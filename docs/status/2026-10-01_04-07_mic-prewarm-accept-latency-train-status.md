# Status: Mic pre-warm train — "speak ASAP after accept"

**When:** 2026-10-01 04:07 CEST · **Scope of this report:** everything since `2026-10-01_02-54_prod-call-demo-webtransport-qa-and-island-latency-diagnosis.md` — the accept-latency answer turned into a shipped island feature.
**Commits:** `c349950` + `4266b8d` (code + tests, auto-daemon), `e26d1aa` (AGENTS.md + CHANGELOG.md doc touches). Working tree clean at write time.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Evidence                                                                                           |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| a1 | **Vendored-API research before design.** Read the minified sip.js 0.21.2 bundle: `SIP.Web.defaultSessionDescriptionHandlerFactory(mediaStreamFactory?)` accepts a custom factory (default = `getUserMedia(constraints)`); the SDH calls `mediaStreamFactory(constraints, session, options)`; **`iceGatheringTimeout` default is 5000 ms** (not the 500 I half-remembered — the bundle corrected me); oxlint `env.browser` covers `navigator`/`MediaStream`.                  | rg dumps of `internal/web/assets/vendor/sip.min.js` (`qt(o)`, `Pt()`, `5e3`), `island/oxlint.json` |
| a2 | **`mic.js` shipped** — the pre-warm state machine: `warmMic()` (generation-token guard, idempotent), `takeWarmMic()` (one-shot handoff + pending-takeover so a late acquisition is STOPPED, not cached — prevents an unowned live mic), `releaseWarmMic()`, `micMediaStreamFactory` (mirrors upstream default semantics: empty-constraints → empty `MediaStream`, insecure-context rejection with the same message). Leaf module, zero imports → island graph stays acyclic. | `internal/web/assets/island/app/mic.js`; arch test green                                           |
| a3 | **Wiring shipped:** `connection.js` — custom factory + `iceGatheringTimeout: 1000`, `warmMic()` in `onInvite`, `releaseWarmMic()` on the missed (caller-gave-up) branch and in `disconnect()`; `calls.js` — release on `rejectIncoming()` and on inbound Terminated in `bindSession` (covers accepted-call-died-in-setup).                                                                                                                                                   | diffs in `c349950`/`4266b8d`                                                                       |
| a4 | **Test suite grown 83 → 98, all green:** new `mic.test.mjs` (12 tests: handoff-once, pending takeover ×2, release ×2, ended-track invalidation, factory warm/video/empty/insecure/no-devices), `connection.test.mjs` +3 (factory wiring + gather cap assertion, warm-during-ring, release-on-missed). Full island suite 98/98.                                                                                                                                               | node:test run, this session                                                                        |
| a5 | **All targeted gates green:** `go test ./internal/arch/` (module graph), `./internal/server/` (DOM contract + CSP untouched and pinned), oxlint exit 0, treefmt/prettier "0 changed" on app files, oxfmt "correct format" on the two `.mjs` test files.                                                                                                                                                                                                                      | command outputs this session                                                                       |
| a6 | **Docs landed per repo convention:** AGENTS.md sip.js bullet extended with the mic pre-warm seam (including the 5000 ms-default discovery and the takeover rule), CHANGELOG `Unreleased → Added` entry written.                                                                                                                                                                                                                                                              | `e26d1aa`                                                                                          |

## b) PARTIALLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                      | Works                    | Open                                                                                           | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ | ---------------------------------------------------------------------------------------------- | ------ |
| b1 | **The feature itself — code-complete, NOT live-verified.** Unit + contract gates prove the wiring; zero real calls have been placed through the changed island. No smoke binary booted, no stack browser E2E re-run (the E2E exercises registration + real WebRTC media and is the closest thing to a live check we own). | 98 island tests          | one real incoming call (accept→speak timing, mic indicator at ring) + stack E2E + a smoke boot | M      |
| b2 | **BuildFlow gate not run.** Targeted gates only; the repo's ONE quality gate never executed on this train. Nothing indicates breakage, but "green" here is my claim, not the gate's.                                                                                                                                      | —                        | run `buildflow`                                                                                | S      |
| b3 | **Ring-silence AudioContext bug (carried from the 02:54 report, still unfixed).** Diagnosed (`audio.js:10,37` create contexts outside a gesture → suspended → ring tone possibly silent). The mic pre-warm does NOT touch this — it is about capture, not playback. Awaiting owner decision.                              | diagnosis + proposed fix | implementation + island test                                                                   | S      |
| b4 | **Field data for the latency budget.** Still no ICE panel `path:`/`rtt:` numbers from a real call and no headset type known — every remaining latency claim is theory.                                                                                                                                                    | in-product panel exists  | one call's panel screenshot/numbers                                                            | S      |

