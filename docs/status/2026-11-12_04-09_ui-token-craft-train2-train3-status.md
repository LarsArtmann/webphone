# UI token/craft completion, Train 2 + Train 3, harness upgrades — session status and self-review

**Claims verified as of 04:11 CET, 2026-11-12** — push/CI statements rot fastest; the push verdict was re-read live (`git ls-remote` = HEAD `0131bd0a`) minutes after a parallel session healed and pushed the tip. The island-suite verdict below was re-read from `node --test` output this session: **64 pass / 1 fail — the fail is mine** (the Phone-tab spec, §d-1).

**Scope of this report.** VERIFIED: the token batch (state layers, focus ring, type ramp, nav fit, easings sweep) in both stylesheets with pixel-evidenced captures; the row-gap cycle fix + its 8 consumers identified; the capture-harness `--hover` flag (self-test green); the pre-commit Python gate (selftest 10/10); the empty-state brand pass (5 surfaces, templ-pinned, view+server suites green); Train 2 (History table, Reports dashboard grid, contacts fitted grid — view+server suites green); Train 3 shell.js Phone tab (tsc-clean via vtsls, CSS gated); desktop + mobile 26-shot captures eyeballed. NOT verified: the full gate battery (buildflow, smoke, pre-push-gates did NOT run this session); `nix flake check`; the Reports/History surfaces were never re-captured after the Train 2 restructure (the codegen + pins are green, the LOOK is unproven — the owner bar is the LOOK); hover/focus/reduced-motion captures of the NEW surfaces (reports grid, history table).

---

## a) FULLY DONE

