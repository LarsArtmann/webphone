# AGENTS.md

Enduring context for AI sessions working in this repo. Rules live
here; war stories, evidence, and train chronology live in
[docs/lessons.md](docs/lessons.md), the error-surface table in
[docs/error-contract.md](docs/error-contract.md), the DOM contract in
[docs/dom-contract.md](docs/dom-contract.md), the dedup rulings in
[docs/dedup-registry.md](docs/dedup-registry.md), the release dance in
[docs/release-runbook.md](docs/release-runbook.md). This file is capped
at 377 lines by BuildFlow's `docs/agents-md-size` preflight — move
evidence to docs/, keep rules here.

## What this is

A single-binary Go unified-communications web app: the proven SIP.js
call island plus server-rendered tabs (Messages SMS/MMS, Fax,
Voicemail, History, Contacts, Settings) on one page. Built on
cqrs-htmx (root library only) + templ-components `layout.Base` +
SQLite (modernc), wired through a samber/do v2 composition root
(`internal/app`) with go-health probes and an optional
go-health-dashboard at `/health`. The consuming stack
([nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony))
imports this repo's `nixosModules.default`, fronts the binary with
TLS + the WSS `/sip` proxy, and RIDES webphone `main` (per-train lock
bump); pbx-artmann consumes the stack via a rev pin. Current version
lives in `flake.nix` (`webphoneVersion`); train chronology in
docs/lessons.md.

The cqrs-htmx `setup` bundle stays REJECTED (split-brain identity + a
measured +68% binary-size NO-GO); the shell VALUE landed via
`httputil.NewServer` instead. Evidence: docs/lessons.md.

## Tri-repo integration rules

- Changes must be PUSHED in webphone before the stack re-pins, and the
  stack tree must be CLEAN before pbx-artmann re-locks (its narHash
  covers the whole tree). Re-lock ritual: [docs/release-runbook.md](docs/release-runbook.md).
