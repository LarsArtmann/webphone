# AGENTS.md

Enduring context for AI sessions working in this repo.

## What this is

A single-binary Go unified-communications web app (v2.0.0, released
2026-09-19):
the proven SIP.js call island plus server-rendered tabs (Messages
SMS/MMS, Fax, Voicemail, History, Contacts, Settings) on one page.
Extracted 2026-09-17 from `nix-international-telephony` as a static
site (v1), rebuilt 2026-09-18 as this service on cqrs-htmx (root
library only) + templ-components `layout.Base` + SQLite (modernc).
[nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
is the intended consumer: it fronts the binary with TLS and the WSS
`/sip` proxy. DECIDED 2026-09-18: this repo ships `nixosModules.default`
(package/nixos-module.nix) so binary and deployment shape stay in sync;
the stack may import it or keep reverse-proxying — the module is
additive. Stack-side switchover DONE 2026-09-18/19: the stack imports
`nixosModules.default`, its nginx vhost proxies the service, its browser
E2E is green, and its `webphone` input rides webphone `main` — bumped to
the v2.1.0 tag commit `d815004` (stack commit `2289e89`).

The cqrs-htmx `setup` bundle was rejected deliberately: it wires
event-sourced usermgmt users, but this product's identity is the PBX
extension + directory password (proven by the island's SIP REGISTER) —
a second user database would be a split brain.

## Tri-repo integration state (2026-09-19)

- The stack (`nix-international-telephony`) imports
  `nixosModules.default` and RIDES webphone `main` (not tags; pins move
  fast, the runbook's stack-bump step covers releases). pbx-artmann
  consumes the stack via a `path:` input — so changes must be PUSHED in
  webphone before the stack re-pins, and the stack tree must be CLEAN
  before pbx-artmann re-locks (the path narHash covers the whole tree).
- The NixOS module this repo ships: options `enable`, `package`,
  `dataDir` (MUST live under `/var/lib/` — assertion, because systemd
  StateDirectory is derived from it), `settings` (freeform),
  `environmentFile`, `memoryMax` (null = uncapped, wires systemd
  MemoryMax), `nginx.{enable,hostName}`. Its vhost proxies `/`, the
  websocket path (upgraded, 3600s), and `/events` (SSE: buffering off,
  HTTP/1.1, 3600s). `nixosModules.webphone` is an alias of `.default`.
- The `webphone-module` flake check evaluates the module with stand-in
  options (nginx/systemd/users + `assertions` — NixOS's modules.nix
  normally provides `assertions`; new config keys the module writes
  need a stand-in there) and asserts the three vhost locations plus the
  `webphone` systemd unit. Statix pins single-assignment style: all
  `locations` in ONE attrset (three `locations.X =` assignments fail
  the gate).
- webphone's gateway seam (loopback vs webhook) is consumed by
  pbx-artmann's `telnyx-webhooks.py` bridge (secrets via LoadCredential/
  EnvironmentFile under `/var/lib/telephony-secrets`) — contracts in
  the plan doc `docs/planning/2026-09-19_11-51_SUPERB-*`.

## Commands

```console
nix develop                        # Go, templ, golangci-lint, esbuild, … — the shell exports GOEXPERIMENT=jsonv2 + GOTOOLCHAIN=local, so bare `go` commands work inside it
templ generate ./internal/web/views/   # after ANY .templ edit (committed *_templ.go)
GOEXPERIMENT=jsonv2 go test -count=1 ./...  # jsonv2 REQUIRED for every go command OUTSIDE the devShell (templ-components); -count=1: the result cache has lied during investigations
python3 scripts/webphone-smoke.py          # 26-check live smoke over real HTTP (boots a fresh binary + temp data dir; --base URL reuses a running server)
buildflow                                  # the quality gate; BUILDFLOW_NO_RESULT_CACHE=1 for full
nix run .#vulnix                           # vulnix --closure over the RUNTIME closure (network; exits non-zero with triage guidance on findings)
nix flake check                            # package build + tests in sandbox + treefmt + island-lint
nix build .#webphone --system aarch64-linux   # cross-builds
./update.sh [version]              # repin vendored sip.js (fetch → esbuild IIFE → swap)
```

Smoke a binary quickly (loopback gateway = whole product, zero PBX):

```console
GOEXPERIMENT=jsonv2 go build -o /tmp/webphone-bin ./cmd/webphone
WEBPHONE_ADDR=127.0.0.1:18099 WEBPHONE_DATA_DIR=/tmp/wp-data \
  WEBPHONE_GATEWAY__WEBHOOK_SECRET=devsecret /tmp/webphone-bin
```

## The DOM + bundle contract (DO NOT BREAK CASUALLY)

The consuming stack's browser E2E (`tests/browser-e2e.py` there) drives
the island remotely — re-run it after any markup change. The Go test
`TestServedPageHoldsTheDomContract` (`internal/server/server_test.go`)
asserts all 35 island element ids (`reg-status`, `login-view`, `keypad`,
`vm-wrap`, `history-list`, `contacts-wrap`, `ice-wrap`, `log`, …) on
every build; it is the local tripwire, not a replacement for the E2E.

- The island modules under `internal/web/assets/island/app/` are served
  VERBATIM (no bundling, no minification), so the E2E's greppable
  strings (`dtmf-relay`, `userAgent.reconnect()`, `.refer(`,
  `blindTransfer`, `attendedTransfer`, `Notification.requestPermission`,
  `titleFlashStart`, `ringToneStart`, `phone-api/history`,
  `phone-api/voicemail`, `candidate-pair`, `currentRoundTripTime`)
  survive by construction. Keep it that way.
- E2E also drives `#reg-status` text (`registered` /
  `registration rejected`), `.hidden` toggles on `#login-view` /
  `#phone-view`, `.call-card` / `.call-state-text` (`in call`),
  `.hangup-btn`, `.transfer-btn`, `.transfer-dest`, keypad
  `button[data-tone]`, and `window.__pcs` (Map of live
  RTCPeerConnections — media proof).
- The event log (`#log`) stays **English** in both UI languages: it is
  operator-facing diagnostics and the runbook greps it.

## Architecture invariants

- **Middleware chain** (server `New`): `ContextEnrichmentMiddleware(nil)`
  outermost → `RequestLoggingSlog` → `ServerTimingMiddlewareWhen`
  (env-gated) → `SecurityHeaders` → `cqrshtmx.RecoveryMiddleware` →
  routes (2026-09-19 superb-adoption plan). Enrichment must stay OUTSIDE
  the request log: the logger reads the request context AFTER the
  handler returns, so a logger outside enrichment can never see the
  RequestID — the whole point is `request_id=` in every log line plus
  the `X-Request-ID` response header. The user extractor stays nil:
  library user ids are ULIDs from the rejected usermgmt module, and
  extensions are not ULIDs. Server-Timing uses the library middleware
  (W3C `total;desc="Total request";dur=…`, CRLF-sanitized, SSE-safe
  writer) — the hand-rolled `timingWriter` is gone; `securityHeadersConfig()`
  is the single source for header config (test parity is structural).
  `Permissions-Policy` ships calibrated: `microphone=(self)`, camera/
  display-capture/geolocation/payment/usb denied — the library's
  `RecommendedPermissionsPolicy` stays rejected (denies microphone).
  Adopted
  2026-09-18 per the Pareto plan (links below): panics log full stacks
  and re-raise `http.ErrAbortHandler`; every request logs exactly one
  line (method/path/status/duration — never bodies/credentials); the
  login/hook limiters are `httputil.KeyedRateLimiter` with port-stripped
  peer-host keys (`remoteHostKey` — port-qualified keys would silently
  disable limiting behind the stack's proxy; flip to
  `KeyExtractorFromClientIP` only once the stack proves XFF
  sanitization); `/healthz` is honest readiness (`sqlite` ping +
  `blob-dir` write probe, 503 names the failing check, library JSON
  shape; GET-open by decision — probers need no session and the body
  leaks only check names/errors, never secrets). It is
  readiness-ONLY — no liveness endpoint exists and
  `cqrshtmx.ReadinessCheck` carries no per-check timeout (both current
  checks are local, so no realistic hang source); the DI/health review
  of 2026-09-19 (`docs/architecture-understanding/`) scored the posture
  and left F1 (timeout guard) / F2 (liveness decision) in TODO_LIST;
  `/events` rides `Broadcaster.ServeSSE` (its `connected`
  handshake frame is additive; htmx sse-swap listeners ignore it;
  payloads stay swap-safe fragments).
- **The island never unloads.** Tab navigation swaps partials into
  `#tab-content` via HTMX; the SIP island lives outside that region so
  calls survive tab switches. Deep links (`/messages`, `/fax`, …)
  render the full shell server-side.
- **Module graph stays acyclic**: the island's `state.js` + `auth.js`
  exist so calls/ice/connection never import each other; the UA lives
  in `state.userAgent` and ice syncs via the `wp:calls-changed`
  CustomEvent (no direct imports — enforced by
  `internal/arch/arch_test.go`, which also asserts `domain` imports
  nothing internal and services never import `server`/`web`); the
  server mirrors the same rule.
- **Sessions**: the island POSTs `/api/session` and the server VERIFIES
  the submitted extension/password against the PBX directory before
  minting (2026-09-19, live-probe finding: the old trust-the-island
  design let a forged POST open a session scoped to any extension —
  tab partials, fax/attachment streams and SSE fragments scope by the
  session alone, so stored threads/fax/contacts of any extension were
  readable without its password). Fail-closed: 401 rejected
  credentials, 502 PBX unreachable; deployments without a phone API
  (loopback dev) skip verification and WARN at boot. The server keeps
  sessions in an in-memory TTL store + HttpOnly cookie. `session.js`
  attaches `sse-connect` to `.wp-root` post-login (no reload — the
  password is memory-only) and reloads the page on logout. Login and
  hooks are per-IP rate limited (the hook limiter wraps, not sits
  inside, the secret gate). Server handlers self-gate through the
  shared `requireSession` helper; `Sessions.Require` middleware
  additionally wires `/events` and `/phone-api/` — dual-layer by
  design (pages must render anonymously; routes stay gated regardless
  of wiring). The contract test allowlists exactly three 401 writers:
  actions.go (session gate), webhooks.go (secret gate),
  session_api.go (login credential gate).
- **Language**: UI language is per extension — `wp-lang` cookie
  (written by the island's `setLang`, samesite=strict) →
  `Accept-Language: de*` → English default. `ExtensionHubs` remember
  the negotiated lang so SSE fragments render in it (the notifier has
  no request). Service validation reasons stay English: operator-facing,
  runbook-greppable — same policy as `#log`.
- **SSE payloads are swap-safe fragments** (`ThreadsList`, `Transcript`,
  `FaxList` — no wrappers, no composers): `sse-swap` replaces
  innerHTML, so a wrapped payload would nest panels and wipe drafts.
  The `voicemail` event is a payload-less NUDGE: the voicemail panel
  re-fetches its partial on receipt (it needs per-session PBX
  credentials the notifier does not have). Event names: `threads`,
  `thread`, `fax`, `voicemail`.
- **Gateway seam**: loopback (dev) vs webhook (multipart to
  `{url}/message|/fax`, Bearer secret, `{"provider_ref"}` receipt).
  Inbound hooks `/hooks/*` share the same secret and fail CLOSED
  (503) when none is configured. Status hooks are idempotent:
  `hooksIdem` (in-memory TTL idem store) dedupes replayed
  `provider_ref` — only successes are recorded, so failures stay
  retryable, and replays answer `202 Accepted` inertly.
- **Owner scoping everywhere**: every store query is extension-scoped;
  attachments/faxes stream through session-gated handlers only.
- **CSP**: same-origin only, `connect-src wss:` for SIP; no CDN, no
  webfonts, no inline handlers. One inline script is allowed by exact
  hash (templ-components' theme preload — no opt-out knob upstream as
  of v1.18.0, and it is inert here: theming rides `data-theme`, not the
  Tailwind dark class); `TestServedPageSatisfiesStrictCSP` checks the
  hash two ways so a dependency bump that changes the script fails the
  build until the hash is refreshed deliberately.
- `window.PBX_CONFIG` (`/config.js`, rendered by this server): keys
  `sipDomain`, `websocketPath`, `iceServers`, `phoneApi`, `contacts`.

## Hard-won knowledge

- cqrs-htmx audit trail: deep-dive
  `docs/research/2026-09-18_cqrs-htmx-deep-dive.html`, execution plan
  `docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`,
  utilization audit
  `docs/research/2026-09-19_cqrs-htmx-deep-dive.html` (78/100 baseline),
  and superb-adoption plan
  `docs/planning/2026-09-19_09-30_cqrs-htmx-superb-100-adoption-plan.md`
  (executed 2026-09-19: request-ID enrichment, `templ.JSONString` CSRF
  wiring, calibrated Permissions-Policy, servertiming middleware, webhook
  5xx redaction via `webhookFail`/`SafeDetail` — all landed with tests).
  Still open: the idiomorph experiment (gated on the stack's browser E2E).
  Adoption posture: middleware + assets only; the `setup` bundle, CQRS
  dispatch layer and usermgmt stay rejected (split-brain identity, see
  above); security presets are NEVER adopted wholesale — the library's
  `RecommendedPermissionsPolicy` denies `microphone`, which would kill
  the WebRTC phone. `toastDetail` is a type alias of
  `cqrshtmx.ToastDetail` (root-package type; NOT dispatch-layer — the
  old "kept local" comment was wrong), so a wire-shape change upstream
  fails this build. Trap: the dispatch-layer `Notify*` options emit
  `{level,message}`, NOT the island's `{message,kind}` shape.
- Two CSRF constraints discovered 2026-09-19 while hardening the wiring
  (both test-caught before they shipped):
  (1) `httputil.CSRFTokenHXHeaders`/`CSRFTokenHTMLMeta` are for RAW-HTML
  contexts — they HTML-escape their output. templ escapes attribute
  values itself, so the helpers would double-escape: hx-headers would
  fail JSON.parse and every HTMX request silently loses CSRF protection.
  In templ, build the value with `templ.JSONString` (pages.go renderShell)
  and let templ escape it once. `TestShellRendersValidJSONCSRFHxHeaders`
  pins this.
  (2) CSRF token rotation on login (SHIPPED 2026-09-19): login/logout
  call `httputil.InvalidateCSRFCookie` (fixation defense) and the island
  adopts the fresh token WITHOUT a reload via `GET /api/csrf`
  (`session.js adoptFreshCsrfToken`, `csrf_api.go refreshCSRF`): the GET
  rides the CSRF middleware, where nosurf regenerates the deleted cookie
  and exposes the new masked token; every token consumer reads live
  (meta tag in session.js/auth.js, htmx re-reads body `hx-headers` per
  request), so updating those two spots re-arms all POSTs. Adoption
  failure falls back to a reload (server session cookie survives).
  Tests that POST after logging in must go through the client `login`
  helper (it adopts) — raw login POSTs leave a dead token; the rate-limit
  and request-log loops re-arm between attempts like a scripted flooder.
- CSRF behind a TLS-terminating proxy needs `csrf.trusted_*` (found
  2026-09-19 via the stack E2E 403s): the browser sends
  `Origin: https://host` + `Sec-Fetch-Site: same-origin`, the listener
  sees plain HTTP, and httputil's attestation check reads the mismatch
  (scheme-only: `r.Host` matches) as a FORGED attestation → 403 on
  every POST. v2.0.0 shipped this; tab logins silently never worked in
  fronted deployments (island calls kept working, so the E2E stayed
  green). Fix: `csrf.trusted_proxies` (loopback nginx) +
  `csrf.trusted_origins` (the https vhost) — `requestScheme` honors
  X-Forwarded-Proto only from trusted proxies. Module and stack set
  both; `TestCSRFTrustsTheFrontingProxy` pins the request shape.
- Verify dependency internals at the CONSUMED tag (module cache or
  `git show v4.9.0:<path>`), never master: the 2026-09-18 audit
  over-credited v4.9.0's `ServeSSE` with a `retry:` hint that only
  exists on master — tag-checking before the port caught it.
- `GOEXPERIMENT=jsonv2` is required for every `go` command —
  templ-components/errorpage needs `encoding/json/v2`.
- `.templ` files must NOT import `github.com/a-h/templ` (the generator
  injects the symbol); conditionals are bare `if x {` statements, not
  `@if`.
- go-branded-id: `id.ID.String()` renders `"Brand:value"`; the domain
  `mustID` parsers strip an optional `"Prefix:"`.
- DTMF must be `application/dtmf-relay` with `Signal=<d>` (equals form)
  — the colon form is 200-OK'd by FreeSWITCH and silently dropped.
- FreeSWITCH executes transfers server-side on the REFER; the browser
  only sends it and parses the NOTIFY sipfrag verdict (`referOnNotify`).
- sip.js is pinned at 0.21.2 (vendored tarball + IIFE bundle). The 0.x
  series hangs in `userAgent.reconnect()` after transport loss; the
  bounded watchdog in the island's `connection.js` (5s per attempt,
  full rebuild on timeout) is load-bearing. Do not "simplify" it away.
- Env config nests with `__`: `WEBPHONE_GATEWAY__MODE` →
  `gateway.mode`; single underscores stay literal (`WEBPHONE_DATA_DIR`
  → `data_dir`). Scalars via env; lists (`ice_servers`, `contacts`)
  via the JSON file.
- `buildflow -s nix-hash-fix --fix` computes the right vendorHash but
  never writes it here (buildflow itself warns). Documented deviation:
  placeholder hash → `nix build` → read `got:` → apply. Only when
  go.mod/go.sum actually changed.
- erraudit honors `//nolint:erraudit // reason`; branching-flow honors
  NO nolint — its remaining policy-opinion findings are triaged as a
  documented skip in `.buildflow.yml` (same for go-structure-linter,
  cqrs-lint, nix-hash-fix).
- htmx loads deferred from `headExtras`, after the `htmx-config` meta
  that disables its inline indicator-style injection (`app.css` ships
  the same rules so hx-indicator keeps working). The meta must precede
  the script or htmx never reads it; Shell leaves `HTMXSrc` unset so
  `layout.Base` does not also emit a synchronous htmx tag.
- The app.css `[data-theme]` `color-scheme` rules carry `!important` so
  the CSP-hash-pinned framework theme script's inline `colorScheme`
  cannot undo a forced theme. If templ-components ships a ThemeScript
  opt-out knob, take it and drop the hash plus these `!important`s.
- `pbx.Client` owns the timeout-bounded HTTP client; the `/phone-api`
  proxy must ride `PhoneAPI.HTTPClient()`, never `http.DefaultClient`.
  Join path and query separately — `url.JoinPath` percent-encodes `?`
  (a test caught upstream receiving `history%3Flimit=30`).
- i18n dictionaries live in `views/i18n.go`; unknown keys surface
  themselves in the page (deliberate) and a test keeps en/de in sync —
  add new keys to BOTH maps.
- vulnix: `vulnix --closure <out-path>` is the ONLY scoped mode — plain
  vulnix expands whatever it is given into the BUILD closure (bootstrap
  toolchains, binutils, gcc: dozens of findings that never deploy), and
  per my 2026-09-19 verification even passing `$(nix-store -qR <out>)`
  paths does NOT scope it; `nix run .#vulnix` wraps the correct call.
  vulnix also range-matches distro-patched versions: it prints glibc
  CVE-2026-5450 against glibc-2.42-84, but the fix shipped in nixpkgs
  2.42-67 (PR #517918 — the locked tree's glibc `2.42-master.patch`
  carries it); NVD ranges cannot see patch suffixes. Runtime closure
  (8 derivations) carries zero real advisories (re-verified 2026-09-19
  with `--closure`).
- Formatting: treefmt (prettier) owns everything under
  `internal/web/assets/island/`; `.buildflow.yml` excludes the island
  so BuildFlow's oxfmt cannot fight prettier (same war the telephony
  repo fought — pre-settled here).
- Island no-undef gate: `nix flake check` runs `island-lint` — oxlint
  with `internal/web/assets/island/oxlint.json` (all categories off,
  `no-undef` on, `SIP` declared readonly). It pins exactly the bug class
  that surfaced as a silent browser ReferenceError (accept/reject):
  calls to undefined identifiers in the island modules. New browser
  globals go in the config's `globals` block. The check fails closed
  and records the scanned file list (a green gate must prove it
  scanned).
- Live-transcript contract (shell.js ↔ server): `#thread-transcript`
  carries `data-page` + `data-thread`. shell.js cancels the sse
  extension's `htmx:sseBeforeMessage` while `data-page != "0"` (paging
  state survives live pushes) and POSTs `/messages/{id}/read` plus a
  `GET /partials/nav?active=<tab>` into `#wp-nav` after a newest-page
  push (a live swap never re-GETs the partial, so only the client can
  clear the unread badge). The sse extension (v2.2.4 served by
  cqrs-htmx v4.9.0) fires cancelable `htmx:sseBeforeMessage` before
  `htmx:sseMessage` — verified in the extension source, not assumed.
- Nav language mechanism (decided): the island's language switch is a
  client-side cookie write with no server round-trip, so it cannot carry
  HX-Trigger; it dispatches `wp:lang-changed` and shell.js re-fetches
  `/partials/nav` — nav labels switch language without a full reload.
  `/partials/nav` renders labels anonymously (no badges) and
  signed-in with fresh badge caches.

## Release runbook (v2.x)

The dance that cut v2.0.0, written down so the next release is a
checklist, not archaeology. The auto-commit daemon commits AND pushes
continuously — work in small, explicitly-committed units.

1. **Fold**: CHANGELOG `Unreleased` → dated section; sync
   FEATURES/TODO_LIST/ROADMAP; explicit commit per doc group.
2. **Bump `webphoneVersion`** in flake.nix (package version AND the
   `/version` ldflags injection — one let-binding; keep it equal to the
   new tag).
3. **Gates**: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`,
   `GOEXPERIMENT=jsonv2 go test -count=1 ./...`, `nix flake check`,
   `python3 scripts/webphone-smoke.py`.
4. **Tag + push**: annotated `git tag -a vX.Y.Z -m ...`, push main +
   tag (verify with `git ls-remote` — the daemon may have pushed
   already).
5. **Link check**: `nix run nixpkgs#lychee -- .` — after the push, so
   the new tag link resolves.
6. **Stack bump** (`~/projects/nix-international-telephony`):
   `nix flake lock --update-input webphone`, commit the lock.
7. **Stack gates**: `nix build -L .#telephony-browser` (browser E2E,
   chromium NixOS test — THE island regression gate, deliberately
   outside `checks` for its ~1-2 GB chromium closure);
   `nix build -L .#checks.x86_64-linux.telephony-webphone` (webphone
   VM test); full `nix flake check` with the new lock before
   announcing.
8. **aarch64**: `nix build .#webphone --system aarch64-linux` — plain
   `nix flake check` silently omits aarch64 (it says so in a warning).
   Do NOT reach for `nix flake check --all-systems` as the fix: it is
   evaluation-only for other systems (verified 2026-09-19 — zero
   derivations built, "running 0 flake checks") and gates nothing.
   Checks DO cross-build if you name them:
   `nix build .#checks.aarch64-linux.island-lint` ran green cross-arch
   (oxlint substitutes from cache.nixos.org), so the aarch64 gate is
   explicit cross-builds of the package + the checks you care about.

## Buildflow health warning, itemized (2026-09-19)

`buildflow` ends with "9 tools unavailable (health check failed)". All
nine are noise for THIS repo: every one additionally reports "missing
prerequisite files" and lands in "not applicable" — their steps never
run (no package.json, no Python package layout): jest, knip, madge,
publint, svelte-check, vitest, vue-tsc (JS/TS), c8/js-coverage,
interrogate (Python). The one real prerequisite gap was go-licenses
(license-scan preflight) — now in the devShell, so run buildflow inside
`nix develop` or the preflight warns "go-licenses binary not found".
markdown-lint/gitleaks/codespell only run in build mode `full`.

## Conventions

- One home per fact: README sells + documents contracts, FEATURES
  inventories status, TODO_LIST holds open work, CHANGELOG logs
  history, this file keeps session-durable knowledge.
- Cite stable names (ids, function names, option names), not `file:line`.
- Behavior parity rules ports: port logic verbatim first, refactor in a
  second, separately-verified change.
- An auto-commit daemon commits continuously AND pushes (observed
  2026-09-19: origin/main tracks HEAD within minutes); never revert
  changes you did not author, and verify end states with
  `git ls-remote`, not push logs — a "local" commit may already be
  public.
- Markdown is NOT in treefmt scope (flake.nix prettier includes only
  `*.css` + island/shell JS): `nix fmt` saying "0 files" on `.md` edits
  means unformatted-by-tooling, not clean — struck tables in annotated
  reports keep manual alignment.
- Recording is two-level: the PBX stack records every dialled call
  server-side (`record_session`, stereo WAV under the stack's
  `/recordings/`, operator basic-auth; `*97<ext>` skips) — this repo's
  island and server have NO recording capability, only CDR history
  rows. Product-level capability questions get answered per level
  (island / Go server / consuming stack).
