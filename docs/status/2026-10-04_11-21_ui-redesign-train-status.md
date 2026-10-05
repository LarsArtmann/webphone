# Status — templ-components UI redesign train (2026-10-04 11:21)

Branch `main`, HEAD `afa46ff`, **pushed** (verified `git ls-remote`). Working
tree clean (the auto-commit daemon committed my in-flight app.css edit as
`afa46ff`). Plan doc: `docs/planning/2026-10-04_08-19_SUPERB-ui-redesign-completion-plan.md`.

## a) FULLY DONE

| Item                                     | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T1 contract-green foundation             | Full `go test -count=1 ./...`: every package green EXCEPT `internal/arch` `TestErrorCodeRegistryIsFresh` — the other session's standing error-contract drift (`store.*`/`webhook.*` registry-vs-doc lines; their files, daemon commit `04b1d94`). Attributed, not stomped. All markup-pin, DOM-contract, CSP, token-mirror, i18n-parity tests GREEN.                                                                                                          |
| T2 tw.css adoption layer                 | `internal/web/assets/tw.css.input` (committed): `@custom-variant dark` → `data-theme`, `@theme inline` remapping ALL Tailwind ramps used by the scan set (blue→accent, gray→graphite neutrals, red→danger, amber→warn, green→ok, white→on-accent), radii pinned to 10/14px with a drift assertion in the build script. `scripts/build-tw-css.sh` (staged build, locked nixpkgs rev, canary greps, scoped prettier). Artifact rebuilt + committed (`418df76`). |
| Drift gate for the adoption layer        | NEW `internal/server/twcss_test.go` `TestTwCssCoversAdoptedComponentClasses`: renders every adopted surface (Button primary/secondary/sm/icon, Input bare/labeled/error, Textarea, EmptyState+action, layout.Base) and asserts each emitted class token has a selector in tw.css. GREEN.                                                                                                                                                                      |
| `tc-auto-grow` ownership                 | `forms.Textarea` AutoGrow emits library class `tc-auto-grow` with NO CSS anywhere in v1.19.4 (consumer-owned). app.css now owns `.tc-auto-grow { field-sizing: content }` (like `.sr-only`).                                                                                                                                                                                                                                                                  |
| Visual capture round 1                   | Binary rebuilt (`nix build .#webphone`), 14/14 shots captured into `ui-shots/` (both themes, 7 tabs), committed by daemon. Phase-1 (pre-tw.css) shots overwritten as intended.                                                                                                                                                                                                                                                                                |
| Root-cause analysis of the white buttons | See d) — mechanism CONFIRMED, fix path designed.                                                                                                                                                                                                                                                                                                                                                                                                              |

## b) PARTIALLY DONE

| Item                        | State                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T3 visual verification loop | Round-1 shots reviewed (messages/thread/settings, light+dark). **Nits found**: composer Send + thread-head Call render as island "key" buttons (white/dark-gray, muted ink) instead of Primary green — root-caused (see d); snippet chip renders light-surface in dark (same family, needs one computed-style confirmation); adopted inputs show BOTH app.css border AND the utility ring shadow (subtle double edge). **Fix loop NOT closed** — one edit landed (`.wp-compose button` un-nesting + `.wp-older button` deletion, `afa46ff`) but it CANNOT take effect until the cascade-layer leak is fixed, and its committed comment currently overstates the effect. Re-capture round 2 pending. |
| T5 brand coherence sweep    | `.wp-older button` dead rule REMOVED. `button.wp-primary`/`button.wp-danger` were scheduled for deletion but are **NOT dead** (error.templ reload, contacts/voicemail row buttons) — kept. NOT DONE: `layout.templ` ThemeColor still old teal `#0f766e`/`#10161a` (layout.templ:55-56); favicon tint unchecked.                                                                                                                                                                                                                                                                                                                                                                                     |

## c) NOT STARTED

- T4 quality gates: `nix fmt` (repo-wide), island node tests, buildflow full.
- T6: webphone-smoke 48-check; AGENTS.md adoption paragraph; CHANGELOG entry.
- T7 contrast audit (WCAG token-pair ratios + fixes).
- T8 handoff: dedup-registry sweep-log line, FEATURES/TODO_LIST updates, stack
  browser-E2E obligation note (served markup changed → release-runbook item).
- Island-E2E re-capture after the cascade fix (round 2 of T3).

## d) TOTALLY FUCKED UP (honest list)

