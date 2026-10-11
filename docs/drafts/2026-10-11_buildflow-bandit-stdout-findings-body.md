> [!NOTE]
> This filing was drafted by glm-5.3 via Crush while executing the owner's 2026-10-11 fix-all mandate on webphone, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: the bandit provider runs bandit bare and chews its human-readable stdout line-by-line into findings, so a run with 17 real issues reports 335 "findings" — and a fully `# nosec`-clean tree still reports ~330. Ask: run with `-f json` and map actual issues to findings.

BuildFlow 25b07c4 (installed), webphone 2026-10-11.

## Evidence

`tools/providers/python_tools.go:45-52` — the provider is a bare
FormatterProvider call with no format flags:

```go
func NewBanditCheckProvider(runner Runner) *dtool.Tool {
	return NewFormatterProvider(
		config.ToolBanditCheck, "bandit", nil, ".", runner,
		...
	)
}
```

Result on webphone (17 real issues at the time): 335 findings, whose
messages are literally bandit's console furniture:

```text
bandit-check: "[main]\tINFO\tprofile include tests: None"
bandit-check: "Run started:2026-10-11 04:41:52.446847+00:00"
bandit-check: "------------------------------------" (rules header)
bandit-check: ">> Issue: [B310:blacklist] …"      (the real ones)
bandit-check: "Location: ./scripts/perf-baseline.py:23:13"
bandit-check: "Code scanned:" / "Run metrics:" …
```

Every multi-line issue block yields ~6 finding rows (issue + severity +
CWE + more-info + location + blank separator), the banner adds ~5, the
rules dump adds one per rule. After triaging all 17 real rows to
`# nosec`-with-reason, the tool STILL reports ~330 findings (all
furniture) — detect-only, but it makes the tool's count meaningless as
a drift signal and it was the single largest blocker when evaluating
`strict: true` (warning-gate) on this repo: 335 phantom warnings.

## Ask

Invoke bandit with `-f json` (bandit ≥1.7 supports it) and map each
`results[]` entry to one finding (rule `issue_text`/`test_id`,
position `filename:line_number:col_offset`); keep nosec suppression
count visible via the JSON's metrics if useful. Same pattern likely
applies to any other FormatterProvider-fed tool whose native output is
multi-line human text — vulture on this repo reports 3 rows and looks
fine, so bandit may be the outlier worth fixing first.

💘 Generated with Crush
