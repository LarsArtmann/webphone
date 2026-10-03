# Master-todo train — session 4 status (2026-10-02, ~08:30 → 09:03)

Execution order: `docs/planning/2026-10-01_17-32_SUPERB-master-todo-pareto-execution-plan.md`
(T01–T27). Session 3 report: `2026-10-02_08-26_master-todo-train-session3-status.md`.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): T13 closed green (the CJS-vs-ESM
> `?case=` discovery) and T14 shipped whole (player engine, fax timeline +
> resend, the C4 decline recorded); the remainder (T15–T19, T21–T23, T25.1)
> executed by sessions 5–8; the gates closed at the 12:57 close-out. Per-item
> verdicts inline.

---

## a) What was completed this session

### T13 — CLOSED (was ~80% at session end)

The one failing E3 test ("a plain / load restores the remembered tab")
is green, and its root cause was a REAL harness discovery, not a test
bug:

- **Root cause**: `shell.js` has no ESM syntax, so node parses it as
  **CommonJS** — and the CJS module cache keys on the resolved path
  WITHOUT the query string. Every `import("../shell.js?case=…")`
  returned the cached module and re-evaluated NOTHING (probe: 0
  listeners registered). The two "passing" E3 guard tests were
  VACUOUSLY green. Island `.js` modules contain `import`/`export`, so
  module-detection parses them as ESM where the query IS the cache key
  — that is why the same pattern works everywhere else.
- **Fix (sanctioned refactor)**: `restoreLastTab` is now a named
  function called immediately at load, exported through a
  CJS-guarded seam (`if (typeof module !== "undefined" && …)`,
  no-op on the wire) alongside `refreshNav` (and later `vmPeaks`/
  `vmClock`). The three E3 specs call the seam directly; the guard
  test now honestly exercises the welcome-hint path (routed selector
  installed) plus a new garbage-tab guard.
- **E8 pinned**: `refreshNav` calls `scrollIntoView({block:"nearest",
  inline:"nearest"})` on the active link; negative case (node without
  `scrollIntoView`) covered. NOTE: `refreshNav` reads `dataset.tab` —
  disconnected from attributes in the stub world.
- `module` added to the island-lint `globals` block (the seam tripped
  `no-undef`; oxlint gate green again).
- Lessons entry added (docs/lessons.md, Tooling traps): the CJS-vs-ESM
  `?case=` trap + "prove re-evaluation with a counter before trusting
  the pattern".
- Stack-E2E obligation note appended to docs/release-runbook.md
  (T11–T13 served-markup deltas: head preloads, island gear, history
  panel).
- Verification: shell tests 28/28 → island suite 144/144 → oxlint
  green → `nix fmt` → full `go test ./...` green.

### T14 — COMPLETE and verified (M13 + M14)

**M14 fax depth (C5 + C6 shipped, C4 declined with rationale):**

- `fax.Service`: the gateway-driving tail of `Send` extracted into
  `drive()` (one home for self-send refusal + gateway + status
  transitions); new `Resend(owner, id)` — failed OUTBOUND jobs only,
  reuses the spooled PDF (existence probe first: a loopback gateway
  would otherwise "transmit" a resend whose PDF vanished), creates a
  NEW job, keeps the original failed row as evidence.
- Handler `resendFax` + route `POST /fax/{id}/resend`; owner-scoped
  pre-fetch gives foreign/missing ids a plain 404 (initial version
  leaked the lookup error into sendFailure → 422; test caught it).
- `FaxRow`: outbound rows render a 3-step status timeline
  (`FaxSteps`: queued → sending → delivered, failed state in
  danger-red; aria-hidden — the badge beside it carries the textual
  truth; title adds the last-transition time). Failed outbound rows
  gain a Resend button. Status text/class mapping extracted to
  `faxStatusText`/`faxStatusClass` (badge now uses them — one home).
- **C4 (PDF thumbnail) DECLINED**: pdf.js (+1.2 MB always-loaded JS)
  violates the lean-serving posture; poppler = deployment-shape change
  for a cosmetic; no sane pure-Go rasterizer; fake placeholder =
  dishonest UI. Rationale + re-open trigger:
  `docs/planning/2026-10-02_09-05_fax-thumbnail-decline-note.md`.

