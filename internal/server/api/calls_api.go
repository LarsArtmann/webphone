package api

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
)

// LogCall receives the island's post-call report and appends one call
// activity to the CRM contact that owns the remote number. Contract:
//
//   - 204 for every delivered-or-deliberately-dropped report: disabled CRM
//     (island raced a config change), unknown number (no contact exists —
//     the integration NEVER creates contacts from raw callers), everything
//     landed. The call is a fact about the phone, but a CRM fact needs a
//     contact stream; silently dropping keeps the island's UX clean.
//   - 400 on malformed bodies (island bug, must surface in dev).
//   - 502 when the CRM itself failed (down, 5xx): the island toasts.
//   - A non-empty "key" (island-generated UUID) makes a retry inert: the
//     same key answers 204 without journaling again. The key is recorded
//     only on a delivered outcome (both 204 shapes), so a 502 stays
//     retryable; an absent or empty key never dedupes (legacy shape).
//
// The session's extension scopes nothing here — the CRM is single-user —
// but the endpoint stays behind the session gate like every /api surface
// so the auth posture never depends on integration wiring.
// recordCallIdem marks a call-log attempt processed. Only successes reach
// it (failures stay retryable); an empty key means the client sent no
// Idempotency-Key, so there is nothing to dedupe on.
func (h *Handlers) recordCallIdem(key string) {
	if key != "" {
		h.deps.CallsIdem.Record(key)
	}
}

// CallLogRequest is the POST /api/calls wire shape: the island's
// post-call report. Named (not inline) so the machine-API contract
// derives its schema from this type (internal/server/apidocs.go).
type CallLogRequest struct {
	Number    string `description:"The remote number the call was made with, e.g. +4917012345678." json:"number"`
	Direction string `description:"Call direction: in or out." json:"direction"`
	Seconds   int    `description:"Duration in seconds; negative values are clamped to 0." json:"seconds"`
	Outcome   string `description:"answered or missed." json:"outcome"`
	Key       string `description:"Island-generated idempotency key (UUID, max 128 chars). A retry with the same key answers 204 without journaling again; a failed (502) attempt stays retryable. Absent or empty means no dedupe." json:"key"`
}

func (h *Handlers) LogCall(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}

	var body CallLogRequest
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Direction != "in" && body.Direction != "out" {
		http.Error(w, "direction must be in or out", http.StatusBadRequest)
		return
	}

	phone, err := domain.ParsePhone(body.Number)
	if err != nil {
		http.Error(w, "enter a valid number", http.StatusUnprocessableEntity)
		return
	}

	if !h.deps.CRM.Enabled() {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Extension-namespaced so the dedupe memory never crosses tenants,
	// even though today's CRM is single-user — the seam should not leak
	// that implementation detail into the key space.
	idemKey := ""
	if body.Key != "" {
		if len(body.Key) > 128 {
			http.Error(w, "idempotency key too long", http.StatusBadRequest)
			return
		}
		idemKey = sess.Extension.String() + ":" + body.Key
		if h.deps.CallsIdem.Seen(idemKey) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	match, found := h.deps.CRM.Resolve(r.Context(), phone.String())
	if !found {
		// Deliberate: an unknown caller gets no CRM journal entry. The
		// integration never mints contacts; the number still lives in the
		// island's recent calls and the PBX CDR.
		slog.Debug("crm: call not logged; no contact for number", "extension", sess.Extension.String())
		h.recordCallIdem(idemKey)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if body.Seconds < 0 {
		body.Seconds = 0
	}

	if err := h.deps.CRM.LogCall(r.Context(), match.ID, body.Direction, phone.String(), body.Seconds, body.Outcome); err != nil {
		slog.Warn("crm: call logging failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "the CRM could not record the call", http.StatusBadGateway)
		return
	}
	h.recordCallIdem(idemKey)

	slog.Debug("crm: call logged", "extension", sess.Extension.String(), "direction", body.Direction)
	w.WriteHeader(http.StatusNoContent)
}
