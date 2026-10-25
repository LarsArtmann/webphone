# Status: T23 /login visual harness — GO executed, dark-mode bug found + fixed

Snapshot header: **claims verified as of 18:37 CET, 2026-10-25** (remote main =
`7762b6b` = local HEAD, CI run 40216219081 green, tree clean apart from this
report). Scope: the owner's GO on report question g.1 — everything between the
17-51 report and now. Nothing else was touched.

---

## a) FULLY DONE (verified)

1. **Harness extension (the GO item itself)**: `scripts/ui-capture.py` boots
   passkey-ON (render-only config — rp_id `webphone.test`, no ceremony ever
   runs), and per theme captures two ANONYMOUS surfaces ahead of the
   session'd tabs: `login-card` (island card + its `/login` link) and
   `login-page` (the library page). Session cookie dropped for those shots,
   re-injected after (a live session 303s `/login` home and hides the card).
   Library theme driven by CDP `Emulation.setEmulatedMedia` — it follows
   prefers-color-scheme, not `wp-theme` (the documented forced-theme caveat),
   so emulation shows exactly what a same-scheme OS user sees. DOM markers
   asserted per shot (`#passkey-login` + its copy; `#lp-login-form`).
   Threshold 18 → 22 (12 surfaces × 2 = 24 shots, skip-tolerant).
2. **REAL BUG found by the new dark shot and FIXED**: the login stylesheet's
   gray family was remapped through app tokens (byte-pinned from the shell
   build). The loginpage uses Tailwind's standard dark idiom where one gray
   step serves BOTH a background and a text slot across themes — the shell
   remap (900 = `--text`, 50 = `--surface-2`) inverted dark mode: LIGHT page
   background (wrapper `dark:bg-gray-900` → light ink token) + INVISIBLE
   heading (`dark:text-gray-50` → dark surface token). Fix: grays stay
   Tailwind stock in the login build; accent/white/danger/warn/ok/radii
   remaps keep the brand. Pixel-verified: dark login shot mean brightness
   **220 → 27**; light shots unchanged (239/247).
3. **Second bug found + fixed (tw.css leak)**: the shell artifact grew 12
   lines of loginpage utilities (`dark:bg-gray-900`, `dark:text-gray-50`).
   Root cause after a wrong first guess (staging order — kept as hardening
   anyway): **a comment in `login.css.input` naming a utility string makes
   Tailwind v4's auto-scan EMIT that utility into the sibling build**. Fixed
   three ways: comment reworded (no literal dark-variant utility names), the
   parity gate now compares NON-GRAY slots only + a new gate REFUSES gray
   overrides in the login input, and the shell build runs before loginpage
   sources are staged. tw.css restored to its committed bytes (-12 = the
   leak exactly); rebuild deterministic.
4. **Gates**: shellcheck clean on the script; `nix fmt` clean on touched
   files; artifact gates in build-tw-css.sh green; full harness re-run 24/24;
   commit `7762b6b` pushed, **CI green** (40216219081, 2m55s — includes
   ruff/treefmt/contracts in CI where ruff actually lives).
5. **Docs**: AGENTS.md T23 comment (20 → 24 shots, anonymous-surface +
   CDP note), AGENTS.md passkey bullet (gray-stock rule + the
   comment-emits-utility trap), CHANGELOG `[Unreleased]` entry covering the
   dark fix + harness extension + build-script gates.

## b) PARTIALLY DONE

1. **Heading visibility is inferred, not pixel-proven**: the invisible-heading
   diagnosis came from reading the remap semantics; the mean-brightness
   verification proves the page/card backgrounds but NOT text contrast. The
   fixed dark shot (card crop 53 vs page 27) is consistent with a correct
   render, but no assertion samples INK against its background. An owner
   eyeball of `ui-shots/14-login-page-dark.png` (2 seconds) or a
   contrast-aware harness check would close this honestly.
2. **Local buildflow / flake check not re-run after the final artifact
   rebuild** — CI green IS the parity gate and it ran everything, so coverage
   is real; recorded here because the earlier session ran both locally and
   this one leaned on CI for the second half.

## c) NOT STARTED (deliberate — awaiting instructions)

1. Report questions **g.2** (loginpage pin policy: tag v4.13.0 vs HEAD) and
   **g.3** (branching-flow stale baseline: refresh vs quarterly audit) —
   owner has not answered; both stay as reported.
2. The remaining 48-item next-list from the 17-51 report (smoke-script
   `/login` probes, strict-CSP exclusion verification, freshness CI gate,
   etc.) — untouched, correctly so: the instruction was the GO item only.

## d) TOTALLY FUCKED UP (process waste, honestly named)

