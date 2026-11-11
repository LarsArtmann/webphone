# cqrs-lint delta resolved: it was the result cache — the follow-up session report

> Claims verified as of 2026-11-11 01:58 CET. CI verdicts and push states rot
> fastest; each is timestamped below.

Session scope: the owner said GO on the 00-36 report's verification tail ("GET
SHIT DONE. The WHOLE TODO LIST."). This report covers exactly that follow-up
turn — the delta investigation, the verification sweep, the gates healing, and
what this turn got wrong. Companion: the 00-36 triage report (annotated
in-place this turn).

## The one-paragraph verdict

The BuildFlow-vs-CLI "6-finding delta" is CLOSED and the answer embarrasses:
**BuildFlow's result cache**, not a stale binary. Three falsification probes
(binary version, comma parsing, toolspec bridge) all came back clean before the
decisive check — `BUILDFLOW_NO_RESULT_CACHE=1` reads **12 == CLI** — and the
flag was documented in AGENTS.md's Commands block the whole time. Everything
else on the tail landed: treefmt red on main healed and pushed (CI GREEN on
`99953a95`), full test suite green (35 packages), CHANGELOG entries, the
upstream draft written with a FRESH end-to-end repro (new fact discovered: the
poison needs an ODD `/*` count — parity, not presence), docs-gates healed to
full PASS (two findings in my own report fixed, the concurrent session's §f row
harvested), AGENTS.md gained the directive convention + the two gotchas, and
the full buildflow run passed (73 steps, exit 0).

---

## a) FULLY DONE (executed AND verified)

### The delta investigation (the turn's centerpiece)

1. **Four probes, all logged**: (i) `go version -m` on the installed binary —
   it embeds go-cqrs-lite cqrs-lint **v4.21.0**, the same version as the CLI,
   killing "stale embed" immediately; (ii) a scratch module in `/tmp/commaprobe`
   calling v4.21.0's `ParseSuppressions` directly — the comma form
   `ignore(C015,C023)` resolves BOTH rule IDs; (iii) the toolspec chain
   BuildFlow actually runs (`ToolFromSpec` → `Spec().Detect`) invoked directly
   against this tree — **12 findings**; (iv) the PATH binary
   (`~/go/bin/cqrs-lint`, a v4.20.1-0.20261107 pseudo-version) run standalone —
   **12 findings, exit 0**.
2. **The decisive check**: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s
   cqrs-lint` — **12, identical to the CLI**. The 18 was a cached step verdict
   from a pre-fix tree state. Re-verified at 01:23 (still 12). Nothing to
   rebuild, nothing to rewrite.
3. **Every stale claim corrected at its source**: the 00-36 report got a dated
   RESOLVED banner above the verdict paragraph, plus in-place corrections to
   its §b1 (open tail), §e3 (directive-convention gotcha), §f3 (immediate
   item), §f26 (rebuild rationale), and §g Q1 (the owner question answered by
   evidence). `.buildflow.yml` was checked and needed NO correction — its
   triage note never carried the stale hypothesis.

### The verification sweep (the tail the 00-36 report owed)

4. **`nix fmt`**: repaired 13 files — the daemon had committed pre-fmt bytes
   (the trailing `//cqrs-lint:ignore` comments' gofmt alignment). The format
   authority (`scripts/format-check.sh`, the treefmt-check drv) passed.
5. **Main's treefmt red HEALED**: the four consecutive red CI tips
   (00:00–00:00 window) failed treefmt-check on exactly those bytes; the fmt
   repair was the fix, pushed under the ratified manual-push bar (arm 1: red
   main + verified fix; anchor: the local green drv).
6. **Full test suite**: `nix develop -c env -u GOEXPERIMENT go test -count=1
   ./...` — exit 0, 35 packages green, zero FAIL lines.
7. **The pbx rename's test ran for real** (not just compile):
   `go test ./internal/pbx/` green — `UpstreamURL` keeps query and path
   behavior under its new name.
8. **CHANGELOG**: two `[Unreleased]` § Changed entries — the decider-hygiene
   train (Version.Increment arithmetic, schema-version stamps, command-carried
   `At` stamps, `decider.StrictApply` folds, exact-name JSON tags on the
   journaled tombstones — argued as replay-safety, not behavior change) and the
   `pbx.Client.UpstreamURL` rename.
9. **Full buildflow run**: EXIT 0 — 73 steps success, 0 failed, findings gate
   green (46 accepted via the baseline). Timing regressions in the tail are
   advisory (machine under load; the test-race +202% needs a quiet host).

