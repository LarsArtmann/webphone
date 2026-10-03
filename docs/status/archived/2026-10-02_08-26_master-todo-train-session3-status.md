# Master TODO Train — Session 3 Status

**When:** 2026-10-02 08:26 CEST (session started 2026-10-01 ~23:10, overnight run)
**Scope:** the master plan's remaining assistant legs — T11 (perf extras), T12 (M9+M12), T13 (M11+M16, ~80%), plus the owed memory lines. T14–T25, HARVEST, and the final full gates NOT reached.
**Inputs:** master plan `docs/planning/2026-10-01_17-32_SUPERB-master-todo-pareto-execution-plan.md`, domain plans (ui-ux `03-53`, verification/perf `05-35`), session-2 report `2026-10-01_23-11`.
**Standing order:** "NOW GET SHIT DONE! The WHOLE TODO LIST!" — Verschlimmbessern guard honored throughout (verbatim island serving, strict CSP zero-inline, no-store pins, DOM contract, /events uncompressed, no assistant ssh).

> ARCHIVED 2026-10-03 (docs-health v6 sweep): session 3's trains (T11 perf
> extras, T12 dial affordances + countdown, T13 filters/URL state) stood; the
> remainder (T14–T19, T21–T23, T25.1) executed by sessions 4–8; T13's tail
> closed green (E3 pinned); owner legs stay on their rows. Per-item verdicts
> inline.

---

## a) FULLY DONE (implemented AND verified this session)

~~1. **Owed memory lines (session-2 (f) #42/#43).**~~ done — this session (report of record)
   - `AGENTS.md` Formatting bullet now ends with the rule: after ANY
     island/shell/css edit run `nix fmt` BEFORE the gates (buildflow's
     formatter skips island files; drift has failed full gates).
   - `docs/lessons.md` Tooling traps: the "always-pending stub promise
     hangs node:test" gotcha + the `lastToast()`-is-an-ordering-claim
     companion trap.
~~2. **T11.1 — modulepreload for the island ESM graph.**~~ done — this session (report of record)
   - `assets.IslandModules()`: sorted serving URLs of every
     `island/app/*.js` (computed once from the embed; now 18 modules
     after `pcsetup.js` joined).
   - `TestIslandModulesMatchImportClosure` (assets package): walks
     main.js's static import closure (named + side-effect imports, no
     dynamic imports) and pins closure == listing, so "preload the
     directory" can never silently drift from the real graph.
   - `layout.templ headExtras`: one `<link rel="modulepreload">` per
     module + a `<link rel="preload" as="script">` for the vendored
     sip.min.js (the one big classic script, discovered just as late
     without the hint). `templ generate` run.
   - `TestServedPagePreloadsIslandGraph`: per-module link present,
     count == graph size, all inside `<head>`, sip preload present.
~~3. **T11.2 — outgoing-call mic warm (the incoming mirror).**~~ done — this session (report of record)
   - `mic.js warmMic(ttlMs = 0)`: bounded-expiry warm — the dial-focus
     warm carries a TTL; the timer is identity-guarded (an expired timer
     can never kill a LATER warm — load-bearing when a warm device dies
     on its own mid-TTL) and cleared on take/release.
   - `calls.js initDialWarm()`: focus on `#dest` starts the acquisition
     with `DIAL_WARM_TTL_MS = 45_000`. **Deliberate deviation from the
     plan's µ25:** session-open is NOT a trigger — the mic indicator
     lighting right after login expresses no call intent (privacy).
     Documented in the code comment.
   - Tests: 3 TTL-semantics tests in `mic.test.mjs` (expire+release,
     consumed-before-expiry survives, stale-timer identity guard) + the
     dial-focus wiring test in `calls.test.mjs` (one acquisition,
     idempotent while pending).
