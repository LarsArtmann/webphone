package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json/v2"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

// contractIDs are the DOM elements the upstream browser E2E drives (see
// AGENTS.md). The Go app must keep serving every one of them.
var contractIDs = []string{
	"reg-status", "login-view", "login-form", "login-error", "ext", "pass",
	"remember", "phone-view", "whoami-ext", "logout", "dial-form", "dest",
	"call-btn", "dial-error", "calls", "keypad", "incoming-call",
	"incoming-from", "accept-btn", "reject-btn", "contacts-wrap",
	"contacts-list", "history-wrap", "history-list", "vm-wrap", "vm-badge",
	"vm-list", "vm-refresh", "vm-status", "ice-wrap", "ice-panel", "log",
	"toasts", "remote-audio", "lang",
}

// testServer bundles the httptest server with the internals the SSE and
// proxy tests need to reach (hubs for subscriptions, phone API wiring).
type testServer struct {
	*httptest.Server
	hubs     *ExtensionHubs
	messages *store.Messages
	faxes    *store.Faxes
	phoneAPI *pbx.Client
}

func newTestServer(t testing.TB) *testServer {
	t.Helper()
	return newTestServerWithPhoneAPI(t, "")
}

// newTestServerWithPhoneAPI builds the full server; the variadic mutators
// let a test swap individual Deps (e.g. a closed DB for healthz 503 tests)
// after the standard wiring but before the handler is built.
func newTestServerWithPhoneAPI(t testing.TB, phoneAPIURL string, mutate ...func(*Deps)) *testServer {
	return newTestServerWithConfig(t, phoneAPIURL, nil, mutate...)
}

// newTestServerWithConfig additionally tweaks the config BEFORE the deps
// are wired (so services like Messaging pick up the tweaked gateway).
func newTestServerWithConfig(
	t testing.TB, phoneAPIURL string, tweakCfg func(*config.Config), mutate ...func(*Deps),
) *testServer {
	t.Helper()

	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	phoneAPI, err := pbx.NewClient(phoneAPIURL)
	if err != nil {
		t.Fatal(err)
	}

	hubs := NewHubs()
	notifier := NewNotifier(hubs, store.NewMessages(db), store.NewFaxes(db))
	cfg := config.Config{
		Addr: ":0", DataDir: t.TempDir(), WebsocketPath: "/sip",
		SessionTTL: time.Hour,
		Gateway:    config.Gateway{Mode: config.GatewayLoopback, WebhookSecret: "test-secret"},
	}
	if tweakCfg != nil {
		tweakCfg(&cfg)
	}
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	deps := Deps{
		Config:    cfg,
		Sessions:  session.NewMemStore(time.Hour),
		Messages:  messages,
		Faxes:     faxes,
		Contacts:  store.NewContacts(db),
		Messaging: messaging.New(messages, blobs, gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient()), notifier.MessagesChanged),
		Fax:       fax.New(faxes, blobs, gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient()), notifier.FaxChanged),
		PhoneAPI:  phoneAPI,
		Hubs:      hubs,
		Shared:    []domain.SharedContact{{Name: "Support", Number: "2000"}},
		DB:        db,
		BlobRoot:  blobs.Root(),
	}
	for _, m := range mutate {
		m(&deps)
	}
	handler := New(deps)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &testServer{Server: server, hubs: hubs, messages: messages, faxes: faxes, phoneAPI: phoneAPI}
}

type client struct {
	t      testing.TB
	base   string
	server *testServer
	token  string
	http   *http.Client
}

func newClient(t *testing.T) *client {
	return clientFor(t, newTestServer(t))
}

func clientFor(t testing.TB, server *testServer) *client {
	t.Helper()
	// server.Client() returns the SAME cached *http.Client on every
	// call — setting Jar on it directly would hijack the cookies of
	// every client built earlier in the test (the first two-client test
	// tripped this as mysterious CSRF 403s). Build a private client
	// sharing only the TLS-configured transport.
	base := server.Client()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: base.Transport, Jar: jar}
	c := &client{t: t, base: server.URL, server: server, http: httpClient}
	c.token = c.csrfToken()
	return c
}

func (c *client) csrfToken() string {
	resp, err := c.http.Get(c.base + "/")
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body := readAll(c.t, resp)
	match := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindSubmatch(body)
	if match == nil {
		c.t.Fatal("no csrf token in page")
	}
	return string(match[1])
}

