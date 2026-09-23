# SUPERB: The UX raw-ideas train (2026-09-22)

Train owner: agent session. Trigger: the owner pasted the
"Composer/UX raw ideas (2026-09-22 trains, unshipped)" list with an
execute order. Six of the eight ideas are implementable without owner
input; two are owner-gated by design and get refined TODO rows instead.

## Context and evidence

- Dial typeahead is the list's own "recommended next train" (17:53 f7).
- Jump-to-latest, missed-call badge, thread search, hover timestamps and
  the audio output picker are all S/M items from the same brainstorm.
- Peer hub is an M-L view train of its own; the Tailwind v4 spike is a
  recorded OWNER call (deep-dive report) — both go to TODO_LIST, not code.
- Constraint set unchanged: strict CSP (no inline anything), served-DOM
  contract (TestServedPageHoldsTheDomContract parses docs/dom-contract.md),
  morph-swap surfaces need stable ids, en output stays byte-stable where
  pins exist, shell survives JS failure, island module graph stays acyclic.

## Pareto breakdown

| Layer     | Items                                            | Rationale                                        |
| --------- | ------------------------------------------------ | ------------------------------------------------ |
| 1% → 51%  | A hover timestamps (XS) + B jump-to-latest (S)   | Instant polish on the two hottest surfaces       |
| 4% → 64%  | + C missed-call badge (S)                        | The only missed-call signal today is a transient |
| 20% → 80% | + D dial typeahead (M, the recommended train)    | Highest daily value, zero round-trips            |
| the rest  | E thread search (M), F audio output picker (S-M) | Complete the list                                |
| deferred  | Peer hub (M-L view), Tailwind spike (owner call) | TODO_LIST rows, not code                         |

## Decisions this train

- **A**: new `fullStamp(lang, t)` helper (de "02.01.2006 15:04", en
  "Jan 2, 2006, 4:09PM") as the `title` on thread-row relative times and
  bubble clock meta. New attribute output — no byte-stability churn.
- **B**: server renders the chip hidden inside a new
  `.wp-transcript-wrap` (position:relative) around `#thread-transcript`;
  shell.js §3b measures near-bottom BEFORE the sse swap, then scrolls or
  shows "↓ N new" (language-neutral, call-badge precedent). Returning to
  bottom hides + resets. No new ids.
- **C**: island dispatches `wp:call-missed` on the two genuinely-missed
  paths (caller gave up pre-answer; accepted call died before
  established) — a user REJECT is not missed. shell.js renders a
  `#missed-badge` header pill ("missed · N", English shell copy), cleared
  when the History tab opens. Client-only state: the PBX CDR has no
  missed flag, and the badge is session-scoped by nature.
- **D**: new island module `typeahead.js` over `PBX_CONFIG.contacts`
  (sharedContacts), ranked name-prefix < name-contains < number-contains,
  capped at 6, full keyboard support (↑/↓/Enter/Escape), listbox created
  by JS (green-dot precedent: served DOM untouched). main.js stays the
  only importer — graph stays acyclic.
- **E**: `store.SearchThreads` (LIKE over `threads.remote` + any message
  body, `%`/`_`/`\` escaped, owner-scoped) → `messaging.ThreadSearch` →
  `messagesPanel` reads `?q=` (full page AND partial — deep links work).
  Search form on ThreadsPanel: hx-get /partials/messages, debounced
  input (300ms) + submit, target #tab-content. SSE guard: while the
  search input holds text, `htmx:sseBeforeMessage` cancels "threads"
  pushes (same pattern as the paging guard). New i18n keys in BOTH maps.
- **F**: new island module `audioout.js` — feature-detects
  `setSinkId`, lists audiooutput devices, persists the pick
  (`wp-sink`), applies to `#remote-audio` only (ring tones are
  room-alarms on purpose, not headset-bound). Picker hidden unless
  supported and >1 output. Island i18n keys en/de.
- **Shell vs island placement**: dial typeahead + audio picker are pure
  island UI (island modules); jump chip + missed badge + SSE search
  guard are tab/shell surfaces (shell.js) per the AGENTS rule.

