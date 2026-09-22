# Status: v2.6.0 train mid-flight — CRM hardening shipped, release blocked on concurrent session

2026-09-22 21:12 CEST. Session started from the pasted TODO_LIST sweep
(2026-09-22 night). One line summarized: **the release train's fold is
committed and pushed; all four dedup-train contract pins plus the CRM
disabled-policy/single-flight/counters work are implemented and green
under `-race`; the release run itself is parked because a concurrent
session is actively landing a UX raw-ideas train in the same tree, and
two release.sh precondition runs correctly refused to run.**

## What was forgotten / could be better (asked-for critique)

- **Forgot to verify the push state before release attempt #1.** The
  runbook itself says "verify with ls-remote, never push logs". The
  fold commit was local, the push daemon was stalled (known mode,
  precedent `e57faf6`), so step 1 failed with "main diverged from
  origin/main". One `git ls-remote` before launching would have caught
  it. Fixed by hand-pushing, then re-running.
- **Forgot the concurrent-session protocol when sequencing.** AGENTS
  documents live multi-session work; I launched release attempt #2
  anyway, which failed cleanly on "tree not clean" (their in-flight
  `shell.js`/`messages.templ`, later `typeahead.js`). No harm done —
  the script fails fast — but the correct move was to define the
  quiet-point policy up front: poll the tree, fold their landed work
  into the 2.6.0 changelog, then one release run.
- **Wrote a racy test first.** The listRows iteration-failure pin
  originally used the context-cancel-mid-iteration trick; it failed
  (nil error — modernc drains before the cancel lands). Replaced with
  a deterministic `RegisterScalarFunction` UDF that fails on the second
  call. Should have reached for fault injection immediately.
- **Shipped a draft with pseudo-helpers once.** First `db_test.go`
  write referenced non-existent `cancelCtx()`/`testCtx()`. Caught by
  the compiler, but sloppy.
- **My own counter test asserted the wrong resolver** (expected the
  failure on `resolver` while the dead lookup ran on a second one) —
  test bug, not code bug; restructured.
- **Three edit round-trips wasted on the mtime guard.** A watcher
  (formatter/buildflow daemon) touches file mtimes; I should re-read
  immediately before every edit instead of reasoning from an earlier
  read.
- **`job poll`** — typo'd a shell command that doesn't exist. Noise.

## a) FULLY DONE

1. **Fold (release runbook step 1)**: `CHANGELOG.md` [Unreleased] →
   `## [2.6.0] - 2026-09-22` (kept an empty Unreleased on top, matching
   the v2.5.0 fold shape). Completed the link refs the v2.5.0 fold
   forgot: added `[2.6.0]` and `[2.5.0]` tag links, moved the
   `[Unreleased]` compare to `v2.6.0...HEAD` — the lychee gate can
   resolve them. Committed as `711fff5` with a narrative message and
   pushed by hand after the stalled daemon left it local.
2. **Dedup-train contract pins — all four requested tests shipped**:
   - `internal/pbx`: `TestNilClientReturnsErrDisabled` — a nil
     `*Client` short-circuits every method (incl. `VerifyCredentials`)
     to `ErrDisabled` without panicking; proves the `do()` chokepoint
     covers the "Deps field nobody wired" case.
   - `internal/store/db_test.go`: `TestListRowsErrorShapes` — pins the
     helper's documented error contract: query failures wrap the op
     (`walk nums: …`), ITERATION failures wrap the op-rows shape
     (`walk nums rows: …`) via a deterministic modernc UDF
     (`wp_test_boom_second_call`), scan failures pass through
     UNWRAPPED (deliberate contract, now pinned so nobody "fixes" it by
     accident), and empty result sets return non-nil empty slices.
   - `internal/store/owner_scoping_test.go`: added the `ListThreads`
     owner/foreign subtest — the one `listRows`-driven list method the
     owner-scoping suite did not cover.
   - `internal/server/send_form_test.go`:
     `TestRequireMultipartToAnswersPerTab422` — drives BOTH send routes
     (`/messages/send`, `/fax/send`) with an undialable "to" (`&&&`;
     letters are deliberately VALID — short codes — so "not-a-number"
     was a wrong first guess that the probe run exposed) and asserts
     422 + the per-tab i18n reason, with negative assertions that
     neither tab's reason leaks into the other's body.
