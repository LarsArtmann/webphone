# Red-main repair session: clone gate, flake hardening, PR #234 triage

> **Claims verified as of 21:56 CET, 2026-11-13.** Push/CI statements rot fastest; this
> is a point-in-time snapshot.

**Scope line.** VERIFIED here: the CI failure logs for runs 43487357527, 43493935137,
43495245945 (read back via `gh run view --log-failed`); the art-dupl baseline re-record
(pinned v0.7.2 `check` = 0 new after write, output quoted in the session); local race
suites on the touched packages (`go test -race -count=1` green for push, hooks, domain,
cqrs, views, messaging); `scripts/docs-gates.sh --check` PASS after each markdown
landing; CI verdicts read workflow-name-verified through `scripts/ci-verdict.sh` for
`0bb68384`, `35966b33`, `b0f322ee` (all GREEN, run ids in the sections below); issue
#246/#252 close confirmations and the PR #234 comment URL from `gh`. NOT verified here:
the stack browser E2E status for the configjs quartet's served-markup change (owned by
the stack repo; issue #252's filing session claimed a VM run — I did not re-verify that
claim against the stack's run records); the render-byte-parity of `crmMutedItem` beyond
the view tests (no dump-harness diff was run; see §b-1); the two unpushed daemon commits
`a4805106`/`d1706028` (another session's TODO bookkeeping — content read only from
commit subjects); whether the go-idempotency injectable-clock upstream ask (named in a
2026-11-07 ledger row) was ever filed. NOT attempted: the fleet-wide dependabot queue
(out of scope), issue #9 (owner-gated).

## a) FULLY DONE

1. **Red-main clone-gate repair** (CI run 43487357527, red at `dc8e0c89`): diagnosed
   from the failed-job log — two new clone groups from the `7349b2c0` digit-keyed-upsert
   train. Fixed by extraction, not suppression: `domain.keepRunes`/`domain.DigitsOf`
   (internal/domain/ids.go) is the ONE home of the digits-only rune reduction, consumed
   by `contactKeyOf` (internal/cqrs/contacts.go), `sanitizeDialable` (via `isDialable`),
   and the `avatarFor` digit tail (internal/web/views/helpers.go — a THIRD hand-kept
   copy the detector had never flagged, found and folded in the same pass). The crm
   strip's two identical muted spans collapsed onto one `crmMutedItem` templ component.
   Evidence: baseline re-recorded 110 → 109 with the CI-pinned v0.7.2 at the gate's
   exact flags, post-write `check` = 0 new; targeted + race suites green; CI GREEN on
   `0bb68384` (verified 20:39). Registry sweep-log row 2026-11-13 20:3x.
2. **Issue #246 closed** with the full adjudication (what was extracted, what was
   accepted under which standing ruling, why merging the DOM-ordered sites would be
   wrong), plus the note that the issue's original `auth.go`↔`internal_directory.go`
   pair had already dissolved with the platform carve.
3. **Issue #252 closed** with evidence: the configjs quartet landed via
   `7349b2c0`/`0142643e`; at `dc8e0c89` the ONLY failing job was art-dupl (format,
   island, lints all green on the landed quartet); the full local suite's only failure
   was the (then-unattributed) messaging load-flake, green in isolation.
4. **Ring-hook CI flake: root-caused and hardened** (CI run 43495245945,
   `TestHookRingValidatesAndStaysInert`): the ring hook's detached push fan outlived
   the test and its sqlite registry read raced `t.TempDir()`'s RemoveAll ("directory
   not empty" during cleanup; the run's own `push: fan-out aborted … sql: database is
   closed` WARN names the goroutine alive at teardown). Rerun of the same run GREEN
   confirmed the flake. Hardening landed per the ledger's "hardening supersedes
   confirmation" precedent: `push.Fanout.Settle(ctx)` (WaitGroup over detached fans),
   the ring fixture settles in `t.Cleanup` (LIFO orders settle before TempDir removal),
   pinned by `TestFanSettleAwaitsTheDetachedFan` (settle-then-assert without polling).
   Ledger row CLOSED with the mechanism; race-green on push + hooks; CI GREEN on
   `35966b33` (verified 21:10).
5. **PR #234 (dependabot ginkgo bump) triaged to its true verdict**: resolved the
   conflict in an isolated worktree (merge main, tidy, vendorHash verified by
   `nix build .#webphone.goModules --rebuild` + full `nix build .#webphone`), pushed
   `233101d5`, dispatched CI (43495381486) — then discovered main had ALREADY carried
   ginkgo v2.34.0 since `be84a883` (2026-11-11): the bump work was moot, dependabot
   closed the PR itself, and the closing verdict was posted for the record
   (comment 6863547507). Worktree + local branch cleaned up. nextd item marked done.
