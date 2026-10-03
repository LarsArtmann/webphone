# Status — docs-health v6 RESUMED: 11 more files annotated+archived (15/34), 19 remain, living docs re-routed

**Date:** 2026-10-03 02:09 CEST (session ran ~00:50 → 02:09, resuming the 00:35 paused sweep).
**Mandate:** the owner's resume order ("READ, UNDERSTAND… Keep going until done") lifted the
paused report's §g wait. Read as answering its three questions by plan-default:
q1 = archive the 8 closed 2026-10-02 chain files (12:57 close-out + 13:50 session-8 stay
live), q2 = archive v5 with a superseding v6 banner, q3 = the 4 owner-gated planning docs
stay LIVE pending explicit ratification. Documentation-only: zero product code touched.

## a) FULLY DONE (verified this session)

1. **Skill + tooling semantics re-derived from source** (not memory):
   `annotate-status-items.py` (whole-line `~~line~~ verdict` wrap, first-unstruck-match,
   duplicate-key hard error, `any:` last-resort keys, banned open verdicts) and
   `check-rows.py` (PARTIAL-cell + mixed-table detection; wrapped rows drop out of table
   blocks, so the grep gate + my per-item accounting carry the completeness bar). The
   accepted v3/v4 archived examples inspected for the house banner/strike form; v3
   confirmed whole-row table strikes ARE precedent (19 rows).
2. **Evidence base loaded for verdict derivation:** the 12:57 close-out report, full
   TODO_LIST/CHANGELOG/ROADMAP (both halves), plus fresh verify-greps: tabular-nums
   actually SHIPPED (19 sites both sheets — the ROADMAP straggler bullet was stale),
   `text-wrap: pretty` only on bubbles, `Receipt.Resolution` at gateway.go:43–56, config
   fail-closed at config.go:319–334 (config.identities/config.contacts family codes), the
   data-model read-claim sentence still present (immutable HTML), registry protocol state
   (step-4 link rule present; tiers recorded, flags not), `mustShell` gone (salvage
   revert), i18n.js `cookieLang()` fix LANDED (split brain closed), lessons.md has
   PIPESTATUS but not pkill, store.ping/close/blob.probe family pins ABSENT, error-contract
   has no dashboard rows.
3. **ANNOTATE + ARCHIVE — 11 files, 690 inline verdicts, all with banner +
   check-rows "complete" + `git mv`** (sweep total now 15/34; archived dir 77→92 .md):
   - 2026-09-30 remainder (7): typography 12:08 (79 verdicts incl. 4 manual
     split-brain-table rows via `any:` keys), review-series 12:46 (61), dedup-t2 12:59
     (60), fax-paperless 13:09 (59), dedup-t1 13:35 (50), caddy-front 15:49 (41),
     setup-adoption 15:49 (93).
   - 2026-10-01 (4): visual-verification-loop 01:21 (70), post-v2-8-0-full-execution
     01:45 (33, incl. 2 `any:` keys for text-first b-table cells), samber-do 02:12 (91),
     02:14 self-review (53).
     Every spec built mechanically (`--emit-keys`), `--verify` first with `rc=$?` capture
     (no pipes), eyeballed match samples before every apply.
4. **Living-doc re-routing done same-breath as the verdicts citing it:**
   - ROADMAP typography stragglers bullet corrected (tabular-nums struck SHIPPED with
     evidence; focus-visible added; citation now f 4/8/11/12/14/30/35).
   - TODO_LIST OWNER-calls row extended: dedup suppression-SCOPE ruling (12:59 g3 —
     unrouted until now).
   - docs/dedup-registry.md open-calls list: item 4 added (suppression scope, routed).
