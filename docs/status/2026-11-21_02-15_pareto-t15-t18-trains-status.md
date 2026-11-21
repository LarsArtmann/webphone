# Pareto execution — T15–T18 trains landed; T19+ ladder open (session status)

> **Snapshot-rot header:** claims verified as of **2026-11-21 02:15 CET**
> (push/CI statements rot fastest — they are timestamped, not timeless).
> Plan under execution:
> [docs/planning/2026-11-19_19-46_SUPERB-pareto-backlog-execution-plan.md](../planning/2026-11-19_19-46_SUPERB-pareto-backlog-execution-plan.md)
> (T01–T27 + micro m01–m27). Prior session reports:
> [2026-11-20_12-20](2026-11-20_12-20_pareto-trains-t10-t15-midflight-status.md),
> [2026-11-20_18-13](2026-11-20_18-13_pareto-t15-gap-and-red-main-heal-verification.md).
> Concurrent sessions were ACTIVE throughout (verification-ring TA–TH train,
> a docs battery session, a history/lens train landing `f251aa33`/`6fbd26dc`,
> and the push-seam 403-prune session) — boundary named per item below.

## a) FULLY DONE this session (all committed + pushed)

1. **T15 closeout — the review's pin tests, plus a REAL bug they caught**
   (`e404ac74` + daemon-swept code):
   - `TestFcmTokenMintIsSingleFlight` (internal/push/fcm_internal_test.go):
     8 concurrent sends against a parked leader → exactly ONE grant
     exchange, ALL shippers carry `Bearer shared-token`, cached follow-up
     makes no new grant. Writing it exposed that the committed joiner
     shape (copied from `sso.providerFor`) shipped the ZERO result to
     every joiner that missed the direct channel handoff — an empty
     Bearer, then a 401 invalidating the freshly-filled cache. Fixed
     with the loop-back (joiners re-read the cache on the zero result);
     `sso.providerFor` carried the identical latent shape (nil provider,
     nil error) and got the same fix. **The bug is real, not theorized:**
     an isolated probe of the OLD shape (`/tmp/sfprobe`) reproduced
     `empty tokens: 2 of 8` on run 5 of 6. Suites green: push, sso, blob
     (incl. `-race`), plus the session-behavior suite (`TestServer`) for
     the sso touch. Ledger: CHANGELOG Fixed entry + R37 strike corrected
     to name the REAL pin (a prior strike named the wrong test).
   - `TestSaveMalformedExtensionLeavesNoOrphanedFile`
     (internal/blob/store_test.go): a `ParseRelPath`-refused extension
     errors BEFORE any bytes land; walk proves zero files remain — the
     brand-before-write reorder's edge, now pinned.
2. **T16 — cadence pre-drain automation** (`2e2cba46`): new
   `scripts/status-predrain.sh` classifies every unarchived report
   READY/BLOCKED through the REAL archive gates (annotate
   `check-rows --strict` + `bare-rows`, tool auto-built per session),
   oldest first, printing the exact banner + `git mv` ritual per READY
   file and the blocking harvest line per BLOCKED one. `--self-test`
   pins the classifier (the self-test itself caught my first wrong
   fixture — a strike needs `~~` in EVERY cell). Wired into
   `scripts/docs-gates.sh` as the advisory `predrain` WARN lane (never
   fails the battery; SKIPs in CI like table-damage). **First run found
   the pin at 66/66 with 14 gates-green candidates; the 8 oldest were
   drained the same train (66 → 58), full battery + `--check` mode
   green.** R36 struck with evidence.
3. **T17 — stack drift audit + coverage arms** (`7601f37f`,
   `752dd1cc`): verdict **NO DRIFT**, now permanently pinned —
   `TestModuleGoldenSettingsAreKnownConfig` (internal/config/moduledrift_test.go)
   feeds BOTH module-output goldens' rendered settings through the same
   `unknownConfigPaths` walk the boot WARN uses, so a module writing a
   key the binary dropped reds at the repo owning both halves. The
   stack-side writes were greped in context (`sip.credential_secret_file`,
   `textagent.secret_file`, `asr.provider`, gateway/csrf/timezone/ice/
   identities — all known; the `nix.settings.*`/`telnyx-webhooks` hits
   were foreign attrspaces). The R08 arms (browser voicemail playback,
   verify-live vm audio fetch, operator.nix 206-through-the-proxy) were
   verified ALREADY LANDED in the stack tree — struck as such. Rows 171,
   207, 1592 struck.
