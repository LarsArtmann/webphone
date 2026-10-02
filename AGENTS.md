# AGENTS.md

Enduring context for AI sessions. Rules live here; war stories, evidence, and train chronology: [docs/lessons.md](docs/lessons.md). Error table: [docs/error-contract.md](docs/error-contract.md). DOM contract: [docs/dom-contract.md](docs/dom-contract.md). Dedup rulings: [docs/dedup-registry.md](docs/dedup-registry.md). Release dance: [docs/release-runbook.md](docs/release-runbook.md).
Capped at 377 lines by BuildFlow's `docs/agents-md-size` preflight — move evidence to docs/, keep rules here.

## What this is

A single-binary Go unified-communications web app: the proven SIP.js call island plus server-rendered tabs (Messages SMS/MMS, Fax, Voicemail, History, Contacts, Settings) on one page. Built on cqrs-htmx (root library only) + templ-components `layout.Base` + SQLite (modernc), wired through a samber/do v2 composition root (`internal/app`) with go-health probes and an optional go-health-dashboard at `/health`. The consuming stack ([nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)) imports `nixosModules.default`, fronts the binary with TLS + the WSS `/sip` proxy, and RIDES webphone `main` (per-train lock bump); pbx-artmann consumes the stack via a rev pin. Version: `flake.nix` (`webphoneVersion`).
The cqrs-htmx `setup` bundle stays REJECTED (split-brain identity + a measured +68% binary-size NO-GO); the shell VALUE landed via `httputil.NewServer`. Evidence: docs/lessons.md.

## Tri-repo integration rules

- PUSH webphone before the stack re-pins; the stack tree must be CLEAN
  before pbx-artmann re-locks (its narHash covers the whole tree). Ritual:
  [docs/release-runbook.md](docs/release-runbook.md).
- NixOS module options: `enable`, `package`, `dataDir` (MUST be under `/var/lib/` — assertion), `settings` (freeform), `environmentFile`, `memoryMax`, `csrf.{trustedProxies,trustedOrigins}`, `serverTiming.enable`, `backup.{enable,destDir,calendar,retentionDays}` (null = single snapshot; set = dated history + prune), `caddy.{enable,hostName,sipUpstream}`, `caddy.hsts.{enable,maxAge}`. The front is CADDY (the old nginx generator was never used).
  The vhost reverse-proxies the app, bridges the websocket path ONLY when `caddy.sipUpstream` is set (the Go app never terminates the SIP wss), flushes `/events` unbuffered; probes fence via `remote_ip` matchers. `backup.destDir` shares the `/var/lib/` assertion; both units set `UMask=0077` + `StateDirectoryMode=0750`; the backup oneshot orders after `webphone.service`.
- The `webphone-module` flake check evaluates the module with stand-in
  options (new config keys the module writes need a stand-in there).
  `checks.x86_64-linux.webphone-backup` (KVM-gated) +
  `webphone-backup-drill` cover snapshot AND restore.
- flake.nix is a slim ENTRY; the meat lives in `nix/packages.nix`,
  `nix/checks.nix`, `nix/module-check.nix` (+ base/csrf/backup case files),
  `nix/vm-tests.nix`, `nix/apps.nix`, `nix/devshell.nix`, `nix/treefmt.nix`.
  The `webphoneVersion` let-binding MUST stay in flake.nix (`release.sh`
  greps/seds it) and reaches `nix/packages.nix` via
  `{ _module.args.webphoneVersion = ...; }`. Gotcha: `self` is a
  TOP-level-only flake-parts module arg (declare it on the module function).
- webphone's gateway seam (loopback vs webhook) is consumed by
  pbx-artmann's `telnyx-webhooks.py` bridge — contracts in the plan docs
  under `docs/planning/archived/2026-09-19_11-51_SUPERB-*`.

## Owner decisions (2026-09-20)

- Stack input policy: ride webphone `main` with a per-train lock bump.
  pbx-artmann keeps its pin type while all three repos live on this host.
