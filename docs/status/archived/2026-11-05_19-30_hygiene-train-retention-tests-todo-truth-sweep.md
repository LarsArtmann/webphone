> ARCHIVED 2026-11-08 (docs-health drain): gates clean (check-rows --strict + bare-rows green), live facts carried in AGENTS/TODO_LIST — cadence headroom for the SUPERB phase-closeout reports.

# Status: hygiene train — retention pins, TODO truth sweep, status index (assistant session)

**Snapshot-rot header: claims verified as of 2026-11-05 19:30 CET.** Push/CI/doc-state
claims rot fastest; everything below is timestamped to this minute.

## Scope — what this report VERIFIED vs did NOT verify

VERIFIED this session, with the artifact: `go build ./...` exit 0 (18:14, on the
bumped dep set) · full `go test -count=1 ./...` exit 0 (~18:20) ·
`go test -race ./internal/retention` green (twice: after each mutation restore) ·
two mutation proofs executed live (corrupt-path `continue`→`return` → pin FAILed;
window-guard removal → pin FAILed; both restored → pass) ·
`go test -race -count=1 ./internal/app ./internal/retention` isolated rerun exit 0
(19:26, after killing the wedged gate) · CI run 42075231935 GREEN via
`gh run watch --exit-status` (the concurrent bump train's re-pin commit) ·
`python3 scripts/status-index.py --check` exit 0, and exit 1 with a planted stray
report (mutation-proven) · `nix build .#webphone.goModules --rebuild` exit 0
(vendorHash `T2Q64uGG…` current) · tree clean, `git rev-list --count origin/main..HEAD`
= 0 at 19:27 (daemon pushed everything).

NOT verified, and why: the `buildflow --build-mode full` run never printed its own
final summary (wedged 25 min inside one step — see d1; its lint and test-coverage
steps passed in-log at 18:50) · island JS tests, `tsc` type gate, treefmt and the
KVM-gated VM checks (`nix flake check`) were not run — no island/nix file was
touched by this session, and CI's green run covers the fast checks on the same
commits · the wedge's root cause was not diagnosed beyond process evidence
(1 s CPU / 25 min, `futex_do_wait`) · the outbound-SMS 422 red row and every
owner-gated TODO row are untouched (owner-terminal legs) · this session's changes
have not been through a human eyeball — `git show` review of the daemon's commits
was spot-checks only.

## Context: two trains, one tree

This session ran CONCURRENTLY with another session's dependency bump train
(go-sse v0.7.0 + templ v0.3.1100 + loginpage v4.16.0 + cqrs-htmx v4.16.0 +
usermgmt v4.19.0 + go-health v0.10.0 + go-health-dashboard v0.13.2 +
templ-components v1.27.0). That train is the one this repo's TODO had called
BLOCKED since 2026-11-04 (loginpage was the blocker). Its session executed it,
re-pinned vendorHash after the named-hash CI red, and watched CI to green —
verified by THIS session, recorded in the fleet-libs TODO row, and left entirely
in its author's hands (one home per fact; no double bookkeeping).

## a) FULLY DONE (this session)

1. **`internal/retention` test package** — the destructive daily sweeper (deletes
   messages, attachments, fax documents, and unlinks blob files from disk) had
   ZERO tests since plan T25. Now: a deterministic `sweepOnce(ctx, db, blobs,
   engine, cutoff) sweepSummary` seam extracted from `Start` (behavior-preserving;
   `Start` keeps the disabled guard, panic recovery, and 24h ticker), plus four
   race-green tests asserting REAL outcomes: expired message/fax rows go AND their
   files leave the disk while live tenants survive untouched; a corrupt stored
   path (`/abs/escape`) skips itself without killing the sweep (the NEIGHBOURING
   valid blob still unlinks); window ≤ 0 spawns nothing (0 and −1h subtests);
   `Start` sweeps immediately on launch and honors ctx cancel. Both containment
   pins are mutation-proven (break the rule → pin FAILs → restore → pass).
   Semantic discovery pinned in the test: **a transcript's retention clock is its
   ARRIVAL time** (`Append` stamps `Seq = time.Now().UnixNano()`, not
   `StartedAt`) — a backdated segment is never sweepable, and the expired-segment
   deletion semantics stay owned by the readmodels suite.
2. **TODO_LIST truth surgery** (1113 → ~830 lines): executed sections DELETED
   per the file's own rule (live-smoke CI; the dock-train M18 + E2E tails; the
   error-agents shipped enumeration — compacted to its two remaining owner calls;
   the cqrs-htmx leverage tails — all four items resolved; the loginpage tails;
   catalog + wire-contracts shipped prose — open rows kept; the declarative tier
   log — owner calls kept; gate-recovery — F86 closed 10-25; the v2.8.0 deploy
   tail — prod verified on v2.9.0 remotely; the island-honesty E2E leg —
   discharged by the 2026-10-29 green run; the samber/do tails — folded into
   /health; the god-package carve record — replaced by a 10-line dormant-trigger
   section, blueprint in `docs/modularization/`; the hub cross-link `[x]`).
   Contradictions fixed at their rows: "#7/#8 remain unstarted" → actual states;
   catalog "push pending" header; the ai-phone dangling "caption lag above"
   clause. The 8-home stack-E2E sprawl collapsed to ONE home (Cross-repo
   obligations, which now records the 2026-10-29 discharge AND the standing
   re-owe rule). Every deletion was verified against CHANGELOG / AGENTS / git
   evidence before cutting.
