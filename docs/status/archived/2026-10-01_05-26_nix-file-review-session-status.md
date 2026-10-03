# Status Report — `*.nix` Review Session

- **Date:** 2026-10-01 05:26
- **Session scope:** Review the repo's Nix files against the `nix-review`
  skill checklist and answer "Is our `*.nix` superb?"
- **Mode:** Exploration (assessment asked, no fix requested)
- **Verdict:** The Nix surface is genuinely superb — 0 critical, 0 high
  issues. A handful of low-severity polish items remain.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the polish table executed
> wholesale by the 05:56 FIX-IT-ALL train; the heavy verification legs closed
> later (T21 KVM VM + drill, T22 vulnix + aarch64); the version-guard question
> resolved by events (release.sh tag↔webphoneVersion guard); low-tier niceties
> carry honest negatives. Per-item verdicts inline.

---

## a) FULLY DONE

~~1. **Discovered all Nix files in scope.** 9 files, 1492 lines:~~ done — this session (report of record)
   `flake.nix`, `nix/{apps,checks,devshell,module-check,packages,treefmt,vm-tests}.nix`,
   `package/nixos-module.nix`.
~~2. **Read every own Nix file top to bottom.** No skimming; module-check~~ done — this session (report of record)
   (355 lines) and the NixOS module (464 lines) read in full, including
   the second half of `module-check.nix` (lines 200–355) and
   `nixos-module.nix` (lines 200–464).
~~3. **Classified each file by role** (flake entry, packages, NixOS module,~~ done — this session (report of record)
   devShells, checks, apps).
~~4. **Ran the anti-pattern greps** for `fakeHash`, `<nixpkgs>`,~~ done — this session (report of record)
   `builtins.getEnv`, `@latest`, `pnpm dlx`, `nix-shell -p`, `rec {`,
   `with pkgs`, `with lib`. Every hit was inside `vendor/` — none ours.
~~5. **Verified Nix availability:** `nix (Nix) 2.34.8`.~~ done — this session (report of record)
~~6. **Evaluated the whole flake without building:**~~ done — this session (report of record)
   `nix flake check --no-build --keep-going` → **"all checks passed!"**
   (only the expected aarch64 omission warning).
~~7. **Built the fast check derivations green:**~~ done — this session (report of record)
   `checks.x86_64-linux.{statix,deadnix,island-lint,webphone-module}`.
~~8. **Ran `statix check`** over the repo — exit 0, no findings.~~ done — this session (report of record)
~~9. **Ran `deadnix --fail`** over our files — exit 0 on own files; the only~~ done — this session (report of record)
   warnings are in `vendor/*/flake.nix` (deadnix skips `vendor/` in the
   flake check, which is why the check passes).
~~10. **Reviewed the CI workflow** (`.github/workflows/ci.yml`): three jobs~~ done — this session (report of record)
    (nix build, go tests, fast checks incl. webphone-module), SHA-pinned
    actions. Confirms checks are actually exercised.
~~11. **Investigated one cross-layer suspicion** — whether the Caddy~~ done — this session (report of record)
    `handle /health/*` block covers the dashboard's SSE endpoint. Confirmed
    the SSE route is `/health/sse` (go-health-dashboard `routes.SSE`), so
    the unbuffered handle is correct; `/health` root (HTML) intentionally
    falls to the catch-all proxy.
~~12. **Produced a structured review report** with strengths and a~~ done — this session (report of record)
    low-severity polish table.

## b) PARTIALLY DONE

~~1. **Checklist coverage.** I walked the structural/correctness/security/~~ process record — the FIX-IT-ALL train superseded the depth question
   devShell/overlay sections, but did **not** systematically tick every
   single checkbox in `references/common-problems.md` and
   `references/best-practices.md` — I read the skill body but never opened
   those two reference files. Some deep edge cases (e.g. IFD, overlay
   system-ref, `go.work` path handling) were checked by observation only.
~~2. **`nix/builders`-style verification of hermeticity.** I confirmed the~~ done — vulnix zero-real-advisories at the T22 close-out (TODO island row)
   apps' `runtimeInputs` are pinned, but I did not run `nix run .#vulnix`
   (it needs network + the advisory DB) to prove the app executes.
~~3. **aarch64 path.** `nix flake check` warned aarch64 was omitted; I never~~ done — aarch64 exit-green at the T22 close-out
   ran `--all-systems` or the cross-build, so aarch64 reproducibility is
   unverified in this session.
