package server

import (
	"net/http"
	"sync"
	"time"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/web/views"
)

// EventName is one SSE event kind the shell's tab listeners swap on.
// A defined type over the closed vocabulary makes the classic
// Publish(extension, event, html) transpose a compile error instead of
// a silent cross-tab breakage.
type EventName string

// SSE event names the tabs listen for (sse-swap="..." in the views).
const (
	sseEventThreads   EventName = "threads"
	sseEventThread    EventName = "thread"
	sseEventFax       EventName = "fax"
	sseEventVoicemail EventName = "voicemail"
	sseEventContacts  EventName = "contacts"
)

// ExtensionHubs gives every signed-in extension its own broadcast hub, so
// live updates reach exactly the tabs of that extension — no filtering, no
// cross-extension leakage. Hubs are created lazily and never torn down
// (they hold no per-subscriber state beyond live connections).
type ExtensionHubs struct {
	mu sync.RWMutex
	// Keyed by the branded extension — the map can never mix a raw
	// string (a phone number, a free-text id) into hub addressing.
	hubs map[domain.Extension]*cqrshtmx.Broadcaster
	// Per-extension UI language, remembered so the notifier renders SSE
	// fragments (which have no request) in the tabs' language.
	langs map[domain.Extension]views.Lang
	// Reaper bookkeeping: when each hub was last fetched or published to.
	// Mutated only under the write lock, alongside hubs, so an entry can
	// never exist in one map without the other.
	seen    map[domain.Extension]time.Time
	sweptAt time.Time
}

// Reaper settings: a hub that no one has touched for hubIdleTTL AND that
// has no live subscribers is deleted (a fresh hub is created on demand).
// The sweep runs opportunistically on the get() write path — no timer
// goroutine, deterministic under test.
const (
	hubIdleTTL    = 10 * time.Minute
	hubSweepEvery = time.Minute
)

// NewHubs builds the per-extension hub registry.
func NewHubs() *ExtensionHubs {
	return &ExtensionHubs{
		hubs:  make(map[domain.Extension]*cqrshtmx.Broadcaster),
		langs: make(map[domain.Extension]views.Lang),
		seen:  make(map[domain.Extension]time.Time),
	}
}

// SetLang remembers the extension's current UI language (called on shell
// renders and SSE connects, where the request is available).
func (h *ExtensionHubs) SetLang(extension domain.Extension, lang views.Lang) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.langs[extension] = lang
}

// Lang returns the extension's remembered UI language (default English).
func (h *ExtensionHubs) Lang(extension domain.Extension) views.Lang {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if lang, ok := h.langs[extension]; ok {
		return lang
	}
	return views.LangEN
}

func (h *ExtensionHubs) get(extension domain.Extension) *cqrshtmx.Broadcaster {
	key := extension
	h.mu.RLock()
	hub, ok := h.hubs[key]
	h.mu.RUnlock()
	if ok {
		return hub
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	h.sweep(now)
	if hub, ok := h.hubs[key]; ok {
		return hub
	}
	hub = cqrshtmx.NewBroadcaster()
	h.hubs[key] = hub
	h.seen[key] = now
	return hub
}

// Shutdown closes every hub and empties the registry (idempotent — the
// container cascade calls it too after the process layer already did).
// Closing a hub closes its subscriber channels, and ServeSSE returns the
// moment its channel closes: that is the whole point — a live tab can no
// longer hold the HTTP drain hostage for the server's full shutdown
// budget. In-flight tab nudges are dropped, deliberately: SSE payloads
// are refresh hints, not guaranteed delivery, and the next process
// serves fresh state on reconnect.
func (h *ExtensionHubs) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, hub := range h.hubs {
		hub.Close()
		delete(h.hubs, key)
		delete(h.seen, key)
	}
}

// sweep deletes hubs that are both untouched for hubIdleTTL and carry no
// live subscribers. A hub mid-connect is safe twice over: get() refreshed
// its seen stamp microseconds before ServeSSE subscribes, and once the
// subscriber exists SubscriberCount() alone protects it. Callers must
// hold the write lock.
func (h *ExtensionHubs) sweep(now time.Time) {
	if now.Sub(h.sweptAt) < hubSweepEvery {
		return
	}
	h.sweptAt = now
	for key, hub := range h.hubs {
		if hub.Hub().SubscriberCount() > 0 {
			continue
		}
		if now.Sub(h.seen[key]) < hubIdleTTL {
			continue
		}
		delete(h.hubs, key)
		delete(h.seen, key)
	}
}

// Publish renders a tab partial for one extension and pushes it to her
// tabs. The sse-swap mechanism replaces the element's innerHTML with the
// event data, so the data IS the rendered HTML.
func (h *ExtensionHubs) Publish(extension domain.Extension, eventName EventName, html string) {
	key := extension
	hub := h.get(extension)

	h.mu.Lock()
	h.seen[key] = time.Now()
	h.mu.Unlock()

	hub.Broadcast(sse.Event{Event: string(eventName), Data: html})
}

// events is the session-gated SSE feed for the signed-in extension.
// Webphone authenticates and remembers the negotiated UI language, then
// hands the connection to the library's ServeSSE: subscribe, an initial
// "connected" frame, a 15 s heartbeat, pump until disconnect, unsubscribe,
// close. The hand loop that duplicated that lifecycle is gone — and with
// it, the heartbeat that outlived its request context.
func (h *handlers) events(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}

	hub := h.deps.Hubs.get(sess.Extension)
	h.deps.Hubs.SetLang(sess.Extension, h.lang(r))

	hub.ServeSSE(w, r)
}
