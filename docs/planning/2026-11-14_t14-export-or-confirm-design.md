# §5 export-or-confirm tooling — design memo (no production code)

**Date:** 2026-11-14 · **Executes:** the death-map correction addendum (2026-11-13) and the TODO_LIST row "Design the §5 export-or-confirm tooling before any DROP window" · **Gates:** the ladder's final DROP migration (storage-ladder §5), operator-gated.

## 1. What the §5 DROP would delete

`webphone.db` `contacts` (v1 baseline, `internal/store/db.go:218`): `id TEXT PK, owner, name, phone, created_at INTEGER (unix SECONDS), UNIQUE(owner, phone)`. Three facts make this the dangerous DROP: the reader is GONE (`4fb24b61` deleted `ReadLegacyContacts` with its importers), NO retention sweep covers the table (`store.Sweep` deletes only messages/threads — contacts are deliberately out of scope, `sweep.go`), and the one-time boot import (deleted `internal/app/legacy.go`) mapped rows oldest-first through `Save`, SKIPPING rows whose owner or phone did not parse and minting ids where the id did not. Legacy-only rows would vanish silently.

## 2. Read-path decision: raw read-only sqlite, NOT a resurrected reader

The dump opens `webphone.db` with the modernc driver in read-only mode (`file:...?_pragma=query_only(1)`), never via `store.Open`: `store.Open` runs the migration chain and stamps WAL — a dump binary at a different schema version could EXECUTE the §5 DROP step before dumping anything. The tool must be inert at ANY point in the version chain, including after v6 (the DROP) lands. Do NOT resurrect the deleted legacy reader (dead class per the correction); the tool owns its own SELECT against the v1 shape. The new-world side reads through the normal composition (the `contacts` journal collection) — read-only too.

## 3. Tool sketch (CLI shape only)

New pre-server one-shot slot in `cmd/webphone/main.go`, beside `-enroll-passkey` / `catalog` / `contracts` (no server wiring, exits after output):

- `webphone export-legacy -table contacts -format json|csv [-data DIR]` — dumps every row (id, owner, name, phone, created_at raw + ISO), stdout; emits a self-describing header: row count + sha256 + timestamp.
- `webphone export-legacy -check [-confirmations PATH]` — classifies legacy rows against the journal projection by the digit-string key (`domain.DigitsOf`, the aggregate's upsert axis) and prints: legacy total, in-both, legacy-only, skipped/unparseable, new-world total. Exits NONZERO while legacy-only rows exist without a confirmation file covering them — the machine gate.

`-table` is extensible: contacts is the pilot; snippets/transcripts are the same unswept-imported class (the §5 obligation is family-wide per the correction); the threads/messages family keeps its intact retention-empty-then-drop posture.

## 4. Per-row-class confirmation flow

- **A — in both worlds** (digit-key match): inert; the DROP deletes no unique data. Listed in the report, no action.
- **B — legacy-only, parseable**: named per owner in the dump (name, phone, created). Either the operator re-imports (re-run of the import path or a manual re-save in the UI — the row then re-classifies as A) or the owner confirms the loss IN WRITING: one record per owner naming its rows by dump line reference, archived beside the dump with the dump's sha256.
- **C — legacy-only, unparseable** (skipped at import; can never auto-migrate): same named-in-dump + written confirmation, or the operator repairs the row by hand in the UI first.
- **D — new-world-only**: never deleted by the DROP; context only.

## 5. The precondition checklist gating the DROP

1. Dump archived per deployment: JSON output + sha256 + timestamp recorded in the deployment record. Fresh installs archive the zero-row dump as proof of absence.
2. `-check` classification run; every legacy-only row (B and C) accounted for: re-imported, or covered by a per-owner written confirmation.
3. The legacy side is FROZEN (no writer since the importer died in `7122de92`), so dump-time numbers hold at window time; only the new-world side can move (users deleting contacts is legitimate). Re-run `-check` immediately before the window; the invariant is: **new-world count ≥ legacy count − confirmed-losses**, and in-both ≥ the archived dump's in-both.
4. The hand-authored §5 migration step cites the dump hash + confirmation records in its comment; one transaction, one version bump — the ladder's last act, after which `store/db.go` shrinks to the sessions schema.
5. Confirmation records live OUTSIDE the databases (deployment record / signed checklist): `app_flags` is machinery-and-cursors only, and a confirmation flag would outlive its meaning.

## Follow-up implementation row for TODO_LIST

- [ ] **Implement the §5 export-or-confirm tool** (Medium): a pre-server one-shot in `cmd/webphone` — `webphone export-legacy -table contacts -format json|csv` (raw read-only sqlite SELECT, sha256 + count header, archived per deployment) plus `-check` (classifies legacy rows against the journal projection by digit-string key; in-both / legacy-only / unparseable / new-world totals; nonzero exit while unconfirmed legacy-only rows exist). Gates the §5 DROP per the checklist in docs/planning/2026-11-14_t14-export-or-confirm-design.md (§4 per-owner written confirmations, §5 counts invariant). Contacts first; `-table` extends to the same-class snippets/transcripts. Evidence: storage-ladder correction addendum; memo 2026-11-14.
