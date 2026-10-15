# Status — M11 closed (S001 + 59-finding triage + baseline), #74 filed, backup VM green, release held for a quiet host

**When:** 2026-10-15 21:26 CEST · **Repo:** webphone @ `410aa4d` (pushed, main green) · **Session:** continuation of the SUPERB metaengine-train remainder under GET SHIT DONE (G3 = "Everything", M9 = "file upstream now")
**Predecessor report:** `docs/status/2026-10-15_20-26_metaengine-remainder-m1-m12-release-staged-status.md` (its §g three questions were answered this session by re-affirmation: Q1 suppress+file → done, Q2 status quo → no action, Q3 fire-when-quiet → pending).

---

## a) FULLY DONE

1. **M11 S001 false positive — suppressed, verified, filed.**
   - Source-verified the entire mechanism before touching anything: the suppression directive (`cmd/cqrs-lint/pkg/suppression/parser.go`, `directives.go` — own-line or first-non-blank-line-above, end-of-line form, block form, stale-suppression detection) and the S001 heuristic (`cmd/cqrs-lint/pkg/rules/security/rules.go:37-45` substring list incl. `"token"`, `:55` length floor, `:59` URL/placeholder allowlist, `:65-70` report at the literal's line; `catalog_security.go` — S001 `hardcoded-secrets`, critical/medium).
   - Applied `//cqrs-lint:ignore(S001)` + rationale at `internal/readmodels/enroll.go:24`.
   - Verified: `buildflow -s cqrs-lint` → 13 findings, all info/warning, **zero errors, exit 0**.
2. **THE FIND this session: branching-flow findings had grown 19 → 59 since the 2026-10-11 triage.** The baseline's 189 entries matched NOTHING (stale 3 runs — ID drift, the CLI-vs-SDK issue already filed upstream as branching-flow#13). Composition: 2 large-struct, 27 COMPOSITION_mixin, 9 DUPLICATE_TYPE, 8 DO_must-in-runtime, 2 IFACE_COMPLETE, 2 PHANTOM-PLUMBING, 10 PHANTOM-TRANSPOSE. The metaengine train (M8) added ~19 untriaged TypeDecl sites, and the statement-level convention ("fixed or nolinted with a reason") was silently violated for 22 sites.
3. **All 59 findings triaged at the site, per the repo convention:**
   - Nolint applicability verified at branching-flow source first: only panicanalyzer, phantom (all modes), roleak, doanalyzerv2, parser consult `core.IgnoreComments` — composition/dupe/ifacecomplete do NOT, so TypeDecl+IFACE findings cannot carry nolints and must be baselined with site comments.
   - **13 statement-level findings** now carry `//nolint:branching-flow` with honest per-site reasons (sipcred.Derive HMAC order; textagent New/Approve/DraftFor loud-failure arguments; app dedup Seen/CheckAndRecord metaengine-port shape; store.SetFlag single-site legacy; readmodels Mint parity port; faxes UpdateStatus/AdvanceIdempotent ProviderRef-ripple ruling; userauth.RemoveCredential handler-door inputs).
   - **app.go hoist (the one real code change):** the four `importLegacy*` calls each carried a duplicated `do.MustInvokeNamed` pair → hoisted to one shared `(sqlite, engine)` resolution. 8 DO findings → 2, both nolinted; genuine dedup, not lint appeasement.
   - **40 TypeDecl/IFACE findings** annotated with `(baselined)` site rulings across 15 files (rationales tied to existing doctrine: strong-id wire ruling, parity ports, closed event set, consumer-side seams).
4. **Baseline re-recorded:** `.buildflow-baseline.json` now holds exactly the 40 annotated findings; the staleness warning is GONE; 189 dead entries dropped (commit `e1fc301`: 954 insertions / 8,421 deletions).
5. **Upstream filing #74:** searched dups open+closed (empty), drafted per github-voice (banner, TL;DR, Problem/Impact/Proposal, check-draft 0 FAIL / 0 WARN), filed **go-cqrs-lite#74 — "cqrs-lint S001 flags schema NAMES as critical hardcoded secrets (substring heuristic)"** — verified landed. Draft kept at `docs/drafts/2026-10-15_go-cqrs-lite_s001-token-substring.md`.
6. **webphone-backup KVM VM test verdict COLLECTED — GREEN (exit 0).** The previous session launched it and never collected; the re-run executed the full driver + vm-test-run.
7. **Gates:** `go build ./...` OK; `nix fmt` clean (treefmt + gofmt); session-behavior suite green twice (`go test -count=1 ./internal/server/ -run TestServer` → ok, once post-hoist); whitespace-drift clean; targeted `buildflow -s branching-flow` → 40 findings, all baselined, exit 0.
8. **Push + CI state:** daemon pushed all session commits; main verified GREEN at `160e984` (run 38788192274). HEAD now `410aa4d` (fresh push — CI-watch still open, see f).
9. Commits this session: `b3df82c` (daemon sweep of the 20-file triage), `e1fc301` (baseline + full narrative), `410aa4d` (upstream draft).

## b) PARTIALLY DONE

1. **M11 close: ~95%.** Everything green on targeted runs (cqrs-lint, branching-flow, build/tests, fmt); the ONE missing piece is a **full `buildflow` run verdict** on this tree. Deliberate deferral: release.sh runs `BUILDFLOW_NO_RESULT_CACHE buildflow` as its first gate anyway — under a host at load 90–145, running the ~55-min full gate twice is pure waste. The release's buildflow leg IS the M11 verdict if you accept that; otherwise it is one command when the host quiets.
2. **M13 release 2.9.0: staged, not fired.** CHANGELOG folded (prior session), runbook + script read, all preconditions green — blocked ONLY on the host-load gate (`< 8`). Load during this session's wait: 20 → 109 → 145 → 31 → 72 → 90 (sibling sessions: NixOS toplevel rebuild + browser E2Es). A heartbeat watch loop is running (expires ~21:49).
3. **Owner questions from the predecessor report:** Q1 and Q2 resolved by re-affirmation; Q3 (release autonomy) is being executed as "fire when quiet" — but see g) below, the load volatility makes the strict threshold worth revisiting.
4. **§f harvest into TODO_LIST:** the prior report's 50-item list lives in that report; only the train row + standing watches landed in TODO_LIST. Full harvest pending (docs-health HARVEST).

