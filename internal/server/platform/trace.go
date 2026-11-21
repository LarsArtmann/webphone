package platform

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/larsartmann/httputil"
)

// TraceRecorder is the OPTIONAL flight-recorder seam: a rolling in-memory
// execution-trace window that can be snapshotted after something went
// wrong (a slow request, an upstream 5xx). Nil — the default — means the
// deployment never pays for tracing.
//
// The concrete implementation lives in the composition root (the injector
// stays there); the server only ever sees this interface.
type TraceRecorder interface {
	// Observe reports one completed request. The implementation owns the
	// capture policy: the server states the facts, it does not decide
	// what is interesting.
	Observe(ctx context.Context, duration time.Duration, status int)
	// Snapshot writes the buffered window to disk and returns its path.
	Snapshot(ctx context.Context) (string, error)
}

// traceMiddleware reports every completed request to the recorder —
// except /metrics self-scrapes, which the latency histogram already
// excludes (the 2026-11-21 alignment ruling, TODO drain of T22 open
// leg f26: a fast 200 scrape can never fire the 2s/5xx capture
// trigger, so observing it was dead weight AND the only surface whose
// exclusion the two observers disagreed on; skipped requests also skip
// the ResponseRecorder wrap, so scrapes ride the uninstrumented
// handler). The middleware is added to the chain ONLY when the seam is
// wired, so an untraced deployment pays nothing at all.
func TraceMiddleware(rec TraceRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			recorder := httputil.NewResponseRecorder(w)
			next.ServeHTTP(recorder, r)
			status := recorder.Status()
			if status == 0 {
				status = http.StatusOK
			}
			rec.Observe(r.Context(), time.Since(start), status)
		})
	}
}

// traceSnapshot writes a trace window on demand:
// GET /internal/flightrecorder.
//
// Machine-facing and loopback-only, like the SIP directory bridge: an
// operator on the box asks for a window, nothing in the browser can reach
// it (a successful snapshot hands back a filesystem path). Disabled means
// the seam is off and the surface does not exist — the ASR-seam
// convention, where a disabled feature answers like a missing one.
func (p *Hub) TraceSnapshot(w http.ResponseWriter, r *http.Request) {
	if !machineSurface(w, r, p.server.NotFound, p.deps.Trace != nil) {
		return
	}
	path, err := p.deps.Trace.Snapshot(r.Context())
	if err != nil {
		p.server.InternalError(w, r, "write trace snapshot", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := fmt.Fprintln(w, path); err != nil {
		slog.WarnContext(r.Context(), "trace snapshot path write failed", "error", err)
	}
}
