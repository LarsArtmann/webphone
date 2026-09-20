# UI Redesign — Screenshot-Driven Review of pbx.artmann.tech (2026-09-20 12:14 CEST)

Session scope: the user asked why https://pbx.artmann.tech/ is "still so fucking
ugly" and demanded ACTUAL screenshots, a diagnosis, a fix, and verified
re-screenshots. Helper: /home/lars/projects/vision-review-agent (CDP plumbing).
No other research was done.

## Answer to the original question (why it looked ugly)

Verified via real screenshots (light/dark × desktop/mobile):

1. **Visible "Skip to main content" link on every page.** templ-components'
   Base emits the skip link with Tailwind `sr-only focus:not-sr-only`
   utilities, but webphone never loads Tailwind — so `sr-only` was dead CSS
   and the link rendered as plain blue underlined text at the top of every
   screenshot. Single biggest "unpolished" tell.
2. **Box-in-box-in-box hierarchy.** Page bg → panel card → gray row cards →
   island container card → login card inside it. One border+radius+bg recipe
   everywhere = no hierarchy, no focal point.
3. **Three nearly identical grays** in the light theme (#edf1f2 / #fbfdfd /
   #f1f5f6) — washed out, flat, no depth.
4. **Anemic signed-out state.** The whole right column was one dashed box
   with a single muted sentence. First impression = empty page, not product.
5. **Compose-first information architecture** on Messages (compose bar above
   the inbox), raw native "Choose Files" browser chrome, full-size solid
   Call + red Remove buttons on every contacts row (button wall), fax rows as
   an undecipherable one-line text soup, transcript bubbles up to 78% of an
   1100px-wide card.
6. **No avatars/visual anchors** in any list; numbers only; timestamps and
   previews all the same size.

## Tooling built this session (in /tmp/wpshoot/, throwaway)

- `wpshoot.py` — CDP screenshot driver reusing
  vision-review-agent's `scripts/cdp-shoot.py` classes (WS/CDP). Spec-driven
  (JSON): pages × login action × JS actions, light+dark, desktop+mobile,
  full-page captures. ~36 shots per round.
- `fake-sip.js` — in-browser WebSocket stub injected via
  `Page.addScriptToEvaluateOnNewDocument`: answers the island's SIP REGISTER
  with a synthesized 200 OK so the LOCAL island reaches the logged-in
  phone-panel state without a PBX. Status pill stays "offline" (only
  transport is faked) — good enough for layout review.
- `seed.py` — logs in (loopback dev mode), adopts CSRF, seeds messages
  (incl. one MMS with attachment), outbound message, two inbound faxes
  (real tiny PDF), four contacts, via the same webhook/API contracts the
  smoke suite uses.

## a) FULLY DONE

- Live-site screenshots: login page light/dark × desktop/mobile (the
  pre-auth state the user sees).
- Full local reproduction: loopback gateway + seeded SQLite + fake SIP.
- Ugliness diagnosis (above) with screenshot evidence.
- **Skip-link fixed**: app.css now owns `.sr-only` + `.sr-only:focus`
  (hidden until focused, visible when tabbed to). Verified gone in all
  new shots.
- **app.css rewritten** as a design system: new token block (deeper dark
  palette #0d1317/#151d22/#1b252b, light theme with teal cast #eef2f1,
  --border-strong, --radius-sm, softer two-layer shadow), one elevation
  model (panels elevated, rows flat tiles with hover border), welcome
  panel styles, row grids, avatar styles, dir-chip, wp-mini quiet action
  buttons, `::file-selector-button` styling, 46rem centered transcript,
  asymmetric bubble tails, compose-as-footer, mobile `:has()` order flip
  (login card leads when signed out), reduced-motion + indicator rules
  preserved.
- **island/style.css**: tokens synced with the shell (per AGENTS.md
  mirror rule), island's outer double-frame removed (border/shadow on
  .card only), bigger login h1 (1.3rem/-0.02em), keycap gradient +
  inset highlight on keypad buttons, 15px base font.
