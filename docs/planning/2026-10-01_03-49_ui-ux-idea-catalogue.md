# UI/UX Idea Catalogue

**Date:** 2026-10-01
**Status:** Idea bank — not a commitment, not ordered as a backlog for execution
**Scope:** The whole served surface — the persistent SIP island and the six
HTMX tabs (Messages, Fax, Voicemail, History, Contacts, Settings), the shell,
the CSS/token system, and the island JS modules.
**Purpose:** A sorted, tagged reservoir of improvements to pull from when
planning a UI/UX train. Nothing here is ratified.

## How to read this

Each idea carries three tags:

- **Priority** — `Now` (clear win, low risk), `Next` (worth doing once a train
  exists), `Later` (nice, bigger or speculative).
- **Effort** — `S` (hours), `M` (a day or two), `L` (multi-day / cross-repo).
- **Constraint** — anything the idea collides with or depends on. Left off when
  none. The constraints that matter in this repo:
  - `CSP` — no inline scripts or `style` attributes are allowed; all new
    behavior lives in `shell.js` / island modules, all new styles in
    `app.css` or the scoped `tw.css` build.
  - `MORPH` — stateful nodes inside morph-swapped surfaces need a stable `id`
    (idiomorph persists by id); focus/drafts/audio only survive with one.
  - `DOM` — new ids/classes greppable by the consuming stack's browser E2E;
    changing them means updating `docs/dom-contract.md` and re-running that
    suite.
  - `TW` — if it needs Tailwind utilities, they come from the adopted-component
    `tw.css` set, rebuilt with `nix run nixpkgs#tailwindcss_4`.
  - `E2E` — needs the stack browser E2E to prove it on the wire.
  - `SEAM` — needs a new server/provider capability, not just presentation.

## At a glance

| # | Theme                          | Ideas | Highest-leverage cluster          |
| - | ------------------------------ | ----- | --------------------------------- |
| A | Calls & the island             | 10    | live call controls + name-on-dial |
| B | Messaging                      | 20    | delivery state + optimistic send  |
| C | Voicemail & Fax                | 10    | waveform playback + fax preview   |
| D | Contacts, directory & history  | 10    | search + section jumplist         |
| E | Navigation, layout & shell     | 10    | command palette + URL-state       |
| F | Visual design & theming        | 10    | icon set + skeletons              |
| G | Accessibility                  | 10    | swap focus + live announcements   |
| H | Keyboard & power users         | 10    | shortcut help + global keys       |
| I | Mobile & responsive            | 10    | call overlay + bottom tabs        |
| J | Feedback, trust & errors       | 10    | rollback + reconnect banner       |
| K | Onboarding & discovery         | 5     | first-run tour + demo mode        |
| L | Internationalization & content | 5     | locale formatting + RTL           |

**Total: 120 ideas.** Of these, 36 are `Now`/S and form the Pareto shortlist at
the end.

---

## A. Calls & the island

Ordered by impact: the island is the product, and today it is keyboard- and
log-driven rather than visibly stateful.

| ID  | Idea                                                                                                                           | Priority | Effort | Constraint      |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | -------- | ------ | --------------- |
| A1  | Active-call control cluster: explicit, labeled Mute / Hold / Hangup / Keypad buttons on the live call card, not keyboard-only. | Now      | M      | —               |
| A2  | Call timer on the active call card (elapsed duration), so the user knows the call is live.                                     | Now      | S      | —               |
| A3  | Ringback state: a _calling…_ pulse on the remote avatar instead of only a `#log` line.                                         | Now      | S      | CSP (JS-driven) |
| A4  | Name-as-you-type resolution: show the CRM/contact name under `#dest` before dialing.                                           | Next     | S      | —               |
| A5  | Number normalization hint: surface how `+`, spaces, and country codes will be sent while typing in `#dest`.                    | Next     | S      | —               |
| A6  | Incoming-call focus mode: dim everything but the incoming card and move focus to Accept.                                       | Next     | M      | a11y            |
| A7  | Missed-call badge on the island header, cleared on view.                                                                       | Next     | M      | —               |
| A8  | DTMF feedback: animate the pressed key and show a transient list of tones sent.                                                | Next     | S      | —               |
| A9  | Re-dial affordance on the last-call row (island history + tab history).                                                        | Next     | S      | —               |
| A10 | Pre-call media check: a button that runs an ICE/audio-device test before the first call.                                       | Later    | M      | —               |

