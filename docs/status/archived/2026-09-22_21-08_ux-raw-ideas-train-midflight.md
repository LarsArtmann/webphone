# Status 2026-09-22 21:08 — UX raw-ideas train MIDFLIGHT (A–D shipped, E in flight, F+docs owed)

> CLOSED 2026-09-23 (docs-health): superseded by the 21:43 close-out
> — E and F landed with green gates, the plan verdict is filled, and
> the whole train (A–F) rode v2.6.0 (tag `807ca0c`) with docs complete
> (CHANGELOG/FEATURES rows for all six ideas; the four A–D bullets
> completed by the 2026-09-23 sweep). Still open, routed: peer hub +
> Tailwind spike (ROADMAP), browser-truth E2E for the train's JS = the
> v2.6.0 release TAIL (TODO row).

Scope: THIS session only — the owner pasted the "Composer/UX raw ideas
(2026-09-22 trains, unshipped)" list with an execute order. Plan doc:
`docs/planning/archived/2026-09-22_21-05_SUPERB-ux-raw-ideas.md` (verdict still
PENDING — written before execution, not yet filled). End state at
freeze: four of six implementable ideas shipped with green scoped
gates, the fifth (thread search) fully designed but ONE LINE written,
the sixth (audio picker) untouched. No narrative commit yet — the
auto-commit daemon swept everything into chore commits as I went.

## a) FULLY DONE (this session, verified from actual runs)

| Item                                                                                                                                                                                                                                                                                                                                              | Evidence                                                                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A: Hover timestamps — `fullStamp` helper (de "22.09.2026 16:09", en "Sep 22, 2026, 4:09PM") as `title` on thread-row relative times and bubble clocks                                                                                                                                                                                             | helpers_test.go `TestFullStampCarriesDateYearAndTime`; server pin `TestTranscriptCarriesHoverStampsAndJumpChip`; `go test ./internal/...` green at that point |
| B: Jump-to-latest chip — transcript wrapped in `.wp-transcript-wrap`, hidden chip server-rendered, shell.js §3b-2 counts scrolled-away SSE pushes ("↓ N new"), pins near-bottom pushes, click returns to bottom, scroll-to-bottom resets                                                                                                          | shell.test.mjs ×2 specs green; templ regenerated; app.css wrap+chip styles                                                                                    |
| C: Missed-call badge — island dispatches `wp:call-missed` on the two genuinely-missed paths (caller gave up pre-answer in connection.js; accepted call died before media in calls.js); a REJECT never dispatches. shell.js §2d renders `#missed-badge` ("missed · N", English D3), cleared when History opens                                     | missed-call.test.mjs ×4 specs + shell.test.mjs badge spec, all green                                                                                          |
| Dead `callEnded(target)` duplicate removed from BOTH island i18n maps (shadowed by the `callEnded(dur)` entry — a real split-brain, zero-behavior fix)                                                                                                                                                                                            | `rg callEnded` = one key per map                                                                                                                              |
| D: Dial typeahead — new island module typeahead.js over `PBX_CONFIG.contacts` (zero round-trips): ranking name-prefix < name-contains < number-contains, tie-broken, capped 6; full keyboard support (↑/↓/Enter/Escape); idempotent init; JS-created listbox (served DOM untouched); main.js wiring + lang-change relabel; en/de `typeaheadLabel` | typeahead.test.mjs + typeahead-empty.test.mjs green; island suite 70/70; island/style.css dropdown styles                                                     |
| Stub upgrades toward real-DOM semantics: `parentElement` getter, `focus()` tracker, classList now mutates `className` observably                                                                                                                                                                                                                  | helpers.mjs; kept all 44 pre-existing tests green                                                                                                             |
| Research pass that de-risked E before code: store `listRows` shape, `messaging.Service` field names, notifier's `ThreadsList` SSE payload call site, LIKE-escape design, focus-preservation decision (morph swap + input id)                                                                                                                      | in-head + plan doc fine plan E.1–E.3                                                                                                                          |

Gates actually RUN this session: `nix develop -c go test -count=1 ./...`
green (mid-train, after A–C markup), scoped server+views green after the
pin test, island node suite 70/70, templ generate diff = messages_templ.go
only (verified before running).

## b) PARTIALLY DONE

