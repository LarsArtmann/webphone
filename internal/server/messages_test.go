package server

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
)

func TestMessageSendAndThreadFlow(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "contract test"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}

	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	page := string(body)
	if !strings.Contains(page, "+441632960961") || !strings.Contains(page, "contract test") {
		t.Fatalf("thread list missing the new thread: %s", page[:min(200, len(page))])
	}

	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatal("no thread link in list")
	}
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	threadView := string(body)
	if !strings.Contains(threadView, "contract test") || !strings.Contains(threadView, "wp-status-sent") {
		t.Fatal("thread view missing message or sent badge (loopback gateway)")
	}
}

func TestMMSAttachmentRoundTrip(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

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

	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`href="(/messages/[^"]+)"`).FindSubmatch(body)
	_, body = c.do(http.MethodGet, "/partials"+string(match[1]), nil, "")
	attachment := regexp.MustCompile(`href="(/attachments/[^"]+)"`).FindSubmatch(body)
	if attachment == nil {
		t.Fatal("attachment link missing in thread view")
	}
	resp, body = c.do(http.MethodGet, string(attachment[1]), nil, "")
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, png) {
		t.Fatalf("attachment round trip broken: %d bytes=%d", resp.StatusCode, len(body))
	}
}

// TestNavPartialAndLiveMarkRead covers the live-polish contract: the nav
// partial re-renders labels (the island's language switch re-fetches it,
// so nav labels follow without a full reload), and the explicit read
// endpoint clears the badge for an open conversation after an SSE swap —
// a live push never re-GETs the partial, so only the client can mark read.
func TestNavPartialAndLiveMarkRead(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	deliverInbound(t, server, "+441632960961", "live polish")
	threadRow := regexp.MustCompile(`hx-get="/partials/messages/([^"]+)"`)
	_, listBody := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := threadRow.FindSubmatch(listBody)
	if match == nil {
		t.Fatal("no thread row rendered")
	}
	threadID := string(match[1])

	if badge := unreadBadge(t, c); badge != "1" {
		t.Fatalf("badge before live mark-read: got %q, want 1", badge)
	}

	// Signed-in nav partial: active tab honored, badge fresh, language
	// from the wp-lang cookie (same negotiation the tabs use).
	req, err := http.NewRequest(http.MethodGet, server.URL+"/partials/nav?active=history", nil)
	if err != nil {
		t.Fatal(err)
	}
	var jar []string
	base, _ := url.Parse(server.URL)
	for _, cookie := range c.http.Jar.Cookies(base) {
		jar = append(jar, cookie.Name+"="+cookie.Value)
	}
	req.Header.Set("Cookie", strings.Join(append(jar, "wp-lang=de"), "; "))
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	navBody := string(readAll(t, resp))
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("nav partial: %d", resp.StatusCode)
	}
	activeLink := regexp.MustCompile(`data-tab="history"[^>]*`).FindString(navBody)
	if !strings.Contains(navBody, `data-tab="history"`) || !strings.Contains(activeLink+navBody, "wp-active") {
		t.Errorf("nav partial lost the active tab: %.200s", navBody)
	}
	if !strings.Contains(navBody, "Nachrichten") {
		t.Errorf("nav partial ignored the wp-lang cookie: %.200s", navBody)
	}
	if !strings.Contains(navBody, `wp-nav-badge">1<`) {
		t.Errorf("nav partial missing the unread badge: %.200s", navBody)
	}

	// The live-swap read endpoint: clears the badge, idempotent.
	if resp, body := c.do(http.MethodPost, "/messages/"+threadID+"/read", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark read: %d %s", resp.StatusCode, body)
	}
	if badge := unreadBadge(t, c); badge != "" {
		t.Fatalf("badge after live mark-read: got %q, want none", badge)
	}
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/read", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark read must be idempotent, got %d", resp.StatusCode)
	}

	// Anonymous visitors get labels (no session, no badge): the nav is
	// visible pre-login and the language switch works there too.
	anon := newClient(t)
	resp, body := anon.do(http.MethodGet, "/partials/nav", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("anonymous nav partial: %d", resp.StatusCode)
	}
	navBody = string(body)
	if !strings.Contains(navBody, `data-tab="messages"`) {
		t.Errorf("anonymous nav partial missing links: %.200s", navBody)
	}
	if strings.Contains(navBody, "wp-nav-badge") {
		t.Errorf("anonymous nav partial must not carry badges: %.200s", navBody)
	}
}

// TestSendClassifiesGatewayOutageAs502 pins the error taxonomy: an
// unreachable send gateway is a server-side failure (502, message saved
// as failed) while a validation mistake stays 422. Before this pin an
// outage masqueraded as a client error with the gateway's internal
// error text leaked into the page.
func TestSendClassifiesGatewayOutageAs502(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Gateway = config.Gateway{
			Mode:          config.GatewayWebhook,
			WebhookURL:    "http://127.0.0.1:1", // nothing listens there
			WebhookSecret: "test-secret",
		}
	})
	c := clientFor(t, server)
	c.login("1001", "pw")

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "hi"}, nil)
	resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("gateway outage: %d %s (want 502)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "unreachable") {
		t.Fatalf("502 body must tell the user the message was saved: %.200s", body)
	}

	form, contentType = multipartBody(t, map[string]string{"to": "+441632960961", "body": ""}, nil)
	resp, _ = c.do(http.MethodPost, "/messages/send", form, contentType)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("validation mistake: %d (want 422)", resp.StatusCode)
	}
}
