// Package server wires every HTTP surface of the webphone: the templ
// shell and tab partials, the island's assets and /config.js, the session
// API, the phone-api proxy, the message/fax actions, the inbound webhooks,
// and the SSE feed. cqrs-htmx supplies the embedded HTMX scripts and the
// SSE broadcaster; httputil supplies security headers and CSRF.
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

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-health"
	"github.com/larsartmann/httputil"
	servertiming "github.com/larsartmann/httputil/server_timing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
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
type Deps struct {
	Config    config.Config
	Sessions  session.Store
	Messages  *store.Messages
	Faxes     *store.Faxes
	Contacts  *store.Contacts
	Messaging *messaging.Service
	Fax       *fax.Service
	PhoneAPI  *pbx.Client
	Hubs      *ExtensionHubs
	Shared    []domain.SharedContact
	// CRM is the OPTIONAL Ledger CRM integration (name enrichment + call
	// logging). Nil or disabled: every surface renders raw numbers and the
	// island's call-log POST is a no-op.
	CRM *crm.Resolver
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
	h := &handlers{
		deps:            deps,
		loginLimiter:    newKeyedRateLimiter(loginLimit, loginBurst),
		hookLimiter:     newKeyedRateLimiter(hookLimit, hookBurst),
		eventsLimiter:   newKeyedRateLimiter(hookLimit, hookBurst),
		csrfLimiter:     newKeyedRateLimiter(hookLimit, hookBurst),
		contactsLimiter: newKeyedRateLimiter(contactsLimit, contactsBurst),
		unread:          newUnreadCache(5 * time.Second),
		hooksIdem:       newIdemStore(hookIdempotencyTTL),
	}

	// The CSRF-protected surface: pages, partials, tab actions, the
	// session API and the phone-api proxy. Assets, SSE, webhooks and
	// health live outside it (GET-only or secret-authed). The CSRF
	// config carries the fronting deployment shape: without it, a
	// browser behind the TLS-terminating proxy sends Origin https://host,
	// which the plain-HTTP listener reads as a forged same-origin
	// attestation and rejects — every POST (logins included) would 403.
	csrf := httputil.CSRFMiddleware(httputil.CSRFConfig{
		TrustedProxies: deps.Config.CSRF.TrustedProxies,
		TrustedOrigins: deps.Config.CSRF.TrustedOrigins,
		Secure:         csrfSecureFromOrigins(deps.Config.CSRF.TrustedOrigins),
	})

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
	// GET is the island's boot resume: a live cookie gets its SIP
	// credentials back and the page opens signed-in, no form. Read-only
	// (no CSRF surface), self-gated through requireSession.
	protected.HandleFunc("GET /api/session", h.getSession)
	protected.HandleFunc("DELETE /api/session", h.destroySession)
	// JSON surface for the island's contact panel (same store as the
	// Contacts tab; the island migrated off localStorage onto this).
	protected.HandleFunc("GET /api/contacts", h.apiListContacts)
	protected.Handle("POST /api/contacts", h.contactsLimiter.Middleware()(http.HandlerFunc(h.apiSaveContact)))
	protected.HandleFunc("DELETE /api/contacts", h.apiDeleteContact)
	// The island's post-call report to the CRM integration: session-gated,
	// CSRF via authedFetch, sharing the contacts write budget (same class:
	// one island JSON POST per user action). A disabled CRM answers 204 —
	// the island normally gates on PBX_CONFIG.crm, but a config change
	// under a long-lived session must not error.
	protected.Handle("POST /api/calls", h.contactsLimiter.Middleware()(http.HandlerFunc(h.apiLogCall)))
	// GET /api/csrf shares the flood budget: the endpoint hands out masked
	// tokens anonymously, so a client must not churn it unbounded. One
	// per-peer-host bucket (60/min burst 60) is orders of magnitude above
	// real traffic (one fetch per login rotation).
	protected.Handle("GET /api/csrf", h.csrfLimiter.Middleware()(http.HandlerFunc(h.refreshCSRF)))
	protected.Handle("/phone-api/", session.Require(h.deps.Sessions, http.HandlerFunc(h.proxyPhoneAPI), lifetime))
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
	// /metrics is the ops scrape surface: AGGREGATES only (counts +
	// build facts, never per-extension data), fenced per-location by
	// the module's vhost like the probe triple.
	open.HandleFunc("GET /metrics", h.metrics)
	// GET /events is rate-limited like the other unauthenticated-by-secret
	// surfaces: a reconnecting tab (or a broken client) must not churn
	// unlimited streams. One bucket per peer host reuses the hook budget
	// (60/min burst 60) — generous for real tabs, bounded for churn.
	open.Handle("GET /events",
		h.eventsLimiter.Middleware()(session.Require(h.deps.Sessions, http.HandlerFunc(h.events), lifetime)))
	// readiness replaces the old constant-"ok" healthz: the endpoint now
	// tells the truth about the two backing resources the app needs. Each
	// check is bounded (see boundedCheck) so a hung resource degrades the
	// probe to 503 within the budget instead of hanging the prober.
	// Per-check timeouts ride the library (v4.11.0's NamedCheck.Timeout):
	// an overdue check is reported failed with "<name>: timed out after
	// <n>" and the probe answers 503 instead of hanging the prober — the
	// same contract the local boundedCheck guard used to provide (F1 of
	// the DI/health review, now upstream).
	readiness := cqrshtmx.ReadinessHandler(
		cqrshtmx.NamedCheck{Name: "sqlite", Check: deps.DB.Ping, Timeout: checkTimeout},
		cqrshtmx.NamedCheck{Name: "blob-dir", Check: func() error { return probeBlobDir(deps.BlobRoot) }, Timeout: checkTimeout},
	)
	open.Handle("GET /healthz", readiness)

	// Liveness + startup complete the standard probe triple WITHOUT a second
	// readiness truth: /healthz above stays the continuous readiness gate,
	// while go-health serves /livez (fetch-free process liveness) and
	// /startupz (503 until the backing resources first pass, then latched).
	// The probe's checks share the same functions /healthz evaluates — same
	// truth, different probe lifecycles. JSON only: no scripts, CSP-neutral.
	// GET-open like /healthz: probers need no session and the bodies carry
	// check names/statuses only, never secrets. No background refresh loop
	// (live mode): two local checks per probe hit are cheap.
	selfHealth := health.NewChecks(map[string]health.CheckFunc{
		"sqlite": func(ctx context.Context) error { return deps.DB.PingContext(ctx) },
		"blob-dir": func(_ context.Context) error {
			return probeBlobDir(deps.BlobRoot)
		},
	}, health.WithCriticalServices("sqlite", "blob-dir"), health.WithRefreshInterval(0))
	open.Handle("GET /livez", selfHealth.LivenessHandler())
	open.Handle("GET /startupz", selfHealth.StartupHandler())
	open.Handle("GET /version", versionHandler())
	open.HandleFunc("GET /openapi.json", openapiHandler)
	open.Handle("/hooks/", h.hookLimiter.Middleware()(h.secretGate(http.HandlerFunc(h.webhooks))))

	root := http.NewServeMux()
	root.Handle("/", session.Attach(h.deps.Sessions, csrf(protected), lifetime))
	root.Handle("/htmx.min.js", open)
	root.Handle("/htmx-ext.js", open)
	root.Handle("/assets/", open)
	root.Handle("/config.js", open)
	root.Handle("/events", open)
	root.Handle("/healthz", open)
	root.Handle("/livez", open)
	root.Handle("/startupz", open)
	root.Handle("/version", open)
	root.Handle("/openapi.json", open)
	root.Handle("/hooks/", open)
	root.Handle("/favicon.svg", open)
	root.Handle("/favicon.ico", open)
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
	return cqrshtmx.Chain(
		cqrshtmx.ContextEnrichmentMiddleware(nil),
		cqrshtmx.RequestLoggingSlog(slog.Default()),
		servertiming.ServerTimingMiddlewareWhen(func(*http.Request) bool { return timingEnabled }),
		security,
		cqrshtmx.RecoveryMiddleware,
	)(root)
}

