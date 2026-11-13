# Closeout round 2 — the self-review: what I forgot, what I did worse than I knew, what still to improve

> ARCHIVED 2026-11-11 (docs-health cadence sweep): gate-clean archive candidate (check-rows --strict + bare-rows clean) — the two-oldest CLEAN pair archived to bring the live count back under the 65 cap after the 13-54 report filed.

> **Date:** 2026-11-09 07:53 CET (verified via `date`)
> **Session:** the execute mandate that resumed the 03-40 honest report
> ("READ, UNDERSTAND… Execute and Verify… Keep going until everything
> works") — one continuous run 03:45–07:53 covering the red-main heal,
> the v2.12.0 completion chain, the enroll split-brain fix, two stack
> test modernizations, and the A46/A47 leftovers.
> **Snapshot rot:** claims verified as of 07:53; every CI verdict below
> is workflowName-verified (see d-1 for why that qualifier exists).

## a) FULLY DONE (verified this session)

1. **Main red diagnosed from the CI logs, not the stale theory.** The
   03-40 report's "docs ratchet crossed" reading was half-wrong: the
   battery itself PASSED (annotate exactly AT pin 203); the reds were
   dprint format (my report's wrap points — a line starting `+ branch`
   is parsed as a bullet and would be rewritten to `- branch`,
   meaning damage; reworded, never reflowed), freshness (one §f row
   whose every token is DF-generic; `fresh:skip` with the
   push-execution reason), and strike-safety on a 6-cell FEATURES row
   (already healed by the sibling's `4788bc52`).
2. **art-dupl healed by the book**: all 6 new groups adjudicated BEFORE
   the baseline write (registry row + sweep-log: the three readmodels
   fax twins — the store is prod-dead since P7 and survives only as
   the fax service's test double, a TODO row owns its deletion which
   dissolves them; the ProjectionHostOptions registration idiom; the
   sweep trivia). Baseline 131 → 135 with the CI-pinned v0.7.2,
   set-diff exactly +6 adjudicated / −2 A43-deleted importers,
   post-write `check` = 0 new.
3. **release.sh lychee quoting fixed and SCRIPT-verified** (the scar's
   own demanded ritual): the script's exact extracted line against a
   fixture — with the pattern the private links are Excluded and
   rc=0; without it pbx-artmann 404s and rc=2.
4. **THE ENROLL SPLIT-BRAIN — found, fixed, triple-proven.** The
   stack's enroll-verify VM red was never a go-cqrs-lite gap:
   webphone's `-enroll-passkey` CLI still minted through the retired
   readmodels engine while the server — since P7 — verifies from
   system.db. Cross-process mint/verify split-brain shipped in every
   2.1x since the engine's retirement. Fix (`internal/cqrs/enroll.go`
   - `cmd/webphone/main.go`): the CLI boots the minimal enroll-only
     journal system over the SAME system.db (`cqrs.NewEnrollSystem`) and
     `Mint` awaits its own projection fold (the D7 mechanism) before
     exit. Proven by (i) `TestEnrollCrossProcessMintVisibleToRunningServer`
     (server idling, CLI mints + closes, server sees it), (ii) local
     end-to-end: fresh verify 200, burned 503, no readmodels.db, (iii)
     the stack's enroll-verify VM check green on the relocked pin.
5. **v2.12.0 COMPLETED** — tag re-pointed onto `efb0ff1a` (no gh
   release, no consumer had the old tag; the owner §g-1 question was
   answered by the mandate with the rationale recorded), lychee green
   twice, stack relocked `3da66985 → f5062592 → efb0ff1a` (both moves
   recorded in its CHANGELOG; lock-move-guard green), stack gates
   green with rc files (never pipe tails), aarch64 rebuilt + ELF b700
   guard PASS, `gh release create v2.12.0` published 04:20Z as
   Latest with the 36-line notes carrying the unification + the fix.
6. **Two stack VM tests modernized to the journal world** — their
   legacy-SQL assertions had been red since the 2.10/2.11 relocks,
   hidden by pipe-swallowed verdicts: `fax-feed` now reads system.db's
   `meta_planned_faxes`; `messaging` dropped raw-SQL SEEDS for real
   sends through webphone's own `/messages/send` lane (session + CSRF
   rotation in-test) — the full send path now rides the VM test,
   STRONGER than the seeds it replaced. Six fix commits; the stack's
   full flake check rc=0; stack CI green on `a44010d`.
