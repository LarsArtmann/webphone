# Status Report: WebPhone v2.0 — SSE Liveness, Proxy Hardening, Docs Overhaul

**Date:** 2026-09-18 16:21
**Scope:** This session only (~16:00–16:21). It closed the named gaps
from `docs/status/2026-09-18_15-25_webphone-v2-rebuild-status.md`
(items 1, 2, 6, 9 + docs) and nothing else. All statements below are
based on work performed and observed in this session.

> Format note: the status-report skill's canonical format is HTML; the
> user explicitly requested Markdown — honored.

---

## Opening self-critique (the three questions)

**What did I forget?** Three things, all caught late: nav badges don't
live-update (noticed mid-session, only now documented); the `/events`
HTTP endpoint has no dedicated Go test (hub layer does); the 14-check
smoke suite lives in `/tmp` and evaporates on reboot.

**What could I have done better?** Start simpler. My first notifier
design was an over-clever generic type-switch; my first fax-status test
draft was literal placeholder junk that briefly sat in the file; my
smoke script failed three times on MY wrong expectations before it
tested anything real; and I asserted "34 contract ids" from the previous
session's summary instead of counting (it's 35).

**What can still be improved?** The liveness story is server-verified
only — no browser ever executed the SSE swaps this session. The nav
badges, transcript mark-read, and i18n for tabs remain open. Details in
(e)/(f).

---

## a) FULLY DONE ✅

### SSE liveness (the report's #2 gap)

- **Swap-safe fragments:** `ThreadsList`, `Transcript`, `FaxList`
  extracted in the views; `sse-swap` now targets inner list regions
  (`.wp-thread-list`, `.wp-fax-list`, `#thread-transcript`). Live
  pushes can no longer nest panels or wipe a half-typed composer draft
  — the old payloads (full panel incl. its own `sse-swap` wrapper)
  silently nested a section on every push.
- **`thread` events:** `Notifier.MessagesChanged` now publishes the
  affected thread's transcript — open conversations update live.
  Previously `sseEventThread` was defined but never sent.
- **Voicemail nudge:** the `voicemail` event is deliberately
  payload-less (the panel needs per-session PBX credentials the
  notifier must never hold); the panel re-fetches its partial on
  receipt. Wired into `deleteVoicemail` and into the `/phone-api`
  proxy (successful island voicemail polls — i.e. after registration
  and after each ended call — refresh an open voicemail tab). The dead
  `Notifier.VoicemailChanged` was deleted (ghost code removed).
- **SSE connects at island login:** `session.js` attaches
  `sse-connect="/events"` to `.wp-root` and calls `htmx.process` after
  the REGISTER-proven session POST — no reload, so the in-memory
  password and any call survive. Logout reloads the page (safe by
  construction; `keepalive` DELETE so the session dies even mid-nav).

### Proxy & client hardening (report's #6 gap)

- `/phone-api` proxy rides `pbx.Client.HTTPClient()` (15s timeout)
  instead of the timeout-less `http.DefaultClient`.
- Fixed a real latent bug in `pbx.Client.do()`: request bodies were
  JSON-encoded and then never sent (`req.Body = http.NoBody`); now the
  reader is passed to `NewRequestWithContext` with a Content-Type.
- Missing data dir is created at startup (0700) — previously the
  server died with SQLite's cryptic "unable to open database file
  (14)". Found by the smoke suite, not by code review.

### Tests (all green, `-count=1`)

- `TestSSEPushesSwapSafeFragments` — threads/thread/fax events carry
  fragments; asserts the forbidden strings (`<section`, `wp-compose`,
  `wp-back`, `wp-panel-head`) stay out of payloads.
- `TestPhoneAPIProxyInjectsCredentialsAndNudgesVoicemail` — Basic-auth
  injection (exact header), JSON/content-type passthrough, voicemail
  nudge fires on voicemail paths, does NOT fire on history (negative),
  unknown upstream paths → 404.
- `TestPhoneAPIProxyDisabledReturns503`, `TestFaxStatusWebhookUpdatesJob`
  (202 + panel shows failure; unknown ref → 404; invalid status → 400).
- Harness: `testServer` exposes hubs/faxes/phoneAPI;
  `newTestServerWithPhoneAPI`, `clientFor` helpers.

### Verification

- Live smoke over real HTTP (fresh binary, fresh data dir): **14/14
  PASS** — CSRF, session, signed-in SSE connect, voicemail trigger,
  thread list + transcript swap regions, unknown hook 404, inbound
  webhook → live `threads` + `thread` events with bubble fragments and
  no nested sections, open-thread view, proxy 503.
- `go build/vet/test ./...` ✅ · `nix fmt` ✅ · `nix build .#webphone` ✅
  · `nix flake check` + `--all-systems` ✅ · **aarch64-linux cross-build
  ✅** (build verified; no aarch64 runtime host) ·
  `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` **exit 0** ✅.
- `update.sh` shfmt findings (pre-existing, uncommitted at session
  start) fixed via `buildflow -s shfmt --fix`; shellcheck + `bash -n`
  clean.

### Docs overhaul (the report's #1 gap) — docs-health BUILD+HARVEST+VERIFY

