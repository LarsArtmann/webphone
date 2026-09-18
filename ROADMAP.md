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
