# CRM-integration follow-up + gates train — status (2026-09-22 19:11 CEST)

> CLOSED 2026-09-23 (docs-health): c1 (vulnix NVD-404) resolved itself
> transient — every later gate green incl. the release run; c2 quiet-host
> reruns landed (E2E ×2 green + buildflow + flake check, 19:55); c4–c6
> and c12's counters shipped in the 21:12/02:47 trains (island test,
> idempotency, single-flight, `/metrics` family). Still open, routed:
> c7–c9 stack-side CRM wiring (TODO CRM row), c10 aarch64/stack bump +
> the v2.6.0 TAIL (TODO row), c11 + the c12 policy tails (owner-calls),
> c13 daemon pre-sweep (ROADMAP infra ask), c3 CRM-repo gates (CRM repo).

Session scope: execute the nine REMAINING items from
`docs/status/2026-09-22_16-54_crm-integration-train-status.md` (the
integration itself was already landed and green): re-verify both repos'
suites, run every outstanding gate, close the openapi/docs gaps, sweep
TODO_LISTs, and triage what the gates surfaced. Three concurrent sessions
worked the same machine throughout (T27/ThemeScript-CSP train in webphone,
a dedup/GContacts/CV-plan train in CRM, plus at least one more host
saturator — load peaked at 142).

---

## a) FULLY done

### Verification + gate work

- **CRM full suite GREEN in the pinned sandbox** (`scripts/verify-pinned.sh`,
  `-race -count=1`, all 9 packages, quiet machine) — including
  `internal/identity`, whose `TestRegistrationRateLimitContract` had failed
  twice earlier only under concurrent-suite load. The load-starve
  attribution from the prior session is confirmed.
- Prerequisite fix, verified TIDY-CLEAN: the CRM committed go.mod was
  untidy (a daemon sweep had downgraded `go 1.26.7` → `go 1.26`; stale
  `go-cqrs-lite/claiming` require; ginkgo/gomega direct/indirect marker
  drift). Tidied IN THE SANDBOX — the source-repo tidy fails on the drifted
  local `../go-cqrs-lite` (requires go ≥ 1.27.1; the pinned siblings are
  the truth) — and go.mod/go.sum copied back.
- **`/api/calls` added to the OpenAPI spec** (`internal/server/server.go`
  `openapiSpec`): POST `logCall`, request schema (number + direction
  required; direction enum in/out; seconds clamps ≥ 0; outcome enum
  answered/missed — all derived from the handler, the contract tests, and
  the island's `recordCrmCall` payload), responses 204 / 400 / 401 / 403 /
  422 / 429+Retry-After / 502.
- **Spec-vs-handler pin tests, green**: `TestOpenAPIEndpoint` extended
  (calls must document POST) + new `TestOpenAPICallLogMatchesHandler`
  (every documented code must exist; 429 must carry Retry-After) — same
  style as the contacts pin.
- **`crmLogFailed` row added** to the failure→feedback table (its real
  home, `docs/error-contract.md` — AGENTS.md had been restructured since
  the prior report): warn toast, en/de copy quoted truthfully from
  i18n.js, test home = crm_test.go 502 contract + i18n parity.
- **Webphone buildflow 51/52 green** (full mode, no result cache): the
  single failure is vulnix (external, see d/b). Fixed en route:
  - `TestMetricsServesAggregatesOnly` could NEVER have passed (T26a):
    the test harness wired `session.NewMemStore`, so the `sessions`
    TABLE did not exist and `ReadCounts`' sessions probe 500'd. Test
    server now uses `NewSQLiteStore(db, time.Hour)` (the prod wiring);
    3× green under `-race` + whole package green.
  - errcheck warning `messages_test.go:478` (another train's committed
    code): `defer func() { _ = hookResp.Body.Close() }()`.
- **CRM buildflow**: the vendorHash mismatch my tidy caused was fixed the
  sanctioned way (`buildflow -s nix-hash-fix --fix` → one-line flake.nix
  change, committed); `nix build -L .` EXIT=0; nix-build step green.
- **Unblocked the whole webphone build**: committed main had been red
  since commit `9d68645` — `type Counts` + `func Counts` collided in
  `internal/store/sweep.go` (T26a sweep). Applied EXACTLY the fix the
  dedup session's own status doc prescribed as their #1 TODO (rename the
  function → `ReadCounts`, keep the type; update the one caller in
  metrics.go; drop their now-unused `context` import; gofmt-align the
  struct — the last caught by treefmt-check).
