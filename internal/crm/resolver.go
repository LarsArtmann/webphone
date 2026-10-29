package crm

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
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

// cacheEntry is one resolved number. A miss stores a zero Person with
// miss=true so both states share the TTL mechanics.
type cacheEntry struct {
	person  Person
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
	done   chan struct{}
	person Person
	ok     bool
}

// Resolver answers "who is this number?" with a process-local TTL cache in
// front of the CRM client. Every failure — CRM down, slow, unauthorized —
// degrades to "no match": enrichment is decoration, never a prerequisite,
// and the failure map keeps it out of the user's way (the raw number
// renders; the operator sees one debug log per failed attempt).
type Resolver struct {
	client *Client
	// profileURL is the deployment's deep-link template into the CRM's own
	// UI ({id} = the CRM's contact id). Empty = no links, the zero-config
	// product. Both facts a link needs — the template (config) and the id
	// (CRM wire data) — are known only here, so the link is built here and
	// the views merely render it.
	profileURL string
	log        *slog.Logger

	mu       sync.Mutex
	cache    map[string]cacheEntry
	inflight map[string]*lookupFlight

	lookups lookupCounters
}

// NewResolver wraps a client with the lookup cache and the optional
// deep-link template. A disabled client yields a resolver that never
// matches without touching the map. profileURL must already be validated
// by config (absolute http(s), contains {id}); an empty one means "this
// deployment renders names without links".
func NewResolver(client *Client, profileURL string, log *slog.Logger) *Resolver {
	if log == nil {
		log = slog.Default()
	}

	return &Resolver{
		client:     client,
		profileURL: profileURL,
		log:        log,
		cache:      make(map[string]cacheEntry),
		inflight:   make(map[string]*lookupFlight),
	}
}

// ProfileURL expands the configured template for one contact, or "" when
// this deployment configured no template. Nil-safe: an absent resolver (a
// deployment without a CRM) is a deployment without links.
func (r *Resolver) ProfileURL(ref ContactRef) string {
	if r == nil || r.profileURL == "" {
		return ""
	}

	return strings.Replace(r.profileURL, "{id}", ref.Get(), 1)
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

// Resolve returns the CRM's person for a number, consulting the TTL cache
// first. ok=false when the CRM is disabled, holds no contact, or fails —
// callers treat all three identically: render the raw number / skip the
// logging. Numbers are cache keys verbatim; distinct spellings of one
// number cost at most distinct entries until the CRM's canonical answer
// dominates the window.
func (r *Resolver) Resolve(ctx context.Context, number string) (Person, bool) {
	if number == "" || !r.Enabled() {
		return Person{}, false
	}

	if entry, fresh := r.cached(number); fresh {
		return entry.person, !entry.miss
	}

	if flight := r.join(number); flight != nil {
		select {
		case <-flight.done:
			return flight.person, flight.ok
		case <-ctx.Done():
			// The waiting caller gave up; the leader's answer still
			// lands in the cache for everyone after.
			return Person{}, false
		}
	}

	person, err := r.client.LookupByPhone(ctx, number)

	switch {
	case err == nil:
		r.remember(number, cacheEntry{person: person, fetched: time.Now()})
		r.settle(number, person, true)
		r.lookups.hit.Add(1)

		return person, true
	case errIsMiss(err):
		r.remember(number, cacheEntry{miss: true, fetched: time.Now()})
		r.settle(number, Person{}, false)
		r.lookups.miss.Add(1)

		return Person{}, false
	default:
		// Do not cache transport failures: the next render retries, and
		// one debug line per attempt keeps the operator trail honest.
		r.log.Debug("crm: lookup failed; rendering raw number", "error", err)
		r.settle(number, Person{}, false)
		r.lookups.failure.Add(1)

		return Person{}, false
	}
}

// Name resolves the display name for a number, or "" on any miss.
func (r *Resolver) Name(ctx context.Context, number string) string {
	person, ok := r.Resolve(ctx, number)
	if !ok {
		return ""
	}

	return person.Name
}

// namesWorkerCap bounds the fan-out of People: enough parallelism that a
// slow CRM costs ONE timeout per page (not one per number), few enough
// requests that a page render never stampedes the CRM.
const namesWorkerCap = 8

// People resolves a batch of numbers and returns the number→person map
// for view props. Unresolved numbers are absent, not zero-personed, so the
// displayName fallback needs no extra policy. Resolution runs bounded-
// concurrent: a page holds tens of numbers and each lookup may take the
// full request timeout against a slow CRM — sequential resolution would
// hang the tab render for N×timeout, violating the "decoration, never a
// prerequisite" contract of the whole seam.
//
// The PERSON, not a bare name: the CRM's contact ref is what makes a named
// caller reachable (deep link, call journal). Dropping it here was the
// seam's original lossy projection.
func (r *Resolver) People(ctx context.Context, numbers []string) map[string]Person {
	people := make(map[string]Person, len(numbers))
	if len(numbers) == 0 {
		return people
	}

	type resolved struct {
		number string
		person Person
	}
	results := make(chan resolved, len(numbers))
	jobs := make(chan string)

	workers := min(len(numbers), namesWorkerCap)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for number := range jobs {
				if person, ok := r.Resolve(ctx, number); ok && person.Name != "" {
					results <- resolved{number: number, person: person}
				}
			}
		})
	}
	for _, number := range numbers {
		jobs <- number
	}
	close(jobs)
	wg.Wait()
	close(results)

	for res := range results {
		people[res.number] = res.person
	}

	return people
}

// LogCall forwards one call activity to the CRM (no caching — a call is a
// fact, not a lookup).
func (r *Resolver) LogCall(ctx context.Context, contactID ContactRef, direction, number string, seconds int, disposition domain.Disposition) error { //nolint:branching-flow // pass-through forwarder of Client.LogCall (the verbatim-number ruling, see there)
	return r.client.LogCall(ctx, contactID, direction, number, seconds, disposition)
}

// HealthCheck forwards the authenticated health leg. The go-health probe
// discovers it as an OPTIONAL NON-CRITICAL service check when the
// composition registered the resolver NAMED "crm" (only when the seam
// is configured — a CRM-less deployment carries no crm leg at all).
func (r *Resolver) HealthCheck(ctx context.Context) error {
	return r.client.HealthCheck(ctx)
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
func (r *Resolver) settle(number string, person Person, ok bool) {
	r.mu.Lock()
	flight := r.inflight[number]
	delete(r.inflight, number)
	r.mu.Unlock()

	if flight != nil {
		flight.person = person
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