### The upstream draft — with a NEW empirical fact

10. **The draft is written**: `~/projects/BuildFlow/docs/drafts/
    cqrs-lint-suppression-block-comment-state-machine.md`, in the established
    D005 draft format (note-header, symptom, root cause with cited function,
    smallest-correct-fix sketch, environment).
11. **The repro was rebuilt from scratch and it taught something new**: the
    first fixture with `e.g. image/* and video/*` did NOT poison (both runs
    suppressed) — because two `/*` on one line is an EVEN count and the state
    machine toggles on odd parity. ONE `/*` (the real incident's shape)
    poisons: C015 fires, the directive goes inert, zero diagnostics. Clean
    run: `1 finding(s) suppressed`. This parity fact is now the draft's
    sharpest evidence — it rules out "strip all `/*` from prose" fixes and
    points squarely at the counting logic.
12. **Bug confirmed live at go-cqrs-lite HEAD** (`3d267ffdb`):
    `computeBlockCommentLines` is unchanged from v4.21.0 — the parity toggle
    counts raw `/*`/`*/` substrings with no line-comment awareness, while the
    adjacent `commentTextStart` lexer already knows exactly where comment text
    begins. The fix sketch (strip from `commentTextStart` before counting) is
    in the draft.

### Gates, CI, docs

13. **docs-gates healed to full PASS** — and both red gates were MY OWN
    residue: (i) my multiedit had mangled the report's verdict paragraph into
    ``` ``.buildflow.yml`. The` ``` (a broken code span) — restored; (ii) my
    §g Q1 strike used "ANSWERED 2026-11-11 00:55:" as its verdict, which the
    scanner reads as a struck item with NO verdict marker — rewritten into the
    house grammar `done (answered 2026-11-11 00:55: …)`. Battery verdict:
    status-f-section PASS, battery PASS, table-damage PASS, strike-safety PASS
    (410 files safe).
14. **The concurrent session's §f row 1 harvested**: its "Answer §g Q1–Q3" row
    was the freshness gate's last red — struck with a routed verdict citing
    TODO_LIST §Release-green + B5-staged (which that session had already
    written) and answering its Q2 directly: the treefmt red was this train's
    pre-fmt bytes and THIS session healed them.
15. **The status index regenerated** (66 reports) and `status-index --check`
    exits 0.
16. **CI GREEN on the fixed tip**: run 42930726762 (commit `99953a95`)
    completed success at ~01:29. Watch lanes green throughout. At 01:58 the
    daemon's docs commit (`47d69d8e`, the §f15 strike) is pushed with run
    42935283945 in flight.
17. **AGENTS.md gained two additions** (203/377 lines, both into existing
    bullets): the cqrs-lint directive convention (comment text must START with
    `cqrs-lint:`, own line or immediately-preceding non-blank line,
    comma-lists fine, C044's prefixed-key form is a DECLINED architecture via
    `.cqrs-lint.json` — never "fix" `id.ID.String()` keys to bare values) plus
    the `/*` landmine with the draft pointer; and the BuildFlow result-cache
    gotcha with the baseline scoped-run noise note.
18. **The `/*` landmine sweep (00-36 §f item 15) executed and recorded**: six
    files carry one-`/*` line comments (`/hooks/*`, `/api/*`, `app/*.js`
    prose) — but ZERO live directives sit below any of them. assets.go's C009
    directive sits ABOVE its poison line; the other five files carry no
    directives. Loaded landmines for future directives only; struck done in
    the 00-36 report with the full file list.

---

## b) PARTIALLY DONE

1. **Both upstream drafts are drafts, not filings.** The state-machine draft
   (new) and the D005 draft (pre-existing) sit in BuildFlow `docs/drafts/`
   awaiting the owner review tick (the note-header's unchecked
   MANUALLY REVIEWED box) before becoming actual issues/PRs upstream.
2. **One docs commit in flight**: `47d69d8e` (the §f15 strike) is pushed with
   CI running as of 01:58. Docs-only, annotate verified 203 locally, freshness
   can only improve from a strike — no red risk identified, but the verdict
   was not yet read at write time.
3. **The AGENTS.md erraudit bullet still carries the inline date "next
   2026-11-05"** while the watch calendar (the ONE home for cadence, per
   TODO_LIST) already re-measured on 2026-11-08 with next due 2026-12-08. A
   two-line fix skipped to avoid churning a dense bullet mid-verification;
   the calendar owning the date means the drift is cosmetic.
