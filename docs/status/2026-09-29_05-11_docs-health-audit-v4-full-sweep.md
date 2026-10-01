# Status Report — docs-health AUDIT v4 (full sweep): living docs superb, annotate phase mid-flight, archive pending

Date: 2026-09-29 05:11 CEST · Session scope: the "View ALL `**/2026-0*`
files; execute the docs-health skill PROPERLY; the six living docs
SUPERB; archive FULLY done and UPDATED (inline strikethrough) .md
files" mandate. Documentation-only delta. Point-in-time snapshot per
the STOP-AND-WAIT instruction — the archive leg has NOT run.

---

## a) FULLY DONE (verified this session)

1. **Skill + inventory**: docs-health SKILL.md loaded before any
   action; ALL 140 `2026-0*` artifacts inventoried and classified;
   archived dirs' completeness gate verified green at start; all 18
   non-archived status `.md` files read IN FULL; 6 plan candidates
   read in full; verdict/decision/operative docs (csrf, pwa,
   webhook-idem, openapi, session-persistence, tailwind, sip.js eval,
   oob, hub-fanout, coverage, briefing, command sheet, announcements)
   classified SKIP/LEAVE per skill policy (they carry their own
   resolution or remain operative); HTML/D2/SVG = LEAVE (immutable —
   the accepted v3-sweep precedent; the two oldest HTML reports were
   content-sampled, not just name-triaged).
2. **VERIFY findings, all fixed on sight**:
   - AGENTS "pre-2.8 binaries" claim → commit-pinned (`e6ea2c7`,
     2026-09-29, unreleased) + the never-write-version-claims rule.
   - AGENTS `.templ` formatter-ownership question (open since the
     09-23 15:58 sweep, carried by four reports) → CLOSED by
     inspecting nix/treefmt.nix: `.templ` sources are deliberately
     formatter-unowned; AGENTS formatting bullet records it.
   - README: `environmentFiles` (plural) module option row added
     (existed in the module since `2356ec8`, missing from the table);
     one-paragraph `nix/` flake-layout pointer in Development.
   - FEATURES: Browser E2E row refreshed (195s/184s @ `271f5ef` with
     the FOUC scenario aboard — was stale 293s/322s); aarch64 row
     refreshed (2026-09-26, ELF e_machine=183 — was 2026-09-18).
   - CHANGELOG [Unreleased]: Fixed entry for the flake-version
     drift-test regex (the 09-29 flake-split comment false-match);
     Changed entry for the pbx `errors.New` sentinels (`b03327f`).
   - DOM-contract "35 ids" claim re-derived (35 — correct);
     v2.7.0-untagged re-verified (`git tag` tail = v2.6.0);
     scripts-vs-flake.nix sweep post-split clean (buildflow.sh
     existence-only, release.sh sed verified by the train,
     webphone-smoke.py existence-only); `checks.treefmt` proven
     pre-split (grep at `46ad1d3`); `createFilePart` proven the sole
     multipart file-part writer under internal/.
3. **HARVEST landed**: TODO_LIST rebuilt — honest header (the
   "every claim checked" overclaim the 09-26 §d5 report flagged is
   GONE), fresh closure narrative (≤8 lines), 11 evidence-cited rows:
   kept release-tail / SMS-bridge / owner-calls (agenda extended to
   ~30 with every 2026-09-25→29 question) / announcements / watches /
   templ-components-release-E2E; DELETED the DONE-parked go-health
   row (park trigger moved into the watches row); ADDED erraudit
   tier-2 conversions (config.go 22 → store/messages.go 20 →
   pbx/client.go 8, due-shrink 2026-10-22), nix-review follow-ups
   batch (§f-cited wholesale), gateway honest-Content-Type follow-ups,
   `/version` enrichment, AGENTS compaction (498→≤377, owner-gated),
   tooling hygiene (markdown-lint posture, codespell policy, buildflow
   reconcile, binary freshness, render-diff script).
   ROADMAP: +12 open questions (2.8.0 numbering, attachment_limit,
   sniff lifespan, push-lag threshold, session push ratification,
   tier-2 intent, erraudit tree upstream, AGENTS restructure, prod
   chmod, destDir nesting + whitespace, stack verification timing) +
   a 2026-09-29 raw-ideas cluster (metrics scraper fencing, stack
   identity endpoint, sentinel `errors.New` micro-check).
