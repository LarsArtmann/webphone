# go-branded-id — adjacent library audit (closeout T23)

**Verdict:** the type-safety substrate is adopted exactly where it pays — six branded ids typed as `id.ID[Brand, nanoid.ID]`, the raw-value JSON wire contract, and the opt-in linter (BD003) clean since 2026-10-25. The two boundaries the 2026-09-24 audit flagged as gaps (hand-rolled parse, hand-rolled SQL conversion) are RULED DELIBERATE at this audit: the domain owns invariants the library's generic seams do not express — prefix-tolerant parsing with a 21-char nanoid shape check and a Corruption-classed error, and rows that must not mint ids from unvalidated strings.
**Pin:** v0.7.1 (direct; local checkout rides v0.7.1 + linter/v0.5.1) · **Latest:** v0.7.1 · **Currency:** current · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (68/100 — "parse hand-rolled", "SQL Valuer/Scanner unused"). **Movement: → 74 at this audit (+6 — the flagged gaps re-ruled deliberate; the linter adoption landed)**.
**Run all commands from the webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

## What the library offers

`id.ID[B, V]` generic strong id: `NewID`/`Get`/`IsZero`/`Equal`/`Compare`/`Or`/`Reset`; `String()` renders `Brand:value`; `MarshalText`/`UnmarshalText` (RAW value only — the prefixed form is not parsed for string ids); `Scan`/`Value` SQL pair (string path scans raw string/[]byte as-is, no validation); JSON v1/v2 with raw-value marshaling; `GoString`/`Format`; the `linter` module (BD-rules; webphone runs the opt-in BD003 brand-adoption scan, clean since 2026-10-25, every decided raw site carrying a single-line `//brandid-lint:ignore(BD003)` directive).

## Webphone usage (verified)

- **Six brands** (`internal/domain/ids.go`): `ThreadID`, `MessageID`, `AttachmentID`, `FaxID`, `ContactID`, `SnippetID` — all `id.ID[Brand, nanoid.ID]`, each with `Parse*ID` (client-supplied ids → 404 not panic), `Must*ID` (database rows → corruption is fatal), and `Generate*ID`.
- **The hand parse is the invariant owner** (`parseID`, ids.go:221): strips an optional `Brand:` prefix (the DB carries prefixed ids because the store writes `.String()`), enforces the 21-char nanoid shape, and codes failures `Corruption` (`domain.id`) — none of which `UnmarshalText` expresses (it mints an id from ANY raw string).
- **The hand SQL conversion is the same ruling**: the store boundary stringifies into params and parses rows through `Must*ID` so every id read back from sqlite passes the shape check. The library's `Scan` would mint ids from unvalidated strings — a weaker seam, not a missing adoption. The 09-24 finding "SQL Valuer/Scanner unused — the store boundary converts by hand" described real code but mis-priced it: the conversion IS the validation.
- **userauth does not ride this library**: the identity seam uses `usermgmt.UserID` (ULID) parsed STRICTLY at the handler door (`parsePasskeyUserID`; AGENTS strong-id ruling 2026-10-05).
- **Unused surface**: `Compare`/`Or`/`Reset` (zero production uses), `UnmarshalText` (cannot read the prefixed wire form for string ids), `Scan`/`Value` (by the invariant ruling above).

## Fit analysis

| Candidate                              | What it buys             | Why not (yet)                                                                                                                              |
| -------------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Replace `parseID` with `UnmarshalText` | one less hand function   | loses prefix tolerance, shape validation, and the Corruption class — the domain's invariant set, enforced at the one door client ids enter |
| Ride `Scan`/`Value` in the store       | typed SQL boundary       | mints ids from unvalidated strings; the current conversion validates every row-read id at the same shape bar                               |
| Adopt `Compare`/`Or`                   | sort/fallback ergonomics | no call site wants them today                                                                                                              |

## Ruling

Adoption stands as-is; the 09-24 score's deductions were premise-inverted (the hand parse is not an adoption gap but the domain invariant's home). **Refresh triggers:** the library grows a parse-with-rules hook (a `ParseFunc` seam `UnmarshalText` consults — then webphone's invariant could migrate INTO the library seam); a second Go consumer imports `internal/domain` (the ratified modularization trigger — the brands would need their own module); the linter ships a rule webphone currently lacks.

## How to verify

```console
rg -n "type \w+ = id\.ID\[" internal/domain/ids.go        # expect 6 branded declarations
rg -n "func parseID" internal/domain/ids.go              # expect 1 — the invariant owner (prefix + 21-char + Corruption)
rg -n "MustThreadID|MustMessageID" internal/store/messages.go | wc -l   # expect > 5 (rows parse through the invariant)
rg -n "\.Compare\(|\.Or\(" internal/ --type go | wc -l   # expect 0 (unused surface, ruled as-is)
rg -n "go-branded-id" go.mod                             # DIRECT pin v0.7.1
git -C /home/lars/projects/go-branded-id describe --tags --abbrev=0   # v0.7.1 vs pin v0.7.1
```
