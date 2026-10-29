package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/textagent"
	"github.com/larsartmann/webphone/internal/web/views"
)

// Panel builders turn service reads into tab components. Each returns a
// ready-to-render component so pages and SSE pushes share one renderer.

func (h *handlers) messagesPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	return h.messagesPanelPrefilled(r, sess, "")
}

// messagesPanelPrefilled renders the Messages tab with an AI-drafted
// body pre-seeded into the new-conversation composer (the ai-draft
// round trip); the plain panel passes "".
func (h *handlers) messagesPanelPrefilled(r *http.Request, sess session.Session, prefillBody string) (templ.Component, error) {
	// The search box's q rides the same panel route; empty/whitespace
	// means "no filter" and renders the plain list. ?archived=1 flips to
	// the filing cabinet (design D6): the archived list, no search box,
	// a way back. ArchivedCount feeds the active view's toggle link.
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	archived := r.URL.Query().Get("archived") == "1"
	var (
		threads       []store.ThreadSummary
		archivedCount int
		err           error
	)
	if archived {
		threads, err = h.deps.Messaging.ArchivedThreads(r.Context(), sess.Extension)
	} else if query == "" {
		threads, err = h.deps.Messaging.Threads(r.Context(), sess.Extension)
	} else {
		threads, err = h.deps.Messaging.ThreadSearch(r.Context(), sess.Extension, query)
	}
	if err != nil {
		return nil, err //nolint:erraudit // context_loss FP: threads is the OUT param, garbage exactly when the fetch failed
	}
	if !archived {
		if archivedCount, err = h.deps.Messaging.ArchivedCount(r.Context(), sess.Extension); err != nil {
			return nil, err //nolint:erraudit // context_loss FP: archivedCount is an OUT param, garbage exactly when the count failed
		}
	}
	numbers := crmNumbers(threads, func(summary store.ThreadSummary) string { return summary.Thread.Remote.String() })
	reviews, draftsErr := h.textAgentReviews(r)
	// The drafts panel shares the thread rows' name feed: every pending
	// draft's contact resolves through the same CRM map.
	if reviews != nil {
		numbers = append(numbers, crmNumbers(reviews.Drafts, func(draft textagent.Draft) string { return draft.Contact })...)
	}
	contacts := h.crmDirectory(r.Context(), numbers)
	return views.ThreadsPanel(views.ThreadsPanelProps{
		Threads: threads, Query: query, Archived: archived, ArchivedCount: archivedCount,
		Identity: h.identityFor(sess.Extension), CRM: contacts, Lang: h.lang(r),
		TextAgent: h.textAgentOn(), Drafts: reviews, DraftsError: draftsErr, PrefillBody: prefillBody,
	}), nil
}

func (h *handlers) threadPanel(r *http.Request, sess session.Session, id domain.ThreadID, page int) (templ.Component, error) {
	return h.threadPanelPrefilled(r, sess, id, page, "")
}

