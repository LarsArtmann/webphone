# Status — red-main repair, ginkgo bump, and one honest self-review — 2026-11-11 16:02 CET

**Snapshot-rot header:** claims verified as of **16:02 CET**, HEAD `571b89fd` local /
`e1cb0831` pushed origin tip (the daemon and two concurrent sessions moved the tree
continuously through this session — every sha below was read back at the time named).

**Scope line — verified here:** the red-main diagnosis from live CI logs (`gh run view
43029324445 --log-failed`: race `[setup failed]` on proxy.golang.org stream errors +
docs-gates format rc=20 with the dprint diff read from the log); the format repair
landed as `75ba79c5` (CI run 43031922522 green, 7m39s); the race transient healed on
retry (job green on `6d2306ed`); the ginkgo bump's full tail — `bump-preflight.sh`
verdict (additive-only, no removed symbols), `go mod tidy`, vendorHash re-pin via
`nix build .#webphone.goModules --rebuild` (`dzvcEzCA` → `VSNx9zdY`, FOD rebuild
green), fresh `go test -count=1 ./...` 36 pkgs ok, batch-tip CI run 43035831397 green
9m11s, `scripts/ci-verdict.sh e1cb0831` exit 0 at 15:26; PR #234 closed superseded
with comment. **NOT verified here:** anything on the live deployment (no creds);
the concurrent sessions' train internals (only their commits observed: `6d2306ed`,
`a36283ea`, `e1cb0831`, `571b89fd`); a full `buildflow` run and `nix flake check`
(not run — CI covered the shipped tips instead); the race suite locally post-bump
(ran `-count=1` without `-race`; race proof is CI's green race job on the bump tip);
the exact final count of `cqrs-lint` findings after the concurrent session's fixes
(their `6d2306ed` landed mid-session; I read their TODO_LIST verdict row, did not
re-run the detector).

## a) FULLY DONE

| Work                                                                                       | Evidence                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Red-main diagnosed to two independent causes after 5 consecutive red CI runs (13:10–13:36) | `gh run view 43029324445 --log-failed`: (1) race job setup failures on proxy.golang.org `stream error ... INTERNAL_ERROR` for `projectionadapter/v4@v4.6.2` + `identity-model/v4@v4.17.0`; (2) docs-gates format gate rc=20 on `TODO_LIST.md` + `docs/dedup-registry.md` non-canonical bytes |
| Format gate repaired and green                                                             | the fmt battery (`scripts/docs-gates.sh`) green locally; identical-byte repair landed `75ba79c5` (concurrent session converged on the same dprint output); CI 43031922522 success                                                                                                            |
| Race setup failure classified transient and verified healed                                | job `go test -race ./...` success on `6d2306ed`'s run 43031617167; every later tip's race job green through 16:00                                                                                                                                                                            |
| Main green again, workflow-verified                                                        | `scripts/ci-verdict.sh e1cb0831...` exit 0 at 15:26 (first attempt hit the tool's transient NO-RUNS — index lag; retry green)                                                                                                                                                                |
| Ginkgo v2.33.1 → v2.34.0 bump with the mandatory tail                                      | `bump-preflight.sh`: additive-only (`DeferOnFailure`), no removed symbols, 16 changed .go files; `go.mod`/`go.sum` bumped + tidied; vendorHash re-pinned `nix/packages.nix:144`; goModules FOD rebuilt green; full suite 36 pkgs ok                                                          |
| Bump pushed and CI-green on the batch tip                                                  | landed inside daemon commit `be84a883`; run 43035831397 (tip `e1cb0831`) all-jobs success incl. the race job and the vendorHash-sensitive `nix build .#webphone`                                                                                                                             |
| Dependabot PR #234 closed as superseded                                                    | `gh pr close 234 --comment` at 14:19 UTC: "superseded: landed on main in be84a883 with the vendorHash re-pin"; PR state CLOSED                                                                                                                                                               |
| Archive headroom for this report paid per today's precedent                                | `2026-10-26_20-10_error-agents-train-status-and-self-review.md` banner-adjudicated (live remainder named with fresh 2026-11-11 greps) and `git mv`'d to `docs/status/archived/`; live count 64 → 65 with this file                                                                           |

## b) PARTIALLY DONE

1. **Red-main incident persistence is incomplete.** Diagnosis and repair are done,
   but the `ci-flake-ledger` row for the proxy stream-error class is NOT written —
   and the class has now bitten TWICE (run 42308961304 on 2026-11-06, whose tracker
   #236 closed with its draft ledger row still `TBD`, and today's 5-red stretch).
   The watcher's recurrence grep said "no matches — apparently a new class" for
   42308961304; today was the second sighting, which is exactly when the ledger
   row becomes load-bearing. Effort: S.
2. **The `cqrs-lint` surface was not re-measured post-train.** The 12 findings in
   the train-start scan were all already adjudicated (keep-visible rulings in
   `.buildflow.yml` for C014/D006/E023/F009/F021; the concurrent session's
   `6d2306ed` fixed F031/F019 inside internal/cqrs), but the detector was not
   re-run to record the new count. Effort: S.
3. **The buildflow timing-regression advisory was noted, not triaged** (test-race
   90s→292s +225%, 20 more steps +50-127%). Exit 0, non-gating; plausibly host
   load from concurrent sessions, but nobody ruled that out with
   `buildflow timings --regressions`. Effort: M.
4. **The nextd queue was read, not maintained.** `webphone#234` and the closed
   `#236` were handled in GitHub but never marked `nextd done`; the queue still
   carries them as Needs review / Active. Effort: S.

## c) NOT STARTED (observed open, none of it touched today)

- **Voicemail playback prod legs** (TODO_LIST): stack redeploy riding main + the
  play-click verdict, host-side diagnostics, the three stack coverage arms.
  Owner/deployment-gated.
- **The T05 authenticated prod capture chain** (creds owner-gated): the capture
  itself, then gallery walkthrough → ugliness inventory → first fix → first
  compare chain → calibration → mobile pass.
- **Night-run stability pair** and the **ring-guard positive path** (need real
  prod + a caller).
- **BuildFlow bandit-provider issue filing** — the drafted issue sits in
  `docs/drafts/`, owner banner unchecked.
- **webphone#9** — switch on the Kith CRM integration (deployment + CRM access).
- **The 2.14.0 release dance** — the concurrent session folded the CHANGELOG
  (`a36283ea`); the runbook's clean-clone `release.sh`, tag, stack re-lock, and
  pbx-artmann relock are all still ahead.
- **Docs-health residue**: annotate_max 203 un-drained, ~1600 un-inventoried §a/§b
  rows in live reports, cadence at 65/65 cap with every October-tail report
  failing the strict archive gates.
- **Type-safety roadmap rows** (push.Transport defined type, sso.Finish value
  object, FaxJob.DocumentPath as RelPath, asr/matrix wiring structs,
  machineSurface bool-last) and the **per-CALL transcript floor** reopen trigger.
- **Datastar NO-GO residue**: the `/health` unsafe-eval → nonce-lane retirement.

## d) TOTALLY FUCKED UP (all mine, none shipped damage)

1. **A dependency bump rode a daemon commit** — the owner's 2026-10-11 rule
   (bumps never ride the auto-commit daemon, after two vendorHash red-main
   incidents) was violated in fact: `go.mod`/`go.sum`/`nix/packages.nix` landed
   inside `be84a883 chore: auto-commit 4 changed file(s)`. Root cause: I lost two
   commit races in one session — the daemon committed between my `git add` and
   `git commit` both times (the format repair converged byte-identically, so only
   the bump was a real doctrine break). The bytes and the mandatory tail were
   verified and CI is green, so the risk the rule guards (stale vendorHash
   reaching main) did not materialize — but the rule exists precisely because
   "verified this time" is how the 10-06/10-09 incidents started. Severity:
   process, not product. Workaround: none needed retroactively; prevention is
   the commit-race protocol in §e1.
2. **I ran `scripts/ci-verdict.sh` with a mangled sha mid-session** and reported
   GREEN off it before cross-checking. The 2026-11-09 misread-green class is
   exactly this; the final tip verdict was re-read cleanly (exit 0), but for one
   stretch the session's own posture claim rested on a typo'd instrument read.
   The tool naming "1 run(s) concluded success" for a sha that matches no run is
   its own small defect — worth a self-test case (see §f8).
3. **The archive-gate detour burned ~25 minutes on doctrine archaeology.** I ran
   `check-rows --strict`/`bare-rows` across eight October reports, wrote a scan
   whose match pattern was wrong (`incomplete` contains `complete`, so eight
   failing files printed as CLEAN), suspected a binary-path typo, re-scanned
   with exit codes — all before reading today's ACTUAL precedent, which showed
   the operative archive ritual is banner + move (gates scan live reports only).
   The right order was: check what today's successful archive did FIRST, then
   act. ~20 minutes of pure waste from sequencing.
4. **I launched the race suite before checking the documented host leak.** The
   agent shell exports `GOEXPERIMENT=jsonv2`; the first background run was
   tainted and had to be killed and relaunched with `env -u GOEXPERIMENT`. One
   `echo` before launching would have saved a ~5-minute background slot.
5. **Nothing shipped is broken** — no user-visible surface, no data path, no gate
   was damaged by this session's work; the failures above are process-level and
   each one is diagnosed with a prevention below.

## e) WHAT WE SHOULD IMPROVE

1. **Commit-race protocol for surgical repairs**: before committing on this
   tree, read `git log origin/main..HEAD --format='%h %cI'` — if the daemon is
   mid-cycle (commit younger than ~2 min), wait for its push leg instead of
   racing it; when racing is necessary, `add + commit + push` in ONE call and
   accept that a lost race means verifying the daemon's commit rather than
   re-committing. Impact: eliminates the §d1 class. Cost: one git call.
2. **Ledger the proxy-stream-error class NOW** (it is two sightings old): one
   row in `docs/ci-flake-ledger.md` naming the `[setup failed]` +
   `stream error ... INTERNAL_ERROR` signature and the re-run-and-go-green
   verdict, so the watcher's recurrence grep finds it next time. Impact: turns
   the third occurrence into a 30-second triage. Cost: S.
3. **`ci-verdict.sh` hardening**: the transient NO-RUNS (GitHub search-index
   lag) deserves either a built-in single retry or a self-test case pinning
   that a sha with zero matching runs can NEVER print "concluded success"
   (the §d2 loophole). Impact: protects the workflow-name-verified verdict
   contract. Cost: S.
4. **`GOEXPERIMENT` guard**: prepend `env -u GOEXPERIMENT` habitually to every
   `go test`/`go build` launched from the agent shell, not only when a
   local-vs-CI divergence is already suspected. Impact: kills a documented
   recurring trap. Cost: none.
5. **docs-health doctrine vs practice**: AGENTS.md prescribes strict
   `check-rows` + `bare-rows` BEFORE `git mv`, while today's two successful
   archives used the banner + move pattern (gates scan the live set only, so
   pre-move strictness is unenforceable noise). Either write the banner
   pattern into the doctrine or re-implement the strict ritual — one home per
   fact applies to rituals too. Impact: stops every future archiver from
   repeating my §d3 detour. Cost: S (doc edit).
6. **Archive-headroom economics**: the 65-cap plus 28-row adjudications per
   slot means archives starve unless drain trains are scheduled; the cadence
   gate now blocks EVERY new report on an archive. The v10-style drain should
   be a recurring calendar item, not an emergency. Impact: report-writing
   stops costing an archaeology dig. Cost: M per drain.

## f) Up to 50 things we should get done next

<!-- check-rows:ledger -->

| #  | Area        | Task                                                                                                                                                  | Why / verification                                     |
| -- | ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| 1  | ci-ledger   | Write the proxy-stream-error class row in docs/ci-flake-ledger.md (42308961304 + today's 5-red stretch, re-run-and-green verdict)                     | second sighting unledgered; watcher grep found nothing |
| 2  | queue       | `nextd done LarsArtmann/webphone#234` and `#236`                                                                                                      | handled in GitHub, queue still lists them              |
| 3  | ci-tool     | ci-verdict.sh: pin the zero-runs-never-green contract + index-lag retry in its self-test                                                              | the §d2 typo loophole                                  |
| 4  | gates       | Re-run `buildflow -s cqrs-lint` and record the post-train finding count                                                                               | 6d2306ed changed the surface; count unverified         |
| 5  | gates       | Triage the timing-regression advisory (`buildflow timings --regressions`)                                                                             | test-race +225% unexplained                            |
| 6  | release     | Run the 2.14.0 release dance from a clean clone (release.sh, tag, CI verdict)                                                                         | a36283ea staged the CHANGELOG fold                     |
| 7  | release     | After the cut: nix-international-telephony lock bump + pbx-artmann relock (tri-repo ritual)                                                           | consuming stack rides webphone main                    |
| 8  | release     | Preflight battery BEFORE release.sh attempt 1 (buildflow + race + lychee + docs battery)                                                              | the five-attempt v2.13.0 lesson                        |
| 9  | docs-health | Drain train: adjudicate the October-tail reports one slot per report                                                                                  | every candidate fails strict gates; cadence at cap     |
| 10 | docs-health | Re-measure and lower annotate_max below 203 via an ANNOTATE sweep                                                                                     | ratchets move down only                                |
| 11 | docs-health | Harvest the ~1600 un-inventoried §a/§b rows in live reports                                                                                           | measured residue from v8                               |
| 12 | prod        | Stack redeploy riding main + voicemail play-click verdict recorded                                                                                    | the toast text is the live trigger's name              |
| 13 | prod        | Host-side voicemail diagnostics (journalctl send_file grep)                                                                                           | needs deployment access                                |
| 14 | stack       | Browser E2E voicemail playback arm in the stack repo                                                                                                  | zero voicemail surface in browser-e2e.py               |
| 15 | stack       | verify-live.sh authenticated audio fetch + operator.nix 206 pin                                                                                       | the Safari Range class                                 |
| 16 | prod        | T05 authenticated capture: `nix run .#ui-track -- --base-url https://pbx.artmann.tech --track`                                                        | creds owner-gated, asked twice                         |
| 17 | prod        | Owner gallery walkthrough → ranked ugliness inventory (T05 act 2)                                                                                     | unblocks T06-T10                                       |
| 18 | ui          | First restyle fix with before/after proof (T06)                                                                                                       | the owner-eyes bar                                     |
| 19 | ui          | Second prod run → first real compare chain (T07)                                                                                                      | diff noise floor unknown                               |
| 20 | ui          | Night-run stability pair ≥8h apart for calibration values (T09/T10)                                                                                   | 0.5% default proven optimistic locally                 |
| 21 | ui          | Mobile pass (T08)                                                                                                                                     | first-class surface per 2026-11-11 ruling              |
| 22 | ui          | Ring-guard positive path during a real inbound call (T13 live half)                                                                                   | selector verified, never exercised live                |
| 23 | ui          | First real-display `--open` pass                                                                                                                      | landed, never exercised                                |
| 24 | crm         | webphone#9: switch on Kith CRM integration (pki trust + bearer provisioning)                                                                          | deployment + CRM access owner-gated                    |
| 25 | crm         | crm#5 per-consumer bearer tokens (repo: crm)                                                                                                          | webphone shares one identity with CLI today            |
| 26 | android     | webphone-android#7 dependabot (AGP 9.4.1→9.5.1, Fix CI) (repo: webphone-android)                                                                      | 0d old, blocked on CI red                              |
| 27 | fleet       | Sweep nextd P0 externals for the webphone family (439 items)                                                                                          | queue is the source of truth                           |
| 28 | fleet       | cqrs-htmx#42 actions-group bump, 5d stale (repo: cqrs-htmx)                                                                                           | Fix CI class                                           |
| 29 | fleet       | go-sqlitestore#5 + storbi#11 same-day dependabots (repos: fleet)                                                                                      | 0d old, cheap merges                                   |
| 30 | drafts      | File the BuildFlow bandit-provider issue (draft exists)                                                                                               | owner banner still unchecked                           |
| 31 | typesafety  | push.Transport defined type + ParseTransport                                                                                                          | open roadmap row                                       |
| 32 | typesafety  | sso.Finish value object into Complete (auth: convention)                                                                                              | open roadmap row                                       |
| 33 | typesafety  | FaxJob.DocumentPath as domain.RelPath                                                                                                                 | open roadmap row                                       |
| 34 | typesafety  | asr/matrix wiring structs + machineSurface bool-last                                                                                                  | open roadmap row                                       |
| 35 | cqrs        | Per-CALL transcript floor: reopen trigger armed on nested-cap addressing                                                                              | the only remaining WithCapPerGroup gap                 |
| 36 | admin       | D4 delegation ruling (write-shaped admin actions)                                                                                                     | REJECTED posture, TODO_LIST owns                       |
| 37 | csp         | `/health` unsafe-eval → nonce-lane retirement                                                                                                         | filed in TODO_LIST from datastar NO-GO                 |
| 38 | identity    | AuditLog bounded projection consumption decision                                                                                                      | refresh trigger armed in pin v4.19.0                   |
| 39 | metrics     | Review island latency beacons on first real prod data                                                                                                 | metric is wired, data unseen                           |
| 40 | smoke       | Full smoke matrix against the 2.14.0 binary post-cut                                                                                                  | scenario counts printed, never hardcoded               |
| 41 | ui-capture  | Refresh the local 26-shot visual baseline post-indigo-restyle                                                                                         | the new baseline needs a fresh ledger                  |
| 42 | watcher     | Verify the mid-incident watch-lane fix (fafdbb74) deduped today's 5 red runs correctly                                                                | the fix landed DURING the incident                     |
| 43 | lessons     | docs/lessons.md row for the daemon-swallow + commit-race class                                                                                        | two same-day losses, unwritten                         |
| 44 | agents      | Note ci-verdict's transient NO-RUNS index lag in AGENTS.md conventions                                                                                | one-retry rule undocumented                            |
| 45 | docs-health | Write the banner+move archive pattern into the doctrine (or restore strict pre-move gates)                                                            | §e5: one home per ritual                               |
| 46 | agenda      | Schedule recurring v10-style drain trains on the watch calendar                                                                                       | archive economics are starving                         |
| 47 | upstream    | go-cqrs-lite: file F031-through-variadic + F032-scope detector-limit reports IF they recur                                                            | conditional per verify-before-filing bar               |
| 48 | pma         | The pma GIT_CONFIG hook-override fix (repo: nix-config, fleet-wide)                                                                                   | guard hole documented since 2026-10-15                 |
| 49 | deps        | Vendored sip.js currency check (update.sh, pinned 0.21.2)                                                                                             | routine                                                |
| 50 | restyle     | Merge this report's residue with the 15-15 restyle report's §f (rows 3-7 overlap: full buildflow, flake check, owner walkthrough, icons, focus rings) | two reports, one backlog — dedupe at harvest           |

## g) Top 3 questions I can NOT answer myself

1. **Daemon-commit ownership policy**: is a dependency bump landing inside a
   daemon `chore:` commit acceptable when the full tail is verified and CI is
   green (today's case), or do you want agent sessions to treat any
   daemon-swallowed bump as a defect to re-land properly (rebase out, recommit,
   force-with-lease)? I tried the re-commit path twice and lost both races; the
   guard hook cannot help (env-level override, documented hole).
2. **CI transient policy**: the proxy stream-error setup failure has now red
   main twice (2026-11-06 and today, 5 consecutive reds). Want a ledger row +
   an auto-retry (re-run failed jobs once) workflow, or keep re-runs manual and
   the ledger as the only persistence? I verified re-run-goes-green both times.
3. **Release timing**: the 2.14.0 CHANGELOG fold is staged (`a36283ea`) and the
   restyle/front-door trains are on main — should the next session run the
   release dance immediately (clean-clone `release.sh` per the runbook), or
   hold until you have done the owner eyes-on walkthrough of the restyle that
   the 15-15 report says nobody has seen yet?
