# go-daemon — library utilization audit

_Produced by F (subagents 18-22) of the 2026-11-22 full-fleet re-audit's fan-out (runbook step 7); every other line is the audit's own._

**Verdict:** NO-GO — the name misleads: this is NOT daemonization (fork/daemonize is an explicit non-goal upstream); it is mechanism-layer plumbing for unix-socket services, and every slice it offers is already owned by webphone code or its platform. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v0.4.0 (unchanged 2026-11-22; [Unreleased] queues codec work only — see re-verify) · **Status:** not a dependency
**Audited:** 2026-11-04 · 2026-11-22 re-verify · **Prior reports:** none (first audit).

## Status

|             |                                    |
| ----------- | ---------------------------------- |
| Module      | `github.com/larsartmann/go-daemon` |
| In webphone | absent (0 references)              |

## What the library offers

`SocketServer` lifecycle (stale-socket cleanup, 0600 unix sockets, graceful drain, `Lifecycle{Startup,OnListen,OnShutdown}` hooks); **standalone sd_notify** (`NotifyReady`/`NotifyStopping`/`Watchdog`/`ReadNotifyCounters`, no-ops off systemd); unix-socket HTTP client transports; codec negotiation (JSON/CBOR q-value) + `WriteJSON`/`WriteErrorf`; SSE data parsing; flight recorder.

## Webphone overlap analysis (verified)

- **Supervision**: the NixOS unit is `Type=simple` with SIGTERM → `signal.NotifyContext` (cmd/webphone/main.go:185) → `httputil.NewServer` graceful drain (`ShutdownTimeout: 30 * time.Second`, main.go:215) with the hubs-first cascade (main.go:234) → App.Shutdown. No fork, no pidfiles — and the library offers none, so there is no conflict, only irrelevance.
- **Unix-socket server/client + codec negotiation**: webphone is a TCP HTTP app behind Caddy; `rg "sd_notify|NOTIFY_SOCKET" --type go` → zero hits. The one unix-socket surface in the tree is a CLIENT of the external signal-cli daemon (internal/signal/client.go — hand-rolled JSON-RPC 2.0 over `net.Dialer`, not HTTP), which none of the library's HTTP-shaped transport slices cover.
- **SSE parsing**: production side only; go-sse/ssetest owns the test-side parsing ([go-sse report](go-sse.md)).
- **Flight recorder**: webphone rides `go-flightrecorder` directly as a DIRECT dep (internal/app/flightrecorder.go, AGENTS.md flight-recorder seam) — adopting go-daemon's embedded variant would be a second recorder truth.

## The one borderline slice

`NotifyReady`/`Watchdog` under a future `Type=notify` unit would let systemd gate on actual listen-readiness instead of process-start (a real, small operational nicety: probes stop racing the listener). But it requires a unit change (`Type=notify` + `WatchdogSec` policy), the current bootreport + socket activation race is harmless in practice — and the conversion is now declined by doctrine in-tree: package/nixos-module.nix:253-255, "Deliberately Type=simple (the default): readiness is the /startupz latch, not sd_notify — see README 'Readiness vs systemd'. Do not add Type=notify or watchdog restarts." New dependency not justified by that alone.

## Ruling

**Decline.** If `Type=notify` ever happens, re-evaluate ONLY the sd_notify slice (it is self-contained and no-op-safe off systemd). Everything else duplicates owned seams.

**Refresh trigger:** a `Type=notify` conversion in the webphone NixOS module (now codified against — see the doctrine comment); any unix-socket machine surface the library's HTTP-shaped slices actually cover.

## 2026-11-22 re-verify

v0.4.0 is still the tag; [Unreleased] queues two codec additions — `JSONv2Codec` (an `encoding/json/v2` adapter with caller-chosen marshal/unmarshal options) and injectable negotiation codecs (`WithJSONCodec`/`WithCBORCodec`) — intent: let json/v2-committed socket consumers adopt content negotiation without regressing to the v1-backed default. That is still integrator machinery for unix-socket HTTP services webphone is not: the supervision, sd_notify, and SSE slices are untouched, and `go-daemon` still replays 0 references. One trigger owned honestly: "any unix-socket machine surface" has fired in letter — the signal-cli JSON-RPC client (internal/signal/client.go, live since 2026-10-23) dials a unix socket — but not in substance: the library's transports are HTTP-shaped (socket-server lifecycle, HTTP-over-socket clients, HTTP content-type negotiation), and the signal seam is a bespoke JSON-RPC demuxer no slice covers. The `Type=notify` slice is declined by doctrine in-tree (package/nixos-module.nix:253-255 — "Do not add Type=notify or watchdog restarts"; readiness is the /startupz latch per AGENTS.md), so that flip trigger is codified against, not merely unfired. Verdict unchanged.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -rn "go-daemon" --type go internal/ cmd/ go.mod | wc -l      # expect 0
rg -n "signal.NotifyContext" cmd/webphone/main.go               # the shutdown path (main.go:185)
rg -n "Type=simple|Type=notify" package/nixos-module.nix | head -3   # the Type doctrine, in-tree (the webphone unit lives here; the nix checkout's modules sit at modules/*.nix, no nix/ prefix)
```
