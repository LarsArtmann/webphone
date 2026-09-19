# Status: TODO sweep completion, browser-E2E bug hunt, and the version cut — 2026-09-19 06:42

Scope of this report: the overnight run in `webphone` (continuing the
2026-09-18 TODO_LIST sweep through its final items), the upstream browser
E2E failure hunt and fix in both `webphone` and
`nix-international-telephony`, and the (interrupted) v2.1.0 version cut.
Parallel-session activity observed in the same repo is noted where it
changed the ground under this session.

Baseline when this run started: 16 of 23 TODO items already done and
verified; all local gates green; the upstream switchover wired by a
parallel session; browser E2E never run against v2.

---

## a) FULLY DONE

1. **go.mod hygiene verified** — `golang.org/x/time` already direct at
   go.mod:18; the gopls "should be direct" warning was stale. No change
   needed; verified by build.
2. **`.buildflow.yml` policy completion** — added documented
   `cqrs-lint` skip (all 3 findings contradict repo design: rejected
   setup bundle, transitive root-library-only imports, cqrs-htmx-owned
   version pins) and `nix-hash-fix` skip (buildflow itself warns it can
   never write the hash here; 10/10 chronic failures).
   `buildflow verify-config` passes.
3. **`-coverpkg` TODO resolved as upstream-blocked** — BuildFlow's
   canonical `.buildflow.yml` key list has no per-step-args mechanism;
   a repo-local coverage script would duplicate an orchestrated step.
   Routed to ROADMAP.md as a BuildFlow feature idea.
4. **Docs overhaul** — TODO_LIST pruned to open work only; FEATURES.md
   upserted (fax page parsing, transcript pagination, history filter,
   vCard, login rate limiting, keyboard shortcuts, NixOS module, arch
   tests, i18n → FULLY_FUNCTIONAL; session persistence →
   WORTH_CONSIDERING); CHANGELOG Unreleased extended (i18n, NixOS
   module, rate limiting, vCard, pagination, history filter,
   shortcuts, theme toggle, fax pages, arch tests, test coverage,
   unread cache, pbx query fix); README gained the NixOS module
   section, vCard in the capability table, i18n/deployment rows
   updated; AGENTS.md records the deployment decision, arch-test
   enforcement, `wp-lang`/SSE-language contract, pbx JoinPath gotcha,
   i18n dictionary rule, vulnix build-vs-runtime closure scope, and
   the extended skip list; ROADMAP.md created.
5. **server_test.go split by concern + twin dedupe** — split into
   session_test.go, messages_test.go, fax_test.go, sse_test.go,
   proxy_test.go; the three webhook tests joined webhooks_test.go;
   the duplicated 14-line sign-in/subscribe blocks (jscpd finding)
   collapsed into the existing `signIn` helper plus a new
   `subscribeEvents` helper (t.Cleanup-based). Full suite green after.
6. **sip.js "0.22" evaluation** — verified via npm registry + GitHub
   tags: **no 0.22 exists**; npm latest is 0.21.2 (2022-10-27); an
   unreleased 0.21.3 tag adds a 7-line SimpleUser option irrelevant to
   the island (raw UserAgent). Verdict: stay on 0.21.2; the reconnect
   watchdog remains load-bearing. Report:
   `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`. TODO row
   deleted; FEATURES row → WORTH_CONSIDERING.
7. **Nixpkgs channel bump** — flake.lock nixpkgs → 2026-09-17
   (e554fab); `nix build .#webphone` rebuilt clean on the new channel.
8. **Honest CVE scoping** — vulnix over the whole build closure (20
   derivations, mostly bootstrap toolchains that never deploy) vs the
   **runtime closure** (`nix-store -qR`, 8 derivations): only
   glibc-2.42-84 carries advisories, headlined by CVE-2026-5450 (9.8),
   still unfixed in the freshest nixos-unstable. TODO re-scoped to
   "track until nixpkgs fixes it". AGENTS.md documents the
   build-vs-runtime scan distinction.
9. **treefmt-check unblocked** — the flake's prettier glob swept the 3
   historical docs HTML snapshots (one unparseable); removed `*.html`
   from prettier's includes (no app HTML exists; historical snapshots
   are immutable by docs-health policy). treefmt-check builds again.
10. **NixOS module eval check fixed** — the `webphone-module` flake
    check failed on missing `services.nginx`/`systemd`/`users` options
    in the minimal evalModules; added grouped stub options (statix-
    clean form). Check builds; also silenced the remaining erraudit
    warnings in the vCard import with reasoned `nolint`s (on the
    `if err != nil` line — placement matters).
