# Draft — go-idempotency: the contract suite's TTL subtests race the wall clock under -race

**Target repo:** `LarsArtmann/go-idempotency` (own repo) · **Kind:** bug
report body · **Provenance:** unsolicited (AI-found; banner required) ·
**Status:** DRAFT, refreshed 2026-11-21 against v0.6.0 (originally drafted
2026-11-08 against v0.3.2 and held because the sweep plan believed
`WithClock` would retire the band-aid — it does not reach the contract
package; see body). Evidence chain: webphone `docs/ci-flake-ledger.md`
row 42407471289 + dependency-currency sweep train 4.

**Proposed title:** `contract suite: NoopOnExistingKey fails under -race when a scheduling stall outlasts the scaled shortTTL — the two Record calls race the wall clock`

```markdown
> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: The contract suite's TTL subtests sleep real wall-clock time, so a consumer running the suite over its own backend still races the scheduler — under `-race`, `NoopOnExistingKey` holds only while its two `Record` calls land within `shortTTL`; a scheduling stall between them lets the second `Record` legitimately re-claim the expired window, and the assertion reads correct store behavior as "Record extended the TTL". v0.6.0's `WithClock` fixed this class for `MemoryStore` and the library's own suite, but it is a `MemoryStoreOption` — the `contract` package still has no clock seam. Ask: the same logical-time seam for the contract suite (`Options.Clock`-shaped), so consumers drive the TTL subtests deterministically.

## Problem

Consuming repo: LarsArtmann/webphone, CI run 42407471289 (2026-11-07),
`go test -race`, subtest
`TestDedupWindowSatisfiesIdempotencyContract/Record/NoopOnExistingKey`.

The invariant (go-idempotency `contract/contract_invariants.go`, `runRecordTests` — unchanged in v0.6.0):

    t.Run("NoopOnExistingKey", func(t *testing.T) {
        store.Record(ctx, "noop-key", tm.shortTTL)
        store.Record(ctx, "noop-key", tm.longTTL)   // must NOT extend
        time.Sleep(tm.waitAfter)                     // past shortTTL
        if mustSeen(t, store, ctx, "noop-key") {
            t.Fatal("Record extended the TTL: ...")
        }
    })

The assertion is sound ONLY if both `Record` calls land while the first
window is still live — i.e. within `shortTTL` of the first Record. Under
`-race` a scheduling stall between the two calls can outlast the scaled
`shortTTL` (60 ms at the consumer's then-TimingScale 3). The second
`Record` then hits the store's `expires_at <= now` guard and legitimately
re-claims a fresh `longTTL` — correct per `CheckAndRecord`/`Record`
semantics — and the subtest fails with "Record extended the TTL".

## Why the existing levers do not fix it

- `Options.TimingScale` (contract/contract.go) stretches every wall-clock
  timing, but the stall budget scales WITH the timings — a sufficiently
  slow `-race` run can always outlast `shortTTL`, whatever the scale. The
  docs say it themselves: "TTL-based subtests sleep on wall-clock
  timings". Webphone hardened from TimingScale 3 to 5 (stall budget
  60 → 100 ms) the same day; that widens the window, it does not close
  the race class.
- v0.6.0's `WithClock` (issue #14) pins the clock of a `MemoryStore`
  (`clock.go`, a `MemoryStoreOption`) and made the library's OWN suite
  deterministic — but consumers run `contract.RunTestsStrict` over their
  own backends through a `StoreFactory`; the suite's `Options` still
  carries only `TimingScale` (contract.go), and expiry is driven by
  `time.Sleep(tm.waitAfter)` at four sites in `contract_invariants.go`.
  Webphone audits its journal-backed sqlite window (go-cqrs-lite
  metaengine `DedupStore`, caller-injected `now`) — no `MemoryStore`
  exists on that path to pin, and the wall-clock bet moves into the
  consumer adapter unchanged.

## Evidence of the flake class (not a regression)

- Run 42407471289 red on this subtest; the SAME hooks code was green on
  the four prior runs and on the docs-only successor run 42407606027.
- 3× local `-race` green on the identical tree (and 3× again on the
  v0.6.0 pin, 2026-11-21 — the band-aid holds, it just is a band-aid).
- Verified at go-idempotency v0.6.0 source: `contract_timing.go`
  (wall-clock `timings`, scaled sleeps) and `contract_invariants.go`
  (the four `time.Sleep(tm.waitAfter)` sites); `contract.Options` has
  exactly one field, `TimingScale`.

## Ask

A deterministic seam for the contract suite's TTL subtests — for example
an `Options.Clock` that clock-aware factories share (the suite steps the
clock instead of sleeping when set), or an equivalent logical-time
arrangement. A documented stall-tolerant reformulation of
`NoopOnExistingKey` that asserts the invariant without betting on
scheduler latency would also close the class. The consumer band-aid
(TimingScale 5) stays in webphone meanwhile.

💘 Generated with Crush

Assisted-by: Crush:GLM-5.3
```

## Verification ledger (for the filing ritual, refreshed 2026-11-21)

- Gate 1 (band-aid does something): webphone TimingScale 3→5 widened the
  window; mechanism verified at v0.6.0 `contract_timing.go` +
  `contract_invariants.go`; 3× `-race` green on the v0.6.0 pin today.
- Gate 2 (not already default): v0.6.0 `contract/contract.go` — `Options`
  has exactly `TimingScale`; no clock anywhere in the contract package's
  non-test source. `WithClock` exists ONLY as a `MemoryStoreOption`.
- Gate 3 (platform): `-race` scheduling stalls are platform-universal.
- Gate 4 (would the fix work): a stepped logical clock removes the
  wall-clock bet entirely; stores keep their own semantics.
- Gate 5 (already exists): `gh issue list --state all` at refresh time =
  #14 (MemoryStore clock, closed — different scope), #13, #12 closed;
  nothing on the contract-suite timing race. The sweep doc's
  "do NOT file" instruction was premised on WithClock retiring the
  band-aid; it does not (MemoryStore-only), so the ask stands.
