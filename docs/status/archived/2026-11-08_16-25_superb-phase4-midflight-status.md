# Status: SUPERB Phase 4 mid-flight — D7 await, contacts + transcripts aggregates — 2026-11-08

Claims verified as of 16:25 (push/CI statements rot fastest; the CI
verdict line names its run).

## The session in one paragraph

Started from the Phase-3 handoff (A22–A24 landed, A25 ~90%). First
verified push/CI state: 37eb4783 was unpushed and main was RED (CI
42488854067: docs-health cadence — 66 unarchived reports vs the 65
ratchet). Executed the archive ritual (annotate toolchain built from
the skill, six gate-clean reports bannered + `git mv`-ed, INDEX
regenerated at 62, full docs-gates battery green), committed bc754a39
and pushed under the manual-push bar — which also carried the Phase-3
train to remote; CI green (42493268358; a concurrent session landed a
parallel cadence drain 42fdf9e3, no conflict). Then closed A25's last
gate: the full no-cache buildflow ran RED on exactly two
file-size-check errors (dial.go 359, dialimport.go 436 > 350 max) —
~~fixed by splitting both, with dialimport.go restructured into the~~ v: verified: internal/app/importutil.go (appendJournalFacts) lives at HEAD
shared importer primitive Phase 4 needs anyway. Then opened Phase 4
with a plan-mandated mechanism the earlier phases had deferred: the
D7 read-your-writes await. Research found Phase 3 had shipped a LATENT
race (the Settings partial re-renders from the projection in the same
request as the snippet save); the await mechanism (journal-tail read +
projection-worker checkpoint poll, 250 ms bounded) now exists, is wired
into the snippets backend, and is pinned by a test. On top of it: the
A26 contacts aggregate is code-complete, and A27 transcripts is written
but DOES NOT YET COMPILE (one helper export owed). Working tree:
daemon chunked most in-flight files; transcripts.go + transcriptsstore.go
sit untracked.

## a) FULLY DONE (this session)

