package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
)

// SSE event names the tabs listen for (sse-swap="..." in the views).
const (
	sseEventThreads   = "threads"
	sseEventThread    = "thread"
	sseEventFax       = "fax"
	sseEventVoicemail = "voicemail"
)

// ExtensionHubs gives every signed-in extension its own broadcast hub, so
// live updates reach exactly the tabs of that extension — no filtering, no
// cross-extension leakage. Hubs are created lazily and never torn down
// (they hold no per-subscriber state beyond live connections).
type ExtensionHubs struct {
	mu   sync.RWMutex
	hubs map[string]*cqrshtmx.Broadcaster
}

// NewHubs builds the per-extension hub registry.
func NewHubs() *ExtensionHubs {
	return &ExtensionHubs{hubs: make(map[string]*cqrshtmx.Broadcaster)}
}

func (h *ExtensionHubs) get(extension domain.Extension) *cqrshtmx.Broadcaster {
	key := extension.String()
	h.mu.RLock()
	hub, ok := h.hubs[key]
	h.mu.RUnlock()
	if ok {
		return hub
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if hub, ok := h.hubs[key]; ok {
		return hub
	}
	hub = cqrshtmx.NewBroadcaster()
	h.hubs[key] = hub
	return hub
}

// Publish renders a tab partial for one extension and pushes it to her
// tabs. The sse-swap mechanism replaces the element's innerHTML with the
// event data, so the data IS the rendered HTML.
func (h *ExtensionHubs) Publish(extension domain.Extension, eventName string, html string) {
	h.get(extension).Broadcast(sse.Event{Event: eventName, Data: html})
}

// events is the session-gated SSE feed for the signed-in extension.
func (h *handlers) events(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}

	stream := sse.NewStream(w, r)
	hub := h.deps.Hubs.get(sess.Extension)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go stream.Heartbeat(ctx, 15*time.Second)

	ch := hub.Hub().Subscribe()
	defer hub.Hub().Unsubscribe(ch)

	for {
		select {
		case <-stream.Context().Done():
			_ = stream.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
			return
		case event, open := <-ch:
			if !open {
				_ = stream.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return
			}
			if err := stream.Send(event); err != nil {
				slog.Debug("sse send failed", "error", err)
				_ = stream.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return
			}
		}
	}
}
