# UI Redesign — Verification & Polish Session (continuation)

**When:** 2026-09-20 13:22 CEST
**Scope:** This session continues the screenshot-driven UI redesign reported in
`2026-09-20_12-14_ui-redesign-screenshot-review.md`. That report covered the
diagnosis + full CSS/templ rewrite; THIS session is the verification round
(reviewing the 12:24 reshoot) plus the queued polish items. Session-scoped per
instruction; carried-over state is marked as such.

> Format note: the status-report skill's canonical output is a styled HTML
> dashboard; the explicit `.md` instruction in the request wins here. If this
> report should be re-rendered as an HTML dashboard later, the sections below
> map 1:1 onto the skill's components.

---

## Context: where the redesign stands overall

The v2 shell redesign (token-system `app.css` rewrite, island stylesheet sync,
avatar/dir-chip/welcome-panel templ changes) is implemented, test-green, and
visually verified across 3 screenshot rounds (36 shots each, light/dark ×
desktop/mobile) against a local loopback harness with seeded data and an
in-browser fake-SIP WebSocket. The 12:24 reshoot contained the last two fixes
from the prior round: avatar country signum and history dir-chips.

## a) FULLY DONE

1. **Reshoot verified, not assumed.** Confirmed the server on :18099 serves the
   NEW build before trusting any screenshot: fetched `/assets/app.css` over
   HTTP and grepped for the marker strings (`operator desk`, `wp-avatar`,
   `wp-dir-chip`) — all present.
2. **Avatar country signum confirmed visually.** Messages thread rows show
   `+1` (US) and `+4` (UK/DE) signa in the avatar circles — the fix that
   replaced the broken `14`-style digits works.
3. **Island dark theme confirmed visually.** Keycap keypad with letter
   sub-labels and inset highlight, tightened topbar, mono event log, dial input
   with Call button — coherent, no box-in-box.
4. **Contacts tab confirmed visually.** Word-initial avatars (TS/OT/MW/AK)
   with deterministic hues, name+number stacked, quiet ghost Call/Remove
   buttons, clean `wp-row` grid.
5. **Stray green dot identified.** The small dot in the bottom-right corner of
   EVERY screenshot is `#wp-sse-live` — the deliberate SSE-connected indicator
   from session.js (green = feed live). Not a bug; no action taken.
6. **Polish fix 1 — contacts import row grouping** (`internal/web/assets/
   app.css`): the file input's `flex-basis: 100%` forced "Choose File" onto its
   own full-width row with the Import button orphaned below. Now
   `flex: 1 1 220px; min-width: 0` plus `.wp-compose form.wp-compose { flex: 1
   1 auto; min-width: 0 }` — file input, Import, and Export vCard share one
   row. Code landed (daemon-committed).
7. **Polish fix 2 — island dial placeholder truncation** (`internal/web/assets/
   island/app/i18n.js`): en `Number, e.g. 1001, 2000, +441632960961` →
   `Number or extension`; de → `Nummer oder Durchwahl`. Verified this string is
   NOT among the stack E2E's greppable contract strings, so shortening is safe.
8. **Unit tests for the avatar helpers** (`internal/web/views/helpers_test.go`,
   new): country signum, name initials (incl. unicode + multi-space), blank
   fallback, hue determinism + 0–359 bounds.
9. **Test-found real bug fixed** (`internal/web/views/helpers.go`): the test
   round caught `avatarFor("   ")` returning `""` (empty avatar glyph) instead
   of `"?"` — whitespace-only names rendered an EMPTY avatar everywhere.
   Fixed at the root with `strings.TrimSpace` before the blank check.
10. **Views package tests + gofmt green** in the devShell
    (`GOEXPERIMENT=jsonv2 go test ./internal/web/views/`, `gofmt -l` clean).

## b) PARTIALLY DONE

1. ~~**History dir-chip arrows — implemented, NOT visually verified.** The~~ done (done (16:28 session verified the chips; render pins shipped))
   ~~loopback harness has no CDR rows: the History tab honestly shows "Needs the~~
   ~~operator phone API (phone_api_url)", so there is nothing to look at. The~~
   ~~`cdrDirGlyph`/`faxDirGlyph` helpers and `.wp-dir-chip` CSS are in and the~~
   ~~pattern is identical to verified fax rows, but no human eyes have seen a~~
   ~~rendered CDR dir-chip yet. Needs a render test, a seeded CDR path, or the~~
   ~~stack E2E.~~
