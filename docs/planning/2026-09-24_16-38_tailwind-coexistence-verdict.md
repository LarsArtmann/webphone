# Tailwind-coexistence verdict — GREEN

**Experiment:** plan M7 (`2026-09-24_13-33_SUPERB-stack-adoption-pareto.md`),
audit finding #2. Question: can templ-components' Tailwind v4 CSS load
beside webphone's unlayered token CSS (`app.css`) without breaking
either side?

**Verdict: GREEN — the adoption lane (M9/M10/M11) is unblocked.**

## Method (empirical, not theoretical)

- Throwaway session-gated route `/dev/spike/tailwind` (spike.templ +
  spike.go) renders REAL webphone surfaces (`.wp-panel`,
  `.wp-panel-head h2`, `.wp-welcome-points li`, `.wp-error`,
  `.wp-panel-sub`) beside three library components (`display.EmptyState`,
  `display.RelativeTime` with `AutoRefresh: false`, `display.CountBadge`).
- `?tw=0` omits the library stylesheet; `?tw=1` loads it. A same-origin
  probe (`/assets/spike/probe.js`) writes `getComputedStyle` JSON for
  both sides into `#probe-out`; headless Chromium 153 (`--dump-dom`,
  cookie via a loopback injecting proxy) captured both states.
- The library CSS was built with Tailwind v4.3.3 (`nixpkgs#tailwindcss_4`)
  from the RENDERED spike HTML (exact v1.19.2 class set), full
  `@import "tailwindcss"` including preflight, 12.7 KB minified,
  committed at `internal/web/assets/spike/tw.css`.

## Data (2026-09-24, Chromium 153 headless)

| Surface                | tw=0 → tw=1                                                                                                     |
| ---------------------- | --------------------------------------------------------------------------------------------------------------- |
| `.wp-panel`            | IDENTICAL (all 14 probed properties)                                                                            |
| `.wp-panel-head h2`    | IDENTICAL                                                                                                       |
| `.wp-welcome-points li`| IDENTICAL — list markers/margins survive the preflight                                                          |
| `.wp-error`            | IDENTICAL                                                                                                       |
| `.wp-panel-sub`        | IDENTICAL                                                                                                       |
| `EmptyState`           | styles GAIN: padding 60px/15px, icon 22.5px block, title 15px                                                   |
| `RelativeTime`         | styles GAIN: text-sm (13.125px) muted color                                                                     |
| `CountBadge` pill      | styles GAIN: absolute overlay, red bg, white 10px text, ring                                                     |

CSP: the spike page renders with ZERO inline scripts and ZERO `style=`
attributes (the probed components emit only classes;
`RelativeTime.AutoRefresh=false` avoids its injected script).

## Why it works (the mechanism, for the record)

Tailwind v4 emits everything inside `@layer` (theme/base/components/
utilities). Unlayered author CSS — all of `app.css` — outranks EVERY
layer in the cascade, so webphone tokens win every collision by spec,
and the class vocabularies barely overlap anyway (`wp-*` vs utilities).
The preflight reset applies only to properties `app.css` never sets —
on the probed surfaces that set is empty. Scope note: the probe covers
the surfaces above, not every page; the M9 wave still rides the full
DOM-contract tests + smoke + stack browser E2E as the safety net.

## Consequences for the waves

- **M9 (EmptyState ×9)**: proceed. Brand note: library text colors render
  as Tailwind `oklch` neutrals, not `--wp-*` tokens — acceptable for v1,
  override via `BaseProps.Class` if it reads off-brand.
- **M10 (RelativeTime + CountBadge)**: RelativeTime only with
  `AutoRefresh: false` (its live-update script is inline → dead under
  the zero-inline-script CSP; the static server-rendered text is the
  correct behavior here anyway). CountBadge fits `wp-nav-badge` spots —
  keep the DOM-contract id stable if the E2E greps it.
- **M11 (errorpage)**: evaluate as planned; the CSS question is now
  answered for its styling layer too.
- **Build recipe for waves**: regenerate `tw.css` by rendering the
  adopted components to HTML and running
  `nix run nixpkgs#tailwindcss_4 -- -i input.css -o out.css --minify`
  with `@source` pointing at the rendered HTML (exact version pin).
  nixpkgs `tailwindcss` (v3) does NOT work — v4 only.
- **Cleanup**: the spike route + `assets/spike/*` are THROWAWAY — remove
  before the next release fold (tracked in TODO_LIST / release checklist).

## Raw evidence

- `/tmp/dom-tw1.html`, `/tmp/dom-tw0.html` (session-local; the diff table
  above is the durable copy)
- diff run: `internal/server` suite green + CSP/DOM-contract tests green
  at commit time of this doc.