4. **The binary-freshness preflight warning stands**: the buildflow binary
   (`168e491`) predates repo HEAD but is functionally current for cqrs-lint —
   the rebuild is an owner infrastructure call, not a correctness need.

---

## c) NOT STARTED (adjacent, deliberately not begun)

1. Filing either draft upstream (blocked on the owner review gate).
2. BuildFlow rebuild + reinstall (`nix build . && nix run .#reinstall` in
   ~/projects/BuildFlow) — clears the preflight warn, nothing else.
3. The test-race +202% timing regression chase (advisory; needs a quiet host
   to attribute).
4. `buildflow format` to clear the templ fingerprint drift advisory
   (0.3.1020 → 0.3.1100, expected after the templ wave).
5. The dprint-remote-plugins advisory decision (nix-store pin vs accept).
6. The 00-36 report's §f items 11–50 (days/week watchers) — unchanged ledger,
   still owned there.
7. The concurrent session's B5 prod-switch authority (its §g Q1) and the E3
   budget number (its Q3) — owner legs, not mine to execute.
8. The 00-36 report's §g Q2/Q3 (cqrs-lint gating posture; C044 permanence) —
   restated as questions below.

---

## d) TOTALLY FUCKED UP (own mistakes, in order of severity)

1. **The answer was in AGENTS.md the whole time.** `BUILDFLOW_NO_RESULT_CACHE=
   1` is documented in this repo's Commands block ("for full"), and the
   result-cache trap is the FIRST thing a scoped-step anomaly should have
   ruled out. Instead I ran three elaborate falsification probes (binary
   version, parser probe, toolspec bridge) over ~20 minutes before the cache
   check — repeating the 00-36 report's own §d3 lesson (edited before reading
   the tool source) with a different costume: debugged tool behavior before
   reading the house gotchas. The probes were not wasted (they make the cache
   verdict bulletproof), but the ORDER was wrong and the cost was real.
2. **My multiedit shipped a typo into a committed, pushed file.** The
   four-edit batch on the 00-36 report mangled the verdict paragraph into
   ``` ``.buildflow.yml`. The` ``` — a broken code span sitting in main for
   one daemon-commit cycle. I verified "edit succeeded" and never re-read the
   rendered span. The annotate gate caught it (F/spacing), which is the system
   working, but the class is mine: batch edits without proofreading.
3. **Three wasted repro cycles on the fixture.** (i) First fixture had no
   receiver — not valid Go, silently unlintable; (ii) second had no
   go-cqrs-lite import — the consumer guard refused it ("none import
   go-cqrs-lite"); (iii) third used two `/*` on one line — even count, no
   poison, and I nearly concluded "cannot reproduce" against the documented
   incident. Each was a fact I could have read first (build the fixture; read
   the guard; count the pattern occurrences). The parity fact was found by
   ACCIDENT (the clean run also suppressed, which made no sense until the
   count was checked) — a lucky catch, not systematic thinking.
4. **I never read the failed run's actual log.** Run 42929376731 (tip
   `57df1bd3`) failed; my `--log-failed` grep returned EMPTY and I moved on to
   pushing the fix without ever establishing that run's true failure cause
   (docs-gates was the plausible inference — annotate 205 was pre-existing per
   the concurrent session's evidence index — but it was never CONFIRMED for
   that run). The fix was verified independently (the next tip went green), so
   no damage, but the diagnosis discipline was skipped.
5. **docs-gates was not in my verification sweep at all.** I ran format
   authority, full tests, scoped buildflow — and only found the red gates by
   accident while checking AGENTS.md's line count. My edits (CHANGELOG
   paragraphs, AGENTS growth, a new report, in-place annotations) are EXACTLY
   the corpus growth the AGENTS.md freshness warning says flips these gates,
   and two of my own edits CREATED new findings. The battery belongs in the
   per-turn sweep the moment any living doc changes, not at the end.

---

## e) WHAT WE SHOULD IMPROVE (process, not code)

1. **Re-read the house gotchas BEFORE debugging tool behavior.** AGENTS.md's
   Commands block is the repo's accumulated scar tissue; the cache flag lived
   there for weeks. A 30-second grep of AGENTS.md for the tool's name should
   precede any hypothesis ladder about its internals.
