## Problem

A fleet deploy bounce of the daemon service turns into a silent push stall:
on 2026-10-27, a `switch-to-configuration test` (SystemNix) terminated
`projects-management-automation.service` at 20:19:41; pushes had stopped at
20:09:15 and only resumed 20:40:36 — webphone's oldest unpushed commit waited
30m37s, follow-on gaps hit 54m and 70m. Nothing in the journal names the
stall, because nothing runs in a dead process and nothing at shutdown says
what was left behind.

Source at `fbd3b1a8`:

- `Stop()` cancels the context and waits — the batch loop's
  `case <-s.ctx.Done(): return` drops whatever sits in the debounce window
  or the queue; there is no final drain:

  ```go
  // pma-daemon/batch.go, batchLoop
  case <-s.ctx.Done():
      return
  ```

- `Start()` boots the watcher and waits for file events; there is no sweep
  of pre-existing unpushed commits, so work stranded by the previous
  termination waits for the NEXT edit in that repo to re-enter the batch
  pipeline.

Per-push logging already exists (`pushProject` logs `pushed changes` /
`push failed`), so the observability gap is not the push itself — it is that
a shutdown with pending work, and a startup with stranded work, are both
indistinguishable from idle.

## Ask

1. On SIGTERM, after the loops stop, log and flush: one line naming the
   pending batches (`terminating with N pending batches: ...`), then a final
   commit-and-push pass for them (bounded, best-effort — a failed flush keeps
   the log line, never blocks exit past the unit's timeout).
2. On startup, one immediate sweep: any repo with commits ahead of its
   remote gets a batch enqueued right away, so a bounce costs seconds, not
   the next edit.
3. One log line at shutdown completion (`stopped; pending=0` or the count),
   so a stall window is greppable from the daemon's own journal instead of
   reconstructed from consumer-side `git ls-remote` forensics.

## Why it matters

Every `switch-to-configuration` on the host predicts a push stall today; the
consumer-side mitigation (webphone's `scripts/push-stall-watch.sh`, a LOCAL
timer) exists only because the daemon's own state is unobservable while its
push leg is dead.

💘 Generated with Crush