- Sanitization: the island keeps letters (`[^\d+*#a-zA-Z]`), matching
  `sanitizeDialable`; pinned on the served asset.
- Own-number feed: static config map (`identities`); a stack `/phone-api`
  identity endpoint is the upgrade path. CDR-derive REJECTED (outbound
  `caller_id_number` is dialplan-dependent and absent before the first call).

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

Host-nix-down fallback: run gates with STORE toolchains directly —
`GOTOOLCHAIN=local /nix/store/*-go-1.27*/bin/go test -count=1 ./...`, the
store nodejs for island tests, and `-ldflags "-X
github.com/larsartmann/webphone/internal/server.buildVersion=vX.Y.Z"` for a
version-stamped smoke binary (NOT `main.displayVersion`). Story: lessons.md.

## The DOM + bundle contract (DO NOT BREAK CASUALLY)

The id list lives in [docs/dom-contract.md](docs/dom-contract.md) — the
SINGLE source, parsed by `TestServedPageHoldsTheDomContract`. The stack's
browser E2E (`tests/browser-e2e.py` there) drives the island remotely —
re-run it after any markup change (release-runbook obligation).

- Island modules under `internal/web/assets/island/app/` are served VERBATIM
  (no bundling, no minification) — keep it that way (the E2E greps strings).
- The event log (`#log`) stays **English** in both UI languages.
- Unknown paths render the STYLED 404 (`notFoundPage`; status stays 404).
- The island never unloads: tab nav swaps partials into `#tab-content`; the
  SIP island lives OUTSIDE that region. Deep links render the full shell.

## Architecture invariants

- **Middleware chain** (server `New`, ORDER): `ContextEnrichmentMiddleware
  (nil)` → `RequestLoggingSlog` → `ServerTimingMiddlewareWhen` (env-gated) →
  `SecurityHeaders` → `cqrshtmx.RecoveryMiddleware` → routes. Enrichment
  stays OUTSIDE the request log (`request_id=` + `X-Request-ID`); user
  extractor nil. `securityHeadersConfig()` is the single header-config
  source; `Permissions-Policy` calibrated (`microphone=(self)`; others
  denied). Limiters are `httputil.KeyedRateLimiter` with port-stripped
  peer-host keys (`remoteHostKey`; flip to `KeyExtractorFromClientIP` only
  once the stack proves XFF sanitization). `/healthz` = honest readiness
  (sqlite ping + blob write via the ONE home `blob.ProbeWrite`; 503 names
  the failing check); `/livez` = fetch-free liveness; `/startupz` = latched
  503-until-first-pass (served from `Deps.Probe` over `sqlite`/`blob-dir`;
  nil Probe = NewChecks fallback for test Deps). `/events` rides
  `Broadcaster.ServeSSE`; pinned by httputil's httpspec suite. The unit
  stays `Type=simple` — README "Readiness vs systemd".
