# Status — docs-health v8 sweep: living docs re-verified, 68 reports archived, the battery adopted (2026-11-06)

**Snapshot-rot header: every claim below was verified 2026-11-06 ~11:00–12:30 CET at
HEAD `54c8bb17` (working tree: this sweep's own changes).** Documentation-only session:
no product code touched, no push.

## TL;DR

The owner's order — "View ALL `**/2026-0*` files, execute docs-health SUPERBLY,
living docs must be superb, archive FULLY done and UPDATED .md files" — executed as a
full AUDIT (BUILD + HARVEST + VERIFY + ANNOTATE) over the September cohort AND the
report backlog, with the `docs-gates` battery adopted so the next sweep is one command
instead of a re-derived ritual.

| Axis                              | Before                                 | After                                                              |
| --------------------------------- | -------------------------------------- | ------------------------------------------------------------------ |
| `docs-gates` battery              | never run here (no config, no wrapper) | **GREEN**, ratcheted: `scripts/docs-gates.sh` + `.docs-gates.json` |
| annotate findings                 | 1002 (20 outside the archive)          | **205**, 100 % pre-canon ARCHIVED (ratchet 205)                    |
| audit-all unharvested §f items    | 346                                    | **330** (ratchet 330)                                              |
| unarchived status reports         | 136                                    | **68** (47 % of the corpus adjudicated and archived)               |
| archived status reports           | 134                                    | **179**                                                            |
| files matching `2026-0*` reviewed | 44                                     | 43 (1 archived, 5 materially updated, 0 lost)                      |
| false claims found in living docs | 3                                      | 0                                                                  |

## a) FULLY DONE (verified this session)

1. **The `2026-0*` cohort, file by file (44 files), classified with a reason each.**
   Every markdown file was READ and judged; every HTML/D2/SVG snapshot was checked for
   obligations (the freshness gate flags un-scanned HTML reports). Result: 1 ARCHIVED
   (`announcements/archived/2026-09-24_v2-7-0_drafts.md` — superseded before posting, its
   content shipped as v2.8.0), 5 KEEP-and-UPDATE (2 standing baselines refreshed with
   today's measurements, 4 executed plans banner-annotated with their verdict + the
   blocker, 1 owner command sheet refreshed against live prod), 38 KEEP (operative
   briefings/command sheets, the standing credentials re-review whose clock runs to
   2026-12-30, and immutable HTML/D2/SVG snapshots — non-markdown archiving has no
   precedent here, and these are the artifacts a blog/README cites by path).
2. **`docs-gates` battery adopted** (`scripts/docs-gates.sh`, `.docs-gates.json`): dprint
   fmt → the five-gate matrix → the table-damage gate (fixture-proven able to fail).
   Ran to a green matrix. AGENTS.md carries the doctrine (archive ritual, ledger-marker
   semantics, ratchet sinkhole rule).
3. **HARVEST.** The newest report (`2026-11-06_10-45`) had NO §f, so the freshness gate
   was passing vacuously; it now carries one (16 ranked items) and 16 rows landed in
   TODO_LIST § 2026-11-06 (push, PR #234, stack `caFile` pass-through, issue #9 criteria,
   `.gitignore` anchoring audit, `crm` health-criticality decision, findings-gate
   posture, quiet-host benchmark, store test train, CI wiring, the archive residue
   plan, `/context` contract snapshot, notes-tab nudge, templ-generate wrapper, the
   attended-transfer UI ruling, `reverseLines`).
4. **ANNOTATE + ARCHIVE — 68 status reports + 1 announcement draft.** Each passed BOTH
   row gates (`check-rows --strict`, `bare-rows`) BEFORE `git mv` (the 2026-10-15
   gate-before-move lesson held) and carries its own `> ARCHIVED 2026-11-06 …` banner.
   Classification + deciding reason per class in the manifest below.
5. **Corpus repair (`repair-mangle`)** — the mechanical damage classes (doubled strikes,
   collapsed empty struck cells, boundary quads) repaired across the archived corpus:
   ~250 rows in 20+ files, content untouched. 6 content-loss quads (a swallowed strike
   boundary, recoverable because the text survived) hand-repaired in
   `planning/archived/2026-09-19_20-01_SUPERB-honest-self-health-upstream-first-plan.md`.
   This alone dropped annotate findings from 950 to 205.
6. **Living docs re-verified against HEAD**, four real lies found and fixed:
   - FEATURES claimed **attended transfer FULLY_FUNCTIONAL** ("bridges the two
     established calls at the network") while this repo's own spike
     (`docs/research/2026-10-27_attended-transfer-spike.md`) says the bridge is
     UNVERIFIED — the stack has no Replaces-aware dialplan handling and no FreeSWITCH
     lab exists here. Status corrected to PARTIALLY_FUNCTIONAL; the UI-vs-ruling
     decision is now a TODO row.
   - README + FEATURES both SELLED that same attended transfer as a working feature
     ("blind + attended transfer"); both now name the client half and the missing
     `Replaces` dialplan support.
   - FEATURES named **go-health-dashboard v0.10.1**; the pin is v0.13.2 (go.mod).
   - ROADMAP's test-infra note still said **`internal/blob` has no direct tests** and
     `views` was 1.3 %; both superseded by measured numbers.
7. **Both standing baselines refreshed with TODAY's numbers** (they are living docs by
   their own rules, not snapshots):
   - `docs/reviews/2026-09-20_coverage-baseline.md` → a 2026-11-06 section (35 packages,
     per-package table, the two packages that moved backwards and why, island JS 22 → 289
     specs).
   - `docs/reviews/2026-09-18_hub-fanout-baseline.md` → a 2026-11-06 section that is
     HONESTLY INCONCLUSIVE: the re-run trigger fired (cqrs-htmx v4.16.0 + go-sse
     v0.7.0) but the host sat at load ~22/32 (foreign nix builds + two other repos'
     `docs-gates` replays) and the spreads were ~2×; the entry prints the numbers WITH
     their extremes and the instruction to re-run below load 8. No regression verdict
     is claimed.
8. **Four owner questions routed out of timestamped reports into ROADMAP § Open
   questions** (they had lived only in status files for weeks): coverage-with-teeth
   (2026-09-18 HTML §g), the CI-bar subset policy and the daemon-format coupling home
   (2026-10-06 §g-Q1/Q3), and the WCAG 1.4.11 `--border-strong` accept-vs-retune call
   (2026-10-04 §g.3).
9. **Rule-F damage eliminated from every live doc** (20 findings → 0 outside the
   archive). The interesting part is the CONFLICT discovered: dprint's markdown reflow
   EATS the space around an inline code span at a wrap point, i.e. the formatter
   inflicts exactly what rule F flags. Restoring the space is undone by the next
   `dprint fmt`; the durable fix is to reword so the code span is not at a wrap point
   (recorded in AGENTS as the corollary of the new command entry).
10. **`docs/status/INDEX.md` regenerated** (69 reports; `--check` green).
11. **Todo-list truth sweep**: the front matter now records THIS sweep, the September
    plans' ratification question was executed (the four owner-gated plans are
    annotated as EXECUTED/CLOSED/SUPERSEDED with their verdicts quoted, so the next
    reader does not have to re-derive it), and 16 new rows carry the harvest.

