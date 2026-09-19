# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Liveness + startup probes complete the health triple: `GET /livez`
  (go-health v0.3.0 `NewChecks`, container-free — fetch-free process
  liveness, 200 while the process serves, no checks run) and
  `GET /startupz` (503 until the `sqlite` + `blob-dir` checks first
  pass, then latched 200). Both are JSON, session-free GETs that share
  `/healthz`'s check functions — readiness keeps a single home at
  `/healthz` (no second readiness truth). Readiness checks are bounded
  by cqrs-htmx v4.11.0's `NamedCheck.Timeout` (2s per check — F1 of the
  2026-09-19 DI/health review, fixed upstream and consumed here; the
  local `boundedCheck` wrapper was deleted). Also ships the Go 1.27.1
  fleet floor (go.mod + flake builder/devShell on `go_1_27`) required
  by go-health v0.3.0 and cqrs-htmx v4.11.0; samber/do appears only as
  go-health's transitive dep — the container stays rejected. New probe
  tests pin the split: broken deps degrade `/startupz` to 503 while
  `/livez` stays 200. F2 liveness decision + fleet/CSP options:
  `docs/architecture-understanding/2026-09-19_20-59_health-probes-fleet-options.md`.

## [2.2.0] - 2026-09-19

### Security

- Login (`POST /api/session`) now verifies the submitted
  extension/password against the PBX directory before a session is
  minted, failing closed (401 rejected credentials, 502 PBX
  unreachable). Previously the endpoint trusted the island's claim
  that a SIP REGISTER had succeeded; a forged request could open a
  session scoped to any extension and read that extension's stored
  message threads, fax documents and contacts (the tab partials,
  attachment streams and SSE fragments scope by the session alone).
  Found by live-probing the production deployment; deployments without
  a phone API (loopback dev) skip verification and log a warning at
  boot.
- The CSRF cookie's `Secure` flag is now derived from the configured
  trusted origins: any `https://` origin marks the cookie `Secure`, so
  it can never ride a plaintext hop in a TLS-fronted deployment
  (`TestCSRFSecureFollowsTrustedOrigins` pins the rule).

### Added

- NixOS module: `services.webphone.nginx.hsts.{enable,maxAge}` opt-in
  Strict-Transport-Security on the generated vhost (default off, with
  the flake check asserting the header appears when enabled).
- Every server-rendered tab can now reach the phone in one click:
  History rows dial the caller (inbound legs) or the dialled
  destination (outbound legs), Voicemail rows call back the sender,
  and an open Messages thread offers a call to the remote number.
  Rows without a dialable number render no button.
- `data-dial` buttons no longer dead-end silently when the phone is
  signed out: the shell focuses the login field and explains via a
  toast (same toast markup and CSS as the island's).
- The shell header shows a live-call presence badge ("on call · N",
  pulsing) driven by the island's existing `wp:calls-changed` event,
  so every tab reflects that the phone is busy.
- Personal contacts have one home: new session-gated JSON endpoints
  (`GET`/`POST`/`DELETE /api/contacts`, extension-scoped) expose the
  same store the Contacts tab renders from. The island's contact
  panel reads and writes the server store and migrates its legacy
  `localStorage` list once after login — cleared only after the
  server accepted every row; failed imports retry on the next login
  (the store upserts by number, so re-import cannot duplicate).

### Fixed

- A hung backing resource no longer hangs the prober: every
  `/healthz` named check now runs under a 2s deadline and a timeout
  degrades the probe to 503 naming the check (`sqlite: timed out`),
  while the underlying call keeps running.
- A transient failure adopting the rotated CSRF token after login now
  retries with backoff (3 attempts) before falling back to the page
  reload, so a blip no longer costs the fresh SIP registration.
- Test harness: two HTTP clients in one server test no longer share
  a cookie jar — `httptest.Server.Client()` returns one cached
  `*http.Client`, and setting a Jar on it hijacked every client the
  test had built earlier (surfaced as CSRF 403s).
- Client-supplied identifiers (path segments, query parameters) that
  are not valid ids now answer `404` instead of panicking the handler:
  the domain gained `ParseThreadID`/`ParseFaxID`/`ParseAttachmentID`/
  `ParseContactID`/`ParseMessageID`, and the handlers use them; the
  `Must*` forms stay reserved for database rows, where a malformed id
  means corruption.
- The `webphone-module` flake check's HSTS variant evaluated the extra
  module config at the wrong nesting level, failing every
  `nix flake check` since the HSTS option landed; corrected, the check
  runs green.

## [2.1.0] - 2026-09-19

### Added

- NixOS module: `services.webphone.memoryMax` option wiring systemd
  `MemoryMax` (default null = uncapped), an `/events` SSE location in
  the module's own nginx vhost (proxy buffering off, HTTP/1.1, 3600s
  read timeout), and `recommendedProxySettings` on the websocket
  location.
- `nixosModules.webphone` alias next to `nixosModules.default`.
- The `webphone-module` flake check now asserts the generated vhost
  locations (`/`, the websocket path, `/events`) and the `webphone`
  systemd unit in the evaluated config, not just the rendered config
  JSON.
- Island lint gate: `checks.island-lint` runs oxlint (fail-closed,
  `no-undef` error) over the island modules and `shell.js` on every
  `nix flake check` — the class of silent ReferenceError behind the
  accept/reject bug can no longer ship.
- Contract-pinning tests: session gates live only in the shared helper
  (401-writer scan), provider multipart field order, verbatim-served
  reload asset, and htmx-config-meta-before-script ordering.
