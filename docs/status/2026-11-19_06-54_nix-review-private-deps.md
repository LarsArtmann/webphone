# Nix review + private-Go-deps posture — 2026-11-19

> Snapshot-rot header: claims verified as of **06:54 UTC, 2026-11-19** against the working tree at that time (`webphoneVersion = "2.14.0"`). Nix-store verdicts below were produced in-session; push/CI claims rot fastest.

Skills executed: `nix-review` (Step 1–4 full checklist over every `.nix` file) + `nix-private-go-repos` (private-dep lens: GOPRIVATE coverage, preparedSource/vendorHash recipe validity, CI token story).

## Executive Summary

- **27 `.nix` files reviewed** (flake.nix, flake.lock, `nix/*` ×17, `package/*` ×9); all read in full, none skipped.
- **0 critical issues** (no placeholder hashes, no impure derivations, no IFD, no hardcoded host paths, no secret-in-store paths).
- **0 high issues.**
- **3 medium findings** (file-length guideline ×2, default-uncapped MemoryMax) — all deliberate-or-restructure-adjacent; reported, not patched, per the repo-discipline rule against restructuring the flake layout.
- **2 low findings fixed in this session** (missing `meta` on two packages); re-verified green.
- **Private-dep posture: NO private Go deps.** Every `github.com/larsartmann/*` module in go.mod is a public repo served by the public Go proxy — verified unauthenticated below. `proxyVendor = true` + pinned `vendorHash` is therefore the *correct* hermetic shape; GOPRIVATE / `mkPreparedSource` / per-dep flake inputs / CI `access-tokens` are all rightly absent.

## The private-Go-repos posture (skill lens)

