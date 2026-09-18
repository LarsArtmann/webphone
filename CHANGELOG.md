# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- cqrs-htmx middleware adoption (the 2026-09-18 Pareto plan's 1%/4%/20%
  tiers, `docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`):
  the hand-rolled panic `recovery()` is now `cqrshtmx.RecoveryMiddleware`
  (full stack trace + method/path in the log, `http.ErrAbortHandler`
  re-raised per net/http convention); `cqrshtmx.RequestLoggingSlog` sits
  outermost so every request leaves one structured log line (200/404/401/
  429/SSE disconnects) with no bodies or credentials; the 83-line
  `keyedLimiter` is deleted in favor of `httputil.KeyedRateLimiter`
  (TTL-evicted per-key buckets, `MaxKeys`-cappable, computed
  `Retry-After`; keys stay port-stripped peer hosts until the stack
  proves XFF sanitization); `/healthz` no longer answers a constant "ok" —
  it serves `cqrshtmx.ReadinessHandler` with named `sqlite` (ping) and
  `blob-dir` (write probe) checks, 503 bodies name the failing check;
  and the SSE `events` handler's 36-line hand loop collapsed onto
  `Broadcaster.ServeSSE` (adds the library's `connected` handshake frame;
  `threads`/`thread`/`fax`/`voicemail` payloads untouched; no `retry:`
  hint at v4.9.0 — tag-verified).

### Added

- Message delivery receipts: providers call `/hooks/message/status`
  (Bearer secret, fail-closed like every hook) with the `provider_ref`
  from the send receipt; the transcript's status badge flips to
  delivered/failed live over SSE, mirroring the fax status callback.
- German for the server-rendered tabs: ~90-key en/de dictionary, chosen
  per extension (settings tab, remembered server-side), with a header
  toggle that writes the `wp-lang` cookie and hot-swaps the tab without a
  reload; SSE-pushed fragments follow the extension's language.
- NixOS module (`nixosModules.default`): hardened systemd unit, all
  settings via a JSON file (`WEBPHONE_CONFIG`), `environmentFile` for
  secrets, optional nginx vhost with the WSS `/sip` proxy; evaluated by
  a flake check. The stale v1 `package/default.nix` is gone.
- Login rate limiting: per-IP token buckets on `/api/session` and the
  inbound `/hooks/*` endpoints (the hook limiter sits outside the
  secret gate so unknown secrets are throttled too).
- vCard import/export for personal contacts: `/contacts/import`
  (upsert-by-number) and `/contacts/export`.
- Transcript pagination: "Load older messages" fetches prior pages
  instead of a fixed window.
- History search/filter: `?q=` text and `?dir=` direction, with a wider
  fetch while filtering.
- Keyboard shortcuts in the island: A answer, H hangup, M mute, P hold,
  Esc reject/cancel, plus headset media keys.
- Manual theme toggle: cycles auto → light → dark, persists the choice
  (`wp-theme` in localStorage), overrides `prefers-color-scheme`.
- Fax page-count parsing: provider status payloads may quote pages as
  number or string (`pages`/`page_count`/`num_pages`); stored as a
  count.
- Architecture enforcement tests (`internal/arch`): the domain imports
  nothing internal, services never import server/web, and the island's
  calls/ice/connection modules stay pairwise independent.
- Expanded test coverage: config loader, gateway webhook mode, pbx
  client error paths, login rate limiting, history filter, vCard, i18n
  dictionary sync, unread-cache invalidation, webhook gates.

### Changed

- The unread badge count is cached per extension (5s TTL, invalidated
  on inbound/status/deletes) instead of rescanning every thread on each
  shell render.

### Fixed

- The on-screen Accept and Reject buttons did nothing: the island's
  module split left them calling `answerIncoming`/`rejectIncoming`
  without importing those functions, so every click died with a silent
  ReferenceError (keyboard shortcuts kept working). Incoming calls
  could not be answered from the UI; the upstream browser E2E caught
  it (callee never sent its 200 OK, the PBX timed the ring out at 60s).
