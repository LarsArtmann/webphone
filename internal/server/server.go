// Package server wires every HTTP surface of the webphone: the templ
// shell and tab partials, the island's assets and /config.js, the session
// API, the phone-api proxy, the message/fax actions, the inbound webhook
// ingest (internal/server/hooks), and the SSE feed. cqrs-htmx supplies
// the embedded HTMX scripts and the SSE broadcaster; httputil supplies
// security headers and CSRF.
package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	loginpage "github.com/larsartmann/cqrs-htmx/loginpage/v4"
	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-health"
	"github.com/larsartmann/httputil"
	servertiming "github.com/larsartmann/httputil/server_timing"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/server/api"
	"github.com/larsartmann/webphone/internal/server/hooks"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/textagent"
	"github.com/larsartmann/webphone/internal/userauth"
)

// contentSecurityPolicy mirrors the strict posture of the static-site
// era: everything same-origin, SIP over wss, no CDN, no webfonts — and
// NO inline scripts: the theme preload ships as a same-origin file
// (/assets/theme-preload.js, rendered by layout.templ's headExtras) and
// templ-components Base is told NoThemeScript, so script-src needs no
// hash and a dependency bump can never silently change served script
// bytes. TestServedPageSatisfiesStrictCSP fails if any inline script
// appears.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; style-src 'self'; " +
	"img-src 'self' data:; media-src 'self'; connect-src 'self' wss:; " +
	"frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

// Rate limits for the two flood-sensitive surfaces, per client IP:
// login attempts (password guessing) and inbound webhooks (provider
// floods share one source behind the stack's proxy). Contacts saves
// get the hook-grade budget: the legacy-import path bursts one POST
// per row, so the burst must cover a realistic import while still
// fencing runaway write loops. Windows are one minute; httputil
// computes Retry-After from the window instead of a hardcoded guess.
const (
	loginLimit    = 30
	loginBurst    = 5
	hookLimit     = 60
	hookBurst     = 60
	contactsLimit = 60
	contactsBurst = 60
)

// securityHeadersConfig is the single source for the security-header
// posture: New wires it for production and the middleware parity test
// reuses it, so the two can never drift. Permissions-Policy is the
// calibrated complement of the library's RecommendedPermissionsPolicy
// (which denies microphone and would kill the WebRTC phone): everything
// power-adjacent is denied, the microphone stays self-origin only.
func securityHeadersConfig() httputil.SecurityHeadersConfig {
	return httputil.SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: contentSecurityPolicy,
		PermissionsPolicy:     "microphone=(self), camera=(), display-capture=(), geolocation=(), payment=(), usb=()",
	}
}

