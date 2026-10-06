# Status Report — 2026-10-07 00:19 CEST — round-5 plan: local-leg execution

Session: webphone, mandate = READ/execute/verify
`docs/planning/2026-10-06_23-47_SUPERB-round5-pareto-plan.md`. The plan's
22 tasks split cleanly into owner-terminal / cross-repo / live-deploy legs
(NOT runnable here) and locally-runnable legs. This session executed the
latter: T04, T05, T06, T11, T12, T21/M80, with T16 attempted and honestly
deferred. No code paths touched — docs, one scratch probe, zero
product-surface deltas.

End state at writing: local HEAD `0949667`, working tree **clean**, remote
`main` at `2c86605` (CI run `37537152790` **green**; `e7cce5f` and `81cfa25`
also green — 37537152790/37536492580/37536460912 all success). HEAD is ONE
commit ahead of remote: the first unpushed commit's timestamp is
**2026-10-07T00:19:09+02:00** — push-lag anchor stated per the manual-push
bar: 0 min < 60 min, so the daemon's push leg is still inside its normal
window and NO manual push was performed. Daemon committed this session's
work in three commits (`d711cc8` ledger+note+stories, `2c86605` plan/eval
table reflow, `0949667` AGENTS/TODO/lessons tails).

---

## a) FULLY DONE

| #  | What                                                                                                                                                                                                                                                                                                                                                     | Evidence                                                                                                                                        | Scope                       |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| 1  | **T04/r16 greppable-class byte check on the v1.20.1 tree** — built a fresh loopback binary, logged in, sent a real outbound message, byte-grepped the SERVED payloads: `wp-thread-row` present in the thread-list fragment, `wp-bubble` present in the transcript page. The templ-components v1.20.0→v1.20.1 ride does not disturb the fragment contract | Probe run output: "send status: 200 / wp-thread-row in LIST: True / wp-bubble in TRANSCRIPT: True"; `messages.templ` remains the authoring home | served-markup contract      |
| 2  | **T04/r17 DOM-contract coverage-reasoning note** — new "§ Notes" in dom-contract.md recording the probe method so future templ-components bumps can re-run it verbatim                                                                                                                                                                                   | `docs/dom-contract.md` § Notes (committed in `d711cc8`)                                                                                         | docs/dom-contract.md        |
| 3  | **T05/r2a outage-timeline lessons story** — 15:58 proxy outage → ~20:40 recovery → 22:06 declared; flock cascade mechanics, 3h44m unnoticed wedge, wchan/procfd forensics, BuildFlow misdiagnosis lesson, GOPROXY/retry-wrapper routing                                                                                                                  | `docs/lessons.md` Tooling traps (verified against the 22:08 report row 3 before writing)                                                        | docs/lessons.md             |
| 4  | **T05/r2b h1 ≠ sha256 oracle story** — `/go.mod` h1 is go's dirhash, not a file digest; the go-oracle (`go mod download -json` reproduces sumdb exactly) beats ad-hoc hashers; the GOSUMDB-off escape that the false "tampering" reading tempted was correctly never taken                                                                               | `docs/lessons.md` Nix section, cross-linked to the AGENTS vendorHash bullet                                                                     | docs/lessons.md             |
| 5  | **T06/r6 CI-flake ledger** — new ONE-home file with run-id/date/test/verdict format; seeded 37496169265 (verdict: hardening supersedes confirmation), 37524833682 (NOT a flake — deterministic vendorHash infra), 37536492580 (green); AGENTS links it, TODO_LIST cites it                                                                               | `docs/ci-flake-ledger.md`; AGENTS.md header line; run verdicts re-checked live via `gh run view/list`                                           | docs + wiring               |
| 6  | **T06/r30 run-37496169265 closure** — the ledger's verdict column records the OR-arm ("hardening supersedes confirmation"), which the TODO row itself defined as valid closure                                                                                                                                                                           | ledger row 1; 23:22 report §3 as source                                                                                                         | TODO row closed             |
| 7  | **T11/r1 whitespace-drift `.nix` self-test** — the `.nix` matcher had ALREADY landed earlier today (verified via git log on the script); scratch-repo self-test executed: whitespace-only `.nix` staged edit → **exit 1** with the e62fe34 fix message; real content edit → **exit 0**                                                                   | bash self-test transcript (mktemp repo, both arms)                                                                                              | scripts/whitespace-drift.sh |
| 8  | **T12/e4 delta hygiene — codespell** clean over the v7 delta set (both 23:38 + 23:43 v7 status docs, the round-5 plan, dom-contract, lessons, ledger)                                                                                                                                                                                                    | `codespell …; echo $?` → 0                                                                                                                      | v7 docs                     |
| 9  | **T12/e4 delta hygiene — markdownlint real hit fixed** — one true positive: MD018 (`#4` at line start parsed as ATX heading) in lessons.md; fixed rewrap-proof by backticking `` `#4` `` after the daemon re-reflowed my first fix; final lint (MD013 disabled as posture) exit 0 over all five touched docs                                             | markdownlint runs before/after; final sweep 0                                                                                                   | docs/lessons.md             |
| 10 | **T21/M80 god-package carve trigger check** — `git log --since 2026-10-05 --diff-filter=A -- internal/server/` → zero new files; carve trigger NOT fired                                                                                                                                                                                                 | git query output                                                                                                                                | internal/server             |
| 11 | **TODO_LIST gate-recovery paragraph rewritten** — 6 rows (r1, r2, r6, r16, r17, r30) moved to CLOSED with one-line evidence each; REMAINING legs re-listed incl. the T16 deferral reason and the local wchan scan note                                                                                                                                   | TODO_LIST.md § "Gate-recovery session tail"                                                                                                     | TODO_LIST.md                |
| 12 | **AGENTS.md link wiring** — ci-flake-ledger added to the header home-list (135/377 lines, size preflight safe)                                                                                                                                                                                                                                           | AGENTS.md line 3                                                                                                                                | AGENTS.md                   |

