// Package server wires every HTTP surface of the webphone: the templ
// shell and tab partials, the island's assets and /config.js, the session
// API, the phone-api proxy, the message/fax actions, the inbound webhooks,
// and the SSE feed. cqrs-htmx supplies the embedded HTMX scripts and the
// SSE broadcaster; httputil supplies security headers and CSRF.
package server

import (
	"log/slog"
	"net/http"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

// contentSecurityPolicy mirrors the strict posture of the static-site
// era: everything same-origin, SIP over wss, no CDN, no webfonts.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data:; media-src 'self'; connect-src 'self' wss:; " +
	"frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

// Deps are the wired services the handlers ride on.
type Deps struct {
	Config     config.Config
	Sessions   *session.Store
	Messages   *store.Messages
	Faxes      *store.Faxes
	Contacts   *store.Contacts
	Messaging  *messaging.Service
	Fax        *fax.Service
	PhoneAPI   *pbx.Client
	Broadcaster *cqrshtmx.Broadcaster
	Shared     []SharedContactEntry
}

// SharedContactEntry mirrors domain.SharedContact without the handlers
// importing both.
type SharedContactEntry struct {
	Name   string
	Number string
}

// New builds the full http.Handler.
func New(deps Deps) http.Handler {
	h := &handlers{deps: deps}

	// The CSRF-protected surface: pages, partials, tab actions, the
	// session API and the phone-api proxy. Assets, SSE, webhooks and
	// health live outside it (GET-only or secret-authed).
	csrf := httputil.CSRFMiddleware(httputil.CSRFConfig{})

	protected := http.NewServeMux()
	protected.HandleFunc("GET /", h.page)
	protected.HandleFunc("GET /partials/messages", h.partialMessages)
	protected.HandleFunc("GET /partials/messages/{id}", h.partialThread)
	protected.HandleFunc("GET /partials/fax", h.partialFax)
	protected.HandleFunc("GET /partials/voicemail", h.partialVoicemail)
	protected.HandleFunc("GET /partials/history", h.partialHistory)
	protected.HandleFunc("GET /partials/contacts", h.partialContacts)
	protected.HandleFunc("GET /partials/settings", h.partialSettings)
	protected.HandleFunc("POST /messages/send", h.sendMessage)
	protected.HandleFunc("POST /fax/send", h.sendFax)
	protected.HandleFunc("GET /fax/{id}/document", h.faxDocument)
	protected.HandleFunc("GET /attachments/{id}", h.attachment)
	protected.HandleFunc("POST /voicemail/delete", h.deleteVoicemail)
	protected.HandleFunc("POST /contacts/save", h.saveContact)
	protected.HandleFunc("POST /contacts/delete", h.deleteContact)
	protected.HandleFunc("POST /api/session", h.createSession)
	protected.HandleFunc("DELETE /api/session", h.destroySession)
	protected.Handle("/phone-api/", h.sessionGate(http.HandlerFunc(h.proxyPhoneAPI)))

	open := http.NewServeMux()
	open.Handle("/htmx.min.js", cqrshtmx.HTMXScriptHandler())
	open.Handle("/htmx-ext/sse.js", cqrshtmx.HTMXExtensionHandler("sse"))
	open.Handle("/assets/", h.assets())
	open.HandleFunc("GET /config.js", h.configJS)
	open.HandleFunc("GET /events", h.events)
	open.HandleFunc("GET /healthz", h.healthz)
	open.Handle("/hooks/", h.secretGate(http.HandlerFunc(h.webhooks)))

	root := http.NewServeMux()
	root.Handle("/", csrf(protected))
	root.Handle("/htmx.min.js", open)
	root.Handle("/htmx-ext/sse.js", open)
	root.Handle("/assets/", open)
	root.Handle("/config.js", open)
	root.Handle("/events", open)
	root.Handle("/healthz", open)
	root.Handle("/hooks/", open)

	security := httputil.SecurityHeaders(httputil.SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: contentSecurityPolicy,
	})

	return security(recovery(root))
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic in handler", "path", r.URL.Path, "panic", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *handlers) sessionGate(next http.Handler) http.Handler {
	return h.deps.Sessions.Middleware(next)
}

func (h *handlers) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
