> ARCHIVED 2026-11-12 (docs-health v11 sweep): closeout complete — the near-miss fix + guard landed (M05 train closed)

# Session status — M05 storage-ladder closeout, the driver-registration near-miss, and the flake hunt (2026-11-12, 06:4x CET)

> Claims verified as of 06:45 CET · main `f3fd6cd7`, CI GREEN (run 43160641817) · this
> session resumed the SUPERB plan's ladder spine (M05) and drove it to done. One
> artifact of this report: everything below is what THIS session did, saw, or broke —
> not a corpus sweep.

## a) FULLY DONE

1. **M05 storage-ladder closeout — the ladder is DOCTRINE COMPLETE.** `internal/readmodels`
   is gates + view types ONLY: deleted `push.go`, `push_test.go`, `agentaudit.go`,
   `agentaudit_test.go`, `export.go`, `readmodels.go` (the Engine itself) and
   `readmodels_test.go` (−7 files); wrote `agentauditgate.go` (gate + view + cap,
   shapes copied verbatim from git history), `doc.go` (the doctrine-complete package
   header), and restored `pushgate.go` — which the live sibling session had TRASHED
   while it sat untracked (my fault for leaving it uncommitted, see §d).
2. **Two load-bearing side finds fixed at root:**
   - **The sqlite driver registration had a silent carrier.** `metaengine.LookupDriver`
     consults go-cqrs-lite's OWN init-time registry (NOT database/sql — my first
     theory was wrong and cost real time). `readmodels.go` was the only importer of
     `sqliteengine` in the whole graph; deleting it would have made EVERY
     `system.New` boot — prod included — die `unknown driver "sqlite"`. The ONE new
     home is a blank import in `internal/cqrs/threadsregister.go` (the boot file).
     Every boot now rides the cqrs graph transitively.
   - **The ring-prune read-your-write flake.** `TestHookRingPrunesDeadDevices`
     failed 1/5: the retired engine folded synchronously, the journal folds on the
     projection worker. Fixed with the established D7 `awaitOwnWrite` on
     `PushDevices.Upsert`/`Remove` AND `AgentAudit.Append` (both backends lacked the
     `vers streamVersions` field; now wired). 10/10 green after; the same hazard in
     the audit door closed preemptively.
3. **Consumer flips + purity pin.** Five test sites boot the journal backends now
   (`auth_audit_test`, `push_api_test`, `summary_api_test`, `hooks/ring_test` with its
   own `newRingCoreSystem`, `hooks/idempotency_contract_test` riding `sys.Dedup()`);
   `newTestEngine` deleted; `NewCoreSystem` grew `PushDevicesDomainDecl()` +
   `AgentAuditDomainDecl()`; new arch test `TestReadmodelsStaysGatesAndViewsOnly`
   (no sqliteengine import, no `func Open(` under readmodels — forever).
4. **Parity-comment sweep.** ~12 sites across 10 files reworded from
   engine-as-current-actor to the one-journal truth (contacts, snippets, pushdevices,
   agentaudit, enroll, dialprojections, callnotes, callreviews, transcriptsstore,
   app.go ×5 wiring comments, push.go package header). Historical-comparative
   comments deliberately kept.
5. **Full verification battery, all green:** `go test -count=1 ./...` (36 pkgs), race
   suite on the 8 touched packages, art-dupl 0 new (baseline holds 117), dprint +
   treefmt canonical, island-typecheck + webphone-module nix checks, and the live
   smoke (54 checks + all 7 scenarios) on a fresh binary — the boot-critical proof
   for the registration move.
