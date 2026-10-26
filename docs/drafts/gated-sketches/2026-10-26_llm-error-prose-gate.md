# GATED SKETCH — LLM error-prose gate (NOT implemented)

Status: sketch only, recorded by the SUPERB error-agents plan (M16).
Trigger to build: owner GO + a CI LLM-budget decision. Nothing here is
wired; this document exists so the future GO starts from a menu, not a
blank page.

## What it would be

A DETECT-ONLY CI lane that reads every user-facing error string and
flags prose that violates the error contract's five-part user standard
(what / reassure / why / fix / escape — where applicable): blaming the
user, leaking internals, dead-end messages ("something went wrong"),
or copy that dropped its de translation. It never edits, never blocks;
findings land as PR annotations or a job-summary table.

## Scan surface inventory (verified against the repo at sketch time)

| Surface | Where | Volume class | Notes |
|---|---|---|---|
| Island i18n values | `internal/web/views/i18n.go` (en/de maps) | ~hundreds of keys | parity tests exist; an LLM gate adds QUALITY checks, not parity |
| `http.Error` texts | `internal/server/**` (api, passkey_api, transcribe) | ~40 sites | the 2026-10-23 ruling keeps these plain-text; copy quality still applies |
| Boot report copy | `cmd/webphone/bootreport.go` | 6 classes × 5 parts | already English-only, golden-pinned; gate would check clarity drift |
| Server validation reasons | services (`Rejection` errors' messages) | distributed | the `<seam>.<op>` codes are the stable keys; prose rides alongside |
| HTML error panels | `internal/web/views` (`errorBanner`, `panelError`) | few | templ copy, en/de |

## Runtime / cost options

| Option | Runtime | Cost profile | Notes |
|---|---|---|---|
| A: local model via the ASR seam's box | self-hosted (Speaches-class server) | zero marginal, needs a runner with the model | repo doctrine prefers local (GDPR posture); adds infra |
| B: hosted API on-demand | CI calls an OpenAI-compatible endpoint | cents per run, only on diff'd copy files | needs API key secret + budget cap; sender-data review for copy (no caller data — only dev-authored strings) |
| C: pre-commit, agent-side | Crush/similar reads the contract during trains | zero CI cost, no enforcement | weakest guarantee, zero infra |

## Detect-only contract (if built)

- Reads ONLY dev-authored string literals — never journal lines, never
  caller content (the self-hosted caller-data doctrine holds).
- Output: a findings table (file, key/site, rule, suggested rewrite).
  No autofix, no merge gate, no i18n key changes without the parity
  tests staying green.
- Rules pinned as data (a rules file the gate loads), so rulings land
  in one reviewable diff.

## Trigger criteria (owner calls, recorded not made)

1. GO/NO-GO on building it at all.
2. If GO: option A vs B vs C.
3. If B: which provider, budget cap per month, key custody.
4. Which lane owns it (PR annotation vs job summary).
