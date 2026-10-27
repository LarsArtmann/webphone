package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// postSummaryDraft is the island's summarize ask: the call id rides the
// path, the body is empty.
func postSummaryDraft(t *testing.T, c *client, callID string) (*http.Response, []byte) {
	t.Helper()
	return c.do(http.MethodPost, "/api/calls/"+callID+"/summary-draft", nil, "")
}

// serverWithAgentAndASR wires the textagent seam to agentURL AND turns
// the ASR seam on: the summary chain reads transcripts, and the capture
// lane (/api/transcripts) only exists while ASR is enabled — without the
// provider stub every seed POST would 404 at the gate.
func serverWithAgentAndASR(t *testing.T, agentURL string) *testServer {
	t.Helper()
	provider, _ := fakeProvider(t, http.StatusOK, "hi")
	return newTestServerWithPhoneAPI(t, "", func(deps *Deps) {
		deps.ASR = asrClientFor(t, provider.URL)
		deps.TextAgent = agentClientFor(t, agentURL)
	})
}

// TestConfigJSCarriesAgentFlag pins the island's summarize gate (M12):
// the served window.PBX_CONFIG mirrors the textagent seam's presence —
// on with a client wired, absent (the zero value) without.
func TestConfigJSCarriesAgentFlag(t *testing.T) {
	stub := newStubReviewAPI(t)
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.TextAgent = agentClientFor(t, stub.server.URL)
	})
	c := clientFor(t, server)
	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js: %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"agent":true`) {
		t.Errorf("config.js must carry agent:true with the seam wired: %.300s", body)
	}

	off := newTestServerWithPhoneAPI(t, "")
	oc := clientFor(t, off)
	_, obody := oc.do(http.MethodGet, "/config.js", nil, "")
	if strings.Contains(string(obody), `"agent":true`) {
		t.Errorf("config.js must not carry agent:true without the seam: %.300s", obody)
	}
}

// TestSummaryDraftSeamOffIs404: a server without the textagent seam has
// no summarize surface (the disabled-seam convention).
func TestSummaryDraftSeamOffIs404(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")
	resp, _ := postSummaryDraft(t, c, "call-1")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("summary draft with the seam off: %d, want 404", resp.StatusCode)
	}
}

func TestSummaryDraftRoundTripOwnerScopedAndIdempotent(t *testing.T) {
	stub := newStubReviewAPI(t)
	c := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c.login("1001", "pw")
	other := newClient(t)
	other.login("2002", "pw")

	text := "guten tag, hier ist die zusammenfassung"
	if resp, _ := postTranscript(t, c, "call-9", "in", "+4989123456", text); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}

	resp, body := postSummaryDraft(t, c, "call-9")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary draft: %d, want 200", resp.StatusCode)
	}
	var got struct {
		ID        string `json:"id"`
		Contact   string `json:"contact"`
		Text      string `json:"text"`
		Duplicate bool   `json:"duplicate"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "d-ond" || !strings.Contains(got.Text, "Guten Morgen") || got.Duplicate {
		t.Errorf("unexpected draft answer: %+v", got)
	}
	callsAfterFirst := stub.calls.Load()

	// The idempotency window: an immediate replay answers duplicate:true
	// WITHOUT another agent round trip.
	resp, body = postSummaryDraft(t, c, "call-9")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary replay: %d, want 200", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if !got.Duplicate {
		t.Errorf("the replay must answer duplicate:true, got %+v", got)
	}
	if stub.calls.Load() != callsAfterFirst {
		t.Errorf("the replay must not reach the agent (calls %d -> %d)", callsAfterFirst, stub.calls.Load())
	}

	// Owner scoping: another extension's call id is the same 404.
	if resp, _ := postSummaryDraft(t, other, "call-9"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("another owner's call: %d, want 404", resp.StatusCode)
	}

	// An unknown call (no transcript) is 404 too.
	if resp, _ := postSummaryDraft(t, c, "call-unknown"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown call: %d, want 404", resp.StatusCode)
	}
}

// TestSummaryDraftAgentVerdictPassesThrough: an agent-side refusal
// (4xx/5xx Error verdict) reaches the client with the agent's status
// and its own human-phrased message — never a re-guessed generic.
func TestSummaryDraftAgentVerdictPassesThrough(t *testing.T) {
	stub := newStubReviewAPI(t)
	stub.status.Store(http.StatusServiceUnavailable)
	c := clientFor(t, serverWithAgentAndASR(t, stub.server.URL))
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-v", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, body := postSummaryDraft(t, c, "call-v")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("agent 503 verdict: %d, want 503 passthrough", resp.StatusCode)
	}
	if !strings.Contains(string(body), "fail-closed") {
		t.Errorf("the agent's own phrasing must ride the answer, got: %s", body)
	}
}

