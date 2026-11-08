# SUPERB Phase 5 mid-flight status — gates closed, upstream trim shipped, core swap 80 %

Snapshot-rot header: claims verified as of **2026-11-08 22:46 CET** (local
runs; push/CI statements rot fastest — timestamps inline).

Scope of this report: THIS SESSION ONLY (resumed from
`docs/status/2026-11-08_17-22_superb-phase4-implementation-status.md`).
Owner's standing order was "GET SHIT DONE, THE WHOLE TODO LIST"; the last
instruction was "STATUS THEN WAIT" — this report, then the session waits.

---

## a) FULLY DONE (verified this session)

### The Phase-4 closeout (A25/A30b debt) — ALL GATES GREEN

1. **No-cache buildflow: 74/74 steps green, 100 %** (run
   `20261108-204740`, coalesced across 3 racing invocations, findings
   gate green `suppressed=46`). The A25 debt is paid.
2. **`nix flake check`: all checks passed** — INCLUDING the KVM-gated
   `webphone-backup` VM test and the error drill (the §f-2 residue from
   the Phase-4 report is thereby also cleared: the module-check and
   backup legs ran locally).
3. **Live smoke ×2 green**: legacy pins 100 checks (51+4+8+9+9+8+11),
   the 2.10.0 binary 101 checks (adds the `--expect-version` assert).
4. **CHANGELOG Phase-4 entry + plan Errata Phase-4 as-built notes**
   (per-owner-stream ruling, the A27 panic lesson = battery-first rule,
   the EnforceCaps deadlock story, migration twins, cqrs.engine_count,
   decision-level cap tests) + the release-cadence note (below).

### Upstream go-cqrs-lite — two capabilities designed, fixed, released

5. **#123 EnforceCap deadlock FIXED** (upstream `de39a43a9`, pushed,
   CI-verified): root cause = lock-order inversion — `EnforceCap` held
   the query's fold lock while waiting for a pooled connection; the fold
   path (`dispatchFoldsCoreLocked`) holds the connection inside `RunInTx`
   while wanting the fold lock. Fix: `withCapTrimTx` runs the trim inside
   the engine's `RunInTx` with the fold lock INSIDE (the replicator's
   order). Regression test `TestSQLiteEnforceCapDoesNotDeadlockAgainstFolds`
   (1-conn pool, 400 applies racing 60 trims, 60 s watchdog): **pre-fix
   FAIL (90 s hang, goroutines parked in WaitGroup.Wait), post-fix green
   ×3 in 0.066 s** — proven in a worktree at the pre-fix commit. Issue
   #123 filed (voice-checker clean) and pushed.
6. **#118 journal trim SHIPPED and RELEASED**: `metaengine.StreamTrimmer`
   (TrimStream: per-stream erasure, snapshot co-trim in the SAME
   transaction under the caller's snapshot collection, fresh-append-at-0
   contract) + `metaengine.JournalSweeper` (TrimJournalBefore: age sweep,
   pre-delete snapshot subquery, NULL-timestamp rows conservatively
   kept); sqliteengine implements both; the memory engine implements
   TrimStream; `System.TrimStream(ref)`/`System.TrimJournalBefore(cutoff)`
   surfaces with sticky-key resolution and `system.ErrTrimUnavailable`.
   Tests: 3 system-level (trim → empty load → fresh re-append; sweep
   boundary; snapshot co-trim through the REAL surface) + 2 engine-level
   (mixed ages, NULL kept, per-stream snapshot fates). Released as
   **metaengine/v4.23.0 + metaengine/sqliteengine/v4.8.0 +
   system/v4.18.0** via `batch-release.sh` in dependency order (the
   standalone-compile guard caught sqliteengine needing an explicit
   `go mod edit -require` pin — the documented chicken-and-egg case),
   proxy smoke green ×3, issue #118 closed with the implementation
   comment.
7. **webphone rides the wave**: metaengine v4.22.0→v4.23.0,
   sqliteengine v4.7.0→v4.8.0, system v4.17.0→v4.18.0; vendorHash
   re-pinned (`5a8DQ2WjIR…`, the reliable `goModules --rebuild` recipe);
   full `go test ./...` green; 2.10.0 binary builds and smokes.

### Phase 5 opened (A31–A33 complete, A34 ~80 %)

