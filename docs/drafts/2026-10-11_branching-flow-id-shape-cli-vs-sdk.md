# Draft: finding ID divergence CLI vs toolsdk

- Proposed title: `finding IDs: CLI is repo-relative, toolsdk lane is absolute — CLI-built baselines never match the SDK gate`
- Repo: LarsArtmann/branching-flow
- Kind: body-issue, --unsolicited, --ai-drafted

---

> [!NOTE]
> This filing was drafted by glm-5.3 via Crush executing the 2026-10-11 webphone ratification's route-upstream ruling (dedup-registry ruling 5 + FP policy), not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: the same finding carries a repo-relative path when reported by the `branching-flow` CLI and an absolute host path when reported through the toolsdk lane (BuildFlow) — a baseline saved from CLI output therefore matches nothing under the SDK gate, which then fails on already-reviewed noise while a staleness doctor insists the baseline is dead. Ask: normalize finding IDs to repo-relative paths at the construction seam, on both lanes.

branching-flow 0.6.4 + BuildFlow 25b07c4, LarsArtmann/webphone, 2026-10-11.

## Evidence

Same tree, same finding, two lanes:

```text
CLI  (branching-flow all . --format finding):
  branching-flow:DUPLICATE_TYPE:internal/config/config.go:179
  branching-flow:COMPOSITION_mixin:internal/asr/asr.go:85:6

SDK  (buildflow -s branching-flow --format finding):
  branching-flow:DUPLICATE_TYPE:/home/lars/projects/webphone/internal/config/config.go:179
  branching-flow:COMPOSITION_mixin:/home/lars/projects/webphone/internal/asr/asr.go:85:6
```

Two divergences in one ID: the path base, and the column (DUPLICATE_TYPE carries none on either lane, mixin rows carry `:6` — the shape inconsistency is minor, the path base is the killer).

Sequence that surfaced it (webphone, 2026-10-11):

1. Baseline built from CLI output via `buildflow baseline add` (19 exact-ID entries, all matching the CLI report).
2. `buildflow -s branching-flow` gate: exit 69, `7 findings at or above severity "warning" remain` — the 7 warning-severity rows of the very set just baselined.
3. Doctor, same run: `baseline staleness: none of the 19 baseline entries matched any remaining finding in each of the last 3 runs`.
4. `buildflow -s branching-flow --save-baseline` (SDK-side save) → same 19 findings, absolute-path IDs → gate green, staleness silent.

So the two lanes cannot share a baseline file today, and nothing warns the operator that the ID shapes differ — the failure reads as "your baseline is stale", sending the triager in the wrong direction.

## Ask

Whatever lane absolutizes (the SDK consumer passing absolute file args, or the analyzer echoing fset positions): the finding-ID constructor should relativize against the module root before building `tool:rule:file:line[:col]`. Toolsdk consumers cannot know the consumer's checkout layout, so the ID must be checkout-independent for baselines to be portable across machines at all — today's SDK IDs embed `/home/lars/...`, so even a second host running the same repo redoes the baseline.

While in there: make the column presence consistent (always emit `:col`, even when 0) so ID shapes are stable across rule families.

💘 Generated with Crush
