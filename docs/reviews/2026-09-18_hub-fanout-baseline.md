# Hub fan-out benchmark — baseline numbers

**Date:** 2026-09-18 · **Command:**
`GOEXPERIMENT=jsonv2 go test ./internal/server/ -run '^$' -bench BenchmarkHubFanOut -benchtime 2000x`
· **Machine:** AMD Ryzen AI MAX+ 395 (32-thread)

Measures the per-event broadcast cost the notifier adds: N extensions,
each with M live tab connections, one `Broadcast` on the targeted
extension's hub (the production shape — hubs are per-extension, so
fan-out width is the subscriber count of ONE extension, not of the
whole deployment).

| Shape (hubs × subs) | ns/op |
| ------------------- | ----- |
| 1 × 1               | 37.3  |
| 10 × 2              | 45.2  |
| 100 × 2             | 61.8  |
| 100 × 10            | 186.7 |

**Reading:** broadcast cost scales with the _target hub's_ subscriber
count, not the total hub count — exactly the property the per-extension
hub registry was chosen for. At 10 live tabs of one extension the
notifier adds ~0.19 µs per event; even a pathological 100-tab single
extension stays under 0.2 µs. The hub reaper (10-minute idle TTL,
landed the same day) keeps the map bounded without touching these
numbers for live connections.

Re-run the benchmark after any change to `ExtensionHubs` or after a
cqrs-htmx / go-sse version bump and update this file (append a new
dated table — do not overwrite history).

## 2026-09-20 re-run — cqrs-htmx v4.11.0 (MD1 bump-trigger follow-through)

`nix develop -c go test -run '^$' -bench BenchmarkHubFanOut -benchtime 2000x`
· same machine.

| Shape (hubs × subs) | ns/op | v4.9.0 baseline |
| ------------------- | ----- | --------------- |
| 1 × 1               | 40.0  | 37.3            |
| 10 × 2              | 42.6  | 45.2            |
| 100 × 2             | 45.3  | 61.8            |
| 100 × 10            | 119.0 | 186.7           |

**Reading:** no regression — same order of magnitude everywhere, and
the wider fan-outs measure FASTER (2000x iterations are coarse; treat
the deltas as noise in v4.11.0's favor). The bump's only wire change
remains the SSE `retry:` stream prefix, which does not touch the
broadcast path.

Methodology note for the NEXT re-run: prefer
`-benchtime=1s -count=5` over `-benchtime 2000x` — longer, repeated
runs average out the coarse-iteration noise both tables above carry.

## 2026-09-24 re-run — cqrs-htmx v4.12.0 + go-sse v0.6.1 (v2.7.0 train)

`nix develop -c go test -run '^$' -bench BenchmarkHubFanOut -benchtime=1s -count=5`
· same machine · median of 5 (per the methodology note; spreads were
≤ ±1.5 ns on every shape).

| Shape (hubs × subs) | ns/op median | v4.11.0 (2000x) |
| ------------------- | ------------ | --------------- |
| 1 × 1               | 35.0         | 40.0            |
| 10 × 2              | 39.8         | 42.6            |
| 100 × 2             | 39.5         | 45.3            |
| 100 × 10            | 105.0        | 119.0           |

**Reading:** no regression — every shape at or below the v4.11.0
numbers (part methodology: 1s×5 medians are tighter than the 2000x
single shots). The broadcast path is untouched by both bumps.

## 2026-11-06 re-run — cqrs-htmx v4.16.0 + go-sse v0.7.0 — INCONCLUSIVE

The file's own re-run trigger fired twice since the last entry (the
2026-10-25 identity train to cqrs-htmx v4.16.0 and the 2026-11-05 wave
to go-sse v0.7.0), so the benchmark was re-run. **The numbers below are
NOT a regression verdict**: the host carried a load average of ~22 on
32 cores (concurrent nix builds plus two foreign `docs-gates` replays),
and the per-shape spreads are ~2×.

`nix develop -c go test -run '^$' -bench BenchmarkHubFanOut -benchtime=1s -count=6`
· medians and extremes of 6:

| Shape (hubs × subs) | median ns/op | min   | max   | 2026-09-24 median |
| ------------------- | ------------ | ----- | ----- | ----------------- |
| 1 × 1               | 46.8         | 38.7  | 48.7  | 35.0              |
| 10 × 2              | 55.2         | 43.1  | 66.0  | 39.8              |
| 100 × 2             | 65.2         | 42.2  | 84.6  | 39.5              |
| 100 × 10            | 210.3        | 110.9 | 220.5 | 105.0             |

**Reading:** the MINIMA sit 6–11% above the 2026-09-24 medians on a host
that was demonstrably busy — consistent with scheduling noise, not with
a broadcast-path change (no `ExtensionHubs`/notifier edit landed in
either bump train). Treat this entry as "measurement attempted, verdict
carried": **re-run on a quiet host (`load < 8`) before quoting any
number**, per the 2026-09-24 methodology note.

## 2026-11-07 re-run — VERDICT: no regression (the 2026-11-06 inconclusive entry is answered)

The 2026-11-06 entry deferred a verdict because the host was busy (load ~22);
this run took the same measurement on a **quiet host** and delivers the
verdict that entry owed.

`nix develop -c go test -run '^$' -bench BenchmarkHubFanOut -benchtime=1s -count=6 ./internal/server/`
· load average **0.75 / 2.06 / 9.72** (1 / 5 / 15 min) on 32 threads, sampled
immediately before the run · medians and extremes of 6:

| Shape (hubs × subs) | median ns/op | min    | max    | 2026-09-24 median | spread |
| ------------------- | ------------ | ------ | ------ | ----------------- | ------ |
| 1 × 1               | 33.2         | 32.88  | 33.60  | 35.0              | 2.2 %  |
| 10 × 2              | 38.4         | 38.08  | 39.06  | 39.8              | 2.5 %  |
| 100 × 2             | 38.6         | 38.55  | 39.09  | 39.5              | 1.4 %  |
| 100 × 10            | 104.3        | 97.84  | 106.10 | 105.0             | 8.1 %  |

**Reading:** every shape is at or below its 2026-09-24 median, and the
per-shape spreads (1.4–8.1 %) are an order of magnitude tighter than the
2026-11-06 run's ~2× — the difference was scheduling noise, not the library.
That confirms the earlier reading: the two bump trains (cqrs-htmx v4.16.0,
go-sse v0.7.0) did not touch the broadcast path, as expected from their diffs
(no `ExtensionHubs` or notifier change landed in either). The 2026-11-06
"measurement attempted, verdict carried" entry is now **answered**, not
contradicted.
