// The native-shell sign-in handoff's HTTP surface: the capability
// probe, the begin hop's shell form, the callback's bounce and the
// ticket redemption, driven through the real mux with the IdP stubbed
// (the same ssoStub the browser-lane tests ride). Each of the handoff's
// three fences is pinned here as the SERVER's refusal: what happens when
// the ticket is forged, replayed, presented without its proof, or spent
// from a browser.
package server

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/server/auth"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sso"
)

var (
	// 43 characters each: the shape the shell generates (32 random
	// bytes in unpadded base64url).
	handoffNonce = strings.Repeat("n", 43)
	handoffProof = strings.Repeat("p", 43)
	// handoffTicketShape is what the server mints.
	handoffTicketShape = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
)

// countingSSOStub counts the IdP challenges: a refused begin hop must
// cost no IdP round trip at all.
type countingSSOStub struct {
	ssoStub
	challenges int
}

func (s *countingSSOStub) Challenge(ctx context.Context) (sso.Challenge, error) {
	s.challenges++
	return s.ssoStub.Challenge(ctx)
}

// handoffBrowser stands in for the system browser: a cookie jar that
// carries the attempt cookies from the begin hop to the callback, and no
// redirect following (every hop's own answer is the assertion).
func handoffBrowser(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

// bounceHandoff drives a well-formed shell attempt from the begin hop to
// the callback and returns the callback's response (left open for the
// caller's cookie assertions) with the return link it carries.
func bounceHandoff(t *testing.T, server *testServer, browser *http.Client) (*http.Response, *url.URL) {
	t.Helper()
	resp, err := browser.Get(server.URL + "/api/auth/oauth2/begin?shell=android&app_state=" + handoffNonce + "&proof=" + handoffProof)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("begin: status %d, want 302 to the IdP", resp.StatusCode)
	}
	resp, err = browser.Get(server.URL + "/api/auth/oauth2/callback?code=good-code&state=" + happySSOStub().challenge.State)
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("callback: status %d, want 303", resp.StatusCode)
	}
	link, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("callback Location %q: %v", resp.Header.Get("Location"), err)
	}
	return resp, link
}

// redemption is the shell's finish request; an empty header value means
// the header is not sent.
func redemption(t *testing.T, server *testServer, ticket, shell, proof string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/auth/oauth2/finish?ticket="+url.QueryEscape(ticket), nil)
	if err != nil {
		t.Fatal(err)
	}
	if shell != "" {
		request.Header.Set("X-WebPhone-Shell", shell)
	}
	if proof != "" {
		request.Header.Set("X-WebPhone-Shell-Proof", proof)
	}
	return request
}

// send performs one request on a jar-less client that follows nothing.
func send(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s: %v", request.URL.Path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// cookieNamed is the response's Set-Cookie of that name, or nil.
func cookieNamed(resp *http.Response, name string) *http.Cookie {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// doBounce hands a well-formed shell attempt through to its return link
// and returns the bounce response and the link.
func doBounce(t *testing.T, stub *ssoStub) (*testServer, *http.Client, *http.Response, *url.URL) {
	t.Helper()
	server := newSSOServer(t, stub)
	browser := handoffBrowser(t)
	resp, link := bounceHandoff(t, server, browser)
	return server, browser, resp, link
}

func TestSSOShellProbeAdvertisesTheContract(t *testing.T) {
	server := newSSOServer(t, happySSOStub())
	resp, err := server.Client().Get(server.URL + "/api/auth/oauth2/shell")
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	body := readAll(t, resp)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("probe: status %d, want 204", resp.StatusCode)
	}
	if len(body) != 0 {
		t.Errorf("probe: body %q, want none", body)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("probe: Cache-Control = %q, want no-store (a stack that upgrades must flip the shell's lane)", got)
	}
	if len(resp.Cookies()) != 0 {
		t.Errorf("probe: set %d cookie(s), want none (it hands out nothing)", len(resp.Cookies()))
	}
}

func TestSSOBeginShellFormParksTheBoundMarker(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	resp, err := handoffBrowser(t).Get(server.URL + "/api/auth/oauth2/begin?shell=android&app_state=" + handoffNonce + "&proof=" + handoffProof)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("begin: status %d, want 302", resp.StatusCode)
	}
	marker := cookieNamed(resp, "wp_sso_shell")
	if marker == nil {
		t.Fatal("the shell form set no wp_sso_shell cookie")
	}
	if !strings.HasPrefix(marker.Value, handoffNonce+".") {
		t.Errorf("wp_sso_shell = %q, want the nonce and the proof's digest", marker.Value)
	}
	if strings.Contains(marker.Value, handoffProof) {
		t.Error("wp_sso_shell carries the raw proof: the server must keep only its digest")
	}
	if !marker.HttpOnly || marker.MaxAge <= 0 || marker.Path != auth.SSOCallbackPath || marker.SameSite != http.SameSiteLaxMode {
		t.Errorf("wp_sso_shell attributes = %+v, want the attempt cookies' (HttpOnly, Lax, path-scoped, short-lived)", marker)
	}
	// The attempt cookies are the browser lane's, unchanged.
	for name, want := range map[string]string{
		auth.SSOStateCookie:    stub.challenge.State,
		auth.SSONonceCookie:    stub.challenge.Nonce,
		auth.SSOVerifierCookie: stub.challenge.Verifier,
	} {
		cookie := cookieNamed(resp, name)
		if cookie == nil || cookie.Value != want {
			t.Errorf("%s = %+v, want %q", name, cookie, want)
		}
	}
}

