# Session 5 — Master-Todo Train: T15 + T16 closed, T17 opened

Date: 2026-10-02, ~09:05 → 09:34. Resumed from session-4 briefing
(`db68c50`/`f3b87d1`), executed the standing "whole list" order.

## a) FULLY DONE (this session)

### Housekeeping / verification

- Todo list recreated (16 items) per briefing; tree verified clean at
  `f3b87d1`, remote lagging (`ffaa03f`) — daemon owns pushing.
- **Runbook E2E obligation extended T11–T13 → T11–T14** (fax
  timeline/resend + voicemail player chrome deltas) — the small debt
  the briefing flagged.
- **Suite-count mystery resolved**: briefing said 149 island tests;
  glob run summarized 142. Proven per-file + by counting ✔/✖ lines:
  **all 149 execute and pass** — the summary undercounts under
  `--test-force-exit` when late microtasks resolve after aggregation.
  Known quirk now; don't re-investigate.

### T15 — M17 feedback/trust (COMPLETE, tests green)

- **J2 SSE auto-recovery notice** (`session.js`): `sseAnnouncedDrop`
  flag — the "restored" ok-toast fires ONLY on `htmx:sseOpen` after a
  drop that itself was announced (the 3-failure threshold). Quiet
  flaps never narrate their healing. i18n `sseRestored` en/de.
- **J9 aria-live politeness audit → real gap fixed**: `onInvite` never
  announced through ANY live region — screen readers missed ringing
  calls entirely (hidden→visible banner swap is invisible to AT).
  `announce(msg, kind, { assertive: true })` (ui.js) sets `role=alert`
  on the toast node; the incoming call uses it (`incomingCall` key
  en/de). Policy pinned: everything else stays polite.
- **J3/J4 Retry + Dismiss in the failed optimistic bubble**
  (`shell.js` §3e): `.wp-opt-actions` row with two `wp-mini` buttons
  (English, D3). Retry removes the bubble + re-submits the CURRENT
  reply composer (draft was restored); Dismiss removes the bubble and
  keeps the draft editable. CSS in app.css.
- **J7 draft navigation pinned**: new shell test — drafts never
  cross-contaminate between threads (A/B switching, restore-on-return).
- **J8 honest empty states**: history empty is FILTER-AWARE
  (`history.emptyFiltered` en/de + conditional in history.templ — an
  active filter with no rows no longer claims "No calls recorded
  yet"); `vm.empty` disambiguated from SMS ("No voicemail yet." /
  "Noch keine Mailbox-Nachrichten.").
- Tests: SSE test updated (recovery toast now part of contract) + new
  recovery-matrix test; connection test asserts the assertive
  announce; shell retry/dismiss test; Go
  `TestHistoryEmptyStateIsFilterAware`. **152/152 island, full Go
  green, templ + nix fmt clean.**
- Docs: error-contract rows (SSE recovery, send-transport failure with
  retry/dismiss); runbook obligation → T15; AGENTS.md gained the
  Feedback/trust bullet (had to re-read first — the daemon CONDENSED
  AGENTS.md mid-session; adapted insert to the new style).

### T16 — M15 tokens + M20 theming + M26 sizing (COMPLETE, tests green)

- **Token audit**: hardcoded on-accent colors (`#fff` ×4 app.css,
  `#f4fbf9`/`#fff` island), overlay scrim literal, two stale var()
  fallbacks (`#9ca3af`, `#22c55e`) that disagreed with real values.
- **New tokens** `--on-accent` + `--scrim` (theme-stable → dark :root
  blocks only, inheriting into light exactly like `--radius*` — this
  is the established pattern, verified against how radius/hairline
  already work). All hardcoded sites replaced; fallbacks cleaned.
- **M20 data-theme propagation ANALYZED + PINNED**: island/style.css
  has NO `[data-theme]` selectors — manual theming reaches island
  surfaces purely via app.css's higher-specificity override blocks on
  the same `:root`. Works ONLY while token sets stay aligned →
- **`TestCSSTokenBlocksAreMirrored`** (new
  `internal/web/assets/tokens_test.go`): parses all five token blocks
  (app dark / light-auto / light-manual, island dark / light), asserts
  app-internal light consistency, shared-name value equality across
  files, and theme-dependent core presence everywhere. First run
  caught `--danger-soft` (island never consumes it) → moved to
  shared-equality-only coverage with rationale comment.
- **M26 touch targets**: existing 44px `.wp-mini` rule was
  width-gated only — tablets (coarse pointer, wide viewport) missed
  it. Added `@media (pointer: coarse)` 44px rule. Fluid heading:
  `.wp-panel-head h2` now `clamp(1.15rem, 0.95rem + 0.8vw, 1.3rem)`
  (root stays a percentage — user font preference still wins).
