# go-health — library utilization audit

**Verdict:** exactly the composition the 09-24 audit praised — injector-discovered probe, one readiness truth, zero duplication. The response-sanitization gap **LANDED 2026-11-04** (bumped + the library applies `CoerceValidUTF8` at its own write seam). **Adoption 100/100 (top band, post-bump)** — 6 of 6 applicable rows.
**Pin:** v0.9.0 (go.mod) · **Latest:** v0.9.0 · **Currency:** current · **Status:** direct dep (bumped 2026-11-04 with go-health-dashboard v0.13.0 — the coordinated wave's consumer side; the dashboard's release commit existed but its tag was missing, which the closeout train completed: verified the release commit green in a detached worktree, tagged v0.13.0, pushed, waited for proxy propagation, then bumped both)
**Audited:** 2026-11-04 · **Prior reports:** [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (80/100 at v0.3.0).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-health` |
| Pinned | v0.8.1 direct |
| Latest release | v0.9.0 (local checkout tag) — branded `InstanceID`, `CoerceValidUTF8` (successor of `SanitizeResponse`), `StatusSince` rename |
| Usage | internal/app/app.go (probe construction + lifecycle), internal/server/server.go (livez/startupz fallback) |

## What the library offers

`New(injector, opts...)` — discovers checks via `do.HealthcheckerWithContext` implementations in a samber/do container; `NewChecks(map[string]CheckFunc)` for explicit compositions; `Probe` lifecycle (`Start` refresh loop, `Shutdown`, `MarkShuttingDown`); critical vs non-critical roll-up (`WithCriticalServices` — critical failures fail readiness; others visible but non-gating); handlers `LivenessHandler`/`ReadinessHandler`/`StartupHandler`/`Healthz`/`RegisterRoutes`; evaluation/caching (`Evaluate`, `CachedResponse*` — the dashboard consumes these internally); options (timeouts, refresh, instance ID, version, boot time, evaluation hooks, recorders); sub-packages `aggregate`, `federation`, `checks` (disk/memory batteries).

## How webphone uses it

- **Injector discovery**: `health.New(injector, health.WithCriticalServices("sqlite","blob-dir"), health.WithRefreshInterval(probeRefreshIf(cfg.Dashboard.Enable)))` (internal/app/app.go:367-370) — the container's services implement `do.HealthcheckerWithContext` and are discovered; refresh 1s only while the dashboard is on.
- **Critical pair** `sqlite` + `blob-dir` roll up to readiness; userauth/asr/crm legs are non-critical (visible, non-gating).
- **Lifecycle**: `Probe.Start(ctx)` / `Probe.Shutdown()` ride App.Start/App.Shutdown (app.go:587,615); the probe is deliberately NOT registered in its own injector (self-recursion guard, app.go:105-110) and is threaded to server.Deps + dashboard, never registered in the injector it reads.
- **`/livez` + `/startupz`** from the same probe instance (server.go:639-640, `LivenessHandler`/`StartupHandler`; `health.NewChecks` fallback for test compositions).
- **`/healthz` is deliberately NOT go-health**: it is `cqrshtmx.ReadinessHandler(cqrshtmx.NamedCheck{...})` with per-check timeouts (server.go:561-590) — one readiness truth, split documented in AGENTS.md.

## Gap analysis

| # | Capability area | Status | Evidence / reason |
|---|---|---|---|
| 1 | Injector discovery (`do.HealthcheckerWithContext`) | 🟢 Fully leveraged | app.go:367 |
| 2 | Critical/non-critical roll-up | 🟢 Fully leveraged | `WithCriticalServices("sqlite","blob-dir")` |
| 3 | Probe lifecycle + refresh cadence | 🟢 Fully leveraged | dashboard-on ⇒ 1s, else live-eval |
| 4 | Liveness + startup handlers | 🟢 Fully leveraged | server.go:639-640 |
| 5 | Readiness serving | 🟡 Partial | Deliberately delegated to `cqrshtmx.ReadinessHandler` for per-check timeouts and the projections leg — documented divergence, not neglect; using go-health here too would create a second readiness truth |
| 6 | Probe-eval timeout (`WithTimeout`) | 🟡 Partial | unused; per-check budgets live in cqrs-htmx `NamedCheck`; the probe eval itself runs on defaults |
| 7 | **Response sanitization (`CoerceValidUTF8`, v0.9.0)** | 🟢 Adopted 2026-11-04 | Bumped to v0.9.0; the coercion is applied INSIDE the library at its single serialization seam (handlers.go `writeHealthResponse` → `marshalResponse(CoerceValidUTF8(resp))`), so `/livez`+`/startupz` are fixed with zero webphone-side wiring — webphone never marshals a `health.Response` itself. `/healthz` (cqrs-htmx `ReadinessHandler`) marshals via encoding/json v1, which coerces invalid UTF-8 silently — no v2 500 risk there |
| n/a | Instance ID / version / boot-time observability | ⚪ N/A | single binary; `/version` + bootreport own identity |
| n/a | Evaluation hooks / recorders / `ChainRecorders` | ⚪ N/A | the dashboard subscribes internally; no second recorder use case |
| n/a | `aggregate` / `federation` / `checks` batteries | ⚪ N/A | one process, no federated probes; disk/memory thresholds are a NixOS monitoring concern |

**Score:** rows 1-4 = 1.0 (4.0) + rows 5-6 = 0.5 (1.0) + row 7 = 1.0 (bumped 2026-11-04) → **6.0 / 6 = 100/100, top band.**

## Ruling

**LANDED 2026-11-04:** bumped to v0.9.0 together with go-health-dashboard v0.13.0 (the wave's coordinated consumer side — see the header note). No webphone-side sanitizer wiring exists or is needed: the library applies `CoerceValidUTF8` at its own serialization seam before marshaling, and webphone never marshals a `health.Response` itself. The dashboard half required finishing the wave's stalled release (the v0.13.0 release commit was pushed untagged): verified green in a detached worktree, tagged, proxy-verified, then bumped. Everything else is at its deliberate boundary.

**Refresh trigger:** a second readiness consumer; a multi-instance deployment; the go-health-dashboard gaining a surface that needs `WithEvaluationHook`.

## How to verify

```console
rg -n "health.New\(" internal/app/app.go
rg -n "LivenessHandler|StartupHandler" internal/server/server.go | head -4
rg -n "cqrshtmx.ReadinessHandler" internal/server/server.go | head -2    # the one readiness truth
rg -n "go-health v" go.mod                                               # expect v0.9.0 (bumped 2026-11-04)
```
