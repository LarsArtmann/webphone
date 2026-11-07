# Status — M20 close-out, the json/v2 main-heal, and what the evening exposed (2026-11-07 19:56)

**Snapshot-rot header: claims verified as of 2026-11-07 19:56 CET.** Two
CI runs on my pushes are green (42400725432 heal, 42401403869 snapshot);
the repo's newest red (42402222719) belongs to the CONCURRENT session's
docs-health battery train — diagnosed (§c.1), not mine to fix, their fix
stream was landing as this report was written (latest head `6007c093`
19:54, run 42403427494 in flight). The daemon owns most commits; two of
mine pushed manually under the red-main bar.

Scope: everything in this report is THIS session (resumed ~18:35 CET).
No unrelated research was done; the concurrent session's work is
reported only where it collided with mine or red-ed main.

## a) FULLY DONE

| #  | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Verification                                                                                                                                                                                                                     |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **Stale todo list repaired** per the session-summary FIRST ACTION: T24a–d marked completed, M20 / cadence / §g routed in as tracked items                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | todos tool; every item now completed or explicitly owner-routed                                                                                                                                                                  |
| 2  | **M20 SHIPPED** (the last open frontend-ux train row): voicemail reload failure got its own copy — `vmLoadFailed` for transport/parse errors, the 401 arm KEEPS `vmAuthFailed` (the review's "no distinct copy" defect: a network error used to answer with the auth line, a lie about the cause); the call-history load (`refreshServerHistory`) and the contacts load (`loadContacts`) announce warn toasts on BOTH the `!res.ok` arm and the thrown-fetch arm via new `historyLoadFailed`/`contactsLoadFailed` keys (en+de) — their silent `return` was the other half of the finding                                                                                    | commit `5c368085` (daemon-swept, message lost — see §d.4); `panels-load-failures.test.mjs` 6/6; island suite 318/318; `tsc -p .` green; no served-bytes pin collisions (grepped smoke + Go tests); CHANGELOG `[Unreleased]` line |
| 3  | **TODO row closed**: the M20 residue row struck `[x]` with commit + pin evidence (mtime guard tripped twice first — §d.5)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | TODO_LIST frontend-ux residue block; commit swept as `66ae9184`                                                                                                                                                                  |
| 4  | **§g Q1 (cadence strategy) dissolved by events**: the v9 docs-health sweep (18:32, before this session started) drained the corpus with gated archives; I re-verified state myself — `docs-gates cadence .` exit 0, 70/70 at the pin, oldest live report 32d < 45d threshold, purge healthy. The 11:07 session's "archive-gate verbs missing from the installed build" premise didn't hold for whatever build v9 used; the installed binary genuinely lacks `check-rows`/`bare-rows` as verbs, so the canonical-binary question stays routed (v9 §g.8) and does not block                                                                                                   | command output in-session                                                                                                                                                                                                        |
| 5  | **§g Q2 substance written**: ledger rows for the four settled red runs — 42343813135, 42352673965, 42358288828 (+sister 42358387321), all the same two deterministic lanes (stale login.css after app.css edits + unbaselined clone groups), healed by `8a090ac5`, green successor 42362014234; plus the 42398615048 row recording the evening's json/v2 break and its heal. Factual finding: the watcher filed NO tracker issues for those runs (gh issue list verified) — the §d accounting in the 11:07 report was the only record until now                                                                                                                             | docs/ci-flake-ledger.md; commits `89c89030` + `d0ab66bb`                                                                                                                                                                         |
| 6  | **Main healed — the json/v2 marshal break**: the concurrent session's daemon sweep `43f70b9c` (18:49) moved sso_test.go onto `encoding/json/v2` (import swap + `MarshalWrite`), whose strict Marshal ERRORS on the unexported-only `domain.Extension` ("Go struct has no exported fields") — a DETERMINISTIC go-tests red on every run since (verified: run 42398615048, `--- FAIL: TestLoginClaimsAuditLineCarriesGroupsOnlyInTheJournal`; reproduced locally 3×). Fixed WITHOUT touching their in-flight file: `MarshalJSON` on domain.Extension rendering the raw value (the id.ID doctrine; `authmap.User` is never JSON-serialized in prod, so prod wire is untouched) | commit `2c50b2c7`; sso `-race` green 3× then 5×; full `go test -race ./...` exit 0; CI 42400725432 GREEN all jobs                                                                                                                |
| 7  | **Main healed — issue #236's real data race** (same commit): the audit-line journal test swapped the process-global `slog.SetDefault` handler to its own `bytes.Buffer` while `t.Parallel()` let sibling sso tests log through it — exactly the race run 42308961304 detected (concurrent `bytes.Buffer.Write` vs the test's `buf.String()` read at sso_test.go:605). The test now runs sequentially, mechanism documented at the site; the race window is eliminated by construction (sso spawns no goroutines)                                                                                                                                                            | commit `2c50b2c7`; 5× `-race` green locally; successor run green                                                                                                                                                                 |
| 8  | **Issue #236 closed on evidence** in the house verdict-comment style (#232 precedent): verdict line, root cause with the failed-log frame, why-not-just-a-flake, heal commit, successor run id, ledger pointer — posted only AFTER the successor run went green so every claim in it was true at post time                                                                                                                                                                                                                                                                                                                                                                  | gh issue view 236: state CLOSED, comment 1320 chars landed; draft kept in docs/drafts/ (`3a1766bd`)                                                                                                                              |
| 9  | **Host-env gotcha recorded where the next session will hit it**: the agent shell exports `GOEXPERIMENT=jsonv2`; mkShellNoCC inherits it, and encoding/json flips to v2 semantics LOCALLY ONLY — `json.Marshal` errors on unexported-only structs where CI (clean env) passes `{}`. AGENTS § Commands now carries the signature (local-only red on a branded-struct marshal), the neutralization (`env -u GOEXPERIMENT`), and the rule: rule out the leak before trusting a local-vs-CI diverging verdict                                                                                                                                                                    | AGENTS.md, commit `d0ab66bb`                                                                                                                                                                                                     |
| 10 | **Session status report + INDEX**: this file's predecessor (`2026-11-07_19-25_m20-and-main-heal-status.md`) written, INDEX regenerated (72 reports) + `--check` clean — committed and CI-green (42401403869)                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `python3 scripts/status-index.py --check` exit 0                                                                                                                                                                                 |
| 11 | **Push hygiene honored**: every one of my pushes was watched to a verdict in the same session (the §d failure mode of the 11:07 session NOT repeated); the two manual pushes rode ratified bars (red-main-with-verified-fix for `2c50b2c7`; the docs commits rode the daemon once its leg recovered)                                                                                                                                                                                                                                                                                                                                                                        | /tmp/ci-heal.log `HEAL_CI_VERDICT=0`; /tmp/ci-final.log `FINAL_CI_VERDICT=0`                                                                                                                                                     |

## b) PARTIALLY DONE

1. **The concurrent battery red — diagnosed, not fixed** (deliberately).
   Run 42402222719 fails `fast flake checks` at the NEW docs-health
   battery step (their `43f70b9c`-train feature): `nix develop -c bash
   scripts/docs-gates.sh --check` exits 2 in CI. Verified mechanism: the
   step depends on the `docs-gates` binary that exists only in the host
   `~/.local/bin` — `nix develop .#ci -c command -v docs-gates` answers
   `/home/lars/.local/bin/docs-gates` HERE (host PATH leak) but the CI
   runner has no such path. Structural, not flaky. Their fix stream was
   landing as this report was written (5 commits 19:39–19:54 incl.
   `dcd11c05`, `fe0038c8`, `6007c093`; run 42403427494 in flight;
   uncommitted `scripts/strike-safety.py` + untracked
   `internal/app/health_crm_test.go` = their live working tree, left
   untouched per the concurrent-sessions rule).
2. **The §g trio from the 11:07 report**: Q1 dissolved (§a.4), Q2
   answered with evidence (§a.5) — but Q3 (ui-shots: human eyeball pass
   vs pixel-diff harness for the ten most-changed surfaces) remains
   genuinely owner-only and is now re-asked in this report's §g.
3. **Ledger completeness for the race row**: the 42308961304 row (the
   race itself) is still #236's draft wording "owner fills the verdict" —
   I closed the issue with the full verdict but have not yet appended the
   race's own settled row to docs/ci-flake-ledger.md (the successor-run
   green arrived minutes before this report; it is a 2-minute follow-up,
   listed in §f.3).
