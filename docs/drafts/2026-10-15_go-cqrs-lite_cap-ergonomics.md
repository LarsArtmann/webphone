> [!NOTE]
> This filing was drafted by GLM-5.3-flash via Crush from an AI-run investigation, not at my request. The evidence comes from migrating a production consumer (webphone) onto metaengine v4.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

**TL;DR:** metaengine has no primitive for "keep the newest N rows" — every consumer hand-rolls a check-then-expunge converging loop (bounded scan, tail scan, N remove-applies, repeat), and the check and the apply can't be atomic. Ask: a cap/trim primitive at the fold or store level, or a documented recipe.

## Problem

`TypedReader` exposes `Get`, `Exists`, `GetBatch`, `Count`, `Sum`/`Min`/`Max`/`Avg`, `Scan`, `ScanPage` (`typed_reader.go`, `typed_reader_scan.go`, `typed_reader_aggregates.go` in `metaengine/v4@v4.19.0`). Writes are event-shaped: `Store.Apply` with a fold-declared `OnRecord` — including `Remove[V]()` for keyed removal. What the API can't express is the compound rule "after this apply, at most N rows survive in this collection/group, oldest expunged first".

So every bounded read model re-implements the same machinery:

1. Detect excess: `Count` (SQL pushdown, `typed_reader_aggregates.go:8`) or a bounded scan at `cap+1`.
2. Identify the victims: a second, sorted scan of the tail.
3. Expunge: one `Apply` per victim with a `Remove`-shaped event.
4. Repeat from 1 — a bounded scan can't size the excess past `cap+1`, so a flood converges only over passes.

Between (1) and (3) the check is stale: another writer can append concurrently, and the loop is per-consumer code that each new bounded collection must re-derive (and re-test).

## Evidence from a consumer

webphone's transcript read model (`internal/readmodels/transcripts.go`, this train's migration) carries two caps (5,000 segments per extension, 1,500 per call). Its append path grew ~60 lines of converging-loop trim machinery; the in-source comment states the constraint directly:

> one bounded scan sees only cap+1 rows, so a flood converges over passes (each removes at least one row)

Snippets (cap 100), faxes (retention window), and enroll tokens (expiry sweep) each grew their own variant of the same scan-expunge loop. That is four hand-rolled implementations of one concept in one codebase.

## Proposed shape

Options, no strong preference:

1. **Fold-declared cap:** a plan-time option on the collection (e.g. `WithCap(N, OrderByField)`) so the engine trims as part of applying the insert — the store already owns the apply path, so the trim could be atomic with it.
2. **Store-level helper:** `Store.Trim(ctx, collection, maxRows, orderBy)` (or per-filter-group) that emits the `Remove` events for the oldest overflow rows.
3. **At minimum:** a documented cap-enforcement recipe in the metaengine docs — the converging-loop pattern (why `cap+1`, why repeat, the concurrency caveat) took real work to get right, and it is currently tribal knowledge.

Affects `metaengine/v4` v4.19.0; consumer stack Go 1.27, sqlite engine v4.5.3.

💘 Generated with Crush
