> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation (the webphone go-appkit fleet-frame re-review, owner-directed), not from a direct filing request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

Readiness grading is per-check only: critical failures fail readiness, non-critical failures degrade to warn (`health/probe.go:54-55`, sha `ed79e4f`):

> Status grading is unchanged: critical failures fail readiness, non-critical
> failures degrade to warn.

That is the right default, but some deployments need a SECOND, stricter readiness surface over the same probe: 503 when ANY check fails, including non-critical ones, while the dashboard keeps the graded view. webphone is the concrete case: `/healthz` is the only surface that reports a revoked CRM token (a deliberately non-critical leg there so the dashboard WARNs instead of flapping), so it must fail on every configured check while the dashboard stays lenient — one probe, two gradings, currently only expressible by hand-rolling a second handler outside the module (webphone's `/healthz` is cqrs-htmx's `ReadinessHandler` for exactly this reason: `internal/server/server.go:673-713`).

## Ask

An option to expose a strict readiness route beside the graded surfaces, e.g.:

```go
appkithealth.New(probe,
    appkithealth.WithProbeRoutes(...),
    appkithealth.WithStrictReadiness("/healthz"), // 503 on ANY failure; graded view unchanged
)
```

(or an equivalent `health.Option` that grades an alternate route all-critical). Mount's default `/healthz /readyz /startupz` set (`health/mount.go:111`) stays as-is; only the strict variant is new.

## Why it generalizes

"Honest readiness" (an all-legs gate for probers) beside a graded operator view is a standing ops pattern, not a webphone quirk — any service with a hard external dependency it deliberately keeps non-critical on the dashboard wants this. Today every such consumer re-rolls a second handler, which is exactly the duplication the composition mission exists to delete.

Cross-links: sibling asks #60 (label-free metrics mode), #61 (success-status override) — together they are webphone's remaining framework-adoption walls (evidence: webphone `docs/review/libs/go-appkit.md`, 2026-11-12 re-review).

💘 Generated with Crush