func readAll(t testing.TB, resp *http.Response) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func (c *client) do(method, path string, body []byte, contentType string) (*http.Response, []byte) {
	c.t.Helper()
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(method, c.base+path, nil)
	} else {
		req, err = http.NewRequest(method, c.base+path, bytes.NewReader(body))
	}
	if err != nil {
		c.t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("X-CSRF-Token", c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp, readAll(c.t, resp)
}

// login creates the server session the way the island does and then adopts
// the CSRF token the login rotated: createSession invalidates the old CSRF
// cookie (fixation defense), so a fresh masked token must come from
// GET /api/csrf before any further POST. The same dance session.js
// performs in the browser.
func (c *client) login(extension, password string) {
	c.t.Helper()
	payload, err := json.Marshal(map[string]string{"extension": extension, "password": password})
	if err != nil {
		c.t.Fatal(err)
	}
	resp, body := c.do(http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		c.t.Fatalf("session create: %d %s", resp.StatusCode, body)
	}
	c.adoptCsrfToken()
}

// loginRaw posts the session create WITHOUT the adoption follow-up:
// flood/limiter tests inspect the raw status themselves and re-arm via
// adoptCsrfToken only on the attempts that actually succeeded.
func (c *client) loginRaw(extension, password string) *http.Response {
	c.t.Helper()
	payload, err := json.Marshal(map[string]string{"extension": extension, "password": password})
	if err != nil {
		c.t.Fatal(err)
	}
	resp, _ := c.do(http.MethodPost, "/api/session", payload, "application/json")
	return resp
}

// adoptCsrfToken mirrors the island's post-login token adoption.
func (c *client) adoptCsrfToken() {
	c.t.Helper()
	resp, body := c.do(http.MethodGet, "/api/csrf", nil, "")
	if resp.StatusCode != http.StatusOK {
		c.t.Fatalf("csrf adoption: %d %s", resp.StatusCode, body)
	}
	var got struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &got); err != nil || got.Token == "" {
		c.t.Fatalf("csrf adoption body %q: %v", body, err)
	}
	c.token = got.Token
}

// multipartBody builds a multipart form with fields and files.
func multipartBody(t *testing.T, fields map[string]string, files map[string]struct {
	Name    string
	Content []byte
}) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	for field, file := range files {
		part, err := writer.CreateFormFile(field, file.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(file.Content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), writer.FormDataContentType()
}

func TestServedPageHoldsTheDomContract(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page status %d", resp.StatusCode)
	}
	page := string(body)
	for _, id := range contractIDs {
		if !strings.Contains(page, fmt.Sprintf("id=%q", id)) {
			t.Errorf("DOM contract broken: id %q missing from served page", id)
		}
	}
	for _, script := range []string{"/config.js", "/assets/vendor/sip.min.js", "/assets/island/app/main.js", "/htmx.min.js"} {
		if !strings.Contains(page, script) {
			t.Errorf("island script %s not referenced", script)
		}
	}
	if !strings.Contains(page, "WebPhone") {
		t.Error("brand missing")
	}
	// The toast host is the WHOLE feedback channel's live region: without
	// role="status" + aria-live, every toast (action feedback, errors,
	// the dead-session 401 notice) is invisible to screen readers. The
	// island and shell only append children to this host — they never
	// set the attributes (pinned client-side by ui.test).
	if !strings.Contains(page, `id="toasts" role="status" aria-live="polite"`) {
		t.Error("toast host lost its live-region attributes (screen readers go blind)")
	}
	// The durable inline error slot + the responseHandling override are
	// one contract: 4xx/5xx panel errors swap the server's .wp-error
	// banner into #wp-tab-error (outside #tab-content so tab swaps can
	// never remove the retarget anchor), while 401 stays swap-free (its
	// plain body must never land in a tab). Dropping either half silently
	// reverts failed actions to toast-only feedback.
	if !strings.Contains(page, `id="wp-tab-error"`) {
		t.Error("error slot #wp-tab-error missing from the shell")
	}
	configMeta := regexp.MustCompile(`name="htmx-config" content='([^']+)'`).FindStringSubmatch(page)
	if configMeta == nil {
		t.Fatal("htmx-config meta missing")
	}
	for _, marker := range []string{`"responseHandling"`, `"401","swap":false`, `"select":".wp-error"`, `"target":"#wp-tab-error"`} {
		if !strings.Contains(configMeta[1], marker) {
			t.Errorf("htmx-config responseHandling contract broken: %q missing", marker)
		}
	}
}

