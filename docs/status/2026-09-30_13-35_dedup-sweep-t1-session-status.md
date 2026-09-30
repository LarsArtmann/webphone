# Status Report — 2026-09-30 13:35 CEST — `-t 1` Dedup Sweep Session + Self-Review

**Session scope:** one task — owner pasted an `art-dupl -t 1` report and asked
"anything worth deduplicating?". Judged per `docs/dedup-registry.md` (the ONE
acceptance home) and the `deduplicate-code` skill. This report covers ONLY this
session's run and what it surfaced, per owner instruction. Sections follow the
owner's a–g shape; `.md` format is an explicit owner override of the skill's HTML
default (second time — see question 3).

**Evidence base:** owner paste; my re-run at HEAD (identical: 560 detected, 42
shown, 395 non-actionable, 123 filtered suppressed); HEAD during session
`97d9c79` → `9209a2d` (daemon). No production code changed this session — the
quality gates (buildflow, go test, island tests) are N/A by construction; only
docs moved.

**One-line verdict:** the codebase is CLEAN at `-t 1` (deepest sweep yet) —
42/42 shown groups attributed, 0 harmful, 0 extractions; 4 new ACCEPT rulings
recorded; the session's own process had 2 gaps, both found and closed below.

---

## a) FULLY DONE

1. **Protocol steps 1–2 (skill + registry first).** Loaded `deduplicate-code`
   SKILL.md and read `docs/dedup-registry.md` BEFORE judging anything. No group
   was re-litigated against a standing ruling.
2. **Attribution against MY OWN run, never the paste.** Re-ran
   `art-dupl --sort total-tokens -t 1 --suggest-generics --timing --rich-text`
   (dropped `--type-aware` per the tool's own precedence warning). Result:
   **identical to the paste** — same 42 shown groups, same windows. The paste's
   flags were owner muscle-memory; results unaffected.
3. **Every occurrence site re-read at HEAD, not one per group.** All 5
   `data-dial` buttons (contacts ×2, history, messages thread header, voicemail),
   both `wp-dir-chip` chips + the group's avatar third site, both
   delivery-verdict guards + the `FaxTransmitted/FaxFailed` twin, both
   `wp-danger` delete affordances, all tab-view submit buttons, both composer
   textareas, and the SingleStatusDelivered/Failed view references.
4. **4 new standing ruling rows, all ACCEPT** (registry:49–53):
   `data-dial` ×5 (shell.js contract; 3 params for a 1-line body),
   `wp-dir-chip` pair (per-domain predicate + glyph/label pairing),
   delivery-verdict guard pair (two validation LAYERS: HTTP 400 on untrusted
   wire input vs errorfamily Rejection on typed domain values — defense-in-depth),
   `wp-danger` delete-affordance pair (different endpoints/keys, 3 params for
   8 lines). Plus the composer/submit row widened to the cross-tab submit family.
5. **Protocol steps 3–4.** Mid-session status doc written
   (docs/status/2026-09-30_13-17_dedup-sweep-t1-status.md) and the sweep-log
   line appended (registry:73).
6. **Concurrent-edit race handled per AGENTS.md.** The registry edit hit
   "file modified since read" (daemon commit between my 13:09 read and 13:15
   edit); re-read, verified content unchanged, re-applied cleanly.
7. **Sibling-site greps for all 4 new rulings** (run during THIS report — see
   gap b1): `data-dial` = 4 templ files + shell.js handler + a comment-only hit
   in helpers.go + 7 test files; `wp-dir-chip` = exactly the 2 ruled sites +
   app.css; `wp-danger` = exactly the 2 ruled sites; `StatusDelivered/Failed` =
   the 2 production guard sites + domain defs + tests + view badges;
   submit census = 10 (8 tab-view, ruled; 2 phone.templ island).

## b) PARTIALLY DONE

1. **The 13:17 sweep-log line omits the sibling-grep evidence.** The 12:59
   `-t 2` sweep recorded its greps in the log line; mine didn't — I flagged
   this during report writing and ran all four greps (results in a7), but the
   13:17 log line stands as written (append-only). Evidence lives here and in
   the ruling rows instead. Effort to fully reconcile: S (one addendum line if
   the owner wants log symmetry).
2. **`messaging.verdict` Rejection branch is UNTESTED.** Discovered while
   verifying the verdict-guard ruling: the happy path is tested through the
   webhook (`webhooks_test.go:250` exercises "delivered" over HTTP), and
   `internal/messaging/family_test.go` exists, but NO test calls
   `DeliveryReceipt` with a non-verdict status — the guard at
   `internal/messaging/service.go:273-275` is reachable only by direct service
   callers and nothing pins it. Effort S (one table case). First step: check
   whether family_test.go already covers the family classification.
3. **"42/42 attributed" = shown groups only.** The 395 non-actionable and 123
   filtered-suppressed groups are counts, never enumerated — the tool prints
   no breakdown of what suppression caught. Attribution is complete for the
   actionable surface and blind beneath it. Not fixable in-session (tool
   limitation); see e5/f10.
4. **Submit-button census correction.** My widened row initially implied 7
   sites; the census shows 10 (8 tab-view — all covered by the composer/send/
   family clauses — plus 2 phone.templ island submits that are a DIFFERENT
   family). Fixed in this session: the row now states island submits ride the
   `data-i18n` island ruling instead.
5. **AGENTS.md's "`-t 3` is the working baseline" line is now ambiguous** —
   today saw `-t 2` (12:59) and `-t 1` (13:17) owner-requested sweeps. The
   line is pending the owner's ratification (open call #1), not mine to rewrite.

## c) NOT STARTED

1. **Open owner call #1 — baseline ratification.** `-t 3` ritual vs `-t 2`
   deep sweeps; today's `-t 1` shows the noise floor (15 single-token T()
   clones, script/link/head trivia — 9s runtime, low signal). Waiting on owner.
