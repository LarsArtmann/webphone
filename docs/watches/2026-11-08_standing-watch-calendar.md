# Standing watches — the ONE dated calendar (refreshed 2026-11-13)

**Status:** CURRENT (every date re-derived this session) · **Supersedes:** the
prose watch list inside `TODO_LIST.md` § Standing watches (that section now
points here; this file is the ONE home so a date can never fork between a
report and a TODO row).

Rule of the house: a watch is either (a) CALENDAR — it fires on a date, or (b)
TRIGGER — it fires on an event, and the event is named precisely enough that a
fresh session can evaluate it. Every row carries: next due, the trigger, the
evidence home, and what to DO when it fires. Named triggers always beat their
calendar.

| Watch                                                                                      | Cadence         | Next due       | Fires earlier when (trigger)                                                                                                                                                                       | Evidence home                                                                                                           | On fire                                                                                                               |
| ------------------------------------------------------------------------------------------ | --------------- | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| erraudit tier-2 re-measure (`check --max --no-suppress` + `nolint-audit`)                  | monthly         | **2026-12-08** | any error-path refactor lands                                                                                                                                                                      | AGENTS § erraudit bar; last FULL measure 2026-11-08 (0 violations); decay pass 2026-11-13: 69/69 directives needed, 0 stale | fix or reasoned-nolint the findings; re-run `nolint-audit` after any error-path refactor (the 41-of-109 decay lesson) |
| Quarterly buildflow re-check (tiers 1+2, dedup baseline, `.buildflow-baseline.json` audit) | quarterly       | **2026-12-20** | a red findings gate, or the baseline-staleness warn naming matches for 3 runs                                                                                                                      | AGENTS § BuildFlow health warning; the tier bar (`--enforce-go-error-family` AND `--enforce-coded-errors` = 0); branching-flow baseline re-recorded 2026-11-13 from this checkout (220 findings, cache-disabled) |
| Fleet tag sweep (`scripts/fleet-tags.sh`)                                                   | trigger-only    | —              | a dependency bump train, a new library joining the fleet, or a release cut in any fleet repo                                                                                                       | AGENTS § Commands (fleet-tags.sh one-liner); the 2026-11-13 catalog tag-shape incident is the pinned false-clean class  | run the sweep before re-pinning vendorHash; a wrong-shape tag must FLAG, not read as tagged | re-measure tiers; regenerate the baseline ONLY from this checkout, CI-pinned detector                                 |
| Sessions credentials-at-rest re-review                                                     | periodic        | **2026-12-30** | ANY session-store schema change (re-opens the same day)                                                                                                                                            | `docs/reviews/2026-09-30_sessions-credentials-at-rest-review.md`                                                        | re-run the at-rest audit against the new schema before merging it                                                     |
| Modularize re-trigger re-review                                                            | quarterly       | **~2027-01**   | a second Go consumer imports `internal/domain` or `internal/gateway` (THE named reopen trigger; fires earlier than any calendar)                                                                   | `docs/modularization/2026-10-17_PROPOSAL.html` §09 (detection commands)                                                 | run the detection commands; only a consumer sighting re-opens the split                                               |
| Dedup ritual sweep (art-dupl `-t 3 --type-aware` + baseline/suppressed-set audit)          | quarterly       | **2027-01-11** | a new one-home helper lands unpinned                                                                                                                                                               | `docs/dedup-registry.md` sweep log; baseline pinned to CI detector v0.7.2                                               | triage per registry rules; append a sweep-log line                                                                    |
| sip.js 0.22 watch                                                                          | trigger-only    | —              | Chromium WebRTC breakage · a sip.js security advisory · a needed capability                                                                                                                        | AGENTS § island bullets (0.21.2 pin; JsSIP 3.13.8 named fallback); npm re-check 2026-09-23 (latest still 0.21.2)        | `./update.sh` + the full island gates + stack E2E                                                                     |
| templ-components upstream watch                                                            | trigger-only    | —              | an adopted component's class set changes (next tw.css regen ONLY then)                                                                                                                             | AGENTS § templ-components bullet (tw.css @source pair)                                                                  | rebuild via `scripts/build-tw-css.sh`, diff committed bytes                                                           |
| oxlint globals                                                                             | trigger-only    | —              | a new browser global appears in island/shell JS                                                                                                                                                    | `internal/web/assets` oxlint config `globals` block                                                                     | add the global to the config, fail-closed stays                                                                       |
| E2E wall-time budget                                                                       | trigger-only    | —              | TWO consecutive over-budget runs (budget 445 s; greens 195/184 s)                                                                                                                                  | AGENTS § stack browser E2E; release-runbook                                                                             | budget stays until two consecutive breaches, then re-derive                                                           |
| go-health M18 park watch                                                                   | trigger-only    | —              | go-health grows startup-hook evaluation, or cqrs-htmx grows a readiness hook                                                                                                                       | TODO § standing watches (wiring reverted `f7028b4`; verified v0.3.0 AND v0.4.0)                                         | re-adjudicate the hook wiring only on those upstream shapes                                                           |
| branded-id Valuer/Scanner                                                                  | trigger-only    | —              | next storage-format touch (schema migration touching branded columns)                                                                                                                              | AGENTS § strong-id ruling                                                                                               | adopt at the store seam in the same train                                                                             |
| go-cqrs-lite upstream filings #71/#72                                                      | trigger-only    | —              | the library ships either (TypedReader.Scan truncation signal / keep-newest-N cap primitive)                                                                                                        | TODO § standing watches (filed 2026-10-15)                                                                              | re-visit webphone's converging-loop trims + enroll sweep as the named trigger                                         |
| QMD `get`/`multi_get` render bug                                                           | trigger-only    | —              | next Crush upgrade ships                                                                                                                                                                           | blocked upstream on crush `#3846` (qmd 2.8.3 spec-compliant)                                                            | re-test `get`; workaround stays `query` + disk reads                                                                  |
| push-stall watchdog                                                                        | standing script | continuous     | silent daemon push lag > 60 min = BROKEN (30-min WARN line pending the D7 ruling)                                                                                                                  | `scripts/push-stall-watch.sh`; AGENTS § manual-push bar                                                                 | invoke the bar; state the anchor                                                                                      |
| Invalidations/doctrine watches NOT carried here                                            | —               | —              | the Fleet-libs reopen-trigger registry, the DECLINED/NO-GO lists, and every ruling with a written reopen trigger live in `TODO_LIST.md` § Fleet-libs reopen-trigger registry (T25) — the ONE table | TODO_LIST § Fleet-libs reopen-trigger registry                                                                          | sweep THAT table, never re-read 24 reports                                                                            |

