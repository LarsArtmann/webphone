# The core streams design (SUPERB A31 — threads / messages / attachments)

Status: DESIGN (the map for A32–A38). Surface facts verified against the
tree 2026-11-08 (the p5 surface map); upstream prerequisites are SHIPPED:
journal trim (`System.TrimStream` / `TrimJournalBefore`,
go-cqrs-lite #118, metaengine v4.23.0 / sqliteengine v4.8.0 /
system v4.18.0) and the EnforceCap lock-order fix (#123).

## 1. Stream topology

ONE stream per thread, keyed by the thread's identity axis
**(owner, network, remote)** — the same key the legacy `threads` table
unique-indexes (migration v5).

- Stream type: `"thread"`.
- Stream key: `ThreadStreamKey(owner, network, remote) →
  owner + "|" + network + "|" + remote`. The separators are safe because
  `owner` is an extension (digits) and `network` is a closed enum, but
  `remote` is NOT separator-safe (matrix `@local:server` contains `:`;
  phones contain `+*#`). Ruling: **hex-encode the remote** in the key
  (`hex(remote)`), keeping the key greppable for owner+network and unique
  for every remote shape: `100|sms|2b343930…`. `id.ParseStreamID` accepts
  it (arbitrary non-empty strings; the contacts/dial precedents already
  key by owner string). The ThreadID itself stays the 21-char nanoid the
  legacy rows carry — the STREAM key and the ROW identity are separate
  axes, and the importer must map (owner, network, remote) → existing
  ThreadID so deep links and `thread-<id>` wrapper ids survive byte-parity.
- Facts on the stream (fact-name continuity: these tables were never in
  readmodels, so the names are NEW but follow the legacy vocabulary):

| Fact                      | Payload (domain-typed)                     | Decided by         |
| ------------------------- | ------------------------------------------ | ------------------ |
| `ThreadStarted`           | first message (full `domain.Message`)      | `message.append`   |
| `MessageAppended`         | full `domain.Message` (attachments inside) | `message.append`   |
| `MessageDeliveryAdvanced` | messageID + `domain.Delivery`              | `message.advance`  |
| `ThreadFlagChanged`       | flag + desired                             | `thread.set_flag`  |
| `ThreadUnreadCleared`     | —                                          | `thread.mark_read` |
| `ThreadDeleted`           | reason (retention \| owner)                | `thread.delete`    |

`ThreadStarted` and `MessageAppended` carry the same payload type; the
distinction is structural (state was empty), so the projection folds both
with one handler and the decider emits `ThreadStarted` only on the empty
state — the list fold's "created" ordering comes from the first event
either way.

## 2. The decider (A32): AppendToThread moves IN

`threadState` folds the full aggregate:

```go
type threadState struct {
    ThreadID   string // 21-char id; adopted from the first message (importer carries legacy ids)
    Network    domain.Network
    Remote     string
    LastActivity int64 // nanos
    Unread     int
    Pinned, Archived, Muted bool
    Outbound   map[string]outboundSlot // id → {status, providerRef, failureKind, failureDetail}
}
```

- `message.append` runs **`domain.AppendToThread` as its apply** — the ONE
  legality home survives verbatim (plan §3): creation adopts the message's
  identity; appends bump LastActivity/Unread; inbound AUTO-UNARCHIVES.
  The decision then emits the fact(s); caps: NONE today (threads are
  uncapped in the legacy store — the decider adds none).
- `message.advance` consumes **`deliveryMoves` + `DeliveryAdvance`** as the
  legality table (out-of-order verdicts legal from every state — the table
  stays the ONE home; tightening stays owner call A1.7). Same-status
  verdicts are legal refreshing moves (replay dedup is `hooksIdem`'s job,
  unchanged). A delivered verdict CLEARS failure kind+detail; a failed
  verdict records them (legacy `UpdateOutboundStatus` semantics, verbatim).
- Self-send (`gateway.SelfSendRejection`) stays REJECTION AFTER PERSIST —
  `messaging.Service.Send` keeps persisting (dispatch) then refusing the
  gateway arm (422) exactly as today; the seam is untouched by the swap.
- `thread.set_flag` / `thread.mark_read` / `thread.delete` are thin
  decisions (existence check → fact; mark-read on unread=0 is a legal
  zero-event dispatch — the Phase-3 inert contract).
- `thread.delete` is the RETENTION tombstone: legal from every state; the
  decider does NOT delete blobs (sweep owns files) and does not trim the
  journal (the sweep orchestrates TrimStream after fold visibility).

## 3. Projections (A33): same shapes, same ids

Three metaengine collections in system.db (views are new structs in
`internal/cqrs`; the server keeps reading through gate interfaces):

| Collection            | Key          | Shape                                                       | Folds                                                                                                                                                                                         |
| --------------------- | ------------ | ----------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `threads`             | threadID     | `ThreadSummaryView` (all summary fields)                    | MessageAppended upsert (LastActivity/Unread/lastBody/lastDirection/lastChannel/lastAttachment flag from attachments>0), ThreadFlagChanged set, ThreadUnreadCleared zero, ThreadDeleted Remove |
| `thread_messages`     | messageID    | `MessageView` (full message + owner+threadID + SortKey)     | MessageAppended upsert, MessageDeliveryAdvanced status/providerRef/failure patch, ThreadDeleted Remove                                                                                        |
| `message_attachments` | attachmentID | `AttachmentView` (messageID, owner, name, mime, size, path) | MessageAppended upsert per attachment, ThreadDeleted Remove-per-thread (fold queries by threadID? NO — Remove needs the key; see note)                                                        |

Attachment removal on ThreadDeleted: the projection Remove is keyed by
the collection key, so `ThreadDeleted`'s folds for `thread_messages` and
`message_attachments` need the KEY SET. Two legal shapes: (a)
`metaengine.RemoveWhere` if the engine has it; (b) the fold is a
remove-by-filter. Verified: metaengine projections support
`OnRecord(…, metaengine.Remove[V]())` keyed removals only — so the
retention SWEEP (not the fold) removes message/attachment rows: the sweep
reads the thread's message ids before dispatching delete and issues the
removals through the store's MapDelete lane... **NO — that breaks replay**
(a reset+replay would resurrect messages of deleted threads). CORRECT
shape: `ThreadDeleted` carries the full id set it deletes (messageIDs +
attachmentIDs gathered from the aggregate state — bounded by the thread's
own history; the DECISION knows every message id it ever saw via
`Outbound` + an `AllMessages []string`... that makes threadState heavy.
**RULING: the decider's state carries `Messages []string` (ids only)** —
a phone thread is hundreds, not millions; ThreadDeleted embeds the three
id lists; every fold is a pure keyed Remove; replay is exact. Snapshot
strategy (D8, below) bounds the fold cost for long-lived threads.

Sort/stability: `MessageView.SortKey` = `CreatedAt.UnixNano()` (int64 —
exact under CBOR, per the Phase-3 ② ruling the payload timestamp IS the
order authority and the event sequence breaks ties — the reader sorts
`SortKey` desc; equal keys stay in fold/journal order (stable engine sort
pinned by the A37 battery; the importer appends in original rowid order
so imported ties keep their legacy order).

Search: `SearchThreads` cannot push down (engine filters are Eq/Ge/Count
— no LIKE). Ruling: the backend scans the owner's ACTIVE threads
(owner-filtered reader, the same page ListThreads renders) and filters
in-memory with the legacy semantics (ASCII case-insensitive contains over
remote and lastBody... **legacy searches EVERY message body**, not just
the last). The in-memory twin therefore scans the owner's
`thread_messages` too (owner-filtered, bounded by the owner's own
history; single-tenant human scale — measured against the P2 perf
baseline in A38). LIKE metacharacters: the legacy `likeEscape` made them
literal — a `strings.Contains(strings.ToLower)` twin matches that
contract exactly.

Admin seams: `ListAllThreads` / `GetThreadAny` / `AttachmentByIDAny`
become unfiltered reader scans (or key-direct gets) on the SAME
collections — they stay the only cross-owner reads, behind
`deps.Admin`, pinned by the parity battery.

## 4. The service swap (A34): messaging.Service dispatches

`messaging.Service` keeps its PUBLIC contract (RegisterSender, Receive,
Send, flags, compose-new routing, ChangeFunc) — the internals swap:

- `Receive` (webhook lane + matrix/signal receivers): builds the
  `domain.Message`, resolves the stream key, dispatches `message.append`,
  awaits fold visibility (D7 — the SSE push re-reads the list), emits the
  SAME `ChangeFunc` values post-append (threads list-only nudge rules
  intact: zero thread id skips the transcript push).
- `Send`: resolves network by address shape (`ParseComposedAddress` —
  untouched), dispatches `message.append` (outbound), THEN the gateway
  arm (self-send rejection after persist preserved). Provider ref
  patching rides `message.advance` via the webhook lane —
  `MessageByProviderRef` becomes a reader lookup over `thread_messages`
  (filter providerRef+direction; single-row expectation).
- Flag toggles + mark-read: dispatch + await (row buttons re-render from
  the projection in the same request; head buttons morph `#wp-thread-head`
  — both read paths need D7 await, matching contacts).

## 5. The importer (A35)

`internal/app/coreimports.go`, the established marker discipline:

- Marker `threads_journal_imported`; legacy reads ordered threads →
  messages per thread (`ORDER BY created_at, rowid`) → attachments.
- Per thread: dispatch `message.append` per message **carrying the legacy
  ThreadID and CreatedAt in the payload** (timestamp fidelity §4.4: list
  ordering folds payload stamps; journal seq = import order = legacy
  rowid order for ties). Attachments ride their message's fact.
- Crash-safe: marker only after EVERY thread dispatched; re-run retries
  the remainder. A partially-imported thread re-dispatches from its
  legacy rows — `message.append` is NOT idempotent on message id... the
  decider must reject a duplicate message id (state knows `Messages`) —
  **the decider rejects message ids it already carries** (Rejection), so
  importer retries are safe AND webhook replays double-append safely.
- Dual-run harness (test-only): row-by-row diff legacy store reads vs
  projection reads (every owner, every thread, every message+attachment,
  summary fields) — part of the battery, not shipped code.
- Pre-cutover snapshot: the release train's backup oneshot covers
  webphone.db; A35 records a manual pre-cutover backup in the runbook
  commit message (rollback = revert + restore).

## 6. Retention (A36): tombstone → fold → trim → blob

The sweep (port of `store/sweep.go` semantics; the file retires at A45):

1. `retention_days = 0` → no sweep (keep-forever, config-validated
   upstream of the seam — unchanged).
2. Read threads with `LastActivity < cutoff` from the PROJECTION
   (owner-by-owner; threads never swept for contacts-style classes —
   threads ARE the swept class per D3).
3. Per thread: dispatch `thread.delete` (decision emits ThreadDeleted
   with the id sets; folds Remove thread row + message rows + attachment
   rows).
4. Await fold visibility, then **`sys.TrimStream(ref)`** — the shipped
   upstream primitive erases journal rows + snapshot atomically (bytes
   GONE; GDPR arm).
5. Blob deletes: the existing file-deletion logic moves to the sweep
   pass (paths from the removed attachment views, gathered BEFORE the
   dispatch).
6. THE GATE (test-pinned, plan §4.2): after a sweep with
   `retention_days=1`, a message older than the cutoff is GONE from
   (a) the journal (`EventStore().Load` empty / no rows),
   (b) every projection (threads, messages, attachments),
   (c) the blob dir, AND (d) a fresh projection reset+replay rebuilds
   WITHOUT the deleted content. `TrimJournalBefore` is NOT used by the
   core sweep (TrimStream per closed thread is the precise tool; the
   age-sweep primitive stays available for later windowed classes).

## 7. Read-your-writes (D7) + snapshots (D8)

- Every same-request re-read site (A34 list) uses the established
  `awaitOwnWrite` helper keyed on the thread's own stream — the
  projection worker name and checkpoint compare are the Phase-2/4
  mechanism, unchanged.
- Snapshots: `RegisterDecider[threadState]` with
  `system.WithSnapshotStrategy(snapshot.EveryNEvents(500))` — D8's
  conservative start; tune from A38's perf data. ThreadState carries the
  id lists (§3) so a snapshot restore is complete; TrimStream co-trims
  snapshots (shipped), so retention and snapshots cannot disagree.

## 8. Perf budget (A38)

P2 baseline (A19): dispatch 805 µs/op, deny 30 µs, projection read 76
µs/op. The core's budget: `message.append` dispatch p95 < 2 ms (single
event, single-digit-µs decision), ListThreads read p95 within 2× of the
legacy SQL read (the engine scan replaces an indexed SQL read; measured,
not assumed), transcript read p95 within 2×. A38 records the numbers
against the baseline; regression beyond budget = remediate twice then
abort per §4.9.

## 9. Execution order (A32–A38)

1. A32 decider + delivery machine (+ battery: append/advance/flags/read/
   delete legality, duplicate-id rejection, scoping, restart).
2. A33 projections + backend gates (readmodels `ThreadGate`-class
   interfaces; server/api Deps swap; byte-pins green).
3. A34 service swap + ChangeFunc parity (island tests + SSE pins).
4. A35 importer + dual-run diff + snapshot record.
5. A36 sweep + deletion gate + keep-forever pin.
6. A37 parity battery: owner_scoping, threadsdual, threads_flags,
   byte-pins, `TestProviderFormPartHeaderBlockGolden`, webhook pins.
7. A38 gates: full smoke, stack browser E2E (445 s), perf re-check,
   release train per D4.