1. **The component adoption is currently visually INERT — and round 1 nearly
   shipped as "done".** tw.css utilities live in Tailwind's `@layer
   utilities`; **any unlayered rule beats any layered one regardless of
   specificity**. `island/style.css` carries a page-global bare
   `button { background: var(--key); padding: 11px 16px; ... }` (line 248)
   that leaks far outside the island DOM — it defeats EVERY adopted button's
   bg/padding/radius/ink. The screenshots caught it: Send/Call render as
   island keys, not green primaries. The old `.wp-compose button` descendant
   rule was the historical BANDAID over this leak; the "coexistence verdict"
   ("Tailwind emits @layer only, unlayered app.css wins") documented the
   mechanism as a feature for hand-styling but its blast radius on the
   adoption layer was never run to ground until now.
2. **My committed app.css comment lies right now.** `afa46ff` adds a comment
   claiming the Send button's "Tailwind utilities must decide the look" —
   true only AFTER the island leak is fixed. Until then the utilities lose.
   Rework lands with the fix.
3. **The previous session summary's "dead CSS" claim was wrong.** It said
   `button.wp-primary` rules were dead; grep shows `wp-primary` alive in
   error.templ and `wp-danger` in contacts/voicemail. I almost deleted live
   styles — caught by re-grepping before deletion.
4. **Diagnosis detour.** I first suspected markup, then specificity, and
   burned three hops before hitting the cascade-layer mechanism. A
   computed-style probe in the capture harness would have answered it in one.
5. The coverage gate verifies "selector EXISTS in the artifact", not
   "selector WINS the cascade" — that's why everything stayed green while
   the visible buttons were wrong.

## e) Improvements to work method (learned, being applied)

- Never declare CSS dead from a session summary — grep views + island JS +
  shell.js for every class before deleting.
- Verify cascade outcomes with computed styles (or a rendered probe), not
  artifact greps; selector presence ≠ selector victory.
- Capture ALL tabs in BOTH themes before calling any phase visually
  verified; the light thread shot alone hid nothing, but dark mode exposed
  the leak's full shape.
- Root-cause leaks at their home (island/style.css global element rules),
  not at the bandaid (composer descendant rules).
- The daemon commits mid-edit: re-View before every edit held true and
  saved the train twice.

## f) Next steps (Pareto-ordered)

1. **Fix the cascade-layer leak** (blocks everything visual): scope
   island/style.css's bare element rules (`button`, `input`, `h1`, `label`,
   their focus/active variants) to the island root container —
   behavior-preserving for the island (its own class rules keep winning by
   order), frees the tab region. Re-run island JS tests + DOM contract +
   round-2 capture to prove zero island regression.
2. Re-capture 14 shots; confirm Send/Call render Primary green in both
   themes; fix remaining nits (chip-in-dark computed-style probe, input
   double-edge) in ONE token/CSS batch.
3. Fix the committed comment in app.css (compose rule) to state the
   post-fix reality.
4. `nix fmt` + island node tests + buildflow full (T4).
5. ThemeColor hexes (`#0f766e`/`#10161a` → new accent/bg) + `templ generate`
   - favicon tint check (T5 rest).
6. webphone-smoke 48-check on the rebuilt binary (T6).
7. AGENTS.md: adoption list (Button/Input/Textarea/EmptyState/Base/icons),
   tw.css.input + build script note, the cascade-layer contract (utilities
   are layered; unlayered app/island rules win by design — adopted
   components must never share element+property with an unlayered global
   rule), tc-auto-grow ownership. Respect 377-line cap.
8. CHANGELOG entry for the redesign train.
9. Contrast audit script over token pairs (text/bg, on-accent/accent,
   muted/surface, status-on-fill); fix via tokens only, mirrored in both
   sheets (T7).
10. Explicit i18n test confirmation (already green in suite — call it out),
    dedup-registry sweep-log line, FEATURES/TODO_LIST updates (T8).
11. Note stack browser-E2E obligation for the release train (served markup
    changed).
12. Optionally: add a computed-style check to the capture harness so
    "styled, not just present" becomes provable per shot.
13. Owner-signoff loop + follow-up polish train (plan § "three owner
    questions").

## g) Questions I cannot settle alone

1. **Island stylesheet surgery approval**: the leak fix touches
   `island/style.css` (contract-sensitive: served verbatim, island JS tests
   pin behaviors). Scoping its bare element rules to the island root is my
   recommendation (zero island-internal change, frees the tab region) — but
   it is THE load-bearing stylesheet of the phone island, so: proceed in
   this train, or defer to a dedicated follow-up train with its own capture
   cycle?
2. **Palette sign-off** (standing owner question): signal-green on graphite
   is fully shipped behind tokens — sign off, or redirect?
3. **Dependency**: stay pinned v1.19.4 or bump to the local +195-commit
   version (Checkbox/Toggle, newer form kit) in this train? The tw.css
   remap + coverage gate make a bump cheap to verify, but it re-opens the
   capture loop.
