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
E2E is green, and its `webphone` input rides webphone `main` — as of
2026-09-20 pinned to the v2.4.0 release commit `a59f0d1` (stack commit
`a273d3f`; pbx-artmann relocked on top at `5858430` and its prod
toplevel pre-builds green).

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
  MemoryMax), `csrf.{trustedProxies,trustedOrigins}` (typed fronts for
  `settings.csrf.*` — empty preserves the `nginx.enable` defaults,
  non-empty overrides them), `serverTiming.enable` (sets the
  `WEBPHONE_DEBUG_TIMING` env gate), `backup.{enable,destDir,calendar}`
  (daily online snapshot timer), `nginx.{enable,hostName}`. Its vhost
  proxies `/`, the websocket path (upgraded, 3600s), `/events` (SSE:
  buffering off, HTTP/1.1, 3600s) and the probe triple
  (`/healthz`/`/livez`/`/startupz` as DEDICATED locations so fleet
  scrapers can be fenced per location without touching `/`).
  `nixosModules.webphone` is an alias of `.default`.
- The `webphone-module` flake check evaluates the module with stand-in
  options (nginx/systemd/users + `assertions` — NixOS's modules.nix
  normally provides `assertions`; new config keys the module writes
  need a stand-in there) and asserts all six vhost locations, the
  `webphone` systemd unit, the csrf fronting defaults AND the typed
  override, the backup timer/oneshot pair, and the serverTiming env
  gate. Statix pins single-assignment style: all `locations` in ONE
  attrset (repeated `locations.X =` assignments fail the gate).
  `checks.x86_64-linux.webphone-backup` is a kvm-gated NixOS VM test
  (fax-feed-test style): boots the real service with `backup.enable`,
  runs the oneshot, asserts snapshot + `pragma integrity_check` + timer
  wiring + `NRestarts=0`; without KVM, `nix flake check` skips it with
  a warning instead of crawling under TCG.
- webphone's gateway seam (loopback vs webhook) is consumed by
  pbx-artmann's `telnyx-webhooks.py` bridge (secrets via LoadCredential/
  EnvironmentFile under `/var/lib/telephony-secrets`) — contracts in
  the plan doc `docs/planning/2026-09-19_11-51_SUPERB-*`.

## Owner decisions (2026-09-20)

The four one-liners gating the 10:14 SUPERB pareto plan (T14), decided
with rationale so they stop blocking work:

- Stack `webphone` input policy: DECIDED — ride webphone `main` with a
  per-train lock bump (runbook step 6), matching the g2 cadence
  recommendation. Tag pins buy reproducibility at the cost of a manual
  bump on every fix; revisit if a hotfix must ever ship inside an hour.
- pbx-artmann input type: DECIDED — keep `path:` while all three repos
  live on this host. The 2026-09-19 "burn" was an uncommitted stack
  tree (now a documented relock precondition), not the input type.
  Revisit when pbx-artmann leaves the host or gains a second consumer
  (this supersedes the earlier github-input recommendation in the
  15:37 SUPERB plan P8).
- Sanitization side: DECIDED — the island keeps letters
  (`[^\d+*#a-zA-Z]`), matching `sanitizeDialable`; shipped with the
  served-asset pin in internal/server (the ids.go parity comment is
  true again).
- Own-number feed: DECIDED — static config map (`identities`) now; a
  stack `/phone-api` identity endpoint remains the upgrade path
  (blocked on the stack tree reconciliation). CDR-derive REJECTED on
  evidence: the stack's phone-api reshapes FreeSWITCH `Master.csv`,
  where outbound `caller_id_number` is dialplan-dependent (webphone's
  fixtures show the extension `1001`, not the DID) and absent before
  the first call.

## Commands

