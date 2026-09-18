package server

import (
	"bytes"
	"context"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/web/views"
)

// Notifier turns service-layer changes into per-extension SSE pushes.
// It is the ChangeFunc target main wires into the messaging and fax
// services; it queries the stores directly so wiring has no cycle.
type Notifier struct {
	hubs     *ExtensionHubs
	messages *store.Messages
	faxes    *store.Faxes
}

// NewNotifier builds the change notifier.
func NewNotifier(hubs *ExtensionHubs, messages *store.Messages, faxes *store.Faxes) *Notifier {
	return &Notifier{hubs: hubs, messages: messages, faxes: faxes}
}

// MessagesChanged pushes a fresh thread list to the extension's tabs.
func (n *Notifier) MessagesChanged(ctx context.Context, owner domain.Extension, _ domain.ThreadID) {
	threads, err := n.messages.ListThreads(ctx, owner)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	component := views.ThreadsPanel(views.ThreadsPanelProps{Threads: threads})
	if err := component.Render(ctx, &buf); err != nil {
		return
	}
	n.hubs.Publish(owner, sseEventThreads, buf.String())
}

// FaxChanged pushes a fresh fax job list to the extension's tabs.
func (n *Notifier) FaxChanged(ctx context.Context, owner domain.Extension, _ domain.FaxID) {
	jobs, err := n.faxes.List(ctx, owner, 100)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	component := views.FaxPanel(views.FaxPanelProps{Jobs: jobs})
	if err := component.Render(ctx, &buf); err != nil {
		return
	}
	n.hubs.Publish(owner, sseEventFax, buf.String())
}

// VoicemailChanged nudges the voicemail tab.
func (n *Notifier) VoicemailChanged(_ context.Context, owner domain.Extension) {
	n.hubs.Publish(owner, sseEventVoicemail, "")
}
