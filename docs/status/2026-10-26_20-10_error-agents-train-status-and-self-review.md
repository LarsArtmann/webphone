# Status — Error-Agents Train: Shipped, Self-Reviewed, and One Real Bug Found By Its Own Machinery

**Snapshot: claims verified as of 2026-10-26 20:12 CET** at HEAD `c793a86`
(push/CI statements rot fastest — every green claim below names its run id).
Scope: this session's three phases — the AI-agent error-management research
scan, the brutal self-review of that scan, and the SUPERB error-agents
Pareto train (all 20 M-tasks / 79 F-tasks), plus the red-main incident the
train's birth surfaced (run 40301017699) and its repair. Format note: the
status-report skill's canonical format is a styled HTML dashboard; the
operator's instruction explicitly named `.md` — honored per the
user-instruction-wins rule, flagged here per the same rule.

---

## a) FULLY DONE

| Work                                                                                                                                                                                                                                                             | Evidence                                                                                                                                                                                                                                                                                                                                          |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **M1–M2 error-context router** (`scripts/error-context.py`): journal line in → bundle out (registry row, seam files first-site-first, family-test pin, lessons/ledger mentions, last-5 commits; `--json`)                                                        | `--self-test` 37/37; README troubleshooting entry; AGENTS Commands line                                                                                                                                                                                                                                                                           |
| **M3–M5 reproduce-first gate** (`scripts/reproduce-first-check.sh` + advisory `reproduce-first` CI job): a fix train's new `*_test.go` diff must FAIL on the merge base                                                                                          | `--self-test` 3/3 (REPRODUCES/SKIP/NO-REPRO scratch repos); `continue-on-error` advisory posture pinned in the job comment                                                                                                                                                                                                                        |
| **M6–M9 the observers** (`watch.yml` */15 cron → `main-red-watch.sh`, `daemon-audit.sh`, new `whitespace-drift.sh --ref` mode)                                                                                                                                   | self-tests 8/8 + 3/3; drift `--ref` verified against a real indentation-shift commit (the e62fe34 class)                                                                                                                                                                                                                                          |
| **M10–M11 bump-preflight** (`scripts/bump-preflight.sh`): proxy zips → capped diff, removed exported symbols, usage-site map                                                                                                                                     | `--self-test` 4/4; verified on two REAL histories (go-health v0.5.0→v0.5.1, go-error-family v0.11.0→v0.12.0 — the latter renders the actual `Unwrap` diff)                                                                                                                                                                                        |
| **M12–M13 incident-bundle** (`scripts/incident-bundle.sh`): traces/runs/commits windows + error-context sweep → drafted bundle with TBD-verdict lessons/CHANGELOG/ledger skeletons                                                                               | `--self-test` 6/6; runbook pointer inside every bundle                                                                                                                                                                                                                                                                                            |
| **M14–M15 error-drill VM test** (`webphone-error-drill`, KVM-gated): designed boot failure renders the 5-part block (`class=data-dir`), exit 1, NRestarts>1; `/hooks` fail-closed 503; wedged blob store → `/healthz` 503 naming `blob-dir`, recovers on unwedge | built green locally in ~20s VM time (run during the session); KVM-gated so consumer CI never runs it                                                                                                                                                                                                                                              |
| **M16–M17 gated sketches** (LLM prose gate; agent dispatch): scan inventory, runtime/cost decision menus, owner-GO trigger criteria — nothing wired                                                                                                              | `docs/drafts/gated-sketches/2026-10-26_*.md`                                                                                                                                                                                                                                                                                                      |
| **M18–M20 docs + verification + closeout**                                                                                                                                                                                                                       | CHANGELOG Added bullet; README one-liner; AGENTS 174/377 with Commands lines + bump-bullet pre-read clause; TODO_LIST flipped to SHIPPED-with-evidence; plan §7 errata annotated; `nix fmt` 0-changed; full `buildflow` rc=0; all 8 fast checks exit 0; red-watch dry-run `NONE conclusion=success`; final push verified `git ls-remote` == local |
| **Red-main incident repair** (run 40301017699): three island test-side breaks from the concurrent multicall train, invisible ~2.5h because six daemon commits landed with zero CI runs                                                                           | fixed same day (233/233 island tests); ledger row for 40301017699 written; successor runs 40311328217 + 40313259454 green                                                                                                                                                                                                                         |
| **Live bug found by the shipped machinery, fixed within the session**: watch.yml's explicit `permissions:` block omitted `actions: read` → `gh run list` 403'd under `set -e` → all four first-night watch runs (17:02–17:53Z) RED                               | diagnosed from triggers+perms (log API is 403 for this token); fixed in `c793a86`, CI 40385454030 green. The watch workflow failing = its DESIGNED semantics ("red means the watcher itself is broken")                                                                                                                                           |

