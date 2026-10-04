package server

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	"github.com/larsartmann/go-error-family"

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
// RAW ceremony body (attestation/assertion JSON). The island coerces
// base64url fields both ways (see assets/island/app/webauthn.js).

// maxPasskeyBodySize bounds the ceremony bodies (attestation/assertion
// JSON): a few KB of base64; 64 KiB leaves generous headroom while
// keeping a hostile POST out of the parser.
const maxPasskeyBodySize = 64 << 10

// passkeyBeginLogin starts the passkey ceremony for an email. Unknown
// emails and credential-less accounts answer the SAME 401 — the endpoint
// must not double as an email-enumeration oracle (usermgmt's own handler
// has the same posture; its log keeps the honest detail).
func (h *handlers) passkeyBeginLogin(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil || body.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	resp, err := h.deps.UserAuth.BeginLogin(r.Context(), body.Email)
	switch {
	case err == nil:
	case errors.Is(err, usermgmt.ErrUserNotFound), errors.Is(err, usermgmt.ErrNoCredentials):
		http.Error(w, "unknown email or no passkey enrolled", http.StatusUnauthorized)
		return
	case errors.Is(err, usermgmt.ErrAccountLocked):
		http.Error(w, "account temporarily locked", http.StatusTooManyRequests)
		return
	default:
		h.passkeyServerError(w, r, err, "begin passkey login")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// passkeyFinishLogin verifies the assertion (the raw body, per the
// usermgmt wire contract), resolves the account's extension mapping,
// sources the SIP directory password from its operator-managed file and
// verifies it against the PBX directory (fail-closed on a stale file,
// exactly like the extension login), then mints the webphone session.
// The response is the createSession shape plus display_name and numbers
// so the island can show the user's real identity and numbers.
func (h *handlers) passkeyFinishLogin(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPasskeyBodySize)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}
	mapped, err := h.deps.UserAuth.FinishLogin(r.Context(), userID, r)
	switch {
	case err == nil:
	case errors.Is(err, usermgmt.ErrUserNotFound):
		http.Error(w, "unknown email or no passkey enrolled", http.StatusUnauthorized)
		return
	default:
		h.passkeyServerError(w, r, err, "finish passkey login")
		return
	}
	extension := mapped.SessionExtension()
	password, err := h.deps.UserAuth.SIPPassword(extension)
	if err != nil {
		h.passkeyServerError(w, r, err, "source extension credentials")
		return
	}
	if !h.verifyDirectoryCredentials(w, r, extension, password) {
		return
	}
	if err := h.mintSession(w, r, extension, password); err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	// The finish response carries the file-sourced SIP password back to
	// the island (its REGISTER needs it; the extension login never sends
	// one because the user just typed it). Same no-store posture as the
	// resume endpoint — the credential stays out of every cache.
	w.Header().Set("Cache-Control", "no-store")
	response := h.sessionIdentityResponse(extension)
	response["password"] = password
	writeJSON(w, http.StatusCreated, response)
}

// passkeyEnrollVerify resolves AND burns a one-time enrollment token
// (minted by the CLI). Unknown, expired and used tokens are
// indistinguishable — the endpoint must not leak token state.
func (h *handlers) passkeyEnrollVerify(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil || body.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}
	enrolled, err := h.deps.UserAuth.VerifyEnrollToken(r.Context(), body.Token)
	if err != nil {
		h.passkeyServerError(w, r, err, "verify enrollment token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"email":   enrolled.Email,
		"user_id": enrolled.UserID,
	})
}

// passkeyEnrollBegin starts the credential registration ceremony for the
// user the (already verified) enrollment token named.
func (h *handlers) passkeyEnrollBegin(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil || body.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	resp, err := h.deps.UserAuth.BeginRegistration(r.Context(), body.UserID)
	if err != nil {
		h.passkeyServerError(w, r, err, "begin passkey enrollment")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// passkeyEnrollFinish persists the attested credential. After this the
// account can passkey-log in; the page sends the user to the login form.
func (h *handlers) passkeyEnrollFinish(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPasskeyBodySize)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}
	credentialName := r.URL.Query().Get("credential_name")
	if credentialName == "" {
		credentialName = "passkey"
	}
	if err := h.deps.UserAuth.FinishRegistration(r.Context(), userID, r, credentialName); err != nil {
		h.passkeyServerError(w, r, err, "finish passkey enrollment")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

// enrollPage renders the standalone passkey-enrollment page: the
// landing surface for a CLI-minted enrollment link. No session, no
// island runtime — the one-time token is the gate, and the page's
// module (assets/enroll/enroll.js) drives verify → begin → finish
// against the three POST endpoints above.
func (h *handlers) enrollPage(w http.ResponseWriter, r *http.Request) {
	if h.deps.UserAuth == nil {
		http.NotFound(w, r)
		return
	}
	props := views.EnrollProps{
		Lang:      h.lang(r),
		CSRFToken: csrfToken(r),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.EnrollPage(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// passkeyServerError maps a userauth error to its HTTP shape. Rejections
// are operator-fixable states (missing/unreadable password file, burned
// token, unmapped account); transient is infrastructure — both answer
// 503, the honest "try again later, the operator must fix something"
// verdict, while the family log line carries the code and context.
func (h *handlers) passkeyServerError(w http.ResponseWriter, r *http.Request, err error, op string) {
	status := http.StatusServiceUnavailable
	if errors.Is(err, usermgmt.ErrAccountLocked) {
		status = http.StatusTooManyRequests
	}
	slog.Warn("webphone: passkey ceremony failed", "op", op, "error", err,
		"family", errorfamily.Classify(err).String(), "code", errorfamily.Code(err))
	http.Error(w, "passkey authentication temporarily unavailable", status)
}

// writeJSON is the one home for the passkey surface's JSON responses.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.MarshalWrite(w, payload) //nolint:erraudit // best-effort write; the response is already committed
}
