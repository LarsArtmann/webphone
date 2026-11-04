# httputil/server_timing — library utilization audit

**Verdict:** wired exactly as designed (env-gated edge middleware on an SSE-safe writer); named spans unused. **Adoption 67/100 (moderate)** — 2 of 3 applicable rows.
**Pin:** v1.0.3 (go.mod:32) · **Latest:** v1.0.3 · **Currency:** current · **Status:** direct dep (separate module of the httputil repo)
**Audited:** 2026-11-04 · **Prior reports:** covered inside [httputil](httputil.md) and the [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/httputil/server_timing` (zero-dep module) |
| Pinned | v1.0.3 direct |
| Latest release | v1.0.3 |
| Usage | 1 import site (internal/server/server.go:25), env-gated |

## What the library offers

`ServerTiming` collector (`NewServerTiming`, `Enabled` nil-safe, `Record(name,desc,dur)`, `Measure(name) func()`, `HeaderValue`); context plumbing (`WithServerTiming`/`RecordServerTiming`/`MeasureServerTiming`); SSE-safe response writer `WrapServerTiming` (forwards Flush/Hijack); `ServerTimingMiddleware()` and `ServerTimingMiddlewareWhen(pred)` (nil pred = disabled). Renders spec-compliant `Server-Timing` headers with auto `total;dur`.

## How webphone uses it

- `servertiming.ServerTimingMiddlewareWhen(timingEnabled)` in the middleware chain (internal/server/server.go:733), gated by `WEBPHONE_DEBUG_TIMING` env — off by default, so production responses carry no Server-Timing header.
- Because the middleware wraps the response writer SSE-safely, `/events` streaming is unaffected when the gate is on.

## Gap analysis

| # | Capability area | Status | Evidence / reason |
|---|---|---|---|
| 1 | Edge middleware (env-gated) | 🟢 Fully leveraged | server.go:733; nil-pred-disabled semantics used correctly |
| 2 | SSE-safe writer integration | 🟢 Fully leveraged | inherited by the chain; SSE compatibility is why this module was chosen over hand-rolling |
| 3 | **Named spans** (`MeasureServerTiming` in hot handlers: SSE broadcast, sqlite apply, gateway forward) | 🔴 Missed | Only the auto `total` is produced; no handler records a named span, so a debug session sees one number, not a breakdown |

**Score:** rows 1-2 = 1.0 (2.0) + row 3 = 0 → **2.0 / 3 = 67/100, moderate band.**

## Ruling

The seam is small and the default-off gate is right for a phone app. Named spans are worth adding **only** when a latency investigation actually needs them (the Prometheus histogram already owns production latency visibility; Server-Timing is a per-request debug tool). Do not pre-instrument.

**Refresh trigger:** a latency investigation where the histogram's aggregates cannot localize the cost; then instrument the suspected 2-3 spans behind the same env gate.

## How to verify

```console
rg -n "servertiming" internal/server/server.go | head -3     # middleware + import
rg -n "MeasureServerTiming|RecordServerTiming" --type go | wc -l   # expect 0
WEBPHONE_DEBUG_TIMING=1 go test ./internal/server/ -run TestServerTiming -v 2>/dev/null || rg -n "WEBPHONE_DEBUG_TIMING" internal/server/
```
