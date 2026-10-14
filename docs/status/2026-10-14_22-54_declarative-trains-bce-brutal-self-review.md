# Status: SUPERB trains B/C/E — brutal self-review + full session status (2026-10-14 22:54)

Session: continuation of the 2026-10-14 dispatch (same five documents;
trains A + docs hygiene landed earlier that day). This report covers
ONLY this session's run (B verification → C → E → wrap), what I got
wrong, and what follows. Owner questions stayed minted throughout.

## a) FULLY DONE (verified, CI green at `51cca0d`)

1. **Probe collected + g2 mechanism grounded.** The instrumented run
   caught `no such table: sessions` — full chain: `store.Open` sets
   `ConnMaxIdleTime(5m)` with `MaxOpenConns(1)`; the oversize subtest's
   25 MiB crawl idles the only conn for 9m1s; the idle timer closes
   it; a `:memory:` database dies with its connection; the next lookup
   opens a fresh EMPTY memory DB → fail-closed 401. Harness-only
   (production is file-backed). Attached to g2 in the 21-15 report per
   its item 24. The un-minted remainder: an optional one-line
   test-side hardening.
2. **CI `38621960858` verified** (the pending CHANGELOG push from the
   prior session) — success.
3. **Train B closed as delivered-by-prior-train** (g1's stated
   default): B1/B2/B3 verified in code WITH test coverage; backfilled
   the three real gaps — B1.5 `TestDriveBranchesOnReceiptResolution`
   (non-loopback immediate gateway settles; deferred keeps sending),
   B1.6 the 09-30 review P5 resolved-annotation, B2.3 README
   validation-contract rows. Commit `e0c5564`.
4. **Owed full buildflow for the train-A+B tree:** every gating step ✔
   (golangci, erraudit, branching-flow, art-dupl, samber-linter,
   test-compile, nix-build incl. sandboxed tests + KVM backup drill);
   exit 1 attributed solely to the documented 7/7 `nix-build-verify`
   box-degradation casualty.
5. **Train C1 — `domain.AppendToThread`:** append invariants (unread
   on inbound, last-activity follows the message, inbound unarchives,
   unconditional stamp overwrite) moved from the store's SQL CASE
   into one pure function; the append transaction became
   read-modify-write inside the same tx; pinned by the existing
   flags/scoping suites, new domain parity tables, and a new
   store-level creation-fallback test. Commit `a7a9073`.
6. **Train C2 — `faxMoves`/`FaxAdvance`:** legality table + pure
   advance; `drive` (submit-failed/accepted/settled) and
   `UpdateProviderStatus` (verdict) rewired onto it. Commit `dadb75e`.
7. **Train E1 — typed events + projection:** closed `domain.Event` set
   (`MessageReceived`, `MessageDeliveryAdvanced`, `ThreadFlagChanged`,
   `FaxAdvanced`), single `domain.ChangeFunc` seam replacing the two
   per-service shapes, `Notifier.Project` as the notification-policy
   table (message events = list+transcript; flag change = list-only BY
   DESIGN — no thread id on the event; fax = fax list). SSE payload
   bytes unchanged (existing goldens green); new
   `TestProjectNotificationPolicy` pins the table. Commits `0c53551` +
   `7d49ac1`.
8. **Train E2:** DOMAIN_LANGUAGE ratification memo — drafted,
   owner-gated, deliberately unpicked; ROADMAP cross-link restored.
9. **Train E3 docs sweep:** CHANGELOG lines (C1/C2/E1), AGENTS
   machine-tables invariant (149/377 — cap fine), plan completion
   annotation, TODO_LIST row refresh, dedup sweep-log line.
10. **End-state gates:** full go suite `./...` minus the documented
    environmental flake = 22/22 ok; fresh-binary smoke 47/0 + restart
    4/0 + boot-failure 8/0; `nix fmt` + whitespace-drift clean at every
    commit; CI `38633368094` and `38633875595` SUCCESS; local ==
    remote == `51cca0d`.

## b) PARTIALLY DONE

