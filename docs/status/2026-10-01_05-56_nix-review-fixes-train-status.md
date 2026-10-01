# Status Report — Nix-review fixes train (2026-10-01 05:56)

Session scope: the follow-up to the earlier nix-review assessment
(`docs/status/2026-10-01_05-26_nix-file-review-session-status.md`). The
owner instruction was **"FIX IT ALL"** — apply every polish item the
assessment surfaced, verify each, and leave the flake better than found.
This report covers only that work and what was noticed during it.

---

## a) FULLY DONE

All committed by the auto-commit daemon (HEAD `2b4807b`; tree clean at
report time). Every item below was verified with a real gate, not by
inspection alone.

| # | Change                                                                                                                                                                                                | Verification                                                                                                                                                                                                  |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Split `nix/module-check.nix` (355 → 185 lines) into `module-check-base.nix` (shared eval + NixOS stand-in), `module-check-csrf.nix`, `module-check-backup.nix`.                                       | `nix build .#checks.x86_64-linux.webphone-module` green; all **15/15** linkFarm entries present; the check derivation was **byte-identical** afterward (Nix reused it — no rebuild), proving no output drift. |
| 2 | Stand-in module gained `freeformType = lib.types.attrsOf lib.types.anything` in `module-check-base.nix`.                                                                                              | Eval green; a new top-level config key no longer breaks the check.                                                                                                                                            |
| 3 | `services.webphone.package` moved from raw `mkOption` to `lib.mkPackageOption pkgs "webphone" { default = null; ... }`.                                                                               | `nix eval` of the option: `hasDefault: false`, `type: package`, description + consumer recipe intact (byte-checked the `${pkgs.system}` escape).                                                              |
| 4 | Added `devShells.ci` (`mkShellNoCC`, go_1_27 + templ + golangci-lint, `GOTOOLCHAIN=local`) in `nix/devshell.nix`.                                                                                     | `nix develop .#ci -c go version` (go1.27.1); appears in `nix flake check`.                                                                                                                                    |
| 5 | `.github/workflows/ci.yml` `go-tests` job switched to `nix develop .#ci -c go test ./...`.                                                                                                            | `nix develop .#ci -c go build ./...` + `go test -count=1 ./internal/domain/...` both green in the new shell.                                                                                                  |
| 6 | Split `package/nixos-module.nix` (464 → 213 lines) under the ~300-line guideline: config only; imports `package/options.nix` (209), `package/caddy-vhost.nix` (39), `package/backup-script.nix` (39). | **Byte-identical output proven** by an eval diff (retention=7: Caddy vhost + both backup scripts + config JSON, 1907 B, `diff` clean); module check reused its derivation again.                              |
| 7 | Added a complete `meta` block to the `webphone-vulnix-triage` package.                                                                                                                                | `nix eval ...webphone-vulnix-triage.meta` shows description/license/mainProgram/maintainers; `checks.webphone-vulnix-triage` green.                                                                           |
| 8 | Documented the whole train: `AGENTS.md` (flake layout bullet + Commands block) and `CHANGELOG.md` (Unreleased → Changed).                                                                             | Read-back; `nix fmt` clean.                                                                                                                                                                                   |

Gate summary for the session: `nix flake check --no-build --keep-going`
→ **"all checks passed!"**; `checks.{format,statix,deadnix,webphone-module,vulnix-triage}`
all build green; repo-wide `statix check` 0 findings; `deadnix` 0 on own
files; `nix fmt` idempotent.

---

## b) PARTIALLY DONE

