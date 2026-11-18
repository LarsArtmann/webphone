# Notes autosave train — session status

> Snapshot-rot header: claims verified as of 2026-11-18 18:39 CET.
> Scope: THIS session's run only (the "notes save themselves as you
> type, event-sourced, 200-300ms debounce" train). No repo-wide
> harvest — the living inventory stays TODO_LIST/FEATURES.
>
> SUPERSEDED (2026-11-18, docs honesty): the closeout report
> `2026-11-18_21-55_notes-autosave-closeout-and-self-review.md` owns
> this train's final state. This report's central "next" claim, the
> shell.js retry loop still being live, was overtaken by events: the
> carried/landed fix, both test-file repairs, and every gate landed
> 19:35-21:55 the same evening. Sections d/e below read as the
> mid-flight diagnosis, not the end state.

## The ask

Owner: "Notes should save themselves as soon as I type" + "be event
sourced with a small debounce like 200ms-300ms". Two note editor
surfaces exist: the History tab note box (server-rendered templ +
shell.js delegated handlers) and the island ended-call card editor
(`postcall.js` `noteRow`). The save lane `PUT /api/calls/{id}/note`
was ALREADY event-sourced (cqrs `note.save` decider → one
`domain.CallNoteSaved` event per save → projection → readmodel), so
"event sourced" needed zero new wiring — the work was autosave UX on
both editors plus journal/rate hygiene for typing-frequency saves.

## a) FULLY DONE (implemented + its targeted tests GREEN)

- **cqrs no-op suppression** (`internal/cqrs/callnotes.go`
  `decideSaveNote`): an unchanged body on an existing slot appends NO
  event — the journal records changes, not echoes. First save keeps
  minting the slot. Pinned by new
  `TestNoteSameBodySaveIsNotARevision` (same body → 0 events, changed
  body → 1, first save → 1). `go test ./internal/cqrs/ -run TestNote`
  GREEN.
- **Dedicated autosave rate bucket** (`internal/server`): PUT
  `/api/calls/{id}/note` moved off the shared contacts grade
  (60/min) onto its own `noteAutosaveLimiter` (120/min, burst 30) —
  sustained typing must not drain the contacts bucket it used to
  share. Constants `noteAutosaveLimit/Burst` in server.go, field in
  pages.go, mount in server.go. Build + server bucket/session tests
  GREEN.
- **History status span** (`internal/web/views/history.templ`):
  `<span class="wp-muted wp-note-status" data-note-status={CallID}
  aria-live="polite">` after the buttons; server-rendered empty (a
  freshly painted box IS the saved state). templ regenerated with the
  pinned compiler v0.3.1100; `panels_test.go` now pins
  `data-note-status="call-1"` + `wp-note-status`. Full views package
  GREEN (twice).