// checkTimeout bounds every /healthz named check (NamedCheck.Timeout): a
// hung backing resource must degrade the probe to 503 within this budget
// instead of hanging the prober. The overdue call keeps running — the
// timeout bounds the probe's wait, it does not cancel the check.
const checkTimeout = 2 * time.Second

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

// buildVersion is injected at link time by the flake build
// (-X ...internal/server.buildVersion=<version>), so /version reports the
// released version instead of Go's "(devel)" for source builds. Empty in
// `go build`/`go test` runs — the handler then falls back to build info.
var buildVersion string

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

// versionHandler reports build metadata for the operator's curl one-liner
// (library DebugHandler pattern): module version, Go version, module path.
// Captured at construction time — it is build info, not live state.
func versionHandler() http.HandlerFunc {
	info, ok := debug.ReadBuildInfo()
	version := DisplayVersionWith(info, ok)
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

// openapiSpec is the hand-written OpenAPI 3.1 document for the session
// API — the one machine-facing contract worth publishing. The spec is a
// constant: it compiles into the binary and a route test keeps the route
// and its content-type honest.
const openapiSpec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "webphone session API",
    "version": "1.0.0",
    "description": "Server session for the tabs and the phone-api proxy. Credentials are verified against the PBX directory before the session is minted."
  },
  "paths": {
    "/api/session": {
      "post": {
        "operationId": "createSession",
        "summary": "Create a server session (sets the HttpOnly cookie)",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["extension", "password"],
                "properties": {
                  "extension": {"type": "string", "examples": ["1001"]},
                  "password": {"type": "string", "format": "password"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {"description": "Session created"},
          "400": {"description": "Invalid extension or body"},
          "401": {"description": "Credentials rejected by the PBX directory"},
          "502": {"description": "PBX credential verification unavailable"},
          "500": {"description": "Session store failure"}
        }
      },
      "get": {
        "operationId": "resumeSession",
        "summary": "Resume a live session (island boot: returns the SIP credentials for the silent re-register)",
        "responses": {
          "200": {
            "description": "The cookie's live session; the browser re-registers without the login form",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["extension", "password"],
                  "properties": {
                    "extension": {"type": "string", "examples": ["1001"]},
                    "password": {"type": "string", "format": "password"},
                    "did": {"type": "string", "description": "Presented PSTN number (config identities), when known"}
                  }
                }
              }
            }
          },
          "401": {"description": "No live session — the island shows the login form"}
        }
      },
      "delete": {
        "operationId": "destroySession",
        "summary": "Destroy the current session",
        "responses": {
          "200": {"description": "Session destroyed"},
          "401": {"description": "No live session"}
        }
      }
    },
    "/api/csrf": {
      "get": {
        "operationId": "refreshCSRF",
        "summary": "Fetch the fresh masked CSRF token (login rotation invalidated the old one)",
        "responses": {
          "200": {
            "description": "Current masked CSRF token, pairing with the csrf_token cookie",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["token"],
                  "properties": {
                    "token": {"type": "string"}
                  }
                }
              }
            }
          },
          "429": {
            "description": "Per-client flood budget exhausted; retry after the Retry-After seconds",
            "headers": {
              "Retry-After": {"schema": {"type": "integer"}, "description": "Seconds until the bucket refills"}
            }
          }
        }
      }
    },
    "/api/contacts": {
      "get": {
        "operationId": "listContacts",
        "summary": "List the session extension's personal contacts plus the configured shared contacts",
        "responses": {
          "200": {
            "description": "Personal (store-backed, per extension) and shared (config) contacts",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["personal", "shared"],
                  "properties": {
                    "personal": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["id", "name", "number"],
                        "properties": {
                          "id": {"type": "string", "description": "Branded contact id; the LIST is the only drift-free id source (renames keep the old id)"},
                          "name": {"type": "string"},
                          "number": {"type": "string", "examples": ["+4917012345678"]}
                        }
                      }
                    },
                    "shared": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["name", "number"],
                        "properties": {
                          "name": {"type": "string"},
                          "number": {"type": "string"}
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "401": {"description": "No live session"},
          "500": {"description": "Store failure"}
        }
      },
      "post": {
        "operationId": "saveContact",
        "summary": "Upsert one personal contact (a repeated number renames; mutations answer 204 and the list is the id source)",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["number"],
                "properties": {
                  "name": {"type": "string"},
                  "number": {"type": "string", "examples": ["+4917012345678"]}
                }
              }
            }
          }
        },
        "responses": {
          "204": {"description": "Saved (upsert by number within the session's extension)"},
          "400": {"description": "Invalid body (or body over 4 KiB)"},
          "401": {"description": "No live session"},
          "422": {"description": "Number failed dialable validation, or the per-extension list is full (500; renames never count)"},
          "429": {
            "description": "Per-client write flood budget exhausted (generous: the legacy import bursts one POST per row); retry after the Retry-After seconds",
            "headers": {
              "Retry-After": {"schema": {"type": "integer"}, "description": "Seconds until the bucket refills"}
            }
          },
          "500": {"description": "Store failure"}
        }
      },
      "delete": {
        "operationId": "deleteContact",
        "summary": "Delete one personal contact by id (scoped to the session's extension)",
        "parameters": [
          {"name": "id", "in": "query", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {
          "204": {"description": "Deleted"},
          "401": {"description": "No live session"},
          "404": {"description": "Unknown id (or another extension's contact)"}
        }
      }
    },
    "/api/calls": {
      "post": {
        "operationId": "logCall",
        "summary": "Journal one finished call on the CRM contact that owns the number (no-op when the CRM is disabled or the number matches no contact — the integration never creates contacts)",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["number", "direction"],
                "properties": {
                  "number": {"type": "string", "examples": ["+4917012345678"]},
                  "direction": {"type": "string", "enum": ["in", "out"]},
                  "seconds": {"type": "integer", "description": "Duration in seconds; negative values are clamped to 0"},
                  "outcome": {"type": "string", "enum": ["answered", "missed"]}
                }
              }
            }
          }
        },
        "responses": {
          "204": {"description": "Logged, or deliberately dropped (CRM disabled, or the number matches no CRM contact)"},
          "400": {"description": "Malformed body, or direction not in/out"},
          "401": {"description": "No live session"},
          "403": {"description": "CSRF token missing or invalid"},
          "422": {"description": "Number failed dialable validation"},
          "429": {
            "description": "Per-client write flood budget exhausted (shares the contacts budget); retry after the Retry-After seconds",
            "headers": {
              "Retry-After": {"schema": {"type": "integer"}, "description": "Seconds until the bucket refills"}
            }
          },
          "502": {"description": "The CRM could not record the call (unreachable or 5xx) — the island toasts"}
        }
      }
    }
  }
}`

func openapiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/schema+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := w.Write([]byte(openapiSpec)); err != nil {
		slog.WarnContext(r.Context(), "openapi spec write failed mid-response", "error", err)
	}
}
