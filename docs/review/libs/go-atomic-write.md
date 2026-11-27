# go-atomic-write — library utilization audit

_Produced by E (subagents 14-17) of the 2026-11-22 full-fleet re-audit's fan-out (runbook step 7); every other line is the audit's own._

**Verdict:** the sharpest NEW finding of this audit series: webphone's one runtime file-write path persisted user data with a bare `os.WriteFile` — no fsync, no atomic rename — while this library exists precisely for that. **Adoption: LANDED 2026-11-04 (v0.6.2) — `blob.Save` rides `WriteWithPerm(…, 0o640)`, both CLI spec writers ride `WriteIfChanged`; vendorHash re-pinned in the same train.**
**Pin:** v0.6.2 · **Latest:** v0.6.2 (local checkout tag) · **Status:** direct dependency
**Audited:** 2026-11-04 · **Re-verified:** 2026-11-22 (currency + `[Unreleased]` read — verdict unchanged). · **Prior reports:** none (first audit).

## Errata — verified write ordering (T01, 2026-11-04)

Step-0 verification of the plan (`docs/planning/2026-11-04_18-41_SUPERB-fleet-libs-closeout.md`) confirmed the finding and sharpened the crash window:

- **Ordering is file-then-row at all five `Save` call sites** (re-counted 2026-11-24 by the citation sweep — the 2026-11-19 retry-owns-its-copy fix added a fifth site the 2026-11-22 refresh missed, and its fourth item was the retry lane mislabelled inbound): `internal/messaging/service.go:141→159` (outbound attachments), `internal/messaging/service.go:210→224` (inbound attachments), `internal/fax/service.go:118→124` (outbound spool), `internal/fax/service.go:219→225` (retry's own blob copy), `internal/fax/service.go:244→255` (inbound spool). The DB row (SQLite, fsynced WAL) lands only after `Save` returns.
- **The precise corruption window is row-committed-before-file-durable**: `os.WriteFile` returning ≠ bytes on stable storage; a power loss after the row commit but before page-cache flush leaves the row referencing a missing/zeroed/truncated file. For inbound faxes the blob is the ONLY copy — this is the user-data-loss class. A crash _between_ file-create and the row does NOT corrupt referenced data (no row exists yet → orphan file at worst).
- **The report's original phrasing** ("a crash between file-create and final bytes corrupts the blob") named the wrong half of the window; the fix is the same either way: fsync-file + rename + dir-fsync makes "Save returned ⇒ content durable" true, closing the window exactly because the row follows the return.
- **Keep file-then-row**: swapping to row-then-file would create the strictly worse direction (row referencing a never-landed file). The orphan-file window (crash between Save and row persist) is a leak, not corruption; no sweep exists for untracked blobs — a possible future retention-sweep extension, NOT part of the adoption train.
- **`Save` signature claim verified**: `internal/blob/store.go:50` — bare `os.WriteFile(abs, content, 0o640)`, 0o640 perms, nanoid names, `ParseRelPath` branding after write.

## Status

|             |                                            |
| ----------- | ------------------------------------------ |
| Module      | `github.com/larsartmann/go-atomic-write`   |
| In webphone | adopted v0.6.2 (blob.Save + 2 CLI writers) |

## What the library offers

`Write(path, data)` — unique-`.tmp` staging, fsync file, flock+verify, atomic rename, directory fsync; `WriteVerified(path, data, fp)` with `Fingerprint` (xxhash64) TOCTOU check; `WriteIfChanged(path, data) (changed, err)` — skips the write (and the mtime churn) when bytes are identical; `WriteFunc`/`WriteFuncVerified` streaming variants; `ErrConcurrentModification`; platform-split rename (unix/windows).

## Webphone overlap (verified)

| Site                                                                                                          | Today                                    | Exposure                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------------------------------------------------------------------- | ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **internal/blob/store.go:50** — `Save` persists attachments/fax PDFs with `os.WriteFile(abs, content, 0o640)` | bare write, file-then-row (verified T01) | `WriteFile` returning ≠ bytes on disk; a power loss **after the DB row commits** can leave a truncated/zero-length blob behind a row that references it — the read path then 404s/500s on user data (inbound faxes: the only copy). fsync+rename makes "Save returned ⇒ durable", closing the window exactly because the row follows the return |
| cmd/webphone/catalog.go:56,64 + contracts.go:36 — generated spec trees/docs                                   | bare `os.WriteFile`                      | freshness gates diff these files; a torn write produces a confusing diff. `WriteIfChanged` also stops needless mtime churn                                                                                                                                                                                                                      |
| Sessions, enroll tokens, matrix cursor, snippets…                                                             | SQLite (WAL)                             | engine-owned durability — correctly NOT file-write territory                                                                                                                                                                                                                                                                                    |
| The `CreateTemp` in the health probe                                                                          | probe-only                               | not a persistence path                                                                                                                                                                                                                                                                                                                          |

## Fit analysis

- **`blob.Save`**: adopt `atomicwrite.Write` (content is already fully in memory as `[]byte` — no streaming need). ~2-line change + dependency add (xxhash + flock are its deps; flock is stdlib-adjacent). Impact: closes the only user-data durability hole found in this audit series.
- **CLI spec writers**: adopt `WriteIfChanged` — makes `contracts`/`catalog` runs idempotent on disk.
- Streaming variants (`WriteFunc`) have no consumer today; fingerprints are optional hardening for concurrent writers (webphone has none per path).

## Ruling

**ADOPT for `blob.Save`** (the durability argument stands on its own) and opportunistically for the two CLI writers. Sequencing note: a new direct dependency means the vendorHash re-pin ritual rides the same commit (`go mod tidy` → re-pin → CI-watch), per the dependency-bump doctrine. Alternative if the dependency is unwanted: hand-roll fsync+rename at blob.Save (12 lines) — but then the one home for the pattern is a comment, not a library.

**Refresh trigger:** any new file-persisting feature (exports, backups beyond the NixOS oneshot, log rotation).

## 2026-11-22 re-verify

Currency: pin v0.6.2 = remote latest (local checkout tag v0.6.2). `[Unreleased]` read in full — two Go-module items, neither touching the durability contract webphone rides (fsync + rename + dir-fsync unchanged in `WriteWithPerm`):

- **`WriteFuncWithPerm(path, fn, perm)`** — the streaming sibling of the `WriteWithPerm` `blob.Save` rides, closing the 0644-default mode gap for streamed writes (previously only byte-slice writes could pick a mode). No webphone consumer today — the fit analysis's "content is already fully in memory" stands — but it becomes the natural pick if the refresh trigger's exports/backups ever stream.
- **Parent-dir `MkdirAll` on every write entrypoint** — belt-and-braces for webphone, which pre-creates directories at the call sites (`internal/blob/store.go:40`, `os.MkdirAll(…, 0o750)`); no action needed, and the library stops depending on caller hygiene.

The rest of `[Unreleased]` is website/CI plumbing (landing redesign, sitemap postbuild gate, lockfile drift guards, a deterministic collision test for the temp-path contract). Verify block replayed green: 0 bare `os.WriteFile` in `internal/blob/store.go`, 0 in `cmd/webphone/{catalog,contracts}.go`, 5 `atomicwrite` hits — `internal/blob/store.go:67` (`WriteWithPerm(…, 0o640)`) plus `WriteIfChanged` at `cmd/webphone/catalog.go:67,75` and `cmd/webphone/contracts.go:40,59`.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "os.WriteFile" internal/blob/store.go              # expect 0 (the probe rides CreateTemp+Remove since the closeout train; nothing bare remains) (health probe still uses CreateTemp); the Save write is atomicwrite now
rg -n "os.WriteFile" cmd/webphone/catalog.go cmd/webphone/contracts.go   # expect 0 hits (WriteIfChanged adopted)
rg -n "atomicwrite" --type go internal/ cmd/ | wc -l     # expect > 0 (adopted 2026-11-04)
```
