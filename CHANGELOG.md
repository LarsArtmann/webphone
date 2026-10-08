# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Live-transcription resilience on both client surfaces: the shell's
  auto-start transcription SERIALIZES rendered clips through one promise
  chain (a dozen swapped-in rows no longer stampede the seam's flood
  budget into 429s) and a 429 is retried exactly once after the
  `Retry-After` delay (bounded to 10 s — a hostile header cannot freeze
  the tab); the island's live capture loop honors a 429 with a 10 s
  segment cooldown (capture keeps running, the budget recovers, no
  retry storm).
- Dropped `/api/transcripts` saves no longer lose segments: each call
  carries an in-memory save queue (order-preserving, capped at 50
  segments per call, drop-oldest) that flushes one-at-a-time on the
  next successful save and once more when the call ends. A failed save
  warns once per call and never blocks the live card.
- Google Cloud Speech-to-Text V2 as a second ASR provider kind
  (`asr.provider: google` + `asr.project`/`asr.location`/`asr.language`;
  `asr.token` becomes the `x-goog-api-key`). The seam
  (`internal/asr.Provider`) now has two implementations — the
  OpenAI-compatible wire (Speaches/whisper.cpp) and the Google JSON
  recognize wire (base64 audio, `telephony` model, regional
  `europe-west3` endpoint so caller audio stays in the EU; V2 bills per
  second with no per-request minimum). The composition root picks the
  kind per config; `/api/transcribe`, `/config.js` and every UI gate are
  provider-agnostic.
- Passkey (WebAuthn) login mode, config-gated and DEFAULT OFF (a
  deployment without `auth.passkey.*` keeps the byte-identical
  extension login card): an embedded cqrs-htmx/usermgmt v4 identity
  layer (`internal/userauth`, own `usermgmt.db` under the data dir)
  backs an email-first front door; the finish ceremony resolves the
  configured email→extension mapping, sources the SIP directory
  password from operator-managed runtime files (read per login, never
  cached, fail-closed on empty/missing), directory-verifies it and
  mints the webphone session (the password returns to the island
  no-store so it can REGISTER; the whoami line leads with the display
  name and the user's numbers). The extension login stays as a
  break-glass `<details>`; sessions resume with the same identity
  fields. Enrollment is CLI-minted one-time tokens
  (`webphone -enroll-passkey <email>` — registers idempotently, prints
  a 15-minute single-use link; `GET /enroll` runs the browser ceremony
  on a standalone CSP-clean page). New island modules `webauthn.js`
  (the one base64url↔ArrayBuffer wire coercion, node:test-pinned) and
  `passkey.js`; `session.js` extracts `adoptServerSession` as the one
  post-session sequence both login paths share. Uniform anti-enumeration
  401 on begin/finish (no account oracle); userauth.* error codes ride
  the error-family registry.
- Live transcription of every audio stream (calls, voicemail, MMS audio),
  config-gated and DEFAULT OFF via the new `asr.url`/`asr.token`/`asr.model`
  seam (a URL alone enables; a token without a URL fails closed). One
  OpenAI-compatible endpoint (`POST /api/transcribe`, session-gated,
  CSRF-protected, flood-budgeted) forwards a whole audio segment to a
  self-hosted whisper.cpp / OpenAI-compatible provider (`internal/asr`,
  error-family pinned) and returns the text; a disabled seam answers 404.
  The island captures live-call segments (remote + operator audio mixed
  through an AudioContext that now CLOSES on stop, `MediaRecorder` every
  4 s) and streams them into the call card; the Voicemail and Messages
  tabs gain transcribe buttons through one delegated shell.js handler.
  New island module `transcribe.js`; `window.PBX_CONFIG.asr` gates every
  affordance (server-rendered buttons included); the Settings tab reports
  the seam's state.
- Transcription auto-start and persistence: with the seam on, live capture
  begins when a call connects (pausing on hold, resuming after), voicemail
  and MMS audio transcribe themselves once per page life (morph re-renders
  never re-POST), and every live-call segment is appended to an
  owner-scoped `call_transcripts` store (schema v3, 5000-segment cap per
  extension) that the History tab renders as its own "Call transcripts"
  section. Segments carry the UI language as the provider hint; transcript
  deltas ride the polite live region; a failed save warns once per call
  and never blocks the capture.
- Error-code registry (error-contract.md): a 104-code table of every
  `<seam>.<op>` error code, generated from the source and freshness-pinned
  by `TestErrorCodeRegistryIsFresh` — a code rename now fails the suite
  instead of silently breaking journal greps. The page also tells the
  "error families are total" story (families, principles P1–P7, the
  `[family:code]` log vocabulary, and the two runtime-split codes).
- Tooling: the health.css rebuild script now runs the treefmt formatter on
  its artifact (a raw rebuild used to re-break the `format` gate), and the
  smoke suite is mypy- and ruff-clean (the 8 tuple-shape warnings were
  lying annotations, now truthful).
- Thread organization (M21/M22, T18): conversations can be pinned,
  muted, and archived — three 0/1 columns behind a VERSIONED migration
  (schema_version v2; migrations now run as ordered steps, never ad-hoc
  ALTERs). Row toggles answer 204 + a list-only SSE nudge; the open
  thread's head buttons re-render in place (one handler branches on the
  HX-Target header, so the transcript, scroll position, and draft stay
  untouched while the buttons flip to the network-confirmed state).
  Inbound messages auto-unarchive their thread; mute hides the unread
  badge without touching stored truth. Thread rows carry stable wrapper
  ids so live morph updates never rebuild a row mid-click. Design
  decisions D1–D14: docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md.
- Reply snippets (M21, T18): a per-extension snippet store (cap 100)
  managed from Settings; up to five "quick" snippets render as chips
  above the reply composer with every snippet also behind a picker.
  Inserting replaces the draft and focuses it. Image attachments open
  in a lightbox dialog. Scheduled sends were cut (NO-GO on a gateway
  seam with no deferred-send concept — a scheduled UI would promise a
  state the transport cannot honor).
