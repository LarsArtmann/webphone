# Status: UX Raw-Ideas Train — Features E+F landed, gates green, docs owed

_2026-09-22 21:43 CEST · webphone · continuing `docs/status/2026-09-22_21-08_ux-raw-ideas-train-midflight.md`_

## TL;DR

All 6 implementable ideas from the raw-ideas list now have code landed and
green scoped gates. This session shipped **thread search (E)** and the
**audio output picker (F)** end to end, plus the full BuildFlow gate
(52 steps, 0 failed — nix flake check, gitleaks, Go suite, island suite all
green). Owed before fold: TODO_LIST/CHANGELOG/FEATURES rows, the plan-doc
verdict, one narrative commit, smoke, and the deferred stack browser E2E.
Everything so far was committed by the auto-daemon as heuristic
`chore:` commits — no narrative commit exists yet.

## Honest self-review first (what you asked)

**What I forgot:**

- To verify my own implementation detail (guard reads `getElementById`)
  before writing the test for it — the spec stubbed `querySelector` and
  failed red.
- That `audioout.js`'s module-level `initialized` guard makes one init per
  process — my first test file called init repeatedly and could never pass.
  Two rewrites before the split into full-env + unsupported-env files.
- That the full `/messages` page always contains `+441632960961` (composer
  placeholder), so a raw substring assertion is wrong — wrote the assertion,
  ran it, watched it fail on exactly that.
