# Health probes — fleet options and the dashboard-HTML/CSP tradeoff

> Recorded 2026-09-19 (SUPERB honest-self-health execution, T14) · Status:
> RECORDED, NOT EXECUTED — adoption decisions belong to the owner.

## Context

webphone v2.x ships three JSON probe endpoints: `/healthz` (readiness —
cqrshtmx.ReadinessHandler, bounded checks), `/livez` (process liveness) and
`/startupz` (boot latch) from go-health's container-free `NewChecks`
constructor. This memo records the two follow-on options that deliberately
did NOT ship.

## Option A — fleet federation (stack-level, zero webphone change)

go-health's `federation` package (v0.3.0) pulls N remote go-health instances
over HTTP into one merged response: `federation.New(remotes []Remote, ...)
(*Prober, error)` (verified: federation/federation.go:131), namespacing checks
as `name/check` and surfacing an unreachable remote as a synthetic
`name/reachable` FAIL check (fail-closed, never silent).

The consuming stack (`nix-international-telephony`) could run a tiny
federation service scraping `pbx.artmann.tech/livez`, `/startupz` (and any
future instances) and expose one merged surface — a fleet health view without
touching webphone. Alternatively go-health-dashboard's aggregate/federation
handlers serve the same merge as HTTP.

**Effort:** stack-side service + module wiring. **Not done because:** no
operator need expressed yet; webphone's single-instance probes already answer
the runbook's checks.

## Option B — the go-health-dashboard HTML face in webphone

The dashboard (go-health-dashboard v0.9.x) serves a real-time HTML view at
`/health` (Datastar SSE) beside JSON-only kubelet probes. The HTML face
requires CSP `script-src 'unsafe-eval'` because the Datastar SDK evaluates
code from SSE patches (verified: dashboard csp.go:19 `// 'unsafe-eval' is
required because the Datastar SDK`, csp.go:33 appends it).

webphone's CSP is `script-src 'self'` + one exact hash for the framework's
inert theme-preload script (verified: internal/server/server.go security
config) — relaxing it would undo the CSP posture pinned by
`TestServedPageSatisfiesStrictCSP` for one optional view.

**What would have to change to adopt it:** either (a) the owner relaxes the
CSP stance explicitly (accept `unsafe-eval` on a dedicated path — CSP is
response-header based, not per-path, so this means weakening the whole
origin), or (b) go-health-dashboard ships a nonce/CSP-safe Datastar mode
upstream, or (c) webphone renders the dashboard's JSON into its own
server-rendered partial (HTMX polling, no Datastar) — cheap but re-implements
the live face. Until one of those happens the JSON probes are the whole
adopted surface, by decision.

## Related

- Review: `2026-09-19_18-49_samber-do-di-health-service-orientation.md`
  (F1/F2/F3 findings)
- Plan: `docs/planning/2026-09-19_20-01_SUPERB-honest-self-health-upstream-first-plan.md`
- ROADMAP long-shots reference this memo.
