package server

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/a-h/templ"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/web/views"
)

// Notifier turns service-layer changes into per-extension SSE pushes.
// It is the ChangeFunc target main wires into the messaging and fax
// services; it queries the stores directly so wiring has no cycle.
//
// Every payload is a swap-safe fragment (ThreadsList, Transcript,
// FaxList): exactly the inner region a sse-swap element replaces, never
// a wrapper — so live pushes cannot nest sections or wipe a draft that
// sits in a composer outside the swapped region.
type Notifier struct {
	hubs     *ExtensionHubs
	messages readmodels.ThreadGate
	faxes    readmodels.FaxGate
	contacts readmodels.ContactGate
	crm      *crm.Resolver
	asr      bool
	faxLane  views.FaxCapability
}

// NewNotifier builds the change notifier. The resolver may be nil (CRM
// integration off): fragments then render raw numbers, same as a miss.
// The contacts gate may be nil (test compositions): personal names then
// degrade away the same way. Both halves ride the ONE directoryFor merge
// the request handlers use — a pushed list must never rename a row a
// full render named differently, or a morph flips the name back to a
// number on the next live push. asrOn rides thread pushes so
// audio-attachment transcribe buttons survive SSE re-renders exactly
// when the seam is configured (a boot-time flag, not a per-request
// read). faxLane does the same job for the fax fragment's empty state:
// an SSE payload must not promise a send affordance its deployment
// lacks (the panel keeps the same flags).
func NewNotifier(hubs *ExtensionHubs, messages readmodels.ThreadGate, faxes readmodels.FaxGate, contacts readmodels.ContactGate, crmResolver *crm.Resolver, asrOn bool, faxLane views.FaxCapability) *Notifier {
	return &Notifier{hubs: hubs, messages: messages, faxes: faxes, contacts: contacts, crm: crmResolver, asr: asrOn, faxLane: faxLane}
}

// Project is the change-notification projection (train E1): each
// domain event kind maps to the SSE pushes it owes. The map is the
// ONE home of notification policy — services emit facts, this switch
// decides what surfaces re-render:
//
//   - message events push the thread list AND the transcript (an open
//     conversation updates live);
//   - ThreadFlagChanged pushes the list ONLY — the event carries no
//     thread id on purpose, an empty transcript would wipe the open
//     conversation;
//   - FaxAdvanced pushes the fax list;
//   - ReviewSaved is a payload-less NUDGE on the History tab (the
//     same pattern as the voicemail/contacts events): the pending
//     review row a reader has open must not stay stale until the next
//     manual fetch — the event carries the fact, the panel re-reads
//     its own store so a pushed payload cannot disagree with a render.
func (n *Notifier) Project(ctx context.Context, e domain.Event) {
	switch ev := e.(type) {
	case domain.MessageReceived:
		n.messagesChanged(ctx, ev.Owner, ev.Thread)
	case domain.MessageDeliveryAdvanced:
		n.messagesChanged(ctx, ev.Owner, ev.Thread)
	case domain.ThreadFlagChanged:
		n.messagesChanged(ctx, ev.Owner, domain.ThreadID{})
	case domain.FaxAdvanced:
		n.faxesChanged(ctx, ev.Owner)
	case domain.ReviewSaved:
		n.reviewChanged(ev.Owner)
	}
}

// reviewChanged publishes the payload-less review nudge (F89): the
// History panel re-fetches its partial via hx-trigger="sse:review".
// An unparseable owner cannot happen for a fact the store folded, but
// a silently dropped nudge beats a panic in the notification path.
func (n *Notifier) reviewChanged(owner domain.Extension) {
	n.hubs.Publish(owner, sseEventReview, "")
}

// notifierProject hands the Notifier's projection to consumers built
// outside the injector (the api JSON surface) with the typed-nil guard
// kept in ONE place — a test composition's nil Notifier must degrade
// to "no nudge", never to a panic inside a detached job.
func notifierProject(n *Notifier) func(context.Context, domain.Event) {
	if n == nil {
		return func(context.Context, domain.Event) {}
	}
	return n.Project
}

// messagesChanged pushes a fresh thread list to the extension's tabs
// and, for the affected thread, a fresh transcript so an open
// conversation updates live. A zero threadID means a list-only nudge:
// no transcript is pushed — an empty-thread query would render an
// empty transcript and wipe whatever conversation is open. A failed
// or missing read is skipped: one lost push is cosmetic, the next
// change catches up.
func (n *Notifier) messagesChanged(ctx context.Context, owner domain.Extension, threadID domain.ThreadID) {
	lang := n.hubs.Lang(owner)
	if threadViews, err := n.messages.ListThreads(ctx, owner); err == nil {
		threads := readmodels.SummariesOfView(threadViews)
		numbers := crmNumbers(threads, func(summary store.ThreadSummary) string { return crmNumberFor(summary.Thread.Remote) })
		n.publish(ctx, owner, sseEventThreads, views.ThreadsList(threads, directoryFor(ctx, n.contacts, n.crm, owner, numbers), lang))
	} else {
		slog.Debug("sse: render thread list failed", "error", err)
	}
	if threadID == (domain.ThreadID{}) {
		return
	}
	if msgs, err := n.messages.ListMessages(ctx, owner, threadID, messaging.MessagePageSize); err == nil {
		n.publish(ctx, owner, sseEventThread, views.Transcript(msgs, lang, n.asr))
	} else {
		slog.Debug("sse: render transcript failed", "error", err)
	}
}

// faxesChanged pushes a fresh fax job list to the extension's tabs.
func (n *Notifier) faxesChanged(ctx context.Context, owner domain.Extension) {
	jobs, err := n.faxes.List(ctx, owner, 100)
	if err != nil {
		slog.Debug("sse: render fax list failed", "error", err)
		return
	}
	numbers := crmNumbers(jobs, func(job domain.FaxJob) string { return job.Remote.String() })
	n.publish(ctx, owner, sseEventFax, views.FaxList(jobs, directoryFor(ctx, n.contacts, n.crm, owner, numbers), n.faxLane, n.hubs.Lang(owner)))
}

func (n *Notifier) publish(ctx context.Context, owner domain.Extension, event EventName, component templ.Component) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		slog.Debug("sse: render push failed", "event", string(event), "error", err)
		return
	}
	n.hubs.Publish(owner, event, buf.String())
}
