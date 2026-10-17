# Round-5 execution status — carve train landed, data-model batch 1 in flight

**Written:** 2026-10-17 23:24 CEST · **Session start:** ~21:00 · **Mandate:** "execute the WHOLE round-5 Pareto plan, do not stop"
**Scope of this report:** this session only (CI check → carve train E7/E8 → E9 partial). Predecessors: 10-17 20:16 + 20:34 status reports; the plan itself (`docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md`).

## Session result in one paragraph

The round-5 plan was dispatched with owner gates intact; I executed every assistant-executable item in plan order and got through the biggest one completely: the **`internal/server` carve train (E7 phases 1+2) is DONE and verified end-to-end** (full suites + session-behavior suite + live smoke 47+4+8+9, 0 failed), with E8 (session_api) adjudicated to stay on dependency-map evidence rather than moved mechanically. **E9 (data-model batch 1) is code-complete + 2 of 5 test files written but NOT yet verified** (no test run after the final edits; the full buildflow gate I started in the background was never harvested — shell evaporated). Owner-gated items (E1 sitting, E2–E6 deploy/cutover cluster, E24–E26 terminals) are untouched by design. **12 commits sit UNPUSHED** — the daemon's push leg has stalled past the ratified manual-push bar; I did not invoke the bar because the dispatch ordered report-then-wait.

## a) FULLY DONE (verified)

1. **CI verdict for the round-5 plan push** — run 39060422698 (`7c4fcbc`): `completed / success` (4m12s). The pending poll from the previous session is closed.
2. **E7 phase 1 — `internal/server/hooks` carve** (daemon commits `d13ed31`/`bef333e`/`be9d709`):
   - `webhooks.go` + `idempotency.go` moved verbatim into the new package; exported surface: `DedupWindow`, `DedupCollection`, `Gate` (Seen/Record), `NewGate`, `CallIdempotencyTTL`, `BearerGate`, `RetryAfterHint`, `Handlers`/`Deps`/`New`, `SecretGate`, `UnreadCache` (satisfied by `unreadCache.Drop`).
   - Wiring keeps mount + flood budgets; `app/dedup.go` adapter + sweeper retarget `hooks.DedupWindow`/`hooks.DedupCollection`; `sipDirectoryGate` relocated to `internal_directory.go` over `hooks.BearerGate`; proxy uses `hooks.RetryAfterHint`.
   - Tests moved with code (applyStatusWebhook contract, idemStore/engine-gate, hook decode fuzz, webhookFail redaction); the 401-writer contract test now scans `../hooks`; error-code registry regenerated (`webhook.fax_pages` → new path).
   - **Verified:** `go test ./internal/server/... ./internal/app/... ./internal/arch/...` all ok (server 547s), arch green after registry regen.
3. **E7 phase 2 — `internal/server/api` carve** (daemon commits `5322be8`/`8bf1af3`/`d9edb9e` + narrative `947cc93`):
   - The five TRUE leaves moved: csrf (`RefreshCSRF`), contacts (`List/Save/DeleteContact` + shared `ContactSaveFailed`), calls (`LogCall` + `recordCallIdem`), transcribe (`Transcribe`), transcripts (`Save/DeleteTranscript`).
   - `requireSession`→`RequireSession` (27 sites), `notifyContactsChanged`→`NotifyContactsChanged` (5 sites) — the wiring hub implements the narrow `api.Server` seam so the 401 wording and the SSE nudge keep one home.
   - `callsIdem` left the handlers struct into `api.Deps`; contacts fuzz + recordCallIdem contract tests moved (with a `contextSessionGate` test stub); contract test asserts ZERO inline 401s in `../api`.
   - Branching-flow: all 42 findings reviewed individually; the only new one (`api.Server` IFACE_COMPLETE) got the DedupWindow-precedent in-source rationale; baseline re-saved for the moved HOST-ABSOLUTE paths; `-s branching-flow` exit 0.
   - **Verified:** server/api/hooks/app/arch suites green; `nix fmt` idempotent.