**M13 voicemail playback (C1/C2/C3/C9/C10 shipped):**

- `voicemail.templ`: row restructured into a two-line action card
  (`wp-vm-row`; caller + actions top, player bottom) with player
  chrome: play button (uuid id, `data-label-play/pause` carrying BOTH
  session-language labels so the JS toggle honors the per-extension
  language invariant), waveform `<canvas role="slider" tabindex=0>`
  with aria-valuemax=duration, speed button, `0:00 / 0:12` readout,
  and the `<audio>` now **controls-free** (native UI is only the
  fallback). `vmClock` Go helper mirrors the JS clock shape.
- `layout.templ`: the voicemail nav badge gained `id="wp-nav-vm-badge"`
  (C9 mirror target; conditional render, so deliberately NOT a
  dom-contract id).
- `shell.js` §3i — the player engine: delegated clicks (swap-proof),
  lazy per-audio wiring, single-active-player, WebAudio fetch+decode →
  `vmPeaks` (48 bars) → canvas draw colored from CSS tokens
  (`--accent-strong`/`--muted`), pointer seek + drag-to-scrub,
  keyboard seek ±5s, speed ladder 1→1.5→2→0.5×, timeupdate repaint
  that SELF-HEALS the `wp-playing` row class after morph wipes, and
  the honest fallback: no WebAudio/fetch/canvas → native controls on
  - chrome hidden + one English shell toast. Unread clears on play
    and decrements the nav badge (optimistic, grounded in the play
    action; next server render converges).
- i18n: `fax.resend`, `vm.play/pause/speed/scrub` — en AND de.
- CSS (app.css): `.wp-vm-player` (grid-column 1/-1), `.wp-vm-wave`,
  `.wp-vm-time`, `.wp-row.wp-playing`, `.wp-fax-steps` states.
- Tests: Go — `TestFaxResendFailedJobCreatesNewJob` (provider flips
  refuse→accept across phases; asserts evidence row survives, 2 rows,
  new job sending), `TestFaxResendRefusesNonFailedRows` (transmitted
  422 + reason, inbound 422, foreign 404),
  `TestVoicemailRowsRenderPlayerChrome` (full chrome contract incl.
  no-controls pin and no-chrome-without-audio_url). Island — `vmPeaks`
  unit (found and fixed a REAL left-fill bug: bars beyond the samples
  now stretch from the nearest valued neighbor), `vmClock` unit,
  click-to-play + honest-fallback (async: the fallback rides a
  microtask), playing/label-swap/unread+badge, speed-ladder walk.

Verification at close: island suite **149/149** (re-run AFTER the last
`nix fmt`), full `go test -count=1 ./...` green, templ regenerated,
`nix fmt` clean, tree auto-committed and pushed (`db68c50`).

## b) What is INCOMPLETE / NOT verified

~~- **T14 leftovers**: the runbook's stack-E2E obligation note covers~~ done in part — the standing runbook rule covers markup deltas generically
  T11–T13 only — it must be extended for T14's markup deltas (fax row
  timeline/resend, voicemail player chrome, no more native audio
  controls in rows). Not yet done.
~~- `nix flake check` NOT run this session (island-lint green via direct~~ resolved by events — later flake checks green
  oxlint; module checks untouched but unproven since T13/T14 edits).
~~- BuildFlow NOT run this session.~~ resolved by events — the later full batteries are green
~~- No live-binary smoke of the new markup (deferred to the final-gates~~ resolved by events — the 12:57 closeout battery smoked the fresh binary
  checkpoint, as with T12).
~~- The whole train after T14 is untouched: T15→T19, T21→T23, T25,~~ done — sessions 5–8 executed the whole remainder (T15–T19, T21–T23, T25.1, harvest, gates)
  HARVEST, final gates (see f).

## c) Unexpected findings / deviations from the plan

~~1. **The `?case=` fresh-import pattern is a trap for CJS targets** —~~ done — the lesson is recorded (docs/lessons.md Tooling traps; this session)
   and shell.test.mjs's two E3 guard tests had been green for the
   wrong reason. Fixed + lesson recorded. No other test file uses the
   pattern against a CJS module (audited: island modules are ESM).
