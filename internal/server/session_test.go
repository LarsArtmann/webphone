package server

import (
	"bytes"
	"encoding/json/v2"
	"html"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func TestSessionGatesAndFlows(t *testing.T) {
	c := newClient(t)

	// Anonymous partials are gated.
	resp, _ := c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous partial: %d", resp.StatusCode)
	}

	c.login("1001", "pw")

	// Wrong CSRF token is rejected.
	req, _ := http.NewRequest(http.MethodPost, c.base+"/messages/send", bytes.NewReader(nil))
	req.Header.Set("X-CSRF-Token", "wrong")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusForbidden {
		t.Fatalf("bad CSRF token: %d (want 403)", badResp.StatusCode)
	}

	// Signed-in partial renders.
	resp, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "No conversations yet") {
		t.Fatalf("messages partial: %d", resp.StatusCode)
	}
}

// TestLoginRotatesCsrfToken pins the full rotation lifecycle: login deletes
// the CSRF cookie (fixation defense) which kills the page's old token, the
// island adopts the fresh one via GET /api/csrf, and the adopted token
// validates again. Logout rotates the cookie a second time.
func TestLoginRotatesCsrfToken(t *testing.T) {
	c := newClient(t)
	old := c.token

	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	resp, body := c.do(http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("session create: %d %s", resp.StatusCode, body)
	}
	rotated := false
	for _, ck := range resp.Header.Values("Set-Cookie") {
		if strings.HasPrefix(ck, "csrf_token=") && strings.Contains(ck, "Max-Age=0") {
			rotated = true
		}
	}
	if !rotated {
		t.Fatalf("login did not invalidate the CSRF cookie: %v", resp.Header.Values("Set-Cookie"))
	}

	// The page's pre-login token is dead now — POSTs would 403.
	resp, _ = c.do(http.MethodPost, "/messages/send", nil, "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("stale CSRF token after login: %d (want 403)", resp.StatusCode)
	}

	// Adoption hands out a different token, and it validates.
	c.adoptCsrfToken()
	if c.token == old {
		t.Fatal("adoption returned the stale token")
	}
	resp, _ = c.do(http.MethodPost, "/messages/send", nil, "")
	if resp.StatusCode == http.StatusForbidden {
		t.Fatal("adopted CSRF token rejected — island would be bricked")
	}

	// Logout rotates once more so the token never outlives its session.
	resp, _ = c.do(http.MethodDelete, "/api/session", nil, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: %d", resp.StatusCode)
	}
	rotated = false
	for _, ck := range resp.Header.Values("Set-Cookie") {
		if strings.HasPrefix(ck, "csrf_token=") && strings.Contains(ck, "Max-Age=0") {
			rotated = true
		}
	}
	if !rotated {
		t.Fatalf("logout did not invalidate the CSRF cookie: %v", resp.Header.Values("Set-Cookie"))
	}
}

// TestShellRendersValidJSONCSRFHxHeaders guards the hx-headers wiring: htmx
// JSON.parses the attribute for every request it sends, so the shell must
// render JSON-valid content that survives the single templ attribute
// escaping. A regression to string concatenation (breaks on exotic tokens)
// or to the pre-escaped httputil helper (double-escapes in templ context)
// would drop CSRF protection from every HTMX request — this test fails on
// both.
func TestShellRendersValidJSONCSRFHxHeaders(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shell status %d", resp.StatusCode)
	}
	m := regexp.MustCompile(`hx-headers="([^"]+)"`).FindSubmatch(body)
	if m == nil {
		t.Fatal("hx-headers attribute missing from shell")
	}
	var hdrs map[string]string
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &hdrs); err != nil {
		t.Fatalf("hx-headers not JSON after attribute unescape: %v (%q)", err, m[1])
	}
	if hdrs["X-CSRF-Token"] == "" {
		t.Error("X-CSRF-Token missing or empty in hx-headers")
	}
}
