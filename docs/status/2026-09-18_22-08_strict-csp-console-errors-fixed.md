# Status Report — Strict-CSP Console Errors Fixed

- **Timestamp:** 2026-09-18 22:08 CEST
- **Session scope:** User pasted four browser-console errors from the
  deployed site (`pbx.artmann.tech` vhost). Root-cause, fix, verify, and
  document — nothing else. No unrelated research was done.
- **Verdict in one line:** All three app-caused console errors are
  root-caused, fixed, and verified against the locally served binary;
  the fourth is a browser-extension artifact; production redeploy and
  real-browser re-verification are still open.

The four reported errors and their dispositions:

| # | Console error                                                                               | Cause found                                                                                                                                                                                                        | Disposition                                                                  |
| - | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------- |
| 1 | Inline script violates `script-src 'self'` (hash `sha256-AO4Oq…`)                           | templ-components `layout.Base` emits its theme-preload script unconditionally (no opt-out knob as of v1.18.0, the latest version); webphone passes `Nonce: ""` so no nonce attribute                               | **FIXED** — hash-pinned in `contentSecurityPolicy` (server.go)               |
| 2 | htmx "Applying inline style violates `style-src 'self'`" (function `Wn` @ htmx.min.js:3467) | htmx's `insertIndicatorStyles` injects an inline `<style>` at startup (`includeIndicatorStyles`); htmx.min.js loads synchronously BEFORE `HeadContent`, so a config meta in headExtras alone would arrive too late | **FIXED** — `htmx-config` meta + htmx deferred in `headExtras`               |
| 3 | `/favicon.ico` 404                                                                          | Shell's `layout.PageProps` never set `Favicon`, so Base emitted no `<link rel="icon">` and the browser probed the default path                                                                                     | **FIXED** — `Favicon: "/favicon.svg"`                                        |
| 4 | `Unchecked runtime.lastError: The message port closed…`                                     | Chrome-extension artifact, not this app's code                                                                                                                                                                     | **NOT APPLICABLE** — explicitly not "fixed"; no action possible in this repo |

---

## a) FULLY DONE

1. **Inline script CSP violation (error 1).** Root-caused rigorously:
   fetched the page served by the locally built binary, extracted every
   inline `<script>`, computed sha256, and matched the console's hash
   byte-for-byte (`AO4OqWm6Ms8LvRxbwPXsO4Fy1QauZJl6sD4NNzrR7S0=`) — the
   culprit is templ-components' theme script, not webphone code.
   Verified it is inert here (webphone themes via `data-theme` from
   shell.js, not Tailwind's `dark` class; the script reads a foreign
   `theme` localStorage key). Fixed by allowing exactly that one script
   by hash in `script-src` (`contentSecurityPolicy`, server.go).
2. **Forced-theme hardening.** Because the hash-pin lets the foreign
   script RUN, its inline `documentElement.style.colorScheme` would
   beat app.css's `[data-theme]` rules and undo forced themes. Both
   forced-theme `color-scheme` rules in app.css now carry `!important`
   (inline style loses to `!important`). The `:root` default (auto
   mode) is untouched — the script's OS-derived value is correct there.
3. **htmx inline-style CSP violation (error 2).** Inspected the actual
   served bundle (cqrs-htmx v4.9.0 embed): `Wn` is the indicator-style
   injection gated on `includeIndicatorStyles`. Fixed at the root:
   `<meta name="htmx-config" content='{"includeIndicatorStyles":false}'>`
   placed in `headExtras` BEFORE the htmx script, and htmx now loads
   `defer` (same pattern as the already-deferred sse.js; defer scripts
   execute in document order, so extension registration is safe).
   Equivalent indicator rules shipped in app.css so future
   `hx-indicator` usage keeps working.
4. **favicon 404 (error 3).** `Favicon: "/favicon.svg"` added to Shell's
   PageProps; Base emits `<link rel="icon" href="/favicon.svg">`
   (verified in served HTML). Modern browsers stop probing
   `/favicon.ico` when an icon link exists.
5. **Error-panel inline handler.** `error.templ`'s
   `onclick="location.reload()"` (a guaranteed future CSP violation on
   the error page) replaced by a `data-reload` attribute plus a
   delegated click listener in shell.js — the same pattern as the
   existing `data-dial` handler.
