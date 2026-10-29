# Caller-perceived latency — spike design (2026-10-29 16:31)

**Status:** SPIKE DESIGN (M17/F72–F73); sample collection (F74) owes one
live session or one stack-E2E instrumentation pass — the numbers are
deliberately NOT invented here. The histogram
(`webphone_http_request_duration_seconds`) covers HTTP only; nothing
measures what the CALLER hears.

## The t0 points (F72)

ai-phone's rule is the calibration: **t0 is the last speech, and the
silence is INSIDE the budget.** For webphone the caller-perceived ladder:

| mark | definition | source seam |
|------|------------|-------------|
| **t_dial** | the operator presses the dial affordance (`data-dial` handler / `wp:dial` submit) | shell.js delegated handler, `calls.js` invite path (`entry.session.invite()`, calls.js:131) |
| **t_invite_ok** | sip.js resolves the INVITE send (0.21.2 resolves `invite()` on SEND — see the call-failure mapping note in error-contract.md) | the `invite()` promise resolution |
| **t_progress** | first SIP 1xx with SDP (ringing/early media) | session state change to `Establishing` + `progress` event |
| **t_media** | first REMOTE audio actually rendered — `pc.getReceivers()[0].track.mute` flipping false, or the ICE `connected` + track unmute pair | `ice.js` state sync + track events |
| **t_asr_rt** | ASR segment round trip: `capturedAt` (the M15 stamp) → `onText` arrival | `transcribe.js` — the lag measurement ALREADY computes `Date.now() - capturedAt`; the lag warn is its threshold form |

The headline number for a placed call is `t_media − t_dial`; for a
received call it is `t_media − t_invite_in`. The ASR round trip is the
second headline (the caption-lag warn fires at 12 s — the spike should
learn what the DISTRIBUTION looks like, not just the tail).

## Capture design (F73) — log-only, no new metric home

Island-side, one `log()` line at call end (the `#log` trail is already
the operator record; console mirrors it):

    call timing: dial→media 3.4s (invite 0.2s, progress 1.1s, media 2.1s)

Rules the capture must keep:

1. **No new telemetry seam.** The AGENTS latency ruling (aggregates only,
   no path/extension labels on `/metrics`) stays untouched; this is the
   operator's own log, fire-and-forget.
2. **Stamps ride existing state.** `entry.startedAt` exists (call start);
   the capture adds `entry.timing = { dial, inviteOk, progress, media }`
   on the same entry object the dock already carries — no new store.
3. **The ASR round trip reuses the M15 lag** — its threshold warn IS the
   productized form; the spike's job is only to see the distribution.
4. Missing marks render as "—" (a declined call has no media mark; the
   line must never lie about what it measured).

## Collection plan (F74 — open, needs one live vehicle)

Two vehicles, cheapest first:

1. **The stack browser E2E** (`telephony-browser`): it already places a
   REAL 1000→1001 call through the VM on every run. Adding the timing
   line to the island turns every E2E run into a sample; the harness log
   greps it. Non-invasive, CI-free.
2. **One real operator call**: boot with the island line, dial once,
   read `#log`. 30 seconds of work, one sample.

Verdict template (fill at collection): dial→media p50/p95 across N
samples; if p95 > 5 s, the follow-up is per-phase attribution (DNS/TLS
on the WSS, ICE crossing, FreeSWITCH answer supervision) BEFORE any
instrumentation grows. If p95 < 2 s, file the number and STOP — the
metric does not owe a permanent home.

*Context: the ai-phone steal-list L3 row (TODO_LIST); the M15 lag warn
shipped the ASR half of this spike already.*
