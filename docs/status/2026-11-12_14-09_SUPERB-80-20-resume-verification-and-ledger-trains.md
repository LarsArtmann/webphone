# Status: SUPERB 80/20 resume — verification, ledger, dependency + prose trains (2026-11-12 14:09 CET)

**Claims verified as of 14:09 CET.** Push/CI statements rot fastest; the concurrent-session
picture changed at least four times during this session (auth/platform carves healing, the
push-Transport + fax-RelPath type trains, a `release-line` branch checkout at 14:0x).

Session scope: resume the 06-55 SUPERB 80/20 execution after the 12-06 status handoff. The
owner's blanket GO ("GET SHIT DONE! The WHOLE TODO LIST!") stood; the three owner questions
from the morning re-posed. One instruction, executed end-to-end.

---

## a) FULLY DONE

1. **State re-verification on resume.** HEAD had moved 8a319a77 → 62b627c5 while I was away:
   the sibling's `auth` carve (0849037c) landed, a `platform/` carve was in flight (staged
   deletions + untracked files), and main was red. Root-caused the red WITHOUT touching the
   sibling's files: the art-dupl gate flagged ONE new clone group from their committed auth
   carve (`auth.go` vs `internal_directory.go` JSON-header trio). Filed tracker **#246**
   naming surfaces, files, and both fix options (one-home helper or pinned baseline). The
   sibling repaired it themselves (`a960d19c`, baseline 117→110) — tracker honored, no
   duplication.
2. **F034 — fixed-element caveat print** in `scripts/ui-capture.py` main(): one print per
   full-page run naming the CDP position:fixed-at-original-viewport-bottom behavior. Ruff
   clean, self-test green (all checks incl. the 5 extended cases).
3. **f-31 — whitespace-drift CSS-indent verification EXECUTED**: planted a 1-space-indent
   CSS probe in the served tree; the flake's treefmt re-indents it ("1 changed") and `--ci`
   mode exits 1 on the non-canonical bytes; canonical bytes pass. The §d-3 incident class is
   provably caught.
4. **§f ledger strikes — 16 rows with evidence** (TODO_LIST): f-13 (tap-through GREEN 3/3 +
   the focus-trap defect story), f-14 (stack E2E touches nav only via exact `[data-tab]`
   selectors — safe by construction), f-19 (caveat print), f-20 (helpers.mjs doc line),
   f-21 (day-head class split + green suites), f-22 (focus-visible ring), f-25 (state-layer
   ruled DELIBERATE, ruling written), f-31 (probe), f-35 (z-index census + ladder comment),
   f-12 (DE lane measured, see 5), f-18 (`--hover-map` proven live), f-32 (seeded
   phone-focus shots), f-30 (art-dupl grid-lane sweep clean, registry row). Plus the
   stale-owner-gate catalog row struck as SUPERSEDED (pin already v4.14.0), and a broken
   "mirror-sync" paragraph re-joined (dprint reflow artifact).
5. **M03/f-12 — the German stacked-table lane, measured honestly.** Browser probe at a TRUE
   390px viewport: every `Zeit`/`Dauer` stacked cell fits (scrollW == clientW across 3
   seeded CDR rows), no lane-wrapper horizontal overflow, `html lang=de` confirmed. The
   probe needed a fake PBX (`/phone-api/history` + the voicemail-summary credential verify)
   and the CSRF cookie dance — five iterations, each failure a real lesson (see §d-3).
6. **M05/M06 — harness capabilities proven live**: hover-map filtered run (messages/history/
   settings × light/dark at 390x844 = exactly 6 shots), seed-class run (2 shots of the
   dialer-focus layout without taps). The F034 caveat print fired on every run.
