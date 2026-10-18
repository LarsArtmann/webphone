package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// TestTranscribeCeilingBoundsInFlight pins the memory ceiling (#47):
// at most transcribeConcurrency requests hold provider slots; the rest
// wait one queue window and then answer 503 (the island's backoff
// absorbs it) instead of piling unbounded buffered bodies.
func TestTranscribeCeilingBoundsInFlight(t *testing.T) {
	release := make(chan struct{})
	var inFlight atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		inFlight.Add(1)
		<-release
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	defer provider.Close()

	client, err := asr.NewClient(provider.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	h := New(Deps{Server: contextSessionGate{}, ASR: client})

	previousWait := transcribeQueueWait
	transcribeQueueWait = 50 * time.Millisecond
	t.Cleanup(func() { transcribeQueueWait = previousWait })

	const burst = 6
	var releaseOnce sync.Once
	codes := make(chan int, burst)
	for range burst {
		go func() {
			ctx := session.With(context.Background(), session.Session{Extension: domain.MustParseExtension("1001")})
			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/transcribe", bytes.NewReader([]byte("audio-bytes")))
			rec := httptest.NewRecorder()
			h.Transcribe(rec, req)
			codes <- rec.Code
		}()
	}

	// The over-ceiling requests 503 after the (shortened) queue wait.
	got := map[int]int{}
	for saw := 0; saw < burst; {
		select {
		case code := <-codes:
			got[code]++
			saw++
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out collecting verdicts; got %v", got)
		}
		if got[http.StatusServiceUnavailable] == burst-transcribeConcurrency {
			releaseOnce.Do(func() { close(release) })
		}
	}
	releaseOnce.Do(func() { close(release) })

	if got[http.StatusServiceUnavailable] != burst-transcribeConcurrency {
		t.Errorf("503s: %d, want exactly the over-ceiling %d (verdicts: %v)",
			got[http.StatusServiceUnavailable], burst-transcribeConcurrency, got)
	}
	if got[http.StatusOK] != transcribeConcurrency {
		t.Errorf("200s: %d, want exactly the in-ceiling %d (verdicts: %v)",
			got[http.StatusOK], transcribeConcurrency, got)
	}
}
