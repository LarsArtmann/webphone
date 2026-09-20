package server

import (
	"bufio"
	"context"
	"net/http"
	"strconv"
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

// TestSSEStreamCarriesConnectedThenEvents pins the wire shape the library
// ServeSSE produces (tag-verified at v4.11.0): exactly one `retry:` reconnect
// hint frame, then the `connected` handshake (`event:` + `data:` lines), then
// broadcasts as `event: <name>` + swap-safe data. The htmx sse extension
// consumes the retry hint; sse-swap listeners ignore the connected frame.
func TestSSEStreamCarriesConnectedThenEvents(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("events status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("events content-type %q", ct)
	}

	reader := bufio.NewReader(resp.Body)

	// v4.11.0 leads with exactly one reconnect hint (sse.WriteRetry:
	// `retry: <millis>` + blank frame terminator), then the connected
	// handshake frame — assert the shape instead of skipping past it, so a
	// future bump that changes the stream head fails here, loudly.
	retryLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no retry hint line: %v", err)
	}
	if !strings.HasPrefix(retryLine, "retry: ") {
		t.Fatalf("first stream line %q, want a `retry:` reconnect hint", retryLine)
	}
	if millis, convErr := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(retryLine, "retry: "))); convErr != nil || millis <= 0 {
		t.Fatalf("retry hint %q is not a positive millisecond value", retryLine)
	}
	blank, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no blank line after retry hint: %v", err)
	}
	if blank != "\n" {
		t.Fatalf("line after retry hint %q, want the blank frame terminator", blank)
	}

	head, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no connected frame: %v", err)
	}
	if head != "event: connected\n" {
		t.Errorf("first event line %q, want %q", head, "event: connected\n")
	}
	data, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no connected data line: %v", err)
	}
	if data != "data: connected\n" {
		t.Errorf("connected data line %q, want %q", data, "data: connected\n")
	}

	// Broadcast after connect: arrives as event + data on the same stream.
	// Blank lines are frame terminators — skip them while scanning for the
	// next event line.
	ext := domain.MustParseExtension("1001")
	server.hubs.Publish(ext, sseEventThreads, "<div class=\"wp-thread-row\">wire</div>")

	var line string
	for {
		var err error
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("no threads event line: %v", err)
		}
		if line != "\n" {
			break
		}
	}
	if line != "event: threads\n" {
		t.Errorf("event line %q, want %q", line, "event: threads\n")
	}
	dataLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no threads data line: %v", err)
	}
	if !strings.Contains(dataLine, "wp-thread-row") {
		t.Errorf("threads data line missing the fragment: %q", dataLine)
	}
}
