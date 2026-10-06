# Webhook idempotency durability — verdict: in-memory is sufficient

- **Date:** 2026-09-22 17:10 CEST
- **Plan:** SUPERB T27e ("webhook-idempotency durability decision:
  post-restart replay acceptable?")
- **Verdict:** YES — post-restart replay is acceptable. The idem store
  stays in-memory (1h TTL). Persisting dedup keys is REJECTED.

> ARCHIVED 2026-10-06 (docs-health v7 sweep): verdict standing — the
> in-memory `hooksIdem` (1h TTL, successes only, replay 202) is the
> shipped behavior, re-verified by the byte-exact part-header and
> webhook contract tests since. Per-item verdicts inline.

## The analysis

`hooksIdem` (in-memory, `hookIdempotencyTTL` = 1h, only successes
recorded) exists to absorb the provider's BURST retries (Telnyx
re-delivers within minutes). What actually happens on a replay that
slips THROUGH the window (restart mid-burst, or a >1h-late delivery):

- Both status hooks carry TERMINAL statuses only — message
  `delivered|failed`, fax `transmitted|failed`. A replay re-applies a
  same-value SET: `UpdateOutboundStatus` /
  `UpdateProviderStatus` converge, no counter moves twice, no state
  regresses.
- A cross-restart replay of a DIFFERENT status for the same
  `provider_ref` (failed, restart, delivered) is not a replay artifact
  at all — it is a legitimate late status change, and applying it is
  the CORRECT behavior; persisting dedup keys would make us IGNORE it
  (worse).
- The 410/422-shaped provider misbehaviors fail the handler BEFORE
  the idem record (failures stay retryable by design).

## Why not persist

Persisted keys would add a store round-trip on the provider-facing
hot path to guard exclusively against same-value no-op writes. That
trade buys nothing observable and costs latency + a new table. The
only theoretical residue — duplicate `wp-thread-row` SSE nudges for a
replayed status — is morph-absorbed (idempotent swap) and invisible.

## Pin

~~This decision rides the existing idempotency test surface
(`hooksIdem` replay → `202` inert) and the convergence argument
above.~~ done — standing: the replay-202 pin is part of the live
webhook suite; revisit trigger unchanged (non-terminal statuses).
