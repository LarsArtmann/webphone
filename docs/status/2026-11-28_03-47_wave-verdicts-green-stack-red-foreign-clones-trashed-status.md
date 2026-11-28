# Status — wave verdicts confirmed green; stack red is foreign; clones trashed; owner questions open

**Snapshot:** 2026-11-28 03:47 CET · claims verified as of 03:47 · **webphone local = remote = `4d14e453`** (one daemon commit past the session's `70d91cb7`; CI on `70d91cb7` GREEN, `4d14e453`'s own run still PENDING at report time) · **android local = remote tip = `19d3405`** (`d53a808` GREEN, tip run GREEN) · **stack remote tip = `f40ffca`** (my three CI legs GREEN; the run red on the FOREIGN bridge suite; one unpushed local daemon commit `caf5a70`) · Mode: solo-serial; this half-session is a verdict-collection + hygiene turn — no code changed by me.

**Verified here** (command + artifact for each):

- webphone `70d91cb7` CI **GREEN** — `scripts/ci-verdict.sh 70d91cb7` via the watch job's concluded output ("GREEN: CI (1 run(s)) concluded success").
- webphone parity — `git rev-parse HEAD` = `4d14e453`, `git log origin/main..HEAD` = 0 unpushed, `git status --porcelain` clean.
- android `d53a808` **GREEN** — `gh run view 46088056429 --json headSha,conclusion` → `success`; tip `19d3405` **GREEN** — run `46095846970` (8m20s, `success`); `ls-remote` tip = `19d3405`.
- stack `f40ffca` **job-level read** — `gh run view 46088071210 --json jobs`: `aarch64 VM test (telephony-boot, TCG)` ✓, `nix flake check (eval, packages, VM test)` ✓, `marker history arms (full checkout)` ✓, `stdlib python suites (engine, bridge, agents, oidc)` ✗; failing test extracted from `--log-failed`: `test_fetch_media_custom_cap_bounds_the_read` (`tests/test_telnyx_bridge.py:1184`).
- Clone hygiene — `merge-base --is-ancestor d53a808 HEAD` and `f40ffca HEAD` both LANDED in the real repos FIRST, then `/var/tmp/android-t16` + `/var/tmp/stack-t16` trashed (`ls` confirms gone); watcher jobs 1AD/1AE killed once their verdicts were collected directly.

**NOT verified here:**

- `4d14e453`'s own CI (PENDING at report time; its deltas — `scripts/audit-red-watch.sh` 8 lines + `scripts/lsp.sh` 2 lines — are NOT this session's edits; concurrent-session/daemon bytes, left untouched per the fix-forward protocol).
- The foreign bridge red does NOT reproduce locally — the file is another session's in-flight surface; diagnosis stopped at the CI log by design.
- No smoke / browser E2E / buildflow / flake check ran in this half-session (no code changes were made; docs + verdicts only).
- The stack's unpushed `caf5a70` content (daemon commit, 01:42 UTC, ~2 h old and under the manual-push bar at collection time — not read).

---

## a) FULLY DONE

