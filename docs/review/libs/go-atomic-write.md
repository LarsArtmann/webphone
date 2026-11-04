# go-atomic-write — library utilization audit

**Verdict:** the sharpest NEW finding of this audit series: webphone's one runtime file-write path persisted user data with a bare `os.WriteFile` — no fsync, no atomic rename — while this library exists precisely for that. **Adoption: LANDED 2026-11-04 (v0.6.2) — `blob.Save` rides `WriteWithPerm(…, 0o640)`, both CLI spec writers ride `WriteIfChanged`; vendorHash re-pinned in the same train.**
**Pin:** v0.6.2 · **Latest:** v0.6.2 (local checkout tag) · **Status:** direct dependency
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Errata — verified write ordering (T01, 2026-11-04)

Step-0 verification of the plan (`docs/planning/2026-11-04_18-41_SUPERB-fleet-libs-closeout.md`) confirmed the finding and sharpened the crash window:

- **Ordering is file-then-row at all four `Save` call sites**: `messaging/service.go:141→159` (outbound attachments), `messaging/service.go:210→224` (inbound attachments), `fax/service.go:118→124` (outbound spool), `fax/service.go:219→225` (inbound spool). The DB row (SQLite, fsynced WAL) lands only after `Save` returns.
- **The precise corruption window is row-committed-before-file-durable**: `os.WriteFile` returning ≠ bytes on stable storage; a power loss after the row commit but before page-cache flush leaves the row referencing a missing/zeroed/truncated file. For inbound faxes the blob is the ONLY copy — this is the user-data-loss class. A crash *between* file-create and the row does NOT corrupt referenced data (no row exists yet → orphan file at worst).
- **The report's original phrasing** ("a crash between file-create and final bytes corrupts the blob") named the wrong half of the window; the fix is the same either way: fsync-file + rename + dir-fsync makes "Save returned ⇒ content durable" true, closing the window exactly because the row follows the return.
- **Keep file-then-row**: swapping to row-then-file would create the strictly worse direction (row referencing a never-landed file). The orphan-file window (crash between Save and row persist) is a leak, not corruption; no sweep exists for untracked blobs — a possible future retention-sweep extension, NOT part of the adoption train.
- **`Save` signature claim verified**: `internal/blob/store.go:50` — bare `os.WriteFile(abs, content, 0o640)`, 0o640 perms, nanoid names, `ParseRelPath` branding after write.

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-atomic-write` |
| In webphone | adopted v0.6.2 (blob.Save + 2 CLI writers)

## What the library offers

`Write(path, data)` — unique-`.tmp` staging, fsync file, flock+verify, atomic rename, directory fsync; `WriteVerified(path, data, fp)` with `Fingerprint` (xxhash64) TOCTOU check; `WriteIfChanged(path, data) (changed, err)` — skips the write (and the mtime churn) when bytes are identical; `WriteFunc`/`WriteFuncVerified` streaming variants; `ErrConcurrentModification`; platform-split rename (unix/windows).

## Webphone overlap (verified)

| Site | Today | Exposure |
|---|---|---|
| **internal/blob/store.go:50** — `Save` persists attachments/fax PDFs with `os.WriteFile(abs, content, 0o640)` | bare write, file-then-row (verified T01) | `WriteFile` returning ≠ bytes on disk; a power loss **after the DB row commits** can leave a truncated/zero-length blob behind a row that references it — the read path then 404s/500s on user data (inbound faxes: the only copy). fsync+rename makes "Save returned ⇒ durable", closing the window exactly because the row follows the return |
| cmd/webphone/catalog.go:56,64 + contracts.go:36 — generated spec trees/docs | bare `os.WriteFile` | freshness gates diff these files; a torn write produces a confusing diff. `WriteIfChanged` also stops needless mtime churn |
| Sessions, enroll tokens, matrix cursor, snippets… | SQLite (WAL) | engine-owned durability — correctly NOT file-write territory |
| The `CreateTemp` in the health probe | probe-only | not a persistence path |

## Fit analysis

- **`blob.Save`**: adopt `atomicwrite.Write` (content is already fully in memory as `[]byte` — no streaming need). ~2-line change + dependency add (xxhash + flock are its deps; flock is stdlib-adjacent). Impact: closes the only user-data durability hole found in this audit series.
- **CLI spec writers**: adopt `WriteIfChanged` — makes `contracts`/`catalog` runs idempotent on disk.
- Streaming variants (`WriteFunc`) have no consumer today; fingerprints are optional hardening for concurrent writers (webphone has none per path).

## Ruling

**ADOPT for `blob.Save`** (the durability argument stands on its own) and opportunistically for the two CLI writers. Sequencing note: a new direct dependency means the vendorHash re-pin ritual rides the same commit (`go mod tidy` → re-pin → CI-watch), per the dependency-bump doctrine. Alternative if the dependency is unwanted: hand-roll fsync+rename at blob.Save (12 lines) — but then the one home for the pattern is a comment, not a library.

**Refresh trigger:** any new file-persisting feature (exports, backups beyond the NixOS oneshot, log rotation).

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "os.WriteFile" internal/blob/store.go              # expect 0 (the probe rides CreateTemp+Remove since the closeout train; nothing bare remains) (health probe still uses CreateTemp); the Save write is atomicwrite now
rg -n "os.WriteFile" cmd/webphone/catalog.go cmd/webphone/contracts.go   # expect 0 hits (WriteIfChanged adopted)
rg -n "atomicwrite" --type go internal/ cmd/ | wc -l     # expect > 0 (adopted 2026-11-04)
```
