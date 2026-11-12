# Deepening plan T17–T24 execution — session close-out

> Claims verified as of 18:53 local, 2026-11-12. Push/CI statements rot
> fastest; the shas and run ids are inline. This closes the 05-25
> SUPERB deepening-and-integrity plan's second half (T17–T24; T01–T16
> closed in the 13-15 and 14-20 reports).

## a) FULLY DONE (committed + pushed + verified)

- **T17 — admin CDR party fallback (F48)**: `cdrParty` (CallerIDName →
  CallerIDNumber → unknown) now leads the admin call-record row, and the
  section's whole ladder is pinned by a structural view test
  (`TestAdminCDRSectionLadderIsStructural`: off explains itself and
  renders no rows, a failed line is named while survivors render, empty
  gets the honest empty state). `e897efe8`; views + admin server suites
  green.
- **T18 — receiver hardening (F49–F52)**: matrix answers cap at 16 MiB
  with a DISTINCT `matrix.too_large` Transient (the old 1 MiB read
  truncated real syncs into a confusing Infrastructure decode error);
  the receiver self-heals a POISONED cursor (400 M_UNKNOWN — resets to a
  fresh sync instead of stopping loudly on self-inflicted Rejections)
  and an OVERSIZED answer (reset rides the backoff ladder, no hot-spin);
  `homeserverError` carries status/errcode for deliberate reactions; FCM
  drops the cached OAuth grant on 401/403 (cache test proves the next
  send mints fresh); CRM lookup/context reads cap at 1 MiB with
  `crm.too_large` (family pins, registry rows landed same-commit).
  Source carried by daemon commits `17022f1a`/`632f0a07`, tests
  `d53a4753`, tail `d15dcc89` (the resetCursor dedup the art-dupl gate
  demanded — extracted, not baselined).
- **T19 — postcall.js extraction (F53)**: the ended-reason mapping, note
  editor, disposition chips and AI-draft affordances moved to
  `postcall.js` (calls.js 2 025 → 1 693; `q850CauseOf`'s one home moved
  with it); three specs re-pointed; import-closure pin, island suite
  336/336, tsc green. `e5fcccfc`. One dead `authedFetch` import survived
  the move and was caught by the island-lint gate ONLY in CI — fixed in
  `033cacb1`; lesson recorded below.
- **T20 — glyph→SVG sweep (F54)**: the history row's redial/save and the
  contacts row's call/remove buttons carry the island's `wp-icon` SVG
  vocabulary (`b1b01b87`). Deliberately out of scope: the shell's pinned
  "✕ Close" label (shell.test.mjs pins the bytes) and the OS-notification
  titles.
- **T21 — owner-calls batch (F55/F62)**: Q7–Q17 fill the batch doc's
  stub section — the CRM trio, cqrs-lint gating posture, C044
  permanence, B5 authority, traces-in-backup, ASR first-live gates, the
  blob crash-window sweep, the storage-ladder DROP date, the race-probe
  flake policy — each with an evidence link and a default (`e9dfc02e`).
  F62: 75 brandid-lint directives swept for the placement contract — 1
  real displacement fixed (matrix ClientConfig; the two
  `settings_templ.go` flags are generator reflow of correctly-authored
  sources); #48 and the ASR deploy-proof rows confirmed still
  owner-held.
- **T22 — nolint-audit + calendar sweep (F56/F61)**: `erraudit
  nolint-audit` = 69 directives, 69 needed, **0 stale** (the carve
  trains decayed nothing); the dated obligations re-verified true
  (erraudit/buildflow tier re-measure 2026-12-08; docs-gates cadence
  65/45d green at 64 live reports).
- **T23 — cqrs-lint upstream drafts (F57)**: VERIFIED COMPLETE, not
  re-done — both drafts already exist in BuildFlow `docs/drafts/`
  (suppression block-comment state machine; D005 first-token
  misattribution), each with released-binary repro evidence; the owner
  tick gates the filing.
