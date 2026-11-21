# Draft — BuildFlow: `--save-baseline` writes the vanish remainder, not the findings snapshot

**Target repo:** `LarsArtmann/BuildFlow` (own repo) · **Kind:** bug
report body · **Provenance:** unsolicited (AI-found; banner required) ·
**Status:** DRAFT — file only after `@LarsArtmann` review (the webphone
TODO row's owner gate).

**Re-route note (2026-11-21):** the TODO row that owns this finding
("fold the `--save-baseline` remainder-write footgun into the
docs-gates issue draft") pointed at the wrong repo — the flag is
BuildFlow's (`docs-gates` has no `--save-baseline`), so the draft
lives here and the docs-gates draft stays at its four defects.

**Proposed title:** `--save-baseline after the stay-vanish change writes only the unaccepted remainder — against a fully-accepted baseline it silently empties the file`

````markdown
> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `--save-baseline` now runs after the stay-vanish transform
removed the baseline-accepted findings from the workflow result, so the
re-save writes only the unaccepted REMAINDER over the existing file —
against a fully-accepted (green) set that is 0 findings, silently
destroying the accepted-noise baseline. Ask: make the save write the
pre-vanish snapshot (the documented "superset of remaining" contract),
pinned by a regression test.

## Problem

GOTCHAS #200 documents the flag's contract: "`--save-baseline` writes
`workflowResult.BuildReport()` (ALL original detect findings, a
SUPERSET of remaining — extra baseline entries are harmless no-ops)".
The P-11 stay-vanish change (2026-10-12) broke that ordering:

```go
// internal/cli/workflow_tail.go (ceb25caa1)
	stampBaselineSuppression(config, workflowResult)   // → applyBaselineVanish
	...
	handleBaselineComparison(cmd, workflowResult)      // → --save-baseline
```
````

`applyBaselineVanish` mutates the step reports in place —

```go
// execution/workflow_result_filter.go:51-52
		filtered := gofinding.NewReportFromFindings(result.Report.Original.Tool, kept)
		result.Report.Original = filtered
```

— so `BuildReport()` (`ModelReportToFinding` returns `report.Original`)
sees the post-vanish remainder, and `SaveBaseline` overwrites the
configured baseline file with it.

Live reproduction at `cb5471ee8` (contains the vanish change; scratch
git repo, 23 findings):

```console
$ buildflow --save-baseline          # no baseline configured yet
INFO  Baseline saved to .buildflow-baseline.json (23 findings)
$ printf 'baseline: .buildflow-baseline.json\n' >> .buildflow.yml
$ buildflow --save-baseline          # all 23 accepted → gate green
INFO  baseline-accepted findings removed from machine output (stay-vanish) removed=23 ...
INFO  Baseline saved to .buildflow-baseline.json (1 findings)   # the 1 NEW finding only
```

The 23 accepted entries are gone from the file; against a green set
with no new findings the write is 0 entries, and the next run
re-surfaces every accepted finding as new. Observed on webphone
2026-11-08: a 124-entry baseline came back as 0 findings; recovery was
only from git history (`ed9ccb63`).

## Fix / Proposal

1. Save the PRE-vanish findings: either run `handleBaselineComparison`
   before `stampBaselineSuppression`, or snapshot the originals when
   vanish applies and save that. The remainder is what the gate
   prints, not what the baseline should contain.
2. Regression test: with a baseline that accepts every finding,
   `--save-baseline` must keep those entries (and still append new
   ones).

💘 Generated with Crush

```
**Filing checklist (skill order):** verify-before-filing → all five
gates passed this session: source traced end-to-end at `ceb25caa1`
(order + in-place mutation + report conversion quoted above), live
reproduction at `cb5471ee8` (contains the vanish change `90f1d6771`,
2026-10-12), the documented contract it contradicts (GOTCHAS #200),
Gate 5: no existing issue (only #28/#40, different defects — searched
"baseline" and "vanish save remainder", open and closed). Re-run the
line-accurate greps at the exact commit you post from. github-voice →
run `check-draft --kind body-issue --ai-drafted --unsolicited` before
posting; `gh issue create --body-file <file>`, never `-`.
```
