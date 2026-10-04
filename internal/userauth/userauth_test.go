// Package userauth tests: the embedded identity layer's webphone-side
// semantics — mapping resolution, fail-closed password sourcing, the
// one-time enrollment token lifecycle — driven through the deterministic
// WebAuthn stub (no browser signatures needed; usermgmt's own suite
// pioneered the shape).
package userauth_test

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // registers the "sqlite" driver for the test's own connection

	errorfamily "github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/userauth"
)

// stubWebAuthn is the deterministic provider: every ceremony succeeds,
// every response is canned (mirrors usermgmt's webauthn_stub_test.go).
type stubWebAuthn struct{}

func (stubWebAuthn) BeginRegistration(_ context.Context, _ []byte) ([]byte, []byte, error) {
	return []byte(`{"publicKey":{"challenge":"dGVzdA=="}}`), []byte(`{"sid":"reg"}`), nil
}

func (stubWebAuthn) FinishRegistration(_ context.Context, _, _, _ []byte) ([]byte, error) {
	return []byte(`{"id":"dGVzdA==","public_key":"dGVzdA==","attestation_type":"none","sign_count":0}`), nil
}

func (stubWebAuthn) BeginLogin(_ context.Context, _ []byte) ([]byte, []byte, error) {
	return []byte(`{"publicKey":{"challenge":"dGVzdA=="}}`), []byte(`{"sid":"login"}`), nil
}

func (stubWebAuthn) FinishLogin(_ context.Context, _, _, _ []byte) error {
	return nil
}

const (
	mappedEmail  = "lars@example.com"
	otherEmail   = "alice@example.com"
	ext1000      = "1000"
	passwordFile = "dir-password"
)

// newService builds the service with the stub provider, one mapped user
// (lars/1000) and a password file holding a fixed secret; returns the
// service, its data dir and the password file path.
func newService(t *testing.T) (*userauth.Service, string, string) {
	t.Helper()
	dir := t.TempDir()
	passFile := filepath.Join(dir, "ext1000")
	if err := os.WriteFile(passFile, []byte(passwordFile+"\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	ext, err := domain.ParseExtension(ext1000)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := userauth.New(context.Background(), userauth.PasskeyRuntime{
		RPID:          "pbx.example.org",
		RPDisplayName: "WebPhone",
		RPOrigins:     []string{"https://pbx.example.org"},
		Users: map[string]userauth.MappedUser{
			mappedEmail: {Email: mappedEmail, DisplayName: "Lars", Extensions: []domain.Extension{ext}},
		},
		ExtensionPasswordFiles: map[string]string{ext1000: passFile},
		WebAuthn:               stubWebAuthn{},
	}, dir, slog.Default())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })
	return svc, dir, passFile
}

func TestRegisterIsIdempotent(t *testing.T) {
	svc, _, _ := newService(t)
	ctx := context.Background()
	first, err := svc.Register(ctx, mappedEmail)
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	second, err := svc.Register(ctx, mappedEmail)
	if err != nil {
		t.Fatalf("second register (add-a-device flow): %v", err)
	}
	if first != second {
		t.Errorf("re-register returned a different userID: %q vs %q", first, second)
	}
}

// TestShutdownClosesTheIdentityDatabase pins the lifecycle contract the
// composition root rides (do.ShutdownerWithError): Shutdown answers nil
// on the healthy path and actually closes usermgmt.db — the health
// check must fail afterwards, never silently keep passing.
func TestShutdownClosesTheIdentityDatabase(t *testing.T) {
	svc, _, _ := newService(t)
	if err := svc.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck before Shutdown: %v", err)
	}
	if err := svc.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if err := svc.HealthCheck(context.Background()); err == nil {
		t.Error("HealthCheck after Shutdown: want the closed-database error")
	}
}

