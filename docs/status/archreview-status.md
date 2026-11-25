# Architecture review #7 — session status

> Claims verified as of **01:04 CET 2026-11-25** (push/CI statements rot
> fastest). Session: one order, "review this project and create a .html
> report from a Software Architect perspective", executed with the
> architecture-review skill as series #7; the sitting ran from 2026-11-24
> into 2026-11-25. This file first landed as an uncorrected draft
> (`53686f32`, 2026-11-24 23:31) and was corrected in place in the next
> sitting; §d says what the draft had wrong.

> **Scope.** Verified here, each with the command that produced it: the
> package graph (`go list -json ./internal/... ./cmd/...`: 38 packages, 149
> internal edges, 0 cycles) and the full test suite with a union coverage
> profile (`go test -count=1 -coverpkg=./internal/...,./cmd/... ./internal/... ./cmd/...`:
> rc 0, 37 packages ok, 0 failed), both at `c592dd22` on 2026-11-24;
> `git diff --stat c592dd22..HEAD -- '*.go'` is empty, so they hold for every
> later commit; the arch suite (15 PASS); the review report's integrity
> checks on the final file (`check.py`: balanced tags, 26 unique ids, 8
> anchors; `recount.py`: 144 of 144); renders at 1440px (eleven slices),
> 820px and 390px (four slices), plus a text-level overflow probe at six
> widths (320 to 1440); `scripts/ci-verdict.sh --self-test` (12 of 12) and
> four live shas; `dprint check` rc 0; five docs-gates pre-flights, all
> SAFE; `scripts/push-stall-watch.sh --no-fetch`;
> `scripts/docs-gates.sh --check` (PASS at 22:53 and 22:57 on 2026-11-24,
> and on the landed tree at 01:04 on 2026-11-25, every gate PASS); `python3 scripts/status-index.py --check` rc 0; the git
> state read at 01:04 (`git ls-remote origin main`,
> `git reflog show origin/main --date=iso`,
> `gh api repos/LarsArtmann/webphone/actions/runs/45501299454`).
>
> **NOT verified here:** CI for the commits that carry this report and the
> script fix (unpushed at report time; `scripts/ci-verdict.sh` answers
> NO-RUNS until the daemon pushes); `nix flake check`, buildflow, smoke and
> workflow-e2e (a docs session plus one shell-script fix that no CI job
> runs); the T08 dependency train (not opened); any number from reports #1
> to #6 beyond the ones re-measured; the 2026-11-19 D2 diagrams (not
> regenerated); a phone narrower than 390px (the stat cards overflow there,
> d-20).
>
> `git rev-parse --short HEAD` at report time: `77e29b38`.

## a) FULLY DONE

1. **Architecture review #7 report** —
   `docs/architecture-understanding/2026-11-24_22-51_architecture-review.html`,
   spliced into the editorial template with `splice-report.sh` (the
   stylesheet is byte-identical to the template; only `<title>` differs).
   Seven dimensions held at 4.36; seven new findings F6 to F12; a
   seven-item roadmap with a table of what would move the two lowest
   scores. The first final splice (95,562 bytes) is at `e23d35ac`, byte
   for byte (`cmp` with `/tmp/archreview/out.html`, 23:14); the phone-width
   amendment (a-6) is 96,101 bytes, 2,622 lines.
2. **Evidence the series did not have before** — the hub's mass by file
   family (the 742-line machine-contract family is a leaf: 0 references
   out, 1 in); AST in-package coupling for seven packages (10 of 11 cqrs
   aggregates lean on an eight-name kernel and nothing aggregate-specific
   from each other); the depth-pin ledger from git history (4 raises, 0
   lowers in 12 days; server at 5300 of 5300); a fresh union coverage pass
   (76.2% of 18,052 statements, 82.6% of the 9,014 hand-written ones, three
   zero-coverage files); the first client-side measurement (15,892 raw
   lines, no size governor); SQL confinement (store 10, session 9, cmd 1,
   zero elsewhere); feature ripple of 1 to 4 packages across six features;
   a concept-to-module map for five glossary concepts.
3. **Series index** — one row in `docs/architecture-understanding/INDEX.md`
   (`bc230306`); `dprint check` rc 0 after the edit.
4. **TODO harvest of the review's roadmap** — seven rows under "2026-11-24
   architecture review #7 harvest" at the top of `TODO_LIST.md`
   (`a27227ca`; one row reworded in `bc230306`); both pre-flights SAFE.
