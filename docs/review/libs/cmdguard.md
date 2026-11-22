# cmdguard — library utilization audit

**Verdict:** NO-GO — webphone is a single-binary server whose process front is a systemd unit, not a terminal; the CLI surface is four ritual one-shots (`-enroll-passkey`, `catalog`, `contracts`, `export-legacy` — the fourth comes from the 2026-11-14 T14 export-or-confirm design, in-tree by 2026-11-20, and was hand-rolled without a framework pull, confirming the shape) behind a ~25-line hand dispatch, and the boot-failure surface is a DESIGNED artifact (the five-part operator contract) that a framework's error printer would bypass. Adopting the Cobra wrapper would buy typed flags nobody types on, styled output nobody looks at, and a SECOND composition root beside internal/app's samber/do container. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v5.1.0 (2026-11-11: `WithStrictPositionalArgs`/`WithArgsPassthrough`, concurrency-safe Builder; the v5 line opened 2026-11-08 — v4 retired at v4.3.1, frozen on `release/v4`) — UNCHANGED at the 2026-11-22 re-verify; the cmdguard-lint submodule moved to v0.4.0 (2026-11-21) and `[Unreleased]` is docs/website/fleet-tooling work with no bearing on this ruling (below) · **Status:** not a dependency
**Audited:** 2026-11-11 · **2026-11-22 re-verify** (currency held; Unreleased intent summarized; one-shot count re-measured at 4) · **Prior reports:** none (first audit).

## 2026-11-22 re-verify — currency held, Unreleased read, no still means no

- **Currency:** the root tag is still v5.1.0 (2026-11-11); the only newer tag family is `cmdguard-lint` v0.4.0 (2026-11-21) — the lint ruleset submodule webphone does not consume either.
- **The ~76-line `[Unreleased]` is glamour/docs/tooling, not capability.** In intent: (a) truth-drift tripwires — a doc-table drift test pinning README/website validate-verb and 13-tag tables to the live `registerBuiltins()`/`FlagTag` sources, plus a `pre-tag-check.sh` release-readiness battery; (b) website/marketing hardening from the 41-review fleet audit (guide-sync/link/smoke guards, `llms.txt`, the v4 RetirementBanner, README "replace your hand-rolled" conversion table, since-annotations, a CI-generated coverage badge); (c) fleet tooling (`fleet-bump.sh`, `fleet-sweep.sh`); (d) a large TODO_LIST purge closing ~47 struck rows. The ONE runtime-adjacent change is SU48: `describe` now serves pristine ANSI-free Long/Example text via a `cmdguard.io/pristine-help` annotation — machine-contract hygiene for agents driving `describe` output. None of it touches the webphone premise: no new server-side capability, no exit-contract change, no flag-depth feature beyond v5.1.0's. **A justified no stays a no.**
- **Premise re-checked against tree drift:** the dispatch grew to four one-shots (main.go:126-148 — `-enroll-passkey`, `catalog`, `contracts`, `export-legacy`) and three flat `flag.NewFlagSet`s (catalog.go:22, contracts.go:26, exportlegacy.go:47 — table/format/check flags, no nesting, no aliases). The new `export-legacy` is the strongest evidence yet FOR the ruling: a genuinely new operational surface (read-only legacy-table dump with sha256+count header, or `-check` classification against the journal projection with a nonzero exit while unconfirmed legacy-only rows exist) landed as a plain function behind a ~6-line dispatch slot — exportlegacy.go runs ~400 lines, but the bulk is dump/classification logic no CLI framework would own, and the flag surface is one flat FlagSet with errors through the designed operator surface.

## Status

|             |                                      |
| ----------- | ------------------------------------ |
| Module      | `github.com/larsartmann/cmdguard/v5` |
| In webphone | absent (0 references)                |

## What the library offers

