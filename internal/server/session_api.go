package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// createSession is called by the island AFTER the PBX accepted its SIP
// REGISTER — the credentials are proven good. The server trusts that call
// only to open the tab/proxy session; every phone-api access revalidates
// against the PBX anyway (it would 401 on wrong credentials).
func (h *handlers) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Extension string `json:"extension"`
		Password  string `json:"password"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	extension, err := domain.ParseExtension(body.Extension)
	if err != nil {
		http.Error(w, "invalid extension", http.StatusBadRequest)
		return
	}
	if body.Password == "" {
		http.Error(w, "missing password", http.StatusBadRequest)
		return
	}

	token, err := h.deps.Sessions.Create(extension, body.Password)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	session.SetCookie(w, r, token, h.deps.Config.SessionTTL)
	// Rotate the CSRF token on login (fixation defense, the nosurf
	// documented pattern): the response deletes the CSRF cookie, and the
	// island immediately adopts the fresh token via GET /api/csrf —
	// the no-reload login needs that adoption step because the served
	// page carries the old masked token.
	htputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.MarshalWrite(w, map[string]string{"extension": extension.String()}) //nolint:erraudit // best-effort write; the response is already committed
}

// destroySession signs the tab session out (island logout).
func (h *handlers) destroySession(w http.ResponseWriter, r *http.Request) {
	h.deps.Sessions.Delete(session.TokenFromRequest(r))
	session.ClearCookie(w)
	// Logout rotates too: the CSRF token must not outlive the session it
	// was issued alongside. The island reloads right after (its contract),
	// so the fresh page re-renders meta/hx-headers from the regenerated
	// cookie.
	htputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.WriteHeader(http.StatusNoContent)
}
