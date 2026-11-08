# Measure-first latency design: dial-press → first audio, and the ASR round trip (2026-11-08)

**Status:** DESIGN NOTE (no code; L3 of the ai-phone steal-list) · **Owner
ask behind it:** "is the phone fast?" is currently unanswerable — nothing
measures the interval between the human's action and the human's audio, and
the ASR round trip is equally invisible. This note defines the two
intervals, names where each number lands, and estimates the
instrumentation cost. No seam in this note exists yet.

## Interval 1 — dial-press → first remote audio (outbound)

**t0** — the moment the user activates dial (the `click` on the dial
control, or Enter in the dial field). Chosen over `invite()` because the
gap between gesture and `invite()` IS part of the latency the user feels
(if state assembly costs 40 ms, the user pays it).

**t1** — first non-silent remote audio sample rendered to the output
device. Not "call established": established means signaling succeeded;
the user perceives AUDIO. The island already owns the render path (the
shared remote-audio element, `audio.js`), so t1 is observable where the
media already flows — an `AnalyserNode` on the existing WebAudio graph
that watches for the first RMS above a noise floor after `established`,
and stamps once.

**The interval splits into named legs** (the whole point of measuring:
a total without legs cannot be improved):

| Leg             | From → to                                      | Observed at                 |
| --------------- | ---------------------------------------------- | --------------------------- |
| island-overhead | t0 → `invite()` issued                         | calls.js, one timestamp     |
| signaling       | `invite()` → 200 OK / established              | SIP.js session state events |
| media attach    | established → first remote RTP decoded         | `ontrack` → first frame     |
| audio gap       | first decoded sample → first non-silent sample | the RMS watcher             |

Leg 4 is the honest one: ringback vs instant-audio deployments differ
exactly here, and without it leg 3 silently absorbs the PBX's ringback
policy.

## Interval 2 — ASR round trip (a spoken segment → text on screen)

**t0** — the segment's last audio sample is captured (not when capture
STARTS: the user waits on the END of what they said).

**t1** — the transcript text is inserted into the DOM
(`#thread-transcript` / the History section), not when the HTTP response
arrives — render is the perceivable event.

Split legs: capture-flush → POST issued (island), POST → provider first
byte (PerformanceObserver on the fetch), first byte → response complete
(provider), parse → DOM insert (island). The 25 MiB cap means large
segments dominate the provider legs; per-segment size MUST ride the
record or every percentile is meaningless.

## The reporting home — one histogram, server-side

**Decision:** the island measures, the SERVER owns the numbers. The
island sends one fire-and-forget beacon per completed event
(`POST /api/metrics/latency {interval: "dial_to_audio"|"asr_round_trip",
legs_ms: {...}, bytes?: n}`), session-gated, riding the existing
`api.writeJSONBody`/JSONBodyLimit pattern with a tiny body cap (2 KiB).
The server folds it into Prometheus histograms
(`webphone_island_latency_seconds{interval,leg}`), bucketed like the
existing request-duration histogram. Aggregates only — no extension, no
remote label, same leak posture as `webphone_http_request_duration_seconds`.

Why not the island reporting directly: the island has no aggregation
home, a per-tab view dies with the tab, and the operator's one
observability surface is already `/metrics`. Why a beacon and not
Resource Timing: the intervals are app-defined, not resource-defined.

**Failure posture:** the beacon endpoint returns 204 and is
best-effort; a failed beacon is dropped SILENTLY (it is instrumentation;
alerting on the instrument would be the alarm-cries-wolf class). The
endpoint answers 404 when disabled — a config flag
(`metrics.island_beacon`, default OFF = byte-identical product), the
same opt-in shape as `serverTiming.enable`.

## Instrumentation cost estimate

| Piece                                                                    | Cost                                                                   |
| ------------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| `audio.js`: one AnalyserNode + RMS threshold watcher on the shared graph | ~40 LOC, one island spec with a scripted GraphStub                     |
| `calls.js`: t0 stamp + beacon call                                       | ~15 LOC                                                                |
| transcript path: t0/t1 stamps + beacon                                   | ~15 LOC                                                                |
| server: beacon route (session-gated, cap, 204)                           | ~60 LOC + the apidocs entry + registry codes (`metrics.beacon` family) |
| histograms + leg labels                                                  | ~30 LOC in the `httputil.Metrics` consumer                             |
| island type-gate + oxlint + specs                                        | included above                                                         |
| **Total**                                                                | **~160 LOC, one session tranche (≈90 min with gates)**                 |

## What this note deliberately does NOT do

- No code, per the L3 ruling. The note is the design; the train that
  implements it inherits the definitions above verbatim.
- No inbound-call interval yet (ring → answer → audio) — the outbound
  dial path is the owner's stated pain; inbound adds the ring-routing
  fork (island vs Android FGS) and deserves its own note when the
  Android shell lands its first device train.
- No SLO target. Measure first; a target without a distribution is a
  wish. The first week of histograms proposes the target.
