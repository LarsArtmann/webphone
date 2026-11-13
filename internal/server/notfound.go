package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

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
