package server

import (
	"net/http"
	"strings"

	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

// handlers bundles the deps; every file in this package hangs methods off it.
type handlers struct {
	deps Deps
	// Per-client flood protection for the two unauthenticated-by-session
	// surfaces: login attempts and inbound webhooks.
	loginLimiter *keyedLimiter
	hookLimiter  *keyedLimiter
	// Memoized nav-badge totals, invalidated on every unread mutation.
	unread *unreadCache
}

// page renders the full shell for the root URL.
func (h *handlers) page(w http.ResponseWriter, r *http.Request) {
	h.renderShell(w, r, views.TabMessages)
}

// tabPage renders the full shell for a tab URL (deep links).
func (h *handlers) tabPage(w http.ResponseWriter, r *http.Request) {
	h.renderShell(w, r, tabFromPath(r.URL.Path))
}

func (h *handlers) renderShell(w http.ResponseWriter, r *http.Request, tab views.Tab) {
	lang := h.lang(r)
	var props views.ShellProps
	props.ActiveTab = tab
	props.CSRFToken = csrfToken(r)
	props.Lang = lang

	if sess, ok := session.From(r.Context()); ok {
		props.SignedIn = sess.Extension.String()
		// Keep the hub's language fresh so SSE fragments match the tabs.
		h.deps.Hubs.SetLang(sess.Extension, lang)
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
		if component, err := h.tabComponent(r, tab, sess); err == nil {
			props.TabContent = component
		} else {
			props.TabContent = errorPanel(err.Error(), lang)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Shell(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// partial renders just the tab region for an HTMX swap.
func (h *handlers) partial(w http.ResponseWriter, r *http.Request, tab views.Tab) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	component, err := h.tabComponent(r, tab, sess)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = errorPanel(err.Error(), h.lang(r)).Render(r.Context(), w) //nolint:erraudit // best-effort write; the response is already committed
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (h *handlers) partialMessages(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabMessages)
}

func (h *handlers) partialFax(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabFax)
}

func (h *handlers) partialVoicemail(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabVoicemail)
}

func (h *handlers) partialHistory(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabHistory)
}

func (h *handlers) partialContacts(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabContacts)
}

func (h *handlers) partialSettings(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabSettings)
}

func tabFromPath(path string) views.Tab {
	switch path {
	case "/fax", "/fax/":
		return views.TabFax
	case "/voicemail", "/voicemail/":
		return views.TabVoicemail
	case "/history", "/history/":
		return views.TabHistory
	case "/contacts", "/contacts/":
		return views.TabContacts
	case "/settings", "/settings/":
		return views.TabSettings
	default:
		return views.TabMessages
	}
}

// langCookie carries the island's language choice to the server so the
// tabs and SSE fragments render in the same language as the phone panel.
const langCookie = "wp-lang"

// lang resolves the request's UI language: explicit cookie (written by
// the island's EN/DE switch), then the Accept-Language header, then
// English.
func (h *handlers) lang(r *http.Request) views.Lang {
	if cookie, err := r.Cookie(langCookie); err == nil {
		return views.ParseLang(cookie.Value)
	}
	if accept := r.Header.Get("Accept-Language"); len(accept) >= 2 && strings.EqualFold(accept[:2], "de") {
		return views.LangDE
	}
	return views.LangEN
}

// T translates a UI string in the request's language.
func (h *handlers) T(r *http.Request, key string) string {
	return views.T(h.lang(r), key)
}
