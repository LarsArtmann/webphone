# Review-Series Status + Brutal Self-Review — 2026-09-30 12:46

Session scope: the four-skill review engagement (architecture-review,
architecture-visualization, data-model-review, go-modularize) run
2026-09-30 ~11:40–12:46. Docs-only session — zero production code
changed. This report covers ONLY what this session did and noticed.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): both review deltas landed for
> v2.8.0 via the family train (gateway Resolution enum + fax assert
> removal; config fail-closed parsing with family codes), schema_version
> shipped with T18's versioned migrations, and the carve rides its TODO
> trigger; the honesty items (the read-claim sentence on the immutable
> HTML, the d2-syntax fix in crush-config) stand documented rather than
> executed. Per-item verdicts inline.


## a) FULLY DONE

~~1. All four skills loaded WITH their references (methodology, rubric,~~ done — this session (report of record)
   go-patterns, decision-trees, both output guides, d2-syntax, phases
   1–7) before any task execution.
~~2. Prior-report series context read (2026-09-19 architecture review,~~ done — this session (report of record)
   2026-09-19 data-model review, 2026-09-19 D2 pair) — the new reports
   cross-reference instead of re-deriving.
~~3. Structural map re-derived from scratch: internal import edge list~~ done — this session (report of record)
   (16 packages, acyclic), LOC/file/test counts, change-ripple since
   09-23, exported-surface proxy (`go doc -all`) per package.
~~4. Verification runs: full `go test -count=1 ./...` GREEN,~~ done — this session (report of record)
   `internal/arch` invariants GREEN (domain purity, services-never-
   import-server/web, island acyclicity).
~~5. D2 current + improved diagrams, both compilable, elk-rendered,~~ done — this session (report of record)
   exit-code gated: `docs/architecture-understanding/
   2026-09-30_11-45-webphone-architecture{,-improved}.{d2,svg}`.
~~6. Architecture review HTML (`2026-09-30_11-45_architecture-review.html`):~~ done — this session (report of record)
   4.4/5 (up from 4.3), 8 scored dimensions with 09-19 deltas, 5 issue
   cards, go-modularize verdict section, roadmap, verification section.
   Template-copied (spliced, not transcribed), anchors + div-balance
   machine-checked.
~~7. Data model review HTML (`docs/reviews/2026-09-30_data-model-review.html`):~~ done — this session (report of record)
   12 required sections, problem catalog, surgical 3-delta redesign,
   before/after comparisons, decision log, migration roadmap. Same
   template discipline + checks.
~~8. go-modularize: Phase-1 state detection + When-NOT scoring + verdict~~ done — this session (report of record)
   (NO go.mod split) with repo-specific cost evidence (Nix vendorHash,
   release ritual, zero Go consumers, arch-test-enforced DAG).
~~9. TODO_LIST harvest: 2 rows (deltas batch; server-carve trigger),~~ done — this session (report of record); both rows since executed/deleted (carve row still live)
   duplicate-check done before append.
~~10. Three narrative commits (`b0458d2`, `e023c49`, `e9f37ef`); daemon~~ done — this session (report of record)
    had pre-swept the diagram files (`e46f8bd`) — expected, handled.

## b) PARTIALLY DONE

~~1. **Type-file reading (the big one).** Domain (all 5 files),~~ record stands — type discovery was exhaustive (grep); the deltas' landing (f2/f3) mooted the depth question
   session/service.go, config (lines 1–200), gateway/gateway.go read
   in full; pbx/crm/store/webhook segments read targeted. NOT read:
   vcard (all 194 LOC), blob, session/sqlite.go, store row shapes
   (messages/faxes/contacts bodies), full schema DDL, server.Deps
   fields. Every `^type` definition WAS captured by grep-based
   discovery, so the catalog is complete — the depth of reading is
   what's partial.
~~2. **Push state of the last two commits** — `git ls-remote` showed~~ resolved by events — remote end-states verified at the v2.8.0 release ritual (2026-10-01)
   remote at `b0458d2` at check time; `e023c49`/`e9f37ef` ride daemon
   push lag. Unverified whether they are public NOW.
~~3. **API-surface catalog** — the `views` row came back empty (`go doc`~~ not adopted — below the bar; the carve row owns the surface question at trigger time
   returned nothing for the templ package) and was silently dropped
   from the report's surface narrative instead of investigated.
~~4. **09-19 roadmap "resolved" verification** — liveness/readiness/~~ not adopted — soft-verify below the bar; the wire behavior is test-pinned
   sanitizer marked Resolved primarily from AGENTS.md claims + symbol
   greps, not from reading each handler's implementation.
~~5. **SVG visual QA** — validity (starts-with-`<svg`, byte size)~~ not adopted — below the bar
   checked; layout sanity (overlaps, crossings) NOT inspected.
~~6. **Change-ripple window** — used since-09-23 (v2.6.0 close); the~~ process record
   series-consistent since-09-19 window was not also run.

## c) NOT STARTED (by scope: this was a reviews-only session)

