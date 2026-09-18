// Package server wires every HTTP surface of the webphone: the templ
// shell and tab partials, the island's assets and /config.js, the session
// API, the phone-api proxy, the message/fax actions, the inbound webhooks,
// and the SSE feed. cqrs-htmx supplies the embedded HTMX scripts and the
// SSE broadcaster; httputil supplies security headers and CSRF.
package server

import (
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

// contentSecurityPolicy mirrors the strict posture of the static-site
// era: everything same-origin, SIP over wss, no CDN, no webfonts. The
// one script-src hash pins the inline theme-preload script that
// templ-components layout.Base emits unconditionally (no opt-out knob
// as of v1.18.0, and its logic is inert here: webphone themes via
// data-theme, not the Tailwind dark class). The app.css forced-theme
// color-scheme rules must keep !important so the script's inline
// colorScheme can never win. TestServedPageSatisfiesStrictCSP fails
// when the dependency's script bytes change, forcing a deliberate
// hash refresh.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'sha256-AO4OqWm6Ms8LvRxbwPXsO4Fy1QauZJl6sD4NNzrR7S0='; style-src 'self'; " +
	"img-src 'self' data:; media-src 'self'; connect-src 'self' wss:; " +
	"frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

// Rate limits for the two flood-sensitive surfaces, per client IP:
// login attempts (password guessing) and inbound webhooks (provider
// floods share one source behind the stack's proxy). Windows are one
// minute; httputil computes Retry-After from the window instead of a
// hardcoded guess.
const (
	loginLimit = 30
	loginBurst = 5
	hookLimit  = 60
	hookBurst  = 60
)

// newKeyedRateLimiter builds the httputil keyed limiter webphone uses
// for both flood-sensitive surfaces. MaxKeys stays uncapped here: keys
// are direct-peer hosts, so the map is bounded by the number of proxy
// source addresses, and TTL eviction handles the churn.
func newKeyedRateLimiter(limit, burst uint) *httputil.KeyedRateLimiter {
	return httputil.NewKeyedRateLimiter(httputil.KeyedRateLimiterConfig{
		Limit:        limit,
		Window:       time.Minute,
		Burst:        burst,
		KeyExtractor: remoteHostKey,
	})
}

// remoteHostKey keys the bucket by the direct peer's host. The port MUST
// be stripped: behind the consuming stack's TLS terminator every request
// arrives from the proxy socket with an ephemeral source port, and a
// port-qualified key would hand every request its own bucket — rate
// limiting silently off. The shared secret and the PBX-proven session
// remain the real boundaries; this only throttles flooding.
// Flip rule: switch to httputil.KeyExtractorFromClientIP only once the
// stack proves it sanitizes X-Forwarded-For on these routes.
func remoteHostKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Deps are the wired services the handlers ride on.
type Deps struct {
	Config    config.Config
	Sessions  *session.Store
	Messages  *store.Messages
	Faxes     *store.Faxes
	Contacts  *store.Contacts
	Messaging *messaging.Service
	Fax       *fax.Service
	PhoneAPI  *pbx.Client
	Hubs      *ExtensionHubs
	Shared    []domain.SharedContact
	// Readiness probes only (healthz): the SQLite handle for the ping
	// and the blob files root for the write probe. Nothing else may use
	// them — data access rides the services above.
	DB       *sql.DB
	BlobRoot string
}

// New builds the full http.Handler.
func New(deps Deps) http.Handler {
	h := &handlers{
		deps:          deps,
		loginLimiter:  newKeyedRateLimiter(loginLimit, loginBurst),
		hookLimiter:   newKeyedRateLimiter(hookLimit, hookBurst),
		eventsLimiter: newKeyedRateLimiter(hookLimit, hookBurst),
		unread:        newUnreadCache(5 * time.Second),
	}

	// The CSRF-protected surface: pages, partials, tab actions, the
	// session API and the phone-api proxy. Assets, SSE, webhooks and
	// health live outside it (GET-only or secret-authed).
	csrf := httputil.CSRFMiddleware(httputil.CSRFConfig{})

	protected := http.NewServeMux()
	protected.HandleFunc("GET /{$}", h.page)
	protected.HandleFunc("GET /messages", h.tabPage)
	protected.HandleFunc("GET /messages/{id}", h.tabPage)
	protected.HandleFunc("GET /fax", h.tabPage)
	protected.HandleFunc("GET /voicemail", h.tabPage)
	protected.HandleFunc("GET /history", h.tabPage)
	protected.HandleFunc("GET /contacts", h.tabPage)
	protected.HandleFunc("GET /settings", h.tabPage)
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
	protected.HandleFunc("POST /contacts/import", h.importContacts)
	protected.HandleFunc("GET /contacts/export", h.exportContacts)
	protected.Handle("POST /api/session", h.loginLimiter.Middleware()(http.HandlerFunc(h.createSession)))
	protected.HandleFunc("DELETE /api/session", h.destroySession)
	protected.Handle("/phone-api/", h.deps.Sessions.Require(http.HandlerFunc(h.proxyPhoneAPI)))

	open := http.NewServeMux()
	open.Handle("/htmx.min.js", cqrshtmx.HTMXScriptHandler())
	open.Handle("/htmx-ext/sse.js", cqrshtmx.HTMXExtensionHandler("sse"))
	open.Handle("/assets/", h.assets())
	open.HandleFunc("GET /config.js", h.configJS)
	open.HandleFunc("GET /favicon.svg", h.favicon)
	// GET /events is rate-limited like the other unauthenticated-by-secret
	// surfaces: a reconnecting tab (or a broken client) must not churn
	// unlimited streams. One bucket per peer host reuses the hook budget
	// (60/min burst 60) — generous for real tabs, bounded for churn.
	open.Handle("GET /events",
		h.eventsLimiter.Middleware()(h.deps.Sessions.Require(http.HandlerFunc(h.events))))
	// readiness replaces the old constant-"ok" healthz: the endpoint now
	// tells the truth about the two backing resources the app needs.
	readiness := cqrshtmx.ReadinessHandler(
		cqrshtmx.NewNamedCheck("sqlite", deps.DB.Ping),
		cqrshtmx.NewNamedCheck("blob-dir", func() error { return probeBlobDir(deps.BlobRoot) }),
	)
	open.Handle("GET /healthz", readiness)
	open.Handle("/hooks/", h.hookLimiter.Middleware()(h.secretGate(http.HandlerFunc(h.webhooks))))

	root := http.NewServeMux()
	root.Handle("/", h.deps.Sessions.Attach(csrf(protected)))
	root.Handle("/htmx.min.js", open)
	root.Handle("/htmx-ext/sse.js", open)
	root.Handle("/assets/", open)
	root.Handle("/config.js", open)
	root.Handle("/events", open)
	root.Handle("/healthz", open)
	root.Handle("/hooks/", open)
	root.Handle("/favicon.svg", open)

	security := httputil.SecurityHeaders(httputil.SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: contentSecurityPolicy,
	})

	// requestLog outermost: it sees every status written anywhere below
	// (429s, panics, SSE disconnects) — the server's blind twin of the
	// browser event log, runbook-greppable at 3 a.m.
	requestLog := cqrshtmx.RequestLoggingSlog(slog.Default())

	return requestLog(security(cqrshtmx.RecoveryMiddleware(root)))
}

// probeBlobDir proves the blob store accepts writes: temp file in the
// files root, then remove it. A full disk or a lost mount fails here and
// the operator sees it in /healthz instead of silently losing attachments.
func probeBlobDir(root string) error {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return err
	}
	file, err := os.CreateTemp(root, ".healthz-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return err
	}
	return os.Remove(name)
}

// versionHandler reports build metadata for the operator's curl one-liner
// (library DebugHandler pattern): module version, Go version, module path.
// Captured at construction time — it is build info, not live state.
func versionHandler() http.HandlerFunc {
	info, ok := debug.ReadBuildInfo()
	version := "devel"
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	goVersion := runtime.Version()
	title := "webphone"
	if ok {
		title = info.Main.Path
	}
	return cqrshtmx.DebugHandler(map[string]any{
		"version":   version,
		"goVersion": goVersion,
		"title":     title,
	})
}