- Call island feedback and trust (M17, T15): aria-live politeness
  everywhere except the assertive incoming-call announce; a failed
  optimistic send grows Retry/Dismiss buttons inside the failed bubble
  with draft restore; SSE recovery announces once, only after an
  announced drop; history's empty state is filter-aware.
- Incoming-call focus mode + device self-test (A6/A10, T19): an
  incoming call dims the chrome and focuses Accept (typing in an input
  never loses focus); the advanced panel gains a pre-call device check
  (mic label, immediate release, speaker count, honest degradation).
- Voicemail playback (M13, T14): inline player with scrubber, speed
  control, and elapsed-time readout. Fax depth (M14, T14): per-job
  status timeline and resend.
- Island performance (T11): modulepreload for the island module graph,
  mic pre-warm timing baseline (scripts/perf-baseline.py), and an ICE
  gathering-time panel in the advanced area.
- Passkey health leg: with the mode on, `/healthz` gains a `userauth`
  named check pinging the identity layer's own `usermgmt.db` (a 503
  naming `userauth` replaces unexplained passkey 503s when that database
  breaks); the go-health probe picks the service up NON-critical (a
  broken identity DB degrades one login mode — it must not flap liveness
  or hold the startup latch), so the dashboard shows it as warn-state.
  Mode-off deployments keep probing exactly sqlite + blob-dir.
- Tier-2 pins for the passkey train: config family pins for every
  `config.auth.passkey.*` rejection code (in-package table + an
  env-driven Load row), enroll begin/finish handler tests (ceremony walk
  + honest 400s), and the `userauth.Shutdown` lifecycle test (closes
  `usermgmt.db`; the health check fails afterwards, never silently
  passes).
- README § "Passkey sign-in (optional)": the one home for the
  enrollment/token/wire/health detail (moved out of AGENTS.md), next to
  the CRM and Paperless integration contracts.
- Dial affordances (M9, T12) + SMS segment countdown (M12, T12):
  number-bearing rows offer one-tap dial into the island; the composer
  counts billable GSM-7/UCS-2 segments past the first.
- Visual tokens, theming, and shell sizing (M15/M20/M26, T16): mirrored
  light/dark token blocks across app and island styles, a three-state
  theme toggle persisted pre-paint (theme-preload.js), calmer shell
  sizing.
- Onboarding and mobile (M18/M19, T17): a dismissible welcome note for
  fresh sessions; fixed bottom tab bar and touch sizing on small
  screens.
- i18n/RTL groundwork (M24, T19): stylesheets swept to CSS logical
  properties (a guard test pins it); dead dictionary keys are now a
  test failure (the guard found and removed two on day one).
- Visual capture harness (T23): scripts/ui-capture.py boots a fresh
  binary, seeds deterministic content over HTTP, injects the session
  cookie into headless chromium, asserts each surface's DOM markers,
  and captures the 14-shot light/dark matrix into ui-shots/. Local-only
  by budget decision (see AGENTS § Commands).
- Operator boot-error contract: the operator is a user. Every boot
  failure (config, data dir, timezone, paperless, bind/listen, panic,
  or unclassified) now renders the five-part error contract (WHAT /
  REASSURE / WHY / FIX / ESCAPE) plus the underlying error to the
  journal, version-stamped and class-tagged, English-only, with a
  documented exit taxonomy (1 = designed failure, 2 = panic). The
  fail-fast panic mechanism at the composition root is unchanged;
  `App.Start`'s three panics became returned errors; guards: golden +
  enum-coverage + classification + EN-only pins (`bootreport_test.go`),
  an arch test confining samber/do to the composition root, and a new
  smoke `boot failure scenario`.
- Speak ASAP after accepting a call: the island now pre-warms the
  microphone while a call RINGS (new `mic.js` seam) and hands the warm
  stream to sip.js at accept time through a custom session-description
  media factory, so mic acquisition (up to seconds on Bluetooth
  headsets) no longer sits inside the answer path. The ICE gathering
  wait before the 200 OK is capped at 1 s (sip.js 0.21.2 default: 5 s).
  The mic indicator lights while ringing (deliberate tradeoff), and
  the device is released on reject, missed calls, and logout.
- Honest hold UI + loud offline banner in the call island: the hold
  button no longer flips to the settled state optimistically — a
  pending re-INVITE shows the WORK ("holding…" / "resuming…", pulsing
  chip, disabled button) and only a settled 200-OK shows "on hold" /
  "in call"; a failed toggle snaps the card back to the state that
  actually still holds. The phone view gains an `#offline-banner`
  (new DOM-contract id) driven by the same registration truth as the
  status pill — visible from login until the REGISTER lands, on any
  transport loss or rejected registration, and on the browser's own
  `offline` event (whose `online` counterpart nudges recovery without
  ever claiming registered on its own). Banner text rides data-i18n
  (en/de); new island i18n keys ship in both dictionaries.
- Optional Paperless-ngx archive for inbound faxes (`paperless.url` +
  `paperless.token`, both-or-neither like the CRM seam): each inbound
  fax is offered to a go-paperless-backed archiver fire-and-forget after
  persist + SSE notify — tagged `fax`, typed `Fax`, titled
  `Fax from <number> <date>`, provenance custom field carrying the
  webphone fax id; a content-hash duplicate refusal is inert. The blob
  store stays the only storage truth: a slow or dead Paperless never
  delays or fails a fax, failures WARN in the log, and unconfigured
  deployments behave exactly as before. Outbound faxes are v1
  out-of-scope (seam supports them; hook lands later).
