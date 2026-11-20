# CQRS / Event-Sourcing / catalog-compliance assessment — session status

> Claims verified as of 2026-11-19 20:23 CET. Session scope: a READ-ONLY
> architecture assessment (three questions: "how Command + Event-Sourced are
> we?", "how CQRS are we?", "how catalog-compliant are we?") delivered in
> chat, plus this report. No repo files were modified before this report.

**Scope line.** Verified here: live catalog exports run twice
(`go run ./cmd/webphone catalog` default tree AND `-plain`, both exit 0,
trees listed and frontmatter inspected); command/decider/event/fold
inventories enumerated by ripgrep over `internal/cqrs` (RegisterCommand,
RegisterDecider, EventTypeName, OnRecord); composition root read
(`internal/app/cqrs.go`); `internal/domain/catalog.go`, `events.go`,
`catalog_test.go`, `cmd/webphone/catalog.go`, `.eventcatalog-hub.json`,
`scripts/check-catalog-contract.py`, `.github/workflows/ci.yml` read in
full; go-cqrs-lite `catalog/v4/README.md` read; the server write/read seam
map produced by a read-only subagent whose key citations (dispatch helpers,
`system.New`, gates list) I spot-verified myself. NOT verified here:
`go test` was never run (the pin tests cited were READ, not executed); the
federation hub's LIVE v2 contract URL was not fetched —
`check-catalog-contract.py` was exercised against a local stand-in schema I
wrote inline (exit 0), which proves the script runs, not hub compliance;
the subagent's full writes/reads tables rest on its reads, spot-checked
only; no benchmarks, no flake check, no CI verdict read for HEAD.

## a) FULLY DONE

1. **Q1 answered: fully command-driven + event-sourced for all product
   state.** 26 registered commands over 11 event-sourced aggregates
   (`system.RegisterDecider` + `system.RegisterCommand` pairs in
   `internal/cqrs`: threads 5, faxes 3, transcripts 3, call notes 3,
   contacts/snippets/dial/enroll/push 2 each, reviews + agent-audit 1
   each); 30 distinct journal event types; ~28 `OnRecord` folds; machine
   tables (`deliveryMoves`/`faxMoves`/`AppendToThread`) own legality;
   dedup CAS (`sys.Dedup`), declarative caps (`WithCapPerGroup`), GDPR
   tombstones + `TrimStream`, projectionhost DLQ all live. ~28-30 HTTP
   write lanes dispatch commands; zero domain writes touch legacy SQL.
   Evidence: greps cited in chat; seam map below.
2. **Q2 answered: textbook single-process CQRS on the ADR-0123 path.**
   `system.New` composition root (`internal/app/cqrs.go:228`; sqlite
   source-of-truth + projections roles, WAL, DLQ at `:203`); read side =
   11 `metaengine.Query` projections behind 11 readmodels gates;
   `readmodels` is gates + views only (purity arch test); read-your-writes
   via `awaitOwnWrite` (`internal/cqrs/visibility.go:127`). Non-CQRS
   islands enumerated, each with a standing ruling: sessions (webphone.db,
   D2), usermgmt.db (library-owned), blob filesystem store (journaled via
   owning commands), remote-only lanes (voicemail/CRM/ASR/textagent).
3. **Q3 answered: catalog compliance strong, one real gap.** Export
   machinery verified live: version stamped on every resource (test-pinned),
   schema-mirror parity pinned, both export trees (composite
   `webphone-1.0.0` refs default / bare IDs `-plain`) render clean,
   bare-filename spec paths carry the live 553-line asyncapi.yaml +
   2364-line openapi.json, CI `catalog-contract` lane validates frontmatter
   against the hub's versioned schema pre-push.
4. **Gap found and evidenced: `ReviewSaved` is the 5th seam event**
   (`internal/domain/events.go:54`, live in the notifier switch at
   `internal/server/notifier.go:77`) **but `BuildCatalog` documents only
   4** — the closed seam set and the catalog disagree; no ruling comment
   justifies the omission. Looks like the 2026-10-28 review train never
   bumped the catalog.
5. **Full server write/read seam map** (subagent, citations spot-checked):
   every write lane → command/journal, direct SQL, blob, or remote; every
   read surface → gate query, PBX API, CDR, or in-memory; the non-CQRS
   islands table (sessions, frozen webphone.db ladder with its live users
   — matrix/signal cursors, retention sweep, metrics counts — blob dir,
   usermgmt.db, system.db dedup CAS).
6. This report filed and the generated status index refreshed.

## b) PARTIALLY DONE

1. **Pin-test verification is read-only.** I read
   `TestBuildCatalog`/`TestCatalogSchemaParity`/`TestCatalogExportFrontmatter`/
   `TestRunCatalogWritesSpecifications` but never RAN even the narrow
   `go test ./internal/domain -run Catalog` — seconds of cost, skipped to
   keep the session read-only. Remaining: one command. Effort S.
   > RESOLVED 2026-11-20 (verification-ring session): the pin suite ran
   > green, and both fresh export trees carry `review.saved` at version
   > 1.1.0 — TA's mechanical half independently verified.
2. **Hub contract check ran against a stand-in schema.** The script exited
   0 against a schema I wrote inline; the hub's live v2 URL was not
   fetched (offline-safe shortcut). Remaining: run with the real contract
   (or confirm the hub-side gate is the authority anyway). Effort S.
   > RESOLVED 2026-11-20: the LIVE oracle ran clean over both fresh
   > exports; the vendored CI copy is sha256-identical to live; the
   > hub's -plain lint consumption confirmed at
   > eventcatalog-hub@84fcada scripts/lint-sources.sh:272. TB's full
   > evidence in the TODO_LIST verification-ring row.
3. **cqrs inventory depth.** Counts come from exhaustive greps; I did not
   individually read all ~25 non-test cqrs files (e.g. per-aggregate
   snapshot presence, fold-by-fold view shapes). The aggregates'
   conclusions rest on the greps + AGENTS.md cross-checks. Effort M to
   deepen; not needed for the assessment verdicts.
   > DEEPENED 2026-11-20 (TF): the snapshot map is now verified from
   > source — threads carries D8's EveryNEvents(200)+ReadPressure(50);
   > the ten other aggregates are per-owner streams via
   > dispatchOwnerCommand; the assessment's "snapshots never assessed"
   > premise was half-stale (see the snapshot verdict row).

## c) NOT STARTED

1. **The ReviewSaved catalog fix** — one `catalog.Event[CatalogReviewSavedSchema]`
   block + mirror + `wantEventIDs` row + parity pair (+ possible
   CatalogVersion bump, see question 1). Offered in chat; awaiting owner
   GO. Priority high, effort S.
   > DONE 2026-11-19 evening (verification session) + 2026-11-20
   > independent verification: the event row, mirror, bump 1.0.0→1.1.0,
   > and the WithVersion stamp on all five events landed; render
   > verified in both trees. TODO_LIST row struck with evidence.
2. **Recurrence guard** `TestCatalogSeamEventSetParity` — reflect over
   `domain.Event` implementers, require each to appear in `BuildCatalog`.
   Would have caught the gap at landing time. Effort S.
   > LANDED 2026-11-19 as the mechanically enforceable half:
   > TestCatalogSchemaParity gained a pairs↔catalog COUNT cross-check
   > (the two hand lists must agree); full implementer enumeration
   > stays impossible via reflection (unexported marker method) — the
   > honest ceiling, recorded in the TODO row.
3. **HARVEST of section (f)** into TODO_LIST.md per docs-health — the
   skill's loop-closing step; deliberately deferred because the owner
   instructed report-then-wait.
   > DONE: seven rows harvested (TODO_LIST § 2026-11-19 catalog/CQRS
   > assessment harvest); ALL SEVEN now struck or ruled with evidence
   > as of 2026-11-20 — the 20:51 plan's TA–TH executed to completion
   > (docs/planning/2026-11-19_20-51_SUPERB-catalog-integrity-verification-ring.md).

## d) TOTALLY FUCKED UP

1. **Both cqrs-inventory subagent dispatches died on provider rate
   limits** (zai "too many requests"), and the SECOND retry repeated the
   same provider/model instead of switching — one wasted dispatch.
   Recovered by doing the inventory directly (grep-first, then targeted
   reads). No repo damage: session was read-only until this report.
2. **One tool call blocked by the `rg --replace` hook** — my `rg -rn`
   habit (flagged as the third strike of that trap on this host).
   Re-ran with correct flags. Process note: `rg -n` is already recursive.
3. **An interim chat message asserted "22 commands"** before the verified
   count of 26 (final report said 26). Self-corrected, but the wrong
   number was spoken once: count before asserting, not after.

## e) WHAT WE SHOULD IMPROVE

1. **Subagent retry policy**: on a provider rate-limit failure, the
   redispatch should switch provider/model, never repeat the same one.
   Cost here: one wasted dispatch + wait latency.
2. **Run the cheap pins after reading them**: after assessing
   catalog pins, running the narrow test subset would have upgraded
   "read the pins" to "pins pass on this tree" for seconds of cost.
3. **Catalog seam-set integrity is hand-maintained** (`wantEventIDs`
   literal); a reflect-driven parity test closes the class permanently —
   this is repo (f) item 3, and the ReviewSaved gap is its proof.
4. **CI catalog lane covers only the default tree**: `ci.yml` exports the
   render tree; the `-plain` governance tree is the hub's job. Cross-repo
   assumption, not verified here — worth one look at the hub's lint config.

## f) Next tasks (what THIS session surfaced — 16, not padded to 50)

| #  | Task                                                                                                        | Impact | Effort | Category      |
| -- | ----------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Land `ReviewSaved` in `BuildCatalog` (+ mirror, `wantEventIDs`, parity pair)                                | High   | S      | Bug (docs)    |
| 2  | Decide the CatalogVersion bump rule for ADDED messages (bump with the fix?)                                 | High   | S      | Documentation |
| 3  | Add `TestCatalogSeamEventSetParity` (domain.Event implementers ⊆ catalog)                                   | High   | S      | Quality       |
| 4  | Re-run `check-catalog-contract.py` against the hub's live v2 URL post-fix                                   | Medium | S      | Quality       |
| 5  | Verify the hub actually lints the `-plain` tree (cross-repo config read)                                    | Medium | S      | Quality       |
| 6  | systemscenario spike for the cqrs Ginkgo suites (notes first) — or write the why-not ruling                 | Medium | M      | Quality       |
| 7  | Confirm `readmodels/faxesgate.go`'s `metaengine.Query` sits right with the gates-and-views-only purity test | Low    | S      | Quality       |
| 8  | Codify the CatalogVersion bump policy in `catalog.go`'s comment block                                       | Low    | S      | Documentation |
| 9  | HARVEST this (f) into TODO_LIST.md (docs-health)                                                            | Medium | S      | Process       |
| 10 | Assess snapshot need for long-lived streams (threads per owner, unbounded versions?)                        | Low    | M      | Quality       |
| 11 | Delete test-only `store.Messages` (messages.go:19-24)                                                       | Low    | S      | Cleanup       |
| 12 | Route metrics counts (`ReadCounts`/`FlagPresent`) off the frozen webphone.db ladder                         | Low    | M      | Cleanup       |
| 13 | Operator note: readmodels.db is an inert orphan on upgraded deployments (README)                            | Low    | S      | Documentation |
| 14 | When landing #1, document ReviewSaved's payload-less nudge semantics in its summary                         | Low    | S      | Documentation |
| 15 | Consider a CI job variant linting the `-plain` export pre-push                                              | Low    | S      | Quality       |
| 16 | DLQ depth visibility: verify an operator surface exists for poisoned projections (metrics/log)              | Medium | S      | Feature       |

## g) Questions (cannot figure these out myself)

1. **Land the ReviewSaved fix now — and does ADDING a fifth event bump
   `CatalogVersion` 1.0.0 → 1.1.0 in the same train?** The comment's rule
   says bump "when a declared message changes shape"; adding a message is a
   catalog-shape change under a strict reading, silent under a loose one.
   Owner intent decides; the fix is S-effort either way.
2. **Is seam-only catalog scope the permanent ruling?** The 30 internal
   journal events (`MessageAppended`, `FaxStatusAdvanced`, …) are
   deliberately uncataloged (catalog = service boundary). Confirm that is
   the standing ruling, so nobody "completes" the catalog with internals
   later.
3. **Keep the hand-rolled `bootDomainSystem` BDD boots, or migrate the
   cqrs suites to the library's `systemscenario`?** The skill's
   recommendation is systemscenario (same DomainConfig the binary uses);
   the current boots work and are real-system. Owner taste; not
   discoverable from the tree.
