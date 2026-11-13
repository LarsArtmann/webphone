package server

import (
	"context"
	"log/slog"
	"maps"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/web/views"
)

// crmContacts is the ONE conversion from the integration seam's type to the
// view projection: the published name always, a deep link only when the
// deployment configured a template. Views never import internal/crm, so no
// surface can grow a second, disagreeing notion of "who this is".
//
// A nil resolver (test compositions, integration off) yields nil, which is
// what displayName's raw-number fallback already expects.
func crmContacts(people map[string]crm.Person, resolver *crm.Resolver) views.CRMContacts {
	if len(people) == 0 {
		return nil
	}

	contacts := make(views.CRMContacts, len(people))
	for number, person := range people {
		contacts[number] = views.CRMContact{
			Name: person.Name,
			URL:  resolver.ProfileURL(person.Ref),
		}
	}

	return contacts
}

// personalNames is the number→name map the owner's own contacts store
// answers for the numbers a surface is about to render. One home for the
// personal half of every display directory: request handlers AND the SSE
// push lane build their directory through directoryFor, so a live morph
// can never rename a row a full render named differently (the
// split-brain this seam exists to prevent).
//
// A nil gate (test compositions) or a failed read yields nil: names
// degrade to the CRM half, a contacts outage must not blank the panels.
func personalNames(ctx context.Context, contacts readmodels.ContactGate, owner domain.Extension, numbers []string) map[string]string {
	if contacts == nil || len(numbers) == 0 {
		return nil
	}

	personal, err := contacts.List(ctx, owner)
	if err != nil {
		slog.Warn("contacts directory read failed; names degrade to CRM/numbers", "owner", owner.String(), "error", err)
		return nil
	}

	wanted := make(map[string]struct{}, len(numbers))
	for _, number := range numbers {
		wanted[number] = struct{}{}
	}

	names := make(map[string]string, len(numbers))
	for _, contact := range personal {
		phone := contact.Phone.String()
		if _, needed := wanted[phone]; needed && contact.Name != "" {
			names[phone] = contact.Name
		}
	}

	return names
}

// directoryFor merges the display directory every row-label view reads:
// the owner's own contacts name a number FIRST (the curated personal
// record beats the CRM — owner call 2026-11-10, flippable by reordering
// this merge), the optional CRM answers the numbers nobody saved. Both
// lanes — request handlers and the SSE push lane — merge through this
// ONE home.
//
// Personal names carry no deep link (the Contacts panel owns those). Zero
// config on both seams renders raw numbers with no link: byte-identical
// to the pre-integration product.
func directoryFor(ctx context.Context, contacts readmodels.ContactGate, resolver *crm.Resolver, owner domain.Extension, numbers []string) views.CRMContacts {
	if len(numbers) == 0 {
		return nil
	}

	directory := make(views.CRMContacts, len(numbers))
	remaining := make([]string, 0, len(numbers))
	personal := personalNames(ctx, contacts, owner, numbers)
	for _, number := range numbers {
		if name, ok := personal[number]; ok {
			directory[number] = views.CRMContact{Name: name}
			continue
		}
		remaining = append(remaining, number)
	}

	maps.Copy(directory, crmContacts(resolver.People(ctx, remaining), resolver))

	return directory
}

// crmNumberFor reports the string the CRM should be asked about for a thread's
// remote, or "" when the CRM cannot possibly answer. Matrix user ids and
// Session ids are not phone numbers: a lookup miss at best, a 400 at worst, and
// always a request nobody needed. crmNumbers' existing ""-skip does the rest.
func crmNumberFor(remote domain.Address) string {
	if !remote.IsPhoneShaped() {
		return ""
	}

	return remote.String()
}

// crmContextShown is how many open deals a strip renders. Two is the most a
// single line can carry before the chips stop being scannable; the tail is
// counted, never shipped to the markup.
const crmContextShown = 2

// crmContextFor resolves the CRM's published context for a thread's remote
// party, or nil when there is nothing to show. It is deliberately scoped to
// PHONE-SHAPED addresses (the same rule the identity lookups obey, via
// crmNumberFor): the CRM indexes phone numbers, so asking it about a matrix or
// session address would be a guaranteed 400 per page view.
//
// Every degradation returns nil: seam off, no contact, no context published,
// CRM unreachable. The surface then renders exactly what a CRM-less
// deployment renders.
func (h *handlers) crmContextFor(ctx context.Context, remote domain.Address) *views.CRMContext {
	if !remote.IsPhoneShaped() {
		return nil
	}

	person, ok := h.deps.CRM.Resolve(ctx, remote.String())
	if !ok {
		return nil
	}

	published, fromCache, ok := h.deps.CRM.ContextOfCached(ctx, person.Ref)
	if !ok {
		return nil
	}

	view := crmContextView(*published)
	if view != nil {
		view.ServedFromCache = fromCache
	}

	return view
}

// crmContextView projects the seam's context onto the view's. It returns nil
// for a context that carries nothing renderable: an empty shell (a bordered
// strip with no chips) is worse than no strip, so the decision is made here,
// once, instead of in the template.
func crmContextView(published crm.Context) *views.CRMContext {
	shown := published.Deals
	if len(shown) > crmContextShown {
		shown = shown[:crmContextShown]
	}

	out := &views.CRMContext{
		Company:      published.Company,
		OpenTasks:    published.OpenTasks,
		LastActivity: published.LastActivity,
	}

	for _, deal := range shown {
		out.Deals = append(out.Deals, views.CRMDeal{
			Title:      deal.Title,
			Stage:      deal.Stage,
			ValueCents: deal.ValueCents,
			Currency:   deal.Currency,
		})
	}

	more := published.DealCount - len(out.Deals)
	if more > 0 {
		out.MoreDeals = more
	}

	if published.NextTask != nil {
		out.NextTask = published.NextTask.Title
		out.NextTaskDue = published.NextTask.DueAt
		out.Overdue = published.NextTask.Overdue
	}

	if out.Company == "" && len(out.Deals) == 0 && out.OpenTasks == 0 && out.LastActivity.IsZero() {
		return nil
	}

	return out
}