3. **TODO_LIST front matter**: the 72-line append-only sweep diary (one >12 KB
   line) replaced by ONE "Last sweep" paragraph naming today's work — the file's
   own rule ("Done work is DELETED; one home per fact") now holds; sweep history
   lives in CHANGELOG + docs/status.
4. **Repo hygiene**: tracked regenerable artifacts untracked —
   `buildflow-fsprobe-2193036538` (blank buildflow probe file) and
   `ui-shots-before/` (14 stale PNGs, same "eyeball material" doctrine as the
   10-29 /ui-shots/ budget decision). Local files retained; gitignore extended
   (first pattern was dir-only and missed the file — caught and fixed same call,
   `git check-ignore` verified).
5. **`docs/status/INDEX.md` + `scripts/status-index.py`**: one row per report
   (135 today), newest first, generated — never hand-edited. `--check` compares
   content (names + titles), so the daemon's table reflow cannot trip it, and it
   is mutation-proven (a planted stray report flips it to exit 1). AGENTS.md
   Conventions carries the one-line rule.
6. **Bump-train verification recorded**: fleet-libs row rewritten from BLOCKED to
   LANDED with the CI-green run id, the vendorHash tail, and the one residual
   (ssetest `WithLastEventID(sse.EventID)` branding — correctly deferred: no
   Last-Event-ID consumer exists; the /events heartbeat row remains NO-GO).
7. **Five deferred engineering findings** (each verified at HEAD before writing)
   recorded as a TODO_LIST section — see e) below.

## b) PARTIALLY DONE

1. **The closing `buildflow --build-mode full` gate**: its golangci-lint and
   test-coverage legs passed in-log (✔ test-coverage 207.5 s, coverage reported),
   then the run WEDGED inside `go test -race ./internal/app` — 25 minutes
   elapsed, 1 second of CPU, `futex_do_wait`. Killed via the AGENTS
   python-kill recipe; the wedged leg re-ran green in isolation well under a
   300 s timeout. Net: every leg passed, but the TOOL's own final verdict never
   printed — the "gate green" claim is composed of legs, not the orchestrator's
   summary. Not re-run a second time (a 40-minute rerun to re-print a summary
   whose legs are all individually green is not a good trade; flagged here
   instead).
2. **status-index adoption**: generated + check exists + AGENTS rule written, but
   NOT wired into any CI lane (deliberate — whether it gates, and against what
   archive policy, is an owner call; see g2).