4. **E8 phase-3 ADJUDICATION (executed as a decision, documented in the TODO row)** — `session_api` STAYS: `verifyDirectoryCredentials`/`mintSession`/`sessionIdentityResponse` are the shared spine of BOTH login modes (passkey_api reuses all three) and the handlers ride wiring-private state (cookies, csrfCfg); moving the three HTTP handlers alone splits the auth subsystem behind an 8-leg interface. Honest future move = both login modes together (own train). `passkey_settings` (view-coupled: partial re-render, toast, lang) and `internal_directory` (machine surface over wiring-owned seams) also stay. Post-carve: server = 19 non-test files (under the 20-file trigger), hooks +2, api +6. TODO row rewritten to 🟢 DONE with the full record (commit `043f6c7`).
5. **Carve end-to-end proof** — session-behavior suite (`-run TestServer`) ok; fresh binary built; **live smoke: 47 passed + restart 4 + boot-failure 8 + asr 9, 0 failed**.
6. **CHANGELOG [Unreleased]** — two carve bullets landed (phase 1 + phase 2 incl. the adjudication).

## b) PARTIALLY DONE (E9 — data-model batch 1, F37–F42)

Code landed (daemon commit `cf138ca`, 8 files); tests partially landed (`ce7318f`):

| Item | State |
| --- | --- |
| F37 Must* enum parsers (Direction/Channel/OutboundStatus incl. legal zero) in `domain/message.go` | ✅ code + `domain/enums_test.go` |
| F38 fax leg: `Parse/MustFaxStatus/FaxDirection` in `domain/fax.go`; `legacy_faxes.go` refactored onto them (reject-on-error, one home) | ✅ code (corrupt-row panic test = `store/scan_panic_test.go` ✅) |
| F39 transcripts: `legacy_transcripts.go` onto `ParseDirection`; island boundary `SaveTranscript` uses `domain.ParseDirection` | ✅ code |
| F40 `retention_days < 0` rejected in `config.validate` (`config.retention_days`) | ✅ code, ❌ test missing |
| F41 gateway exhaustive dispatch: unknown mode PANICS (both constructors), Loopback fallback dropped | ✅ code, ❌ test missing |
| F42 `ParsePhone` at the transcript POST boundary (stores the SANITIZED remote; 400 on non-dialable) | ✅ code, ❌ test missing (garbage-remote 400 + sanitized round-trip cases) |
| **Build + test run after the final edits** | ❌ NOT DONE — last verified state predates the E9 edits |
| `nix fmt` + gates + narrative commit | ❌ NOT DONE |

**The full buildflow gate I started (post-carve) was never harvested** — the background shell vanished before I read it; its verdict is unknown/lost (see d).

## c) NOT STARTED (this session; from the round-5 plan)

