# cmdguard — library utilization audit

**Verdict:** NO-GO — webphone is a single-binary server whose process front is a systemd unit, not a terminal; the CLI surface is three ritual one-shots (`-enroll-passkey`, `catalog`, `contracts`) behind a 15-line hand dispatch, and the boot-failure surface is a DESIGNED artifact (the five-part operator contract) that a framework's error printer would bypass. Adopting the Cobra wrapper would buy typed flags nobody types on, styled output nobody looks at, and a SECOND composition root beside internal/app's samber/do container. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v5.1.0 (2026-11-11: `WithStrictPositionalArgs`/`WithArgsPassthrough`, concurrency-safe Builder; the v5 line opened 2026-11-08 — v4 retired at v4.3.1, frozen on `release/v4`) · **Status:** not a dependency
**Audited:** 2026-11-11 · **Prior reports:** none (first audit).

## Status

|             |                                      |
| ----------- | ------------------------------------ |
| Module      | `github.com/larsartmann/cmdguard/v5` |
| In webphone | absent (0 references)                |

## What the library offers

A Cobra wrapper that catches missing handlers, duplicate commands, and invalid flags at construction: typed struct-tag flags (no string lookups), per-command flag structs, DI with lifecycle management over a samber/do scope, `DoctorCommand`/health checks, zero-panic error contract with error-printed-once + correct exit codes, styled output (fang + lipgloss), config-file/env middleware, and the cmdguard-lint rules. v5.1.0 adds strict positional-args control and a concurrency-safe Builder.

## Webphone overlap analysis (verified)

- **The dispatch**: `cmd/webphone/main.go:126-140` — three `os.Args` one-shots, each a plain function, then server boot. `flag.NewFlagSet` × 2 (catalog.go:22 `catalog -o [-plain]`, contracts.go:26 `contracts -o`) with one bare positional for `-enroll-passkey`. No nesting, no aliases, no help surface to keep true — `rg "cmdguard\|cobra" go.mod cmd/ internal/ -i` → 0.
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
rg -n "os.Args\[1\]" cmd/webphone/main.go                  # the 3 one-shot dispatch sites
rg -n "flag.NewFlagSet" cmd/webphone/*.go | wc -l          # expect 2 (catalog, contracts)
rg -c "bootStepError|staticBootError" cmd/webphone/main.go # expect 12 — the designed exit-contract owner
git -C /home/lars/projects/cmdguard describe --tags --abbrev=0   # expect v5.1.0
```