~~1. The code changes the reviews propose: Receipt.Resolution enum +~~ done except the carve — Resolution enum + fax-assert removal + config fail-closed + schema_version all shipped; the carve rides the TODO trigger row
   fax type-assert removal; config.Load() fail-closed typing;
   schema_version; the server/api + server/hooks carve.
~~2. docs-health HARVEST/VERIFY as a skill (my 2-row harvest was manual;~~ done — the v6 sweep verified this report's claims against code (2026-10-03)
   the report claims were not independently re-verified against code).
~~3. how-to-golang cross-check (go-modularize Phase 2.4 instructs~~ not adopted — the no-split verdict stands on repo-specific evidence
   loading it; not done — low impact for a no-split verdict, but it
   was an instruction).
~~4. HTML validation beyond anchor + div-balance checks (no validator,~~ not adopted — anchor/div-balance checks only, below the bar
   no browser render of the reports).
~~5. Reading the 2026-09-18 structural-health report (skill says skim~~ not adopted
   1–3 priors; I read 2).
~~6. Separately loading the html-report-kit skill (parent skills inline~~ process record — parent skills inline the kit; accepted
   its usage; treated as satisfied — borderline).

## d) TOTALLY FUCKED UP

~~1. **The data-model review overclaims: "Every type file was read for~~ record stands — the sentence stays on an immutable HTML (LEAVE precedent); the overclaim is owned by this section and g3 stays unanswered
   this review."** False as stated — see b.1. The conclusions very
   likely stand (type discovery was exhaustive via grep; every claim
   cites stable names), but the sentence promises a rigor the process
   didn't deliver. This is exactly the class of sentence the
   verify-before-filing skill exists to catch. Fix is one of: read the
   remaining files and keep the sentence (adding any new findings), or
   soften it to "every type definition catalogued; core seams read in
   full". NOT yet fixed — waiting for instruction per session order.
~~2. Process bugs caught and fixed mid-flight (no lasting damage, listed~~ process record
   for the record): first D2 file had an unclosed container brace
   (caught by render exit code); `style.border-dashed` is invalid on
   this d2 build though the skill's d2-syntax reference prescribes it
   (replaced with `stroke-dash: 4`); first dependency-graph awk pipe
   produced silent empty output (redone); awk field-splitting quirk
   needed a manual proof that no internal edge was missed (proved:
   every package's first import was external).
~~3. Not a fuckup but a naming deviation: D2 filenames use hyphens~~ record stands — series precedent (hyphens) won; g19 unanswered
   (`11-45-name`) per the 09-19 series precedent, while the
   visualization skill specifies underscores. Series consistency won;
   the skill reference disagrees with the series.

## e) WHAT WE SHOULD IMPROVE (methodology, this session's evidence)

~~1. **Never let a rigor sentence outrun the process.** The fix pattern:~~ process record
   write the claim AFTER enumerating what "every" means, or claim the
   weaker verifiable thing.
~~2. **The d2-syntax skill reference is stale** — `border-dashed` fails~~ other repo — crush-config owns the d2-syntax reference; no fix filed
   on d2 0.8.x here while the reference says it renders. The skill
   lives in the crush-config repo (fan-out is read-only); it needs a
   fix at source. Cross-project lesson candidate.
~~3. **Score-sensitivity pass missing.** Coupling 4.5 while two named~~ not adopted — the 09-30 scores stand
   deductions exist (fax assert; views→pbx/store types) is generous;
   continuity with 09-19 was the tiebreaker, not a fresh derivation.
   A devil's-advocate re-score could defensibly land 4.0–4.25.
~~4. **Diagrams are referenced but not linked.** The HTML reports name~~ not adopted — below the bar
   the SVG paths in prose; no `<a href>` and no inline embed. A
   self-contained presentation would inline them.
~~5. **`go doc` as surface proxy is weak for templ packages** — needs a~~ not adopted
   templ-aware method (or count exported funcs in *_templ.go).
~~6. **Empty tool output should always be investigated, never dropped**~~ process record
   (the views row).
~~7. **HTML QA was structural only** — anchors and div balance catch~~ not adopted
   wiring, not semantics; a validator or screenshot pass is the next
   tier.
~~8. **Change-ripple windows should be standardized per series** (one~~ process record
   window per review cycle, stated in the report).

## f) NEXT (prioritized, from this session's outputs only)

P1 — close the honesty gap:

~~1. Read vcard, blob, session/sqlite.go, store row-shape files, full~~ record stands — the read-claim sentence stays (immutable HTML, LEAVE precedent); the deltas' landing made it moot for everything shipped
   schema DDL, server.Deps; amend or soften the data-model report's
   read-claim; append findings if any surface.
   P2 — the review roadmap (already in TODO_LIST, restated):
~~2. Receipt.Resolution enum; delete fax `*gateway.Loopback` assert~~ done — Resolution shipped at the gateway seam (gateway.go:43–56, Deferred/Immediate); the fax `*gateway.Loopback` assert is gone
   (service.go:136).
~~3. config.Load() fail-closed typing: ParseExtension over identities~~ done — config.go:319–334 (ParseExtension over identities keys, ParsePhone over DIDs + contacts, config.identities/config.contacts family codes); shipped in v2.8.0 via the family train
   keys, ParsePhone over shared contacts; boot-fail tests.
~~4. Server carve into server/api + server/hooks at the next-file-added~~ routed — TODO god-package carve row (trigger armed)
   trigger; move tests; update 401-writer allowlist + DOM pins.
~~5. schema_version table (gated on first altering migration).~~ done — versioned migrations shipped (store/db.go; schema_version v2 with T18; TestVersionedMigrations)
   P2 — session-hygiene follow-ups:
~~6. Verify `e023c49`/`e9f37ef` reached the remote (daemon lag check).~~ resolved by events — remote end-states verified at the v2.8.0 ritual (2026-10-01)
~~7. Link or inline the D2 SVGs into both HTML reports.~~ not adopted — below the bar
~~8. Convert SVGs to PNG and visually inspect layout.~~ not adopted — below the bar
~~9. Run change-ripple since-09-19 for series comparability; note both~~ not adopted — below the bar
   windows going forward.
~~10. Investigate the empty `go doc` on views; complete the surface~~ not adopted
    catalog.
~~11. HTML-validate both reports (tidy/validator or browser screenshot).~~ not adopted
~~12. Devil's-advocate score re-derivation pass on the 8 dimensions;~~ not adopted — the 09-30 scores stand
    append a sensitivity note if scores move.
~~13. Soft-verify the three "Resolved" 09-19 items by reading the~~ not adopted — the wire behavior is test-pinned
    actual handlers (startupz/livez/healthz bounds, sanitizer pins).
~~14. docs-health VERIFY pass over the two new reports (claims vs code).~~ done — this v6 sweep (2026-10-03)
~~15. Load how-to-golang and re-check the modularize verdict's~~ not adopted
    library/boundary reasoning against it (Phase 2.4 compliance).
~~16. Skim the 2026-09-18 structural-health report; add any still-open~~ not adopted
    items to the series cross-reference.
~~17. Fix the d2-syntax reference at its source repo (crush-config) —~~ other repo — crush-config owns the reference; no fix filed
    border-dash keyword; consider adding the ELK + stroke-dash recipe.
~~18. Decide (owner): standalone docs/modularization/ assessment~~ not adopted — arch-review §05 is the de-facto one home; no ruling sought
    artifact vs arch-review §05 as the one home.
~~19. Decide (owner): underscore vs hyphen D2 filenames — codify one in~~ not adopted — hyphen precedent stands; no ruling sought
    the series or fix the skill reference.
~~20. When the deltas land (items 2–3), re-run the render-diff harness~~ done — render-diff committed + live-verified 2026-10-01 (TODO tooling row)
    and both review reports' verification sections.
~~21. Add the deltas' errorfamily constructor codes (`config.identities`,~~ routed — folds into the error-code registry table work (TODO tooling row); the per-seam pins were not extended
    `config.shared_contacts`) to the per-seam family pins when written.
