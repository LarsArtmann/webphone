# UI/UX destination — "CCAI + Google Voice, but really good"

**Status:** DESTINATION (owner-ratified direction, 2026-11-11) · **Amended same day ~15:25:** the accent is GREEN (owner veto 15:15 — "my app is green, not Tailwind default"; indigo read as default-SaaS and is wrong for this product). Wherever an earlier draft of this file said indigo, green is the ruling. · **Author:** session with owner direction calls · **Execution:** Train 1 begins this session; Trains 2–3 are queued.

The bar is the OWNER'S EYES (owner ruling 2026-11-11): a train that leaves the look unchanged is a failed train, no matter what a compliance review says. Every shipped train on this file must pass one test — _side-by-side screenshots where anyone can point at the difference in under five seconds._

---

## 1. Owner direction calls (2026-11-11, this session)

| Call                       | Ruling                                                                                                                                                                                                                                                                                                              |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Product character          | **Hybrid with an operator-desk lean** — everyday surfaces (Messages, island) stay Google-Voice-clean; power surfaces (History, Reports, Admin, live-call context) get CCAI-grade density and information richness. When in doubt, density wins over airiness on power surfaces, calm wins on conversation surfaces. |
| Typography                 | **System stack only.** The CSP stays untouched (no webfonts, no CDN, ruling stands). The premium feel is manufactured from weight/size/tracking rhythm, tabular numerals, and spacing — not from a distinct typeface.                                                                                               |
| First train                | **Full visual pass** — every surface visibly restyled in one train, NO layout restructuring. Judgment rule: the owner must SEE the difference.                                                                                                                                                                      |
| Desktop layout destination | **Per-surface layouts under one shell** — Messages keeps master–detail, History becomes a scannable table, Reports becomes a dashboard, contacts/settings get their fitting layouts (Trains 2+).                                                                                                                    |

## 2. Reference DNA — what we actually take

**Google Voice** (consumer-clean half): conversation-first; quiet chrome that recedes behind content; list rows with a clear primary line, muted preview, right-aligned timestamp; owner-side bubbles in the accent, inbound bubbles neutral; pill-shaped navigation affordances; search as a first-class surface.

