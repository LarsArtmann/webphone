# Status Report — release 2.9.0 held 7.5h on the load gate; watcher hardened to detached; gate mechanics corrected

- **Written:** 2026-10-16 05:05 CEST
- **Session span:** resumed ~21:26 (2026-10-15) after the predecessor report's WAITING state → now.
- **Scope discipline:** this report covers ONLY this session's run (load-watch operation, daemon-commit audits, one manual push, watcher hardening, release precondition verification) plus what it noticed. No unrelated research.
- **Format note:** the status-report skill is HTML-canonical; the operator's prompt explicitly demands `.md` at this path, so `.md` wins (flagged per skill contract, not propagated into the skill).

---

## Executive summary

The session executed the M11 handoff's waiting contract: fire release 2.9.0 when the host quiets (the report §g Q1 default: strict load < 8, no override given). The quiet window **never opened in 7.5 hours** of continuous watching (load oscillated 5.7 → 1175; a sibling rebuild held > 1000 for 33 minutes). The release did not fire; **nothing is broken** — all preconditions stay green and the watcher is armed. The session's real work became: auditing and pushing 7 daemon commits (including a sibling session's completed BD003 train), verifying release preconditions, and — after catching my own watcher **dead silently** for ~2 hours — hardening it through four iterations into a detached, log-file-based process that survives session death. One material self-correction: I had characterized `release.sh` as refusing to start under load; source verification at 03:55 showed its load gate has a **single call site late in the script** (before the stack E2E/VM phase, after tag+push). The strict-hold policy is the owner's deliberate extra safety, not a script constraint — this sharpens §g Q1.

| Metric                 | Value                                                                         |
| ---------------------- | ----------------------------------------------------------------------------- |
| Load-watch coverage    | ~7.5 h (21:35 → 05:07), heartbeats every 3 min                                |
| Daemon commits audited | 7 (all benign; 2 carried defect-class go.sum churn)                           |
| Manual pushes          | 1 (5 commits, push-bar anchored, CI green)                                    |
| Release fired          | NO — strict < 8 gate never met; max quiet streak 2/3 polls                    |
| Closest load to gate   | 5.66 (00:18, single poll); dips 10.08 / 12.42 not sustained                   |
| Worst burst            | 1175 (01:57–02:30, > 1000 for 33 min)                                         |
| Watcher iterations     | v1 in-session (died) → v2 +CI gate → v3 +clean-tree → v4 detached +sync-check |
| My commits             | 0 (all pushes were daemon-authored; none of mine pending)                     |

---

## a) FULLY DONE

1. **Predecessor §g interpreted, defaults applied.** The resumption prompt was read as the standing GET SHIT DONE re-affirmation; the three unanswered questions took their report defaults: strict load < 8 for firing, art-dupl stays CI-pinned v0.7.2, per-finding baseline stays. Evidence: predecessor report §g read at session start; every firing decision tonight cites the strict default.
2. **Resume-state audit found and cleared a commit backlog.** At 21:34 the repo held 2 unpushed daemon commits (`4831358` go.sum +26 + baseline trailing-newline; `8fbb8d6` go.sum −26 + the 148-line status report). Audited both at diff level: net effect = report file + a newline; the go.sum churn fully cancels. Evidence: `git show` of both; net `git diff origin/main..HEAD --stat` showed only 2 files.
3. **Both open CI watches from the predecessor closed GREEN.** `410aa4d` → run 38788192274's successor 38793829175 success; `e1fc301` covered by the earlier verified-green `160e984` run. Evidence: `gh run list` at 21:34.
4. **Three gap daemon commits audited before pushing** (`6c00695`, `e52e04e`, `c029741`, landed 21:40–21:44 while the session idled): a sibling session's `brandid-lint` BD003 suppression triage — `//brandid-lint:ignore(BD003)` directives with per-site rationales layered onto the M11 triage sites, 36 files. Verified the tree coherent BEFORE pushing: `go build ./...` rc 0; session suite ok 0.771s; readmodels ok 0.808s; domain ok 0.003s. Evidence: this session's test runs at 00:01–00:03.
5. **Manual-push bar executed and documented.** Daemon push-leg stalled 2h39m (anchor: oldest unpushed `4831358` @ 21:26:48+02, bar = 60 min). Pushed `410aa4d..c029741` at ~00:05; `git ls-remote` verified main = `c029741`; CI on `c029741` green (run 38814348023, cross-confirmed by the sibling's own report and my watcher's CI polling). Unpushed = 0 since.
6. **Release preconditions all verified green** (the part of the release that does not depend on load): CHANGELOG folded (`## [2.9.0] - 2026-10-15` present, compare link points v2.9.0...HEAD); v2.9.0 tag free locally and on origin; tree clean; HEAD synced with origin.
7. **`.buildflow.yml` integrity re-verified after daemon activity** (predecessor §f item 44, standing rule triggered by tonight's commits): all three 2026-10-15 UNSKIPPED annotations (cqrs-lint, nix-hash-fix, gomod-check) intact; the 18:02 daemon sweep of that file was the predecessor session's own un-skip edit. Evidence: `git log -3 -- .buildflow.yml` + grep of the annotations.
8. **Sibling train attributed and read, not fought.** The sibling session's BD003 train report (`docs/status/2026-10-16_00-31_brandid-bd003-brand-adoption-train.md`): 35 findings → 0 via suppression directives, zero runtime behavior change except one tautology removal (`userauth.Service.Shutdown`), full 24-package suite green (server suite 550s — which explains the 40-minute load plateau 01:09–01:51). Its later commits (`0970b38` go.sum +26, `9ff6730` go.sum −26 + table-reformat of MY status report + its train report) audited: formatter churn only, §g of my report untouched.
9. **release.sh gate mechanics verified at source** (03:55, after 6 h of quoting them from memory): `load_gate` defined at line 47, called exactly ONCE at line 224 — after tag+push, immediately before the stack relock + browser E2E + VM-test phase. There is NO load check at script start. Consequence documented in §d/e and §g Q1.
10. **Watcher hardened to a detached, session-independent process.** After discovering the in-session watcher dead (see §d 1): wrote `/tmp/webphone-release-watch.sh` (outside the repo — a repo file would trip the clean-tree precondition), `setsid`-detached as pid 2008959, logging every 3-min heartbeat + all release output to `/tmp/webphone-release-watch.log`. Conditions to fire: 3 consecutive polls with load < 8 AND latest main CI run green AND webphone tree clean AND synced with origin. Syntax-checked (`bash -n`), double-fire prevented (in-session watcher killed BEFORE the detached one started), liveness + first heartbeat verified. Any future session reads the log file to catch up — the "handoff must carry terminal verdicts" lesson, applied to a live process.

## b) PARTIALLY DONE

1. **Release 2.9.0 — armed, not fired.** What works: every precondition green (§a 6), watcher armed and healthy (heartbeat 05:07:08, load 40.45, dirty=0, unpushed=0, CI green). What remains: the sustained < 8 window (max streak tonight: 2/3 polls; the floor touched 5.66 once at 00:18 but the next poll rebounded to 8.55). Blocker: fleet load from sibling sessions (a Rust visual E2E `cargo test --test visual_ds` runs single-threaded for long stretches; a NixOS rebuild pushed load > 1000). Effort to finish: zero if the owner overrides the threshold (§g Q1); otherwise unbounded wait.
2. **Watcher correctness — iterated 4 times instead of designed once.** v1 (in-session bash loop) died with the session idle; v2 added the CI gate; v3 added the clean-tree check; v4 (detached) added the unpushed-sync check and log-file state. All versions correct for what they checked; the arcs are documented in §e 1/3.
3. **Daemon-commit audit depth.** All 7 commits audited for DANGER (none found), but the recurring go.sum ±26 churn (twice tonight) was never content-identified — the 30-second `git show` that would name the 26 rows is still undone (§f 8). The churn is net-zero and the tree is correct, so this is an open question, not an open wound.

## c) NOT STARTED

1. **Everything downstream of the tag** — blocked on the release itself, all pre-staged in predecessor §f and re-listed in §f below: stack notification, TODO_LIST G3 GREEN line + HARVEST of the §f list, addendum status report, prod deploy tail (deploy → `--expect-version 2.9.0` smoke → prod backup drill → pbx-artmann re-pin — owner terminal).
2. **Review boxes on go-cqrs-lite #71 / #72 / #74** — owner terminal, unchanged.
3. **docs-health ANNOTATE of the predecessor report** — its §f items 44/45 were executed tonight (buildflow.yml re-verify; daemon-commit audit ×7), but the ANNOTATE pass that would mark them done in the old report has not run (this report records them; HARVEST will reconcile).
4. **§f items 17–20 of the predecessor** (lessons.md war stories, AGENTS.md green-line update, dedup-registry sweep-log line) — not verified whether the sibling session's report closed any; untouched by this session.

## d) TOTALLY FUCKED UP

1. **My watcher died silently for ~2 hours and I did not notice — the owner's prompt did.** The in-session background shell (1F3) was reaped when the session idled (~21:55 → 00:00). I found it at 00:00 only because a "status?" message made me look. Severity: process risk, not data risk — no sub-8 window demonstrably missed (lowest load in the gap ~30; the all-night floor 5.66 occurred later, under watch). Root cause: Crush reaps background shells on idle; I rebuilt a pattern that had ALREADY failed once — the predecessor session's handoff literally says "Load-watch loop killed." A known failure mode was re-created instead of engineered away. Mitigation now live: detached watcher + log-file state (§a 10).
2. **I mischaracterized release.sh's load gate to the owner for ~6 hours.** At 21:35 and again at 00:00 I reported "release.sh's own gate refuses to fire at load ≥ 8" — wrong. Source-verified 03:55: `load_gate` is called exactly once (line 224), after tag+push, fencing only the stack E2E/VM phase; the script's start has NO load check. Severity: decision-quality, not correctness — the strict-hold policy stood either way (it is the owner's §g Q1 default), but the tradeoff framing ("firing into a spike wastes the run") was overstated: the script can run its gates under load, and a mid-run spike yields a partial-but-resumable state (tag pushed, stack phase pending), not a wasted run. Root cause: I quoted the gate from the handoff instead of reading the call-sites — the exact discipline the M11 S001 work (verify directive mechanics at source) had just taught. Mitigation: corrected in the 00:00 answer's follow-up, §g Q1 re-sharpened below.
3. **Defect-class daemon commits keep landing on a hook the daemon demonstrably bypasses — and I audited for danger without asking why the churn exists.** The `.githooks/commit-msg` guard (AGENTS.md, 2026-10-15 note) is neutralized fleet-wide by the pma unit's `GIT_CONFIG_*` exports; tonight produced two fresh go.sum ±26 daemon commits in that unguarded lane (`0970b38`/`9ff6730`, repeating the `4831358`/`8fbb8d6` pair from 21:26). I verified each pair nets to zero and moved on — twice — without identifying WHAT the 26 rows are or WHO dirties go.sum on a cadence. An unidentified recurring write into a dependency file, on a lane with no guard, is precisely the shape of the vendorHash incidents the guard was built for. Severity: latent. Mitigation: §f 8 (identify the rows), §g Q2 (re-bind the hook fleet-wide or accept advisory).

## e) WHAT WE SHOULD IMPROVE

1. **Detach every long-running watch on day one.** The pattern now proven: write the loop to a script OUTSIDE the repo, `setsid` it, log heartbeats to a file, and treat "log stale > 10 min" as a downed watch. Impact: this session lost ~2 h of coverage to a reapable in-session loop; the predecessor lost its watcher the same way. Fix: make it the default first move (candidate: a lessons.md entry + a skill note).
2. **Verify gate/linter MECHANICS AT SOURCE before characterizing them upstream.** Same root cause as the M11 S001 find (verify the suppression parser before suppressing), violated again 6 h later on release.sh. Fix: any claim of the form "tool X refuses/gates Y" gets a same-session source check — it is a 2-minute grep and it changes decisions.
3. **Design watch conditions upfront, not live.** v1→v4 grew four gates (quiet-streak, CI, clean tree, origin-sync) one discovery at a time; each missing gate was a real fire-into-a-bad-state risk. Fix: the detached script (v4) is now the canonical artifact — future sessions copy it instead of evolving a new one.
4. **Cross-session load coordination is a fleet gap.** The sibling's 550s server suite and my release both need "the machine is busy" signals from each other; tonight they only met as load numbers. Fix idea (fleet-level, owner's call): a tiny well-known file convention (`/tmp/fleet-heavy-run` with pid + ETA) that heavy runners write and watchers respect — not built this session, just named.
5. **Audit churn for CONTENT, not just for danger.** "The diff is benign" and "I know what this diff is" are different verdicts; tonight's report needed the second and settled for the first. Fix: a churn-audit habit — identify the changed rows before declaring a daemon commit clean (30 s), especially on guard-bypassed lanes.

## f) NEXT — up to 50 things to get done

**Release-critical (the active train):**

| #  | Task                                                                                                                                                                                   | Impact               | Effort | Category |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- | ------ | -------- |
| 1  | Answer §g Q1 (fire policy under load) — every other release item keys off it                                                                                                           | BLOCKS the train     | S      | Decision |
| 2  | If override authorized: fire `WEBPHONE_RELEASE_MAX_LOAD=N nix develop -c scripts/release.sh 2.9.0` and watch the stack E2E for the known flake (one re-run)                            | BLOCKS the train     | M      | Release  |
| 3  | If strict holds: detached watcher (pid 2008959, `/tmp/webphone-release-watch.log`) auto-fires on sustained quiet — verify liveness at next session start (log stale > 10 min = downed) | Safety net           | S      | Release  |
| 4  | After firing: CI-watch the release run, verify `git ls-remote` shows v2.9.0, watch watcher log for the per-step verdicts (buildflow count = the M11 full-run verdict)                  | Release hygiene      | M      | Release  |
| 5  | Post the stack notification ("2.9.0 train done; relock window open") after the stack relock leg lands                                                                                  | G3 tail              | S      | Comms    |
| 6  | TODO_LIST: G3 GREEN line + HARVEST this §f list (docs-health HARVEST)                                                                                                                  | Loop closure         | M      | Docs     |
| 7  | Addendum status report with the release verdict                                                                                                                                        | Owner record         | S      | Docs     |
| 8  | Identify the go.sum ±26 rows (one `git show 0970b38 -- go.sum`): which checksums, which tool writes them, which session owns the writer                                                | Kill the churn class | S      | Bug      |
| 9  | Prod deploy tail after release: deploy → prod smoke `--expect-version 2.9.0` → prod backup drill → pbx-artmann re-pin                                                                  | Owner terminal       | M      | Release  |
| 10 | Decide whether the sibling BD003 train (landed post-handoff: `c029741`..`9ff6730`) belongs in 2.9.0's notes — its CHANGELOG delta, if any, rides the same tag                          | Record accuracy      | S      | Docs     |

**Upstream reviews + watches (owner terminal unless noted):**

| #  | Task                                                                                                                    | Impact                    | Effort | Category |
| -- | ----------------------------------------------------------------------------------------------------------------------- | ------------------------- | ------ | -------- |
| 11 | Review boxes on go-cqrs-lite #71 / #72 / #74 (tick + timestamp)                                                         | Provenance ledger         | S      | Upstream |
| 12 | Standing watch: adopt #74's S001 fix when released → drop the suppression directive                                     | Debt removal              | S      | Upstream |
| 13 | Standing watch: adopt #71 truncation signal in scans                                                                    | Silent-100 fix            | S      | Upstream |
| 14 | Standing watch: adopt #72 cap primitive → replace converging-loop trims                                                 | Debt removal              | S      | Upstream |
| 15 | File upstream: branching-flow composition/dupe/ifacecomplete don't consult IgnoreComments (TypeDecl nolints would work) | FP class fix              | M      | Upstream |
| 16 | File upstream: IFACE_COMPLETE misses the nil-seam idiom (server.DedupWindow)                                            | Rule accuracy             | S      | Upstream |
| 17 | Consider upstream note: DO_must-in-runtime can't see the registration side                                              | Rule accuracy             | S      | Upstream |
| 18 | When cqrs-lint ships the S001 fix: bump + drop directive (pairs with #12)                                               | Debt removal              | S      | Upstream |
| 19 | art-dupl pin direction (§g Q3): resolve once — CI v0.7.2 vs workflow v0.9.0 re-record                                   | Kills a standing question | S      | Decision |

**Process + tooling (from §e):**

| #  | Task                                                                                                                       | Impact          | Effort | Category |
| -- | -------------------------------------------------------------------------------------------------------------------------- | --------------- | ------ | -------- |
| 20 | lessons.md: "detach long watches; log-file state; stale log = downed" war story (this session's §d 1)                      | Memory          | S      | Docs     |
| 21 | lessons.md: "verify gate mechanics at source before characterizing upstream" (§d 2)                                        | Memory          | S      | Docs     |
| 22 | Canonicalize the detached-watcher script (copy into scripts/ with docs) so future trains stop re-growing it                | Reuse           | S      | Quality  |
| 23 | Fleet heavy-run convention (`/tmp/fleet-heavy-run` pid+ETA file) — owner-level, cross-session courtesy                     | Fleet hygiene   | M      | Feature  |
| 24 | §g Q2: re-bind commit-msg hook fleet-wide (nix-config pma module) or formally accept advisory-only                         | Guard integrity | M      | Decision |
| 25 | Check whether the sibling session closed predecessor §f 17–20 (lessons/AGENTS/dedup-registry entries) before re-doing them | Dedup of effort | S      | Docs     |

**Docs + records hygiene (carried, still open):**

| #  | Task                                                                                                          | Impact         | Effort | Category |
| -- | ------------------------------------------------------------------------------------------------------------- | -------------- | ------ | -------- |
| 26 | AGENTS.md: update the stale "buildflow = 1 known FP" line to green + the finding-count rule (#19 prior)       | Memory         | S      | Docs     |
| 27 | dedup-registry: append the 59-finding triage sweep-log line (#20 prior) — verify not done by sibling          | Registry rule  | S      | Docs     |
| 28 | Reconcile the 10-14 plan's over-counted census or keep the honest note (#22 prior)                            | Record hygiene | S      | Docs     |
| 29 | ROADMAP verified-no-changes note (#23 prior)                                                                  | Record hygiene | S      | Docs     |
| 30 | docs-health ANNOTATE the 21-26 predecessor report with tonight's executed items (44/45 done, release pending) | Doc hygiene    | S      | Docs     |
| 31 | Decide the CHANGELOG classification for the app.go hoist (#21 prior: internal refactor — line or not)         | Bar compliance | S      | Docs     |
| 32 | lychee: whitelist the drafts' issue-link pattern if it trips the release leg (#46 prior)                      | Release tail   | S      | Release  |
| 33 | After train close: re-read the plan's G-gates, formally mark G3 done (#50 prior)                              | Train close    | S      | Docs     |

**Standing watches (re-affirmed tonight, unchanged):**

| #  | Task                                                                                                      | Impact          | Effort | Category |
| -- | --------------------------------------------------------------------------------------------------------- | --------------- | ------ | -------- |
| 34 | Re-verify `.buildflow.yml` after ANY daemon activity near it (standing; re-verified tonight §a 7)         | Integrity       | S      | Standing |
| 35 | Audit future daemon sweeps for unintended content (7 clean tonight; the audit habit is the control)       | Trust           | S      | Standing |
| 36 | Watch for the Crush background-shell reaping behavior biting other long tasks (this session's §d 1 class) | Process         | S      | Standing |
| 37 | Watch the next Crush upgrade for the QMD `get` render fix (#48 prior)                                     | Standing        | S      | Standing |
| 38 | threadsdual harness re-run after every metaengine minor bump (#37 prior)                                  | Parity evidence | S      | Standing |
| 39 | Monthly tier 1+2 re-measure (next 2026-11-05; tier-2 must stay 0)                                         | Error contract  | M      | Standing |
| 40 | Quarterly art-dupl baseline/suppressed-set audit (#ratified baseline rule)                                | Hygiene         | M      | Standing |

**Backlog / recorded paths (unchanged):**

| #  | Task                                                                                                                                                   | Impact              | Effort | Category |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------- | ------ | -------- |
| 41 | 429-combined-load test (#5 prior)                                                                                                                      | Backlog             | M      | Feature  |
| 42 | Legacy cleanup train: delete `legacy_*.go` + flags when the import window closes (#34 prior)                                                           | Census −6 sites     | M      | Cleanup  |
| 43 | `TestEngineCollectionsMatchPlan` arch-test (#35 prior)                                                                                                 | Hygiene             | S      | Quality  |
| 44 | Boot-log line for ignored legacy `wp_enroll_tokens` rows (#36 prior)                                                                                   | Operator visibility | S      | Feature  |
| 45 | CI-ify the threadsdual harness as a scheduled job (#38 prior)                                                                                          | Evidence automation | M      | Quality  |
| 46 | `pgengine` behind `readmodels.OpenWithEngine` ONLY on real need (#39 prior)                                                                            | Recorded path       | L      | Recorded |
| 47 | Backup script comment: why `sqlite3 .backup` needs no `-wal` handling (#40 prior)                                                                      | Doc debt            | S      | Docs     |
| 48 | Consider CI-ifying the detached load-watch pattern for release trains (#22 generalized)                                                                | Evidence automation | M      | Quality  |
| 49 | `TestServedPageHoldsTheDomContract` + stack browser E2E re-run owed after ANY served-markup change (the release's stack leg covers it — verify it ran) | Contract proof      | S      | Quality  |
| 50 | After 2.9.0 ships: bump-train sweep (templ-components, go-health, metaengine) rides the NEXT train's `buildflow update`, never the daemon              | Dependency policy   | M      | Quality  |

_(Carried items renumbered from the predecessor's list where still open; predecessor numbering in parentheses where applicable.)_

## g) Questions for the owner (max 3)

1. **Release firing policy — now with corrected mechanics and 7.5 h of evidence.** Correction first: `release.sh` does NOT refuse to start under load; its single `load_gate` call (line 224) fences only the final stack E2E/VM phase, AFTER tag+push — so firing at moderate load cannot waste the gates; worst case is a partial-but-resumable release (tag pushed, stack leg pending the runbook). Given that: (a) keep strict hold for a fleet-quiet night, (b) authorize a bounded override (e.g. `WEBPHONE_RELEASE_MAX_LOAD=16`) and I fire immediately, or (c) set a deadline ("if not quiet by HH:MM, hold until tomorrow")? I could not answer this myself: the fleet's overnight load floor tonight was ~5.7 with bursts to 1175, and the choice between "ship now with E2E-retry risk" vs "hold for quiet" is yours.
2. **Re-bind the commit-msg hook fleet-wide?** Tonight added two more defect-class daemon commits (go.sum ±26 churn) on the lane where the pma systemd unit's `GIT_CONFIG_*` exports neutralize `core.hooksPath` — the guard is advisory against the daemon (AGENTS.md 2026-10-15 note). The fix lives in the nix-config pma module and re-binds hooks for the ENTIRE daemon fleet. That is your fleet policy call: enable enforcement fleet-wide, or keep advisory and accept the churn (I can identify the churn's writer first — §f 8)?
3. **art-dupl pin direction (carried third time).** Keep CI pinned at v0.7.2 and force locals to match (current state, zero workflow edits), or bump the workflow to v0.9.0 once with one canonical baseline re-record so the newer detector is the standard? The workflow change is the part I cannot decide.

---

_Evidence trail: watcher log `/tmp/webphone-release-watch.log` (heartbeats 21:35→05:07, all CI verdicts, per-poll load); pushes `410aa4d..c029741` (ls-remote verified, CI 38814348023 green) and daemon's own `0970b38`/`9ff6730` (CI green per watcher poll 01:15); build + session/readmodels/domain suites green at 00:01–00:03; release preconditions verified §a 6; detached watcher pid 2008959 alive at write time. My commits this session: 0 — everything pushed was daemon-authored, audited before landing on origin._
