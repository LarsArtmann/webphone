# Go policy + error-modernization audit (skills: go-error-modernization + how-to-golang)

> Snapshot-rot header: claims verified as of **07:06 UTC, 2026-11-19** against
> working tree of that moment (erraudit binary: `/run/current-system/sw/bin/erraudit`).
> Tier exit codes and finding counts rot with the next landed file.

## A. Error modernization (go-error-modernization skill)

### A.1 `errors.As` inventory — per-site verdicts

Sweep: `rg 'errors\.As\(' --glob '*.go'` over the whole tree **including vendor/ and the
`internal/arch/testdata/erraudit-rule-probes` fixture module: 0 matches.**
The migration this skill exists for is **already complete** — the repo is
fully on `errors.AsType[E]` (15 files carry it, e.g. `internal/cqrs/dialwindows.go`
`errors.AsType[*DialDeniedError]`, `cmd/webphone/bootreport.go`
`errors.AsType[*bootError]`, `internal/server/actions.go`
`errors.AsType[*gateway.ErrProviderRejected]`). No `//nolint:legacyerrors`
directives exist either (0 matches repo-wide), so the skill's
`nolint-audit --fix` trap is structurally absent.

Verdict per skill step:

| Skill step             | Command                                             | Result                                      |
| ---------------------- | --------------------------------------------------- | ------------------------------------------- |
| 1. fix dry-run         | `erraudit fix ./...` (main module + fixture module) | `No fixes found.` exit 0 — nothing to apply |
| 2. apply fixes         | n/a                                                 | zero auto-fixable transformations           |
| 3. build + test        | island suite + touched Go package (below)           | green                                       |
| 4. `lint --type-aware` | see tier table below                                | Go findings: 0                              |

`errors.Is` posture: ~100 sites, all sentinel-style (`store.ErrNotFound`,
`context.Canceled`, `io.EOF`, `push.ErrUnregistered`, errorfamily classified
values) — spot-checked the seams (`internal/server/actions.go`,
`internal/session/sqlite.go`, `internal/push/push.go`); no value-matching site
would benefit from `AsType`. **Zero migrations warranted**; sentinel discipline
untouched, as mandated.

### A.2 erraudit tier measurements (exact commands, 2026-11-19 ~06:45–07:00 UTC)

| Tier             | Exact command                                                         | Before                                      | After this train                                     |
| ---------------- | --------------------------------------------------------------------- | ------------------------------------------- | ---------------------------------------------------- |
| 1 (enforced bar) | `erraudit lint ./... --type-aware --disable-extensions`               | exit 2, **15 findings — all JS, 0 Go**      | exit 2, **11 findings — all JS, 0 Go**               |
| 2a               | `erraudit lint ./... --enforce-go-error-family`                       | exit 2, 15 (all JS, 0 Go)                   | exit 2, 11 (all JS, 0 Go)                            |
| 2b               | `erraudit lint ./... --enforce-coded-errors`                          | exit 2, 15 (all JS, 0 Go)                   | exit 2, 11 (all JS, 0 Go)                            |
| 3 (owner-only)   | `erraudit lint ./... --max`                                           | not run before                              | exit 2, 11 JS + **2 Go `generic_return` advisories** |
| directives       | `erraudit nolint-audit .`                                             | exit 2: `70 directives: 69 needed, 1 stale` | exit 0: `69 directives: 69 needed, 0 stale`          |
| fixture module   | `erraudit fix ./...` in `internal/arch/testdata/erraudit-rule-probes` | `No fixes found.` exit 0                    | unchanged                                            |

**The Go-side bars HOLD**: tier 1 Go = 0, tier 2 = 0 Go findings both flags.
The nonzero exits come **entirely from erraudit's built-in JS analyzer**
(empty-ish `.catch` handlers + one floating promise in
`internal/web/assets/island/app/**` + `shell.js`), which `--disable-extensions`
does **not** disable. The AGENTS.md "tier 1 … exit 0" wording predates that
analyzer reaching this binary (skill-verified in the tool 2026-11-02; the
2026-10-27 `--max`-clean run predates it). **Instrument drift, not code rot**
for the pre-existing sites; the four gaps below were real doctrine misses.

### A.3 JS findings — all 15 sites adjudicated (owner ruling 2026-10-27: every

error path lands in a visible surface OR carries a written justification at the site)