## Where the next dates come from

- erraudit: monthly from the last measure (2026-11-08 → 2026-12-08). The
  2026-11-05 due date lapsed unmeasured for three days — this refresh ran the
  measure same-session so the calendar restarts from a real number, not a
  promise.
- Quarterly set (buildflow 12-20, modularize ~2027-01, dedup 2027-01-11):
  anchored to their prior entries, unchanged.
- Credentials 12-30: anchored to the 2026-09-30 baseline review + one quarter.

## What the 2026-11-13 refresh changed

1. erraudit row annotated with the train-specific decay pass (69/69, 0 stale) —
   the monthly tier re-measure stays due 2026-12-08 (a decay pass is not the
   quarterly re-measure; the house rule held).
2. branching-flow baseline row annotated: re-recorded 2026-11-13 (the
   moved-checkout protocol exercised ahead of the quarterly due date).
3. NEW trigger-only row: the fleet tag sweep (the catalog v4.15.0 incident —
   a release tagged in a shape Go cannot resolve read as "tagged" until the
   bump failed — is the class this watch exists for).
4. No dated row moved: every calendar date was re-derived and none had fired.

## What the 2026-11-08 refresh changed

1. erraudit re-MEASURED (0 violations) instead of re-dated — the watch was
   overdue and a calendar that moves a lapsed date without running the check
   would be the vacuous-PASS class again.
2. Every trigger-only row gained an explicit "fires earlier when" clause with
   the exact upstream shape to watch for.
3. The reopen-trigger registry is referenced, not duplicated — one home per
   fact.
