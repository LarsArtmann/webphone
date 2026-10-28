// Line-switcher handler tests (D1): the switch surface's gates — the
// mapped-membership authorization, the no-op switch, the CSRF + cookie
// rotation, the audit line naming the human move, and the mode-off 404.
// The identity layer runs with the deterministic stub provider (see
// internal/userauth tests), mapped to TWO extensions so a switch has a
// legal target.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/userauth"
)

const (
	switchTargetPassword = "dir-password-1001"
	switchPath           = "/api/session/switch"
)

// newSwitchIdentity builds the identity layer mapped to TWO extensions
// (1000 + 1001, one password file each): the switch's legal caller and
// its legal target.
func newSwitchIdentity(t *testing.T) *userauth.Service {
	t.Helper()
	dir := t.TempDir()
	fileFor := func(ext, password string) string {
		file := filepath.Join(dir, "ext"+ext)
		if err := os.WriteFile(file, []byte(password+"\n"), 0o640); err != nil {
			t.Fatal(err)
		}
		return file
	}
	engine, err := readmodels.Open(dir)
	if err != nil {
		t.Fatalf("readmodels.Open: %v", err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	svc, err := userauth.New(context.Background(), userauth.PasskeyRuntime{
		RPID:          "localhost",
		RPDisplayName: "WebPhone",
		RPOrigins:     []string{"http://" + "localhost"},
		Users: map[string]userauth.MappedUser{
			pkEmail: {Email: pkEmail, DisplayName: "Lars", Extensions: []domain.Extension{
				domain.MustParseExtension("1000"),
				domain.MustParseExtension("1001"),
			}},
		},
		ExtensionPasswordFiles: map[string]string{
			"1000": fileFor("1000", pkPassword),
			"1001": fileFor("1001", switchTargetPassword),
		},
		WebAuthn: pkStub{},
	}, dir, engine.EnrollTokens(), slog.Default())
	if err != nil {
		t.Fatalf("userauth.New: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })
	return svc
}

// passkeyLoginOn walks the stubbed ceremony on the given server and
// adopts the rotated CSRF token, leaving the client signed in as the
// mapped user.
func passkeyLoginOn(t *testing.T, c *client, svc *userauth.Service) {
	t.Helper()
	enrollCredential(t, svc, pkEmail)
	begun, err := svc.BeginLogin(context.Background(), pkEmail)
	if err != nil {
		t.Fatalf("begin login: %v", err)
	}
	resp, body := c.do(http.MethodPost,
		passkeyFinish+"?user_id="+begun.SessionKey,
		[]byte(`{"id":"x","rawId":"eA","type":"public-key","response":{"clientDataJSON":"eA"}}`),
		"application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("passkey finish: %d %s", resp.StatusCode, body)
	}
	c.adoptCsrfToken()
}

// newSwitchServer wires the switch identity into the loopback fixture
// (directory verify skipped, like every dev boot) and signs in via the
// stubbed passkey ceremony.
func newSwitchServer(t *testing.T) (*testServer, *client) {
	t.Helper()
	svc := newSwitchIdentity(t)
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Identities = map[string]string{"1000": "+17287289311", "1001": "+17287289312"}
	}, func(deps *Deps) {
		deps.UserAuth = svc
	})
	c := clientFor(t, server)
	passkeyLoginOn(t, c, svc)
	return server, c
}

// sessionCookie snapshots the client's current session cookie value, so
// a test can replay a token that a later mint should have invalidated.
func (c *client) sessionCookie() *http.Cookie {
	c.t.Helper()
	parsed, err := url.Parse(c.base)
	if err != nil {
		c.t.Fatal(err)
	}
	for _, cookie := range c.http.Jar.Cookies(parsed) {
		if cookie.Name == session.CookieName {
			return cookie
		}
	}
	c.t.Fatalf("no %s cookie in the jar", session.CookieName)
	return nil
}

func TestSessionSwitchReMintsForMappedExtension(t *testing.T) {
	_, c := newSwitchServer(t)

	// The mapped identity carries every line: the switcher's render
	// condition (more than one) and the wire the island reads.
	resp, body := c.do(http.MethodGet, "/api/session", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resume: %d %s", resp.StatusCode, body)
	}
	var identity struct {
		Extension  string   `json:"extension"`
		Extensions []string `json:"extensions"`
	}
	if err := json.Unmarshal(body, &identity); err != nil {
		t.Fatalf("resume body %q: %v", body, err)
	}
	if identity.Extension != "1000" || strings.Join(identity.Extensions, ",") != "1000,1001" {
		t.Fatalf("mapped identity = %+v", identity)
	}

	// Snapshot the pre-switch session cookie: the switch must kill this
	// row, and the proof is a raw replay of the old cookie afterwards.
	oldCookie := c.sessionCookie()

	resp, body = c.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("switch: %d %s", resp.StatusCode, body)
	}
	var switched struct {
		Extension  string   `json:"extension"`
		Extensions []string `json:"extensions"`
		Password   string   `json:"password"`
	}
	if err := json.Unmarshal(body, &switched); err != nil {
		t.Fatalf("switch body %q: %v", body, err)
	}
	if switched.Extension != "1001" {
		t.Fatalf("switched extension = %q", switched.Extension)
	}
	if strings.Join(switched.Extensions, ",") != "1000,1001" {
		t.Fatalf("switched extensions = %v (the mapping does not change)", switched.Extensions)
	}
	// Legacy mode answers no credential on switch (the reload's boot
	// probe fetches the row's own payload) — mirror of createSession.
	if switched.Password != "" {
		t.Errorf("switch body must not carry the password: %q", switched.Password)
	}

	// The new cookie resumes on the new line.
	resp, body = c.do(http.MethodGet, "/api/session", nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"extension":"1001"`) {
		t.Fatalf("resume after switch: %d %s", resp.StatusCode, body)
	}
	// The OLD row is dead: a raw replay of the old cookie gets 401.
	oldOnly := &http.Client{Transport: c.http.Transport}
	req, _ := http.NewRequest(http.MethodGet, c.base+"/api/session", nil)
	req.AddCookie(oldCookie)
	if oldResp, _ := oldOnly.Do(req); oldResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("old session cookie must be dead after the switch, got %d", oldResp.StatusCode)
	}
}

func TestSessionSwitchRotatesCsrfAndNamesTheHumanMove(t *testing.T) {
	_, c := newSwitchServer(t)

	var audit bytes.Buffer
	prevLog := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&audit, nil)))
	t.Cleanup(func() { slog.SetDefault(prevLog) })

	resp, _ := c.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("switch: %d", resp.StatusCode)
	}
	// The switch response invalidated the CSRF cookie: the NEXT POST
	// with the stale token must 403 before the adoption.
	req, _ := http.NewRequest(http.MethodPost, c.base+"/messages/send", bytes.NewReader(nil))
	req.Header.Set("X-CSRF-Token", c.token)
	if staleResp, _ := c.http.Do(req); staleResp.StatusCode != http.StatusForbidden {
		t.Errorf("stale CSRF token after switch must 403, got %d", staleResp.StatusCode)
	}
	c.adoptCsrfToken()

	if !strings.Contains(audit.String(), "auth switch") ||
		!strings.Contains(audit.String(), "email="+pkEmail) ||
		!strings.Contains(audit.String(), "1000 -> 1001") {
		t.Errorf("switch audit must name the human move:\n%s", audit.String())
	}
}

func TestSessionSwitchGates(t *testing.T) {
	_, c := newSwitchServer(t)

	// Not mapped to this account: one 403 for a foreign extension.
	resp, body := c.do(http.MethodPost, switchPath, []byte(`{"extension":"1099"}`), "application/json")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("unmapped target: %d %s", resp.StatusCode, body)
	}
	// No dialable characters at all: 400 (letters alone are legal
	// dialable — sanitizeDialable keeps them — and die at the 403
	// membership gate instead).
	resp, _ = c.do(http.MethodPost, switchPath, []byte(`{"extension":"---"}`), "application/json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("undialable extension: %d", resp.StatusCode)
	}
	// Switching to the session's OWN extension is an idempotent no-op:
	// the same cookie stays live (no re-mint, no rotation).
	resp, body = c.do(http.MethodPost, switchPath, []byte(`{"extension":"1000"}`), "application/json")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"extension":"1000"`) {
		t.Fatalf("no-op switch: %d %s", resp.StatusCode, body)
	}
	if resp, body = c.do(http.MethodGet, "/api/session", nil, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("session must survive the no-op switch: %d %s", resp.StatusCode, body)
	}
}

func TestSessionSwitchRequiresMappedLogin(t *testing.T) {
	// An extension login on an identity-wired server: LoginEmail is
	// empty, and the switch must refuse without consulting the mapping
	// (the classic login authenticated exactly one line).
	server, _ := newSwitchServer(t)
	ext := clientFor(t, server)
	ext.login("1000", pkPassword)
	resp, body := ext.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("extension-login switch: %d %s", resp.StatusCode, body)
	}
	// The refusing switch leaves the session untouched.
	if resp, body = ext.do(http.MethodGet, "/api/session", nil, ""); resp.StatusCode != http.StatusOK ||
		!strings.Contains(string(body), `"extension":"1000"`) {
		t.Fatalf("session must survive the refused switch: %d %s", resp.StatusCode, body)
	}
}

