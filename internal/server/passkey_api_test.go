// Passkey handler tests: the HTTP surface's gates — mode-off 404s, the
// uniform anti-enumeration 401, the finish→session mint (the one place
// a passkey becomes a webphone session), and the one-time enroll token
// burn. The identity layer runs with the deterministic stub provider
// (see internal/userauth tests), so every ceremony succeeds and the
// HANDLER's mapping is what's under test.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/userauth"
)

const (
	pkEmail      = "lars@example.com"
	pkUnknown    = "nobody@example.com"
	pkPassword   = "dir-password"
	uniformBody  = "unknown email or no passkey enrolled\n"
	passkeyBegin = `/api/auth/passkey/begin`
)

// pkStub mirrors userauth's test stub: every ceremony succeeds.
type pkStub struct{}

func (pkStub) BeginRegistration(_ context.Context, _ []byte) ([]byte, []byte, error) {
	return []byte(`{"publicKey":{}}`), []byte(`{}`), nil
}

func (pkStub) FinishRegistration(_ context.Context, _, _, _ []byte) ([]byte, error) {
	return []byte(`{"id":"dGVzdA==","public_key":"dGVzdA==","attestation_type":"none","sign_count":0}`), nil
}

func (pkStub) BeginLogin(_ context.Context, _ []byte) ([]byte, []byte, error) {
	return []byte(`{"publicKey":{}}`), []byte(`{}`), nil
}

func (pkStub) FinishLogin(_ context.Context, _, _, _ []byte) error {
	return nil
}