| Site                                                                                                                      | Verdict                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `island/app/audio.js:94` ringback `resumeAudio().catch(()=>{})`                                                           | **GAP → FIXED**: no at-site justification existed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `island/app/latency.js:30` beacon POST `.catch(()=>{})`                                                                   | **GAP → FIXED** (doctrine lived only in AGENTS.md)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `shell.js:1485` read-receipt `.catch(function(){})`                                                                       | **GAP → FIXED**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `island/app/calls.js:1444` floating `placeCall(...)`                                                                      | **GAP → FIXED** with `void` marker (force path skips the gate consult; `invite()` is try-caught inside `placeCall`, so the promise cannot reject there)                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `island/app/audio.js:122`, `calls.js:992/1462/1656/1685`, `main.js:228/230`, `transcribe.js:243/433`, `shell.js:918/1174` | **KEEP — documented deliberate**: each already carries a written justification directly above the call (gesture-unlock semantics, transfer-verdict surface, page-teardown, teardown-dies-with-page, live-context doctrine, best-effort decoration). The scanner wants the comment _inside_ the handler body; converting 11 doctrine comments is scanner-pleasure churn the skill's fix-to-zero warning exists for. Empirically the in-body shape DOES suppress (the 4 fixed sites cleared: 15 → 11), so an owner may choose to convert them later — posture call, not a defect |

Tier-3 Go advisories (owner-only tier, both ACCEPTED per the decoded
`generic_return` semantics in AGENTS.md — the tool's own remedy text says
"Keep the bare error signature"):

- `internal/server/panels.go:90 threadPanelPrefilled` — error _returns_ are
  pure forwarded dep errors; the only stdlib constructor in the body is a
  `fmt.Errorf` built for a `LogErrorContext` **log line**, not a return.
- `internal/signal/receiver.go:138 readAttachments` — constructs errors solely
  via `errorfamily.WrapInfrastructuref` (code `signal.attachment_read`,
  registered) + `errors.Join`; exactly the repo's typed-error-PRODUCT pattern.

### A.4 Changes applied (error-modernization half)

1. `internal/web/assets/island/app/audio.js` — in-body justification on the ringback resume catch.
2. `internal/web/assets/island/app/latency.js` — in-body justification on the beacon catch.
3. `internal/web/assets/shell.js` — in-body justification on the read-receipt catch.
4. `internal/web/assets/island/app/calls.js` — `void` fire-and-forget marker + rationale comment on the dial-anyway `placeCall`.
5. `internal/server/voicemail_chain_test.go` — removed the 1 stale `//nolint:erraudit` directive (audit: "safe to remove"; no `legacyerrors` directives exist, so the `--fix` trap was structurally absent; removed by hand, not `--fix`).
6. `CHANGELOG.md` — `[Unreleased]` line per the changelog bar.

No Go error paths were created, moved, or retyped → **no new errorfamily
codes, no registry moves**; `TestErrorCodeRegistryIsFresh` run anyway (green,
`docs/error-contract.md` untouched).

### A.5 Deliberately unchanged (error-modernization half)

- The 11 documented-deliberate JS catch sites (A.3 table) — justification
  exists at the site; converting to in-body comments or `//nolint:erraudit`
  is an owner posture call (nolint DECAY + `--max` bypass noted in AGENTS.md).
- The 2 tier-3 `generic_return` advisories — false positives per the decoded
  rule semantics; tier 3 never gates.
- Zero `errors.Is` → `AsType` migrations — every inspected site is sentinel
  value matching (mandate: never migrate).
- AGENTS.md tier wording — updating "exit 0" to reflect the JS analyzer is an
  owner/docs-health edit outside this train's scope; this report is the record.

## B. Go policy audit (how-to-golang skill)

### B.1 go.mod banned-dependency review — **PASS**

Direct requires: **zero banned**. Aligned-with-required: `knadh/koanf/v2`
(config), `samber/do/v2` (DI — v2, not the banned v1), `onsi/ginkgo/v2` +
`gomega` (BDD testing), `a-h/templ`, `sixafter/nanoid`, `go-branded-id`
(domain ids — the skill's branded-ID doctrine, ratified in AGENTS.md
strong-id ruling), first-party fleet libs (cqrs-htmx, go-cqrs-lite,
go-error-family, go-health, httputil, templ-components, go-sse, go-retry,
go-etag, go-flightrecorder, go-paperless, loginpage, usermgmt).

