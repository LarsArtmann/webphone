package hooks

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/larsartmann/webphone/internal/push"
)

// hookRing is the inbound-call wake lane: the stack's FS dialplan POSTs
// here the moment an extension starts ringing, and again with state
// "end" when the ringing stops. The handler fans a push out to every
// device the owner registered and acknowledges 202 — the dialplan must
// never wait on (or fail because of) push health. Per-device failures
// log and continue; a transport reporting the device dead (FCM
// UNREGISTERED, ntfy's 404 lane) prunes the registration on the spot.
// RingRequest is the POST /hooks/ring wire shape (the inbound-call wake
// lane from the stack's FS dialplan). Named (not inline) so the
// wire-contract catalog derives its schema from this type
// (internal/server/wirecontracts.go).
// brandid-lint:ignore(BD003) wire contract: the PBX call id is an opaque foreign token, correlated verbatim by the shell (strong-id wire ruling)
type RingRequest struct {
	Owner  string `json:"owner"`
	From   string `json:"from"`
	CallID string `json:"call_id"`
	State  string `json:"state"` // "ring" | "end"
}

func (h *Handlers) hookRing(w http.ResponseWriter, r *http.Request) {
	var payload RingRequest
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	owner, from, ok := parseOwnerFrom(w, payload.Owner, payload.From)
	if !ok {
		return
	}

	var event push.Event
	switch payload.State {
	case "ring":
		event = push.EventRing
	case "end":
		event = push.EventRingEnd
	default:
		http.Error(w, "state must be \"ring\" or \"end\"", http.StatusBadRequest)
		return
	}

	if h.push == nil || !h.push.Enabled() || h.devices == nil {
		// Inert ack: the lane exists (the dialplan configured it), the
		// seam does not. Same answer as the no-devices case below.
		w.WriteHeader(http.StatusAccepted)
		return
	}

	targets, err := h.devices.List(r.Context(), owner)
	if err != nil {
		webhookFail(w, "ring-devices", err)
		return
	}
	ring := push.Ring{Event: event, From: from.String(), CallID: payload.CallID}
	for _, target := range targets {
		if err := h.push.Send(r.Context(), target.Transport, target.Token, ring); err != nil {
			if errors.Is(err, push.ErrUnregistered) {
				if err := h.devices.Remove(r.Context(), owner, target.DeviceID); err != nil {
					slog.Warn("push: prune dead device failed", "device", target.DeviceID, "err", err)
				} else {
					slog.Info("push: pruned dead device", "device", target.DeviceID, "transport", target.Transport)
				}
				continue
			}
			slog.Warn("push: ring fan-out failed",
				"device", target.DeviceID, "transport", target.Transport, "event", event, "err", err)
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
