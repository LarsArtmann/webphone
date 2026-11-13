package auth

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	identitymodel "github.com/larsartmann/cqrs-htmx/identity-model/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/web/views"
)

// The passkey (WebAuthn) login + enrollment surface. Routes are
// registered ONLY when Deps.UserAuth is non-nil (the config-gated
// embedded usermgmt identity layer); disabled deployments keep answering
// the styled 404, exactly as before the mode existed.
//
// Wire contract mirrors usermgmt's own AuthHandler (its shape is the
// library's pinned ceremony contract): begin answers
// {options, session_key}; finish takes ?user_id=<session_key> plus the
// RAW ceremony body (attestation/assertion JSON). The enroll page
// coerces base64url fields both ways (see assets/enroll/webauthn.js);
// the /login library page owns its own coercion.

// parsePasskeyUserID is the strict door for browser-supplied ceremony
// keys: usermgmt's SessionKey is always a minted ULID, so garbage is
// malformed input (400) — never a silently hashed pseudo-user. Zero
// UserID with ok=false means the 400 is already written.
func parsePasskeyUserID(w http.ResponseWriter, raw string) (identitymodel.UserID, bool) {
	userID, err := identitymodel.ParseUserID(raw)
	if err != nil {
		http.Error(w, "user_id must be the session_key the begin step returned", http.StatusBadRequest)
		return identitymodel.UserID{}, false
	}
	return userID, true
}

// passkeyUserIDQuery is the ONE home for the ceremony callbacks'
// user_id prologue (both finish handlers share it): read the query
// parameter, refuse the empty form, parse strictly at the door.
func passkeyUserIDQuery(w http.ResponseWriter, r *http.Request) (identitymodel.UserID, bool) {
	raw := r.URL.Query().Get("user_id")
	if raw == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return identitymodel.UserID{}, false
	}
	return parsePasskeyUserID(w, raw)
}

// PasskeyBeginLogin starts the passkey ceremony for an email. Unknown
// emails and credential-less accounts answer the SAME 401 — the endpoint
// must not double as an email-enumeration oracle (usermgmt's own handler
// has the same posture; its log keeps the honest detail).
func (a *Handlers) PasskeyBeginLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.UnmarshalRead(r.Body, &body); err != nil || body.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	resp, err := a.deps.UserAuth.BeginLogin(r.Context(), body.Email)
	switch {
	case err == nil:
	case errors.Is(err, identitymodel.ErrUserNotFound), errors.Is(err, identitymodel.ErrNoCredentials):
		http.Error(w, "unknown email or no passkey enrolled", http.StatusUnauthorized)
		return
	case errors.Is(err, identitymodel.ErrAccountLocked):
		http.Error(w, "account temporarily locked", http.StatusTooManyRequests)
		return
	default:
		a.passkeyServerError(w, err, "begin passkey login")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// PasskeyFinishLogin verifies the assertion (the raw body, per the
// usermgmt wire contract), resolves the account's extension mapping,
// sources the SIP directory password from its operator-managed file and
// verifies it against the PBX directory (fail-closed on a stale file,
// exactly like the extension login), then mints the webphone session.
// The response is the CreateSession shape plus display_name and numbers
// so the island can show the user's real identity and numbers.
func (a *Handlers) PasskeyFinishLogin(w http.ResponseWriter, r *http.Request) {
	userID, ok := passkeyUserIDQuery(w, r)
	if !ok {
		return
	}
	mapped, err := a.deps.UserAuth.FinishLogin(r.Context(), userID, r)
	switch {
	case err == nil:
	case errors.Is(err, identitymodel.ErrUserNotFound):
		http.Error(w, "unknown email or no passkey enrolled", http.StatusUnauthorized)
		return
	default:
		a.passkeyServerError(w, err, "finish passkey login")
		return
	}
	extension := mapped.SessionExtension()
	password, err := a.deps.UserAuth.SIPPassword(extension)
	if err != nil {
		a.passkeyServerError(w, err, "source extension credentials")
		return
	}
	if !a.verifyDirectoryCredentials(w, r, extension, password) {
		return
	}
	// The session remembers WHO minted it (the multi-user substrate):
	// the audit lanes read this back on every later action — the
	// extension alone cannot distinguish two humans on a shared line.
	token, err := a.mintSession(w, r, extension, password, mapped.Email)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	// The finish response is the OTHER legacy-carrier surface (the
	// browser REGISTER needs the file-sourced password when brokering
	// is off): its credential plus the unconditional no-store ride this
	// site; the derived-credential policy itself lives in sessionAnswer.
	w.Header().Set("Cache-Control", "no-store")
	response := a.sessionAnswer(w, r.Context(), token, extension, mapped.Email)
	if !a.deps.Config.SIP.DerivedCredentials() {
		response.Password = password
	}
	AuthAudit("login", "email", mapped.Email, "passkey verified; session minted")
	writeJSON(w, http.StatusCreated, response)
}

// PasskeyEnrollVerify resolves AND burns a one-time enrollment token
// (minted by the CLI). Unknown, expired and used tokens are
// indistinguishable — the endpoint must not leak token state.
func (a *Handlers) PasskeyEnrollVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.UnmarshalRead(r.Body, &body); err != nil || body.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}
	enrolled, err := a.deps.UserAuth.VerifyEnrollToken(r.Context(), body.Token)
	if err != nil {
		a.passkeyServerError(w, err, "verify enrollment token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"email":   enrolled.Email,
		"user_id": enrolled.UserID.Get().String(),
	})
}