- **Post-ThemeScript-train re-verification**: the concurrent CSP train
  (templ-components `NoThemeScript`, v1.19.2) landed mid-session and had
  transiently red the sandbox build + CSP test; afterwards: `-race` green
  for internal/server + crm + store, treefmt-check green, golangci-lint
  green.
- **Final git end-states verified with ls-remote**: webphone clean, HEAD
  `d91b96c` == origin/main (my openapi/test/doc work daemon-committed and
  pushed); CRM integration + my tidy/hash/TODO changes committed.

### Docs + sweeps

- **Webphone TODO_LIST**: added (1) CRM follow-ups row (island
  `recordCrmCall` test, idempotency key, single-flight resolver, lookup
  counters, stack-side secrets wiring, runbook cross-doc, restore drill);
  (2) OWNER policy row (typed NixOS crm options / multi-contact "+N more"
  / English-only journal bodies); (3) **vulnix NVD-404 blocker row
  (High)**.
- **CRM TODO_LIST**: added phones-column CSV import E2E row (unit-tested
  only today; replay_e2e row + fixture keeps the alias list honest).

## b) PARTIALLY done

- **Webphone `nix flake check`**: not green this session. First attempt
  hit the ThemeScript train's mid-edit state (`unknown field
  NoThemeScript`); the train has since landed and the suite is green, but
  the full check (sandbox package build + island checks + KVM VM test)
  has NOT been re-run post-landing — needs a quiet host.
- **Webphone full buildflow re-run post-metrics-fix**: deliberately
  deferred (host load 68–142 makes `-race` timing unreliable); verified
  with targeted steps instead (race server/crm/store, treefmt,
  golangci — all green).
- **Stack browser E2E**: 2 runs, BOTH failed at DIFFERENT
  load-sensitive markers (run 1: `after-register-WS` frame timeout with
  3 failed outbound calls; run 2: `CONTACTS-ROUNDTRIP-OK` 180 s timeout
  — its call/transfer flow itself worked). Host load was 68 rising to
  142 (other sessions). The chain was E2E-green twice earlier today.
  Attribution: load-induced, not a regression. Third run deferred.
- **CRM buildflow**: nix-build green after the hash fix; the remaining
  red steps (go-generate, golangci-lint, govalid-generate, license-check,
  test-coverage + go-* warnings) ALL stem from the documented BLOCKED
  sibling-drift item (`../go-cqrs-lite` requires go ≥ 1.27.1 vs pinned
  `GOTOOLCHAIN=local`). gitleaks/codespell not run (on-demand).
- **CRM push state**: my changes committed, but the daemon lagged ~2
  commits behind origin at last check, and the tree carries another
  session's 151-line uncommitted `cmd/crm-server/main.go` WIP (not
  mine — hands off).

## c) NOT started

1. ~~**vulnix NVD-404 fix** (newly discovered; blocks the next release) —
   TODO_LIST High row, no fix attempt (owner/triage call).~~ done
   (transient — `nix run .#vulnix` green with full triage since 19:55
   and inside the v2.6.0 release gates)
2. ~~Quiet-host reruns: stack E2E attempt 3, webphone full buildflow,
   webphone `nix flake check`.~~ done (E2E ×2 green 19:55; buildflow
   RC 0; flake check ALL PASS — 19:55)
3. CRM gitleaks + codespell runs. ← CRM repo
4. ~~Island unit test for `recordCrmCall` (coverage is still server
   contract + i18n parity only).~~ done (02:47, `panels-crm.test.mjs`)
5. ~~Idempotency key on `POST /api/calls` (a retry can double-journal).~~
   done (02:47, UUID key + `callsIdem`, 4 contract subtests)
6. ~~Single-flight in `crm.Resolver` (history burst fans out).~~ done
   (21:12, leader/waiter single-flight, race-tested)
7. Stack-side `crm.url`/`crm.token` wiring story (secrets dir). ← TODO CRM row
8. Cross-doc CRM surfaces into the stack runbook § error contract. ← TODO CRM row
9. Restore-drill proving `call_logged` survives a CRM journal restore. ← TODO CRM row
10. ~~aarch64 cross-build + stack input bump~~ (release territory; gated on
    the vulnix fix anyway). ← release-TAIL row
11. NixOS typed `crm.{url,token}` options decision (owner). ← owner-calls row
12. ~~Lookup hit/miss/timeout counters;~~ done (21:12 + 02:47 metrics
    family) multi-contact "+N more"; English-only journal confirm; CSV
    import E2E (CRM row added); plan v1.1 (CRM + webphone on different
    hosts) — ← owner-calls / CRM repo.
