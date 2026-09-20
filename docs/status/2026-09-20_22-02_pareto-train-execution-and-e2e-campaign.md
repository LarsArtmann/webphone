# Status Report — Testing & Error-Feedback Pareto Train Execution

- **Created:** 2026-09-20 22:02 CEST (`date`)
- **Session scope:** execution of the full 27-task plan
  `docs/planning/2026-09-20_17-23_SUPERB-testing-error-feedback-pareto-plan.md`
  ("GET SHIT DONE" instruction), plus honest findings along the way.
- **Format note:** user explicitly requested `.md`; the status-report
  skill's canonical format is HTML — one-off override, not propagated.
- **End state:** webphone `260fbee` pushed (origin/main == HEAD), stack
  `de0feec` pushed, stack `nix flake check` ALL GREEN same day.

---

## a) FULLY DONE (verifiable, committed, gated)

| # | Item | Evidence | Scope |
|---|---|---|---|
| A1 | **SQLite-backed session store** (T11 spike + T12): `Store` interface seam, `NewMemStore` (tests) / `NewSQLiteStore` (prod) over the same `webphone.db`, TTL parity (millis, sweep on Create, delete-on-expired-read), constructor guards | `f0cd868`/`6ade868`; `TestSQLiteSessionStoreSurvivesRestart` + 5 more store tests green; verdict doc `docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md` | `internal/session/`, `cmd/webphone/main.go`, `internal/server/server.go` |
| A2 | **kill -9 restart smoke scenario**: login → SIGKILL → reboot same data dir → old cookie still authorized; anonymous still 401 | `a76df8f`; smoke "restart scenario: 4 passed" (re-verified at 22:00) | `scripts/webphone-smoke.py` |
| A3 | **Toast accessibility** (T01): keyboard dismissal (Enter/Space/Escape) in island `announce()` + shell `shellToast`, focusable but never focus-stealing; served live-region attributes (`role="status" aria-live="polite"`) pinned Go-side | `2bf56cd`; 24/24 island tests; `TestServedPageHoldsTheDomContract` extended | `ui.js`, `shell.js`, `phone.templ` (attr already present), `server_test.go` |
| A4 | **Durable inline tab errors** (T02): `#wp-tab-error` slot rendered outside `#tab-content`; htmx `responseHandling` override (401 swap-free exception; 4xx/5xx select `.wp-error` → slot); 502 test asserts HX-Trigger + banner presence; config contract pinned on served page | `2178896`; `TestSendClassifiesGatewayOutageAs502` extended; server suite green | `layout.templ` (+regenerated), `app.css`, `messages_test.go`, `server_test.go` |
| A5 | **Toast feedback map completed** (T03): status-specific login-failure toasts (401/429/other/net) in en+de; shell 429-specific wording; SSE-dead toast after 3 consecutive failures with recovery reset | `74412d0`; new `session.test.mjs` 6 specs; i18n parity tests green | `session.js`, `i18n.js` (both maps), `shell.js` |
| A6 | **SSE pill a11y + dedup** (T14): pill now `role="img"` with localized state label (was `aria-hidden`); identical-consecutive toast dedup in both hosts | daemon-committed (same session); 24/24 island specs incl. pill state-flip test | `session.js`, `ui.js`, `shell.js`, `i18n.js` |
| A7 | **Store owner-scoping suite** (T15): every owner-taking query pinned against a foreign extension — reads → `ErrNotFound`, mutations inert; webhook lookups documented as deliberately ref-scoped | `51036e4` (integrated with the concurrent session's byte-identical store file); store 59.4% → **75.3%** | `internal/store/owner_scoping_test.go` |
| A8 | **Gateway error-branch table** (T17): transport, timeout, 5xx + truncated detail, JSON-ish receipt, oversized token, empty-ref acceptance contract, fax-lane 503 | `51036e4` (compile-fixed superset of the concurrent session's file); gateway 76.5% → **77.6%** | `internal/gateway/webhook_errors_test.go` |
| A9 | **Domain parser edge table + JSON fuzz** (T18): dialable prefix/length/charset/empty incl. RTL-paste; branded-id junk table; both round-trip forms; fuzz over contacts API (400/422/204, never 5xx) — 148k execs clean | `51036e4`; domain 54.1% → **72.1%** | `internal/domain/ids_parse_test.go`, `internal/server/contacts_api_fuzz_test.go` |
| A10 | **Island test infra** (T19): shell load-error boundary (#log breadcrumb), injectable clock (`window.__wpClock`), harness policy README with the recorded main.js skip decision | `9ea3cfd` | `shell.js`, `shell.test.mjs`, `island-tests/README.md`, `helpers.mjs` |
| A11 | **Toolchain self-heal** (T20): bare smoke + `scripts/buildflow.sh` detect the go-floor trap and re-exec via `nix develop -c`; verified bare (`ambient go 1.26.7 < floor 1.27.1; re-executing…` → full suite green) | `d1ddf62`; verified live twice | `scripts/webphone-smoke.py`, `scripts/buildflow.sh` |
| A12 | **Codespell zero + gate promotion** (T21): `.codespellrc` (German UI copy + technical terms); wrapper appends gitleaks+codespell to the default run — buildflow 52/52 green | `d1ddf62`; `BUILDFLOW_NO_RESULT_CACHE=1` verified | `.codespellrc`, `scripts/buildflow.sh` |
| A13 | **Backup/restore drill as flake check** (T23): `checks.webphone-backup-drill` runs the restore path sandbox-safe (pkill probed); byte-identical attachment retrieval asserted | `d1ddf62`; check green ("[4] restore drill PASSED") | `flake.nix`, `scripts/webphone-backup-drill.py` |
| A14 | **Train-hygiene script** (T24): fanout-if-bumped / vulnix / lychee / origin-vs-HEAD as one command; caught a real push drift the same day | `5d55ef0` (daemon); `--no-net` run green after sync | `scripts/release-hygiene.sh` |
| A15 | **Bandit triage** (T22): `.bandit` excludes the drill script wholesale with rationale (loopback-only, dummy creds, asserts ARE the verification) | `6d01355`/`110fb7c`; buildflow bandit quiet on the drill | `.bandit` |
| A16 | **AGENTS.md trim** (T22): 705 → 600 lines, history condensed, every invariant/gotcha preserved (spot-checked 10 key markers) | `110fb7c` | `AGENTS.md` |
| A17 | **Docs bundle** (T13/T25): failure→feedback map table (11 failure classes → surface/owner/test-home), session-seam invariant rewrite, hooksIdem 1h-TTL rationale, htmx-bump checklist (4 wire contracts), D3 English-shell decision, BDD posture | `0d6bc18`, `aba9855` | `AGENTS.md` |
| A18 | **CHANGELOG + FEATURES + HARVEST** (T04/T06): full train recorded; FEATURES session/feedback rows synced; ROADMAP + TODO_LIST harvested; coverage baseline doc with deltas + views decision (T26) | `6b7e7ba`, `3eb45dd`, `7df801a`, `50c0534` | `CHANGELOG.md`, `FEATURES.md`, `ROADMAP.md`, `TODO_LIST.md`, `docs/reviews/2026-09-20_coverage-baseline.md` |
| A19 | **aarch64 gate** (T05): cross-build + ELF machine-byte check `b7 00` = EM_AARCH64 | verified during train | build artifact |
| A20 | **Fax upload indicator + reduced-motion parity** (T26): `.htmx-indicator` ellipsis on the fax submit; island reduced-motion upgraded to the global catch-all | `50c0534` | `fax.templ`, `island/style.css` |
| A21 | **Stack re-pin + browser E2E** (T07/T08): stack lock bumped to `260fbee`; E2E run 1 green (149 s) incl. the existing wrong-password leg (T08); run 15 GREEN with all new scenarios; stack `nix flake check` ALL GREEN | stack `99c669b` + `de0feec`, pushed | `~/projects/nix-international-telephony` |
| A22 | **New stack E2E scenarios** (T07/T09/T10): webphone-restart mid-session (`RESTART-SESSION-KEPT`), unallocated-transfer verdict drill (`TRANSFER-VERDICT-SURFACED`), FreeSWITCH-stop mid-call + recovery (`FS-OUTAGE-DETECTED`/`FS-RECOVERED`); reload-retry dial helper | stack `de0feec` (216 insertions); run 15 green | stack `tests/browser-e2e.py`, `tests/browser.nix` |
| A23 | **Concurrent-session integration**: the same two test files independently pushed by another session mid-train were integrated without loss (store file byte-identical; gateway file had a compile error my version fixes) | `git log` shows clean ancestry; FF push `c66598f..5d55ef0` | two test files |

## b) PARTIALLY DONE

1. **AGENTS.md size target** — trimmed 705 → 600 lines (−15%), but the
   plan's own DoD said "≤ ~500" and buildflow's preflight suggests ≤377.
   What remains: history-heavy paragraphs in Architecture invariants and
   Hard-won knowledge that still carry dates and narrative. Effort to
   finish: **M** (2 more surgical passes). Not blocking anything.
2. **Fuzzing program** — `FuzzContactsAPISave` runs its seed corpus in
   every `go test ./...` (continuous), but no scheduled long `-fuzztime`
   runs exist, and there is exactly one fuzz target (send bodies, vcard
   parser are obvious next targets). Effort: **S** per additional target.
3. **Stack E2E scenario stabilization** — all green in run 15, but the
   12-run validation campaign surfaced three flake classes now only
   partially mitigated: first-dial INCOMING misses (3/13 runs, mitigated
   by re-run but NOT root-caused), REFER-verdict lag up to minutes
   (budgeted, not fixed), transferee stale-dialog lag (actively worked
   around by hangup). The runbook's "×2 green runs" rule for new
   scenarios is NOT yet satisfied. Effort: **M–L**.
4. **Coverage deltas recorded but follow-ups unstarted** — blob, fax,
   messaging still have no direct tests (exercised only via the server
   suite); recorded honestly in the baseline doc as accepted-for-now.
5. **CHANGELOG `Unreleased` pile** — the train recorded everything, but
   the pile now includes two full trains' worth of features (identities,
   styled 404, sessions, error-feedback). No release decision taken; the
   TODO_LIST train-cut question (owner call) is still open.

## c) NOT STARTED

1. **Local Playwright island harness + golden toasts (T27)** —
   consciously ROADMAP'd with trigger conditions (chromium closure cost
   vs. stack E2E coverage); not "forgotten", a recorded decision.
2. **Union test coverage (`-coverpkg=./...`)** — pre-existing ROADMAP
   item, blocked on BuildFlow feature; untouched this session.
3. **Prod redeploy with the new build** — the owner deploy (TODO_LIST
   High item) is unchanged; everything shipped here rides main but prod
   still runs the pre-v2.4.0 binary per the last known state.
4. **Release announcement posts** — drafts still waiting on owner
   channel/disclosure decisions (pre-existing).
5. **docs-health ANNOTATE over the 2026-09-19 status reports** —
   pre-existing TODO_LIST item; this session annotated only its own plan
   file.

## d) TOTALLY FUCKED UP

1. **The 15-run E2E validation campaign burned ~4 hours for a bug in MY
   OWN test code.** Severity: blocked the train's final gate for the
   session. Root cause chain, in order of discovery: (1) I added
   scenarios that dialed fresh calls after an echo teardown — the exact
   wedged-transport class this suite keeps hitting; (2) I reordered the
   testScript markers but NOT the python phase order (run 5 failed on a
   mismatch I introduced); (3) the decisive one: Selenium's `.text`
   returns EMPTY for content inside a closed `<details>` drawer — my
   `#log` verdict assertion read nothing, and I spent runs 9–14 chasing
   VM-load theories, budget raises, and zombie-card diagnostics before
   re-reading the API semantics. The correct tool was
   `textContent` via `execute_script` from the start. Mitigation:
   fixed + validated green (run 15), and the lesson is now written into
   the scenario comments.
2. **My `GOTOOLCHAIN: auto` theory was wrong.** I added the env pin to
   `.buildflow.yml`, then buildflow's own preflight told me the caller's
   shell WINS over project-side env — the key was a silent no-op I had
   just documented as the fix. Reverted within the same hour and
   replaced with the re-exec wrapper, but I shipped a wrong fix first
   because I did not read buildflow's preflight output before writing it.
3. **The zombie-card false alarm.** I burned two runs (13–14) and added
   diagnostic machinery chasing "the island leaves a zombie call card
   after REFER" — the card evidence in the console dump was the
   `.text`-reads-nothing bug's mirror image (I could not see the log
   entries my assertions needed, so I invented a DOM race). There is no
   confirmed zombie; the diagnostic hook remains in the scenario.