## b) PARTIALLY DONE

1. **Observers are shipped but have ZERO green live runs.** Four runs fired
   on cron cadence and all failed on the permissions bug; the fix is pushed
   (CI green) but the next cron tick (~19:15/19:30 UTC) had not been
   observed green at snapshot time. Until one green run exists, the
   issue-filing path (`gh issue create` with the workflow token) is
   fixture-tested only, never live-proven.
2. **`reproduce-first` has never run on a real PR.** It is
   `pull_request`-triggered; no PR opened since it landed. Local
   self-tests are strong; the CI-side run (nix develop .#ci, merge-base
   against origin/main on a real PR ref) is unexercised.
3. **The plan's substrate promise is only half-wired.** Tier 1 says
   watchers "draft ledger rows through" error-context; `main-red-watch`
   currently drafts the row itself and greps raw codes from log tails
   WITHOUT invoking the router. The bundle enrichment (registry row +
   seam files per code in the issue body) is the missing half.
4. **F071 AGENTS one-liners cover 3 of 5 shipped scripts.**
   `reproduce-first-check.sh` and `incident-bundle.sh` have no Commands
   line (cap headroom exists at 174/377).
5. **bump-preflight's committed self-test has a false-green shape.** Its
   fixture pair (go-health v0.5.0→v0.5.1) legitimately changed zero .go
   files, so the self-test passes even though the diff path was broken at
   one point today (double-stripped zip prefix — caught only by my ad-hoc
   run against go-error-family). The CHANGES-DETECTED path needs a
   diff-bearing fixture pair in `--self-test`.
6. **`docs/lessons.md` got no war-story row for this train.** The
   observability-gap incident, the watch permissions class, and the
   proxy-zip `@version` prefix anchor are lessons; today they live in the
   ledger row, plan errata, and commit messages — three homes instead of
   the one (lessons.md).
7. **The research scan's R1–R8 proposal statuses are now stale** (several
   shipped today). Point-in-time artifact, acceptable drift — but a
   one-line "status: see plan errata" annotation would keep the series
   honest.

## c) NOT STARTED

- **Blocking flip for reproduce-first** (owner call): after N clean
  advisory trains, remove `continue-on-error`. No counter mechanism
  exists yet.
- **LLM error-prose gate** (gated, sketch only) — needs owner GO + CI LLM
  budget decision.
- **Agent dispatch on red main** (gated, sketch only) — needs owner GO +
  runtime/secrets/cost decisions.
- **The pma hook-override fix** (env-level `GIT_CONFIG_*` lets every
  daemon commit run hookless) — lives in the nix-config systemd unit;
  fleet-wide policy call, not started.
- **README incident-runbook rewrite** — deliberately deferred until the
  first real incident proves the bundle shape (plan §6).
- **Everything outside this train** (multicall-dock P1 agent chain,
  transcript search, push-notified approvals, cqrs-htmx AuditLog bounded
  projection upstream ask, …) — owned by other trains, untouched here.

## d) TOTALLY FUCKED UP

Nothing shipped destroys data or regresses the product — the binary is
untouched all train (verified: build job green on every push, zero
Go-source diffs outside tests). The fucked-up list is process-level, and
it is honest:

1. **The watchers' first night was 4/4 RED on my own permissions
   mistake.** An explicit GHA permissions block nulls everything
   unlisted; I granted `issues: write` but not `actions: read`, so the
   detection step died before any verdict. The "verify against live gh"
   I ran was a HOST-side dry-run with the host's admin-ish token — the
   exact wrong token class to prove a GITHUB_TOKEN permission. The bug
   was guaranteed to be found the moment the machinery ran as itself.
   (Silver lining, and it is real: the workflow's contract — red means
   the watcher is broken — worked exactly as designed, and the fix is
   one line. But I should have caught it BEFORE four red runs.)
2. **I edited `main.js` inside an active concurrent session's blast
   radius** and my change was silently clobbered by their stale buffer
   write; I found it only when oxlint flagged the now-orphaned duplicate
   block. AGENTS says re-View before each Edit in concurrent sessions —
   I re-Viewed before, not after, and paid a fix cycle.
