# TODO_LIST

Short- and mid-term improvement tasks. Done work is DELETED, never struck
through (docs-health: shipped work moves to CHANGELOG / FEATURES — one home
per fact). Long-shot ideas and owner calls live in ROADMAP.md; every row
cites its source report in Evidence.

Last sweep: 2026-11-06 (docs-health v8, assistant): harvested the GitHub-issue
train's residue (§ below), re-verified every living doc against HEAD `54c8bb17`,
archived the August/September cohort's resolved snapshots, and adopted the
`docs-gates` battery (`scripts/docs-gates.sh` + `.docs-gates.json`) as the
standing docs health gate. Prior sweeps live in CHANGELOG + docs/status — this
front matter does not carry their history (one home per fact, per the rule above).

## Open tasks

### BuildFlow tool-disposition residue (harvested 2026-11-08 from the 2026-10-11 findings-zeroing plan — one home now)

Six tool dispositions the 2026-10-11 plan owned but TODO_LIST never carried;
each verified still-open at harvest time (the plan doc keeps its detail):

- [ ] go-auto-upgrade: add the `warnings_budget` marker + rationale comment to
      `.buildflow.yml` (lo/jsonv1 declined per doctrine; plan T6 µ20).
- [ ] jscpd: attribute the 6 island-test clone rows in docs/dedup-registry.md
      (test-fixture class; plan T6 µ22).
- [ ] lychee: rerun `-v` and reconcile the 5-vs-0 delta; the standing
      authenticate-vs-exclude policy waits on sitting pack §F9.
- [ ] nix: add `meta.description` to the vulnix app (nix/apps.nix; plan T6 µ23).
- [ ] File the drafted BuildFlow bandit-provider issue
      (docs/drafts/2026-10-11_buildflow-bandit-stdout-findings.md; owner banner
      still unchecked).
- [x] **DONE 2026-11-08 (P4-7): file-size-check configurable cap** — ALREADY
      SHIPPED upstream: buildflow #91 (closed COMPLETED 2026-11-06) fixed the
      inert `max_file_size` override at the root (the CLI clobbered the
      materialized value; now flag-Changed-gated) with a regression test;
      #66/#75 were the same defect family. No filing owed. The webphone-side
      posture call (raise the cap vs keep the 46 per-finding accepts) stays
      with sitting pack §D10 — untouched by this closure.

### 2026-11-09 cross-repo session residue (harvested from the 00-49 report)

New open items from the legs session, not homed elsewhere yet.

- [x] **Re-record the art-dupl baseline for the Phase-6 fax-journal
      groups** (adjudicate-before-write, pinned v0.7.2 detector — the
      same repair shape as the 2026-11-08 threads-aggregate gate fix);
      run 42538856915 red on it after the A39–A40 landing. Then read
      run 42539029787's failing legs (the second red was unnamed at
      snapshot time). Their train may own both — the 03-38 report's
      §g-3 asks the owner.
      DONE 2026-11-09 (SUPERB closeout train): the baseline re-record
      rides that commit (131 → 137, all six groups adjudicated in
      docs/dedup-registry.md — the readmodels fax twins, the
      ProjectionHostOptions registration idiom, the sweep trivia);
      run 42539029787's unnamed second leg was TWO: dprint format rc=20
      (the honest report's wrap points, reworded) and strike-safety on
      a 6-cell FEATURES row (healed by 4788bc52). The owner's execute
      mandate gave the train to the SUPERB session.
- [x] **Finish the v2.12.0 release train (legs 6–9)** — the train died
      at 6/9 on release.sh's mis-quoted lychee `--exclude` (a
      double-backslash inside single quotes reaches lychee as a
      literal backslash and matches nothing); fix the quoting and
      verify by running exactly that script line, then the lychee link
      check → stack relock to v2.12.0 (plus its CHANGELOG lock-move
      record, browser E2E, telephony-VM check, flake check, push with
      exit codes captured — no pipe-swallowing) → aarch64 cross-build
      + ELF b700 guard → `gh release create v2.12.0`. Every push owes
      the `gh run watch` to green.
      DONE 2026-11-09 (closeout train, report 04-50): quoting fixed +
      fixture-verified; the relock surfaced the ENROLL CLI SPLIT-BRAIN
      (fixed at efb0ff1a, E2E-proven 200/503) and two stale stack VM
      tests (fax-feed + messaging, modernized to the journal world);
      tag re-pointed onto the green fix commit; aarch64 b700 PASS; gh
      release published 04:20Z as Latest.
- [ ] **Delete the prod-dead readmodels faxes store** — the
      engine-backed `Faxes` in internal/readmodels/faxes.go lost its
      prod wiring at P7 (app.go wires only `cqrs.NewFaxes`); it
      survives as the fax service's test double
      (internal/fax/teststore_test.go, archiver_test.go). Move those
      tests to the cqrs backend, keep `FaxesProjection`/
      `FaxJobOfView`/`FaxGate` (faxesgate.go, the shared fold builder
      both DBs ride), then delete the store — dissolves 3 of the 6
      clone groups baselined 2026-11-09 (the registry's post-P7 fax
      twins row).
- [x] **fleet-libs `go-cqrs-lite.md` re-score (B159) + ROADMAP ES
      annotation (B160)** — the A46 leftovers: SUPERB made webphone the
      library's deepest consumer (threads/faxes aggregates, system
      composition, snapshots, DedupStore); the 83/100 score predates
      the whole train. ROADMAP gains the event-sourcing doctrine
      paragraph (ONE journal world; sessions stay SQL per D2).
      DONE 2026-11-09 (closeout train): re-scored 96/100 on a 14-row
      frame (11 direct modules; EnforceCap NO-GO superseded by
      WithCapPerGroup; audit-verify 8/8), the ROADMAP carries the
      doctrine paragraph, index.html + AGENTS updated (mean 80).