6. **Canary test `TestServedPageSatisfiesStrictCSP`** (server_test.go):
   two-way check that every inline script served is covered by an exact
   `sha256-` hash in the CSP header AND no stale hash lingers; also
   asserts no inline handlers, the htmx-config meta, and the icon link.
   This closes the specific gate hole that let all three violations
   ship. Written, passing.
7. **End-to-end verification:** `templ generate` (committed generated
   files), build via `GOEXPERIMENT=jsonv2 go build`, smoke server run,
   served-page assertions via node one-liners (exactly one htmx
   reference, meta-before-script, icon link, zero `onclick`, CSP header
   contains the hash), full `go test ./...` green across all packages,
   new canary test green.
8. **Buildflow quality gate:** full run then cached rerun, **exit 0**
   ("passed with warnings"). Zero failed steps. Remaining findings are
   all pre-existing and documented: vulnix (BUILD-closure advisories —
   AGENTS.md documents the honest number is the runtime closure),
   lychee (404 on the not-yet-published v2.0.0 GitHub release link),
   nix-checker (vendorHash style advisory), go-auto-upgrade (4
   warning-level `lo.SliceToMap` style suggestions; `samber/lo` is not
   even a dependency here — deliberately not acted on).
9. **Docs:** AGENTS.md CSP invariant updated (one hash-pinned script
   allowed, canary test named) + two new hard-won-knowledge bullets
   (htmx meta-ordering; the `!important`/hash/`HTMXSrc` trio and how to
   remove it if upstream ships an opt-out); CHANGELOG Unreleased →
   Fixed entries. All 10 changed files committed by the auto-commit
   daemon (c7c25f1 code, 512068c tests/docs/generated).

## b) PARTIALLY DONE

1. **Real-browser + production verification.** The user's console log
   came from the deployed vhost; the fix exists only in this repo.
   Verified: locally served binary, CSP semantics, unit tests. NOT
   verified: the production site after redeploy, and any real browser.
   The stack-side browser E2E (AGENTS.md: re-run after ANY markup
   change) has not been run. — UPDATE 2026-09-19: the stack E2E has since run GREEN with real headless chromiums (06:42 report §a.14, fix `00f13fe`); the production vhost redeploy + console eyeball remains open → TODO_LIST.
2. ~~**htmx deferred runtime behavior.** Deferred-load ordering is
   reasoning-verified (document order: htmx.min.js → sse.js → shell.js;
   htmx initializes on DOMContentLoaded, so the meta is read and hx-*
   attributes still process), but no browser has actually exercised
   nav swaps, SSE liveness, or form POSTs against the new head.~~ — done: exercised by the green stack E2E (06:42).
3. **Forced-theme visual matrix.** The `!important` guard is
   logic-verified, not visually verified across the
   theme(auto/light/dark) × OS(prefers light/dark) combinations. — UPDATE 2026-09-19: mechanically verified that app.css re-declares the island token names under `:root[data-theme=…]` blocks (so forced themes cascade into the island); the visual matrix itself stays open (ROADMAP, browser-level gates).
4. ~~**Lint-cleanliness of the new test.** go-auto-upgrade flags my new
   set-building loop (suggests `lo.SliceToMap`). Deliberately kept:
   adding `samber/lo` for a warning-level style nit is negative value,
   and the repo already carries 3 identical tolerated findings. "Done"
   as a decision; the debt counter went from 3 to 4.~~ — decided (kept); recorded here as the policy answer.
5. ~~**This report itself feeds TODO_LIST** — per the status-report
   skill, section (f) should be harvested into TODO_LIST.md /
   ROADMAP.md (docs-health HARVEST). Not done yet in this session.~~ — done (2026-09-19 docs-health sweep: Tier items routed to TODO_LIST/ROADMAP).

## c) NOT STARTED

1. **Upstream templ-components fix** — add a ThemeScript opt-out knob
   (zero-value = current behavior for backward compat), release, bump
   here, then drop the CSP hash + the two `!important`s. — standing watch item (AGENTS.md CSP bullet); upstream knob as of v1.18.0 still absent.
2. ~~**Stack-side browser E2E re-run** (mandated by AGENTS.md for markup
   changes; lives in nix-international-telephony, needs stack/PBX env).~~ done at `00f13fe` (green 06:42 report §a.14)
3. **A browser-console cleanliness gate** — nothing in this repo's
   gates executes the page in a browser. → ROADMAP (browser-level gates — the leverage play).