- Service-oriented composition root: `internal/app` wires every service
  through a samber/do v2 container (promoted from indirect — it arrived
  via go-health). The critical pair (`sqlite`, `blob-dir`) registers as
  named services implementing the container's health-check interface,
  the go-health probe reads them from the injector, and `App.Shutdown`
  runs the container cascade (dashboard pusher drain, SQLite close)
  after the HTTP drain. Wire contracts are unchanged: `/healthz` keeps
  the cqrshtmx readiness shape, `/livez` and `/startupz` serve from
  the injected probe (a nil `Deps.Probe` falls back to the equivalent
  standalone probe, so test composition is untouched).
- Optional health dashboard (go-health-dashboard v0.10.1) at `/health`,
  disabled by default (`WEBPHONE_DASHBOARD__ENABLE=true` or
  `dashboard.enable`; `dashboard.title` names the deployment). Live
  check cards over SSE, status trend, JSON via content negotiation,
  and its own scoped Tailwind build at `/assets/health.css` (the app's
  `tw.css` stays untouched). CSP: per-request nonces + the dashboard's
  verified policy with `unsafe-eval` scoped to the subtree (the
  Datastar SDK compiles its expressions); every other surface keeps
  the stricter app-wide policy. The NixOS module's Caddy vhost proxies
  `/health/*` unbuffered like `/events`. The smoke suite boots with
  the dashboard enabled and covers the page, the CSP nonce, the
  same-origin SDK, and the probe aliases. Footprint: +1.1 MB unstripped
  (+5.1%), stripped release binary 15.25 → ~15.4 MB — far inside the
  ≤ +8 MB / ≤ +20 % gate that rejected the 2026-09-30 setup-shell
  adoption.
- Trust + accessibility batch (2026-10-01 UI/UX train, committed after the
  v2.8.0 tag): the reply composer shows an optimistic pending bubble that
  flips to failed and restores the draft on a send error; transcripts gained
  day separators (Today / Yesterday / weekday) with a group count and an
  "N unread" divider; navigation swaps move focus to the new panel heading
  and announce through a screen-reader live region while the nav marks the
  current tab with `aria-current`; a tab-loading skeleton and a short panel
  transition make swaps feel instant; a skip-to-content link and form
  `aria-describedby` aid keyboard/screen-reader users; the nav becomes a
  fixed bottom tab bar on phones with a sticky call-prominent island; and a
  Ctrl/Cmd-K command palette plus a "?" shortcut help (mirrored as a
  Settings cheat-sheet) expose the shell's actions. A contacts-manager
  workstream built in the same train was reverted — contacts depth is
  Ledger's domain, not this app's.

### Changed

- Dependency train (buildflow `go-mod-update`): templ-components
  v1.20.0 → v1.20.1 (v1.20.0 shipped an unresolvable
  `errorpage@00010101…` require that fails every module-graph
  resolution — vendor mode masked it), the go-cqrs-lite v4 submodule
  set to latest (storage v4.10.4, watermill v4.6.4, stack v4.4.3,
  snapshot v4.6.1, scheduling v4.6.1, …), go-flightrecorder v0.2.1,
  go-sse/sseparse v0.2.1; `vendorHash` re-pinned to match (and again
  after the post-train `go mod tidy` moved go.mod/go.sum — a tidy
  landing after the pin red-ed CI with a hash mismatch).

- Session stores (`internal/session`) take an injectable clock: the
  SQLite and mem stores read time through an unexported `now` field
  (default `time.Now`), so the TTL/sweep tests drive expiry with an
  explicit `Advance` instead of real sleeps — the "young session not
  live" CI flake (test lost a race against a loaded runner) is
  structurally impossible now. No behavior change.

- Server lifecycle: `cmd/webphone` now serves via `httputil.NewServer`
  (the primitive the rejected cqrs-htmx setup bundle's RunHandler
  wraps) — SSE-safe timeouts (ReadHeaderTimeout 5s bounds slowloris,
  IdleTimeout 60s reaps dead keep-alives, deliberately NO read/write
  deadlines so SSE streams outlive any fixed bound) and a 30s shutdown
  drain budget replacing the hand-rolled 10s one. This is the salvage
  of the 2026-09-30 setup-shell adoption NO-GO: adopting the bundle
  measured +10.40 MB (+68.2%) against the recorded ≤ +8 MB / ≤ +20%
  footprint gate — importing the setup package links its +72-module
  import graph even through the prunable constructor — so the bundle
  stays rejected (identity split-brain + footprint) and the lifecycle
  value rides the already-dependency at +8 KB. Verdict + measurement:
  `docs/planning/2026-09-30_10-37_SUPERB-setup-shell-adoption.html`.
- Nix flake polish (2026-10-01 nix-review train): the `webphone-module`
  check split into base/csrf/backup files, its NixOS stand-in gained a
  `freeformType` so a new config key can no longer break the check, and
  the `services.webphone.package` option moved to `mkPackageOption`
  (still required). Added a minimal `devShells.ci` (go + templ +
  golangci-lint) that the `go-tests` CI job now enters instead of the
  heavy interactive shell. The NixOS module was split under the
  ~300-line guideline — config stays in `nixos-module.nix`, with
  options, the Caddy vhost body, and the backup shell in their own
  files — output proven byte-identical by an eval diff of the generated
  vhost and both backup scripts.
- Island module homes (2026-10-04 tail): `csrf.js` is the CSRF token
  reader's ONE exported home (session.js, passkey.js and the standalone
  enroll page import it — three private copies were drift bait);
  `whoamiLine` moved from passkey.js to the neutral `ui.js` (both login
  paths render it); the session-identity wire shape (`GET /api/session`,
  login, passkey finish) is now the typed `sessionIdentity` struct
  instead of a `map[string]any` — omitempty tags keep the wire
  byte-compatible.
