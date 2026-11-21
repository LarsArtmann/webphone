# Pareto T24–T27 + T21/T19 session status — 2026-11-21 03:19

> **Snapshot-rot warning:** push/CI statements rot fastest. CI was last
> checked 03:19 CET via `scripts/ci-verdict.sh` (see §d-1 — the verdict
> for the resumed-against sha is genuinely UNKNOWN). Webphone origin/main
> was at `560edfa2` (concurrent session's docs commits on top of my
> `8402b3d3`); stack main at `53ab2cc`.

**One-line verdict:** the owner's second "WHOLE TODO LIST, DO NOT STOP"
directive resumed the pareto plan; four tasks closed (T24 verified as
concurrent-landed, T26 FILED upstream as BuildFlow#113, T27 gate verdict
recorded, plus the session-start CI check), T21 is written-and-RED with
a diagnosed fixture gap (one settings line from green), T19's research
is complete with the wire drift proven red by the sync gate — and the
session STOPS here per the owner's status-report-then-wait order.

## a) FULLY DONE

1. **Session-start mandate**: check the CI verdict for `0f5690dd`.
   Answer: inconclusive-by-instrument (§d-1) — no run recorded for the
   sha the watcher sampled, and origin/main has since moved three
   concurrent-session commits forward (`10e88574`, `560edfa2`,
   `92174cb9`). Local gates for everything I pushed were run or verified
   in-session; no red attributed to my commits.
2. **T24 (notes-in-search) — VERIFIED as concurrent-landed, not
   double-claimed**: the feature lives at HEAD
   (`NotesGate.SearchBodies` + the panels.go note-only join); the pin
   `TestHistoryQuerySurfacesNoteOnlyCalls` re-run green this session
   (`--- PASS`, verbose); the R25 row was already struck by the
   concurrent session with full evidence. Nothing owed.
