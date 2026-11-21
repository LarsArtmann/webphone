# Whole-list resume: red-main heal + governance rows + cross-repo trains (2026-11-21 03:16)

> Claims verified as of 03:16 CET — push/CI/load statements rot fastest.
> Session mandate: resume the 22:05 pareto whole-list execution after the
> 02:28 hard wait. Predecessor: `2026-11-21_02-28_whole-list-execution-midtrain-status.md`.

Landscape on entry (02:51): the concurrent release train had stalled
PRE-TAG (`2be14b73` bump + `aa6245b8` merge pushed, no `v2.15.0` tag
local or remote — still true at 03:16); a sibling session was live
(T15-T18 trains, bandit filing, report harvesting — my 02-28 report was
committed AND harvested by them, INDEX fresh `--check` rc=0); and **main
was RED** — the gate repair this session opened with.

## a) FULLY DONE (this session, committed + verified)

1. **Red-main heal** (CI run 44850987580 failed on `0f5690dd`, ~25 min
   red): BOTH failures were my 02:2x trains' residue. (i) art-dupl: 2 new
   clone groups from the T14/lens history.templ edits shipped WITHOUT
   their baseline ride-along — adjudicated (both the standing chip/span
   idiom classes: the 2026-11-18 conditional-strong family + the
   admin.templ span/timestamp class), baseline RE-RECORDED with the
   CI-pinned detector v0.7.2 (112 → 114, post-write `check` = 0 new),
   dedup-registry sweep row appended. (ii) tw-freshness: locked-nixpkgs
   drifted tailwindcss 4.3.4 → 4.3.5 so the committed artifacts went
   stale with ZERO source change — `build-tw-css.sh` rebuild (banner-only
   diff; the daemon swept the artifacts into `404da12c`). Commit
   `9077bc28`, pushed; fix verified by construction (local check + the
   two CI gates re-run inside later pushes).
2. **Error-contract detection column (L9) — the one undone small row**:
   the 2026-10-29 section existed with six rows, but its own ride-along
   contract was violated by three post-10-29 classes. Added rows for the
   projection DLQ (WARN per landing + boot `warnDLQDepth` re-WARN + the
   passkey-gated `/healthz` projections leg), the detached review job
   (the `failed` review state + the `review: transcript fetch failed`
   WARN), and the backup oneshot (systemd journal + the KVM-gated
   drill check). Signatures verified against `dlq.go`, `review_api.go`,
   `push.go`; chat-receiver was already covered in the main table.
   TODO L9 struck with evidence. Commit `10e88574`.
