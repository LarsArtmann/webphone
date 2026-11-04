# go-output — library utilization audit

**Verdict:** NO-GO — a strong "write once, render anywhere" data-rendering library with no data-rendering consumer in this product; webphone's CLI prints single lines, and its one rich report (the boot report) is a designed product artifact, not tabular data. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** family tag `bdd/v0.38.7` (local checkout) · **Status:** not a dependency
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-output` (multi-module: root table/tree/graph builders + markdown, markup, delimited, serialization, graph, d2, plantuml, daghtml, tree, tui, nom, escape, testhelpers) |
| In webphone | absent (0 references) |

## What the library offers

"Write your data once, render anywhere": CQRS-shaped builder → snapshot → renderer over 16 formats (text tables, Markdown, HTML, CSV/TSV, JSON/YAML/TOML, D2, DOT, Mermaid, PlantUML, AsciiDoc, XML), plus NOM-style live terminal progress.

## Webphone overlap analysis (verified)

- CLI surface is three one-shot subcommands printing one line each: `catalog: wrote eventcatalog tree to %s` (cmd/webphone/catalog.go:69), `contracts` (contracts.go:41), the passkey enroll link (`fmt.Printf`, main.go:110).
- The failure path is cmd/webphone/bootreport.go:168-254 — the hand-written five-part operator contract (WHAT / REASSURE / WHY / FIX / ESCAPE + stack), pinned by docs/error-contract.md § Boot surface and its own tests. That copy is product design; a table renderer has nothing to say about it.
- No tabular data, no report generation, no terminal UI anywhere in the binary.

## Ruling

**Decline.** The natural home for go-output is tooling repos (linters, auditors, dashboards) that emit the same data in many formats. If webphone's CLI ever grows a data-bearing command (e.g. an operator `webphone threads --format=md` dump), that is the re-open trigger — and the root module is dependency-light enough to adopt then without ceremony.

**Refresh trigger:** any CLI subcommand whose output is structured data rather than a sentence.

## How to verify

```console
rg -rn "go-output" --type go go.mod cmd/ internal/ | wc -l       # expect 0
rg -n "fmt\\.Print" cmd/webphone/*.go | head -6                  # the one-line outputs
rg -n "bootStepError|REASSURE|WHY|FIX" cmd/webphone/bootreport.go | head -6
```
