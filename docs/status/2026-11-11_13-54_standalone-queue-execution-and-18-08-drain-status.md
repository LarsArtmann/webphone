# Status — standalone queue execution: stack CI heal, 18-08 drain, cadence hygiene

**Written:** 2026-11-11 13:54 CET
**Scope:** this session's executed work only — the stack CI heal, the oldest-live-report
drain + archive, the runbook/script/TODO hygiene, the erraudit and clone-attribution
re-measures, and the cross-session attribution of webphone main's ongoing red.
**Snapshot-rot notice:** every claim below was verified as of 13:54 CET 2026-11-11.
Push/CI statements rot fastest; treat them as timestamped, not timeless.

---

## a) FULLY DONE

1. **Stack CI healed and CI-verified green.** Two sweep-damage classes, both fixed in
   `bb9421a` and confirmed by the workflow-verified verdict: the lock-move-guard wanted
   CHANGELOG attribution for the daemon's second lock sweep (webphone
   `99953a95785b` → `9b47bc783527`, pbx-lens `51777b6e` → `27fd19a72355`, disko
   untracked) — the previous entry named only the superseded intermediate rev; and the
   marker-history arm caught the 2026-10-07 snapshot's f#1 verdict cell re-dropped by a
   later sweep — restored byte-identical to its historical peak (146 markers, verified
   by diff). Local gates green (lock-guard PASS + self-test, marker sweep 0 findings)
   before the push.
2. **The 18-08 report drained gates-first and archived** (the session's largest item):
   102 bare rows + one wholly-untouched 14-row table, all 129 items adjudicated with
   fresh evidence — resolved-at-run (fact/achievement tables), resolved-since (the
   standing release gates: flake check, vulnix, smoke, ui-capture, the errorfamily
   tiers), routed to TODO_LIST's new 18-08 harvest block (traces-in-backup owner
   ruling, observability docs rows, flight-recorder hardening cluster, lessons
   candidates, the go-appkit direct-dep fact), NOT-DO-superseded (closed plan-file
   rows), wont-implement (considerations with no reopen pressure). check-rows --strict
   and bare-rows both clean BEFORE the git mv; INDEX regenerated; committed `1da1bc73`.
3. **The drain's cell mangle repaired.** My bulk verdict wrap swallowed the closing
   pipe into 102 bold markers; strike-safety refused the rows; repaired to the
   canonical `**→ verdict** |` shape and re-verified (both archive gates + strike
   safety clean). Landed via the daemon sweep in `dadf48e8`.
