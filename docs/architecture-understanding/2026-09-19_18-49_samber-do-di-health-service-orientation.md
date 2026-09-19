# DI, Health Checks & Service-Orientation Review — webphone

**Date:** 2026-09-19 18:49 · **Scope:** samber/do v2 posture, health checks, service
orientation, composability, resilience, self-health · **Method:** code-read + module-graph
verification + test execution (log at the end). Point-in-time snapshot; re-verify before
treating claims as current truth.

---

## 1. Verdict (the two questions, answered)

**Q1 — "Are we using samber/do v2 in combination with Health Checks superbly?"**

**We are not using samber/do v2 — at all — and that is the correct, deliberate decision.**
Verified facts, not assumptions:

- `github.com/samber/do` has **zero imports** in this repo (grep over all `.go` files).
- It is **absent from the entire module graph**: not in `go.mod` (direct) and not in
  `go list -m all` (transitive). The consumed `cqrs-htmx/v4 v4.9.0` root library does not
  depend on it either (its `go.mod` has no samber entry — verified in the module cache at
  the consumed tag).
- The absence is a **documented architectural decision** (AGENTS.md): the cqrs-htmx
  `setup` bundle — which wires event-sourced usermgmt users through a DI container — was
  **rejected** because this product's identity is the PBX extension + directory password.
  Adoption posture: "middleware + assets only; the setup bundle, CQRS dispatch layer and
  usermgmt stay rejected".

So the literal answer is *N/A by design* — but the **intent** of the question (is DI sound
and are health checks superb?) is fully answerable, and section 3–6 does exactly that.
Short version: **the health-check pattern samber/do prescribes (interface-based
`Healthchecker` + `Shutdowner` lifecycle) is fulfilled here by an equivalent — and for
this codebase better-fitting — mechanism**: named readiness check functions wired to
`GET /healthz` via `cqrshtmx.ReadinessHandler`, plus a static composition root with
correctly ordered self-contained cleanup. Scored verdict in section 7.

**Q2 — "Do we have a PROPER Service Oriented, composable, resilient, and self-health
architecture?"**

**Yes on all four, with one genuine gap (liveness/watchdog) and one small hardening
opportunity (per-check timeout).** No anti-pattern findings. Details and scores below.

---

## 2. What the DI actually is: a static composition root

`cmd/webphone/main.go` (`run()`) is the **single composition root** — manual constructor
injection, layered in strict order:

1. **Infrastructure:** `store.Open` (SQLite), `blob.New` (files), `pbx.NewClient`
   (empty URL ⇒ disabled client returning `ErrDisabled`).
2. **Services:** `store.NewMessages/Faxes/Contacts`, `session.NewStore`,
   `server.NewHubs` + `NewNotifier` (observer wiring), `gateway.NewMessageGateway` /
   `NewFaxGateway` (config-picked seam), `messaging.New`, `fax.New`.
3. **HTTP layer:** one `server.New(server.Deps{…})` receiving everything.

There is no service locator, no global injector, no runtime resolution. Every dependency
is a compile-time-checked constructor parameter. Boot failures are fail-fast with named
errors (`open store: …`, `phone api client: …`).

### DO-1 → DO-6 audit (the samber/do anti-pattern rules, applied to the equivalent constructs)

| Rule  | samber/do smell                          | webphone equivalent audited                                                                                                   | Finding |
| ----- | ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | ------- |
| DO-1  | `MustInvoke` in runtime paths            | No container ⇒ no runtime resolution anywhere. Handlers receive resolved deps via `handlers.deps`.                            | ✅ none |
| DO-2  | `do.New()` without `Shutdown()`          | `db.Close()` deferred in `run()`; `blob.Store` is stateless (no held fds); `http.Server.Shutdown` bounded 10 s; pbx client needs no close. **Shutdown ordering is correct: HTTP drains first, then SQLite closes.** | ✅ none |
| DO-3  | `Override*` outside tests                | No overrides exist; implementation selection happens once, at construction, by config (`GatewayWebhook` vs loopback).         | ✅ none |
| DO-4  | Global package-level injector            | None. Only package-level state is `buildVersion` (ldflags-injected build metadata — idiomatic, not a service).                | ✅ none |
| DO-5  | `Invoke` inside loops                    | N/A — no resolution calls exist.                                                                                              | ✅ none |
| DO-6  | `Shutdown()` reaching into other services | Each cleanup is self-contained (`db.Close`, `httpServer.Shutdown`); no cross-service teardown.                                | ✅ none |

