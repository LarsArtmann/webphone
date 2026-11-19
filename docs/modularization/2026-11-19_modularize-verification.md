# Modularization verification — 2026-11-19 (6th series entry)

> **Point-in-time snapshot (will rot).** Assessed at HEAD `536a7b20` (2026-11-19,
> clean tree). Facts below were true today and are not maintained — the rulings
> index ([INDEX.md](INDEX.md)) is the entry point, and `TODO_LIST.md`'s carve
> row owns the live refinement. Series: verification posture — a single go.mod
> is RATIFIED (2026-10-14, third verdict, re-affirmed 10-17 and 11-11); this
> entry checks the named reopen trigger and adjudicates the below-module
> refinement. It does not re-litigate the ratified ruling.

## 00 Verdict

**Single go.mod AFFIRMED (6th consecutive entry). The reopen trigger did NOT
fire.** Every boundary claim the ratification rests on was re-measured at HEAD
and holds: zero external Go consumers (GitHub-wide, re-searched today), domain
imports nothing internal, services never import server/web, zero cycles,
solo build green, and the 2026-11-12 auth+platform carves landed exactly as
the refinement plan said they would.

| Probe                         | Result                                | Method                                                                                                                  |
| ----------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| External Go consumers         | **0**                                 | `gh search code` (3 queries) + Sourcegraph public index + local sweep of `~/projects/webphone-android`                  |
| Packages                      | 38, **all** `internal/` or `cmd/`     | `go list ./...` in `nix develop`                                                                                        |
| Real go.mod files             | 1 (root)                              | `find` — the second hit is the `internal/arch/testdata/erraudit-rule-probes` lint-probe fixture, not a buildable module |
| domain internal imports       | **0** (fan-in 28)                     | import map + `TestDomainImportsNothingInternal`                                                                         |
| services importing server/web | **0 violations**                      | import map + `TestServicesNeverImportServerOrWeb`                                                                       |
| Cycles                        | **0** (147 internal production edges) | DFS over the full production import set                                                                                 |
| Solo build (gate 5)           | **GREEN**                             | `GOWORK=off go build ./...` in `nix develop` (go1.27.3)                                                                 |
| `internal/arch` suite         | **GREEN** (13 tests)                  | `go test -count=1 ./internal/arch/`                                                                                     |

## 01 Trigger check — did "a second Go consumer importing internal/domain or internal/gateway" appear?

**No.** Four independent searches, all run this session:

1. `gh search code larsartmann/webphone --match file` — every Go-import hit is
   inside `LarsArtmann/webphone` itself. Outside the repo the only matches are
   `LarsArtmann/webphone-android:SECURITY.md` (an advisory-page **link**) and
   `LarsArtmann/go-health:tools/fleet-repos.txt` (the fleet list — repo name +
   `go.mod`, no import). Both are the same doc-only hits the 10-14/10-17
   verdicts already discounted.
2. Exact trigger paths — `gh search code 'larsartmann/webphone/internal/domain'`:
   **0 hits outside the repo**. `'.../internal/gateway'`: exactly one outside
   hit, `LarsArtmann/.hermes:workspace/projects/.pdg-sweep/logs/webphone.log` —
   a sweep-tool **log** of `go test` output from running _this_ module's own
   tests (`ok github.com/larsartmann/webphone/internal/gateway 0.398s`). Test
   output is not an import; the `.hermes` repo contains no Go module consuming
   webphone. Not a consumer.
3. Sourcegraph public index (`larsartmann/webphone`, `LarsArtmann/webphone`,
   `.../internal`) — **0 matches** for all variants (the org is evidently
   unindexed there; GitHub search above is the authoritative sweep).
4. `~/projects/webphone-android` — **verified Kotlin/Gradle, zero Go**: no
   `.go` files, no `go.mod`/`go.sum`; `build.gradle.kts`/`settings.gradle.kts`/
   `gradlew` present (per the platform decision). Its only webphone reference
   outside `build/` is a markdown draft. It consumes the server as **wire
   contracts** — JSON fixtures from `docs/contracts/wire/` decoded by a Kotlin
   `WireContractTest` — never Go packages. The sibling is a _client_ of the
   HTTP surface, which trigger 3 (a separately deployed artifact) explicitly
   treats as not-module-boundary material.

Standing alongside the empirical zero: the 2026-11-11 fifth entry's
**structural disarm** still holds — 38/38 packages live under `internal/` or
`cmd/` (re-counted today), and Go refuses cross-module imports of `internal/`
packages at compile time, so no second consumer can exist until packages are
deliberately re-homed out of `internal/`. The trigger is closed by
construction, not by vigilance.

## 02 Package-DAG audit at HEAD

All commands inside `nix develop -c` (go1.27.3). Method: full production
import map via
`go list -f '{{.ImportPath}}|{{join .Imports ";"}}' ./internal/... ./cmd/...`
(test imports excluded), then scripted checks.

- **domain purity**: `internal/domain` imports zero internal packages —
  measured, and pinned by `TestDomainImportsNothingInternal` (green). Its
  internal fan-in is **28**, matching the fresh 2026-11-19 review's Ca=28.