1. **Narrative commit hygiene.** Content complete on origin, but the
   daemon's heuristic commits split my narratives ~6×; amend/squash
   recovery worked only while unpushed. Two heuristic-named commits
   (`a87b215`, `271a627`) carry E1's core files ON ORIGIN — history
   readability scar, not fixable without force-push (prohibited).
2. **C2.5 semantics.** Shipped correct-by-tests, but my FIRST
   implementation encoded the plan's letter ("replay = inert") and was
   wrong against the pinned panel contract. Caught by the server
   suite before any push — the safety net worked; my design-phase
   reading of the pinned tests did not (see d3).
3. **Session's own status trail.** The 22-40 report is written and
   pushed; THIS self-review exists because the owner asked — I had not
   planned one unprompted (the previous session's report set the
   pattern; I should have continued it).

## c) NOT STARTED

1. **Train D** — island `CALL_TRANSITIONS` (D1.1–D1.6): deliberately
   untouched, blocked on owner question g3 (sequencing; my lean: fresh
   session + stack browser E2E immediately after).
2. **Owner calls requiring answers before work:** A1.7 (verdict
   tightening — recommend keep), E2 (DOMAIN_LANGUAGE ratify-vs-delete —
   memo in the file), g1 formal blessing of B's re-adjudication
   (default already executed), g2's optional test hardening.
3. **Nothing else from the plan** — all other rows shipped or
   re-adjudicated this session.

## d) TOTALLY FUCKED UP (owned, with root causes)

1. **I shipped a broken build to main (`dc0a52f`).** Root cause: I
   wrote `internal/domain/thread.go` declaring `type Thread` WITHOUT
   verifying the type didn't already exist — my earlier discovery grep
   DID cover `internal/domain/*.go` but I piped it through `head -8`
   and the store-package matches consumed the budget before the
   domain match printed. The daemon committed AND PUSHED the broken
   intermediate; main sat red ~18 minutes until I pushed the fixed
   tip under manual-push bar clause 1. Self-inflicted, fully
   preventable: never truncate discovery greps; grep the exact symbol
   before declaring a type.
2. **Two CI-red pushes from the error-code registry.** C1 renamed
   `store.thread_upsert` → `thread_lookup`/`thread_update`; I gated on
   the packages I touched but not `internal/arch`, where the generated
   registry lives. Red: `dadb75e` (and `a7a9073` before it). Fix:
   `017ac3e`. Process fix applied from then on: gate `./...` before
   every push. I had even seen the registry test's name earlier in the
   session and failed to connect it.
3. **Designed C2 against the plan, not against the pins.** I read
   `webhooks.go` (closed verdict vocabulary) but NOT
   `webhooks_test.go`/`idempotency_test.go` before encoding replay
   semantics; my own new tests encoded my misreading and passed. The
   existing server suite caught it. Rule going forward: before
   encoding any state-machine semantics, read the pinned tests for
   that seam FIRST — they are the contract, the plan is a proposal.
4. **CI ran red in four windows this evening** (~30 min total):
   dc0a52f (d1), registry ×2 (d2), dup-gate (0c53551 — the gate
   correctly rejected twin `ChangeFunc` aliases I'd introduced; fixed
   `7d49ac1`). All healed same-session, but four reds in one evening
   is the worst CI discipline of any session this month.
5. **Small sloppiness, all caught in-flight:** a garbled const/alias
   block in `fax.go` (two naming attempts pasted together — cleaned
   before commit); `git add -A` once (riskier than scoped adds under a
   racing daemon + concurrent-session policy); piped test output
   through `tail` and lost failure detail (the exact anti-pattern the
   handoff warned about for logs); an unescaped grep pattern
   (`--- FAIL|...`) that ate its own flags. None reached origin.

## e) WHAT WE SHOULD IMPROVE (systems, not self-flagellation)

1. **Discovery greps must never be truncated** (`head` hid the
   existing `Thread` type → d1). Also grep the exact symbol before any
   new type declaration.
2. **Pre-push gate = `./...`**, always — package-scoped gating missed
   the arch registry twice. Consider a tiny wrapper that runs
   `go test ./...` + `TestErrorCodeRegistryIsFresh` explicitly before
   any `git push`.