| Question (skill checklist) | Answer | Evidence |
| --- | --- | --- |
| Are the LarsArtmann Go deps private? | **No — all public.** | Unauthenticated `api.github.com/repos/...` → HTTP 200 for `LarsArtmann/cqrs-htmx`, `go-cqrs-lite`, `go-health`, `templ-components`; `proxy.golang.org/.../cqrs-htmx/v4/@latest` → HTTP 200. `proxyVendor = true` (nix/packages.nix:166) could not work otherwise — a private module 410s off the public proxy. |
| GOPRIVATE coverage | **Correctly absent.** No `GOPRIVATE` anywhere in the repo (grep over `.nix`/`.yml`/`.sh`), and none needed: `go mod tidy`/`go get` in the devShell resolve every dep from the public proxy. The skill's case-sensitivity gotcha (mixed-case `larsartmann`/`LarsArtmann`) does not apply because there is nothing to mark private. |
| vendorHash re-pin recipe validity | **Valid and current.** AGENTS.md documents `nix build .#webphone.goModules --rebuild` → read `got:` → apply; that recipe defeats the cached-FOD masking the skill warns about (`--rebuild` forces the FOD re-run, so the declared-hash/output-path collision class cannot bite). Current pin `sha256-yixN5HYyQyMrrM9Biskf9+7RGQsJdQGCVxMoLbwTG2E=` re-verified this session: `nix build .#webphone` green (see Checks run). No `vendor/` directory exists (committed-vendor stance: never had one, per AGENTS 2026-10-17 correction). |
| mkPreparedSource / go-standard needed? | **No.** `mkPreparedSource` exists for *sandbox-unfetchable private repos*; with all deps public, `proxyVendor = true` is strictly simpler and hermetic. `go-standard` was considered and ruled out: this flake is a 68-line slim entry with 4 inputs (not the skill's "5-input, 80+ line boilerplate" trigger), and the repo carries surfaces go-standard does not own (NixOS module + eval-goldens + VM tests + the release-runbook `webphoneVersion` sed contract). |
| CI token story | **None needed, and none configured — correctly.** `.github/workflows/ci.yml` pins `DeterminateSystems/nix-installer-action` by sha with **no** `extra-conf`/`access-tokens`; the `go-tests` job enters `nix develop .#ci` with no GOPRIVATE and no `git insteadOf` rewrite; the art-dupl job `go install`s `github.com/LarsArtmann/art-dupl@v0.7.2` straight off the public proxy. All of this only works because every dependency is public — it is the empirical proof of the posture. |
| The one private-input interaction | **Handled by the right workaround.** `docs-gates` is a `flake = false`, rev-pinned input (flake.nix:24-27) *because the upstream repo's own flake pulls a PRIVATE erraudit input over ssh* which a public-repo CI job cannot fetch. Source-only pinning + narHash in flake.lock keeps the private transitive input out of webphone's lock entirely — exactly the skill's recommended shape for this situation. |

## Critical Issues (Must Fix)

None.

## High (Should Fix)

None.

## Medium (Reported — fixes would restructure or change deployment behavior)

| # | File | Lines | Issue | Why reported, not fixed |
| --- | --- | --- | --- | --- |
| M1 | `nix/vm-tests.nix` | 523 | Exceeds the repo's own "~300 lines" monolith guideline (4 VM tests + the sandbox drill in one file). | Splitting = flake restructure; repo discipline forbids it in this session. Cohesive test family; each test is self-contained. |
| M2 | `package/nixos-module.nix` + `nix/module-check-oauth2.nix` | 374 / 310 | Slightly over the same guideline; the module is already split five ways (options/oauth2/crm/textagent/passkey + vhost + backup script), the residue is the wiring itself. | Same restructure bar; the split that existed to make was already made (comments record the 2026-10-27 splits). |
| M3 | `package/options.nix:129-137` | — | `memoryMax` defaults to `null` ⇒ the long-running service ships **uncapped** unless the consumer sets it (checklist: MemoryMax on all long-running services). | Changing the default silently alters the consuming stack's deployment (nix-international-telephony rides this module). Owner/deployment call, not a drive-by. The option exists and is documented — the fence is one line for any consumer. |

## Low (fixed this session)

| # | File:line | Issue | Fix applied |
| --- | --- | --- | --- |
| L1 | `nix/packages.nix:32-43` | Pinned `templ` package had no `meta` (description/license/mainProgram) — incomplete vs. the webphone + triage packages beside it. | Added full `meta` (MIT verified against `api.github.com/repos/a-h/templ` → spdx MIT). |
| L2 | `nix/packages.nix:51-59` | `docs-gates` package had no `meta`. | Added full `meta` (MIT verified against upstream repo). |

Re-verification after the fix: `nix build .#checks.x86_64-linux.format` green ("formatted 495 files (0 changed)") and `nix build .#checks.x86_64-linux.webphone-module` green — see Checks run.

## Informational (checklist items adjudicated as deliberate deviations)

- **Version not git-derived** — `webphoneVersion = "2.14.0"` is hardcoded in flake.nix:37 by DESIGN: `release.sh` seds that exact binding and the tri-repo release ritual greps it; commit-time enrichment (`self.rev`/`dirtyRev`/`lastModifiedDate`) rides ldflags instead (nix/packages.nix:170-189), preserving byte-reproducibility (no build clock). The skill's `self.rev or self.dirtyRev` pattern is satisfied in spirit via the ldflags lane.
- **Systemd booleans** — `ProtectHome = true` / `RestrictNamespaces = true` / `NoNewPrivileges = true` (package/nixos-module.nix:261-281) are booleans, not strings. NixOS's option layer coerces these to valid systemd values and booleans are the nixpkgs-idiomatic form; the skill's string rule targets raw unit files. No change.
- **`apps.vulnix` is networked at run time** — by declared purpose ("Networked by nature, so it is an app, not a flake check", nix/apps.nix:1-3). The hermeticity invariant is honored where it applies: every *tool* in both apps comes from the store (`pkgs.vulnix`, pinned `nix`, chromium+selenium closure); only the vulnerability-advisory DB and the scanned deployment are live-network reads, which is the app's product, not a hidden `@latest` fetch.
- **No GOWORK=off in devShells** — harmless today (no `go.work` exists; verified), and the `.githooks/commit-msg` guard refuses commits touching `go.work*`. Worth adding the env var if a workspace file ever lands.
- **`flake.nixosModules` at top level** with a `webphone` alias — correct placement (never inside perSystem); both point at `package/nixos-module.nix`.
- **`self` top-level-only gotcha** handled exactly as AGENTS documents: `nix/checks.nix` and `nix/vm-tests.nix` declare `self` at the module top level and close over it lexically into perSystem.
- **No overlays at all** — packages-only flake; no empty-overlay anti-pattern to remove.

## Strengths (what's done well)

- **Source filtering is exemplary**: `lib.fileset.unions [ ../cmd ../go.mod ../go.sum ../internal ]` (nix/packages.nix:150-162) — exactly the Go build inputs, no `.git`, no docs, no `package/`; `builtins.path`-grade reproducibility with better precision.
- **Eval-time module guards are a model**: the store-leak guard (path values and `/nix/store` strings refused — nixos-module.nix:41-56), the inline-secret walker (secret-shaped keys with literal values, one shape over — :65-90), and metrics fencing (non-loopback bind without the vhost warns naming the surfaces — :92-128). Each is pinned by a dedicated module-check case (`module-check-secrets.nix`, `metrics-fencing`, `oauth2-secret-store-guard`), so the guards cannot silently regress.
- **Secrets doctrine is uniform**: no sops/agenix/plaintext mixing — operator files → systemd `LoadCredential` (oauth2/crm/textagent), env-borne secrets → `EnvironmentFile`; every typed seam fails eval on half-config and fails the unit start on a missing source file, with VM-test proof (`webphone-oauth2` removes the source file and asserts the restart loop).
- **Systemd hardening is deep and duplicated deliberately** into the backup oneshot: `ProtectSystem=strict`, `ProtectProc=invisible`, `CapabilityBoundingSet=""`, `SystemCallFilter=@system-service ~@privileged @resources`, `MemoryDenyWriteExecute`, `RestrictAddressFamilies` (INET/INET6/UNIX for the service; **AF_UNIX-only** for the backup — correctly narrower), plus the UMask/StateDirectoryMode private-data pins asserted at eval time AND in the VM.
- **Typed-vs-raw seam discipline**: every typed option renders *plain* so a raw `settings.*` duplicate is an eval **conflict** (one home per fact) instead of a silent winner — pinned per seam (`oauth2-typed-raw-conflict`, `crm-typed-raw-conflict`, `textagent-typed-raw-conflict`).
- **Two full-text goldens** (`module-output.golden`, `module-output-oauth2.golden`) pin the entire generated artifact set (vhost + backup script + settings JSON) against reordering/whitespace drift that substring checks miss.
- **Checks reuse the package** (`checks.webphone = self'.packages.webphone`): the goModules FOD is shared, never rebuilt per check.
- **DevShells**: `mkShellNoCC`, `packages` (not buildInputs), env-var attributes (`GOTOOLCHAIN=local` in both shells), pinned `go_1_27` matching the go.mod floor, the templ compiler pinned to go.mod's library version (nixpkgs' lags and emits incompatible codegen), and a genuinely minimal `ci` shell.

## Checks run (in-session verdicts)

| Command | Result |
| --- | --- |
| `nix build -L .#checks.x86_64-linux.webphone-module` | ✅ green (eval guards, goldens, all csrf/backup/oauth2/crm/textagent/passkey/secrets cases — also the post-fix regression gate) |
| `nix build -L .#checks.x86_64-linux.format` | ✅ green — "formatted 495 files (0 changed)" after the meta edits |
| `nix build -L .#webphone` | see verdict appended below (vendorHash freshness + sandbox build + `doCheck` tests) |
| `nix flake check` | **deliberately NOT run** — KVM-gated VM tests ride it and a full check may be running in another session (repo discipline); the targeted checks above cover the touched surfaces |
| grep sweeps (`rec`, top-level `with`, `inputs.self`, `getEnv`, `<nixpkgs>`, `GOPRIVATE`, `vendor/`, `go.work`) | all clean / absent |

`nix build .#webphone` verdict: **green** — goModules FOD accepted the pinned vendorHash (`sha256-yixN5…TG2E=`), compile + `doCheck` test suite passed in the sandbox. Confirms the re-pin recipe's current pin is fresh as of this snapshot.