## c) NOT STARTED

1. `scripts/release.sh 2.9.0` execution (gates → tag v2.9.0 → push → lychee → stack relock + browser E2E + VM test → aarch64 → gh release).
2. CI-watch of the release run + `git ls-remote` end-state verification.
3. Stack notification ("webphone 2.9.0 train done; relock window open").
4. TODO_LIST train row G3 GREEN line; addendum status report with the release verdict.
5. Owner terminal items: review boxes on #71/#72/#74; prod deploy; prod smoke (`--expect-version 2.9.0`); prod backup drill; pbx-artmann re-pin.
6. CI-watch for the just-pushed `e1fc301`/`410aa4d` (run list still showed `160e984` as newest completed at report time).

## d) TOTALLY FUCKED UP

Nothing destructive — but three honest failures, two of them inherited patterns:

1. **The 19 → 59 finding drift sat undetected for ~a day.** The M8 train committed ~19 new TypeDecl sites and 22 un-nolinted statement sites without a full buildflow run — violating the AGENTS "run buildflow at TRAIN START" rule (the 2026-10-08 incident produced exactly this rule). The detector's warnings-only severity cap made the drift INVISIBLE in gate results: the run was "1 error finding" while the triage debt quietly doubled. Only insisting on a fresh `-s branching-flow` inventory surfaced it. (Fix candidate: f-31 below.)
2. **Lost the narrative-commit race to the daemon AGAIN.** The 20-file triage sweep rode a `chore: auto-commit` (`b3df82c`); my `auth:` narrative only covers the baseline file. Same as M8. Root cause: I batch the commit until after verification (gates, fmt) — a multi-minute window the daemon always wins. (Fix: f-41.)
3. **Re-ran the backup VM test instead of collecting it.** The predecessor launched it in a shell that died with the session; the verdict was never written down, so this session re-executed the full VM run under load. Cost: ~minutes of 32-core contention. Handoff hygiene: terminal verdicts must land in the handoff text, not in a dead shell.