- AGENTS.md compacted 419 → 129 lines (the `docs/agents-md-size`
  preflight cap is 377): every rule kept, evidence parentheticals moved
  to docs/ (passkey detail → README), the userauth health leg, the
  csrf.js/whoamiLine homes and the python-not-jq lock-read guard folded
  in.
- Nix-review batch 2 (2026-10-02). The generated module output (Caddy
  vhost body + the retention=7 backup script + the settings JSON) is now
  pinned by a full-text golden (`nix/module-output.golden`, rendered by
  `nix/module-output.nix`) so the reordering/whitespace drift the
  substring checks miss fails `checks.webphone-module`. `release.sh`
  gained a release-time guard that the flake's `webphoneVersion` matches
  the newest tag — drift the eval can never see; override with
  `WEBPHONE_RELEASE_SKIP_VERSION_GUARD=1`. The two devShells now share
  one `goTools`/`goEnv` binding so the CI and interactive Go pins cannot
  drift. `deadnix` excludes vendored third-party `flake.nix` files. The
  scoped `/health` stylesheet gained a committed Tailwind input
  (`health.css.input`) recording the class-based dark variant (the
  missing line that silently dropped every `dark:` utility), a rebuild
  script, and a `checks.health-css` canary. Reply-snippet persistence
  gained error-family pins. The `go-standard` migration stays declined
  (this repo's bespoke flake checks are not modelled by it); the
  hardcoded `webphoneVersion` and the module-check stand-in's
  permissiveness remain accepted exceptions.
- Dependency sweep (2026-10-05, the 15:42 train): templ v0.3.1020 →
  v0.3.1070, templ-components v1.19.4 → v1.20.0 (with its icons/utils/
  datastar/htmx submodules), cqrs-htmx/usermgmt v4.13.1 → v4.14.0,
  go-health v0.4.1 → v0.5.0, indirect go-webauthn v0.18.1 → v0.18.2 and
  opentelemetry v1.46.0 → v1.47.0 (go-health-dashboard held at v0.10.2).
  The templ-components ride moved the button outline-warning/success
  variants from amber/green-600 text to -700, so the tw.css token remap
  gained matching -700 entries and the artifact was rebuilt — its
  utilities keep resolving through `--warn`/`--ok`, not the raw palette.
  vendorHash re-pinned; full suite and `nix flake check` green post-sweep.
- The error-code registry freshness pin (`TestErrorCodeRegistryIsFresh`)
  now compares table cell content with padding collapsed: the auto-commit
  daemon's markdown table re-alignment broke the byte-exact pin twice on
  2026-10-05 (main CI-red twice); column re-padding is inert now, while
  missing/stale/family-drifted codes still fail the suite by name.
- The same registry's `-update` WRITER now emits the daemon-aligned
  table shape (columns at max content width, plus the blank line the
  formatter keeps after the BEGIN marker), so regeneration is
  byte-idempotent — the daemon has nothing left to re-pad, killing
  the 2026-10-05 red-class churn outright rather than merely making
  it inert.
- Island module homes (2026-10-06 auth tail): `auth.js`'s `authedFetch`
  imports the `csrf.js` token reader instead of querying the meta tag
  itself — the last private copy of the CSRF read, closing the one-home
  rule the 10-04 tail established for session.js, passkey.js and the
  enroll page.

### Fixed

- Enrollment error copy (2026-10-04 tail): a network failure on the
  standalone `/enroll` page rendered "Enrollment failed (HTTP 0)" — the
  status-0 transport sentinel leaked to the user. It now maps to the
  purpose-written network message, and the (browser-unreachable, but
  honest) empty-token guard says to paste the token from the link
  instead of masquerading as an HTTP status.
- The adopted templ-components layer was visually inert: the island
  stylesheet's bare element rules (`button`, `input`, …) were unlayered,
  and ANY unlayered rule beats ANY Tailwind `@layer utilities` rule at
  any specificity — so Send/Call rendered as island-styled keys instead
  of Primary green everywhere the components reached. The island
  stylesheet now scopes its element rules to `.island` (box-sizing and
  the reduced-motion block deliberately stay global), the history filter
  button gets `wp-mini` (it would have gone UA-gray unclassed), and the
  tab region's snippet-picker summary keeps its focus ring. The capture
  harness waits on the island call view per shot instead of a fixed
  sleep (the resume reveal raced the old 0.4s).
- A resumed session was silently deleted whenever the SIP WebSocket was
  unreachable at page load: the boot-resume path treated ANY connect
  failure as "stored credentials are stale" and dropped the server
  session — kicking the user out of every tab (messages, fax, all of
  it) because the phone transport was briefly down. The island now
  distinguishes a refused REGISTER (stale credentials → drop + login)
  from a transport failure (keep the session, show the honest offline
  phone, let the reconnect backoff own recovery; tabs stay usable).
  Found by the T23 harness: the bare-boot island deleted its session
  within a second of every page load.
- Thread deep links rendered the wrong surface: `GET /messages/{id}` —
  the URL every thread row pushes to the address bar — always rendered
  the conversation LIST, so a refresh or shared link lost the open
  thread (caught by the T23 harness: the thread shots were byte-
  identical to the messages shots). The deep link now renders the open
  conversation; unresolvable ids degrade to the list with a
  gone-notice. Pinned by TestThreadDeepLinkOpensConversation.
- `config.js` hardcoded `wss://` for the SIP WebSocket, so a bare-HTTP
  dev boot could never connect the island; the scheme now follows the
  page protocol (HTTPS deployments unchanged).
- The committed `/assets/health.css` was stale: it carried zombie
  classes from earlier builds and missed the `--blur-xs` token. Rebuilt
  via the new staged, locked-nixpkgs script; a canary check now guards
  the dark-variant fingerprint.
- `deadnix` no longer fails on vendored third-party `flake.nix` files
  (excluded) — a latent red final gate.

### Security

- Auth-hardening pass over the login surfaces (2026-10-06): the session
  cookie now carries `Secure` on TLS-fronted deployments — behind the
  consuming stack's TLS-terminating proxy `r.TLS` is always nil, so the
  flag is derived from the same https trusted-origin signal the CSRF
  cookie uses (`session.CookiePolicy` wired once in `server.New`;
  directly-TLS requests still upgrade on their own); login/logout CSRF
  rotation deletes the cookie with the REAL config's attributes (a
  non-Secure deletion of a Secure cookie is rejected by strict
  browsers, silently degrading the rotation); and the `/hooks/*` shared
  secret is compared constant-time over SHA-256 digests (no
  early-exit/length leak — the same bar the CRM seam holds). Pinned by
  TestCookieSecurePolicy and the TLS-fronted login spec.

