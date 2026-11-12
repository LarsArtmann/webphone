// Package auth holds BOTH login modes and their shared spine, carved out
// of the server wiring: the extension+SIP-password login, the passkey
// (WebAuthn) ceremony and enrollment, the OIDC (Pocket ID) dance, the
// library-owned /login page, the Settings passkey management, and the
// session endpoints every mode resumes through (create/destroy/whoami/
// switch). The 2026-10-17 adjudication kept session_api in the server
// package because it was the shared spine of one login mode; the
// 2026-11-11 login-page front-door train made auth a real sub-surface
// (library page + ExtraSection lanes + ceremony endpoints), which fired
// the row's reopen trigger. The HTTP wiring (mux, flood budgets, CSRF
// middleware, nonce lane) stays in internal/server; these handlers reach
// the session gate and the HTML render plumbing through the narrow Shell
// seam, so the 401 wording and the render helpers keep their one home
// there.
package auth

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"net/http"

	loginpage "github.com/larsartmann/cqrs-htmx/loginpage/v4"
	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sso"
	"github.com/larsartmann/webphone/internal/textagent"
	"github.com/larsartmann/webphone/internal/userauth"
	"github.com/larsartmann/webphone/internal/web/views"
)

// MaxPasskeyBodySize bounds the ceremony bodies (attestation/assertion
// JSON): a few KB of base64; 64 KiB leaves generous headroom while
// keeping a hostile POST out of the parser. The server's route table
// mounts httputil.MaxBodySize(MaxPasskeyBodySize) on the finish lanes.
const MaxPasskeyBodySize = 64 << 10

// The three short-lived cookies carrying one OIDC authorization attempt.
// They are HttpOnly (no script needs them), SameSite=Lax (the IdP
// redirect is a top-level GET navigation, so they ride it while
// cross-site POSTs cannot trigger the dance), Path-scoped to the callback
// and expired as soon as the callback consumes them — a replay finds
// nothing to replay.
const (
	SSOStateCookie    = "wp_sso_state"
	SSONonceCookie    = "wp_sso_nonce"
	SSOVerifierCookie = "wp_sso_verifier"
	SSOCallbackPath   = "/api/auth/oauth2"
	ssoCookieTTL      = 600
)

// SSOClient is the OIDC login seam the auth surfaces drive: one challenge
// (the IdP URL plus the state/nonce/PKCE values the browser must echo
// back) and one completion (code → mapped identity + SIP directory
// password). The concrete *sso.Client is the only production
// implementation; tests supply a stub, which is why the field is an
// interface.
type SSOClient interface {
	Challenge(ctx context.Context) (sso.Challenge, error)
	Complete(ctx context.Context, code, verifier, nonce string) (authmap.User, string, error)
}

// Deps is the auth surfaces' narrowed dependency view: exactly what the
// login modes, the ceremonies and the session spine read.
type Deps struct {
	Config    config.Config
	Sessions  session.Store
	UserAuth  *userauth.Service
	SSO       SSOClient
	PhoneAPI  *pbx.Client
	TextAgent *textagent.Client
}

// Shell is the narrow wiring seam the carved handlers ride on,
// implemented by the server's handlers hub: the session gate (the ONE
// 401 writer for missing sessions) and the HTML render plumbing the two
// HTML surfaces borrow (the settings partial re-render and the styled
// 404). Exported methods so internal/server can implement it.
type Shell interface {
	// RequireSession gates a handler on the attached session; on
	// failure it has already written the 401.
	RequireSession(w http.ResponseWriter, r *http.Request) (session.Session, bool)
	// Partial re-renders one tab's partial (the HX response body).
	Partial(w http.ResponseWriter, r *http.Request, tab views.Tab)
	// Lang resolves the request's UI language.
	Lang(r *http.Request) views.Lang
	// SafeDetail logs the honest detail and returns the safe operator
	// line.
	SafeDetail(r *http.Request, op string, err error) string
	// NotFound renders the styled 404 (or the htmx one-liner).
	NotFound(w http.ResponseWriter, r *http.Request)
	// IdentityFor returns the extension's presented PSTN number.
	IdentityFor(extension domain.Extension) string
	// NotifyToast pushes one toast onto the response (HX-Trigger).
	NotifyToast(w http.ResponseWriter, kind, msg string)
}

