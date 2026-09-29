# Status Report — Nix Review Train (flake + NixOS module)

_Point-in-time snapshot: 2026-09-29 03:05 CEST · Session scope: `nix*` skills review train only — no unrelated research was done._
_Repo: `webphone` @ `2808a90` (daemon commits 83cdf46 → 2808a90 are this session's work). Format: Markdown (explicit user request — skill's canonical format is HTML; flagged per skill contract)._

---

## Executive summary

A full nix-review (checklist-driven, both `.nix` files, 1,267 lines read in full) of the flake and the shipped NixOS module. Every "found issue" was premise-verified before touching code — which **refuted the biggest suspected finding** (checks allegedly not gating) and confirmed three real ones. Fixed: 1 correctness gap (missing `destDir` assertion), 1 ordering race (backup oneshot vs `webphone.service`), 1 security gap (world-readable private comms data), 1 structural debt (803-line flake.nix). Final state: **`nix flake check` — all 20 checks green** including the KVM backup VM test running the hardened units, `nix build .#webphone` → 2.7.0, aarch64 eval green, statix/deadnix/nixfmt clean. Documentation (AGENTS.md, CHANGELOG) updated. Nothing in the end state is broken.

The honest headline for section (d): the **process** stumbled three times (self-inflicted split bugs caught before or by gates), and one premature "fix" was stopped by evidence-checking. The end state is clean; the first draft was not.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                    | Evidence                                                                            |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| 1  | **Skill-driven review completed** — both `.nix` files categorized and checked against every checklist category (critical, purity, structural, correctness, consistency, NixOS module, security, performance, devShell, overlays)        | 0 files skipped; findings logged below                                              |
| 2  | **Premise verification before fixing** — locked-stdenv `pipefail` check (`set -euo pipefail` at setup:12 via drv-path extraction) **refuted** the suspected "four non-gating checks" finding; no pointless change shipped               | `/nix/store/c29a7xz…-stdenv-linux-no-cc/setup` lines 11–12                          |
| 3  | **`backup.destDir` assertion** — same `/var/lib/` + not-root shape as `dataDir`, gated on `backup.enable` via `lib.optionals`                                                                                                           | `package/nixos-module.nix` assertions block                                         |
| 4  | **Assertion pinned by a test** — new `backup-destdir-assertion` linkFarm entry in the module check: bad destDir trips exactly 1 failed assertion, clean eval trips 0 (the avatarFor lesson applied on day one)                          | artifact output: `backup.destDir assertion fires exactly off-/var/lib`              |
| 5  | **Backup ordering race fixed** — oneshot orders `after = [ "webphone.service" ]` so a Persistent timer catch-up at boot cannot race db creation                                                                                         | module diff; VM test green                                                          |
| 6  | **Private data no longer world-readable** — `UMask=0077` + `StateDirectoryMode=0750` on BOTH units (main + backup)                                                                                                                      | module diff; VM test (NRestarts=0, integrity_check, prune assertions) green         |
| 7  | **flake.nix split** — 803-line monolith → slim entry + 7 focused flake-parts modules under `nix/`: `packages.nix`, `checks.nix`, `module-check.nix`, `vm-tests.nix`, `apps.nix`, `devshell.nix`, `treefmt.nix`                          | flake.nix now ~57 lines; all 11 checks + 3 packages + app + devShell eval and build |
| 8  | **release.sh contract preserved** — `webphoneVersion` binding stayed in flake.nix (release.sh greps/seds it there; grep count = 1, sed pattern still matches); reaches `nix/packages.nix` via `{ _module.args.webphoneVersion = …; }`   | `grep -c 'webphoneVersion = "2.7.0";' flake.nix` → 1                                |
| 9  | **Full gate green** — `nix flake check`: **all 20 checks passed**, including the KVM `webphone-backup` VM test executing the hardened units, the drill, island-js, island-lint, vulnix-triage, format, statix, deadnix                  | gate output: `all checks passed!`                                                   |
| 10 | **Explicit builds verified** — `nix build .#webphone` → `webphone-2.7.0`; `.#checks.x86_64-linux.webphone-module` → artifacts built; aarch64 version evals `"2.7.0"`                                                                    | store paths in transcript                                                           |
| 11 | **Deliberate decisions re-confirmed, not churned** — hardcoded version (release contract), no-`go-standard` (zero private deps + public-flake burden), rejected items left rejected                                                     | report in final message                                                             |
| 12 | **Memory maintenance** — AGENTS.md: new "flake.nix layout" block (file map, version-binding contract, `self`-is-top-level-only gotcha, hardening summary, pipefail-safe note); CHANGELOG Unreleased: Changed / Fixed / Security entries | AGENTS.md + CHANGELOG.md, committed by daemon (2808a90)                             |
| 13 | **Path-bug caught pre-eval** — `nix/packages.nix` `root = ./.` would have sourced the `nix/` directory; self-caught during writing, fixed to `./..` before any gate ran                                                                 | write + immediate multiedit in transcript                                           |

