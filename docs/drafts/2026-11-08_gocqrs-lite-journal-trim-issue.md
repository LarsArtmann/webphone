## Why

webphone is migrating its domain state onto `system/v4` (the SUPERB plan, 2026-11-08). Its `retention_days` feature is a hard product invariant: content older than N days must be DELETED — journal, projections, blobs — and a test-pinned gate proves it. The A05 spike verified against the published pins that this is currently impossible:

- `event.Store` / `Journal` / `SeekableJournal` carry no delete/trim/purge surface (event/v4 `store.go:101-135` at v4.15.2)
- the tombstone is a metadata mark — replay folds skip the stream, but the payload BYTES stay (GDPR erasure unmet for personal-data content)
- sqliteengine's DELETEs are all projection-side (caps, map deletes, snapshots, graph, reset) — nothing touches the stream log

Doctrine so far is append-only + tombstone-over-delete (core.md §3.1, ADR-0114), which is right for replay integrity and wrong for GDPR erasure + unbounded journal growth on trimmed/capped domains.

## Source-level cause

Not a bug — a missing engine-family capability (lane B of the I/O rule: the capability should exist in the substrate, designed once, never satellite SQL beside it — cqrs-lint E023 exists precisely for that workaround).

## Design

Two operator-invoked maintenance methods on the engine family (sqliteengine first), NOT read/write-path behavior:

- [ ] `TrimStream(ctx, streamRef, keepHead)` — GDPR erasure: delete a stream's event rows (keep the head event or a stream-tombstone marker), CO-TRIM its snapshot row in the SAME transaction (a trimmed stream's snapshot without its events must be impossible — that state resurrects partial state on replay; snapshot DELETE already exists engine-side, `snapshot.go:96` at v4.7.0)
- [ ] `TrimJournalBefore(ctx, cutoff)` — age-based sweep primitive: delete event rows older than cutoff for tombstoned (v1: conservative) or otherwise-marked trimmable streams
- [ ] Replay consistency: a trimmed stream folds to NOTHING; `VerifyProjections` and the plan-drift scream rules stay untouched
- [ ] Projection checkpoints unaffected (position-based, stream-agnostic)
- [ ] `systemscenario` recipe covering trim → replay → empty view

Consumer evidence for sizing the need: webphone's swept set (messages/threads/attachments) plus cap-trim domains (dial windows, transcripts) all want exactly these two calls; the alternative in every consumer is either keep-forever journals or satellite SQL, both wrong.

## To verify

webphone Phase 5's deletion gate (`docs/planning/2026-11-08_journal-retention-policy.md` §4): after a sweep with `retention_days=1`, journal rows gone (`EventStore().Load` empty), projection row gone, blob gone, AND a fresh `ResetProjection` replay rebuilds WITHOUT the deleted content. The capability must be released before that gate or the migration's pre-committed P5 abort fires (plan §4.9).

💘 Generated with Crush
