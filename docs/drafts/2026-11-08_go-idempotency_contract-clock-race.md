# Draft — go-idempotency: the contract suite's TTL subtests race the wall clock under -race

**Target repo:** `LarsArtmann/go-idempotency` (own repo) · **Kind:** bug
report body · **Provenance:** unsolicited (AI-found; banner required) ·
**Status:** DRAFT — file only after `@LarsArtmann` review. Routed from
webphone TODO §fleet-libs item 3 (2026-11-07); evidence chain:
webphone `docs/ci-flake-ledger.md` row 42407471289.

**Proposed title:** `contract suite: NoopOnExistingKey fails under -race when a scheduling stall outlasts the scaled shortTTL — the two Record calls race the wall clock`

```markdown
> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `NoopOnExistingKey` holds only while its two `Record` calls execute within the scaled `shortTTL` (20 ms default); a `-race` scheduling stall between them lets the second `Record` legitimately re-claim the expired window, and the assertion reads correct store behavior as "Record extended the TTL". Ask: an injectable clock (or an equivalent logical-time seam) so the contract's TTL subtests become deterministic instead of racing the scheduler.

## Problem

Consuming repo: LarsArtmann/webphone, CI run 42407471289 (2026-11-07),
`go test -race`, subtest
`TestDedupWindowSatisfiesIdempotencyContract/Record/NoopOnExistingKey`.

The invariant (go-idempotency `contract/contract_invariants.go`, `runRecordTests`):

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
`shortTTL` (60 ms at the consumer's TimingScale 3). The second `Record`
then hits the store's `expires_at <= now` guard and legitimately re-claims
a fresh `longTTL` — correct per `CheckAndRecord`/`Record` semantics — and
the subtest fails with "Record extended the TTL".

## Why the existing lever does not fix it

`Options.TimingScale` (contract/contract.go) stretches every wall-clock
timing, but the stall budget scales WITH the timings — a sufficiently slow
`-race` run always can outlast `shortTTL`, whatever the scale. The docs say
it themselves: "TTL-based subtests sleep on wall-clock timings". Webphone
hardened from TimingScale 3 to 5 (stall budget 60 → 100 ms) the same day;
that widens the window, it does not close the race class.

## Evidence of the flake class (not a regression)

- Run 42407471289 red on this subtest; the SAME hooks code was green on
  the four prior runs and on the docs-only successor run 42407606027.
- 3× local `-race` green on the identical tree.
- Verified at go-idempotency source: `contract_timing.go` (wall-clock
  `timings`, `time.Sleep` waits) and `contract_invariants.go:72`
  (NoopOnExistingKey). No clock seam exists anywhere in the non-test
  source; `example/main.go` uses `time.Now()` directly.

## Ask

A deterministic seam for the TTL subtests — for example an
`Options.Clock func() time.Time` that clock-aware stores consult for
`expires_at` (with the suite stepping it instead of sleeping), or a
documented stall-tolerant reformulation of `NoopOnExistingKey` that
asserts the invariant without betting on scheduler latency. The consumer
band-aid (TimingScale 5) is in place in webphone meanwhile.

💘 Generated with Crush
```

## Verification ledger (for the filing ritual)

- Gate 1 (band-aid does something): webphone TimingScale 3→5 widened the
  window; mechanism verified at `contract_timing.go` + `contract_invariants.go`.
- Gate 2 (not already default): grep at go-idempotency HEAD `9a82577` — no
  clock/clock seam in non-test source; only `TimingScale`.
- Gate 3 (platform): `-race` scheduling stalls are platform-universal.
- Gate 4 (would the fix work): a stepped logical clock removes the wall-clock
  bet entirely; stores keep their own semantics.
- Gate 5 (already exists): `gh issue list --repo LarsArtmann/go-idempotency
  --state all` at drafting time = #12 (SQLite TEXT expiry docs) and #13
  (Wise adoption) closed; nothing on the contract timing race.
