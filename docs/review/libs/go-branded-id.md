# go-branded-id — adjacent library audit (closeout T23)

_Produced by F (subagents 18-22) of the 2026-11-22 full-fleet re-audit's fan-out (runbook step 7); every other line is the audit's own._

**Verdict:** the type-safety substrate is adopted exactly where it pays — six branded ids typed as `id.ID[Brand, nanoid.ID]`, the raw-value JSON wire contract, and the opt-in linter (BD003) clean since 2026-10-25. The two boundaries the 2026-09-24 audit flagged as gaps (hand-rolled parse, hand-rolled SQL conversion) are RULED DELIBERATE at this audit: the domain owns invariants the library's generic seams do not express — prefix-tolerant parsing with a 21-char nanoid shape check and a Corruption-classed error, and rows that must not mint ids from unvalidated strings.
**Pin:** v0.7.1 (direct) · **Latest:** v0.8.0 (2026-11-21; zero-API go-floor bump — adjudicated in the re-verify: ride the next deps train) · **Currency:** one minor behind, deliberately · **Status:** direct dep (local checkout rides v0.8.0 + linter/v0.6.1)
**Audited:** 2026-11-04 · 2026-11-22 re-verify · **Prior reports:** [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (68/100 — "parse hand-rolled", "SQL Valuer/Scanner unused"). **Movement: → 74 at this audit (+6 — the flagged gaps re-ruled deliberate; the linter adoption landed)**.
**Run all commands from the webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

## What the library offers

`id.ID[B, V]` generic strong id: `NewID`/`Get`/`IsZero`/`Equal`/`Compare`/`Or`/`Reset`; `String()` renders `Brand:value`; `MarshalText`/`UnmarshalText` (RAW value only — the prefixed form is not parsed for string ids); `Scan`/`Value` SQL pair (string path scans raw string/[]byte as-is, no validation); JSON v1/v2 with raw-value marshaling; `GoString`/`Format`; the `linter` module (BD-rules; webphone runs the opt-in BD003 brand-adoption scan, clean since 2026-10-25, every decided raw site carrying a single-line `//brandid-lint:ignore(BD003)` directive).

## Webphone usage (verified)

- **Six brands** (`internal/domain/ids.go`): `ThreadID`, `MessageID`, `AttachmentID`, `FaxID`, `ContactID`, `SnippetID` — all `id.ID[Brand, nanoid.ID]`, each with `Parse*ID` (client-supplied ids → 404 not panic), `Must*ID` (database rows → corruption is fatal), and `Generate*ID`.
- **The hand parse is the invariant owner** (`parseID`, ids.go:276): strips an optional `Brand:` prefix (the DB carries prefixed ids because the store writes `.String()`), enforces the 21-char nanoid shape, and codes failures `Corruption` (`domain.id`) — none of which `UnmarshalText` expresses (it mints an id from ANY raw string).
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

## 2026-11-22 re-verify

v0.8.0 (2026-11-21) is a floor bump and nothing else: the root `go` directive rises to `go 1.27` (flake go_1_26 → go_1_27 and the CI pins move in the same change) — ZERO API change; the motivation is JSON determinism (on a 1.27 toolchain, unpinned v1 legs silently compile v2, and the library now pins GOEXPERIMENT per CI leg). webphone is already there: `go 1.27.3` is its own floor, `encoding/json/v2` is stable in-tree (CONTRIBUTING: "no GOEXPERIMENT is needed"; the signal lane imports encoding/json/v2 directly) — the bump is compatible and zero-code. Adjudication: **ride v0.8.0 on the next deps train** (`go get github.com/larsartmann/go-branded-id@v0.8.0`, no code change — it removes the very v1/v2 ambiguity class webphone documented as a host-env leak on 2026-11-07); a standalone pin-bump commit is not warranted while v0.7.1's wire behavior is already pinned green by webphone's own raw-value contract tests. The linter moved v0.5.1 → v0.6.1 the same day — CLI ergonomics only (`-version`, `./...` path patterns, flag permutation, `-rules all`), no new rule, so the "linter ships a rule webphone lacks" trigger has not fired (68 BD003 ignore directives, each carrying its reason). [Unreleased] is empty in both sections. Adoption surface unchanged: six brands (ids.go:158-253), parseID still the invariant owner (ids.go:276 — ref refreshed), rows still parse through `Must*ID` — the store seam moved internal/store → internal/cqrs since 2026-11-04 (11 non-test call sites; verify line re-pointed) — and `Compare`/`Or` remain uncalled. Ruling unchanged.

## How to verify

```console
rg -n "type \w+ = id\.ID\[" internal/domain/ids.go        # expect 6 branded declarations
rg -n "func parseID" internal/domain/ids.go              # expect 1 — the invariant owner (prefix + 21-char + Corruption)
rg -n "MustThreadID|MustMessageID" internal/cqrs/ --type go -g '!*_test.go' | wc -l   # expect >= 10 (rows parse through the invariant; store seam moved internal/store → internal/cqrs since 2026-11-04 — scope re-pointed)
rg -n "\.Compare\(|\.Or\(" internal/ --type go | grep -vE "cmp\.|strings\.Compare|Date\.Compare|_test" | wc -l   # expect 0 (unused surface, ruled as-is; scope added 2026-11-09 — later trains brought stdlib cmp.Or/strings.Compare collisions into the tree, the branded surface itself stays uncalled)
rg -n "go-branded-id" go.mod                             # DIRECT pin v0.7.1 (v0.8.0 ride adjudicated: next deps train)
git -C /home/lars/projects/go-branded-id tag --list "v*" --sort=-v:refname | head -1   # expect v0.8.0 (root-module tags; `describe` now answers linter/vX.Y.Z — the sub-module out-tags the root)
```

## Index-cell adjudications (2026-11-27 tighten train)

### From the index verdict cell (go-branded-id)

the 09-24 gaps re-ruled DELIBERATE; v0.8.0 (2026-11-21) is a zero-API go-floor bump — ride the next deps train, no standalone commit warranted while the wire is pinned green
