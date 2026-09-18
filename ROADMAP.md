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
- sip.js 0.22: evaluation report in `docs/reviews/`; an actual bump
  also needs the upstream browser E2E re-run (see TODO_LIST).
- Stack-side switchover of nix-international-telephony onto this
  service (input swap, WSS proxy, config migration, E2E re-run) —
  decided in principle, execution pending (TODO_LIST).
