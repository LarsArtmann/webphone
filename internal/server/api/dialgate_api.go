package api

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// dialGateMax bounds the dialable remote string (the island sanitizes
// before asking; the bound is the abuse guard).
const dialGateMax = 64

// DialGateRequest is the pre-invite ask (issue #7): the sanitized
// remote plus, on the click-through, the override flag.
// brandid-lint:ignore(BD003) wire DTO: the remote is the dialable string the island dials verbatim (strong-id wire ruling)
type DialGateRequest struct {
	Remote   string `description:"The sanitized dialable remote the island is about to invite." json:"remote"`
	Override bool   `description:"True on the click-through: spend this remote's one-shot override." json:"override,omitzero"`
}

// DialGateResponse is the verdict: allowed=true means the attempt is
// RECORDED and the invite may go out; allowed=false carries the slot's
// last-at stamp and the window length so the island can phrase the
// remaining wait without guessing the configured size.
// brandid-lint:ignore(BD003) wire DTO: the remote rides the request, the stamp is the slot's last attempt (strong-id wire ruling)
type DialGateResponse struct {
	Allowed bool  `description:"True when the dial may proceed (the attempt is recorded)." json:"allowed"`
	LastAt  int64 `description:"Unix seconds of the slot's last allowed dial (denied asks only)." json:"last_at,omitzero"`
	Window  int64 `description:"The configured cooldown window in seconds (denied asks only; powers the remaining-wait phrasing)." json:"window,omitzero"`
}

// DialGate is the advisory pre-invite gate (issue #7): one
// check+record transaction per dial attempt. Contract:
//
//   - 200 allowed:true — the attempt is recorded, dial.
//   - 200 allowed:false — inside the cooldown window (nothing
//     recorded: a retry loop must not extend the window it fights).
//   - 404 when the cooldown is off (the disabled-seam convention —
//     the island reads it as "no gate, dial").
//   - 503 when the window store is nil (test compositions).
//   - 502 on a store failure: the island dials anyway (the gate is
//     advisory — a phone must dial), but the failure is logged.
//
// DialSlotView is one active cooldown window in the typeahead's
// decoration payload.
// brandid-lint:ignore(BD003) wire DTO: the remote is the dialable string (strong-id wire ruling)
type DialSlotView struct {
	Remote string `description:"The sanitized remote inside its cooldown window." json:"remote"`
	LastAt int64  `description:"Unix seconds of the slot's last allowed dial." json:"last_at"`
	Window int64  `description:"The configured cooldown window in seconds." json:"window"`
}

// DialSlotsResponse lists the owner's ACTIVE cooldown windows (newest
// first) — the typeahead decorates matching contact rows with the
// remaining wait instead of state nobody can see.
type DialSlotsResponse struct {
	Slots []DialSlotView `description:"Active windows, newest first (empty when none are live)." json:"slots"`
}

// DialSlots serves the owner's active cooldown windows (issue #7: the
// typeahead-row state). Same gate shape as the POST: 404 when the
// cooldown is off, 503 when the store is nil, 502 on a store failure —
// the island hides the decoration on failure, it never blocks a dial.
func (h *Handlers) DialSlots(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.DialCooldown <= 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if h.deps.Dials == nil {
		http.Error(w, "dial window store unavailable", http.StatusServiceUnavailable)
		return
	}
	slots, err := h.deps.Dials.Active(r.Context(), sess.Extension, h.deps.DialCooldown)
	if err != nil {
		slog.Warn("dial-gate: active-slot scan failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not read the dial windows", http.StatusBadGateway)
		return
	}
	views := make([]DialSlotView, len(slots))
	for i, s := range slots {
		views[i] = DialSlotView{Remote: s.Remote, LastAt: s.LastAt, Window: s.Window}
	}
	writeJSONBody(w, DialSlotsResponse{Slots: views})
}

func (h *Handlers) DialGate(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.DialCooldown <= 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if h.deps.Dials == nil {
		http.Error(w, "dial window store unavailable", http.StatusServiceUnavailable)
		return
	}
	dec := jsontext.NewDecoder(r.Body, json.RejectUnknownMembers(true))
	var req DialGateRequest
	if err := json.UnmarshalDecode(dec, &req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.Remote = strings.TrimSpace(req.Remote)
	if req.Remote == "" || len(req.Remote) > dialGateMax || strings.ContainsAny(req.Remote, "/%") {
		http.Error(w, "invalid remote", http.StatusBadRequest)
		return
	}

	if req.Override {
		if err := h.deps.Dials.ArmOverride(r.Context(), sess.Extension, req.Remote); err != nil {
			slog.Warn("dial-gate: override arm failed", "error", err, "remote", req.Remote, "extension", sess.Extension.String())
			http.Error(w, "could not arm the dial override", http.StatusBadGateway)
			return
		}
	}
	allowed, stamp, err := h.deps.Dials.Attempt(r.Context(), sess.Extension, req.Remote, h.deps.DialCooldown)
	if err != nil {
		slog.Warn("dial-gate: attempt failed", "error", err, "remote", req.Remote, "extension", sess.Extension.String())
		http.Error(w, "could not consult the dial gate", http.StatusBadGateway)
		return
	}
	resp := DialGateResponse{Allowed: allowed}
	if !allowed {
		resp.LastAt = stamp.Unix()
		resp.Window = int64(h.deps.DialCooldown / time.Second)
	}
	writeJSONBody(w, resp)
}
