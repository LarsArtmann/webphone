# Status — SUPERB deepening + integrity execution (carves, depth budget, views props)

**Snapshot:** 2026-11-12 13:15 · claims verified against the tree at this moment ·
Branch `main`, all session work pushed · CI: GREEN on `a960d19c` (the dupl
repair), in-progress on `190a8c2e` (views props) · Full Go suite green at the
last full run; island/tsc untouched this session (no JS/markup bytes changed).

Executing `docs/planning/2026-11-12_05-25_SUPERB-deepening-and-integrity-plan.md`
(the T01–T24 / F01–F62 Pareto plan) task by task, one revertable unit per task.

---

## a) FULLY DONE (committed + pushed + verified)

| Task | What landed | Evidence |
| --- | --- | --- |
| T01 ledger batch | Hold-verdict + ring-fan-out TODO rows struck with pin sites; CHANGELOG gained the four deepening entries | `7475f9fb` + daemon `01b5e9b9` |
| T02 race probe | `TestConcurrentFirstSightYieldsOneThread` (race-detector clean): **NOT REPRODUCIBLE** — both appends hit ONE journal stream, `cqrs.thread_identity` rejects the loser, re-delivery lands on the winner's thread. No guard needed | `internal/messaging/raceprobe_test.go`; TODO row struck with the full verdict |
| T03 blob orphans | Compensating deletes on all four lanes (messaging Send/Receive save loops incl. mid-loop failure, fax Send/Receive spool); `rollbackSaved`/`unspool`; 4 failure-path tests pin "persist fails ⇒ blob tree empty". Crash-window sweep question KEPT for the owner batch | service.go + blobcomp_test.go both packages |
| T04 auth carve | `internal/server/auth`: both login modes, ceremonies, OIDC, library /login, session spine behind the narrow `auth.Shell` seam (the api/hooks precedent). Route table/flood budgets/nonce lane untouched — every session/passkey/SSO spec passes unchanged. Error registry regenerated (3 moved sso.* sites); i18n dead-key scan extended to carved subpackages; `Deps.SSO` retyped `auth.SSOClient` | `0849037c` (+ daemon `7175ba5b`/`7f524963` carried pieces) |
| T06 platform carve | `internal/server/platform`: assets, /metrics + recorder types (exported `LatencyRecorder`/`IslandLatency`), CSP intake + Reporting-Endpoints middleware, flight-recorder seam (+ `platform.TraceRecorder` new home, `TraceMiddleware`), /phone-api proxy, directory bridge (+ `machineSurface`/`loopbackPeer`), /config.js, export. Voicemail SSE nudge + bearer gate STAY in server (one-home) | `1dbc1bd5` |
| T05 html carve | **RULED OUT structurally** — pages.go holds the `handlers` hub itself; panels/actions reference admin/crm/unread/auth-audit in both directions; moving them renames the package rather than splitting it. Recorded in the TODO carve row (rewritten: executed auth+platform, html NO, reopen only on a real architecture change) and in the `1dbc1bd5` commit message | TODO_LIST carve section |
| T07 depth budget | `internal/arch/depthbudget_test.go`: every package fits 1500 hand-written LOC / 60 exports or carries a NAMED measured pin (cqrs 6000/100, domain 2500/180, server 5200/60, api 1900/60, store 1900/60, app 1650/40). Ratchet semantics (raise = number+reason same commit). **Mutation-proven**: planted a 170-export package → red naming the fix → removed → green. AGENTS module-graph bullet extended | `4c1c7b32` |
| art-dupl red-main repair | CI went red on the carve commits (2 new clone groups). Adjudicated BEFORE any baseline write: platform `verifyTURNPair` was DEAD once its test moved (deleted — the honest fix); the per-package `writeJSON` 5-liners are the carve precedent's intentional parallel (api's variant is status-free by wire contract — a shared home would be a false unification) → baselined. Registry sweep-log line appended. Baseline **117 → 110** (the carves dissolved 7 groups); post-write check = 0 new; **CI green on the repair** | `a960d19c`, run green |
| T08 views props | `views/types.go` mirrors (VoicemailItem, VoicemailSummary, AgentReviews, AgentDraft, AgentFrozen, AgentSuppressed, CDREntry); ONE converter home each in panels.go (`voicemailItems`/`agentReviewsOf`/`cdrEntries`); templates + tests swapped; compose cap moved to `domain.MaxBodyLength` (messaging aliases it). **views' import census = domain + readmodels + reports only** — the two ratified shape homes. templ regenerated with the pinned CLI v0.3.1100; no markup/class bytes changed | `190a8c2e` (+ daemon `69020979`/`9ded77a4` carried pieces) |

