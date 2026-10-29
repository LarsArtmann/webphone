# Error contract — the failure → feedback map

What the user SEES when X fails (2026-09-20 train, plan T13). Every
error path lands in at least one VISIBLE surface (toast, inline
banner, or panel); `#log` is always the operator trail, never the
only user feedback.

| Failure                                                                                          | User sees                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Owner (wording)                                                                                     | Test home                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Tab session dead (401 on tab actions)                                                            | Throttled error toast: "Tab session ended; calls keep working…" — never auto-reload                                                                                                                                                                                                                                                                                                                                                                                                           | shell.js §3c (English, D3)                                                                          | shell.test.mjs + `TestShellJSSurfacesHtmxErrors`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Validation mistake (422)                                                                         | `.wp-error` banner swapped into `#wp-tab-error` + error toast                                                                                                                                                                                                                                                                                                                                                                                                                                 | server (en/de via `h.T`)                                                                            | `TestSendClassifiesGatewayOutageAs502` + renderPanelError tests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Provider refused the send (422, its reason)                                                      | Banner + toast carry the provider's own refusal text; saved as failed, NO retry affordance. Self-sends to the own DID (config identities) ride the same arm with a locally synthesized refusal (train C) — same 422, same failed row, zero provider roundtrip                                                                                                                                                                                                                                 | server i18n `err.messageRejected`/`err.faxRejected`                                                 | messages_test refusal arm + `TestFaxProviderRefusalAnswers422` + `TestThreadViewWarnsOnSelfSend`/`TestFaxToOwnNumberIsRefusedLocally`                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Rate limited (429)                                                                               | Toast: "Too many requests — wait a moment…"                                                                                                                                                                                                                                                                                                                                                                                                                                                   | shell.js (htmx) / island i18n (login)                                                               | shell.test.mjs + session.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| Gateway outage (502)                                                                             | Banner + toast; message/fax saved as failed; retry affordance offered (also the arm for provider 5xx ANSWERS since train E)                                                                                                                                                                                                                                                                                                                                                                   | server                                                                                              | 502 test (HX-Trigger + banner pinned)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Other 4xx/5xx on htmx actions                                                                    | Banner (`.wp-error` selected) + generic toast "HTTP N"                                                                                                                                                                                                                                                                                                                                                                                                                                        | shell.js generic copy                                                                               | contract test (config markers)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Network down (htmx)                                                                              | Toast: "Network request failed…"                                                                                                                                                                                                                                                                                                                                                                                                                                                              | shell.js                                                                                            | shell.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Version probe (`GET /version`)                                                                   | JSON: `version` (+ `commit`, `commitDate` when known — ldflags rev/dirtyRev + RFC 3339 date, or go-build vcs settings fallback; empty when unknown). Not an error surface: the enrichment kills the store-path chain-verification dance                                                                                                                                                                                                                                                       | server `versionHandler` + `buildCommitWith`/`buildCommitDateWith`                                   | `version_test.go` + smoke key-checks                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Network down (island session POST)                                                               | Toast: "Could not reach the server…" + `#log` line                                                                                                                                                                                                                                                                                                                                                                                                                                            | island i18n `sessionNetFailed`                                                                      | session.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| PBX rejects login (401 island REGISTER)                                                          | `loginError` inline + `reg-status` pill "registration rejected"                                                                                                                                                                                                                                                                                                                                                                                                                               | island i18n `loginError`/`regRejected`                                                              | i18n parity tests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| SSE feed dead (3 consecutive errors)                                                             | One warn toast + pill label flips to "not connected"; the later reconnect announces "restored" once — only when the drop itself was announced (M17 J2)                                                                                                                                                                                                                                                                                                                                        | island i18n `sseDropped`/`ssePillDown`/`sseRestored`                                                | session.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Reply send transport failure (network / 5xx on htmx)                                             | Failed optimistic bubble (danger tint) with Retry + Dismiss buttons in the bubble; the exact draft is restored into the composer — retry resends it, dismiss keeps it editable (M17 J3/J4)                                                                                                                                                                                                                                                                                                    | shell.js §3e (English, D3)                                                                          | shell.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| CRM call journal failed (502 / network)                                                          | Warn toast: "Call not recorded in the CRM (…) — you can log it manually." en/de                                                                                                                                                                                                                                                                                                                                                                                                               | island i18n `crmLogFailed`                                                                          | crm_test.go 502 contract + i18n parity tests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Contact write failed (JSON API 500)                                                              | Warn toast: "contact save failed (…) ; kept locally" / "Kontakt speichern fehlgeschlagen (…); lokal behalten" — the same 500 text the tab banner shows                                                                                                                                                                                                                                                                                                                                        | server `contactSaveFailed` (one home)                                                               | `helpers_contract_test.go` (500 text + routing)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Store/render failure on a panel or thread load (500)                                             | Body carries only the op prefix + family default ("load conversation: A temporary error occurred. …") — the raw store text goes to the operator log (`error=` + `family=`), never the browser                                                                                                                                                                                                                                                                                                 | server `internalError`/`safeDetail` (SafeDetail)                                                    | `TestInternalErrorRedactsDetail`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Upload body over the 61 MB envelope (400)                                                        | Plain-text 400 "could not read the form: http: request body too large" (English, operator-facing) — htmx surfaces the generic "HTTP 400" toast; legit uploads never see it                                                                                                                                                                                                                                                                                                                    | server `requireSessionMultipart` (MaxBytesReader)                                                   | `TestUploadBodyIsBounded` + smoke "oversized upload rejected"                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Unknown path (404)                                                                               | Styled 404: shell + `views.NotFoundPanel` (literal 404, the requested path echoed back, link home, no reload button), `Cache-Control: no-store`, status stays 404. Covers `/health` + `/health/*` while the dashboard is off. Shape follows the request — see § "404 — one home, two shapes" below                                                                                                                                                                                            | server `error.notfound`, `error.notfound.title`, `error.notfound.back` (en/de)                      | `TestNotFoundPanelRendersTheContract`, `TestNotFoundPanelEscapesThePath`, `TestUnknownPathsAnswerTheStyledNotFoundPage`, `TestNotFoundPageOffersNoReloadButton`, `TestNotFoundPageIsNotCacheable`, `TestNotFoundAnswersHeadWithoutBody`, `TestNotFoundEchoesThePathWithoutTheQuery`, `TestNotFoundPageFollowsTheLanguage`, `TestMissingAssetAnswersStyledForNavigationsOnly`, `TestHTMXMissAnswersANoticeNotTheShell`, `TestDocumentMissOnAPartialURLAnswersTheShell`, `TestActionMissAnswersByRequestShape`, `TestStaleFaxDocumentLinkAnswersTheStyledPage`, `TestAttachmentMissKeepsTheTerseAnswer` |
| Method mismatch on a known path (405)                                                            | Bare stdlib 405, deliberately: every unsafe method crosses the CSRF gate first (a tokenless POST is 403, not 405), so reaching it takes a signed-in, token-bearing, hand-crafted request — machine traffic, not a user surface. Wrapping the mux to restyle it would mean buffering every protected response (templ streams; `/api/export` is a zip) for a case no UI can produce                                                                                                             | net/http ServeMux                                                                                   | none (decision, not a gap); revisit only if the island ever issues a method-mismatched request                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Handler panic                                                                                    | Recovery middleware logs stack + re-raises; user gets htmx/browser failure surface                                                                                                                                                                                                                                                                                                                                                                                                            | cqrshtmx.RecoveryMiddleware                                                                         | library + server middleware tests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| Passkey login: unknown email / no credential (401)                                               | Inline `#passkey-login-error`: "No passkey matches this email — check the address or use extension and password below." — begin and finish answer the SAME body, no account oracle; the login flood bucket is shared with the extension login                                                                                                                                                                                                                                                 | server uniform 401 + island i18n `passkeyUnknown`                                                   | `TestPasskeyBeginIsUniformAgainstEnumeration` + passkey.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Passkey ceremony throttled (429)                                                                 | Inline `#passkey-login-error`: "Too many attempts — wait a minute, then try again."                                                                                                                                                                                                                                                                                                                                                                                                           | island i18n `passkeyThrottled`                                                                      | passkey.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Passkey ceremony key malformed (400)                                                             | Plain-text 400: "user_id must be the session_key the begin step returned" — a garbage ceremony key is caller error, strictly ULID-parsed at the door (never silently hashed into a pseudo-user that would answer the unknown-account 401)                                                                                                                                                                                                                                                     | server `parsePasskeyUserID`                                                                         | `TestPasskeyFinishRejectsAMalformedUserID` + `TestPasskeyEnrollHandlersRejectMissingUserID`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| Passkey ceremony rejected (5xx: unmapped account, password-file missing/empty/unreadable, store) | Inline `passkeyFinishFailed`/`passkeyBeginFailed` "(HTTP N)" — operator-fixable states all answer 503 with the code in the journal (`userauth.*`), never the internal detail                                                                                                                                                                                                                                                                                                                  | server `passkeyServerError` + island i18n                                                           | `TestPasskeyFinishMintsTheSessionWithIdentity` class + userauth suite                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Passkey prompt dismissed / timed out                                                             | Quiet: no error shown (the form is still there), `#log` keeps the class — dismissal is a user choice, not a failure                                                                                                                                                                                                                                                                                                                                                                           | passkey.js                                                                                          | passkey.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Enrollment token invalid / expired / already used (503)                                          | Inline `#enroll-error`: ONE message naming the fresh-link remedy — the token state (unknown vs used) is not distinguishable client-side; expired carries its own operator-log code only                                                                                                                                                                                                                                                                                                       | island i18n `enrollFailed`                                                                          | enroll.test.mjs + `TestPasskeyEnrollVerifyBurnsTheToken`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Enrollment prompt cancelled                                                                      | Inline `#enroll-error`: "Passkey creation was cancelled — submit the token again to retry." (the token burned at verify; the copy is honest about it)                                                                                                                                                                                                                                                                                                                                         | island i18n `enrollDismissed`                                                                       | enroll.test.mjs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Passkey revoke failed (store error on `DELETE /api/passkeys/{id}`)                               | Error toast `passkeys.revokeFailed` ("Revoking the passkey failed — try again.", en/de) + the whole Settings partial re-renders with STORE truth — the list never shows a passkey the store still holds; detail goes to the journal (`safeDetail`), never the browser                                                                                                                                                                                                                         | server `revokePasskey` + island toast i18n                                                          | `TestPasskeySettingsListsAndRevokes` + i18n parity tests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Passkey revoke on an extension with no mapped identity account                                   | Plain-text 404 "no passkey account" (machine-facing state, unreachable through the UI — the section only renders when a mapping exists)                                                                                                                                                                                                                                                                                                                                                       | server `revokePasskey`                                                                              | `TestPasskeySettingsListsAndRevokes` (unmapped arm)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Internal directory seam off (`GET /internal/directory/{extension}`)                              | Honest 404 — a disabled feature answers like a missing one (ASR-seam convention). Machine-facing (the stack's FS xml_curl bridge), never the browser: loopback + Bearer secret only (403 otherwise), invalid extension 400, unknown/session-less extension answers an EMPTY list so the bridge's static fallback keeps the lockout decision. Stack side — the bridge, fail-closed posture, <2-min rollback drill: nix-international-telephony `docs/ops-runbook.md` § "Dynamic SIP directory" | server `internalDirectory` (bare statuses like every gate — no `sipcred.*` codes, 2026-10-12 train) | `TestInternalDirectoryFailsClosedWithoutASecret` + `TestInternalDirectoryAnswersTheDerivedCredentials` + `TestInternalDirectoryRejectsBadSecrets` + `TestInternalDirectoryUnknownExtensionAnswersEmpty` + `TestInternalDirectoryInvalidExtensionIs400` + `TestInternalDirectoryLoopbackOnly`                                                                                                                                                                                                                                                                                                          |

## `/api` error bodies — plain text, RULED (2026-10-23)

The island JSON endpoints answer failures as plain-text copy with the
honest reason and the status that carries the meaning (`http.Error`, so
`text/plain; charset=utf-8`). Success bodies are JSON; failures are not.
That asymmetry is deliberate, not an oversight, and it stays:

- **No consumer parses the body.** `authedFetch`, `shell.js` and the
  passkey/enroll modules branch on the STATUS and show their own toast or
  inline message; `passkey.js` renders "(HTTP N)". A JSON envelope would
  be decoded by nobody.
- **The copy is already the user-facing explanation** — "contact list is
  full — delete one first", "transcription is busy — retry shortly".
  Wrapping it in `{"error": …}` would only add a field to unwrap.
- **A partial migration splits the contract.** Forty-odd failure sites
  across `internal/server/api`, `passkey_api.go` and `transcribe.go`
  would ship both shapes for as long as the sweep takes, and every client
  would need a "maybe JSON" branch — strictly worse than one shape.
- **`errorpages`-style JSON is for machine consumers that do not exist
  here.** The only machine consumers of this deployment are the provider
  webhooks and the stack's SIP directory bridge; both have their own
  byte-stable contracts (see the webhook rows above) and neither touches
  `/api`.

**Reopen trigger:** a non-island consumer of `/api` appears (a CLI, a
mobile client, a second in-repo service) that must branch on an error
CODE rather than a status. Then the contract becomes
`{"code": "<seam>.<op>", "message": "…"}` for every `/api` failure in one
change, versioned at the endpoint, and the plain-text shape is retired —
never run side by side.

## 404 — one home, two shapes

`h.notFound` (`internal/server/pages.go`) is the one writer for the HTML
router's misses (the asset mux keeps its own `Accept` split and the
unknown-path/health-off arms call `notFoundPage` directly, since those are
always navigations). The shape follows what the caller can consume:

