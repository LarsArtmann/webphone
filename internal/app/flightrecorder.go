package app

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	flightrecorder "github.com/larsartmann/go-flightrecorder"

	"github.com/larsartmann/webphone/internal/server"
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

// traceRecorder adapts go-flightrecorder to the server's TraceRecorder
// seam. It implements do.Shutdowner so the container drains in-flight
// captures on exit (the library blocks Close until they finish).
type traceRecorder struct {
	rec     *flightrecorder.Recorder
	trigger flightrecorder.TriggerFunc
}

// newTraceRecorder starts the recorder writing snapshots into dir. A
// failure here is a boot failure, not a silent degradation: the operator
// asked for tracing and must learn immediately that it is not running.
func newTraceRecorder(dir string) (*traceRecorder, error) {
	rec, err := flightrecorder.New(
		flightrecorder.WithSnapshotDir(dir),
		flightrecorder.WithMaxSnapshots(traceRetention),
		flightrecorder.WithCompression(gzip.BestSpeed),
		flightrecorder.WithMinAge(traceMinAge),
	)
	if err != nil {
		return nil, wrapf("configure flight recorder: %w", err)
	}
	if err := rec.Start(); err != nil {
		return nil, wrapf("start flight recorder: %w", err)
	}
	return &traceRecorder{
		rec:     rec,
		trigger: flightrecorder.OnErrorOrLatency(traceSlowRequest),
	}, nil
}

// Observe implements server.TraceRecorder. The request context is detached
// from cancellation: the snapshot write outlives the response, and the
// library skips a capture whose context is already dead.
func (t *traceRecorder) Observe(ctx context.Context, d time.Duration, status int) {
	tc := flightrecorder.TriggerContext{Kind: "http.request", Duration: d}
	if status >= http.StatusInternalServerError {
		tc.Err = errTraceRequestFailed
	}
	t.rec.SnapshotIfAsync(context.WithoutCancel(ctx), tc, t.trigger)
}

// Snapshot implements server.TraceRecorder.
func (t *traceRecorder) Snapshot(ctx context.Context) (string, error) {
	path, err := t.rec.SnapshotToDir(ctx)
	if err != nil {
		return "", wrapf("write trace snapshot: %w", err)
	}
	return path, nil
}

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
func traceSeam(rec *traceRecorder) server.TraceRecorder {
	if rec == nil {
		return nil
	}
	return rec
}