## b) PARTIALLY DONE

| # | Item                                       | Done                                                                               | Missing                                                                                                                                                                                                                                                                                          |
| - | ------------------------------------------ | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | aarch64 verification                       | flake eval + version eval for `aarch64-linux`                                      | No cross **build** + ELF-bytes check (release-time step per docs/lessons.md; deliberately deferred)                                                                                                                                                                                              |
| 2 | Live behavior proof of the hardened module | KVM VM test green (exercises UMask/StateDirectoryMode/after incidentally)          | No **eval-time pin** in the module check for UMask/StateDirectoryMode/`after` (see f-13/14), and no behavioral `stat` assertions in the VM test (f-32/33)                                                                                                                                        |
| 3 | Tri-repo integration                       | Local gates green; daemon committed everything                                     | **Stack-side verification not started**: nix-international-telephony rides `main`; its relock + `nix flake check` + browser E2E happen at next train close. Also **remote is behind**: `git ls-remote` = `46ad1d3`, local HEAD = `2808a90` — six session commits not yet pushed (daemon cadence) |
| 4 | Documentation                              | AGENTS.md + CHANGELOG updated                                                      | docs/lessons.md war-story entry for the stdenv-pipefail verification story not written (currently one line in AGENTS.md)                                                                                                                                                                         |
| 5 | Existing-deployment hardening              | New files created by the units are now 0600/0700                                   | **Existing** production files keep old modes until a one-time chmod / re-backup — owner ops decision, documented in CHANGELOG only                                                                                                                                                               |
| 6 | Output-parity proof of the split           | Attr-by-attr evals (11 checks / 3 packages / app / devShell / modules) + full gate | No recorded pre/post `nix flake show --json` diff against the pre-split commit (still doable post-hoc via a worktree at `46ad1d3` — f-28)                                                                                                                                                        |

## c) NOT STARTED

| # | Item                                                                                                                                          | Why it surfaced                                                                                                          |
| - | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| 1 | Backup-unit hardening to main-unit depth (`SystemCallFilter`, `ProtectProc`, `CapabilityBoundingSet`)                                         | Review noted the asymmetry; deferred to avoid churn on a drill-verified oneshot without a dedicated re-verification pass |
| 2 | Shell-quoting of interpolated `${cfg.dataDir}`/`${dest}` paths in the backup script (spaces would break them)                                 | Noted; declined this train (drill-verified script bytes; gain ≈ 0 for realistic configs)                                 |
| 3 | `dataDir`/`destDir` **whitespace** policing (systemd `StateDirectory` treats spaces as list separators)                                       | Noticed while writing the assertion; not decided                                                                         |
| 4 | `destDir`↔`dataDir` **nesting** assertion (backup dest inside live data — or the reverse — would make `rsync --delete` semantics treacherous) | Noticed during assertion work; needs an owner decision on legality first                                                 |
| 5 | `ReadWritePaths` redundancy in the backup unit (`destDir` == its StateDirectory)                                                              | Noticed; harmless, kept                                                                                                  |
| 6 | vulnix app cwd guard (assumes repo root; `jq` on missing `flake.lock` fails opaquely)                                                         | Review observation                                                                                                       |
| 7 | CI devShell (skill checklist item)                                                                                                            | Checks run sandboxed with `nativeBuildInputs`; low value here                                                            |
| 8 | HARVEST of section (f) into TODO_LIST.md / ROADMAP.md                                                                                         | Skill contract: belongs to `docs-health` HARVEST; awaiting instruction (user said wait)                                  |
| 9 | Re-run `nix flake check` after the AGENTS/CHANGELOG edits                                                                                     | Docs are not gated (markdown outside treefmt scope); final-green snapshot not re-taken                                   |