- **Document request** (browser navigation: a stale link, a partial URL
  pasted out of devtools) -> the styled page (`notFoundPage`).
- **HTMX request** (`HX-Request: true`) -> a one-line `.wp-error` notice,
  status still 404. htmx absorbs non-2xx bodies and reports the status
  through its own responseError toast, so shipping the shell there would
  be bytes nobody renders.

Machine surfaces deliberately stay out of it: JSON APIs
(`internal/server/api`, `passkey_api.go`, `sso_api.go`), provider webhooks
(`/hooks/*`) and fetched byte streams (`/attachments`, `/assets/*`) keep
their terse stdlib answers, because a browser is not their client and an
HTML body would be meaningless. A fax document link
(`/fax/{id}/document`) DOES route through it: that link is a manual
browser download, so a stale one deserves the readable page.

The 405 stance is the neighbouring row: a method mismatch is unreachable
without a valid CSRF token, so restyling it would mean buffering every
protected response for a case no UI can produce.
| Flight-recorder seam off (`GET /internal/flightrecorder`) | Styled 404 (a disabled feature answers like a missing path). Machine-facing and loopback-only when on: a browser can never reach it, a non-loopback peer gets 403, and success answers the snapshot's filesystem path as `text/plain`. Enabled with `WEBPHONE_DEBUG_FLIGHTRECORDER=1` | server `traceSnapshot` + `app.traceRecorder` | none (decision, not a gap) + `TestTraceSnapshotIsAbsentWhenTheSeamIsOff`, `TestTraceSnapshotReportsTheWrittenPath` |

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
code missing from, stale in, or family-drifted against this page. The pin
compares cell CONTENT with padding collapsed, so markdown formatters
re-aligning the table's columns (the auto-commit daemon does) cannot break
the suite. Two codes split their family at runtime — `pbx.http` and `crm.http`
classify the HTTP answer 4xx → Rejection / 5xx → Transient, the same
split the gateway seam pins; the table marks them `runtime-split` and
each seam's `family_test.go` pins both arms. After
a deliberate code change, regenerate:

    go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update

