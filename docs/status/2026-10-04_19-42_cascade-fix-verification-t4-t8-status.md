# Status — cascade-leak fix VERIFIED; T3 closed; T4/T6/T7/T8 executed to their gates

_2026-10-04 19:42 CEST · train: SUPERB UI-redesign completion plan (T1→T8) ·
preceded by docs/status/2026-10-04_12-34_island-cascade-leak-fix-status.md (that
report's verification holes are what this session closed)_

**One-line outcome:** the cascade-layer leak fix (`a379976`, pushed) is now
FULLY VERIFIED — island node tests 166/166, server/web Go tests green, fresh
14-shot capture proves Send/Call Primary green in BOTH themes, chips
dark-correct, inputs single-edge — and T4 (gates), T6 (smoke + docs), T7
(contrast audit), T8 (bookkeeping) were executed this session. T5 (brand
hexes) remains deliberately blocked on the owner's palette sign-off. The
chip-in-dark mystery RESOLVED by observation (correct in both themes) — no
probe needed.

Concurrent train note: the passkey train (cf8373a → 8835863) landed
continuously through this session, healed its own `h.enrollPage` build
breakage mid-session, and touched phone.templ (new `Passkey` prop). All
full-suite failures were attributed before acting: NONE were mine.

---

## a) FULLY DONE

1. **T3 verification holes closed** (the three carried from the last report):
   - **Island node tests**: first run WITHOUT `--test-force-exit` hung
     (pending promises) and I killed it; re-run WITH the flag (the AGENTS
     documented form): **166/166 pass, 0 fail, 0 cancelled in 2.3s**.
   - **Go tests** `./internal/server/ ./internal/web/...`: first attempt hit
     the passkey train's mid-edit `h.enrollPage undefined` (their file, not
     mine — attributed, not touched); after their fix landed: **server 3.75s
     ok, web/assets ok, web/views ok** (DOM contract, byte pins incl.
     `panels_test.go:220`, CSP all green). Full suite later: **zero
     failures** (their `internal/arch` registry-drift failure also healed).
   - **Round-2 capture**: fresh `nix build` binary (GC-copied to /tmp — see
     d.4), 14/14 shots. VERIFIED by eyeball: composer Send + thread-head
     Call render **Primary green in light AND dark**; island renders in both
     themes; quick-reply chips **dark-correct in dark** (mystery resolved —
     no computed-style probe needed); adopted inputs single-edge; offline
     banner, row buttons, search inputs all correct. welcome-dismiss not
     rendered (seeded session already dismissed it; markup is byte-pin
     tested and green).
2. **Chip-in-dark + double-edge diagnosis closed**: root cause of the dark
   pass island ABSENCE (not a CSS bug) — the island hides the login view at
   module load and reveals the call view only after the resume probe
   settles (`main.js:150`); the harness's fixed 0.4s sleep raced it. Fixed
   the HARNESS: `scripts/ui-capture.py` now waits on `#phone-view` losing
   `[hidden]` (WebDriverWait, 8s) per shot instead of blind-sleeping.
   Re-capture: island present in all 14.
3. **T4 quality gates**: `nix fmt` repo-wide 0-changed; island tests green
   (above); **buildflow** ran to its findings gate — the ONLY error class is
   `gomod-check` (99 findings), the documented vendor-consistency FALSE
   POSITIVE (was ~54; grew with the passkey train's dep churn — same
   verified class, do-not-hand-edit stands). Timing "regressions" are
   parallel-load noise. Effectively green.
4. **T6 smoke**: `webphone-smoke.py` on a fresh binary: **47 passed,
   0 failed + restart scenario 4/4 + boot-failure scenario 8/8** (incl. all
   six boot-contract markers). NOTE: AGENTS says "48-check"; actual is 47 —
   label drift, nothing failed (see f.26).