// csrfSecureFromOrigins derives the CSRF cookie's Secure flag from the
// configured trusted origins: any https origin means the deployment is
// fronted by TLS, so the cookie must never ride a plaintext hop. Plain
// http origins (loopback dev) keep the flag off.
func csrfSecureFromOrigins(origins []string) bool {
	for _, origin := range origins {
		if strings.HasPrefix(origin, "https://") {
			return true
		}
	}
	return false
}

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
// branching-flow large-struct (baselined): THE DI surface — one field
// per optional seam, nil-optionality documented per field (AGENTS.md
// composition doctrine).
type Deps struct {
	Config   config.Config
	Sessions session.Store
	Messages *store.Messages
	Faxes    *readmodels.Faxes
	// Contacts is the metaengine-backed personal-contacts read model
	// (readmodels.db) — same method contract the SQL store held.
	Contacts *readmodels.Contacts
	// Snippets is the metaengine-backed reply-snippet read model
	// (readmodels.db) — same method contract the SQL store held.
	Snippets *readmodels.Snippets
	// Transcripts persists live-call transcript segments (owner-scoped,
	// rendered in the History tab). Nil (test compositions): the save
	// endpoint answers 503 and the History section stays empty.
	Transcripts *readmodels.Transcripts
	// Notes is the post-call-note store (owner-scoped, the History
	// tab's note affordance). Nil (test compositions): the endpoints
	// answer 503.
	Notes     *readmodels.Notes
	Messaging *messaging.Service
	Fax       *fax.Service
	PhoneAPI  *pbx.Client
	Hubs      *ExtensionHubs
	Shared    []domain.SharedContact
	// CRM is the OPTIONAL Ledger CRM integration (name enrichment + call
	// logging). Nil or disabled: every surface renders raw numbers and the
	// island's call-log POST is a no-op.
	CRM *crm.Resolver
	// ASR is the OPTIONAL speech-to-text seam behind /api/transcribe.
	// Nil or disabled: the route answers the styled 404 and the island
	// hides every transcribe affordance. The concrete kind (OpenAI-
	// compatible or Google Cloud STT v2) is chosen in the composition
	// root; the server only ever sees the seam.
	ASR asr.Provider
	// ASRLocality is the config-honest trust label for the ASR seam
	// ("local" | "remote", asr.LocalityOf): the caption-strip footer chip
	// and the Settings row name where caller audio is processed. Empty
	// when the seam is off.
	ASRLocality string
	// TextAgent is the OPTIONAL supervised text-agent seam (the stack's
	// services.telephony.textagent review API on loopback). Nil: the
	// drafts routes answer the styled 404 and the Messages tab renders
	// neither the drafts panel nor the AI compose buttons.
	TextAgent *textagent.Client
	// Probe is the go-health probe serving /livez and /startupz. When
	// nil (tests composing Deps by hand), New falls back to building the
	// equivalent NewChecks probe over the same backing checks — one
	// fallback home, identical wire behavior.
	Probe *health.Probe
	// Dashboard mounts the go-health-dashboard subtree at /health when
	// non-nil (config-gated in the composition root). Nil: no /health
	// surface exists, the styled 404 answers.
	Dashboard http.Handler
	// Dedup is the OPTIONAL engine-backed dedup window behind the
	// webhook and call idem gates (metaengine DedupStore, adapted in
	// the composition root). Nil: the in-process map — byte-identical
	// legacy behavior, dedup memory lost on restart (tests, rollback).
	Dedup hooks.DedupWindow
	// UserAuth is the OPTIONAL embedded passkey identity layer
	// (config-gated). Nil: the passkey mode is off — every passkey
	// surface 404s and the login page renders the extension form only.
	UserAuth *userauth.Service
	// ProjectionStatus overrides the projection-status source behind the
	// passkey-gated "projections" readiness check (nil: the wired
	// UserAuth's own usermgmt service reports — the only prod source).
	// The override exists for tests, mirroring the Probe nil-fallback
	// idiom: the wiring's honesty is pinnable without a real failed
	// projectionhost.
	ProjectionStatus cqrshtmx.ProjectionStatusProvider
	// SSO is the OPTIONAL OpenID Connect login seam (config-gated). Nil:
	// the OIDC mode is off — the begin/callback routes 404 and the login
	// card renders no SSO entry.
	SSO SSOClient
	// Trace is the OPTIONAL flight-recorder seam (operator opt-in, default
	// off). Nil: no request observation and no snapshot endpoint. Non-nil:
	// every request is reported and GET /internal/flightrecorder writes a
	// trace window (loopback only).
	Trace TraceRecorder
	// Push is the OPTIONAL push fanout behind ring-while-closed (the
	// /hooks/ring lane and the /api/push-devices registry). Nil or
	// disabled: the device API answers 404 and the ring hook
	// acknowledges inertly.
	Push *push.Fanout
	// PushDevices is the registered-device store (readmodels-backed;
	// present whenever the composition built the engine).
	PushDevices *readmodels.PushDevices
	// Readiness probes only (healthz): the SQLite handle for the ping
	// and the blob files root for the write probe. Nothing else may use
	// them — data access rides the services above.
	DB       *sql.DB
	BlobRoot string
}

