// Package handoff is the native-shell sign-in handoff: the half of the
// contract that needs no HTTP. The server half's HTTP surface is the
// auth package's sso_handoff.go; the consumer is the Android call shell
// (~/projects/webphone-android), whose cross-repo note is
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
package handoff

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// Android is the one native shell the handoff speaks today. Every other
// name is refused at the door: an unvalidated name must never reach a
// return URL.
const Android = "android"

// The wire names. Each is a cross-repo contract (the shell's
// SignInHandoff and SessionApi mirror them).
const (
	// AndroidFinishURL is the Android shell's return address: the
	// private-use URI the callback bounces its ticket off and Android
	// launches the shell with. RFC 8252 §7.1 wants a domain the app's
	// author controls, reversed; the host and path are pinned so nothing
	// but the handoff enters through it. PAIRED with the shell's manifest
	// VIEW filter: change both or neither.
	AndroidFinishURL = "cloud.larsartmann.webphone://auth/finish"
	// ShellHeader names the redeeming shell. Required on every
	// redemption: it is the browser's fence (see the package comment).
	ShellHeader = "X-WebPhone-Shell"
	// ShellProofHeader carries the proof of the attempt the ticket was
	// issued for.
	ShellProofHeader = "X-WebPhone-Shell-Proof"
	// ShellCookie marks ONE OIDC attempt as a shell handoff and binds it
	// (the auth package sets it with the attempt cookies' attributes).
	ShellCookie = "wp_sso_shell"

	ShellParam    = "shell"
	AppStateParam = "app_state"
	ProofParam    = "proof"
	TicketParam   = "ticket"
)

// secretShape is the alphabet of every secret the handoff passes through
// a URL or a header: the nonce, the proof and the ticket (32 random
// bytes in unpadded base64url are 43 characters; the ceiling only bounds
// a hostile value before it is hashed or compared).
var secretShape = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

const (
	// ticketTTL is how long a parked identity waits for its shell. The
	// shell redeems within a second or two of the bounce; the rest is
	// patience for a chooser dialog or a slow device.
	ticketTTL = 5 * time.Minute
	// ledgerCap bounds the ledger against a flood. Every entry costs a
	// completed IdP login, so a server ever reaching it is already under
	// a storm the login limiter is fencing; the cap keeps that storm out
	// of memory regardless.
	ledgerCap = 256
	// ticketBytes is the ticket's entropy: 256 bits.
	ticketBytes = 32
)

// Marker is what the shell cookie binds one attempt to: the nonce the
// return link must echo and the digest of the proof its ticket will
// demand.
type Marker struct {
	AppState    string
	ProofDigest [sha256.Size]byte
}

// CookieValue encodes the marker for the shell cookie.
func (m Marker) CookieValue() string {
	return m.AppState + "." + base64.RawURLEncoding.EncodeToString(m.ProofDigest[:])
}