// threadPanelPrefilled renders an open conversation with an AI-drafted
// reply pre-seeded into the composer (the ai-draft round trip); the
// plain panel passes "".
func (h *handlers) threadPanelPrefilled(r *http.Request, sess session.Session, id domain.ThreadID, page int, prefillDraft string) (templ.Component, error) {
	thread, msgs, hasMore, err := h.deps.Messaging.ThreadWindow(r.Context(), sess.Extension, id, page)
	if errors.Is(err, store.ErrNotFound) {
		return threadGone(h.lang(r)), nil
	}
	if err != nil {
		return nil, err
	}
	if err := h.deps.Messaging.MarkRead(r.Context(), sess.Extension, id); err != nil {
		// marking read is cosmetic; a failure must not block the transcript
		_ = err
	}
	h.unread.Drop(sess.Extension)
	// The master rail (P1): the same active-thread list the Messages
	// landing view shows, rendered beside the conversation. A failed read
	// degrades to NO rail — the conversation is the page's subject, and
	// hiding it behind a list error would be a regression against the
	// pre-split single-column shape. The failure is LOGGED (never
	// swallowed) because "the split silently stopped splitting" is exactly
	// the class of invisible breakage the error-surface ruling forbids.
	rail, railErr := h.deps.Messaging.Threads(r.Context(), sess.Extension)
	if railErr != nil {
		errorfamily.LogErrorContext(r.Context(), fmt.Errorf("conversation rail: %w", railErr), nil)
		rail = nil
	}
	numbers := append(
		[]string{thread.Remote.String()},
		crmNumbers(rail, func(summary store.ThreadSummary) string { return summary.Thread.Remote.String() })...,
	)
	contacts := h.crmDirectory(r.Context(), numbers)
	// Reply snippets feed the composer's chip lane + picker (M21);
	// nil-safe: hand-composed test Deps may carry no snippet store.
	var snippets []domain.Snippet
	if h.deps.Snippets != nil {
		if snippets, err = h.deps.Snippets.List(r.Context(), sess.Extension); err != nil {
			return nil, err
		}
	}
	return views.ThreadView(views.ThreadViewProps{
		Thread: thread, Messages: msgs, Page: page, HasMore: hasMore, Snippets: snippets, Rail: rail,
		Identity: h.identityFor(sess.Extension), CRM: contacts, Lang: h.lang(r),
		ASR: h.asrOn(), TextAgent: h.textAgentOn(), PrefillDraft: prefillDraft,
	}), nil
}

// threadGone is the graceful fallback when a linked conversation cannot
// be resolved (deleted, foreign, or unparseable id): the list plus a
// notice. The deep link keeps the full shell instead of a bare 404.
func threadGone(lang views.Lang) templ.Component {
	return views.ThreadsPanel(views.ThreadsPanelProps{Error: "That conversation no longer exists.", Lang: lang})
}

// threadPageFromQuery reads the transcript paging cursor (?older=N);
// absent or out-of-range values mean page 0 (the newest window).
func threadPageFromQuery(r *http.Request) int {
	if raw := r.URL.Query().Get("older"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 10000 {
			return parsed
		}
	}
	return 0
}

func (h *handlers) partialThread(w http.ResponseWriter, r *http.Request) {
	sess, id, ok := h.requireThread(w, r, h.notFound)
	if !ok {
		return
	}
	component, err := h.threadPanel(r, sess, id, threadPageFromQuery(r))
	if err != nil {
		h.internalError(w, r, "load conversation", err)
		return
	}
	renderHTML(w, r, component)
}

func (h *handlers) faxPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	jobs, err := h.deps.Fax.List(r.Context(), sess.Extension)
	if err != nil {
		return nil, err
	}
	numbers := crmNumbers(jobs, func(job domain.FaxJob) string { return job.Remote.String() })
	return views.FaxPanel(views.FaxPanelProps{Jobs: jobs, Identity: h.identityFor(sess.Extension), CRM: h.crmDirectory(r.Context(), numbers), Lang: h.lang(r)}), nil
}

func (h *handlers) voicemailPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	if !h.deps.PhoneAPI.Enabled() {
		return views.VoicemailPanel(views.VoicemailPanelProps{Lang: h.lang(r), ASR: h.asrOn()}), nil
	}
	creds := sess.PBXCredentials()
	summary, messages, err := h.fetchVoicemail(r, creds)
	if err != nil {
		return views.VoicemailPanel(views.VoicemailPanelProps{
			Enabled: true, Error: h.T(r, "vm.unreachable"), Lang: h.lang(r), ASR: h.asrOn(),
		}), nil
	}
	return views.VoicemailPanel(views.VoicemailPanelProps{
		Enabled: true, Summary: summary, Messages: messages,
		CRM: h.crmDirectory(r.Context(), crmNumbers(messages, func(msg pbx.VoicemailMessage) string { return msg.CIDNumber })), Lang: h.lang(r),
		ASR: h.asrOn(),
	}), nil
}

