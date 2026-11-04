# httputil — library utilization audit

**Verdict:** configured, not defaulted — the security surface runs on this library by explicit choice. One standing body-cap gap. **Adoption 85/100 (very high)** — 8.5 of 10 applicable rows (inputs below).
**Pin:** v1.8.0 (go.mod:31) + `httputil/server_timing` v1.0.3 · **Latest:** v1.8.0 / v1.0.3 · **Currency:** current · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (82/100 at v1.2.0 — its finding #1, `MaxBodySize`, is **still open**); server_timing has its own [report](httputil-server-timing.md).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/httputil` (root + `httpspec` subpackage; `server_timing` separate module) |
| Pinned | v1.8.0 direct; server_timing v1.0.3 direct |
| Latest release | v1.8.0 (local checkout tag) |
| Usage | 11 files import root; ~40% of the root API touched; 8 keyed rate limiters; server lifecycle, CSRF, CSP, nonce, metrics all ride it |

## What the library offers

Root package: `NewServer`/`ServerConfig` (timeouts, graceful shutdown, `StartTLS`, listener variants); middleware family — `SecurityHeaders`, `Nonce` (+ CSP builders), `CSRFMiddleware` (+ full nosurf configurator, token formatters, `InvalidateCSRFCookie`), `CORS`, `Compression`, `Metrics`, `RequestID`, `Language`, `Logging`, `Recovery`, `Timeout`, `MaxBodySize`, `KeyedRateLimiterMiddleware`; `KeyedRateLimiter` (Check/Middleware/ActiveKeys + key extractors + trusted-proxy set); response recorder + `Chain`/`Compose`/`MiddlewareStack`; health handlers; clientip; queryparam helpers; `httpspec` conformance suite (specs for tracing/language/CORS/rate-limit/private-network). `server_timing` and `go-etag/server` are sibling modules.

## How webphone uses it

- **Server lifecycle**: `httputil.NewServer(ServerConfig{Addr, ReadHeaderTimeout: 5s, IdleTimeout: 60s, ShutdownTimeout: 30s})` (cmd/webphone/main.go:197-202) — deliberately NO Read/Write deadline (SSE streams); this replaced the rejected cqrs-htmx `setup` bundle's shell (AGENTS.md: "the shell VALUE landed via httputil.NewServer").
- **Middleware chain** (internal/server/server.go:704-737): `cqrshtmx.ContextEnrichmentMiddleware(nil)` → `httputil.Metrics` (feeds the latency histogram) → [trace] → `cqrshtmx.RequestLoggingSlog` → `servertiming.ServerTimingMiddlewareWhen(env)` → `httputil.SecurityHeaders(securityHeadersConfig())` → `cqrshtmx.RecoveryMiddleware` → routes.
- **Security headers fully explicit** (server.go:85-93): nosniff, `X-Frame-Options: DENY`, Referrer-Policy, the strict same-origin CSP (`connect-src 'self' wss:`), `Permissions-Policy: microphone=(self)…`. Login page + `/health` override CSP per-subtree (nonce'd variants), test-pinned.
- **CSRF**: `CSRFMiddleware` + `CSRFConfig` (trusted origins/proxies), `InvalidateCSRFCookie` at login/logout rotation (internal/server/session_api.go:218,355), `CSRFTokenFromRequest` for the island's token reader (internal/server/api/csrf_api.go:25).
- **Rate limiting**: `NewKeyedRateLimiter` ×8 (login 30/min, hooks 60, events, csrf, transcribe, passkey, contacts, switch 6/min) via factory at server.go:112; custom `remoteHostKey` extractor (port-stripped peer host) deliberately instead of `KeyExtractorFromClientIP` (flip rule documented at server.go:127).
- **Nonce lane**: `httputil.Nonce` + `NonceFromRequest` feed both the login page and the health dashboard (internal/server/login_page.go:43, internal/app/dashboard.go:72,88-90).
- **Conformance**: `httpspec.Run` pins the middleware chain against 19 specs (internal/server/httpspec_test.go:16).

## Gap analysis

| # | Capability area | Status | Evidence / reason |
|---|---|---|---|
| 1 | Server lifecycle (`NewServer` + graceful drain) | 🟢 Fully leveraged | cmd/webphone/main.go:197 |
| 2 | SecurityHeaders, fully explicit config | 🟢 Fully leveraged | server.go:85-93; `TestServedPageSatisfiesStrictCSP` |
| 3 | CSRF suite | 🟢 Fully leveraged | middleware + rotate + token reader |
| 4 | KeyedRateLimiter | 🟢 Fully leveraged | 8 limiters + custom extractor + documented IP-flip rule |
| 5 | Nonce (CSP `script-src 'nonce-…'`) | 🟢 Fully leveraged | login + dashboard subtrees |
| 6 | Metrics hook | 🟢 Fully leveraged | feeds `webphone_http_request_duration_seconds` |
| 7 | server_timing middleware | 🟢 Fully leveraged | see [server_timing report](httputil-server-timing.md) |
| 8 | httpspec conformance | 🟢 Fully leveraged | 19-spec chain pin |
| 9 | **`MaxBodySize`** | 🔴 Missed | **Standing finding since the 2026-09-24 audit: still open.** 12 files hand-roll `http.MaxBytesReader` (internal/server/actions.go:652, transcribe.go:54, passkey_api.go:90,203, hooks/webhooks.go:425, api/*.go 4 KiB JSON guards); `rg MaxBodySize` → 0 uses. Mechanical migration |
| 10 | Compression | 🟡 Partial | assets.go hand-rolls *scoped* gzip for embedded assets by design (SSE-safety + ETag interaction); the generic `Compression` middleware is unused |
| n/a | CORS | ⚪ N/A | same-origin product; CSP forbids cross-origin by policy |
| n/a | Language extractor | ⚪ N/A | the `wp-lang` cookie → Accept-Language chain is session-scoped and SSE-aware (views/i18n.go); the generic extractor cannot express it |
| n/a | RequestID | ⚪ N/A | `cqrshtmx.ContextEnrichmentMiddleware` owns request enrichment |
| n/a | Timeout middleware | ⚪ N/A | long-lived SSE + `/phone-api` proxy need per-route budgets, not a blanket wrapper |
| n/a | `MiddlewareStack`/`Compose` | ⚪ N/A | chain composed via `cqrshtmx.Chain` deliberately (cqrs-htmx round-3 ruling) |

**Score:** rows 1-8 = 1.0 (8.0) + row 10 = 0.5 + row 9 = 0 → **8.5 / 10 = 85/100, very high band.**

## Ruling: should we use it to the max?

Yes with one action item: **retire the 12 hand-rolled `MaxBytesReader` sites for `httputil.MaxBodySize`** (or justify a written ruling at each site). It is the last standing finding from the 09-24 audit and the library ships exactly this. The remaining unused middleware (CORS, Language, RequestID, Timeout) each carry a written no-fit reason above — re-litigate only if the constraint changes.

**Refresh trigger:** a new body-consuming route; a cross-origin consumer; a per-route timeout need that outgrows handler-local contexts.

## How to verify

```console
rg -l "MaxBytesReader" --type go internal/ | wc -l        # expect 12
rg -n "MaxBodySize" --type go internal/ | wc -l           # expect 0
rg -n "httputil.NewServer" cmd/webphone/main.go           # lifecycle
go test ./internal/server/ -run TestHttpspec -v           # conformance pin
```