~~4. **T11.3 — ICE panel setup timings (gathering duration + time-to-first-media).**~~ done — this session (report of record)
   - NEW leaf module `island/app/pcsetup.js`: `instrumentSessionDescriptionHandler`
     attaches `icegatheringstatechange`/`iceconnectionstatechange`
     listeners at SDH construction; `setupSummary(pc)` reports
     `gatherMs`, `gatherCapped` (stale gathering past the 1000 ms cap +
     200 ms slack), `gatheringStartedAt`, `iceConnectedAt`.
   - `connection.js`: the SDH factory is wrapped (eager base factory
     call PRESERVED — the existing eager-capture contract test depends
     on it; first attempt made it lazy and was restructured), each
     fresh handler's PC is instrumented before the first offer/answer.
   - `ice.js`: `firstMediaByPC` WeakMap + `setupLine` — the panel now
     renders `setup: gather 812 ms · ice +0.6 s · first media +1.4 s`
     (all measured from gathering start; on the answering side that is
     effectively accept→speak).
   - Tests: NEW `pcsetup.test.mjs` (5 tests; the capped branch uses
     node:test mock timers `t.mock.timers.enable({apis:["Date"]})`) and
     NEW `ice.test.mjs` (2 panel-render tests: full setup story once
     media flows; uninstrumented call shows live lines but no setup
     line). **Real bug found by the tests:** `marks.x && …` falsiness
     vs `!= null` — with mocked epoch 0 (and any future 0 value) the
     gather duration silently read as null; fixed to explicit null
     checks in pcsetup.js.
~~5. **T11.4 — iceServers trimming evaluation (note, no change).**~~ done — this session (report of record)
   - `docs/planning/2026-10-01_23-50_iceservers-trimming-evaluation.md`:
     the list lives in the STACK config; STUN-trim is the only
     candidate, TURN is off the table; the decision gate table keys on
     the panel's new `setup:`/`path:` numbers from the owner's T01
     ritual; standing revisit trigger documented.
~~6. **T10.6 — timing baseline (Python, curl is banned in the harness).**~~ done — this session (report of record)
   - `scripts/perf-baseline.py` (executable): urllib + perf_counter;
     measures cold GET / gzip GET / If-None-Match revalidation with a
     gzip-roundtrip assertion and a served-page preload count. (First
     run failed: urllib raises on 304 — the script now treats
     HTTPError-304 as the response it wants.)
   - Measured against a fresh binary: `/assets/vendor/sip.min.js`
     cold 273,356 B / 0.8 ms; gzip 62,942 B (77% saving) / 3.8 ms;
     304 revalidation 0 B / 0.4 ms; head carries 18 modulepreload
     links. Numbers + interpretation appended to the verification
     plan's verdict appendix (`2026-10-01_05-35`, table + notes).