| # | Item                                           | What's missing / why                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| - | ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Verification breadth                           | I verified the **changed surfaces** but did not run the heavy end-to-end gates (full `nix build .#webphone`, `nix flake check --all-systems`, the KVM backup VM test, the backup drill, `nix run .#vulnix`, `buildflow`). Rationale: the package derivation was untouched (only the triage CLI got `meta`), and the module-split parity was proven at eval time. Honest gap: the **VM test** actually executes the rewritten backup script source — my eval diff proves the _text_ matches, not that the VM still boots it. |
| 2 | aarch64                                        | `nix flake check` omitted aarch64-linux (only warns). The module split / devShell changes are arch-neutral, but cross-system eval was not run.                                                                                                                                                                                                                                                                                                                                                                              |
| 3 | The earlier status report                      | `docs/status/2026-10-01_05-26_nix-file-review-session-status.md` still describes the pre-fix state. I left it as a historical snapshot rather than annotate it with "FIXED in this train".                                                                                                                                                                                                                                                                                                                                  |
| 4 | `go-standard` migration (§ check "Structural") | The nix-review checklist says LarsArtmann Go projects should import `go-nix-helpers`' `go-standard`. This project deliberately does not. I decided **not** to migrate (webphone-specific checks — module-check, vm-tests, island-lint, treefmt scoping — are not modelled by `go-standard`, and the AGENTS flake layout is load-bearing), but I did **not** record that decision in `TODO_LIST.md` / FEATURES.                                                                                                              |
| 5 | `TODO_LIST.md`                                 | No row was added recording the accepted exceptions (hardcoded `webphoneVersion`, the module-check stand-in permissiveness) or the declined `go-standard` migration. It lives only in `AGENTS.md` now.                                                                                                                                                                                                                                                                                                                       |

---

## c) NOT STARTED

| # | Item                                       | Why not started                                                                                                                      |
| - | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | Version/tag drift guard                    | Decided **against** intentionally: a _pure_ eval-time guard is impossible (Nix cannot read git tags without impurity/IFD). See `e)`. |
| 2 | `actionlint` on `.github/workflows/ci.yml` | The YAML edit is small and mirrors the existing `nix develop` invocation, but I did not run a YAML/action linter.                    |
| 3 | Splitting any remaining files              | After the two splits, every own `.nix` file is now ≤ 213 lines — nothing else exceeds the ~300 guideline.                            |
| 4 | `nixos-module.nix` further decomposition   | 213 lines is comfortably inside the guideline; no further split warranted.                                                           |

---

## d) TOTALLY FUCKED UP

**Nothing shipped broken.** One real hazard was hit and fixed in-flight:

- **New files are invisible to Nix until `git add`.** The first rebuild
  after creating `package/backup-script.nix` / `package/caddy-vhost.nix`
  **failed evaluation**: _"Path 'package/backup-script.nix' in the
  repository is not tracked by Git."_ Flake sources are git-scoped, so an
  untracked new file silently 404s the build. Fixed with `git add` (no
  commit). Worth remembering as a standing gotcha.
- **Auto-commit daemon races the tracking requirement.** The daemon had
  already committed the earlier new files (`nix/module-check-*.nix`), so
  they built; the newest two were still untracked when I first built.
  This is benign but makes "build fails on a file you just wrote"
  non-deterministic in this repo.
- **No destructive actions taken.** No `rm`, no `git reset`, no force
  push, no revert of the concurrent session's files.

---

## e) WHAT WE SHOULD IMPROVE

1. **The module check is now _more permissive_ by design.** `freeformType
   = attrsOf anything` means the stand-in accepts any config the module
   writes — great for robustness, but it also **stops catching typo'd
   option paths** inside the module (they'd be silently accepted as
   freeform config instead of erroring). Net judgement: the check is an
   _evaluation smoke test_, not a schema validator, and the webphone
   module's own option tree still enforces its types; so the trade is
   acceptable. If stricter checking is ever wanted, the freeform type
   should be paired with an eval of `options.services.webphone` presence.
2. **`mkPackageOption` with `default = null` is the correct idiom but
   subtle.** Future maintainers may "helpfully" drop `default = null`
   and accidentally make the package optional (it would then default to
   `pkgs.webphone`, which does not exist → confusing eval error in the
   consumer). The in-code comment guards this; a one-line note in the
   module docs would help.