| #  | Item                                                                                                                                                                                                                                                                             | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~ | 1                                                                                                                                                                                                                                                                                | Todo reconciliation + push/CI truth established (37eb4783 unpushed; main RED on docs cadence)                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ~~ | 2                                                                                                                                                                                                                                                                                | Docs-health archive ritual: annotate toolchain rebuilt, 6 READY reports bannered + archived, INDEX regen (62), battery ALL green                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ~~ | 3                                                                                                                                                                                                                                                                                | Manual-push bar invoked with anchor (main red + verified fix); push verified remote + CI green                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ~~ | 4                                                                                                                                                                                                                                                                                | A25 full `BUILDFLOW_NO_RESULT_CACHE=1` run executed — everything green EXCEPT file-size-check (2 error findings: dial.go, dialimport.go)                                                                                                                                                                                                                                                                                                                                                                                                                            |
| ~~ | 5                                                                                                                                                                                                                                                                                | dial.go split: `Domain` combinator moved to new internal/cqrs/domain.go (336 lines, under cap)                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 6  | dialimport.go restructured: NEW internal/app/importutil.go (`appendJournalFacts`, the ONE shared importer primitive: per-stream tail-sync append); dialimport.go rewritten dial-only (121); NEW journalimports.go carrying snippets + agent-audit + push-devices importers (187) | vet + `go test ./internal/app ./internal/cqrs` green                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| ~~ | 7                                                                                                                                                                                                                                                                                | D7 research: dispatch returns before the fold (~1–2 ms, A06); Phase-3 snippets lane had a real same-request re-read race; upstream has NO await primitive; checkpoints are per-event durable by default; the system's ONE projection worker is named "projections" (NOT the collection name — first attempt compared the wrong name and WARNed)                                                                                                                                                                                                                     |
| ~~ | 8                                                                                                                                                                                                                                                                                | NEW internal/cqrs/visibility.go: `streamVersions` cache + `committedTail` (LoadFromVersion tail read) + `awaitVisibility` (poll ProjectionHost().Status() until worker "projections" checkpoint ≥ our last event ID; 250 ms budget; WARN + proceed on expiry — the D7 availability posture)                                                                                                                                                                                                                                                                         |
| ~~ | 9                                                                                                                                                                                                                                                                                | Snippets backend Save/Delete wired to the await — the Phase-3 latent race is FIXED, pinned by TestSnippetSaveIsImmediatelyVisible (Save returns ⇒ List already sees the write; passes in 0.01 s, no WARN)                                                                                                                                                                                                                                                                                                                                                           |
| ~~ | 10                                                                                                                                                                                                                                                                               | NEW internal/cqrs/engineaggregate.go: `engineCount` via `TimerEngine().(metaengine.AggregateReader)` (the System's only engine exposure; sqlite implements it) — the metrics totals' new home                                                                                                                                                                                                                                                                                                                                                                       |
| ~~ | 11                                                                                                                                                                                                                                                                               | A26 contacts aggregate code-complete: internal/cqrs/contacts.go — per-owner stream "contacts", state map[phone]slot, cap 500 + rename-keeps-id/createdAt + delete-by-id resolution ALL in the decision, never-sweep structural (no trim events), ContactView projection byte-parity, backend implements readmodels.ContactGate (Save/List/Count/Delete) with D7 await on Save/Delete                                                                                                                                                                                |
| ~~ | 12                                                                                                                                                                                                                                                                               | A27 readmodels gates + exports: TranscriptGate (9 methods) + NotesGate + ReviewsGate + ExportAll ×4 (the importer sources)                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ~~ | 13                                                                                                                                                                                                                                                                               | A27 transcripts aggregate written: internal/cqrs/transcripts.go (stream per (owner,call) via TranscriptStreamKey of owner then callID; per-CALL floor 1500 in the DECISION as an append+expunge batch; commands transcript.append/.expunge/.delete_call; projection mirrors the legacy query incl. WithCap + WithCapPerGroup) + internal/cqrs/transcriptsstore.go (backend: Append + EnforceCaps for the 5000/owner projection trim, the four reassembly reads, DeleteCall with decision-authoritative count via `dispatchCounting`, Count, converging SweepBefore) |

## b) PARTIALLY DONE

| #  | Item | What exactly remains           |
| -- | ---- | ------------------------------ |
| ~~ | 1    | A25 Phase-3 closeout           |
| ~~ | 2    | A27 transcripts aggregate      |
| ~~ | 3    | A28 notes + reviews aggregates |
| ~~ | 4    | Phase-4 wiring                 |

## c) NOT STARTED

| #  | Item |
| -- | ---- |
| ~~ | 1    |
| ~~ | 2    |
| ~~ | 3    |
| ~~ | 4    |
| ~~ | 5    |

## d) TOTALLY FUCKED UP / honest misses

| #  | What | Consequence + state                                                                                                        |
| -- | ---- | -------------------------------------------------------------------------------------------------------------------------- |
| ~~ | 1    | The tree does NOT compile right now: transcriptsstore.go references `readmodels.ReassembleTranscripts` (unexported helper) |
| ~~ | 2    | A25's "full buildflow to green" was executed and came back RED (file-size-check ×2)                                        |
| ~~ | 3    | Phase 3 shipped the snippets same-request D7 race (the settings partial re-render)                                         |
| ~~ | 4    | First awaitVisibility implementation compared the WRONG worker name (collection name vs the system's worker "projections") |
| ~~ | 5    | I pushed bc754a39 without immediately CI-watching (watched ~an hour late)                                                  |

## e) IMPROVE (do better)

| #  | What |
| -- | ---- |
| ~~ | 1    |
| ~~ | 2    |
| ~~ | 3    |
| ~~ | 4    |

## f) Next (the ledger — SUPERB Phase 4 completion, then the plan's tail)

| #  | Item | Size                                                                                                                                                                                                                                  |
| -- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~ | 1    | Export `readmodels.ReassembleTranscripts`; repoint internal callers; vet the tree GREEN                                                                                                                                               |
| ~~ | 2    | A28: call-notes aggregate (stream per (owner,call) "call_notes"; SaveCallNote body-merge + SaveDisposition merge, snapshot facts, 2000 cap → ErrNoteListFull WrapConflictf in the DECISION; command names note.save/disposition.save) |
| ~~ | 3    | A28: call-reviews aggregate (wholesale-replace SaveCallReview; 2000 cap → ErrReviewListFull Rejection; system-written doctrine in the command shape)                                                                                  |
| ~~ | 4    | Notes + reviews backends implementing NotesGate/ReviewsGate (no await needed: 204 + payload-less nudges)                                                                                                                              |
| ~~ | 5    | Contacts + transcripts + notes + reviews test batteries (canonical shape: real system over temp dir, restart-reload, caps, semantics, D7 pin for contacts)                                                                            |
| ~~ | 6    | app/cqrs.go: four once-accessors + DomainDecls in NewCqrs; app.go Deps swap to the cqrs backends + importers wiring                                                                                                                   |
| ~~ | 7    | server.Deps + api.Deps type swaps to the four gates (admin_reads_test + owner-scoping pins must stay green)                                                                                                                           |
| ~~ | 8    | A29: four importers on appendJournalFacts + markers (contacts_journal_imported, transcripts_journal_imported, call_notes_journal_imported, call_reviews_journal_imported) + port tests                                                |
| ~~ | 9    | A29: dual-run diff harnesses (legacy engine read vs journal projection read, row-by-row) — the parity proof the plan demands per cutover                                                                                              |
| ~~ | 10   | Error-contract registry: `go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update` for the new cqrs.* codes — SAME COMMIT as the codes                                                                                      |
| ~~ | 11   | art-dupl extract-first sweep + baseline regen (CI-pinned v0.7.2) + dedup-registry sweep row                                                                                                                                           |
| ~~ | 12   | `nix fmt` + full gates (`go test -count=1 ./...`, session suite, buildflow no-cache CLEAN GREEN, nix flake check) — the A25 debt closes here                                                                                          |
| ~~ | 13   | Smoke via `nix develop -c` (contacts/transcripts/notes/reviews scenarios still pass byte-identically)                                                                                                                                 |
| ~~ | 14   | CHANGELOG [Unreleased] entry; plan Errata: Phase-4 as-built notes (D7 mechanism, worker name, per-call floor in decision, EnforceCaps posture, migration twins)                                                                       |
| ~~ | 15   | A30 release train per D4: version bump in flake.nix, narrative commit at the phase boundary, push + CI-watch to green; stack re-pin coordination note                                                                                 |
| ~~ | 16   | Phase 5 design doc (A31): per-thread streams keyed (owner, network, remote); AppendToThread as the decider apply; OCC conflict-retry spec                                                                                             |
| ~~ | 17   | A36 upstream dependency: go-cqrs-lite#118 (TrimStream / journal trim) — MUST ship + bump before the Phase-5 deletion gate or the §4.9 P5 abort fires                                                                                  |
| ~~ | 18   | Phases 5–7 remainder per the plan's §f ledger (messages core, fax, handler unification, contracts, dedup migration + readmodels.db death certificate, sessions ruling, ladder freeze, doctrine docs, final release)                   |

## g) Questions for the owner

~~1. **A36's upstream blocker (go-cqrs-lite#118, journal trim) — where do I spend the effort?** The plan says the GDPR-delete gate (A36) cannot ship without an upstream TrimStream-style capability that does not exist yet. Options: (a) I implement it in go-cqrs-lite and release (I own that repo, but it is a library change with its own review burden), (b) defer A36 to the end of Phase 5 and land the rest of the phase, (c) park Phase 5 entirely until upstream ships. Bare "continue" = (a) when I reach it.~~ v: verified: moot — #118 shipped upstream; option (a) never needed
~~2. **D4 release cadence, re-confirmed:** A09 ruled one release per phase; the previous session's report re-opened it (per-phase release vs one release at P7). A30 expects a release train. Cut a real version bump + push per phase, or defer all releases to P7? Bare "continue" = per-phase release per the standing D4 ruling.~~ v: verified: answered by events — the consolidated v2.14.0 one-tag posture superseded per-phase releases
~~3. **`.buildflow-baseline.json` reappeared** (46 accepted-noise findings) — AGENTS.md records it RETIRED 2026-10-28 ("suppressed nothing and looked reviewed"). Did you (or a concurrent session) intentionally re-create it, or should the retirement be re-executed and the 46 findings re-triaged into .buildflow.yml classifications?~~ v: verified: resolved by events — the file exists today as the branching-flow baseline; AGENTS documents the 2026-10-28 retirement and the moved-checkout re-save recipe (upstream #13)
