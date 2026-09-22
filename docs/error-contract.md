# Error contract — the failure → feedback map

What the user SEES when X fails (2026-09-20 train, plan T13). Every
error path lands in at least one VISIBLE surface (toast, inline
banner, or panel); `#log` is always the operator trail, never the
only user feedback.

| Failure                                 | User sees                                                                           | Owner (wording)                        | Test home                                                       |
| --------------------------------------- | ----------------------------------------------------------------------------------- | -------------------------------------- | --------------------------------------------------------------- |
| Tab session dead (401 on tab actions)   | Throttled error toast: "Tab session ended; calls keep working…" — never auto-reload | shell.js §3c (English, D3)             | shell.test.mjs + `TestShellJSSurfacesHtmxErrors`                |
| Validation mistake (422)                | `.wp-error` banner swapped into `#wp-tab-error` + error toast                       | server (en/de via `h.T`)               | `TestSendClassifiesGatewayOutageAs502` + renderPanelError tests |
| Rate limited (429)                      | Toast: "Too many requests — wait a moment…"                                         | shell.js (htmx) / island i18n (login)  | shell.test.mjs + session.test.mjs                               |
| Gateway outage (502)                    | Banner + toast; message/fax saved as failed                                         | server                                 | 502 test (HX-Trigger + banner pinned)                           |
| Other 4xx/5xx on htmx actions           | Banner (`.wp-error` selected) + generic toast "HTTP N"                              | shell.js generic copy                  | contract test (config markers)                                  |
| Network down (htmx)                     | Toast: "Network request failed…"                                                    | shell.js                               | shell.test.mjs                                                  |
| Network down (island session POST)      | Toast: "Could not reach the server…" + `#log` line                                  | island i18n `sessionNetFailed`         | session.test.mjs                                                |
| PBX rejects login (401 island REGISTER) | `loginError` inline + `reg-status` pill "registration rejected"                     | island i18n `loginError`/`regRejected` | i18n parity tests                                               |
| SSE feed dead (3 consecutive errors)    | One warn toast + pill label flips to "not connected"                                | island i18n `sseDropped`/`ssePillDown` | session.test.mjs                                                |
| CRM call journal failed (502 / network) | Warn toast: "Call not recorded in the CRM (…) — you can log it manually." en/de     | island i18n `crmLogFailed`             | crm_test.go 502 contract + i18n parity tests                    |
| Unknown path (404)                      | Styled 404 (shell + error panel), status stays 404                                  | server `error.notfound` en/de          | `TestNotFoundRendersTheShell`                                   |
| Handler panic                           | Recovery middleware logs stack + re-raises; user gets htmx/browser failure surface  | cqrshtmx.RecoveryMiddleware            | library + server middleware tests                               |

Shell copy (toasts, dedup, throttle wording) stays ENGLISH by decision
D3 (2026-09-20): matches the `#log` operator-channel precedent; the
island's user-facing copy is fully en/de. Localizing shell copy only
if a tabs-style per-extension UX demand shows up.

## Send-failure UX layering (2026-09-22)

(plan `docs/planning/2026-09-22_16-07_SUPERB-send-failure-ux.md`):
the durable 4xx/5xx banner lands in `#wp-tab-error` ABOVE the tab
region while the reply composer swaps `show:window:bottom` — the
reason is durable but OFF-VIEWPORT, and the failed bubble in view
says only "failed". Shipped mitigations: every outbound send form
(new message, reply, fax) carries `hx-disabled-elt="find
button[type=submit]"` (double-submit guard — a live 40310 self-send
session produced TWO identical failed rows), and a self-thread
(remote == the config `identities` DID, both sides through
`ParsePhone` so config spacing cannot hide the match) renders the
`wp-notice` caution (en/de, `thread.selfNotice`). The durable
in-bubble failure story (persisted reason + kind, retry only where
retryable) is the planned follow-up in TODO_LIST.

## Rate-limit keying (ops note, 2026-09-22)

The login/hook/contacts limiters key on the PEER host
(`remoteHostKey`, port-stripped). Behind the consuming stack's
fronting nginx every browser shares the proxy's address — one rate
bucket for ALL users of a deployment. This is accepted while the
deployments are small; the widening to per-client keys
(`KeyExtractorFromClientIP`) is deliberately gated on the stack
proving X-Forwarded-For sanitization first (spoofable XFF would let
one client dodge the limiter by rotating a header). If a deployment
ever sees collective 429s on login, THIS is the first suspect —
check the limiter's key, not the users.

## Cross-repo sync

The operator-facing semantics of these surfaces — status meanings,
the gateway/bridge string contract (rendered verbatim in webhook
mode), and the `family=` log vocabulary — are cross-documented in the
stack runbook: `nix-international-telephony/docs/ops-runbook.md`
§ "Webphone error contract" (2026-09-22). Keep both sides in sync
when error copy or families change.
