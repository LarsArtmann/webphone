# Status: Metaengine Train — Faxes + Transcripts Swaps, G2 GO, M10 Closed Elsewhere (2026-10-15 19:13 CEST)

**Scope of this session:** M1–M7 of the remainder Pareto plan
(`docs/planning/2026-10-15_17-14_SUPERB-metaengine-train-remainder-pareto.md`),
i.e. T09 snippets, the G2 verdict, T11+T12 faxes, the T16 census interim,
and T13+T14 transcripts — plus the discovery that another session closed
M10 (guard hole). NOT done yet: M8 enroll, M9 upstream filings, M11
buildflow hygiene, M12 docs sweep, M13 G3 close, the M14–M18 tail.

## TL;DR

The engine pilot is now THREE seams deep (contacts, snippets, faxes) plus
a fourth (transcripts) — the two hardest parity surfaces (faxes' declarative
lifecycle, transcripts' caps + search + reassembly) landed green, smoke
holds at 47+4+8+9 zero-fail throughout, and **G2 is recorded GO** in the
train plan §7. Remaining SQL sites under the stated census rule: **40 →
lower after enroll** (see below). main is green through the fax push
(CI 38771614852); the transcripts push (CI 38775269720) was in progress at
report time with the full local suite already green (server leg 548 s).

## What landed

1. **T09 snippets (M1, μ01–μ12)** — narrative commit `197ee02`, CI
   38764205893. Read-model facts (`SnippetSaved/SnippetDeleted`),
   folds, seam with the exact SQL contract (replace-keeps-creation-time,
   owner-scoped oldest-first list, cross-owner delete = ErrNotFound, cap
   100 → `ErrSnippetListFull` homed in readmodels). Legacy snippets ride
   the T07 `app_flags` one-time boot import (flag `legacy_snippets_imported`).
   `store/snippets.go` deleted; tests ported not deleted (cap leg, replace
   leg, restart durability, scoping, family pins).

2. **T10/G2 VERDICT: GO (M2, μ13–μ16)** — narrative commit `953f81a`.
   Recorded in train plan §7 with: census rule stated (ExecContext/
   QueryContext/QueryRowContext/listRows call sites in internal/store
   production files; 46 pre-train → 44 post-snippets), an HONEST note that
   no pre-train latency baseline existed (plan gap) plus the recorded
   post-swap reference (boot 26–67 ms; `/api/contacts` 0.7 ms empty /
   1.13 ms @ 58 rows; partials ≤0.8 ms), and the rollback posture. TODO_LIST
   train row updated. The 108→≤73 headline used a broader counting rule;
   the stated rule is the measurement of record going forward.

3. **T11+T12 faxes (M3+M4, μ17–μ29)** — landed across daemon commits
   (swept before I could narrative-commit; the story lives here + plan).
   - Facts: `FaxReceipted` (insert), `FaxStatusAdvanced` (status move —
     NAMED AWAY from the plan's "FaxAdvanced" shorthand: that name is the
     existing SSE nudge event), `FaxExpunged` (retention).
   - The update fold is a real prev-merging fold — the provider-ref
     keep-when-empty rule lives in ONE place.
   - **faxMoves stays the only legality home**: the projection is a dumb
     store (parity with the blind SQL UPDATE); a test drives every
     faxMoves row through the seam plus the deferred-resolution,
     replay-refresh, and late-reversal arms.
   - **`AdvanceIdempotent`**: the webhook lane rides the engine's
     in-process idempotent apply (`faxadv:<ref>:<status>`). Documented
     honestly in the seam: ApplyIdempotent is process-local by design;
     DURABLE replay protection remains the hooks seam's engine-backed
     dedup window (T04) — no false durability claims.
   - **Retention moved with the data**: `Faxes.SweepBefore` expunges and
     returns spooled document paths for unlinking; the faxes metric counts
     via the engine aggregate (the dead-table freeze/500 class closed for
     faxes like T07 closed it for contacts).
   - Legacy fax import (flag `legacy_faxes_imported`, corrupt rows
     skipped: unknown lifecycle enums cannot poison the projection).
   - `fax.Service` now takes a `Store` interface (framework-free);
     `store/faxes.go` deleted; all test helpers ported to the engine.

4. **T16 interim census (M5, μ30–μ31)** — 44 → **40 sites** post-faxes
   (−5 fax store, +1 transitional legacy reader). Latency smoke within
   noise of the recorded reference. Final number lands after enroll (T15).

5. **T13+T14 transcripts (M6+M7, μ32–μ39)** — the hardest parity surface,
   green across all suites.
   - Facts: `TranscriptAppended` / `TranscriptExpunged`; segments are
     KEYED (SQL used a private autoincrement; the seam mints
     `GenerateTranscriptSegmentKey`) and carry `Seq` (append-order nanos
     — the rowid tiebreak's engine equivalent).
   - Caps ported with a semantic difference the SQL got for free: the
     SQL trim was ONE range-delete statement; the engine trim is a
     converging loop of bounded scans + key expunges (each pass removes
     ≥1). Both trims (owner 5000, per-call 1500) proven by the ported
     flood tests (bulk-seeded straight through the engine, mirroring the
     old raw-INSERT seed).
   - Search is a seam-side ASCII-case-insensitive contains over the same
     fetch window (the SQL LIKE's escaping is moot for a Contains scan;
     metacharacters literal — test-pinned).
   - Reassembly (`reassembleTranscripts`/`reverseLines`) moved verbatim
     from the deleted SQL file; Recent/Search/Export share one
     fetchSegments home; DeleteCall idempotent; Count + SweepBefore
     engine arms wired into metrics and retention (retention.Start now
     takes the engine and drives both sweeps).
   - Legacy transcript import (flag `legacy_transcripts_imported`).
   - `store/transcripts.go` + its test file deleted; suite ported.

## What I got wrong (fixed in-session)

1. **Test split-brain**: my first server-test rewire gave the notifier and
   the Deps TWO different engine instances — `TestSSEPushesSwapSafeFragments`
   caught it (fax push rendered the empty state). One engine everywhere now.
2. **Trim bugs, two layers**: a slice-bounds panic (no len guard), and the
   deeper one — a single bounded scan cannot see the excess BEYOND cap+1,
   so one pass left 1502/1500 rows. The converging loop fixes it; note the
   SQL statement was strictly more powerful per pass. Ported tests caught
   both (the flood seeds are load-bearing).
3. Smaller: wrong import path (metaengine vs /v4), helper name collision
   (`newTranscripts`), a botched double-seed in the flood test (caught
   pre-commit), repeated edit-tool failures from daemon touches (adapted:
   python re-read-and-apply).

## Done by someone else (verified, not duplicated)

**M10 guard hole (μ44–μ46) is CLOSED** — `scripts/commit-msg-guard-selftest.sh`
(3/3 pass) + the TODO_LIST row: the hook NEVER bound daemon commits because
the pma systemd unit exports `GIT_CONFIG core.hooksPath=/var/empty` (env
config beats repo + global). The hook's own logic is verified; the fix is
an OWNER edit of the nix-config pma module (fleet-wide). I did not touch it.

## State

- HEAD pushed; clean tree. CI: fax push GREEN (38771614852); transcripts
  push in progress (38775269720) with the full local suite green (server
  548 s), smoke 47+4+8+9 zero-fail, art-dupl clean at the NEW baseline 66
  (one added group: the three sibling legacy-import skip-shapes in
  `app/legacy.go` — intentional symmetry; registry ruling line owed in M12).
- **Owed in M12**: dedup-registry ruling for baseline 66, CHANGELOG
  `[Unreleased]` lines for faxes+transcripts, AGENTS.md readmodels
  paragraph update (it currently stops at T07), TODO_LIST progress row.
- **NOT started**: M8 enroll tokens, M9 upstream filings (needs
  verify-before-filing + github-voice), M11 buildflow full run + the stale
  `.buildflow-baseline.json` (189 dead entries), M13 G3 (release dance:
  version bump, CHANGELOG cut, vulnix, tag, push, CI, stack notification),
  M14–M18 tail (sessions memo, threads spike, v5 watch, upload `-short`
  (owner approval already given in the standing order), fuzz perf, backup
  verify, postgres proof).
- Watch item: the not-mine `.buildflow.yml` working-tree change
  (build_mode fast→full + un-suppressing cqrs-lint/nix-hash-fix/gomod-check)
  was swept into daemon commit c81f17f and is now on main — I had planned
  to leave it uncommitted; the daemon outran me. gomod-check is the
  documented vendor-consistency false positive; expect findings noise at
  M11's buildflow run.

---

## Appendix (19:35 CEST) — owner answers + M8 started, then WAITING

**Owner decisions received (both questions answered):**

1. **G3 scope: "Everything (plan default)"** — after M8+M9+M11+M12, run the
   complete G3 release INCLUDING the tag and the stack notification.
2. **M9: "Yes, file now"** — both go-cqrs-lite issues filed upstream with
   the evidence (scan-ceiling vs WithLimit semantics with the 1506-row
   trim-scan observation; check-then-apply cap ergonomics), links into
   TODO_LIST Standing watches.

**M8 (T15 enroll tokens) — first half landed, second half owed:**

- DONE: `domain/enroll.go` facts (`EnrollTokenKey` keyed by the SHA-256
  hex, `EnrollTokenMinted`, `EnrollTokenBurned`); `readmodels/enroll.go`
  seam — hash-keyed `Mint/Lookup/Burn`, expired rows swept opportunistically
  on every mint (the SQL mint's preamble, now a bounded scan+expunge);
  engine wired + Plan'd, readmodels+userauth suites green. Deliberate
  deviation, documented in the seam: **NO legacy import** — tokens live
  15 minutes, an upgrade carries no meaningful continuity (in-flight
  links get re-minted; the old `wp_enroll_tokens` table in usermgmt.db
  goes dead storage).
- OWED (next commit): `userauth/enroll.go` rewrite onto the seam
  (classification and all `userauth.enroll.*` codes/copy stay in
  userauth — parity of the error contract), `userauth.New` gains the
  store param, `app.go` wires `engine.EnrollTokens()`, external tests
  seed expiry through the returned store handle, gates, then the
  **`auth:` commit prefix + session-behavior suite** (passkey surface —
  AGENTS auth-regression convention).
- The seam-half commit was swept by the daemon (messageless, as usual
  for sweeps); the narrative lives here and in the next commit.

**Then (per the owner's answers):** M9 filings (verify-before-filing +
github-voice skills first), M11 buildflow full run + stale baseline,
M12 docs sweep (incl. the dedup-registry ruling for baseline 66 and the
AGENTS readmodels paragraph), M13 full G3 release, then the M14–M18 tail.

**State at wait:** clean tree, all M8-partial work committed (daemon
sweep), main green through CI 38775790916. WAITING for the owner's go.
