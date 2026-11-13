# Android wire-contract sync — mid-flight status (SUPERB train)

**Status:** 🟡 IN FLIGHT — webphone half LANDED and CI-GREEN; android half ~15% done
**Snapshot:** claims verified as of **2026-11-10 15:57 CET** (push/CI statements rot fastest)
**Plan:** `docs/planning/2026-11-10_14-36_SUPERB-android-wire-contract-sync.md`
**Session scope:** keep `~/projects/webphone-android` in sync with this repo via
byte-pinned wire fixtures + local red-lines on both sides; fix the found-live
session-field drift.

## What this session verified (the short story)

- **LIVE DRIFT CONFIRMED:** android `SessionApi.resumeSession` reads
  `sip_domain`/`sip_password`/`websocket_url` from `GET /api/session`; the
  server's `sessionIdentity` sends NONE of those (real wire: `password` OR
  `sip_credential`; domain/websocket live in `/config.js` keys `sipDomain`/
  `websocketPath`). The shell has been SIP-registering on EMPTY credentials —
  same class the android repo's own wake-lane diagnosis (d820810, same day)
  was chasing. **Fix designed, not yet landed.**
- **Webphone fixtures LANDED:** `webphone contracts` now renders
  `docs/contracts/wire/*.json` (9 files: both transports × 3 events, both
  sessionIdentity arms, the registration request) through the SAME pure
  payload builders the senders marshal; `contracts-fresh` gate (script +
  nix check) diffs the whole tree; pinned by
  `TestWireFixtureFilesPinTheConsumerContract`. CI **GREEN** at
  `6936be2a` (run 42832784499, success 14:28Z).
- **Two of my defects red-ed CI in between** (owned + fixed, see §d): a
  wrong grep-assertion version of the pin test and an art-dupl clone group,
  both pushed by the auto-commit daemon before my verified fixes landed;
  plus a docs-format (dprint) drift from my plan-doc table. All healed in
  the green run.

## a) FULLY DONE