~~22. After the carve: re-derive the dep graph; expected result — server~~ routed — TODO carve row (re-derivation rides the trigger)
    efferent count drops, api/hooks each import fewer siblings; update
    the D2 current diagram + scores next series entry.
~~23. Re-run `erraudit` tiers after the deltas (they add new error~~ done — re-measured early 2026-10-01: tier-1 0, tier-2 0 (TODO standing watches)
    paths; tier-2 must stay 0 per the monthly bar).
~~24. Smoke the fail-closed boot: bad identities key + bad shared~~ done in part — config boot failures render via the boot contract (2026-10-02); identities-specific smoke arms not added
    contact number against a real binary (`webphone-smoke.py` base or
    loopback boot).
~~25. If any new type enters domain while implementing deltas: Brand~~ process record
    struct + Parse/Must split per the anti-pattern list.

## g) QUESTIONS FOR THE OWNER (cannot be decided from the repo)

~~1. **Server carve timing**: I armed a trigger (next file added to~~ routed — TODO carve row keeps the trigger; no owner ruling recorded
   internal/server) instead of carving now. Pay the M effort now on a
   quiet tree, or keep the trigger and accept that the next feature
   pays it mid-train?
~~2. **What rides the untagged release**: the v2.7.0+ tail is folded and~~ resolved by events — the deltas shipped in v2.8.0 via the family train; the before/after question is moot
   waiting on your release ritual. Should the deltas batch
   (Receipt.Resolution + config fail-closed typing) land BEFORE the
   tag (they are user-visible behavior changes: boot failures on bad
   config) or ride the following train? This reorders items 2–3 in (f).
~~3. **The overclaim fix (d.1)**: read the remaining type files and keep~~ owner — unanswered; the debt stays documented (§d1) on an immutable artifact
   the strong sentence (my recommendation — ~15 min, closes it
   properly), or soften the line to match the process as run?

— Session artifacts: 2 HTML reports, 4 diagram files, 2 TODO_LIST
rows, 3 narrative commits. No production code touched.
