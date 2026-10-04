# Status — island cascade-leak fix (UI-redesign train, T3 continuation)

_2026-10-04 12:34 · HEAD `a379976` (my fix, **pushed**, verified vs `origin/main`) · this report covers THIS session's run only._

## Context

Resumed mid-T3 (visual verification of the tw.css adoption layer). Round-1 shots had shown the adopted
Buttons (composer Send, thread-head Call) still rendering as island "keys": the cascade-layer leak —
island/style.css's unlayered bare element rules (`button`, `input`, `h1`, `label`, focus groups) beat
Tailwind's layered utilities at any specificity — made the component adoption visually INERT.
This session designed and applied the root fix: scope the island stylesheet to `.island`.

## a) FULLY DONE

1. **Todo list recreated** (T1/T2 complete, T3 in_progress, T4–T8 pending) per handoff.
2. **Git state verified**: working tree clean at start, `afa46ff` pushed; the unpushed `f466945` was the
   prior status doc (daemon pushed it later).
3. **Full blast-radius research before touching anything** (the session's core work):
   - Island root identified: `<div class="island">` inside `<aside class="wp-island">`
     (layout.templ:111, phone.templ:10). `#toasts` and `#remote-audio` live INSIDE `.island`
     (phone.templ:133-135) — so scoping to `.island` covers every island DOM surface, including
     JS-created toasts. The shell.js lightbox close button (`wp-mini`, appended to `body`) does NOT
     need island styles — app.css owns `.wp-mini`.
   - Inventoried EVERY tab-region element that currently freeloads on the leak:
     - `.wp-mini` (voicemail/contacts/history/fax/flags/retry/welcome-dismiss/lightbox): fully
       self-sufficient in app.css (border, bg, color, radius, padding, font, cursor, transition). ✓
     - `.wp-chip` (both app.css definitions ~714/~1867), `.wp-snippet-item`, `.wp-theme-toggle`,
       `.wp-jump-latest`, `.wp-welcome-dismiss` (already carries `wp-mini`, pinned byte-exact in
       panels_test.go:220 — no markup change needed): all self-sufficient. ✓
     - `button.wp-primary`/`button.wp-danger` (error reload, contacts/voicemail mini-danger): own
       padding 9px 16px etc. in app.css:758-800 — fully self-sufficient. ✓
     - Island JS-created buttons (calls.js `mkBtn` — `ghost`, `mute-btn`, `hold-btn`,
       `hangup-btn`, transfer row `ghost small`; panels.js `ghost small call`): all inside `.island`. ✓
     - **One true orphan found**: `history.templ:60` filter submit `<button>` had NO class — it would
       have gone UA-gray after scoping. Fixed by adding `class="wp-mini"` (design-system-consistent,
       no CSS clone, no byte-pin covers it).
     - No bare `<h1>`/`<label>` exists in any tab view (settings' label carries
       `.wp-snippet-quick-label`, self-styled).
     - Fax + messages composers ride `.wp-compose input/textarea` flex rules (app.css:667/675);
       adopted forms.Input/Textarea carry Tailwind classes. ✓
   - Cascade arithmetic verified for the island itself after scoping: `.island button` (0,1,1) and
     island class rules `button.primary` (0,1,1) are EQUAL specificity → later-in-file class rules
     win by order, exactly as today; `#keypad button` (1,0,1) unchanged; app.css's global
     `input`/`h1,h2,h3`/focus rules now correctly lose to `.island …` inside the island.
4. **island/style.css scoped** (committed `a379976`): `h1`, `label`, `input`, `input:focus`,
   `button`, `button:active`, and the 4-selector `*:focus-visible` group all prefixed `.island`;
   contract comment added at the top stating the cascade-layer rule (unlayered beats layered at any
   specificity) and why the `prefers-reduced-motion` block deliberately stays GLOBAL (killing motion
   page-wide under a user preference is correctness, not leakage; app.css has no reduced-motion block
   of its own and its tab transitions rely on it).
5. **app.css repaired**: the comment at the `button.wp-primary/wp-danger` rule no longer lies — it now
   states the true story (adoption works BECAUSE island scoped its element rules). Added
   `summary:focus-visible` to app.css's focus-visible group (tab-region `<details><summary>` — the
   snippet picker — would otherwise have lost its focus ring entirely after scoping, since app.css's
   group lacked summary and island's rule was scoped away).
6. **history.templ filter button** → `class="wp-mini"`; `templ generate` re-ran; scoped
   `nix fmt` clean (0 changed — formatting already correct).
7. **Committed + pushed**: `a379976` (app.css, island/style.css, history.templ, history_templ.go),
   verified against `origin/main`.

## b) PARTIALLY DONE

- **T3 verification loop** — edits are in, verification is NOT:
  - Island node test suite was launched (`nix run nixpkgs#nodejs -- --test
    internal/web/assets/island-tests/*.test.mjs`, background shell 051) — **no output returned before
    this report; result UNKNOWN**. Never waited on it. This is the biggest verification hole: the
    island stylesheet is verbatim-served and JS-tested; the suite must be green before T3 closes.
  - `go test` for `./internal/server/ ./internal/web/...` (DOM contract, byte pins, CSP) NOT run yet.
- **Round-2 capture** (binary rebuild → 14 shots → verify Send/Call render Primary green in BOTH
  themes, history filter looks right, summaries focusable): NOT started. T3 cannot close without it.
- **Chip-in-dark mystery**: un-diagnosed this session. `.wp-chip { background: var(--surface-2) }`
  (0,1,0 unlayered) should beat island's bare button today AND after scoping; the round-1 dark-theme
  shot still showed a light surface. Pure reasoning did NOT resolve it; needs the computed-style
  probe in the capture harness (or a minimal HTML repro). Note: scoping may have changed the picture
  (the chip's focus/active/inheritance context differs) — re-check in round 2 before digging deeper.

## c) NOT STARTED

- T3 closure items: computed-style probe idea for the capture harness; `.wp-chip` diagnosis;
  adopted-input double-edge nit (island's `input:focus` outline + border-color:transparent used to
  stack on the adopted inputs' ring — scoping should have REMOVED it; verify in shots rather than
  assume).
- T4 (gates: repo-wide `nix fmt`, full suite, buildflow), T5 (ThemeColor hexes `#0f766e`→accent,
  favicon tint), T6 (smoke 48-check, AGENTS.md, CHANGELOG), T7 (WCAG token-pair ratios),
  T8 (i18n confirm, dedup-registry sweep-log line, FEATURES/TODO_LIST).

## d) TOTALLY FUCKED UP

- **Ran a verification command and walked away from it.** The island test suite went to background
  (shell 051) and I never collected it — at report time it is still "running" with zero output.
  Either it hangs (node waiting on something) or output is buffered; either way the #1 rule
  ("test after changes") was left open-ended on a contract-sensitive file.
- Minor: first multiedit pass on island/style.css failed 2 of 8 edits (`input {` and the bare
  `button:active {` — the latter appears 3× in the file; uniqueness rule caught it). Caught and fixed
  immediately with context-anchored edits, no harm done — but the uniqueness check should have been
  done before writing the edit list.
- Nothing else. No pinned contract touched: DOM ids, byte pins, CSP, verbatim serving, token mirror,
  i18n — all untouched by the diff (11+29+2+4 lines across 4 files).

## e) WHAT WE SHOULD IMPROVE

1. **Never leave a background test uncollected** — if a command backgrounds, the next action is
   `job_output` (wait=true), before any other step.
2. **Compute edit uniqueness first**: grep candidate old_strings for occurrence count before
   batching multiedits.
3. **Capture harness gap (recurring)**: cascade victory is not provable by selector-existence tests
   (twcss_test.go's known limit). Add a computed-style assertion per shot to ui-capture.py
   (e.g. assert Send button's computed background ≈ accent token) — would have caught the round-1
   inertness automatically.
4. **AGENTS.md should pin the cascade-layer contract** (island styles scoped to `.island`; unlayered
   vs layered rule) — it currently documents the coexistence verdict ("unlayered app.css wins") but
   not the scoping obligation that the adoption layer depends on. Planned for T6.
5. The `input {` scoping changed tab-region input metrics (island leak gave padding 11px 12px /
   radius 10px; app.css gives 9px 12px / radius-sm 7px). Judged correct (shell owns the shell), but
   round-2 shots must eyeball the search box, settings, fax number field for layout regressions.

## f) NEXT (priority order)

1. Collect island test job 051; if hung, re-run foreground with a timeout and diagnose.
2. `nix develop -c go test -count=1 ./internal/server/ ./internal/web/...` (DOM contract + pins).
3. Re-run full suite later; re-attribute the known concurrent-session `internal/arch` drift failure
   if it still exists (note: the concurrent train has CHANGED identity — it is now a PASSKEY train:
   `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`, go.mod/go.sum + userauth in flight,
   uncommitted app/server/userauth edits in the tree — do NOT stomp).
4. Rebuild binary (`nix build .#webphone`) → round-2 capture → verify: Send/Call Primary green both
   themes; history filter compact; adopted inputs single-edge focus; chip dark-correct.
5. Diagnose chip-in-dark with computed-style probe if still wrong.
6. Batch remaining nits into ONE CSS pass (avoid daemon-trample piecemeal edits).
7. T4 gates (nix fmt repo-wide, full suite, buildflow), T5 ThemeColor/favicon, T6 smoke + AGENTS.md
   (cascade contract!) + CHANGELOG, T7 WCAG ratios, T8 docs sweep + dedup sweep-log.
8. Surface the 3 owner questions again (island-surgery timing — now MOOT for this fix since it's
   committed; palette sign-off; templ-components v1.19.4 vs local bump).

## g) QUESTIONS FOR THE OWNER

1. **Chip mystery budget**: the `.wp-chip` light-in-dark rendering has survived two rounds of
   reasoning. OK to spend a capture-harness iteration on a computed-style probe (adds an assertion
   utility to ui-capture.py, permanent value), or should I timebox it to one minimal repro and move on?
2. **Tab-input metric shift**: scoping moves tab inputs from the island leak look (11px padding,
   10px radius) to app.css's (9px, 7px radius-sm). Accept as the designed end state, or do you want
   the shell's input look bumped to match the old metrics (one token change, both sheets)?
3. **Palette sign-off** (carried from round 1): signal-green `--accent #17a467` on graphite — approved
   as-is before T5 stamps it into `ThemeColor`/favicon, or do you want a variant pass first?
