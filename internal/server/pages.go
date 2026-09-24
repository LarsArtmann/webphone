package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

// handlers bundles the deps; every file in this package hangs methods off it.
type handlers struct {
	deps Deps
	// Per-client flood protection for the unauthenticated-by-session
	// surfaces: login attempts, inbound webhooks, SSE connects, and the
	// anonymous CSRF token endpoint. Contacts saves carry a session but
	// join the flood budget class: the legacy import bursts one POST
	// per row, so it gets the generous hook-grade bucket, not the tight
	// login one.
	loginLimiter    *httputil.KeyedRateLimiter
	hookLimiter     *httputil.KeyedRateLimiter
	eventsLimiter   *httputil.KeyedRateLimiter
	csrfLimiter     *httputil.KeyedRateLimiter
	contactsLimiter *httputil.KeyedRateLimiter
	// Dedupe memory for replayed provider status callbacks (provider_ref).
	hooksIdem *idemStore
	// Dedupe memory for the island's post-call journal reports (key).
	callsIdem *idemStore
	// Memoized nav-badge totals, invalidated on every unread mutation.
	unread *unreadCache
	// go-health evaluation outcome counters for /metrics (M18).
	health *healthOutcomes
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
	// Raw JSON for the hx-headers attribute: templ HTML-escapes attribute
	// values exactly once on render, so the pre-escaped httputil helper
	// would double-escape here and drop CSRF protection. json.Marshal of a
	// string map cannot fail.
	props.CSRFHxHeaders, _ = templ.JSONString(map[string]string{"X-CSRF-Token": props.CSRFToken}) //nolint:erraudit // json.Marshal of map[string]string cannot fail
	props.Lang = lang

	if sess, ok := session.From(r.Context()); ok {
		props.SignedIn = sess.Extension.String()
		props.SignedInDID = h.identityFor(sess.Extension)
		// Keep the hub's language fresh so SSE fragments match the tabs.
		h.deps.Hubs.SetLang(sess.Extension, lang)
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
		if component, err := h.tabComponent(r, tab, sess); err == nil {
			props.TabContent = component
		} else {
			props.TabContent = errorPanel(h.safeDetail(r, "render tab", err), lang)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Shell(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// identityFor returns the extension's presented PSTN number (config
// identities) for display, or "" when none is configured. Validation at
// config load guarantees keys are normalized extensions, so the lookup
// by the signed-in extension cannot silently miss.
func (h *handlers) identityFor(ext domain.Extension) string {
	return h.deps.Config.Identities[ext.String()]
}

// safeDetail logs the full error for the operator (English op text,
// via errorfamily.LogErrorContext — family/code/retryable attrs) and
// returns the client-safe copy: the family default message, never the
// raw internal detail. Panels that degrade a page keep their own op
// string so the log line names the surface that failed.
func (h *handlers) safeDetail(r *http.Request, op string, err error) string {
	errorfamily.LogErrorContext(r.Context(), fmt.Errorf("%s: %w", op, err), nil)
	return cqrshtmx.SafeDetail(err, http.StatusInternalServerError, false)
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
		_ = errorPanel(h.safeDetail(r, "render tab", err), h.lang(r)).Render(r.Context(), w) //nolint:erraudit // best-effort write; the response is already committed
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

// notFoundPage renders unknown paths as the shell around an error panel —
// error-page parity with the pre-templ-components era: a stray deep link
// gets the app chrome and the reload affordance instead of Go's bare
// "404 page not found" text. The status stays 404 so probers, crawlers
// and the smoke suite's stale-token probe all keep reading the truth.
// (Handler-level 500s mid-render keep http.Error by necessity — bytes may
// already be on the wire; store failures render this same panel inside
// the shell via tabComponent's error path.)
func (h *handlers) notFoundPage(w http.ResponseWriter, r *http.Request) {
	lang := h.lang(r)
	props := views.ShellProps{
		ActiveTab:  views.TabMessages,
		Lang:       lang,
		CSRFToken:  csrfToken(r),
		TabContent: views.ErrorPanel(views.T(lang, "error.notfound"), lang),
	}
	props.CSRFHxHeaders, _ = templ.JSONString(map[string]string{"X-CSRF-Token": props.CSRFToken}) //nolint:erraudit // json.Marshal of map[string]string cannot fail
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if err := views.Shell(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// partialNav re-renders the nav links: labels in the negotiated language,
// badges fresh from the caches. The island's language switch re-fetches
// this partial (shell.js, on wp:lang-changed) so the nav switches language
// together with the tabs — no full reload, the island never unloads.
// Anonymous visitors get labels without badges (the nav is visible pre-login).
func (h *handlers) partialNav(w http.ResponseWriter, r *http.Request) {
	props := views.ShellProps{
		ActiveTab: tabFromPath("/" + r.URL.Query().Get("active")),
		Lang:      h.lang(r),
	}
	if sess, ok := session.From(r.Context()); ok {
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.NavLinks(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
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