8. **A31 design doc** (`docs/planning/2026-11-08_22-19_core-stream-design.md`):
   per-thread streams keyed (owner, network, hex(remote)); facts,
   decider, projections, service swap, importer, retention
   (tombstone → fold → **TrimStream** → blob), D7 await, D8 snapshots
   (EveryNEvents(500)), search strategy (engine has no LIKE — in-memory
   contains over owner scans), perf budget, execution order.
9. **A32+A33 the thread aggregate, battery green (8/8)**:
   - `internal/domain/threadfacts.go`: 9 facts (MessageAppended with
     full row image, OutboundAdvanced, ThreadFlagSet,
     ThreadUnreadCleared, ThreadDeleted with id sets, MessageRemoved,
     AttachmentStored/Removed) + `DeliveryStep` (the DeliveryEvent wire
     form) + row-key types + `ThreadSummaryImage`.
   - `internal/cqrs/threads.go`: 5 commands; **`domain.AppendToThread` is
     the apply** (one home consumed verbatim); deliveryMoves legality at
     the decision via `DeliveryAdvance`; duplicate-message-id rejection
     (importer-retry + webhook-replay guard); delete emits the tombstone
     + per-row removal batch; D8 snapshot strategy;
     `NewThreadsSystem` boot helper.
   - `internal/readmodels/threads.go`: `ThreadGate` (20-method surface =
     every store.Messages method the server/service call, plus
     `AdvanceOutbound(step)`, `CountMessages`, `DeleteThread(reason)`).
   - `internal/cqrs/threadsstore.go`: 3 projections (full-image inserts
     + keyed removes — the contacts typed-Key pattern) + the `Threads`
     backend (owner-scoped lists/search/pages, admin seams,
     provider-ref lookup, delete = tombstone + await + **TrimStream**).
   - `internal/cqrs/threads_test.go`: append/advance, duplicate
     rejection, the delivery machine (out-of-order verdicts, inert
     moves), flags/read/search/archive/auto-unarchive, attachments
     (name-ordered hydration), owner scoping + admin seams,
     **delete removes rows AND journal bytes** (the A36 deletion
     invariant at the aggregate level), restart reload.
10. **A34 service/handler swap (~80 %, compiles)**: `messaging.Service`
    takes the gate (send/refuse/verdict sites construct the SAME events
    and pass `.Step()` — legality now lives in ONE place, the decider);
    `Deps.Messages`/Notifier/export swapped to the gate
    (`readmodels.SummariesOfView` converter, one home); app wiring
    (`ThreadsSvc()` accessor + `ThreadsDomainDecl` in `NewCqrs` + the 3
    app.go sites); messaging suite green on the gate (incl. the
    twin-thread v5 pin); server/app test sites re-pointed
    (messaging ×2, notifier ×2, server_test, session_behaviors).

### Release train 2.10.0 (per D4)

11. CHANGELOG 2.10.0 section cut (P2–P4 + latency beacon + the cqrs
    bump bullet), `webphoneVersion` 2.10.0, Errata cadence ruling
    ("catch-up release carrying P2–P4; per-phase releases resume with
    Phase 5"). **No tag exists yet — the train died twice mid-flight
    (see §d); re-cut is the first next step.**

---

## b) PARTIALLY DONE

1. **A34 final leg**: tree compiles and vets clean; suites verdicts are IN:
   `internal/app` **green**, Ginkgo session-behaviors **9/9 green**,
   but `internal/server` has **4 red tests on the swapped seam**:
   `TestAdminAttachmentAndFaxDocumentFallback` (admin attachment
   fallback — likely AttachmentByIDAny semantics or seeding shape),
   `TestThreadPaginationLoadsOlderPages` (ListMessagesPage — likely the
   page window or hasMore edge over the new read),
   `TestMetricsServesAggregatesOnly` (the known counts re-pointing,
   §b-2), and `TestSSEPushesSwapSafeFragments` (notifier fragment
   parity). Diagnosis/fix is next-step #1.
2. **`webphone_messages_total` metric regression (KNOWN, unfixed)**:
   `metrics.go` still reads `store.ReadCounts` (webphone.db
   threads/messages tables) — but new messages now land in system.db, so
   the metric FREEZES at legacy counts for new traffic. The gate already
   has `CountMessages` (engine-native COUNT) — re-pointing metrics.go is
   a 20-minute fix + test, deliberately not done mid-report.
3. **A36's aggregate half exists**: `DeleteThread` (tombstone → folds →
   await → TrimStream) is implemented and battery-pinned, but the
   retention SWEEP pass (store/sweep.go semantics: cutoff scan over the
   projection, blob deletes, keep-forever gating) is not ported.