A Cobra wrapper that catches missing handlers, duplicate commands, and invalid flags at construction: typed struct-tag flags (no string lookups), per-command flag structs, DI with lifecycle management over a samber/do scope, `DoctorCommand`/health checks, zero-panic error contract with error-printed-once + correct exit codes, styled output (fang + lipgloss), config-file/env middleware, and the cmdguard-lint rules. v5.1.0 adds strict positional-args control and a concurrency-safe Builder.

## Webphone overlap analysis (verified)

- **The dispatch**: `cmd/webphone/main.go:126-148` — four `os.Args` one-shots (the enroll-passkey, catalog, contracts and export-legacy pre-server slots), each a plain function, then server boot. `flag.NewFlagSet` × 3 (catalog.go:22 `catalog -o [-plain]`, contracts.go:26 `contracts -o`, exportlegacy.go:47 `export-legacy -table/-format/-data/-check/-confirmations`) with one bare positional for `-enroll-passkey`. No nesting, no aliases, no help surface to keep true — `rg "cmdguard\|cobra" go.mod cmd/ internal/ -i` → 0.
- **The exit-code contract is OWNED**: boot failures render the five-part operator report (bootreport.go) and exit 1 (designed) / 2 (panic) — pinned by tests, consumed by the smoke script (`--expect-version`, boot-failure scenario counts) and by systemd's restart policy. cmdguard's exit-code and error-printing contract would be a second printer over a designed surface.
- **The DI story is a split-brain risk, not a win**: webphone's doctrine is ONE container in `internal/app` (samber/do v2, `MustInvoke*` composition-root-only, arch-test enforced). cmdguard's scope is a CLI-app composition root — for this binary it would either wrap nothing (dead weight) or fork the container.
- **The audience**: operators read journald + the NixOS module options + `/healthz`; the binary is `Type=simple` behind Caddy. Spinners, styled help, and glamour-rendered markdown have no terminal to render to.
- **The dep tree**: cobra + mousetrap + the charmbracelet/glamour/chroma stack (see cmdguard's go.mod) — a large addition to a deployment binary whose vendorHash and binary size are both watched (the `setup` bundle was rejected at a measured +68% for less).

## The one borderline slice

A future `webphone doctor` (config validation + probe dry-run against the live datadir) is the shape cmdguard's `DoctorCommand` is best at. But the same surface is already half-owned: boot REFUSES bad config before listening (wrong config fails boot, unknown keys WARN via `internal/config/unknown.go`), and `/healthz` answers the live-probe half honestly. A doctor subcommand, if ever wanted, is ~50 lines against the existing constructors.

## Ruling

**Decline.** The hand dispatch is smaller than the framework's quick-start example and every error path lands in the designed operator surface. Revisit only if this repo grows a second, human-driven binary (an interactive admin/ops CLI with real flag depth) — at that point cmdguard v5 is the fleet default and the report re-runs as an adoption audit.

**Refresh trigger:** a second binary in this repo; an interactive operator CLI (`doctor`, import/export shells) with more than ritual flags; the enroll one-shot ever needing flag-shaped options.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "cmdguard|cobra" go.mod cmd/ internal/ -i | wc -l    # expect 0
rg -n "os.Args\[1\]" cmd/webphone/main.go                  # the 4 one-shot dispatch sites (export-legacy joined by 2026-11-20)
rg -n "flag.NewFlagSet" cmd/webphone/*.go | wc -l          # expect 3 (catalog, contracts, export-legacy)
rg -c "bootStepError|staticBootError" cmd/webphone/main.go # expect 12 — the designed exit-contract owner
git -C /home/lars/projects/cmdguard tag --list 'v*' | sort -V | tail -1        # expect v5.1.0 (root family only — `describe --tags` now answers lint/v0.4.0, the newer submodule tag)
git -C /home/lars/projects/cmdguard tag --list 'lint/*' | sort -V | tail -1    # expect v0.4.0 (the lint submodule, 2026-11-21 — unconsumed here)
```
