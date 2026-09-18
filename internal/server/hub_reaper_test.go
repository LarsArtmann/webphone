package server

import (
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestHubReaperDeletesOnlyIdleHubs pins the reaper contract: a hub with
// no live subscribers that nobody touched for the TTL is deleted (fresh
// one created on demand), while a hub with a live subscriber is never
// reaped no matter how old its seen stamp is — and a broadcast after a
// reap lands on the fresh hub without panic.
func TestHubReaperDeletesOnlyIdleHubs(t *testing.T) {
	ext := domain.MustParseExtension("2001")
	ext2 := domain.MustParseExtension("2002")
	hubs := NewHubs()

	idle := hubs.get(ext)
	busy := hubs.get(ext2)
	sub := busy.Hub().Subscribe()
	t.Cleanup(func() { busy.Hub().Unsubscribe(sub) })

	// Age both hubs past the TTL and sweep explicitly (the sweep also
	// runs opportunistically inside get()'s write path in production).
	stale := time.Now().Add(-hubIdleTTL - time.Second)
	hubs.mu.Lock()
	hubs.seen[ext.String()] = stale
	hubs.seen[ext2.String()] = stale
	hubs.sweep(time.Now())
	hubs.mu.Unlock()

	if hubs.get(ext) == idle {
		t.Fatal("idle hub survived the sweep; a fresh hub was expected")
	}

	hubs.mu.Lock()
	_, busyAlive := hubs.hubs[ext2.String()]
	hubs.mu.Unlock()
	if !busyAlive {
		t.Fatal("hub with a live subscriber was reaped — active hubs must never be reaped")
	}

	// Broadcast after a reap: no panic, and the fresh hub receives it.
	events := hubs.get(ext).Hub().Subscribe()
	t.Cleanup(func() { hubs.get(ext).Hub().Unsubscribe(events) })
	hubs.Publish(ext, sseEventThreads, "<div>after-reap</div>")
	select {
	case ev := <-events:
		if ev.Event != sseEventThreads {
			t.Errorf("post-reap broadcast event %q", ev.Event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("post-reap broadcast never reached the fresh hub")
	}
}

// TestHubReaperKeepsFreshIdleHubs: a just-created hub (no subscribers
// yet — the connect window) is not reaped by the sweep that created it.
func TestHubReaperKeepsFreshIdleHubs(t *testing.T) {
	ext := domain.MustParseExtension("2003")
	hubs := NewHubs()

	hub := hubs.get(ext)

	hubs.mu.Lock()
	hubs.sweep(time.Now()) // just-created hub: seen stamp is brand new
	hubs.mu.Unlock()

	hubs.mu.Lock()
	stillThere := hubs.hubs[ext.String()] == hub
	hubs.mu.Unlock()
	if !stillThere {
		t.Fatal("freshly created hub was reaped — the connect window must be protected")
	}
}