**Honest measured outcome for the god-package finding:** server is now
4 989 hand-written LOC / 13 top-level exports (was ~8 500+ across what became
four packages), with auth (1 112 LOC) and platform (1 343 LOC) as real
subpackages beside api (1 776) and hooks (662). The html residual is the
named, pinned exemption — the debt is mechanical now, not narrative.

## b) PARTIALLY DONE

- **T09 CI verdicts**: verdicts checked via raw `gh run list` (repair green,
  T08 in progress). BUT `scripts/ci-verdict.sh` answered NO-RUNS for shas that
  demonstrably have runs — the instrument disagreed with the raw source and I
  did NOT debug it this session. The verdict-log half (F31) is not written.
- **CHANGELOG coverage**: entries added for auth carve, platform carve,
  views props, depth budget — but the platform-carve CI run showed the
  CHANGELOG job "skipped" and the daemon folded some CHANGELOG edits into
  noise commits; the render/coverage state needs one confirming look.
- **The plan document** still lists T05 as pending work — the "amend the plan
  in the same turn" lesson (itself a TODO_LIST candidate) says the T05
  NO-GO ruling should be written INTO the plan file, not only into the TODO
  row + commit message.

## c) NOT STARTED (from the plan, in execution order)

- T10 docs-gates battery + ratchet re-measure (F33–F34)
- T11 store-ladder retirement row + legacy-reader headers (F35–F36)
- T12 push.Transport defined type + boot rejection (F37–F38)
- T13 FaxJob.DocumentPath → domain.RelPath (F39–F40)
- T14 wiring structs + machineSurface bool-last (F41–F42) — NOTE: machineSurface
  MOVED to platform this session; F42's site marker changed
- T15 verdict-lane UI pins (F43–F44) — the sipverdict stubs verified in tree
- T16 DE locale fixes (F45–F47)
- T17 admin CDR party fallback (F48) — the admin CDR surface now converts
  through `cdrEntries`, so the fallback belongs in the view type's extractor
- T18 receiver hardening (F49–F52)
- T19 postcall.js extraction (F53)
- T20 glyph→SVG sweep (F54)
- T21 owner-calls prep doc (F55, F62)
- T22 erraudit nolint-audit + calendar sweep (F56, F61) — due date 2026-12-08
- T23 cqrs-lint upstream drafts (F57)
- T24 release tails: 18-08 drain, flight-recorder docs, release-log policy (F58–F60)

## d) TOTALLY FUCKED UP (owned)

1. **The daemon repeatedly committed my in-flight work as noise commits**
   (`7175ba5b` carried half the auth carve before my narrative commit;
   `f1c5ac6d`/`61162f96`/`3631c6d5`/`69020979`/`9ded77a4` the same for
   T02/T03/T07/T08). The plan's one-revertable-commit discipline was
   repeatedly defeated: narrative messages sit on partial content while
   daemon `chore:` commits carry the rest. Nothing was lost, but the history
   is noisier than the plan promised.
2. **`git add -A` swept the concurrent session's file into my commit** —
   `1dbc1bd5` (platform carve) includes their `docs/status/2026-11-12_12-52_go-appkit-*`
   report. Harmless content-wise, but it is THEIR work in MY commit.
3. **Mid-task red states hit the daemon's commit window**: the T08 templ
   swap sat stale-generated (build red) across one daemon commit cycle; an
   earlier python one-liner failed silently (syntax error) making a
   mutation test look green when the bomb never existed. Both healed in the
   next step, but the pattern (heal FIRST, verify, then commit) was violated
   twice before I tightened it.
4. **The messaging suite threw one unexplained FAIL** (first combined run
   after the race probe), then passed 3× singles + count=3 + combined + race.
   Never root-caused; attributed to daemon/CPU noise without evidence.
   Recorded here so the pattern is searchable.
5. **The export census I promised in the plan ("342 exports") was measured
   with a sloppy method** (method-line grep inflating the count). The honest
   instrument is top-level exported declarations: server 13, auth 27,
   platform 29, api 33, hooks 21. F23's "≤80 across the three" framing was
   miscalibrated from the start.

## e) WHAT WE SHOULD IMPROVE

- **Carve trains need daemon coordination**: stage+commit in ONE call is not
  enough when the daemon's window is ~1 commit/5 min and a carve is ~40 min
  of in-flight edits. Either pause the daemon during carves or accept
  scaffold-then-squash.
