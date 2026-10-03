# Status — docs-health AUDIT v6 (interrupted mid-sweep): 4 files annotated+archived, 18 remain, living docs updated

**Date:** 2026-10-03 00:35 CEST (session ran 2026-10-02 ~13:30 → 00:35).
**Mandate:** "View ALL `**/2026-0*` files; execute the docs-health SKILL
PROPERLY; TODO_LIST / CHANGELOG / AGENTS / README / ROADMAP / FEATURES all
SUPERB; archive FULLY done and UPDATED (inline strikethrough) .md files."
**Mode:** docs-health AUDIT (BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE);
SKILL.md + annotate tooling read before acting; the accepted v3/v4/v5 sweeps
used as the bar. Documentation-only: zero product code touched.

---

## a) FULLY DONE (verified this session)

1. **Skill + inventory.** docs-health SKILL.md loaded first; ALL
   `2026-0*` artifacts enumerated (12 unarchived status files: 11 `.md`
   - 1 immutable HTML; 77 already archived — archived-dir completeness
     gate verified CLEAN at start); all 11 unarchived 2026-09 reports read
     IN FULL; the entire 2026-10-01/02 train chain (24 reports, ~5800
     lines) read in full as the annotation evidence base.
2. **VERIFY sweep against code** (every claim by grep/read, not trust):
   the shipped-test set confirmed present (`TestShellHtmlLangFollowsSessionLang`,
   `TestCSSTokenBlocksAreMirrored`, `TestAssetsCarryContentETag`,
   `TestNoUnusedDictionaryKeys`, `TestStylesUseLogicalProperties`,
   `TestVersionedMigrations`, `TestThreadDeepLinkOpensConversation`);
   NOT present (honest routing verdicts): `--font-sans` consolidation,
   `DeliveryReceipt` non-verdict pin, caddy `validate`/`adapt` in the
   module check, caddy VM coverage, `binary-size-gate.sh`, A9 re-dial,
   `data-dial` in the DOM contract (id-scoped by design); confirmed:
   templ-components v1.19.4 vendored WITHOUT the round-12 asks (they ride
   their `[Unreleased]`), stale `docs/status/dedup-registry.md` copy
   GONE, PIPESTATUS lesson in lessons.md, `build-health-css.sh` still
   missing its trailing `nix fmt`.
3. **Living docs updated (the six made true):**
   - **TODO_LIST**: sweep-log rewritten (v6); tooling-hygiene row
     extended (health-css fmt append, smoke mypy triage, error-code
     registry table); OWNER-calls row extended with the three docs-health
     decisions (4-plan routed-as-resolved ratification, hash-bar
     ratification, check-rows 37-file baseline); three STALE rows
     DELETED per BUILD rules (UI/UX M9–M26 remainder — T12–T19 executed
     2026-10-02; UI/UX gates row — buildflow/flake/smoke/pins ALL green
     at the 12:57 close-out; verification/perf plan row — T02/T04–T09/T18
     executed, owner/stack legs already live in the mic, island-honesty,
     cross-repo rows).
   - **ROADMAP**: 2026-10-01 harvest cluster corrected (M21/M22, M24/M25,
     ETag/gzip/preload struck SHIPPED); NEW "2026-10-02 stragglers"
     cluster (A9 re-dial, I3/I6/I7/I9, feedback polish, mic extras,
     typography extras, guard ideas incl. the DeliveryReceipt/pin-class
     bullet); ring-silence open question struck RESOLVED; C4
     fax-thumbnail ratification added.
   - **AGENTS.md**: stale "40-check live smoke" line fixed →
     "48-check (+4 restart +8 boot-failure)" — line-neutral, file stays
     at 376/377.
   - **README**: 6 capability rows extended with the T11–T19 features
     (thread org + snippets, fax timeline/resend/Paperless, voicemail
     player, history filters/day grouping, device self-test + setup
     timings, five-part boot-error contract).
   - **CHANGELOG + FEATURES**: verified current against the close-out
     harvest — no drift found, no edit needed.
4. **ANNOTATE + ARCHIVE — 4 files complete** (banner + per-item strikes
   via annotate-status-items.py, `--verify` before every write,
   check-rows complete on each, `git mv` to `archived/`):
   - v3 audit (`2026-09-23_04-26`): 47 verdicts (incl. one manual
     whole-row strike — row 13's own text contains `'~~'`, the tool
     rightly refuses it).
   - v4 audit (`2026-09-29_05-11`): 42 verdicts + manual f11 strike.
   - templ-components round-12 (`2026-09-30_11-43`): 80 verdicts
     (webphone items done; everything else = "other repo" with their TODO
     numbers; v1.19.4 verified NOT to contain the four asks).
   - family-adoption train (`2026-09-30_11-45`): 74 verdicts (train
     shipped in v2.8.0; leftovers routed to TODO/ROADMAP/briefing/owner).

