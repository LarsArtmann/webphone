# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Contacts API hardening (plan T12): `POST /api/contacts` is rate
  limited (60/min, burst 60 — sized for legacy imports, which POST one
  row at a time), the per-extension store enforces an atomic cap of
  500 contacts (renames never consume slots; a full list answers 422),
  and the OpenAPI 3.1 spec now documents `GET`/`POST`/`DELETE
  /api/contacts` (a spec-vs-handler test pins the three paths).
- Live contacts nudge (plan T13): every contacts mutation — tab
  save/delete/import AND island API writes — publishes a payload-less
  `contacts` SSE event; the open contacts tab re-fetches its own panel
  (morph swap, stable input ids), so island edits appear without a
  reload. The last stale live surface goes live.
- Transient rebuild pill (plan T17): while the connection watchdog
  rebuilds a lost registration the pill shows an explicit
  "rebuilding registration…" state (en/de) instead of a stale
  "connected"; island tests pin the re-entrancy collapse guard and the
  transient state, and the CSRF adoption retry ladder (recover on
  retry 2, reload only after 3 failures) is pinned by island tests
  (T14; rotation-on-slide itself stays a documented NOT-DO — verdict
  in `docs/planning/2026-09-22_14-45_csrf-rotation-on-slide-verdict.md`).
- `scripts/webphone-smoke.py --expect-version X.Y.Z`: asserts the
  running server's `/version` (verified positive and negative) — the
  deploy-verification companion.

### Changed

- `GOEXPERIMENT=jsonv2` is gone everywhere (devShell, buildGoModule,
  buildflow env, release.sh, smoke, README/CONTRIBUTING/AGENTS): Go
  1.27.1 ships a stable `encoding/json/v2`, the flag had become a
  footgun (it contributed to a failed release run outside the
  devShell), and the full suite is green 13/13 without it.

## [2.5.0] - 2026-09-22

### Added (2026-09-22 error-excellence train)

- Typed error families at the two outbound seams (plan:
  `docs/planning/2026-09-22_01-20_SUPERB-error-excellence.md`):
  gateway transport/receipt failures classify as Transient, form-build
  and store failures as Infrastructure, validation and provider 4xx
  answers as Rejection (`github.com/larsartmann/go-error-family` —
  already an indirect dependency, promoted to direct at the same
  version; zero new dependencies). The message and fax handlers now
  share ONE failure ladder (`sendFailure` + `classifyForUser`) instead
  of two duplicated type-ladders; every rendered string and status is
  byte-identical to before (pinned by tests) and the family rides the
  log line (`family=rejection|transient|infrastructure`).
- Boundary observability (logs only, zero behavior change): mid-stream
  response-write failures (fax/attachment/vCard streams, phone-api
  proxy, error banner, openapi) and contact-import skips (count +
  first reason in one line) are now visible in the journal instead of
  silently discarded.
- The erraudit bar is defined (AGENTS): the enforced set
  (`--type-aware`) gates green; `--enforce-go-error-family` becomes
  family-adoption tracking (102 stdlib-constructor sites remain
  outside the seams — must shrink, never grow); oops/generic-return
  stay owner-audit-only.

### Changed (same train)

- Removed the dead `messaging.ErrThreadNotFound` sentinel (zero
  producers, zero consumers — thread absence already flows through
  `store.ErrNotFound`).

### Added

- Sign-in once per device, not once per visit: the island now RESUMES a
  live cookie session at boot (`GET /api/session` hands the session's
  SIP credentials back, the browser re-registers silently, the login
  form only appears when there is genuinely no live session — first
  visit, logout, expired, or credentials that no longer register, in
  which case the stale row is dropped with an explanatory hint in
  en/de). Sessions slide while in use: activity past the halfway point
  of the idle window (`session_ttl`, now `7d` by default) renews the
  session and re-issues the cookie with the server's remaining
  lifetime, so a regularly used device never re-signs-in. The new
  `session_max_ttl` (`30d`) is the absolute cap from sign-in that
  expires even a continuously renewed (or stolen) session. This
  deliberately widens the idle window the session-persistence spike
  bounded at 24h (7d idle + 30d absolute vs the old flat 24h) — the
  trade is documented on both config keys, and both stay operator-settable.

