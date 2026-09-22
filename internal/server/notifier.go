package server

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/a-h/templ"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/messaging"
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
	messages *store.Messages
	faxes    *store.Faxes
	crm      *crm.Resolver
}

// NewNotifier builds the change notifier. The resolver may be nil (CRM
// integration off): fragments then render raw numbers, same as a miss.
func NewNotifier(hubs *ExtensionHubs, messages *store.Messages, faxes *store.Faxes, crmResolver *crm.Resolver) *Notifier {
	return &Notifier{hubs: hubs, messages: messages, faxes: faxes, crm: crmResolver}
}

// MessagesChanged pushes a fresh thread list to the extension's tabs and,
// for the affected thread, a fresh transcript so an open conversation
// updates live. A failed or missing read is skipped: one lost push is
// cosmetic, the next change catches up.
func (n *Notifier) MessagesChanged(ctx context.Context, owner domain.Extension, threadID domain.ThreadID) {
	lang := n.hubs.Lang(owner)
	if threads, err := n.messages.ListThreads(ctx, owner); err == nil {
		numbers := crmNumbers(threads, func(summary store.ThreadSummary) string { return summary.Thread.Remote.String() })
		n.publish(ctx, owner, sseEventThreads, views.ThreadsList(threads, n.crm.Names(ctx, numbers), lang))
	} else {
		slog.Debug("sse: render thread list failed", "error", err)
	}
	if msgs, err := n.messages.ListMessages(ctx, owner, threadID, messaging.MessagePageSize); err == nil {
		n.publish(ctx, owner, sseEventThread, views.Transcript(msgs, lang))
	} else {
		slog.Debug("sse: render transcript failed", "error", err)
	}
}

// FaxChanged pushes a fresh fax job list to the extension's tabs.
func (n *Notifier) FaxChanged(ctx context.Context, owner domain.Extension, _ domain.FaxID) {
	jobs, err := n.faxes.List(ctx, owner, 100)
	if err != nil {
		slog.Debug("sse: render fax list failed", "error", err)
		return
	}
	numbers := crmNumbers(jobs, func(job domain.FaxJob) string { return job.Remote.String() })
	n.publish(ctx, owner, sseEventFax, views.FaxList(jobs, n.crm.Names(ctx, numbers), n.hubs.Lang(owner)))
}

func (n *Notifier) publish(ctx context.Context, owner domain.Extension, event string, component templ.Component) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		slog.Debug("sse: render push failed", "event", event, "error", err)
		return
	}
	n.hubs.Publish(owner, event, buf.String())
}
