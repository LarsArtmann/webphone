package platform

import (
	"strings"
	"testing"
)

// The counter family renders the server-fixed outcome vocabulary and the
// byte total; nothing else (no paths, no operation kinds — the debug
// seam must not leak through its own observability).
func TestTraceSnapshotMetricsExposition(t *testing.T) {
	m := NewTraceSnapshotMetrics()
	m.RecordWritten(1024)
	m.RecordWritten(512)
	m.RecordFailed()
	m.RecordSkipped()

	var b strings.Builder
	m.writePrometheus(&b)
	page := b.String()
	for _, want := range []string{
		"webphone_flightrecorder_snapshots_total{outcome=\"written\"} 2",
		"webphone_flightrecorder_snapshots_total{outcome=\"failed\"} 1",
		"webphone_flightrecorder_snapshots_total{outcome=\"skipped_low_disk\"} 1",
		"webphone_flightrecorder_snapshot_bytes_total 1536",
		"# TYPE webphone_flightrecorder_snapshots_total counter",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("exposition missing %q:\n%s", want, page)
		}
	}
	for _, leak := range []string{".trace", "/tmp", "http.request", "manual"} {
		if strings.Contains(page, leak) {
			t.Errorf("exposition leaks %q — aggregates only:\n%s", leak, page)
		}
	}
}