2. **Run docs-gates immediately after ANY living-docs growth** (CHANGELOG,
   AGENTS, new/edited status reports) — not as a sweep-day ritual. The DF
   brittleness means the gate's verdict is valid only against the corpus it
   measures; measuring after the growth is the only honest reading.
3. **Proofread batch edits**: re-View the edited span after any multiedit
   before declaring done. The edit tool guarantees the replacement happened;
   it cannot guarantee the replacement is SANE.
4. **For upstream repros: build the fixture first, read the guard second,
   count the pattern third.** A repro that "shows nothing" is a fixture bug
   until proven otherwise — and my fixture was wrong twice before the bug was
   even reachable.
5. **The scratch-probe discipline paid for itself again**: calling
   `ParseSuppressions` and `Spec().Detect` directly isolated the layer in
   minutes where reasoning about wrappers had stalled. Keep "import the
   library, call the seam" as the default move whenever a wrapper (BuildFlow,
   a pipeline, a provider) is suspected.
6. **Attribute new annotate findings at once**: after living-docs edits, a
   quick `docs-gates annotate . --annotate-max=<pin>` names file:line owners
   immediately — the two findings this turn created were trivially fixable
   once named, and would have been invisible until the next full sweep.

---

## f) Up to 50 things to get done next (prioritized)

**Immediate (this train, unverified tails):**

1. ~~Re-run `nix fmt` + format authority~~ done (13 files repaired,
   treefmt-check drv green).
2. ~~Full `go test -count=1 ./...`~~ done (35 packages, exit 0).
3. ~~Resolve the 6-finding BuildFlow delta~~ done (result cache; NO_CACHE
   reads 12 == CLI; four probes logged).
4. ~~CHANGELOG entries~~ done (decider-hygiene train + rename).
5. ~~Upstream draft~~ done (with the parity repro; lives in BuildFlow
   docs/drafts/).
6. ~~CI watch~~ done (GREEN on `99953a95`; the docs-only `47d69d8e` push has
   CI in flight as of 01:58 — read its verdict when it lands).
7. ~~AGENTS.md additions~~ done (directive convention, landmine, cache
   gotcha, C044 cross-ref).
8. ~~status-index regen~~ done (66 reports; `--check` exit 0).
9. ~~Full buildflow at train end~~ done (EXIT 0, 73 steps).
10. ~~`/*` landmine sweep~~ done (six files, zero live directives exposed;
    recorded in the 00-36 report).

**Follow-ups on this triage (days):**

11. File the suppression state-machine draft upstream as the actual
    issue/PR (gated on the owner review tick).
12. File the D005 draft upstream (same gate; it predates this session).
13. Track the upstream `/*` lexer fix; on the next cqrs-lint bump, remove the
    actions.go hedge comment if the fix shipped.
14. Re-check the enroll.go S001 removal: confirm the heuristic change is
    intentional upstream (the wirefixtures twin still fires).
15. Split the two C014 rulings' rationales in `.buildflow.yml` (app.go
    boot-time zone vs reports.go deployment-zone windows).
16. Run `erraudit nolint-audit` after the C007/At-command train (the decay
    lesson says run it after error-adjacent refactors) — NOTE: the watch
    calendar already owns the quarterly erraudit re-measure (next due
    2026-12-08); this item is the train-specific decay pass, not the
    calendar's.
17. Verify the D014 tags with a replay test: round-trip an OLD-shape journal
    row through DecodePayloadAuto (byte-neutrality is argued, not pinned).
18. Add a comment cross-referencing the `.cqrs-lint.json` C044 reason beside
    the ids_test.go prefix pin (stops the next reader re-deriving the wrong
    first analysis).
19. `cqrs-lint scorecard` before/after: confirm the disabled F-series did not
    distort the adoption score.
20. `cqrs-lint doctor` once: read the rendered disable-reasons as an operator
    would at six months.

**Doctrine/duty items (this week):**

21. Write the triage decision table (config-disable / directive /
    keep-visible / fix) into one home (`.buildflow.yml` or AGENTS.md).
22. Move the scoped-verifier/baseline lesson (00-36 §d1) into
    docs/lessons.md.
23. Move the read-the-pins-first lesson (00-36 §d2) into docs/lessons.md.
24. Add the result-cache story (three falsified hypotheses, one documented
    flag) to docs/lessons.md — AGENTS.md carries the rule; the lesson carries
    the war story.
25. Pin the B007 ruling with a comment at the server.New registration block
    ("explicit registrations ARE the table") so nobody loops it.