13. Pre-sweep build check in the auto-commit daemon (fleet-level ask;
    today alone saw THREE sessions hit committed-broken main). ← ROADMAP infra ask

## d) TOTALLY fucked up

1. **Used a hand-rolled rsync for the CRM suite first** — failed with
   "go mod tidy needed", sent me chasing go.mod state, when
   `scripts/verify-pinned.sh` (whose header DOCUMENTS the sibling-drift
   trap) was the canonical entry. One wasted cycle, entirely avoidable.
2. **Ran `go mod tidy` in the CRM SOURCE repo** — failed on drifted local
   siblings; should have gone straight to the sandbox (same lesson as 1).
3. **Invented a nil comparison in the pin test** (`post == nil` on a
   struct map value → compile error) instead of copying the comma-ok
   idiom used 30 lines above in the same file.
4. **Tried to edit sweep.go without reading it** — the tool refused;
   rule violation, no excuse.
5. **Nearly clobbered my own TODO_LIST rows**: anchored the vulnix-row
   insert on the CRM-follow-ups row I had just written (a file that
   other sessions + the daemon touch every minute). The
   modified-since-read refusal saved me; the anchor should have been a
   stable neighbor.
6. **Ran two ~8-minute VM E2Es before checking `uptime`** — load was 68
   (climbing to 142); both runs were doomed and burned ~15 minutes.
   Check load BEFORE scheduling VM tests, not after two failures.
7. **`nix flake check | tail; echo EXIT=$?` reported tail's exit code** —
   almost announced a false green for a failed build. Pipefail/PIPESTATUS
   from now on.
8. Minor: mis-modeled webphone push state twice in a row because HEAD
   moved between my own commands (concurrent session); resolved with
   `merge-base --is-ancestor`, but I should have expected HEAD
   instability from command one in this repo.

## e) What we should improve

1. **Load-gate the timing-sensitive tests**: run `uptime` first; defer
   VM tests and full race suites when 1-min load > ~10 on this shared
   host. Encode it in the runbook, not tribal memory.
2. **Never read exit codes after a pipe** (`PIPESTATUS`/`set -o
   pipefail`), or redirect to a log file and echo `$?` separately.
3. **Canonical scripts before improvisation**: each repo's
   verify/buildflow script header is documentation; read it first (it
   saved the CRM twice today once actually read).
4. **Shared-file edit discipline**: re-read immediately before every
   edit (daemon commits + multiple live sessions make every read stale
   within minutes); anchor inserts on stable text, never on your own
   fresh edits.
5. **Attribution-first worked and should stay**: concurrent-session
   breakage (Counts collision, CSP mid-flight, ThemeScript) was
   attributed before acting, and where the other session had DOCUMENTED
   the intended fix (ReadCounts), completing their prescription beat
   inventing a new one.
6. **The auto-commit daemon needs a pre-sweep build gate** — today it
   committed broken main repeatedly across sessions (Counts; untidy
   go.mod downgrades; mid-flight CSP state). Suggested fix lives in the
   dedup session's doc too: skip the sweep while `go build ./...` fails.
7. **Test-harness realism**: T26a's metrics test shipped green-confident
   but could never pass (MemStore vs SQLite schema). Any new test that
   depends on infrastructure wiring (schema, stores, clock) must run at
   least once against the prod-shaped wiring before being trusted.
8. **Exit-code honesty in gates**: buildflow's treefmt/vulnix failure
   modes (formatter diff, NVD 404) are precise, but my own shell habits
   weren't; fix the human half too.
9. **CRM go-directive story needs one decision**: `go 1.26` (committed)
   vs `1.26.7` (tidy's demand) vs the pinned toolchain — the
   go-version warnings in CRM buildflow all trace to this unresolved
   triple.

## f) Next things (ordered, 50 — brainstorm, not a commitment list; the

first ~15 are already harvested into TODO_LIST.md)

1. Fix the vulnix NVD-404 release-gate blocker (direction = question 2
   below).
2. Stack browser E2E attempt 3 on a quiet host (load < ~10).
3. Webphone full buildflow on a quiet host (post metrics-fix).
4. Webphone `nix flake check` on a quiet host (post ThemeScript
   landing).
5. Decide the CRM sibling-push question (BLOCKED item) — unlocks 6 red
   buildflow steps, flake pin bumps, and the nightly stress proof.
6. Verify CRM daemon push lag (2 commits) + coordinate the other
   session's uncommitted `cmd/crm-server/main.go` WIP.