- **Composition root `internal/app`** (samber/do v2): ONE container owns
  object lifetime; `cmd/webphone` owns process concerns. The injector lives
  ONLY in `internal/app` (service packages stay framework-free; lifecycle
  conformance asserted adapter-side in app.go). `MustInvoke*` stays
  composition-root-ONLY by doctrine (fail-fast miswires, DO-1; arch-test
  enforced); `App.Start` resolves via `InvokeNamed`+`wrapf` (de-panicked
  2026-10-02). Critical pair registers
  NAMED (`sqlite`, `blob-dir` — the probe's critical-service contract);
  both + the handler are EAGERLY invoked in `New`; the probe is built ONCE
  and threaded to server.Deps + the dashboard (never registered in the
  injector it reads — self-registration recurses); shutdown = HTTP drain →
  `Probe.Shutdown` → `do.Shutdown` (idempotent). `Deps.Probe`/
  `Deps.Dashboard` nil-fallback keeps existing test compositions working.
  `samber-linter` HW-4 on the pair is suppressed (eagerly resolved via
  `MustInvokeNamed`; the linter can't see transitive resolution).
- **Health dashboard seam** (go-health-dashboard v0.10.1): config-gated
  `dashboard.{enable,title}`, DEFAULT OFF — fence like the probe triple; the
  Caddy vhost proxies `/health/*` unbuffered. Mounted at `/health` (subtree
  patterns `/health` AND `/health/`); aliases
  `/health/{livez,readyz,startupz}` (SAME probe instance). CSP: subtree
  override via `httputil.Nonce` + `dashboard.RecommendedCSP` — `unsafe-eval`
  scoped to `/health` ONLY (Datastar compiles expressions); elsewhere strict
  (`TestAppServesMainPageUnderStrictCSP` pins it). Own scoped stylesheet
  `/assets/health.css` (rebuild with `nix run nixpkgs#tailwindcss_4`, NEVER
  v3; never merges with `tw.css`). Probe refresh 1s ONLY while enabled; off
  = 0. The smoke boots WITH the dashboard on. Evidence: lessons.md.
- **Module graph acyclic**: island `state.js` + `auth.js` exist so
  calls/ice/connection never import each other; ice syncs via
  `wp:calls-changed` — enforced by `internal/arch/arch_test.go` (also
  `domain` imports nothing internal; services never import `server`/`web`).
- **Sessions**: verify extension/password against the PBX directory before
  minting (fail-closed; loopback dev skips + WARNs). Store SEAM: prod
  `NewSQLiteStore` (survives restarts; row carries the extension + directory
  password the `/phone-api` proxy needs — at rest accepted, swept on
  read/Create), tests/loopback `NewMemStore`. Sessions SLIDE:
  `GET /api/session` resumes a live cookie at boot (no-store,
  requireSession-gated); activity past the idle halfway renews
  (`session_ttl`=7d idle, `session_max_ttl`=30d absolute). Login/hooks
  per-IP rate limited. Handlers self-gate via `requireSession`;
  `Sessions.Require` also wires `/events` + `/phone-api/`. The contract test
  allowlists exactly three 401 writers: actions.go, webhooks.go,
  session_api.go.
- **Language**: per extension — `wp-lang` cookie (samesite=strict) →
  `Accept-Language: de*` → English. `ExtensionHubs` remember the negotiated
  lang for SSE fragments. Service validation reasons stay English. i18n
  dictionaries in `views/i18n.go`; add new keys to BOTH maps (test enforces
  parity; unknown keys surface themselves, deliberately).
- **Live-update surfaces morph-swap**: thread list, `#thread-transcript`,
  fax list, contacts panel, voicemail re-fetch, `refreshNav` carry
  `hx-swap="morph:innerHTML"` (idiomorph via `/htmx-ext.js`, ONE bundle).
  Morph preserves focus/drafts/listeners; STATEFUL nodes need stable ids.
  Payloads are bare fragments. Events: `threads`, `thread`, `fax`,
  `voicemail`, `contacts` (the last two are payload-less NUDGES).
  The `threads` nudge is also LIST-ONLY by design: `MessagesChanged`
  with a zero thread id skips the transcript push (notifier.go guard —
  an empty `thread` payload would wipe the open conversation).
- **Thread organization + snippets seam (T18 M21/M22, 2026-10-02)**:
  `pinned/archived/muted` are three 0/1 thread columns (schema_version
  v2; migrations are VERSIONED in `store/db.go` — a new schema change
  is a new migration step, never an ad-hoc ALTER). Toggles ride
  `POST /messages/{id}/{flag}?on=N` posting the DESIRED state: row
  buttons answer 204 + the list-only `threads` nudge; head buttons
  target `#wp-thread-head` outerHTML and get the freshly rendered head
  back (setThreadFlag branches on the `HX-Target` header). Inbound
  messages AUTO-UNARCHIVE (AppendMessage upsert CASE); mute is
  presentation-only (`countUnread` skips muted rows — unread truth
  stays in the store). Thread rows carry stable `thread-<id>` wrapper
  ids (idiomorph). Reply snippets: per-extension store (cap 100,
  replace-by-id upsert), Settings CRUD actions, quick snippets (cap 5)
  render as chips + every snippet behind a `<details>` picker in the
  REPLY composer only; shell.js `data-snippet` fill REPLACES the
  textarea (data-sms precedent). Image attachments carry
  `data-lightbox`; shell.js opens the singleton `#wp-lightbox` dialog
  (native ESC/backdrop close). Design decisions D1–D14:
  `docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md`.
- **Gateway seam**: loopback (dev) vs webhook (multipart to
  `{url}/message|/fax`, Bearer secret, `{"provider_ref"}` receipt). File
  parts carry their HONEST Content-Type (`createFilePart`; fax parts
  `application/pdf`) — the Telnyx bridge prefers it, sniffing only as the
  octet-stream fallback for pre-`e6ea2c7` binaries. Never write version
  claims for unreleased code — pin by date or commit. Part-header block
  byte-pinned by `TestProviderFormPartHeaderBlockGolden`; a webhook-lane
  smoke probe stays DECIDED-AGAINST. Self-sends to the owner's DID (config
  `identities`) never reach the provider: `gateway.SelfSendRejection`
  refuses locally after persist, riding the 422 arm. `/hooks/*` share the
  secret and fail CLOSED (503) when none is configured. Status hooks
  idempotent (`hooksIdem`, 1h TTL; only successes recorded; replay 202).
- **Owner scoping everywhere**: every store query is extension-scoped;
  attachments/faxes stream through session-gated handlers only.
- **One-home helpers**: `listRows[T]` (store/db.go) owns the query→close→
  scan→`rows.Err()` lifecycle; `pbx.do()` is the single disabled-policy
  chokepoint; `session.makeSession` owns the birth invariant;
  `server.requireMultipartTo` is the send-form prologue; `domain.must`,
  `domain.OrClock`, `store.updatedOrNotFound`, `views.formatFor`,
  `server.crmNumbers`, `server.applyStatusWebhook`, `server.recordCallIdem`,
  `server.contactSaveFailed`, `server.apiContactSaved`, `views.panelHead`,
  `views.errorBanner`, `views.panelError`, `views.identityLine` — each with
  its own micro-test (several byte-exact in `views/panels_test.go`).
  **Dedup registry**: [docs/dedup-registry.md](docs/dedup-registry.md) is
  the ONE home for every clone ruling — read it BEFORE re-litigating an
  accepted similarity; append a sweep-log line per run. `-t 3` is the
  working baseline pending owner ratification.
- **Personal contacts have ONE home**: the per-extension SQLite store via
  `/api/contacts` (session-gated, 60/min POST limiter, 500/extension cap).
  Mutations answer 204; the LIST is the only id source (store upserts by
  (owner, phone), keeping the old id on rename). Legacy
  `localStorage["pbx-contacts"]` imports once post-login and is REMOVED
  only after the server accepted every row; load trigger is session.js's
  `wp:session-opened` AFTER cookie mint + CSRF adoption. History stays
  hybrid BY DESIGN (local session log + same CDR API).
- **Tab→island affordances live in shell.js**, never island modules (the
  shell must work when island scripts fail). shell.js owns the delegated
  `data-dial` handler (hidden `#phone-view` → toast + focus `#ext`), the
  live-call badge (`wp:calls-changed` → `#call-badge`), the htmx
  error-surfacing listener, the live-transcript paging guard, the nav
  language re-fetch (`wp:lang-changed`), the tab skeleton reveal/hide, the
  morph focus-to-heading move, and the swap-time `aria-current` mirror.
- **CSP**: same-origin only, `connect-src wss:` for SIP; no CDN, no
  webfonts, no inline handlers, NO inline `style` (silently dead —
  lessons.md) and NO inline scripts (`TestServedPageSatisfiesStrictCSP`
  fails on any; theme preload is same-origin `/assets/theme-preload.js`;
  templ-components Base told `NoThemeScript`).
- **templ-components adoption**: `layout.Base` (with `NoThemeScript` +
  `CSSPath`/`HTMXVersion` suppressed); `display.EmptyState` (six true
  empty-state sites) with a PERMANENT scoped Tailwind v4 build at
  `/assets/tw.css` (`@source` of exactly the adopted components; rebuild
  with `nix run nixpkgs#tailwindcss_4`, NEVER v3; coexistence PROVEN —
  Tailwind emits `@layer` only, unlayered app.css wins). DELIBERATE custom
  hand-rolls stay: avatars (`avatarFor`/`avatarHue`), nav badges
  (`wp-nav-badge`), three informational `wp-empty`, error panel
  (error.templ), timestamps (`formatClock`/`formatStamp`), brand SVG.
  `display.RelativeTime`, `display.CountBadge`, the errorpage module are
  REJECTED (rationale in the coexistence verdict doc).
- `window.PBX_CONFIG` (`/config.js`): `sipDomain`, `websocketPath`,
  `iceServers`, `phoneApi`, `contacts`, `crm`.

## Failure → feedback

Full table: [docs/error-contract.md](docs/error-contract.md)
(cross-documented with the stack runbook § "Webphone error contract" — keep
both in sync). Every error path lands in at least one VISIBLE surface;
shell copy stays ENGLISH (D3, 2026-09-20), island copy en/de. Boot
failures render the 5-part operator contract on the journal
(`cmd/webphone/bootreport.go`; exit 1 designed / 2 panic, 2026-10-02):
copy + pins in error-contract.md § "Boot surface". BDD posture:
Ginkgo for state machines (session suites), table-driven Go tests where
clearer, island `node:test` specs; no Ginkgo ports of pinned paths.

## Hard-won rules (stories + evidence: docs/lessons.md)

- `.templ` files must NOT import `github.com/a-h/templ`; bare `if x {`, not
  `@if`. JSON-carrying templ values build with `templ.JSONString` (CSRF
  helpers double-escape otherwise).
- go-branded-id: `id.ID.String()` renders `"Brand:value"`; domain `mustID`
  parsers strip an optional `"Prefix:"`.
- DTMF must be `application/dtmf-relay` with `Signal=<d>` (equals form) —
  the colon form is 200-OK'd by FreeSWITCH and silently dropped. FreeSWITCH
  executes transfers server-side on the REFER; the browser sends it and
  parses the NOTIFY sipfrag verdict.
- sip.js pinned 0.21.2; the bounded watchdog in `connection.js` is
  load-bearing (0.x hangs in `userAgent.reconnect()`; rebuilds on
  registration loss). JsSIP 3.13.8 is the NAMED fallback — swap only on
  named triggers. All Go-side telephony REJECTED. sip.js fires NO
  stateChange on re-register — recovery paths set UI state explicitly. Mic
  pre-warm seam (`mic.js`): evidence in docs/lessons.md.
- Env config nests with `__` (`WEBPHONE_GATEWAY__MODE` → `gateway.mode`);
  single underscores stay literal. Scalars via env; lists (`ice_servers`,
  `contacts`) + the `identities` map via the JSON file.
- **CRM seam** (Ledger `~/projects/crm`): optional, OFF by default —
  `crm.url`+`crm.token` (both or neither) build `internal/crm.Resolver`
  (Client + TTL cache: positive 6h, negative 5min, cap 1024, transport
  failures NOT cached) injected as `Deps.CRM`; nil-safe everywhere.
  Enrichment read-only: panels → `crmNames` → `Names` in view props →
  `displayName` falls back to the raw number. Number matching lives in the
  CRM; webphone sends numbers verbatim. Call logging: island `recordCrmCall`
  (gated on `PBX_CONFIG.crm`, fire-and-forget) → `POST /api/calls` → 204
  (unknown numbers 204, no contacts minted; outage 502, island toasts
  `crmLogFailed`). CRM side mounts `/api/*` only with `-api-token` (bearer,
  constant-time, no CSRF). `crm.Client` uses the SAME `do()` chokepoint as
  `pbx.Client`; `Resolver` single-flights misses, `/metrics` renders
  `webphone_crm_lookups_total` only when enabled. Call journal idempotent
  (`callsIdem`, extension-namespaced `key`, 1h TTL; replay inert 204).
- **Paperless seam**: optional, OFF — `paperless.url`+`paperless.token`
  both-or-neither build `internal/paperless.Archiver` as `fax.New`'s
  archiver; nil-safe. Inbound faxes only, fire-and-forget after
  persist+notify; metadata ids lazily ensured per first SUCCESS; duplicate
  refusal = inert success; blob store stays the only storage truth.
- erraudit honors `//nolint:erraudit // reason`; branching-flow honors NO
  nolint (documented `.buildflow.yml` skip; same for go-structure-linter,
  cqrs-lint, nix-hash-fix). The erraudit bar: tier 1 enforced (`--type-aware
  --disable-extensions`, exit 0); tier 2 enforced green
  (`--enforce-go-error-family` AND `--enforce-coded-errors`, both 0); tier 3
  owner-only (never gates). NEW error paths MUST use errorfamily.New*/
  Wrap* + a stable `<seam>.<op>` code, never bare fmt.Errorf, never a
  family-fixed wrap over a polymorphic cause. Per-seam pins in each
  package's family_test.go. Re-measure tiers 1+2 monthly (next 2026-10-22);
  tier-2 must STAY 0. Evidence: lessons.md.
- `pbx.Client` owns the timeout-bounded HTTP client; the `/phone-api` proxy
  rides `PhoneAPI.HTTPClient()`, never `http.DefaultClient`. Join path and
  query separately (`url.JoinPath` percent-encodes `?`).
- Formatting: treefmt/prettier owns `internal/web/assets/island/**` +
  `shell.js` + `*.css`; oxfmt (BuildFlow) owns everything else Go AND
  `internal/web/assets/island-tests/*.mjs`. Markdown NOT in treefmt scope;
  `*_templ.go` and `vendor/` excluded everywhere; `.templ` SOURCES
  deliberately formatter-unowned. After ANY island/shell/css edit run
  `nix fmt` BEFORE the gates — buildflow's formatter skips island files,
  and drift has failed full gate runs before (2026-10-01).
- Island no-undef gate: `nix flake check` runs `island-lint` (oxlint, all
  categories off, `no-undef` on, `SIP` readonly). New browser globals go in
  the config's `globals` block; fails closed and records the file list.
- UI token system: `:root`/dark token blocks MIRRORED between app.css and
  island/style.css — change both. `.sr-only` OWNED by app.css. Avatars:
  `avatarFor`/`avatarHue` (helpers.go, WITH tests). SSE payloads keep the
  greppable row classes (`wp-thread-row`, `wp-bubble`, `wp-fax-row`). The
  green dot is `#wp-sse-live` (JS-created).
- Typography: `font-size: 93.75%` on `html` (PERCENTAGE, never px);
  rendering baseline on the html rule; `text-wrap: balance` on headings,
  `pretty` on `.wp-bubble-body`; local mono stack for `#log`/`.ice`.
  `html lang` follows the session (layout.templ passes
  `Locale: string(props.Lang)`). Evidence: lessons.md.
- CSRF: login/logout rotate the token; the island adopts the fresh one
  WITHOUT reload via `GET /api/csrf` (recover on retry 2, reload only after
  3 failures). Tests that POST after login must use the client `login`
  helper. Behind TLS proxies: `csrf.trusted_*` (lessons.md).
- Island tests: `window.location.reload` must be stubbed as
  `globalThis.window.location = {...}`; stubs lacking DOM methods make
  renders THROW silently — assertions must cover render OUTCOMES.
- **Island honesty contract**: the UI never shows a state the network
  hasn't confirmed. Hold is a pending-state machine (`entry.holdPending`
  "holding"/"resuming" → pulsing chip + disabled button until the re-INVITE
  settles; a failed toggle returns the card to the SETTLED state; a
  mid-flight toggle queues `holdQueued`). Offline truth: `#offline-banner`
  (DOM-contract id, role=status, data-i18n) flips exactly with the
  registration pill; `online` only nudges `connection.networkOnline()`.