7. **M14 — go-paperless v0.5.0 → v0.5.2 bump train, complete**: preflight (fail-open proxy
   404 recorded), `go get`+`tidy` clean single-dep delta, paperless+fax suites green, full
   suite green (one load-flake in the sibling's concurrency test, green ×5 isolated),
   vendorHash re-pinned via the goModules rebuild recipe (got: applied, build closes),
   report currency line flipped, CHANGELOG entry, audit-verify replay green.
8. **M15 — catalog adjudication with the blocker recorded**: `go list -m -versions` shows
   latest = v4.14.0 = our pin; F045's expected v4.15.0 does not exist upstream. No bump
   possible; the stale owner-gate row struck as superseded. The "E1 catalog half" in the
   owner-calls doc corrected to the truth (it had claimed the v4.15.0 bump was dispatched —
   wrong on two counts).
9. **M22 — wire contracts**: `webphone contracts` regen is byte-identical against the tree;
   README split-rule verified truthful; the three open rows re-adjudicated (all correctly
   owner/trigger-gated, no strikes owed).
10. **M23 — Datastar CSP rows adjudicated with a corrected premise**: the library's nonce
    lane SHIPPED (go-health-dashboard 449aeb3, first released v0.13.1 — older than our
    v0.13.2 pin), and webphone already consumes + pins it
    (`TestAppDashboardMountsHealthSubtree` asserts nonce + per-request rotation + inline
    nonce'd scripts). The "nonce INSTEAD OF unsafe-eval" premise is unsatisfiable: csp.go
    adds the nonce BESIDE `'unsafe-eval'` because the Datastar SDK compiles data-*
    expressions with the Function constructor regardless of nonces (three consistent
    sources). Re-open trigger moved where it belongs: Datastar upstream or a dashboard
    drop of Datastar. Two rows struck done-by-train.
11. **M24 — cross-repo 11-09 residue audited**: every storage-ladder leg (A/B/C/closeout)
    already landed 2026-11-11/12 with evidence lines in the strikes; the section's
    remaining open rows are hub-side (runner uid, ALERT_WEBHOOK_URL) and the stack-side
    ringhook docs row — correctly homed, nothing owed by this repo's code.
12. **M11/f-30 — art-dupl sweep with the CI-pinned detector v0.7.2**: the three grid lanes
    (table stack / reports / contacts) carry ZERO clone groups — CSS/templ idiom, not Go
    duplication; baseline 110 holds. Sweep-log row in docs/dedup-registry.md.
13. **The VoicemailSummary twin — adjudicated and baselined** (main red ~1h on the CI
    gate): `pbx.VoicemailSummary` (wire, json-tagged) vs `views.VoicemailSummary` (the
    vocabulary move's ratified view-side home) with `voicemailItems` as the explicit
    conversion seam — deliberate two-vocabulary similarity, BASELINE entry (the a960d19c
    writeJSON precedent's reasoning), post-write check 0 new. The sibling session
    independently reached the same verdict minutes later (`37ceda39`); the registry carries
    both lines.
14. **M17 — deep prose sweep of the ten Nov-5/6 reports** (subagent): 12 unique findings —
    10 drift-by-train (the readmodels→cqrs/gate moves dominate), 3 stale. Live reports
    fixed in place (morphToEndedCall → settleEndedCall rename provenance, ui-capture
    surface counts framed as snapshots, announcements path corrected, reverseLines row
    annotated with the train outcome); four archived reports carry ONE-LINE errata banners
    (verdicts untouched, corpus-closed rule honored).
15. **M26 — docs battery**: `scripts/docs-gates.sh --check` PASS (format, status-f-section,
    battery, table-damage, strike-safety 439 files safe); INDEX regenerated (66 reports) +
    `--check` exit 0.
16. **Artifact hygiene**: tw/login.css verified FRESH (sibling rebuilt at 12:56 — my app.css
    classes present in the inlined artifact); dprint/treefmt over every touched file.
17. **Sync checks**: dom-contract `wp-phone-focus` entry corrected to the content-region
    truth (nav stays visible); stack `browser-e2e.py` greps ZERO assumptions about
    `wp-phone-focus`/`wp-day-head`/`wp-table-day-head`/`wp-main-col`; index.html appkit row
    consistent with the concurrent re-review (both place appkit absent; rationale-only
    delta); smoke suite over a fresh binary **54 passed + all 7 scenarios green**.
18. **CI verdicts, workflow-name-verified** (`scripts/ci-verdict.sh`): my commits' lineage
    green at the time of each check (6969b7ef SUCCESS); the residual reds are the
    sibling's active type-safety trains (push-Transport defined type, fax DocumentPath→
    RelPath), each with their own land→baseline cadence — their session owns them, tracker
    precedent standing.
19. **Gated rows staged** in the owner-calls doc: M18 redeploy recipe (4 steps incl. the
    journalctl ask), M19 command + run-dir path + the creds dependency, M20 night-window
    calendar note (piggyback guidance), M21 real-inbound-call note, and **Q6 added**
    (the M19 credentials ask, now numbered in the doc).

## b) PARTIALLY DONE

1. **Full buildflow re-run after the bump train.** The 13:05 run completed but carried 3
   step-fails — all attributed to MY OWN in-flight vendorHash state (the run measured the
   tree WHILE I was re-pinning; nix-hash-fix + nix-build + test-race). The specific gates
   were verified green afterward (goModules closes, full suite green), but a CLEAN
   post-train full-buildflow pass was not re-executed. The battery-on-commit posture
   (stage everything fmt touched) was honored for the files I committed; the final state
   is certified by the targeted gates, not by a second full run.
2. **CI verdict for the moving tip.** My lineage is verified green; the tip keeps moving
   under the sibling's trains (a `release-line` branch checkout appeared in the reflog at
   14:0x — someone is cutting a release). "Main green" is only claimable per-sha.
3. **M02/M03/M05/M06** are fully executed (runs + strikes) — listed here only because the
   DE lane's verification used a throwaway /tmp probe script, not a committed harness
   assertion; re-assertability rides the TODO row's recorded evidence, not a test.
4. **E1 in the owner doc** — corrected to the truth but still needs the owner's
   ratify-or-revert on the bump-timing precedent.

## c) NOT STARTED

1. **M18** — stack redeploy + voicemail verdict recording (owner-gated; recipe staged).
2. **M19** — authenticated prod capture + gallery walkthrough (owner-gated on CREDS — the
   third ask, now Q6; command + run-dir staged).
3. **M20** — night-run stability pair (owner-gated on a quiet window; calendar note staged).
4. **M21** — ring-guard positive path during a REAL inbound call (owner-gated).
5. **M25/Q5** — the BuildFlow bandit draft banner (owner provides one line or says archive).
6. **The stack browser E2E re-run** — the runbook obligation for served-markup changes
   (the day-head rename + focus-trap fix are served bytes). Today's grep proved no E2E
   ASSUMPTION breaks, and the smoke + captures re-asserted the surfaces, but the actual
   stack-side E2E run (445s budget) was not executed this session.
7. **The sibling plan's T21 questions** (05-25 plan) — the owner doc stub awaits that
   session; unchanged.
8. **Proactive docs-health drain of the two 11-12 reports** (12-06 + this one) into
   TODO_LIST rows — the cadence gate allows it; next sweep's job.

## d) TOTALLY FUCKED UP

1. **I nearly struck f-12 on a lying GREEN.** The DE-fit probe first "passed" at
   `window.innerWidth == 500` — chromium silently CLAMPS the `--window-size` flag; my
   "390px" measurement was a 500px measurement. The page-geometry check I added on a hunch
   (inner/doc vs requested) caught it; a CDP override then made it WORSE (980) before
   `driver.set_window_size` (the WebDriver API, no clamp) gave the true 390. A verdict
   that cannot name its viewport is not a verdict — the honesty contract applies to
   measurement claims, not just UI states. The harness itself has no guard against this
   class (see §f).
2. **Six probe iterations for one measurement.** Missing CSRF header → missing cookie jar
   (the token is double-submit) → missing voicemail-summary credential verify (setting
   `phone_api_url` FLIPS the login policy from loopback-skip to PBX-verify) → missing
   `/phone-api` path prefix → driver resolution (Selenium Manager can't exec on nix) →
   the viewport saga. Each failure was legible in hindsight; the sequence was not
   cheap. The reusable parts (fake PBX shape, credential-verify coupling, path prefix)
   are only in this report and the TODO row, not a committed probe.
3. **The commit-race dance, three times.** The daemon swept my staged files between
   `git add` and `git commit`: `585f630a` ("deps: bump go-paperless…") actually carried
   CHANGELOG+TODO_LIST while the DEP files rode daemon batch `e924ee0a` (the daemon
   bypasses the commit-msg hook — the documented advisory-guard hole, live again);
   `6969b7ef` ("M17 prose-sweep…") landed as 2-of-10 files with the other 8 in daemon
   batches. I verified every file reached origin (the end states are correct), but my
   commit MESSAGES misstate their contents and the attribution is buried — exactly the
   "narrative eaten by chore commits" class the doctrine warns about, re-loaded.
4. **I ran buildflow WHILE invalidating it.** The train-start buildflow (owed from the
   previous session) was started at 13:05 and then the vendorHash re-pin landed under it —
   3 step-fails were my own moving tree, and I burned a diagnostic cycle reading `buildflow
   history` to attribute them. A full gate must own a frozen tree: run it BEFORE a bump
   train starts OR after it lands, never astride it.
5. **I raced myself.** Two parallel `node --test` invocations of the same island suite
   (sent in one message as "independent" calls) produced an exit-1 I then spent two
   diagnostic rounds attributing — it was my own parallelism, the suite is green (333/333
   serial). Same suite, same host: serial, always.
6. **The ci-verdict label "RED" for pending runs.** I reported "RED: CI concluded
   ['pending']" twice — a pending run is not a red; the wording overstates. (The script's
   exit code is honest; my narration wasn't.)

## e) WHAT WE SHOULD IMPROVE

1. **Viewport honesty belongs in the harness, not in probe scripts**: ui-capture should
   assert `window.innerWidth == requested viewport width` (± scrollbar) before shooting —
   the clamp class would otherwise mislead every narrow-viewport run silently.
2. **Commit-race protocol for dep files**: stage go.mod/go.sum/vendorHash in a commit with
   NO other files and issue add+commit as the very next statement; treat every other file
   as daemon-owed attribution. Fighting the daemon for narrative messages on non-guarded
   files is wasted effort — verified end state is the contract.
3. **Never run the same test suite twice concurrently on this host** — parallel tool calls
   with the same underlying node/go test process clash.
4. **A reusable loopback-CDR probe** (fake PBX with `/phone-api/history` +
   voicemail-summary verify + the CSRF dance) would make DE/i18n lane checks a one-command
   affair; it lives now only as this report's prose and a /tmp script that dies on reboot.
5. **Duplicate-adjudication cost is real**: both sessions adjudicated the same twin within
   minutes. The tracker-issue pattern (file FIRST, name the surface, let the owning
   session pick extract-or-baseline) scales better than waiting-then-doing — #246 worked
   exactly this way; the twin red would have too, if I'd filed instead of baselining at
   the one-hour mark.
6. **E1-style owner-doc claims need a same-session truth check** — the doc said the
   catalog bump was "dispatched" when the resolvability check had not run; I inherited
   the claim and corrected it, but the check should have gated the original wording.

## f) Up to 50 things we should get done next

**Owner-gated (blocking):**

1. M19: provide prod capture creds → run the tracked capture + gallery walkthrough (Q6).
2. M18: stack redeploy riding main + record the live voicemail-play verdict.
3. M20: the two quiet-window tracked runs ≥8h apart → badge-threshold calibration.
4. M21: ring-guard positive path during a real inbound call in a tracked run.
5. Q5/M25: BuildFlow bandit draft — provide the banner line or archive.
6. Q1: Phone-tab label i18n ruling (D3 stands by default).
7. Q2: empty-state action buttons (decline stands by default).
8. Q3: index-render.jpg refresh vs caption.
9. Q4: templ-components taste re-litigation (declines stand; name a component to re-open).
10. Q6: the creds answer itself (or order M18 before M19).
11. E1: ratify or revert the bump-timing precedent (go-paperless executed same-day).
12. The sibling plan's T21 questions (CRM trio, cqrs-lint gating, C044, B5, traces-in-
    backup, M1-first-live) — that session's stub awaits.

**Harness + verification (self-dispatchable, small):**
13. Add the viewport-honesty guard to ui-capture (fail/warn when innerWidth ≠ requested).
14. Commit the loopback-CDR probe as a real script (`scripts/` or a ui-capture mode) so DE
lane checks are re-assertable.
15. Re-run the full buildflow on a frozen post-bump tree (close §b-1).
16. Run the stack browser E2E against current main (close §c-6).
17. File the `--window-size` clamp gotcha into ui-capture's docstring/AGENTS recipe.
18. Deduplicate the twin adjudication rows in dedup-registry.md into one canonical line
(mine + the sibling's overlap).
19. Sweep the 12-06 + this report into TODO_LIST rows at the next docs-health cadence.
20. `--hover-map` selector verification: log when a selector matches nothing (silent no-op
today is by design but unlogged).
21. Add a `--seed-class` summary.json ledger field so seeded runs are diffable against
baseline runs.
22. The `/tmp/f12-probe.py` geometry check (inner/doc/body vs requested) → fold into the
harness post-shot assertions.

**Ledger + corpus:**
23. TODO_LIST voicemail-playback production legs (row exists; M18's verdict feeds it).
24. The traces-in-backup owner ruling (f-22 routed row) — still unanswered.
25. Dialer-focus return pattern (owner question 2 from the morning — still open in the
ledger even though the trap is fixed).
26. Ringhook operator docs + typed mirror row (stack-side deploy docs).
27. Hub runner durable fix + ALERT_WEBHOOK_URL decision (hub-side rows).
28. webphone-android runtime digest handshake (trigger-gated; first skew incident).
29. Wire-fixture follow-ups (a) android loopback live-check, (b) /config.js key fixture
(trigger-gated).
30. OpenAPI golden-fixture owner call (the /openapi.json conditional-GET row's remainder).

**Type-safety roadmap (open items from the 2026-10-25 review — the sibling is actively
eating these; coordinate, don't fork):**
31. `sso.Finish` value object into Complete (auth: convention).
32. asr/matrix wiring structs.
33. `machineSurface` bool-last ordering.
34. A1.7 out-of-order delivery verdict tightening (owner call).
35. D4 admin delegation row (write-shaped admin actions).
36. The transport-isolation pattern audit (post-Isolated sweep — any new bare clients?).

**Fleet-libs + currency:**
37. go-appkit flightrecorderhealth module-level adoption (demand-gated TODO row; ask #59/
#60/#61 upstream status watch).
38. go-sse/server_timing 67-scores: re-examine after the next dashboard wave.
39. templ-components 50-score: the cheap-wins verdict is landed; the re-score waits for
the next upstream train.
40. catalog/v4: watch for the v4.15.0 release (bump rides that train; F045's blocker).

**Hygiene:**
41. Audit-verify corpus: the 105-claim set is green; quarterly re-measure of the floors
(`>= 115` rode errorfamily 131→124 — re-derive at the next sweep).
42. depthbudget_test.go: the sibling's carves moved packages — re-measure the pin table.
43. `internal/server` carve completion: the platform/ move in flight + auth/ landed — the
TODO_LIST carve row's remaining half needs the post-carve verdict.
44. Errata-banner convention: decide whether ARCHIVED banners get a standard ERRATA shape
(mine are ad-hoc one-liners; the docs-gates annotate tool could own them).
45. ChangeLOG: the day-head class rename rides the earlier train's entry — verify the
[Unreleased] block reads coherently after three sessions of parallel edits.
46. The `release-line` branch that appeared in the reflog — reconcile with the release
runbook (whose session cut it, what rides it).
47. Tracker #246 + the twin: close both issues/rows with the landing shas once the
sibling's train settles.
48. UI-token mirror check: island/style.css ↔ app.css token blocks (the §f-25 ruling
noted deliberate divergence; a diff pass keeps it deliberate).
49. The 2026-11-12_12-06 status report's §f list: items 15-21 of MY today-list are done —
strike them in that report's table or fold into the ledger (avoid two competing
next-step lists).
50. Post-concurrent-train smoke: after the sibling's Transport/RelPath trains settle, one
full smoke + island suite pass as the cheap regression lane (they touch wire shapes).

## g) Questions for the owner (cannot self-answer)

1. **M19 credentials (fourth ask)** — `WEBPHONE_CAPTURE_EXTENSION` +
   `WEBPHONE_CAPTURE_PASSWORD` for pbx.artmann.tech, or confirm "M18 first" ordering so
   the deploy mints its own access. The command is staged and verified against the
   harness; only the creds block the owner-eyes gallery walkthrough.
2. **Dialer-focus return pattern** — the trap is fixed (nav stays visible), but the design
   question stands: is nav-as-return the final pattern, or do you want an explicit
   in-island "back" control in the dialer view?
3. **Q5 bandit banner** — one line of owner-banner text for the BuildFlow upstream issue
   draft, or say "archive" and the zombie draft dies at the next sweep.
