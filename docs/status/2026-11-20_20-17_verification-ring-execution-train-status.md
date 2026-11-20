# Verification-ring execution train status — 2026-11-20 20:17

> Snapshot-rot header: claims verified as of 20:17 CET. CI on the
> pushed tip `6498c8ca` reads GREEN via `scripts/ci-verdict.sh` (exit
> 0, workflow-name-verified). Scope: THIS session only — executing
> `docs/planning/2026-11-19_20-51_SUPERB-catalog-integrity-verification-ring.md`
> (TA–TH) under the owner's "whole TODO list" directive, plus the
> honest self-review the owner asked for afterward.

## a) FULLY DONE (verified this session)

1. **TA — verified, not rebuilt.** The sibling verification session
   had already landed the bump + rulings (`9ba336cc`, TODO:1744). This
   session independently verified the mechanical half: the domain
   catalog pin suite ran green, and fresh exports carry `review.saved`
   at version 1.1.0 in BOTH trees.
2. **TB — the verification ring, all three legs** (`132e1bb0`).
   (1) LIVE oracle: `check-catalog-contract.py` against
   `https://catalog.home.lan/…/catalog-frontmatter.v2.schema.json`
   over BOTH fresh exports — clean, exit 0; the vendored CI copy is
   sha256-identical to live (no drift). (2) Cross-repo read:
   eventcatalog-hub@84fcada `scripts/lint-sources.sh:272` consumes
   each source's `lint.export`/`lint.tree` (the `-plain` recipe) —
   webphone's lint block is load-bearing, sources without one are
   skipped. (3) CI `catalog-contract` job now exports-then-validates
   BOTH trees; the job ran green on the pushed tip.
3. **TC — projection DLQ is observable** (`5a3f385f` daemon + docs in
   `2db39ba4`). Inspection confirmed the blind spot exactly: the
   library's `sendToDLQ` stores silently; `/healthz` only sees
   failed/draining workers (a dead-lettering worker moves on);
   `/metrics` had nothing. Shipped BOTH halves:
   `internal/app/dlq.go`'s `loggingDeadLetterStore` (poison-moment
   WARN naming projection/event_type/event_id/error/code + replay
   hint) and `warnDLQDepth` at `CqrsService.Start` (boot re-WARN,
   count + oldest_age via `DeadLetterStoreAdmin`; non-admin stores
   skip silently). Pinned by `dlq_test.go` (unit + seeded/empty/
   non-admin) AND the real-path WARN assertion added to
   `TestCqrsPoisonEventDeadLetters`. Ops copy: README §Troubleshooting,
   AGENTS invariant bullet, CHANGELOG.
4. **TE — metrics off the frozen ladder, RETIRED not rerouted**
   (`1402b98e` + `1c51a807`). The assessment found the legacy
   contribution reads ZERO on every reachable deployment (the A35
   importer stamped its marker on every deployment with legacy rows —
   the upgrade chain forces v2.10/v2.11 passage; fresh boots never
   wrote legacy rows), so `ReadCounts`' legacy halves + the
   `FlagPresent` marker check were permanently-zero per-scrape cost
   and a "no such table" 500 for the whole surface once §5 DROP
   lands. `/metrics` totals now read journal counts directly (names +
   values unchanged; the pin test survived byte-identical);
   `ReadCounts` shrank to `CountSessions`; `FlagPresent` deleted with
   its last reader.
5. **TD — systemscenario ruled ADOPT-LATER** (hands-on spike, no repo
   churn). A throwaway /tmp module against webphone's EXACT pins
   (system v4.18.0, metaengine v4.23.0, sqliteengine v4.8.0,
   systemscenario v4.3.0) drove a notes-shaped toy domain through
   `When(cmd).Then(eventType).ThenQueryFunc(...)`: journal-diff + polled
   assertions both work; boot 23ms / chain 2ms vs the hand lane's
   0.01–0.05s (a wash). Why not now: no behavior gap (machine tables
   pin the write side), a new direct dep rides the 2026-10-11 protocol
   tail, and "fixture-from-production" doesn't close the boot-drift
   gap by itself. Reopen triggers recorded in the row (TODO:1748).
6. **TF — snapshot assessment, premise half-stale** (TODO:1749). The
   ONE unbounded aggregate (threads) is ALREADY snapshotted (D8:
   EveryNEvents(200) + ReadPressure(50), A38-resolved). The remaining
   exposure is per-OWNER streams (transcripts, agent_audit, call_notes,
   …) whose trim tombstones grow versions unbounded under capped
   collections; worst shape modeled (heavy ASR user ⇒ ~300–500k
   versions after 6 months ⇒ tens-to-hundreds of ms per warm append).
   Operational adopt-later trigger: any owner stream's max version
   crossing ~50k, or measurable transcript-append latency.