- To run `nix fmt` after the final `audioout.test.mjs` edits — BuildFlow had
  to tell me oxfmt was unhappy. Also: `nix fmt` does NOT fix island-tests
  (its treefmt scope differs from BuildFlow's oxfmt scope); the fix is
  `buildflow -s oxfmt --fix`, which is what finally cleared it.

**What I could have done better:**

- Land small explicit narrative commits as I go instead of letting the
  daemon sweep broken intermediate states (a red `TestSearchThreads` case
  sat in a public `chore:` commit for ~2 minutes).
- Write tests against the STUB semantics I already know (helpers.mjs
  auto-registers `getElementById` ids; makeEl listeners are inspectable) —
  I re-derived them by failure instead of by reading.
- Run fmt + scoped lint BEFORE handing work to the big gate.

**What I could still improve (product):**

- All session JS is stub-verified only — the browser E2E is still owed.
- SQLite `LIKE` is ASCII-case-insensitive only: a search for "MÜNCHEN" will
  not match "münchen". Unicode-insensitive search needs `lower()` +
  `lower_lithuanian`-style collation or an FTS5 column — real design work,
  not a patch.

---

## a) FULLY DONE (implemented + verified green this session)

1. **E.1 — store search**: `SearchThreads(ctx, owner, query)` in
   `internal/store/messages.go` (same row shape as `ListThreads`, matches
   remote OR any message body via `EXISTS`, `ESCAPE '\'`, `likeEscape`
   turns `%`/`_`/`\` literal, ordering by last activity). 7-case table test
   (`TestSearchThreads`): body-in-older-message, remote match, ASCII
   case-insensitivity, literal `%`, literal `_`, no-match, owner scoping,
   newest-first ordering. `go test ./internal/store/` green.
2. **E.2 — wiring**: `Service.ThreadSearch` proxy; `messagesPanel` reads
   `q` (trimmed; empty → plain `Threads`), passes `Query` into
   `ThreadsPanelProps`; search form in `messages.templ`
   (`role="search"`, `hx-get="/partials/messages"`, 300 ms debounce +
   submit, `hx-sync="this:replace"`, input id `wp-thread-search-input`
   echoing the query for morph-safe focus); query-aware no-match empty
   state with localized quotes („" / ""); 5 new i18n keys in BOTH en+de
   maps; `.wp-thread-search` CSS; templ regenerated; build green.
3. **E.2 — server test**: `TestThreadSearchFiltersPanel` pins partial+page
   filtering, query echo in the input, literal `%`, quoted no-match state,
   and the unfiltered empty-q baseline.
4. **E.3 — shell guard**: `shell.js` §3b-3 cancels `htmx:sseBeforeMessage`
   for `.wp-thread-list` while the search input holds text (live pushes
   must not stomp a filtered view); 4-assertion spec (query set → cancelled,
   whitespace → cancelled, empty → passes, other targets untouched).
5. **F — audio output picker**: `internal/web/assets/island/app/audioout.js`
   (feature-detects `setSinkId` + `enumerateDevices`; only UNHIDES for ≥2
   physical outputs — `default`/`communications` aliases excluded; persists
   pick in `localStorage["wp-sink"]`; applies via `setSinkId` to
   `#remote-audio` ONLY (ring tones stay room-alarms); `devicechange`
   repopulates keeping live selections, falling back off dead ones;
   re-labels on `wp:lang-changed`; every failure logs and stays hidden).
   `phone.templ` hidden `#audio-out-wrap` + `#audio-output`; i18n
   `audioOutput`/`audioDefault` in en+de; `main.js` boots it after the
   typeahead; `.audio-out` island CSS; 6 specs across
   `audioout.test.mjs` (full env, one-init-per-process) +
   `audioout-unsupported.test.mjs` (single-output env, own process).
6. **Verification state**: full `go test -count=1 ./...` green (14 pkgs);
   island suite 76/76; `nix flake check` — ALL CHECKS PASSED (includes
   treefmt, island-js/oxlint, webphone-module eval, KVM-gated backup VM
   test, vulnix-triage); BuildFlow `52 success, 0 failed`; gitleaks: no
   leaks across 543 commits.

## b) PARTIALLY DONE

1. **Gates**: Go + island + buildflow + flake check are green, but the
   **smoke run** (`python3 scripts/webphone-smoke.py`, 38-check + restart
   scenario) has NOT been run this session.
2. **Formatting**: the two oxfmt findings on my island-test files are fixed
   and re-verified (76/76), but the daemon hasn't necessarily swept/pushed
   the final state yet — `git ls-remote` check still owed.
3. **The train itself**: 6/6 implementable ideas have code + tests; the
   train's closing sweep (docs, narrative commit, E2E decision) is open.
4. **BuildFlow warnings triage**: pass is green, but 12 tools reported
   findings (see d/e) — most pre-existing, none mine except oxfmt (fixed).

## c) NOT STARTED (owed from the train plan)

1. **TODO_LIST.md rows**: peer-hub single view (contact → thread + history +
   voicemail, M-L) and the Tailwind v4 scoped-layer spike (owner call).
2. **CHANGELOG.md [Unreleased]** entries: hover stamps, jump chip, missed
   badge, dial typeahead, thread search, audio picker.
3. **FEATURES.md** inventory rows for the six shipped ideas.
4. **Plan-doc verdict** (`docs/planning/2026-09-22_21-05_SUPERB-ux-raw-ideas.md`
   still says PENDING) — must record the stub-evidence caveat honestly.
5. **One narrative commit** for the train + `git ls-remote` verification.
6. **Stack browser E2E** — owner question from last report, still
   unanswered (see g).

## d) TOTALLY FUCKED UP

Nothing catastrophic. Damage report, honestly:

1. A **red test state was pushed** briefly (`TestSearchThreads` empty-query
   case asserting behavior opposite to LIKE semantics) — daemon committed
   it mid-fix; fixed within minutes. Lesson: my drafting, the daemon's
   velocity.
2. **Three wasted red-green cycles** on tests written before validating
   assumptions (querySelector vs getElementById; module-init guard; page
   placeholder content). Each was cheap but avoidable.
3. **Pre-existing BuildFlow findings I noticed but did NOT touch** (not
   mine to fix in this train, flagged for triage):
   - `ruff-format` + `mypy` + `bandit` on `scripts/webphone-smoke.py`
     (another session's theme-preload additions, 213 bandit notes are
     mostly its self-signed-TLS test client);
   - `golangci-lint`: `internal/store/db.go` `updatedOrNotFound` unused,
     `export_test.go` unchecked `rc.Close`;
   - `lychee`: CHANGELOG links to `v2.6.0` tag/compare **404** — CHANGELOG
     references a tag that was never cut (or was renamed to 2.5.0?);
   - `nix-checker`: `vendorHash` stale vs `go.sum` (buildflow warns `nix
     build` may fail on hash mismatch — but `nix-build` PASSED this run,
     so likely a false-positive freshness heuristic worth checking);
   - `go-version-auto-configure`: `go.mod` pins `go 1.27.1` (patch floor);
   - `AGENTS.md` is 383 lines vs max 377 (6 over);
   - vulnix: binutils/bison/coreutils CVE advisories in the toolchain
     closure (triage CLI's job to gate).

## e) WHAT WE SHOULD IMPROVE

1. **Test-first discipline inversion**: when a design is fully decided
   (like E was), read the exact implementation surface (which DOM API, which
   stub method) before writing the spec, not after it fails.
2. **Commit rhythm**: make the narrative commit at each feature boundary so
   the daemon's heuristic commits only carry leftovers, never mid-states.
3. **Formatter scope confusion**: `nix fmt` ≠ BuildFlow oxfmt scope for
   island-tests. One home for "format everything" would end this class of
   gate surprise (either teach treefmt the island-tests scope or drop it
   from BuildFlow).
4. **Unicode search**: ASCII-only `LIKE` folding is a real product limit for
   de (umlauts). FTS5 or `lower()`-based matching deserves a small ADR.
5. **Search UX depth**: no clear-button, no result count, no URL persistence
   (owner question pending), no CRM-name matching (search matches raw
   numbers only, not the display name from CRM) — all small, all deferred
   deliberately.
6. **Stub-evidence gap**: every JS feature this train shipped is
   node:test-verified only; the browser E2E debt from the last three trains
   keeps growing and already bit the stack once (transfer-step flake).

## f) NEXT 50 (prioritized, roughly in order)

**Close this train (1–8)**

1. Run smoke: `python3 scripts/webphone-smoke.py` (add probes if the new
   surfaces warrant them — search `q=`, `#audio-out-wrap` presence).
2. Fill the plan-doc verdict honestly (shipped list + stub caveat).
3. TODO_LIST.md: peer-hub row + Tailwind v4 spike row (refined, with sub-steps).
4. CHANGELOG.md [Unreleased]: one entry per shipped idea (6).
5. FEATURES.md rows for the six ideas (DONE).
6. One narrative commit for the train; verify `git ls-remote origin main` == HEAD.
7. Answer the three owner questions (g) — they gate E2E + small refactors.
8. If approved: run the stack browser E2E (budget ~445 s; watch the known
   ~90 s transfer-step flake; cover typeahead, chip, badge, search, picker).

**Correctness / hardening (9–16)**
9. Unicode case-insensitive search design (FTS5 vs `lower()` ADR).
10. Search over CRM display names, not just raw numbers (needs a name-aware
store path or in-memory filter after `crmNames` enrichment).
11. Clear-button affordance in the search form (JS-free `type=reset` doesn't
re-render the list — needs a small shell handler or an `hx-get` link).
12. Decide + implement search URL semantics per owner answer (pushState ?q=
vs ephemeral).
13. setSinkId browser-truth check on Safari/Firefox (feature-detect hides
the picker — confirm no console noise; document support matrix).
14. Missed-call semantics per owner answer (REJECT counting; badge clears on
island Recent panel too?).
15. `htmx:sseBeforeMessage` search guard: also skip the _unread nav refresh_
storm while filtering? (Currently only the list push is cancelled; the
nav badge may flap while a filtered view hides the unread row.)
16. Add a contract test that the search input id stays OUTSIDE the
sse-swap region (markup regression guard for the morph-focus design).

**Pre-existing findings worth their own micro-train (17–24)**
17. Fix `scripts/webphone-smoke.py` ruff-format + mypy findings (another
session's file — coordinate first).
18. Remove or wire `internal/store/db.go` `updatedOrNotFound` (golangci:
unused).
19. Check `rc.Close` in `internal/server/export_test.go`.
20. Fix CHANGELOG v2.6.0 links (lychee 404) — tag was likely meant to be
v2.5.0-era; verify release history.
21. Investigate vendorHash staleness warning (flake built fine this run —
either fix the hash or the heuristic).
22. `go.mod`: consider `go 1.27` (drop patch floor) per go-version finding.
23. AGENTS.md trim to ≤377 lines (move the CRM seam + erraudit tier details
into docs/).
24. Rebuild the BuildFlow binary (predates HEAD by 59 h — preflight warn).

**Feature follow-ons from the same ideas list (25–34)**
25. Peer-hub single view (contact-centric thread + history + voicemail).
26. Tailwind v4 scoped-layer spike (owner-gated; one-component proof).
27. Search result count + active-filter chip in the panel header.
28. Search deep-link: open a thread then return with the query preserved.
29. Typeahead: also search thread remotes (currently contacts only).
30. Typeahead: recent-calls recency boost in `rankContacts`.
31. Audio picker: remember per-device label changes; handle revocation
(permission) path explicitly.
32. Jump-chip: coalesce rapid pushes into one count update (currently
increments per push — fine, but text thrashes).
33. Hover stamps: also title the nav-badge and voicemail rows.
34. Search: debounce indicator (subtle spinner) for slow phones.

**Platform / hygiene (35–42)**
35. Re-measure erraudit tier-2 family adoption (due 2026-10-22; current
baseline 127 stdlib_constructor / 113 outside crm).
36. aarch64 cross-build verify (`nix build .#webphone --system
    aarch64-linux`, ELF-bytes check) before the next stack re-pin.
37. Consider `--all-systems` for `nix flake check` in CI (warning noted).
38. Stack re-lock ritual (fold → stack bump → pbx-artmann relock) per
release runbook once the narrative commit lands.
39. Record the "island-tests module-init guard → one-init-per-process +
separate unsupported-env file" pattern in AGENTS.md test section (it's
now used twice).
40. Record "nix fmt vs buildflow oxfmt scope" gotcha in docs/lessons.md.
41. Consider smoke probes for the six new surfaces (grep assertions on the
served HTML).
42. Prune BuildFlow cache DBs (1.34 GB cache, 0.16 GB state — preflight).

**Nice-to-have polish (43–50)**
43. Localize the jump-chip ("↓ N new") + missed badge if the owner ever
overturns the language-neutral-shell decision (D3).
44. Search input `enterkeyhint="search"` + `type=search` clear-native on
mobile Safari.
45. Audio picker: add `title` tooltips with the full device label (CSS
truncation).
46. Empty-state illustration/tone pass (the quoted no-match line is plain).
47. Consider `aria-live` announcement for search result counts (a11y).
48. Typeahead listbox: `aria-activedescendant` wiring check.
49. Consider persisting the audio pick per-output-device availability
windows (currently global; churn falls back sanely already).
50. Doc: one paragraph in README's tab docs about the Messages search +
picker (user-facing surface).

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Search URL semantics** (carried over, still gates task 12): should a
   typed query push `?q=` into browser history (deep-linkable, back-button
   unwinds filters) or stay ephemeral as now? I implemented ephemeral; if
   you want history, the morph-swap needs `hx-push-url` handling and a
   back/forward test.
2. **Missed-call semantics** (carried over): does a deliberately REJECTED
   incoming call count as missed (current: no — a reject hides the banner
   first and never dispatches), and should the badge also clear when you
   open the island's _Recent calls_ panel (current: History tab click only)?
3. **Browser E2E now or at fold?** (carried over): running the stack browser
   E2E now verifies this train's five JS surfaces in a real browser
   (~445 s budget, plus stack boot; the transfer-step flake may cost one
   re-run). Deferring keeps the debt for the fourth consecutive train.
   Run it now, or record it as owed-at-fold?

---

### Verification ledger (what "green" means right now)

| Gate                                    | Result                                              |
| --------------------------------------- | --------------------------------------------------- |
| `nix develop -c go test -count=1 ./...` | 14 pkgs ok, 0 fail                                  |
| Island suite (`node --test`)            | 76/76 pass                                          |
| `nix flake check` (buildflow-driven)    | all checks passed (75 s)                            |
| BuildFlow (`scripts/buildflow.sh`)      | 52 success / 0 failed, 12 detect-only warning tools |
| gitleaks                                | no leaks (543 commits)                              |
| oxfmt on this train's files             | fixed + re-verified                                 |
| Smoke (38-check)                        | **not yet run this session**                        |
| Stack browser E2E                       | **owed (owner decision pending)**                   |
| Narrative commit                        | **owed** (daemon heuristic commits only)            |
