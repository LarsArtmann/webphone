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
	"github.com/larsartmann/webphone/internal/userauth"
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
	if !h.verifyDirectoryCredentials(w, r, extension, body.Password) {
		return
	}
	if err := h.mintSession(w, r, extension, body.Password); err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, h.sessionIdentityResponse(extension))
}

// verifyDirectoryCredentials is the ONE fail-closed PBX directory check
// behind both login modes (extension and passkey): the submitted (or
// file-sourced) credentials must directory-verify before any session is
// minted — a forged POST must not mint a session scoped to another
// extension, and a stale password file must not mint a dead one (the
// passkey mode sources the password server-side, so this check is its
// only directory proof). Deployments without a phone API (loopback dev)
// skip verification, matching the mode where no PBX-backed panels exist.
// It answers the client itself on failure; false means the response is
// done.
func (h *handlers) verifyDirectoryCredentials(w http.ResponseWriter, r *http.Request, extension domain.Extension, password string) bool {
	if !h.deps.PhoneAPI.Enabled() {
		return true
	}
	ctx, cancel := context.WithTimeout(r.Context(), verifyTimeout)
	err := h.deps.PhoneAPI.VerifyCredentials(ctx, pbx.Credentials{
		Extension: extension.String(),
		Password:  password,
	})
	cancel()
	switch {
	case err == nil:
		return true
	case errors.Is(err, pbx.ErrUnauthorized):
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "credential verification timed out", http.StatusGatewayTimeout)
	default:
		http.Error(w, "credential verification failed", http.StatusBadGateway)
	}
	return false
}

// mintSession is the ONE birth path of a webphone session row + cookie:
// create, set the cookie for the idle window, and rotate the CSRF token
// (fixation defense, the nosurf documented pattern): the response
// deletes the CSRF cookie, and the client immediately adopts the fresh
// token via GET /api/csrf — the no-reload login needs that adoption
// step because the served page carries the old masked token.
func (h *handlers) mintSession(w http.ResponseWriter, r *http.Request, extension domain.Extension, password string) error {
	token, err := h.deps.Sessions.Create(extension, password)
	if err != nil {
		return err //nolint:wrapcheck // classified by the store seam; the caller answers 500
	}
	session.SetCookie(w, r, token, h.deps.Config.SessionTTL)
	httputil.InvalidateCSRFCookie(w, httputil.CSRFConfig{})
	return nil
}

// sessionIdentity is the typed wire shape of what a freshly minted
// (or resumed) session tells the client about itself: the extension (the
// island's SIP REGISTER needs it), its presented DID when configured
// (identities), and — when the extension is passkey-mapped — the user's
// display name and every number mapped to that user, so the whoami line
// can lead with the human identity instead of a bare extension@sip_domain.
// Password rides the same shape on the surfaces that must return the
// credential (resume, passkey finish); the struct's omitempty tags keep
// the wire identical to the historical map form.
type sessionIdentity struct {
	Extension string `json:"extension"`
	DID       string `json:"did,omitempty"` //nolint:branching-flow // wire DTO: the DID is a config-validated identity rendered verbatim (omitempty); no id ever flows back IN through this field
	DisplayName string   `json:"display_name,omitempty"`
	Numbers     []string `json:"numbers,omitempty"`
	Password    string   `json:"password,omitempty"`
}

// sessionIdentityResponse shapes what a freshly minted (or resumed)
// session tells the client about itself (see sessionIdentity).
func (h *handlers) sessionIdentityResponse(extension domain.Extension) sessionIdentity {
	response := sessionIdentity{Extension: extension.String()}
	response.DID = h.identityFor(extension)
	if h.deps.UserAuth != nil {
		if mapped, ok := h.deps.UserAuth.MappedByExtension(extension); ok {
			response.DisplayName = mapped.DisplayName
			response.Numbers = h.numbersFor(mapped)
		}
	}
	return response
}

// numbersFor lists the mapped user's presented numbers (identities of
// every mapped extension), deduplicated in mapping order — the "numbers
// available to this user" the whoami line leads with.
func (h *handlers) numbersFor(mapped userauth.MappedUser) []string {
	seen := make(map[string]bool, len(mapped.Extensions))
	numbers := make([]string, 0, len(mapped.Extensions))
	for _, ext := range mapped.Extensions {
		did := h.identityFor(ext)
		if did == "" || seen[did] {
			continue
		}
		seen[did] = true
		numbers = append(numbers, did)
	}
	return numbers
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
	w.Header().Set("Cache-Control", "no-store")
	response := h.sessionIdentityResponse(sess.Extension)
	// The password is the session's own payload (the row already carries
	// it for the phone-api proxy; the browser REGISTER needs it by
	// design) and it is served only to the cookie that proved itself at
	// login. no-store keeps the credential out of every cache.
	response.Password = sess.Password
	writeJSON(w, http.StatusOK, response)
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
