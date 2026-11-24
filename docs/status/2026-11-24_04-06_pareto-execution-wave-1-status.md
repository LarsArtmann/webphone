# Pareto plan execution — wave 1 status (the 2026-11-24 run)

> **Claims verified as of 04:06 CET 2026-11-24.** Push/CI statements rot
> fastest; everything below was re-checked at that minute unless a command
> output is quoted.
>
> **Created:** 2026-11-24 04:06 CET · the pareto-execution session · backlog
> basis: TODO_LIST @ `586ff894` (this session's own tip).

**Scope.** This report covers ONE thing: executing
`docs/planning/2026-11-24_01-24_SUPERB-post-deploy-pareto-replan.md` (27
tasks) after the owner's "get shit done" order. VERIFIED: every landed commit
below is on local `main` and was pushed at 04:04 (`94157138..586ff894`, manual
push, see §d-6); each landing's evidence is the agent's gate/self-test output
re-quoted here, plus the docs battery run in a shared clone for the living-doc
batch (PASS: audit-all 74/74, annotate 254/260). NOT VERIFIED: CI green on
`586ff894` (run was `in_progress` at 04:05, `gh run list --workflow ci.yml`);
the six interrupted agents' end states (no reports received — §b); the
cross-repo clones (stack/android) whose work sits unlanded and unpushed;
subagent report claims beyond the spot-checks named inline.

**Run context (one paragraph, the rest is sections).** Nine agents were
dispatched at once; the provider's rate limit killed six at birth and three
more later (T08 three times). Two fallback providers were tried and have NO
credits (hyper, huggingface/fireworks: "payment required"); `synthetic` works
and carried T03 and T09 to completion. A host DNS outage (~02:55–03:05,
`api.github.com`/`api.synthetic.new` unresolvable, confirmed by a python
`socket.gethostbyname` probe at 03:01) killed every synthetic agent mid-run;
a Crush process exit additionally interrupted T16. DNS recovered by 04:04.
Six commits landed; six tasks sit interrupted at known states (§b).

## a) Fully done

| # | What                                                                        | Commit     | Evidence                                                                                                                                                                                                                                                                                                                                               |
| - | --------------------------------------------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | T04 `scripts/ci-verdict.sh` pages past the watch-lane flood                 | `362956ab` | Self-test 11/11 incl. the new page-2 fixtures (proven failing on the old script: rc 1 masquerade); live: `10327ed8` now rc 2 RED, `3f91cf61`/`7f39d6a7`/`94157138` rc 0 GREEN; PAGE_CAP=20 answers greppable `PAGE-CAP-EXCEEDED` rc 3                                                                                                                  |
| 2 | T10 `scripts/citation-sweep.py` + cite hygiene                              | `e81781fb` | Sweep before: 16 bad cites (2 OUT-OF-RANGE, 8 MISSING, 4 AMBIGUOUS, 2 UNKNOWN) — after: `182 cite(s): OK 182`; self-test sabotage-proven; rides as the advisory tail of `scripts/audit-verify.sh` (124 commands: 69 PASS + 55 RUN-OK, exit contract unchanged); claim rot fixed (five `blob.Save` sites, `.wp-error` pins 391→422 / 206→208 / 251→173) |
| 3 | T03 release closure: CHANGELOG deploy ledger + FEATURES stamp + prod probes | `a4efbfc3` | Live anonymous read-only probes 6/6 (`/version` v2.15.0 @ `75913f1a`, `/login`, manifest, openapi 3.0.3, `/healthz` 5 checks ok, styled 404); `webphone-smoke --base` deliberately NOT run (its foreign arms POST); FEATURES row corrected (OpenAPI 3.0.3); freshness pre-flight SAFE; battery PASS                                                    |
| 4 | T09 `scripts/fleet-tags.sh` no-v headings + tally line + re-sweep           | `78a10215` | Self-test 6/6 (new fixture proven silent-skip on the old script); coverage: `swept 24/24 repos (53 CHANGELOGs checked, 0 skipped)`; ZERO missing tags — the one UNTAGGED (go-cqrs-lite root v4.14.1) is a release-train heading artifact; 6 vendored-CHANGELOG false findings excluded                                                                 |
| 5 | T01 owner sitting pack refreshed + the top-12 sheet                         | `4de5cf43` | 53/58 lettered rulings open (5 ticked with evidence, D10 premise gone), 29/38 F, 7/8 G; top-12 = 162 owner minutes; battery PASS (74/74, 254/260)                                                                                                                                                                                                      |
| 6 | Living-doc batch for the three script fixes                                 | `586ff894` | AGENTS: blind-spot-closed sentence, fleet-tags contract, citation-sweep Commands line, 391→422 pin, five Save sites, `GO-2026-6859 / CVE-2026-97054`; CHANGELOG 3 bullets; TODO registry sweep-log line. First draft pre-flighted UNSAFE (3 historical items would flip), reworded → SAFE; battery PASS                                                |
| 7 | Manual push after the daemon stall (bar invoked)                            | —          | Anchor: first unpushed commit `362956ab` @ 02:45:44; pushed 04:04 (79 min > 60-min bar); `94157138..586ff894` on origin                                                                                                                                                                                                                                |

T01's top-12 sheet is the owner-facing deliverable of the whole 1% tier: it
lives at the top of `docs/planning/2026-11-07_21-30_owner-sitting-pack.md`
("Top 12 for the next sitting (prepared 2026-11-24)").

## b) Partially done

| Task                    | State at 04:06                                                                                                                    | Remaining                                                                                                                                                                  |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T06 backlog truth       | Agent gunnar: clone `/tmp/wp-clone-t06` exists at base `94157138`, clean; killed twice (provider rate limit, then the DNS outage) | Everything: re-verify §5.2's 16 rows at HEAD, strike, dedupe twins, battery                                                                                                |
| T07 own-defects (index) | Agent gustavo: killed at birth twice (rate limit, credits, DNS); no work done                                                     | index.html: vuln id ×2 cells, go-health/go-error-family row, appendix block, Findings cards 3–5; NOTE: the AGENTS vuln sentence is already fixed by the batch (`586ff894`) |
| T08 deps train          | Agent mateus: killed 3× (rate limit) + 1× (DNS); no work done                                                                     | All of m08.1–m08.8; must run on `synthetic` with one-command pacing                                                                                                        |
| T16 stack window A      | Agent alejandra: interrupted by a Crush process exit; session kept, work state unknown (clone `/tmp/stack-clone-t16` if created)  | Resume via its session; then the two legs + stack gates                                                                                                                    |
| T20 announcements       | Agent geraldo: killed at birth twice (rate limit, credits, DNS); no work done                                                     | Inventory + draft + channel options                                                                                                                                        |
| T21 android slice 1     | Agent chuck: killed twice (rate limit, DNS); no work done                                                                         | All of m21.1–m21.7 + wire fixture (a)                                                                                                                                      |

## c) Not started

T02 (gated: needs the owner's sitting answers on the top-12 sheet), T05
(red-window detection; unblocked by T04+T10 landing), T11/T12 (the two
read-throughs), T13 (index polish + vision QA), T14 (`RegisterStdlibDefaults`
decision), T15 (lessons + SKILLS runbook), T17/T19 (stack slices, after T16),
T18/T26 (the drains), T22/T23 (gated: the owner's D2/D3 GO), T24 (gated: the
capture credentials), T25 (chat-network stack slice), T27 (gated register
sweep).

## d) Totally fucked up

1. **The 9-agent simultaneous dispatch tripped the provider rate limit** —
   six of nine died at birth, and the limit kept killing replacements (T08
   three times, T01 and T21 once each). Wave 1 landed 6 of 13 tasks; most of
   the wall-clock went into failures, not work.
2. **Provider-fallback improvisation burned starts on dead ends**: hyper
   ("no credits yet", charm.land billing) and huggingface/fireworks
   ("purchase pre-paid credits") both fail with payment errors — discovered
   only by launching agents into them. Only `zai` (rate-limited) and
   `synthetic` (works) are usable.
3. **A host DNS outage (~02:55–03:05) killed every synthetic agent mid-run**
   (T06, T07, T08, T20, T21) and broke `gh`; my own session survived on its
   persistent connection, which made the outage look like "only the
   subagents are broken".
4. **My dispatch preamble carried a broken landing recipe** (`git -L=$cl`
   typo; `merge $cl/main` cannot see a shared clone's new ref): T04 and
   T01/T09 each had to re-derive the working form
   (`git -C real fetch <clone> main && merge --ff-only FETCH_HEAD`) on their
   own. Two agents improvised around my bug; nothing landed wrong, but the
   recipe must be fixed once, centrally.
5. **The living-doc batch's first draft pre-flighted UNSAFE**: "coverage" and
   "repo" were one line away from the generic-token bar (27), and three
   historical report items would have flipped to DRIFT. The pre-flight tool
   caught it before landing; the reword is the landed form. A near-miss, not
   damage — but it confirms the batch protocol (clone → pre-flight → battery
   → land) is load-bearing.
6. **Six commits sat unpushed for 79 minutes** (the daemon's push-leg stalled
   through the DNS outage); the manual-push bar was invoked with the anchor
   named in §a-7 and the push landed at 04:04. CI on the pushed tip was still
   `in_progress` at 04:05 — the green verdict is owed, not assumed.

## e) What we should improve

1. **Dispatch cap ≤ 3 concurrent agents**, each instructed to issue one tool
   call per step; the rate limit is the binding constraint, not task size.
2. **Probe provider health before fan-out**: a one-line DNS probe plus one
   throwaway agent on the target provider would have surfaced both the rate
   limit and the missing credits in minutes instead of six failures.
3. **Centralize the landing recipe** in a tiny helper (the fetch + rebase +
   ff-only triple with retry) so every agent stops re-deriving it.
4. **Record the clone-protocol edges in one place**: `audit-verify.sh` inside
   a clone needs the `/tmp/go-appkit` symlink bridge (T10's discovery), and
   `gh` inside a clone needs `GH_REPO=LarsArtmann/webphone` (T04's).
5. **Keep the "cqrs-htmx" living-doc token at its boundary in mind**: the
   batch pushed it to DF 27 (generic at 27); every future TODO/CHANGELOG line
   that mentions it must pre-flight first.
6. **Wave-2 prompts should be written once and reused** (T05, T11–T15,
   T17–T19, T25–T27) instead of composed at dispatch time under failure
   pressure.
7. **Verify CI before claiming green anywhere**: `scripts/ci-verdict.sh` on
   the pushed tips is now trustworthy (T04 fixed it), but nobody has run it
   on `586ff894` yet.

## f) Up to 50 things we should get done next

Routed already: rows 1–18 map one-to-one onto plan tasks T05–T27, whose homes
are the TODO_LIST sections "2026-11-24 fleet-libs re-audit harvest" and
"2026-11-24 report-harvest batch" (no new TODO rows minted — HARVEST is
already satisfied by the plan's §6 ledger; this list re-states the order of
attack).

1. Resume and land T06 (strike the 15 done-unstruck rows + dedupe) — clone and checklist ready.
2. Resume and land T07 (index.html defects: vuln cells, go-health row, appendix, Findings cards).
3. Resume and finish T08 (the dependency train) on synthetic, one command per step; its CHANGELOG bullet is drafted.
4. Resume and land T16 (stack: styled-404 rule in `tests/browser-e2e.py`, dashboard-on VM leg).
5. Resume and land T20 (the consolidated announcements draft v2.9.0 → v2.15.0).
6. Resume and land T21 (android slice 1 + the loopback live-check script).
7. T05 red-window detection: `audit-verify.sh` into the watch lane, tip-unchanged skip, the pma detection-lane ask (four missed windows).
8. T11 read-through A over the 11 first-half fleet-libs reports + provenance lines + the dprint reflow inspection.
9. T12 read-through B over the other 14 reports + latest re-derivation (needs GitHub reachable).
10. T13 index polish: one-line verdict cells, summary rewrite, trace stylesheet, fresh render, a vision-capable QA pass.
11. T14 decide `RegisterStdlibDefaults`: adopt behind boot registration, or write the decline into the go-error-family report.
12. T15 lessons batch (four war stories) + the SKILLS repo runbook additions.
13. T17 stack hardening slice 1 (operator-watched units, secretFile fixture, sipcred vector) — after T16 lands.
14. T18 status-report drain slice 1: the two big 2026-11-10 reports, gates-first, ratchets measured DOWN.
15. T19 stack residues slice 1: the paperless module option + module-check case + deploy.md secret column.
16. T25 chat-network stack slice 1: the signal-cli unit + QR bootstrap + module-check; the homeserver memo.
17. T26 status-report drain slice 2: the eight-report 2026-11-19/22 cohort.
18. T27 gated register sweep + the upstream ticks batch: the go-structure-linter #13 skip fix, the cqrs-lint P012 ask, the go-sse `[0.7.0]` heading note, and the docs-gates / htmx filings awaiting the owner tick.
19. T02 post-sitting conversion pass — the moment the owner answers the top-12 sheet.
20. Verify CI green on `586ff894` (and every later tip) through `scripts/ci-verdict.sh`.
21. Pass B of TODO_LIST: strike the harvest rows for everything landed in this run (T01/T03/T04/T09/T10 + the batch).
22. Pre-flight every future TODO_LIST/CHANGELOG addition (the "cqrs-htmx"/"coverage"/"repo" boundary is live).
23. T22/T23 stay gated behind the owner's one-word GO on sheet row 10; both designs are ready.
24. T24 stays gated on the capture credentials (sheet row 1, asked four times now).

## g) Questions I cannot figure out myself

1. **Provider budget**: `zai` rate-limits under concurrency and both fallback
   providers (hyper, huggingface/fireworks) have no credits on this account.
   Should I run the remaining ~20 tasks serially through `zai` + `synthetic`
   (slow but steady, my default), or do you want to top up credits / raise
   the limit so I can parallelize again?
2. **D2/D3 GO or defer** (top-12 sheet row 10): T22 (veto window) and T23
   (agent liveness) build only behind OFF-by-default flags and wait for your
   security review either way — but the BUILD itself needs your one word.
   GO both, GO one, or defer?
3. **The production capture credentials** (sheet row 1, now asked four
   times): `WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD` into
   this session's environment only, never written to the repo. Without them
   the whole prod-verification window (T24) stays closed. Provide them here,
   or explicitly park T24?
