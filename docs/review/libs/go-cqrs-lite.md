# go-cqrs-lite — library utilization audit

**Verdict:** SUPERB (2026-11-08/09, A01–A47) made webphone the family's flagship consumer: EVERY domain aggregate is command/event-sourced through the `system` composition root into one journal (`system.db`), projections rebuild the read views, snapshots keep the hot streams fast, and metaengine remains the read-model/projection substrate plus the catalog lanes. **Adoption 96/100 (flagship)** — 13.5 of 14 applicable rows.
**Pin:** system/v4 v4.18.0, event/v4 v4.15.2, command/v4 v4.14.1, decider/v4 v4.8.0, projectionhost/v4 v4.7.1, snapshot/v4 v4.7.0, metaengine/v4 v4.23.0, metaengine/sqliteengine/v4 v4.8.0, catalog/v4 v4.14.0, record/v4 v4.6.2, id/v4 v4.7.5 (11 direct; dispatcher/dedup/commandlifecycle/kv/listing/metadata/projectionadapter/claiming ride indirect) · **Currency:** every module at its latest except catalog/v4 — v4.15.0 (2026-11-11, re-verified against remote tags) is a directory restructure (`catalog/` → `catalog/v4/`, same module path) + the hub's `@eventcatalog/core` npm pin ^5.1.0→^5.9.1 + json/v2 in examples; no Go API delta webphone consumes, bump rides the next go-cqrs-lite train · **Status:** direct dep · **Audited:** 2026-11-09 (SUPERB closeout re-score, B159; supersedes the 2026-11-04 83/100 snapshot below) · **Currency re-verified:** 2026-11-11.

## Errata — EnforceCap NO-GO (2026-11-04), SUPERSEDED 2026-11-06

