package crm

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// ErrNoContext is returned when the CRM publishes no context for a contact: a
// 404, which covers both "this contact no longer exists" and "this CRM build
// predates the context product". Both mean the same thing to a caller: render
// nothing. Absence is a normal state of an optional integration, never an
// error surface in the phone's UI.
var ErrNoContext = errors.New("crm has no context for this contact")

// Context is the CRM's published context for one contact, as far as the phone
// can see it: the company they belong to, the deals in flight, the follow-up
// that is due, and when the CRM last saw activity.
//
// It is a SNAPSHOT of a published product, never a replica: the phone keeps it
// only in the render-time cache below, never in SQLite, never in an event.
// Every field is optional; a CRM that publishes less simply yields a thinner
// answer, and the view renders fewer chips.
//
// Identity is deliberately ABSENT here. The strip's job is context; who the
// caller is arrives from the identity product as a crm.Person, so a name field
// on this type would be a second, disagreeable home for the same fact.
type Context struct {
	Ref       ContactRef
	Company   string
	Deals     []Deal
	DealCount int
	OpenTasks int
	NextTask  *Task
	// LastActivity is zero when the CRM published no activity timestamp.
	LastActivity time.Time
}

// Deal is one open deal, reduced to what a single line of UI can say.
type Deal struct {
	Title      string
	Stage      string
	ValueCents int64
	Currency   string
}

// Task is the next follow-up that is due.
type Task struct {
	Title   string
	DueAt   time.Time
	Overdue bool
}

// contextWire is the producer's wire shape, narrowed to the facts the strip
// renders. Unknown fields are tolerated and EXPECTED (a newer CRM must not
// break an older phone) — the producer's id/name/tags/email arrive here and are
// dropped on purpose, because identity belongs to the lookup product.
type contextWire struct {
	Company   string `json:"company"`
	OpenDeals []struct {
		Title      string `json:"title"`
		Stage      string `json:"stage"`
		ValueCents int64  `json:"value_cents"`
		Currency   string `json:"currency"`
	} `json:"open_deals"`
	OpenDealCount int `json:"open_deal_count"`
	OpenTasks     int `json:"open_tasks"`
	NextTask      *struct {
		Title   string `json:"title"`
		DueAt   string `json:"due_at"`
		Overdue bool   `json:"overdue"`
	} `json:"next_task"`
	LastActivityAt string `json:"last_activity_at"`
}

// ContactContext fetches one contact's context from the CRM's machine API.
// ErrNoContext on 404; ErrUnauthorized when the bearer was rejected; a coded
// error for anything else. The ref travels verbatim: the CRM minted it.
func (c *Client) ContactContext(ctx context.Context, ref ContactRef) (Context, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/contacts/"+ref.Get()+"/context", nil, nil)
	if err != nil {
		return Context{}, err
	}
	defer drainClose(resp)

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusNotFound:
		return Context{}, ErrNoContext
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return Context{}, ErrUnauthorized
	default:
		// Local variable, not a call argument: the registry scan needs to
		// see the code (see the same note in client.go).
		family := statusFamily(resp.StatusCode)

		return Context{}, errorfamily.Newf(family, "crm.context",
			"crm: context failed with status %d", resp.StatusCode)
	}

	var buf bytes.Buffer

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return Context{}, errorfamily.WrapTransientf(err, "crm.context_read", "crm: read context response")
	}

	var parsed contextWire
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		return Context{}, errorfamily.WrapTransientf(err, "crm.context_decode", "crm: decode context response")
	}

	return parsed.toContext(ref), nil
}

// toContext maps the wire shape onto the seam's type. Optional fields degrade
// individually: a malformed timestamp or a negative count costs that one chip,
// never the whole strip, because the phone must render whatever arrived.
func (w contextWire) toContext(ref ContactRef) Context {
	out := Context{
		Ref:       ref,
		Company:   w.Company,
		DealCount: max(w.OpenDealCount, len(w.OpenDeals)),
		OpenTasks: max(w.OpenTasks, 0),
	}

	for _, deal := range w.OpenDeals {
		out.Deals = append(out.Deals, Deal{
			Title:      deal.Title,
			Stage:      deal.Stage,
			ValueCents: deal.ValueCents,
			Currency:   deal.Currency,
		})
	}

	if w.NextTask != nil && w.NextTask.Title != "" {
		task := &Task{Title: w.NextTask.Title, Overdue: w.NextTask.Overdue}
		if due, err := time.Parse(time.RFC3339, w.NextTask.DueAt); err == nil {
			task.DueAt = due
		}

		out.NextTask = task
	}

	if at, err := time.Parse(time.RFC3339, w.LastActivityAt); err == nil {
		out.LastActivity = at
	}

	return out
}

// Context cache tuning. The strip is read once per opened conversation, and
// the facts it shows move far faster than a contact's name: a deal changes
// stage, a task gets completed. So the positive window is deliberately short,
// and a MISS ("not published", "contact gone") is remembered longer — asking a
// CRM without the endpoint on every page view would be the loudest possible
// way to poll a missing product.
const (
	contextTTL        = 2 * time.Minute
	contextAbsentTTL  = 15 * time.Minute
	contextMaxEntries = 256
)

// contextEntry is one cached answer. absent is the "nothing to say" state: no
// context published, the contact is gone, or the CRM refused the token.
type contextEntry struct {
	context Context
	absent  bool
	fetched time.Time
}

// ContextOf resolves the CRM's published context for one contact, memoized for
// a short window. nil/false means "nothing to say" and covers every
// degradation: seam off, no contact, endpoint not published, CRM unreachable.
//
// No single-flight here on purpose: a page holds ONE contact's strip, so there
// is no fan-out to join, and the names single-flight exists precisely because
// a list page resolves tens of numbers.
func (r *Resolver) ContextOf(ctx context.Context, ref ContactRef) (*Context, bool) {
	if !r.Enabled() || ref.Get() == "" {
		return nil, false
	}

	if entry, fresh := r.cachedContext(ref.Get()); fresh {
		if entry.absent {
			return nil, false
		}

		out := entry.context

		return &out, true
	}

	parsed, err := r.client.ContactContext(ctx, ref)

	switch {
	case err == nil:
		r.rememberContext(ref.Get(), contextEntry{context: parsed, fetched: time.Now()})

		return &parsed, true
	case errors.Is(err, ErrNoContext):
		r.rememberContext(ref.Get(), contextEntry{absent: true, fetched: time.Now()})
	case errors.Is(err, ErrUnauthorized):
		// A revoked token is a real condition the health leg names loudly;
		// the strip stays silent and stops asking for a while.
		r.log.Debug("crm: context unauthorized; no strip", "ref", ref)

		r.rememberContext(ref.Get(), contextEntry{absent: true, fetched: time.Now()})
	default:
		// Transport/shape failures are NOT cached: the next open retries.
		r.log.Debug("crm: context failed; no strip", "ref", ref, "error", err)
	}

	return nil, false
}

func (r *Resolver) cachedContext(ref string) (contextEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.contexts[ref]
	if !ok {
		return contextEntry{}, false
	}

	ttl := contextTTL
	if entry.absent {
		ttl = contextAbsentTTL
	}

	if time.Since(entry.fetched) > ttl {
		delete(r.contexts, ref)

		return contextEntry{}, false
	}

	return entry, true
}

func (r *Resolver) rememberContext(ref string, entry contextEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Same bound-and-reset policy as the name cache: a decoration memo, not a
	// store worth an LRU.
	if len(r.contexts) >= contextMaxEntries {
		r.contexts = make(map[string]contextEntry, contextMaxEntries)
	}

	r.contexts[ref] = entry
}
