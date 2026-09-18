package server

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestIdemStoreSeenRecord pins the dedupe primitives: a key is a replay
// only after record() (failed attempts are never recorded), and an
// expired entry is no longer a replay.
func TestIdemStoreFirstTime(t *testing.T) {
	s := newIdemStore(50 * time.Millisecond)
	if s.seen("k") {
		t.Fatal("unrecorded key reported as replay")
	}
	s.record("k")
	if !s.seen("k") {
		t.Fatal("recorded key not reported as replay")
	}
	time.Sleep(60 * time.Millisecond)
	if s.seen("k") {
		t.Fatal("expired entry still reported as replay")
	}
}

// TestWebhookStatusIdempotent pins the replay contract end to end: the
// same provider_ref verdict answers 202 on every replay, and a replay
// carrying DIFFERENT data cannot touch the store again — the panel still
// shows the first verdict. A failed attempt (unknown ref) is not recorded,
// so a genuine retry of the same ref stays possible.
func TestWebhookStatusIdempotent(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	pdf := []byte("%PDF-1.4 pages\n%%EOF\n")
	form, contentType := multipartBody(t,
		map[string]string{"to": "+441632960961"},
		map[string]struct {
			Name    string
			Content []byte
		}{"document": {Name: "doc.pdf", Content: pdf}})
	if resp, body := c.do(http.MethodPost, "/fax/send", form, contentType); resp.StatusCode != http.StatusOK {
		t.Fatalf("fax send: %d %s", resp.StatusCode, body)
	}
	owner, err := domain.ParseExtension("1001")
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := server.faxes.List(context.Background(), owner, 10)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("seeded fax jobs: %d (err %v)", len(jobs), err)
	}
	ref := jobs[0].ProviderRef

	// First verdict: transmitted, 3 pages.
	if got, body := postHook(t, server, "/hooks/fax/status",
		`{"provider_ref":"`+ref+`","status":"transmitted","pages":3}`); got != http.StatusAccepted {
		t.Fatalf("first verdict: %d %s", got, body)
	}

	// Replay, same payload: 202, no second write.
	if got, _ := postHook(t, server, "/hooks/fax/status",
		`{"provider_ref":"`+ref+`","status":"transmitted","pages":3}`); got != http.StatusAccepted {
		t.Fatalf("same-payload replay: %d", got)
	}

	// Replay, CONTRADICTORY payload (9 pages, failed): must be inert —
	// 202 with the stored job untouched.
	if got, _ := postHook(t, server, "/hooks/fax/status",
		`{"provider_ref":"`+ref+`","status":"failed","pages":9}`); got != http.StatusAccepted {
		t.Fatalf("contradictory replay: %d", got)
	}
	_, panel := c.do(http.MethodGet, "/partials/fax", nil, "")
	if !strings.Contains(string(panel), "3 page(s)") {
		t.Fatalf("replay mutated the job; panel: %.300s", panel)
	}
	if strings.Contains(string(panel), "9 page(s)") {
		t.Fatalf("replay wrote 9 pages; panel: %.300s", panel)
	}

	// Failed attempts are not recorded: an unknown ref 404s twice.
	for i := 0; i < 2; i++ {
		if got, _ := postHook(t, server, "/hooks/fax/status",
			`{"provider_ref":"unknown-ref","status":"failed"}`); got != http.StatusNotFound {
			t.Fatalf("unknown ref attempt %d: %d, want 404 (retryable)", i+1, got)
		}
	}
}