- **templ structure**: ThreadRow avatar + two-line text + unread tint +
  unread badge side slot; ContactRow/Shared row avatar + quiet actions;
  FaxRow/CDRRow/VoicemailRow → wp-row grid with arrow dir-chip
  (aria-labelled), status/pages inline, mini buttons; ThreadsPanel
  compose moved BELOW the list; SignInHint → full welcome panel
  (headline, body, three capability points, accent CTA hint).
- **helpers.go**: `avatarFor` (name initials / "+1" country signum —
  fixed after first version returned confusing "14") and `avatarHue`
  (deterministic oklch hue per peer).
- **i18n**: welcome.* ×3 keys + fax.pdf added to BOTH en and de;
  threads.empty copy fixed in both ("start one below" — compose moved).
- **Swap-safety contract kept**: `wp-fax-row` stays in the class list
  (TestSSEPushesSwapSafeFragments greps it) — documented in fax.templ.
- Verification rounds: three full re-shoot rounds; latest round visually
  confirmed: Messages (light+dark), Thread, Fax, Contacts, Island
  registered + revealed (keypad/log), mobile Messages. The improvement
  is large and coherent in both themes.
- `GOEXPERIMENT=jsonv2 go test ./internal/...` — ALL PASS (incl. the 35-id
  DOM contract test, SSE swap-safety, vm morph-id tests, i18n sync test).

## b) PARTIALLY DONE

- **Final verification reshoot** — a full round (36 shots) was launched at
  ~12:10 and was still RUNNING at 12:14 (background shell 03B, had printed
  "ALL_TESTS_PASS", build step underway). The avatar "+1" fix and the
  history dir-chip fix are IN this round but NOT yet visually reviewed.
- **Visual review of Voicemail/History/Settings** in the latest round —
  earlier rounds reviewed; the newest round's shots not yet viewed.
- **Live-site sign-off** — impossible without live PBX credentials; the
  deployed binary is also behind (stack pins v2.4.0 a59f0d1). Local
  verification stands in.
- **Stack browser E2E** (`nix build -L .#telephony-browser` in
  nix-international-telephony) — NOT run. Markup changes are additive
  (ids intact, classes added; E2E-greppable island strings untouched) so
  it SHOULD pass, but it is the island regression gate and remains
  unproven.
- **Commits** — nothing explicitly committed by me this session (the
  auto-commit daemon may have swept working-tree changes; state at 12:14
  not re-verified). Per the release runbook, doc folds + explicit commits
  per task are still owed.
- **AGENTS.md update** — new session-durable facts not yet written:
  app.css owns `.sr-only` (templ-components assumes Tailwind), `wp-fax-row`
  is a pinned SSE contract class, dir-chip arrow glyph + aria-label
  pattern, avatarFor/avatarHue helpers, /tmp tooling pattern.

## c) NOT STARTED

- `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` quality gate.
- `nix flake check` (incl. island-lint oxlint gate — new CSS/JS untouched
  by oxlint scope, but unproven).
- `python3 scripts/webphone-smoke.py` (28-check live smoke).
- Stack-side browser E2E + webphone VM test re-run.
- CHANGELOG/FEATURES fold for the redesign; version train decision.
- Unit tests for `avatarFor`/`avatarHue` (pure functions, untested).
- de-language screenshot pass (all reviewed shots are EN).
- Contacts import-row grouping (file picker + Import on one row) —
  identified as awkward, fix not implemented.
- Dial input placeholder truncation in the island at narrow widths
  (placeholder "+44163…" clips) — noted, not fixed.
- flake.nix `webphoneVersion` bump (release-time task).

## d) TOTALLY FUCKED UP (and how it was recovered)

1. **Stale-server shambles (cost ~2 full shoot rounds).** `kill` by
   pid-file did not stop the old binary (graceful shutdown keeps waiting —
   SIGTERM swallowed); new servers then failed to bind :18099 while the
   OLD binary kept serving OLD embedded assets. Two reshoot rounds
   "verified" nothing. Recovery: `pkill -9 -f webphone-bin` + served-CSS
   content check before trusting any round.
