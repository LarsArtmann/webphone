package api

import (
	"encoding/json/v2"
	"net/http"
	"strings"

	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/session"
)

// JSON push-device API for the native shells (the Android call shell
// first): one register call per install at login, one deregister at
// logout, list for the Settings/debug surface. The device's wake
// address (ntfy topic or FCM registration token) is write-only — the
// list NEVER echoes it back (a credential stays where it was trusted
// in, the same doctrine as the SIP directory password).

// PushDeviceBodyLimit bounds the register body (a service-ready JSON
// object with room to spare).
const PushDeviceBodyLimit = 8192

// deviceIDShape bounds a registered device id: the shell mints a
// UUID-shaped per-install id; the charset keeps engine keys sane.
const (
	deviceIDMinLen = 8
	deviceIDMaxLen = 64
	tokenMaxLen    = 4096
)

// apiPushDevice is the LIST projection: identity and recency only —
// never the token.
// brandid-lint:ignore(BD003) wire DTO: the device id is a foreign client-generated install identity (a UUID), not a domain entity (strong-id wire ruling)
type apiPushDevice struct {
	// DeviceID is the shell-minted per-install id.
	DeviceID string `description:"The shell-minted per-install id." json:"device_id"`
	// Transport names the wake wire ("ntfy" or "fcm").
	Transport string `description:"The wake wire: ntfy or fcm." json:"transport"`
	// UpdatedAt is the last successful registration, unix nanoseconds.
	UpdatedAt int64 `description:"Last successful registration, unix nanoseconds." json:"updated_at"`
}

// PushDeviceListResponse is the GET /api/push-devices wire shape. Named
// (not inline) so the machine-API contract derives its schema from this
// type (internal/server/apidocs.go).
type PushDeviceListResponse struct {
	Devices []apiPushDevice `description:"The extension's registered devices, newest registration first. The wake address (token) is write-only and never echoed." json:"devices"`
}

// PushDeviceRegisterRequest is the POST /api/push-devices wire shape:
// one register call per install at login.
// brandid-lint:ignore(BD003) wire DTO: foreign client-generated install identity, shape-checked at the door (strong-id wire ruling)
type PushDeviceRegisterRequest struct {
	DeviceID  string `description:"The shell-minted per-install id (8-64 chars of letters, digits, hyphens, underscores)." json:"device_id"`
	Transport string `description:"The wake wire to register: ntfy or fcm (must be configured on the server)." json:"transport"`
	Token     string `description:"The device's wake address: an ntfy topic or an FCM registration token. Write-only — never echoed back." json:"token"`
}

// pushSeamOff answers the disabled-seam 404 — the same contract as
// /api/transcribe: without a configured transport the surface does not
// exist, and the shell must not believe registration succeeded.
func (h *Handlers) pushSeamOff() bool {
	return h.deps.Push == nil || !h.deps.Push.Enabled() || h.deps.Devices == nil
}

// pushSession is the push surfaces' prologue: a live session, then the
// enabled seam. Both gates write their own answer (401 / honest 404), so
// the caller only returns when this reports false.
func (h *Handlers) pushSession(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return sess, false
	}
	if h.pushSeamOff() {
		http.NotFound(w, r)
		return sess, false
	}
	return sess, true
}

// ListPushDevices answers the shell's/Settings' read: the session
// extension's registered devices, newest registration first, without
// any wake address.
func (h *Handlers) ListPushDevices(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.pushSession(w, r)
	if !ok {
		return
	}
	views, err := h.deps.Devices.List(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list push devices", http.StatusInternalServerError)
		return
	}
	body := PushDeviceListResponse{Devices: make([]apiPushDevice, 0, len(views))}
	for _, view := range views {
		body.Devices = append(body.Devices, apiPushDevice{
			DeviceID:  view.DeviceID,
			Transport: view.Transport,
			UpdatedAt: view.UpdatedAt,
		})
	}
	writeJSONBody(w, body)
}

// SavePushDevice registers (or re-registers) the calling install's
// wake address. Re-registration replaces the stored transport/token
// row — FCM token churn must not strand a device. Registering a
// transport the server cannot send to is a 422, not a silent dead
// letter: the fanout's Knows is the honest signal.
func (h *Handlers) SavePushDevice(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.pushSession(w, r)
	if !ok {
		return
	}
	var body PushDeviceRegisterRequest
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, PushDeviceBodyLimit), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validDeviceID(body.DeviceID) {
		http.Error(w, "device_id must be 8-64 characters of letters, digits, hyphens or underscores", http.StatusUnprocessableEntity)
		return
	}
	if !h.deps.Push.Knows(body.Transport) {
		http.Error(w, "transport is not configured on this server", http.StatusUnprocessableEntity)
		return
	}
	if !validDeviceToken(body.Transport, body.Token) {
		http.Error(w, "token is not a plausible device address for the transport", http.StatusUnprocessableEntity)
		return
	}

	err := h.deps.Devices.Upsert(r.Context(), sess.Extension, body.DeviceID, body.Transport, body.Token)
	if err != nil {
		if storeAtCapacity(w, err, "device list is full — deregister one first") {
			return
		}
		http.Error(w, "could not register the device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeletePushDevice deregisters the calling install (logout) scoped to
// the session's extension; an unknown or foreign id is a plain 204 —
// the deregistration outcome is the same.
func (h *Handlers) DeletePushDevice(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.pushSession(w, r)
	if !ok {
		return
	}
	deviceID := r.URL.Query().Get("id")
	if !validDeviceID(deviceID) {
		http.NotFound(w, r)
		return
	}
	if err := h.deps.Devices.Remove(r.Context(), sess.Extension, deviceID); err != nil {
		http.Error(w, "could not deregister the device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validDeviceID checks the install-id shape (UUID-shaped client ids;
// the charset keeps engine keys and query strings sane).
// brandid-lint:ignore(BD003) validator door: the param is the raw wire value under test, not a domain id (strong-id wire ruling)
func validDeviceID(id string) bool {
	if len(id) < deviceIDMinLen || len(id) > deviceIDMaxLen {
		return false
	}
	return asciiIdentifier(id, "-_")
}

// validDeviceToken checks the wake-address shape per transport: an
// ntfy topic (ValidNtfyTopic) or an FCM registration token (long,
// colon/hyphen/alnum charset — Google's shape is stable enough to
// bound, not to pin).
func validDeviceToken(transport, token string) bool {
	if token == "" || len(token) > tokenMaxLen {
		return false
	}
	if transport == push.TransportNtfy {
		return push.ValidNtfyTopic(token)
	}
	return asciiIdentifier(token, ":-_")
}

// asciiIdentifier reports whether every rune is an ASCII letter, digit, or
// one of extra — the shared device-id/token charset rule.
func asciiIdentifier(value, extra string) bool {
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9', strings.ContainsRune(extra, ch):
		default:
			return false
		}
	}
	return true
}
