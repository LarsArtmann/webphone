# go-datastar — adjacent library audit (closeout T24)

**Verdict:** transitive-only, deliberately: the ONE consumer is the health dashboard's embedded SDK (`dashboard.WithEmbeddedDatastarSDK` → `/health/datastar.js`), and the import lives in go-health-dashboard, not webphone. The app's own realtime deliberately rides SSE (go-sse) + htmx morph swaps — Datastar's signal-patch model would be a second reactive idiom for zero product gain. templ-components/datastar v1.27.0 rides indirect and unconsumed for the same reason.
**Pin:** v0.6.2 + `go-datastar/static` v0.6.1 (both indirect) · **Latest:** main v0.7.0 (2026-11-07 four-module lockstep, re-verified 2026-11-11) · `static` v0.8.0 (2026-11-17 — named version constants + `VersionIsValid`, static.go +37/−0 additive) · **Currency:** rides the consumers — the behind-delta (main v0.6.2→v0.7.0, static v0.6.1→v0.8.0) flows in when go-health-dashboard/templ-components bump their requires; dashboard v0.13.3 still requires v0.6.2/v0.6.1 and webphone's indirect pins mirror it exactly — not webphone's bump to make · **Status:** indirect-only
**Audited:** 2026-11-04 · 2026-11-22 re-verify · **Prior reports:** none.
**Run all commands from the webphone checkout root** (repo-relative paths).

## Ruling

NO-GO as a direct dependency. The CSP story already prices Datastar honestly: `unsafe-eval` is scoped to `/health` ONLY (Datastar compiles at runtime) and everything else stays strict. **Refresh trigger:** the dashboard library swapping its embedded SDK version (then webphone re-verifies the CSP scope); a product surface that genuinely wants signal-patch reactivity (none does — the island's state machines own their own rendering).

**Superseded in part 2026-11-07:** the signal-patch refresh trigger was examined and does NOT fire — the client-side deep-dive ([docs/research/2026-11-07_datastar-deep-dive.html](../../research/2026-11-07_datastar-deep-dive.html), posture 78/100) affirmed the NO-GO with per-module evidence (36% of the client JS is media-imperative behind Datastar's own fences; tabs already declarative on htmx) and left a sharpened trigger: a live-Reports-shaped surface wanting server-pushed reactive state. The verify block below stays authoritative for the dependency shape.

**2026-11-22 re-verify:** currency unchanged in effect — both pins still indirect and still riding go-health-dashboard v0.13.3's requires (webphone go.mod:109-110 = the dashboard's go.mod:10-11, v0.6.2/v0.6.1 verbatim; that embed is how the module rides in at all). The sharpened trigger re-adjudicated and still cold: every `datastar` mention in webphone Go is the /health lane — `WithEmbeddedDatastarSDK` + the `DatastarJS` route (dashboard.go:62,77), the app-test pins on `/health/datastar.js` (app_test.go:150,177,180), and the css-source comment (assets.go:76); zero Datastar idioms in the client (`data-signals`/`data-store`/`data-computed` → 0 hits anywhere in views/web) and no live-Reports-shaped surface exists to want server-pushed reactive state. The only datastar-adjacent change since the deep-dive is the /health `unsafe-eval` retirement (commit `ae26ba18`, riding the dashboard v0.13.3 data-nonce bridge) — the sanctioned embedded-SDK lane, not adoption. Upstream deltas behind the pins are dashboard-facing, never imported here: main v0.7.0 (lockstep cut — `ScriptTagVersioned` cache-heal, embedded client v1.0.3→v1.0.4, breaking `ScriptHandlerWith(bundle)` second-parameter drop) and static v0.8.0 (additive constants). NO-GO stands.

## How to verify

```console
rg -n "go-datastar" go.mod                           # BOTH modules indirect
rg -n "datastar" --type go internal/ cmd/ | wc -l    # expect 5 (the dashboard route name, its test pins, one css-source comment)
rg -n "WithEmbeddedDatastarSDK" internal/app/dashboard.go   # the one consumer path
rg -n "go-datastar" /home/lars/projects/go-health-dashboard/go.mod   # expect v0.6.2 + static v0.6.1 — the riding lane
```
