# SUPERB Phase 3 second train (A22–A24) — landed, A25 closeout interrupted mid-gates

> claims verified as of 15:48 CET, 2026-11-08 (push/CI statements rot fastest; timestamp the claims, don't present them as timeless)

Session scope: resolve the two owner gate questions, then continue the SUPERB
plan (`docs/planning/2026-11-08_04-04_SUPERB-command-event-sourcing.md`).
Executed: owner rulings ①②, Phase 3's A22 + A23 + A24 in full, most of A25.
The status request arrived while A25's final full buildflow run was still owed.

## a) FULLY DONE (verified this session)

- **CI verdict on the prior docs push (53929848): GREEN** (run 42477798797,
  completed success) — the unawaited risk from last session is closed. The code
  run before it (42476804599) also green.
- **Owner ruling ① — templ "use LATEST":** latest IS v0.3.1100. The proxy's
  `@latest` AND GitHub's tag list agree; the feared "v0.3.1114 CLI" exists as
  no tag anywhere. The tree is already fully consistent: go.mod pin = CLI pin
  = all 15 `*_templ.go` codegen markers at v0.3.1100. No bump needed; the
  pin rule (CLI must equal go.mod) stands. Documented in the plan Errata.
- **Owner ruling ② — CBOR quantization: event-sequence tiebreaks ONLY.**
  Root cause confirmed in `/home/lars/projects/go-codec`: both CBOR codecs
  ride `cbor.TimeUnixDynamic` — README: "float64 epoch with sub-second
  precision (9 bytes)" — a deliberate upstream design choice (~200 ns ulp at
  2026 epoch). Integer-epoch payload fields REJECTED: they fork one timestamp
  into two representations (split brain) to win fidelity no surface renders.
  If µs+ fidelity ever matters, the fix home is a `TimeUnixNano` mode
  upstream in go-codec. Documented in the plan Errata (both rulings, one block).
- **A22 — enroll tokens answer from the journal** (`internal/cqrs/enroll.go`):
  one stream per token HASH (SHA-256 hex is a legal `ParseStreamID` key; no
  owner axis by construction); commands `enroll.mint`/`enroll.burn`; the
  consume-once burn decides an INERT no-op for unknown/already-burned hashes
  (zero-event decisions verified legal dispatches — empirically pinned); the
  mint-side opportunistic sweep survives as burn commands on each expired
  token's own stream (burned-set guard so projection lag cannot re-dispatch);
  expiry classification STAYS at the userauth door (Lookup exposes it); NO
  importer (15-min TTL re-mint doctrine). Seam: `readmodels.EnrollGate`;
  `userauth.New` signature swapped; app wires `CqrsService.EnrollSvc()`.
  5 tests in `internal/cqrs/enroll_test.go` (round-trip, burn-inert,
  sweep-on-mint, expired-still-resolves, restart-reload) — all green.
- **A23 — agent-audit answers from the journal**
  (`internal/cqrs/agentaudit.go`): per-owner APPEND-ONLY stream; audit events
  carry actor metadata (`id.ParseActorID("user:<email>")` — a UserID itself
  must be a ULID, so the email rides as the user-kind raw) and command
  causation (`event.WithCausation("agent_audit.append", cmd.ID())`); the
  500-row window trim moved INTO the decision (append + expunges in ONE event
  batch; state already bounded, no scan loop); Seq stays unix-nanos (int64 —
  exact under CBOR). Seam: `readmodels.AgentAuditGate`; server `Deps.AgentAudit`
  swapped; Settings' newest-20 unchanged. Importer `importAgentAudit`
  (marker `agent_audit_journal_imported`, tail-sync AppendBatch, ids/actors/
  seq preserved). 3 aggregate tests + 1 app-level import test — green.
- **A24 — push devices answer from the journal**
  (`internal/cqrs/pushdevices.go`): per-owner stream; the 8-per-owner cap for
  NEW device ids and rotation-replaces semantics live in the DECISION;
  `ErrListFull` passes through dispatch verbatim (errors.Is); the wake token
  stays write-only (LIST wire shape untouched — that mapping stays at the API
  door). Seam: `readmodels.PushDeviceGate`, taken by server.Deps, api.Deps AND
  hooks.Deps (the ring hook). Importer `importPushDevices` (marker
  `push_devices_journal_imported` — ring continuity is the whole point).
  4 aggregate tests (incl. rotation-at-cap + restart) + 1 import test — green.
- **Cross-cutting, all green:** error-code registry regenerated (+6 enroll,
  +4 agent-audit, +6 push_device codes in `docs/error-contract.md`);
  CHANGELOG `[Unreleased]` entry; plan Errata "Phase 3 second train" notes;
  full Go suite `go test -count=1 ./...` green; session-behavior suite
  (`TestServer`) green; userauth + app + push + hooks + readmodels packages
  green; `nix fmt` applied; **live smoke FULLY green against the nix binary**:
  smoke 51/51, restart 4/4, boot-failure 8/8, dialgate 9/9, faxcap 8/8.
