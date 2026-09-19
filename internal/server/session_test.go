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

	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	resp, body := c.do(http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("session create: %d %s", resp.StatusCode, body)
	}

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
