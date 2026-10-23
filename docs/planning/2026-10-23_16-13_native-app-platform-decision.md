# Native Linux + Android 16 app — platform decision (evidence pack)

- **Date:** 2026-10-23 16:13 CEST
- **Owner demand (2026-10-23):** native Linux and Android 16 apps; **calls are the
  product** ("best in class"); ringing while the app is CLOSED is a HARD
  requirement; push transport = **BOTH** ntfy (self-hosted) and FCM.
- **Status:** DECIDED + SHIPPING — owner go on P1+P2 (2026-10-23,
  second session): P0 (PWA-LITE) and P2 (the push seam:
  `internal/push`, ntfy + FCM senders, `/api/push-devices` registry,
  `/hooks/ring` lane) SHIPPED; P1 scaffolded the same day as the
  separate repo `~/projects/webphone-android` (v0.1, architecture
  complete, untested against live FreeSWITCH until the first device
  train — honesty markers in its README). The stack's FS dialplan must
  POST `/hooks/ring` on inbound calls (consuming-stack work).
- **Unblocks the parked verdict:** this demand signal is the one the
  2026-09-22 PWA spike parked the manifest behind (`../archived/2026-09-22_17-05_pwa-spike-verdict.md`).
  That verdict's core ruling stands unchanged: **NO service worker** (stale
  island = bug class invisible to every gate; offline is impossible for a
  phone).

## The one architectural fact that frames everything

