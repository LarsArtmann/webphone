# T23 — `appkit/flightrecorderhealth` demand-gate evaluation

**Date:** 2026-11-15 · **Ruling: PARK** — candidate AFFIRMED, surface verified end-to-end, wiring memo'd; adoption stays behind the named demand (TODO_LIST row) · One home: docs/review/libs/go-appkit.md § module table · framework layer NO-GO stands (AGENTS.md appkit bullet; asks #59/#60/#61 filed).

## Surface verification (flightrecorderhealth@v0.1.7, module cache 2026-11-15)

- `NewTrigger(rec *fr.Recorder, ...TriggerOption)` implements go-health's `HealthRecorder` by RUNNING the batch itself (`injector.HealthCheckWithContext`), wraps it in `fr.TriggerContext{Kind:"health.check", Type:"batch", Err: firstError}`, cooldown-gates under a mutex, then `rec.SnapshotIfAsync(context.WithoutCancel(ctx), ...)` — non-blocking, race-safe, returns the batch unchanged (health-data pass-through).
- Options: `WithTriggerFunc` (default `fr.OnError`), `WithCooldown` (0 default; 30–60s recommended for dir sinks like webphone's), `WithTriggerLogger`, `WithServiceName`. Nil rec = pure pass-through (safe with the seam off).
- Constructor is standalone — it only READS the injector for the batch; `Register`/`NewCheckable` (the dashboard row for the recorder's own state) is a separate opt-in. No conflict with the injector doctrine.
- Module go.mod requires go-health v0.5.1 (older than webphone's v0.10.0 — MVS upgrades; `WithHealthRecorder` exists in the v0.10.0 pin at options.go:127), go-flightrecorder v0.2.1 (exact match), go-error-family v0.11.0, samber/do v2.

## Deps + go-health semantics (verified)

- webphone go.mod: go-flightrecorder v0.2.1 ✓, go-health v0.10.0 ✓ — both hard deps pinned.
- go-health v0.10.0 honors the recorder on the INJECTOR path only (`health.New` → `resolveHealthCheck`); function-path constructors DROP it (recorderDropped). Webphone uses the injector path — honored.
- The recorder sees EVERY batch, including the background refresh loop's ticks.

## Does webphone have this glue hand-rolled today? NO.

- Verified: zero occurrences of `flightrecorderhealth`/`WithHealthRecorder`/`frhealth` in internal/ and go.mod. The traceRecorder (internal/app/flightrecorder.go) captures ONLY via HTTP requests (TraceMiddleware → Observe, >2s or 5xx) and the manual `GET /internal/flightrecorder` endpoint (internal/server/server.go:879). Health-batch→snapshot glue does not exist — adoption is NEW capability, not dedup. Different bar: demand-gated.

## The demand, sharpened (vs the 2026-11-12 row wording)

- The always-on wedge detector exists only when the dashboard is ON: `probeRefreshIf` returns the 1s `probeRefresh` loop only under `cfg.Dashboard.Enable` (app.go:566), and the dashboard is default-OFF behind the fence decision. Dashboard-off = live mode (interval 0): the lane fires only when something actually polls /healthz — and nothing does routinely (`Type=simple`; health_crm_test.go: "Nothing routes traffic on /healthz").
- Delta over the request arm is real but conditional: dashboard-on deployments get wedge snapshots with zero HTTP traffic; every deployment gets a cooled snapshot whenever anyone curls /healthz during an incident.
- One row correction: the module's cooldown hardens the NEW health lane only — it does not retroactively cool webphone's un-cooled 5xx request arm.

## Wiring prototype (memo-level; package-internal access is fine)

1. app.go: build `trace := traceRecorderFor(cfg.DataDir)` BEFORE the probe (~8-line order flip), then, when `trace != nil`, add to `health.New(...)`: `health.WithHealthRecorder(frhealth.NewTrigger(trace.rec, frhealth.WithTriggerFunc(trace.trigger), frhealth.WithCooldown(30*time.Second), frhealth.WithTriggerLogger(log), frhealth.WithServiceName("webphone")))` — `trace.trigger` already composes the disk guard, so the health lane inherits the 1 GiB floor + skip counting for free.
2. Optional: `frhealth.Register(injector, trace.rec, "flight-recorder")` for the dashboard row; skip while the seam is off so /healthz stays byte-identical.
3. Cost when off: one import + the nil-recorder pass-through; zero behavior change.

## Why PARK, not ADOPT

New capability ≠ dedup, so the bar is demonstrated demand: a wedge incident the request-trigger misses, or a flight-recorder train to ride. Neither has arrived. The candidate is verified real and wireable in ~6 lines the day demand lands.

**Reopen trigger (unchanged in substance):** the first wedge incident the request-trigger misses (esp. dashboard-on, or any /healthz poll during an incident), or the next go-flightrecorder minor touching trigger semantics.

Evidence: ~/go/pkg/mod/github.com/larsartmann/go-appkit/flightrecorderhealth@v0.1.7/{adapter.go,doc.go,go.mod}; go-health@v0.10.0/{probe.go,options.go,doc.go}; webphone internal/app/{app.go,flightrecorder.go,dashboard.go}, internal/server/server.go, go.mod.
