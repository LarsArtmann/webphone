package server

import (
	"bytes"
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

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	return newTestServerWithPhoneAPI(t, "")
}

func newTestServerWithPhoneAPI(t *testing.T, phoneAPIURL string) *testServer {
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
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	handler := New(Deps{
		Config:    cfg,
		Sessions:  session.NewStore(time.Hour),
		Messages:  messages,
		Faxes:     faxes,
		Contacts:  store.NewContacts(db),
		Messaging: messaging.New(messages, blobs, gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient()), notifier.MessagesChanged),
		Fax:       fax.New(faxes, blobs, gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient()), notifier.FaxChanged),
		PhoneAPI:  phoneAPI,
		Hubs:      hubs,
		Shared:    []domain.SharedContact{{Name: "Support", Number: "2000"}},
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &testServer{Server: server, hubs: hubs, messages: messages, faxes: faxes, phoneAPI: phoneAPI}
}

type client struct {
	t      *testing.T
	base   string
	server *testServer
	token  string
	http   *http.Client
}

func newClient(t *testing.T) *client {
	return clientFor(t, newTestServer(t))
}

func clientFor(t *testing.T, server *testServer) *client {
	t.Helper()
	httpClient := server.Client()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	httpClient.Jar = jar
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

func readAll(t *testing.T, resp *http.Response) []byte {
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
}

func TestStaticAssetsServe(t *testing.T) {
	c := newClient(t)
	for path, wantContains := range map[string]string{
		"/htmx.min.js":                    "htmx",
		"/htmx-ext/sse.js":                "sse",
		"/assets/app.css":                 "--accent",
		"/assets/shell.js":                "data-dial",
		"/assets/island/app/main.js":      "loginForm",
		"/assets/island/app/shortcuts.js": "MediaPlayPause",
		"/assets/vendor/sip.min.js":       "UserAgent",
		"/config.js":                      "window.PBX_CONFIG",
		"/favicon.svg":                    "<svg",
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
