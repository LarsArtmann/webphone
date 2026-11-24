# Architecture review #7 — session status

> Claims verified as of **23:27 CET 2026-11-24** (push/CI statements rot
> fastest). Session: one order, "review this project and create a .html
> report from a Software Architect perspective", executed with the
> architecture-review skill as series #7. HEAD at report time: `e23d35ac`.

> **Scope.** Verified here, each with the command that produced it: the
> package graph (`go list -json ./internal/... ./cmd/...`: 38 packages, 149
> internal edges, 0 cycles); the full test suite
> (`go test -count=1 -coverpkg=./internal/...,./cmd/... ./internal/... ./cmd/...`,
> rc 0, 37 packages ok, 0 failed); the arch suite
> (`go test -count=1 -v ./internal/arch/`, 15 PASS); the docs battery
> (`nix develop -c bash scripts/docs-gates.sh --check`, rc 0, runs ending
> 22:53 and 22:57); the TODO pre-flight (`docs-gates freshness .
> --pre-flight`, SAFE at 22:52 and 22:58); the report's own integrity checks
> (144 of 144 recount checks, HTML parse, anchors, stylesheet identity); and
> the git state read at 23:24 (`git fetch`, then
> `git rev-list --count origin/main..HEAD` = 3). Every measurement was taken
> at `c592dd22`; the three commits since touch no Go file
> (`git diff --stat c592dd22..e23d35ac -- '*.go'` is empty), so they hold for
> `e23d35ac`.
>
> **NOT verified here:** the CI verdict (nothing was pushed, see b-1); the
> T08 dependency train (the 07-23 report's item; not opened); `nix flake
> check`, buildflow, smoke and workflow-e2e (a docs-only session, none run);
> any number from reports #1 to #6 beyond the ones re-measured; a
> mobile-width render of the report (desktop slices only); the 11-19 D2
> diagrams (not regenerated).

## a) FULLY DONE (landed in local main; the push is pending, see b-1)

1. **Architecture review #7 report** —
   `docs/architecture-understanding/2026-11-24_22-51_architecture-review.html`,
   2,613 lines, 95,562 bytes, spliced into the editorial template with
   `splice-report.sh` (`--self-test` OK; the stylesheet is byte-identical to
   the template, only `<title>` differs). Seven dimensions held at 4.36;
   seven new findings F6 to F12; a seven-item roadmap with a table of what
   would move the two lowest scores. The bytes at `e23d35ac` are identical
   to the final splice (`cmp`, 23:15).
2. **Evidence the series did not have before** — the hub's mass by file
   family (the 742-line machine-contract family is a leaf: 0 references out,
   1 in); AST in-package coupling for seven packages (10 of 11 cqrs
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
4. **TODO harvest** — seven rows under "2026-11-24 architecture review #7
   harvest" at the top of `TODO_LIST.md` (`a27227ca`; one row reworded in
   `bc230306`); both pre-flights answered SAFE.
5. **Report integrity checklist** — recount 144 of 144 against the measured
   data; balanced tags, 26 unique ids, 8 anchors resolving, no inline style
   or handlers; every cited symbol grep-verified and every proposed name
   labelled; twelve desktop render slices inspected.
6. **Docs battery on the changed tree** — "PASS — formatted + all doc gates
   green", twice (22:53 and 22:57).
7. **Full test suite green** as a byproduct of the coverage pass: 37
   packages ok, 0 failed.
8. **Nothing stray in the repo** — the pre-flight ledger
   `.docs-gates-preflight.json` was trashed twice; all scratch (43 MB)
   stayed under `/tmp/archreview`.

## b) PARTIALLY DONE (in flight)

1. **Push and CI verdict** — three commits sit unpushed: `a27227ca`
   (22:53:18), `bc230306` (22:57:03), `e23d35ac` (23:07:24); `origin/main`
   read `c592dd22` at 23:24:40. The daemon committed three times and pushed
   zero times in that window (watcher heartbeats 22:58 to 23:13, a fetch at
   23:24). The manual-push bar opens at 23:53:18, sixty minutes after the
   first unpushed commit; it was neither reached nor invoked. Remaining: the
   push (daemon or bar), then `scripts/ci-verdict.sh e23d35ac` to green.
   Effort S. Blocker: none technical; the instruction to wait.
2. **Narrative commit** — none was written (the harness forbids commits
   without an explicit request), so the work rides three heuristic
   "chore: auto-commit" commits. The report carries a dated amendment note
   in its footer because five amendments followed the first splice.
3. **F12 ("hub" carries three meanings)** — in the report, bound to the
   ROADMAP DOMAIN_LANGUAGE row by reference; the ROADMAP text was not
   edited.
4. **This status report** — written at 23:27. The INDEX regen and the
   battery check for it run in the same step; their verdict is recorded in
   the handoff line at the end of this file.
5. **Series D2 pair** — not refreshed (not requested); the 11-19 diagrams
   predate the two new edges.

## c) NOT STARTED (the review's code work; the review changed no code)

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
   reported `ids.go` at 24.3% (it is 100%). The recount caught it 30 minutes
   later, after the first splice was already committed. Root cause: a
   `-coverpkg` profile carries one block per test binary. Mitigation: the
   recount harness, 144 checks, now part of how the report was verified.
2. **Regex coupling gave three false references.** Counting identifiers
   with a regex read `events` and `lang` out of string literals and
   comments; the machine-contract family first measured "3 references into
   the hub". A go/parser and go/scanner tool re-measured it: 0 out, 1 in.
   Caught before it reached a report file.
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
   footer lists the amendments.
6. **A TODO row shipped with a wrong title.** "Test-drain row precedent is
   wrong" landed in `a27227ca`; the truth is "half right" (platform yes,
   auth no). Reworded in `bc230306`, re-flighted in replace mode.
7. **Five amendments after the daemon had committed.** Three commits carry
   one artifact, and the report carries an amendment note. Root cause: the
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
    re-derives them unless f-25 lands.
11. Noticed: **the daemon's push leg pushed nothing in 31 minutes** while
    three commits waited (22:53 to 23:24); the 07-23 report lists the same
    stall.
12. Noticed: **`TODO_LIST.md` front matter contradicts the strike canon.**
    It says done work is deleted, never struck; the file holds 172 struck
    rows and AGENTS.md (owner ruling 2026-11-11) rules the opposite.
13. Noticed: **two stale numbers in AGENTS.md.** "Host go is below the
    1.27.1 floor" (host reads go1.27.3, equal to go.mod) and "`audit_max`
    72" (`.docs-gates.json` says 74).
14. Noticed: **`.docs-gates-preflight.json` is not git-ignored** (the tool
    says so itself); a session that forgets the trash step hands it to the
    daemon.
15. Noticed: **the status-report skill contradicts itself** — its Output
    section lists four HTML steps under a `.md` default.

## e) WHAT WE SHOULD IMPROVE

1. **Stage a report outside the repo and drop it in once.** Body, splice,
   recount, read-through and render all under `/tmp`; one `cp` at the end.
   Tonight that would have been one commit and no amendment note.
2. **Recount before the first splice, not after.** The recount exists; its
   place in the sequence was wrong.
3. **Persist the review tooling in the repo.** Three small programs would
   make series #8 a re-run instead of a re-derivation.
4. **AST or scanner for any coupling claim.** Regex over source counts
   strings and comments; it was wrong twice tonight and right never.
5. **A generated index for `docs/architecture-understanding/`.** It is the
   only report directory with a hand-maintained index.
6. **The architecture itself, in priority order (all harvested):** decide
   the hub pin's third raise with the contracts seam on the table (F6); put
   a ceiling on 15.9k ungoverned client lines (F7); close the readiness-only
   DB rule and pin it (F8); give the journal one boot recipe and one
   aggregate list (F9); correct the test-drain precedent (F10); name the
   cqrs kernel (F11); settle "hub" with the glossary call (F12).
7. **Daemon push-leg liveness.** Two reports in one day describe the same
   stall; a tracker issue is due if the bar fires again tonight.
8. **Align `TODO_LIST.md` front matter with the strike canon** so a new
   session does not "clean up" 172 struck rows.
9. **A mobile-width render in the report checklist.** The kit asks for one
   render; 1024px is where the sidebar hides and nothing was checked there.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Landed this session (struck here so the ledger stays honest):

1. ~~Write architecture review #7 as an HTML report~~ → done 2026-11-24 — `docs/architecture-understanding/2026-11-24_22-51_architecture-review.html`, final bytes at `e23d35ac`.
2. ~~Verify #6's roadmap and re-measure its claims~~ → done 2026-11-24 — report sections 01 and 03; 38 packages, 149 edges, 0 cycles; 4.36 held.
3. ~~Add the series INDEX row~~ → done 2026-11-24 — `docs/architecture-understanding/INDEX.md` (`bc230306`).
4. ~~Harvest the review roadmap into TODO_LIST~~ → done 2026-11-24 — seven rows, both pre-flights SAFE (`a27227ca`, `bc230306`).
5. ~~Run the docs battery on the changed tree~~ → done 2026-11-24 — `scripts/docs-gates.sh --check` PASS at 22:53 and 22:57.
6. ~~Run the report-integrity checklist~~ → done 2026-11-24 — recount 144 of 144, parse, anchors, twelve render slices.

Open, in execution order (each binds its TODO_LIST row):

7. **Owner ruling on the server pin's third raise** — carve `internal/server/contracts` (apidocs, wirecontracts, wirefixtures; 742 LOC) as a fifth surface carve, or ride with the runway note naming the seam. TODO row "Server pin full: decide the third raise". Impact Critical, effort S for the ruling, M for the carve. Owner-gated (g-1).
8. **Two ratchet clicks** — `store` pin 1900 to 650; `server/platform` to the named pin the runway note ruled. TODO row "Depth ratchet: first click down + the platform pin". Impact High, effort S.
9. **Close the readiness-only DB rule** — `store.CountSessions` behind `session.Store`, `DB` out of `platform.Deps`, the arch pin. TODO row "Readiness-only DB rule". Impact Medium, effort S, `auth:` convention.
10. **Client-side depth pins** — four measurement-only pins in the `templBudgetPins` shape. TODO row "Client-side depth pins". Impact Medium, effort S.
11. **One home for the journal boot recipe** — six `system.DeploymentConfig` literals, one aggregate list, two app-owned one-shot builders. TODO row "One home for the journal boot recipe". Impact Medium, effort S to M; cheapest right after the T08 dependency train.
12. **Name the journal kernel** — eight shared helpers and `NewCoreSystem` into one file. TODO row "Name the journal kernel". Impact Low, effort S.
13. **Reword the standing test-drain row** — api, hooks and platform carried tests; auth did not. TODO row "Test-drain row precedent is half right". Impact Low, effort S.
14. **Push and CI-watch the three review commits** — b-1; `scripts/ci-verdict.sh e23d35ac` must read green before anything rides on them. Impact High, effort S.
15. **T08 tail (m08.6 to m08.8)** — vendorHash re-pin in one command, buildflow, the `auth:` commit, CI watch; carried from the 07-23 report, not touched here. TODO row "Dependency train".
16. **Union coverage as a build step** — tonight is a baseline (76.2% of all statements, 82.6% hand-written); a profile carries one block per test binary, merge on the maximum count. TODO row "`buildflow -coverpkg` union coverage step".
17. **File-size gate owner ratify** — `max_file_size` 1500 stands reversible; 20 hand-written files are over 350 lines tonight. TODO row "BuildFlow file-size-check disposition".
18. **Storage-ladder §5 DROP** — the store pin can click down again after it. TODO row "Storage-ladder §5 DROP migration".
19. **calls.js valve** — 1,715 lines; the deepening brief's candidate (dock and transcript surfaces out of the call module) is the named split if the F7 pin trips.
20. **Refresh the D2 pair** — the 11-19 diagrams predate two edges. TODO row "Refresh the architecture-understanding diagrams".
21. **DOMAIN_LANGUAGE ratify-or-delete** — also settles F12's "hub" naming. ROADMAP row.
22. **domain exports-axis annotation** at 170 (162 tonight), the runway note's own schedule.
23. **config's two ceilings** — package pin 1600 with 67 lines left, file ceiling 1500 with 73; plan the validate-lane subpackage before either bites.
24. **Tests ride the F6 carve** — the contracts move takes 173 test lines out of the server root; make it a checklist line on that carve.

Standing maintenance and process (from tonight's observations; NEW rows
were harvested with this report):

25. NEW **Persist the review tooling in-repo** — the in-package coupling tool, the coverage merge and the recount harness lived under `/tmp/archreview` and die with `/tmp`.
26. NEW **Generate `docs/architecture-understanding/INDEX.md`** — hand-maintained today; `scripts/status-index.py` covers only `docs/status`.
27. NEW **Git-ignore `.docs-gates-preflight.json`** — the tool asks for it in its own note; tonight it was trashed by hand twice.
28. NEW **Align the TODO_LIST front matter with the strike canon** — "deleted, never struck" against 172 struck rows and the 2026-11-11 owner ruling.
29. NEW **Two stale numbers in AGENTS.md** — the host-go floor sentence and the `audit_max` pin (72 written, 74 in the file).
30. **Tracker issue for the daemon push leg** if the bar fires again tonight (zero pushes in 31 minutes; the 07-23 report lists the same stall).
31. **Architecture review #8 trigger** — after the F6 ruling lands or at the next release; forward rule: scalability drops to 3.5 if the server pin rides a third time with no carve.
32. **SKILLS repo: architecture-review skill** — add "recount before the first splice" and the coverage merge note to its method notes.
33. **SKILLS repo: status-report skill** — the Output section still lists four HTML steps under a `.md` default.
34. **Re-measure `audit_max` and `generic_df`** at the next sweep; this report adds a dense §f.
35. **Trash `/tmp/archreview`** (43 MB) once the push and the CI verdict are in.
36. **Mobile-width render** of the review report; only desktop slices were inspected.
37. **A durable home for render slices**, or drop the "screenshot for the record" step from the kit's checklist.
38. **CI leg for `scripts/status-index.py --check`** — the 07-23 report's item 45; tonight's INDEX regen is the same class.

38 rows: 6 landed, 18 open against TODO_LIST or ROADMAP rows, 14 standing.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **The F6 ruling.** The server pin is full at 5300 of 5300. The
   machine-contract family (742 lines, 0 references into the hub) can leave
   as `internal/server/contracts`, which makes it a fifth surface carve and
   amends the 2026-11-12 "four carves plus the budget" end state. Carve, or
   ride the pin again? I measured the seam; whether the end-state ruling
   moves is yours.
2. **The push.** Three review commits have waited since 22:53 and the
   daemon has not pushed. The manual-push bar opens at 23:53:18. Invoke it
   and watch CI, or leave the push to the daemon?
3. **The glossary.** `docs/DOMAIN_LANGUAGE.md` has carried its
   ratify-or-delete memo unanswered since 2026-10-14. Ratify it (and keep
   "hub" for the SSE broadcaster, F12), or delete it and let the code stay
   the only vocabulary?

## Handoff

Stopped on instruction after this report. Waiting for instructions.
