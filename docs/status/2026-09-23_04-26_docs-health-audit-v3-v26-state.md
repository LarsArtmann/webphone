# Status — docs-health AUDIT v3 over the v2.6.0 state: living docs truthful, 21 files archived, gates green

_2026-09-23 04:26 CEST. Session mandate: "View ALL `**/2026-0*` files;
execute the docs-health skill PROPERLY; the six living docs SUPERB;
archive FULLY done and UPDATED (inline strikethrough) .md files."
Documentation-only: zero product code touched. Skills loaded:
`docs-health` (SKILL.md + all 7 references) before any action;
precedent calibrated against the two accepted sweeps (2026-09-19 08:33,
2026-09-22 20:19)._

## Self-review (the three questions)

**What did I forget?**

1. **Annotate-verify-THEN-archive, not archive-then-annotate.** I
   `git mv`'d the ux-raw-ideas plan into `docs/planning/archived/`
   with a filled verdict but ZERO strikethroughs — the completeness
   gate (`grep -rLn '~~'`) failed on exactly that file. Fixed by
   striking its 15 fine-plan rows; the gate is green now. The order
   error is mine, not the tooling's.
2. **The whole-row strike rule for manual table edits.** check-rows
   caught THREE PARTIAL rows I created (16-29 c-table, 16-48 c-table,
   21-08 b-table): my manual strikes hit only the Item cell, leaving
   the Note cell clean — the exact F12.3 planted-miss class the skill
   documents. All repaired to uniform strikes.
3. **Two manual-strike clause drops.** In 16-52's c-section I dropped
   the "T25 (retention job)" clause from the struck text; in 19-11's
   c12 I dropped "plan v1.1 (CRM + webphone on different hosts)".
   Both violate the never-replace-text rule; both caught by my own
   re-read and repaired. The annotate tools do not have this bug
   class — my manual edits where the tools' shapes didn't fit did.
4. **Marked-done-before-verify once.** I struck 01-12 f17 / 03-01
   f20 ("confirm no pre-2026-09-18 dup registers linger") as done and
   only verified the claim in the LATER gate phase (it checked out:
   only session reports mention art-dupl; no register-shaped doc
   exists). Verify-before-claim is the rule; I inverted it once.
5. **Two edit refusals from bash reads** (CHANGELOG, 16-48 post-tool
   write) — I read files via `sed`/`cat` in the inventory phase and
   the edit tracker rightly refused. The repo's own lessons say
   View-tool-first; I paid the round-trip anyway.
6. The HTML/D2/SVG snapshots were triaged by name/class only, never
   content-viewed — accepted by both prior sweeps (immutable
   snapshots), disclosed here for completeness.

**What could you have done better?**

- The 02-47 verification: I proved the chained release retry NEVER
  FIRED (log gone, no gh release, origin == HEAD) but only NOW
  checked whether a wait-for-quiet process is still ARMED (pgrep:
  none found — but that check belonged in the verification step, not
  the report-writing step).
- Scoped quality gate: `go test ./cmd/webphone ./internal/server` +
  codespell + marker/link gates, not buildflow/flake-check — a
  docs-only-delta call on a host at load 58, matching the accepted
  precedent, but it is a scope cut I chose, not a gate verdict.
- The TODO sweep-log paragraph first shipped ~16 lines, violating the
  ≤8-line constraint yesterday's audit recorded (f31/f48); caught in
  final verification and trimmed.
- AGENTS.md grew 407 → 411 lines (my release-state paragraph); no
  enforced cap (full gates passed at 407), but the drift direction is
  wrong for that file.

**What could you still improve?** → (e)/(f). Also: the annotation
hash bar is still train-level `v` markers (same disclosed downgrade
as both prior sweeps) — ratify or retire formally; and the
DOMAIN_LANGUAGE.md gap scored −1 Fitness and I deliberately did not
fabricate a thin glossary at 04:00 (question g2).

**Did you lie?** No. Every verdict marker cites either a real commit
hash from the reports' verified ledger or an explicitly-labeled
`done (<evidence>)` string; the one claim I struck before verifying
(dup registers) was verified before the session ended and would have
been retracted had it failed. Numbers in this report are from actual
command outputs in this session.

