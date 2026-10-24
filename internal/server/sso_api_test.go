// The OIDC/SSO handler surface: mode-off 404s, the begin hand-off (state,
// nonce and PKCE parked in short-lived cookies), the callback's state
// gate, and the one place an IdP identity becomes a webphone session.
// The IdP itself is a stub (the real wire protocol lives in internal/sso's
// suite), so what is under test here is the HANDLER contract.
package server

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sso"
)

type ssoStub struct {
	challenge sso.Challenge
	user      authmap.User
	password  string
	err       error

	gotCode     string
	gotVerifier string
	gotNonce    string
}

func (s *ssoStub) Challenge(context.Context) (sso.Challenge, error) {
	return s.challenge, s.err
}

func (s *ssoStub) Complete(_ context.Context, code, verifier, nonce string) (authmap.User, string, error) {
	s.gotCode, s.gotVerifier, s.gotNonce = code, verifier, nonce
	return s.user, s.password, s.err
}

func newSSOServer(t *testing.T, stub *ssoStub) *testServer {
	t.Helper()
	return newTestServerWithConfig(t, "", nil, func(deps *Deps) {
		deps.SSO = stub
	})
}

func happySSOStub() *ssoStub {
	return &ssoStub{
		challenge: sso.Challenge{
			URL:      "https://auth.example.org/authorize?client_id=webphone",
			State:    "state-1",
			Nonce:    "nonce-1",
			Verifier: "verifier-1",
		},
		user: authmap.User{
			Email:       "lars@example.com",
			DisplayName: "Lars",
			Extensions:  []domain.Extension{domain.MustParseExtension("1000")},
		},
		password: "dir-password",
	}
}

func cookieFor(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: ssoCookiePath}
}

func TestSSOOffAnswersTheStyled404(t *testing.T) {
	server := newTestServer(t) // standard fixture: no SSO
	client := server.Client()
	for _, path := range []string{"/api/auth/oauth2/begin", "/api/auth/oauth2/callback"} {
		resp, err := client.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body := readAll(t, resp)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404 (mode off)", path, resp.StatusCode)
		}
		if !strings.Contains(string(body), "WebPhone") {
			t.Errorf("%s: the 404 must stay the styled shell, got %q", path, body)
		}
	}
}

func TestSSOBeginParksTheAttemptAndRedirects(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse // observe the 302, do not follow it
	}}
	resp, err := client.Get(server.URL + "/api/auth/oauth2/begin")
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("begin: status %d, want 302", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != stub.challenge.URL {
		t.Errorf("Location = %q, want the IdP authorize URL", got)
	}
	cookies := map[string]*http.Cookie{}
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie
	}
	for name, want := range map[string]string{
		ssoStateCookie:    stub.challenge.State,
		ssoNonceCookie:    stub.challenge.Nonce,
		ssoVerifierCookie: stub.challenge.Verifier,
	} {
		cookie, ok := cookies[name]
		if !ok {
			t.Fatalf("begin did not set %s", name)
		}
		if cookie.Value != want {
			t.Errorf("%s = %q, want %q", name, cookie.Value, want)
		}
		if !cookie.HttpOnly || cookie.MaxAge <= 0 || cookie.Path != ssoCookiePath {
			t.Errorf("%s cookie attributes = %+v", name, cookie)
		}
	}
}

func TestSSOCallbackMintsTheSession(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse // the 303 IS the assertion
	}}

	u, err := url.Parse(server.URL + "/api/auth/oauth2/callback")
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("code", "good-code")
	query.Set("state", stub.challenge.State)
	u.RawQuery = query.Encode()
	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(cookieFor(ssoStateCookie, stub.challenge.State))
	request.AddCookie(cookieFor(ssoNonceCookie, stub.challenge.Nonce))
	request.AddCookie(cookieFor(ssoVerifierCookie, stub.challenge.Verifier))

	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("callback: status %d, want 303", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/" {
		t.Errorf("Location = %q, want /", got)
	}
	if stub.gotCode != "good-code" || stub.gotVerifier != stub.challenge.Verifier || stub.gotNonce != stub.challenge.Nonce {
		t.Errorf("Complete arguments = (%q, %q, %q), want the code/cookie values",
			stub.gotCode, stub.gotVerifier, stub.gotNonce)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == session.CookieName && cookie.Value != "" {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil {
		t.Fatal("callback minted no webphone session cookie")
	}
	// The 303 clears the attempt cookies so a replay has nothing to use.
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case ssoStateCookie, ssoNonceCookie, ssoVerifierCookie:
			if cookie.MaxAge >= 0 {
				t.Errorf("%s must be expired on the callback, got MaxAge=%d", cookie.Name, cookie.MaxAge)
			}
		}
	}
}