3. **T26 (BuildFlow bandit-provider issue) — FILED and verified
   landed**: `LarsArtmann/BuildFlow#113`, OPEN, body intact (3,274
   bytes). Process: all five verify-before-filing gates passed — the
   provider re-verified BARE at BuildFlow `ceb25caa1`
   (python_tools.go:45, output_parser.go:111-120 mechanism quoted),
   fresh reproduction at webphone HEAD (339 furniture findings sharing
   ONE id, tree itself 0-real/63-nosec), no existing issue (only closed
   #64, the statix class, which does not cover bandit); the body passed
   the github-voice checker (`--kind body-issue --ai-drafted
   --unsolicited`, 0 FAIL 0 WARN) and ships the provenance banner
   UNCHECKED for owner review (the appkit-asks precedent). Both webphone
   TODO rows struck with evidence (the tool-disposition row + the
   gate-setup residue's "owner-tick gated" row — its stricter reading is
   superseded by the banner precedent the filing follows). Final text:
   `docs/drafts/2026-11-21_bandit-issue-final.md`. Commit `e343d849`,
   pushed, origin verified.
4. **T27 (flightrecorderhealth demand check) — verdict: gate HOLDS,
   recorded in the row** (commit `8402b3d3`, daemon-swept + pushed):
   (a) zero webphone health-batch wedge sightings in docs/status +
   docs/lessons since the gate armed 2026-11-12 (the corpus's "wedge"
   stories are CI-runner and dev-machine go-get classes); (b) the
   request-trigger arm is STRONGER than the row assumed wherever the
   stack's monitoring is on — `telephony-health` curls webphone
   `/healthz` per tick (monitoring.nix:97), so a failing leg makes that
   poll itself a 5xx/slow traced request and the snapshot fires WITHOUT
   the module. Remaining gap: the narrow "background loop dies while
   every HTTP answer stays fast-200" class; reopen per the registry
   trigger.
5. **Concurrent-train coordination survey** (protects the plan from
   double-fire): the 02:28 whole-list session is ALIVE and executing
   their queue (three docs commits landed during this session: the
   resume report, the T25 owner-pack addendum §G with 8 rulings, the
   error-contract detection column). Their declared queue owns the
   deploy chain (their items 1–6 = my T01–T03), four dependency bump
   trains, the stack lens option, android `lines[]` (their item 12),
   and the pma/hub fleet legs (their items 15–16). Division observed:
   this session kept to T24/T26/T27/T21/T19 (T21's stack surface is
   tests-only; their lens leg is modules/ + eval needles — non-overlapping
   files).

## b) PARTIALLY DONE

1. **T21 (SSO-only identity resolution in the stack fixture)** — the
   test is WRITTEN, registered, eval-green, and the first VM run is RED
   with a DIAGNOSED gap:
   - `tests/webphone-sso-only.nix` (stack repo): boots the full fixture
     (FreeSWITCH + nginx TLS front + webphone) with `oauth2.enable`,
     passkey OFF, the shared users map (`lars@pbx.test` → 1000), the
     password-file override; a loopback **https stub IdP on :8443**
     (cert adopted into the guest trust store, hosts entry) serving
     discovery + authorize + token + JWKS and minting REAL RS256 ID
     tokens via python-cryptography (nonce stashed per code, iss/aud/
     exp/iat/email claims); the dance driven by curl exactly like the
     app-level pin (`TestSSOOnlySessionResolvesIdentity`): begin →
     discovered-authorize 302 (PKCE S256 + nonce asserted) → stub
     authorize → callback 303 → `GET /api/session` resolves
     `"extension":"1000"` + `"display_name":"Lars"` + `"lines"`; plus a
     negative anonymous-401 arm.
   - Registered as `telephony-webphone-sso-only` in the stack flake's
     plainSuites; `nix eval` green; full VM build run.
   - **RED root cause (diagnosed from the boot log)**: `auth.passkey:
     rp_id is required when the passkey mode is configured` — the
     shared identity layer (the users map) configures passkey-mode
     derivation and webphone requires rp keys even with the ceremony
     OFF. Webphone's OWN oauth2 VM test documents exactly this and pins
     `settings.auth.passkey.rp_id/rp_origins` explicitly; the stack's
     eval-only fixture escapes because its host module derives rp from
     the domain. **Fix is one raw-settings rp pin in the fixture**
     (mirroring webphone's own test), then re-run.
   - The daemon auto-committed the test + registration to stack main
     (`53ab2cc`) BEFORE the fixture was green — a landed-red on the
     stack's flake-check surface; fix-forward is the next action.
2. **T19 (android `lines[]` wire)** — research complete, implementation
   not started:
   - Drift PROVEN: `scripts/sync-wire-fixtures.sh --check` RED — the
     vendored `session-identity-brokered.json` still carries the old
     parallel `numbers`/`extensions` lists; the server fixture ships
     `lines:[{extension,did},…]` (the 2026-11-18 typed wire).
   - The app's wire model read in full: `SessionApi.SessionIdentity`
     parses only `extension` + the credential arm today (the old
     parallel lists were never consumed client-side — there is nothing
     to DELETE, only to ADOPT); the label surface is the
     `line_and_wake` pill pair in MainActivity (format string
     `%1$s · %2$s` already matches the "DID · extension" separator
     idiom); `CallLayer` registers SIP on the ACTIVE line only.
   - The suite home (`WireContractTest`), the gate
     (`scripts/verify-gate.sh`: assembleDebug → testDebugUnitTest →
     ktlint → zipalign → abi → contrast → dprint, Gradle-busy-guarded),
     and the android repo's TODO (no mirror row yet — owed when picked
     up) all surveyed.

## c) NOT STARTED

1. **T20** (android wire-fixture trigger follow-ups — sequenced after
   T19 lands).
2. **T22** (fleet health batch): pma push heuristic is a CROSS-REPO pma
   train (LIVE daemon; age gate + canary + rollback story — mirror row
   in pma's TODO_LIST owed); hub runner durable fix lives in the hub
   host repo's own TODO; `ALERT_WEBHOOK_URL` is an OWNER channel pick
   (their queue item 39) — not agent-convertible.
3. **T01–T03** (stack redeploy + prod verification windows): the
   concurrent session's queue explicitly owns these (their items 1–6);
   v2.15.0 is released (`2be14b73`) and their vm-tests post-split fix
   (`81b52f3b`) suggests they are mid-deploy-chain. This session did
   NOT touch the deploy (double-firing a prod switch is the worst
   collision class).
4. **Final report-back** rendering both plan tables (§2 T01–T27 and
   §3 m01–m27) — owed when the list is done.
5. Watching the concurrent v2.15.0 train's end state (tag/gh-release
   verify) — the harvest row exists; not re-checked this session.

## d) TOTALLY FUCKED UP

1. **The CI watcher exited on a transient state (instrument flaw #3)**:
   my background poll loop treated ANY non-pending verdict as
   conclusive, but it sampled `origin/main` in the window after a
   concurrent push and BEFORE CI registered a run — `ci-verdict.sh`
   answered exit 1 ("no runs for this sha") and the loop STOPPED. The
   verdict for `0f5690dd` (and the current tip) is genuinely UNKNOWN.
   Fix owed: exit-1 must re-poll inside a grace window; only a
   workflow-concluded green/red stops the watcher.
2. **T21 first VM run was avoidably red**: 25 minutes of build time
   taught what webphone's own oauth2 VM test comment already warned
   (the identity layer requires rp keys with the ceremony off). I
   re-derived the whole stub-IdP machinery from first principles
   WITHOUT grepping the sibling fixture's failure-prevention comments
   first — the rp pin was one `git grep` away.
3. **Landed-red on the stack main** (the daemon raced the fixture fix):
   `53ab2cc` carries the red test + its registration. Consequence: a
   stack `nix flake check` from that tree would fail the new suite
   until the one-line fix lands. This is the documented daemon-race
   class, now on a second repo.

## e) WHAT WE SHOULD IMPROVE

1. **Watcher resilience**: the CI poll loop needs the exit-1 grace
   window (§d-1) and a heartbeat line per poll (the prior sessions'
   silent-death class). Worth 10 minutes as a script fix.
2. **Read sibling fixtures' comments BEFORE building VM tests** — the
   fixtures encode paid-for lessons (the rp pin, the cred-mode 0440
   pin, the projection-restart race). A `grep -n "must\|required\|pin"
   <sibling-test>` pass is cheaper than any red VM run.
3. **Start long builds earlier**: the T21 VM build (~25 min) could have
   run during the T26/T27 reading and filing instead of after.
4. **Cross-repo daemon awareness**: the stack repo's daemon also
   auto-commits; any red-by-construction state there must be fixed
   forward FAST or the landed-red window grows.

## f) Next actions (the pareto tail this session owes)

| #  | Item                                                                                                                                                                                                                                             | Owner       | Pri  |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------- | ---- |
| 1  | T21 fix-forward: pin `settings.auth.passkey.rp_id/rp_origins` in the sso-only fixture, re-run `nix build -L .#checks.x86_64-linux.telephony-webphone-sso-only`, green → stack CHANGELOG + the webphone TODO §f-6 strike                          | agent       | High |
| 2  | T19 implement: re-vendor fixtures (sync copy), `Line(extension, did?)` + label helper in SessionApi, WireContractTest pins (lines parse + old parallel lists asserted GONE), `scripts/verify-gate.sh` green, android CHANGELOG + mirror TODO row | agent       | High |
| 3  | T20: check the android wire-fixture triggers; execute only FIRED follow-ups                                                                                                                                                                      | agent       | Med  |
| 4  | T22 pma leg: mirror row into pma's TODO_LIST (the heuristic itself is a careful LIVE-daemon train)                                                                                                                                               | agent       | Med  |
| 5  | T22 hub leg: mirror/verify the hub runner durable-fix row in the hub repo                                                                                                                                                                        | agent       | Med  |
| 6  | T01–T03: coordinate with the concurrent session's deploy chain (take over ONLY if their train stalls)                                                                                                                                            | agent+owner | High |
| 7  | CI-watcher fix (exit-1 grace window + heartbeat)                                                                                                                                                                                                 | agent       | Med  |
| 8  | Final report-back: render BOTH plan tables (T01–T27 + m01–m27) with outcomes                                                                                                                                                                     | agent       | High |
| 9  | Verify the concurrent v2.15.0 train's end state (tag cut, gh release) per the harvest row                                                                                                                                                        | agent       | Low  |
| 10 | Answer §g-3 and reconcile TODO_LIST canonicity once                                                                                                                                                                                              | owner       | Med  |

## g) Questions for the owner

1. **Deploy authority (T01–T03)**: the concurrent whole-list session's
   queue owns the stack redeploy + both prod windows and is actively
   landing commits. Should I (a) let their train own the deploy and
   verify their end state afterward, or (b) take the deploy myself now
   (with the ssh-owner-only harvest row in mind — is "WHOLE TODO LIST"
   the standing authorization for `nixos-rebuild switch` on
   pbx.artmann)? I cannot determine whether two sessions may share a
   prod switch window, and a double-fire is the one collision that
   costs uptime.
2. **Android line-picker scope (T19)**: the native shell is
   single-active-line by design (one SIP registration; a second call
   waits). Is the wanted v1 (a) the wire adoption + "DID · extension"
   LABELS wherever identity renders (parse, display, fixtures,
   contract pins), or (b) a full line SWITCHER (POST
   /api/session/switch + SIP re-register mid-shell)? The plan's m19.2
   says wire + labels; the webphone row's "render its line picker"
   reads like (b) — a product call I cannot make from the code.
3. **TODO_LIST canonicity (standing g-3 from the 02-15 report, still
   unanswered)**: a concurrent commit deleted TODO_LIST.md wholesale
   (`f251aa33`, later restored by daemon sweeps). Treat the CURRENT
   file as canonical (my working assumption — I struck rows against
   it), or do you want a reconciliation pass against the pre-delete
   revision first?

— drafted by GLM-5.3 via Crush, 2026-11-21 03:19 CET