## d) TOTALLY FUCKED UP

**Nothing in the end state** — every gate green, no regressions shipped, no reverted work. What _was_ fucked up mid-flight, caught and corrected (recorded so the pattern is visible):

1. **Almost shipped a no-op "fix"**: the top-ranked finding was "four checks don't gate (`\| tee $out` masks failures)". One grep of the locked stdenv refuted it (`pipefail` is set). Had I "fixed" it, the diff would have been harmless but the claim false — a textbook unverified-premise mistake, avoided by verification **only because I checked before editing**.
2. **First split draft had three self-inflicted bugs**: (a) `flake-parts-lib.importApply` used outside its scope (`undefined variable`); (b) `self` destructured inside `perSystem` (flake-parts error: top-level-only arg); (c) `root = ./.` in `nix/packages.nix` pointing at the wrong directory — self-caught before eval. All were structural-split mistakes from not re-reading flake-parts' module-args model first.
3. **Malformed verification probe**: `nix eval .#nixosModules.default --apply 'm: m.import'` — wrong shape, wasted a cycle and produced a misleading "error" in the transcript (the module was fine).
4. **Edit-tool mtime-guard fumbling**: two rejected edits after the daemon touched file mtimes; required a re-view cycle (correct behavior, sloppy sequence).

## e) WHAT WE SHOULD IMPROVE

