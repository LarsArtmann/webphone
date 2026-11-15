# Status — SUPERB pareto-plan execution (T01–T03, T07, T12–T16, T19–T24) + the concurrent rail train

**Snapshot: 2026-11-15 12:20 CET.** HEAD `b469d03f`; this session executed the owner-GO'd
`docs/planning/2026-11-14_11-41_SUPERB-open-queue-pareto-plan.md` across two working days
(2026-11-14 → 11-15), in parallel with a CONCURRENT sidebar-rail session (its train landed
mid-session and is the attributed owner of the current CI reds — see §d/§e).

## 0. The three numbers

| Metric                              | Value                                                                          |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| Plan tasks fully done               | **15 of 24** (T01–T03, T07, T12–T16, T19–T21, T22, T23, T24, T20-pins-half)    |
| Plan tasks not started              | **6** (T04, T05, T06, T08–T11) + T17/T18 pending + 2 new follow-up rows minted |
| Island suite / tsc at last full run | **344/344 pass · tsc rc=0**                                                    |

## 1. a) FULLY DONE

- **T01 Summarize correctness** — the ended card's Summarize button is gated on
  agent AND ASR AND media (a button that could only 404 no longer renders; pinned by
  `calls-summary-noasr.test.mjs`; the both-on spec boots `asr:true`); the manual
  summarize producer audits through the Server seam (`agent summarize`,
  ok/refused/failed — the manual counterpart of `auto_review`; the idem replay audits
  nothing), pinned by `TestSummaryDraftAuditsTheProducerLine`. CHANGELOG landed.
  Commit `48f5b1f7` (+ daemon-carried code commits).
- **T02 Notes completeness** — `DELETE /api/calls/{id}/note` end to end: the
  `CallNoteDeleted` tombstone (mirror of SnippetDeleted), `NotesGate.Delete` +
  `Count`, the cqrs decider (`cqrs.note_missing` Rejection via the sentinel-wrap
  pattern; the cap Conflict stays 409), the `metaengine.Remove` projection fold, the
  api handler (204/404/503, the notes nudge on success), the route mount, the
  apidocs entry + regenerated `openapi.golden.json` (diff read: exactly the
  noteDelete addition), the error-registry regen (+4 codes). The 409 remedy is real:
  `TestNoteDeleteFreesCapPressure` proves a freed slot admits a new note at the cap.
  History rows gained the glanceable "Noted" chip (empty slot renders nothing);
  Settings gained the fail-open "Noted calls N / 2000" row. Commits `ddce7a24` et al.
- **T03 Agent-pill trio** — per-mode tooltip + aria-label decoding the raw mode names
  (AGENT_MODE_HELP; unknown future modes keep a sane label), the pill re-reads the
  live identity on `wp:drafts-changed` (stale-pill guard; the boot detail stays
  authoritative — the session-opened refresh RACED it and was removed, caught by the
  new spec), the `D` shortcut cycles the mode (typing-guarded, shell-owned so it
  works when island scripts fail), Settings cheat-sheet + `?` help list it, i18n
  en/de. Three new shell specs. Island 344/344, tsc rc=0. Commit `26014b9b`.
- **T07 Sessions at-rest re-review** (subagent) — every 09-30 claim re-verified
  (table shape + login_email ALTER, sweep invariants, one birth path, NO agent_mode
  at rest); residual risk LOW; dated verdict section appended to the review; TODO row
  struck.
- **T12 Schema spike** (subagent) — verdict: reflection walk via `webphone contracts`
  - the existing contracts-fresh gate; deliverable `docs/contracts/config.schema.json`
    (NOT an HTTP route); `x-unset-means-on` for pointer-bools; time.Duration → string.
    Memo written; row struck; implementation row minted.
- **T14 Export-or-confirm design** (subagent) — `webphone export-legacy` dump/check
  sketch (raw read-only sqlite — never store.Open), four-row-class classification by
  `DigitsOf(phone)` under (owner, …), the DROP-window checklist. Memo written; row
  struck; implementation row minted.
- **T19 Deps post-ladder audit** (subagent + code residue) — surface ruled DELIBERATE
  (7→10 gates, each one implementation home, wiring-parity pinned); the 2026-11-08
  collapse design is dead (its engine target was deleted A43). Row closed; the two
  missing `var _` gate pins LANDED in code (`snippets.go`, `dialwindows.go`) and the
  Contacts/Snippets nil-doc drift corrected in `server.go`.
- **T21 HITL memos** (subagent, retried after a rate limit) — D2 veto-window (5th
  stack mode; stack-side held drafts + countdown chip + `agent veto`/`agent expire`
  audit lines; five owner questions parked) and D3 agent-registry stack-contract
  (opaque per-agent bearers, scope ceilings, revocation allowlist ≤60s cache,
  `X-Agent-Name` → `actor=agent:<name>`). Both rows updated.