// crmNumbers collects the numbers of a page of rows for CRM resolution,
// skipping blanks (withheld/blank caller ID, CDRs without a dial target).
// Order is irrelevant; the resolver dedupes via its cache.
func crmNumbers[T any](rows []T, number func(T) string) []string {
	numbers := make([]string, 0, len(rows))
	for _, row := range rows {
		if n := number(row); n != "" {
			numbers = append(numbers, n)
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

// crmDirectory resolves numbers against the optional CRM integration for
// view props: the published name plus, when a deep-link template is
// configured, the address of that contact in the CRM's own UI.
//
// A disabled (or nil) resolver returns nil, so the views' displayName
// fallback renders raw numbers with no link: zero CRM config renders
// byte-identically to the pre-integration product.
func (h *handlers) crmDirectory(ctx context.Context, numbers []string) views.CRMContacts {
	if len(numbers) == 0 {
		return nil
	}
	return crmContacts(h.deps.CRM.People(ctx, numbers), h.deps.CRM)
}

// asrOn is the one gate every server-rendered transcribe affordance
// reads: nil-safe over the optional seam, so a hand-composed test Deps
// (no ASR) renders no dead buttons.
func (h *handlers) asrOn() bool {
	return h.deps.ASR != nil && h.deps.ASR.Enabled()
}

func (h *handlers) historyPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	lang := h.lang(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	dir := r.URL.Query().Get("dir")
	if dir != "in" && dir != "out" {
		dir = ""
	}
	outcome := r.URL.Query().Get("outcome")
	if outcome != "answered" && outcome != "missed" {
		outcome = ""
	}
	disposition := historyDispositionFilter(r.URL.Query().Get("disposition"))
	// Transcripts are the extension's OWN records (no phone API needed):
	// they render even when the CDR half of the tab is disabled or
	// unreachable. A store failure degrades to no section — the CDR list
	// must not die with it.
	transcripts := h.recentTranscripts(r, sess, query, disposition)
	if disposition != "" {
		// A verdict filter cannot address the CDR rows (they carry no call
		// uuid to join the disposition with — the documented invariant),
		// so the CDR section stays empty rather than showing rows the
		// filter did not choose. The transcripts carry the answer.
		return views.HistoryPanel(views.HistoryPanelProps{
			Enabled: true, Disposition: disposition.String(), Lang: lang, Transcripts: transcripts,
			AgentEnabled: h.textAgentOn(),
		}), nil
	}
	if !h.deps.PhoneAPI.Enabled() {
		return views.HistoryPanel(views.HistoryPanelProps{
			Lang: lang, Query: query, Transcripts: transcripts,
			AgentEnabled: h.textAgentOn(),
		}), nil
	}
	// A filter needs a wider window than the unfiltered top-30 view.
	limit := historyPageSize
	if query != "" || dir != "" || outcome != "" {
		limit = historyFilterFetchSize
	}
	page, err := h.deps.PhoneAPI.History(r.Context(), sess.PBXCredentials(), limit)
	if err != nil {
		return views.HistoryPanel(views.HistoryPanelProps{
			Enabled: true, Error: h.T(r, "history.unreachable"), Lang: lang,
			AgentEnabled: h.textAgentOn(),
		}), nil
	}
	entries := filterCDRs(page.Entries, query, dir, outcome)
	if len(entries) > historyPageSize {
		entries = entries[:historyPageSize]
	}
	numbers := crmNumbers(entries, views.CDRDialTarget)
	return views.HistoryPanel(views.HistoryPanelProps{
		Enabled: true, Entries: entries, Query: query, Dir: dir, Outcome: outcome,
		CRM: h.crmDirectory(r.Context(), numbers), Lang: lang, Transcripts: transcripts,
		AgentEnabled: h.textAgentOn(),
	}), nil
}

// recentTranscripts builds the History tab's transcript groups with CRM
// name enrichment. Nil store (test compositions) and read failures both
// yield nil — the section simply does not render. An active query turns
// the fetch into the owner-scoped LIKE search: the section then shows
// only calls whose transcript text matched (matching lines only — the
// hint in the view says so). An active disposition filter narrows the
// groups to calls whose note slot carries that confirmed verdict (T3)
// and widens the window, since the verdict lives on the note slot and
// older transcribed calls must stay reachable for the join.
func (h *handlers) recentTranscripts(r *http.Request, sess session.Session, query string, disposition domain.Disposition) []views.TranscriptGroup {
	if h.deps.Transcripts == nil {
		return nil
	}
	window := historyTranscriptGroups
	if disposition != "" {
		window = historyFilterFetchSize
	}
	var (
		recent []domain.CallTranscript
		err    error
	)
	if query != "" {
		recent, err = h.deps.Transcripts.Search(r.Context(), sess.Extension, query, window)
	} else {
		recent, err = h.deps.Transcripts.Recent(r.Context(), sess.Extension, window)
	}
	if err != nil || len(recent) == 0 {
		if err != nil {
			slog.Warn("history: transcript fetch failed", "error", err, "extension", sess.Extension.String())
		}
		return nil
	}
	numbers := crmNumbers(recent, func(tx domain.CallTranscript) string { return tx.Remote })
	contacts := h.crmDirectory(r.Context(), numbers)
	groups := make([]views.TranscriptGroup, 0, len(recent))
	for _, tx := range recent {
		note := ""
		noteDisposition := ""
		if h.deps.Notes != nil {
			if view, err := h.deps.Notes.Get(r.Context(), sess.Extension, tx.CallID); err != nil {
				slog.Warn("history: note fetch failed", "error", err, "extension", sess.Extension.String())
			} else if view != nil {
				note = view.Body
				noteDisposition = view.Disposition
			}
		}
		reviewState, reviewText, reviewConfidence := "", "", 0
		if h.deps.Reviews != nil {
			if view, err := h.deps.Reviews.Get(r.Context(), sess.Extension, tx.CallID); err != nil {
				slog.Warn("history: review fetch failed", "error", err, "extension", sess.Extension.String())
			} else if view != nil {
				reviewState, reviewText, reviewConfidence = view.State, view.Text, view.Confidence
			}
		}
		if disposition != "" && noteDisposition != disposition.String() {
			continue
		}
		groups = append(groups, views.TranscriptGroup{
			CallID: tx.CallID,
			Remote: tx.Remote, Direction: string(tx.Direction), StartedAt: tx.StartedAt,
			Lines: tx.Lines, Name: names[tx.Remote], Note: note, Disposition: noteDisposition,
			ReviewState: reviewState, ReviewText: reviewText, ReviewConfidence: reviewConfidence,
		})
	}
	return groups
}

// historyDispositionFilter validates the disposition query param: only
// the five operator-confirmable verdicts filter; anything else (empty,
// unknown, garbage) means no filter — a filter param never errors the
// page, it just doesn't filter.
func historyDispositionFilter(raw string) domain.Disposition {
	disposition, err := domain.ParseDisposition(raw)
	if err != nil || disposition == domain.DispositionUnknown {
		return ""
	}
	return disposition
}

// historyTranscriptGroups caps the History tab's transcript section: the
// newest transcribed calls, whole conversations each.
const historyTranscriptGroups = 10

// History sizes: the unfiltered view and the upstream window a filter
// may search before the display cap applies again.
const (
	historyPageSize        = 30
	historyFilterFetchSize = 100
)

// filterCDRs keeps records whose number/name contains the query (case-
// insensitive), whose direction matches "in" (public context) or "out"
// (internal context), and whose outcome matches: "missed" is an inbound
// leg that never carried media (billsec 0), "answered" is any leg with
// billsec > 0. Outbound legs bill nothing until answered, so they can
// never be "missed" by this definition.
func filterCDRs(entries []pbx.CDR, query, dir, outcome string) []pbx.CDR {
	if query == "" && dir == "" && outcome == "" {
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
		if outcome == "missed" && (cdr.Context != "public" || cdr.Billsec != 0) {
			continue
		}
		if outcome == "answered" && cdr.Billsec == 0 {
			continue
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

// settingsCRM states — the Settings tab's CRM row renderings (T7).
const (
	settingsCRMUnset      = "unset"
	settingsCRMSetup      = "setup"
	settingsCRMConnected  = "connected"
	settingsCRMAuthFailed = "auth-failed"
)

// crmSettingsProbeTimeout bounds the Settings row's CRM health ask —
// one authenticated round trip, fail-open (the agent-row probe's
// posture at a tighter budget: the CRM is a loopback/LAN service).
const crmSettingsProbeTimeout = 2 * time.Second

func (h *handlers) settingsPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	websocketURL := "wss://<this-host>" + h.deps.Config.WebsocketPath
	// The snippets section is the tab's one managed list (M21); nil-safe
	// for hand-composed test Deps without a snippet store.
	var snippets []domain.Snippet
	if h.deps.Snippets != nil {
		var err error
		if snippets, err = h.deps.Snippets.List(r.Context(), sess.Extension); err != nil {
			return nil, err
		}
	}
	// The Settings row names the provider kind + model when the seam is
	// on (raw service vocabulary, like the gateway-mode row).
	asrDetail := ""
	if h.asrOn() {
		asrDetail = h.deps.ASR.Describe()
	}
	// The agent row (M17) renders only when the textagent seam is wired;
	// the mode probe is the identity lane's bounded, fail-open ask — an
	// unreachable agent still shows the row naming its home.
	agentOn, agentMode := h.deps.TextAgent != nil, ""
	if agentOn {
		ctx, cancel := context.WithTimeout(r.Context(), agentModeProbeTimeout)
		if reviews, err := h.deps.TextAgent.Reviews(ctx); err == nil {
			agentMode = reviews.Mode
		}
		cancel()
	}
	// The push row + device list render only when a transport is live
	// (the chat-lane pattern; nil-safe for hand-composed test Deps). The
	// wake address never leaves the store.
	pushOn, pushTransports := false, ""
	var pushDevices []views.PushDeviceRow
	if h.deps.Push != nil && h.deps.Push.Enabled() {
		pushOn = true
		pushTransports = h.deps.Push.Describe()
		if h.deps.PushDevices != nil {
			registered, err := h.deps.PushDevices.List(r.Context(), sess.Extension)
			if err != nil {
				return nil, err
			}
			pushDevices = make([]views.PushDeviceRow, 0, len(registered))
			for _, device := range registered {
				pushDevices = append(pushDevices, views.PushDeviceRow{
					ID:        device.DeviceID,
					Transport: device.Transport,
					UpdatedAt: time.Unix(0, device.UpdatedAt),
				})
			}
		}
	}
	// The CRM row (T7) is the bounded-probe tri-state: not configured
	// (no seam), connected (probe 200), sign-in failed (401/403 — rotate
	// the token credential), and the fail-open "configured but the probe
	// did not answer" for transport failure or budget exhaustion — four
	// renderings because three cannot say that honestly. Same posture as
	// the agent row's mode probe: bounded, fail-open, never blocking.
	crmState := settingsCRMUnset
	if h.deps.CRM != nil && h.deps.CRM.Enabled() {
		crmState = settingsCRMSetup
		ctx, cancel := context.WithTimeout(r.Context(), crmSettingsProbeTimeout)
		err := h.deps.CRM.HealthCheck(ctx)
		cancel()
		switch {
		case err == nil:
			crmState = settingsCRMConnected
		case errorfamily.Code(err) == crm.HealthCodeUnauthorized:
			crmState = settingsCRMAuthFailed
		}
	}
	return views.SettingsPanel(views.SettingsPanelProps{
		SIPDomain:      h.deps.Config.SIPDomain,
		WebsocketURL:   websocketURL,
		GatewayMode:    string(h.deps.Config.Gateway.Mode),
		PhoneAPI:       h.deps.PhoneAPI.Enabled(),
		CRMState:       crmState,
		ASR:            h.asrOn(),
		ASRDetail:      asrDetail,
		ASRLocality:    h.deps.ASRLocality,
		Agent:          agentOn,
		AgentMode:      agentMode,
		ChatMatrix:     h.deps.Config.Matrix.Enabled(),
		ChatSignal:     h.deps.Config.Signal.Enabled(),
		Push:           pushOn,
		PushTransports: pushTransports,
		PushDevices:    pushDevices,
		ICEServers:     len(h.deps.Config.ICEServers),
		SharedContacts: len(h.deps.Shared),
		Snippets:       snippets,
		AgentAuditRows: h.agentAuditRows(r, sess),
		Passkeys:       h.passkeySectionProps(sess.Extension),
		Lang:           h.lang(r),
	}), nil
}

// agentAuditRows reads the durable agent-action audit window for the
// Settings history (D5): the owner's newest rows, actor resolved to
// the login email when the row names one and the line itself otherwise
// (the multi-actor substrate — the row says WHO, not just which line).
// A nil store or a read failure degrades to no section input: the
// journal line remains the audit truth, the panel never lies with a
// half list.
func (h *handlers) agentAuditRows(r *http.Request, sess session.Session) []views.AgentAuditRow {
	if h.deps.AgentAudit == nil {
		return nil
	}
	rows, err := h.deps.AgentAudit.Recent(r.Context(), sess.Extension, 20)
	if err != nil {
		slog.Warn("settings: agent audit rows unavailable", "error", err)
		return nil
	}
	out := make([]views.AgentAuditRow, 0, len(rows))
	for _, row := range rows {
		actor := row.Actor
		if actor == "" {
			actor = sess.Extension.String()
		}
		out = append(out, views.AgentAuditRow{
			Actor:   actor,
			Event:   row.Event,
			Subject: row.Subject,
			Outcome: row.Outcome,
			At:      time.Unix(row.At, 0),
		})
	}
	return out
}

// renderTab adapts the builder result for the shell.
func (h *handlers) tabComponent(r *http.Request, tab views.Tab, sess session.Session) (templ.Component, error) {
	switch tab {
	case views.TabMessages:
		// A thread deep link (GET /messages/{id} — the URL the row
		// anchors push) opens the conversation, not the list: the
		// pushed URL must round-trip through a refresh or a share.
		if raw := r.PathValue("id"); raw != "" {
			id, err := domain.ParseThreadID(raw)
			if err != nil {
				return threadGone(h.lang(r)), nil
			}
			return h.threadPanel(r, sess, id, threadPageFromQuery(r))
		}
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
		return h.settingsPanel(r, sess)
	}
}

func errorPanel(message string, lang views.Lang) templ.Component {
	return views.ErrorPanel(message, lang)
}

// internalError is the one-home 500 writer (SafeDetail consistency,
// 2026-09-24): the full detail goes to the operator log via
// errorfamily.LogErrorContext (message "op: detail", attrs family/code/
// retryable/exit_code; a transient classification logs Warn — the
// library's self-healing semantics) while the client body carries only
// the family default message — internal store errors (SQL text, file
// paths) never reach the browser. Sub-500 writers keep their own
// texts: client-caused 4xx detail is safe by construction and the
// webhook bodies are byte-stable-pinned for providers.
func (h *handlers) internalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	errorfamily.LogErrorContext(r.Context(), fmt.Errorf("%s: %w", op, err), nil)
	http.Error(w, op+": "+cqrshtmx.SafeDetail(err, http.StatusInternalServerError, false), http.StatusInternalServerError)
}

func csrfToken(r *http.Request) string {
	return httputil.CSRFTokenFromRequest(r)
}