2. ~~**UI redesign overall (~95%).** Implementation + local visual verification~~ done (resolved: E2E green ×2 on the v2.5.0 chain 2026-09-22 (registered state included))
   ~~done in both themes; what is missing is (i) registered-state sign-off~~
   ~~against the REAL PBX (fake SIP only proves transport, pill stays~~
   ~~"offline"), and (ii) the stack browser E2E re-run required after markup~~
   ~~changes.~~
3. ~~**Commit hygiene.** The auto-commit daemon swept this session's CSS + i18n~~ done (superseded: daemon + narrative-commit convention; helpers committed)
   ~~polish into anonymous "heuristic" commits; `helpers.go` (TrimSpace fix) and~~
   ~~the new `helpers_test.go` are still uncommitted right now. The runbook's~~
   ~~"explicit commit per task" discipline was not followed this session.~~
4. ~~**This session's two polish fixes are NOT re-screenshotted yet** — the~~ done (done (16:28 verified the fixes in both themes))
   ~~import row and the placeholder changes are code-verified only. Pattern says~~
   ~~screenshot after every visual edit; that round is pending.~~

## c) NOT STARTED

1. ~~AGENTS.md knowledge write-back (see e/f — planned item list is defined).~~ done (done (AGENTS facts written))
2. ~~Full gates: `go test -count=1 ./...` (whole repo), `webphone-smoke.py`,~~ done (done (gates green across trains))
   ~~`BUILDFLOW_NO_RESULT_CACHE=1 buildflow`, `nix flake check`, vulnix.~~
3. ~~Stack browser E2E (`.#telephony-browser`) + stack webphone VM test — the~~ done (done (E2E ×2 on the release chain))
   ~~REQUIRED island regression gate after markup/CSS changes.~~
4. ~~Explicit per-task commits for this session's units.~~ done (superseded: daemon + narrative commits)
5. ~~CHANGELOG/FEATURES fold for the redesign (Unreleased section).~~ done (done (v2.5.0 fold))
6. ~~Live PBX registered-state screenshot round (blocked on credentials).~~ done (superseded: registered state proven via the stack E2E)
7. ~~Release decision + execution (v2.5.0 train vs hold) — blocked on user.~~ done (done (v2.5.0 released 2026-09-22))
8. ~~TODO_LIST harvest of the 12:14 report's 50-item backlog (docs-health).~~ done (done (docs-health sweeps 2026-09-20/22))
9. ~~The three open questions (below) — asked twice now, still unanswered.~~ done (resolved: harness decisions superseded by the E2E sign-off; identity verified)

## d) TOTALLY FUCKED UP

Nothing destroyed or broken beyond repair. Honest warts, worst first:

1. **The whitespace-avatar bug shipped in the prior round.** `avatarFor` went
   in WITHOUT tests and carried a real defect (whitespace-only input → empty
   glyph). It only surfaced because tests were written this session. The
   project rule "every change raises the bar" was violated by pairing a new
   helper with zero tests.
2. **Per-task commits were skipped, and the daemon blurred the history.** The
   CSS polish, i18n change, and (soon) helper fix are now spread across
   anonymous heuristic commits. No data lost, but the release-notes and
   archaeology value of history is degraded for this session's units.
3. **Visual fixes slipped rounds.** Dir-chips have now gone two rounds without
   visual confirmation because the harness cannot render CDR rows — the
   harness gap should have been closed the moment the fix landed.
4. **Standing noise, not mine but unignored:** gopls/golangci-lint have been
   red across the project ALL session (go.mod needs go ≥ 1.27.1, LSP runs
   1.26.7 with GOTOOLCHAIN=local). Every file view carries phantom errors.
   Gates are only real inside `nix develop`.

## e) WHAT WE SHOULD IMPROVE

1. **Tests ship WITH the helper, same commit.** The whitespace bug is the
   proof; "I'll add tests later" means shipping bugs.
2. **Verify visual changes in the same round that makes them.** Two queued
   polish items sat unverified across a session pause; the screenshot loop
   must close before the task is called done.
3. **Commit per task immediately.** The daemon's heuristic commits are a
   safety net, not a history strategy.
4. **Close harness gaps as first-class tasks.** "Cannot visually verify X"
   should spawn "make X renderable locally", not linger.
5. **Don't trust background shells across session pauses.** Re-verify
   artifacts on disk (HTTP greps saved this round).
6. **Keep the screenshot harness reproducible.** It lives in `/tmp/wpshoot/`
   (throwaway); promoting it to `scripts/` would make every future UI change
   verifiable in one command.
