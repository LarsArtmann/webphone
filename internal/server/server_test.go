package server

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// domContractIDs loads the island DOM-contract id list from
// docs/dom-contract.md — the file is the single source of truth
// (AGENTS and the stack runbook link it), so the asserted list can
// never drift from the documented one. Marker comments fence the
// block; the emptiness guard stops a truncated file from passing
// vacuously.
func domContractIDs(t testing.TB) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "dom-contract.md"))
	if err != nil {
		t.Fatalf("read docs/dom-contract.md: %v", err)
	}
	const (
		beginMark = "<!-- dom-contract:begin -->"
		endMark   = "<!-- dom-contract:end -->"
	)
	body := string(raw)
	i := strings.Index(body, beginMark)
	j := strings.Index(body, endMark)
	if i < 0 || j < 0 || j < i {
		t.Fatal("docs/dom-contract.md lost its dom-contract marker comments")
	}
	var ids []string
	for _, line := range strings.Split(body[i+len(beginMark):j], "\n") {
		if id := strings.TrimSpace(line); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) < 20 {
		t.Fatalf("docs/dom-contract.md carries only %d ids — truncated?", len(ids))
	}
	return ids
}

// testServer bundles the httptest server with the internals the SSE and
// proxy tests need to reach (hubs for subscriptions, phone API wiring).
type testServer struct {
	*httptest.Server
	handler  http.Handler
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
	notifier := NewNotifier(hubs, store.NewMessages(db), store.NewFaxes(db), nil)
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
	// SQLite session store (not the mem store): the /metrics aggregates
	// count the sessions TABLE, which only the SQLite store's schema
	// creates — TestMetricsServesAggregatesOnly pins that surface.
	sessions, err := session.NewSQLiteStore(db, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	deps := Deps{
		Config:    cfg,
		Sessions:  sessions,
		Messages:  messages,
		Faxes:     faxes,
		Contacts:  store.NewContacts(db),
		Messaging: messaging.New(messages, blobs, gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient()), notifier.MessagesChanged, cfg.Identities),
		Fax:       fax.New(faxes, blobs, gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient()), notifier.FaxChanged, cfg.Identities),
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
	return &testServer{Server: server, handler: handler, hubs: hubs, messages: messages, faxes: faxes, phoneAPI: phoneAPI}
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
	for _, id := range domContractIDs(t) {
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
func TestFaviconAnswersBothNames(t *testing.T) {
	c := newClient(t)
	for _, path := range []string{"/favicon.svg", "/favicon.ico"} {
		resp, body := c.do(http.MethodGet, path, nil, "")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d", path, resp.StatusCode)
			continue
		}
		if ct := resp.Header.Get("Content-Type"); ct != "image/svg+xml" {
			t.Errorf("%s: content-type %q", path, ct)
		}
		if len(body) == 0 {
			t.Errorf("%s: empty body", path)
		}
	}
}

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

// TestServedPageSatisfiesStrictCSP guards the strict-CSP contract: the
// page serves ZERO inline scripts and script-src allows none — no hashes,
// no unsafe-inline — so a dependency bump can never silently change
// served script bytes (the theme preload ships as the same-origin
// /assets/theme-preload.js instead; see contentSecurityPolicy in
// server.go). It also pins the htmx-config meta (keeps htmx from
// injecting CSP-hostile inline indicator styles) and the icon link
// (without it browsers request /favicon.ico and 404).
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
	if hashes := regexp.MustCompile(`'sha256-[^']+'|'unsafe-inline'`).FindAllString(scriptSrc, -1); len(hashes) > 0 {
		t.Errorf("script-src must allow no hashes and no unsafe-inline (zero inline scripts), found %v", hashes)
	}
	for _, match := range regexp.MustCompile(`(?s)<script([^>]*)>`).FindAllSubmatch(body, -1) {
		if !strings.Contains(string(match[1]), "src=") {
			t.Errorf("page serves an inline <script> without src: %.80s", match[0])
		}
	}
	if !strings.Contains(page, `<script src="/assets/theme-preload.js"></script>`) {
		t.Error("theme preload script tag missing: forced themes would flash the OS theme before shell.js runs")
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

func TestMetricsServesAggregatesOnly(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "metric probe"}, nil)
	if resp, _ := c.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d", resp.StatusCode)
	}

	resp, body := c.do(http.MethodGet, "/metrics", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("metrics content-type: %q", ct)
	}
	page := string(body)
	for _, want := range []string{
		"webphone_build_info",
		"webphone_uptime_seconds",
		"webphone_threads_total 1",
		"webphone_messages_total 1",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("metrics missing %s:\n%s", want, page)
		}
	}
	// AGGREGATES ONLY: a scraped metrics body must never carry
	// per-extension or per-number data.
	for _, leak := range []string{"1001", "+441632960961", "metric probe"} {
		if strings.Contains(page, leak) {
			t.Errorf("metrics leaks %q — aggregates only:\n%s", leak, page)
		}
	}
	// The CRM integration is off in this deployment: its metric family
	// must be absent entirely, not published as zeros.
	if strings.Contains(page, "webphone_crm_lookups_total") {
		t.Errorf("metrics publishes CRM lookups with the integration off:\n%s", page)
	}
}
