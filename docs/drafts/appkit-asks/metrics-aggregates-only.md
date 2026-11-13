> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation (the webphone go-appkit fleet-frame re-review, owner-directed), not from a direct filing request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

The opt-in Prometheus surface labels every series with `{method, route, status}` (`metrics.go:31-32`, sha `ed79e4f`), route from `r.Pattern` with an `"unmatched"` fallback (`metrics.go:149-151`). The module's own comment names the leak class (`metrics.go:46`):

> Unauthenticated metrics endpoints leak request cardinality, route ...

The current answer is auth-fencing the endpoint. For multi-tenant or privacy-sensitive deployments that is not enough: the labels themselves are the problem, regardless of who can scrape. webphone is the concrete case — its `/metrics` is loopback-fenced by the NixOS front, and a repo test still greps the exposition for path/extension-shaped labels and fails the build if any appear (aggregates-only is a pinned privacy line: `webphone_http_requests_total{status_class}` + the latency histogram, deliberately no route/method/extension label). Such a service cannot adopt core's metrics surface at all today; there is no knob in `MetricsConfig` (only `Path` + basic-auth fields, `metrics.go:51-63`).

## Ask

An aggregates-only mode on `MetricsConfig`, e.g.:

```go
cfg.Metrics = &appkit.MetricsConfig{
    AggregatesOnly: true, // fold {method, route, status} -> {status_class} (or status only)
}
```

Series names stay the stable contract; the mode only drops/reduces labels. A route-attributed mode and an aggregates mode can coexist behind the existing config — consumers choose per deployment threat model.

## Why it generalizes

Any fleet service fronting customer traffic has the same call to make: per-route histograms are gold on an internal ops box and a data-minimization liability on a tenant-facing one. One knob keeps both postures inside the framework instead of pushing privacy-sensitive deployments back to hand-rolled `/metrics` handlers (webphone's `httputil.Metrics` wiring today).

Cross-links: sibling asks #59 (strict readiness route), #61 (success-status override) — together they are webphone's remaining framework-adoption walls (evidence: webphone `docs/review/libs/go-appkit.md`, 2026-11-12 re-review).

💘 Generated with Crush