7. **Ask blocking questions earlier.** The same 3 questions have now blocked
   sign-off twice.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Brainstorm, roughly impact-ordered; most items beyond ~15 are ROADMAP fuel
(docs-health HARVEST should apply routing rigor).

1. ~~Commit `helpers.go` fix + `helpers_test.go` explicitly (per-task).~~ done (superseded: daemon + narrative commits)
2. ~~Run full `GOEXPERIMENT=jsonv2 go test -count=1 ./...`.~~ done (superseded: GOEXPERIMENT removed 2026-09-22 (T15b); suite green)
3. ~~Run `python3 scripts/webphone-smoke.py` (28-check live smoke).~~ done (done (smoke runs every train))
4. ~~Run `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`.~~ done (done (buildflow no-cache runs))
5. ~~Run `nix flake check` (includes island-lint + treefmt).~~ done (done (flake check ALL PASS 2026-09-22))
6. ~~Re-shoot the harness to visually confirm import row + placeholder fixes.~~ done (done (16:28 reshoot review))
7. ~~Mobile-width check of the contacts import row (shrink/overflow behavior).~~ done (ROADMAP-fuel (mobile polish))
8. ~~Stack browser E2E re-run (`nix build -L .#telephony-browser`,~~ done (done (E2E ×2 on the release chain))
   ~~`--override-input webphone` in the stack) — REQUIRED after markup changes.~~
9. ~~Stack webphone VM test (`checks.x86_64-linux.telephony-webphone`).~~ done (done (VM test green every train))
10. ~~aarch64 cross-build + verify ELF machine bytes (never trust exit code).~~ done (superseded: release.sh ELF guard)
11. ~~CDR render test: pin `CDRRow` markup incl. dir-chip glyph + aria-label.~~ done (done (CDR render pins: dir-chip glyph + aria-label tests))
12. ~~Extend the harness/seed so History renders CDR rows locally (fake~~ done (covered by the E2E history drill)
    ~~phone-api or seeded CDR path) — closes the dir-chip verification gap.~~
13. ~~AGENTS.md write-back: app.css owns `.sr-only` (no Tailwind loaded;~~ done (done (AGENTS facts written))
    ~~templ-components Base emits Tailwind classes), `wp-fax-row` grep contract,~~
    ~~dir-chip pattern, token mirror rule app.css ↔ island/style.css, avatar~~
    ~~helper semantics, `#wp-sse-live` indicator note.~~
14. ~~Live PBX registered-state screenshots (needs credentials — see g1).~~ done (superseded: registered state via the stack E2E)
15. ~~Visually verify the parallel session's fax Identity block (g3).~~ done (done (identity surfaces verified + pinned))
16. ~~Fold redesign into CHANGELOG `Unreleased` (release runbook step 1).~~ done (done (v2.5.0 fold))
17. ~~Update FEATURES.md UI inventory (welcome panel, avatars, dir chips,~~ done (done (FEATURES UI rows current))
    ~~morph-swap surfaces, import row).~~
18. ~~Harvest the 12:14 report's 50-item backlog into TODO_LIST (docs-health).~~ done (done (HARVEST sweeps))
19. ~~Decide + execute release train (v2.5.0: version bump → gates → tag →~~ done (done (v2.5.0 released))
    ~~push → lychee → stack bump → stack gates — see g2).~~
20. ~~vulnix runtime-closure re-run (`nix run .#vulnix`).~~ done (vulnix rides release.sh)
21. ~~Island i18n sync check: the Go en/de test does not cover~~ done (done (island i18n parity node test))
    ~~`island/app/i18n.js`; add a cheap parity check (grep test or node assert).~~
22. ~~Verify prettier/treefmt pass on the edited island JS (`nix fmt`).~~ done (done (treefmt owns island assets))
23. ~~Promote `/tmp/wpshoot` harness to `scripts/` (or document as throwaway).~~ **Won't implement — harness stayed throwaway (documented).**
24. ~~Shoot the DE locale round (`wp-lang` cookie) — German copy never~~ done (ROADMAP-fuel (de screenshot pass))
    ~~visually verified in this effort.~~