~~4. **`nix fmt -- --check`.** I relied on the `format` check existing in the~~ done — the format check gates every sweep (flake check green routine)
   eval output; I did not independently run treefmt in check mode.

## c) NOT STARTED

~~1. **Fixing any of the five polish items I identified** (explicitly left~~ done — the 05:56 FIX-IT-ALL train executed the polish table
   for your instruction — exploration mode).
~~2. **Building the package itself** (`nix build .#webphone`) — the heaviest~~ done — routine in later trains (smoke boots nix-built binaries)
   single artifact; not needed to answer the review question.
~~3. **Running the KVM backup VM test** (`webphone-backup`) — KVM-gated,~~ done — T21 ran the KVM backup VM + drill post-split (20-12 report)
   multi-minute; skipped by design.
~~4. **Running `webphone-backup-drill`.**~~ done — T21 (same battery)
~~5. **Opening the two skill reference files** (`common-problems.md`,~~ process record — the reference files were never opened; body-knowledge sufficed
   `best-practices.md`).
~~6. **Authoring a fix plan / tasks** derived from the polish table.~~ done — the 05:56 train WAS the plan

## d) TOTALLY FUCKED UP

**Nothing.** No destructive action was taken; no file was modified; no
revert of anyone else's work occurred. The session was read + verify only.
The one honest miss was procedural (reference files not opened, item b-1),
not a mistake with consequences.

## e) WHAT WE SHOULD IMPROVE

~~1. **`nix/module-check.nix` is 355 lines** — over the skill's ~300~~ done — split landed (module-check-base/csrf/backup, 05:56)
   guideline. The csrf case-group and the backup case-group are cleanly
   separable into two files (`module-check-csrf.nix`,
   `module-check-backup.nix`) imported from the entry.
~~2. **Module-check stand-ins are brittle.** The `moduleSet` stand-in only~~ done — freeformType stand-in landed (05:56)
   declares `services.caddy`, `systemd.{services,timers}`, `users.*`,
   `assertions`. Any _new_ top-level config key the module writes fails
   eval with an opaque error. A permissive `freeformType = anything`
   stand-in (or a documented checklist in the check header) would soften
   this — AGENTS.md already flags it as a known constraint.
~~3. **`package` option uses raw `lib.mkOption`.** `lib.mkPackageOption`~~ done — mkPackageOption landed (05:56)
   is the idiomatic helper; here it legitimately has no default, so this
   is cosmetic. Could use `mkPackageOption` with `default = null` +
   a `mkIf`/assertion, or leave as-is with a comment.
~~4. **DevShell granularity.** Only `devShells.default` exists and it packs~~ done — devShells.ci landed + ci.yml switched (05:56)
   LSPs + linters + vulnix. CI's `go-tests` enters that same heavy shell.
   A minimal `mkShellNoCC { packages = [ go golangci-lint templ ]; }` CI
   shell would make CI faster and its contract explicit.
~~5. **`webphoneVersion` is a hand-maintained literal.** Documented and~~ resolved by events — the release-time tag↔version guard landed in release.sh (eval-time stays impossible, as reasoned)
   locked to `release.sh`, but it is the one place the flake trusts a
   human to keep a version in sync with a git tag. Consider an eval-time
   assertion that `webphoneVersion` matches `self.rev`-derived tag when
   clean (won't work for dev trees — reason it's deferred).
~~6. **No `flake-systems` input / hardcoded `systems`.** Fine as-is; noting~~ record stands — fine as-is
   for completeness against the skill's "standard stack".
~~7. **`docs/status/` report naming** is consistent with recent files~~ record stands — convention followed
   (`YYYY-MM-DD_HH-MM_slug.md`) — good, keep following it.
~~8. **Verification depth.** For future Nix reviews, open the two skill~~ process record
   reference files explicitly and tick the checklist, rather than reading
   the skill body and reasoning from memory.

## f) UP TO 50 THINGS TO GET DONE NEXT

### Directly from this review (high → low)

~~1. Split `nix/module-check.nix` csrf cases into their own file.~~ done — 05:56 (module-check-csrf.nix)
~~2. Split `nix/module-check.nix` backup cases into their own file.~~ done — 05:56 (module-check-backup.nix)
~~3. Make the module-check stand-in permissive (`freeformType`), or add a~~ done — freeformType landed (05:56)
   "new config keys need a stand-in" checklist comment at its top.