- **CI red forensics this session** (three reds, all diagnosed, all
  healed): `17f5c3af`/`e897efe8` red = docs-gates (the unknown
  `_annotate_max_evidence` config key + a cadence breach — see §c);
  `e5fcccfc` red = the dead import above; `b1b01b87` red = the same
  already-healed class (its tree predates the fix).

## b) PARTIALLY DONE

- **T24 release tails (F58–F60)**: the cadence gate FORCED the first
  drain tranche — three reports archived gates-first (05-22 trains,
  06-40 m05 closeout, 12-52 appkit re-review), the flagged residue
  harvested into TODO_LIST (the island small-hardening row), live
  corpus 67 → 64 (≤65). F59's FEATURES `/metrics` row now names the
  latency histogram + beacon vocabulary (the traces-backup half is
  owner Q13). F60's release-log POLICY landed in the runbook (raw logs
  stay out of git; verdict tails go into the filing report).
  **REMAINS**: the full 18-08 archived-corpus drain — the annotate pin
  260's way back DOWN — before the ~2026-12-07 cadence window.
- **CHANGELOG coverage**: T17–T20 all carry [Unreleased] entries; the
  proofread pass across the whole block (the daemon co-mingles other
  sessions' entries) is still owed before the next release cut.

## c) THE DOCS-GATES RED, FIXED AT THE ROOT (the session's first hour)

The day's CI red (`17f5c3af` onward) was the docs battery, not code:
the `_annotate_max_evidence` key I added to `.docs-gates.json` was
never a known key (warning-only noise — removed), the freshness gate
fails on ANY unharvested item in the NEWEST report (no ratchet — my own
14-20 report's T18/T24 next-step enumerations), and cadence had breached
65 live reports. Fixed by harvest + drain, NOT pin raises: the open
work got a TODO_LIST home (the deepening-plan section), one flagged
item got routed with a strike, three fully-struck reports archived
gates-first (`check-rows --strict` + `bare-rows` verified BEFORE each
`git mv`). All nine gates green in CI mode afterward.

## d) TOTALLY FUCKED UP (owned)

- **The `e5fcccfc` dead import cost a CI red that local gates could not
  see.** tsc, the island suite and the closure pin all pass with an
  unused import; only the flake's island-lint (oxlint no-unused-vars)
  catches it, and that gate runs nowhere locally in my flow. Lesson:
  an island module MOVE owes `nix build
  .#checks.x86_64-linux.island-lint` before push — the same
  before-attempt-1 discipline the release runbook already ratified for
  releases.
- **The family-test imports I scripted went up unformatted** (python
  import-inserts ignored gofmt order; `nix fmt` healed the worktree but
  the daemon snapshotted pre-fmt bytes twice) — two CI reds of the same
  mechanical class. Lesson: after ANY scripted multi-file edit, fmt +
  stage + commit in ONE call before anything else can observe the tree.

## e) WHAT WE SHOULD IMPROVE

- **A pre-push "island gate trio" script** (tsc + island suite +
  island-lint) for any train that moves served JS — one command would
  have caught the dead import locally.
- **`docs-gates freshness` in repair flows**: harvest-at-write-time (a
  report's next-step items into TODO_LIST the moment the report is
  written) instead of harvesting after the gate reds — the DF-brittle
  flare is now a documented recurring cost, and this session paid it
  twice.

## f) NEXT (up to 50, best-first)

1. Watch CI green on `7dfcab3e`/`2ecd1831` (the tail pushes; the
   pending runs at 18:53 cover the whole healed chain).
2. The full 18-08 drain train (T24 remainder) — annotate pin 260 back
   DOWN before ~2026-12-07.
3. The CHANGELOG [Unreleased] proofread before the next release cut.
4. Owner sitting: Q7–Q17 in the batch doc (eleven defaults ready to
   ratify or reverse in one pass).
5. The island small-hardening residue row (4 items, all small).
6. `buildflow format` quiet-window run (the concurrent session's
   gate-setup row 2; not mine to self-schedule).
7. T23's two drafts: file when the owner ticks the banners.

## g) QUESTIONS FOR YOU (cannot answer from the repo)

None new — eleven sit in the batch doc (Q7–Q17, each with a default),
plus the standing M18–M21 gated recipes from this morning's batch.
