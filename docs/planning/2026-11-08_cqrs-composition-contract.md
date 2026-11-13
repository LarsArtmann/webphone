# system-in-do composition contract (A03)

> **Date:** 2026-11-08 · **Plan:** [SUPERB command & event sourcing](2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 0 · **Status:** contract as-built for Phase 1 (A11–A14 implement exactly this)

## 1. The one rule

`system.New` does NOT replace `internal/app`. It runs INSIDE the samber/do container as ONE named eager service (`"cqrs"`), exactly like `"sqlite"`, `"blob-dir"`, and `"metaengine"` today. The container keeps owning object lifetime; `System` keeps owning its internal engines. Two worlds, one composition root, zero lifecycle fights.

## 2. Registration (internal/app/cqrs.go, Phase 1)

```go
// CqrsService adapts *system.System to the container's lifecycle contract.
type CqrsService struct {
    Sys *system.System
    // started latches App.Start's single sys.Start call (System.Start is
    // once-only; a second App.Start must not red-tip it).
}

// Provider: do.ProvideNamed(injector, "cqrs", ...) → NewCqrs(cfg, log)
```

- `NewCqrs` builds `system.DeploymentConfig`:
  - ONE engine: `"db" = {Driver: "sqlite", DSN: filepath.Join(cfg.DataDir, "system.db"), Pragmas: ["wal"]}`.
  - TWO instances on that engine: `RoleSourceOfTruth` + `RoleProjections` — journal, projections, dedup, checkpoints, snapshots, DLQ all in the ONE file (the §2 terminal topology; spike-proven).
  - `DurabilityNormal` explicit on the source-of-truth instance.
- `DomainConfig` in Phase 1 is the SKELETON: no commands, no projections — the projection host is therefore absent and `Start` is inert. The hollow-root WARN is EXPECTED and LOGGED (`log.Warn("cqrs hollow root", "reason", "phase 1 plumbing — first aggregate lands in phase 2")`). **Forbidden after Phase 2's first aggregate registers** (risk R5's gate: an arch or unit test asserts `DomainConfig.Commands != nil` once dial windows land).
- Eager instantiation: `MustInvokeNamed[*CqrsService](injector, "cqrs")` inside `New` (fail-fast miswires, same doctrine as the critical pair).
- Lifecycle conformance asserted adapter-side in app.go's guard block:
  `var _ do.ShutdownerWithError = (*CqrsService)(nil)` and `var _ do.HealthcheckerWithContext = (*CqrsService)(nil)`.

## 3. Boot order (unchanged sequence, one insertion)

```
app.New:    container registers everything; MustInvokeNamed("cqrs") constructs
            system.db (file + engines exist; NO workers run yet)
cmd/webphone: signal ctx → app.Start
app.Start:  dashboard → Probe.Start → resolve sqlite/blob/metaengine
            → **cqrs.Sys.Start(ctx)** (starts projection workers + timers)
            → retention.Start → dedup sweeper → chat loops
cmd/webphone: httpServer.Start
```

`sys.Start` runs BEFORE the HTTP listener accepts traffic, so the projection workers' initial journal drain never races a request that expects a caught-up view. The chat-loop and retention seams keep their existing order (they do not touch the journal until their phases).

## 4. Shutdown order (the constraint the plan called out)

Current chain (cmd/webphone main): `Hubs.Shutdown()` pre-drain → `httpServer.Shutdown` (30 s) → `application.Shutdown()` (Probe.Shutdown → `injector.Shutdown()` cascade → Trace.Close last).

Insertion: the cascade calls `CqrsService.Shutdown()`, which runs `Sys.GracefulClose(ctx)` with a 10 s bound (drains any registered Drainers, stops projection workers first, then engines — System.Close internal order), joining errors. Because `App.Shutdown` only runs AFTER the HTTP drain completed, no dispatch can be in flight; because `system.db` is its own file, there is NO ordering constraint against `"sqlite"`/`"metaengine"` closing in the same cascade — the cascade order is irrelevant by construction. `Trace.Close` stays last (unchanged).

```
signal → Hubs pre-drain (SSE released instantly)
       → HTTP drain (waits in-flight requests ⇒ waits in-flight dispatches)
       → Probe.Shutdown
       → injector.Shutdown cascade:
             cqrs.GracefulClose(10s)   (workers → engines → closers)
             sqlite / metaengine / …   (own files, order-independent)
       → Trace.Close
```

Rollback: Phase 1 is one revert; `system.db` left behind by a reverted build is inert (an unknown file the next boot ignores).

## 5. Deps threading (designed now, wired per phase)

- `server.Deps` grows `CQRS *CqrsRuntime` (nil = seam off — test compositions keep working, the nil-fallback doctrine). `CqrsRuntime` exposes exactly what each phase consumes: `Dispatch(ctx, cmd) error`, `MetaEngine() *metaengine.Store` (projection reads), `Dedup()`, `HealthCheck/WorkerStatus/LagPerProjection` (health legs). Phase 1 ships the field nil + the wiring-parity test unchanged; Phase 2 wires it.
- The projection reader for migrated collections is `metaengine.NewReader[V](sys.MetaEngine(), collection)` — the SAME call readmodels uses today, so `readmodels.go`'s TypedReader plumbing ports as-is.
- The Notifier/SSE path is untouched: command handlers keep emitting `domain.ChangeFunc` post-append (post bounded catch-up await per the A06 verdict).

## 6. Health wiring

- Probe leg name: `"cqrs"`, OPTIONAL NON-CRITICAL while the seam carries only advisory features (dial gate fail-open posture), flipping to critical at the owner's call once core threads ride the journal (Phase 5 gate).
- `HealthCheck(ctx)`: `Sys.HealthCheck(ctx)` + worker triage: any `WorkerStatus()` in `WorkerFailed`/draining-too-long or `LagPerProjection` beyond a bound (start: 30 s) ⇒ the leg fails naming the worker — the same shape cqrs-htmx's projections leg uses on `/healthz`.
- DLQ: `projectionhost.WithDeadLetterStore` in `DomainConfig.ProjectionHostOptions` (A12); the DLQ's own readability surface is A12's work, not the probe's.

## 7. Invariants this contract preserves

1. `internal/app` stays the ONLY injector home; services stay framework-free (cqrs.go is an adapter, like PasskeyRuntime).
2. `MustInvoke*` composition-root-only (arch test already enforces).
3. Engine DDL never meets the hand ladder: `system.db` is engine-owned; `webphone.db`/`readmodels.db` untouched by system (N5/N6 two-file spirit).
4. `App.Start`/`Shutdown` idempotence semantics unchanged; `CqrsService` adds no second shutdown path (cascade owns it).
5. Phase-1 behavior parity: zero routes change, zero bytes change; the only observable deltas are the new `system.db` file, one log line, and (option-gated) the health leg.
