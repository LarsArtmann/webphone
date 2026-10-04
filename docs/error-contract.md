# Error contract — the failure → feedback map

What the user SEES when X fails (2026-09-20 train, plan T13). Every
error path lands in at least one VISIBLE surface (toast, inline
banner, or panel); `#log` is always the operator trail, never the
only user feedback.

| Failure                                              | User sees                                                                                                                                                                                                                                                     | Owner (wording)                                                   | Test home                                                                                                                             |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Tab session dead (401 on tab actions)                | Throttled error toast: "Tab session ended; calls keep working…" — never auto-reload                                                                                                                                                                           | shell.js §3c (English, D3)                                        | shell.test.mjs + `TestShellJSSurfacesHtmxErrors`                                                                                      |
| Validation mistake (422)                             | `.wp-error` banner swapped into `#wp-tab-error` + error toast                                                                                                                                                                                                 | server (en/de via `h.T`)                                          | `TestSendClassifiesGatewayOutageAs502` + renderPanelError tests                                                                       |
| Provider refused the send (422, its reason)          | Banner + toast carry the provider's own refusal text; saved as failed, NO retry affordance. Self-sends to the own DID (config identities) ride the same arm with a locally synthesized refusal (train C) — same 422, same failed row, zero provider roundtrip | server i18n `err.messageRejected`/`err.faxRejected`               | messages_test refusal arm + `TestFaxProviderRefusalAnswers422` + `TestThreadViewWarnsOnSelfSend`/`TestFaxToOwnNumberIsRefusedLocally` |
| Rate limited (429)                                   | Toast: "Too many requests — wait a moment…"                                                                                                                                                                                                                   | shell.js (htmx) / island i18n (login)                             | shell.test.mjs + session.test.mjs                                                                                                     |
| Gateway outage (502)                                 | Banner + toast; message/fax saved as failed; retry affordance offered (also the arm for provider 5xx ANSWERS since train E)                                                                                                                                   | server                                                            | 502 test (HX-Trigger + banner pinned)                                                                                                 |
| Other 4xx/5xx on htmx actions                        | Banner (`.wp-error` selected) + generic toast "HTTP N"                                                                                                                                                                                                        | shell.js generic copy                                             | contract test (config markers)                                                                                                        |
| Network down (htmx)                                  | Toast: "Network request failed…"                                                                                                                                                                                                                              | shell.js                                                          | shell.test.mjs                                                                                                                        |
| Version probe (`GET /version`)                       | JSON: `version` (+ `commit`, `commitDate` when known — ldflags rev/dirtyRev + RFC 3339 date, or go-build vcs settings fallback; empty when unknown). Not an error surface: the enrichment kills the store-path chain-verification dance                       | server `versionHandler` + `buildCommitWith`/`buildCommitDateWith` | `version_test.go` + smoke key-checks                                                                                                  |
| Network down (island session POST)                   | Toast: "Could not reach the server…" + `#log` line                                                                                                                                                                                                            | island i18n `sessionNetFailed`                                    | session.test.mjs                                                                                                                      |
| PBX rejects login (401 island REGISTER)              | `loginError` inline + `reg-status` pill "registration rejected"                                                                                                                                                                                               | island i18n `loginError`/`regRejected`                            | i18n parity tests                                                                                                                     |
| SSE feed dead (3 consecutive errors)                 | One warn toast + pill label flips to "not connected"; the later reconnect announces "restored" once — only when the drop itself was announced (M17 J2)                                                                                                        | island i18n `sseDropped`/`ssePillDown`/`sseRestored`              | session.test.mjs                                                                                                                      |
| Reply send transport failure (network / 5xx on htmx) | Failed optimistic bubble (danger tint) with Retry + Dismiss buttons in the bubble; the exact draft is restored into the composer — retry resends it, dismiss keeps it editable (M17 J3/J4)                                                                    | shell.js §3e (English, D3)                                        | shell.test.mjs                                                                                                                        |
| CRM call journal failed (502 / network)              | Warn toast: "Call not recorded in the CRM (…) — you can log it manually." en/de                                                                                                                                                                               | island i18n `crmLogFailed`                                        | crm_test.go 502 contract + i18n parity tests                                                                                          |
| Contact write failed (JSON API 500)                  | Warn toast: "contact save failed (…) ; kept locally" / "Kontakt speichern fehlgeschlagen (…); lokal behalten" — the same 500 text the tab banner shows                                                                                                        | server `contactSaveFailed` (one home)                             | `helpers_contract_test.go` (500 text + routing)                                                                                       |
| Store/render failure on a panel or thread load (500) | Body carries only the op prefix + family default ("load conversation: A temporary error occurred. …") — the raw store text goes to the operator log (`error=` + `family=`), never the browser                                                                 | server `internalError`/`safeDetail` (SafeDetail)                  | `TestInternalErrorRedactsDetail`                                                                                                      |
| Upload body over the 61 MB envelope (400)            | Plain-text 400 "could not read the form: http: request body too large" (English, operator-facing) — htmx surfaces the generic "HTTP 400" toast; legit uploads never see it                                                                                    | server `requireSessionMultipart` (MaxBytesReader)                 | `TestUploadBodyIsBounded` + smoke "oversized upload rejected"                                                                         |
| Unknown path (404)                                   | Styled 404 (shell + error panel), status stays 404                                                                                                                                                                                                            | server `error.notfound` en/de                                     | `TestNotFoundRendersTheShell`                                                                                                         |
| Handler panic                                        | Recovery middleware logs stack + re-raises; user gets htmx/browser failure surface                                                                                                                                                                            | cqrshtmx.RecoveryMiddleware                                       | library + server middleware tests                                                                                                     |
| Passkey login: unknown email / no credential (401)   | Inline `#passkey-login-error`: "No passkey matches this email — check the address or use extension and password below." — begin and finish answer the SAME body, no account oracle; the login flood bucket is shared with the extension login | server uniform 401 + island i18n `passkeyUnknown`                 | `TestPasskeyBeginIsUniformAgainstEnumeration` + passkey.test.mjs                                                                     |
| Passkey ceremony throttled (429)                     | Inline `#passkey-login-error`: "Too many attempts — wait a minute, then try again."                                                                                                                                                                           | island i18n `passkeyThrottled`                                    | passkey.test.mjs                                                                                                                      |
| Passkey ceremony rejected (5xx: unmapped account, password-file missing/empty/unreadable, store) | Inline `passkeyFinishFailed`/`passkeyBeginFailed` "(HTTP N)" — operator-fixable states all answer 503 with the code in the journal (`userauth.*`), never the internal detail | server `passkeyServerError` + island i18n                         | `TestPasskeyFinishMintsTheSessionWithIdentity` class + userauth suite                                                                 |
| Passkey prompt dismissed / timed out                 | Quiet: no error shown (the form is still there), `#log` keeps the class — dismissal is a user choice, not a failure                                                                                                                                           | passkey.js                                                        | passkey.test.mjs                                                                                                                      |
| Enrollment token invalid / expired / already used (503) | Inline `#enroll-error`: ONE message naming the fresh-link remedy — the token state (unknown vs used) is not distinguishable client-side; expired carries its own operator-log code only                                                                      | island i18n `enrollFailed`                                        | enroll.test.mjs + `TestPasskeyEnrollVerifyBurnsTheToken`                                                                             |
| Enrollment prompt cancelled                           | Inline `#enroll-error`: "Passkey creation was cancelled — submit the token again to retry." (the token burned at verify; the copy is honest about it)                                                                                                          | island i18n `enrollDismissed`                                     | enroll.test.mjs                                                                                                                      |

