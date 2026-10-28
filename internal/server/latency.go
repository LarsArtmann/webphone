package server

import (
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"
)

// latencyBuckets are the histogram's upper bounds in seconds. They follow
// the shape of this app's request mix: sub-10 ms for cached panels, tens of
// milliseconds for WebAuthn and provider round-trips, seconds for the rare
// upstream timeout. The last entry before +Inf is deliberately generous —
// anything slower is a transport story, not a latency distribution.
var latencyBuckets = [...]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30}

// latencyRecorder aggregates request durations into a Prometheus-style
// histogram plus a status-class counter.
//
// AGGREGATES ONLY, the same rule the rest of /metrics follows: no path, no
// extension and no request id ever becomes a label, so a scrape cannot turn
// into a data leak or blow up cardinality. The one exclusion is /metrics
// itself — self-scrapes would otherwise dominate the slow tail.
//
// Safe for concurrent use (satisfies httputil.MetricsRecorder).
type latencyRecorder struct {
	mu sync.Mutex
	// counts[i] is the number of observations that fell in bucket i
	// (non-cumulative; the exposition accumulates).
	counts [len(latencyBuckets) + 1]uint64
	sum    float64
	total  uint64
	// classes maps the status class (2xx -> 2) to its request count.
	classes map[int]uint64
}

func newLatencyRecorder() *latencyRecorder {
	return &latencyRecorder{classes: map[int]uint64{}}
}

// Record implements httputil.MetricsRecorder. The method and path arguments
// are accepted for interface conformance only: neither is labelled.
func (l *latencyRecorder) Record(_, path string, status int, d time.Duration) {
	if path == "/metrics" {
		return
	}
	secs := d.Seconds()
	idx := len(latencyBuckets)
	for i, bound := range latencyBuckets {
		if secs <= bound {
			idx = i
			break
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[idx]++ //nolint:branching-flow // counts is sized len(latencyBuckets)+1 and idx defaults to the overflow bucket, so the index is always in range
	l.sum += secs
	l.total++
	l.classes[status/100]++
}

// writePrometheus appends the histogram and the status-class counter to the
// exposition. Cumulative bucket semantics are built here, not stored, so the
// hot path stays one increment.
func (l *latencyRecorder) writePrometheus(b *strings.Builder) {
	l.mu.Lock()
	counts := l.counts
	sum, total := l.sum, l.total
	classes := make(map[int]uint64, len(l.classes))
	maps.Copy(classes, l.classes)
	l.mu.Unlock()

	fmt.Fprintf(b, "# HELP webphone_http_request_duration_seconds Wall-clock duration of completed HTTP requests (aggregates only; /metrics self-scrapes excluded).\n")
	fmt.Fprintf(b, "# TYPE webphone_http_request_duration_seconds histogram\n")
	var cumulative uint64
	for i, bound := range latencyBuckets {
		cumulative += counts[i] //nolint:branching-flow // i ranges over latencyBuckets, which sizes counts
		fmt.Fprintf(b, "webphone_http_request_duration_seconds_bucket{le=\"%g\"} %d\n", bound, cumulative)
	}
	cumulative += counts[len(latencyBuckets)] //nolint:branching-flow // the overflow bucket exists: counts is sized len(latencyBuckets)+1
	fmt.Fprintf(b, "webphone_http_request_duration_seconds_bucket{le=\"+Inf\"} %d\n", cumulative)
	fmt.Fprintf(b, "webphone_http_request_duration_seconds_sum %g\n", sum)
	fmt.Fprintf(b, "webphone_http_request_duration_seconds_count %d\n", total)

	fmt.Fprintf(b, "# HELP webphone_http_requests_total Completed HTTP requests by status class.\n")
	fmt.Fprintf(b, "# TYPE webphone_http_requests_total counter\n")
	for _, class := range []int{1, 2, 3, 4, 5} {
		fmt.Fprintf(b, "webphone_http_requests_total{status_class=\"%dxx\"} %d\n", class, classes[class])
	}
}