3. **Read pinned tests before encoding semantics** (d3). The plans
   come from reviews; the tests are the contract.
4. **The daemon needs a compile/fast-check guard before push.** It
   pushed a build-breaking commit while I was mid-edit; the commit-msg
   hook guards dep files but nothing guards compilability. This is
   owner-surface tooling — flagged, not self-implemented (see g2
   below).
5. **Commit micro-milestones immediately.** My trains each spent
   10–20 min between "files written" and "commit" — longer than the
   daemon's sweep cadence; that's why narratives kept getting
   swallowed. Write one file group → gate → commit.
6. **The flake ledger was not updated.** I grounded the
   TestUploadBodyIsBounded mechanism but wrote it into the g2 addendum
   only; `docs/ci-flake-ledger.md` is the named one-home for CI-flake
   history and still lacks the line. (Deliberately NOT done now: the
   instruction was report-then-wait.)
7. **Stack-runbook cross-sync unverified.** The error-contract table
   changed rows; AGENTS says the stack runbook mirrors it. I did not
   check whether the runbook lists per-code rows or links the table
   wholesale (one grep next session settles it).
8. **Session reports should be unprompted.** The previous session set
   the pattern; I only wrote this one when asked.

## f) Next up to 50 (ordered: owner calls → D train → this session's
loose ends → harvested backlog)

1. Owner g3: decide Train D sequencing (lean: fresh session + stack
   browser E2E right after).
2. Owner A1.7: tighten out-of-order verdicts or keep today's overwrite
   (rows live in `deliveryMoves`).
3. Owner E2: ratify or delete `docs/DOMAIN_LANGUAGE.md` (memo inside).
4. Owner g1: formally bless Train B's delivered-by-prior-train closure
   (default was executed).
5. Owner g2 remainder: bless/refuse the one-line `:memory:` test
   hardening (shared-cache DSN or nil `ConnMaxIdleTime` for memory
   paths).
6. Add the flake-mechanism line to `docs/ci-flake-ledger.md` (e5.6).
7. Implement g2's test hardening if blessed (one line + a comment).
8. Train D1.1: inventory call-state moves from `calls.js` (ringing/
   established/ending, hold pending/queued).
9. Train D1.2: `CALL_TRANSITIONS` table + single `applyTransition`.
10. Train D1.3: fold the hold pending/queued machine into the table.
11. Train D1.4: node:test specs per move incl. honesty contract
    (failed toggle returns the SETTLED state; render-outcome
    assertions, not silent-throw stubs).
12. Train D1.5–D1.6: island-lint + island tests + `nix fmt` (prettier
    owns island files).
13. After D: stack browser E2E against the island (served-verbatim
    strings are greppable contract).
14. Check the stack runbook's error-table mirror for the renamed
    `store.thread_*` codes (e5.7).
15. Harvest row: collapse `UpdateOutboundStatus`'s raw-string failure
    params onto typed `FailureKind` (deferred last session).
16. TODO_LIST #37: voicemail transcribe-button test.
17. TODO_LIST #5: 429-combined-load behavior test.
18. TODO_LIST #43/#44: ASR retry/backoff + toast.
19. TODO_LIST docs cluster (the standing docs-debt row).
20. TODO_LIST branching-flow unskip re-triage (next owner window).
21. Monthly error-family re-measure (tier 1+2; next due 2026-11-05,
    must stay 0/0).
22. Quarterly art-dupl baseline/suppressed-set audit (due ~2027-01).
23. DOMAIN_LANGUAGE refresh pass IF ratified (align sections with
    machine names; AGENTS links it as language home).
24. If E2 = delete: remove the file, record the ruling in ROADMAP,
    drop the TODO/ROADMAP cross-links.
25. Daemon guard (owner-surface, see g2 below): compile-before-push or
    fast-flake-check gating for heuristic commits.
26. Consider a pre-push wrapper script (e5.2) so the registry miss
    class is mechanically impossible.
27. `internal/server` carve refinement row (live in TODO_LIST; the
    below-module refinement from the modularization verdict).
28. `TestDriveBranchesOnReceiptResolution`: extend with a
    provider-ref-preserving assert on the settled ROW (I asserted the
    returned job; the row's `provider_ref` CASE-preservation is pinned
    only indirectly today).