Minor: tripped the edit-guard once on `app.go` (mtime drift, content identical — sed-reads don't refresh the read-state; the concurrent-sessions rule wants View before Edit and I gambled on bash reads). The guard caught it; no harm.

## e) WHAT WE SHOULD IMPROVE

1. **Gate verdicts must carry COUNTS, not just pass/fail.** "buildflow green" hid a 3× finding-growth. Every train report should state per-tool finding counts (like the T16 census and the art-dupl group count already do).
2. **Train-start buildflow is load-bearing when detectors are warnings-capped** — the severity cap turns "untriaged growth" into silence. Either re-measure at train start (rule exists — enforce it) or elevate specific rules via `error_rules` when the installed binary supports it.
3. **Baseline staleness is a signal, not noise:** "none of 189 matched, stale 3 runs" was printed by every run since the drift and nobody read it. Add it to the train checklist.
4. **Handoffs must carry terminal verdicts** (VM test results, run exit codes), never "launched in background".
5. **Commit-per-edit-batch beats commit-after-verification** when a daemon commits-and-pushes continuously: narrative first, gates after is ALREADY the repo rule — I verified then committed. Follow the rule as written.
6. **check-draft before gh, always** — worked (caught my dropped-"I" typo pre-file via re-read; the mechanical run then passed 0/0). Keep the two-step (human re-read + checker) as the standing pair.
7. **Status-report §f should be harvested into TODO_LIST within the same session** — the skill says the loop isn't closed until HARVEST runs; both this and the predecessor report left it open.

## f) NEXT — up to 50 things to get done

**Release-critical (the active train):**

| # | Task                                                                                                                              | Impact                 |
| - | --------------------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| 1 | Fire `scripts/release.sh 2.9.0` when host load < 8 (gates → tag → push → lychee → stack relock + E2E + VM → aarch64 → gh release) | BLOCKS the train close |
| 2 | Its buildflow leg doubles as the M11 full-run verdict — record the count                                                          | Closes M11             |
| 3 | CI-watch the release + verify `git ls-remote` end state                                                                           | Release hygiene        |
| 4 | CI-watch `e1fc301`/`410aa4d` (pushed 21:2X, runs may still be queued)                                                             | Main-green proof       |
| 5 | Post the stack notification ("2.9.0 train done; relock window open")                                                              | G3 tail                |
| 6 | TODO_LIST: G3 GREEN line + HARVEST this §f list                                                                                   | Loop closure           |
| 7 | Addendum status report with release verdict                                                                                       | Owner record           |
| 8 | Kill/expire the running load-watch loop cleanly (self-expires ~21:49)                                                             | Clean state            |

**Upstream reviews + watches (owner terminal unless noted):**

| #  | Task                                                                                                                                       | Impact            |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------ | ----------------- |
| 9  | Review boxes on go-cqrs-lite #71 / #72 / #74 (tick + timestamp after reading)                                                              | Provenance ledger |
| 10 | Standing watch: adopt #74's fix when released → DROP the S001 suppression directive                                                        | Debt removal      |
| 11 | Standing watch: adopt #71 truncation signal in webphone scans                                                                              | Silent-100 fix    |
| 12 | Standing watch: adopt #72 cap primitive → replace converging-loop trims (sweep + enroll)                                                   | Debt removal      |
| 13 | File upstream: composition/dupe/ifacecomplete don't consult IgnoreComments (same class as branching-flow#12) — TypeDecl nolints would work | FP class fix      |
| 14 | File upstream: IFACE_COMPLETE misses the nil-seam idiom (server.DedupWindow)                                                               | Rule accuracy     |
| 15 | Consider upstream note: DO_must-in-runtime can't see the registration side (composition-root doctrine)                                     | Rule accuracy     |
| 16 | When cqrs-lint ships the S001 fix: bump + drop directive (pairs with #10)                                                                  | Debt removal      |

**Docs + rules upkeep:**

| #  | Task                                                                                                                                                    | Impact         |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| 17 | docs/lessons.md: add the "warnings-capped detectors hide count drift" war story (19→59, ~1 day)                                                         | Memory         |
| 18 | docs/lessons.md: add the handoff-must-carry-terminal-verdicts lesson (VM re-run)                                                                        | Memory         |
| 19 | AGENTS.md: update the stale "buildflow = 1 known FP" line to green + the finding-count rule                                                             | Memory         |
| 20 | dedup-registry: append this session's sweep-log line (59-finding triage)                                                                                | Registry rule  |
| 21 | Decide: CHANGELOG line for the app.go hoist? (internal refactor, no behavior change — the bar says code-path changes get a line; verify classification) | Bar compliance |
| 22 | Decide (#15 prior): reconcile the 10-14 plan's over-counted census inline or keep the honest note                                                       | Record hygiene |
| 23 | Record (#30 prior): ROADMAP verified-no-changes note                                                                                                    | Record hygiene |
| 24 | `.buildflow.yml` ownership confirm (#26 prior): the 2026-10-15 un-skips say "owner decision" — see g) Q3                                                | Provenance     |

**Process hardening:**

| #  | Task                                                                                                     | Impact                     |
| -- | -------------------------------------------------------------------------------------------------------- | -------------------------- |
| 25 | Add "per-tool finding counts" to the train-start buildflow checklist (e.1)                               | Drift detection            |
| 26 | Add "baseline staleness line must be read" to the same checklist (e.3)                                   | Drift detection            |
| 27 | scripts/art-dupl-ci.sh wrapper pinning the CI version (#19 prior)                                        | Skew guard                 |
| 28 | Decide (#18 prior): art-dupl pin direction (see g) Q2)                                                   | One re-record vs recurring |
| 29 | Add the version-skew war story to docs/lessons.md (#20 prior)                                            | Memory                     |
| 30 | erraudit tier-1/2 monthly re-measure — due 2026-11-05 (#17 prior)                                        | Standing gate              |
| 31 | Quarterly art-dupl baseline audit — next ~2027-01 (#16 prior)                                            | Standing gate              |
| 32 | Consider `rule_accepts` for whole-rule acceptance (mixin fleet-wide) vs per-finding baseline — see g) Q3 | Baseline ergonomics        |
| 33 | Commit-per-edit-batch discipline vs the daemon (e.5) — try it next train                                 | Narrative survival         |

**Deferred product/store work (unchanged, carried):**

| #  | Task                                                                                                      | Impact              |
| -- | --------------------------------------------------------------------------------------------------------- | ------------------- |
| 34 | Legacy cleanup train: delete `legacy_*.go` + flags when the import window closes (#44/#45 prior)          | Census −6 sites     |
| 35 | `TestEngineCollectionsMatchPlan` arch-test so new read models can't skip Plan registration (#46)          | Hygiene             |
| 36 | Boot-log line when legacy `wp_enroll_tokens` rows exist but are ignored (#34 prior)                       | Operator visibility |
| 37 | threadsdual harness: re-run after every metaengine minor bump (#23 prior)                                 | Parity evidence     |
| 38 | Consider CI-ifying the threadsdual harness as a scheduled job (#24 prior)                                 | Evidence automation |
| 39 | `pgengine` behind `readmodels.OpenWithEngine` ONLY on real need (#35 prior)                               | Recorded path       |
| 40 | Backup script comment: why `sqlite3 .backup` needs no `-wal` handling (#36 prior)                         | Doc debt            |
| 41 | Prod smoke after deploy: `--base https://pbx.artmann.tech --expect-version 2.9.0` (#50 prior)             | Owner terminal      |
| 42 | Prod backup drill against deployed 2.9.0                                                                  | Owner terminal      |
| 43 | pbx-artmann re-pin after the stack relock (#38 prior)                                                     | Owner terminal      |
| 44 | Re-verify `.buildflow.yml` after daemon activity near it (standing rule #27 prior)                        | Standing rule       |
| 45 | Audit today's daemon commits for unintended sweeps (3–4 spot checks; `b3df82c` verified mine) (#25 prior) | Trust               |
| 46 | lychee: whitelist the drafts' issue-link pattern if it trips (#33 prior)                                  | Release tail        |
| 47 | 429-combined-load test (#5) + ASR follow-ups — parked from older sweeps (#40 prior)                       | Backlog             |
| 48 | Watch the next Crush upgrade for the QMD `get` render fix (#39 prior)                                     | Standing watch      |
| 49 | docs-health ANNOTATE this report when superseded by the addendum                                          | Doc hygiene         |
| 50 | After the train closes: re-read the plan's G-gates and formally mark G3 done (#37 prior)                  | Train close         |

## g) Questions for the owner (max 3)

1. **Release firing policy under this load:** load is oscillating 31 ↔ 145 with sibling sessions. Strictly hold for < 8 (release.sh's own gate), or authorize a bounded threshold (e.g. < 16) and/or a deadline ("if not quiet by HH:MM, hold until tomorrow")? The run eats ~30+ min of heavy compute; firing into a spike wastes it, holding forever stalls G3.
2. **art-dupl pin direction (carried from the predecessor report):** keep CI pinned at v0.7.2 and force locals to match (current fix, zero workflow edits), or bump the workflow to v0.9.0 once + one canonical baseline re-record so the newer detector is the standard? Workflow change = your call.
3. **`.buildflow.yml` posture (carried, #26 prior):** the 2026-10-15 un-skips (cqrs-lint, nix-hash-fix, gomod-check) are annotated "owner decision" — confirm that was actually you, and do you want whole-rule `rule_accepts` (e.g. accept COMPOSITION_mixin fleet-wide) instead of 40 per-finding baseline entries?

---

_Evidence: `buildflow -s cqrs-lint` exit 0 / 0 errors; `-s branching-flow` 40 findings all baselined, staleness gone; session suite ok (0.896s, post-hoist); webphone-backup exit 0; CI 38788192274 green; #74 landed. Commits `b3df82c`/`e1fc301`/`410aa4d` pushed._