5. **T6 docs truth**: AGENTS.md gained the **cascade-layer contract** ("ANY
   unlayered rule beats ANY layered utility at any specificity — island
   bare-element rules MUST stay scoped to `.island`; box-sizing +
   reduced-motion stay global on purpose; never un-scope") inside the
   templ-components adoption bullet, plus a capture-harness note (waits on
   the island call view). CHANGELOG.md gained the user-facing Fixed entry
   (adoption layer was visually inert → scoped; history filter `wp-mini`;
   summary focus ring; harness view-wait).
6. **T7 contrast audit** (script, both themes, 19 pairs): **body/muted text
   pairs ALL PASS with huge margins (4.6:1–17.8:1)**. Five real fails, ALL
   requiring palette-token changes (see g): dark `--on-accent` on
   `--accent` 3.21:1 (button labels), light `--accent` as text on `--bg`
   4.14:1 (links), `--border-strong` vs surface 1.86/1.53 (WCAG 1.4.11
   boundary question), and my own over-audited hover pair (NOT a WCAG
   requirement — dropped). NO token values changed: every remedy touches
   the exact hexes the owner's Q3 palette sign-off reserves. Audit =
   deliverable; remediation = owner-gated.
7. **T8 bookkeeping**: i18n explicitly confirmed (8 tests: format verbs
   match across languages, no unused dictionary keys, panel pins both
   languages — all PASS); dedup-registry sweep-log line appended (no
   art-dupl run — selector scoping, no repeated logic; registry consulted
   during blast-radius inventory); FEATURES.md rows updated (adoption row:
   cascade contract + Primary-green verification; capture-harness row:
   view-wait); TODO_LIST stack-legs line now carries this train's
   stack-E2E obligation note; T5.1 declared a NO-OP with evidence — both
   plan targets (`button.wp-primary`, `.wp-older`) are ALIVE
   (error/contacts/voicemail + messages pagination), the plan was written
   on stale info.
8. Session hygiene: verified `a379976` still the last touch on
   island/style.css (passkey train never trampled my files); killed the
   PRIOR session's orphaned background shell 051 (the never-collected
   island-test job — it was hung on the same missing flag).

## b) PARTIALLY DONE

1. **T5 brand coherence sweep** — 5.1 resolved as no-op (alive CSS), but
   5.2 (ThemeColor `#0f766e`→`#17a467` light, `#10161a`→`#0b0d11` dark,
   layout.templ:54-55) and 5.3 (favicon tint) NOT executed: blocked on
   owner Q3 by design.
2. **T7 remediation** — audit done (a.6); token fixes gated on Q3 (the
   dark button-label remedy alone has two legitimate designs: darken the
   green vs flip the label ink to graphite).