## Fine plan (≤ 12 min each)

| #   | Task                                                                    | Verifies via                   |
| --- | ----------------------------------------------------------------------- | ------------------------------ |
| A.1 | helpers.go fullStamp + tests                                            | `go test ./internal/web/views` |
| A.2 | messages.templ titles (thread-when, bubble clock) + templ generate      | server render test             |
| B.1 | messages.templ transcript-wrap + chip; app.css                          | server render test             |
| B.2 | shell.js §3b chip logic                                                 | shell.test.mjs                 |
| C.1 | calls.js + connection.js wp:call-missed; dead callEnded key removed     | island node tests              |
| C.2 | shell.js missed badge + History-tab clear; app.css                      | shell.test.mjs                 |
| D.1 | island typeahead.js + main.js wiring + i18n-free list markup            | typeahead.test.mjs             |
| D.2 | app.css typeahead dropdown                                              | served asset                   |
| E.1 | store.SearchThreads + test                                              | `go test ./internal/store`     |
| E.2 | messaging.ThreadSearch + panels.go ?q= + templ search form + i18n keys  | `go test ./...`                |
| E.3 | shell.js SSE search guard                                               | shell.test.mjs                 |
| F.1 | audioout.js + main.js + phone.templ wrap + island i18n keys + templ gen | audioout.test.mjs              |
| G.1 | TODO_LIST rows: peer hub + Tailwind spike                               | review                         |
| G.2 | CHANGELOG, FEATURES, plan verdict                                       | review                         |
| G.3 | Gates: go test full, node tests, nix fmt, island-lint, buildflow        | green                          |

## Constraints (do not break)

- CSP: no inline styles/scripts; every element from createElement +
  classes; the chip/typeahead/badge ship styles in app.css.
- Served DOM contract untouched (all new nodes are JS-created or hidden
  attribute additions on existing surfaces; the wrap div carries no id).
- Morph surfaces: the chip lives OUTSIDE `#thread-transcript`; the
  search form lives OUTSIDE the `sse-swap="threads"` region; idiomorph
  keeps focus in the search input via restoreFocus.
- English rendered output byte-stable where pins exist (titles are new
  attributes; the chip text is language-neutral; shell copy English).
- Store queries stay owner-scoped; search LIKE patterns escaped.
- Island module graph acyclic: typeahead.js and audioout.js import
  config/ui/i18n only; main.js imports them.

## Verification gates

1. `templ generate` diff = expected files only.
2. `nix develop -c go test -count=1 ./...` green.
3. `nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs` green.
4. `nix fmt` clean; island-lint green (no new globals); buildflow green.
5. Smoke 38/0 against a loopback boot.

## Verdict

SHIPPED — all six implementable ideas (A hover stamps, B jump chip,
C missed-call badge, D dial typeahead, E thread search, F audio
output picker) landed on `main` and rode the v2.6.0 release (tag
`807ca0c`, 2026-09-23). Gates at train close (21:43 report): full
`go test -count=1 ./...` 14 pkgs, island suite 76/76, `nix flake
check` ALL PASS, BuildFlow 52/0, gitleaks clean; the release gates
re-ran the full battery green on the tagged tree. Two deferred ideas
stayed deferred by design: peer-hub single view and the Tailwind v4
coexistence spike (ROADMAP raw ideas; the Tailwind spike is
owner-gated).

Honest caveats, on record:

- **Stub-evidence gap**: every JS surface (typeahead, chip, badge,
  audioout, search guard) is node:test-verified only — the stack
  browser E2E has not exercised them yet (it is part of the open
  v2.6.0 release tail: two load-shaped E2E stalls, chained retry
  never fired).
- **Search is ASCII-case-insensitive only** (SQLite `LIKE`): a query
  for "MÜNCHEN" will not match "münchen". Unicode-insensitive search
  needs `lower()` collation or an FTS5 column — real design work,
  deliberately not patched (raw idea on the ROADMAP).
- Search UX depth (clear button, result count, URL persistence, CRM
  display-name matching) deliberately out of scope; owner question
  on `?q=` URL semantics is open on the ROADMAP.