5. **Report-integrity checklist** — recount 144 of 144 against the measured
   data, re-run on the final file; balanced tags, 26 unique ids, 8 anchors
   resolving, no inline style or handlers; every cited symbol
   grep-verified and every proposed name labelled; all eleven desktop
   render slices, one 820px render and four 390px slices inspected.
6. **Phone-width check of the review report** — a text-level probe at six
   widths found one "Where" line (F9) widening the page to 503px at 390px;
   six `<wbr>` break opportunities fixed it (the probe reads clean at 390,
   820, 1024 and 1440). The footer carries a dated amendment note.
7. **Docs battery on the changed tree** — "PASS — formatted + all doc gates
   green" at 22:53 and 22:57 on 2026-11-24, and on the landed tree at 01:04 on 2026-11-25, every gate PASS.
8. **Full test suite green** as a byproduct of the coverage pass: 37
   packages ok, 0 failed.
9. **Push and CI for the review chain** — the daemon pushed the three
   review commits at 23:28:43 on 2026-11-24, 35 minutes after the first;
   CI run 45501299454 concluded success at 23:40:11;
   `scripts/ci-verdict.sh e23d35ac` answers GREEN.
10. **`scripts/ci-verdict.sh` zero-run fix** — a sha with no workflow runs
    answered rc 3 with a traceback instead of the documented rc 1 (found
    23:33 on 2026-11-24, reading the verdict of this report's own commit).
    Root cause: the paging rewrite of that day left a page's stats line in
    the row list. Fixed; self-test 12 of 12, the new case proven failing
    before the fix; live rc 1 on an unpushed sha; CHANGELOG `[Unreleased]`
    line.
11. **Stale numbers in AGENTS.md corrected** — the host-go sentence (host
    go read go1.27.3 = go.mod; the text said "below the 1.27.1 floor"),
    `audit_max` 72 → 74, cadence 65 → 66, each with its commit.
12. **TODO_LIST front matter aligned with the strike canon** (owner ruling
    2026-11-11): it said done work is deleted; 172 struck rows and AGENTS.md
    say struck with evidence.
13. **`.docs-gates-preflight.json` is git-ignored** — the tool's own note
    asks for it; the daemon would commit a forgotten ledger.
14. **This report's open rows harvested** — eight TODO rows and one ROADMAP
    standing watch; five pre-flights SAFE.
15. **Nothing stray in the repo** — the pre-flight ledger was trashed after
    every run; all scratch (43 MB) stayed under `/tmp/archreview`.

## b) PARTIALLY DONE (in flight)

1. **This report's own push and CI verdict** — At 01:04, 2 commit(s) sit unpushed, the oldest `53686f32` (23:31, 93 minutes old), the newest `77e29b38`; `origin/main` reads `e23d35ac` and `scripts/push-stall-watch.sh --no-fetch` reads STALL.
   The manual-push bar opened at 00:31:05 on 2026-11-25 (60 minutes after
   `53686f32`) and was not invoked: the harness forbids a push without an
   explicit instruction, and the order was to wait. Remaining after that
   reading: the push, then `scripts/ci-verdict.sh <sha>` to GREEN. Effort S.
   Blocker: none technical.
2. **Narrative commit** — none was written (the harness forbids commits
   without an explicit request), so the review chain rides three heuristic
   "chore: auto-commit" commits and this report rides more. The review
   report carries dated amendment notes because amendments followed its
   first splice.
3. **F12 ("hub" carries three meanings)** — in the review report, bound to
   the ROADMAP DOMAIN_LANGUAGE row by reference; the ROADMAP text was not
   edited.
4. **This report's name** — `docs/status/archreview-status.md` (the owner's
   path) sits outside the `2*.md` report glob: `docs/status/INDEX.md` does
   not list it, the docs-gates battery does not read it, and the wrapper's
   `status-f-section` gate picks it as "newest" because it sorts every
   `*.md`. It is rename-ready: a convention-named copy of this file read "0
   unharvested" in a scratch clone. The file stays at that path until you
   say otherwise.
5. **Series D2 pair** — not refreshed (not requested); the 2026-11-19 pair
   predates two edges. Now an open TODO row.