- **All three in-flight CI verdicts collected — the wave's last open surfaces are closed.** webphone `70d91cb7` GREEN; android `d53a808` GREEN plus its tip `19d3405` GREEN; stack `f40ffca` green on every leg I own. The 00-30 report's §f row 1 is answered.
- **The stack run red is fully attributed — foreign, moving, and not mine to patch.** The failed job is the `stdlib python suites` leg; the failing test on `f40ffca` is `test_fetch_media_custom_cap_bounds_the_read` (media-cap read bound, `tests/test_telnyx_bridge.py:1184`) in the `telnyx-webhooks.py` lineage. On the PREVIOUS run (`aaf3df1`) the same job failed on a DIFFERENT test (`voicemail_db` AttributeError) — the other session's bridge surface is in flux between runs. My three legs (`nix flake check` incl. the dashboard VM suite + the FEATURES 53→54 count, `marker history arms`, the aarch64 VM test) are proven green in CI.
- **Working-set cleanup landed:** both landed clones trashed AFTER content verification (`merge-base --is-ancestor` for `d53a808` and `f40ffca` against the real repos' HEAD), the two redundant poller jobs killed, webphone parity re-verified clean. The 00-30 §f row 33 is answered.
- **T16 is CLOSED cross-repo and every leg now carries its CI stamp** (carried from the 00-30 half, verdicts now attached): webphone wire-fixture row struck (`3dd89c70` + the TODO tick train); android `e8ab38c` (`decodeIslandConfig` split, `LoopbackLiveContractTest` under Robolectric with the `java.net` transport, `scripts/loopback-live-check.sh` — `loopback live-check: PASS` against a real boot, nosurf CSRF dance with the masked cookie + clear `<meta>` token, 201 mint); stack `aaf3df1` (dashboard-on VM suite, `styled_404_arm()` first in the browser-E2E main, plainSuites wiring) — the suite itself ✅ in CI both runs.
- **The whole wave inventory is CI-green end to end**: `69ffa88a` (docs-gates red fix) → `b25476a0` (TODO ticks) → `a618ac9f` (T20 announcements drafts A/B) → `80a2eb7f`+`ffc062b9` (T05 `audit-red-watch.sh` + watch.yml wiring) → `303f2e2c` (T11/T12/T13/T14 fleet polish, go-error-family 64→68) → `a0c89506` (T15 lessons) → `70d91cb7` (the 00-30 status report) — every run GREEN.
- **T09/T10 verified** (carried): fleet-tags sweep 6 repos / 30 CHANGELOGs — one real finding, the go-cqrs-lite ROOT `v4.14.1` untagged upstream; citation sweep 182/182 clean, self-test 9/9.
- **pma#355 filed** (the hourly persistence-alarm detection lane draft from `docs/drafts/2026-11-27_pma-red-window-detection-lane.md`) — upstream adoption pending (watch row, §f).
- **Fresh `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` rc=0** mid-train (carried; re-certification at current HEAD is §f row 10).

## b) PARTIALLY DONE

- **Stack CI on `f40ffca`**: my halves (dashboard suite, styled-404 arm, FEATURES count, marker arms) are CI-green; the RUN stays red on the foreign `telnyx-webhooks.py` bridge suite. Remaining: the owning session's fix (or a tracker issue — §g Q2). Blocker: not my surface; effort once unblocked: S (their fix) / S (the issue).
- **T13** (fleet-libs index polish): every mechanical half landed (`303f2e2c`); the vision-QA half (row 21 — a vision-capable model reading the regenerated `index-render.jpg` trace) is provider-gated. Remaining: the vision pass. Effort: M once a lane lives.
- **T21 (android picker)**: slice 1 landed (`097d58a` + `d53a808`, GREEN); slices 2–6 remain — slice 2 ring-payload contact resolution, slice 3 push re-register after a line switch, slice 4 digest-handshake sketch (incident-gated), slices 5–6 the SIP-UA device legs (operator-blocked, `docs/SIP-UA-SMOKE-CHECKLIST.md`). Effort: S–M per slice, owner/operator-gated.
- **T18/T26 (the report drains)**: scoped, NOT executed — at 00:30 the pre-drain said READY 5 / BLOCKED 51, cadence 56/66 (pin 66), oldest 17d vs the 45d pin; NOT re-measured this half-session (honest gap). No cadence pressure; the READY set includes the wave's own live queue. Effort: L (hours of per-report adjudication), fresh-session task.
- **The wave queue itself**: T17 → T19 → T25 remain, deliberately not started after the owner's THEN-WAIT directive. Priority: on your GO (§g Q3 context).

## c) NOT STARTED

- **T17** (stack window B slice 1: fs-directory `operatorWatchedUnits`, secretFile fixture, answer-count helper, sipcred vector, suite merge). Why: owner WAIT. Priority: high — first of the remaining queue.
- **T19** (stack residues slice 1: paperless module option + module-check, deploy.md secret PATH column, demo-call recipe). Why: owner WAIT. Priority: medium.
- **T25** (stack chat-network slice 1: signal-cli unit + QR bootstrap + module-check; matrix homeserver choice memo — TODO_LIST:1574 carries it). Why: owner WAIT. Priority: medium.
- **T22/T23/T24** — owner-gated (GO + stack 5th mode / GO + security review / capture creds), deliberately closed; the "whole list" directive does not lift owner gates. Priority: yours.
- **T02 gated register** — the Kith CRM live e2e #9 row and the authenticated prod capture row stay closed; registry + watch calendar own the triggers. Priority: yours.
- **The drain ritual proper** (T18/T26 above). Priority: low until ~2026-12-07.

## d) TOTALLY FUCKED UP

1. **I burned two calls on the KNOWN `gh run list --commit` trap this half-session.** Empty output, rc=0, on both android and stack — the lag trap was documented in tonight's own technical context, and I still reached for it first out of habit. Third call used the reliable shape (`gh run list --limit` + `gh run view <id> --json`). Cost: two dead calls; severity: cosmetic; the pattern (consult the trap list, don't trust memory) is §e row 1.
2. **My 00-30 §g Q2 names a test that is already stale**: it says "the voicemail_db python-suite red", but on `f40ffca` the same job fails on `test_fetch_media_custom_cap_bounds_the_read` — a DIFFERENT test in the same `telnyx-webhooks.py` lineage. If a tracker issue had been filed naming one test, it would describe a moving target. Root cause: I attributed by the first failure's name instead of by suite + lineage. Correction rule in §e row 2; §g Q2 here is re-asked with the corrected scope.
3. **The two watcher jobs sat alive and silent long after their answers arrived by another route** — job 1AD/1AE polled with no output while two direct `gh` reads settled everything; I killed them ~15 minutes later than useful. Cost: hygiene only.
4. **Carried from tonight's first half — all eight remain on record with evidence in the 00-30 report §d** (docs-gates red shipped on main ~1.5 h; the android dprint skip; the python string-surgery that mangled `SettingsActivityTest.kt`; the tail-pipe fake verdict; the real-repo direct edit that raced the daemon; the stale Gradle cache; the self-check grep written from memory instead of served bytes; the ci-verdict call in a repo that lacks the script). None recurred this half-session; the rules they produced are §e.

