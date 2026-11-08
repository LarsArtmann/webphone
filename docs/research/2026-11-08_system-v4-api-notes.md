# system/v4 adoption spike — API notes + verdicts (A01, A02, A06)

> **Date:** 2026-11-08 · **Plan:** [2026-11-08_04-04_SUPERB-command-event-sourcing.md](../planning/2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 0
> **Verdict:** **GO** on all three spikes — API (A01), sqlite/nix (A02), read-your-writes + replay (A06). The pinned version is `system/v4 v4.17.0` (latest published; local checkout is 50 commits past tags and is NOT the pin).

## A01 — the API surface (verified against the published module, not the local checkout)

All paths relative to `$(go env GOMODCACHE)/github.com/larsartmann/go-cqrs-lite/system/v4@v4.17.0`.

### Construction and lifecycle

- `system.New(ctx, DomainConfig, DeploymentConfig, ...Option) (*System, error)` — `constructor.go:26`. Runs `CheckSafety` first (SCREAM-tier violations block construction, WARN/ADVISORY land on `ScreamReport`). Creates engines in sorted name order (deterministic errors), wires source-of-truth + projection instances, plans projections via `metaengine.Plan`, builds the projection host, then calls the consumer's `domain.Commands/Queries/Timers` closures LAST (`constructor.go:336-347`).
- `System.Start(ctx)` — `system_lifecycle.go:47`. Once-only (`ErrAlreadyStarted`); starts projection host + managed timers.
- `System.Close()` — `lifecycle.go:78`. Idempotent (`stopped` latch); order: managed timers → projection host → engines (topologically ordered by `ShutdownDependencies`) → registered closers; errors joined. **Pinned by the spike: double Close returns nil.**
- `System.GracefulClose(ctx)` — drains registered Drainers under the context, then Close unconditionally. This is the SIGTERM shape webphone's `App.Shutdown` should call.
- `New` error paths tear everything down via `sys.fail(err)` (projection host → engines → closers, errors joined) — no leaked handles on boot failure.

### DomainConfig (consumer side — `config_types.go`)

- `Commands func(*System)` / `Queries func(*System)` — registration closures; called at the END of `New`.
- `Projections []ProjectionDeclaration` — sealed interface; values from `Lookup/QuerySet/Count/RawQuery`. **`system.RawQuery(metaengine.QueryDecl)` is webphone's path: the existing readmodels `metaengine.Query[Q,R]` declarations move here VERBATIM** (spike-proven below).
- `Events []event.Type` — coeffect gate: consuming an undeclared type is a hard error (`ErrDanglingEventSubscription`); declared-but-unconsumed is advisory. Webphone SHOULD declare its full fact-type universe per phase.
- `Schema []schema.EventSchema` — payload upcasting (v-current reads over old stored bytes). Relevant when payload shapes evolve post-migration.
- `Middleware []command.Middleware` — command-level (validation, authz); `UseCommandMiddleware` wires them.
- `ProjectionHostOptions []projectionhost.HostOption` — **the DLQ mounts here** (`projectionhost.WithDeadLetterStore(dlq, retries)`).
- `CheckpointStore` — nil = persisted as a `system_checkpoints` Map collection on the engine (ADR-0142); exactly what we want on sqlite.
- `ShutdownDependencies []ShutdownDependency{Before, After string}` — close-order edges over engine names.

### DeploymentConfig (operator side)

- `Engines map[string]EngineConfig{Driver, DSN, Pragmas, MaxOpenConns, MaterializedViews, Priority}` — driver `"sqlite"` comes from importing `sqliteengine/v4` (its `init()` registers with `metaengine.RegisterDriver`, over modernc `database/sql`). `Pragmas: ["wal"]` matches webphone's SQLite posture. **`MaxOpenConns` matters: sqlite engines default to pool 1; raise it only with evidence.**
- `Instances []InstanceConfig{Role, Engine|Engines, Durability, Publish, Cache}` — `RoleSourceOfTruth` ("source-of-truth") holds events+commands+queries+snapshots+checkpoints; `RoleProjections` plans the projection store over its engine pool.
- Durability tiers: strict (fsync every commit) / normal (WAL, crash-safe not power-safe) / relaxed. Webphone parity: webphone.db rides normal-class durability today; recommend `DurabilityNormal` explicitly at the instance, decide strict at D-ruling time if the journal deserves it.
- No source-of-truth configured ⇒ implicit MEMORY fallback with an ADVISORY diagnostic — a config typo must never reach production silently; the boot check will assert the sqlite instance exists.

### Commands, deciders, dispatch (register.go, system.go)

- `system.RegisterDecider[State](sys, streamType, decider.Decider[State]{Initial, Apply})` — Apply is the pure fold `(state, evt) (state, error)`. Snapshot store auto-wires when the engine implements SnapshotBackend; `WithSnapshotStrategy(...)` enables writes.
- `system.RegisterCommand[Cmd, State](sys, name, func(ctx, Cmd) system.Op[State])` — the handler returns `system.Execute(ctx, streamID, streamType, DecideFunc)`. `DecideFunc` = `(state, currentVersion) ([]event.Event, error)`; events via `event.New(type, streamID, streamType, ver+1, payload)` (CBOR default codec; payload struct → typed bytes).
- Dispatch: `command.New(name, streamID)` builds a `*command.BasicCommand`; `sys.CommandDispatcher().Dispatch(ctx, cmd)` runs middleware → handler → `repo.ExecuteRef` (load → decide → save → publish). Errors from DecideFunc reject the command (nothing saved).
- Reads: `sys.MetaEngine()` exposes the projection store — `metaengine.NewReader[V](store, collection)` is the SAME reader readmodels uses today. `sys.EventStore()`, `sys.Dedup()`, `sys.SnapshotStore()` expose the rest.

### Projection plumbing (the part the plan cared most about)

- ONE adapter named `"projections"` wraps the WHOLE projection store (`projection_wiring.go`); the host registers that single adapter as one worker (`constructor.go:283`). Event type → collection routing happens inside `Store.ApplyEncodedRecord`, which detects CBOR/JSON and decodes into the fold's typed param — **zero consumer decoder code when the journal event type string equals `metaengine.EventTypeName(FactStruct{})`** (bare struct name; `metaengine/fold.go:257`).
- Worker model: one goroutine per projection (here: exactly one), initial journal drain → live subscriber callback, serialized by `handleMu`; `seenIDs` ring dedups the drain/live overlap; checkpoint batching live. Restart = backoff with full jitter; `maxRestarts` exhausted ⇒ `WorkerFailed` + `onFailed` callback (health wiring point).
- Rebuild recipe: host `Stop → System.ResetProjection(ctx, "projections") → host.Start` — "projections" is the one worker name. Reset clears checkpoint AND the adapter's read-model state (`projectionadapter/reset.go`).

### Introspection (health legs)

`HealthCheck(ctx) error`, `HealthCheckDetailed(ctx) []EngineHealth`, `WorkerStatus() []WorkerState` (status/processed/errors/restarts/checkpoint/lastError/lag), `LagPerProjection() map[string]time.Duration`, `ShutdownOrder() []string`, `VerifyProjections(ctx)`, `ScreamReport()`, `ProjectionPlan()`.

## A02 — sqlite driver + nix proof

- `sqliteengine/v4 v4.7.0` registers driver `"sqlite"` via `init()` over `modernc.org/sqlite v1.61.1` — the SAME driver family webphone's `webphone.db`/`readmodels.db` already ride (CGO-free, nix-proven for years in this repo).
- Spike binary built with `CGO_ENABLED=0 go build` inside the devshell: **statically linked ELF, runs, journal+projection+checkpoint all in ONE 164 KiB `system.db`**.
- Module-graph delta vs webphone's current go.mod: **zero new direct modules; the only new module anywhere in the sum graph is `github.com/knadh/koanf/parsers/yaml`** (system's config loader). The A10 vendorHash re-pin is therefore a version-bump re-pin, not a new-vendor re-pin.
- Residual nix risk (bounded, lands at A10): the FOD rebuild under `nix build .#webphone.goModules --rebuild` is the authoritative nix proof; devshell-static-ELF plus a two-module delta makes a red there very unlikely.

