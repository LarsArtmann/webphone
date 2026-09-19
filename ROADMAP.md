# ROADMAP

Long-term direction and raw ideas not yet refined into actionable tasks.
Actionable work lives in TODO_LIST.md; shipped work in FEATURES.md.

## Open ideas

- Union test coverage (`-coverpkg=./...`) in the buildflow test-coverage
  step: per-package attribution hides cross-package coverage (tests in
  `internal/server` exercising `internal/store` count as 0% for the
  store). Blocked on a BuildFlow feature — the canonical
  `.buildflow.yml` key list has no per-step-args mechanism, and a
  repo-local coverage script would duplicate a step BuildFlow already
  orchestrates. Fleet-wide value → belongs upstream.
- sip.js: no 0.22 exists; upstream is dormant at 0.21.2 (an
  unreleased 0.21.3 tag adds only a SimpleUser option). Evaluation:
  `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`. Revisit only on
  a reconnect-hang fix, a security advisory, or a needed capability —
  any future bump re-runs the upstream browser E2E.
- Stack-side switchover of nix-international-telephony onto this
  service: DONE 2026-09-18 (stack imports `nixosModules.default`,
  nginx vhost proxies the service, browser E2E green after the
  accept/reject wiring fix). Fully resolved 2026-09-19 with the v2.1.0
  release: the stack input rides the v2.1.0 tag commit `d815004`
  (stack commit `2289e89`), and the stack's FULL `nix flake check` is
  green with that lock (browser E2E, webphone VM test included).

## WORTH_CONSIDERING cluster (one-line specs, SUPERB plan P26 2026-09-19)

- Session persistence: move the in-memory TTL session store behind a
  restart-survivable backing (SQLite table + TTL sweep) — the island
  login flow is unchanged; weigh against "sessions are ephemeral by
  design" before building.
- Retention/cleanup job: bounded deletion for old CDR rows, read
  faxes/voicemail blobs, and expired sessions (a `retention_days`
  setting + a systemd timer in the module).
- PWA: manifest + service worker so the webphone installs to a home
  screen; the SIP island must survive SW caching rules (no cache for
  `/events`, verbatim island modules pinned by hash).
- Video calls: SIP.js video negotiation + a `<video>` call card —
  FreeSWITCH side needs a video-capable profile; large surface, only
  on demand.
- Recording UI (see raw ideas below): product-intent decision first
  (consent/jurisdiction), then the panel.

## Standing watches (SUPERB plan P27 2026-09-19)

Drift gets caught by routine, not luck — each row names the trigger to
re-check:

- sip.js: revisit on a 0.22 release, a reconnect-hang fix, or a
  security advisory (evaluation: `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`).
- templ-components: ship a ThemeScript opt-out knob and this repo drops
  the CSP hash pin plus the app.css `!important` color-scheme rules.
- oxlint globals watchlist: any new browser global in the island needs
  an entry in `internal/web/assets/island/oxlint.json` (the gate fails
  closed on undeclared identifiers by design).
- E2E wall-time budget: the stack browser E2E baseline is ~150s; a run
  drifting far above it is a perf regression signal, not noise.
- cqrs-htmx root tag: the next release after v4.9.0 changes the
  `/events` byte stream (master already adds the SSE `retry:` hint) —
  on bump, re-run the browser E2E and update the AGENTS retry note.

## cqrs-htmx adoption long tail (plan P5-P7, 2026-09-18)

Source: `docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`
§Phases 5-7 (items below are the road to ~100% adoption; P0-P3 landed
2026-09-18, P4 + small P5 tasks graduated to TODO_LIST).

- OOB badge push spike (OO1-OO3): `hx-swap-oob` unread-badge fragment
  inside the `threads` SSE payload, env-flag-gated, validated against the
  stack browser E2E before any default-on. UB1 (badge push from
  unreadCache invalidation) only if the spike adopts.
- Toasts (TO1-TO3): island `HX-Trigger` listener rendering the library's
  ToastDetail shape; server sets Notify headers on send/save/delete;
  i18n copy in BOTH en/de maps.
- `/version` endpoint via the library `DebugHandler` pattern (VE1);
  Server-Timing middleware behind an env flag (ST1); OpenAPI 3.1 for
  `/api/session` at `/openapi.json` (OA1/OA2); compose the root stack
  with `cqrshtmx.Chain` + ordering parity test (CH1).
- Hardening: fuzz `/hooks/*` JSON decoding (FZ1/FZ2); ordering-invariant
  test pinning limiter-wraps-secret-gate (OR1); island 429/Retry-After
  handling in phone-api fetch wrappers (RA1); session TTL sweeper
  interaction test (TT1); cqrs-htmx transitive drift check + root
  v4.10.0 bump trigger note (DR1); periodic vulnix runtime-closure rescan
  (last: buildflow 2026-09-18, clean).
  STATUS 2026-09-18: fuzz target shipped (844k execs clean), OR1/RA1/TT1
  shipped; DR1 checked — no newer releases of cqrs-htmx/httputil/go-sse;
  VL1 re-verified — runtime closure still 8 derivations, unchanged glibc.
- Root-tag bump trigger (MD1 finding, 2026-09-18): master already adds
  the SSE `retry:` hint to `Broadcaster.ServeSSE` that v4.9.0 lacks —
  the next root tag changes the `/events` byte stream. On bump: update
  the AGENTS.md retry note, re-run `BenchmarkHubFanOut`, and re-run the
  upstream browser E2E (payload-shape rule).
