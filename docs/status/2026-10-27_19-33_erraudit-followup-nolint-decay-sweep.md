# Status — erraudit follow-up session: nolint-decay sweep, probe preservation, buildflow attribution

Claims verified as of 19:33 CET, 2026-10-27 (push/CI and buildflow rows are the first to rot).
Scope: THIS follow-up session only (~18:47–19:33). The parent train's report with its own
a–g lives at `docs/status/2026-10-27_18-23_erraudit-max-zero-typed-error-train.md`; this one
executes its §h addendum and closes (or attributes) every open row.

Snapshot anchors: HEAD = c08c36f, `origin/main == HEAD` (0 unpushed), latest push-CI run
SUCCESS (18:07Z, 4m09s), full Go suite 31/31 ok, `erraudit lint ./...` AND `--max` exit 0,
`erraudit nolint-audit`: 72 directives / 68 needed / 4 stale (all 4 in the sibling-owned
crm seam), standalone `nix build .#checks.x86_64-linux.treefmt` exit 0.

## a) FULLY DONE

1. **Dead-nolint sweep, evidence-driven, beyond the narrow grep.** The parent report
   scoped "comments whose fmt.Errorf no longer exists"; I ran the purpose-built tool
   instead: `erraudit nolint-audit` over the whole repo found **41 of 109 directives
   stale**. Removed 37 across 26 files with a per-line sanity-checked script (dry-run
   first, bottom-up per file, marker-position verified before every strip). The
   headline: typed constructors (`*bootStepError`, `*formError`, `*archiveError`,
   `*pageError`, `*migrateError`) decay their own suppressions — every
   "family-neutral propagation" comment died the moment its helper started returning
   a specific type. Verified inert: normal-mode lint exit 0, `--max` exit 0, build ok,
   vet clean, gofmt clean, full suite green. The 4 crm-seam stales are deliberately
   deferred (concurrent-session ownership, see b4).
2. **Probe fixtures rescued from /tmp and committed.**
   `internal/arch/testdata/erraudit-rule-probes/` (a/b/c/g/h.go + `other/o.go` +
   go.mod) — the 7-file module that decoded the `generic_return` rule — plus a README
   pinning every clause (bare-error + construction; composite-literal counting;
   propagation/New-prefix/typed-return exemptions; same-package transparency; the
   opaque cross-package `Wrapf`-name heuristic) against tool build `61a744c`. The
   parent session left these one /tmp-wipe away from loss; they are now the
   re-verification harness for any future erraudit upgrade.
3. **lessons.md war story** (Tooling traps): construction-shaped-not-signature-shaped
   rule, the probe method ("reverse-engineering a linter from findings is guesswork"),
   the registry-scan trap, and the nolint-decay law with its audit cadence.
4. **End-to-end sentinel proof** (parent §c2 closed):
   `TestNotFoundVerdictSurvivesLiveEngineCall` boots a REAL engine (`Open(t.TempDir())`)
   and drives `Contacts.Delete` on a missing id — `errors.Is(err, store.ErrNotFound)`
   matches (the exact contract `actions.go:577` / `panels.go:87` consume) AND the coded
   half stays `errors.AsType`-able with code `readmodels.contact_missing`. Both notfound
   tests pass. This also settles the LSP "unused notFoundError" warnings as phantom
   (same-package call sites + tests; `go vet` clean; the file has no `errors.As` at all,
   so the govet `errorsas` warning there is equally phantom).
