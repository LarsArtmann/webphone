package server

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
)

// apiLogCall receives the island's post-call report and appends one call
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
// but the endpoint stays behind requireSession like every /api surface so
// the auth posture never depends on integration wiring.
func (h *handlers) apiLogCall(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	var body struct {
		Number    string `json:"number"`
		Direction string `json:"direction"`
		Seconds   int    `json:"seconds"`
		Outcome   string `json:"outcome"`
		Key       string `json:"key"`
	}
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
		if h.callsIdem.seen(idemKey) {
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
		if idemKey != "" {
			h.callsIdem.record(idemKey)
		}
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
	if idemKey != "" {
		h.callsIdem.record(idemKey)
	}

	slog.Debug("crm: call logged", "extension", sess.Extension.String(), "direction", body.Direction)
	w.WriteHeader(http.StatusNoContent)
}
