# Draft upstream ask: `test-coverage` should support `-coverpkg` (union coverage)

> DRAFT 2026-11-15 (T08/F38) — owner tick to file. Not sent. Target repo:
> BuildFlow. Verified against the installed binary (2026-11-15): the
> `test-coverage` step's command construction carries `-coverprofile`
> but no `-coverpkg` knob, and no `.buildflow.yml` key reaches the
> step's argv (config view exposes mode/concurrency/excludes only).

## Body

**`test-coverage`: expose `-coverpkg` (union coverage)**

The `test-coverage` step runs per-package coverage
(`go test -coverprofile=... ./...`), which answers "how well is each
package tested BY ITS OWN suite". It cannot answer the cross-package
question: "how much of the whole module do ALL tests together reach?"
That number needs `-coverpkg=./...` (or the equivalent per-module
union), because in a single-module repo with heavy
integration-through-HTTP testing, the per-package numbers
structurally undershoot:

- `internal/server/api` measures ~7% by its own tests while its bytes
  are contract-pinned through `internal/server`'s HTTP suites;
- `internal/blob`, `internal/views` render ~0-1% direct while every
  server test exercises them transitively.

A union pass (`go test -coverpkg=./... -coverprofile=... ./...` +
`go tool cover -func`) reports the cross-package truth we currently
measure by hand at every release (docs/reviews/2026-09-20_coverage-baseline.md,
appended per re-run since 2026-09-20).

Ask: a `.buildflow.yml` key on the step (e.g.
`coverpkg: "./..."`) or a `go-coverage-extra-args`-shaped override,
so the fleet's coverage records carry both numbers (per-package AND
union) from the same orchestrated step. Alternative that also works:
emit BOTH profiles in one step (per-package as today + a second
union-profiled invocation) and report both totals.