### Fixed

- A registration lost AFTER it was established no longer wedges the
  phone silently: the island's connection watchdog now rebuilds the
  whole agent (fresh Registerer) when a Registerer that had reached
  `Registered` later goes `Unregistered`/`Terminated` outside logout or
  a rebuild (a rejected re-REGISTER after a transport reconnect, a
  server-side contact drop — the never-root-caused 1001 E2E anomaly
  where sofia said `user_not_registered` while the island kept its dead
  Registerer forever). The reconnect retry path also rebuilds on a
  Terminated registerer instead of retrying `register()` on a dead
  object, a registration lost during a transport OUTAGE stays on the
  backoff (no rebuild into a dead network), and an INDEPENDENT
  cycle-deadline timer (15s) force-rebuilds if a reconnect cycle wedges
  without settling. A bogus-credentials LOGIN still shows the
  `registration rejected` pill with no rebuild loop (E2E contract
  preserved). Pinned by `island-tests/connection.test.mjs` (nine
  scenarios over a SIP.js stub: happy path, lost/terminated-after-
  registered rebuilds, bogus-login pill, Terminated-registerer retry
  rebuild, logout inertness, outage gate, cycle deadline, stale pill).
- The reconnect pill no longer lies after a successful recovery:
  sip.js fires NO `stateChange` when a Registerer re-registers without
  having left `Registered` (a transport loss does not demote it), so
  the pill kept showing the last backoff state ("reconnecting in 4s
  (try 2)") while the phone was fully re-registered — the 2026-09-22
  E2E failure chain started exactly there (the suite read the stale
  pill as "stuck", fell back to reloads, and the reloaded pages
  resumed without a login click). The reconnect success path now sets
  the registered pill explicitly (plus the preserved-sessions note).
  Same class as the watchdog's bounded rebuild: recovery must never
  depend on a state event that only fires on a transition.
- Provider rejections no longer masquerade as "The message gateway is
  unreachable": a non-2xx gateway ANSWER is now a typed
  `gateway.ErrProviderRejected` whose detail (unwrapped from the
  `{"error": "…"}` envelope) is shown in the panel, for messages and fax
  alike. Transport failures keep the generic banner + log detail (they can
  carry internal URLs). 2026-09-21 production burn: sending an SMS to the
  PBX's own DID is structurally refused by Telnyx (400 "Source and
  destination cannot be the same number") and surfaced as "gateway
  unreachable" — a working system telling the user it is broken.
- Avatars no longer emit inline `style` attributes: the deterministic
  per-peer hue rides a `wp-av-h<deg>` class (10° buckets, 36 rules in
  app.css) instead. The strict `style-src 'self'` CSP silently blocked
  the old `style="--av-h: …"` attribute on every contacts/messages
  render, so avatar tints never applied in production browsers (console:
  "Applying inline style violates …"). Style attributes cannot be
  hash-allowlisted per-value, so class bucketing is the only CSP-strict
  shape.
- No more 401 storm on cold login: the voicemail badge and server
  history refreshes now run on `wp:session-opened` (after the session
  cookie is minted and the fresh CSRF adopted) instead of racing the
  un-awaited session POST at login-submit. A signed-out shell
  (welcome-hint tab area) additionally auto-loads the default
  messages tab + nav once the session opens, so tabs work without a
  manual reload after the island login.

### Added

- SQLite-backed session store: tab sessions now survive service
  restarts (the pre-2.0 "every deploy signs everyone out" failure class
  is deleted at the root, not narrated). Same cookie, same TTL
  semantics, fail-closed gates unchanged; the in-memory store remains
  for tests and the design/threat review lives in
  `docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md`
  (`TestSQLiteSessionStoreSurvivesRestart` + a kill -9 smoke scenario
  pin it).
- Durable inline errors for failed tab actions: the shell ships an htmx
  `responseHandling` override that swaps the server's `.wp-error` panel
  banner into a dedicated `#wp-tab-error` slot (rendered outside
  `#tab-content` so tab swaps and compose drafts are never touched);
  401 stays swap-free by exception. The toast-plus-inline pairing on
  502 gateway outages is test-pinned.
- Completed toast feedback map: server-session login failures toast
  status-specific advice in the user's language (credentials, 429,
  other HTTP, unreachable), the shell names 429 as
  client-correctable instead of a generic failure, and a dead SSE feed
  toasts once after three consecutive failures (recovery resets the
  counter). New copy in both en/de dictionaries.
