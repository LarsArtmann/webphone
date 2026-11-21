# Pareto T07–T09 journey rework status — 2026-11-20 11:17

> ARCHIVED 2026-11-21 (docs-health v10 sweep): pre-drain verified (check-rows --strict + bare-rows green), drained for cadence — the pin sat at 66/66 with 12 gates-passing candidates

> Snapshot-rot header: claims verified as of 11:17 CET. `origin/main`
> (5b75b8b6) is RED in the `workflow e2e (browser)` job — the daemon
> swept this train's half-reworked journey file and the T06 promotion
> made that job blocking; the local rework (4e74d606, 71acd93a,
> unpushed at snapshot) carries the fix, still one 3-line patch short of
> a green local run. This report covers THIS session's run only:
> resuming `docs/planning/2026-11-19_19-46_SUPERB-pareto-backlog-execution-plan.md`
> (T01–T27) per the owner's "whole TODO list" directive.

## a) FULLY DONE (verified this session)

1. **Docs-gates battery: FULLY GREEN on arrival.** The resume note's
   three reds (freshness / audit-all 74>72 / cadence 68>65) were healed
   BETWEEN sessions by the concurrent session's sweeps — verified, not
   assumed: newest-report freshness FRESH (0 unharvested), audit-all 70
   ≤ 72, cadence 66 ≤ 66, full `scripts/docs-gates.sh --check` PASS.
2. **Manual-push bar invoked with its anchor**: 6 unpushed commits,
   oldest 779 min (2026-11-19T21:38:20+01:00, bar 60 min) —
   `push-stall-watch.sh` STALL verdict recorded, pushed
   d9d2a251..5b75b8b6.
3. **CI red on the pushed tip root-caused in minutes** (the sibling-§f
   discipline paid off): run 44713194391, job `workflow e2e (browser)`,
   exactly the three diagnosed failures of the known mid-rework file
   (deny-path switcher toast, admin nav absent for an extension login,
   the ElementNotInteractable crash). No new defect.
4. **T07–T09 rework implemented in full** (scripts/workflow-e2e.py,
   syntax + self-test green):
   - Port-pinned boot + `rp_origins: ["http://webphone.test:<port>"]`
     (go-webauthn compares scheme+host WITH the non-default port —
     verified in the v0.18.2 source, `isOriginEqual` on `url.Host`) +
     `--unsafely-treat-insecure-origin-as-secure` for WebAuthn's
     secure-context need on plain http.
   - `mint_enroll_token` (CLI one-shot under the same config/datadir,
     one retry) restructured to mint BEFORE boot — the smoke's
     shared-journal ordering.
   - `journey_passkey`: chromedriver virtual authenticator (CTAP2,
     resident, UV) → /enroll ceremony (token prefill = module-bound
     proof) → /login `#lp-login-form`/`#lp-email` ceremony → shell +
     whoami ("Ops User · +49 30 1110001").
   - `journey_line_switcher` rewritten for the DESIGNED success path
     (select 1002 → island POSTs + reloads → re-rendered switcher +
     whoami + `#wp-sse-live` resubscribe).
   - `journey_dial_gate`: JS value-set + `requestSubmit()` (the
     interactability fix) + the override re-allows leg.
   - `journey_voicemail`: stale-retry clicks (the boot-time re-fetch
     morph), toast-once as a PEAK-count over a 7s settle window
     (toasts live 6s, dedup consecutive), `window.confirm` override
     instead of racing `switch_to.alert`.
   - Journey isolation: `attempt()` wrapper so one crash never orphans
     the suite (the first run's lesson, applied).
5. **Three decisive probes** (probe-first debugging, each ~1 min):
   - VM audio state dump: `err:4 / ns:3` — the cross-ORIGIN (different
     port) recording URL is CSP-blocked; production recordings ride the
     same-origin `/phone-api` proxy.
   - Phone-API path shapes through the proxy: summary/messages 200 but
     history + `/vm/…` 404.
   - Upstream path capture: the app PREPENDS `/phone-api` to the
     configured `phone_api_url` upstream (both direct and proxied calls
     carry the prefix) — the fake's substring checks (summary, messages)
     accidentally matched while its prefix checks (`/vm/`, `/history`)
     never did.
6. **Two full e2e runs measured**: 24/30 → 27/29. The remaining two
   failures (mint datadir, live play) are both root-caused with fixes
   designed (see §b).

## b) PARTIALLY DONE (in-flight, committed by the daemon, unpushed)

1. **T07–T09 final 3-line fix NOT yet applied or run**:
   - `FakePhoneAPI`: strip the `/phone-api` prefix at the top of
     `do_GET`/`do_DELETE` (the config contract appends it upstream).
   - `audio_url` rows: RELATIVE `/phone-api/vm/{uuid}…` so the browser
     rides the session-gated same-origin proxy (CSP-clean, production
     shape).
   - `mint_enroll_token`: `os.makedirs(datadir, exist_ok=True)` — the
     CLI's SQLite open fails on a missing parent (the smoke makedirs;
     I missed it).
2. **CI remains RED** until the fix lands green twice locally, then
   commits + push.

## c) NOT STARTED

