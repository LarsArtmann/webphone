# Draft — go-idempotency: injectable clock

Target: LarsArtmann/go-idempotency (own repo). Genre: enhancement request, body register. Provenance: unsolicited (fleet-libs audit finding, AI-run investigation). File with `gh issue create --body-file` (never the piped form), then verify with `gh issue view`.

---

## Title

MemoryStore: injectable clock — every expiry-path test sleeps real time

## Body

> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `MemoryStore` calls `time.Now()` directly at four sites, so nothing that expires can be tested without a real sleep — the library's own suite does exactly that, eight times. Ask: an injectable `now` function defaulting to `time.Now()`.

## Problem

`MemoryStore` reads the wall clock inline in every expiry decision (v0.3.2, `store.go`):

- `store.go:153` — `Seen` lazily deletes on `time.Now().Before(exp)`
- `store.go:174` — `Record` re-records on `now.Before(exp)`
- `store.go:194` — `CheckAndRecord` claims on `now.Before(exp)`
- `store.go:218` — `sweep` bulk-deletes on `now.After(exp)` (plus the `time.NewTicker` at the top of the same function)

There is no seam: `NewMemoryStore(sweepInterval)` takes no clock, and the field set has no `now`.

## Cost

The suite pays for it in sleep time and in race exposure. Every expiry-path test sleeps a real `ttl + ε`:

| File                           | Sleep        |
| ------------------------------ | ------------ |
| `store_test.go:77`             | 40ms         |
| `store_test.go:120`            | 40ms         |
| `store_lifecycle_test.go:27`   | 60ms         |
| `store_lifecycle_test.go:102`  | 120ms        |
| `property_test.go:145`         | `ttl + 20ms` |
| `store_concurrency_test.go:99` | 5ms          |
| `fuzz_test.go:134`             | 20ms         |
| `bench_test.go:185`            | 30ms         |

The 40ms sleeps guard 20-40ms TTLs: a loaded CI runner that delays the test goroutine past the sleep margin flips "expired" to "not yet expired" and the assertion fails on timing, not on logic. Consumers with hour-scale TTLs (webphone's replay guards use 1h) cannot test expiry at all — they either skip the path or fake the whole `Store` interface, which leaves the real `MemoryStore` expiry behavior unpinned.

## Suggested shape

One field, defaulted in the constructor, keeps every existing call site working:

```go
type MemoryStore struct {
    // ...
    now func() time.Time
}

func NewMemoryStore(sweepInterval time.Duration) *MemoryStore {
    s := &MemoryStore{
        // ...
        now: time.Now,
    }
```

plus an option (or exported field) to pin it in tests: `store.now = func() time.Time { return fake }`. The sweep ticker can stay real — the lazy-delete paths in `Seen`/`Record`/`CheckAndRecord` are the ones tests actually exercise, and they are all clock-driven once `now` is injectable.

## Version

go-idempotency v0.3.2 (webphone pin, `go.mod`; local checkout at `9a82577`).

💘 Generated with Crush

Assisted-by: Crush:GLM-5.3
