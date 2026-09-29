# Status Report — 2026-09-29 02:26 CEST

**Session scope:** erraudit `tree` investigation (triggered by "looks very little"), the pbx sentinel fix it surfaced, and the pre-existing broken `nix build` the quality gate exposed. Nothing else researched, per instruction.

**Session start state:** HEAD `0816f15`, tree clean. `0816f15` had landed `go-error-family@v0.11.0` in go.mod/go.sum via the auto-commit daemon — without the vendorHash roundtrip, so `nix build` was already broken before this session touched anything.

---

## a) FULLY DONE

| # | What | Evidence | Files |
|---|------|----------|-------|
| 1 | **"Very little" tree mystery solved** — 7 package-level sentinel declarations exist (crm×3, pbx×2, store×2) but only 4 unique names (`ErrDisabled`/`ErrUnauthorized`/`ErrNotFound` each in 2 packages, `ErrListFull` once). `erraudit tree` dedupes same-named sentinels to one row and draws edges only from package-level declarations; all wrapping is inline (89 sites) by design, so max depth 0 is the true shape, not a tooling gap. | Declaration grep (7 sites), both `tree` runs, count arithmetic; documented in AGENTS.md erraudit bullet | AGENTS.md:385-392 (commit `b03327f`) |
| 2 | **pbx sentinels fixed** — `ErrDisabled`/`ErrUnauthorized` were built with `fmt.Errorf` with no directives (revealed by the `[fmt.errorf]` tag in the `--type-aware` tree run); converted to `errors.New` matching crm/store. Zero behavior change (same var, same message), identity preserved for `errors.Is`. | `go build ./...` green; `go test -count=1 ./internal/pbx/...` ok (2.020s); `erraudit tree` now all-`[sentinel]`; tier-1 erraudit (`--type-aware --disable-extensions`) exit 0 | internal/pbx/client.go:13,43,48 (commit `b03327f`) |
| 3 | **Stale vendorHash repaired** — `nix build .#webphone` was failing (`go-error-family@v0.11.0 ... no such file or directory` from the go-modules FOD). Followed the sanctioned manual deviation (.buildflow.yml:74-78): zero-hash placeholder → read `got:` → applied `sha256-eGRrw4YeGgNFJ552iz/4CZThZ30jcuvVH98HLQgoIKA=`. | `nix build .#webphone` green (drv `webphone-2.7.0`); committed `46ad1d3` | flake.nix:143 |
| 4 | **Full buildflow gate green** — after the hash repair: exit 0, 54 steps success / 0 failed / 0 skipped (+8 via config). | `/tmp/bf2.log` line "v2 results: 54 success, 0 failed" | — |
| 5 | **Tier-1 erraudit gate re-verified** post-fix: `erraudit ./... --type-aware --disable-extensions` exits 0. | `/tmp/erraudit_t1.txt`, `tier1 exit=0` | — |

End state: working tree **clean** at 02:26; the daemon swept all session edits (flake.nix → `46ad1d3`, pbx+AGENTS → `b03327f`).

## b) PARTIALLY DONE

1. **%w-count reconciliation — right conclusion, sloppy arithmetic.** I told you "91 raw `%w`, 14 in tests" — actual: 12 in tests, and my grep covered `internal/` only while erraudit scans `./...` (`cmd/webphone/main.go` carries ~9 wrap sites), so "89 vs 91" was never apples-to-apples. The conclusion (display dedupe + flat-by-design) is unaffected, but the numbers I quoted were partly wrong. **Remaining:** one precise recount (include `cmd/`, count `fmt.Errorf`-with-`%w` in non-test function bodies). Effort S. No blocker.
2. **AGENTS.md note added, but the file is over budget.** Now 475 lines vs buildflow's 377 max (98 over — pre-existing; my +6 lines made it marginally worse). A compaction pass is not started. Effort M/L.
3. **Remote end-state not verified.** AGENTS says verify daemon end states with `git ls-remote`; I confirmed the local tree is clean but did not confirm the commits are pushed. Effort S.

## c) NOT STARTED

Things this session surfaced but deliberately left alone:

1. **Tier-2 family-adoption migration** (113 stdlib_constructor findings outside the crm seam; top seams config.go 22, store/messages.go 20, pbx/client.go ~8). The `go-error-family@v0.11.0` dep landing suggests this project is beginning, but zero migration work happened here. Waiting on owner intent (question 1).
2. **Monthly tier-1/tier-2 erraudit re-measure** — scheduled for 2026-10-22 per AGENTS.md; not due, not done.
3. **docs-health HARVEST** of this report's section (f) into TODO_LIST.md/ROADMAP.md — per the skill handoff; waiting for your go (you said WAIT).
4. **BuildFlow binary staleness** — preflight warn: binary built at `e881e96`, BuildFlow repo HEAD `991080a`. Advisory; `nix build . && nix run .#reinstall` not run here.
5. **go.mod go-line flipflop** — preflight warn: the `go` line changed 8 times in 20 commits; fleet tooling dispositions fighting. Out of this repo's scope alone; not started.
6. **buildflow regression-baseline normalization** — the "slower than baseline" verdict came from a cold-cache first run (0% hit rate); advisory only, not re-baselined.