6. **CI flake fixed at root (my tip's red).** Run 43159504844 red the race suite on
   `TestMetricsServesAggregatesOnly`: the duration SUM's decimal digits
   (`0.027210017000000003`) happen to contain the digit sequence "1001" — the
   aggregates-only pin greps the whole body for the extension literal. Fixed: the
   extension is asserted in its quoted label form (`"1001"`), the phone number and
   body text stay whole-string (not float-representable). CI green on the tip
   containing the fix (43160641817). Error-contract registry regen (−10
   `readmodels.push_*`/`agent_audit_*` codes) rode the code commit.
7. **Ledger + battery.** CHANGELOG `[Unreleased]` Removed entry (naming both side
   finds), TODO_LIST closeout strike row with evidence, AGENTS ONE-journal-world
   bullet rewritten to the doctrine-complete state, plan-doc §9 landing record
   naming every landing commit. docs-gates battery green end-to-end.
8. **Cadence drain.** 4 of 5 clean candidates archived gates-first (annotate
   check-rows --strict + bare-rows green BEFORE each `git mv`; banners + INDEX regen,
   62 reports): the 15-15 ui-restyle snapshot, the 04-10 fleet-libs refresh, MY 04-42
   ladder report, the 04-45 release closeout. 64 → 60 unarchived. The sibling's live
   05-22 report deliberately left for their own drain.
9. **The freshness/audit reds healed with the documented remedies.** Struck the
   2026-10-24 report's §f row 46 with evidence (`c85aefb` verified as a main
   ancestor — the pushed end state it asked for); re-derived `generic_df` 22 → 27
   per the standing 0.5%-of-living-lines rule (corpus grew to 5 432 lines; my new
   prose flipped one more historical row over the old bar — the exact DF-drift class
   AGENTS warns about), WITH the planted-row control re-run and passing.
10. **cqrs-lint re-run** (pinned v4.20.1 over `./internal/cqrs`): 5 findings, ALL
    previously-adjudicated classes (F032 catalog-scope, F031 the fetchSegments
    variadic forwarder, F009 the visibility await, F021×2 fold heuristics) — M05
    added ZERO new classes and one F019 healed. Delta recorded in the TODO row.
11. **Ownership calls recorded:** M09–M12 belong to the sibling's live UI lane (their
    capture-matrix/nav-overflow session is actively landing — I did not touch it);
    the uncommitted `.githooks/pre-commit` change from the morning self-resolved
    (clean tree by the time I checked); M14 stays owner-gated.

## b) PARTIALLY DONE

1. **The full buildflow gate** — my run finished 74/75 with ONE step failure, which I
   adjudicated as the concurrent-commit race (the daemon committed the sibling's
   17-file templ sweep mid-run; the historically-failing steps `island-typecheck` +
   `webphone-module` verified green directly afterward). I did NOT get a clean full
   76/76 buildflow run on the exact final tree — CI's race suite + every gate ran
   green separately, but the buildflow-shaped single verdict is owed.
2. **The dedup-registry sweep-log** — the registry's canon says append a sweep-log
   line per run; art-dupl ran clean (0 new) this session but the line is NOT written
   yet (last entry is the 21:0x repair from the prior session). Owed.
3. **docs/lessons.md** — the driver-registration war story is recorded in
   CHANGELOG/TODO/AGENTS but NOT yet in the war-story home it belongs in (the
   database/sql-vs-own-registry mechanism cost ~30 minutes to prove; that is exactly
   what lessons.md exists for). Owed.
4. **The freshness-drain discipline** — I fixed the two red gates after they failed;
   the skill's `--pre-flight` verb (judge candidate living-doc lines BEFORE writing)
   exists precisely to avoid this loop and I did not use it. The outcome is green,
   but the loop ran backward.

## c) NOT STARTED (all consciously out of this session's lane)

1. **M09–M12** (visual verify, hover capture, bubble bursts, empty states) — sibling's
   active lane; their session landed a capture-matrix report minutes before I
   finished.
2. **The server carve (sibling plan T04–T06)** — the new 1% item; actively owned
   elsewhere.
3. **M14 prod capture** — blocked on owner credentials (unchanged).
4. **M18 owner banner, M25 hub access, M26 trigger-gated, M27** (not due until
   2026-12-08/2027-01-11) — owner/external-gated tails, untouched by design.
5. **The 60 dirty cadence candidates** — each owes its own drain train; the 4 clean
   ones I drained were the batch this session could safely take.
6. **No dependency bumps, no vendorHash work** — zero go.mod deltas this session
   (the guard had nothing to guard).

## d) TOTALLY FUCKED UP (the honest list)

1. **I let my own finished work sit untracked and it got destroyed.** `pushgate.go`
   was written and verified in the PREVIOUS session, left uncommitted through the
   handoff, and the sibling's live session trashed it mid-morning. I restored it
   verbatim (I had the full text in context), but the loss was avoidable: the
   standing synthesis says stage+commit in ONE call with verification, and I
   violated it for the most valuable single file of the train. Untracked artifacts
   near a live sibling session are debris waiting to be swept.
