# Whole-list execution: T23 android landed, pma push-heat gate shipped, hub-fix verification interrupted (2026-11-21 13:13)

> Claims verified as of 13:13 CET — push/CI/load statements rot fastest.
> Session mandate: the standing "WHOLE TODO LIST" directive (22:05 pareto
> plan), resumed from the 03:16 report's ranked queue. Predecessor:
> `2026-11-21_03-16_whole-list-resume-red-heal-and-crossrepo-trains.md`.
> Interrupted mid-verification by the owner's status order.

## a) FULLY DONE (this session, verified)

1. **Session-start recon + bar-fired pushes**: webphone main was >4h
   unpushed (92174cb9 at 03:16) — pushed per the ratified manual-push
   bar; the stack was 4h unpushed — pushed; CI verified GREEN at
   webphone origin/main via `scripts/ci-verdict.sh` (re-verified again
   at 13:13, still green). GitHub API quota restored (4869 remaining).
   Host load started at 18.28 and only got worse (92.72 peak).
2. **T20 verified as ALREADY DONE by siblings** — no work owed, rows
   already struck: the stack coverage arms (`voicemail_playback_arm` in
   tests/browser-e2e.py with the honest-toast contract + the
   206/200-through-vhost observables; verify-live.sh's authenticated vm
   audio fetch; operator.nix's Range 206 + through-vhost seek arms),
   the drift audit R04 (NO DRIFT both directions,
   `TestModuleGoldenSettingsAreKnownConfig` owns the fence), and the
   ringhook docs/health-leg half. The 03:16 report queued these; the
   sibling's T17 audit sweep had already landed and struck them.
3. **T12 lens stack-leg verified + webphone row struck**: the stack's
   HEAD (53ab2cc, pushed this session) carries the full typed front —
   `services.telephony.webphone.lens.url` + `lensContextDays` options,
   the `lib.optionalAttrs` settings arm (byte-identical off), and the
   `webphoneLensCheck`/`webphoneLensNegativeCheck` eval arms wired into
   telephony-eval (eval built green 03:1x). Webphone TODO row 1758
   struck with evidence (daemon-swept; committed).
4. **T23 android `lines[]` adoption — COMPLETE, green, pushed**
   (daemon-swept into aec9957/520cef5/bf33458/e0787d9; Forgejo
   origin/main == HEAD, verified):
   - Picker UI in the wake-detail sheet (`renderLineRows`): a
     "Your lines" section between the last-checked line and Retry,
     visible ONLY for multi-line accounts, rows DID-first
     ("+15551234567 · 1001", bare extension fallback), the active line
     (lines[0] by the server contract) accent-marked and inert, every
     other owned line tappable.
   - `switchToLine`: live-call guard (toast "Hang up before switching"),
     the switch POST on a background thread, verdict landed via
     `runOnUiThread`; `applySwitchVerdict`: SWITCHED dismisses the
     sheet, toasts the landed line, stops the ntfy lane, reloads the
     web surface, and `CallLayer.rewire`s; refusals keep the sheet and
     say WHICH refusal (file-level `switchToastText` table, 8 new
     strings, English-only per the house convention).
   - Tests: `WireContractTest` +4 (the typed `lines[]` wire active-first
     with the parallel `numbers`/`extensions` lists asserted GONE, the
     key-less old-server fallback, DID-first labels),
     `SessionApiSwitchVerdictTest` (5 classifier pins),
     `MainActivityWakeDetailTest` +3 (picker render incl. single-line
     hide, SWITCHED dismiss+toast+lane-teardown, refusal keeps sheet).
     Fixtures re-vendored (`sync-wire-fixtures.sh`): the brokered
     fixture now carries `lines[]`, drift closed. Gates:
     `testDebugUnitTest` green (suites 7/15/5 tests, 0 failures),
     `ktlintCheck` green after two wrap fixes. The model/parse/switch
     half (SessionApi/CallLayer, landed last session) was
     daemon-committed as 2cf3c15 and rides the same push.
5. **T21 pma push-heat gate — code half LANDED in pma** (commits
   177da3c2 daemon-swept + `1bf82129` mine; pma-daemon module fully
   green, treefmt'd): `CommitSettings.PushDeferSeconds` (default 0 =
   off — the deployed daemon is byte-neutral until armed),
   `commitWithPushHeatGate` (the ONE commit entry for batch + SIGTERM
   flush: a hot tree commits WITHOUT the immediate push — the
   commit-race class is a PUBLISH problem, not a commit problem), the
   `pushProject` heat gate (startup sweep + flush), the per-repo
   cool-down timer (deduped, context-guarded, Stop-cancelled, recovered
   by the next startup sweep), and the `CommitWithoutPush` seam (shares
   history — no MinInterval split brain). Pinned by
   `pushheat_test.go` (heat signal: .git writes are NOT heat; knob
   arithmetic; unstatable = cool) + `commitwithoutpush_test.go`
   (AutoPush ON, push leg suppressed, commit lands). Mirror row PD in
   pma's TODO_LIST with the canary/rollback story; webphone's row 69
   struck with the pointer. NOT pushed — pma's own TODO P1 rule:
   never push without explicit owner approval.

## b) PARTIALLY DONE

1. **T21 hub runner durable fix — strike evidence ~90% gathered, strike
   not written** (interrupted here): the fix LANDED 2026-11-13
   (SystemNix forgejo.nix: UMask=002, setgid 2770 roots, root
   ExecStartPre repair leg) and has since EVOLVED beyond the row's
   description — the host now runs
   `forgejo-runner-evo-x2-cache-repair.timer` (ACTIVE; periodic repair
   decoupled from the start chain) beside the active runner unit, and
   hub CI runs #352–#370 (2026-11-18/19, the canary lane proven green)
   built through that machinery live. The row's strike condition ("one
   real switch + one CI run pass with the leg live") is met by that
   evidence — the hub TODO edit is the only missing half.
2. **T23 android verification tail**: the FULL `scripts/verify-gate.sh`
   (assembleDebug → zipalign → abi → contrast → dprint legs beyond the
   unit+ktlint gates) was started in the background — its verdict was
   LOST (see §d3). The github mirror (origin is Forgejo) push state is
   unchecked (`scripts/mirror-lag.sh` exists for exactly this).
3. **Release v2.15.0**: preconditions re-verified this session (NO tag
   local or remote; fold on main; CI green; no sibling train alive),
   but the runbook's load gate never opened — host load ranged
   7.8→18→92 on the 1-minute average all session (gate wants <8 on 1+
   5-minute). Not hand-fired; the deliberate sequencing (bumps after
   the tag) therefore also did not start.

## c) NOT STARTED (from the ranked queue; unchanged from 03:16 §f)

- Post-release legs: ci-verdict on the release sha, stack relock
  verification, aarch64 cross-build, gh release notes.
- Bump trains 1–4: go-cqrs-lite engine wave (v4.19.0/v4.26.0/v4.9.1) →
  templ v0.3.1118 + CLI-paired regen → modernc sqlite v1.62.0 (+
  aarch64 + backup-drill belt) → small SDKs ×5 + TimingScale
  retirement via `idempotency.WithClock` (templ-components v1.27.2
  changelog read rides the preflight).
- 18-08 docs drain (102 bare §a/§b rows; `status-predrain.sh` names
  the READY set; due ≤2026-12-07).
- One KVM `webphone-backup` re-prove under commonHardening;
  `build-tw-css.sh --check` post-lens sanity; watcher fire-drill
  canary.
- ALERT_WEBHOOK_URL presentation (owner channel pick → pack §G).
- Final batteries (full no-cache buildflow, docs-gates --check, race
  suite) and the final report-back rendering both plan tables (the
  sibling report's §c4 debt).

## d) TOTALLY FUCKED UP (honest ledger)

1. **A multiedit leg targeted the WRONG FILE** (flush.go's call-site
   text pasted into the operations.go edit set): "Applied 1 of 2"
   silently half-landed. Caught by reading the result line, healed by
   View+re-edit. Same class as the last session's missed leg — the
   lesson (one file per edit set, verify every leg's target) keeps not
   sticking.
2. **Three edit-tool "must read first" rejections** (pma batch.go,
   service.go, TODO_LIST): I had bash-grep/sed context but not
   View-tool reads. No damage (the tool is the guard), but three lost
   cycles to a rule I know.
3. **A gate verdict was LOST**: the android verify-gate ran as a
   background job whose shell was wiped between turns — no recorded
   green. A gate without a persisted verdict did not run; rerun owed
   (§f2).
4. **The webphone TODO strike edit raced the daemon twice** (file moved
   from row 69→122 under me mid-edit): first edit rejected on
   mod-time, healed by re-View. Self-inflicted: I edited a shared file
   minutes after watching the daemon sweep it.
5. **Idle gap 08:50→13:13**: the hub-runner verification sat
   half-finished across the gap; evidence gathered at 08:5x, strike
   still unwritten at 13:13. Finish what a turn starts or write the
   state down at the turn boundary.

## e) WHAT WE SHOULD IMPROVE

1. **Persist background verdicts to disk** (`… 2>&1 | tee /tmp/<gate>.log;
   echo rc=$?`), never trust the job system across turn boundaries —
   §d3 is the second lost-verdict incident.
2. **One file per edit-set leg**, and re-verify the target path of
   every leg before firing (§d1).
3. **The release load gate needs a decision, not patience**: two
   sessions have now idled against a host that never drops below 8
   (44 users, CI runners, sibling sessions). Either the scheduled
   quiet-window watch-loop (fires release.sh when the gate passes) or
   an owner override ruling — see §g1.
4. **Commit at the turn boundary**: narrative commits were eaten by the
   daemon twice this session (webphone strike, android train) because
   gates finished and the turn ended before `git add+commit` ran in
   the SAME call. The pma commit shows the correct shape.

## f) NEXT (ranked)

1. Write the hub runner row strike (evidence in §b1) into the hub's
   TODO_LIST + commit.
2. Re-run android `scripts/verify-gate.sh`, persist the verdict; run
   `scripts/mirror-lag.sh` and push the github mirror if behind.
3. ALERT_WEBHOOK_URL → owner pack §G ruling G9 (ntfy vs email vs
   Forgejo webhook; needed before hub failures alert anywhere).
4. Release v2.15.0 the moment the gate opens (or per §g1's answer);
   tag check first, clean clone, `scripts/release.sh 2.15.0`.
5. Post-release legs: ci-verdict to green, stack relock verified,
   aarch64, gh release, KVM backup re-prove, tw-css --check.
6. Bump train 1: go-cqrs-lite engine wave (preflight → tidy →
   vendorHash → CI-watch).
7. Bump train 2: templ v0.3.1118 + CLI-paired views regen.
8. Bump train 3: modernc sqlite v1.62.0 + aarch64 + backup-drill.
9. Bump train 4: small SDKs ×5 + TimingScale retirement (read
   templ-components v1.27.2 changelog in the preflight).
10. 18-08 docs drain (pre-drain automation idea from T08 rides it).
11. Watcher fire-drill canary.
12. Final batteries: full no-cache buildflow, docs-gates --check, race
    suite, docs battery; final report-back with both plan tables.
13. pma deploy/canary leg once §g2 is answered (rebuild daemon, arm
    `push_defer_seconds` on webphone as the canary, watch a day, fleet;
    rollback = knob to 0).
14. Host units noticed FAILED while verifying the runner:
    `btrbk-forgejo-clean.service` + `forgejo-backup.service` — verify/
    file (owner-visible; found by side-glance, not audited).
15. Stack: the sibling's landed-red sso-only fixture (`53ab2cc` carries
    the red VM suite; their diagnosed one-line rp-pin fix-forward) —
    leave to them, re-verify at the next stack touch before any
    `nix flake check` there.
16. Prod verification windows A/B once the owner answers deploy (G7)
    and capture credentials (G8) — unchanged gates.

## g) QUESTIONS FOR THE OWNER

1. **Release load override?** The <8 gate has now blocked two sessions
   on a host that idles above it. Options: (a) I fire
   `WEBPHONE_RELEASE_MAX_LOAD=999` now accepting E2E/VM flake risk,
   (b) a scheduled watch-loop fires release.sh only when the gate
   genuinely passes (recommended), or (c) keep manual-only and you
   name the window.
2. **pma deployment authorization**: the push-heat gate is committed
   (unpushed — pma's never-push rule). May I push pma master, rebuild/
   redeploy the daemon, and arm `push_defer_seconds` on webphone as
   the one-repo canary — or is the daemon's deploy nix-config-side
   owner-only?
3. **Hub ALERT_WEBHOOK_URL channel** (the 2026-11-08 row's open
   decision): ntfy, email, or a Forgejo webhook — hub failures
   currently alert nowhere. This is a channel + secret pick only you
   can make.
