# Status: branching-flow re-triage + gate enablement (2026-10-11 ~06:00)

Executed the 2026-10-11 ratification sitting's open condition: "the
`.buildflow.yml` whole-tool skip survives until the finding set is
re-triaged into targeted nolints + a baseline." Input: the owner-pasted
03:40 `branching-flow all . --format finding` report (40 rows) — which
turned out to be PARTIAL: its package load hit
`/mnt/buildcache … no space left on device` and dropped cmd/ plus the
panic rows (61 files scanned either way; the panic family reappeared on
the post-edit full run). Every site was re-read at HEAD before any
edit; the full set is 45 rows.

## What landed

| Disposition | Rows | Mechanism |
| ----------- | ---- | --------- |
| Fixed honestly | 2 | `likeEscape`/`SearchThreads` unified on `needle` (the transcripts.go:94 collision died by vocabulary unification, not types); `server.EventName` was the earlier 02:19 fix |
| nolint-suppressed | 26 | statement/field/signature-line `//nolint:branching-flow // reason` directives, each citing its registry/ruling source |
| Baselined | 19 | type-node findings (12 COMPOSITION_mixin + 5 DUPLICATE_TYPE + 2 ANTI_PATTERN_large-struct) in `.buildflow-baseline.json`, each with a plain `branching-flow ... (baselined)` rationale comment at the site |

The 26 nolints cover: the vcard nil-guard FP, bootreport/transcripts
index FPs (bounds conditions and the map-to-slice invariant), the two
go:embed panic invariants, the three `App.Start` InvokeNamed
composition-root resolutions (AGENTS doctrine), nine transpose-constructor
pairs (single composition-root caller + boot-time both-or-neither
validation), the two i18n key-plumbing collisions, and the four CallID
strong-id rows (opaque island token, verbatim passthrough — the 09-18
pbx-UUID class; wire DTOs stay raw).

## Suppression mechanics discovered (the load-bearing findings)

1. `//nolint:branching-flow // reason` WORKS for findings reported on
   statement, field, and signature lines (the apiContact.ID precedent
   generalized; verified live on the full set).
2. It does NOT work for TypeDecl-node findings (COMPOSITION_mixin,
   DUPLICATE_TYPE, ANTI_PATTERN_large-struct) in v0.6.4. Five placements
   tested and all ignored: inline on the decl line, standalone comment
   line above, first-field line, bare directive (no reason), and a
   `//nolint:start:branching-flow` block. Those findings therefore ride
   the baseline, with the ruling text kept at the site as a plain
   comment so the next triage reads it without opening this file.
3. CLI and SDK emit DIFFERENT finding-ID shapes: the standalone CLI
   relativizes paths (`internal/...`, DUPLICATE_TYPE without column),
   the buildflow-embedded SDK absolutizes them
   (`/home/lars/.../internal/...`). A baseline written from CLI output
   never matches the SDK's gate (observed: "none of the 19 baseline
   entries matched any remaining finding"). The baseline must be saved
   from the SDK side: `buildflow -s branching-flow --save-baseline`.

## The gate

- `.buildflow.yml`: `branching-flow` REMOVED from `skip_steps`;
  `strict: true` + `baseline: .buildflow-baseline.json` added.
- Semantics achieved: strict promotes the findings gate to warning
  severity for every tool; the baseline accepts the 19 reviewed
  type-node findings → the gate fails on NEW findings only, fleet-wide
  (not just branching-flow).
- `error_rules` (finer per-rule elevation, exists in unreleased
  BuildFlow source) is UNKNOWN to the installed binary (25b07c4) and
  warns at every load — left OUT until the installed binary knows it.
- Verified: focused step exit 0 with 19 accepted findings; negative
  test — removing one baseline entry trips the gate (restored);
  full pipeline 64/64 steps green under `--strict`.
- Brittleness (accepted): baseline IDs embed host-absolute paths. Every
  buildflow consumer on this repo runs from the canonical checkout on
  this host (CI's go-tests job does not run buildflow; release.sh and
  the dev loop are host-local). A moved checkout re-baselines with one
  command (AGENTS.md documents it).

## Rulings consumed (no re-litigation)

All dispositions trace to standing rulings: the 10-05 mixin doctrine
(field-shape similarity is architectural ROLE), the 10-05 dupe triage
(boundary trio; CRM/Paperless twin = owner ruling 5: two seams stay,
noise routed upstream), the 09-18/09-22 phantom policy class, the
02:11 02:19 report triage, and the strong-id wire-DTO ruling.

## Upstream (routed, per ruling 5 and the FP policy)

branching-flow (private) gets the three findings of this session:
**#12** TypeDecl nolint blindness — composition + dupe analyzers never
consult the ignore-comment system (five-placement evidence + source grep;
the enum also lacks a dupe rule name), **#13** the CLI↔SDK finding-ID
divergence (absolute vs relative paths; baselines must be saved from the
SDK side), both filed with provenance banners. The panic-FP shapes were
ALREADY routed as **#9** by the earlier session (covers the vcard
closure guard, the map-to-slice invariant, and the two-pointer swap).
Drafts: docs/drafts/2026-10-11_branching-flow-{composition-dupe-suppression,id-shape-cli-vs-sdk}.md.

## CHANGELOG

Silent by the ratified bar: comment-only suppressions, a config-only
gate change, and a behavior-neutral param rename are not code-path
changes. This doc + the registry sweep line are the record.