Structural smells: no injector passed deep into business logic; no service holds a
container. **A runtime DI container would add the DO-1 risk class and zero benefit for a
13-package single binary with one static wiring site.** Not adopting samber/do here is
not a gap — it is the better engineering call, and it is what the samber/do skill itself
prescribes for greenfield design ("composition root that returns a cleanup function" —
`run()` is exactly that).

---

## 3. Self-health: the health-check deep dive

### What exists (all verified against code and the consumed tag)

`GET /healthz` = `cqrshtmx.ReadinessHandler` with two **named** checks
(`internal/server/server.go`, route `open.Handle("GET /healthz", readiness)`):

| Check      | Probe                                                        | Catches                                        |
| ---------- | ------------------------------------------------------------ | ---------------------------------------------- |
| `sqlite`   | `deps.DB.Ping`                                               | closed/corrupt handle, lost database           |
| `blob-dir` | `probeBlobDir`: `MkdirAll` → `CreateTemp` → `Close` → `Remove` under `deps.BlobRoot` | full disk, lost/`ro` mount, permission drift — the failure mode that would silently eat attachments |

Verified at the **consumed tag** (cqrs-htmx v4.9.0, module cache — per the
verify-at-tag rule, not master): checks run **in parallel** (goroutine per check,
`sync.WaitGroup`); response is `{"status":"ok","checks":{…}}` with **200**, or
`{"status":"degraded","checks":{"<name>":{"status":"fail","error":"…"}}}` with **503**.
Every failure names its check. GET-open by decision (probers need no session; the body
leaks only check names/errors).

**Deliberate scope decision, judged correct:** the PBX phone API is **not** a readiness
check. The product degrades gracefully without it (loopback mode; login verification
fails closed 401/502 with a boot WARN). Gating readiness on the PBX would take the whole
service down when the PBX blips — the opposite of self-health.

**Deliberate absence of liveness-by-readiness:** only `/healthz` (readiness) exists; no
`/livez`. Under the NixOS module (`Restart = "on-failure"`, `RestartSec = 5`,
`MemoryMax` option) crash-recovery is covered. A **hung** process is not self-detected —
see Finding F2.

### Test evidence (executed during this review, all green)

- `TestHealthzReportsOkWhenBackingResourcesAnswer` — 200 with both check names present.
- Two 503 tests — closed DB and failed blob-dir probe each produce degraded + named check.
- Middleware parity tests hit `/healthz` through the full chain.

### Health-check scorecard vs. the samber/do best-practice shape

