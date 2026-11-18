# Status — line-ownership follow-up closeout: the whole remaining list, verified

**Claims verified as of 2026-11-18 23:42 CET.** Push/CI statements rot
fastest; everything below names its evidence and verification moment.
Session scope: the owner's "NOW GET SHIT DONE! The WHOLE TODO LIST!"
directive over the 20-03 report's remaining items (f.3, f.4, f.8-row,
f.16, f.22-recheck, final battery, cadence drain) plus every red the
verification surfaced.

## a) FULLY DONE (verified this session)

| # | Item                                                                                                                                                                                                                                                                                               | Evidence                                                       |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| 1 | **f.16 hot-reload sketch** — `docs/planning/2026-11-18_ownership-hot-reload-sketch.md` written (MapSource atomic-swap seam, boot-equal validation doctrine, live-session semantics, SIGHUP-vs-file-watch table, non-binding SIGHUP recommendation); the TODO_LIST row's dangling reference is DEAD | commit `b13e79f5`; implement arm stays owner-gated by design   |
| 2 | **f.8 TODO row** — the pma push-heuristic lands as a cross-repo pma/go-commit train row (age heuristic + rollback + canary), never a webphone edit                                                                                                                                                 | TODO_LIST § Line-ownership residue, commit `b13e79f5`          |
| 3 | **f.4 visual matrix** — fixture-boot run GREEN: 28 shots + the layout proof + the 1600px wide-canvas probe; the run EXPOSED a stale check (below) and a false fixture assumption (below)                                                                                                           | `scripts/ui-capture.py` run 22:02–22:08, exit 0; self-test 8/8 |
| 4 | **f.3 stack browser E2E** — the stack re-pins webphone `b63bb819` → `cfba7252` and the telephony-browser VM lane ran GREEN FIRST TRY: E2E-OK, DTMF 5, watchdog reconnect-recovery, 350.76 s (< 445 s budget, no flake rerun)                                                                       | `/tmp/stacke2e.log`; stack commit `248bcab` (its daemon)       |
| 5 | **f.22 recheck** — issue #253 CLOSED by the concurrent train (postcall-note reds fixed); island suite 364/364 locally confirms; NO split needed                                                                                                                                                    | `gh issue view 253`; island run exit 0                         |
| 6 | **Full battery** — island 364/364; `go test -count=1 ./...` GREEN twice (37 pkgs, including post-refactor); tsc exit 0; smoke 200/200 across 13 scenarios; `format` sandbox check green ×2; docs-gates ALL GREEN                                                                                   | logs cited inline below                                        |
| 7 | **Cadence drain** — 3 October reports archived (68 → 65 unarchived): 10-28 red-heal pair + 10-29 layout-leverage; check-rows --strict + bare-rows green BEFORE each `git mv`; strike-safety 3/3; banners carry the per-row adjudication; INDEX regenerated                                         | commit `cf4a289e`; docs-gates matrix PASS                      |
| 8 | **CI main GREEN** — tip `cfba7252` concluded success 23:40:41 with ALL jobs green (incl. the two advisory lanes actually running)                                                                                                                                                                  | `scripts/ci-verdict.sh cfba7252`                               |

## b) The reds the verification surfaced (all healed same session)

| # | Red                                                                                                                                                                                                                             | Heal                                                                                                                                                                                                                                                                                                                |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **ui-capture layout proof RED** ("the split must be a row, got 'column'") — the Oct-29 check hardcoded "row at 1280 expanded" from pre-rail geometry; today's docked rail (216px) + island (400px) leaves 576px < 44rem at 1280 | commit `15b1a085`: the expanded probe derives its expectation from the MEASURED column width (the ladder's own rule) + a new 1600px wide-canvas probe keeps the desktop-split proof alive where it is true                                                                                                          |
| 2 | **CI art-dupl RED on bde1ebeb** (4 new groups; 3 in my line-ownership rework, 1 foreign 2-liner)                                                                                                                                | commit `b5f39d2b` + `e92f490d`: `dirChip` ONE home across 4 sites (fax ×2, admin CDR, history CDR — the admin comment's "reuses History vocabulary" is now structural), `adminNumberMeta` composed line, 2 intentional idioms baselined with the pinned v0.7.2 (registry sweep-log entry; 111 → 108; check = 0 new) |
| 3 | **CI docs battery RED** — dprint format (the third session's unaligned report tables) + audit count (healed by the corpus growth of the same-day docs commits; measured 66 ≤ 72 pin)                                            | commits `e9518673` (dprint-align, format-only) + `cf4a289e` (cadence drain)                                                                                                                                                                                                                                         |
| 4 | **CI workflow-e2e advisory NEVER RAN** — `getFlake "path:."` relative-path eval error (the concurrent train's new job)                                                                                                          | commit `2076c4c0`: absolute `path:$PWD` (the AGENTS recipe's form); the journeys ran GREEN on cfba7252                                                                                                                                                                                                              |

## c) NOT STARTED (owner-gated or foreign, none silently dropped)

1. The 20-03 §g questions stay owner-gated (switcher/account home;
   config-edit ritual — the f.16 sketch now FEEDS the second one).
2. The pma train itself (f.8) — row landed, work is cross-repo.
3. Android `lines[]` adoption + SSO-only stack-fixture OIDC leg —
   standing TODO rows with named triggers.

## d) TOTALLY FUCKED UP (mine, all recovered)

1. **`strike_row` never wrote the file** — the drain's first strike
   pass edited lines in memory and silently discarded them; caught by
   the gate re-run (bare count unchanged), fixed by adding the write.
2. **Verdict cells appended AFTER the final table pipe** — an extra
   cell per row; strike-safety REFUSED on exactly this class, and
   dprint's table formatter then dropped the overflow cells (the
   verdict prose vanished). Outcome: the adjudication lives in the
   banners (the v10 precedent's shape anyway) and the rows carry the
   strikes — gates green, safety green. Lesson: verdicts go INSIDE the
   last cell or the banner, never past the final pipe.
3. **A mangled layout.templ edit** (`}}` merge) — my own old_string
   omitted a newline; caught by immediate re-view, fixed before
   generate.
4. **A multiedit applied against a stale read** (TODO_LIST mtime race,
   daemon) — re-View then re-apply, no damage.
5. **Commit message typo** ("f.18" for the f.8 row) in `b13e79f5` —
   cosmetic, left in history rather than amend-racing the daemon.

## e) WHAT WE SHOULD IMPROVE

1. The fixture's admin-surface assumption was WRONG twice over (the
   prior session assumed it, this session's run disproved it): the
   harness session is an extension login, admin is identity-gated,
   ceremonies cannot be automated — the comment now records WHY
   (`cfba7252`), and /admin stays a prod --track lane.
2. The daemon's push leg stalled 60+ min while CI-eligible heals sat
   local — the manual-push bar fired exactly as ratified (anchor
   `e9518673` 22:31:03, pushed 23:31:38). The pma row (f.8) owns the
   daemon-side fix.

## f) NEXT

1. Owner answers to the 20-03 §g trio (the f.16 sketch arms q2).
2. The standing TODO rows: pma push heuristic; Android lines[]; the
   SSO-only OIDC fixture leg; ownership-hot-reload implement (gated).
3. Next stack fixture pass picks up the OIDC lane check (f.6's
   trigger).

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

Unchanged from the 20-03 report §g — nothing new was minted that needs
an owner call; this session executed, it did not decide.