## e) WHAT WE SHOULD IMPROVE

- **Consult the CI-trap list before ANY verdict collection on these repos** — go straight to `gh run list --limit` + `gh run view <id> --json headSha,conclusion,jobs`; the `--commit` flag has now wasted calls in two sessions on the same night. Impact: small but recurring; fix: one AGENTS watchers-entry line naming the reliable shape.
- **Attribute foreign reds by SUITE + lineage, never by one test's name** — an in-flight session's failing test changes between runs (proven twice tonight); a tracker issue that names a test describes a snapshot. Impact: avoids filing stale bug reports; fix: the issue body quotes the job name + the file lineage and says "failing test rotates".
- **Kill watcher jobs the moment their answer arrives by another route** — `job_kill` is cheap, silent running jobs are noise that outlives their purpose. Impact: hygiene; fix: habit.
- **Verify every NEW landing commit's own CI, not just the pinned one** — HEAD moved `70d91cb7` → `4d14e453` (daemon) between my green verdict and this report; the report pins both and says PENDING where it is PENDING. Fix: the scope line always carries the CURRENT sha's own verdict state.
- **Carried: the six rules from the 00-30 §e stand unchanged** — docs-gates `--check` before any docs commit; fresh-read before scripted edits; exit codes, never pipeline tails; clone-only landing; self-checks from served bytes; watch runs to verdict before writing §a rows.

## f) Up to 50 things we should get done next

<!-- check-rows:ledger -->

