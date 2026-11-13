> ARCHIVED 2026-11-08 (docs-health drain): gates clean (check-rows --strict + bare-rows green), live facts carried in AGENTS/TODO_LIST — cadence headroom for the SUPERB phase-closeout reports.

# Dedup sweep + tw-freshness repair — 2026-11-06 22:4x (red-main run 42209008464)

> Claims verified as of 22:45 local. CI statements rot fastest; the run
> verdict below is from the local re-run, the push verdict lands in the
> ledger row.

## What ran

1. **Red reproduction at HEAD** (not at the red run's sha): CI run
   42209008464 went red on the 197-file docs-gates daemon push
   (`58996d95`) with two lanes — `tw.css/login.css freshness` and
   `art-dupl check`. The six follow-up daemon commits touched neither
   the CSS artifacts nor `.art-dupl-baseline.json`, so BOTH lanes were
   re-run locally at HEAD before touching anything.
2. **tw-freshness**: `nix develop -c scripts/build-tw-css.sh` →
   rebuild diffed against the committed bytes: a missing `.table-cell`
   utility (both artifacts) and a stale `.wp-live-context` selector
   pair in login.css. Healed by committing the rebuild.
3. **art-dupl**: CI-pinned v0.7.2 at the gate's exact flags →
   **10 new groups** (all 2026-11-06 trains: Reports issue #12,
   History review/empty-state, live-context issue #10, asr/crm seams,
   config CA validator).

## Adjudication (before any baseline write)

- **EXTRACTED (1)**: `reportsStat` — the six hand-rolled `wp-stat`
  card blocks in reports.templ. Old-vs-new render diff: snapshot
  harness rendered `ReportsPanel` before and after, **diff exit 0 —
  byte-identical**. Pinned by `reports_view_test.go` (card bytes
  plain+alert, six-card grid, alert-only-on-missed).
- **ACCEPTED (9)**: rows added to the standing table (templ local
  binds; two-arm text conditionals + count>0 notices; per-seam
  `NewClient`/validator prologues; one-home call-site pairs; mount
  divs; guard-clause returns).

## Baseline

Re-recorded with the pinned binary + gate flags: **98 → 105**.
Set-diff verified: **+9 = exactly the adjudicated hashes**, −2 drops
(a messages.templ pair and a 15-file server-test group) re-keyed by
the week's line drift — the phantom-identity class, not blessings.
Post-write `check` = **0 new**.

## Tail work

- Registry: 5 ruling rows + sweep-log line (docs/dedup-registry.md).
- CHANGELOG `[Unreleased]` line for the `reportsStat` one-home.
- ci-flake-ledger row for run 42209008464 (filled after push+green).
