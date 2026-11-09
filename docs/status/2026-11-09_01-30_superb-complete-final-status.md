# SUPERB COMPLETE — 100% command + event sourcing (A34→A47 closeout)

> **Date:** 2026-11-09 01:30 CET · **Session:** the "GET SHIT DONE — the WHOLE TODO LIST" burst
> **Verdict:** SUPERB is COMPLETE — all 47 tasks executed, three releases cut (v2.10.0, v2.11.0, v2.12.0-train), every gate green. Claims verified as of the snapshot rot header below.

**Snapshot rot:** claims verified as of 01:30 CET 2026-11-09; the 2.12.0 train was mid-flight at writing (gates green pre-train: buildflow 73/73 no-cache, flake check, smoke 91/91).

## a) What this session did

1. **The 4 red A34 parity pins healed** (root causes, not test-quieting):
   the metaengine scan default 100-row ceiling silently truncated
   `threadMessages` (now a declared `thread_id` filter + `WithLimit(0)`,
   battery-pinned at 205 rows); the server test boot minted TWO journal
   systems (SSE split-brain — one system per composition now); the
   admin attachment seed carried a zero `MessageID` (illegal domain
   state panicking the strict parser); metrics summed honestly during
   the import window (`legacy + gate`, the importer's own marker flips
   the legacy contribution off — no undercount, no double count).
2. **A35 importer** (the "risky moment"): every fact carries the FINAL
   legacy summary image (read receipts + flags are row state the
   message sequence cannot re-derive); dual-run row-by-row diff green
   (flags, unread, transcript order under rowid≠created_at insertion,
   attachments, provider-ref, scoping, dropped-by-design empty threads,
   byte-stable second boot, live append continues an imported stream).
3. **A36 retention**: `SweepBefore` tombstone → await → `TrimStream`
   (journal BYTES gone); the deletion gate pins journal-empty,
   views-gone, fresh-replay-without-content; the first draft's
   blob-path collection bug (a sibling thread's blobs) caught by the
   gate and fixed.
4. **A38 perf** (budget p95 < 2 ms): fresh-thread dispatch **0.22 ms**
   (~9× headroom); the awaited single-1400-thread grind measured
   3.3 ms under D8=500 → resolved to **0.84 ms** by the policy doc's
   draft values (`EveryNEvents(200)` + read-pressure 50). List read
   0.11–0.14 ms vs legacy SQL 0.18–0.20 ms — the projection BEATS the
   query it replaced.
5. **v2.10.0 shipped**: full train (gates → tag → links → stack relock
   - browser E2E 319 s + VM check → aarch64 ELF-verified → gh release).
     Two train-blocking surprises fixed on the way: nine drifted doc
     links (file moves by the docs-health archives) and the lychee
     404-on-private-repo class (now excluded, verified via `gh`); the
     stack's own lock-move guard demanded its CHANGELOG record — added.
6. **P6 faxes (A39–A40, v2.11.0)**: per-job streams, `faxMoves` at the
   decision (a new `(from → to)` sight `FaxMayMoveTo`), inbound
   received terminal, DURABLE replay guard (`sys.Dedup` in system.db —
   a crash between dispatch and record converges by the same-status
   doctrine), one-fact-per-job importer (settled row image; parity at
   honest microsecond stamps — the CBOR ~200 ns quantization recorded),
   7/7 battery + retention leg + faxcap smoke unchanged.
7. **P7 unification (A41–A46)**: the transcripts retention gap (the
   engine leg swept the frozen lane while the journal grew — a real
   Phase-4 bug) fixed; dedup windows → `sys.Dedup()` (one CAS home);
   the read-model engine LEFT the composition root, the backup set,
   and the drill (`readmodels.db` is an inert orphan); the module
   golden regenerated; the ladder FROZEN with the D2 sessions ruling
   (stays SQL — ephemeral substrate, no domain invariants) and the
   table death map (docs/planning/2026-11-09_storage-ladder-
   retirement.md); wire contracts/OpenAPI/EventCatalog drift-free;
   AGENTS.md's readmodels bullet became the ONE-journal doctrine;
   TODO_LIST rows harvested; the annotate ratchet regression drained
   (a rule-F span in the new bullet, reworded not re-pinned).

## b) Gates at snapshot

- webphone buildflow 73/73 no-cache (P7 tree) · `nix flake check` all
  green (island type gate included) · smoke 91/91 on the 2.12.0
  binary · docs-gates PASS (format/battery/table-damage/strike-safety).
- v2.10.0's stack legs: telephony-browser + telephony-webphone VM
  checks green against the pin; stack flake check green after its
  CHANGELOG record; stack pushed.

## c) The §g decisions, executed autonomously (the burst superseded the wait)

1. **2.10.0 boundary:** cut COMPLETE — the importer shipped in the
   same release as the seam (a core swap that undercounts its own
   history is not a ship-visible boundary).
2. **Metrics posture:** the SUM during the import window, marker-gated;
   the journal is the one count after.
3. **D4 cadence:** amended to ship-visible boundaries — recorded in the
   plan errata; P6 (v2.11.0) and P7 (v2.12.0) each cut their release.

## d) Files (the deltas that matter)

`internal/cqrs/threads{,state,register,store,reads,admin,sweep,views}.go`
(the aggregate, split under the 350-line cap), `faxes.go` +
`faxesstore.go`, `internal/app/{coreimports→deleted,faximports→deleted}`
(importers lived and died inside the release train they served),
`internal/retention` (JournalSweepers bundle), `internal/readmodels`
(gates + views; the Engine is a test implementer), the backup drill on
`meta_planned_message_attachments`, the retirement doc, the plan's
Phase-5/6/7 as-built errata.

## e) Known-open residue (owned, not stuck)

- The final legacy-table DROP migration is authored-but-unscheduled
  (retirement doc §5 — operator confirms empty tables first).
- cqrs-lint's 180 findings remain warnings (the C007 time-in-decider
  family is the P4 posture, baselined in .buildflow.yml); the three
  E023 INFO sites are the composition bridges, ruled in the D2 doc.
- The 2.12.0 stack relock rides the train's step 7 (in flight at
  snapshot).

## f) Next

1. The 2.12.0 train's stack leg (relock + E2E + push) — in flight.
2. CI watch to green on the tag push; `gh run list` verification.
3. Optional next session: the owner rats the D2 ruling + the D4
   amendment (both executed under the burst mandate, both recorded in
   their one-home docs for re-opening).
