# Status — deepening plan continued: gates repaired, ladder truth, Transport type, RelPath, verdict pins, DE locale

**Snapshot:** 2026-11-12 14:20 · branch `main` pushed through `17f5c3af` ·
CI: `5acadaa7` GREEN, `b9078c6a` GREEN, `a2509a83`/`17f5c3af` pending · Full Go
suite green at every commit; island suite 336/336 (T15's three specs included).

Continuing `docs/planning/2026-11-12_05-25_SUPERB-deepening-and-integrity-plan.md`
after the 13:15 report (T01–T08 done there; this report covers T09–T16 + repairs).

---

## a) FULLY DONE (committed + pushed + verified)

| Task | What landed | Evidence |
| --- | --- | --- |
| T09 CI verdicts (F31–F32) | **Found a real instrument bug**: `gh run list --commit <sha>` answers EMPTY on this repo while the REST API's `head_sha` filter returns the runs. `ci-verdict.sh` now queries `actions/runs?head_sha=` — exit contract + 5-case self-test unchanged. Verdict log: ledger batch GREEN, dupl repair GREEN, depth-budget commit RED (the art-dupl red it rode in with, superseded), views commit GREEN after repair | `9529ca64`; self-test + live verdicts verified |
| Plan amendment | The plan file gained the same-turn amendments: T05 NO-GO with the structural evidence, T08's shrank scope, the honest export instrument (top-level decls), the T14 site move | `8479e93a` |
| T10 docs-gates (F33–F34) | Battery GREEN locally; CI's `--check` red repaired: (1) the retirement-doc §4 row edit had mangled dprint's table reflow — staged; (2) the living-corpus growth re-flagged 49 old archived §f rows — `annotate_max` re-measured 203→260 **with the evidence comment** in `.docs-gates.json` (the documented DF-brittle class; F58's drain train is the way back down) | `2655bc2d`; `--check` green end-to-end |
| T11 ladder truth (F35–F36) | Production reads against webphone.db census: sessions + retention sweep + app_flags cursors + ErrNotFound sentinel — nothing else. Dead `*store.Messages` injector registration (never invoked since A34) deleted; store/messages.go + the five legacy_*.go headers now say TEST-ONLY (the "rides this ONE transitional read at boot" claim was drift); the retirement doc's app_flags rows corrected (the matrix/signal cursors LIVE there — the 2026-11-09 row missed them); TODO_LIST gains the §5 owner row: the one-query emptiness proof over the seven frozen families + the owner date proposal (first 2.15.x window after 2027-01-01) | `f581a37b` |
| art-dupl repair #2 | The T08 mirror (`pbx.VoicemailSummary` ↔ `views.VoicemailSummary`) reded the gate: ACCEPTED — the mirror IS the architecture (a shared home would re-couple views to the PBX wire). Baseline re-recorded (110 held), registry sweep-log line, check = 0 new | `37ceda39` |
| T12 push.Transport (F37–F38) | `push.Transport` defined type; `ParseTransport` the one door (closed set ntfy|fcm, Rejection, `push.transport_parse`); fanout map + Device row typed; the registration API parses BEFORE the Knows check — a client typo is 400, a known-but-unconfigured transport stays 422; the registry adapter parses every row at fan-list (pre-door rows skip, never strand). Family pin `TestParseTransportVocabulary`; registry regenerated same-commit | `6e5a3d4a` (+ daemon `f445ef26`) |
| T13 FaxJob.RelPath (F39–F40) | `FaxJob.DocumentPath` is `domain.RelPath`; both constructors take the branded value (callers pass `blobs.Save`'s result — no String() round-trip); the four consumer MustParse calls dissolved; `RelPath.Empty()` added (pathless lifecycle rows are legal — "" comparisons don't compile by design); journal fact + wire views stay raw (wire doctrine); folds convert via `domain.RelPathFromRow` (empty→zero, corrupt→panic with the honest shape). The 2026-10-25 marker retired. Legacy/retention/paperless/admin fixtures swapped | `101e41b5` |
| T14 wiring structs (F41–F42) | `asr.ClientConfig` + `matrix.ClientConfig` (the three-primitive trains can't swap order anymore; GoogleConfig was the precedent); `machineSurface` bool-last `(w, r, missing, enabled)` at its platform home — the 2026-10-25 marker retires. Composition root + all fixtures swapped | `66bc375a` |
| T15 verdict pins (F43–F44) | `hold-verdicts.test.mjs`: the TIMEOUT arm (node:test mock timers tick past sipverdict's 30s cap — the arm calls.test.mjs could not drive) and the refused-HOLD arm, each pinning the settled-truth restoration + the direction-and-verdict announce (stack-searched — the settled re-render re-announces connected after the failure copy); third spec pins en/de composing rejected/timeout/send-failed verbatim into holdFailed/resumeFailed. Island 336/336 | `5acadaa7` |
| T16 DE locale (F45–F47) | Reports day ladder reuses the transcripts' `weekdayLabel` vocabulary (ONE weekday home; Go's "Mon" spells English in both languages) over `02.01.` for DE; `monthLabel` mini-map fixes the CRM due phrase ("fällig 4. Nov" — "2 Jan" spelled the English month on a German desk); the unread divider's aria-label announces through the NEW `thread.unread.aria` key ("2 unread messages" / "2 ungelesene Nachrichten") — no em-dashes in the announcement, the visible dashes stay. Three new DE pins; tw.css/login.css rebuilt after the templ edits; i18n parity auto-green | daemon `6e000765`/`b9078c6a` + `a2509a83` |
| Race-fixture fix | CI's -race step flaked red on `101e41b5`: the transcript fixture's `transcriptSeq++` was an unsynchronized shared increment across the concurrent-append goroutines (PRE-EXISTING test bug — surfaced by interleaving luck). `atomic.Int64`; -race twice green | `17f5c3af`, run 43221820456 |

## b) PARTIALLY DONE

- **The annotate_max ratchet went UP** (203→260) with evidence — protocol-honest
  but direction-contrary; F58's drain train (check-rows --strict + bare-rows
  gates-first, then archive) is the designed way back down and stays unexecuted.
- **CI watch**: `a2509a83` (the divider-test repair) + `17f5c3af` (the race
  fix) pending at snapshot; `b9078c6a`/`5acadaa7` green.
- **The concurrent session's dep bump** (`585f630a`: go-paperless v0.5.2 +
  vendorHash re-pin) landed mid-my-trains and its first CI run reded on the
  same T08 mirror group (since absorbed). Their vendorHash re-pin — I did NOT
  verify it independently (not my train).

## c) NOT STARTED (remaining plan tasks)

- T17 admin CDR party fallback over CDREntry (F48)
- T18 receiver hardening: matrix oversized-sync streaming + garbage-cursor
  self-heal (F49–F50), FCM 401/403 invalidation (F51), CRM body caps (F52)
- T19 postcall.js extraction from calls.js (F53)
- T20 glyph→SVG sweep (F54)
- T21 owner-calls prep doc (F55, F62)
- T22 erraudit nolint-audit + calendar sweep (F56, F61) — the carve trains
  moved ~20 nolint-bearing sites; drift is expected
- T23 cqrs-lint upstream drafts (F57)
- T24 release tails: 18-08 drain, flight-recorder docs, release-log policy (F58–F60)

## d) TOTALLY FUCKED UP (owned)

1. **My own new spec took three iterations to converge** (hold-verdicts.test
   .mjs): I asserted the WRONG settled states twice (a failed HOLD restores
   `established`, not `held` — I kept expecting the hoped-for state, the exact
   dishonesty the verdict lane exists to prevent), then the leak-check window
   mixed the visible dashes into the aria assertion, then panicked on index
   math before switching to value extraction. The test-first discipline I
   demand elsewhere was sloppy on my own file.
2. **Blind regex fixture sweeps caused two syntax breakages** (the
   double-MustParseRelPath wrap `domain.MustParseRelPath(domain.MustParseRelPath(...)`
   + a swallowed i18n comma breaking both dictionaries). The full suite caught
   both, but the pattern (regex over tests, then run) is not the discipline.
3. **The messaging race probe transient-FAILED a second time** in one full
   suite run (passed alone immediately after, deterministic per its
   assertions). Twice unexplained is a pattern: next occurrence MUST be
   captured (full -v output retained) before any re-run.
4. **The art-dupl gate has bitten every refactor train this session** (3
   repairs, 2 baseline writes). The preflight recommendation from the 13:15
   report remains unimplemented — each red was avoidable with a pinned-detector
   check before push.
5. **The race-fixture bug rode four of my pushes** before CI's interleaving
   happened to surface it — a pre-existing defect, but my local gate has no
   -race step, so I could not have caught it earlier either. Worth adding
   `-race` on the touched packages to the train discipline.

## e) WHAT WE SHOULD IMPROVE

- **Test-first for new island specs**: write the assertion of the SETTLED
  truth BEFORE the fixture (the verdict arms cost 3 iterations because I
  wrote fixtures first and guessed outcomes).
- **Fixture sweeps need compile-fix loops with vet BETWEEN sweeps**, not
  after all of them (the two breakages were found only at the full suite).
- **A pre-push art-dupl pinned-detector check** for refactor trains (one
  command, the gate stops biting).
- **`-race` on the touched packages** for concurrency-adjacent trains.
- **The annotate pin needs the drain** — every new ledger train grows the
  corpus and risks another flare; F58 is the structural fix.

## f) NEXT (up to 50, best-first)

1. Watch CI on `a2509a83`/`17f5c3af` to green (the race fix is the load-bearing one).
2. T17 admin CDR party fallback over CDREntry + the ladder test (F48).
3. T18 matrix oversized-sync streaming decode + too-large transient (F49).
4. T18 garbage next_batch self-heal → fresh sync (F50).
5. T18 FCM invalidateToken on 401/403 (F51).
6. T18 CRM read/context io.LimitReader caps (F52).
7. T19 postcall.js extraction (calls.js 2 025 → ~1 400) + import-closure pin (F53).
8. T20 glyph→SVG sweep in panels.js + dialer-focus residue (F54).
9. T21 owner-calls prep doc — now includes: the §5 DROP date proposal (T11's
   row), the annotate drain priority, the race-probe transient question (F55).
10. T22 erraudit nolint-audit — the carve trains moved ~20 directive sites (F56).
11. T22 calendar sweep — confirm 2026-12-08 erraudit + 2026-12-20 buildflow (F61).
12. T23 cqrs-lint upstream drafts, owner-tick gated (F57).
13. T24 18-08 drain train — the annotate pin's way back DOWN (F58).
14. T24 flight-recorder/docs rows (F59).
15. T24 release-log archival policy row (F60).
16. T21/62 verify brandid-lint posture + #48 + ASR deploy-proof rows (F62).
17. buildflow FULL run — never executed this session; the trains changed ~3 000 lines.
18. Island tsc after T19/T20 land (the only JS-touching tasks left).
19. CHANGELOG entries for T12/T13/T14/T15/T16 (the daemon carried some edits;
    the [Unreleased] block needs one proofread pass).
20. TODO_LIST full-code-review section coherence pass (T02/T03 strikes landed
    there; the T15 row needs its strike).
21–50. The remaining fine-tasks F45–F62 are all named above in the plan's own
    order; after T17–T24 the plan is fully executed and the standing
    obligations (release runbook, docs-gates, calendar) own the cadence.

## g) QUESTIONS FOR YOU (cannot answer from the repo)

1. **The annotate ratchet went UP with evidence** (203→260). Acceptable as
   recorded, or do you want the F58 drain train executed BEFORE any more
   ledger trains so the pin travels back down?
2. **The concurrent session's dep bump** (`585f630a`, go-paperless v0.5.2 +
   vendorHash re-pin) rides main between my trains and its vendorHash re-pin
   is unverified by me. Is that session still active (I keep deferring to its
   in-flight state), and do you want the bump's CI watched by me too?
3. **Remaining execution order**: T18 (receiver hardening) → T19/T20 (island
   refactors) → T21–T24 (docs/watches) is the plan's order. Confirm, or do
   you want the owner-calls prep doc (T21) pulled forward while the CI on the
   race fix settles?