2. **`;`-chains masked failures.** A chained `templ generate` failed
   (exit 1), `&&` skipped the rebuild silently, but the `;`-separated
   boot+seed+shoot still ran — another round of old-markup screenshots.
   Recovery: fail-fast chains + `rg "↓" fax_templ.go` style post-condition
   checks; lesson matches the AGENTS.md "pipeline masking" entry — I
   repeated it anyway.
3. **templ attribute if-expressions are illegal.** My first fax fix put
   `aria-label={ if … }` in the attribute — compile error at generate.
   Recovery: `faxDirGlyph`/`faxDirLabel` helper funcs (same for CDR).
4. **Edit-without-re-read collisions with a PARALLEL SESSION.** Another
   session added an Identity feature mid-flight (fax.templ Identity prop +
   `wp-identity` block, panels.go `identityFor`, i18n `identity.from`
   en+de). My stale old_strings bounced off the mod-time guard three
   times; I re-read and rebased my edits on the current file each time and
   did NOT touch the foreign work.
5. **Dirty demo data in early rounds** (badge "12", duplicate MMS threads)
   because seed.py ran against surviving old servers — early judgment made
   on noisy data; later rounds wiped the data dir and seeded exactly once.

## e) WHAT WE SHOULD IMPROVE (process/code, in priority order)

1. One fail-fast driver script for the whole loop
   (generate→test→build→boot-on-fresh-port→seed→shoot→verify-by-content)
   instead of hand-chained shell; verify server identity by fetching a
   build-specific marker (e.g. CSS content hash) before shooting.
2. Never kill by pid-file: `pkill -9 -f` by binary name + `ss -ltn` port
   owner check; or boot each round on a NEW port (17991+) to make stale
   servers structurally irrelevant.
3. Add go unit tests for avatarFor/avatarHue (table-driven; "/" blank
   edge cases).
4. Review the de locale + remaining tabs (voicemail/history/settings
   empty states) in the FINAL shot round before declaring done.
5. Contacts import/export row: group file+Import on one row, export link
   right-aligned (small templ+CSS change).
6. Island dial placeholder truncation (shorten placeholder or grow input).
7. Avatar chroma/lightness tuning (currently very subtle in light mode).
8. Consider a `prefers-contrast` / focus-ring contrast audit pass.
9. Fold the session into AGENTS.md (sr-only ownership, wp-fax-row
   contract, dir-chip pattern, fake-SIP-in-CDP testing trick).
10. Commit per task with explicit messages (daemon history is noise).

## f) UP TO 50 THINGS TO GET DONE NEXT (ordered, actionable)

**Ship this redesign**
1. Review the 12:10 reshoot round (all 36) — confirm avatar "+1", history
   arrows, voicemail/history/settings, mobile dark.
2. Fix contacts import row grouping (templ class + CSS).
3. Island dial placeholder truncation fix.
4. Add avatarFor/avatarHue unit tests.
5. Update AGENTS.md (sr-only, wp-fax-row contract, dir-chip pattern,
   token-mirror reminder now includes --radius-sm/--border-strong).
6. Explicit git commits in task-sized groups (CSS system / island CSS /
   templ rows / welcome panel / i18n / helper tests).
7. Run `python3 scripts/webphone-smoke.py` (fresh binary + data dir).
8. Run `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`.
9. Run `nix flake check` (island-lint included).
10. Run the stack browser E2E against the branch (`--override-input`) —
    the real DOM-contract gate.
11. Cross-build `nix build .#webphone --system aarch64-linux` + verify
    ELF machine bytes (per runbook).
12. Fold CHANGELOG (Unreleased → dated section) + FEATURES entries.
13. Decide release train (v2.5.0), bump `webphoneVersion` in flake.nix.
14. Tag + push, lychee link check, stack bump (`nix flake lock
    --update-input webphone`), stack gates, announce.
15. Verify the LIVE site post-deploy with the same CDP harness
    (light/dark login + a real logged-in pass with real credentials).

**Visual polish backlog (from screenshot critique)**
16. Mobile: island card `order` flip for signed-in state review —
    confirm keypad reachability thumbs-wise.
17. Bubble meta: show read-status icon (✓) instead of underlined text?
18. Thread header: resolve contact NAME for the remote number (contacts
    store lookup) instead of raw number — product gap visible in shots.