7. **TG — the cleanup cluster, bigger than the row's 8-minute
   estimate** (daemon commits + `7c274492`). `messages.go` (696
   lines, test-only since the journal swap) deleted WITH seven
   dead-surface test files (store_test, threads_flags_test,
   threads_bench_test, admin_reads_test, owner_scoping_test,
   scan_panic_test, threadsdual_test — the build-tagged T20 spike
   could not compile after the deletion). `ErrNotFound` moved to
   db.go (production dependents in cqrs). The PRODUCTION-behavior
   pins kept, rewired to raw-SQL seeding against the frozen tables
   (store sweep_test + retention's four sweep pins — they die with
   the §5 DROP, noted in each header). Error-code registry
   regenerated (`store.query`/`store.count_archived` gone with the
   file). App depth-budget pin raised 1650→1700 in-commit with the
   DLQ reason. README's `readmodels.db` note now says "safe to
   delete".
8. **TH — harvest wiring complete.** All seven harvested rows struck
   or ruled with evidence; the 20:23 assessment report's (b)/(c)
   items carry RESOLVED/DONE annotations (non-destructive).
9. **Final gates + push + CI.** Full go suite 37 packages ok, zero
   FAIL; docs-gates battery ALL PASS; art-dupl pinned v0.7.2 check 0
   new clones; dprint clean. The daemon's push stalled 60 minutes
   (anchor 18:15:08, first unpushed commit) → manual push under the
   ratified bar → CI GREEN on `6498c8ca`.

## b) PARTIALLY DONE

1. **dprint/format healing — fixed the right way on the second
   attempt.** The final docs-gates run red'd `format (rc=20)` on the
   TRACKED `work/` catalog-export trees (my TB exports regenerated
   them; the committed bytes predated any dprint pass). First move
   was wrong: `dprint fmt work/` (formatted 20 generated artifacts,
   coupling every regen to dprint). Self-corrected: `git restore
   work/` + `"work/**"` added to dprint excludes — the `docs/contracts/**`
   precedent (generated wire docs are formatter-unowned). Green since.
2. **a01.5's letter vs spirit.** The plan asked to "record in the R31
   row"; the R31/fcr-02 row (TODO:1727) was already struck by the
   verification session and I recorded my independent verification in
   the verification-ring row + the report annotation instead — no
   cross-note was added to 1727 itself. Functionally covered,
   literally not.
3. **TD's spike depth.** The spike mirrored the notes domain's SHAPE
   in a toy module rather than booting the real `NotesDomainDecl`
   (internal packages are unimportable from a separate module; a
   throwaway git-worktree clone + `go get` would have tested the real
   thing at the cost of dep churn inside a worktree). The ruling
   records this openly; the ergonomics/runtime conclusions hold.

## c) NOT STARTED

1. **DLQ replay tooling.** The new WARN line tells the operator "fix
   the handler, then replay" — but webphone ships NO replay command;
   the library's `ReplayDeadLetters` path exists and nothing calls
   it. The observability half shipped; the remediation half is a row
   (§f).
2. **TF's trigger automation.** The 50k-version adopt trigger is
   written into the row but nothing measures or alarms it — it is an
   operator-run SQL today.
3. **The §5 DROP migration planning update.** TE + TG reduced the
   frozen tables' remaining readers to the sweep + its raw-SQL
   seeders; the DROP window is closer than the retirement doc's last
   audit assumed, but that doc was not touched this session.

## d) TOTALLY FUCKED UP

1. **The first `sweep_test.go` write was MALFORMED** — I wrote a
   half-sketch helper (an invented interface, an empty function body,
   an undefined `queryer`) straight to disk and only noticed on
   re-read. The daemon commits within seconds; the broken file sat
   on disk for roughly two minutes. Luck, not discipline, kept it
   out of history. Lesson: compose new test files in one clean write,
   or write to a scratch name and rename.
2. **Three avoidable compile-fix loops in the TG cluster**: the
   appflags import cycle (test in package `store` importing `store`),
   the `FlagValue` 3-return mismatch (wrote the call from memory
   instead of checking the signature I had READ minutes earlier),
   and the missed `scan_panic_test.go` dependency (my dependency
   catalogue before deleting messages.go covered six files and missed
   the seventh). Each was caught by the compiler before commit — but
   all three were the same failure mode: editing from memory while
   the file contents were one View away.
3. **`rg -rn` hit the hook AGAIN** (once, hunting the A38 docs) —
   after the handoff explicitly warned about it. The third-strike
   hook message exists because of sessions like this one.
4. **Commit message typo `feat(dlx)`** for the DLQ surface commit
   (`2db39ba4`) — wrong scope token, left standing (amending
   daemon-shared history is the worse evil).