func TestSSOCallbackStateMismatchReturnsToTheLoginCard(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	request, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/auth/oauth2/callback?code=good-code&state=forged", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(cookieFor(ssoStateCookie, stub.challenge.State))
	request.AddCookie(cookieFor(ssoNonceCookie, stub.challenge.Nonce))
	request.AddCookie(cookieFor(ssoVerifierCookie, stub.challenge.Verifier))

	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Location"); got != "/?sso_error=stale" {
		t.Errorf("Location = %q, want the stale-attempt verdict", got)
	}
	if stub.gotCode != "" {
		t.Error("a forged state must never reach the token exchange")
	}
}

func TestSSOCallbackFailuresMapOntoTheLoginCardVocabulary(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"unmapped email", errorfamily.NewRejection("sso.unmapped_email", "no mapping"), "/?sso_error=unmapped"},
		{"unverified email", errorfamily.NewRejection("sso.email_unverified", "unverified"), "/?sso_error=unverified"},
		{"bad secret", errorfamily.NewRejection("sso.client_secret.read", "gone"), "/?sso_error=config"},
		{"idp down", errorfamily.NewRejection("sso.discovery", "unreachable"), "/?sso_error=unavailable"},
		{"anything else", errorfamily.NewRejection("sso.code", "no code"), "/?sso_error=failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := happySSOStub()
			stub.err = tc.err
			server := newSSOServer(t, stub)
			client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}}
			request, err := http.NewRequest(http.MethodGet,
				server.URL+"/api/auth/oauth2/callback?code=c&state="+stub.challenge.State, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.AddCookie(cookieFor(ssoStateCookie, stub.challenge.State))
			request.AddCookie(cookieFor(ssoNonceCookie, stub.challenge.Nonce))
			request.AddCookie(cookieFor(ssoVerifierCookie, stub.challenge.Verifier))
			resp, err := client.Do(request)
			if err != nil {
				t.Fatalf("callback: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if got := resp.Header.Get("Location"); got != tc.want {
				t.Errorf("Location = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSSOCallbackWithoutAttemptCookiesIsStale(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(server.URL + "/api/auth/oauth2/callback?code=c&state=whatever")
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if got := resp.Header.Get("Location"); got != "/?sso_error=stale" {
		t.Errorf("Location = %q, want the stale-attempt verdict", got)
	}
}

func TestSSOErrorCodeVocabulary(t *testing.T) {
	for code, want := range map[string]string{
		"sso.unmapped_email":            "unmapped",
		"authmap.password_file.missing": "directory",
		"authmap.password_file.read":    "directory",
		"authmap.password_file.empty":   "directory",
		"sso.email_unverified":          "unverified",
		"sso.state.mismatch":            "stale",
		"sso.nonce.mismatch":            "stale",
		"sso.attempt.missing":           "stale",
		"sso.client_secret.read":        "config",
		"sso.client_secret.empty":       "config",
		"sso.exchange":                  "config",
		"sso.discovery":                 "unavailable",
		"sso.userinfo":                  "unavailable",
		"sso.id_token.verify":           "unavailable",
		"sso.something.unclassified":    "failed",
	} {
		if got := ssoErrorCode(errorfamily.NewRejection(code, "x")); got != want {
			t.Errorf("ssoErrorCode(%q) = %q, want %q", code, got, want)
		}
	}
}