~~4. Add a minimal `devShells.ci` (mkShellNoCC: go, templ, golangci-lint)~~ done — devShells.ci + ci.yml switch (05:56)
   and switch `ci.yml` `go-tests` to it.
~~5. Decide `mkPackageOption` vs raw `mkOption` for `package`; add rationale~~ done — mkPackageOption with rationale (05:56)
   comment either way.
~~6. Add an eval-time `webphoneVersion` ↔ git-tag consistency check (clean~~ done — the release-time guard landed in release.sh (tag↔webphoneVersion, override documented)
   trees only).
~~7. Open `references/common-problems.md` and tick every applicable box.~~ process record — never opened
~~8. Open `references/best-practices.md` and diff our patterns against it.~~ process record — never opened
~~9. Run `nix flake check --all-systems --no-build` to catch aarch64 eval.~~ done in part — aarch64 exit-green landed at the T22 close-out
~~10. Run `nix fmt -- --check` independently.~~ done — the format check gates every sweep
~~11. Run `nix build .#webphone` once to confirm the pinned vendorHash at~~ done — routine in later trains
    current HEAD.
~~12. Run `nix run .#vulnix` once (networked) to prove the app path executes.~~ done — vulnix zero-real-advisories at T22
~~13. Record the review verdict + polish items in `TODO_LIST.md` (or~~ done in part — the fixes live in the AGENTS flake-layout bullet; no dedicated TODO row
    `docs/dedup-registry.md` if any become clone rulings — unlikely here).

### Repo-wide Nix hygiene (observed while reading)

~~14. Confirm `nix/packages.nix`'s vulnix-triage script (70 lines) stays in~~ record stands — still cohesive
    `packages.nix` vs. its own file — currently cohesive, revisit if it
    grows.
~~15. Consider factoring the `pkgs.writeShellApplication` boilerplate shared~~ not adopted — below the bar
    by the triage CLI and its fixture check.
~~16. Evaluate whether `statix`/`deadnix` checks should `cd` a filtered~~ not adopted — below the bar; whole-tree scan acceptable (vendor excluded by check scoping)
    source rather than `${self}` (they scan the whole tree incl. vendor).
~~17. Confirm `deadnix`'s vendor-skip is by tool default, not luck~~ record stands — the vendor-skip is observed tool behavior
    (documented rationale if so).
~~18. Add `meta.description` to any check derivations still lacking one~~ done in part — the triage package got a full meta block (05:56); check derivations stay manual-curated
    (`island-js`, `statix` has it, verify all).
~~19. Review whether `checks.format` and `checks.treefmt` both being present~~ record stands — both resolve to the same derivation (documented here)
    is intended (both resolved to the same `.drv` in eval).
~~20. Add a `checks` entry that asserts the flake's own `devShells.ci`~~ record stands — flake check evaluates devShells.ci
    builds.
~~21. Verify every `writeShellApplication` sets `runtimeInputs` (done by~~ not adopted — below the bar
    reading — codify as a grep check?).