**CCAI Platform agent desktop** (operator-desk half; structure verified from Google's docs — visual presentation is our synthesis): a panel workspace where the active interaction, the customer context, and assistive knowledge are all visible at once; session tabs; transcripts as first-class artifacts; density that stays scannable. We take the _posture_ — context-at-a-glance, disciplined panels, information-dense power surfaces — not a clone of any specific screenshot.

**Material Design 3** (craft vocabulary; from working knowledge of the public spec): surface-container tiers instead of border-heavy boxes; state layers (hover ≈ 8%, focus/pressed ≈ 10–12% overlays) instead of color swaps; a shape scale from pill (chips/badges) through medium (controls) to large (cards); elevation as a shadow ladder, not one shadow everywhere; an emphasized easing curve for motion; a type ramp where every step pairs a size with a weight and tracking.

## 3. Current-state gap list (measured on the 2026-11-11 14:12 capture set)

1. **Type has no hierarchy.** Everything sits in one weight band; headings are just bigger; timestamps and counts use proportional numerals; the all-caps eyebrow ("CALL TRANSCRIPTS") reads as template chrome.
2. **Mobile thread rows strangle their own content.** Four icon buttons per row leave the preview a two-character stub at 390px.
3. **The island card is a foreign object** inside the shell: native selects, mismatched control styles, a hard red-bordered alert, floating EN/reconnecting pills above it.
4. **Empty states are dashed boxes** — a Bootstrap-era idiom, no icon, no action, no hierarchy.
5. **Buttons and rows lack state layers.** Hover/press/focus are flat color flips; focus rings are browser defaults.
6. **One shadow for everything** (panels, popovers, cards all share `--shadow`), so elevation carries no meaning.
7. **The header has no identity** — brand mark, two pills and a text-button theme toggle on a bare strip; the nav is plain underlined text.

## 4. Design pillars (permanent rules for every future train)

1. **Quiet chrome, loud content.** The shell (header, nav, island frame) recedes: neutral surfaces, hairlines, no decoration. Color belongs to content: bubbles, badges, state.
2. **One accent, tonal depth.** GREEN is the accent (the brand — owner veto 2026-11-11 15:15, same-day against the restyle's first indigo draft; indigo/blue reads as default-SaaS). Depth comes from surface tiers (`--surface` → `--surface-2` → `--surface-3`) and translucent state layers, never from new hues.
3. **Type is the hierarchy.** A real ramp (see §5); every heading pairs size with weight AND tracking; all data that ticks (timestamps, counts, durations) renders in tabular numerals; sentence case everywhere. The tiny small-caps section labels (CALL TRANSCRIPTS, Settings group heads) are functional dividers in the Google idiom and stay — the ban is on decorative eyebrows above content headings.
4. **The phone is an object.** The island is the product's hero: a crafted device card with a status identity (registration dot, live-call state), consistent controls, its own header strip — recognizably "the phone", not "another form panel".
5. **Density where the pro lives.** History, Reports, Admin and live-call context optimize for scanning: tighter rows, tabular numbers, aligned columns, more information per pixel — while Messages, Voicemail playback and the island stay calm.
6. **Motion answers action.** Transitions confirm state changes (tab cross-fades exist; control state layers get 150ms; panels get 250ms emphasized easing). Nothing moves that the user didn't cause; `prefers-reduced-motion` stays honored globally.

## 5. Token architecture destination

The token block stays mirrored between `app.css` and `island/style.css` (standing rule). Train 1 upgrades it in place:

- **Type ramp** (roles, not raw sizes at call sites): `--text-xs 12 / 13` caption+meta (muted, 400), `--text-sm 14` secondary, `--text-md 15` body, `--text-lg 17` emphasized body, `--text-xl 20 / 600 / -0.01em` panel titles, `--text-2xl 24 / 600 / -0.02em` page headings. Weights used: 400 body, 500 emphasis/controls, 600 headings — never lighter than 400, never bolder than 600 (system faces synthesize poorly beyond that; `font-synthesis: none` already guards).
- **Numerals**: `--numeric: font-variant-numeric: tabular-nums` applied to every timestamp, count, badge and metric — digits stop dancing as values tick.
- **Elevation ladder**: `--shadow-1` (resting cards: faint contact + tight ambient), `--shadow-2` (raised: nav, island, popovers), `--shadow-3` (overlays: palette, lightbox, toasts). Dark theme deepens spreads; light theme warms the umbra.
- **State layers**: `--state-hover: 8%`, `--state-focus: 10%`, `--state-pressed: 12%` (accent or ink alpha overlays on controls — M3 state-layer idiom, replacing flat color flips).
- **Shape scale**: `--radius-sm 8` (chips, inputs), `--radius 12` (controls, rows-cards), `--radius-lg 16` (panels, island), pill `999px` (badges, nav active pill). Bubbles get their own asymmetric scheme (§6).
- **Motion tokens**: `--dur-fast: 150ms`, `--dur: 250ms`, `--ease: cubic-bezier(0.2, 0, 0, 1)` (M3 emphasized easing); all component transitions reference the tokens.
- **Focus ring**: `--focus-ring` (2px accent, 2px offset) on every interactive element — visible, consistent, never the browser default.
- **Spacing rhythm**: keep `--panel-pad`/`--row-pad`; tighten row internals to a 4px grid so lists scan (8/12/16/24 steps only).

## 6. Per-surface destination

- **Shell header**: a slim elevated bar; refined brand lockup (mark + wordmark, tighter); identity pill with the DID; theme toggle restyled as a compact control with state icon; island toggle keeps its label. (Markup-light: CSS-first.)
- **Navigation (desktop)**: Google-apps-style pill tabs — active tab gets the tonal accent pill + 500 weight; hover gets the state layer; the underline idiom retires. Mobile bottom bar: elevated surface + blur backdrop, active item as accent pill, icon+label legible at 390px.
- **Island (the phone object)**: header strip with registration status dot + line identity; sections separated by hairlines, not nested boxes; selects and buttons rebuilt to match shell controls (no native select look); the offline banner becomes a tonal container (danger-soft fill, danger text, no hard border); EN + connection pills dock into the strip.
- **Messages**: rail rows get breathing room (14px pad, 3-line clamp preview, tabular timestamps, per-row actions as ghost icon buttons that solidify on hover/focus and shrink to the two essentials on mobile via CSS only); bubbles get the asymmetric radius scheme (owner: accent fill, 16px with 4px tail corner; inbound: `--surface-2` fill + hairline, mirrored tail) and grouped spacing (tight within a burst, airy between turns); composer becomes a single elevated field row with the send control as a filled accent pill.
- **History**: (Train 1 visual) rows densify with tabular numerals and aligned meta columns; (Train 2) full table layout with column heads.
- **Reports**: (Train 2) dashboard-grade cards; (Train 1) stat cards and meters get the tonal card treatment.
- **Voicemail / Fax**: rows adopt the unified list-row craft; voicemail player keeps its custom chrome but inherits the new tokens; fax timeline steps get the pill treatment.
- **Contacts / Settings**: grouped rows with hairline separators and 500-weight labels; destructive actions stay quiet until hovered.
- **Empty states**: retire dashed boxes — tonal circle with a content glyph, one sentence of "what lives here", and the action that fills it, all left-aligned in the list's own rhythm.
- **Login**: library-owned page; token flow-through via `login.css` only (grays stay Tailwind stock per the standing rule).

## 7. Trains

| Train                                   | Scope                                                                                                                                                | Acceptance                                                                                                                                                                                                            |
| --------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **1 — Full visual pass** — SHIPPED 2026-11-11 (capture evidence: `ui-shots/` desktop 1280x900 + mobile 390x844 runs, both themes; status report `docs/status/2026-11-11_16-44_ui-ux-destination-train1-visual-pass-status.md`) | §5 tokens + §6 chrome, island, list rows, bubbles, empty states, dark-mode parity. CSS-first; markup edits only where a pin-safe win justifies them. | 26-shot capture set before/after where every surface shows a pointable difference; all gates green (fmt, tw/login.css rebuild, tsc, island tests, view pins, smoke, buildflow); ink discipline WCAG note still holds. |
| **2 — Per-surface desktop IA**          | History table, Reports dashboard, contacts/settings layouts under the shared shell.                                                                  | Each tab's layout fits its content type; screenshots show restructures, not just paint.                                                                                                                               |
| **3 — Mobile dialer surface**           | Island-as-dialer on mobile (ring/dial reachable like an app tab), bottom-bar integration.                                                            | Ring-while-closed UX parity with the PWA-lite goals.                                                                                                                                                                  |

## 8. Fences (unchanged, restated because they bound every train)

CSP: same-origin only, no webfonts, no CDN, no inline `style`, no inline scripts. GREEN is the baseline accent (owner veto 2026-11-11 15:15; values in app.css/island style.css token blocks). Island modules serve verbatim inside main.js's import closure. The DOM contract (`docs/dom-contract.md`) governs ids. Before touching served markup: grep the three pin sites (Go tests, `scripts/webphone-smoke.py`, stack `tests/browser-e2e.py`). The 2026-11-11 green restyle is the floor — trains evolve it.
