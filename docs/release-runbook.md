# Release runbook (v2.x)

The dance that cut v2.0.0, written down so the next release is a
checklist, not archaeology. The auto-commit daemon commits AND pushes
continuously — work in small, explicitly-committed units, and leave an
EXPLICIT narrative commit at every phase boundary (the daemon's
"chore: auto-commit" sweeps are never the record; before switching
phases, verify `git log -1` carries your message). The daemon has
also re-introduced formatting drift by sweeping foreign files
mid-train (2026-09-22 operator.js in the stack) — before any gate run
on a repo the daemon touched, `git diff` the last daemon commit and
re-run the formatter if it moved styled files.

1. **Fold**: CHANGELOG `Unreleased` → dated section; sync
   FEATURES/TODO_LIST/ROADMAP; explicit commit per doc group.
2. **Bump `webphoneVersion`** in flake.nix (package version AND the
   `/version` ldflags injection — one let-binding; keep it equal to
   the new tag).
3. **Gates**: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`,
   `nix develop -c go test -count=1 ./...`, `nix flake check`,
   `python3 scripts/webphone-smoke.py`. If the train bumped
   cqrs-htmx or go-sse: also re-run `BenchmarkHubFanOut` and append a
   dated table to
   `docs/reviews/2026-09-18_hub-fanout-baseline.md` (MD1 rule; next
   re-run should use `-benchtime=1s -count=5` instead of
   `-benchtime 2000x` for tighter numbers).
4. **Tag + push**: annotated `git tag -a vX.Y.Z -m ...`, push main +
   tag (verify with `git ls-remote` — the daemon may have pushed
   already).
5. **Link check**: `nix run nixpkgs#lychee -- .` — after the push, so
   the new tag link resolves.
6. **Stack bump** (`~/projects/nix-international-telephony`):
   `nix flake lock --update-input webphone`, commit the lock.
7. **Stack gates**: `nix build -L .#telephony-browser` (browser E2E,
   chromium NixOS test — THE island regression gate, deliberately
   outside `checks` for its ~1-2 GB chromium closure; forced rebuild:
   `nix build -L --rebuild .#telephony-browser`, budget 445s);
   `nix build -L .#checks.x86_64-linux.telephony-webphone` (webphone
   VM test); full `nix flake check` with the new lock before
   announcing.
8. **aarch64**: `nix build .#webphone --system aarch64-linux` — plain
   `nix flake check` silently omits aarch64 (it says so in a
   warning). The script then asserts the built ELF's machine bytes
   (`b700` = EM_AARCH64) because `--system` is a restricted setting
   an untrusted client's nix may silently ignore while exiting 0.
   Do NOT reach for `nix flake check --all-systems` as the fix: it is
   evaluation-only for other systems (verified 2026-09-19 — zero
   derivations built, "running 0 flake checks") and gates nothing.
   Checks DO cross-build if you name them:
   `nix build .#checks.aarch64-linux.island-lint` ran green
   cross-arch (oxlint substitutes from cache.nixos.org), so the
   aarch64 gate is explicit cross-builds of the package + the checks
   you care about.
9. **Closing sweep**: any command that boots a server for verification
   ends by PROVING the process dead (`pgrep -f <pattern> || echo
   dead`) — only processes YOU booted; CHANGELOG link edits outside a
   train get a `nix run nixpkgs#lychee -- .`; the post-train tree
   sees `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` once (gitleaks/
   codespell on demand — or via `scripts/buildflow.sh`, which
   promotes them); the daemon's last commits verified pushed
   (`git ls-remote origin main` vs local HEAD).

## pbx-artmann relock ritual (2026-09-21/22)

The stack is consumed by pbx-artmann via a rev pin in its flake.nix
URL (not a path input for telephony). After ANY stack change that
should reach prod:

1. Swap the rev (ALWAYS via `git rev-parse`, never typed) in
   pbx-artmann's `flake.nix` telephony URL.
2. `nix flake update telephony` (old `--update-input` is a no-op
   alias there).
3. `nix run .#lock-drift-probe` — checks narHash + sibling rev + the
   webphone-pin cross-check against the sibling's own lock.
4. Build BOTH toplevels: `nix build
   .#nixosConfigurations.pbx.config.system.build.toplevel` AND the
   `pbx-aarch64` equivalent (NOT `.#pbx-toplevel` — wrong attr; NOT
   `--system` — restricted-setting trap).
5. Verify the webphone unit's ExecStart store path MOVED (the old
   build's path stays in the journal or your notes).
6. The daemon beats manual commits — amend its unpushed commit into
   the narrative message.
