package server

import (
	"sync"
	"time"
)

// idemStore remembers recently processed webhook idempotency keys (the
// providers' provider_ref values) so a replayed callback gets the original
// outcome without a second store write. Only successful applies are
// recorded — a failed attempt (unknown ref, transient store error) stays
// retryable by design.
//
// In-memory on purpose: this binary owns the whole store (no
// multi-instance deployment), the session store already makes the same
// process-local trade, and a restart merely re-arms dedupe — the worst
// case is one harmless re-apply of an already-terminal status.
type idemStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]time.Time
}

// hookIdempotencyTTL: provider replays land within minutes; an hour of
// dedupe memory covers every real retry window while keeping the map
// bounded by one hour of callback traffic.
const hookIdempotencyTTL = time.Hour

func newIdemStore(ttl time.Duration) *idemStore {
	return &idemStore{ttl: ttl, entries: make(map[string]time.Time)}
}

// seen reports whether key was already recorded as processed.
func (s *idemStore) seen(key string) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	at, ok := s.entries[key]
	return ok && now.Sub(at) < s.ttl
}

// record marks key as processed (with drop-on-sight of expired entries,
// so the map never grows beyond live traffic).
func (s *idemStore) record(key string) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if at, ok := s.entries[key]; ok && now.Sub(at) >= s.ttl {
		delete(s.entries, key)
	}
	s.entries[key] = now
}
