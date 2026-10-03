# Session 7 — T18 finish + T19 + T21 + T22 + T23: status

Date: 2026-10-02 (~10:15 → 11:41). Resumed from the session-6 briefing
(`docs/status/2026-10-02_10-11_master-todo-train-session6-status.md`),
executed §f items 1–29 in order. Standing order: the whole TODO list,
verified. This report ends with a WAIT — the remaining work is the
harvest + final gates (items below).

> ARCHIVED 2026-10-03 (docs-health v6 sweep): T18 closed whole, T19/T21/T22
> landed, the T23 harness shipped working (14 shots, LOCAL-ONLY); the tail
> (eyeballs, DOM assertions, harvest, final gates) closed at the 12:57
> closeout; the §g defaults stood. Per-item verdicts inline.

## a) Fully done (verified green this session)

- **T18 views layer COMPLETE** (the two known bugs + everything §f 2–9):
  - `@if` → bare `if`; `pickLabel` → `flagLabel(lang, state, action,
    reverse)` — the SAME expression feeds label and posted value.
  - `ThreadHead` extracted (its own component): head flag buttons post
    to `flagURL` with `hx-target="#wp-thread-head"` + `hx-swap=
    "outerHTML"`; `setThreadFlag` branches on the `HX-Target` header
    and answers the FRESH head (island-honesty: the buttons flip to the
    network-confirmed state; transcript/scroll/draft untouched). Row
    buttons keep 204 + SSE list nudge.
  - Snippet lane: `SnippetChips` (quick, cap 5) + `snippetPicker`
    (`<details>`, opens UP) inside the REPLY composer only;
    `quickSnippets` + `chipLabel` (28-rune ellipsis) helpers.
  - `settings.templ`: `SnippetsSection` (list + ⚡ marker + delete +
    add form with quick checkbox); `SettingsPanelProps.Snippets`.
  - i18n en+de: `threads.pin/unpin/mute/unmute/archive/unarchive/
    archived/back`, `snippets.*` (8 keys), `toast.snippetSaved/
    Deleted`, `settings.crm`. Sync/verb/exists tests green.
  - `panels.go`: `archived=1` routing + `ArchivedThreads`/`ArchivedCount`
    (count fetched only in the active view), search hidden in the
    archived view (D6), Snippets into threadPanel (nil-safe) +
    settingsPanel (nil-safe; signature now `(r, sess)` returning error).
  - CSS T18 block (flags quiet-at-0.55, pinned tint, archived toggle,
    chips, picker, manage list, add form, lightbox over `--scrim`).
  - shell.js §2e (data-snippet fill = REPLACE + focus + picker close)
    and §2f (singleton `#wp-lightbox` native dialog).
  - **openapi entries SKIPPED, deliberately**: the spec documents the
    JSON API only (`/api/session|csrf|contacts`); htmx form actions
    (`/contacts/save` precedent) are out of its scope by design. Deviation
    from design-note line "Routes + openapi.json entries" — the
    /contacts/save precedent wins.
  - Full test map landed: store `threads_flags_test.go` (flag round-trips
    incl. owner-scope + unknown-flag rejection, archive exclusion
    list+search+count, D7 inbound-unarchives/outbound-stays, pinned
    ordering list+search, snippets CRUD/cap/replace-by-id, summary flags
    - LastAttachment), server `threads_flags_test.go` (anon 401/403,
      unknown flag/bad id/missing thread 404s, 204, HX-Target head
      re-render with CONFIRMED labels, mute drops+restores the nav badge,
      archive loop through HTTP incl. `Archived (N)` + auto-unarchive,
      snippet save/delete/422s/cap/partial re-render), views (row controls
      both langs + stable `thread-<id>` ids, archived toggle + search
      hiding, chips cap + picker-holds-all + no-lane-when-empty, settings
      section + empty state, `data-lightbox` on images only), shell
      (snippet fill + stranger no-op + orphan no-op, lightbox singleton +
      reuse + close). Runbook E2E obligation → T18; AGENTS.md seam bullet.
