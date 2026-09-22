# Status — v2.3.0 train complete, pbx-artmann riding it, one statix gate red

**Written:** 2026-09-20 00:14 CEST · **Scope:** continuation after the
23:43 report — closing out the release train and the TODO_LIST tail.
Point-in-time snapshot; annotate, never rewrite.

**Headline:** **v2.3.0 is fully released** (RELEASE-EXIT=0: gates, tag,
lychee, stack relock + browser E2E + full stack flake check, aarch64
cross-builds, gh release object) and **pbx-artmann is relocked to the
stack, pushed, and its prod toplevel pre-builds green** — the owner
redeploy is a single `nixos-rebuild test` away (g1). P5 closed: browser
E2E green twice (instrumented standalone run + in-train run, transfer
initiated in 4.5 s). One real red opened at the buzzer: webphone's full
`nix flake check` now fails the **statix single-assignment gate** on my
new backup-timer module block — known fix, not yet applied (d1).

---

## a) FULLY DONE (all verified, not assumed)

| #  | Work                                                                                                                                                                                                                                                                                                                                                                           | Evidence                                                                                             |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| a1 | **v2.3.0 released end to end** — the credential-verification fix, contacts/dial integration, health triple (`/livez` `/startupz`), Go 1.27.1 floor, nanoid v1.65.1, drift-guard release-window tolerance, 28-check smoke                                                                                                                                                       | `/tmp/release-2.3.0.log` RELEASE-EXIT=0; https://github.com/LarsArtmann/webphone/releases/tag/v2.3.0 |
| a2 | **P5 closed — browser E2E green ×2**: standalone instrumented run (18.8 s to transfer) + in-train run (4.5 s); the timestamped transfer-row dumps stay as a permanent flake probe                                                                                                                                                                                              | `/tmp/e2e-run5.log`, `/tmp/release-2.3.0.log`                                                        |
| a3 | **aarch64 gates for the new build** ran inside the train (package + island-lint cross-builds) — closes the sibling session's b8                                                                                                                                                                                                                                                | release.sh step 8, exit 0                                                                            |
| a4 | **pbx-artmann relocked and pushed**: webphone input c27a22b → 7581858 (v2.3.0); lock commit `7a1c081` on master (the earlier push failure was mine: wrong branch name — the repo's default is `master`)                                                                                                                                                                        | `git log origin/master -1`                                                                           |
| a5 | **pbx toplevel pre-build GREEN**: the full prod system closure builds with v2.3.0                                                                                                                                                                                                                                                                                              | `/tmp/pbx-toplevel.log` TOPLEVEL-EXIT=0                                                              |
| a6 | **P17 deliverables shipped**: `scripts/webphone-backup-drill.py` (committed, prod-safe pkill scoping, re-run PASSED after commit), README "Backups and restore" section (inventory, online snapshot commands, drill-verified restore steps), NixOS module `services.webphone.backup.{enable,destDir,calendar}` daily online timer (sqlite `.backup` + blob rsync, no downtime) | commits `95fb273`, `bdaff93`, the module commit; module eval check EXIT=0                            |
| a7 | **TODO_LIST harvested**: drift-guard, P17, P24 rows deleted; watch row updated (nanoid closed, FEATURES VERIFY done, hotfix pre-draft moot); redeploy row now names v2.3.0                                                                                                                                                                                                     | TODO_LIST commits                                                                                    |
| a8 | **Stack flake check "all checks passed" ×3 inside the train** (fax-feed included since tonight)                                                                                                                                                                                                                                                                                | `/tmp/release-2.3.0.log`                                                                             |

## b) PARTIALLY DONE

| #      | Item                                                                           | State                                                                                                                                                                                                                                            | What remains                                                                                                                                                                      |
| ------ | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~b1~~ | ~~**webphone full `nix flake check` is RED at the statix gate**~~              | ~~~`package/nixos-module.nix` now has THREE top-level `systemd` assignments (webphone, webphone-backup service, backup timer); statix demands one `systemd = { ... }` attrset — the exact pin AGENTS documents for the stack's nginx locations~~ | ~~Merge the three blocks into one attrset, re-run `nix flake check`, push. Fix is mechanical, ~10 min~~ done at `8b39eab` (01:04 report a1)                                       |
| ~~b2~~ | ~~**Release-ops remainders**~~                                                 | ~~three gh release objects live~~                                                                                                                                                                                                                | ~~CHANGELOG bottom link refs + announcement drafts (owner approves posting)~~ done (link refs + drafts shipped 01:04; posting remains the owner-gated TODO row)                   |
| ~~b3~~ | ~~**P25 idiomorph**~~                                                          | ~~untouched, gate is green (E2E ×2)~~                                                                                                                                                                                                            | ~~branch + experiment + verdict doc~~ done (executed 01:04, merged to main; morph swap is the live-update contract)                                                               |
| ~~b4~~ | ~~**Sibling session's self-health leftovers** (their 22:29 report owns them)~~ | ~~~—~~                                                                                                                                                                                                                                           | ~~~`/livez` consumer decision, 18-49 review annotation, go-health doc.go, cqrs-htmx v4.11.0 release-object audit~~ done (routed — annotated in the 22:29 report, 2026-09-22 pass) |

## c) NOT STARTED (owner-only, unchanged)

1. **Prod redeploy** — v2.3.0 folds everything; one rebuild (g1 below).
2. ~~Island sanitization alignment (owner decision).~~ done (sanitization aligned + pinned, DECIDED 2026-09-20 (CHANGELOG))
3. ~~Own-number DID feed (owner decision).~~ done (identities config map shipped, DECIDED 2026-09-20 (CHANGELOG))
4. SMS-bridge journal grep (owner-only access) (g3 below).

## d) TOTALLY FUCKED UP (this stretch, no spin)

1. **The statix gate slipped through my "verified" claims twice.** I ran the module eval check and declared it green — but the first "green" was the pipe-masked exit-code trap (`nix build … | tail -1` reports tail's status), and when I did look, I fixed the WRONG surface (added the `systemd.timers` stand-in to the flake check) instead of reading the statix error that the FULL flake check then surfaced. The eval stand-in fix was necessary but never the whole story. The honest state at 00:14: main's full `nix flake check` is red.
2. **My module edit was clobbered by the sibling session's editor buffer** — my first application of the backup block vanished from disk (file matched pre-edit HEAD) while git status showed nothing of it. I detected it only because a post-edit `grep -c` came back 0. The re-apply + immediate explicit commit is the working defense; the real defense is not letting two agents own one file.
3. **The daemon committed my half-finished work at least three times tonight** (splitting commits, committing before my explicit `git add`). No damage this stretch, but the commit-at-phase-boundary discipline from the sibling report (their e1) is clearly still aspirational.
4. **Wrong branch name on pbx-artmann's first push** (`main` vs `master`) — trivial, but it means I assumed instead of reading `git status -b`.

## e) WHAT WE SHOULD IMPROVE

1. **Exit codes never through pipes** — tonight it bit me twice more (`nix build | tail` masking failures). Use `set -o pipefail` or capture-to-file + separate verdict command, as muscle memory.
2. **Post-write grep-back for EVERY structural edit** when a sibling session is live (`grep -c` the inserted marker immediately after write, not after the next gate).
3. **statix belongs in the inner loop for module edits**: run the statix check (or full flake check) the moment a .nix file changes — the module eval check alone cannot catch style-gate failures.
4. **Single-owner files**: the module + CHANGELOG need an explicit claim convention between concurrent sessions (a `TODO_LIST` line or commit-message lock note) — tonight's clobber was invisible until grepped.
5. **The release train is now proven twice** (2.2.0 resume-mode, 2.3.0 fresh) — the remaining script gap is the statix/treefmt class of gate: `nix flake check` inside the train already covers it; nothing to add, only to keep the tree green BEFORE starting the train.

## f) NEXT STEPS (ordered; owner-independent unless marked)

1. ~~Fix the statix single-assignment violation in `package/nixos-module.nix` (merge the three `systemd` blocks) → full `nix flake check` green → push.~~ done at `8b39eab`
2. ~~Re-run `nix run .#vulnix` on the NEW runtime closure (go-health + samber/do transitive + cqrs-htmx v4.11.0 are new since the last scan; glibc triage note is the baseline).~~ done (re-scanned 01:04; release.sh now gates on vulnix)
3. ~~CHANGELOG bottom link refs for v2.1.0/v2.2.0/v2.3.0 + announcement drafts (owner approves posting).~~ done (link refs + drafts done 01:04; posting stays owner-gated (TODO_LIST))
4. ~~P25 idiomorph experiment branch + verdict doc (gates are green).~~ done (idiomorph merged to main)
5. ~~Fold-back check: confirm the stack's lock still rides webphone main after (1) lands — if the fix commit lands after the relock, one more `nix flake lock --update-input webphone` in the stack + pbx-artmann relock dance.~~ done (stack re-pins verified per train; 2026-09-22 again (550aaea))
6. (owner, g1) Prod redeploy of v2.3.0 + post-deploy probes (bogus-creds gate must go green).
7. (owner, g2) Release cadence preference.
8. (owner, g3) SMS-bridge journal grep.
9. ~~Sibling-session handoffs (b4 list).~~ done (routed - annotated in the 22:29 report (2026-09-22 pass))
10. Version-drift guard wiring INTO buildflow (currently a Go test + release-script fold-check; the buildflow step was the row's original ask).

## g) OWNER QUESTIONS (3)

1. **Prod redeploy go + version**: v2.3.0 is fully released, pbx-artmann's toplevel pre-builds green — deploy v2.3.0 now (`nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech`, probe `scripts/webphone-smoke.py --base https://pbx.artmann.tech` must show `bogus credentials rejected` green, then `switch`)? Recommended: yes, now — it closes tonight's found-and-fixed forged-session hole on prod.
2. **Release cadence**: same-day minor releases per batch (2.2.0 + 2.3.0 tonight, recommended: small, semver-honest, deployable) or one release per day maximum?

   _(Answered — the recorded cadence rule is "train on a user-visible
   theme", cited by the TODO_LIST train-cut row; v2.4.0 followed it.)_
3. **SMS bridge**: the 422/502 root cause needs prod-only journal access — run `journalctl -u telnyx-webhooks --since today | grep -iE "sms|422|error"` on pbx and paste the tail, or should bridge health move into the operator window so the SMS lane stops depending on owner hands?
