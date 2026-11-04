# go-cqrs-lite — library utilization audit

**Verdict:** the 2026-10-15 readmodels train turned this from "0% direct" into the deepest dogfooding in the repo: metaengine IS the read-model substrate, and the catalog lanes (OpenAPI/AsyncAPI/EventCatalog) ride its exporters. **Adoption 83/100 (high)** — 7.5 of 9 applicable rows.
**Pin:** catalog/v4 v4.12.0, metaengine/v4 v4.20.0, metaengine/sqliteengine/v4 v4.6.0, record/v4 v4.6.2 (all direct; ~21 more /v4 modules indirect via cqrs-htmx) · **Latest (family tag):** cmd/cqrs-lint/v4.20.0 · **Currency:** current at the four required modules · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** [2026-10-24 catalog deep-dive](../../research/2026-10-24_go-cqrs-lite-catalog-deep-dive.html) (0/100 snapshot, gap closed same day — read its Addenda first); [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (n/a at the time: zero direct imports — fully superseded).

## Errata — EnforceCap ruled NO-GO (T07 verification, 2026-11-04)

The plan's step-2 verification ("confirm newest-rows semantics") FAILED, and the row-8 finding was **wrong on semantics**. `WithCap(maxRows, orderBy)` + `Store.EnforceCap` trim a whole **collection** to maxRows — there is no per-key grouping option (verified against cap.go + query_config.go: the keep page is one descending prefix; capKeyExtractor only maps victims back to keys). Webphone's trims are **per-owner** (`trimOwner`: owner-filtered 5000) and **per-call** (`trimCall`: owner+call_id filtered) and **per-owner** again (`trimWindows`: 500/owner). Swapping would change the guarantee from "each owner keeps 5000" to "all owners share N total" — on a multi-extension deployment one owner's flood would silently evict ANOTHER owner's transcripts/dial windows, which is guard §5.1's exact wrong (reject-caps protect owner data; per-owner caps do too). The hand loops are the correct one-home for the grouped semantics: bounded cap+1 scans, converging passes, owner-scoped. **Refresh trigger for a re-look: metaengine growing a per-group cap (`WithCapPerGroup(groupCol, maxRows, orderBy)` or equivalent) — then the three trims migrate in one train.**

## Status

| | |
|---|---|
| Modules | catalog/v4, metaengine/v4, metaengine/sqliteengine/v4, record/v4 (direct) |
| Usage | 19 files: internal/readmodels/* (12), internal/app/dedup.go, internal/domain/catalog.go, cmd/webphone/catalog.go, internal/server/{apidocs,wirecontracts}.go |

## What the library offers (the four required modules)

- **metaengine/v4**: cost-based planner over engines; `Query[Q,R]` declarations + `OnRecord` folds/`Remove`, `FilterOnField`/`SortOnField`/`Volume`; `Plan(engines)` → `Store`; `TypedReader` (`NewReader[R]`) with `ScanOptions` (WithFilter/WithRange/WithIn/WithOr/WithSort/WithLimit/WithCursor…); capabilities `DedupStore` (`DedupCheckAndRecord/DedupSeen/DedupSweep`) + `AggregateReader`; **`EnforceCap(name)`/`EnforceCaps(ctx)`** (cap.go:144,228 — newest-rows trim); failover ops (`AddEngine`/`Backfill`/`Replan`), materialized views, temporal `ExecuteAsOf`, Export/Import, HealthSnapshot.
- **sqliteengine/v4**: `NewFromDSN(dsn)` (WAL/busy_timeout pragmas), planned variants, `WithMaterializedViews`/`WithCellVersioning`; implements DedupStore + AggregateReader + reset.
- **record/v4**: `Record` envelope (ID/Type/Payload/Encoding/StreamRef/MetaData: actor, cause, stamps, schema version) + `StreamRef`/`Actor`/`Cause`/`Stamp` primitives.
- **catalog/v4**: `NewBuilder` → AddService/AddDomain/Registry().AddSecurityScheme → `Event[T]`/`Command[T]`/`Query[T]` with WithName/WithSummary/MsgOperation/MsgSecurity/Response[T]/WithResponse; exporters: `openapi`, `asyncapi` (`WithServer(name,"","")`), `eventcatalog` (`WithPlainRefIDs`), plus d2/architecture/docserver/huma/compat.

## How webphone uses it (internal/readmodels/readmodels.go pattern)

1. `sqliteengine.NewFromDSN(dataDir/"readmodels.db")` — dedicated engine-owned DB (N5/N6: engine DDL never meets the hand ladder), separate from hand-migrated webphone.db.
2. **Boot-time capability assertions**: `eng.(metaengine.DedupStore)`, `eng.(metaengine.AggregateReader)` — an engine swap that drops one fails boot, not first use.
3. `metaengine.Plan([]Engine{eng}, <11 QueryDecl vars>)` → `*metaengine.Store`; one `Query` declaration per collection (contacts, snippets, faxes, transcripts, call notes, call reviews, dial windows, enroll tokens, push devices, agent audit).
4. Views are plain structs with json tags; folds are `OnRecord(Saved→upsert)` / `OnRecord(Deleted→Remove[View]())`; owner scoping via `FilterOnField("owner", FilterEq)`.
5. Reads: `reader.Get/Scan(... WithFilter(owner), WithSort(col, true), WithLimit(cap+1))`; counts via AggregateReader; **reject-caps enforced app-side** (ErrListFull, internal/readmodels/contacts.go:25,146); **trim-caps hand-rolled** (transcripts trimOwner/trimCall, transcripts.go:106-130; dial_windows trimWindows, dialwindows.go:229).
6. Dedup: internal/app/dedup.go adapts `metaengine.DedupStore` into the server-facing `hooks.DedupWindow` port + 15-min sweep loop (hooksIdem/callsIdem ride it).
7. Catalog lanes: cmd/webphone/catalog.go (EventCatalog export), internal/server/apidocs.go:391 (OpenAPI), internal/server/wirecontracts.go:129 (AsyncAPI `WithServer("webphone","","")`) — the 10-24 audit's gaps, closed same-day.

## Gap analysis

| # | Capability area | Status | Evidence / reason |
|---|---|---|---|
| 1 | Engine Open + capability assertions | 🟢 Fully leveraged | readmodels.go; fail-fast boot doctrine |
| 2 | `Plan` + Query declarations (11 collections) | 🟢 Fully leveraged | readmodels.go + 10 collection files |
| 3 | TypedReader Filter/Sort/Limit reads | 🟢 Fully leveraged | `WithLimit` is load-bearing (AGENTS.md) |
| 4 | OnRecord folds + Remove | 🟢 Fully leveraged | ~20 fold declarations |
| 5 | DedupStore adaptation | 🟢 Fully leveraged | internal/app/dedup.go; idem windows are DURABLE since the readmodels train |
| 6 | AggregateReader counts | 🟢 Fully leveraged | contacts count etc. |
| 7 | catalog builder + 3 exporters | 🟢 Fully leveraged | apidocs.go, wirecontracts.go, cmd/webphone/catalog.go |
| 8 | **`EnforceCap`/`EnforceCaps` (native trim)** | 🔵 NO-GO (ruled 2026-11-04) | The audit's "exactly the shape" claim was wrong: EnforceCap trims a whole COLLECTION; webphone's trims are per-OWNER and per-CALL (owner+call_id filters). A global cap would let one owner's flood evict another owner's rows — the guard-§5.1 wrong. Hand loops stay as the one-home for grouped semantics; reopen on a per-group cap API upstream — FILED 2026-11-04 as [go-cqrs-lite#114](https://github.com/LarsArtmann/go-cqrs-lite/issues/114); ruling re-verified HOLDING at the pin (metaengine v4.20.0 cap.go:134 trims the whole collection; the key-extractor machinery addresses rows for deletion, it does not group; capScanPage takes no query args so a parameterized owner-scoped query cannot be enforced per owner either)
| 9 | Advanced ScanOptions (WithCursor keyset, WithRange/WithIn/WithOr) | 🟡 Partial | History/transcript paging scans the newest window bounded by WithLimit; no keyset pagination, no engine-side compound filters — fine at 5000-row caps, becomes interesting at ten-fold growth |
| n/a | record envelope metadata (Actor/Cause/Stamp) | ⚪ N/A | every fold discards the envelope (`func(_ record.Record, e domain.E)` — deliberate: webphone's domain events carry their own facts) |
| n/a | Multi-engine ops (Replan/AddEngine/Backfill/Export/materialized views/temporal) | ⚪ N/A | single-engine-per-DB is the documented posture; no hot tier |
| n/a | GroupedAggregateReader | ⚪ N/A | counts only |
| n/a | ES modules (event/command/decider/system/projectionhost/snapshot/scheduling/watermill/kv/stack/otel) | ⚪ N/A | no event sourcing in this product; readmodels is a record store — the ratified single-go.mod + no-ES doctrine (AGENTS.md); `stack` is deprecated upstream |

**Score:** rows 1-7 = 1.0 (7.0) + row 9 = 0.5 + row 8 = 0 → **7.5 / 9 ≈ 83/100, high band.**

## Ruling

Deep adoption is correct and should stay the pattern for every new read model ("New read models grow as metaengine.Query declarations"). ~~One real opportunity: declare the two trim caps with metaengine caps and delete the trims~~ **RULED NO-GO 2026-11-04** (see the errata: EnforceCap is collection-global; webphone's caps are per-owner/per-call — the hand loops ARE the correct semantics). Reject-caps stay app-side as before. The remaining gap between the engine and webphone is the missing per-group cap — an upstream ask candidate (T20).

**Refresh trigger:** any read model crossing ~10k rows per owner (keyset pagination via WithCursor); a second engine (failover/backfill); an upstream bounded-AuditLog projection.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "metaengine.Plan\(" internal/readmodels/readmodels.go
rg -n "func .*trim" internal/readmodels/*.go | head -5
rg -n "EnforceCap" /home/lars/projects/go-cqrs-lite/metaengine/cap.go | head -2
rg -n "catalogopenapi.NewExporter|asyncapi.NewExporter|eventcatalog.NewExporter" internal/server/apidocs.go internal/server/wirecontracts.go cmd/webphone/catalog.go
```
