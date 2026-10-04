# Status — post-verification housekeeping: §f executed, gomod-check suppressed, buildflow exits 0

_2026-10-04 23:20 · HEAD `ee45946`+sweeps · this report covers THIS session's run only
(the 19:42 report of the cascade-verification session remains the baseline for the train;
superseded-by pointers: none yet — this one EXTENDS it)._

## Context

Resumed after the 19:42 comprehensive report with the owner's standing
"WAIT FOR INSTRUCTIONS" on the palette-gated work. The resuming prompt
directed: read, understand, execute the unblocked remainder. So this
session closed every §f item that was NOT owner-gated: the §f.8–f.13
housekeeping block, plus eight zero-risk evidence-closure greps, plus a
fresh re-verification of both test lanes against the passkey train's
landed commits. The owner-gated set (palette, tab-input metrics,
border-strong policy, WCAG-script fate) is untouched and still blocking.

## a) FULLY DONE

1. **Todo list recreated** from the handoff (7 items, accurate statuses).
2. **Git state surveyed**: the daemon had swept the prior session's
   FEATURES/dedup/TODO_LIST edits (`a1942e6`, exactly those 3 files);
   the untracked 19:42 report got swept later (`142d14b`, alongside
   this session's plan-doc and 12:34 edits). Push lag (remote `11ea2cd`
   vs local `ee45946`+sweeps) is the daemon's cadence, not an error.
3. **CSS-lane trampling check**: ZERO commits touched
   `internal/web/assets/island/style.css` or `internal/web/assets/app.css`
   since the cascade fix `a379976` — the passkey train never entered my
   lanes.
4. **Both test lanes re-verified post-train**: island node tests
   **184/184 pass, 0 fail** (was 166 — the passkey train added 18
   webauthn specs: base64url, prepareLoginOptions,
   prepareRegistrationOptions, serializeCredential); full Go suite
   **green** (unpiped exit 0).
5. **§f.8** — the 12:34 cascade-fix report now carries a SUPERSEDED
   banner pointing at the 19:42 report (fix-design record preserved).
6. **§f.9 + §f.49** — the completion plan doc carries an "Execution
   status (2026-10-04 19:42)" block marking T1–T8 (DONE except T5
   BLOCKED-on-owner, T7 audit-done/fixes-blocked), and micro-task 5.1 is
   annotated **STALE — NO-OP: both targets are ALIVE** so no future
   session deletes the live `.wp-older` pagination CSS.
7. **§f.10** — three war stories appended to docs/lessons.md:
   unlayered-beats-layered (cascade arithmetic), the `--test-force-exit`
   green-then-hang, and the capture resume race (wait on events, not
   clocks; reveal markers must be static markup).
8. **§f.11 + §f.35** — AGENTS.md smoke command line no longer hardcodes
   "48-check": the script PRINTS the counts (`smoke:/restart scenario:/
   boot failure scenario: N passed`), grep them, never assume. Also
   pinned the `--bin` flag (a stale handoff string cost a wasted run
   last session).
9. **§f.12** — gomod-check SUPPRESSED via `skip_steps` with full
   rationale in `.buildflow.yml`: 99/99 error-severity findings are the
   documented vendor-consistency FALSE POSITIVE; the one new
   `ignore-needed` finding ("add ignore vendor directive") is
   info-severity and same-family (this repo commits `vendor/` and builds
   `-mod=vendor` deliberately). The real gate stands: `nix flake check`
   builds the vendored tree in the sandbox. **`buildflow` now exits 0**
   — verified by an UNPIPED run (`REAL_BUILDFLOW_EXIT=0`) after I caught
   my own pipe-masked measurement (see §d.1).
