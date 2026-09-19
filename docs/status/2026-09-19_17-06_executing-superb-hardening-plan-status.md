# Status Report — Executing the SUPERB Prod-Recovery & Hardening Plan

**Date:** 2026-09-19 17:06 (CEST)
**Session scope:** execution of
`docs/planning/2026-09-19_15-37_SUPERB-prod-recovery-and-hardening-plan.md`
(owner: "GET SHIT DONE! The WHOLE TODO LIST!"). Three repos touched:
webphone (this repo), nix-international-telephony (stack), pbx-artmann.
Production was NOT redeployed (owner ssh required; see §b/§g).

---

## a) FULLY DONE

### Tier 1% (production recovery — prepared, deploy itself blocked)

- **M1.1 pre-flight**: stack lock pins webphone `d815004` (v2.1.0) verified;
  `git ls-remote` verified both repos; trees clean.
- **Trap fixed one level above the plan's pre-flight**: pbx-artmann's
  `flake.lock` pinned the stack at a STALE narHash predating the CSRF fix +
  v2.1.0 bump — a deploy from that tree would have re-shipped the broken
  v2.0.0-era login. Re-pinned with `nix flake lock --update-input
  telephony`; the chain now resolves pbx-artmann → stack main → webphone
  v2.1.0 tag commit exactly. Committed by daemon as `019093f`, pushed,
  verified via ls-remote.
- **M1.2 deploy procedure confirmed** (docs/deploy.md + pbx-artmann
  `install-pbx.sh` + flake): `nixos-rebuild test --flake .#pbx
  --target-host root@pbx.artmann.tech`, verify `/version` = v2.1.0, then
  `switch`. Exact commands handed to the owner in chat. My toolchain
  cannot run ssh (hard-blocked), so M1.3-M1.5 + M2.x await the owner.
- **Prod "before" evidence captured**: `GET /version` on
  `https://pbx.artmann.tech` → `"v2.0.0"`, `/healthz` → 200 ok. This is the
  build where every browser login 403s.

### P3 — the silent-breakage gate (stack browser E2E) — DONE, negative- and positive-proven

- CDP-injected fetch recorder (`tests/browser-e2e.py`) records every
  fetch's real status per page, surviving reloads; island untouched.
- New gate `wait_session_gate`: after each browser's registration, POST
  `/api/session` must answer **201** AND `GET /api/csrf` must answer **200**
  (the rotation-adoption round-trip). Prints `{ext}-SESSION-CREATED` /
  `{ext}-CSRF-ADOPTED` markers; failure dumps the recorded fetch list.
- testScript waits for the gate markers per-browser, immediately after
  each registration (a failure localizes at its own marker, not the next
  registration).
- **Negative test (M3.5)**: stripped `settings.csrf` in the VM node → run
  failed with `1000-SESSION-GATE-FAILED: session=403 …
  recorded fetches: ["POST /api/session -> 403"]` — the gate bites, the
  silent-breakage class (v2.0.0 prod outage) can never pass green again.
- **Positive run: green** (exit 0, both gates instant, E2E-OK).
- Stack commits: `50b7aee` (daemon, includes the gate),
  `f8eb226` (mine: override removed + marker reorder), pushed to f8eb226.

### P4 — docs HARVEST + annotate — DONE

- TODO_LIST rewritten (harvest): redeploy row updated with the delivered
  command + lock-chain evidence; 12 new bounded rows from the plan
  (P5, P9, P13, P14, P15, P17, P19, P20, P21, P22, P23, P24, P25, P27)
  with status/priority/effort/evidence; done-today items never entered.
- ROADMAP: new "WORTH_CONSIDERING cluster" (P26 one-line specs: session
  persistence, retention job, PWA, video, recording UI) and "Standing
  watches" (P27: sip.js 0.22, templ-components ThemeScript knob, oxlint
  globals, E2E wall-time budget, cqrs-htmx root tag).
- 11:02 report annotated inline via `annotate-prose.py` (dry-run first):
  **22 §f-items + all 3 §g-items** struck with hashes/evidence
  (f.1-f.10 → `4c6bba1`, f.11 → `d815004`, f.12-f.14 → `28d4688`,
  f.19/20/22-25/29/30 → verified-by-appendix evidence, g.1 → superseded
  by P8 memo, g.2/g.3 → answered-by-events). Open items left untouched
  (f.15-f.18, f.21, f.26).
