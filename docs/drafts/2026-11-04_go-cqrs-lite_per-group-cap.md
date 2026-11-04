## Problem

`metaengine.EnforceCap` trims one capped collection to its newest N rows — the keep-newest-N primitive from #72, shipped and working. But N is per **collection**. Consumers whose retention rule is per **key group** — per-owner quotas, per-tenant windows — still hand-roll check-then-expunge loops, which is exactly the duplication #72 set out to kill.

## Evidence

webphone (the consuming deployment) bounds two read models per OWNER, not per collection:

- `call_transcripts`: 5000 rows per owner
- `dial_windows`: 500 rows per owner

A collection-global `EnforceCap` cannot express either without letting one owner's flood evict another owner's rows, so both ride hand-written converging trim loops (`internal/readmodels/transcripts.go`, `dialwindows.go` — scan cap+1, expunge the tail, repeat).

Source at the pinned version (`metaengine/v4@v4.20.0`, `cap.go:134`): `EnforceCap(ctx, queryName)` keeps `capRows` newest rows of the query's whole result set, ordered by `capOrderBy`. The key-extractor machinery (`buildKeyExtractor`, `cap.go:126`) addresses rows for deletion — it does not group. `capScanPage` takes no query args, so a parameterized owner-scoped query cannot be enforced per owner either.

## Proposed change

A per-group cap variant. The shape I'd reach for: `EnforceCapGrouped(ctx, queryName, groupCol string, perGroup int)` — keep the newest N rows **per distinct value of a group column** (the same `capOrderBy` picks the survivors inside each group). Declaration could also live on `WithCap` (`WithCapPerGroup(groupCol, rows)`), matching how the rest of the cap config rides the query.

## Scope

Single-writer SQLite consumers with multi-tenant read models are the beneficiaries; single-tenant users are unaffected and need not touch it. Not needed for replay (the existing replay-then-EnforceCap rule carries over unchanged).
💘 Generated with Crush

Assisted-by: Crush:glm-5.3-flash
