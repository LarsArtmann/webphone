# Draft — docs-gates: four sweep defects (one reproduction each)

**Target repo:** `Artmann-Technologies/docs-gates` (own repo) · **Kind:** bug
report body · **Provenance:** unsolicited (AI-found; banner required) ·
**Status:** DRAFT — file only after `@LarsArtmann` review.

**Proposed title:** `audit-all's token-genericity flips verdicts as living docs grow, standalone verbs ignore .docs-gates.json, and two table defects corrupt sweeps silently`

```markdown
> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `docs-gates freshness --audit-all` re-classifies already-judged §f rows as "unharvested" when the living corpus grows past a token threshold, and a standalone verb ignores `.docs-gates.json`'s ratchets — so the same repo scores 72 findings through `run` and fails through `freshness`. Two separate table-handling defects (an escaped pipe splitting a struck cell, an unvalidated verdict value) then corrupt strike sweeps silently. Ask: make token genericity scale with corpus size explicitly, make standalone verbs read the config, and refuse the two table shapes.

## Problem 1 — DF brittleness: the audit's verdict flips as living docs grow

A §f row survives harvest only while its tokens stay rare in the LIVING corpus.
When the corpus passes ~8 living lines containing a row's token, that token
turns "generic", and a row judged harvested re-answers UNHARVESTED — on a
report nobody touched.

Observed on one repo over two days, all on the SAME row
(`docs/status/archived/2026-10-15_01-39_metaengine-dogfooding-train-p0-pilot.md` row 2, work landed
2026-10-15):

- 2026-11-07: audit_max ratchet moved 71 → 72 when living lines passed
  ~4,360; the row re-flagged with the tool's own note "accidental overlap,
  not a harvest".
- 2026-11-08: five new planning documents pushed the corpus further; the
  SAME row re-flagged again, plus one more row (74 findings vs the 72 pin).

Reproduction (webphone checkout):

    $ docs-gates run .                       # passes at its .docs-gates.json pins
    $ # append ~30 lines to TODO_LIST.md mentioning any §f row's tokens
    $ docs-gates run .                       # now reports +1 unharvested, same report

The counter-move is the consumer re-measuring its ratchet after every corpus
growth — which is a treadmill, not a gate. webphone now pins `generic_df: 22`
in `.docs-gates.json` and re-derives it per sweep (~0.5% of living lines), but
the verdict flip itself lives in the tool.

## Problem 2 — standalone verbs ignore `.docs-gates.json`

`docs-gates run .` honours the repo's ratchets; the SAME gate invoked
standalone ignores them:

    $ cat .docs-gates.json        # "audit_max": 72
    $ docs-gates run .            # PASS (72 ≤ 72)
    $ docs-gates freshness . --audit-all   # FAIL — evaluated against no pin

The wrapper therefore cannot hand a single verb to CI; it must pass
`--audit-max=N` explicitly (two config readers, one file). One reader, one
config.

## Problem 3 — an escaped pipe inside a struck cell splits the cell

`|` inside a code span within a table cell (`` `a\|b` ``) terminates the cell
at the WRONG position when a strike rewrite runs, silently mangling the row.
The consumer needs a pre-scan refusal BEFORE any strike is written — webphone
now ships one (`scripts/strike-safety.py`, `scan_rows()` flags ragged rows and
struck rows containing `\|`; 13 self-test cases), but the tool itself still
writes the mangled table.

## Problem 4 — the verdict cell accepts anything

A struck row whose verdict reads `UNMAPPED` (or any other garbage) passes
every gate — the strike machinery cannot tell a real verdict from noise, so a
sweep that LOSES its evidence table produces a green report. webphone now
ships a canonical-verdict allowlist (`strike-safety.py`, 13 self-test cases);
the tool should enforce the same shape at write time.

## Environment

- docs-gates at `2278406afdfd1b8df7674e2e594b6ea880694c90`
  (`flake = false` input, `cmd/docs-gates`), reporting `dev`.
- Local guards that work around 3/4: `scripts/strike-safety.py` +
  `.docs-gates.json` `generic_df` in
  `github.com/LarsArtmann/webphone` (linked for shape reference, not an ask).

💘 Generated with Crush
```

**Filing checklist (skill order):** verify-before-filing → own repo, all four
defects are first-hand with commands run this session; Gate 5 (already
reported?) → search `Artmann-Technologies/docs-gates` issues before filing.
github-voice → body register, banner + TL;DR above; run
`check-draft --kind body-issue --ai-drafted --unsolicited` before posting;
`gh issue create --body-file <file>`, never `--body-file -`.