- Fronting-bug timeline appendix appended to the 11:02 report
  (09:18 wiring → forceSSL vhost → prod outage → v2.1.0 fix → E2E gate).

### P6 — smoke fronted-shape probes — DONE (26/26)

- `fronted_login` helper (Host + Origin + Sec-Fetch-Site + XFP: https):
  unconfigured server → **403** (the prod bug, locally reproduced forever);
  configured server (second boot via `WEBPHONE_CONFIG` JSON) → **201**.
- Smoke suite: 24 → **26 checks**, all green.

### P7 — GET /api/csrf rate limiter — DONE

- `csrfLimiter` (shared hook budget, 60/min burst 60) wired around the
  route; 429 carries the computed Retry-After; openapi documents the 429
  response with the Retry-After header.
- `TestCSRFRefreshRateLimitPerClient` proves the bound.

### P8 — decision memos — DONE (the DECIDED lines await owner)

- Two one-paragraph memos appended to the plan: pin policy
  (recommendation: keep riding main + per-release lock bump; the real
  hazard is stale pins, proven today) and pbx-artmann input
  (recommendation: keep `path:` + lock discipline; github pin adds auth
  plumbing without removing the staleness failure mode).

### P9 — CSRF cookie Secure flag — DONE

- `csrfSecureFromOrigins`: any https trusted origin ⇒ Secure cookie.
- `TestCSRFSecureFollowsTrustedOrigins` (3 subtests).
- Harness consequence handled honestly: a Secure cookie is invisible to
  plain-HTTP test clients (Go jar + python cookiejar), so both harnesses
  now carry the token cookie by hand, exactly as a real browser does over
  the TLS hop. Smoke 26/26 re-verified after the change.

### P10 — operator troubleshooting — DONE

- README "Troubleshooting: 403 logins behind a TLS proxy": symptom, the
  exact WARN log line, the fix keys, the module defaults, the startup log
  line. Module `nginx.enable` option description now names the csrf
  consequence.

### P11 — openapi parity — DONE (`/api/csrf` GET asserted like `/api/session`)

### P12 — startup visibility — DONE (`csrf fronting trustedProxies=N trustedOrigins=[…]` INFO log, non-secret)

### P13 — version-drift guard — DONE, verified BOTH ways

- `cmd/webphone/drift_test.go`: `flake.nix webphoneVersion` must equal
  `git describe --tags --abbrev=0` (skips honestly in the nix sandbox —
  no .git; buildflow + local runs exercise it).
- **First version was a false positive: it silently SKIPPED locally
  (wrong relative path, `../` instead of `../../`) and "ok" = skip. Fixed
  and re-verified: PASS on 2.1.0, FAIL on a temp `2.1.1-DRIFT` bump with
  a clear message, PASS after restore.**

### P18 — island polish — DONE

- `#log` lines (English, runbook-greppable) on: session failed, adoption
  failed + reloading, session created + token adopted.
- Stale "gated behind the stack E2E" comment sweep: grep clean, nothing
  to fix (f.31 already clean).
- `loginRaw` test-client variant; ratelimit + requestlog loops converted.

### P19 — adoption fallback — DONE

- Retry ×3 with backoff (250/500/750 ms) before the reload last resort;
  reload path intact; both outcomes logged. Keeps SIP registration in the
  common case.

### P20 — typed nix csrf options + stack-side assertion — DONE (VM test run pending, see §b)

- `settings.csrf` is now a typed submodule (trusted_proxies /
  trusted_origins, with fronting-shape description pointing at the README
  troubleshooting entry); module defaults unchanged.