// newPasskeyServer wires the identity layer (stub provider, mapped
// lars→1000, password file) into the standard fixture and returns the
// service alongside, so tests can register accounts and mint tokens
// the way the CLI does.
func newPasskeyServer(t *testing.T) (*testServer, *userauth.Service) {
	t.Helper()
	dir := t.TempDir()
	passFile := filepath.Join(dir, "ext1000")
	if err := os.WriteFile(passFile, []byte(pkPassword+"\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	ext := domain.MustParseExtension("1000")
	svc, err := userauth.New(context.Background(), userauth.PasskeyRuntime{
		RPID:          "localhost",
		RPDisplayName: "WebPhone",
		RPOrigins:     []string{"http://" + "localhost"},
		Users: map[string]userauth.MappedUser{
			pkEmail: {Email: pkEmail, DisplayName: "Lars", Extensions: []domain.Extension{ext}},
		},
		ExtensionPasswordFiles: map[string]string{"1000": passFile},
		WebAuthn:               pkStub{},
	}, dir, slog.Default())
	if err != nil {
		t.Fatalf("userauth.New: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })

	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Identities = map[string]string{"1000": "+17287289311"}
	}, func(deps *Deps) {
		deps.UserAuth = svc
	})
	return server, svc
}

// enrollCredential registers the account and walks the enrollment
// ceremony so BeginLogin stops answering ErrNoCredentials.
func enrollCredential(t *testing.T, svc *userauth.Service, email string) string {
	t.Helper()
	ctx := context.Background()
	userID, err := svc.Register(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	begun, err := svc.BeginRegistration(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/finish?user_id="+begun.SessionKey, strings.NewReader(`{}`))
	if err := svc.FinishRegistration(ctx, begun.SessionKey, r, "test-device"); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestPasskeyOffAnswersTheStyled404(t *testing.T) {
	c := newClient(t) // standard fixture: no UserAuth
	for _, target := range []struct{ method, path string }{
		{http.MethodPost, passkeyBegin},
		{http.MethodPost, "/api/auth/passkey/finish?user_id=x"},
		{http.MethodPost, "/api/auth/passkey/enroll/verify"},
		{http.MethodGet, "/enroll"},
	} {
		resp, body := c.do(target.method, target.path, []byte(`{}`), "application/json")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s %s: status %d, want 404 (mode off)", target.method, target.path, resp.StatusCode)
		}
		if !strings.Contains(string(body), "WebPhone") {
			t.Errorf("%s %s: the 404 must stay the styled shell, got %q", target.method, target.path, body)
		}
	}
}

func TestPasskeyBeginIsUniformAgainstEnumeration(t *testing.T) {
	server, svc := newPasskeyServer(t)
	_ = svc
	c := clientFor(t, server)

	// Unknown email and enrolled-account-less email must answer the
	// SAME 401 body — begin must not double as an account oracle.
	payload, _ := json.Marshal(map[string]string{"email": pkUnknown})
	resp, body := c.do(http.MethodPost, passkeyBegin, payload, "application/json")
	if resp.StatusCode != http.StatusUnauthorized || string(body) != uniformBody {
		t.Fatalf("unknown email: %d %q", resp.StatusCode, body)
	}

	payload, _ = json.Marshal(map[string]string{"email": pkEmail}) // registered later? not yet: same class
	resp, body = c.do(http.MethodPost, passkeyBegin, payload, "application/json")
	if resp.StatusCode != http.StatusUnauthorized || string(body) != uniformBody {
		t.Fatalf("credential-less email: %d %q — must match the unknown-email verdict", resp.StatusCode, body)
	}

	// Malformed bodies are honest 400s (the class is different: the
	// caller, not the account, is wrong).
	resp, _ = c.do(http.MethodPost, passkeyBegin, []byte(`not-json`), "application/json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("garbage body: %d, want 400", resp.StatusCode)
	}
}

func TestPasskeyFinishMintsTheSessionWithIdentity(t *testing.T) {
	server, svc := newPasskeyServer(t)
	enrollCredential(t, svc, pkEmail)
	c := clientFor(t, server)

	begun, err := svc.BeginLogin(context.Background(), pkEmail)
	if err != nil {
		t.Fatalf("begin login: %v", err)
	}
	resp, body := c.do(http.MethodPost,
		"/api/auth/passkey/finish?user_id="+begun.SessionKey,
		[]byte(`{"id":"x","rawId":"eA","type":"public-key","response":{"clientDataJSON":"eA"}}`),
		"application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("finish: %d %s", resp.StatusCode, body)
	}
	var identity struct {
		Extension   string   `json:"extension"`
		Password    string   `json:"password"`
		DisplayName string   `json:"display_name"`
		Numbers     []string `json:"numbers"`
	}
	if err := json.Unmarshal(body, &identity); err != nil {
		t.Fatalf("finish body %q: %v", body, err)
	}
	if identity.Extension != "1000" || identity.Password != pkPassword {
		t.Errorf("session credentials = %+v", identity)
	}
	if identity.DisplayName != "Lars" {
		t.Errorf("display_name = %q, want Lars", identity.DisplayName)
	}
	if len(identity.Numbers) != 1 || identity.Numbers[0] != "+17287289311" {
		t.Errorf("numbers = %v, want the config identity", identity.Numbers)
	}

	// The minted cookie is a real webphone session: the resume endpoint
	// hands the same credentials back.
	resp, body = c.do(http.MethodGet, "/api/session", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resume with the minted cookie: %d %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"extension":"1000"`) || !strings.Contains(string(body), pkPassword) {
		t.Errorf("resume body %q lost the session credentials", body)
	}
}

func TestPasskeyFinishWithoutUserIDIsA400(t *testing.T) {
	server, _ := newPasskeyServer(t)
	c := clientFor(t, server)
	resp, _ := c.do(http.MethodPost, "/api/auth/passkey/finish", []byte(`{}`), "application/json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing user_id: %d, want 400", resp.StatusCode)
	}
}

func TestPasskeyEnrollVerifyBurnsTheToken(t *testing.T) {
	server, svc := newPasskeyServer(t)
	userID := enrollCredential(t, svc, pkEmail)
	c := clientFor(t, server)

	token, err := svc.MintEnrollToken(context.Background(), pkEmail, userID)
	if err != nil {
		t.Fatal(err)
	}

	verify := func(tok string) (*http.Response, []byte) {
		payload, _ := json.Marshal(map[string]string{"token": tok})
		return c.do(http.MethodPost, "/api/auth/passkey/enroll/verify", payload, "application/json")
	}

	resp, body := verify(token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first verify: %d %s", resp.StatusCode, body)
	}
	var resolved struct {
		Email  string `json:"email"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(body, &resolved); err != nil || resolved.Email != pkEmail || resolved.UserID != userID {
		t.Fatalf("verify body %q: %v", body, err)
	}

	// The burn: the same token answers 503 like every other rejection.
	resp, _ = verify(token)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("replayed token: %d, want 503", resp.StatusCode)
	}
	resp, _ = verify("not-a-token-at-all-1234")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("garbage token: %d, want 503 (same shape)", resp.StatusCode)
	}
}

func TestPasskeySurfacesRequireCSRF(t *testing.T) {
	server, svc := newPasskeyServer(t)
	enrollCredential(t, svc, pkEmail)

	// A POST without the CSRF header must never reach the ceremony.
	req, err := http.NewRequest(http.MethodPost, server.URL+passkeyBegin,
		strings.NewReader(`{"email":"`+pkEmail+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("headerless POST: %d, want 403", resp.StatusCode)
	}
}