7. Island unit test for `recordCrmCall` (helpers.mjs fetch stubs).
8. Idempotency key on `POST /api/calls`.
9. Single-flight lookups in `crm.Resolver`.
10. Lookup hit/miss/timeout counters (replace debug-only slog).
11. Stack-side wiring story for `crm.url`/`crm.token` (secrets dir).
12. Cross-doc CRM surfaces into the stack runbook § error contract.
13. Restore-drill: prove `call_logged` entries survive a CRM journal
    restore.
14. CRM phones-column CSV import E2E (TODO row exists).
15. erraudit tier-2 re-measure incl. `internal/crm` (due 2026-10-22).
16. Island local recent-calls enrichment via a lookup endpoint (owner
    scope call).
17. Multi-contact numbers: surface "+N more"? (owner)
18. Confirm English-only CRM journal bodies (owner).
19. Typed NixOS `crm.{url,token}` options vs documented freeform
    (owner).
20. Plan v1.1: CRM + webphone on different hosts (TLS, token rotation).
21. Pre-sweep `go build` gate in the auto-commit daemon (fleet ask).
22. Load-gate policy for VM tests in the stack runbook + E2E harness.
23. Backoff/Retry-After handling when the CRM rate limits upstream.
24. OWNER: deploy the v2.5.0 chain to prod (standing TODO row).
25. OWNER: post-deploy verification (smoke + 40310 banner check).
26. Outbound SMS bridge prod root-cause (stack-side, OWNER terminal).
27. Post the release announcements (drafts exist through v2.5.0).
28. Send-failure UX follow-ups C/D/E/F.
29. T26b stack half: coturn REST secret.
30. If E2E attempt 3 fails clean: analyze with the shipped
    `transfer_dbg()` dumps.
31. CRM: manual browser pass (passkey registration + deal-owner
    attribution).
32. CRM: manual print/dark-mode eyes pass.
33. CRM: credential-management UI decision (own templ page vs
    API-only).
34. CRM: Google Contacts sync v1 (plan exists; shared dependency with
    email-less identity D1).
35. CRM: Twenty migration owner decisions D1/D2.
36. CRM: extraction script `scripts/extract-twenty/`.
37. CRM: first real migration rehearsal + counts reconciliation.
38. CRM: cutover execution per the migration plan (tasks 10-20).
39. CRM: verify real Twenty CSV headers against the alias parser.
40. CRM: migrate `internal/identity` off deprecated `stack/v4` APIs
    before the go-cqrs-lite v5 cut.
41. CRM: root-cause or nightly-prove the `TestSQLiteRestartReplaysJournal`
    tags flake.
42. CRM: README screenshots (headless browser).
43. Both repos: gitleaks + codespell on-demand runs.
44. Webphone: triage the 35 govulncheck warnings (mostly stdlib noise —
    confirm or fix).
45. Webphone: triage mypy/ruff/shellcheck/lychee/jscpd warnings
    (nix-run-env noise vs real findings; silence properly or fix).
46. Webphone: 2 nix-checker findings (documented oscillation?) — triage.
47. CRM: 38 go-auto-upgrade findings — adapt or document non-fix.
48. CRM: settle the go-directive triple (1.26 / 1.26.7 / pinned
    toolchain) — kills the go-version + gomod-check warnings.
49. CRM: 1 oxlint finding — fix.
50. Session claim markers / lockfile to reduce multi-session collision
    overhead (pairs with 6/21).

## g) Questions I can NOT figure out myself

1. **CRM sibling push (the standing BLOCKED item, now load-bearing):**
   may go-cqrs-lite (119 commits) and cqrs-htmx (50) be pushed? This
   single decision unlocks CRM buildflow's 6 red steps, the flake pin
   bumps, and the nightly-stress proof. The recorded alternative: adopt
   a flake-managed go 1.27 posture and retire the pinned sandbox for
   dev runs. Which way?
2. **vulnix NVD-404 fix direction** (blocks the next release): bump/
   patch vulnix in nixpkgs, repoint the wrapper at a working mirror, or
   swap the gate to govulncheck (already runs in buildflow) and demote
   vulnix to advisory? My recommendation is the govulncheck swap, but
   the release-gate risk tolerance is yours.
3. **Concurrent-session red-commit policy** (asked by the dedup session,
   recurred three times today): when committed-but-broken main is
   detected from another session, fix-and-commit immediately (unblock
   everyone, small collision risk) or keep hands-off-and-report? And do
   you want the daemon pre-sweep build gate + per-session claim markers
   pursued as fleet tooling?
