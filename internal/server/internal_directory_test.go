// Internal directory endpoint tests: the gates (mode-off fail-closed,
// bearer secret, loopback posture) and the wire answer (derived
// credentials of the live sessions only, empty list for unknown
// extensions).
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/sipcred"
)

const dirSecret = "dir-derive-secret"

func newDirectoryServer(t *testing.T, sip config.SIP) *testServer {
	t.Helper()
	return newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.SIP = sip
	})
}

func TestInternalDirectoryFailsClosedWithoutASecret(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{})
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/internal/directory/1001", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("no-secret request: %d, want 503 (fail closed)", res.StatusCode)
	}
}

func TestInternalDirectoryAnswersTheDerivedCredentials(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{CredentialSecret: dirSecret})
	c := clientFor(t, ts)
	resp := c.loginRaw("1001", "pw")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("login: %d", resp.StatusCode)
	}
	token := sessionTokenFromResponse(t, resp)
	want := sipcred.Derive(dirSecret, token)

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/internal/directory/1001", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+dirSecret)
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, res)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("directory: %d %s", res.StatusCode, body)
	}
	var got struct {
		Credentials []string `json:"credentials"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("body %q: %v", body, err)
	}
	if len(got.Credentials) != 1 || got.Credentials[0] != want {
		t.Fatalf("credentials = %v, want [%s]", got.Credentials, want)
	}
}

func TestInternalDirectoryRejectsBadSecrets(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{CredentialSecret: dirSecret})
	for _, tc := range []struct{ name, auth string }{
		{"missing", ""},
		{"wrong", "Bearer not-the-secret"},
	} {
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/internal/directory/1001", nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.auth != "" {
			req.Header.Set("Authorization", tc.auth)
		}
		res, err := ts.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: %d, want 401", tc.name, res.StatusCode)
		}
	}
}

func TestInternalDirectoryUnknownExtensionAnswersEmpty(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{CredentialSecret: dirSecret})
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/internal/directory/4242", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+dirSecret)
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, res)
	if res.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != `{"credentials":[]}` {
		t.Fatalf("unknown extension: %d %s, want 200 {\"credentials\":[]}", res.StatusCode, body)
	}
}

func TestInternalDirectoryInvalidExtensionIs400(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{CredentialSecret: dirSecret})
	// %20 sanitizes to an empty extension (no dialable characters) — the
	// one class ParseExtension rejects.
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/internal/directory/%20", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+dirSecret)
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid extension: %d, want 400", res.StatusCode)
	}
}

// TestInternalDirectoryLoopbackOnly pins the posture: a non-loopback
// peer is refused even with the correct bearer (defense in depth — the
// module wires the bridge on the same host).
func TestInternalDirectoryLoopbackOnly(t *testing.T) {
	ts := newDirectoryServer(t, config.SIP{CredentialSecret: dirSecret})
	req := httptest.NewRequest(http.MethodGet, "/internal/directory/1001", nil)
	req.Header.Set("Authorization", "Bearer "+dirSecret)
	req.RemoteAddr = "192.0.2.1:4444"
	rec := httptest.NewRecorder()
	ts.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-loopback peer: %d, want 403", rec.Code)
	}
}

// sessionTokenFromResponse extracts the minted session token from a
// login response's Set-Cookie header.
func sessionTokenFromResponse(t *testing.T, resp *http.Response) string {
	t.Helper()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "webphone_session" {
			return cookie.Value
		}
	}
	t.Fatal("login response carried no webphone_session cookie")
	return ""
}
