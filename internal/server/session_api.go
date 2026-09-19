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
	// Rotate the CSRF token on login: a token minted before authentication
	// must never survive into the authenticated session (CSRF fixation).
	httputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.MarshalWrite(w, map[string]string{"extension": extension.String()}) //nolint:erraudit // best-effort write; the response is already committed
}

// destroySession signs the tab session out (island logout).
func (h *handlers) destroySession(w http.ResponseWriter, r *http.Request) {
	h.deps.Sessions.Delete(session.TokenFromRequest(r))
	session.ClearCookie(w)
	httputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.WriteHeader(http.StatusNoContent)
}