10. **AGENTS.md buildflow-health entry updated** with the suppression —
    merged into the passkey train's mid-session rewrite of that passage
    without losing their content (their "grows with the vendor tree:
    ~54 before, 99 after" wording carried over verbatim).
11. **Eight evidence loops closed with commands, not arguments**:
    - §f.45: `#remote-audio` has NO visual rules in either stylesheet —
      scoping-proof.
    - §f.33: panels byte-pins pin MARKUP (classes, aria-labels, URLs),
      not CSS metrics — nothing pins the old leaked styling.
    - §f.40: userauth error codes (`nclose.db`, `nclose.users`,
      `nconfig`, `ndb.open` + passkey rows) are in the
      docs/error-contract.md registry; the freshness test enforces it.
    - §f.39: oxlint config needs NO new globals for the passkey modules —
      `env.browser` covers `navigator`/`PublicKeyCredential`; only
      `SIP`/`module` are manual. PROVEN by building
      `.#checks.x86_64-linux.island-lint` green.
    - §f.20 (utilities half): the passkey enroll page loads
      `/assets/app.css` + island style.css, NOT tw.css, and uses only
      the app's own `card` class — no tw.css dependency at all.
    - §f.21: rebuilt the tw.css artifact with the locked tailwindcss_4 —
      **byte-identical** to the committed one; the passkey train's view
      churn demands zero new utilities.
    - §f.34: health.css's bare-element rules are scoped by LOAD SITE
      (dashboard-only mount, pinned by `app_test.go` asserting the
      dashboard page references its scoped build) — it cannot reach app
      surfaces.
    - §f.28: the daemon's sweep commit `a1942e6` contains exactly the 3
      files I authored — attribution clean.

## b) PARTIALLY DONE

1. **AGENTS.md line cap**: the file is ~40 lines over its own 377 cap
   (preflight warns every run: "AGENTS.md has 417 lines (max: 377)" at
   22:53; `wc -l` says 419 after my +2 — the counter delta is itself
   unreconciled, §d.4). Debt row added to TODO_LIST with the trim
   recipe (move evidence to docs/, keep rules) and the safety note: do
   NOT trim while the passkey train is mid-flight on that file.
   Ironic confession: my own suppression note added +2 lines to the
   over-cap file.
2. **The 19:42 report's §f list**: this session closed items
   8, 9, 10, 11, 12 and the evidence-closers 20 (half), 21, 28, 33, 34,
   39, 40, 45 — ~14 of 50. The remaining ~36 are re-numbered in §f
   below.
3. **Tw.css drift automation**: this session PROVED artifact freshness
   by manual rebuild+diff; a `checks.tw-css` drift gate in the flake
   (mirroring the existing `checks.health-css`) would make that
   automatic. Proposed in §f, not built.

## c) NOT STARTED (owner-gated or lane-gated — deliberately untouched)

1. **T5.2** ThemeColor hexes in layout.templ — blocked on the owner's
   palette sign-off (§g.1).
2. **T5.3** favicon.svg teal check/re-tint — same gate.
3. **WCAG remediation** via tokens (5 real fails: dark on-accent 3.21:1,
   light accent-as-text 4.14:1, border-strong 1.86/1.53:1) — blocked on
   §g.1 + §g.3; mirrored-both-sheets rule applies when it lands.
4. **Round-3 capture matrix** — depends on 1–3.
5. **Stack browser E2E re-run** — stack lane; the obligation is noted in
   TODO_LIST's cross-repo section.
6. **Tab-input metric ruling** (accept vs bump) — owner §g.2.
7. **border-strong 1.4.11 policy** — owner §g.3.
8. **WCAG script permanent home** (`scripts/wcag_check.py` + drift gate
   vs throwaway) — owner §g.3; script still at `/tmp/wcag_check.py`.
9. **Release train** (version bump, CHANGELOG cut, vulnix, aarch64 ELF
   verify, module check) and the **release dance** (push → stack re-pin
   → pbx-artmann relock) — next train after sign-off.
10. **AGENTS.md trim to ≤377** — waiting for the passkey train to close
    (shared file, mid-flight edits).

## d) TOTALLY FUCKED UP (this session's mistakes — no dressing up)

1. **Pipe-masked exit codes, TWICE.** (a) The first island-test
   invocation piped into `grep -E '^# (tests|pass|fail)'` — node
   emitted `ℹ`-prefixed summary lines, grep matched nothing, the PIPE
   exited 1, and the run looked like a test failure. One wasted
   invocation + a false alarm before re-running raw. (b) Worse: my
   first "buildflow EXIT=0" claim was `echo $?` AFTER a `| tail` pipe —
   that measured TAIL's exit code, not buildflow's. The claim happened
   to be true, but the measurement was invalid; I re-ran unpiped
   (`> /tmp/bf.log 2>&1; echo $?` → `REAL_BUILDFLOW_EXIT=0`) before
   letting the claim stand in docs. Rule going into lessons-proposals:
   capture exits unpiped (or `${PIPESTATUS[0]}`), and never grep a
   tool's output before seeing it raw once.
2. **Edited a shared file off a stale read.** My AGENTS.md suppression
   edit failed exact-match because the passkey train had rewritten that
   exact passage between my session-start read and my edit (their
   "grows with the vendor tree" version). The failure mode was SAFE
   (exact-match refused; re-read showed their content; merge preserved
   every concept) — but I KNEW the train was active and should have
   re-read immediately before the first attempt, not after the failure.
3. **Grew an over-cap file.** The suppression note (+2 lines) landed in
   an AGENTS.md already 40 lines past the cap the repo itself declares.
   A one-line pointer ("gomod-check FP suppressed via skip_steps,
   rationale in .buildflow.yml — do NOT hand-edit vendor markers")
   would have carried the same policy at a third of the cost.
4. **Unreconciled counter delta.** The preflight said 417 lines at
   22:53; `wc -l` said 416 before my edit and 419 after a +2 edit. The
   1-line discrepancy (counting method vs concurrent edit) was never
   chased down, and the TODO_LIST debt row quotes the wc number without
   noting the gate's own counter disagrees. The gate's number is the
   truth that matters; say so.
5. **Not repeated from last session**: none of the 10 confessed mistakes
   in the 19:42 report §d recurred (no stale-command trust, no
   tail-piped verdicts left standing, no wrong reveal markers, no
   CSS-first diagnosis).

## e) WHAT WE SHOULD IMPROVE

1. **Exit-code and output-format discipline** (from §d.1): raw output
   once, THEN filter; exits via redirect+`$?` or PIPESTATUS. This is
   the portable lesson of the session — worth a lessons.md line.
2. **Re-read shared files immediately before editing during active
   concurrent trains** — "read this session" is not fresh enough when
   another train's daemon commits between your read and your edit.
3. **Generalize the printed-count rule**: AGENTS now says it for smoke
   counts; the same ban on hardcoded counts should apply to island test
   totals and capture-shot counts wherever docs mention them (166 → 184
   already drifted once).
4. **Cap discipline as standing practice**: any session that adds lines
   to AGENTS.md trims an equal number elsewhere in the same edit — the
   file only ever grows otherwise.
5. **A `checks.tw-css` flake drift gate** (mirror of `checks.health-css`)
   would replace this session's manual rebuild-and-diff proof; tw.css
   staleness would fail `nix flake check` instead of waiting for a
   human to notice.
6. **buildflow `--format finding` ergonomics**: the JSON is
   banner-prefixed and needs a `sed -n '/^{/,$p'` dance; a documented
   one-liner in AGENTS (or an upstream raw-JSON format) saves the next
   person the discovery.
7. **Trust the gate's own counter**: for cap/limit discussions, quote
   the preflight's number, not `wc`'s.
8. **Evidence-closers as a batch habit worked well** — eight loops
   closed with greps/builds in two tool calls. Keep "close every loop
   with a command, not an argument" as the default posture.

## f) NEXT (up to 50, impact-ordered; owner rulings first)

1. **Owner §g.1** — palette sign-off incl. the two sub-rulings (dark
   button-label ink; light link-green). Unblocks 4–8.
2. **Owner §g.2** — tab-input metric shift: accept or bump.
3. **Owner §g.3** — border-strong 1.4.11 policy + WCAG script fate
   (permanent `scripts/wcag_check.py` + flake gate vs throwaway).
4. T5.2: ThemeColor hexes in layout.templ → `templ generate` → views
   tests.
5. T5.3: favicon.svg — check for old teal, re-tint.
6. WCAG remediation via tokens, mirrored in BOTH app.css and
   island/style.css; re-run the ratio script.
7. Round-3 14-shot capture after 4–6.
8. Island + Go suites after 4–6.
9. Stack browser E2E re-run (stack lane; cascade train's served-markup
   obligation, noted in TODO_LIST).
10. lessons.md: append the pipe-masked-exit lesson (§d.1) — one line,
    next session opener.
11. Trim AGENTS.md to ≤377 (move evidence parentheticals to docs/;
    only after the passkey train closes; quote the preflight counter).
12. `nix flake check` solo clean run at train close (documented gate
    deserves a solo green; buildflow covered it so far).
13. Check whether the passkey train's userauth error paths need an
    early erraudit tier-1/tier-2 re-run (else next scheduled
    2026-10-22; their erraudit-nolint commits suggest they handled it —
    verify, don't assume).
14. Confirm the passkey enroll page is inside the strict-CSP served-page
    test's page set (§f.20 closed only the tw.css-utilities half).
15. Release prep when sign-off lands: version bump in flake.nix,
    CHANGELOG release cut, `nix run .#vulnix`, aarch64 ELF-byte verify,
    `webphone-module` check.
16. Release dance: push → stack re-pin → pbx-artmann relock
    (docs/release-runbook.md ritual).
17. Wire `--expect-version` into the release smoke invocation.
18. TODO_LIST rows: add the WCAG findings + tab-input metric acceptance
    as owner-gated rows (currently only in the two status reports).
19. WCAG script permanent home IF §g.3 says permanent: move to
    `scripts/`, wire a drift gate mirroring `checks.health-css`.
20. `checks.tw-css` drift gate in the flake (§e.5) — would have caught
    any passkey-train utility demand automatically.
21. Capture seeding idempotency: chips accumulated 2→3 across runs on a
    reused /tmp data dir — make seeding idempotent or always fresh-dir.
22. Toast-visible capture shot (make `#toasts` styling evidence, not
    inference).
23. Fresh-data capture for welcome-dismiss + a focus-visible shot
    (harness extensions).
24. History `wp-mini` visual proof: one capture with a stubbed
    `phone_api_url` so the filter form renders.
25. `.wp-older` (messages pagination) — alive but never eyeballed in
    the matrix; verify its look in a seeded-history run.
26. Review the remaining 8 round-2 shots individually (b.4 carryover).
27. Record the chip-resolution story in the tw coexistence verdict doc
    (one line closes the loop for the next reader).
28. `settingsRow` dedup owner call — FOURTH surfacing pending.
29. Dedup baseline `-t 3` ratification — FIFTH surfacing due.
30. Push-lag threshold owner call (when is ~4-commit lag BROKEN?).
31. Plan Q3-alt (copy in scope?) — unresolved owner call.
32. Archive `ui-shots-before/` comparison against round-2 (plan 3.3's
    side-by-side was replaced by per-criterion verification; either do
    it or strike it).
33. Version the capture matrix: named round dirs (`ui-shots-round2/`)
    if per-train persistence wins.
34. Retire the hardcoded /nix/store PATH export from handoff knowledge;
    investigate why `nix shell nixpkgs#...` resolves as a PATH flake on
    this host (the capture recipe depends on the hardcoded form).
35. Docs-health sweep at train close: FEATURES "18.9KB tw.css" size
    claim still true after any regen.
36. Keep the /tmp binaries out of release claims (capture/smoke
    artifacts, not release candidates).
37. Capture harness: fail-with-explanation if `#phone-view` never
    reveals (name the resume race in the timeout message).
38. CHANGELOG harness wording sync if the harness grows (f.46
    carryover).
39. templ-components dep-bump owner question (v1.19.4 vs local
    +195 commits) — still open, still cheap to defer.
40. Double-check no byte-pin depends on old island input metrics in the
    tab region beyond panels_test.go (suite green post-scoping; wider
    grep is cheap).
41. Sweep for any other stylesheet loaded AFTER tw.css that could
    re-leak (health.css is load-site scoped — PROVEN this session;
    confirm no future page loads it outside the dashboard mount).
42. Island oxlint globals: standing rule — new browser globals go in the
    config's `globals` block (passkey modules needed none; PROVEN).
43. Full-matrix eyeball pass + owner sign-off — closes the plan's
    80% → 100% gap.
44. After sign-off + 4–8: re-check FEATURES/CHANGELOG claims against
    the final state (palette hexes, WCAG numbers).
45. Investigate the wc-vs-preflight line-counter delta (§d.4) — one
    command, closes a small honesty gap.

## g) QUESTIONS FOR THE OWNER (cannot self-answer — the standing three)

1. **Palette sign-off (§g.1 at 19:42, expanded)**: is
   signal-green-on-graphite CONFIRMED? If yes, the two sub-rulings in
   the same breath: (a) dark-theme button labels — keep white-on-
   `#17a467` (fails AA at 3.21:1) or darken the ink/brighten the green?
   (b) light-theme link green `#0e8752` on bg fails AA at 4.14:1 as
   text — darken to the ~4.5:1 neighbor or accept?
2. **Tab-input metric shift** (§g.2): accept the post-scoping input
   metrics in the tab region, or bump the tab design?
3. **border-strong 1.4.11 policy + WCAG script fate** (§g.3): accept
   and record the 1.86:1/1.53:1 boundary contrast, or retune the
   tokens? And: does `/tmp/wcag_check.py` become permanent
   (`scripts/wcag_check.py` + flake drift gate) or stay throwaway?
