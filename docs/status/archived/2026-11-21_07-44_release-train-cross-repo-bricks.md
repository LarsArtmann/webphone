# 2026-11-21 07:44 — release train v2.15.0: three cross-repo bricks found and repaired, release blocked on the host storm

> ARCHIVED 2026-11-21 (docs-health v10 sweep): pre-drain verified (check-rows --strict + bare-rows green), drained for cadence — the pin sat at 66/66 with 12 gates-passing candidates

Session arc: resume the 2026-11-15 sidebar-rail handover → discover six days of parallel
sessions had already landed the rail, folded the CHANGELOG, and drained cadence → the real
remaining work was "cut v2.15.0 and deploy" → cutting it surfaced three UNRELATED bricks
(a red main, a poisoned lens FOD, a broken tag ancestry), all repaired → the release itself
is still pending, blocked on a machine-wide load storm from parallel sessions.

> **Live remainder:** the release + deploy chain lives in §f rows 1–8 (TODO_LIST §
> release-v2150-chain); the owner questions in §g.

## a) FULLY DONE

1. **The §g integrity repair (the handover's top item).** The archived rail report carried
   two FABRICATED owner answers ("ANSWERED: 5 + More", "ANSWERED: T2 next") on questions
   the owner never answered. Both struck with inline corrections; Q1 keeps its answer (the
   2026-11-15 "SUPERB mix" directive IS the ratification). Committed `afc37d21`, pushed.
2. **Webphone main was EVAL-RED and is healed.** The hardening-parity refactor moved the
   NixOS module `nix/package/` → `package/` and updated flake.nix but left the three
   vm-tests resolving `../package/` INSIDE nix/ — every eval of the VM-test checks died
   "Path 'nix/package/nixos-module.nix' does not exist", which is what looped the release
   runbook's buildflow nix-build step five consecutive times (rc=69). Fixed the three
   imports to `../../package/nixos-module.nix` (commit `81b52f3b`, pushed); all 16 checks
   evaluate again; CI's `nix build .#webphone` and `go test -race ./...` jobs are green at
   that commit.
3. **Tag-ancestry repair.** v2.14.0 was cut from the `release-line` branch, NOT main, so
   `git describe` from main answered v2.13.0 and the release version-guard refused.
   `git cherry` proved both release-line commits patch-equivalent to commits already on
   main; merged release-line into main (`aa6245b8`) as a PROVEN content no-op
   (`git diff HEAD^1 HEAD` = 0 files). describe now answers v2.14.0.
4. **The pbx-lens vendorHash brick (third occurrence of the 2026-11-02 class).** The lens
   repo's `vendorHash.nix` carried the `sha256-AAAA…` placeholder again (daemon
   auto-commits re-poisoned it after the f0d275d repair), so every fresh build of the lens
   FOD died "hash mismatch" — this is EXACTLY why telephony main's CI has been failing its
   serialized VM suites since 2026-11-19 (`ccd3249`). Repaired: probe-built the real hash
   in an isolated worktree (`sha256-ELxYnKt…`), committed it ALONE (`0b83c4f`) so the
   parallel session's in-flight serve-package refactor stayed staged, verified the
   committed tree builds green (rc=0), pushed. go.mod/go.sum unchanged since a82a283, so
   the hash is valid for the whole committed line.