- **Feedback/trust (M17, 2026-10-02)**: aria-live politeness — everything
  is POLITE (`#toasts` + `#wp-live` are role=status); the one assertive
  exception is the incoming-call announce
  (`announce(msg, kind, { assertive: true })` → role=alert on the toast
  node). SSE recovery announces once, only after an ANNOUNCED drop
  (3-failure threshold). A failed optimistic send grows Retry + Dismiss
  buttons inside the failed bubble (shell.js §3e, English D3; retry
  re-submits the reply composer, dismiss keeps the draft editable).
  History's empty state is filter-aware (`history.empty` vs
  `history.emptyFiltered`).
- **Shell & accessibility contract**: the shell owns the command palette
  (Ctrl/Cmd-K), the "?" help + Settings cheat-sheet, the skip-to-content
  link, the `#wp-tab-skeleton` reveal + panel transition, the morph
  focus-to-heading move, `#wp-live` announcements, the fixed bottom tab bar
  (mobile), `aria-current` on the nav, and the optimistic send bubble with
  draft-restore-on-failure. OPERATOR RULING (2026-10-01): contacts depth is
  LEDGER's domain — this app does NOT grow a contacts manager.
  Served-markup changes owe a fresh stack browser E2E.
- Stack browser E2E: budget 445s; a ~90s transfer-step death after green
  registration+DTMF+ICE is a known flake mode (re-run once before digging).

