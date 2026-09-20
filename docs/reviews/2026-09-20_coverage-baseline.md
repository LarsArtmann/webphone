# Coverage Baseline — 2026-09-20

- **Command:** `nix develop -c go test -count=1 -cover ./...`
- **Why persisted:** the 2026-09-20 audit session computed numbers that
  lived only in chat. This file is the reference point the next coverage
  train diffs against. **Re-run at every release** (release runbook step 3
  extension) and after any new package.

## Per-package statement coverage

| Package          | Coverage | Statement count | Note                                                    |
| ---------------- | -------- | --------------- | ------------------------------------------------------- |
| `internal/vcard` | 92.6%    | small           | Best in repo; table-driven parser tests                 |
| `internal/server`| 87.6%    | large           | Contract tests + behavior suites; views rendering is transitive |
| `internal/pbx`   | 87.1%    | small           | Live-probe + httptest client                            |
| `internal/config`| 83.0%    | small           | Env/file precedence                                     |
| `internal/gateway`| 76.5%   | small           | Webhook error branches are the gap (plan T17)           |
| `internal/session`| 67.0%   | small           | Was 63.1%; SQLite store tests raised it (T12)           |
| `internal/store` | 59.4%    | medium          | Owner-scoping behavior suite missing (plan T15)         |
| `internal/domain`| 54.1%    | small           | Parser edge table missing (plan T18)                    |
| `internal/views` | 1.3%     | large (templ)   | Direct coverage only; rendered via server tests (plan T26 decision) |
| `internal/blob`  | 0.0%     | small           | No direct tests; exercised via server attachment paths  |
| `internal/fax`   | 0.0%     | small           | No direct tests; exercised via server fax tests         |
| `internal/messaging` | 0.0% | small           | No direct tests; exercised via server messages tests    |
| `internal/web/assets` | 0.0% | n/a (embed)     | JS behavior lives in `internal/web/assets/island-tests/` (node:test, not go cover) |
| `cmd/webphone`   | 0.0%     | small           | Wiring only; smoke suite boots the real binary          |
| `internal/arch`  | —        | no statements   | Pure test package (import-graph gate)                   |

## Island JS (outside go cover)

`nix run nixpkgs#nodejs -- --test --test-force-exit
'internal/web/assets/island-tests/*.test.mjs'` — 22 specs green
2026-09-20 (ui 7, shell 6, session 6, i18n parity 1, i18n misc 2).

## Honest reading

- The numbers above measure STATEMENTS, not error-path coverage; the
  audit (status report 2026-09-20 17:06) judged error paths, not
  percentages. High % ≠ high trust: `views` at 1.3% direct is fine
  (rendered through every server test); `store` at 59.4% is the real
  gap because owner-scoping regressions there leak data.
- Delta rule: a coverage train records its before/after HERE, one table
  row per touched package, dated.

## Deltas

| Date       | Package  | Before → After | Cause                          |
| ---------- | -------- | -------------- | ------------------------------ |
| 2026-09-20 | session  | 63.1% → 67.0%  | SQLite store suite (plan T12)  |
| 2026-09-20 | store    | 59.4% → 75.3%  | Owner-scoping suite (plan T15) |
| 2026-09-20 | gateway  | 76.5% → 77.6%  | Webhook error branches (T17)   |
| 2026-09-20 | domain   | 54.1% → 72.1%  | Parser edge table (T18)        |

## Decision: views coverage stays transitive (plan T26)

`internal/views` renders templ components; DIRECT render tests would
duplicate what every server test already exercises through the real
router (page shells, partials, error panels, CSP/DOM contracts all
asserted over HTTP responses). Decision (2026-09-20): accept transitive
coverage; revisit only if a view grows real logic (loops with
computation, conditionals that encode business rules) — and then
extract that logic into helpers WITH tests, not render snapshots.
