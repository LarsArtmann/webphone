> [!NOTE]
> This filing was drafted by GLM-5.3-flash via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

`TestConcurrentFirstSightYieldsOneThread` (`internal/messaging/raceprobe_test.go`)
fails intermittently under repetition and under full-suite parallelism:

```
raceprobe_test.go:113: after re-delivery: 1 threads, unread 3 — want 1 thread, both messages counted
```

Reproduction, observed 2026-11-12 (all runs `env -u GOEXPERIMENT go test`):

| probe | verdict |
| --- | --- |
| `-count=8 -run TestConcurrentFirstSightYieldsOneThread` at `b1b01b87` (pre-session HEAD, detached worktree) | FAIL |
| `-count=5`, same test at then-HEAD | FAIL (≥1 of 5) |
| single run, same test | ok (multiple) |
| package alone, `-count=1`, 3× | ok 3/3 |
| full `go test ./...` | FAIL observed twice, ok once |

The test is pre-existing flaky, not a regression: it reproduces at a
commit that predates every working-tree change of the reporting session.

## Why

The final assertion (`raceprobe_test.go:112`) reads `svc.Threads` and
demands `Unread == 2` immediately after the re-delivery `Receive`
returns — no projection-await poll. The two first-sight racers plus the
re-delivery interleave inside the projection fold window; when the
loser's append folds before the `cqrs.thread_identity` rejection
settles, the read observes three folded messages on the one thread. The
transcript tests solved the same shape with an explicit poll helper
(`awaitTranscriptCount`, `internal/cqrs/transcripts_test.go:41`); this
test has no equivalent, so its verdict depends on fold timing.

Not yet observed red in CI (the go-tests workflow runs without
`-count`), but any `-count>1`, `-race`, or full-buildflow test step can
red on it spuriously.

💘 Generated with Crush
