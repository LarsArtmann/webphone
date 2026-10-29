# The island DOM contract

The element ids every served full shell MUST carry. The consuming
stack's browser E2E (`tests/browser-e2e.py` in
nix-international-telephony) drives the island through them, and
`TestServedPageHoldsTheDomContract` (internal/server) parses THIS
file — the block between the marker comments below is the single
source of truth, so the test can never drift from the documented
list. Change the page and this list in the same commit; the test
fails otherwise.

Two ids are deliberately NOT mirrored in `types/island.d.ts`
(`wp-live`, `wp-tab-skeleton`) and one is markup-only (`call-btn`):
the first two are shell-owned affordances (shell.js must work when the
island scripts fail), and the last is driven by the shell's delegated
`data-dial` handler, never by a typed `getElementById`. The mirror
rule is one-directional — every id the island's type declarations
carry must appear here, not the reverse — and
`TestIslandDomIdsMirrorTheDomContract` (internal/web/assets) pins it.

Beyond the ids, the contract (asserted in the same test): the script
references (`/config.js`, `/assets/vendor/sip.min.js`,
`/assets/island/app/main.js`, `/htmx.min.js`), the `WebPhone` brand,
the toast host's live-region attributes (`#toasts` carries
`role="status" aria-live="polite"`), the durable error slot
`#wp-tab-error`, and the `htmx-config` meta. Unknown paths answer the
styled 404 (`section.wp-panel.wp-notfound`: the literal 404 in the
heading, the requested path echoed back, one link home, no
`data-reload`) while the status stays 404. The E2E also greps served
island sources verbatim — see AGENTS.md "The DOM + bundle contract".

<!-- dom-contract:begin -->

reg-status
login-view
login-form
login-error
ext
pass
remember
phone-view
offline-banner
whoami-ext
logout
dial-form
dest
call-btn
dial-error
wp-dial-hint
calls
keypad
incoming-call
incoming-from
incoming-queue
accept-btn
reject-btn
contacts-wrap
contacts-list
history-wrap
history-list
vm-wrap
vm-badge
vm-list
vm-refresh
vm-status
wp-adv-toggle
wp-advanced
wp-devtest-btn
ice-wrap
ice-panel
log
toasts
remote-audio
audio-out-wrap
audio-output
lang
wp-live
wp-tab-skeleton

<!-- dom-contract:end -->

## Conditional ids

Rendered only under the named seam; the always-on block above stays
complete without them. The fenced list below is the machine-readable
counterpart (the type-declaration mirror in `types/island.d.ts` is
checked against BOTH blocks) — keep it in step with the prose sections
that follow.

<!-- dom-contract-conditional:begin -->

sso-login
sso-error

<!-- dom-contract-conditional:end -->

## Conditional ids (SSO lane)

Rendered when the deployment configures the OIDC lane
(`auth.oauth2.*`): `sso-login` (the Pocket ID front-door link) and
`sso-error` (its failure line). The island's `ssoerror.js` writes into
the latter, which is why the id is declared in its typed lookups even
though the markup is conditional.

## Conditional ids (passkey mode)

These ids render ONLY when the deployment enables the passkey
(WebAuthn) login mode (`auth.passkey.*`); the always-on contract above
stays complete without them, and the E2E must not assume them:

- `passkey-login` — the `/login` front-door LINK on the island login
  card (the cqrs-htmx `loginpage` library page owns the whole ceremony;
  the card only navigates there — a plain anchor, works without
  scripts)