7. **A46 leftovers closed**: fleet-libs `go-cqrs-lite.md` re-scored
   **96/100** on a 14-row frame (11 direct modules; the EnforceCap
   NO-GO superseded by WithCapPerGroup; audit-verify 8/8 on the
   rewritten report), ROADMAP carries the ES doctrine paragraph,
   index.html + AGENTS score lines updated (mean 80).
8. **audit-verify full replay repaired — 91/91** (five drifted
   count-claims fixed in their reports with dated evidence; the
   branded-id probe's pattern scoped past stdlib `cmp.Or` collisions).
9. **`nix flake check --all-systems` PASSED** (session-first full run
   incl. KVM backup VM + aarch64 checks).
10. Smaller §f closes: worktree cleanup (`/tmp/wp-p6` + branch;
    pruned), buildflow timings reviewed (all faster or
    version-crossed), `.docs-gates.json` pins verified un-drifted,
    cqrs-lint triage TODO row (185 warnings; **C007 is ×3, not 180** —
    the 03-40 miscount corrected), runbook lessons added
    (worktree-release-interleave; verdicts-ride-rc-files).
11. **End state, workflowName-verified**: webphone main green at
    `fe31b3e1` (CI runs 42562946635 + 42563028873 success), stack main
    green at `a44010d` (CI run 42561260197 success), the v2.12.0
    release live.

## b) PARTIALLY DONE

1. **The docs-gates-in-pre-commit decision** (03-40 §f-25): decided
   DECLINED (commit-time cost + DF-brittleness noise; the
   battery-on-commit rule + CI's fast-flake leg already cover it) —
   but the decision record lives only in my closing message, not in a
   one-home doc. This report's §f-9 homs it.
2. **The cqrs-lint advisory debt**: counted, corrected (C007 ×3),
   TODO row added — the actual triage (adjudicate / file upstream /
   baseline) is untouched.
3. **The enroll-fix commits' CI story**: healed (`4bbe6f48`), but see
   d-1/d-2 — the healing landed AFTER the release published, and the
   tag still points at the commit whose CI run carried the reds
   (repo-side checks only; the shipped artifact is byte-stable across
   the healing commits).

## c) NOT STARTED

1. pbx-artmann relock (the owner's standing §g-2; nobody touched it).
2. The readmodels faxes store deletion (TODO row; needs the fax test
   doubles moved to the cqrs backend).
3. The legacy-table DROP migration (operator-gated; retirement doc §5).
4. The registry-suppressions sweep and C007/C044 adjudication (ride
   the cqrs-lint triage row).

## d) TOTALLY FUCKED UP (the honest ledger — this is the section the owner asked for)

1. **I misread CI green TWICE and published a release on the false
   belief.** `gh run list` without `workflowName` shows the WATCH
   workflow's scheduled runs — my jq filters showed
   `completed/success` rows on the enroll-fix shas and I declared
   them green, re-pointed the tag onto `efb0ff1a`, and created the gh
   release. The CI-workflow runs on `85ba79d2` and `efb0ff1a` were
   FAILURE (registry + baseline legs). The reds surfaced only on the
   NEXT commit's run. Even my `gh run watch --exit-status` on
   `85ba79d2` printed rc=0 for a failed run and I did not question
   it. The 2026-10-05 standing rule ("check CI after EVERY push") was
   followed in MOTION but defeated in INSTRUMENT — I checked the
   wrong workflow's verdict and never noticed.
2. **The enroll fix shipped WITHOUT its registry line** — the exact
   2026-10-24 lesson printed in AGENTS.md ("new errorfamily codes
   register IN THE SAME COMMIT"): `cqrs.enroll_boot` landed
   unregistered because I ran the cqrs suite but never the arch
   package's freshness test before pushing.
3. **The baseline was not re-recorded in the same commit as the
   line-moving fix** — the very ride-along rule I had re-asserted in
   the registry sweep-log HOURS earlier (the e-5 scar). My fix's line
   insertions re-keyed the Phase-3-ruled enroll twin; CI redded on a
   re-key, not new duplication — but the gate cannot know that, and I
   made it find out the loud way.
4. **The release published before its tag's CI verdict was truly
   verified.** Local gates were green and the ritual's stack legs
   were done, but the publish rode the d-1 misread. The artifact
   itself is fine (the healing commits changed only repo-side
   checks + docs); the SEQUENCE was wrong and the tag now sits on a
   CI-red commit by the letter of the run record.
5. **The first local E2E was contaminated** — my kill of the leftover
   test server targeted the wrong PID through a subshell variable
   mixup; a stale server on the port answered my probes and nearly
   read as a route regression (404). Clean rerun on a fresh port
   produced the proof.
6. **The messaging-test rewrite took FOUR rounds** — wrong lane first
   (`/gateway/message` is the bridge's provider target, not a webphone
   surface; the row never existed), then 201-vs-200 on login, then
   `/messages` vs `/messages/send`, then the cookie `-c` + CSRF
   re-fetch after login rotation. Every round was caught by the VM
   check itself (the system working), but I burned four ~8-minute
   cycles reading interfaces I could have read first.
7. **Two daemon commit races lost** (the red-fix split 2+2 files; the
   A46 batch rode two `chore:` commits). Nothing was lost — verified
   by `git log --stat` after each — but the narrative history is
   scattered again, and I knew the one-call rule going in; I lost by
   editing-and-verifying with uncommitted bytes sitting in the tree.
8. **Smaller self-inflicted cuts**: a stray `>` typo in index.html
   (caught at once); a §d numbering botch in the first report draft
   (fixed); a first big-block multiedit on the stack's messaging.nix
   failed wholesale on whitespace and I should have started with
   small chunks; the first ELF-guard one-shot redirected its evidence
   into an empty file (`2>/dev/null` ate it) and had to be re-run
   foreground.

## e) WHAT WE SHOULD IMPROVE (process, from this session's scars)

1. **CI verdicts are only verdicts with `workflowName`.** The watch
   workflow's scheduled runs masquerade as CI verdicts in a name-less
   run list. Standing correction for every future session: cite the
   CI-WORKFLOW run id, or filter `--workflow CI`. Better: a tiny
   `scripts/ci-verdict.sh <sha>` that prints exactly that, so the
   check is mechanical, not a jq discipline. And question any
   `gh run watch` rc that disagrees with the run's conclusion.
2. **The registry and baseline lessons are now THREE-time repeaters**
   (2026-10-24, 2026-11-08 e-5, tonight). Prose in AGENTS.md is not
   catching them at commit time — they need a mechanical pre-push
   probe (a battery step: "new errorfamily codes? regen + diff" and
   "art-dupl check" for ANY code train, not just refactor trains).
3. **Never publish a release from a tag whose CI-workflow verdict is
   not personally verified green** — local gates + stack legs do not
   substitute for the tagged commit's own CI record.
4. **Read the interface before driving the VM** — the four-round
   messaging rewrite was four rounds because I probed instead of
   reading the route table and the session/CSRF contract I already
   had in AGENTS.md.
5. **Kill-by-PID discipline**: verify the kill landed (pgrep again)
   before trusting any subsequent probe against that port.

## f) NEXT TASKS (ledger — harvest inventory, not commitments)

<!-- check-rows:ledger -->

| #  | Task                                                                                                                                          | Impact | Effort | Category |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 1  | pbx-artmann relock to the stack's new pin (owner g-1: which session owns that tree)                                                           | High   | S      | owner    |
| 2  | Ratify D2 + D4-amendment + the v2.12.0 tag re-point, incl. the tag-on-CI-red-record acceptance or a final re-point (g-2)                      | High   | S      | owner    |
| 3  | Release-policy ruling: may a gh release publish after LOCAL gates + stack push, before the stack's REMOTE CI verdict (g-3)                    | High   | S      | owner    |
| 4  | `scripts/ci-verdict.sh <sha>` — print the CI-workflow conclusion for a sha (the d-1 instrument fix)                                           | High   | S      | webphone |
| 5  | Mechanical pre-push probes: errorfamily-code freshness + art-dupl check on EVERY code train (the e-2 fix)                                     | High   | M      | webphone |
| 6  | Smoke scenario for `-enroll-passkey` (mint → verify 200 → burned 503): webphone's own CI would have caught d-2 pre-release                    | High   | M      | webphone |
| 7  | Delete the prod-dead readmodels faxes store (TODO row; dissolves 3 baselined clone groups)                                                    | Med    | M      | webphone |
| 8  | Check whether `readmodels.EnrollTokens` is now prod-dead too (the CLI moved to the journal; only userauth tests may hold it)                  | Med    | S      | webphone |
| 9  | cqrs-lint 185-warning triage (TODO row; C007 ×3, C044 ×50 — adjudicate, file, or baseline)                                                    | Med    | M      | webphone |
| 10 | The legacy-table DROP migration once the operator confirms empty tables (retirement doc §5)                                                   | Med    | S      | owner    |
| 11 | Give the WATCH workflow a distinct displayTitle prefix so name-less run lists cannot masquerade as CI verdicts                                | Med    | S      | webphone |
| 12 | release.sh resume-from-step flag (03-40 §e-4; two trains died at 6/9 in two nights)                                                           | Med    | M      | webphone |
| 13 | AGENTS.md "Concurrent sessions"/commands sections gain the workflowName-verified-verdict rule                                                 | Med    | S      | webphone |
| 14 | The committed benchmarks ride a docs note or CI lane (03-40 §f-15, carried)                                                                   | Low    | S      | webphone |
| 15 | `facesDrive` test helper inline (03-40 §f-16, carried)                                                                                        | Low    | S      | webphone |
| 16 | Stack's on-demand Browser-E2E CI lane ran 0s (skipped) on the relock push — confirm the on-demand trigger policy                              | Low    | S      | stack    |
| 17 | The messaging VM test logs in through the fail-open directory-verify — consider a dedicated machine-test credential lane                      | Low    | M      | stack    |
| 18 | Registry-suppressions sweep (03-30 report §f, carried)                                                                                        | Low    | M      | webphone |
| 19 | `fresh:skip` semantics: consider upstreaming a self-executing skip class (push-watch rows are structurally unharvestable)                     | Low    | S      | docs     |
| 20 | AGENTS.md enroll/one-home paragraphs still describe the CLI as opening the readmodels engine — the doctrine docs need the journal-world sweep | Med    | S      | webphone |
| 21 | The 2.10.0 gh-release notes trim guidance for future catch-up releases (03-40 §f-26, carried)                                                 | Low    | S      | webphone |
| 22 | Datastar unsafe-eval→nonce retirement lane (TODO, carried)                                                                                    | Low    | M      | webphone |

## g) QUESTIONS (cannot be figured out from the repos)

1. **pbx-artmann** (f-1, standing from two reports back): should the
   next session drive its relock to the stack's new pin now, or does
   the sibling stack/android session own that tree? I will not
   interleave a third repo into a checkout another session may be
   mutating without your call.
2. **Ratify or unwind** (f-2): D2 (sessions stay SQL), the D4-amendment
   (releases cut at ship-visible boundaries), and the v2.12.0 tag
   re-point — now with the addendum that the tag sits on `efb0ff1a`,
   whose CI-workflow run was red on the registry/baseline legs
   (healed in `4bbe6f48`, one commit later; the shipped artifact is
   byte-stable — the healing changed only repo-side checks and docs).
   Accept as documented, or re-point the tag once more onto a fully
   green commit?
3. **Release-policy ruling** (f-3): tonight's sequence published the
   gh release after ALL local gates + the stack relock/push, but
   while the stack's REMOTE CI was still running (it went green
   later). Is local-green + pushed the bar, or must the remote CI
   verdict gate the publish? The runbook is ambiguous exactly here,
   and both readings defensible — your call sets the standing rule.

— END OF REPORT. WAITING FOR INSTRUCTIONS. —
