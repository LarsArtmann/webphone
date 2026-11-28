package auth

// The native-shell sign-in handoff: the server half of the contract the
// Android call shell (~/projects/webphone-android) redeems a Pocket ID
// login through. The cross-repo note is the shell's
// docs/drafts/2026-11-28_webphone-oidc-shell-handoff.md.
//
// WHY a handoff at all. The OAuth dance cannot complete inside the
// shell's WebView: RFC 8252 §8.12 forbids embedded user-agents for
// authorization requests, and this IdP makes the rule bite. Pocket ID is
// passkey-ONLY and accepts exactly its own App URL as a WebAuthn origin
// (RPOrigins in its backend/internal/webauthn/service.go; the request to
// relax it, pocket-id/pocket-id#1845, closed NOT_PLANNED on 2026-10-15:
// "open the login page in an external browser"). Android's FOR_APP
// WebAuthn signs a WebView ceremony with the APP's origin
// (android:apk-key-hash:...), so no assetlinks file bridges the gap. The
// system browser can, because a browser's WebAuthn origin IS the web
// origin.
//
// So the shell hands the dance to the system browser, which then holds
// the finished sign-in in ITS cookie jar, where the shell cannot read it.
// The handoff carries the sign-in across in four hops:
//
//	GET /api/auth/oauth2/shell
//	    204 when this server speaks the contract (the styled 404 when the
//	    OIDC seam is off, which is also what an older server answers)
//	GET /api/auth/oauth2/begin?shell=android&app_state=<nonce>&proof=<secret>
//	    302 to the IdP, the attempt cookies plus the shell marker; 400 for
//	    an unknown shell or a malformed nonce or proof
//	GET /api/auth/oauth2/callback?code&state        (the browser, cookies)
//	    303 cloud.larsartmann.webphone://auth/finish?app_state=<nonce>&ticket=<ticket>
//	    with NO session cookie and every attempt cookie expired
//	GET /api/auth/oauth2/finish?ticket=<ticket>
//	    X-WebPhone-Shell: android
//	    X-WebPhone-Shell-Proof: <secret>
//	    303 / plus the session cookie (the ticket dies here); 303
//	    /login?sso_error=stale for an unknown, expired or used ticket
//
// A begin hop WITHOUT the shell parameter is the plain browser lane and
// stays byte-identical to the pre-handoff server.
//
// THREE SECRETS, THREE FENCES. A ticket is a bearer credential: whoever
// presents one with its proof redeems the sign-in it was issued for. The
// secrets fence the three ways this can go wrong.
//
//   - The nonce (app_state) is the SHELL's fence against a forged return
//     link. The shell opens every attempt with a fresh nonce and drops
//     any return link that does not echo it, before any request: a link
//     carrying an attacker's own ticket (anyone can finish a handoff
//     dance for their own account) would otherwise sign the shell into
//     that account, and would void the user's open attempt.
//   - The proof is the ticket's BINDING to the one shell attempt that
//     started it. A private-use scheme is claimable: any app on the
//     device can register for it (RFC 8252 §7.2), so a rogue app that
//     catches the return link holds the ticket. It never holds the
//     proof, which rides only the begin hop (over the stack's own TLS)
//     and the redemption. The ledger keeps the proof's digest; a
//     redemption without the matching proof is refused and leaves the
//     ticket alive for its rightful shell.
//   - The X-WebPhone-Shell header is the BROWSER's fence. A navigation,
//     an image or a form can carry no custom header, so a cross-site
//     page cannot spend a ticket into a browser's cookie jar (the
//     login-CSRF shape the callback's state cookie closes for its own
//     lane).
//
// The session is BORN at the redemption, not at the callback. An
// abandoned handoff (the user never returns, the shell is gone) then
// leaves a dead ledger entry, never a live session row that nobody can
// reach and nothing reaps. The ledger is in memory and per process by
// the same reasoning: a restart drops every pending handoff, the shell
// lands on the stack's own "attempt expired" sentence and the user signs
// in again, which costs less than persisting a bearer credential for a
// finished sign-in.

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// AndroidShellFinishURL is the Android shell's return address: the
// private-use URI the callback bounces its ticket off and Android
// launches the shell with. RFC 8252 §7.1 wants a domain the app's author
// controls, reversed; the host and path are pinned so nothing but the
// handoff enters through it. PAIRED with the shell's manifest VIEW
// filter: change both or neither.
const AndroidShellFinishURL = "cloud.larsartmann.webphone://auth/finish"

