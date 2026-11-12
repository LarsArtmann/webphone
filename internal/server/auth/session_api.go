package auth

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sipcred"
	"github.com/larsartmann/webphone/internal/userauth"
)

// verifyTimeout bounds the PBX round-trip at login: the island awaits
// this POST before showing the signed-in state, and a hung PBX must not
// hang logins for the full client timeout.
const verifyTimeout = 5 * time.Second

// agentModeProbeTimeout bounds the best-effort draft-mode decoration of
// the session identity (the nav pill's data, M13). The agent is
// loopback; a slow or dead seam degrades to NO pill, never to a broken
// login (the SSO boot doctrine: wrong config fails boot, an
// unreachable peer only warns).
const agentModeProbeTimeout = 3 * time.Second

// LoginRequest is the POST /api/session wire shape: the extension and
// its PBX directory password. Named (not inline) so the machine-API
// contract derives its schema from this type (internal/server/apidocs.go).
type LoginRequest struct {
	Extension string `description:"The PBX extension to sign in, e.g. 1001." json:"extension"`
	Password  string `description:"The extension's PBX directory password." json:"password"`
}

// SwitchRequest is the POST /api/session/switch wire shape: another
// extension mapped to the signed-in user (D1, the line switcher). The
// server resolves the mapping itself — the client's suggestion is
// never authority for what a session may become.
type SwitchRequest struct {
	Extension string `description:"A mapped extension of the signed-in user to re-mint the session for." json:"extension"`
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

// SwitchExtension re-mints the session for ANOTHER extension mapped to
// the same human (D1, the line switcher): a multi-line operator moves
// between their lines without a second ceremony. The authorization is
// the session itself — only a mapped login (passkey/OIDC, LoginEmail
// set) may switch, and only to an extension THAT email maps; the
// classic extension login authenticated exactly one line and stays
// pinned to it. The target's directory password is sourced fresh from
// its password file and directory-verified (a stale file must not mint
// a dead line — the same fail-closed rule as login). The old session
// row dies with the switch; CSRF rotates (mintSession); the client
// reloads, and the boot probe re-registers, resubscribes and re-renders
// everything for the new line.
func (a *Handlers) SwitchExtension(w http.ResponseWriter, r *http.Request) {
	sess, ok := a.shell.RequireSession(w, r)
	if !ok {
		return
	}
	var body SwitchRequest
	if err := json.UnmarshalRead(r.Body, &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	target, err := domain.ParseExtension(body.Extension)
	if err != nil {
		http.Error(w, "invalid extension", http.StatusBadRequest)
		return
	}
	if sess.LoginEmail == "" {
		http.Error(w, "switching lines requires a mapped login", http.StatusForbidden)
		return
	}
	mapped, ok := a.deps.UserAuth.Resolve(sess.LoginEmail)
	if !ok || !slices.Contains(mapped.Extensions, target) {
		// One answer for an unmapped account and a not-mapped target
		// alike: the switch surface never enumerates which lines exist.
		http.Error(w, "extension not mapped to this account", http.StatusForbidden)
		return
	}
	if target == sess.Extension {
		// Idempotent no-op: answer the current identity without a mint,
		// exactly what a resumed boot probe would say.
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, a.sessionIdentityResponse(r.Context(), sess.Extension, sess.LoginEmail))
		return
	}
	password, err := a.deps.UserAuth.SIPPassword(target)
	if err != nil {
		slog.Warn("webphone: line switch could not source the target credentials", "error", err,
			"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err), "extension", target.String())
		http.Error(w, "target line credentials unavailable", http.StatusServiceUnavailable)
		return
	}
	if !a.verifyDirectoryCredentials(w, r, target, password) {
		return
	}
	oldToken := session.TokenFromRequest(r)
	token, err := a.mintSession(w, r, target, password, sess.LoginEmail)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	// The old row dies only after the mint succeeded: a failed switch
	// must leave the operator on their working line.
	a.deps.Sessions.Delete(oldToken)
	AuthAudit("switch", "email", sess.LoginEmail, sess.Extension.String()+" -> "+target.String())
	writeJSON(w, http.StatusOK, a.sessionAnswer(w, r.Context(), token, target, sess.LoginEmail))
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
func (a *Handlers) verifyDirectoryCredentials(w http.ResponseWriter, r *http.Request, extension domain.Extension, password string) bool {
	if !a.deps.PhoneAPI.Enabled() {
		return true
	}
	ctx, cancel := context.WithTimeout(r.Context(), verifyTimeout)
	err := a.deps.PhoneAPI.VerifyCredentials(ctx, pbx.Credentials{
		Extension: extension.String(),
		Password:  password,
	})
	cancel()
	switch {
	case err == nil:
		return true
	case errors.Is(err, pbx.ErrUnauthorized):
		AuthAudit("login_denied", "extension", extension.String(), "directory rejected the credentials")
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
// step because the served page carries the old masked token. Returns
// the minted token (the derived-credential input).
func (a *Handlers) mintSession(w http.ResponseWriter, r *http.Request, extension domain.Extension, password, loginEmail string) (session.Token, error) {
	token, err := a.deps.Sessions.Create(extension, password, loginEmail)
	if err != nil {
		return session.Token{}, err //nolint:wrapcheck // classified by the store seam; the caller answers 500
	}
	session.SetCookie(w, r, token, a.deps.Config.SessionTTL, a.cookies)
	httputil.InvalidateCSRFCookie(w, a.csrfCfg)
	return token, nil
}

// sessionAnswer builds the post-mint credential answer — the ONE home
// of the derived-credential policy: with brokering on, the response
// carries the session-derived REGISTER credential (the directory
// password never reaches the browser) under no-store; with it off, the
// legacy wire stays silent HERE, and the login surfaces that must hand
// back the file-sourced password (passkey finish, session resume) add
// their credential plus the no-store that credential demands at their
// own site — the policy lives once, the real divergences stay visible
// where they diverge.
func (a *Handlers) sessionAnswer(
	w http.ResponseWriter, ctx context.Context, token session.Token, extension domain.Extension, loginEmail string,
) SessionIdentity {
	response := a.sessionIdentityResponse(ctx, extension, loginEmail)
	if a.deps.Config.SIP.DerivedCredentials() {
		response.SIPCredential = sipcred.Derive(a.deps.Config.SIP.CredentialSecret, token.String())
		w.Header().Set("Cache-Control", "no-store")
	}
	return response
}

// SessionIdentity is the typed wire shape of what a freshly minted
// (or resumed) session tells the client about itself: the extension (the
// island's SIP REGISTER needs it), its presented DID when configured
// (identities), and — when the extension is passkey-mapped — the user's
// display name and every number mapped to that user, so the whoami line
// can lead with the human identity instead of a bare extension@sip_domain.
// Exactly ONE credential field rides an answer: `sip_credential` (the
// session-derived REGISTER credential) when credential brokering is on —
// the directory password never reaches the browser — and `password` (the
// legacy wire, the session's own payload) only when it is off, keeping
// unconfigured deployments byte-identical. The omitempty tags keep the
// wire identical to the historical map form.
// brandid-lint:ignore(BD003) wire DTO: config-validated DID rendered verbatim, no id ever flows back in (strong-id ruling)
type SessionIdentity struct {
	Extension        string   `description:"The signed-in extension." json:"extension"`
	DID              string   `description:"Presented PSTN number (config identities), when known." json:"did,omitempty"` //nolint:branching-flow // wire DTO: the DID is a config-validated identity rendered verbatim (omitempty); no id ever flows back IN through this field
	DisplayName      string   `description:"The passkey user's display name, when the extension is passkey-mapped." json:"display_name,omitempty"`
	Numbers          []string `description:"Every number mapped to the passkey user, deduplicated." json:"numbers,omitempty"`
	Extensions       []string `description:"Every extension mapped to the signed-in user (mapped logins only); more than one renders the line switcher." json:"extensions,omitempty"` //nolint:branching-flow // wire DTO: config-derived extension list rendered verbatim (omitempty); the switch endpoint re-validates membership server-side
	Password         string   `description:"Legacy wire only (credential brokering off): the session's directory password for the browser REGISTER." json:"password,omitempty"`
	SIPCredential    string   `description:"Session-derived REGISTER credential (credential brokering on; the directory password never reaches the browser)." json:"sip_credential,omitempty"`
	AgentMode        string   `description:"The textagent's effective draft mode (agent seam on only; empty hides the nav pill)." json:"agent_mode,omitempty"`
	AgentModeSource  string   `description:"Where the draft mode comes from: config or a runtime pin." json:"agent_mode_source,omitempty"`
	AgentDefaultMode string   `description:"The configured draft mode a reset restores." json:"agent_default_mode,omitempty"`
	AgentModes       []string `description:"The draft modes the agent supports, in cycle order." json:"agent_modes,omitempty"`
}

// sessionIdentityResponse shapes what a freshly minted (or resumed)
// session tells the client about itself (see SessionIdentity).
// loginEmail is the session's authenticated identity when it has one
// (passkey/OIDC): the whoami resolution prefers THAT user — on a shared
// extension the session must name its own human, not a map-order guess
// (the MappedByExtension fallback stays deterministic but cannot know
// which person is signed in). Empty = the classic extension login.
func (a *Handlers) sessionIdentityResponse(ctx context.Context, extension domain.Extension, loginEmail string) SessionIdentity {
	response := SessionIdentity{Extension: extension.String()}
	response.DID = a.shell.IdentityFor(extension)
	if a.deps.UserAuth != nil {
		if mapped, ok := a.deps.UserAuth.MappedByLogin(extension, loginEmail); ok {
			response.DisplayName = mapped.DisplayName
			response.Numbers = a.numbersFor(mapped)
			for _, ext := range mapped.Extensions {
				response.Extensions = append(response.Extensions, ext.String())
			}
		}
	}
	a.fillAgentMode(ctx, &response)
	return response
}

// fillAgentMode decorates the identity with the textagent's draft mode
// (the nav pill's data, M13). Best-effort by contract: the agent is
// loopback, the ask is bounded, and a failure leaves the fields empty —
// the shell hides the pill rather than lying about an autonomy state it
// could not see.
func (a *Handlers) fillAgentMode(ctx context.Context, response *SessionIdentity) {
	if a.deps.TextAgent == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, agentModeProbeTimeout)
	defer cancel()
	reviews, err := a.deps.TextAgent.Reviews(ctx)
	if err != nil {
		slog.Warn("agent mode probe failed; the nav pill stays hidden", "error", err)
		return
	}
	response.AgentMode = reviews.Mode
	response.AgentModeSource = reviews.ModeSource
	response.AgentDefaultMode = reviews.DefaultMode
	response.AgentModes = reviews.Modes
}

// numbersFor lists the mapped user's presented numbers (identities of
// every mapped extension), deduplicated in mapping order — the "numbers
// available to this user" the whoami line leads with.
func (a *Handlers) numbersFor(mapped userauth.MappedUser) []string {
	seen := make(map[string]bool, len(mapped.Extensions))
	numbers := make([]string, 0, len(mapped.Extensions))
	for _, ext := range mapped.Extensions {
		did := a.shell.IdentityFor(ext)
		if did == "" || seen[did] {
			continue
		}
		seen[did] = true
		numbers = append(numbers, did)
	}
	return numbers
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