| Item                                      | What's missing                                                                                                                                                                                                                                                                                                                                                                                                        |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~E: Thread search~~ done — landed at 21:43 with 7-case store test, panel wiring, shell guard, i18n; shipped in v2.6.0 | ~~DESIGN COMPLETE… ZERO lines landed → superseded by the close-out~~ |
| ~~Gates at train end~~ done — full suite + island 76/76 + flake check + buildflow 52/0 at 21:43 | ~~none run → ran at close-out~~ |
| ~~Plan doc verdict~~ done — filled (shipped 6/6 + stub caveat) | ~~PENDING placeholder → filled 2026-09-23~~ |
| ~~Narrative commit~~ done — the release-train boundary commits carry it (711fff5 fold, b162e22 bump, tag 807ca0c) | ~~daemon chore commits → superseded~~ |
| ~~CHANGELOG / FEATURES / TODO_LIST / AGENTS~~ done — 02:47 (E+F + AGENTS CRM) + the 2026-09-23 sweep (A–D bullets + FEATURES rows; ROADMAP cluster rewritten) | ~~untouched → synced~~ |

## c) NOT STARTED (none silently skipped)

- ~~F: Audio output picker (`setSinkId`): audioout.js module, phone.templ
  picker wrap, island i18n keys, audioout.test.mjs~~ done — landed at
  21:43 (apply to `#remote-audio` only; ring tones stay room-alarms);
  shipped in v2.6.0
- TODO_LIST rows for the two deferred ideas: peer hub (M-L view train)
  and the Tailwind v4 coexistence spike (owner-gated by design). ←
  routed (ROADMAP composer cluster)
- Browser-truth verification of ANY of this session's JS (same blind
  spot as the previous trains — everything is stub-level evidence). ←
  routed (v2.6.0 release TAIL row: stack E2E ×2 on the new chain)

## d) TOTALLY FUCKED UP (mine, no excuses)

1. **Repeated the import()-cache-bust mistake ONE SESSION after it was
   documented.** The 20:19 self-review recorded "wrote a test on the
   assumption that query strings defeat node's ESM module cache" as a
   lesson; I then designed the "without contacts" typeahead test on
   `import("../island/app/typeahead.js?empty")` and expected it to see
   fresh config state — the fresh module still imported the CACHED
   config.js, so the test asserted the wrong thing. Caught by the run,
   fixed by splitting into typeahead-empty.test.mjs (per-file process
   isolation). The lesson existed in writing; I did not re-read it.
   That is now strike three on this lesson class.
2. **The helpers.mjs rewrite dropped `style`/`dataset`/`attrs`** —
   ~14 tests across five files went red in one sweep. The classList
   upgrade itself was right (real-DOM semantics); the execution was a
   big-bang replacement instead of a surgical addition. Restored the
   props and everything came back green, but a file-wide rewrite of
   the SHARED test harness mid-train was the wrong shape of edit.
3. **Two drafting-garbage lines typed into test files** (a nonsense
   ternary assertion in typeahead.test.mjs; a `!Contains(...) == false`
   double-negative in messages_test.go). Both caught before they could
   run — but they were typed at all, which means draft-then-scan, not
   draft-clean.
4. **shell.js newline collapse**: an edit whose old_string ended with a
   trailing newline deleted the line break before the next
   `addEventListener`, producing `swapped in.    document.addEventListener` —
   LSP flagged it instantly; fixed in one edit. Small, but it was an
   avoidable whitespace-mismatch class the AGENTS explicitly warns
   about.
5. **Unicode ellipsis mismatch** in the i18n dead-key removal (ASCII
   `...` in my old_string vs `…` in the file) made a multiedit fail
   half-applied, and I briefly misread WHICH half applied (the sed
   check settled it). The en/de parity test would have caught a real
   error; this was process noise, not product risk.
6. **Two edit-tool refusals on main.js** ("modified since read" —
   daemon mtime churn). Recovered correctly by re-reading (content was
   unchanged), but I should batch main.js edits in ONE call from the
   start in this repo.

## e) WHAT WE SHOULD IMPROVE

1. **Re-read the prior session's lessons BEFORE designing tests** —
   lessons.md and the last self-review are the exact lookup for "how do
   I import a fresh module under node:test". Strike three means the
   trigger must move earlier: at test-design time, not at review time.
