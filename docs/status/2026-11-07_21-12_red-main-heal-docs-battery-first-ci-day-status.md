# Status — the red-main heal: the docs battery's first CI day, fixed end to end (2026-11-07 21:12)

**Snapshot-rot header: every claim below was verified as of 2026-11-08 00:24
CET** (the fix-train claims were verified 2026-11-07 21:12–21:5x CET; the
report-verification pass — gates, planted control — completed 00:24). At
HEAD `605d2289` (pushed; CI run 42408436799 green — all seven jobs). This
report covers ONE session's run: the `fix!` mandate that found main red on
the docs-health battery's first CI day and healed it through four distinct
defects, ending fully green.

## a) FULLY DONE

| # | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Fix 1 — dprint pinned into the devShell.** The battery's CI step fail-closes on a missing `dprint` (the wrapper's exit-2 gate), and the devShell only ever found dprint through the host PATH — true on this workstation (`~/.nix-profile/bin/dprint`), false on a GitHub runner. `pkgs.dprint` (nixpkgs 0.64.3, host 0.64.1, same minor; formatting comes from the wasm plugins pinned in `dprint.json`, so no drift) now rides `nix/devshell.nix` beside `config.packages.docs-gates`. Proved under CI conditions locally: a stripped-PATH `nix develop -c` resolves both tools from the store and the battery runs green. Commit `6e3fb06a` (daemon-swept edit) + `0d4f0eb9` (the CHANGELOG half).                                                                                                                                                                                                                                                                                   |
| 2 | **Fix 2 — full history for the battery's job.** With dprint fixed, the next run failed DEEPER: the annotate gate counted **373 findings against the 205 ratchet** in CI vs 205 locally — same 343 files, same 6 148 struck items, same binary. Reproduced exactly with a local `git clone --depth 1`: the annotate counting resolves verdict references against GIT HISTORY, and a depth-1 checkout leaves most unresolvable. `fetch-depth: 0` on the `fast-checks` job's checkout (the `changelog-guard` job already did this for the same class of reason). The successor run measured 205 in CI — exact match. Commit `b77b502e` (daemon-swept).                                                                                                                                                                                                                                                                                                                                       |
| 3 | **Fix 3 — audit_max re-derived 71 → 72.** The audit-all gate then counted 72 unharvested open items on the grown corpus: the metaengine-dogfooding report's row 2 (`Deps.Contacts` → readmodels, a train LANDED 2026-10-15 and documented in the readmodels seam notes) crossed the `generic_df=22` boundary as living lines passed ~4 360 — the tool itself printed "accidental overlap, not a harvest". This is the documented DF-brittleness class (one CHANGELOG paragraph moved the count 330 → 331 in the v8 sweep), so the pin followed the measurement per the ratchet protocol, with the mechanism in the commit. Commit `5ac98889`.                                                                                                                                                                                                                                                                                                                                             |
| 4 | **Fix 4 — the go-test flake attributed, hardened, ledgered.** Run 42407471289's `go test -race` red-ed `TestDedupWindowSatisfiesIdempotencyContract/Record/NoopOnExistingKey`. Attributed FLAKE, not regression: the same job was green on the FOUR prior runs of identical hooks code, the docs-only successor run 42407606027 was fully green, and 3× local `-race` runs passed. Mechanism: the contract's NoopOnExistingKey holds only while its two `Record` calls land within the scaled shortTTL (60 ms at TimingScale 3); a `-race` scheduling stall outlasting that gap makes the second `Record` legitimately re-claim the expired window — CORRECT `DedupCheckAndRecord` semantics (`expires_at <= now` guard in claimkit) — which the wall-sleep assertion misreads as "extended the TTL". Hardened `TimingScale 3 → 5` (stall budget 60 → 100 ms) with the mechanism at the test; ci-flake-ledger row appended naming the mechanism and the upstream path. Commit `605d2289`. |
| 5 | **Main verified green end to end.** Final run 42408436799: all seven jobs success (fast flake checks incl. the battery, go test -race, nix build, smoke, advisories). Local legs: full `go test -count=1 ./...` green, the hardened contract test green 2× under `-race`, `scripts/format-check.sh` (the treefmt authority) green, and `scripts/docs-gates.sh --check` green in a stripped-PATH CI simulation at every step.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 6 | **Memory updated where the next session reads it.** AGENTS.md: the docs-gates bullet now carries the HISTORY-GATE GOTCHA (annotate resolves git history; shallow checkout inflates 205 → 373; every docs-gates checkout must be full-depth) and current pin values. CHANGELOG `[Unreleased]` § Fixed carries the double CI failure with both root causes. The todo candidate line for the upstream ask was PRE-FLIGHTED (`docs-gates freshness . --pre-flight` → SAFE) before writing.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| 7 | **The stalled push leg was worked around per the ratified bar.** Seven daemon commits sat unpushed > 60 min (oldest `4b41b13e` 20:02Z-era) while the daemon's push leg was dead; the manual-push bar (main red + verified fix in hand; then the 60-min time anchor) authorized both pushes, each followed by explicit CI verification via `gh run view --json conclusion` (not the watch exit code — see d) #1).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 8 | **The planted-nonsense control re-ran after the pin move and HELD.** AGENTS.md mandates the control (plant one §f row with zero-matching tokens, expect one finding) whenever the pin moves; audit_max moved 71 → 72 in this session. Planted `Qwxyz blorb vxyzzy zzzqqq flurgenheim` inside this report's §f table: audit-all replayed 261 reports → **73 findings = REGRESSION past 72** — the instrument still catches a planted row at the moved threshold. Two side-findings from the control itself: a first plant whose tokens included "planted/nonsense/control" self-HARVESTED off AGENTS.md's own control documentation (a token-match lesson for future control authors), and a row appended outside the §f table was invisible to the section-scoped parser (the scoping works). Row removed; 72/72 green re-verified after removal.                                                                                                                                         |