- [ ] **cqrs-lint advisory triage over internal/cqrs** — 185
      warnings (C044 ×50, C006 ×20, D014 ×17, C023 ×14, C009 ×16 …;
      C007 time-in-decider is only ×3 — the closeout report's "180
      C007" miscount corrected), ALL severity-warning and non-gating
      under the SDK cap. Adjudicate the real ones (A032 branded-id
      suggestions collide with the JSON-wire raw-string ruling), file
      the noise classes upstream, or baseline with rationale once the
      error_rules-bearing BuildFlow release lands.
- [ ] **Clone the transport-isolation fix into pbx, asr, matrix,
      paperless** — the crm client's Clone() repair (2026-11-08) killed
      the parallel-test flake class at ONE seam; the same bare
      http.Client-literal rides the process-global transport at the
      other four (pbx/client.go two sites, asr.go + google.go,
      matrix/client.go). Production-neutral; the class bites only
      under parallel httptest suites.
- [ ] **Hub runner durable fix** — ExecStartPre provisioning (or a
      static uid / group-write) for the build-state roots, so the next
      DynamicUser renumber cannot red every hub build (the 2026-11-08
      chown was in-place; the durable half lives in the hub's TODO).
- [ ] **Hub ALERT_WEBHOOK_URL decision** — the failure log says
      channel decision pending (ntfy / email / Forgejo webhook); pick
      the channel and set the secret so hub failures alert.
- [ ] **Verify catalog.home.lan serves run #240's publish** (the
      hourly sync half) and the site freshness WARN cleared.
- [ ] **Ringhook operator docs + typed mirror**: the stack's deploy
      docs gain the enable recipe (config keys, secret slot, expected
      dialplan arms); consider a typed ringhook front in webphone's own
      NixOS module for non-stack hosts; a ringhook health leg joining
      telephony-health; the operator command sheet recipe.
- [ ] **Confirm crm's own gates on the vendorHash repair** (`37a1957`):
      the hub build proves the build; crm's treefmt/buildflow verdicts
      were not read this session.

### 2026-11-08 SUPERB Phase-5 mid-flight residue (harvested from the 22:46 report)

All three rows CLOSED by the 2026-11-09 execution session (A38-A47
complete; the plan doc's Phase-5/6/7 as-built errata carry the
evidence):

- [x] **A38 phase-gate: stack browser E2E** — ran green against the
      v2.10.0 pin in the stack leg (319 s, under the 445 s budget);
      telephony-webphone VM check green in the same train.
- [x] **A43 dedup unification: hooksIdem/CallsIdem → sys.Dedup()** —
      done: the dedup window resolves from the system (ONE CAS home in
      system.db), the readmodels engine left the composition root, the
      backup set, and the drill; readmodels.db is an inert orphan.
- [x] **The metrics-count decision** — executed as the SUM posture
      during the import window (`legacy + gate`, marker-gated by
      `threads_journal_imported`; the importer flips the legacy
      contribution to zero — no double count, no undercount), then the
      journal is the one count.

### 2026-11-07 Frontend UX fix-train (one home: the SUPERB plan)

The 2026-11-07 design review
(docs/review/2026-11-07_01-51_frontend-design-ux-review.html) produced 20
carded findings + a 47-item agent matrix + 8 flow GAPs. The task inventory
lives in the plan, NOT here:
docs/planning/2026-11-07_04-09_SUPERB-frontend-ux-fix-train.md (Pareto
tiers, 24 comprehensive tasks, 119 micro-tasks, execution graph, gate
contracts). Harvest rows here only as trains close them. Owner calls open:
F1 ink direction, F19 signed-out shape.

Residue harvested 2026-11-07 (the execution train's open items):

- [x] M20 (report §f item 37) DONE 2026-11-07 (commit 5c368085): voicemail
      reload failure got its own copy (`vmLoadFailed` — 401 keeps the auth
      line), and the history/contacts loads announce a warn toast on both
      the `!res.ok` and thrown-fetch arms (`historyLoadFailed`/
      `contactsLoadFailed`, en+de). Pinned by
      panels-load-failures.test.mjs (6 specs, suite 318/318, tsc green).
- [x] Gate evidence closed 2026-11-07: train-start buildflow 74/74 (run
      20261107-032044), train-end 74/74 (run 20261107-084827), `nix flake
  check` all-checks-passed; run ids filed in the status report.

Residue harvested 2026-11-08 (the §f rows audit-all flagged; each verified
against the tree at harvest time):

- [ ] Drafts-poll skip-if-dirty guard (F11): no dirty check exists in
      shell.js's drafts poll — verify-skip when a draft textarea is
      mid-edit (grep-verified absent).
- [ ] Localize island-built aria-labels (F9): 8 `aria-label` literals in
      island/app JS render English in the de locale (grep-verified count).
- [ ] Chip-X 24px hit area (F14): `.wp-chip-x` is 18px (app.css:998-997),
      WCAG 2.5.8 asks 24 — grow the hit area without growing the glyph.
- [ ] History empty state rewording (F15): the filter-aware copy predates
      the review; reword per F15 if the reviewer's phrasing still reads
      better.
- [ ] Radius/spacing token adoption (F17): tokens exist
      (`--radius-sm/--radius/--radius-lg` app.css:38-40) but hardcoded
      radii remain across panels — sweep the stragglers.
- [ ] Phase-4 release preflight (harvested 2026-11-08 from the 17-22
      SUPERB report §f-2 residue): the KVM/local-only flake legs —
      `nix build .#checks.x86_64-linux.webphone-module`, the
      `webphone-backup` + `webphone-backup-drill` VM tests — before the
      SUPERB release train; CI cannot exercise them.

### 2026-11-07 Datastar deep-dive residue (one home: the report)

The full client-side adoption question is CLOSED (NO-GO, posture 78/100,
docs/research/2026-11-07_datastar-deep-dive.html — supersedes the 2026-11-04
go-datastar adjacent ruling's trigger paragraph). One actionable item fell out:

- [ ] Retire `/health` `'unsafe-eval'`: go-health-dashboard upstream adopts the
      embedded Datastar v1.0.3 nonce + Trusted Types lane (emit `data-nonce` on its
      page, grant `'nonce-…'` instead of `'unsafe-eval'` in its csp.go), then
      re-pin `TestAppDashboardMountsHealthSubtree` (app_test.go:135) to assert the
      nonce form. Evidence: report F4.
- [ ] **Implement the dashboard lane in the library repo** (go-health-dashboard):
      the nonce emit + CSP grant above, then run ITS gates there — this repo cannot
      ship the change and must not fake it locally.
- [ ] **Release the dashboard version** (tag) once the lane lands — owner call per
      the report's §g1; only then can webphone bump and re-pin the CSP assertion.

### 2026-11-06 GitHub-issue train residue (harvest from the 10:45 report)

The 2026-11-06 train closed 4 of 5 open GitHub issues (Reports tab #12, live-call
context + ring routing #10, two stale red-CI flags) and shipped `crm.ca_file`.
What remains, each with its ONE home:

1. **Push the 14 local commits** and watch CI. Manual-push bar MET (first unpushed
   commit `216ed28e`, 2026-11-06 09:30; ~75 min at 10:45). Nothing else here is
   cheaper until remote `main` carries the train.
2. **Merge PR #234** (ginkgo 2.33.1→2.34.0, green on `a2ad2815`, vendored-module
   hash already re-pinned). Owner call; the tidy → re-pin → watch tail is done.
3. **Stack `crm.caFile` pass-through** — `nix-international-telephony`'s typed CRM
   front renders only `crm.url` + the env-file token, so the option cannot reach the
   PBX host. Cross-repo; the standing home is the Cross-repo obligations section.
4. **Issue #9 acceptance criteria 1–2** — needs `/run/secrets/crm_api_token`
   (root/unit-scoped by design) and the PBX host: an inbound call must render the
   Kith name and `call_logged` must land on the journal. The webphone slice is
   complete; the residue is pbx-artmann-side.
5. [x] **Anchor `.gitignore` against package shadows — DONE 2026-11-07 (P1-2)**:
       `/result`, `/result-*`, `/.devenv`, `/.direnv` are root-anchored, and
       `internal/arch/gitignore_test.go` now walks `go list ./...` and fails when any
       package directory resolves to an ignore rule. `--no-index` is load-bearing:
       without it the index silently excuses already-tracked packages, which made the
       planted-rule canary pass until the flag was added. The rules inside the
       buildflow-managed block (`build/`, `dist/`, `vendor/` …) cannot be edited —
       `buildflow --fix` re-generates them — so THIS test is their guard.
6. **Decide the `crm` health leg's criticality** — a CRM outage reds `/healthz`
   readiness although the phone keeps working, while the library treats crm as
   non-critical. Nothing routes traffic on it today (unit `Type=simple`, Caddy
   fences the probes), so it stays an alarm channel until made explicit.
7. **Findings-gate posture** — `buildflow` cannot exit 0 while 50 pre-existing
   tier-3 findings stand, which trains readers to ignore its exit code (the
   retired-baseline failure mode one layer up). Decide: elevate and fix, or a
   sanctioned, visible baseline.
8. **Re-run the hub fan-out benchmark on a quiet host** — the 2026-11-06 attempt was
   inconclusive (load ~22 on 32 cores; spreads ~2×). Its trigger (cqrs-htmx v4.16.0 +
   go-sse v0.7.0) has been fired for two trains.
9. [x] **Store test train — DONE 2026-11-07 (P1-1)**: `internal/store` reads
       **82.4%** (was 64.8%, and 75.3% on 2026-09-20). The gain came from the block
       nobody had pinned at all — the transitional reads of the retired SQL tables
       plus the `app_flags` helpers, all four readers and four translators at 0.0%
       (`internal/store/legacy_import_test.go`). The admin read seams were already
       pinned both ways by `store/admin_reads_test.go` and are now measured
       (77–86%; the rest is the infrastructure-error branch), and one assertion is
       mutation-proved. Delta recorded in
       `docs/reviews/2026-09-20_coverage-baseline.md` § 2026-11-07.
10. **Wire the `docs-gates` battery into CI** — adopted locally this sweep
    (`scripts/docs-gates.sh`, `.docs-gates.json`); the pre-canon archive residue is
    RATCHETED, so nothing reds main when new drift appears until the CI leg lands.
11. **Bulk verdict pass over the archive residue** — after the v8 sweep archived 68
    reports, the live status-report set is blocked by un-inventoried §a/§b table rows
    (~1 600 measured 2026-11-06; those reports use a Verdict COLUMN instead of the
    strike canon). PROGRESS 2026-11-07 (v9 sweep): the four executed September plans
    (error-excellence 95, 20-year-durability 217, pareto-v260 118, stack-adoption 123
    rows) were per-row verified, struck with evidence and ARCHIVED (gates held before
    each move; citation repoints incl. archived citers). Remaining: the live status
    reports (~1 600 rows). Plan: per-file `repair` strikes for §a/§b shipped tables
    (Verdict-column evidence transcribed + hash-resolved), routed strikes for §f rows,
    then `check-rows`/`bare-rows` BEFORE each `git mv`. The cadence ratchet holds the
    residency and may only move DOWN.
12. **`/context` contract snapshot test** pinning webphone's parse against Kith's
    published shape (a CRM-side schema change is invisible today).
13. [x] **History nudge for the notes save path — DONE 2026-11-07 (P2-3)**: a
        successful `call_notes` PUT publishes the payload-less `notes` SSE event
        (`NotifyCallNotesChanged` on the api Server seam, same shape as the contacts
        nudge, fired only on success), the History panel subscribes with
        `hx-trigger="sse:review, sse:notes"`, and the generated wire contract now
        declares BOTH `review` (F89 shipped it undeclared) and `notes`. Pinned by
        `TestCallNoteSaveNudgesHistory` (exactly one payload-less nudge to the
        owner's hub).
14. **`scripts/templ-generate.sh` wrapper** that pins the regen compiler to go.mod's
    version and asserts the generated header stamp (the 2026-11-05 near-miss class).
15. **Attended-transfer UI vs the NO-GO ruling** — `FEATURES.md` claimed the row
    FULLY_FUNCTIONAL while `docs/research/2026-10-27_attended-transfer-spike.md` says
    the bridge is UNVERIFIED (the stack has no Replaces dialplan handling); the status
    is now PARTIALLY_FUNCTIONAL. Decide whether the dock's "Attended" button should be
    hidden until a FreeSWITCH lab proves `Replaces` — the island's honesty contract
    forbids a 200-OK that silently does nothing.
16. [x] **`reverseLines` → `slices.Reverse` — DONE 2026-11-07 (P2-5)**: the
        hand-rolled two-pointer helper in `internal/readmodels/transcripts.go` is
        gone; its only caller now uses `slices.Reverse`, which also retires the
        `//nolint:branching-flow` the helper's own arithmetic needed. Behavior is
        identical (newest-first per call) and the readmodels suite is green.

17. **Make freshness WARN on a §f-less newest report** — the gate PASSED on the
    2026-11-06 10:45 report because it carried no §f at all ("nothing to check"),
    so the ritual certified nothing (the vacuous-PASS class). A newest report
    without §f should be a finding, not a green.
18. **Enforce CHANGELOG coverage for code-path commits** — the metaengine
    per-group cap train (2026-11-06 17:59) landed with no `[Unreleased]` line and
    nothing reds. Cheapest shape: a buildflow step diffing non-test Go/JS paths
    against CHANGELOG.md.
19. **Record the DF-brittleness + standalone-verb gotchas UPSTREAM** (docs-gates
    repo, Artmann-Technologies): `audit_max` shifts with living-doc growth because
    "generic" is a document-frequency threshold (observed 330 → 331 from one
    CHANGELOG paragraph, then 354 → 360 in one session); **LOCAL CALIBRATION
    2026-11-07 (P1-4 tail)**: `generic_df` is now 16 in `.docs-gates.json`
    (**the rule: a token on more than ~0.5 % of the living corpus's lines is too
    common to prove a harvest**, so it is re-derived from the line count each
    sweep) instead of the upstream default 8, which alone dropped the residue
    360 → 71 — the sweep of values (8→360, 12→181, 16→101, 20→76, 22→71, 30→40)
    is the evidence, and a planted nonsense §f item is still flagged. Upstream should scale the default with corpus size
    and expose it on the CLI (it is config-file-only today). Also: a standalone
    verb ignores `.docs-gates.json`'s ratchets while `run` honors them.
20. [x] **Frontend-UX verification tail — DONE 2026-11-07 (P1-4)**: three pins
        landed in `internal/web/views/settings_structure_test.go`. (a) the admin
        line-count copy is pinned in BOTH languages (it must name the transcript,
        keep its leading space, and stay one unbreakable German compound — so there
        is no wrap point inside "Transkriptzeilen"); (b) the F20 three-audience
        Settings structure is pinned as an ORDER (prefs < data < collapsed server
        `<details>` between them) plus a per-section dl-count pin (the 10
        hand-written shortcut rows, which no helper generates); (c)
        `#wp-palette-open` is now a real disclosure —
        `aria-controls="wp-palette"` + `aria-expanded`, kept in step by `shell.js`
        at runtime, documented in docs/dom-contract.md § Notes.
21. **annotate.go/docs-gates tooling nits** (found by the docs-health v9 sweep
    2026-11-07, upstream Artmann-Technologies repos): (a) escaped pipes inside a
    code span (`\|`) defeat the table cell splitter, so a per-cell strike
    silently crosses a cell boundary — hand-fixed one in the durability
    archive; **LOCAL MITIGATION SHIPPED 2026-11-07 (P1-3)**:
    `scripts/strike-safety.py rows` refuses the shape before a sweep writes and
    is a wrapper gate (13-case self-test); still worth the upstream scanner
    fix; (b) spec verdict values are not shape-validated — a literal "UNMAPPED"
    verdict passes the banned-words check and would strike as garbage;
    **LOCAL MITIGATION SHIPPED 2026-11-07 (P1-3)**:
    `scripts/strike-safety.py verdict` enforces the canonical heads; upstream
    should hard-error the same way;
    (c) "live reports" has three disagreeing counters (INDEX regen count, raw
    ls count, cadence gate count read 76/75/70 in one moment) — derive one
    from the other or name the differences;
    (d) the documented `<!-- fresh:skip -->` exemption does NOT actually exempt
    an item (tried directly above the item and above the §f heading; the gate
    still reported 2 unharvested). What DID work was a distinctive token in the
    item text (`go-health-dashboard`) — so the marker is either undocumented in
    its real placement or broken; verify upstream and fix the doc or the code.

### 2026-11-05 hygiene train (assistant sweep — tests + docs + repo hygiene)

Executed: the destructive sweeper `internal/retention` (zero pins since plan T25)
got its test package — a deterministic `sweepOnce` seam plus race-green pins for the
expired-tenant sweep with REAL file unlinks, the corrupt-path containment, the
disabled default, and the immediate-first-sweep loop contract (both containment pins
mutation-proven: they FAIL when the rule is broken). One semantic discovery pinned in
place: a transcript's retention clock is its ARRIVAL time (`Seq = now` at Append), not
`StartedAt`. Repo hygiene: the tracked regenerable artifacts untracked
(`buildflow-fsprobe-2193036538`, `ui-shots-before/` — the same doctrine as the
/ui-shots/ budget decision). TODO_LIST surgery: executed sections deleted, the
E2E-mention sprawl collapsed to ONE home (Cross-repo obligations), contradictions
fixed at their rows.

Deferred engineering findings (each verified at HEAD 2026-11-05; none urgent):

1. **CSP has no violation reporting** (`report-uri`/`report-to` absent from the
   server.go CSP const) — a live XSS-class event is unobservable. Needs an owner
   decision on where reports land (a `/csp-report` endpoint + Caddy wiring) before
   anything is wired.
2. **`/metrics`, `/version`, `/openapi.json` are fenced only by the module's vhost** —
   with `caddy.enable = false` they are open to anyone who can reach the port (the
   server.go comment says "fenced per-location by the module's vhost"; nothing
   enforces it). A module-level eval-time warning (or a bind-interface default)
   closes the silent-exposure class.
3. **The NixOS module's secret-leak guard misses freeform secrets**: it asserts the
   three typed secret files but not `gateway.webhook_secret` or a TURN secret inside
   freeform `settings` (those still land in the world-readable store if an operator
   writes them as literals).
4. **The templ form-POST surface has no rate bucket**: `/messages/send`, `/fax/send`,
   contacts import, snippets, flags are unlimited — only the `/api/*` lanes are
   bucketed. Decide: bucket them or document the session gate as the bound.
5. **`server.Deps` grows a field per read model** (seven read-model fields, each with
   a nil-means-503 convention riding every test composition) — DESIGNED 2026-11-08:
   docs/planning/2026-11-08_server-deps-accessor-collapse-design.md owns the shape,
   the migration order, and the test-composition impact. EXECUTE only when the
   eighth read model lands (the note's trigger).

### Fleet-libs closeout residue (2026-11-04, the SUPERB closeout train)

The 2026-11-04 fleet-libs audit's 26-task closeout plan
(`docs/planning/2026-11-04_18-41_SUPERB-fleet-libs-closeout.md`) executed T01–T15 the same
day; the residue below is what remains OPEN, with its reopen triggers. The landed work
(blob durability, MaxBodySize mounts, go-health/dashboard wave, Conflict family, entitytag
swap, fax badge, 2 empty states) lives in CHANGELOG + the per-library reports — not here
(one home per fact).

1. **go-sse v0.7.0 bump train — TRIGGER FIRED, TRAIN LANDED 2026-11-05**: loginpage
   v4.16.0 (with cqrs-htmx v4.16.0, usermgmt v4.19.0, go-health v0.10.0) accepts
   templ ≥ v0.3.1100; go.mod carries go-sse v0.7.0 + templ v0.3.1100 +
   templ-components v1.27.0, the views were regenerated, and the vendorHash tail was
   re-pinned (`T2Q64uGG…`) after the named-hash red. Full local suite green AND CI GREEN on the re-pin commit (run
   42075231935, 2026-11-05) — the train is fully landed.
   Residual, no consumer today: the ssetest `WithLastEventID(sse.EventID)` branding
   ride — schedule it only when a Last-Event-ID replay surface exists (the
   `/events` heartbeat row below is NO-GO until drop-class evidence).
2. **templ-components taste calls** — owner decisions, not adoption gaps: the ✓ delivered
   chip (or a Badge prefix/icon slot) and whether section subtitles should render as muted
   empty-states. Re-litigate only if the owner asks.
3. **go-idempotency upstream ask (2026-11-07)**: the contract suite's NoopOnExistingKey
   red-ed CI (run 42407471289) when a -race scheduling stall outlasted the scaled shortTTL —
   the second Record legitimately re-claims the expired window and the wall-sleep assertion
   reads it as an extension. Webphone's band-aid is TimingScale 5 (stall budget 60 → 100 ms);
   the durable fix is upstream: an injectable clock instead of wall sleeps. **2026-11-08
   (P4-7): the filing draft is committed** — docs/drafts/2026-11-08_go-idempotency_contract-clock-race.md
   (all five verify-before-filing gates run at go-idempotency HEAD 9a82577; check-draft 0 FAIL).
   Filing awaits the owner's review tick (the pack's D-slot is unanswered, so the unsolicited
   banner stands and posting is not the agent's call). Evidence: docs/ci-flake-ledger.md row 42407471289.

### Fleet-libs reopen-trigger registry (2026-11-04, T25 — the ONE table)

Every NO-GO/declined/keep-as-is ruling carries a written trigger in its report
(`docs/review/libs/*.md` § Ruling — the one home). This table consolidates them so a
sweep answers "is any trigger fired?" in one read; re-litigate a ruling ONLY when its
trigger fires, never on a schedule.

| Fires when                                                                                      | Library                | Then                                                                |
| ----------------------------------------------------------------------------------------------- | ---------------------- | ------------------------------------------------------------------- |
| metaengine grows NESTED group addressing (a per-call floor inside a per-owner group)            | go-cqrs-lite           | `trimCall` is the last hand loop; delete it for the declarative cap |
| usermgmt major bump, or a second identity consumer                                              | cqrs-htmx              | re-run the family audit (adminui re-check)                          |
| `internal/domain` gains a second Go consumer                                                    | go-branded-id          | ratified modularization trigger — brands need their own module      |
| A client-facing `Idempotency-Key` contract appears on a public API                              | go-idempotency         | re-run the adoption ruling (metaengine DedupStore vs the SDK)       |
| A self-managed versioned blob format or event signing requirement                               | go-codec               | `DeterministicCodec` is the entry point                             |
| A drop-folder/intake feature, or live-config-reload demand                                      | go-filewatcher         | re-run the NO-GO ruling                                             |
| A `Type=notify` conversion, or a unix-socket machine surface                                    | go-daemon, go-appkit   | re-run both rulings together                                        |
| A CLI subcommand outputs structured data                                                        | go-output              | adopt the renderer                                                  |
| A non-systemd target, credential-compliance rule, mutable secrets                               | go-credentials-store   | re-run the NO-GO ruling                                             |
| A new file-persisting feature (export, backup, rotation)                                        | go-atomic-write        | ride `WriteWithPerm`/`WriteIfChanged`                               |
| A latency investigation the histogram aggregates cannot localize                                | httputil/server_timing | instrument 2–3 named spans behind the same env gate                 |
| User-visible missed-update reports on flaky mobile links                                        | go-sse                 | ssetest-pinned Last-Event-ID replay                                 |
| A second retry-shaped loop (e.g. ASR uploads)                                                   | go-retry               | `retry.Do` is the natural home                                      |
| A dashboard deployment without the Caddy front                                                  | go-health-dashboard    | re-decide the fence (see the 2026-11-04 fence memo)                 |
| A read model crossing ~10k rows/owner, or a second engine                                       | go-cqrs-lite           | keyset pagination / failover audit                                  |
| A second readiness consumer or multi-instance deployment                                        | go-health              | re-run the probe-topology ruling                                    |
| A templ-components release with a new status/toast component; the subtitle/✓-chip taste rulings | templ-components       | owner-taste re-litigation, not adoption gaps                        |
| The dashboard library swaps its embedded SDK version                                            | go-datastar            | re-verify the /health CSP scope                                     |
| A product surface wants server-pushed reactive state (live Reports-shaped, not fragment morphs) | go-datastar            | re-run the client-side NO-GO (2026-11-07 deep-dive, posture 78/100) |
| The library ships a parse-with-rules hook or a new linter rule                                  | go-branded-id          | migrate the invariant INTO the seam / adopt the rule                |

### Fax enablement (owner + carrier action, 2026-10-30)

The app now DECLARES the fax lanes instead of offering a form that can
only fail (`fax.send` / `fax.receive`; the stack derives them: receive
follows the fax feed, send is off). Turning fax ON is not a flag flip —
it is two carrier facts, one per direction:

1. **Inbound**: buy/port a fax-capable DID and set
   `services.telephony.gateways.<gw>.faxDid` on the stack (the module
   asserts it differs from the voice DID), plus
   `services.telephony.fax.feed.enable` + `feed.secretFile` so rxfax
   TIFFs reach `/hooks/fax`. Then `fax.receive` turns on by derivation.
2. **Outbound**: nothing in the stack transmits faxes (the messaging
   bridge answers `/gateway/fax` with an honest 503). Needs either a
   Telnyx Programmable Fax application + fax-enabled number (a number
   binds to ONE connection, so that DID cannot also carry the voice
   trunk) or a bridge that implements `/fax`; then set
   `services.webphone.settings.fax.send = true` in the host.

Verifying the UI after either step: `python3 scripts/webphone-smoke.py
--bin <binary>` prints the `faxcap scenario` verdicts.

**DECISION (2026-10-30, owner): fax stays as-is — declared off.** No DID
purchase, no outbound path, no deletion. Record + consequences:
`docs/planning/2026-10-30_04-15_fax-decision.md`. This section remains
the re-open recipe; nothing here is scheduled.

### Paperless: the invisible seam (capability audit T31, 2026-10-30)

The fax audit's one real gap: `paperless.url`/`token` build the fax
archiver with NO Settings row and NO per-job archive marker — a
deployment that believes it archives learns otherwise only from the
journal. Suggested shape + the other three seams' clean verdicts:
`docs/planning/2026-10-30_14-05_capability-seam-audit.md`. Owner call
on the shape (Settings row alone vs per-fax archive marker); recorded,
not built. **The shape now has a decision brief** (2026-11-08,
recommending the Settings row, with the per-fax marker's cost and the
correlation machinery it would need):
[docs/planning/2026-11-08_decision-briefs.md](docs/planning/2026-11-08_decision-briefs.md)
§ Brief 1 — answer in the sitting pack.

### CRM data-mesh train (2026-10-29) — deferred tail

Source: [docs/status/2026-10-29_19-44_crm-data-mesh-train-closeout-status.md](docs/status/2026-10-29_19-44_crm-data-mesh-train-closeout-status.md)
§f. The train itself (typed `crm.Person`, deep links, context strip, the
CRM-side `GET /api/contacts/{id}/context`, the outcome-vocabulary bridge) is
SHIPPED and green; these are the deliberately deferred edges.

1. **"Add to CRM" for misses**: today a caller the CRM does not know is a dead
   end. The affordance is a link into the CRM's create form (still no
   browser→CRM traffic, no contact creation in the phone).
2. **Batch identity (`by-phone?numbers=` on the producer)**: lists ask N times;
   a batch product collapses it to one. Owner-side (Kith) work, worth doing
   when a list ever exceeds ~20 numbers per render.
3. **Contract snapshot test**: pin webphone's exact `/context` request + parse
   against the producer's published shape, so the vocabulary-break class (a
   closed set drifting under a consumer that keeps sending the old value)
   fails in CI instead of on a customer's answered call.
4. **Full `nix flake check` as ONE gate**: the KVM-gated
   `webphone-backup`, `webphone-backup-drill` and `webphone-module` checks were
   not part of this train's verdict set (each verified individually when it
   last changed).
5. **Kith side (`blocked` on Q3, below)**: the endpoint's docs landed
   (README, FEATURES, CHANGELOG); `/docs` is generated from the EventCatalog and
   does NOT cover machine-API REST routes, so no generated-wire entry is owed.
   What remains is the PUSH: the branch is 16 commits ahead of `origin/master`
   and mixes another session's claim/outreach work, whose `internal/catalog`
   gate is red.

**Owner questions (still open):**

- **Q1 — Email/SMS/fax facts into the CRM timeline?** Today only call metadata
  crosses. Sending message bodies would make the phone a producer of the
  CRM's relationship history (privacy + noise decision).
- **Q2 — Actionable strip?** Completing a task or moving a deal from the phone
  would flip the read-only ownership rule this train is built on.
- **Q3 — Who pushes Kith?** Whole branch as-is / cherry-pick the CRM-context
  commits onto a clean branch / leave it to that session.

### issue #9 + SUPERB execution tails (2026-10-29 closeout train)

Harvested from the SUPERB execution plan (T22-T24 + the owed-debt sweep). The
dial-gate seam (#7) is SHIPPED end-to-end: server gate + island chip + API
tests + smoke scenario + the apidocs/registry landing; the automatic review
(#8) shipped server + History render; the external-claims debt (T19) closed
with the verification comment on #6.

1. **Kith CRM live e2e (#9 — OWNER-GATED, `blocked`)**: the pbx-artmann host
   wiring landed (`5df19b5` — dnsblockd CA into the system trust bundle via
   `security.pki.certificateFiles`, `services.telephony.webphone.crm` at
   https://crm.home.lan, the `webphone_crm_token` secret slot in
   generate.sh; issue #9's `trustedCertificateAuthorities` option name does
   not exist in nixpkgs and is corrected in the #9 comment). Owner steps:
   mint the Kith API token, write it to
   `~/.pbx-prod-secrets/webphone_crm_token`, run `scripts/push-secrets.sh`,
   deploy the pbx-artmann train, then verify the acceptance bar (inbound
   caller name; `call_logged` on the contact journal; a revoked token turns
   the Settings tri-state red).

### ai-phone steal-list (2026-10-29 deep dive — 3 SHIPPED, 4 actions open)

Source + grading: [docs/research/2026-10-29_ai-phone-deep-dive.html](docs/research/2026-10-29_ai-phone-deep-dive.html)
(oxnw/ai-phone read at source; the report carries the ten mechanics and the six declines).
Shipped the same session: the docs-versus-code config pin (`internal/config/docpin_test.go`, which
immediately documented eight previously undocumented keys) and generation-closing for live transcripts
(`closeCallSegments` + the announce guard, two island specs).

1. **Config typo strictness (L7, sharpest).** koanf `Unmarshal`
   (the koanf `Unmarshal` inside `config.Load`) silently drops unknown keys from the JSON file AND from
   `WEBPHONE_*`, so a typo turns a seam off instead of failing the boot. Design is constrained by
   the surface itself: strict (error) for the FILE, warn-only for the ENVIRONMENT, because a dozen
   `WEBPHONE_*` variables are operational and not config keys (`WEBPHONE_CONFIG`,
   `WEBPHONE_DEBUG_TIMING`, `WEBPHONE_RELEASE_MAX_LOAD`, `WEBPHONE_SSE_OOB`, …) and a false reject
   is an outage on every deployment. Owner call: the NixOS module writes a freeform `settings` map,
   so file-strictness changes boot semantics for a deployment carrying an extra key today. Pairs
   with the `docpin_test.go` key tree (same reflection helper can feed the allowed-key set).
2. **Agent liveness contract for D3 (L6).** Prior art worth borrowing before the agent-token lane is
   designed: ai-phone's relay mode answers a stalled client with a filler at 1.5 s, another at 10 s,
   a polite goodbye and BYE at 25 s; a dead `listen` client frees the listener slot on socket EOF so
   the next call falls back to the autonomous mode; `hangup` waits 30 s + 10 s and then fails with a
   named exit code. webphone's textagent has a 3 s probe and a 502 toast (covers unreachable, not
   "accepted then silent") and no designed fallback for the human waiting on a draft. Design input
   for D3 LANDED 2026-10-29 (M18/F75-F77):
   `docs/planning/2026-10-29_16-17_D3-agent-liveness-design.md` — liveness ladder, slot release,
   named exit codes, and the webphone seam map; still owner-GO + security-review gated.
3. **Detection column in the error contract (L9).** `docs/error-contract.md` answers "what does the
   operator SEE"; ai-phone's failure table also answers "how would anyone KNOW it happened"
   (counters, warm-start timings, a logged line-override). Add a detection column for the quiet
   classes at the next pass over that document.
4. **Caller-perceived latency (L3).** The histogram covers HTTP durations by status class; nothing
   measures dial press → first audible media, or the ASR segment round trip. ai-phone's rule is the
   sharpest phrasing of why that matters: t0 is the last speech, and the silence is INSIDE the
   budget. Lowest of the five: needs a reporting home and an island metric, so measure the question
   before instrumenting everything.

### Island type gate tails (2026-10-27/28 — gate SHIPPED and CI-WIRED; refinements below)

The gate itself is done and enforced: 680 type errors → 0 over 26 served files, `island-typecheck`
now runs in CI's `fast-checks` job (it was green locally and unwired in CI until 2026-10-28), the
how-to is [docs/type-gate.md](docs/type-gate.md), and the pins from `typecheck_pin_test.go` grew from
5 to 10 assertions (excluded glob reaches the bundle and nothing else, no narrowed include, the
`IslandDomIds` mirror of docs/dom-contract.md, plus a `globMatcher` table test).
Evidence: [docs/status/2026-10-27_22-00_island-type-gate-complete-status.md](docs/status/2026-10-27_22-00_island-type-gate-complete-status.md).

The three refinement rows that sat here (whitespace-drift false positive,
`__wpClock`, branching-flow triage) are SHIPPED 2026-10-28 — the gate now
adjudicates formatter output with the flake's own treefmt over the
materialized staged blobs (`--ref` mode too), the shell reads all time
through one exported clock holder with the global deleted, and the
branching-flow set is fully governed (36 reasoned nolints + a fresh
78-entry baseline wired into `.buildflow.yml`). See CHANGELOG [Unreleased].

- **DECLINED with reason — do not re-litigate without new evidence:**
  - _Split `shell.js` at its seams_ (voicemail player / palette / composer). The gate now makes a
    split SAFE (a broken import is a compile error), but the classic-script load order is a
    served-byte contract the stack E2E greps, and the payoff is readability only. Reopen with an E2E
    budget.
  - _Type-check `island-tests/*.mjs`._ Needs `@types/node` (not in the devshell) plus a third
    tsconfig, and the specs' stubs are deliberately partial duck types — the checker would fight the
    design the `hasMethod` doctrine rests on.
  - _Move the shell's copy into a typed key map._ The shell must work when the island scripts fail
    (that is why it is a classic script); a shared map adds a load-order dependency in the one place
    the shell's independence is the point.
  - _`tsc --explainFiles` in CI._ The coverage pin derives the checked set from the EMBEDDED tree, so
    a file that stops being included fails the pin — proved 2026-10-28 by narrowing the include to a
    single file (both coverage pins failed with an actionable message).

### Error-surface + crash-containment plan (2026-10-27 — P0–P2 SHIPPED, host legs + evidence-gated tails open)

Plan + execution log: [docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md](docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md).
Shipped same-session (see the plan's execution log + CHANGELOG): recover() guards on ALL
five bare-goroutine sites + containment test, the island catch sweep (1 bug fixed:
summarize network-dead split; justifications at every silent site), the call-failure
reason card + THE SIP-cause mapping (error-contract.md, coordinates with #6 below),
oxlint `no-empty`/`no-unused-vars`/`no-console(allow info·warn·error)` gates + finding
fixes, the shutdown release-count log, push draft wrap-code pins, errors.AsType, the
502-proxy + island-JS error-contract sections (+ stack runbook sync), lessons entries.
**Priority:** High · **Effort:** S remainder.

1. **Host-leg verdicts (OWNER, read-only ssh — banned in the agent shell)**: run
   `ssh root@pbx.artmann.tech 'systemctl show webphone -p NRestarts -p ExecMainStartTimestamp; journalctl -u webphone --since today | grep -iE "panic|fatal"'`
   and paste the output; a panic line reopens P0 forensics on the named seam. Same
   pass covers the Caddy /events flush+logging config check and the dialplan
   `state:end` ring-hook probe (plan M03/M53/M54). Prod `/version` already verified
   REMOTELY: v2.9.0 at `ec5de5e` (deploy fired; `agent:true` live).
2. **`/events` heartbeat + server `retry:` hint — NO-GO until drop-class evidence**
   (plan T19): recovery is guaranteed by the upstream recreate loop; no long-idle
   drop has ever been observed. Reopen trigger: a journal/SSE-log pattern showing
   feeds dying while the process stayed up.
3. **Upstream filing (owner + github-voice)**: the htmx SSE ext `console.error("[object
   Event]")` defect — diagnosis source-verified, draft note ready:
   [docs/research/2026-10-27_htmx-sse-ext-object-event-logging-note.md](docs/research/2026-10-27_htmx-sse-ext-object-event-logging-note.md).
4. **Post-deploy deep-verify (plan T18/T22 tail)**: after the next release reaches
   prod, confirm the reason card + offline toast + reconnect pill title render on
   the live deployment (browser E2E owed only if served markup changed — it did
   not: all surfaces are island-JS-created).
5. **Flush-announce island spec (harvested 2026-11-08 from the 10-27 error-surface
   report §f-15)**: the announce paths landed (`transcribeLag`/`transcribeUnavailable`
   in transcribe.js) but no island spec covers them — pin the laggy-streak
   announce-once behavior with a node:test spec.

### Language stance: island type gate SHIPPED, Go→JS + full KMP rejected (2026-10-27 plan)

> **Item 1 CLOSED 2026-10-27 21:00** — the gate is enforced and green over all 26 served asset
> files (`tsc -p . --noEmit`, wired as `island-typecheck`), with 680 type errors retired and ten
> real defects fixed on the way. `tsconfig.all.json` is deleted and the debt list was replaced by
> stronger coverage pins. Execution record:
> [docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md](docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md)
> §9. Item 2 (the error-surface conventions) landed in the same train; items 3–4 stay open.

Plan + evidence: [docs/planning/2026-10-27_16-04_SUPERB-typescript-and-platform-language-pareto-plan.html](docs/planning/2026-10-27_16-04_SUPERB-typescript-and-platform-language-pareto-plan.html)
(the plan owns the task count; this row points at it). **Priority:** High · **Effort:** M.

3. **KMP protocol-layer spike (CONDITIONAL — T26/T27, `~/projects/webphone-android`)**:
   the codec + digest move to `commonMain` cleanly; `SipEngine` needs its okhttp transport behind
   `expect`/`actual` first (re-estimated at 180 min, not 15). Android stays the ONLY target, then
   STOP — no second target without a linuxX64 WebRTC 3-call audio smoke. Gate: a committed second
   target in ROADMAP. **DEFERRED 2026-10-28**: the android repo carried a LIVE concurrent
   call-stack train (daemon commits minutes old, Kotlin files mid-edit) — Gradle surgery on
   settings.gradle.kts + a new shared/ module would collide head-on. Tier-0 by the plan's own
   wording ("may stay unstarted indefinitely"); execute when that repo is quiet.
4. **REJECTED, do not re-propose**: Go server → JavaScript (57,349 LOC plus the
   single-binary/nix deployment premise for zero product gain) · full Kotlin Multiplatform (the
   2026-10-23 platform decision stands: Linux WebRTC is the weak leg) · a Kotlin/Wasm web target
   (rewrites the island + templ/htmx model) · island → `.ts` with emit (deferred behind three
   named triggers: a justified npm dependency, a JSDoc overhead ceiling, framework adoption).

### Dialer-adoption issues #6–#8 + Kith switch-on #9 (2026-10-27 — claims verified; #6 + the #9 config/health slice SHIPPED same evening)

UPDATE 2026-10-27 evening: #6 is SHIPPED — typed `domain.Disposition`
(answered/missed/voicemail/ivr/failed/unknown per the stop-the-lying
ruling, NOT the dialer's seven-value set), legacy outcome demotion,
ended-card confirm, `call_logged` verdict travel, history verdict
filter + badge. #9's typed `crm` NixOS lane, `crm.token_file`, and the
`/healthz` CRM leg are SHIPPED. #7 is SHIPPED 2026-10-28 (row 2); #8 is
MOSTLY SHIPPED (row 3 — the dimensions slice is the open rest). The
per-item rows keep their original wording for the audit trail.

Owner-filed adoption candidates from the dialer postmortem
(matthewdonsemail-lab/dialer PRs #20/#54); every gap claim verified against
HEAD, corrections posted as issue comments. Separate trains, none started.

1. **Call dispositions** (#6): operator-confirmed taxonomy
   (`connected`/`voicemail`/`ivr`/`no_answer`/`wrong_number`/`dnc`/`disconnected`)
   in `internal/domain`, default `connected` on Established, confirm through
   the CallNote UI; ONE total CDR-status→disposition mapping (unknown ⇒
   failure class, never answered); awaited write with surfaced failure; carry
   the disposition in the Kith `call_logged` payload. Correction recorded:
   `Outcome` IS filled today — the island sends
   `established ? "answered" : "missed"` (the island's `recordCrmCall` → `api.Handlers.LogCall`),
   so a voicemail call journals as "answered" in Kith right now.
2. **Per-remote redial cooldown** (#7) — SHIPPED 2026-10-28 (seam by the
   morning dialer train, island + tests + docs completed by the
   whole-list evening session): `dial.cooldown` (off = byte-identical),
   readmodels `dial_windows`/`dial_overrides` (restart-proof, backed
   up), `POST /api/dials` check+record gate (denied asks record
   NOTHING — a retry loop cannot extend its own window; the verdict
   carries `last_at` + `window`), one consult in `placeCall` covering
   EVERY dial path (form submit, redial, click-to-dial), the inline
   state line + one-shot "Dial anyway" chip under the dial form (never
   a modal), 404/unreachable dial anyway with the off verdict latched.
   Pins: dialgate_test.go (3 e2e cases) + dialgate.test.mjs (6).
   DECLINED for now: a pre-dial cooldown badge in the typeahead ROW
   would need a peek (read-only) endpoint — the gate's check RECORDS,
   so typing must never consult it; reopen when an operator asks.
   (Was: no cooldown exists today — true at the 2026-10-27
   verification; the morning train landed the seam.)
3. **AI call review artifact** (#8): UPDATE 2026-10-28 (multi-actor train
   read at HEAD): MOSTLY SHIPPED by the concurrent dialer train — the
   post-call review is AUTOMATIC (`settleEndedCall` dispatches
   `POST /api/calls/{id}/review`, 202 detached, idempotent with the manual
   button; island calls.js), the artifact IS persisted
   (`domain.ReviewState` pending/present/failed via readmodels
   `call_reviews`, rendered in History — `TestHistoryTranscriptReviewStates`
   pins the four renderings incl. the not-configured chip). Remaining
   open slice: the summary/points/score DIMENSIONS the plan named (today
   the artifact stores the draft text + confidence only) — reopen only if
   the operator asks for richer artifacts.
4. **Kith CRM switch-on, webphone slice** (#9) — WEBPHONE SIDE FULLY
   SHIPPED (2026-10-27/28; verified at HEAD 2026-10-28 evening): typed
   `crm` module lane + `crm.token_file` (LoadCredential), the `/healthz`
   CRM leg (a revoked token reddens healthz naming "crm" on the very
   next request — pinned in healthz_test.go), and the Settings
   auth-failed surface (`crmDetail` renders FOUR states: unset /
   configured / connected / AUTH-FAILED via `crm.HealthCodeUnauthorized`
   on the bounded per-render probe — panels.go settingsCRMAuthFailed).
   The issue's steps 1/2/4 (PKI trust, token provisioning, e2e verify)
   remain pbx-artmann-side cross-repo work; the cited diagnosis doc
   lives there too. (The row's earlier "no auth-failed surfacing"
   wording was stale — the T7 four-state work had landed it.)

### Multi-actor HITL (2026-10-28 — substrate SHIPPED same evening; escalation tiers below)

Plan + HITL ladder + actor model:
[docs/planning/2026-10-28_21-54_SUPERB-multi-actor-hitl-pareto-plan.html](docs/planning/2026-10-28_21-54_SUPERB-multi-actor-hitl-pareto-plan.html).
SHIPPED in the same train: session `LoginEmail` (passkey/OIDC mint,
guarded migration, byte-parity for extension logins), audit lines naming
the human actor (`agent *` + logout; `TestAgentAuditNamesTheHumanActor`
pins both shapes), `authmap.MappedByLogin` whoami preference +
deterministic shared-extension lookup, drafts-panel pending age
(`draftAge`, foreign-wire tolerant). **Priority:** per-row · **Effort:**
per-row.

1. **Extension switcher (D1)** — SHIPPED 2026-10-28 (whole-list
   execution session): `POST /api/session/switch` re-mints the session
   for another extension mapped to the SAME login email (fresh
   password-file read + directory verify + CSRF rotation + old row
   deleted after the mint; tight PBX-bound rate bucket; `auth switch`
   audit line names the human move; extension logins refused, own-line
   switch an idempotent no-op, 404 with the identity layer off). The
   session identity carries `extensions`; more than one renders a
   JS-created select by the whoami line (no served-markup delta, no DOM
   contract change); selecting a line POSTs + reloads through the ONE
   boot path (re-REGISTER, hub resubscribe, panels — the no-reload
   sequence stays future polish), and a LIVE CALL blocks the switch.
   Pinned: session_switch_test.go (6 cases) + session-switch.test.mjs
   (5 cases).
2. **Veto-window mode / HITL L2 (D2)**: draft auto-sends after T unless a
   human vetoes; the missing rung between approve-everything (L1) and
   confidence-gated autopilot (L3, shipped agent-side). Lives in the
   stack's textagent (send policy) — webphone side only surfaces age +
   countdown (the age substrate landed). Owner GO + stack train.
3. **Agent registry + scoped machine tokens (D3)**: today ONE machine
   actor (the textagent bearer secret, loopback). Named agents with
   scopes (draft/read/send) + per-agent audit prefixes + an API-token
   lane for agents-as-users. Owner GO; security review first. V1
   DESIGN (ready for that review — nothing below is implemented):
   config `auth.agents: [{name, token_file, scopes: [read|draft|send]}]`
   (LoadCredential like oauth2/crm — never a raw token in config);
   a `RequireAgent(scope)` middleware answering `Authorization: Bearer`
   with constant-time compares, initially LOOPBACK-ONLY (the textagent
   posture — widen only with the stack's mTLS/TLS front); audit lines
   carry `actor=agent:<name>` through the EXISTING agentAudit
   vocabulary (the D5 durable rows inherit it for free); scope classes
   map to route families (read → GET /api/*, draft → review/draft ops,
   send → approve/send). NO second machine consumer exists yet (the
   textagent is the only one) — the registry's value materializes when
   one does, so the security review can also rule "defer until the
   second agent is real".
4. **Delegation + roles (D4)**: approve-on-behalf-of, per-contact scopes,
   admin/member split — needs the actor substrate (shipped) AND a second
   real user first. Owner call. UPDATE 2026-11-04: the READ-ONLY half of
   the admin split SHIPPED (the Admin tab: all messages/faxes/calls/
   users/numbers across lines for `admin: true` accounts, resolved per
   request — see FEATURES § Contacts & sessions); what remains open is
   delegation (approve-on-behalf-of) and any WRITE-shaped admin surface,
   both still gated on a second real user.
5. **Durable audit read model (D5, F107 reopen trigger)** — SHIPPED
   2026-10-28 (whole-list execution session): readmodels `agent_audit`
   (append-only, owner-scoped, 500/extension oldest-first trim) +
   `agentAudit` DUAL-WRITE (the greppable line stays; a store failure
   WARNs, never blocks the action) + the Settings "Recent agent
   actions" section (`wp-agent-audit-*`, rows-or-seam conditional,
   actor resolved to the login email). Pins: agentaudit_test.go (store)
   - TestAgentAuditDualWritesTheDurableRow + TestSettingsAgentAuditSection.
     The LOGIN-event durable journal stays usermgmt's (identity seam
     untouched); an in-UI login-history surface would need an upstream
     read API — reopen on an operator ask.
6. **Branching-flow baseline staleness (observed 2026-10-28 train-start
   buildflow)** — VERIFIED HEALTHY at HEAD (2026-10-28 evening, whole-list
   session): the 11:20 re-establishment (`ba6e189`, 78 entries) is LIVE —
   74/78 ids match current findings exactly; the earlier "matched nothing
   in 3 runs" observation described the PRE-re-establishment state and is
   retired. Current drift (expected cadence work, findings are non-gating
   warnings under the SDK severity cap): 4 baseline entries healed away,
   7 live findings are new since 11:20 (the day's trains: D1's
   `session_api.go` DUPLICATE_TYPE + Deps-grown `server.go` large-struct,
   the chat trains' `matrix/receiver.go` + `chatloop.go` sites). The next
   governance triage sweep (per-site rationale or fix for the 7, then
   `buildflow -s branching-flow --save-baseline` from THIS checkout)
   owns absorbing them — not re-baselined unilaterally here.

### Daemon push-stall forensics (2026-10-27 T10 — root cause found, fleet fix pending)

The evening's push stalls are the daemon's push leg dying with its
PROCESS, and fleet deploys bounce that process: the 20:19:41
`switch-to-configuration test` (SystemNix) terminated
`projects-management-automation.service`; pushes stopped 20:09:15,
resumed 20:40:36 (webphone's oldest unpushed commit waited 30m37s;
follow-on gaps hit 54m and 70m). The journal logs NOTHING about pushes
(the T21 LOCAL-timer verdict stands), and the commit-message LLM
provider (127.0.0.1:52625) refused connections 13:06→20:59 — ~8h of
heuristic-only messages fleet-wide, though the commit leg itself stayed
fast (177 ms fails). FLEET ASK FILED 2026-10-28 as
projects-management-automation#348 (verified at source first: Stop()
drops pending batches with no drain, Start() has no unpushed sweep,
per-push logging already exists — so the ask is the SIGTERM flush, the
startup sweep, and a shutdown-completion line; draft in
docs/drafts/pma-push-leg-durability-issue-2026-10-28.md). Until it
lands, a fleet deploy window predicts a push stall — run
`scripts/push-stall-watch.sh` after every `switch-to-configuration`.

### Multi-call dock + agent-UX train (2026-10-26/27 SUPERB plan — SHIPPED, two tails)

Pareto plan: [docs/planning/2026-10-26_08-11_SUPERB-multicall-dock-agent-ux.md](docs/planning/2026-10-26_08-11_SUPERB-multicall-dock-agent-ux.md)
(20 medium / 120 micro tasks; round-two errata in the plan doc: M15 search
already shipped, notes home = readmodels not the store ladder, M18 durable
journal NO-GO, M19 NO-GO on this host). M1–M14 + M16–M18 shipped across
2026-10-26/27 (ring queue, dock, captions, ended-card summarize chain,
agent-mode pill, draft-ready push, call notes with AI prefill, locality
labels, audit lines). Open tails:

1. **M19 re-open trigger — attended transfer**: the stack grows an FS
   lab with an attended-REFER-with-Replaces VM scenario; the island
   delta is small (sketch in
   `docs/research/2026-10-27_attended-transfer-spike.md`).

### Error-agents tooling train (2026-10-26 — SHIPPED; owner calls remain)

Pareto plan: [docs/planning/2026-10-26_10-33_SUPERB-error-agents.md](docs/planning/2026-10-26_10-33_SUPERB-error-agents.md)
— all 20 M-tasks shipped the same day (error-code router, reproduce-first
check + advisory CI job, the two watch.yml observers, bump-preflight,
incident-bundle, the KVM-gated error-drill VM test, the two gated
sketches; evidence in the plan + CHANGELOG). REMAINING owner calls: the
reproduce-first blocking flip (after N clean advisory trains) and the
two gated GOs (LLM prose gate, agent dispatch — trigger criteria
recorded in `docs/drafts/gated-sketches/`).

### Chat networks follow-ups (2026-10-23 train — matrix + signal shipped, session deferred)

Foundation (network axis + sender registry + receivers) and both shippable lanes
landed (`internal/matrix`, `internal/signal`, store v5, config gates, Settings rows;
plan + research evidence:
[docs/planning/2026-10-23_15-40_SUPERB-chat-networks-pareto-plan.html](docs/planning/2026-10-23_15-40_SUPERB-chat-networks-pareto-plan.html)).
Remaining, unscheduled:

- **Session messenger: NO-GO** (researched at source: no Go library anywhere, no
  official headless daemon/bot API; `libsession-util` is C++ config-layer only;
  the community `session.js` is Bun + bus-factor 1). Reopen ONLY on a named
  trigger: (a) an official/foundation headless API, (b) a maintained Go
  implementation, (c) an owner call accepting a research-grade `session.js`
  sidecar (then the stack repo owns the service first). **Priority:** parked ·
  **Effort:** n/a until a trigger fires.
- **Stack deployment work** (nix-international-telephony repo, NOT this one):
  signal-cli systemd unit (nixpkgs 0.14.9, `daemon --socket`, linked-device QR
  bootstrap ritual, 3-month freshness watch) + matrix homeserver choice/config;
  both feed webphone via `settings`/`environmentFile` — no webphone module
  change required. **Priority:** Medium · **Effort:** M.
- **Lane v2 fences**: read receipts (matrix `m.read` → delivered) and the
  compose-new-DM affordance are SHIPPED 2026-10-28 (`ParseComposedAddress`
  shape routing + `chatloop.Sink.RemoteRead`; CHANGELOG [Unreleased]). Still
  open: matrix group rooms (model spike DONE 2026-10-29 —
  `docs/planning/2026-10-29_12-20_matrix-group-rooms-model-spike.md`
  recommends A1: room-shaped `!…` remotes + nullable `Message.Sender`,
  awaiting the owner ruling before any execution), matrix E2EE (research
  DONE 2026-10-29 — `docs/research/2026-10-29_12-35_matrix-e2ee-go-story.md`:
  mautrix-go goolm is the only pure-Go path, no official rust-sdk Go
  bindings exist (verified at source); recommends P2 OlmMachine-bridge,
  decrypt-first; owner owes P1/P2/P3 + crypto-store-backup + trust-posture
  rulings), signal inbound attachments — **SHIPPED 2026-10-29** (M12/F55-F57:
  `signal.attachments_dir` arms them, per-attachment failures keep the text,
  fence stays byte-identical without the key; the signal-cli side is just a
  directory mount, the consuming stack owes nothing),
  signal groups (inherit the spike's Sender design verbatim once ruled).
  **Priority:** Low · **Effort:** M each.

### /health operator surface (2026-10-23 — stack option shipped, deploy pending)

`services.telephony.webphone.dashboard` now mounts the go-health dashboard at
`/health` with the nginx location UNBUFFERED and fenced (operator SSO in
`forwardAuth` mode, loopback-only otherwise); pbx-artmann sets it. **Shipped
2026-10-23:** webphone and the stack are pushed, pbx-artmann relocked
(`2fed91b`), and the host evaluates — the `pbx.artmann.tech` vhost renders
`= /health` + `/health/` with `auth_request /_telephony_auth_operator` and
`error_page 401 = @telephony_signin_operator`, and
`services.webphone.settings.dashboard = { enable = true; title = "PBX webphone"; }`.
What remains is the HOST REBUILD plus a browser check that logged-out `/health`
redirects to the SSO gate and a signed-in operator sees the cards, plus the
exposure-policy decision (remote_ip vs PublicMode vs basic auth — carried from
the samber/do train follow-ups; everything else in that section shipped).
**Priority:** High · **Effort:** S.

### Catalog adoption (2026-10-24 — SHIPPED + pushed 2026-10-25; the hub re-pin is cross-repo)

webphone is a full federated source of catalog.home.lan: the four seam events
(documented AS change notifications, NO commands declared — nothing fabricates a
CQRS write model), the `catalog` CLI subcommand, export tests, and
`.eventcatalog-hub.json` (the 2026-10-19 inventory-only record is superseded).
`46cc6db` is on origin/main. **The cross-repo leg EXECUTED 2026-11-08 (P4-10):**
the hub rides `branch: main` (no sha pin to bump — the "re-pin" is the mirror

- build), the Forgejo webphone mirror was force-synced to the GitHub tip
  (`mirror-sync-sources.sh webphone` → ✓ at tip), and hub run #240 is **GREEN
  end-to-end** (export, merge, governance lint, site build, publish) — reached
  through three real repairs: the CI host's runner state-root ownership (a
  NixOS switch renumbered the DynamicUser; chown + a durable-fix row in the
  hub's TODO), the crm source's stale go-modules vendorHash (re-pinned at the
  66f37f3 tip in the crm repo, `37a1957`), and a hub workflow step-scoping bug
  (`current_contract` derived in the build step but read in the verify step —
  `7551d4a`). The site published; the hourly sync pulls it to catalog.home.lan.

### Wire contracts (2026-10-24 — SHIPPED; three open rows)

The three catalog lanes cover the three real vocabularies (eventcatalog federation
= seam events, OpenAPI = browser JSON, AsyncAPI = machine-to-machine; the split
rule lives in [docs/contracts/README.md](docs/contracts/README.md); generation
evidence in the 16-50 plan + its execution status). Open rows:

- [ ] **OWNER-GATED: catalog v4.9.0→v4.10.0 bump** — 213 commits
      (architecture-reference + export-selection features webphone does
      not use; the seven upstream candidates are NOT obviously fixed).
      Lean: ride the next natural train. Same notes file. **Priority:**
      Low · **Effort:** S (decision) / M (bump).
- [ ] **webphone-android consumes the specs** — fetch/parse
      `/openapi.json` in the client bootstrap, derive the push-devices
      client calls, parse the asyncapi ring contract, verify the CSRF-dance
      paths against the spec (completion-plan N14; other repo's gates).
      **Priority:** Medium · **Effort:** M.
- [x] **DONE 2026-11-04 (ETag half): /openapi.json conditional GET** —
      the go-etag/server middleware serves bodyless 304 revalidation on
      `/openapi.json` + `/manifest.webmanifest` (mount is the skip set;
      `TestConditionalGETRootDocuments` pins it). Remaining OWNER CALL:
      the golden fixture (byte-pinned generated spec). **Priority:** Low
      · **Effort:** S.

### Declarative/DDD model train (2026-10-14 — EXECUTED; owner calls remain)

Trains A/B/C/E DONE 2026-10-14, D DONE 2026-10-18 (island CALL_TRANSITIONS +
table-gated hold machine, `ef29ea9`; delivery machine, typed events +
`Notifier.Project`, `AppendToThread`/`faxMoves` invariants — evidence in the plan:
[docs/planning/2026-10-14_17-15_SUPERB-declarative-ddd-model.md](docs/planning/2026-10-14_17-15_SUPERB-declarative-ddd-model.md)).
Remaining OWNER CALLS: A1.7 out-of-order verdict tightening (recommend: keep
today's overwrite); E2 ratify-vs-delete DOMAIN_LANGUAGE.md (memo in the file,
ROADMAP cross-linked); g3 sequencing (lean: own fresh session).

### SUPERB unified-auth train tails: owner-gated cutover + hardening remainder (2026-10-12 13:45 harvest)

**Status:** 🟡 `PARTIALLY DONE` (T08/T11/T12/T14 closed 2026-10-12; the same-day harvest session executed the 10-52 report's non-gated items — VM suites green with the new asserts; only the rows below remain) · **Priority:** High (cutover cluster) / Medium (tail) · **Effort:** M

- **OWNER-GATED — T13/T15 cutover cluster (do not self-answer; questions in ROADMAP § Open questions):** pbx-artmann relock to stack HEAD (clean tree — narHash covers all; release-runbook dance) → staged prod cutover `telephony.directory.mode = "xml_curl"` + explicit `staticFallbackExtensions` (operator + ring groups; members that cannot webphone-login MUST be listed) → pre-cutover backup + the <2-min rollback drill rehearsed → verify the derivation secret file lands with correct owner/mode for BOTH webphone (DynamicUser read) and the bridge LoadCredential → post-cutover CI + journal watch (`auth logout`, `directory query` lines prove the seam live) → T15 live proofs (derived-credential REGISTER browser+one softphone, logout-revocation live, static extensions unaffected, rollback drill ONCE on prod) → pbx-artmann CHANGELOG entry written IN the cutover commit.
- **Stack hardening tail:** port the bridge answer-count journal + seam-content-assert pattern into the OTHER webphone suites (shared helper in `tests/common.nix`) · shared `secretFile` fixture helper (`user = <service>` pattern) in `tests/common.nix` · consider merging telephony-webphone-dynamic-directory INTO telephony-webphone (one machine, halved CI cost; keep assertions readable) · wire `fs-directory.service` into `operatorWatchedUnits` under xml_curl + consider a health leg joining telephony-health (verified open 2026-10-12, pbx.nix `operatorWatchedUnits` lacks it) · pin the sipcred wire vector into the stack fixture (the VM proves behavior; the cross-repo HMAC vector is untested there) · xml_curl × fail2ban interplay check (digest-failure bans still apply?) · fs-directory journal noise under REGISTER storms (proxy-side rate limit?) · stack-repo one-liner at next touch: name the single-go-mod re-open trigger in the stack's consuming docs (routed from the 10-14 go-modularize §f item 40, 2026-10-18).
- **Webphone/infra tail:** daemon push-stall watchdog LANDED 2026-10-12 (`scripts/push-stall-watch.sh`, exit 1 when the oldest unpushed commit exceeds the 60-min ratified anchor; `--threshold-minutes`/`--no-fetch`; 4 arms self-tested in a scratch repo) · daemon-enforcement hook LANDED 2026-10-12 (`.githooks/commit-msg` blocks daemon-pattern commits touching `go.mod`/`go.sum`/`go.work*`/`flake.lock`/`.buildflow.yml`; `core.hooksPath` now RESOLVES — it had dangled since repo birth; `--no-verify` bypasses for real tool bumps) · **GUARD HOLE triaged 2026-10-15 (M10 μ44, evidence `89cf3b9`): the hook NEVER binds daemon commits** — the pma unit exports `GIT_CONFIG core.hooksPath=/var/empty` (env config overrides repo + global config), so daemon git runs hookless; hook logic itself VERIFIED (temp-index repro refuses exit 1; `scripts/commit-msg-guard-selftest.sh` pins 3 cases). FIX = owner edit of the nix-config pma module (drop the GIT_CONFIG_* override; fleet-wide call — hooks then bind the daemon in ALL projects). Interim exposure: daemon can still land go.mod/.buildflow.yml — watch daemon commits touching guarded files. · smoke-harness vacuous-pass guard: VERIFIED CLEAN 2026-10-12 (bool() guards at every wire compare in webphone-smoke.py; no double-extraction compares shipped — apply the rule when a credential-equality leg lands).
- Closed 2026-10-12 by the harvest session (rows deliberately NOT minted): VM suites value-pins (`sipCredential:true`, auth-logout audit assert, static auth-register leg), `apps.check-vm` sequential VM driver, FS module-mirror eval check + ONE-home list, ci-posture ledger, CI-registration investigation (`25cf08d` HAS run 38211100788 — the 10-52 claim was stale; `c2caead` was simply unpushed), README/error-contract/lessons/runbook docs, wsprobe exit-code docs, AGENTS tri-repo FS line. The /tmp passkey smoke script evaporated with /tmp — its arms live in the VM suite + go tests now; recreate only on demand.

**Evidence:** `docs/status/2026-10-12_10-52_SUPERB-train-completion-status.md` (f)+§g; stack runbook § Dynamic SIP directory; session VM run log (telephony-webphone + dynamic-directory green with the new asserts, 2026-10-12 ~13:35).

### Dedup/branching-flow sitting (2026-10-11 ratifications — execution tail)

**Status:** 🟡 `PARTIALLY DONE` (all five owner calls answered + recorded in registry/AGENTS; executions below) · **Priority:** Medium · **Effort:** M

1. **settingsRow ADOPTION** (owner: BUILD → resolved to ADOPT) — EXECUTED 2026-10-12: settings.templ renders `display.DefinitionList` (Class `wp-settings`), tw.css `@source` rebuilt, old-vs-new render diff 6/7 panels byte-identical (radius literals synced to the redesign train's 9/12px), registry ruling #2 annotated EXECUTED. templ-components ALREADY ships the component, so no new one anywhere.
2. **branching-flow upstream filings** (owner: file): (a) phantom-analysis class scoping/config — 13/14 rows this session were non-identity plumbing, per-finding nolint + baseline exist but no class-level knob; (b) panic-analysis FP precision — 3 guard shapes (closure nil-guard, map-to-slice invariant, i<j two-pointer) produced 6/8 false rows. Drafts land in `docs/drafts/`, evidence: docs/status/2026-10-11_02-19 triage doc.
3. **Re-triage the branching-flow skip** — EXECUTED 2026-10-11 (stale premise retired): the tool runs UNSKIPPED (19 baselined findings with in-source rationales; the SDK severity cap keeps new findings non-gating until the error_rules-bearing BuildFlow release); `skip_steps` now holds only pytest-test (verified 2026-10-25).
4. **art-dupl `-t 1` + baseline ritual** — EXECUTED 2026-10-12: first deep sweep 867→58 actionable groups, ZERO harmful; `.art-dupl-baseline.json` recorded (tool v0.7.2) and CI gained the new-only `dup-gate` job; sweep-log line in the registry. Unfixable detector FPs: none surfaced this sweep (upstream-filing duty stays with future sweeps).
5. **Daemon dep-lane move** (owner: MOVE): configure pma to stop bumping go.mod/go.sum (owner-side daemon config); bumps ride `buildflow update`/Dependabot with the tidy → re-pin → CI-watch tail (AGENTS Conventions row added).
6. **Quarterly baseline/suppressed-set audit** — first 2027-01-11 (registry Protocol step 1).

**Evidence:** registry 2026-10-11 entries + ratified-rulings section; status docs `2026-10-11_02-19` (triage) + `2026-10-11_02-48` (session report); AGENTS.md dedup/nolint/dep-lane/INTERIM edits

### brandid-lint enforcement posture (2026-10-16 scan train — unanswered policy call, harvested at the 2026-11-08 archive)

**Status:** 🟡 `OPEN DECISION` (owner call: fleet-policy precedent) · **Priority:** Low · **Effort:** S (decide) / M (a BuildFlow step)

The BD003 directives (87 sites at HEAD, grep 2026-11-08) are policed only when someone hand-runs `brandid-lint -rules all` — last clean run 2026-10-25 (AGENTS strong-id bullet). The 2026-10-16 train's proposal (BuildFlow step / CI job / flake version pin) was never answered; the report's g2 asked exactly this and waits. Decide: gate it (every train keeps it green) or ratify ad-hoc (directives rot only if an edit touches a directive site without a re-run; AGENTS's strong-id bullet is the convention home).

### ASR/SUPERB train follow-ups (2026-10-08 harvest: 16-54/18-15 §f + plan remainder)

**Status:** 🟡 `PARTIALLY DONE — hardening + polish + test debt CLOSED 2026-10-18; #13 affordance + #21 whisper.cpp variant SHIPPED 2026-10-25; remains: #48 privacy call, owner gates` · **Priority:** Medium · **Effort:** S-M

SHIPPED by the train (see CHANGELOG `[Unreleased]`): transcript persistence with per-extension + per-call caps and the near-silence RMS guard (M12), History transcript search + hint (M15), copy-to-clipboard on card/History/chips (M16), export-zip transcripts leg (M17), Settings provider kind·model row (M18), queued-hold + concurrency + cap-speedup specs (M19/M20), single-pass `Recent` (M21), confidence marker REJECTED on provider evidence (M22, decisions doc), ui-capture ASR-on surfaces (M13, 18-shot matrix), buildflow-at-train-start lesson + AGENTS rule + `.buildflow.yml` load-bearing warning (M14). REMAINING, grouped:

- **OWNER-gated (14-08 report §g — do not self-answer):** M1 first-live provider + retention/autoplay ratification; M2 Speaches deployment on the consuming stack (stack-repo work); M4 live-reality transcription (throwaway GCP key or Speaches instance); M6 stack browser E2E re-run (served markup changed twice on 10-08); 429 cooldown Retry-After vs fixed 10 s; ship `SILENCE_RMS` 0.004 as-is.
- **Provider hardening (16-54 §f):** CLOSED 2026-10-18 — retry/backoff + "transcription unavailable" toast (#43/#44) VERIFIED ALREADY-SHIPPED; provider latency metric (#23) landed as the rolling outcomes tracker + debug line; optional non-critical ASR health leg (#22) landed (recent-outcomes verdict, never probes); memory ceiling (#47) landed as the /api/transcribe concurrency ceiling (4 in flight, 503 past a short queue wait). #21 `/inference` whisper.cpp native variant SHIPPED 2026-10-25 (`asr.provider = "whispercpp"`: same multipart + {"text"} wire, /inference path, kind-aware Settings row).
- **Client polish (16-54 §f):** CLOSED except #48 — language hint (#11/#12), DOM cap (#41), per-call clear affordance (#39), transcript download (#40) landed (round-5 E16); "saved to History" discoverability (18-15 #13) SHIPPED 2026-10-25: the first server-confirmed segment save reveals a persistent saved-count line on the call card (en/de), announced once per call; REMAINS: remote-leg-only transcription default — privacy call (#48, owner).
- **Test debt (16-54 §f):** ALL CLOSED — 2026-10-12 batch (#32/#33/#36/#38/#34) plus 2026-10-18: voicemail panels transcribe-button test (#37) VERIFIED PRE-SHIPPED · 429 combined live+auto load (18-15 #5) pinned by the flood spec (transcribe + transcripts share one bucket).
- **Docs/smoke:** CLOSED 2026-10-12: NixOS module README `asr.*` options (#26 — already shipped by the google-wire train; all 7 rows verified against `config.ASR` fields) · error-contract ASR freshness (VERIFIED: all 8 `asr.*` codes in code match the table; `TestErrorCodeRegistryIsFresh` enforces it continuously) · smoke ASR leg with a fake provider (already shipped as `asr_scenario` in webphone-smoke.py — openai + google wires, wired into main) · "audio leaves the box" flag note (#49, README § Live transcription) · near-live segmented-design note (#50, same section).

**Evidence:** plan `docs/planning/2026-10-08_07-36_SUPERB-asr-provider-rollout.md`; decisions `docs/planning/2026-10-08_14-05_SUPERB-asr-decisions-and-spikes.md`; pause report `docs/status/2026-10-08_14-08_SUPERB-asr-train-m12-m25-execution-report.md` §g; source §f lists in `docs/status/2026-10-07_16-54_*` and `2026-10-07_18-15_*`

### Passkey tail: deploy proof (owner terminal)

**Status:** IN_PROGRESS (code + the 3-repo deploy train DONE 2026-10-04; owner legs remain) · **Priority:** High · **Effort:** S

Remaining: the OWNER switch + post-switch ritual (rotate `/tmp/pbx-toplevel-current`, record the fresh diff-closures baseline) and the LIVE passkey proof with the fail-closed password-file drill (pbx-artmann TODO §1). OWNER CALLS outstanding: ratify runbook-only enroll + Lars-only v1 mapping; the enroll-begin ULID-secrecy verdict (bind the ceremony to a token-verified server-side session — a usermgmt upstream change — or ratify ULID unguessability as the standing posture; audit 2026-10-06); switch-now-vs-CI-first (VERDICT 2026-10-05: every recent stack `main` CI run cancels at the 1h ceiling inside `nix flake check` — the mod_enum/timeout diagnosis precedes any lock bump); installer release republish timing.

**Evidence:** status `docs/status/2026-10-04_17-28_passkey-train-resumed-session-brutal-status.md`; plan `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`; the 2026-10-06 audits (HTTP surface, slog-leak sweep, island audit) live in that report + CHANGELOG.

### Boot-contract tail: D3 retry-loop owner call + stack runbook patch

**Status:** 🟡 `PLANNED` (owner call + stack dispatch) · **Priority:** Medium · **Effort:** S

The boot-error contract SHIPPED 2026-10-02 (5-part render on every boot failure, exit 1 designed / 2 panic, EN-only, class-tagged; `App.Start` de-panicked; arch test confines samber/do; smoke `boot failure scenario` green). The stack runbook patch APPLIED 2026-10-04 (stack `9a21893`: boot-surface + passkey error-contract rows; the parked patch doc is spent). Remaining: (1) OWNER CALL (D3 follow-up) — cap systemd's boot-failure retry loop (`StartLimitBurst`/`StartLimitIntervalSec`) or ratify the 5s `Restart=on-failure` retry as desirable liveness (module deliberately unchanged this train). The boot-surface re-grade landed EARLY 2026-10-05 (contract table vs code: 7 classes + drift pins intact; see the erraudit line in Standing watches)

**Evidence:** plan `docs/planning/2026-10-02_10-23_SUPERB-operator-boot-contract.md`; ruling `docs/error-contract.md` § "Boot surface"; pins `cmd/webphone/bootreport_test.go`; runbook patch = stack `nix-international-telephony@9a21893`

### Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error)

**Status:** 🔴 `TODO` (owner journal leg) · **Priority:** High · **Effort:** S

Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error) when probed 2026-09-19 — webphone-side classification is 502 + provider-rejection detail since `1d53f44`, refusals answer 422 since `6ac8962` (`1d53f44` live on prod since the 2.6.0 deploy; `6ac8962` rides the untagged release), but the ROOT cause is stack-side (telnyx-webhooks bridge down/rejecting or Telnyx creds). OWNER (assistant ssh is forbidden by pbx-artmann AGENTS): journal the telnyx-webhooks unit, grep for sms/422/error lines, restart or fix creds per findings, send a test SMS; record the root cause here + in the stack runbook. Also still unproven live: the 2.6.0 post-deploy sign-in rejection-banner (self-send) browser check — narrows to a journal-only confirmation if a human test SMS worked since 2026-09-25

**Evidence:** 2026-09-19 probe; webphone-side fix `1d53f44`; refusal-422 `6ac8962`; OWNER TRIAGE PACK ready 2026-10-01: command sheet §4 decision tree (unit status → journal grep → the bridge's own actionable cred strings → restart → test SMS; self-send 422 = expected, not a failure)

### Scheduled sends (M21.6): NO-GO on the current gateway seam

**Status:** ⛔ `DECIDED-AGAINST` (gateway-seam dependency; revisit only with a deferred-send seam) · **Priority:** Low · **Effort:** M

Deferred/scheduled message sending was cut from T18 (M21.6): the gateway seam has NO deferred-send concept (loopback sends immediately; the webhook bridge is fire-and-forget to the provider), so a scheduled-send UI would show a state the transport cannot honor — violating the island-honesty contract. Revisit only if the provider/stack grows a deferred-send lane; the snippets seam (M21) already covers the "reply faster" problem it was meant to solve.

**Evidence:** design note `docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md` (M21.6 cut); owner ratification pending (2026-10-02 §g) — answer surface: the sitting pack §F-side wording. **Audited 2026-11-08** (nothing lost; the NO-GO premise re-verified at HEAD: [docs/planning/2026-11-08_parked-decisions-audit.md](docs/planning/2026-11-08_parked-decisions-audit.md)).

### OWNER-calls batch session

**Status:** 🟡 `PLANNED` (sitting pack ready, sitting owed) · **Priority:** High · **Effort:** S

**THE ONE ANSWER SURFACE (2026-11-08): [docs/planning/2026-11-07_21-30_owner-sitting-pack.md](docs/planning/2026-11-07_21-30_owner-sitting-pack.md)** — 58 rulings (A product/UX, B privacy/data, C infra, D tooling, E cross-repo) + the F bulk table that folds the briefing's 38 rows, each with a recommendation, the other branch's cost, and a payoff map. Every surface below (the 2026-10-29 packet, the briefing, the ROADMAP open questions) is consolidated INTO it; answer THERE and this row closes.

Superseded surfaces kept for evidence only: the 2026-10-29 packet (six cards, folded into pack rows B1/B2/B3/C1/C2/D17), the briefing (38 rows, folded into pack §F), the ROADMAP §Open questions (prune in the same sitting per the pack's already-resolved section).

OWNER-calls batch session (decisions, one sitting — briefing at `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`): the original ~15 + CRM policy trio + self-send train-C semantics + templ-components history-blemish + tail items g1 force-push ratification / release.sh load-gate default / g2 KVM-timeout / art-dupl `-t 3` baseline (FIFTH surfacing) + registry ratification (`docs/dedup-registry.md` as THE home) + `settingsRow` build-or-retire (FOURTH surfacing) + suppression-bucket doc + dedup suppression-SCOPE ruling (shown-only vs periodic suppressed-set audit; 12:59 report g3, unrouted until v6) + webhook 400 body-text dependents + `msg/`→`message/` idem-key rename + helper micro-test bar + missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy. NEW since 2026-10-06 (briefing rows 36–38): the `auth:`-prefix × daemon-sweep interface (manual `auth:` commit before the sweep window vs accept a daemon `chore:` commit when the pinned session-behavior suite provably ran — 23:22 g2) · devShells.default tool additions (dprint/prettier/ruff/lychee, silences 4 buildflow warnings) + lychee authenticate-vs-exclude policy (22:08 f13/f14) · docs-only express-push ratification (23:22 g3). NEW since 2026-10-05 (briefing rows 29–32, addendum in the briefing doc): mixin/dupe detector-policy home (registry sweep-log only vs AGENTS promotion) · `config.CRM`/`Paperless` URL+Token twin standing ruling (resurfaced by TWO detectors) · `vcard.Card`/`SharedContact`/`apiSharedContact` triad boundary-layer rejection ratification · `errorfamily.Join` gap (add upstream constructor vs keep 4 nolint aggregate sites). NEW since 2026-10-02 (docs-health v6 sweep): ratify "routed-as-resolved → archive" for the four live owner-gated plans (error-excellence, 20-year-durability, 04-29 pareto, 13-33 stack-adoption — v4's g1); ratify train-level `v` verdict markers as the standing annotation hash bar (v3's g3, six sweeps running); the check-rows marker-cell baseline for the 37 pre-2026-09-18 archived files (restyle or accept). NEW since 2026-09-26: push-lag threshold policy (when is a silent daemon push stall BROKEN — 10 min? 1 h) + session push ratification (phase-boundary hand-pushes) · is the next webphone release 2.8.0 (making the bridge's ">= 2.8" claims true) or reword them · `gateway.attachment_limit` knob vs bridge-422-teaches design as final · sniff-fallback lifespan (keep forever vs delete-after-deploy-confirmed) · tier-2 family-adoption intent (dep swept-but-unused?) · report erraudit `tree` same-name-dedupe upstream? (owner's own tool) · AGENTS restructure permission · existing-prod-data chmod/re-backup for the UMask tightening · destDir nesting legality · stack verification now-vs-train-close

**Evidence:** briefing at 38 rows (15–28 added 2026-10-01; 29–32 on 10-05; 33–35 on 10-06 by the round-2 sessions; 36–38 on 10-06 by docs-health v7) + a closed-since section (go-cqrs-lite wontfix etc.)

### Post the release announcements

**Status:** 🟡 `PLANNED` (owner posts) · **Priority:** Low · **Effort:** S · **Gate:** the next release train (v2.9.x+) is the trigger — drafts refresh to cover it before posting

Post the release announcements (drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 at `docs/announcements/`; owner picks channel(s), approves wording, decides the disclosure posture — fix-acknowledged, no exploit detail)

**Evidence:** v2.8.0 drafts A/B/C = `docs/announcements/2026-09-30_v2-8-0_drafts.md` (covers the merged 2.7.0+2.8.0 story); v2.7.0 drafts annotated superseded, links repointed; gh object for v2.8.0 verified published 2026-10-01

### Standing watches

**Status:** 🟢 `ONE CALENDAR` (refreshed 2026-11-08) · **Priority:** Low · **Effort:** S

Standing watches live in ONE dated calendar now: [docs/watches/2026-11-08_standing-watch-calendar.md](docs/watches/2026-11-08_standing-watch-calendar.md) — every row carries cadence, next due, the named trigger, the evidence home, and the on-fire action. The refresh re-MEASURED the lapsed erraudit monthly check same-session (2026-11-08: `check --max --no-suppress` = 0 violations, `nolint-audit` 68/68 directives needed, 0 stale) instead of silently moving the date; next erraudit due **2026-12-08**, quarterly buildflow **2026-12-20**, credentials **2026-12-30**, modularize **~2027-01**, dedup **2027-01-11**. The Fleet-libs reopen-trigger registry (T25) stays the ONE home for ruling reopen triggers — the calendar references it, never duplicates it.

**Evidence:** the calendar file itself (every date derived from a verified prior entry); prior prose list re-checked 2026-09-23/24 during the pareto cycle (C22/M22)

### internal/server auth-package carve (dormant trigger)

The 2026-10-17 hooks/api carve is DONE and phase-3 adjudicated (session_api
STAYS — it is the shared auth spine of both login modes; passkey_api reuses
its three shared methods, and the handlers ride wiring-private state). The
honest future move — BOTH login modes into an auth package — reopens on the
next auth-surface change; the blueprint and dependency map live in
`docs/modularization/2026-10-17_PROPOSAL.html`, the full adjudication in git
history (the 2026-10-12 + 2026-10-17 entries).

### Mic pre-warm live verification + gates

**Status:** 🟡 `PARTIALLY DONE` (code + gates green; live ritual owner) · **Priority:** Medium · **Effort:** S

`mic.js` pre-warm (speak ASAP after accept) shipped with T07 (2026-10-01, code + island pins) and the T11 perf train added the mic-indicator timing baseline; suite/smoke green. Remaining: ONE live-call ritual on the stack (accept→speak sub-second, indicator at ring, warm release on reject/missed) — owner terminal.

**Evidence:** `docs/status/archived/2026-10-01_04-07_mic-prewarm-accept-latency-train-status.md` §b/§f; T11 timing baseline `scripts/perf-baseline.py`.

### Visual harness: shot disposition + optional vision cross-check

**Status:** 🟡 `PARTIALLY DONE` (harness shipped; disposition is owner) · **Priority:** Low · **Effort:** S

`scripts/ui-capture.py` is the persisted harness (T23): config-file boot with the fronted CSRF shape, HTTP seeding (multipart sends incl. an audio attachment, meta-CSRF → `/api/session` → rotated token, transcript seed via `/api/transcripts`), session-cookie INJECTION into chromium (the island's login gates on the SIP WS a bare boot cannot serve), per-surface DOM assertions (marker + optional rendered-copy pin) before every shot, and the ASR-ON 20-shot matrix (10 surfaces × light/dark — M13 extension 2026-10-08, styled 404 added 2026-10-23) into `ui-shots/`. LOCAL-ONLY by budget decision — see the AGENTS command entry. Remaining (owner §g1): eyeball the current matrix; decide per-release vs per-train persistence (default: per release); the optional vision-CLI cross-check needs a provider/key decision.

**Evidence:** T23 in `docs/status/archived/2026-10-02_11-41_t18-t23-train-session7-status.md` §a/§b; run recipe in AGENTS § Commands.

### Cross-repo obligations (stack / pbx-artmann)

**Status:** MOSTLY EXECUTED (2026-11-08 legs train) · **Priority:** Medium — THE ONE HOME for the standing stack-E2E obligation. The four standing legs have a verified step-by-step runbook: [docs/planning/2026-11-08_cross-repo-handover-pack.md](docs/planning/2026-11-08_cross-repo-handover-pack.md) (stack `crm.caFile` pass-through · `/hooks/ring` dialplan leg · `mod_enum` CI-ceiling repair · the #9 owner sequence). **Per-leg verdicts (2026-11-08):** Leg 1 `crm.caFile` — RESOLVED (already landed in the stack: options + threading + eval render pin); Leg 2 `/hooks/ring` — SHIPPED as-built (loopback shim `ringhook.py` + `services.telephony.ringhook.*` unit + conditional dialplan arms, off = byte-identical; `ringhookCheck` eval arm + 7-spec unittest; the RING-GROUP path uncovered — owner question §g-1); Leg 3 `mod_enum` — EXCLUDED with the ruling comment (zero call sites grep-verified), telephony-eval green, pbx toplevel builds; run 42529951959 finished the flake leg in 46m40s (under the old ceiling) with every suite green EXCEPT the sibling session-2's own named-open `telephony-voicemail` quota red — the twice-green acceptance is blocked on THAT train, not on these legs; Leg 4 — owner-gated (unchanged). Remaining open: the voicemail quota red (sibling's), the ring-group boundary ruling, and the residues below.

Owner/stack legs that block or derive from webphone work: repair the stack
FreeSWITCH `mod_enum` build (it produces the stack CI 1h ceiling that cancels every
`nix flake check` run and gates each relock) → relock to webphone main +
pbx-artmann re-pin. The standing stack browser E2E obligation was DISCHARGED
2026-10-29 16:05 (green with the full owed markup set in the tree, VM test green);
a FUTURE served-markup change re-owes it per the release-runbook standing rule
(M03/F16). Also: stack `services.webphone.paperless` module option + smoke arm;
the WebTransport-not-adopted verdict doc; telephony `deploy.md` secret PATH column;
ops-runbook demo-call recipe (`originate user/1000 &playback(local_stream://moh)` +
the `/var/lib/telephony-secrets/` password path); MOH audibility + `/recordings/` +
CDR check; the gateway stack-side bits (`ftypqt`→`video/quicktime` sniff, E2E
MMS-outbound, pbx-artmann FEATURES:87 stale text); the 2026-10-23 404-health
stack residues (a dashboard-on VM test leg — boots the stack with
`webphone.dashboard.enable = true` and asserts the page renders through the
vhost fence — and the app's styled-404 shape rule in `tests/browser-e2e.py`,
grep-verified absent 2026-11-08). No assistant ssh — verified
handovers only. Evidence: owner command sheet
`docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md`.

### Round-5 plan tail (2026-10-17 harvest — d2 + markdownlint posture remain)

2026-10-18 closed F75/F76/F79–F84; 2026-10-25 row-hygiene sweep closed F74/F85/F86
(evidence re-verified at HEAD; the closed-work paragraphs were deleted with the
2026-11-05 sweep — history lives in the plan + its status reports).
REMAINS: d2 cheat-sheet upstream fix (`style.stroke-dash` on v0.9.0; blocked on the
sitting's Q3) + the human eyeball of the 2026-10-17 D2 pair (image-blind assistant
ceiling) + markdownlint posture execution (post-sitting row-18 verdict). Plan:
[docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md](docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md).

### Native-app train: ring-while-closed legs (2026-10-23 ship, webphone side closed)

**Status:** 🟡 `P2 SHIPPED server-side; P1 shell scaffolded (webphone-android v0.1, BUILD GREEN + 16KB-aligned); the two remaining legs live OUTSIDE this repo` · **Priority:** High · **Effort:** M (each leg) · **Preflight 2026-11-08:** the android repo is QUIET (clean tree, last commit 11-05, 2 commits/7d — half the KMP gate met, so the spike STAYS deferred; the go/no-go note + the first-device-train checklist: [docs/planning/2026-11-08_native-app-preflight.md](docs/planning/2026-11-08_native-app-preflight.md))

- **Stack leg (nix-international-telephony) — LANDED 2026-11-08:** the FS dialplan now POSTs the ring facts (`{owner, from, call_id, state: ring|end}`) to the `ringhook.py` loopback shim, which Bearer-signs the forward to `webphone:/hooks/ring` (`services.telephony.ringhook.*`; dialplan arms only when enabled — off = byte-identical; eval arm + 7-spec unittest). REMAINS: the end-to-end ring proof (needs a push device registered + the shim enabled in a VM or the live host) and the ring-group-path ruling (§g-1). The seam is no longer inert stack-side; it is OFF by default until an deployment sets `ringhook.enable`.
- **Shell leg (webphone-android):** first device train — exercise the SIP UA against the live WSS gateway (REGISTER/INVITE/BYE/DTMF/hold + incoming INVITE + REFER/NOTIFY are implemented but UNTESTED against FreeSWITCH), compose the real CallStyle notification (v0.1 posts the FSI heads-up), the origin-picker settings UI (App.origin is a constant today), and the `/api/session` field-name check against the served sessionIdentity shape.

**Evidence:** decision doc `docs/planning/2026-10-23_16-13_native-app-platform-decision.md` (status: DECIDED + SHIPPING); `~/projects/webphone-android` README honesty markers; `internal/push` + `internal/server/hooks/ring_test.go` + `internal/server/push_api_test.go` for the shipped server contract.
