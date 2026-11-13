# File-size split train closeout — #248 claimed, cadence CI red fixed

> Claims verified as of 2026-11-13 10:34 CET. Every number below was
> re-read with a fresh command this session: `wc -l` on both split
> files (300 + 88), `buildflow -s file-size-check` exit 0 with an
> empty findings array, `scripts/ci-verdict.sh` (RED on `7f179792`,
> the cadence gate), `docs-gates` full battery PASS after the archive,
> and the `-race -count=3` probe (`ok`, exit 0).

## What this session did

Yesterday's split train closed with two deliberate parks and one open
CI question; this session closed all three.

1. **#248 claimed and landed** — the matrix cursor-reset train had
   closed (2026-11-12 18:57) without splitting its file, so the
   fix-forward hold expired. Read the whole file first, then carved the
   `/sync` wire surface (response shapes + the `Client.Sync`
   long-poll) to `internal/matrix/sync.go`, leaving the loop (Cursor,
   Receiver, Run, self-heal arms, ingest family, the DM-map folders)
   in `receiver.go`. Verified **byte-pure motion**: the 76 moved lines
   diff clean against `HEAD:receiver.go`, and the import delta is
   exactly the three removals (`net/url`, `strconv`, `jsontext`).
   `receiver.go` 382 → 300, `sync.go` 88. Matrix + arch suites green;
   the scoped gate reads **exit 0, findings: []**. #248 closed with
   the evidence comment.
2. **The cadence CI red found and fixed** — `scripts/ci-verdict.sh` on
   the pushed tip read RED (the obligation paid off): the docs-gates
   cadence gate crossed 66 unarchived reports vs threshold 65, exactly
   the §f-16 warning from yesterday's report. None of the 66 reports
   passes both archive gates clean, so the lightest candidate
   (the 2026-11-11 15:08 storage-ladder report, 5 open rows) was
   adjudicated row by row — the CI-verdict row re-verified GREEN via
   `ci-verdict.sh` on `7aa38c83`, four residue rows routed to their
   TODO_LIST / standing-doctrine homes (annotate `r` records) — then
   archived with the banner. Cadence back to 65 (ARCHIVE-HEALTHY);
   the full `docs-gates.sh` battery reads PASS on all eight gates.
3. **The §f harvest landed** — the still-open rows from yesterday's
   22-50 report now live in TODO_LIST (raceprobe decision, baseline
   refresh, conditional ledger row, AGENTS one-liner), and the struck
   "raceprobe NOT REPRODUCIBLE" row carries its inline correction:
   #249's repetition evidence supersedes the single-run probe verdict.
4. **The `-race` probe ran** — `-race -count=3` over the messaging
   package: clean, no race-detector signature. One green run neither
   proves nor falsifies the fold-window claim (the issue's own table
   shows single runs usually pass); the reproducible signal stays
   `-count>1` / full-suite parallelism. Result posted on #249.
5. **The stray `go.mod` delta adjudicated** — the working tree carried
   an un-authored `go 1.27.3` → `go 1.27` downgrade. Root cause found
   in yesterday's pasted buildflow log: the `go-version-auto-configure`
   tool wants a `major.minor`-only directive and attempted the
   downgrade; `go mod tidy` vetoed it ("downgrade is not
   dependency-floor-safe" — cqrs-htmx/loginpage require 1.27.3). The
   delta self-healed under tidy; nothing committed.
6. **Closeout docs** — the CHANGELOG bullet extended to six files;
   the receiver draft marked historical.

## Honest note (the near-miss)

The first write of `sync.go` corrupted one wire tag: `roomSync`'s
inner `Events` field was typed as `json:"ephemeral"` instead of the
correct `json:"events"` — a silent receipt-ingestion break, exactly
what "pure code motion" forbids. The multiedit deletion failed on the
same mistyped byte in its old_string, which forced a re-read, and the
byte-pure diff against HEAD then proved the shipped bytes. Lesson:
when moving code verbatim, extract the bytes from the source (or diff
the moved block against `git show`), never re-type them.

## Not done here (owners/triggers)

- The **full buildflow exit-0 re-certification** — the morning full run
  (before the cadence fix) read exit 69 with the cadence gate as the
  ONLY error finding; every other step was green. Re-run lands after
  this report pushes.
- The **T24 release-tails drain** still owns the remaining 65-report
  corpus (the oldest, 2026-10-25 03-54, carries 13 untouched rows);
  this session archived one report to green CI, not the drain.
- **§g3 owner questions** (raceprobe semantics, §D10 posture, the
  filing review boxes) stay with the owner — the harvest rows and
  #249 carry them.

## f) Top next tasks (this session's delta)

| # | Task                                                                                                                                                                | Impact | Effort | Category      |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1 | Re-run the full buildflow after this push — the exit-0 certification of the whole file-size split train (the morning run's only error finding was the cadence gate) | High   | S      | Quality       |
| 2 | T24 release-tails drain: archive-gate the remaining 65-report corpus oldest-first (the 2026-10-25 03-54 report carries 13 untouched rows)                           | Medium | M      | Documentation |
| 3 | Raceprobe fix decision (#249): await-poll vs fold-window semantics — the owner question parked in the file-size split train residue harvest row                     | High   | S–M    | Quality       |
| 4 | Branching-flow baseline refresh at the next governance sweep: re-home COMPOSITION_mixin, prune the two stale session-carve DUPLICATE_TYPE entries                   | Medium | S      | Quality       |
| 5 | AGENTS.md one-liner at the next legitimate touch: the error-registry regen fires on MOVE, not only on NEW codes                                                     | Low    | S      | Documentation |