## d) TOTALLY FUCKED UP

**Nothing this session shipped is broken** — every gate is green. Radical honesty about the process, though:

1. **The repo's `nix build` was silently broken when the session started** (severity: blocks every full gate and release preflight; root cause: daemon dep-sweep landed go.mod changes without the vendorHash roundtrip — the *same* failure mode as the documented `0a7a732` repair). Fixed this session, but the loop that lets it recur is still open (see e6/f11).
2. **I violated the skill-loading order.** Ran `buildflow` before loading the buildflow SKILL.md; the skill's triage table had the exact answer (`nix build` hash mismatch → hash repair; "don't hand-fix vendorHash" + the project's documented exception). Cost: one wasted `-s nix-hash-fix --fix` attempt that died with "no executable nodes" because the step is `skip_steps`'d in this repo (.buildflow.yml:79-84).
3. **Invalid first placeholder hash** — used `sha256-PLACEHOLDER=`, which nix rejects as an invalid SRI hash before it can print `got:`; wasted one build cycle. The all-zeros/A placeholder is the correct incantation.
4. **Delivered wrong arithmetic in-chat** (14 vs 12 test `%w`; missing `cmd/` from the grep) — see b1. Small, but it was the session's *headline* reconciliation and I stated it with more confidence than the evidence supported.

## e) WHAT WE SHOULD IMPROVE

1. **Load the matching skill before the first tool call, not after the first failure.** The buildflow skill had the answer pre-packaged; the detour was pure waste.
2. **Baseline before change.** A 10-second `nix build .#webphone` (or `git log -p -- go.mod flake.nix`) before editing would have shown the pre-existing breakage and kept my change's verification clean instead of conflating the two. Worth an AGENTS.md hard-won rule.
3. **`erraudit tree` display is actively misleading** — it hides 3 of 7 declarations behind same-name dedupe, and the pick site flips between runs (crm rows without `--type-aware`; pbx/store rows with it). That ambiguity is what triggered this whole investigation. Qualified rows (`crm.ErrDisabled` + `pbx.ErrDisabled`) would fix it. Upstream erraudit/BuildFlow candidate.
4. **Recount before asserting reconciliations.** The 89/91/12/14 mess came from eyeballing a grep summary mid-answer.
5. **The vendorHash roundtrip has no forcing function.** It is documented in two places (AGENTS.md + .buildflow.yml comment) and has now broken the build twice (`0a7a732`, today) because the daemon sweeps deps and nothing runs the roundtrip. Either repair buildflow's nix-hash-fix for this layout (chronic 10/10 no-write) or add a daemon-side hook. The real gate (`nix flake check` sandbox build) only catches it after the fact.
6. **AGENTS.md needs a compaction policy**, not append-forever — it is 98 lines past the buildflow cap and every session adds more.

## f) NEXT TASKS (brainstorm — HARVEST fuel, not commitments)