2. **The driver registration nearly shipped broken — and I held the wrong theory
   for too long.** My first explanation (modernc's database/sql registration) was
   flatly wrong; go-cqrs-lite consults its own `metaengine.RegisterDriver` registry.
   I burned several investigation rounds on the wrong mechanism before reading the
   library source. Worse: I initially read the hooks/api test failures as
   package-local test plumbing and only later connected that the SAME deletion
   bricked the PROD boot. If the ladder recipe had been followed
   commit-quickly-without-the-deep-boot-look, main would have shipped a binary
   that cannot start. The smoke run (which caught nothing at first because my own
   env leak broke it, see #4) is the only reason the final state is boot-proven.
3. **All narrative commits were eaten by the daemon again.** The M05 code landed as
   three `chore: auto-commit` blobs and the docs ledger as two more; only the
   cadence drain, the landing record, and the flake fix kept narrative messages.
   I knew the daemon's timing, I had the one-call synthesis, and the multi-gate
   verify loops still out-ran it. The history is honest-but-ugly, and per the
   fix-forward protocol it stands — the cost is narrative, not correctness.
4. **I broke the smoke with my own environment, then almost mis-diagnosed it as a
   product bug.** Running the smoke with the session's `WEBPHONE_DATA_DIR` still
   exported leaked the data dir into the smoke's SECOND server (same dir as the
   live first server → boot fail → "configured server did not become ready").
   One false-alarm investigation round before I read the script's env inheritance.
   The AGENTS smoke line shows the env-free invocation; I should have used it.
5. **The first planted control row was invalid and I nearly accepted it.** I planted
   "Plant the zqxwavr control row for the kryptonym threshold check" — real words
   (`plant`, `control`, `threshold`) matched living docs, so the tool's verdict was
   "harvested" and the control proved nothing. I caught it only because I checked
   why the grep was empty. The valid control is pure gibberish; the ritual should
   be scripted so nobody plants a plausible-sounding row again.
6. **The purity arch test's first draft contained literal garbage** — a nonsense
   `runtime.Caller` lambda I wrote and replaced within the same edit sequence. It
   never landed, but it shipped-quality-tax on my own review.
7. **Edit-tool freshness races cost three wasted round trips** (two "modified since
   read" rejects on shared files + one edit attempted off a bash-sed read instead
   of a view). Mechanical, known, and still repeated.

## e) WHAT WE SHOULD IMPROVE

1. **Commit before the gates, not after.** The daemon sweeps within seconds; every
   multi-minute verification loop loses the race. The working pattern this session:
   verify → IMMEDIATELY stage+commit in one call → run the long gates → fix-forward
   if anything reds. Codify it as the default instead of the exception.
2. **State the mechanism, then read the source before theorizing.** The
   driver-registration hunt went hypothesis-first (database/sql) when one grep of
   `LookupDriver` would have ended it. For any library-magic question: read the
   lookup function FIRST.
3. **Script the generic_df ritual.** Re-derivation is `floor(living_lines × 0.005)`
   plus a gibberish-row control — both mechanical, both currently hand-run, and one
   hand-run was wrong. One script (`scripts/generic-df-control.sh`?) removes both
   failure modes.
4. **Use `--pre-flight` before writing living-doc prose near the report corpus.**
   The freshness/audit loop ran backward (red → diagnose → fix) when the tool
   exists to run it forward.
5. **The arch purity test's `func Open(` heuristic is comment-brittle.** It does a
   raw `strings.Contains` over file bytes — a future DOC COMMENT in readmodels that
   mentions `func Open(` (as my own AGENTS prose does) would red the test with a
   misleading message. Parse or match declarations properly, or accept + document
   the footgun at the test site.
6. **Give new ledger rows rare tokens** (the §f corollary): my closeout TODO row
   uses common words and WILL DF-drift as the corpus grows — exactly the failure
   the canon warns about. Retro-fit an anchor phrase.
7. **Smoke invocations must scrub the session env** (`env -u WEBPHONE_*`), or the
   smoke script itself should unset inherited WEBPHONE_* config vars before booting
   its servers — the leak class will bite the next session the same way.

## f) Up to 50 things we should get done next (session-derived; the standing queues in TODO_LIST and the sibling's 05-25 plan own the rest)

1. Append the dedup-registry sweep-log line for this train (art-dupl 0-new, baseline
   117) — the registry canon requires it per run.
2. Write the driver-registration war story into docs/lessons.md (database/sql vs
   go-cqrs-lite's own registry; the silent transitive carrier class; the smoke as
   the boot proof).
3. Add a named boot-canary test: a system.New with Driver "sqlite" and an explicit
   failure message pointing at the cqrs blank import (today only incidental test
   boots catch its removal).
4. Get one clean full buildflow run (76/76) on the settled tree — the session's run
   carried the concurrent-commit transient.
5. Script the generic_df re-derivation + gibberish control
   (`scripts/generic-df-control.sh`).
6. Fix the purity test's comment-brittle `func Open(` heuristic (match declarations,
   not bytes).
