# Dashboard fence decision: no app-side middleware, the proxy is the contract

**Date:** 2026-11-04 · **Task:** closeout plan T19 (`docs/planning/2026-11-04_18-41_SUPERB-fleet-libs-closeout.md`) · **Status:** DECIDED — `WithMiddleware` declined, the written Caddy-fence contract stands, plus one boot WARN.

## The question

The go-health-dashboard library ships `WithMiddleware(mw func(http.Handler) http.Handler)` — a hook to wrap every dashboard route in caller-supplied middleware. Webphone mounts the dashboard at `/health` with rate-limiting and a scoped CSP but **no authentication**; the fence lives entirely in the reverse proxy (the stack module's Caddy/nginx vhost: operator SSO gate in `forwardAuth` mode, loopback-only otherwise). Should the app add an authentication middleware as defense-in-depth?

## Decision

**Decline `WithMiddleware`; keep the proxy fence as the ONE authentication home.** Add a boot WARN on `dashboard.enable` naming the fence's location so an operator who enabled the dashboard cannot miss whose job the fence is.

## Why (the reasons the decision rests on)

1. **The app is proxy-blind.** Every production request arrives from the reverse proxy's loopback address; webphone deliberately has no proven XFF plumbing (`KeyExtractorFromClientIP` is parked until the stack proves XFF sanitization — AGENTS.md). An app-side loopback check would therefore see Caddy, not the client: it cannot enforce "operator only", and a wrong enforcement here either breaks the deployment or lies about securing it.
2. **A header gate is a split-brain fence.** The module already owns the fence-mode decision (`forwardAuth` SSO vs loopback-only). An app-side middleware demanding, say, an SSO-injected header would need its own config axis mirroring that mode — two homes for one fact, exactly the drift class this repo rules against. When the operator flips the module's fence mode, an app-side twin silently disagrees.
3. **The proxy fence is fail-closed and eval-checked.** The stack module fences by default (SSO gate or loopback-only); opening the dashboard to the network is an explicit module-level choice. The residual risk — an operator bypassing the module with a hand-rolled vhost — is a stack-side discipline problem; the WARN closes the awareness gap app-side.
4. **The leaked surface is recon-grade, not secret-grade.** The unfenced dashboard publishes check names, version and trend shapes — no credentials, no per-extension data (the /metrics leak doctrine applies to the dashboard too). Defense-in-depth worth having would be IP-honest; a dishonest gate is worse than a documented edge.

## What landed instead

- Boot WARN (2026-11-04): `App.New` logs `health dashboard mounted WITHOUT app-side authentication; the fence is your reverse proxy (module default: SSO gate in forwardAuth mode, loopback-only otherwise)` when `dashboard.enable` is set — pinned by test.
- README § "Health dashboard" already carries the fence contract ("the dashboard itself carries no authentication"); AGENTS.md § Health dashboard seam rules the mount. This memo is the one home for the WHY.

## Reopen triggers

- The stack proves XFF sanitization and `KeyExtractorFromClientIP` ships: an IP-honest loopback/allowlist middleware becomes buildable — re-decide.
- The library grows an auth-aware option (e.g. `WithBearerToken` validating a configured secret): a symmetric app-side gate with ONE config fact stops being a split-brain and becomes worth taking.
- The dashboard grows a surface that leaks secret-grade data (it does not today).

## Verification

```console
rg -n "WITHOUT app-side authentication" internal/app/      # the WARN + its pin
rg -n "WithMiddleware" internal/app/                        # expect 0 (declined)
```