3. **crm.Client disabled-policy split brain RESOLVED** (`internal/crm/
   client.go`): both methods now funnel through one `do()` chokepoint
   with the same shape as `pbx.Client.do` — Enabled guard first (no URL
   built when disabled), single URL-join home, bearer header cannot be
   forgotten, Content-Type only with a body. Wire paths, status
   mapping, and sentinel errors unchanged; the package's tests pass
   unmodified except the strengthened nil-client pin in
   `TestClientDisabledByDefault`.
4. **crm.Resolver single-flight** (`internal/crm/resolver.go`):
   concurrent `Resolve` calls for one number share ONE upstream
   lookup (leader + `lookupFlight` joiners, in-flight map retired on
   settle). Waiting callers honor their own ctx (bail out without
   corrupting the leader's flight); cache semantics untouched
   (failures still uncached); misses/failures still degrade silently.
5. **crm.Resolver lookup counters**: `hit`/`miss`/`failure` atomics
   counting UPSTREAM round-trips only (cache hits don't count — they
   aren't CRM traffic), per-resolver (no cross-talk), nil-safe
   `LookupCounters()`.
6. **New crm tests, all green under `-race`**:
   `TestResolveSingleFlightsConcurrentLookups` (10 racers, one number →
   exactly 1 upstream hit, counters 1/0/0) and
   `TestResolverLookupCounters` (hit+miss+failure classification,
   cache-hit-doesn't-count, per-resolver isolation).
7. **Repo hygiene**: all work is committed (the auto-daemon swept the
   tree clean at report time); nothing of mine is dangling uncommitted.

## b) PARTIALLY DONE

1. **/api/calls idempotency (CRM follow-up b)** — design settled, code
   not written yet: body gains `key` (island-generated UUID);
   handlers gain `callsIdem *idemStore` (the `hooksIdem` mechanism,
   own 1h TTL const); key namespaced by session extension; seen-check
   placed AFTER the CRM-enabled short-circuit but BEFORE the resolve
   (a replayed fact short-circuits to 204); record ONLY on success
   (both the unknown-number 204 and the logged 204; the 502 stays
   retryable — same fail-retryable contract as the status hooks).
   The OpenAPI `/api/calls` schema needs the `key` property.
2. **Island half of idempotency**: `recordCrmCall` (panels.js:48) must
   add `key: crypto.randomUUID()` per finished call; the Terminated
   branch (calls.js:382) fires once per call, and the key also covers
   wire-level fetch retries. Not touched yet — deliberately last among
   island edits because the concurrent session is actively in island
   files (`typeahead.js` at last check).
3. **/metrics rendering of the counters** — design settled: three
   labeled `webphone_crm_lookups_total{outcome="hit|miss|failure"}`
   counter lines via the existing `writeMetric`, rendered only when
   the CRM is enabled (keeps the aggregates-only contract trivially).
   Not implemented.
4. **Island unit test for `recordCrmCall` (CRM follow-up a)** — located
   the subject and its seams (config `crmEnabled` flag, `authedFetch`,
   `announce`/`log` paths); panels.test.mjs addition not written.
5. **Release train v2.6.0** — fold done; release.sh has NOT completed
   (see d/f for why) — no tag exists yet, so the 2.6.0 section can
   still absorb my CRM hardening and the concurrent train's landed
   work.
6. **CHANGELOG 2.6.0 addendum** for this session's hardening
   (chokepoint, single-flight, counters, idempotency key, four pins) —
   drafted here, not yet written into the file (it should land
   together with the idempotency implementation so the section is
   written once, honestly).

## c) NOT STARTED (all remaining TODO rows)

- Full-suite verification of tonight's code changes (`go test
  -count=1 ./...`, buildflow no-cache, island node tests, `nix flake
  check`, smoke `--expect-version 2.6.0`).
- pbx-artmann relock #4 (rev swap → `nix flake update telephony` →
  lock-drift-probe → both toplevels → ExecStart-moved check).
- Everything OWNER-terminal, untouched by design: prod deploy of the
  released chain, post-deploy probes (incl. the rejection-banner
  check), the telnyx-webhooks journal/root-cause for the outbound SMS
  422, the OWNER-calls batch (~15 decisions), announcement posting +
  disclosure posture, the oops non-fix ratification, T26b's stack
  half, quarterly standing watches (next due 2026-12-20).
- Send-failure UX (C) pre-flight self-send fast path (OWNER call),
  (E) provider refusal → 422 + honest log family (test + error
  contract + stack runbook must move together), (F) conditional.
- Theme-knob verifications: FOUC screenshot pair, German-native copy
  review, composer-affordance screenshot QA.
- CRM follow-ups (e) stack secrets wiring, (f) stack-runbook
  cross-doc, (g) restore-drill, (h) the island→server integration
  level decision.
- TODO_LIST harvest (rows for the pins + CRM follow-ups a–d close
  only after the full gates run green), AGENTS.md CRM-seam paragraph
  update (chokepoint is now SHARED — the split-brain note is stale the
  moment this ships), FEATURES.md CRM row notes.

## d) TOTALLY FUCKED UP

Nothing destructive: no reverts, no lost work, no broken main, no
data risk. The honest list of self-inflicted damage: one release.sh
attempt wasted on an unpushed fold (my sequencing), one on the
concurrent session's tree (unavoidable, but I could have predicted
it), one compiler-caught draft with phantom helpers, one racy test
design that failed before I replaced it with the UDF approach, one
test that failed due to MY wrong assertion, and a handful of mtime-
guard round-trips. All fixed; every package I touched is green under
`-race` at package scope.

## e) WHAT WE SHOULD IMPROVE

1. **Release precondition checklist before launching release.sh**:
   `git ls-remote origin main` == local HEAD, tree clean, AND a
   concurrency check (no uncommitted foreign files; if a concurrent
   train is active, wait or coordinate).
2. **A written quiet-point protocol** for concurrent sessions: who
   folds whose work into the release changelog, and what the release
   waits for. Tonight's collisions were all avoidable with one rule:
   "release only when the tree has been quiet for N minutes and
   origin == HEAD".
3. **Deterministic fault injection as the default test design** — the
   modernc UDF trick (register a scalar function that fails on call N)
   is a reusable pattern for ANY iteration-error pin in this repo;
   worth a docs/lessons.md entry.
4. **Re-read before every edit** while the mtime-touching watcher is
   active (three wasted tool calls tonight).
5. **Counter-test design**: assert each counter on the exact resolver
   instance that performed the action (tonight's failure was an
   assertion aimed at the wrong instance).
6. **Fold-then-code ordering**: my CRM hardening extended the train
   after the fold — harmless because no tag exists yet, but the
   cleaner order is: finish all train work → fold once → release.

## f) UP TO 50 NEXT THINGS (ordered: finish tonight's work first)

1. Implement /api/calls idempotency server-side (`key` field,
   `callsIdem` store, extension-namespaced, seen-before-resolve,
   record-on-success-only).
2. Add `key` to the OpenAPI `/api/calls` schema.
3. Extend `TestAPICallLoggingContract`: replay-with-same-key journals
   once (both 204s); 502 keeps the key retryable (failure recorded
   only after success).
4. Island: add `key: crypto.randomUUID()` in `recordCrmCall`.
5. Island test (a): recordCrmCall posts the right body, toasts on
   failure, no-ops when `crmEnabled` is false.
6. Re-read panels.js + oxlint globals BEFORE the island edit (the
   concurrent session is in island files; `crypto` may need the
   island-lint globals block).
7. Render `webphone_crm_lookups_total{outcome=…}` in /metrics (only
   when enabled).
8. Metrics test: lines present when enabled, absent when disabled,
   the no-extension-strings leak guard still passes.
9. `nix develop -c go test -count=1 ./...` — full suite; attribute any
   concurrent-session breakage before acting on it.
10. Island node tests (`nix run nixpkgs#nodejs -- --test
    --test-force-exit internal/web/assets/island-tests/*.test.mjs`).
11. `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` inside nix develop (erraudit
    tier-1 must exit 0 on the touched packages).
12. Decide fold policy with the owner (question 1 below) — then either
    fold the concurrent UX train's landed items into 2.6.0 or hold
    them for 2.7.0.
13. Write the CHANGELOG 2.6.0 addendum for the CRM hardening (one
    coherent edit once 12 is answered).
14. Update AGENTS.md CRM-seam paragraph: the disabled-policy split
    brain is closed (crm shares the pbx `do()` shape) + single-flight
    + counters + idempotency contract one-liners.
15. Update FEATURES.md CRM row (single-flight, counters, idempotency).
16. Run `nix flake check` (module eval, island-lint, treefmt gates).
17. Cross-build aarch64 sanity EARLY if the stack gates will rebuild
    anyway (release.sh does it; avoid duplicate full builds).
18. Explicit narrative commits per doc group (daemon beats me
    otherwise); verify `git log -1` carries MY message at each phase
    boundary.
19. Push by hand if the daemon is still stalled; verify ls-remote.
20. Wait for tree quiet (concurrent session), then run
    `scripts/release.sh 2.6.0` to completion (gates → signed tag →
    push → lychee → stack relock → stack gates incl. browser E2E →
    aarch64 with the b700 ELF guard → gh release).
21. Post-release: `python3 scripts/webphone-smoke.py --expect-version
    2.6.0` (positive AND a negative probe against a wrong version).
22. pbx-artmann relock #4 per the ritual (rev via `git rev-parse`,
    `nix flake update telephony`, `nix run .#lock-drift-probe`, BOTH
    toplevels, ExecStart store path moved, narrative commit).
23. TODO_LIST harvest: close the dedup-train pins row and CRM
    follow-ups (a)–(d); annotate (e)–(h) status; close the
    release-train row after the tag lands.
24. Update docs/status/ with the release-outcome report.
25. Closing sweep per runbook: prove any booted process dead, final
    ls-remote verify, gitleaks/codespell via `scripts/buildflow.sh`.
26. Document the UDF fault-injection pattern in docs/lessons.md.
27. CRM follow-up (e): stack-side `crm.url`/`crm.token` secrets-dir
    wiring (stack repo).
28. CRM follow-up (f): cross-doc the CRM surfaces into the stack
    runbook (keep both error-contract sides in sync).
29. CRM follow-up (g): restore-drill proving `call_logged` entries
    survive a CRM journal restore.
30. CRM follow-up (h): settle the island→server→CRM integration level
    (contract test here vs stack E2E).
31. Send-failure UX (E): provider refusal → 422 + honest log family;
    test + error-contract table + stack runbook move together.
32. Fax-lane self-send guard (rides C).
33. Send-failure UX (C): OWNER decision — instant self-send refusal vs
    evidence-preserving failed row.
34. Send-failure UX (F): own-DID on the session payload + composer
    warning — only if self-sends recur after B(+C).
35. Theme-knob: FOUC headless first-paint screenshot pair.
36. Theme-knob: German-native review of the new copy.
37. Theme-knob: screenshot QA of the four composer affordances.
38. After the owner deploys: post-deploy probes — smoke
    `--base https://pbx.artmann.tech --expect-version <deployed>` +
    the rejection-banner check (needs a real extension session).
39. Outbound SMS bridge root cause: owner journals the telnyx-webhooks
    unit; record the cause here + in the stack runbook.
40. OWNER-calls batch session (~15 decisions; briefing doc exists at
    docs/planning/2026-09-22_13-50_owner-calls-briefing.md).
41. Post the release announcements (drafts exist; owner picks channels
    + disclosure posture).
42. OWNER ratify the oops non-fix (T15).
43. T26b TURN REST stack half (coturn static-auth-secret wiring).
44. Quarterly standing watches re-check (2026-12-20; sip.js 0.22,
    templ-components upstream, oxlint globals, E2E budget 445s).
45. Re-measure erraudit tier-2 monthly (next 2026-10-22) — tonight's
    crm work shrinks the family-adoption backlog slightly.
46. After the concurrent UX train lands: verify `templ generate` ran
    (committed `*_templ.go`) and its island tests pass.
47. Consider an idempotency-key note in the AGENTS gateway-seam
    paragraph (the calls endpoint now has the hooks' dedupe semantics).
48. Keep [Unreleased] accumulating after the tag (post-2.6.0 work
    starts with an empty section, per the fold shape).
49. Watch for the push daemon recovering; keep hand-pushing at phase
    boundaries until confirmed.
50. When v2.6.0 ships: update the stack browser-E2E budget watch if
    the forced rebuild lands outside 445s (two consecutive over-budget
    runs trigger action).

## g) QUESTIONS I CANNOT ANSWER MYSELF (max 3)

1. **Fold policy for the concurrent train**: the UX raw-ideas session
   is landing commits right now (hover timestamps, typeahead, …).
   Should v2.6.0 WAIT for that train to finish and fold its items into
   the 2.6.0 section (one honest release), or ship v2.6.0 with the
   three announced themes + my CRM hardening and let their train be
   2.7.0? I can wait out the tree, but the changelog decision is yours.
2. **Idempotency contract shape**: is a browser-generated UUID `key`
   acceptable as the sole dedupe contract for `POST /api/calls`
   (server trusts the client's key, absent key = no dedupe), or do you
   want a server-side fingerprint fallback (extension + number +
   direction + a coarse time bucket) so even keyless double-posts
   cannot double-journal?
3. **Push daemon**: it stalled earlier tonight (my fold sat unpushed
   until I pushed by hand — the `e57faf6` precedent). Is the daemon
   supposed to be alive right now (i.e., should I investigate/restart
   it), or is hand-pushing at phase boundaries the expected mode until
   you say otherwise?

— Waiting for instructions.
