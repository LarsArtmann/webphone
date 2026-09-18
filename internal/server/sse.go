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
	"github.com/larsartmann/webphone/internal/web/views"
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
	// Per-extension UI language, remembered so the notifier renders SSE
	// fragments (which have no request) in the tabs' language.
	langs map[string]views.Lang
}

// NewHubs builds the per-extension hub registry.
func NewHubs() *ExtensionHubs {
	return &ExtensionHubs{hubs: make(map[string]*cqrshtmx.Broadcaster), langs: make(map[string]views.Lang)}
}

// SetLang remembers the extension's current UI language (called on shell
// renders and SSE connects, where the request is available).
func (h *ExtensionHubs) SetLang(extension domain.Extension, lang views.Lang) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.langs[extension.String()] = lang
}

// Lang returns the extension's remembered UI language (default English).
func (h *ExtensionHubs) Lang(extension domain.Extension) views.Lang {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if lang, ok := h.langs[extension.String()]; ok {
		return lang
	}
	return views.LangEN
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
// Webphone authenticates and remembers the negotiated UI language, then
// hands the connection to the library's ServeSSE: subscribe, an initial
// "connected" frame, a 15 s heartbeat, pump until disconnect, unsubscribe,
// close. The hand loop that duplicated that lifecycle is gone — and with
// it, the heartbeat that outlived its request context.
func (h *handlers) events(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	hub := h.deps.Hubs.get(sess.Extension)
	h.deps.Hubs.SetLang(sess.Extension, h.lang(r))

	hub.ServeSSE(w, r)
}