Shell copy (toasts, dedup, throttle wording) stays ENGLISH by decision
D3 (2026-09-20): matches the `#log` operator-channel precedent; the
island's user-facing copy is fully en/de. Localizing shell copy only
if a tabs-style per-extension UX demand shows up.

## Boot surface (operator as user, 2026-10-02)

Ruling: the OPERATOR is a user of this contract. Every boot failure
renders the five-part error contract (WHAT / REASSURE / WHY / FIX /
ESCAPE) plus the underlying error, version-stamped and phase-tagged,
ENGLISH-ONLY (same D3 precedent as the shell copy: the journal is the
operator trail; no i18n entries). The renderer is pure copy data in
`cmd/webphone/bootreport.go`; `run()` tags designed errors at their
sites, a deferred recover catches everything else, and the panic
MECHANISM at the composition root stays fail-fast (DO-1) — only the
PRESENTATION was added, plus removing `App.Start`'s three gratuitous
panics (`InvokeNamed` + `wrapf`, the method already returned error).

| Boot failure class                             | Operator sees                                                 | Classification                                        | Test home                                                   |
| ---------------------------------------------- | ------------------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------------- |
| Panic escaping `run()`                         | 5-part block + `panic trace` marker + raw stack below, exit 2 | recover arm (`bootPanicMiswire`)                      | `TestReportBootPanicRendersBlockTraceAndExitsTwo`           |
| Config rejected (env / `WEBPHONE_CONFIG` JSON) | class=config, exit 1                                          | tagged at the `config.Load` site                      | `TestClassifyBootError` + smoke                             |
| Data dir inaccessible                          | class=data-dir, exit 1                                        | app.New wrap prefix `create data dir:` (drift-pinned) | `TestAppWrapPrefixesStillExistInAppSource` + smoke scenario |
| IANA zone unloadable                           | class=timezone, exit 1                                        | app.New wrap prefix `load timezone:`                  | same                                                        |
| Paperless pair unusable                        | class=paperless, exit 1                                       | app.New wrap prefix `paperless:`                      | same                                                        |
| Bind/serve failure (the most common real one)  | class=listen, exit 1                                          | tagged at the NewServer/serve sites                   | golden pin `TestRenderBootFailureGoldenListen`              |
| Anything else (incl. shutdown-phase errors)    | class=generic, exit 1                                         | fallback                                              | `TestRenderBootFailureCoversEveryClass`                     |

