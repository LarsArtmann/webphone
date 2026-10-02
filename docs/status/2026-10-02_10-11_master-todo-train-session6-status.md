# Session 6 — Master-Todo Train: T17 closed, T18 mid-flight (store+seams green, views half-landed)

Date: 2026-10-02, ~09:35 → 10:11. Resumed from session-5 briefing
(report `09-34`), standing "finish everything" order, user's latest
directive: report now, then WAIT.

## a) FULLY DONE (this session)

### Housekeeping / verification

- Todo list recreated (16 items) per briefing; tree confirmed: session-5
  T15/T16 changes uncommitted, remote lagging (`ffaa03f` vs `f3b87d1`);
  daemon owns commits/pushes — end-of-train `git ls-remote` assert
  stays planned.
- Session-5 report §f/§g read. Owner questions unanswered → documented
  defaults stand (welcome-tab dismissal, localStorage persistence,
  schema_version same train as M22).

### Gap closure (session-5's queued item #1)

- oxlint island gate re-run after T15/T16 JS edits: exit 0. Island
  suite 152/152. Lint-after-JS-edits habit restored.

### T17 M18 — welcome dismissal (COMPLETE, all suites green)

- **Flash-free design**: `wp-welcome-dismissed` class lives on `<html>`,
  set by **theme-preload.js BEFORE first paint** (extended: reads the
  flag in the same try/catch), toggled by shell.js §1c on the dismiss
  click (localStorage `wp-welcome-dismissed=1` + root class — collapse
  applies without reload).
- layout.templ SignInHint: server-rendered dismiss button
  (`wp-mini wp-welcome-dismiss`) + compact line
  (`wp-welcome-compact-line`, hidden unless dismissed). Copy is
  per-session-language: i18n `welcome.compact`/`welcome.dismiss` en+de.
- app.css: `:root.wp-welcome-dismissed` hides body/points/hint/dismiss,
  reveals the compact line, tightens padding; `align-self` keeps the
  button from stretching in the flex column.
- Tests: theme-preload.test.mjs harness extended (classList capture) +
  2 new cases (dismissed collapses pre-paint; garbage/absent/thrown
  storage never sets the class); shell.test.mjs new case (click →
  localStorage + root class + stranger-click negative); Go
  `TestSignInHintRendersDismissAndCompactLineBothLanguages`.
  **Island count now 155** (152 + 3).

### T17 M19 — mobile/keyboard (COMPLETE, all suites green)

- **Viewport**: templ-components `layout.Base` hardcodes the viewport
  meta (base.templ:228, no PageProps field); HeadContent renders AFTER
  it → a SECOND viewport meta merges per the CSS Viewport spec. Added
  `<meta name="viewport" content="interactive-widget=resizes-content">`
  to headExtras (Chromium: keyboard resizes the LAYOUT viewport so the
  sticky composer rides above it; others ignore the key; no fork).
- `.wp-root` gains `min-height: 100dvh` (after the 100vh fallback) —
  no gap under the mobile URL bar.
- **Composer audit (no change needed)**: composer is in document flow
  (iOS focus auto-scroll reveals it); bottom tab bar already
  fixed + safe-area. No visualViewport JS — CSS-first honored.
- `TestServedPageSatisfiesStrictCSP` now pins the keyboard viewport
  meta on the served page.
- Runbook E2E obligation note extended T11–T15 → **T11–T17** (welcome
  button + keyboard viewport meta are served markup).
- Close-out gates: `templ generate` + `nix fmt` clean, full
  `go test ./...` green, island 155/155, oxlint 0.

### T18 — scope resolution + design note (COMPLETE)

- **Briefing scope error RESOLVED**: sessions 4/5 called T18 "M21
  settings + M22 store ALTER" — wrong. Master plan, UX pareto plan and
  TODO_LIST all say **M21 messaging richness + M22 pin/archive/mute**;
  session-4's "settings depth" was a mis-paraphrase. Proceeding on the
  authoritative scope.