The T10–T27 ladder (export tool, catalog parity, renames, type tails,
comment truths, fcr batch, cadence pre-drain, stack legs, android,
SSO fixture, fleet health, CUE record, notes spike, hardening pin,
bandit issue, flightrecorderhealth) and T01–T03 (deploy + prod
verification windows, gated on §g questions). Final report-back with
both plan tables.

## d) TOTALLY FUCKED UP (honest ledger)

1. **The daemon swept the HALF-reworked journey file and CI ran it as
   a BLOCKING job** — main sat red ~2 h. T06's own promotion bit this
   train: editing a blocking-job's input across multiple turns without
   a coherent landing point guarantees a red window.
2. **I skipped a documented ordering** the resume notes and the smoke
   both stated (mint the enroll token BEFORE boot); the 503 taught it
   again. Reading the recipe before writing the code was the miss.
3. **The fake phone API was prefix-blind**: I matched the WIRE SHAPES
   (summary/messages/vm/history) but not the URL CONTRACT (the
   `/phone-api` prepend) — two of four routes silently 404'd, and the
   voicemail rows still rendered because two substring checks
   accidentally matched. A honest 404 log line in the fake would have
   caught it in run one.
4. Two edit escapes on the first pass (a `·` escape mismatch in an
   edit old_string; a dead try/except I wrote then had to simplify) —
   sloppy first-pass matching against bytes I had just viewed.

## e) WHAT WE SHOULD IMPROVE

- **Probe-first for browser-media semantics**: CSP, origin/port, and
  codec behavior diverge in headless; a 1-minute state-dump probe
  beats a 15s timeout assertion loop.
- **Fake servers should LOG unmatched paths** (a one-line WARN): this
  failure class (substring-vs-prefix accidental matches) is otherwise
  invisible while rows keep rendering.
- **Blocking-job edit discipline**: land journey reworks as one
  verified batch (or park the job's input green at every turn) — a
  red main is now a real signal, not advisory noise.
- **Read the sibling recipe before porting it**: the smoke's
  enroll/mint ordering, makedirs, and env shape were all written down.

## f) Up to 50 things to get done next (this train's tail, ordered)

1. Apply the 3-line fake/URL/makedirs fix in scripts/workflow-e2e.py.
2. Full e2e run → all-green; then a SECOND consecutive green run
   (blocking-lane stability bar).
3. Commit + push; watch `scripts/ci-verdict.sh` to GREEN on the new
   tip (heals the red window).
4. CHANGELOG `[Unreleased]` line for the three journeys; strike
   R26/R27/R28 in TODO_LIST with run evidence.
5. Clean up /tmp probes (vm-probe, hist-probe, path-probe) —
   disposable, verify nothing references them.
6. T10: export-or-confirm tool per the 2026-11-14 design (m10.1–m10.5).
7. T11: check the concurrent session's newest report FIRST for catalog
   overlap (R30 lesson).
8. T12: FindThread → ResolveThreadID (LSP rename + call sites).
9. T13: Locality constants, ASRProvider enum, LoginPageExtra params.
10. T14: apidocs response-header comment fix + favicon token
    unification.
11. T15: fcr small-findings batch (verify what the concurrent session
    already landed).
12. T16: cadence pre-drain automation (this session hit the 66-report
    boundary again — the class is live).
13. T17: stack drift audit + coverage arms.
14. T18: ringhook operator docs + typed mirror option.
15. T19/T20: android lines[] adoption + spec/trigger lanes.
16. T21: SSO-only stack fixture verify.
17. T22: fleet health batch (pma push, hub runner, ALERT_WEBHOOK_URL).
18. T23: CUE ruling one-liner + pointer.
19. T24: notes-in-search corpus spike.
20. T25: hardening-parity pin test.
21. T26: bandit-provider issue filing (verify-before-filing first).
22. T27: flightrecorderhealth demand check + gate verdict.
23. T01–T03 deploy cycle once §g answers land (stack lock bump, gates,
    pbx-artmann relock, switch, /version verify, capture, ring guard,
    night pair).
24. e2e: assert the break-glass disclosure STATE (closed) in J1.
25. FakePhoneAPI: pin its shapes against internal/pbx/client.go
    structs in a comment (drift alarm).
26. Re-run `scripts/audit-verify.sh` after the docs changes.
27. status-index regen after this report (done in the same commit).
28. Consider a `--repeat 2` flag for workflow-e2e (the twice-green bar
    as one command).

## g) Questions for the owner (cannot be answered from here)

1. **Production cutover authority (T01, standing)**: may this session
   run `nixos-rebuild switch --flake .#pbx --target-host
   root@pbx.artmann.tech` once the pre-deploy gates pass, or do you
   press that button yourself?
2. **v2.15.0 release order (standing)**: redeploy main-pin first and
   tag v2.15.0 after this train's quality work, or tag now off the
   green tip and deploy the tag?
3. **Blocking-job workflow**: `workflow e2e (browser)` now reds main
   whenever journey work lands unverified (this session: ~2 h red from
   a daemon-swept half-edit). Park in-flight journey work on a short
   branch until green, or is a red window during trains acceptable
   noise?