<!-- error-code-registry: BEGIN (generated block; do not edit by hand; go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update rewrites it) -->

| Code                                         | Families                  | First site                         |
| -------------------------------------------- | ------------------------- | ---------------------------------- |
| agent.summarize                              | Transient                 | internal/server/api/review_api.go  |
| api.note_save                                | Infrastructure            | internal/server/api/note_api.go    |
| asr.decode                                   | Transient                 | internal/asr/asr.go                |
| asr.empty                                    | Rejection                 | internal/asr/asr.go                |
| asr.encode                                   | Infrastructure            | internal/asr/asr.go                |
| asr.http                                     | Transient, runtime-split  | internal/asr/asr.go                |
| asr.read                                     | Transient                 | internal/asr/asr.go                |
| asr.request                                  | Infrastructure            | internal/asr/asr.go                |
| asr.transport                                | Transient                 | internal/asr/asr.go                |
| asr.url                                      | Rejection                 | internal/asr/asr.go                |
| authmap.password_file.empty                  | Rejection                 | internal/authmap/authmap.go        |
| authmap.password_file.missing                | Rejection                 | internal/authmap/authmap.go        |
| authmap.password_file.read                   | Rejection                 | internal/authmap/authmap.go        |
| blob.name                                    | Infrastructure            | internal/blob/store.go             |
| blob.open                                    | Infrastructure            | internal/blob/store.go             |
| blob.probe                                   | Infrastructure            | internal/blob/store.go             |
| blob.remove                                  | Infrastructure            | internal/blob/store.go             |
| blob.root                                    | Infrastructure            | internal/blob/store.go             |
| blob.subdir                                  | Infrastructure            | internal/blob/store.go             |
| blob.write                                   | Infrastructure            | internal/blob/store.go             |
| config.addr                                  | Rejection                 | internal/config/config.go          |
| config.asr.project                           | Rejection                 | internal/config/config.go          |
| config.asr.provider                          | Rejection                 | internal/config/config.go          |
| config.asr.url                               | Rejection                 | internal/config/config.go          |
| config.auth.extension_password_files         | Rejection                 | internal/config/config.go          |
| config.auth.oauth2.client_id                 | Rejection                 | internal/config/config.go          |
| config.auth.oauth2.client_secret_file        | Rejection                 | internal/config/config.go          |
| config.auth.oauth2.issuer                    | Rejection                 | internal/config/config.go          |
| config.auth.oauth2.redirect_url              | Rejection                 | internal/config/config.go          |
| config.auth.oauth2.scopes                    | Rejection                 | internal/config/config.go          |
| config.auth.passkey.extension_password_files | Rejection                 | internal/config/config.go          |
| config.auth.passkey.rp_id                    | Rejection                 | internal/config/config.go          |
| config.auth.passkey.rp_origins               | Rejection                 | internal/config/config.go          |
| config.auth.passkey.users                    | Rejection                 | internal/config/config.go          |
| config.auth.users                            | Rejection                 | internal/config/config.go          |
| config.contacts                              | Rejection                 | internal/config/config.go          |
| config.crm.contact_url                       | Rejection                 | internal/config/config.go          |
| config.crm.token                             | Rejection                 | internal/config/config.go          |
| config.crm.token_file                        | Rejection                 | internal/config/config.go          |
| config.crm.token_sources                     | Rejection                 | internal/config/config.go          |
| config.crm.url                               | Rejection                 | internal/config/config.go          |
| config.csrf.trusted_origins                  | Rejection                 | internal/config/config.go          |
| config.csrf.trusted_proxies                  | Rejection                 | internal/config/config.go          |
| config.data_dir                              | Rejection                 | internal/config/config.go          |
| config.defaults                              | Orchestration             | internal/config/config.go          |
| config.dial.cooldown                         | Rejection                 | internal/config/config.go          |
| config.env                                   | Rejection                 | internal/config/config.go          |
| config.file                                  | Rejection                 | internal/config/config.go          |
| config.gateway.mode                          | Rejection                 | internal/config/config.go          |
| config.gateway.webhook_url                   | Rejection                 | internal/config/config.go          |
| config.gateway_secret_sources                | Rejection                 | internal/config/config.go          |
| config.identities                            | Rejection                 | internal/config/config.go          |
| config.matrix.access_token                   | Rejection                 | internal/config/config.go          |
| config.matrix.homeserver                     | Rejection                 | internal/config/config.go          |
| config.matrix.owner                          | Rejection                 | internal/config/config.go          |
| config.matrix.user_id                        | Rejection                 | internal/config/config.go          |
| config.paperless.token                       | Rejection                 | internal/config/config.go          |
| config.paperless.url                         | Rejection                 | internal/config/config.go          |
| config.push.fcm                              | Rejection                 | internal/push/fcm.go               |
| config.push.ntfy.url                         | Rejection                 | internal/config/config.go          |
| config.retention_days                        | Rejection                 | internal/config/config.go          |
| config.session_max_ttl                       | Rejection                 | internal/config/config.go          |
| config.session_ttl                           | Rejection                 | internal/config/config.go          |
| config.signal.account                        | Rejection                 | internal/config/config.go          |
| config.signal.attachments_dir                | Rejection                 | internal/config/config.go          |
| config.signal.owner                          | Rejection                 | internal/config/config.go          |
| config.signal.socket_path                    | Rejection                 | internal/config/config.go          |
| config.sip.credential_secret_file            | Rejection                 | internal/config/config.go          |
| config.sip.credential_secret_sources         | Rejection                 | internal/config/config.go          |
| config.textagent.secret_file                 | Rejection                 | internal/config/config.go          |
| config.textagent.url                         | Rejection                 | internal/config/config.go          |
| config.timezone                              | Rejection                 | internal/config/config.go          |
| config.turn_rest.dead_secret                 | Rejection                 | internal/config/config.go          |
| config.turn_rest.ttl                         | Rejection                 | internal/config/config.go          |
| config.unmarshal                             | Rejection                 | internal/config/config.go          |
| config.webhook_secret_file                   | Rejection                 | internal/config/config.go          |
| crm.call_log.failed                          | Transient                 | internal/crm/client.go             |
| crm.call_log.rejected                        | Rejection                 | internal/crm/client.go             |
| crm.context                                  | runtime-split             | internal/crm/context.go            |
| crm.context_decode                           | Transient                 | internal/crm/context.go            |
| crm.context_read                             | Transient                 | internal/crm/context.go            |
| crm.decode                                   | Transient                 | internal/crm/client.go             |
| crm.encode                                   | Infrastructure            | internal/crm/client.go             |
| crm.health.failed                            | Transient                 | internal/crm/client.go             |
| crm.health.unauthorized                      | Rejection                 | internal/crm/client.go             |
| crm.http                                     | runtime-split             | internal/crm/client.go             |
| crm.read                                     | Transient                 | internal/crm/client.go             |
| crm.request                                  | Infrastructure            | internal/crm/client.go             |
| crm.transport                                | Transient                 | internal/crm/client.go             |
| crm.url                                      | Rejection                 | internal/crm/client.go             |
| domain.address                               | Rejection                 | internal/domain/network.go         |
| domain.channel                               | Rejection                 | internal/domain/message.go         |
| domain.direction                             | Rejection                 | internal/domain/message.go         |
| domain.disposition                           | Rejection                 | internal/domain/disposition.go     |
| domain.extension                             | Rejection                 | internal/domain/ids.go             |
| domain.fax_direction                         | Rejection                 | internal/domain/fax.go             |
| domain.fax_status                            | Rejection                 | internal/domain/fax.go             |
| domain.id                                    | Corruption                | internal/domain/ids.go             |
| domain.network                               | Rejection                 | internal/domain/network.go         |
| domain.outbound_status                       | Rejection                 | internal/domain/message.go         |
| domain.phone                                 | Rejection                 | internal/domain/ids.go             |
| domain.relpath                               | Rejection                 | internal/domain/relpath.go         |
| gateway.form                                 | Infrastructure            | internal/gateway/webhook.go        |
| gateway.receipt                              | Transient                 | internal/gateway/webhook.go        |
| gateway.request                              | Infrastructure            | internal/gateway/webhook.go        |
| gateway.transport                            | Transient                 | internal/gateway/webhook.go        |
| matrix.attachment_read                       | Infrastructure            | internal/matrix/client.go          |
| matrix.create_room                           | Infrastructure            | internal/matrix/client.go          |
| matrix.decode                                | Infrastructure, Transient | internal/matrix/client.go          |
| matrix.download                              | Rejection, Transient      | internal/matrix/client.go          |
| matrix.empty                                 | Rejection                 | internal/matrix/client.go          |
| matrix.encode                                | Rejection                 | internal/matrix/client.go          |
| matrix.mxc                                   | Rejection                 | internal/matrix/client.go          |
| matrix.network                               | Rejection                 | internal/matrix/client.go          |
| matrix.request                               | Infrastructure, Transient | internal/matrix/client.go          |
| matrix.send                                  | Infrastructure            | internal/matrix/client.go          |
| matrix.token                                 | Rejection                 | internal/matrix/client.go          |
| matrix.upload                                | Infrastructure, Transient | internal/matrix/client.go          |
| messaging.sender_register                    | Rejection                 | internal/messaging/service.go      |
| messaging.verdict                            | Rejection                 | internal/messaging/service.go      |
| pbx.decode                                   | Transient                 | internal/pbx/client.go             |
| pbx.encode                                   | Infrastructure            | internal/pbx/client.go             |
| pbx.http                                     | runtime-split             | internal/pbx/client.go             |
| pbx.request                                  | Infrastructure            | internal/pbx/client.go             |
| pbx.transport                                | Transient                 | internal/pbx/client.go             |
| pbx.url                                      | Rejection                 | internal/pbx/client.go             |
| push.draft_devices                           | Infrastructure            | internal/server/api/summary_api.go |
| push.draft_send                              | Infrastructure            | internal/server/api/summary_api.go |
| push.fcm_send                                | Infrastructure, Rejection | internal/push/fcm.go               |
| push.fcm_token                               | Infrastructure, Rejection | internal/push/fcm.go               |
| push.ntfy_send                               | Infrastructure, Rejection | internal/push/ntfy.go              |
| push.transport                               | Rejection                 | internal/push/push.go              |
| readmodels.agent_audit_append                | Infrastructure            | internal/readmodels/agentaudit.go  |
| readmodels.agent_audit_recent                | Infrastructure            | internal/readmodels/agentaudit.go  |
| readmodels.agent_audit_trim                  | Infrastructure            | internal/readmodels/agentaudit.go  |
| readmodels.close                             | Infrastructure            | internal/readmodels/readmodels.go  |
| readmodels.contact_count                     | Infrastructure            | internal/readmodels/contacts.go    |
| readmodels.contact_delete                    | Infrastructure            | internal/readmodels/contacts.go    |
| readmodels.contact_list                      | Infrastructure            | internal/readmodels/contacts.go    |
| readmodels.contact_list_full                 | Rejection                 | internal/readmodels/contacts.go    |
| readmodels.contact_missing                   | Rejection                 | internal/readmodels/contacts.go    |
| readmodels.contact_save                      | Infrastructure            | internal/readmodels/contacts.go    |
| readmodels.dial_attempt                      | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.dial_override_arm                 | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.dial_override_consume             | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.dial_override_load                | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.dial_trim                         | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.dial_window_load                  | Infrastructure            | internal/readmodels/dialwindows.go |
| readmodels.disposition_save                  | Infrastructure            | internal/readmodels/callnotes.go   |
| readmodels.enroll_burn                       | Infrastructure            | internal/readmodels/enroll.go      |
| readmodels.enroll_lookup                     | Infrastructure            | internal/readmodels/enroll.go      |
| readmodels.enroll_mint                       | Infrastructure            | internal/readmodels/enroll.go      |
| readmodels.enroll_sweep                      | Infrastructure            | internal/readmodels/enroll.go      |
| readmodels.fax_by_ref                        | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_count                         | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_get                           | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_insert                        | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_list                          | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_missing                       | Rejection                 | internal/readmodels/faxes.go       |
| readmodels.fax_sweep                         | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.fax_update                        | Infrastructure            | internal/readmodels/faxes.go       |
| readmodels.note_get                          | Infrastructure            | internal/readmodels/callnotes.go   |
| readmodels.note_list_full                    | Rejection                 | internal/readmodels/callnotes.go   |
| readmodels.note_save                         | Infrastructure            | internal/readmodels/callnotes.go   |
| readmodels.open                              | Infrastructure            | internal/readmodels/readmodels.go  |
| readmodels.ping                              | Infrastructure            | internal/readmodels/readmodels.go  |
| readmodels.push_list                         | Infrastructure            | internal/readmodels/push.go        |
| readmodels.push_list_full                    | Rejection                 | internal/readmodels/push.go        |
| readmodels.push_remove                       | Infrastructure            | internal/readmodels/push.go        |
| readmodels.push_upsert                       | Infrastructure            | internal/readmodels/push.go        |
| readmodels.review_get                        | Infrastructure            | internal/readmodels/callreviews.go |
| readmodels.review_list_full                  | Rejection                 | internal/readmodels/callreviews.go |
| readmodels.review_save                       | Infrastructure            | internal/readmodels/callreviews.go |
| readmodels.snippet_delete                    | Infrastructure            | internal/readmodels/snippets.go    |
| readmodels.snippet_list                      | Infrastructure            | internal/readmodels/snippets.go    |
| readmodels.snippet_list_full                 | Rejection                 | internal/readmodels/snippets.go    |
| readmodels.snippet_missing                   | Rejection                 | internal/readmodels/snippets.go    |
| readmodels.snippet_save                      | Infrastructure            | internal/readmodels/snippets.go    |
| readmodels.transcript_append                 | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_call                   | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_count                  | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_delete                 | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_scan                   | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_sweep                  | Infrastructure            | internal/readmodels/transcripts.go |
| readmodels.transcript_trim                   | Infrastructure            | internal/readmodels/transcripts.go |
| session.insert                               | Infrastructure            | internal/session/sqlite.go         |
| session.migrate                              | Infrastructure            | internal/session/sqlite.go         |
| session.migrate_login_email                  | Infrastructure            | internal/session/sqlite.go         |
| session.nil_db                               | Infrastructure            | internal/session/sqlite.go         |
| session.sweep                                | Infrastructure            | internal/session/sqlite.go         |
| session.token                                | Infrastructure            | internal/session/service.go        |
| session.ttl                                  | Infrastructure            | internal/session/sqlite.go         |
| signal.attachment_read                       | Infrastructure            | internal/signal/client.go          |
| signal.connect                               | Transient                 | internal/signal/client.go          |
| signal.decode                                | Infrastructure            | internal/signal/client.go          |
| signal.encode                                | Rejection                 | internal/signal/client.go          |
| signal.network                               | Rejection                 | internal/signal/client.go          |
| signal.rpc                                   | Rejection                 | internal/signal/client.go          |
| signal.send                                  | Infrastructure            | internal/signal/client.go          |
| signal.write                                 | Infrastructure, Transient | internal/signal/client.go          |
| sso.attempt.missing                          | Rejection                 | internal/server/sso_api.go         |
| sso.claims                                   | Rejection                 | internal/sso/sso.go                |
| sso.client_id                                | Rejection                 | internal/sso/sso.go                |
| sso.client_secret.empty                      | Rejection                 | internal/sso/sso.go                |
| sso.client_secret.read                       | Rejection                 | internal/sso/sso.go                |
| sso.client_secret_file                       | Rejection                 | internal/sso/sso.go                |
| sso.code                                     | Rejection                 | internal/sso/sso.go                |
| sso.discovery                                | Rejection                 | internal/sso/sso.go                |
| sso.email_missing                            | Rejection                 | internal/sso/sso.go                |
| sso.email_unverified                         | Rejection                 | internal/sso/sso.go                |
| sso.exchange                                 | Rejection                 | internal/sso/sso.go                |
| sso.groups_denied                            | Rejection                 | internal/sso/sso.go                |
| sso.id_token                                 | Rejection                 | internal/sso/sso.go                |
| sso.id_token.verify                          | Rejection                 | internal/sso/sso.go                |
| sso.idp_error                                | Rejection                 | internal/server/sso_api.go         |
| sso.issuer                                   | Rejection                 | internal/sso/sso.go                |
| sso.mapping                                  | Rejection                 | internal/sso/sso.go                |
| sso.nonce                                    | Infrastructure            | internal/sso/sso.go                |
| sso.nonce.mismatch                           | Rejection                 | internal/sso/sso.go                |
| sso.random_token                             | Infrastructure            | internal/sso/sso.go                |
| sso.redirect_url                             | Rejection                 | internal/sso/sso.go                |
| sso.state                                    | Infrastructure            | internal/sso/sso.go                |
| sso.state.mismatch                           | Rejection                 | internal/server/sso_api.go         |
| sso.unmapped_email                           | Rejection                 | internal/sso/sso.go                |
| sso.userinfo                                 | Rejection                 | internal/sso/sso.go                |
| sso.userinfo.claims                          | Rejection                 | internal/sso/sso.go                |
| sso.verifier                                 | Infrastructure, Rejection | internal/sso/sso.go                |
| store.attachment_get                         | Infrastructure            | internal/store/messages.go         |
| store.attachment_insert                      | Infrastructure            | internal/store/messages.go         |
| store.attachment_save                        | Infrastructure            | internal/messaging/service.go      |
| store.attachment_scan                        | Infrastructure            | internal/store/messages.go         |
| store.attachments_list                       | Infrastructure            | internal/store/messages.go         |
| store.attachments_rows                       | Infrastructure            | internal/store/messages.go         |
| store.close                                  | Infrastructure            | internal/store/db.go               |
| store.count_archived                         | Infrastructure            | internal/store/messages.go         |
| store.counts                                 | Infrastructure            | internal/store/sweep.go            |
| store.exec                                   | Infrastructure            | internal/store/sweep.go            |
| store.fax_create                             | Infrastructure            | internal/fax/service.go            |
| store.fax_spool                              | Infrastructure            | internal/fax/service.go            |
| store.flag_read                              | Infrastructure            | internal/store/legacy_contacts.go  |
| store.flag_write                             | Infrastructure            | internal/store/legacy_contacts.go  |
| store.message_append                         | Infrastructure            | internal/messaging/service.go      |
| store.message_insert                         | Infrastructure            | internal/store/messages.go         |
| store.message_scan                           | Infrastructure            | internal/store/messages.go         |
| store.migrate                                | Infrastructure            | internal/store/db.go               |
| store.open                                   | Infrastructure            | internal/store/db.go               |
| store.outbound_status                        | Infrastructure            | internal/messaging/service.go      |
| store.ping                                   | Infrastructure            | internal/store/db.go               |
| store.query                                  | Infrastructure            | internal/store/db.go               |
| store.sweep_scan                             | Infrastructure            | internal/store/sweep.go            |
| store.thread_find                            | Infrastructure            | internal/store/messages.go         |
| store.thread_flag                            | Infrastructure, Rejection | internal/store/messages.go         |
| store.thread_get                             | Infrastructure            | internal/store/messages.go         |
| store.thread_insert                          | Infrastructure            | internal/store/messages.go         |
| store.thread_lookup                          | Infrastructure            | internal/store/messages.go         |
| store.thread_mark_read                       | Infrastructure            | internal/store/messages.go         |
| store.thread_reread                          | Infrastructure            | internal/store/messages.go         |
| store.thread_resolve                         | Infrastructure            | internal/messaging/service.go      |
| store.thread_scan                            | Infrastructure            | internal/store/messages.go         |
| store.thread_update                          | Infrastructure            | internal/store/messages.go         |
| store.tx_begin                               | Infrastructure            | internal/store/messages.go         |
| store.tx_commit                              | Infrastructure            | internal/store/messages.go         |
| textagent.badid                              | Rejection                 | internal/textagent/client.go       |
| textagent.decode                             | Infrastructure            | internal/textagent/client.go       |
| textagent.encode                             | Infrastructure            | internal/textagent/client.go       |
| textagent.reach                              | Transient                 | internal/textagent/client.go       |
| textagent.read                               | Transient                 | internal/textagent/client.go       |
| textagent.request                            | Infrastructure            | internal/textagent/client.go       |
| textagent.secret                             | Rejection                 | internal/textagent/client.go       |
| userauth.close.db                            | Infrastructure            | internal/userauth/userauth.go      |
| userauth.close.users                         | Infrastructure            | internal/userauth/userauth.go      |
| userauth.config                              | Rejection                 | internal/userauth/userauth.go      |
| userauth.db.open                             | Infrastructure            | internal/userauth/userauth.go      |
| userauth.db.optimize                         | Infrastructure            | internal/userauth/userauth.go      |
| userauth.enroll.entropy                      | Infrastructure            | internal/userauth/enroll.go        |
| userauth.enroll.expired                      | Rejection                 | internal/userauth/enroll.go        |
| userauth.enroll.invalid                      | Rejection                 | internal/userauth/enroll.go        |
| userauth.enroll.userid                       | Corruption                | internal/userauth/enroll.go        |
| userauth.mapping                             | Rejection                 | internal/userauth/userauth.go      |
| userauth.passkey.bad_id                      | Rejection                 | internal/userauth/userauth.go      |
| userauth.passkey.unmapped                    | Rejection                 | internal/userauth/userauth.go      |
| userauth.register.exists_unreadable          | Infrastructure            | internal/userauth/userauth.go      |
| userauth.unmapped_email                      | Rejection                 | internal/userauth/userauth.go      |
| webhook.fax_pages                            | Rejection                 | internal/server/hooks/webhooks.go  |
| webphone.cli.catalog                         | Infrastructure            | cmd/webphone/catalog.go            |
| webphone.cli.contracts                       | Infrastructure            | cmd/webphone/contracts.go          |
| webphone.server.openapi_render               | Infrastructure            | internal/server/apidocs.go         |
| webphone.server.wire_render                  | Infrastructure            | internal/server/wirecontracts.go   |

