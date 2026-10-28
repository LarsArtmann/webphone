# Status — SUPERB session 3 (2026-10-28 ~03:00 → 10:52 CET)

Claims verified as of **10:52 CET**: `main...origin/main [ahead 1]`, CI run
40666424112 (09:43, on `3584652`) is **RED** — see §d. The 09:49 push's CI
run was still in progress at snapshot time. One foreign in-flight edit sits
in the tree: `scripts/pre-commit-guard-selftest.sh` (a concurrent session
extending the hook guard — not mine, untouched).

Supersedes the 01:43 report. Continuing the ratified 27-task plan
(`docs/planning/2026-10-27_17-25_SUPERB-issue-triage-execution.md`).

## a) FULLY DONE

Carried from sessions 1–2 (verified green): red-main heal (`4c39734`), T3
history disposition filter, T4 Kith contract + issue #6 comment, T7
Settings CRM tri-state, T8 orphan guard (`2f3ca80`), T9 pre-commit drift
gate (`c2ef67a`), T10 daemon forensics (`4cb55e0`), T20 baseline audit
(`de2e517`), T21 push-stall modeling (`0004eb7`).

This session:

- **T11 closed out.** The apidocs entry for `POST /api/calls/{id}/review`
  (`7806355`), the M66 server suite in `internal/server/review_api_test.go`
  (6 tests: happy 202 + idem replay, seam-off/scoping 404s, nil-store 503,
  oversized-id 400, outage-retryable, in-flight guard dedup), and — found
  by writing the tests — a **real 200-vs-202 bug**: the handler answered
  200 while the island checks `status === 202`, so every successful
  dispatch would have shown "failed". Fixed with `WriteHeader(202)` on all
  three answers. Island suite 261/261 green, tsc green, full `go test ./...`
  green at close.