1. **Premise-check discipline is load-bearing** — the pipefail refutation is the session's most valuable non-change. Codify: no fix diff without evidence of the failure mode (locked-stdenv grep, drv inspection, eval probe).
2. **Read the framework's module-args model before restructuring** — all three split bugs were flake-parts scoping assumptions (importApply, `self`, path roots). A 2-minute docs read would have saved three gate cycles.
3. **Move-rewrites need a path checklist** — every `./x` becomes `../x` and every `./.` becomes `./..`; make it an explicit reflex, not a catch (this one _was_ caught, by luck of rereading).
4. **New behavior should be pinned at the same moment it is written** — the UMask/StateDirectoryMode/`after` hardening shipped with VM-test coverage but no eval-time pin in the module check (the repo's own standard: "the avatarFor lesson"). Fix in f-13/14.
5. **statix/deadnix belong in the devShell** — the style finding surfaced only at gate time; local tools would catch it per-edit.
6. **Record output-parity evidence for pure moves** — a pre/post `nix flake show --json` diff would turn "I moved code" into "outputs are byte-equivalent" (f-28).
7. **Local commit ≠ pushed** — end-state verification (`git ls-remote`) must stay a standing step; the daemon's push lags its commit (currently 6 commits behind origin).

## f) NEXT — up to 50 things to get done

_**[R]** = READY (small, session-adjacent, actionable now) · **[T]** = needs a train/ritual · **[?]** = needs an owner decision (see g) · **[O]** = ops/host action. Ordered by impact._

1. **[R]** Add eval-time pins for the new hardening to the module check: assert `UMask == "0077"` and `StateDirectoryMode == "0750"` on both units (currently only incidentally covered).
2. **[R]** Add eval-time pin for backup ordering: `service.after == [ "webphone.service" ]` in a linkFarm entry.
3. **[?]→[R]** Assert `backup.destDir` is not inside `dataDir` (and vice versa) — needs your legality call (g-3).
4. **[R]** Add statix + deadnix to `devShells.default.packages` (local catch before the gate).
5. **[R]** VM test: assert backup files are not world-readable (`stat -c '%a'` → 600/700) — behavioral pin of the UMask fix.
6. **[R]** VM test: assert `/var/lib/webphone` mode is 0750.
7. **[T]** Stack-side integration: nix-international-telephony relock at next train close + its `nix flake check` + browser E2E against the hardened module.
8. **[R]** Verify daemon push caught up (`git ls-remote` vs `2808a90`) — end-state contract.
9. **[T]** Live smoke of the built binary before next release: `scripts/webphone-smoke.py --bin $(nix build .#webphone) --expect-version v2.7.0`.
10. **[T]** aarch64 cross build + ELF-bytes verification (release ritual step, not yet run this train).
11. **[R]** HARVEST this report's section (f) into TODO_LIST.md / ROADMAP.md via docs-health (skill contract; awaiting instruction).
12. **[R]** docs/lessons.md entry: "stdenv sets `set -euo pipefail` — verify premises against the locked stdenv before 'fixing' check plumbing" (with the drv-grep technique).
13. **[R]** Pin the `dataDir` assertion at eval level too (only `destDir` has a dedicated check entry today).
14. **[?]→[R]** Whitespace policy for `dataDir`/`destDir` (StateDirectory space-splitting) — assertion vs documented constraint (g-3 follow-up).
15. **[R]** One-command `nix build .#checks.x86_64-linux.<name>` convention documented in AGENTS.md Commands block (used repeatedly this session, currently tribal knowledge).
16. **[R]** Output-parity proof: worktree at `46ad1d3`, `nix flake show --json` diff vs HEAD (closes b-6; expected: identical).
17. **[R]** vulnix app cwd guard: fail with "run from the repo root" when `flake.lock` is absent.
18. **[R]** Exhaustive grep for other scripts referencing `flake.nix` structure (release.sh + .buildflow.yml done; one final sweep of scripts/).
19. **[R]** Confirm `checks.treefmt` pre-existed the split (one git-archaeology command) — closes the one unproven observation.
20. **[R]** Remove-or-justify `ReadWritePaths` in the backup unit (redundant with StateDirectory) — decide, then comment or delete.
21. **[O]** Existing-deployment tightening: one-time `chmod`/re-backup so current files match the new UMask (g-1).
22. **[?]** Whether the stack should be verified NOW vs at train close (g-2) — if NOW, pull item 7 forward.
23. **[R]** Backup-unit hardening (item c-1) behind a re-run of the VM test + drill in the same change.
24. **[R]** Shell-quote the interpolated paths in the backup script + drill re-run (item c-2).
25. **[R]** `nix flake check --all-systems` trial: close the "omitted incompatible systems: aarch64-linux" warning in the release ritual.
26. **[R]** Verify the KVM-less skip warning path once (AGENTS claims skip-with-warning; unproven this session).
27. **[T]** Release ritual for the next train will fold CHANGELOG Unreleased (Changed/Fixed/Security) — nothing to do now, listed to keep the fold honest.
28. **[R]** Re-run `nix flake check` for a truly-final green snapshot post-doc-edits (c-9) — cheap.
29. **[R]** nil/LSP sanity over the new `nix/*.nix` via the project crushrc (crushrc line 8 references the devShell; confirm the LSP picks the new dir up).
30. **[R]** AGENTS.md one-liner: `self` gotcha is now documented; consider the same note inline as a comment in `nix/checks.nix`/`nix/vm-tests.nix` (already present — verify it survived fmt; it did).
31. **[R]** Consider a `formatter` alias sanity check (`nix fmt` used this session; confirm it's in `nix flake show`).
32. **[R]** Decide the fate of the `nixosModules.webphone` alias (documented as consumer convenience; keep — pin with a one-line check that both attrs import the same file).
33. **[R]** `memoryMax`: add a README recommendation paragraph (checklist wanted MemoryMax set; owner prefers null-by-default — document the tradeoff).
34. **[R]** `/metrics` vhost location: consider a module helper for scraper fencing (`allow/deny` snippet generation) — ROADMAP fuel, product decision.
35. **[R]** backup `calendar` option: doc example uses `*:00/15:00` — verify that systemd syntax form is valid in a VM test run (untested option path).
36. **[R]** `nginx.hsts.maxAge` unit-test the rendered header value with a custom maxAge (only default 63072000 pinned today).
37. **[R]** `serverTiming` env gate: pin `WEBPHONE_DEBUG_TIMING` absence when the option is off (only presence is pinned).
38. **[R]** CSRF typed-override check: add the empty-typed-lists case ("empty lists never clobber raw values" is documented but unpinned).
39. **[R]** Module `environmentFiles` (plural) seam: no check entry exercises it — add one (stack uses this seam per its header comment).
40. **[R]** VM test currently asserts NRestarts=0 for the no-restart claim — also assert the oneshot's own result is `success` (`systemctl show -p Result`).
41. **[R]** Drill: assert restored file mode ≤ 0640 (privacy carries through restore).
42. **[T]** Next erraudit tier-2 re-measure is due 2026-10-22 (standing AGENTS cadence; unrelated to this train, listed so it isn't lost).
43. **[R]** README: mention the new `nix/` module layout in the contributor/hacking section (one paragraph; README sells, so keep it short).
44. **[R]** Consider `deadnix --report` (non-fail) locally documented for quick scans while authoring.
45. **[R]** AGENTS.md "Tri-repo integration rules": add the UMask/StateDirectoryMode hardening to the module options bullet (options unchanged, but deployment shape changed — the bullet lists deployment shape).
46. **[R]** Add `lib.optionals cfg.backup.enable`-style gating audit: confirm no other assertion can fire for unused options (dataDir always used; only destDir gated — verified this session; note it in the module header comment).
47. **[R]** Consider renaming `checks.webphone` → keep as-is (checklist's `checks.build` analog; naming matches the package) — documented decision, no action; listed to close the checklist item explicitly.
48. **[T]** At next release: bump `webphoneVersion` via release.sh ONLY (sed target verified this session) — ritual reminder.
49. **[R]** Brutal-self-review pass over this nix train (sibling skill; "what did we get wrong" beyond this report's inventory).
50. **[R]** Delete stale store garbage from the failed first statix run (`l44w1dpp…-statix-check` output path from the pre-fix build) — cosmetic hygiene, `nix store gc` covers it.

## g) Questions I can NOT figure out myself

1. **Existing data on the production host**: should the UMask/`StateDirectoryMode` tightening be applied retroactively (one-time `chmod -R go-rwx /var/lib/webphone /var/lib/webphone-backup` + re-backup), or is new-files-only acceptable until the next backup cycle? I cannot see the host, and the right answer depends on your downtime/ops window.
2. **Stack verification timing**: is the consuming stack's per-train relock the intended vehicle for this module change (my default assumption), or do you want nix-international-telephony relocked + browser-E2E'd **now** (the module's unit environment changed, not just packaging)?
3. **`backup.destDir` / `dataDir` nesting legality**: would you ever legitimately configure `destDir` inside `dataDir` (or the reverse)? If never, I'll add a forbid-assertion (f-3); if sometimes, it needs documented semantics instead.

---

_Skill-contract notes: format override (user requested `.md`; canonical is styled HTML) — flagged, not propagated into the skill. Manual commit skipped per Crush harness contract ("NEVER COMMIT unless the user says commit"); the auto-commit daemon picks the file up. Section (f) is the primary input for docs-health HARVEST — run it when instructed._
