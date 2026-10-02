package server

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/store"
)

// threadIDFromList extracts the first thread id from the messages panel
// (the row anchor carries /messages/{id}).
func threadIDFromList(t *testing.T, c *client) string {
	t.Helper()
	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	match := regexp.MustCompile(`hx-get="/partials/messages/([^"]+)"`).FindSubmatch(body)
	if match == nil {
		t.Fatalf("no thread row rendered: %.300s", body)
	}
	return string(match[1])
}

// TestThreadFlagActions covers the M22 toggles end to end: the auth
// gate, the desired-state form value, owner-scoped misses, the unknown
// flag, and the #wp-thread-head-targeted re-render that flips the open
// conversation's buttons to the confirmed state.
func TestThreadFlagActions(t *testing.T) {
	server := newTestServer(t)

	anon := &client{t: t, base: server.URL, server: server, http: server.Client()}
	// Anonymous state-changing verbs sit behind CSRF too — the accepted
	// pattern is 401 OR 403, never success.
	resp, _ := anon.do(http.MethodPost, "/messages/x/pin?on=1", nil, "")
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("anon toggle: got %d, want 401/403", resp.StatusCode)
	}

	c := signIn(t, server)
	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "flag me"}, nil)
	if resp, body := c.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}
	threadID := threadIDFromList(t, c)

	// Unknown flag and bad id are plain 404s.
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/explode?on=1", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown flag: got %d, want 404", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodPost, "/messages/bogus/pin?on=1", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("bad id: got %d, want 404", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodPost, "/messages/Thread:does-not-exist/pin?on=1", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing thread: got %d, want 404", resp.StatusCode)
	}

	// Pin: 204 for the row buttons; the list reflects the pin.
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/pin?on=1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("pin: got %d, want 204", resp.StatusCode)
	}
	_, body := c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(body), "wp-pinned") {
		t.Fatalf("pinned row class missing: %.300s", body)
	}

	// A head-targeted post answers the FRESH head: the mute button has
	// flipped to Unmute posting on=0. The request needs the raw cookie
	// (c.do sets the CSRF token but has no extra-header hook).
	headReq, _ := http.NewRequest(http.MethodPost, server.URL+"/messages/"+threadID+"/mute?on=1", nil)
	for _, cookie := range c.http.Jar.Cookies(mustURL(t, server.URL)) {
		headReq.AddCookie(cookie)
	}
	headReq.Header.Set("X-CSRF-Token", c.token)
	headReq.Header.Set("HX-Target", "wp-thread-head")
	headResp, err := server.Client().Do(headReq)
	if err != nil {
		t.Fatal(err)
	}
	head := string(readAll(t, headResp))
	if headResp.StatusCode != http.StatusOK {
		t.Fatalf("head toggle: got %d, want 200", headResp.StatusCode)
	}
	if !strings.Contains(head, `id="wp-thread-head"`) {
		t.Fatalf("head response must be the head element: %s", head)
	}
	if !strings.Contains(head, `aria-label="Unmute"`) || !strings.Contains(head, "/mute?on=0") {
		t.Fatalf("head buttons must show the CONFIRMED state: %s", head)
	}
	if strings.Contains(head, "thread-transcript") {
		t.Fatalf("head response must not carry the transcript: %s", head)
	}
}

// TestMuteDropsNavBadge pins D8's countUnread leg: a muted thread's
// unread stays in the store but the nav badge stops counting it.
func TestMuteDropsNavBadge(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	deliverInbound(t, server, "+441632960961", "muted pings")
	if badge := unreadBadge(t, c); badge != "1" {
		t.Fatalf("badge before mute: got %q, want 1", badge)
	}

	threadID := threadIDFromList(t, c)
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/mute?on=1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mute: got %d, want 204", resp.StatusCode)
	}
	if badge := unreadBadge(t, c); badge != "" {
		t.Fatalf("badge after mute: got %q, want none", badge)
	}

	// Unmute brings the attention surface back — the unread truth never
	// left the store.
	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/mute?on=0", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unmute: got %d, want 204", resp.StatusCode)
	}
	if badge := unreadBadge(t, c); badge != "1" {
		t.Fatalf("badge after unmute: got %q, want 1", badge)
	}
}

