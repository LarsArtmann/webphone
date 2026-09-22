# Status Report — Full Gate Sweep & E2E Flake Verdict Session

**When:** 2026-09-20, 16:28 CEST
**Scope:** THIS session only (the continuation after
`2026-09-20_13-22_ui-redesign-verification-and-polish-session.md`; that
report's scope is NOT re-researched here). Session window ≈ 14:30–16:28 CEST.
**Context:** Resumed from "WAIT FOR INSTRUCTIONS" with the command to execute
the remaining plan (AGENTS.md write-back, gates, reshoot, E2E) until done.
No release authorization was given; the three open questions from 13:22 are
still unanswered.

---

## a) FULLY DONE

1. **Session-state verification.** Git tree was CLEAN at session start;
   `helpers.go` (TrimSpace fix) + `helpers_test.go` confirmed committed by
   the daemon (commit `90084e1`). Suspected origin/local anomaly
   (`ls-remote` showed a different HEAD) investigated with `git fetch`:
   local was 16 daemon commits AHEAD, origin had nothing extra — no
   foreign history, no action needed.
2. **Fact research for AGENTS.md.** Every planned AGENTS.md claim verified
   against the tree before writing: `.sr-only` block (app.css:108–124 with
   the "no Tailwind loaded" comment), `#wp-sse-live` pill (session.js:125+,
   JS-created, `aria-hidden`, `data-live` toggling), SSE grep contract
   (`TestSSEPushesSwapSafeFragments` greps `wp-thread-row`/`wp-bubble`/
   `wp-fax-row`), dir-chip helpers (`cdrDirGlyph`/`cdrDirLabel` in
   history.templ, `faxDirGlyph`/`faxDirLabel` in fax.templ, "↓"/"↑",
   aria-hidden chip + `role="img"` sibling), `avatarFor` TrimSpace
   implementation, token blocks in both CSS files, dial placeholder strings
   (en line 23 / de line 104 in i18n.js), import-row flex rules.
3. **AGENTS.md write-back.** One new bullet `UI redesign invariants
   (2026-09-20)` (token mirror rule app.css ↔ island/style.css, `.sr-only`
   ownership, `avatarFor`/`avatarHue` semantics + "helpers ship WITH tests"
   lesson, dir-chip pattern + the loopback renderability gap, SSE row-class
   grep contract, `#wp-sse-live` green-dot note, import-row and dial-
   placeholder polish) and one new bullet `Stack browser E2E flake mode
   (2026-09-20)` (see a7). All claims evidence-checked first.
4. **templ sync check.** `nix develop -c templ generate ./internal/web/views/`
   → 0 updates; committed `*_templ.go` are in sync. No .templ edits existed.
5. **Full Go test suite.** `GOEXPERIMENT=jsonv2 go test -count=1 ./...` in
   devShell: every package ok, including `internal/web/views` with the four
   new helper tests.
6. **Smoke test.** `scripts/webphone-smoke.py` in devShell: **32/32 passed**
   on a fresh boot + temp data dir (the script's current total; AGENTS.md
   already says 32-check on disk — my "28" memory was a stale-context false
   alarm, caught by checking the file before "fixing" it).
7. **Stack browser E2E: GREEN.** `nix build -L .#telephony-browser
   --override-input webphone /home/lars/projects/webphone` — run 2 exited 0
   with the FULL flow in both call directions (2× call established, ICE
   panel, blind transfer initiated, caller released). This is the island
   regression gate passing over the ENTIRE redesign + morph + polish +
   parallel-session fax Identity tree.
8. **E2E run-1 failure root-caused and documented.** Run 1 died at the
   transfer step: FreeSWITCH hung the call with `RECOVERY_ON_TIMER_EXPIRE`
   almost exactly 90s after DTLS-ready (13:42:03.8 → 13:43:33.4), the
   island removed the dead call card, and the E2E's fresh `.transfer-btn`
   click went `StaleElementReference`. Verified NOT a webphone regression:
   island call-path JS (calls/ui/connection/ice) is absent from the
   `6bb792e..HEAD` diff; registration/DTMF/ICE-stats were all green before
   the drop. Verdict rule written into AGENTS.md: green-until-~90s = flake,
   re-run once before digging.
9. **Screenshot reshoot (36 shots).** Fresh binary built in devShell, stale
   server `pkill -9`-ed, clean data dir (`data-r2`), reseeded, full
   light/dark × desktop/mobile sweep into `/tmp/wpshoot/local/`.
10. **Both queued polish fixes VISUALLY CONFIRMED** (desktop, light + dark):
    contacts import row on ONE line (Choose File + Import + Export vCard;
    the `flex: 1 1 220px` fix), island dial placeholder "Number or
    extension" live next to the Call button. Avatar signums (+1/+4), word
    initials (TS/OT/MW/AK), dark keycap keypad, token-driven theming all
    re-verified in passing.
11. **buildflow full: exit 0.** `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` in
    devShell. The visible ruff warnings (PLW1510 in `scripts/*.py`) are
    pre-existing script-level policy notes, warning-severity, not mine.
12. **nix flake check: exit 0.** Explicitly re-run to capture the exit code;
    KVM present, the current tree's `webphone-backup` VM-test derivation is
    realized (dry-run: nothing to build), island-lint/treefmt/island-js
    checks green. aarch64 omission is the documented `nix flake check`
    limitation, handled at release time.
13. **Closing sweep.** Harness server + chromium proven dead
    (`pgrep || echo dead` both); no manual commits (harness rule — the
    daemon owns commits).

## b) PARTIALLY DONE

1. ~~**Visual verification breadth.** What works: both polish fixes confirmed~~ done (accepted; later review rounds covered the variants)
   ~~on desktop light + dark (4 shots viewed). What remains: the 4 MOBILE~~
   ~~variants of the fixed pages are on disk but were never VIEWED; this~~
   ~~round's `Island-Registered` shots carry a WARN ("island never~~
   ~~registered" — fake-sip race) and were not viewed; registered-state~~
   ~~proof rests on `Island-Revealed` (which did show a signed-in,~~
   ~~connected island). Blocker: none — ~10 minutes of viewing. Effort: S.~~
2. ~~**E2E green state.** What works: run 2 exit 0, full flow both directions.~~ done (resolved: the ~90s transfer-step death is the documented re-run-once flake mode (AGENTS); dest-clear fix removed the stall class)
   ~~What remains: the exact FreeSWITCH knob behind the ~90s ceiling is~~
   ~~unidentified (no explicit timer found in the stack config; "cold caches~~
   ~~made run 1 slower" is timestamp inference, not a controlled comparison).~~
   ~~Flake rate unquantified (n=1 fail, n=1 pass). Blocker: needs stack-side~~
   ~~FS config archaeology. Effort: M.~~
3. ~~**AGENTS.md persistence.** What works: both bullets written and verified~~ done (done (narrative commits at phase boundaries))
   ~~on disk. What remains: uncommitted at report time; the daemon will fold~~
   ~~them into a heuristic auto-commit, blurring the docs unit. Blocker:~~
   ~~explicit commits need user authorization. Effort: S.~~
4. ~~**History dir-chip rendering.** Carried from 13:22 unchanged: chips are~~ done (superseded: E2E renders CDR chips in the history drill)
   ~~test-verified (`cdrDirGlyph`/`faxDirGlyph` greps + unit tests) but have~~
   ~~NEVER rendered in a screenshot — loopback has no phone API, so history~~
   ~~renders its empty-state. Blocker: needs a fake phone-api in the harness~~
   ~~or live PBX access. Effort: M.~~
5. ~~**The 13:22 report's three questions.** Still unanswered; this session~~ done (resolved: owner batch consolidated in the TODO owner-batch row)
   ~~treated "keep going until done" as authorization for non-release work~~
   ~~only. Release-related work stayed untouched.~~

## c) NOT STARTED

1. ~~**Release train v2.5.0** — fold CHANGELOG/FEATURES/TODO_LIST/ROADMAP,~~ done (done (v2.5.0 released 2026-09-22))
   ~~bump `webphoneVersion`, full gates, tag+push, lychee, stack lock bump,~~
   ~~stack gates, aarch64 cross-builds (+ ELF-byte proof), vulnix, closing~~
   ~~sweep. Waiting on: your release decision. All gates this session are~~
   ~~green, so the train is ready when you say go.~~
2. ~~**TODO_LIST harvest** of this report's section f (+ the 13:22 backlog).~~ done (done (HARVEST sweeps 2026-09-20/22))
   ~~Waiting on: authorization + the release decision (it changes priorities).~~
3. ~~**Live-PBX registered-state screenshots.** Waiting on: credentials, or~~ done (superseded: registered state via the stack E2E)
   ~~your acceptance of the green E2E as sign-off.~~
4. ~~**Fake phone-api harness mode** (unblocks b4). Planned, no code.~~ done (superseded: E2E history drill renders chips)
5. ~~**Stack re-pin** to a release tag (currently rides v2.4.0 `a59f0d1`).~~ done (stack rides main per-train (DECIDED))
   ~~Waiting on: release.~~
6. ~~**Fax Identity (parallel session) verification/fold.** Hands-off was~~ done (identity surfaces verified + pinned (TestIdentitySurfacesOwnNumber))
   ~~respected; note it rode through the green E2E untouched. Waiting on:~~
   ~~your answer on whether that session is still active.~~

## d) TOTALLY FUCKED UP

Nothing was destroyed, reverted, or left broken — every gate is green and
the closing sweep proved the harness dead. The honest list is process
failures that cost time or produced misleading evidence:

1. **buildflow verdict swallowed by a pipe.** I ran `buildflow … | tail -30`
   in a background shell: the exit code AND the verdict banner were cut, so
   a full gate result had to be recovered via a second (cached) run.
   Severity: wasted ~minutes + a false "did it pass?" ambiguity. Root
   cause: gating through a filter without keeping the status. Mitigation:
   capture exit code first (`> log; echo EXIT=$?`), pretty-tail second —
   applied for every later gate this session.
2. **First E2E launched on cold caches with no flake guard.** Run 1 failed
   at the transfer step, costing ~10 minutes of diagnosis before the
   correct flake verdict. Severity: gate noise, no damage — and it
   produced the now-documented flake mode (silver lining). Root cause:
   known passing run was 148s; I didn't anticipate cold-cache slowness
   pushing the transfer step past the ~90s FS ceiling. Mitigation now
   exists: the AGENTS.md verdict rule + retry-once.
3. **`Island-Registered` WARN ignored instead of fixed.** The shooter
   warned "island never registered" on the registered-state pages; I
   worked around it by viewing `Island-Revealed` instead of investigating
   the fake-sip race. Severity: those specific screenshots may
   misrepresent the registered state; the harness can silently produce
   lying shots. Root cause: unknown — needs investigation (likely a
   login/WS-stub ordering race in the spec).
4. **`PIPESTATUS` capture came back empty** on the first flake check,
   costing one extra verification round. Root cause: shell quirk under the
   tool's interpreter; fixed by re-running with `> log; echo EXIT=$?`.
5. **Sloppy research grep:** used `rg -rn` (`-r` is REPLACE) once, got a
   misleading empty result before re-running correctly. Harmless, sloppy.
6. **Carried from last session (bookkeeping):** the todo list had claimed
   "AGENTS.md update in_progress" when in fact it had not been started —
   the summary carried it correctly, but the stale status itself was an
   error. Fixed this session.

## e) WHAT WE SHOULD IMPROVE

1. **Gate-runner discipline:** never pipe a gate through a filter without
   keeping its exit status. Concrete rule: `cmd > /tmp/log 2>&1;
   echo EXIT=$?` first, inspect the log second. (Cost this session: one
   full buildflow re-run cycle + one flake-check re-run.)
2. **E2E retry policy in the runbook:** add an explicit "on transfer-step
   failure with green registration/DTMF/ICE, re-run once before
   investigating" step 7 note, so the next session doesn't re-derive the
   verdict. Cost: ~10 min per occurrence.
3. **View ALL FOUR variants per fix** (light/dark × desktop/mobile) as a
   hard checklist item after every reshoot. This session viewed 2 of 4 for
   each fix; the mobile shots exist but were skipped.
4. **Harden the harness:** wpshoot should exit non-zero (or annotate the
   PNG filename) when the island fails to register, so screenshots can't
   lie silently; seed should pick a fresh data dir automatically instead
   of a manual `data-r2`.
5. **Fake phone-api in the harness** so history dir-chips finally render —
   the only remaining unverified visual surface of the redesign.
6. **Pin the FS timer:** identify the exact profile param behind the ~90s
   `RECOVERY_ON_TIMER_EXPIRE` (stack-side grep of the FS defaults) and
   record it — upgrades the flake doc from symptom to mechanism, and
   informs whether the E2E or FS config should change.
7. **Explicit commit-per-task when authorized:** the daemon already blurred
   the CSS/i18n/helper units into heuristic commits; my two AGENTS.md
   edits will suffer the same. History readability is a standing cost.
8. **Doc-count drift guard:** counts embedded in prose (e.g. smoke-check
   count) drift; prefer naming the script over the number, or test-assert
   the count. (Today's "28 vs 32" was a stale-context false alarm caught
   before an unnecessary edit.)

## f) Up to 50 things we should get done next

Brainstorm, ranked by impact; HARVEST should route rigorously (TODO_LIST vs
ROADMAP). Impact/Effort/Category per item. Effort: S <30min, M 30min–2h,
L >2h.

**Release & integration**

1. ~~Answer the three open questions (g1–g3) — unblocks 2–9 and the harvest. — Critical/S/Decision~~ done (resolved: owner batch consolidated (TODO owner-batch row))
2. ~~Cut v2.5.0: fold CHANGELOG `Unreleased` → dated section (redesign + morph + polish + helpers). — Critical/S/Release~~ done (done (v2.5.0 released 2026-09-22))
3. ~~Sync FEATURES.md / TODO_LIST.md / ROADMAP.md for the redesign + idiomorph merge. — High/M/Documentation~~ done (done (FEATURES/TODO/ROADMAP current))
4. ~~Bump `webphoneVersion` in flake.nix to the new tag. — Critical/S/Release~~ done (done (version bump per train; drift test))
5. ~~Re-run the four gates fresh (buildflow no-cache, go test, flake check, smoke). — Critical/M/Release~~ done (done (gates green across trains))
6. ~~Tag + push; verify with `git ls-remote` (daemon races pushes). — Critical/S/Release~~ done (done (tags signed + pushed; ls-remote verified))
7. ~~`nix run nixpkgs#lychee -- .` link check after the push. — Medium/S/Release~~ done (lychee rides release.sh)
8. ~~Stack lock bump (`nix flake lock --update-input webphone`) + stack gates (telephony-browser E2E, telephony-webphone VM test, full stack flake check). — Critical/M-L/Release~~ done (done (stack relock ritual per train))
9. ~~aarch64: explicit cross-builds of package + island-lint, verified by ELF machine bytes. — High/M/Release~~ done (done (ELF guard))
10. ~~`nix run .#vulnix` runtime-closure rescan at the new package. — Medium/S/Release~~ done (done (vulnix rides release.sh))
11. ~~Closing sweep per runbook §9 (process-death proofs, post-train buildflow). — Medium/S/Release~~ done (done (runbook §9 closing sweep))
12. ~~Re-pin the stack to the release tag instead of riding main. — Medium/S/Integration~~ done (superseded: ride-main DECIDED)

**Verification gaps (this session's honest leftovers)**
13. ~~View the 4 mobile variants of import-row + dial placeholder (shots already on disk). — High/S/Quality~~ done (accepted (mobile variants covered by later review rounds))
14. ~~Investigate + fix the fake-sip registration race; re-shoot `Island-Registered` cleanly. — High/M/Quality~~ done (superseded: registered state via the E2E)
15. ~~Add a fake phone-api mode to the harness; render + view history dir-chips. — High/M/Quality~~ done (superseded: E2E history drill)
16. ~~Add a server-side render test pinning dir-chip glyphs + aria-labels (independent of visuals). — High/S/Quality~~ done (done (CDR render pins))
17. ~~Identify the FS profile param behind the ~90s timer; update the AGENTS.md flake bullet. — Medium/M/Documentation~~ done (resolved: transfer flake = re-run-once mode documented)
18. ~~Run the E2E 2–3× back-to-back to quantify the flake rate before the release train. — High/M/Quality~~ done (superseded: ×2-green norm satisfied repeatedly; 445s budget)
19. ~~Extend helper tests: whitespace variants (tabs, NBSP) for `avatarFor`. — Low/S/Quality~~ done (avatarFor whitespace variants covered in helpers_test)
20. ~~Visual review of the parallel session's fax Identity panel (light + dark) — unreviewed foreign DOM. — High/S/Quality~~ done (done (identity surfaces verified))
21. ~~Screenshot baseline/goldens + compare script to catch visual drift mechanically. — Medium/M/Quality~~ done (ROADMAP-fuel (golden screenshots consciously deferred))
22. ~~Verify `Island-Revealed` after_js selectors still match the redesigned DOM ids (works today; pin it). — Medium/S/Quality~~ done (done (E2E selectors re-verified on the v2.5.0 chain))
23. ~~a11y pass over the new tokens: contrast, focus rings, sr-only usage inventory. — Medium/M/Quality~~ done (ROADMAP-fuel (a11y cluster))

**Harness/tooling**
24. ~~wpshoot: fail non-zero or filename-annotate on registration WARN (no silent lying shots). — High/S/Quality~~ done (harness stayed throwaway (documented))
25. ~~wpshoot: auto-fresh data dir per round; fold the seed step in. — Medium/S/Quality~~ **Won't implement — throwaway harness.**
26. ~~wpshoot: emit an index.html contact sheet of all 36 shots for fast review. — Medium/S/Quality~~ **Won't implement — throwaway harness.**
27. ~~seed.py: make re-runs idempotent against an existing data dir. — Medium/S/Quality~~ **Won't implement — throwaway harness.**
28. ~~Promote the /tmp harness (wpshoot/seed/fake-sip) into `scripts/` so /tmp cleanup can't kill it. — Medium/S/Cleanup~~ **Won't implement — throwaway harness.**
29. ~~Add the E2E retry-once flake rule to the stack runbook/AGENTS.md step 7. — High/S/Documentation~~ done (done (re-run-once rule in AGENTS/runbook))

**Docs**
30. ~~Explicit docs commit for the two new AGENTS.md bullets (needs authorization; else daemon blur). — Medium/S/Documentation~~ done (superseded: narrative-commit convention)
31. ~~HARVEST this section f into TODO_LIST.md / ROADMAP.md (docs-health). — High/M/Documentation~~ done (done (HARVEST sweeps))
32. ~~ANNOTATE the 13:22 report with this session's answers (docs-health). — Low/S/Documentation~~ done (done (this sweep))
33. ~~README: check whether its screenshots/claims predate the redesign; refresh if so. — Medium/M/Documentation~~ done (done (README current — screenshots not used))
34. ~~Record the "cite the script, not the count" policy for check-count prose. — Low/S/Documentation~~ **Won't implement — count-citation policy informal.**

**Island/UX backlog (carried top items, still valid)**
35. ~~Voicemail rows: playback progress + played/unplayed state polish. — Medium/M/Feature~~ done (ROADMAP-fuel (voicemail polish))
36. ~~Fax compose: drag-and-drop PDF affordance alongside the file input. — Medium/M/Feature~~ done (ROADMAP-fuel (fax affordances))
37. ~~Contacts import: surface partial-failure feedback (which rows failed, why). — Medium/M/Feature~~ done (ROADMAP-fuel (import feedback))
38. ~~Keypad: long-press "0" → "+"; haptic-ish press feedback audit. — Low/S/Feature~~ done (ROADMAP-fuel (keypad polish))
39. ~~Thread list: relative timestamps with absolute time on hover/title. — Low/S/Feature~~ done (ROADMAP-fuel (timestamps on hover))
40. ~~Welcome panel behavior when already registered (is it stale?) — verify + fix if wrong. — Medium/S/Bug~~ done (superseded: welcome behavior verified in later sessions)
41. ~~Event log: copy button + level filter for operator ergonomics. — Low/S/Feature~~ done (ROADMAP-fuel (operator ergonomics))
42. ~~Mobile island: evaluate bottom-sheet call card (current card layout on 412px). — Medium/M/Feature~~ done (ROADMAP-fuel (mobile polish))
43. ~~Theme tri-state (auto/light/dark) persistence check across reloads. — Low/S/Bug~~ done (verified (theme persistence tested))
44. ~~i18n audit: remaining English-only island strings beyond the deliberate `#log`. — Medium/S/Quality~~ done (#log English-only is the deliberate exception (AGENTS))

**Server/robustness (carried)**
45. ~~Rate-limiter rejection counts surfaced in request logs (operational visibility). — Medium/S/Quality~~ done (ROADMAP-fuel (ops visibility))
46. ~~Contract test pinning `/healthz` check names (fleet dashboards depend on them). — Medium/S/Quality~~ done (done (probe-shape tests pin check names))
47. ~~SSE client reconnect backoff review (htmx sse ext defaults) under proxy hiccups. — Low/M/Quality~~ **Won't implement — SSE backoff stays library defaults.**
48. ~~`Retry-After` header on 429 responses. — Low/S/Feature~~ done (done (computed Retry-After ships; 429 surfacing pinned))

**Hygiene**
49. ~~Post-daemon-churn repo check: stale worktrees/branches from the release era. — Low/S/Cleanup~~ **Won't implement — stale worktrees not observed since.**
50. ~~Retire the `ud1` leftover directory in /tmp/wpshoot (unidentified round artifact). — Low/S/Cleanup~~ **Won't implement — /tmp hygiene left to the owner machine.**

## g) Three questions I cannot answer myself

1. **Live-PBX screenshots vs E2E sign-off:** Can you give me a real
   extension + password on pbx.artmann.tech for registered-state
   screenshots (call card, dir-chips in history, in-call UI) — or do you
   accept the green stack browser E2E as final visual sign-off for the
   redesign? (I cannot mint PBX credentials myself; loopback cannot render
   history rows at all.)
2. **Release decision:** Cut v2.5.0 now via the runbook (all gates are
   green; the train includes stack re-pin + aarch64 + vulnix), or hold the
   redesign unreleased? (The stack currently rides v2.4.0; every day on
   `main` widens the eventual lock diff.)
3. **Parallel fax Identity session:** Is that other session still active?
   If it is, I keep hands-off (it already rode through the green E2E); if
   it is done, I should visually review + verify-and-fold its fax
   Identity work this week. (I cannot see other sessions' state.)

Also welcome whenever you decide: authorization for explicit commit-per-task
commits and for the section-f HARVEST into TODO_LIST/ROADMAP.

---

_Point-in-time snapshot. HARVEST candidate: section f. Written per the
status-report skill; Markdown instead of the skill's HTML dashboard because
the user explicitly specified the `.md` path._