- Accessible, keyboard-dismissable toasts: the `#toasts` live region
  (`role="status" aria-live="polite"`) is now pinned on the served page,
  and toasts take focus on Tab and dismiss with Enter/Space/Escape in
  both the island and the shell.
- Island tests for the server-session feedback map (login failure
  classes, quiet-success, SSE failure counter) — session.js was the
  last untested island module with user-visible error paths.
- Smoke suite restart scenario: login → `kill -9` → reboot on the same
  data dir → the old session cookie still opens session-gated surfaces
  while anonymous requests stay rejected.
- `checks.webphone-backup-drill`: the backup/restore drill (tar →
  restore → byte-identical attachment retrieval) now runs as a
  sandbox-safe flake check, proving the RESTORE path, not just the
  online snapshot the kvm-gated VM test covers.
- `scripts/release-hygiene.sh`: the post-train sweeps (fanout
  benchmark when the SSE libraries moved, vulnix, lychee, origin/main
  vs HEAD) as one command with a failing exit status.
- Depth suites: store owner-scoping table (every owner-taking query
  pinned against a foreign extension), webhook error-branch table
  (transport/timeout/5xx-with-truncated-detail/JSON-ish/oversized
  receipts), dialable + branded-id parser edge table (pasted RTL
  numbers, charset, boundary lengths), and a fuzz target over the
  contacts JSON API (malformed bodies are always 400/422/204).
- Shell load-error boundary: a throw during shell wiring leaves a
  breadcrumb in `#log` + console instead of a silently half-wired
  page; the error-toast throttle reads an injectable clock
  (`window.__wpClock`) so tests stop monkey-patching `Date.now`.
- Fax compose shows an upload indicator on the submit button while the
  PDF posts; `prefers-reduced-motion` now kills all island animation/
  transitions (parity with app.css).
- Toolchain self-heal: `scripts/webphone-smoke.py` and
  `scripts/buildflow.sh` detect the host go-below-floor trap
  (`GOTOOLCHAIN=local` with an old go) and re-exec inside
  `nix develop -c`; the buildflow wrapper also promotes the gitleaks
  and codespell scans into the default run (`.codespellrc` silences
  the deliberate German copy, taking codespell to zero real findings;
  `.bandit` documents the drill-script exclusion).

### Changed

- Island JS test runner (the standing gap is closed): `node:test` with
  minimal DOM stubs under Nix — tests live in `internal/web/assets/
island-tests/` (a sibling of the served tree, never embedded) and run
  as the `island-js` flake check. First ports replace grep-only
  tripwires with real behavior tests: toast rendering (kind, stack cap)
  and island i18n en/de key parity (mirroring the Go-side sync test).
- `checks.vulnix-triage`: the vulnix triage verdict logic is now the
  `webphone-vulnix-triage` CLI (shared by `nix run .#vulnix` and the
  fixture check), so the grep-in-pipeline verdict-inversion bug class
  from 2026-09-20 fails at check time instead of at the next train.
- `scripts/release.sh` step 8 asserts the aarch64 cross-build's ELF
  machine bytes (`b700` = EM_AARCH64): `--system` is a restricted nix
  setting an untrusted client's daemon may silently ignore while
  exiting 0 — the false-green cross-build gate now fails loudly.