4. **The manual-push bar timer**: the daemon push leg stalled ~35–40 min
   again this evening (my ledger + AGENTS commits sat unpushed ~20 min
   before I pushed the heal myself under the red-main clause). The
   `scripts/push-stall-watch.sh` tool exists precisely for this; I did
   not run it (computed the bar by hand instead) — wiring it into the
   watch loop is still TODO row territory.

## c) NOT STARTED (correctly out of mandate)

1. **The battery-train completion** — the concurrent session's own work
   (battery hardening, strike-safety, palette announce, CHANGELOG range
   guard): theirs, actively landing; adopting it mid-flight would
   collide.
2. **The 50-item §f list** in the 11-07 resumed-completion report —
   rides TODO row 20's harvest; not this session's scope (the still-open
   items I noticed are carried into this report's §f so they survive).
3. **The ~65-report archive residue / cadence drain** — TODO row 11,
   the v9 recipe exists, long work.
4. **Datastar F4 (retire `/health` unsafe-eval)** — owner gate (TODO
   datastar row), untouched.

## d) TOTALLY FUCKED UP

1. **The ~40-minute forensics spiral on the local sso failure — the
   big one.** The failure was deterministically reproducible from a
   ONE-LINE read (`encoding/json/v2` in sso_test.go's imports, changed
   by `43f70b9c`), but I reached it LAST. In between I burned the
   session's credibility on a theory chain: host-env leak (true but
   NOT the cause) → build-cache poisoning → fresh GOCACHE → module
   graph import scans → go env file → binary build metadata → a
   wrong "linked-json/v2 flips semantics" hypothesis I half-believed.
   Each probe was individually sound; the SEQUENCE was backwards.
   Read the changed code first (`git log -L` on the failing lines /
   the file's imports), theorize second. The env leak I DID find is
   real and now recorded (§a.9), but it was never the cause of this
   failure — GOEXPERIMENT-unset still failed, and I kept probing
   anyway.
2. **25 idle minutes on a deterministic red main.** After diagnosing
   the break as the concurrent session's, I followed the
   concurrent-sessions deference into watching-and-waiting (~25 min,
   18:5x–19:1x) while main sat RED with a verified diagnosis in my
   head. The doctrine bars REVERTING others' work; it never barred
   healing at a neutral site — the domain-side MarshalJSON was
   available the moment the mechanism was understood. The
   "keep going until everything works" directive outranks deference
   to a possibly-finished session. I pushed at ~19:19 under the
   red-main bar; I should have pushed ~25 minutes earlier.
3. **Watch-verdict discipline nearly relapsed.** My FIRST CI watch
   (background shell 28C on run 42397931294) completed and its log
   sat unread for many minutes — the exact §d failure mode of the
   11:07 session (background `gh run watch`, verdict never read,
   red sat 35 min). I corrected on every later push (verdicts read
   promptly from the log files), but the relapse risk is real and
   the pattern needs a rule: a watch without a scheduled read is a
   decoration.
4. **My first commit's narrative was eaten by the daemon.** The M20
   commit raced the daemon's sweep: my `git add && git commit` lost
   ("nothing to commit, working tree clean"), the daemon landed the
   identical content as `chore: auto-commit 4 changed file(s)`
   (`5c368085`). Content survived; the WHY-record did not (the
   CHANGELOG line carries it instead). The stage+commit-in-one-call
   rule already exists; the daemon can still win that race — the
   only real mitigation is committing IMMEDIATELY after the last
   edit, not after a format+status detour (which is what I did).
5. **mtime-guard trips twice on TODO_LIST** — I edited without
   re-Viewing after daemon touches; the guard caught both. Cheap
   failures, but they cost two tool round-trips each.
6. **One direction slip in the sync check** — I read
   `HEAD..origin/main` when I meant `origin/main..HEAD`, briefly
   concluded "remote ahead" when the truth was "local ahead, daemon
   push stalled". Caught within one command; recorded because
   SHA-direction confusion during a stalled-push investigation is
   exactly when it hurts most.

## e) WHAT WE SHOULD IMPROVE

1. **Diagnose by reading the diff, not by probing the environment.**
   New rule candidate for AGENTS: when a suite reds after a daemon
   sweep, `git log -L` the failing lines (or read the file's imports)
   BEFORE any environment probe. The env-leak gotcha is real (§a.9)
   but it should be checked SECOND.
2. **A time-box on deference.** Concurrent-sessions hands-off needs a
   clock: if main is red on a defect another session owns, diagnose
   fully, wait bounded (≤15 min for a heal commit), then heal at a
   neutral site and say so loudly in the commit message. Deference
   without a deadline is abandonment of main.
3. **Watch = read.** A background `gh run watch` MUST come with a
   follow-up read of its log file in the same session phase; better,
   the AGENTS rule should say "never start a watch you will not read
   before your next commit".
4. **Commit at the speed of the daemon.** Post-edit detours (fmt,
   status checks) between the last edit and `git commit` are the
   daemon's window. Ideal: fmt + stage + commit in ONE call (fmt is
   deterministic for island files).
5. **The battery gate should not depend on a host binary.** Whatever
   the battery train decides (flake-shipped docs-gates, or a vendored
   runner), the gate must be reproducible on a bare CI runner —
   `~/.local/bin` is not a supply chain (this is v9 §g.8's canonical-
   binary question now wearing a red CI hat).
6. **Ledger rows should be appended at verdict time.** I closed the
   issue and updated the OTHER row but left the race's own row for
   later (§b.3). "After settling one, append a row" should be part of
   the close ritual, not a follow-up.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered by impact; items 1–10 from THIS session's evidence, 11–50
carried open items I noticed remain unaddressed.

| #  | Task                                                                                                                                                                                                                                   |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Land the battery train's CI fix (their stream is close: latest `6007c093` + run 42403427494 in flight) — watch it to green and record the verdict in the ledger                                                                        |
| 2  | Ship `docs-gates` through the flake (both shells) so the battery gate is reproducible on a bare CI runner — the structural half of §b.1, whatever their train decides about timing                                                     |
| 3  | Append the settled 42308961304 race row to docs/ci-flake-ledger.md (successor run 42400725432 green — the verdict is final, the row is owed)                                                                                           |
| 4  | Hunt the `GOEXPERIMENT=jsonv2` source in the shell rc (crushrc/config) — fix at the source or bless it and keep the AGENTS neutralization as permanent doctrine                                                                        |
| 5  | Sweep domain for sibling branded structs with unexported-only fields that the json/v2 migration will marshal next (domain.Phone, domain.RelPath, …) — add raw-value MarshalJSON proactively where a test or wire path will hit it      |
| 6  | Verify the concurrent session's next sso_test.go sweep PRESERVES the race-fix comment + non-parallel signature (a stale-buffer rewrite would silently regress #236)                                                                    |
| 7  | Wire `scripts/push-stall-watch.sh` into the watch.yml loop (or a session-start check) — the daemon push leg has now stalled 3 evenings running; the tool exists, nothing invokes it automatically                                      |
| 8  | Re-run the FULL island suite + tsc + buildflow once the battery train settles — their sweeps touched island tests (`palette.test.mjs` et al.) and my 318-green baseline predates their last three commits                              |
| 9  | §g Q3 carried: decide ui-shots evidence policy — recorded human eyeball verdicts vs pixel-diff assertions on the ten most-changed surfaces (ink flip, icons, type scale, uppercase removal all shipped visually unverified by a human) |
| 10 | §g Q2 residual ruling: do you want tracker issues for UNFILED red runs (the watcher missed four) — or is the ledger now the one home and the watcher's LIST-API suppression rule should be fixed instead?                              |
| 11 | `#wp-palette-open` aria-expanded state feedback (11-07 §f 39; v9 tail)                                                                                                                                                                 |
| 12 | Admin filter's pushed `/partials/admin?q=…` URL renders the bare partial on reload — decide if admin follows the History pattern (11-07 §f 42)                                                                                         |
| 13 | Keypad `✱`/`#` DE aria-label parity (11-07 §f 30 — parity test enforces keys, not semantics)                                                                                                                                           |
| 14 | `toggleTransferRow` returns null on toggle-close — dock chips could double-open; verify or return the row (11-07 §f 31)                                                                                                                |
| 15 | beforeunload guard with EMPTY returnValue — verify Firefox/Chrome honor it (Safari needs non-empty legacy value) (11-07 §f 32)                                                                                                         |
| 16 | In-call tab title vs notify.js titleFlash race during ringing→established (11-07 §f 33)                                                                                                                                                |
| 17 | Login Connecting label vs mid-flight language switch (data-i18n re-apply may overwrite it) (11-07 §f 34)                                                                                                                               |
| 18 | Sweep wirecontracts.go/apidocs.go for the new 422 contact-name response (11-07 §f 35)                                                                                                                                                  |
| 19 | Add the contacts-name cap to /openapi.json (the JSON door's 422 must appear in generated schema) (11-07 §f 36)                                                                                                                         |
| 20 | Admin read-only chip: bare "read-only" SR text — name WHAT is read-only or rule the subtitle sufficient (11-07 §f 40)                                                                                                                  |
| 21 | rateLimited(seconds) DE number formatting audit (11-07 §f 41)                                                                                                                                                                          |
| 22 | DTMF digit-shortcuts option: decide or close explicitly (11-07 §f 22)                                                                                                                                                                  |
| 23 | 0.68/0.72/0.75 micro sizes — schedule surfaces for the corpus catch-up (11-07 §f 24)                                                                                                                                                   |
| 24 | INDEX regen races the daemon — consider `--check` inside docs-gates (11-07 §f 25)                                                                                                                                                      |
| 25 | CDR missed-chip icon render at both themes — pin-tested but not screenshot-compared (11-07 §f 27)                                                                                                                                      |
| 26 | `#wp-tones` settle animation/politeness contract (hides after 3s; SR re-announce check) (11-07 §f 28)                                                                                                                                  |
| 27 | hx-disabled-elt audit: AI draft button + notes forms lack it (double-click class) (11-07 §f 29)                                                                                                                                        |
| 28 | sendIndicator ellipsis at button scale in dark AND light (11-07 §f 45)                                                                                                                                                                 |
| 29 | 0.9em→0.875em 12-site migration — one screenshot diff to prove visual harmlessness (11-07 §f 44)                                                                                                                                       |
| 30 | F20 Settings dl-count pins (v9 tail item)                                                                                                                                                                                              |
| 31 | Root-cause the three different live-report counters (INDEX 76 vs ls 75 vs cadence 70 — v9 §b.3, TODO row 21c)                                                                                                                          |
| 32 | Canonical docs-gates binary ruling (v8 §g.2 / v9 §g.8 — item 2 above is its implementation)                                                                                                                                            |
| 33 | THE SITTING: the 2026-10-29 decision packet + remaining briefing rows (unlocks ~10 planned tasks)                                                                                                                                      |
| 34 | Release announcements backlog post                                                                                                                                                                                                     |
| 35 | /context Kith-contract snapshot test (TODO row 12)                                                                                                                                                                                     |
| 36 | History nudge for the notes save path (TODO row 13)                                                                                                                                                                                    |
| 37 | scripts/templ-generate.sh wrapper pinning the compiler (TODO row 14)                                                                                                                                                                   |
| 38 | Carrier-leg owner question execution once ruled: inbound-MMS isolation test + GV verdict + number ownership (ROADMAP § Open questions, minted by v9)                                                                                   |
| 39 | Datastar F4: retire /health unsafe-eval — BLOCKED on owner green-light of the upstream dashboard nonce-lane change (TODO datastar row)                                                                                                 |
| 40 | Archive-residue drain: ~65 live reports / ~1600 §a-§b rows remain (TODO row 11, v9 recipe at scale)                                                                                                                                    |
| 41 | 2026-10-* planning/ cohort (~60 files) — next sweep's candidates as trains close (v9 §c)                                                                                                                                               |
| 42 | M38 arm: consider clearing pendingHistoryClear on navigation AWAY from history (state lives long; benign) (11-07 §f 38)                                                                                                                |
| 43 | DTMF trail role=status was set outside the create branch for stub semantics — verify the real DOM node still gets it on first create (my M20 session's neighbor; unverified in real browser)                                           |
| 44 | Simulate a load-failure STORM against the new toasts (SSE nudge + poll both failing) to prove announce()'s dedupe holds in the real browser, not just the stub                                                                         |
| 45 | The 11-07 §f 48 narrative pass over the session's ~20 daemon commits — still never done; yesterday's plan doc got 299 lines of changes nobody read                                                                                     |
| 46 | Fresh eyes re-review of the review (the 11-07 §f 50): the frontend-ux review that spawned this train has now been implemented + annotated twice — a third pass is how the NEXT train starts                                            |
| 47 | Consider a `docs/drafts/` README (or prune policy) — the dir now exists with one posted-comment draft and no stated retention rule                                                                                                     |
| 48 | Verify `git config core.hooksPath` resolves on fresh clones AFTER tonight's daemon commits (the commit-msg guard is the last defense the daemon bypasses via env)                                                                      |
| 49 | buildflow full run at the NEXT train start (the 74/74 pair this train relied on is now stale after ~15 more commits)                                                                                                                   |
| 50 | Re-measure .docs-gates.json ratchets at the next docs-health sweep (audit_max moved 333→355 in one day; DF-brittle per AGENTS — growing living docs flag old §f items)                                                                 |

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **The concurrent train (json/v2 migration + docs-health battery) is
   red-ing main while its fixes stream in — is that session yours and
   still active?** Concretely: the next time a deterministic red from
   THEIR files sits on main longer than ~15 minutes, do you want me to
   heal-forward at neutral sites (as I did tonight with
   domain.Extension), keep hands-off entirely, or ping you first?
2. **Should `docs-gates` ship through the flake** so the battery gate
   (now wired into CI's fast checks) runs on a bare runner without your
   `~/.local/bin`? And is `~/.local/bin/docs-gates` even the canonical
   build you want pinned (the v8 §g.2 dev-vs-nix question, now
   load-bearing in CI)?
3. **The host shell exports `GOEXPERIMENT=jsonv2`** — deliberate for
   your other projects (I keep the per-command `env -u` neutralization
   as permanent doctrine), or a shell-rc leftover to delete? Tonight a
   deterministic-looking local red from it cost real diagnosis time
   while the REAL cause was something else entirely.