5. **Buildflow attribution complete — the parent's b1 "unconfirmed" row is now
   CONFIRMED WITH NAMED CAUSES.** Fresh full run (18:52→19:05): 2 step(s) failed,
   both foreign to this train: (1) `type-check` — 680 findings, 100% in
   island/shell/enroll JS, caused by an **untracked root `tsconfig.json`** (born
   18:34:10, strict `checkJs` over the whole island) from the SIBLING session's island
   type-gate train (their execution-status doc landed 19:0x; top errors are ordinary
   strictness: implicit any, `EventTarget.closest`, `unknown` err — not a broken
   d.ts cascade); (2) `nix-hash-fix` — **no dependency bump happened today** (last
   go.mod touch: the Oct 25 daemon commits), the documented no-bump noise class.
   `test-race` PASSED this run (+54% slower — load, matching parent §c3's hypothesis).
   Everything this train owns is green; the earlier 18:33-era test-race failure did
   not reproduce.
6. **AGENTS.md updated in place**: the train's erraudit bullet now carries the
   nolint-decay law, the `erraudit nolint-audit` cadence rule, probe/testdata pointers,
   and the live-engine test reference.
7. **Prior report §h addendum** written with per-row closures (b3, c2, g1-partial).
8. **Push/CI verified per doctrine**: origin == HEAD; the sweep rode daemon commits
   b838328 + 4396116, covered by the green 18:01Z CI run; the 17:59Z red was the
   sibling's own `fix(ci)` commit, healed by their follow-up push. g1 from the parent
   report self-resolved: the daemon's push-leg recovered (17:42Z run pushed ba20a1c).

## b) PARTIALLY DONE

1. **Full buildflow gate green: NOT on this tree.** Attribution is complete (a5) but
   the gate itself stays red until (i) the sibling's island type-gate burn-down lands
   and (ii) the nix-hash-fix question is answered (g3 below). A green full run has not
   been seen since the 17:42-era pushes.
2. **Upstream `Wrapf`-heuristic filing** (parent g3, owner-unanswered): the
   verify-before-filing evidence half is now DONE (minimal repro committed and
   importable); the go-finding prior-art search, the draft, and the filing itself are
   not started. I did not advance this without the owner's answer — but I could have
   done the read-only prior-art search in parallel; didn't.
3. **Timing-regression baseline** (parent §c3): the fresh run's regression list is
   captured (test-race +54%, govulncheck +112%, golangci-lint +71%, …) but the
   "concurrent-session load" attribution is still hypothesis-only — no quiet-machine
   measurement exists to compare against.
4. **4 sibling-crm stale nolints** (`crm/client.go:38,42,45`, `crm_test.go:37`):
   identified, deferred, and now tracked ONLY in this report. If the sibling's crm
   train lands and nobody reads this, they rot. Needs a TODO_LIST row (harvest, f9).
5. **TODO_LIST harvest**: this report's section f is harvest input per the status-report
   skill, but the standing instruction for this session is "report, then WAIT" — so
   TODO_LIST.md is untouched and the harvest is queued behind your word.
6. **CHANGELOG stance**: the sweep + probes + test are comment-only/test-only/docs-only
   deltas — per the ratified bar these stay changelog-silent. Decision made and
   recorded here; nothing written. (If you disagree, one line goes in.)

## c) NOT STARTED

1. **Upstream erraudit issue** (see b2) — not filed.
2. **vendorHash re-pin** — owner-gated (g3); the proving run (one
   `nix build .#webphone.goModules --rebuild` + re-pin) not spent.
3. **Baseline refresh** — 47 stale baseline entries warned across 3 consecutive runs;
   refresh is tier-3 owner territory (the daemon has mangled policy files before, so I
   refuse to regenerate unprompted) and I did not even prep the legwork (a preview diff
   of what `--save-baseline` would write) — flagged in e6, question g1.
4. **Island type-gate burn-down support** — 680 findings are the sibling train's
   blast radius; I touched nothing (correct), but also offered nothing (a triage of the
   error classes per file would take 20 minutes and could help them — not done).
5. **`nolint-audit` as a buildflow step** — idea born this session (decay should be
   caught at gate time, not months later); zero work done.
6. **erraudit-upgrade tripwire** — the probes pin behavior to `61a744c` but nothing
   FAILS when the tool build changes; a cheap version-check guard is undesigned.
7. **Smoke re-run** — deliberately skipped this session (comment-only delta; reasoning
   recorded in the parent report §h6). Next behavior-bearing change owes a fresh one.
8. **Prior report f-items 11–50**: untouched (this session scoped its own Tier-1 set).

## d) TOTALLY FUCKED UP

Nothing destroyed, nothing reverted that wasn't mine, no red left behind by this
train, zero findings in any file I touched. Honest misses, all self-caught but all
preventable:

1. **The narrative commit was EATEN AGAIN — the third documented instance.** I
   followed the "stage+commit in ONE tool call" rule (explicitly designed against the
   2026-10-17 lesson) and STILL lost: the daemon committed my staged 27-file sweep
   mid-call (b838328, "chore: auto-commit 27 changed file(s)"), my `git commit` found
   "nothing to commit, working tree clean". The one-call rule reduces the race window;
   it does not close it. The narrative now lives in this report instead of history.
   A real mitigation needs a different mechanism (commit-message file + `-F` executed
   within the same call as add is equally race-exposed; a daemon pause hook would be
   fleet policy — question for you, not something I'll improvise).
2. **I repeated the exact mistake the parent report warned against, in the same day.**
   Parent §e4: "check the first run's artifacts before re-running." I launched the
   fresh buildflow piped through `tail -40` (07B), which DESTROYED the identity of the
   second failed step; then burned two more invocations (09B verbose-grep that matched
   nothing because the step-vs-findings-gate shape differs; 09C full-capture to a file)
   to learn what one properly-captured log would have said. ~25 minutes of gate time
   for nothing. New rule for lessons: capture to a file, grep the file, never pipe a
   long gate through `tail`.
3. **First `nolint-audit` invocation was wrong** (`./...` path arg → false "No
   //nolint:erraudit directives found") — one wasted round trip because I assumed
   git-style path args instead of reading `--help` first. The tool takes a plain
   directory or nothing.
4. **My strip script left a formatting drift behind**: removing trailing comments from
   gofmt-aligned comment columns requires a re-align (`internal/server/actions.go` was
   flagged by `gofmt -l`). Caught by my own follow-up check BEFORE any gate — but the
   script should have ended with `gofmt -l`/`-w` on its touched files by construction.
   Had I not checked, `treefmt-check` would have red the next full gate run.
5. **The probes sat in /tmp across a session boundary** — the parent session ended
   without committing them (its own report listed the preservation as future work).
   Evidence that matters should be committed when created, not "next session".

## e) WHAT WE SHOULD IMPROVE

1. **Gate-capture discipline**: long-running gates (buildflow, full suites, VM tests)
   ALWAYS write full output to a named file; analysis greps the file. `| tail` is for
   humans watching, never for agents triaging. (Direct fix for d2.)
2. **Script-hygiene convention**: any session script that edits Go files ends with
   `gofmt -l` on its touched set (and `-w` when the drift is its own). (Fix for d4.)
3. **Tool-CLI rule**: read `--help` before the first invocation of any subcommand,
   even "obviously git-shaped" ones. (Fix for d3.)
4. **Commit-phase doctrine**: narrative commits at phase boundaries are hostage to the
   daemon. Options: (a) accept daemon authorship + put narrative in the phase report
   (today's de-facto outcome), (b) fleet-level daemon pause/resume around narrative
   commits, (c) `--no-verify`-style escape hatch doesn't exist for the daemon. This is
   a policy call, recorded here so it stops recurring as a surprise.
5. **Phantom-LSP-diagnostics awareness**: golangci_lint_ls emitted stale
   `errorsas`/`unused` findings against freshly refactored files for the entire
   session; the parent summary knew, I still spent a vet-run re-proving it. AGENTS
   already documents concurrent-session staleness for shared files; a one-line
   "arbiter is `go vet` + tests, never the IDE panel" addition would canonize it.
6. **Prepare owner decisions, don't just escalate them**: for the baseline question I
   could have run `--save-baseline` into a scratch worktree and shown you the exact
   diff (47 entries → presumably near-empty) so your decision is yes/no, not research.
   Doing that legwork next time.
7. **Untracked configs that gate local behavior are a drift class**: the sibling's
   root `tsconfig.json` makes LOCAL buildflow fail in ways a fresh clone won't
   reproduce (CI is unaffected). Whether it becomes tracked-and-enforced or stays a
   private scaffold must be explicit (g2) — otherwise the next session reds on it
   without context, exactly as mine did.
8. **Cross-train conflict audit on mixed commits**: d035691 interleaved my probes with
   sibling island JS. CI green either way, but a 30-second `git show --stat` review of
   daemon commits that mix trains would catch a wrong-file landing early. (Not done
   this session; landed correctly by luck of timing.)

## f) Up to 50 things we should get done next (Pareto-ordered in tiers)

Tier 1 — unblocks the gate or closes an owner question (do first):

1. Wait out the sibling's island type-gate train, then burn down or finalize the 680
   checkJs findings so `type-check` returns to the green side of the gate.
2. Decide + implement the `tsconfig.json` end-state: track it (and own the 680
   findings as a real gate) or remove it when the sibling train ends (g2).
3. Baseline refresh with legwork done: I preview `buildflow --save-baseline` in a
   scratch worktree, you approve, we commit the shrunk baseline (g1).
4. nix-hash-fix: spend ONE proving run (the goModules rebuild recipe from AGENTS) to
   answer stale-pin-vs-noise with evidence instead of doctrine (g3).
5. Upstream filing for the opaque `Wrapf` heuristic: go-finding prior-art search →
   verify-before-filing checklist → github-voice draft → file. Evidence half is
   already committed (probes).
6. Remove the 4 crm-seam stale nolints the moment the sibling's crm train lands
   (`erraudit nolint-audit --fix` will do it mechanically).
7. Re-run full buildflow to GREEN once (1) + (4) land — the actual closure of the
   parent's b1 row.
8. TODO_LIST harvest of this section (docs-health HARVEST) — say the word.
9. Quiet-machine timing baseline: one `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` run on
   an idle tree to replace the load-skewed regression verdicts (b3).
10. Fresh smoke on the next behavior-bearing change (precedent + reasoning for the
    skip is recorded; the precedent must not become a habit).

Tier 2 — hardening and cheap wins from this session's findings:

11. Propose `nolint-audit` as a buildflow step (advisory severity; decay caught at
    gate time).
12. Erraudit-upgrade tripwire: pin/fail-loudly when the erraudit build changes from
    `61a744c`, prompting a probe re-verification.
13. lessons.md append: the daemon-race third instance (d1) + the capture-to-file rule
    (d2) + the script-gofmt rule (d4) — one combined "agent gate hygiene" entry.
14. AGENTS one-liner: phantom-LSP findings arbiter = `go vet` + tests (e5).
15. Upstream DX note to go-finding: `nolint-audit ./...` silently reports "0
    directives" for a git-style path arg — either accept `./...` or error.
16. Prepare the `Wrapf`-footgun rationale comments at the remaining 3 `Wrapf` call
    sites (`userauth.go`) regardless of the upstream filing outcome (parent e5/46).
17. `grep -n "errors\.As(" --type go` audit: confirm zero non-test legacy-`As` sites
    remain (the `--max` gate implies it; one grep makes it visible).
18. Cross-train audit of mixed daemon commits (e8): read-only `git show --stat` pass
    over today's daemon commits; confirm no wrong-file landings.
19. `/tmp` hygiene: `tc.log`, `buildflow_v3.log`, the strip script, and the now-redundant
    `/tmp/errtest` — remove after confirming the committed copies (trivial).
20. docs/error-contract.md: decide whether the typed-error PRODUCTS (ForwardError,
    ShutdownError, notFoundError, …) deserve a short inventory section, or stay
    registry-only (they carry no codes; drift risk low — a ruling prevents relitigating).
21. Sweep the untracked-file population (`git status --ignored` review) for OTHER
    stray configs that alter local gate behavior like tsconfig did.
22. Confirm the island-tests file that rode b838328 with my sweep
    (`calls-disposition.test.mjs`) is fully the sibling's (read-only blame check).

Tier 3 — roadmap fuel (most belong in ROADMAP, not TODO_LIST):

23. errorfamily upstream: a `Join`-family constructor (the parent train hit the gap:
    shutdown aggregates had to leave errorfamily to use errors.Join).
24. A "typed error products" map of the codebase (which envelope for which seam) if
    the inventory in (20) grows beyond a table.
25. Revisit the parent report's f-items 21–50 for anything harvest-worthy (none looked
    urgent from this session's vantage; docs-health ANNOTATE can mark them).
26. Consider `errors.AsType` adoption in readmodels_test helpers as the canonical
    example for future test code (parent e2, ongoing discipline — no work item, a
    convention note).
27. Evaluation: should `erraudit lint` (normal mode, not `--max`) run in CI as a
    fast job, so suppressions and family shapes are checked off-machine too?
28. Longer-term: fleet-level daemon pause protocol for narrative commits (e4 option b).

(Stopped at 28: items 29–50 would be padding — the honest backlog above is complete
for this session's horizon; ROADMAP owes breadth, not this report.)

## g) Questions I cannot answer myself

1. **Baseline refresh authorization**: 47 accepted finding entries have matched
   nothing for 3 consecutive runs — the typed-error conversions dissolved their
   sites. Refreshing the baseline is tier-3 owner territory and the daemon has
   mangled `.buildflow.yml` before, so: do I run `buildflow --save-baseline` (I'll
   show you the diff in a scratch worktree first), delete the file outright, or leave
   the warning nagging?
2. **The untracked root `tsconfig.json`** (sibling session's strict island type-gate,
   source of the 680-findings red): should this become a TRACKED, enforced gate after
   the sibling's train lands (meaning someone owes the 680-finding burn-down as a real
   train), or is it scaffolding to remove when they finish? I cannot know their
   train's intended end-state, and the drift risk (local gate ≠ fresh clone) is live
   either way.
3. **nix-hash-fix, still red, still no bump**: spend the one proving run now (goModules
   rebuild recipe → re-pin vendorHash if it really moved) or wait for a quiet tree?
   This is parent-g2 re-asked because the red persists across every run today and the
   doctrine ("a red step IS the stale-pin signal") now contradicts the no-bump
   evidence — I need your call to break the tie.

---

_Rendered per the 2026-10-27 ruling: every claim above names its verification; the
push/CI/buildflow rows are the first to rot. WAITING FOR INSTRUCTIONS._