## b) PARTIALLY DONE — the archive residue (measured, not hand-waved)

68 reports remain live. Every one is blocked by the SAME thing: un-inventoried
table rows in its §a/§b sections (~1 600 rows), because those reports use a "Verdict"
COLUMN (`**DONE**`, `**PARTIAL**`) instead of the strike canon the gates read. They
are not lost work — they are un-adjudicated rows.

- Why not done now: striking 1 600 rows with a generic verdict would be the
  trophy-case anti-pattern the docs-health skill forbids (a verdict that could apply
  to ANY file adds nothing). Each row needs its own evidence, and the honest evidence
  is usually the report's own §a narrative.
- What IS done: the class is characterised, counted per file, and owned by TODO_LIST
  item 11 (per-file `status-items` strikes for §a/§b shipped tables, `r` routed
  verdicts for §f rows, then the gates BEFORE each move).
- Cadence is ratcheted (`max_unarchived` 68, `oldest_days` 45) so the residency cannot
  grow while the drain proceeds, and the ratchet can only move DOWN.

The pre-canon ARCHIVED corpus keeps 205 ratcheted findings (108 no-verdict strikes, 56
strike-crossing rows, 17 spacing, 13 missing-path citations, 7 nonexistent tag cites,
4 unresolvable hashes). Those are CLOSED HISTORY: rewriting their verdicts to buy a
green would be the Verschlimmbesserung the skill's case study documents.

## c) What the gates can and cannot see (findings worth carrying)

1. **A §f-less newest report makes freshness vacuously green.** The gate said "no §f
   section — nothing to check" and PASSED for the 10:45 report. Fixed by writing the
   §f; worth remembering that "PASS" can mean "did not look".