6. **Review tooling** — the AST coupling tool, the coverage merge and the
   recount harness exist only under `/tmp/archreview`; the row "Persist
   the review tooling in-repo" is open and the scratch is kept until it
   lands.

## c) NOT STARTED (the review's code work; the review changed no Go code)

- F6: the owner ruling, then the `internal/server/contracts` carve.
- F6: the two ratchet clicks (store pin 1900 to 650; server/platform named
  pin).
- F8: the session count behind `session.Store`, `DB` out of
  `platform.Deps`, the arch pin.
- F7: the four client-side pins.
- F9: one boot recipe, one aggregate list, two app-owned one-shot builders.
- F11: the cqrs kernel file.
- F10: the rewording of the standing test-drain row.
- The T08 tail (m08.6 to m08.8) from the 07-23 report: untouched, its state
  unknown here.

All of these are harvested (rows in `TODO_LIST.md`), none is started; none
is blocked except F6, which waits on the ruling in g-1.

## d) TOTALLY FUCKED UP (honest list; mine unless marked "noticed")

1. **A wrong number nearly shipped.** A first coverage script kept the
   first-seen profile block per location instead of the maximum and
   reported `ids.go` at 24.3% (it is 100%). The recount caught it within
   minutes, but the first splice was already committed (22:53). Root
   cause: a `-coverpkg` profile carries one block per test binary.
   Mitigation: the recount harness, 144 checks, re-run on the final file.
2. **Regex coupling gave three false references.** Counting identifiers
   with a regex read `events` and `lang` out of string literals and
   comments; the machine-contract family first measured "3 references from
   the family into the rest of the hub". A go/parser and go/scanner tool
   re-measured it: 0 out, 1 in. Caught before it reached a report file.
3. **A raw-SQL grep counted `r.URL.Query()`.** It reported 25 SQL call
   sites in the server root; the real count is 0 (store 10, session 9, cmd
   1). Caught before citing.
4. **A corrected table cell left a stale sentence.** The hub table said 4
   names for textagent_api.go after the AST tool; a prose sentence kept
   "17 names" until the render review.
5. **Over-claims fixed only after the first splice:** "every #6 claim
   re-measured" (structural claims only), "the first data read" (the one
   data read), a `serve` subcommand (there is none: the default boot,
   `-enroll-passkey`, `catalog`, `contracts`, `export-legacy`), two quotes
   that were not verbatim, and "the precedent is api and hooks" (platform
   carried 5 test files on its carve day too). Each was corrected and the
   review report's footer lists the amendments.
6. **A TODO row shipped with a wrong title.** "Test-drain row precedent is
   wrong" landed in `a27227ca`; the truth is "half right" (platform yes,
   auth no). Reworded in `bc230306`, re-flighted in replace mode.
7. **Amendments after the daemon had committed.** Three commits carry one
   review artifact, and it carries two amendment notes. Root cause: the
   first splice went into `docs/` before the recount and the read-through.
   The 07-23 report's own lesson ("work everything in clones; the daemon
   commits any dirty file within minutes") was in the repo and was not
   applied to the report artifact.
8. **I blamed the tool for my typo.** A one-character mistake at a call
   site (`os.path_is_dir` for `os_path_is_dir`) cost four tool rounds while
   I suspected the bash tool of mangling output; a probe at 23:24 shows no
   mangling.