## c) NOT STARTED

| #  | Item                                                                                                                                                                                                                                                       | Why                                                                                               | Wanted?        |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | -------------- |
| c1 | HARVEST of BOTH reports' (f) sections into `TODO_LIST.md`/`ROADMAP.md`                                                                                                                                                                                     | last report ended "wait for instructions"; this one same — needs owner word (docs-health HARVEST) | Yes            |
| c2 | `devicechange` re-warm (BT headset connects after warm → stale stream bound to old device)                                                                                                                                                                 | edge case, not asked for                                                                          | Medium         |
| c3 | Mic-permission-denied operator visibility (warm failure is silent by design; accept-time fallback produces the error — no early signal)                                                                                                                    | design choice, revisit if it bites                                                                | Medium         |
| c4 | Cross-browser check: Safari/Firefox behavior for gesture-less `getUserMedia` at ring time (Chrome verified by policy docs only)                                                                                                                            | no non-Chromium device in the loop today                                                          | Medium         |
| c5 | Answer-latency `#log` timestamps + ice-panel gathering-duration/time-to-media (carried items #7/#8 from the 02:54 report)                                                                                                                                  | not started, still valid                                                                          | Yes            |
| c6 | WebTransport verdict doc, deploy.md secret-path column, ops-runbook recipe + password path, MOH audibility check, `/recordings/` check, `/tmp/song.wav` cleanup, Option B teaser verification (all carried from the 02:54 report — untouched this segment) | out of scope for the mic train                                                                    | Yes (mostly S) |

## d) TOTALLY FUCKED UP

Honest accounting — all test-authoring, zero product-code bugs found after the fact:

1. **First `mic.test.mjs` draft was incoherent.** It asserted across `?case=` fresh module instances (each load = SEPARATE module state, so cross-instance assertions tested nothing) and contained a labeled-`break` junk block. Self-caught BEFORE the first run and rewritten — but it should never have been drafted that way. Severity: wasted effort, not a false green.
2. **Three failed test runs to green, all my test bugs:** (a) `gum.fn = () => new Promise(() => {})` replaced the recording wrapper, so the assertion counted 0 by construction; (b) test isolation — the prior connection test left `state.incomingSession` set, so the next `onInvite` hit the second-call guard and never warmed (0 ≠ 1); (c) connection.test's local `StubEmitter` has only `addListener` — I assumed the `fire()` from the SIBLING file's emitter and got a TypeError. Root cause across all three: I wrote tests against the stubs I IMAGINED, not the ones in the file I was extending. The product code itself passed on first execution.
3. **`go test ./internal/web/` — "no Go files".** Wrong package path (the contract tests live in `internal/server`); recovered in one step. Minor, same class: assert the map before driving.
4. **Edit-tool rejection on AGENTS.md** ("read before editing" — my earlier in-session read had gone stale) and a formatter/daemon race on `connection.test.mjs` (modified between view and edit; re-read, re-applied). Both handled correctly, both cost a round trip. The repo's concurrent-sessions rule (re-read shared files immediately before editing) exists precisely for this — I followed it reactively, not proactively.
5. **"Done" was declared without the primary gate.** I reported the feature complete on targeted gates alone (see b2). The repo's own bar is buildflow; a strict reading says the train is NOT done until it runs.

## e) WHAT WE SHOULD IMPROVE

1. **Inventory the test file's OWN stubs before extending it.** The `fire()`/`StubEmitter` bug and the isolation bug both came from pattern-matching a sibling file. Rule: when extending a test file, list its stub classes and its module-level state FIRST (they differ per file by design).
2. **Write state-machine tests with the fresh-instance + explicit-reset pattern from line one.** Every connection-family test needs `resetStubs()` + the shared-state resets it touches (`state.incomingSession`, mic cache). Retrofitting isolation after a failure costs more than starting with it.
3. **Read the vendored dependency FIRST when the design hinges on its API** — this train got that right (the bundle decided the whole design and even corrected the timeout folklore). Keep doing exactly that.
4. **Run the primary gate before calling a train done, or label the report "gate pending."** Partial-gate green is not the repo's green.
5. **Narrative commits per phase boundary.** The daemon split this train across `c349950`/`4266b8d` mid-flight (tests separated from code by an unrelated docs commit) — the repo convention asks for explicit narrative commits at phase boundaries; the daemon's heuristic grouping is the fallback, not the plan.
6. **Carry a "measured vs theorized" tag on latency claims** until the ICE panel numbers exist (b4) — theory shipped twice now; measurement keeps not happening because it needs a human on a call.

## f) Up to 50 things to get done next

> Items marked ⟲ are carried open items from the 02:54 report. 1–10 are TODO_LIST-ready; the rest are ROADMAP/cleanup fuel for HARVEST. Impact / Effort / Category.

1. **Place one real incoming call through the PBX** — verify accept→speak is sub-second, mic indicator lights at ring, warm release on reject. High / S / Quality
2. **Re-run the stack browser E2E** (island call path changed; E2E proves registration + real media end-to-end). High / M / Quality
3. **Run the full buildflow gate** on this train. High / S / Quality
4. **Fix the ring-silence AudioContext bug** (b3 — create/resume inside a gesture; island test pins ctx-creation order). High / S / Bug
5. **Capture ICE panel `path:`/`rtt:` from a real call** and close the theory gap (b4). High / S / Quality
6. **Boot a smoke binary** (`go build` + `scripts/webphone-smoke.py`) over the changed island. Medium / S / Quality
7. **HARVEST both status reports** into `TODO_LIST.md`/`ROADMAP.md` (c1). High / S / Documentation
8. **Answer-latency `#log` lines** (click → 200-sent → Established; operator-greppable English). Medium / M / Feature
9. **Ice panel: gathering duration + time-to-first-media + a "gather cap hit" hint.** Medium / M / Feature
10. **`devicechange` re-warm** (BT plugs in after warm → refresh the stream). Medium / M / Feature
11. ⟲ MOH audibility check on prod + `local_stream` inventory (still unconfirmed you HEARD music). Medium / S / Quality
12. ⟲ Check the demo call landed in `/recordings/` + CDR. Medium / S / Quality
13. ⟲ Write the demo-call recipe (originate + corrected password path) into the telephony ops-runbook. High / S / Documentation
14. ⟲ Add a PATH column next to secret names in telephony `deploy.md`. Medium / S / Documentation
15. ⟲ Verify Option B (teaser file playback on prod, DynamicUser read through the symlink). Low / S / Quality
16. ⟲ Verify the `sofia/gateway/itsp/+<num>` mobile variant (owner-approved spend). Low / S / Quality
17. ⟲ Persist the WebTransport verdict as a planning doc. Medium / S / Documentation
18. ⟲ `/tmp/song.wav` cleanup + decide Option A vs B as the standing demo. Low / S / Cleanup
19. Test: warm mic survives an agent rebuild (module-level state is intentional — pin it). Low / S / Quality
20. Test: outgoing dial after a missed ring consumes the stale warm stream (intended reuse — pin it). Low / S / Quality
21. Test: second `onInvite` while a first call rings (second-call guard keeps ONE warm, rejects the invite — pin it). Low / S / Quality
22. Dedupe test helpers: move `makeTrack`/`makeStream`/mediaDevices stub into `helpers.mjs` (now duplicated in mic.test + connection.test). Low / S / Quality
23. `pagehide` release (mid-ring navigation keeps the warm mic until unload). Low / S / Feature
24. Mic warm failure: one `announce()` at ring time instead of full silence (permission-denied visibility). Low / S / Feature
25. Warm-on-login option (indicator tradeoff; owner decision). Low / S / Feature
26. Cross-browser: Safari/Firefox gesture-less `getUserMedia` at ring (c4). Medium / M / Quality
27. `iceServers` trimming evaluation (fewer allocations, shorter gather). Medium / M / Quality
28. Trickle-ICE feasibility scan for a future sip.js (named-trigger policy). Low / L / Feature
29. Share ONE AudioContext between ringback and ring tone. Medium / S / Refactor
30. README FAQ: "why answering is fast now / what the mic indicator at ring means". Low / S / Documentation
31. FEATURES.md island row: add the mic pre-warm seam. Low / S / Documentation
32. Fold the CHANGELOG entry into the next release train's version section. Low / S / Documentation (comes with release)
33. Verify `e26d1aa`/`4266b8d`/`c349950` are PUSHED (`git ls-remote`, per repo convention). Low / S / Ops
34. Annotate the 02:54 report's open items that this train closed (#9 pre-warm done; #1 autoplay still open) — docs-health ANNOTATE. Low / S / Documentation
35. pbx-artmann runbook note: island warms at ring (no PBX-side action, but the next debug session should expect the mic live before ACCEPT). Low / S / Documentation
36. Battery impact sanity: warm mic keeps the audio route active during long rings on mobile. Low / L / Quality (ROADMAP)
37. Make the gather cap observable: log when `iceGatheringTimeout` actually bites. Low / S / Feature
38. Decide 1000 ms vs PBX_CONFIG-configurable gather cap. Low / S / Decision
39. Operator `#log` line on warm handoff ("mic warm → call") for support correlation. Low / S / Feature
40. Consider releasing warm on `visibilitychange→hidden` during a very long ring. Low / S / Feature (YAGNI flag)
41. ⟲ AGENTS.md (webphone): "media is P2P SRTP; WS is signaling-only" one-liner if absent. Low / S / Documentation
42. ⟲ Ice-hints copy: explain slow answers in plain language (relay → slow path). Low / S / Feature
43. ⟲ `sofia status` / runbook: originate caller-ID `0000000000` — document or set from-user for demos. Low / S / Quality
44. ⟲ MOH originate timeout knob documented next to the recipe. Low / S / Documentation
45. ⟲ `pbx-fs <cmd>` wrapper on prod (shorter handover commands). Low / S / Feature
46. ⟲ Telephony VM test: `local_stream://moh` wired under the default sounds package. Medium / M / Quality
47. ⟲ Telephony VM test: originate + `&playback` smoke. Medium / M / Quality
48. ⟲ lessons.md candidate (crush-config): "derive runtime paths from config values, not docs tables" (from the 02:54 train, still unrecorded). Medium / S / Lesson
49. Shorten the AGENTS.md mic bullet after the next release (it is long; fold the how into docs, keep the rules). Low / S / Documentation
50. Re-run the whole verification set (1+2+3+6) as one clean record once the autoplay fix (4) also lands. Medium / S / Quality

## g) Questions I can NOT answer myself (answer to unblock)

1. **Place a real call now (the island changed — hard-refresh the tab first): how long from tapping Accept until you can actually speak, and did the mic indicator light up WHILE it was ringing?** This is the only evidence that the train did what it claims; everything else is unit-level truth.
2. **Was your headset Bluetooth during the original slow call, and what did the ICE panel show (`path:` host/srflx/relay + `rtt:`)?** Still unanswered from last report; it decides whether the remaining latency knob is gather-cap, TURN path, or device warm-up.
3. **Green-light the ring-silence AudioContext fix (#4) and authorize the gates (#2 stack E2E, #3 buildflow) — and do you want HARVEST (#7) to fold both reports into `TODO_LIST.md` now?** All three are one-word decisions that unblock the next train.

---

_Point-in-time snapshot. The mic pre-warm train is code-complete and gate-green at the targeted level; live verification, the primary buildflow gate, and the carried autoplay bug are the open threads. Markdown per explicit user instruction (skill default is HTML — override honored, flagged in the closing message)._
