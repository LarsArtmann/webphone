# UI/UX layout leverage — diagnosis, decisions, execution ledger

**Date:** 2026-10-29
**Status:** EXECUTED (P1 + P2 + P3 shipped; P4 deferred with named triggers)
**Scope:** the served surface — shell grid, the six HTMX tabs, `app.css`,
`shell.js`, `theme-preload.js`. NOT the island's in-call internals (they are
already the deepest-worked surface in the repo).
**Question driving it:** "How can we be smarter from a UI/UX perspective —
layout, multipage, CSS? And what could templ-components / cqrs-htmx still
give us?"

---

## 1. What was audited

| Surface                               | State found                                                                                                                            |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `views/layout.templ` (shell)          | header → grid `[island 320-400px \| main-col]` → horizontal wrapping nav strip → `#tab-content`. One grid, two breakpoints.            |
| `views/messages.templ` (904 lines)    | Thread list and open conversation are TWO separate panel renders of the same `#messages-panel`; opening a thread replaces the whole tab. |
| `views/history.templ` (404)           | filter form → day-grouped CDRs → transcripts section → reviews. One column, everything stacked.                                         |
| `views/settings.templ` (406)          | `DefinitionList` + 5 conditional sections + shortcut table in one column.                                                              |
| `assets/app.css` (2351 lines)         | Hand-rolled design system, token-based, 2 viewport media queries + 1 reduced-motion guard. Zero container queries.                      |
| `assets/shell.js` (2355 lines)        | Palette, shortcuts, theme, skeleton, voicemail player, composer, SSE recovery. Mature.                                                 |
| `docs/planning/2026-10-01_03-49_ui-ux-idea-catalogue.md` | 120 ideas; ~36 marked `Now`. The layout cluster (E1–E10) is the least-executed: E5 (reclaim island width) never landed. |

---

## 2. Diagnosis — the six things actually wrong

Ranked by user-visible damage per unit of work. Everything below is a
*structural* complaint, not a coat of paint.

1. **The Messages tab has no master–detail.** Opening a conversation destroys
   the list you were reading. Every desktop messaging product (Signal Desktop,
   iMessage, Google Messages, WhatsApp Web) keeps the list beside the
   conversation. This is the single largest layout defect in the product, and
   it is worse here than in a generic app because the island eats 320–400px of
   the width before the tab even starts.
2. **The island is a fixed 320–400px tax.** It is the product when you are
   calling and dead weight when you are reading a thread, a fax list, or
   settings. Catalog item E5 has been open since 2026-10-01.
3. **Layout decisions are made on the viewport, not on the column.** Every
   breakpoint in `app.css` is a `@media` query. But the tab column's width is
   `viewport − island − padding`: change the island and the column changes
   *without the viewport changing*. Any layout built on `@media` is therefore
   structurally wrong for this shell. This is the root cause that makes #1 and
   #2 hard to fix independently — and fixing it makes them one change.
4. **`History` and `Settings` are unscannable single columns.** History mixes
   two unrelated datasets (CDRs and persisted call transcripts) in one scroll
   region; Settings stacks six conditional sections under one heading.
5. **Disabled/unavailable/empty states are three different shapes.** History's
   "needs phone API" renders a dashed `.wp-empty` box; voicemail's renders a
   `.wp-empty` paragraph with inline `<code>`; the empty cases use
   `display.EmptyState`. Same ladder, three renderings, no shared spoken cue.
6. **Navigation is a wrapping text strip.** Six labels, no icons, wraps to two
   rows at tablet widths. Fine, but it is the cheapest remaining scannability
   win and it becomes load-bearing once the island can collapse.

---

## 3. Decisions

### D1 — Multipage: **NO.** Deepen the in-shell view model instead.

