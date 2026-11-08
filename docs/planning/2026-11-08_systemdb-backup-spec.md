# system.db backup/restore integration spec (A07)

> **Date:** 2026-11-08 · **Plan:** [SUPERB command & event sourcing](2026-11-08_04-04_SUPERB-command-event-sourcing.md) Phase 0 · **Implements at:** Phase 1 (A11/A14), updated per phase

## 1. Principle

`system.db` is USER DATA the moment Phase 2 lands (dial windows) and the SOURCE OF TRUTH from Phase 5. The backup set, the drill, and the README restore path carry it with the SAME pattern as `readmodels.db`: an online `sqlite3 .backup` with an existence guard (data dirs that predate the file still back up cleanly).

## 2. Exact integration points

| Site                                    | Change                                                                                                    | When        |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------- | ----------- |
| `package/backup-script.nix`             | after the readmodels block: `if [ -f "${cfg.dataDir}/system.db" ]; then sqlite3 ... ".backup '${dest}/system.db'"; fi` — in BOTH the top-level and the dated-history branch | Phase 1 (file exists, empty of user data) — one commit, never touched again |
| `nix/vm-tests.nix` (backup drill)       | assert `system.db` present + `pragma integrity_check` ok, top-level AND snapshots branches                 | Phase 1     |
| `nix/module-check-backup.nix`           | extend the script-content assertions to include the system.db guard line                                  | Phase 1     |
| `README.md` § Backups and restore       | add `system.db` to the file list + restore steps                                                          | Phase 1     |
| `scripts/webphone-backup-drill.py`      | round-trip `system.db`; once Phase 2 lands, assert a restored boot REBUILDS the dial-window view from the journal (the A06 replay proof, in the drill) | Phase 1 guard / Phase 2 assertion |
| `docs/release-runbook.md`               | the per-train backup note names system.db from Phase 2 on                                                 | Phase 2     |

## 3. Per-phase readmodels shrink rule

Each aggregate phase REMOVES its collections from the readmodels engine plan (e.g. B65 dial, B82 snippets, B99 contacts). Consequences:

- The readmodels.db FILE stays in the backup set until A43/B152 retires it (harmless: a shrinking file, still the dedup/idem home until then).
- The backup drill does NOT change per phase (it asserts file-level integrity, not collection membership).
- **A43/B152** (engine registration removed, file retired): THEN drop the readmodels block from backup-script.nix + drill + README in the SAME train, and add a restore-path test WITHOUT readmodels.db (B153 — a restored data dir missing the file must boot byte-identically; the existence guards make this a non-event).

## 4. Restore semantics (why journal restore is stronger than table restore)

- Restoring `system.db` restores journal + projections + checkpoints + dedup + snapshots as ONE consistent sqlite snapshot (the `.backup` API guarantee, same as today's files).
- If a projection were somehow stale/corrupt in the restored file, the documented recovery is the A06-proven recipe: boot → host `Stop → ResetProjection("projections") → Start` — the worker replays the journal and rebuilds every view (measured 1.6 ms for a spike-sized journal; the drill's restored-boot assertion covers it end-to-end).
- Pre-Phase-5, restoring WITHOUT `webphone.db` is still a data loss (the hand ladder owns threads/messages until then) — README wording carries that boundary honestly per phase.
- WAL sidecar files (`system.db-wal`) are NOT backed up (same as webphone.db today): the `.backup` API captures committed state including the WAL window.

## 5. Sizing note

`system.db` grows with the journal (bounded by §3 of the retention policy: caps reject upstream, trim passes erase). The dated-history branch's `--link-dest` blob hardlinking is unaffected (blobs never live in system.db — the journal carries paths/metadata, never bytes, per plan §2).
