package auth

import (
	"encoding/json/v2"
	"net/http"
	"time"

	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// verifyTimeout bounds the PBX round-trip at login: the island awaits
// this POST before showing the signed-in state, and a hung PBX must not
// hang logins for the full client timeout.
const verifyTimeout = 5 * time.Second

// LoginRequest is the POST /api/session wire shape: the extension and
// its PBX directory password. Named (not inline) so the machine-API
// contract derives its schema from this type (internal/server/apidocs.go).
type LoginRequest struct {
	Extension string `description:"The PBX extension to sign in, e.g. 1001." json:"extension"`
	Password  string `description:"The extension's PBX directory password." json:"password"`
}

// CreateSession opens the tab/proxy session. The submitted credentials
// are verified against the PBX directory first (VerifyCredentials: the
// same directory the SIP REGISTER checks): a forged POST must not mint
// a session scoped to another extension — the tab partials, fax and
// attachment streams, and SSE fragments all scope by the session alone.
// Deployments without a phone API (loopback dev) skip verification,
// matching the mode where no PBX-backed panels exist.
func (a *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := json.UnmarshalRead(r.Body, &body); err != nil {
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
	if !a.verifyDirectoryCredentials(w, r, extension, body.Password) {
		return
	}
	// The classic extension login has no identity beyond the extension
	// itself: the session's LoginEmail stays empty (byte-parity with
	// every pre-identity session row).
	token, err := a.mintSession(w, r, extension, body.Password, "")
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	response := a.sessionAnswer(w, r.Context(), token, extension, "")
	AuthAudit("login", "extension", extension.String(), "ok")
	writeJSON(w, http.StatusCreated, response)
}

// GetSession resumes a live session for an island page load: the
// browser probes BEFORE rendering the login form, and a live cookie
// gets the session's SIP credentials back so the browser-side SIP
// REGISTER re-runs without the user typing anything. The password is
// the session's own payload (the row already carries it for the
// phone-api proxy; the browser REGISTER needs it by design) and it is
// served only to the cookie that proved itself at login. no-store keeps
// the credential out of every cache.
func (a *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	sess, ok := a.shell.RequireSession(w, r)
	if !ok {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	response := a.sessionAnswer(w, r.Context(), session.TokenFromRequest(r), sess.Extension, sess.LoginEmail)
	// The resume endpoint is one of the two legacy-carrier surfaces (the
	// browser REGISTER needs the row's own payload); its credential plus
	// the unconditional no-store ride this site, not the shared policy.
	if !a.deps.Config.SIP.DerivedCredentials() {
		response.Password = sess.Password
	}
	writeJSON(w, http.StatusOK, response)
}

// DestroySession signs the tab session out (island logout).
func (a *Handlers) DestroySession(w http.ResponseWriter, r *http.Request) {
	if sess, ok := a.deps.Sessions.Get(session.TokenFromRequest(r)); ok {
		// The logout line names the HUMAN when the session has one
		// (passkey/OIDC) — same identity-event parity as the login lines.
		if sess.LoginEmail != "" {
			AuthAudit("logout", "email", sess.LoginEmail, "ok")
		} else {
			AuthAudit("logout", "extension", sess.Extension.String(), "ok")
		}
	}
	a.deps.Sessions.Delete(session.TokenFromRequest(r))
	session.ClearCookie(w, a.cookies)
	// Logout rotates too: the CSRF token must not outlive the session it
	// was issued alongside. The island reloads right after (its contract),
	// so the fresh page re-renders meta/hx-headers from the regenerated
	// cookie.
	httputil.InvalidateCSRFCookie(w, a.csrfCfg)
	w.WriteHeader(http.StatusNoContent)
}