## [2.8.0] - 2026-09-30

Ships together with the never-separately-tagged [2.7.0] content below
(prod jumps 2.6.0 → 2.8.0 in one release).

### Breaking (NixOS module)

- `services.webphone.nginx.*` is REPLACED by `services.webphone.caddy.*`:
  the module now generates a Caddy vhost (automatic HTTPS, `encode zstd
  gzip`, unbuffered `/events` via `flush_interval -1`) instead of an
  nginx one. The old nginx generator had zero users — the telephony
  stack force-disables it (`nginx.enable = mkForce false` plus an
  assertion) and fronts the app with its own vhost. `nginx.gzip.enable`
  is gone with it (Caddy compresses unconditionally). The csrf fronting
  defaults (`trusted_proxies`/`trusted_origins`) now derive from
  `caddy.enable`/`caddy.hostName` with the same precedence. New
  `caddy.sipUpstream` bridges the SIP WebSocket path to the PBX when
  set — fixing a latent dead-end: the old vhost proxied the websocket
  path to the app, which never terminates the SIP wss (the island's
  connection must reach the PBX; the stack routes it there itself).

### Changed

- Error architecture: completed the tier-2 family adoption — every
  internal error constructor site now builds `go-error-family` errors
  with stable dot-notation codes (blob/store/session/config/pbx/crm/
  domain/messaging/fax/gateway seams; sentinels keep their
  `errors.New` identity and classify via package registration).
  Wire behavior is unchanged: rendered strings, HTTP statuses, and
  the classify ladder are pinned by tests; log lines gain the
  `[family:code]` prefix. Startup wiring wraps stay family-neutral by
  design (inner constructors own the family). The enforced erraudit
  tier-2 set reads zero findings; the decision record lives in the
  error-excellence plan appendix.
- Added `SECURITY.md` (nix-ssh-config parity adapted to a runtime
  service): GitHub private vulnerability reporting as the preferred
  channel, per-surface triage guidance (island / tabs / phone-api
  proxy / bundle / NixOS module), posture facts verified at module
  source, tags-trail-versions note for lock-riding consumers.
- Webhook gateway: attachment and fax file parts now carry their honest
  Content-Type on the wire (the attachment's stored mime; fax parts are
  `application/pdf`) instead of the `application/octet-stream` default
  Go's `CreateFormFile` stamps on every part. The producing side owns
  the type, so consuming bridges can stop magic-byte sniffing; the
  Content-Disposition bytes are unchanged.
- Nix: split the 803-line `flake.nix` into focused flake-parts modules
  under `nix/` (packages, checks, module-check, vm-tests, apps,
  devshell, treefmt). The `webphoneVersion` binding stays in
  `flake.nix` — `scripts/release.sh` continues to sed it there.
- Internal: the pbx client's `ErrDisabled`/`ErrUnauthorized` sentinels
  construct via `errors.New` (matching crm/store) instead of
  directive-less `fmt.Errorf` — zero behavior change, identity
  preserved for `errors.Is`.
- Typography accessibility + craft (retro bullet, added by the
  docs-health v6 sweep 2026-10-03: the 2026-09-30 typography train
  shipped in this release without an entry). The root `html` rule now
  sizes text as a percentage (93.75%) so browser font-size preferences
  are honored (px ignores them), carries the rendering baseline
  (antialiasing), and sets a local mono stack for the event log and
  ICE panel; headings get `text-wrap: balance` and message bubbles
  `text-wrap: pretty`. One real defect fixed with it: `html lang` now
  follows the session language instead of staying static
  (`TestShellHtmlLangFollowsSessionLang` pins it).

### Fixed

- NixOS module: `services.webphone.backup.destDir` now gets the same
  `/var/lib/` assertion as `dataDir` (an off-`/var/lib` path used to
  render a unit systemd refuses to load, with no clear message).
- Nix: the flake-version drift test no longer false-matches the
  version string inside comments — the 2026-09-29 flake split added a
  comment containing the literal `webphoneVersion = "`, tripping the
  unanchored regex; it now anchors to the assignment.
- NixOS module: backup oneshot orders `after = [ "webphone.service" ]`
  so a boot-time Persistent timer catch-up cannot race the database
  into existence.

### Security

- NixOS module: both units set `UMask=0077` and
  `StateDirectoryMode=0750` — message threads, faxes, and voicemail
  blobs are no longer world-readable on disk (existing files keep
  their old modes; re-backup or `chmod` to tighten in place).

## [2.7.0 (staged 2026-09-24, never tagged; ships in v2.8.0 above)]

### Added

- Self-sends to your own number are refused before the provider
  roundtrip (send-failure train C): a message or fax addressed to the
  extension's own DID (config `identities`) fails fast with the 422
  refusal arm — the provider (Telnyx 40310) would reject it anyway —
  while the failed row/job is still persisted first, so the evidence a
  provider refusal would have left survives. The guard is off entirely
  when no identities are configured; both lanes pinned by service and
  server tests.
