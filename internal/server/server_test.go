package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

func newTestServer(t *testing.T) *httptest.Server {
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
	phoneAPI, err := pbx.NewClient("")
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
	return server
}

type client struct {
	t      *testing.T
	base   string
	server *httptest.Server
	token  string
	http   *http.Client
}

func newClient(t *testing.T) *client {
	server := newTestServer(t)
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
		"/htmx.min.js":               "htmx",
		"/htmx-ext/sse.js":           "sse",
		"/assets/app.css":            "--accent",
		"/assets/shell.js":           "data-dial",
		"/assets/island/app/main.js": "loginForm",
		"/assets/vendor/sip.min.js":  "UserAgent",
		"/config.js":                 "window.PBX_CONFIG",
		"/favicon.svg":               "<svg",
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

func TestMessageSendAndThreadFlow(t *testing.T) {
	c := newClient(t)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "contract test"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}

	resp, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	page := string(body)
	if !strings.Contains(page, "+441632960961") || !strings.Contains(page, "contract test") {
		t.Fatalf("thread list missing the new thread: %s", page[:min(200, len(page))])
	}

	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	resp, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	threadView := string(body)
	if !strings.Contains(threadView, "contract test") || !strings.Contains(threadView, "wp-status-sent") {
		t.Fatal("thread view missing message or sent badge (loopback gateway)")
	}
}

func TestMMSAttachmentRoundTrip(t *testing.T) {
	c := newClient(t)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")

	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 16)...)
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961", "body": "with media"},
		map[string]struct {
			Name    string
			Content []byte
		}{"attachment": {Name: "pic.png", Content: png}})
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mms send: %d %s", resp.StatusCode, body)
	}

	resp, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	resp, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	attachment := regexp.MustCompile(`href="(/attachments/[^"]+)"`).FindSubmatch(body)
	if attachment == nil {
		t.Fatal("attachment link missing in thread view")
	}
	resp, body = c.do(http.MethodGet, string(attachment[1]), nil, "")
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, png) {
		t.Fatalf("attachment round trip broken: %d bytes=%d", resp.StatusCode, len(body))
	}
}

func TestFaxSendAndDocument(t *testing.T) {
	c := newClient(t)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")

	pdf := []byte("%PDF-1.4\n%test\ntrailer<<>>\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fax send: %d %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "transmitted") {
		t.Fatal("loopback fax must resolve to transmitted")
	}

	match := regexp.MustCompile(`href="(/fax/[^/]+/document)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("document link missing")
	}
	resp, body = c.do(http.MethodGet, string(match[1]), nil, "")
	if resp.StatusCode != http.StatusOK || !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatalf("fax document: %d", resp.StatusCode)
	}

	// Non-PDF is rejected.
	form, contentType = multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "x.txt", Content: []byte("not a pdf")}})
	resp, _ = c.do(http.MethodPost, "/fax/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("non-pdf fax: %d (want 422)", resp.StatusCode)
	}
}

func TestWebhooksSecretAndInbound(t *testing.T) {
	c := newClient(t)
	server := newTestServer(t)
	c.base = server.URL

	inbound := map[string]any{"owner": "1001", "from": "+4915112345678", "body": "hooked"}
	payload, _ := json.Marshal(inbound)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("webhook without secret: %d (want 401)", badResp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("webhook with wrong secret: %d", badResp.StatusCode)
	}

	// Right secret stores the message; reading it back needs a session.
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/hooks/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	if okResp, _ := c.http.Do(req); okResp.StatusCode != http.StatusAccepted {
		t.Fatalf("webhook with secret: status %d", okResp.StatusCode)
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

func TestInboundFaxWebhookStoresDocument(t *testing.T) {
	server := newTestServer(t)
	pdf := []byte("%PDF-1.4 hook\n%%EOF\n")
	payload, _ := json.Marshal(map[string]any{
		"owner": "1001", "from": "+491700000000", "pages": 2,
		"pdf_base64": base64.StdEncoding.EncodeToString(pdf),
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("fax webhook: %d", resp.StatusCode)
	}
}