## b) PARTIALLY DONE

| # | What                                                                                 | Done                                                                         | Missing                                                                                                                                                                                                                                                     |
| - | ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **T12 full scope** — v7 self-review e4 named "codespell + markdownlint over v7 docs" | The 23:38/23:43 v7 status pair, the plan, and all files this session touched | The older v7-cohort reports (22:08, 23:22) were NOT in the lint set — scope was interpreted as "the v7 sweep's own delta", which matches "the sweep's own e4 debt paid" but a broader read exists                                                           |
| 2 | **T04 SSE-lane evidence**                                                            | HTTP GET payloads (list fragment + transcript page) byte-probed live         | The SSE STREAM payloads (`/events` thread push) were not independently live-probed this session — they stay covered by `sse_test.go` byte assertions; the plan's "served payloads" wording is satisfied but the live SSE lane is inference, not fresh bytes |
| 3 | **T21 raw-idea triage**                                                              | Only the automatable M80 leg (carve trigger)                                 | The pick/park triage legs (paperless follow-ups, mic/typography extras) are owner-judgment prep, not started                                                                                                                                                |
| 4 | **T16 quiesced-host full flake check**                                               | Quiescence pre-check executed with process forensics (the M62 leg)           | The check itself deferred — host NOT quiesced (see d/e)                                                                                                                                                                                                     |

## c) NOT STARTED (in this session's slice — all routed, none dropped)

| # | What                                                                                                                                                                                                                                                                                                                                                                      | Where it lives                                                                                                                                                  |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | T16 the actual 90-min `nix flake check` incl. KVM backup leg                                                                                                                                                                                                                                                                                                              | TODO_LIST r15, DEFERRED with reason                                                                                                                             |
| 2 | T01 deploy chain, T02/T03 sittings, T07 passkey live ceremony, T08 live-call ritual, T09 stack batch, T10 BuildFlow binary/doctor/timings, T13 GOPROXY/retry dispatch, T14 fleet cleanup (local /tmp prune + gopls legs included), T15 run-closure/habit pins, T17 announcements, T18 docs v8 cohort sweep, T19 link-gate hardening, T20 upstream asks, T22 calendar pins | Round-5 plan; owner-terminal / cross-repo / live-deploy legs — not runnable from this repo/session                                                              |
| 3 | r14 kill-builtin empirical verify                                                                                                                                                                                                                                                                                                                                         | Gated on "the next natural kill" — this session killed processes via python `Popen.terminate()`, never via the broken shell builtin, so the trigger never fired |

## d) TOTALLY FUCKED UP

Nothing destructive: no data loss, no CI red, no cross-session interference,
no revert of anyone's work. The honest failures, ranked:

