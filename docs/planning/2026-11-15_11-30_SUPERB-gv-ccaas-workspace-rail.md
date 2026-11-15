# GV + CCaaS workspace: the sidebar rail, and the IA ladder behind it

**Date:** 2026-11-15
**Status:** RAIL LANDED (source + tests); rollout ladder named below.
**Owner directive this answers:** "I asked for multipage 5 fucking times! I want
this to be a SUPERB mix of Google Voice + [Google Cloud Contact Center as a
Service]."

---

## 1. What the record got wrong

D1 of the 2026-10-29 layout-leverage doc answered "multipage?" with **NO**. The
reasoning was sound about one thing and wrong about the question:

- What is true: a **document reload** tears down the SIP island
  (`RTCPeerConnection`, registration, mic pre-warm). The phone must live in one
  continuously-mounted document.
- What is false: that this settles "multipage". **Google Voice and Google's
  CCaaS agent workspace are both single-document apps** with a left sidebar,
  a work area and a call panel, and a real URL per view. Neither reloads a
  document on navigation. Nobody using them experiences "not multipage".
- What actually was missing is the **information architecture**: a grouped
  sidebar, a persistent workspace, and a call dock. Five owner asks for
  "multipage with a sidebar nav" were read as five asks to reload documents,
  and the rail question (2026-11-11 restyle report, §g.3) was left open.

D1 is therefore **superseded**, not obeyed. The reload constraint survives as
an implementation detail (HTMX swap + pushState), exactly as it does in the
reference products.

## 2. Decision

**The shell is rail | workspace | phone dock.**

- **Rail (left, 216px):** four labelled groups (Workspace, Calls, People,
  System) with one glyph per destination and unread badges, the active row a
  filled accent pill with a leading bar. Compacts to 72px via a foot toggle
  (stored in `localStorage`, applied before first paint by
  `theme-preload.js`); auto-compacts below 1200px.
- **Workspace (centre):** `#tab-content`, unchanged mechanics: every tab is a
  real route (`/messages`, `/voicemail`, `/fax`, `/history`, `/reports`,
  `/contacts`, `/settings`, `/admin`), navigated by an HTMX partial swap with
  `hx-push-url`, so back/forward and deep links work and the island never
  unloads.
- **Phone dock (right):** the existing island column, now the call panel the
  way a contact-centre workspace has one; the island collapse and its
  live-call force-expand are unchanged.
- **Phones:** the SAME `<nav>` markup becomes the fixed bottom bar
  (`display: contents` on the rail wrapper); content stays first, the rail
  toggle is absent, no second nav tree exists to drift.

Layout templates live as tokens (`--grid-full`, `--grid-no-island`,
`--grid-door`, ...) because prettier breaks long `minmax()`/`var()` chains
mid-function and an unreadable grid is an unmaintainable one.

## 3. Shipped in this train

| File                                     | Change                                                          |
| ---------------------------------------- | --------------------------------------------------------------- |
| `internal/web/views/layout.templ`        | rail aside, grouped `NavLinks`, `wp-rail-toggle`, dock reorder  |
| `internal/web/views/i18n.go`             | group labels + rail toggle, en + de                             |
| `internal/web/assets/app.css`            | rail/workspace/dock grid, compact rail, mobile bottom-bar reuse |
| `internal/web/assets/theme-preload.js`   | pre-paint `wp-rail-collapsed`                                   |
| `internal/web/assets/shell.js`           | rail toggle, persisted, aria-expanded in step                   |
| `internal/web/views/layout_rail_test.go` | rail/group/route/signed-out pins                                |

Two pre-existing red tests were repaired on the way (both confirmed red at the
pre-train commit): the settings agent-row assertion matched the shortcut table's
"Cycle the AI drafts mode", and the shortcut-row count still said 10 after the
`D` shortcut landed.

## 4. The IA ladder (next trains, in order)

1. **T2 - workspace pages.** Per-page header (section title + one primary
   action), master-detail for Voicemail and Fax (Messages already has it), a
   letter index for Contacts. This is what makes each destination feel like a
   page rather than a panel.
2. **T3 - call dock parity.** Presence pill in the rail foot
   (available / on-call / do-not-disturb), a live-call workspace with
   disposition + wrap-up + notes, recent-call context beside the active call.
3. **T4 - omnichannel queue.** One ordered queue across messages / voicemail /
   fax with per-channel filters and counts already in the rail.
4. **T5 - analytics workspace.** Reports becomes a real page (range picker,
   per-contact drill-down) on the existing data.

Explicitly **not** on the ladder: document reloads between views. The island
must stay mounted, and no reference product does otherwise.

## 5. What this does NOT do

- No queue/presence product logic yet (T3/T4): this train is the shell and IA.
- No visual regression capture: the capture harness is owned by a parallel
  front-door session and was mid-edit; the shell structure is pinned by tests
  instead.
- Not deployed: rollout is the usual ladder (webphone push + CI, telephony
  lock bump, pbx-artmann relock, owner switch).
