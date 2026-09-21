package server

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"time"

	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
)

// verifyTimeout bounds the PBX round-trip at login: the island awaits
// this POST before showing the signed-in state, and a hung PBX must not
// hang logins for the full client timeout.
const verifyTimeout = 5 * time.Second

// createSession opens the tab/proxy session. The submitted credentials
// are verified against the PBX directory first (VerifyCredentials: the
// same directory the SIP REGISTER checks): a forged POST must not mint
// a session scoped to another extension — the tab partials, fax and
// attachment streams, and SSE fragments all scope by the session alone.
// Deployments without a phone API (loopback dev) skip verification,
// matching the mode where no PBX-backed panels exist.
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
	if h.deps.PhoneAPI.Enabled() {
		ctx, cancel := context.WithTimeout(r.Context(), verifyTimeout)
		err := h.deps.PhoneAPI.VerifyCredentials(ctx, pbx.Credentials{
			Extension: body.Extension,
			Password:  body.Password,
		})
		cancel()
		switch {
		case err == nil:
		case errors.Is(err, pbx.ErrUnauthorized):
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		default:
			http.Error(w, "credential verification failed", http.StatusBadGateway)
			return
		}
	}

	token, err := h.deps.Sessions.Create(extension, body.Password)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	session.SetCookie(w, r, token, h.deps.Config.SessionTTL)
	// Rotate the CSRF token on login (fixation defense, the nosurf
	// documented pattern): the response deletes the CSRF cookie, and the
	// island immediately adopts the fresh token via GET /api/csrf,
	// the no-reload login needs that adoption step because the served
	// page carries the old masked token.
	httputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// did is the extension's presented PSTN number (config identities);
	// the island appends it to the whoami line so users see their real
	// number, not just extension@sip_domain.
	response := map[string]string{"extension": extension.String()}
	if did := h.identityFor(extension); did != "" {
		response["did"] = did
	}
	_ = json.MarshalWrite(w, response) //nolint:erraudit // best-effort write; the response is already committed
}

// getSession resumes a live session for an island page load: the
// browser probes BEFORE rendering the login form, and a live cookie
// gets the session's SIP credentials back so the browser-side SIP
// REGISTER re-runs without the user typing anything. The password is
// the session's own payload (the row already carries it for the
// phone-api proxy; the browser REGISTER needs it by design) and it is
// served only to the cookie that proved itself at login. no-store keeps
// the credential out of every cache.
func (h *handlers) getSession(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"extension": sess.Extension.String(),
		"password":  sess.Password,
	}
	if did := h.identityFor(sess.Extension); did != "" {
		response["did"] = did
	}
	_ = json.MarshalWrite(w, response) //nolint:erraudit // best-effort write; the response is already committed
}

// destroySession signs the tab session out (island logout).
func (h *handlers) destroySession(w http.ResponseWriter, r *http.Request) {
	h.deps.Sessions.Delete(session.TokenFromRequest(r))
	session.ClearCookie(w)
	// Logout rotates too: the CSRF token must not outlive the session it
	// was issued alongside. The island reloads right after (its contract),
	// so the fresh page re-renders meta/hx-headers from the regenerated
	// cookie.
	httputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	w.WriteHeader(http.StatusNoContent)
}
