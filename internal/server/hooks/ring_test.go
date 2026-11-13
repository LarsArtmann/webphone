package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// fakeTransport records the rings shipped to it. Mutex-guarded: the
// fanout delivers from a detached goroutine, so the test's polling
// reads race with the sends.
type fakeTransport struct {
	enabled bool
	err     error

	mu    sync.Mutex
	rings []push.Ring
}

func (f *fakeTransport) Enabled() bool    { return f.enabled }
func (f *fakeTransport) Describe() string { return "fake" }
func (f *fakeTransport) Send(_ context.Context, _ string, ring push.Ring) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rings = append(f.rings, ring)
	return f.err
}
func (f *fakeTransport) SendDraft(_ context.Context, _ string, draft push.Draft) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rings = append(f.rings, push.Ring{Event: draft.Event, From: draft.Contact})
	return f.err
}
func (f *fakeTransport) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.rings)
}
func (f *fakeTransport) snapshot() []push.Ring {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]push.Ring(nil), f.rings...)
}

// fanRegistry adapts the journal-backed device store to the fanout's
// prune seam (the same conversion the composition root performs in
// prod; push.DeviceRegistry's List flattens the view).
type fanRegistry struct{ devices readmodels.PushDeviceGate }

func (r fanRegistry) List(ctx context.Context, owner domain.Extension) ([]push.Device, error) {
	views, err := r.devices.List(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]push.Device, 0, len(views))
	for _, v := range views {
		out = append(out, push.Device{DeviceID: v.DeviceID, Transport: push.Transport(v.Transport), Token: v.Token})
	}
	return out, nil
}

func (r fanRegistry) Remove(ctx context.Context, owner domain.Extension, deviceID string) error {
	return r.devices.Remove(ctx, owner, deviceID)
}

// newRingCoreSystem boots the throwaway journal system (every gate
// decl registered) — hooks cannot reuse package server's newCoreSystem
// (different package; server imports hooks, never the reverse).
func newRingCoreSystem(t *testing.T) *system.System {
	t.Helper()
	sys, err := cqrs.NewCoreSystem(context.Background(), "file:"+filepath.Join(t.TempDir(), "system.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sys.Close() })
	return sys
}

func newRingHandlers(t *testing.T, transport *fakeTransport) (*Handlers, readmodels.PushDeviceGate) {
	t.Helper()
	devices := cqrs.NewPushDevices(newRingCoreSystem(t))
	fanout := push.NewFanout(map[push.Transport]push.Sender{push.TransportNtfy: transport}, fanRegistry{devices: devices})
	handlers := New(Deps{
		Secret: "hook-secret",
		Push:   fanout,
	})
	// The hooks detach their fans; settle them before TempDir's cleanup
	// removes the journal's directory (cleanup order is LIFO: this
	// registers after newRingCoreSystem's TempDir). The 2026-11-13 CI
	// flake was the detached fan's store read racing that removal.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := fanout.Settle(ctx); err != nil {
			t.Logf("ring fans did not settle: %v", err)
		}
	})
	return handlers, devices
}

func postRing(t *testing.T, h *Handlers, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/hooks/ring", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Webhooks(rec, req)
	return rec
}

func TestHookRingFansOutToRegisteredDevices(t *testing.T) {
	transport := &fakeTransport{enabled: true}
	handlers, devices := newRingHandlers(t, transport)
	owner := domain.MustParseExtension("1001")
	if err := devices.Upsert(context.Background(), owner, "device-aaaa", string(push.TransportNtfy), "wp-topic"); err != nil {
		t.Fatal(err)
	}

	rec := postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-1","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202 (body %q)", rec.Code, rec.Body.String())
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && transport.count() < 1 {
		time.Sleep(5 * time.Millisecond)
	}
	if transport.count() != 1 {
		t.Fatalf("transport rings %d, want 1", transport.count())
	}
	got := transport.snapshot()[0]
	if got.Event != push.EventRing || got.From != "+15550000001" || got.CallID != "fs-1" {
		t.Fatalf("ring = %+v", got)
	}

	// ring_end stands the device down.
	rec = postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-1","state":"end"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("end status %d", rec.Code)
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && transport.count() < 2 {
		time.Sleep(5 * time.Millisecond)
	}
	if transport.count() != 2 || transport.snapshot()[1].Event != push.EventRingEnd {
		t.Fatalf("rings after end = %+v", transport.snapshot())
	}
}

func TestHookRingValidatesAndStaysInert(t *testing.T) {
	transport := &fakeTransport{enabled: true}
	handlers, _ := newRingHandlers(t, transport)

	for _, tc := range []struct {
		name string
		body string
		code int
	}{
		{"bad state", `{"owner":"1001","from":"+15550000001","call_id":"c","state":"maybe"}`, http.StatusBadRequest},
		{"empty owner", `{"owner":"","from":"+15550000001","call_id":"c","state":"ring"}`, http.StatusBadRequest},
		{"empty from", `{"owner":"1001","from":"","call_id":"c","state":"ring"}`, http.StatusBadRequest},
		{"garbage", `not json`, http.StatusBadRequest},
	} {
		rec := postRing(t, handlers, tc.body)
		if rec.Code != tc.code {
			t.Fatalf("%s: status %d, want %d", tc.name, rec.Code, tc.code)
		}
	}
	if transport.count() != 0 {
		t.Fatalf("no ring may ship from a rejected payload, got %+v", transport.snapshot())
	}

	// No devices registered: inert 202, nothing shipped.
	rec := postRing(t, handlers, `{"owner":"1099","from":"+15550000001","call_id":"c","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("no-devices status %d, want 202", rec.Code)
	}
	if transport.count() != 0 {
		t.Fatalf("no-devices shipped %+v", transport.snapshot())
	}
}

func TestHookRingPrunesDeadDevices(t *testing.T) {
	// The transport reports the device dead: the registration is pruned
	// and the hook still answers 202.
	transport := &fakeTransport{enabled: true, err: push.ErrUnregistered}
	handlers, devices := newRingHandlers(t, transport)
	owner := domain.MustParseExtension("1001")
	if err := devices.Upsert(context.Background(), owner, "device-aaaa", string(push.TransportNtfy), "wp-topic"); err != nil {
		t.Fatal(err)
	}

	rec := postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-2","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202", rec.Code)
	}
	// The fan is detached: the prune happens in its goroutine, so poll
	// the store instead of asserting synchronously.
	deadline := time.Now().Add(2 * time.Second)
	var views []readmodels.PushDeviceView
	var err error
	for time.Now().Before(deadline) {
		views, err = devices.List(context.Background(), owner)
		if err != nil {
			t.Fatal(err)
		}
		if len(views) == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(views) != 0 {
		t.Fatalf("dead device not pruned: %+v", views)
	}
}

func TestHookRingDisabledSeamAcknowledgesInertly(t *testing.T) {
	// No push seam wired: the lane still exists for the dialplan and
	// acknowledges inertly.
	handlers := New(Deps{Secret: "hook-secret"})
	rec := postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"c","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202", rec.Code)
	}
}
