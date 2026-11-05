# go-daemon — library utilization audit

**Verdict:** NO-GO — the name misleads: this is NOT daemonization (fork/daemonize is an explicit non-goal upstream); it is mechanism-layer plumbing for unix-socket services, and every slice it offers is already owned by webphone code or its platform. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v0.4.0 (local checkout tag) · **Status:** not a dependency
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

|             |                                    |
| ----------- | ---------------------------------- |
| Module      | `github.com/larsartmann/go-daemon` |
| In webphone | absent (0 references)              |

## What the library offers

`SocketServer` lifecycle (stale-socket cleanup, 0600 unix sockets, graceful drain, `Lifecycle{Startup,OnListen,OnShutdown}` hooks); **standalone sd_notify** (`NotifyReady`/`NotifyStopping`/`Watchdog`/`ReadNotifyCounters`, no-ops off systemd); unix-socket HTTP client transports; codec negotiation (JSON/CBOR q-value) + `WriteJSON`/`WriteErrorf`; SSE data parsing; flight recorder.

## Webphone overlap analysis (verified)

- **Supervision**: the NixOS unit is `Type=simple` with SIGTERM → `signal.NotifyContext` (cmd/webphone/main.go:175) → `httputil.NewServer` graceful drain → App.Shutdown cascade with the 30s SSE-drain budget (internal/app/app.go:614). No fork, no pidfiles — and the library offers none, so there is no conflict, only irrelevance.
- **Unix-socket server/client + codec negotiation**: webphone is a TCP HTTP app behind Caddy; `rg "sd_notify|NOTIFY_SOCKET|daemon" --type go` → zero hits.
- **SSE parsing**: production side only; go-sse/ssetest owns the test-side parsing ([go-sse report](go-sse.md)).
- **Flight recorder**: webphone rides `go-flightrecorder` directly as a DIRECT dep (internal/server/trace.go, AGENTS.md flight-recorder seam) — adopting go-daemon's embedded variant would be a second recorder truth.

## The one borderline slice

`NotifyReady`/`Watchdog` under a future `Type=notify` unit would let systemd gate on actual listen-readiness instead of process-start (a real, small operational nicety: probes stop racing the listener). But it requires a unit change (`Type=notify` + `WatchdogSec` policy) in the consuming stack's module, and the current bootreport + socket activation race is harmless in practice. New dependency not justified by that alone.

## Ruling

**Decline.** If `Type=notify` ever happens, re-evaluate ONLY the sd_notify slice (it is self-contained and no-op-safe off systemd). Everything else duplicates owned seams.

**Refresh trigger:** a `Type=notify` conversion in the nix-international-telephony module; any unix-socket machine surface added to the app.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -rn "go-daemon" --type go internal/ cmd/ go.mod | wc -l      # expect 0
rg -n "signal.NotifyContext" cmd/webphone/main.go               # the shutdown path
rg -n "Type=simple|Type=notify" /home/lars/projects/nix-international-telephony/nix/modules/*.nix 2>/dev/null | head -3
```