- **History autosave** (`internal/web/assets/shell.js` §2b7 rewrite):
  delegated `input` → 250ms debounce (per-call state seeded from the
  server-rendered `defaultValue`), `change` (blur) flush, `pagehide`
  flush with `keepalive`, quiet status line ("Saving…"/"Saved"/"Not
  saved"), failure toasts ONCE per episode, PUTs serialized per call
  (in-flight guard + completion re-dispatch), manual `[data-note-save]`
  button kept: immediate save + toast, skips the PUT when the body is
  already saved. Tuning seam `noteAutosave {delayMs}` exported on
  `module.exports`. Editor TS errors (closure narrowing) fixed with
  bound locals.
- **Island ended-card autosave** (`internal/web/assets/island/app/
  postcall.js`): `noteRow` reworked — input debounce 250ms, change
  flush, quiet `call-note-status` line, once-per-episode announce,
  serialized PUTs; exported `flushPendingCallNotes(keepalive)` +
  module-level `pagehide` flush; manual button is now the "done"
  affordance (saves AND collapses; collapses WITHOUT a PUT when the
  body is already autosaved). `calls.js` Close button flushes pending
  notes BEFORE `card.remove()` (hand-written content never silently
  expunged).
- **i18n + style**: 3 new keys in BOTH maps
  (`callNoteSaving`/`callNoteStatusSaved`/`callNoteUnsaved`, en+de —
  parity tests enforce); `.call-note-status` (+ `:empty{display:none}`)
  in island/style.css. No app.css touch (wp-muted reused on the shell
  side — no login.css rebuild owed).

## b) PARTIALLY DONE

- **postcall-note.test.mjs**: 5 new specs written (autosave+status,
  burst coalescing, failure episode, flush + pagehide keepalive,
  done-button collapse). First run: 5 pass / 1 fail — the fail was the
  infinite-retry loop (see d), now fixed in postcall.js; **the file has
  NOT been re-run after the fix**.
- **shell.js History autosave**: implemented but its mirror of the
  retry bug is NOT yet fixed (see d) and shell.test.mjs autosave specs
  are NOT yet written.

## c) NOT STARTED

- shell.test.mjs: History autosave specs (debounce PUT, coalesce,
  failure toast-once, pagehide flush, manual-save no-double-PUT when
  settled).
- Full island suite re-run (all *.test.mjs — calls tests may react to
  the Close-handler flush import).
- Type gate `nix develop -c tsc -p . --noEmit`; island-lint (oxlint).
- `nix fmt` over island/** + shell.js + css (treefmt owns them;
  oxfmt-vs-treefmt trap documented in AGENTS).
- Full `go test -count=1 ./...` (only cqrs/TestNote*, server subsets,
  full views package ran).
- CHANGELOG `[Unreleased]` line (code-path change — changelog bar).
- AGENTS.md notes bullet update (autosave semantics, no-op suppression,
  dedicated limiter, status-line vocabulary, flush contract).
- buildflow battery; commit (narrative, staged+committed in ONE call
  per repo convention).

## d) TOTALLY FUCKED UP (diagnosed; one fixed, one NOT)

- **Infinite retry loop** in the autosave completion hook: on a FAILED
  save the hook compared `live.value !== saved` — true forever on
  failure — and re-ran the save immediately: unbounded PUT loop, hung
  the island test file for 6 minutes ("Promise resolution is still
  pending"). postcall-note.test caught it.
  - postcall.js: FIXED (carried/landed logic — only input NEWER than
    the flight's body re-dispatches; success carries immediately,
    failure re-arms the debounce once, bounded).
  - **shell.js `runNoteSave` still has the SAME bug** (`if (live &&
    !state.timer && live.value !== state.saved) runNoteSave(...)`) — a
    failing server + typing pauses = unbounded PUT loop against
    History. THE next action; the carried/landed fix ports verbatim.

## e) WHAT WE SHOULD IMPROVE (this session's lessons)

- **Symmetric code, asymmetric fix**: I wrote the same flawed pattern
  in two files, the test caught it in one, and I fixed only that one
  before reporting. A diagnosed bug must be fixed at EVERY site in the
  same breath — the twin is a known-bad shape I left live.
- **Batch-then-test**: I wrote all 5 island specs before the first run;
  a per-spec run would have localized the loop in seconds instead of a
  6-minute hang (and I let a background `nix run nodejs` burn 6 min
  before reading it).
- **First-draft retry design was structurally wrong**: "re-run when
  value differs from saved" is unbounded on failure by construction;
  "re-run only when input is NEWER than the flight" is the terminating
  invariant. Should have derived that before writing either file.
- modern-web-guidance had no autosave-specific guide (searched; forms
  guide retrieved, nothing applied) — the value this round was the
  local patterns (300ms composer-draft debounce precedent), not the
  catalog.

## f) NEXT (ordered; ~the honest 25, not padded to 50)

1. Port the carried/landed retry fix to shell.js `runNoteSave` (d).
2. Re-run postcall-note.test.mjs — expect 6/6.
3. Write shell.test.mjs History autosave specs (use
   `shellApi.noteAutosave.delayMs` seam).
4. Run the FULL island suite (calls tests + i18n parity + type-gate
   pins).
5. `nix develop -c tsc -p . --noEmit` (the only authority).
6. `nix fmt`, then `bash scripts/format-check.sh` (treefmt cache trap).
7. Full `env -u GOEXPERIMENT nix develop -c go test -count=1 ./...`.
8. CHANGELOG `[Unreleased]` entry.
9. AGENTS.md: update the post-call-notes bullet (autosave, suppression,
   limiter, status line, flush contract, the retry invariant).
10. Consider a server-side test pinning the PUT's dedicated limiter
    (the bucket-test pattern), if the suite has a home for it.
11. buildflow battery inside `nix develop`.
12. Commit (narrative message; stage+commit+verify in ONE call).
13. Optional: smoke boot to eyeball the History status line end-to-end.
14. Optional: ui-capture shot of the History note box with a live
    status ("Saved") so the surface lands in the visual ledger.
15. Decide + record: does the AI-draft insert autosave? (question g2
    below — doctrine currently says no.)
16. Later/optional: debounce could ride one shared constant for both
    editors (250ms is duplicated as two tuning seams — fine for tests,
    but one exported shared value would prevent drift if ever changed).

## g) Questions for the owner (cannot be derived)

1. **The explicit "Save note" button**: keep it as the "done"
   affordance (immediate save + collapse on the island card; toast on
   History), or remove it now that autosave exists?
2. **AI-draft insert**: the doctrine "the AI never writes the note —
   the save is the owner's keystroke" means an inserted draft sits
   UNSAVED until the owner types/clicks. With autosave, should an
   explicit draft insert now save too (it is still an owner action),
   or does the doctrine stand?
3. **Failure loudness**: today a failed autosave = one warn toast per
   episode + a persistent "Not saved" status line + automatic retry on
   the next keystroke. Loud enough, or do you want a visible Retry
   affordance (like the failed send bubble's Retry/Dismiss)?

## Verdict lines for the impatient

- Server side (event-sourced journal + rate bucket): DONE, targeted
  tests green.
- Island editor (postcall.js): DONE incl. loop fix, tests written but
  not re-run.
- History editor (shell.js): implemented, KNOWN LIVE BUG (retry loop
  twin), tests not written.
- Gates/docs/changelog: not started.