- **T19 COMPLETE** (M24 + M25, plan-faithful; §f16–18, A6/A10):
  - `TestNoUnusedDictionaryKeys` (dead-key guard; found and removed
    `vm.from` and `lightbox.close` — the latter was mine, dead on
    arrival: the shell close button is English by D3).
  - RTL logical-property sweep: 20 conversions (margins, borders,
    text-align, my two new offsets → inset-inline) across app.css +
    island/style.css + `TestStylesUseLogicalProperties` guard test.
  - Settings status dots (J10): `.wp-service-on/off` dots on the phoneAPI
    - new CRM row (`settings.crm`; `CRM.Enabled()` is nil-safe). Config
      truth, not live probes — documented in the CSS.
  - **A6 incoming focus mode**: `setIncomingFocusMode`/`showIncomingBanner`/
    `hideIncomingBanner` in ui.js (one home; all four banner transitions
    funnel through them), connection.js focuses Accept with a typing
    guard (INPUT/TEXTAREA/contentEditable never lose focus), CSS dims
    nav + tab content + header actions (visual only, nothing
    pointer-blocked). Two connection specs (dim+focus, typing guard,
    give-up clears).
  - **A10 pre-call device check**: `selftest.js` (`runDeviceCheck`:
    getUserMedia → label + immediate release, speaker count; missing
    mediaDevices degrades honestly), `wp-devtest-btn` in phone.templ
    advanced area, main.js wiring, `deviceTest` island i18n en+de, DOM
    contract entry. Four selftest specs.
  - `config.js` fix (see f): the WS scheme now follows the page protocol.
- **T21 COMPLETE** (batch 2): 21.1 golden + 21.2 tag guard + 21.5 dedupe
  were already landed (verified, not re-done). NEW: actionlint clean
  (21.4); **deadnix gate fixed** — vendored upstream flake.nix files
  (templ, tailwind-merge-go) fail `deadnix --fail .`; the check now
  excludes `./vendor` (was a latent red final gate); island-lint
  auto-covers selftest.js (dir glob — no file list to grow); backup-drill
  check builds green (21.3 drill half; the KVM `webphone-backup` VM runs
  at the final-gate flake check).
- **T22 largely done**: family pins added for `store.thread_flag` +
  `store.count_archived` (22.3; snippet pins existed). **health.css**
  (22.1/22.2): artifact + script were committed before; this session
  root-caused and fixed the REBUILD pipeline (below), rebuilt the
  artifact (it was stale — shipped zombie classes and MISSED
  `--blur-xs`), and wired `checks.health-css` (canary form — below).
  22.5: `nix run .#vulnix` exit 0 (all findings distro-patched, zero
  real advisories); aarch64 cross-build exit-green (ELF-byte verify owed
  at final gates per AGENTS).