// TestArchiveFlowViaActions walks the D6/D7 filing-cabinet loop through
// the HTTP surface: archive hides the row from the list and search,
// the archived view lists it, and an inbound message brings it back.
func TestArchiveFlowViaActions(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	deliverInbound(t, server, "+441632960961", "to be filed")
	threadID := threadIDFromList(t, c)

	if resp, _ := c.do(http.MethodPost, "/messages/"+threadID+"/archive?on=1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("archive: got %d, want 204", resp.StatusCode)
	}

	_, listBody := c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(listBody), "Archived (1)") {
		t.Fatalf("active view must offer the archived toggle with count: %.400s", listBody)
	}
	if strings.Contains(string(listBody), "wp-thread-rowwrap") {
		t.Fatalf("archived thread must leave the active list: %s", listBody)
	}
	if _, searchBody := c.do(http.MethodGet, "/partials/messages?q=filed", nil, ""); strings.Contains(string(searchBody), "wp-thread-rowwrap") {
		t.Fatalf("search must not resurrect the archived thread: %s", searchBody)
	}

	_, archivedBody := c.do(http.MethodGet, "/partials/messages?archived=1", nil, "")
	archived := string(archivedBody)
	if !strings.Contains(archived, "wp-thread-rowwrap") || !strings.Contains(archived, `aria-label="Unarchive"`) {
		t.Fatalf("archived view must list the thread with unarchive: %.400s", archived)
	}
	if strings.Contains(archived, `class="wp-thread-search"`) {
		t.Fatalf("archived view must hide the search box: %s", archived)
	}

	// D7: inbound auto-unarchives; the count follows.
	deliverInbound(t, server, "+441632960961", "hello again")
	_, listAfter := c.do(http.MethodGet, "/partials/messages", nil, "")
	if !strings.Contains(string(listAfter), "wp-thread-rowwrap") {
		t.Fatalf("inbound must return the thread to the active list: %.400s", listAfter)
	}
	if strings.Contains(string(listAfter), "Archived (1)") {
		t.Fatalf("archived count must drop after the inbound: %s", listAfter)
	}
}

// TestSnippetActions covers the Settings CRUD: save happy path (quick
// flag), the 422s for empty and oversized bodies and the cap, delete
// with its 404, and the settings partial re-render both answers carry.
func TestSnippetActions(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	save := func(body, quick string) (*http.Response, string) {
		fields := map[string]string{"body": body}
		if quick != "" {
			fields["quick"] = quick
		}
		form, contentType := multipartBody(t, fields, nil)
		resp, respBody := c.do(http.MethodPost, "/snippets/save", form, contentType)
		return resp, string(respBody)
	}

	resp, body := save("Thanks, on it!", "1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("save: %d %s", resp.StatusCode, body)
	}
	if !strings.Contains(body, `class="wp-snippet-row"`) || !strings.Contains(body, "Thanks, on it!") {
		t.Fatalf("save must re-render the settings partial with the row: %.400s", body)
	}
	if !strings.Contains(body, "wp-snippet-quick") {
		t.Fatalf("quick marker missing: %s", body)
	}

	if resp, body := save("   ", ""); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty body: got %d %s, want 422", resp.StatusCode, body)
	}
	if resp, body := save(strings.Repeat("x", 1601), ""); resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "too long") {
		t.Fatalf("oversized body: got %d %s, want 422", resp.StatusCode, body)
	}

	// The cap: fill to the store limit, then one more answers 422.
	for i := 1; i < store.SnippetsMaxPerExtension; i++ {
		if resp, body := save("filler", ""); resp.StatusCode != http.StatusOK {
			t.Fatalf("fill save %d: %d %s", i, resp.StatusCode, body)
		}
	}
	if resp, body := save("one too many", ""); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("cap: got %d %s, want 422", resp.StatusCode, body)
	}

	// Delete: extract an id from the partial, delete it, double-delete 404.
	_, settings := c.do(http.MethodGet, "/partials/settings", nil, "")
	match := regexp.MustCompile(`/snippets/delete\?id=([^"]+)"`).FindSubmatch([]byte(settings))
	if match == nil {
		t.Fatalf("no delete button in settings: %.400s", settings)
	}
	id := string(match[1])
	if resp, _ := c.do(http.MethodPost, "/snippets/delete?id="+url.QueryEscape(id), nil, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: got %d, want 200 (partial)", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodPost, "/snippets/delete?id="+url.QueryEscape(id), nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("double delete: got %d, want 404", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodPost, "/snippets/delete?id=bogus", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("bad id delete: got %d, want 404", resp.StatusCode)
	}
}
