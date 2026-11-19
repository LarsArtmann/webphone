# samber/do v2 best-practices audit — internal/app + consumer seams

> Snapshot-rot header: claims verified as of 2026-11-19 07:13 UTC (working tree; the
> auto-commit daemon swept the two code fixes into `bb7850ee` minutes after landing).
> Skill: `samber-do-best-practices` (DO-1 → DO-6). Scope: `internal/app/**` +
> `cmd/webphone` lifecycle + `server.Deps` injection points. DI surface: `samber/do/v2`.
> Test verdicts below re-runnable; the mandated gates were executed AFTER the fixes.

## Verdict

**COMPLIANT — 6/6 DO rules pass.** 30 distinct do-API sites audited; 28 clean on
arrival, 2 fixed (one real DO-2 gap, one stale doc comment). No doctrine-level
violations; no restructure recommended. The composition root is a textbook
implementation of the skill's canonical patterns with repo-specific rulings
(named critical pair, probe-outside-injector, trace-outside-injector) that the
skill would themselves prescribe (self-registration recursion, typed-nil).

## DO-rule compliance summary

| Rule | Verdict | Evidence |
| --- | --- | --- |
| DO-1 `Must*` on runtime paths | PASS | Every `MustInvoke*` sits inside a provider closure, `New`, or a composition-root helper (`wireChatNetworks`, `systemDedup`) — the skill's allowed homes. `App.Start` uses `InvokeNamed`+`wrapf` (error-propagating). Enforced repo-wide by `internal/arch/arch_test.go` `TestMustInvokeStaysInTheCompositionRoot` (flags both the `samber/do` import and `do.MustInvoke` markers outside `internal/app`). |
| DO-2 `do.New()` without `Shutdown` | PASS (after fix) | Exactly one `do.New()` (app.go:194); one `injector.Shutdown()` (app.go:752) via `App.Shutdown`, idempotent (`TestAppShutdownClosesDatabaseAndIsIdempotent`). Pre-fix, the `App.Start`-failure exit path in `cmd/webphone/main.go` returned without `application.Shutdown()` — a partial Start (cqrs workers + probe + dashboard pusher up, then a rejected chat-network config) leaked the cascade. FIXED (see Fixes). |
| DO-3 `Override*` outside tests | PASS | Zero `do.Override*` anywhere in the tree (grep over all `*.go`). Test compositions build a real `New` over temp dirs instead of overriding — the stronger pattern. |
| DO-4 package-level injector | PASS | The injector is a field of `App` (the composition-root wrapper — the skill's `Container` pattern); it never escapes `internal/app`. No service struct stores `do.Injector` (`CqrsService`, `pushRegistryAdapter`, `traceRecorder` all hold resolved deps only). |
| DO-5 `Invoke` inside loops | PASS | No invoke site sits inside a loop; the only loop over container-derived values (`App.Start`'s chat receiver loops) runs already-resolved closures. |
| DO-6 `Shutdown` touching other services/injector | PASS | Every registered Shutdowner is self-contained and idempotent: `store.Database.Shutdown` (own handle), `userauth.Service.Shutdown` (own usermgmt.db + users store), `CqrsService.Shutdown` (GracefulClose + own DLQ handle, latched), `ExtensionHubs.Shutdown` (Broadcaster.Close), `dashboard.Dashboard.Shutdown` (pusher drain). Cross-cutting order lives in `App.Shutdown` — exactly the "lifecycle manager" coordination the skill prescribes. |

Structural smells: none. `server.Deps` is a plain struct (the injector never
reaches handler code); no service-locator; no injector-in-service. Child scopes:
none used, none warranted (no request/session/tenant isolation need — skill says
scopes only when isolation is real). `do.Package`: not used — grouping is via
`registerCqrs`-style helpers; introducing it would move the `server.Deps` literal
the wiring-parity test parses; declined (observation, not a defect).

## Site-by-site table

| # | Site | Rule lens | Verdict | Action |
| --- | --- | --- | --- | --- |
| 1 | app.go:194 `do.New()` | DO-2 | PASS post-fix | — (exit-path fix landed in cmd/webphone) |
| 2 | app.go:203 `ProvideNamed "sqlite"` | named/eager doctrine | PASS | HW-4 suppressed w/ reason; eagerly invoked at :525 |
| 3 | app.go:207 `ProvideNamed "blob-dir"` | named/eager | PASS | same; invoked at :528 |
| 4 | app.go:216 `Provide *pbx.Client` | lazy singleton | PASS | — |
| 5 | app.go:220 `Provide *crm.Client` + :269 named alias | aliasing (DO-3 alternative) | PASS | HW-4 suppressed; alias only when seam on |
| 6 | app.go:224 `Provide *crm.Resolver` w/ `MustInvoke[*crm.Client]` in closure | DO-1 | PASS | provider-closure Must = allowed home |
| 7 | app.go:237-256 eager ASR construction + `Provide(asr.Provider)` + :259 named alias | eager optional seam | PASS | present-but-unusable fails boot = documented seam doctrine |
| 8 | app.go:302 `Provide *push.Fanout` (prebuilt) | value-in-closure | PASS | — |
| 9 | app.go:311 `ProvideNamed "userauth"` w/ MustInvokeNamed cqrs | DO-1, named | PASS | AGENTS-doc'd named registration; 30s boot bound |
| 10 | app.go:327 `ProvideNamed "sso"` | named optional | PASS | lazy discovery (IdP-down ≠ boot fail) = documented |
| 11 | app.go:354 `Provide *session.SQLiteStore` | DO-1 (closure Must) | PASS | — |
| 12 | app.go:358 `Provide *ExtensionHubs` | health wiring | PASS | hw-1 suppression (probe owns health) documented |
| 13 | app.go:361-371 `Provide *Notifier` | DO-1 (closure Musts) | PASS | — |
| 14 | app.go:372/375 gateway Provides | lazy | PASS | — |
| 15 | app.go:378-396 messaging/fax Provides | DO-1 (closure Musts) | PASS | nolint:branching-flow reasons present |
| 16 | app.go:437-518 `Provide http.Handler` (the Deps graph) | service-locator check | PASS | typed-nil SSO discipline (:444-455) is the skill-worthy lesson, pinned by `TestAppSSOOffRendersPlainLoginAndUnregistersRoutes` |
| 17 | app.go:525-556 eager invokes (`InvokeNamed`+`wrapf`) | DO-1 | PASS | error-propagating form — stronger than doctrine requires |
| 18 | app.go:560 `MustInvoke[*ExtensionHubs]` in `New` return | DO-1 | PASS | composition root; cache hit (handler already resolved it); nolint reason |
| 19 | app.go:686-743 `App.Start` | DO-1/DO-5 | PASS | all resolutions hoisted, errors wrapped |
| 20 | app.go:750-769 `App.Shutdown` | DO-2/DO-6, drain order | PASS | Probe → injector → Trace last; `ShutdownError` aggregates w/ `Unwrap() []error` |
| 21 | app.go:68-79 lifecycle guards | interface conformance | PASS | HealthcheckerWithContext ×5, ShutdownerWithError ×3, Shutdowner ×2, adapter-side by design |
| 22 | cqrs.go:320 `ProvideNamed "cqrs"` | named/eager | PASS | invoked at app.go:531; self-contained latched Shutdown (DO-6) |
| 23 | cqrs.go:259-274 `CqrsService.Shutdown` | DO-6 | PASS | own System + own DLQ handle only |
| 24 | chat.go:39 `MustInvoke[*messaging.Service]` in `wireChatNetworks` | DO-1 | PASS | composition-root helper; nolint reason at site |
| 25 | dedup.go:88-95 `systemDedup` MustInvokeNamed + panic | DO-1 | PASS | boot-time miswire panic, documented impossible-by-construction |
| 26 | dashboard.go:86 `ProvideNamedValue "dashboard"` | eager value | PASS | registered only when mounted; drains via cascade |
| 27 | flightrecorder.go traceRecorder (outside injector) | DO-2/DO-6 | PASS post-fix | stale comment claiming container drainage corrected |
| 28 | cmd/webphone main.go run() serve/serve-fail/listen-fail exits | DO-2 | PASS | all call `application.Shutdown()` (bootJoin) |
| 29 | cmd/webphone main.go run() signal exit | drain ordering | PASS | Hubs pre-drain → HTTP drain (30s) → `App.Shutdown` — matches AGENTS sequence exactly |
| 30 | cmd/webphone enrollPasskey one-shot | container bypass | PASS | no `do.New` — builds the same runtime pieces with own defers; no DO-2 exposure |

## Health-check wiring (skill §2/§4.3)

Probe built ONCE outside the injector it reads (self-registration would recurse —
`*health.Probe` conforms to the interface it iterates); threaded to `server.Deps`
and the dashboard; `WithCriticalServices("sqlite","blob-dir")` is the named-check
contract; optional legs (`asr`, `crm`, `cqrs`, `userauth`, `dashboard`) register
NAMED only when their seam is on, so off = byte-identical probes. Per-SURFACE
criticality (vs per-check) is the documented owner ruling, pinned by
`health_crm_test.go` — respected, not re-litigated.

## Findings worth acting on

1. **FIXED — DO-2 exit-path gap** (`cmd/webphone/main.go`): `App.Start` failure
   returned without `application.Shutdown()`. A Start that fails after
   `cqrs.Start`/`Probe.Start` (e.g. matrix `Whoami` Rejection at
   `wireChatNetworks`) left workers, the pusher and both SQLite handles
   undrained. Now `bootJoin(bootErr(...), application.Shutdown())` — the exact
   shape of the listener-failure path 20 lines below.
2. **FIXED — stale lifecycle comment** (`internal/app/flightrecorder.go:55`):
   claimed the container drains the recorder; the doctrine (typed-nil lesson)
   keeps it OUTSIDE the injector with `App.Shutdown` draining it last. Comment
   now states the real contract.
3. **Low, no action (recorded to prevent re-flagging)**: `blob.Store` has
   HealthcheckerWithContext but no Shutdowner — verified correct (directory
   store, no held OS resources). `systemDedup`'s panic vs `App.Start`'s
   wrapf is a style inconsistency inside the composition root, documented
   impossible-by-construction. `do.Package` unused — introducing it would
   break `TestAppWiresEveryServerDep`'s literal-parsing; keep flat.
4. **Doctrine-level, flagged not touched**: HW-4 suppressions (8 sites) and the
   single hw-1 all carry one-line reasons matching AGENTS.md; the
   probe-outside-injector and trace-outside-injector rulings are correct
   applications of the skill's own logic. No changes proposed.

## Fixes applied + test results

- `cmd/webphone/main.go` — Start-failure path drains the container (bootJoin).
- `internal/app/flightrecorder.go` — traceRecorder doc comment corrected.
- `CHANGELOG.md` — `[Unreleased]` line for the code-path change (comment-only
  delta rides silently).
- Gates (post-fix, inside `nix develop -c`, `env -u GOEXPERIMENT`):
  `go test -count=1 ./internal/app/... ./internal/arch/... ./internal/server/ ./cmd/webphone/`
  → **all ok** (app 0.254s, arch 0.579s, server 13.572s, cmd 0.011s). Note:
  `TestAppWiresEveryServerDep` lives in `./internal/arch` (wiring_parity_test.go),
  so the full arch package — including it and `TestMustInvokeStaysInTheCompositionRoot` —
  ran green; the full `./internal/server/` package also ran green.