| Best practice (samber/do shape)              | webphone realization                                                                        | Status |
| -------------------------------------------- | ------------------------------------------------------------------------------------------- | ------ |
| `Healthchecker` interface on resource holders | Named check **functions** wired at the single wiring site — simpler, no interface ceremony on `*sql.DB` (which we don't own) | ✅ equivalent, better fit |
| `Shutdowner` lifecycle, self-contained       | `defer db.Close()` + bounded `httpServer.Shutdown`; correct drain-then-close ordering        | ✅ |
| Fail-closed readiness naming the failing part | Library `ReadinessHandler`, verified at tag: parallel, named, 503 + error strings            | ✅ |
| Health surfaced without auth friction        | GET-open by decision, body leaks check names only                                            | ✅ |
| Per-check timeout / context                  | `ReadinessCheck func() error` has neither — a hung check would hang the probe (server deliberately has no `WriteTimeout` for SSE) | ⚠️ F1 |
| Liveness distinct from readiness             | Readiness only; hung-process detection is absent end-to-end                                  | ⚠️ F2 |

Residual risk for F1 is small — both probes are local (in-process SQLite ping; local-disk
temp-file write under `dataDir`, which the NixOS module asserts lives under `/var/lib/`).
No network hop can hang them today. It is still worth a cheap guard (F1) because the
check set is the natural place future checks (e.g. PBX with timeout) would land.

---

## 4. Service orientation

**Strong.** The seams are real, not ceremonial:

- **Outbound carrier seam:** `gateway.MessageGateway` / `gateway.FaxGateway` interfaces;
  `NewMessageGateway`/`NewFaxGateway` pick `Webhook` vs `Loopback` by `config.Gateway`
  mode. Loopback makes the whole product run with zero external dependencies — the seam
  is consumed by the stack's `telnyx-webhooks.py` bridge and by every test.
- **Observer seam:** `messaging.ChangeFunc` / notifier callbacks — services notify
  changes without importing the SSE/server layer; direction preserved by construction
  *and* by test (below).
- **Identity seam:** `pbx.Client` is the single point that proves extension credentials
  against the PBX directory (the 2026-09-19 forged-session fix); sessions carry
  `PBXCredentials()` derived from that proof.
- **Enforced direction (executable, not aspirational):** `internal/arch/arch_test.go` —
  `TestDomainImportsNothingInternal` (domain is pure), `TestServicesNeverImportServerOrWeb`
  (services stay usable without HTTP), `TestIslandModulesStayIndependent` (island module
  graph acyclic). **All three executed green during this review.**
- One composition root, zero singletons, no hidden global state (only `buildVersion`).

Split-brain check: none found. There is exactly one user identity (PBX extension), one
session store, one notifier wiring, one store per aggregate. The rejected usermgmt/CQRS
identity would have been the split brain; its rejection holds.

## 5. Composability

**Strong where variability is real; concrete where variability is imaginary — the right
split.** Interfaces exist only at the carrier seam (two implementations, config-chosen)
and at the change-notification seam (function seam). Stores (`*store.Messages` etc.) are
concrete: swapping the storage engine is not a real requirement for a single-binary
SQLite product, and concrete types keep the data model honest. Test composition uses
**real implementations** (in-memory SQLite via `store.Open(":memory:")`, `t.TempDir()`
blob roots) plus a variadic `func(*Deps)` mutator so tests can swap single deps
(e.g. a closed DB for the healthz 503 tests) — composition-root testability without a
container. Library surfaces are pinned by contract tests (CSP hash parity, DOM contract,
CSRF JSON shape), so upstream drift fails the build instead of runtime.

## 6. Resilience

**Strong at process and request level; adequate at deployment level.**

- **Graceful shutdown:** `signal.NotifyContext` (SIGINT/SIGTERM) → `httpServer.Shutdown`
  with 10 s budget (no `WriteTimeout`, deliberately, so SSE survives) → then `db.Close()`.
  In-flight handlers drain before the database closes.
- **Flood control:** keyed per-peer-host limiters with port-stripped keys
  (`remoteHostKey` — port-qualified keys would silently disable limiting behind the
  proxy): login 30/min burst 5, hooks 60/min burst 60, `/events` and `/api/csrf` share
  the hook budget.
- **Fail-closed boundaries:** webhook secret gate (503 without secret, limiter *wraps*
  the gate), session credential verification (401 rejected / 502 unreachable), CSRF
  proxy-trust config logged at boot, webhook 5xx redaction via `SafeDetail`.
- **Chaos-proofing the SSE plane:** hub reaper deletes only idle hubs, never hubs with
  live subscribers, and post-reap broadcasts land on the fresh hub without panic
  (test-pinned: `TestHubReaperDeletesOnlyIdleHubs`).
- **Idempotency:** `hooksIdem` TTL store dedupes replayed `provider_ref`; only successes
  are recorded so failures stay retryable; replays answer `202` inertly.
- **Panic containment:** `cqrshtmx.RecoveryMiddleware` logs full stacks and re-raises
  `http.ErrAbortHandler`.
- **Deployment:** `Restart = "on-failure"` + `RestartSec = 5`, optional `MemoryMax`
  systemd cap, `dataDir` asserted under `/var/lib/`, sessions deliberately ephemeral on
  restart (island re-logs — documented tradeoff).
- **Client-side (island):** bounded reconnect watchdog for the sip.js 0.21.2 hang class.

---

## 7. Scores (1–5 rubric)

| Dimension              | Score | Rationale |
| ---------------------- | ----- | --------- |
| Coupling               | 4.5   | One explicit Deps wiring site; seams at carrier + events; no global state; concrete store types are a deliberate coupling, not an accident |
| Cohesion               | 5     | One concept per package (session, gateway, blob, pbx, messaging, fax); package docs state the responsibility |
| Modularity             | 4.5   | Clean layering with executable import-direction gates; `server` is large but is the composition of the HTTP surface, not a god module |
| Composability          | 4     | Real seams where variability exists; test composition via real impls + Deps mutators; no container overhead |
| Scalability            | 4     | Single-node SQLite by design (correct for the product); SSE fan-out per-extension hubs; horizontal scale would require the storage seam to materialize — YAGNI today |
| Service orientation    | 4.5   | Services usable without HTTP (arch-tested); observer seam; identity seam; config-picked implementations |
| Dependency direction   | 5     | domain → nothing internal; services ↛ server/web — enforced by tests that ran green in this review |
| Self-health            | 4     | Honest named parallel readiness with tests; **gap:** no liveness/watchdog story (F2), no per-check timeout (F1) |
| DI posture             | 4.5   | Static composition root is the *superior* tool at this size; container adoption would be over-engineering (and the rejected `setup` bundle proved the failure mode) |

**Overall: 4.3 / 5 — "proper" yes; "superb" once F1/F2 land.**

---

## 8. Findings register

| ID  | Severity | Finding                                                                                                                                                          | Recommendation |
| --- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| F1  | Low-Med  | `cqrshtmx.ReadinessCheck` is `func() error` — no timeout, no context. A hung check would hang `GET /healthz` indefinitely (no `WriteTimeout` by SSE design). Today both checks are local, so no realistic hang source — but the check set is where future (networked) checks would land. | Wrap each check in a bounded-timeout helper at the wiring site (goroutine + `time.After` returning a named timeout error), or upstream a `ReadinessHandlerTimeout` to cqrs-htmx. ~20 lines + one test. |
| F2  | Medium   | **No liveness story for a hung process.** `/healthz` is readiness-only; systemd restarts on *crash* (`Restart = "on-failure"`) but nothing detects a wedged event loop or stuck disk. A hung webphone serves the nginx vhost forever. | Decide and document: minimum = document `/healthz` as readiness-only in README/module docs; better = systemd watchdog (`WatchdogSec` + sd_notify heartbeat from a goroutine) or stack-side monitoring that acts on `/healthz` failures. Needs an owner call (behavior change at deploy level). |
| F3  | Low      | `sharedContacts(cfg)` in `cmd/webphone/main.go` is a pointless indirection (returns `cfg.Contacts` unchanged).                                                    | Inline `Shared: cfg.Contacts`. Trivial cleanup, zero risk. |
| F4  | Info     | `Deps` mixes high-level services with readiness-only primitives (`DB`, `BlobRoot`) — guarded by a comment ("Nothing else may use them") but nothing enforces it. | Acceptable as-is; if it ever grows, split a `Readiness` sub-struct. No action now. |

No high-severity findings. No DO-1..DO-6 equivalents, no service-locator smell, no split
brains, no global state, no cleanup-order hazards.

---

## 9. Action roadmap

- **P1 · F2 decision:** owner call on liveness/watchdog (document-only vs `WatchdogSec`
  + sd_notify vs stack-side gating). If document-only: README + NixOS module comment,
  effort S.
- **P2 · F1:** bounded-timeout wrapper for readiness checks at the wiring site, with a
  test that a hanging check yields `503` naming `"<check>: timed out"`. Effort S.
  Also consider proposing the timeout upstream to cqrs-htmx (both webphone and the stack
  would inherit it).
- **P3 · F3:** inline `sharedContacts`. Effort S.

F1+F3 are mechanical; F2 needs a decision before code. All three are harvested into
`TODO_LIST.md` (entries "Readiness check timeout guard" and "Liveness/watchdog decision",
plus the inline cleanup folded into this review's follow-ups).

---

## 10. Verification log (what was actually run, 2026-09-19)

1. `grep -r 'github.com/samber/do'` over the repo → **no matches**.
2. `go list -m all | grep -i samber` → **no matches** (not even transitive).
3. cqrs-htmx v4.9.0 `go.mod` in module cache → no samber dependency; `readiness.go` read
   at the consumed tag (parallel named checks, 200/503 shape confirmed).
4. `go test -count=1 -run 'TestHealthz|TestDomainImportsNothingInternal|
   TestServicesNeverImportServerOrWeb|TestIslandModulesStayIndependent'
   ./internal/server ./internal/arch` → **ok, ok** (healthz 200 + both 503 paths, all
   three arch invariants).
5. Manual reads: `cmd/webphone/main.go`, `internal/server/server.go` (wiring + readiness
   + probes), `internal/arch/arch_test.go`, `internal/gateway/gateway.go`,
   `internal/messaging/service.go`, `internal/fax/service.go` (constructor),
   `internal/blob/store.go`, `internal/session/service.go`, `internal/pbx/client.go`,
   `package/nixos-module.nix` (Restart/MemoryMax), healthz/hub-reaper tests.

*N/A-by-design note kept explicit: any future adoption of samber/do (e.g. if the
rejected `setup` bundle's tradeoffs are ever revisited) must re-run this review's
DO-1..DO-6 audit against the container surface.*
