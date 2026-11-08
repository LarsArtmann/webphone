# ES migration importer strategy (A08)

> **Date:** 2026-11-08 · **Plan:** [SUPERB command & event sourcing](2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 0 · **Precedent:** the 2026-10-15 legacy-contacts import (`internal/store/legacy_contacts.go` + `readmodels` boot import, schema-v4 `app_flags` markers)

## 1. The pattern (extracted from the shipped precedent)

1. **Marker**: one `app_flags` key per aggregate (`es_import_dial_windows`, `es_import_contacts`, …) — stamped ONLY after FULL success; a crash mid-import retries on next boot (never flags partial).
2. **Read old rows oldest-first** (`ORDER BY created_at ASC, rowid ASC` semantics) — order preservation.
3. **Parse into domain types at the door**; skip+log corrupt rows (a corrupt legacy row must not brick a boot).
4. **Write to the new store**, then flag.

## 2. What changes for a journal (the two new problems)

### 2.1 Append is not idempotent by key — resumability must be STRUCTURAL

Engine upserts dedup by key; journal appends do not. The importer is therefore **resumable by construction, not by marker alone**:

- **Deterministic stream IDs** per aggregate (dial: the existing `DialWindowKey{owner, remote}` string; contacts: owner-scoped contact stream; threads: the thread id). The importer derives the SAME stream id on every run.
- **Per-stream tail sync**: `EventStore().Load(ctx, ref)` → the stream's current event count IS the resume cursor; append only the missing versions, in deterministic order (source row order). A crash anywhere converges on re-run with zero duplicates — no transaction gymnastics, no two-phase marker.
- **Marker stamps the aggregate's LAST stream** — after that, the boot path skips the importer entirely.

### 2.2 Timestamp fidelity — journal order is NOT list order

Folds must order by PAYLOAD stamps (`created_at`, `last_activity_at`, `At`), never by journal append order or `OccurredAt`. This is exactly how readmodels folds behave today (`SortOnField("at", …)`), so the rule is: **import events carry the ORIGINAL row timestamps in their payloads; the projections' existing sort declarations keep ordering them correctly** (B72/B128 pin it per aggregate). `OccurredAt` = import time; nothing user-visible may read it.

## 3. The importer shape (one per aggregate phase, `internal/migrate`)

```go
// RunDialImport: reads readmodels.db rows (OLD engine, read-only), tail-syncs
// synthetic events into system.db, awaits projection catch-up, diffs, flags.
func RunDialImport(ctx, old *readmodels.Engine, sys *system.System, flags FlagSink, log) error
```

- **Synthetic events, raw append** (not command dispatch): `event.New(factType, streamID, streamType, nextVersion, payload)` with metadata actor `system:importer` + `imported_from: <aggregate>`. Deciders stay decision code; the importer is state TRANSPLANT, and dispatch would route rows through decisions that were already made.
- **Runs in `app.New` BEFORE `sys.Start`** — the projection host's initial journal drain folds the imported events exactly once; no live-phase race, no await needed at boot.
- **Event types = today's fact names** (`DialAttempted`, `ContactSaved`, …) — the same `metaengine.EventTypeName` strings the folds already consume (A01: adapter routes by that name).
- **Payload types = today's domain fact structs, verbatim** — one home per fact (plan §3 fact-name continuity).

## 4. The dual-run diff harness (gate, not tooling)

After import (and per B71/B106–B108/B128), for EVERY owner:

1. read the OLD surface rows (old engine/store),
2. read the NEW projection rows (TypedReader over `sys.MetaEngine()`),
3. compare row-by-row as VIEW STRUCTS (byte-compare of the Go values — the view types are shared),
4. zero diffs ⇒ cutover proceeds; any diff ⇒ the phase train STOPS before the port swap (the old surface stays live — the importer is additive until the swap).

The harness lives as a `go test` per aggregate phase (dual-run test) plus a one-shot boot log line (`migrate: dial diff owners=N rows=M diffs=0`). The pre-cutover backup snapshot (B129) is taken by the OPERATOR runbook step, recorded in the phase commit.

## 5. Per-aggregate notes

| Aggregate       | Old source                       | Exemptions / notes                                                            |
| --------------- | -------------------------------- | ----------------------------------------------------------------------------- |
| Dial windows    | `readmodels.db` collections      | first importer (A18); overrides consumed-state replays as `Consumed:true` events |
| Snippets        | `readmodels.db`                  | cap 100 enforced by decider on NEW writes; import may exceed → converging trim on next save (posture: trim, documented) |
| Enroll tokens   | —                                | **EXEMPT** (15-min TTL, in-flight links re-mint — the 2026-10-15 precedent)   |
| Agent audit     | —                                | **EXEMPT** (starts empty; journal IS the audit from Phase 3 on; the readmodels dual-write row stays until P7) |
| Push devices    | `readmodels.db`                  | token stays write-only (importer reads it once server-side; LIST never echoes) |
| Contacts        | `readmodels.db`                  | rename-keeps-id semantics fold from `ContactSaved` events with ORIGINAL ids   |
| Transcripts     | `readmodels.db`                  | per-(owner, call) streams; floor loop unchanged                              |
| Notes/reviews   | `readmodels.db`                  | one-slot-per-call folds                                                       |
| Threads/messages/attachments | `webphone.db` hand ladder | TWO-part importer (B125/B126): threads → stream heads, then messages+attachments → events in (created_at, rowid) order; attachments events carry blob paths (bytes never move) |
| Faxes           | `webphone.db`                    | `received` inbound facts replay as facts; lifecycle statuses as `faxMoves`-legal events |

## 6. Invariants

- Importers are TRANSITIONAL: one file per aggregate, deleted at ladder/file retirement (the legacy_contacts precedent: "exists to be deleted").
- The boot NEVER blocks on import failure beyond retry-on-next-boot: a failing import logs ERROR + names the aggregate; the OLD surface stays live (additive-only until each swap) — no brick, no silent split.
- Import reads are OWNER-SCOPED loops (all owners, each scoped) — no cross-owner seams added.
- The flag sink is webphone.db's `app_flags` (D2: markers stay there until ladder retirement; system.db never grows an import-marker table).
