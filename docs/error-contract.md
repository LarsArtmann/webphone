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
| blob.escape | Rejection | blob/store.go |
| blob.name | Infrastructure | blob/store.go |
| blob.open | Infrastructure | blob/store.go |
| blob.probe | Infrastructure | blob/store.go |
| blob.remove | Infrastructure | blob/store.go |
| blob.root | Infrastructure | blob/store.go |
| blob.subdir | Infrastructure | blob/store.go |
| blob.write | Infrastructure | blob/store.go |
| config.addr | Rejection | config/config.go |
| config.contacts | Rejection | config/config.go |
| config.crm.token | Rejection | config/config.go |
| config.crm.url | Rejection | config/config.go |
| config.csrf.trusted_origins | Rejection | config/config.go |
| config.csrf.trusted_proxies | Rejection | config/config.go |
| config.data_dir | Rejection | config/config.go |
| config.defaults | Orchestration | config/config.go |
| config.env | Rejection | config/config.go |
| config.file | Rejection | config/config.go |
| config.gateway.mode | Rejection | config/config.go |
| config.gateway.webhook_url | Rejection | config/config.go |
| config.gateway_secret_sources | Rejection | config/config.go |
| config.identities | Rejection | config/config.go |
| config.paperless.token | Rejection | config/config.go |
| config.paperless.url | Rejection | config/config.go |
| config.session_max_ttl | Rejection | config/config.go |
| config.session_ttl | Rejection | config/config.go |
| config.timezone | Rejection | config/config.go |
| config.turn_rest.dead_secret | Rejection | config/config.go |
| config.turn_rest.ttl | Rejection | config/config.go |
| config.unmarshal | Rejection | config/config.go |
| config.webhook_secret_file | Rejection | config/config.go |
| crm.call_log.failed | Transient | crm/client.go |
| crm.call_log.rejected | Rejection | crm/client.go |
| crm.decode | Transient | crm/client.go |
| crm.encode | Infrastructure | crm/client.go |
| crm.http | runtime-split | crm/client.go |
| crm.read | Transient | crm/client.go |
| crm.request | Infrastructure | crm/client.go |
| crm.transport | Transient | crm/client.go |
| crm.url | Rejection | crm/client.go |
| domain.extension | Rejection | domain/ids.go |
| domain.id | Corruption | domain/ids.go |
| domain.phone | Rejection | domain/ids.go |
| gateway.form | Infrastructure | gateway/webhook.go |
| gateway.receipt | Transient | gateway/webhook.go |
| gateway.request | Infrastructure | gateway/webhook.go |
| gateway.transport | Transient | gateway/webhook.go |
| messaging.verdict | Rejection | messaging/service.go |
| pbx.decode | Transient | pbx/client.go |
| pbx.encode | Infrastructure | pbx/client.go |
| pbx.http | runtime-split | pbx/client.go |
| pbx.request | Infrastructure | pbx/client.go |
| pbx.transport | Transient | pbx/client.go |
| pbx.url | Rejection | pbx/client.go |
| session.insert | Infrastructure | session/sqlite.go |
| session.migrate | Infrastructure | session/sqlite.go |
| session.nil_db | Infrastructure | session/sqlite.go |
| session.sweep | Infrastructure | session/sqlite.go |
| session.token | Infrastructure | session/service.go |
| session.ttl | Infrastructure | session/sqlite.go |
| store.attachment_get | Infrastructure | store/messages.go |
| store.attachment_insert | Infrastructure | store/messages.go |
| store.attachment_save | Infrastructure | messaging/service.go |
| store.attachment_scan | Infrastructure | store/messages.go |
| store.attachments_list | Infrastructure | store/messages.go |
| store.attachments_rows | Infrastructure | store/messages.go |
| store.close | Infrastructure | store/db.go |
| store.contact_delete | Infrastructure | store/contacts.go |
| store.contact_list_full | Rejection | store/contacts.go |
| store.contact_missing | Rejection | store/contacts.go |
| store.contact_save | Infrastructure | store/contacts.go |
| store.contact_scan | Infrastructure | store/contacts.go |
| store.count_archived | Infrastructure | store/messages.go |
| store.counts | Infrastructure | store/sweep.go |
| store.exec | Infrastructure | store/sweep.go |
| store.fax_create | Infrastructure | fax/service.go |
| store.fax_insert | Infrastructure | store/faxes.go |
| store.fax_scan | Infrastructure | store/faxes.go |
| store.fax_spool | Infrastructure | fax/service.go |
| store.fax_update | Infrastructure | store/faxes.go |
| store.message_append | Infrastructure | messaging/service.go |
| store.message_insert | Infrastructure | store/messages.go |
| store.message_scan | Infrastructure | store/messages.go |
| store.migrate | Infrastructure | store/db.go |
| store.open | Infrastructure | store/db.go |
| store.outbound_status | Infrastructure | messaging/service.go |
| store.ping | Infrastructure | store/db.go |
| store.query | Infrastructure | store/db.go |
| store.snippet_delete | Infrastructure | store/snippets.go |
| store.snippet_save | Infrastructure | store/snippets.go |
| store.snippet_scan | Infrastructure | store/snippets.go |
| store.sweep_scan | Infrastructure | store/sweep.go |
| store.thread_find | Infrastructure | store/messages.go |
| store.thread_flag | Infrastructure, Rejection | store/messages.go |
| store.thread_get | Infrastructure | store/messages.go |
| store.thread_insert | Infrastructure | store/messages.go |
| store.thread_mark_read | Infrastructure | store/messages.go |
| store.thread_reread | Infrastructure | store/messages.go |
| store.thread_resolve | Infrastructure | messaging/service.go |
| store.thread_scan | Infrastructure | store/messages.go |
| store.thread_upsert | Infrastructure | store/messages.go |
| store.tx_begin | Infrastructure | store/messages.go |
| store.tx_commit | Infrastructure | store/messages.go |
| webhook.fax_pages | Rejection | server/webhooks.go |

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
