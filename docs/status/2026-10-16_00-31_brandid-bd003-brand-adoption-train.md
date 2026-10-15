# Status Report — brandid-lint BD003 brand-adoption train

- **Written:** 2026-10-16 00:31 CEST
- **Scope discipline:** this report covers ONLY the current session's run (the
  `brandid-lint -rules all .` scan the operator handed over with exit 1 / 35
  findings, and everything that fell out of resolving it). No unrelated
  research was done.
- **Format note:** the status-report skill is HTML-canonical; the operator's
  prompt explicitly demands `.md` at this path, so `.md` wins (flagged per
  skill contract, not propagated into the skill).

---

## Executive summary

The session took an exit-1 `brandid-lint -rules all .` scan (35 BD003
findings: string-typed identifiers the brand-adoption rule wants branded) to
exit 0 without changing a single runtime byte of behavior — every finding was
resolved by recording the ALREADY-EXISTED ruling at the site as a
`//brandid-lint:ignore(BD003) <reason>` directive, the tool's own designed
mechanism for "this primitive stays primitive on purpose". One real code
change rode along: a tautological dead condition in `userauth.Service.Shutdown`.
All gates green; pushed under the manual-push bar after a 154-minute daemon
push stall; CI green.

Headline numbers:

| Metric                     | Value                                      |
| -------------------------- | ------------------------------------------ |
| BD003 findings in          | 35 (2 ERROR, 2 WARNING, 31 INFO)           |
| Findings resolved          | 35 (100%)                                  |
| Real type changes made     | 0 (deliberate — all sites carried rulings) |
| Code deltas                | 1 (tautology removal, `userauth.Shutdown`) |
| Files touched              | 18 Go files + AGENTS.md + CHANGELOG.md     |
| Full test suite            | green (24 packages; server suite 550s)     |
| buildflow                  | exit 0                                     |
| CI on pushed tip `c029741` | success (run 38814348023)                  |

---

## a) FULLY DONE

1. **All 35 BD003 findings resolved — scan exits 0.** Verified twice (after
   edits and after `nix fmt`). Also verified under the default rule set
   (BD001+BD002): exit 0 — which additionally proves every inserted directive
   is well-formed AND effective (a no-op directive would have fired BD002).
2. **Context gathering before touching anything.** Read the linter's source
   (`~/projects/go-branded-id/linter`): BD003 is detection-only (`-fix` never
   touches it — the operator's `-fix` flag was a no-op by design), directive
   grammar/placement rules (trailing / directly-above / doc-group-last-line),
   and the severity model (user-domain signatures → ERROR).
3. **Per-finding categorization against recorded rulings.** All 14 flagged
   files read; every site turned out to carry a fresh, written rationale
   (wire-shape facts/views, opaque island SIP Call-ID, retired legacy SQL
   rows, external textagent contract, base64url WebAuthn credential ids,
   parse-strictly-at-the-userauth-door enroll signatures). Suppression with
   the recorded reason was the correct and honest remediation — minting the
   tool's suggested mechanical brands (`FaxViewID`, `faxInputID`,
   `LegacyContactID`, …) would have split-brained the existing
   `domain.FaxID`-class brands.
4. **The two ERROR-severity signatures decided, not defaulted.**
   `EnrollTokens.Mint`/`Lookup` keep raw `userID string`: the consumer
   (userauth) is already fully typed (`identitymodel.UserID`, strict parse at
   its door with a dedicated corruption code); typing the seam would couple
   readmodels → cqrs-htmx/identity-model and relocate auth-forensics error
   classification. Reasoning recorded in the directive at the site.
5. **Tautology fix:** `userauth.go` Shutdown's first `&& firstErr == nil` was
   dead (gopls nilness warning in a file I was editing) — removed; `go vet`
   - build green; disk-verified (the LSP kept serving a stale snapshot —
     see e-7).