- Empty states now render through the templ-components
  `display.EmptyState` (six true empty-state sites: messages search,
  thread list, fax list, contacts, voicemail inbox, history) with a
  permanent scoped Tailwind v4 build served at `/assets/tw.css`
  (18.9KB): coexistence with the hand-rolled token CSS is PROVEN, not
  assumed — a Chromium A/B spike found all existing surfaces
  byte-identical across 14 computed properties because Tailwind v4
  emits `@layer` only and unlayered CSS wins every collision (verdict
  with data: docs/planning/archived/2026-09-24_16-38).

### Changed

- Provider refusals answer 422 with their reason (send-failure train
  E): a provider that ANSWERS with a 4xx (Telnyx 40310 self-send,
  invalid destination, content policy) is input feedback the user can
  act on, not a system fault — the refusal arm of the shared
  send-failure ladder moved from 502 to 422 while keeping the
  provider's own detail text verbatim and the no-retry affordance. A
  provider 5xx ANSWER classifies transient by its own status and now
  lands on the 502 transport arm with the honest `family=transient`
  log line. Pinned by the classify table, the messages refusal arm,
  and a new fax-lane refusal test; error-contract + stack runbook
  ladders updated in lockstep.
- Dependency train swept: cqrs-htmx v4.12.0, httputil v1.3.0,
  go-error-family v0.10.2, go-sse v0.6.1 (go-health held at v0.3.0 —
  the evaluation-hook telemetry idea was parked with evidence: the
  hook cannot fire in this wiring; see TODO_LIST).
- Internal quality: family-field logging via
  `errorfamily.LogErrorContext`, library asserts via
  `errorfamilytest`, and the middleware chain pinned by httputil's
  19-spec httpspec suite (all green first run — zero divergences).

### Fixed

- Upload bodies are bounded (61MB envelope) before multipart parsing:
  `http.MaxBytesReader` wraps the request body, so an oversized upload
  answers a plain 400 instead of streaming unbounded into memory — the
  ParseMultipartForm argument was only a memory threshold, not a size
  cap. Pinned by a unit test and a live smoke probe.
- Shared contacts render in the dial typeahead again: `SharedContact`
  marshaled Go-style `Name`/`Number` onto `window.PBX_CONFIG` while
  the island reads lowercase `name`/`number` (the README always
  documented lowercase), so every operator-configured shared contact
  silently vanished from the dial suggestions. The json tags now pin
  the wire shape, asserted capitalized-absent by
  `TestConfigJSContactsWireKeys` (re-do of the fix the 2026-09-24 host
  reboot destroyed in a /tmp worktree).
- Store/render failures on panel and thread loads send only the
  op-prefixed family default to the browser; the raw error text goes
  to the operator log (one-home `internalError`/`safeDetail` helpers,
  SafeDetail at every 500 writer).
- Fail-closed surfaces answer 503 with a `Retry-After: 1` hint: the
  unconfigured inbound-hooks gate and the unconfigured phone-api proxy
  tell well-behaved callers when to come back (liveness probes ignore
  headers, so startupz deliberately carries none).
- The message composer's attachment picker now offers the audio and
  video types the messaging bridge already delivers (`bd77669`): the
  dialog was filtered to images, PDFs and vCards, so a user could not
  even select a voice note or video clip the backend would have
  accepted. The accept list mirrors the bridge's media types
  (mp3/wav/amr/ogg audio, mp4/3gpp/mov video) and is pinned by a
  render test — the island suite stubs the file input and cannot see
  the attribute.

## [2.6.0] - 2026-09-23

### Added

- Call cards carry a state chip + spoken transitions (plan T20a):
  `data-state` (ringing/established/ending) drives a colored dot —
  pulsing while ringing, green when established,
  `prefers-reduced-motion` aware — and state TRANSITIONS announce
  through the toast live region (en/de); the per-second duration tick
  stays silent.