~~2. **`vmPeaks` shipped with a left-fill bug** in its first cut (empty~~ done — the unit test caught it pre-release; fixed + pinned
   bars carried `0` instead of stretching from the nearest valued
   bar); the new unit test caught it before any release exposure.
~~3. **Resend 404-vs-422**: the service's owner-scoped Get error~~ done — the handler pre-fetch shipped (honest 404 semantics)
   classified as a family verdict through sendFailure; the handler now
   pre-fetches for honest 404 semantics (same shape as faxDocument).
~~4. **C4 declined** (see a) — the plan's 14.3 said "C4–C6"; C5+C6~~ record stands — the decline note + re-open trigger are recorded (docs/planning fax-thumbnail-decline-note)
   shipped, C4 is a documented NO-GO pending owner ratification.
~~5. Judgment calls, all documented in code: play button glyph swap~~ record stands — the judgment calls are documented in code
   keeps aria-label in the session language via server-rendered data
   attributes (D3 English-copy rule does NOT apply — the labels are
   server-rendered per-lang, JS only swaps between them); the fallback
   toast is English (shell policy); `wp-playing` self-heals from the
   audio element (truth) rather than persisting client state through
   morphs.

## d) Screw-ups this session (process lessons)

~~1. A no-op edit glued a comment onto the following `func` line in~~ process record
   actions.go (old/new differed only by a newline) — caught by the
   immediate sed check, fixed before build. Lesson: never use an edit
   whose old_string and new_string differ only in whitespace unless
   that IS the change.
~~2. First multiedit on fax.templ left `</article>` mis-indented (my~~ process record
   old_string guess, not a dump) — re-viewed and fixed. Lesson
   (again): dump exact bytes before templ edits; the daemon reforms
   files between reads.
~~3. My first speed-ladder assertion was an unreadable nested~~ process record
   map/&&-expression (intuition-based, exactly the session-3 trap) —
   rewritten to record the label per click. Derive assertions from the
   implementation, keep them dead-simple.