9. **The series lost its headline metric for this window.** Per-commit
   co-change (#6: 57.8% single-cluster) is unmeasurable here: 39 of the 49
   Go-touching commits are daemon sweeps. Ripple by identifier (1 to 4
   packages) replaced it; the two numbers are not comparable.
10. **The tooling is ephemeral.** The AST coupling tool, the coverage merge
    and the recount harness exist only under `/tmp/archreview`; series #8
    re-derives them unless the row "Persist the review tooling in-repo"
    lands.
11. **This report's first landing was an uncorrected draft.** The draft
    went into the repo before its correction list was applied and the
    daemon committed it as `53686f32`. It claimed twelve desktop slices
    (eleven), a `cmp` at 23:15 (23:14), and three unpushed commits that the
    daemon had pushed at 23:28:43, two minutes before that commit. It also
    bound one item to a TODO row that was already struck. Corrected in
    place here; the struck row's item is a new TODO row now.
12. **The review report overflowed on a phone.** The draft listed a
    phone-width render as "not verified"; doing it found one "Where" line
    (F9) widening the page to 503px at 390px. Fixed with `<wbr>` (a-6).
    Eleven desktop slices and one 820px render could not show it: the
    checklist's one-render step was too weak.
13. Noticed, fixed: **`scripts/ci-verdict.sh` crashed on a zero-run sha**
    (the 2026-11-24 paging rewrite shipped without a zero-run self-test
    case; a-10).
14. Noticed, fixed: **the TODO_LIST front matter contradicted the strike
    canon** (a-12).
15. Noticed, fixed: **three stale numbers in AGENTS.md** (the draft said
    two; cadence 65 was the third) (a-11).
16. Noticed, fixed: **the pre-flight ledger was not git-ignored** (a-13).
17. Noticed: **the daemon's push leg waited 35 minutes** (22:53:18 →
    23:28:43) with three commits waiting. Under the 60-minute bar, so no
    issue was filed; TODO_LIST § "Daemon push-stall forensics" and the
    ROADMAP "Push-lag policy" question own the topic. At 01:04, 2 commit(s) sit unpushed, the oldest `53686f32` (23:31, 93 minutes old), the newest `77e29b38`; `origin/main` reads `e23d35ac` and `scripts/push-stall-watch.sh --no-fetch` reads STALL.
18. Noticed: **two definitions of "newest report".** `scripts/docs-gates.sh`
    (`status_f_gate`) sorts every `docs/status/*.md`; the battery and
    `scripts/status-index.py` read `report_glob` (`2*.md`). The two agree
    on a convention-named corpus and split on this file's name. Harvested.
19. Noticed: **the freshness gate passes rows on accidental token
    overlap.** A convention-named copy of the draft read "0 unharvested"
    while `--explain-item` showed rows that match nothing they meant
    ("desktop slices", "trash"). The gate cannot tell a harvest from a
    coincidence, so the harvest here was checked by hand, row by row.
20. Noticed: **the stat cards of the kit's editorial template cannot
    shrink below 390px** (six-character values at 2.6rem in a forced
    two-column phone layout); they overflow by 30px at 360px. Harvested
    (repo: SKILLS).
21. Noticed: **the status-report skill contradicts itself** — its Output
    section lists four HTML steps under a `.md` default. Harvested
    (repo: SKILLS).
22. **The landing script forgot the fixed script.** The one-call landing
    copied every rehearsed file except `scripts/ci-verdict.sh`; the
    byte-parity check against the rehearsal clone caught it, and the copy
    landed at 00:58, before any commit. A landing needs a parity check,
    not a checklist.

## e) WHAT WE SHOULD IMPROVE

1. **Stage a report outside the repo and drop it in once.** Body, splice,
   recount, read-through and render all under `/tmp`; one copy at the end.
   The lesson applied twice tonight: the review needed three commits, and
   this report's draft landed before its correction list was applied.
2. **Recount before the first splice, not after.** The recount exists; its
   place in the sequence was wrong.
3. **Persist the review tooling in the repo.** Three small programs would
   make series #8 a re-run instead of a re-derivation.
4. **AST or scanner for any coupling claim.** Regex over source counts
   strings and comments; it was wrong twice tonight and right never.
5. **A generated table of contents for `docs/architecture-understanding/`.**
   It is the only report directory with a hand-maintained one.
6. **A text-level overflow probe in the report checklist.** One run at six
   widths found in seconds what eleven desktop slices and one tablet
   render could not.
7. **The architecture itself, in priority order (all harvested):** decide
   the hub pin's third raise with the contracts seam on the table (F6); put
   a ceiling on 15.9k ungoverned client lines (F7); close the
   readiness-only DB rule and pin it (F8); give the journal one boot recipe
   and one aggregate list (F9); correct the test-drain precedent (F10);
   name the cqrs kernel (F11); settle "hub" with the glossary call (F12).
8. **Daemon push lag.** Two reports in one day describe a slow or dead push
   leg; the review chain's 35-minute lag stayed under the bar, and the next
   batch crossed it (b-1). The ROADMAP "Push-lag policy" question decides
   whether a watchdog is due.
9. **A zero-row case in every script's self-test.** The paging rewrite was
   verified against a flooded listing and never against an empty one.
10. **One rule for "which report is newest".** `report_glob` in three
    tools, not two.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Landed this session (struck here so the ledger stays honest):