## b) PARTIALLY DONE

1. **The upstream go-idempotency ask.** The TODO_LIST row is written (Fleet-libs closeout
   residue, item 3), the evidence is assembled (ledger row, mechanism, the claimkit
   `expires_at <= now` guard reading), but the FILING has not started — it owes the
   verify-before-filing gate and the github-voice ritual first, and those were out of this
   session's `fix!` scope. TimingScale 5 is a band-aid, not the fix.
2. **Plumbing the battery into `.#ci`.** The battery's tools (dprint + docs-gates) are pinned
   in the DEFAULT devShell only, which is what the CI step enters by design; the minimal `.#ci`
   shell still lacks dprint. Fine today (one consumer, one shell), but the contract lives only
   in the commit message and the devshell comment — a second battery consumer would trip it.

## c) NOT STARTED

1. **The gh `run watch --exit-status` defect.** Observed once this session: the watch returned
   exit 0 while the run it was watching concluded `failure` (run 42407471289 — I read the
   matrix from `gh run view --json` and found the contradiction). Not reproduced, not verified
   against other gh versions, not filed. Until then: never trust the watch exit alone; check
   `conclusion` explicitly (now the session habit).
2. **The BuildFlow stale-toolchain findings** the parallel `--fix` run surfaced (go 1.27.0 tool
   binaries vs the catalog v4.13 floor ≥ 1.27.2 red-ing go-generate/govalid-generate/license-check
   locally while CI stays green; nix-build-verify 20/23 failures; vulnix 0/5 retries) — that is
   BuildFlow-repo work (`repo:` tag class), untouched here.
3. **The `annotate_max` drain.** The corpus sits exactly AT the 205 pin (zero headroom, same as
   audit_max was). Draining it is a deliberate docs-health ANNOTATE session, not a fix-train tail.

## d) TOTALLY FUCKED UP

1. **I misread `gh run watch --exit-status` exit 0 as a green run.** The watch printed job
   checkmarks from a mid-flight poll, exited 0, and I nearly reported success — the run had
   actually concluded `failure` (the annotate 373-vs-205 red). Caught it because the close-out
   check queried `conclusion` explicitly. Lesson now written into the process: the watch exit
   code is not evidence; `gh run view --json conclusion` is.
2. **The daemon scrambled my commit narrative.** My `git add devshell.nix CHANGELOG && commit`
   landed the fix MESSAGE on the CHANGELOG-only half (`0d4f0eb9`), because the daemon had swept
   the devshell.nix edit into `6e3fb06a` (a `chore:` commit) one minute earlier. Content was
   never at risk, but history reads scrambled: the "fix:" commit carries docs, the `chore:`
   carries the code. I judged a rebase too risky against an adversarial daemon (the 2026-10-27
   detached-HEAD incidents) and pushed as-is — the right call under the concurrent-sessions
   doctrine, but the narrative cost was real.
3. **I almost trusted a locally-green battery as CI-green evidence.** The commit message for
   the battery said "present in the devShell CI enters" while the devShell only worked via
   host-PATH inheritance — locally every check passed for an environment-dependent reason.
   The stripped-PATH CI simulation (what actually caught it) should have been run BEFORE the
   feature's own commit, not discovered by a red runner two hours later.
4. **The first stash-based attribution attempt was void.** `git stash push` found nothing to
   stash (the daemon had already committed my edits) and the measurement I took was of the
   WRONG tree state — I reported the 72 count as "without my edits" in my head before noticing
   the stash was empty. The `--no-local --depth 1` clone reproduction was the honest experiment;
   the stash attempt was noise I nearly treated as signal.