// New builds the full http.Handler.
func New(deps Deps) http.Handler {
	// Session lifetime policy: the sliding idle window plus the absolute
	// cap. The config validates Max >= Idle; the normalized() guard is
	// the belt-and-suspenders for Deps-built instances (tests) whose
	// config skipped validation — they degrade to the pre-sliding
	// behavior instead of renewing against a zero cap.
	lifetime := session.Lifetime{Idle: deps.Config.SessionTTL, Max: deps.Config.SessionMaxTTL}
	// The CSRF-protected surface: pages, partials, tab actions, the
	// session API and the phone-api proxy. Assets, SSE, webhooks and
	// health live outside it (GET-only or secret-authed). The CSRF
	// config carries the fronting deployment shape: without it, a
	// browser behind the TLS-terminating proxy sends Origin https://host,
	// which the plain-HTTP listener reads as a forged same-origin
	// attestation and rejects — every POST (logins included) would 403.
	csrfCfg := httputil.CSRFConfig{
		TrustedProxies: deps.Config.CSRF.TrustedProxies,
		TrustedOrigins: deps.Config.CSRF.TrustedOrigins,
		Secure:         csrfSecureFromOrigins(deps.Config.CSRF.TrustedOrigins),
	}
	// The session cookie rides the same TLS-deployment signal the CSRF
	// cookie derives: behind the TLS-terminating proxy r.TLS is always
	// nil, so Secure must come from the wired policy, not the request.
	cookies := session.CookiePolicy{Secure: csrfCfg.Secure}
	// The library passkey login page over the identity layer. Our config
	// is constant (fixed paths, valid CSS URL), so a build failure is a
	// wiring defect, not an operator state: log it loudly and leave the
	// handler nil — /login then answers the styled 404 instead of a
	// half-built login surface.
	var loginPage *loginpage.Handler
	if deps.UserAuth != nil {
		lp, err := newLoginPage(deps)
		if err != nil {
			slog.Error("webphone: login page build failed; /login answers the styled 404", "error", err)
		} else {
			loginPage = lp
		}
	}
	h := &handlers{
		deps:              deps,
		loginLimiter:      newKeyedRateLimiter(loginLimit, loginBurst),
		hookLimiter:       newKeyedRateLimiter(hookLimit, hookBurst),
		eventsLimiter:     newKeyedRateLimiter(hookLimit, hookBurst),
		csrfLimiter:       newKeyedRateLimiter(hookLimit, hookBurst),
		contactsLimiter:   newKeyedRateLimiter(contactsLimit, contactsBurst),
		transcribeLimiter: newKeyedRateLimiter(hookLimit, hookBurst),
		passkeyLimiter:    newKeyedRateLimiter(hookLimit, hookBurst),
		loginPageHandler:  loginPage,
		cookies:           cookies,
		csrfCfg:           csrfCfg,
		unread:            newUnreadCache(5 * time.Second),
		latency:           newLatencyRecorder(),
	}
	hooksHandlers := hooks.New(hooks.Deps{
		Secret:    deps.Config.Gateway.WebhookSecret,
		Messaging: deps.Messaging,
		Fax:       deps.Fax,
		Unread:    h.unread,
		Dedup:     deps.Dedup,
		Push:      deps.Push,
		Devices:   deps.PushDevices,
	})
	// The carved island JSON surface (internal/server/api): the hub
	// implements the narrow Server seam (session gate + contacts nudge)
	// so the 401 wording and the SSE push keep their one home here.
	apiHandlers := api.New(api.Deps{
		Server:      h,
		Contacts:    deps.Contacts,
		Shared:      deps.Shared,
		CRM:         deps.CRM,
		CallsIdem:   hooks.NewGate(deps.Dedup, hooks.CallIdempotencyTTL),
		ASR:         deps.ASR,
		Transcripts: deps.Transcripts,
		Notes:       deps.Notes,
		TextAgent:   deps.TextAgent,
		Push:        deps.Push,
		Devices:     deps.PushDevices,
	})

	csrf := httputil.CSRFMiddleware(csrfCfg)

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
	protected.HandleFunc("GET /partials/nav", h.partialNav)
	protected.HandleFunc("POST /messages/{id}/read", h.markThreadRead)
	// M22: one pattern route covers pin|archive|mute — the handler's
	// action map rejects anything else with 404, and the literal "read"
	// pattern above still wins for that exact segment.
	protected.HandleFunc("POST /messages/{id}/{flag}", h.setThreadFlag)
	protected.HandleFunc("POST /snippets/save", h.saveSnippet)
	protected.HandleFunc("POST /snippets/delete", h.deleteSnippet)
	protected.HandleFunc("POST /messages/send", h.sendMessage)
	protected.HandleFunc("GET /partials/messages/drafts", h.partialDrafts)
	protected.HandleFunc("POST /messages/drafts/mode", h.setDraftMode)
	protected.HandleFunc("POST /messages/drafts/{id}/approve", h.approveDraft)
	protected.HandleFunc("POST /messages/drafts/{id}/discard", h.discardDraft)
	protected.HandleFunc("POST /messages/ai-draft", h.aiDraft)
	protected.HandleFunc("POST /fax/send", h.sendFax)
	protected.HandleFunc("POST /fax/{id}/resend", h.resendFax)
	protected.HandleFunc("GET /fax/{id}/document", h.faxDocument)
	protected.HandleFunc("GET /attachments/{id}", h.attachment)
	protected.HandleFunc("POST /voicemail/delete", h.deleteVoicemail)
	protected.HandleFunc("POST /contacts/save", h.saveContact)
	protected.HandleFunc("POST /contacts/delete", h.deleteContact)
	protected.HandleFunc("POST /contacts/import", h.importContacts)
	protected.HandleFunc("GET /contacts/export", h.exportContacts)
	protected.Handle("POST /api/session", h.loginLimiter.Middleware()(http.HandlerFunc(h.createSession)))
	// The passkey (WebAuthn) login + enrollment surface: registered ONLY
	// when the identity layer is wired (config-gated in the composition
	// root) — a disabled deployment keeps answering the styled 404.
	// Login ceremonies share the login flood budget (email probing is
	// password-guessing class); enrollment rides its own bucket (the
	// one-time token is the real gate, the bucket only fences runaway
	// clients).
	if deps.UserAuth != nil {
		// The library login page's front door. The nonce middleware
		// mints the per-request CSP nonce; loginPage overrides the
		// app-wide strict CSP for this surface alone (the library's
		// inline ceremony scripts carry the nonce). Anonymous surface:
		// the handler redirects a live session to the shell.
		protected.Handle("GET /login",
			httputil.Nonce(httputil.NonceConfig{})(http.HandlerFunc(h.loginPage)))
		// The passkey login ceremonies at the usermgmt-canonical paths
		// the library page's JS drives (usermgmt's pinned wire
		// contract; the handlers are webphone's session-minting own).
		// Login ceremonies share the login flood budget (email probing
		// is password-guessing class); enrollment rides its own bucket
		// (the one-time token is the real gate, the bucket only fences
		// runaway clients).
		protected.Handle("POST /api/auth/webauthn/login/begin",
			h.loginLimiter.Middleware()(http.HandlerFunc(h.passkeyBeginLogin)))
		protected.Handle("POST /api/auth/webauthn/login/finish",
			h.loginLimiter.Middleware()(http.HandlerFunc(h.passkeyFinishLogin)))
		protected.Handle("POST /api/auth/passkey/enroll/verify",
			h.passkeyLimiter.Middleware()(http.HandlerFunc(h.passkeyEnrollVerify)))
		protected.Handle("POST /api/auth/passkey/enroll/begin",
			h.passkeyLimiter.Middleware()(http.HandlerFunc(h.passkeyEnrollBegin)))
		protected.Handle("POST /api/auth/passkey/enroll/finish",
			h.passkeyLimiter.Middleware()(http.HandlerFunc(h.passkeyEnrollFinish)))
		protected.HandleFunc("GET /enroll", h.enrollPage)
		// Settings-tab passkey management (A2): list lives in the
		// settings partial; revoke is session-gated + CSRF-protected and
		// owner-matched against the session extension's mapping.
		protected.HandleFunc("DELETE /api/passkeys/{id}", h.revokePasskey)
	}
	// The OIDC/SSO login surface (Pocket ID): registered ONLY when the
	// seam is wired. GET-only navigations (the island's SSO link and the
	// IdP's redirect back), so they share the login flood budget and the
	// CSRF wrapper's safe-method pass-through — no form, no token.
	if deps.SSO != nil {
		protected.Handle("GET /api/auth/oauth2/begin",
			h.loginLimiter.Middleware()(http.HandlerFunc(h.ssoBegin)))
		protected.Handle("GET /api/auth/oauth2/callback",
			h.loginLimiter.Middleware()(http.HandlerFunc(h.ssoCallback)))
	}
	// GET is the island's boot resume: a live cookie gets its SIP
	// credentials back and the page opens signed-in, no form. Read-only
	// (no CSRF surface), self-gated through RequireSession.
	protected.HandleFunc("GET /api/session", h.getSession)
	protected.HandleFunc("DELETE /api/session", h.destroySession)
	// JSON surface for the island's contact panel (same store as the
	// Contacts tab; the island migrated off localStorage onto this).
	protected.HandleFunc("GET /api/contacts", apiHandlers.ListContacts)
	protected.Handle("POST /api/contacts", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.SaveContact)))
	protected.HandleFunc("DELETE /api/contacts", apiHandlers.DeleteContact)
	// The per-extension data export: one zip with every thread, fax job
	// and personal contact the signed-in extension owns (settings-tab
	// download link). GET-only and read-only, so the session cookie alone
	// is the right gate — no CSRF surface.
	protected.HandleFunc("GET /api/export", h.exportData)
	// The island's post-call report to the CRM integration: session-gated,
	// CSRF via authedFetch, sharing the contacts write budget (same class:
	// one island JSON POST per user action). A disabled CRM answers 204 —
	// the island normally gates on PBX_CONFIG.crm, but a config change
	// under a long-lived session must not error.
	protected.Handle("POST /api/calls", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.LogCall)))
	// The post-call summarize chain: one call's transcript becomes an AI
	// follow-up draft in the review queue. Session-gated, CSRF via
	// authedFetch, sharing the contacts write budget (one island POST per
	// user action). A disabled seam answers 404.
	protected.Handle("POST /api/calls/{id}/summary-draft", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.SummaryDraft)))
	protected.Handle("GET /api/calls/{id}/note", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.GetCallNote)))
	protected.Handle("PUT /api/calls/{id}/note", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.SetCallNote)))
	// The operator's post-call disposition verdict (issue #6): the one
	// authority that may claim "answered", persisted on the call's
	// review slot and corrected upstream in the same awaited write.
	// Session-gated, CSRF via authedFetch, sharing the contacts write
	// budget (one island POST per user action).
	protected.Handle("POST /api/calls/{id}/disposition", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.ConfirmCallDisposition)))
	// The push-device registry the native shells ride (the Android call
	// shell first): session-gated, CSRF via the same authedFetch dance,
	// sharing the contacts write budget (the same one-write-per-action
	// class). A disabled seam answers 404 — a shell must not believe
	// registration succeeded on a server that cannot ring it.
	protected.HandleFunc("GET /api/push-devices", apiHandlers.ListPushDevices)
	protected.Handle("POST /api/push-devices", h.contactsLimiter.Middleware()(http.HandlerFunc(apiHandlers.SavePushDevice)))
	protected.HandleFunc("DELETE /api/push-devices", apiHandlers.DeletePushDevice)
	// The transcription surface: session-gated, CSRF via authedFetch, and
	// the ONE endpoint behind every audio stream's transcribe affordance
	// (live calls, voicemail, MMS audio). A disabled seam answers 404.
	protected.Handle("POST /api/transcribe", h.transcribeLimiter.Middleware()(http.HandlerFunc(apiHandlers.Transcribe)))
	protected.Handle("POST /api/transcripts", h.transcribeLimiter.Middleware()(http.HandlerFunc(apiHandlers.SaveTranscript)))
	protected.Handle("DELETE /api/transcripts", h.transcribeLimiter.Middleware()(http.HandlerFunc(apiHandlers.DeleteTranscript)))
	// GET /api/csrf shares the flood budget: the endpoint hands out masked
	// tokens anonymously, so a client must not churn it unbounded. One
	// per-peer-host bucket (60/min burst 60) is orders of magnitude above
	// real traffic (one fetch per login rotation).
	protected.Handle("GET /api/csrf", h.csrfLimiter.Middleware()(http.HandlerFunc(apiHandlers.RefreshCSRF)))
	protected.Handle("/phone-api/", session.Require(h.deps.Sessions, http.HandlerFunc(h.proxyPhoneAPI), lifetime, cookies))
	// Unknown paths render the styled 404 (shell + error panel), not Go's
	// bare-text default — the catch-all sits inside the CSRF layer so the
	// response shape matches every other full page.
	protected.HandleFunc("/", h.notFoundPage)

	open := http.NewServeMux()
	open.Handle("/htmx.min.js", cqrshtmx.HTMXScriptHandler())
	// One request serves both extensions (cqrshtmx.HTMXExtensionsHandler
	// concatenates sse + idiomorph with per-extension version comments and a
	// composite ETag) — the page loads exactly one script tag instead of two.
	open.Handle("/htmx-ext.js", cqrshtmx.HTMXExtensionsHandler(cqrshtmx.HTMXExtSSE, cqrshtmx.HTMXExtIdiomorph))
	open.Handle("/assets/", h.assets())
	open.HandleFunc("GET /config.js", h.configJS)
	open.HandleFunc("GET /favicon.svg", h.favicon)
	// Browsers and hard-coded scrapers ask for /favicon.ico when no icon
	// link is parsed; the modern SVG answers both names (content-type
	// carries the format, and every current browser sniffs it fine).
	open.HandleFunc("GET /favicon.ico", h.favicon)
	// PWA-LITE install manifest (2026-10-23 decision: owner demand signal
	// unlocked the parked manifest; the service-worker NOT-DO of the
	// 2026-09-22 spike stands — no fetch interception, zero staleness
	// risk). application/manifest+json is the type Chrome requires to
	// parse it as a manifest.
	open.HandleFunc("GET /manifest.webmanifest", h.manifest)
	// /metrics is the ops scrape surface: AGGREGATES only (counts +
	// build facts, never per-extension data), fenced per-location by
	// the module's vhost like the probe triple.
	open.HandleFunc("GET /metrics", h.metrics)
	// GET /events is rate-limited like the other unauthenticated-by-secret
	// surfaces: a reconnecting tab (or a broken client) must not churn
	// unlimited streams. One bucket per peer host reuses the hook budget
	// (60/min burst 60) — generous for real tabs, bounded for churn.
	open.Handle("GET /events",
		h.eventsLimiter.Middleware()(session.Require(h.deps.Sessions, http.HandlerFunc(h.events), lifetime, cookies)))
	// readiness replaces the old constant-"ok" healthz: the endpoint now
	// tells the truth about the backing resources the app needs. Each
	// check is bounded (see boundedCheck) so a hung resource degrades the
	// probe to 503 within the budget instead of hanging the prober.
	// Per-check timeouts ride the library (v4.11.0's NamedCheck.Timeout):
	// an overdue check is reported failed with "<name>: timed out after
	// <n>" and the probe answers 503 instead of hanging the prober — the
	// same contract the local boundedCheck guard used to provide (F1 of
	// the DI/health review, now upstream).
	//
	// The passkey identity layer joins ONLY when it is wired: a
	// deployment that never enabled the mode keeps probing exactly the
	// two resources it has, and one that did gets an honest 503 naming
	// "userauth" when usermgmt.db breaks (instead of unexplained passkey
	// 503s with a green healthz).
	checks := []cqrshtmx.NamedCheck{
		{Name: "sqlite", Check: deps.DB.Ping, Timeout: checkTimeout},
		{Name: "blob-dir", Check: func() error { return blob.ProbeWrite(deps.BlobRoot) }, Timeout: checkTimeout},
	}
	if deps.UserAuth != nil {
		checks = append(checks, cqrshtmx.NamedCheck{
			Name:    "userauth",
			Check:   func() error { return deps.UserAuth.HealthCheck(context.Background()) },
			Timeout: checkTimeout,
		})
		// The projectionhost behind the identity layer joins the same
		// truth: a worker still draining its journal (async startup) or
		// exhausted past its restart budget degrades healthz naming
		// "projections" (and the failed worker's name in the error
		// text) instead of green-lighting traffic into a passkey lane
		// that cannot answer logins. Library check, bounded like the
		// rest; the provider defaults to the wired service — only
		// tests override it.
		provider := deps.ProjectionStatus
		if provider == nil {
			provider = deps.UserAuth.Usermgmt()
		}
		projectionCheck := cqrshtmx.ProjectionReadinessCheck(provider)
		projectionCheck.Timeout = checkTimeout
		checks = append(checks, projectionCheck)
	}
	readiness := cqrshtmx.ReadinessHandler(checks...)
	open.Handle("GET /healthz", readiness)

	// Liveness + startup complete the standard probe triple WITHOUT a second
	// readiness truth: /healthz above stays the continuous readiness gate,
	// while go-health serves /livez (fetch-free process liveness) and
	// /startupz (503 until the backing resources first pass, then latched).
	// The composition root passes the container-built probe in Deps.Probe;
	// the nil fallback (tests) builds the equivalent NewChecks probe over
	// the same backing checks — same truth, different construction.
	// JSON only: no scripts, CSP-neutral. GET-open like /healthz: probers
	// need no session and the bodies carry check names/statuses only,
	// never secrets.
	selfHealth := deps.Probe
	if selfHealth == nil {
		healthChecks := map[string]health.CheckFunc{
			"sqlite": func(ctx context.Context) error { return deps.DB.PingContext(ctx) },
			"blob-dir": func(_ context.Context) error {
				return blob.ProbeWrite(deps.BlobRoot)
			},
		}
		if deps.UserAuth != nil {
			// Same conditional as the readiness handler above: the
			// identity layer probes only when it is wired. It stays
			// NON-critical on this surface — a broken usermgmt.db degrades
			// one login mode, it must not flap liveness or hold the
			// startup latch.
			healthChecks["userauth"] = deps.UserAuth.HealthCheck
		}
		if deps.ASR != nil && deps.ASR.Enabled() {
			// The optional ASR leg joins the same way: NON-critical,
			// outcomes-based (never probes the provider), only when the
			// seam exists. The prod probe builds this via the named
			// "asr" service; this fallback covers test compositions.
			if hc, ok := deps.ASR.(interface {
				HealthCheck(context.Context) error
			}); ok {
				healthChecks["asr"] = hc.HealthCheck
			}
		}
		selfHealth = health.NewChecks(healthChecks,
			health.WithCriticalServices("sqlite", "blob-dir"), health.WithRefreshInterval(0))
	}
	open.Handle("GET /livez", selfHealth.LivenessHandler())
	open.Handle("GET /startupz", selfHealth.StartupHandler())
	// The dashboard subtree (config-gated upstream): its own mux carries
	// /health, /health/sse, the embedded Datastar SDK and its namespaced
	// probe aliases. Both patterns delegate: Go's mux matches "/health"
	// exactly and "/health/" as the subtree — one without the other
	// would 404 either the page or everything under it. Not registered
	// when Deps.Dashboard is nil.
	if deps.Dashboard != nil {
		open.Handle("/health", deps.Dashboard)
		// The subtree pattern exists for /health/sse and friends; the
		// dashboard answers its own page at "/health" only, so the bare
		// "/health/" would fall through to the stdlib's text 404. Send it
		// to the canonical form instead, like any directory index.
		open.Handle("/health/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health/" {
				http.Redirect(w, r, "/health", http.StatusMovedPermanently)
				return
			}
			deps.Dashboard.ServeHTTP(w, r)
		}))
	} else {
		// Dashboard off: BOTH patterns must still answer with the styled
		// page. Without this "/health" reached the protected catch-all
		// (styled) while "/health/" fell through the open mux into the
		// stdlib's bare-text default, so one absent feature had two
		// different 404s depending on a trailing slash.
		open.HandleFunc("/health", h.notFoundPage)
		open.HandleFunc("/health/", h.notFoundPage)
	}
	open.Handle("GET /version", versionHandler())
	open.HandleFunc("GET /openapi.json", openapiHandler)
	open.Handle("/hooks/", h.hookLimiter.Middleware()(hooksHandlers.SecretGate(http.HandlerFunc(hooksHandlers.Webhooks))))
	// The internal directory endpoint (the stack's FS directory bridge is
	// the only consumer): outside CSRF/session (webhooks-class — secret
	// gated), hook-grade flood budget, loopback-only posture in-handler.
	open.Handle("GET /internal/directory/{extension}",
		h.hookLimiter.Middleware()(h.sipDirectoryGate(http.HandlerFunc(h.internalDirectory))))
	// The flight-recorder snapshot endpoint: machine-facing, registered
	// unconditionally so the disabled seam answers the styled 404 rather
	// than falling through the open mux; loopback-only in-handler.
	open.HandleFunc("GET /internal/flightrecorder", h.traceSnapshot)

	root := http.NewServeMux()
	root.Handle("/", session.Attach(h.deps.Sessions, csrf(protected), lifetime, cookies))
	root.Handle("/htmx.min.js", open)
	root.Handle("/htmx-ext.js", open)
	root.Handle("/assets/", open)
	root.Handle("/config.js", open)
	root.Handle("/events", open)
	root.Handle("/healthz", open)
	root.Handle("/livez", open)
	root.Handle("/startupz", open)
	root.Handle("/health", open)
	root.Handle("/health/", open)
	root.Handle("/version", open)
	root.Handle("/openapi.json", open)
	root.Handle("/hooks/", open)
	root.Handle("/internal/directory/", open)
	root.Handle("/internal/flightrecorder", open)
	root.Handle("/favicon.svg", open)
	root.Handle("/favicon.ico", open)
	root.Handle("/manifest.webmanifest", open)
	root.Handle("/metrics", open)

	security := httputil.SecurityHeaders(securityHeadersConfig())

	// Enrichment outermost: it only wraps the request context (RequestID,
	// echoed back as X-Request-ID) and writes nothing, so the request log
	// inside it still sees every status written anywhere below — while
	// gaining request_id/correlation_id fields from the enriched context.
	// The user extractor stays nil on purpose: the library's user identity
	// is a ULID from the rejected usermgmt module, and extensions are not
	// ULIDs — forcing them in would misuse the concept.
	// Server-Timing stays an operator opt-in (WEBPHONE_DEBUG_TIMING), the
	// same gate the removed hand-rolled writer had. The library middleware
	// auto-prepends total;dur, sanitizes values against CRLF injection,
	// and its writer forwards Flush/Hijack (SSE keeps flushing). Read the
	// env once at construction like the old middleware did.
	timingEnabled := os.Getenv("WEBPHONE_DEBUG_TIMING") != ""
	// Latency sits directly under enrichment and above the request log:
	// it measures the same span the log line reports, for every status,
	// and its ResponseRecorder forwards Flush/Hijack so /events and the
	// phone-api streaming keep working unchanged. Trace is conditional: an
	// opt-out deployment carries no extra wrapper at all.
	middlewares := []func(http.Handler) http.Handler{
		cqrshtmx.ContextEnrichmentMiddleware(nil),
		httputil.Metrics(httputil.MetricsConfig{Recorder: h.latency}),
	}
	if deps.Trace != nil {
		middlewares = append(middlewares, traceMiddleware(deps.Trace))
	}
	middlewares = append(middlewares,
		cqrshtmx.RequestLoggingSlog(slog.Default()),
		servertiming.ServerTimingMiddlewareWhen(func(*http.Request) bool { return timingEnabled }),
		security,
		cqrshtmx.RecoveryMiddleware,
	)
	return cqrshtmx.Chain(middlewares...)(root)
}