6. **Full verification chain:** `nix fmt` before the gates (18 files
   reflown into the daemon commits), `go test -count=1 ./...` all green
   (server's 550s includes the session-behavior suite — the auth-regression
   convention is satisfied), `go vet`/`go build` green, `buildflow` exit 0.
7. **Push + CI.** Daemon push stalled 154 min past the first unpushed
   commit; manual-push bar invoked WITH the stated anchor (2026-10-15
   21:26:48+02:00), `scripts/push-stall-watch.sh --no-fetch` exit 1 as
   evidence. The daemon raced my push (remote already at `c029741` — its
   push leg recovered). CI on `c029741`: success.
8. **Memory updated:** `AGENTS.md` strong-id bullet now records the brandid
   convention (BD003 clean as of 2026-10-15; never mint the tool's
   mechanical brand names) and `CHANGELOG.md` got its `[Unreleased]` line
   per the changelog bar.

Commits carrying this train (all daemon-shaped, per the auto-commit
doctrine): `4831358`, `8fbb8d6`, `6c00695`, `e52e04e`, `c029741` (the 18-file
directive + fmt snapshot), `9a66329` (AGENTS.md + CHANGELOG).

## b) PARTIALLY DONE

1. **Memory persistence of the decision rationale.** The AGENTS.md bullet
   records the convention and the "never mint mechanical brands" rule, but
   the per-seam WHYS (why Mint/Lookup stay raw, why textagent ids stay
   foreign, etc.) live only in the inline directives — correct per
   one-home-per-fact, but a future `-rules all` re-run by a fresh session
   will not know the ERROR-class suppression was a deliberate doctrine call
   without reading this report or the git history.
2. **Directive policing.** The directives now exist, but nothing RUNS
   `brandid-lint -rules all` in any gate. BD002 will keep them honest only
   when someone actually runs the tool (see f-1/f-2).
3. **The docs-only commit `9a66329`** (AGENTS.md + CHANGELOG) was unpushed at
   report time (origin at `c029741`). Committed 2026-10-16 00:29-ish; the
   60-minute bar clock started then; the daemon's push leg was proven alive
   this session, so it rides the daemon unless it stalls again.
4. **LSP hygiene.** The stale gopls diagnostic (see e-7) was worked around
   (disk + `go vet` verification) but not root-caused or reported.

## c) NOT STARTED

1. **HARVEST of section (f) into `TODO_LIST.md`/`ROADMAP.md`.** The
   status-report skill mandates the loop-close; the operator's prompt says
   THEN WAIT FOR INSTRUCTIONS — the wait wins. Harvest is one `docs-health`
   invocation away once instructions arrive.
2. **Making the scan durable** (buildflow step / CI job / flake check for
   `brandid-lint -rules all`) — proposed below, not started.
3. **Codemod evaluation for future BD003 adoption** (the linter explicitly
   says a real fix is "a codemod's job") — not started, probably not wanted
   given this repo's wire-shape doctrine.
4. **Upstream report of the stale-LSP-snapshot behavior** — not started.

## d) TOTALLY FUCKED UP

