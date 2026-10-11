# Draft: composition + dupe ignore-comment suppression

- Proposed title: `composition + dupe analyzers never consult ignore comments — mixin/large-struct/duplicate-type findings are unsuppressible`
- Repo: LarsArtmann/branching-flow
- Kind: body-issue, --unsolicited, --ai-drafted

---

> [!NOTE]
> This filing was drafted by glm-5.3 via Crush executing the 2026-10-11 webphone ratification's route-upstream ruling (dedup-registry ruling 5 + FP policy), not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: every suppression directive silently does nothing for COMPOSITION_mixin, ANTI_PATTERN_large-struct and DUPLICATE_TYPE findings — `pkg/composition` and `pkg/dupe` never call the ignore-comment system, so the finding stays no matter where the directive sits. Ask: wire both analyzers into `core.IgnoreComments`, and add the missing `dupe` enum value so DUPLICATE_TYPE has a rule name at all.

branching-flow 0.6.4, run against LarsArtmann/webphone during the 2026-10-11 re-triage; every claim re-verified at HEAD today.

## What was tested

`COMPOSITION_mixin` on `internal/asr/asr.go` (`asr.Client`), six placements, finding stays in all six:

| Placement                                                                | Result         |
| ------------------------------------------------------------------------ | -------------- |
| `//nolint:branching-flow // reason` inline on the `type X struct {` line | still reported |
| same form, standalone comment line directly above                        | still reported |
| same form, on the first field line                                       | still reported |
| bare `//nolint:branching-flow` (no reason)                               | still reported |
| `//nolint:start:branching-flow` block wrapping the decl                  | still reported |
| native `//branching-flow:ignore composition <reason>` above the decl     | still reported |

## Why (source)

`core.ExtractIgnoreComments` / `IgnoreComments.Matching` are consulted by panicanalyzer, strongid, phantom (collision/deep/transpose), roleak and doanalyzerv2 — and by nothing else:

```console
$ rg 'IsIgnored|\.Matching\(|ExtractIgnoreComments' pkg --glob '*.go' | rg -v '_test|pkg/core'
pkg/panicanalyzer/analyzer.go
pkg/strongid/analyzer.go
pkg/phantom/{analyzer,collision,deep,transpose}.go
pkg/roleak/analyzer.go
pkg/doanalyzerv2/analyzer.go
```

`pkg/composition` and `pkg/dupe` have zero references (their only "ignore" hit is a comment about SkipDirs). The same is true of boolblind, splitbrain, contextguard, nakedreturn, flagparam and ifacecomplete — none of those produced findings in this repo today, so the composition/dupe pair is the evidenced ask, the rest a one-line sweep once the pattern exists.

Secondary gap: the `IgnoreCommentType` enum (`all, context, phantom, strong-id, bool, panic, composition, do, ro`, `pkg/core/ignore_comments.go:17`) has NO value for dupe. Even with plumbing, there is no rule name a `//branching-flow:ignore <rule>` directive could carry for DUPLICATE_TYPE.

## Consumer impact

webphone's 2026-10-11 re-triage had to park 19 type-node findings in a BuildFlow findings baseline, with plain-text rationale comments at the sites — a `//nolint` that does not suppress would mislead every future reader auditing suppressions. Working line-level suppression collapses that baseline to zero and makes the rationale live next to the code it defends.

💘 Generated with Crush
