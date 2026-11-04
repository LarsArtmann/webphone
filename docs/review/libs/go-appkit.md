# go-appkit — library utilization audit

**Verdict:** AFFIRMED — not a dependency and stays that way. The 2026-10-23 audit's per-module ruling stands unchanged; this report re-verifies absence and re-states the ruling without re-litigating. **Adoption: n/a (non-adoption affirmed).**
**Pin:** none (absent from go.mod; `go-flightrecorder` is the one family module webphone rides DIRECTLY) · **Latest (family tag):** health/v0.3.0 · **Status:** not a dependency
**Audited:** 2026-11-04 · **Authoritative prior:** [2026-10-23 deep-dive](../../research/2026-10-23_go-appkit-deep-dive.html) (score 77/100 over 14 rows — the score grades the *audit target*, the ruling is DO-NOT-ADOPT; the per-module calls live there and in AGENTS.md).

## Re-verification (this tree, 2026-11-04)

| Prior claim | Status |
|---|---|
| `core`/`health`/`realtime`/`security`/`cqrs` re-wrap libraries webphone composes directly (httputil, go-health, go-sse, metaengine) | ✅ unchanged — none of those seams moved to appkit |
| `errorpages` DECLINED — the styled 404 must render inside the session/CSRF shell (`views.NotFoundPanel`), not a standalone minimal page | ✅ unchanged — `h.notFound` contract intact (AGENTS.md § error contract) |
| `otel`/`docs` solve problems this deployment does not have | ✅ unchanged — no OTel surface in the product |
| `systemd` (sd_notify needs `Type=notify` + a module change) and `flightrecorder` unadopted | ✅ flight-recorder need is met DIRECTLY by `go-flightrecorder` v0.2.1 (go.mod:25, internal/server/trace.go); `systemd` still unscheduled (see [go-daemon](go-daemon.md) for the sd_notify borderline slice) |
| Absence from the module graph | ✅ `rg "go-appkit" go.mod` → 0 |

## Ruling

The 2026-10-23 verdict holds: adopting the framework layer would replace a composition root webphone controls with no capability gain. Nothing in this audit's fresh pass (a full sweep of the other 18 family libraries) surfaced a capability that only appkit provides — the closest candidates (sd_notify, unified framework glue) each have a cheaper direct answer already in the tree.

**Refresh trigger:** a `Type=notify` conversion; a second service wanting webphone's exact HTTP-shell recipe (then appkit/`httputil.NewServer` becomes the extraction question, not an adoption); any appkit module that stops re-wrapping and provides a unique capability.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-appkit" go.mod | wc -l                          # expect 0
rg -n "go-flightrecorder" go.mod                          # the direct family ride
rg -n "NotFoundPanel" internal/web/views/error.templ internal/server/*.go | head -3
```