Exit taxonomy (deliberate, previously de-facto): **1** designed boot
failure, **2** panic (Go's default code, now contract). The NixOS
module ships `Restart=on-failure` + `RestartSec=5`, so each attempt
renders once per 5s until fixed; `systemctl stop webphone` ends the
loop. Open owner call (D3 follow-up): cap the retry with
`StartLimitBurst`/`StartLimitIntervalSec`, or keep the 5s retry as
desirable liveness.

Audit appendix (the gap this closes, graded 2026-10-02): before this
train a miswiring panic scored 1/5 contract parts
(`panic: do: service not found: …`, exit 2) and designed boot errors
2/5 (`slog.Error("webphone exited", …)`, main.go). Both render 5/5
now; the live proof is the smoke suite's `boot failure scenario`.

## Send-failure UX layering (2026-09-22)

(plan `docs/planning/archived/2026-09-22_16-07_SUPERB-send-failure-ux.md`):
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

## Error families are total (2026-09-30 train)

Every error this binary constructs or wraps carries two stable names: a
FAMILY — Rejection (caller input, not retryable), Conflict, Transient
(retryable), Corruption (damaged truth), Infrastructure (environment),
Orchestration (our own wiring) — and a `<seam>.<op>` CODE (stable,
non-empty, the grep target). New error paths use `errorfamily.New*` /
`Wrap*` constructors; bare `fmt.Errorf` is banned by the AGENTS erraudit
bar (tier 2 = `--enforce-go-error-family` + `--enforce-coded-errors`,
both held at 0 since 2026-09-30), and equally banned is a family-FIXED
wrap over a polymorphic cause: propagation stays family-neutral (P2),
marked by reasoned nolints where deliberate. Sentinels stay sentinels
and take their family from registration (P3); classification never
changes rendered strings (P4); codes are stable (P5). The seven
principles and the per-seam table live in the error-excellence plan
appendix (`docs/planning/2026-09-22_01-20_SUPERB-error-excellence.md`).

`Error()` renders `[family:code] message` and the handler logs add
`error=`/`family=` fields, so these names are an operator- and
cross-repo-visible vocabulary (the stack runbook § "Webphone error
contract" syncs it). A rename would silently break journal greps, so
the registry below is PINNED: `TestErrorCodeRegistryIsFresh`
(internal/arch) generates it from the source and fails the suite on any
code missing from, stale in, or family-drifted against this page. Two codes split their family at runtime — `pbx.http` and `crm.http`
classify the HTTP answer 4xx → Rejection / 5xx → Transient, the same
split the gateway seam pins; the table marks them `runtime-split` and
each seam's `family_test.go` pins both arms. After
a deliberate code change, regenerate:

    go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update

<!-- error-code-registry: BEGIN (generated block; do not edit by hand; go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update rewrites it) -->
| Code | Families | First site |
| ---- | -------- | ---------- |
| blob.escape | Rejection | internal/blob/store.go |
| blob.name | Infrastructure | internal/blob/store.go |
| blob.open | Infrastructure | internal/blob/store.go |
| blob.probe | Infrastructure | internal/blob/store.go |
| blob.remove | Infrastructure | internal/blob/store.go |
| blob.root | Infrastructure | internal/blob/store.go |
| blob.subdir | Infrastructure | internal/blob/store.go |
| blob.write | Infrastructure | internal/blob/store.go |
| config.addr | Rejection | internal/config/config.go |
| config.auth.passkey.extension_password_files | Rejection | internal/config/config.go |
| config.auth.passkey.rp_id | Rejection | internal/config/config.go |
| config.auth.passkey.rp_origins | Rejection | internal/config/config.go |
| config.auth.passkey.users | Rejection | internal/config/config.go |
| config.contacts | Rejection | internal/config/config.go |
| config.crm.token | Rejection | internal/config/config.go |
| config.crm.url | Rejection | internal/config/config.go |
| config.csrf.trusted_origins | Rejection | internal/config/config.go |
| config.csrf.trusted_proxies | Rejection | internal/config/config.go |
| config.data_dir | Rejection | internal/config/config.go |
| config.defaults | Orchestration | internal/config/config.go |
| config.env | Rejection | internal/config/config.go |
| config.file | Rejection | internal/config/config.go |
| config.gateway.mode | Rejection | internal/config/config.go |
| config.gateway.webhook_url | Rejection | internal/config/config.go |
| config.gateway_secret_sources | Rejection | internal/config/config.go |
| config.identities | Rejection | internal/config/config.go |
| config.paperless.token | Rejection | internal/config/config.go |
| config.paperless.url | Rejection | internal/config/config.go |
| config.session_max_ttl | Rejection | internal/config/config.go |
| config.session_ttl | Rejection | internal/config/config.go |
| config.timezone | Rejection | internal/config/config.go |
| config.turn_rest.dead_secret | Rejection | internal/config/config.go |
| config.turn_rest.ttl | Rejection | internal/config/config.go |
| config.unmarshal | Rejection | internal/config/config.go |
| config.webhook_secret_file | Rejection | internal/config/config.go |
| crm.call_log.failed | Transient | internal/crm/client.go |
| crm.call_log.rejected | Rejection | internal/crm/client.go |
| crm.decode | Transient | internal/crm/client.go |
| crm.encode | Infrastructure | internal/crm/client.go |
| crm.http | runtime-split | internal/crm/client.go |
| crm.read | Transient | internal/crm/client.go |
| crm.request | Infrastructure | internal/crm/client.go |
| crm.transport | Transient | internal/crm/client.go |
| crm.url | Rejection | internal/crm/client.go |
| domain.extension | Rejection | internal/domain/ids.go |
| domain.id | Corruption | internal/domain/ids.go |
| domain.phone | Rejection | internal/domain/ids.go |
| gateway.form | Infrastructure | internal/gateway/webhook.go |
| gateway.receipt | Transient | internal/gateway/webhook.go |
| gateway.request | Infrastructure | internal/gateway/webhook.go |
| gateway.transport | Transient | internal/gateway/webhook.go |
| messaging.verdict | Rejection | internal/messaging/service.go |
| pbx.decode | Transient | internal/pbx/client.go |
| pbx.encode | Infrastructure | internal/pbx/client.go |
| pbx.http | runtime-split | internal/pbx/client.go |
| pbx.request | Infrastructure | internal/pbx/client.go |
| pbx.transport | Transient | internal/pbx/client.go |
| pbx.url | Rejection | internal/pbx/client.go |
| session.insert | Infrastructure | internal/session/sqlite.go |
| session.migrate | Infrastructure | internal/session/sqlite.go |
| session.nil_db | Infrastructure | internal/session/sqlite.go |
| session.sweep | Infrastructure | internal/session/sqlite.go |
| session.token | Infrastructure | internal/session/service.go |
| session.ttl | Infrastructure | internal/session/sqlite.go |
| store.attachment_get | Infrastructure | internal/store/messages.go |
| store.attachment_insert | Infrastructure | internal/store/messages.go |
| store.attachment_save | Infrastructure | internal/messaging/service.go |
| store.attachment_scan | Infrastructure | internal/store/messages.go |
| store.attachments_list | Infrastructure | internal/store/messages.go |
| store.attachments_rows | Infrastructure | internal/store/messages.go |
| store.close | Infrastructure | internal/store/db.go |
| store.contact_delete | Infrastructure | internal/store/contacts.go |
| store.contact_list_full | Rejection | internal/store/contacts.go |
| store.contact_missing | Rejection | internal/store/contacts.go |
| store.contact_save | Infrastructure | internal/store/contacts.go |
| store.contact_scan | Infrastructure | internal/store/contacts.go |
| store.count_archived | Infrastructure | internal/store/messages.go |
| store.counts | Infrastructure | internal/store/sweep.go |
| store.exec | Infrastructure | internal/store/sweep.go |
| store.fax_create | Infrastructure | internal/fax/service.go |
| store.fax_insert | Infrastructure | internal/store/faxes.go |
| store.fax_scan | Infrastructure | internal/store/faxes.go |
| store.fax_spool | Infrastructure | internal/fax/service.go |
| store.fax_update | Infrastructure | internal/store/faxes.go |
| store.message_append | Infrastructure | internal/messaging/service.go |
| store.message_insert | Infrastructure | internal/store/messages.go |
| store.message_scan | Infrastructure | internal/store/messages.go |
| store.migrate | Infrastructure | internal/store/db.go |
| store.open | Infrastructure | internal/store/db.go |
| store.outbound_status | Infrastructure | internal/messaging/service.go |
| store.ping | Infrastructure | internal/store/db.go |
| store.query | Infrastructure | internal/store/db.go |
| store.snippet_delete | Infrastructure | internal/store/snippets.go |
| store.snippet_save | Infrastructure | internal/store/snippets.go |
| store.snippet_scan | Infrastructure | internal/store/snippets.go |
| store.sweep_scan | Infrastructure | internal/store/sweep.go |
| store.thread_find | Infrastructure | internal/store/messages.go |
| store.thread_flag | Infrastructure, Rejection | internal/store/messages.go |
| store.thread_get | Infrastructure | internal/store/messages.go |
| store.thread_insert | Infrastructure | internal/store/messages.go |
| store.thread_mark_read | Infrastructure | internal/store/messages.go |
| store.thread_reread | Infrastructure | internal/store/messages.go |
| store.thread_resolve | Infrastructure | internal/messaging/service.go |
| store.thread_scan | Infrastructure | internal/store/messages.go |
| store.thread_upsert | Infrastructure | internal/store/messages.go |
| store.tx_begin | Infrastructure | internal/store/messages.go |
| store.tx_commit | Infrastructure | internal/store/messages.go |
| userauth.close.db | Infrastructure | internal/userauth/userauth.go |
| userauth.close.users | Infrastructure | internal/userauth/userauth.go |
| userauth.config | Rejection | internal/userauth/userauth.go |
| userauth.db.open | Infrastructure | internal/userauth/userauth.go |
| userauth.db.optimize | Infrastructure | internal/userauth/userauth.go |
| userauth.enroll.burn | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.entropy | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.expire | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.expired | Rejection | internal/userauth/enroll.go |
| userauth.enroll.invalid | Rejection | internal/userauth/enroll.go |
| userauth.enroll.lookup | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.migrate | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.mint | Infrastructure | internal/userauth/enroll.go |
| userauth.enroll.sweep | Infrastructure | internal/userauth/enroll.go |
| userauth.password_file.empty | Rejection | internal/userauth/userauth.go |
| userauth.password_file.missing | Rejection | internal/userauth/userauth.go |
| userauth.password_file.read | Rejection | internal/userauth/userauth.go |
| userauth.register.exists_unreadable | Infrastructure | internal/userauth/userauth.go |
| userauth.unmapped_email | Rejection | internal/userauth/userauth.go |
| webhook.fax_pages | Rejection | internal/server/webhooks.go |

<!-- error-code-registry: END -->

## Cross-repo sync

The operator-facing semantics of these surfaces — status meanings,
the gateway/bridge string contract (rendered verbatim in webhook
mode), and the `family=` log vocabulary — are cross-documented in the
stack runbook: `nix-international-telephony/docs/ops-runbook.md`
§ "Webphone error contract" (2026-09-22). Keep both sides in sync
when error copy or families change. The boot-surface classes and exit
taxonomy (2026-10-02) join the sync set; the stack-side patch text is
prepared in
`docs/planning/2026-10-02_11-05_boot-contract-stack-runbook-patch.md`
and applies under the tri-repo ritual (webphone first, clean stack
tree, then relock).
