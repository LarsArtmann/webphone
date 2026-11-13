# Draft — erraudit `tree`: same-name dedupe hides same-named sentinels across packages, and the surviving pick is arbitrary

**Target repo:** `LarsArtmann/erraudit` (own repo) · **Kind:** bug report
body · **Provenance:** unsolicited (AI-found; banner required) ·
**Status:** DRAFT — file only after `@LarsArtmann` review. Routed from
the owner sitting pack §D17 (recommendation on record) + webphone
`docs/status/archived/2026-09-29_02-26_erraudit-tree-investigation-and-vendorhash-repair.md`
(the original incident).

**Proposed title:** `tree: same-name sentinels collapse to one row (3 of 4 ErrDisabled declarations hidden) — dedupe should not merge distinct declaration sites`

```markdown
> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `erraudit tree` shows ONE row for a sentinel name even when several packages declare their own, so 4 of 6 same-named sentinel declarations in the consuming repo are invisible — and which declaration survives is arbitrary (the pick site flipped between runs in the 2026-09-29 investigation). Ask: qualify same-name rows by package (or show every declaration site) and make the pick deterministic.

## Problem

Consumer: LarsArtmann/webphone at the drafting tip. Declared sentinels
(`rg 'Err(Disabled|NotFound) = errors.New' internal/ --glob '!*_test.go'`):

    internal/asr/asr.go:61      var ErrDisabled = errors.New("asr not configured")
    internal/crm/client.go:38   var ErrDisabled = errors.New("crm not configured")
    internal/pbx/client.go:57   var ErrDisabled = errors.New("phone api not configured")
    internal/push/push.go:81    var ErrDisabled = errors.New("push not configured")
    internal/crm/client.go:45   var ErrNotFound = errors.New("crm has no contact for this number")
    internal/store/messages.go:17 var ErrNotFound = errors.New("not found")

`erraudit tree ./internal/...` renders:

    ├── ErrNotFound [sentinel] @internal/crm/client.go:45 · 5 checks
    ├── ErrNoContext [sentinel] @internal/crm/context.go:19 · 1 check
    ├── ErrDisabled [sentinel] @internal/pbx/client.go:57 · unreferenced

Six declarations, two rows. The asr/crm/push `ErrDisabled` declarations and
the store `ErrNotFound` are invisible — and the shown pick is the worst
possible one for triage here: `pbx.ErrDisabled` reads as `unreferenced`
while the three same-named sentinels that ARE classified and checked
everywhere never appear at all.

## Mechanism (source-verified)

`cmd/erraudit/cmd/output/output_strategy.go` calls
`findingapi.DeduplicateFindings` (pkg/finding/bridge_filter.go), which
feeds every finding into one `go-finding` report and iterates
`findingsdk.MergeIter(..., WithDeduplication(true))`. The dedup key is the
finding identity the tree display groups by — sentinel NAME — so distinct
declaration sites of the same name collapse to a single row, and the
survivor depends on iteration order over the report's contents (the
2026-09-29 investigation observed the pick site flip between runs: crm
rows won without `--type-aware`, pbx/store rows with it).

## Why this misleads triage

`tree` reads as the full inventory of error surface. A reviewer auditing
"which sentinels exist and where are they checked" sees one `ErrDisabled`
and concludes the seam has a single disabled-sentinel; the three other
seams' classifications, family tests and checks render as nothing. The
flat display may be deliberate for duplicate findings — but declaration
SITES are not duplicates; each is a distinct fact.

## Ask

Two parts, either acceptable alone:

1. `tree` should show every declaration site, qualified by package when a
   name is shared (`crm.ErrDisabled`, `pbx.ErrDisabled`, …) — the display
   fix at the erraudit layer, independent of the dedup key.
2. If dedupe stays, make the survivor deterministic (stable sort before
   the pick) so a re-run at least shows the same row.

Reproduction: build erraudit at the filing commit, run `tree ./internal/...`
in webphone, compare against the `rg` above. The counts have diverged
further since the 2026-09-29 investigation (3 of 7 hidden then, 4 of 6 now).

💘 Generated with Crush
```

## Verification ledger (for the filing ritual)

- Gate 1 (reproduced at HEAD): erraudit `c78b8db5` binary built 2026-11-08
  23:33; two consecutive `tree ./internal/...` runs byte-identical
  (/tmp/tree-run1.txt vs /tmp/tree-run2.txt) — the deterministic-hiding half
  is live; the flip half observed 2026-09-29.
- Gate 2 (not already fixed): tree output at HEAD still collapses; no
  package qualification in the rendered rows.
- Gate 5 (already exists): `gh issue list --repo LarsArtmann/erraudit
  --state all` at drafting time shows no tree-dedupe issue (#11/#10/#3/#2
  open, all unrelated; #14 close-out class).
- Mechanism: pkg/finding/bridge_filter.go `DeduplicateFindings` →
  go-finding `MergeIter(WithDeduplication(true))`; call site
  cmd/erraudit/cmd/output/output_strategy.go:287.
