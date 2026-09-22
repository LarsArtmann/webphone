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
whoami-ext
logout
dial-form
dest
call-btn
dial-error
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
ice-wrap
ice-panel
log
toasts
remote-audio
lang
<!-- dom-contract:end -->