// checkTimeout bounds every /healthz named check (NamedCheck.Timeout): a
// hung backing resource must degrade the probe to 503 within this budget
// instead of hanging the prober. The overdue call keeps running — the
// timeout bounds the probe's wait, it does not cancel the check.
const checkTimeout = 2 * time.Second

// buildVersion is injected at link time by the flake build
// (-X ...internal/server.buildVersion=<version>), so /version reports the
// released version instead of Go's "(devel)" for source builds. Empty in
// `go build`/`go test` runs — the handler then falls back to build info.
var buildVersion string

// buildCommit and buildCommitDate are link-time injections too (the
// flake passes self.rev/self.dirtyRev and self.lastModifiedDate) —
// commit-time facts ONLY, never a build clock, so the byte-reproducibility
// the aarch64 ELF assert depends on stays honest. Empty in plain go
// builds; the handler falls back to ReadBuildInfo's vcs settings, which
// carry the same facts for `go build` runs inside the repo.
var (
	buildCommit     string
	buildCommitDate string
)

// DisplayVersion names the running build for operator surfaces
// (/version, /metrics).
func DisplayVersion() string {
	info, ok := debug.ReadBuildInfo()
	return DisplayVersionWith(info, ok)
}

// DisplayVersionWith resolves the version against an existing build
// info: the ldflags injection wins, a real module version follows,
// source builds stay "devel".
func DisplayVersionWith(info *debug.BuildInfo, ok bool) string {
	switch {
	case buildVersion != "":
		return buildVersion
	case ok && info.Main.Version != "" && info.Main.Version != "(devel)":
		return info.Main.Version
	default:
		return "devel"
	}
}

