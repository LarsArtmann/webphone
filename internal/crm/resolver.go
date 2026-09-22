package crm

import (
	"context"
	"log/slog"
	"sync"
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

// cacheEntry is one resolved (or definitively missing) number.
type cacheEntry struct {
	name    string // empty = known miss
	fetched time.Time
}

// Resolver answers "who is this number?" with a process-local TTL cache in
// front of the CRM client. Every failure — CRM down, slow, unauthorized —
// degrades to an empty name: enrichment is decoration, never a
// prerequisite, and the failure map keeps it out of the user's way (the
// raw number renders; the operator sees one debug log per miss window).
type Resolver struct {
	client *Client
	log    *slog.Logger

	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewResolver wraps a client with the lookup cache. A disabled client
// yields a resolver whose Name always returns "" without touching the map.
func NewResolver(client *Client, log *slog.Logger) *Resolver {
	if log == nil {
		log = slog.Default()
	}

	return &Resolver{client: client, log: log, cache: make(map[string]cacheEntry)}
}

// Enabled reports whether the underlying CRM client is wired up.
func (r *Resolver) Enabled() bool { return r != nil && r.client.Enabled() }

// Name resolves the display name for a number, or "" when the CRM is
// disabled, holds no contact, or fails. Numbers are the cache keys verbatim;
// distinct spellings of one number cost at most distinct entries until the
// CRM's canonical answer dominates the window.
func (r *Resolver) Name(ctx context.Context, number string) string {
	if number == "" || !r.Enabled() {
		return ""
	}

	if name, ok := r.cached(number); ok {
		return name
	}

	match, err := r.client.LookupByPhone(ctx, number)

	switch {
	case err == nil:
		r.remember(number, match.Name, positiveTTL)

		return match.Name
	case errIsMiss(err):
		r.remember(number, "", negativeTTL)

		return ""
	default:
		// Do not cache transport failures: the next render retries, and
		// one debug line per attempt keeps the operator trail honest.
		r.log.Debug("crm: lookup failed; rendering raw number", "error", err)

		return ""
	}
}

// Names resolves a batch of numbers in one call and returns the number→name
// map for view props. Unresolved numbers are absent, not empty-stringed, so
// templates distinguish "no CRM" from "no match" by lookup, not semantics.
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

func (r *Resolver) cached(number string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.cache[number]
	if !ok {
		return "", false
	}

	ttl := positiveTTL

	if entry.name == "" {
		ttl = negativeTTL
	}

	if time.Since(entry.fetched) > ttl {
		delete(r.cache, number)

		return "", false
	}

	return entry.name, true
}

func (r *Resolver) remember(number, name string, ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Bound the map: real deployments hold hundreds of numbers; on the
	// (unrealistic) overflow, dropping the whole cache is simpler and
	// cheaper than tracking LRU order for a decoration cache.
	if len(r.cache) >= maxEntries {
		r.cache = make(map[string]cacheEntry, maxEntries)
	}

	r.cache[number] = cacheEntry{name: name, fetched: time.Now()}
}