- **README** rewritten for v2: quick start (loopback = whole product,
  zero PBX), full config reference incl. the `__` env convention,
  integration contracts (PBX_CONFIG, phone API, outbound gateway,
  inbound hooks), FreeSWITCH bridge example, nginx deployment shape,
  dev commands.
- **FEATURES** v2 inventory with honest statuses (incl. i18n
  PARTIALLY_FUNCTIONAL: island en/de, tabs en-only; upstream browser
  E2E PARTIALLY_FUNCTIONAL pending switchover).
- **TODO_LIST** harvested: report f-list items done this session routed
  to CHANGELOG, the rest kept with evidence citations; stale v1 items
  merged/dropped.
- **CHANGELOG** 2.0.0 entry (Added/Changed/Fixed).
- **AGENTS.md** rewritten for v2: commands (GOEXPERIMENT, templ path),
  DOM/bundle contract (35 ids + E2E hooks), architecture invariants
  (island never unloads, acyclic graphs, session model, SSE fragment
  rule, gateway seam, owner scoping, CSP), hard-won knowledge.
- CONTRIBUTING templ command aligned; stale `config.Load()` comment
  (old env convention) fixed; cross-file consistency checked (one
  status conflict fixed: session persistence now PLANNED in both
  FEATURES and TODO_LIST).

## b) PARTIALLY DONE ⚠️

- **README deployment section** is written under the current
  ASSUMPTION (the telephony stack owns TLS/WSS/nginx and reverse-proxies
  to the binary). It becomes wrong the day the user decides this repo
  ships the NixOS module — the decision is still open (see g).
- **SSE verification depth:** hub-level Go tests + one live HTTP smoke
  pass. No browser executed the actual innerHTML swaps this session
  (reasoned against the vendored sse.js source: default swap spec is
  innerHTML, `hx-trigger="sse:…"` is supported by the extension's own
  listener path).
- **Island JS behaviors** (`connectLiveUpdates` → `htmx.process`,
  logout reload): CSP-safe and simple, but reasoning-verified only.

## c) NOT STARTED ❌

Deliberately routed to TODO_LIST, untouched this session: German
translations for tabs; login rate limiting; session persistence; SMS
delivery-receipt webhook; NixOS module (blocked on g1); upstream
switchover in nix-international-telephony; message pagination; vCard
import/export; history search; manual theme toggle; keyboard/media
keys; `countUnread` caching; sip.js 0.22 evaluation; fax page-count
parsing; retention/cleanup; richer /healthz.

## d) TOTALLY FUCKED UP! 💥

Nothing shipped is broken (every gate green at close). The honest
mishap list:

1. **Placeholder junk in a test file:** my first draft of
   `TestFaxStatusWebhookUpdatesJob` contained `_ = statusPayload`-style
   filler scaffolding — literally garbage code briefly in the file.
   Caught on read-back and fully rewritten. Drafting in the editor
   instead of thinking first.
2. **Three smoke failures that were my fault:** I asserted the
   voicemail trigger and transcript swap region on `/`, which renders
   the default tab — I debugged a correct app twice. Then a partial
   multiedit ("1 of 2 applied") left stale lines I only caught via
   `rg`. Lesson re-earned: verify file state after every partial edit;
   derive smoke expectations from the routing table before writing
   assertions.
3. **"34 contract ids" was wrong:** I trusted the previous session's
   summary instead of counting. VERIFY counted: **35**. Fixed in
   FEATURES + AGENTS. (Compute counts from the repo, always.)
4. **Over-clever first design:** the initial notifier refactor used a
   generic `renderFragment[P]` with a type switch on props — discarded
   for the simple three-read version after LSP failed. One wasted
   design cycle; "simple first" is cheaper.
5. **Product gap found by smoke, not by review:** the missing-data-dir
   crash should have been caught reading `main.go` during research. It
   took a failed smoke run to see it.

## e) WHAT WE SHOULD IMPROVE! 🛠️

1. **Nav badges don't live-update** (unread messages, new voicemail):
   they render with the shell only. SSE events never touch them. New
   finding this session — now documented, needs a decision (badge swap
   target vs badge-refresh endpoint).
2. **Live transcripts don't mark threads read:** the server correctly
   cannot know whether any tab actually swapped the fragment (one hub
   per extension, many tabs), so an inbound message read live still
   badges the thread as unread until re-opened. Honest default; a
   client-side `htmx:sseMessage` → POST mark-read would be
   visibility-accurate.
3. **`connectLiveUpdates` assumes `window.htmx` exists** at REGISTER
   time (near-certain — user gesture long after deferred loads — but
   unguarded against exotic orders).
4. **`/events` endpoint lacks a Go test** (anonymous 401, heartbeat on
   the wire, per-extension isolation). Hub layer is tested; the HTTP
   handler is smoke-covered only.
5. **The smoke suite is ephemeral** (`/tmp/webphone-smoke.py`): 14
   checks that prove the product works should live in the repo.
6. **No browser-level test of tab SSE behavior anywhere** (upstream E2E
   covers the island only).