## b) PARTIALLY DONE

1. **The annotation/archive sweep itself — 4 of 22 files done.**
   Remaining, in planned order: the 7 other 2026-09-30 reports
   (typography 12:08, review-series 12:46, dedup-t2 12:59, fax-paperless
   13:09, dedup-t1 13:35, caddy-front 15:49, setup-adoption 15:49), the
   13 reports of 2026-10-01 (incl. the v5 audit — to be superseded by a
   v6 report of record), and the closed 2026-10-02 chain (sessions 1–7,
   boot-contract, t21-t22, mustinvoke — banner-only today, items
   unstruck). Evidence for every verdict is already gathered (all files
   read; code greps done); the remaining work is mechanical spec
   building. The 12:57 close-out + 13:50 session-8 reports stay LIVE
   (owner §g open, freshest).
2. **One typography miss surfaced, not yet fixed:** the typography train
   (2026-09-30, shipped pre-v2.8.0-tag) has NO CHANGELOG entry — grep
   confirms none exists in [v2.8.0] or [Unreleased]. Owes a retro bullet.
3. **Gates NOT yet run** (interrupted first): archived-dir completeness
   grep, check-rows over every newly annotated file (the 4 done files
   each passed individually), link check over touched files, scoped
   `go test ./cmd/webphone ./internal/server` (docs-only-delta
   precedent), `git ls-remote` end-state.
4. **The v6 inline health report (Accuracy/Fitness, visible math)** —
   not yet delivered; needs the sweep + gates first.