- **Owner-gated (cannot self-execute):** E1 sitting (38 rows + 10 questions), E3 pbx deploy, E4 journal watch, E5 SMS-422 triage, E6 auth cutover T13/T15, E24 announcements, E25 mic ritual, E26 harness disposition, E18 markdownlint posture (row-18 verdict), E19 d2 upstream (Q3).
- **Assistant-executable, not reached:** E10 (brands + constructors), E11 (audits + perf), E12 (island CALL_TRANSITIONS), E14 (three smalls), E15/E16 (ASR hardening + polish), E17 (gate-recovery stragglers), E20 (branching-flow upstream drafts), E21/E22/E23 residual folds (only the carve-row fold of E21/E23's scope is done), E2 (stack relock — cross-repo, sequenced behind E1).
- Standing watches (CAL): calendar items, untouched by design.

## d) TOTALLY FUCKED UP (honest ledger)

1. **The post-carve full buildflow verdict is LOST.** I launched it in the background (shell 062) and never read it before the shell evaporated. The carve's gate evidence rests on the per-phase suites + smoke — good, but the FULL gate (golangci tier-1/2, erraudit, art-dupl, nix build) has no observed verdict for the current tree. Must re-run.
2. **Train-start baseline was garbage.** I started a ~25-min full buildflow at train start and edited files WHILE it ran — its "1 step failed" verdict reflects a mid-edit tree, and the failing step's name was lost to `tail -60`. Process rule violated: long gates run at boundaries, never concurrent with edits.
3. **Write-tool glitches ×3** (dropped quote in an import line, dropped `Set` argument, one stray CJK character `吞` in a comment). All caught by diagnostics/build/own review, but each was avoidable sloppiness on big pastes.
4. **The daemon ate 3 of 4 narrative stage+commit pairs** (phases 1, and both E9-adjacent attempts); only `947cc93` (CHANGELOG-only) and `043f6c7` (TODO row) survived. Content always landed via daemon `chore:` commits — nothing lost, but the history reads as opaque chunks for the carve.
5. **whitespace-drift false positive bypassed, not fixed.** The gate flagged treefmt's OWN realignment (the daemon had committed my pre-format alignment as HEAD). I diagnosed it correctly and committed with justification, but the script still cannot distinguish "hand-edited whitespace" from "formatter's correction of a raced commit" — a script-level fix or upstream note was not filed. Close-the-loop rule: this is recorded here, not yet filed/fixed.
6. **E9 left unverified at session end** — code + 2 test files landed through the daemon while 3 test files and the verification run were still pending; the tree on disk is mid-train (compiles? unproven this session for the final state).
7. **Stale gopls throughout** (deleted-file diagnostics persisted for hours; one LSP rename attempt failed with a closed connection → fell back to sed). Should have restarted the LSP after the file deletions.
8. Minor: two pre-existing-looking staticcheck SA4006 warnings surfaced in `transcripts_api_test.go` (266/278, unused `resp`) — not investigated, not mine, noted only.

## e) WHAT WE SHOULD IMPROVE

1. **Never run long gates concurrent with planned edits** — gate at boundaries only (the AGENTS train-start rule needs a "and then WAIT for it" clause in my own workflow).
2. **Harvest background jobs before moving on** — a started gate that is never read is worse than no gate (false confidence + lost evidence).
3. **Fix or file the whitespace-drift blind spot**: staged diffs that are byte-identical to `nix fmt` output are NOT drift; the script could re-run the formatter in a temp checkout to compare (fix in `scripts/whitespace-drift.sh`).
4. **Daemon-race mitigation**: stage+commit in one call is already the rule; consider committing immediately after observing a daemon commit (the daemon's quiet window right after its own commit is the least contested slot), or accept `chore:`-only history for mechanical trains and put the narrative in the TODO/CHANGELOG/lessons instead.
5. **Restart the LSP after bulk deletions/moves** — stale diagnostics cost attention all session.
6. **Write tests in the same batch as the code** for small batches (E9's three missing tests would have taken minutes in-flow).
7. **Verify-before-waiting**: never end a phase with "edits on disk, no build" — run at least `go build ./... && go test <touched pkgs>` before yielding.

## f) NEXT — up to 50, in execution order

**Immediate (unblock the current tree):**
1. Run `go build ./...` + `go test ./internal/domain/ ./internal/store/ ./internal/config/ ./internal/gateway/ ./internal/server/...` for the E9 state; fix anything red.
2. Write the three missing E9 tests: retention-days rejection (config), gateway unknown-mode panic (gateway), transcript boundary (garbage remote → 400; `" +44 1632 960961 "` → stored `+441632960961`, history round-trip).
3. `nix fmt`; regenerate the error-code registry if E9's new codes (`domain.direction` etc.) trip `TestErrorCodeRegistryIsFresh`.
4. **Re-run the FULL buildflow** (BUILDFLOW_NO_RESULT_CACHE=1) and READ it this time; fix findings forward.
5. Run `nix flake check` (sandbox + island-lint + treefmt + KVM backup VM) before any push.
6. **Invoke or explicitly waive the manual-push bar** (anchor: first unpushed `463829b`, 2026-10-17T21:03:03+02:00; stall already >2h at report time) — then watch CI to green (`gh run watch`), per the ratified convention.
7. CHANGELOG [Unreleased] line for E9 (vocabulary closes).
8. Run `bash scripts/push-stall-watch.sh --no-fetch` as the bar's formal evidence when invoked.

**E10 (brands + constructors, F43–F48):**
9. `session.Token` brand part 1: type + constructors through the Store seam.
10. `session.Token` brand part 2: `sipcred.Derive` + TestServer suite green (`auth:` prefix, session-behavior suite).
11. `domain.RelPath` type home + blob edge conversion + arch-test note.
12. `NewInboundMessage`/`NewOutboundMessage` constructors + illegal-pair test.
13. `NewInboundFax`/`NewOutboundFax` + test.
14. Collapse `UpdateStatus` loose params onto the delivery-machine value.

**E11 (audits + perf, F49–F54):**
15. Fax tests: cover both Resolution arms.
16. Dedup-registry end-to-end cross-check of cited rulings.
17. idempotency TTL-map invariant audit note + `views.Lang` 10th-enum stat fix.
18. Time-granularity doc note (millis vs s) + Lifetime single-home ruling.
19. `attachAttachments` N+1 → single IN-query + test.
20. Caps × transcriptFetchWindow table test; `ExtensionHubs` keyed by `domain.Extension`.

**E14 (three smalls):** 21. fax receipt-verbosity seam (typed levels + tests). 22. config fail-closed twin-option struct validation + boot test. 23. schema_version gate (table + first migration step wired).

**E15/E16 (ASR):** 24. segment retry/backoff cap + unavailable toast. 25. provider latency metric + optional non-critical health leg. 26. memory ceiling for concurrent live transcriptions. 27. language hint from `wp-lang` + `asr.language` wire. 28. client transcript DOM cap + per-call scroll/clear affordance. 29. transcript download leg. 30. voicemail transcribe-button test (#37). 31. 429 combined load test (#5).

**E12 (D-train close):** 32. island CALL_TRANSITIONS emission points + `state.js` map. 33. island test + DOM-contract conditional check. 34. DOMAIN_LANGUAGE ratification memo fold (gated on sitting g3).

**E17 (stragglers):** 35. ui-capture 14-shot pass on the v1.20.1 tree. 36. module-cache janitor + zero-byte .tmp sweep. 37. errorpage tagging discipline upstream note.

**E20–E23 folds:** 38. branching-flow drafts finalized + filed upstream (verify-before-filing first). 39. d2 cheat-sheet verify-on-v0.9.0 + skills-repo PR (gated Q3). 40. 10-14 status 26-followup routing audit. 41. carve-row number refresh (two-method count) — partially done in the row rewrite; verify numbers. 42. metaengine row close (verify G3-via-2.9.0, F019/F026 hints). 43. lessons folds: self-verifying scripts, verify-before-describing. 44. snapshot-rot header convention. 45. GOPROXY fallback + BuildFlow retry (BuildFlow repo). 46. whitespace-drift script fix or upstream filing (from e.3). 47. investigate the two SA4006 warnings in transcripts_api_test.

**Watch items (not tasks):** 48. CI green after the eventual push (verify with `gh run list`, not push logs). 49. Stack browser E2E owed only on served-markup changes — the carve changed none; confirm no E2E debt accrued. 50. Daemon push leg health — if the stall persists across sessions, the pma unit's push leg may need the same attention the commit hook got.

## g) Questions I cannot answer myself

1. **Next-train sequencing (standing question 1, now concrete):** E9 is one verification away from done. Continue straight into E10→E11 (data-model series, plan order), or insert E14 (three smalls) / E15–E16 (ASR) first? The plan orders data-model; the sitting was to ratify — ratify or invert?
2. **Adjudication ratification:** I closed the carve WITHOUT moving `session_api`/`passkey_settings`/`internal_directory` (dependency-map veto; record in the TODO row), and proposed the honest future carve = BOTH login modes together into an auth package as its own train on the next auth-surface change. Ratify, or do you want the mechanical session_api move anyway under the `auth:` convention?
3. **Push decision now:** the manual-push bar is MET (anchor `463829b` at 21:03, stall >2h, 12 unpushed commits) but the tree is mid-E9 and the full gate is unobserved. Push immediately after item f.1–f.5 verify (my recommendation), or hold the push until E9's tests + full gates are green in one verified train?

## Artifacts this session

- Code: `internal/server/hooks/` (7 files), `internal/server/api/` (8 files), rewiring across `server.go`/`pages.go`/`actions.go`/`proxy.go`/`internal_directory.go`/`unread.go`/`panels.go`/`app/dedup.go`, E9 deltas in `domain/message.go`/`fax.go`, `store/messages.go`/`legacy_faxes.go`/`legacy_transcripts.go`, `config/config.go`, `gateway/gateway.go`, `server/api/transcripts_api.go`.
- Tests: moved+adapted (hooks ×5, api ×2), new contract-scan legs, `domain/enums_test.go`, `store/scan_panic_test.go`.
- Docs: TODO_LIST carve row (🟢 DONE + adjudication), CHANGELOG ×2 bullets, error-contract registry regen, `.buildflow-baseline.json` refresh.
- Commits (local, unpushed at report time): `463829b`…`ce7318f` — 12 total, incl. narratives `947cc93`, `043f6c7`.