---

## a) FULLY DONE (verified this session)

| #  | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Evidence                                                                        |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| 1  | Skill discipline: docs-health SKILL.md + all 7 references loaded BEFORE any action; AUDIT mode (BUILD+HARVEST+VERIFY+ANNOTATE+ARCHIVE); both accepted sweeps read as the bar                                                                                                                                                                                            | session transcript                                                              |
| 2  | All 96 `**/2026-0*` artifacts viewed/triaged: 21 status + 12 planning `.md` read in full; verdict/decision records (csrf, webhook-idem, openapi, pwa, session-persistence, sip.js eval, oob, hub-fanout, coverage, briefing, announcements) = SKIP per skill policy; 11 HTML/D2/SVG = LEAVE (immutable)                                                                                                  | glob + ls inventories                                                           |
| 3  | VERIFY sweep against code, all by direct command: v2.6.0 tag `807ca0c` on origin but **NO gh release object + the chained release retry NEVER FIRED** (`/tmp/release26d.log` gone — this fact was documented nowhere); stack pins webphone `7503561`; origin == HEAD at sweep start; `SearchThreads`/`crmNumbers`/`fullStamp`/`apiContactSaved`/`applyStatusWebhook`/`recordCallIdem`/`typeahead.js`/`audioout.js` all present in code; smoke prints 40+4 (AGENTS said 38); CHANGELOG [2.6.0] MISSING the four A–D ux features (only E+F folded) | git ls-remote/rev-parse, gh release view, greps                                 |
| 4  | CHANGELOG: [2.6.0] date corrected 2026-09-22 → 2026-09-23 (the tag's real date); four Added bullets added (hover stamps A, jump chip B, missed-call badge C, dial typeahead D) with test-evidence cites                                                                                                                                                                                                                                | CHANGELOG.md                                                                    |
| 5  | FEATURES: +4 rows (Dial typeahead → Calling; Hover timestamps + Jump-to-latest chip → Messaging; Missed-call badge → Awareness)                                                                                                                                                                                                                                                                                                     | FEATURES.md                                                                     |
| 6  | AGENTS.md: release-state paragraph rewritten to the v2.6.0 truth (tag `807ca0c`, open TAIL enumerated, stack pin `7503561`); smoke count 38 → 40                                                                                                                                                                                                                                                             | AGENTS.md                                                                       |
| 7  | TODO_LIST rebuilt (13 → 17 evidence-cited rows): release row rewritten around the verified TAIL (gh release + stack E2E ×2 + aarch64 + `--expect-version` + pbx-artmann relock #4); dedup-pins row DELETED (shipped 21:12); CRM row shrunk to (e)–(h); +7 harvested rows (helper micro-tests, `-race` full-package, error-contract cross-check + rationale registry, runbook hardening, LSP nolint fix-or-declare, full-code-review of the interleaved day); watches row grew the monthly erraudit re-measure | TODO_LIST.md (full rewrite)                                                     |
| 8  | ROADMAP: composer cluster rewritten (six shipped ideas marked, follow-on fuel added: unicode search ADR, search depth, typeahead boost, transcription, `ValidOutboundStatus`); Open questions +9 (release load-gate, pin-vs-tag, host-contention protocol, art-dupl `-t 3` baseline, webhook 400 + idem-key rename, helper-test bar, missed-call semantics, `?q=` semantics, `Must*` panic policy); infra ask extended (pre-sweep build gate + release-boundary tree assert)                           | ROADMAP.md                                                                      |
| 9  | UX plan verdict filled (was PENDING): SHIPPED 6/6 + the stub-evidence E2E caveat + the ASCII-only-search limit + deferred ideas routed                                                                                                                                                                                                                                                                                               | `docs/planning/archived/2026-09-22_21-05_*`                                     |
| 10 | README: capability table gained the six user-facing surfaces (search, typeahead, state chip, missed badge, audio picker)                                                                                                                                                                                                                                                                                                             | README.md                                                                       |
| 11 | ANNOTATE: ~330 inline verdicts across 22 files — dated `CLOSED/UPDATED` banners on all 22 (inline-correcting headlines per the skill's placement rule) + per-item strikes via `annotate-rows.py`/`annotate-prose.py` (dry-run first on every new file shape; shape-checked read-backs)                                                                                                                                                  | grep marker counts; tool outputs                                                |
| 12 | ARCHIVE: `git mv` of 20 status reports + 1 planning doc into `archived/`; every stale reference rewritten (TODO_LIST ×6 refs + 5 in-file refs); only the operative release-tail report (02:47) stays live — the healthy shape                                                                                                                                                                                                          | git status R-count 21; ref-grep clean                                           |
| 13 | Gates, all green: completeness `grep -rLn '~~'` over both archived dirs = EMPTY; zero PARTIAL rows after repair (check-rows); dup-register claim verified (no register-shaped doc exists pre-2026-09-18); codespell 0 findings; `go test ./cmd/webphone ./internal/server` ok ×2 packages; TODO table uniform (19 lines × 7 fields)                                                                                                                                                             | command outputs in transcript                                                   |
| 14 | Health report delivered inline: Accuracy 6.25 → 10 (9 findings, all fixed), Fitness 9/10 (DOMAIN_LANGUAGE gap), visible math per the format spec                                                                                                                                                                                                                                                                                   | conversation                                                                    |
| 15 | End state: tree clean (daemon swept everything incl. the 67-move archive batch, commit `142f6c1`); no processes booted by me                                                                                                                                                                                                                                                                                                        | git status; pgrep                                                               |

## b) PARTIALLY DONE

| # | What works | What remains |
| - | ---------- | ------------ |
| 1 | Quality gate | Scoped (drift + DOM-contract + codespell + marker/link gates) — buildflow full, `nix flake check`, island suite NOT run (docs-only delta; host load 58 from parallel sessions) |
| 2 | The v2.6.0 TAIL verification | Proven never-fired + no gh release (routed to the TODO release row); the TAIL itself (stack E2E ×2, gh release, aarch64 re-verify, `--expect-version`, pbx relock #4) NOT executed — owner/next-session work |
| 3 | Push state | All work committed by the daemon; the daemon's PUSHER lags again (HEAD `0023e3b` vs origin `0cb6d2c`) — known mode, runbook covers it, not a docs session's call to push main |
| 4 | Annotation hash depth | Train-level `v` markers throughout (same disclosed bar as both accepted sweeps); per-item `git log -S` hashing would roughly double sweep cost — ratify-or-retire still open (g3) |

## c) NOT STARTED

- `docs/DOMAIN_LANGUAGE.md` — the docs-health must-have table lists it; it
  has never existed here; I flagged it (−1 Fitness) instead of
  fabricating a thin glossary at 04:00. Decision = g2.
- Everything on the rebuilt TODO_LIST (release TAIL, owner console,
  micro-tests, `-race`, error-contract cross-check, runbook hardening,
  LSP nolint, full-code-review) — routed, deliberately untouched: this
  session was documentation-only by mandate.
- Restyling the 3 prior-sweep archived files (`16-37`, `21-38`, `22-55`)
  to the strict check-rows bar — still the accepted-baseline question
  from yesterday's audit (its g3, now mine).

## d) TOTALLY FUCKED UP (this session's own goals, no spin)

1. **Archived a file with zero strikethroughs** (the 21-05 plan) —
   the completeness gate caught it; the correct order
   (annotate → verify → archive) was in the skill all along.
2. **Manufactured three PARTIAL table rows** with manual first-cell-only
   strikes — the exact planted-miss class this skill's tooling exists
   to catch; check-rows caught what I should not have created.
3. **Two clause-dropping manual strikes** (16-52 T25, 19-11 plan-v1.1) —
   content loss inside a "never replace text" operation, both
   self-caught on re-read and restored inside the strikes.
4. **One claim struck before it was verified** (dup-register sweep) —
   verified later and it held, but the order was wrong and would have
   been a false marker had the grep come back otherwise.
5. Nothing shipped broken; no product code touched; no data loss; all
   gate failures were caught by the gates themselves and repaired
   before the session ended.

## e) WHAT WE SHOULD IMPROVE (process, not docs)

1. **Order discipline for archives**: marker-completeness gate BEFORE
   `git mv`, every file, no exceptions — this session's one gate
   failure was pure ordering.
2. **Manual table strikes are the bug farm**: when a table's shape
   defeats the annotate tools, strike the ENTIRE row (all cells) or
   don't strike at all — three of my four d-items came from
   half-row strikes and clause-drops.
3. **Verify-then-strike, never strike-then-verify** — the marker is a
   claim; claims get evidence first.
4. **pgrep the armed job when verifying "never fired"** — a missing
   log proves the run, not the absence of a waiting process (I did
   the pgrep, just too late in the sequence).
5. **View-tool-only reads in this repo**: the mtime/edit-tracker
   refusals are cheap lessons; stop re-learning them per session.
6. **AGENTS.md drift direction**: 411 lines and growing; the next
   content add should pay for itself by removing a line (the
   file's own T16a discipline).

## f) NEXT — honest backlog, no padding (the live set IS the rebuilt TODO_LIST; restated in execution order)

1. **v2.6.0 release TAIL** (TODO row 1): re-run release.sh steps 6–9 on
   a quiet host → gh release → `--expect-version 2.6.0` → pbx-artmann
   relock #4 → closing ls-remote sweep.
2. Check the daemon pusher (HEAD ahead of origin again).
3. Owner: deploy the released chain to prod (command in TODO row 2) +
   post-deploy probes (row 3).
4. Owner: SMS-lane journalctl (row 4).
5. Quiet-machine full gate over the post-tag tree (buildflow full +
   flake check + island suite) — doubles as the art-dupl post-tag
   verification (b2 of the 03:01 report).
6. Helper micro-tests for the ten one-home helpers + coverage
   inventory of the two JSON contact handlers (TODO row 5).
7. `go test -race ./internal/server/...` full-package pass (row 6).
8. Error-contract cross-check vs `applyStatusWebhook` + JSON-204
   contract + the AGENTS acceptance-rationale registry line (row 7).
9. Release-runbook hardening once the owner picks the mechanism
   (load-gate vs E2E-retry + flake heuristic + logging rule + the
   coordination-hazard note) (row 8).
10. golangci-lint LSP nolint fix-or-declare + full-tree lint re-run
    (row 9).
11. Full-code-review over the interleaved 2026-09-22 day (row 10).
12. Theme-knob verification: FOUC pair, de-native review, composer
    screenshots (row 11).
13. Send-failure trains C/E/F + fax guard + stack-runbook CSP cross-doc
    (row 12).
14. CRM follow-ups (e)–(h): stack wiring, runbook cross-doc,
    restore-drill, integration test (row 13).
15. OWNER-calls batch session (~24 decisions now — the row carries the
    full list; briefing doc + ROADMAP questions ready) (row 14).
16. T26b stack half; announcements (v2.6.0 draft owed after the tail);
    standing watches + monthly erraudit (2026-10-22) (rows 15–17).
17. Decide DOMAIN_LANGUAGE.md (g2) and the annotation hash bar (g3).

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Is anything still armed for the v2.6.0 release TAIL?** The
   chained wait-for-quiet job's log is gone and my pgrep finds no
   live process — but the 02:47 session launched it from a shell I
   do not own. If anything of yours is still waiting to fire
   `release.sh 2.6.0`, this session's TODO rewrite assumed manual
   re-execution; tell me which truth holds.
2. **DOMAIN_LANGUAGE.md**: create a real domain glossary (the
   docs-health model lists it as a must-have; the domain terms
   currently live in README "Integration contracts" + AGENTS), or
   record its deliberate absence as accepted for this repo?
3. **The annotation hash bar** (carried from yesterday's audit, now
   mine): ratify train-level `v` markers ("done (shipped v2.6.0)")
   as the standing bar for archive sweeps, or order per-item
   `git log -S` hashes at roughly double sweep cost?

---

_Point-in-time snapshot — annotate, never rewrite, when bringing
current later. The auto-commit daemon owns the sweep (this report
included)._
