package server

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/sso"
)

// SSOClient is the OIDC login seam the server drives: one challenge (the
// IdP URL plus the state/nonce/PKCE values the browser must echo back) and
// one completion (code → mapped identity + SIP directory password). The
// concrete *sso.Client is the only production implementation; tests
// supply a stub, which is why the field is an interface.
type SSOClient interface {
	Challenge(ctx context.Context) (sso.Challenge, error)
	Complete(ctx context.Context, code, verifier, nonce string) (authmap.User, string, error)
}

// The three short-lived cookies carrying one authorization attempt. They
// are HttpOnly (no script needs them), SameSite=Lax (the IdP redirect is a
// top-level GET navigation, so they ride it while cross-site POSTs cannot
// trigger the dance), Path-scoped to the callback and expired as soon as
// the callback consumes them — a replay finds nothing to replay.
const (
	ssoStateCookie    = "wp_sso_state"
	ssoNonceCookie    = "wp_sso_nonce"
	ssoVerifierCookie = "wp_sso_verifier"
	ssoCookiePath     = "/api/auth/oauth2"
	ssoCookieTTL      = 600
)

// ssoBegin starts the authorization-code + PKCE dance: mint the attempt's
// secrets, park them in cookies, and hand the browser to the IdP.
func (h *handlers) ssoBegin(w http.ResponseWriter, r *http.Request) {
	if h.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	challenge, err := h.deps.SSO.Challenge(r.Context())
	if err != nil {
		h.ssoFailed(w, r, err, "begin oidc login")
		return
	}
	h.setSSOCookie(w, ssoStateCookie, challenge.State)
	h.setSSOCookie(w, ssoNonceCookie, challenge.Nonce)
	h.setSSOCookie(w, ssoVerifierCookie, challenge.Verifier)
	// The browser leaves for the IdP: no referrer (the authorize URL
	// carries the state; it must not leak through third-party scripts on
	// our own page) and no caching of the redirect itself.
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, challenge.URL, http.StatusFound)
}

// ssoCallback completes the dance: verify the state against the cookie,
// exchange the code, resolve the mapped identity, verify the SIP
// credentials against the PBX directory and mint the ONE webphone session.
// Every failure lands back on the login card with a short reason code; the
// honest detail goes to the journal (a browser cannot act on it).
func (h *handlers) ssoCallback(w http.ResponseWriter, r *http.Request) {
	if h.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query()
	if idpErr := query.Get("error"); idpErr != "" {
		h.ssoFailed(w, r, errorfamily.NewRejection("sso.idp_error",
			"the IdP refused the authorization request (%s)").WithContext("error", idpErr), "idp refused")
		return
	}
	state, err := h.readSSOCookie(r, ssoStateCookie)
	if err != nil {
		h.ssoFailed(w, r, err, "read state cookie")
		return
	}
	nonce, err := h.readSSOCookie(r, ssoNonceCookie)
	if err != nil {
		h.ssoFailed(w, r, err, "read nonce cookie")
		return
	}
	verifier, err := h.readSSOCookie(r, ssoVerifierCookie)
	if err != nil {
		h.ssoFailed(w, r, err, "read verifier cookie")
		return
	}
	h.clearSSOCookie(w, ssoStateCookie)
	h.clearSSOCookie(w, ssoNonceCookie)
	h.clearSSOCookie(w, ssoVerifierCookie)
	if subtle.ConstantTimeCompare([]byte(state), []byte(query.Get("state"))) != 1 {
		h.ssoFailed(w, r, errorfamily.NewRejection("sso.state.mismatch",
			"the authorization response state does not match this browser's sign-in attempt"), "state mismatch")
		return
	}
	mapped, password, err := h.deps.SSO.Complete(r.Context(), query.Get("code"), verifier, nonce)
	if err != nil {
		h.ssoFailed(w, r, err, "complete oidc login")
		return
	}
	extension := mapped.SessionExtension()
	if !h.verifyDirectoryCredentials(w, r, extension, password) {
		return
	}
	// The session remembers WHO minted it (the multi-user substrate):
	// passkey and OIDC logins carry the same identity field.
	if _, err := h.mintSession(w, r, extension, password, mapped.Email); err != nil {
		slog.Error("webphone: oidc session mint failed", "error", err, "extension", extension.String())
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	authAudit("login", "email", mapped.Email, "oidc verified; session minted")
	// A 303 lands the browser on the shell as a fresh GET, where the
	// island's boot probe (GET /api/session) signs it in and registers —
	// the exact resume path a returning browser takes.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ssoFailed logs the classified detail and returns the browser to the
// login card with a short reason code the island can localize.
func (h *handlers) ssoFailed(w http.ResponseWriter, r *http.Request, err error, op string) {
	slog.Warn("webphone: oidc login failed", "op", op, "error", err,
		"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err))
	h.clearSSOCookie(w, ssoStateCookie)
	h.clearSSOCookie(w, ssoNonceCookie)
	h.clearSSOCookie(w, ssoVerifierCookie)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/?sso_error="+ssoErrorCode(err), http.StatusSeeOther)
}

// ssoErrorCode maps a failure onto the island's small reason vocabulary:
// the browser gets an actionable sentence, never an internal error string.
func ssoErrorCode(err error) string {
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

func (h *handlers) setSSOCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // HttpOnly + Secure (wired policy) + SameSite=Lax + path-scoped: the attempt secrets never reach scripts or another site
		Name:     name,
		Value:    value,
		Path:     ssoCookiePath,
		MaxAge:   ssoCookieTTL,
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *handlers) readSSOCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return "", errorfamily.Newf(errorfamily.Rejection, "sso.attempt.missing",
			"this sign-in attempt has no live %s cookie — restart the sign-in from this browser", name)
	}
	return cookie.Value, nil
}

func (h *handlers) clearSSOCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // deletion mirrors setSSOCookie's attributes; a Mismatched Secure attribute would leave the cookie behind
		Name:     name,
		Value:    "",
		Path:     ssoCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}