// vcsSetting reads one build-info settings pair (vcs.revision,
// vcs.time, vcs.modified) — that is where go build embeds VCS facts.
func vcsSetting(info *debug.BuildInfo, key string) string {
	if info == nil {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}

// buildCommitWith resolves the source revision: the ldflags injection
// wins, build-info vcs settings follow (suffixed -dirty when modified),
// unknown builds report "".
func buildCommitWith(info *debug.BuildInfo, ok bool) string {
	if buildCommit != "" {
		return buildCommit
	}
	if !ok {
		return ""
	}
	if rev := vcsSetting(info, "vcs.revision"); rev != "" {
		if vcsSetting(info, "vcs.modified") == "true" {
			return rev + "-dirty"
		}
		return rev
	}
	return ""
}

// buildCommitDateWith resolves the commit date: ldflags injection
// (RFC 3339) first, build-info vcs time second, "" otherwise.
func buildCommitDateWith(info *debug.BuildInfo, ok bool) string {
	if buildCommitDate != "" {
		return buildCommitDate
	}
	if !ok {
		return ""
	}
	if raw := vcsSetting(info, "vcs.time"); raw != "" {
		if stamp, err := time.Parse(time.RFC3339, raw); err == nil {
			return stamp.UTC().Format(time.RFC3339)
		}
		return raw
	}
	return ""
}

// versionHandler reports build metadata for the operator's curl one-liner
// (library DebugHandler pattern): module version, source revision +
// commit date (when known — kills the store-path chain-verification
// dance), Go version, module path. Captured at construction time — it is
// build info, not live state.
func versionHandler() http.HandlerFunc {
	info, ok := debug.ReadBuildInfo()
	version := DisplayVersionWith(info, ok)
	goVersion := runtime.Version()
	title := "webphone"
	if ok {
		title = info.Main.Path
	}
	payload := map[string]any{
		"version":   version,
		"goVersion": goVersion,
		"title":     title,
	}
	if commit := buildCommitWith(info, ok); commit != "" {
		payload["commit"] = commit
	}
	if date := buildCommitDateWith(info, ok); date != "" {
		payload["commitDate"] = date
	}
	return cqrshtmx.DebugHandler(payload)
}