- **T22 Obs cluster audit** (subagent) — 6 of 11 legs landed (cited), 0 obsolete,
  5 truly-open rows minted; the routed row struck; traces-in-backup obsolete-struck;
  observability-docs row done-struck.
- **T23 appkit flightrecorderhealth** (subagent) — PARK: surface verified in the
  module source (Trigger runs the batch, cooldown-gated, non-blocking; both hard deps
  pinned; NOT dedup — new capability); demand sharpened (dashboard-on wedge or any
  /healthz poll during an incident). Row updated.
- **T13 Legacy readers** — the five dead sibling readers deleted (zero production
  callers since the 2026-11-09 unification; only their own tests read them); the
  app-flag seam (FlagPresent/FlagValue/SetFlag — the sync cursors' ride) was misfiled
  INSIDE `legacy_contacts.go` and now lives verbatim as `appflags.go`; registry
  regenerated for the file moves. Full internal suite green.
- **T15 Errata strikes** — the shipped-claims rows struck with evidence: CSP
  reporting, metrics fencing, secret-leak guard, form-POST bucketing, config-typo
  WARN, §f-less freshness gate, attended-transfer honesty, treefmt-healed CI.
- **T16 Cadence drain** — the two oldest candidates (the island-type-gate mid-flight
  pair) adjudicated + archived gates-first: their campaign COMPLETED same day (26/26
  files, 0 errors, gate enforced), every open §b/§c/§d leg carries a landed verdict,
  inventory tables ledger-marked; cadence healed to 65 ≤ 65; INDEX regenerated. The
  concurrent rail session's freshness debt was healed WITHOUT touching their file
  (their pbx-lens ruling + the usermgmt.db backup question homed as a TODO row).
- **T24 nolint decay** — `erraudit nolint-audit`: 69/69 directives needed, 0 stale;
  calendar intact (next re-measure 2026-12-08); bonus `check --max --no-suppress` =
  0 violations today. Row struck.
- **T20 CRM contract pins (verified-already-shipped half)** — the request-shape pin
  (`context_request_test.go`, P2-2) and the producer-rendered snapshot
  (`context_snapshot_test.go` + `testdata/kith_context_snapshot.json`) were already
  in the tree; the row is struck with that evidence.

## 2. b) PARTIALLY DONE

- **T20 batch-ask half** — the `by-phone?numbers=` draft is WRITTEN
  (`docs/drafts/kith-by-phone-batch-ask.md`, owner tick to file) but uncommitted at
  snapshot time (working tree) and the underlying row stays owner-gated (Kith-side
  work). The draft + strike commit is the next action.
- **Train-end gates** — the battery was green in CI mode at the T15/T16 push; the
  FINAL train battery + `nix flake check` + CI watch for the whole train are still
  owed (the concurrent rail session's reds currently block a clean CI read — see §d).

## 3. c) NOT STARTED

- **T04** summarize tests/docs tail (deleted-transcript idem pin, duplicate
  no-second-push pin, error-context families, error-contract prose, ring_test native
  drafts, FCM SendDraft wire test, server test-harness extraction)
- **T05** smoke boot probes (agent flag grep, notes roundtrip) + lychee excludes
- **T06** typed `settings.textagent.*` module lane + module-check case
- **T08** `buildflow -coverpkg` union coverage step
- **T09** OIDC module VM test · **T10** oauth2-enabled module-check golden ·
  **T11** typed passkey module options
- **T17** architecture D2/SVG refresh · **T18** night-run stability pair
- The two minted follow-ups (not in the original plan): ship `config.schema.json`
  (T12 verdict), implement the §5 export-legacy tool (T14 verdict); T22's five open
  legs; the T19-residue nil-write-guard question (prod-safe, doc-only).

## 4. d) TOTALLY FUCKED UP (owned)

1. **T16 ritual violation**: candidate #2 was `git mv`-ed to archived/ BEFORE its
   gates passed — my `&&` chain piped the gate output through `tail`, which exits 0,
   so the red never stopped the chain. Repaired in place at the archived path (both
   gates then passed), but the gates-before-mv order is the canon and I broke it once.
2. **T03 race, self-caught**: my first design refreshed the pill on
   `wp:session-opened` too — the async identity fetch raced the boot detail and the
   spec failed (`actual: 'AI drafts: full · pinned'`). The listener was wrong, not
   the test; removed (boot detail is authoritative). Lesson: never add a second
   async writer to state a sync event already owns.
3. **T13 trash slip**: `trash internal/store/legacy_contacts.go` deleted the file
   that WAS the app-flag seam (misnamed since birth) — the build caught it
   immediately and the file was restored verbatim as `appflags.go` (the honest name).
   Net effect positive (naming fixed), but the pre-delete grep missed a second
   identity of the file.
4. **T02 test-first misses**: the refusal-arm test initially hit the 404-no-transcript
   arm (forgot to seed the second transcript); the dispatch wrap initially swallowed
   the missing-slot Rejection (`errors.Is` blind) — fixed with the sentinel-wrap
   pattern mirroring the cap's WrapConflictf. Two red→green cycles that a slower
   read of the stack would have avoided.
5. **CI attribution discipline held under pressure**: `5bd34dc0` pushed RED with three
   failures (fast-flake, tw.css freshness, art-dupl) — all three traced to the
   CONCURRENT rail session's markup (`layout.templ` toggle-span clones, their CSS
   classes), none to my surfaces. No fix was attempted on their files (fix-forward
   protocol); the tracker question goes to the owner (§g).

## 5. e) WHAT WE SHOULD IMPROVE

- **`&&` chains must not launder exit codes through `tail`** — gate rc must gate the
  next command (use pipefail or capture before branching).
- **Subagent fan-out worked** (7 parallel adjudications, ~15 minutes wall instead of
  hours) but needs two guardrails next time: (a) read-only subagents CANNOT write
  files in this setup (T07 returned paste-ready text instead of an append — plan for
  the parent to land), (b) one retry lane for rate-limited agents (T21).
- **The daemon's push-leg stalls silently** (>24 h this session); the manual-push bar
  carried the train, but the anchor timestamp check should be a standing first-step
  of every session, not a discovery.
- **Concurrent-session friction is now the norm**: two trains touched shell.js,
  CHANGELOG and the battery in the same hours. The edit-tool freshness trips caught
  every race, but a pre-flight `git log --oneline origin/main..HEAD` + a glance at
  the newest docs/status file would have saved three re-reads.
- **The plan's line-number citations rotted within a day** (the errata rows moved);
  token-based grep is the only stable locator — the plan should say so.

## f) UP TO 50 THINGS TO GET DONE NEXT

| #  | Item                                                                                                                                                      | Impact | Effort |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 1  | Commit + push the T20 draft + strike (working tree now)                                                                                                   | S      | S      |
| 2  | T04: deleted-transcript idem pin (retryability after deletion)                                                                                            | H      | S      |
| 3  | T04: duplicate-replay ⇒ no-second-push pin                                                                                                                | H      | S      |
| 4  | T04: `error-context.py --self-test` families (`push.draft_*`, `readmodels.note_*`, `agent.summarize`)                                                     | M      | S      |
| 5  | T04: error-contract.md prose (`textagent.http` passthrough vs `agent.summarize` wrap)                                                                     | M      | S      |
| 6  | T04: ring_test fakeTransport records drafts natively                                                                                                      | M      | S      |
| 7  | T04: FCM `SendDraft` wire test                                                                                                                            | H      | S      |
| 8  | T04: extract the server test harness (fakeProvider/stubReviewAPI/asrClientFor/recordingTransport) into one helpers file                                   | M      | M      |
| 9  | T05: smoke greps `window.PBX_CONFIG.agent` at boot                                                                                                        | M      | S      |
| 10 | T05: smoke notes API PUT/GET/DELETE roundtrip on a fresh binary                                                                                           | H      | S      |
| 11 | T05: lychee excludes for the two archived doc dirs                                                                                                        | L      | S      |
| 12 | T06: `settings.textagent.{url,secret_file}` module options + LoadCredential                                                                               | M      | M      |
| 13 | T06: eval assertions (both-or-neither) + module-check case + secrets pass                                                                                 | M      | M      |
| 14 | T08: `.buildflow.yml` `-coverpkg=./...` union pass + one recorded run                                                                                     | L      | S      |
| 15 | T09: OIDC module VM test (credential 0400, app reads it)                                                                                                  | M      | L      |
| 16 | T10: oauth2-enabled module-check golden                                                                                                                   | M      | S      |
| 17 | T11: typed passkey module options (derived rp)                                                                                                            | L      | M      |
| 18 | T17: re-derive the package import graph + regenerate the D2/SVG pair (dated snapshots)                                                                    | L      | M      |
| 19 | T18: quiet-window overnight `--track` run script + "stable" definition                                                                                    | L      | M      |
| 20 | T18: ring-guard positive path (a real ring during `--track`)                                                                                              | L      | M      |
| 21 | Ship `config.schema.json` (the T12 verdict row)                                                                                                           | M      | M      |
| 22 | Implement the §5 export-legacy tool (the T14 verdict row)                                                                                                 | M      | M      |
| 23 | T22 open leg f26: rule trace-middleware vs `/metrics` exclusion (skip or wont-implement)                                                                  | S      | S      |
| 24 | T22 open leg f47: arch test — `h.notFound` the only HTML-router 404 writer                                                                                | M      | S      |
| 25 | T22 open leg f49: quarterly `http.NotFound` re-rule cadence row (standing)                                                                                | S      | S      |
| 26 | T22 open leg e6: measure-or-wont-implement the ResponseRecorder sendfile cost                                                                             | L      | S      |
| 27 | T22 open leg e8: the IPv6 `::1` pinned case through `machineSurface`                                                                                      | S      | S      |
| 28 | T19 residue: nil-write-guard question for Contacts/Snippets (doc says degrade, writes panic) — rule it                                                    | S      | S      |
| 29 | Train-end: full battery + `nix flake check` + CI watch to green                                                                                           | H      | M      |
| 30 | Tracker issue for the rail session's three CI reds (layout.templ clones, tw.css staleness) if their train doesn't heal them                               | M      | S      |
| 31 | Rail-train rebaseline: regenerate `.art-dupl-baseline.json` with the CI-pinned detector AFTER their markup lands                                          | M      | S      |
| 32 | Drain the next cadence candidate (22-00 complete-status is now oldest at 19 days)                                                                         | S      | M      |
| 33 | Verify CI GREEN on the tip this train finally leaves behind                                                                                               | H      | S      |
| 34 | Re-check the api-dial tripwire (watch calendar, due 2026-11-17)                                                                                           | M      | S      |
| 35 | F58 tranche #3: adjudicate the 2026-10-28 cohort (5 DIRTY candidates queued)                                                                              | S      | M      |
| 36 | Settings cap row: add the DELETE affordance beside it (the cap row names the pressure; the remedy deserves a click)                                       | M      | S      |
| 37 | Notes DELETE: island-side affordance on the ended-card note row (the API exists; the island doesn't call it yet)                                          | M      | S      |
| 38 | `webphone_island_latency_seconds` dashboards: fold the new summarize/refresher legs into the review                                                       | L      | S      |
| 39 | Re-run `scripts/ui-capture.py` after the rail train settles (26-shot matrix owns the visual verdict)                                                      | M      | L      |
| 40 | CHANGELOG `[Unreleased]` audit for the whole train (T02/T03 lines landed; T13/T15-T16 docs-only stay silent per the bar)                                  | S      | S      |
| 41 | AGENTS.md: note the `appflags.go` rename + the deleted-reader state in the storage-ladder bullet                                                          | S      | S      |
| 42 | AGENTS.md: the `D` shortcut + pill-refresh belong in the shell-contract bullet                                                                            | S      | S      |
| 43 | FEATURES.md: the notes DELETE + Settings cap row + pill semantics inventory lines                                                                         | S      | S      |
| 44 | Consider `AgentModes` tooltip localization (pill copy is D3-English; the island-localized vocabulary was retired — confirm that's still the owner's call) | L      | S      |
| 45 | The T12 follow-up's contracts-fresh wiring order (schema before golden) when picked up                                                                    | S      | S      |
| 46 | Kith e2e (#9) + the batch-ask filing — owner tick                                                                                                         | M      | S      |
| 47 | usermgmt.db backup coverage: the verification half of the homed lens-analysis debt (§d3 of that report)                                                   | H      | S      |
| 48 | Retire the plan's stale line-number errata table (T15 drained it; the plan file should point at the strikes)                                              | L      | S      |
| 49 | Post-train `buildflow` full run (the battery-on-commit rule wants the working tree clean first)                                                           | M      | M      |
| 50 | Next session starts with: `git log origin/main..HEAD`, newest docs/status glance, daemon push-leg check                                                   | S      | S      |

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

1. **The rail session's three CI reds** (layout.templ clone pairs, tw.css/login.css
   staleness, the art-dupl baseline): heal-them-myself with a rebaseline + artifact
   rebuild, or does their active train own its own healing? I have NOT touched their
   surfaces per the fix-forward protocol; main stays red until one of us moves.
2. **Queue order after this report**: the two minted follow-ups (ship
   `config.schema.json`, implement the §5 export-legacy tool) versus the plan's
   remainders (T04/T05 test armor first, then the nix cluster T06/T09–T11) — which
   half do you want first next session?
3. **The `api-dial` tripwire revisit is due 2026-11-17** (the standing watch
   calendar): take it inside this train's tail, or leave it to the calendar owner on
   the day?