~~7. **T12 — M9 dial affordances + M12 over-limit countdown** (after the~~ done — this session (report of record)
   12.1 audit; see d)4 for the audit's "already exists" findings):
   - **A5 dial hint**: `calls.js sanitizeDialable()` (ONE home for the
     dialable charset — placeCall and the hint share it) +
     `initDialHint()` → a `→ +4930123456` preview under `#dest` while
     typing (formatted numbers reveal their dialable form BEFORE the
     Call button commits). Served `#wp-dial-hint` in phone.templ
     (aria-hidden visual aid), `els.dialHint` in ui.js, wired in
     main.js, `.wp-dial-hint` CSS, DOM contract id added. 2 tests.
   - **A4 typeahead personal contacts**: `typeahead.js
     setExtraContacts()` + `suggestionSource()` — personal
     (per-extension server store) contacts join the ranking and WIN
     number collisions (shared duplicate deduped away; ranking order
     unchanged); boot guard now allows personal-only deployments.
     `panels.js renderContacts` syncs the source on EVERY contacts
     render (load, SSE nudge, save, remove — one call site, no drift)
     and re-attempts the typeahead boot. 2 tests.
   - **A8 DTMF feedback**: `sendDtmf` now pulses the pressed key
     (`wp-key-sent` class, animation-restart on rapid repeats; fires on
     INVOCATION so keyboard-shortcut tones flash too) and collects a
     transient `#wp-tones` trail (last ≤12 tones, dot-separated, 3 s
     fade-to-hidden; #log stays the durable record). CSS keyframe +
     prefers-reduced-motion guard. 1 test.
   - **M12.3 byte-limit countdown**: `shell.js updateSegcount` extended
     — `MAX_BODY_BYTES = 1600` mirrors `messaging.MaxBodyLength` (Go
     len = BYTES, so the client counts TextEncoder bytes, not chars);
     `LIMIT_WINDOW_BYTES = 200`; near the cap the counter appends
     "bytes/1600", over the cap it goes loud (`.wp-segcount-over`,
     danger color). **Design choice:** the counter WARNS but never
     blocks — the server's 422 banner stays the enforcing verdict.
     `.wp-segcount-over` CSS in app.css. 2 tests.
   - **K5 diagnostics gear**: `phone.templ` wraps the ICE + event-log
     details in `#wp-advanced` behind `#wp-adv-toggle` (⚙). **Ships
     OPEN** (aria-expanded=true) — the stack E2E and the operator
     runbook must never meet a hidden diagnostics tree; collapse is
     opt-in and persisted (`wp-advanced-collapsed`, ui.js
     `initAdvancedToggle` mirrors aria-expanded). i18n
     `advancedSection` en/de; `.wp-advanced-bar` CSS; DOM contract ids.
     2 tests.
   - T12 checkpoint: island suite **140/140**, `go test` views+server
     green, `nix fmt` run.
~~8. **T13 server side — M11 history filters (D8/D9/D10) + E2:**~~ done — this session (report of record)
   - `panels.go`: `outcome` query param ("missed" = inbound +
     billsec 0; "answered" = billsec > 0; outbound legs can never be
     "missed" by this definition), `filterCDRs` gained the 4th param,
     fetch window widens for it.
   - `history.templ`: outcome select; **E2** `hx-push-url="true"` on
     the filter form (filter state is URL-addressable/shareable);
     **D9** day grouping — `groupCDRsByDay` buckets by the `Start`
     10-byte prefix ("YYYY-MM-DD HH:MM" contract; unparsable/short
     starts render headless rather than under a wrong label) and the
     rows render under the SAME `dayHead` vocabulary the message
     transcript established (Today/Yesterday/weekday+date + count);
     **D10 partial** missed styling — `cdrMissed()`, ✖ glyph,
     `wp-row-missed`/`wp-dir-missed` classes (danger).
   - i18n keys en/de (`history.outcome.{label,any,answered,missed}`,
     `history.missed`); `.wp-dir-chip.wp-dir-missed` +
     `.wp-row.wp-row-missed` CSS in app.css.
   - `TestHistoryOutcomeFilterAndDayGroups`: missed/answered
     filtering, select echo, two day heads in order, missed styling +
     glyph, hx-push-url pin. GREEN. Views + server suites green.
~~9. **T13 client side — E3 + E8 (implemented, test WIP, see b)):**~~ done — this session (report of record)
   - `shell.js`: §1a stores `wp-last-tab` on every tab navigation;
     §1b `restoreLastTab` IIFE at boot — only on a plain `/` load,
     only when the stored tab is one of the RENDERED nav links
     (garbage can never 404 the boot), never over the sign-in hint
     (`.wp-welcome` guard), fetches partial + nav morph and
     `replaceState`s the address to match the content (all guards
     defensive: `window.location`/`window.history` may be absent).
     E8: `refreshNav` scrolls the active link into view with
     `scrollIntoView({block:"nearest", inline:"nearest"})` after morph
     re-renders (only scrolls when genuinely out of the strip's view).
   - island `main.js restoreTabsAfterSignIn` now honors the stored tab
     (validated against the six known tabs) instead of always
     "messages".
   - `helpers.mjs`: `location.pathname` + `window.location` reference
     added to the stub world (the empty `window` object made the first
     restore implementation throw at load — the shell's load-error
     boundary surfaced it as a whole-file test failure).

---

## b) PARTIALLY DONE

~~1. **T13 — one test from green.** The three new E3 shell specs~~ done — T13 finished green in the follow-on sessions (E3 restore landed + pinned: shell.test wp-last-tab)
   (`shell.test.mjs`) are written; two pass ("tab navigation stores the
   last-active tab", "restore never overrides a deep link or the
   sign-in hint"). The third — "a plain / load restores the remembered
   tab" — still fails: the restore path runs but no `htmx.ajax` calls
   are recorded. Standalone debug repro was started, exceeded the 60 s
   backgrounding threshold (undecided whether a real hang or harness
   slowness — the `?case=` fresh-import of shell.js in a bare node
   process), and the shell was killed when this report was requested.
   Remaining T13 tail after that test: an E8 scrollIntoView assertion
   (currently only implemented, not pinned), `nix fmt` for the
   shell.js/helpers/main.js edits, full island + go suite re-run, and
   the served-markup stack-E2E obligation note (T04 owner-blocked).
~~2. **The whole train after T13** — see c).~~ done — the whole train after T13 executed (T14–T19, T21–T23, T25-partial; see the c-verdicts)

---

## c) NOT STARTED

~~- **T14** — M13 voicemail playback (C1–C3/C9/C10) + M14 fax depth (C4–C6).~~ done — T14 shipped (voicemail player + fax timeline/resend)
~~- **T15** — M17 feedback/trust: J2 reconnect banner (via `#wp-live`),~~ done — T15 shipped (offline banner + SSE recovery announce + failed-bubble retry)
  J3 undo, J4 retry-in-banner, J7 confirm consistency, J8 spinner,
  J9 success pulse.
~~- **T16** — M15 visual tokens + M20 theming depth + M26 shell sizing.~~ done — T16 shipped (mirrored tokens, theme system, shell sizing)
~~- **T17** — M18 onboarding/demo + M19 mobile extras (I3/I6/I7/I9).~~ done — T17 shipped (onboarding + mobile legs)
~~- **T18** — M21 messaging richness + M22 pin/archive/mute (SEAM:~~ done — T18 shipped (snippets + thread flags; the design pass ran FIRST)
  server design pass FIRST; M22 would be the first store schema change
  → couples to T25.1).
~~- **T19** — M24 i18n locale/RTL/status dots + M25 call depth (A6/A10).~~ done in part — T19 shipped (RTL groundwork + focus mode); status dots weren't taken
~~- **T21** — nix-review batch 2 (module golden, tag guard, VM/drill,~~ done — golden + tag guard + post-split VM; actionlint/dedupe/exception rows not adopted
  actionlint, devshell dedupe, exceptions).
~~- **T22** — samber/do + dashboard follow-ups (health.css commit+CI,~~ done — health.css script + canary, family pins, divergence documented; the /health policy + v2.9.0 fold ride the TODO health row (owner)
  family pins, divergence, `/health` policy, aarch64+vulnix, v2.9.0
  fold decision).
~~- **T23** — visual verification harness (`scripts/ui-capture.py`,~~ done — scripts/ui-capture.py + the 14-shot matrix + the AGENTS note
  12-shot matrix, AGENTS note, vision-CLI decision).
~~- **T25** — schema_version (trigger-gated; no new server files were~~ resolved by events — schema_version shipped with T18; the carve trigger stayed unfired; the gateway bits ride the cross-repo row
  added this session — panels/history/i18n changes all stayed in
  EXISTING files, so the carve trigger stayed unfired) + server carve
  decision + gateway stack-side bits.
~~- **Docs-health HARVEST** — fold session-2 (f) + this report's (f)~~ done — the rows moved + this v6 sweep harvested
  into `TODO_LIST.md`; move shipped T05/T06/T07/T08/T10/T11/T12 (+T13
  when green) rows to CHANGELOG/FEATURES.
~~- **Final gates** — `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`~~ resolved by events — the later full batteries are green
  full + `nix flake check` + fresh-binary smoke + `git ls-remote`.
~~- Owner legs (handover-only): T01, T02, T04, T09, T24, T27.~~ routed — TODO owner rows (handover legs)

---

## d) TOTALLY FUCKED UP (and how)

~~1. **Python heredoc write discipline — TWICE.** Scripts did~~ process record
   `s = s.replace(old, new)` and then only
   `open(p, "a").write(addition)` — the modified `s` was NEVER
   written; the append landed on the unmodified file. Symptom: missing
   imports that looked like module bugs. **Lesson:** one
   read-modify-write per script, and verify the marker actually changed
   (print a grep after writing).
~~2. **templ comments inside an attribute list are a parse error.**~~ process record
   history.templ's `<form … // comment …>` broke `templ generate`
   ("if: expected nodes"). Comments must live OUTSIDE the tag; moved
   above the form.
~~3. **Three failed anchored edits from formatter drift.** The~~ process record
   daemon/templ tooling reindented `.templ` files BETWEEN my read and
   my edit (tab counts changed), so byte-exact anchors stopped
   matching mid-script (scripts assert-before-write, so nothing
   corrupted — but three re-anchoring rounds burned time).
   **Lesson:** re-dump the exact bytes (`cat -A`) immediately before
   every templ edit; prefer content anchors over remembered
   indentation.
~~4. **T12 audit finding worth recording:** A9 (re-dial) ALREADY EXISTS~~ process record
   both sides (island history rows have ↻ buttons; CDR rows have
   data-dial) — I almost re-built it. The 12.1 "already exists?"
   audit caught it before any code. Also: E7 (breadcrumb) assessed as
   ALREADY SERVED by the ThreadView back-link (one level of depth
   needs no trail) — recorded here rather than in code.
~~5. **A run of self-inflicted test bugs** (all caught by the run, none~~ process record
   shipped): assumed dedupe-preference implies ranking-preference
   (Anna Licht legitimately outranks Anna Privat); a "unique" query
   that also matched a shared contact ("zahn" → Zahnpasta Hotline);
   expected a hidden counter at 1400 chars (any body that long is
   multi-segment by construction); passed `event.target.value` to a
   listener that reads `els.dest.value`; used `dataset.tab` on a stub
   whose dataset is disconnected from `getAttribute`; forgot the stub
   `document.querySelector` is a null-sink (had to route two
   selectors around it); `nav.querySelectorAll` needs manual wiring on
   stub elements.
~~6. **Wrong node:test mock-timer API** (`t.mock.timer()` →~~ process record
   `t.mock.timers.enable({apis:["Date"]})`) — one run to notice, but
   the mock timers then EXPOSED a real falsiness bug (see a)4).