2. **Open owner call #2 — `settingsRow` build-or-retire.** Surfaced in every
   sweep to date including this one (the 7-token dt/dd group). Waiting on owner.
3. **Open owner call #3 — ratify the registry as the ONE acceptance home.**
   Proposed 03-01, re-raised twice, still unratified.
4. **HARVEST of section (f) into TODO_LIST.md.** Per the status-report skill,
   the f-list dies entombed in this timestamped file unless harvested. Not
   started (report written moments ago).
5. **Canonical dedup-flags wrapper.** Owner's command carried
   `--type-aware --suggest-generics` (the tool warns the first is ignored);
   today's two sweeps used different thresholds. A `scripts/dedup-sweep.sh`
   pinning canonical flags would end the drift — NOT started, LOW priority,
   arguably pointless until (c1) ratifies a threshold.
6. **Suppression audit.** The 123-filtered bucket has never been inspected;
   verifying no false negatives needs a manual flag flip and a read-through.
   NOT started, M effort.

## d) TOTALLY FUCKED UP

**Nothing.** No code was harmed: the session produced only docs
(registry + 2 status files). No build, test, or lint surface touched. The two
nearest misses, for the record:

- The registry stale-read race (a6) — cost one tool round-trip, zero damage,
  recovered by the AGENTS.md concurrent-session rule.
- The missing sibling-greps (b1) — would have left 4 new rulings without
  site-census evidence; caught during this report's self-review and closed.

## e) WHAT WE SHOULD IMPROVE

1. **Make sibling-greps a WRITTEN protocol step, not tribal memory.** The gap
   (b1) happened because the 12:59 sweep's grep habit lives in a log line, not
   in the registry's protocol section. Fix: insert "2.5. Sibling-grep every NEW
   ruling's shape before recording it" into the protocol list. S effort.
2. **Re-read daemon-active shared files in the SAME breath as the edit.** Read
   at 13:09, edit at 13:15 = stale. The concurrent-session rule exists; the
   practical habit should be re-read-then-edit back-to-back for
   registry/i18n/pages/flake files. Process habit, zero code.
3. **`-t 1` is mostly noise.** 15 of 42 groups are single-token fragments
   (T() calls, `" "` literals, script tags). Unless the owner wants `-t 1` as
   ritual, deep sweeps should floor at `-t 2`; `-t 1` fits targeted questions
   ("is THIS shape duplicated"), not routine sweeps.
4. **Same-day sweep docs are proliferating** (12:59, 13:17, 13:35 = 3 files).
   The log lines cross-link them, but a future HARVEST must not triple-count
   shared items (owner calls appear in all three). Consider one rolling
   "dedup-sweeps" doc per day. LOW.
5. **Suppression opacity.** "123 filtered suppressed" is unauditable from the
   report alone. Either the tool grows an explain mode (f10, upstream) or each
   deep sweep flips the filter once and records "suppressed bucket clean"
   in the log. The 12:59 and 13:17 sweeps both skipped this.