1. **File the stack tracker issue for the foreign `voicemail_db`-class python-suite red** — body names the `stdlib python suites` job + the `telnyx-webhooks.py` lineage and notes the failing test rotates (`AttributeError` → `test_fetch_media_custom_cap_bounds_the_read` between runs); not my files (fix-forward protocol). Impact: High · Effort: S · Category: Bug. (Owner call — §g Q2.)
2. **T25 stack chat-network slice 1**: signal-cli systemd unit + QR bootstrap + module-check + the matrix homeserver choice memo (TODO_LIST:1574). Impact: High · Effort: M · Category: Feature.
3. **T17 stack window B slice 1**: fs-directory `operatorWatchedUnits`, secretFile fixture, answer-count helper, sipcred vector, suite merge. Impact: High · Effort: M · Category: Feature.
4. **T19 stack residues slice 1**: paperless module option + module-check, deploy.md secret PATH column, demo-call recipe. Impact: Medium · Effort: M · Category: Feature.
5. **T18/T26 drain**: the 51 BLOCKED reports' §a–§g adjudication ritual — drive with `scripts/status-predrain.sh`, tranche commits, gates-first archives (no cadence pressure until ~2026-12-07; fresh-session task). Impact: Medium · Effort: L · Category: Documentation.
6. **The go-cqrs-lite ROOT tag ruling** (fleet-tags finding): root CHANGELOG says `v4.14.1`, remote has no root tag — tag it upstream or retitle the heading (owner/release act). Impact: Low · Effort: S · Category: Cleanup. (§g Q3.)
7. **library-policy#80 watch**: on ship, UNSKIP `library-policy` in `.buildflow.yml` + delete the skip comment. Impact: Low · Effort: S · Category: Quality.
8. **pma#355 watch**: when the hourly persistence-alarm lane lands, verify it files on a planted red. Impact: Medium · Effort: S · Category: Quality.
9. **Go 1.27.4 toolchain roll** (nixpkgs-gated watch row, 2026-11-08 calendar) — GO-2026-6859 / MaxBytesReader crash class flagged in httputil [Unreleased]. Impact: High · Effort: M · Category: Quality.
10. **Re-certify the full `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` at current HEAD** (`4d14e453`) — the mid-train rc=0 predates the daemon's post-report commits. Impact: Medium · Effort: M · Category: Quality.
11. **T21 slice 2**: the ring-payload number → native contact resolution decision (app-side cache vs raw number). Impact: Medium · Effort: S · Category: Feature.
12. **T21 slice 3**: push re-register AFTER a line switch (make the switch → re-register → refresh identity sequence explicit; the picker currently relies on the shell resume re-wire). Impact: Medium · Effort: M · Category: Feature.
13. **T21 slice 4**: the runtime digest-handshake sketch (gated on the first version-skew incident — keep deferred unless the v2.15-vs-old-app skew appears). Impact: Low · Effort: S · Category: Feature.
14. **T22** (owner GO + stack 5th mode) — D2 veto-window webphone half. Impact: Medium · Effort: M · Category: Feature.
15. **T23** (owner GO + security review) — D3 agent liveness slice 1. Impact: Medium · Effort: M · Category: Feature.
16. **T24** (owner capture creds) — the authenticated prod capture + ugliness inventory. Impact: Medium · Effort: S · Category: Quality.
17. **T02 gated register** — the Kith CRM live e2e #9 row and the authenticated prod capture row stay closed; the registry and the watch calendar own the triggers. Impact: Low · Effort: S · Category: Documentation.
18. **Add `audit-red-watch.sh` to the AGENTS watch-calendar entry** so its monthly self-test has a due date. Impact: Low · Effort: S · Category: Documentation.
19. **Prove the tip-unchanged skip LIVE**: after two watch runs on one tip, read the second run's log for the SKIP line (the self-test pins the decision core; the live path is unproven). Impact: Medium · Effort: S · Category: Quality.
20. **Plant a rotten verify claim in a scratch clone** and prove audit-red-watch files its issue (the CREATE path is dry-run-proven only). Impact: Medium · Effort: S · Category: Quality.
21. **templ-components v1.27.1/v1.27.2** ride the next templ train (JSON-LD hardening N/A here; auto-fit-grid absent from the tw.css `@source` set — carried struck, loop closed). Impact: Low · Effort: S · Category: Cleanup.
22. **Android: `verify-gate.sh` cache-hygiene leg** — a `--rerun-tasks`-equivalent or cache-key note so the stale-ktlint-report class cannot bite CI the way it bit local. Impact: Medium · Effort: S · Category: Quality.
23. **Android: wire the picker into the push re-registration contract test** (slice 3's test half). Impact: Medium · Effort: M · Category: Feature.
24. **Stack: the NEXT full browser E2E run keeps `styled_404_arm()` green against prod-shaped data** (the VM run passed; the live-stack run is the release-runbook obligation). Impact: Medium · Effort: S · Category: Quality.
25. **The `audit_max` 74 sits exactly AT the pin** — the next living-docs growth reds audit-all again; the drain (row 5) is the way back down. Impact: Medium · Effort: L · Category: Documentation.
26. **The `sipDomain` compact-JSON grep lesson → docs/lessons.md tooling traps one-liner** (self-checks from served bytes). Impact: Low · Effort: S · Category: Documentation.
27. **Keep the android run ids in its TODO row** (`097d58a`/`d53a808`/`19d3405` all GREEN) for the next session's freshness check. Impact: Low · Effort: S · Category: Documentation.
28. **The wave hand-off**: this file + the 00-30 report (+ the 16-59 one) are the next session's handover pair. Impact: Low · Effort: S · Category: Documentation.
29. **Re-run `scripts/audit-verify.sh` at HEAD** (green mid-train; the polish train touched reports after it). Impact: Medium · Effort: S · Category: Quality.
30. **Sweep the stack for OTHER stale 53-count pins** the dashboard suite broke (AGENTS/README may name 53 too — the FEATURES 53→54 fix covered one). Impact: Low · Effort: S · Category: Documentation.
31. **Run the smoke's dashboard nonce check** if a NON-default dashboard title ever boots (the fence copy is title-sensitive). Impact: Low · Effort: S · Category: Quality.
32. **The origin-picker first screen (m21.4) vs the Settings-level picker** — a design call (§g Q1). Impact: Medium · Effort: M · Category: Feature.
33. **The KMP spike stays deferred** (quiet-repo gate) — leave it. Impact: Low · Effort: S · Category: Documentation.
34. **Re-check `nix flake check --no-build` on the stack at its tip** (`f40ffca` local + the unpushed `caf5a70`) — one command, confirms the eval gate post-FEATURES-edit. Impact: Low · Effort: S · Category: Quality.
35. **The lessons rule-F class**: future `docs/lessons.md` appends keep the code-span-not-at-wrap-point rule. Impact: Low · Effort: S · Category: Documentation.
36. **The `doc-counts` gate lesson → a stack lessons/AGENTS one-liner**: every new check owes its FEATURES row the same day. Impact: Low · Effort: S · Category: Documentation.
37. **Keep the `library-policy` skip comment dated** — verify it survived tonight's daemon sweeps (it did at last read; re-verify after `4d14e453` touched scripts). Impact: Low · Effort: S · Category: Documentation.
38. **`docs/review/libs/index.html` regen discipline**: the next audit re-score updates BOTH the report and the index row in one commit (tonight's pattern held; go-error-family 68 is the latest pair). Impact: Low · Effort: S · Category: Documentation.
39. **The vision-QA half of T13** unblocks the moment any vision lane lives — top of the provider-gated register. Impact: Medium · Effort: M · Category: Quality.
40. **The android TODO row 46 E2E ring proof** stays blocked on stack switches + phone sign-in (operator). Impact: Medium · Effort: M · Category: Feature.
41. **The 16-59 report's §g questions** (provider budget, T22/T23 GO, capture creds) — superseded by §g here but carried for the next session. Impact: Low · Effort: S · Category: Documentation.
42. **Fresh-session full-status-report ritual** at the end of the next wave (the 00-30 report is the template). Impact: Low · Effort: S · Category: Documentation.
43. **The android first-device-train row** stays operator-blocked (§g Q1 feeds it). Impact: Medium · Effort: M · Category: Feature.
44. **Read the stack's `caf5a70` daemon commit when it pushes** (read-only check; its deltas are not this session's). Impact: Low · Effort: S · Category: Quality.
45. **Sleep the wave here** — the remaining tasks are owner-gated, operator-blocked, or fresh-session rituals; the fleet is green on every verdict that has concluded. Impact: Low · Effort: S · Category: Documentation.

## g) Three questions I cannot answer myself

1. **The android origin-picker scope** (carried unchanged from the 00-30 §g Q1): the plan's m21.4 said "origin-picker first screen"; I built the handover pack's Settings-level "Active line" picker instead (the pack IS the design doc and its §2 proposal is explicit). Do you ALSO want a first-launch origin entry screen (the app currently boots straight into the WebView with the default origin), or is Settings-level sufficient for the P1 shell?
2. **The stack tracker issue** (updated from the 00-30 §g Q2 — the failing test changed between runs): file a tracker issue NOW naming the `stdlib python suites` job + the `telnyx-webhooks.py` lineage (with "failing test rotates" noted), or leave it for the owning session to notice (their next push-triggered run reds the same way)?
3. **The go-cqrs-lite ROOT module** (carried from the 00-30 §g Q3): the root CHANGELOG heading says `v4.14.1` but the remote has no root tag — the only UNTAGGED row in the sweep. Should the root get a real `v4.14.1` tag upstream (a release act in your repo), or is the heading stale and should be retitled/removed so the sweep stops flagging it?
