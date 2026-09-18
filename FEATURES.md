# FEATURES

Honest feature inventory by status: FULLY_FUNCTIONAL, PARTIALLY_FUNCTIONAL,
BROKEN, DISABLED (externally switched off), PLANNED, WORTH_CONSIDERING.
Code wins when doc and code disagree.

## Calling (the SIP island)

| Feature                               | Status              | Notes                                                                                     |
| ------------------------------------- | ------------------- | ----------------------------------------------------------------------------------------- |
| Register over WebSocket (wss)         | 🟢 FULLY_FUNCTIONAL | Vendored sip.js 0.21.2 browser global; no CDN, no runtime fetch                           |
| Outgoing calls (audio)                | 🟢 FULLY_FUNCTIONAL | Dial string sanitization strips Unicode direction marks, spaces, dashes                   |
| Incoming calls: accept/reject         | 🟢 FULLY_FUNCTIONAL | Second simultaneous invite auto-rejects (one call at a time)                              |
| Multiple concurrent calls             | 🟢 FULLY_FUNCTIONAL | Per-call card; hold/focus/mute; focused call gets the audio device                        |
| Blind transfer (REFER)                | 🟢 FULLY_FUNCTIONAL | Server-side execution (FreeSWITCH); NOTIFY sipfrag verdict surfaced                       |
| Attended transfer (REFER w/ Replaces) | 🟢 FULLY_FUNCTIONAL | Bridges the two established calls at the network                                          |
| DTMF keypad                           | 🟢 FULLY_FUNCTIONAL | `application/dtmf-relay` INFO, equals-form `Signal=` (FreeSWITCH-compatible)              |
| Hangup / cancel                       | 🟢 FULLY_FUNCTIONAL | Correct verb per session state (cancel before answer, bye after, reject for incoming)     |
| Reconnect watchdog                    | 🟢 FULLY_FUNCTIONAL | Bounded (5s/attempt) rebuild when `userAgent.reconnect()` hangs; calls survive            |
| Honest failure states                 | 🟢 FULLY_FUNCTIONAL | Status pill reports WHY (TLS/cert vs network vs rejected credentials), not just "offline" |

## Messaging (SMS/MMS)

| Feature                       | Status              | Notes                                                                       |
| ----------------------------- | ------------------- | --------------------------------------------------------------------------- |
| Send SMS/MMS                  | 🟢 FULLY_FUNCTIONAL | ≤1600 chars, ≤5 attachments, ≤10 MiB each; loopback + webhook gateways      |
| Inbound SMS/MMS via webhook   | 🟢 FULLY_FUNCTIONAL | `/hooks/message`, base64 attachments, Bearer secret, fail-closed            |
| Threads with unread badges    | 🟢 FULLY_FUNCTIONAL | Owner-scoped upsert; unread increments on inbound (regression-tested)       |
| Attachment round trip         | 🟢 FULLY_FUNCTIONAL | Content-addressed blob store, owner-scoped streaming, path-escape refusal   |
| Live thread list + transcript | 🟢 FULLY_FUNCTIONAL | SSE `threads`/`thread` events carry swap-safe fragments (tested)            |
| Transcript pagination         | 🟢 FULLY_FUNCTIONAL | "Load older messages" fetches prior pages (`?older=`); LIMIT+1 hasMore      |
| Delivery receipts             | 🟢 FULLY_FUNCTIONAL | `/hooks/message/status` flips by `provider_ref`; badge updates live via SSE |

## Fax

| Feature                  | Status              | Notes                                                                    |
| ------------------------ | ------------------- | ------------------------------------------------------------------------ |
| Send PDF                 | 🟢 FULLY_FUNCTIONAL | `%PDF-` sniff, ≤20 MiB, spooled to blob store                            |
| Inbound fax via webhook  | 🟢 FULLY_FUNCTIONAL | `/hooks/fax` with base64 PDF, page count, provider ref                   |
| Provider status callback | 🟢 FULLY_FUNCTIONAL | `/hooks/fax/status` flips job to transmitted/failed with error text      |
| Page-count parsing       | 🟢 FULLY_FUNCTIONAL | flexPages: `pages`/`page_count`/`num_pages` as number or string (tested) |
| Document download        | 🟢 FULLY_FUNCTIONAL | Session-gated, owner-scoped PDF streaming                                |

## Voicemail & history (phone API)

| Feature                    | Status               | Notes                                                                    |
| -------------------------- | -------------------- | ------------------------------------------------------------------------ |
| Voicemail list/play/delete | 🟢 FULLY_FUNCTIONAL  | Per-extension API with the session's credentials; needs `phone_api_url`  |
| CDR call history           | 🟢 FULLY_FUNCTIONAL  | Server-rendered tab + island panel via the same API                      |
| History search/filter      | 🟢 FULLY_FUNCTIONAL  | `?q=` text + `?dir=` in/out filter, widened fetch when filtering         |
| Honest disabled states     | 🟢 FULLY_FUNCTIONAL  | Tabs say what is missing instead of pretending when no API is configured |
| Live voicemail refresh     | 🟢 FULLY_FUNCTIONAL  | Payload-less SSE nudge on deletes and island polls                       |
| Phone-api reverse proxy    | 🟢 FULLY_FUNCTIONAL  | Same paths/JSON as the static era, Basic auth injected server-side       |
| Voicemail transcripts      | ⚪ WORTH_CONSIDERING | Surfacing only if the PBX API ever provides them                         |

## Contacts & sessions