Banned items present **indirect-only** (attributed via
`nix develop -c env -u GOEXPERIMENT go mod why`):

| Indirect                               | Banned entry                                         | Pulled by                                                    | Action                            |
| -------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------ | --------------------------------- |
| `github.com/pkg/errors v0.9.1`         | pkg/errors (unmaintained)                            | `cqrs-htmx/usermgmt/v4` → `ThreeDotsLabs/watermill`          | upstream ask, not actionable here |
| `github.com/fsnotify/fsnotify v1.10.1` | fsnotify                                             | `koanf/providers/file` (koanf is the RECOMMENDED config lib) | upstream's choice                 |
| `go.yaml.in/yaml/v3 v3.0.6`            | yaml.v3 (the maintained yaml.in fork, not gopkg.in)  | upstream chain                                               | note only                         |
| `github.com/go-viper/mapstructure/v2`  | mitchellh/mapstructure (this is the maintained fork) | upstream chain                                               | note only                         |

`math/rand`: 0 usages. No testify, gorm, viper, gin, gorilla, logrus/zerolog,
pkg/errors-direct, satori/uuid, avast/retry-go, swaggo, go-playground/validator,
charm v1 paths, json-iterator/sonic/easyjson — none direct or indirect.

### B.2 Library-choice / pattern compliance

- **HTTP stack**: no gin/huma — the repo rides `net/http` + first-party
  `httputil.NewServer` + cqrs-htmx, a documented architecture ruling
  (AGENTS.md: setup bundle REJECTED; module-graph invariants). The
  required-libraries table is a default, not a mandate over first-party fleet
  rulings. **Compliant-by-ruling.**
- **JSON**: go 1.27.3 with `encoding/json` v1 semantics in CI (clean env);
  the `GOEXPERIMENT=jsonv2` host-env leak is documented in AGENTS.md and the
  dial-gate bullet shows v2 divergence is live knowledge. Switching to
  `encoding/json/v2` would change wire bytes product-wide — **owner call,
  documented, not changed**.
- **Validation**: `sivchari/govalid` used as the TOOL (`govalid-generate`,
  toolchain-floor lesson in AGENTS.md); no banned validator lib present.
- **Rate limiting**: `golang.org/x/time/rate` (stdlib-adjacent, recommended) via httputil limiters.
- **Resilience**: first-party `go-retry` (`ComputeDelay`, chatloop ladder) in
  place of failsafe-go (which rides indirect via cqrs-lite) — first-party
  equivalent, fine.
- **Domain types**: `go-branded-id` direct ✓; hand-rolled domain definitions
  (`domain.Address`, `domain.Phone`, network axis) follow the
  definition-vs-alias doctrine; `go-composable-business-types` absent — this
  product has no Email/Money-class primitives. Fine.
- **Testing**: ginkgo/gomega for state machines, table-driven Go, node:test
  island specs — matches the skill's testing matrix (go-snaps not needed: no
  snapshot lane; smoke/E2E are Python harnesses).
- **Errors as values / no panics in library code**: erraudit Go analyzers at 0
  across every tier; boot panics are the documented C009 boot-time fail-fast
  class (cqrs-lint baselined).

### B.3 Test results (this train)

- Island suite: `node --test --test-force-exit internal/web/assets/island-tests/*.test.mjs` — **364/364 pass**.
- Touched Go package: `nix develop -c env -u GOEXPERIMENT go test -count=1 ./internal/server/ -run 'TestVoicemail'` — **ok**.
- Registry: `nix develop -c env -u GOEXPERIMENT go test -count=1 ./internal/arch/ -run TestErrorCodeRegistryIsFresh` — **ok** (no regen needed).
- `nix fmt` — `formatted 6 files (0 changed)`; the four JS diffs are exactly
  the inserted comments/marker (no reflow).

## C. Follow-ups surfaced (not acted on)

1. Owner posture call: convert the 11 documented JS catch sites to in-body
   comments (or `//nolint:erraudit`) to restore tier-1 exit 0 as written, or
   amend AGENTS.md's tier wording to "0 Go findings" — the JS analyzer is not
   disable-able via `--disable-extensions`.
2. Upstream asks (fleet-level): watermill's `pkg/errors`, koanf/file's
   `fsnotify` — both outside webphone's control.
3. Tier re-measure cadence: AGENTS.md anchors the next at 2026-12-08; this
   report is an early re-measure (the 11-05 overdue flag in the 11-11 status
   row predates it).