~~4. Test helper built elements via `createElement` and set `.id` — but~~ process record
   `getElementById` only resolves REGISTRY entries; the shell resolved
   fresh empty nodes. Fixed by building stubs FROM `doc.getElementById`
   instances. (Same class as session-3's dataset/querySelector traps.)
~~5. The fallback test asserted microtask results synchronously —~~ process record
   needed `await tick`. Promise-chained production code ⇒ async tests.

## e) What could be improved

~~- Run `nix flake check` per TASK close-out (cheap subset: `nix build~~ process record
  .#checks.x86_64-linux.webphone-module` when only app code changed)
  instead of hoarding it to the final gates.
~~- The resend handler's pre-fetch duplicates the service's Get — an~~ record stands — the double read is the accepted cost of clean 404 semantics
  acceptable cost for clean status semantics, but a typed
  NotFound from Resend would remove the double read.
~~- `vmDraw` reads computed CSS tokens per draw (per timeupdate, ~4 Hz) —~~ record stands — not worth it until a profile shows it
  cacheable per row if it ever shows on a profile; not worth it now.
~~- The E2E-obligation note should become a standing checklist the~~ done — the standing runbook rule IS the checklist now
  moment ANY served markup moves, not a close-out step (this session
  nearly forgot it for T14).

## f) Next ~50 items (execution order)

~~1. Extend the runbook E2E obligation note with T14's markup deltas.~~ done in part — the standing runbook rule covers markup deltas generically; no per-train enumeration was added
~~2. Quick gate: `nix build .#checks.x86_64-linux.island-lint` (or full~~ resolved by events — later flake checks green
   `nix flake check`) to prove the tree in sandbox.
~~3. **T15 M17 feedback/trust**: J2 reconnect banner via `#wp-live`~~ done in part — the landed surfaces (offline banner, Retry/Dismiss, filter-aware empty state, politeness ladder, per-thread drafts); undo wasn't taken
   (registration lost → visible surface); J3 optimistic-send undo;
   J4 retry affordance inside the failed-state banner; J7 confirming
   sends do not lose drafts on navigation; J8 honest empty-state
   copy pass; J9 aria-live politeness audit (assertive vs polite).
~~4. T15 tests: island specs for the banner state machine + shell specs~~ done — the T15 surfaces carry specs (calls/shell/session tests; the AGENTS M17 contract)
   for undo/retry; failure-feedback table row for each new surface.
~~5. **T16 M15 visual tokens**: audit hardcoded colors in app.css /~~ done — mirrored light/dark token blocks (TestCSSTokenBlocksAreMirrored)
   island/style.css → tokens; mirror rule for :root/dark blocks.
~~6. T16 M20 theming: data-theme attr propagation to island surfaces;~~ done — the three-state theme persisted pre-paint (theme-preload.js); reduced-motion guards in place
   reduced-motion audit for new animations (wp-key-sent, morph).
~~7. T16 M26 shell sizing: fluid type scale clamp() pass; touch-target~~ done — calmer shell sizing + 44px mini targets
   audit ≥44px on wp-mini buttons (player + rows).
~~8. T16 tests: token mirror test, computed-style assertions where~~ done — the token mirror test pins the sheets
   cheap, screenshot-diff deferred to T23 harness.
~~9. **T17 M18 onboarding**: first-login hint layer (pointer to phone~~ done — the dismissible welcome note (T17)
   sign-in, tabs); dismissible, i18n en/de, session-language aware.
~~10. T17 M19 mobile extras: bottom tab bar active-state polish, safe-~~ done in part — T17 mobile legs; the iOS-keyboard extras weren't taken
    area insets, composer sticky-footer behavior on iOS keyboard.
~~11. T17 tests: island/shell specs for hint dismissal persistence~~ done — hint dismissal persistence pinned (T17 specs)
    (localStorage), a11y labels both langs.
~~12. **T18 M21/M22 SEAM design note FIRST** (docs/planning/…): enumerate~~ done — the T18 seam design (D1–D14)
    the store ALTER (M22), the schema_version mechanism, the carve.
~~13. T18: implement M21 (settings depth) + M22 (store change) per note;~~ done — snippets + thread flags shipped
    expect T25.1 to fire here.
~~14. T18↔T25 decision: schema_version lands in the same train (default)~~ resolved by events — same train (T18 + versioned migrations)
    or deferred (owner question, see g).
~~15. T18 tests: migration test (old row → new schema), store family~~ done — the migration/versioned-migrations tests (TestVersionedMigrations)
    pins, server handler tests for new settings surface.
~~16. **T19 M24 i18n/RTL**: dictionary audit for missing keys (the test~~ done — the unused-key test exists (AGENTS: parity + unused-key tests enforce)
    keeps en/de in sync — extend it to flag UNUSED keys too); RTL
    smoke via dir=attr on a rendered page (logical CSS properties
    audit: margin-inline etc.).
~~17. T19 M25 call depth: transfer/merge affordances in the island call~~ done in part — focus mode + device test shipped (A6/A10); transfer/merge affordances weren't taken (transfer rides REFER as-is)
    card (REFER semantics already pinned — read DTMF/REFER lessons
    first); hold-state parity for transfers.
~~18. T19 tests: island call-card specs for transfer state machine; Go~~ done in part — RTL groundwork shipped; the Go RTL render tests weren't taken
    render tests for RTL strings (no truncation regressions).
~~19. **T21 nix-review batch 2**: re-run the nix-review skill pass over~~ not adopted — below the bar (no second nix-review pass recorded)
    flake.nix + nix/*.nix (module split done in batch 1); checks.nix
    island-lint file list grew nothing this session (player lives in
    shell.js) — verify.
~~20. T21: fix findings in nix files; `nix flake check` after.~~ done — the batch-2 wins landed (module-output.golden + release.sh guard); the rest dispositioned individually
~~21. **T22 samber/do + dashboard follow-ups**: re-check AGENTS.md~~ process record — the AGENTS composition-root bullets are current
    composition-root bullets against internal/app for drift; dashboard
    refresh loop test still green; do-lifecycle conformance asserts.
~~22. T22: dashboard CSS rebuild ONLY if sources changed (tailwindcss_4,~~ done — the 2026-10-02 rebuild fixed a genuinely stale artifact (TODO health row)
    never v3 — AGENTS rule).
~~23. **T23 visual verification harness**: pick shape (Playwright?~~ done — scripts/ui-capture.py (chromedriver shape; 14 shots)
    chromedriver? — research what the stack's browser-e2e.py already
    reuses); golden-page screenshots for tab surfaces incl. the new
    player + timeline; wire into scripts/ + flake check (KVM-gated if
    it needs a browser).
~~24. T23: decide budget (screenshot count, runtime) before building.~~ done — budget decided (14 shots, LOCAL-ONLY by budget decision)
~~25. **T25 schema_version + carve** (if not landed with T18): version~~ resolved by events — versioned migrations shipped with T18; the carve trigger stayed unfired
    table + migration runner pattern; internal/server carve decision
    (file count trigger from M22's additions).
~~26. **HARVEST (docs-health)**: fold session-3 (f) + this report's (f)~~ done — the TODO rows + CHANGELOG/FEATURES moves + this v6 sweep
    into TODO_LIST.md; move shipped rows (T11–T14) to CHANGELOG
    Unreleased + FEATURES.md status flips (voicemail player, fax
    resend/timeline, history filters, E3/E8, modulepreload…).
~~27. HARVEST: mark the master plan's T11–T14 rows DONE with pointers.~~ done — the master plan rows marked DONE (12:57 close-out)
~~28. **Final gates**: `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`~~ resolved by events — the later full batteries are green
    (full), fix findings.
~~29. `nix flake check` (KVM backup VM included — quiet host, nothing~~ resolved by events — later flake checks green (incl. the KVM VM)
    else running).
~~30. Fresh-binary smoke: build `/tmp/webphone-bin`, boot loopback,~~ resolved by events — the 12:57 closeout battery smoked the fresh binary (48 checks)
    `python3 scripts/webphone-smoke.py --base http://127.0.0.1:…`
    (41+4 checks) — eyeball the fax + voicemail partials in the same
    boot (player chrome, timeline, resend button on a failed fax).
~~31. `git ls-remote origin main` == local HEAD end-state assert.~~ done — the ls-remote ritual is standing
~~32. Update AGENTS.md architecture bullets if T14 introduced durable~~ done — AGENTS carries the durable T14 rules (morph surfaces incl. voicemail re-fetch)
    rules (player seam exports, vm ids) — one line each, rules only.
~~33. Owner handover notes for T01/T02/T04/T09/T24/T27 (standing).~~ routed — TODO owner rows (the handover legs)
~~34. Re-measure erraudit tiers 1+2 if the month marker passes (next:~~ routed — standing watch (TODO watches row; next due 2026-10-22)
    2026-10-22) — not this train.
~~35. Consider a webhook-lane smoke probe if the gateway lane changed~~ record stands — the webhook-lane probe stays DECIDED-AGAINST (AGENTS)
    (it did NOT this train — resend rides the same seam; skip unless
    the bridge contract moves).

## g) Questions for the owner

~~1. **C4 fax thumbnail — ratify the decline?** pdf.js = +1.2 MB~~ record stands — the decline + re-open trigger are recorded; the owner word never came
   always-loaded JS (lean-serving posture says no); poppler in the
   module = deployment-shape change for a cosmetic; placeholder =
   dishonest. Re-open trigger recorded. Say the word if you want the
   poppler variant despite the closure growth.
~~2. **T18↔T25 sequencing** (carried from session 3): M22's store ALTER~~ resolved by events — schema_version landed in the same train (versioned migrations)
   fires T25.1 schema_version — land schema_version in the SAME train
   (my default: yes, one migration story, one verification) or defer
   T25 to its own train after T19?
~~3. **Voicemail first-play latency**: the waveform fetches the audio a~~ record stands — the double-fetch accepted (browser cache dedupes; no owner ruling)
   second time for WebAudio decode (the `<audio>` element fetches
   independently; browser cache should dedupe same-origin GETs). On a
   slow PBX link the first play could pay the transfer twice until the
   cache warms. Acceptable, or should the player fetch once into a
   Blob URL and feed both consumers (more code, one transfer)?

— End of report. Waiting for instructions.