4. **The 2.10.0 tag**: prep complete (§a-11), the train failed twice on
   environmental causes (§d-1, §d-5), no tag pushed.

---

## c) NOT STARTED (beyond what §b covers)

- **A35** the importer (webphone.db threads/messages/attachments →
  streams; full-success marker; dual-run row-by-row diff; timestamp
  fidelity; pre-cutover snapshot record) — the riskiest data step.
- **A36** sweep port + the end-to-end deletion gate (journal + views +
  blob + replay-rebuild) + `retention_days=0` pin.
- **A37** the parity battery re-pointing (owner_scoping, threadsdual,
  threads_flags, message byte-pins, webhook pins — several exist against
  the legacy store and stay green as import-source tests; the SERVER
  pins are the ones at risk from the swap).
- **A38** P5 gates + stack browser E2E + perf re-check vs the P2
  baseline + the P5 release train per D4.
- **P6 (A39–A40)** faxes: faxMoves decider, durable provider_ref
  idempotency, fax importer/retention, faxcap gates + release.
- **P7 (A41–A47)**: handler unification sweep, contracts/catalog diffs,
  hooksIdem/CallsIdem → sys.Dedup(), readmodels.db death certificate +
  backup/restore updates, D2 sessions ruling, ladder freeze, doctrine
  rewrite (AGENTS.md no-ES bullet → ES doctrine etc.), final battery +
  release.
- Smaller owed items from the Phase-4 report's §f that remain open:
  exporting the 3 transcripts consts from readmodels (kills the accepted
  const-mirror twin), history-surface smoke scenarios, docs-gates
  cadence run for the new reports (this one makes it 63).

---

## d) TOTALLY FUCKED UP (honest ledger — all recoverable, all cost time)

1. **The env leak that killed release train #1**: I ran release.sh with
   `WEBPHONE_RELEASE_SKIP_VERSION_GUARD=1`; the app's env parser maps
   every `WEBPHONE_*` var onto a config key, so the override leaked into
   the gate's child processes and **failed
   `TestLoadCleanConfigWarnsNothing` from inside my own release run**
   (~35 min of gates burned). Fix: release.sh now unsets the override
   before any child runs (commit `9ff137c7`). Lesson: an override env
   var in this repo is a config-key injection by construction.
2. **Release train #2 died on my mid-edit tree** (3 steps failed): the
   train gates the CURRENT working tree, and it ran while A34 was
   half-landed (red test files). Lesson: NEVER run the release script
   while landing a phase — the train is a whole-tree gate; the daemon
   commits mid-flight chunks underneath it.
3. **`NewThreadsSystem` forgot `sys.Start()`** — the projection worker
   never ran, and I debugged a phantom "projection lag" for ~30 min
   (empty worker checkpoint was the tell I initially misread). The
   working `bootDomainSystem` helper was 40 lines away the whole time;
   diffing against the KNOWN-GOOD boot first would have answered it in
   one look.
4. **Key-inference rabbit hole (upstream metaengine)**: burned real
   context discovering that `deriveKeys` anchors on the INSERT fold's
   key TYPE and that bare-string keys are ambiguous — the answer (a
   typed `Key` struct field in each fact, the exact
   `ContactKey`-in-`ContactDeleted` pattern from Phase 4) was visible in
   shipping code from the start. Corollary discovery that WAS worth it:
   update folds need an insert in the same Query to get their key type;
   that forced the full-image-facts design (decision carries the row
   image), which is simpler than the prev-aware folds I first wrote.
5. **Draft artifacts shipped into files repeatedly**: placeholder
   loops/stubs in threads.go + threadsstore.go + one test draft with
   literal nonsense (`fmt.Fprintln(nil)`), each caught by build/vet but
   only after I'd moved on. Root cause: writing long files in one pass
   without compiling between logical chunks.
6. **`FindThread` semantics changed silently** (no empty-thread row
   until a message lands — legacy created the row eagerly). Design-doc
   Errata note written, but the handlers' compose flow was not re-pinned
   against it yet — the server suite verdict covers it.