~~7. **The first factory wrap broke an existing contract test.** Moving~~ process record
   the base SDH-factory call inside the per-session arrow made the
   eager capture (`capturedMediaStreamFactory`) never run —
   restructured to eager base + thin per-session wrap.
~~8. **Left a background debug shell running** (11B, the restore-repro~~ process record
   hang investigation) until the report request killed it. Should have
   terminated or diagnosed it before starting the report.

---

## e) WHAT WE SHOULD IMPROVE

~~- **Single-write python scripts** for edits (see d)1) — and print a~~ process record
  post-write grep so a silent no-op can't hide.
~~- **A `templ` edit protocol:** dump exact bytes immediately before the~~ process record — the templ protocol is AGENTS/lessons material
  edit; place comments outside tags; run `templ generate` after EVERY
  templ edit batch, not at checkpoints.
~~- **Test assertions about ordering/dedupe should be derived from the~~ process record
  implementation's sort/dedupe code, not from intuition** — every
  failed expectation this session was an intuition bug.
~~- **The `?case=` fresh-import pattern** re-registers ALL~~ process record
  document-level listeners per import (acceptable while such tests sit
  at the file end). If restore-style boot code grows, extract it into
  a named, individually-importable function instead.
~~- **Kill or conclude background debug shells before switching tasks.**~~ process record

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered by the plan's tier order; [owner] = handover-only.