3. **First-draft quality on the scripts was below the repo's bar**:
   reproduce-first needed four fix cycles (leftover draft line crashing
   under `set -u`; scratch branches not forking from base; `git init`
   default-branch assumption; EXIT trap referencing an out-of-scope
   local), daemon-audit's assert harness had inverted expectations twice,
   incident-bundle force-set `GH_TOKEN=""` breaking gh auth, and the
   bump-preflight zip-prefix handling had a double-strip bug the
   self-test's no-op fixture could not see. Every one was caught by the
   scripts' own self-tests or my verification runs — the safety net
   held — but the net was worked far harder than it should have been.
4. **Documented host traps bit again**: `PIPESTATUS` (empty verdict
   variable, the AGENTS-documented unavailability) and the `rg -r`
   hook-block (the host hook says third strike). Both cost a round trip
   each and both were written down before I hit them.
5. **I ran buildflow concurrently with my own edits** and misread its
   step-failure as a real regression before re-running clean (rc=0).
   Gates need a quiet tree; I knew that and did it anyway.

## e) WHAT WE SHOULD IMPROVE

1. **CI-workflow trains need a "trigger the real lane" verification step.**
   A workflow file's honest test is a dispatch (or waiting one cron tick)
   with the REAL token the runner uses. Fixture self-tests + host dry-runs
   proved logic, not permissions. Make this part of the train-end gate
   whenever `.github/workflows/*` changes.
2. **Self-test fixtures must pin the INTERESTING path, not just a passing
   one.** A no-op pair proves plumbing, not the diff engine. Rule of
   thumb: every verdict branch a script can emit needs at least one
   committed fixture that walks it (reproduce-first does this; the
   watchers do via `decide`; bump-preflight does not yet).
3. **Concurrent-session discipline**: re-View immediately after every edit
   to shared files (not just before), and re-verify the file's sha before
   staging.
4. **Commit narrative AT the phase boundary** — the daemon raced me into
   five contentless narrative commits today (content landed in daemon
   blobs first). The stage+commit-in-one-call rule works only if the
   commit happens the moment the unit is verified, not at tier end.
5. **Harvest `docs/lessons.md` in the same session** — war stories rot
   into commit messages otherwise; the one-home rule for stories was
   written for exactly this failure.
6. **Keep the verification ladder**: script self-tests → local real-data
   run → targeted nix checks → full buildflow → push+CI. Every script
   today climbed all five rungs; that is why four bug-cycles never
   reached main. Preserve it even when a train feels routine.
7. **Read the permissions block of any new workflow as an allowlist
   audit**, line by line, against every `gh`/API call the steps make.

## f) NEXT — up to 50, sorted by impact (brainstorm for HARVEST, not a commitment list)

**Verify/close the shipped train**

1. Observe the first GREEN `watch.yml` cron run (check `gh run list
   --workflow watch.yml` after the next :15/:30/:45/:00 UTC tick) — the
   fix's live proof.
2. Prove the issue-filing path end-to-end: on the next real red main (or a
   deliberately dispatched red via workflow_dispatch on a scratch branch's
   CI), confirm `main-red-watch` files the `ci-red` issue with a usable
   draft row.
3. Wire error-context into `main-red-watch`'s issue body (per-code registry
   row + seam files when the log tail yields codes) — completes the plan's
   substrate promise.
4. Confirm `gh run view --log-failed` works with the workflow's GITHUB_TOKEN
   (host token is 403; if the runner token also 403s, add `actions: read`
   is already done — then the LOG_CODES section should degrade loudly, not
   silently).
5. Open one real PR (any one-line docs branch) to exercise the
   `reproduce-first` CI job end-to-end; then record its runtime cost.
6. Add a diff-bearing fixture pair to bump-preflight's `--self-test`
   (go-error-family v0.11.0→v0.12.0) so CHANGES-DETECTED is pinned, not
   just NO-GO-FILE-CHANGES.
7. Add AGENTS Commands one-liners for `reproduce-first-check.sh` and
   `incident-bundle.sh` (F071 completion; cap at 174/377 has room).
8. Append the `docs/lessons.md` war-story row for this train: the
   zero-CI-runs observability gap, the watch permissions allowlist class,
   and the proxy-zip `@version` prefix anchor.
9. Annotate the research scan's R1–R8 table with shipped-status pointers
   (one line, ANNOTATE-not-rewrite).