7. **The daemon raced my edits twice** (mtime changed mid-edit, one
   multiedit rejected) — recovered each time via fresh-read + python
   surgery, but it reinforces the write-atomically rule for this repo.

---

## e) WHAT WE SHOULD IMPROVE (process, from this session's scars)

1. **Release trains and phase-landing are mutually exclusive** — gate
   the train on "no phase in flight" (a checklist line in release.sh or
   just discipline).
2. **Env-override hygiene**: any `WEBPHONE_*`-prefixed override must be
   consumed-and-unset at the top of the script that reads it (the
   release.sh fix is the pattern; audit other scripts for the same
   shape).
3. **Diff against the known-good twin BEFORE debugging** (the
   `sys.Start` lesson generalizes: every new "why doesn't X work" starts
   with a diff against the closest working example).
4. **Compile between chunks**: write aggregates in 2–3 passes with
   `go build` between them; the placeholder-artifact class disappears.
5. **Upstream API notes as a living doc**: the deriveKeys/type-matching
   rule, the insert-anchors-key-type rule, the WithCapPerGroup/WithCap
   postures, and the trim capabilities should live in
   `docs/research/2026-11-08_system-v4-api-notes.md` (or a successor) so
   the next aggregate doesn't rediscover them. (Not done this session —
   listed in §f.)
6. **Metric re-pointing should have ridden the swap commit**: any seam
   swap must sweep the metric/export/count readers in the same train —
   the `webphone_messages_total` freeze is the exact failure mode of
   leaving one behind.

---

## f) NEXT (ordered; up to 50 — realistic completion list, not padding)

1. Read job 3F3's verdict (server+app suites); fix any red pins.
2. Re-point `metrics.go` counts: threads/messages via the gate
   (`CountMessages` + a `CountThreads`), sessions stays SQL; pin it.
3. Commit the A34 completion; push; CI-watch.
4. Re-run `scripts/release.sh 2.10.0` on a green, idle tree (the prep is
   all in place; the tag + stack relock + aarch64 + gh release ride it).
5. CI-watch 2.10.0 to green; record the tag in the plan Errata.
6. A35: the importer (`internal/app/coreimports.go`) — threads ordered
  by (created_at, rowid) per thread, carrying legacy ThreadIDs +
  CreatedAt; marker `threads_journal_imported`; duplicate-id rejection
  makes retries safe.
7. A35: the dual-run row-by-row diff harness (test-only) — legacy reads
  vs projection reads, every owner/thread/message/attachment.
8. A35: pre-cutover backup snapshot record (runbook note + commit msg).
9. A36: port the sweep (`retention` package over the gate: cutoff scan,
  DeleteThread per swept thread, blob deletes, keep-forever gate).
10. A36: the deletion-gate test (journal empty, all three projections
    empty, blob gone, reset+replay rebuilds WITHOUT the content).
11. A36: `retention_days=0` keep-forever pin.
12. A37: run the full store battery (owner_scoping, threadsdual,
    threads_flags) — keep as import-source tests.
13. A37: server byte-pins + `TestProviderFormPartHeaderGolden` +
    webhook pins green on the swapped seam.
14. A37: SSE fragment parity (notifier pushes byte-stable — morph
    targets, `wp-thread-row`/`wp-bubble` classes).
15. A37: admin surface parity (`/admin/threads/{id}` deep link +
    ListAll).