## Release runbook

Full v2.x dance + pbx-artmann relock ritual:
[docs/release-runbook.md](docs/release-runbook.md). The auto-commit daemon
commits AND pushes continuously — work in small explicitly-committed units,
leave a narrative commit at every phase boundary, verify end states with
`git ls-remote`.

## Concurrent sessions

More than one Crush session can work this repo at once (tell-tale:
uncommitted files you did not author, mid-edit compile failures that heal on
re-run). Never revert/"fix" their in-flight files; re-read shared files
(i18n.go, pages.go, flake.nix) immediately before editing; attribute
full-suite failures before acting; leave their dev servers running. The
auto-commit/treefmt daemon is adversarial to in-flight edits — re-View
immediately before each Edit or write atomically.

## Buildflow health warning

"9 tools unavailable (health check failed)" is NOISE here (all nine are
JS/TS or Python "not applicable"). go-licenses and codespell are in the
devShell — run buildflow inside `nix develop` (or `scripts/buildflow.sh`).
On-demand `-s gitleaks`/`-s codespell` need the REAL binary (BuildFlow's
fallback spellchecker ignores `.codespellrc`). markdown-lint is detect-only:
never reflow the corpus. KNOWN TOOL BUG — gomod-check's vendor-consistency
rule reports ~54 findings: verified FALSE POSITIVE 2026-10-01 (`go mod
vendor` regenerates modules.txt byte-identically; `go build -mod=vendor`
green). Do NOT hand-edit vendor markers (fix belongs upstream).

## Conventions

- One home per fact: README sells + documents contracts, FEATURES
  inventories status, TODO_LIST holds open work, CHANGELOG logs history,
  docs/lessons.md holds war stories, this file keeps the rules.
- Cite stable names (ids, function names, option names), not `file:line`.
- Behavior parity rules ports: port logic verbatim first, refactor in a
  second, separately-verified change.
- The auto-commit daemon commits AND pushes; never revert changes you did
  not author; verify end states with `git ls-remote`, not push logs.
- Recording is two-level: the PBX stack records every dialled call
  server-side (`record_session`, stereo WAV under the stack's
  `/recordings/`, operator basic-auth; `*97<ext>` skips) — this repo's
  island/server have NO recording capability, only CDR history rows.
  Product-level questions get answered per level (island / server / stack).
