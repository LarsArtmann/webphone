package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