- **services never import server/web**: 0 violations across the 38-package
  map (allowed importers: `app`, `cmd/*`, `server/*`, `web/*`, `arch`);
  pinned by `TestServicesNeverImportServerOrWeb` (green).
- **Acyclic**: **147 internal production edges, 0 cycles** (DFS over the full
  set) — same edge count the 2026-11-19 07:53 architecture review measured
  independently.
- **Leaves** (zero internal prod imports): `arch`, `domain`, `httpclient`,
  `sipcred`, `vcard`, `web/assets`. (The 10-17 leaf list grew `httpclient`
  and dropped `asr`/`crm`/`pbx`/`textagent` — those adapters now share the
  carved `internal/httpclient`; expected drift, not a violation.)
- **Solo build**: `GOWORK=off go build ./...` GREEN — no workspace fiction.
- **Enforcement**: `internal/arch` suite green — 13 tests, now including the
  five structural fences the verdicts cite (`TestDomainImportsNothingInternal`,
  `TestServicesNeverImportServerOrWeb`, `TestMustInvokeStaysInTheCompositionRoot`,
  `TestIslandModulesStayIndependent`, `TestReadmodelsStaysGatesAndViewsOnly`)
  plus `TestPackageDepthBudget` (see §03). Boundary drift reds CI in minutes.

The compile-time-boundary benefit a split would buy remains realized at
package granularity and test-enforced. Nothing in the DAG audit moves the
ruling.

## 03 The live refinement — internal/server carve row

**Status: EXECUTED 2026-11-12 (auth + platform); the planned html carve RULED
OUT with evidence; four surface carves are the declared end state.** The row
is struck through in `TODO_LIST.md` ("EXECUTED 2026-11-12, extended to
platform; html carve RULED OUT with evidence") — its one home.

- **What landed**: `internal/server/auth` (both login modes, the ceremonies,
  the OIDC dance, the library `/login` page and the session spine, behind a
  narrow `auth.Shell` seam — the 2026-10-17 api/hooks pattern), then
  same-day `internal/server/platform` (assets, `/metrics` + recorders, CSP
  intake, flight-recorder seam + TraceRecorder's new home, `/phone-api`
  proxy, directory bridge + machineSurface, `/config.js`, export). All four
  carved packages verified present on disk today: `server/{api,auth,hooks,platform}/`.
- **The 2026-11-12 html ruling (carve OUT)**: `pages.go` holds the `handlers`
  hub itself and panels/actions reference admin/crm/unread/auth-audit in both
  directions — moving them "renames the package rather than splitting it."
  The god-package's honest end state is the **four surface carves
  (api/hooks/auth/platform) + the depth-budget arch test**, with the html hub
  as the measured, named residual: 19 files / **5212 LOC** / 13 exports
  (re-measured today: 5212 exactly), down from 106 files / ~7.1k LOC /
  342 exports. Reopen the html question ONLY on a real architecture change
  (per-tab render packages over a shared kernel), never as a mechanical move.
- **Fresh finding — depth-pin saturation** (2026-11-19 07:53 architecture
  review, F1, P2·High): the depth-budget pins that govern the residuals are
  nearly full — `server` 5212/5250 (38 lines headroom), `app` 1606/1650 (44),
  `server/api` 1841/1900 (59), `cqrs` 5672/6000 (328). The ratchet's designed
  escape ("split the package — the carve precedent") has **no pre-sketched
  next seam** for three of them: the html carve is ruled out for the server
  hub, and no next seam is sketched anywhere for `app` or `server/api`. The
  2026-11-18 line-ownership train already rode the server pin 5200→5250 (the
  legitimate release valve, reason in-commit). The review's prescription:
  **pre-sketch the next splits (effort S) before the budget fires mid-train**,
  and ratify pin-riding as the standing valve for the html hub. Companion P3s:
  `.templ` sources (3136 LOC in views) sit outside the ratchet's count, and
  test mass concentrates at the server root (77 files / 15.5k lines, 37% of
  repo test LOC).

The carve train is governed, measured, and closed as designed; its successor
work (pin sketches) is owned by the 11-19 review's roadmap, not by the
modularization question. No below-module fact reaches the go.mod ruling.

## 04 Ruling

**Single go.mod AFFIRMED — 6th consecutive verdict (09-19, 09-30, 10-14,
10-17, 11-11, today).** The named reopen trigger ("a second Go consumer
importing `internal/domain` or `internal/gateway`") did not fire, empirically
(zero imports GitHub-wide; the android sibling is Kotlin/Gradle wire-contract
client) and structurally (compiler-refused while 38/38 packages stay under
`internal/`). The DAG claims underpinning the ratification all hold at HEAD
`536a7b20`. Next scheduled re-review: quarterly, ~2027-01.

_Generated 2026-11-19 · verification posture (trigger check + DAG audit +
carve adjudication; no re-litigation) · all probes run this session inside
`nix develop` (go1.27.3) except the GitHub/Sourcegraph searches · series
home: [INDEX.md](INDEX.md)_
