package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/larsartmann/go-error-family"
)

// ssoBegin starts the authorization-code + PKCE dance: mint the attempt's
// secrets, park them in cookies, and hand the browser to the IdP.
func (a *Handlers) SSOBegin(w http.ResponseWriter, r *http.Request) {
	if a.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	challenge, err := a.deps.SSO.Challenge(r.Context())
	if err != nil {
		a.ssoFailed(w, r, err, "begin oidc login")
		return
	}
	a.setSSOCookie(w, SSOStateCookie, challenge.State)
	a.setSSOCookie(w, SSONonceCookie, challenge.Nonce)
	a.setSSOCookie(w, SSOVerifierCookie, challenge.Verifier)
	// The browser leaves for the IdP: no referrer (the authorize URL
	// carries the state; it must not leak through third-party scripts on
	// our own page) and no caching of the redirect itself.
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, challenge.URL, http.StatusFound)
}

// SSOCallback completes the dance: verify the state against the cookie,
// exchange the code, resolve the mapped identity, verify the SIP
// credentials against the PBX directory and mint the ONE webphone session.
// Every failure lands back on the login card with a short reason code; the
// honest detail goes to the journal (a browser cannot act on it).
func (a *Handlers) SSOCallback(w http.ResponseWriter, r *http.Request) {
	if a.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query()
	if idpErr := query.Get("error"); idpErr != "" {
		a.ssoFailed(w, r, errorfamily.NewRejection("sso.idp_error",
			"the IdP refused the authorization request (%s)").WithContext("error", idpErr), "idp refused")
		return
	}
	state, err := a.readSSOCookie(r, SSOStateCookie)
	if err != nil {
		a.ssoFailed(w, r, err, "read state cookie")
		return
	}
	nonce, err := a.readSSOCookie(r, SSONonceCookie)
	if err != nil {
		a.ssoFailed(w, r, err, "read nonce cookie")
		return
	}
	verifier, err := a.readSSOCookie(r, SSOVerifierCookie)
	if err != nil {
		a.ssoFailed(w, r, err, "read verifier cookie")
		return
	}
	a.clearSSOCookie(w, SSOStateCookie)
	a.clearSSOCookie(w, SSONonceCookie)
	a.clearSSOCookie(w, SSOVerifierCookie)
	if subtle.ConstantTimeCompare([]byte(state), []byte(query.Get("state"))) != 1 {
		a.ssoFailed(w, r, errorfamily.NewRejection("sso.state.mismatch",
			"the authorization response state does not match this browser's sign-in attempt"), "state mismatch")
		return
	}
	mapped, password, err := a.deps.SSO.Complete(r.Context(), query.Get("code"), verifier, nonce)
	if err != nil {
		a.ssoFailed(w, r, err, "complete oidc login")
		return
	}
	extension := mapped.SessionExtension()
	if !a.verifyDirectoryCredentials(w, r, extension, password) {
		return
	}
	// The session remembers WHO minted it (the multi-user substrate):
	// passkey and OIDC logins carry the same identity field.
	if _, err := a.mintSession(w, r, extension, password, mapped.Email); err != nil {
		slog.Error("webphone: oidc session mint failed", "error", err, "extension", extension.String())
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	AuthAudit("login", "email", mapped.Email, "oidc verified; session minted")
	// A 303 lands the browser on the shell as a fresh GET, where the
	// island's boot probe (GET /api/session) signs it in and registers —
	// the exact resume path a returning browser takes.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ssoFailed logs the classified detail and returns the browser to the
// login page with a short reason code the page translates into its alert
// (ssoLoginFailureText). The island card no longer renders any login
// surface, so the failure lands on /login directly.
func (a *Handlers) ssoFailed(w http.ResponseWriter, r *http.Request, err error, op string) {
	slog.Warn("webphone: oidc login failed", "op", op, "error", err,
		"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err))
	a.clearSSOCookie(w, SSOStateCookie)
	a.clearSSOCookie(w, SSONonceCookie)
	a.clearSSOCookie(w, SSOVerifierCookie)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/login?sso_error="+SSOErrorCode(err), http.StatusSeeOther)
}

// SSOErrorCode maps a failure onto the browser-visible reason
// vocabulary (the ?sso_error= short codes): the browser gets an
// actionable sentence, never an internal error string. Exported: the
// vocabulary is a wire contract the server tests pin.
func SSOErrorCode(err error) string {
	switch errorfamily.Code(err) {
	case "sso.unmapped_email":
		return "unmapped"
	case "authmap.password_file.missing", "authmap.password_file.read", "authmap.password_file.empty":
		return "directory"
	case "sso.email_unverified":
		return "unverified"
	case "sso.groups_denied":
		return "groups"
	case "sso.state.mismatch", "sso.nonce.mismatch", "sso.attempt.missing":
		return "stale"
	case "sso.client_secret.read", "sso.client_secret.empty", "sso.exchange":
		return "config"
	case "sso.discovery", "sso.userinfo", "sso.id_token.verify":
		return "unavailable"
	default:
		return "failed"
	}
}

func (a *Handlers) setSSOCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // HttpOnly + Secure (wired policy) + SameSite=Lax + path-scoped: the attempt secrets never reach scripts or another site
		Name:     name,
		Value:    value,
		Path:     SSOCallbackPath,
		MaxAge:   ssoCookieTTL,
		HttpOnly: true,
		Secure:   a.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Handlers) readSSOCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return "", errorfamily.Newf(errorfamily.Rejection, "sso.attempt.missing",
			"this sign-in attempt has no live %s cookie — restart the sign-in from this browser", name)
	}
	return cookie.Value, nil
}

func (a *Handlers) clearSSOCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // deletion mirrors setSSOCookie's attributes; a Mismatched Secure attribute would leave the cookie behind
		Name:     name,
		Value:    "",
		Path:     SSOCallbackPath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}
