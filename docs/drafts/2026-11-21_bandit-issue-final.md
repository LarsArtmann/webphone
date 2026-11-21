> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: the bandit provider runs bandit bare and chews its human-readable stdout line-by-line into findings, so a tree with ZERO real issues still reports 339 "findings" — all console furniture sharing one id. Ask: run with `-f json` and map actual issues to findings.

BuildFlow `ceb25caa1` (2026-11-20), webphone 2026-11-21.

## Problem

`tools/providers/python_tools.go:45-53` — the provider is a bare
FormatterProvider call with no format flags:

```go
func NewBanditCheckProvider(runner Runner) *dtool.Tool {
	return NewFormatterProvider(
		config.ToolBanditCheck, "bandit", nil, ".", runner,
		...
	)
}
```

The output then rides the generic path
(`tools/providers/output_parser.go:111-120`): `FindingsFromOutput` →
`ParseColonFormat` → line-based `parseLineFormat` fallback — one finding
per stdout line.

Fresh sighting (webphone, 2026-11-21, buildflow `cb5471ee8`):
`buildflow -s bandit-check --format finding` reports **339 findings**,
every row severity "warning", sharing ONE id
(`bandit-check:bandit-check:b1308d29e4a43254`) with position
`file: ".", line: 0`. BuildFlow's own report builder flags the shape:
`patched findings missing identity fields tool=bandit-check
missing_file=244`. The messages are literally bandit's console
furniture:

```text
bandit-check: "[main]\tINFO\tprofile include tests: None"
bandit-check: "Run started:2026-10-11 04:41:52.446847+00:00"
bandit-check: "------------------------------------" (rules header)
bandit-check: ">> Issue: [B310:blacklist] …"      (the real ones)
bandit-check: "Location: ./scripts/perf-baseline.py:23:13"
bandit-check: "Code scanned:" / "Run metrics:" …
```

Context:

- webphone's tree is `# nosec`-clean: a direct bandit run (`bandit -r .`,
  1.9.4) reports **0 issues, 63 nosec'd sites** — so all 339 rows are
  parser furniture. It makes the tool's count meaningless as a drift
  signal and was the single largest blocker when evaluating
  `strict: true` on this repo (the original sighting: 17 real issues
  surfaced as 335 findings).
- The multi-line-diagnostic class got #64 (statix, CLOSED 2026-11-06),
  but that fix did not cover bandit, and `usageOrCrashMarkers`
  intentionally passes bandit's report through (it is not a usage
  banner).
- The 6-vs-5 real-issue count in mangled output is a #64-class line
  fold: bandit's own metric block (Low 5 + Medium 1) is the honest
  count; B603+B607 firing at the same line shared one recovered header.

## Proposed fix

Invoke bandit with `-f json` (bandit ≥1.7) and map each `results[]`
entry to one finding — the JSON carries every identity field the report
builder names as missing (`filename`, `line_number`, `test_id`,
`issue_severity`, `issue_confidence`); nosec suppression counts stay
visible via the JSON's metrics. vulture on this repo reports 3 rows and
looks fine, so bandit is the outlier worth fixing first.

One extra for the fix's tests: bandit's line-anchored `# nosec` accepts
SPACE-separated ids (`# nosec B603 B607`); the comma form left B603
firing while B607 went quiet (mechanism unclear — possibly per-test line
attribution).

💘 Generated with Crush
