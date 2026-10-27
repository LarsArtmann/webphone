package api

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// CallDispositionRequest is the POST /api/calls/{id}/disposition wire
// shape: the operator's post-call verdict plus the call facts the CRM
// correction write re-resolves. Named (not inline) so the machine-API
// contract derives its schema from this type (internal/server/apidocs.go).
type CallDispositionRequest struct {
	Disposition string `description:"The operator's verdict: answered|voicemail|ivr|missed|failed|unknown. Only a human may claim answered." json:"disposition"`
	Number      string `description:"The remote number (verbatim) the CRM contact resolves against." json:"number"`
	Direction   string `description:"Call direction: in or out." json:"direction"`
	Seconds     int    `description:"Duration in seconds; negative values are clamped to 0." json:"seconds"`
}

// ConfirmCallDisposition records the operator's post-call verdict for
// one call (issue #6): the ONE authority that may claim "answered".
// Contract:
//
//   - The verdict persists FIRST (it is a fact about the human's
//     knowledge, independent of CRM reachability); the CRM write is the
//     awaited correction afterward. A CRM 502 leaves the verdict
//     stored and the retry open — the island re-POSTs the same body.
//   - CRM disabled or a number without a contact: verdict-only 204 (the
//     same never-mint-contacts doctrine as the auto-report).
//   - Idempotent per (call, verdict): the same verdict again is an
//     inert 204 without a second CRM journal entry; a DIFFERENT verdict
//     is a correction and journals again.
func (h *Handlers) ConfirmCallDisposition(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.Notes == nil {
		http.Error(w, "note storage is unavailable", http.StatusServiceUnavailable)
		return
	}
	callID := r.PathValue("id")
	if !validNoteCallID(callID) {
		http.Error(w, "invalid call id", http.StatusBadRequest)
		return
	}

	var body CallDispositionRequest
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, transcriptBodyLimit), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	disposition, err := domain.ParseDisposition(body.Disposition)
	if err != nil {
		http.Error(w, "disposition must be one of answered, missed, voicemail, ivr, failed, unknown", http.StatusBadRequest)
		return
	}
	if body.Direction != "in" && body.Direction != "out" {
		http.Error(w, "direction must be in or out", http.StatusBadRequest)
		return
	}
	if body.Seconds < 0 {
		body.Seconds = 0
	}

	if err := h.deps.Notes.SetDisposition(r.Context(), sess.Extension, callID, disposition); err != nil {
		if errors.Is(err, readmodels.ErrNoteListFull) {
			http.Error(w, "the note list is full — delete another note first", http.StatusConflict)
			return
		}
		slog.Warn("disposition save failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not save the disposition", http.StatusInternalServerError)
		return
	}

	if !h.deps.CRM.Enabled() {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	idemKey := sess.Extension.String() + ":" + callID + ":" + disposition.String()
	if h.deps.CallsIdem.Seen(idemKey) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	phone, err := domain.ParsePhone(body.Number)
	if err != nil {
		// The verdict stands; without a parseable number there is no
		// CRM correction to make and nothing retryable — the same key
		// would resolve identically every time.
		slog.Debug("crm: disposition has no parsable number; correction skipped", "extension", sess.Extension.String())
		h.recordCallIdem(idemKey)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	match, found := h.deps.CRM.Resolve(r.Context(), phone.String())
	if !found {
		slog.Debug("crm: disposition not corrected; no contact for number", "extension", sess.Extension.String())
		h.recordCallIdem(idemKey)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.deps.CRM.LogCall(r.Context(), match.ID, body.Direction, phone.String(), body.Seconds, disposition); err != nil {
		slog.Warn("crm: disposition correction failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "the CRM could not record the correction", http.StatusBadGateway)
		return
	}
	h.recordCallIdem(idemKey)

	slog.Debug("crm: disposition corrected", "extension", sess.Extension.String(), "disposition", disposition.String())
	w.WriteHeader(http.StatusNoContent)
}