29. Store: pin the `store.thread_lookup` error path (driver failure
    mid-append → wrapped `Infrastructure` code) — new branch from C1
    has no direct test.
30. Fax: `Resend` drives the retry job but its `FaxAdvanced` event
    fires twice on the resend path (once at create, once in drive) —
    verify both pushes are harmless (identical list) or dedup at the
    projection; note in E1's policy test if intentional.
31. Notifier: `ThreadFlagChanged` has no payload-id field; document in
    events.go that this is the list-only mechanism (it IS documented —
    add a cross-link from sse.go's event list).
32. Check whether `MessageDeliveryAdvanced` from `Send`'s refused arm
    (row already failed) double-nudges vs the pre-E1 behavior (same
    call count — verify with the SSE test once more).
33. Island tests: add a projection-equivalence note? No — island has
    no SSE knowledge; skip (recorded to prevent re-litigating).
34. `docs/lessons.md`: add the truncated-grep lesson (d1) and the
    read-pinned-tests-first lesson (d3) as war stories with evidence.
35. AGENTS: consider one line under Conventions: "pre-push gate is
    `go test ./...`" (currently only my session reports carry it).
36. The 22-40 report §g daemon-guard signal: owner to route (buildflow
    repo? crush-config?) — needs an owner home decision.
37. Visual harness (T23) re-run after Train D lands (island surface
    changes).
38. Re-run the full buildflow at the START of the D session (train-
    start rule; also validates the daemon hasn't eaten
    `.buildflow.yml` again).
39. Re-read `.buildflow.yml` integrity at D-session start (the
    truncated-skip_steps incident class).
40. `git config core.hooksPath .githooks` sanity check on this clone
    (should resolve; verify once after the daemon activity today).
41. Consider `SetConnMaxIdleTime` documentation: if the owner refuses
    the test hardening, document the 5-minute/memory interaction at
    `store.Open` so the next investigator doesn't re-derive it.
42. Fax machine: add `FaxSettled`-from-`FaxSending`-only rationale to
    the D-session's island table notes (the same lane discipline will
    apply to CALL_TRANSITIONS).
43. If a second consumer of `internal/domain` ever appears, the
    single-go.mod verdict reopens (named trigger; ROADMAP holds it).
44. Sweep: verify no `//nolint` was needed anywhere in the new machine
    code (none was — keep it that way; branching-flow found nothing).
45. Status-report naming: this file doubles as the session's
    self-review; if the owner prefers them split, say so and I'll
    split future ones.

(46–50 deliberately unspent: the harvest rows above already enumerate
the rest — re-minting them here would break the one-home rule.)

## g) Questions I cannot figure out myself (3)

1. **Train D sequencing (g3).** Execute `CALL_TRANSITIONS` in a fresh
   session with a stack browser E2E immediately after (my lean —
   hottest file, verbatim-served, string-grepped, and this session
   already showed my error rate climbs late in long sessions), or
   continue per plan order whenever you next dispatch?
2. **May the daemon get a compile-before-push guard?** It pushed a
   build-breaking heuristic commit to main while I was mid-edit (d1).
   The fix lives in daemon/automation tooling I do not own. Do you
   want (a) a compile-check gate in the daemon, (b) daemon pushes
   gated on the fast flake checks, or (c) leave it — the
   check-CI-after-every-push discipline caught it in 18 minutes?
3. **g2's one-line test hardening — worth spending?** Grounded
   mechanism says the flake is unreachable in production. Options:
   shared-cache memory DSN, or nil `ConnMaxIdleTime` for `:memory:`
   paths in tests. It de-flakes local runs but touches
   `store.Open`'s documented pool settings (an `auth:`-adjacent
   surface). Bless, or leave documented-environmental?

*(A1.7 and E2 remain open owner calls too — they are fully framed in
their own artifacts and did not make the 3-question cut.)*

---

*End state at write time: local == remote == `51cca0d`, CI green,
tree clean, one probe log archived at `/tmp/upload-flake-probe.log`.
Waiting for instructions.*