- Own-number identity (config `identities`): map an extension to its
  presented PSTN number (DID) and the UI finally answers "what is my
  number" — the signed-in header shows `extension · DID`, the island's
  whoami line gains the DID from the session response, and the messages
  and fax composers show the sending identity. Display-only and
  session-scoped (never rendered into the unauthenticated `/config.js`);
  unmapped extensions keep today's extension-only display
  (`TestIdentitySurfacesOwnNumber` pins all surfaces).
- NixOS module: dedicated nginx locations for the probe triple
  (`/healthz`, `/livez`, `/startupz`) instead of riding `/` — a fleet
  health hub can be scraped from or fenced (`allow`/`deny` via
  `extraConfig`) per location without touching the app's location; the
  `webphone-module` flake check asserts all six vhost locations.
- NixOS module: typed `services.webphone.csrf.trustedProxies` /
  `trustedOrigins` options rendering into `settings.csrf` (empty lists
  preserve the `nginx.enable` defaults; non-empty lists override them)
  and `services.webphone.serverTiming.enable` setting the
  `WEBPHONE_DEBUG_TIMING` env gate — both pinned by new flake-check
  assertions.
- Backup-timer NixOS VM test (`checks.x86_64-linux.webphone-backup`,
  kvm-gated): boots the real service with `backup.enable`, runs the
  oneshot, and asserts the snapshot pair lands under `destDir`, the
  copied database passes `pragma integrity_check`, the timer renders
  `OnCalendar`/`Unit`, and the service never restarted (the online
  claim). Fax-feed-test style from the consuming stack.
- Styled 404 page: unknown paths render the app shell around the error
  panel (reload affordance included) instead of Go's bare
  `404 page not found` text — error-page parity with the pre-2.0 static
  site, re-verified and restored after the templ-components adoption
  (`TestNotFoundRendersTheShell` pins it; the status stays 404).
- Smoke suite: `/partials/nav` anonymous-vs-signed-in shape checks —
  labels render anonymously and badges never do; the signed-in re-fetch
  shows the unread badge in its honest window before the thread opens.

### Changed

- The styled 404 message is translated: `error.notfound` joins the
  en/de dictionaries ("There is nothing at this address." /
  "Unter dieser Adresse gibt es nichts."), ending the one hardcoded
  string on an otherwise i18n'd error panel (both languages
  test-pinned; smoke covers the English side).
- NixOS module csrf precedence under conflict is now deterministic and
  pinned: typed `csrf.trustedProxies/trustedOrigins` WIN over raw
  `settings.csrf.trusted_*` values (previously both set = Nix's generic
  duplicate-definition eval error); raw still beats the nginx-derived
  defaults, so the escape hatch stands when typed options are empty.
  Pinned by a new `csrf-conflict-precedence` case in the
  `webphone-module` check.
- cqrs-htmx v4.11.0 (from v4.9.0, with the idiomorph adoption): the one
  wire change is the SSE stream — `/events` now opens with the
  library's `retry:` reconnect hint before the `connected` frame
  (`TestSSEStreamCarriesConnectedThenEvents` pins it; the htmx sse
  extension consumes it). MD1's bump-trigger obligations closed
  2026-09-20: the hub fan-out benchmark re-ran clean (baseline doc
  updated) and the stack browser E2E passed on the bumped tree.
- The dial alphabet now keeps letters end to end: the island's
  dial/transfer/contact sanitize regex matches the server's
  `sanitizeDialable` exactly (`[^\d+*#a-zA-Z]`), so alphanumeric SIP
  user parts survive pasting and saving instead of being silently
  stripped client-side (DECIDED 2026-09-20; both sides pinned —
  served-asset grep + `TestParsePhoneSanitizesLikeTheIsland`).
- The two htmx extensions now load as one bundle: `/htmx-ext.js` serves
  sse + idiomorph concatenated via `cqrshtmx.HTMXExtensionsHandler`
  (composite ETag, per-extension version comments) — one script tag and
  one request instead of two.
- Voicemail rows and their `<audio>` elements carry stable uuid-derived
  ids (`vm-<uuid>`, `vm-audio-<uuid>`): idiomorph persists any element
  whose id exists in both trees, so a voicemail re-fetch morph can
  reorder rows without re-creating an audio element mid-playback (the
  im-preserve audit's one stateful morph surface;
  `TestVoicemailRowsCarryStableMorphIds` pins the ids).
