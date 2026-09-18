# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Message delivery receipts: providers call `/hooks/message/status`
  (Bearer secret, fail-closed like every hook) with the `provider_ref`
  from the send receipt; the transcript's status badge flips to
  delivered/failed live over SSE, mirroring the fax status callback.

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