- **`git add -A` is banned for my commits** — explicit paths only (the
  concurrent session's files keep landing in mine).
- **The art-dupl gate belongs in the pre-carve preflight** (like the release
  preflight batch): the carve moved ~2 000 lines and the gate red was
  predictable — a `buildflow -s art-dupl` or local pinned-detector check
  before push would have caught both groups pre-CI.
- **ci-verdict.sh's NO-RUNS vs gh's real runs** must be reconciled before it
  is trusted for release posture (it is the instrument the release runbook
  leans on).
- **Plan amendments belong in the plan file the same turn** (the existing
  lessons candidate — T05 proved it again).

## f) NEXT (up to 50, best-first; T-numbers = plan tasks)

1. Finish T09: confirm CI green on `190a8c2e`; reconcile ci-verdict.sh's
   NO-RUNS mismatch; write the verdict log (F31–F32).
2. T10 docs-gates battery (`scripts/docs-gates.sh` in nix develop) + ratchet
   re-measure rows (F33–F34).
3. Amend the plan file with the T05 NO-GO + measured carve outcomes.
4. CHANGELOG coverage confirming look (the CI job skipped on the carve run).
5. T12 push.Transport defined type + ParseTransport + boot rejection (F37–F38).
6. T13 FaxJob.DocumentPath → domain.RelPath; retire the fax.go marker (F39–F40).
7. T14 asr/matrix wiring structs; machineSurface bool-last — at its NEW
   platform home (F41–F42).
8. T15 verdict-lane UI pins: rejected/timeout/send-failed island specs (F43–F44).
9. T16 DE locale: reports weekday, crm month label, unreadDivider aria-label (F45–F47).
10. T17 admin CDR party fallback + ladder test — now over `CDREntry` (F48).
11. T18 receiver hardening: matrix oversized-sync streaming + garbage-cursor
    self-heal (F49–F50).
12. T18 FCM 401/403 token invalidation (F51).
13. T18 CRM read/context body caps (F52).
14. T19 postcall.js extraction from calls.js (~2 000 → ~1 400) (F53).
15. T20 glyph→SVG sweep + dialer-focus residue (F54).
16. T21 owner-calls prep doc: number every sitting question with evidence (F55).
17. T22 erraudit nolint-audit post-refactor decay check — the carves moved
    ~20 nolint-comment-bearing sites; expect drift (F56).
18. T22 calendar sweep: confirm next-due dates after this week (F61).
19. T23 cqrs-lint upstream drafts (suppression state machine + D005), owner
    tick gate (F57).
20. T24 18-08 drain train with archive gates-first ritual (F58).
21. T24 flight-recorder/docs cluster + traces-in-backup question into the
    owner doc (F59).
22. T24 release-log archival policy row (F60).
23. T21/62: verify brandid-lint posture + Client polish #48 + ASR deploy-proof
    rows still owner-gated (F62).
24. Update TODO_LIST full-code-review section: the two strike rows + probe
    verdict are done — sweep the section for coherence.
25. Re-run `buildflow` full at the next train boundary (the carve train
    changed ~2 400 lines; the scoped steps were never run this session).
26. Island suite + tsc after T19/T20 (the only planned JS-touching tasks).
27. Stack browser E2E obligation: NOT owed (no served-markup bytes changed) —
    verify that claim once the island tasks land.
28. Docs-gates freshness after this report (this file adds a status doc —
    INDEX regen via scripts/status-index.py).
29. Consider a `depthbomb`-style permanent mutation test harness note in the
    dedup/registry docs (the F25 proof was manual).
30. CHANGELOG [Unreleased] final proofread before the next release cut.
31–50. Remaining plan fine-tasks F37–F62 in the plan's order table; each is
    ≤100 min and named above — the plan's own table is the checklist.

## g) QUESTIONS FOR YOU (cannot answer from the repo)

1. **T05 ruling acceptance**: I ruled the html carve structurally impossible
   (evidence in TODO_LIST + `1dbc1bd5`). Do you ACCEPT that NO-GO, or do you
   want the deeper redesign (per-tab render packages over a shared render
   kernel) scoped as its own plan before any code moves?
2. **Daemon vs carve discipline**: during carve trains, should I (a) keep
   fighting the daemon with one-call commits (current), (b) you pause the
   daemon for carve sessions, or (c) I stop trying for revertable units and
   let the daemon noise stand?
3. **ci-verdict.sh**: it answered NO-RUNS for shas with real runs (raw gh
   showed them). Want me to debug/fix the instrument now (it gates release
   posture), or is raw `gh run list --workflow CI` acceptable as the
   authority until the next release train?

---

**Unexplained transient worth remembering:** one messaging-suite FAIL
(non-reproducible across 5 subsequent runs incl. -race) immediately after the
race probe landed. If it recurs, suspect daemon sqlite/tmp contention first.
