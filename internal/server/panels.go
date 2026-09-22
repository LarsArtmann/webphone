package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	numbers := make([]string, 0, len(threads))
	for _, summary := range threads {
		numbers = append(numbers, summary.Thread.Remote.String())
	}
	return views.ThreadsPanel(views.ThreadsPanelProps{Threads: threads, Identity: h.identityFor(sess.Extension), Names: h.crmNames(r.Context(), numbers), Lang: h.lang(r)}), nil
}

func (h *handlers) threadPanel(r *http.Request, sess session.Session, id domain.ThreadID, page int) (templ.Component, error) {
	thread, msgs, hasMore, err := h.deps.Messaging.ThreadWindow(r.Context(), sess.Extension, id, page)
	if errors.Is(err, store.ErrNotFound) {
		return views.ThreadsPanel(views.ThreadsPanelProps{Error: "That conversation no longer exists.", Lang: h.lang(r)}), nil
	}
	if err != nil {
		return nil, err
	}
	if err := h.deps.Messaging.MarkRead(r.Context(), sess.Extension, id); err != nil {
		// marking read is cosmetic; a failure must not block the transcript
		_ = err
	}
	h.unread.drop(sess.Extension)
	names := h.crmNames(r.Context(), []string{thread.Remote.String()})
	return views.ThreadView(views.ThreadViewProps{
		Thread: thread, Messages: msgs, Page: page, HasMore: hasMore,
		Identity: h.identityFor(sess.Extension), Names: names, Lang: h.lang(r),
	}), nil
}

func (h *handlers) partialThread(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	id, err := domain.ParseThreadID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	page := 0
	if raw := r.URL.Query().Get("older"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 10000 {
			page = parsed
		}
	}
	component, err := h.threadPanel(r, sess, id, page)
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
	numbers := make([]string, 0, len(jobs))
	for _, job := range jobs {
		numbers = append(numbers, job.Remote.String())
	}
	return views.FaxPanel(views.FaxPanelProps{Jobs: jobs, Identity: h.identityFor(sess.Extension), Names: h.crmNames(r.Context(), numbers), Lang: h.lang(r)}), nil
}

func (h *handlers) voicemailPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	if !h.deps.PhoneAPI.Enabled() {
		return views.VoicemailPanel(views.VoicemailPanelProps{Lang: h.lang(r)}), nil
	}
	creds := sess.PBXCredentials()
	summary, messages, err := h.fetchVoicemail(r, creds)
	if err != nil {
		return views.VoicemailPanel(views.VoicemailPanelProps{
			Enabled: true, Error: h.T(r, "vm.unreachable"), Lang: h.lang(r),
		}), nil
	}
	return views.VoicemailPanel(views.VoicemailPanelProps{
		Enabled: true, Summary: summary, Messages: messages,
		Names: h.crmNames(r.Context(), voicemailNumbers(messages)), Lang: h.lang(r),
	}), nil
}

// voicemailNumbers collects the resolvable caller numbers of a voicemail
// page (skipping withheld/blank CID). Order is irrelevant; the resolver
// dedupes via its cache.
func voicemailNumbers(messages []pbx.VoicemailMessage) []string {
	numbers := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.CIDNumber != "" {
			numbers = append(numbers, msg.CIDNumber)
		}
	}
	return numbers
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

// crmNames resolves numbers against the optional CRM integration for view
// props. A disabled (or nil) resolver returns an empty map: the views'
// displayName fallback renders the raw numbers unchanged.
func (h *handlers) crmNames(ctx context.Context, numbers []string) map[string]string {
	if len(numbers) == 0 {
		return nil
	}
	return h.deps.CRM.Names(ctx, numbers)
}

func (h *handlers) historyPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	lang := h.lang(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	dir := r.URL.Query().Get("dir")
	if dir != "in" && dir != "out" {
		dir = ""
	}
	if !h.deps.PhoneAPI.Enabled() {
		return views.HistoryPanel(views.HistoryPanelProps{Lang: lang}), nil
	}
	// A filter needs a wider window than the unfiltered top-30 view.
	limit := historyPageSize
	if query != "" || dir != "" {
		limit = historyFilterFetchSize
	}
	page, err := h.deps.PhoneAPI.History(r.Context(), sess.PBXCredentials(), limit)
	if err != nil {
		return views.HistoryPanel(views.HistoryPanelProps{
			Enabled: true, Error: h.T(r, "history.unreachable"), Lang: lang,
		}), nil
	}
	entries := filterCDRs(page.Entries, query, dir)
	if len(entries) > historyPageSize {
		entries = entries[:historyPageSize]
	}
	numbers := make([]string, 0, len(entries))
	for _, cdr := range entries {
		if dial := views.CDRDialTarget(cdr); dial != "" {
			numbers = append(numbers, dial)
		}
	}
	return views.HistoryPanel(views.HistoryPanelProps{
		Enabled: true, Entries: entries, Query: query, Dir: dir, Names: h.crmNames(r.Context(), numbers), Lang: lang,
	}), nil
}

// History sizes: the unfiltered view and the upstream window a filter
// may search before the display cap applies again.
const (
	historyPageSize        = 30
	historyFilterFetchSize = 100
)

// filterCDRs keeps records whose number/name contains the query (case-
// insensitive) and whose direction matches "in" (public context) or
// "out" (internal context).
func filterCDRs(entries []pbx.CDR, query, dir string) []pbx.CDR {
	if query == "" && dir == "" {
		return entries
	}
	needle := strings.ToLower(query)
	filtered := entries[:0:0]
	for _, cdr := range entries {
		if dir == "in" && cdr.Context != "public" {
			continue
		}
		if dir == "out" && cdr.Context == "public" {
			continue
		}
		if needle != "" {
			haystack := strings.ToLower(cdr.CallerIDNumber + " " + cdr.CallerIDName + " " + cdr.DestinationNumber)
			if !strings.Contains(haystack, needle) {
				continue
			}
		}
		filtered = append(filtered, cdr)
	}
	return filtered
}

func (h *handlers) contactsPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	personal, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		return nil, err
	}
	return views.ContactsPanel(views.ContactsPanelProps{Personal: personal, Shared: h.deps.Shared, Lang: h.lang(r)}), nil
}

func (h *handlers) settingsPanel(r *http.Request) templ.Component {
	websocketURL := "wss://<this-host>" + h.deps.Config.WebsocketPath
	return views.SettingsPanel(views.SettingsPanelProps{
		SIPDomain:      h.deps.Config.SIPDomain,
		WebsocketURL:   websocketURL,
		GatewayMode:    string(h.deps.Config.Gateway.Mode),
		PhoneAPI:       h.deps.PhoneAPI.Enabled(),
		ICEServers:     len(h.deps.Config.ICEServers),
		SharedContacts: len(h.deps.Shared),
		Lang:           h.lang(r),
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
		return h.settingsPanel(r), nil
	}
}

func errorPanel(message string, lang views.Lang) templ.Component {
	return views.ErrorPanel(message, lang)
}

func csrfToken(r *http.Request) string {
	return httputil.CSRFTokenFromRequest(r)
}
