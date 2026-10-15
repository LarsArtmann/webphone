# Metaengine train — T07 swap + main-red repair + G1 GREEN (2026-10-15 17:10)

Session resumed the SUPERB metaengine dogfooding train (plan: `docs/planning/2026-10-14_23-05_SUPERB-metaengine-dogfooding.md`) from the 01:39 handoff. Mid-session, main was RED (two CI failures); it closed GREEN with T07 complete.

## Fully done (verified green)

1. **T07 contacts swap** — `Deps.Contacts` is `*readmodels.Contacts`; `*store.Contacts` provider deleted; `store/contacts.go` removed; `ErrListFull` homed into readmodels (sentinel + family registration); `/api/contacts`, fuzz, contract, owner-scoping, session-behavior suites rewired via `newTestReadmodels` helper; cap test ported with the delete-frees-slot leg; store family pins moved to `internal/readmodels/family_test.go`.
2. **Data continuity (plan gap closed)** — migration v4 adds `app_flags`; boot-time one-time import re-saves legacy SQL contacts into the engine (idempotent upsert-keeps-id; flag stamped only after full success; broken import logs + never blocks boot; crash retries next boot).
3. **Metrics re-source** — `webphone_contacts_total` via the engine's native COUNT (`AggregateReader` asserted at Open, same doctrine as DedupStore); `store.ReadCounts` dropped the dead-table subquery (which would have 500'd the whole metrics surface).
4. **Main-red repair** — the daemon's go.mod sweep (cqrs-htmx v4.14.1, go-health v0.7, go-paperless v0.5.0) broke `internal/paperless`: adopted the v0.5.0 branded-id API (deleted webphone's duplicate local brand layer — the SDK owns the brands); vendorHash re-pinned (`sha256-MrjH…`, matched CI's value); error-code registry regenerated.
5. **Gates** — full local suite green (24 packages; server 548 s under host load ~95); CI GREEN run 38689670760 (race suite, nix build, treefmt, art-dupl); art-dupl baseline 65 (twin capability assertions accepted with ruling in the dedup registry); `nix fmt` idempotent; smoke 47+4+8+9 zero-fail (loopback + ASR scenarios).
6. **G1 verdict GREEN** recorded in the plan §7.
7. **Housekeeping** — CHANGELOG `[Unreleased]` (durable dedup, contacts read model, `:memory:` hardening); TODO_LIST train row (next: T09); AGENTS.md readmodels-seam paragraph; lessons.md (daemon-bisection worktree rule, SIGQUIT forensics + `:memory:` fix, scan-ceiling-100, assertion-to-type-parameter); dedup-registry sweep line.

## Partially done

- **buildflow**: full run completed; visible tail = host-load timing regressions only + a PRE-EXISTING stale-baseline warning (189 accepted entries match nothing — predates this train). The findings-verdict line was still being collected when the report was demanded; treat buildflow as green-pending-confirmation (CI's own lint/build legs were green).
- **T08 byte-parity golden (F30b)**: API/parity suites green and ports done; an explicit byte-level golden file for list payloads was not added as a standalone test.

## Not started

T09 snippets swap → T10 G2 verdict → T11–T12 faxes (`ApplyIdempotent`) → T13–T14 transcripts → T15 enroll tokens → T16–T18 metrics/doc sweep → T19 sessions spike → T20 threads spike → T21 dogfood-findings memo (2 candidates already logged) → T22 close. Three carried owner questions unanswered (hook-guard hole, G2 hard-stop, upload-test `-short`).

## Mistakes (self-caught, all recovered)

- Two-step "PLACEHOLDER-DELETE" deletion in store_test.go (clumsy; final state clean).
- `store_ErrNotFoundForJoin()` placeholder in the new family test (fixed one step later).
- Multiple daemon race rejections on edits (re-View + re-apply each time).
- whitespace-drift exit-1 on my staged formatter-emitted alignment — verified false positive (formatter idempotent; CI gates content and was green).

## Next (top of the queue)

T09 snippets (the contacts pattern is proven); then T10 G2 verdict respecting the owner's hard-stop answer; buildflow verdict collection; investigate the commit-msg guard hole.