**T13 finish (immediate)**

~~1. Diagnose the restore-test failure (why restoreLastTab records no~~ done — the restore landed + pinned (shell.test.mjs:728; main.js:207)
   htmx.ajax calls under the stub world) and land it green — or split
   the restore into a testable named function if the IIFE-at-import
   shape keeps fighting the harness.
~~2. Add the E8 pin: refreshNav scrolls the active link into view~~ not adopted — below the bar; the E8 scroll ships unpinned
   (stub `scrollIntoView`, assert called with `nearest`/`nearest`).
~~3. `nix fmt` (shell.js, helpers.mjs, main.js edited since the last~~ done — routine in the follow-on batteries (nix fmt + suites green)
   fmt) + full island suite + `go test ./...`.
~~4. Record the T11/T12/T13 served-markup delta in the stack-E2E~~ routed — TODO island-honesty row (T04 owner-blocked)
   obligation note (T04 [owner] still blocked).

**T14–T19 (the UI/UX train, in plan order)**
~~5. T14.1 — M13 C1–C3: voicemail waveform from audio peaks + scrubber~~ done in part — T14 player (speed + scrubber); the waveform wasn't taken

- speed control.

~~6. T14.2 — M13 C9/C10: playing-row highlight (MORPH id) + unread~~ done in part — the player shipped; the row-highlight/unread legs rode the visual pass
   clears on play mirroring the nav badge.
~~7. T14.3 — M14 C4–C6: fax first-page thumbnail + status timeline +~~ done in part — timeline + resend shipped (T14); the thumbnail wasn't taken
   resend on failed rows.