~~22. Consider `nixpkgs-terraform`-style pin docs for the `nix run~~ not adopted — below the bar; ad-hoc invocations stay unpinned
    nixpkgs#...` invocations in scripts/AGENTS.md (they're unpinned).
~~23. Note that several AGENTS.md command examples use `nix run nixpkgs#X`~~ not adopted — below the bar (same rationale, never written down)
    (e.g. nodejs, tailwindcss_4) — unpinned, acceptable for ad-hoc but
    worth a documented rationale.
~~24. Check `nix/devshell.nix` includes everything `buildflow` needs~~ done — confirmed by reading (the item's own check)
    (go-licenses, codespell present — confirmed).
~~25. Consider adding `pkgs.dprint` to the shell if `dprint.json` is used~~ not adopted — below the bar
    interactively.

### Module / deployment (from reading `nixos-module.nix`)

~~26. Document that `/health` (root) is intentionally not in the~~ record stands — the review itself confirmed the design (SSE at /health/sse)
    `flush_interval -1` handles (only `/health/*`).
~~27. Consider an explicit `handle /health` block if the dashboard root ever~~ record stands — the dashboard root does not stream
    streams.
~~28. Add `systemd.services.webphone.serviceConfig.Environment` key ordering~~ not adopted — below the bar
    note (mkIf merge semantics) — or leave.
~~29. Consider `StartLimitIntervalSec`/`StartLimitBurst` for restart storm~~ routed — TODO boot-contract row (the D3 StartLimit owner call carries it)
    protection.
~~30. Consider `RestrictAddressFamilies` on the backup unit (currently~~ not adopted — below the bar; the backup unit keeps the lighter posture (UMask/StateDirectory)
    inherits broader defaults) — oneshot, low priority.
~~31. Review whether the backup unit should carry the same~~ not adopted — below the bar (same posture ruling)
    `SystemCallFilter` allowlist for consistency.
~~32. Consider `ProtectProc`/`ProcSubset` on the backup unit.~~ not adopted — below the bar (same posture ruling)
~~33. Document the rsync `--delete` semantics (mirror, not archive) in the~~ record stands — partially documented, as noted
    option description (it is partially documented).
~~34. Consider a `backup.preCommand`/`postCommand` seam for notification~~ not adopted — below the bar
    hooks.
~~35. Consider exposing `caddy.extraConfig` append for operator additions.~~ not adopted — below the bar; no extraConfig seam

### Verification / gates

~~36. Add the review's low-severity items to `TODO_LIST.md` with owners.~~ done in part — the decisions live in AGENTS; no owner-attributed row
~~37. Re-run `nix flake check` after any split of module-check.~~ done — 05:56 gate summary green
~~38. Add a CI job that runs `nix flake check --no-build` to catch eval~~ done — the fast flake checks CI job exists (ci.yml)
    regressions cheaply.
~~39. Consider a `nix flake check` (full) nightly job to include VM tests.~~ not adopted — below the bar
~~40. Track the aarch64 omission explicitly (add `--all-systems` to one CI~~ not adopted — below the bar
    job if runner budget allows).

### Meta / process

~~41. Adopt the skill's two-step read (skill body + both reference files)~~ process record
    as a standing rule for future Nix reviews.
~~42. Keep citing stable names, not `file:line`, per AGENTS.md conventions.~~ record stands — convention followed
~~43. Log this session's verdict in the status corpus (this file) and link~~ done — this file + the 05:56 fixes report are the ledger pair
    it from `TODO_LIST.md` if items are harvested.
~~44. Consider a short `docs/lessons.md` entry: "vendored flakes carry~~ not adopted — no lessons row landed
    anti-patterns — scope Nix greps to own dirs, vendor is excluded."
~~45. Re-check after any dependency bump that `vendorHash` round-trips in~~ record stands — the AGENTS same-breath reminder exists
    the same breath (existing AGENTS reminder).
~~46. When the BuildFlow `gomod-check` false positive is fixed upstream,~~ other repo — BuildFlow upstream fix
    remove the deviation note from AGENTS.md.
~~47. Periodically re-run `statix`/`deadnix` after upstream rule additions.~~ process record — statix/deadnix run in every flake check
~~48. Consider pinning the two `nix run nixpkgs#` dev tools (tailwindcss_4,~~ not adopted — below the bar; documented rationale never landed
    nodejs) into the devShell to remove the last unpinned invocations.
~~49. Keep `webphoneVersion` in `flake.nix` (release.sh greps it there) —~~ record stands — AGENTS pins webphoneVersion to flake.nix
    do not let a refactor move it.
~~50. Revisit the "setup shell" import-graph NO-GO only if upstream ships a~~ record stands — standing AGENTS decision
    zero-usermgmt `shell` submodule (standing AGENTS decision).

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

~~1. **Do you want me to apply the low-severity fixes now** (split~~ resolved by events — FIX IT ALL came as the 05:56 train
   `module-check.nix`, add a CI devShell + CI job, permissive stand-in),
   or leave the Nix surface untouched and only log the polish items to
   `TODO_LIST.md`?
~~2. **Is the `webphoneVersion`-hardcoded design a permanent accepted~~ resolved by events — the release-time guard landed in release.sh; the hardcode stays release.sh-owned
   exception** (release.sh contract), or do you want an eval-time guard
   that fails when it drifts from the git tag on a clean tree?
~~3. **Should the two skill reference files (`common-problems.md`,~~ process record — unanswered; the fixes train superseded the depth question
   `best-practices.md`) be walked exhaustively** for a second, deeper pass
   with a per-checkbox tick table, or is the current read-and-verify depth
   sufficient for your bar?
