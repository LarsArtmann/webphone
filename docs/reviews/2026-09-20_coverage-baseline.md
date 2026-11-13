# Coverage Baseline — 2026-09-20

> **LIVING BASELINE — append a new dated section per re-run; never overwrite an
> older one.** The 2026-09-20 table below is the original snapshot; the newest
> section is the current reference. Latest re-run: **2026-11-06** (§ 2026-11-06
> re-run, HEAD `54c8bb17`).

- **Command:** `nix develop -c go test -count=1 -cover ./...`
- **Why persisted:** the 2026-09-20 audit session computed numbers that
  lived only in chat. This file is the reference point the next coverage
  train diffs against. **Re-run at every release** (release runbook step 3
  extension) and after any new package.

## Per-package statement coverage

| Package               | Coverage | Statement count | Note                                                                               |
| --------------------- | -------- | --------------- | ---------------------------------------------------------------------------------- |
| `internal/vcard`      | 92.6%    | small           | Best in repo; table-driven parser tests                                            |
| `internal/server`     | 87.6%    | large           | Contract tests + behavior suites; views rendering is transitive                    |
| `internal/pbx`        | 87.1%    | small           | Live-probe + httptest client                                                       |
| `internal/config`     | 83.0%    | small           | Env/file precedence                                                                |
| `internal/gateway`    | 76.5%    | small           | Webhook error branches are the gap (plan T17)                                      |
| `internal/session`    | 67.0%    | small           | Was 63.1%; SQLite store tests raised it (T12)                                      |
| `internal/store`      | 59.4%    | medium          | Owner-scoping behavior suite missing (plan T15)                                    |
| `internal/domain`     | 54.1%    | small           | Parser edge table missing (plan T18)                                               |
| `internal/views`      | 1.3%     | large (templ)   | Direct coverage only; rendered via server tests (plan T26 decision)                |
| `internal/blob`       | 0.0%     | small           | No direct tests; exercised via server attachment paths                             |
| `internal/fax`        | 0.0%     | small           | No direct tests; exercised via server fax tests                                    |
| `internal/messaging`  | 0.0%     | small           | No direct tests; exercised via server messages tests                               |
| `internal/web/assets` | 0.0%     | n/a (embed)     | JS behavior lives in `internal/web/assets/island-tests/` (node:test, not go cover) |
| `cmd/webphone`        | 0.0%     | small           | Wiring only; smoke suite boots the real binary                                     |
| `internal/arch`       | —        | no statements   | Pure test package (import-graph gate)                                              |

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

| Date       | Package | Before → After | Cause                          |
| ---------- | ------- | -------------- | ------------------------------ |
| 2026-09-20 | session | 63.1% → 67.0%  | SQLite store suite (plan T12)  |
| 2026-09-20 | store   | 59.4% → 75.3%  | Owner-scoping suite (plan T15) |
| 2026-09-20 | gateway | 76.5% → 77.6%  | Webhook error branches (T17)   |
| 2026-09-20 | domain  | 54.1% → 72.1%  | Parser edge table (T18)        |

## 2026-11-06 re-run — HEAD `54c8bb17` (docs-health v8 sweep)

Same command, same host. Island JS separately:
`nix run nixpkgs#nodejs -- --test --test-force-exit
'internal/web/assets/island-tests/*.test.mjs'` → **289 specs, 289 pass,
0 fail** (was 22 specs on 2026-09-20).

