# go-idempotency — library utilization audit

**Verdict:** NO-GO for PROD semantics, reasoned and standing — webphone's idempotency semantics are deliberately different (check-first / record-on-success), and the runtime capability already lives in webphone's stack via metaengine's `DedupStore`. But the library is no longer absent: its **contract suite is a TEST-side consumer** since the T18 train — the 18-invariant suite audits the REAL journal-backed window, which is exactly the "residual value" this report first filed as optional. **Adoption: test-side contract pin (LANDED 2026-11-04); prod stays by-decision absent.**
**Pin:** v0.6.0 (go.mod:39, **direct require since the T18 contract pin**; previously `// indirect`; walked v0.3.2 → v0.4.0 → v0.5.0 → v0.6.0 by the 2026-11-18 fleet treadmill) · **Latest:** v0.6.0 · **Currency:** current · **Status:** direct (test-only) — `internal/server/hooks/idempotency_contract_test.go` is the sole import; prod consumption stays transitive via cqrs-htmx (root + usermgmt) and go-cqrs-lite/middleware
**Audited:** 2026-11-04 · **Updated:** 2026-11-13 (the T18 test-side adoption recorded — it landed 23:33 the same day this report was written and the body never learned it; prod verdict unchanged). · **Re-verified:** 2026-11-22 (v0.6.0 currency + premise re-audit — NO-GO unchanged, test-side value grew). · **Prior reports:** none (first audit).

## Status

|             |                                                             |
| ----------- | ----------------------------------------------------------- |
| Module      | `github.com/larsartmann/go-idempotency`                     |
| In webphone | direct require, test-side only; **zero production imports** |

## What the library offers

`Store` interface (`Seen(ctx,key)`, `Record(ctx,key,ttl)`, `CheckAndRecord(ctx,key,ttl) error` — atomicity contract on the third); sentinels `ErrDuplicate` (→409, non-retryable) and `ErrInvalidTTL`; `Forever` — a `math.MaxInt64` TTL sentinel for permanent claims (v0.4.0); deprecated in-memory store for dev (`NewMemoryStore(sweepInterval, ...MemoryStoreOption)` since v0.6.0 — `WithClock` injects its clock; the old two-arg call sites compile unchanged, and the store keeps its `Deprecated:` notice); `middleware.HTTP(store, ttl, next)` (400 missing key / 409 dup / 503 store-down, `Idempotency-Key` header); `contract.RunTests` — 18 invariants (13 TTL-window + 5 `Forever` permanent-claim) incl. 200-goroutine atomicity; the `idemtest` consumer-drill kit as a separate nested module (v0.5.0). No production backends by design (the README points stacks with real stores at their own implementation).

## Webphone's own answer (the overlap)

- `internal/server/hooks/idempotency.go`: `DedupWindow` port (`Seen`/`CheckAndRecord`/`Sweep`) implemented by an adapter over **metaengine's DedupStore** (`DedupCheckAndRecord`/`DedupSeen`/`DedupSweep`, internal/app/dedup.go) with a TTL-map fallback for tests; 1h TTLs for hooks and calls idem windows; consumers: hooks/webhooks.go, api/{calls,summary,review,disposition}_api.go; wired at internal/server/server.go:442/453 (`deps.Dedup` ← internal/app/app.go:511 — the line moved off server.go:318 in later trains).
- Semantics: **check-first, record-on-success** — a failed/5xx attempt stays retryable, only successes are recorded, and a replay answers the original outcome (`duplicate:true`), per AGENTS.md idem doctrine.

## How webphone uses it (test-side, since T18)

`TestDedupWindowSatisfiesIdempotencyContract` (internal/server/hooks/idempotency_contract_test.go) runs the library's `contract.RunTestsStrict` suite against the REAL dedup window — a fresh `system.db` journal per subtest:

- `contractWindow` adapts the hooks `DedupWindow` port (collection-named, caller-injected clock) onto the library's flat-key `Store` interface; the `Record` shim rides `CheckAndRecord`'s duplicate verdict, which is a no-op on an existing key — exactly what the invariant suite asserts.
- `engineWindow` is the metaengine `DedupStore` twin of the production adapter in internal/app (`engineDedupWindow`) — the copy exists only because hooks cannot import app.
- `TimingScale: 5` keeps the wall-clock expiry subtests honest on loaded runners (ci-flake-ledger row, run 42407471289); the durable fix is upstream — a contract-suite clock seam (go-idempotency#15). v0.6.0's `WithClock` is NOT it: it pins only the deprecated MemoryStore, which this suite never sees (the window under audit is the metaengine journal store), and `[Unreleased]`'s `PinnedClock` drill is idemtest-side.

A red there means the idem-gate seam's semantics drifted from the standard idempotency contract the webhook and call lanes depend on. This is the same pattern as go-codec's test-side tombstone pin: the library audits webphone's seam, webphone's prod path stays its own.

## Why NOT to adopt in prod (each a claim)

1. **Semantic mismatch.** `CheckAndRecord` fail-closed 409-on-duplicate contradicts record-on-success: with the library, a transport failure after the key was recorded would permanently 409 the retry — exactly what webphone's outage-retryable design forbids. Webphone would have to use only `Seen`+`Record`, the racy pair the library's own README warns against, gaining nothing.
2. **The capability already exists in-tree.** metaengine's `DedupStore` IS the durable window (the upstream README itself says cqrs-lite stacks don't need this module). Adopting it would put two idempotency truths side by side.
3. **HTTP-key semantics differ.** The library's middleware derives keys from an `Idempotency-Key` header; webphone's keys are derived server-side (provider refs, call ids) — no client-facing key contract to enforce.