4. ~~**`/favicon.ico` route** for old bookmarks/scrapers~~ **Won't implement —** decided: the `<link rel="icon">` tag is enough for browsers; non-browser clients get the 404.
5. **CSP nonce mode** (documented future option). → ROADMAP (raw ideas).
6. ~~**TODO_LIST harvest** of this report (b5/f).~~ done (2026-09-19 sweep).

## d) TOTALLY FUCKED UP

Nothing this session broke the product — full suite green, buildflow
exit 0, the 35-element DOM contract intact. But honestly:

1. **v2.0.0 shipped with three console errors on every page load.**
   The AGENTS.md invariant said "no inline scripts/styles. Keep it that
   way" — and it was already untrue the day it was written: Base always
   emitted the theme script, htmx always injected its style block, and
   the favicon 404 fired for every visitor. The invariant was dead
   letter from the v2 rebuild until today. The gates (DOM contract
   test, buildflow, arch tests) grep HTML strings but never EXECUTE the
   page in a browser, so CSP violations were structurally invisible.
   The user — not any gate — found this. The canary test closes the
   hash/meta/icon holes at the HTML level; a browser-level gate still
   does not exist.
2. **The `!important` workaround is clever-but-smelly.** One fact (the
   pinned theme script must never win over forced themes) now lives in
   three coupled places: the CSP hash comment (server.go), the app.css
   rules, and AGENTS.md. Cross-referenced and two-way-tested, yes — but
   it is a deliberate split-brain-by-design that a single upstream knob
   would delete.

## e) WHAT WE SHOULD IMPROVE (self-review)

**What did I forget?**

- Did not check whether templ-components is checked out locally before
  declaring the upstream fix out of scope — "exhaust paths first" was
  violated by assumption, not by evidence.
- No test pins the `data-reload` handler: TestStaticAssetsServe asserts
  shell.js contains `data-dial`, but `data-reload` (the handler I
  added) is asserted nowhere. The global no-`onclick` canary only runs
  against `/`, which never renders the error panel.
- Did not fully read island/style.css (first 70 lines only). app.css's
  comment claims the explicit-choice rule wins "in BOTH stylesheets",
  which implies island/style.css also has `[data-theme]` rules —
  unverified. If it does not, the island ignores forced themes and the
  comment lies. Left unverified; flagged, not fixed.
- Did not organize the E2E handoff — just "should be re-run" in my
  closing message instead of proposing concretely who/what/when.

**What is stupid that we do anyway?**

- Shipping "strict CSP" claims with zero browser-level execution
  testing (the structural cause of d1).
- Three homes for the hash-pin fact (mitigated by the two-way canary,
  but still three homes).

**What could I have done better?**

- Surfaces the fix-choice decision matrix (hash-pin vs nonce vs replace
  layout.Base vs upstream knob) to the user earlier — the chosen fix
  touches a documented architectural invariant, and per my own rules
  that deserved a visible tradeoff statement before execution rather
  than only after.
- Throwaway node one-liners did the served-page verification before
  the Go canary existed; writing the Go test first would have made the
  throwaway steps unnecessary.
- The final message said "All four console issues fixed" — accurate
  for 1–3, but the fourth was never fixable here. The report now says
  so in a table; the earlier phrasing invites over-reading.

**What could I still improve?**

- Replace "hash in const + comment + docs" with a generated golden
  value (test validates a checked-in golden file; update = one
  deliberate diff).
- Headless-browser smoke test asserting a clean console on `/` — the
  single highest-value gate addition this session exposed.
- Self-review mid-task instead of on demand: the two genuine misses
  (local templ-components checkout, data-reload assertion) were both
  catchable by a mid-task checklist.

**Did I lie?** No. All verification claims are scoped to what was
actually measured (local binary, tests, buildflow exit code); the
unverified parts (production, real browser, E2E) are labeled as open.

**Ghost systems?** One deliberate, documented near-ghost: the app.css
`.htmx-indicator` rules are currently unused (no `hx-indicator` in any
template). Kept as htmx-parity with a why-comment; first use or first
cleanup should decide its fate.

**Split brains?** (a) The hash-pin trio above. (b) The island/app.css
token blocks duplicate design tokens with a "keep in sync" comment —
pre-existing, documented, unchanged this session.

**Scope creep?** Avoided: did not rip out `layout.Base`, did not fix
the (suspected) island theme gap, did not release upstream without
authorization, did not touch the unrelated buildflow findings.

