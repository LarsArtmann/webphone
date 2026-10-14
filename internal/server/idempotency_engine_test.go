package server

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeDedupWindow records the calls the engine gate makes.
type fakeDedupWindow struct {
	seenKeys        []string
	recordKeys      []string
	seenResult      bool
	seenErr         error
	recordResult    bool
	recordErr       error
	seenCollection  string
	recordTTL       time.Duration
	recordTimestamp time.Time
}

func (f *fakeDedupWindow) Seen(_ context.Context, collection, key string, _ time.Time) (bool, error) {
	f.seenCollection = collection
	f.seenKeys = append(f.seenKeys, key)
	return f.seenResult, f.seenErr
}

func (f *fakeDedupWindow) CheckAndRecord(
	_ context.Context, collection, key string, ttl time.Duration, now time.Time,
) (bool, error) {
	_ = collection
	f.recordKeys = append(f.recordKeys, key)
	f.recordTTL = ttl
	f.recordTimestamp = now
	return f.recordResult, f.recordErr
}

func (f *fakeDedupWindow) Sweep(_ context.Context, _ string, _ time.Time) (int, error) {
	return 0, nil
}

// TestNewIdemGateFallsBackToMap pins the constructor seam: a nil
// window must produce the in-process map (byte-identical legacy
// behavior for test compositions and the documented rollback).
func TestNewIdemGateFallsBackToMap(t *testing.T) {
	gate := newIdemGate(nil, time.Minute)
	if _, ok := gate.(*idemStore); !ok {
		t.Fatalf("nil window produced %T, want *idemStore", gate)
	}
	if gate.seen("k") {
		t.Fatal("fresh map gate reports seen")
	}
	gate.record("k")
	if !gate.seen("k") {
		t.Fatal("map gate forgot a recorded key")
	}
}

func TestEngineIdemGateRoutesToWindow(t *testing.T) {
	window := &fakeDedupWindow{}
	gate := newIdemGate(window, hookIdempotencyTTL)

	if gate.seen("ref-1") {
		t.Fatal("unseen key reported seen")
	}
	if len(window.seenKeys) != 1 || window.seenKeys[0] != "ref-1" {
		t.Fatalf("Seen not routed: %v", window.seenKeys)
	}
	if window.seenCollection != DedupCollection {
		t.Fatalf("Seen collection %q, want %q", window.seenCollection, DedupCollection)
	}

	gate.record("ref-1")
	if len(window.recordKeys) != 1 || window.recordKeys[0] != "ref-1" {
		t.Fatalf("CheckAndRecord not routed: %v", window.recordKeys)
	}
	if window.recordTTL != hookIdempotencyTTL {
		t.Fatalf("record TTL %v, want %v", window.recordTTL, hookIdempotencyTTL)
	}
	if window.recordTimestamp.IsZero() {
		t.Fatal("record ran without a timestamp")
	}
}

// TestEngineIdemGateFailsOpen pins the error policy: an unreadable or
// unwritable dedup window must never block the guarded lane — a failed
// read treats the key as unseen (work re-runs, idempotent at rest), a
// failed write leaves the key retryable.
func TestEngineIdemGateFailsOpen(t *testing.T) {
	window := &fakeDedupWindow{seenErr: errors.New("boom"), recordErr: errors.New("boom")}
	gate := newIdemGate(window, time.Minute)

	if gate.seen("ref-2") {
		t.Fatal("failed read must treat the key as unseen")
	}
	gate.record("ref-2") // must not panic
}

func TestEngineIdemGateHonorsWindowVerdict(t *testing.T) {
	window := &fakeDedupWindow{seenResult: true}
	gate := newIdemGate(window, time.Minute)
	if !gate.seen("ref-3") {
		t.Fatal("window's seen verdict not propagated")
	}
}