## A06 — read-your-writes + replay (the R2/R12 kill shots)

Spike: `/tmp/hello-system` (dial-window-shaped domain: DialAttempted fact → dial_windows collection fold — mirrors `internal/readmodels/dialwindows.go` verbatim). Measurements on this host:

| Measurement                                             | Result            |
| ------------------------------------------------------- | ----------------- |
| Command dispatch (decider + journal append + publish)   | 0.9–1.1 ms        |
| Projection visibility AFTER dispatch returned           | 1.1–2.0 ms        |
| Replay rebuild (wipe projection → host replay)          | 1.6 ms            |
| Close idempotency                                       | nil, nil          |

Findings:

1. **Read-your-writes is NOT synchronous** — dispatch returns before the fold lands (v1's retracted claim stays retracted). Visibility is milliseconds: the gochannel bus + worker goroutine apply the fold ~1-2 ms after the append.
2. **The honest D7 mechanism: bounded catch-up await at the seams that must see their own write.** Commands whose ANSWER is the decider's verdict (dial-gate: allowed/denied + stamp) answer from the decision itself — zero added latency. Seams that re-read state for the response (messaging send → SSE fragment) await projection visibility of the command's own event IDs, bounded (250 ms budget, ~2 ms typical) before emitting the `ChangeFunc` nudge. Fallback on budget expiry: emit anyway (staleness, never data loss) + WARN — same availability posture as today's `idemGate` fail-open.
3. **Replay is proven**: Stop → Reset("projections") → Start rebuilt the view from the journal alone in 1.6 ms. This is also the importer primitive (A08: synthesize events, replay, diff) and the corruption-recovery primitive (backup restore of system.db replays for free).
4. Projection lag is observable (`LagPerProjection`, `WorkerStatus`) — the `/healthz` leg can gate on it honestly.

## Resolved plan unknowns (§12 OPEN list)

| Open item                                              | Resolution                                                                     |
| ------------------------------------------------------ | ------------------------------------------------------------------------------ |
| Exact DomainConfig field surface on the pinned version | Mapped above (v4.17.0)                                                         |
| cqrs-lint finding inventory                            | Still open — A04 worktree run owns it                                          |
| systemscenario import path                             | `github.com/larsartmann/go-cqrs-lite/systemscenario/v4` (rides the same graph) |
| projectionhost apply model for sqlite single-instance  | Async goroutine worker, ~1-2 ms visibility (A06 table)                         |

## Reproduction

```console
# spike sources: /tmp/hello-system/{go.mod,main.go} (dial-window-shaped, plan-mirroring)
nix develop -c sh -c 'cd /tmp/hello-system && go mod tidy && CGO_ENABLED=0 go build . && ./hello-system /tmp/hello-system/system.db fresh && ./hello-system /tmp/hello-system/system.db reopen'
```