func TestSessionSwitchDirectoryRejectedKeepsTheOldLine(t *testing.T) {
	// The PBX directory rejects the TARGET line's credentials: the
	// switch answers 401 and the old session stays minted (a failed
	// switch must leave the operator on their working line).
	svc := newSwitchIdentity(t)
	// The phone API stub accepts 1000's summary (the login verify) and
	// rejects 1001's (the switch verify).
	pbxStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/1001/") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":0}`))
	}))
	t.Cleanup(pbxStub.Close)

	server := newTestServerWithPhoneAPI(t, pbxStub.URL, func(deps *Deps) {
		deps.UserAuth = svc
	})
	c := clientFor(t, server)
	passkeyLoginOn(t, c, svc)

	resp, body := c.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("directory-rejected switch: %d %s", resp.StatusCode, body)
	}
	if resp, body = c.do(http.MethodGet, "/api/session", nil, ""); resp.StatusCode != http.StatusOK ||
		!strings.Contains(string(body), `"extension":"1000"`) {
		t.Fatalf("the old line must stay minted: %d %s", resp.StatusCode, body)
	}
}

func TestSessionSwitchOffAnswers404(t *testing.T) {
	c := newClient(t) // standard fixture: no identity layer, no route
	resp, body := c.do(http.MethodPost, switchPath, []byte(`{"extension":"1001"}`), "application/json")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("mode-off switch: %d %s", resp.StatusCode, body)
	}
}