5. **The 2026-09-30 cohort is now FULLY archived** (all 9 same-day reports + the two
   audit files v3/v4 from the prior session's leg).

## b) PARTIALLY DONE

1. **The annotation/archive sweep — 19 of 34 files remain:** eleven 2026-10-01 reports
   (02-54 READ with verdicts derivable — ring-tone fix = T06 audio shipped+audio.test.mjs
   pinned; demo-call recipe + WebTransport verdict doc = cross-repo/ROADMAP legs; then
   03-52, 04-07, 05-26 ×2, 05-27, 05-56, 06-59, 17-27 v5, 20-12, 23-11) and the 8 closed
   2026-10-02 chain files (incl. completing the mustinvoke report's banner-only §f items
   1–50 — the skill's #1 failure mode if left). Evidence for every verdict is gathered;
   per-file pipeline is now ~3 tool calls (emit → paired spec + verify → apply/banner/
   check-rows/mv).
2. **Forward-debt I created (must clear next):** several annotations cite "retro [2.8.0]
   bullet added by the v6 sweep (2026-10-03)" — the bullet is NOT landed yet. It must
   land before the sweep closes or those verdicts lie (see g1 for the shape question).
3. **v5 archive + superseding banner** — deliberately waiting for its archive moment
   (last of the 2026-10-01 cohort, after the v6 report-of-record story is complete).

## c) NOT STARTED (deliberate or queued)

- The retro CHANGELOG bullet (blocked on g1 shape: [2.8.0] retro vs [Unreleased] note).
- Archive manifest (one line per archived file — owed for the 34-file bulk archive).
- Gates: archived-dir completeness grep, check-rows over all newly annotated files,
  link check over touched files, scoped `go test ./cmd/webphone ./internal/server`
  (docs-only precedent), `git ls-remote` end-state.
- The inline v6 health report (Accuracy/Fitness, visible math — format reference to be
  loaded first) + `/tmp/spec-*.tsv` sweep.
- Keep live: 12:57 close-out, 13:50 session-8, the 4 owner-gated plans, operative
  verdicts/briefings/announcements, the immutable HTML snapshot, this report + the 00:35
  paused report (both v6 session records).

## d) TOTALLY FUCKED UP (all caught pre-write or fixed within a minute)

1. **Banner-less archive window (~1 min):** the typography file was `git mv`-ed BEFORE
   its banner landed — two anchor mismatches (a guessed `**`-prefix; then the trailing
   `**` artifact my earlier view showed had silently vanished from the file — git says
   it was already gone at HEAD~1, so my anchor was built from a stale read). Fixed by
   inserting into the archived path with a structural anchor (`\n---\n\n## Self-Review`).
   Lesson: banners use structural anchors; land them BEFORE the `git mv`.
2. **Miscounted verdict lists twice** (60-vs-59 fax; 93-vs-92 setup) and **transcribed
   line numbers wrong twice** (samber-do f32–f45 drifted by 2; review-series f15/f16 off
   by one) — every one caught by my length asserts, emit-key stderr, or `--verify` gate
   BEFORE any write; fixed by switching to mechanical `grep -n` line maps for tail
   sections instead of eyeballing view output.
3. **One stray token in a spec line** (`values` after a tab in the split-brain row) —
   caught by my own tab-count check pre-verify.
4. **A failed assert left a spec file unwritten** while the verify ran against a stale
   path — surfaced as "no key lines"; rebuilt cleanly. No file was touched by it.

## e) WHAT WE SHOULD IMPROVE

1. **Land cited living-doc edits BEFORE or WITH the annotations that cite them** — the
   CHANGELOG bullet debt (b2) is exactly this class; the ROADMAP/TODO/registry edits
   were done same-breath and set the right pattern.
2. **Mechanical line maps always** (`grep -n -E '^[0-9]+\. '`) — my view-transcription
   drifts on 300-line files; grep never does.
3. **Structural banner anchors** (`\n---\n\n## Heading`), never prose-line tails.
4. **The per-file pipeline is now fast and safe** (emit → python-paired verdicts with
   length asserts → verify with rc capture → apply → banner → check-rows → mv): keep it
   exactly; do not compress the verify eyeball out of it.
5. **`any:` keys are the sanctioned tool** for text-first table cells (split-brains
   table, b-table rows) — cleaner than manual strikes and still atomic.

## f) Up to 50 things to get done next (execution order)

1. Land the retro typography bullet (CHANGELOG — shape per g1; clears the b2 debt).
2. Annotate+archive 02-54 (verdicts derived; ring-tone=T06 shipped, recipe/verdict-doc
   = cross-repo legs).
   3.–12. Annotate+archive the other ten 2026-10-01 reports (03-52, 04-07, 05-26 island,
   05-26 nix, 05-27, 05-56, 06-59, 20-12, 23-11 — then 17-27 v5 LAST with its
   superseding v6 banner).
   13.–20. Annotate+archive the 8 closed 2026-10-02 chain files (sessions 3–6, mustinvoke
   §f completion, session 7, boot-contract, t21-t22).
3. Write the archive manifest (34 lines, classification + deciding reason) — appended
   as a completion addendum to the 00:35 v6 report or the archive-dir README.
4. Gates: `grep -rLn '~~' docs/status/archived/ docs/planning/archived/` = empty;
   check-rows over every newly annotated file; link check; scoped
   `nix develop -c go test -count=1 ./cmd/webphone ./internal/server`.
5. `git ls-remote origin main` end-state verify (daemon push lag).
6. Load health-report-format.md, deliver the inline v6 health report (Accuracy/Fitness,
   per-doc findings, visible math).
7. Sweep `/tmp/spec-*.tsv` session scratch.
8. Re-read TODO_LIST/ROADMAP once after the sweep closes (harvest-parity check: nothing
   the annotations routed is missing from its named home).

## g) Questions I can NOT figure out myself

1. **The retro typography CHANGELOG bullet — which shape?** My annotations already cite
   "retro [2.8.0] bullet added by the v6 sweep". Options: (a) a `### Changed` bullet
   under `[2.8.0] - 2026-09-30` documenting the shipped typography train (where a reader
   looks; technically edits a released section — append-only in spirit, additive in
   practice), or (b) an `[Unreleased]` Documentation note ("the 2026-09-30 typography
   train shipped without an entry; its content: …"). I recommend (a); confirm or pick
   (b) — the bullet lands either way as the sweep's next act.
2. **Same-day 2026-10-02 archiving — confirmed?** Your "keep going" read as assent to
   q1 of the paused report, but before I touch the 8-file closed chain (incl. completing
   the mustinvoke report's bare §f items): archive them this sweep, or hold the whole
   10-02 cohort one more sitting while you may still be reading the chain?
3. **The 4 owner-gated planning docs** (error-excellence, 20-year-durability, 04-29
   pareto, 13-33 stack-adoption): ratify "routed-as-resolved → archive" now (OWNER-calls
   row) so the next sweep archives all four in one pass — or keep waiting for the
   sitting? (Carried from the paused report §g3; no action taken either way.)

---

_Point-in-time snapshot — annotate, never rewrite. The auto-commit daemon owns the
commits (verified sweeping: 6-file heuristic commits through the session). SWEEP
COMPLETED 2026-10-03 03:29: all 19 remaining files annotated + archived (34/34;
manifest in the 00:35 report); gates clean — closure report:
`2026-10-03_03-29_docs-health-v6-closed-19-files-34-of-34-complete.md`._