- **Dedup:** the three `ExportAll` methods extracted into ONE home
  (`readmodels.exportAllViews`, `internal/readmodels/export.go`); the six
  remaining new groups adjudicated as the Phase-2 migration-twin classes and
  baselined with the CI-pinned art-dupl v0.7.2 (112 → 118, post-write
  check = 0 new); sweep-log row appended to `docs/dedup-registry.md`.

## b) PARTIALLY DONE

- **A25 (Phase-3 closeout) — interrupted mid-gates by this status request:**
  - DONE: full Go suite, session suite, registry, art-dupl, fmt, smoke (all green).
  - FIXED JUST NOW: the full buildflow run had **2 failed steps**; the real one
    was `test-compile` (go vet): my new `export.go` passed a VARIABLE as the
    format string to `WrapInfrastructuref` — fixed to the non-f
    `WrapInfrastructure`; `go vet` + `buildflow -s test-compile` now green.
    (The other failure was the known-flaky `nix-build-verify` class — 88%
    historical fail rate, the failed-only rerun passed it.)
  - STILL OWED: one clean full `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` run;
    the narrative `auth:` commit for the train remainder; push; CI-watch.
- **Commit hygiene:** the daemon landed most of the train in heuristic chunks
  (25081489, 3b77e10f, b7d9a254, 2fcd0c35 — all already pushed per
  `git log origin/main..HEAD` = empty). Still uncommitted/staged: the dedup
  baseline + registry row, the export.go extraction + vet fix, the three
  ExportAll reroutes. No narrative commit survived for most of the train.

## c) NOT STARTED

- **A25 final full-gate run + push + CI-watch** (the remaining sliver).
- **Phase 4 (A26–A30):** contacts aggregate (cap 500, rename-keeps-id, never-
  sweep), transcripts aggregate (per-owner 5000 trim stays, per-CALL floor
  keeps the loop — nested-cap gap untouched), call notes + reviews aggregates
  (notes 409 Conflict, reviews 2000 Rejection), P4 importers + markers +
  dual-run diffs, P4 gates + release train.
- **Phase 5 (A31–A38) — the mountain:** core stream design doc (per-thread
  streams keyed owner+network+remote, `AppendToThread` as the decider's apply,
  Conflict-retry R7), delivery machine integration (`DeliveryAdvance` verdicts
  become decisions), message projections (byte-identical fragments, stable
  ids), `messaging.Service` swap, THE importer (A35, webphone.db
  threads/messages/attachments), retention port (A36 — **BLOCKED on upstream
  go-cqrs-lite#118 trim capability** per the A05 finding), core parity
  battery, P5 gates + stack browser E2E.
- **Phase 6 (A39–A40):** fax aggregate (`faxMoves` in the decider,
  provider_ref dedup → durable command idempotency), fax importer + retention
  + release.
- **Phase 7 (A41–A47):** handler unification sweep, contracts/catalog diffs,
  dedup migration to `sys.Dedup()` + readmodels.db death certificate,
  sessions/app_flags ruling execution (D2), ladder freeze, doctrine rewrite
  (AGENTS.md no-ES bullet → ES doctrine etc.), final battery + release.

## d) TOTALLY FUCKED UP (found, owned, fixed or fenced)

- **The export.go vet defect (mine):** `WrapInfrastructuref(err, code, what)`
  with a variable `what` — `go build` was green, only buildflow's `test-compile`
  (go vet) caught it. Fixed same session; step re-verified green. LESSON:
  any `*f` wrapper needs a constant format string; run `go vet` (or the
  buildflow step) after EVERY new helper file, not only `go build`.
- **Smoke binary exec outside the devshell:** `/nix/store/.../webphone` exists
  (verified ELF bytes) yet direct exec failed with FileNotFoundError / "not
  found" — running inside `nix develop -c` works fine (interpreter/env
  resolution). Fenced by always running the smoke via
  `nix develop -c python3 scripts/webphone-smoke.py`. Two smoke attempts
  burned on this before the fence.
- **Stale LSP diagnostics:** `props.ASR` errors in messages.templ persisted in
  every tool result all session. Verified harmless (package builds, CI green,
  field present since 7d4a078c) — vtsls staleness, not the tree. Noise only,
  but it polluted every diagnostic banner.

## e) WHAT WE SHOULD IMPROVE

1. **Narrative commits keep losing the race.** AGENTS.md already says
   stage+commit in ONE call with verification; I staged but deferred the
   narrative commit — the daemon chunked the train. Next trains: commit the
   narrative at each phase boundary IMMEDIATELY after the phase's tests pass.
