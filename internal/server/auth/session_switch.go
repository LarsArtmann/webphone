package auth

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"slices"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// SwitchRequest is the POST /api/session/switch wire shape: another
// extension mapped to the signed-in user (D1, the line switcher). The
// server resolves the mapping itself — the client's suggestion is
// never authority for what a session may become.
type SwitchRequest struct {
	Extension string `description:"A mapped extension of the signed-in user to re-mint the session for." json:"extension"`
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
