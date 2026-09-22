# Status report: composer UX train, mid-flight freeze (2026-09-22 16:48)

Scope: THIS session's second train (composer UX batch) as frozen by
the owner's report request. The train is ~80% implemented and gated on
its fast paths; the remaining 20% is enumerated below so resuming is a
checklist, not archaeology. First-train state (send-failure UX,
`56caf37`, pushed+verified) is unchanged. A concurrent session (CRM
names → contacts/crm, backup retention) collided twice; both
collisions are documented.

## Self-review (the three questions)

**What did you forget?**

1. shell.js is inside the island-lint scope (`oxlint … shell.js`,
   flake.nix) — the AGENTS rule "new browser globals go in the config's
   globals block" existed and I applied it only at the LAST recon step:
   `DataTransfer` is referenced by my new chip code and is NOT in
   `island/oxlint.json` globals. `nix flake check` would fail closed.
   Caught, not yet fixed.
2. Same verification-class gap I flagged in the 16:29 report: the
   Enter-to-send path (requestSubmit → htmx submit binding) is verified
   through stubs and reasoning only, not at a browser or the served
   htmx.min.js — I criticized exactly this yesterday-evening for
   hx-disabled-elt and repeated the pattern.
3. `field-sizing: content` degradation claimed, not proven in the E2E
   chromium.

**What could you have done better?**

- My first time-format test fixture used `time.UTC` while the helpers
  render `.Local()` — failed on this host's TZ; fixed by riding
  `time.Local` (the failure was mine, caught by my own test — the pin
  worked as intended).
- Drafted a server test against a client helper (`doRaw`) that doesn't
  exist; caught at edit time and replaced with the established
  jar-merge idiom (should have read the harness for the pattern first —
  I DID have it in context from line 107 and missed it).
- Two genuine stub bugs surfaced via failing specs (helpers.mjs `append`
  not variadic; className vs ".class"-selector mismatch in closest) —
  fixed in the STUB toward real-DOM semantics, not by weakening the
  tests. That was the right call; the repro-then-fix loop cost ~4 min.

**What could you still improve?** → (e)/(f). Nothing shipped broken;
the risk is entirely in the not-yet-run gates.

**Did you lie?** No. 49/49 island node tests, views+server go suites
green — all re-runnable. Everything else is honestly labeled unrun.

## a) FULLY DONE

| Item                                                                                                                                     | Evidence                                               |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| Plan doc (pareto 20/4/1, coarse+fine tables, mermaid, constraints)                                                                       | `docs/planning/2026-09-22_16-36_SUPERB-composer-ux.md` |
| T3 time helpers `formatClock`/`formatStamp` + unit tests (en byte-stable, de 24h)                                                        | helpers.go/helpers_test.go green                       |
| T3 call sites: Bubble, relativeTime(+lang), FaxRow, vmWhen(+lang)                                                                        | messages/fax/voicemail .templ, regenerated             |
| T1 textarea composers (both message forms, `wp-compose-body`) + segcount spans + attach containers (reply + fax)                         | `TestComposerCarriesSegmentCounterAndTextarea` green   |
| T1/T2/T4 shell.js §4: GSM-7/UCS-2 segmentation, Enter/Shift+Enter (+IME guard), counter DOM update, chips render/remove via DataTransfer | `composer.test.mjs` — 5 specs                          |
| helpers.mjs upgrades: variadic append, selector matcher (class selectors), DataTransfer stub                                             | 49/49 node tests green                                 |
| CSS: field-sizing textarea, segcount, chips (+`[hidden]` guard vs author display)                                                        | app.css                                                |
| Server lang test: en meridiem vs de 24h bubble clock (wp-lang cookie)                                                                    | `TestBubbleClockFollowsLanguage` green                 |
| Concurrent-collision handling: multiedit conflict re-read+reapplied; their store generics breakage waited out, not touched               | git history                                            |

## b) PARTIALLY DONE (the frozen 20%)

