> ARCHIVED 2026-11-08 (docs-health drain, portfolio-plan P2-1): gates clean (check-rows --strict + bare-rows green, 45 rows struck with same-session evidence), citations repointed — the six still-open tool dispositions are harvested to TODO_LIST § BuildFlow tool-disposition residue; the rest landed or live in AGENTS/.buildflow.yml.

# Status: fix-all mandate — research window + parked session (2026-10-11 09:38)

Scope: THIS session only. Window covered: the 06:41–06:47 research burst
after the owner's three directives, the ~3 h park (session left open,
nothing ran), and the 09:36 close-out. The two earlier reports
(`06-00_branching-flow-retriage-gate-status`, `06-34_branching-flow-
retriage-session-status`) cover everything before; this one adds the
mandates, the 13-tool findings classification, and what remains.

Repo state at write: clean tree; `f3bf706` pushed at 09:37 under the
manual-push bar (docs-only commit, daemon push leg stalled from
06:37:25 — ~3 h, far past the 60-min anchor). CI on `da24a12`: green;
`f3bf706` CI verdict to confirm (docs-only delta).

---

## The three mandates (received 06:41)

1. **"FUCKING GO AND START FIXING ALL! PROPERLY"** — the 642 findings
   across 13 reporting tools get honest dispositions: real findings
   fixed, noise classes properly routed (config/rationale), nothing
   blanket-muted.
2. **"CHANGELOG should stay up to date!"** — `[Unreleased]` gets the
   needle-rename + gate-rewiring lines (currently missing), and the
   changelog bar is applied going forward.
3. **error_rules: undecided; go-error-family is "always nice to
   have"** — the ready-to-paste elevation block stays parked in
   `.buildflow.yml`; go-error-family gets a light look during execution
   (nothing in the mandate requires adopting anything).

Plus the SUPERB template: Pareto tiers (1%/4%/20% + the other 20%),
comprehensive plan (30–100 min tasks) TABLE, micro-breakdown (≤12 min)
TABLE, plan doc at `docs/planning/<ts>_SUPERB-*.md` with a mermaid/d2
graph, then execute/verify, detailed commits, push.

## Self-review (this window — brutal first)

**What did you forget?**

1. **The research never left /tmp.** The entire 13-tool classification
   (the single most valuable artifact of this window) lived only in
   `/tmp/f-*.json` and my context for ~3 h of parked session. Had the
   session died, it would be gone — the exact failure mode the
   github-voice skill warns about for drafts. This report now captures
   it; the lesson: persist research to the repo the moment it exists.