1. ~~Write architecture review #7 as an HTML report~~ → done 2026-11-24 — `docs/architecture-understanding/2026-11-24_22-51_architecture-review.html`, first final bytes at `e23d35ac`.
2. ~~Verify #6's roadmap and re-measure its claims~~ → done 2026-11-24 — report sections 01 and 03; 38 packages, 149 edges, 0 cycles; 4.36 held.
3. ~~Add the series INDEX row~~ → done 2026-11-24 — `docs/architecture-understanding/INDEX.md` (`bc230306`).
4. ~~Harvest the review roadmap into TODO_LIST~~ → done 2026-11-24 — seven rows, both pre-flights SAFE (`a27227ca`, `bc230306`).
5. ~~Run the docs battery on the changed tree~~ → done 2026-11-24 — `scripts/docs-gates.sh --check` PASS at 22:53 and 22:57.
6. ~~Run the report-integrity checklist~~ → done 2026-11-25 — recount 144 of 144, parse, anchors, eleven desktop slices, 820px and 390px renders.
7. ~~Fix the review report's phone-width overflow~~ → done 2026-11-25 — six `<wbr>` in one Where line; overflow probe clean at 390, 820, 1024 and 1440.
8. ~~Push and CI-watch the review commits~~ → done 2026-11-24 — daemon push 23:28:43; CI run 45501299454 success; `scripts/ci-verdict.sh e23d35ac` GREEN.
9. ~~Fix `scripts/ci-verdict.sh` for a sha with zero workflow runs~~ → done 2026-11-25 — self-test 12 of 12; live rc 1 on an unpushed sha.
10. ~~Correct the stale AGENTS.md numbers~~ → done 2026-11-25 — host-go sentence, `audit_max` 74, cadence 66.
11. ~~Align the TODO_LIST front matter with the strike canon~~ → done 2026-11-25.
12. ~~Git-ignore `.docs-gates-preflight.json`~~ → done 2026-11-25 — the tool's own note asked for it.
13. ~~Harvest this report's open rows~~ → done 2026-11-25 — eight TODO rows and one ROADMAP standing watch; five pre-flights SAFE.

Open, in execution order (each names its TODO_LIST or ROADMAP home):

14. **Owner ruling on the server pin's third raise** — carve `internal/server/contracts` (apidocs, wirecontracts, wirefixtures; 742 LOC) as a fifth surface carve, or ride with the runway note naming the seam. TODO row "Server pin full: decide the third raise". Impact Critical, effort S for the ruling, M for the carve. Owner-gated (g-1).
15. **Two ratchet clicks** — `store` pin 1900 to 650; `server/platform` to the named pin the runway note ruled. TODO row "Depth ratchet: first click down + the platform pin". Impact High, effort S.
16. **Close the readiness-only DB rule** — `store.CountSessions` behind `session.Store`, `DB` out of `platform.Deps`, the arch pin. TODO row "Readiness-only DB rule". Impact Medium, effort S, `auth:` convention.
17. **Client-side depth pins** — four measurement-only pins in the `templBudgetPins` shape; the calls.js valve (1,715 lines) is the named split if the pin trips. TODO row "Client-side depth pins". Impact Medium, effort S.
18. **One home for the journal boot recipe** — six `system.DeploymentConfig` literals, one aggregate list, two app-owned one-shot builders. TODO row "One home for the journal boot recipe". Impact Medium, effort S to M; cheapest right after the T08 dependency train.
19. **Name the journal kernel** — eight shared helpers and `NewCoreSystem` into one file. TODO row "Name the journal kernel". Impact Low, effort S.
20. **Reword the standing test-drain row** — api, hooks and platform carried tests; auth did not. TODO row "Test-drain row precedent is half right". Impact Low, effort S.
21. **T08 tail (m08.6 to m08.8)** — vendorHash re-pin in one command, buildflow, the `auth:` commit, CI watch; carried from the 07-23 report, not touched here. TODO row "Dependency train".
22. **Union coverage as a build step** — tonight is a baseline (76.2% of all statements, 82.6% hand-written); a profile carries one block per test binary, merge on the maximum count. TODO row "`buildflow -coverpkg` union coverage step".
23. **File-size gate owner ratify** — `max_file_size` 1500 stands reversible; 20 hand-written files are over 350 lines tonight. TODO row "BuildFlow file-size-check disposition".
24. **Storage-ladder §5 DROP** — the store pin can click down again after it. TODO row "Storage-ladder §5 DROP migration".
25. **Tests ride the F6 carve** — the contracts move takes 173 test lines out of the server root; a checklist line on item 14's carve.
26. **domain exports-axis annotation** at 170 (162 tonight) — the ruling lives in the runway note, section `internal/domain`, cited by the struck runway row in TODO_LIST.
27. **config's two ceilings** — package pin 1600 with 67 lines left, file ceiling 1500 with 73; the runway note, section `internal/config`, rules the validate-lane subpackage for when the pin bites.
28. **DOMAIN_LANGUAGE ratify-or-delete** — also settles F12's "hub" naming. ROADMAP row.
29. **Refresh the architecture D2 pair again** — the `2026-11-19_07-41` pair predates two edges. TODO row "Refresh the architecture D2 pair again".
30. **Persist the review tooling in-repo** — the AST coupling tool, the coverage merge and the recount harness die with `/tmp`. TODO row "Persist the review tooling in-repo".
31. **Generate `docs/architecture-understanding/INDEX.md`** — TODO row "Generate `docs/architecture-understanding/INDEX.md`".
32. **`scripts/docs-gates.sh` newest-report rule** — resolve the report it vouches for through `report_glob`. TODO row "`scripts/docs-gates.sh` vouches for the report `report_glob` names".
33. **CI leg for `scripts/status-index.py --check`** — the 07-23 report's item 45, never rowed. TODO row "CI leg for `scripts/status-index.py --check`".
34. **SKILLS: editorial template at phone width** — the stat cards need 390px; `.where` does not wrap long paths. TODO row "SKILLS: editorial template at phone width" (repo: SKILLS).
35. **SKILLS: architecture-review method notes** — recount before the first splice, the profile-merge note, a kept home for render slices. TODO row "SKILLS: architecture-review method notes" (repo: SKILLS).
36. **SKILLS: status-report Output section** — four HTML steps under a `.md` default. TODO row "SKILLS: status-report Output section" (repo: SKILLS).
37. **Architecture review #8** — after the F6 ruling lands or at the next release. ROADMAP standing watch "Architecture review #8".

