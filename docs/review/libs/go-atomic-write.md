# go-atomic-write — library utilization audit

**Verdict:** the sharpest NEW finding of this audit series: webphone's one runtime file-write path persists user data with a bare `os.WriteFile` — no fsync, no atomic rename — while this library exists precisely for that. **Adoption: n/a today; fit: HIGH (small surface, real durability win).**
**Pin:** none (absent from go.mod) · **Latest:** v0.6.2 (local checkout tag) · **Status:** not a dependency
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-atomic-write` |
| In webphone | absent (0 references) |

## What the library offers

`Write(path, data)` — unique-`.tmp` staging, fsync file, flock+verify, atomic rename, directory fsync; `WriteVerified(path, data, fp)` with `Fingerprint` (xxhash64) TOCTOU check; `WriteIfChanged(path, data) (changed, err)` — skips the write (and the mtime churn) when bytes are identical; `WriteFunc`/`WriteFuncVerified` streaming variants; `ErrConcurrentModification`; platform-split rename (unix/windows).

## Webphone overlap (verified)

| Site | Today | Exposure |
|---|---|---|
| **internal/blob/store.go:50** — `Save` persists attachments/fax PDFs with `os.WriteFile(abs, content, 0o640)` | bare write | power-loss/kill -9 can leave a **truncated or zero-length blob on disk while the DB row already references it** — the read path then 404s/500s on user data. The write order (DB row vs file) does not fence this: a crash between file-create and final bytes corrupts the blob. fsync+rename closes the window |
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

```console
rg -n "os.WriteFile" internal/blob/store.go              # the finding (:50)
rg -n "os.WriteFile" cmd/webphone/catalog.go cmd/webphone/contracts.go
rg -rn "atomicwrite" --type go internal/ | wc -l         # expect 0 until adopted
```