3. **T25 owner-pack §G addendum + reopen-trigger sweep (m25.1+m25.2)**:
   §G appended to the 2026-11-07 sitting pack — 8 rulings in the one
   answer surface (deepening pick 1-of-5, bump-train timing, the
   pbx-lens 5-question register, DOMAIN_LANGUAGE ratify-or-delete, §5
   DROP date, §D10 file-size ratification, v2.15.0 deploy authorization,
   prod-capture credentials), answer-block line added. The registry
   sweep recorded ZERO fired triggers with its machine-checked arms
   named (go-appkit remote tags top out at v0.12.0; templ-components
   v1.27.2 is a version delta whose trigger is component-shaped — it
   rides the small-SDK bump train's preflight read). Journey arms
   already closed by siblings (dial-gate + voicemail 2026-11-20; appkit
   demand check HOLDS). Commit `560edfa2`.
4. **T12 stack leg — the lens typed front** (nix-international-telephony,
   daemon-swept into `3130a62`/`53ab2cc`): `services.telephony.webphone.lens.url`
   (str, default "" = the rendered settings carry NO lens key,
   byte-identical off-arm), the `lens.url` settings arm in web.nix, and
   `webphoneLensCheck` + `webphoneLensNegativeCheck` eval arms wired
   into telephony-eval's inherit + shell loop. **telephony-eval check
   built GREEN** (`nix build --no-link .#checks.x86_64-linux.telephony-eval`,
   derivation built, no FAIL line). G3(d) in the owner pack is now a
   ratify-only row.
5. **T20's webphone-side ringhook ruling: verified ALREADY DONE** by the
   sibling (pareto T18 recorded the DECLINE with reasoning — the shim
   arms are intrinsically stack-side). No duplicate work; closed by
   verification.

## b) PARTIALLY DONE (in flight at report time)

1. **T23 android `lines[]` adoption** (webphone-android, 2 files
   UNCOMMITTED): `SessionApi` gained the `SessionLine(extension, did)`
   model with the DID-first `label()` (webphone's one convention), the
   `lines[]` parse (active first; an old server's key-less answer
   degrades to the single top-level line), the `SwitchVerdict` enum +
   pure classifier (200/403/404/null), and `switchLine()` riding the
   same CSRF dance as the push register; `CallLayer.lastIdentity`
   exposes the resume-resolved identity for the picker. REMAINING:
   MainActivity picker rows in the wake-detail sheet (tap a non-active
   line → switch → `NtfySubscriptionService.stop` + `webSurface.retry()`
   - `CallLayer.rewire`), strings, WireContractTest lines assertions +
     verdict-classifier pin, fixture re-vendor (`sync-wire-fixtures.sh`
     --check is RED as expected: vendored `numbers/extensions` vs server
     `lines[]`), then ktlint + JVM suite green.
2. **v2.15.0 release**: preconditions all hold (fold on main, FEATURES/
   TODO synced, version pinned, red healed); BLOCKED on the runbook's
   load gate — host load sat at 1066-1100 the whole session (39 D-state
   - self-hosted CI runners + multiple sibling sessions; the gate wants
     <8, the CRM sibling's watcher wants <40). Also the GitHub API rate
     limit was EXHAUSTED at 01:59 (watch daemon + sessions) — CI-verdict
     polling degraded until the hourly reset. No tag fired; no
     double-train risk taken.

## c) NOT STARTED (from the ranked remainder)

Bump trains 1-4 (deliberately sequenced AFTER the tag); T20 stack
coverage arms (browser-E2E voicemail lane, verify-live authenticated
audio fetch, operator.nix 206-through-proxy pin); T21 fleet legs (pma
push heuristic, hub runner durable fix, ALERT_WEBHOOK_URL decision);
the 18-08 docs drain (due ≤2026-12-07); watcher fire-drill; KVM backup
re-prove under commonHardening; `build-tw-css.sh --check` post-lens
verify (low risk: `wp-mini` reuses an existing utility).

## d) TOTALLY FUCKED UP (honest ledger)

1. **My own trains reddened main** (the 0f5690dd CI red): the T14/lens
   history.templ edits shipped without the art-dupl baseline ride-along
   AND without running `scripts/pre-push-gates.sh` — the registry now
   carries SIX rows of exactly this class with my name on the latest.
   The gate exists precisely for this and I skipped it twice in one
   night (the 02:2x session and its residue here).
2. **Two edit-tool failures from reconstructed text** (guessed
   `INVALID` where the file said `AUTH`; a 3-way multiedit where one
   leg missed): both retried green after re-reading exact bytes, but
   both were preventable View-first violations under speed pressure.
3. `fleet-tags.sh` invoked bare (usage error — it needs `--repos`):
   wasted a cycle; replaced with direct per-repo `ls-remote` checks for
   the two rows that mattered.

## e) WHAT WE SHOULD IMPROVE

1. **Run `scripts/pre-push-gates.sh` before EVERY code push** (art-dupl
   new-only + registry freshness): it is cheap, exists, and would have
   caught §d1 both times. Consider wiring it into this session's
   standing checklist alongside ci-verdict.
2. **The release load gate needs a realistic quiet-window strategy** on
   a host that now idles at load ~1000: either a scheduled 3-4 AM
   watch-loop that fires release.sh when the gate passes (the
   buildflow-when-quiet pattern from the CRM sibling, lifted to
   webphone with the <8 threshold), or an owner ruling to raise
   WEBPHONE_RELEASE_MAX_LOAD for IO-shaped load. Do NOT hand-fire it
   mid-storm.
3. **gh API rate limit is a fleet-level exhaustion** (watch daemon +
   sessions + CI polling share one budget): ci-verdict should
   exponential-backoff and the watch lane's poll cadence deserves a
   look before the next multi-session night.

## f) NEXT (ranked)

1. Finish T23: picker UI + strings + tests + fixture re-vendor; ktlint
   - `testDebugUnitTest` green; commit + push (mirror remotes).
2. Re-verify the stack T12 landing survived the daemon sweep intact
   (`git show 53ab2cc --stat` + one more telephony-eval build), then
   strike webphone TODO's T12/T26 stack-leg rows.
3. Watch for a quiet window (load <8): fire `scripts/release.sh 2.15.0`
   from a clean clone; confirm no sibling fired it first (tag check
   stays step zero).
4. Post-release: `ci-verdict.sh` to green; stack relock + browser E2E +
   VM + flake check; aarch64 cross-build; gh release notes.
5. Bump train 1: go-cqrs-lite engine wave (v4.19.0/v4.26.0/v4.9.1) —
   tidy → vendorHash → CI-watch.
6. Bump train 2: templ v0.3.1118 + CLI-paired regen (views + fixtures).
7. Bump train 3: modernc sqlite v1.62.0 (+ aarch64 + backup-drill belt).
8. Bump train 4: small SDKs ×5 + retire the TimingScale band-aid via
   `idempotency.WithClock`; read templ-components v1.27.2's changelog in
   the preflight (the registry's component-shaped trigger).
9. T20 stack arms: voicemail browser-E2E lane; verify-live
   authenticated audio fetch; operator.nix 206-through-proxy pin.
10. T21: pma push heuristic (age gate + canary + rollback story);
    hub runner ExecStartPre provisioning; ALERT_WEBHOOK_URL decision.
11. 18-08 report drain train (102 bare rows; status-predrain.sh names
    the READY set) — before ~2026-12-07.
12. One KVM `webphone-backup` run to re-prove the oneshot under
    commonHardening.
13. `scripts/build-tw-css.sh --check` post-lens sanity.
14. Watcher fire-drill (deliberate canary through the systemd user env).
15. CI watch to name-verified green on the final push of the night.
16. Owner sitting on pack §A-G (the §G addendum is the newest sheet).

## g) QUESTIONS FOR THE OWNER (the three that gate the most)

1. **Deploy v2.15.0?** (pack §G7 / the 00-51 §g Q2): the fold is on
   main, red healed, release staged behind the load gate — ssh to
   pbx.artmann is owner-only either way.
2. **Ratify §D10 = `max_file_size: 1500`?** (pack §G6): the reversible
   default stands unless overridden; two clean no-cache batteries
   proved 47 → 0 findings.
3. **Provision the prod capture credentials?** (pack §G8 / §g Q3):
   `WEBPHONE_CAPTURE_EXTENSION`/`_PASSWORD` unblock T02/T03 (the
   authenticated capture matrix, night-run pair, ring-guard positive
   path) — the only blocker left on those rows.
