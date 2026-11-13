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

### UI/UX destination trains (ratified 2026-11-11 — plan: docs/planning/2026-11-11_15-11_ui-ux-destination.md)

- ~~**Train 3 tap-through proof** (§f-13 of the 04-09 report): a driver script taps
  Phone → island full-screen → a real tab returns, capturing the three states —
  the tab's LOOK is captured (bar fits 8 slots after the `min-width: 0` fix,
  probe-verified `scrollW 358 ≤ nav 360` at 375px) but the FLOW is unproven.~~ →
  done 2026-11-12 — `ui-capture.py --tap-through` (390x844, asserts all 3 states in DOM, aborts non-zero) ran GREEN 3/3; and it caught a REAL shipped defect: the phone-focus rule hid `.wp-main-col` (nav inside it — the only return affordance), fixed to hide `#tab-content` + `#wp-tab-error` only, re-proven green (app.css `@media (max-width:900px)` block)
- ~~**Stack E2E strip assumption** (§f-14): the strip gained a JS-created button child;
  `refreshNav`'s `[data-tab]` walk is safe by construction — confirm the stack's
  `tests/browser-e2e.py` never assumes a strip child count.~~ →
  done 2026-11-12 — verified: the stack E2E touches the nav ONLY via `[data-tab='{tab}']` exact selectors (line 680), no child-count walks; the JS-created Phone tab carries NO data-tab attribute — safe by construction
- ~~**German stacked-table lane** (§f-12): the mobile History stack renders
  `td[data-label]` copy from markup — verify the longest DE labels ("Zeit"/"Dauer")
  fit their cells at 390px (the lane was verified EN-only).~~ →
  done 2026-11-12 — measured GREEN at a TRUE 390px viewport (WebDriver `set_window_size`; the `--window-size` flag silently clamps at 500): a fake PBX (`/phone-api/history` + the voicemail-summary credential verify) served 3 CDR rows and every Zeit/Dauer stacked cell fits (scrollW == clientW, no lane-wrapper overflow); gotcha recorded — a harness `--window-size 390x844` boot measures 500px wide unless the WebDriver API resizes
- ~~**Phone label i18n** (§f-15): the label is English-only by the D3 shell rule while
  the localized tabs sit beside it — owner eyes or an owner exception.~~ →
  done 2026-11-12 — batched as Q1 of docs/planning/2026-11-12_07-05_owner-calls-batch.md (default stands: D3 keeps the label English)
- ~~**Destination-doc wording** (§f-29): Train 3's acceptance line says
  "ring-while-closed parity"; the shipped tab is navigation focus, not push —
  align the plan wording with what shipped.~~ →
  done 2026-11-12 — the acceptance line now reads navigation-focus parity (wording-only diff in docs/planning/2026-11-11_15-11_ui-ux-destination.md)

### UI token/craft + Train 2/3 residue (2026-11-12, harvested from the 04-09 report §f)

- ~~**Hover matrix per surface** (§f-18): `--hover SURFACE=SELECTOR` pairs so the
  hover lane can vary per shot instead of one global selector.~~ →
  done 2026-11-12 — `--hover-map` shipped (repeatable SURFACE=SELECTOR, last-wins parse) and proven live: messages/history/settings × light/dark at 390x844 = exactly 6 shots
- ~~**Fixed-element caveat print** (§f-19): full-page CDP captures render
  `position: fixed` at the original viewport's bottom — one harness print
  prevents the next "the bar is mid-image" misread (cost a §d scare once).~~ →
  done 2026-11-12 — main() prints the caveat once on every full-page run (ruff + self-test green)
- ~~**helpers.mjs registry-root doc** (§f-20): the stub's querySelector walks
  registry-root SUBTREES only (a root can never match itself; `body` is not a
  root) — write it where the next spec author will read it (cost a debug session).~~ →
  done 2026-11-12 — doc line landed at the document.querySelector stub in island-tests/helpers.mjs
- ~~**`.wp-day-head` dual context** (§f-21): one class, two display semantics
  (flex in transcripts, table-cell in the CDR table) — split the classes at the
  next touch of either.~~ →
  done 2026-11-12 — split landed: CDR rows now `wp-table-day-head` (app.css counter-rule block + 2 container-query sites, history.templ + regenerated `_templ.go`, 3 history_test.go pins renamed); transcripts keep `.wp-day-head`; both suites green
- ~~**Table region focus ring** (§f-22): the scroll wrapper's `tabindex="0"` needs a
  focus-visible check INSIDE the container — sticky heads clip the ring.~~ →
  done 2026-11-12 — `.wp-table-wrap:focus-visible` ring added (outline-offset −2px)
- ~~**Empty-state action buttons** (§f-23, owner eyes): rows point at the composer
  ("add one above") but carry no button — the library supports ActionText;
  ruled out as a duplicate affordance, owner may disagree.~~ →
  done 2026-11-12 — batched as Q2 of docs/planning/2026-11-12_07-05_owner-calls-batch.md (the decline stands by default)
- ~~**Island state-layer mirror** (§f-25): `island/style.css` took the token
  additions but not the hover state-layer sweep — deliberate (island buttons
  use :active scale) or a gap? Decide, write the ruling.~~ →
  done 2026-11-12 — ruled DELIBERATE: island buttons own color-shift hovers (accent-strong/danger-strong/surface-2) + press-scale (reduced-motion-neutralized); the shell's 6% overlay sweep is a dense-row idiom, not an island obligation
- ~~**art-dupl sweep** (§f-30): the table stack lane / reports grid / contacts grid
  share grid idioms — adjudicate fresh clones against the pinned baseline.~~ →
  done 2026-11-12 — pinned v0.7.2 `-t 1` sweep: the three grid lanes carry ZERO clone groups (CSS/templ idiom, not Go duplication); baseline 110 holds; the run's one new group is another session's in-flight VoicemailSummary twin (fix-forward; sweep-log row in docs/dedup-registry.md)
- ~~**whitespace-drift CSS-indent verification** (§f-31): confirm the formatter
  adjudication would catch a 1-space-indent CSS commit (the §d-3 incident class).~~ →
  done 2026-11-12 — planted probe: treefmt re-indents the 1-space CSS ("1 changed") and `--ci` mode exits 1 on the non-canonical bytes; canonical file passes
- ~~**`wp-phone-focus` seeded capture** (§f-32): a 390-wide pass with the focus class
  pre-seeded via script — the dialer-focus LAYOUT is captureable without taps.~~ →
  done 2026-11-12 — `--seed-class html=wp-phone-focus` run captured the dialer-focus layout (2 shots, 390x844) without taps
