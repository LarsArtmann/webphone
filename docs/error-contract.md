# Error contract — the failure → feedback map

What the user SEES when X fails (2026-09-20 train, plan T13). Every
error path lands in at least one VISIBLE surface (toast, inline
banner, or panel); `#log` is always the operator trail, never the
only user feedback.

| Failure                                     | User sees                                                                                                                                              | Owner (wording)                                     | Test home                                                       |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------- | --------------------------------------------------------------- |
| Tab session dead (401 on tab actions)       | Throttled error toast: "Tab session ended; calls keep working…" — never auto-reload                                                                    | shell.js §3c (English, D3)                          | shell.test.mjs + `TestShellJSSurfacesHtmxErrors`                |
| Validation mistake (422)                    | `.wp-error` banner swapped into `#wp-tab-error` + error toast                                                                                          | server (en/de via `h.T`)                            | `TestSendClassifiesGatewayOutageAs502` + renderPanelError tests |
| Provider refused the send (422, its reason) | Banner + toast carry the provider's own refusal text; saved as failed, NO retry affordance                                                             | server i18n `err.messageRejected`/`err.faxRejected` | messages_test refusal arm + `TestFaxProviderRefusalAnswers422`  |
| Rate limited (429)                          | Toast: "Too many requests — wait a moment…"                                                                                                            | shell.js (htmx) / island i18n (login)               | shell.test.mjs + session.test.mjs                               |
| Gateway outage (502)                        | Banner + toast; message/fax saved as failed; retry affordance offered (also the arm for provider 5xx ANSWERS since train E)                            | server                                              | 502 test (HX-Trigger + banner pinned)                           |
| Other 4xx/5xx on htmx actions               | Banner (`.wp-error` selected) + generic toast "HTTP N"                                                                                                 | shell.js generic copy                               | contract test (config markers)                                  |
| Network down (htmx)                         | Toast: "Network request failed…"                                                                                                                       | shell.js                                            | shell.test.mjs                                                  |
| Network down (island session POST)          | Toast: "Could not reach the server…" + `#log` line                                                                                                     | island i18n `sessionNetFailed`                      | session.test.mjs                                                |
| PBX rejects login (401 island REGISTER)     | `loginError` inline + `reg-status` pill "registration rejected"                                                                                        | island i18n `loginError`/`regRejected`              | i18n parity tests                                               |
| SSE feed dead (3 consecutive errors)        | One warn toast + pill label flips to "not connected"                                                                                                   | island i18n `sseDropped`/`ssePillDown`              | session.test.mjs                                                |
| CRM call journal failed (502 / network)     | Warn toast: "Call not recorded in the CRM (…) — you can log it manually." en/de                                                                        | island i18n `crmLogFailed`                          | crm_test.go 502 contract + i18n parity tests                    |
| Contact write failed (JSON API 500)         | Warn toast: "contact save failed (…) ; kept locally" / "Kontakt speichern fehlgeschlagen (…); lokal behalten" — the same 500 text the tab banner shows | server `contactSaveFailed` (one home)               | `helpers_contract_test.go` (500 text + routing)                 |
| Store/render failure on a panel or thread load (500) | Body carries only the op prefix + family default ("load conversation: A temporary error occurred. …") — the raw store text goes to the operator log (`error=` + `family=`), never the browser | server `internalError`/`safeDetail` (SafeDetail) | `TestInternalErrorRedactsDetail` |
| Upload body over the 61 MB envelope (400)     | Plain-text 400 "could not read the form: http: request body too large" (English, operator-facing) — htmx surfaces the generic "HTTP 400" toast; legit uploads never see it | server `requireSessionMultipart` (MaxBytesReader)   | `TestUploadBodyIsBounded` + smoke "oversized upload rejected"    |
| Unknown path (404)                          | Styled 404 (shell + error panel), status stays 404                                                                                                     | server `error.notfound` en/de                       | `TestNotFoundRendersTheShell`                                   |
| Handler panic                               | Recovery middleware logs stack + re-raises; user gets htmx/browser failure surface                                                                     | cqrshtmx.RecoveryMiddleware                         | library + server middleware tests                               |

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

## Provider verdict hooks (ops note, 2026-09-23)

The outbound verdict hooks (`/hooks/{message,fax}/status`) share one
tail, `applyStatusWebhook`: the STATUS is validated BEFORE the ref
(consolidation flipped the old ref-first precedence), the 400 texts
are byte-stable ("status must be delivered or failed" /
"…transmitted or failed", "provider_ref is required") because
providers log them, a replayed ref answers 202 inertly, an unknown
ref is a 404 that names the lane ("could not update message: …") so
the provider stops retrying it, and any other apply failure is a 500
that stays retryable — only a SUCCESSFUL apply consumes the idem key
(`hooksIdem`, in-memory 1h). Pinned by `TestApplyStatusWebhookContract`
(`internal/server/helpers_contract_test.go`). The related JSON
mutation contract: contact mutations answer bare 204 (no body — the
island re-fetches, the list is the only id source), pinned by
`TestContactsAPIMutationsNudgeThen204` + the round-trip tests.

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