4. **Runbook train-hygiene rules landed** (docs/release-runbook.md § "Train-hygiene
   rules from the v2.13.0 seven-attempt war"): check-mode docs battery as the pre-push
   verdict; HEAD-assert inside the same tool call as the push; archive-when-filing;
   release logs copied home at attempt completion; the entangled-red-train rule (no
   pushable prefix heals a red whose fix is split across another session's unpushed
   commits — today's 09:47 island-js red was the evidence).
5. **scripts/cadence-candidates.sh landed**: the archive-candidate scan behind the
   cadence budget (check-rows --strict + bare-rows per live report, oldest-first CLEAN
   list, INDEX.md excluded, builds the session-fresh annotate tool into /tmp/annotool
   when tmpfiles wipes it). First run: 9 clean / 57 dirty.
6. **erraudit tier re-measure (was overdue since 2026-11-05): CLEAN.** BuildFlow's
   erraudit step 1/1 green (RC=0, detect 3.2s). The raw CLI's 119 findings are the
   bypass-suppression class, not tree regressions — the authoritative gate is
   BuildFlow's wiring. Next calendar due 2026-12-08.
7. **Post-refactor attribution ran clean**: internal/arch tests green on the swept
   tree (error-code registry freshness, module-graph acyclicity); art-dupl at the
   pinned v0.7.2 detector reports 2 new clone groups, both in the active session's
   in-flight templ files (login_page/layout/error) — theirs to baseline, left alone.
8. **Two more gate-clean reports archived in the same push as this filing**
   (the archive-when-filing rule, executed not just written): the 2026-11-09 07-53
   closeout-round2 and the 2026-11-10 14-42 prod-screenshot-tracking reports — the
   two-oldest CLEAN candidates; live count lands at exactly the 65 cap.
9. **TODO_LIST hygiene**: struck the runbook-additions row (+ its daemon duplicate),
   the cadence-script row, and the erraudit row with dated verdicts; updated the
   red-main-watcher row with the 4h-no-issue evidence; added the 18-08 harvest block
   (5 rows) and the cadence-continuation row.
10. **Every push ran HEAD-assert** (capture HEAD, assert unchanged at push): three
    pushes this session, zero foreign-sweep surprises published.

## b) PARTIALLY DONE

1. **Webphone main's red is attributed but not healed.** All four failing ci.yml jobs
   (smoke, go test -race, fast flake checks, art-dupl) trace to the active session's
   login-page train mid-flight: island ids missing from the served shell, a nonce'd
   inline script tripping TestAppServesMainPageUnderStrictCSP, the styled-404 surface
   bare, and the two in-flight templ clone groups. Verified per failing job on my own
   push's run — none of the failures trace to my deltas. The owning session is
   actively committing (tip moved three times during this session) and built
   scripts/ci-verdict.sh mid-train, so the heal is theirs; I left their files alone
   per the concurrent-sessions doctrine.
2. **B5 staging is stale.** The daemon swept the pbx lock past the staged exact pin:
   it now rides stack `aceaf598`/webphone `9b47bc78` (superseding the `bef56520`
   v2.13.0 staging). The pre-state verification and toplevel build must be redone
   before any switch; deliberately deferred because the switch itself is gated on Q1.
3. **The cadence budget is balanced only at this instant** (65/65 after this filing).
   The active session's two 12:2x-12:3x filings are what pushed it over; if they file
   again without archiving, the budget breaches again. The next-oldest DIRTY
   candidates (57 of them) owe drain trains per the TODO row.

## c) NOT STARTED

1. **Q1 — the B5 prod switch** (authority + timing). Everything downstream waits:
   the fresh pre-state verification, the toplevel rebuild on the swept lock, the
   switch itself, the post-switch battery (prod version, journal quiet, ui-capture
   --track, deploy-ledger row), and the post-switch pbx telephony bump.
2. **Q3 — the E3 budget number**, which gates e3.2.
3. **The owner legs** (explicitly do-not-start): A3.4 Telnyx reply send, C3.2 inbound
   mobile call, E5.2–E5.4 hikaru legs, E6 first real fax.
4. **The red-main watcher investigation** — the row is updated with evidence (the
   09:47 red sat unfiled for 4h+), but the watcher's red-detection has not been
   read, reproduced, or fixed.
5. **The annotate_max ratchet re-measure** after today's archive moves (the
   continuation TODO row owns it; the battery was red all session on the other
   session's report, so a clean measurement was not obtainable).
6. **The next-oldest DIRTY candidates' drain trains** (57 reports; the 45-day
   thresholds start landing 2026-12-07).

## d) TOTALLY FUCKED UP

1. **I inflicted the exact mangle the strike-safety gate exists for.** My bulk
   bold-arrow wrap put the closing `**` after the cell pipe in 102 rows
   (`**→ … |** |`). check-rows/bare-rows TOLERATED the shape (their regex matches
   across the misplaced pipe); strike-safety refused it — but only after the archive
   commit had landed. I ran the strike gate at the archive ritual, not before the
   first commit. The repair is in, but a one-line pre-commit strike-safety run would
   have kept main's history clean of the mangle entirely.
2. **I ran the raw erraudit CLI before the authoritative gate and alarmed myself.**
   119 violations, 61 CRITICAL — all of it the bypass-suppression class that
   BuildFlow's wiring (and the lessons' read-the-pins-first rule) already governs.
   The scoped buildflow run answered 1/1 green in 3.2s. The expensive order was
   mine; the cheap order was documented.
3. **I repeated the misread-green class while it was being fixed.** My first CI check
   read watch.yml successes as "CI green" — the exact failure the other session
   shipped scripts/ci-verdict.sh for TODAY (its AGENTS rule landed in `dadf48e8`).
   I corrected to the workflow-verified tool mid-session, but the first read was
   wrong in the documented direction.