func TestSSOBeginShellFormRefusesMalformedHandshakes(t *testing.T) {
	for name, query := range map[string]string{
		"an unknown shell":      "shell=ios&app_state=" + handoffNonce + "&proof=" + handoffProof,
		"a missing nonce":       "shell=android&proof=" + handoffProof,
		"a short nonce":         "shell=android&app_state=short&proof=" + handoffProof,
		"a missing proof":       "shell=android&app_state=" + handoffNonce,
		"an off-alphabet proof": "shell=android&app_state=" + handoffNonce + "&proof=" + strings.Repeat("p", 42) + "%2B",
	} {
		t.Run(name, func(t *testing.T) {
			stub := &countingSSOStub{ssoStub: *happySSOStub()}
			server := newTestServerWithConfig(t, "", nil, func(deps *Deps) { deps.SSO = stub })
			resp, err := handoffBrowser(t).Get(server.URL + "/api/auth/oauth2/begin?" + query)
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			body := readAll(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status %d, want 400", resp.StatusCode)
			}
			if !strings.Contains(string(body), "Start the sign-in again from the app.") {
				t.Errorf("body %q, want the actionable sentence", body)
			}
			if strings.Contains(string(body), "sso.handoff") {
				t.Errorf("body %q leaks the internal error code (it belongs in the journal)", body)
			}
			if len(resp.Cookies()) != 0 {
				t.Errorf("a refused handshake set %d cookie(s), want none", len(resp.Cookies()))
			}
			if stub.challenges != 0 {
				t.Errorf("a refused handshake cost %d IdP challenge(s), want none", stub.challenges)
			}
		})
	}
}

func TestSSOBeginPlainFormExpiresAPlantedShellMarker(t *testing.T) {
	server := newSSOServer(t, happySSOStub())

	// Parity first: a plain begin with no marker around answers exactly
	// as the pre-handoff server did.
	resp, err := handoffBrowser(t).Get(server.URL + "/api/auth/oauth2/begin")
	if err != nil {
		t.Fatalf("plain begin: %v", err)
	}
	_ = resp.Body.Close()
	if cookieNamed(resp, "wp_sso_shell") != nil || len(resp.Cookies()) != 3 {
		t.Errorf("a plain begin set %d cookie(s) including a shell marker, want exactly the three attempt cookies", len(resp.Cookies()))
	}

	// A cross-site navigation to the shell form can plant the marker in
	// a browser that never asked for the handoff; the plain begin must
	// not let the next login bounce to a scheme that browser cannot
	// open.
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/auth/oauth2/begin", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(cookieFor("wp_sso_shell", "planted.marker"))
	resp = send(t, request)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("plain begin with a planted marker: status %d, want 302", resp.StatusCode)
	}
	planted := cookieNamed(resp, "wp_sso_shell")
	if planted == nil || planted.MaxAge >= 0 {
		t.Errorf("the planted marker was not expired: %+v", planted)
	}
}

