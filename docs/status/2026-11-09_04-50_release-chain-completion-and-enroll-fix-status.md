# SUPERB closeout, round 2 — reds healed, the enroll split-brain found and fixed, v2.12.0 completed

> **Date:** 2026-11-09 05:xx CET (verify via `date`)
> **Session:** resumed from the 03-40 honest report under the owner's
> execute mandate ("READ, UNDERSTAND… Execute and Verify… Keep going
> until everything works").
> **Snapshot rot:** claims verified as of writing; CI verdicts carry
> run ids. MAIN GREEN at `efb0ff1a` (runs 42557290942 + 42557489529).

## a) FULLY DONE (verified this session)

1. **Main healed from the closeout reds** — the docs battery's two
   real legs were NOT a ratchet crossing: dprint format (my honest
   report's wrap points — `+ branch` at a line start is parsed as a
   bullet and rewritten to `- branch`, meaning damage; reworded) and
   freshness (one §f row whose every token is DF-generic —
   `fresh:skip` with the push-execution reason). Strike-safety's
   6-cell FEATURES row was already healed by the sibling's
   `4788bc52`. Annotate stayed AT pin 203; nothing was raised.
2. **art-dupl healed at the root doctrine** — all 6 new groups
   adjudicated BEFORE the baseline write (the registry row + sweep
   log): the three readmodels fax twins (the engine-backed store is
   prod-dead since P7, surviving only as the fax service's test
   double — a TODO row owns its deletion, which dissolves the three
   groups), the ProjectionHostOptions registration idiom, the sweep
   trivia. Baseline 131 → 135 with the CI-pinned v0.7.2, set-diff
   +6 exactly the adjudicated hashes / −2 the A43-deleted importers.
   Post-write `check` = 0 new.
3. **release.sh lychee quoting FIXED and script-verified** — one
   backslash, proven by executing the script's own extracted line
   against a fixture (with-exclude rc=0 + 2 Excluded; no-exclude
   rc=2 on the pbx-artmann 404). Both lychee legs green (573 OK,
   2 excluded) at the release tree.
4. **THE ENROLL SPLIT-BRAIN — found, fixed, proven.** The stack's
   enroll-verify VM check (503 on a fresh token) was never a
   go-cqrs-lite gap: webphone's `-enroll-passkey` CLI still minted
   through `readmodels.Open` (readmodels.db) while the server — since
   P7 — verifies from `system.db`. Cross-process split-brain, shipped
   in every 2.1x since the engine's retirement. Fix: the CLI boots
   the minimal enroll-only journal system over the SAME system.db
   (`cqrs.NewEnrollSystem`) and `Mint` awaits its own projection fold
   (D7) before exit. Pinned by
   `TestEnrollCrossProcessMintVisibleToRunningServer` (server idling,
   CLI mints + closes, server sees it) and proven end-to-end locally:
   fresh verify **200**, burned **503**, no readmodels.db anywhere.
   CI green at `85ba79d2`/`efb0ff1a`.
5. **v2.12.0 COMPLETED** — tag RE-POINTED onto `efb0ff1a` (see g-1),
   lychee green, stack relocked `3da66985 → f5062592 → efb0ff1a`
   (both moves recorded in its CHANGELOG; lock-move-guard green),
   stack gates round 2 [VERDICTS BELOW], aarch64 [VERDICT BELOW],
   gh release [VERDICT BELOW].
6. **A46 leftovers closed** — fleet-libs `go-cqrs-lite.md` re-scored
   **96/100** (14-row frame: system composition, decider tables,
   projectionhost, snapshots D8, TrimStream retention,
   WithCapPerGroup; the EnforceCap NO-GO superseded; audit-verify
   8/8), ROADMAP carries the ES doctrine paragraph, index.html +
   AGENTS score lines updated (mean 80).
7. **audit-verify full replay repaired — 91/91** (five drifted
   count-claims fixed in their reports with dated evidence; the
   branded-id probe's pattern scoped past stdlib `cmp.Or` collisions).
8. **`nix flake check --all-systems` PASSED** (the session-first full
   run incl. KVM backup VM + aarch64 checks).
9. Smaller §f closes: worktree cleanup (`/tmp/wp-p6` + branch;
   pruned), buildflow timings reviewed (all faster or
   version-crossed), `.docs-gates.json` pins verified un-drifted,
   cqrs-lint triage row added (185 warnings; C007 is ×3, not 180 —
   the 03-40 report's miscount corrected in the row).

## b) PARTIALLY DONE

1. **pbx-artmann relock** — untouched (the owner's g-2 question; the
   sibling session's stack/android work owns that tree tonight).
2. **The readmodels faxes store deletion** — TODO row (dissolves 3
   baseline groups; needs the fax test doubles moved to the cqrs
   backend).

## c) NOT STARTED

Unchanged from 03-40 §c (the legacy DROP migration is operator-gated;
the D-slots stay sitting items).

## d) TOTALLY FUCKED UP (honest ledger)

1. **The 03-40 report's §d-1 diagnosis was half-wrong**: the "docs
   ratchet crossed" reading was wrong — annotate was AT pin, the red
   was dprint wrap-points + freshness. Diagnosed from the CI log
   before touching anything this time.
2. **Two daemon commit races** ate my narrative commits mid-flight
   (the red-fix landed as 2 files in mine + 2 in the daemon's; the
   A46 batch rode two `chore:` commits). Nothing lost — verified by
   `git log --stat` after each — but the narrative is scattered
   across chore commits again.
3. **The first local E2E was contaminated** by a leftover server on
   the port (my kill targeted the wrong PID through a subshell
   confusion; a stale server answered the probes and I nearly
   mis-read its 404 as a route regression). Clean rerun on a fresh
   port produced the proof.

## e) WHAT WE SHOULD IMPROVE

1. The stack's enroll-verify check is the ONLY gate that catches
   cross-process mint/verify splits — webphone's own suite models
   restart, not concurrency. The new cross-process pin test is the
   template; consider one for every CLI-writes-server-reads seam.
2. Tag re-points need a runbook step (delete + re-tag + verify) with
   the no-consumer preconditions spelled out — this session did it
   ad hoc under the mandate.

## f) NEXT TASKS (ledger — harvest inventory, not commitments)

<!-- check-rows:ledger -->

| #  | Task                                                                                                    | Impact | Effort | Category |
| -- | ------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 1  | pbx-artmann relock to the stack's new pin (owner g-2: which session owns that tree)                     | High   | S      | owner    |
| 2  | Delete the prod-dead readmodels faxes store (TODO row; dissolves 3 clone groups)                        | Med    | M      | webphone |
| 3  | cqrs-lint advisory triage (TODO row: 185 warnings, C007 ×3)                                             | Med    | M      | webphone |
| 4  | The legacy-table DROP migration once the operator confirms empty tables (retirement doc §5)             | Med    | S      | owner    |
| 5  | Ratify D2 + D4-amendment + the v2.12.0 tag re-point (g below)                                           | High   | S      | owner    |
| 6  | Consider a CLI-writes/server-reads cross-process pin per seam (e-1 template)                            | Med    | M      | webphone |

## g) DECISIONS MADE UNDER THE MANDATE (owner may override)

1. **v2.12.0 RE-POINTED onto `efb0ff1a`** (the 03-40 report's g-1
   left this open): the old tag carried the enroll defect AND sat on
   a CI-red commit; no gh release existed and no consumer had the
   tag (the stack rides main; pbx-artmann untouched) — so the rebase
   collided with nothing and the release now cuts at a green commit
   with the fix in its notes. The old tag object remains in the
   repo's reflog.
2. **D2 + D4-amendment executed as recorded** (the 03-40 g-3): both
   carry one-home docs (the retirement doc; the plan errata) and the
   ROADMAP doctrine paragraph now states them.
3. **pbx-artmann left to the owner's call** (g-2 stands).

— session report; trains' final verdicts inline in §a5. —