4. **Living-doc fixes on sight**: docs/lessons.md +3 (the gawk-5.4.1
   escape-drop awk variant + comments-claim-fixes trap; the
   `--no-link`/stale-`result` artifact trap; silent daemon push
   stalls); docs/dedup-registry.md protocol upgraded (attribute
   against YOUR re-run never the paste; re-read EVERY occurrence;
   sweep-log lines carry status-doc links) — resolving the 09-27
   report's §e1/e.4/e.6/§f5/§f26 on the spot.
5. **ANNOTATE: 832 inline verdicts across 16 status reports**
   (banner + per-item strikes via annotate-status-items.py,
   `--verify` before EVERY write, zero failed-writes): 09-29
   04-32/03-05/02-26, 09-27 23-43, 09-26 19-24, 09-25 05-10/05-08/
   04-04, 09-24 18-03/18-05/16-56/12-41, 09-23 22-11/21-15/15-58/
   15-41. Verdict vocabulary: `done at <hash>` / `DONE (<evidence>)`
   / `routed — <TODO row/ROADMAP §>` / `standing ritual|watch` /
   `NOT-DO` / `resolved by events`. Plus superseded-by-registry
   banners added to the three already-archived art-dupl reports
   (22-55, 01-12, 03-01); 16-56/15-58/18-05 carry theirs in-banner.
6. **On-the-spot closes that shrank the harvest**: the 09-27 §f12
   "helper micro-test remainder" was a STALE carry (every named
   helper already had its micro-test at `7bf32a3` + crm_test.go
   subtests — struck DONE with that evidence); the `.templ` ownership
   and lessons entries closed live rather than routed.

## b) PARTIALLY DONE

1. **02-47 + 04-26 status reports**: unstruck-remainder keys EMITTED
   (02-47: §a1-7, §b1-2, §c1-3, §d1-3, §f8; 04-26: §a1-15, §b1-4,
   §d1-4, §e1-6, §f1/3/13-17) — specs not yet written/applied.