// Handlers is the auth hub: every file in this package hangs methods off
// it. The server's wiring mounts the routes; the Shell seam is the
// server's handlers hub.
type Handlers struct {
	deps  Deps
	shell Shell
	// cookies is the wired session-cookie policy (Secure mirrors the
	// TLS-fronted deployment; r.TLS is nil behind the proxy).
	cookies session.CookiePolicy
	// csrfCfg is the REAL CSRF config — login/logout cookie invalidation
	// must carry the same Secure/Path attributes as the cookie it deletes
	// (strict browsers reject a non-Secure deletion of a Secure cookie).
	csrfCfg httputil.CSRFConfig
	// loginPage is the library-owned passkey login page
	// (cqrs-htmx/loginpage), built ONLY when the identity layer is wired;
	// nil keeps /login on the styled 404 (mode off, or the page build
	// failed loudly at boot).
	loginPage *loginpage.Handler
}

// New wires the auth hub. The server's New builds the narrowed Deps, its
// own Shell implementation, the cookie policies and the library login
// page, then mounts the routes through the returned hub.
func New(deps Deps, shell Shell, cookies session.CookiePolicy, csrfCfg httputil.CSRFConfig, loginPage *loginpage.Handler) *Handlers {
	return &Handlers{deps: deps, shell: shell, cookies: cookies, csrfCfg: csrfCfg, loginPage: loginPage}
}

// NewLoginPage builds the library login page over the embedded identity
// layer (when wired) plus webphone's own sign-in section. Registration
// stays HIDDEN: enrollment is token-gated by design (CLI-minted,
// operator-provisioned; the /enroll page owns that flow) — a phone system
// never grows self-service accounts. The page carries no library OAuth2
// buttons: the OIDC seam is webphone's own lane, hosted as part of the
// ExtraSection, so each sign-in method keeps exactly one door.
func NewLoginPage(deps Deps) (*loginpage.Handler, error) {
	var svc *usermgmt.Service
	if deps.UserAuth != nil {
		svc = deps.UserAuth.Usermgmt()
	}
	passkeyOn := deps.Config.Auth.Passkey.Enabled()
	return loginpage.New(loginpage.Config{
		Service: svc,
		Title:   "Sign in",
		Brand:   "WebPhone",
		// A live session post-login lands on the shell, where the boot
		// probe resumes and registers.
		Redirect: "/",
		// The scoped Tailwind build carrying the app tokens + the
		// loginpage/templ-components classes (see login.css.input).
		CSSPath: "/assets/login.css",
		// The ceremony endpoints at the usermgmt-canonical paths the
		// library's JS posts to: /api/auth/webauthn/login/{begin,finish}.
		AuthPrefix: "/api",
		// The nonce the CSP override pins onto the page's inline
		// scripts (ceremony JS + config JSON + the theme preload). The
		// ExtraSection's script is same-origin external ('self'), no
		// nonce needed.
		NonceFromRequest: httputil.NonceFromRequest,
		NoRegistration:   true,
		// webphone's own sign-in lanes: extension form always (the
		// break-glass lifeline under the passkey ceremony, the primary
		// method when the identity layer is off), Pocket ID above it
		// when the OIDC seam is wired.
		ExtraSection: views.LoginPageExtra(!passkeyOn, deps.SSO != nil),
	})
}

// AuthAudit is the ONE home of the semantic auth events (A3): one
// greppable slog line per login, denied login, logout, passkey
// enrollment and passkey revocation — the operator's audit trail over
// the journal. It logs IDENTIFIERS only (extension, email, credential
// name); never a password, a session token, a derived credential, or a
// ceremony body (the request log middleware never sees those either —
// bodies stay unparsed). Exported: the agent-action branch (agentAudit)
// stays in the server package and the audit vocabulary is one contract.
func AuthAudit(event string, key, value, outcome string) {
	slog.Info("auth "+event, key, value, "outcome", outcome)
}

// writeJSON is the auth surfaces' JSON writer (the api package carries
// the same one-liner as writeJSONBody; the server package keeps its own
// for internal_directory — the carve precedent accepts the per-package
// writer so no cross-package unexported call survives).
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.MarshalWrite(w, payload) //nolint:erraudit // best-effort write; the response is already committed
}
