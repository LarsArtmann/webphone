package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestToastHeadersOnActionPaths pins the toast contract (TO2/TO3): htmx
// tab actions carry the cqrs-htmx ToastDetail wire shape in HX-Trigger —
// "ok"-kind on success, "error"-kind with the localized message on the
// panel-error path — so the island's showMessage listener can surface
// feedback without touching the event log.
func TestToastHeadersOnActionPaths(t *testing.T) {
	t.Run("success: contact saved", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		form := url.Values{"name": {"Ada"}, "number": {"+441632960962"}}
		resp, body := c.do(http.MethodPost, "/contacts/save", []byte(form.Encode()), "application/x-www-form-urlencoded")
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("contacts/save: %d %s", resp.StatusCode, body)
		}
		header := resp.Header.Get("HX-Trigger")
		for _, want := range []string{"showMessage", `"kind":"ok"`, "Contact saved."} {
			if !strings.Contains(header, want) {
				t.Errorf("HX-Trigger %q missing %q", header, want)
			}
		}
	})

	t.Run("error: invalid message recipient", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		form := url.Values{"to": {"not-a-number"}, "body": {"hi"}}
		resp, body := c.do(http.MethodPost, "/messages/send", []byte(form.Encode()), "application/x-www-form-urlencoded")
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("messages/send: %d %s", resp.StatusCode, body)
		}
		header := resp.Header.Get("HX-Trigger")
		for _, want := range []string{"showMessage", `"kind":"error"`} {
			if !strings.Contains(header, want) {
				t.Errorf("HX-Trigger %q missing %q", header, want)
			}
		}
	})
}
