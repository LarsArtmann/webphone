package crm

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"net/http"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// ErrNoContext is returned when the CRM publishes no context for a contact:
// a 404, which covers both "this contact no longer exists" and "this CRM build
// predates the context product". Both mean the same thing to a caller: render
// nothing. It is deliberately NOT a Rejection guess or an infrastructure
// failure; absence is a normal state of an optional integration.
var ErrNoContext = errNoContext{}

type errNoContext struct{}

// Error implements the error interface.
func (errNoContext) Error() string { return "crm has no context for this contact" }

// Is lets errors.Is(err, ErrNoContext) work through wraps.
func (errNoContext) Is(target error) bool { _, ok := target.(errNoContext); return ok }

// Context is the CRM's published context for one contact, as far as the phone
// can see it: the company they belong to, the deals in flight, the follow-up
// that is due, and when the CRM last saw activity.
//
// It is a SNAPSHOT of a published product, never a replica: the phone keeps it
// only in the render-time cache below, never in SQLite, never in an event.
// Every field except Name is optional; a CRM that publishes less simply yields
// a thinner answer, and the view renders fewer chips.
type Context struct {
	Ref       ContactRef
	Name      string
	Company   string
	Tags      []string
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

// contextWire is the producer's wire shape. Unknown fields are tolerated (a
// newer CRM must not break an older phone), and every field except id/name is
// optional.
type contextWire struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Tags      []string `json:"tags"`
	OpenDeals []struct {
		Title      string `json:"title"`
		Stage      string `json:"stage"`
		ValueCents int64  `json:"value_cents"`
		Currency   string `json:"currency"`
	} `json:"open_deals"`
	OpenDealCount  int    `json:"open_deal_count"`
	OpenTasks      int    `json:"open_tasks"`
	NextTask       *struct {
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
		family := errorfamily.Transient
		if resp.StatusCode < 500 {
			family = errorfamily.Rejection
		}

		return Context{}, errorfamily.Newf(family, "crm.context", "crm: context failed with status %d", resp.StatusCode)
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
		Name:      w.Name,
		Company:   w.Company,
		Tags:      w.Tags,
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