// The wire names of the redemption and of the attempt's shell marker.
const (
	// ShellHeader names the redeeming shell. Required on every
	// redemption: it is the browser's fence (see the file comment).
	ShellHeader = "X-WebPhone-Shell"
	// ShellProofHeader carries the proof of the attempt the ticket was
	// issued for.
	ShellProofHeader = "X-WebPhone-Shell-Proof"
	// SSOShellCookie marks ONE OIDC attempt as a shell handoff and binds
	// it: the value is the shell's nonce (which the return link echoes)
	// and the proof's digest, separated by a dot. It carries the attempt
	// cookies' attributes and dies with them.
	SSOShellCookie = "wp_sso_shell"
)

// The query-parameter names of the handoff. Each is a cross-repo
// contract (the shell's SignInHandoff and SessionApi mirror them).
const (
	shellParam    = "shell"
	appStateParam = "app_state"
	proofParam    = "proof"
	ticketParam   = "ticket"
)

// shellAndroid is the one native shell the handoff speaks today. Every
// other value is refused at the door: an unvalidated name must never
// reach a return URL.
const shellAndroid = "android"

// shellSecretShape is the alphabet of every secret the handoff passes
// through a URL or a header: the nonce, the proof and the ticket (32
// random bytes in unpadded base64url are 43 characters; the ceiling only
// bounds a hostile value before it is hashed or compared).
var shellSecretShape = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

const (
	// shellHandoffTTL is how long a parked identity waits for its shell.
	// The shell redeems within a second or two of the bounce; the rest is
	// patience for a chooser dialog or a slow device.
	shellHandoffTTL = 5 * time.Minute
	// shellHandoffCap bounds the ledger against a flood. Every entry
	// costs a completed IdP login, so a server ever reaching it is
	// already under a storm the login limiter is fencing; the cap keeps
	// that storm out of memory regardless.
	shellHandoffCap = 256
	// shellTicketBytes is the ticket's entropy: 256 bits.
	shellTicketBytes = 32
)

// shellMarker is what the shell cookie binds one attempt to: the nonce
// the return link must echo and the digest of the proof its ticket will
// demand.
type shellMarker struct {
	appState string
	proof    [sha256.Size]byte
}

// cookieValue encodes the marker for the shell cookie.
func (m shellMarker) cookieValue() string {
	return m.appState + "." + base64.RawURLEncoding.EncodeToString(m.proof[:])
}