- Strict-CSP console errors on every page load: htmx no longer injects
  its inline indicator styles (disabled via the `htmx-config` meta, with
  htmx deferred after it and the same rules shipped in `app.css`); the
  error panel's reload button uses a delegated `data-reload` listener
  instead of an inline `onclick`; and the framework's theme-preload
  script is allowed by exact CSP hash, two-way-guarded by
  `TestServedPageSatisfiesStrictCSP`.
- `/favicon.ico` 404: the shell now declares
  `<link rel="icon" href="/favicon.svg">`.
- The phone-api client percent-encoded `?` in request URLs
  (`history%3Flimit=30`), so history/voicemail requests 404'd upstream;
  path and query are now joined correctly (caught by the new client
  tests).

## [2.0.0] - 2026-09-18

### Added

- **Go unified-communications service** replacing the static site: the
  proven SIP call island plus server-rendered tabs for Messages
  (SMS/MMS), Fax, Voicemail, History, Contacts and Settings on one page
  (templ + HTMX partial swaps — the island never unloads, calls survive
  tab switches).
- Messaging: threads with unread badges, attachments in and out
  (content-addressed blob store), validation limits, live updates over a
  per-extension SSE feed (`threads`, `thread`, `fax`, `voicemail`).
- Fax: send PDFs (signature-checked, ≤20 MiB), receive documents via
  webhook, provider status callbacks, owner-scoped downloads.
- Voicemail and CDR history through the per-extension phone API, both
  directly and via a transparent `/phone-api/*` reverse proxy that
  injects the session's Basic credentials server-side.
- Gateway seam for outbound traffic: `loopback` (zero dependencies —
  everything works with no PBX) and `webhook` (multipart to a provider
  URL, `{"provider_ref"}` receipts). Inbound webhooks `/hooks/message`,
  `/hooks/fax`, `/hooks/fax/status` are guarded by a Bearer secret and
  fail closed when none is configured.
- Session model: the island's SIP REGISTER proves the extension
  credentials; the tabs share that login via a cookie session, and the
  SSE feed connects without a page reload after sign-in.
- Single-binary Nix package (`buildGoModule`, tests run inside the
  sandbox); aarch64-linux cross-build verified.
- Go test suite: DOM contract (34 island element ids), store lifecycle
  (caught a real unread-upsert bug), SSE fragment pushes, phone-api
  proxy, webhook gates, fax status callbacks.
- Integration contracts (gateway, hooks, phone API, `window.PBX_CONFIG`,
  FreeSWITCH bridge example, reverse-proxy deployment) documented in the
  README.

### Changed

- The static-site derivation is gone: consumers run the service binary
  and put TLS + the WSS `/sip` proxy in front (see README deployment).
- `window.PBX_CONFIG` is rendered by the server at `/config.js` from its
  own configuration instead of being supplied by the serving PBX.

### Fixed

- SSE payloads are swap-safe fragments: live pushes can no longer nest
  panels or wipe a half-typed composer draft, and open transcripts now
  update live (`thread` events were defined but never published before).
- The voicemail tab refreshes live on deletes and island voicemail polls
  (payload-less SSE nudge; the event was previously never sent).
- The `/phone-api` proxy rides a timeout-bounded HTTP client instead of
  the timeout-less `http.DefaultClient`.
- The phone-API client built request bodies it then never sent.
- A missing data directory is created at startup instead of failing with
  SQLite's cryptic "unable to open database file (14)".

## [0.1.0] - 2026-09-17

### Added

- Standalone webphone UI, extracted from
  [nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
  `packages/webphone` (the UI lived there since v0.1.0 of that stack).
  That stack now consumes this flake as its `webphone` input.
- The 1590-line single-file app is restructured into ES modules
  (`src/app/`: config, i18n, ui, state, auth, audio, notify, panels,
  ice, connection, calls, main) bundled by esbuild into one
  same-origin `app.js` — unchanged behavior, reviewable pieces.
- Light theme alongside dark: full design-token system following
  `prefers-color-scheme` (previously dark-only).
- LED-style status pill, tactile keypad/button press states, focus
  rings, `prefers-reduced-motion` support, incoming-call pulse, stricter
  toast styling — same DOM contract as before (see AGENTS.md).
- `package/update.sh` for repinning the bundled sip.js tarball.

[Unreleased]: https://github.com/LarsArtmann/webphone/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.0.0
[0.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v0.1.0