func TestSSOCallbackOnAShellAttemptBouncesATicket(t *testing.T) {
	server, browser, resp, link := doBounce(t, happySSOStub())

	if link.Scheme != "cloud.larsartmann.webphone" || link.Host != "auth" || link.Path != "/finish" {
		t.Errorf("return link = %s://%s%s, want the shell's pinned return address", link.Scheme, link.Host, link.Path)
	}
	if got := link.Query().Get("app_state"); got != handoffNonce {
		t.Errorf("return link app_state = %q, want the shell's nonce", got)
	}
	if got := link.Query().Get("ticket"); !handoffTicketShape.MatchString(got) {
		t.Errorf("return link ticket = %q, want 43 characters of unpadded base64url", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store (the Location carries a bearer ticket)", got)
	}
	// The browser gets NO session, and loses the whole attempt.
	if cookieNamed(resp, session.CookieName) != nil {
		t.Error("the bounce set a session cookie: the sign-in must land in the shell's jar, not the browser's")
	}
	for _, name := range []string{auth.SSOStateCookie, auth.SSONonceCookie, auth.SSOVerifierCookie, "wp_sso_shell"} {
		if cookie := cookieNamed(resp, name); cookie == nil || cookie.MaxAge >= 0 {
			t.Errorf("%s was not expired by the bounce: %+v", name, cookie)
		}
	}
	probe, err := browser.Get(server.URL + "/api/session")
	if err != nil {
		t.Fatalf("session probe: %v", err)
	}
	_ = probe.Body.Close()
	if probe.StatusCode != http.StatusUnauthorized {
		t.Errorf("the browser resumes with %d after the bounce, want 401 (it stays signed out)", probe.StatusCode)
	}
}

func TestSSOFinishRedeemsTheTicketOnce(t *testing.T) {
	server, _, _, link := doBounce(t, happySSOStub())
	ticket := link.Query().Get("ticket")

	shell := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := shell.Do(redemption(t, server, ticket, "android", handoffProof))
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/" {
		t.Fatalf("finish: %d %q, want 303 /", resp.StatusCode, resp.Header.Get("Location"))
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	var minted *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == session.CookieName && cookie.Value != "" {
			minted = cookie
		}
	}
	if minted == nil {
		t.Fatal("the redemption set no session cookie")
	}
	if minted.Path != "/" {
		t.Errorf("session cookie path = %q, want the root", minted.Path)
	}

	// The cookie is a real session in the shell's jar: the island's boot
	// probe resumes it, and the identity is the IdP's.
	resume := send(t, mustRequest(t, http.MethodGet, server.URL+"/api/session", minted))
	body := readAll(t, resume)
	if resume.StatusCode != http.StatusOK || !strings.Contains(string(body), `"extension":"1000"`) {
		t.Errorf("resume with the redeemed cookie: %d %s", resume.StatusCode, body)
	}

	// Single use: the replay finds nothing and lands on the stack's own
	// "attempt expired" sentence.
	replay, err := shell.Do(redemption(t, server, ticket, "android", handoffProof))
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	defer func() { _ = replay.Body.Close() }()
	if replay.StatusCode != http.StatusSeeOther || replay.Header.Get("Location") != "/login?sso_error=stale" {
		t.Errorf("replay: %d %q, want 303 /login?sso_error=stale", replay.StatusCode, replay.Header.Get("Location"))
	}
	if cookieNamed(replay, session.CookieName) != nil {
		t.Error("the replay set a session cookie")
	}
}

func TestSSOFinishNeedsTheShellHeader(t *testing.T) {
	server, _, _, link := doBounce(t, happySSOStub())
	ticket := link.Query().Get("ticket")

	// A navigation or an image: no custom header, no session, and not
	// even a hint about the ticket (a cross-site page must not spend a
	// stolen ticket into a browser's jar).
	resp := send(t, redemption(t, server, ticket, "", handoffProof))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("finish without the shell header: status %d, want 400", resp.StatusCode)
	}
	if cookieNamed(resp, session.CookieName) != nil {
		t.Error("a header-less redemption set a session cookie")
	}
	// ...and it must not burn the ticket the rightful shell is about to
	// present.
	after := send(t, redemption(t, server, ticket, "android", handoffProof))
	if after.StatusCode != http.StatusSeeOther || after.Header.Get("Location") != "/" {
		t.Errorf("the rightful redemption after a header-less one: %d %q, want 303 /", after.StatusCode, after.Header.Get("Location"))
	}
}

func TestSSOFinishRefusesAnUnknownShellHeader(t *testing.T) {
	server, _, _, link := doBounce(t, happySSOStub())
	resp := send(t, redemption(t, server, link.Query().Get("ticket"), "ios", handoffProof))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("finish with an unknown shell: status %d, want 400", resp.StatusCode)
	}
}