6. **Docs battery red repaired** (CI run 43493935137, `format rc=20` on the TODO-strike
   push `b6d57e83`): the strike was treefmt-clean but not dprint-clean. Repaired with
   the battery's own fmt step, `--check`-certified locally before the push (`4a1124b0`).
7. **AGENTS.md memory updated** (`b0f322ee`, CI GREEN verified 21:12): (1)
   `ci-verdict.sh` answers exit 2 on a PENDING run — poll loops must wait for the
   verdict to stop saying pending; (2) the CI docs-gates battery owns markdown bytes
   via dprint even though treefmt does not — run `docs-gates.sh --check` before
   pushing md edits.
8. **TODO_LIST stale row struck with evidence**: the "AGENTS.md one-liner" row was
   already satisfied (AGENTS.md carried the MOVES/RENAMES regen fact since the asr-split
   train); struck per the strike canon, committed `b6d57e83`.
9. **CHANGELOG entries** for both code-path trains (the dedup one-home; the Settle
   hardening) under `[Unreleased] → Changed`.

## b) PARTIALLY DONE

1. **`crmMutedItem` render parity — tested, not render-diff-proven.** The view tests
   (crm_context_test.go) pin content and pass, but the 2026-11-10 protocol used a
   temporary render-dump harness to prove BYTE-identity for a templ refactor; I relied
   on the tests + a token-level argument instead. Remaining open: run the dump-harness
   diff or write the ruling that view-test coverage is sufficient for span-level
   refactors. Effort: S. Blocker: none.
2. **The detached-fan hardening covers ONE seam.** `Fanout.Settle` guards the ring
   fixture. The OTHER detached-goroutine sites (the review-draft job at
   internal/server/api/review_api.go:89 with its `context.WithoutCancel` + 15s budget,
   and any other test pairing `t.TempDir()` with a goroutine that touches files after
   test end) have the same latent race un-audited. Remaining open: a sweep of
   TempDir+detached-goroutine pairs; extend Settle to the draft-job fixture if the
   pattern matches. Effort: M. Blocker: none.
3. **The clone acceptance is baselined, not eliminated.** The residual crm templ
   if-shape pair rides the standing ACCEPT ruling; it is a tracked baseline entry, not
   gone. Reopen trigger recorded in the registry row (a third muted line joining the
   strip forces re-extraction). Effort: 0 (deliberate). Blocker: none.
4. **`Settle`'s timeout path is untested.** The pin covers the happy path
   (settle-then-assert); the `ctx.Done()` branch returning `ctx.Err()` has no test.
   Effort: S. Blocker: none.
