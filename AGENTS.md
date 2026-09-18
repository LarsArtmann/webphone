# AGENTS.md

Enduring context for AI sessions working in this repo.

## What this is

A single-binary Go unified-communications web app (v2.0.0, 2026-09-18):
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
additive. Stack-side switchover remains open work (TODO_LIST).

The cqrs-htmx `setup` bundle was rejected deliberately: it wires
event-sourced usermgmt users, but this product's identity is the PBX
extension + directory password (proven by the island's SIP REGISTER) —
a second user database would be a split brain.

## Commands

```console
nix develop                        # Go, templ, golangci-lint, esbuild, …
templ generate ./internal/web/views/   # after ANY .templ edit (committed *_templ.go)
GOEXPERIMENT=jsonv2 go test ./...  # jsonv2 REQUIRED for every go command (templ-components)
buildflow                          # the quality gate; BUILDFLOW_NO_RESULT_CACHE=1 for full
nix flake check                    # package build + tests in sandbox + treefmt
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

- **Middleware chain** (server `New`): `RequestLoggingSlog` outermost →
  `SecurityHeaders` → `cqrshtmx.RecoveryMiddleware` → routes. Adopted
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
  leaks only check names/errors, never secrets); `/events` rides `Broadcaster.ServeSSE` (its `connected`
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
- **Sessions**: the island POSTs `/api/session` AFTER its REGISTER
  succeeds (credentials proven against the PBX); the server keeps them
  in an in-memory TTL store + HttpOnly cookie. `session.js` attaches
  `sse-connect` to `.wp-root` post-login (no reload — the password is
  memory-only) and reloads the page on logout. Login and hooks are
  per-IP rate limited (the hook limiter wraps, not sits inside, the
  secret gate).
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
  (503) when none is configured.
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
  `docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`.
  Adoption posture: middleware + assets only; the `setup` bundle, CQRS
  dispatch layer and usermgmt stay rejected (split-brain identity, see
  above); security presets are NEVER adopted wholesale — the library's
  `RecommendedPermissionsPolicy` denies `microphone`, which would kill
  the WebRTC phone.
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
- vulnix against `./result` scans the BUILD closure (bootstrap
  toolchains, binutils, gcc — dozens of findings that never deploy).
  The honest number is the runtime closure: `nix-store -qR result`
  (8 derivations); only glibc carried advisories as of 2026-09-18.
- Formatting: treefmt (prettier) owns everything under
  `internal/web/assets/island/`; `.buildflow.yml` excludes the island
  so BuildFlow's oxfmt cannot fight prettier (same war the telephony
  repo fought — pre-settled here).

## Conventions

- One home per fact: README sells + documents contracts, FEATURES
  inventories status, TODO_LIST holds open work, CHANGELOG logs
  history, this file keeps session-durable knowledge.
- Cite stable names (ids, function names, option names), not `file:line`.
- Behavior parity rules ports: port logic verbatim first, refactor in a
  second, separately-verified change.
- An auto-commit daemon commits continuously; do not be surprised by
  commits you did not make, and never revert changes you did not author.
