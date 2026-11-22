# Status — armed set holds, push/CI closed (2026-11-22 evening)

> Claims verified as of **2026-11-22 17:39 CET**. CI/push claims re-check
> with `git ls-remote origin main` + `scripts/ci-verdict.sh <sha>`.

**Scope line.** Covers only this session since report
`2026-11-22_15-59_cqrs-lint-exportlegacy-triage-status.md`: the "why
still issues" verification over the user's 17:33 paste and the closure
of that report's push/CI duty. Verified here: push state at 17:39
(remote = HEAD = 7f39d6a7, tree clean), `scripts/ci-verdict.sh
7f39d6a7` → "GREEN: CI (1 run(s)) concluded success" (workflow-name
verified, read ~17:38), and the 17:33 findings list compared
rule-by-rule against the cache-bypassed 15:52 run (identical
composition). NOT verified here: a fresh no-cache cqrs-lint run at
17:33 (the user's paste was a 100% result-cache hit; trust rests on the
composition match), the BuildFlow repo delta behind the new
binary-freshness warn (sibling repo, deliberately not read), and every
carried item from the 15:59 report that is listed below as still open.

## a) FULLY DONE

- **Push + CI duty from report #1: CLOSED.** The daemon pushed; at 17:39
  remote main = HEAD = 7f39d6a7. `scripts/ci-verdict.sh 7f39d6a7`
  answers GREEN for the CI workflow itself (not a watch-lane green).
  The manual-push bar was never invoked — the daemon landed it inside
  the window.
- **The 17:33 run adjudicated as the ruled state.** Composition is
  identical to the 15:52 cache-bypassed verification: 13 findings =
  C014x2, D006, E023x6, F009, F021x3; P012/P013 absent; nothing new,
  nothing lost. The armed keep-visible set holds with no drift.
- **"Why still issues" answered with the ruling table** (per-class
  standing rulings), the noise decode (baseline staleness and timing
  regressions are documented scoped-run artifacts; exit 0 because the
  findings gate trips only at error severity), and a keep-armed
  recommendation with the suppress-on-order alternative named — owner
  doctrine stands absent an explicit order.
- **Report #1 fully landed**: the 15:59 file, the regenerated INDEX (49
  reports), and the dprint repairs are all inside 7f39d6a7's ancestry
  (daemon commits f05e4833 → f15a6195 → 7f39d6a7).

## b) PARTIALLY DONE

- **Preflight warn triage — now three warns, zero dispositions.**
  Carried: state DB 1.09 GB (VACUUM or wait for post-purge GC) and
  dprint.json's 3 remote plugin fetches. New: binary-freshness (binary
  built at 9a6af21, BuildFlow repo HEAD d57a64a — a concurrent session
  advanced the tool's own repo past the installed binary; advisory).
  All three remain outside the fixed/filed/deliberate-ignored states
  the buildflow skill's close-the-loop rule demands.
- **Full-train buildflow battery** — still not run (carried; scoped
  runs only all session).
- **Baseline-staleness + timing-regression verdicts** — still dismissed
  as scoped-run noise without a full-run confirmation (carried).

## c) NOT STARTED

- **BuildFlow delta characterization** — the warn's own triage hint
  (`git log 9a6af21..d57a64a --oneline` in ~/projects/BuildFlow) was
  not run; kept out of scope by the session's research discipline. It
  would say whether the stale binary can mis-report cqrs-lint at all.
- **Binary rebuild/reinstall** — not attempted (owner call; see g).
- **HARVEST routing** of both reports' durable (f) rows — awaiting the
  owner go promised in report #1.
- **AGENTS.md "Buildflow health warning" one-liner** for the
  binary-freshness-on-shared-checkouts class — candidate, not written.
- **/tmp overlay artifacts** (`/tmp/exportlegacy-old.go`,
  `/tmp/overlay.json`) — still there (trivial).

## d) TOTALLY FUCKED UP

Nothing broken: no red gates, clean tree, CI green, no lost or reverted
work. Two honest process notes:

- **The close-the-loop miss repeated.** Report #1 flagged two unrulged
  preflight warns; the 17:33 run added a third and I again explained it
  without dispositioning it. Recording it in a report is discharge of
  the record, not of the rule — the warn was in front of me twice and
  both times the disposition (even "deliberate-ignore, owner question
  open") lived only in prose after the fact.
- **The 17:33 verdict I vouched for was cache-sourced.** My "no drift"
  claim rests on composition-matching the 15:52 no-cache run — sound,
  but a 30-second `BUILDFLOW_NO_RESULT_CACHE=1` spot-check would have
  made it airtight and I skipped it as redundant without saying so at
  the time.

## e) WHAT WE SHOULD IMPROVE

- **Make the preflight-warn disposition mechanical**: any warn seen in
  a run output gets its one-line fix/file/ignore verdict in the same
  reply that discusses the run. Two reports in a row have now had to
  retro-list them.
- **A standing triage note for binary-freshness on shared checkouts**
  (AGENTS.md Buildflow-health section candidate): which deltas matter
  (findings plumbing vs docs) and the PATH-shadow check (BuildFlow#81)
  before trusting a reinstall.
- **Cache-hit verdicts quoted to the owner should say so** — "cached,
  matches the last no-cache verification at HH:MM" is one clause that
  removes the only soft spot in an otherwise evidence-backed answer.

## f) Top next tasks

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Owner answers the three standing questions (g) — they gate tasks 3-5 and the upstream ask | Critical | S | Decision |
| 2 | Full `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` run: certify the train, settle baseline-staleness + timing verdicts against a FULL run | High | M | Quality |
| 3 | Characterize `9a6af21..d57a64a` in ~/projects/BuildFlow; rebuild/reinstall the binary if findings plumbing moved (PATH-shadow check first) | Medium | S | Tooling |
| 4 | dprint.json → local wasm plugins (`nix build .#dprint-plugins`), re-run `docs-gates.sh --check` after | Medium | S | Tooling |
| 5 | VACUUM the 1.09 GB buildflow state DB or record deliberate wait-for-GC | Medium | S | Maintenance |
| 6 | Verify-at-source + file the cqrs-lint P012-on-query_only detector ask upstream | Low | S | Tooling |
| 7 | AGENTS.md Buildflow-health one-liner for the binary-freshness class | Low | S | Documentation |
| 8 | HARVEST both reports' durable (f) rows into TODO_LIST on owner go | Medium | S | Documentation |
| 9 | docs/lessons.md entry for the `go test -overlay` negative-proof pattern | Low | S | Documentation |
| 10 | Clean /tmp overlay artifacts | Low | S | Cleanup |

## g) Questions for the owner

1. **(carried)** VACUUM the 1.09 GB buildflow state DB now, or wait for
   the post-purge GC? Live-run contention is unverifiable from this
   session; it is your host-state call.
2. **(carried)** Switch dprint.json to local wasm plugins now
   (prophylactic; rewrites a config the CI docs battery reads) or park
   it?
3. **(new)** The installed buildflow binary is one-plus session behind
   its own repo (9a6af21 → d57a64a, concurrent session). Want me to
   characterize the delta and rebuild/reinstall (PATH-shadow check
   included), or leave the installed binary until that BuildFlow
   session declares its work? The earlier upstream-ask question
   (cqrs-lint P012 detector) also remains open from report #1.

Both reports' (f) sections are HARVEST-ready; routing awaits your go.