Impact: Critical/High/Medium/Low · Effort: S <30min, M 30min-2h, L >2h

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Repair the vendorHash forcing function: fix buildflow `nix-hash-fix` for this flake layout (chronic no-write) OR add a daemon dep-sweep hook that runs the documented placeholder→`got:`→apply dance | High | L | Quality (upstream) |
| 2 | Plan the tier-2 family-adoption migration seam-by-seam now that `go-error-family` is a real dependency (config.go 22 → store/messages.go 20 → pbx/client.go) | High | L | Quality |
| 3 | Get owner intent on Q1 (below) before #2 — migrating to the wrong family shape would be months of compound interest | High | S | Decision |
| 4 | Slim AGENTS.md 475 → ≤377 lines: move war-story prose to docs/lessons.md, keep rules only | High | M | Cleanup |
| 5 | File upstream erraudit fix: qualified sentinel rows in `tree` (show all 7 declarations; stop same-name dedupe + nondeterministic pick site) | High | S | Quality (upstream) |
| 6 | Run docs-health HARVEST on this section into TODO_LIST.md/ROADMAP.md | High | S | Documentation |
| 7 | Verify daemon commits pushed: `git ls-remote` vs local HEAD (`46ad1d3`) | Medium | S | Quality |
| 8 | Add CHANGELOG entries: pbx sentinel `errors.New` fix + vendorHash repair | Medium | S | Documentation |
| 9 | Precise %w reconciliation (include `cmd/`, non-test function bodies) and record the number where the 89 is cited | Low | S | Documentation |
| 10 | Monthly erraudit tier-1+2 re-measure (due 2026-10-22): refresh 127/113 counts + seam list in AGENTS.md | Medium | S | Quality (scheduled) |
| 11 | Verify erraudit is actually one of buildflow's 54 gate steps (I ran tier-1 manually; the step list this run is unverified) | Medium | S | Quality |
| 12 | Rebuild/reinstall stale BuildFlow binary (`nix build . && nix run .#reinstall` in the BuildFlow repo) | Low | S | Quality |
| 13 | Fix the go.mod `go`-line flipflop fleet-side (align go-version-auto-configure and go-mod-update dispositions) | Medium | M | Quality (upstream) |
| 14 | Re-run buildflow once more so timing baselines settle and the regression verdict clears | Low | S | Quality |
| 15 | Add "baseline `nix build` before editing" to AGENTS.md hard-won rules (this session's lesson) | Medium | S | Documentation |
| 16 | Version-stamped smoke of the built binary: `--bin $(nix build .#webphone)` + `--expect-version v2.7.0` per AGENTS (the drv built this session is unverified at runtime) | Medium | S | Quality |
| 17 | Triage the 40 ignored errors from the audit metrics (`--format finding`): classify intentional vs real; the nolint'd ones (main.go:79 `db.Close`, pbx/client.go:205 body close, session/sqlite.go:103,111 deletes, actions.go:489,506 import skips) are documented but worth one pass | Medium | M | Quality |
| 18 | Consider `slog.Debug` on the best-effort session DELETE failures (session/sqlite.go:103,111) so hygiene failures leave a trace | Low | S | Quality |
| 19 | Fold the `generic_return` findings (~25 store/blob/messaging/crm funcs) into task #2 as one error-type family decision instead of piecemeal `FooError` structs | High | L | Quality |
| 20 | Repo-wide grep for other no-directive `fmt.Errorf` constructors (package-level is clean now; inline sites unchecked) | Low | S | Quality |
| 21 | Record the "erraudit tree dedupes same-named sentinels" gotcha in crush-config `references/lessons.md` if it generalizes fleet-wide | Low | S | Documentation |
| 22 | Add the 4-name/7-declaration sentinel inventory to docs/error-contract.md (it is the error-surface table's natural home) | Low | S | Documentation |
| 23 | Strengthen the release-runbook preflight with an explicit `nix build .#webphone` step (catches vendorHash drift before gates, not during) | Medium | S | Documentation |
| 24 | Tag daemon auto-commits with a session/marker for attribution (this session had to archaeology `0816f15` to find who landed the dep) | Medium | M | Process |
| 25 | Upstream: clearer buildflow error when `-s` selects a `skip_steps`'d tool ("no executable nodes" cost a debug cycle) | Low | S | Quality (upstream) |
| 26 | Schedule a tier-3 owner-only full audit run (`--enforce-samber-oops --enforce-generic-return`) when you want the full picture | Low | M | Quality |
| 27 | After #2 lands, re-check whether the tier-2 family-adoption tracking should replace the per-seam counts with per-package adoption percentages | Low | M | Quality |
| 28 | Consider archive cadence for docs/status/ (5 unarchived reports; docs-health ANNOTATE owns it) | Low | S | Cleanup |
| 29 | Pin the erraudit version used by gates in the devShell so binary-freshness warns stop appearing every run | Medium | S | Quality |
| 30 | Add a micro-check that fails when a package-level sentinel is built with anything but `errors.New` (protects fix #2's pattern) — or fold into erraudit upstream | Low | S | Quality |
| 31 | Re-verify `nix flake check` (incl. KVM backup VM) on this HEAD before the next release train — buildflow ran its nix steps, but the runbook gate is the fuller check | Medium | S | Quality |
| 32 | Decide whether the stack needs a re-pin for this train (Go-source + flake changes; no markup, so stack E2E re-run not required — but the lock-drift probe rule may still apply) | Medium | S | Decision |
| 33 | Once upstream erraudit dedup fix (#5) lands, update the AGENTS.md tree note to describe the new display | Low | S | Documentation |
| 34 | Fleet sweep: apply the `errors.New` sentinel fix pattern to sibling repos if they share the `fmt.Errorf`-sentinel habit | Low | M | Quality (fleet) |

Items 1-6 are the ones I'd start with; the rest are harvest/roadmap fuel.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Is the tier-2 family-adoption migration now actively starting?** `go-error-family@v0.11.0` appeared in go.mod via daemon auto-commits (HEAD was `0816f15` at session start) with no narrative commit, no AGENTS.md update, and no migration code. I checked git log, .buildflow.yml, and the AGENTS tier-2 paragraph — none state the plan. The answer decides whether I should start converting seams (#2/#19) and in what order, or treat the dep as swept-but-unused.
2. **Do you want the `erraudit tree` same-name dedupe reported upstream as a defect?** It hid 3 of 7 sentinels and its pick site flips between runs — but erraudit is your tool, and maybe the flat display is deliberate. I can't file (or decide not to file) against your own project without your call.
3. **May I restructure AGENTS.md (not just append) to get under buildflow's 377-line cap?** Every session appends; the file is 98 lines over and my note added more. A compaction moves prose to docs/lessons.md and rewrites sections — that touches the file every concurrent session depends on, so I want explicit permission + timing.

---

*Report is a point-in-time snapshot. Section (f) is the input for docs-health HARVEST into TODO_LIST.md/ROADMAP.md — say the word. Waiting for instructions.*