// TestSummaryDraftOutageIs502: a dead agent (transport failure, not a
// verdict) answers 502 so the island toasts and the attempt stays
// retryable (the idempotency key never recorded).
func TestSummaryDraftOutageIs502(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	hc := clientFor(t, serverWithAgentAndASR(t, dead.URL))
	hc.login("1001", "pw")
	if resp, _ := postTranscript(t, hc, "call-dead", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, _ := postSummaryDraft(t, hc, "call-dead")
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("dead agent: %d, want 502", resp.StatusCode)
	}
}

// mustExt parses a test extension or fails the test.
func mustExt(t *testing.T, raw string) domain.Extension {
	t.Helper()
	owner, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatal(err)
	}
	return owner
}

// recordingTransport satisfies push.Sender and records the draft facts
// shipped through it (the M14 fan-out specs). The fan-out runs in a
// fire-and-forget goroutine, so every field access is mutex-guarded.
type recordingTransport struct {
	mu      sync.Mutex
	enabled bool
	drafts  []push.Draft
	err     error
}

func (r *recordingTransport) Enabled() bool    { return r.enabled }
func (r *recordingTransport) Describe() string { return "recording" }
func (r *recordingTransport) Send(context.Context, string, push.Ring) error {
	return nil
}
func (r *recordingTransport) SendDraft(_ context.Context, _ string, draft push.Draft) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drafts = append(r.drafts, draft)
	return r.err
}

// snapshot returns a copy of the recorded drafts (race-free read).
func (r *recordingTransport) snapshot() []push.Draft {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]push.Draft(nil), r.drafts...)
}

// waitForDrafts polls until the fan-out goroutine recorded n facts (or
// times out) — deterministic, no sleep windows.
func waitForDrafts(t *testing.T, transport *recordingTransport, n int) []push.Draft {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		drafts := transport.snapshot()
		if len(drafts) >= n {
			return drafts
		}
		if time.Now().After(deadline) {
			t.Fatalf("fan-out recorded %d draft facts, want %d", len(drafts), n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// summaryServerWithPush wires the full summarize chain plus a push
// seam over the engine's device registry; the registry handle comes
// back so a test can register devices directly.
func summaryServerWithPush(t *testing.T, transport push.Sender) (*testServer, *client, *readmodels.PushDevices) {
	t.Helper()
	stub := newStubReviewAPI(t)
	engine := newTestEngine(t)
	fanout := push.NewFanout(map[string]push.Sender{"recording": transport})
	provider, _ := fakeProvider(t, http.StatusOK, "hi")
	server := newTestServerWithPhoneAPI(t, "", func(deps *Deps) {
		deps.ASR = asrClientFor(t, provider.URL)
		deps.TextAgent = agentClientFor(t, stub.server.URL)
		deps.Push = fanout
		deps.PushDevices = engine.PushDevices()
	})
	return server, clientFor(t, server), engine.PushDevices()
}

// TestSummaryDraftPushesOnlyWhenSeamOn: a fanout with no live transport
// ships nothing, and the summarize answer is unaffected either way.
func TestSummaryDraftPushesOnlyWhenSeamOn(t *testing.T) {
	transport := &recordingTransport{enabled: false}
	_, c, _ := summaryServerWithPush(t, transport)
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-p1", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, _ := postSummaryDraft(t, c, "call-p1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary: %d", resp.StatusCode)
	}
	// Give the fire-and-forget goroutine its window; a disabled
	// transport must still receive nothing.
	waitForDrafts(t, func() *recordingTransport { return transport }(), 0)
	if drafts := transport.snapshot(); len(drafts) != 0 {
		t.Fatalf("a disabled transport must receive nothing, got %v", drafts)
	}
}

// TestSummaryDraftFansOutToRegisteredDevices: the draft-ready fact
// reaches every registered device of the OWNING extension only, with
// the agent's draft id and the contact on the wire.
func TestSummaryDraftFansOutToRegisteredDevices(t *testing.T) {
	transport := &recordingTransport{enabled: true}
	server, c, devices := summaryServerWithPush(t, transport)
	ctx := context.Background()
	for _, device := range []string{"dev-1", "dev-2"} {
		if err := devices.Upsert(ctx, mustExt(t, "1001"), device, "recording", "tok-"+device); err != nil {
			t.Fatal(err)
		}
	}
	// Another extension's device must never receive owner 1001's fact.
	other := clientFor(t, server)
	other.login("2002", "pw")
	if err := devices.Upsert(ctx, mustExt(t, "2002"), "dev-other", "recording", "tok-other"); err != nil {
		t.Fatal(err)
	}

	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-p2", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, body := postSummaryDraft(t, c, "call-p2")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary: %d %s", resp.StatusCode, body)
	}
	drafts := waitForDrafts(t, transport, 2)
	for _, fact := range drafts {
		if fact.Event != push.EventDraftReady || fact.Contact != "+4989123456" || fact.DraftID != "d-ond" {
			t.Fatalf("draft fact mismatch: %+v", fact)
		}
	}
	// The other extension's device row survives untouched (owner scoping
	// at the registry, not a broadcast).
	live, err := devices.List(ctx, mustExt(t, "2002"))
	if err != nil || len(live) != 1 {
		t.Fatalf("other owner's device must survive: %v %v", live, err)
	}
}

