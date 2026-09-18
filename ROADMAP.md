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
  accept/reject wiring fix). Residual: push the fix and bump the
  stack's input lock to it (TODO_LIST).

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
- Decision records / doc notes (P7): StructuredError for `/api/session`
  (adopt only if the island branches on codes); sync/ multi-tab module
  N.A. (per-tab SIP UA by design); DecodePagination N.A. (cursor
  `older=` semantics); hub fan-out benchmark baseline; hx-boost
  non-adoption; Default vs JSONLogFormatter for the stack's sink;
  readiness-body contract for the stack's probes; notify/ack
  applicability close-out; cross-link the deep-dive and structural-health
  reports; ClientIP-trust note upstream in httputil if XFF turns out
  sanitized; re-diff cqrs-htmx master vs v4.9.0 for new middleware.