## B. Messaging

The composer is already strong (segment count, morph-safe search). The gap is
the _state_ of a message and the richness of an MMS.

| ID  | Idea                                                                                             | Priority | Effort | Constraint           |
| --- | ------------------------------------------------------------------------------------------------ | -------- | ------ | -------------------- |
| B1  | Delivery state per outbound bubble (sent / delivered / failed) fed by the existing status hooks. | Now      | M      | MORPH                |
| B2  | Optimistic bubble: render the sent message instantly with a pending style.                       | Now      | M      | MORPH                |
| B3  | Retry affordance on a failed bubble (tap to resend).                                             | Now      | S      | MORPH                |
| B4  | Date separators in the transcript ("Today", "Yesterday", weekday).                               | Now      | S      | —                    |
| B5  | Unread divider inside the transcript ("— 3 unread —").                                           | Next     | S      | —                    |
| B6  | Inline MMS image preview with a lightbox, instead of a bare attachment.                          | Next     | M      | —                    |
| B7  | Attachment-type preview in the thread list ("📷 Photo").                                         | Next     | S      | —                    |
| B8  | Over-limit countdown on the segment counter (existing `.wp-segcount`, currently just a count).   | Next     | S      | —                    |
| B9  | Snippet/template picker insertable into the composer.                                            | Next     | M      | SEAM (snippet store) |
| B10 | Quick-reply chips above the composer driven by those snippets.                                   | Later    | S      | SEAM                 |
| B11 | Send-on-Enter / newline-on-Shift-Enter, with a setting.                                          | Next     | S      | —                    |
| B12 | Per-thread draft persistence across reloads.                                                     | Next     | M      | —                    |
| B13 | Thread pinning / starring.                                                                       | Next     | M      | SEAM (store)         |
| B14 | Thread archiving as a distinct action from delete.                                               | Later    | M      | SEAM (store)         |
| B15 | Per-thread mute controlling the badge.                                                           | Later    | M      | SEAM (store)         |
| B16 | Mark-all-read on the messages panel.                                                             | Next     | S      | —                    |
| B17 | Copy number / copy message affordance on rows and bubbles.                                       | Next     | S      | —                    |
| B18 | Deep link from a notification straight into the exact thread.                                    | Later    | M      | MORPH                |
| B19 | Schedule a message for later (provider permitting).                                              | Later    | L      | SEAM                 |
| B20 | Grouped day headers with per-day message counts in long threads.                                 | Later    | S      | —                    |

## C. Voicemail & Fax