19. Thread list: apply contact-name resolution to rows too.
20. Fax compose: collapse number+file+send into one bordered unit card.
21. Voicemail rows: show duration as mm:ss and a waveform-ish progress
    (audio only has native controls today).
22. Empty states: add one-line "what will appear here" sub-copy per tab.
23. Welcome panel: add version string + gateway mode (settings-lite).
24. Theme toggle: cycle label "Auto → Light → Dark" with state icon.
25. Signed-in chip: add a green LED dot mirroring reg-status.
26. Nav: keyboard focus ring consistency check across browsers.
27. Print stylesheet? (fax/history lists) — probably YAGNI, decide.
28. Dark theme: border-strong maybe too subtle on rows in the latest
    shots — verify at 100% zoom.
29. Light theme bg teal cast: confirm it does not band on cheap panels.
30. `oklch()` fallback for older Safari (island ships raw; check
    browserslist reality for Lars's fleet — Chromium only today).

**Architecture/clean-up surfaced this session**
31. Kill the `/tmp/wpshoot` throwaway: consider promoting the shoot
    harness into the repo (scripts/ui-shoot/) or into vision-review-agent
    as a reusable "app harness" (spec + seed + fake-SIP).
32. Document the fake-SIP WebSocket trick in AGENTS.md testing section.
33. Smoke suite: add a check that /assets/app.css contains the current
    design-token marker (catches stale-embed deployments — the exact
    failure mode that burned this session twice).
34. Consider adding an E2E-greppable design marker to the login page
    (like the island strings) so the stack E2E proves the new shell.
35. i18n: `vm.from` key now unused after row redesign — remove or reuse.
36. `wp-fax-dir`, `wp-thread-side`… sweep for dead CSS classes left from
    the old row layout and delete (dedupe pass).
37. Check `.wp-filter` (history) still looks right with new input styles.
38. Contacts: sort rows alphabetically (currently reverse-insertion —
    looked arbitrary in shots).
39. Fax list: group outbox/inbox or add a status filter (product call).
40. Messages compose: Enter-to-send in the reply field (JS island-side,
    shell must not depend on it).

**Verification debt**
41. Re-run the FULL `go test ./...` (not just ./internal/...) after the
    parallel Identity session settles.
42. gitleaks/codespell on-demand buildflow steps on the final tree.
43. Vulnix `nix run .#vulnix` re-run at release time (post-rev bump).
44. Re-verify `wp-` class inventory vs CSS (grep both directions) to
    catch dead/misnamed classes before release.
45. Add the reshoot step to the release runbook §9 closing sweep
    (server process dead-proof + served-marker check).

**Bigger swings (needs a decision, not urgent)**
46. Two-column → three-region responsive plan for wide screens (transcript
    + thread list side-by-side) — screenshots show the thread view wastes
    the right half; a desktop-class messaging layout would exploit it.
47. Move the island's topbar (EN/OFFLINE) INTO the login/phone card —
    the floating strip above the card still reads slightly disjoint.
48. Consider icons in nav (inline SVG, CSP-safe) — skipped this session
    for scope, would aid scanability.
49. Dark-first as DEFAULT (force dark on first visit before cookie)?
    The dark theme is objectively the stronger look; product decision.
50. Real PBX-credentials staging account so UI verification can include
    the REGISTERED island state with live data (reg pill green, ICE
    stats, real voicemail) instead of the fake-SIP stand-in.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Live credentials**: can you give me a test extension + password on
   pbx.artmann.tech (or spin a staging PBX) so the final sign-off includes
   the REAL registered island (green pill, real voicemail/history)? Local
   loopback + fake SIP covers layout but not those live states.
2. **Release train**: should this redesign ship as v2.5.0 after the stack
   browser E2E passes, or do you want it split (CSS-system commit now,
   templ structure behind review later)?
3. **The parallel session's Identity feature** (fax "sending as" +
   `identityFor` in panels.go): is that yours and still in flux? My final
   reshoot may interleave with it — should I wait for it to settle or
   just verify around it?

---

*Report per instruction; waiting for instructions. Background reshoot
(shell 03B) may have finished by the time you read this — shots land in
/tmp/wpshoot/local/.*