The passkey ceremony UI itself lives on the SEPARATE `/login` page
(same gate), rendered by the library: `lp-login-form`, `lp-email`,
`lp-login-btn`, `lp-error` (plus the `loginpage-config` JSON blob its
embedded nonce'd script reads). That markup is library-owned contract —
do not pin it beyond presence checks.

Settings passkey management (passkey mode on AND the signed-in
extension has a mapped account) renders `wp-passkeys-head`,
`wp-passkeys-empty` (mapped account, zero enrolled keys) and
`wp-passkeys-list` (the enrolled list with revoke buttons); absent
markup means the mode is off or the extension has no mapped account.

Settings agent-action history (D5; rows exist OR the textagent seam is
on) renders `wp-agent-audit-head` and either `wp-agent-audit-empty`
(no rows yet, seam on) or `wp-agent-audit-list` (the newest audit
rows); absent markup means the seam is off with no recorded history.

With the mode on, the extension login moves INSIDE a
`details.passkey-breakglass` disclosure (summary: "Use extension and
password instead") but keeps its own ids and behavior unchanged.

The standalone `/enroll` page (same gate) carries `enroll-view`,
`enroll-form`, `enroll-token`, `enroll-credential-name`,
`enroll-status`, `enroll-error` — a one-purpose surface with NO island
runtime; its module is `/assets/enroll/enroll.js`.

## Conditional ids (textagent seam)

These ids render ONLY when the deployment configures the supervised
text-agent seam (`textagent.url` + `textagent.secret_file`, auto-wired
by the stack's `services.telephony.textagent`); the always-on contract
above stays complete without them, and the E2E must not assume them:

- `wp-drafts` — the review panel section (its own 30 s poll at
  `GET /partials/messages/drafts` swaps it outerHTML)
- `wp-draft-count` — the pending-count badge
- `wp-draft-<id>` — one article per pending AI draft (Approve / Edit /
  Discard ride `POST /messages/drafts/<id>/…`)
- `wp-ai-draft` — the composer's on-demand-draft button class (both
  the new-conversation form and the reply composer)
- `wp-draft-mode` — the drafts panel's auto-reply mode selector
  (`POST /messages/drafts/mode` pins a runtime override; a
  `value="reset"` submit returns to the configured mode)

## Notes

### Layout ids (served, but NOT required — 2026-10-29)

These are in the served bytes of every signed-in shell but are deliberately
absent from the pinned block above: the block is "ids every served full shell
MUST carry", and a contract that grows every time a surface gains an
affordance stops being a contract. They are documented here so the E2E author
can find them:

- `wp-island` — the phone column (`<aside>`). Its width is the shell's
  `--island-w-*` token range; it is `display: none` while the root carries
  `wp-island-collapsed`, EXCEPT when it holds a live `.call-card` (a call
  always re-expands the column).
- `wp-island-toggle` — the header button that collapses/expands the phone
  column (rendered only when signed in; signed out the island IS the login
  card). `aria-controls="wp-island"`; `aria-expanded` is the state channel and
  is re-asserted on load so it matches the pre-paint class.
- `wp-conversation-rail` — the Messages master pane (`<nav aria-label>`); one
  `a.wp-rail-row` per active conversation, the open one carrying
  `aria-current="page"`. Hidden below the `44rem` container threshold, where
  the conversation's "All messages" link is the way back.
- `wp-msg-detail` — the Messages conversation column (head + transcript +
  composer). The panel keeps its id `messages-panel`, which now names the
  master–detail CONTAINER rather than either pane.

Behaviour/rationale: AGENTS.md § Architecture invariants, first bullet.

### Runtime-created ids (NOT in the served-bytes block above)

The multi-call train (2026-10-26) adds ids the island creates at
runtime — they can never appear in the served bytes, so they stay out
of the pinned block and are documented here for the E2E and future
sessions:

- `wp-call-dock` — the call dock strip (JS-created when a second call
  goes live; removed again below two live calls)
- `wp-dock-chip-<callId>` — one dock chip per live call (stable per
  call, click = hold & take)
- `wp-captions-<callId>` — the caption strip wrapper (aria-hidden) on
  an ASR-enabled call card; the inner `call-transcript-<callId>` node
  carries the bounded line scrollback (`wp-caption-line` children)
- `wp-incoming-slot-<n>` — ringing-queue slots 2..N inside
  `#incoming-queue` (slot 1 IS the pinned `#incoming-from` /
  `#accept-btn` / `#reject-btn` markup above)

### Greppable SSE row classes on the templ-components v1.20.1 tree (2026-10-06)

`wp-thread-row` / `wp-bubble` (plus `wp-thread-rowwrap`, `wp-bubble-body`,
`wp-bubble-meta`) are the greppable row/bubble classes the E2E and the
AGENTS contract rely on for SSE fragments. They are authored in
`internal/web/views/messages.templ` and land in the served bytes through
the generated `messages_templ.go`. Verified on the v1.20.1 tree by
booting a fresh loopback binary, logging in, sending an outbound
message, and byte-grepping the SERVED payloads (not the source): the
thread list fragment contains `wp-thread-row` and the transcript page
contains `wp-bubble`. The templ-components v1.20.0 → v1.20.1 ride
therefore does not disturb the fragment contract; this note records the
coverage reasoning so future bumps can re-run the same probe.