2. **TODO_LIST staleness noticed twice, fixed zero times.** While
   reading it I saw sitting items 2b and 3 are already DONE (the
   re-triage + #9 filing were this session's earlier work) — and again
   did not edit on sight. Same slip class as the 02-19 doc formatting
   in the last report. Twice now. Stop noticing; start fixing.
3. **The lychee discrepancy was left unexplained**: 5 findings in the
   cached full run, 0 in the isolated drill. Possibly network-dependent
   or a files-scanned difference — flagged, not resolved.
4. **No visible phase marker before drilling**: the SUPERB template
   says PLAN FIRST; I legitimately needed per-tool sizing first, but I
   never SAID "research phase, plan next" — ten minutes of silent
   tool calls is indistinguishable from ignoring the template.

**What could you have done better?**

1. Announce phase transitions ("research for sizing → now the plan").
2. Reconcile count discrepancies the moment they appear (one more
   drill with `-v` would likely explain lychee).
3. Push the parked docs commit earlier — the 60-min bar was met at
   ~07:37; the push happened at 09:37 only because the session was
   parked. Nothing was lost (docs-only), but the bar exists to be
   used, not admired.

**What could you still improve?** Session-boundary TODO_LIST sync;
research artifacts written to the repo immediately; the plan doc
(below) is still unwritten — it is the first execution task.

**Did you lie?** No new claims beyond evidence; the 45→46 correction
from last report is committed and pushed (`f3bf706`).

**Verschlimmbessern risk watch:** the fix-all mandate's biggest danger
is "fixing" intentional design into worse code (splitting the i18n
dictionary for a line count, adopting `lo` against repo doctrine,
migrating the passkey identity seam against a standing ruling). The
plan therefore routes policy-shaped items to questions, not code.

---

## The 13-tool classification (the window's product)

Verdict classes: **REAL** (fix code) · **FP/RULE** (per-site
suppression with reason) · **NOISE-TOOL** (tool/config defect — fix
config or upstream) · **POLICY** (owner-shaped disposition).

| Tool                | Count   | Verdict                      | Substance                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ------------------- | ------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~bandit-check~~    | ~~335~~ | ~~**NOISE-TOOL + FP/RULE**~~ | ~~~329 = BuildFlow provider swallowing bandit's stdout (banner/progress lines counted as findings) → upstream BuildFlow issue; ~6 real rows (B101 asserts, B105/B107 dev secrets, B108 tmp, B306 mktemp, B310 urlopen, B404/B603/B607 subprocess) in scripts/tests — nosec-with-reason class~~ done — the real rows landed as nosec-with-reason, reasons on every marker (scripts/perf-baseline.py:23 `# nosec B310 - dev perf harness…`, scripts/render-diff.py:84 `# nosec B105 - the documented dev-only secret…`, scripts/ui-capture.py:717 `# nosec B108 - LOCAL-ONLY visual harness…`; grep-verified this session) and `.bandit` documents the drill-script exclusion; the provider-stdout defect is a minted issue draft (docs/drafts/2026-10-11_buildflow-bandit-stdout-findings.md) awaiting owner filing per plan T2 |
| ~~govulncheck~~     | ~~94~~  | ~~**NOISE-TOOL**~~           | ~~87 "requires newer Go version go1.27" (binary built go1.26, cannot load the module — zero real scans) + 3 parse crumbs → `tool_paths` pin to a go1.27-built binary~~ done — pin landed: `.buildflow.yml` lines 63-64 `tool_paths: govulncheck: /home/lars/go/bin` carry the go1.27-built-binary rationale comment ("94 \"requires newer Go version\" noise rows and ZERO real scans (2026-10-11)", read-verified)                                                                                                                                                                                                                                                                                                                                                                                                            |
| ~~art-dupl~~        | ~~90~~  | ~~**POLICY→REAL**~~          | ~~the clone groups; registry-ratified `-t 1` sweep + `art-dupl baseline` file + CI `art-dupl check` (TODO_LIST sitting item 4 — never built)~~ done — ritual built: `.art-dupl-baseline.json` exists (repo root, 24.8 KB) and CI runs the new-only `dup-gate` job (`art-dupl check . -t 1 --baseline-path .art-dupl-baseline.json`, .github/workflows/ci.yml:145-160, read-verified); TODO_LIST § Dedup/branching-flow sitting: "art-dupl `-t 1` + baseline ritual — EXECUTED 2026-10-12: first deep sweep 867→58 actionable groups, ZERO harmful"                                                                                                                                                                                                                                                                             |
| ~~file-size-check~~ | ~~28~~  | ~~**POLICY**~~               | ~~real list (i18n.go 424, webhooks_test 418, middleware_test 351, …) — is 350 a real limit here? i18n is the one-home dictionary~~ superseded by docs/planning/2026-11-07_21-30_owner-sitting-pack.md §D10 ("`file-size-check`'s hardcoded 350-line cap (plan posture call 2)", recommend: push upstream for a configurable cap) — meanwhile the 46 per-finding accepts live in `.buildflow-baseline.json` (ids `file-size-check:go-file-too-large:…`, grep-verified)                                                                                                                                                                                                                                                                                                                                                          |
| ~~vulnix~~          | ~~21~~  | ~~**POLICY**~~               | ~~real advisories, but the release gate (`nix run .#vulnix` + `webphone-vulnix-triage`) owns verdicts — buildflow step redundancy question~~ done — release-gate ownership is the standing ruling: AGENTS.md quality-gates block "release.sh also gates on `nix run .#vulnix`" + ROADMAP.md "`nix run .#vulnix` rides the release.sh gates every train since 2026-09-20"; verdict logic is the fixture-checked `webphone-vulnix-triage` CLI (nix/apps.nix reads "Verdict logic lives in the shared CLI so the fixture check (checks.vulnix-triage) exercises the same code"; CI fast-checks builds `.#checks.x86_64-linux.vulnix-triage`)                                                                                                                                                                                      |
| ~~go-auto-upgrade~~ | ~~19~~  | ~~**POLICY**~~               | ~~11 lo.Map/Reduce advisories (3 note lo is not even a dependency), ~5 jsonv1-boundary advisories (the deliberate mixed-jsonv2 state) → budget/skip-with-rationale class~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 20 ("go-auto-upgrade: add `warnings_budget` marker + rationale comment (lo/jsonv1 declined per doctrine)") — still open there: no `warnings_budget` anywhere in `.buildflow.yml` (grep-verified)                                                                                                                                                                                                                                                                                                                                                               |
| ~~golangci-lint~~   | ~~18~~  | ~~**REAL + QUESTION**~~      | ~~errcheck (rows.Close, testDB.Close — real) + SA1019 usermgmt deprecations; `identity-model/v4` exists in vendor, usermgmt/v4 used in exactly 5 files → bounded migration, but the 2026-10-05 ruling says "userauth rides usermgmt.UserID" → owner gate (g/2)~~ done — errcheck fixed: `listRows` in internal/store/db.go:380-401 is the ONE home carrying `defer func() { _ = rows.Close() }()` + `rows.Err()` (read-verified), and internal/server/server_test.go:120 carries `t.Cleanup(func() { _ = db.Close() })`; SA1019 landed: internal/userauth/userauth.go:25 and enroll.go:12 import `cqrs-htmx/identity-model/v4` directly, CHANGELOG [2.9.0]: "every deprecated `usermgmt` re-export (user ids, sentinel errors, the WebAuthn provider seam) is gone from webphone code"                                         |
| ~~jscpd~~           | ~~6~~   | ~~**FP/RULE**~~              | ~~island-tests helper similarity (stub wiring) — intentional test-fixture class, attribute against registry~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 22 ("jscpd: attribute the 6 island-test clone rows (test-fixture class; registry note)") — still open there: no jscpd/test-fixture ruling exists in docs/dedup-registry.md (grep-verified)                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ~~lychee~~          | ~~5→0~~ | ~~**INVESTIGATE**~~          | ~~full-run said 5, isolated drill says 0 — unexplained, reconcile~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 24 ("lychee: rerun with -v; reconcile the 5-vs-0 delta"); the standing authenticate-vs-exclude policy waits in the owner sitting pack §F9 ("exclude archived dirs (the buildflow warn names the exact config)")                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ~~nix-checker~~     | ~~2~~   | ~~**POLICY (minor)**~~       | ~~vendorHash-to-dedicated-file style advice~~ done — moot by current shape: vendorHash already lives in the dedicated nix/packages.nix (line 144, `vendorHash = "sha256-NzY62…"`) and nowhere in flake.nix (grep-verified) — the advised dedicated-file shape holds                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| ~~nix-flake-check~~ | ~~2~~   | ~~**REAL (trivial)**~~       | ~~`apps.*.vulnix` lacks `meta.description`; aarch64-omitted warning~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 23 ("nix: add `meta.description` to the vulnix app") — still open there: nix/apps.nix `apps.vulnix` returns only `{ type = "app"; program = … }` with no meta (read-verified)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ~~vulture~~         | ~~3~~   | ~~**FP/RULE + REAL (1)**~~   | ~~`do_POST`/`log_message` are http.server overrides (FP); `binary_location` unused attr likely real~~ done — FP verified in the current tree: `options.binary_location = chrome` (scripts/ui-capture.py:420) is consumed by `webdriver.Chrome(service=service, options=options)` (ui-capture.py:434, read-verified), so the "unused attr" is a false positive with nothing to remove; `log_message` keeps its reason in-code ("# Silence: the harness's stdout is the shot ledger", ui-capture.py:269) and `do_POST` is the stdlib dispatch hook                                                                                                                                                                                                                                                                               |
| ~~branching-flow~~  | ~~19~~  | ~~DONE~~                     | ~~baselined (earlier this session)~~ done — UNSKIPPED 2026-10-11 with the 19-row `.buildflow-baseline.json` (skip_steps now holds only pytest-test, .buildflow.yml:171-172 read-verified); TODO_LIST § Dedup/branching-flow sitting: "Re-triage the branching-flow skip — EXECUTED 2026-10-11 (stale premise retired): the tool runs UNSKIPPED (19 baselined findings with in-source rationales)"                                                                                                                                                                                                                                                                                                                                                                                                                              |

Net: **~42 genuinely actionable rows** hiding inside 642 displayed
findings; three config/tool defects (bandit provider, govulncheck pin,
missing art-dupl baseline) account for ~510 of the display count.

## a) FULLY DONE (this window)

1. All 13 tools drilled and classified (table above) — evidence:
   `/tmp/f-{tool}.json` + this report (now the durable home).
2. Migration surface bounded: `usermgmt/v4` usage = 5 files
   (internal/userauth ×3, internal/server/passkey* ×2);
   `identity-model/v4` present in vendor.
3. TODO_LIST + CHANGELOG current-state read (staleness identified —
   see b).
4. Skills loaded for the train (pareto-planning, docs-health; the
   earlier buildflow/dedup/github-voice set still in session context).
5. Parked docs commit pushed under the bar (`f3bf706`, anchor
   06:37:25, pushed 09:37).

## b) PARTIALLY DONE

1. **Fix-all mandate** — research DONE (above); Pareto tiers sized in
   my analysis but the two TABLES, the plan doc, and ALL execution are
   pending. Blocker: none (interrupted).
2. **CHANGELOG currency** — mandate received; the `[Unreleased]` lines
   not yet written. Effort S.
3. **TODO_LIST sync** — staleness identified (items 2b/3 done, item 4
   open, plus the new tool-disposition tasks to harvest); not edited.

## c) NOT STARTED

Everything in the execution plan (detailed in f): bandit nosecs +
BuildFlow provider issue; govulncheck pin; art-dupl baseline ritual;
file-size policy; vulnix ownership decision; golangci errcheck fixes;
SA1019 migration (gated by g/2); jscpd/vulture/nix fixes; lychee
reconciliation; go-auto-upgrade disposition; drift alarm; full
`nix flake check`; whitespace-drift + pre-commit exercise; 02-19 doc
formatting fix; `.buildflow.yml` UNSKIPPED-note dedup; drafts
`-body.md` cleanup; smoke re-run; go-error-family peek; CI verdict for
`f3bf706`.

## d) TOTALLY FUCKED UP

Nothing broke this window. Honest failures, all process-class:

1. Research parked in /tmp for 3 h (single point of loss — corrected
   by this report).
2. Fix-on-sight slipped AGAIN (TODO_LIST staleness) — second
   occurrence of the same class in one session.
3. The plan doc demanded by the template was never begun before the
   park; had the mandate's author checked, zero visible progress
   existed despite completed research.

## e) WHAT WE SHOULD IMPROVE

1. **Persist research immediately** — findings classifications belong
   in the repo the minute they exist (this report is the pattern).
2. **Announce phases** — "research → plan → execute" checkpoints make
   template compliance visible instead of inferred.
3. **Session-boundary TODO_LIST sync** — cheap, noticed twice, done
   never; make it a hard habit.
4. **Count-discrepancy reflex** — when two runs disagree (lychee
   5 vs 0), resolve before moving on; unexplained deltas compound.
5. **Noise hygiene upstream** — two of the three big noise classes are
   BuildFlow/branching-flow defects (bandit stdout-swallowing, the
   earlier ID divergence): file both upstream rather than nursing
   their output locally forever.

## f) Next tasks (ranked; feeds HARVEST — supersedes the 06-34 list where overlapping)

| #      | Task                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Impact     | Effort  | Cat               |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------- | ----------------- |
| ~~1~~  | ~~Write the SUPERB plan doc (Pareto tiers, both tables, mermaid graph) at docs/planning/~~ done — verified present: docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md carries the Pareto tiers (1%→51% / 4% / 20% / other), the T1–T11 comprehensive TABLE, the 37-microtask TABLE and the mermaid execution graph (read-verified) — and its header states it "Supersedes the execution lists of docs/status/2026-10-11_09-38_* §f"                                                                                                                                                  | ~~High~~   | ~~S~~   | ~~Documentation~~ |
| ~~2~~  | ~~CHANGELOG `[Unreleased]`: needle rename + gate rewiring + this window's fixes~~ done — the bar got teeth: ci.yml's `changelog-guard` job ("Require an `[Unreleased]` line for product-code changes", comment "The CHANGELOG bar (RATIFIED 2026-10-11)") plus scripts/changelog-guard.sh (both read-verified); this window's fixes carry [2.9.0] lines — server.EventName, the dup-gate job, scripts/push-stall-watch.sh, the .githooks commit-msg hook (CHANGELOG.md:1150/1210-1219, read-verified). Caveat: no line found naming the `needle` rename itself (grep-verified CHANGELOG)     | ~~High~~   | ~~S~~   | ~~Documentation~~ |
| ~~3~~  | ~~TODO_LIST sync: strike done items 2b/3, add tool-disposition tasks~~ done — TODO_LIST was restructured per its docs-health convention "Done work is DELETED, never struck through" (TODO_LIST.md:3, read-verified): the stale 2b/3 numbering is gone and the tool-disposition tasks landed as TODO_LIST § "Dedup/branching-flow sitting (2026-10-11 ratifications — execution tail)" with EXECUTED verdicts (rows 3 "Re-triage the branching-flow skip — EXECUTED 2026-10-11" and 4 "art-dupl `-t 1` + baseline ritual — EXECUTED 2026-10-12", quoted from the file)                       | ~~High~~   | ~~S~~   | ~~Documentation~~ |
| ~~4~~  | ~~govulncheck: pin a go1.27-built binary via `tool_paths` (kills 94 noise rows, restores real scanning)~~ done — verified in .buildflow.yml:63-64: `tool_paths:` / `govulncheck: /home/lars/go/bin`, with the rationale comment citing the exact 94-noise diagnosis ("The ~/go/bin copy is go1.27.2-built")                                                                                                                                                                                                                                                                                  | ~~High~~   | ~~S~~   | ~~Quality~~       |
| ~~5~~  | ~~bandit: nosec-with-reason the ~6 real rows (dev secrets/asserts/subprocess in scripts)~~ done — nosec-with-reason markers landed with a reason on each: scripts/perf-baseline.py:23 `# nosec B310 - dev perf harness fetching the local server under test`, scripts/render-diff.py:36 `# nosec B404 - deliberately spawns the local webphone binary under test`, scripts/ui-capture.py:717 `# nosec B108 - LOCAL-ONLY visual harness by budget decision (AGENTS T23)` (grep-verified this session; `.bandit` excludes the drill separately)                                                | ~~High~~   | ~~S~~   | ~~Quality~~       |
| ~~6~~  | ~~File BuildFlow issue: bandit provider parses stdout as findings (banner/progress included); should run `-f json`~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T2/microtask 8 — the issue is DRAFTED, not filed: docs/drafts/2026-10-11_buildflow-bandit-stdout-findings.md (+ -body.md, evidence `tools/providers/python_tools.go:45-52`) carries an unchecked "MANUALLY REVIEWED by @Lars Artmann" owner banner (read-verified) — still open with the owner                                                                                                  | ~~High~~   | ~~S~~   | ~~Quality~~       |
| ~~7~~  | ~~golangci errcheck: fix rows.Close/testDB.Close per repo idiom~~ done — verified in tree: internal/store/db.go `listRows` (lines 380-401) is the ONE home with `defer func() { _ = rows.Close() }()` + the `rows.Err()` check, and internal/server/server_test.go:120 carries `t.Cleanup(func() { _ = db.Close() })`; AGENTS.md one-home-helpers bullet documents the listRows "query→close→scan→`rows.Err()` lifecycle"                                                                                                                                                                    | ~~High~~   | ~~S~~   | ~~Bug~~           |
| ~~8~~  | ~~art-dupl ratified ritual: `-t 1` sweep, attribute vs registry, `art-dupl baseline` file, CI `art-dupl check`~~ done — all four pieces verified: `.art-dupl-baseline.json` exists (repo root), CI runs `art-dupl check . -t 1 --baseline-path .art-dupl-baseline.json` in the `dup-gate` job (.github/workflows/ci.yml:145-160), the registry protocol step 1 codifies the sweep (docs/dedup-registry.md), and TODO_LIST § Dedup/branching-flow sitting row 4: "EXECUTED 2026-10-12: first deep sweep 867→58 actionable groups, ZERO harmful"                                               | ~~High~~   | ~~M~~   | ~~Quality~~       |
| ~~9~~  | ~~SA1019: migrate usermgmt→identity-model in 5 files IF g/2 approves (auth: prefix + session suite per convention)~~ done — the migration LANDED: internal/userauth/userauth.go:25 and enroll.go:12 import `cqrs-htmx/identity-model/v4` directly (grep-verified), CHANGELOG [2.9.0]: "The identity layer imports `cqrs-htmx/identity-model/v4` directly: every deprecated `usermgmt` re-export … is gone from webphone code", and AGENTS.md cqrs-htmx audit: "the identity train (usermgmt v4.16.0 + identity-model v4.14.0 + webauthn v4.13.0) is fully LANDED"                            | ~~High~~   | ~~M~~   | ~~Quality~~       |
| ~~10~~ | ~~vulnix: resolve gate ownership (g/3) then skip-or-wire the buildflow step with rationale~~ done — g/3 resolved to release-gate ownership, recorded durably: AGENTS.md quality-gates "release.sh also gates on `nix run .#vulnix`" + ROADMAP.md "`nix run .#vulnix` rides the release.sh gates every train since 2026-09-20"; the verdict logic is the fixture-checked `webphone-vulnix-triage` CLI (nix/apps.nix "Verdict logic lives in the shared CLI so the fixture check (checks.vulnix-triage) exercises the same code"; CI fast-checks builds `.#checks.x86_64-linux.vulnix-triage`) | ~~High~~   | ~~S~~   | ~~Quality~~       |
| ~~11~~ | ~~file-size: apply g/1 decision (enforce-and-split vs tuned limit/excludes-with-reasons)~~ superseded by docs/planning/2026-11-07_21-30_owner-sitting-pack.md §D10 — "`file-size-check`'s hardcoded 350-line cap (plan posture call 2)… Recommend: push upstream for a configurable cap, and keep the 46 accepts meanwhile" (the sitting pack is TODO_LIST's ONE answer surface); the 46 per-finding accepts are in `.buildflow-baseline.json` (grep-verified)                                                                                                                               | ~~Medium~~ | ~~S/M~~ | ~~Quality~~       |
| ~~12~~ | ~~go-auto-upgrade: disposition (budget marker or skip-with-rationale; lo-adoption contradicts root-library doctrine)~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 20 ("add `warnings_budget` marker + rationale comment (lo/jsonv1 declined per doctrine)") — still open there: no `warnings_budget` exists in `.buildflow.yml` (grep-verified)                                                                                                                                                                                                    | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~13~~ | ~~vulture: remove `binary_location` if dead; nosec/attribute the http.server override FPs~~ done — verified not-dead, so nothing to remove: `options.binary_location = chrome` (scripts/ui-capture.py:420) feeds `webdriver.Chrome(service=service, options=options)` (ui-capture.py:434, read-verified) — the "unused attr" report is a false positive; the override FPs are attributed in-code: `log_message` carries "# Silence: the harness's stdout is the shot ledger, not access logs" (ui-capture.py:268-270) and `do_POST` is the http.server dispatch hook                         | ~~Medium~~ | ~~S~~   | ~~Cleanup~~       |
| ~~14~~ | ~~jscpd: attribute the 6 island-test clones against the registry's test-fixture class~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 22 — still open there: no jscpd/test-fixture attribution exists in docs/dedup-registry.md (grep-verified)                                                                                                                                                                                                                                                                                                       | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~15~~ | ~~lychee: reconcile 5-vs-0 (rerun with -v; network dependence?)~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 24 ("lychee: rerun with -v; reconcile the 5-vs-0 delta; fix real dead links if any") — still open there; the standing authenticate-vs-exclude policy sits in the owner sitting pack §F9                                                                                                                                                                                                                                               | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~16~~ | ~~nix-flake-check: add `meta.description` to the vulnix app; assess aarch64 warning~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T6/microtask 23 — still open there: nix/apps.nix `apps.vulnix` returns only `{ type = "app"; program = … }` with no meta attribute (read-verified this session)                                                                                                                                                                                                                                                                | ~~Medium~~ | ~~S~~   | ~~Cleanup~~       |
| ~~17~~ | ~~nix-checker: vendorHash-to-file advice — adopt or decline with reason~~ done — the advised shape already holds, verified: vendorHash lives in the dedicated nix/packages.nix (line 144) and flake.nix carries no vendorHash (grep-verified)                                                                                                                                                                                                                                                                                                                                                | ~~Low~~    | ~~S~~   | ~~Cleanup~~       |
| ~~18~~ | ~~Drift alarm: assert branching-flow findings == baseline count in CI~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T7/microtask 27 — still open there; the recorded constraint is `.buildflow.yml`'s baseline caveat ("findings stay under the severity cap everywhere, so a non-matching baseline warns, never gates", lines 48-53, read-verified) and CI instead gained the analogous `dup-gate` new-only assertion for art-dupl (ci.yml:145-160)                                                                                                             | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~19~~ | ~~Full `nix flake check` (owed since the re-triage delta)~~ done — later full runs recorded green: TODO_LIST § 2026-11-07 frontend-UX residue "train-end 74/74 (run 20261107-084827), `nix flake check` all-checks-passed" (quoted from the file), and docs/ci-flake-ledger.md's 2026-10-12 row 38282396056 records "its own full go suite, buildflow rc=0, and `nix flake check` all-pass ran locally pre-push"                                                                                                                                                                             | ~~Medium~~ | ~~M~~   | ~~Quality~~       |
| ~~20~~ | ~~Pre-commit hook exercise under the new baseline config~~ done — hooks landed and exercised: `.githooks/` carries commit-msg + pre-commit (ls-verified), and TODO_LIST § unified-auth tails: "daemon-enforcement hook LANDED 2026-10-12 … hook logic itself VERIFIED (temp-index repro refuses exit 1; `scripts/commit-msg-guard-selftest.sh` pins 3 cases)" — both self-test scripts exist in scripts/ (ls-verified)                                                                                                                                                                       | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~21~~ | ~~02-19 status doc formatting fix (third notice)~~ done — superseded by archive: the 02-19 doc was archived 2026-11-07 in the docs-health v10 sweep with ledger-marked, well-formed tables (docs/status/archived/2026-10-11_02-19_branching-flow-report-triage-status.md header + table read-verified this session)                                                                                                                                                                                                                                                                          | ~~Low~~    | ~~S~~   | ~~Documentation~~ |
| ~~22~~ | ~~`.buildflow.yml` UNSKIPPED-note/baseline-block narrative dedup~~ done — verified disjoint in the current file: the branching-flow note (lines 89-107: unskip history + suppression mechanics) and the baseline block (lines 42-53: 2026-10-28 re-establishment + the CI host-absolute-ids caveat) share no duplicated narrative (read-verified this session)                                                                                                                                                                                                                               | ~~Low~~    | ~~S~~   | ~~Cleanup~~       |
| ~~23~~ | ~~drafts `-body.md` artifacts: keep-or-gitignore decision~~ done — keep won by usage: docs/drafts/ is the tracked evidence home — TODO_LIST § Dedup/branching-flow sitting: "Drafts land in `docs/drafts/`" (quoted) and the owner sitting pack cites ../drafts/ files in its evidence table ("[pma push-leg durability draft](../../drafts/pma-push-leg-durability-issue-2026-10-28.md)", read-verified); nothing in .gitignore excludes them                                                                                                                                                  | ~~Low~~    | ~~S~~   | ~~Cleanup~~       |
| ~~24~~ | ~~smoke re-run post-rename~~ done — the smoke is now a standing CI job, not a one-off: ci.yml `smoke` ("webphone-smoke (live boot)") builds the binary and runs scripts/webphone-smoke.py against it on every push (read-verified)                                                                                                                                                                                                                                                                                                                                                           | ~~Low~~    | ~~S~~   | ~~Quality~~       |
| ~~25~~ | ~~go-error-family peek: confirm nothing actionable for the error_rules wait-state; note result~~ done — the library got its full audit instead of a peek: docs/review/libs/go-error-family.md (2026-11-04, adoption 64/100) rules the gaps — Conflict family LANDED, retry consumption rides go-retry (ADOPTED 2026-11-04 per AGENTS.md: `chatloop.Backoff` gates `errorfamily.Classify`), error context "do not adopt ceremonially" — nothing actionable for the error_rules wait-state, whose block stays parked in `.buildflow.yml` lines 35-45 (read-verified)                           | ~~Low~~    | ~~S~~   | ~~Quality~~       |
| ~~26~~ | ~~Upstream sweep-extension comment on branching-flow#12 for the 8 unsuppressed analyzers~~ done — harvested: TODO_LIST § Dedup/branching-flow sitting, row "branching-flow upstream filings (owner: file)" — the suppression-extension draft is minted at docs/drafts/2026-10-11_branching-flow-composition-dupe-suppression.md ("composition + dupe analyzers never consult ignore comments … Ask: wire both analyzers into `core.IgnoreComments`", read-verified), owner filing pending                                                                                                    | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~27~~ | ~~Owner: review + tick banners on #12/#13~~ done — harvested: TODO_LIST § Dedup/branching-flow sitting, row "branching-flow upstream filings (owner: file)" — the drafts carry the owner-review banner "[ ] MANUALLY REVIEWED by `@Lars Artmann`" (docs/drafts/2026-10-11_branching-flow-*.md, read-verified); the 10-47 plan's T11 also lists the #12/#13 banner ticks as owner-gated                                                                                                                                                                                                       | ~~High~~   | ~~S~~   | ~~Documentation~~ |
| ~~28~~ | ~~When error_rules ships: apply the parked block~~ superseded by docs/planning/2026-10-11_10-47_SUPERB-findings-zeroing.md T11 ("error_rules application post-release", owner-gated) — the parked ready-to-paste block is verified present at `.buildflow.yml` lines 35-45 ("GATE COMPLETION (pending tooling): the installed BuildFlow (25b07c4) caps SDK detector findings at warning severity")                                                                                                                                                                                           | ~~Medium~~ | ~~S~~   | ~~Quality~~       |
| ~~29~~ | ~~Quarterly audit 2027-01-11 → TODO_LIST standing row~~ done — harvested: TODO_LIST § Dedup/branching-flow sitting, row "Quarterly baseline/suppressed-set audit — first 2027-01-11 (registry Protocol step 1)" (quoted from the file, line 820)                                                                                                                                                                                                                                                                                                                                             | ~~Medium~~ | ~~S~~   | ~~Documentation~~ |
| ~~30~~ | ~~Daemon push-stall watchdog (the 60-min bar applied automatically)~~ done — scripts/push-stall-watch.sh exists (ls-verified) and CHANGELOG [2.9.0]: "exits non-zero when the oldest unpushed commit is older than the threshold (default 60 minutes, the ratified manual-push anchor)"; TODO_LIST § unified-auth tails: "LANDED 2026-10-12 … 4 arms self-tested in a scratch repo"                                                                                                                                                                                                          | ~~Medium~~ | ~~M~~   | ~~Quality~~       |
| ~~31~~ | ~~CI verdict for `f3bf706` (docs-only push)~~ done — verified green via the GitHub API this session: check-runs on f3bf706 all `conclusion: success` (nix build .#webphone, fast flake checks, go test -race)                                                                                                                                                                                                                                                                                                                                                                                | ~~Low~~    | ~~S~~   | ~~Quality~~       |
| ~~32~~ | ~~settingsRow adoption train (carried, pre-existing)~~ done — ADOPTED and executed: internal/web/views/settings.templ:105,111 render `@display.DefinitionList(display.DefinitionListProps{…})` (read-verified); TODO_LIST § Dedup/branching-flow sitting: "settingsRow ADOPTION (owner: BUILD → resolved to ADOPT) — EXECUTED 2026-10-12", and CHANGELOG [2.9.0] records the DefinitionList swap with the 6/7-panels byte-identical diff                                                                                                                                                     | ~~Medium~~ | ~~M~~   | ~~Feature~~       |

## g) Questions I cannot answer myself

1. **file-size 350-line limit**: enforce for real (split `i18n.go`'s
   dictionary, webhooks_test, …) or is that a BuildFlow default to tune
   with reasoned excludes (the i18n dictionary is a deliberate one-home
   blob; splitting it is churn with no behavior value)? I can execute
   either, but the limit's authority is yours.
2. **SA1019 usermgmt → identity-model migration**: the 2026-10-05
   strong-id ruling says "userauth rides `usermgmt.UserID`" — but
   upstream now deprecates the whole usermgmt surface in favor of
   `identity-model/v4` (in vendor; 5 files affected). Does the ruling
   FORBID the direct import, or was it about the ID type only (making
   the migration a conforming refinement)? It touches the passkey
   seam, so I won't move without your read.
3. **vulnix gate ownership**: the buildflow-embedded vulnix reports 21
   advisories ungated, while the release runbook's `nix run .#vulnix`
   - `webphone-vulnix-triage` CLI is the fixture-checked verdict path.
     Should the buildflow step be skipped as documented redundancy (the
     release gate owns it), or should its output be wired through the
     same triage verdict? Which gate is canonical is a runbook call.

— End of report. Waiting for instructions.