- **T23 harness WORKING**: `scripts/ui-capture.py` — boots a fresh
  binary (config-file boot with the fronted CSRF shape), seeds two
  threads + two snippets + a contact over HTTP (multipart sends,
  meta-CSRF → /api/session → rotated token), injects the session cookie
  into chromium (the island's own login gates on the SIP WS a bare boot
  cannot serve — caddy owns /sip on the stack), pins themes via
  `localStorage.wp-theme`, captures the **14-shot matrix** (7 surfaces ×
  light/dark) to `ui-shots/`. Selenium + nix chromium/chromedriver (the
  stack E2E's driver family). Verified end-to-end twice; shots differ
  per tab. Budget decision (§f28): LOCAL-ONLY, not a flake check —
  headless chromium cannot run in a nix sandbox without a KVM VM, and a
  KVM browser VM here would duplicate the stack's lane at double cost.

## b) Partially done

~~- **T23**: the shots are mechanically good but NOT eyeballed (no image~~ done — the visual pass eyeballed the shots; the DOM assertion landed (AGENTS: asserts DOM per shot)
viewing in this session) and the harness asserts nothing about DOM
content at capture time — a signed-in-content assertion (thread rows
present) would make it self-verifying. AGENTS note not yet added.
~~- **T22.4** (local-behind-remote divergence): observed live this session~~ done — the divergence is documented in the TODO health row (LIVE daemon behavior)
(see d) but the note is only here, not yet in the harvest docs.
~~- **T21.6** (exceptions record): the exceptions exist across~~ done — the batch-2 evidence lives here + the archived 05:56 report; the TODO row closed at harvest
TODO_LIST row 158 + AGENTS + the 05:56 report; the TODO row still
reads PLANNED and needs flipping at harvest with the actionlint/VM
evidence attached.

## c) Not started (deliberately or pending)

~~- **T25.2 carve trigger-gate**: nothing fired this session (zero new~~ record stands — the carve trigger never fired; the D2/D3 decision text stands in the design note
internal/server files; D2 honored). The standing decision text is in
the design note; nothing to do until the trigger fires.
~~- **HARVEST** (TODO_LIST/CHANGELOG/FEATURES folds incl. T11–T19) —~~ done — the folds landed (CHANGELOG/FEATURES/TODO)
next session's first block.
~~- **Final gates**: `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`,~~ resolved by events — the 12:57 closeout battery
quiet-host full `nix flake check` (KVM backup VM included), fresh
smoke (`--expect-version` + the one-boot eyeball list), `git ls-remote
  origin main` == HEAD assert.
~~- Owner legs untouched as always (T01/T02/T04/T09/T24/T26/T27).~~ routed — TODO owner rows (handover legs)

## d) Fucked up / went sideways (all recovered)

~~- **Interrupted tool call rolled back five file edits** (calls.js,~~ process record — the mtime-guard lesson re-confirmed
phone.templ, main.js, i18n.js, selftest.js, the A6 connection tests):
detected via selftest module-not-found, re-applied everything
atomically, re-verified 164/164 + Go green. Lesson re-confirmed: the
edit-tool mtime guard is real; re-verify survivors before rebuilding
on top.
~~- **The auto-commit daemon REWROTE git history mid-session**: the~~ process record — the T22.4 observation, live; documented in the TODO health row
morning's named commit `a00dadc` (T18 seams) vanished from the log,
replaced by a fresh auto-commit chain (same content, new hashes).
All work survived under the new hashes (verified file-by-file);
nothing was lost. This is T22.4's divergence observation, live.
~~- **health.css byte-drift gate rabbit hole** (three failed attempts,~~ process record — the canary rationale is documented in the script comment
then right-sized): (1) naive rebuild-diff caught REAL drift → good;
(2) root-caused a self-feedback loop — the artifact sits inside a
Tailwind @source root, so in-place rebuilds re-ingest their own
output (zombie classes survived three rebuilds); fixed the script to
a staged, locked-nixpkgs build; (3) the drift STILL failed because
`vendor/` is UNTRACKED — the flake source never contains the scan
roots, so a from-${self} rebuild can never equal a vendor-built
artifact. Byte-drift gate = architecturally impossible here; the
check is the CANARY (dark-variant fingerprint) and the script owns
the honest rebuild. The rebuilt artifact was genuinely stale (see a).
~~- **ui-capture needed seven iterations**: multipart send prologue,~~ process record — the seven-iteration recipe (cookie injection, config-file boot) is the reusable one
meta-CSRF for /api/session, browser Origin rejected without
`csrf.trusted_*` (config-file boot now), `wss://` hardcoded in
config.js (bare-HTTP boots could NEVER connect — fixed, product bug),
`form.submit()` skips the submit event (island listener), and the
island login gating on the unservable SIP WS (cookie injection wins).
One syntax error of my own making (`len('…'` unclosed) caught by
python before anything ran.

## e) Improvements I'm proud of / would keep

- The HX-Target head-render pattern: one handler, two response shapes,
  honesty without page yank — worth reusing for future head-level
  toggles.
- `TestNoUnusedDictionaryKeys` + `TestStylesUseLogicalProperties`: two
  cheap guards that pay rent forever (found 2 dead keys day one).
- The staged-build script comment (feedback loop + vendor invisibility)
  documents WHY, so nobody re-attempts the byte gate.
- The capture harness's cookie-injection insight (session ≠ SIP) is the
  reusable recipe for any future browser-against-bare-boot testing.

## f) Next up (execution order)

~~1. fmt + full re-verify of the latest edits (config.js, health.css,~~ done — the suites re-verified green (this session's close + the 12:57 battery)
ui-capture, checks.nix): island suite + oxlint + `go test ./...`.
~~2. Eyeball `ui-shots/` (owner or a session with image viewing); add the~~ done — the visual pass eyeballed the shots; the DOM assertion landed (AGENTS: asserts DOM per shot)
signed-in-content DOM assertion to ui-capture (thread rows present).
~~3. AGENTS.md: one durable line for the visual harness (local-only, the~~ done — the AGENTS Commands block documents the harness (LOCAL-ONLY + budget rationale)
budget rationale, how to run).
~~4. T22.6 default note: fold T11–T19 into the v2.9.0 changelog at~~ routed — TODO health row (the v2.9.0 fold decision, owner §g2)
harvest unless the owner splits.
~~5. HARVEST: TODO_LIST (flip the nix-review batch-2 row → DONE with~~ done — the harvest landed (TODO flips, the M21.6 NO-GO in CHANGELOG, FEATURES flips)
actionlint/drill/deadnix evidence; flip the schema_version row →
DONE pointing at the T18 migration runner; add the M21.6 NO-GO row
naming the gateway-seam dependency; add the visual-harness row),
CHANGELOG Unreleased (T11–T19 + the health.css artifact fix + the
config.js ws fix + the deadnix vendor exclusion), FEATURES.md flips.
~~6. Session-3/4/5/6/7 §f folds into TODO_LIST (the standing harvest).~~ done — the TODO state + this v6 sweep
~~7. Final gates: `BUILDFLOW_NO_RESULT_CACHE=1 ./scripts/buildflow.sh`~~ resolved by events — the 12:57 closeout battery ran the final gates
(gomod-check vendor false-positive = documented exception; run inside
`nix develop`), quiet-host full `nix flake check`, fresh-binary
smoke with the one-boot eyeball list (fax timeline/resend, voicemail
player, filter-aware empty, Retry/Dismiss, welcome dismiss+compact,
archived toggle, snippet chips, service dots), `git ls-remote origin
   main` == HEAD.
~~8. Post-gates: report + wait (owner §g rulings outstanding: D7 archive~~ process record — the §g rulings stood at their defaults (D7, M21.6 NO-GO, carve staging, C4, double-fetch)
semantics, M21.6 NO-GO ratification, carve staging, C4 fax-thumbnail
decline, voicemail waveform double-fetch).

## g) Questions for the owner (cannot figure these out myself)

~~1. **Shot review**: `ui-shots/` (14 PNGs) needs a human eyeball — is the~~ done — the shots were eyeballed at the T23 visual pass (per-train; no pixel-goldens, as recommended)
visual matrix what you want persisted per release, or per-train
only? (No pixel-goldens proposed; timestamps make byte-goldens flake.)
~~2. **v2.9.0 fold**: T11–T19 is a large changelog block — one release or~~ routed — TODO health row (the fold decision, owner §g2)
split (my default: one).
~~3. **The history rewrite**: the auto-commit daemon replaced a named~~ done — the daemon divergence is documented (TODO: never verify via push logs, only git ls-remote)
commit mid-session (same content, new hashes). Harmless here, but if
you rely on named-commit archaeology the daemon's heuristic may be
fighting you — worth a look when convenient.