~~8. T15.1 — J2 reconnect banner (completes the honesty-contract story;~~ done — the honesty-train banner + T15's SSE recovery announce
   rides `#wp-live` + the SSE reconnect surface).
~~9. T15.2 — J3 undo + J4 retry-in-banner.~~ done in part — the retry surface is the failed bubble's Retry/Dismiss; undo wasn't taken
~~10. T15.3 — J7 confirm consistency + J8 spinner + J9 success pulse.~~ done in part — spinner/pulse weren't taken
~~11. T16.1 — M15 F3/F4/F7/F9 visual tokens (icon set, emoji→icons,~~ done in part — the mirrored tokens shipped; the icon set wasn't refreshed
    state colors, theme toggle icon).
~~12. T16.2 — M20 theming depth (F5/F6/F8/F10).~~ done in part — the theme system shipped; depth extras weren't taken
~~13. T16.3 — M26 shell sizing E5/E6 (island collapse, resizable~~ done in part — shell sizing shipped; the sidebar resize wasn't taken
    sidebar).
~~14. T17.1 — M18 K1–K4 onboarding/demo.~~ done — T17 onboarding
~~15. T17.2 — M19 I3/I6/I7/I9 mobile extras.~~ done in part — T17 mobile; the I3/I6/I7/I9 extras weren't all taken
~~16. T18.1 — server design note: M21 snippets/schedule (SEAM).~~ done — the T18 seam design (D1–D14)
~~17. T18.2 — server design note: M22 pin/archive/mute (SEAM).~~ done — same design doc
~~18. T18.3 — implement per design (+ tests); M22's store change is the~~ done — T18 implementation + tests; the first ALTER fired T25.1 (versioned migrations)
    FIRST ALTER → fires T25.1 (see Q3).
~~19. T19.1 — M24 locale switch UI + RTL audit + status dots.~~ done in part — RTL groundwork shipped; the locale-switch UI/status dots weren't taken
~~20. T19.2 — M25 A6 focus mode + A10 media test.~~ done — T19 focus mode + device self-test

**Infra / nix / tooling**
~~21. T21.1 — golden module-output fixture (vhost + backup scripts) + check entry.~~ done — nix/module-output.golden + the module-check golden case
~~22. T21.2 — release.sh `webphoneVersion`↔`git describe` guard (dry-run).~~ done — release.sh tag↔webphoneVersion guard (override documented)
~~23. T21.3 — KVM backup VM + drill runs post-split.~~ done — post-split VM ran green (the 20:12 battery)
~~24. T21.4 — `actionlint` over `.github/workflows/ci.yml`.~~ not adopted — actionlint never run
~~25. T21.5 — dedupe `devShells.ci`/`default` Go env.~~ not adopted — the shells stay separate
~~26. T21.6 — record accepted exceptions + declined `go-standard`.~~ not adopted — the exceptions live in AGENTS only
~~27. T22.1 — commit `health.css` + regeneration script (dark recipe).~~ done — scripts/build-health-css.sh (TODO health row)
~~28. T22.2 — wire CI rebuild check for `health.css`.~~ done — checks.health-css canary (TODO health row)
~~29. T22.3 — `family_test.go` pins for any new error codes this train added.~~ done — T22 family pins (store.thread_flag, store.count_archived, snippets)
~~30. T22.4 — local-behind-remote divergence check (daemon pushes; verify `git ls-remote`).~~ done — the divergence is documented LIVE daemon behavior (TODO row)
~~31. T22.5 — aarch64 cross-build ELF verify + `nix run .#vulnix`.~~ done — T22 close-out (aarch64 + vulnix zero-real-advisories)
~~32. T22.6 — v2.9.0 fold decision note.~~ routed — TODO health row (the v2.9.0 fold decision, owner)
~~33. T23.1 — persist `scripts/ui-capture.py`.~~ done — scripts/ui-capture.py (AGENTS T23 harness)
~~34. T23.2 — finish the 12-shot matrix.~~ done — the matrix grew to 14 shots
~~35. T23.3 — review + fix what the shots surface.~~ done — the 2026-10-02 rebuild fixed a genuinely stale artifact (TODO health row)
~~36. T23.4 — AGENTS "visual gate" note.~~ done — the AGENTS Commands block documents the visual harness
~~37. T23.5 — vision-CLI provider/key decision [owner input].~~ owner — the vision-CLI decision wasn't taken (the harness is DOM-assert based, LOCAL-ONLY)
~~38. T25.1 — `schema_version` table implementation if T18 fires the ALTER trigger.~~ done — shipped with T18 (versioned migrations, schema_version v2)
~~39. T25.2 — `internal/server` carve decision note (trigger stayed unfired this session — all work landed in existing files).~~ record stands — the carve trigger stayed unfired (all work landed in existing files)
~~40. T25.3 — gateway stack-side bits (ftypqt sniff, MMS-outbound, pbx FEATURES row) — preparation notes; the wire work is XREPO.~~ routed — TODO cross-repo row (the gateway stack-side bits)