5. **Queue triage depth**: only the three webphone items were adjudicated (#234 done,
   #236 already closed, #9 owner-gated). The fleet queue's 259 external-human P0 items
   and the parked set were untouched — by scope decision, but they remain the
   ranking's top of the pile. Effort: ongoing. Blocker: session scope.

## c) NOT STARTED (observed during the session, deliberately untouched)

1. **OIDC lane residuals** (NixOS VM test for the LoadCredential lane; second
   module-check golden with `oauth2.enable`; typed passkey options) — TODO rows exist;
   waiting on a train slot. Priority: per TODO ordering.
2. **Schema-export spike** (settings schema generated from config structs) + the
   durable CUE ruling record — TODO rows; not started.
3. **Stack drift audit** (stack repo's webphone module assertions vs internal/config)
   — rides the next stack module touch; not started here.
4. **go-appkit/flightrecorderhealth adoption** (the first module-level candidate,
   demand-gated TODO row) — not started.
5. **Issue #9 (Kith CRM switch)** — owner-gated (pki trust + bearer provisioning);
   untouched.
6. **Voicemail prod legs** (stack redeploy riding main; host-side diagnostics; stack
   coverage arms) — untouched; deployment-gated.
7. **Prod capture train** (T05 authenticated capture, night-run stability, ring-guard
   positive path) — untouched; owner-cred gated.
8. **BuildFlow bandit-provider issue filing** — drafted; needs verify-before-filing;
   not filed.
9. **Hub runner durable fix + ALERT_WEBHOOK_URL decision + ringhook operator docs** —
   untouched.
10. **The concurrent session's new TODO rows** (`a4805106`: dead readers deletion,
    export tooling) — landed by the other session while this session ran; not read,
    not adjudicated here.

## d) TOTALLY FUCKED UP

1. **I did the PR #234 merge resolution BEFORE verifying the bump was still needed.**
   One `rg ginkgo go.mod` on main at minute zero would have shown v2.34.0 landed two
   days earlier (`be84a883`) and ended the task immediately. Instead I spent ~20
   minutes on fetch/worktree/merge/conflict-resolution/tidy/vendorHash/full-build/push/
   dispatch before the discovery. Severity: wasted effort, no damage (the branch CI ran
   green; dependabot closed the PR). Root cause: I anchored on the queue item's
   framing ("CONFLICTING, needs the tail") instead of verifying the premise first.
   Mitigation: the closing verdict documents reality; the pre-check is now §f-5.
2. **I pushed a markdown edit without running the CI-shaped docs battery**, and reded
   main for ~25 minutes (run 43493935137, `format rc=20`): the TODO strike was
   treefmt-clean, `nix fmt` + `format-check.sh` green, but the CI battery's dprint
   gate reds on the bytes. Severity: red main window; zero user impact (no deploy
   rides CI red directly). Root cause: treefmt-unowned ≠ battery-unowned — a
   distinction the docs-health docs carry that I did not consult before pushing.
   Mitigation: the AGENTS gotcha now records it; the fix push was --check-certified.
3. **I misread `ci-verdict.sh`'s pending contract twice.** Pending answers exit 2
   ("RED"), so my first poll loop broke instantly on a pending run, and I briefly
   reported a pending run as a red in-stream. Severity: minutes of confusion, a killed
   and rewritten background job. Root cause: I wrote the loop from the exit-code table
   without reading the script's verdict function. Mitigation: AGENTS gotcha recorded;
   the corrected loop greps for the literal `pending`.
4. **I staged a merge with `git add -A` and committed 846 files without listing the
   staged set.** It was correct (the merge delta) and the whitespace gate adjudicated
   it, but the habit is a loaded gun: one wrong resolution would have shipped unseen.
   Severity: process risk only. Mitigation: none recorded yet — see §e-5.
5. **I ran the full test suite concurrently with art-dupl and templ regen**, creating
   the exact load that expired `TestConcurrentFirstSightYieldsOneThread`'s 5s await
   deadline and produced a false failure I then had to attribute. The isolation re-run
   rescued the verdict, but the suite ran under conditions I manufactured. Severity:
   misattribution risk. Mitigation: sequence heavy verification; the race suite is the
   CI arbiter anyway.
6. **Minor: printed a meaningless exit-code echo** (`ARTDUPL_EXIT=$?` captured `tail`'s
   exit, not the detector's; PIPESTATUS is unavailable in this shell — a documented
   trap). The verdict came from grepping output, but the printed `=0` read like
   verification and wasn't.

## e) WHAT WE SHOULD IMPROVE

1. **Verify-the-premise-first for every queue item**: before working a dependabot PR,
   `rg <dep> go.mod` on main; before acting on a tracker issue, re-check whether its
   red is still red at HEAD. Both failures this session (#234, #252's premise) were
   stale-premise work.
2. **Push markdown through the CI-shaped battery, not just treefmt**: md edits get
   `scripts/docs-gates.sh --check` before push. (Now in AGENTS; needs the habit.)
3. **Sequence heavy local verification**: never run the full suite while other
   compile/scanning jobs hold the CPU — race deadlines are wall-clock sensitive.
4. **Read a tool's verdict contract before writing loops around it** (the ci-verdict
   exit table is not the whole contract; the verdict TEXT is the state machine).
5. **Never `git add -A` a merge without `git status --short` review of the staged
   list** — stage the resolved files explicitly; let the merge commit carry the rest.
6. **Attribute flake candidates by rerun FIRST** (this session did this right for the
   ring hook — rerun before digging — keep it), and record the ledger row in the same
   hour as the hardening (also done — keep it).
7. **The fix-forward protocol worked**: the configjs quartet (another session's
   in-flight sweep) was tracker-filed, not touched, and healed by its author. Keep.
8. **Adjudicate-before-baseline with set-diff verification** for clone-gate writes
   worked cleanly (110→109, 0 new). Keep; it is the registry protocol.

## f) Top 50 next things (HARVEST fuel — brainstorm, not commitments)

Items below are ranked roughly by impact; Impact/Effort/Category per item. This
section feeds docs-health HARVEST into TODO_LIST/ROADMAP.

| #  | Task                                                                                                                                                                         | Impact | Effort | Category      |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Run the render-dump byte-parity proof (or write the sufficiency ruling) for the `crmMutedItem` refactor                                                                      | Medium | S      | Quality       |
| 2  | Sweep all `t.TempDir()` + detached-goroutine test pairs; extend `Settle` to the review-draft fixture if the pattern matches                                                  | High   | M      | Quality       |
| 3  | Add a `Settle` timeout-path test (ctx expiry returns `ctx.Err()`)                                                                                                            | Low    | S      | Quality       |
| 4  | Cadence drain: archive the oldest eligible status report through the annotate gates (this report pushed the live count to 66 vs pin 65)                                      | High   | M      | Documentation |
| 5  | Queue pre-check for dependabot PRs: `rg <dep> go.mod` on main before any merge work; auto-close already-landed ones with the verdict                                         | High   | S      | Quality       |
| 6  | Re-verify the stack browser E2E obligation for the configjs quartet's served-markup change (claim in #252 unverified from this side)                                         | High   | M      | Quality       |
| 7  | Migrate push_test polling (`wait()`) fixtures to `Settle` where they fan                                                                                                     | Low    | S      | Quality       |
| 8  | Make `TestConcurrentFirstSightYieldsOneThread` load-tolerant (scale the 5s await deadline, TimingScale precedent) if it flakes again                                         | Medium | S      | Bug           |
| 9  | Add `ci-verdict.sh --wait` (poll-until-final) so ad-hoc poll loops stop re-implementing the pending contract                                                                 | Medium | S      | Quality       |
| 10 | Re-measure `.docs-gates.json` ratchets (CI ran `--annotate-max=260`; AGENTS prose says 203 — doc drift spotted in the CI log)                                                | Medium | S      | Documentation |
| 11 | Check whether the go-idempotency injectable-clock upstream ask (ledger row 42407471289) was ever filed; file it if not                                                       | Low    | S      | Cleanup       |
| 12 | OIDC lane NixOS VM test (`webphone-oauth2`-style: LoadCredential lands, mode 0400, app reads it)                                                                             | High   | L      | Feature       |
| 13 | Second module-check golden with `oauth2.enable` (exact rendered JSON)                                                                                                        | Medium | M      | Quality       |
| 14 | Typed passkey options in the NixOS module (derived rp for passkey-only configs)                                                                                              | Medium | M      | Feature       |
| 15 | Settings schema export spike (generated from config structs, the openapi.json doctrine)                                                                                      | Medium | L      | Feature       |
| 16 | Record the CUE ruling durably as an owner decision + park the cuelang cost measurement                                                                                       | Low    | S      | Documentation |
| 17 | Stack drift audit: stack repo webphone module assertions vs `internal/config`                                                                                                | Medium | M      | Quality       |
| 18 | Adopt `appkit/flightrecorderhealth` (demand-gated TODO row, the first go-appkit module candidate)                                                                            | Medium | M      | Feature       |
| 19 | Issue #9: Kith CRM switch (owner: PBX pki trust + bearer token provisioning)                                                                                                 | High   | M      | Feature       |
| 20 | Stack redeploy riding main so the voicemail playback fixes reach prod; click play and record the toast verdict                                                               | High   | S      | Bug           |
| 21 | Host-side voicemail diagnostics (journalctl send_file/audio-not-found grep; fs-state-acl status)                                                                             | Medium | S      | Bug           |
| 22 | Stack coverage arms: browser E2E voicemail playback; verify-live.sh authenticated audio fetch; operator.nix pin for 206-through-proxy                                        | Medium | L      | Quality       |
| 23 | T05 authenticated prod capture (owner creds required)                                                                                                                        | Medium | M      | Quality       |
| 24 | Night-run stability pair for the capture harness (quiet-window cron + prune check)                                                                                           | Low    | M      | Quality       |
| 25 | Ring-guard positive path exercise in the capture harness                                                                                                                     | Low    | S      | Quality       |
| 26 | Verify-before-filing pass, then file the drafted BuildFlow bandit-provider issue                                                                                             | Low    | S      | Documentation |
| 27 | Hub runner durable fix (ExecStartPre provisioning or equivalent)                                                                                                             | Medium | M      | Bug           |
| 28 | Hub ALERT_WEBHOOK_URL decision (the failure log names the missing knob)                                                                                                      | Low    | S      | Cleanup       |
| 29 | Ringhook operator docs + typed mirror in the stack                                                                                                                           | Low    | M      | Documentation |
| 30 | Read + adjudicate the concurrent session's new TODO rows (`a4805106`: dead readers deletion, export tooling)                                                                 | Medium | S      | Cleanup       |
| 31 | Check the daemon's push-leg delivered `a4805106`/`d1706028` (unpushed at session end, minutes old)                                                                           | Low    | S      | Cleanup       |
| 32 | Consider dependabot auto-rebase config to kill the stale-base PR class at the source                                                                                         | Medium | S      | Quality       |
| 33 | Re-open trigger hygiene: the crm muted-span baseline entry wants a re-extract if a third muted line joins the strip                                                          | Low    | S      | Quality       |
| 34 | Sweep the CI-flake ledger for rows whose named fix never landed; convert each to a TODO row or close with evidence                                                           | Medium | M      | Documentation |
| 35 | Give the status-index a `--check` run in the next docs sweep (report landed mid-corpus-growth)                                                                               | Low    | S      | Documentation |
| 36 | Audit remaining local branches for staleness after the pr-234 cleanup                                                                                                        | Low    | S      | Cleanup       |
| 37 | Owner triage of the nextd parked set (4 items) and webphone #9                                                                                                               | Medium | S      | Cleanup       |
| 38 | Fleet-libs reopen-trigger registry sweep (the T25 table — sweep it instead of re-reading 24 reports)                                                                         | Medium | M      | Documentation |
| 39 | Freshness-gate watch: this report's prose may trip the DF-brittle audit/annotate pins on the next CI battery run                                                             | Low    | S      | Documentation |
| 40 | erraudit/branching-flow monthly re-measure is due 2026-12-08 per the watch calendar                                                                                          | Low    | M      | Quality       |
| 41 | Consider a `.art-dupl-baseline.json` age audit (110→109 this session; quarterly audit of the accepted set is doctrinal)                                                      | Low    | S      | Cleanup       |
| 42 | Verify `TestServedPageHoldsTheDomContract` + smoke pins still hold after the crm templ helper extraction (they ran green in the suites; a CI re-confirm rides the next push) | Low    | S      | Quality       |
| 43 | Decide whether `Fanout.Settle` should ALSO be called in prod shutdown (drain fans before process exit) — deliberate ruling, currently test-only                              | Medium | S      | Quality       |
| 44 | Pre-commit hook could run the docs-gates fmt step on staged md (kills the dprint rc=20 class at commit time)                                                                 | Medium | S      | Quality       |
| 45 | Read the two newest pareto status reports (15-14, 17-05) before the next TODO sweep so their §f rows are harvested, not re-derived                                           | Medium | S      | Documentation |
| 46 | The `--annotate-max` ratchet moved 203→260 in CI while AGENTS says 203 — reconcile which is canonical at the next docs-health sweep                                          | Low    | S      | Documentation |
| 47 | Add the queue's "already-satisfied" pre-check as a nextd preflight (item 5 generalized: is the upstream still unsatisfied?)                                                  | Low    | M      | Quality       |
| 48 | Re-run the full race suite on a quiet machine before the next release preflight (today's only suite failure was self-inflicted load)                                         | Medium | S      | Quality       |
| 49 | Check whether dependabot recreates the ginkgo PR against the new main; if it does, close with the same verdict template                                                      | Low    | S      | Cleanup       |
| 50 | Capture-harness `--track` lane: compare.html diff health after today's landing (the prod lane is gitignored; no local run was made)                                          | Low    | S      | Quality       |

## g) Three questions I cannot answer myself

1. **Force-push policy for dependabot branches.** When dependabot resets its branch
   (as it did after my merge pushes), is maintainer `--force-with-lease` to a
   dependabot branch pre-approved for you, or should sessions always route around via
   direct-to-main landing (what I did here)? The global rule requires your approval
   for force pushes; I need the standing policy so the next stale-PR conflict resolves
   in one step instead of twenty minutes.
2. **The configjs quartet's stack browser E2E**: issue #252's filing session wrote
   "the browser-E2E VM run and the smoke battery are unaffected", but I did not (and
   per your instruction, did not) re-verify that claim against the stack repo's run
   records for the LANDED tree. Do you know whether the E2E was re-run after
   `7349b2c0`, or should the next stack ride treat the E2E re-run as owed before
   relock?
3. **Queue automation policy**: for dependabot PRs that turn out already-landed on
   main (today's #234, and plausibly dozens across the 1511-item queue), do you want a
   standing pre-check that closes them automatically with a verdict comment — or do
   you prefer each to be adjudicated by a session so the verdict stays human? This
   decides whether §f-5/§f-47 becomes a script or stays a discipline.
