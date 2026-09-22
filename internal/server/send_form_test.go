package server

import (
	"net/http"
	"strings"
	"testing"
)

// requireMultipartTo is the send-form prologue for BOTH outbound tabs
// (2026-09-22 dedup train): a malformed "to" must answer 422 through
// the per-tab i18n reason, so each tab's error copy names its own
// failure without leaking the shared helper's generality.
func TestRequireMultipartToAnswersPerTab422(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	t.Run("messages tab", func(t *testing.T) {
		form, contentType := multipartBody(t, map[string]string{"to": "not-a-number", "body": "hi"}, nil)
		resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("invalid to: %d (want 422)", resp.StatusCode)
		}
		if !strings.Contains(string(body), "Enter a valid number to send to.") {
			t.Fatalf("messages 422 lost its per-tab reason: %.200s", body)
		}
		if strings.Contains(string(body), "Enter a valid fax number.") {
			t.Fatalf("messages 422 must not carry the fax tab reason: %.200s", body)
		}
	})

	t.Run("fax tab", func(t *testing.T) {
		form, contentType := multipartBody(t, map[string]string{"to": "not-a-number"}, nil)
		resp, body := c.do(http.MethodPost, "/fax/send", form, contentType)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("invalid to: %d (want 422)", resp.StatusCode)
		}
		if !strings.Contains(string(body), "Enter a valid fax number.") {
			t.Fatalf("fax 422 lost its per-tab reason: %.200s", body)
		}
		if strings.Contains(string(body), "Enter a valid number to send to.") {
			t.Fatalf("fax 422 must not carry the messages tab reason: %.200s", body)
		}
	})
}
