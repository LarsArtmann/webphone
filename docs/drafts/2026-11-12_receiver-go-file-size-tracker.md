> [!NOTE]
> This filing was drafted by GLM-5.3-flash via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

`buildflow -s file-size-check` exits 69 since the cursor-reset self-heal
train: `internal/matrix/receiver.go` grew past the 350-line cap.

```
$ buildflow -s file-size-check --format finding
internal/matrix/receiver.go: 381 lines (max 350)
```

The growth landed in `d15dcc89` (2026-11-12 15:59, "refactor(matrix): the
cursor-reset self-heal arms share one helper", receiver.go +56). The file
has sat at 381 lines since; a re-scan at 18:43 still reports the finding.

For context: the same scan flagged five other oversize files
(`auth/session_api.go`, `asr/asr.go`, `cqrs/transcripts_test.go`,
`server/pages.go`, `views/crm_test.go`) — all five were split along their
natural seams the same day and the scoped gate now reds on this file
alone.

The split is deliberately NOT landed from the reporting session: the file
belongs to a concurrent session's active train, and the fix-forward
protocol says patch their files never. The natural seam I saw while
reading it — the cursor-reset arms vs the ingest loop — is recorded here
so the owning train (or a later session) can move fast.

💘 Generated with Crush