- Affordances across the surfaces (plan T20c/d/e): thread-LIST rows
  offer a dial button OUTSIDE the anchor; history and voicemail rows
  offer `data-sms` (Messages tab, recipient prefilled) and the ☆
  save-as-contact bridging to the island (`wp:save-contact` — the
  island keeps sole /api/contacts ownership, saved/failed toasts en/de,
  the CDR's caller-id name rides along first).
- Bounded retention (plan T25): `retention_days` (default 0 = keep
  everything forever). When set, a daily sweep (boot + 24h ticker in
  the binary — no extra endpoint to protect) deletes messages with
  their attachments, fax jobs with their documents, and the threads
  those deletions empty; blob files are collected before their rows go
  and unlinked after. Session rows stay out of scope: the session
  store already sweeps its own expiry. Pinned by a store sweep test
  (old-only deletion, live content kept, idempotent second pass,
  empty-thread age rule).
- Short-lived TURN credentials (plan T26b): `turn_rest.secret` +
  `turn_rest.ttl` (default 48h). When the secret is set, `/config.js`
  derives a coturn REST pair per response (username = unix expiry,
  credential = base64(HMAC-SHA1(secret, username))) for every
  `turn:`/`turns:` entry, so a long-lived TURN password never sits in
  config or the browser; STUN-only entries and the static passthrough
  (secret unset) stay verbatim. Validation rejects a secret without a
  TURN URL (dead config) and non-positive TTLs.
- Per-extension data export (plan T26c): `GET /api/export`
  (session-gated, Settings-tab download link) zips everything the
  signed-in extension owns — messages.json (every thread with its
  messages and attachment manifest), faxes.json, contacts.vcf. Owner
  scoping comes from the store queries; document blobs stay in the
  blob store (names/sizes listed in the JSON), keeping the archive a
  portable manifest.
- Per-thread composer drafts (plan T21d): message text survives tab
  and thread switches — the paths that re-render the composer empty.
  Drafts save debounced (4k cap), restore only into an EMPTY composer,
  and clear after a successful send; a failed send keeps the draft for
  the retry. Pinned by a shell spec driving the real listeners.
- The failed bubble tells its story (plan T21a/b): failure kind +
  reason persist on the message row (idempotent column migration);
  the send path classifies (rejection vs transient via the error
  family), the delivery webhook persists the provider reason, and a
  delivered verdict clears it. The bubble shows the reason and a
  retry form (same recipient + body) ONLY for transient failures; the
  delivered badge gained the ✓ glyph.
- Inline image thumbnails with sniffed types (plan T21c + T26e):
  image attachments render lazy CSS-scaled inline; the server no
  longer trusts client-declared multipart types
  (http.DetectContentType when missing or octet-stream). The
  legacy-contacts migration announces itself (en/de toast).
- Single-source DOM contract (plan T22): the 35 island element ids now
  live in `docs/dom-contract.md`; `TestServedPageHoldsTheDomContract`
  parses that file as the golden source, so the test, AGENTS, and the
  stack-side docs can never drift from one hand-maintained list.
- Optional Ledger CRM integration (both repos, off by default):
  `crm.url` + `crm.token` (config / `WEBPHONE_CRM__*`) point at a
  Ledger CRM running with `-api-token`. Phone numbers in History,
  Messages, Fax and Voicemail resolve against the CRM's contacts
  (digit matching, trunk/country-code tolerant, TTL-cached 6 h,
  misses 5 min; any failure degrades to the raw number), and every
  finished call is reported by the island to `POST /api/calls`
  (session + CSRF gated) which journals it on the matching CRM
  contact — unknown numbers are never logged and the integration
  never mints contacts. CRM side: contacts gained a `phones` list
  (form, search, CSV import) plus the machine API
  (`GET /api/contacts/by-phone`, `POST /api/contacts/{id}/calls`)
  behind `-api-token`. Pinned by `internal/crm` unit tests, the
  `/api/calls` contract test, the history-enrichment render test,
  and the CRM's api_test + phone-matching tables.
- Dated backup history (plan T18a):
  `services.webphone.backup.retentionDays` (default `null` = today's
  single-snapshot behavior). When set — e.g. `30` — each daily run
  also writes `destDir/snapshots/<date>/` (blob tree hardlinked
  against the previous snapshot, so unchanged blobs cost no space) and
  prunes snapshot directories older than the given days; point-in-time
  restore copies db + blobs back from the chosen directory. Pinned by
  the module eval check (option gates the history/prune script) and
  the KVM backup VM test (stale dir pruned, today's snapshot intact,
  same-day rerun never links against itself). Also documents the
  readiness contract (T18c): the unit deliberately stays
  `Type=simple` — `/startupz` is the readiness truth, not sd_notify.
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
  in `docs/planning/archived/2026-09-22_14-45_csrf-rotation-on-slide-verdict.md`).
- `scripts/webphone-smoke.py --expect-version X.Y.Z`: asserts the
  running server's `/version` (verified positive and negative) — the
  deploy-verification companion.
- Double-submit guard on outbound sends: the new-message, thread
  reply, and fax forms disable their submit button for the duration
  of the request (`hx-disabled-elt`) — a live self-send test produced
  TWO identical failed messages because the multi-second gateway
  round-trip left Send live. Pinned by `TestSendFormsDisableWhileInFlight`.
- Self-send notice (plan
  `docs/planning/archived/2026-09-22_16-07_SUPERB-send-failure-ux.md`): opening
  a thread whose remote number is the extension's own DID (config
  `identities`) renders a warn notice (en/de) at intent time —
  providers refuse self-addressed sends (Telnyx 40310), and the
  notice lands before the user can discover that by failing. The
  comparison runs both sides through `ParsePhone` (config DID spacing
  cannot hide the match); pinned by `TestIsSelfThread` and
  `TestThreadViewWarnsOnSelfSend`.
- Composer UX batch (plan
  `docs/planning/archived/2026-09-22_16-36_SUPERB-composer-ux.md`): message
  composers are auto-growing textareas (Enter sends, Shift+Enter breaks
  a line — IME-composition guarded, shell.js; the multipart wire format
  is unchanged), a correct SMS segment counter shows "N SMS" past one
  segment (GSM-7 160/153 with two-unit extension characters, UCS-2
  70/67 — language-neutral by design), attachment chips with remove
  cover the reply and fax file inputs (DataTransfer rebuild,
  feature-detected; the native input stays the fallback), and German
  renders 24h timestamps (`formatClock` "16:09" for bubbles,
  `formatStamp` "22.09. 16:09" for row fallbacks in messages, fax and
  voicemail) while English output stays byte-identical. Pinned by
  `TestFormatClockAndStampFollowLanguage`,
  `TestComposerCarriesSegmentCounterAndTextarea`,
  `TestBubbleClockFollowsLanguage`, and five composer specs in
  `island-tests/composer.test.mjs` (49/49 node tests).
- `/metrics` endpoint (plan T26a): Prometheus text format, AGGREGATES
  only by design — build info, uptime, table counts — never
  per-extension data; a test fails on any extension-like string, and
  the NixOS module gained a dedicated fenced `/metrics` vhost location
  (asserted by the module flake check).
- `services.webphone.nginx.gzip.enable` (plan T27a): flips nginx's
  recommended gzip settings on the generated vhost (SSE is never
  gzipped by nginx itself); module-eval stand-in pinned.
- Signed release tags (plan T27c): `scripts/release.sh` now cuts
  `git tag -s` and verifies the signature locally before pushing.
- Thread search in the Messages tab (ux-raw-ideas E): a debounced
  search form (300 ms, morph-safe focus via `wp-thread-search-input`)
  filters threads by remote or any message body — `SearchThreads`
  matches with `ESCAPE '\''` handling so `%`/`_` search literally,
  orders by last activity, and stays owner-scoped (7-case store test).
  Live SSE pushes are suppressed while a query is active (shell.js
  guard) so a filtered view is never stomped by a newest-row push.
