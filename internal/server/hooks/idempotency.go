package hooks

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DedupWindow is the engine-backed dedup seam: a check-and-set window
// with per-key TTL (the metaengine DedupStore port, consumer-owned so
// this package stays library-free). Nil (test compositions, map
// fallback) degrades to the in-process idemStore.
//
// branching-flow IFACE_COMPLETE (baselined): consumer-side seam kept
// small on purpose (map fallback in tests, engine window in prod) —
// the "one implementer" reading misses the nil seam.
type DedupWindow interface {
	// Seen reports whether key is currently recorded and unexpired.
	Seen(ctx context.Context, collection, key string, now time.Time) (bool, error)
	// CheckAndRecord atomically records key (expiring at now+ttl) when
	// absent; returns true when the key was already seen.
	CheckAndRecord(ctx context.Context, collection, key string, ttl time.Duration, now time.Time) (bool, error)
	// Sweep deletes every expired record in the collection and returns
	// how many.
	Sweep(ctx context.Context, collection string, now time.Time) (int, error)
}

// DedupCollection is the single engine collection both idem gates
// share: TTL is carried per key (a DedupWindow property), so the hook
// and call windows can diverge inside one sweepable collection.
const DedupCollection = "webphone.idem"

// Gate is the check-then-record contract the webhook and call lanes
// consume: check first (only successes are recorded), then the caller
// records after the guarded work succeeded. Both impls honor it — the
// map (rollback seam, tests) and the engine window.
type Gate interface {
	Seen(key string) bool
	Record(key string)
}

// idemStore remembers recently processed webhook idempotency keys (the
// providers' provider_ref values) so a replayed callback gets the original
// outcome without a second store write. Only successful applies are
// recorded — a failed attempt (unknown ref, transient store error) stays
// retryable by design.
//
// In-memory fallback (and the rollback seam for the engine window):
// a restart merely re-arms dedupe — the worst case is one harmless
// re-apply of an already-terminal status.
// branching-flow COMPOSITION_mixin (baselined): coincidental mu+ttl
// shape; hook dedupe and the unread cache are different domains.
type idemStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]time.Time
}

// hookIdempotencyTTL: provider replays land within minutes; an hour of
// dedupe memory covers every real retry window while keeping the map
// bounded by one hour of callback traffic.
const hookIdempotencyTTL = time.Hour

// CallIdempotencyTTL bounds the island's post-call report dedupe
// window: a replay is a browser-level retry or a double-fire, which
// lands within seconds of the first attempt. The own constant keeps
// the journal-dedupe window free to diverge from the webhook one.
const CallIdempotencyTTL = time.Hour

func newIdemStore(ttl time.Duration) *idemStore {
	return &idemStore{ttl: ttl, entries: make(map[string]time.Time)}
}

// Seen and Record repeat the clock+lock prologue deliberately: the
// fail-retryable contract needs check-then-record as separate steps,
// and a lock helper would sever Lock from its deferred Unlock.
// Seen reports whether key was already recorded as processed.
func (s *idemStore) Seen(key string) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	at, ok := s.entries[key]
	return ok && now.Sub(at) < s.ttl
}

// Record marks key as processed (with drop-on-sight of expired entries,
// so the map never grows beyond live traffic).
func (s *idemStore) Record(key string) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if at, ok := s.entries[key]; ok && now.Sub(at) >= s.ttl {
		delete(s.entries, key)
	}
	s.entries[key] = now
}

// engineIdemGate adapts a DedupWindow into the Gate contract: the
// durable engine-backed implementation. Unlike the map, it survives
// restarts — a restart inside the TTL window no longer forgets seen
// hooks (the defect the swap fixes).
//
// Error policy is fail-open on purpose, on both legs: a dedup window
// that cannot be read must not block a callback whose guarded work is
// idempotent at rest anyway (treat as unseen → the work re-runs → the
// store's own terminal-state handling absorbs the replay), and a
// record that cannot be written leaves the key retryable — the exact
// contract a failed attempt already has. Both log once per occurrence.
//
// Calls run on context.Background: the dedup record is persistence
// that should outlive the request it guards, and the underlying ops
// are bounded single-writer SQLite statements.
type engineIdemGate struct {
	window DedupWindow
	ttl    time.Duration
}

// NewGate picks the implementation behind the constructor seam: the
// engine window when the composition root wired one, the in-process
// map otherwise (test compositions, rollback).
func NewGate(window DedupWindow, ttl time.Duration) Gate {
	if window == nil {
		return newIdemStore(ttl)
	}
	return &engineIdemGate{window: window, ttl: ttl}
}

func (g *engineIdemGate) Seen(key string) bool {
	seen, err := g.window.Seen(context.Background(), DedupCollection, key, time.Now())
	if err != nil {
		slog.Warn("idem gate: dedup window read failed; treating key as unseen", "key", key, "error", err)
		return false
	}
	return seen
}

func (g *engineIdemGate) Record(key string) {
	if _, err := g.window.CheckAndRecord(context.Background(), DedupCollection, key, g.ttl, time.Now()); err != nil {
		slog.Warn("idem gate: dedup window write failed; key stays retryable", "key", key, "error", err)
	}
}