```console
nix develop                        # Go, templ, golangci-lint, esbuild, … — the shell exports GOEXPERIMENT=jsonv2 + GOTOOLCHAIN=local, so bare `go` commands work inside it
templ generate ./internal/web/views/   # after ANY .templ edit (committed *_templ.go)
GOEXPERIMENT=jsonv2 go test -count=1 ./...  # jsonv2 REQUIRED for every go command OUTSIDE the devShell (templ-components); -count=1: the result cache has lied during investigations. Since the go 1.27.1 floor, OUTSIDE-the-shell go commands also need the 1.27 toolchain — prefer `nix develop -c` wrappers
python3 scripts/webphone-smoke.py          # 32-check live smoke over real HTTP (boots a fresh binary + temp data dir; --base URL reuses a running server; includes /livez + /startupz, /version, /openapi.json, the styled 404)
buildflow                                  # the quality gate; BUILDFLOW_NO_RESULT_CACHE=1 for full (release.sh now also gates on `nix run .#vulnix`)
nix run .#vulnix                           # vulnix --closure over the RUNTIME closure (network; exits non-zero with triage guidance on findings; the verdict logic is the `webphone-vulnix-triage` CLI, fixture-checked by `checks.vulnix-triage`)
nix flake check                            # package build + tests in sandbox + treefmt + island-lint + island-js (node:test) + the kvm-gated backup VM test (skipped with a warning without /dev/kvm)
nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs   # the island JS tests alone (toast rendering, i18n en/de parity); stubs in island-tests/helpers.mjs
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
- Unknown paths render the STYLED 404 (`notFoundPage` in pages.go: the
  full shell + ErrorPanel, status stays 404) — error-page parity with
  the pre-2.0 static site, restored 2026-09-20 (the templ-components
  adoption had silently reverted it to Go's bare-text 404). Partial
  swaps never see it (they target existing regions); the smoke's
  stale-CSRF probe still reads a plain 404 status.

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
  leaks only check names/errors, never secrets). Each named check runs
  under a 2s webphone-side bound (`boundedCheck`: a timeout degrades
  the probe to 503 naming the check, the underlying call keeps
  running) — F1 of the DI/health review closed, now via the UPSTREAM
  mechanism: cqrs-htmx v4.11.0's `NamedCheck.Timeout` (the local
  `boundedCheck` wrapper was deleted when webphone bumped). The probe
  triple is completed by go-health v0.3.0's container-free
  `NewChecks` (2026-09-19, F2 DECIDED + shipped: go-health JSON
  probes, the plan's recommended posture): `/livez` is fetch-free
  process liveness (200 while serving — the split from dependency
  health is the point), `/startupz` is 503 until the same two backing
  checks first pass, then latched. Readiness stays `/healthz` alone —
  no second readiness truth; the go-health probe SHARES the check
  functions (same truth, two lifecycles). JSON only (CSP-neutral);
  the Datastar HTML dashboard face stays rejected (`unsafe-eval`) —
  options memo:
  `docs/architecture-understanding/2026-09-19_20-59_health-probes-fleet-options.md`.
  samber/do appears only as a TRANSITIVE dep of go-health's recorder
  interfaces — the container itself stays rejected. The
  self-health plan
  (`docs/planning/2026-09-19_20-01_SUPERB-honest-self-health-upstream-first-plan.md`)
  is EXECUTED. The review
  of 2026-09-19 (`docs/architecture-understanding/`) scored the posture;
  `/events` rides `Broadcaster.ServeSSE` (its `connected`
  handshake frame is additive; htmx sse-swap listeners ignore it;
  payloads stay swap-safe fragments; v4.11.0 leads the stream with a
  `retry:` reconnect hint — the one wire change of the
  v4.9.0→v4.11.0 bump, pinned by
  `TestSSEStreamCarriesConnectedThenEvents`; the bump-trigger
  follow-through history lives in CHANGELOG and the fan-out baseline
  doc, not here).
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
  (loopback dev) skip verification and WARN at boot. Since 2026-09-20
  (plan T11/T12, verdict
  `docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md`)
  the store is a SEAM: prod runs `NewSQLiteStore` over `webphone.db`
  (sessions survive restarts; the row carries the extension + directory
  password the `/phone-api` proxy needs — credentials at rest accepted,
  bounded by the same 24h TTL, swept on read/Create), tests and loopback
  keep `NewMemStore` (in-memory TTL + GC, sessions die with the
  process). Cookie unchanged (HttpOnly, Max-Age = SessionTTL). The
  "Tab session ended" toast is now the rare path (hard crash mid-TTL,
  manual cookie clear). `session.js`
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
- **Live-update surfaces morph-swap** (since 2026-09-20): the five
  SSE/nav surfaces — thread list, `#thread-transcript`, fax list,
  voicemail panel re-fetch, shell.js `refreshNav` — carry
  `hx-swap="morph:innerHTML"` and are reconciled by idiomorph
  (cqrs-htmx v4.11.0's bundled self-contained exts, served as ONE
  bundle at `/htmx-ext.js` via `cqrshtmx.HTMXExtensionsHandler` —
  sse + idiomorph concatenated, composite ETag; the sse ext resolves
  swaps via htmx `getSwapSpecification`, so the attribute IS honored on `sse-swap`
  elements). Morph preserves matched nodes in place: focus, draft
  text, container attrs (`data-page`/`data-thread`) and shell.js
  listeners survive live pushes. im-preserve audit conclusion
  (2026-09-20): idiomorph 0.7 persists any element whose id exists in
  BOTH trees (morphed in place, pantry-pulled across reorders) and its
  `restoreFocus` equally needs an id — so STATEFUL nodes in morph
  surfaces must carry stable ids (voicemail rows + `<audio>` do:
  `vm-<uuid>`/`vm-audio-<uuid>`, pinned by
  `TestVoicemailRowsCarryStableMorphIds`; thread/fax rows and bubbles
  are static content, re-creation there is harmless). Payloads still render as bare
  fragments (`ThreadsList`, `Transcript`, `FaxList` — no wrappers,
  no composers) — the shape is convention even though morph no
  longer wipes drafts on it. The `voicemail` event stays a
  payload-less NUDGE: the voicemail panel re-fetches its partial on
  receipt (it needs per-session PBX credentials the notifier does
  not have). Event names: `threads`, `thread`, `fax`, `voicemail`.
  New live surfaces should follow the morph pattern.
- **Gateway seam**: loopback (dev) vs webhook (multipart to
  `{url}/message|/fax`, Bearer secret, `{"provider_ref"}` receipt).
  Inbound hooks `/hooks/*` share the same secret and fail CLOSED
  (503) when none is configured. Status hooks are idempotent:
  `hooksIdem` (in-memory TTL idem store, `hookIdempotencyTTL` = 1h)
  dedupes replayed `provider_ref` — only successes are recorded, so
  failures stay retryable, and replays answer `202 Accepted` inertly.
  TTL rationale (2026-09-20): the idem window only has to cover the
  provider's BURST retries (Telnyx re-delivers within minutes); a
  re-delivery after the hour re-applies a status SET, and status
  transitions converge, so no persistence is needed here.
- **Owner scoping everywhere**: every store query is extension-scoped;
  attachments/faxes stream through session-gated handlers only.
- **Personal contacts have ONE home** (2026-09-19): the per-extension
  SQLite store. The island reads/writes it via `/api/contacts` (JSON,
  session-gated, extension-scoped, CSRF via `authedFetch`); mutations
  answer 204 and the LIST is the only id source — the store upserts by
  (owner, phone) and keeps the old id on rename, so a returned minted
  id could drift. The legacy `localStorage["pbx-contacts"]` list
  imports once post-login and is REMOVED only after the server
  accepted every row (failed imports retry next login; upsert makes
  re-import idempotent). Load trigger: session.js dispatches
  `wp:session-opened` AFTER cookie mint + CSRF adoption — loading
  earlier would POST on a dead token. History stays hybrid BY DESIGN
  (local session log + same CDR API) — not a split brain, don't "fix"
  it.
- **Tab→island affordances live in shell.js** (2026-09-19): the
  delegated `data-dial` handler (guarded: hidden `#phone-view` → toast
  in `#toasts` + focus `#ext`, never a silent submit) and the
  live-call badge (`wp:calls-changed` → `#call-badge` in the header,
  counting `.call-card` in `#calls`). New shell-side listeners go in
  shell.js, never island modules — the shell must keep working when
  island scripts fail. The htmx error-surfacing listener (shell.js §3c,
  2026-09-20) is the client half of the error-feedback story: htmx
  swaps NOTHING on 4xx/5xx (verified in the v4.11.0 htmx.min.js
  `responseHandling` defaults), so a dead tab session (the in-memory
  store is cleared by every server restart) used to make every tab
  click and form submit fail silently. shell.js now toasts on
  `htmx:responseError` (401 wording: session ended, calls keep working,
  reload when convenient — never an AUTO reload, the island must
  survive) and `htmx:sendError`, skips responses carrying HX-Trigger
  (those already toast server-authored feedback via `renderPanelError`),
  and throttles to one error toast per 8s (an expired session 401s
  every later request). Pinned behaviorally island-side
  (island-tests/shell.test.mjs drives the real document listeners with
  a fake clock; helpers.mjs now returns the stub document and
  dispatches through it) and by `TestShellJSSurfacesHtmxErrors` on the
  SERVED asset.
- **CSP**: same-origin only, `connect-src wss:` for SIP; no CDN, no
  webfonts, no inline handlers. One inline script is allowed by exact
  hash (templ-components' theme preload — no opt-out knob upstream as
  of v1.18.0, and it is inert here: theming rides `data-theme`, not the
  Tailwind dark class); `TestServedPageSatisfiesStrictCSP` checks the
  hash two ways so a dependency bump that changes the script fails the
  build until the hash is refreshed deliberately.
- `window.PBX_CONFIG` (`/config.js`, rendered by this server): keys
  `sipDomain`, `websocketPath`, `iceServers`, `phoneApi`, `contacts`.

## Failure → feedback map (2026-09-20 train, plan T13)

The single table that answers "what does the user SEE when X fails".
Every error path lands in at least one VISIBLE surface (toast, inline
banner, or panel); `#log` is always the operator trail, never the only
user feedback.

| Failure | User sees | Owner (wording) | Test home |
|---|---|---|---|
| Tab session dead (401 on tab actions) | Throttled error toast: "Tab session ended; calls keep working…" — never auto-reload | shell.js §3c (English, D3) | shell.test.mjs + `TestShellJSSurfacesHtmxErrors` |
| Validation mistake (422) | `.wp-error` banner swapped into `#wp-tab-error` + error toast | server (en/de via `h.T`) | `TestSendClassifiesGatewayOutageAs502` + renderPanelError tests |
| Rate limited (429) | Toast: "Too many requests — wait a moment…" | shell.js (htmx) / island i18n (login) | shell.test.mjs + session.test.mjs |
| Gateway outage (502) | Banner + toast; message/fax saved as failed | server | 502 test (HX-Trigger + banner pinned) |
| Other 4xx/5xx on htmx actions | Banner (`.wp-error` selected) + generic toast "HTTP N" | shell.js generic copy | contract test (config markers) |
| Network down (htmx) | Toast: "Network request failed…" | shell.js | shell.test.mjs |
| Network down (island session POST) | Toast: "Could not reach the server…" + `#log` line | island i18n `sessionNetFailed` | session.test.mjs |
| PBX rejects login (401 island REGISTER) | `loginError` inline + `reg-status` pill "registration rejected" | island i18n `loginError`/`regRejected` | i18n parity tests |
| SSE feed dead (3 consecutive errors) | One warn toast + pill label flips to "not connected" | island i18n `sseDropped`/`ssePillDown` | session.test.mjs |
| Unknown path (404) | Styled 404 (shell + error panel), status stays 404 | server `error.notfound` en/de | `TestNotFoundRendersTheShell` |
| Handler panic | Recovery middleware logs stack + re-raises; user gets htmx/browser failure surface | cqrshtmx.RecoveryMiddleware | library + server middleware tests |

Shell copy (toasts, dedup, throttle wording) stays ENGLISH by decision
D3 (2026-09-20): matches the `#log` operator-channel precedent; the
island's user-facing copy is fully en/de. Localizing shell copy only if
a tabs-style per-extension UX demand shows up.

**BDD posture** (plan T13): Ginkgo where it earns its keep — the
session behavior suites (`session_behaviors_test.go`) describe
observable auth behavior; table-driven Go tests everywhere else where
they are the clearer idiom; the island uses `node:test` black-box specs
driving real document listeners. No Ginkgo port of already-pinned error
paths (YAGNI, audit 2026-09-20). New BEHAVIOR surfaces should consider
Ginkgo DescribeTable when the subject is a state machine.

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
  The idiomorph experiment MERGED 2026-09-20: the stack browser E2E
  passed against the branch via `--override-input` (148 s, full
  call/transfer/DTMF/reconnect flow) — verdict doc
  `docs/research/2026-09-20_p25-idiomorph-morph-swap-verdict.md`,
  merge `6bb792e`.
  Adoption posture: middleware + assets only; the `setup` bundle, CQRS
  dispatch layer and usermgmt stay rejected (split-brain identity, see
  above); security presets are NEVER adopted wholesale — the library's
  `RecommendedPermissionsPolicy` denies `microphone`, which would kill
  the WebRTC phone. `toastDetail` is a type alias of
  `cqrshtmx.ToastDetail` (root-package type; NOT dispatch-layer — the
  old "kept local" comment was wrong), so a wire-shape change upstream
  fails this build. Trap: the dispatch-layer `Notify*` options emit
  `{level,message}`, NOT the island's `{message,kind}` shape. Their
  KIND vocabulary is dispatch-layer too: the server's `notifyToast`
  emits island kinds (ok/error/warn/info), which main.js's old
  `TOAST_KINDS` (copied from the dispatch vocabulary success/warning)
  silently recolored every success toast to info — the mapping now
  lives in ui.js `toastKindFor`, pinned by ui.test.mjs.
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
- Nix cross-build trust trap (2026-09-20): `nix build --system
  aarch64-linux` prints "ignoring the client-specified setting 'system',
  because it is a restricted setting and you are not a trusted user"
  when the client is untrusted (`trusted-users = root` here) — and on
  truly restricted paths it silently builds the DEFAULT system with
  EXIT=0 (a false-green aarch64 gate). Verify cross-builds by the ELF
  machine bytes (`od -An -tx1 -j18 -N2 <binary>`: `b7 00` =
  EM_AARCH64, `3e 00` = EM_X86_64), never by exit code alone. On this
  host the flake eval still honored aarch64 despite the warning — the
  trap is the silent variant elsewhere.
- Dynamic awk/grep patterns must ESCAPE `[`/`]`: release.sh's notes
  extractor matched `^## [2.4.0]` as a dynamic awk regex, where
  `[2.4.0]` is a bracket expression (one char of {2,.,4,0}) — it never
  matched the literal heading and shipped v2.3.0/v2.4.0 with EMPTY
  GitHub release bodies. The fold-check grep passed because ITS
  pattern had shell-escaped `\[`. Found by the 2026-09-20 release
  audit; both objects backfilled from CHANGELOG.
- Verify dependency internals at the CONSUMED tag (module cache or
  `git show v4.9.0:<path>`), never master: the 2026-09-18 audit
  over-credited v4.9.0's `ServeSSE` with a `retry:` hint that only
  exists on master — tag-checking before the port caught it. (True at
  the time; v4.11.0 DOES ship the hint — MD1's bump trigger fired and
  was executed 2026-09-20.)
- htmx/cqrs-htmx BUMP CHECKLIST (plan T25, 2026-09-20): re-verify at the
  new version (1) the `responseHandling` contract — entry keys `code`,
  `swap`, `error`, `target`, `select` in the SERVED htmx.min.js plus the
  htmx-config markers pinned in `TestServedPageHoldsTheDomContract`;
  (2) the `/events` `retry:` hint (`TestSSEStreamCarriesConnectedThenEvents`);
  (3) `HX-Trigger` still fires BEFORE the swap decision (the skip-guard
  in shell.js §3c depends on it); (4) the sse extension still fires
  cancelable `htmx:sseBeforeMessage`.
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
- SDK decision 2026-09-19 (SUPERB integration plan, all claims
  primary-source verified): KEEP sip.js 0.21.2; **JsSIP 3.13.8 is the
  named fallback** (npm-published 2026-05, actively maintained — the
  only maintained alternative stack; license nuance: npm says MIT,
  GitHub license field NOASSERTION). Swap ONLY on: Chromium WebRTC
  API breakage, a sip.js security advisory, or a needed capability —
  never speculatively; a swap is a full island call-path rewrite plus
  a stack browser-E2E re-run. All Go-side telephony REJECTED (sipgo
  signaling proxy, pion B2BUA/SBC, FreeSWITCH ESL — no maintained Go
  ESL client anyway): the browser terminates media regardless, so
  server-side signaling only adds state to the hottest path. Research
  table: docs/planning/2026-09-19_19-37_SUPERB-island-server-integration.md.
- Env config nests with `__`: `WEBPHONE_GATEWAY__MODE` →
  `gateway.mode`; single underscores stay literal (`WEBPHONE_DATA_DIR`
  → `data_dir`). Scalars via env; lists (`ice_servers`, `contacts`)
  and the `identities` ext→DID map via the JSON file.
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
  vulnix also range-matches distro-patched versions: against
  glibc-2.42-84 it prints 8 CVEs (2026-5450, 2025-15281, 2026-4046,
  2026-4437, 2026-5928, 2026-5435, 2026-6238, 2026-4438), and ALL EIGHT
  appear verbatim in the locked tree's glibc `2.42-master.patch`
  (verified 2026-09-19 by grepping the patch — NVD ranges cannot see
  patch suffixes; future rescans: grep the flagged CVE ids against the
  LOCKED rev's `nix eval github:NixOS/nixpkgs/<rev>#glibc.patches`
  before believing a finding; re-run clean 2026-09-20 — all 8 patched
  at rev `20b1ddd`). Runtime
  closure (8 derivations) carries zero real advisories (re-verified
  2026-09-20 with `--closure`; the exit-nonzero-on-findings shape is
  expected noise for this class).
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
- UI redesign invariants (2026-09-20): app.css is a token system — the
  `:root`/dark token blocks are MIRRORED between app.css and
  island/style.css; change both or the island drifts. `.sr-only` is
  OWNED by app.css (templ-components' Base emits Tailwind utility
  classes, but no Tailwind CSS loads, so `sr-only focus:not-sr-only`
  only works because app.css defines both). Avatars are
  `avatarFor`/`avatarHue` (views/helpers.go): country signum for
  numbers ("+1", "+4", "0"), word initials for names ("AK"), "?" for
  blanks — TrimSpace first, the tests caught the whitespace-only
  escape; hue is deterministic 0-359. Helpers ship WITH tests in the
  same commit: the untested first cut of `avatarFor` shipped a real
  bug. Direction chips: history `cdrDirGlyph`/`cdrDirLabel` and fax
  `faxDirGlyph`/`faxDirLabel` render "↓"/"↑" as aria-hidden chips with
  a sibling `role="img"` aria-label; loopback has no phone API so
  history rows never render locally — chips are pinned by tests, not
  screenshots. SSE payloads must keep the greppable row classes
  (`wp-thread-row`, `wp-bubble`, `wp-fax-row` —
  `TestSSEPushesSwapSafeFragments` greps them and forbids wrappers).
  The green dot in every screenshot is `#wp-sse-live` (session.js,
  JS-created so the served DOM contract stays untouched) — intentional
  live-feed indicator, not a bug. The contacts import row stays on one
  line via `flex: 1 1 220px` on its file input (not `flex-basis:
  100%`), and the island dial placeholder is the short
  "Number or extension" / "Nummer oder Durchwahl".
- Stack browser E2E flake mode (seen 2026-09-20): a SLOW run can die at
  the transfer step — FreeSWITCH hangs the call with
  RECOVERY_ON_TIMER_EXPIRE almost exactly 90s after DTLS-ready (ICE/media
  inactivity ceiling; no explicit timer is set in the stack's config),
  the island removes the dead call card, and the E2E's fresh
  `.transfer-btn` click goes StaleElementReference. Cold chromium
  caches plus boot-time transport reconnects on both browsers made the
  first run miss the window; the immediate re-run with warm caches
  passed the FULL flow in both call directions. Verdict rule:
  registration + DTMF + ICE stats all green before a ~90s death = flake,
  not an island regression — re-run once before digging.

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
   `python3 scripts/webphone-smoke.py`. If the train bumped
   cqrs-htmx or go-sse: also re-run `BenchmarkHubFanOut` and append a
   dated table to
   `docs/reviews/2026-09-18_hub-fanout-baseline.md` (MD1 rule; next
   re-run should use `-benchtime=1s -count=5` instead of `-benchtime
   2000x` for tighter numbers).
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
   The script then asserts the built ELF's machine bytes (`b700` =
   EM_AARCH64) because `--system` is a restricted setting an untrusted
   client's nix may silently ignore while exiting 0.
   Do NOT reach for `nix flake check --all-systems` as the fix: it is
   evaluation-only for other systems (verified 2026-09-19 — zero
   derivations built, "running 0 flake checks") and gates nothing.
   Checks DO cross-build if you name them:
   `nix build .#checks.aarch64-linux.island-lint` ran green cross-arch
   (oxlint substitutes from cache.nixos.org), so the aarch64 gate is
   explicit cross-builds of the package + the checks you care about.
9. **Closing sweep** (added 2026-09-20 after the §e lessons): any
   command that boots a server for verification ends by PROVING the
   process dead (`pgrep -f <pattern> || echo dead`) — and counts only
   processes YOU booted; a concurrent session's server (e.g.
   `/tmp/wpshoot/*`) is not yours to kill; CHANGELOG link
   edits outside a train get a `nix run nixpkgs#lychee -- .` run (the
   script covers post-push, not manual edits); the post-train tree
   sees `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` once (gitleaks and
   codespell are on-demand: `buildflow -s gitleaks`, `-s codespell`);
   and the daemon's last commits are verified pushed
   (`git ls-remote origin main` vs local HEAD).

## Concurrent sessions (observed 2026-09-20)

More than one Crush session can work this repo at once. Tell-tale:
uncommitted files you did not author (e.g. `internal/web/views/helpers.go`

- an untracked `helpers_test.go`) and mid-edit compile failures that heal
  on re-run. Rules: never revert/"fix" their in-flight files; re-read any
  shared file (i18n.go, pages.go, flake.nix) immediately before editing;
  a full-suite gate run may catch THEIR transient breakage — attribute
  failures before acting; and leave their booted dev servers running.

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
