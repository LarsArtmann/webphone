# Status: SUPERB declarative trains B/C/E executed — D pending owner g3 (2026-10-14 22:40)

Continuation of `2026-10-14_21-15_declarative-ddd-train-a-docs-hygiene-status.md`
(same dispatch: read, understand, research, reflect; execute and verify
step by step). Owner questions g1/g2/g3 stayed minted throughout; the
stated defaults were executed, never the calls themselves.

## a) Delivered this session

- **Train B closed as delivered-by-prior-train (g1's default).** B1
  (Resolution enum, no type-asserts), B2/B3 (config validations) were
  verified in code WITH test coverage; the three genuinely missing
  obligations were backfilled: B1.5 `TestDriveBranchesOnReceiptResolution`
  (a non-loopback immediate gateway settles the job — the exact
  wrong-branch the old type-assert risked), B1.6 the 09-30 review P5
  resolved-annotation, B2.3 README validation-contract rows. Commit
  `e0c5564` (amend-recovered from a daemon heuristic commit).
- **Train C1:** `domain.AppendToThread` — the append invariants
  (unread/lastActivity/auto-unarchive, unconditional stamp overwrite)
  moved from the store's SQL CASE into one pure function; the append
  transaction is now read-modify-write inside the same tx. Parity
  pinned by the existing flags/scoping suites + new domain tables +
  a store-level creation-fallback test. Commit `a7a9073`.
- **Train C2:** `faxMoves`/`FaxAdvance` machine; `drive` and
  `UpdateProviderStatus` rewired. Commit `dadb75e`.
- **Train E1:** typed domain events (`MessageReceived`,
  `MessageDeliveryAdvanced`, `ThreadFlagChanged`, `FaxAdvanced`) on a
  single `domain.ChangeFunc` seam; `Notifier.Project` is the
  notification-policy table (message events = list+transcript, flag
  change = list-only BY DESIGN, fax = fax list). SSE payload bytes
  unchanged (existing goldens green); new policy test pins the table.
  Commits `0c53551` + `7d49ac1` (daemon-split, content complete).
- **Train E2/E3:** DOMAIN_LANGUAGE ratification memo (drafted,
  owner-gated, deliberately unpicked) + ROADMAP cross-link; CHANGELOG
  lines for C1/C2/E1; AGENTS machine-tables invariant (149/377 lines);
  plan completion annotation; TODO_LIST row refresh; dedup sweep-log
  line. Commit `7d49ac1`, CI `38633368094` SUCCESS.

## b) The one substantive re-adjudication (C2.5)

The plan said "replay verdict = no legal move = inert". The pinned
panel contract disagreed: `TestFaxStatusWebhookParsesFlexiblePageCounts`
and `TestWebhookStatusIdempotent` pin that a same-status verdict
REFRESHES the row (a later hook carries the real page count), and
replay dedup is the hooks seam's ref-keyed `hooksIdem` (it refutes
even contradictory replays). Resolution: same-status verdicts are
legal refreshing moves in `faxMoves` (mirroring `deliveryMoves`'
duplicate-verdict doctrine); the machine's inert class is events
against an inbound `received` fact (no lifecycle). Documented at the
table, pinned at both layers (`TestVerdictSemantics`,
`TestFaxAdvanceInertMoves`). The plan's completion annotation records
the deviation.

## c) g2's mechanism — grounded and attached

The flake probe caught it: `WARN session: lookup failed error="SQL
logic error: no such table: sessions (1)"`. Chain: `store.Open` sets
`ConnMaxIdleTime(5m)` + `MaxOpenConns(1)`; the oversize subtest's 25
MiB crawl holds the request 9m1s with the DB conn idle; the idle
timer closes the conn and a `:memory:` database dies with it; the
next lookup opens a fresh EMPTY memory DB. **Harness-only** —
production paths are file-backed. Full write-up attached to g2 in the
21-15 report (item 24 honored). The un-minted remainder: whether to
spend a one-line test-side hardening (shared-cache DSN or nil idle
time for `:memory:` paths).

## d) Incidents + heals (all CI green at end state)

1. The daemon pushed `dc0a52f`, a broken intermediate (duplicate
   `Thread` declaration, build-fail), while I was mid-C1. Caught by
   the check-CI-after-every-push discipline; the green tip was pushed
   under manual-push bar clause 1 (red main + verified fix). CI healed.
2. Two pushes failed CI on the error-code registry (C1 renamed
   `store.thread_upsert` → `thread_lookup`/`thread_update`); the arch
   package was outside my first gate selection — the lesson (gate
   `./...` before push) is now applied; registry regenerated (`017ac3e`).
3. The new-clones dup-gate correctly rejected twin `ChangeFunc` alias
   blocks from E1; both aliases deleted in favor of direct
   `domain.ChangeFunc` references (`7d49ac1`). The gate did its job.
4. The daemon split narrative commits ~6 more times; amend/squash
   recovery worked only while unpushed (used 3×). Two pushed history
   seams remain heuristic-named (`a87b215`/`271a627` carry E1's core;
   content verified complete on origin).

## e) Gates at end state

- Full go suite `./...` (minus the documented environmental flake):
  22/22 packages ok. The flake itself reproduced twice more locally,
  never in CI — consistent with the grounded harness-only mechanism.
- Fresh-binary smoke: 47/0, restart 4/0, boot-failure 8/0.
- `nix fmt` clean; whitespace-drift clean at every commit; art-dupl
  step green post-fix; owed full buildflow (train A+B tree): every
  gating step ✔, exit 1 attributed solely to the documented 7/7
  `nix-build-verify` box-degradation casualty.
- CI `38633368094` SUCCESS on the final tree; local == remote ==
  `7d49ac1`.

## f) What remains (post-train backlog)

- **Train D** (island `CALL_TRANSITIONS`) — owner question g3
  (sequencing; lean: fresh session + stack browser E2E). The plan's
  D-rows are untouched.
- **Owner calls outstanding:** A1.7 (verdict tightening), E2
  (DOMAIN_LANGUAGE ratify-vs-delete — memo in the file), g1 formal
  blessing of B's re-adjudication (default executed), g2's optional
  test-side hardening, g3.
- Served-markup changes: none this session (no templ/island edits
  beyond train A's already-E2E'd `.IsRetryable()` swap) — no fresh
  stack browser E2E owed from B/C/E.

## g) Signals for the owner

- The dup-gate and the registry gate both caught real defects in my
  own work within minutes of pushing; both were fixed same-session.
  The gate lattice is earning its keep.
- The daemon pushed a red build to main mid-train (incident 1). The
  commit-msg hook guards dep files but nothing gates the daemon's
  heuristic commits against build breakage. A daemon-side
  "compile-check before push" (or gating pushes on the fast flake
  checks) would remove the only main-went-red vector left. Not
  self-implemented — daemon tooling is owner surface.
