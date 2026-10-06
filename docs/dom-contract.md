# The island DOM contract

The element ids every served full shell MUST carry. The consuming
stack's browser E2E (`tests/browser-e2e.py` in
nix-international-telephony) drives the island through them, and
`TestServedPageHoldsTheDomContract` (internal/server) parses THIS
file — the block between the marker comments below is the single
source of truth, so the test can never drift from the documented
list. Change the page and this list in the same commit; the test
fails otherwise.

Beyond the ids, the contract (asserted in the same test): the script
references (`/config.js`, `/assets/vendor/sip.min.js`,
`/assets/island/app/main.js`, `/htmx.min.js`), the `WebPhone` brand,
the toast host's live-region attributes (`#toasts` carries
`role="status" aria-live="polite"`), the durable error slot
`#wp-tab-error`, and the `htmx-config` meta. The E2E also greps
served island sources verbatim — see AGENTS.md "The DOM + bundle
contract".

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
lang
wp-live
wp-tab-skeleton

<!-- dom-contract:end -->

## Conditional ids (passkey mode)

These ids render ONLY when the deployment enables the passkey
(WebAuthn) login mode (`auth.passkey.*`); the always-on contract above
stays complete without them, and the E2E must not assume them:

- `passkey-login-form` — the email-first front door (island module
  `passkey.js` binds it; absent markup means the mode is off and the
  module no-ops)
- `passkey-email` — the email input
- `passkey-login-error` — the inline error slot

With the mode on, the extension login moves INSIDE a
`details.passkey-breakglass` disclosure (summary: "Use extension and
password instead") but keeps its own ids and behavior unchanged.

The standalone `/enroll` page (same gate) carries `enroll-view`,
`enroll-form`, `enroll-token`, `enroll-credential-name`,
`enroll-status`, `enroll-error` — a one-purpose surface with NO island
runtime; its module is `/assets/enroll/enroll.js`.

## Notes

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