// parseShellMarker is the cookie's parse door: false for every value this
// server did not write, so a malformed marker can never mint a return
// link.
func parseShellMarker(value string) (shellMarker, bool) {
	nonce, digest, found := strings.Cut(value, ".")
	if !found || !shellSecretShape.MatchString(nonce) {
		return shellMarker{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(digest)
	if err != nil || len(raw) != sha256.Size {
		return shellMarker{}, false
	}
	return shellMarker{appState: nonce, proof: [sha256.Size]byte(raw)}, true
}

// shellAttempt validates the begin hop's shell form BEFORE any IdP round
// trip or cookie is spent: the shell must be one this server can hand a
// sign-in to, and the nonce and the proof must have the shape the shell
// generates (anything else is a broken or hostile client, never a user).
// Each refusal is a coded error so the journal names the exact reason.
func shellAttempt(query url.Values) (shellMarker, error) {
	if query.Get(shellParam) != shellAndroid {
		return shellMarker{}, errorfamily.NewRejection("sso.handoff.request",
			"this server hands a finished sign-in only to the android shell")
	}
	nonce := query.Get(appStateParam)
	if !shellSecretShape.MatchString(nonce) {
		return shellMarker{}, errorfamily.NewRejection("sso.handoff.request",
			"an android shell sign-in needs an app_state of 16 to 128 characters from A-Za-z0-9_- (the nonce the return link echoes)")
	}
	proof := query.Get(proofParam)
	if !shellSecretShape.MatchString(proof) {
		return shellMarker{}, errorfamily.NewRejection("sso.handoff.request",
			"an android shell sign-in needs a proof of 16 to 128 characters from A-Za-z0-9_- (the secret its ticket is bound to)")
	}
	return shellMarker{appState: nonce, proof: sha256.Sum256([]byte(proof))}, nil
}

// shellReturnURL is the redirect the callback answers on a shell
// attempt: the shell's return address carrying the ticket and the nonce
// the shell requires the link to echo.
func shellReturnURL(ticket, appState string) string {
	query := url.Values{appStateParam: {appState}, ticketParam: {ticket}}
	return AndroidShellFinishURL + "?" + query.Encode()
}

// handoffIdentity is a sign-in the IdP already vouched for and the
// directory already verified, waiting for its shell: exactly what
// mintSession takes, and nothing else.
type handoffIdentity struct {
	extension domain.Extension
	password  string
	email     string
}

// parkedHandoff is one live ticket: the identity it redeems, the proof
// digest it is bound to, and when it dies.
type parkedHandoff struct {
	identity handoffIdentity
	proof    [sha256.Size]byte
	expires  time.Time
}

// ticketVerdict is what one redemption attempt found.
type ticketVerdict int

const (
	// ticketRedeemed: the ticket was live, its proof matched, and it is
	// gone now.
	ticketRedeemed ticketVerdict = iota
	// ticketUnknown: no live ticket of that name (never issued, expired
	// or already redeemed).
	ticketUnknown
	// ticketForeign: the ticket is live but the proof is not its own.
	// The ticket stays alive.
	ticketForeign
)

// handoffTickets is the ledger of live handoff tickets. In memory, per
// process, bounded (see the file comment for why nothing persists).
type handoffTickets struct {
	mu   sync.Mutex
	live map[[sha256.Size]byte]parkedHandoff
	// now is the ledger's clock, so the pins drive expiry without
	// sleeping.
	now func() time.Time
}

func newHandoffTickets() *handoffTickets {
	return &handoffTickets{
		live: make(map[[sha256.Size]byte]parkedHandoff),
		now:  time.Now,
	}
}

// issue parks one verified identity under a fresh single-use ticket
// bound to the proof digest, and returns the ticket. The ledger indexes
// by the ticket's hash: the raw ticket exists in exactly one place, the
// return link.
func (l *handoffTickets) issue(identity handoffIdentity, proof [sha256.Size]byte) (string, error) {
	raw := make([]byte, shellTicketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "sso.handoff.entropy", "generate handoff ticket entropy") //nolint:erraudit // crypto/rand takes no ctx; the wrap owns the code
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepLocked()
	if len(l.live) >= shellHandoffCap {
		l.evictOldestLocked()
	}
	l.live[sha256.Sum256([]byte(ticket))] = parkedHandoff{
		identity: identity,
		proof:    proof,
		expires:  l.now().Add(shellHandoffTTL),
	}
	return ticket, nil
}

// redeem is the one place a ticket turns into its identity. A live
// ticket redeemed with its own proof is consumed; anything else answers
// without touching the ledger's other entries, and a foreign proof never
// burns the ticket (a holder without the proof must not be able to void
// the sign-in it is trying to steal).
func (l *handoffTickets) redeem(ticket string, proof [sha256.Size]byte) (handoffIdentity, ticketVerdict) {
	if !shellSecretShape.MatchString(ticket) {
		return handoffIdentity{}, ticketUnknown
	}
	key := sha256.Sum256([]byte(ticket))
	l.mu.Lock()
	defer l.mu.Unlock()
	parked, found := l.live[key]
	if !found {
		return handoffIdentity{}, ticketUnknown
	}
	if !l.now().Before(parked.expires) {
		delete(l.live, key)
		return handoffIdentity{}, ticketUnknown
	}
	if subtle.ConstantTimeCompare(parked.proof[:], proof[:]) != 1 {
		return handoffIdentity{}, ticketForeign
	}
	delete(l.live, key)
	return parked.identity, ticketRedeemed
}

// sweepLocked drops the dead entries. The ledger is bounded, so a full
// pass on every issue is the whole garbage collection it needs (no
// janitor goroutine to leak or to forget).
func (l *handoffTickets) sweepLocked() {
	now := l.now()
	for key, parked := range l.live {
		if !now.Before(parked.expires) {
			delete(l.live, key)
		}
	}
}

// evictOldestLocked makes room at the cap by dropping the entry that
// dies first. Every entry is born with the same TTL, so that is the
// oldest one; the caller holds the lock.
func (l *handoffTickets) evictOldestLocked() {
	var (
		oldestKey [sha256.Size]byte
		oldest    time.Time
	)
	for key, parked := range l.live {
		if oldest.IsZero() || parked.expires.Before(oldest) {
			oldestKey, oldest = key, parked.expires
		}
	}
	delete(l.live, oldestKey)
}

// SSOShellProbe answers whether this server speaks the native-shell
// sign-in handoff: 204 when the OIDC seam is on, the styled 404 when it
// is not, which is also what a server that predates the contract
// answers. One GET therefore tells a shell whether it may hand a sign-in
// to the browser at all, and a shell that never asked cannot be stranded
// in a lane where the browser signs in and the shell never does. The
// probe hands out nothing, and is never cached: a stack that upgrades
// must flip the shell's lane on its next ask.
func (a *Handlers) SSOShellProbe(w http.ResponseWriter, r *http.Request) {
	if a.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

// SSOFinish redeems a handoff ticket: the ONE request that turns the
// ticket the browser bounced back into a session cookie in the shell's
// own jar (the request is the shell's, not a navigation, so the cookie
// lands exactly there and nowhere else). The response is a redirect to
// the page the shell must land on: `/` signed in, or the login page
// carrying the stack's own refusal sentence for a ticket that is no
// longer live. The audit line lands here, where the session is born.
func (a *Handlers) SSOFinish(w http.ResponseWriter, r *http.Request) {
	if a.deps.SSO == nil {
		http.NotFound(w, r)
		return
	}
	// The browser's fence comes first: a request that names no shell is
	// a navigation or an image, and must not even learn whether the
	// ticket exists.
	if r.Header.Get(ShellHeader) != shellAndroid {
		a.handoffRefused(w, r, http.StatusBadRequest,
			"Sign-in refused: only the app that started this sign-in can finish it. Start the sign-in again from the app.",
			errorfamily.NewRejection("sso.handoff.request", "a handoff redemption must name its shell in the "+ShellHeader+" header"),
			"redeem handoff ticket")
		return
	}
	proof := sha256.Sum256([]byte(r.Header.Get(ShellProofHeader)))
	identity, verdict := a.handoffs.redeem(r.URL.Query().Get(ticketParam), proof)
	switch verdict {
	case ticketUnknown:
		a.ssoFailed(w, r, errorfamily.NewRejection("sso.handoff.ticket",
			"the handoff ticket is unknown, expired or already redeemed"), "redeem handoff ticket")
		return
	case ticketForeign:
		// Someone holds the ticket without the proof that never rode the
		// return link: that is an interception attempt, worth a line.
		a.handoffRefused(w, r, http.StatusForbidden,
			"Sign-in refused: this ticket belongs to another sign-in attempt. Start the sign-in again from the app.",
			errorfamily.NewRejection("sso.handoff.proof", "the handoff ticket was presented without the proof of the attempt that started it"),
			"redeem handoff ticket")
		return
	case ticketRedeemed:
	}
	if _, err := a.mintSession(w, r, identity.extension, identity.password, identity.email); err != nil {
		slog.Error("webphone: oidc handoff session mint failed", "error", err, "extension", identity.extension.String())
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	AuthAudit("login", "email", identity.email, "oidc verified; session minted through the android shell handoff")
	// The island's boot probe (GET /api/session) signs the shell in on
	// the landing page, exactly as the callback's 303 does for a browser.
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handoffRefused journals one refused handoff request with its classified
// reason (the line ssoFailed writes for the redirect refusals) and
// answers the caller's actionable sentence. The sentence is written at
// the call site on purpose: the body says what the client should do,
// while the coded error keeps the operator's detail in the journal.
func (a *Handlers) handoffRefused(w http.ResponseWriter, r *http.Request, status int, sentence string, err error, op string) {
	slog.Warn("webphone: oidc handoff refused", "op", op, "error", err,
		"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err))
	w.Header().Set("Cache-Control", "no-store")
	http.Error(w, sentence, status)
}
