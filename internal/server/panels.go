package server

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/web/views"
)

// Panel builders turn service reads into tab components. Each returns a
// ready-to-render component so pages and SSE pushes share one renderer.

func (h *handlers) messagesPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	threads, err := h.deps.Messaging.Threads(r.Context(), sess.Extension)
	if err != nil {
		return nil, err
	}
	return views.ThreadsPanel(views.ThreadsPanelProps{Threads: threads}), nil
}

func (h *handlers) threadPanel(r *http.Request, sess session.Session, id domain.ThreadID) (templ.Component, error) {
	thread, msgs, err := h.deps.Messaging.Thread(r.Context(), sess.Extension, id)
	if errors.Is(err, store.ErrNotFound) {
		return views.ThreadsPanel(views.ThreadsPanelProps{Error: "That conversation no longer exists."}), nil
	}
	if err != nil {
		return nil, err
	}
	if err := h.deps.Messaging.MarkRead(r.Context(), sess.Extension, id); err != nil {
		// marking read is cosmetic; a failure must not block the transcript
		_ = err
	}
	h.unread.drop(sess.Extension)
	return views.ThreadView(views.ThreadViewProps{Thread: thread, Messages: msgs}), nil
}

func (h *handlers) partialThread(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	id := domain.MustThreadID(r.PathValue("id"))
	component, err := h.threadPanel(r, sess, id)
	if err != nil {
		http.Error(w, "load conversation: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (h *handlers) faxPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	jobs, err := h.deps.Fax.List(r.Context(), sess.Extension)
	if err != nil {
		return nil, err
	}
	return views.FaxPanel(views.FaxPanelProps{Jobs: jobs}), nil
}

func (h *handlers) voicemailPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	if !h.deps.PhoneAPI.Enabled() {
		return views.VoicemailPanel(views.VoicemailPanelProps{}), nil
	}
	creds := pbx.Credentials{Extension: sess.Extension.String(), Password: sess.Password}
	summary, messages, err := h.fetchVoicemail(r, creds)
	if err != nil {
		return views.VoicemailPanel(views.VoicemailPanelProps{
			Enabled: true, Error: "Voicemail is unreachable right now.",
		}), nil
	}
	return views.VoicemailPanel(views.VoicemailPanelProps{
		Enabled: true, Summary: summary, Messages: messages,
	}), nil
}

func (h *handlers) fetchVoicemail(r *http.Request, creds pbx.Credentials) (pbx.VoicemailSummary, []pbx.VoicemailMessage, error) {
	ctx := r.Context()
	summary, err := h.deps.PhoneAPI.VoicemailSummary(ctx, creds)
	if err != nil {
		return pbx.VoicemailSummary{}, nil, err
	}
	page, err := h.deps.PhoneAPI.VoicemailMessages(ctx, creds)
	if err != nil {
		return summary, nil, err
	}
	return summary, page.Messages, nil
}

func (h *handlers) historyPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	if !h.deps.PhoneAPI.Enabled() {
		return views.HistoryPanel(views.HistoryPanelProps{}), nil
	}
	page, err := h.deps.PhoneAPI.History(r.Context(), pbx.Credentials{
		Extension: sess.Extension.String(), Password: sess.Password,
	}, 30)
	if err != nil {
		return views.HistoryPanel(views.HistoryPanelProps{
			Enabled: true, Error: "Call records are unreachable right now.",
		}), nil
	}
	return views.HistoryPanel(views.HistoryPanelProps{Enabled: true, Entries: page.Entries}), nil
}

func (h *handlers) contactsPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	personal, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		return nil, err
	}
	return views.ContactsPanel(views.ContactsPanelProps{Personal: personal, Shared: h.deps.Shared}), nil
}

func (h *handlers) settingsPanel() templ.Component {
	websocketURL := "wss://<this-host>" + h.deps.Config.WebsocketPath
	return views.SettingsPanel(views.SettingsPanelProps{
		SIPDomain:      h.deps.Config.SIPDomain,
		WebsocketURL:   websocketURL,
		GatewayMode:    string(h.deps.Config.Gateway.Mode),
		PhoneAPI:       h.deps.PhoneAPI.Enabled(),
		ICEServers:     len(h.deps.Config.ICEServers),
		SharedContacts: len(h.deps.Shared),
	})
}

// renderTab adapts the builder result for the shell.
func (h *handlers) tabComponent(r *http.Request, tab views.Tab, sess session.Session) (templ.Component, error) {
	switch tab {
	case views.TabMessages:
		return h.messagesPanel(r, sess)
	case views.TabFax:
		return h.faxPanel(r, sess)
	case views.TabVoicemail:
		return h.voicemailPanel(r, sess)
	case views.TabHistory:
		return h.historyPanel(r, sess)
	case views.TabContacts:
		return h.contactsPanel(r, sess)
	default:
		return h.settingsPanel(), nil
	}
}

func errorPanel(message string) templ.Component {
	return views.ErrorPanel(message)
}

func csrfToken(r *http.Request) string {
	return httputil.CSRFTokenFromRequest(r)
}