- **T12 — `call_reviews` artifact slot.** `domain.ReviewState`
  (pending/present/failed) + `CallReviewSaved` fact; readmodels
  `call_reviews` collection (one slot per owner+call, replace semantics,
  2000/owner reject-cap like notes); `Reviews.Get/Set`; wired through
  `api.Deps`, `server.Deps`, app composition root; `runReviewJob` writes
  pending→present/failed, the manual summarize path writes the same
  vocabulary; `GET /api/calls/{id}/review` (200-with-artifact,
  empty-body-is-absent, disabled-seam 404) + apidocs entry + drift row;
  store + API lifecycle tests; error codes `readmodels.review_*`
  registered same-session; family pin added.
  **Bonus bug found and fixed while wiring:** the composition root never
  set `Notes:` — production note endpoints answered 503 since M16 shipped
  (test harness set it, `internal/app` didn't). One-line fix, verified by
  the existing note tests now running against a non-nil store.
- **T13 + T14 — History render + not-configured chip.** `TranscriptGroup`
  grew `ReviewState/ReviewText/ReviewConfidence`; `history.templ` renders
  the artifact under the transcript text (present = draft text +
  confidence %, pending, failed), and a row with no review renders the
  "AI review not configured" chip ONLY when the seam is off — with the
  seam on, silence means nobody asked. 5 i18n keys × 2 maps, 4
  `AgentEnabled` wirings in `historyPanel`, `wp-call-review` CSS (logical
  properties after the lint caught `border-left`), view byte-pin test
  (`TestHistoryTranscriptReviewStates`). Committed as `e70f90f` + sweeps.
- **T15 — model-routing decision doc.**
  `docs/planning/2026-10-28_04-45_T15-model-routing-decision.md` (`e08126d`):
  the review drafts through the existing textagent seam (one drafting
  path, one idem key, loopback = the local default); a dedicated client
  has no consumer; three reopen triggers recorded.
- **T16 server side — dial-cooldown gate (code complete, locally green).**
  `domain.DialWindowKey/DialAttempted/DialOverrideSaved` + expunge facts;
  readmodels `DialWindows` (windows + one-shot overrides collections,
  check+record under one mutex, denied asks never record — a retry loop
  cannot extend its own window, cap 500/owner with transcripts-posture
  trim); `dial.cooldown` config block (`time.Duration`, negative rejects
  via `config.dial.cooldown`); `POST /api/dials` gate handler (404 when
  off, 503 nil store, 400 bad remote, 502 store failure = island dials
  anyway); `dialCooldownSeconds` in `/config.js`; routes + Deps + app
  wiring. Store tests: window deny/expiry, one-shot override (spent by
  the next attempt, follow-up denies again), restart survival, owner
  scoping — all green locally.

## b) PARTIALLY DONE

- **T16 landed RED (see §d).** Server code + store tests are in `main`
  (swept by the daemon at 10:35–10:50), but: the `readmodels.dial_*` error
  codes were never registered (`-update` not re-run after T12's), the
  route has **no apidocs entry** (the drift test class I fixed for T11 —
  repeated anyway), and `nix fmt` was not re-run after the last wiring
  edits. CI 40666424112 failed on `TestErrorCodeRegistryIsFresh`,
  `treefmt-check` (server.go alignment), and island-js microphone device
  check (possibly concurrent-session or env flake — my island files were
  untouched at that commit; the other session's 10:49 addendum says their
  work is not the cause either).
- **T17 — island cooldown UX: designed, zero code.** Full design exists
  (gate ask in `placeCall` before the INVITE, fail-open on 404/transport,
  deny renders a JS-created confirm chip with a "dial anyway" override
  that re-enters `placeCall(raw, {override:true})`, i18n keys, typing in
  `types/globals.d.ts`, test stubs patterned on `calls-summary.test.mjs`).
  My first island edit was cancelled mid-flight; nothing landed. The
  contacts-row server-rendered chip (M96) is unbuilt and I recommend
  deferring it behind the deny-time chip (see §g).
- **T18 — hardening tests: store layer done, API+island layers missing.**
  Window/expiry/override/restart/scoping pinned in
  `dialwindows_test.go`; no API-level test for `POST /api/dials` (404-off,
  400-bad-remote, 502-fail-open, override flow), no opt-in-off
  byte-identical pin, no island tests.

## c) NOT STARTED

- T19 external-claims verification (dialer PRs #20/#54, Kith issue, pbx
  open-issue review) — blocked on `gh auth` (still 401; interactive login
  needed).
- T22–T24 pbx-artmann work (CA trust, token provisioning, e2e verify) —
  tri-repo ritual applies; the stack repo still carries my unpushed
  `d34ea70` bridge note from session 2.
- T25–T27 closeout: docs polish, triage labels/templates, HARVEST,
  M18 tail refresh, FEATURES/CHANGELOG on #6, **plan errata rows** (T12
  PUT-endpoint deviation — no consumer, GET-only shipped; T17 chip
  placement pending §g; M62 95s budget; M19-in-Go; M22/M23
  obsolete-by-design).
- Stack browser E2E re-run (owed by T3/T7/T11/T13 served-markup changes).
- Debt: CHANGELOG entries for T7/T11/T12/T13/T16, AGENTS.md row for the
  review route + `AuditAgent` seam + the island i18n test gap + the
  `Notes` prod-wiring bug class, dedup sweep after the trains, status
  index update, CI-watch verification for this session's pushes.

## d) TOTALLY FUCKED UP

1. **I red-ed main by re-committing the exact defect class I had fixed
   six hours earlier.** For T11 I documented and executed: apidocs entry +
   registry `-update` + fmt before the gates. For T16 I built the route,
   ran the tests green, and then let the daemon sweep the files WITHOUT
   the apidocs entry, WITHOUT re-running `-update` for the seven new
   `readmodels.dial_*` codes, and WITHOUT a final `nix fmt`. Three gates
   violated in one landing. The AGENTS rule is explicit ("register IN THE
   SAME COMMIT — the 2026-10-24 red-main class") and I was its most recent
   beneficiary. Root cause: I treated "tests pass locally" as the finish
   line instead of "the landing checklist ran at the tip".
2. **Repeated narrative-commit losses to the daemon.** T11, T12, T13, T16
   all lost their stage+commit races despite the one-call discipline from
   sessions 1–2. Content survived (verified in HEAD each time) but the
   history is daemon-chauffeur noise, and the sweep-then-red pattern is
   exactly how the registry drifted: my commit-time checklist never ran
   because my commit never ran.
3. **`validatePush` collateral edit.** One multiedit silently deleted the
   ntfy token/URL both-or-neither check; I caught it on the next read and
   restored it before any test ran. A multiedit whose new_string omits
   half the old_string's content should never have been constructed —
   that is the session-2 data-loss class one step from shipping.
4. **A cancelled island edit as a stopping point.** My T17 first edit was
   cancelled; I had no re-entry plan written down, and the session ended
   on that state rather than on a green, committed boundary.

## e) WHAT WE SHOULD IMPROVE

- **Landing checklist as code, not memory.** New-route trains need a
  mechanical pre-commit sequence: `nix fmt` → registry `-update` (when any
  `errorfamily.New*` was added) → apidocs entry + drift row → gates →
  stage+commit in ONE call at the tip. I will run this verbatim on T16's
  heal and every future train; a `scripts/landing-checklist.sh` or
  buildflow step would make it non-memory.
- **Never trust "last fmt".** Any edit after `nix fmt` invalidates it; the
  fmt must be the LAST write before staging, every time.
- **Commit BEFORE the daemon sees the files.** The one-call discipline is
  necessary but not sufficient — the sequence must be: finish → fmt →
  registry → apidocs → targeted tests → stage+commit+verify-markers within
  one tool call, ideally within the daemon's 2–5 min window.
- **Test windows must respect storage granularity** (the 50ms cooldown
  test vs unix-second stamps — flake-by-design). Record granularity next
  to the column, and pick test windows ≥ 10× the tick.
- **The `Notes` prod-503 bug class deserves a pin**: every new `server.Deps`
  field gets a same-train app.go wiring + a boot-level assertion (or a
  test that reflects over Deps construction in app.go). I found it by
  accident; a one-off test would find the next one on purpose.
- **T17 island work should start from the cancelled edit's design notes**
  (this report §b) rather than re-deriving them.
- **CI-watch was passive this session** — the red sat from 09:43 to 10:49
  until the other session's addendum. The manual-push-bar watcher logic
  should extend to "CI red on a commit containing my files".

## f) NEXT 50 (ordered; the first 12 are the actual queue)

1. **Heal the red main**: re-run `go test ./internal/arch -run
   TestErrorCodeRegistryIsFresh -update` (7 dial codes), add the
   `POST /api/dials` apidocs entry + drift row, `nix fmt`, full gates,
   one-call commit at the tip, CI-watch to green.
2. Re-run `nix build .#checks.x86_64-linux.island-typecheck` +
   island-lint + island-js locally to reproduce/attributed the microphone
   device-check failure before assuming flake.
3. Write the `POST /api/dials` API tests (T18): 404-off, 400-bad-remote,
   503-nil-store, 502-fail-open, allow-records, deny-unrecorded,
   override-arm+consume.
4. Opt-in-off byte-identical pin (T18): zero-config boot has no
   `/api/dials` route and no config.js key.
5. Implement T17 island: `consultDialGate` (fail-open, 3s abort,
   404-caches-off) + `placeCall(raw, {override})` + JS-created confirm
   chip in ui.js + i18n keys ×2 + `types/globals.d.ts` dialCooldownSeconds.
6. Island tests (M100): deny → chip renders + no INVITE; override →
   INVITE; gate off → direct INVITE; 502 → INVITE (fail-open).
7. Morph-safety pass on the chip (stable ids, no re-render tear) (M101).
8. Family pins for the dial codes (`readmodels.dial_*`) in
   `readmodels/family_test.go` + arch registry check green.
9. T16 gates + ONE narrative commit (post-heal) + CI-watch.
10. Decide T17 contacts-row chip with owner (§g Q2); implement if GO.
11. T19: `gh auth login` (owner-side), then dialer PR #20/#54 claims,
    Kith issue location, pbx open-issue review, verification notes.
12. T22–T24 pbx-artmann: dnsblockd CA into `security.pki`, CRM token
    provisioning + `crm.token_file` wiring, stack e2e verify — with the
    tri-repo ritual (push webphone BEFORE stack re-pins; push the stack's
    `d34ea70` first).
13. T25 docs polish A: TODO rows → stable-name citations; issue-body
    corrections; comment render check.
14. T26 docs polish B: triage labels; adoption-candidate issue template;
    issue-review runbook; concurrent-session habit row.
15. T27 HARVEST: plan-born items → TODO_LIST; M18 tail refresh;
    FEATURES/CHANGELOG for #6–#8; **plan errata** (T12 GET-only, T17
    placement, M62 95s, M19-in-Go, M22/M23 obsolete).
16. Stack browser E2E re-run (T3/T7/T11/T13 markup).
17. CHANGELOG `[Unreleased]` lines: T7 tri-state, T11 auto-review + 202
    fix, T12 artifact + Notes-wiring fix, T13/T14 render, T16 gate.
18. AGENTS.md rows: review route + `AuditAgent` seam; dial-gate seam +
    config; the `Notes`-wiring bug class + Deps-wiring pin idea; island
    i18n test gap.
19. Status index update + mark this report's predecessor superseded.
20. Dedup sweep (`art-dupl -t 1`) after the #7/#8 trains; new baselines
    with the CI-pinned v0.7.2 detector.
21. Re-verify T3 semantics survived the concurrent dedup commit
    (`15a5a8d` reworked my history selects) — filter + disposition pins.
22. Verify the concurrent `go-health v0.8.1` bump's tail landed end-to-end
    (vendorHash + CI green on that train).
23. Attributed the island-js microphone NotAllowedError (flake vs real).
24. Merge/verify the concurrent session's in-flight
    `pre-commit-guard-selftest.sh` + `whitespace-drift.sh` extension once
    it lands (selftest still 3/3 after their edit).
25. Add the `server.Deps` ↔ app.go wiring parity test (the Notes bug
    class) — reflect over `server.Deps` fields, assert each non-optional
    one is set by the app composition root.
26. Add the missing island i18n key-existence test (the T11 red class:
    `t("…")` literals are uncovered).
27. Consider a `scripts/landing-checklist.sh` (fmt → registry → apidocs
    grep → gates) wired as a pre-push hook.
28. Refresh `docs/dom-contract.md` if T13/T17 markup ids enter the served
    page (the review block is server-rendered — confirm contract parse).
29. Re-run `nix flake check` end-to-end once main is green (last full run
    predates T13/T16).
30. Smoke-boot a fresh binary (`scripts/webphone-smoke.py`) covering the
    review + gate surfaces loopback.
31. Backup drill sanity: `readmodels.db` now carries dial_windows +
    call_reviews — confirm the drill asserts them or note as
    acceptable-by-scheme.
32. Update `docs/error-contract.md` cross-reference for the dial-gate
    family (handler codes) — after the `-update` heal, verify the table
    rows name `dialgate` sites.
33. Check `/openapi.json` render after the dial entry lands
    (`webphone contracts` untouched by /api/dials? no — it's
    session-gated JSON API, so apidocs only).
34. Consider making `placeCall`'s gate ask skip-able for internal dials
    (redial from history?) — UX ruling needed if the E2E dials twice.
35. Confirm the `history.review.*` de strings render (grep the served de
    page once E2E runs).
36. Write the T16/T17 sections into FEATURES.md (PLANNED → DONE moves).
37. Verify `makeModule`/island import closure unaffected (no new island
    module files planned — confirm).
38. Re-check `TestServedPageSatisfiesStrictCSP` after any JS-created
    chip (no inline styles — use classes only).
39. Sweep for other `deps.X` fields added since the Notes bug and not
    wired in app.go (AuditAgent? Dials done; Reviews done).
40. Reply to/resolve the open owner questions from the 01:43 report once
    answered; fold decisions into the plan errata.
41. Quarterly erraudit/branching-flow re-measure due 2026-11-05 — note the
    new codes were added with family pins (tier-2 stays 0).
42. Consider `dial.cooldown` NixOS module option + module-check stand-in
    (the module lags the app config; same for `textagent` if unwired).
43. Confirm `/hooks/ring` and gate endpoints sit in the right limiter
    buckets (gate shares contacts — one POST per user action; OK).
44. Documentation: README § Passkey/Settings mention of the review chip
    if user-facing wording needs it (probably plan-level, skip unless).
45. Add `dial.cooldown` to the config example/docs block (README env
    table) if one exists.
46. Double-check `DisallowUnknownFields` on the gate body won't reject
    the island's real payload (it sends exactly {remote, override}).
47. Prune `safeTail`'s export shape if dedup flags the shared tail
    helper (private, fine — check art-dupl verdict).
48. Update TODO_LIST: move #7/#8 rows to SHIPPED-pending-E2E; add the
    §f queue's 1–12 as bounded tasks.
49. After green: manual-push-bar check (`scripts/push-stall-watch.sh`)
    before any manual push.
50. Schedule the M18 tail refresh (T27 dependency) — dispositions vocab
    unchanged, but the review states join the unknown⇒failure ruling doc.

## g) QUESTIONS FOR THE OWNER (3)

1. **T16/T17 visibility split:** the deny-time confirm chip ("dialed
   recently — dial anyway") fully covers #7's acceptance bar;
   the typeahead/contacts-row server-rendered chip (M96, ~90 min) is
   progressive visibility at the cost of a per-render store read on the
   contacts panel. GO on the row chip too, or is deny-time-only the
   right size for v1?
2. **The island-js `microphone` device-check failure** on the red run: my
   commits touched no island files; the other session denies ownership
   too. If the heal-push re-reds on it, do you want me to dig into the
   check's environment (KVM/headless mic emulation) or treat it as fleet
   infra and file it to the runner?
3. **Deviation batch, still open from the 01:43 report plus new:** (a)
   keep the 95s (not 15s) draft budget; (b) keep T12's GET-only artifact
   API (the plan's PUT has no consumer; replace semantics live in the
   store seam); (c) accept the daemon-swept commit narratives for this
   session or should I rebase-history-tidy before the stack re-pins? All
   three default to "keep as shipped" if unanswered.
