# Status — Master TODO execution train (session 2026-10-01 20:12 CEST)

**Program:** execute the consolidated backlog —
`docs/planning/2026-10-01_17-32_SUPERB-master-todo-pareto-execution-plan.md`
(authorised by the owner's "NOW GET SHIT DONE — the WHOLE TODO LIST").
**Scope actually attempted:** the assistant-executable legs of T03, T05, T06,
T08, T07-partial, T20-partial. Owner-only legs (T01/T02/T04/T09/T24) untouched
by design.
**Repo:** `/home/lars/projects/webphone`, branch `main`. Tree left **clean**
(the auto-commit daemon committed every edit as it landed: `c14f88c`,
`8f68722`, `f18d925`).

---

## a) FULLY DONE (verified)

### T03 — Baseline gates on the UI/UX batch (the row marked 🔴 "ran NO gate")
The 2026-10-01 UI/UX markup had never been gated. It now has been:

- **`buildflow` (full)** via `./scripts/buildflow.sh`: **58 success / 0 failed**,
  `nix flake check` green **including the KVM `webphone-backup` VM test**
  ("all checks passed!"), `nix-build` green, all Go tests green.
  - Exit code **69** = the findings gate, tripped by **`gomod-check` (54) ALONE**
    — the **documented known false positive** in `AGENTS.md` ("verified FALSE
    POSITIVE 2026-10-01 … do NOT hand-edit vendor markers"). No other tool
    contributed gate-level findings.
- **Fresh-binary smoke**: `scripts/webphone-smoke.py --bin $(nix build .#webphone)`
  → **47 passed, 0 failed** + restart scenario **4 passed, 0 failed**.
- Verdict: the served surface is **green**; a deploy is not blocked by a red gate.

### T05 — Island boot language split-brain fix
`internal/web/assets/island/app/i18n.js`: boot order is now
**`wp-lang` cookie → `localStorage` → `navigator.language`** (was localStorage →
navigator, ignoring the cookie the server renders tabs/`<html lang>` from).
New island spec `island-tests/i18n-boot.test.mjs` pins: cookie boots German,
cookie beats localStorage, localStorage leads when no cookie, navigator last,
and `applyI18n()` writes the cookie-derived `<html lang>`.

### T06 — Ring-silence `AudioContext` fix
`audio.js` rewritten to **ONE shared `AudioContext`** created lazily (never at
import) with a new `resumeAudio()`; wired into the gesture handlers in `calls.js`
(`answerIncoming`, `rejectIncoming`, `placeCall`) plus a best-effort resume in
both tone starters. New spec `island-tests/audio.test.mjs` pins lazy creation,
single shared context, and resume-once semantics.
(Note: this row was marked "owner go needed" — see question 1.)

### T08 (partial) + T20 (partial) — hold copy split + pending-clear
- `holdFailed` vs **new `resumeFailed`** copy split per direction, **en + de**
  (both maps, parity test holds).
- `calls.js` clears `holdPending`/`holdQueued` in the `Terminated` branch so a
  pending chip cannot outlive the call (covers the watchdog-rebuild teardown).
- `island-tests/calls.test.mjs` extended: failed-resume names the resume
  direction, failed-hold names the hold direction (new test), and a
  Terminated-while-pending test asserts the state machine settles.

### T07 (partial) — mic pre-warm pins (3 owed)
- `mic.test.mjs`: **dial-after-missed consumes the stale stream** → the outgoing
  call acquires a fresh stream; the missed call's tracks stay released.
- `connection.test.mjs`: **warm survives a watchdog rebuild** (registration-lost
  rebuild leaves the warm mic live) and **second-`onInvite` guard** (second INVITE
  rejected, first banner/session/mic untouched).

---

## b) PARTIALLY DONE

- **Island suite verification is INCOMPLETE.** I ran the full island suite once;
  it surfaced a failure in **my own new test** (wrong assumption — after a failed
  resume the session is still *held*, so the injected click issued a *resume*, not
  a *hold*; the assertion expected `holding`). I fixed the test, but the
  confirming re-run had not returned when this report was requested (two
  `node --test` background shells were still running — see (d)). **The new green
  state is NOT yet witnessed.**
- **T08 owed pins still missing:** `aria-current="page"/"false"` nav-partial pin,
  skeleton reveal/hide on nav swaps, optimistic-bubble-morph-removed edge, the
  `#wp-live` over-promising comment fix, and the release-runbook obligation note.
- **T20 remainder:** the 3 errcheck findings, `nix run .#vulnix` verdict, aarch64
  ELF verify, VM-timeout root-cause, and the visual screenshot pass — all open.
- **T07 live verification:** owner-only (a real call); gates for it reuse T03
  (done) and T04 (blocked).

---

## c) NOT STARTED

T04 (stack E2E, XREPO), T09 (owner sitting), T01/T02 (owner terminal), T10
(ETag+304 + scoped gzip), T11 (modulepreload + outgoing mic warm + ICE timing +
`iceServers` eval + curl baseline), **T12–T19 (the entire UI/UX M9–M26 train)**,
T21 (nix-review batch 2), T22 (samber/do follow-ups), T23 (visual harness),
T25 (server carve / schema_version — trigger-gated), T26 (AGENTS compaction,
markdownlint posture, announcements, watches), T27 (docs-health continuation).

---

## d) TOTALLY FUCKED UP!

1. **A test I authored failed on first run** (the hold-pending block in
   `calls.test.mjs`). Root cause: I asserted the post-failed-resume state was
   *un-held*. Fixed in-place (assert `holdPending === "resuming"` instead of the
   card's `data-state`), **but the fix is unverified**.
2. **Two background `node --test` runs never returned** (shell IDs `001` and
   `007`). Possibly the `nix run nixpkgs#nodejs` fetch is slow, or a test leaves
   an open handle (`ringToneStart`/`ringbackStart` install `setInterval`s;
   `--test-force-exit` should reap them). **Verification latency is a real
   process defect in this session** — I should have run host `node` (v24.20.0 is
   on PATH) on the *changed files only* first.
3. **I wrote one new test on a guessed state model** instead of reading the
   existing test's own sequence to the end (it had already driven the session to
   `held`). Process failure, not just a code bug.

---

## e) WHAT WE SHOULD IMPROVE

- **Verification order:** run the *changed test files* under host `node` after
  each edit, then the whole suite once at the end — never a full-suite run as the
  first feedback.
- **Stop using `nix run nixpkgs#nodejs` for iteration** — host node v24 is
  present; the nix path adds fetch latency and ambiguity.
- **Read a test to its last line before extending it** (the state it leaves the
  subject in is part of its contract).
- **Governance:** `AGENTS.md` is now **705 lines** (buildflow warns
  `max 377, excess 328`) — materially worse than the 498 the TODO row cites.
- **Stale duplicate doc:** `docs/status/dedup-registry.md` carries **4 broken
  file links** (lychee) — it is a stale copy of the live `docs/dedup-registry.md`.
- **Small real findings left un-dispatched:** 3 `errcheck` (`fax/service.go:215`,
  `paperless_test.go:77`, `version_test.go:75`), 8 `mypy` findings in
  `scripts/webphone-smoke.py`, `ruff` EXE001 (`scripts/render-diff.py` shebang not
  executable), 7 `oxlint` unused-var warnings in island tests, 2 `samber-linter`
  HW-4 infos on `internal/app/app.go` (the named services ARE eagerly invoked in
  `New` — the linter cannot see it; a reasoned suppression is the honest fix).

---

## f) Up to 50 things to get done next (priority-ordered)

| # | Task | Parent |
|---|------|--------|
| 1 | Confirm island suite green (host node, changed files) | T03 |
| 2 | T08: `aria-current` nav-partial pin | T08 |
| 3 | T08: skeleton reveal/hide island pin | T08 |
| 4 | T08: optimistic-bubble-morph-removed edge pin | T08 |
| 5 | T08: fix `#wp-live` over-promising comment (`layout.templ`) | T08 |
| 6 | T08: release-runbook stack-E2E obligation note | T08 |
| 7 | T20: dispatch 3 errcheck findings | T20 |
| 8 | T20: `nix run .#vulnix` verdict via `webphone-vulnix-triage` | T20 |
| 9 | T20: aarch64 ELF `e_machine=183` verify | T20 |
| 10 | T20: root-cause the KVM VM-test timeout | T20 |
| 11 | T10: ETag+304 for `/assets/*` (reconcile `server.go` comment) | T10 |
| 12 | T10: scoped gzip for static handlers (never `/events`) | T10 |
| 13 | T10: curls timing baseline before/after | T10 |
| 14 | T11: `modulepreload` for the island ESM graph | T11 |
| 15 | T11: outgoing-call mic warm (mirror incoming) | T11 |
| 16 | T11: ICE panel gathering-duration + time-to-first-media | T11 |
| 17 | T11: `iceServers` trimming evaluation | T11 |
| 18 | T12: M9 dial affordances (A4/A5/A8/A9/K5) | T12 |
| 19 | T12: M12 over-limit segment countdown | T12 |
| 20 | T13: M11 history filters (D8–D10) | T13 |
| 21 | T13: M16 URL state (E2/E3/E7/E8) | T13 |
| 22 | T14: M13 voicemail playback (C1–C3/C9/C10) | T14 |
| 23 | T14: M14 fax depth (C4–C6) | T14 |
| 24 | T15: M17 feedback/trust (J2/J3/J4/J7/J8/J9) | T15 |
| 25 | T16: M15 visual tokens + M20 theming + M26 sizing | T16 |
| 26 | T17: M18 onboarding + M19 mobile extras | T17 |
| 27 | T18: M21 messaging richness + M22 pin/archive/mute (server design first) | T18 |
| 28 | T19: M24 i18n/RTL/status dots + M25 call depth | T19 |
| 29 | T21: module-output golden fixture + check entry | T21 |
| 30 | T21: `release.sh` `webphoneVersion`↔`git describe` guard | T21 |
| 31 | T21: run KVM backup VM + drill post-split | T21 |
| 32 | T21: `actionlint` over CI workflow | T21 |
| 33 | T21: dedupe `devShells.ci`/`default` Go env | T21 |
| 34 | T21: record accepted exceptions + declined `go-standard` | T21 |
| 35 | T22: commit `health.css` build script + CI rebuild wiring | T22 |
| 36 | T22: `family_test.go` pins for new error codes | T22 |
| 37 | T22: investigate local-main-ahead-of-remote divergence | T22 |
| 38 | T23: persist `scripts/ui-capture.py` + 12-shot matrix | T23 |
| 39 | T23: AGENTS "visual gate" note | T23 |
| 40 | Fix `docs/status/dedup-registry.md` broken links (or remove duplicate) | hygiene |
| 41 | `chmod +x scripts/render-diff.py` (ruff EXE001) | hygiene |
| 42 | Fix 8 `mypy` findings in `scripts/webphone-smoke.py` | hygiene |
| 43 | Clear 7 `oxlint` unused-var warnings in island tests | hygiene |
| 44 | Suppress 2 `samber-linter` HW-4 infos with reasoned comment | hygiene |
| 45 | T26: compact `AGENTS.md` 705 → ≤377 lines (war stories → lessons) | T26 |
| 46 | T26: markdownlint posture ratification | T26 |
| 47 | T25: document `schema_version` trigger (gate on first ALTER) | T25 |
| 48 | T25: `internal/server` carve (trigger: next file added) | T25 |
| 49 | T27: annotate/archive recent status reports | T27 |
| 50 | Final gates: buildflow full + flake check + smoke after all edits | T03 |

---

## g) Questions I cannot answer myself (3)

1. **Ring-silence authorisation (T06).** The row says implementation "awaits the
   owner's go (tied to 'was your ring actually silent?')". Your blanket "do the
   whole list" read to me as the go, so I shipped it. Confirm, or revert?
2. **AGENTS.md compaction (T26).** It needs "explicit owner permission + a quiet
   window" and is now **705 lines** vs the 377 cap. May I compact it this
   session, or keep waiting for a declared quiet window?
3. **Stack E2E / T04.** The island-honesty row says the browser E2E is blocked by
   a **stack-side FreeSWITCH `mod_enum` build break**. Is that repaired yet — do
   I defer T04 (and its dependents) or is a stack session taking it?

---

_Point-in-time snapshot. `TODO_LIST.md` remains the living source; (f) is
harvest material for `docs-health` HARVEST when the session continues._

---

## Session 2 addendum (2026-10-01 continued)

Answers received: (Q1) T06 ring fix **KEEP**; (Q2) AGENTS.md compaction **go**;
(Q3) stack E2E status **"i do not know"** → T04 stays BLOCKED (handover only).

Completed + verified this session:

- **Suite confirmation**: island `node:test` green (120/120) and full
  `go test -count=1 ./...` green. The earlier hang was an authoring bug in
  `calls.test.mjs` (a gated `invite()` that the initial `placeCall` awaited) —
  fixed; also corrected the "failed hold" toast assertion (the settled state
  re-announces "connected", so the failure toast is not last).
- **T08 pins**: shell spec for the swap-time `aria-current` mirror + the tab
  skeleton reveal/hide + the "settled send is never rolled back" morph edge;
  server `TestServedPageHoldsTheDomContract` now asserts `aria-current="page"`
  and `"false"`; the over-promising `#wp-live` comment was corrected; the
  release runbook gained the "served-markup changes owe the stack E2E" rule.
- **T05 owed pin**: `TestShellHtmlLangFollowsSessionLang` (4 cases: cookie wins,
  Accept-Language fallback, EN default).
- **T20 hygiene**: `chmod +x scripts/render-diff.py`; 3 errcheck `defer
  Close()` findings fixed (fax/service.go, paperless_test.go, version_test.go);
  8 oxlint warnings fixed across theme-preload.js + island tests; 2
  samber-linter HW-4 infos suppressed with a reason at the ProvideNamed sites.
  (The `docs/status/dedup-registry.md` link finding was already gone.)
- **T10 perf**: `/assets/*` now serves a strong content ETag (sha256) and
  answers `If-None-Match` with a bodyless 304, plus scoped gzip (never
  `/events`); the stale "caching buys nothing" comment rewritten. New tests
  `TestAssetsCarryContentETag` + `TestAssetsGzipWhenAccepted`.
- **T26**: AGENTS.md compacted 705 → 364 lines (≤377); the moved train
  chronology/evidence appended to `docs/lessons.md` under "Provenance moved out
  of AGENTS.md".

Verification: fresh-binary smoke 47+4 green with the new asset caching; full
`go test` green; golangci-lint 0 issues; full buildflow run logged separately.
**Not started (still on `TODO_LIST.md`)**: T11, T12–T19 (the UI/UX M9–M26
train), T20 remainder, T21–T23, T25. Owner legs T01/T02/T04/T09/T24/T27 remain
handover-only.

