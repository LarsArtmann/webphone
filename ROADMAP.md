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
  accept/reject wiring fix). Residual resolved 2026-09-19: the stack
  input is bumped to the v2.0.0 tag commit (stack commit `fc6bc81`,
  package builds green). Still open: the stack's FULL `nix flake check`
  with the new lock (FreeSWITCH/operator derivations unexercised) —
  TODO_LIST.

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
  key-sync, island/app.css shared-token extraction.
