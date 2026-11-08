# Status: SUPERB Phase 0 executed — all eight spikes GO, gate A09 recorded, D1–D8 ruled

> **Snapshot rot:** claims verified as of **2026-11-08 08:52 CET** (CI verdict on the latest push still `in_progress` at write time — re-check `gh run view 42460315090` before relying on it).
> **Scope:** this report covers THIS session's run only — Phase 0 (A01–A09) of the [SUPERB command & event sourcing plan](../planning/2026-11-08_04-04_SUPERB-command-event-sourcing.md). Verified via: live commands (module-cache source reads, the hello-system spike build/runs, the cqrs-lint worktree run, docs-gates battery), git log/gh runs, and the committed docs themselves. NOT verified: the concurrent sessions' work (the 07:19/07:38 CI reds are ATTRIBUTED to their commits, not investigated), the in-flight CI run's outcome, and anything prior to this session.

## a) FULLY DONE

1. **Skills loaded before action:** go-cqrs-lite (SKILL + module sources), buildflow, github-voice (+ duplicate search + checker), status-report.
2. **A01 — system/v4 API spike:** deep-read of the PUBLISHED `system/v4 v4.17.0` (not the 50-commits-ahead local checkout): `New`/`Start`/`Close`/`GracefulClose`, DomainConfig/DeploymentConfig, RegisterDecider/RegisterCommand/`Op`/`Execute`, ONE "projections" adapter wrapping the whole store, `RawQuery(QueryDecl)` passthrough, introspection surface. Doc: `docs/research/2026-11-08_system-v4-api-notes.md`.
3. **A02 — sqlite/nix proof:** `CGO_ENABLED=0` static ELF; `sqliteengine` registers "sqlite" over modernc (same driver family as webphone); module delta = system/v4 + `koanf/parsers/yaml` ONLY, plus the discovered **metaengine v4.21.0 → v4.22.0 floor**. Journal+projection+checkpoints in ONE 164 KiB `system.db`.
4. **A06 — read-your-writes + replay spike** (`/tmp/hello-system`, dial-window-shaped, mirrors `internal/readmodels/dialwindows.go` verbatim): dispatch 0.9–1.1 ms; projection visible 1.1–2.0 ms AFTER dispatch returns (NOT synchronous — v1's retraction upheld); **replay rebuild 1.6 ms** via the documented host `Stop → ResetProjection("projections") → Start` recipe; Close idempotent (nil, nil). D7 mechanism decided: answer-from-decision + bounded catch-up await (250 ms budget) before SSE nudges.
5. **A04 — cqrs-lint inventory:** isolated `git worktree` on throwaway branch `spike/cqrs-lint-inventory` (deleted clean after): 56 findings, ALL info-severity; **E023 = exactly 3 sites** (app/chat.go, readmodels.go, server.go), each migrating with its phase — no baselines needed, no surprise reds possible.
6. **A03 — composition contract:** `docs/planning/2026-11-08_cqrs-composition-contract.md` — system as ONE named eager service `"cqrs"` in the do container; boot inserts `sys.Start` before the listener; shutdown rides the cascade via `GracefulClose(10s)` with zero cross-file ordering constraints; Deps threading + health wiring designed.
7. **A05 — retention + GDPR policy doc** (`docs/planning/2026-11-08_journal-retention-policy.md`): classes table, the finding that **journal trim does not exist upstream** (interfaces carry no delete; tombstone is metadata-only), the upstream capability spec, the sweep-port design, D8 snapshot thresholds. **Owner has since scoped the follow-through OUT of my lane ("not my worry") — see §g Q1.**
8. **A07 — backup spec:** `docs/planning/2026-11-08_systemdb-backup-spec.md` — 6 integration points (backup-script.nix guard-if pattern, vm-test assertions, module-check, README, drill, runbook), per-phase readmodels shrink rule, restore-is-stronger-than-tables semantics.
9. **A08 — importer strategy:** `docs/planning/2026-11-08_es-importer-strategy.md` — extracted the shipped legacy-contacts pattern; journal-specific designs: resumable-by-construction (deterministic stream IDs + per-stream tail sync), timestamp fidelity (payload stamps, never journal order), raw-append synthetic events with importer identity, dual-run diff harness as a gate, per-aggregate notes incl. enroll/agent-audit exemptions.
10. **A09 — gate recorded:** plan Errata now carries Phase 0 verdicts + owner GO (granted in-session: "execute the whole list") + **D1–D8 ruled per recommendation** (CBOR; sessions-migrate-P7; retention classes; one release per phase; toolchain unchanged; ONE system.db; bounded catch-up await; EveryNEvents(200)).
11. **Upstream filing:** [go-cqrs-lite#118](https://github.com/LarsArtmann/go-cqrs-lite/issues/118) — journal trim capability (TrimStream + TrimJournalBefore), voice-checked (0 FAIL), duplicate-searched, verified at published-source level first. Draft preserved at `docs/drafts/2026-11-08_gocqrs-lite-journal-trim-issue.md`.
12. **docs-gates:** annotate pin 205→206 raised WITH evidence (verified my 6 new docs carry zero findings first — the +1 is the documented DF-brittleness, same class as the sanctioned 71→72); full battery GREEN (fmt, freshness, audit-all, routing, cadence, table-damage, strike-safety).
13. **Landed + pushed:** all Phase 0 docs on `main` and on the remote (daemon sweep commits `6908eaad`/`781d7a52` carried the docs; my `e762f5af` carried the pin + narrative). New CI run `42460315090` started 07:49.

## b) PARTIALLY DONE

- **CI green-verification of my push:** run `42460315090` was `in_progress` at 08:52 — outcome unverified (see §d).
- **A05 scope:** the deliverable (policy doc + upstream issue) is done, but the owner's "not my worry" correction arrived AFTER completion; whether the §4.2 deletion invariant survives for Phase 5 is OPEN (§g Q1).
- **Narrative commit hygiene:** content 100% landed, but the daemon swept 7 files before my atomic commit — my explicit commit carried only `.docs-gates.json`. Authorship messy, content intact.

## c) NOT STARTED

- **Phase 1 (A10–A14):** the dependency train (system/v4 + metaengine v4.22.0, vendorHash re-pin), `internal/app/cqrs.go`, shutdown-order/DLQ/health wiring, byte-parity boot smoke, full gates.
- **Phases 2–7 (A15–A47):** dial windows end-to-end, small/medium aggregates, the core (threads/messages/attachments), faxes, unification + readmodels.db retirement + doctrine rewrite.
- **TODO_LIST harvest** of Phase-0 completion rows + reopen-registry updates (this report §f feeds it).
- **Spike preservation:** `/tmp/hello-system` dies with /tmp — the research doc carries reproduction, but the Go shapes should be ported into a repo test during Phase 1.
- CHANGELOG entry: correctly OWED NOTHING (docs-only delta — changelog bar).

## d) TOTALLY FUCKED UP (process stumbles; no deliverable damage)

1. **Pushed onto a RED main without checking first:** the 07:38 run (42459589773, `fast flake checks` failure) belongs to a concurrent session's openapi-golden fix; my 07:49 push stacked on top. The repo's own bar is check-THEN-push; I checked only after. If my run inherits/exposes that red, attribution must precede any "fix".
2. **Lost the narrative commit to the daemon:** wrote 7 files, then batched one commit — the daemon committed them first (the exact AGENTS race, mitigated only by speed). Lesson honored too late.
3. **Spike iteration waste:** reopen phase needed 2 fix rounds (System.Start once-only; Reset-while-running) — both documented in code I had already read; reading lifecycle docs BEFORE writing would have saved both.
4. **Worktree/tidy ordering:** stub written after tidy dropped system/v4 (re-add needed); first `git worktree add main` tripped the in-use fatal (branch from HEAD fixed it).
5. **One `nix develop` from /tmp without a flake** (expected error, 1 wasted call).

## e) WHAT WE SHOULD IMPROVE

- **Pre-push CI check as a hard gate** (`gh run list --workflow CI --limit 1` BEFORE `git push`) — this session proves the cost of skipping it.
- **Commit-per-deliverable** (or explicitly accept daemon authorship and verify content on the remote instead of narrative commits).
- **Read-then-write for lifecycle-sensitive spike code** — the upstream sources were already in my context both times.
- **Scope-check design work against owner appetite** when an area is flagged "not my worry" — A05's policy doc was plan-prescribed and is now potentially over-built relative to the owner's lane ruling.
- **Todo-list discipline:** I marked A05 completed before the scope-down arrived; statuses should track owner rulings, not just artifact existence.

## f) NEXT (up to 50, priority-ordered)

1. Watch CI run `42460315090` to a verdict; if red, attribute (mine vs inherited) before touching anything.
2. Rule §g Q2 (red-main ownership: the 07:19 island + 07:38 openapi reds are concurrent sessions' commits).
3. P1/A10: manual dependency train — `go get system/v4@v4.17.0` (metaengine floor v4.22.0 rides along), `go mod tidy`, vendorHash re-pin via `nix build .#webphone.goModules --rebuild`, push, CI-watch.
4. P1/A11: `internal/app/cqrs.go` per the contract (named `"cqrs"`, eager, hollow-root WARN logged).
5. P1/A11: lifecycle guard asserts + unit test (resolve named, Close idempotent).
6. P1/A12: place `GracefulClose(10s)` in the cascade + shutdown-order test (fake system records sequence).
7. P1/A12: DLQ wiring (`WithDeadLetterStore`) + readable DLQ surface + poison-event ⇒ worker-named health test.
8. P1/A13: `/healthz` cqrs leg (HealthCheck + WorkerStatus + LagPerProjection, 30s lag bound) + `/startupz` join + nil-fallback for test compositions.
9. P1/A14: TRAIN-START full buildflow (`BUILDFLOW_NO_RESULT_CACHE=1`) — the plan's §4.3 obligation.
10. P1/A14: boot-smoke byte-diff of every served page vs pre-change.
11. P1/A14: full smoke + `nix flake check` + train-end buildflow + CHANGELOG + narrative commit + push + CI-watch.
12. A07 Phase-1 slice: `backup-script.nix` system.db guard (both branches) + vm-test assertions + module-check + README + drill guard.
13. Preserve the spike: port hello-system shapes into a repo-side Go test before /tmp wipes.
14. P2/A15: dial commands + decider + events (rejection shapes preserved: 404 off-latch, 502 store failure).
15. P2/A16: register decider; dial folds via `RawQuery` (existing declarations verbatim); remove dial collections from the readmodels plan (first shrink) + drill update.
16. P2/A17: port `dial.Service` to dispatch; `/api/dials` handler diff nil; `/config.js` pointer semantics kept.
17. P2/A18: dial importer (rows → events, marker, crash-retry, dual-run diff, timestamp fidelity test).
18. P2/A19: `TestDialGate*` ports, `systemscenario` spec, wiring parity, command p95 benchmark vs SQL path, `/metrics` unchanged check.
19. P2/A20: `dialgate` smoke + full gates + release.
20. P3/A21: snippets aggregate (cap 100 decider-side; fill-replaces-textarea parity).
21. P3/A22: enroll commands (expiry stays at the userauth door; importer EXEMPT documented).
22. P3/A23: agent-audit stream (Appended events only) + actor/causality enrichers; `auth:` prefix + session suite.
23. P3/A24: push devices (8/owner cap; token write-only projection; `push.*` family pins).
24. P3/A25: smoke additions + gates + release.
25. P4/A26: contacts aggregate (rename-keeps-id, delete tombstone, cap Rejection).
26. P4/A27: transcripts streams per (owner, call) + floor loop unchanged.
27. P4/A28: notes (409 mapping) + reviews (wholesale replace, 2000 cap).
28. P4/A29: importers for contacts/transcripts/notes/reviews (markers + diffs).
29. P4/A30: `admin_reads_test` + owner-scoping pins + gates + release.
30. P5/A31: core design doc (stream keys, payloads, folds, `AppendToThread` as decider apply, conflict-retry spec).
31. P5/A32: delivery decider consuming `deliveryMoves`; self-send Rejection AFTER persist.
32. P5/A33: thread-list + transcript + attachment projections; snapshot wiring per D8.
33. P5/A34: `messaging.Service` Receive/Send/flags → dispatch; `ChangeFunc` emission preserved.
34. P5/A35: two-part core importer + full-success marker + crash-retry + dual-run + pre-cutover snapshot.
35. P5/A36: sweep port (tombstone+trim per D3) + deletion gate + `retention_days=0` test — **pending §g Q1 scope ruling**.
36. P5/A37–A38: parity suites + stack browser E2E + perf re-check + release.
37. P6/A39: fax decider consuming `faxMoves` + durable provider_ref idempotency.
38. P6/A40: fax importer + `faxcap` smoke + release.
39. P7/A41: sweep remaining write handlers to the dispatcher + `internal/arch` import-direction rules.
40. P7/A42: `/openapi.json` byte-diff + `webphone contracts` regen + catalog render diff.
41. P7/A43: `hooksIdem`/`CallsIdem` → `sys.Dedup()`; readmodels engine registration removed; restore-path test WITHOUT readmodels.db.
42. P7/A44: D2 sessions ruling execution + session-behavior suite.
43. P7/A45: ladder freeze + retirement plan doc.
44. P7/A46: doctrine rewrite (AGENTS/ROADMAP/TODO_LIST/FEATURES/CHANGELOG) + fleet-libs `go-cqrs-lite.md` re-score.
45. P7/A47: FULL no-cache buildflow + flake check + VM drills + vulnix + final E2E + version release + stack re-pin note.
46. docs-health HARVEST: this report's §f → TODO_LIST; Phase-0 completion rows; reopen-trigger registry refresh.
47. Re-measure docs ratchets at the next sweep (audit_max/annotate_max DF-brittleness is documented and LIVE — my errata flipped one archived row).
48. Owner lane: upstream #118 implementation decision (§g Q3) — blocks P5's deletion gate timing either way.
49. Verify at P2: webphone-android push seam unaffected (registry moves readmodels→system; the `/api/push-devices` wire contract must stay byte-identical).
50. Release-runbook obligation at the FIRST release train: stack (nix-international-telephony) per-train lock bump + pbx-artmann relock ritual.

## g) QUESTIONS I CANNOT ANSWER MYSELF (top 3)

1. **A05 scope ruling:** "not my worry" — does that also relax the plan's §4.2 deletion invariant (Phase 5's journal-trim gate: journal may keep-forever while views+blobs delete), or does the invariant STAND and only the upstream implementation lane is yours? The answer decides whether P5's B130–B132 stays, shrinks, or rides the §4.9 abort criterion.
2. **Red-main ownership:** the 07:19 (island no-undef) and 07:38 (openapi golden) CI failures come from CONCURRENT sessions' commits. Do you want me to verify-and-fix main's red myself before starting Phase 1, or leave it to the owning sessions (repo rule: never revert others' in-flight work — but red main blocks every train's start)?
3. **Upstream #118 (journal trim):** it is your library. Will you implement TrimStream/TrimJournalBefore in go-cqrs-lite yourself, or should a webphone-side session pick it up when Phase 5 nears (either way it must ship before the deletion gate or the pre-committed P5 abort fires)?