- In-repo live smoke suite (`scripts/webphone-smoke.py`, 21 checks,
  stdlib-only): builds the binary, boots it on a temp data dir and
  probes shell/CSRF/session/SSE/webhook/phone-api end to end.
- `vulnix` flake app scanning the RUNTIME closure (`--closure`), so the
  honest advisory count (8 derivations, currently zero real findings)
  is one command instead of tribal knowledge.
- Transcript paging state (`data-page`) survives SSE `thread` pushes
  (cancelable `htmx:sseBeforeMessage` guard), and a live transcript swap
  marks the thread read client-side and refreshes the nav badges via the
  new `GET /partials/nav` + `POST /messages/{id}/read` endpoints.
- Nav labels follow the extension language switch without a full reload
  (`wp:lang-changed`).
- Request-ID correlation on every request: the cqrs-htmx enrichment
  middleware now sits outermost in the server chain, so each response
  carries an `X-Request-ID` header and each request-log line records the
  identical `request_id` — a response can be matched to its 3 a.m. log
  line.
- A calibrated `Permissions-Policy` header (`microphone=(self)` for the
  WebRTC phone; camera, display-capture, geolocation, payment and usb
  denied). The library's recommended policy stays rejected because it
  denies the microphone outright.
- Webhook 5xx responses are redacted: internal error detail (store paths,
  SQL state) now goes only to the server log (`webhook apply failed`);
  providers receive the library's redacted SafeDetail text.

### Changed

- NixOS module: a `dataDir` outside `/var/lib/` fails evaluation with
  an explanatory assertion (systemd StateDirectory is derived from it).
- devShell exports `GOEXPERIMENT=jsonv2` + `GOTOOLCHAIN=local`, so
  bare `go` commands work inside `nix develop`.
- Flake style: `lib.*` (flake-parts perSystem lib) instead of
  `pkgs.lib.*`, no `with pkgs;` in the devShell, and the package `src`
  narrowed to `cmd/`, `internal/`, `go.mod`, `go.sum` via
  `lib.fileset` (smaller store source).
- `/version` reports the build version injected by the flake via
  ldflags (previously `(devel)` outside `go install` contexts).
- Frontend CSRF wiring is JSON-valid by construction: the shell's
  `hx-headers` attribute is built with `templ.JSONString` instead of
  string concatenation (a regression test parses the rendered attribute).
- CSRF token rotation on login and logout (`InvalidateCSRFCookie`, the
  nosurf fixation defense): the island adopts the fresh masked token
  without a reload via the new `GET /api/csrf` endpoint (`session.js`
  updates the meta tag and the body `hx-headers`; every token consumer
  reads live). Adoption failure falls back to a reload, which the
  post-login server session survives.
- The `Server-Timing` debug header is produced by the httputil
  servertiming middleware instead of a hand-rolled response writer
  (~40 lines deleted), keeping the same `WEBPHONE_DEBUG_TIMING` opt-in
  gate while adding CRLF sanitization and an SSE-safe writer.
- The toast wire-shape struct is now a type alias of the library's
  `cqrshtmx.ToastDetail`, so an upstream shape change fails this build
  instead of silently breaking the island's toast listener.

- `scripts/release.sh`: the release runbook (preconditions, fold
  check, version bump, gates, tag/push/verify, link check, stack
  relock + gates, aarch64 cross-builds, GitHub release) as one
  fail-fast command with `--dry-run`.
- `scripts/webphone-smoke.py --base` now probes foreign servers
  honestly: secret/injection-dependent checks are skipped with a
  stated reason, and bogus login credentials must be REJECTED (a 201
  flags a pre-v2.1.1 build that still mints sessions without
  verification).

### Fixed

- CSRF rejected every browser login behind the TLS-terminating proxy:
  a truthful browser POST arrives with `Origin: https://host` +
  `Sec-Fetch-Site: same-origin`, which the plain-HTTP listener read as
  a forged same-origin attestation and answered 403 — tabs never
  unlocked in any fronted deployment (v2.0.0 shipped this). The new
  `csrf.trusted_proxies` / `csrf.trusted_origins` config teaches the
  middleware the fronting shape; the NixOS module sets both when its
  vhost is enabled, and a subtest pins the exact fronted request.
- `hookFaxStatus` error mapping: empty `provider_ref` is now 400 (was
  404), unknown ref 404, other store failures 500 — parity with the
  message status hook.
- `messages.provider_ref` carries the same UNIQUE partial index as
  `fax_jobs`, closing the duplicate-receipt window on the messages side.
- Delivered messages render the distinct `wp-status-delivered` badge
  instead of reusing the sent style.
- Session→PBX credential construction is centralized
  (`Session.PBXCredentials()` + `SignInFirst`), removing four inline
  copies in the server handlers.
- Unreachable send gateways (message/fax) now answer 502 with a
  localized "saved as failed" note instead of 422 with the gateway's
  internal error text; validation mistakes keep their 422 reasons.

## [2.0.0] - 2026-09-19

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

- The static-site derivation is gone: consumers run the service binary
  and put TLS + the WSS `/sip` proxy in front (see README deployment).
- `window.PBX_CONFIG` is rendered by the server at `/config.js` from its
  own configuration instead of being supplied by the serving PBX.
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
- The unread badge count is cached per extension (5s TTL, invalidated
  on inbound/status/deletes) instead of rescanning every thread on each
  shell render.

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