The 2026-11-04 ruling (EnforceCap is collection-GLOBAL; webphone's caps are per-OWNER) was correct and produced the upstream ask [go-cqrs-lite#114](https://github.com/LarsArtmann/go-cqrs-lite/issues/114); the library then shipped **`WithCapPerGroup`** (ADR-0164, metaengine v4.21.0) and the supersession LANDED 2026-11-06: the transcripts per-OWNER bound and the dial windows/overrides caps are now DECLARATIVE (`WithCapPerGroup("owner", …)` at readmodels/transcripts.go:94, readmodels/dialwindows.go:78, cqrs/transcripts.go:298) and those hand loops are GONE. The per-CALL transcript floor (1500, nested owner+call addressing) stays a loop — a cap addresses ONE group; that nested-addressing gap is the row's only remaining reopen trigger (TODO_LIST reopen registry).

## Status

|         |                                                                                                                                                                                                                                                                                                                                                                           |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Modules | system/v4, event/v4, command/v4, decider/v4, projectionhost/v4, snapshot/v4, metaengine/v4, sqliteengine/v4, catalog/v4, record/v4, id/v4 (direct)                                                                                                                                                                                                                        |
| Usage   | internal/cqrs/* (25 source files — threads/faxes/dials/enroll/contacts/snippets/transcripts/notes/reviews/agent-audit/push-devices aggregates + projections + retention), internal/retention (JournalSweepers over `TrimStream`), internal/app/cqrs.go (composition), internal/readmodels (engine store as the frozen import source), cmd/webphone + server catalog lanes |

## What the library offers (the consumed modules)

- **system/v4**: `system.New(ctx, Domain(decls…), deployment)` composition root — journal store, bus, projection hosting, snapshot wiring, health/status; `RegisterDecider[S]`; `sys.Dedup()` (durable CAS window); `sys.TrimStream` (retention: journal bytes leave the store); `system.Domain(decls...)` composes aggregate declarations.
- **event/command/decider/v4**: the machine-table core — `decider.Decider[S]{Initial, Decide, Apply}` decides legality from state; commands carry an `ID()` METHOD; webphone's legality tables (`deliveryMoves`/`faxMoves`/`AppendToThread`) are consumed verbatim at the decision.
- **projectionhost/v4**: declarative projection registration (`ProjectionHostOptions` + `FilterOnField`/`SortOnField` declarations); readiness surface feeds the passkey-gated `/healthz` "projections" leg.
- **snapshot/v4**: `EveryNEvents(200)` + `NewReadPressure(50, WithInnerStrategy(every))` — the resolved D8 strategy (grind 3.3 ms → 0.84 ms, measured in the plan errata).
- **metaengine/v4**: `Query[Q,R]` declarations + folds/Remove, `TypedReader` ScanOptions, capabilities DedupStore + AggregateReader, **`WithCapPerGroup`** (per-group declarative caps, the #114 product), `WithLimit` (default scan ceiling 100 — load-bearing).
- **sqliteengine/v4**: `NewFromDSN` (WAL pragmas); planned tables `meta_planned_<collection>`; implements both capabilities (asserted at boot).
- **catalog/v4 + record/v4 + id/v4**: the OpenAPI/AsyncAPI/EventCatalog export lanes; the Record envelope + StreamRef (retention refs); branded ids.

## Gap analysis

| #   | Capability area                                                    | Status                | Evidence / reason                                                                                                                                                  |
| --- | ------------------------------------------------------------------ | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | Engine Open + capability assertions                                | 🟢 Fully leveraged    | readmodels.go; fail-fast boot doctrine                                                                                                                             |
| 2   | `Plan` + Query declarations (11 collections)                       | 🟢 Fully leveraged    | readmodels.go:87 + collection files                                                                                                                                |
| 3   | TypedReader Filter/Sort/Limit reads                                | 🟢 Fully leveraged    | `WithLimit` is load-bearing (AGENTS.md); the 100-row ceiling lesson is battery-pinned (205-row test)                                                               |
| 4   | OnRecord folds + Remove                                            | 🟢 Fully leveraged    | ~20 fold declarations; `FaxesProjection` shared by two engine homes                                                                                                |
| 5   | DedupStore adaptation                                              | 🟢 Fully leveraged    | app/dedup.go (hooks/calls windows) + cqrs/faxesstore.go `sys.Dedup()` (durable fax replay guard)                                                                   |
| 6   | AggregateReader counts                                             | 🟢 Fully leveraged    | contacts count etc.                                                                                                                                                |
| 7   | catalog builder + 3 exporters                                      | 🟢 Fully leveraged    | apidocs.go, wirecontracts.go, cmd/webphone/catalog.go — drift-free since 2026-10-24                                                                                |
| 8   | **`WithCapPerGroup` (native grouped caps)**                        | 🟢 Adopted 2026-11-06 | transcripts per-owner + dial windows/overrides caps are declarative (3 sites); the per-CALL floor stays a hand loop (nested addressing — the named reopen trigger) |
| 9   | Advanced ScanOptions (WithCursor keyset, WithRange/WithIn/WithOr)  | 🟡 Partial            | newest-window scans bounded by WithLimit; no keyset pagination — fine at 5000/owner caps, interesting at ten-fold growth                                           |
| 10  | system composition root (`system.New`, Domain, lifecycle)          | 🟢 Fully leveraged    | app/cqrs.go:228; the ONE journal world (system.db); `ExtensionHubs.Shutdown`-style pre-drain doctrine mirrored by the container shutdown                           |
| 11  | Decider machine tables (event/command/decider)                     | 🟢 Fully leveraged    | 11+ `system.RegisterDecider` aggregates (enroll/callnotes/callreviews/transcripts/snippets/pushdevices/…); legality single-home doctrine (AGENTS.md)               |
| 12  | projectionhost declarative registration + readiness                | 🟢 Fully leveraged    | ProjectionHostOptions declarations (threadsregister, faxesstore); `/healthz` projections leg                                                                       |
| 13  | Snapshot strategy (EveryNEvents + ReadPressure)                    | 🟢 Adopted (D8)       | threadsregister.go: `snapshotEveryNEvents=200`, `ReadPressure(50)` — measured grind 3.3→0.84 ms                                                                    |
| 14  | Journal retention (`TrimStream`)                                   | 🟢 Adopted            | retention.JournalSweepers {Threads, Faxes, Transcripts}: tombstone → await own write → TrimStream (bytes GONE; GDPR arm)                                           |
| n/a | record envelope metadata (Actor/Cause/Stamp)                       | ⚪ N/A                | folds read the typed domain event; webphone's events carry their own facts (actor/causation live in the agent-audit DOMAIN facts, not the envelope)                |
| n/a | Multi-engine ops (Replan/AddEngine/Backfill/materialized/temporal) | ⚪ N/A                | single-engine-per-DB posture; no hot tier                                                                                                                          |
| n/a | GroupedAggregateReader                                             | ⚪ N/A                | counts only                                                                                                                                                        |

**Score:** rows 1-8 + 10-14 = 1.0 (12.0) + row 9 = 0.5 → **13.5 / 14 ≈ 96/100, flagship band.**

## Ruling

The doctrine INVERTED with SUPERB: every new aggregate grows on the system composition root (`cqrs` package + `system.Domain` decl + decider + projections + retention leg), never as a new SQL table — the storage ladder is FROZEN (docs/planning/2026-11-09_storage-ladder-retirement.md; sessions stay SQL per D2). The engine-backed readmodels store stays as the frozen import source/rollback lane until the operator-gated legacy DROP. Reject-caps stay app-side. The per-CALL nested cap is the library's one open addressing gap (reopen trigger).

**Refresh trigger:** the per-CALL nested-addressing cap ask (upstream); any owner crossing ~10k rows per collection (keyset pagination); a second engine (failover/backfill).

## How to verify

All commands run from the **webphone checkout root**.

```console
rg -n "system.New\(" internal/app/cqrs.go
rg -n "RegisterDecider" internal/cqrs/enroll.go
rg -n "WithCapPerGroup" internal/readmodels/transcripts.go internal/readmodels/dialwindows.go internal/cqrs/transcripts.go
rg -n "EveryNEvents|ReadPressure" internal/cqrs/threadsregister.go
rg -n "TrimStream" internal/cqrs/threadssweep.go internal/cqrs/faxesstore.go
rg -n "EnforceCap" /home/lars/projects/go-cqrs-lite/metaengine/cap.go
rg -n "metaengine.Plan\(" internal/readmodels/readmodels.go
rg -n "catalogopenapi.NewExporter|asyncapi.NewExporter|eventcatalog.NewExporter" internal/server/apidocs.go internal/server/wirecontracts.go cmd/webphone/catalog.go
```
