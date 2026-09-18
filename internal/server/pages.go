package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

// handlers bundles the deps; every file in this package hangs methods off it.
type handlers struct {
	deps Deps
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
	var props views.ShellProps
	props.ActiveTab = tab
	props.CSRFToken = csrfToken(r)

	if sess, ok := session.From(r.Context()); ok {
		props.SignedIn = sess.Extension.String()
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
		if component, err := h.tabComponent(r, tab, sess); err == nil {
			props.TabContent = component
		} else {
			props.TabContent = errorPanel(err.Error())
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Shell(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// partial renders just the tab region for an HTMX swap.
func (h *handlers) partial(w http.ResponseWriter, r *http.Request, tab views.Tab) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	component, err := h.tabComponent(r, tab, sess)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = errorPanel(err.Error()).Render(r.Context(), w) //nolint:erraudit // best-effort write; the response is already committed
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
