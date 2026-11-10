package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	loginpage "github.com/larsartmann/cqrs-htmx/loginpage/v4"
	"github.com/larsartmann/cqrs-htmx/v4/httperr"
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
	// Transcription uploads (one island/shell POST per audio segment; live
	// call chunks arrive every few seconds): the hook-grade bucket.
	transcribeLimiter *httputil.KeyedRateLimiter
	// Passkey enrollment ceremonies (token-gated by design; the bucket
	// only fences runaway clients, the one-time token is the real gate).
	passkeyLimiter *httputil.KeyedRateLimiter
	// Line switches (D1): session-gated + CSRF-gated, but PBX-bound — a
	// tight bucket keeps a looping client off the directory.
	switchLimiter *httputil.KeyedRateLimiter
	// CSP violation reports: session-free intake (the browser delivers
	// without credentials), so the bucket is the only flood fence.
	cspLimiter *httputil.KeyedRateLimiter
	// Templ form-POST buckets (2026-11-08): the write lanes outside
	// /api/* keyed by cost class — compose (provider/paid surfaces,
	// tight), the contacts-grade forms, and the cheap state toggles.
	// Rejections carry the .wp-error element htmx's responseHandling
	// selects, so a flooded lane is honest inside the tab.
	composeLimiter       *httputil.KeyedRateLimiter
	contactsTemplLimiter *httputil.KeyedRateLimiter
	stateTemplLimiter    *httputil.KeyedRateLimiter
	// loginPageHandler is the library-owned passkey login page
	// (cqrs-htmx/loginpage), built ONLY when the identity layer is wired;
	// nil keeps /login on the styled 404 (mode off, or the page build
	// failed loudly at boot).
	loginPageHandler *loginpage.Handler
	// cookies is the wired session-cookie policy (Secure mirrors the
	// TLS-fronted deployment; r.TLS is nil behind the proxy).
	cookies session.CookiePolicy
	// csrfCfg is the REAL CSRF config — login/logout cookie invalidation
	// must carry the same Secure/Path attributes as the cookie it deletes
	// (strict browsers reject a non-Secure deletion of a Secure cookie).
	csrfCfg httputil.CSRFConfig
	// Memoized nav-badge totals, invalidated on every unread mutation.
	unread *unreadCache
	// Request-duration aggregates for /metrics (the ONE recorder instance:
	// the middleware writes it, the metrics handler reads it).
	latency *latencyRecorder
	// Island latency-beacon histograms for /metrics (the measure-first
	// latency design; built ONLY when metrics.island_beacon is on — nil
	// keeps the beacon route on a terse 404 and /metrics without the
	// family, so the flag-off product stays byte-identical).
	island *islandLatency
}

// page renders the full shell for the root URL.
func (h *handlers) page(w http.ResponseWriter, r *http.Request) {
	h.renderShell(w, r, views.TabMessages)
}

// tabPage renders the full shell for a tab URL (deep links).
func (h *handlers) tabPage(w http.ResponseWriter, r *http.Request) {
	h.renderShell(w, r, tabFromPath(r.URL.Path))
}

