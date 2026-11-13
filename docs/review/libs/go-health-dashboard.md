# go-health-dashboard — library utilization audit

**Verdict:** deeply and carefully wired — the stack-facing operator surface rides this library with its CSP, nonce and stylesheet conventions handled properly. **Adoption 88/100 (very high)** — 3.5 of 4 applicable rows.
**Pin:** v0.13.2 (go.mod, bumped 2026-11-05) · **Latest:** v0.13.3 · **Currency:** behind by one patch, re-verified 2026-11-13 — the v0.13.2→v0.13.3 delta is the **`data-nonce` root bridge** (upstream issue #23: `rootNonceAttrs` stamps the CSP nonce onto the `<html>` root so the Datastar SDK evaluates its expressions WITHOUT the Function path a strict CSP forbids) + CSP/concurrency/lifecycle test hardening + release infra; its dependency floor is templ v0.3.1100 + templ-components v1.27.0 + go-health v0.10.0 — **exactly webphone's current pins**, so the bump is zero-collision and rides the next train (or, better, the `/health` unsafe-eval→nonce-lane retirement it is the prerequisite for) · **Status:** direct dep. History: the v0.13.0 release commit was pushed in the upstream repo with its TAG missing (the release dance stalled mid-flight) — the webphone closeout train verified the release commit green in a detached worktree, tagged v0.13.0, and rode it the same day (the coordinated go-health v0.9.0 wave's consumer side). v0.13.0 → v0.13.2 landed 2026-11-05 in the templ-wave train (v0.13.2 requires templ v0.3.1100 + templ-components v1.27.0 + go-health v0.10.0 — exactly the wave); the delta is release-infra fixes (release-cut commits/tags before the battery, rollback on red) + docs (compatibility policy, push-fingerprint stability contract), no API change.
**Audited:** 2026-11-04 · **Prior reports:** none specific (context: 2026-09-24 fleet stack audit; AGENTS.md § health dashboard seam).

## Status

|                |                                                                                             |
| -------------- | ------------------------------------------------------------------------------------------- |
| Module         | `github.com/larsartmann/go-health-dashboard`                                                |
| Pinned         | v0.13.2 direct (bumped 2026-11-05)                                                          |
| Latest release | v0.13.3 (2026-11-13 re-verify; webphone pin v0.13.2 — zero-collision bump pending, see Currency) |
| Usage          | internal/app/dashboard.go (construction), internal/server/server.go:641-668 (mount + fence) |

## What the library offers

`New(probe, opts...)` over a go-health `Prober`; Datastar SSE pusher (real-time check patches, trend sparkline, `WithTrend(n)`); routes (`WithRoutes`) for dashboard/SSE/trend/export/probe aliases/datastar.js; `WithCSSPath` (fallback: Tailwind Play CDN), `WithEmbeddedDatastarSDK`, `RecommendedCSP(nonce)`; `WithMiddleware` (**auth seam** wrapping HTML/SSE/favicon/metrics but never the kubelet-style probes); `WithRateLimit` (probe endpoints never limited); `WithNonceExtractor`; plus untaken knobs: push intervals/limits/drain, `WithWebhook`, `WithMetrics`, `WithIntrospection`, `WithGrouping`, `WithPublicMode`, `WithBasePath`.

## How webphone uses it

- **Construction** (internal/app/dashboard.go:46-75): `WithTitle`, `WithCSSPath("/assets/health.css")`, `WithEmbeddedDatastarSDK()`, `WithTrend(240)`, `WithRateLimit(60, time.Minute)`, and a full custom `WithRoutes` table that aliases probe names under `/health/*` (because webphone owns the root `/healthz`,`/livez`,`/startupz` triple) — **all aliases served from the SAME probe instance**.
- **CSP**: per-request override with `dashboard.RecommendedCSP(nonce)` on the `/health` subtree only (dashboard.go:88-92), fed by `httputil.NonceFromRequest` via `WithNonceExtractor`; `unsafe-eval` scoped to this subtree alone (test-pinned elsewhere-strict).
- **Stylesheet**: scoped `/assets/health.css` built by scripts/build-health-css.sh (Tailwind v4 over the dashboard + templ-components vendor trees) because the CSP blocks the Play CDN fallback.
- **Mount + fence**: private mux + `dash.RegisterRoutes`; dashboard-off (default) renders the styled 404 at both `/health` patterns (server.go:660-668); the STACK fences the surface (SSO `forwardAuth` or loopback) because the library has no auth.
- **Lifecycle**: registered as named `"dashboard"` service (cascade shutdown), `dash.Start(ctx)` in App.Start, `_ do.Shutdowner` asserted.

## Gap analysis

| #   | Capability area                                                                         | Status             | Evidence / reason                                                                                                                                                                                                                                                           |
| --- | --------------------------------------------------------------------------------------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Full wiring (routes, rate limit, trend, datastar, CSS, nonce)                           | 🟢 Fully leveraged | dashboard.go:46-75                                                                                                                                                                                                                                                          |
| 2   | `RecommendedCSP` + scoped stylesheet                                                    | 🟢 Fully leveraged | dashboard.go:88-92; scripts/build-health-css.sh                                                                                                                                                                                                                             |
| 3   | Same-probe aliases (`/health/livez                                                      | readyz             | startupz`)                                                                                                                                                                                                                                                                  |
| 4   | **Auth seam (`WithMiddleware`)**                                                        | 🟡 Partial         | unused: the surface sits on the open mux behind the Caddy fence (SSO forwardAuth / loopback). The seam exists precisely for in-process fencing; the current posture relies on the reverse proxy. A standalone-binary deployment (no Caddy) would leak check names + version |
| n/a | Push knobs (intervals, connection caps, drain)                                          | ⚪ N/A             | defaults fit a single-operator dashboard; no fleet scale                                                                                                                                                                                                                    |
| n/a | `WithWebhook` / `WithMetrics` / `WithIntrospection` / `WithGrouping` / `WithPublicMode` | ⚪ N/A             | webhook = push owned by `internal/push`; metrics owned by `/metrics`; introspection/grouping are fleet-scale conveniences                                                                                                                                                   |

**Score:** rows 1-3 = 1.0 (3.0) + row 4 = 0.5 → **3.5 / 4 = 88/100, very high band.**

## Ruling

Adoption is essentially complete. The one decision to make explicit: either wrap the dashboard mux with an in-process fence via `WithMiddleware` (defense-in-depth for bare-binary deployments), or write the "fence lives at Caddy" ruling into the deployment docs as the contract (it is implied in AGENTS.md today). Cheap either way.

**Refresh trigger:** any deployment of this binary without the Caddy front; operator demand for trend export/introspection; the `/health` nonce-lane retirement (v0.13.3's `data-nonce` bridge is its upstream prerequisite).

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "dashboard.New\(|WithRoutes|WithMiddleware" internal/app/dashboard.go   # WithMiddleware absent
rg -n "RecommendedCSP" internal/app/dashboard.go
rg -n "\"/health\"" internal/server/server.go | head -3
```