3. **HARVEST routing**: this report's section (f) is brainstorm-grade. The five
   verified hygiene findings ARE already in TODO_LIST; the remaining ~45 rows
   are deliberately unrouted (a larger N is a brainstorm, not a commitment
   list).
4. **CHANGELOG**: nothing added by this session — correct per the ratified bar
   (test-only and docs-only deltas stay silent; the retention refactor is
   behavior-preserving). FLAGGED, not fixed: the concurrent bump train has no
   `[Unreleased]` entry yet — that is its session's duty; dep bumps arguably
   deserve a standing bar ruling (see g-question 3, item 8 below).
5. **Decision-record consolidation inside TODO_LIST**: the surgery kept
   small NO-GO records in place (scheduled sends, session messenger, M19
   transfer) because the fleet-libs trigger registry covers LIBRARIES only —
   folding the product-side NO-GOs into one table was out of this session's
   budget and is listed in (f).

## c) NOT STARTED (observed this session, deliberately out of scope)

- CSP `report-uri`/`report-to` endpoint (violations unobservable today).
- Module-level eval warning for unfenced `/metrics` + `/version` +
  `/openapi.json` when `caddy.enable = false`.
- NixOS module secret-leak guard for freeform `settings` secrets
  (`gateway.webhook_secret`, TURN).
- Rate bucket on the templ form-POST surface (only `/api/*` lanes are limited).
- `server.Deps` read-model field collapse (accessor-style refactor).
- The stack `mod_enum` repair (gates the stack CI ceiling + every relock) —
  cross-repo, owner/stack-terminal.
- `nix flake check` on the post-surgery tree (island/type/format/VM legs).
- Any touch of the Android repo, the CRM push question (Q3), passkey owner
  legs, unified-auth cutover, fax carrier facts — all owner-gated rows, all
  left exactly as found.

## d) TOTALLY FUCKED UP (honest, in order of severity)

1. **The buildflow wedge itself** (~35 min wall): killed and worked around, but
   the run's final summary was lost and no stall detector exists — a step that
   prints nothing for 25 minutes should have alerted at 5.
2. **A multiedit anchored on the wrong trailing header** mangled TODO_LIST
   structure mid-surgery (deleted the Chat-networks header, duplicated
   Boot-contract). Caught within two tool calls and repaired, but a broken
   intermediate state existed on disk while an adversarial auto-committer was
   sweeping every few minutes. Section-scale edits should have gone through
   python span surgery from the start (which is what the remaining deletions
   used — and those went clean).
3. **`git restore` on a daemon-managed file during mutation testing**: the
   daemon had already committed my refactor, so restore landed on HEAD = the
   refactor — zero damage **by luck, not by care**. The correct tool (python
   reverse-replace) was used for the second mutation after learning. On this
   repo, `git restore` of my own just-written file is a loaded gun.
4. **Two consecutive parser bugs in `status-index.py`** (wrong backtick/bracket
   order in the link pattern; then naive splitting broken by dprint padding) —
   and worse, the check reported "everything drifted" instead of failing loud on
   the real signal: **parsed 0 rows**. A checker that parses zero rows must exit
   with a structural error, not enumerate 135 phantom findings.
5. **Overlapping edits inside one multiedit** produced a hybrid `(- [ ]` line —
   self-inflicted, repaired from raw bytes (`cat -A`), avoidable by
   non-overlapping anchors.
6. **Process deviation with consequences**: this harness forbids commits without
   an explicit user instruction, so per-task commits were left to the daemon —
   deviating from the repo's own "stage+commit in ONE call" convention. The
   git-restore near-miss in (3) is a direct child of that gap. The daemon
   covered the work (everything pushed, 0 unpushed), but the workflow honesty
   belongs in this report.

## e) WHAT WE SHOULD IMPROVE (lessons, each actionable)

