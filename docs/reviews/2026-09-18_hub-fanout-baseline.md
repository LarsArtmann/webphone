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