- Island JS test runner (standing gap, 2026-09-19): the island has
  none — toasts listener, live pill and 429 surfacing are pinned by Go
  asset tripwires + code review only (§b.3 of the 2026-09-19 status).
  When a runner lands, port the tripwires into real DOM tests.
- Decision records / doc notes (P7): StructuredError for `/api/session`
  (adopt only if the island branches on codes); sync/ multi-tab module
  N.A. (per-tab SIP UA by design); DecodePagination N.A. (cursor
  `older=` semantics); hub fan-out benchmark baseline; hx-boost
  non-adoption; Default vs JSONLogFormatter for the stack's sink;
  readiness-body contract for the stack's probes; notify/ack
  applicability close-out; cross-link the deep-dive and structural-health
  reports (both live in-tree under `docs/` — a cross-link adds nothing;
  closed as NOT-DO 2026-09-19); ClientIP-trust note upstream in httputil
  if XFF turns out sanitized; re-diff cqrs-htmx master vs v4.9.0 for new
  middleware (checked 2026-09-18 — only the SSE `retry:` hint, see MD1).

## Open questions (owner calls)

- Rate-limit keys: does the stack's proxy sanitize `X-Forwarded-For`?
  Gates the flip from port-stripped peer-host keys
  (`remoteHostKey`) to `KeyExtractorFromClientIP`; standing since the
  2026-09-18 adoption (safe default shipped, flip rule documented).
- Stack `webphone` input policy: pin tag refs (`?ref=v2.x`) with
  explicit bumps vs track main (today: tracks main — the next
  `nix flake update` silently moves the deployment input).
- GitHub Release objects per tag (notes/visibility) vs tags +
  CHANGELOG only (v2.0.0 has a tag, no Release object).
- Go module-path policy for v2+ tags: `/v2` suffix vs NOT-DO record
  (a flake-consumed app never `go get`s itself).
- Loopback gateway semantics: simulate `delivered` (like fax loopback
  resolves `transmitted`) or keep `sent` as the honest terminal state?
- Handler gating: keep the deliberate dual layer (`requireSession`
  helpers + `Sessions.Require` on `/events`, `/phone-api/`) or drop
  the middleware wiring?
- Recordings in the webphone UI (see raw ideas below) — product
  intent, consent/jurisdiction posture, per-extension vs operator
  access.
- Stack-side: keep or revert the TEMP-DIAG answer-phase dump in the
  stack's browser E2E (commit `b96d4c2` there).

## Harvested raw ideas (2026-09-19 docs-health sweep)

Grouped from the annotated 2026-09-18/19 status reports; nothing here
is committed work — refine into TODO_LIST only on demand.

- Recording integration (PBX records everything today; webphone has no
  surface): Recordings panel/in-island playback, live REC indicator,
  per-extension access via the phone-api proxy pattern, `*97` opt-out
  visibility, recorded-badge on CDR rows, retention surfacing,
  stereo-WAV compression job, consent/privacy README note.
- Browser-level gates: headless-browser console-cleanliness check on
  `/` (the gate hole that let three CSP violations ship), tab-SSE
  browser test, two-browser E2E under CPU-constrained TCG.
- Testing long tail: SSE handler edge tests (anonymous 401, heartbeat
  on the wire), store/domain edge-case round-out (contacts CRUD,
  branded-ID property round-trips), `-race` stress of SSE hubs,
  delivery-receipt SSE-event assertion, comment-vs-code contract
  sweep, provider error-text pinning, periodic `art-dupl` ritual.
- Messaging polish: persist + display failure reasons for failed
  messages (parity with fax), distinct delivered badge style (see
  TODO_LIST), random loopback provider refs (drop UnixNano), image
  thumbnails, draft persistence, fax cover pages.
- Calls/UX: shortcut help overlay, new-conversation UX (open the
  thread after first send), a11y pass (focus order post-swap, aria-live
  SSE regions), multiple-tab glare warning, SSE connection-loss banner
  beyond the live pill.
- Platform: short-lived TURN REST credentials via `/config.js`,
  metrics endpoint, SQLite backup/restore runbook + drill,
  per-extension data export, webhook payload versioning header,
  timezone-aware timestamps, server-side PDF page counting, MIME
  sniffing on attachments, rate-limit tuning knobs (only on operator
  demand), CSP nonce mode, `/favicon.ico` route for non-browser
  clients, reusable "HTMX tabs + island + session" pattern doc
  (third LarsArtmann app with this shape), i18n dynamic-template
  key-sync, island/app.css shared-token extraction, webhook-idempotency
  durability decision (memory TTL means a post-restart provider replay
  re-applies — acceptable?), OpenAPI boundary decision (extend beyond
  `/api/session` or record as deliberate), limiter-key widening runbook
  line (NAT offices), signed tags (`git tag -s`), nixpkgs lock + vulnix
  rescan cadence, daemon-config exclusion of `docs/status/` from
  heuristic commits (upstream infra decision), wrapper-flake bisect
  trick write-up (module-from-HEAD + package-from-rev).

## Session persistence (raw idea, owner call)

Sliding-session TTL refresh (extend expiry on activity) with CSRF token
rotation at each refresh point, riding the island's existing adoption path
(`GET /api/csrf`; see the SUPERB plan P23 verdict for why this is deferred).