1. Section-scale doc edits: python span surgery (find header → next header),
   never multiedit chains that share anchors — proven clean 7 deletions in a
   row after the two multiedit failures.
2. Checkers must fail loud on structural emptiness ("parsed 0 rows" ≠ "all
   drifted") — noise drowned the real bug for two runs.
3. Mutation testing on daemon-managed trees: python reverse-replace only;
   `git restore` assumes HEAD semantics the daemon silently changes.
4. `git check-ignore <exact-path>` on the exact artifact TYPE (file vs dir) as a
   standard step of every gitignore edit.
5. A buildflow stall-watch (no log line for N minutes → notify + offer the
   kill-and-narrow-rerun recipe) — the same observer culture as
   main-red-watch/daemon-audit, pointed at the gate itself.
6. Run a gate's heavy legs individually BEFORE the orchestrator when the tree
   carries a concurrent train — the orchestrator's verdict is then a formality,
   and a wedge costs minutes, not a third of an hour.
7. Keep the concurrent-session protocol (re-View before every edit, attribute
   foreign commits, never revert) — it worked; the one near-miss was
   self-inflicted, not interference.

## f) Up to 50 things to get done next (brainstorm — most are ROADMAP fuel; the five verified findings already sit in TODO_LIST)

1. CSP violation reporting: `report-to` endpoint (loopback) + Caddy route +
   module wiring decision.
2. Module eval-time warning when `caddy.enable = false` leaves
   `/metrics`/`/version`/`/openapi.json` unfenced.
3. Module secret guard: eval-reject literal `gateway.webhook_secret`/TURN
   secrets in freeform `settings` (environmentFile assert).
4. Rate bucket (or a documented bound) for the templ form-POST surface.
5. `server.Deps` read-model accessor collapse (next read model is the trigger).
6. Wire `status-index.py --check` into a CI lane (after the archive-policy
   ruling, g2).
7. `status-index.py` hard-error on parsed-0-rows + a `--self-test` (repo script
   culture).
8. `--json` output mode for status-index (agents; error-context.py precedent).
9. Attribute the buildflow race-test wedge (module-cache flock? memory? two
   concurrent buildflows?) and file it if reproducible.
10. Buildflow stall-watch script (no log line N min → alert; kill +
    narrow-rerun recipe printout).
11. Mutation-check convention: `scripts/mutation-check.sh` standardizing
    break→FAIL→restore evidence.
12. Retention: direct pin for the panic-recovery path via a sweep-deps
    interface seam.
13. Retention: pin the 24h ticker path with an injectable clock.
14. Retention counters → `/metrics` (sweeps, removed counts, duration) —
    observability for the destructive pass.
15. Retention dry-run mode ("what WOULD be deleted") for operator trust.
16. Blob-orphan sweeper: files with no row, older than a grace window
    (the documented "orphans stay until an operator prunes" gap).
17. Verify the session-store lazy expiry sweep actually fires (claimed in the
    retention doc, no test cites it — possible second 0-pin class).
18. Document that contacts are NEVER swept even with `retention_days` set
    (deliberate, operator-visible consequence).
19. TODO_LIST: migrate product-side NO-GO records (scheduled sends, session
    messenger, M19 transfer) into one trigger table mirroring the fleet-libs
    registry.
20. TODO_LIST: stable row ids so evidence links survive line drift.
21. README: consolidate the five "code complete, unverified" seams into one
    Limitations section (passkey ceremony, fast-answer live call, Android vs
    live FreeSWITCH, ring-while-closed seam, fax lanes).
22. FEATURES: sweep "owed" language now discharged (browser E2E ran green
    2026-10-29).
23. /health: ONE doc home for the auth posture (README says unauthenticated;
    module fences via vhost forwardAuth).
24. Rate-limit NAT-fold note (per-IP buckets fold shared egress) — document or
    session-key the buckets.
25. Verify passkey login rotates the CSRF token like password login.
26. SECURITY.md: incident-response runbook pointer (reporting exists; response
    steps don't).
27. templ compiler pin guard: check that go.mod's templ and the devshell CLI
    agree (guards the "v0.3.1114" class recurrence TODO explicitly says nothing
    guards).
28. CHANGELOG bar ruling for dependency bumps (owner call; the just-landed
    train shipped without an [Unreleased] line).
29. `nix flake check` on the post-surgery tree (island-js, typecheck, treefmt,
    module, KVM legs) — none run this session.
30. CHANGELOG entry for the bump train (its session's closeout; verify it
    lands).
31. Archive cull for docs/status (134+ reports; docs-health ANNOTATE battery;
    resolved reports → `archived/`, INDEX already excludes it).
32. docs-gates battery adoption (HARVEST/freshness tooling — currently manual).
33. `md-table-shape.py` run over the generated INDEX (shape-tool cross-check).
34. Decide local fate of the now-untracked `ui-shots-before/` files (delete
    locally or keep as eyeball reference).
35. `.git/objects` garbage: six `tmp_obj_*` files observed in
    `git count-objects` — `git prune` hygiene.
36. `legacy_*` store files (contacts/faxes/snippets/transcripts): confirm
    load-bearing or delete (dead-code sweep candidate).
37. usermgmt v4.19.0: AuditLog projection consume decision (owner call,
    recorded in AGENTS 2026-10-23).
38. go-health-dashboard v0.13.2: re-check the fence memo still holds (registry
    row exists).
39. Stack-side: `services.webphone.paperless` module option + smoke arm
    (cross-repo row).
40. Stack-side: ring hook dialplan POST (`webphone:/hooks/ring`) — until it
    lands the push seam is inert (High, cross-repo).
41. Stack-side: ops-runbook demo-call recipe + secret PATH column.
42. Owner packet: the six 2026-10-29 rulings (group rooms, E2EE, pma#348,
    tw-freshness, config typo strictness, branching-flow triage) — one sitting
    unlocks ~10 tasks.
43. Outbound SMS 422 on prod: the journal leg (the file's only 🔴 row).
44. Passkey owner legs: live ceremony + fail-closed drill + ULID-secrecy
    verdict.
45. Unified-auth T13/T15 cutover cluster (owner-gated, High).
46. Config typo strictness (L7): file-strict/env-warn design + owner ruling.
47. Branching-flow governance triage: the 7 new findings + baseline save.
48. Kith CRM live e2e (#9 owner steps: token → push-secrets → deploy →
    acceptance bar).
49. webphone-android: first live device train (the seam both repos owe).
50. This report's own closeout: TODO_LIST HARVEST for any of (f) the owner
    promotes (the five verified findings are already routed).

## g) Three questions I cannot figure out myself

1. **Buildflow wedge posture**: should the fleet get an automated stall-watch
   (no gate-log line for N minutes → notify + print the kill/narrow-rerun
   recipe), or do you prefer manual triage? It is fleet-wide BuildFlow
   behavior, so it is your call, not a per-repo script I should add.
2. **docs/status archive policy**: keep ~135 reports flat with the generated
   INDEX as the only access layer, or move fully-resolved/old reports to
   `archived/` (INDEX already excludes them)? This decides whether `--check`
   becomes a CI gate now or after a cull.
3. **Promotion call**: of the five verified hygiene findings (CSP reporting,
   probe fencing, freeform-secret guard, templ POST rate bucket, Deps
   collapse) — which, if any, should be promoted from "recorded finding" to a
   scheduled train? I can rank them (secret guard and probe fencing are
   security-shaped; the rest are polish), but scheduling is yours.

**Evidence pointers:** retention diff + tests (`internal/retention/`), TODO_LIST
diff (this commit range), `scripts/status-index.py`, `.gitignore` tail, AGENTS.md
Conventions, CI run 42075231935, `/tmp/buildflow-final.log` (wedge evidence),
`git rev-list --count origin/main..HEAD` = 0.
