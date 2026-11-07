# Status — M20 shipped, main healed after the concurrent json/v2 break (2026-11-07 19:2x)

**Snapshot-rot header: claims verified as of 2026-11-07 19:2x CET.** The
heal push `2c50b2c7`'s CI run (42400725432) was still in flight at
writing; its verdict is the one claim this report cannot yet make.
Everything else below is verified in-tree or by command output this
session. The daemon owns the commits; one manual push (2c50b2c7) rode
the ratified red-main-with-verified-fix bar.

## a) FULLY DONE

| # | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Verification                                                                                                                                                                                                                         |
| - | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **M20 shipped** (the one open harvested frontend-ux train row): voicemail reload failure got its own copy (`vmLoadFailed` — the 401 arm keeps `vmAuthFailed`), and the call-history + contacts loads announce a warn toast on both the `!res.ok` arm and the thrown-fetch arm (`historyLoadFailed`/`contactsLoadFailed`, en+de keys) — M20's review finding was "voicemail reload failure has no distinct copy; history/contacts loads swallow !res.ok silently"; both halves closed                                                           | commit `5c368085`; panels-load-failures.test.mjs 6/6; island suite 318/318; tsc green; no served-bytes pin collisions (grepped smoke + Go tests); CHANGELOG `[Unreleased]` line added                                                |
| 2 | **TODO row closed**: the M20 residue row in TODO_LIST is `[x]` with the commit + pin evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                  | TODO_LIST frontend-ux residue block                                                                                                                                                                                                  |
| 3 | **Cadence question (§g Q1 of the 11:07 report) answered by events**: the v9 docs-health sweep (18:32) drained the cadence to green with gated archives — `docs-gates cadence .` reads 70/70 at the pin, oldest report 32d < 45d, exit 0; the "missing archive-gate verbs" premise from the 11:07 session didn't hold for the build v9 used (the canonical-binary question stays routed as its own TODO row, not blocking)                                                                                                                      | `docs-gates cadence .` exit 0 this session                                                                                                                                                                                           |
| 4 | **Ledger rows for the train's settled reds** (§g Q2 substance): 42343813135, 42352673965, 42358288828 (+sister 42358387321) recorded as the same two deterministic lanes (stale login.css after app.css edits + unbaselined clone groups), healed by `8a090ac5`, green successor 42362014234; row updated with the evening heal                                                                                                                                                                                                                | docs/ci-flake-ledger.md, commits `89c89030` + `d0ab66bb`; §g Q2's "did the watcher file trackers?" answer: NO issue exists for those four runs (gh issue list verified) — #236 exists only for the separate sso-race run 42308961304 |
| 5 | **Main healed after the concurrent json/v2 break**: the concurrent session's daemon sweep `43f70b9c` (18:49) moved sso_test.go onto `encoding/json/v2`, whose strict Marshal ERRORS on the unexported-only `domain.Extension` — deterministic go-tests red (run 42398615048, "--- FAIL: TestLoginClaimsAuditLineCarriesGroupsOnlyInTheJournal"). Fixed WITHOUT touching their file: `MarshalJSON` on domain.Extension rendering the raw value (the id.ID doctrine; `authmap.User` is never JSON-serialized in prod, so prod wire is untouched) | commit `2c50b2c7`; sso 3x+5x `-race` green locally; full `go test -race ./...` exit 0                                                                                                                                                |
| 6 | **Issue #236's race fixed** (same commit): the audit-line journal test swapped the PROCESS-GLOBAL slog default to its own bytes.Buffer while `t.Parallel()` let sibling sso tests log through it — the data race run 42308961304 detected (read at sso_test.go:605 vs concurrent Buffer.Write). The test now runs sequentially; race repro window eliminated by construction                                                                                                                                                                   | commit `2c50b2c7`; issue close comment pending the successor run verdict (voice profile followed)                                                                                                                                    |
| 7 | **Host-env gotcha recorded** (cost ~30 min of forensics): the agent shell exports `GOEXPERIMENT=jsonv2`, mkShellNoCC inherits it, and encoding/json flips to v2 semantics LOCALLY ONLY — json.Marshal errors on unexported-only structs where CI passes. AGENTS Commands section now carries the neutralization (`env -u GOEXPERIMENT`) and the local-vs-CI divergence rule                                                                                                                                                                    | AGENTS.md, commit `d0ab66bb`                                                                                                                                                                                                         |

## b) PARTIALLY DONE

1. **Issue #236 close comment** — drafted in the house verdict style,
   held until run 42400725432 goes green so every claim in it is true
   when posted (verify-before-claiming; the ledger row for the race
   itself also waits on that verdict).
2. **The §g trio** — Q1 (cadence) dissolved by events; Q2 (tracker
   issues) answered with evidence above; Q3 (ui-shots eyeball vs
   pixel-diff) stays genuinely owner-only and is routed to the owner in
   this session's final message. No TODO row minted for Q3 — the
   question lives in the 11-07 report §g and now in the owner briefing.

## c) NOT STARTED (correctly out of mandate)

- The concurrent json/v2 migration session's remaining files — their
  sweep, their train; only the deterministic main-red they left behind
  was healed, minimally, in their spirit.
- The 50-item §f list in the 11-07 resumed-completion report — rides
  TODO row 20's harvest; not this session's scope.

## d) The fuck-up

One, small and self-caught: I started editing TODO_LIST's M20 row
without re-Viewing after the daemon touched it (mtime guard tripped
twice before the edit landed). No damage — the guard worked as
designed. The evening's real near-miss was process, not defect: the
first M20 commit was swept by the daemon mid-`git commit` (race
between my staging and its sweep); content survived intact in
`5c368085` but the narrative message was lost to a `chore:` header.
Lesson re-confirmed: stage+commit in ONE call, and the daemon can win
even that race.

## e) Environment notes for the next session

- The GOEXPERIMENT leak (AGENTS, § Commands) is live in THIS tool
  shell — check `env | grep GOEXPERIMENT` if local-vs-CI suites
  disagree.
- The concurrent json/v2 session was last heard from at 18:57 (daemon
  sweep); its migration may continue — sso_test.go, authmap, and the
  other `*_test.go` files it touched are ITS files; the domain-side
  MarshalJSON added here is deliberately in `internal/domain/ids.go`,
  not in their test files.
- Daemon push leg stalled ~35 min again this evening (66ae9184 18:57
  → manual push of `2c50b2c7` 19:2x under the red-main bar; `d0ab66bb`
  left for the daemon).
