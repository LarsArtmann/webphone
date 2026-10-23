package app

import (
	"context"
	"os"
	"testing"
	"time"
)

// A nil concrete recorder must become a TRUE nil interface: a typed nil
// reads as wired to `deps.Trace != nil`, which would put the trace
// middleware in front of a nil receiver.
func TestTraceSeamKeepsNilNil(t *testing.T) {
	if traceSeam(nil) != nil {
		t.Fatal("traceSeam(nil) must be a nil interface")
	}
}

func TestDebugFlightRecorderIsOffUnlessOptedIn(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"no", false},
		{"garbage", false},
		{"1", true},
		{"true", true},
	} {
		t.Setenv("WEBPHONE_DEBUG_FLIGHTRECORDER", tc.value)
		if got := debugFlightRecorderEnabled(); got != tc.want {
			t.Errorf("WEBPHONE_DEBUG_FLIGHTRECORDER=%q: got %v, want %v", tc.value, got, tc.want)
		}
	}
}

// The adapter really starts the runtime tracer, writes a snapshot into the
// configured directory and drains on close. One recorder per process is a
// Go constraint, so this is the only test that starts one.
func TestTraceRecorderCapturesAndCloses(t *testing.T) {
	dir := t.TempDir()
	rec, err := newTraceRecorder(dir)
	if err != nil {
		t.Fatal(err)
	}

	// A request observation must not block or panic, whatever the trigger
	// decides.
	rec.Observe(context.Background(), 10*time.Millisecond, 200)
	rec.Observe(context.Background(), 3*time.Second, 200)

	path, err := rec.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("snapshot %q was reported but not written: %v", path, statErr)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}