4. **Pre-existing, still fucked (observed, not caused, NOT fixed):**
   (a) the first-dial INCOMING flake (~23% of runs) is still
   unroot-caused — the TODO_LIST "1001-registration anomaly" item
   stands, now with more data points; (b) the REFER-verdict can lag
   minutes under VM load — budgeted around, mechanism unknown;
   (c) prod still runs a build without the session-verification fix
   (owner deploy pending since 2026-09-19).

## e) WHAT WE SHOULD IMPROVE

1. **Read the tool's self-diagnostics BEFORE designing around it.**
   Buildflow's preflight literally printed the constraint that
   invalidated my first fix. New rule: when a tool has a `doctor`/
   preflight output, read it before writing config for it.
2. **Assert on rendered-independent DOM properties in Selenium.**
   `.textContent` via `execute_script` for any content inside
   `<details>`/hidden regions; `.text` only for visible text. This cost
   ~4 hours; it should become a standing note in the stack E2E's
   helpers (comment added, but the rule belongs in common.nix docs).
3. **Change python AND nix sides of the marker contract atomically.**
   Runs 3–5 wasted on python/testScript phase-order mismatches. A
   single comment block listing the marker ORDER (kept in both files,
   or one canonical list in the nix file) would make drift obvious.
4. **Retry-friendly phase design in the E2E.** The fresh-dial wedge bit
   three times. `dial_into_call`'s reload-retry should become the ONLY
   way the suite dials (refactor the original dial phase onto it too).