// TestNotFoundRendersTheShell pins error-page parity (re-verified
// 2026-09-20 after the templ-components adoption): unknown paths answer
// 404 with the app shell and the styled error panel, not Go's bare
// "404 page not found" text — a stray deep link keeps the chrome and the
// reload affordance.
func TestNotFoundRendersTheShell(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/nope", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	page := string(body)
	if !strings.Contains(page, `class="wp-panel"`) || !strings.Contains(page, "data-reload") {
		t.Errorf("404 body lacks the styled error panel: %.200s", page)
	}
	if !strings.Contains(page, `id="login-view"`) {
		t.Errorf("404 body lacks the island shell: %.200s", page)
	}
	if !strings.Contains(page, "There is nothing at this address.") {
		t.Errorf("404 body lacks the English not-found message: %.200s", page)
	}
}

// TestServedPageSatisfiesStrictCSP guards the strict-CSP contract: every
// inline script the page serves must be covered by an exact hash in the
// script-src directive, and vice versa, so a stale hash cannot linger.
// The single allowed inline script is templ-components' theme preload
// (see contentSecurityPolicy in server.go); a dependency bump that
// changes its bytes fails here until the hash is refreshed deliberately.
// It also pins the htmx-config meta (keeps htmx from injecting
// CSP-hostile inline indicator styles) and the icon link (without it
// browsers request /favicon.ico and 404).
func TestServedPageSatisfiesStrictCSP(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page status %d", resp.StatusCode)
	}
	page := string(body)

	scriptSrc := regexp.MustCompile(`script-src[^;]*`).FindString(resp.Header.Get("Content-Security-Policy"))
	if scriptSrc == "" {
		t.Fatal("no script-src directive in CSP header")
	}
	allowed := map[string]bool{}
	for _, hash := range regexp.MustCompile(`'sha256-[^']+'`).FindAllString(scriptSrc, -1) {
		allowed[hash] = true
	}
	served := map[string]bool{}
	for _, match := range regexp.MustCompile(`(?s)<script([^>]*)>(.*?)</script>`).FindAllSubmatch(body, -1) {
		if strings.Contains(string(match[1]), "src=") {
			continue
		}
		sum := sha256.Sum256(match[2])
		served[fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(sum[:]))] = true
	}
	for hash := range served {
		if !allowed[hash] {
			t.Errorf("served inline script %s is not allowed by script-src", hash)
		}
	}
	for hash := range allowed {
		if !served[hash] {
			t.Errorf("script-src allows %s but no served inline script matches it (stale hash?)", hash)
		}
	}

	for _, forbidden := range []string{" onclick=", " onload=", " javascript:"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("page contains CSP-hostile inline handler %q", forbidden)
		}
	}
	if !strings.Contains(page, `"includeIndicatorStyles":false`) {
		t.Error("htmx-config meta missing: htmx would inject CSP-hostile inline indicator styles")
	}
	if !strings.Contains(page, `<link rel="icon" href="/favicon.svg">`) {
		t.Error("icon link missing: browsers would request /favicon.ico and 404")
	}
}

func TestStaticAssetsServe(t *testing.T) {
	c := newClient(t)
	for path, wantContains := range map[string]string{
		"/htmx.min.js":                    "htmx",
		"/htmx-ext.js":                    "Idiomorph",
		"/assets/app.css":                 "--accent",
		"/assets/shell.js":                "data-dial",
		"/assets/island/app/main.js":      "loginForm",
		"/assets/island/app/shortcuts.js": "MediaPlayPause",
		"/assets/island/app/session.js":   "initSseLiveIndicator",
		// The dial alphabet is a cross-side contract: the island regex must
		// keep letters exactly like domain.sanitizeDialable (see
		// TestParsePhoneSanitizesLikeTheIsland). Grepping the literal pins
		// a one-sided regex drift at build time.
		"/assets/island/app/calls.js":  "replace(/[^\\d+*#a-zA-Z]/g, \"\")",
		"/assets/island/app/panels.js": "replace(/[^\\d+*#a-zA-Z]/g, \"\")",
		"/assets/island/style.css":     "#wp-sse-live",
		"/assets/vendor/sip.min.js":    "UserAgent",
		"/config.js":                   "window.PBX_CONFIG",
		"/favicon.svg":                 "<svg",
	} {
		resp, body := c.do(http.MethodGet, path, nil, "")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d", path, resp.StatusCode)
			continue
		}
		if !strings.Contains(string(body), wantContains) {
			t.Errorf("%s: body does not contain %q", path, wantContains)
		}
	}
}

func TestConfigJSContract(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode)
	}
	cfg := string(body)
	for _, fragment := range []string{"window.PBX_CONFIG", `"websocketPath":"/sip"`, `"phoneApi":false`} {
		if !strings.Contains(cfg, fragment) {
			t.Errorf("config.js missing %s", fragment)
		}
	}
}
