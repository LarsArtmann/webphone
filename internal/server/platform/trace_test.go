package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type recordingTrace struct {
	observed int
}

func (r *recordingTrace) Observe(_ context.Context, _ time.Duration, _ int) {
	r.observed++
}

func (r *recordingTrace) Snapshot(_ context.Context) (string, error) { return "", nil }

// TestTraceMiddlewareSkipsMetricsScrapes pins the 2026-11-21 alignment
// ruling: the trace observer and the latency histogram share ONE
// exclusion — /metrics self-scrapes never reach Observe (and skip the
// ResponseRecorder wrap entirely; a fast 200 scrape could never fire
// the capture trigger, so observing it was dead weight).
func TestTraceMiddlewareSkipsMetricsScrapes(t *testing.T) {
	rec := &recordingTrace{}
	handler := TraceMiddleware(rec)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.observed != 0 {
		t.Fatalf("/metrics self-scrape must never be observed, got %d", rec.observed)
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.observed != 1 {
		t.Fatalf("an ordinary request must be observed exactly once, got %d", rec.observed)
	}
}
