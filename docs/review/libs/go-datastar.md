# go-datastar — adjacent library audit (closeout T24)

**Verdict:** transitive-only, deliberately: the ONE consumer is the health dashboard's embedded SDK (`dashboard.WithEmbeddedDatastarSDK` → `/health/datastar.js`), and the import lives in go-health-dashboard, not webphone. The app's own realtime deliberately rides SSE (go-sse) + htmx morph swaps — Datastar's signal-patch model would be a second reactive idiom for zero product gain. templ-components/datastar v1.26.0 rides indirect and unconsumed for the same reason.
**Pin:** v0.6.2 + `go-datastar/static` v0.6.1 (both indirect) · **Latest:** v0.6.3 (one patch behind, transitive — not webphone's bump to make) · **Currency:** rides the consumers · **Status:** indirect-only
**Audited:** 2026-11-04 · **Prior reports:** none.
**Run all commands from the webphone checkout root** (repo-relative paths).

## Ruling

NO-GO as a direct dependency. The CSP story already prices Datastar honestly: `unsafe-eval` is scoped to `/health` ONLY (Datastar compiles at runtime) and everything else stays strict. **Refresh trigger:** the dashboard library swapping its embedded SDK version (then webphone re-verifies the CSP scope); a product surface that genuinely wants signal-patch reactivity (none does — the island's state machines own their own rendering).

**Superseded in part 2026-11-07:** the signal-patch refresh trigger was examined and does NOT fire — the client-side deep-dive ([docs/research/2026-11-07_datastar-deep-dive.html](../research/2026-11-07_datastar-deep-dive.html), posture 78/100) affirmed the NO-GO with per-module evidence (36% of the client JS is media-imperative behind Datastar's own fences; tabs already declarative on htmx) and left a sharpened trigger: a live-Reports-shaped surface wanting server-pushed reactive state. The verify block below stays authoritative for the dependency shape.

## How to verify

```console
rg -n "go-datastar" go.mod                           # BOTH modules indirect
rg -n "datastar" --type go internal/ cmd/ | wc -l    # expect 5 (the dashboard route name, its test pins, one css-source comment)
rg -n "WithEmbeddedDatastarSDK" internal/app/dashboard.go   # the one consumer path
```