5. **My own 20:23 assessment shipped a stale claim** that this
   session's TF disproved: "snapshots never assessed" was false —
   threads has carried a deliberate D8 snapshot strategy since the
   SUPERB phases, one `view threadsregister.go` away. The
   assessment's grep-derived inventory was presented with more
   confidence than grep-depth warrants. (The (b)3 item hedged this;
   the hedge was correct, the headline wasn't.)

## e) WHAT WE SHOULD IMPROVE

1. **Verify tracked-path state BEFORE writing into it.** The work/
   incident exists because I exported into a directory I assumed was
   scratch. `git ls-files <path>` before any generator writes is one
   command.
2. **Strike rows the moment their task lands, not in a closing TH
   pass.** TC/TE/TG were each done ~30 minutes before their TODO
   strikes existed — a session death in that window would have left
   done work looking open. The plan assigned strikes to TH; the risk
   window was still real.
3. **The daemon-commit race was lost three times** (the TG closure
   commit even came up empty — exit 1, nothing to commit). Commit
   immediately after each file cluster instead of batching per task;
   verify with `git log -1 --stat` when the race hits.
4. **Grep inventories need source spot-checks for ARCHITECTURE
   claims.** "26 commands / 11 aggregates / never assessed" is a
   census, not an audit — TF proved one cell of it wrong. Future
   assessments: every "X does not exist" claim gets one open-file
   verification.
5. **Depth-budget pins absorb too easily.** The app pin went
   1650→1700 the same session the DLQ decorator landed. Sanctioned
   (reason in-commit), but the honest alternative — a tiny
   `internal/app/dlq` carve — was not seriously weighed before
   raising. Direction-down doctrine deserves a fight first.
6. **Flake-level gates were leaned onto CI.** This session ran go
   suite + docs-gates + dprint + art-dupl locally but never a full
   `nix flake check` (island gates, treefmt sandbox). Nothing island-
   adjacent was touched, and CI green covers it — but the local-gates
   -first habit slipped.

## f) Next tasks (from THIS session's observations — 12, not padded)

<!-- check-rows:ledger -->

| #   | Task                                                                                                    | Why (evidence from this run)                                                                       | Size |
| --- | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ---- |
| N1  | DLQ replay CLI: `webphone dlq list/replay` over `DeadLetterStore.ReplayDeadLetters`                     | the WARN's own hint promises replay; nothing in the repo can perform it (c)1                       | M    |
| N2  | DLQ depth as a go-health WARN-grade leg (non-critical) so the dashboard shows it                        | WARN journal lines are invisible to the dashboard; CRM already models the non-critical leg pattern | S    |
| N3  | Automate TF's trigger: max(version) per stream WARN at boot beside `warnDLQDepth` (threshold 50k)       | the adopt-later trigger is a written row, nothing measures it (c)2                                 | S    |
| N4  | Decide the tracked work/ trees' fate: regen ritual in the catalog train vs move under docs/ vs untrack  | they rotted silently once already and forced the dprint fence (b)1                                 | S    |
| N5  | §5 DROP-migration plan refresh: readers are now ONLY the sweep + its raw-SQL seeders                    | TE+TG shrank the blast radius; the retirement doc's audit predates it (c)3                         | M    |
| N6  | Carve `dlq.go` out of internal/app; lower the 1700 pin back toward 1650                                 | the raise was sanctioned but unresisted (e)5                                                       | S    |
| N7  | Cross-note the R31/1727 row with this session's independent TA verification                             | a01.5's letter (b)2                                                                                | XS   |
| N8  | CI catalog job: consider running the hub's own @eventcatalog/linter over the -plain tree (node in CI)   | deeper parity than frontmatter keys; owner cost/benefit call                                       | M    |
| N9  | Vendored-schema staleness alarm (scheduled LAN job comparing sha vs live)                               | drift is invisible until a hub-side schema bump reds CI cryptically                                | S    |
| N10 | Systemscenario: write the NEXT new aggregate's specs in it from day one (armed trigger)                 | TD ruling's reopen path; porting pinned lanes stays banned                                         | —    |
| N11 | Sweep seeders carry DROP-migration death notes — verify each header names §5 before the DROP lands      | files already note it; a checklist row for the DROP train                                          | XS   |
| N12 | Assessment protocol note: every "does not exist" architecture claim requires one open-file verification | the TF stale-premise lesson, made a rule (d)5/e)4                                                  | XS   |

## g) Questions (cannot figure these out myself)

1. **The push-protocol tension.** The handoff said "push only if the
   owner asks", but the standing 60-minute stall bar (ratified
   2026-10-11) fired at 19:15 with 12 commits stalled and I pushed
   under it. Correct call, or should a session in "await owner"
   posture let the bar win only when main is red? This sets the
   precedent for every future handoff.
2. **The tracked `work/` catalog trees** (N4): keep tracked with a
   regen ritual in every catalog train, move under `docs/` as frozen
   evidence, or untrack and let the CI job's fresh export be the only
   proof? Each has a different rot profile and I have no taste signal
   for which you read.
3. **DLQ replay tooling (N1): build now, or defer until the first
   real dead-letter lands in production?** The WARN promises a replay
   path; building it now costs a CLI afternoon, deferring costs
   nothing until the first poisoned projection — and then it costs
   an outage-window scramble.