// PasskeyEnrollBegin starts the credential registration ceremony for the
// user the (already verified) enrollment token named.
func (a *Handlers) PasskeyEnrollBegin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.UnmarshalRead(r.Body, &body); err != nil || body.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	userID, ok := parsePasskeyUserID(w, body.UserID)
	if !ok {
		return
	}
	resp, err := a.deps.UserAuth.BeginRegistration(r.Context(), userID)
	if err != nil {
		a.passkeyServerError(w, err, "begin passkey enrollment")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// PasskeyEnrollFinish persists the attested credential. After this the
// account can passkey-log in; the page sends the user to the login form.
func (a *Handlers) PasskeyEnrollFinish(w http.ResponseWriter, r *http.Request) {
	userID, ok := passkeyUserIDQuery(w, r)
	if !ok {
		return
	}
	credentialName := r.URL.Query().Get("credential_name")
	if credentialName == "" {
		credentialName = "passkey"
	}
	if err := a.deps.UserAuth.FinishRegistration(r.Context(), userID, r, credentialName); err != nil {
		a.passkeyServerError(w, err, "finish passkey enrollment")
		return
	}
	AuthAudit("passkey_enroll", "credential", credentialName, "ceremony accepted")
	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

// EnrollPage renders the standalone passkey-enrollment page: the
// landing surface for a CLI-minted enrollment link. No session, no
// island runtime — the one-time token is the gate, and the page's
// module (assets/enroll/enroll.js) drives verify → begin → finish
// against the three POST endpoints above.
func (a *Handlers) EnrollPage(w http.ResponseWriter, r *http.Request) {
	props := views.EnrollProps{
		Lang:      a.shell.Lang(r),
		CSRFToken: csrfToken(r),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.EnrollPage(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// csrfToken reads the per-request CSRF token (the same source the
// server's shell renders: httputil carries it on the request context).
func csrfToken(r *http.Request) string {
	return httputil.CSRFTokenFromRequest(r)
}

// passkeyServerError maps a userauth error to its HTTP shape. Rejections
// are operator-fixable states (missing/unreadable password file, burned
// token, unmapped account); transient is infrastructure — both answer
// 503, the honest "try again later, the operator must fix something"
// verdict, while the family log line carries the code and context.
func (a *Handlers) passkeyServerError(w http.ResponseWriter, err error, op string) {
	status := http.StatusServiceUnavailable
	if errors.Is(err, identitymodel.ErrAccountLocked) {
		status = http.StatusTooManyRequests
	}
	slog.Warn("webphone: passkey ceremony failed", "op", op, "error", err,
		"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err))
	http.Error(w, "passkey authentication temporarily unavailable", status)
}
