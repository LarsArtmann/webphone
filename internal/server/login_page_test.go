package server

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// Login page tests: the library-owned passkey front door (cqrs-htmx
// loginpage at /login) — the nonce CSP lane (the ONE surface allowed
// inline scripts, each carrying the per-request nonce), the live-session
// redirect, the passkey-off styled 404, and the CSRF meta the embedded
// ceremony JS posts with.

// noncePattern matches a base64url CSP nonce (httputil mints one per
// request; anything else in the directive would be a CSP break-out).
var noncePattern = regexp.MustCompile(`^script-src 'self' 'nonce-([A-Za-z0-9_-]+)'$`)

func TestLoginPageServesBehindANonceCSP(t *testing.T) {
	server, _ := newPasskeyServer(t)
	c := clientFor(t, server)

	resp, body := c.do(http.MethodGet, "/login", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/login status %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("/login Cache-Control = %q, want no-store (auth surface)", got)
	}

	scriptSrc := regexp.MustCompile(`script-src[^;]*`).FindString(resp.Header.Get("Content-Security-Policy"))
	m := noncePattern.FindStringSubmatch(strings.TrimSpace(scriptSrc))
	if m == nil {
		t.Fatalf("script-src must be exactly 'self' + one nonce, got %q", scriptSrc)
	}
	nonce := m[1]
	for _, forbidden := range []string{"unsafe-inline", "'sha256-"} {
		if strings.Contains(scriptSrc, forbidden) {
			t.Errorf("script-src carries %q — the login lane allows ONLY the nonce", forbidden)
		}
	}
	// Every inline script on the page must carry the SAME nonce — one
	// unnonced tag is a dead script under this CSP (the ceremony would
	// silently no-op in a browser).
	inline := regexp.MustCompile(`(?s)<script((?:[^>]|\n)*?)>`).FindAllStringSubmatch(string(body), -1)
	if len(inline) < 2 {
		t.Fatalf("expected the ceremony JS + config JSON inline scripts, found %d script tags", len(inline))
	}
	nonced := 0
	for _, tag := range inline {
		if strings.Contains(tag[1], "src=") {
			continue // external script, allowed by 'self'
		}
		if !strings.Contains(tag[1], `nonce="`+nonce+`"`) {
			t.Errorf("inline script without the CSP nonce: %.80s", tag[0])
		}
		nonced++
	}
	if nonced == 0 {
		t.Error("no inline script carried the nonce — the page would be inert in a browser")
	}

	// The library page's wire contract: the form the embedded JS binds
	// plus the config blob it reads, and the CSRF meta it posts with.
	for _, want := range []string{
		`id="lp-login-form"`,
		`id="lp-email"`,
		`id="loginpage-config"`,
		`<meta name="csrf-token"`,
		`/assets/login.css`,
		`/api/auth/webauthn/login/begin`,
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("/login lacks %s:\n%.400s", want, body)
		}
	}
	// The dead island form ids must never leak onto the library page.
	if strings.Contains(string(body), "passkey-login-form") {
		t.Error("/login renders the dead hand-rolled form id")
	}
}

func TestLoginPageRedirectsALiveSessionHome(t *testing.T) {
	server, _ := newPasskeyServer(t)
	c := clientFor(t, server)
	c.login("1000", pkPassword)

	// The shared client follows redirects; the 303 itself is the
	// contract, so stop at the first response.
	c.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	defer func() { c.http.CheckRedirect = nil }()
	resp, _ := c.do(http.MethodGet, "/login", nil, "")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("/login with a live session: status %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Errorf("/login redirect Location = %q, want /", loc)
	}
}

func TestLoginPageAnswersTheStyled404WhenPasskeyIsOff(t *testing.T) {
	c := newClient(t) // standard fixture: no UserAuth
	resp, body := c.do(http.MethodGet, "/login", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("/login with the mode off: status %d, want 404", resp.StatusCode)
	}
	if !strings.Contains(string(body), "WebPhone") {
		t.Errorf("the off-mode 404 must stay the styled shell, got %q", body)
	}
}
