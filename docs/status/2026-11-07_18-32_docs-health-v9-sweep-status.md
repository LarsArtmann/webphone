# Status — docs-health v9: the 2026-0* corpus archived, cadence drained to green (2026-11-07 18:32)

**Snapshot-rot header: claims verified as of 2026-11-07 18:32 CET.** The
final `docs-gates` matrix was ALL GREEN at 18:2x (freshness, audit-all,
annotate at its 205 pin, routing, cadence ARCHIVE-HEALTHY at 70 live,
table-damage). Four concurrent frontend-ux sessions landed reports and
trains DURING this sweep — the last hour of the session was partly
triage of their fallout. The daemon owns the commits (18 changed files
at writing); nothing pushed by me, nothing reverted.

**Scope:** the owner's repeat of the docs-health mandate over the
`2026-0*` corpus. v8 (yesterday) had classified all 44 files but
deferred the row-level pass on the four executed September plans; this
session executed that pass, archived the plans, drained the oldest
status reports to bring cadence under the pin, and updated the living
docs. Documentation-only: zero product code touched.

## a) FULLY DONE

| #  | Work                                                                                                                                                                                                                                                                                                                                                                                                                                           | Verification                                                                                                                                |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **All four executed September plans taken through the full row-level pass and ARCHIVED** — error-excellence (95 rows + 25 prose), 20-year-durability (217 + 6), pareto-execution-v260 (118), stack-adoption-pareto (123): ~553 table rows + ~31 prose items, each struck per-cell with evidence verified THIS session (git grep / CHANGELOG quotes / AGENTS bullets / code paths), not transcribed on faith                                    | `check-rows --strict` + `bare-rows` + `prose-items` green per file BEFORE each `git mv`; post-move `grep -L '~~'` clean (completeness gate) |
| 2  | **Mechanical row enumeration** for the durability plan (awk over table line ranges → 182-row spec, 0 unmapped) — the method the other three plans should have used from row one (see §d)                                                                                                                                                                                                                                                       | `repair` dry-runs matched every spec line before each apply                                                                                 |
| 3  | **Verification finds en route**: retention job SHIPPED (`internal/retention`, ROADMAP says so), nginx gzip SHIPPED, data-sms + history ☆-save shipped, image thumbnails FULLY_FUNCTIONAL, `turn_rest` shipped, `canonicalRow` pin live, `httpspec_test.go` live — each turned into a per-row verdict instead of a guess                                                                                                                        | grep receipts in session                                                                                                                    |
| 4  | **Cadence drained 71 → 70 → 74 (concurrent landings) → 70 → GREEN**: seven oldest status reports archived with full row passes — 10-04 ui-redesign (8 + 21 prose), 10-05 mixin-review (8 + 3), 10-06 00-52 (6 + 4), 10-06 03-40 (6), 10-05 21-41 round2-brutal (59 + 18), 10-06 01-40 registry-writer-brutal (9 + 54), 10-06 15-25 round4 (8 + 29)                                                                                             | same gates before each move; `status-index.py --check` clean after each                                                                     |
| 5  | **Citation repoints (12 sites)** before/around the moves — CHANGELOG ×1, error-contract ×1, owner-calls-briefing ×1, dedup-registry ×1, five archived citers + one partial-path mention — so the moves minted ZERO new broken-path findings (the 205 ratchet held)                                                                                                                                                                             | repo-wide basename grep pre-move; residual grep post-move: one hit, fixed                                                                   |
| 6  | **ARCHIVED banners** (v9 sweep form) on all ten moved files, each naming the verdict + where open legs route                                                                                                                                                                                                                                                                                                                                   | head-check per file                                                                                                                         |
| 7  | **Living docs updated**: TODO row 11 rewritten (plans drained, recipe named), TODO row 20 ADDED (harvest of the fresh frontend-ux §f), TODO row 21 ADDED (tooling nits found this session), ROADMAP § Open questions gains the carrier-leg owner question (MMS isolation test + GV verdict + number ownership — found orphaned in the error-excellence plan, routed nowhere since September), AGENTS residue bullet updated with the v9 recipe | edits in tree; AGENTS 196 lines (cap 377)                                                                                                   |
| 8  | **Concurrent-session fallout triaged**: 4 new reports + 2 trains landed mid-sweep; freshness went FAIL (3 unharvested §f items in the 11-07 frontend-ux report) → harvested as TODO row 20; annotate went 206 > 205 → root-caused to MY OWN defect (see §d.5) and repaired                                                                                                                                                                     | final gate matrix ALL GREEN                                                                                                                 |
| 9  | **Formatter-stability held**: dprint fmt re-run over every annotated file; `check-rows` re-verified AFTER fmt (the canon's both-sides rule) — per-cell strikes survived every pass                                                                                                                                                                                                                                                             | post-fmt gate runs, all clean                                                                                                               |
| 10 | **INDEX regenerated + `--check` clean** after every status/ move (3 regens)                                                                                                                                                                                                                                                                                                                                                                    | `python3 scripts/status-index.py --check` exit 0                                                                                            |

## b) PARTIALLY DONE

1. **The archive residue itself (~65 live status reports, ~1 600 §a/§b
   rows)** — this session proved the per-row recipe at scale (ten files)
   but the corpus is the same work in the Verdict-column dialect. Cadence
   is green AT the pin (70), so residency cannot grow silently, but the
   drain is TODO row 11 and it is long.
2. **Living-docs verification was spot-check, not re-audit.** README and
   FEATURES were fully verified by v8 eighteen hours ago; I verified the
   claims this session's verdicts lean on (reportsStat in `[Unreleased]`,
   MMS rows, turn_rest, export, favicon) but did NOT re-walk every row —
   and two frontend-ux trains landed TODAY touching admin/settings copy
   that I did not re-verify against FEATURES.
