> [!NOTE]
> This filing was drafted by GLM-5.3-flash via Crush from an AI-run investigation, not at my request. The evidence comes from migrating a production consumer (webphone) onto metaengine v4.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

**TL;DR:** `TypedReader.Scan` returns at most 100 rows by default and gives the caller no way to know it truncated — `ScanResult.HasMore` exists at the raw level but never surfaces through the typed reader. Ask: make the truncation visible at the typed call site (a HasMore, an exact-mode option, or anything in between).

## Problem

`metaengine/v4@v4.19.0`, `typed_reader_scan.go:16`:

```go
func (r *TypedReader[V]) Scan(ctx context.Context, opts ...ScanOption) ([]V, error) {
```

The default limit lives in `scanConfig{limit: 100}` (`typed_reader_scan.go:30`) and the raw path's `limit = 100` (`execute.go:286`). The godoc is honest about it ("a silent truncation for larger collections") and the knobs exist — `WithLimit(0)` for unbounded (`scan_options.go:97`), `WithDefaultLimit` store-wide (`planner_options.go:106`), `ScanPage` for cursor pagination. But the signature is the trap: `rows, err := reader.Scan(ctx)` reads as list-all, and nothing in the return shape distinguishes "exactly these 100 rows" from "the first 100 of 5,000".

The raw `Store` path knows better: `executeFilteredScan` returns a `ScanResult` with `HasMore` (`execute.go:312`). The typed reader computes the page metadata internally (`scanItems` returns `pageMeta`) and throws it away before returning `([]V, error)`.

## Evidence from a consumer

Migrating webphone's read models onto metaengine (v4.19.0) hit this in every bounded-collection enforcement: transcript trims (per-extension 5,000 / per-call 1,500 caps), the enroll-token expiry sweep, fax retention. Each one had to become a converging loop — scan `cap+1` by sort key, expunge the tail, repeat — because a single bounded scan cannot size the excess: `Scan(WithLimit(cap+1))` answers `cap+1` rows whether 3 or 3,000 rows overflow, and `Scan(WithLimit(0))` on a hot 100k-row collection is not a default anyone should reach for by accident. The loops work (proven at 501 and 5,001 seeded rows), but the reason they must loop is precisely that `Scan` won't say "there is more".

## Why the existing options don't cover it

- `WithLimit(0)` means unbounded — "zero means no limit" is its own falsy-default trap for a caller who reaches for it as "no option set".
- `WithDefaultLimit` re-pins the ceiling store-wide but doesn't add a truncation signal; a scan that returns exactly `defaultLimit` rows is still indistinguishable from an exact one.
- `ScanPage` is pagination-shaped (caller opts in), not a list-all shape.

## Proposed shape

Any of these would close the gap; no strong preference:

1. `Scan` returns a truncation signal (`([]V, bool, error)` or a small `ScanMeta{HasMore bool}`) — `pageMeta` is already computed internally.
2. A `WithExact()` option that errors when the result was truncated (turns the silent case into a loud one without breaking the signature).
3. At minimum, upgrade the godoc warning into a `WARNING:`-boxed contract on every typed scan entry point — today it lives only on `TypedReader.Scan`, not on the filtered/raw scan paths a caller might read first.

Affects `metaengine/v4` v4.19.0; consumer stack Go 1.27, sqlite engine v4.5.3.

💘 Generated with Crush