| Package                 | 2026-11-06 | 2026-09-20 snapshot / delta | Note                                                                   |
| ----------------------- | ---------- | --------------------------- | ---------------------------------------------------------------------- |
| `internal/sipcred`      | 100.0%     | new package                 | Derived-credential primitive (2026-10-12 brokering train)              |
| `internal/admin`        | 95.5%      | new package                 | Read-only cross-line surface (2026-11-04)                              |
| `internal/chatloop`     | 94.7%      | new package                 | Matrix/signal receiver loops (2026-10-23)                              |
| `internal/reports`      | 94.4%      | new package                 | Pure CDR projection (2026-11-06, issue #12)                            |
| `internal/paperless`    | 92.7%      | new package                 | Inbound-fax archiver seam                                              |
| `internal/vcard`        | 92.4%      | 92.6%                       | Unchanged, still the best parser suite                                 |
| `internal/session`      | 88.5%      | 67.0% (→ 67.0%)             | The SQLite store + sliding-renewal suites                              |
| `internal/pbx`          | 85.7%      | 87.1%                       | Live-probe + httptest client                                           |
| `internal/sso`          | 85.2%      | new package                 | OIDC lane                                                              |
| `internal/config`       | 85.0%      | 83.0%                       | Env/file precedence + the both-or-neither validation families          |
| `internal/domain`       | 81.9%      | 54.1% (→ 72.1%)             | Parser edge table + the network/address axis                           |
| `internal/server`       | 81.9%      | 87.6%                       | Denominator grew with the admin/reports/live-context handlers          |
| `internal/textagent`    | 82.1%      | new package                 | Agent seam                                                             |
| `internal/asr`          | 82.4%      | new package                 | ASR client                                                             |
| `internal/authmap`      | 80.9%      | new package                 | Identity mapping                                                       |
| `internal/crm`          | 80.8%      | new package                 | CRM resolver + vocabulary bridge                                       |
| `internal/gateway`      | 80.0%      | 76.5% (→ 77.6%)             | Webhook error branches + honest part headers                           |
| `internal/retention`    | 80.0%      | new package                 | The 2026-11-05 hygiene train's sweeper pins                            |
| `internal/web/assets`   | 80.0%      | 0.0%                        | Go-side embed helpers; JS behavior stays in island-tests               |
| `internal/push`         | 76.4%      | new package                 | ntfy + FCM transports                                                  |
| `internal/matrix`       | 75.4%      | new package                 | Chat-network receiver                                                  |
| `internal/blob`         | 73.7%      | 0.0%                        | Direct tests since the 2026-11-04 atomic-write train                   |
| `internal/readmodels`   | 73.2%      | new package                 | Engine store, caps, admin read seams                                   |
| `internal/fax`          | 64.2%      | 0.0%                        | Direct lifecycle tests beside the server-level fax suite               |
| `internal/store`        | 64.8%      | 59.4% (→ 75.3%)             | **Below the 2026-09-20 post-suite figure** — see the reading below     |
| `internal/signal`       | 68.8%      | new package                 | Signal socket reader                                                   |
| `internal/userauth`     | 58.2%      | new package                 | Identity layer; ceremony paths run through the server suites           |
| `internal/app`          | 56.5%      | new package                 | Composition root; wiring asserted adapter-side                         |
| `internal/server/hooks` | 53.8%      | new package                 | `/hooks/*` idempotency + ring lane                                     |
| `cmd/webphone`          | 42.2%      | 0.0%                        | Wiring + boot-report; the smoke suite boots the real binary            |
| `internal/web/views`    | 42.4%      | 1.3%                        | Helpers carry tests; render paths stay transitive (T26 decision holds) |
| `internal/messaging`    | 37.9%      | 0.0%                        | Sender registry + address routing; receive paths exercised via server  |
| `internal/server/api`   | 7.1%       | new package                 | Thin JSON writers; the DTO contract tests live one level up            |
| `internal/arch`         | —          | —                           | Pure test package (import-graph gate)                                  |

**Honest reading (2026-11-06):**

- Two numbers moved the WRONG way and both are explained by denominator
  growth, not by deleted tests: `internal/server` (new admin/reports/
  live-context handlers) and `internal/store`. The store figure is the one
  worth a train: the 2026-09-20 owner-scoping suite still pins the original
  surface, but the store has since grown the admin read seams
  (`ListAllThreads`/`GetThreadAny`/`AttachmentByIDAny`), the thread-flag
  columns and migration v5 — each with a narrower direct suite than the
  surface it adds.
- The 20 packages that are new since 2026-09-20 all arrive WITH tests (the
  lowest are `server/api` 7.1% and `messaging` 37.9%, both deliberately thin
  writers/registries whose behavior is asserted one level up).
- Island JS is where the real growth happened: 22 → 289 specs, with the
  `node:test` harness now gating `nix flake check` (`island-js`).
- Percentages still measure STATEMENTS, not error paths. The 2026-09-20
  reading stands: high % ≠ high trust, and `server/api` at 7.1% is fine
  because its bytes are contract-pinned.

## 2026-11-07 delta — `internal/store` 64.8% → 82.4 % (P1-1 store test train)

One new file, `internal/store/legacy_import_test.go`, closed the biggest
untested block in the package: the transitional reads of the RETIRED SQL tables
(`contacts`, `fax_jobs`, `snippets`, `call_transcripts`) plus the `app_flags`
helpers. These are not screen paths — they run once at boot on an upgraded
deployment and then stamp "imported", so a regression imports nothing and
leaves the flag behind. Before this train four readers and four translators sat
at 0.0%.

| Function                                | before | after                 |
| --------------------------------------- | ------ | --------------------- |
| `ReadLegacyContacts` / `ToContact`      | 0.0%   | 100.0%                |
| `ReadLegacyFaxes` / `ToFax`             | 0.0%   | 100% / 81.2%          |
| `ReadLegacySnippets` / `ToSnippet`      | 0.0%   | 100.0%                |
| `ReadLegacyTranscripts` / `ToSegment`   | 0.0%   | 100.0%                |
| `FlagPresent` / `FlagValue` / `SetFlag` | 0.0%   | 85.7% / 85.7% / 66.7% |
| **package total**                       | 64.8%  | **82.4%**             |

The admin read seams the 2026-11-06 reading flagged were already pinned by
`store/admin_reads_test.go` (both directions: the `*Any` seam crosses lines, its
owner-scoped sibling still refuses); this train only measured them
(`ListAllThreads` 76.9%, `GetThreadAny` 85.7%, `AttachmentByIDAny` 85.7% — the
remainder is the `errorfamily.WrapInfrastructuref` branch) and mutation-proved
one assertion: replacing `ToContact`'s owner parse with a constant makes
`TestToContactSkipsCorruptRowsAndDemotesTheID` fail, then passes again on
restore.

## Decision: views coverage stays transitive (plan T26)

`internal/views` renders templ components; DIRECT render tests would
duplicate what every server test already exercises through the real
router (page shells, partials, error panels, CSP/DOM contracts all
asserted over HTTP responses). Decision (2026-09-20): accept transitive
coverage; revisit only if a view grows real logic (loops with
computation, conditionals that encode business rules) — and then
extract that logic into helpers WITH tests, not render snapshots.