4. **I nearly pushed a partial fix as "the fix".** The 09:47 island-js red looked
   one-commit-deep (`711f0ac5`'s guard change); the detached-worktree check showed
   that commit ALONE still fails the test — the test-side repair landed three commits
   later inside their train. The worktree verification caught it before any push,
   but I had already framed it internally as the manual-push-bar case. Verify the
   fix claim BEFORE adopting the bar.

## e) WHAT WE SHOULD IMPROVE

1. **Gate order: strike-safety before ANY commit that touches verdict rows**, not
   just before the move. The battery's gates each answer a different corruption
   class; passing two of three proves nothing about the third.
2. **Adopt scripts/ci-verdict.sh as the reflex.** `gh run list` without the
   workflow filter reads the watch lane's greens as CI verdicts. The tool exists,
   is self-tested, and its AGENTS rule landed today — use it first, every time.
3. **Run the authoritative wiring first for every re-measure.** Raw CLI flags
   bypass BuildFlow's suppression layers and rule configs; the scoped
   `buildflow -s <step>` run is both faster to interpret and the actual bar.
4. **The /tmp toolchain rule exists because of this exact session shape** — rebuild
   the annotate tool at session start, not "check it exists" (it survived today by
   luck of the 06:52 build).
5. **Shared-file edits want HEAD-capture around the edit batch, not just the
   push.** The daemon swept the 18-08 file between my read and my banner edit
   (benign — my edits were intact — but the pattern is the 2026-10-27 loss class).
6. **Attribution before push for swept foreign work**: the daemon commit I published
   carried the other session's AGENTS/lessons additions; I read the diffs before
   pushing, which is right — make it a stated step, not an afterthought.

## f) UP TO 50 THINGS TO DO NEXT (ledger — harvest inventory, not commitments)

<!-- check-rows:ledger -->

| #  | Task                                                                                                                                      | Impact | Effort | Category    |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----------- |
| 1  | Answer Q1: B5 prod-switch authority + timing (the gate for the whole deploy line)                                                         | High   | S      | owner       |
| 2  | Answer Q3: the E3 budget number (gates e3.2)                                                                                              | High   | S      | owner       |
| 3  | Decide heal-or-wait for the active login-page train's docs-gate reds if that session idles (the 12-31 report's missing §f + format)       | Med    | S      | owner       |
| 4  | B5 pre-state re-verification after the daemon's lock sweep (fresh ExecStart, generation, toplevel rebuild on the new pin)                 | High   | M      | deploy      |
| 5  | Execute the B5 switch once Q1 lands (runbook shape: nixos-rebuild switch --flake .#pbx --target-host root@pbx.artmann.tech)               | High   | M      | deploy      |
| 6  | Post-switch battery once B5 executes: prod version, journal quiet, ui-capture --track, pbx-artmann deploy-ledger row                      | High   | M      | deploy      |
| 7  | pbx telephony bump to stack main post-switch, lock-drift-probe, toplevel rebuild                                                          | Med    | M      | deploy      |
| 8  | Investigate the red-main watcher's 4h-no-issue gap (LIST-API dedup vs red-detection; the TODO row carries today's evidence)               | Med    | M      | tooling     |
| 9  | Re-measure the annotate_max pin at the next sweep (the archives moved corpus counts; ratchets move down only)                             | Med    | S      | docs-health |
| 10 | Drain-train the next-oldest DIRTY candidates from scripts/cadence-candidates.sh before their 45-day thresholds (~2026-12-07 first)        | Med    | L      | docs-health |
| 11 | Traces-in-backup ruling (owner): should `<dataDir>/traces` ride the backup set or be excluded + documented                                | Med    | S      | owner       |
| 12 | Observability docs rows: FEATURES histogram wording + flight-recorder rows, ROADMAP observability entry (grep-verified absent)            | Med    | S      | docs        |
| 13 | Flight-recorder + metrics hardening cluster (the 18-08 harvest row: trace/metrics alignment, module surface, rate limit, and the rest)    | Med    | L      | webphone    |
| 14 | Lessons candidates: the "plan says X, the code pins Y — amend the plan in the same turn" lesson + the on-disk-artifact consequence lesson | Low    | S      | docs        |
| 15 | go-appkit report refresh: correct the report's flight-recorder fact (indirect dep → the DIRECT dep reality) per its own trigger           | Low    | S      | docs        |
| 16 | The login-page train's island suite + smoke pins once that session lands (their ssoerror/resume reds + the shell-id smoke pins)           | High   | M      | webphone    |
| 17 | art-dupl baseline update for the login-page train's templ clones once their tip is final                                                  | Low    | S      | tooling     |

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Q1 (still open): do I execute the B5 prod switch over ssh — and does "now"
   mean despite the daemon having swept the pbx lock past the staged v2.13.0 pin
   onto the login-page train's era?** The switch would today deploy a webphone
   whose own main CI is red under that train. Wait-for-green is my default; your
   authority call is the gate.
2. **Q3 (still open): the E3 budget number** for e3.2.
3. **Cross-session intervention: if the login-page session idles with main still
   red (docs gates + island suite + smoke pins), do you want me to heal its
   in-flight surfaces (fmt its 12-31 report, add the missing §f, update the smoke
   pins to the new served shape), or does that train stay exclusively its
   owner's?** Today I stayed hands-off; a standing rule would settle every future
   instance.
