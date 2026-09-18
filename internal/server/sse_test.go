package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/webphone/internal/domain"
)

// subscribeEvents subscribes the test to extension 1001's hub events;
// the subscription is torn down automatically with the test.
func subscribeEvents(t *testing.T, server *testServer) <-chan sse.Event {
	t.Helper()
	hub := server.hubs.get(domain.MustParseExtension("1001"))
	events := hub.Hub().Subscribe()
	t.Cleanup(func() { hub.Hub().Unsubscribe(events) })
	return events
}

func expectEvent(t *testing.T, events <-chan sse.Event, name string) sse.Event {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		select {
		case ev := <-events:
			if ev.Event == name {
				return ev
			}
		case <-deadline.C:
			t.Fatalf("timed out waiting for SSE event %q", name)
		}
	}
}

func assertNoEvent(t *testing.T, events <-chan sse.Event, name string) {
	t.Helper()
	select {
	case ev := <-events:
		if ev.Event == name {
			t.Fatalf("unexpected %q event (data %.80q)", name, ev.Data)
		}
	case <-time.After(150 * time.Millisecond):
	}
}

// The SSE payloads must be swap-safe fragments: exactly the inner region
// the sse-swap element replaces — never a wrapper section and never the
// composer (a live push must not wipe a draft or nest panels).
func TestSSEPushesSwapSafeFragments(t *testing.T) {
	server := newTestServer(t)
	events := subscribeEvents(t, server)
	c := signIn(t, server)

	form, contentType := multipartBody(t, map[string]string{"to": "+441632960961", "body": "live push"}, nil)
	if resp, body := c.do(http.MethodPost, "/messages/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("send: %d %s", resp.StatusCode, body)
	}

	threads := expectEvent(t, events, "threads")
	if !strings.Contains(threads.Data, "wp-thread-row") || !strings.Contains(threads.Data, "+441632960961") {
		t.Fatalf("threads payload missing the thread row: %.200s", threads.Data)
	}
	for _, forbidden := range []string{"<section", "wp-compose", "wp-panel-head"} {
		if strings.Contains(threads.Data, forbidden) {
			t.Errorf("threads payload must not contain %q: %.200s", forbidden, threads.Data)
		}
	}

	thread := expectEvent(t, events, "thread")
	if !strings.Contains(thread.Data, "wp-bubble") || !strings.Contains(thread.Data, "live push") {
		t.Fatalf("thread payload missing the bubble: %.200s", thread.Data)
	}
	for _, forbidden := range []string{"<section", "wp-compose", "wp-back", "thread-transcript"} {
		if strings.Contains(thread.Data, forbidden) {
			t.Errorf("thread payload must not contain %q: %.200s", forbidden, thread.Data)
		}
	}

	pdf := []byte("%PDF-1.4 live\n%%EOF\n")
	form, contentType = multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	if resp, body := c.do(http.MethodPost, "/fax/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("fax send: %d %s", resp.StatusCode, body)
	}
	fax := expectEvent(t, events, "fax")
	if !strings.Contains(fax.Data, "wp-fax-row") {
		t.Fatalf("fax payload missing the fax row: %.200s", fax.Data)
	}
	for _, forbidden := range []string{"<section", "wp-compose"} {
		if strings.Contains(fax.Data, forbidden) {
			t.Errorf("fax payload must not contain %q: %.200s", forbidden, fax.Data)
		}
	}
}
