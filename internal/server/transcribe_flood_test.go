package server

import (
	"net/http"
	"testing"
)

// TestTranscribeFlood429SpansTheCombinedSurfaces pins #5's server side:
// /api/transcribe (the island's live + on-demand lane) and
// /api/transcripts (the live loop's persistence lane) share ONE
// hook-grade bucket per IP, so a transcription flood exhausts every
// surface together — the island's shared cooldown (#5 client half,
// transcribe.test.mjs) is the response to exactly this verdict.
func TestTranscribeFlood429SpansTheCombinedSurfaces(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	seen429 := false
	for range hookBurst + 5 {
		resp, body := c.do(http.MethodPost, "/api/transcribe", []byte("audio"), "application/octet-stream")
		if resp.StatusCode == http.StatusTooManyRequests {
			seen429 = true
			break
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("transcribe under budget: %d %s", resp.StatusCode, body)
		}
	}
	if !seen429 {
		t.Fatal("flood of transcription requests never got limited")
	}

	// The SAME exhausted budget answers the persistence lane: the
	// combined-surface contract the island cooldown coordinates around.
	resp, _ := c.do(http.MethodPost, "/api/transcripts",
		[]byte(`{"callId":"c","direction":"out","remote":"+49","startedAt":1,"text":"hi"}`), "application/json")
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("transcripts POST after the transcribe flood: %d, want 429 (one shared bucket)", resp.StatusCode)
	}
}