3. **Embedded programs (`caddy-vhost.nix`, `backup-script.nix`) are the
   riskiest text in the module and are only covered by substring
   assertions.** The module check greps fragments (`handle /events`,
   `flush_interval -1`, `snapshots`, …). A **full-text golden** of the
   generated vhost + backup script (like `TestProviderFormPartHeaderBlockGolden`
   did for the gateway) would catch reordering/whitespace drift the
   substring checks miss. My session accidentally produced exactly such
   a baseline (`/tmp/wp-baseline.json`) — it should become a committed
   fixture instead of a throwaway.
4. **CI-shell vs interactive-shell divergence.** `devShells.ci` and
   `devShells.default` now duplicate the `GOTOOLCHAIN=local` env and the
   Go pin. They can drift. Consider a shared `let goEnv = { GOTOOLCHAIN
   = "local"; };`/`let goTools = [ pkgs.go_1_27 ... ];` binding in
   `devshell.nix`, or `inputsFrom`-style composition.
5. **`webphoneVersion` remains a manual single point of failure.** The
   hardcode is documented and `release.sh`-owned, but nothing _checks_
   that a released tag matches the flake binding. A **release-time**
   (not eval-time) assertion in `release.sh`/CI — compare the parsed
   `webphoneVersion` against `git describe --tags --abbrev=0` — is the
   only place this can be enforced hermetically. Decide whether that
   guard is wanted; it has a false-positive edge (a bump commit before
   the tag).
6. **Documentation home drift.** The flake-layout facts now live in a
   dense `AGENTS.md` bullet; the module-check case split and the
   `package/` split are also in `CHANGELOG.md`. That's two homes for
   overlapping facts. Consider a short `docs/nix-layout.md` and point
   both at it, or keep AGENTS as the single home and let CHANGELOG only
   log _history_.
7. **Concurrent sessions keep colliding in this repo.** This session
   observed another session editing `shell.js`, `i18n.go`,
   `layout.templ`, `layout_templ.go`, `messages_test.go` mid-flight. The
   `nix fmt` sweep is safe (scoped), but a future `nix fmt` / gate run
   could reformat or catch _their_ transient breakage. The AGENTS
   warning exists; a cheap mitigation is to always `git add` + build
   only your own paths.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered roughly by value/effort (Pareto: the first ~8 are the 80%).

1. Commit the eval baseline (`/tmp/wp-baseline.json` shape) as a golden
   fixture driving a new `module-output-golden` check entry.
2. Add a release-time tag↔`webphoneVersion` guard to `release.sh` (and
   document the false-positive edge).
3. Run the KVM `checks.x86_64-linux.webphone-backup` VM test once after
   the module split to prove the rewritten backup script still boots.
4. Run `webphone-backup-drill` once post-split.
5. Run `nix flake check --all-systems` (or at least `nix build .#webphone
   --system aarch64-linux`) and verify by ELF bytes.
6. Run `nix build .#webphone` once end-to-end (paranoia; derivation was
   untouched).
7. Run `buildflow` (the quality gate) on the branch.
8. Add `devShells.ci` composition to remove the default/ci env duplication.
9. Run `actionlint` (or equivalent) over `.github/workflows/ci.yml`.
10. Add the accepted-exception rows to `TODO_LIST.md` (version hardcode,
    check permissiveness).
11. Record the declined `go-standard` migration + rationale in
    `TODO_LIST.md`/`FEATURES.md`.
12. Decide on a `docs/nix-layout.md` single home for the flake structure.
13. Consider `checks.webphone-module` asserting `options.services.webphone`
    exists alongside the freeform type (restore typo-catching).
14. Add a `meta` completeness check across all packages (not just manual).
15. Add `deadnix`/`statix` to the CI `fast-checks` job if not already
    (it builds the check derivations — confirm).
16. Verify `nix flake show` output-parity again after this train (the
    earlier parity proof was pre-split).
17. Add a golden for the `services.webphone.settings` JSON output.
18. Add a golden for the generated systemd unit (`webphone.service`).
19. Pin the `mkPackageOption` no-default invariant with an eval assertion
    in module-check.
