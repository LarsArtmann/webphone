# templ-components — library utilization audit

**Verdict:** growing adoption on deliberate boundaries: the shell, buttons, empty states, forms and icons ride the library; domain panels are hand-rolled templ by design. Four cheap wins sit inside unruled territory. **Adoption 50/100 (moderate)** — 3.5 of 7 applicable rows.
**Pin:** v1.26.0 + icons v1.26.0 + utils v1.26.0 (direct); datastar + htmx submodules indirect v1.26.0 · **Latest (family tag):** charts/echarts/v1.26.0 · **Currency:** current · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** [2026-09-22 deep-dive](../../research/2026-09-22_templ-components-deep-dive.html) (adopted layout.Base + EmptyState + icons; RelativeTime/CountBadge/errorpage REJECTED; that audit shipped `PageProps.NoThemeScript` upstream); [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (15/100 — "1 of 121 components"; superseded: adoption is materially deeper now).

## Status

| | |
|---|---|
| Module family | root + icons + utils (+ datastar/htmx indirect; feedback/navigation/recipes/charts/errorpage present upstream) |
| Usage | `.templ` surfaces: icons ×24 call sites (16 distinct of 105 constants), display ×15 (Button ×8, EmptyState ×6, DefinitionList ×1), forms ×4, layout.Base ×2, utils.BaseProps ×4 |

## What the library offers

Root: `display` (49 components — badges, cards, charts, tables, tabs, modal/drawer, StatusBadge, Avatar, PageHeader, SimpleEmptyState, RelativeTime, CountBadge…), `forms` (26 — inputs, selects, FilterBar/Chips, Combobox, DatePicker…), `feedback` (Alert/InlineError, Toast, Skeletons, StepIndicator), `layout` (Base/Minimal/AppShell, theme), `navigation`, `recipes`, `templates`. Submodules: `icons` (105 named stroke SVGs + Spinner/RTL/animated), `utils` (BaseProps, format helpers, tone, wire), `datastar`/`htmx` (integration components), `charts/echarts`, `errorpage` (ErrorPage/NotFound404 + HTTP handler).

## How webphone uses it

- **layout.Base ×2** (internal/web/views/layout.templ:63 Shell, enroll.templ:23) with the full deliberate option set: `CSSPath:"/assets/app.css"`, `Locale` = session lang, `SEO{NoIndex:true}`, **`NoThemeScript:true`** (CSP — theme-preload is the same-origin `/assets/theme-preload.js`), ThemeColor/DarkThemeColor, Favicon; HTMX script served from cqrs-htmx assets (SRI default suppressed path untouched).
- **display**: `Button` ×8 (secondary/sm variants), `EmptyState` ×6, `DefinitionList` in Settings (13 `DefinitionItem`s).
- **forms**: `Input` ×3, `Textarea` ×1 with `utils.BaseProps`.
- **icons**: `icons.Icon(name, "wp-icon …")` ×24 across settings/messages/history/fax/voicemail/phone/contacts.
- **Boundaries held from the 09-22 audit**: `rg "RelativeTime|CountBadge|errorpage"` → 0 hits; the gray-family/token remap and cascade-layer contract documented in AGENTS.md still govern.

## Gap analysis

| # | Capability area | Status | Evidence / reason |
|---|---|---|---|
| 1 | layout.Base option surface | 🟢 Fully leveraged | both invocations set the deliberate option set incl. NoThemeScript |
| 2 | display adoption (Button/EmptyState/DefinitionList) | 🟢 Fully leveraged | ×15 sites; matches the adopted verdicts |
| 3 | icons | 🟢 Fully leveraged | 16/105 pulled on demand — an icon library is pull-as-need, not max-out |
| 4 | forms | 🟡 Partial | 4 components used; contacts add form, snippet picker and quick-send compose are hand-rolled `<form>` markup |
| 5 | **Status badges** (`FaxStatusBadge`, `OutboundStatusBadge`, `wp-status` spans) vs `display.StatusBadge`/`Badge` | 🔴 Missed | hand-rolled at fax.templ:214, messages.templ:738-744; the library component maps status→label+tone with the same semantics |
| 6 | **`wp-empty` inline paragraphs** (×10) vs `display.SimpleEmptyState` | 🔴 Missed | inconsistent with the ADOPTED EmptyState used elsewhere in the same panels; cheap consistency win |
| 7 | **errorBanner/ErrorPanel** vs `feedback.Alert`/`InlineError` | 🔴 Missed | feedback package not imported at all; error banner markup is hand-rolled (error.templ) |
| 8 | **panelHead** vs `display.PageHeader`/`SectionHeading` | 🔴 Missed | one-home helper (views/panels_test.go pins it) — a swap would be a component-for-component refactor, listed for the next panel-touching train |
| n/a | Avatars (`avatarFor`/`avatarHue`) vs `display.Avatar` | ⚪ N/A | ruled DELIBERATE hand-roll (AGENTS.md: avatars stay, with helper tests) |
| n/a | `RelativeTime`, `CountBadge`, `errorpage` | ⚪ N/A | ruled REJECTED 2026-09-22 (timestamps are server-formatted with stable classes; 404 must render inside the session shell) — do not re-propose without new evidence |
| n/a | navigation/recipes/charts/feedback-toasts/datastar/htmx components | ⚪ N/A | no consuming surface: nav is the shell's own bottom bar + rail; toasts are island-owned (`#toasts` aria-live); charts have no product home; datastar ships only via the health dashboard bundle |

**Score:** rows 1-3 = 1.0 (3.0) + row 4 = 0.5 + rows 5-8 = 0 → **3.5 / 7 = 50/100, moderate band.**

## Ruling

The thin-but-deep pattern is right for a domain-specific phone shell — do NOT chase component coverage for its own sake. The four red rows are worth a small train because they remove hand-rolled CSS/markup with no behavior change: **StatusBadge (fax + messages), SimpleEmptyState (×10), feedback.Alert (errorBanner)**; panelHead follows when a panel is next touched. Each keeps the cascade-layer contract (layered utilities only) and the greppable row classes.

**Refresh trigger:** a templ-components release with a new status/toast component; any panel redesign; the error surface train revisiting banners.

## How to verify

```console
rg -o "icons\.Icon" -N -g '*.templ' | wc -l                       # expect 24
rg -n "RelativeTime|CountBadge|errorpage" -g '*.templ' --type go | wc -l   # expect 0
rg -n "wp-empty" -g '*.templ' | wc -l                             # expect 10
rg -c "display\.StatusBadge|feedback\.Alert|SimpleEmptyState" -g '*.templ' | wc -l   # expect 0 today
```
