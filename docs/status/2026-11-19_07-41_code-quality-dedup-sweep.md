# Code-quality + dedup sweep — 2026-11-19 07:41 UTC

> Snapshot-rot header: **claims verified as of 07:41 UTC** (2026-11-19). Push/CI statements rot fastest; every verdict below was measured against the working tree at HEAD `25301790` (clean tree at sweep start, one fix landed mid-sweep — see §Fixes).

Combined execution of the `code-quality-scan` + `deduplicate-code` skills against the repo's ratified process. Output-format deviation, noted for the record: the scan skill's default artifact is a styled HTML dashboard under `docs/reviews/`; this sweep was tasked to deliver ONE markdown report here instead — the dashboard structure (stat cards → badge table → sorted findings) is preserved in markdown form.

## Verdict summary

**GREEN across every gate; zero harmful clones; one Low finding (dead const) found and fixed.**

| Gate                                                                                                                           | Result                                              | Exit  |
| ------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------- | ----- |
| Build (`go build ./...`, devshell)                                                                                             | 38 packages compile clean                           | 0     |
| `go vet ./...`                                                                                                                 | no findings                                         | 0     |
| `golangci-lint run ./...`                                                                                                      | 1 finding (unused const) → **fixed** → re-run clean | 1 → 0 |
| Go tests (`env -u GOEXPERIMENT go test -count=1 ./...`)                                                                        | 38/38 packages `ok`, 0 fail                         | 0     |
| Island JS suite (node --test, 33 specs)                                                                                        | **364/364 pass, 0 fail**                            | 0     |
| BuildFlow full run `20261119-064013-1` (the in-flight baseline run; 75 steps, 6m25s exec — real execution, not a cache replay) | **75/75 pass, 0 fail, 100%**                        | 0     |
| art-dupl clone gate (CI-pinned **v0.7.2**, `-t 1`, vs `.art-dupl-baseline.json`)                                               | **0 new clones** (baseline: 108 groups)             | 0     |

Cache discipline (scan-skill step 5): the test suite ran `-count=1`; the art-dupl gate compares live-scan hashes against the committed baseline (no result cache); buildflow's 06:40 full run carries 6m25s of real step execution and its `findings_gate` numbers were read from the state DB (`buildflow history`), not from a streamed log tail.

## Tool-by-tool results (exact commands + exit codes)

All commands run inside the devshell (`nix develop -c …`) per repo discipline; the host `go` is below the 1.27.3 floor.

| # | Command                                                                                                                                                                                     | Exit | Result                                                                                                                                  |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- | --------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | `nix develop -c go build ./...`                                                                                                                                                             | 0    | clean                                                                                                                                   |
| 2 | `nix develop -c go vet ./...`                                                                                                                                                               | 0    | clean                                                                                                                                   |
| 3 | `nix develop -c golangci-lint run ./...`                                                                                                                                                    | 1    | 1 finding: `internal/arch/model_api_test.go:36:7: const modelAPIRoot is unused (unused)` — see §Fixes; package re-run after fix: exit 0 |
| 4 | `nix develop -c env -u GOEXPERIMENT go test -count=1 ./...`                                                                                                                                 | 0    | 38/38 `ok` (slowest: cqrs 30.5s, server 18.0s)                                                                                          |
| 5 | `nix develop -c nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs`                                                                             | 0    | 364 tests / 364 pass / 0 fail                                                                                                           |
| 6 | `nix develop -c buildflow history`                                                                                                                                                          | 0    | full run `20261119-064013-1`: 75 steps, 75 pass / 0 fail, 100% (1h ago at read time)                                                    |
| 7 | `nix develop -c go run github.com/LarsArtmann/art-dupl/cmd/art-dupl@v0.7.2 check . -t 1 --baseline-path .art-dupl-baseline.json`                                                            | 0    | **"Found total 0 clone groups … No new clones detected (baseline: 108 groups)"**                                                        |
| 8 | `nix develop -c go run github.com/LarsArtmann/art-dupl/cmd/art-dupl@v0.7.2 . -t 1` (live inventory; note: v0.7.2 has no `scan` subcommand — a stray `scan` arg is stat-rejected harmlessly) | 0    | "Detected 1902 clone groups, 108 shown (1303 non-actionable, 491 filtered suppressed)" — shown set == baseline set                      |

Threshold note: the dedup skill defaults to `-t 5` when no threshold is given; this repo has **ratified `-t 1`** (2026-10-11 sitting, registry §Protocol) — the deepest threshold — and the sweep ran at exactly that, with the CI-pinned v0.7.2 detector (never the ambient host v0.9.x, which reports a different group universe).

## Dedup sweep-log table (clone groups → ruling)

`check` = 0 new, so **no new rulings were produced**; the 108 shown groups are the pre-adjudicated baseline. Sampled occurrences re-read at HEAD and attributed against standing registry rows:

| Clone group (live sighting at HEAD)                                                 | Ruling                  | Reason (registry row)                                                                                              |
| ----------------------------------------------------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `cqrs/faxesstore.go` + `threadsregister.go` ×2 — `ProjectionHostOptions` appends ×3 | intentional             | machine-table registration idiom; genericizing decode needs reflection over the event union (2026-11-09)           |
| `internal/matrix/receiver.go` !alive blocks + `signal/receiver.go` loop-exit        | intentional             | shared logic IS extracted (`chatloop.Backoff`); residue is the irreducible call shape (2026-11-04 T17)             |
| `cmd/webphone/main.go` vs `internal/app/app.go` — `switch len(e.errs)` join body    | intentional             | typed multi-error aggregates; extraction deferred to the NAMED trigger (a 4th aggregate; exactly 3 verified 11-04) |
| `internal/web/views/admin.templ` muted-span / chip pairs                            | intentional             | admin-train markup-shape coincidence class; upstream art-dupl#14 filed (2026-11-05)                                |
| `cqrs/faxesstore.go` vs `threadssweep.go` retention-sweep err-guard fragments       | intentional             | guard/trivia, 3–6 lines; a helper couples two aggregates' retention internals (2026-11-09)                         |
| `asr/client.go` vs `asr/google.go` — `providerRoundTrip` call pair                  | intentional             | call-sites-are-the-dedup, panelHead class (2026-10-18)                                                             |
| `cqrs/callnotes.go` / `callreviews.go` guard prologues                              | intentional             | per-seam error-code parity — the registry forbids burying `<seam>.<op>` codes behind arguments (2026-11-11)        |
| remaining ~100 shown groups                                                         | intentional (baselined) | every hash is in `.art-dupl-baseline.json` with a registry ruling; `check` proves none are new                     |

**Harmful clone count: 0.** Detector FP posture: the short-span `.templ` matching class is already FILED upstream (art-dupl#14); nothing new to file this sweep.

## Sorted findings list

| # | Severity                          | File                              | Line | Issue                                                                                                                                                     | Disposition                                                                                                                                               |
| - | --------------------------------- | --------------------------------- | ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Low                               | `internal/arch/model_api_test.go` | 36   | `const modelAPIRoot = "."` declared, never used (`unused`); comment claimed "resolved against the repo root below" but `repoRoot()` uses `runtime.Caller` | **FIXED this sweep** — dead const deleted, doctrine comment preserved above `repoRoot`; arch tests + package lint re-run green                            |
| 2 | Low (tracked posture, non-gating) | repo-wide (47 findings)           | —    | buildflow `findings_gate`: **47 file-size-check errors**                                                                                                  | ATTRIBUTED, not re-derived: the tracked **D10** posture (size budget ratchet owned by its TODO/plan row) — not a new regression; gate itself passed 75/75 |
| 3 | Low (standing class)              | island-tests + docs/contracts     | —    | buildflow `findings_gate`: **1 jscpd** finding                                                                                                            | standing island-test bootstrap-harness / generated-vs-generated class (registry 2026-11-11 rows; `.jscpd.json` owns the work/ exclusion)                  |
| 4 | Low (standing class)              | docs corpus                       | —    | buildflow `findings_gate`: **1 docs-gates** finding                                                                                                       | docs battery ratchet class (`.docs-gates.json` owns ratchets; markdown-lint detect-only)                                                                  |
| 5 | Info                              | `golangci-lint` stderr            | —    | `unknown linters in //nolint directives: branching-flow, erraudit`                                                                                        | noise: those are BuildFlow-internal linters whose directives golangci-lint cannot know; no action                                                         |

No Critical, High, or Medium findings. The irony of #1 is noted on the record: the dead const lived in the very test that rejects dead exported model API.

## Fixes applied (and verification)

- `internal/arch/model_api_test.go`: deleted the unused `modelAPIRoot` const; the file-level doctrine comment now sits directly above `repoRoot` (whose `runtime.Caller` logic is the real root resolution). Verified: `go test -count=1 ./internal/arch/` → `ok … 0.612s` (exit 0); `golangci-lint run ./internal/arch/` → exit 0; `gofmt -l` → empty. File was in a clean tree (no concurrent session owned it); the auto-commit daemon picks up the change.
- Nothing else was touched. The 47 file-size / jscpd / docs-gates findings are deliberately left to their owning ratchets and tracked postures (report, don't fix).

## What I deliberately did not touch

- **Baseline file** `.art-dupl-baseline.json` — untouched: `check` = 0 new means no re-record is owed (re-recording without new rulings risks blessing identity drift).
- **The 47 file-size findings (D10)** — tracked posture, owned by its plan/TODO row; not this sweep's to ratchet.
- **jscpd / docs-gates standing findings** — pre-adjudicated classes (registry 2026-11-11; `.docs-gates.json` ratchets).
- **No commits/pushes** — daemon-owned per repo discipline; no `git checkout`/`reset`/`rm` run at any point.
- The buildflow full run was already in flight from another shell; it was left alone and its verdict read from `buildflow history`.

## Sweep-log line

Appended to `docs/dedup-registry.md` per the registry's ritual (0-new precedent, 2026-11-04 05:36 class): pinned v0.7.2 `-t 1` re-run, **0 new vs 108**, no extraction, no baseline write; one `unused`-lint fix in `internal/arch/model_api_test.go` (not a clone-ruling change).
