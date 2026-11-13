package app

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	flightrecorder "github.com/larsartmann/go-flightrecorder"

	"github.com/larsartmann/webphone/internal/server/platform"
)

// The flight recorder is an OPT-IN debugging seam: it keeps a rolling
// window of the last seconds of execution trace in memory and writes a
// snapshot when a request turns slow or answers 5xx. Off by default — the
// runtime tracer costs CPU and the snapshots are operator artifacts, not
// user data.
const (
	// traceMinAge is ~2x the debug window: the trace data the recorder
	// promises to have retained before a snapshot is requested.
	traceMinAge = 10 * time.Second
	// traceSlowRequest is the capture threshold. Every request slower than
	// this, and every 5xx, lands a snapshot; anything faster is ordinary
	// traffic and would only fill the directory.
	traceSlowRequest = 2 * time.Second
	// traceRetention bounds the snapshot directory, so a pathological
	// upstream cannot fill the disk with trace files.
	traceRetention = 20
	// traceDiskFloor is the free-space floor the disk guard enforces
	// before any snapshot write: below it the recorder stops capturing
	// (and counts the skip), so a pathological upstream cannot fill the
	// DISK — the count bound alone assumes every snapshot is small.
	traceDiskFloor = uint64(1) << 30 // 1 GiB
)

// errTraceRequestFailed is the trigger vocabulary's error signal, used for
// an HTTP 5xx: the recorder library reasons about operations that returned
// errors, and a request that answered 5xx is this deployment's equivalent.
var errTraceRequestFailed = fmt.Errorf("request answered 5xx")

// debugFlightRecorderEnabled reads the opt-in once at construction, the
// same convention WEBPHONE_DEBUG_TIMING follows: a debug surface an
// operator turns on with an environment variable, never a config file.
func debugFlightRecorderEnabled() bool {
	on, err := strconv.ParseBool(os.Getenv("WEBPHONE_DEBUG_FLIGHTRECORDER"))
	return err == nil && on
}

// traceRecorder adapts go-flightrecorder to the platform package's TraceRecorder
// seam. It implements do.Shutdowner so the container drains in-flight
// captures on exit (the library blocks Close until they finish).
type traceRecorder struct {
	rec     *flightrecorder.Recorder
	trigger flightrecorder.TriggerFunc
	metrics *platform.TraceSnapshotMetrics
	dir     string
}

// diskHasRoom reports whether the snapshot directory's filesystem still
// has more free space than the floor. Best-effort: a stat failure reads
// as "room" — the capture proceeds and the library's own write error
// handling owns the failure (a broken mount must not silently disable
// tracing, it must fail the capture visibly).
func diskHasRoom(dir string) bool {
	var st syscall.Statfs_t
	if err := syscall.Statfs(dir, &st); err != nil {
		return true
	}
	return st.Bavail*uint64(st.Bsize) > traceDiskFloor
}

// newTraceRecorder starts the recorder writing snapshots into dir. A
// failure here is a boot failure, not a silent degradation: the operator
// asked for tracing and must learn immediately that it is not running.
func newTraceRecorder(dir string) (*traceRecorder, error) {
	metrics := platform.NewTraceSnapshotMetrics()
	rec, err := flightrecorder.New(
		flightrecorder.WithSnapshotDir(dir),
		flightrecorder.WithMaxSnapshots(traceRetention),
		flightrecorder.WithCompression(gzip.BestSpeed),
		flightrecorder.WithMinAge(traceMinAge),
		// Every capture that reaches the write stage folds into the
		// /metrics counter family — the debug tool observing itself
		// (T13): an operator can see captures happening and how much
		// disk they eat without grepping the directory.
		flightrecorder.WithMetrics(func(e flightrecorder.SnapshotEvent, err error) {
			if err != nil {
				metrics.RecordFailed()
				return
			}
			metrics.RecordWritten(e.Bytes)
		}),
	)
	if err != nil {
		return nil, wrapf("configure flight recorder: %w", err)
	}
	if err := rec.Start(); err != nil {
		return nil, wrapf("start flight recorder: %w", err)
	}
	base := flightrecorder.OnErrorOrLatency(traceSlowRequest)
	return &traceRecorder{
		rec: rec,
		// The disk guard composes into the trigger: below the floor the
		// auto-capture path refuses BEFORE the write (the skip is
		// counted, the disk is spared).
		trigger: func(tc flightrecorder.TriggerContext) bool {
			if !diskHasRoom(dir) {
				metrics.RecordSkipped()
				return false
			}
			return base(tc)
		},
		metrics: metrics,
		dir:     dir,
	}, nil
}

// Observe implements platform.TraceRecorder. The request context is detached
// from cancellation: the snapshot write outlives the response, and the
// library skips a capture whose context is already dead.
func (t *traceRecorder) Observe(ctx context.Context, d time.Duration, status int) {
	tc := flightrecorder.TriggerContext{Kind: "http.request", Duration: d}
	if status >= http.StatusInternalServerError {
		tc.Err = errTraceRequestFailed
	}
	t.rec.SnapshotIfAsync(context.WithoutCancel(ctx), tc, t.trigger)
}

// Snapshot implements platform.TraceRecorder. The manual path rides the
// same disk guard as the trigger path: below the floor it refuses (and
// counts the skip) instead of writing.
func (t *traceRecorder) Snapshot(ctx context.Context) (string, error) {
	if !diskHasRoom(t.dir) {
		t.metrics.RecordSkipped()
		return "", wrapf("trace snapshot skipped: free disk below the %d MiB floor", traceDiskFloor>>20)
	}
	path, err := t.rec.SnapshotToDir(ctx)
	if err != nil {
		return "", wrapf("write trace snapshot: %w", err)
	}
	return path, nil
}

// Metrics exposes the snapshot counter family the /metrics exposition
// renders (nil never: the adapter is constructed with it).
func (t *traceRecorder) Metrics() *platform.TraceSnapshotMetrics { return t.metrics }

// Close implements do.Shutdowner: in-flight captures are drained before the
// trace buffer is released.
func (t *traceRecorder) Close() error {
	if err := t.rec.Close(); err != nil {
		return wrapf("close flight recorder: %w", err)
	}
	return nil
}

// traceDir is where snapshots land: a sibling of the database inside the
// service's data directory, so the module's StateDirectory and the backup
// oneshot's UMask cover it without a second path to configure.
func traceDir(dataDir string) string {
	return filepath.Join(dataDir, "traces")
}

// traceRecorderFor builds the recorder: nil when the operator has not
// opted in, which is the zero-cost default.
func traceRecorderFor(cfgDataDir string) (*traceRecorder, error) {
	if !debugFlightRecorderEnabled() {
		return nil, nil
	}
	return newTraceRecorder(traceDir(cfgDataDir))
}

// traceSeam converts the concrete recorder into the server seam without
// the typed-nil trap: a nil *traceRecorder must become a TRUE nil
// interface, or `deps.Trace != nil` would wire middleware over a nil
// receiver (the same class as the SSO nil-client break).
func traceSeam(rec *traceRecorder) platform.TraceRecorder {
	if rec == nil {
		return nil
	}
	return rec
}

// traceMetricsSeam is the same typed-nil discipline for the counter
// family: nil recorder → nil metrics, so /metrics stays byte-identical
// when the debug seam is off.
func traceMetricsSeam(rec *traceRecorder) *platform.TraceSnapshotMetrics {
	if rec == nil {
		return nil
	}
	return rec.Metrics()
}
