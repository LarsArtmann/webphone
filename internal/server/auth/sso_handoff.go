package auth

// The HTTP surface of the native-shell sign-in handoff (the Android call
// shell's Pocket ID login): the capability probe and the ticket
// redemption. The contract, its wire names and the ticket ledger live in
// internal/sso/handoff, the half that needs no HTTP; the begin hop and
// the callback's bounce sit in sso_api.go beside the rest of the OIDC
// dance.

import (
	"log/slog"
	"net/http"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/sso/handoff"
)

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
	if r.Header.Get(handoff.ShellHeader) != handoff.Android {
		a.handoffRefused(w, r, http.StatusBadRequest,
			"Sign-in refused: only the app that started this sign-in can finish it. Start the sign-in again from the app.",
			errorfamily.NewRejection("sso.handoff.request", "a handoff redemption must name its shell in the "+handoff.ShellHeader+" header"),
			"redeem handoff ticket")
		return
	}
	identity, verdict := a.handoffs.Redeem(r.URL.Query().Get(handoff.TicketParam),
		handoff.ProofDigest(r.Header.Get(handoff.ShellProofHeader)))
	switch verdict {
	case handoff.UnknownTicket:
		a.ssoFailed(w, r, errorfamily.NewRejection("sso.handoff.ticket",
			"the handoff ticket is unknown, expired or already redeemed"), "redeem handoff ticket")
		return
	case handoff.ForeignProof:
		// Someone holds the ticket without the proof that never rode the
		// return link: that is an interception attempt, worth a line.
		a.handoffRefused(w, r, http.StatusForbidden,
			"Sign-in refused: this ticket belongs to another sign-in attempt. Start the sign-in again from the app.",
			errorfamily.NewRejection("sso.handoff.proof", "the handoff ticket was presented without the proof of the attempt that started it"),
			"redeem handoff ticket")
		return
	case handoff.Redeemed:
	}
	if _, err := a.mintSession(w, r, identity.Extension, identity.Password, identity.Email); err != nil {
		slog.Error("webphone: oidc handoff session mint failed", "error", err, "extension", identity.Extension.String())
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	AuthAudit("login", "email", identity.Email, "oidc verified; session minted through the android shell handoff")
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