25. ~~Re-verify the welcome panel (`wp-welcome`) after final CSS settles.~~ done (verified in review rounds)
26. ~~Theme cycle test: light/dark/auto toggle visually (all three states).~~ done (verified (theme cycle tested; island-tests cover theme preload behavior))
27. ~~Keyboard-focus shot of the skip link (`sr-only:focus` reveal).~~ done (skip-link focus verified (sr-only rules + review))
28. ~~Run SSE fragment test explicitly to confirm `wp-fax-row` grep contract~~ done (wp-fax-row contract holds (SSE tests))
    ~~still holds alongside the Identity changes.~~
29. ~~Confirm voicemail stable-id contract (`vm-<uuid>`, `vm-audio-<uuid>`)~~ done (TestVoicemailRowsCarryStableMorphIds green)
    ~~survives — run `TestVoicemailRowsCarryStableMorphIds`.~~
30. ~~Confirm the 35-id DOM contract test (`TestServedPageHoldsTheDomContract`).~~ done (DOM contract single-sourced (T22) and green)
31. ~~Fax tab visual pass with thumbnails/previews (carried backlog).~~ done (ROADMAP-fuel (fax previews))
32. ~~Thread list date-group separators (carried backlog).~~ done (ROADMAP-fuel)
33. ~~Contacts: personal-vs-shared section headers + sort order (carried).~~ done (ROADMAP-fuel)
34. ~~Incoming-call modal styling pass (harness cannot trigger inbound; needs~~ done (covered by the E2E call drills)
    ~~stack E2E footage or a fake-inbound hook).~~
35. ~~Transfer dialog styling pass (needs live call).~~ done (covered by the E2E transfer drill)
36. ~~Toast styling verification (trigger an error toast in harness).~~ done (toast styling pinned by node tests (kind classes))
37. ~~`#wp-sse-live`: consider hiding pre-login (dot shows on the logged-out~~ done (verified (pill shows post-connect only))
    ~~shell where SSE is not connected anyway).~~
38. ~~History empty state: richer guidance (link to Settings/phone_api docs).~~ done (ROADMAP-fuel (empty states))
39. ~~Keypad press-state/DTMF visual check during a real call.~~ done (covered by the E2E DTMF drill)
40. ~~Session-expiry UX: expired-session toast + focus back to login.~~ done (superseded: dead-session class deleted (SQLite sessions 2.5.0 + resume))
41. ~~app.css size budget note (~19 KB — fine; keep an eye as it grows).~~ **Won't implement — size budget note kept informal.**
42. ~~gitleaks + codespell on-demand runs (`buildflow -s gitleaks -s codespell`).~~ done (done (gitleaks/codespell default))
43. ~~Annotate the 12:14 report as superseded by this one (docs-health~~ done (done (this sweep classifies + annotates it))
    ~~ANNOTATE, non-destructive).~~
44. ~~Stack repo: bump the `webphone` input after the push (coordinate with~~ done (stack rides main per-train (DECIDED))
    ~~release decision).~~
45. ~~CSP hash re-check cadence: only on templ-components bumps (no-op now,~~ done (superseded: CSP hash dropped entirely (NoThemeScript, 2026-09-22))
    ~~keep on radar).~~
46. ~~Consider a tiny `make-shoot` wrapper (build → pkill -9 → boot → seed →~~ **Won't implement — wrapper not adopted; ritual documented.**
    ~~shoot) so the 6-step ritual is one command.~~
47. ~~Verify light/dark avatar hue contrast (oklch tints on both themes).~~ done (verified (avatar hues in both themes))
48. ~~Check the morph surfaces still look right after CSS changes (draft text~~ done (morph surfaces pinned (stable ids + E2E))
    ~~- focus preservation on live pushes).~~
49. ~~Post-gates: deliver before/after screenshot set to the user for sign-off.~~ done (superseded: E2E sign-off accepted)
50. ~~Update memory/AGENTS.md with the confirmed harness recipe (chromium path,~~ **Won't implement — harness recipe lives in the 16:28 report only.**
    ~~CDP 9333, PUT /json/new retry) once promoted.~~

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Live PBX credentials** — an extension+password on pbx.artmann.tech (or a
   throwaway test extension) so I can shoot REAL registered-state screenshots;
   or do you accept the stack browser E2E green as registered-state sign-off
   and skip live shots?
2. **Release train** — single v2.5.0 release carrying the whole redesign
   (full runbook: fold → bump → gates → tag → stack bump → stack gates), or
   explicit per-task commits only and hold the release for later?
3. **Parallel session's fax "Identity" feature** — still in flux (I keep
   hands fully off), or settled enough for me to visually verify, test, and
   fold it into the same release?

---

**Awaiting instructions.** Nothing else was started or changed after writing
this report.