7. Retro-fit a rare-token anchor into the M05 closeout TODO row (§f corollary).
8. Make the smoke script unset inherited `WEBPHONE_*` env before booting its
   servers (the leak class from §d-4).
9. Verify-then-strike the 2026-10-24 report's §f row 1 ("Register the 3 error
   codes") — the codes registered long since; the row predates the strike canon.
10. Verify-then-strike or route the remaining ~52 under-ratchet UNHARVESTED rows the
    audit lists (each needs its own evidence pass — batchable into one drain train).
11. M09–M12 (sibling lane): visual verify pass, hover capture, bubble bursts,
    empty-state upgrade — verify their landing and close the plan rows after.
12. Server carve steps 1–3 (sibling plan T04–T06): when landed, run the export
    census and re-check the AGENTS architecture invariants still hold verbatim.
13. Depth-budget arch test (sibling T07) — verify the readmodels purity test and it
    don't double-pin the same shape.
14. M14 prod capture — owner creds still owed.
15. When the sibling's 05-22 deepening report goes stale, archive it gates-first
    (it was deliberately left live this session).
16. Update the fleet-libs `go-cqrs-lite.md` report with the driver-registry
    mechanism finding (the report's "utilization" frame should know the
    registration is consumer-owned).
17. Upstream draft candidate (verify-then-file): go-cqrs-lite's `LookupDriver`
    unknown-driver error says "did you import the driver package?" — naming
    `metaengine/sqliteengine` for the sqlite driver would have saved the 30-minute
    hunt. go-error-family-family error, low effort, high next-session value.
18. Deterministic pin for the metrics-leak fix: a spec asserting the quoted-label
    check specifically (feed a crafted body whose float contains the extension
    digits and expect pass; quoted form expect fail).
19. Boot-failure smoke arm: a driver-unregistered boot (env-simulated) asserting the
    boot report names the cqrs step — turns §d-2's near-miss into a pinned
    scenario.
20. erraudit `nolint-audit` decay pass (standing watch; the D7 await additions
    touched error paths across two backends).
21. The `newTestReadmodels` helper now wraps only Contacts — either generalize or
    rename so the name stops promising more than it boots.
22. Family-test pins: an explicit await-visibility pin for PushDevices/AgentAudit
    (the ring-prune timing test covers it indirectly only).
23. `docs-gates freshness . --margins` report — which living-doc tokens sit just
    under the DF bar before the next prose growth flips them (proactive, one run).
24. Harvest this report's own §f top rows into TODO_LIST (the canon: new report →
    forward the actionable ones).
25. Re-check the CI red-main watcher's tracker issues from the 04:55 and
    landing-record reds — confirm they carry RESOLVED shas or got suppressed by the
    open-ci-red dedupe.
26. The 04-55 run's format rc=20 (the sibling's TODO plan-table spacing) — verify
    their landing healed it everywhere (the 06:19 run is green, so verify the
    specific table survives dprint on THEIR remaining rows).
27. Retire the modernc blank-import comment drift risk: `internal/app/cqrs.go:23`
    still says "for the DLQ handle" — after this train the same import also
    matters for nothing else (cqrs owns the engine registration), but the comment
    should not imply it registers the ENGINE driver (it does not — metaengine's
    registry is separate; worth one clarifying word).
28. Sweep remaining `engine`-wording comments in the OTHER packages the sweep didn't
    cover (fax/, messaging/, store/ headers referencing the retired engine store
    era) — read-and-judge pass, same shape as this session's.
29. The backup-set note: confirm readmodels.db orphan language in the backup unit
    docs matches the doctrine-complete state (AGENTS says it; the NixOS module
    docs may still hedge).
30. CHANGELOG-coverage CI job shows "skipped" on green runs — confirm that is the
    designed gate shape (it looked odd twice this session).
31. Teach the wiring-parity test (`TestAppWiresEveryServerDep`) about the two
    unaffected Deps seams this train touched? — verified unchanged; only do it if
    the next Deps field lands.