| Item                                                                                                            | Evidence                                                                                        |
| --------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Research: both repos' seams, DTOs, parsers, tests, freshness-gate wiring                                        | session; plan §1                                                                                |
| SUPERB plan doc with Pareto + coarse/fine tables + mermaid                                                      | docs/planning/2026-11-10_14-36_SUPERB-android-wire-contract-sync.md (daemon-committed ce9f77f2) |
| Pure payload builders extracted (fcm×2, ntfy×2) — Send/SendDraft now marshal through them                       | internal/push/fcm.go, ntfy.go                                                                   |
| Fixture renderers + CLI fold (`webphone contracts` writes wire/)                                                | internal/push/fixtures.go, internal/server/wirefixtures.go, cmd/webphone/contracts.go           |
| 9 committed fixtures, byte-deterministic (json.Deterministic — v2 map order is random)                          | docs/contracts/wire/                                                                            |
| Pin test (both credential arms, event keys on both transports, register keys)                                   | internal/server/wirefixtures_test.go                                                            |
| Freshness gate extended (script + `contracts-fresh` nix check diff wire/)                                       | scripts/contracts-fresh.sh, nix/checks.nix — `nix build .#checks…contracts-fresh` green         |
| Runbook android leg + contracts README fixture section + CHANGELOG + TODO_LIST row                              | docs/release-runbook.md, docs/contracts/README.md, CHANGELOG.md, TODO_LIST.md                   |
| Local gates on webphone half: push+server+cmd tests, art-dupl v0.7.2 (0 new clones), format-check, docs battery | session log                                                                                     |
| Train-start buildflow + docs battery repair (concurrent session's rc=20 markdown also healed)                   | buildflow history 74/74; battery PASS                                                           |
| CI green on the final tree, pushed                                                                              | run 42832784499 success                                                                         |
| Android: `WireFacts.kt` pure decoders (fromFcmData/fromNtfyFrame) written                                       | app/src/main/…/push/WireFacts.kt (daemon-committed 79f17ea, UNPUSHED)                           |

## b) PARTIALLY DONE

| Item                       | State                                                                                                                         | Remaining                                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| A2 parsers extracted       | WireFacts.kt exists                                                                                                           | FcmReceiver + NtfySubscriptionService.routeMessage still parse inline — delegate to WireFacts |
| SessionApi drift FIX       | designed (parse `password`/`sip_credential`; fetch `/config.js` `sipDomain`/`websocketPath`; origin-derived ws fallback kept) | not written; extract pure `parseSessionIdentity` + `registerPayload` for testability          |
| Android ledger             | none                                                                                                                          | CHANGELOG/TODO_LIST/AGENTS sync-ritual rows                                                   |
| Detailed narrative commits | webphone fix commit `6936be2a` only; android + earlier webphone units rode `chore:` daemon commits                            | accept (history is pushed) — narrative the android commit                                     |

## c) NOT STARTED

- Android fixture vendoring (`app/src/test/resources/wire/`) + `scripts/sync-wire-fixtures.sh` (--check mode).
- `WireContractTest.kt` (Robolectric pattern like NtfySubscriptionServiceTest): decode all 9 fixtures through real parsers; pin register-encode keys; pin both session arms incl. the exactly-one-credential rule.
- Android gates: `nix develop -c ./gradlew testDebugUnitTest ktlintCheck` (assembleDebug optionally).
- Android commit + push origin + github (both remotes lag local; nothing auto-pushes there).
- D1 deferred rows (loopback live-check script; runtime digest handshake; device-train checklist refresh).
- Final webphone buildflow tail on the landed tree (F1 verification leg).

## d) TOTALLY FUCKED UP (owned, fixed, lessons below)

1. **~30 min of red main caused by me, raced out by the daemon:** my first
   pin-test version asserted the ntfy `event` key with a plain substring
   grep (the key lives inside an escaped JSON string) — I caught+fixed it
   locally, but the daemon pushed the broken version (add9580c) between my
   write and my fix; run 42828291169 failed `go test -race`. Healed by the
   later commits; final run green.
2. **art-dupl clone group (same push):** I duplicated the marshal loop in
   push+server renderers. The repo has an explicit new-clones gate — I ran
   plain tests but NOT art-dupl before considering the unit done. Fixed by
   the one-home loop in server (`push.RenderWireFixtures` now yields payload
   VALUES); 0 new clones verified with the CI-pinned v0.7.2.
3. **Nondeterministic fixture bytes mid-train:** moving the marshal loop to
   `encoding/json/v2` silently changed FCM `data` map ordering (v2 iterates
   maps RANDOMLY — 3 renders, 3 orders). Fixed with
   `json.Deterministic(true)` (the catalog exporter's own lesson); freshness
   gate verified stable ×3.
4. **dprint-red plan doc:** my plan-doc table wasn't canonical (the rc=20
   class); local battery repaired it, but the daemon had already pushed the
   raw version, redding the docs battery job on two runs.

## e) WHAT WE SHOULD IMPROVE (process, from this run)

- **Run the repo's OWN clone gate on every Go change:** `art-dupl check`
  (v0.7.2) belongs in my per-unit verification, not "after CI red".
- **Determinism probe for any new generated artifact:** 2-3 consecutive
  renders diffed BEFORE first commit (the freshness gate caught it, but
  only after a push).
- **dprint-canonicalize every new markdown BEFORE the daemon can push it**
  (`scripts/docs-gates.sh` fmt step or `nix develop -c dprint fmt docs/…`).
- **Shrink the daemon race window:** stage+commit+verify in ONE call
  immediately when a unit goes green — the 15-min gap is what shipped the
  broken test. (Narrative message quality also dies in that window.)
- **Battery-on-commit hygiene held:** the concurrent session's in-flight
  files (reports.templ etc.) were never staged by me; their battery-repaired
  files were left for their commit — correct call, kept.

## f) NEXT (impact-sorted; IDs from the plan's fine table)

1. ~~a2.1/a2.2: delegate FcmReceiver + NtfySubscriptionService parsing to WireFacts (verbatim behavior).~~ done in webphone-android (`2c578b5`)
2. ~~a2.3: SessionApi — extract pure `parseSessionIdentity(body, config, wsOrigin)` + `registerPayload(...)`; fix credential keys + `/config.js` seam.~~ done at `2c578b5 7dd7e53`
3. ~~a1.1/a1.2: `sync-wire-fixtures.sh` + vendor the 9 fixtures into test resources.~~ done at `a9d09b1 baafd6f`
4. ~~a2.4: WireContractTest over fixtures through the real decoders (incl. brokered-arm `sip_credential`, register keys, ntfy priorities 5/3).~~ done in webphone-android (`7dd7e53`)
5. ~~a3.1-3: android CHANGELOG/TODO_LIST/AGENTS rows (sync ritual + "live contract checks unblocked locally").~~ done in webphone-android (`748e6f3`)
6. ~~a3.3: gradle gates green (testDebugUnitTest + ktlintCheck; assembleDebug if time).~~ done (docs-health pass verify-gate green on the baafd6f tree)
7. ~~f1.3: android detailed commit + push origin + github (no workflow-file touch ⇒ mirror-safe).~~ done at `baafd6f a58a68c`
8. ~~f1.2-tail: confirm webphone CI stays green through the concurrent session's pushes (attributed, not mine to fix).~~ done — run 42851723276 all-green on 722cef89 after this train healed the red it could own (format bytes, cadence archive, the tw.css CWD leak)
9. ~~Final buildflow tail (BUILDFLOW_NO_RESULT_CACHE=1) on webphone tree.~~ done — full run 80 steps with one local-load flake (TestAutoReviewDispatchRoundTrip, 3/3 green isolated, CI green on identical bytes) cleared by the failed-only rerun (rc=0)
10. ~~d1.1: deferred rows (loopback live-check script; runtime digest handshake w/ trigger; device-train checklist refresh).~~ done at `3f955f2`
11. Consider a `/config.js` key fixture (android now reads it — keys pinned today only by island boot) — small webphone addition + re-vendor. **→ deferred with trigger — TODO rows in both repos name the config.js key fixture**
12. ~~Update the android README honesty marker: session-field drift found+fixed via fixtures (replaces "caught at first device boot").~~ done in webphone-android (`a58a68c`)
13. ~~Re-run `scripts/audit-verify.sh` + docs battery after ledger edits land.~~ done (docs-health pass audit-verify 91/91 + docs battery all green after the archive sweep)
14. ~~Watch the android forgejo→github mirror lag after push (PAT/workflow-scope trap documented in their AGENTS).~~ done — both remotes pushed directly, tips verified equal (`a58a68c` is the webphone-android tip), no mirror lag possible

## g) QUESTIONS FOR THE OWNER (cannot be determined from here)

1. ~~**Is another session actively working the android wake-lane/no-ring~~ done — the wake-lane session committed its ring-or-silence plan at 16:12 (owner-gated waves, no code overlap) — this train landed standalone on a clean tree (android 2c578b5 baafd6f)
   ~~diagnosis right now** (d820810 landed this morning, status reports at~~
   ~~11:48/13:12)? My SessionApi fix touches the same seam — fold it into~~
   ~~their train or land it standalone?~~
2. ~~**Android push policy for this train:** push both remotes with local~~ done — both remotes pushed with the local verify-gate as authority (owner resume mandate + the repo standing hand-push order; Actions billing dead there)
   ~~gates as authority (Actions billing still dead), or hold until the~~
   ~~billing row closes?~~
3. ~~**The drift fix is a behavior change** (SIP REGISTER will now get a real~~ done — the Robolectric WireContractTest suite is the local authority (12 tests, meta-proven); the emulator pass stays the first device train job (TODO rows in both repos)
   ~~credential + domain). Want a quick emulator sanity pass before push, or~~
   ~~is local test green + the device train's job to verify live?~~

## Concurrent-session boundary (kept clean)

A second webphone session worked reports.templ/prod-ui-tracking/docs-health
all train. I never staged their files; their battery-repaired markdown
(AGENTS.md, error-contract.md, lessons.md, their plan) sits in the working
tree for THEIR commit. Webphone main is green as of my push; their pushes
(42831338505/42831796643) failed on MY pushed defects + format drift, healed
by 42832784499.