7. **Voicemail nudge fires on every successful island poll** — each
   nudge costs an open voicemail tab 2 upstream calls. Fine now; a
   content-hash guard would make it free.
8. **README `nix run github:LarsArtmann/webphone`** is shape-verified
   (mainProgram set) but was never executed (no network fetch).
9. **vulnix noise:** every full buildflow run warns about nixpkgs
   CVEs (binutils/glibc/bison) — upstream churn, not actionable here,
   but the noise desensitizes.
10. **Trusting summaries over repos** (the "34" incident) — standing
    risk every session; VERIFY-style counting is the antidote.

## f) Up to 50 things to get done next (impact-sorted)

1. Deployment decision + NixOS module (or stack-side guide) — blocked on g1.
2. Upstream switchover in nix-international-telephony (input swap, WSS proxy, config migration, browser E2E re-run).
3. Client-side mark-read on live transcript swap (`htmx:sseMessage`).
4. Nav badge live updates (unread + new voicemail).
5. German translations for the server-rendered tabs.
6. Login rate limiting on `/api/session` (and hooks beyond the secret).
7. SMS delivery-receipt webhook (`provider_ref` column already exists).
8. `/events` Go tests: anonymous 401, heartbeat, cross-extension isolation.
9. Promote the 14-check smoke suite into the repo (script or Go integration test).
10. Session persistence decision (SQLite sessions vs documented in-memory + security note).
11. Browser E2E for tab SSE behavior (chromium in CI).
12. Harden `connectLiveUpdates` (retry until `htmx` is present).
13. Voicemail nudge content-hash guard (skip no-op re-fetches).
14. Message pagination/virtualization beyond the 200 window.
15. vCard contact import/export.
16. History search/filter.
17. Manual theme toggle (override `prefers-color-scheme`).
18. Keyboard shortcuts (answer/hangup/mute/hold) + media keys.
19. Cache `countUnread` (scans all threads per shell render).
20. sip.js 0.22 evaluation (bundle-contract strings must survive).
21. Fax page-count parsing from provider payloads.
22. Retention/cleanup job (blobs + old messages/faxes).
23. Richer `/healthz` (store ping, gateway mode) for load balancers.
24. Voicemail transcript surfacing (if the PBX API ever provides it).
25. Tag the v2.0.0 release (CHANGELOG entry exists; see g3).
26. Short-lived TURN REST credentials rendered into `/config.js` (now trivial server-side; v1 design anticipated it).
27. Idempotency keys on `/hooks/message` (dedupe provider retries).
28. Attachment MIME sniffing server-side (don't trust declared type).
29. Structured request logging (slog middleware) for operators.
30. Metrics endpoint (or JSON mode on /healthz).
31. Shared i18n dictionary guard between island and tabs (prevent two drifting dictionaries).
32. Accessibility pass on tabs (focus order after HTMX swap, aria-live on SSE regions).
33. Draft persistence for composers (localStorage) — complements swap-safety.
34. Image attachment thumbnails inline in bubbles (link-only today).
35. Fax cover-page templating.
36. SQLite backup/restore runbook (`.backup` procedure documented or exposed).
37. Per-extension data export (messages + faxes dump).
38. Multiple-tab glare warning (two tabs registering one extension).
39. SSE connection-loss banner (the ext already retries with backoff).
40. Webhook payload versioning header (`X-Webphone-Event-Version`).
41. Timezone-aware timestamp rendering (server-local today).
42. Server-side PDF page counting when the provider omits pages.
43. Bridge the SSE voicemail nudge to the island's own badge refresh (one event, two consumers).
44. CSP nonce mode option for stricter deployments.
45. Rate-limit response headers on hooks (transparency for providers).
46. Dependabot/buildflow update cadence as new deps land over time.
47. `docs/status` pruning policy (archive cadence for old reports).
48. CONTRIBUTING: document the dev smoke loop once (9) lands.
49. Consider the `besteffort` helper if the 29 `//nolint:erraudit` sites ever feel noisy (deliberately untouched).
50. Reusable pattern doc: "HTMX tabs + island + session shell" (third LarsArtmann app with this shape).

## g) Questions I cannot figure out myself

1. **Deployment ownership:** should THIS repo ship the NixOS module
   (systemd + nginx vhost + WSS `/sip` proxy), or does
   nix-international-telephony keep owning TLS/proxying and merely
   reverse-proxies to the binary (my documented assumption in README +
   AGENTS)? This decides the shape of (f) 1–2.
2. **New-conversation UX:** after sending the FIRST message to a
   number (from the list composer), should the app open that thread
   view instead of returning to the list? (Carried from the 15:25
   report; replies already keep the thread open.)
3. **Release timing:** cut and tag **v2.0.0** now (CHANGELOG is ready),
   or only after the upstream switchover + browser E2E is green in
   nix-international-telephony?

---

**Bottom line:** the product's named gaps (SSE liveness, proxy
hardening, missing tests, stale docs) are closed and every gate is
green — build, vet, tests, nix (both systems), buildflow full, and a
14-check live smoke. What remains is a decision (deployment/release), a
switchover in the consuming repo, and the honest backlog above.

**WAITING FOR INSTRUCTIONS.**