// TestSummaryDraftPushPrunesDeadDevice: a transport verdict of
// ErrUnregistered prunes the registration exactly like the ring lane.
func TestSummaryDraftPushPrunesDeadDevice(t *testing.T) {
	transport := &recordingTransport{enabled: true, err: push.ErrUnregistered}
	_, c, devices := summaryServerWithPush(t, transport)
	ctx := context.Background()
	if err := devices.Upsert(ctx, mustExt(t, "1001"), "dev-dead", "recording", "tok-dead"); err != nil {
		t.Fatal(err)
	}
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-p3", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, _ := postSummaryDraft(t, c, "call-p3")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary: %d", resp.StatusCode)
	}
	// The dead-device prune happens inside the same goroutine; wait for
	// the send attempt before reading the registry.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if live, _ := devices.List(ctx, mustExt(t, "1001")); len(live) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the dead device was never pruned")
		}
		time.Sleep(5 * time.Millisecond)
	}
	live, err := devices.List(ctx, mustExt(t, "1001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("the dead device must be pruned, still %v", live)
	}
}

// panicTransport satisfies push.Sender and detonates inside SendDraft:
// the containment pin for the fire-and-forget fan-out goroutine.
type panicTransport struct {
	called atomic.Bool
}

func (p *panicTransport) Enabled() bool    { return true }
func (p *panicTransport) Describe() string { return "panic" }
func (p *panicTransport) Send(context.Context, string, push.Ring) error {
	return nil
}
func (p *panicTransport) SendDraft(context.Context, string, push.Draft) error {
	p.called.Store(true)
	panic("push transport exploded mid-send")
}

// TestSummaryDraftPushPanicDoesNotKillServer pins the goroutine
// containment (owner ruling 2026-10-27): the fan-out runs OUTSIDE
// RecoveryMiddleware's reach, so a panicking transport would kill the
// whole process without the recover guard. The summarize answer must
// still be 200 and the server must keep serving afterwards.
func TestSummaryDraftPushPanicDoesNotKillServer(t *testing.T) {
	transport := &panicTransport{}
	_, c, devices := summaryServerWithPush(t, transport)
	ctx := context.Background()
	if err := devices.Upsert(ctx, mustExt(t, "1001"), "dev-panic", "panic", "tok"); err != nil {
		t.Fatal(err)
	}
	c.login("1001", "pw")
	if resp, _ := postTranscript(t, c, "call-panic", "in", "+4989123456", "words"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed transcript: %d", resp.StatusCode)
	}
	resp, _ := postSummaryDraft(t, c, "call-panic")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary answer must not wait on push health: %d, want 200", resp.StatusCode)
	}
	deadline := time.Now().Add(2 * time.Second)
	for !transport.called.Load() {
		if time.Now().After(deadline) {
			t.Fatal("the panicking transport was never invoked")
		}
		time.Sleep(5 * time.Millisecond)
	}
	// A process without the guard would be dead before this answers.
	resp2, _ := postSummaryDraft(t, c, "call-panic")
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("server must survive the fan-out panic: %d, want 200", resp2.StatusCode)
	}
}