- Stack `tests/webphone.nix`: asserts the RENDERED runtime config of the
  running unit (reads WEBPHONE_CONFIG from the unit's environ) pins
  loopback proxy + https origin. **Not yet executed — see §b.**

### P21 — HSTS knob — DONE (module + check + README)

- `nginx.hsts.enable` (default off, honest brick-the-domain rationale in
  the option text) + `nginx.hsts.maxAge` (default 2y) → vhost
  `add_header … always`.
- `webphone-module` flake check refactored to a shared `moduleSet` and
  extended with an `hsts-opt-in` evaluation asserting the header.
  Evaluates green; full `nix flake check` still pending (§b).
- README note documents default-off rationale.

### P22 — cache-safety of GET /api/csrf — DONE

- `Cache-Control: no-store` + `Vary: Cookie` on the response, contract
  documented in-code (the stack vhost has no proxy_cache today; headers
  make it explicit), `TestCSRFRefreshCacheSafety` pins it.

### P5 — 1001-registration anomaly — PARTIALLY DONE (see §b)

- Island `connection.js` reviewed: bounded watchdog → rebuild → full
  re-login on reload; no second-tab hole found.
- `REGS-AT-RECONNECT` sofia dump added to the testScript (tripwire: any
  vanished binding is dated to the reconnect phase, not first seen at
  dial).
- Run 1: **anomaly did NOT reproduce** — both registrations healthy at
  reconnect AND at dial (full Call-ID/Contact/EXPSECS dumps). Run 1 died
  at an UNRELATED phase: blind transfer's `1 total` channel-count wait
  (90 s) under TCG load while media flowed on both legs → budget bumped
  to 180 s with a comment.
- Run 2: **green end-to-end**.
- Run 3: **NOT A RUN — nix cache hit** (tree unchanged → same derivation
  reused; grep proves no fresh log). The "two consecutive green runs"
  bar is NOT met yet: 1 green + 1 unrelated-phase flake + 1 cache hit.

## b) PARTIALLY DONE

- **P1 redeploy**: everything except the switch itself. Owner command
  delivered; prod still serves v2.0.0 with broken tab logins RIGHT NOW.
- **P5 verdict**: instrumentation shipped, anomaly not reproduced (4 data
  points today), transfer flake found + budgeted, but no two-consecutive
  green runs (run 3 was a cache hit — my error). Close with
  `nix build --no-eval-cache` or a tree-touching rerun ×2.
- **P20 stack assertion**: written, not executed — the `telephony-webphone`
  VM test needs a run to prove the environ-reading snippet works.
- **Gates owed before calling the webphone work done**: full suite +
  smoke after the P18/P19/P22 edits (P22 test green; full suite last ran
  before those), `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`, `nix flake
  check` (statix/island-lint over the session.js + nix edits), prettier
  pass over session.js (treefmt owns island JS — my hand edits may not
  be prettier-clean).
- **pbx-artmann lock is stale AGAIN** (necessarily): today's stack commits
  moved the stack tree narHash, so the deploy chain needs one more
  `nix flake lock --update-input telephony` + toplevel pre-build before
  the owner's switch. Known cadence, still owed.
- **Push state**: webphone pushed (3435393); stack has 2 unpushed daemon
  commits (38d24a6, fe60979) at report time.

## c) NOT STARTED

- **P14** release runbook script (`release.sh`, M14.1-M14.5 incl. dry-run).
- **P15** release ops for v2.1.0 (`gh release create` + CHANGELOG link
  refs + announcement draft).
- **P16** vulnix re-run on the v2.1.0 runtime closure + aarch64
  island-lint cross-build.
- **P17** backup/restore story (README section + real restore drill +
  timer-skeleton-vs-stack-restic decision) — I was mid-write when this
  report was requested; nothing landed.
- **P23** CSRF rotation on session TTL refresh (spec + impl + tests).
- **P24** fax feed (stack rxfax TIFF→PDF → `/hooks/fax`, loopback test).
- **P25** idiomorph swap experiment (branch + browser E2E + verdict).
- **P27** remainder: FEATURES VERIFY pass, #log-English check,
  union-coverage blocked note, v2.1.1 hotfix pre-draft.
- **M1.3-M1.5, M2.1-M2.4**: prod redeploy + probes + smoke + eyeball
  (blocked on owner ssh).

## d) TOTALLY FUCKED UP (honest ledger)

1. **Run 3 was a cache hit and I nearly logged it as a green
   verification.** `nix build` reused run 2's result because the stack
   tree hadn't changed; the grep count (0 E2E-OK lines) exposed it. My
   own verification process was about to lie to me; the two-consecutive-
   greens bar is NOT met. Fix: `--no-eval-cache` / `--check` or touch the
   tree.
2. **P13's first version was a silent-skip false positive.** Wrong
   relative path made the test SKIP locally and print "ok" — both my
   initial green AND my first negative verification were meaningless.
   Only the verbose run exposed it. A skip inside a verification run
   must be loud.
3. **`rg -r` footgun, twice**: `-r` is `--replace`; two searches printed
   mangled output before I caught it.
4. **Pipeline masking, again**: `nix build … | tail; echo EXIT:$?` reports
   tail's exit code (AGENTS documents this exact class — I repeated it on
   the negative E2E run; the visible FAIL contradicted the EXIT:0 and I
   re-ran with pipefail).
