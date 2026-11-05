# go-idempotency — library utilization audit

**Verdict:** NO-GO, reasoned. Webphone's idempotency semantics are deliberately different (check-first / record-on-success), and the capability already lives in webphone's stack via metaengine's `DedupStore`. **Adoption: n/a (absent from product by decision).**
**Pin:** v0.3.2 (go.mod:102, `// indirect`) · **Latest:** v0.3.2 · **Currency:** current (at the indirect pin) · **Status:** indirect-only — pulled by cqrs-htmx (root + usermgmt) and go-cqrs-lite/middleware
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

|             |                                         |
| ----------- | --------------------------------------- |
| Module      | `github.com/larsartmann/go-idempotency` |
| In webphone | indirect only; **zero imports**         |

## What the library offers

`Store` interface (`Seen(ctx,key)`, `Record(ctx,key,ttl)`, `CheckAndRecord(ctx,key,ttl) error` — atomicity contract on the third); sentinels `ErrDuplicate` (→409, non-retryable) and `ErrInvalidTTL`; deprecated in-memory store for dev; `middleware.HTTP(store, ttl, next)` (400 missing key / 409 dup / 503 store-down, `Idempotency-Key` header); `contract.RunTests` — 13 invariants incl. 200-goroutine atomicity. No production backends by design (the README points stacks with real stores at their own implementation).

## Webphone's own answer (the overlap)

- `internal/server/hooks/idempotency.go`: `DedupWindow` port (`Seen`/`CheckAndRecord`/`Sweep`) implemented by an adapter over **metaengine's DedupStore** (`DedupCheckAndRecord`/`DedupSeen`/`DedupSweep`, internal/app/dedup.go) with a TTL-map fallback for tests; 1h TTLs for hooks and calls idem windows; consumers: hooks/webhooks.go, api/{calls,summary,review,disposition}_api.go; wired at internal/server/server.go:318.
- Semantics: **check-first, record-on-success** — a failed/5xx attempt stays retryable, only successes are recorded, and a replay answers the original outcome (`duplicate:true`), per AGENTS.md idem doctrine.

## Why NOT to adopt (each a claim)

1. **Semantic mismatch.** `CheckAndRecord` fail-closed 409-on-duplicate contradicts record-on-success: with the library, a transport failure after the key was recorded would permanently 409 the retry — exactly what webphone's outage-retryable design forbids. Webphone would have to use only `Seen`+`Record`, the racy pair the library's own README warns against, gaining nothing.
2. **The capability already exists in-tree.** metaengine's `DedupStore` IS the durable window (the upstream README itself says cqrs-lite stacks don't need this module). Adopting it would put two idempotency truths side by side.
3. **HTTP-key semantics differ.** The library's middleware derives keys from an `Idempotency-Key` header; webphone's keys are derived server-side (provider refs, call ids) — no client-facing key contract to enforce.

## Residual value

One small, cheap pin: adapt `DedupWindow` → the library's `Store` interface and run `contract.RunTests` against it in a test — 13 concurrency/atomicity invariants for ~40 lines of test glue, no runtime dependency added. Filed as an optional hardening, not scheduled.

**Refresh trigger:** a client-facing `Idempotency-Key` contract on a public API; a second consumer needing generic key/ttl windows outside metaengine's reach.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-idempotency" go.mod                            # indirect pin
rg -n "go-idempotency" --type go -g '!*_test.go' internal/ cmd/ | wc -l # expect 0 production imports (the T18 contract adapter is test-only)
rg -n "type DedupWindow" internal/server/hooks/idempotency.go
rg -n "DedupCheckAndRecord" internal/app/dedup.go | head -2
```