- **Design note**: `docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md`
  — 14 decisions (D1–D14), the load-bearing ones:
  - D1 schema_version lands NOW (the "before first ALTER" gate already
    fired 2026-09-24: failure_kind/detail were ad-hoc ALTERs).
  - D2 M22 toggles are htmx actions in actions.go (markThreadRead
    pattern) → **zero new internal/server files → the T25.2 carve
    trigger does NOT fire mid-feature** (D3: carve stays trigger-gated,
    staged into internal/server/api when it does).
  - D4 three 0/1 columns on threads; D5 pin = sort-first; D6 archived
    leaves list AND search unless `?archived=1`; D7 **inbound
    auto-unarchives**; D8 mute is presentation-only (badge + nav total
    suppressed, unread truth stays; SSE unchanged).
  - D9 lightbox = `<dialog>` + shell.js delegated click; D10 attachment
    preview = EXISTS subquery in the thread summary; D11 snippets =
    new table (cap 100) + Settings-tab CRUD actions + server-rendered
    `<details>` picker + `data-snippet` chips filled by shell.js (zero
    JSON API); D12 **M21.6 scheduled sends NO-GO** (gateway has no
    deferred-send concept — filed with the cross-repo dependency);
    D13 M22.5 notification deep-link NO-OP (no message notifications
    exist); D14 stable row ids BEFORE stateful buttons (folded in).

### T18 — store + service + server seams (COMPLETE, green at checkpoint)

- **Versioned migration runner** (store/db.go): `schemaVersion = 2`,
  ordered `migrations` chain — v1 duplicate-tolerant baseline (today's
  CREATEs + the ad-hoc failure-column ALTERs; converges legacy DBs),
  v2 strict + transactional (threads pinned/archived/muted + snippets
  table + index). Single-row `schema_version` table; each step atomic
  (DELETE+INSERT stamp inside the tx). `TestVersionedMigrations`:
  fresh → latest; legacy (tables, no version row) → stamps + converges;
  already-latest rerun no-op; chain contiguity pinned.
- **Domain**: Thread += Pinned/Archived/Muted; new Snippet type +
  SnippetBrand/ID/Generate/Parse/Must (the contacts pattern).
- **store/messages.go**: AppendMessage upsert **auto-unarchives on
  inbound** (`excluded.unread > 0` CASE, commented); ThreadSummary +=
  flags + LastAttachment (EXISTS subquery); shared
  `threadSummaryQuery` builder (pinned-first ordering, archived
  filter); ListThreads (archived excluded), ListArchivedThreads,
  CountArchived, SearchThreads (archived excluded, pinned first),
  GetThread reads flags, `SetThreadFlag` (+threadFlagColumns map,
  ErrNotFound on miss).
- **store/snippets.go (NEW)**: List (oldest first), Save
  (INSERT..WHERE-capped at 100, replace-by-id via ON CONFLICT),
  Delete (owner-scoped); ErrSnippetListFull sentinel + family
  registration.
- **messaging.Service**: Threads/ArchivedThreads/ArchivedCount +
  SetThreadFlag (persists, then list-only SSE nudge).
- **server/notifier.go HARDENED**: MessagesChanged now SKIPS the
  transcript push for a zero threadID — the list-only nudge would have
  rendered an EMPTY transcript and **wiped the open conversation**
  (caught during design, guard + comment).
- **server/actions.go**: countUnread skips muted (D8);
  `setThreadFlag` (action map pin|archive|mute, `on` form value, 404
  owner-scoped, unread cache drop); `saveSnippet`/`deleteSnippet`
  (422 empty/long body + cap, toast + settings partial epilogue).
- **server.go**: Deps += `Snippets *store.Snippets`; routes
  `POST /messages/{id}/{flag}` (literal `read` pattern still wins),
  `POST /snippets/save|delete`. **app.go**: Snippets provider + Deps
  wiring. Whole repo `go build ./...` green; server+app tests green.

## b) PARTIALLY DONE

### T18 — views layer (messages.templ edits 1–3 of ~6 applied)

- LANDED: ThreadsPanelProps += Archived/ArchivedCount; ThreadsPanel
  archived-toggle links (deep-link safe: full renders read the same
  query param); ThreadRow stable wrapper id (`thread-<id>`, D14),
  flag buttons (pin/mute/archive + unarchive in archived view), muted
  badge suppression + 🔇/🔔 glyph, 📎 attachment preview prefix;
  threadFlagButton component + helpers.
- **KNOWN-BROKEN in the last edit batch (self-caught, NOT yet fixed,
  templ generate has NOT been run since)**:
  1. `@if summary.Thread.Archived {` is NOT templ syntax (bare `if`
     is) — templ generate will fail on it.
  2. `pickLabel(lang, activeKey, inactiveKey)` is a pass-through that
     ignores inactiveKey AND the pin/mute call sites pass a constant
     key — the label does not flip with the row state (button value
     posts correctly; label text is wrong for one of the two states).