3. **The three different "live reports" counters.** INDEX regen said 76,
   raw ls said 75→70, the cadence gate said 71→70→74 across the session.
   They count different things (INDEX includes itself? the gate excludes
   non-timestamped files?) — noticed, recorded as TODO row 21c, NOT
   root-caused.
4. **The §Opening bold-numbered headings in the 01-40 report**
   (`**1. What did you forget?**` …) were left unstruck — the scanner
   classifies them narrative, not items, so no gate demands verdicts; I
   concurred but did not record the classification rationale anywhere.

## c) NOT STARTED (correctly out of mandate)

- The 2026-10-* planning/ cohort (~60 files) — outside the `2026-0*`
  pattern; the next sweep's candidates as trains close.
- The pre-canon ARCHIVED corpus (205 ratcheted findings) — CLOSED
  HISTORY per AGENTS; untouched, ratchet unchanged.
- CI wiring for docs-gates (TODO row 10), freshness §f-less WARN (row
  17), DF-brittleness upstream (row 19) — standing rows, not mine today.
- Any push: the daemon owns commits AND pushes; per AGENTS I verify, I
  do not push.

## d) TOTALLY FUCKED UP

1. **Four parallel verification subagents, four rate-limit corpses.** I
   dispatched all four plans to subagents at once; every one died with
   `too many requests` before doing a byte of work. One round of
   planning and a dispatch cycle burned for nothing — and the fallback
   (sequential, myself) was always available. Check the provider's
   appetite BEFORE fanning out, or fan out one-and-see.
2. **I missed a table THREE times** (error-excellence's line-79 design
   table, pareto's line-40 second row, stack-adoption's §5 verification
   table). Each was caught by `check-rows --strict` — the gates did
   their job — but the root cause is that I enumerated tables from the
   headings grep instead of mechanically. The durability plan's awk
   enumeration (0 misses) was the right method all along; I only
   graduated to it mid-session.
3. **Eight spec rows nearly shipped as literal "UNMAPPED" verdicts**
   (single-digit `F1`–`F8` fell through my case patterns in the
   round2-brutal spec). The tool would have APPLIED them — "UNMAPPED"
   is not in the banned-verdict list — had I not grepped for UNMAPPED
   before applying. That is the fake-verdict incident class one grep
   away from shipping. I fixed the rows AND filed the harden request
   (TODO row 21b), but the near-miss was mine.
4. **A "no output" shell failure cost a diagnosis round-trip**: the
   tool shell (mvdan/sh) rejects function definitions inside `&&`
   chains — my spec generator silently died at the first `emit()`.
   Diagnosed by bisecting the command instead of reading anything;
   avoid shell function defs in this tool shell, period.
5. **My own annotation minted the +1 corpus finding.** The durability
   plan's fine table had escaped pipes inside a code span (`\|`); the
   repair tool's cell splitter cut at them, leaving a strike that
   crossed a cell boundary — file-local gates stayed green, the corpus
   annotate gate went 206 and I initially suspected the concurrent
   sessions' new reports. Root-caused to my file, hand-repaired the
   cell (pipes → commas), filed the fixture request (TODO row 21a).
   Lesson: file-local gates prove uniformity, not shape; the corpus
   scanner is the shape authority — run it mid-sweep, not only at the
   end.
6. **Two edit-tool mod-time guard trips** (banner swaps, AGENTS edit)
   after my own annotate/fmt writes moved the files — the documented
   hot-file pattern; each recovered by re-View + retry, but I should
   have gone View-first without being told.

## e) WHAT WE SHOULD IMPROVE

1. **Enumerate rows mechanically, ALWAYS** (emit-keys / awk over line
   ranges), never from a headings grep — the three missed tables were
   all enumeration misses, not verification misses.
2. **Run the corpus annotate gate at sweep midpoint**, not only at the
   end: my durability defect sat in the tree ~40 minutes and got
   attributed (briefly) to the concurrent sessions before I looked.