2. **6 planning docs** (04-29 pareto — verdict CLOSED, 13-33 stack
   adoption M1-M22, send-failure-ux A-F, composer-ux T1-T4,
   error-excellence T01-T18, 20-year-durability T1-T27): read,
   classified FULLY-RESOLVED-or-routed, banners + table strikes NOT
   yet applied.
   ~~3. **ARCHIVE leg not started**: ZERO `git mv` has run — all 16~~ done 2026-10-01 (this sweep): 16 status reports + 2 planning docs git mv-ed to archived/
   annotated files still live in docs/status/. Order will be
   annotate → verify → archive (the 04-26 audit's own lesson).

## c) NOT STARTED

~~- `git mv` of the 22 files (16 status + 6 plans) into `archived/`.~~ done 2026-10-01 — 18 moved (16 status + send-failure-ux + composer-ux); the other 4 plans stay live (owner g1 + owner-open remainder)

- Post-move gates: `grep -rLn '~~'` completeness over both archived
  dirs; check-rows.py uniformity over every annotated file; stale
  reference sweep (TODO_LIST already cites the post-move archived
  paths — they resolve only AFTER the move); markdown link check.
- Scoped quality gate (docs-only-delta precedent: scoped go test +
  codespell) — not run.
- `git ls-remote` end-state verify (push lag observed at session
  start: remote trailed local).
- The inline health report (Accuracy/Fitness, visible math) — the
  health-report-format reference has not been loaded yet.

## d) TOTALLY FUCKED UP (all caught by --verify or immediate re-read; zero wrong writes landed)

1. **Hand-typed spec keys despite the tool's own warning.** The
   annotate tool documents "specs are generated mechanically via
   --emit-keys … never hand-typed from memory"; I pasted emit output
   but hand-typed verdict LINES with substrings copied by eye — five
   failed --verify rounds (backticks silently dropped in a heredoc, a
   pipe typed instead of @, "unflagged" vs the file's "non-flagged",
   a missing tab, a trailing-hyphen mismatch). Every failure was
   caught pre-write; the round-trips were pure waste.
2. **One accidental CHANGELOG mutation**: a multiedit changed
   "NixOS module:" to "Nix:" on an existing Fixed bullet while
   inserting a sibling — caught on the next read, reverted in the
   following edit.
3. **Two premature archive-path references** in the rebuilt
   TODO_LIST (verdict doc + command sheet cited as archived before
   any move) — self-caught within minutes, repointed to live paths.
4. **Skill-reference loading was partial**: SKILL.md + tool
   docstrings fully read; the 7 reference files were NOT all loaded
   before starting (annotate placement rules were followed from
   SKILL.md, but the health-report format + verify checklist
   references are still unread and are needed for the pending legs).

## e) WHAT WE SHOULD IMPROVE

1. Spec verdict lines get composed by EDITING the --emit-keys output
   in place (append `\t<verdict>` to the pasted key) — never retyped.
2. Inserting a sibling bullet must re-read the anchor block after the
   edit (the CHANGELOG prefix mutation class).
3. Paths referenced in living docs must be written LAST, after the
   moves they cite (or written as post-move paths with the move in
   the same breath — my TODO rows now depend on the archive leg
   landing before any gate run).
4. Load the format reference BEFORE the phase that needs it, not at
   the phase.

## f) NEXT — the honest remainder (execution order)

1. Annotate 02-47 remainder (keys already emitted).
2. Annotate 04-26 remainder (keys already emitted).
3. Annotate 04-29 pareto plan (banner cites the filled Verdict; strike
   the 24 C-rows + 84 F-rows as done/routed per the Outcomes section).
4. Annotate 13-33 stack-adoption plan (22 M-rows + 77 F-rows: done at
   their commits; M18 parked; M19 tail = release-tail row).
5. Annotate send-failure-ux plan (coarse 1-8: A/B executed; C
   `37ffc53`; D shipped earlier; E `6ac8962`; F found-already-shipped
   — TODO_LIST records it).
6. Annotate composer-ux plan (T1-T4 executed; the "planned next
   trains" all shipped in v2.6.0 — FEATURES rows).
7. Annotate error-excellence plan (T02-T10, T13-T14, T16-T17 done
   with hashes; T01/T11/T12/T15/T18 routed — owner-calls/release rows).
8. Annotate 20-year-durability plan (execution log already carries
   CLOSED markers; strike the 27-row coarse + fine tables to match,
   routing T16 → the AGENTS-compaction row).
9. `git mv` all 22 files to their `archived/` dirs.
10. Stale-reference sweep: grep living docs + AGENTS for the moved
    filenames; fix every pointer.
11. Completeness gate: `grep -rLn '~~' docs/status/archived/
    docs/planning/archived/` must print NOTHING.
12. check-rows.py over every annotated file (uniform rows).
13. Link check over every touched markdown file.
14. Scoped quality gate: `nix develop -c go test -count=1
    ./cmd/webphone ./internal/server` + codespell over touched files
    (docs-only-delta precedent; full gates already green on main at
    `0230ead`+).
15. `git ls-remote` end-state verify (daemon push lag).
16. Load health-report-format.md, deliver the inline health report
    (Accuracy/Fitness, per-doc findings, visible math).
17. Confirm the daemon-swept tree is clean at close (it already swept
    this session's work to `46bd326` mid-flight — expected).
18. If ratifying "routed-as-resolved" for the two plans with
    owner-open remainder (g1 below): nothing extra; if NOT: keep
    error-excellence + 20-year plans LIVE with partial annotation.
19. ROADMAP hygiene check: verify no double-listing between the new
    open questions and the owner-calls TODO row (they deliberately
    mirror; confirm wording does not drift).
20. Sweep `/tmp/spec-*.tsv` + `/tmp/keys-*.txt` (session scratch).

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Routed-as-resolved for ARCHIVE?** The error-excellence and
   20-year-durability plans still carry owner-open items (carrier-MMS
   isolation test, oops ratification, recordings intent, AGENTS
   compaction) that now live in the owner-calls/TODO rows. Ratify
   routed-as-resolved (archive both) or keep those two plans LIVE
   until the owner-calls sitting lands? (The other 20 files are
   unambiguous.)
2. **Home for the routed owner questions**: is ROADMAP "Open
   questions" the standing home for the 2026-09-25→29 g-questions,
   or should the 2026-09-22 owner-calls briefing doc get a refresh
   pass (it predates ~15 of the ~30 queued decisions)?
3. **AGENTS.md compaction timing**: 500+ lines now against the 377
   buildflow cap; the compaction row is owner-permission-gated (asked
   02:26 §g3, unanswered). Ratify a follow-up session for it, or
   leave the cap warning standing indefinitely?

---

_Point-in-time snapshot. The auto-commit daemon owns the sweep (this
report included — it will land as a heuristic commit). Annotated files
carry their banners + strikes NOW; the archive moves are NOT yet run —
resuming mid-flight is safe: every pending file is either fully
annotated (16) or fully read with keys emitted (2 + 6)._
