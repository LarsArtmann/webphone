package server

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// shellMarker appears only in the full application shell (the HTMX swap
// target), never in a fragment: it is the cheapest way to tell "the user
// got a page" from "the user got a notice".
const shellMarker = `id="tab-content"`

// doWithHeaders is client.do plus request headers: the 404 shape depends on
// HX-Request, which the shared helper cannot express.
func doWithHeaders(c *client, method, path, body, contentType string, headers map[string]string) (*http.Response, []byte) {
	c.t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("X-CSRF-Token", c.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp, readAll(c.t, resp)
}

var htmxHeaders = map[string]string{"HX-Request": "true"}

// A miss on an HTMX endpoint gets a notice, not a shell: htmx discards
// non-2xx bodies (its responseError toast carries the status), so shipping
// the whole application page would be bytes nobody renders.
func TestHTMXMissAnswersANoticeNotTheShell(t *testing.T) {
	c := signIn(t, newTestServer(t))

	resp, body := doWithHeaders(c, http.MethodGet, "/partials/messages/not-a-thread-id", "", "", htmxHeaders)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("content-type %q, want text/html", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	notice := string(body)
	if strings.Contains(notice, shellMarker) {
		t.Errorf("an HTMX miss must not ship the shell, got %d bytes", len(notice))
	}
	if !strings.Contains(notice, "404 — page not found") {
		t.Errorf("the notice must name the 404, got %.200s", notice)
	}
}

// The same URL opened as a document (a partial path pasted out of devtools,
// no HX-Request) is a real navigation and deserves the styled page.
func TestDocumentMissOnAPartialURLAnswersTheShell(t *testing.T) {
	c := signIn(t, newTestServer(t))

	resp, body := doWithHeaders(c, http.MethodGet, "/partials/messages/not-a-thread-id", "", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}
	page := string(body)
	if !strings.Contains(page, shellMarker) {
		t.Fatalf("a document miss must render the shell, got %d bytes", len(page))
	}
	if !strings.Contains(page, "/partials/messages/not-a-thread-id") {
		t.Errorf("the requested path must be echoed, got %.300s", page)
	}
}

// The action endpoints (fan-out buttons, snippet delete, draft controls) are
// HTML surfaces too: a stale one must not answer stdlib text in either shape.
func TestActionMissAnswersByRequestShape(t *testing.T) {
	c := signIn(t, newTestServer(t))

	cases := []struct{ name, method, path string }{
		{"thread flag", http.MethodPost, "/messages/bogus/pin?on=1"},
		{"snippet delete", http.MethodPost, "/snippets/delete?id=bogus"},
		{"contact delete", http.MethodPost, "/contacts/delete?id=short"},
		{"drafts poll", http.MethodGet, "/partials/messages/drafts"},
	}
	for _, tc := range cases {
		t.Run(tc.name+"/htmx", func(t *testing.T) {
			resp, body := doWithHeaders(c, tc.method, tc.path, "", "", htmxHeaders)
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("status %d, want 404", resp.StatusCode)
			}
			if strings.Contains(string(body), shellMarker) {
				t.Errorf("an HTMX miss must not ship the shell (%d bytes)", len(body))
			}
		})
		t.Run(tc.name+"/document", func(t *testing.T) {
			resp, body := doWithHeaders(c, tc.method, tc.path, "", "", nil)
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("status %d, want 404", resp.StatusCode)
			}
			if !strings.Contains(string(body), shellMarker) {
				t.Errorf("a document miss must render the shell, got %d bytes", len(body))
			}
		})
	}
}

// A stale fax link is a browser download, so the miss is a navigation.
func TestStaleFaxDocumentLinkAnswersTheStyledPage(t *testing.T) {
	c := signIn(t, newTestServer(t))

	resp, body := c.do(http.MethodGet, "/fax/not-a-fax-id/document", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}
	if !strings.Contains(string(body), "404 — page not found") {
		t.Errorf("a stale fax link must render the app page, got %.200s", body)
	}
}

// The ruled exclusion: /attachments is a fetched byte stream (<img src>,
// <audio src>). An HTML body there is meaningless, so it keeps the terse
// stdlib answer even when a browser navigates to it directly.
func TestAttachmentMissKeepsTheTerseAnswer(t *testing.T) {
	c := signIn(t, newTestServer(t))

	resp, body := c.do(http.MethodGet, "/attachments/not-an-attachment-id", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("content-type %q, want text/plain (a byte stream must not answer HTML)", got)
	}
	if strings.Contains(string(body), shellMarker) {
		t.Error("the attachment miss must stay terse")
	}
}