Nothing. The tree is green end-to-end (suite, buildflow, CI), the working
tree is clean, and no behavior changed. Two near-misses that were caught
in-flight and cost only seconds (listed honestly under "what could I have
done better" below, since "caught instantly, zero damage" is not "fucked
up"):

- a 4-edit batch was mis-addressed to `transcript.go` (the exact-match
  failure caught it before any damage; re-issued per-file);
- three files read via `sed` were rejected by the edit tool until properly
  Viewed (wasted round trips, no damage).

The one genuinely regrettable NON-event: the whitespace-drift ritual
(`scripts/whitespace-drift.sh` after `git add`) never ran because the daemon
committed the waves before I could stage anything. CI-formatter-green on
`c029741` says no drift landed, but the ritual was skipped, not satisfied.

## e) WHAT WE SHOULD IMPROVE (what I forgot / could do better)

1. **I suppressed all 35 findings without flagging the ERROR class to the
   operator first.** The one-alternative protocol says execute decisively,
   and the reasoning is solid — but "user-domain id in a signature" is the
   exact class the tool grades highest, and the suppression encodes a
   doctrine interpretation (store-seam-transport) the owner may want to
   re-affirm or overrule. A one-line "heads-up, 2 ERRORs suppressed for this
   reason, object if you disagree" in the running commentary would have
   made the call auditable in real time instead of post-hoc.
2. **I never checked whether brandid-lint supports a repo-level config** for
   default rules/exclusions before concluding "ad-hoc scan". If it does, the
   f-1/f-2 work gets cheaper. Not researched (out of the session's scope).
3. **Read-tool discipline.** Three `sed`-first reads cost a re-View each;
   one batch of edits was mis-addressed. The exact-match machinery saved
   me both times — that is the guardrail working, not good aim.
4. **The `-fix` expectation gap.** The operator ran `-fix`; for BD003 that
   flag is defined to do nothing. I should have stated that explicitly in
   my first response instead of letting the flag silently mean "manual
   resolution expected".
5. **Timing-noise in the buildflow tail.** I let the gate's
   timings-regression block (30 warmup-percentage lines) pass uncommented.
   It is exit-0 noise, but a one-line "all timing regressions are
   cache-warmup noise" would have closed the loop faster for the reader.
6. **Stale-diagnostic hygiene.** I verified on disk and moved on; the
   sharper move is a 2-minute isolation (does Crush's LSP integration serve
   pre-edit snapshots after `lsp_restart`?) so it can be reported instead
   of re-encountered.
7. **Background-shell bookkeeping.** The first buildflow run's shell record
   vanished (`background shell not found`) and I simply re-ran it. Fine
   here; in a longer session that pattern can silently drop a gate verdict.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Operator asked for "up to 50". Grounded strictly in what this session
Touched or Noticed, the honest list is 25 — padding to 50 with invented
work would violate the no-unrelated-research instruction. Grouped, impact
order inside groups. Items 1–8 are the real next work; 9–25 are the
session's observed smaller polish.

**Durable the scan (highest impact — otherwise the directives rot):**

1. Wire `brandid-lint -rules all` into BuildFlow as a step (or `.buildflow.yml`
   skip-with-rationale decision, same as the cqrs-lint precedent) so exit
   non-zero reds the train.
2. Add the same scan as a CI job (or ride the BuildFlow CI entry once 1
   lands) so the gate runs outside the devshell too.
3. Pin the linter version in `flake.nix` (devShell input) so directives are
   policed by the same grammar that wrote them (v0.5.1 today).
4. Decide the repo-level config question (does brandid-lint read a config
   for default rules?) — if yes, encode `-rules all` as default so bare
   runs police BD003.
5. Add a one-paragraph pointer from the AGENTS.md strong-id bullet to the
   enforcement entry point (wherever 1 lands), so the next session finds
   the gate, not just the convention.
6. `docs-health` HARVEST: pull items 1–8 (and anything the owner promotes
   from 9–25) into `TODO_LIST.md`; the rest are ROADMAP fuel.
7. Re-run `-rules all` after 1–3 land to confirm the gate is green IN the
   gate, not just in a shell.
8. Fleet-level (owner policy): the same BD003 sweep for sibling repos that
   share the strong-id doctrine — out of this repo's scope, owner's call.

**Doctrine audits this scan surfaced (medium):**

9. Owner re-affirmation of the ERROR-class enroll decision: keep
   `Mint`/`Lookup` raw (store-seam-transport reading) or adopt
   `identitymodel.UserID` at the readmodels signature (typed-seam reading).
   Both defensible; currently decided by directive.
10. Sweep the remaining `//nolint:branching-flow` comments that now sit NEXT
    to a `//brandid-lint:ignore(BD003)` directive: if the branching-flow
    strong-id rule is permanently superseded by BD003 for this repo, the
    duplicated rationales could be consolidated (two tools, one reason, two
    comments — drift risk).
11. The 19 `.buildflow-baseline.json` branching-flow entries: after 10,
    check whether any baseline IDs describe strong-id findings that BD003
    now owns inline — retire what became redundant.
12. File the `faxInput`/`snippetInput`/`transcriptInput` one-line key-input
    structs as a named pattern in the dedup registry (three siblings, one
    shape — intentional, now directive-annotated three times).

**Tooling friction observed this session (polish):**

13. Isolate + report the stale-LSP-snapshot behavior (diagnostics pinned to
    pre-edit line:col even after `lsp_restart` and disk change) — Crush
    integration or gopls, needs a minimal repro.
14. Investigate the vanished background-shell record (buildflow run 016:
    "background shell not found" after completion) — bookkeeping bug or
    expected pruning?
15. Make `scripts/push-stall-watch.sh` print the proposed `git push` line
    when the bar trips (saves the operator the anchor restatement).
16. Add the manual-push-bar invocation to the session transcript ritual: the
    anchor timestamp + watch exit code belong in the report that mentions
    the push (done here; make it a habit).
17. Consider a `justfile`-free wrapper (`nix run .#brand-scan`?) so the
    scan is one memorable command instead of a raw binary invocation.

**Docs/memory follow-through (cheap, do with 6):**

18. Cross-link this status report from `TODO_LIST.md` when harvesting.
19. When the linter releases the next version, re-run `-rules all` once —
    directive grammar or detection changes could surface new findings
    (the repo pinned no version expectations yet).
20. Add `brandid-lint` to the BuildFlow "known tools" documentation next to
    branching-flow, with the "BD003 is opt-in, detection-only" note.
21. Record in `docs/lessons.md` (if the owner agrees it is lesson-grade):
    "an adoption scanner's `-fix` may be a no-op for the rule you care
    about — read the rule family table before promising auto-fix."

**Carry-over hygiene noticed in passing (no research done, flag-only):**

22. The AGENTS.md project-context snapshot in this session's context showed
    a much longer file than the 151 lines on disk — the context render and
    the file have diverged in size; worth one glance at how AGENTS.md is
    injected (not investigated, per scope).
23. The daemon's incremental commits split this train across 5 auto-commits
    with heuristic names; the narrative is only recoverable from this
    report. If train-level narrative matters, the release-runbook's
    "narrative commit per phase boundary" convention is currently
    daemon-vs-ritual contested territory.
24. `buildflow` timing-regression noise (30 lines of warmup percentages in
    every tail) — a BuildFlow-side "suppress < N%" knob would shrink every
    future report tail.
25. The 60-second sleep in my final verification loop was a crude daemon
    wait; a tiny `scripts/daemon-settled.sh` (poll `git status` empty AND
    `git log -1 --format=%ct` older than N seconds) would make "wait for
    the daemon" deterministic.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **The ERROR-class doctrine call:** should the enroll-token seam
   (`EnrollTokens.Mint`/`Lookup`) stay raw-string (my reading: transport
   seam, typed at the userauth door, corruption codes stay home), or do you
   want `identitymodel.UserID` AT the readmodels signature (typed seam,
   accepting the cqrs-htmx import edge and the classification relocation)?
   The directive currently encodes MY interpretation of your 2026-10-05
   ruling — confirm or overrule.
2. **Enforcement policy:** should `brandid-lint -rules all` become a
   BuildFlow/CI gate for THIS repo (and eventually the fleet), or stay an
   ad-hoc scan you run by hand? Directives without a runner will drift
   silently; a runner adds a step every train must keep green. Your call —
   it is a fleet policy precedent either way.
3. **Was the stale-LSP diagnostic behavior already known?** (Diagnostics
   pinned to pre-edit positions, surviving `lsp_restart`, contradicting
   disk bytes.) If you have seen it before and it is already tracked, I
   drop item f-13; if not, I isolate and report it.

---

_Report ends. Waiting for instructions (per operator prompt). The skill's
post-report HARVEST obligation is intentionally deferred to that instruction._