func (h *handlers) renderShell(w http.ResponseWriter, r *http.Request, tab views.Tab, content ...templ.Component) {
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
	// The passkey front door renders only when the identity layer is
	// wired; a disabled deployment keeps the extension-only login card.
	props.Passkey = h.deps.Config.Auth.Passkey.Enabled()
	props.SSO = h.deps.SSO != nil
	// The live-call context slot exists only when the CRM seam is on: zero
	// CRM config keeps the served shell byte-identical to a deployment
	// that never had the integration (issue #10).
	props.LiveContext = h.deps.CRM != nil && h.deps.CRM.Enabled()

	if sess, ok := session.From(r.Context()); ok {
		props.SignedIn = sess.Extension.String()
		props.SignedInDID = h.identityFor(sess.Extension)
		// The Admin tab renders only for admin sessions (per-request
		// resolution over the live mapping — see isAdminSession).
		props.Admin = h.isAdminSession(sess)
		// Keep the hub's language fresh so SSE fragments match the tabs.
		h.deps.Hubs.SetLang(sess.Extension, lang)
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
		if len(content) > 0 {
			// A pre-built component (the admin deep link): the handler
			// already ran (and gated) the tab's own builder.
			props.TabContent = content[0]
		} else if component, err := h.tabComponent(r, tab, sess); err == nil {
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
	return httperr.SafeDetail(err, http.StatusInternalServerError, false)
}

// partial renders just the tab region for an HTMX swap.
func (h *handlers) partial(w http.ResponseWriter, r *http.Request, tab views.Tab) {
	sess, ok := h.RequireSession(w, r)
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
	renderHTML(w, r, component)
}

// renderHTML sends a rendered component as the response body. The status
// and headers are already committed when a render fails, so the failure
// can only surface as a plain 500.
func renderHTML(w http.ResponseWriter, r *http.Request, component templ.Component) {
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

func (h *handlers) partialReports(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabReports)
}

func (h *handlers) partialContacts(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabContacts)
}

func (h *handlers) partialSettings(w http.ResponseWriter, r *http.Request) {
	h.partial(w, r, views.TabSettings)
}

// notFoundPage renders unknown paths as the shell around the not-found
// panel: a stray deep link gets the app chrome, the literal 404, the path
// it asked for and a way back, instead of Go's bare "404 page not found"
// text. The status stays 404 so probers, crawlers and the smoke suite's
// stale-token probe all keep reading the truth.
// (Handler-level 500s mid-render keep http.Error by necessity — bytes may
// already be on the wire; store failures render the error panel inside
// the shell via tabComponent's error path.)
func (h *handlers) notFoundPage(w http.ResponseWriter, r *http.Request) {
	lang := h.lang(r)
	props := views.ShellProps{
		ActiveTab:  views.TabMessages,
		Lang:       lang,
		CSRFToken:  csrfToken(r),
		TabContent: views.NotFoundPanel(r.URL.Path, lang),
	}
	props.CSRFHxHeaders, _ = templ.JSONString(map[string]string{"X-CSRF-Token": props.CSRFToken}) //nolint:erraudit // json.Marshal of map[string]string cannot fail
	// The requested path is echoed into the page; keep a crafted one from
	// being cached as a permanent answer under a shared CDN key.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if err := views.Shell(props).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// notFound answers a miss in the shape the caller can actually read.
//
// A document request (a browser navigating to a stale link, or a partial
// URL pasted out of devtools) gets the styled page. An HTMX-originated
// request gets a short notice instead: htmx absorbs non-2xx bodies and
// reports the status through its own toast, so shipping a whole shell
// would be kilobytes nobody renders.
//
// Machine surfaces keep their own terse writers on purpose — JSON APIs,
// provider webhooks and fetched byte streams (/attachments, assets) never
// reach this helper, because a browser is not their client and an HTML
// body would be meaningless to them. A fax document link does reach it:
// it is a browser download, so a stale link deserves the readable page.
func (h *handlers) notFound(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		h.notFoundPage(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, templ.EscapeString(h.T(r, "error.notfound.title"))); err != nil {
		slog.WarnContext(r.Context(), "404 notice write failed", "error", err)
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
		// F19: the signed-in gate lives in NavLinks — an anonymous nav
		// renders zero links — so this partial must carry the session
		// like the full page does or a signed-in re-fetch would wipe the
		// tab row.
		props.SignedIn = sess.Extension.String()
		props.Unread = h.countUnread(r, sess)
		props.NewVoicemail = h.countVoicemail(r, sess)
		props.Admin = h.isAdminSession(sess)
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
	case "/reports", "/reports/":
		return views.TabReports
	case "/contacts", "/contacts/":
		return views.TabContacts
	case "/settings", "/settings/":
		return views.TabSettings
	case "/admin", "/admin/":
		return views.TabAdmin
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