32. Strip the double registration risk: confirm NO second sqliteengine blank import
    crept into any test file during the flips (grep says clean; re-verify at next
    train start).
33. TODO_LIST: strike the storage-ladder PARENT row if one still exists above the
    per-train rows (the per-train strikes are in; the umbrella row may linger).
34. ROADMAP: the event-sourcing doctrine paragraph — add the "readmodels is
    contracts-only, purity-pinned" sentence if the paragraph predates today.
35. The vendored sip.js/update.sh path — untouched, verify the check still passes
    at next release drill (no action this session).
36. `scripts/status-index.py --check` in CI? — the index regen ran locally twice;
    a CI lane would catch a hand-edited index (cheap, prevents drift).
37. Re-run the audit-all after the next big living-doc growth with margins first —
    codify "grow prose → run margins" in the docs-health section of AGENTS.
38. The `pushDevicesKeyInput`/`agentAuditInput` names still say "engine"-era input
    shapes — pure nit, only on touch.
39. Verify `internal/store/threadsdual_test.go`'s direct sqliteengine import is
    still intentional (it predates the ladder; it imports the engine for a legacy
    dual-write pin — confirm the pin is still load-bearing).
40. The AGENTS "ONE journal world" bullet cites the purity test by name — add the
    same pointer to the ARCHITECTURE bullet list entry for arch tests if one exists.
41. Consider exporting the cqrs boot DSN shape as a helper so hooks/ring + the
    contract test stop hand-building `"file:"+filepath.Join(...)` strings (three
    copies).
42. Upstream draft candidate #2: metaengine `register.go` documents self-registering
    engines in a comment that names sqlite — stale now that sqlite lives in the
    submodule; one-line upstream docs fix.
43. Check the sibling's 05-55 §f rows for anything actually mine (nav-overflow,
    capture matrix) before their next drain misattributes ownership.
44. Run `docs-gates freshness . --pre-flight` BEFORE the next AGENTS/TODO prose
    growth (make it a reflex, list it here so it is one).
45. The `engineDedupWindow` name in internal/app now wraps sys.Dedup() — rename
    candidate (`systemDedupWindow`?) to stop lying about the engine.
46. Sweep the TODO_LIST "reopen registry" for the EnforceCap nested-grouping trigger
    — confirm M05's caps (push 8, audit 500) stay loop-side deliberately and the
    reopen trigger text still names the right gap.
47. Add the engine-registration fact to the onboarding README's architecture
    section (the README sells the binary; one sentence prevents the next
    contributor from "cleaning up" the blank import).
48. Next train start: full buildflow BEFORE code (the standing rule — this session
    entered with the resumption and skipped it; CI covered the tail instead).
49. Label-form audit: grep the metrics surface for OTHER bare-number assertions of
    the same phantom class (the leak pin was the only one found, but the pattern
    deserves one grep).
50. Schedule the next tier re-measure (erraudit tiers 1+2 monthly; last recorded
    sweep row is due by 2026-12-05 per the standing cadence).

## g) Questions for the owner (not self-answerable)

1. **Commit-race policy:** the daemon's seconds-scale sweeps keep eating narrative
   commits (three more blobs this session). Should trains switch to
   "commit-first-verify-after" as the DEFAULT working shape (blob history accepted
   as the norm), or is a daemon quiet-window (a pause file the daemon respects
   during an active session's train) worth building? The first keeps history ugly
   but zero-infrastructure; the second costs a daemon change but keeps the
   narrative commits you have repeatedly valued.
2. **The buildflow transient bar:** my run red on one step during the concurrent
   templ-sweep commit and I adjudicated it by verifying the failing class directly
   instead of re-running the full gate. Is a clean full-buildflow verdict on the
   final tree REQUIRED per train (costing ~10 quiet minutes the daemon keeps
   interrupting), or is "every gate green separately + CI green" an acceptable
   equivalent when the failure provably raced a foreign commit?
3. **Driver-registration home:** the blank import now lives in `internal/cqrs` (the
   boot-owning package). The alternative is `internal/app` (the composition root).
   I chose cqrs because every test-boot site imports it and the boot helper lives
   there — but a prod-root purist would argue the composition root should own all
   wiring facts. Ratify the cqrs home (it is committed), or move it?

_Pure projection of one session's work; the standing queues (TODO_LIST, the 05-25
plan, the fleet-libs reopen registry) were NOT re-swept and remain the wider
backlog._
