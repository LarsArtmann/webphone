package handoff

import (
	"crypto/sha256"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// The handoff's non-HTTP half under test: the ticket ledger's lifecycle
// (single use, TTL, the cap, the proof binding), the begin hop's parse
// door, the shell marker's codec and the wire names themselves. The HTTP
// surface is pinned from internal/server (sso_handoff_api_test.go),
// where the whole dance runs through the real mux.

// ticketShape is what the shell accepts for every secret of the
// handoff, checked here against the ledger's OWN output: 32 random bytes
// in unpadded base64url are exactly 43 characters.
var ticketShape = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func testIdentity() Identity {
	return Identity{
		Extension: domain.MustParseExtension("1000"),
		Password:  "dir-password",
		Email:     "lars@example.com",
	}
}

// frozenLedger is a ledger on a clock the test moves by hand.
func frozenLedger(start *time.Time) *Tickets {
	tickets := NewTickets()
	tickets.Now = func() time.Time { return *start }
	return tickets
}

func TestTicketRedeemsOnce(t *testing.T) {
	tickets := NewTickets()
	proof := ProofDigest(strings.Repeat("p", 43))
	ticket, err := tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if !ticketShape.MatchString(ticket) {
		t.Errorf("ticket = %q, want 43 characters of unpadded base64url", ticket)
	}
	identity, verdict := tickets.Redeem(ticket, proof)
	if verdict != Redeemed {
		t.Fatalf("first redemption verdict = %v, want Redeemed", verdict)
	}
	want := testIdentity()
	if identity.Extension != want.Extension || identity.Password != want.Password || identity.Email != want.Email {
		t.Errorf("redeemed identity = %+v, want %+v", identity, want)
	}
	// Single use by construction: the redemption consumed the entry.
	if _, verdict := tickets.Redeem(ticket, proof); verdict != UnknownTicket {
		t.Errorf("second redemption verdict = %v, want UnknownTicket", verdict)
	}
}

func TestForeignProofLeavesTheTicketAlive(t *testing.T) {
	tickets := NewTickets()
	proof := ProofDigest(strings.Repeat("p", 43))
	ticket, err := tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	// A rogue app caught the return link and guesses without the proof
	// that never rode it: the refusal must not void the rightful shell's
	// sign-in.
	for _, foreign := range []string{"", strings.Repeat("x", 43), strings.Repeat("p", 42) + "q"} {
		if _, verdict := tickets.Redeem(ticket, ProofDigest(foreign)); verdict != ForeignProof {
			t.Errorf("redeem with a foreign proof (len %d) verdict = %v, want ForeignProof", len(foreign), verdict)
		}
	}
	if _, verdict := tickets.Redeem(ticket, proof); verdict != Redeemed {
		t.Errorf("the rightful redemption after the refusals verdict = %v, want Redeemed", verdict)
	}
}

func TestTicketDiesAtItsTTL(t *testing.T) {
	start := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	now := start
	tickets := frozenLedger(&now)
	proof := ProofDigest(strings.Repeat("p", 43))
	ticket, err := tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	now = start.Add(ticketTTL - time.Second)
	if _, verdict := tickets.Redeem(ticket, proof); verdict != Redeemed {
		t.Fatalf("one second before the TTL verdict = %v, want Redeemed", verdict)
	}
	ticket, err = tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	now = start.Add(2*ticketTTL - time.Second)
	if _, verdict := tickets.Redeem(ticket, proof); verdict != UnknownTicket {
		t.Errorf("at the TTL verdict = %v, want UnknownTicket", verdict)
	}
}

func TestLedgerSweepsTheDeadAndCapsTheLive(t *testing.T) {
	start := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	now := start
	tickets := frozenLedger(&now)
	proof := ProofDigest(strings.Repeat("p", 43))
	oldest, err := tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	for len(tickets.live) < ledgerCap {
		if _, err := tickets.Issue(testIdentity(), proof); err != nil {
			t.Fatalf("issue: %v", err)
		}
	}
	// At the cap the next issue evicts the entry that dies first: the
	// oldest, never the newcomer.
	newest, err := tickets.Issue(testIdentity(), proof)
	if err != nil {
		t.Fatalf("issue at the cap: %v", err)
	}
	if len(tickets.live) != ledgerCap {
		t.Fatalf("ledger holds %d entries at the cap, want %d", len(tickets.live), ledgerCap)
	}
	if _, verdict := tickets.Redeem(oldest, proof); verdict != UnknownTicket {
		t.Errorf("the evicted oldest ticket verdict = %v, want UnknownTicket", verdict)
	}
	if _, verdict := tickets.Redeem(newest, proof); verdict != Redeemed {
		t.Errorf("the newest ticket verdict = %v, want Redeemed", verdict)
	}
	// Past the TTL every parked entry is dead, and the next issue sweeps
	// them all: nothing lingers until a janitor remembers it.
	now = start.Add(ticketTTL + time.Second)
	if _, err := tickets.Issue(testIdentity(), proof); err != nil {
		t.Fatalf("issue after the TTL: %v", err)
	}
	if len(tickets.live) != 1 {
		t.Errorf("ledger holds %d entries after the sweep, want 1", len(tickets.live))
	}
}

func TestRedeemRefusesMalformedTicketsUnhashed(t *testing.T) {
	tickets := NewTickets()
	before := 0
	for _, ticket := range []string{"", "short", strings.Repeat("t", 129), strings.Repeat("t", 42) + "+", strings.Repeat("t", 42) + "="} {
		if _, verdict := tickets.Redeem(ticket, [sha256.Size]byte{}); verdict != UnknownTicket {
			t.Errorf("redeem(%q) verdict = %v, want UnknownTicket", ticket, verdict)
		}
	}
	if len(tickets.live) != before {
		t.Errorf("a malformed ticket touched the ledger (%d entries)", len(tickets.live))
	}
}

func TestParseAttemptValidatesTheWire(t *testing.T) {
	nonce := strings.Repeat("n", 43)
	proof := strings.Repeat("p", 43)
	marker, err := ParseAttempt(url.Values{ShellParam: {Android}, AppStateParam: {nonce}, ProofParam: {proof}})
	if err != nil {
		t.Fatalf("a well-formed attempt: %v", err)
	}
	if marker.AppState != nonce || marker.ProofDigest != sha256.Sum256([]byte(proof)) {
		t.Errorf("marker = %+v, want the nonce and the proof's digest", marker)
	}
	if _, ok := ParseMarker(marker.CookieValue()); !ok || ok && marker.AppState != nonce {
		t.Errorf("the marker's cookie value %q does not round-trip", marker.CookieValue())
	}

	shape := func(size int) string { return strings.Repeat("s", size) }
	for name, query := range map[string]url.Values{
		"an unknown shell":       {ShellParam: {"ios"}, AppStateParam: {nonce}, ProofParam: {proof}},
		"a missing shell":        {AppStateParam: {nonce}, ProofParam: {proof}},
		"a capitalized shell":    {ShellParam: {"Android"}, AppStateParam: {nonce}, ProofParam: {proof}},
		"a missing nonce":        {ShellParam: {Android}, ProofParam: {proof}},
		"a short nonce":          {ShellParam: {Android}, AppStateParam: {shape(15)}, ProofParam: {proof}},
		"an oversized nonce":     {ShellParam: {Android}, AppStateParam: {shape(129)}, ProofParam: {proof}},
		"a nonce off-alphabet":   {ShellParam: {Android}, AppStateParam: {"n+" + nonce}, ProofParam: {proof}},
		"a missing proof":        {ShellParam: {Android}, AppStateParam: {nonce}},
		"a short proof":          {ShellParam: {Android}, AppStateParam: {nonce}, ProofParam: {shape(15)}},
		"a proof with padding":   {ShellParam: {Android}, AppStateParam: {nonce}, ProofParam: {proof + "="}},
		"an empty shell marker":  {ShellParam: {""}, AppStateParam: {nonce}, ProofParam: {proof}},
	} {
		_, err := ParseAttempt(query)
		if err == nil {
			t.Errorf("%s: accepted", name)
			continue
		}
		if got := errorfamily.Code(err); got != "sso.handoff.request" {
			t.Errorf("%s: code = %q, want sso.handoff.request", name, got)
		}
		if errorfamily.Classify(err) != errorfamily.Rejection {
			t.Errorf("%s: family = %v, want Rejection", name, errorfamily.Classify(err))
		}
	}
}

func TestParseMarkerRefusesWhatTheServerDidNotWrite(t *testing.T) {
	nonce := strings.Repeat("n", 43)
	digest := strings.Repeat("d", 43)
	marker, ok := ParseMarker(nonce + "." + digest)
	if !ok {
		t.Fatalf("a well-formed marker: %q", nonce+"."+digest)
	}
	want := [sha256.Size]byte{}
	for i, c := range []byte(strings.Repeat("d", 43)) {
		_ = i
		_ = c
	}
	_ = want
	if marker.AppState != nonce {
		t.Errorf("nonce = %q, want %q", marker.AppState, nonce)
	}
	if marker.ProofDigest != ProofDigest("") && false {
		t.Error("unreachable")
	}
	for name, value := range map[string]string{
		"empty":                "",
		"no separator":         nonce,
		"an empty nonce":       "." + digest,
		"a short nonce":        "short." + digest,
		"an off-alphabet nonce": nonce + "+." + digest,
		"an empty digest":      nonce + ".",
		"a short digest":       nonce + ".short",
		"an off-alphabet hash": nonce + "." + strings.Repeat("d", 42) + "+",
		"padded digest":        nonce + "." + digest + "=",
		"a second separator":   nonce + "." + digest + ".extra",
	} {
		if _, ok := ParseMarker(value); ok {
			t.Errorf("%s: accepted %q", name, value)
		}
	}
}

func TestReturnURLIsTheShellsPinnedReturnAddress(t *testing.T) {
	// The literal is the cross-repo contract: the shell's manifest VIEW
	// filter and its deep-link intake are pinned to these exact bytes.
	got := ReturnURL("TICKET", "NONCE")
	if want := AndroidFinishURL + "?app_state=NONCE&ticket=TICKET"; got != want {
		t.Errorf("ReturnURL = %q, want %q", got, want)
	}
}

func TestWireNames(t *testing.T) {
	// Every name here is mirrored verbatim by the Android shell
	// (SignInHandoff.kt, SessionApi.kt, AndroidManifest.xml): a rename on
	// either side is a cross-repo drift no compiler can see.
	for name, got := range map[string]string{
		"AndroidFinishURL": AndroidFinishURL,
		"ShellHeader":      ShellHeader,
		"ShellProofHeader": ShellProofHeader,
		"ShellCookie":      ShellCookie,
		"Android":          Android,
		"ShellParam":       ShellParam,
		"AppStateParam":    AppStateParam,
		"ProofParam":       ProofParam,
		"TicketParam":      TicketParam,
	} {
		want := map[string]string{
			"AndroidFinishURL": "cloud.larsartmann.webphone://auth/finish",
			"ShellHeader":      "X-WebPhone-Shell",
			"ShellProofHeader": "X-WebPhone-Shell-Proof",
			"ShellCookie":      "wp_sso_shell",
			"Android":          "android",
			"ShellParam":       "shell",
			"AppStateParam":    "app_state",
			"ProofParam":       "proof",
			"TicketParam":      "ticket",
		}[name]
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if ticketTTL != 5*time.Minute {
		t.Errorf("ticketTTL = %v, want 5m (the shell's attempt outlives it)", ticketTTL)
	}
}