5. **Telephony relocked onto the healthy lens.** `nix flake update pbx-lens` → pins
   0b83c4f; telephony toplevel evals green; pushed as `508c789`. Telephony consumes
   `inputs.pbx-lens.packages.<sys>.default` (the lens's own builder), so the green
   worktree build IS the consumer proof.
6. **pbx-artmann docs-numbers drift healed.** The parallel trains grew the Go suite
   584 → 599 tests without the doc bump; the gate reded the first battery run with the
   swept docs-gates v0.3.1. Bumped the three claims (AGENTS/README/FEATURES); standalone
   `docs-numbers .` answers CONSISTENT. (Edits swept into daemon auto-commits.)
7. **docs-gates v0.3.1 sweep verified legitimate.** The 01:58 daemon lock move is
   eb2fa6b (v0.3.0) → 67d7d27b = the v0.3.1 release commit, CI SUCCESS 2026-11-14.
   Battery run against it: freshness/annotate/routing/cadence/land-parity PASS;
   only `numbers` was red (the 584 drift, fixed in item 6).
8. **Pre-flight verification of the release path.** `release.sh 2.15.0 --dry-run` passes
   all nine steps; CHANGELOG has the folded [2.15.0] section; the tag is free; the
   version bump is already on main (see d-3).

## b) PARTIALLY DONE

1. **The v2.15.0 release itself.** Dry-run green; the REAL run launched, completed step 3
   (version bump commit `2be14b73`), then died rc=69 in step 4 — buildflow's nix-build
   failure loop, root cause the eval-red main of item a-2 (since healed). Re-run blocked
   by the host storm (§c-1).
2. **Webphone CI health at HEAD.** Green at my import fix; the two newest queued runs'
   verdicts pending (runner backlog 1h+). The parallel "whole-list session" already
   repaired the art-dupl red (2 new intentional chip-idiom groups: baselined + registry
   entry 02:5x) and the tw.css freshness advisory (rebuild committed via daemon sweep
   `404da12c`) — I verified both locally: format check rc=0, css rebuild rc=0 with the
   bytes already committed, `art-dupl check` answers 0 new groups.
3. **The release's stack leg.** release.sh step 7 will relock telephony's webphone input
   to the tag and push — not started (depends on the release).

## c) NOT STARTED

1. The release re-run — hard-blocked on the storm: load sustained ~1000+ for hours
   (6+ parallel Crush sessions' builds), 985 commits/8min churn at peak, and
   release.sh's `assert_clean_tree` between the bump and the tag cannot win against
   parallel commits.
2. Everything downstream of the tag: lychee, stack relock, aarch64 cross-builds +
   ELF guard, gh release.
3. pbx-artmann: telephony relock (one `nix flake update telephony` covers BOTH the lens
   fix and the webphone tag once step 7 lands), lock-drift-probe (webphone pin MATCH),
   restage-and-gate, deploy-freshness FRESH over a SETTLED tree (a parallel session was
   mid-refactor there earlier — avbrelay/telnyxaicall renames + cdrline WIP — it has
   since landed and the tree is clean).
4. The owner switch handover (no assistant SSH) + paste-backs.
5. Post-deploy: ui-track capture vs the pre-rail run, /version=v2.15.0 probe,
   verify-live webphone arms, CHANGELOG deploy ledger + FEATURES closure stamp.
6. Watching the queued CI verdicts (webphone ×2, telephony ×2 — one is fe6eb88 which
   will red on the PRE-fix lens pin; superseded by 508c789, annotate if it completes red).

## d) TOTALLY FUCKED UP

1. **Launched a ~30-60 min release into a load-56 box without checking the trajectory.**
   The load storm was already building; the release died on an eval-red main it had done
   nothing to cause, and the storm then ate the re-run window entirely. A 30-second
   `uptime` trend check before `release.sh` was the missing step.
2. **The release dry-run checks PRECONDITIONS, not tree health — I treated it as a green
   light.** `nix flake check --no-build` (or the buildflow nix-build step) is the eval
   authority; the dry-run never evaluates the tree, so an eval-red main passed
   preconditions and wasted a full run. Lesson recorded in §e.
3. **My `art-dupl check` run mutated the shared `.art-dupl-baseline.json`** (recordedAt
   rewrite — a tool side effect I did not expect), at the same moment a parallel session
   was performing its own legitimate re-baseline. I almost triaged and re-committed over
   their repair. The stamp-ownership lesson generalized: verify WHO owns a modified
   shared file before acting on it (check recordedAt timestamps + registry log against
   your own actions).
4. **Wasted rounds on the lens `undefined: serve` ghost.** The compile error was a
   parallel session's MID-FLIGHT refactor (staged rename + untracked wrapper), invisible
   in my first `status` reads because the session STARTED between my commands. I burned
   several diagnostics (ls-files truncated by `head`, check-ignore, local build) before
   taking a full-state snapshot. Full-state-first, not incremental probes, on a box this
   contended.
5. **The half-release state is still on main** (flake.nix says 2.15.0, no tag). It is
   benign for the re-run (step 3 sees "already 2.15.0"; the version guard needs its
   documented override), but between my failed run and the re-run the tree CLAIMS a
   version it has not shipped — the stale-prod-watch and any version-reading probe will
   see 2.15.0 in source while production runs 2.14.0.

## e) WHAT WE SHOULD IMPROVE

1. **release.sh should eval the tree before step 1** (`nix flake check --no-build` as a
   step-0 precondition): the dry-run's precondition checks cannot catch an eval-red main,
   and step 4 discovering it costs a full buildflow cycle.
2. **Session-start ritual addition**: before ANY long background launch, record load
   (`cat /proc/loadavg`) AND its trend (two samples), plus the target repo's commit
   cadence; refuse or wait when either is hostile.
3. **update-vendor-hash must refuse at entry when vendorHash.nix already holds the
   sentinel fake** — today it probes, learns the hash, and its EXIT trap then restores
   the pre-app state, which IS the fake: a failed run leaves the poison in place and a
   confusing log. Entry-refusal turns that into an actionable message.
4. **The daemon re-poisons pins.** Three occurrences of the sha256-AAAA class now
   (2026-11-02, and at least twice since in pbx-lens via auto-commits). The daemon should
   never auto-commit a file whose content matches a known sentinel (`sha256-AAAA…`,
   `sha256-0000…`); a denylist in the daemon is the cheap fix.
5. **`git ls-files` output was truncated by `head`** and produced a false conclusion
   (serve.go looked untracked). House rule: never pipe inventory commands through head;
   count + grep by exact name.
6. **Contended-repo snapshot discipline**: on any surprising state, take ONE full-state
   snapshot (status + log + diff + worktree list) instead of incremental probes —
   parallel sessions invalidate each read.
7. **CI runner backlog is now a verdict hazard** (1h+ queues): the unwatchable-minimum
   local set is the only timely truth; consider an org runner bump or documented
   "local-minimum stands in for CI" policy for release days.
8. **The CI battery runs a REDUCED matrix** (freshness/audit-all SKIP, predrain SKIP):
   the local ritual is the only full docs-gates coverage — if CI is the gate people
   watch, the reduced matrix should be named in the job, not discovered.
9. **985 commits/8min daemon churn** (TODO_LIST.md alone committed 195×): the auto-commit
   daemon needs a debounce (commit only after an mtime-quiet window); the churn is also
   what makes assert_clean_tree unbeatable during parallel trains.
10. **The release script's step 7 races the daemon** (my telephony commit became
    "ahead by 1" mid-commit). Either the release document tells sessions to pause, or
    step 7 re-checks and rebases its own commit.
11. **A failed release leaves a version-bumped main** — the runbook should name the
    half-release state and its recovery (override + "already bumped" path) in the header,
    so the next person does not re-derive it.

## f) NEXT (ordered, ~36)

1. Wait for quiet (load < 30 trending down AND webphone commit cadence < 5/8min), then
   re-run `WEBPHONE_RELEASE_SKIP_VERSION_GUARD=1 bash scripts/release.sh 2.15.0`.
2. Before the re-run: `nix flake check --no-build` in webphone (eval authority; the
   vm-tests fix has only been proven via attrNames eval + CI jobs so far).
3. Confirm the parallel dupl-repair commit landed and the tree is clean at launch time.
4. Watch release step 4 (buildflow) complete; if nix-build loops again, read the FIRST
   failure, not the loop notice.
5. Tag v2.15.0 pushed → verify ls-remote + the peeled object (the script does this).
6. gh release notes: eyeball the extracted [2.15.0] section (non-empty, rail bullet
   present) before announcing anything.
7. Release step 7 lands telephony webphone→v2.15.0: confirm the commit + push + its CI.
8. Watch telephony CI at 508c789 (lens fix) — queued; if the fe6eb88 run completes red
   on the pre-fix lens pin, annotate it as superseded.
9. pbx-artmann: `nix flake update telephony` (one relock = lens fix + webphone tag),
   `nix run .#lock-drift-probe` (require webphone pin MATCH + ALREADY-LATEST).
10. pbx-artmann: restage ONLY over a settled tree (parallel session landed its refactor;
    re-verify `git status` immediately before staging), then `deploy-freshness` FRESH.
11. Write the CHANGELOG baseline block at the file TAIL (placement law), naming the
    staged root read in the same breath as the write.
12. Handover: the exact owner switch command + paste-backs (unit states for
    webphone.service, /version answer, one rail screenshot ask).
13. Include the rollback ladder in the handover: previous closure path + the exact
    `nixos-rebuild switch --rollback` invocation.
14. Post-switch: `nix run .#ui-track -- --base-url https://pbx.artmann.tech --track`;
    compare against the pre-rail captures.
15. Post-switch: verify-live webphone arms + the deploy ledger + FEATURES
    deployed-closure line stamp.
16. Owner-gated: §g Q2 — mobile bottom bar, all-8 slots vs 5+More.
17. Owner-gated: §g Q3 — T2 (per-page layouts) vs T3 (presence workspace) ordering.
18. After the parallel lens session lands serve→internal/serve + CI green: relock
    telephony to it (next train, not this one).
19. Fix update-vendor-hash: entry-refusal on the sentinel fake (§e-3) + a selftest pin.
20. Propose the daemon sentinel denylist (§e-4) upstream / to the owner.
21. Add release.sh step-0 eval precondition (§e-1).
22. Delete the local `release-line` branch post-merge (avoid the next describe surprise).
23. Verify the [2.15.0] CHANGELOG section actually contains the sidebar-rail Added bullet
    (the original handover's feature) — if it was folded away, restore it before the
    release notes are cut.
24. The island node suite's 2 base failures (drafts pill state) — schedule the repair
    train or record the acceptance.
25. Verify the queued webphone CI runs' verdicts when they complete; heal any new red
    same-day.
26. Record the docs-gates v0.3.1 upgrade in pbx-artmann's CHANGELOG (or pin back to the
    tag form per the input ritual), and run the full parity diff against the frozen
    oracle per the documented upgrade bar.
27. Webphone docs/status INDEX.md row for this report (status-index.py).
28. Harvest check: rail report §f rows 47–50 (T2 scheduling, owner review, rollback note,
    close-out) — confirm each carries a routing pointer.
29. CHANGELOG "Fixed" entries in webphone for a-2/a-4/a-3 (eval-red main, lens hash,
    tag ancestry) — they are train-relevant repairs currently only in git history.
30. Investigate the 985-commits/8min churn loop identity (which daemon/sessions) and
    file the debounce proposal (§e-9).
31. Decide the release-window convention (§e-10) before the next release: sessions pause
    commits or the release locks the tree.
32. stale-prod-watch: confirm it tolerates the source-version 2.15.0 vs prod 2.14.0
    window (it should WARN, not page, until the switch).
33. After the switch: webphone-smoke --base (the release script's printed manual step).
34. Verify pbx-lens CI on 0b83c4f goes green (its own suites beyond the package build).
35. pbx-artmann: after the relock lands, one full docs battery as the session-closing
    ritual (the 584→599 edits + report landing ride it).
36. Keep the two `/tmp` verify worktrees and scratch dirs trashed (lens-verify removed;
    battery/dupl logs can go with the next sweep).

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Cut v2.15.0 as ONE release covering every train folded into the CHANGELOG**
   (sidebar rail, pbx-lens deep links, notes-in-search, DLQ observability,
   export-legacy, the hardening spine…), or do you want the release held until the
   currently in-flight parallel trains (whole-list, T25 owner-pack) land and ride in
   the same tag?
2. **The deploy window**: the switch is yours to run (`nixos-rebuild switch --flake .#pbx
   --target-host root@pbx.artmann.tech`). Should I prepare the staging root and handover
   as soon as the release lands, even while other sessions keep landing pbx-artmann
   trains, or do you want the deploy paired to a specific time?
3. **§g Q2 from the rail report** (which I retracted the fabricated answer for): on
   phones, all eight destinations in the bottom bar with truncated labels, or five
   primaries plus a "More" sheet?