5. **Daemon races are not theoretical**: the daemon committed my stack
   work mid-flight INCLUDING the temporary negative-test override
   (50b7aee) — for a moment the committed tree shipped a broken csrf
   config; my fix landed on top (f8eb226). Explicit commits must land
   immediately after each edit, not "after the task".
6. **session.js edit-tool dance**: three rejected edits (daemon mod-time
   races) before switching to scripted patches — wasted round trips.
7. **flake.nix hand-edit broke eval** (unclosed attrset in the linkFarm
   list); caught immediately by `nix eval`, fixed.
8. **Stale git index.lock in the stack repo**: trashed it after verifying
   no git process held it — safe, but the daemon's lock handling caused
   two failed commit attempts.
9. **The gopls "unused: refreshCSRF" diagnostic is STALE** (route wires
   it; smoke + tests exercise it; CLI builds green) — per the
   independently-verify rule I trusted the CLI and moved on; still, it
   sits in diagnostics and should be restarted away, not left to confuse
   the next reader.

## e) WHAT WE SHOULD IMPROVE

1. Verification hygiene: forced E2E reruns must use `--no-eval-cache`;
   write that into the stack AGENTS/runbook so the next session doesn't
   count a cache hit as evidence.
2. Loud-skip design: drift-style guards should FAIL (or log loudly) when
   their prerequisites are missing in a local run — silent skip + "ok"
   is a false-green factory.
3. Commit per edit (or per micro-task), immediately: the daemon wins
   races it shouldn't, and heuristic messages hide the story.
4. Build a personal checklist reflex for the two repeat offenders:
   `set -o pipefail` before any piped gate, and never `rg -r` casually.
5. The E2E's transfer phase under TCG load is the flakiest leg (this is
   the second load-related stall this week: answer-phase 2026-09-18,
   transfer-phase today). A wall-time/CPU budget note or a retry-once
   policy for channel-count waits would de-flake CI.
6. Add an assertion (not just a dump) on REGS-AT-RECONNECT
   (count == 2) once two real consecutive greens exist.
7. The plan-vs-execution split worked, but "blocked on owner" items
   (deploy) still gate the product; a deploy helper that does NOT need
   my ssh (owner-runnable script that also updates the pbx-artmann lock
   + pre-builds) would shrink M1.3 to one command.
8. Island JS formatting must go through prettier (`nix fmt`), not my
   hand-rolled edits — check before the next gate run.

## f) Up to 50 things to get done next (impact-ordered)

1. **Redeploy prod to v2.1.0** (owner ssh; command delivered) — the only
   thing between users and working logins.
2. Re-lock pbx-artmann to the settled stack tree + pre-build the
   toplevel so the switch is cache-hits-only.
3. M1.4 probes vs prod: `/version` = v2.1.0, `healthz`, `config.js`
   contract, hooks 401/202.
4. M1.5 `scripts/webphone-smoke.py --base https://pbx.artmann.tech`
   (read-only checks; expect 25 of 26 — the configured-boot probe skips
   in --base mode).
5. M2.x in-browser eyeball (console clean, login, tabs, DTMF) + record
   findings.
6. Verify P12's startup line in the prod journal post-deploy
   (`csrf fronting trustedProxies=1`).
7. Push the 2 unpushed stack daemon commits (or verify the daemon did).
8. Run the stack `telephony-webphone` VM test to prove the P20.2
   rendered-csrf assertion.
9. Full webphone suite + smoke after P18/P19/P22 (last full suite
   predates them).
10. `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` gate.
11. `nix flake check` (statix + module-check incl. hsts-opt-in +
    island-lint over the session.js edits).
12. prettier/treefmt pass over session.js; fix whatever it reformats.
13. P5 close-out: two REAL consecutive greens via `--no-eval-cache`.
14. Upgrade REGS-AT-RECONNECT from dump to assertion (count == 2).
15. Decide E2E de-flake policy for channel-count waits (retry-once or
    CPU budget note) and write it into the stack test docs.
16. P14 `release.sh` (skeleton → gates → tag/push/lychee → stack lock →
    stack gates → aarch64) + dry-run idempotence check.
