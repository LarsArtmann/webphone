# SUPERB gate-setup closeout — full status report

> Claims verified as of 2026-11-12 22:04 CET. CI/push claims re-read at
> 22:04 (`git status -sb`, `git log -2`); the CI-green claim was captured at
> ~19:08 against the SAME sha that is still HEAD — so it stands for the
> current tip without re-polling.

## Scope line — what this report verified, and what it did not

**Verified this session (artifact named per claim):**

- The SUPERB plan (`docs/planning/2026-11-12_16-14_SUPERB-gate-setup-hardening.md`, commit `22d3afed`) was executed end-to-end: every micro-task M1–M13 has a named verification (the tables below).
- **Full buildflow certification**: my run reads **75/76** in the state DB; the one failure is `go-mod-vendor-tail` (error_family `infrastructure`) — root-caused by querying `~/.local/state/buildflow/buildflow.db` directly (`resume_steps`, latest non-success row), re-verified **scoped green** (`buildflow -s go-mod-vendor-tail` → `1 success, 0 failed`, exit 0), and independently confirmed healed by a concurrent session's **76/76 = 100%** full run minutes later (read from `buildflow history`).
- **Bandit count closed**: `nix run nixpkgs#python312Packages.bandit` works (1.9.4, matches BuildFlow's minor); the "6th issue" identified as **B603 at `scripts/status-index.py:138`** (B607 fires on the same line; the mangled BuildFlow report log carried only one of the two headers); my comma-form `# nosec B603,B607` was proven silently inert (B603 kept firing) and fixed to the space form; fresh full-tree scan reads **0 issues, 63 nosec'd sites**. All recorded in draft addendum 2.
- **Armed-set counts** appended to the `.buildflow.yml` triage note (cqrs-lint 11, samber-linter 0, io-surface 6, jscpd 16, bandit furniture-only) and cross-checked against a live no-cache scoped run (`"total": 11`).
- **Harvest**: TODO_LIST 6-row subsection + ROADMAP owner-infra section (`(repo: BuildFlow)`) + CHANGELOG backfill for `e897efe8` (grep-proven missing before adding); `scripts/status-index.py` regen + `scripts/docs-gates.sh` **PASS on every gate, twice** (pre- and post-edit).
- **Everything pushed and CI-green**: `git ls-remote` == local HEAD at `7dfcab3e`, daemon chunks `e9b9ccf1`/`2ecd1831` carried the closeout files, and `scripts/ci-verdict.sh` reads **GREEN on `3d9515cc`** — which is STILL HEAD and still == origin/main at 22:04. The green is current, not stale.
- The concurrent deepening session landed its own closeout (`3d9515cc`, "T17–T24 close-out") with a real narrative commit — its tail trains are that session's, not re-dispatched here.

**NOT verified here (and why):**

- The contents of the concurrent session's `2026-11-12_18-53_deepening-tail-t17-t24-closeout.md` and its `raceprobe-flake-tracker` draft edit — foreign session's files, observed only by filename.
- Any CI run AFTER `3d9515cc` — none exists (HEAD unchanged); nothing to verify.
- The upstream filing of the bandit ask — human act, deliberately not performed.
- Whether bandit's comma-vs-space nosec parsing asymmetry (comma suppressed B607 but not B603) has a discoverable mechanism in bandit's source — I recorded the FACTS (verified empirically) without reading bandit's parser; the draft says exactly that.

## Session shape

One continuous execution of the 16:14 SUPERB plan: write+ship the plan →
certify (full buildflow, background) → close the bandit count with a real
local runner → record the anti-silent-loss counts → harvest the residue →
re-certify docs gates → push → CI heartbeat to green. Wall-clock 16:14 →
~19:08 execution, reported at 22:04. Concurrent sessions were active
throughout (their full buildflow runs interleaved with mine; their closeout
report landed mid-session).

## a) FULLY DONE

1. **Plan written, committed, pushed** — `docs/planning/2026-11-12_16-14_SUPERB-gate-setup-hardening.md` with the Pareto ladder, 30–100min table, ≤12min table, mermaid execution graph, and an errata section that superseded two weaker ideas from the earlier status report (the directive-pinning arch test → armed-count note; stale-B105 deletion → verify-only). Evidence: commit `22d3afed` in pushed history.
2. **The certification goal (P2)** — full `buildflow` over the post-triage tree: 75/76 with the single failure root-caused (state-DB query: `go-mod-vendor-tail`, infrastructure family — the module-cache contention class under concurrent go commands), the flake **re-verified scoped green**, and a concurrent session's **76/76** full run on the same tree as independent confirmation. This also closes the 2026-11-11 triage report's open b2 (full run owed).
3. **The bandit story (P3 + P8) fully closed** — local runner verified; the 6-vs-5 discrepancy resolved (the count was always 6; the sixth was B603@138 hidden by report-log folding); the comma-form nosec defect discovered, fixed to space form, and the tree now scans **0 real issues**; the upstream draft carries two dated addenda (reobservation + reconciliation) with a locally reproducible command. Evidence: fresh `nix run … bandit -r .` output; draft diff.
4. **The anti-silent-loss guard (P4) landed** — `.buildflow.yml` now records the expected scoped finding counts with the F031-lesson spelled out ("the loss hid because the total coincidentally stayed 12"); a live scoped run agrees with the recorded 11. Evidence: config block in HEAD; scoped `"total": 11`.
5. **The harvest (P5) landed** — TODO_LIST "Gate-setup hardening residue" subsection (6 rows, Evidence-cited, house format), ROADMAP owner-infra section with `(repo: BuildFlow)` tag, and the `e897efe8` CHANGELOG `[Unreleased]` backfill (grep-proven absent first — T17 of the deepening plan had landed its code without the changelog line). Evidence: all three files in pushed history (`e9b9ccf1`/`2ecd1831`).
6. **Docs gates re-certified twice** — `scripts/status-index.py` + `nix develop -c scripts/docs-gates.sh` → PASS on all gates (freshness/audit-all/annotate/routing/cadence/status-f-section/battery/table-damage/strike-safety), with the annotate toolchain rebuilt per the AGENTS recipe.
7. **CI green, current** — `scripts/ci-verdict.sh 3d9515cc…` reads success; `3d9515cc` is still HEAD and == origin/main at 22:04, so the verdict covers the tip as it stands. The earlier docs-health red (14:26/14:43) is healed by the fresh report + harvest, as predicted.
8. **Push discipline held under daemon racing** — all five closeout files verified in the pushed tree by content (`git show <sha>:<path> | grep` on the counts line, the harvest header, the space-form nosec), not by commit-message trust.
9. **F13 closed without a row** — the cqrs-lint suppression-state-machine upstream draft was verified to EXIST in BuildFlow `docs/drafts/` (`ls`), so the status report's "verify it exists" item dissolved instead of becoming a TODO row.

## b) PARTIALLY DONE

1. **The canonical bandit runner (S)** — verified working (`nix run nixpkgs#python312Packages.bandit`) but NOT pinned: the TODO row asks for a devshell pin or `scripts/bandit.sh` wrapper, because today's command rides the UNPINNED registry nixpkgs and could drift. Remains: one flake line or one wrapper file. Blocker: none; deliberately parked as a row.
2. **`buildflow format` re-verify (S)** — deliberately NOT run: it mutates the whole tree and concurrent sessions held in-flight files for the entire session. Rides TODO row 2.
3. **Verify-only dispositions executed as decisions, not edits** — stale-B105 cleanup (TODO row 4) and the armed-counts/mensurability follow-ups are recorded dispositions; nothing was deleted or re-rolled. This is the verschlimmbesser fence working as designed, listed here so the "not done" is visibly deliberate.
4. **CI green scope** — green on `3d9515cc` with no newer commits at 22:04; any FUTURE daemon push re-opens the watch (standing obligation, not a gap in this session).

## c) NOT STARTED (parked with owners/triggers — deliberately)

1. **Upstream filing of the bandit ask** — owner-ticked act (TODO row 5); the draft is filing-ready with two addenda.
2. **BuildFlow binary rebuild/reinstall** (`8936ca0`→`39c4789`) — ROADMAP, owner infra.
3. **BuildFlow state-DB purge decision** (0.90 GB) — ROADMAP; the cache VACUUM was done last session, the state DB is the remaining half.
4. **io-history-ratchet `--profile-io` pass** — ROADMAP.
5. **dprint remote plugins → nix-store pin** — ROADMAP.
6. **templ-fmt devshell pin alignment** (v0.3.1020 vs v0.3.1100) — TODO row 3; the fingerprint-drift advisory's root.
7. **Stale `# nosec B105` re-check on next touch** of those lines — TODO row 4, verify-only.
8. **AGENTS.md b5 one-liner** (per-site cqrs-lint directive convention) — TODO row 6, owner call, cap has room.
9. **Calendar-gated**: monthly tier re-measure due **2026-12-08** (now with the armed-counts note as its baseline); quarterly art-dupl baseline audit; post-`-f json` nosec re-verification (blocked upstream).

## d) TOTALLY FUCKED UP (own errors, radical honesty)

1. **The fabricated sha burned 11 minutes of polling.** My first CI loop polled `7dfcab3eb0f1a5d0` — I INVENTED the tail `a5d0` beyond the 12 characters `cut -c1-12` had shown me. Twelve NO-RUNS polls against a commit that does not exist. Root cause: decorating a truncated sha instead of re-running `git rev-parse HEAD` into a variable — the fix I applied immediately after and which worked first try. Severity: wasted wall-clock, zero damage; but it delayed the green verdict by ~11 minutes and produced a misleading NO-RUNS trail.
2. **"cardiovascular" is permanently in commit `22d3afed`'s message body.** A heredoc authoring slip in the plan commit; it is pushed history and unfixable without force-push (prohibited, and the daemon was racing). Root cause: no self-review pass over the assembled message before committing. Severity: cosmetic scar in the permanent log; disclosure here is the mitigation.
3. **The exit-capture bug — a DOCUMENTED trap I walked into.** `echo "exit=$?"` measured `tail`'s exit through a pipe (PIPESTATUS is unavailable in this shell — AGENTS.md says exactly this). The "exit=0" next to a RED label was meaningless noise. Root cause: piping a verdict script instead of using `v=$(cmd); e=$?` from the start. The corrected pattern is what the successful loop uses.
4. **Closeout commit raced the daemon and lost, twice.** By the time my `git add … && git commit && git push` chain ran, the daemon had already committed AND pushed all five files (2-minute cadence — documented behavior); my chain produced a "no changes added" no-op. Root cause: I assumed my files were still uncommitted instead of re-checking `git status` in the same breath as staging. Severity: none (content verified in tree), but the narrative-commit doctrine lost again to heuristic chunks.
5. **The mid-run file-modified dances.** `.buildflow.yml` and the plan report both hit "modified since read" — once legitimately (the docs-gates fmt step reformatted the report WHILE the gate ran), once my own earlier edit registering. Cost ~3 extra rounds. Root cause: writing docs while a gate that formats those docs is running — the order should be write-first, gate-once.
6. **Clock narration drift.** I misread buildflow's run-ID clock as "+1h skew" and narrated wrong wall times for the certification runs before settling on `date` as the only clock. Severity: none in artifacts (the report uses `date`), but the intermediate narration was wrong twice.
7. **One bash typo** (stray quote after `===BUILDFLOW-JOB===`) — one wasted call, self-explanatory.

## e) WHAT WE SHOULD IMPROVE (process, grounded in this session)

1. **Shas are never typed, extended, or trusted from display truncation** — `SHA=$(git rev-parse HEAD)` into a variable is the ONLY pattern (d1's entire root cause).
2. **Commit message bodies get a read-back before commit** — the heredoc footer template as a shell variable would have caught d2 in five seconds; prose sanity is not gated by any hook.
3. **Daemon-aware staging**: `git status --short` + `git add` + `git commit` in ONE call, with an explicit branch: if my files vanished from status, the daemon won — skip silently instead of running a no-op chain (d4).
4. **`v=$(cmd); e=$?` for every verdict script, no pipes** — elevate the AGENTS PIPESTATUS warning from "known" to "habit" (d3).
5. **Write-all-docs-first, then gate once** — never interleave doc edits with a gate whose fmt step rewrites the same docs (d5); the docs-gates run is a certification, not a collaborator.
6. **Poll loops should label-match from poll one** — 'pending' is in-flight, 'failure' is real; the generic exit-code break conditions wasted a loop design iteration (and the pending→exit-2 mapping needs label-sniffing anyway).
7. **Start the long certification immediately after the first push** — this session got that right (buildflow backgrounded while bandit/harvest proceeded); keep it as the standard shape.
8. **State-DB forensics beat log tails** — `buildflow history` + a `resume_steps` query named the failed step in two commands after the streamed log's tail had already been truncated; go to the DB first for "which step failed".

## f) Top next tasks (session-grounded only — most are ALREADY harvested rows; this section lists the DELTA and the pointers)

| #  | Task                                                                                                                                                                                                                                        | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Land the canonical bandit runner (devshell pin or wrapper) — TODO_LIST row 1; `nix run` works but is registry-unpinned                                                                                                                      | High   | S      | Quality       |
| 2  | `buildflow format` in a quiet window (no concurrent in-flight files) — TODO row 2                                                                                                                                                           | Medium | S      | Quality       |
| 3  | Align the devshell templ pin with the CLI (v0.3.1020 → v0.3.1100) — TODO row 3, retires the fingerprint-drift advisory                                                                                                                      | Medium | S      | Quality       |
| 4  | File the bandit upstream ask (owner-ticked; draft carries both addenda + reproducible command) — TODO row 5                                                                                                                                 | High   | S      | Documentation |
| 5  | BuildFlow rebuild/reinstall (`8936ca0`→`39c4789`) — ROADMAP (repo: BuildFlow)                                                                                                                                                               | Low    | S      | Cleanup       |
| 6  | BuildFlow state-DB purge decision (0.90 GB) — ROADMAP                                                                                                                                                                                       | Low    | S      | Cleanup       |
| 7  | io-history-ratchet `--profile-io` pass — ROADMAP                                                                                                                                                                                            | Low    | M      | Quality       |
| 8  | dprint plugins → nix-store pin (`BUILDFLOW_DPRINT_PLUGINS`) — ROADMAP                                                                                                                                                                       | Low    | M      | Quality       |
| 9  | AGENTS.md b5 one-liner (owner call) — TODO row 6                                                                                                                                                                                            | Low    | S      | Documentation |
| 10 | Stale `# nosec B105` re-check ONLY on next touch of those lines — TODO row 4 (verify-only stands)                                                                                                                                           | Low    | S      | Cleanup       |
| 11 | NEW (this session): the `go-mod-vendor-tail` contention flake cost one certification round under concurrent sessions — consider a BuildFlow-side retry-once/lock discipline for the module-graph lanes (candidate upstream ask, owner call) | Medium | S      | Quality       |
| 12 | NEW: `ci-verdict.sh` maps 'pending' → exit 2, so heartbeat scripts must label-sniff; a distinct in-flight exit (or `--wait` flag) would make poll loops semantic — small script change, policy call                                         | Low    | S      | Quality       |
| 13 | NEW: tie the `.buildflow.yml` armed-counts note to the 2026-12-08 tier re-measure — the re-measure updates the counts (one home, one trigger), else the note drifts                                                                         | Medium | S      | Documentation |
| 14 | NEW: cross-check the concurrent session's T17–T24 closeout report §f against my harvest for semantic duplicates (I observed its landing; not read — foreign session's file)                                                                 | Low    | S      | Documentation |
| 15 | Standing: CI watch on the NEXT daemon push (the green is anchored to `3d9515cc`; any newer commit re-opens it)                                                                                                                              | High   | S      | Cleanup       |
| 16 | Calendar: monthly tier re-measure due 2026-12-08 — update the armed-counts note with its fresh numbers when it fires                                                                                                                        | Medium | M      | Quality       |
| 17 | Calendar: quarterly art-dupl baseline audit (last full match-verification 2026-11-11)                                                                                                                                                       | Low    | M      | Quality       |
| 18 | Blocked-upstream: after the bandit `-f json` fix lands, re-verify the nosec set and watch the furniture count collapse                                                                                                                      | Low    | S      | Quality       |

Deliberately EXCLUDED: everything outside this session's gate/tooling surface
(product trains live in today's earlier plans and the deepening session's
closeout; nothing here re-dispatches them).

## g) Top 3 questions I cannot answer myself

1. **The contention flake:** `go-mod-vendor-tail` failed once under concurrent-session module-cache pressure and passed on every rerun (mine scoped, theirs full). Do you want an upstream BuildFlow ask drafted — a retry-once or lock discipline for the module-graph lanes — or is the flake accepted noise given the rerun culture? I tried: state-DB root-cause, scoped re-verify, cross-run confirmation; the only remaining decision is whether the flake class is worth an upstream ask.
2. **`ci-verdict.sh` semantics:** 'pending' currently maps to exit 2 (red), so heartbeat scripts must sniff labels. Should the script gain a distinct in-flight exit (or a `--wait` flag), or is the conservative mapping deliberate so scripts never mistake in-flight for green? I tried: the script's output contract, AGENTS's three-exit description, and two poll-loop designs; the intent behind the mapping is not written anywhere I can find.
3. **Narrative commits vs the daemon:** my closeout files landed as heuristic `chore:` chunks twice this session because the daemon's 2-minute cadence beat my commit chains. Do you want own-work commits staged+committed IMMEDIATELY after the last edit of each unit (accepting more, smaller commits), or is heuristic-chunk history acceptable for doc/config residue? I tried both orders this session; the daemon wins whenever I batch.

Waiting for instructions.

## Post-report gate state (observed 22:1x — live corpus dynamics)

`scripts/docs-gates.sh` after this report landed reads **freshness FAIL +
cadence FAIL**, and the attribution is precise:

- **freshness** — the unharvested item belongs to the concurrent session's
  brand-new `2026-11-12_22-05_self-review-deepening-tail-honest-accounting.md`
  (its §f item 33, the `.cqrs-lint.json` reason-audit trigger). That session
  is editing the file RIGHT NOW (uncommitted mod at 22:1x) — its harvest is
  in flight; per the fix-forward protocol I did not touch it. This report
  itself reads **0 unharvested**.
- **cadence** — the corpus tapped its ceiling: **66 unarchived reports
  (threshold 65)**; this report was #66. The oldest is 19 days (threshold 45),
  so the drain ritual is what the planned **T24 release-tails row** ("the
  18-08 archived-corpus drain train") already owns — check-rows + bare-rows
  before each `git mv`, banner, re-run gates. Deliberately NOT started here:
  it is a named planned train, and the annotate toolchain's ratchet re-measure
  belongs to it, not to a status-report session.
- audit-all (54 ≤ 72) and annotate (204 ≤ 260) hold under their ratchets.
- CI was GREEN on `3d9515cc` at ~19:08; the corpus state above rides the
  NEXT push (tip moved to `3b56e3b4` with this report).

Net: nothing in this session's deliverables is unharvested or unrouted; the
two red gates are (a) a concurrent session's minutes-old harvest window and
(b) the already-planned drain train's trigger having fired.
