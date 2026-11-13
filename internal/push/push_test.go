package push_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
)

// fakeSender records sends; enabled toggles its liveness. err = the
// per-send verdict (ErrUnregistered exercises the prune lane).
type fakeSender struct {
	enabled bool
	err     error

	mu     sync.Mutex
	got    []push.Ring
	drafts []push.Draft
	tokens []string
}

func (f *fakeSender) Enabled() bool    { return f.enabled }
func (f *fakeSender) Describe() string { return "fake" }
func (f *fakeSender) Send(_ context.Context, token string, ring push.Ring) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.got = append(f.got, ring)
	f.tokens = append(f.tokens, token)
	return f.err
}
func (f *fakeSender) SendDraft(_ context.Context, token string, draft push.Draft) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.drafts = append(f.drafts, draft)
	f.tokens = append(f.tokens, token)
	return f.err
}
func (f *fakeSender) rings() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.got)
}
func (f *fakeSender) sentDrafts() []push.Draft {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]push.Draft(nil), f.drafts...)
}

// fakeRegistry records prunes; devices is the static inventory the
// List answer flattens.
type fakeRegistry struct {
	devices []push.Device
	err     error

	mu      sync.Mutex
	removed []string
}

func (r *fakeRegistry) List(_ context.Context, _ domain.Extension) ([]push.Device, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.devices, nil
}
func (r *fakeRegistry) Remove(_ context.Context, _ domain.Extension, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removed = append(r.removed, deviceID)
	return nil
}
func (r *fakeRegistry) pruned() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.removed...)
}