- **Reduced-motion audit (M20)**: app.css ships a GLOBAL kill switch
  (`* { animation/transition-duration 0.01ms !important; scroll-behavior auto }`) — every new animation (wp-key-sent, morph,
  skeleton, player) is covered by construction. Island has 3 targeted
  blocks. No changes needed; recorded here as the audit result.

## b) PARTIALLY DONE

### T17 — M18 onboarding + M19 mobile (ANALYSIS ONLY, zero edits)

- **M18**: welcome panel already carries full en/de onboarding copy
  (title/body/3 points/sign-in hint, per-session language). Chosen
  design (not yet built): server-rendered dismiss button +
  `.wp-welcome-compact` mode via localStorage, wired in shell.js —
  keeps copy server-side (session language native), dismissal
  client-remembered.
- **M19**: bottom tab bar, safe-area insets (`env(safe-area-inset-bottom)`),
  44px nav links ALREADY exist from earlier trains. Remaining:
  composer sticky-footer/iOS-keyboard behavior + viewport-meta
  investigation (`interactive-widget=resizes-content` — the meta comes
  from templ-components `layout.Base`; local grep found no local
  viewport override; need to check the library's Base props next).

## c) NOT STARTED (train remainder)

T18 (SEAM design note FIRST; M21 settings + M22 store ALTER — fires
T25.1), T19 (M24 i18n/RTL + M25 call depth), T21 (nix-review batch 2),
T22 (samber/do + dashboard follow-ups), T23 (visual harness — research
the stack's browser-e2e.py), T25 (if not landed with T18), docs-health
HARVEST, final gates (buildflow full, quiet-host flake check,
fresh-binary smoke with T14/T15 eyeballs, `git ls-remote` end-state).

## d) TOTALLY FUCKED UP (nothing broken — two self-caught slips)

1. **templ scriptlet mistake**: put Go statements (`emptyKey := …`)
   inside an element body without `{{ }}` — templ wrote them as TEXT
   and generated Go referenced an undefined var. Caught by immediate
   `go test`; fixed with a `{{ }}` block. Lesson: AGENTS's templ rules
   cover `if` blocks; plain statements ALWAYS need scriptlets.
2. **One no-op edit** (identical old/new intended as the scrim swap) —
   exactly the whitespace-only-edit class session 4 warned about.
   Harmless (tool no-op'd it), redone properly. Slower checking next
   time.
3. Minor: the i18n.go edit whitespace-normalized by the tool —
   verified alignment matched file style after (tabs preserved).

## e) WHAT WE SHOULD IMPROVE

- **Run the oxlint island gate after JS edits in the SAME batch** —
  this session edited shell.js/ui.js/session.js/connection.js and only
  re-ran tests + fmt; oxlint was NOT re-invoked (low risk — no new
  browser globals — but the habit should be: tests + fmt + lint
  together). Queued as next-item #1.
- **The daemon condensed AGENTS.md mid-session** (file changed between
  my read and edit — caught by the mtime guard). The condensed version
  dropped detail the old one had (e.g. the long honesty-contract
  wording). Whoever owns that condense: verify nothing load-bearing
  was lost (the DOM-contract/idempotency rules especially).
- **Remote lag**: local has commits the daemon hasn't pushed yet
  (`ffaa03f` remote vs `f3b87d1`+ local, plus uncommitted session
  changes at report time). End-of-train `git ls-remote` assert is
  already planned — do NOT trust push logs mid-train.
- Session-4's open owner questions (C4 ratify, T18↔T25 sequencing,
  waveform double-fetch) remain unanswered — defaults stand.

## f) Next up to ~50 (execution order)

1. `nix develop -c oxlint …` island gate re-run (post-T15/T16 JS
   edits) + island suite re-run after latest fmt.
2. T17 M18: implement welcome dismissal (dismiss button i18n en/de +
   compact-mode CSS + shell.js localStorage wiring + tests).
3. T17 M18 tests: dismissal persistence (localStorage), compact
   reveal, both languages render.
4. T17 M19: check templ-components Base viewport props — add
   `interactive-widget=resizes-content` if the library allows
   override WITHOUT forking Base (else document the library limit).
5. T17 M19: composer sticky-footer audit on iOS keyboard
   (visualViewport behavior; keep it CSS-first).
6. T17: extend runbook E2E obligation note to T17 markup deltas
   (welcome dismissal button IS served markup).
7. T17: templ generate + nix fmt + full suites; T17 done.
8. **T18 design note FIRST** (docs/planning/): enumerate M21 settings
   depth + M22 store ALTER + schema_version mechanism + server carve
   trigger — before ANY code.
9. T18↔T25 default stands (same train) unless owner rules otherwise.
10. T18: implement M21 settings surface per note.
11. T18: implement M22 store ALTER + migration test (old row → new
    schema) + schema_version table + runner.
12. T18: store family pins + server handler tests for the new
    surface.
13. T18: server carve decision if file-count trigger fires
    (internal/server split per the note's criteria).
14. T19 M24: island i18n dictionary audit — extend the en/de sync
    test to flag UNUSED keys too.
15. T19 M24: RTL smoke — render a page with dir=rtl, audit logical
    CSS properties (margin-inline etc.), fix what's broken.
16. T19 M24: Go render tests for RTL strings (no truncation
    regressions).
17. T19 M25: transfer/merge affordances in the island call card —
    READ the DTMF/REFER lessons first (FreeSWITCH executes transfers
    server-side; REFER semantics already pinned).
18. T19 M25: hold-state parity for transfers (holdPending machine
    interplay).
19. T19: island call-card specs for the transfer state machine.
20. T21 nix-review batch 2: re-run the nix-review skill over
    flake.nix + nix/*.nix (module split landed batch 1).
21. T21: verify island-lint file list in checks.nix grew nothing
    (player lives in shell.js — confirm).
22. T21: fix findings; `nix flake check` after.
23. T22: re-check AGENTS composition-root bullets against
    internal/app for drift (the condensed AGENTS makes this MORE
    relevant now).
24. T22: dashboard refresh-loop test still green; do-lifecycle
    conformance asserts.
25. T22: dashboard CSS rebuild ONLY if sources changed
    (tailwindcss_4, never v3).
26. T23: research the stack's browser-e2e.py (in the
    nix-international-telephony repo on this host) — what driver does
    it reuse; pick the same for webphone's harness.
27. T23: decide harness budget (screenshot count, runtime) BEFORE
    building; KVM-gate if it needs a browser.
28. T23: golden-page screenshots for tab surfaces incl. voicemail
    player + fax timeline; wire into scripts/ + flake check.
29. T25: if T18 landed schema_version — carve decision only; else
    full T25.
30. HARVEST: fold session-3 §f + session-4 §f + this §f into
    TODO_LIST.md.
31. HARVEST: T11–T15 rows → CHANGELOG Unreleased + FEATURES.md status
    flips (voicemail player, fax resend/timeline, retry/dismiss, SSE
    recovery, assertive incoming, filter-aware empty, token mirror,
    touch targets).
32. HARVEST: mark the master plan's T11–T16 rows DONE with pointers.
33. Final gates: `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`
    (full), fix findings (gomod-check vendor false-positive is the
    documented exception).
34. Final gates: quiet-host `nix flake check` (KVM backup VM included).
35. Final gates: fresh-binary smoke — boot loopback, 41+4 checks,
    eyeball fax timeline/resend + voicemail player + history
    filter-aware empty + failed-send Retry/Dismiss in ONE boot.
36. `git ls-remote origin main` == local HEAD end-state assert.
37. AGENTS.md durable-rule lines if T17+ added any (welcome-dismissal
    seam, viewport note) — one line each, rules only.
38. Owner handover notes for T01/T02/T04/T09/T24/T27 (standing).
39. Stack browser-E2E run is OWED at the next release tag (T11–T17
    obligation note) — do not let the runbook note rot.
40. Consider extending the smoke script with a
    `history?outcome=missed` empty-filter probe (cheap, pins J8
    server-side) — only if trivially greppable.

## g) Questions for the owner (cannot figure these out myself)

1. **M18 onboarding shape**: my plan turns the existing welcome TAB
   into a dismissible compact mode (localStorage). Alternative: a
   one-time hint layer right after the FIRST successful login
   (wp:session-opened) pointing at the tabs. The first is calmer; the
   second is louder exactly once. Which do you want? (Default if
   silent: welcome-tab dismissal.)
2. **Welcome-dismissal persistence**: localStorage is per-browser. A
   server-side "onboarding seen" flag would need the settings store —
   natural to piggyback on T18's M21/M22 work. Defer to T18, or is
   per-browser fine forever? (Default: per-browser.)
3. **Carried from session 4, still unanswered**: T18↔T25 sequencing —
   land schema_version in the SAME train as M22's store ALTER (my
   default) or defer T25 until after T19? And please ratify the C4
   fax-thumbnail decline + the voicemail waveform double-fetch
   acceptance while you're at it.

— End of report. Waiting for instructions.