**Tests overall:** Go suite is strong (contract, arch, SSE, store,
config, proxy). Blind spots hit this session: browser execution,
JS-behavior assertions (string-level only), i18n of new UI strings
(none added this session), and error-path rendering (ErrorPanel is
never rendered in any test — it only appears on action failure).

## f) Up to 50 things to get done next

Tier 1 — verify this fix (highest impact, small effort):

1. Redeploy the binary to the production vhost and re-open the browser
   console on `pbx.artmann.tech` — confirm errors 1–3 are gone in the
   environment they were reported from. → TODO_LIST (owner/ops row)
2. ~~Run the stack-side browser E2E (`tests/browser-e2e.py` in
   nix-international-telephony) — AGENTS.md mandates it after markup
   changes; the head changed.~~ done at `00f13fe` (green 06:42)
3. Manual browser matrix: forced light/dark/auto × OS light/dark —
   confirm forced themes still win (the `!important` guard). → ROADMAP (browser-level gates)
4. ~~Real-browser regression pass on the deferred htmx: nav tab swaps,
   SSE-driven transcript/fax/voicemail updates, composer POSTs.~~ done at `00f13fe` (E2E exercised the island + registration flows against the new head)
5. ~~Add `"data-reload"` to TestStaticAssetsServe's shell.js assertion.~~ → TODO_LIST (contract-pinning tests row)
6. ~~Read island/style.css in full; settle whether the island honors
   `data-theme` (app.css comment says BOTH stylesheets — verify or fix
   the comment/code).~~ verified 2026-09-19: island/style.css has NO `data-theme` rules — app.css re-declares the island token variable names under `:root[data-theme=…]` blocks (app.css:31–60), so forced themes cascade; the comment is truthful.
7. ~~Decide `/favicon.ico`: serve the SVG bytes there too, or accept the
   404 for non-browser clients (document the choice).~~ decided: link tag only (see c.4)
8. ~~Cut a patch release (v2.0.1) so deployments pick this up; date the
   CHANGELOG's Unreleased section.~~ done differently: the fixes landed BEFORE the v2.0.0 tag — everything shipped in v2.0.0 (`d9d6d03`); no patch needed

Tier 2 — close the gate hole that let this ship:

9. Headless-browser console gate: load `/` (and one signed-in state),
   assert zero console errors/warnings; wire into buildflow or CI. → ROADMAP (browser-level gates)
10. ~~Extend the served-page contract test to assert ORDER: htmx-config
    meta strictly before the htmx script tag (presence is tested;
    order is what makes it work).~~ → TODO_LIST (contract-pinning tests row)
11. Golden-file the CSP hash so a dependency bump produces a
    one-line deliberate diff instead of a test failure to interpret. → ROADMAP (CSP polish)
12. ~~Add an error-path render test: exercise ErrorPanel markup (data-
    reload button present, no inline handler) via a failing action.~~ → ROADMAP (testing long tail)
13. ~~Sweep the 4 `lo.SliceToMap` warnings as a policy decision: adopt
    samber/lo repo-wide, or document the loop idiom as a skip.~~ decided: keep the loop idiom (no `samber/lo`; see §b.4)
14. Keep `TestServedPageSatisfiesStrictCSP` fast/stable (it renders
    the full page; watch for flakiness as page grows). — standing note, no action required
15. Consider asserting the sse.js extension loads after htmx
    (document-order guard) in the same contract test. → TODO_LIST (contract-pinning tests row)
16. ~~Add the island's greppable-contract strings (dtmf-relay,
    reconnect watchdog, etc.) to a served-asset test so AGENTS.md's
    "survive by construction" claim is actually enforced in Go tests.~~ → ROADMAP (testing long tail; DOM contract already pins the ids)

Tier 3 — upstream + docs:

17. Check out templ-components locally; implement the ThemeScript
    opt-out knob (zero value = current behavior) with tests. — standing watch item (needs owner authorization for an upstream release)
18. Release templ-components (needs authorization), bump webphone,
    then delete the CSP hash, the two `!important`s, and the AGENTS.md
    trio note together (the documented removal set). — standing (rides on 17)
19. ~~Harvest this report into TODO_LIST.md (docs-health HARVEST) —
    Tier 1–3 items are actionable now, the rest routes to ROADMAP.~~ done (2026-09-19 sweep)
20. ~~README: note the single hash-pinned framework script in the CSP
    contract section so the sales page stays truthful.~~ **NOT-DO —** "strict same-origin CSP / no CDN" remains truthful (the pinned script is same-origin inline); the nuance lives in AGENTS.md
