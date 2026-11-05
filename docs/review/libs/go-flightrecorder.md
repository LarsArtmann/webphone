# go-flightrecorder — adjacent library audit (closeout T24)

**Verdict:** the opt-in debugging seam rides the library exactly as designed: one `Recorder` configured with snapshot dir/retention/compression/min-age, `SnapshotIfAsync` off a detached request context, `OnErrorOrLatency(2s)` as the trigger, `Close` drained last via the container's Shutdowner. The server-side `TraceRecorder` interface keeps the framework OUT of the handler packages — the app adapter is the only place that sees the library.
**Pin:** v0.2.1 (direct) · **Latest:** v0.2.1 · **Currency:** current · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).
**Run all commands from the webphone checkout root** (repo-relative paths).

## Usage (verified)

`internal/app/flightrecorder.go`: `WEBPHONE_DEBUG_FLIGHTRECORDER=1` gates construction (boot-failure on error, never silent); `New(WithSnapshotDir, WithMaxSnapshots(20), WithCompression(gzip.BestSpeed), WithMinAge(10s))`; `Observe` builds `TriggerContext{Kind: "http.request", Duration}` with 5xx mapped to a synthetic error, then `SnapshotIfAsync(context.WithoutCancel(ctx), tc, OnErrorOrLatency(2*time.Second))` — the snapshot outlives the response by design. `GET /internal/flightrecorder` writes a window on demand (loopback-only, machine surface). Shutdown drains via `do.Shutdowner`.

**Unused surface (ruled as-is):** `Stop`/`Enabled` (lifecycle is constructor+Close here), the direct `Snapshot`/`SnapshotToFile`/`SnapshotToDir` variants (the async-if form covers both trigger paths), `MetricsHook`/`LoggerHook` (the app logs around the seam itself).

## Ruling

Keep as-is; the seam boundary (server declares, app implements) is the same shape as the CRM/paperless adapters. **Refresh trigger:** a second trigger site (e.g. the SSE hub) would reuse the recorder unchanged; a MetricsHook consumer appears only if /metrics ever needs capture diagnostics.

## How to verify

```console
rg -n "flightrecorder\\." internal/app/flightrecorder.go | wc -l   # expect >= 8 (config + trigger + capture calls)
rg -n "TraceRecorder" internal/server/server.go | head -3          # the seam lives server-side, framework-free
rg -n "go-flightrecorder" go.mod                                   # DIRECT pin v0.2.1
```
