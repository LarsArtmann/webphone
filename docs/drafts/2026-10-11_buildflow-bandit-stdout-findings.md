# Draft: bandit provider parses stdout as findings

- Proposed title: `bandit-check counts every stdout line (banner, metrics, blank) as a finding — 17 real issues surfaced as 335`
- Repo: LarsArtmann/BuildFlow
- Kind: body-issue, --unsolicited, --ai-drafted

---

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

## Reobservation 2026-11-12 (webphone, GLM-5.3 via Crush)

Still live, now with upstream naming:

- The class got a number: #64 (statix multi-line diagnostic format,
  CLOSED 2026-11-06). bandit rides the SAME `FindingsFromOutput` →
  `ParseColonFormat` → `parseLineFormat` fallback
  (`tools/providers/output_parser.go:111-120`); the #64 fix did not
  cover bandit, and `usageOrCrashMarkers` intentionally passes its
  report (it is not a usage banner).
- Fresh sighting: `buildflow -s bandit-check --format finding` reports
  244 findings, every row severity "warning", sharing ONE id
  (`bandit-check:bandit-check:b1308d29e4a43254`) with position
  `file: ".", line: 0`. BuildFlow's own report builder flags the
  shape: `patched findings missing identity fields tool=bandit-check
  missing_file=244`. The 6 real issues (5 Low, 1 Medium) are buried
  inside as `">> Issue: …"` message rows.
- bandit's `-f json` carries every identity field the report builder
  names as missing (filename, line_number, test_id, issue_severity,
  issue_confidence) — the ask stands unchanged.
- webphone side: the 5 real sites now carry `# nosec <CODE> - reason`
  (house practice since 2026-10-11), so the tree's real count is 0 and
  every remaining finding row is parser furniture.

### Reobservation addendum 2 (2026-11-12 16:3x): count reconciled to zero, locally reproducible

- Local runner: `nix run nixpkgs#python312Packages.bandit -- -r .`
  (bandit 1.9.4, same minor as the More-Info links). Fresh full-tree
  scan: **0 issues, 63 nosec'd sites** — the 15:57 sighting's six real
  issues are all suppressed at source with reasoned `# nosec`.
- The 6-vs-5 discrepancy is EXPLAINED: bandit's own metric block
  (Low 5 + Medium 1) was right; the sixth issue was B603 at
  `scripts/status-index.py:138` (B607 fires at the same line), and the
  BuildFlow report log carried only one of the two headers for that
  line (the #64-class line folding). Count the metric block, not the
  recovered headers, when triaging through the mangled output.
- New gotcha for the fix review: bandit's line-anchored `# nosec`
  accepts SPACE-separated ids (`# nosec B603 B607`). The COMMA form
  (`# nosec B603,B607`) left B603 firing while B607 went quiet
  (mechanism unclear — possibly per-test line attribution); switching
  to the space form suppressed both and took the tree to 0. Worth a
  line in the bandit docs or the BuildFlow fix's tests.
