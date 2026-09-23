# AGENTS.md

Enduring context for AI sessions working in this repo. Rules live
here; war stories with evidence live in [docs/lessons.md](docs/lessons.md),
the error-surface table in [docs/error-contract.md](docs/error-contract.md),
the DOM contract in [docs/dom-contract.md](docs/dom-contract.md), the
release dance in [docs/release-runbook.md](docs/release-runbook.md).

## What this is

A single-binary Go unified-communications web app: the proven SIP.js
call island plus server-rendered tabs (Messages SMS/MMS, Fax,
Voicemail, History, Contacts, Settings) on one page. Built on
cqrs-htmx (root library only) + templ-components `layout.Base` +
SQLite (modernc). The consuming stack
([nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony))
imports this repo's `nixosModules.default`, fronts the binary with
TLS + the WSS `/sip` proxy, and RIDES webphone `main` (per-train lock
bump); pbx-artmann consumes the stack via a rev pin. v2.6.0 (signed
tag `807ca0c`) released 2026-09-23; the release TAIL (stack browser
E2E green on the relock, aarch64 re-verify, gh release object,
smoke `--expect-version 2.6.0`, pbx-artmann relock #4) is open on
the TODO_LIST release row — two load-shaped E2E stalls, the armed
chained retry never fired. The stack currently pins webphone
`7503561` (post-tag main).

The cqrs-htmx `setup` bundle is rejected deliberately (split-brain
identity): this product's identity is the PBX extension + directory
password, proven by the island's SIP REGISTER — a second user
database would be a split brain.

## Tri-repo integration rules

- Changes must be PUSHED in webphone before the stack re-pins, and the
  stack tree must be CLEAN before pbx-artmann re-locks (its narHash
  covers the whole tree). Re-lock ritual: [docs/release-runbook.md](docs/release-runbook.md).
- The NixOS module options: `enable`, `package`, `dataDir` (MUST live
  under `/var/lib/` — assertion), `settings` (freeform),
  `environmentFile`, `memoryMax`, `csrf.{trustedProxies,trustedOrigins}`,
  `serverTiming.enable`, `backup.{enable,destDir,calendar,retentionDays}`
  (retentionDays null = single snapshot; set = dated history +
  prune), `nginx.{enable,hostName}`, `nginx.hsts.{enable,maxAge}`.
  Its vhost proxies `/`, the websocket path (3600s), `/events` (SSE)
  and the probe triple as DEDICATED locations.
- The `webphone-module` flake check evaluates the module with
  stand-in options (nginx/systemd/users + `assertions`; new config
  keys the module writes need a stand-in there). Statix pins
  single-assignment style (all `locations` in ONE attrset).
  `checks.x86_64-linux.webphone-backup` (KVM-gated) and
  `webphone-backup-drill` cover backup snapshot AND restore.
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
nix develop                        # Go, templ, golangci-lint, esbuild, … — GOTOOLCHAIN=local; bare `go` works inside. Host go (1.26.7) is below the 1.27.1 floor: OUTSIDE-the-shell go commands need `nix develop -c`
templ generate ./internal/web/views/   # after ANY .templ edit (committed *_templ.go)
nix develop -c go test -count=1 ./...  # -count=1: the result cache has lied during investigations
python3 scripts/webphone-smoke.py          # 40-check live smoke (+4-check restart scenario; boots a fresh binary; --base URL reuses a server; --expect-version X asserts /version)
buildflow                                  # the quality gate; BUILDFLOW_NO_RESULT_CACHE=1 for full (release.sh also gates on `nix run .#vulnix`)
nix run .#vulnix                           # vulnix over the RUNTIME closure; verdict logic = `webphone-vulnix-triage` CLI, fixture-checked
nix flake check                            # package + tests in sandbox + treefmt + island-lint + island-js + kvm-gated backup VM test
nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs   # island JS tests alone (stubs in island-tests/helpers.mjs)
nix build .#webphone --system aarch64-linux   # cross-builds — verify by ELF bytes, never exit code alone (docs/lessons.md)
./update.sh [version]              # repin vendored sip.js (fetch → esbuild IIFE → swap)
```

Smoke a binary quickly (loopback gateway = whole product, zero PBX):

```console
nix develop -c go build -o /tmp/webphone-bin ./cmd/webphone
WEBPHONE_ADDR=127.0.0.1:18099 WEBPHONE_DATA_DIR=/tmp/wp-data \
  WEBPHONE_GATEWAY__WEBHOOK_SECRET=devsecret /tmp/webphone-bin
```

## The DOM + bundle contract (DO NOT BREAK CASUALLY)

The id list lives in [docs/dom-contract.md](docs/dom-contract.md) —
the SINGLE source, parsed by `TestServedPageHoldsTheDomContract`. The
consuming stack's browser E2E (`tests/browser-e2e.py` there) drives
the island remotely — re-run it after any markup change.

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
  (`microphone=(self)`, camera/display-capture/geolocation/payment/
  usb denied). Limiters are `httputil.KeyedRateLimiter` with
  port-stripped peer-host keys (`remoteHostKey`; flip to
  `KeyExtractorFromClientIP` only once the stack proves XFF
  sanitization). `/healthz` = honest readiness (sqlite ping +
  blob-dir write probe, 2s bounds, 503 names the failing check);
  `/livez` = fetch-free liveness; `/startupz` = latched
  503-until-first-pass (go-health `NewChecks`, SAME check functions —
  no second readiness truth). `/events` rides
  `Broadcaster.ServeSSE` (v4.11.0 leads with a `retry:` hint).
  The unit deliberately stays `Type=simple` — see README
  "Readiness vs systemd".
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
  boot (SIP credentials go back to the browser, no-store,
  requireSession-gated); activity past the idle halfway point renews
  (`session_ttl`=7d idle, `session_max_ttl`=30d absolute defaults;
  `normalized()` degrades a missing cap to Max=Idle for test
  determinism). Login/hooks per-IP rate limited. Handlers self-gate
  via `requireSession`; `Sessions.Require` additionally wires
  `/events` + `/phone-api/` — dual-layer by design. The contract
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
  morph surfaces must carry stable ids (idiomorph persists by id).
  Payloads stay bare fragments (no wrappers). Event names: `threads`,
  `thread`, `fax`, `voicemail`, `contacts`. The `voicemail` and
  `contacts` events are payload-less NUDGES (the panel re-fetches
  with per-session credentials). New live surfaces follow the morph
  pattern.
- **Gateway seam**: loopback (dev) vs webhook (multipart to
  `{url}/message|/fax`, Bearer secret, `{"provider_ref"}` receipt).
  Inbound hooks `/hooks/*` share the same secret and fail CLOSED
  (503) when none is configured. Status hooks are idempotent:
  `hooksIdem` (in-memory 1h TTL) dedupes replayed `provider_ref` —
  only successes are recorded (failures stay retryable); replays
  answer `202` inertly. The window only has to cover provider BURST
  retries; status transitions converge, so no persistence.
- **Owner scoping everywhere**: every store query is extension-scoped;
  attachments/faxes stream through session-gated handlers only.
- **One-home helpers from the 2026-09-22 dedup train**: `listRows[T]`
  (store/db.go) owns the query→close→scan→`rows.Err()` lifecycle for
  every list query and wraps both failure shapes with the `op`
  string; `pbx.do()` is the single disabled-policy chokepoint (every
  pbx method fails with `ErrDisabled` there; `VerifyCredentials`
  delegates to `VoicemailSummary`); `session.makeSession` owns the
  session birth invariant (`ExpiresAt = CreatedAt + ttl`);
  `server.requireMultipartTo` is the send-form prologue (session +
  multipart + ParsePhone + 422 with the per-tab key). The 2026-09-22
  late-night dedup pass added: `domain.must` (the one panic-unwrap behind
  every Must parser), `domain.OrClock` (inbound events without a provider
  timestamp get the wall clock), `store.updatedOrNotFound` (zero-rows
  status UPDATE = `ErrNotFound`), `views.formatFor` (the language switch
  behind the timestamp helpers), `server.crmNumbers` (collect a page's
  numbers for CRM resolution, blanks skipped),
  `server.applyStatusWebhook` (the shared status-hook tail: 400 empty
  ref, 202 replay, 404 unknown, 500 retryable, record-on-success),
  `server.recordCallIdem` + `server.contactSaveFailed` (call-log
  idempotency record and the one contact-save 500 text);
  `server.apiContactSaved` (the JSON mutation epilogue: contacts nudge
  + bare 204; the tab handlers share the nudge but answer
  toast + partial). Each carries its own micro-test
  (`TestApplyStatusWebhookContract`, `TestRecordCallIdemContract`,
  `TestContactSaveFailedText`, `TestCRMNumbersSkipsBlanks`,
  `TestMustUnwrapsOrPanics`, `TestOrClockPinsTheZeroFallback`,
  `TestUpdatedOrNotFoundShapes`, `TestFormatForSwitchesAndDefaults` —
  the avatarFor lesson). **Dedup acceptance registry**: the WHY behind
  every accepted/declined clone lives at the in-code comment of its
  site; the sweep-level decisions live in the archived
  `2026-09-23_01-12` + `2026-09-23_03-01` art-dupl reports (and the
  2026-09-18 original) — read those BEFORE re-litigating an accepted
  similarity; `-t 3` is the working baseline pending owner
  ratification.
- **Personal contacts have ONE home**: the per-extension SQLite
  store, read/written via `/api/contacts` (session-gated, 60/min
  POST limiter, 500-per-extension atomic cap). Mutations answer 204;
  the LIST is the only id source (the store upserts by (owner,
  phone), keeping the old id on rename). The legacy
  `localStorage["pbx-contacts"]` list imports once post-login and is
  REMOVED only after the server accepted every row (failed imports
  retry; upsert makes re-import idempotent). Load trigger: session.js
  dispatches `wp:session-opened` AFTER cookie mint + CSRF adoption. History stays hybrid BY DESIGN (local session log + same
  CDR API) — not a split brain, don't "fix" it.
- **Tab→island affordances live in shell.js**, never island modules
  — the shell must keep working when island scripts fail. shell.js
  owns: the delegated `data-dial` handler (guarded: hidden
  `#phone-view` → toast + focus `#ext`, never a silent submit), the
  live-call badge (`wp:calls-changed` → `#call-badge`), the htmx
  error-surfacing listener (§3c: toasts on `htmx:responseError` /
  `htmx:sendError`, skips HX-Trigger responses, throttled to one per
  8s, never auto-reloads — pinned by shell.test.mjs +
  `TestShellJSSurfacesHtmxErrors`), the live-transcript paging
  guard (`htmx:sseBeforeMessage` canceled while
  `data-page != "0"`; `/messages/{id}/read` + `/partials/nav`
  re-fetch after a newest-page push), and the nav language re-fetch
  on `wp:lang-changed`.
- **CSP**: same-origin only, `connect-src wss:` for SIP; no CDN, no
  webfonts, no inline handlers, NO inline `style` attributes (they
  are silently dead — docs/lessons.md) and NO inline scripts at all
  (`TestServedPageSatisfiesStrictCSP` fails on any; the theme preload
  is the same-origin `/assets/theme-preload.js`, and templ-components
  Base is told `NoThemeScript` — a dependency bump can never change
  served script bytes).
- **templ-components adoption** (grep-able table per the library's
  consumer tip): `layout.Base` adopted (layout.templ, with
  `NoThemeScript` + `CSSPath`/`HTMXVersion` suppressed via props);
  everything else is a DELIBERATE custom hand-roll — avatars
  (`avatarFor`/`avatarHue`, hue-class CSP workaround), nav badges
  (`wp-nav-badge`), empty states (`wp-empty`), error panel
  (error.templ), timestamps (`formatClock`/`formatStamp`, byte-stable
  pins), brand SVG. The blocker for further component adoption is
  Tailwind: the library emits Tailwind v4 classes and webphone ships a
  hand-rolled token CSS (app.css, no Tailwind build) — coexistence is
  possible (Tailwind output is layered; unlayered app.css wins
  collisions) but unproven here.
- `window.PBX_CONFIG` (`/config.js`): keys `sipDomain`,
  `websocketPath`, `iceServers`, `phoneApi`, `contacts`.

## Failure → feedback

The full table lives in
[docs/error-contract.md](docs/error-contract.md) (cross-documented
with the stack runbook § "Webphone error contract" — keep both sides
in sync). Rules that live here: every error path lands in at least
one VISIBLE surface; shell copy stays ENGLISH (decision D3,
2026-09-20) while island copy is en/de. BDD posture: Ginkgo where the
subject is a state machine (session behavior suites), table-driven Go
tests where clearer, island `node:test` black-box specs; no Ginkgo
ports of already-pinned paths.

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
  only on named triggers, never speculatively. All Go-side telephony
  REJECTED. sip.js fires NO stateChange on re-register — recovery
  paths set UI state explicitly.
- Env config nests with `__`: `WEBPHONE_GATEWAY__MODE` →
  `gateway.mode`; single underscores stay literal. Scalars via env;
  lists (`ice_servers`, `contacts`) + the `identities` map via the
  JSON file.
- **CRM integration seam (2026-09-22, Ledger `~/projects/crm`)**:
  optional, OFF by default — `crm.url`+`crm.token` (both or neither,
  validated) build `internal/crm.Resolver` (Client + TTL cache:
  positive 6h, negative 5min, cap 1024, transport failures NOT
  cached) injected as `Deps.CRM`; nil-safe everywhere. Enrichment is
  read-only display: panels collect page numbers → `crmNames` →
  `Names map[string]string` in view props → `displayName` falls back
  to the raw number, so a dead CRM never breaks a page (debug log
  only — deliberately NOT in the failure-feedback table). Number
  matching lives in the CRM (single home): digits-only normalize +
  suffix ≥8 with prefix-delta ≤4 + trunk-0 variants; webphone sends
  numbers verbatim. Call logging: island `recordCrmCall` (panels.js,
  gated on `PBX_CONFIG.crm`, fire-and-forget beside `recordHistory`
  in the Terminated branch — never a call-path dependency) → `POST
  /api/calls` (requireSession + CSRF + contacts budget) → resolve →
  `LogCall` → 204; unknown numbers 204 (never mint contacts), CRM
  outage 502 (island toasts, i18n `crmLogFailed`). The CRM side
  mounts `/api/*` only with `-api-token` (bearer, constant-time, no
  CSRF — machine surface) and gained `phones` on contacts
  (comma/semicolon-split ONLY — whitespace is formatting inside a
  number; verbatim at rest). Both sides' tests pin the wire shapes.
  Hardening (2026-09-22 night): `crm.Client` uses the SAME `do()`
  disabled-policy chokepoint as `pbx.Client` (nil-safe, `ErrDisabled`
  everywhere, no URL built when off) — the split brain is closed.
  `Resolver` single-flights concurrent misses (one upstream lookup
  per number; waiters honor their own ctx) and counts upstream
  outcomes (`LookupCounters` hit/miss/failure, upstream round-trips
  only) → `/metrics` renders `webphone_crm_lookups_total{outcome=…}`
  only when `CRM.Enabled()`. The call journal is idempotent: the
  island sends `crypto.randomUUID()` per ended call, `POST /api/calls`
  dedupes on the extension-namespaced `key` (`callsIdem`, own 1h TTL;
  replay = inert 204, 502 stays retryable, the unknown-number drop
  consumes its key too, absent key = legacy never-dedupe).
- erraudit honors `//nolint:erraudit // reason`; branching-flow
  honors NO nolint (documented skip in `.buildflow.yml`; same for
  go-structure-linter, cqrs-lint, nix-hash-fix). **The erraudit
  bar** (2026-09-22): tier 1 enforced (`--type-aware
  --disable-extensions`, must exit 0); tier 2 family-adoption
  tracking (`--enforce-go-error-family`; re-measured 2026-09-22
  evening EARLY per TODO: 127 stdlib_constructor findings total, 113
  outside the crm seam — GREW from 102 via the CRM train (store/
  messaging paths) + the turn_rest validation idioms; top unconverted
  seams: config.go 22, store/messages.go 20, pbx/client.go 8 — the
  family-adoption project must shrink it from 113, and new error
  paths keep following the file-local idiom until their seam
  converts wholesale);
  tier 3 owner-only full audit (never gates). Re-measure tiers 1+2
  monthly (next: 2026-10-22) and update the tier-2 count above.
- `pbx.Client` owns the timeout-bounded HTTP client; the
  `/phone-api` proxy rides `PhoneAPI.HTTPClient()`, never
  `http.DefaultClient`. Join path and query separately
  (`url.JoinPath` percent-encodes `?`).
- Formatting: treefmt/prettier owns `internal/web/assets/island/**`
  - `shell.js` + `*.css`; BuildFlow's oxfmt owns everything else Go
    AND `internal/web/assets/island-tests/*.mjs` (prettier does NOT
    claim island-tests — no two-formatter war). Markdown is NOT in
    treefmt scope; `*_templ.go` and `vendor/` are excluded everywhere.
- Island no-undef gate: `nix flake check` runs `island-lint` (oxlint,
  all categories off, `no-undef` on, `SIP` declared readonly). New
  browser globals go in the config's `globals` block; the check fails
  closed and records the scanned file list.
- UI token system: the `:root`/dark token blocks are MIRRORED between
  app.css and island/style.css — change both. `.sr-only` is OWNED by
  app.css. Avatars: `avatarFor`/`avatarHue` (helpers.go, WITH tests —
  the untested first cut shipped a real bug). SSE payloads keep the
  greppable row classes (`wp-thread-row`, `wp-bubble`, `wp-fax-row`).
  The green dot is `#wp-sse-live` (JS-created so the served DOM
  contract stays untouched).
- CSRF: login/logout rotate the token; the island adopts the fresh
  one WITHOUT reload via `GET /api/csrf` (retry ladder: recover on
  retry 2, reload only after 3 failures). Tests that POST after
  logging in must use the client `login` helper. Behind TLS proxies:
  `csrf.trusted_*` (full story: docs/lessons.md).
- Island tests: `window.location.reload` must be stubbed as
  `globalThis.window.location = {...}` (session.js calls
  `window.location.reload()`); stubs that lack DOM methods (e.g.
  `replaceChildren`) make renders THROW silently — assertions must
  cover render OUTCOMES, not just absence of errors.
- Stack browser E2E: budget 445s (2026-09-22 baseline, two
  forced-rebuild runs 384s/373s); a ~90s transfer-step death after
  green registration+DTMF+ICE is a known flake mode (re-run once
  before digging).

## Release runbook

The full v2.x dance (fold → bump → gates → tag → stack bump → stack
gates → aarch64 → closing sweep) plus the pbx-artmann relock ritual
live in [docs/release-runbook.md](docs/release-runbook.md). The
auto-commit daemon commits AND pushes continuously — work in small
explicitly-committed units, leave a narrative commit at every phase
boundary, and verify end states with `git ls-remote`.

## Concurrent sessions (observed 2026-09-20, re-confirmed 2026-09-22)

More than one Crush session can work this repo at once (2026-09-22:
a CRM-integration session landed `internal/crm/` + view signature
changes mid-flight). Tell-tale: uncommitted files you did not author
and mid-edit compile failures that heal on re-run. Rules: never
revert/"fix" their in-flight files; re-read any shared file
(i18n.go, pages.go, flake.nix) immediately before editing; a
full-suite gate run may catch THEIR transient breakage — attribute
failures before acting (erraudit findings in THEIR new packages are
theirs to land); and leave their booted dev servers running.

## Buildflow health warning (as of 2026-09-19)

"9 tools unavailable (health check failed)" is NOISE here: all nine
are JS/TS or Python steps that land "not applicable". The one real
gap was go-licenses — now in the devShell, so run buildflow inside
`nix develop` (or `scripts/buildflow.sh`). gitleaks/codespell/
markdown-lint run in build mode `full` — `scripts/buildflow.sh`
appends the first two by default.

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
  `git ls-remote`, not push logs — a "local" commit may already be
  public.
- Recording is two-level: the PBX stack records every dialled call
  server-side (`record_session`, stereo WAV under the stack's
  `/recordings/`, operator basic-auth; `*97<ext>` skips) — this
  repo's island and server have NO recording capability, only CDR
  history rows. Product-level capability questions get answered per
  level (island / Go server / consuming stack).