| ID  | Idea                                                                                                 | Priority | Effort | Constraint |
| --- | ---------------------------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| C1  | Inline waveform scrubber for voicemail instead of the native `<audio>` control.                      | Next     | M      | MORPH      |
| C2  | Playback speed control (0.5–2×) on voicemail.                                                        | Now      | S      | MORPH      |
| C3  | Explicit "playing" highlight on the row, so a morph swap can't look like playback stopped.           | Now      | S      | MORPH      |
| C4  | Fax first-page thumbnail in the fax list (PDF render).                                               | Next     | M      | —          |
| C5  | Fax send-status timeline (queued → sending → delivered / failed).                                    | Next     | M      | —          |
| C6  | Fax resend button on failed rows.                                                                    | Next     | S      | —          |
| C7  | Bulk select + "download all" for faxes.                                                              | Later    | M      | —          |
| C8  | Auto-transcription line under each voicemail (needs an ASR seam).                                    | Later    | L      | SEAM       |
| C9  | Voicemail unread badge clears on play and mirrors to the nav badge.                                  | Next     | S      | —          |
| C10 | Voicemail row as one card with play + call-back + SMS (consolidates today's scattered mini-buttons). | Next     | M      | —          |

## D. Contacts, directory & history

> **Boundary ruling (2026-10-01):** contact management lives in Ledger
> (~/projects/crm), not here. D4–D7 route to Ledger's roadmap; D1–D3
> stay rejected for the same reason — webphone's personal contacts are a
> dialing scratchpad, and the CRM seam (enrichment + call journal) is
> the integration surface. D8–D10 (history) are unaffected.

| ID  | Idea                                                                               | Priority | Effort | Constraint   |
| --- | ---------------------------------------------------------------------------------- | -------- | ------ | ------------ |
| D1  | Search input at the top of the Contacts panel.                                     | Now      | S      | —            |
| D2  | Alphabetical section headers / jumplist for long contact lists.                    | Next     | M      | —            |
| D3  | Inline contact editing instead of delete + recreate.                               | Now      | M      | SEAM (store) |
| D4  | Merge duplicate contacts (same number, different names).                           | Later    | M      | SEAM         |
| D5  | Favorites / speed-dial row pinned at the top.                                      | Next     | M      | SEAM         |
| D6  | Contact detail drawer with that number's calls, messages, and faxes.               | Later    | L      | SEAM         |
| D7  | Single-contact vCard export (today only whole-list export exists).                 | Next     | S      | SEAM         |
| D8  | History filters: inbound / outbound / missed, date range, per number.              | Next     | M      | —            |
| D9  | History grouped by day with per-day counts.                                        | Next     | S      | —            |
| D10 | Channel + outcome icons on each history row (call / SMS / fax, answered / missed). | Next     | S      | —            |

## E. Navigation, layout & shell

| ID  | Idea                                                                            | Priority | Effort | Constraint |
| --- | ------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| E1  | Global command palette (Ctrl/Cmd-K): jump to tab, thread, contact, or number.   | Now      | M      | CSP        |
| E2  | URL-addressable panel state for filters (`/messages?q=`, `/history?missed=1`).  | Next     | M      | —          |
| E3  | Remember last-active tab and restore it on reload.                              | Next     | S      | —          |
| E4  | Skip-to-content link for keyboard users.                                        | Now      | S      | —          |
| E5  | Collapse/expand the island on desktop so tabs reclaim width.                    | Next     | M      | —          |
| E6  | Resizable sidebar with persisted width (localStorage).                          | Later    | M      | —          |
| E7  | Breadcrumb in the tab header for deep views (Messages › Thread).                | Next     | S      | DOM        |
| E8  | Sticky nav that keeps the active tab visible when the strip scrolls.            | Next     | S      | —          |
| E9  | "Recent activity" summary strip under the header (new SMS, VM, missed).         | Next     | M      | —          |
| E10 | Empty-root state that orients a first-time user instead of only a sign-in hint. | Next     | S      | —          |

## F. Visual design & theming

| ID  | Idea                                                                                    | Priority | Effort | Constraint |
| --- | --------------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| F1  | Skeleton loaders on partial swaps instead of a blank beat.                              | Now      | M      | —          |
| F2  | Subtle `#tab-content` swap transition (fade/slide) honoring reduced-motion.             | Now      | S      | CSP, MORPH |
| F3  | Consistent icon set replacing mixed emoji (`📞`, `☆`) for row actions.                  | Now      | M      | DOM        |
| F4  | State-color tokens (success / warning / danger) unified across badges, banners, toasts. | Next     | S      | —          |
| F5  | Per-tab accent color so each service is recognizable at a glance.                       | Next     | M      | TW         |
| F6  | Density toggle (comfortable / compact) for message and history lists.                   | Next     | M      | —          |
| F7  | Theme toggle shows an icon + current state, not just text ("Theme: auto").              | Now      | S      | —          |
| F8  | Brand empty-art set for the empty states, not only icons.                               | Later    | M      | —          |
| F9  | Dark-mode contrast audit of the avatar hue classes.                                     | Now      | S      | —          |
| F10 | Optional rounded vs sharp visual theme as a token switch.                               | Later    | L      | TW         |

## G. Accessibility

The morph surfaces create real focus/announcement gaps — this theme is
disproportionately high-value.

| ID  | Idea                                                                                           | Priority | Effort | Constraint |
| --- | ---------------------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| G1  | Focus management on partial swap: move focus to the new panel heading.                         | Now      | M      | MORPH      |
| G2  | Live-region announcements for unread-count changes and new-message pushes.                     | Now      | M      | MORPH      |
| G3  | `aria-current="page"` on the active nav link.                                                  | Now      | S      | —          |
| G4  | Announce badges as "3 unread", not a bare "3".                                                 | Now      | S      | —          |
| G5  | `aria-describedby` linking `#dial-error` and `#login-error` to their forms.                    | Now      | S      | —          |
| G6  | Keyboard path through the keypad (roving tabindex + arrow keys).                               | Next     | M      | —          |
| G7  | Contrast pass on `.wp-muted` and `.hint` text.                                                 | Now      | S      | —          |
| G8  | Extend `prefers-reduced-motion` beyond the current single block (toasts, pulses, transitions). | Next     | S      | —          |
| G9  | `forced-colors` (high-contrast) support for badges and borders.                                | Later    | M      | —          |
| G10 | Theme toggle exposes its current state to screen readers.                                      | Now      | S      | —          |

## H. Keyboard & power users

| ID  | Idea                                                                       | Priority | Effort | Constraint |
| --- | -------------------------------------------------------------------------- | -------- | ------ | ---------- |
| H1  | Shortcut help overlay (`?`), since shortcuts only surface in `#log` today. | Now      | M      | CSP        |
| H2  | Shortcut cheat-sheet in the Settings tab.                                  | Now      | S      | —          |
| H3  | Global `n` = new message, `/` = focus search, `g m` = go to Messages.      | Next     | M      | CSP        |
| H4  | Alt+1..6 to jump between the six tabs.                                     | Next     | S      | —          |
| H5  | Arrow-key navigation in the thread list, Enter to open.                    | Next     | M      | —          |
| H6  | `j` / `k` prev/next thread inside a conversation.                          | Later    | S      | —          |
| H7  | Escape closes the thread view back to the list.                            | Next     | S      | —          |
| H8  | Vim-style `q` to go back.                                                  | Later    | S      | —          |
| H9  | Global search shortcut that works from any tab.                            | Next     | M      | —          |
| H10 | Alt+A / Alt+H call keys, to avoid media-key collisions on some OSes.       | Later    | S      | —          |

## I. Mobile & responsive

The phone is the primary device for a phone app; this theme currently has the
fewest affordances.

| ID  | Idea                                                                         | Priority | Effort | Constraint |
| --- | ---------------------------------------------------------------------------- | -------- | ------ | ---------- |
| I1  | Full-screen call overlay on mobile (the island shrinks awkwardly today).     | Now      | M      | —          |
| I2  | Bottom tab bar on narrow viewports instead of horizontal scroll.             | Now      | M      | DOM        |
| I3  | Dedicated mobile action bar (call / message) on contact and history rows.    | Next     | M      | —          |
| I4  | One-tap autofill audit for extension/password (`inputmode`, `autocomplete`). | Now      | S      | —          |
| I5  | Enforce ≥44px tap targets on `.wp-mini` buttons.                             | Now      | S      | —          |
| I6  | Swipe-to-delete on list rows.                                                | Next     | M      | —          |
| I7  | Pull-to-refresh on the SSE-backed lists.                                     | Later    | M      | —          |
| I8  | Safe-area insets for notched devices around the island and toasts.           | Next     | S      | —          |
| I9  | Collapse the header on scroll to reclaim vertical space.                     | Later    | S      | —          |
| I10 | Single-column stack that reorders the island above the tabs on phones.       | Now      | M      | —          |

## J. Feedback, trust & errors

| ID  | Idea                                                                                   | Priority | Effort | Constraint |
| --- | -------------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| J1  | Optimistic-UI rollback visuals when a send fails (pairs with B2).                      | Now      | M      | MORPH      |
| J2  | "Reconnecting…" banner when SSE drops, with an auto-recovery notice.                   | Now      | M      | —          |
| J3  | Undo window on deletes instead of a modal confirm.                                     | Next     | M      | —          |
| J4  | Retry / copy-detail affordance inside every error banner.                              | Next     | S      | DOM        |
| J5  | Toast queue with explicit dismiss and defined stacking rules.                          | Next     | S      | —          |
| J6  | Relative-time refresh so "2m ago" doesn't freeze.                                      | Now      | S      | —          |
| J7  | Consistent confirm dialogs (today native `hx-confirm` mixes with styled surfaces).     | Later    | M      | —          |
| J8  | Inline spinner pinned to the submitted button (extend the existing `hx-disabled-elt`). | Next     | S      | —          |
| J9  | Success pulse when a contact saves or a message sends.                                 | Next     | S      | —          |
| J10 | Per-service status dots in Settings (gateway / phone-API / CRM reachability).          | Later    | M      | SEAM       |

## K. Onboarding & discovery

| ID | Idea                                                                                  | Priority | Effort | Constraint |
| -- | ------------------------------------------------------------------------------------- | -------- | ------ | ---------- |
| K1 | First-run tour highlighting the island, the tabs, and the shortcuts.                  | Next     | M      | CSP        |
| K2 | Contextual hints on the composer and dial field that vanish once used.                | Next     | S      | —          |
| K3 | Demo/sample mode on the loopback gateway so the app is explorable with zero PBX.      | Now      | M      | —          |
| K4 | A "what this can do" panel on the empty root.                                         | Next     | S      | —          |
| K5 | Progressive disclosure of the island's advanced `<details>` (ICE, log) behind a gear. | Next     | S      | —          |

## L. Internationalization & content

| ID | Idea                                                                           | Priority | Effort | Constraint |
| -- | ------------------------------------------------------------------------------ | -------- | ------ | ---------- |
| L1 | Locale-aware date/number formatting (today fixed helpers).                     | Next     | M      | —          |
| L2 | Per-user language persisted server-side, not only the `wp-lang` cookie.        | Next     | M      | SEAM       |
| L3 | Extend beyond en/de with a picker that does not require a reload.              | Later    | L      | —          |
| L4 | RTL readiness via logical CSS properties, even before an RTL locale ships.     | Later    | L      | —          |
| L5 | Optionally translate `#log` diagnostics while keeping English for the runbook. | Later    | M      | —          |

---

## Pareto shortlist — 36 `Now`/S items

The cheapest, highest-signal cluster. Grouped by the outcome they buy:

**Make calls feel alive (A)**
A1, A2, A3

**Make messaging trustworthy (B, J)**
B1, B2, B3, B4, J1, J6

**Make voicemail/fax usable (C)**
C2, C3

**Make lists navigable (D)**
D1, D3

**Make the shell smart (E)**
E1, E4

**Make it feel finished (F)**
F1, F2, F3, F7, F9

**Close the a11y gaps the morph creates (G)**
G1, G2, G3, G4, G5, G7, G10

**Give power users a handle (H)**
H1, H2

**Make the phone a phone (I)**
I1, I2, I4, I5, I10

**Make first contact good (K)**
K3

### Suggested first train (if one is picked)

A vertical slice that proves the pattern end-to-end, then fans out:

1. **Trust slice** — B2 + B3 + B1 + J1 (optimistic send, retry, delivery
   state, rollback). Proves the `MORPH` pattern for server-confirmed state.
2. **A11y slice** — G1 + G2 + G3 + G4 (swap focus + announcements). Proves
   the morph focus contract before more morph surfaces are added.
3. **Shell slice** — E1 + H1 + H2 (palette + shortcut help). Proves the `CSP`
   pattern for a JS-heavy overlay with no inline code.

Each slice should update `docs/dom-contract.md` where ids change and re-run the
stack browser E2E (`E2E`) before it is called done.

---

## Open questions

- Do we treat this as input to `ROADMAP.md` (raw ideas) or promote clusters
  into `TODO_LIST.md` (bounded tasks)? The catalogue deliberately stays
  neutral.
- Which `SEAM` items (B9, B13–B15, B19, C8, D3–D7, D10, J10, L2) are worth a
  server-side design before any UI work, versus deferred?
- Is mobile (`I`) a first-class target or a graceful-degradation target? That
  decision reorders roughly a third of this file.