2. **`bare-rows` uses BASENAMES, not paths.** A script that greps for
   `docs/status/<file>:` silently matches nothing and reports a spotless corpus. This
   sweep was nearly fooled by exactly that (the numbers looked impossibly good);
   the tell was that a file known to carry 137 bare rows showed zero.
3. **The HTML-report warning is real**: `docs/status/2026-09-18_16-52_…html` carries
   §f obligations outside the `.md` scan. Its ten items were harvested by its own
   session; its ONE open question (coverage teeth) was unowned until today.
4. **A feature can be "FULLY_FUNCTIONAL" in FEATURES while its own spike says
   UNVERIFIED** — nothing cross-checks the two. The attended-transfer row had been
   wrong since 2026-10-27.
5. **`audit_max` is DF-brittle by construction**: a token counts as "generic" once it
   appears on ≥ 8 living-doc lines, so GROWING the living docs can flip an old §f
   item from quietly-matched to flagged without anything in the corpus changing.
   Observed live this sweep: 330 → 331 after a CHANGELOG paragraph. The pin is a
   snapshot, re-measured each sweep, never a claim about the corpus.
6. **Generated tables must format themselves.** `scripts/status-index.py` rendered
   cell padding it does not own, so every regen left the tree dirty until a
   formatter pass — a generator that hands the next `dprint check` a red nobody
   edited. It now runs dprint on its own output (best-effort: a missing dprint warns).
7. **The ratchets only bind through `docs-gates run .`** — a standalone verb
   (`docs-gates freshness . --audit-all`) ignores `.docs-gates.json` and exits 1 while
   the battery exits 0 on the same tree. Pass `--audit-max`/`--annotate-max` explicitly
   when running one gate by hand, or a reader concludes the corpus is red when it is
   merely at its pin.
8. **Ratchet values must be measured, never chosen.** The four pinned numbers came from
   running the battery on the repaired tree; picking round numbers would have hidden
   the 205/330 residue forever.

## d) Archive manifest (68 status reports + 1 draft — classification + deciding reason)

Every archived file carries its own `> ARCHIVED 2026-11-06 …` banner naming the reason;
the classes below are the shared criteria. Reading this table plus one banner answers
"why is this archived" for any of them.

| Class                                               | Count | Deciding reason                                                                                             | Files                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --------------------------------------------------- | ----- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| rows-clean, §f harvested                            | 15    | every table row already struck/routed, every §f item represented in the living docs                         | `2026-10-02_12-57_t23-closeout…`, `2026-10-02_13-50_session8-self-review…`, `2026-10-03_03-29_docs-health-v6-closed…`, `2026-10-04_12-34_island-cascade-leak-fix…`, `2026-10-12_10-52_SUPERB-train-completion…`, `2026-10-12_15-09_SUPERB-harvest-hardening…`, `2026-10-12_17-45_TODO-execution-and-hygiene…`, `2026-10-14_22-40_declarative-ddd-trains-bce…`, `2026-10-15_17-10_metaengine-t07-swap-g1-green`, `2026-10-15_19-13_metaengine-faxes-transcripts…`, `2026-10-17_20-16_modularization-proposal…`, `2026-10-17_20-34_improvement-pass…`, `2026-10-17_21-04_vendor-removal-and-doc-drift…`, `2026-10-18_05-34_round5-continuation-clone-gate…`, `2026-10-27_19-33_erraudit-followup-nolint-decay…`                            |
| §f tables are the harvest inventory (ledger-marked) | 6     | §a/§b already struck; the open §f table is a brainstorm whose items are represented (or generic-token only) | `2026-10-04_07-32_tooling-stragglers…`, `2026-10-04_12-39_passkey-users-train`, `2026-10-04_17-28_passkey-train-resumed…`, `2026-10-05_00-01_passkey-train-tail…`, `2026-10-11_06-34_branching-flow-retriage…`, `2026-10-25_19-22_cqrs-htmx-round3-deep-dive…`                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| rows-clean, §f only generic-token matches           | 34    | same as class 1 except the §f residue is the freshness gate's documented generic-token FP shape             | `2026-10-03_00-35…`, `2026-10-03_02-09…`, `2026-10-04_23-20…`, `2026-10-05_14-50…`, `2026-10-05_15-25…`, `2026-10-05_20-27…`, `2026-10-06_04-12…`, `2026-10-06_23-43…`, `2026-10-07_16-54…`, `2026-10-08_09-10…`, `2026-10-08_13-33…`, `2026-10-08_14-08…`, `2026-10-08_16-26…`, `2026-10-12_06-44…`, `2026-10-12_19-20…`, `2026-10-15_20-26…`, `2026-10-18_06-46…`, `2026-10-25_15-49…`, `2026-10-25_16-10…`, `2026-10-25_17-51…`, `2026-10-25_19-50…`, `2026-10-26_23-44…`, `2026-10-27_18-23…`, `2026-10-27_19-31…`, `2026-10-27_19-51…`, `2026-10-28_10-52…`, `2026-10-29_01-31…`, `2026-10-29_11-19…`, `2026-10-29_11-26…`, `2026-10-29_14-09…`, `2026-10-29_14-22…`, `2026-10-29_17-48…`, `2026-10-29_18-26…`, `2026-10-30_20-29…` |
| §f-only rows, verdicts resolved 2026-11-06          | 13    | 19 §f items annotated with verified evidence (or routed) this sweep, then the §f table ledger-marked        | `2026-10-04_08-17…`, `2026-10-04_19-42…`, `2026-10-06_13-12…`, `2026-10-07_18-15…`, `2026-10-14_22-54…`, `2026-10-15_18-48…`, `2026-10-15_21-26…`, `2026-10-23_17-17…`, `2026-10-24_16-22…`, `2026-10-25_15-16…`, `2026-10-25_18-37…`, `2026-10-25_19-53…`, `2026-10-27_15-57…`                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| superseded before posting                           | 1     | v2.7.0 never tagged; content shipped as v2.8.0; live backlog is the v2.8.0 drafts                           | `announcements/2026-09-24_v2-7-0_drafts.md` → `announcements/archived/`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |

The 43 remaining `2026-0*` files are KEEP, each with a written reason (see §a1). The
four executed September plans are the only KEEP entries whose ARCHIVE is blocked, and
their banner names the blocker and the row count (95 / 217 / 118 / 123).

## e) The 19 §f items resolved this sweep (the ones that unblocked archives)

Each verdict was verified against the tree before it was written: `nix fmt` +
`whitespace-drift.sh` are standing AGENTS commands; the smoke prints 89 checks (not
48); `audio.js` resumes the AudioContext defensively; red-main alerting exists
(`scripts/main-red-watch.sh` + `watch.yml`); `bootreport.go` guards both `copyRows`
reads behind one bounds condition; `ci-flake-ledger.md` carries the
mechanism-or-superseding-fix verdict column; `readmodels.ErrListFull` is the one
sentinel; `domain.NewInboundFaxJob`/`NewOutboundFaxJob` exist with tests;
`requireMultipartTo` is documented as the send-form prologue; `docs/research/README.md`
exists; the crm Resolver single-flights name lookups; the AGENTS citation really is
`crm.crmOutcome`; no `fmt.Fprint(fmt.Sprintf(…))` remains in any test; the
`.wp-crm-chip`/`-more` plan names became `.wp-crm-link`/`-context`/`-item` in the
shipped CSS. Two items were routed instead of closed: the hub re-pin (cross-repo) and
fax cover pages (ROADMAP).

## f) Up to 50 things we should get done next

_(brainstorm for HARVEST; every item below is already represented in
TODO_LIST/ROADMAP/CHANGELOG — the freshness gate verifies that at each battery run.)_

1. Push the 14 local commits and watch CI (the manual-push bar was met at 10:45).
2. Merge PR #234 once CI is green.
3. Stack `crm.caFile` pass-through in the typed CRM front.
4. Issue #9 acceptance criteria 1–2 on the PBX host.
5. Audit `.gitignore` for unanchored directory rules.
6. Decide the `crm` health leg's criticality.
7. Decide the findings-gate posture (50 pre-existing tier-3 findings).
8. Re-run the hub fan-out benchmark on a quiet host.
9. Store test train for the admin read seams / thread flags / migration v5.
10. Wire the `docs-gates` battery into CI so drift reds main.
11. Bulk verdict pass over the archive residue (73 → 68 files, ~1 600 rows).
12. `/context` contract snapshot test pinning webphone's parse against Kith's published shape.
13. History nudge for the notes save path.
14. `scripts/templ-generate.sh` wrapper pinning the regen compiler.
15. Attended-transfer UI vs the NO-GO ruling.
16. `reverseLines` → `slices.Reverse` — DONE by a later train; the symbol (and the readmodels engine store it lived in) no longer exists in the tree — the storage-ladder closeout moved the backends to `internal/cqrs`.