16. A37: export parity (the zip's threads/messages legs).
17. A38: full no-cache buildflow + `nix flake check` + full smoke.
18. A38: stack browser E2E (445 s budget, re-run-once flake rule).
19. A38: perf re-check vs the P2 baseline (dispatch p95, list/transcript
    reads) — record in the Errata.
20. A38: P5 CHANGELOG entry + plan Errata Phase-5 as-built + release
    train per D4.
21. A39: fax decider consuming faxMoves (+ inbound `received` fact
    split kept); battery-first.
22. A39: durable provider_ref idempotency (supersedes the process-local
    ApplyIdempotent) — engine dedup or a fax-stream command-id gate.
23. A39: fax projection parity (`wp-fax-row`, badge tones) + collection
    retired from live writes (kept as import source per A16).
24. A39: paperless seam untouched proof (existing pins).
25. A40: fax importer + fax retention (if configured) + markers.
26. A40: faxcap smoke + gates + P6 release train.
27. A41: handler unification sweep (every remaining write dispatches).
28. A41: `internal/arch` import-direction rules updated.
29. A42: `/openapi.json` byte-diff (zero or documented drift + pins).
30. A42: `webphone contracts` regen + asyncapi diff.
31. A42: `catalog -o` render diff + governance-lint tree check.
32. A43: hooksIdem/CallsIdem → `sys.Dedup()` (one dedup home).
33. A43: readmodels engine registration removed; file retired; drill
    updated (restore-path test WITHOUT readmodels.db).
34. A44: D2 sessions ruling execution (system-records migration or
    cqrs-lint baseline) + `auth:` prefix + session-behavior suite.
35. A45: ladder freeze comment + retirement plan doc (tables → death
    releases).
36. A46: AGENTS.md doctrine rewrite (no-ES → ES; machine-tables →
    deciders; topology wording).
37. A46: fleet-libs `go-cqrs-lite.md` re-score (trim + deadlock fix are
    new capability surface).
38. A46: ROADMAP annotation + TODO_LIST harvest + FEATURES update.
39. A47: FULL no-cache buildflow + flake check + VM tests (WITH
    system.db, WITHOUT readmodels.db) + vulnix + full smoke + E2E.
40. A47: final version bump + release + push + CI-watch + stack re-pin
    coordination note.
41. Export the 3 transcripts consts from readmodels (kills the accepted
    const-mirror twin) + dedup-registry row update.
42. Write the upstream API-notes addendum (deriveKeys rules, trim
    capabilities) into docs/research.
43. Consider smoke scenarios for the notes/reviews/transcript HTTP
    surfaces (P4's deferred "history surfaces" line).
44. Regenerate `docs/status/INDEX.md` (this report makes 63) + a
    docs-gates cadence run.
45. File/track the metrics-count decision (see §g-2) wherever it lands.
46. Audit scripts/ for other `WEBPHONE_*` override leaks (the release.sh
    pattern).
47. Sweep the plan Errata for the FindThread empty-row deviation (§d-6)
    once the server suite verdict is in.
48. Re-verify the phase-4 transcripts EnforceCaps workaround can be
    simplified now that upstream v4.23.0 is pinned (optional; the
    await-first posture is defense-in-depth and correct either way).

---

## g) QUESTIONS (cannot be figured out from the tree)

1. **2.10.0's content boundary**: the train never tagged. Two clean
   options: (a) cut 2.10.0 from a green tree as-is — the tag then
   carries the Phase-5 A32–A34 swap (the changelog would need one
   honest line saying the journal-backed thread seam rides along,
   unwired importer = legacy rows invisible until A35); or (b) hold the
   tag until A35 (importer) makes the swap user-visible-correct, cutting
   2.10.1/2.10.0 then. (a) ships the debt-free P2–P4 story now but with
   an in-between seam state; (b) is cleaner but couples the P4 release
   to P5's riskiest task. Which?
2. **The metrics posture for the seam window**: `webphone_messages_total`
   (+ threads count) must move to system.db counts. Until A35 imports
   legacy rows, gate counts undercount (new journal messages only);
   after A35 they converge. Options: (a) re-point now and accept the
   window (metric honest about the NEW store), (b) sum legacy + gate
   counts until the ladder retires (metric continuous, two homes), (c)
   leave frozen until A35 (dishonest — rejected). Which?
3. **Per-phase release cadence going forward**: D4 says one release per
   phase; 2.10.0 slipped to a catch-up because per-phase trains didn't
   run. P5, P6, P7 each end with a full train (buildflow no-cache +
   flake + smoke + E2E + tag + stack relock ≈ 1.5–2 h each at best).
   Keep per-phase releases (D4 as ruled — trivial rollback, real cost),
   or amend D4 to "one release per phase BOUNDARY THAT SHIPS USER-
   VISIBLE CHANGE" (P6 faxes + P7 closeout would then be one train)?
   This decides §f items 20/26/40.

---

**Session posture**: the tree compiles and vets clean; server+app suites
were running at snapshot time (3F3); NOTHING is pushed past `9ff137c7`
except what the daemon lifted. The session WAITS here per the owner's
instruction.