// ParseMarker is the cookie's parse door: false for every value this
// server did not write, so a malformed marker can never mint a return
// link.
func ParseMarker(value string) (Marker, bool) {
	nonce, digest, found := strings.Cut(value, ".")
	if !found || !secretShape.MatchString(nonce) {
		return Marker{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(digest)
	if err != nil || len(raw) != sha256.Size {
		return Marker{}, false
	}
	return Marker{AppState: nonce, ProofDigest: [sha256.Size]byte(raw)}, true
}

// ProofDigest is the one home of the proof's digest: the begin hop and
// the redemption must agree on it byte for byte, and the raw proof is
// never kept.
func ProofDigest(proof string) [sha256.Size]byte {
	return sha256.Sum256([]byte(proof))
}

// ParseAttempt validates the begin hop's shell form BEFORE any IdP round
// trip or cookie is spent: the shell must be one this server can hand a
// sign-in to, and the nonce and the proof must have the shape the shell
// generates (anything else is a broken or hostile client, never a user).
// Each refusal is a coded error so the journal names the exact reason.
func ParseAttempt(query url.Values) (Marker, error) {
	if query.Get(ShellParam) != Android {
		return Marker{}, errorfamily.NewRejection("sso.handoff.request",
			"this server hands a finished sign-in only to the android shell")
	}
	nonce := query.Get(AppStateParam)
	if !secretShape.MatchString(nonce) {
		return Marker{}, errorfamily.NewRejection("sso.handoff.request",
			"an android shell sign-in needs an app_state of 16 to 128 characters from A-Za-z0-9_- (the nonce the return link echoes)")
	}
	proof := query.Get(ProofParam)
	if !secretShape.MatchString(proof) {
		return Marker{}, errorfamily.NewRejection("sso.handoff.request",
			"an android shell sign-in needs a proof of 16 to 128 characters from A-Za-z0-9_- (the secret its ticket is bound to)")
	}
	return Marker{AppState: nonce, ProofDigest: ProofDigest(proof)}, nil
}

// ReturnURL is the redirect the callback answers on a shell attempt: the
// shell's return address carrying the ticket and the nonce the shell
// requires the link to echo.
func ReturnURL(ticket, appState string) string {
	query := url.Values{AppStateParam: {appState}, TicketParam: {ticket}}
	return AndroidFinishURL + "?" + query.Encode()
}

// Identity is a sign-in the IdP already vouched for and the directory
// already verified, waiting for its shell: exactly what a session mint
// takes, and nothing else.
type Identity struct {
	Extension domain.Extension
	Password  string
	Email     string
}

// Verdict is what one redemption attempt found.
type Verdict int

const (
	// Redeemed: the ticket was live, its proof matched, and it is gone
	// now.
	Redeemed Verdict = iota
	// UnknownTicket: no live ticket of that name (never issued, expired
	// or already redeemed).
	UnknownTicket
	// ForeignProof: the ticket is live but the proof is not its own. The
	// ticket stays alive.
	ForeignProof
)

// parked is one live ticket: the identity it redeems, the proof digest it
// is bound to, and when it dies.
type parked struct {
	identity Identity
	proof    [sha256.Size]byte
	expires  time.Time
}

// Tickets is the ledger of live handoff tickets. In memory, per process,
// bounded (see the package comment for why nothing persists).
type Tickets struct {
	mu   sync.Mutex
	live map[[sha256.Size]byte]parked
	// Now is the ledger's clock, so the pins drive expiry without
	// sleeping.
	Now func() time.Time
}

// NewTickets builds an empty ledger on the wall clock.
func NewTickets() *Tickets {
	return &Tickets{live: make(map[[sha256.Size]byte]parked), Now: time.Now}
}

// Issue parks one verified identity under a fresh single-use ticket
// bound to the proof digest, and returns the ticket. The ledger indexes
// by the ticket's hash: the raw ticket exists in exactly one place, the
// return link.
func (t *Tickets) Issue(identity Identity, proofDigest [sha256.Size]byte) (string, error) {
	raw := make([]byte, ticketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "sso.handoff.entropy", "generate handoff ticket entropy") //nolint:erraudit // crypto/rand takes no ctx; the wrap owns the code
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sweepLocked()
	if len(t.live) >= ledgerCap {
		t.evictOldestLocked()
	}
	t.live[sha256.Sum256([]byte(ticket))] = parked{
		identity: identity,
		proof:    proofDigest,
		expires:  t.Now().Add(ticketTTL),
	}
	return ticket, nil
}

// Redeem is the one place a ticket turns into its identity. A live
// ticket redeemed with its own proof is consumed; anything else answers
// without touching the ledger's other entries, and a foreign proof never
// burns the ticket (a holder without the proof must not be able to void
// the sign-in it is trying to steal).
func (t *Tickets) Redeem(ticket string, proofDigest [sha256.Size]byte) (Identity, Verdict) {
	if !secretShape.MatchString(ticket) {
		return Identity{}, UnknownTicket
	}
	key := sha256.Sum256([]byte(ticket))
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, found := t.live[key]
	if !found {
		return Identity{}, UnknownTicket
	}
	if !t.Now().Before(entry.expires) {
		delete(t.live, key)
		return Identity{}, UnknownTicket
	}
	if subtle.ConstantTimeCompare(entry.proof[:], proofDigest[:]) != 1 {
		return Identity{}, ForeignProof
	}
	delete(t.live, key)
	return entry.identity, Redeemed
}

// sweepLocked drops the dead entries. The ledger is bounded, so a full
// pass on every issue is the whole garbage collection it needs (no
// janitor goroutine to leak or to forget).
func (t *Tickets) sweepLocked() {
	now := t.Now()
	for key, entry := range t.live {
		if !now.Before(entry.expires) {
			delete(t.live, key)
		}
	}
}

// evictOldestLocked makes room at the cap by dropping the entry that
// dies first. Every entry is born with the same TTL, so that is the
// oldest one; the caller holds the lock.
func (t *Tickets) evictOldestLocked() {
	var (
		oldestKey [sha256.Size]byte
		oldest    time.Time
	)
	for key, entry := range t.live {
		if oldest.IsZero() || entry.expires.Before(oldest) {
			oldestKey, oldest = key, entry.expires
		}
	}
	delete(t.live, oldestKey)
}