func TestSSOFinishRefusesAForeignProof(t *testing.T) {
	server, _, _, link := doBounce(t, happySSOStub())
	ticket := link.Query().Get("ticket")

	// A rogue app caught the return link and holds the ticket alone.
	rogue := send(t, redemption(t, server, ticket, "android", strings.Repeat("x", 43)))
	if rogue.StatusCode != http.StatusForbidden {
		t.Errorf("finish with a foreign proof: status %d, want 403", rogue.StatusCode)
	}
	if cookieNamed(rogue, session.CookieName) != nil {
		t.Error("a foreign-proof redemption set a session cookie")
	}
	// The refusal leaves the ticket alive for the shell that started the
	// attempt.
	rightful := send(t, redemption(t, server, ticket, "android", handoffProof))
	if rightful.StatusCode != http.StatusSeeOther || rightful.Header.Get("Location") != "/" {
		t.Errorf("the rightful redemption after a refused one: %d %q, want 303 /", rightful.StatusCode, rightful.Header.Get("Location"))
	}
	// Without any proof header the ticket is equally bound.
	server2, _, _, link2 := doBounce(t, happySSOStub())
	bare := send(t, redemption(t, server2, link2.Query().Get("ticket"), "android", ""))
	if bare.StatusCode != http.StatusForbidden {
		t.Errorf("finish without the proof header: status %d, want 403", bare.StatusCode)
	}
}

func TestSSOFinishStaleTicketsLandOnTheLoginCard(t *testing.T) {
	server := newSSOServer(t, happySSOStub())
	for name, ticket := range map[string]string{
		"never issued": strings.Repeat("z", 43),
		"malformed":    "not-a-ticket",
		"missing":      "",
	} {
		resp := send(t, redemption(t, server, ticket, "android", handoffProof))
		if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login?sso_error=stale" {
			t.Errorf("%s ticket: %d %q, want 303 /login?sso_error=stale", name, resp.StatusCode, resp.Header.Get("Location"))
		}
		if cookieNamed(resp, session.CookieName) != nil {
			t.Errorf("%s ticket: set a session cookie", name)
		}
	}
}

func TestSSOCallbackWithAMalformedMarkerFailsClosed(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	request, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/auth/oauth2/callback?code=good-code&state="+stub.challenge.State, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(cookieFor(auth.SSOStateCookie, stub.challenge.State))
	request.AddCookie(cookieFor(auth.SSONonceCookie, stub.challenge.Nonce))
	request.AddCookie(cookieFor(auth.SSOVerifierCookie, stub.challenge.Verifier))
	request.AddCookie(cookieFor("wp_sso_shell", "garbage"))
	resp := send(t, request)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("callback with a malformed marker: status %d, want 303", resp.StatusCode)
	}
	// Neither jar gets the sign-in: no cookie for a browser that may be
	// a shell's Custom Tab, and no ticket for a link nobody could craft
	// from garbage.
	if got := resp.Header.Get("Location"); got != "/login?sso_error=failed" {
		t.Errorf("Location = %q, want the login card's failure verdict", got)
	}
	if cookieNamed(resp, session.CookieName) != nil {
		t.Error("a malformed marker minted a session cookie")
	}
	if marker := cookieNamed(resp, "wp_sso_shell"); marker == nil || marker.MaxAge >= 0 {
		t.Errorf("the malformed marker was not expired: %+v", marker)
	}
}

func TestSSOCallbackFailureOnAShellAttemptClearsTheMarker(t *testing.T) {
	stub := happySSOStub()
	server := newSSOServer(t, stub)
	browser := handoffBrowser(t)
	resp, err := browser.Get(server.URL + "/api/auth/oauth2/begin?shell=android&app_state=" + handoffNonce + "&proof=" + handoffProof)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	_ = resp.Body.Close()
	// The IdP answers with a forged state.
	resp, err = browser.Get(server.URL + "/api/auth/oauth2/callback?code=good-code&state=forged")
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	_ = resp.Body.Close()
	if got := resp.Header.Get("Location"); got != "/login?sso_error=stale" {
		t.Errorf("Location = %q, want the stale-attempt verdict", got)
	}
	if marker := cookieNamed(resp, "wp_sso_shell"); marker == nil || marker.MaxAge >= 0 {
		t.Errorf("the failed attempt kept its shell marker: %+v", marker)
	}
}

// mustRequest builds a request carrying one cookie.
func mustRequest(t *testing.T, method, target string, cookie *http.Cookie) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(cookie)
	return request
}