<!-- error-code-registry: END -->

## 502 from the proxy — deploy-restart signature (2026-10-27)

A burst of `502` on `/events` (and any other proxied path) right after a stack
daemon relock is the webphone restart window, NOT a crash: the unit runs
`Restart=on-failure, RestartSec=5`, `ExtensionHubs.Shutdown` releases live SSE
streams instantly (2026-10-25 feature, working as designed — the release logs
`events: released live SSE streams for shutdown` with hub/stream counts), and
Caddy answers 502 for the ~5s the binary is down. The client's EventSource
recreate loop (upstream-owned, 500ms→64s backoff, reset on open) re-establishes
the feed on its own. What stays alive during the window: the session cookie
(and the tab), the shell markup, and any text the user typed into composers
(draft persistence); what dies with the process: the SIP registration and
therefore every live call. Triage bar before calling anything a bug:
`systemctl show webphone -p NRestarts -p ExecMainStartTimestamp` +
`journalctl -u webphone | grep -iE "panic|fatal"`. A recovered feed announces
itself once (the pill's announced-drop symmetry); an unrecovering feed is the
actual signal.

## Island-JS error surface — every catch lands or justifies (2026-10-27)

Owner ruling (AGENTS.md § Failure → feedback): an unhandled/uncaught error
without a user-visible surface IS a bug — Go and island JS alike. The island
side of that bar:

- EVERY `catch`/rejection either lands in a visible surface (toast/announce/
  pill/banner) or carries a written justification comment at the site. The
  2026-10-27 sweep classified all 62 sites: surfaced (transcript-save flush,
  CRM call log, summarize ask, htmx responseError/sendError toasts, session
  errors), or justified at the site (teardown-path bye/close/resume, poll
  samples, private-mode storage, warm-mic denial). The sweep log lives in the
  16:05 plan's execution log.
- `summarizeCall` failure classes are DISTINCT: 404 = "no transcript",
  network-dead (status 0) = `summarizeOffline`, other HTTP = `summarizeFailed`.
  A conflated toast is a bug of the same class.
- **THE call-failure reason mapping** (the one TODO_LIST #6's disposition work
  must align with; unknown stays a FAILURE class, never "answered"):
  486/600/603 → busy · 404 → number does not exist · 403/604 → blocked ·
  480/488/502/503/504 → far end did not answer · 487 → cancelled · 408 → no
  answer in time · any other code → `callFailedCode` with the code · inbound
  non-established → missed class · established → no reason line. Rendered on
  the ended card (`call-reason-text`) + announced once; the cause is captured
  from the INVITE `requestDelegate.onReject` (sip.js 0.21.2 resolves `invite()`
  on send — the verdict never rejects the promise).
- The readable `htmx:sseError` log line (failure counter + retry note) is the
  operator record; the ext's own `console.error("[object Event]")` spam is an
  upstream defect (draft note:
  `docs/research/2026-10-27_htmx-sse-ext-object-event-logging-note.md`).
- Division of truth, connectivity: `#offline-banner` (role=status) reflects the
  BROWSER's network (`online`/`offline` events — the machine has no route);
  the `#wp-sse-live` pill reflects the FEED (server reachable, SSE healthy).
  A red pill with a hidden banner = server-side trouble; a banner with a red
  pill = the user's own connection. They never collapse into one signal.
- Typed-rejection CONVENTION design (checkJs-era) belongs to the sibling plan
  (16:04) task T24; this file owns the runtime surfaces only.

## Detection column — how anyone would KNOW a quiet class happened (2026-10-29)

The sections above answer "what does the operator SEE when the system
speaks". Five failure classes are quiet BY DESIGN or by drift — they
answer "how would anyone know" (the ai-phone failure-table question,
M16/F69-F70). Each row: the class, why it is quiet, its detection
signature, and the surface that carries it.

| quiet class | why quiet | detection signature |
|---|---|---|
| **Caption lag** (L8, shipped 2026-10-29) | a slow provider degrades freshness, nothing errors | one polite announce + one `webphone: live transcription lagging: segment answered Ns after capture` warn PER STREAK (island `transcribe.js`); a punctual segment re-arms. A provider failing outright is the LOUD failure-streak toast instead |
| **Daemon push stall** (pma#348) | the daemon's push leg dies with its PROCESS — nothing is left to log | `scripts/push-stall-watch.sh` (the LOCAL 60-min timer is the ONLY automation home — GitHub CI checks out the remote tip and cannot see unpushed commits) + `git log origin/main..HEAD --format=%cI \| tail -1` |
| **Push-transport death** (ring-while-closed) | `/hooks/ring` answers 202 inertly ON PURPOSE — the dialplan must never depend on push health | `push.ntfy_send` / `push.fcm_send` log codes per failed publish; an `ErrUnregistered` verdict PRUNES the device row — so the tell is the `push_devices` row count dropping (the list never echoes tokens) while the phone stays silent |
| **ASR degrading to empty** | a provider returning `""` (no speech detected) is a SUCCESS — no failure path fires | the live loop skips empty-text saves silently, so the signature is the ABSENCE: `skipped a near-silent segment` lines while the operator IS talking, and the History transcript section not growing during a spoken call |
| **CRM outage** | enrichment is read-only and nil-safe; a dead CRM must never touch the call | the Settings tri-state goes red (deployed seam), the island toasts the 502 once per failed `recordCrmCall`; the quiet residue is the resolver's 5-min negative cache masking RECOVERY — the tell on the CRM side is `call_logged` journal rows stopping |

The column's contract: a new quiet class gets a row here at the same
commit that ships it (the caption-lag row landed with the feature).
Rows 2–5 are pre-existing classes retroactively documented.

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
tree, then relock). The § "Detection column" rows (2026-10-29) join
the sync set too: the daemon-stall and CRM rows describe fleet/stack
sides the runbook owns, so the next runbook pass carries the table.
