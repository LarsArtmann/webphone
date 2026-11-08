# Journal retention + GDPR policy (A05)

> **Date:** 2026-11-08 · **Plan:** [SUPERB command & event sourcing](2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 0 · **Feeds:** D3 (retention classes), D8 (snapshots), §4.2 invariant, P5 gate B130–B132

## 1. The invariant being ported (today's semantics, `internal/store/sweep.go`)

`retention_days` (0 = keep forever, N = delete content older than N days) drives, per sweep pass:

- messages: `DELETE FROM messages WHERE created_at < cutoff`
- threads: `DELETE ... WHERE last_activity_at < cutoff AND NOT EXISTS (remaining messages)`
- attachments: blob files whose rows vanished (paths collected first, files deleted after)
- **contacts: NEVER swept** ("the user's address book, not transient content" — sweep.go's own words)
- counts remain aggregate table sizes for `/metrics`

Faxes have NO sweep today (B144's "fax retention (if configured)" is a future option, not a port).

## 2. Retention classes under event sourcing

| Class                     | Content                                             | Journal posture (D3 draft)                                | Mechanism                                                                                                                                |
| ------------------------- | --------------------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| GDPR-delete               | messages, threads, attachments (the swept set)      | `retention_days` deletes from JOURNAL + PROJECTION + BLOB | tombstone event → Remove fold (projection row gone, exists today) → **stream trim** (bytes gone, upstream gap §3) → blob delete (exists) |
| GDPR-delete (user action) | a deleted contact; an expunged transcript set       | same as above, event-initiated instead of tick-initiated  | same chain; contact deletion already carries a Remove fold                                                                               |
| Cap-trim                  | dial windows, transcripts                           | trim WITH the view (transcripts posture: expunge oldest)  | projection caps already trim views (`WithCapPerGroup`); journal-side ride the SAME §3 trim pass by age                                   |
| Reject-capped             | contacts 500, snippets 100, reviews 2000, devices 8 | keep-forever journal (bounded by decider-side rejects)    | no trim; the cap IS the bound                                                                                                            |
| Audit                     | agent audit                                         | keep-forever BY DESIGN (owner ruling D3)                  | none                                                                                                                                     |
| Never-sweep               | contacts (tick sweeps)                              | unchanged doctrine                                        | tick pass never touches contact streams                                                                                                  |

## 3. THE FINDING: journal trim does not exist upstream (the hard prerequisite)

Verified against the PUBLISHED pins (event/v4 v4.15.2, system/v4 v4.17.0, metaengine v4.22.0, sqliteengine v4.7.0):

- `event.Store`/`Journal`/`SeekableJournal` interfaces carry NO delete/trim/purge surface (store.go:101–135).
- `event`'s tombstone is a **metadata mark** (`TombstoneMark` on stream metadata; replay folds skip tombstoned streams) — semantic deletion, NOT byte deletion. Payload bytes stay.
- sqliteengine's `DELETE`s are ALL projection-side (caps, map deletes, snapshots, graph, reset) — nothing touches the stream log.
- Upstream doctrine: append-only, tombstone-over-delete (core.md §3.1 / ADR-0114).

Consequence: **§4.2's "delete actually deletes (journal + projection + view)" cannot be met with published pins.** Lane selection per the I/O rule's A/B gate: this is lane **(B) the capability should exist** — design WITH the owner, land it ONCE in the engine family (never satellite SQL beside it, which would be E023 and a fork of the substrate).

### The upstream capability spec (to file + design with the owner)

`metaengine` engine-family extension, sqliteengine first:

1. **`TrimStream(ctx, streamRef, keepHead)`** — GDPR erasure: delete a stream's event rows (optionally keeping the head event or a stream-tombstone marker), CO-TRIM its snapshot row (snapshot DELETE already exists engine-side), leave projection checkpoints intact (they are position-based, stream-agnostic).
2. **`TrimJournalBefore(ctx, cutoff)`** — the age-based sweep primitive for cap-trim classes: delete event rows older than cutoff FOR streams marked trimmable (a per-stream-class registry, or tombstoned streams only — design call with the owner; tombstoned-only is the conservative v1).
3. Replay consistency REQUIREMENT: a trimmed stream must fold to NOTHING (no partial resurrection); a trimmed+snapshot-only stream must be rejected loudly (snapshot without events = corrupt state) — trim and snapshot delete are ONE transaction.
4. Plan-drift/scream rules unchanged; trim is an operator-invoked maintenance method, never a read/write-path behavior.

Sequencing: **file the upstream issue in the Phase 1 train** (go-cqrs-lite issue; same-owner precedent: #114 `WithCapPerGroup`, ADR-0164, shipped in days). The capability must be released + bumped into webphone **before Phase 5's deletion gate (B131)**. If it cannot land: the §4.9 P5 abort criterion fires (core stays CRUD, documented, already-migrated aggregates keep streams) — a pre-committed outcome, not a failure.

## 4. The sweep port (Phase 5, `internal/retention`)

`retention.Start` grows a cqrs leg (injected alongside db.SQL/blobs/engine):

```
per sweep tick (unchanged cadence):
  1. hand-ladder sweep            — UNCHANGED bytes until ladder retirement (P7);
                                    after Phase 5 the threads/messages tables are
                                    import-only history, swept to empty then inert
  2. ES tombstone pass            — for threads whose last_activity < cutoff and no
                                    surviving messages: emit ThreadDeleted (tombstone
                                    event, Remove folds already exist) via the
                                    retention service's own dispatch identity
  3. journal trim pass            — TrimJournalBefore(cutoff) over tombstoned
                                    thread streams (upstream §3 primitive)
  4. blob delete                  — unchanged path (paths from the tombstoned set)
  5. counts                       — projection-side aggregates, same metric names
```

Deletion gate (B131, test-pinned BEFORE core cutover): after a sweep with `retention_days=1` and content older than the cutoff — journal rows for the thread's stream GONE (`EventStore().Load` returns empty), projection row GONE (reader Get misses), blob file GONE, AND a fresh `ResetProjection` replay rebuilds WITHOUT the deleted content (replay-consistency requirement §3.3). `retention_days=0` keeps every byte (B132).

## 5. Snapshot thresholds (D8 draft)

- **Thread streams** (the only unbounded-lived streams): `snapshot.EveryNEvents(200)` composed with `WithReadPressure` (`WithInnerStrategy`) — long threads snapshot on writes AND hot threads on reads; 200 is conservative (typical SMS threads hold < 200 messages; long-lived business threads get one snapshot per 200 appends ≈ kilobytes).
- **Dial-window / small-aggregate streams**: NO snapshot strategy — streams are tiny and short-lived; `RegisterDecider` wires the snapshot store for reads but no strategy writes.
- Review at P5 perf data (B120/B138): raise N if snapshot write amplification shows; lower if decider load latency exceeds the P2 baseline budget.

## 6. What this policy does NOT change

- Contacts never-sweep (carried verbatim; contact streams are reject-capped, keep-forever journal).
- The 2026-11-04 transcript/dial-window trim POSTURES (expunge-oldest) — preserved by class mapping.
- The dedup windows (`hooksIdem`/`CallsIdem`, 1 h TTL) — they ride `DedupStore.Sweep` (already swept in `app.Start`), unaffected until A43 moves them onto `sys.Dedup()`.
- `readmodels.db` collection caps — unchanged until each collection's phase moves it into system.db, where the SAME `WithCap`/`WithCapPerGroup` declarations keep trimming views.