- NOT YET: ThreadViewProps += Snippets; snippetPicker + quick-chips
  components; ThreadView head flags + composer chips/picker; Bubble
  image → lightbox wrapper; settings.templ snippets section +
  SettingsPanelProps + settingsPanel builder; ALL i18n keys (en+de:
  threads.pin/unpin/mute/unmute/archive/unarchive/archived/back,
  snippets.*, toast.snippetSaved/Deleted, lightbox.close);
  panels.go messagesPanel archived param + count + snippets into
  props; CSS (wp-thread-flag/wp-archived-toggle/chips/picker/
  lightbox); shell.js data-snippet fill + lightbox dialog; openapi
  entries for the three new routes; templ generate + suites.

## c) NOT STARTED (train remainder)

T19 (M24 i18n/RTL + M25 call depth), T21 (nix-review batch 2), T22
(samber/do + dashboard follow-ups), T23 (visual harness — research the
stack's browser-e2e.py first), T25 (schema_version LANDED with T18 —
only the carve trigger-gate remains), docs-health HARVEST (note:
TODO_LIST's "schema_version gated on first ALTER" row is now STALE and
must flip to DONE), final gates (buildflow full, quiet-host flake
check, fresh-binary smoke with T14–T17 eyeballs, `git ls-remote`
end-state assert).

## d) TOTALLY FUCKED UP (all self-caught, all fixed except the two flagged view bugs)

1. **defer paren loss** in applyMigration: the edit landed as
   `defer func(){ _ = tx.Rollback() }` (no invocation) — build error,
   fixed immediately. Lesson: build after EVERY edit batch.
2. **Backtick typo** inside ListArchivedThreads (string op arg) —
   build error, fixed.
3. **errorfamily constructor guess**: wrote NewRejectionf (does not
   exist) — the vendored library is New* family constructors +
   NewRejection; fixed by reading the library source.
4. **theme-preload test stub bug**: passed a bare function as the
   `storage` object (getItem undefined → try/catch swallowed the
   TypeError → test failed) — wrapped in an object.
5. **The two UNFIXED view-layer bugs** listed in (b) — `@if` syntax +
   pickLabel pass-through. Nothing generated or committed broken:
   templ generate simply has not run since.

## e) WHAT WE SHOULD IMPROVE

- **Run `templ generate` immediately after every .templ edit batch** —
  it is the only fast check for templ-syntax slips like `@if`; the Go
  build does NOT see .templ sources.
- **Briefings can mislabel scope**: session 4/5's "M21 settings" nearly
  shipped as the wrong workstream. When a briefing summary contradicts
  the authoritative plan doc, the PLAN doc wins — verify scope labels
  against source before building.
- **The mtime guard fired twice** (theme-preload.js, db_test.go) — the
  daemon touches files mid-session; keep re-Viewing immediately before
  every edit (already habit; it saved both edits).
- **Label-value coupling in generated buttons**: state-dependent copy
  must be computed from the SAME expression as the posted value
  (`wantInt(!Pinned)` + the label key) — centralize in the helper, not
  spread across call sites.

## f) Next up to ~50 (execution order)

1. Fix the two flagged messages.templ bugs: `@if` → bare `if`;
   pickLabel → state-selecting helper (single expression for value +
   label), threadFlagButton takes the resolved key.
2. Finish messages.templ: ThreadViewProps += Snippets;
   snippetPicker + SnippetChips; ThreadView head pin/mute/archive;
   reply composer chips + picker; Bubble image lightbox trigger
   (data-lightbox + full-size dialog img).
3. settings.templ: snippets section (list + quick marker + delete +
   add form with quick checkbox) + SettingsPanelProps.Snippets.
4. i18n.go: ALL new keys in BOTH maps (sync test enforces);
   panels_test byte-pins where the note's test map demands.
5. panels.go: messagesPanel reads `archived` param + ArchivedCount +
   Snippets into both props; settingsPanel reads Snippets.List
   (nil-safe for hand-composed test Deps).
6. openapi.json entries for /messages/{id}/{flag}, /snippets/save,
   /snippets/delete.
