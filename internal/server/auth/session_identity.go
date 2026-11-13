package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sipcred"
	"github.com/larsartmann/webphone/internal/userauth"
)

// agentModeProbeTimeout bounds the best-effort draft-mode decoration of
// the session identity (the nav pill's data, M13). The agent is
// loopback; a slow or dead seam degrades to NO pill, never to a broken
// login (the SSO boot doctrine: wrong config fails boot, an
// unreachable peer only warns).
const agentModeProbeTimeout = 3 * time.Second

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