func TestEnrollTokenLifecycle(t *testing.T) {
	svc, _, _ := newService(t)
	ctx := context.Background()
	userID, err := svc.Register(ctx, mappedEmail)
	if err != nil {
		t.Fatal(err)
	}
	token, err := svc.MintEnrollToken(ctx, mappedEmail, userID)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if len(token) < 32 || strings.ContainsAny(token, "+/=") {
		t.Errorf("token must be long, URL-safe base64: %q", token)
	}

	// First use resolves AND burns.
	resolved, err := svc.VerifyEnrollToken(ctx, token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if resolved.Email != mappedEmail || resolved.UserID != userID {
		t.Errorf("resolved %+v, want email=%s userID=%s", resolved, mappedEmail, userID)
	}

	// Second use: the SAME rejection as an unknown token (no used-token
	// oracle), same family, same code.
	_, usedErr := svc.VerifyEnrollToken(ctx, token)
	_, unknownErr := svc.VerifyEnrollToken(ctx, "definitely-not-a-known-token-xyz")
	if usedErr == nil || unknownErr == nil {
		t.Fatal("used and unknown tokens must both reject")
	}
	errorfamilytest.AssertFamily(t, usedErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, usedErr, "userauth.enroll.invalid")
	errorfamilytest.AssertFamily(t, unknownErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, unknownErr, "userauth.enroll.invalid")
}

func TestEnrollTokenExpiryIsOperatorVisible(t *testing.T) {
	svc, dir, _ := newService(t)
	ctx := context.Background()
	userID, err := svc.Register(ctx, mappedEmail)
	if err != nil {
		t.Fatal(err)
	}
	token, err := svc.MintEnrollToken(ctx, mappedEmail, userID)
	if err != nil {
		t.Fatal(err)
	}
	// Force the row into the past through the store's own file (WAL
	// allows the test connection alongside the service's).
	testDB, err := sql.Open("sqlite", filepath.Join(dir, "usermgmt.db")+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	if _, err := testDB.ExecContext(ctx, `UPDATE wp_enroll_tokens SET expires_at = 1`); err != nil {
		t.Fatalf("age the token row: %v", err)
	}
	_, expiredErr := svc.VerifyEnrollToken(ctx, token)
	if expiredErr == nil {
		t.Fatal("expired token must reject")
	}
	errorfamilytest.AssertFamily(t, expiredErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, expiredErr, "userauth.enroll.expired")
}

// finishLoginFor drives begin+finish for an email with the stub provider
// (every ceremony succeeds) and returns the mapped user.
func finishLoginFor(t *testing.T, svc *userauth.Service, email string) (userauth.MappedUser, error) {
	t.Helper()
	ctx := context.Background()
	begun, err := svc.BeginLogin(ctx, email)
	if err != nil {
		return userauth.MappedUser{}, err
	}
	r := httptest.NewRequest(http.MethodPost, "/finish?user_id="+begun.SessionKey, strings.NewReader(`{}`))
	return svc.FinishLogin(ctx, begun.SessionKey, r)
}

func TestRegistrationAndLoginCeremoniesResolveMapping(t *testing.T) {
	svc, _, _ := newService(t)
	ctx := context.Background()
	userID, err := svc.Register(ctx, mappedEmail)
	if err != nil {
		t.Fatal(err)
	}
	// Before any credential exists, login begin fails closed.
	if _, err := svc.BeginLogin(ctx, mappedEmail); err == nil {
		t.Fatal("login before enrollment must fail (no credentials)")
	}
	// Enrollment ceremony: begin (session key == the userID wire
	// contract) then finish with the stub attestation.
	begun, err := svc.BeginRegistration(ctx, userID)
	if err != nil {
		t.Fatalf("begin registration: %v", err)
	}
	if begun.SessionKey == "" {
		t.Fatal("begin registration lost the session key")
	}
	r := httptest.NewRequest(http.MethodPost, "/enroll/finish?user_id="+begun.SessionKey, strings.NewReader(`{}`))
	if err := svc.FinishRegistration(ctx, begun.SessionKey, r, "laptop"); err != nil {
		t.Fatalf("finish registration: %v", err)
	}

	mapped, err := finishLoginFor(t, svc, mappedEmail)
	if err != nil {
		t.Fatalf("finish login: %v", err)
	}
	if mapped.Email != mappedEmail || mapped.DisplayName != "Lars" {
		t.Errorf("mapped = %+v", mapped)
	}
	if got := mapped.SessionExtension().String(); got != ext1000 {
		t.Errorf("session extension = %q, want %q", got, ext1000)
	}
}

func TestFinishLoginUnmappedAccountFailsClosed(t *testing.T) {
	svc, _, _ := newService(t)
	ctx := context.Background()
	// Register an account that is deliberately NOT in the users mapping.
	userID, err := svc.Register(ctx, otherEmail)
	if err != nil {
		t.Fatal(err)
	}
	begun, err := svc.BeginRegistration(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/finish?user_id="+begun.SessionKey, strings.NewReader(`{}`))
	if err := svc.FinishRegistration(ctx, begun.SessionKey, r, "device"); err != nil {
		t.Fatal(err)
	}
	// The stub provider approves the ceremony; the MAPPING is the gate.
	_, err = finishLoginFor(t, svc, otherEmail)
	if err == nil {
		t.Fatal("unmapped account must never mint a session")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "userauth.unmapped_email")
}

func TestSIPPasswordFailsClosedOnEveryBadState(t *testing.T) {
	svc, _, passFile := newService(t)
	ext := domain.MustParseExtension(ext1000)

	got, err := svc.SIPPassword(ext)
	if err != nil {
		t.Fatalf("happy path: %v", err)
	}
	if got != passwordFile {
		t.Errorf("password = %q, want trimmed %q", got, passwordFile)
	}

	// Empty file: the 2026-10-01 outage class.
	if err := os.WriteFile(passFile, []byte("   \n"), 0o640); err != nil {
		t.Fatal(err)
	}
	_, err = svc.SIPPassword(ext)
	if err == nil {
		t.Fatal("empty password file must reject")
	}
	errorfamilytest.AssertCode(t, err, "userauth.password_file.empty")

	// Unreadable file: perms reject with the read code. (WriteFile does
	// not change an existing file's mode — Chmod is the honest lever.)
	if err := os.WriteFile(passFile, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(passFile, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(passFile, 0o640) })
	_, err = svc.SIPPassword(ext)
	if err == nil {
		t.Fatal("unreadable password file must reject")
	}
	errorfamilytest.AssertCode(t, err, "userauth.password_file.read")

	// Unmapped extension: no file configured at all.
	otherExt := domain.MustParseExtension("1009")
	_, err = svc.SIPPassword(otherExt)
	if err == nil {
		t.Fatal("extension without a configured file must reject")
	}
	errorfamilytest.AssertCode(t, err, "userauth.password_file.missing")
}

func TestResolveIsCaseAndWhitespaceInsensitive(t *testing.T) {
	svc, _, _ := newService(t)
	mapped, ok := svc.Resolve("  LARS@Example.COM ")
	if !ok {
		t.Fatal("mapping lookup must normalize case and whitespace")
	}
	if mapped.Email != mappedEmail {
		t.Errorf("resolved email = %q", mapped.Email)
	}
	if _, ok := svc.Resolve(otherEmail); ok {
		t.Error("unmapped email must not resolve")
	}
}

func TestMappedByExtensionCoversEveryMappedExtension(t *testing.T) {
	svc, _, _ := newService(t)
	mapped, ok := svc.MappedByExtension(domain.MustParseExtension(ext1000))
	if !ok {
		t.Fatal("1000 must reverse-resolve")
	}
	if mapped.DisplayName != "Lars" {
		t.Errorf("reverse-resolved %+v", mapped)
	}
	if _, ok := svc.MappedByExtension(domain.MustParseExtension("1009")); ok {
		t.Error("unmapped extension must not reverse-resolve")
	}
}
