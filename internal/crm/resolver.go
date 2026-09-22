package crm

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Cache tuning: names change rarely (contact edits), and a phone page holds
// tens of numbers, so the positive window outlives a typical browsing
// session. Misses are retried sooner: a contact created "just now" in the
// CRM should appear on the next tab visit, not tomorrow.
const (
	positiveTTL = 6 * time.Hour
	negativeTTL = 5 * time.Minute
	maxEntries  = 1024
)

// cacheEntry is one resolved number. A miss stores a zero Match with
// miss=true so both states share the TTL mechanics.
type cacheEntry struct {
	match   Match
	miss    bool
	fetched time.Time
}

// lookupCounters counts UPSTREAM lookups by outcome (one increment per
// actual CRM round-trip — cache hits are not CRM traffic). Rendered as
// aggregate /metrics; never per-extension data.
type lookupCounters struct {
	hit     atomic.Int64
	miss    atomic.Int64
	failure atomic.Int64
}

// lookupFlight is one in-flight upstream lookup. Concurrent Resolves for
// the same number join the leader's round-trip instead of fanning out one
// request per caller (a history page rendering while another tab renders
// must cost the CRM one lookup, not two).
type lookupFlight struct {
	done  chan struct{}
	match Match
	ok    bool
}

// Resolver answers "who is this number?" with a process-local TTL cache in
// front of the CRM client. Every failure — CRM down, slow, unauthorized —
// degrades to "no match": enrichment is decoration, never a prerequisite,
// and the failure map keeps it out of the user's way (the raw number
// renders; the operator sees one debug log per failed attempt).
type Resolver struct {
	client *Client
	log    *slog.Logger

	mu       sync.Mutex
	cache    map[string]cacheEntry
	inflight map[string]*lookupFlight

	lookups lookupCounters
}

// NewResolver wraps a client with the lookup cache. A disabled client
// yields a resolver that never matches without touching the map.
func NewResolver(client *Client, log *slog.Logger) *Resolver {
	if log == nil {
		log = slog.Default()
	}

	return &Resolver{
		client:   client,
		log:      log,
		cache:    make(map[string]cacheEntry),
		inflight: make(map[string]*lookupFlight),
	}
}

// Enabled reports whether the underlying CRM client is wired up.
func (r *Resolver) Enabled() bool { return r != nil && r.client.Enabled() }

// LookupCounters snapshots the upstream-lookup outcome counters. Nil-safe.
func (r *Resolver) LookupCounters() (hit, miss, failure int64) {
	if r == nil {
		return 0, 0, 0
	}

	return r.lookups.hit.Load(), r.lookups.miss.Load(), r.lookups.failure.Load()
}

// Resolve returns the contact a number belongs to, consulting the TTL cache
// first. ok=false when the CRM is disabled, holds no contact, or fails —
// callers treat all three identically: render the raw number / skip the
// logging. Numbers are cache keys verbatim; distinct spellings of one
// number cost at most distinct entries until the CRM's canonical answer
// dominates the window.
func (r *Resolver) Resolve(ctx context.Context, number string) (Match, bool) {
	if number == "" || !r.Enabled() {
		return Match{}, false
	}

	if entry, fresh := r.cached(number); fresh {
		return entry.match, !entry.miss
	}

	if flight := r.join(number); flight != nil {
		select {
		case <-flight.done:
			return flight.match, flight.ok
		case <-ctx.Done():
			// The waiting caller gave up; the leader's answer still
			// lands in the cache for everyone after.
			return Match{}, false
		}
	}

	match, err := r.client.LookupByPhone(ctx, number)

	switch {
	case err == nil:
		r.remember(number, cacheEntry{match: match, fetched: time.Now()})
		r.settle(number, match, true)
		r.lookups.hit.Add(1)

		return match, true
	case errIsMiss(err):
		r.remember(number, cacheEntry{miss: true, fetched: time.Now()})
		r.settle(number, Match{}, false)
		r.lookups.miss.Add(1)

		return Match{}, false
	default:
		// Do not cache transport failures: the next render retries, and
		// one debug line per attempt keeps the operator trail honest.
		r.log.Debug("crm: lookup failed; rendering raw number", "error", err)
		r.settle(number, Match{}, false)
		r.lookups.failure.Add(1)

		return Match{}, false
	}
}

// Name resolves the display name for a number, or "" on any miss.
func (r *Resolver) Name(ctx context.Context, number string) string {
	match, ok := r.Resolve(ctx, number)
	if !ok {
		return ""
	}

	return match.Name
}

// Names resolves a batch of numbers and returns the number→name map for
// view props. Unresolved numbers are absent, not empty-stringed, so the
// displayName fallback needs no extra policy.
func (r *Resolver) Names(ctx context.Context, numbers []string) map[string]string {
	names := make(map[string]string, len(numbers))

	for _, number := range numbers {
		if name := r.Name(ctx, number); name != "" {
			names[number] = name
		}
	}

	return names
}

// LogCall forwards one call activity to the CRM (no caching — a call is a
// fact, not a lookup).
func (r *Resolver) LogCall(ctx context.Context, contactID, direction, number string, seconds int, outcome string) error {
	return r.client.LogCall(ctx, contactID, direction, number, seconds, outcome)
}

func errIsMiss(err error) bool { return err == ErrNotFound || err == ErrDisabled }

// join registers the caller as the lookup's leader (nil return) or hands
// back the existing flight to wait on.
func (r *Resolver) join(number string) *lookupFlight {
	r.mu.Lock()
	defer r.mu.Unlock()

	if flight, ok := r.inflight[number]; ok {
		return flight
	}

	r.inflight[number] = &lookupFlight{done: make(chan struct{})}

	return nil
}

// settle publishes the leader's result to joiners and retires the
// in-flight entry.
func (r *Resolver) settle(number string, match Match, ok bool) {
	r.mu.Lock()
	flight := r.inflight[number]
	delete(r.inflight, number)
	r.mu.Unlock()

	if flight != nil {
		flight.match = match
		flight.ok = ok
		close(flight.done)
	}
}

func (r *Resolver) cached(number string) (cacheEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.cache[number]
	if !ok {
		return cacheEntry{}, false
	}

	ttl := positiveTTL

	if entry.miss {
		ttl = negativeTTL
	}

	if time.Since(entry.fetched) > ttl {
		delete(r.cache, number)

		return cacheEntry{}, false
	}

	return entry, true
}

func (r *Resolver) remember(number string, entry cacheEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Bound the map: real deployments hold hundreds of numbers; on the
	// (unrealistic) overflow, dropping the whole cache is simpler and
	// cheaper than tracking LRU order for a decoration cache.
	if len(r.cache) >= maxEntries {
		r.cache = make(map[string]cacheEntry, maxEntries)
	}

	r.cache[number] = entry
}
