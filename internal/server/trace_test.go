package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeTraceRecorder stands in for the composition root's adapter so the
// server's wired behavior can be asserted without touching runtime/trace.
type fakeTraceRecorder struct {
	mu         sync.Mutex
	observed   []time.Duration
	statuses   []int
	snapshotAt string
}

func (f *fakeTraceRecorder) Observe(_ context.Context, d time.Duration, status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.observed = append(f.observed, d)
	f.statuses = append(f.statuses, status)
}

func (f *fakeTraceRecorder) Snapshot(context.Context) (string, error) {
	return f.snapshotAt, nil
}

// The seam is OFF by default: the endpoint must not exist, so it answers
// the same styled page as any other unknown address.
func TestTraceSnapshotIsAbsentWhenTheSeamIsOff(t *testing.T) {
	c := signIn(t, newTestServer(t))

	resp, body := c.do(http.MethodGet, "/internal/flightrecorder", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404 (a disabled seam answers like a missing path)", resp.StatusCode)
	}
	if !strings.Contains(string(body), "wp-notfound") {
		t.Errorf("the disabled surface must render the styled page, got %.200s", body)
	}
}

// Wired, the endpoint writes a window and hands back its path. It is a
// machine surface: plain text, no HTML, no browser affordance.
func TestTraceSnapshotReportsTheWrittenPath(t *testing.T) {
	fake := &fakeTraceRecorder{snapshotAt: "/var/lib/webphone/traces/snapshot-1.trace.gz"}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Trace = fake })
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/internal/flightrecorder", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("content-type %q, want text/plain", got)
	}
	if !strings.Contains(string(body), "/var/lib/webphone/traces/snapshot-1.trace.gz") {
		t.Errorf("the snapshot path must be reported, got %.200s", body)
	}
}

// Every completed request reaches the recorder once, with the status the
// client saw — that is the whole contract the capture policy is built on.
func TestTraceMiddlewareReportsEveryRequest(t *testing.T) {
	fake := &fakeTraceRecorder{}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Trace = fake })
	c := signIn(t, server)

	if resp, _ := c.do(http.MethodGet, "/", nil, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("probe: status %d", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodGet, "/nope", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("probe: status %d", resp.StatusCode)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.observed) == 0 {
		t.Fatal("no observation reached the recorder")
	}
	sawNotFound := false
	for _, status := range fake.statuses {
		if status == http.StatusNotFound {
			sawNotFound = true
		}
	}
	if !sawNotFound {
		t.Errorf("the 404 must be reported too, got statuses %v", fake.statuses)
	}
	if len(fake.observed) != len(fake.statuses) {
		t.Errorf("durations and statuses must pair up: %d vs %d", len(fake.observed), len(fake.statuses))
	}
}
