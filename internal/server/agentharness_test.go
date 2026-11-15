package server

// agentharness_test.go is the ONE home for the cross-file agent/ASR/push
// test harness: the fake OpenAI-compatible transcription provider, the
// stub review API, the seam clients over them, and the push transport
// that records draft facts. Every handler test in the package rides the
// same seams, so the fakes live together instead of three files' worth
// of near-twins.

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/textagent"
)

// providerSeen records what the fake provider received on the wire so a
// test can assert the pass-through end to end.
type providerSeen struct {
	Audio       string
	Filename    string
	ContentType string
	Language    string
}

// fakeProvider is an OpenAI-compatible transcription stub. It records the
// last request it saw so a test can assert the bytes and the multipart
// shape forwarded end to end.
func fakeProvider(t *testing.T, status int, text string) (*httptest.Server, *providerSeen) {
	t.Helper()
	seen := &providerSeen{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("provider path: %q", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("provider multipart: %v", err)
		}
		if fhs := r.MultipartForm.File["file"]; len(fhs) == 1 {
			fh := fhs[0]
			seen.Filename = fh.Filename
			seen.ContentType = fh.Header.Get("Content-Type")
			if f, err := fh.Open(); err == nil {
				defer func() { _ = f.Close() }()
				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				seen.Audio = string(buf[:n])
			}
		}
		seen.Language = r.FormValue("language")
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_ = json.MarshalWrite(w, struct {
			Text string `json:"text"`
		}{Text: text})
	}))
	t.Cleanup(server.Close)
	return server, seen
}

func asrClientFor(t *testing.T, providerURL string) *asr.Client {
	t.Helper()
	client, err := asr.NewClient(asr.ClientConfig{BaseURL: providerURL})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// stubReviewAPI is the review-API stand-in the handlers' tests ride on:
// scripted status + payload per route, a call log the assertions read.
type stubReviewAPI struct {
	server *httptest.Server
	calls  atomic.Int32
	// draftCalls counts ONLY POST /draft — the drafting path. `calls`
	// counts every request, which is fine for "the seam was reached" but
	// useless for "the agent drafted exactly once": the server also reads
	// /reviews (the drafts panel and the session identity's agent-mode
	// probe) without drafting anything.
	draftCalls atomic.Int32
	status     atomic.Int32
	payload    string
}

func newStubReviewAPI(t *testing.T) *stubReviewAPI {
	t.Helper()
	stub := &stubReviewAPI{payload: `{"drafts":[{"id":"d-7","contact":"441632960961",
		"trigger":"inbound","text":"Guten Morgen","en_gloss":"Good morning",
		"language":"de","confidence":93,"status":"pending",
		"created_at":"2026-10-12T18:00:00Z"}],"frozen":[],"suppressed":[],
		"mode":"drafts","mode_source":"config","default_mode":"drafts",
		"modes":["off","drafts","after-hours","full"]}`}
	stub.status.Store(http.StatusOK)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		stub.calls.Add(1)
		if strings.HasSuffix(r.URL.Path, "/draft") {
			stub.draftCalls.Add(1)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(stub.status.Load()))
		if stub.status.Load() >= 400 {
			// The real agent phrases its refusals as {"error": "..."} —
			// the client extracts that field verbatim, so a passthrough
			// test can pin on the agent's own words.
			_, _ = w.Write([]byte(`{"error":"fail-closed: the agent cannot draft right now"}`))
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/approve"):
			_, _ = w.Write([]byte(`{"id":"d-7","status":"sent","provider_ref":"stub-1"}`))
		case strings.HasSuffix(r.URL.Path, "/discard"):
			_, _ = w.Write([]byte(`{"ok":true}`))
		case strings.HasSuffix(r.URL.Path, "/mode"):
			_, _ = w.Write([]byte(`{"mode":"full","mode_source":"runtime",
				"default_mode":"drafts","modes":["off","drafts","after-hours","full"]}`))
		case strings.HasSuffix(r.URL.Path, "/draft"):
			_, _ = w.Write([]byte(`{"id":"d-ond","text":"Guten Morgen","language":"de","confidence":90}`))
		default:
			_, _ = w.Write([]byte(stub.payload))
		}
	})
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	return stub
}

// agentClientFor wires a real seam client onto the given review-API URL.
func agentClientFor(t *testing.T, url string) *textagent.Client {
	t.Helper()
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("review-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	client, err := textagent.New(url, secret)
	if err != nil {
		t.Fatal(err)
	}
	return client
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

// mustExt parses a test extension or fails the test.
func mustExt(t *testing.T, raw string) domain.Extension {
	t.Helper()
	owner, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatal(err)
	}
	return owner
}