20. Add the `nix flake check` aarch64 note to AGENTS Commands.
21. Wire a `checks.island-lint` reason-list staleness guard (unrelated but
    noticed while reading checks.nix).
22. Confirm `devShells.ci` closure size vs `default` and record it.
23. Add `#ci` to the AGENTS Commands (done) — mirror in README if it
    documents CI.
24. Introduce a `nix/README.md` or section index for the 8 nix/ files.
25. Standardise the header-comment style across all nix/ modules.
26. Add `nix/module-check-*.nix` header cross-links to their sibling
    cases.
27. Consider a shared `evalModule` helper name vs `moduleSet` for clarity.
28. Evaluate whether `webphone-vulnix-triage`'s `platforms = linux` should
    be narrowed (it's a shell script — arch-neutral already).
29. Add a CI job that fails if `nix fmt -- --check` is dirty (confirm
    `format` check is in fast-checks — it is).
30. Add a CI job for `statix`/`deadnix` standalone (confirm present).
31. Add `nix develop .#ci` to `scripts/` docs if any script re-enters the
    shell.
32. Re-measure the erraudit tiers monthly (per AGENTS) — not this train,
    but tracked.
33. Add a golden test for the Caddy vhost HSTS branch specifically.
34. Add a golden for the `sipUpstream`-set vhost variant.
35. Add a negative test: a module config missing `package` must fail eval.
36. Verify the `freeformType` addition doesn't hide a genuinely unknown
    option in the `services.caddy` subtree (typed stand-in).
37. Consider splitting `nix/checks.nix` (166 lines) if it grows.
38. Document the "new file must be `git add`-ed to be visible to Nix"
    gotcha in AGENTS Commands/lessons.
39. Add a smoke assertion that `/version` still reports the flake version
    after any release-tooling change.
40. Consider a `nix flake check` CI job (currently only fast checks).
41. Add `nix/treefmt.nix` scope note for the new `package/*.nix` files
    (nixfmt auto-covers `.nix` — confirm).
42. Add a `pre-commit` style hook (git-hooks-nix) — the checklist's
    "standard stack" mentions it; currently CI-only.
43. Reconcile the two nix-review status reports into one ledger.
44. Add `CHANGELOG` "Unreleased" lint (heading conventions).
45. Confirm the `devShells.ci` `golangci-lint` is actually needed (CI
    doesn't lint) — drop it if it only bloats the closure.
46. Add a test that the ci shell's `go` matches `go.mod`'s floor.
47. Consider `mkShellNoCC` for `default` too if no C toolchain is needed.
48. Verify `nix flake check` includes `devShells.ci` (it does — no action).
49. Add the module-split rationale to `docs/lessons.md` (why byte-diff
    parity is the acceptance method).
50. Re-run the full gate sweep once the concurrent session's templ/JS
    edits settle, to get a clean whole-repo green at one HEAD.

---

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Is a release-time tag↔version guard wanted at all?** A pure eval-time
   guard is impossible; a `release.sh`/CI git-level guard is possible but
   has a false-positive window (a version-bump commit made _before_ the
   tag commit). Do you want that guard, and if so should it compare to
   `git describe --abbrev=0` (last tag) or only fire when HEAD is
   exactly tagged?
2. **Should the module-check stay maximally permissive (current
   `freeformType` post-fix) or should it also assert
   `options.services.webphone` exists** so a typo'd option path inside the
   module fails the check again? This is a robustness-vs-fault-detection
   tradeoff I can't choose for you.
3. **Do you want the `go-standard` migration (`go-nix-helpers`) actually
   pursued, or is webphone's bespoke flake a permanent deliberate
   deviation?** The nix-review checklist flags it; migrating would
   threaten the webphone-specific checks (`module-check`, `vm-tests`,
   `island-lint`, treefmt scoping) and the documented flake layout, so I
   did not attempt it — tell me if it should be a real backlog item.

---

_Session: read/verify + fix; no destructive actions; concurrent-session
files left untouched. All green gates recorded above; the remaining work
is breadth (heavy gates) and the three open questions._