11. **All quality gates green at the sweep's end** — full
    `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` passed (exit 0);
    `nix flake check` all checks passed; aarch64 cross-build ok; live
    loopback smoke (healthz 200, page renders DOM contract + theme
    toggle, config.js serves) ok.
12. **Upstream browser E2E run + root-cause hunt** — first-ever v2-era
    run failed (stall before CALL-ESTABLISHED). Diagnosis chain:
    FS siptrace showed the callee sent 100 Trying + 180 Ringing but no
    200 OK, then NO_ANSWER at exactly 60s; per-browser diagnostics
    (added to the stack's E2E on the stall path) showed the callee
    logged the click but `answerIncoming` never executed.
13. **ROOT CAUSE + FIX** — `internal/web/assets/island/app/main.js`
    imported `{ placeCall, renderCalls, sendDtmf, teardownAll }` from
    calls.js but **not** `answerIncoming`/`rejectIncoming`; every
    on-screen Accept/Reject click died with a silent ReferenceError
    (keyboard shortcuts worked because shortcuts.js imports them).
    Fixed, committed as `00f13fe` ("fix: wire accept/reject button
    handlers to the island's answer functions").
14. **Browser E2E fully green after the fix** —
    `.#telephony-browser` (two headless chromiums through nginx
    wss:// → FreeSWITCH): registration, wrong-password UX, reconnect
    watchdog, real 1000→1001 WebRTC call with media proof, and
    transfer all pass (exit 0).
15. **Stack-side server test green** — `telephony-webphone` NixOS VM
    test passes against the new stack wiring (isolates FS/nginx/
    service wiring from browser JS).
16. **Push happened** — the auto-commit daemon published `main`
    (28+ commits incl. the fix; `00f13fe` verified as ancestor of
    origin/main).
17. **Docs for the fix** — CHANGELOG Unreleased/Fixed gained the
    accept/reject entry; TODO_LIST/ROADMAP updated for the switchover
    outcome.

## b) PARTIALLY DONE

1. **v2.1.0 version cut — made, then overwritten.** I renamed
   `[Unreleased]` → `[2.1.0] - 2026-09-18` and updated the compare
   links; minutes later concurrent session edits (or the daemon)
   overwrote the file back to a single `[Unreleased]` against
   `v2.0.0...HEAD`. Current state: CHANGELOG carries all 2.1.0 content
   under Unreleased; **no `2.1.0` heading, no tags exist** (only
   v0.1.0). The cut must be redone atomically (edit → commit → tag →
   push in one motion) when no other session is mid-edit.
2. **Stack lock bump** — the telephony stack's `webphone` input was
   last seen pinned to `276c596` (pre-fix). The fix is now pushed, so
   the bump is `nix flake update webphone` + commit; not yet done, and
   the pinned rev should be re-checked (the parallel session kept
   committing).
3. **TODO_LIST row drift** — my "push the fix + bump the lock" row is
   half-obsolete (push done by the daemon); needs re-scoping to just
   the lock bump. TODO_LIST also grew to 8 rows: 5 new cqrs-htmx
   Pareto-plan rows arrived from the parallel session.
4. **Stack E2E diagnostics instrumentation** — added a clearly-marked
   TEMP-DIAG answer-phase dump (island `#log`, call-card states,
   `window.__wpDiag`) to `tests/browser-e2e.py`; committed there as
   b96d4c2. Genuinely useful for future stalls, but it is an
   unannounced change to Lars's repo — keep or revert is an owner
   call. After the island breadcrumbs were removed, the `diag:` field
   is inert.
5. **glibc CVE-2026-5450** — everything doable locally is done
   (channel bump, rebuild, honest scoping); the actual fix is
   upstream-blocked. Stays as a tracking row.
6. **This status report itself** — Markdown, not the skill's canonical
   HTML dashboard, on your explicit instruction (override flagged per
   the status-report skill).

## c) NOT STARTED

1. Tags `v2.0.0` and `v2.1.0` (only `v0.1.0` exists) — and therefore
   the lychee 404s on the CHANGELOG release/compare links persist.
2. Post-cut verification: lychee green, compare links resolving.
3. **Browser E2E re-run after tonight's post-green island changes** —
   the parallel session's `718cbe7` ("prettier reflow of island
   429/throttle additions") and adjacent commits touched island assets
   AFTER my green run; the same wiring-bug class (silent
   ReferenceError) is exactly what only the E2E catches.
4. The 5 new cqrs-htmx Pareto-plan TODO rows (parallel session's
   scope: SSE Broadcaster collapse, webhook hardening, hub limiter/
   reaper follow-ups).
5. sip.js bump — deliberately parked by the evaluation (re-trigger
   conditions documented).
6. `-coverpkg` union coverage — parked upstream (BuildFlow feature).

## d) TOTALLY FUCKED UP

Nothing in the product: at report time the build passes, all 10 test
packages pass, and the last full browser E2E is green. Process damage,
honestly:

1. **Concurrent-edit collision on CHANGELOG.md** — I built the v2.1.0
   cut on a file I had read 8 minutes earlier while another session
   was actively editing; my rename was silently overwritten. Cost: the
   cut has to be redone. Lesson (now obvious in hindsight): re-stat the
   file immediately before every edit in a multi-session repo, and do
   version cuts as one atomic edit→commit→tag→push motion.
2. **Wasted control experiment** — the "old webphone under new stack
   wiring" bisect control confounded two variables (old binary's
   config-shape drift vs new vhost), producing an earlier, unrelated
   failure. The per-browser island-state dump was the probe that
   actually cracked it; should have gone there first.
3. **Blocked 26 minutes on a stale `.git/index.lock`** (no git process
   alive) before removing it — recognized it as stale too late.
4. **The question tool failed 4 consecutive times** ("single_choice
   requires at least 2 choices", choices array arrived empty) — burned
   round trips before switching to plain-text questions.
5. **LSP phantom diagnostics cost repeated re-verification** — "bytes
   imported and not used" / "x/time should be direct" / unused-type
   warnings that contradicted green builds; per AGENTS.md the CLI run
   wins, but each false alarm still cost a verification cycle.

## e) WHAT WE SHOULD IMPROVE

1. **Multi-session concurrency protocol**: re-read (or re-stat) every
   file immediately before editing; never trust an earlier read in a
   repo another session is actively committing to.
2. **Version cuts as an atomic motion**: CHANGELOG edit, commit, tag,
   push, and the consuming lock bump in one uninterrupted sequence —
   never left half-applied across a report or pause.
3. **Make the browser E2E a pre-publish gate, not on-demand**: the
   accept/reject wiring bug shipped through heuristic commits and was
   only caught because tonight's mandate forced an E2E run. A GitHub
   Action (KVM runner) or a daemon hook that runs `.#telephony-browser`
   before webphone pushes would close the class.
4. **Island import/export cross-check**: a tiny lint (node script or
   ESLint no-undef on module scope) that catches referenced-but-
   unimported identifiers — `node --check` is syntax-only and missed
   it; the Go DOM contract cannot see JS wiring.
5. **Keep the answer-phase diagnostics** in the stack E2E (with owner
   sign-off) — they turned a 180-second opaque stall into a
   one-look diagnosis.
6. **vulnix should scan the runtime closure by default** in gates;
   build-closure noise (bootstrap gcc/binutils/Go) trains people to
   ignore the scanner.
7. **TODO_LIST needs a concierge**: with two sessions executing in one
   repo, rows went stale within hours (my push row was half-obsolete
   before I finished the run). A quick TODO_LIST re-triage at the end
   of every session.
8. **LSP distrust discipline worked but cost time** — codify: when LSP
   contradicts a green CLI run, file it as noise immediately and move
   on (already an AGENTS.md rule; follow it faster).

## f) NEXT — up to 50 things (brainstorm; most are ROADMAP fuel, top ~10 are TODO_LIST material)

1. Redo the v2.1.0 cut atomically (CHANGELOG rename + links, commit,
   tag v2.1.0 at HEAD, push).
2. Create tag `v2.0.0` at `b390c7a` (the v2-rebuild + sweep state) so
   both lychee 404s resolve.
3. `nix flake update webphone` in the telephony stack onto the
   post-fix rev; commit; verify the stack's pins.
4. Re-run `.#telephony-browser` against the bumped pin (airtight chain:
   published tag → stack lock → E2E green).
5. Re-run `.#telephony-browser` again after tonight's island
   429/throttle commits (`718cbe7`) — island JS changed post-green.
6. CI: wire the browser E2E (or at least an island import-lint smoke)
   into webphone's pre-push/publish path.
7. Island identifier lint: no-undef cross-check over
   `internal/web/assets/island/app/*.js` (would have caught the bug).
8. Go test that greps the island for referenced-but-unimported
   bindings? (cheap stopgap for 7 if no JS toolchain lands).
9. Confirm the telephony stack lock rev post-bump and note it in the
   stack's TODO/report (avoid another silent drift).
10. Prune TODO_LIST: my push row → "lock bump" only; keep tag row
    until tags exist.
11. lychee re-run after tags — expect 0 404s.
12. HARVEST this report into TODO_LIST/ROADMAP per docs-health (the
    (f) list below the top ~10 is ROADMAP fuel).
13. Annotate yesterday's status reports that this run superseded
    (`2026-09-18_18-50_todo-list-sweep-status.md` §f items now done).
14. Stack: decide keep-vs-revert of the TEMP-DIAG E2E block (owner).
15. Stack: `tests/webphone.nix` (server VM test) after the lock bump.
16. BuildFlow upstream feature request: per-step test args
    (`-coverpkg`), fleet-wide value (ROADMAP documents it).
17. vulnix gate scoped to the runtime closure (script or BuildFlow
    step) so advisories that matter are visible.
18. Track glibc CVE-2026-5450 until nixpkgs fixes it (existing row).
19. glibc fix lands → bump + rebuild + re-scan + close the row.
20. sip.js: only on reconnect-hang fix/security/need (evaluation
    report's re-trigger conditions).
21. Session persistence decision (FEATURES WORTH_CONSIDERING) — owner
    product call, security tradeoff documented.
22. Retention/cleanup job for blobs (FEATURES WORTH_CONSIDERING).
23. Richer /healthz (store, gateway mode) for LBs.
24. Video calls (sip.js supports; UI surface needed).
25. PWA/offline shell under strict CSP.
26. cqrs-htmx Pareto plan: SSE Broadcaster collapse row (parallel
    session's, appears in progress — `aacae89` did part).
27. cqrs-htmx plan: webhook hardening row.
28. cqrs-htmx plan: hub limiter/reaper follow-ups (429/throttle landed
    tonight; E2E re-run missing — see 5).
29. Docs: record the browser-E2E invocation recipe
    (`nix build -L .#telephony-browser`, override-input pattern for
    pre-publish testing) in webphone AGENTS.md.
30. Docs: record the bisect wrapper-flake trick (module-from-HEAD +
    package-from-rev) — it unblocked rev-level isolation.
31. AGENTS.md: note that history is occasionally rewritten by the
    daemon (remote 404s on local revs; local refs may differ from
    pushed hashes) — affects anyone pinning revs.
32. Consider tagging discipline: annotated tags + release notes per
    version (the CHANGELOG entries already carry the material).
33. Metrics: island keyboard-vs-button answer usage is unobservable —
    optional telemetry is out of scope under strict CSP; skip unless
    product demands (noted to prevent re-proposal).
34. Fax: surface page count in the fax tab rows (parsing exists; UI
    may already show it — verify, then close or file).
35. History: date-range filter beyond `?q=`/`?dir=` (UI capacity
    exists server-side).
36. Messages: unread badge TTL cache metrics (hit/miss) for tuning.
37. Contacts: vCard export of shared directory (currently personal
    only) — product decision.
38. NixOS module: add tests option / backup hooks (owner ops call).
39. NixOS module: document lmtp/webhook secret rotation runbook.
40. README: deployment matrix (module vs reverse-proxy) already exists
    — add a concrete `imports` snippet for the stack's actual wiring.
41. Accessibility pass on the island (focus order in incoming-call
    panel; shortcuts cheat-sheet visible in UI, not just log).
42. i18n: island dictionary and views dictionary key-sync test
    exists — extend to cover dynamic templates (missedCall etc.).
43. SSE: reconnect backoff tuning after tonight's Broadcaster
    collapse (verify heartbeat behavior under load).
44. Rate limiting: expose 429 Retry-After to the island UI (server
    sends 429s now; island shows generic failure).
45. Load test: two-browser E2E under CPU-constrained VM (TCG) to know
    the NO_ANSWER margin (60s ring timeout was never at risk with KVM;
    TCG may differ).
46. Backup/restore drill: SQLite + blob store round trip.
47. Dependency sweep: templ-components / cqrs-htmx minor bumps through
    the buildflow update flow.
48. Docs-health sweep of `docs/planning/2026-09-18_21-45_*` (parallel
    session's plan) against tonight's actual outcomes.
49. Archive/annotate superseded 2026-09-18 status reports (docs-health
    ANNOTATE mode, inline markers).
50. Retrospective item: agree a hand-off convention between concurrent
    sessions (one-line "I'm editing X" note in a shared scratch file)
    to prevent the next CHANGELOG collision.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Tag naming**: should `v2.0.0` point at `b390c7a` (the v2 rebuild
   + sweep state that the stack pinned) with `v2.1.0` at current HEAD
   — or do you want a single `v2.0.0` at HEAD and no 2.1.0? (Decides
   how I redo the CHANGELOG cut; both resolve the lychee 404s.)
2. **The stack's E2E diagnostics block** (commit b96d4c2 in
   nix-international-telephony, TEMP-DIAG marked): keep it as a
   permanent stall-diagnostic, or revert it now that the bug is fixed?
3. **Post-green island changes** (`718cbe7` 429/throttle prettier
   reflow + adjacent commits): did the parallel session already
   re-run `.#telephony-browser` after those, or should that E2E re-run
   be the next action before the stack lock bump?