1. **Debug instrumentation in a daemon-watched repo**: my temporary DBG print
   in `scripts/ui-capture.py` was swept by a daemon commit (`21e8e31`) while
   I was still investigating. I removed it before the final commit, but had
   the session died mid-hunt, debug noise would have been on main. The
   correct move (copy the script to /tmp and instrument the copy) was my
   FIRST idea, botched via a sed one-liner whose quoting produced a
   traceback; instead of fixing the copy approach I edited the real file.
2. **The sed one-liner itself**: nested-quote hell that "validated" via
   ast.parse but crashed at runtime — I syntax-checked the wrong thing.
3. **Misleading instrumentation target**: the DBG printed
   `getComputedStyle(document.body)` — which was DARK and correct — while
   the light pixels came from the page WRAPPER div (`dark:bg-gray-900` →
   light token). The debug output said "state correct, pixels wrong" and
   cost me a mystery cycle before I went back to reading the CSS remap.
   Instrument the element you can SEE, not the one that's convenient.
4. **Wrong first root-cause for the tw.css leak**: I "fixed" staging order
   (loginpage staged after the shell build) and only discovered via
   unchanged output that the leak persisted — the real source was my own
   comment text. The reorder stays as defense-in-depth, but it shipped as a
   "fix" before being one.

## e) WHAT WE SHOULD IMPROVE (from this run's evidence)

1. **The "new surface" checklist needs a contrast check, not just
   brightness**: mean-brightness deltas caught the catastrophic inversion
   but are blind to ink-on-background failures. A harness assertion
   sampling text-region vs background-region means (or min/max luminance
   delta) would have caught the invisible heading directly.
2. **Never instrument tracked files in this repo** — /tmp copy + edit. The
   daemon makes any temporary code in-tree a race you will sometimes lose
   (I lost once this session; won by luck that the final commit won).
3. **Tailwind auto-scan is a loaded footgun here**: ANY file in the input's
   directory containing utility-shaped strings (comments, docs, test
   fixtures) emits utilities into the build. The parity + gray-refusal gates
   now catch the known class; a broader rule (no utility-shaped literals in
   `*.input` comments) is documented in the input header and AGENTS.
4. **Pixel verification without image rendering**: I cannot see images in
   this environment; ImageMagick mean/crop statistics were the substitute.
   Worked for backgrounds, insufficient for text. Either add the structural
   check (e.1) or route the final eyeball to the owner as a standing step.

## f) NEXT — up to 50 things (delta since the 17-51 report; its 50-list still stands)

1. Owner eyeball: `ui-shots/02-login-page-light.png` + `14-login-page-dark.png`
   (the two new library-page shots) — 2 seconds, closes b.1.
2. Harness: add an ink-vs-background luminance-delta assertion per shot
   (e.1) — catches invisible-text classes the means cannot.
3. Harness: assert the DARK pass actually went dark (mean < threshold) per
   surface — today darkness is verified externally by me, not by the script.
4. Smoke-script `/login` probes (17-51 f.3) — still the cheapest missing
   regression fence for the surface.
5. Verify the strict-CSP test's `/login` exclusion mechanism + comment it
   (17-51 f.4).
6. tw.css/login.css freshness rebuild-diff CI gate (17-51 f.7) — today's
   leak class is exactly what it would have caught in CI instead of by eye.
7. Consider committing a tiny README in `ui-shots/` naming the matrix and
   the "stale shots wiped per run" contract (the dir is now tracked).
8. The `--out-dir /tmp/ui-shots-dbg` leftover + `/tmp/wp-smoke-*` fixtures —
   cosmetic cleanup.
9. loginpage pin policy decision (17-51 g.2) — still open.
10. Branching-flow stale baseline decision (17-51 g.3) — still open.
11. Upstream (templ-components or loginpage): the @theme dual-use gray
    conflict is inherent to token-remapping consumers of the library page —
    a doc note in loginpage's README would spare the next consumer the same
    inversion hunt.
12. Upstream (art-dupl or here): nothing new this run.
13. The remaining ~38 items of the 17-51 f-list (chat networks, carve,
    catalog push, erraudit re-measure 2026-11-05, release obligations incl.
    the stack E2E before the next tag, etc.) — unchanged, still open.

## g) QUESTIONS (cannot figure out myself)

1. **Eyeball delegation**: please look at `ui-shots/14-login-page-dark.png`
   (and `02-login-page-light.png`) — is the library page's look acceptable
   as the product's front door (stock Tailwind neutrals + our green accent
   buttons), or should I invest in a webphone-tinted neutral set that flips
   correctly (new app.css token pairs used only by the login build)?
2. **Standing eyeball step**: should "owner views the new/changed shots"
   become a required line in the T23 runbook entry (the harness stays
   LOCAL-ONLY; the shots land in-repo either way), or is the automated
   marker+brightness verification the intended ceiling?
3. (Carried) **loginpage pin policy** (17-51 g.2): tag v4.13.0 until
   usermgmt pins align, or ride HEAD at the next natural alignment?

---

*Written from session memory + direct observations; no new investigations
were opened. Awaiting owner instructions.*