21. ~~FEATURES.md: refine the Strict-CSP row with the same nuance.~~ **NOT-DO —** same reasoning as 20
22. ~~Publish the v2.0.0 GitHub release (lychee currently 404s its tag
    link — the only link-check finding).~~ tag done at `d9d6d03` (lychee 0 errors); a GitHub Release OBJECT remains an open owner call → ROADMAP open questions
23. Decide the nix-checker vendorHash-extraction advisory: do it or
    document a skip in `.buildflow.yml`. — open, trivial; the documented deviation in AGENTS.md (manual hash application) covers the practice
24. VACUUM the buildflow cache DB (doctor warned: 1.34 GB) and rebuild
    the stale buildflow binary (built at 42fd89b, HEAD is 30c9344). → TODO_LIST (buildflow doctor row)

Tier 4 — known backlog pointers spotted this session (route via
docs-health, do not treat as freshly researched):

25. ~~Stack-side switchover to `nixosModules.default` (AGENTS.md: open
    work; module is additive today).~~ done (switchover executed 2026-09-18/19 — ROADMAP)
26. Call recording support — the investigation report ranks in-island
    playback (same-origin audio) as top option. → ROADMAP (recording cluster)
27. PWA offline shell (FEATURES: WORTH_CONSIDERING; must respect the
    strict CSP — interacts with this session's CSP work). — FEATURES WORTH_CONSIDERING
28. SMS/MMS audit + delivery-receipts self-review items (docs/status
    2026-09-18_16-34, 30 numbered items) — harvest rather than trust. → done (2026-09-19 sweep annotated that report)
29. CSP nonce mode for stricter deployments (docs/status 2026-09-18_
    16-21 item 44) — becomes the natural replacement for the hash-pin. → ROADMAP
30. ~~v2.0.0 status report's remaining numbered items — bring current
    with docs-health ANNOTATE before acting on stale claims.~~ done (2026-09-19 sweep annotated the 15:25/16:21 reports inline)

Tier 5 — small polish noticed in passing:

31. ~~Verify the `Nonce` field (still `""`) in Shell's PageProps is worth
    keeping now that the hash approach is chosen — drop it or comment
    why it stays.~~ minor; the hash approach is two-way-tested — leaving as-is
32. Consider naming the CSP hash constant (e.g. `themeScriptHash`) and
    building `contentSecurityPolicy` from it — self-documenting const. → ROADMAP (CSP polish)
33. The error panel's reload keeps the island-unload behavior of the
    old onclick (full page reload); consider an hx-get refresh of the
    failed tab instead — UX decision, not a bug. → ROADMAP (UX polish)
34. app.css + island/style.css token duplication ("keep in sync"
    comment) — extract shared tokens or generate one from the other. → ROADMAP (UX/platform polish)
35. Give shell.js a count-neutral header comment (it grew from three
    behaviors to five). — trivial, left to the next shell.js touch
36. Double-check `media-src 'self'` covers remote-audio for call
    recording's future in-island playback option. → ROADMAP (recording cluster — verify at build time)
37. `TestConfigJSContract` asserts `phoneApi:false` for the loopback
    smoke path — add a PBX-configured variant if config matrix grows. — conditional; open as written
38. Consider a `docs/status/README` index of reports (they are
    accumulating with no table of contents). — **NOT-DO for now —** the 2026-09-19 sweep archived the resolved reports (docs/status/archived/), shrinking the live set to the newest report

(Items 9, 11, 17–19 are the leverage plays: they prevent the whole
class of "shipped CSP violation" rather than this instance.)

## g) Questions I can NOT figure out myself

1. **Deployment:** where does the production binary get built/deployed
   from (which host/flake input consumes this repo), and may I trigger
   or at least verify the redeploy — or is that your manual step? → TODO_LIST (production redeploy row, owner/ops)
2. **Upstream authorization:** may I patch templ-components (ThemeScript
   opt-out knob) and release it? — standing owner authorization question (blocking nothing; the hash-pin is two-way-tested)
3. **Browser gate tooling:** may I add a headless-browser dependency
   (e.g. Playwright/chromedp + its browser download) to this repo's
   devShell/buildflow for the console-cleanliness gate? → ROADMAP (browser-level gates; owner call on the toolchain surface)

---

_Point-in-time snapshot; goes stale. Feed section (f) to docs-health
HARVEST; annotate, never rewrite._