- The NixOS module options: `enable`, `package`, `dataDir` (MUST live
  under `/var/lib/` — assertion), `settings` (freeform),
  `environmentFile`, `memoryMax`, `csrf.{trustedProxies,trustedOrigins}`,
  `serverTiming.enable`, `backup.{enable,destDir,calendar,retentionDays}`
  (retentionDays null = single snapshot; set = dated history + prune),
  `caddy.{enable,hostName,sipUpstream}`,
  `caddy.hsts.{enable,maxAge}` (the front is CADDY, not nginx — the
  module's old nginx generator was never used). The generated vhost
  reverse-proxies the app, bridges the websocket path ONLY when
  `caddy.sipUpstream` is set (the Go app never terminates the SIP wss),
  and flushes `/events` unbuffered (`flush_interval -1`); probes fence
  via `remote_ip` matchers, not per-location blocks. `backup.destDir`
  shares the `/var/lib/` assertion; both units set `UMask=0077` +
  `StateDirectoryMode=0750`; the backup oneshot orders after
  `webphone.service`.
- The `webphone-module` flake check evaluates the module with stand-in
  options (caddy/systemd/users + `assertions`; new config keys the
  module writes need a stand-in there). Statix pins single-assignment
  style. `checks.x86_64-linux.webphone-backup` (KVM-gated) and
  `webphone-backup-drill` cover backup snapshot AND restore.
- **flake.nix layout**: flake.nix is a slim ENTRY (inputs, systems,
  imports, nixosModules); the meat lives in `nix/packages.nix`,
  `nix/checks.nix`, `nix/module-check.nix` (+ `module-check-base.nix`
  and the `module-check-csrf.nix`/`module-check-backup.nix` case
  groups), `nix/vm-tests.nix`, `nix/apps.nix`, `nix/devshell.nix`,
  `nix/treefmt.nix`. The `webphoneVersion` let-binding MUST stay in
  flake.nix (`scripts/release.sh` greps/seds it there) and reaches
  `nix/packages.nix` via `{ _module.args.webphoneVersion = ...; }`.
  Gotcha: `self` is a TOP-level-only flake-parts module arg — declare
  it on the module function, never inside `perSystem`.
- webphone's gateway seam (loopback vs webhook) is consumed by
  pbx-artmann's `telnyx-webhooks.py` bridge — contracts in the plan
  docs under `docs/planning/archived/2026-09-19_11-51_SUPERB-*`.

## Owner decisions (2026-09-20)

- Stack input policy: ride webphone `main` with a per-train lock
  bump. pbx-artmann keeps its pin type while all three repos live on
  this host.
- Sanitization: the island keeps letters (`[^\d+*#a-zA-Z]`), matching
  `sanitizeDialable`; pinned on the served asset.
- Own-number feed: static config map (`identities`); a stack
  `/phone-api` identity endpoint is the upgrade path. CDR-derive
  REJECTED on evidence (outbound `caller_id_number` is
  dialplan-dependent and absent before the first call).

## Commands

```console
nix develop                        # Go, templ, golangci-lint, esbuild, … — GOTOOLCHAIN=local; bare `go` works inside. Host go is below the 1.27.1 floor: OUTSIDE-the-shell go commands need `nix develop -c`
nix develop .#ci                   # minimal CI shell (go + templ + golangci-lint only); what the `go-tests` CI job enters
templ generate ./internal/web/views/   # after ANY .templ edit (committed *_templ.go)
nix develop -c go test -count=1 ./...  # -count=1: the result cache has lied during investigations
python3 scripts/webphone-smoke.py          # 40-check live smoke (+4-check restart scenario; boots a fresh binary; --base URL reuses a server; --expect-version X asserts /version)
buildflow                                  # the quality gate; BUILDFLOW_NO_RESULT_CACHE=1 for full (release.sh also gates on `nix run .#vulnix`)
nix run .#vulnix                           # vulnix over the RUNTIME closure; verdict logic = `webphone-vulnix-triage` CLI, fixture-checked
nix flake check                            # package + tests in sandbox + treefmt + island-lint + island-js + kvm-gated backup VM test
nix build .#checks.x86_64-linux.webphone-module   # ONE check without the whole flake check (also: webphone-backup-drill; webphone-backup is KVM-gated)
nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs   # island JS tests alone (stubs in island-tests/helpers.mjs). Host `node` also runs them.
nix build .#webphone --system aarch64-linux   # cross-builds — verify by ELF bytes, never exit code alone (docs/lessons.md)
./update.sh [version]              # repin vendored sip.js (fetch → esbuild IIFE → swap)
```

Smoke a binary quickly (loopback gateway = whole product, zero PBX):

```console
nix develop -c go build -o /tmp/webphone-bin ./cmd/webphone
WEBPHONE_ADDR=127.0.0.1:18099 WEBPHONE_DATA_DIR=/tmp/wp-data \
  WEBPHONE_GATEWAY__WEBHOOK_SECRET=devsecret /tmp/webphone-bin
```

Host-nix-down fallback (the daemon can break while the store stays
healthy): run gates with STORE toolchains directly — `GOTOOLCHAIN=local
/nix/store/*-go-1.27*/bin/go test -count=1 ./...`, the store nodejs
for island tests, and `-ldflags "-X
github.com/larsartmann/webphone/internal/server.buildVersion=vX.Y.Z"`
for a version-stamped smoke binary (NOT `main.displayVersion`). Full
story: docs/lessons.md.

## The DOM + bundle contract (DO NOT BREAK CASUALLY)

The id list lives in [docs/dom-contract.md](docs/dom-contract.md) —
the SINGLE source, parsed by `TestServedPageHoldsTheDomContract`. The
consuming stack's browser E2E (`tests/browser-e2e.py` there) drives
the island remotely — re-run it after any markup change (obligation
noted in the release runbook).

- The island modules under `internal/web/assets/island/app/` are
  served VERBATIM (no bundling, no minification), so the E2E's
  greppable strings survive by construction. Keep it that way.
- The event log (`#log`) stays **English** in both UI languages:
  operator-facing diagnostics, grepped by the runbook.
- Unknown paths render the STYLED 404 (`notFoundPage` in pages.go;
  status stays 404). Partial swaps never see it; the smoke's
  stale-CSRF probe still reads a plain 404 status.
- The island never unloads: tab navigation swaps partials into
  `#tab-content`; the SIP island lives OUTSIDE that region so calls
  survive tab switches. Deep links render the full shell.

## Architecture invariants

- **Middleware chain** (server `New`, in this ORDER):
  `ContextEnrichmentMiddleware(nil)` → `RequestLoggingSlog` →
  `ServerTimingMiddlewareWhen` (env-gated) → `SecurityHeaders` →
  `cqrshtmx.RecoveryMiddleware` → routes. Enrichment stays OUTSIDE
  the request log (`request_id=` in every line + `X-Request-ID`);
  user extractor stays nil. `securityHeadersConfig()` is the single
  header-config source; `Permissions-Policy` ships calibrated
  (`microphone=(self)`; camera/display-capture/geolocation/payment/usb
  denied). Limiters are `httputil.KeyedRateLimiter` with port-stripped
  peer-host keys (`remoteHostKey`; flip to `KeyExtractorFromClientIP`
  only once the stack proves XFF sanitization). `/healthz` = honest
  readiness (sqlite ping + blob write probe via the ONE home
  `blob.ProbeWrite`, 2s bounds, 503 names the failing check); `/livez`
  = fetch-free liveness; `/startupz` = latched 503-until-first-pass
  (served from `Deps.Probe` over the named services `sqlite`/`blob-dir`;
  nil Probe = the NewChecks fallback for hand-composed test Deps).
  `/events` rides `Broadcaster.ServeSSE`; the chain is pinned by
  httputil's 19-spec httpspec suite (`TestHTTPSpectChainConformance`).
  The unit deliberately stays `Type=simple` — see README "Readiness
  vs systemd".
- **Composition root is `internal/app`** (samber/do v2): ONE container
  owns object lifetime; `cmd/webphone` owns process concerns (config,
  logging, signals, the HTTP listener). Rules: the injector lives ONLY
  in `internal/app` (services hold resolved deps, never the container,
  so service packages stay framework-free; the do lifecycle-interface
  conformance is asserted adapter-side in app.go); the critical pair
  registers NAMED (`sqlite`, `blob-dir`) because the names are the
  probe's critical-service contract; both + the handler are EAGERLY
  invoked in `New`; the probe is built ONCE and threaded to both
  server.Deps and the dashboard (never registered in the injector it
  reads — `*health.Probe` conforms to the health-check interface,
  self-registration recurses); shutdown order = HTTP drain →
  `Probe.Shutdown` → `do.Shutdown` cascade (idempotent).
  `Deps.Probe`/`Deps.Dashboard` nil-fallback keeps every existing test
  composition working. `samber-linter` HW-4 on the named pair is
  suppressed with a reason (eagerly resolved via `MustInvokeNamed`;
  the linter cannot see transitive resolution).
- **Health dashboard seam** (go-health-dashboard v0.10.1):
  config-gated `dashboard.{enable,title}`, DEFAULT OFF — fence it like
  the probe triple; its Caddy vhost proxies `/health/*` unbuffered
  like `/events`. Mounted at `/health` (subtree patterns `/health` AND
  `/health/`); probe aliases at `/health/{livez,readyz,startupz}` (SAME
  probe instance). CSP: subtree override via `httputil.Nonce` +
  `dashboard.RecommendedCSP` — `unsafe-eval` scoped to `/health` ONLY
  (Datastar compiles expressions); every other surface keeps the strict
  app policy (`TestAppServesMainPageUnderStrictCSP` pins the boundary).
  Its stylesheet is its OWN scoped build `/assets/health.css` (treefmt
  owns it; rebuild with `nix run nixpkgs#tailwindcss_4` — NEVER
  nixpkgs#tailwindcss v3; never merges with `tw.css`). Probe refresh:
  1s loop ONLY while enabled; off = interval 0. The smoke boots WITH the
  dashboard on. Evidence: docs/lessons.md.
- **Module graph stays acyclic**: the island's `state.js` + `auth.js`
  exist so calls/ice/connection never import each other; ice syncs
  via the `wp:calls-changed` CustomEvent — enforced by
  `internal/arch/arch_test.go` (also: `domain` imports nothing
  internal; services never import `server`/`web`).
- **Sessions**: the server VERIFIES extension/password against the
  PBX directory before minting (fail-closed; loopback dev skips and
  WARNs). The store is a SEAM: prod `NewSQLiteStore` (sessions
  survive restarts; the row carries the extension + directory
  password the `/phone-api` proxy needs — credentials at rest
  accepted, swept on read/Create), tests/loopback `NewMemStore`.
  Sessions SLIDE: `GET /api/session` resumes a live cookie at island
  boot (no-store, requireSession-gated); activity past the idle
  halfway point renews (`session_ttl`=7d idle, `session_max_ttl`=30d
  absolute defaults). Login/hooks per-IP rate limited. Handlers
  self-gate via `requireSession`; `Sessions.Require` additionally
  wires `/events` + `/phone-api/` — dual-layer by design. The contract
  test allowlists exactly three 401 writers: actions.go, webhooks.go,
  session_api.go.
- **Language**: per extension — `wp-lang` cookie (samesite=strict) →
  `Accept-Language: de*` → English. `ExtensionHubs` remember the
  negotiated lang so SSE fragments render in it. Service validation
  reasons stay English (operator-facing, same policy as `#log`).
  i18n dictionaries live in `views/i18n.go`; add new keys to BOTH
  maps (a test keeps en/de in sync; unknown keys surface themselves
  in the page, deliberately).
- **Live-update surfaces morph-swap**: the SSE/nav surfaces — thread
  list, `#thread-transcript`, fax list, contacts panel, voicemail
  panel re-fetch, shell.js `refreshNav` — carry
  `hx-swap="morph:innerHTML"` (idiomorph via `/htmx-ext.js`, ONE
  bundle). Morph preserves focus/drafts/listeners; STATEFUL nodes in
  morph surfaces must carry stable ids. Payloads stay bare fragments.
  Event names: `threads`, `thread`, `fax`, `voicemail`, `contacts`.
  The `voicemail` and `contacts` events are payload-less NUDGES. New
  live surfaces follow the morph pattern.
- **Gateway seam**: loopback (dev) vs webhook (multipart to
  `{url}/message|/fax`, Bearer secret, `{"provider_ref"}` receipt).
  File parts carry their HONEST Content-Type (`createFilePart` in
  internal/gateway/webhook.go: the attachment's stored mime; fax parts
  `application/pdf`) — the producer owns the type; the Telnyx bridge
  prefers it, magic-byte-sniffing only as the octet-stream fallback
  for pre-`e6ea2c7` binaries. Never write version claims for unreleased
  code — pin by date or commit. The part-header block is byte-pinned by
  `TestProviderFormPartHeaderBlockGolden`; a webhook-lane smoke probe
  stays DECIDED-AGAINST. Self-sends to the owner's own DID (config
  `identities`) never reach the provider: `gateway.SelfSendRejection`
  refuses locally after the row is persisted, riding the 422 refusal
  arm. Inbound hooks `/hooks/*` share the same secret and fail CLOSED
  (503) when none is configured. Status hooks are idempotent
  (`hooksIdem`, in-memory 1h TTL; only successes recorded; replays
  answer `202` inertly). Evidence: docs/lessons.md.
- **Owner scoping everywhere**: every store query is extension-scoped;
  attachments/faxes stream through session-gated handlers only.
- **One-home helpers**: `listRows[T]` (store/db.go) owns the
  query→close→scan→`rows.Err()` lifecycle; `pbx.do()` is the single
  disabled-policy chokepoint; `session.makeSession` owns the session
  birth invariant; `server.requireMultipartTo` is the send-form
  prologue; `domain.must`, `domain.OrClock`, `store.updatedOrNotFound`,
  `views.formatFor`, `server.crmNumbers`, `server.applyStatusWebhook`,
  `server.recordCallIdem`, `server.contactSaveFailed`,
  `server.apiContactSaved`, `views.panelHead`, `views.errorBanner`,
  `views.panelError`, `views.identityLine` — each with its own
  micro-test, several pinned byte-exact in `views/panels_test.go`.
  **Dedup acceptance registry**: the ONE home for every accepted/
  declined clone ruling + its WHY is
  [docs/dedup-registry.md](docs/dedup-registry.md) — read it BEFORE
  re-litigating an accepted similarity, and append a sweep-log line
  per run. `-t 3` is the working baseline pending owner ratification.
- **Personal contacts have ONE home**: the per-extension SQLite
  store, read/written via `/api/contacts` (session-gated, 60/min POST
  limiter, 500-per-extension atomic cap). Mutations answer 204; the
  LIST is the only id source (the store upserts by (owner, phone),
  keeping the old id on rename). The legacy
  `localStorage["pbx-contacts"]` list imports once post-login and is
  REMOVED only after the server accepted every row. Load trigger:
  session.js dispatches `wp:session-opened` AFTER cookie mint + CSRF
  adoption. History stays hybrid BY DESIGN (local session log + same
  CDR API) — not a split brain, don't "fix" it.
- **Tab→island affordances live in shell.js**, never island modules
  — the shell must keep working when island scripts fail. shell.js
  owns: the delegated `data-dial` handler (guarded: hidden
  `#phone-view` → toast + focus `#ext`), the live-call badge
  (`wp:calls-changed` → `#call-badge`), the htmx error-surfacing
  listener, the live-transcript paging guard, the nav language
  re-fetch on `wp:lang-changed`, the tab skeleton reveal/hide on
  navigating swaps, the morph focus-to-heading move, and the
  swap-time `aria-current` mirror.
- **CSP**: same-origin only, `connect-src wss:` for SIP; no CDN, no
  webfonts, no inline handlers, NO inline `style` attributes (they
  are silently dead — docs/lessons.md) and NO inline scripts at all
  (`TestServedPageSatisfiesStrictCSP` fails on any; the theme preload
  is the same-origin `/assets/theme-preload.js`, and templ-components
  Base is told `NoThemeScript` — a dependency bump can never change
  served script bytes).
- **templ-components adoption**: `layout.Base` adopted (layout.templ,
  with `NoThemeScript` + `CSSPath`/`HTMXVersion` suppressed via
  props); `display.EmptyState` adopted (six true empty-state sites)
  with a PERMANENT scoped Tailwind v4 build at `/assets/tw.css`
  (`@source` of exactly the adopted components from the module cache —
  rebuild with `nix run nixpkgs#tailwindcss_4`, NEVER
  nixpkgs#tailwindcss which is v3; coexistence PROVEN — Tailwind emits
  `@layer` only, unlayered app.css wins every collision). DELIBERATE
  custom hand-rolls stay: avatars (`avatarFor`/`avatarHue`, hue-class
  CSP workaround), nav badges (`wp-nav-badge`), the three
  informational `wp-empty` occurrences, error panel (error.templ),
  timestamps (`formatClock`/`formatStamp`), brand SVG.
  `display.RelativeTime`, `display.CountBadge`, and the errorpage
  module are REJECTED with rationale in the coexistence verdict doc.
- `window.PBX_CONFIG` (`/config.js`): keys `sipDomain`,
  `websocketPath`, `iceServers`, `phoneApi`, `contacts`, `crm`.

## Failure → feedback

The full table lives in
[docs/error-contract.md](docs/error-contract.md) (cross-documented
with the stack runbook § "Webphone error contract" — keep both sides
in sync). Rules: every error path lands in at least one VISIBLE
surface; shell copy stays ENGLISH (decision D3, 2026-09-20) while
island copy is en/de. BDD posture: Ginkgo where the subject is a state
machine (session behavior suites), table-driven Go tests where
clearer, island `node:test` black-box specs; no Ginkgo ports of
already-pinned paths.

## Hard-won rules (stories + evidence: docs/lessons.md)

- `.templ` files must NOT import `github.com/a-h/templ`; bare `if x {`
  conditionals, not `@if`. templ values that carry JSON build with
  `templ.JSONString` (CSRF helpers double-escape otherwise).
- go-branded-id: `id.ID.String()` renders `"Brand:value"`; the domain
  `mustID` parsers strip an optional `"Prefix:"`.
- DTMF must be `application/dtmf-relay` with `Signal=<d>` (equals
  form) — the colon form is 200-OK'd by FreeSWITCH and silently
  dropped. FreeSWITCH executes transfers server-side on the REFER;
  the browser only sends it and parses the NOTIFY sipfrag verdict.
- sip.js pinned at 0.21.2; the bounded watchdog in `connection.js` is
  load-bearing (0.x hangs in `userAgent.reconnect()`; also rebuilds
  on registration loss). JsSIP 3.13.8 is the NAMED fallback — swap
  only on named triggers. All Go-side telephony REJECTED. sip.js
  fires NO stateChange on re-register — recovery paths set UI state
  explicitly. Mic pre-warm seam (`mic.js`): evidence in docs/lessons.md.
- Env config nests with `__`: `WEBPHONE_GATEWAY__MODE` →
  `gateway.mode`; single underscores stay literal. Scalars via env;
  lists (`ice_servers`, `contacts`) + the `identities` map via the
  JSON file.
- **CRM integration seam** (Ledger `~/projects/crm`): optional, OFF by
  default — `crm.url`+`crm.token` (both or neither, validated) build
  `internal/crm.Resolver` (Client + TTL cache: positive 6h, negative
  5min, cap 1024, transport failures NOT cached) injected as
  `Deps.CRM`; nil-safe everywhere. Enrichment is read-only display:
  panels collect page numbers → `crmNames` → `Names map[string]string`
  in view props → `displayName` falls back to the raw number (debug
  log only, NOT in the failure-feedback table). Number matching lives
  in the CRM; webphone sends numbers verbatim. Call logging: island
  `recordCrmCall` (panels.js, gated on `PBX_CONFIG.crm`, fire-and-forget
  beside `recordHistory`) → `POST /api/calls` → resolve → `LogCall` →
  204 (unknown numbers 204, never mint contacts; CRM outage 502, island
  toasts `crmLogFailed`). The CRM side mounts `/api/*` only with
  `-api-token` (bearer, constant-time, no CSRF) and `phones` on
  contacts (comma/semicolon-split ONLY). `crm.Client` uses the SAME
  `do()` chokepoint as `pbx.Client`; `Resolver` single-flights misses
  and counts outcomes (`LookupCounters`) → `/metrics` renders
  `webphone_crm_lookups_total` only when `CRM.Enabled()`. The call
  journal is idempotent (`callsIdem`, extension-namespaced `key`, 1h
  TTL; replay = inert 204; absent key = legacy never-dedupe).
- **Paperless seam**: optional, OFF by default — `paperless.url`+
  `paperless.token` both-or-neither build `internal/paperless.Archiver`
  injected as `fax.New`'s archiver; nil-safe everywhere. Inbound faxes
  only, fire-and-forget after persist+notify; metadata ids lazily
  ensured per first SUCCESS; duplicate refusal = inert success; blob
  store stays the only storage truth. Evidence: docs/lessons.md.
- erraudit honors `//nolint:erraudit // reason`; branching-flow
  honors NO nolint (documented skip in `.buildflow.yml`; same for
  go-structure-linter, cqrs-lint, nix-hash-fix). **The erraudit bar**:
  tier 1 enforced (`--type-aware --disable-extensions`, must exit 0);
  tier 2 enforced green (`--enforce-go-error-family` AND
  `--enforce-coded-errors`, both 0); tier 3 owner-only full audit
  (never gates). NEW error paths MUST follow the convention:
  errorfamily.New*/Wrap* + a stable `<seam>.<op>` code, never bare
  fmt.Errorf, and never a family-fixed wrap over a polymorphic cause.
  Per-seam family pins live in each package's family_test.go.
  Re-measure tiers 1+2 monthly (next: 2026-10-22) — tier-2 must STAY
  0. Rules + evidence: docs/lessons.md.
- `pbx.Client` owns the timeout-bounded HTTP client; the
  `/phone-api` proxy rides `PhoneAPI.HTTPClient()`, never
  `http.DefaultClient`. Join path and query separately
  (`url.JoinPath` percent-encodes `?`).
- Formatting: treefmt/prettier owns `internal/web/assets/island/**` +
  `shell.js` + `*.css`; BuildFlow's oxfmt owns everything else Go
  AND `internal/web/assets/island-tests/*.mjs` (prettier does NOT
  claim island-tests — no two-formatter war). Markdown is NOT in
  treefmt scope; `*_templ.go` and `vendor/` are excluded everywhere.
  `.templ` SOURCES are deliberately formatter-unowned (nix/treefmt.nix
  scope excludes them).
- Island no-undef gate: `nix flake check` runs `island-lint` (oxlint,
  all categories off, `no-undef` on, `SIP` declared readonly). New
  browser globals go in the config's `globals` block; the check fails
  closed and records the scanned file list.
- UI token system: the `:root`/dark token blocks are MIRRORED between
  app.css and island/style.css — change both. `.sr-only` is OWNED by
  app.css. Avatars: `avatarFor`/`avatarHue` (helpers.go, WITH tests).
  SSE payloads keep the greppable row classes (`wp-thread-row`,
  `wp-bubble`, `wp-fax-row`). The green dot is `#wp-sse-live`
  (JS-created).
- Typography craft: `font-size: 93.75%` on `html` (PERCENTAGE, never
  px), rendering baseline on the html rule, `text-wrap: balance` on
  headings, `text-wrap: pretty` on `.wp-bubble-body`, local mono stack
  for `#log`/`.ice`. `html lang` follows the session (layout.templ
  passes `Locale: string(props.Lang)` to `layout.Base`). Evidence:
  docs/lessons.md.
- CSRF: login/logout rotate the token; the island adopts the fresh
  one WITHOUT reload via `GET /api/csrf` (retry ladder: recover on
  retry 2, reload only after 3 failures). Tests that POST after
  logging in must use the client `login` helper. Behind TLS proxies:
  `csrf.trusted_*` (full story: docs/lessons.md).
- Island tests: `window.location.reload` must be stubbed as
  `globalThis.window.location = {...}`; stubs that lack DOM methods
  (e.g. `replaceChildren`) make renders THROW silently — assertions
  must cover render OUTCOMES, not just absence of errors.
- **Island honesty contract**: the UI never shows a state the network
  hasn't confirmed. Hold is a pending-state machine —
  `entry.holdPending` ("holding"/"resuming") renders the pulsing chip +
  disabled button until the re-INVITE settles; a failed toggle returns
  the card to the SETTLED state (no optimistic flip); a toggle arriving
  mid-flight queues (`holdQueued`). Offline truth: `#offline-banner`
  (DOM-contract id, role=status, data-i18n) flips exactly with the
  registration pill — ON at `connect()` until the REGISTER lands, ON
  for transport loss / rejected registration, ON for the browser
  `offline` event; `online` only nudges `connection.networkOnline()`
  and never claims registered.
- **Shell & accessibility contract**: the shell owns the command
  palette (Ctrl/Cmd-K), the "?" shortcut help + Settings cheat-sheet,
  the skip-to-content link, the `#wp-tab-skeleton` reveal + panel
  transition, the morph focus-to-heading move, the `#wp-live`
  live-region announcements (both new DOM-contract ids), the fixed
  bottom tab bar (mobile), `aria-current` on the nav, and the
  optimistic send bubble with draft-restore-on-failure. OPERATOR
  RULING (2026-10-01): contacts depth is LEDGER's domain
  (~/projects/crm) — this app does NOT grow a contacts manager. The
  served-markup change owes a fresh stack browser E2E.
- Stack browser E2E: budget 445s; a ~90s transfer-step death after
  green registration+DTMF+ICE is a known flake mode (re-run once
  before digging).

## Release runbook

The full v2.x dance (fold → bump → gates → tag → stack bump → stack
gates → aarch64 → closing sweep) plus the pbx-artmann relock ritual
live in [docs/release-runbook.md](docs/release-runbook.md). The
auto-commit daemon commits AND pushes continuously — work in small
explicitly-committed units, leave a narrative commit at every phase
boundary, and verify end states with `git ls-remote`.

## Concurrent sessions

More than one Crush session can work this repo at once. Tell-tale:
uncommitted files you did not author and mid-edit compile failures
that heal on re-run. Rules: never revert/"fix" their in-flight files;
re-read any shared file (i18n.go, pages.go, flake.nix) immediately
before editing; a full-suite gate run may catch THEIR transient
breakage — attribute failures before acting; and leave their booted
dev servers running. The auto-commit/treefmt daemon is adversarial to
in-flight edits: it can reformat a file between View and Edit
(silently discarding the edit), so re-View immediately before each
Edit or write the file atomically.

## Buildflow health warning

"9 tools unavailable (health check failed)" is NOISE here: all nine
are JS/TS or Python steps that land "not applicable". Real gaps,
fixed: go-licenses and codespell are in the devShell now — run
buildflow inside `nix develop` (or `scripts/buildflow.sh`). On-demand
scanners (`buildflow -s gitleaks`, `-s codespell`) need the REAL
binary in the shell: BuildFlow's built-in fallback spellchecker does
NOT read `.codespellrc`. markdown-lint runs detect-only by posture:
never "fix" the corpus by reflowing. KNOWN TOOL BUG — gomod-check's
vendor-consistency rule reports ~54 findings on this repo: verified
FALSE POSITIVE 2026-10-01 (`go mod vendor` regenerates modules.txt
byte-identically; `go build -mod=vendor` green; the marker heuristic
disagrees with the toolchain). Do NOT hand-edit vendor markers — the
findings gate tripping on gomod-check ALONE is this known deviation
(fix belongs upstream in BuildFlow).

## Conventions

- One home per fact: README sells + documents contracts, FEATURES
  inventories status, TODO_LIST holds open work, CHANGELOG logs
  history, docs/lessons.md holds war stories, this file keeps the
  rules.
- Cite stable names (ids, function names, option names), not
  `file:line`.
- Behavior parity rules ports: port logic verbatim first, refactor in
  a second, separately-verified change.
- An auto-commit daemon commits continuously AND pushes; never revert
  changes you did not author, and verify end states with
  `git ls-remote`, not push logs.
- Recording is two-level: the PBX stack records every dialled call
  server-side (`record_session`, stereo WAV under the stack's
  `/recordings/`, operator basic-auth; `*97<ext>` skips) — this
  repo's island and server have NO recording capability, only CDR
  history rows. Product-level capability questions get answered per
  level (island / Go server / consuming stack).