- ~~**Sticky-head stacking** (§f-35): table thead z-index (1, inside the wrapper)
  vs palette (200) and toasts (100) — confirm no bleed when a toast crosses a
  scrolled table.~~ →
  done 2026-11-12 — census + ladder commented in app.css above `.wp-table thead th` (thead 1 < header 10 < snippet picker 30 < mobile island 50 < mobile nav 90 < skip links + #toasts 100 < palette 200); `.wp-table-wrap` is overflow-only (NOT a stacking context), so no overlay bleeds under the sticky head

### 2026-11-12/13 file-size split train residue (harvested from the 22-50 report §f)

The five-file split train closed 2026-11-12; the sixth file
(`internal/matrix/receiver.go`, #248) was split 2026-11-13 (`sync.go`
wire-surface carve, byte-pure motion, package + arch suites green) and
`buildflow -s file-size-check` reads exit 0. This section holds the
report's still-open §f rows.

- [x] **Raceprobe fix decision** (#249, High): ~~await-poll vs
      product-semantics ruling (owner question §g3 of the 22-50
      report)~~ → DECIDED + SHIPPED 2026-11-13 (pareto T14): await-poll,
      not semantics — the probe's exact-count `svc.Threads` reads raced
      the async projection fold; `awaitThreadFold` polls to convergence
      (5s deadline, the `awaitTranscriptCount` pattern), both reads
      route through it (`c46b9666`). Pre-fix reproduction at current
      HEAD: NONE (the ladder retirement shifted fold timing), post-fix
      gauntlet green; #249 CLOSED with the evidence comment. The
      semantics question stays closed with it — a refused message never
      folds; only the PROBE's read timing was wrong.
- [x] **Branching-flow baseline refresh** (Medium): ~~at the next
      governance sweep~~ → done 2026-11-13 (pareto T23, pulled forward):
      `buildflow -s branching-flow --save-baseline` re-run from this
      checkout with the result cache disabled — 220 findings
      re-recorded at current paths (COMPOSITION_mixin re-homed, the
      stale old-path DUPLICATE_TYPE entries gone), all under the
      severity cap (non-gating per the config).
- [x] **ci-flake-ledger row** (Low, conditional): → resolved 2026-11-13
      — the condition never fired (#249 closed with the await-poll fix
      before CI ever red; local-repetition class only, `c46b9666`
      carries the gauntlet). No ledger row; reopen only if a repetition
      red ever lands on CI.
- ~~**AGENTS.md one-liner** (Low): the error-registry regen fires on
  MOVE, not only on NEW codes (proved 2026-11-12 on the asr split);
  add it at the next legitimate AGENTS touch (cap has room).~~ →
  done 2026-11-13 — already in: AGENTS.md carries the full fact
  (MOVES/RENAMES regen + the `-update` command + the asr-split
  proof + the same-push reading for two live sessions); nothing
  left to add.
- §D10 (split-on-sight vs raise-cap vs warn-only) stays with its
  existing 2026-11-08 pack above — no duplicate row. The corpus cadence
  drain beyond the one CI-greening archive stays with the T24
  release-tails train.

### 2026-11-02 cue-assessment residue (harvested 2026-11-13 — the report's own §f-5 ask)

- [ ] **Stack drift audit**: read the stack repo's webphone module
      assertions and diff every rule against `internal/config` — the
      CUE-era verification task, still valid under the typed NixOS
      lane (rides the next stack module touch).
- [ ] **Record the CUE ruling durably** (REJECTED; schema-export
      deferred with a named trigger) as an owner decision line, and
      park the cuelang.org/go cost measurement next to it.
- [ ] **Schema-export spike**: settings schema GENERATED from the
      config structs (the openapi.json doctrine applied to settings).

### OIDC lane residuals (routed 2026-11-13 from the archived 2026-10-25 15-16 report)

- [ ] **NixOS VM test for the OIDC lane** (`webphone-oauth2`-style:
      systemd `LoadCredential` lands, mode 0400, the app actually
      reads the credential at login) — the eval-time assertions +
      `module-check-oauth2` cases are the current fence.
- [ ] **Second module-check golden with `oauth2.enable`** (exact
      rendered JSON, mirroring the zero-config golden).
- [ ] **Typed passkey options** in the module (passkey-only configs
      get derived rp the same way the OIDC lane derives them).

### Voicemail playback train residue (2026-11-10 — the fix is pushed; production legs open)

The two product defects are fixed and CI-green (`fc352e66`: honest toast +
Range forwarding; `shell.js` one-honest-toast split-brain repair). The LIVE
root cause of the original "play does nothing" report is still unconfirmed —
no credentials on the deployment. Discriminating evidence: the next failure
now names itself via the toast (silence = fixed; text = live trigger named).

- [ ] **Stack redeploy** riding webphone `main` so the fixes reach
      `pbx.artmann.tech`; then click play and RECORD the verdict (the new
      toast text is the live trigger's name).
- [ ] **Host-side diagnostics** (needs deployment access):
      `journalctl -u telephony-operator | grep -E "send_file|audio not found"`
      + `systemctl status telephony-fs-state-acl` — the runbook's audio-404
      class (`docs/ops-runbook.md` § audio).
- [ ] **Stack coverage arms**: browser E2E voicemail playback (zero
      voicemail surface in `tests/browser-e2e.py` today);
      `verify-live.sh` authenticated audio fetch; `operator.nix` pin for
      206-through-the-proxy.
- ~~**Local freshness replay** before push (rebuild `tw.css` + `git diff`
  inside the battery)~~ → done 2026-11-11 — `scripts/build-tw-css.sh --check`: the same staged pipeline, then byte-compare against the committed artifacts, restore + exit 1 on drift (verified in-sync at HEAD; the sibling's docs-gates `--check` CI leg covers the docs half)

### Reports tab quality train (2026-11-10, owner triage of the live panel — plan: docs/planning/2026-11-10_10-15_SUPERB-reports-tab-quality-plan.html)

DONE 2026-11-10 (trains cabdeeb8 / abed8247 / 3d1c1ba2 / 12cee0b3): C1 is the one-line root fix
(73/73 CDR rows undated because FreeSWITCH writes `start_stamp` with seconds,
`%Y-%m-%d %T`, and `reports.startLayout` accepts none):

- [x] C1: `startLayout` "2006-01-02 15:04" → "2006-01-02 15:04:05" +
      real-format fixtures in `reports_test.go` (restores This-week, By-day,
      Hours — the 1%). Minute-only timestamps now pin as Undated
      (counted, never dropped).
- [x] C2: honest unknown rendering — all-undated ⇒ EmptyState instead of
      "This week — 0"; guard the By-day heading; move the undated caveat next
      to the temporal block; new i18n keys in BOTH maps.
- [x] C3: resolve top-contact names through the contacts store first, CRM
      fallback, at the shared `crmDirectory` seam (extended into
      `directoryFor` — the SSE push lane merges through the SAME home, so a
      morph can never rename a row a full render named differently);
      muted mono for non-dialable remotes ("pifo42h1" class).
- [x] C4: `Totals.NoAnswer` (unanswered outbound) as a seventh stat card when
      > 0; `.wp-reports-totals` checked — auto-fit takes any card count,
      no CSS change. Card renders NEUTRAL (deviation from the plan's
      "alert tone": the alert red is the inbound-miss tone by contract).
- [x] C5: pins + gates + live verification (reports/view suites, nix fmt,
      buildflow, CHANGELOG line, prod ui-capture of the reports surface).

### Prod screenshot-tracking train residue (2026-11-10 14:42 report, §f-24 — harvested by the cadence-sweep session)

- [ ] **T05 — the authenticated prod capture** (owner-gated: needs
      `WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD` env
      creds, asked twice). The command is staged:
      `nix run .#ui-track -- --base-url https://pbx.artmann.tech --track`.
      Unblocks in order: owner gallery walkthrough → ranked ugliness
      inventory (T05 act 2) → first fix with before/after proof (T06) →
      second run / first real prod compare chain (T07) → calibration
      values (T09/T10) → mobile pass (T08). Also rides that session:
      the real-display `--open` pass and the ring-guard positive path
      (row below).

- [x] **ui-capture `--open` flag**: launch the gallery/compare in a browser
      after a run — LANDED 2026-11-10 (harness v2; opens compare.html on
      a display, prints a headless note otherwise; not yet exercised live
      — first authenticated prod session should pass it once).

- [ ] **Night-run stability pair**: a quiet-window (overnight)
      ui-capture run pair on real prod data to set the diff noise floor
      (badge bar / threshold calibration values; the 0.5% default is
      optimistic — identical-seed local runs peaked at 2.5% on relative
      timestamps). Home: the tracking plan's T25; pair the runs ≥8h
      apart so relative times settle. SUPPORT PACK LANDED 2026-11-13
      (pareto T18/F61): `scripts/ui-night-pair.sh` schedules both runs
      as transient systemd user units (run 1 now, run 2 +8h) with the
      env credentials — the wall-clock half stays a scheduled item until
      the first pair actually runs.

- [x] **Self-test pins for the dormant paths** (tracking plan T28's
      second half + §f-3 partial): the EMPTY-run discard guard
      (empty crashed dir removed, shots-bearing run kept) and the
      `--open` decision (`open_target`: produced artifact wins,
      out-dir fallback, headless refusal) are pinned in
      `--self-test` (35 checks, both envs) — LANDED 2026-11-10.
      The real-display `--open` pass stays with the first prod session.

- [ ] **Ring-guard positive path**: exercise the
      `#incoming-call:not([hidden])` re-shoot during a REAL inbound
      call in a tracked run (first prod session + a caller; the
      selector is verified against the served markup 2026-11-10 —
      phone.templ renders `hidden`, the island toggles the attribute).
      Home: the tracking plan's T13 live half.

### BuildFlow tool-disposition residue (harvested 2026-11-08 from the 2026-10-11 findings-zeroing plan — one home now)

Six tool dispositions the 2026-10-11 plan owned but TODO_LIST never carried;
each verified still-open at harvest time (the plan doc keeps its detail):

- ~~go-auto-upgrade: add the `warnings_budget` marker + rationale comment to
  `.buildflow.yml` (lo/jsonv1 declined per doctrine; plan T6 µ20).~~ → done 2026-11-11 — VERIFIED HEALED, no marker owed: `buildflow -s go-auto-upgrade` runs clean at the current tip (the lo/jsonv1 warning premise the marker was drafted for no longer fires — upstream healed it), so a rationale comment would document a warning that does not exist
- ~~jscpd: attribute the 6 island-test clone rows in docs/dedup-registry.md
  (test-fixture class; plan T6 µ22).~~ → done 2026-11-11 — attributed: `buildflow -s jscpd` = 15 findings, ALL island-test bootstrap harnesses; standing-ruling row + sweep-log line landed in docs/dedup-registry.md (13:4x)
- ~~lychee: rerun `-v` and reconcile the 5-vs-0 delta; the standing
  authenticate-vs-exclude policy waits on sitting pack §F9.~~ → verification half done 2026-11-11 — rerun: 579 OK, 0 errors, 3 redirects (`nix run nixpkgs#lychee -- . --no-progress`, lychee.toml config). The reconcile found an UPSTREAM display gap: lychee 0.24.2 names redirect URLs in NO local mode (-v/-vv/--json/detailed carry only the count), so the 5-vs-0 delta cannot be decomposed locally — no URL guessing. The §F9 policy call stays owner-gated; the upstream gap question rides §g of the 14-02 report
- [x] ~~nix: add `meta.description` to the vulnix app (nix/apps.nix; plan T6 µ23).~~ → done — VERIFIED ALREADY-SHIPPED (meta.description present at both app sites in nix/apps.nix; landed with the sibling session's sweep)
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
- ~~**Delete the prod-dead readmodels faxes store** — the
  engine-backed `Faxes` in internal/readmodels/faxes.go lost its
  prod wiring at P7 (app.go wires only `cqrs.NewFaxes`); it
  survives as the fax service's test double
  (internal/fax/teststore_test.go, archiver_test.go). Move those
  tests to the cqrs backend, keep `FaxesProjection`/
  `FaxJobOfView`/`FaxGate` (faxesgate.go, the shared fold builder
  both DBs ride), then delete the store — dissolves 3 of the 6
  clone groups baselined 2026-11-09 (the registry's post-P7 fax
  twins row). **RIDER (2026-11-11 doctrine sweep): the same train
  deletes the prod-dead `readmodels.EnrollTokens`** (the
  composition root wires `CqrsService.EnrollSvc()`; the
  engine-backed store has zero non-test callers — same
  test-double pattern; readmodels/enroll.go + the engine field in
  readmodels.go).~~ → done 2026-11-11 — faxes.go DELETED (FaxView/faxInput moved into faxesgate.go; FaxesProjection/FaxJobOfView/FaxGate kept); enroll.go slimmed to EnrollGate + EnrollTokenView (the live cqrs seams); Engine lost the faxes/enroll fields + accessors + Plan entries. Tests moved: fax/teststore_test.go boots the journal backend (`cqrs.FaxesDomainDecl` + `NewFaxes`), userauth/session_switch/passkey_api tests boot `cqrs.NewEnrollSystem`+`NewEnrollTokens`; the dying readmodels coverage PORTED (ListAll newest-first + ref guards + the 11-row faxMoves mirror + idempotent-second-key, now cqrs tests); admin_reads fax pins moved to cqrs/faxes_test.go (transcripts pin stays). Registry regen'd (-14 codes); art-dupl scan: 0 clones reference readmodels/faxes.go (3 twin groups dissolved — registry row updated); full `go test -count=1 ./...` green.
- [x] ~~**Storage-ladder train A: delete the prod-dead readmodels
      transcripts store** (the SUPERB 80/20 plan's M02; the faxes+enroll
      recipe verbatim — prod has answered transcripts from the cqrs
      journal since A27; the engine-backed `Transcripts` had zero
      non-test callers)~~ → done 2026-11-11 — transcripts.go DELETED;
      the surviving surface moved to transcriptsgate.go (TranscriptGate,
      TranscriptSegmentView, both cap constants, ReassembleTranscripts —
      the gate+view home the doctrine keeps); Engine lost the
      calls field/accessor/Plan entry. The server test harness rides
      `cqrs.NewTranscripts(coreSys)` (NewCoreSystem already carried the
      decl; the gate's Append awaits its own projection fold, so the
      harness stays deterministic). The admin RecentAll pin + the
      window/caps table (2230-segment flood via direct projection
      folds), the owner-cap trim (5010→5000 through EnforceCaps), the
      metacharacter-literal search pin, and the 8×25 concurrent-append
      no-loss test all PORTED to cqrs/transcripts_test.go; the per-call
      floor stays decision-pinned (TestTranscriptPerCallFloorEvictsOldest).
      Registry regen'd (−8 `readmodels.transcript_*` codes); art-dupl:
      0 clones reference the deleted file.
- [x] ~~**Storage-ladder train B: delete the prod-dead readmodels
      notes + reviews stores** (the plan's M03; prod has answered both
      from the cqrs journal since Phase 4 — the engine-backed twins had
      zero non-test callers beyond the server harness)~~ → done
      2026-11-11 — callnotes.go/callreviews.go DELETED; the surviving
      surface moved to notesgate.go/reviewsgate.go (NotesGate/
      CallNoteView, ReviewsGate/CallReviewView, the 2000-cap constants
      and sentinels with their family classifications); Engine lost the
      notes/reviews fields/accessors/Plan entries. The server test
      harness rides `cqrs.NewNotes(coreSys)`/`cqrs.NewReviews(coreSys)`
      (NewCoreSystem grew the two decls per its one-journal charter).
      The two missing slot-shape pins ported (verdict-first slot
      creation, empty-body-clears-still-a-slot — cqrs
      TestNoteVerdictFirstAndClearedBody); the cap/merge/scoping/
      restart pins already lived at the cqrs decision. BOTH backends'
      write doors now carry the D7 awaitOwnWrite beat (the engine
      store's synchronous fold used to provide it; the History panel
      re-fetches right behind the SSE nudge — the aggregate header's
      old "no D7 await" premise retired with the twin). Registry regen'd
      (−7 note/review/disposition codes).
- [x] ~~**Storage-ladder train C: delete the prod-dead readmodels
      contacts + snippets + dialwindows stores** (the plan's M04; all
      three answered from the cqrs journal since the SUPERB swaps —
      the engine twins had zero non-test callers beyond the server
      harness and the fuzz/cap pins)~~ → done 2026-11-12 —
      contacts.go/snippets.go/dialwindows.go DELETED (plus the
      engine-store dial_bench and the now-userless notFound envelope +
      its test); the gate/view surface moved to contactsgate.go/
      snippetsgate.go/dialwindowsgate.go (gates, views, cap constants,
      sentinels); Engine lost the contacts/snippets/dials fields/
      accessors/Plan entries and the aggregate-reader assertion (its
      last consumer was the contacts metrics count). The server harness
      rides `cqrs.NewContacts`/`NewSnippets`/`NewDialWindows` over the
      core journal (NewCoreSystem grew the three decls); the fuzz spec,
      the API cap pin, and the store-failure lanes boot the journal
      through one newCoreSystem helper; the 500-contact API cap test
      seeds through the backend (dispatch+await, ~2s). The live-engine
      delete-miss envelope pins retired with the stores — the cqrs
      delete tests carry the contract. Registry regen'd (readmodels
      contact/snippet/dial codes out; the regen also caught the
      concurrent train's uncommitted codes). Engine = push devices +
      agent audit ONLY (M05 empties it).
- [x] ~~**Storage-ladder closeout (the plan's M05): delete the last
      two prod-dead readmodels stores (push devices + agent audit)**
      (both answered from the cqrs journal since the SUPERB Phase 3
      swaps; zero prod callers)~~ → done 2026-11-12 —
      push.go/agentaudit.go + their tests + export.go + readmodels.go
      (the Engine itself) DELETED; the gate/view surface moved to
      pushgate.go/agentauditgate.go (gates, views, cap constants) and
      doc.go now carries the doctrine-complete package header. A new
      arch test (`TestReadmodelsStaysGatesAndViewsOnly`) pins the
      purity: no sqliteengine import, no `func Open(` — the
      second storage brain can never grow back. The five consumer test
      sites (hooks ring + idempotency contract, auth-audit, push-api,
      summary-push) boot the journal backends through the core system;
      both write doors carry the D7 `awaitOwnWrite` (the engine folded
      synchronously; the ring prune flaked 1/5 without the await).
      LOAD-BEARING SIDE-FIND: readmodels.go silently carried the
      sqlite engine-driver registration for every boot —
      internal/cqrs is now the ONE home (blank import in
      threadsregister.go; without it every `system.New`, prod
      included, dies unknown-driver). NewCoreSystem grew the two decls;
      registry regen'd (10 `readmodels.push_*`/`agent_audit_*` codes
      out); art-dupl 0 new (baseline holds 117); full suite + race
      green.
- **Storage-ladder §5 DROP migration: owner date + precondition proof**
  (F35, the deepening plan's T11). The frozen families
  (threads/messages/attachments/fax_jobs/contacts/snippets/call_transcripts)
  are reader-less in production (the legacy readers are test-only
  since v2.12.0; the dead `*store.Messages` injector registration
  deleted 2026-11-12); the retention sweep empties them at
  `retention_days`. The precondition is provable per deployment with
  ONE sqlite3 query against `webphone.db`:
  `SELECT 'threads',COUNT(*) FROM threads UNION ALL SELECT
      'messages',COUNT(*) FROM messages UNION ALL SELECT 'attachments',
      COUNT(*) FROM attachments UNION ALL SELECT 'fax_jobs',COUNT(*) FROM
      fax_jobs UNION ALL SELECT 'contacts',COUNT(*) FROM contacts UNION
      ALL SELECT 'snippets',COUNT(*) FROM snippets UNION ALL SELECT
      'call_transcripts',COUNT(*) FROM call_transcripts;` — all zeros ⇒
  §5 is safe. **Owner proposal: schedule the DROP migration for the
  first 2.15.x maintenance window after 2027-01-01** (gives every
  deployment ≥30 days of default-retention sweep time past the
  2026-11/12 releases). app_flags is NOT inert — the matrix/signal
  sync cursors live there (doc corrected 2026-11-12); their move to
  system.db (or their stated home) is §5's own precondition, listed
  in the retirement doc now.
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
- ~~**cqrs-lint advisory triage over internal/cqrs** — 185
  warnings (C044 ×50, C006 ×20, D014 ×17, C023 ×14, C009 ×16 …;
  C007 time-in-decider is only ×3 — the closeout report's "180
  C007" miscount corrected), ALL severity-warning and non-gating
  under the SDK cap. Adjudicate the real ones (A032 branded-id
  suggestions collide with the JSON-wire raw-string ruling), file
  the noise classes upstream, or baseline with rationale once the
  error_rules-bearing BuildFlow release lands.~~ → done 2026-11-11 — the premise DECAYED: the pinned detector v4.21.0 over ./internal/cqrs reports SIX findings (F032 ERROR, F031 WARNING, F019/F009/F021×2 INFO), not 185 — the older count came from a different linter vintage/scope. Adjudicated ALL six: FIXED the two real ones (F031 `SearchThreads`' messages scan rode the engine's silent 100-row ceiling — a thread search over >100 messages MISSED hits; now `WithLimit(0)`, the package's lifted-ceiling contract, the legacy LIKE walk is the whole corpus; F019 all 13 hand-built cqrs query decls lacked the `Volume(humanScaleVolume)` hint the readmodels side carries — one const + 13 options); ADJUDICATED NOISE with rationale the remaining four (F032 catalog-not-wired — the wiring lives in internal/domain/catalog.go, invisible to a package-scoped scan; F031 `fetchSegments` variadic forwarder — all FOUR callers pass explicit WithLimit, the detector cannot see through the helper; F021×2 fold-count heuristics on the deliberate threads/messages projection models; F009 the visibility await's `time.After` is best-effort read-your-write, journal-durable by construction — not the durable-deadline class the scheduling module owns). Package green after the fixes; the noise classes are recorded HERE, not filed upstream — the F031-through-variadic and F032-scope classes are detector-limitation reports to go-cqrs-lite ONLY if they recur in practice (verify-before-filing bar: one live false-positive sighting each, no noise-issue spray). RE-RUN 2026-11-12 post-ladder-closeout (pinned cqrs-lint v4.20.1 over ./internal/cqrs): 5 findings, ALL recorded adjudications (F032 catalog-scope, F031 the fetchSegments variadic forwarder, F009 the visibility await's time.After, F021×2 fold-count heuristics) — the M05 push/agent-audit backends added ZERO new classes and one F019 healed; delta negative, no action
- ~~**CI-verdict instrument fixes (the 2026-11-09 misread-green
  class)**~~ → done 2026-11-11 — (a) `scripts/ci-verdict.sh <sha>` names the workflow and answers ONLY the CI run's conclusion (5-case self-test; exit 0/1/2 = green/no-run/red); (b) the watch lane's `run-name` now says "NOT a CI verdict"; (c) the workflowName-verified rule sits beside the check-CI rule in AGENTS. Live proof the day it landed: the tool exposed a real CI red (vendorHash mismatch) hiding under green watch rows
- ~~**Mechanical ride-along probes for code trains**~~ → done 2026-11-11 — `scripts/pre-push-gates.sh` runs the six push-critical gates in ONE command (registry freshness, art-dupl pinned check, changelog-guard, templ-generate --check, build-tw-css --check, docs-gates --check); `--only=`/`--skip=` per gate
- ~~**Smoke scenario for `-enroll-passkey`**~~ → done 2026-11-11 — `enroll_scenario` in webphone-smoke.py: CLI one-shot + server under ONE config file (the shared-journal contract), fresh token verifies 200, replay answers 503, /enroll page renders; full suite 8/8 scenarios green (the split-brain class is now webphone-caught)
- ~~**Doctrine-doc sweep after the enroll CLI journal swap**~~ → done 2026-11-11 — AGENTS' ONE-journal-world bullet and README § Passkey were already current; the one stale artifact (userauth's `readmodels.enroll_*` code reference) fixed to `cqrs.enroll_*`; `readmodels.EnrollTokens` confirmed PROD-DEAD (composition root wires `CqrsService.EnrollSvc()`; zero non-test callers) — its deletion JOINS the readmodels faxes-store row below
- ~~**Clone the transport-isolation fix into pbx, asr, matrix,
  paperless**~~ → done 2026-11-11 — the crm Clone pattern extracted into the ONE home `internal/httpclient.Isolated` (stdlib-only leaf + micro-test) and migrated: pbx ×2 (incl. the HTTPClient fallback), asr ×2, matrix, push ×2 (unnamed in this row but the same bare-literal class), crm's plain branch; paperless rides the go-paperless SDK (no local client — the row's title overcounted). Production-neutral
- [ ] **Hub runner durable fix** — ExecStartPre provisioning (or a
      static uid / group-write) for the build-state roots, so the next
      DynamicUser renumber cannot red every hub build (the 2026-11-08
      chown was in-place; the durable half lives in the hub's TODO).
- [ ] **Hub ALERT_WEBHOOK_URL decision** — the failure log says
      channel decision pending (ntfy / email / Forgejo webhook); pick
      the channel and set the secret so hub failures alert.
- ~~**Verify catalog.home.lan serves run #240's publish** (the
  hourly sync half) and the site freshness WARN cleared.~~ → done 2026-11-11 (the observable half) — the site serves a coherent publish: root 200 with generation stamp 2026-11-10T22:59:53 (run #240's window), and `/docs/domains/telephony/1.0.0` server-renders webphone's CURRENT event vocabulary (Fax Advanced / Message Received / Thread Flag Changed, 34 webphone mentions) — the freshness proof that does not need hub access. The hub-side WARN log line itself stays deployment-access territory (forgejo.home.lan host); if a newer publish lands and the WARN recurs, that read belongs to the hub session
- [ ] **Ringhook operator docs + typed mirror**: the stack's deploy
      docs gain the enable recipe (config keys, secret slot, expected
      dialplan arms); consider a typed ringhook front in webphone's own
      NixOS module for non-stack hosts; a ringhook health leg joining
      telephony-health; the operator command sheet recipe.
- ~~**Confirm crm's own gates on the vendorHash repair** (`37a1957`):
  the hub build proves the build; crm's treefmt/buildflow verdicts
  were not read this session.~~ → done 2026-11-11 — crm's full test suite green (`nix develop -c env -u GOEXPERIMENT go test -count=1 ./...` EXIT=0, all internal packages ok at e383ba9); treefmt step is not_applicable there (no treefmt.toml — the tool answers REJECTION honestly, not a red); the vendorHash repair never touched crm's tree (webphone-side pin only)

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

- ~~Drafts-poll skip-if-dirty guard (F11)~~ → done 2026-11-11 — the dirty flag (WeakSet keyed by the textarea node) gates the poll alongside the existing `[open]` guard; folded-shut-but-typed edits no longer lose text; 3-case spec in drafts-guard.test.mjs
- ~~Localize island-built aria-labels (F9)~~ → done 2026-11-11 — the last hardcode (typeahead's listbox) rides `t("typeaheadLabel")` (both i18n maps carried the key; main.js already re-labels on switch); the other seven were localized by later trains; shell.js copy stays English by the D3 ruling
- ~~Chip-X 24px hit area (F14)~~ → verified LANDED at HEAD before this sweep — the glyph stays 18px and the invisible `::after` (`inset: -3px`) extends the hit to 24px with the F14 comment at the site (app.css)
- ~~History empty state rewording (F15)~~ → verified LANDED 2026-11-07 (T16) — the review card itself carries the FIXED banner; `history.needsAPI.pre/post` carries the three-beat shape (en+de)
- ~~Radius/spacing token adoption (F17)~~ → done 2026-11-11 — the stragglers swept: voicemail waveform, mono input, admin-ext chip → `--radius-sm` byte-identical; island `#lang`/small buttons/audio-out selects (8px literals between the tokens) normalized onto `--radius`; tw.css/login.css rebuilt; the 2px fax-step bar stays (a 4px-tall shape no token describes)
- ~~Phase-4 release preflight (harvested 2026-11-08 from the 17-22
  SUPERB report §f-2 residue): the KVM/local-only flake legs —
  `nix build .#checks.x86_64-linux.webphone-module`, the
  `webphone-backup` + `webphone-backup-drill` VM tests — before the
  SUPERB release train; CI cannot exercise them.~~ → done 2026-11-11 — all three named legs green on KVM at the 6d2306ed tree: `webphone-module` (eval check, ad03g91w), `webphone-backup-drill` (8d1f8sw), `webphone-backup` (the full VM run, 24bnbx00). Side finding from the same sweep: the FULL buildflow run's go step flaked on `TestAutoReviewDispatchRoundTrip` (`sql: statement is closed` — the detached review goroutine racing the fixture engine's cleanup under sandbox timing; the test passes plain AND `-race` locally including the whole server package at 137s) — a shutdown-race flake class, not a regression; the CI `-race` legs that reached test execution passed

### 2026-11-07 Datastar deep-dive residue (one home: the report)

The full client-side adoption question is CLOSED (NO-GO, posture 78/100,
docs/research/2026-11-07_datastar-deep-dive.html — supersedes the 2026-11-04
go-datastar adjacent ruling's trigger paragraph). One actionable item fell out:

- ~~Retire `/health` `'unsafe-eval'`: go-health-dashboard upstream adopts the
  embedded Datastar v1.0.3 nonce + Trusted Types lane (emit `data-nonce` on its
  page, grant `'nonce-…'` instead of `'unsafe-eval'` in its csp.go), then
  re-pin `TestAppDashboardMountsHealthSubtree` (app_test.go:135) to assert the
  nonce form. Evidence: report F4.~~ →
  adjudicated 2026-11-12 — the "nonce INSTEAD OF unsafe-eval" premise is unsatisfiable: the shipped csp.go (v0.13.1, commit 449aeb3) adds the nonce BESIDE `'unsafe-eval'` because the Datastar SDK compiles its data-* expressions with the Function constructor regardless of nonces (library comment + webphone dashboard.go + deep-dive F4 agree). The ACHIEVABLE half already landed and is pinned: per-request nonce'd inline scripts + the nonce'd CSP are asserted by TestAppDashboardMountsHealthSubtree (both tokens + rotation). The eval-free re-open trigger lives with Datastar upstream (expression compilation without eval) or a dashboard drop of Datastar — not this repo. ~~2026-11-13: dashboard v0.13.3 shipped the SDK-side half (the `data-nonce` root bridge, upstream issue #23 — the SDK reads the page nonce to evaluate "without the new Function path"); when the v0.13.3 bump rides, re-check whether csp.go can finally drop `'unsafe-eval'` on a nonce'd host.~~ → done 2026-11-13 — the v0.13.3 bump LANDED and the eval retirement SHIPPED WITH IT: the runtime can already avoid the Function path under the root nonce (upstream browser-verifies the exact no-eval policy in TestBrowser_StrictCSPNoEvalConnects), so webphone's `dashboardCSP` now strips the `'unsafe-eval'` token from RecommendedCSP (one line; the library stays the single source for every other directive); /health serves `script-src 'self' 'nonce-…'` with NO eval, pinned by the app tests (inverted assertions) + the smoke `dashboard nonce csp` check; full smoke 54 + all scenarios green; vendorHash re-pinned.
- ~~**Implement the dashboard lane in the library repo** (go-health-dashboard):
  the nonce emit + CSP grant above, then run ITS gates there — this repo cannot
  ship the change and must not fake it locally.~~ →
  done (by train) — the library shipped cspPolicy(nonce, hardened) + strict variant (go-health-dashboard 449aeb3, first released v0.13.1); local checkout verified 2026-11-12
- ~~**Release the dashboard version** (tag) once the lane lands — owner call per
  the report's §g1; only then can webphone bump and re-pin the CSP assertion.~~ →
  done — v0.13.1/2/3 tagged; webphone pins v0.13.2 and the CSP assertion already asserts the nonce form (app_test.go TestAppDashboardMountsHealthSubtree)

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
10. ~~**Wire the `docs-gates` battery into CI**~~ — done 2026-11-11 (the sibling's CI leg): `ci.yml` fast-checks runs `scripts/docs-gates.sh --check` (line ~115), the non-mutating mode; the ratchets hold
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
12. ~~**`/context` contract snapshot test**~~ — done 2026-11-11: the producer's RENDERED bytes captured from the CRM's own machine API (crm@85c7f2e, throwaway dumper deleted, tree verified clean) pinned in `internal/crm/testdata/kith_context_snapshot.json`; one test parses those exact bytes into the mapped Context, a second keeps fixture↔snapshot honest both directions (every producer-emitted key must exist in the fixture with the same JSON type; the fixture may not invent keys)
13. [x] **History nudge for the notes save path — DONE 2026-11-07 (P2-3)**: a
        successful `call_notes` PUT publishes the payload-less `notes` SSE event
        (`NotifyCallNotesChanged` on the api Server seam, same shape as the contacts
        nudge, fired only on success), the History panel subscribes with
        `hx-trigger="sse:review, sse:notes"`, and the generated wire contract now
        declares BOTH `review` (F89 shipped it undeclared) and `notes`. Pinned by
        `TestCallNoteSaveNudgesHistory` (exactly one payload-less nudge to the
        owner's hub).
14. ~~**`scripts/templ-generate.sh` wrapper**~~ — done (landed 2026-11-07 by the sibling session; verified in-sync at templ v0.3.1100 today) — reads the pin from go.mod, warns on a disagreeing PATH templ, `--check` snapshots + regenerates + restores
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
18. ~~**Enforce CHANGELOG coverage for code-path commits**~~ — done (landed by the sibling session): `scripts/changelog-guard.sh` (+ selftest) — product-code deltas without a CHANGELOG.md change exit 1, tooling/tests/docs exempt per the ratified bar
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

### go-appkit module candidate (2026-11-12 re-review — docs/review/libs/go-appkit.md)

- [x] ~~**File the webphone walls UPSTREAM as appkit asks (fleet frame, owner ruling
      2026-11-12: the walls are extraction backlog, not waiting conditions — the owner
      owns both repos)**: (1) appkit/health per-SURFACE criticality grading (the same
      check leg 503s a strict readiness route but WARNs the dashboard — health_crm_test.go
      is the spec); (2) an aggregates-only metrics mode (no `route`/`method` labels — the
      /metrics leak-test privacy line); (3) httpx success-status override (202-class
      lanes) + per-route family→status tables.~~ **DONE 2026-11-12 — filed as
      LarsArtmann/go-appkit#59 (strict readiness), #60 (aggregates-only metrics), #61
      (httpx success-status; the ask NARROWED during source verification — error-side
      statuses were already expressible via errorfamily's `HTTPStatuser`, so the
      per-route-tables half was withdrawn as unnecessary). All three carry the
      provenance banner for owner review; drafts in docs/drafts/appkit-asks/.** Adoption
      order stands: simpler apps ride preset first (appkit's own battery-swap recipes);
      webphone migrates last. Evidence: docs/review/libs/go-appkit.md § the open-mind
      pass (issue links inline).
- [ ] **Adopt `appkit/flightrecorderhealth` (first go-appkit MODULE adoption; framework stays NO-GO)** —
      wire `frhealth.NewTrigger(rec, WithCooldown, WithTriggerFunc(fr.OnError))` into
      `health.New` via go-health's `WithHealthRecorder` (options.go:127 in the pinned
      v0.10.0) where app.go already builds both the probe and the recorder: health-batch
      failure auto-captures a trace snapshot (cooldown-hardened, non-blocking, race-safe
      — all pinned by the module's own tests), covering wedges that never surface as a
      traced request (the dashboard's 1s loop failing with nobody polling /healthz) and
      adding the cooldown the un-cooled 5xx snapshot arm lacks. Both hard deps already
      pinned (go-flightrecorder v0.2.1, go-health v0.10.0); standalone constructor — no
      injector-doctrine conflict. Demand-gated: the first wedge incident the
      request-trigger misses, or ride the next flight-recorder train. Evidence:
      docs/review/libs/go-appkit.md § module table (2026-11-12).

### Fleet-libs reopen-trigger registry (2026-11-04, T25 — the ONE table)

Every NO-GO/declined/keep-as-is ruling carries a written trigger in its report
(`docs/review/libs/*.md` § Ruling — the one home). This table consolidates them so a
sweep answers "is any trigger fired?" in one read; re-litigate a ruling ONLY when its
trigger fires, never on a schedule.

| Fires when                                                                                           | Library                | Then                                                                                                                   |
| ---------------------------------------------------------------------------------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| metaengine grows NESTED group addressing (a per-call floor inside a per-owner group)                 | go-cqrs-lite           | `trimCall` is the last hand loop; delete it for the declarative cap                                                    |
| usermgmt major bump, or a second identity consumer                                                   | cqrs-htmx              | re-run the family audit (adminui re-check)                                                                             |
| `internal/domain` gains a second Go consumer                                                         | go-branded-id          | ratified modularization trigger — brands need their own module                                                         |
| A client-facing `Idempotency-Key` contract appears on a public API                                   | go-idempotency         | re-run the adoption ruling (metaengine DedupStore vs the SDK)                                                          |
| A self-managed versioned blob format or event signing requirement                                    | go-codec               | `DeterministicCodec` is the entry point                                                                                |
| A drop-folder/intake feature, or live-config-reload demand                                           | go-filewatcher         | re-run the NO-GO ruling                                                                                                |
| A `Type=notify` conversion, or a unix-socket machine surface                                         | go-daemon, go-appkit   | re-run both rulings together                                                                                           |
| An appkit tag beyond v0.12.0 ships a new battery/module (httpx B6/B8, preset legs, W4 worker/polite) | go-appkit              | sweep the module table in docs/review/libs/go-appkit.md — the 2026-11-12 re-review proved the inventory drifts in days |
| A wedge incident the request-triggered trace snapshots miss (loop failure with no traced request)    | go-appkit              | adopt `flightrecorderhealth` (the module-level candidate row above)                                                    |
| A CLI subcommand outputs structured data                                                             | go-output              | adopt the renderer                                                                                                     |
| A non-systemd target, credential-compliance rule, mutable secrets                                    | go-credentials-store   | re-run the NO-GO ruling                                                                                                |
| A new file-persisting feature (export, backup, rotation)                                             | go-atomic-write        | ride `WriteWithPerm`/`WriteIfChanged`                                                                                  |
| A latency investigation the histogram aggregates cannot localize                                     | httputil/server_timing | instrument 2–3 named spans behind the same env gate                                                                    |
| User-visible missed-update reports on flaky mobile links                                             | go-sse                 | ssetest-pinned Last-Event-ID replay                                                                                    |
| A second retry-shaped loop (e.g. ASR uploads)                                                        | go-retry               | `retry.Do` is the natural home                                                                                         |
| A dashboard deployment without the Caddy front                                                       | go-health-dashboard    | re-decide the fence (see the 2026-11-04 fence memo)                                                                    |
| A read model crossing ~10k rows/owner, or a second engine                                            | go-cqrs-lite           | keyset pagination / failover audit                                                                                     |
| A second readiness consumer or multi-instance deployment                                             | go-health              | re-run the probe-topology ruling                                                                                       |
| A templ-components release with a new status/toast component; the subtitle/✓-chip taste rulings      | templ-components       | owner-taste re-litigation, not adoption gaps                                                                           |
| The dashboard library swaps its embedded SDK version                                                 | go-datastar            | re-verify the /health CSP scope                                                                                        |
| A product surface wants server-pushed reactive state (live Reports-shaped, not fragment morphs)      | go-datastar            | re-run the client-side NO-GO (2026-11-07 deep-dive, posture 78/100)                                                    |
| The library ships a parse-with-rules hook or a new linter rule                                       | go-branded-id          | migrate the invariant INTO the seam / adopt the rule                                                                   |
| A second binary in this repo, or an interactive operator CLI beyond the three ritual one-shots       | cmdguard               | re-run as an adoption audit (docs/review/libs/cmdguard.md, 2026-11-11)                                                 |

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
5. ~~**Flush-announce island spec**~~ — done 2026-11-11: `island-tests/transcribe-announce.test.mjs` (steppable clock) pins both arms — the lag warn announces once per streak and a fast segment re-arms it; the unavailable warn waits for the third consecutive failure and stays quiet while the streak continues (the 60s backoff gate is part of the sequence)

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
the hub rides `branch: main` (no sha pin to bump — the "re-pin" is the
mirror-sync build), the Forgejo webphone mirror was force-synced to the GitHub tip
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

- ~~**OWNER-GATED: catalog v4.9.0→v4.10.0 bump** — 213 commits
  (architecture-reference + export-selection features webphone does
  not use; the seven upstream candidates are NOT obviously fixed).
  Lean: ride the next natural train. Same notes file. **Priority:**
  Low · **Effort:** S (decision) / M (bump).~~ →
  superseded 2026-11-12 — the pin is ALREADY v4.14.0 (go.mod; later
  natural trains rode it well past v4.10.0), and the plan's M15
  resolvability check (F045) found NO v4.15.0 published upstream —
  the pin is CURRENT, this row is moot; a future bump needs an
  upstream release first.
- [ ] **webphone-android consumes the specs** — REMAINING HALF is the
      runtime digest handshake: fetch a contract digest of exactly the
      paths/shapes the app uses at origin-config time (the 304 ETag
      lane makes it free) and render a named "server older than app"
      surface on mismatch. The BUILD-TIME half LANDED 2026-11-10:
      `webphone contracts` renders byte-pinned consumer fixtures
      (`docs/contracts/wire/` — push facts both transports, both
      sessionIdentity arms, the registration request), freshness-gated
      with the asyncapi document; the android repo vendors them and
      its JVM tests decode them through its real parsers
      (`scripts/sync-wire-fixtures.sh` there; the runbook carries the
      re-vendor leg). Trigger for the runtime half: first real
      server/app version-skew incident. **Priority:** Low ·
      **Effort:** M→S (build-time half done).
- [ ] **Wire-fixture follow-ups (deferred, with triggers)** — (a) the
      android-side loopback live-check script: boot a loopback webphone
      and drive the app's real decoders against its LIVE `/api/session`
      + `/config.js` answers (fixtures pin the shape, this would pin a
      deployment; trigger: device-train prep wanting deployment-level
      proof); (b) CONSIDER a `/config.js` key fixture pinning
      `sipDomain`/`websocketPath` in the same sync (trigger: the first
      config.js key drift or a second config consumer — the android
      repo owns the regex reads). BOTH TRIGGERS RE-CHECKED UNFIRED
      2026-11-13 (pareto T20; config.js keys stable, no second
      consumer, first device train still operator-blocked) — the
      handover pack (docs/planning/2026-11-13_16-40_android-handover-pack.md)
      carries the check + the line-picking/CallStyle contracts. Rows mirrored in webphone-android's
      TODO_LIST; the KMP spike stays deferred per the 2026-11-08
      preflight (quiet-repo gate). **Priority:** Low · **Effort:** S each.
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

### ~~internal/server auth-package carve~~ → EXECUTED 2026-11-12, extended to platform; html carve RULED OUT with evidence

The trigger fired (the 2026-11-11 login-page front-door train) and the carve
LANDED as `internal/server/auth` — both login modes, the ceremonies, the OIDC
dance, the library /login page and the session spine, behind a narrow
`auth.Shell` seam (the 2026-10-17 api/hooks pattern). Step 2 landed the same
day: `internal/server/platform` — assets, /metrics + recorders, CSP intake,
flight-recorder seam (+ TraceRecorder's new home), /phone-api proxy,
directory bridge (+ machineSurface), /config.js, export. **The planned html
carve (pages/panels/actions/notifier/sse) is RULED OUT structurally**:
pages.go holds the `handlers` hub itself and panels/actions reference
admin/crm/unread/auth-audit in both directions — moving them renames the
package rather than splitting it. The god-package's honest end state is the
four surface carves (api/hooks/auth/platform) + the depth-budget arch test
(landed beside this row), with the html hub as the measured, named residual.
Reopen the html question ONLY on a real architecture change (per-tab render
packages over a shared kernel), never as a mechanical move. The 2026-10-17
adjudication lives in git history (the 2026-10-12 + 2026-10-17 entries).

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
- **Shell leg (webphone-android):** first device train — exercise the SIP UA against the live WSS gateway (REGISTER/INVITE/BYE/DTMF/hold + incoming INVITE + REFER/NOTIFY are implemented but UNTESTED against FreeSWITCH), compose the real CallStyle notification (v0.1 posts the FSI heads-up), and the origin-picker settings UI (App.origin is a constant today). The `/api/session` field-name check LANDED 2026-11-10: the wire-contract sync train fixed the android `resumeSession` drift (it read `sip_domain`/`sip_password`/`websocket_url`, keys the server never sends — REGISTER ran on empty credentials) and now pins both sessionIdentity arms + the push payloads + the registration DTO via byte-pinned fixtures (`docs/contracts/wire/` here, vendored + decoded by `WireContractTest` there; re-vendor ritual in the release runbook).

**Evidence:** decision doc `docs/planning/2026-10-23_16-13_native-app-platform-decision.md` (status: DECIDED + SHIPPING); `~/projects/webphone-android` README honesty markers; `internal/push` + `internal/server/hooks/ring_test.go` + `internal/server/push_api_test.go` for the shipped server contract.

### Trunk truth + v2.13.0 release train — open rows (2026-11-10 harvest)

**Status:** 🔴 the rows below are the OPEN tail of the 19:14 report's §f (the satisfied half is struck in the report itself) · **Priority:** see each row · **Effort:** S each unless noted

- **OWNER legs (a sitting, not a script):** A3.4 — send the assembled Telnyx reply into ticket #832080 (paste-ready at pbx-artmann `docs/evidence/2026-11-10-trunk-siptrace/FINDINGS.md`, attach the capture, reply INSIDE the thread); C3.2 — one real inbound mobile call to the DID (closes the six-week zero-inbound ruling only when a `context=public` CDR lands); E3 — pick the `reports.weekly_talk_budget` number and set it in the host config; E5.2–E5.4 — hikaru's passkey one-time link + enroll + the Super Admin 3-surface check; E6 — the first real outbound fax (owner picks the destination, watch webhooks → delivered); row-33/34/35 confirmations — listen to the `pifo42h1` recording once, say what 9100 is, decide whether internal test destinations get contact-store names.
- **R-URI `+` dialplan train (8):** if Telnyx engineering confirms the leading-`+` R-URI question, the dialplan fix is OURS — same-day train, `trunk-lift-watch` watches for the answer.
- **C5 lift day (15):** probe rc=3 twice → capture the LIFTED wire beside the screened one (re-run `trunk-debug-matrix.sh` with `PBX_MATRIX_KEEP=1` into a durable path — the 2026-11-10 /tmp cell logs rotted) → first real PSTN dial + CDR + recording → inbound green → strike `trunk-lift-watch` with the evidence lines.
- **Watcher fire-drill (13):** deliberately trip one alert class (canary against a deliberately-wrong host in a scratch run) to prove the GitHub-issue filing + suppression path works inside the systemd user environment — the alert leg has never fired in earnest.
- **api-dial tripwire (14):** the 2026-11-17 revisit of the DEFER recommendation self-surfaces here (memo: pbx-artmann `docs/planning/2026-11-10_17-30_api-dial-stopgap-decision-memo.md`).
- **Release hygiene on a quiet host (22/30/31):** owner call on `.buildflow.yml`'s stale baseline file (157 accepted entries match nothing — delete or refresh deliberately); re-run `buildflow timings --regressions` on a quiet host before rebaselining the +200% race-step timing; correlate the one SQLITE_BUSY warning from attempt 2's tail if it recurs.
- ~~**lychee redirect hints (32):**~~ partially — 0 errors at today's run (579 OK, the 5-vs-0 delta reconciles as green); the 3 redirecting URLs are UNNAMED by lychee 0.24.2 in every local output mode (-v/-vv/json/detailed carry only the count — an upstream display gap worth filing, owner-gated); no URL guessing

### Release-green + B5-staged — open rows (2026-11-11 harvest)

- **Answer the 00-45 report's §g Q1–Q3:** (1) B5 prod-switch authority + timing, (2) main's treefmt red — heal over the other session's refactor files or let them self-heal, (3) the E3 budget number.
- **Verify webphone CI green on the treefmt-healed tip:** the daemon-swept refactor files owe a formatter pass; main sat red across four consecutive tips (`a3832e4a` → `346cad88`), treefmt-check failing on gofmt alignment around `cqrs-lint:ignore` comments.
- **Post-switch verification battery** once B5 executes: prod `/version` = `v2.13.0`, journal quiet (no unknown-config-key WARNs), `ui-capture --track` Reports causes subline + voicemail honest-toast verdict on real data, pbx-artmann deploy-ledger row.
- **Release-log archival:** the `/tmp/release-2.13.0-{5..11}.log` evidence was LOST to a tmpfiles purge before it reached the docs evidence home — row 32's rot warning materialized. The runbook addition below is the fix; the raw logs are unrecoverable (the verdicts live in the CHANGELOG + the status docs).
- **aarch64 toplevel ruling:** the `pbx-aarch64` toplevel is EVAL-proven (it evaluated and reached building) but its full build is QEMU-bound on `stalwart` (a Rust mail server compiled from source for aarch64 — hours, silent). Prod is x86_64. Revisit only with an aarch64 host, or via a remote aarch64 builder.
- ~~**Runbook additions:** `--check`-mode docs battery as the pre-push verdict (the mutating battery hid the rule-F glue that red-ed CI); verify the HEAD sha is unchanged in the same call as the push (the daemon swept 28 files between verify and push); archive one gate-clean report in the same push that files a new one; copy release attempt logs into the evidence home AT attempt completion — never "later" (tmpfiles purges faster than a session idles).~~ → done 2026-11-11 — all four landed as release-runbook.md § "Train-hygiene rules from the v2.13.0 seven-attempt war" (check-mode pre-push, HEAD-assert-in-push-call, archive-when-filing, logs-home-at-creation), plus the entangled-red-train rule the day surfaced
- **18-08 report drain train** before the 45-day cadence age threshold (~2026-12-07): 102 bare rows across 13 tables, zero adjudication — the oldest live report.
- ~~**cadence-candidates script:** turn the archive-candidate scan (check-rows + bare-rows over every live report, oldest-first CLEAN list) into `scripts/cadence-candidates.sh`.~~ → done 2026-11-11 — script landed; first run: 9 clean candidates / 57 dirty (INDEX.md excluded); builds the session-fresh annotate tool into /tmp/annotool when missing
- ~~**Runbook additions (dup):** `--check`-mode docs battery as the pre-push verdict (the mutating battery hid the rule-F glue that red-ed CI); verify the HEAD sha is unchanged in the same call as the push (the daemon swept 28 files between verify and push); archive one gate-clean report in the same push that files a new one.~~ → done 2026-11-11 — duplicate of the row above; the canonical row carries the verdict
- ~~**erraudit tier re-measure** (overdue since 2026-11-05): tier 1 exit 0, tier 2 must stay 0.~~ → done 2026-11-11 — BuildFlow erraudit step 1/1 green (RC=0, detect 3.2s); the raw CLI's 119 unconfigured findings are the bypass-suppression class, not tree regressions; next calendar due 2026-12-08
- ~~**Red-main watcher visibility:** no tracker issue visible through four red tips — verify the LIST-API title-dedup still matches when every sweep changes the commit subject. UPDATE 2026-11-11: the 09:47 island-js red (tip `9b47bc78`, fixed only inside another session's unpushed train) has NO tracker issue either, 4h+ red — the watcher's red-detection needs a live look, not just the title-dedup theory.~~ → done 2026-11-11 — the live look FOUND the defect the theory missed, two layers deep: (1) the dedup lane rode `gh issue list --search` (the lagging search index — the exact 2026-11-04 class), and (2) `--label ci-red --limit 1` picked ANY ci-red issue — six open daemon-audit flags share that label BY DESIGN, so the next red-main would have commented its evidence onto a daemon-audit flag and then spammed it every 15 min (CREATE was unreachable). Fixed: `red_main_issues()` = LIST api over ci-red filtered by the `red-main run` title prefix (daemon-audit's exact-title dedup untouched — its own lane already used the LIST api); decide() core unchanged; self-test 11/11 with a lane-fixture case; live dry-run answers CREATE against the real red tip (run 43027836672, the sibling train's art-dupl leg) — correct on both inputs.

### 18-08 report drain harvest — open rows (2026-11-11, the archive-drain routes)

- **Flight-recorder + metrics hardening cluster** (routed from the 18-08 report's f26/f38/f39/f40/f41/f42/f47/f49, e6/e8/e10): trace observation vs the `/metrics` exclusion alignment; a NixOS-module declared surface for the env-only recorder; a rate limit on `/internal/flightrecorder`; a snapshots-taken metric (count + bytes); a startup log line when the recorder is on; disk-aware `maxBytes`; an arch test pinning `h.notFound` as the only HTML-router 404 writer; a quarterly re-rule of the remaining `http.NotFound` sites; the `ResponseRecorder` sendfile cost on downloads; an IPv6-loopback (`::1`) pinned case; the route-family test's say-so where it lives.
- **Traces-in-backup ruling (owner)** (routed from f22/f23, d4, g1): should `<dataDir>/traces` ride the backup set or be excluded from the module's backup unit + documented. The 2026-11-09 journal swap moved the db backup set to system.db; the traces directory question was left unanswered by it.
- **Observability docs rows** (routed from f32/f33, c5): FEATURES.md owes the `/metrics` histogram wording (the row still says "build info, uptime, table counts") and flight-recorder rows; ROADMAP.md owes an observability-depth entry. Grep-verified absent 2026-11-11.
- **Lessons candidates** (routed from f35, e3, e4): the "plan says X, the code pins Y — amend the plan in the same turn" lesson; the "check the backup/test-path consequence of every new on-disk artifact under dataDir" process lesson. Grep-verified absent from docs/lessons.md 2026-11-11; the daemon-attribution lesson (f34) ALREADY landed (lessons line ~1062) and is not part of this row.
- **go-appkit report refresh trigger fired** (routed from f37): the report's flight-recorder fact says "indirect dep"; the seam is a DIRECT dep since the 2026-10-23 train (AGENTS flight-recorder bullet). Re-scoring rides the report's own refresh trigger, not a re-litigation.
- **Cadence drain continuation + ratchet re-measure** (the 2026-11-11 13-54 report's §f): re-measure the annotate_max pin at the next sweep (the 18-08 + two-candidate archives moved corpus counts; ratchets move DOWN only, so a lower measured count is the expected direction); drain-train the next-oldest DIRTY candidates from `scripts/cadence-candidates.sh` output before their 45-day age thresholds; decide heal-or-wait for the active login-page train's docs-gate reds (the 12-31 report's missing §f + dprint format) if that session idles. That train LANDED 2026-11-11: the island suite (330/330) + smoke pins (54/54) healed, the art-dupl templ clones dissolved at the sites (baseline untouched, registry sweep log appended), and the fold shipped in 2.14.0.

### cqrs-lint triage follow-ups — open rows (2026-11-11 harvest, the 01-58 report's §f)

- **File the suppression state-machine draft upstream** (`~/projects/BuildFlow/docs/drafts/cqrs-lint-suppression-block-comment-state-machine.md`; owner review tick first) — and the **D005 draft** beside it (same gate, older).
- **Track the upstream `/*` lexer fix**: on the next cqrs-lint bump, remove the actions.go hedge comment if the fix shipped; until then any directive below a one-`/*` line comment in the six swept files is dead on arrival.
- **Confirm the enroll.go S001 heuristic change upstream** (the wirefixtures twin still fires — intentional change or drift?).
- **Train-specific decay pass: `erraudit nolint-audit`** after the C007/At-command train (the calendar's quarterly re-measure is separate, next due 2026-12-08).
- ~~**Pin the D014 tags with a replay test**~~ → done 2026-11-11 — `internal/cqrs/tombstone_wire_test.go`: untagged old-shape witnesses replay through `DecodePayloadAuto` and must re-encode byte-identically; JSON wire keys pinned to the Go field names with byte goldens; MUTATION-PROVEN (a planted `json:"reason"` drift fails both legs)
- ~~**`cqrs-lint scorecard` before/after** and **`cqrs-lint doctor` render check**~~ → done 2026-11-11 — scorecard: adoption 11/36 relevant modules (30%, Grade Sparse) and the score counts MODULE adoption, so the 11 disabled LINT rules could not distort it; doctor renders all 11 disable-reasons (217 rules total, 206 active, 11 disabled, each `why` verbatim)
- ~~**Write the triage decision table**; move three lessons~~ → done 2026-11-11 — the 4-way table (config-disable / directive / keep-visible / fix) is the standing rule beside the CHANGELOG bar in AGENTS (cites the 00-36 report); docs/lessons.md § Tooling traps gains the result-cache war story, the scoped-verifier near-miss, and the vendorHash-tail red (CI's own-hash repair)
- ~~**Pin the B007 ruling**; consider feeding the stream-vocab vars to the catalog (B010)~~ → done/ruled 2026-11-11 — the ruling comment sits at server.New's registration block (explicit rows ARE the route table); B010 DECLINED: the catalog's scope is the four SSE seam events by ruling (nothing fabricates a CQRS write model into the federation), the stream vocab is the internal write model, and the seam events do not map 1:1 onto streams
- **Owner decisions pending**: cqrs-lint gating posture (answer or TODO row), C044 prefixed-key permanence (answer or priced TODO row), BuildFlow rebuild + reinstall (freshness only), dprint-remote-plugins advisory, `buildflow format` fingerprint drift, test-race +202% on a quiet host.
- ~~**Cosmetic**: the AGENTS.md erraudit bullet's stale inline "next 2026-11-05" date~~ → done 2026-11-11 — the bullet now defers to the watch calendar (next 2026-12-08, last measured 2026-11-08)

### Full-code-review open rows (2026-11-11 harvest, the 18-23 report's verified findings)

- ~~**Hold/resume must settle on the re-INVITE verdict, not the send**~~ → done 2026-11-12 — `sipverdict.js` (`inviteVerdict`) wraps the re-INVITE in a requestDelegate whose onAccept/onReject settle the promise; `holdSession` (calls.js:117) commits only the far-end verdict (accepted/rejected/send-failed/timeout, 30s cap) and failed toggles return the SETTLED prior state. The verdict-arm UI pins landed the same day (`hold-verdicts.test.mjs`: TIMEOUT arm under mock timers, refused arm, settled-truth restoration + en/de copy pinned verbatim; island suite 336/336).
- ~~**Detach the /hooks/ring fan-out from the request**~~ → done 2026-11-12 — `push.Fanout.Fan(ctx, owner, fact)` (push.go:197) is the ONE delivery policy (WithoutCancel detach + 15s budget + prune-on-unregistered); the ring hook hands off at `internal/server/hooks/ring.go:58` and answers 202 without waiting on transports (pinned by `TestHookRingFansOutToRegisteredDevices`).
- ~~**Probe the concurrent first-sight thread-split race**~~ → probed 2026-11-12, **NOT REPRODUCIBLE — no guard needed** — `TestConcurrentFirstSightYieldsOneThread` (internal/messaging/raceprobe_test.go, race-detector clean): two barrier-started Receives for one unseen remote both miss FindThread's projection scan, but both appends target the SAME journal stream (ThreadStreamKey = owner|network|remote, threads.go:33), the first adopter wins the thread identity, and the decider's `cqrs.thread_identity` rejection (threads.go:103) refuses the loser. Exactly one thread exists after the race; the loser's append fails `store.message_append` and is recovered by re-delivery (the retried Receive resolves the winner's thread, unread counts both) — the webhook lane's provider retry / the receiver loops' retry ladder own that. No uniqueness guard to add. ~~REOPENED 2026-11-12 by #249: the single-run probe missed the repetition class — `-count>1` and full-suite parallelism red the final assertion ("unread 3", pre-existing at `b1b01b87`); the await-poll-vs-semantics decision is parked with the file-size split train residue row above.~~ → CLOSED 2026-11-13 (#249) — decision: await-poll, not semantics: the probe's exact-count `svc.Threads` reads raced the async projection fold; `awaitThreadFold` polls to convergence (5s deadline, the `awaitTranscriptCount` pattern) and both reads route through it. Pre-fix reproduction at current HEAD: NONE (`-count=20`, `-count=10 -race`, package `-count=2` all green — the ladder retirement shifted the fold timing since the b1b01b87 observations); post-fix gauntlet green again. Never red CI (no ledger row — local-repetition class only).
- ~~**Compensating blob delete on persist failure**~~ → landed 2026-11-12 (cheap half) — both messaging lanes (Send/Receive save loops, incl. the mid-loop save failure) and both fax lanes (Send/Receive spool) now remove every saved blob when the persist fails (`rollbackSaved`/`unspool`, best-effort — the persist error stays the verdict); pinned by `TestFailed{Send,Receive}LeavesNoOrphaned{Blobs,Spool}` in both packages. **Owner question stands:** the crash window (file durable, row lost, process dies before compensation) still orphans files — an orphan sweep decision rides the owner-calls batch.

### SUPERB deepening plan — remaining trains (2026-11-12; plan + evidence: docs/planning/2026-11-12_05-25_SUPERB-deepening-and-integrity-plan.md)

T01–T16 are done (strikes live in the plan file + the 2026-11-12 status reports); T17–T23 landed the same day:

1. ~~**T17** — admin CDR party fallback over `views.CDREntry` (F48)~~ → done 2026-11-12 — `cdrParty` (name → number → unknown) leads the admin row; the CDROff/Failed/None ladder is pinned by a structural view test (`e897efe8`).
2. ~~**T18 receiver hardening** (F49–F52)~~ → done 2026-11-12 — matrix answerCap 16 MiB + `matrix.too_large` verdict; poisoned-cursor (400 M_UNKNOWN) + oversized arms self-heal via `resetCursor` with receiver tests; FCM drops the cached grant on 401/403 (cache test); CRM reads capped 1 MiB with `crm.too_large` at both sites (family pins + registry rows landed in the same commit set; tail `d15dcc89`).
3. ~~**T19** — postcall.js extraction (F53)~~ → done 2026-11-12 — the ended-reason/note/disposition/AI-draft family moved to `postcall.js` (calls.js 2 025 → 1 693; q850CauseOf's one home moved with it); three specs re-pointed; closure pin + island 336/336 + tsc green (`e5fcccfc`; one dead import caught by island-lint and removed — lesson: run the flake lint gate before pushing island moves).
4. ~~**T20** — glyph→SVG sweep (F54)~~ → done 2026-11-12 — redial/save/call/remove carry the island's `wp-icon` SVG vocabulary (`b1b01b87`); the shell's pinned "✕ Close" label and the OS-notification titles stay deliberately out of scope.
5. ~~**T21** — owner-calls prep doc (F55/F62)~~ → done 2026-11-12 — Q7–Q17 fill the owner batch's stub section with evidence links + defaults (`e9dfc02e`); F62's posture verification: 1 brandid-lint displacement fixed, #48 + ASR gates confirmed still owner-held.
6. ~~**T22** — erraudit nolint-audit + calendar sweep (F56/F61)~~ → done 2026-11-12 — `erraudit nolint-audit`: 69 directives, 69 needed, 0 stale (the carve trains moved none into decay); the calendar's dated obligations (erraudit/buildflow re-measure 2026-12-08, docs-gates cadence 65/45d) re-verified true.
7. ~~**T23** — cqrs-lint upstream drafts (F57)~~ → verified complete 2026-11-12 — both drafts already exist in BuildFlow `docs/drafts/` (suppression block-comment state machine + D005 first-token misattribution), each with released-binary repro evidence; the owner tick gates the filing.
8. **T24 release tails** (F58–F60) — PARTIALLY done 2026-11-12: the cadence gate FORCED the first drain tranche (3 reports archived gates-first, the residue harvested into TODO_LIST; live corpus 67 → 64); F59's FEATURES `/metrics` row now names the latency + beacon vocabulary (the traces-backup half is owner Q13); F60's release-log POLICY landed in the runbook (raw logs stay out of git; verdict tails go into the filing report). REMAINS: the full 18-08 archived-corpus drain (the annotate pin 260's way back DOWN) before the ~2026-12-07 cadence window.
9. ~~**Island small-hardening residue** (harvested from the 05-22 trains report): note-textarea bytes-vs-chars cap alignment for multibyte; login-page.js localStorage try/catch; dial-form double-submit guard; `htmx.ajax` unhandled-rejection catches in restoreLastTab/live-context paths.~~ → done 2026-11-13 — all four: the note cap aligns at the SERVER (rune count = the wire doc's "chars" and the editors' maxLength; multibyte 4000-accept/4001-reject pins in `TestCallNoteGuards`); login-page.js + i18n.js storage accesses degrade under blocked site data (throwing-storage specs); the dial form's pending flag kills the double-INVITE race (double-submit spec, one consult under a hung gate); restoreLastTab catches → throttled toast, live-context catches → doctrine-silent empty slot (rejection specs assert no unhandled rejections). Island suite 340/340, tsc + island-lint green.

### Gate-setup hardening residue (2026-11-12 five-scoped-run triage — plan: docs/planning/2026-11-12_16-14_SUPERB-gate-setup-hardening.md; status report: docs/status/2026-11-12_15-57_five-scoped-buildflow-triage.md)

1. ~~**Canonical bandit runner**~~ → done 2026-11-13 — `python312Packages.bandit` rides the devshell (locked-nixpkgs pin; 1.9.4 = BuildFlow's runner); fresh full-tree scan green: 0 issues, only the decided-dead B105 nosec furniture warnings (row 4's verify-only class).
2. ~~**`buildflow format` in a quiet window**~~ → done 2026-11-13 — format lane + full battery re-ran green on the quiet tree (23/23, findings gate green); ROOT CAUSE found and fixed UPSTREAM: the drift check returned early without persisting, so a stale record warned forever (webphone's sat frozen 2026-10-23 → three weeks of re-warns) — BuildFlow fix landed `c3eaec3a7` (one-shot semantics, test-pinned; doctor suite green); webphone-side store re-baselined (13 lanes @ templ v0.3.1100, zero warns on re-run). The `~/.nix-profile` buildflow (7e4ca72) picks the fix up at its next fleet profile rebuild.
3. ~~**templ-fmt devshell pin alignment**~~ → done 2026-11-13 — both devshells now carry a source-built templ at go.mod's exact version (`config.packages.templ` in nix/packages.nix, buildGoModule rev v0.3.1100, nixpkgs' lagging 0.3.1020 retired from the shells); `templ generate` with the new pin produced ZERO delta, proving the committed codegen already matches; format gate green (`5186a657`).
4. **Stale `# nosec B105` audit (verify-only, decided)** — bandit warns "nosec encountered, but no failed test" at `render-diff.py:76,77,78,146` + `webphone-smoke.py:1771–1778`; deleting documented suppressions to silence cosmetic warnings is churn (the upstream `-f json` lane removes the whole furniture class anyway) — re-check only when touching those lines. Evidence: BuildFlow bandit stdout, 2026-11-12 runs.
5. **File the bandit upstream ask (owner-tick gated)** — `docs/drafts/2026-10-11_buildflow-bandit-stdout-findings.md` carries the 2026-11-12 reobservation (#64-class linkage, root-cause chain, locally reproducible command, the comma-form nosec gotcha); filing is a human act per verify-before-filing.
6. **AGENTS.md b5 one-liner (owner call)** — one sentence on the per-site cqrs-lint directive convention in the strong-id bullet (2026-11-11 triage report §b5; the 203/377 line cap has room).

- [ ] **Delete the dead sibling legacy readers** (Medium): `legacy_faxes.go` / `legacy_snippets.go` / `legacy_threads.go` / `legacy_transcripts.go` are the same prod-dead class as the contacts reader deleted in `4fb24b61` — zero production callers, and the retention sweep only deletes legacy threads/messages. One deletion pass + FEATURES/docs sync; KEEP the app-flag helpers (matrix/signal sync cursors live there). Evidence: the death-map correction addendum in `docs/planning/2026-11-09_storage-ladder-retirement.md`.
- [ ] **Design the §5 export-or-confirm tooling before any DROP window** (Medium): the legacy contacts table DROP (§5, operator-gated migration) would delete rows that NO retention sweep covers — the death-map correction pins this; the migration window needs export-or-confirm tooling first (dump the table, or record explicit owner confirmation per row class). Evidence: `docs/planning/2026-11-09_storage-ladder-retirement.md` correction addendum; contacts-plan follow-up 23.