4. **T18 — ringhook health leg + docs + mirror verdict** (stack
   `fe6eb88`, webphone `3f9665db`): `telephony-health` now probes the
   shim's `/ringhook/health` per tick when `ringhook.enable` — the shim
   is the one stack service whose death is SILENT by contract. Two new
   needles in `telephony-eval`'s `ringhookCheck` (fixture now enables
   `monitoring`), check built GREEN; runbook ringhook section gained the
   continuous-arm bullet; stack CHANGELOG entry. The typed webphone-side
   mirror is DECLINED with reasoning recorded in the R16 strike (the
   shim exists BECAUSE the store-rendered dialplan cannot hold a secret
   — its arms are intrinsically stack-side; webphone's module already
   exposes the receiver's complete surface).

## b) PARTIALLY DONE

- **T24 (notes-in-search corpus)**: a concurrent session's
  `f251aa33` ("note prose joins the query corpus; depth pin rides the
  feature") looks like the T24 spike LANDED by them. NOT verified by me
  (their train, their evidence); the TODO row state must be re-read
  before T24 is claimed — do not double-claim.
- **CI verification of my last two pushes** (`752dd1cc`, `3f9665db`):
  runs pending at report time (pushed 01:42/02:05; full run ≈12–16 min).
  See (d) — the earlier pushes' red runs were diagnosed and are
  explained by the two-session regen race, with current HEAD green
  locally for every blamed test.

## c) NOT STARTED (the remaining ladder, plan §2 order)

T19 (android `lines[]` wire + `WireContractTest`), T20 (android spec
consumption + wire-fixture triggers), T21 (SSO-only fixture end-to-end),
T22 (fleet health batch: pma push heuristic, hub runner durable fix,
ALERT_WEBHOOK_URL), T23 (CUE ruling record), T24 (verify/re-scope after
the concurrent claim), T25 (hardening-parity pin), T26 (BuildFlow
bandit issue filing), T27 (flightrecorderhealth demand check), T01–T03
(stack redeploy + prod verification windows — owner-gated), final
report-back rendering both plan tables.

## d) TOTALLY FUCKED UP (honest ledger)

1. **Three pushes rode RED CI runs** (`e404ac74`, `2e2cba46`,
   `7601f37f` concluded `failure`). Diagnosis at 02:10: the blamed tests
   (`TestErrorCodeRegistryIsFresh`, `TestEveryReferencedKeyExists`,
   `TestPackageDepthBudget`) ALL PASS at current HEAD locally; the
   blame's newest-touching commits are concurrent-session trains
   (`6fbd26dc` history/lens, `0849037c`) — the documented 2026-11-12
   class (a session lands codes/keys/depth without the regen; whoever
   lands next heals it). NOT my regression — but I pushed without
   watching each run to green, the exact discipline this session vowed
   to keep. Mitigation found after the fact, not before.
2. **The TODO_LIST.md transient-empty incident**: during the daemon
   index.lock race, the tree's TODO_LIST.md was committed EMPTY by the
   concurrent history-train commit (`752dd1cc~1` had 0 bytes); my
   strike commit restored the last-known-good 1764-line content. If
   their session had staged rows in that window, those rows are gone
   from history (recoverable from their session context; they will
   re-land via fix-forward). I did not verify against their session.
3. **The first bite-proof of the single-flight pin was inconclusive**
   (the daemon had swept my fix into HEAD before the old-code swap —
   both sides fixed). Caught by noticing HEAD already contained
   "missed it"; settled with the isolated probe instead. Lesson: in
   daemon-raced repos, verify WHICH BYTES you are testing, not just
   which ref.
4. **The background CI watcher died silently** ("background shell not
   found" at 02:11) — I armed it once and never polled it, so the red
   runs surfaced an hour late via the owner-mandated status check
   instead of the watcher.

## e) WHAT WE SHOULD IMPROVE

- **Per-sha CI watching must be a foreground step** in the commit call
  chain: push → poll `scripts/ci-verdict.sh <sha>` until concluded →
  only then next task. Background watchers die; habits don't.
- **Before any commit of a SHARED file** (TODO_LIST, CHANGELOG, i18n)
  during daemon activity: `git diff --cached --stat` sanity (byte count
  plausible?) — the empty-restore would have been caught by a 0→1764
  insertion count check.
- **The 52 BLOCKED reports are the real cadence debt**: the predrain
  sweep names their harvest work; a dedicated annotate/harvest train
  would drop the pin pressure for weeks, not per-session.
- The predrain WARN lane only helps locally (CI SKIPs it, by design) —
  consider a weekly local ritual line in the docs-gates header.

## f) Up to 50 things we should get done next

1. **T19**: android repo — adopt the typed `lines[{extension,did}]`
   session wire; UI labels "DID · extension" (fixtures already
   regenerated webphone-side).
2. **T19**: update `WireContractTest` fixtures; run the Gradle suite.
3. **T19**: cross-check the deleted parallel `numbers`/`extensions`
   lists are gone client-side.
4. **T20**: check wire-fixture triggers; execute only FIRED follow-ups.
5. **T21**: stack fixture boot SSO-only shape; drive login; assert
   whoami resolves the identity; fix if red.
6. **T22**: pma push-heuristic design (the stall anchor exists in
   `push-stall-watch.sh`).
7. **T22**: hub runner durable fix per the fleet decision record.
8. **T22**: ALERT_WEBHOOK_URL channel decision + wiring.
9. **T23**: CUE ruling one-liner into AGENTS/README with the pointer
   (cheap; do not lose it behind bigger trains).
10. **T24**: verify the concurrent session's notes-in-search landing;
    re-scope or strike the row honestly.
11. **T25**: hardening-parity pin test per the 00:51 spec.
12. **T26**: file the BuildFlow bandit-provider issue (github-voice +
    verify-before-filing gates).
13. **T27**: flightrecorderhealth demand check (health-batch failures
    wanting traces?) → adopt or record the gate.
14. **T01**: stack redeploy window — bump the stack's webphone lock,
    build, switch, `/version` verify (OWNER-GATED: deploy authority).
15. **T02**: prod authenticated capture + ring-guard positive path
    (needs a ringing device).
16. **T03**: night-run stability pair (`ui-night-pair.sh`).
17. **Heal**: watch the pending CI runs on `752dd1cc`/`3f9665db`/`6fbd26dc`
    to conclusion; if red, attribute + fix-forward.
18. **Drain**: run the predrain sweep's harvest pass over the oldest
    BLOCKED reports (the 2026-11-08 four carry 0/N-rows-struck tables).
19. **Final**: report-back rendering BOTH plan tables (§2 T01–T27, §3
    m01–m27) with outcomes.
20. **Investigate**: whether the concurrent TODO_LIST rows were lost in
    the empty-file window (question g-3).

(20 ranked rows — the honest count; padding to 50 would be noise.)

## g) Questions the owner must answer (cannot be self-derived)

1. **Deploy authority + v2.15.0 order (blocking T01–T03):** may this
   session run the full release dance (tag v2.15.0 via release.sh →
   stack relock → `nixos-rebuild switch` on the PBX host → prod
   verification windows) UNATTENDED, or is a supervised window required?
   Production serves real calls; the answer decides whether T01–T03 run
   tonight or wait for you at the keyboard.
2. **Android repo scope (blocking T19/T20):** is `~/projects/webphone-android`
   in scope for THIS session's train (I would branch + land the `lines[]`
   adoption + run its Gradle suite), or does it get its own dedicated
   window?
3. **The TODO_LIST restore:** my `752dd1cc` restored the file from
   last-known-good after a daemon race committed it EMPTY — I cannot
   know what rows the concurrent history-train session had staged in
   that window. Should I treat any rows they re-add as canonical
   (default, fix-forward), or do you want a reconciliation pass against
   their session notes first?