5. **The 4 owner-gated planning docs** (error-excellence, 20-year,
   04-29 pareto, 13-33) — classified STAY-LIVE (v5's standing decision);
   the ratification now sits in the OWNER-calls TODO row.

## c) NOT STARTED (deliberate or interrupted)

- Archiving the HTML snapshot `2026-09-18_16-52_*.html` — LEAVE
  (immutable; the accepted v3/v4/v5 precedent — HTML cannot carry
  markdown strikethrough without failing the completeness gate).
- The 2026-10-01/02 PLANNING docs (master plan, verification plan,
  ui-ux plan, T18 seam design, boot-contract plan, briefings) — stay
  live this sweep (operative references for open owner legs; prior
  sweeps' policy). Their turn comes when their owner legs close.
- Announcements, verdict/decision records, reviews/research/
  architecture-understanding 2026-0* files — SKIP/LEAVE per standing
  classification (v3 a2, v4 a1).
- AGENTS.md line-budget: any future add owes a compensating trim (at
  376/377 now; my edit was line-neutral).

## d) TOTALLY FUCKED UP (all caught; one required a revert)

1. **I fabricated a TODO_LIST item.** While extending the OWNER-calls
   row I introduced "suppress the `.templ`-import rule (same root
   cause)" — content with NO source in any report. Caught on the
   immediate re-read, reverted in the next edit, replaced with the three
   REAL docs-health decisions. Worst mistake of the session: an
   invented decision in the owner's decision queue.
2. **Archived v4 UNANNOTATED via a masked exit code.** I piped
   `annotate --verify` through `tail` — the pipe returned tail's rc, the
   failing spec looked green, and the `&&` chain ran `git mv` on an
   unannotated file. Caught within a minute (the follow-up apply
   printed FAIL), moved back out, spec keys fixed, properly annotated
   (42 verdicts), re-archived. This is VERBATIM the rc-masking trap this
   repo's lessons.md documents — I re-committed it inside the very
   discipline that exists to prevent it.
3. **Spec-key round-trips (≈6 wasted cycles):** substrings not in the
   line, rows whose own text contains `'~~'` (tool skips by design —
   correct), guessed key text instead of emitting keys mechanically for
   every file. The tool's `--verify` gate caught every one pre-write.
4. **One mtime-guard rejection** (edit tool after my own python write) —
   recovered via the atomic python insert the repo sanctions.
5. **Reading order cost:** I read the ENTIRE 2026-10 chain before the
   first annotation landed — the sweep was 7 hours old with 4 files
   archived when interrupted. Defensible (verdicts need evidence) but
   the per-file pipeline could have started at file 3.

## e) WHAT WE SHOULD IMPROVE

1. **Never pipe an annotate/gate command through `tail`/`grep` without
   `PIPESTATUS`** — d2 is the third session in this repo's history to
   relearn it. The one-liner `> log 2>&1; echo rc=$?` is the shape.
2. **Compose living-doc edits ONLY from read lines** — d1's fabricated
   item came from writing a row from memory of what "should" be there.
   Same class as the repo's "edit anchors from bytes" lesson.
3. **Emit keys mechanically per file, always** (`--emit-keys <file>
   <lineno>`) — hand-typed substrings cost 6 round-trips; the tool
   documents this exact rule.
4. **Start the per-file pipeline early, read remaining files in its
   gaps** — the evidence-first ordering was safe but slow.
5. **Retro-CHANGELOG debts found by VERIFY should be fixed on sight**
   (the typography entry) rather than queued — the owner-permission
   rule applies to AGENTS, not CHANGELOG history entries.

## f) Up to 50 things to get done next (execution order)

1. Finish the annotation+archive sweep: 7 remaining 2026-09-30 reports
   (typography, review-series, dedup-t2, fax-paperless, dedup-t1,
   caddy-front, setup-adoption).
2. Annotate+archive the 13 reports of 2026-10-01 (incl. the v5 audit —
   superseded once the v6 report of record exists).
3. Annotate+archive the closed 2026-10-02 chain (sessions 1–7,
   boot-contract, t21-t22) + complete the mustinvoke report's items
   (banner-only today — the skill's #1 failure mode if left).
4. Keep live: 12:57 close-out, 13:50 session-8, the 4 owner-gated plans,
   operative verdicts/briefings/announcements, the immutable HTML.
5. Write the archive manifest (one line per file: classification +
   deciding reason) — required for multi-file bulk archives.
6. Add the retro typography bullet to CHANGELOG [v2.8.0] (b2).
7. Gates: `grep -rLn '~~' docs/status/archived/ docs/planning/archived/`
   = empty; check-rows over every newly annotated file; link check over
   every touched file; scoped `nix develop -c go test -count=1
   ./cmd/webphone ./internal/server` (docs-only-delta precedent).
8. Deliver the inline v6 health report (Accuracy/Fitness, per-doc
   findings, visible math).
9. `git ls-remote` end-state verify (daemon push lag).
10. Write the v6 docs-health report of record (this file is the session
    report; the v5 file gets its superseding banner at archive time).
11. Route-or-do the typography f3-style stragglers my greps surfaced
    (already routed to ROADMAP this session — verify nothing was missed).
12. Sweep `/tmp/spec-*.tsv` session scratch (ephemeral, no repo trace).

## g) Questions I can NOT figure out myself

1. **Scope of the 2026-10-02 archive leg:** the close-out (12:57) and
   session-8 self-review (13:50) clearly stay live (owner §g open). But
   sessions 3–7 + boot-contract + t21-t22 are same-day reports whose
   trains verifiably closed — archive them this sweep (my plan), or is
   same-day archiving too aggressive while you may still be reading the
   chain?
2. **The v5 audit (2026-10-01_17-27):** archive it once its open items
   are routed (b2/b3 already are, or now are via the OWNER-calls row) —
   or keep it live as the previous docs-health report of record until
   you have read the v6 report?
3. **The 4 owner-gated planning docs:** they stay live pending the
   sitting's "routed-as-resolved" ratification (now TODO row). If you
   ratify NOW, say so and the next sweep annotates+archives all four in
   one pass — otherwise they wait for the sitting with the other ~31
   decisions.

---

_Point-in-time snapshot — annotate, never rewrite. The auto-commit daemon
owns the sweep (this report included). SWEEP PAUSED AT 4/22 FILES; resuming
is safe — every remaining file is fully read with verdicts derived, only
spec-building + gates remain._

---

## Archive manifest — v6 sweep completion addendum (2026-10-03)

All 34 files below: classification **ANNOTATE → ARCHIVE** (every numbered item resolved inline;
banner + per-item verdicts; `check-rows` complete; `git mv` to `docs/status/archived/`).
Deciding reason per file:

- `2026-09-30_11-43_templ-components-round12-session-status.md` — templ-components round-12: adoption verdicts shipped; remainders routed
- `2026-09-30_11-45_family-adoption-train-status.md` — family-adoption train: tier-2 zero reached; re-measure watch stands
- `2026-09-30_12-08_typography-font-design-train.md` — typography train: shipped in v2.8.0 (retro CHANGELOG bullet landed by this sweep)
- `2026-09-30_12-46_review-series-status-and-self-review.md` — review series: process records; routed remainders
- `2026-09-30_12-58_go-cqrs-lite-question-session-status.md` — go-cqrs-lite Q&A: decision REJECTED (root-library-only); record stands
- `2026-09-30_12-59_dedup-sweep-t2-status.md` — dedup t2 sweep: rulings in the registry; suppression-scope routed to OWNER-calls
- `2026-09-30_13-09_fax-paperless-plan-session-status.md` — fax-paperless plan: seam shipped + pinned; stack option routed cross-repo
- `2026-09-30_13-17_dedup-sweep-t1-status.md` — dedup t1 sweep: rulings in the registry
- `2026-09-30_13-35_dedup-sweep-t1-session-status.md` — dedup t1 session: rulings in the registry
- `2026-09-30_15-49_caddy-front-diagram-rework-status.md` — caddy front diagram: rework shipped in v2.8.0
- `2026-09-30_15-49_setup-adoption-train-verdict-status.md` — setup adoption: REJECTED verdict stands (split-brain + binary-size NO-GO)
- `2026-10-01_01-21_visual-verification-loop.md` — visual verification loop: T23 harness superseded it
- `2026-10-01_01-45_post-v2-8-0-full-execution-session-status.md` — post-v2-8-0 execution: trains verified; owner legs routed
- `2026-10-01_02-12_samber-do-composition-root-health-dashboard-train.md` — samber-do dashboard train: shipped (probe triple + dashboard seam)
- `2026-10-01_02-14_post-v2-8-0-execution-self-review-status.md` — execution self-review: lessons landed; AGENTS compaction resolved by events
- `2026-10-01_02-54_prod-call-demo-webtransport-qa-and-island-latency-diagnosis.md` — demo/Q&A session: ring-silence fix = T06; demo legs ride the cross-repo row
- `2026-10-01_03-52_post-v2-8-0-harvest-paperless-train-status.md` — paperless harvest: seam shipped + pinned; nix-hash-fix answers the enforcement ask
- `2026-10-01_04-07_mic-prewarm-accept-latency-train-status.md` — mic pre-warm train: shipped + pinned; live-call ritual routed (owner terminal)
- `2026-10-01_05-26_island-honesty-hold-offline-banner-train-status.md` — island honesty train: shipped (cc98c2e); T15/T22/T23 closed the follow-ups
- `2026-10-01_05-26_nix-file-review-session-status.md` — nix review: polish executed wholesale by the 05:56 FIX-IT-ALL train
- `2026-10-01_05-27_performance-inventory-train-status.md` — perf inventory: SHIPPED as T10/T11 + dial-focus warm
- `2026-10-01_05-56_nix-review-fixes-train-status.md` — nix fixes train: splits byte-identical; golden + release guard closed later
- `2026-10-01_06-59_ui-ux-pareto-train-boundary-lesson-status.md` — UI/UX pareto: M1–M8 stood; remainder executed by T12–T19; Ledger ruling = AGENTS doctrine
- `2026-10-01_17-27_docs-health-audit-v5-living-docs-superb-archive-sweep.md` — docs-health v5: SUPERSEDED by this v6 sweep (its f-list fully closed)
- `2026-10-01_20-12_master-todo-execution-train-status.md` — master-todo s1: trains stood; session-2 addendum closed the pins; T10+ shipped later
- `2026-10-01_23-11_master-todo-train-session2-status.md` — master-todo s2: pins + T10 ETag/gzip stood; T11–T19/T21–T23 shipped later
- `2026-10-02_08-26_master-todo-train-session3-status.md` — master-todo s3: T11/T12/T13 stood; T13 tail closed green; remainder shipped
- `2026-10-02_09-03_master-todo-train-session4-status.md` — master-todo s4: T13 closed + T14 shipped whole; C4 decline recorded
- `2026-10-02_09-34_master-todo-train-session5-status.md` — master-todo s5: T15 + T16 closed; T17 shipped next; defaults stood
- `2026-10-02_10-11_master-todo-train-session6-status.md` — master-todo s6: T17 closed + T18 seams green; views finished by s7
- `2026-10-02_10-19_mustinvoke-operator-contract-session-review.md` — mustinvoke review: boot contract shipped same morning; §f completed inline by this sweep
- `2026-10-02_11-41_t18-t23-train-session7-status.md` — master-todo s7: T18 whole + T19/T21/T22 + T23 harness; tail closed at 12:57
- `2026-10-02_11-43_boot-contract-execution-run-status.md` — boot-contract run: shipped whole (65b7f0d); D3 + stack patch ride the boot row
- `2026-10-02_11-43_master-todo-t21-t22-verification-status.md` — T21/T22 verification: batch-2 wins + battery stand; nolints landed; KVM ran at 20:12

Kept live by design: the 12:57 close-out + 13:50 session-8 reports (current), both v6
session reports (00:35, 02:09), the four owner-gated planning docs (error-excellence,
20-year-durability, 04-29 pareto, 13-33 stack-adoption — pending the owner sitting), the
operative verdicts/briefings, and the immutable HTML snapshot.