3. **Validate spec verdicts before apply** — a grep for the sentinel
   (UNMAPPED) saved this run; the tool should refuse non-canonical
   verdict heads outright (filed: TODO row 21b).
4. **Fan out subagents only after a rate-limit probe** — one cheap
   dispatch or a provider check beats four corpses.
5. **The four-count problem** (live reports / INDEX / cadence gate /
   audit-all replay count) needs one derivation — TODO row 21c.
6. **When the freshness gate reds on a NEW concurrent report**, harvest
   immediately (as done for TODO row 20) — the two-minute fix beats
   leaving main red on someone else's §f.

## f) Up to 50 things we should get done next

_(Rows 1–19 below are the standing TODO_LIST rows this session verified
still open; 20–21 are this sweep's own additions; 22+ are session
observations with their homes. The freshness gate matches this §f
against the living docs — every item names its row.)_

1. Wire the docs-gates battery into the devShell + CI (TODO row 10) —
   until then "we have a docs gate" is a local habit.
2. Bulk verdict pass over the archive residue (TODO row 11: the live
   status reports, ~1 600 un-inventoried Verdict-column rows; the v9
   recipe is proven at ten files).
3. Freshness should WARN on a §f-less newest report (TODO row 17 — the
   vacuous-PASS class this session started under).
4. Frontend-UX verification tail (TODO row 20): admin.calls.lines DE
   wrapping, F20 Settings dl-count pins, `#wp-palette-open`
   aria-expanded.
5. The carrier-leg owner question (ROADMAP § Open questions, new):
   inbound-MMS isolation test + GV verdict + `+14084757593` ownership.
6. The annotate/docs-gates tooling nits (TODO row 21): escaped-pipe
   fixture, UNMAPPED hard-error, one live-report counter.
7. Record the DF-brittleness + standalone-verb gotchas upstream (TODO
   row 19).
8. Decide the canonical docs-gates binary (v8 §g.2 — dev vs nix-pinned;
   still unanswered).
9. THE SITTING: the 2026-10-29 fresh decision packet + remaining
   briefing rows (TODO sitting row) — it unlocks ~10 planned tasks.
10. Post the release announcements backlog (TODO announcements row).
11. The `/context` Kith-contract snapshot test (TODO row 12: pin
    webphone's parse against Kith's published shape).
12. History nudge for the notes save path (TODO row 13).
13. `scripts/templ-generate.sh` wrapper pinning the compiler (TODO row
    14).
14. Attended-transfer UI vs the NO-GO ruling (TODO row 15).
15. `reverseLines` → `slices.Reverse` (TODO row 16).
16. Enforce CHANGELOG coverage for code-path commits (TODO row 18).
17. Findings-gate posture: 50 tier-3 findings keep buildflow red (TODO
    row 7).
18. Quiet-host hub fan-out re-run (TODO row 8).
19. Store test train at 64.8% (TODO row 9).
20. Re-run this session's final gate after the daemon commits (and
    after any concurrent session lands) — CI-watch per AGENTS once
    pushed.
21. Sweep-log discipline check: no dedup-registry line was owed this
    session (no art-dupl run, no clone ruling) — keep skipping honestly.
22. Consider a `fresh:skip`-style ledger exemption for §Opening bold
    headings if the scanner ever starts classifying them as items (the
    01-40 precedent — currently correctly narrative).
23. When archiving plans with `|`-bearing code spans, pre-rewrite the
    cell (pipes → commas) BEFORE striking — the repair-tool split-then-
    strike order is what minted the boundary-crossing.
24. Keep the plan-archive recipe: banner LAST, fmt, file gates, corpus
    annotate gate, THEN `git mv` — order held this sweep and caught
    everything.

_(24 real items; the standing TODO_LIST rows 1–19 carry the rest.
Padding to 50 would be inventory theater — same verdict as the round-2
report's 42.)_

## g) Three questions I cannot answer myself

1. **Residue-drain depth.** For the remaining ~64 reports I used the
   row's OWN Verdict-column evidence, hash-spot-checked, and struck
   per-row (v8 §g.1's middle path, executed). Keep that depth for all
   64 (slow, days), or is a lighter standard sanctioned once a report's
   §a/§b tables are uniformly evidence-bearing — and if so, where is
   the line?
2. **The cadence pin sits exactly at the live count (70/70).**
   Concurrent sessions land 3–5 reports/day, so ANY new report reds
   main on cadence until the next drain. Raise `max_unarchived` with
   the landing-rate evidence (ratchets may rise only with evidence —
   this is that evidence), or hold 70 and treat each red as the drain
   drumbeat?
3. **Subagent fan-out policy on this provider:** after the four
   rate-limit corpses — is sequential-only the standing reality (I
   stop attempting fan-out), or should verification legs retry with one
   agent at a time / a different model before falling back?