7. CSS: wp-thread-flag pressed state, wp-archived-toggle, chips,
   snippet picker, lightbox dialog (scrim token reuse).
8. shell.js: data-snippet delegated fill (composer textarea) +
   lightbox dialog open/ESC/close; shell.test.mjs cases for both.
9. templ generate + nix fmt; then the FULL T18 test map:
10. store tests: flag round-trips, auto-unarchive (inbound clears,
    outbound keeps), pinned ordering, archived exclusion from
    list+search, CountArchived, snippets CRUD + cap + replace-by-id.
11. server tests: toggle 401/404/204 + unknown-flag 404, mute drops
    badge from nav (countUnread), snippet save/delete happy + 422s +
    cap, settings partial re-render.
12. views tests: row controls both langs + stable ids, archived
    toggle, chips render from quick snippets, picker hidden when
    empty, lightbox attr on image attachments.
13. Island/shell tests: snippet fill preserves existing draft text
    (append vs replace — DECIDE: replace, matching data-sms prefill),
    lightbox open/close.
14. Runbook E2E obligation → T18 (served markup moved: row buttons,
    toggle, picker, chips, settings section).
15. AGENTS.md: one durable line for the thread-flags seam + snippets
    (rules only) once the shape settles.
16. T19 M24: extend the en/de sync test to flag UNUSED keys; audit the
    dictionaries.
17. T19 M24: RTL pass — render with dir=rtl, audit logical properties
    (margin-inline etc.), fix findings; note `.wp-welcome-points
    li::before { margin-right }` as a known candidate.
18. T19 M24: Go render tests for RTL strings (no truncation).
19. T19 M25: island transfer/merge affordances — READ the DTMF/REFER
    lessons first (FreeSWITCH executes transfers server-side; REFER
    semantics pinned).
20. T19 M25: hold-state parity for transfers (holdPending interplay);
    island call-card specs for the transfer state machine.
21. T21: nix-review skill batch 2 over flake.nix + nix/*.nix.
22. T21: verify island-lint file list in checks.nix grew nothing.
23. T21: fix findings; `nix flake check` after.
24. T22: AGENTS composition-root bullets vs internal/app drift check.
25. T22: dashboard refresh-loop + do-lifecycle conformance still green.
26. T22: dashboard CSS rebuild ONLY if sources changed (tailwindcss_4).
27. T23: research the stack's browser-e2e.py driver; pick the same.
28. T23: decide harness budget BEFORE building; KVM-gate if browser.
29. T23: golden-page screenshots (voicemail player, fax timeline,
    thread row controls) wired into scripts/ + flake check.
30. HARVEST: fold session-3/4/5/6 §f lists into TODO_LIST.md.
31. HARVEST: TODO_LIST "schema_version gated" row → DONE with pointer;
    T11–T18 → CHANGELOG Unreleased + FEATURES.md flips; mark plan rows.
32. Final gates: `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`
    (gomod-check vendor false-positive = documented exception).
33. Final gates: quiet-host `nix flake check` (KVM backup VM included).
34. Final gates: fresh-binary smoke — one boot, eyeball fax
    timeline/resend, voicemail player, filter-aware empty, failed-send
    Retry/Dismiss, welcome dismiss + compact, archived toggle, snippet
    chips.
35. Consider a smoke probe for `history?outcome=missed` empty-filter
    (only if trivially greppable).
36. `git ls-remote origin main` == local HEAD end-state assert.
37. Owner handover notes for T01/T02/T04/T09/T24/T27 (standing).

## g) Questions for the owner (cannot figure these out myself)

1. **Archive semantics (D7)**: I made an INBOUND message auto-unarchive
   its thread (a live conversation must not stay hidden). Alternative:
   archived stays archived until manually unarchived, even with new
   messages. Keep my default?
2. **Scheduled messages (M21.6)**: NO-GO documented — the gateway seam
   (loopback + webhook/Telnyx bridge) has no deferred-send concept,
   and a server-side scheduler adds a queue+timer lifecycle surface.
   Confirm the deferral, or do you want the scheduler built anyway?
3. **Carve staging (T25.2)**: T18 adds zero new internal/server files,
   so the carve trigger did NOT fire. When it does: new handler
   families → new `internal/server/api` package from day one (staged,
   D3), or one big-bang move of the four existing `*_api.go` families
   in a dedicated train (the TODO_LIST's original carve wording)?

— End of report. Waiting for instructions.