| # | Failure                                                                                                                                                                                                           | Cost    | Root cause                                                                                                                                                                       |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **MD018 fix lost a round to the daemon reflow** — my first fix (line-break shuffle) still left `#4` at line start; the daemon then reflowed the long merged line BACK to the broken split. Two wasted edit cycles | ~5 min  | I fixed the symptom (the wrap point) instead of the invariant (a line must not START with `#`), despite AGENTS explicitly warning "the daemon is adversarial to in-flight edits" |
| 2 | **r16 probe 403 on first boot** — POSTed `/api/session` without first fetching the CSRF cookie/token; wrote one nonsense line (`json.load.__doc__` check) in the first draft                                      | ~3 min  | Wrote my own probe before fully reading `Smoke.login`/`adopt_csrf`, which already document the exact dance                                                                       |
| 3 | **Scratch-file edit mangled `/tmp/r16_probe.py`** — a multiedit with mismatched indentation broke the retry loop; needed a second read+fix                                                                        | ~2 min  | Edited without a fresh View on a file I had just written                                                                                                                         |
| 4 | **markdownlint config evaporated mid-session** — `/tmp/mdlrc.json` was gone on the second run (tmp cleaning), so a "verification" rerun silently used default config and dumped 120 lines of MD013 noise          | ~2 min  | Put a gate config in `/tmp`; should have used a stable path or inline config                                                                                                     |
| 5 | **Left new /tmp artifacts behind** — `/tmp/webphone-r16`, `/tmp/wp-r16-data`, `/tmp/r16_probe.py`, `/tmp/mdlrc.json`                                                                                              | hygiene | T14/M57 says PRUNE /tmp probe artifacts; I ADDED four while not pruning any (mine or the stale ones)                                                                             |

## e) WHAT WE SHOULD IMPROVE

| # | Improvement                                                                                | Why                                                                                                                      |
| - | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| 1 | Fix invariants, not wrap points, when the daemon reflows markdown                          | The `#4` episode: a backtick (or hard `&#35;`) survives any rewrap; line-shuffles do not                                 |
| 2 | Read the maintained harness (`webphone-smoke.py`) BEFORE writing a custom probe            | The CSRF/login flow was already solved and documented there; my 403 was avoidable                                        |
| 3 | Check host quiescence FIRST when a plan row is gated on it                                 | T16 was the last item checked; had the window existed, the 90-min check could have run in background during the doc work |
| 4 | Keep gate configs out of `/tmp`                                                            | One evaporated config turned a verification run into noise output                                                        |
| 5 | Pair every new /tmp artifact with its cleanup in the same session                          | The repo keeps a TODO row about pruning probe artifacts; creating more un-pruned is the exact debt                       |
| 6 | Record the markdownlint posture in a rules home (AGENTS conventions or the ledger pattern) | It currently lives only in a TODO closure sentence; a fresh session would re-derive "MD013@80 is noise"                  |
| 7 | Treat "served payloads" claims precisely: page payloads vs SSE-stream payloads             | The ledger/dom note now say which lanes were probed; keep that precision habit                                           |
| 8 | When a plan row is closed by an OR-arm (r30), quote the arm in the closure                 | Done here; making it a habit prevents "closed" rows that actually punted                                                 |

## f) UP TO 50 THINGS TO GET DONE NEXT

Ordered roughly by the plan's Pareto spine; owner/cross-repo legs marked.

**The deploy chain (plan 1%→51%):**

1. T01/M01: verify webphone `main` green + `ls-remote` end-state + pick target rev (CI green at `2c86605`/`0949667`-on-push as of writing)
2. T01/M02: stack — repair/verify FreeSWITCH `mod_enum` build (stack repo)
3. T01/M03: stack — relock webphone input to the target rev (stack repo)
4. T01/M04: stack — `nix flake check` + browser E2E ×1, wall-time vs 445s budget (stack repo)
5. T01/M05: aarch64 cross-build, verify by ELF machine bytes (needs quiesced host for reliability)
6. T01/M06: pbx-artmann relock + gates + staged activation (pbx repo, owner stack tree)
7. T01/M07: owner deploy command + journal watch (OWNER terminal)
8. T01/M08: post-deploy smoke `--base https://pbx.artmann.tech --expect-version <V>`
9. T01/M09: record closure — TODO rows, CHANGELOG, E2E-obligation checkbox

**Sittings (plan 4%→64%):**
10. T02 sitting A: 38-row auth/release briefing part 1 (OWNER; digest prep M10 is assistant-runnable)
11. T03 sitting B: conventions/tooling postures rows 33–38 (OWNER; digest prep M15–M18 assistant-runnable)