- Audio output picker (ux-raw-ideas F): when the machine exposes two
  or more real outputs, the island unhides a `setSinkId` picker for the
  REMOTE audio only (ring tones stay room alarms); the pick persists
  in `localStorage["wp-sink"]`, follows `devicechange` with live
  selection preserved, re-labels on language change, and every
  unsupported/failure path stays hidden (6 island specs incl. the
  single-output env).
- Hover timestamps (ux-raw-ideas A): `fullStamp` titles on thread-row
  relative times and bubble clocks carry the full date (de
  "22.09.2026 16:09", en "Sep 22, 2026, 4:09PM") — relative stamps
  stay relative, the absolute truth is one hover away. Pinned by
  `TestFullStampCarriesDateYearAndTime` and
  `TestTranscriptCarriesHoverStampsAndJumpChip`.
- Jump-to-latest chip (ux-raw-ideas B): while the transcript is
  scrolled away from the bottom, live SSE pushes accumulate into a
  hidden server-rendered chip ("↓ N new", language-neutral); clicking
  returns to the newest message, near-bottom pushes still pin the
  view, scrolling back down resets it (shell.js §3b-2, 2 specs).
- Missed-call badge (ux-raw-ideas C): the island dispatches
  `wp:call-missed` on the two genuinely-missed paths (caller gave up
  pre-answer; accepted call died before media — a deliberate REJECT
  never dispatches); shell.js renders an English "missed · N" header
  badge cleared when History opens (5 island specs).
- Dial typeahead (ux-raw-ideas D): the dial field suggests
  `PBX_CONFIG.contacts` as you type — zero round-trips, ranked
  name-prefix < name-contains < number-contains, capped at 6, full
  keyboard support (↑/↓/Enter/Escape), idempotent init, and a
  JS-created listbox so the served DOM contract stays untouched;
  re-labels on language change (en/de).
- Idempotent CRM call journal: `POST /api/calls` accepts an
  island-generated UUID `key` — a browser-level retry or double-fire
  replays the SAME key and journals once (replay answers inert 204);
  a failed (502) attempt stays retryable, the silent unknown-number
  drop consumes its key too, and an absent/empty key keeps the legacy
  never-dedupe shape (4 new contract tests). The island sends
  `crypto.randomUUID()` per ended call; dedupe is extension-
  namespaced with its own 1h TTL store.
- CRM resolver hardening: a 30-row history cache-miss burst now
  fans out ONE upstream lookup per number (leader/waiter single-
  flight, waiters honor their own context), and the resolver
  counts upstream outcomes (hit/miss/failure, upstream round-trips
  only — cache hits don't count). Both nil-safe, race-tested.
- `/metrics` gains `webphone_crm_lookups_total{outcome="hit|miss
  |failure"}` when the CRM integration is on; the family is absent
  entirely when off (a disabled deploy must not publish zero-lines
  that read as "CRM broken"). Aggregates only — pinned by the
  leak-guard.

### Changed

- `/favicon.ico` answers with the SVG icon (plan T27b); the views
  gained the i18n referenced-keys guard (every `T(lang, "…")` literal
  must exist in the dictionary — T() would render the raw key at
  runtime) and a validated `timezone` config key owning every wall
  clock and log line (plan T27d + T26d).
- The dial field clears once the INVITE actually went out
  (`placeCall` reports it; validation errors keep the typed text) —
  until now the value accumulated, and the stack E2E literally dialed
  "10011001" on its post-transfer redials.
- AGENTS.md size pass (plan T16a): 705 → 337 lines. Every rule stays;
  war stories moved to `docs/lessons.md`, the failure→feedback table
  to `docs/error-contract.md` (now carrying the rate-limit-keying ops
  note), the release dance + pbx-artmann relock ritual to
  `docs/release-runbook.md`.
- `GOEXPERIMENT=jsonv2` is gone everywhere (devShell, buildGoModule,
  buildflow env, release.sh, smoke, README/CONTRIBUTING/AGENTS): Go
  1.27.1 ships a stable `encoding/json/v2`, the flag had become a
  footgun (it contributed to a failed release run outside the
  devShell), and the full suite is green 13/13 without it.
- Zero inline scripts: templ-components v1.19.2 ships the
  `PageProps.NoThemeScript` knob this repo's roadmap asked for, so the
  CSP-hash-pinned library theme preload is replaced by the same-origin
  `/assets/theme-preload.js` (render-blocking, sets `data-theme`
  pre-paint like shell.js §4). script-src drops the hash — a
  dependency bump can never change served script bytes — and the
  app.css forced-theme `color-scheme` rules lose their `!important`
  (nothing inline fights them anymore). `TestServedPageSatisfiesStrictCSP`
  now fails on ANY inline script instead of pinning one.
- templ-components bumped v1.18.0 → v1.19.2 (upstream: popover/dropdown
  positioning fixes, `DropdownProps.Trigger`, `ListNote` count variant,
  `NoMainWrapper`, `NoThemeScript`).
- `crm.Client` adopts the same disabled-policy chokepoint as
  `pbx.Client` (a `do()` prologue that fails every method with the
  sentinel when the integration is unconfigured; a nil client is
  nil-safe) — the deliberate split brain between the two gateway
  seams is closed, and a disabled CRM builds no URL at all.

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
  `docs/planning/archived/2026-09-20_17-41_session-persistence-spike-verdict.md`
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
  tiers, `docs/planning/archived/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`):
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

[Unreleased]: https://github.com/LarsArtmann/webphone/compare/v2.8.0...HEAD
[2.8.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.8.0
[2.7.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.8.0
[2.6.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.6.0
[2.5.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.5.0
[2.4.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.4.0
[2.3.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.3.0
[2.2.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.2.0
[2.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.1.0
[2.0.0]: https://github.com/LarsArtmann/webphone/releases/tag/v2.0.0
[0.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v0.1.0
