package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// fakeTransport records the rings shipped to it.
type fakeTransport struct {
	enabled bool
	rings   []push.Ring
	err     error
}

func (f *fakeTransport) Enabled() bool    { return f.enabled }
func (f *fakeTransport) Describe() string { return "fake" }
func (f *fakeTransport) Send(_ context.Context, _ string, ring push.Ring) error {
	f.rings = append(f.rings, ring)
	return f.err
}
func (f *fakeTransport) SendDraft(_ context.Context, _ string, draft push.Draft) error {
	f.rings = append(f.rings, push.Ring{Event: draft.Event, From: draft.Contact})
	return f.err
}

func newRingHandlers(t *testing.T, transport *fakeTransport) (*Handlers, *readmodels.PushDevices) {
	t.Helper()
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	devices := engine.PushDevices()
	fanout := push.NewFanout(map[string]push.Sender{push.TransportNtfy: transport})
	handlers := New(Deps{
		Secret:  "hook-secret",
		Push:    fanout,
		Devices: devices,
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
	if err := devices.Upsert(context.Background(), owner, "device-aaaa", push.TransportNtfy, "wp-topic"); err != nil {
		t.Fatal(err)
	}

	rec := postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-1","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202 (body %q)", rec.Code, rec.Body.String())
	}
	if len(transport.rings) != 1 {
		t.Fatalf("transport rings %d, want 1", len(transport.rings))
	}
	got := transport.rings[0]
	if got.Event != push.EventRing || got.From != "+15550000001" || got.CallID != "fs-1" {
		t.Fatalf("ring = %+v", got)
	}

	// ring_end stands the device down.
	rec = postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-1","state":"end"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("end status %d", rec.Code)
	}
	if len(transport.rings) != 2 || transport.rings[1].Event != push.EventRingEnd {
		t.Fatalf("rings after end = %+v", transport.rings)
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
	if len(transport.rings) != 0 {
		t.Fatalf("no ring may ship from a rejected payload, got %+v", transport.rings)
	}

	// No devices registered: inert 202, nothing shipped.
	rec := postRing(t, handlers, `{"owner":"1099","from":"+15550000001","call_id":"c","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("no-devices status %d, want 202", rec.Code)
	}
	if len(transport.rings) != 0 {
		t.Fatalf("no-devices shipped %+v", transport.rings)
	}
}

func TestHookRingPrunesDeadDevices(t *testing.T) {
	// The transport reports the device dead: the registration is pruned
	// and the hook still answers 202.
	transport := &fakeTransport{enabled: true, err: push.ErrUnregistered}
	handlers, devices := newRingHandlers(t, transport)
	owner := domain.MustParseExtension("1001")
	if err := devices.Upsert(context.Background(), owner, "device-aaaa", push.TransportNtfy, "wp-topic"); err != nil {
		t.Fatal(err)
	}

	rec := postRing(t, handlers, `{"owner":"1001","from":"+15550000001","call_id":"fs-2","state":"ring"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202", rec.Code)
	}
	views, err := devices.List(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
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
