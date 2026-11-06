# go-sse — library utilization audit

**Verdict:** thin by architecture — cqrs-htmx's `Broadcaster` wraps this library, and webphone touches the wrapped surface; the test kit pins the SSE wire contract (called "the single best test-integration in the repo" by the 09-24 audit). **Adoption 67/100 (moderate)** — 2 of 3 applicable rows.
**Pin:** v0.7.0 + `go-sse/ssetest` v0.5.0 (direct); `go-sse/sseparse` v0.2.1 (indirect) · **Latest:** v0.7.0 / ssetest v0.5.0 · **Currency:** current (v0.7.0 LANDED 2026-11-05 — the blocked 2026-11-04 bump completed once the cqrs-htmx wave published; see errata) · **Status:** direct dep
**Audited:** 2026-11-04 · **Bump landed:** 2026-11-05 · **Prior reports:** [2026-09-24 fleet stack audit](../../research/2026-09-24_larsartmann-stack-deep-dive.html) (72/100 at v0.6.0). **Movement: 67 at the 2026-11-04 audit (the BLOCKED bump cost the currency points the 09-24 score carried) → still 67 after the landing: the score is rubric-bound (2.0/3 capability rows; Last-Event-ID replay remains a deliberate MISS), and the landing heals the currency axis, not the row math.**

## Status

## Errata — bump attempted and reverted (2026-11-04), then LANDED (2026-11-05)

The v0.6.2→v0.7.0 bump was executed and REVERTED in the 2026-11-04 closeout train. The cascade, verified live: v0.7.0's go.mod requires `github.com/a-h/templ v0.3.1100` (webphone pinned v0.3.1070); the upgrade compiles go-sse itself but breaks two templ-consuming pins in webphone's graph — `templ-components v1.26.0` (fixed by the already-published v1.27.0) and then **`cqrs-htmx/loginpage v4.13.0`** (`templ.ResolveAttributeValue`/`templ.JoinURLErrs` undefined), for which NO templ-v0.3.1100-compatible release existed: the local cqrs-htmx checkout sat ~55 commits past an UNPUBLISHED v4.16.0 tag. Landing v0.7.0 therefore required releasing the cqrs-htmx wave first — a fleet release operation outside a consumer currency bump's scope (Verschlimmbessern guard: a boring bump must not drag an unreleased multi-module wave). (The v0.7.0 changelog drift — the undocumented templ requirement is exactly the consumer-critical fact — was FILED 2026-11-04 as [go-sse#12](https://github.com/LarsArtmann/go-sse/issues/12).)

**RESOLVED 2026-11-05:** the reopen trigger FIRED — `cqrs-htmx/loginpage v4.16.0` published with a go.mod requiring templ ≥ v0.3.1100 (plus `templ-components v1.27.0` + `go-sse v0.7.0`), and the cqrs-htmx root v4.16.0 / usermgmt v4.19.0 wave rode the same templ. The prescribed ONE train landed the same day: templ v0.3.1070→v0.3.1100, go-sse v0.6.2→v0.7.0, templ-components v1.26.0→v1.27.0, loginpage v4.13.0→v4.16.0, cqrs-htmx v4.15.0→v4.16.0, usermgmt v4.18.0→v4.19.0, plus the coupled go-health v0.9.0→v0.10.0 and go-health-dashboard v0.13.0→v0.13.2 (dashboard v0.13.2 requires templ v0.3.1100 + templ-components v1.27.0 itself). The anticipated ssetest `WithLastEventID(sse.EventID)` branding adjustments were MOOT: ssetest's latest is v0.5.0 (already pinned) and webphone never calls `WithLastEventID` (its ssetest usage is NewStreamReader/MustReadNextEvent/RequireEventType/RequireData/RequireDataContains-class asserters only). Full gates green post-train: go build + full `go test -count=1` suite, smoke 49+4+8+9+9+8 (0 failed), tsc, island 286, nix build with the re-pinned vendorHash.

|                |                                                                                                                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Module         | `github.com/larsartmann/go-sse` (+ `ssetest` test module, `sseparse` parser module)                                                                                                 |
| Pinned         | v0.7.0 / ssetest v0.5.0; sseparse v0.2.1 indirect                                                                                                                                   |
| Latest release | v0.7.0 — exported `DefaultSubscriberBuffer`, `EventID` branding at boundaries, ssetest `WithLastEventID(sse.EventID)` (breaking), security docs for parsing untrusted Last-Event-ID |
| Usage          | internal/server/sse.go (prod, `sse.Event` only), sse_test.go + notifier_test.go + hub_bench_test.go (ssetest)                                                                       |