// wait spins until cond holds (the fan is detached; its sends are
// synchronous inside the goroutine, so the sender/registry mutations
// become visible once it drains) — the deadline only guards a hang.
func wait(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

var owner = domain.MustParseExtension("1001")

func TestFanRoutesPerTransport(t *testing.T) {
	ntfy := &fakeSender{enabled: true}
	fcm := &fakeSender{enabled: true}
	registry := &fakeRegistry{devices: []push.Device{
		{DeviceID: "d1", Transport: push.TransportNtfy, Token: "topic-a"},
		{DeviceID: "d2", Transport: push.TransportFcm, Token: "reg-b"},
	}}
	fanout := push.NewFanout(map[push.Transport]push.Sender{
		push.TransportNtfy: ntfy,
		push.TransportFcm:  fcm,
	}, registry)
	if !fanout.Enabled() || fanout.Describe() != "fcm+ntfy" {
		t.Fatalf("fanout %q enabled=%v", fanout.Describe(), fanout.Enabled())
	}
	if !fanout.Knows(push.TransportNtfy) || fanout.Knows("apns") {
		t.Fatal("Knows must reflect the live senders")
	}

	ring := push.Ring{Event: push.EventRing, From: "+15550000001", CallID: "c1"}
	if err := fanout.Fan(context.Background(), owner, ring); err != nil {
		t.Fatal(err)
	}
	wait(t, func() bool { return ntfy.rings() == 1 && fcm.rings() == 1 })
	if ntfy.tokens[0] != "topic-a" || fcm.tokens[0] != "reg-b" {
		t.Fatalf("tokens %q / %q, want per-device wake addresses", ntfy.tokens[0], fcm.tokens[0])
	}
}

// TestFanSettleAwaitsTheDetachedFan pins the test seam: Settle returns
// only after the detached fan finished, so settle-then-assert sees the
// send without polling. The ring-hook fixture settles before TempDir
// cleanup (the 2026-11-13 CI flake's hardening).
func TestFanSettleAwaitsTheDetachedFan(t *testing.T) {
	ntfy := &fakeSender{enabled: true}
	fanout := push.NewFanout(map[push.Transport]push.Sender{push.TransportNtfy: ntfy}, &fakeRegistry{devices: []push.Device{
		{DeviceID: "d1", Transport: push.TransportNtfy, Token: "topic-a"},
	}})
	ring := push.Ring{Event: push.EventRing, From: "+15550000001", CallID: "c1"}
	if err := fanout.Fan(context.Background(), owner, ring); err != nil {
		t.Fatal(err)
	}
	if err := fanout.Settle(context.Background()); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if ntfy.rings() != 1 {
		t.Fatalf("rings after settle = %d, want the fan already finished", ntfy.rings())
	}
}

// TestFanRoutesTheDraftFact pins the draft lane's routing: same
// transport table, same per-device tokens — the draft fact is a peer
// of the ring fact, not a special.
func TestFanRoutesTheDraftFact(t *testing.T) {
	ntfy := &fakeSender{enabled: true}
	registry := &fakeRegistry{devices: []push.Device{
		{DeviceID: "d1", Transport: push.TransportNtfy, Token: "topic-a"},
	}}
	fanout := push.NewFanout(map[push.Transport]push.Sender{push.TransportNtfy: ntfy}, registry)
	draft := push.Draft{Event: push.EventDraftReady, Contact: "+15550000001", DraftID: "d-1"}
	if err := fanout.Fan(context.Background(), owner, draft); err != nil {
		t.Fatal(err)
	}
	wait(t, func() bool { return len(ntfy.sentDrafts()) == 1 })
	drafts := ntfy.sentDrafts()
	if drafts[0].DraftID != "d-1" {
		t.Fatalf("drafts %v", drafts)
	}

	// The zero Fanout is the disabled seam: the SYNCHRONOUS contract
	// answers ErrDisabled without detaching anything.
	if err := (&push.Fanout{}).Fan(context.Background(), owner, draft); !errors.Is(err, push.ErrDisabled) {
		t.Fatalf("disabled fanout: %v", err)
	}
}

func TestFanoutDropsDeadSenders(t *testing.T) {
	dead := &fakeSender{enabled: false}
	fanout := push.NewFanout(map[push.Transport]push.Sender{push.TransportNtfy: dead}, nil)
	if fanout.Enabled() {
		t.Fatal("a fanout of dead senders is disabled")
	}
	if err := fanout.Fan(context.Background(), owner, push.Ring{}); !errors.Is(err, push.ErrDisabled) {
		t.Fatalf("error = %v, want ErrDisabled", err)
	}
	var nilFanout *push.Fanout
	if nilFanout.Enabled() || nilFanout.Describe() != "off" {
		t.Fatal("nil fanout must behave as disabled")
	}
}

func TestFanPrunesUnregisteredDevices(t *testing.T) {
	dead := &fakeSender{enabled: true, err: push.ErrUnregistered}
	live := &fakeSender{enabled: true}
	registry := &fakeRegistry{devices: []push.Device{
		{DeviceID: "dead-1", Transport: push.TransportNtfy, Token: "t1"},
		{DeviceID: "live-1", Transport: push.TransportFcm, Token: "t2"},
	}}
	fanout := push.NewFanout(map[push.Transport]push.Sender{
		push.TransportNtfy: dead,
		push.TransportFcm:  live,
	}, registry)
	if err := fanout.Fan(context.Background(), owner, push.Ring{Event: push.EventRing}); err != nil {
		t.Fatal(err)
	}
	wait(t, func() bool { return len(registry.pruned()) == 1 })
	if got := registry.pruned(); len(got) != 1 || got[0] != "dead-1" {
		t.Fatalf("pruned %v, want [dead-1]", got)
	}
	wait(t, func() bool { return live.rings() == 1 })
}

func TestFanSyncErrorsAnswerImmediately(t *testing.T) {
	enabled := push.NewFanout(map[push.Transport]push.Sender{push.TransportNtfy: &fakeSender{enabled: true}}, nil)
	err := enabled.Fan(context.Background(), owner, "not-a-fact")
	if err == nil {
		t.Fatal("an unsupported fact must reject synchronously")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "push.fact")

	err = enabled.Fan(context.Background(), owner, push.Ring{Event: push.EventRing})
	if err == nil {
		t.Fatal("a fan without a registry must reject synchronously")
	}
	errorfamilytest.AssertCode(t, err, "push.registry")
}
