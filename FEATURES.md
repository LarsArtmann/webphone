# FEATURES

Honest feature inventory by status: FULLY_FUNCTIONAL, PARTIALLY_DONE,
PLANNED, WORTH_CONSIDERING.

## Calling

| Feature                                    | Status              | Notes                                                                                                  |
| ------------------------------------------ | ------------------- | ------------------------------------------------------------------------------------------------------ |
| Register over WebSocket (wss)              | 🟢 FULLY_FUNCTIONAL | sip.js 0.21 bundled as a browser global; no CDN, no runtime fetch                                      |
| Outgoing calls (audio)                     | 🟢 FULLY_FUNCTIONAL | Dial string sanitization strips Unicode direction marks, spaces, dashes                                |
| Incoming calls: accept/reject              | 🟢 FULLY_FUNCTIONAL | Second simultaneous invite auto-rejects (one call at a time)                                           |
| Multiple concurrent calls                  | 🟢 FULLY_FUNCTIONAL | Per-call card; hold/focus/mute; focused call gets the audio device                                     |
| Blind transfer (REFER)                     | 🟢 FULLY_FUNCTIONAL | Server-side execution (FreeSWITCH); NOTIFY sipfrag verdict surfaced                                    |
| Attended transfer (REFER w/ Replaces)      | 🟢 FULLY_FUNCTIONAL | Bridges the two established calls at the network                                                       |
| DTMF keypad                                | 🟢 FULLY_FUNCTIONAL | `application/dtmf-relay` INFO, equals-form `Signal=` (FreeSWITCH-compatible)                           |
| Hangup / cancel                            | 🟢 FULLY_FUNCTIONAL | Correct verb per session state (cancel before answer, bye after, reject for incoming)                  |

## Resilience

| Feature                       | Status              | Notes                                                                                     |
| ----------------------------- | ------------------- | ----------------------------------------------------------------------------------------- |
| Reconnect watchdog            | 🟢 FULLY_FUNCTIONAL | Exponential backoff, 5s bound per attempt, full UA rebuild when `reconnect()` hangs       |
| Call preservation             | 🟢 FULLY_FUNCTIONAL | Re-registration after transport loss keeps live calls and says so                         |
| Honest failure states         | 🟢 FULLY_FUNCTIONAL | Status pill reports WHY (TLS/cert vs network vs rejected credentials), not just "offline" |

## Awareness

| Feature                     | Status              | Notes                                                            |
| --------------------------- | ------------------- | ---------------------------------------------------------------- |
| Incoming-call notifications | 🟢 FULLY_FUNCTIONAL | System notification (permission asked from the login gesture)    |
| Ring tone + ringback        | 🟢 FULLY_FUNCTIONAL | Locally synthesized (distinct incoming ring vs outgoing ringback) |
| Tab-title flash             | 🟢 FULLY_FUNCTIONAL | While an incoming call rings                                     |
| ICE/media diagnostics panel | 🟢 FULLY_FUNCTIONAL | Candidate path, RTT, loss, jitter, codec + plain-language hints  |
| Event log                   | 🟢 FULLY_FUNCTIONAL | Operator-facing, English-only (runbook greps it), 100 entries    |

## Data

| Feature                      | Status              | Notes                                                                     |
| ---------------------------- | ------------------- | ------------------------------------------------------------------------- |
| Local call history           | 🟢 FULLY_FUNCTIONAL | localStorage, redial + save-as-contact                                    |
| Server CDR history           | 🟢 FULLY_FUNCTIONAL | Via per-extension phone API (`/phone-api/history`), Basic auth            |
| Voicemail: list/play/delete  | 🟢 FULLY_FUNCTIONAL | Via phone API with expiring stream tokens; new-message badge              |
| Contacts: shared + personal  | 🟢 FULLY_FUNCTIONAL | Shared from config.js, personal in localStorage, click-to-dial            |
| Remember extension           | 🟢 FULLY_FUNCTIONAL | Extension only — the password is never persisted                          |

## Platform

| Feature                        | Status              | Notes                                                                        |
| ------------------------------ | ------------------- | ---------------------------------------------------------------------------- |
| Nix package (static site)      | 🟢 FULLY_FUNCTIONAL | `packages.webphone`; esbuild-bundled, pinned sip.js tarball                  |
| aarch64-linux                  | 🟢 FULLY_FUNCTIONAL | Pure esbuild + tarball build, cross-builds cleanly                           |
| i18n (en/de)                   | 🟢 FULLY_FUNCTIONAL | Browser-language detection, switcher, persisted                              |
| Dark + light themes            | 🟢 FULLY_FUNCTIONAL | Token-based, follows `prefers-color-scheme`                                  |
| Accessibility floor            | 🟢 FULLY_FUNCTIONAL | focus-visible rings, aria-live toasts, labeled controls, reduced-motion      |
| Strict-CSP compatible          | 🟢 FULLY_FUNCTIONAL | Same-origin only; works under `default-src 'self'` + `connect-src wss:`      |
| Browser E2E (upstream stack)   | 🟢 FULLY_FUNCTIONAL | Two chromiums register + call through the real PBX (owned by the consumer)   |

## PLANNED / WORTH_CONSIDERING

| Idea                                   | Status               | Notes                                                          |
| -------------------------------------- | -------------------- | -------------------------------------------------------------- |
| Manual theme override (toggle)         | 🟡 PARTIALLY_DONE    | Tokens exist; only the media-query hook is wired               |
| Video calls                            | ⚪ WORTH_CONSIDERING | sip.js supports it; UI needs a video surface + layout decision |
| PWA (offline shell + install prompt)   | ⚪ WORTH_CONSIDERING | Service worker must respect strict CSP + same-origin only      |
| Opus/DTX preference tuning             | ⚪ WORTH_CONSIDERING | Current SDP is sip.js defaults                                 |
| Headset media keys (mute/answer)       | ⚪ WORTH_CONSIDERING | Keyboard shortcuts as a stepping stone                         |