## What the library offers

Root: `Event{Event,ID,Data,Retry}`; generic `Broadcaster[T]` (Subscribe/SubscribeFilter/Broadcast/Unsubscribe/Close/Shutdown(ctx)/Health, buffer + on-drop hooks); `Stream` (Send/SendJSON/heartbeat/OnDisconnect); `LastEventIDFromRequest`; **replay**: `Replay`/`ReplayFiltered` over an `EventStore` interface; writers (`WriteEvent`/`WriteHeartbeat`). `ssetest`: Collect* harness, stream reader + asserters (`RequireEventType/Data/DataContains/EventCount…`), `WithLastEventID`. `sseparse`: standalone WHATWG-conformant parser with conformance corpus.

## How webphone uses it

- **Prod**: `hub.Broadcast(sse.Event{Event: string(name), Data: html})` (internal/server/sse.go:165). The per-extension hub is `cqrshtmx.NewBroadcaster()` — a wrapper AROUND `*sse.Broadcaster[sse.Event]`; webphone owns `ExtensionHubs` (map[extension]*cqrshtmx.Broadcaster) with an idle-TTL reaper (`SubscriberCount()`) and `Shutdown()` → `hub.Close()` (sse.go:116-122, feeds the pre-drain shutdown budget).
- **Serving**: `hub.ServeSSE(w, r)` (sse.go:183) — cqrs-htmx's wrapper, not go-sse's `Stream` directly.
- **Tests**: ssetest stream reader + asserters pin the morph-swap contract: exact event names (`threads`, `thread`, `fax`, `voicemail`, `contacts`), payload shapes, payload-less nudges (internal/server/sse_test.go:135-224).

## Gap analysis

| #   | Capability area                                                            | Status             | Evidence / reason                                                                                                                                                                                                                                                                                                                                                                          |
| --- | -------------------------------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | Event type + broadcast through the cqrs-htmx wrap                          | 🟢 Fully leveraged | sse.go:165; the wrap is the documented composition                                                                                                                                                                                                                                                                                                                                         |
| 2   | ssetest wire-contract pinning                                              | 🟢 Fully leveraged | sse_test.go; the 09-24 audit's standout praise                                                                                                                                                                                                                                                                                                                                             |
| 3   | **Last-Event-ID replay** (`Replay`/`EventStore`, `LastEventIDFromRequest`) | 🔴 Missed          | `rg -in "last-event" internal/` → 0 hits. Deliberate-adjacent: the shell heals a dropped stream by re-fetching nav + open tab (morph-swap design), so missed events are recovered at reconnect. Replay would make reconnects cheap and gap-free; today a busy thread can lose events seen only by the missed push (the next nudge re-fetches, so the window is bounded by event frequency) |
| n/a | Hub policy (filter hooks, on-drop, buffer sizing)                          | ⚪ N/A             | cqrs-htmx's Broadcaster owns hub policy; webphone's reaper uses `SubscriberCount`                                                                                                                                                                                                                                                                                                          |
| n/a | `Stream`/writers direct use                                                | ⚪ N/A             | cqrs-htmx `ServeSSE` owns serving; raw writers would bypass the wrap                                                                                                                                                                                                                                                                                                                       |
| n/a | `sseparse`                                                                 | ⚪ N/A             | parser for SSE _consumers_; webphone only produces                                                                                                                                                                                                                                                                                                                                         |

**Score:** rows 1-2 = 1.0 (2.0) + row 3 = 0 → **2.0 / 3 ≈ 67/100, moderate band.**

## Ruling

The thin surface is a composition decision, not neglect — keep it. Action (a) is DONE (v0.7.0 landed 2026-11-05). Remaining: (b) adopt `Replay` + `LastEventIDFromRequest` only if a live-update gap is ever _observed_ (the re-fetch design is the written alternative; do not pre-build an event store for it).

**Refresh trigger:** user-visible missed-update reports on flaky mobile links; the cqrs-htmx wrap gaining replay passthrough.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "sse.Event" internal/server/sse.go | head -3
rg -in "last-event" internal/ | wc -l                    # expect 0
rg -o "ssetest\.[A-Za-z]+" --type go internal/ | sed 's/.*ssetest\.//' | sort | uniq -c | sort -rn | head -6
git -C /home/lars/projects/go-sse describe --tags --abbrev=0   # expect v0.7.0 (pin v0.7.0, bumped 2026-11-05)
```