26. Consider feeding the per-aggregate stream-vocab vars to the catalog
    (B010's catalog.Event duplication) — the lists exist as data now.
27. Fix the AGENTS.md erraudit bullet's stale inline "next 2026-11-05" date
    (cosmetic; the calendar owns the real date).

**Upstream/tooling (week+):**

28. Rebuild + reinstall BuildFlow (freshness only; owner infrastructure call).
29. After any rebuild: re-review `buildflow timings --regressions` on a quiet
    host (test-race +202% attribution).
30. `buildflow format` for the formatter-fingerprint drift.
31. Decide the dprint-remote-plugins advisory (nix-store pin vs accept).

**Longer-lived watchers (reopen triggers):**

32. F005: first upcaster at the first v2 event (the `.cqrs-lint.json` reason
    names it).
33. C044: prefixed-key migration stays PRICED AND DECLINED; reopen only via
    that train.
34. D005: re-enable when the first-token misattribution fix ships upstream.
35. C009: any NEW panic finding is a real signal (the boot-time class is
    fully directed) — keep it that way.
36. B005: new folds ship with fold + stream-vocab var together; consider an
    arch-test that every RegisterDecider Apply is StrictApply-wrapped.
37. F031: new Scans without explicit WithLimit are real signals now.
38. S001: watch whether the wirefixtures directive survives future heuristic
    changes (enroll's did not).
39. The `/*` landmine: any NEW directive placed below a one-`/*` line comment
    in the six swept files (or anywhere) is dead on arrival until the lexer
    fix lands — the sweep result in the 00-36 report names the files.

**Repo hygiene (opportunistic):**

40. The 12 visible findings' `.buildflow.yml` prose note: if it grows again,
    move per-rule verdicts into `.cqrs-lint.json` severity-overrides.
41. Confirm `internal/server/crm.go`'s concurrent-session edit landed (it
    predates this session; still not mine to touch).
42. Pin `.cqrs-lint.json` root-relative discovery into the BuildFlow step
    docs (a moved invocation cwd would silently drop all rulings).
43. Verify the lychee `exclude_path` entries take effect (run release.sh
    step 6 once, or lychee directly).
44. Confirm the pbx UpstreamURL doc comment's dedup-playbook reference
    satisfies the preflight inventory on the next run.
45. Decide whether the renamed pbx URL test should also pin the no-query
    path (the old test pinned only the query case).
46. Sweep AGENTS.md for now-stale cqrs-lint sentence numbers ("11
    non-gating findings" era) — the corpus is 12 deliberate.
47. The `.buildflow.yml` 2026-10-15 note's "11 findings" lead-in is
    historical context under the 2026-11-10 paragraph — fine; check tone at
    the next touch.
48. Docs-health sweep owed for this turn's report (this file's own §f rows
    are harvested into TODO_LIST in the same push — see the harvest block).
49. When the state-machine fix lands upstream, delete the actions.go hedge
    comment IF the lexer fix also shipped (two separate upstream fixes).
50. Never re-litigate the 00-36 report's rulings silently — its §f/§g rows
    and this report's carry the reopen triggers.

---

## g) Questions I cannot answer myself

1. **File the drafts upstream NOW?** Both drafts (the `/*` state machine and
   the older D005 misattribution) sit behind the MANUALLY REVIEWED checkbox —
   your convention is drafts wait for your review tick. The verify-before-
   filing gate is fully discharged on my side (source-level root cause +
   end-to-end repro on the released binary). Do you want me to file either as
   an issue in this session, or do they wait for your review?
2. **Rebuild + reinstall BuildFlow now?** It is fleet-wide tooling
   (`nix build . && nix run .#reinstall` in ~/projects/BuildFlow), the delta
   does NOT depend on it (the cache verdict stands on v4.21.0 evidence), and
   its only yield is clearing the preflight freshness warning. Do you want it
   done in a session, or is it yours / a concurrent session's?
3. **The two standing owner questions from the 00-36 report — answer now, or
   TODO_LIST rows?** (a) cqrs-lint gating posture: with noise at zero and the
   pile having regrown 11 → 186 in one train, does the classification-note
   system stay the permanent posture, or does gate-when-available get a
   TODO_LIST owner? (b) C044: is the prefixed-key form the permanent ruling
   (never migrate internal keys to bare values), or do you want a TODO_LIST
   row carrying the migration pricing so it surfaces at planning time?
