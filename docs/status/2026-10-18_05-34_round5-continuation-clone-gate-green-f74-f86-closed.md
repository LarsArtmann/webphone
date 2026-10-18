# Round-5 continuation: clone gate judged + fixed, E17/E20–E23 closed, F74 harness bug found

> **Snapshot-rot header (F85 convention, first adoption): claims below were verified as of 2026-10-18 05:34 CEST. Push/CI statements rot fastest — the final CI verdict for this session's pushes lives in `gh run list`, not here.**

## What this session did

Continuation of the round-5 mandate (`docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md`) after the 02-23 report's WAIT state. The owner's "keep going until everything works" answered the three open §g questions in effect: registry-precedent judgment for the clone (Q1), push-when-red bar (Q2, ratified anchor), run the harness now with the eyeball staying owner (Q3).

### 1. CI RED → GREEN: the clone gate

Main was red on `art-dupl check` (runs 39077421278 → 39083279308). Local re-run with the CI-pinned v0.7.2 at `-t 1` showed FIVE new groups at HEAD, all judged before any baseline write:

- **textagent approve/discard prologue** — the standing 2026-10-12 baselining re-exposed by line drift (the phantom-new class). RE-BASELINED, ruling unchanged.
- **asr transport tail ×3 groups** (Do+drain, 401/403 arm, non-200 family arm across `Client`/`GoogleClient`) — EXTRACTED into `asr.providerRoundTrip`: one transport classifier per wire-agnostic concern. The unification also maps the OpenAI wire's 429 to Transient (retryable) like the Google wire always did — pinned by `TestRateLimitedIsTransient`. Registry: EXTRACTED row + call-site-pair ACCEPT row + sweep-log line; baseline re-recorded 66 groups with the gate's exact flags. Pushed under the manual-push bar (main red + verified fix); CI run **39083735059 GREEN**.
- **transcribe timer-drain idiom ×2** — ACCEPTED (stdlib stop-and-drain, two arms of one select); standing row added.

### 2. Owed full gates for the unpushed trains

Registry fresh; island suite 228/228; **buildflow full `BUILDFLOW_NO_RESULT_CACHE=1` exit 0** (twice — once per tree state); **`nix flake check` exit 0**; branching-flow baseline re-saved (47 findings, the expected line-shift churn); smoke on the final tree **47+4+8 passed, 0 failed**.

### 3. E17 gate-recovery stragglers

- **F74 ui-capture: GREEN 18/18 after fixing a real harness bug.** The first run aborted at settings: the 10-08 pin `dd.wp-service-on` cannot match any served markup since the 10-12 settingsRow→DefinitionList adoption moved the class onto the inner span (verified against the served HTML with a live session; the "passing" 10-17 matrix was a MIXED set — 10 stale shots beside 8 fresh ones, exposed by mtime sort). Fixes: the selector to the descendant form `dd .wp-service-on` (copy pin unchanged) + the out-dir now cleans stale shots so a failed run cannot masquerade as a complete matrix. The AGENTS invocation was also stale — nix ≥2.34 parses paren attrpaths (`nixpkgs#python312.withPackages(ps: …)`) as PATHS; the verified PYSEL expression recipe landed in its place. The 18-shot matrix itself is fresh; the EYEBALL remains owner §g1 (image-blind ceiling stands).
- **F75 module-cache janitor: DONE.** `scripts/module-cache-janitor.sh` — report-only default, `--sweep` deletes only zero-byte AND age-floored `.tmp` (an in-flight download is never touched), wedged non-zero downloads reported but never deleted, `--self-test` proves the three classes and the sweep boundaries. Live: 24 stale zero-byte corpses swept from `/mnt/buildcache/go-mod`, re-run converges to 0. AGENTS command line added.
- **F76: ADJUDICATED ALREADY-UPSTREAM, nothing filed.** Verified at source: templ-components issue #27 (the v1.20.0 zero-pseudo-version incident), `docs/release-checklist.md`, and the release-smoke workflow already encode the tagging discipline. A duplicate filing would have been noise — verify-before-filing Gate 5 paid for itself.

### 4. E20–E23 folds

- **F79: ADJUDICATED ALREADY-FILED-COMPLETED.** Both branching-flow drafts have filed, COMPLETED upstream issues: panic FP shapes = #9 (2026-10-11), phantom class scoping = #15/#16 (2026-10-17). Drafts remain in `docs/drafts/`.
- **F80: the go-modularize §f routing audit — all 40 items annotated inline** (done/routed/NOT-DO verdicts with evidence). Un-routed survivors routed: IO-storm investigation + push-stall 30-min warn threshold → ROADMAP open questions; the stack-repo re-open-trigger one-liner → the cross-repo obligations row.
- **F81: carve numbers refreshed** (two methods agree): wiring-only 19 files / 4,056 non-test LOC (under the 20-file trigger); whole tree incl. api/hooks 27 files / 5,204 LOC.
- **F82: metaengine row CLOSED** — G3's release leg verified inside tag v2.9.0 (`0f1031d`); row deleted per docs-health (CHANGELOG 2.9.0 is the home).
- **F83: F019 Volume hints landed** on all five query declarations via one `humanScaleVolume` const (single-tenant fact rates; the planner's 1000/sec default overestimates by three orders). **F026 ADJUDICATED NOT-DO**: `WithPrefetch` exists to cache cursor-paged next pages — no readmodel Scan paginates by cursor at HEAD; reopens only if a surface adopts cursor paging.
- **F84: lessons folds DONE** — self-verifying scripts + verify-before-describing appended to the daemon-race entry; IO-storm scoping folded into the local-suite corollary.
- **F85: snapshot-rot header adopted** — this report carries it (top); AGENTS Conventions line landed.
- **F86: CLOSED (subsumed + owner-routed).** Verified at source: the retry-or-kill wrapper ask is covered by BuildFlow's 10-13 retry-economics train (starvation-floor retry in `pipeline_builder_*.go` + the no-output watchdog — the 4.5h network-wedge class). The fleet GOPROXY fallback chain stays the owner's §g-3 network-posture decision (crush-config side).

### 5. Harvest

TODO_LIST: sweep header updated; metaengine + series-3 rows deleted (done); DDD row → D executed; ASR row hardening/polish/test-debt bullets closed; gate-recovery tail + round-5 tail rows rewritten to the honest remainder. CHANGELOG: [Unreleased] lines for the island transition table (E12), the transport classifier, the Volume hints, and the janitor.

## Remainders (assistant-executable: none this session)

F87 D2 eyeball + d2 cheat-sheet (Q3-gated), markdownlint posture (post-sitting), the ASR #21/#13/#48 owner items, E1–E6/E18/E19/E24–E26 owner terminals. The stack browser E2E obligation (served-markup changes: history download button, call-card clear) rides the next relock.

## Verdict sources

Local: `/tmp/{buildflow-full2,flake-check,ui-capture2,smoke}.log` (exit codes captured, not masked); CI `gh run list` (39083735059 green); `git log origin/main..HEAD`. Point-in-time snapshot — will go stale; §Remainders is HARVEST input for TODO_LIST.md.