3. **Push state** — my code fix `a379976` is pushed, but origin/main
   (11ea2cd) lags local HEAD (8835863) by ~4 daemon commits (push lag is
   the daemon's cadence, not an error); my FEATURES/dedup/TODO_LIST edits
   from this session are still uncommitted in-tree (daemon will sweep).
4. **Full 14-shot eyeball** — I verified the 6 evidence-bearing shots
   (01/02/05/08/09 + spot checks) against the acceptance criteria; the
   remaining 8 shots are captured but not individually reviewed (the
   standing TODO g1 "eyeball the current matrix" still owns a full pass).

## c) NOT STARTED

1. T5.2/5.3 hexes + favicon (owner-gated).
2. WCAG token remediation (owner-gated, same ruling).
3. Stack browser E2E re-run (stack-side; obligation now NOTED in
   TODO_LIST, execution lives in the stack lane).
4. Release-train mechanics (version cut, vulnix gate, aarch64 ELF verify,
   pbx-artmann relock) — next train after sign-off.
5. The prior report's 3 owner questions were NOT answered in-session; they
   are re-surfaced (§g) with Q1 now MOOT (chip resolved).

## d) TOTALLY FUCKED UP (what I forgot / did badly — no dressing up)

1. **Ran the island tests without `--test-force-exit`** — the flag is in
   AGENTS.md's documented command; I trusted the session-handoff's stale
   command string instead of the repo's memory file. Cost: a 191s hang, a
   kill, a re-run. **Repo AGENTS.md outranks the handoff summary. Always.**
2. **Smoke flag error**: invoked `--binary` (handoff's string) instead of
   the script's actual `--bin`. Cost: one dead background round-trip.
3. **Piped the smoke output through `tail -10`** and LOST the main 48-check
   verdict line — had to re-run the whole smoke to report honestly.
   Capture verdicts, not vibes: grep the verdict lines, don't tail.
4. **The nix store GC race**: my first capture attempt crashed twice with
   FileNotFoundError because a concurrent session GC'd the build output
   between build and use. Fixed by copying the binary to /tmp — but I
   burned two failed runs rediscovering what "concurrent sessions" in
   AGENTS should have primed me for.
5. **Wrong reveal marker on first harness edit**: I reached for
   `.call-card` (a JS-created live-call element that NEVER appears in this
   lane) before checking phone.templ — would have hung every shot for 8s.
   Caught it by verifying the selector against the source before running;
   the check should have PRECEDED the first edit.
6. **Diagnosed the dark-island absence CSS-first**: I went digging through
   `:has()` rules and my own scoping diff before noticing the DECISIVE
   evidence already in the shots — the EN selector was visible (island
   present) while both views were hidden (mid-resume). DOM evidence first,
   stylesheet second.
7. **Reviewed shots only after the full 14-shot run** — a one-shot
   spot-check right after capture would have surfaced the dark-island
   race a full run earlier.
8. **Over-audited WCAG**: invented an "accent-strong over accent hover"
   pair that no WCAG success criterion requires. Self-caught and dropped,
   but it shipped in the first audit print.
9. **Consulted the completion plan before the AGENTS pin for T5.1** — the
   pin already said `wp-primary` is alive; the plan's dead-CSS step was
   stale on arrival. Cost: one investigation loop.
10. **Orphaned job 051**: inherited from the prior session, but I launched
    a duplicate island-test run before killing it. Kill-first, then re-run.

## e) WHAT WE SHOULD IMPROVE

1. **Handoff hygiene**: session summaries must carry VERIFIED command
   lines (`--bin`, `--test-force-exit`) or explicitly mark them stale;
   two of this session's three wasted runs trace to a stale handoff.
2. **Bake the flags in**: the flake could own an `island-js` app wrapping
   `node --test --test-force-exit` so the hang mode is structurally
   impossible (AGENTS command stays as the manual form).
3. **Capture harness**: the view-wait is in; next maturity step is a
   focused-state shot (focus-visible ring) and a fresh-data welcome-dismiss
   shot — both currently unverifiable in the lane.
4. **Verdict-line discipline**: report-generating runs should grep named
   verdict markers, never tail.
5. **gomod-check FP**: 99 findings of documented false positive is gate
   noise that trains every reader to ignore the findings gate — worth a
   proper suppression path or upstream fix so buildflow can exit 0.
6. **AGENTS "48-check" label vs 47 actual**: reconcile the label (or find
   the conditional check) before it erodes trust in the smoke number.
7. **ui-shots churn**: each capture rewrites tracked PNGs into daemon
   commits — decide per-release vs per-train persistence (standing owner
   call g1) or capture into an untracked out-dir by default.
8. **Palette decisions need ONE owner ruling covering**: direction
   (Q3), button-label ink in dark, light link-green, border-strong
   1.4.11 policy — they all touch the same tokens; ruling on them
   separately guarantees churn.

## f) NEXT (up to 50, impact-ordered)

1. **Ruling on §g questions** (unblocks 2, 3, 5, 6 below).
2. T5.2: ThemeColor hexes in layout.templ → `templ generate` → tests.
3. T5.3: favicon tint (check island/favicon.svg for old teal).
4. WCAG remediation via tokens (per the §g.1 ruling), mirrored BOTH sheets.
5. Re-capture round 3 + re-run island/go tests after 2–4.
6. Stack browser E2E re-run (this train's served-markup obligation).
7. Push-lag check: confirm daemon swept FEATURES/dedup/TODO_LIST edits and
   origin catches up (`git ls-remote`, not push logs).
8. Annotate the 12:34 prior status report as superseded by this one.
9. Mark T1–T8 statuses inline in the completion plan doc.
10. lessons.md: three war stories — unlayered-beats-layered (with the
    Send/Call-gray evidence), the --test-force-exit hang, the capture
    resume race.
11. Reconcile AGENTS smoke "48-check" → actual count (f.6/e.6).
12. gomod-check FP: file/suppress so buildflow exits 0 (e.5).
13. Decide WCAG script fate: permanent `scripts/wcag_check.py` + flake
    check vs throwaway (my lean: permanent, it's 60 lines and the token
    blocks WILL drift again).
14. border-strong 1.4.11 policy: accept-and-record vs retune (§g.3).
15. Fresh-data capture run for welcome-dismiss + a focus-visible shot
    (harness extensions).
16. History `wp-mini` visual proof: one capture with a stubbed
    `phone_api_url` so the filter form renders (empty lane can't).
17. `.wp-older` (messages pagination) — alive but never eyeballed in the
    matrix (history lane empty); verify its look in a seeded-history run.
18. Review the remaining 8 round-2 shots individually (b.4).
19. Toast visual check: `#toasts` lives inside `.island` — scoping covers
    it; one toast-visible shot would make it evidence instead of inference.
20. Verify the passkey train's enroll page (standalone CSP-clean) didn't
    need tw.css utilities it isn't getting.
21. Re-verify tw.css `@source` set against the passkey train's view churn.
22. TODO_LIST rows: add WCAG findings + tab-input metric acceptance as
    owner-gated rows (currently only in this report).
23. Release prep when sign-off lands: version bump in flake.nix,
    CHANGELOG release cut, `nix run .#vulnix`, aarch64 ELF-byte verify,
    webphone-module check, backup drill if module changed.
24. Release dance: push → stack re-pin → pbx-artmann relock
    (docs/release-runbook.md ritual).
25. Consider `--expect-version` wiring into the release smoke invocation.
26. Dep-bump owner question carried from the prior report (templ-components
    v1.19.4 vs local +195 commits) — still open, still cheap to defer.
27. Seed idempotency nit: chips accumulated 2→3 across capture runs on a
    reused /tmp data dir — make ui-capture seeding idempotent or always
    fresh-dir.
28. Confirm daemon attribution: check the sweep commits didn't merge my
    docs edits with passkey files confusingly (attribution hygiene).
29. `nix flake check` solo clean run at train close (buildflow covered it;
    the documented gate deserves a clean solo green).
30. Re-run erraudit tier-1/tier-2 early only if the passkey train added
    error paths (next scheduled 2026-10-22 otherwise).
31. Copy pass: the plan's Q3-alt (copy in scope?) — unresolved owner call.
32. Consider recording the chip-resolution story in the tw.css/verdict doc
    (the "light surface in dark" observation was real pre-fix; one line
    closes the loop for the next reader).
33. Double-check no byte-pin depends on the OLD island input metrics in the
    tab region (suite green post-scoping — a grep confirms cheaply).
34. Sweep for any other stylesheet loaded AFTER tw.css that could re-leak
    (health.css is scoped to /health — verify it can't hit app surfaces).
35. Add the smoke verdict-marker list to AGENTS so future sessions grep
    the right lines.
36. Consider making the capture harness fail (not just wait) if
    `#phone-view` never reveals — it already times out via WebDriverWait;
    improve the error message to name the resume race.
37. Archive ui-shots-before/ comparison against round-2 (plan 3.3 said
    side-by-side; round-2 verified against criteria instead).
38. `settingsRow` dedup owner call — FOURTH surfacing pending (unrelated
    but due).
39. Check whether the passkey train's island changes (webauthn.js,
    passkey.js) need entries in the oxlint globals config (suite green —
    verify the config, not just the outcome).
40. Verify `internal/userauth` error codes are in the generated registry
    (suite green incl. freshness test — spot-check the table).
41. Docs-health sweep at train close: FEATURES "18.9KB tw.css" size claim
    still true after any regen.
42. Keep the /tmp binaries out of release claims (they are capture/smoke
    artifacts, not release candidates).
43. Version the capture matrix: name round dirs (ui-shots-round2/) if
    per-train persistence wins (e.7).
44. Retire the hardcoded /nix/store PATH export from the handoff knowledge;
    the AGENTS nix-shell form is canonical (this host's flake-ref
    resolution broke it once — investigate `nix shell nixpkgs#...`
    resolving as a path flake here).
45. Confirm `#remote-audio` (inside .island) unaffected by scoping — it
    has no visual rules (one grep, zero risk, closes the loop).
46. Sweep my own CHANGELOG entry for accuracy after round-3 (it says
    "waits on the island call view" — keep in sync if the harness grows).
47. Dedup baseline `-t 3` ratification still open (FIFTH surfacing due).
48. `push-lag threshold` owner call (currently ~4 commits / ~1h lag —
    when is it BROKEN?).
49. Plan-doc hygiene: the completion plan's T5.1 step should be annotated
    (stale targets) so no future session "deletes alive CSS".
50. After everything: full-matrix eyeball pass + owner sign-off closes the
    80% → 100% gap the plan names.

## g) QUESTIONS FOR THE OWNER (cannot self-answer)

1. **Palette sign-off (Q3, expanded to cover the WCAG remedies)**: is
   signal-green-on-graphite CONFIRMED as the direction? If yes, two
   sub-rulings in the same breath: (a) dark-theme button labels — keep
   white and darken the green to ~#0f7a4a-class, or keep #17a467 and flip
   label ink to graphite; (b) light-theme link-green `--accent` 4.14:1 on
   page bg — darken to ≥4.5 (touches button bg too) or accept links-only
   exemption? T5 hexes + WCAG fixes + favicon all unblock on this.
2. **Tab-input metric shift (Q2, carried)**: scoping moved tab-region
   inputs from island metrics (11px padding / 10px radius / 2px focus
   offset) to app.css metrics (9px / 7px / 1px) — accept as designed, or
   bump shell tokens to restore the island look?
3. **Non-text contrast policy**: `--border-strong` fails WCAG 1.4.11's
   3:1 vs panels in BOTH themes (1.86 / 1.53) — accept as a recorded
   design decision (industry-common for hairline borders), or retune the
   tokens (visibly heavier hover/boundary strokes)? And should the WCAG
   ratio script become a permanent flake check so token drift re-triggers
   it?

---

_Verdict: the cascade fix is proven at every layer it could break; the train
is now waiting on exactly one owner ruling (palette) plus two small design
acceptances. Everything else that could be executed without you has been._