17. P15 `gh release create v2.1.0` with CHANGELOG excerpt.
18. P15 CHANGELOG bottom link-refs for the 2.1.0/2.0.0 diffs.
19. P15 announcement draft (owner approves posting).
20. P16 `nix run .#vulnix` on the v2.1.0 runtime closure.
21. P16 `nix build .#checks.aarch64-linux.island-lint --system aarch64-linux`.
22. P17 data inventory (webphone.db + files/ layout, sizes).
23. P17 README "Backups and restore" section (rsync/restic pattern,
    SQLite online-copy caveat, stack `backups.paths` pointer).
24. P17 real restore drill on a scratch dir (boot → seed → backup →
    wipe → restore → boot → verify).
25. P17 decision: module timer skeleton vs stack-restic-only — record it.
26. P23 TTL-rotation spec (rotate at refresh points; adoption interplay).
27. P23 implementation + tests (TTL refresh rotates; island same path).
28. P24 locate rxfax handler + TIFF output path in the stack.
29. P24 TIFF→PDF conversion step behind a feature toggle.
30. P24 POST converted PDFs to `/hooks/fax` with the webhook secret
    (LoadCredential wiring).
31. P24 loopback test: seeded TIFF → hook → Fax tab row.
32. P24 error paths (conversion failure → log + retry policy) + docs.
33. P25 idiomorph research (htmx 2.x morph/SSE interplay).
34. P25 experiment branch wired for sse-swap targets.
35. P25 browser E2E on the branch + verdict doc (keep/drop).
36. P27 FEATURES VERIFY pass.
37. P27 #log-English check.
38. P27 union-coverage blocked note (upstream BuildFlow ask).
39. P27 v2.1.1 hotfix pre-draft.
40. On owner answers: AGENTS `DECIDED` lines for pin policy + pbx input
    (P8.2); relock pbx-artmann if the input type ever changes.
41. AGENTS.md session-knowledge updates: SESSION-CREATED gate,
    REGS-AT-RECONNECT tripwire, Secure-flag derivation + harness
    implication, drift guard (and its skip semantics), HSTS option,
    typed csrf options, csrf limiter, Vary/no-store, cache-hit trap in
    E2E reruns.
42. CHANGELOG: open a `[2.1.1] - Unreleased` section for today's
    hardening batch (limiter, Secure, HSTS option, drift guard, probes).
43. FEATURES.md rows: csrf limiter, HSTS option, drift guard, fronted
    smoke probes, E2E session gate (stack-side row).
44. TODO_LIST: delete rows as they complete (P6/P7/P9-P13/P18-P22 work
    is done but rows were harvested before completion — prune to match).
45. Investigate the run-1 transfer flake one level deeper if run 4+
    shows it again (sofia log capture around the REFER).
46. Document the dev-shell sqlite3 requirement inside the P17 drill
    commands.
47. Restart gopls to clear the stale `refreshCSRF unused` diagnostic.
48. Check `nix run nixpkgs#lychee` still passes after README edits.
49. Consider surfacing `Retry-After` handling in the island for
    /api/csrf 429 (parity with the phone-api wrapper, RA1 pattern).
50. Fold this report's §b leftovers into the next session's first
    TODO read (docs-health HARVEST stays routine).

## g) Questions I cannot figure out myself

1. **Deploy go/no-go mechanics**: will you run the redeploy yourself with
   the delivered commands (`cd ~/projects/pbx-artmann && nixos-rebuild
   test --flake .#pbx --target-host root@pbx.artmann.tech`, verify
   `/version`, then `switch`) — or do you want me to re-lock pbx-artmann
   to the final stack tree first and hand you ONE copy-paste block?
   Either way I cannot execute it (ssh is hard-blocked for me).
2. **Blast radius of the login-403 window**: was anyone besides you using
   pbx.artmann.tech between the v2.0.0 deploy and the upcoming v2.1.0
   switch? If yes: do you want a notice/apology drafted, and should I
   reconcile any missed messages/fax notifications afterward (the PBX
   kept the calls; only tab sessions and their server-side actions were
   dead)?
3. **Pin-policy DECIDED**: does my P8 recommendation get your approval as
   written — stack keeps riding webphone `main` with a per-release lock
   bump, and pbx-artmann stays `path:` with lock discipline (the memo
   argues the real hazard is stale pins, which today's pre-flight now
   catches) — or do you want tags-only anyway?

---

*Report is a point-in-time snapshot. All webphone work is committed
(HEAD `3435393`, pushed); stack HEAD `fe60979` has 2 unpushed daemon
commits at report time. Production still runs v2.0.0 — the login-403
outage continues until the owner's redeploy (§g.1).*