2. **Never rewrite the shared test harness wholesale mid-train.** Add
   props surgically; the blast radius of helpers.mjs is every island
   test.
3. **Scan typed test code for double-negatives and placeholder
   assertions before saving** — both garbage lines were visible in the
   write I sent.
4. **The near-bottom/chip logic rides stub evidence** (setTimeout-after-
   swap, capture-phase scroll listener). One browser run would retire
   the risk; keep it on the E2E bill for this train.
5. Daemon chore commits are fine as scaffolding, but the train still
   owes ONE narrative commit at the boundary — otherwise history reads
   as "8 heuristic commits" with no why.

## f) NEXT UP (prioritized, this train first)

1. E.1: `store.SearchThreads` + `likeEscape` + store test (body match in ANY thread message, remote match, wildcard-literal `%`, owner scoping).
2. E.2: `messaging.ThreadSearch`; `messagesPanel` reads `?q=`; ThreadsPanel search form (debounced input, hx-sync, morph swap, input id) + noMatch empty state; en/de keys `threads.search*`, `threads.noMatch.pre/post`; templ generate.
3. E.3: shell.js SSE guard — cancel "threads" pushes while `#wp-thread-search-input` holds text; shell.test.mjs spec.
4. F.1: audioout.js (setSinkId feature-detect, enumerate, persist `wp-sink`, devicechange, `#remote-audio` only) + main.js boot + phone.templ hidden picker + island i18n `audioOutput`/`audioDefault` en/de + audioout.test.mjs.
5. Re-run FULL gates: `go test -count=1 ./...`, island node suite, `nix fmt`, island-lint, buildflow, smoke 38/0.
6. TODO_LIST: add the two deferred rows (peer hub view train; Tailwind v4 spike — owner call) with concrete breakdowns.
7. CHANGELOG [Unreleased] entry covering A–F; FEATURES.md inventory rows.
8. Fill the plan doc verdict (honest: what shipped, what's stub-evidence).
9. One narrative commit; verify `git ls-remote` == HEAD (daemon may push first — verify, never assume).
10. AGENTS.md: island-test note if anything new generalizes (the classList real-DOM stub upgrade belongs in the island-test rules).
11. Stack browser E2E run (owner decision — see g3) to convert the four JS behaviors + search markup to browser truth.
12. Missed-call follow-up: badge should probably also clear when the island's own Recent calls opens (currently History tab only) — small, decide + do.
13. Typeahead polish: highlight the matched substring in suggestions; feed the History tab's local log into suggestions alongside shared contacts.
14. Search polish: server-side result cap + "search took N ms" debug log; decide whether remote LIKE should normalize formatting (strip spaces/dashes) before matching.
15. Jump-chip polish: also count pushes that arrive while data-page != "0"? (Currently cancelled by the paging guard — deliberate; document it.)
    16-50. (Standing backlog unchanged — see the 20:19 report's f1–f50; nothing there was touched by this session. Highest remaining after this train: v2.6.0 fold/release chain, owner-calls batch, erraudit seam conversions config.go/store-messages.go/pbx-client.go, CRM follow-ups a–h, send-failure C/E/F.)

## g) Questions I cannot figure out myself

1. **Thread search URL semantics:** ephemeral filter (current design —
   no push-url, deep link `/messages?q=` still renders filtered) or
   push the query into history so back-button undoes a search? Browser
   back/forward behavior is the owner's daily UX call.
2. **Missed-call semantics:** does a DECLINED (rejected) call count as
   missed? Current implementation says NO (a reject is a seen call).
   And should the badge clear on the island's Recent-calls panel too,
   or ONLY on the History tab (current)?
3. **Browser E2E now or at fold?** The stack repo's browser E2E is the
   only browser-truth gate for this session's four JS behaviors plus
   the search form markup. Run it immediately after E+F land (costs
   ~4 min, catches stub-vs-browser drift early), or fold into the next
   stack bump/re-pin like the previous two trains did (which is the
   pattern that produced the "twice-skipped declared gate" incident)?

---

_`.md` per standing instruction. HEAD at freeze: daemon chore commits
hold all session work; no narrative commit yet. Everything claimed
green above was re-run in THIS session with the documented wrappers;
nothing was claimed from memory._