**Docs / hygiene / gates**
~~41. Docs-health HARVEST: fold session-2 (f) + this (f) into `TODO_LIST.md`.~~ done — the one-home moves + this v6 sweep
~~42. Move shipped rows (T05/T06/T07/T08/T10/T11/T12, T13 when green) to CHANGELOG/FEATURES.~~ done — the shipped rows moved to CHANGELOG/FEATURES
~~43. Record the A9-already-exists + E7-declined-as-served audit outcomes in the plan/HARVEST so nobody re-litigates them.~~ done — the A9-already-exists / E7-served outcomes are recorded (this report d4 + the plan docs)
~~44. Session-2 stragglers: vendor-freshness scratch-tree probe; go-line-flipflop disposition; /mnt/buildcache housekeeping; assertToastPresent helper decision; lychee archived-file recheck.~~ not adopted — the stragglers were dispositioned below-the-bar in the earlier verdicts
~~45. Monthly erraudit tier-1+2 re-measure is calendared 2026-10-22 (not due).~~ routed — standing watch (TODO watches row; next due 2026-10-22)
~~46. Final gates: full `./scripts/buildflow.sh` + `nix flake check` + fresh-binary smoke (`--expect-version` semantics unchanged) + `git ls-remote`.~~ resolved by events — the later full batteries are green
~~47. [owner] T04 — stack browser E2E re-run; the T11/T12/T13 markup deltas are stacking behind it.~~ routed — TODO cross-repo row (T04, owner terminal)
~~48. [owner] T01/T02/T09/T24/T27 — handover legs (deploy tail, SMS bridge, owner-calls sitting, cross-repo, docs-health continuation).~~ routed — TODO owner rows (the handover legs)
~~49. [owner] Ratify the `-t 3` dedup baseline + the release-load E2E-retry choice (standing).~~ routed — TODO OWNER-calls row (the -t 3 baseline pending ratification)
~~50. [owner] The lychee 502 on `https://pbx.artmann.tech/` disposition (accepted external condition vs fenced) — still unanswered from session 2.~~ process record — the 502 was the PBX down; lychee green at the release tail

---

## g) Questions I CANNOT answer myself (max 3)

~~1. **E3 landing behavior.** `restoreTabsAfterSignIn` now lands a FRESH~~ record stands — continuity shipped (fresh login lands on the last-used tab); no owner override
   login on the last-used tab (was: always Messages). I chose
   continuity; you may prefer a clean Messages landing after login
   (restore then only applying to reloads). Which do you want?
~~2. **K5 default state.** The diagnostics gear ships OPEN (collapse is~~ record stands — open-by-default confirmed by shipping (E2E/runbook safety)
   opt-in, persisted) to keep the stack E2E and the operator runbook
   grep safe — the catalogue's "progressive disclosure" implied
   default-collapsed. Confirm open-by-default is the wanted posture
   (flipping it later is trivial; flipping it blind before T04's E2E
   re-run is the risk I avoided).
~~3. **T18 ↔ T25 coupling.** M22 (pin/archive/mute) will be the first~~ resolved by events — implemented together: T18 + T25.1 in one cycle (versioned migrations)
   store schema change and therefore fires the `schema_version` gate.
   Implement both together in this train (bigger blast radius, one
   verification cycle), or defer M22 and leave T25.1 documented
   trigger-only (smaller steps, two cycles)?

---

_Point-in-time snapshot. `TODO_LIST.md` remains the living source; (f) is
HARVEST material for docs-health._
