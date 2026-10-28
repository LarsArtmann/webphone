package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// postReview is the automatic review dispatch: the call id rides the
// path, the body is empty.
func postReview(t *testing.T, c *client, callID string) (*http.Response, []byte) {
	t.Helper()
	return c.do(http.MethodPost, "/api/calls/"+callID+"/review", nil, "")
}

// waitFor polls until cond holds or the deadline passes — the review
// job runs in a detached goroutine, so its effects (agent call, idem
// record, slot release) land asynchronously.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not reached within 5s")
}

// TestAutoReviewDispatchRoundTrip: the happy automatic review — 202
// accepted, the detached job drafts, the idem window then answers every
// replay duplicate:true WITHOUT a second agent round trip (exactly one
// draft per call, exactly like the manual button).
func TestAutoReviewDispatchRoundTrip(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-auto", "in", "+4989123456", "guten tag"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}

	resp, body := postReview(t, c, "call-auto")
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("review dispatch: %d, want 202", resp.StatusCode)
	}
	var got struct {
		Dispatched bool `json:"dispatched"`
		Duplicate  bool `json:"duplicate"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode dispatch: %v", err)
	}
	if !got.Dispatched || got.Duplicate {
		t.Errorf("unexpected dispatch answer: %+v", got)
	}

	waitFor(t, func() bool {
		r2, b2 := postReview(t, c, "call-auto")
		if r2.StatusCode != http.StatusAccepted {
			t.Errorf("review replay: %d, want 202", r2.StatusCode)
			return true
		}
		var g2 struct {
			Duplicate bool `json:"duplicate"`
		}
		if err := json.Unmarshal(b2, &g2); err != nil {
			t.Errorf("decode replay: %v", err)
			return true
		}
		return g2.Duplicate
	})
	if calls := stub.draftCalls.Load(); calls != 1 {
		t.Errorf("agent must be drafted exactly once, got %d draft calls (of %d requests)", calls, stub.calls.Load())
	}
}

// TestAutoReviewDispatchSeamOffAndScoping: the disabled-seam 404, the
// unknown-call 404, and owner scoping (another extension's call id is
// the same 404 — no existence leak).
func TestAutoReviewDispatchSeamOffAndScoping(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")
	if resp, _ := postReview(t, c, "call-1"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("review with the seam off: %d, want 404", resp.StatusCode)
	}

	stub := newStubReviewAPI(t)
	c2 := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c2.login("1001", "pw")
	other := newClient(t)
	other.login("2002", "pw")
	if resp, _ := postTranscript(t, c2, "call-scope", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	if resp, _ := postReview(t, other, "call-scope"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("another owner's call: %d, want 404", resp.StatusCode)
	}
	if resp, _ := postReview(t, c2, "call-unknown"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown call: %d, want 404", resp.StatusCode)
	}
	if calls := stub.draftCalls.Load(); calls != 0 {
		t.Errorf("no dispatch may reach the agent on 404 paths, got %d draft calls (of %d requests)", calls, stub.calls.Load())
	}
}

// TestAutoReviewDispatchTranscriptsNilIs503: a wired agent beside a nil
// transcript store (test compositions) answers the honest 503 — the
// job must never start from an input it cannot read.
func TestAutoReviewDispatchTranscriptsNilIs503(t *testing.T) {
	stub := newStubReviewAPI(t)
	srv := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.TextAgent = agentClientFor(t, stub.server.URL)
		d.Transcripts = nil
	})
	c := clientFor(t, srv)
	c.login("1001", "pw")
	if resp, _ := postReview(t, c, "call-x"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil transcript store: %d, want 503", resp.StatusCode)
	}
}

// TestAutoReviewDispatchInvalidCallID: an oversized opaque id is a 400
// before any store or dispatch touches it.
func TestAutoReviewDispatchInvalidCallID(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c.login("1001", "pw")
	resp, _ := postReview(t, c, strings.Repeat("x", 129))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("oversized call id: %d, want 400", resp.StatusCode)
	}
}

// TestAutoReviewFailureStaysRetryable: an agent outage fails the job
// WITHOUT recording the idem key and releases the in-flight slot — the
// next dispatch reaches the agent again (the manual button and a
// re-settle both stay live after a failed attempt).
func TestAutoReviewFailureStaysRetryable(t *testing.T) {
	stub := newStubReviewAPI(t)
	stub.status.Store(http.StatusServiceUnavailable)
	c := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-dead", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}

	if resp, _ := postReview(t, c, "call-dead"); resp.StatusCode != http.StatusAccepted {
		t.Fatalf("dispatch during outage: %d, want 202 (failure is async)", resp.StatusCode)
	}
	waitFor(t, func() bool { return stub.calls.Load() >= 1 })
	stub.status.Store(http.StatusOK)

	waitFor(t, func() bool {
		_, _ = postReview(t, c, "call-dead")
		return stub.calls.Load() >= 2
	})
}

// TestAutoReviewInFlightGuardDeduplicates: while the job is mid-flight
// (a blocked agent holds the slot), a second dispatch answers
// duplicate:true WITHOUT queueing a second draft — the guard closes the
// idem window's success-only blind spot.
func TestAutoReviewInFlightGuardDeduplicates(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	draftCalls := 0
	agentMux := http.NewServeMux()
	agentMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/draft") {
			draftCalls++
			select {
			case entered <- struct{}{}:
			case <-time.After(time.Second):
			}
			<-release
			_, _ = w.Write([]byte(`{"id":"d-race","text":"Guten Morgen","language":"de","confidence":90}`))
			return
		}
		_, _ = w.Write([]byte(`{"mode":"drafts","mode_source":"config","default_mode":"drafts","modes":["off","drafts"]}`))
	})
	agentSrv := httptest.NewServer(agentMux)
	t.Cleanup(agentSrv.Close)

	c := clientFor(t, serverWithAgentAndASR(t, agentSrv.URL))
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-race", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}

	resp, _ := postReview(t, c, "call-race")
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("first dispatch: %d, want 202", resp.StatusCode)
	}
	<-entered
	var got struct {
		Duplicate bool `json:"duplicate"`
	}
	resp, body := postReview(t, c, "call-race")
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("in-flight dispatch: %d, want 202", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode in-flight replay: %v", err)
	}
	if !got.Duplicate {
		t.Errorf("the in-flight replay must answer duplicate:true, got %+v", got)
	}

	close(release)
	waitFor(t, func() bool {
		r2, _ := postReview(t, c, "call-race")
		return r2 != nil && r2.StatusCode == http.StatusAccepted
	})
	if draftCalls != 1 {
		t.Errorf("agent must be drafted exactly once despite the mid-flight replay, got %d calls", draftCalls)
	}
}