5. **VM-load hygiene during long E2E runs.** Run 4 failed while three
   heavy builds competed for CPU. The flake verdict rule says re-run;
   better: a pre-flight check that nothing else heavy is running, or
   simply a documented "quiet machine" note in the runbook.
6. **Concurrency etiquette**: the concurrent session and I wrote the
   same two test files simultaneously — the integration was lucky
   (byte-identical store file). A pre-work `git fetch` + check of
   uncommitted/recent remote commits in BOTH repos (webphone + stack)
   would have saved the collision entirely.
7. **Estimate honesty**: the plan estimated T07 at 100 min; the real
   cost (15 E2E runs × ~3–6 min + diagnosis) was ~4 h for T07–T10
   combined. Future plans should budget E2E-touching tasks at 3–4× the
   Go/JS task rate.
8. **The codespell ignore list is a workaround, not a fix.** German
   words in en/de dictionaries will keep tripping spell gates; the
   honest fix is codespell's dictionary-aware mode or scoping the scan
   to code + English docs only.

## f) TOP 50 THINGS TO GET DONE NEXT (ranked by impact; HARVEST: Critical/High → TODO_LIST, Medium → TODO_LIST, Low/long-shot → ROADMAP)

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 1 | Re-deploy prod with current main (sessions persist + verification fix + error-feedback); verify with `--base` smoke incl. "bogus credentials rejected" | Critical | S | Bug (security) |
| 2 | Root-cause the first-dial INCOMING flake (3/13 runs): instrument the callee's WS/registration state at dial time; fix island or E2E per findings | High | M | Bug |
| 3 | Two consecutive green E2E runs for the new scenario set (runbook rule for new scenarios) | High | M | Quality |
| 4 | Investigate the REFER-verdict latency mechanism (why minutes under load): sofia task queue? log the REFER→NOTIFY interval | High | M | Bug |
| 5 | Owner decision: cut v2.5.0 (sessions + error-feedback + identities pile) vs hold | High | S | Release |
| 6 | Refactor the E2E's ORIGINAL dial phase onto `dial_into_call` (retry everywhere, one dial path) | High | S | Quality |
| 7 | Document the Selenium `.text`-vs-`textContent` rule in the stack's `common.nix`/E2E header (it cost 4 h) | High | S | Documentation |
| 8 | Canonical marker-order list for the browser E2E (one list, referenced by python + nix) to prevent phase drift | High | S | Quality |
| 9 | Second AGENTS.md trim pass toward ≤500 (Architecture invariants + Hard-won knowledge history) | Medium | M | Cleanup |
| 10 | Add fuzz targets: outbound message bodies, vcard parser, webhook multipart forms | Medium | S | Quality |
| 11 | Scheduled long fuzz runs (`-fuzztime 5m`) as an optional flake app | Medium | S | Quality |
| 12 | Direct test suites for `internal/blob`, `internal/fax`, `internal/messaging` (currently server-tested only) | Medium | M | Quality |
| 13 | Fix the E2E flake budget: raise the E2E wall-time watch trigger (151 s → ~180 s) now that scenarios added ~90 s; update the standing watch | Medium | S | Documentation |
| 14 | Quiet-machine pre-flight for E2E runs (warn when nix builds run concurrently) | Medium | S | Quality |
| 15 | Encrypt-at-rest config knob for the sessions table (key from environmentFile), for deployments that demand it | Medium | M | Feature |
| 16 | Add a `sessions` count to `/healthz` detail (ops visibility into the new store) — careful: GET-open body leaks only check names today | Medium | S | Feature |
| 17 | Backfill `docs/announcements/` drafts for v2.5.0 when cut | Medium | S | Documentation |
| 18 | Stack-side csrf assertion in its webphone VM test (typed options render) — pre-existing TODO | Medium | S | Quality |
| 19 | Owner answer: reconcile the stack's previously uncommitted tree (may be resolved by now — re-check) | Medium | S | Cleanup |
| 20 | docs-health ANNOTATE over the 2026-09-19/20 status reports (this one included, when superseded) | Medium | S | Documentation |
| 21 | Island unit test for the transfer-failure `referOnNotify` branch (the E2E cannot reach it — dialplan catch_all) | Medium | S | Quality |
| 22 | Consider `aria-live="assertive"` for error toasts specifically (polite may under-announce urgent failures; needs a screen-reader check) | Medium | S | Feature (a11y) |
| 23 | Fax upload: assert the indicator behavior in a browser context (currently only CSS/attr shipped) | Medium | S | Quality |
| 24 | Prune buildflow's preflight warns: cache DB VACUUM, legacy DB removal, binary freshness (rebuild buildflow) | Low | S | Cleanup |
| 25 | codespell: scope the scan (code + English docs) instead of ignore-list growth | Low | S | Cleanup |
| 26 | Module option for `session_ttl` docs in README module table (TTL now load-bearing for at-rest exposure) | Low | S | Documentation |
| 27 | Add the kill -9 restart scenario as a flake-level check too (currently smoke-only) | Low | M | Quality |
| 28 | Vulnix + lychee full (online) hygiene run — this session ran `--no-net` | Low | S | Quality |
| 29 | Re-run `BenchmarkHubFanOut` is NOT needed (no SSE lib bump) — but record that fact in the baseline doc to close the MD1 question for this train | Low | S | Documentation |
| 30 | Expose the restart scenario's "anonymous still rejected" check in `--base` foreign mode (works today? verify) | Low | S | Quality |
| 31 | Wire `scripts/release.sh` to call `scripts/buildflow.sh` (one gate entry point) | Low | S | Cleanup |
| 32 | Consider per-tab CSRF re-arming cost audit: every htmx request re-reads hx-headers — verify no perf cliff on morph-heavy pages | Low | S | Quality |
| 33 | ICE panel: reduced-motion already global; check the `.incoming` keyframes are the only remaining animation and document it | Low | S | Documentation |
| 34 | Session store: add a started-at/metrics log line on boot (row count) for ops | Low | S | Feature |
| 35 | Unit-test `session.NewSQLiteStore` sweep under concurrent Create (map the mutex parity story) | Low | S | Quality |
| 36 | Move the `window.__wpClock` seam doc from shell.js comments into island-tests/README (it is a public test contract) | Low | S | Documentation |
| 37 | Fuzz the smoke script's own parsing? No — instead pin the smoke suite count (36 checks) so silent check-loss fails | Low | S | Quality |
| 38 | E2E: assert `TRANSFER-UNALLOCATED-INITIATED` also on the callee side (banner/decline path) for full leg coverage | Low | S | Quality |
| 39 | Consider `dial_into_call` adoption by the conference/IVR tests in the stack (same wedge class) | Low | M | Quality |
| 40 | Update TODO_LIST "standing watches" E2E wall-time entry with the 15-run campaign data | Low | S | Documentation |
| 41 | Delete the zombie-card diagnostic hook from the E2E once the card-removal race is understood (it is currently dead weight) | Low | S | Cleanup |
| 42 | Session seam: evaluate whether `MemStore` should log a loud warning when used outside tests (loopback dev is silent today) | Low | S | Quality |
| 43 | i18n: the shell's htmx error copy (401/429/generic) is English-only by D3 — revisit if any de-first user feedback arrives | Low | S | Feature |
| 44 | Add `prefers-contrast`/forced-colors pass over toasts + pills (a11y next frontier after reduced-motion) | Low | S | Feature (a11y) |
| 45 | Golden screenshots of the four toast kinds (T27 fragment, cheap via the E2E's existing chromium, no Playwright needed) | Low | M | Quality |
| 46 | Sweep `.buildflow.yml` skip_steps warn ("match no registered tool") — names drifted upstream again | Low | S | Cleanup |
| 47 | Document in AGENTS that `checks.webphone-backup-drill` needs the sandbox to allow loopback (it does by default — record the verification) | Low | S | Documentation |
| 48 | Consider storing session rows' `extension` as the branded-string with prefix for debuggability (currently raw value — fine, but write the decision down) | Low | S | Documentation |
| 49 | Re-check `git ls-remote` on BOTH repos after the daemon's next burst (post-report commits) | Low | S | Cleanup |
| 50 | Celebrate: the #1 error class (dead sessions on every deploy) is deleted at the root — tell users in the release notes | Low | S | Documentation |

## g) QUESTIONS I CANNOT ANSWER MYSELF (tried and failed)

1. **Should the new stack E2E scenarios (restart / unallocated-transfer
   / FS-outage) become a REQUIRED gate for every webphone train, or stay
   opt-in?** I ran the full browser E2E manually this session (15 runs,
   ~4 h wall-clock, ~1-2 GB closure). If every future webphone push
   triggers it via the daemon, each train gains ~10-20 min wall-time and
   real flake exposure; if opt-in, the scenarios can rot. This is a
   cadence/risk tradeoff only the owner can set (the g2 cadence rule
   governs similar calls).
2. **Is the at-rest credential tradeoff (directory passwords in the
   sessions table, 24 h TTL, 0700 dataDir, plain-text column) acceptable
   for pbx.artmann.tech PRODUCTION, or do you want the
   environmentFile-keyed encryption knob (f-list #15) BEFORE the prod
   redeploy?** I judged it acceptable and wrote the threat model
   (`docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md`
   §4), but this is precisely the call a threat-model author cannot make
   alone — you know who has access to `/var/lib` on that box and whether
   backups leave it.
3. **The first-dial INCOMING flake (~23% of my runs) — have YOU seen its
   signature before (sofia shows the registration, the INVITE never
   renders the banner), e.g. during the 2026-09-19 "1001-registration
   anomaly"?** I know the island reuses a possibly-wedged Registerer
   after a transport reconnect, and my runs all passed on re-run — but
   if you have a sofia dump from a previous occurrence, it would decide
   island-fix vs. E2E-fix without another instrumented campaign.

---

*Point-in-time snapshot. Feed section (f) into `docs-health` HARVEST
(Critical/High → TODO_LIST; Low/long-shot → ROADMAP). When this report
goes stale, ANNOTATE it — do not rewrite.*