2. **`go vet` is not in my local loop.** The repo's own bar caught it, but an
   hour earlier costs less than a red full run. Add `go vet ./...` to the
   per-aggregate verification trio (build/test/vet).
3. **Zero-event dispatch needed empirical proving** — it works (pinned by
   tests now), and the finding is in the plan Errata so no future aggregate
   re-litigates it.
4. **Smoke needs the devshell** — should be written into AGENTS.md commands
   (currently only implied by the nix develop prefix).

## f) NEXT — the queue (ranked, ~47 plan items remain)

| # | Item | Size | Note |
|---|------|------|------|
| 1 | A25 finish: full no-cache buildflow green + narrative commit + push + CI-watch | S | the sliver left |
| 2 | A26 contacts aggregate (flagship parity: cap 500, rename-keeps-id+createdAt, never-sweep) | M | P4 |
| 3 | A27 transcripts aggregate (5000/owner trim stays; per-CALL floor keeps loop) | M | P4 |
| 4 | A28 call notes + reviews aggregates (409 Conflict vs 2000 Rejection) | M | P4 |
| 5 | A29 P4 importers ×4 + markers + dual-run diffs | M | P4 |
| 6 | A30 P4 gates + release train (one release per phase, D4) | M | P4 |
| 7 | A31 core stream design doc (per-thread streams; AppendToThread as ONE Decide; R7 retry) | H | P5 map |
| 8 | A32 delivery machine integration (DeliveryAdvance → decisions; self-send 422 arm preserved) | H | P5 |
| 9 | A33 message projections (byte-identical fragments, `wp-bubble`/`wp-thread-row` pins) | H | P5 |
| 10 | A34 messaging.Service swap (Receive/Send/flags dispatch; nudge rules intact) | H | P5 |
| 11 | A35 THE importer (webphone.db → streams; pre-cutover snapshot; dual-run diff) | H | the risky moment |
| 12 | A36 retention port — **BLOCKED on upstream go-cqrs-lite#118** (trim/Tombstone) | H | needs upstream ship + bump |
| 13 | A37 core parity battery (owner_scoping, admin_reads, threads_flags, byte-pins) | H | P5 |
| 14 | A38 P5 gates + release + stack browser E2E (445s budget) | H | P5 ship |
| 15 | A39 fax aggregate (faxMoves decider; provider_ref dedup → durable idempotency) | M | P6 |
| 16 | A40 fax importer + retention + P6 release | M | P6 |
| 17 | A41 handler unification sweep (every write surface dispatches) | M | P7 |
| 18 | A42 contracts/catalog diff pins (openapi byte check, asyncapi, EventCatalog) | M | P7 |
| 19 | A43 dedup migration to sys.Dedup() + readmodels.db death certificate | M | P7 |
| 20 | A44 sessions + app_flags ruling per D2 (`auth:` + session suite) | M | P7 |
| 21 | A45 ladder freeze (no new store/db.go steps; retirement plan doc) | S | P7 |
| 22 | A46 doctrine rewrite (AGENTS.md no-ES → ES; machine-tables → deciders; fleet-libs re-score) | M | P7 |
| 23 | A47 final battery + release (full flake check, VM tests WITHOUT readmodels.db, vulnix, version bump, stack re-pin) | H | 100% shipped |
| 24 | Regenerate `docs/status/INDEX.md` after this report (done in this commit) | S | hygiene |
| 25 | Monitor upstream go-cqrs-lite#118 (the A36 unblocker); file/ping if idle | S | unblock A36 |

(Items beyond 25 are the per-task micro-steps already enumerated in the plan's
Table B — not duplicated here.)

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **A36/retention upstream dependency:** go-cqrs-lite#118 (journal trim) is
   the P5 deletion gate. When Phase 5 reaches A36 and #118 still hasn't
   shipped: do you want me to (a) implement it myself in go-cqrs-lite and
   release (I own that repo), (b) defer A36 and ship P5 without the
   retention/GDPR arm (the §4.9 P5 abort says don't), or (c) park the whole
   Phase-5 train until upstream moves?
2. **One release per phase (D4) with the daemon racing commits:** Phases land
   as many daemon chunks + at most one surviving narrative commit. Should the
   version bump + tag for each phase release happen at phase END (current
   reading), or do you want a single release at the end of P7 covering the
   whole SUPERB migration?
3. **Concurrent-session coordination:** the other session bumped
   loginpage→v4.17.0 / cqrs-htmx→v4.18.1 / usermgmt→v4.20.0 and is doing CRM
   create-form work mid-train. My phases touch app.go/server.go readmodels
   seams they also edit. Do you want me to keep working straight through
   (current posture: additive merges, never revert, re-View before edit), or
   pause between phases when their activity touches the same files?
