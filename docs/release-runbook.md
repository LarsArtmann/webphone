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

## Hard-won release rules (2026-09-23 tail, the v2.6.0 E2E stalls)

- **Load precondition**: `release.sh` step 7 refuses to start the
  stack E2E/VM gates while `/proc/loadavg` (1-minute) is ≥ 8
  (`WEBPHONE_RELEASE_MAX_LOAD` overrides deliberately). Default
  mechanism = the gate, NOT an E2E auto-retry — a retry masks real
  regressions at +10 min per true failure (owner ratification of the
  choice pending, 02:47 §g1). Nothing of YOURS runs during a release
  E2E: the v2.6.0 attempt-2 stall was self-inflicted (own background
  `nix flake check` competing for KVM/CPU).
- **The flake heuristic**: same E2E step stalls TWICE in a row = dig
  into code; DIFFERENT steps inside the post-FS-restart tail = host
  load — wait for sustained quiet and retry once. (Derived from two
  full 7000-line logs; one line each, no more archaeology.) New data
  point 2026-09-23 evening: the E2E×2 runs flaked at DIFFERENT
  post-drill steps (CONTACTS-ROUNDTRIP, INCOMING-SHOWN) at load ~2 —
  low load does NOT clear the different-steps class, the
  post-reconnect recovery window is itself fragile (2 flakes / 6
  runs). So: different steps twice = re-run ONCE more; a THIRD
  failure = dig.
- **Release-attempt logging**: always
  `> /tmp/release-<v>-<n>.log`; NEVER pipe a 30-minute multi-phase
  script through `tail` (it buffers everything until exit — the
  attempt-1/2 monitoring blind spot). Tail the FILE mid-run instead.
- **Mid-release tree asserts**: `release.sh` re-asserts a clean tree
  at tag time — the daemon committing between gates and tag means the
  tag would carry swept content nobody reviewed (f50e825 precedent).
- **Coordination hazard**: v2.6.0 was cut while another session's
  refactor train was mid-flight — the release train folded their
  landed items and rode per-package verification only. Before cutting,
  check for in-flight sessions (dirty files you did not author in ANY
  of the three repos) and either wait or fold explicitly.
- **Chained auto-retries don't fire themselves**: the armed
  wait-for-quiet chained job NEVER ran (log gone, no process). On a
  contended host, plan MANUAL re-execution of the tail steps rather
  than arming background retries.

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