| Work | Evidence |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Modern-web-guidance searches + 9-guide retrieval run BEFORE the CSS work (the §e-6 debt from the last report paid) | `/tmp/mwg-guides.md` (90 KB: natural-interactions, responsive-table, card, navbar, motion, color, toasts, dark-mode, design-token-reactivity); its rules shaped the state-layer idiom, the focus-ring contrast check, the table pattern, the reduced-motion refinement |
| `--row-gap: var(--row-gap)` cycle FIXED at app.css:72 (intended 12px), and a repo-wide regex sweep proved it was the ONLY cyclic custom property in all four CSS sources | python re-scan over app.css / island style.css / tw.css.input / login.css.input: zero other hits; the 8 consumers identified by name (`.wp-brand`, `.wp-header-actions`, `.wp-skeleton`, `.wp-rail-row`, `.wp-unread-divider`, `.wp-reports-figure/-totals/-row`) and rendered 12px in the capture |
| The archive-button "pale-red plate" hunt CLOSED with pixel evidence, no probe server needed | sampled `04-messages-light.png` at the plate: (229,162,159) hovered / (242,207,205) at rest vs surface-2 (244,246,249); computed the danger icon strokes at the flag's 0.45 rest opacity over surface-2 ≈ (229,161,157) — the plate IS the deliberate at-rest quieting, there is no stray rule. Ruling: keep |
| Token batch shipped in BOTH stylesheets and committed (`37af2d80`) | `--text-lg/xl/2xl` (the fluid M26 display curves preserved INSIDE the tokens), `--state-hover/pressed` (8%/12%), `--focus-ring`+`--focus-ring-offset`, `--dur-press` (80ms); 28 literal easings swept onto tokens (`120ms ease`→`var(--dur-fast) var(--ease)` etc., sed-verified zero literal `ms ease` left in app.css); ghost row actions ride M3 state layers + micro-scale press; island mirrored (focus tokens, `--dur-press` on keypad/a.primary, toast slide-in keyframes, reduced-motion `animation-iteration-count: 1` refinement per the motion guide's 0.01ms-restart hazard) |
| Nav single-row fit at 1280 with the island open (§f 12) — fluid cqw compaction, not a breakpoint | `.wp-nav-link` gap/padding ride `clamp(...cqw)` of the tab column + `@container tabcol (max-width: 880px)` font step; verified IN THE CAPTURE: 1280×900 desktop shows all 7 tabs on ONE row in both themes |
| Bubble burst grouping + rail 2-line clamp + thread-side baseline (§f 10/11/47) | adjacent-sibling margins (−4px in-burst, +4px at polarity change), `.wp-rail-preview` 2-line `-webkit-line-clamp`, `.wp-thread-side` baseline with centered minis |
| Harness `--hover SELECTOR` flag shipped (§f 31/§e-7) — interactive states are now capturable | `apply_hover()` (best-effort, silent no-op on surfaces without a match), threaded through both capture call sites; `ast.parse` + `--self-test` "all checks pass"; the desktop run used `--hover ".wp-thread-rowwrap"` and the shot shows the row's hover state layer + solidified actions |
| Pre-commit Python syntax gate shipped (§f 34, the §d-class that landed a daemon corruption at HEAD once) | Gate 3 in `.githooks/pre-commit` (`py_compile` over staged `scripts/*.py` blobs, python3-missing = WARN like node); selftest extended to 10 cases, all green |
| Empty-state brand pass (§f 14's CSS half, respecting the 2026-11-04 dedup ruling) | `Class: "wp-empty-state"` on the 6 list EmptyState call sites (search-empty, threads, contacts, vm, fax, reports-undated); app.css: accent circle tile + accent icon + `--text-lg` title + left-aligned rhythm, scoped via the component's own Class prop (library markup stock); templ regen'd with the pinned CLI; view + server suites green |
| Train 2 — History is a real TABLE (§f 15; the responsive-table guide's pattern) | semantic thead/tbody, sticky column heads inside a scroll-region wrapper (role=region + tabindex=0), right-aligned tabular When/Duration columns, day groups as tbody sections (the `wp-day-head` pin survives as a table-row head with `display: table-cell` reset — the transcript's flex rule would have collapsed it), `data-label` cells drive the CSP-safe stacked mobile lane (`::before content: attr(data-label)`); missed-row danger edge moved to first-cell inset (TR borders don't render in separate-borders model); i18n keys in BOTH maps; `TestHistory*` + parity + DOM-contract green |
| Train 2 — Reports dashboard grid (§f 16) | `.wp-reports-grid` wrapper; at `tabcol ≥1000px`: hero + totals + undated + top-contacts span full width, day/hour ladders sit side by side; per-ladder modifier classes added; undated EmptyState gets `wp-reports-empty` so the auto-flow never splits it |
| Train 2 — Contacts fitted grid (§f 17) | `.wp-contacts-grid` auto-fill `minmax(340px, 1fr)` — two-up at desktop width, one column on narrow, no container query needed |
| Train 3 — the Phone tab shipped in shell.js + CSS (§f 18) | JS-created button in the strip (a VIEW toggle, not a route — the tab→island-affordances doctrine), `aria-pressed` state channel, clears on real-tab clicks, re-installs after nav morph swaps; CSS: desktop-hidden, mobile-only, `html.wp-phone-focus` makes the island the one body column; button reset rides the shared `.wp-nav-link` class; TS-clean |
| Cargo verification: parallel-session discipline held through a heavy overnight overlap | the sibling session shipped the storage-ladder notes/reviews deletion, the v2.14.0 release end-to-end (`dbe2ed8b` verdict), and a red-main heal (`0131bd0a`, passkey user_id one-home + art-dupl baseline) — my batches committed atomically (CSS `37af2d80`, then daemon-carried trains), zero reverts of their in-flight files |

## b) PARTIALLY DONE

| Work | What exists | What is missing |
| ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Train 3 end-to-end | shell.js + CSS landed and type-clean; two specs written | **the Phone-tab spec is RED** (§d-1) — the suite is 64/1, so the island gate would fail CI if run now; the mobile capture never re-shot with the tab visible; the dialer-focus LOOK is unproven |
| Visual QA passes (§f 7/8/9) | desktop + mobile 26-shot sets captured with hover; messages/history/logins/404 eyeballed both themes | the NEW surfaces (history table, reports grid, contacts grid, phone tab) captured BEFORE their restructure or not at all — no post-Train-2/3 capture run happened; hover/focus/reduced-motion passes over the new surfaces not run |
| §f item 19 (call-card densify, disposition/note craft) | ruled SATISFIED-BY-EVIDENCE: the rows already carry the flex/muted/input craft + Train 1 tokens | nothing further — strike queued, not yet struck in the ledger |
| §f item 32 (seed-helper extraction) | ruled RESOLVED-BY-EVIDENCE: the named duplicate was the throwaway probe, deleted same session; remaining scripts are distinct documented tools | ledger strike pending |
| Item 41 (easing sweep) | app.css fully swept (28 sites), island's two 80ms press sites onto `--dur-press` | island's remaining literals: only the shimmer 1.1s animation (deliberate, out of scope) |
| Logical-property polish (NOT mine, live in the tree right now) | an uncommitted `text-align: left→start / right→end` sweep over my table/empty-state CSS + an `onAccept?.()` stub fix in calls.test.mjs | another session's in-flight edits — I did not touch or commit them; they may or may not land |

## c) NOT STARTED

- The full gate battery over THIS session's trains (buildflow, smoke, pre-push-gates, `nix flake check`) — the runs stopped at view/server suites + island suite + fmt + tsc; the trains were daemon-committed mid-flight by design (concurrent-session doctrine), but a clean-tree full battery never ran.
- Post-restructure captures (History table, Reports grid, contacts grid, Phone tab) — the §7 acceptance for Train 2 says "screenshots show restructures, not just paint"; that evidence does not exist yet.
- A11y audits (§f 24/25: ghost icon contrast, 200% zoom) — untouched.
- docs-health HARVEST of this session's §f rows into TODO_LIST + ratchet re-measure (§f 39/40 of the old ledger) — superseded by THIS report's ledger but not yet struck anywhere.
- Reduced-motion + theme-toggle visual passes over the new surfaces.
- Train 3 mobile capture acceptance ("ring-while-closed UX parity" eyeball).

## d) TOTALLY FUCKED UP (all mine; nothing shipped broken except the red spec)

1. **I committed a RED spec into HEAD (via the daemon's mid-flight commit) and the suite now sits at 64/1.** The Phone-tab spec fails because the node-stub's `querySelector` walks registry-root SUBTREES only: my `doc.body.append(nav)` puts the strip under a parent that is NOT a registry root, so `document.querySelector(".wp-nav")` inside `installPhoneTab` can never find it (a registered root can't match itself, and the body→nav edge is never walked). The root cause is FULLY understood (debugged to a two-line repro: `doc.querySelector('.wp-nav'): MISS, tab: MISS`); the fix is one seam change (`installPhoneTab(nav)` accepting the strip the spec holds, or registering a wrapper root in the spec) + a rerun. I should have run the spec BEFORE ending the Train 3 edit batch — the edit-then-verify ordering slipped exactly once, and it was the batch that mattered.
2. **The sed sweep bypassed the edit tool's mod-time guard and then tripped it on myself** — my own `sed -i` invalidated the tool's read state, costing a wasted multiedit round-trip (the guard rejected, I re-viewed, re-applied). Cheap, but avoidable: after any bash-side file mutation, re-view before the next tool edit.
3. **One multiedit wrote 1-space indentation into app.css** (the tool's "re-indent to file style" heuristic misfired on the CSS blocks; neighbors are 2-space). Caught on read-back, healed by `nix fmt` (prettier owns CSS — canonical bytes verified by `cat -A`). The gate hierarchy worked, but I shipped non-canonical bytes to the working tree twice before the formatter fixed them.
4. **A no-op multiedit against reports.templ** (identical old/new strings — I wrote the edit before noticing the anonymous-loop arm should stay hover-free by design). Cost one round trip; the guard accepted it silently, which means no-op edits DO land — review what you send.
5. **A `question`-shaped gap in my own workflow: I never re-verified the live origin tip before my first push-bar check** and read a 3-hours-stale position as "unpushed backlog" — the daemon had pushed while I was retrieving guides. Cost: one wasted analysis pass; the 04:11 live re-read corrected it (origin == HEAD, everything pushed).
6. **The mobile "mid-page nav" scare was my own misread** — a full-page CDP capture renders `position: fixed` elements at the viewport-bottom of the ORIGINAL viewport inside the taller image, so the fixed bottom bar appears mid-page. I nearly filed a regression against the restyle baseline before checking the media block (intact) and reasoning the artifact through. The capture harness's full-page mode needs a fixed-element caveat note so the next session doesn't burn the same scare.

## e) WHAT WE SHOULD IMPROVE

1. **Never end an edit batch on tests without running those tests.** The Train 3 batch ended at "TS-clean via the editor LSP" — but the repo's own rule says the editor LSP is not the authority for the suite, and the same laziness applied to node:test. Spec-first or spec-immediately-after; never "specs written" as a done state.
2. **Node-stub contract documentation:** the fake DOM's registry-root querySelector limitation (roots can't match themselves; only registry roots are walked) cost a debug session. One paragraph in island-tests/helpers.mjs's header comment (or the README) would have saved it.
3. **The capture harness should print a fixed-element caveat** on full-page runs (mobile bottom bar appears mid-image) — one `print()` in shoot().
4. **The daemon's mid-flight commits carry unverified suites.** This is the standing trade (concurrency beats gating), but when I KNOW a batch contains a red spec I should either fix it forward immediately or annotate the commit — a silent 64/1 in HEAD is how a red main starts (it didn't this time only because CI hadn't run the island suite on that push yet, and the heal session's push raced ahead).
5. **Logical-property sweeps belong to a linter.** The sibling session hand-edited `text-align: left→start` over my fresh CSS — a stylelint-style rule would catch these mechanically instead of by eyeball across sessions.
6. **Pixel sampling beats DOM probing, again** (second confirmation this train): the archive-plate question closed in ONE pillow script after the previous session burned ~10 tool calls on selenium probes.

## f) Up to 50 things we should get done next

<!-- check-rows:ledger -->

| # | Area | Task | Why / verification |
| --- | ---------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| 1 | train3-tests | Fix the red Phone-tab spec: pass the strip into `installPhoneTab(nav)` (or register a wrapper registry root in the spec), rerun to 65/0 | the suite is 64/1 at HEAD `0131bd0a`; root cause debugged to the stub's registry-walk rule |
| 2 | capture | Re-shoot desktop + mobile AFTER Train 2/3 (history table, reports grid, contacts grid, Phone tab) + `--hover` on a row | §7 acceptance: "screenshots show restructures, not just paint" — no post-restructure shots exist |
| 3 | gates | Run the full battery on a clean checkout of `0131bd0a`: buildflow, smoke, pre-push-gates, tsc, island suite, `nix flake check` | this session's trains never saw the full battery (suites-only green) |
| 4 | ledger | Strike the resolved §f rows (2 plate, 19 densify, 32 seed-helper, 33 `--surface` existed, 34 gate shipped) with evidence | entombed-open-rows are the docs-health failure class |
| 5 | harvest | HARVEST the open rows of this report + the 16:44 report into TODO_LIST (the annotate tool rebuilt fresh per session) | the standing close-the-loop rule |
| 6 | ratchets | Re-measure `.docs-gates.json` annotate_max/audit_max after the overnight corpus growth (destination doc, 2 status reports, AGENTS bullets) | DF-brittle pins flip when living docs grow |
| 7 | a11y | Ghost-icon contrast audit (the 0.45-opacity danger minis measured (229,161,157) on surface-2 — borderline for 3:1 non-text) | pixel evidence in hand; WCAG 1.4.11 |
| 8 | a11y | 200% zoom / rem pass over the table + grid surfaces | percentage root, new layouts untested |
| 9 | motion | Reduced-motion capture pass (`prefers-reduced-motion` emulation) over panel transitions + toast slide-in | the `animation-iteration-count: 1` refinement never LOOKED at |
| 10 | theme | Manual `data-theme` toggle capture (not just prefers-color-scheme) | the override path unviewed |
| 11 | reports-dark | Dark-mode check of the budget-over meter + the new two-column grid | §f 28 carried over |
| 12 | history-i18n | Verify the stacked table lane's `data-label` lines in German (longest labels: "Zeit"/"Dauer") | the stack lane renders ::before content from markup — desktop-only so far |
| 13 | train3-mobile | Mobile capture with the Phone tab: tap-through (focus → island full-screen → real tab returns) via a driver script | the tab exists; the LOOK/flow unproven — the §7 Train 3 acceptance |
| 14 | train3-e2e | Stack browser E2E still drives the island desktop-first — confirm no selector depends on the strip having ONLY server tabs | the strip gained a button child; `refreshNav`'s active-tab walk uses `[data-tab]` (safe), but verify |
| 15 | nav-i18n | The Phone label is English-only (D3 shell rule) — confirm the German nav strip reads acceptably or get an owner exception | shell copy English vs the localized tabs beside it |
| 16 | dom-contract | Document `wp-nav-phone` + `wp-phone-focus` + `wp-reports-grid`/`wp-table-wrap` in docs/dom-contract.md if any get pinned | new served ids/classes; the contract is the single source |
| 17 | smoke | Add a smoke check for the reports grid/history table render (the smoke greps served bytes; new surfaces deserve a marker) | the smoke's history checks may still expect `.wp-row` article markup — VERIFY, the restructure changed `article`→`tr` |
| 18 | harness | `--hover` should accept per-surface pairs (`surface=selector`) so the hover matrix can vary per shot | current single selector applies globally per run |
| 19 | harness | Fixed-element caveat print on full-page captures (the §d-6 scare) | one print, prevents the next misread |
| 20 | stubs | Document the registry-root querySelector limitation in helpers.mjs (§e-2) | cost a debug session |
| 21 | css-arch | The `.wp-day-head` dual-context rule (flex in transcripts, table-cell in tables) is a split-brain seed — consider distinct classes at the next touch | one class, two display semantics |
| 22 | table-a11y | The table wrapper's `tabindex="0"` + region label needs a focus-visible check INSIDE the scroll container | sticky heads clip the ring (the motion guide's clip warning) |
| 23 | empty-action | Empty states carry pointed-to sentences ("add one above") but no action button — owner eyes on whether a button duplicating the in-panel composer is wanted | library supports ActionText; ruled out this session as a duplicate affordance |
| 24 | nav-badge | Badge pip + Phone tab coexistence on 390px (8 slots now) — capture and eyeball the crowding | the mobile bar gained a slot |
| 25 | island-style | Mirror-check: island/style.css gained the token additions but NOT the state-layer component sweep (ghost minis are app.css-only) | the island's own buttons use :active scale; hover layers absent — deliberate or gap? |
| 26 | loginpage | Re-shoot /login post-token-flow (login.css rebuilt in the batch) | §f 23 carried over |
| 27 | cqw-fallback | The nav cqw compaction assumes container-query support — verify the fallback path (no `cqw` units = fixed paddings) is acceptable | modern browsers fine; the repo targets latest-only per AGENTS, so likely a non-issue |
| 28 | release | The `[Unreleased]` section needs entries for this session's trains (token batch, table, dashboard, phone tab, harness, gate) | the CHANGELOG bar: committed code-path changes get a line |
| 29 | conspec | Train-3 acceptance wording in the destination doc says "ring-while-closed parity" — the tab is navigation, not push; align the wording with what shipped | §7 table honesty |
| 30 | dupl | Run art-dupl over the new view/CSS blocks (table stack lane vs reports grid vs contacts grid share grid idioms) | the baseline is version-pinned; new clones should be adjudicated fresh |
| 31 | whitespace | The 1-space CSS incident: consider a pre-commit CSS-indent check via whitespace-drift's formatter adjudication (it should already catch — verify it does) | §d-3 |
| 32 | capture-mobile | Mobile runs use `--viewport 390x844` only; add a 390-wide run WITH the island focused (`wp-phone-focus` seeded via script) | the dialer-focus layout is captureable without taps |
| 33 | docs | AGENTS.md: the Commands block's smoke-boot port note (dnsblockd) landed; add the PYSEL `--hover` example to the capture recipe | the flag is discoverable only in the script |
| 34 | i18n-parity | The 5 new history.table/col keys passed parity — run the unused-key test to confirm no dead key | parity tests green, unused-key not re-run explicitly |
| 35 | sticky-heads | Sticky thead z-index vs the palette overlay (z 200) and toasts (z 100): the table heads use z-index 1 inside the wrapper — confirm no stacking bleed | overlay contexts |
| 36 | grid-fallback | `.wp-reports-grid` at exactly 1000px tabcol with BOTH ladders absent (undated arm): eyeball the empty-state span | edge of the container query |
| 37 | arch-test | `internal/arch` import rules unaffected (views changes only) — confirm the arch suite still passes in the full battery | not run this session |
| 38 | sse-fragments | The history table rides the panel outerHTML swap — confirm the SSE `review`/`notes` nudges re-render the TABLE correctly (morph + tbody sections) | morph + table sections is an idiomorph edge |
| 39 | contacts-sse | The contacts grid wrapper rides the contacts nudge re-fetch — verify morph preserves the grid (wrapper is inside the swap region) | same class as 38 |
| 40 | perf | 26-shot capture time grew with hover settle (0.15s/surface) — negligible, but track if the matrix adds more flags | note only |
| 41 | templ-cli | The regen re-emitted the WHOLE package (18 files) from a 4-file invocation — confirm the pre-push templ gate still passes (it did in the 16:44 session's shape) | the collateral re-emission is committed |
| 42 | baseline | art-dupl baseline version pin: this session ran no detector — the CI-pinned detector owns the next sweep | standing rule |
| 43 | island-lint | oxlint covers shell.js — the new globals/vars (`WP_PHONE_SVG`, closures) pass lint only inside `nix flake check`; run it | nothing gates shell.js at commit time |
| 44 | type-pin | `types/island.d.ts` if `wp-nav-phone`/`wp-phone-focus` become pinned ids — the pin test cross-checks dom-contract.md | only if 16 lands |
| 45 | lessons | Record the fixed-element capture artifact + the stub-registry rule in docs/lessons.md | the war-story home |
| 46 | vendored | No vendor/module changes this session — vendorHash untouched, verified by the absence of go.mod/go.sum in my commits | clean |
| 47 | changelog-guard | The pre-push changelog gate will demand the §28 line — write it BEFORE the next push | gate-ordered |
| 48 | next-train | Train 2 residue: settings fitted layout ruled satisfied-by-evidence (definition grid + groups exist) — strike or upgrade by owner eye | the half of §f 17 not touched |
| 49 | owner-eye | The §g answers from the 16:44 report remain unanswered (canonical greens, desk-lean weight) — the trains above assumed keep-current-values + §6's own prescriptions | owner taste |
| 50 | process | The scratch-drain rule (§e-1 of the 16:44 report) worked this session — the row-gap fix, the plate hunt, the easings all left the scratch as work or rows; keep it standing | positive note |

## g) Questions I cannot figure out myself

1. **The Phone tab's fate on the desktop bar.** I shipped it mobile-only (CSS-hidden ≥900px) because desktop has the island column. Do you want a desktop Phone affordance at all (e.g. a header pill that pulses on incoming ring), or is mobile-only the right shape? This decides whether `wp-nav-phone` stays display:none on desktop or becomes a real control.
2. **Is the History table's mobile stacked lane the right mobile shape, or should History keep the old card rows on phones?** The table restructure is a desktop scan win; the stacked lane (thead hidden, labeled lines) is the guide's pattern but denser vertically than the old cards. Your eyes decide — it's a one-container-query revert.
3. **Push/CI posture right now:** origin == HEAD `0131bd0a` (live-verified 04:11), but that tip contains MY red island spec (64/1) and CI had not reported on it as of this report. Do you want me to fix the spec forward immediately (one seam line + rerun) and push that as the next commit, or is the active heal session already owning the island suite? I did not touch the tree after reading your hold instruction.

---

**Self-verdict, one line:** the token/craft/harness/Train-2 program shipped verified and green, Train 3 shipped code-complete with its spec red by one stub-contract line I chose not to hot-fix under the report-first order — the failure mode (edit-batch-without-final-test-run) is named in §d/§e and is the one thing to change.
