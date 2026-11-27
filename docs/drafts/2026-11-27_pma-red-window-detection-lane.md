# pma detection-lane ask — red-window persistence alarm

**Provenance:** drafted by an agent session (pareto plan T05/m05.4,
2026-11-27) for owner review; facts read from the webphone tree at the
time of writing (`.github/workflows/watch.yml`, `scripts/main-red-watch.sh`,
`docs/planning/2026-11-24_01-24_SUPERB-post-deploy-pareto-replan.md`
S22-1409-4 row).

## The record

webphone main sat RED in four windows the 15-minute `watch` observer
lane failed to surface as tracker issues:

1. 2026-11-22 ~17:12Z — ~1.5 h unfiled (S22-1409-4, the third miss).
2. 2026-11-22 17:12Z → 2026-11-24 01:04Z — the longest: ~32 h red with
   no tracker issue, because the watcher's `ci-verdict.sh` answered
   NO-RUN (the watch lane's own runs had pushed the CI run off page one
   of the `actions/runs?head_sha=` listing — since fixed by pagination
   in the script, but the CLASS is the point).

Four misses across two months is not one bug: each repair closed a
symptom. The structural premise — ONE 15-minute lane, whose every
failure mode (listing pagination, workflow-name misreads, search-index
lag, the lane's own runs crowding the listing) silently disables the
only detector — is what keeps producing missed windows.

## The ask

pma's detection lane gains a SECOND, deliberately slower red-window
alarm that does not share the 15-minute lane's code path:

- Cadence ~hourly.
- Read the CI workflow's LATEST CONCLUDED run on main (workflow-name
  verified, no name-less green).
- When the latest conclusion is `failure` AND no open tracker issue
  names it, file one.
- When a red window has persisted (latest still `failure`) for more
  than the bar (4 h suggested), file/comment regardless of dedup state
  — a long red window is itself the signal, even if some issue
  somewhere mentions the run.

The two lanes fail differently (different reader code, different
cadence, different dedup), so a silent failure in one no longer zeroes
detection. The cost is one API listing per hour.
