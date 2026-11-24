# go-output — library utilization audit

**Verdict:** NO-GO — a strong "write once, render anywhere" data-rendering library with no data-rendering consumer in this product; webphone's CLI prints single lines, and its one rich report (the boot report) is a designed product artifact, not tabular data. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** family tag `bdd/v0.38.7` (unchanged 2026-11-22; [Unreleased] empty) · **Status:** not a dependency
**Audited:** 2026-11-04 · 2026-11-22 re-verify · **Prior reports:** none (first audit).

## Status

|             |                                                                                                                                                                                                   |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Module      | `github.com/larsartmann/go-output` (multi-module: root table/tree/graph builders + markdown, markup, delimited, serialization, graph, d2, plantuml, daghtml, tree, tui, nom, escape, testhelpers) |
| In webphone | absent (0 references)                                                                                                                                                                             |

## What the library offers

"Write your data once, render anywhere": CQRS-shaped builder → snapshot → renderer over 16 formats (text tables, Markdown, HTML, CSV/TSV, JSON/YAML/TOML, D2, DOT, Mermaid, PlantUML, AsciiDoc, XML), plus NOM-style live terminal progress.

## Webphone overlap analysis (verified)

- CLI surface is three one-shot subcommands printing one line each: `catalog: wrote eventcatalog tree to %s` (cmd/webphone/catalog.go:80), `contracts` (contracts.go:44), the passkey enroll link (`fmt.Printf`, main.go:113) — line refs refreshed 2026-11-22.
- The failure path is cmd/webphone/bootreport.go:165-190 (`renderBootFailure`; the five parts WHAT / REASSURE / WHY / FIX / ESCAPE at 170-174, + stack), pinned by docs/error-contract.md § Boot surface and its own tests. That copy is product design; a table renderer has nothing to say about it.
- No tabular data, no report generation, no terminal UI anywhere in the binary.

## Ruling

**Decline.** The natural home for go-output is tooling repos (linters, auditors, dashboards) that emit the same data in many formats. If webphone's CLI ever grows a data-bearing command (e.g. an operator `webphone threads --format=md` dump), that is the re-open trigger — and the root module is dependency-light enough to adopt then without ceremony.

**Refresh trigger:** any CLI subcommand whose output is structured data rather than a sentence.

## 2026-11-22 re-verify

v0.38.7 (tagged 2026-10-19) is still the family head — no drift — and [Unreleased] reads "Nothing yet" in both sections. The v0.38.7 delta itself was release engineering (the daghtml coverage gate satisfied at the tag level; a website TypeScript pin repaired so tags can redeploy) — motion in that repo stays tooling-side, the report's premise in miniature. webphone's side is unchanged: the CLI still prints one-liners (cmd/webphone/catalog.go:80, contracts.go:44, main.go:113) and the boot report is still the designed five-part artifact (bootreport.go:170-174), pinned by docs/error-contract.md § Boot surface (the `bootStepError` symbol cited by the original verify line has since been retired; the alternation now carries ESCAPE). No structured-data subcommand exists (`go-output`: 0 references) — the flip trigger has not fired. Verdict unchanged.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -rn "go-output" --type go go.mod cmd/ internal/ | wc -l       # expect 0
rg -n "fmt\\.Print" cmd/webphone/*.go | head -6                  # the one-line outputs
rg -n "REASSURE|WHY|FIX|ESCAPE" cmd/webphone/bootreport.go | head -6
```
