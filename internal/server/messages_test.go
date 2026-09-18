package server

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func TestMessageSendAndThreadFlow(t *testing.T) {
	c := newClient(t)
	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	c.do(http.MethodPost, "/api/session", payload, "application/json")

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