**Verification loops:**
12. T16: quiesced-host full `nix flake check` incl. KVM backup leg (deferred this session — host had a concurrent session's `go test -race` suites + a monitor365 cargo check in flock-wait)
13. T04-tail: live SSE-lane byte probe (grep the `/events` stream payload for `wp-thread-row`/`wp-bubble`) to upgrade the b/2 inference to fresh bytes
14. f6: ui-capture 14-shot visual pass on the v1.20.1 tree (command in AGENTS; local-only by budget decision)
15. f12: vulnix `0/5 deterministic-retry` warning investigation
16. r14: kill-builtin empirical verify at the next natural kill (python one-liner, record the wchan)

**Tooling tail:**
17. T10/M44: rebuild the stale BuildFlow binary in its repo + reinstall (BuildFlow repo)
18. T10/M45: release-runbook prelude standing `buildflow doctor` gate line
19. T10/M46: `buildflow timings --regressions` rebaseline on the quiesced host (pairs with 12)
20. T10/M47: verify tool warnings shrank (4 devshell legs gone post-row-37)
21. T13/M53: crush-config GOPROXY fallback chain config/PR (crush-config repo)
22. T13/M54: BuildFlow retry-or-kill wrapper for network go steps (BuildFlow repo)
23. T13/M55: `--fail-on`/retry policy note + fleet rollout record

**Stack obligations batch (cross-repo):**
24. T09/M37–M38: stack `services.webphone.paperless` option + smoke arm
25. T09/M39: WebTransport NOT-ADOPTED verdict doc (stack repo)
26. T09/M40: deploy.md secret PATH column
27. T09/M41: ops-runbook demo-call recipe + `/var/lib/telephony-secrets/` path
28. T09/M42: `ftypqt` sniff fix + E2E MMS-outbound + pbx FEATURES:87 text
29. T09/M43: stack relock/push ritual for the batch

**Hygiene (host + docs):**
30. T14/M56: fleet wedged-go audit across machines (local leg done: one cargo flock-waiter, concurrent session's, left alone)
31. T14/M57: prune stale /tmp probe artifacts incl. THIS session's four (`/tmp/webphone-r16`, `/tmp/wp-r16-data`, `/tmp/r16_probe.py`, `/tmp/mdlrc.json`) — trash, never rm
32. T14/M58: close stale gopls editor workspaces (four orphaned gopls tidies were implicated in the 10-06 wedge)
33. T18: docs-health v8 — 2026-10 status cohort read + annotate/archive + gates + manifest + report (M68–M71)
34. T18/M72: split the gate-recovery TODO row into named sub-bullets
35. T19/M73: extend the link-existence loop to ALL living docs/*.md
36. T19/M74: pre-mv citation grep as a named gate step
37. T20/M75: templ-components errorpage tagging-discipline upstream issue (use verify-before-filing + github-voice)
38. T20/M76: go-cqrs-lite release-tooling audit (non-/v4 requires)
39. T20/M77: record upstream asks + watch entries
40. T17/M65–M67: announcements — owner picks channel, wording/disclosure posture, post (OWNER)
41. T22/M81–M83: calendar pins (erraudit 11-05, quarterly 12-20, sessions 12-30) + QMD subscribe (OWNER for auth scopes)

**Small, from this session's observations:**
42. Give the markdownlint posture a rules home (AGENTS conventions one-liner or extend the ledger pattern) — currently only a TODO closure sentence
43. Decide whether the r16 served-payload probe becomes a committed harness (`scripts/`) or stays ephemeral; if committed, fold into smoke as a `--grep-classes` arm
44. Extend T12's lint set to the 22:08 + 23:22 reports if the broader e4 scope is ratified
45. Confirm `0949667` pushes (daemon push-lag anchor 00:19 CEST; manual-push bar at >60 min per TODO convention) + check CI verdict after it lands
46. After the next session-boundary commit: `git ls-remote` + `gh run list` (the 2026-10-05 red-main lesson)
47. BuildFlow/whitespace-drift: consider adding the scratch self-test (M49's two-arm check) as a named check so the gate proves itself, not just today's transcript
48. dom-contract note: add the probe command snippet verbatim so the next templ-components bump re-runs it without archaeology
49. T21 remainder: prep the one-page paperless/mic/typography pick/park digest for the owner (M78–M79)
50. Plan hygiene: when round-5 is superseded, docs-health ANNOTATE resolves it — never rewrite (guard already in the plan)

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **T16 window:** the full `nix flake check` (90 min, KVM leg) needs a quiesced host, but another session is actively running `go test -race` suites and a monitor365 `cargo check` sits in flock-wait. Do you want me to (a) wait and re-check quiescence later, (b) run it now accepting contention, or (c) bundle it into T10's timings rebaseline on a scheduled quiet window?
2. **markdownlint posture:** ratify "MD013-off detect-only, no repo config" as the standing baseline (I'll pin it in AGENTS conventions), or adopt a committed `.markdownlint` config for the repo — and if so, which rules do you want enforced beyond detect-only?
3. **r16 probe disposition:** should the served-payload byte probe become a committed, repeatable harness (smoke arm or `scripts/` tool), or is the dom-contract note's re-run recipe enough? (Determines whether the next templ-components bump re-verifies via the maintained smoke or a one-off probe.)

---

_Verdict sources: probe run output; `git log/status/ls-remote`; `gh run list/view` 37537152790 / 37536492580 / 37536460912; codespell/markdownlint exit codes. Point-in-time snapshot — will go stale; section (f) is HARVEST input for TODO_LIST.md._