A conventional multi-page app would reload the document on navigation. The SIP
island lives in that document and **must never unload** (AGENTS: "The island
never unloads"). A reload during a call tears down the `RTCPeerConnection`,
the registration, and the microphone pre-warm. Anything that reads as
"multipage" — addressable URLs, back/forward, deep links, shareable state —
must therefore be delivered as *URL-addressable sub-views inside one document*,
which is exactly what HTMX swap + `hx-push-url` already provides.

So the answer to "multipage?" is: **no routes added for the sake of routes**;
add addressable *state* (`?archived=1`, `?older=N`, `/messages/<id>`) where the
state is worth restoring. The one thing a multipage architecture would have
bought — a real master–detail URL pair (`/messages` and `/messages/:id`
rendering differently) — we already have.

The honest limitation this creates, stated because it is the cost of the
decision: a full-page reload re-registers SIP (~1s), whereas a swap does not.
That is the price of never paging, and it is the right price for a phone.

### D2 — Messages becomes a master–detail split, driven by **container queries**.

`ThreadView` grows a conversation rail, so the open-thread render carries both
panes. The split engages on the *tab column's* width, not the viewport's:

```
.wp-main-col { container-type: inline-size; }
@container (min-width: 66rem) { .wp-msg-split { flex-direction: row } }
```

This makes D2 and D3 **compose for free**: collapsing the island widens the
column, the container query flips, and Messages goes split — no JS involved, no
second breakpoint to keep in sync, no `@media` duplicate.

Why not `layout.Split` from templ-components? Because this repo's ratified
convention is that **Tailwind styles adopted components, `app.css` styles our
own markup** (`tw.css.input`: "Webphone does NOT use Tailwind for its own
markup"). The split is our markup; it belongs in `app.css` where it can use the
design tokens and container queries. Adopting `layout.Split` would also add a
scanned-source dependency and a `tw.css` rebuild for one grid.

Cost accepted: the rail is hidden (not lazy-loaded) below the container
threshold, so mobile pays a few hundred bytes of duplicate list markup it never
shows. That is cheaper than a second render path that can drift.

### D3 — The island becomes collapsible on desktop, persisted pre-paint.

Header toggle + `localStorage`, applied by `theme-preload.js` before first paint
(the same file that already applies the theme and the welcome dismissal, for
the same reason: a late class application flashes the wrong layout). The island
stays in the DOM and keeps its runtime alive; only its column collapses, so SIP
registration is untouched. A live call force-expands it (`:has(.call-card)`),
because hiding the phone during a call is never what the user wants.

### D4 — CSS: container queries for component layout, tokens for shell geometry.

Rules going forward:

- **Component-level layout** (split lists, rail widths, card grids) → `@container`
  on `.wp-main-col`. Never `@media`.
- **Page-level chrome** (bottom tab bar, island stacking, padding) → `@media`,
  because those genuinely track the viewport. Mobile stays `max-width: 900px`.
- **Shell geometry** gets tokens (`--island-w`, `--content-max`) so the collapse
  and the grid agree without duplicating numbers.

Not doing: a wholesale `app.css` rewrite, Tailwind-ifying our markup, or a
`@layer` reorganisation. The token layer is already sound; the missing
ingredient was container queries, not structure.

### D5 — templ-components leverage audit

| Component                                | Verdict                     | Reasoning                                                                                                                                    |
| ---------------------------------------- | --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `display.DefinitionList`                 | **already adopted**         | Settings rows (M21 train).                                                                                                                    |
| `display.EmptyState`                     | **already adopted**         | List/thread/fax/voicemail/history empties.                                                                                                    |
| `layout.Base`                            | **already adopted**         | Head tags, skip link, `NoThemeScript`.                                                                                                        |
| `display.Button`, `forms.Input/Textarea` | **already adopted**         | Composer + fax form.                                                                                                                          |
| `icons.*`                                | **already adopted**         | All row actions. No platform emoji left.                                                                                                      |
| `display.DataState`                      | **ADOPTED NOW (P3)**        | The disabled→unavailable→empty→content ladder is exactly the History/Fax/Voicemail shape; unifies the three renderings found in §2.5.         |
| `display.CommandPalette`                 | **REJECTED**                | The shell's palette is JS-built (keeps the served DOM contract untouched) and covers nav + dial + shell actions. Adopting the component means shipping its `data-tc-palette` runtime and moving palette markup into served bytes to gain grouping/hints we do not need. |
| `layout.Split` / `display.Split`         | **REJECTED**                | See D2: our markup is token-styled `app.css`, and the split must answer a container query.                                                    |
| `forms.FilterBar` / `FilterChips` / `SegmentedControl` | **DEFERRED** | History's filter is a 4-control GET form that already submits + pushes URL. Rewriting it is a net-neutral style swap with real regression risk (it is pinned by `history_test.go`). Trigger: the next time History's filter needs a *new* control. |
| `feedback.Skeleton*`                     | **DEFERRED**                | The tab skeleton is 3 shimmer bars; the component adds a grid + card variants we would not use. Same trigger as the filter bar: adopt when the skeleton needs to know about content shape. |
| `display.Table` / `DataTable`            | **REJECTED**                | Rows are conversation/fax/voicemail objects with domain affordances and SSE swap-safety constraints, not tabular data.                        |
| `display.RelativeTime`, `display.CountBadge`, `errorpage` | **REJECTED (standing)** | Recorded rulings from the 2026-09 coexistence verdict; not reopened here.                                                                     |
| `htmx.ViewTransitions`                   | **DEFERRED**                | Native View Transitions on tab swaps would genuinely lift the perceived quality, but `hx-swap="morph:innerHTML"` and the island's non-unloading geometry make the interaction between the two untested. Trigger: a dedicated train with the stack browser E2E in hand. |

### D6 — cqrs-htmx leverage audit

| Module        | Verdict               | Reasoning                                                                                                                                      |
| ------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| root          | **already adopted**   | SSE broadcaster, CSRF/security middleware, recovery, error mapping, projection readiness, OpenAPI builder.                                     |
| `usermgmt` + `webauthn` + `identity-model` | **already adopted** | Passkey lane, durable auth-audit journal.                                                                                                      |
| `loginpage`   | **already adopted**   | `/login` ceremony page (nonce CSP lane).                                                                                                       |
| `adminui`     | **DEFERRED**          | The only remaining cqrs-htmx module with *user-facing* value: an identity-ops panel (users/tenants/members/audit) over the adopted `usermgmt.Service`. It is a PRODUCT decision (do we want an in-app identity admin surface at all?) plus an authz decision (who may see it), not a UI/UX win — and it is the bundle that pushed the rejected `setup` composition root over the binary-size bar. Trigger: a second human actor actually using the system (the multi-actor substrate that landed 2026-10-28). |
| `dashboardui` | **REJECTED**          | Event-store observability; the go-health dashboard at `/health` already owns ops read-only.                                                     |
| `setup`       | **REJECTED (standing)** | Split-brain identity + measured +68% binary size.                                                                                            |
| `datastar`    | **REJECTED**          | The shell is HTMX + idiomorph by design; a second transport doubles the swap model for no user-visible gain.                                    |

**The honest summary:** cqrs-htmx has no further *UI/UX* leverage to give this
app. Its remaining value is `adminui` as a product feature, gated on a second
actor. templ-components has one clean win left (`DataState`) and the rest is
either already mined or deliberately repo-inappropriate.

---

## 4. Execution ledger

| #  | Change                                                                                        | Verification                                    |
| -- | --------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| P1 | Messages master–detail: `ThreadView` carries the conversation rail; split via container query. | `templ generate`, Go suite, island suite, type gate |
| P2 | Island collapse (header toggle, pre-paint persistence, call force-expand).                      | `theme-preload.test.mjs`, shell tests, type gate   |
| P3 | `display.DataState` adoption for the disabled/unavailable rungs (History, Fax, Voicemail).      | `panels_test.go`, `tw.css` rebuild                 |
| P4 | Nav iconography (per-tab icons) — not shipped this train.                                       | —                                               |

P4 was scoped and dropped: once the island can collapse the strip has room, so
icons buy scannability rather than correctness. Left in TODO_LIST as a polish
item rather than smuggled in unverified.

---

## 5. What this doc does NOT claim

- No stack browser E2E was run (that suite lives in the consuming repo). Served
  markup changed, so AGENTS' release-runbook obligation stands: the next stack
  session re-runs it.
- The visual harness (`scripts/ui-capture.py`, 24 shots) was not re-run — it is
  local-only by budget decision and needs chromium + selenium.
- `nix flake check` was run where feasible; the KVM-gated backup VM test was
  not.