## Residual value — LANDED

The original 2026-11-04 report filed this as "optional hardening, not scheduled"; it landed the same day (T18, commit 7d0af3b6): 13 concurrency/atomicity invariants for ~40 lines of test glue, no runtime dependency added — see above. The v0.4.0 pin walk added the five `Forever` invariants for free (18 total), widening the audit onto permanent-claim semantics against the same journal window.

**Refresh trigger:** a client-facing `Idempotency-Key` contract on a public API; a second consumer needing generic key/ttl windows outside metaengine's reach.

## 2026-11-22 re-verify

Currency: pin v0.6.0 = remote latest (local checkout tag v0.6.0); `[Unreleased]` is idemtest-side only (`RequireAllowsAfterExpiryPinned` + the thread-safe `PinnedClock` double, shipped as `idemtest/v0.1.1`). The pin walked v0.3.2 → v0.4.0 → v0.5.0 → v0.6.0 during the 2026-11-18 fleet treadmill — three upstream releases, all dated 2026-11-17:

- **v0.4.0 — `Forever` permanent claims**: the `math.MaxInt64` TTL sentinel plus five new contract invariants (and six negative scenarios) pinning permanent-claim semantics; the documented backend examples switched to reclaimable conditional upserts.
- **v0.5.0 — the `idemtest` kit**: consumer-side drills (exactly-once retry storm, TOCTOU, fail-closed, TTL expiry) plus spies and counters, as a separate nested Go module. **NOT adopted here** — no `idemtest` require line in webphone's go.mod and zero imports (rg): the kit's drills target dispatch pipelines over a store wired with `WithClock`, a shape webphone's journal-backed window doesn't have. The root contract suite stays the sole consumer. (Note: upstream's CHANGELOG records a fleet library-policy companion rule — "uses go-idempotency → must have `idemtest`" — which would flag webphone as-is; the non-adoption is reasoned above, but the policy mismatch is worth an owner glance.)
- **v0.6.0 — `WithClock`**: `NewMemoryStore` takes variadic options and the (still-deprecated) store's clock is injectable; old call sites compile unchanged.

**Did any premise move? No.** The prod NO-GO stands: `CheckAndRecord`'s fail-closed claim semantics are untouched (v0.4.0's doc fixes made the claim contract stricter, not softer), `middleware.HTTP(store, ttl, next)` keeps its signature, MemoryStore keeps its `Deprecated:` notice (store.go:150), and the no-production-backends stance holds — metaengine's `DedupStore` remains the in-tree capability. The test-side value moved the OTHER way — it grew for free: `RunTestsStrict` audits **18 invariants** now, and the `Forever` subtests push `Record(key, math.MaxInt64)` through `contractWindow` onto the journal store, where claimkit computes `now.Add(ttl)` and encodes it in the dialect's fixed-width storage form (go-cqrs-lite metaengine/v4/claimkit/dedup.go:87) — `time.Time` represents year ~2318 fine and there is no UnixNano path, so the permanent claim lands as a far-future expiry and the widened suite holds by construction (static evidence; no `go test` per audit rules). One hoped-for fix did NOT land: the contract-suite clock seam stays open (go-idempotency#15), so `TimingScale: 5` stays load-bearing. Verify block replayed green — 0 production imports, `RunTestsStrict` at idempotency_contract_test.go:95, pin v0.6.0 at go.mod:39.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-idempotency" go.mod                            # expect v0.6.0 — direct pin since T18 (test-only consumer)
rg -n "go-idempotency" --type go -g '!*_test.go' internal/ cmd/ | wc -l # expect 0 production imports (the T18 contract adapter is test-only)
rg -n "RunTestsStrict" internal/server/hooks/idempotency_contract_test.go
rg -n "type DedupWindow" internal/server/hooks/idempotency.go
rg -n "DedupCheckAndRecord" internal/app/dedup.go | head -2
```