| Item                                | What's missing                                                                                                                |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Island-lint gate                    | `DataTransfer` must be added to `internal/web/assets/island/oxlint.json` globals; then run the oxlint command locally         |
| Full gates                          | `go test -count=1 ./...` (only views+server run), `nix fmt`, buildflow (check concurrent train's erraudit state first), smoke |
| Paper trail                         | CHANGELOG Unreleased entries, plan-doc verdict section                                                                        |
| Narrative commit + push + ls-remote | Not yet (daemon chore-sweeps hold the code so far)                                                                            |
| Browser-truth verification          | requestSubmit↔htmx interplay + field-sizing posture (E2E or served-artifact level)                                            |

> Resolved 2026-09-22 (docs-health): island-lint ran green on resume
> (env.browser already covered DataTransfer — no config change needed);
> full gates + CHANGELOG + narrative commit landed (5cce98d); per-thread
> drafts shipped (T21d). The browser-truth E2E stays with the release
> TODO row; train C/E owner calls stay in the send-failure TODO row.

## c) NOT STARTED

| Item                                                                                                                                   | Note                           |
| -------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Next-train ideas from the plan: dial typeahead (PBX_CONFIG contacts), jump-to-latest chip, per-thread drafts                           | planned, coarse table rows 6-8 |
| Carried from train 1: C (pre-flight self-send, owner call open), D (bubble failure story), E (422 semantics), F (live own-DID warning) | TODO_LIST row                  |
| Stack browser E2E re-run — now covers TWO trains of markup changes                                                                     | declared gate, still skipped   |

## d) TOTALLY FUCKED UP

Nothing shipped broken (nothing final shipped yet at all). Stumbles:

1. The `DataTransfer` globals omission — found by my own last recon
   step, unfixed at freeze. If the daemon's flake check runs before I
   resume, it fails on MY code.
2. The UTC/Local test fixture error (mine, self-caught).
3. The phantom `doRaw` harness call (mine, caught at edit time).
4. Repeated the verify-through-stubs-only pattern after criticizing it
   in the morning report — the honest headline of this freeze.

## e) WHAT WE SHOULD IMPROVE

1. A "new browser global → oxlint.json" reflex step in the plan's
   fine-table for ANY shell.js/island work (the rule is in AGENTS; the
   process step wasn't in my table).
2. Behavior added via JS needs one browser-truth gate per train (E2E
   or served-artifact grep) — stubs prove logic, not integration.
3. Read the test harness for an existing idiom BEFORE drafting new
   test code against imagined helpers.

## f) Next tasks (resume order, impact-sorted)

| #  | Task                                                                                                     | Impact          | Effort |
| -- | -------------------------------------------------------------------------------------------------------- | --------------- | ------ |
| 1  | Add DataTransfer to island/oxlint.json globals; run island-lint command locally                          | High (gate red) | XS     |
| 2  | `nix fmt` (prettier owns shell.js/app.css/island-tests)                                                  | High            | XS     |
| 3  | Full `go test -count=1 ./...`                                                                            | High            | S      |
| 4  | Buildflow (attribute any concurrent-train findings first) + smoke 38/0                                   | High            | S      |
| 5  | CHANGELOG Unreleased entries + plan verdict                                                              | Med             | S      |
| 6  | Narrative commit + push + `git ls-remote` verify                                                         | High            | XS     |
| 7  | Stack browser E2E re-run (covers train 1 + 2 markup; ~6-7 min)                                           | High            | S      |
| 8  | Verify requestSubmit↔htmx at served htmx.min.js / E2E; field-sizing in E2E chromium                      | Med             | S      |
| 9  | Dial typeahead train (PBX_CONFIG contacts, zero round-trips)                                             | High            | M      |
| 10 | Jump-to-latest chip on live pushes while scrolled up                                                     | Med             | S      |
| 11 | Per-thread draft persistence (localStorage)                                                              | Med             | S      |
| 12 | Train C decision + implementation (owner call, see g/1)                                                  | High            | S      |
| 13 | Train D bubble failure story (persist reason+kind, retry-where-retryable)                                | High            | M-L    |
| 14 | Train E: provider refusal → 422 + family vocabulary + runbook sync                                       | Med             | S-M    |
| 15 | Fax-lane self-send guard (with C)                                                                        | Med             | S      |
| 16 | Contacts add/import in-flight guard decision (idempotent upsert — decide deliberately)                   | Low             | XS     |
| 17 | Native de review of new copy (none added this train — `thread.selfNotice` from train 1 still unreviewed) | Low             | XS     |
| 18 | Screenshot QA both themes (notice + chips + counter)                                                     | Low             | XS     |

## g) Questions I cannot figure out myself

1. **Train C (still open from 16:29):** pre-flight self-send = instant
   422 with NO thread row, or today's evidence-preserving FAILED row?
2. **Resume:** finish tasks 1-6 (gates → docs → commit → push) now, or
   change anything about the frozen implementation first?
3. **E2E:** one stack browser-E2E run now to cover both trains' markup
   changes, or fold into the next stack bump? (asked at 16:29, still
   unanswered)

---

_`.md` per explicit owner instruction (skill default is HTML; override
flagged). Not manually committed — the auto-commit daemon owns sweeps
under the harness no-commit-without-request contract. The tree at
freeze mixes my files with the concurrent session's (CHANGELOG.md,
README.md, crm_test.go, store/messages.go are THEIRS — do not sweep
into a narrative commit blindly; `git add` by explicit path on resume._