Standing rules and housekeeping (no row; each names where it lives):

38. **Daemon push lag** — TODO_LIST § "Daemon push-stall forensics" and the ROADMAP "Push-lag policy" question; the review chain's lag was 35m25s, the report batch crossed the bar at 00:31:05 (b-1).
39. **Re-measure `audit_max` and `generic_df`** at the next sweep — the rule lives in AGENTS.md (docs-gates bullet).
40. **`/tmp/archreview` is kept** until item 30 lands: it is the only copy of the review tooling, and `/tmp` dies at reboot.

40 rows: 13 landed, 24 open, 3 standing. The name of this file is not a task; see b-4 and the Handoff.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **The F6 ruling.** The server pin is full at 5300 of 5300. The
   machine-contract family (742 lines, 0 references into the hub) can leave
   as `internal/server/contracts`, which makes it a fifth surface carve and
   amends the 2026-11-12 "four carves plus the budget" end state. Carve, or
   ride the pin again? I measured the seam; whether the end-state ruling
   moves is yours.
2. **The push.** At 01:04, 2 commit(s) sit unpushed, the oldest `53686f32` (23:31, 93 minutes old), the newest `77e29b38`; `origin/main` reads `e23d35ac` and `scripts/push-stall-watch.sh --no-fetch` reads STALL. The manual-push bar is open (anchor: the
   first unpushed commit, `53686f32`, 2026-11-24 23:31:05). Order the
   push and the CI watch, or leave it to the daemon?
3. **The glossary.** `docs/DOMAIN_LANGUAGE.md` has carried its
   ratify-or-delete memo unanswered since 2026-10-14. Ratify it (and keep
   "hub" for the SSE broadcaster, F12), or delete it and let the code stay
   the only vocabulary?

## Handoff

Stopped on instruction after this report. Nothing was pushed by hand.

- This file stays at `docs/status/archreview-status.md` (your path), which
  sits outside the `2*.md` report glob: no index row, no freshness check.
  One `git mv` to a timestamped name puts it back inside both; it is
  rename-ready (b-4).
- `/tmp/archreview` is kept: the only copy of the review tooling (row 30).
- The review report's amendments are in its footer; the corrections of this
  report landed in `77e29b38` and the commits after it.
