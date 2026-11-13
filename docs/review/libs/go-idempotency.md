# go-idempotency — library utilization audit

**Verdict:** NO-GO for PROD semantics, reasoned and standing — webphone's idempotency semantics are deliberately different (check-first / record-on-success), and the runtime capability already lives in webphone's stack via metaengine's `DedupStore`. But the library is no longer absent: its **contract suite is a TEST-side consumer** since the T18 train — the 13 invariant suite audits the REAL journal-backed window, which is exactly the "residual value" this report first filed as optional. **Adoption: test-side contract pin (LANDED 2026-11-04); prod stays by-decision absent.**
**Pin:** v0.3.2 (go.mod:39, **direct require since the T18 contract pin**; previously `// indirect`) · **Latest:** v0.3.2 · **Currency:** current · **Status:** direct (test-only) — `internal/server/hooks/idempotency_contract_test.go` is the sole import; prod consumption stays transitive via cqrs-htmx (root + usermgmt) and go-cqrs-lite/middleware
**Audited:** 2026-11-04 · **Updated:** 2026-11-13 (the T18 test-side adoption recorded — it landed 23:33 the same day this report was written and the body never learned it; prod verdict unchanged). · **Prior reports:** none (first audit).

## Status

|             |                                         |
| ----------- | --------------------------------------- |
| Module      | `github.com/larsartmann/go-idempotency` |
| In webphone | direct require, test-side only; **zero production imports** |

## What the library offers

`Store` interface (`Seen(ctx,key)`, `Record(ctx,key,ttl)`, `CheckAndRecord(ctx,key,ttl) error` — atomicity contract on the third); sentinels `ErrDuplicate` (→409, non-retryable) and `ErrInvalidTTL`; deprecated in-memory store for dev; `middleware.HTTP(store, ttl, next)` (400 missing key / 409 dup / 503 store-down, `Idempotency-Key` header); `contract.RunTests` — 13 invariants incl. 200-goroutine atomicity. No production backends by design (the README points stacks with real stores at their own implementation).

## Webphone's own answer (the overlap)

- `internal/server/hooks/idempotency.go`: `DedupWindow` port (`Seen`/`CheckAndRecord`/`Sweep`) implemented by an adapter over **metaengine's DedupStore** (`DedupCheckAndRecord`/`DedupSeen`/`DedupSweep`, internal/app/dedup.go) with a TTL-map fallback for tests; 1h TTLs for hooks and calls idem windows; consumers: hooks/webhooks.go, api/{calls,summary,review,disposition}_api.go; wired at internal/server/server.go:318.
- Semantics: **check-first, record-on-success** — a failed/5xx attempt stays retryable, only successes are recorded, and a replay answers the original outcome (`duplicate:true`), per AGENTS.md idem doctrine.

## How webphone uses it (test-side, since T18)

`TestDedupWindowSatisfiesIdempotencyContract` (internal/server/hooks/idempotency_contract_test.go) runs the library's `contract.RunTestsStrict` suite against the REAL dedup window — a fresh `system.db` journal per subtest:

- `contractWindow` adapts the hooks `DedupWindow` port (collection-named, caller-injected clock) onto the library's flat-key `Store` interface; the `Record` shim rides `CheckAndRecord`'s duplicate verdict, which is a no-op on an existing key — exactly what the invariant suite asserts.
- `engineWindow` is the metaengine `DedupStore` twin of the production adapter in internal/app (`engineDedupWindow`) — the copy exists only because hooks cannot import app.
- `TimingScale: 5` keeps the wall-clock expiry subtests honest on loaded runners (ci-flake-ledger row, run 42407471289); the durable fix is upstream — an injectable clock instead of wall sleeps.

A red there means the idem-gate seam's semantics drifted from the standard idempotency contract the webhook and call lanes depend on. This is the same pattern as go-codec's test-side tombstone pin: the library audits webphone's seam, webphone's prod path stays its own.

## Why NOT to adopt in prod (each a claim)

1. **Semantic mismatch.** `CheckAndRecord` fail-closed 409-on-duplicate contradicts record-on-success: with the library, a transport failure after the key was recorded would permanently 409 the retry — exactly what webphone's outage-retryable design forbids. Webphone would have to use only `Seen`+`Record`, the racy pair the library's own README warns against, gaining nothing.
2. **The capability already exists in-tree.** metaengine's `DedupStore` IS the durable window (the upstream README itself says cqrs-lite stacks don't need this module). Adopting it would put two idempotency truths side by side.
3. **HTTP-key semantics differ.** The library's middleware derives keys from an `Idempotency-Key` header; webphone's keys are derived server-side (provider refs, call ids) — no client-facing key contract to enforce.

## Residual value — LANDED

The original 2026-11-04 report filed this as "optional hardening, not scheduled"; it landed the same day (T18, commit 7d0af3b6): 13 concurrency/atomicity invariants for ~40 lines of test glue, no runtime dependency added — see above.

**Refresh trigger:** a client-facing `Idempotency-Key` contract on a public API; a second consumer needing generic key/ttl windows outside metaengine's reach.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-idempotency" go.mod                            # direct pin since T18 (test-only consumer)
rg -n "go-idempotency" --type go -g '!*_test.go' internal/ cmd/ | wc -l # expect 0 production imports (the T18 contract adapter is test-only)
rg -n "RunTestsStrict" internal/server/hooks/idempotency_contract_test.go
rg -n "type DedupWindow" internal/server/hooks/idempotency.go
rg -n "DedupCheckAndRecord" internal/app/dedup.go | head -2
```