- `scripts/release.sh`: the gates now include `nix run .#vulnix`
  (runtime-closure advisory scan every train — cadence institutionalized
  instead of remembered).

### Fixed

- Silent dead-tab-session failures: after a server restart (or any 401 /
  network error) htmx tab clicks and form submits failed invisibly —
  htmx swaps nothing on error responses and nothing listened. The shell
  now toasts throttled, honest feedback on `htmx:responseError` /
  `htmx:sendError` (401 wording reassures that calls keep working;
  never an auto-reload, the island must survive) and skips responses
  that already carry `HX-Trigger` server feedback.
- Server-authored toasts rendered in `info` styling: the island's kind
  map spoke the dispatch-layer vocabulary (`success`/`warning`) while
  the server emits island kinds (`ok`/`error`), so every error toast
  looked neutral. `toastKindFor` accepts both vocabularies and is
  pinned by island tests.
- `scripts/release.sh` release-notes extraction: the section-matching
  awk treated `## [2.4.0]` as a regex bracket expression and never
  matched, shipping v2.3.0 and v2.4.0 with empty GitHub release bodies.
  The brackets are now escaped and both release objects were backfilled
  from their CHANGELOG sections (audit 2026-09-20: v2.1.0/v2.2.0 were
  already byte-identical to the CHANGELOG).

### Documentation

- README: module-options table (including the new `csrf.*`,
  `serverTiming`, backup options), the probe-triple +
  fleet-scraping paragraph, the backup drill invocation line, and an
  explicit off-machine restic/borg pointer (`destDir` shares the host
  with the service — it is not a backup until copied off).

## [2.4.0] - 2026-09-20

### Fixed

- The NixOS module satisfies statix's repeated-key gate: the three
  `systemd` assignments (service, backup service, backup timer) are
  merged into one attrset — eval-equivalent, no behavior change.

### Added

- Live-update surfaces now morph-swap instead of innerHTML-replacing:
  thread list, message transcript, fax list, voicemail panel and the
  nav badge refresh carry `hx-swap="morph:innerHTML"` and are
  reconciled by idiomorph (the self-contained extension bundled with
  cqrs-htmx v4.11.0, served same-origin at `/htmx-ext/idiomorph.js`
  — no new dependency). Matched DOM nodes are preserved in place, so
  focus, draft text, paging state (`data-page`/`data-thread`) and
  shell.js listeners survive live SSE pushes. The stack browser E2E
  passed against the branch before merging (148 s, full
  call/transfer/DTMF/reconnect flow).
- Backup story, drill-verified: the NixOS module gains
  `services.webphone.backup.{enable,destDir,calendar}` — a daily
  online-backup timer (sqlite `.backup` + blob-tree rsync, no phone
  downtime) — and `scripts/webphone-backup-drill.py` proves the cold
  restore path end to end: a real binary, an inbound webhook message
  with a binary attachment, tar the data dir, restore to scratch,
  reboot and pull the attachment back byte-identical. The README
  documents inventory, the online snapshot commands and the restore
  steps.

## [2.3.0] - 2026-09-19

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
- The `nanoid` dependency is bumped to v1.65.1 (its Go >= 1.27 floor
  was the watch's blocker; the transitive prng/aes-ctr-drbg siblings
  came along), unblocked by this release's toolchain floor.
- The version-drift guard learned the release window: a flake version
  ahead of the newest tag passes only while a dated CHANGELOG section
  for it exists (the runbook's fold step), so the guard no longer
  rejects the runbook itself while still failing real drift.
- The smoke suite grew to 28 checks: `/livez` and `/startupz`
  assertions keep the new probe pair honest.

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

[Unreleased]: https://github.com/LarsArtmann/webphone/compare/v2.4.0...HEAD
[2.4.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.4.0
[2.3.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.3.0
[2.2.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.2.0
[2.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.1.0
[2.0.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.0.0
[0.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v0.1.0