## e) WHAT WE SHOULD IMPROVE

1. **CI-parity simulations for every "CI enters the devShell" claim.** The stripped-PATH
   `nix develop -c` dance is five minutes of work and would have caught Fix 1 before the first
   push. A tiny wrapper (`scripts/ci-shell-sim.sh`?) or a documented one-liner in AGENTS.md
   would make it the default reflex instead of a rediscovery.
2. **Feature commits that change CI wiring owe a same-session watched CI run.** The battery
   train pushed at 18:37 and the fix train started ~20:20 — nearly two hours of red that a
   "watch the first CI run of any ci.yml/devshell change" rule would have truncated.
3. **The ratchet-at-pin smell.** audit_max moved to a value the corpus EXACTLY hits (72/72), and
   annotate_max was already sitting exactly at 205. A pin with zero headroom reds on the next
   innocent doc paragraph. The sweep protocol should either leave one slot of margin or expect
   the next sweep to be triggered by the red. Worth a line in the docs-health SKILL.md upstream
   ask bundle.
4. **Daemon-vs-narrative commits.** When a fix spans a code file + a docs file, stage and commit
   in ONE call is documented — but the daemon raced even that (it committed between my edit and
   my add). The only robust shape left is: edit → immediately commit (same tool call), which I
   did for the second push but not the first.
5. **Verification vocabulary.** "green" must name its instrument: watch exit, run conclusion,
   job list, or local suite. This session produced one false green (d1) and one vacuous local
   green (d3) — both would have been caught by naming the instrument in the claim.

## f) Up to 50 things we should get done next

Ranked by risk of silent failure, then by what unblocks the most.

| # | Task                                                                                                                                                                        | Impact | Effort | Home                         |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ---------------------------- |
| 1 | File the go-idempotency injectable-clock ask upstream (verify-before-filing + github-voice; evidence: run 42407471289, ledger row, claimkit guard reading)                  | High   | M      | TODO §Fleet-libs closeout n3 |
| 2 | Drain the annotate residue (corpus sits AT the 205 pin, zero headroom — the next innocent doc paragraph reds) so audit_max/annotate_max can move DOWN again                 | High   | L      | TODO §2026-11-06 r11         |
| 3 | Archive-residue wave 2: cadence headroom is ONE slot after this report (64/65); the next ten oldest live reports through the proven gates-before-move recipe                | High   | L      | TODO §2026-11-06 r11         |
| 4 | Append today's push-stall recurrence (7 commits > 60 min, manual bar invoked, push leg silent again) to the forensics section and re-run the daemon-audit leg               | Medium | S      | TODO §Daemon push-stall      |
| 5 | Reproduce the `gh run watch --exit-status` zero-exit-on-failure observation (run 42407471289); file upstream if it reproduces                                               | Medium | S      | TODO §2026-11-06 r21         |
| 6 | Triage the stale-toolchain buildflow findings (go 1.27.0 tool binaries vs catalog v4.13 floor ≥ 1.27.2 red-ing go-generate/govalid/license-check locally) (repo: BuildFlow) | Medium | S      | TODO §2026-11-06 r7          |
| 7 | Re-run the planted-nonsense control after ANY future pin move (this session's 71 → 72 move re-ran it as part of this report's verification pass)                            | Medium | S      | TODO §2026-11-06 r11         |
| 8 | Decide the ratchet-headroom posture: leave one slot of margin at sweep time vs accept at-pin reds as the sweep trigger (docs-health SKILL.md upstream note)                 | Low    | S      | TODO §2026-11-06 r21         |
| 9 | Document the battery's one-shell contract (default devshell is the ONLY shell carrying dprint + docs-gates; `.#ci` deliberately minimal) in AGENTS commands                 | Low    | S      | TODO §2026-11-06 r10         |

## g) Questions for the owner

1. **The go-idempotency filing (f1):** file it NOW as its own ritual, or fold it into the next
   fleet-libs sweep's upstream-ask bundle (r21's batch)? The evidence is fresh either way; the
   difference is whether the upstream clock fix can land before webphone's TimingScale band-aid
   is stretched again by a slower runner.
2. **Ratchet posture (f2/f8):** after the wave-2 drain, do you want pins re-landed with ONE slot
   of headroom (my recommendation — at-pin values red on innocent prose), or keep the honest
   at-pin snapshot and accept a red as the sweep trigger?
3. **The daemon push leg stalled twice today.** The forensics section names the fleet fix
   (the pma unit's GIT_CONFIG override) as pending and owner-gated — it would re-bind commit-msg
   hooks fleet-wide. Fix it now, or keep the advisory posture and the manual-push bar as the
   compensation?
