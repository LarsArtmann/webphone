# Storage ladder retirement plan (SUPERB A44/A45)

> **Date:** 2026-11-09 · **Plan:** [SUPERB command & event sourcing](2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 7 · **Executes:** D2 (sessions + app_flags), the ladder freeze, the table→release death map

## 1. Where the ladder stands after the journal swaps

Every domain aggregate answers from `system.db` (threads, messages,
attachments, faxes, dial windows, contacts, snippets, transcripts,
notes, reviews, enroll tokens, agent audit, push devices, dedup
windows). What still writes `webphone.db`:

| Surface                | Why it is still SQL                                     | Verdict                                    |
| ---------------------- | ------------------------------------------------------- | ------------------------------------------ |
| `sessions` table       | D2's open ruling                                        | **STAYS SQL — the ruling below**           |
| `app_flags` markers    | transitional import bookkeeping                         | dies with the last marker's reader (§4)    |
| `threads`/`messages`/`attachments` | import-only history since v2.10.0          | frozen; the retention sweep empties them   |
| `fax_jobs` (retired)   | untouched since the T11 swap                           | frozen                                     |
| contacts/snippets/transcripts/notes/reviews legacy tables | untouched since their swaps | frozen                                     |

`readmodels.db` is RETIRED (A43): the composition root stopped opening
it, the backup set dropped it, and its dedup window moved to
`sys.Dedup()`. The file lingers on upgraded deployments as an inert
orphan — harmless, no longer backed up, safe to delete by hand.

## 2. The D2 sessions ruling (executed 2026-11-09)

**Sessions stay SQL.** The journal is for DOMAIN state — facts with
legality, replay value, and retention classes. A session is ephemeral
auth substrate: sliding 7-day idle / 30-day absolute windows, lazy
expiry sweeps, high-churn mint/delete with zero domain invariants and
zero replay value. Migrating it to event streams would grow system.db
unboundedly with machinery noise, buy nothing (no decider, no
projections, no retention class), and put the login hot path behind
journal appends. cqrs-lint agrees: `internal/session` mixes no
cqrs-lite usage (E023 fires nowhere near it — the three INFO sites are
the bridge packages `readmodels`/`server`/`app`, which legitimately see
both worlds as the composition seams).

The session row keeps its one SQL home until a future ruling changes
the auth substrate itself (e.g. passkey-only logins removing extension
sessions) — not until the storage doctrine changes.

## 3. The ladder freeze (A45)

`store/db.go` accepts NO new migration steps (the version freeze
comment marks it in-code). The remaining tables are import-only or
machinery; a NEW domain fact goes into `system.db` via a cqrs
aggregate, never into a new SQL table.

## 4. The death map (table → release)

| Table family                                | Dies at           | Mechanism                                                       |
| ------------------------------------------- | ----------------- | --------------------------------------------------------------- |
| `readmodels.db` (whole file)                | **v2.12.0 (A43)** | already done: boot stops opening it, backup drops it            |
| threads/messages/attachments legacy rows    | v2.13+ (operator) | the retention sweep empties them per `retention_days`; then the tables are DROP-candidates in a coordinated migration |
| `app_flags` (import markers)                | with §5           | no reader once the importers are gone (they ARE gone in v2.12.0) — the table is inert; DROP with the family below |
| fax/contacts/snippets/transcripts/notes/reviews legacy tables | v2.13+ (operator) | same retention-empty-then-drop posture |
| `sessions`                                  | never (this doc)  | stays SQL per §2 — the one deliberate hand-ladder survivor       |

Upgrade chain obligation: **v2.12.0 requires passing through v2.10.0 or
v2.11.0 first** on deployments with pre-SUPERB data — the one-time
importers lived in those releases. A fresh install needs nothing.

## 5. The final DROP migration (not yet scheduled)

When the operator confirms every deployment's legacy tables are empty
(retention swept) or expendable, ONE new migration step drops the
frozen families in a single transaction and bumps the schema version —
the ladder's last act, authored per this doc, after which `store/db.go`
shrinks to the sessions schema.
