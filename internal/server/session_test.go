package server

import (
	"bytes"
	"encoding/json/v2"
	"html"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
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
	resp, body := c.do(http.MethodGet, "/partials/messages", nil, "")
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

	// The page's pre-login token is dead now; POSTs would 403.
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
		t.Fatal("adopted CSRF token rejected, island would be bricked")
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

// TestCSRFTrustsTheFrontingProxy pins the TLS-fronting deployment shape.
// nginx terminates TLS, so a truthful browser POST arrives with
// Origin: https://host and Sec-Fetch-Site: same-origin while the listener
// sees plain HTTP. Unconfigured, the attestation check reads that as
// forged and 403s every POST, logins included (the bug v2.0.0 shipped);
// with the loopback proxy and the vhost origin trusted, it passes.
func TestCSRFTrustsTheFrontingProxy(t *testing.T) {
	for name, tc := range map[string]struct {
		csrf config.CSRF
		want int
	}{
		"unconfigured rejects the fronted origin": {
			want: http.StatusForbidden,
		},
		"trusted proxy and origin accept it": {
			csrf: config.CSRF{
				TrustedProxies: []string{"127.0.0.1"},
				TrustedOrigins: []string{"https://pbx.test"},
			},
			want: http.StatusCreated,
		},
	} {
		t.Run(name, func(t *testing.T) {
			server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
				d.Config.CSRF = tc.csrf
			})
			c := clientFor(t, server)

			req, err := http.NewRequest(http.MethodGet, c.base+"/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Host = "pbx.test"
			resp, err := c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body := readAll(t, resp)
			match := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindSubmatch(body)
			if match == nil {
				t.Fatal("no csrf token in page")
			}
			// A https trusted origin marks the CSRF cookie Secure; a
			// plain-HTTP test harness must carry it by hand (a real
			// browser received it over the TLS hop with the vhost).
			cookieHeader := ""
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "csrf_token" && cookie.Secure {
					cookieHeader = "csrf_token=" + cookie.Value
				}
			}

			payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
			req, err = http.NewRequest(http.MethodPost, c.base+"/api/session", bytes.NewReader(payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Host = "pbx.test"
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "https://pbx.test")
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			req.Header.Set("Sec-Fetch-Mode", "cors")
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-CSRF-Token", string(match[1]))
			if cookieHeader != "" {
				req.Header.Set("Cookie", cookieHeader)
			}
			resp, err = c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want {
				t.Fatalf("fronted login: %d (want %d)", resp.StatusCode, tc.want)
			}
		})
	}
}

// TestCSRFSecureFollowsTrustedOrigins pins the cookie-hygiene rule: an
// https trusted origin means TLS fronting, so the CSRF cookie must carry
// the Secure flag; plain-http origins (loopback dev) must not set it.
func TestCSRFSecureFollowsTrustedOrigins(t *testing.T) {
	for name, tc := range map[string]struct {
		origins   []string
		wantStray bool
	}{
		"https origin sets Secure": {origins: []string{"https://pbx.test"}, wantStray: true},
		"http origin stays plain":  {origins: []string{"http://localhost"}, wantStray: false},
		"unconfigured stays plain": {wantStray: false},
	} {
		t.Run(name, func(t *testing.T) {
			server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
				d.Config.CSRF = config.CSRF{TrustedProxies: []string{"127.0.0.1"}, TrustedOrigins: tc.origins}
			})
			req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp.Body.Close() }()
			cookies := resp.Cookies()
			var csrfCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == "csrf_token" {
					csrfCookie = cookie
					break
				}
			}
			if csrfCookie == nil {
				t.Fatalf("no csrf_token cookie in response (cookies: %v)", cookies)
			}
			if got := csrfCookie.Secure; got != tc.wantStray {
				t.Errorf("csrf_token Secure = %v, want %v", got, tc.wantStray)
			}
		})
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