10. Re-verify after 7 days of cron: zero false-positive ci-red issues
    filed on green/infra-flap runs (the dedup's first real soak test).

**Owner decisions pending (blocking work, not tasks I can start)**
11. Blocking-flip GO: how many clean advisory PR trains before
`reproduce-first` blocks; where the counter lives (tracker issue?).
12. Agent dispatch GO: yes/no, runtime (claude-code-action / Devin / Crush
headless), secrets custody, per-red budget.
13. LLM prose-gate GO: yes/no, option A (local) vs B (hosted) vs C
(agent-side only).
14. Fleet-side pma fix: stop the systemd unit from exporting
`GIT_CONFIG_*=/var/empty` so daemon commits run the repo hooks
(nix-config module change; re-binds the daemon fleet-wide).
15. Watch notification channel: label-only today — should ci-red issues
also ping a channel/email you actually read?

**Tooling hardening (small, high-value)**
16. `daemon-audit`: audit the last N daemon commits in the window, not
just the tip — the six-commit gap class would only trip the tip check
if the LAST commit is dirty.
17. `main-red-watch`: on EXISTS, still append the run id to the existing
issue if that run is NEWER than the issue's last-commented run
(today: same-run dedup only).
18. `bump-preflight`: diff the `.mod` files too (declared-dependency
drift, not just source).
19. `incident-bundle`: accept `--run-id` to anchor the gh window on the
failing run instead of a time window.
20. `error-context`: `--open` to print an editor-ready file:line shortlist
(agent ergonomics; cheap).
21. `whitespace-drift --ref`: self-test inside the script (it is verified
manually twice today; pin it like the others).
22. reproduce-first: emit the verdict as a GitHub step summary table when
run in CI (grep-friendly prefix exists; the table is free).
23. watch.yml: `timeout-minutes` is set; add `--limit` caps on every gh
call to bound worst-case API latency.
24. Add `actions: read`-class allowlist audit to the train checklist
(process item, one AGENTS line).
25. Consider a `watch.yml` manual "selftest" dispatch mode that runs both
scripts against fixtures in CI (proves the runner env, not just local
env).

**Adjacent trains (other sessions/owners — listed for completeness)**
26. Multicall-dock P1: agent chain (M11–M13 of that plan).
27. Multicall-dock P1/P2 tails (M14–M20 of that plan).
28. cqrs-htmx bounded AuditLog projection (upstream ask, AGENTS-documented).
29. `nix flake check` full local run on a quiet machine (drill eval +
sandbox legs together; today only targeted legs ran).
30. `webphone-backup-drill` rebuild sanity after the vm-tests.nix edit
(same file; untouched path, cheap re-proof).
31. TODO_LIST HARVEST of this report's (f) section per docs-health.

**Hygiene/debt (non-urgent)**
32. Retire the `REPRODUCE_FIRST_LOCAL=1` env guard once the CI job proves
itself (or keep and document — decide at blocking-flip time).
33. Consider archiving the four RED watch runs' lessons into the ci-flake
ledger as one row (infra class: permissions allowlist).
34. `error-context`: registry parse assumes the generated block's table
shape — pin a format assertion into `--self-test` so a generator
change screams here too.
35. `incident-bundle`: slug collision handling (two incidents same minute)
— today the timestamp suffix saves us; assert it.
36. Review whether `main-red-watch` should early-exit when the failing run
is older than the newest green (stale-red suppression).
37. Document the watch.yml cron-delay reality (GitHub fires late; the
"*/15" is an upper bound, not a promise) in the workflow comment.
38. Deduplicate the recurrence-grep needles (job names today; consider
failing-step names via `gh run view --json jobs` deep-linking).
39. Keep `.art-dupl-baseline` honesty: five new shell scripts added — the
dup-gate only sees Go, but a quick manual `art-dupl` glance at
scripts/ would keep the dedup-registry spirit.
40. After the first real incident: revisit README's incident-runbook
rewrite (currently deferred by plan §6) with the bundle as evidence.

(Items 41–50 intentionally left empty — the honest list is 40; padding a
brainstorm to a round number would violate the routing rigor the skill
demands.)

## g) Questions I cannot answer myself

1. **Blocking flip**: after how many clean advisory PR trains should
   `reproduce-first` start blocking — and do you want the count tracked
   in a tracker issue (visible, manual) or derived from run history
   (automatic, opaque)?
2. **Agent dispatch**: is there ANY appetite for GO on the dispatch half
   this quarter (and if yes, which runtime + who owns the secret), or is
   observer-only the standing posture until a real incident proves the
   human loop too slow?
3. **Fleet-side daemon fix**: the hook-override hole lives in the
   nix-config pma systemd unit and fixing it re-binds the commit hook to
   every daemon commit on every repo on this host — do you want that
   change prepared (and effectively fleet-enabling hooks), or is the
   daemon deliberately hook-exempt by policy?

---

_Prepared per the status-report skill with two operator overrides (`.md`
format, top-50/top-3 instead of top-25/top-1) — both instructed in the
request, flagged here. Then: waiting for instructions._
