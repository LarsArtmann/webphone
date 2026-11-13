package platform

import (
	"fmt"
	"strings"
	"sync"
)

// TraceSnapshotMetrics aggregates the flight recorder's snapshot outcome
// counters. The library's WithMetrics hook feeds it (every capture that
// reaches the write stage — manual, trigger and async alike), and the
// adapter's own disk guard feeds the skip counter.
//
// AGGREGATES ONLY, the same rule the rest of /metrics follows: no path,
// no request id, no operation label beyond the server-fixed outcome
// vocabulary, so a scrape of the debug seam leaks nothing.
//
// Safe for concurrent use.
type TraceSnapshotMetrics struct {
	mu sync.Mutex
	// written / failed count the library's capture attempts (err nil /
	// non-nil in the metrics hook); skipped counts the adapter's
	// disk-guard refusals, which never reach the library.
	written uint64
	failed  uint64
	skipped uint64
	bytes   uint64
}

func NewTraceSnapshotMetrics() *TraceSnapshotMetrics { return &TraceSnapshotMetrics{} }

// RecordWritten folds one successful capture (the metrics hook with a
// nil error): the counter and the byte total.
func (m *TraceSnapshotMetrics) RecordWritten(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.written++
	if bytes > 0 {
		m.bytes += uint64(bytes) //nolint:gosec // a negative byte count is not a real capture; the guard above keeps it out
	}
}

// RecordFailed folds one failed capture attempt.
func (m *TraceSnapshotMetrics) RecordFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failed++
}

// RecordSkipped folds one disk-guard refusal.
func (m *TraceSnapshotMetrics) RecordSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.skipped++
}

// writePrometheus appends the snapshot counters. The family renders ONLY
// when the recorder seam is on (the app threads nil otherwise) — a
// disabled deploy publishes no zero-lines that read as "tracing
// broken".
func (m *TraceSnapshotMetrics) writePrometheus(b *strings.Builder) {
	m.mu.Lock()
	written, failed, skipped, bytes := m.written, m.failed, m.skipped, m.bytes
	m.mu.Unlock()

	fmt.Fprintf(b, "# HELP webphone_flightrecorder_snapshots_total Flight-recorder snapshot write attempts by outcome since process start (written = on disk, failed = the library refused or errored, skipped = the disk guard refused before the write).\n")
	fmt.Fprintf(b, "# TYPE webphone_flightrecorder_snapshots_total counter\n")
	fmt.Fprintf(b, "webphone_flightrecorder_snapshots_total{outcome=\"written\"} %d\n", written)
	fmt.Fprintf(b, "webphone_flightrecorder_snapshots_total{outcome=\"failed\"} %d\n", failed)
	fmt.Fprintf(b, "webphone_flightrecorder_snapshots_total{outcome=\"skipped_low_disk\"} %d\n", skipped)

	fmt.Fprintf(b, "# HELP webphone_flightrecorder_snapshot_bytes_total Compressed bytes written to the snapshot sink since process start.\n")
	fmt.Fprintf(b, "# TYPE webphone_flightrecorder_snapshot_bytes_total counter\n")
	fmt.Fprintf(b, "webphone_flightrecorder_snapshot_bytes_total %d\n", bytes)
}