6. **Sweep status docs should inline their sibling-grep section** (this report
   does; 13:17 didn't). Cheap, and it is exactly the evidence that makes a
   ruling trustworthy six months later.
7. **Flag the .md-vs-HTML divergence once, decisively.** Two sweep reports in
   a row are `.md` by owner instruction while the skill's canonical format is
   HTML. See question 3 — one answer retires this note forever.

## f) NEXT TASKS (session-scoped, impact-sorted; "up to 50" = 22 real ones — padding to 50 would be ROADMAP spam, per the skill's own warning)

| #  | Task                                                                                                                     | Impact | Effort | Category       |
| -- | ------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | -------------- |
| 1  | Add direct unit test: `DeliveryReceipt` non-verdict status → Rejection `messaging.verdict` (check family_test.go first)  | High   | S      | Bug (test gap) |
| 2  | Registry protocol: insert mandatory sibling-grep step (2.5) for new rulings                                              | High   | S      | Quality        |
| 3  | Owner call: ratify sweep baseline threshold (`-t 3` ritual / `-t 2` deep / `-t 1` targeted-only)                         | High   | S      | Decision       |
| 4  | HARVEST this f-list into TODO_LIST.md (docs-health)                                                                      | High   | S      | Process        |
| 5  | Owner call: `settingsRow` — build the component or accept dt/dd permanently                                              | Medium | M      | Decision       |
| 6  | Owner call: ratify registry as the ONE acceptance home (3rd request)                                                     | Medium | S      | Decision       |
| 7  | Verify `data-dial` attribute is in docs/dom-contract.md; add it if absent (wire contract currently only in AGENTS prose) | Medium | S      | Documentation  |
| 8  | Update AGENTS.md dedup line after baseline ratification (currently says `-t 3` pending)                                  | Medium | S      | Documentation  |
| 9  | One addendum sweep-log line recording the a7 sibling-greps (log symmetry with 12:59)                                     | Low    | S      | Documentation  |
| 10 | Upstream (verify-first): art-dupl flag to enumerate/break down the filtered-suppressed bucket                            | Low    | M      | Quality        |
| 11 | One-time suppression audit: flip filter, read the 123, record "clean" in the sweep log                                   | Medium | M      | Quality        |
| 12 | `scripts/dedup-sweep.sh` canonical-flags wrapper (post-ratification only)                                                | Low    | S      | Cleanup        |
| 13 | Cross-link the three same-day sweep docs (related/supersedes lines) so HARVEST can't triple-count                        | Low    | S      | Documentation  |
| 14 | Record art-dupl version in each sweep-log line (reproducibility)                                                         | Low    | S      | Process        |
| 15 | Decide sweep-report format policy (.md vs HTML) — see question 3                                                         | Low    | S      | Process        |
| 16 | `DeliveryReceipt`: decide whether a provider failure with EMPTY errMsg should still `slog.Warn` (currently silent)       | Low    | S      | Quality        |
| 17 | Registry hygiene: refresh "sites at last sighting" counts on pre-09-27 rows (some cite stale shapes)                     | Low    | S      | Documentation  |
| 18 | Fold the ratification outcome into the registry protocol text ("`-t 2` deep sweeps by owner request" line)               | Low    | S      | Documentation  |
| 19 | Consider a rolling per-day sweep doc instead of per-sweep files (e5)                                                     | Low    | S      | Process        |
| 20 | Re-run `art-dupl -t 2` after tasks 1–2 land to confirm the group set is unchanged (regression ritual)                    | Low    | S      | Quality        |
| 21 | Grep-audit remaining unruled shapes from the -t 1 report (option rows, hidden inputs) for one-line registry notes        | Low    | S      | Documentation  |
| 22 | Close the loop: annotate 12:59 + 13:17 docs with "superseded by rulings in registry, not by this report" pointer         | Low    | S      | Documentation  |

Items 3/5/6 are the owner calls — everything else is executable without input.

## g) QUESTIONS (3, none answerable from the repo)

1. **Baseline:** with today's `-t 2` AND `-t 1` sweeps both clean, do you want
   to ratify `-t 2` as the permanent deep-sweep floor ( `-t 3` ritual stays),
   retiring `-t 1` to targeted questions only? This unblocks f3/f8/f12/f18.
2. **`settingsRow`:** the settings dt/dd group has surfaced in every sweep
   (09-18 → today). Build the `settingsRow` component, or accept the rows
   permanently and close owner call #2?
3. **Report format:** the status-report skill's canonical output is a styled
   HTML dashboard; you've now requested `.md` for two sweep reports running.
   Bless `.md` for the sweep/status series (one-line skill note), or should
   future reports go back to HTML?
