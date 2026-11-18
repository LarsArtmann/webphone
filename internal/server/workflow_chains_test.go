package server

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// chainThreadLink pulls the first thread id out of a messages listing.
func chainThreadLink(t *testing.T, listing string) string {
	t.Helper()
	match := regexp.MustCompile(`href="/messages/([^"]+)"`).FindStringSubmatch(listing)
	if match == nil {
		t.Fatal("no thread link in listing")
	}
	return match[1]
}

// chainExtractAttr pulls the first capture group of pattern out of body,
// empty string when absent.
func chainExtractAttr(body, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(body)
	if match == nil {
		return ""
	}
	return match[1]
}

// The inbound-message workflow as ONE chain over the real mux: the
// webhook lands, the thread appears with its unread badge, BOTH SSE
// events fire with swap-safe payloads, a second inbound consolidates
// into one row, and an inbound to an ARCHIVED thread auto-unarchives it
// (the AppendToThread invariant, proven over HTTP — the store-level pin
// in store/threads_flags_test.go cannot see the route wiring).
func TestInboundWorkflowChain(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)
	events := subscribeEvents(t, server)

	postInbound := func(from, body string) {
		t.Helper()
		payload, err := json.Marshal(map[string]any{"owner": "1001", "from": from, "body": body})
		if err != nil {
			t.Fatal(err)
		}
		if status, _ := postHook(t, server, "/hooks/message", string(payload)); status != http.StatusAccepted {
			t.Fatalf("inbound hook from %s: %d (want 202)", from, status)
		}
	}

	postInbound("+441632960201", "chain inbound one")

	listEv := expectEvent(t, events, "threads")
	if !strings.Contains(listEv.Data, "wp-thread-row") {
		t.Fatalf("threads event payload is not a swap-safe list fragment: %.120q", listEv.Data)
	}
	threadEv := expectEvent(t, events, "thread")
	if !strings.Contains(threadEv.Data, "wp-bubble") || strings.Contains(threadEv.Data, "wp-compose") {
		t.Fatalf("thread event payload must be bubbles only: %.120q", threadEv.Data)
	}

	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(body), "+441632960201") || !strings.Contains(string(body), "wp-unread") {
		t.Fatalf("list missing the new unread thread: %.200s", body)
	}

	// A second inbound to the same remote consolidates: one row, unread 2.
	postInbound("+441632960201", "chain inbound two")
	expectEvent(t, events, "threads")
	owner := domain.MustParseExtension("1001")
	threads, err := server.messages.ListThreads(context.Background(), owner)
	if err != nil || len(threads) != 1 {
		t.Fatalf("after two inbounds: %d threads (err %v)", len(threads), err)
	}
	if threads[0].Unread != 2 {
		t.Fatalf("unread = %d, want 2", threads[0].Unread)
	}

	// Archive, then prove the LIVE channel re-opens the conversation: an
	// inbound to the archived thread auto-unarchives it (D7).
	if resp, respBody := c.do(http.MethodPost, "/messages/"+threads[0].ThreadID+"/archive?on=1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("archive: %d %s", resp.StatusCode, respBody)
	}
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	if strings.Contains(string(body), "+441632960201") {
		t.Fatal("archived thread still in the default list")
	}

	postInbound("+441632960201", "chain inbound unarchives")
	expectEvent(t, events, "threads")
	_, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(body), "+441632960201") {
		t.Fatal("inbound did not auto-unarchive the archived thread")
	}
}

// The owner-to-owner fences over HTTP: another extension's session must
// not read a foreign attachment, a foreign fax document, or a foreign
// thread deep link (the gone-notice, never the transcript). The *Any
// store seams are admin-gated at the handler layer; this pins the
// default-deny direction the admin fallback carves its exception into.
func TestForeignOwnerSurfacesStayDenied(t *testing.T) {
	server := newTestServer(t)

	seed := signIn(t, server)
	other := clientFor(t, server)
	other.login("1002", "pw")

	// Owner 1001 seeds an inbound MMS attachment through the hook lane.
	payload, err := json.Marshal(map[string]any{
		"owner": "1001", "from": "+441632960301", "body": "with attachment",
		"attachments": []map[string]string{
			{"name": "dot.png", "mime_type": "image/png", "data_base64": "iVBORw0KGgo="},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status, _ := postHook(t, server, "/hooks/message", string(payload)); status != http.StatusAccepted {
		t.Fatalf("mms hook: %d", status)
	}
	_, list := seed.do(http.MethodGet, "/partials/messages", nil, "")
	threadID := chainThreadLink(t, string(list))
	_, detail := seed.do(http.MethodGet, "/partials/messages/"+threadID, nil, "")
	attachment := chainExtractAttr(string(detail), `/attachments/([^"]+)"`)
	if attachment == "" {
		t.Fatalf("no attachment link in the transcript: %.200s", detail)
	}

	if resp, _ := other.do(http.MethodGet, "/attachments/"+attachment, nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign attachment: %d (want 404)", resp.StatusCode)
	}

	// The same fence on the fax document lane.
	form, contentType := multipartBody(t, map[string]string{"to": "+441632960302"}, map[string]struct {
		Name    string
		Content []byte
	}{"document": {Name: "doc.pdf", Content: []byte("%PDF-1.4 stub")}})
	if resp, body := seed.do(http.MethodPost, "/fax/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("fax send: %d %s", resp.StatusCode, body)
	}
	_, faxPanel := seed.do(http.MethodGet, "/partials/fax", nil, "")
	faxDoc := chainExtractAttr(string(faxPanel), `href="/fax/([^"]+)/document"`)
	if faxDoc == "" {
		t.Fatal("no fax document link in the panel")
	}
	if resp, _ := other.do(http.MethodGet, "/fax/"+faxDoc+"/document", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign fax document: %d (want 404)", resp.StatusCode)
	}

	// A foreign thread deep link degrades to the list + gone-notice —
	// never a 404 shell and never the other owner's transcript.
	resp, body := other.do(http.MethodGet, "/messages/"+threadID, nil, "")
	if resp.StatusCode != http.StatusOK || strings.Contains(string(body), "with attachment") || !strings.Contains(string(body), "no longer exists") {
		t.Fatalf("foreign thread deep link: %d — must render the gone-notice, not the transcript", resp.StatusCode)
	}
}