| Feature                        | Status              | Notes                                                                                   |
| ------------------------------ | ------------------- | --------------------------------------------------------------------------------------- |
| Shared directory (config)      | 🟢 FULLY_FUNCTIONAL | Rendered into every contacts tab + island panel                                         |
| Personal contacts (server DB)  | 🟢 FULLY_FUNCTIONAL | Upsert-by-number, delete, click-to-dial into the island                                 |
| vCard import/export            | 🟢 FULLY_FUNCTIONAL | `internal/vcard`; `/contacts/import` + `/contacts/export`, upsert-by-number             |
| Single sign-on with the island | 🟢 FULLY_FUNCTIONAL | REGISTER-proven credentials open the tab session; logout closes it                      |
| Session store                  | 🟢 FULLY_FUNCTIONAL | In-memory, TTL + GC, HttpOnly cookie; lost on restart by design                         |
| Login rate limiting            | 🟢 FULLY_FUNCTIONAL | Per-IP token buckets on `/api/session` and `/hooks/*` (limiter outside the secret gate) |

## Live updates (SSE)

| Feature                    | Status              | Notes                                                              |
| -------------------------- | ------------------- | ------------------------------------------------------------------ |
| Per-extension event feed   | 🟢 FULLY_FUNCTIONAL | `/events`, heartbeats, no cross-extension leakage                  |
| Swap-safe fragments        | 🟢 FULLY_FUNCTIONAL | `threads`/`thread`/`fax` payloads never wipe a composer draft      |
| Connect after island login | 🟢 FULLY_FUNCTIONAL | `session.js` attaches `sse-connect` post-REGISTER without a reload |

## Awareness

| Feature                     | Status              | Notes                                                                                    |
| --------------------------- | ------------------- | ---------------------------------------------------------------------------------------- |
| Incoming-call notifications | 🟢 FULLY_FUNCTIONAL | System notification (permission asked from the login gesture)                            |
| Ring tone + ringback        | 🟢 FULLY_FUNCTIONAL | Locally synthesized (distinct incoming ring vs outgoing ringback)                        |
| Tab-title flash             | 🟢 FULLY_FUNCTIONAL | While an incoming call rings                                                             |
| Keyboard shortcuts          | 🟢 FULLY_FUNCTIONAL | A answer · H hangup · M mute · P hold · Esc + headset media keys (island `shortcuts.js`) |
| ICE/media diagnostics panel | 🟢 FULLY_FUNCTIONAL | Candidate path, RTT, loss, jitter, codec + plain-language hints                          |
| Event log                   | 🟢 FULLY_FUNCTIONAL | Operator-facing, English-only (runbook greps it), 100 entries                            |

## Platform

| Feature                      | Status                  | Notes                                                                                                                                                                                   |
| ---------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Single Go binary             | 🟢 FULLY_FUNCTIONAL     | `cmd/webphone`; SQLite (pure Go) + blob store; data dir auto-created                                                                                                                    |
| Nix package                  | 🟢 FULLY_FUNCTIONAL     | `buildGoModule`, tests run in the sandbox, pinned vendorHash                                                                                                                            |
| aarch64-linux                | 🟢 FULLY_FUNCTIONAL     | Cross-builds cleanly (verified 2026-09-18)                                                                                                                                              |
| Strict-CSP compatible        | 🟢 FULLY_FUNCTIONAL     | Same-origin only; `default-src 'self'` + `connect-src wss:`; no CDN                                                                                                                     |
| Security posture             | 🟢 FULLY_FUNCTIONAL     | CSRF on all mutations, security headers, owner-scoped queries everywhere                                                                                                                |
| DOM contract test            | 🟢 FULLY_FUNCTIONAL     | 35 island element ids asserted by `internal/server/server_test.go`                                                                                                                      |
| NixOS module                 | 🟢 FULLY_FUNCTIONAL     | `nixosModules.default`: hardened systemd unit, JSON settings via `WEBPHONE_CONFIG`, `environmentFile` for secrets, optional nginx WSS vhost; evalModules-checked                        |
| Import-direction arch tests  | 🟢 FULLY_FUNCTIONAL     | `internal/arch`: domain imports nothing internal, services never import server/web, island modules pairwise independent                                                                 |
| i18n (en/de)                 | 🟢 FULLY_FUNCTIONAL     | Island + server tabs (~90-key dictionary); `wp-lang` cookie / Accept-Language; SSE fragments follow the extension's language; service-validation reasons stay English (operator-facing) |
| Dark + light themes          | 🟢 FULLY_FUNCTIONAL     | Token-based, follows `prefers-color-scheme`; manual toggle cycles auto→light→dark (`wp-theme`)                                                                                          |
| Browser E2E (upstream stack) | 🟡 PARTIALLY_FUNCTIONAL | Suite exists for the v1 surface; must be re-run after the v2 switchover                                                                                                                 |

## PLANNED / WORTH_CONSIDERING

| Idea                                  | Status               | Notes                                                                 |
| ------------------------------------- | -------------------- | --------------------------------------------------------------------- |
| Session persistence across restarts   | ⚪ WORTH_CONSIDERING | Passwords in RAM only today; persistence has security cost            |
| Retention/cleanup job (blobs, old)    | ⚪ WORTH_CONSIDERING | Data grows unbounded today                                            |
| Richer /healthz (store, gateway mode) | ⚪ WORTH_CONSIDERING | For load balancers                                                    |
| Video calls                           | ⚪ WORTH_CONSIDERING | sip.js supports it; UI needs a video surface                          |
| PWA (offline shell)                   | ⚪ WORTH_CONSIDERING | Service worker must respect strict CSP                                |
| sip.js 0.22 bump                      | ⚪ PLANNED           | Evaluation report in docs/reviews/; gated on the upstream browser E2E |