The webphone is a CLIENT of the stack: data (messages, fax, voicemail,
history, contacts) and the SIP registration domain live server-side; the
device is a browser today — the ISLAND registers the device's session over
WSS (Caddy-fronted) and renders everything else from the same origin. Any
"app" is therefore a **client shell around that origin**, NOT a server on
the device. A gomobile/Wails-embedded local server cannot show the user's
messages (the gateway webhooks arrive at the stack's instance), so every
"ship the Go server inside the app" framing is wrong-topology for this
product.

## Comparison (call UX is the ranking axis)

| Option                                                          | Ring while CLOSED                                                                                              | Screen-off call survival                                                                                                | Audio quality / OS call integration                                                                                                                                                                                                                                                         | Linux story                                                                                      | Reuses island+server                                                                                               | Effort                             | Verdict                                                                                                                                                 |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **PWA-LITE** (manifest + maskable icons, no SW)                 | NO (app must be opened to answer)                                                                              | Best-effort GOOD: Chrome de-throttles tabs playing audio / using WebRTC+WebSockets; no guarantee under extreme pressure | Chrome's WebRTC stack (AEC/NS) — the same engine desktop Chrome ships                                                                                                                                                                                                                       | Chrome "Install app" = real standalone window, dock icon, auto-start                             | 100%                                                                                                               | **Days**                           | ✅ **P0 — shipped this train**                                                                                                                          |
| TWA / WebAPK (Bubblewrap APK)                                   | NO — same web platform, same push limits                                                                       | Same engine as Chrome → same as PWA                                                                                     | Same engine                                                                                                                                                                                                                                                                                 | Play distribution of the PWA                                                                     | 100%                                                                                                               | Days-weeks                         | ✖ No capability over PWA-LITE; only store presence (irrelevant for a personal stack)                                                                    |
| Android WebView shell (Kotlin wraps remote origin)              | NO                                                                                                             | **WORSE than PWA**: WebView suspends with the Activity; no Chrome de-throttling guarantees; needs FGS hacks             | getUserMedia works only via onPermissionRequest plumbing; empty device labels; no getDisplayMedia                                                                                                                                                                                           | n/a                                                                                              | 100%                                                                                                               | Weeks                              | ✖ Rejected BY CATEGORY: strictly worse call behavior than the browser it replaces                                                                       |
| **Android native-call shell** (Kotlin call layer + web content) | **YES** — FCM/ntfy high-priority push → CallStyle notification + full-screen intent; Telecom ConnectionService | **YES** — typed FGS (`microphone`/`phoneCall`); audio focus + BT routing survive backgrounding                          | **Best in class**: native libwebrtc AudioProcessing, system call UI, BT/watch/wired answer via InCallService                                                                                                                                                                                | n/a (desktop stays web — see below)                                                              | Call signaling must be re-owned by Kotlin (SIP-over-WSS + island call contract); tabs stay 100%                    | Weeks-months                       | ✅ **P1 — the recommended native path**                                                                                                                 |
| Full native Kotlin + Compose (all tabs native)                  | YES                                                                                                            | YES                                                                                                                     | Best in class (same call layer as above)                                                                                                                                                                                                                                                    | n/a                                                                                              | ~0%: needs NEW JSON APIs for every surface (today: HTML fragments + SSE), rewrites all proven UI + contract suites | Months                             | ✖ Rejected BY CATEGORY for now: the call gain over the hybrid shell is zero, the cost is the whole product                                              |
| **Kotlin Multiplatform** (Compose Android + desktop Linux)      | YES (Android)                                                                                                  | YES (Android)                                                                                                           | Android = best in class; **Linux desktop = the weak leg**: webrtc-kmp (shepeliev) has NO desktop target and is stagnating (0.125.x, ~1y quiet); Piasy kmp-webrtc linuxX64 is 2025-new (9★); webrtc-java (JVM) is the workable fallback — all far less battle-tested than any browser engine | Compose Desktop is production-ready, but the native Linux WebRTC audio path is the immature part | ~0% (same full-native cost) + a young call-engine dependency                                                       | Months, high risk on the call path | ✖ Rejected: pays full-native cost AND lands the call engine on its least mature platform                                                                |
| Flutter (flutter_webrtc: Android + Linux ✔ native)              | YES                                                                                                            | YES                                                                                                                     | Native libwebrtc both platforms — the only cross-platform stack with a REAL Linux audio path                                                                                                                                                                                                | Real native Linux app                                                                            | ~0% (full UI+API rewrite, Dart toolchain inside a Go/Nix repo — nix+flutter is notoriously painful)                | Months                             | ✖ **HARD NO (owner ruling 2026-10-23)** — option closed for ANY path, not subject to cost/benefit re-litigation                                         |
| Instant Apps                                                    | —                                                                                                              | —                                                                                                                       | —                                                                                                                                                                                                                                                                                           | —                                                                                                | —                                                                                                                  | —                                  | ✖ **DEAD platform**: Google Play Instant shutdown — since Dec 2025 Instant Apps cannot be published and the APIs no longer work (Play docs, 2026-06-24) |

## Android 16 (API 36) constraints the native path must design for (verified 2026-10-23)

- **16 KB page sizes**: any bundled native lib must be 16 KB-aligned on
  arm64/x86_64. Play release block from **2027-02-01** (apps targeting
  API 35+); new-app/updates targeting API 36 required from **2026-08-31**.
  stream-webrtc-android was NOT aligned ≤1.3.9, reported aligned at
  **1.3.10** (2026-03) — verify per release (`zipalign -c -P 16`), never
  assume.
- **USE_FULL_SCREEN_INTENT** (the incoming-call wake): Android 14+ made it
  special; calling/alarm apps qualify for the default grant. **Sideloaded**
  apps keep the default-enabled grant; Play revokes it for apps that don't
  declare calling as core function. Sideload (this product's distribution)
  is the favorable lane. Runtime-check `canUseFullScreenIntent()` and
  degrade to heads-up (60s pill) gracefully.
- **FGS types**: `microphone` (+ `phoneCall` with `MANAGE_OWN_CALLS` or
  ROLE_DIALER) required, declared + typed at startForeground; camera/mic
  FGS cannot START from the background (the push → activity/CallStyle
  full-screen intent is the legitimate launch path). Android 15+: no
  `phoneCall`/`microphone` FGS from BOOT_COMPLETED.
- **TargetSdk 36 behavior**: edge-to-edge enforced (no opt-out), predictive
  back on (`onBackPressed` gone), orientation locks ignored ≥600dp, 16 KB
  compat-mode dialog for unaligned libs. All call-UI-relevant.
- **Notification.CallStyle** (`forIncomingCall`) + ContactKey: the ONLY way
  to render as a real incoming call (ringtone, accepted/declined, lockscreen
  prominence). Web surfaces cannot reach it — this is the native shell's
  core value.

## Push transport (owner: BOTH)

Server-side `push` seam in webphone (`internal/push`, shaped like the
asr/crm/paperless optional seams — both-or-neither config per transport,
nil-safe everywhere):

- `push.ntfy.*` — self-hosted: topic + token; GDPR-friendly, no Google;
  delivery subject to Doze/battery-manager variance.
- `push.fcm.*` — FCM HTTP v1 (service-account credentials); reliable on
  GMS devices; needs a Google Cloud project.

Trigger seam: the stack signals "inbound call for extension X" (FS dialplan
→ the existing webhook-bridge pattern) → webphone push seam fans out to all
configured transports → the Android shell rings with CallStyle; tap opens
the app which (re)establishes the WSS registration and answers.

## Rejected framings (by category, with the deciding evidence)

- **Tauri v2 + Go sidecar**: Android sidecars UNSUPPORTED (tauri#9774,
  open 2025-10; Android 10+ blocks executing bundled binaries) — and the
  sidecar is wrong-topology anyway (see the framing fact above).
- **Wails v3**: desktop stable-beta, Android still EXPERIMENTAL
  (v3.0.0-beta.34, 2026-10-20) — and its in-process Go adds nothing for a
  remote-origin client.
- **gomobile-embedded server**: wrong-topology (device cannot host the
  user's data; no trunk on-device).
- **Flutter**: HARD NO by owner ruling (2026-10-23) — closed regardless
  of its unique position as the only cross-platform stack with a real
  native Linux WebRTC path. Recorded so no future session re-proposes
  it; the Linux call story stays the browser engine (P0), per the phases
  below.

## Phases

- **P0 — PWA-LITE (this train, DONE)**: `manifest.webmanifest` + maskable
  192/512 PNGs + head link + root endpoint. NO service worker (NOT-DO
  stands). Buys: Android 16 home-screen WebAPK-style install, Linux
  standalone window via Chrome install, dock/taskbar presence, splash
  colors. Served-markup change → owes the stack browser E2E at next
  release (runbook obligation).
- **P1 — Android native-call shell** (new repo or `android/` subtree —
  owner call): Kotlin; native call layer (Telecom ConnectionService,
  CallStyle, FGS microphone/phoneCall, native WebRTC via
  stream-webrtc-android ≥1.3.10) re-owning the island's call contract
  (SIP-over-WSS REGISTER/INVITE, hold re-INVITE, DTMF
  `application/dtmf-relay` equals-form, REFER/NOTIFY transfer, bounded
  watchdog); tabs stay web (WebView into the same origin, session shared).
  Push per the seam above. Distribution: sideload APK first (FSI grant
  stays enabled), Play later if ever.
- **P2 — webphone push seam** (this repo): `internal/push` with ntfy + FCM
  senders, config-gated, plus the stack's inbound-call webhook → push
  fan-out. Can land BEFORE P1 (a push with no client is inert) but is only
  verifiable end-to-end once P1 exists.
- **Linux**: P0 IS the native-app experience (Chrome engine = the same
  best-in-class WebRTC desktop Chrome ships). A packaged GTK/Flutter
  native window would add a toolchain for zero call-quality gain — revisit
  only on a named trigger (e.g. Chrome-independent distribution demand).

## Unverified cells (honesty contract — no fabricated claims)

- **liblinphone SIP-over-WSS transport support** (would matter only for a
  full-native variant that swaps the island's sip.js for a turnkey stack):
  research blocked this session (search infra outage); treat as UNKNOWN.
  If P1 ever prefers liblinphone over native-libwebrtc + hand-owned
  signaling, run a 1-day spike FIRST (verify-external-claims: read the
  linphone-sdk source, not marketing pages). Note the server executes
  transfers on REFER server-side (FreeSWITCH) — any native stack must
  reproduce the island's NOTIFY sipfrag verdict parsing.
- **Stream-webrtc-android 16 KB alignment**: claimed aligned at 1.3.10 by
  a third-party reporter (issue #312, 2026-03); VERIFY per shipped
  release before pinning.
