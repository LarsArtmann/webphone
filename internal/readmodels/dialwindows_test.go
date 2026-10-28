package readmodels_test

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// TestDialWindowsGate pins the gate's check+record semantics (issue
// #7): the first dial records and allows, an immediate re-dial inside
// the window denies WITHOUT recording (the window must not grow by
// retrying), and after expiry the dial records and allows again.
func TestDialWindowsGate(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	dials := engine.DialWindows()
	ctx := context.Background()
	owner := mustExtension(t, "1001")
	cooldown := 50 * time.Millisecond

	allowed, _, err := dials.Attempt(ctx, owner, "+4989123456", cooldown)
	if err != nil || !allowed {
		t.Fatalf("first dial: %v %v", allowed, err)
	}
	denied, stamp, err := dials.Attempt(ctx, owner, "+4989123456", cooldown)
	if err != nil || denied {
		t.Fatalf("immediate re-dial must deny: %v %v", denied, err)
	}
	if stamp.IsZero() {
		t.Fatal("a denied ask carries the slot's last-at stamp")
	}
	time.Sleep(60 * time.Millisecond)
	allowed, _, err = dials.Attempt(ctx, owner, "+4989123456", cooldown)
	if err != nil || !allowed {
		t.Fatalf("post-expiry dial must allow: %v %v", allowed, err)
	}

	// Owner scoping: another extension's window is a different slot.
	other := mustExtension(t, "2002")
	if allowed, _, err := dials.Attempt(ctx, other, "+4989123456", cooldown); err != nil || !allowed {
		t.Fatalf("another owner's first dial: %v %v", allowed, err)
	}
}

// TestDialWindowsOverrideOneShot pins the bypass: arming spends on the
// NEXT attempt only — the follow-up attempt (window still hot) denies
// again, and the override never refreshes itself.
func TestDialWindowsOverrideOneShot(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	dials := engine.DialWindows()
	ctx := context.Background()
	owner := mustExtension(t, "1001")
	cooldown := time.Hour

	if _, _, err := dials.Attempt(ctx, owner, "+4989", cooldown); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if allowed, _, _ := dials.Attempt(ctx, owner, "+4989", cooldown); allowed {
		t.Fatal("the hot window must deny before the override")
	}
	if err := dials.ArmOverride(ctx, owner, "+4989"); err != nil {
		t.Fatalf("arm: %v", err)
	}
	if allowed, _, err := dials.Attempt(ctx, owner, "+4989", cooldown); err != nil || !allowed {
		t.Fatalf("armed override must allow the next attempt: %v %v", allowed, err)
	}
	if allowed, _, _ := dials.Attempt(ctx, owner, "+4989", cooldown); allowed {
		t.Fatal("the override is ONE-SHOT: the follow-up must deny again")
	}
}

// TestDialWindowsRestartSurvival pins the durable window: the slot
// survives a store close/reopen (the fleet restarts must not reset a
// hot cooldown).
func TestDialWindowsRestartSurvival(t *testing.T) {
	dir := t.TempDir()
	engine, err := readmodels.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	owner := mustExtension(t, "1001")
	if _, _, err := engine.DialWindows().Attempt(context.Background(), owner, "+4989", time.Hour); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := engine.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	reopened, err := readmodels.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Shutdown() })
	allowed, _, err := reopened.DialWindows().Attempt(context.Background(), owner, "+4989", time.Hour)
	if err != nil || allowed {
		t.Fatalf("the window must survive the restart: %v %v", allowed, err)
	}
}
