# Status: v2.6.0 shipped through the tag — release tail blocked on host contention, chained retry armed

> UPDATED 2026-09-23 (docs-health): the chained attempt NEVER FIRED —
> verified: `/tmp/release26d.log` gone, NO gh release object for
> v2.6.0, origin == HEAD (`0cb6d2c`), tag `807ca0c` on origin; the
> stack pins webphone `7503561`. The tail (stack E2E ×2 green, gh
> release, aarch64 re-verify, smoke `--expect-version 2.6.0`,
> pbx-artmann relock #4) is the TODO release row — THIS report stays
> the operative pointer. g1–g3 routed to ROADMAP open questions + the
> owner-calls row; e1–e5 improvements ride the TODO runbook-hardening
> row + the ROADMAP infra/protocol asks.

_2026-09-23 02:47 CEST. Session: CRM idempotency + metrics train executed to green; release
train advanced to the signed tag; remaining steps (stack E2E → aarch64 → gh release) blocked
on host load from parallel agent sessions. A chained wait-for-sustained-quiet job is live and
will fire the release automatically (background job, log at `/tmp/release26d.log`)._

## TL;DR

All code work is DONE and green (CRM idempotency end-to-end, metrics family, docs). The v2.6.0
tag is **created, signed, pushed, and gate-verified** (`807ca0c` on origin). The release's
remaining steps (lychee → stack E2E → aarch64 → gh release) failed twice on E2E stalls that
were load-shaped, not code-shaped — the host ran at load 60–174 tonight from parallel AI
sessions. Honest count: the first stall had a self-inflicted component I own (my own background
flake check). What we should improve: a load precondition in release.sh and a cross-session
gate protocol.

## a) FULLY DONE (implemented + verified green this session)

1. **CRM call-journal idempotency, end to end.** Server: `POST /api/calls` accepts a UUID
   `key` (≤128 chars, extension-namespaced, own 1h `callsIdem` idemStore with its own TTL
   constant). Replay = inert 204; the key is recorded only on delivered outcomes (both 204
   shapes, including the unknown-number drop); a 502 stays retryable; an absent/empty key
   keeps the legacy never-dedupe shape. OpenAPI schema documents `key` with the contract.
   Pinned by four new contract subtests in `TestAPICallLoggingContract`
   (`internal/server/crm_test.go`): same-key-journals-once + key-scoping (different key
   journals again), 502→retry→succeed→replay-deduped, absent-key-never-dedupes (two
   keyless reports journal twice), drop-consumes-the-key (unknown number first, known
   number with the same key after → still zero journals). All green under `-race`.
2. **Island side.** `recordCrmCall` (`internal/web/assets/island/app/panels.js`) sends
   `crypto.randomUUID()` per ended call; oxlint `no-undef` passes (`browser` env covers
   `crypto` — no new globals needed). New `panels-crm.test.mjs` exists BECAUSE
   `config.js` reads `PBX_CONFIG.crm` once per module registry — a second test file gives
   a fresh process with crm ON; it pins the exact wire body (number/direction/seconds/
   outcome/key), fresh-UUID-per-call, and the 502 warn-toast copy carrying `HTTP 502`;
   `panels.test.mjs` gained the crm-off no-op assertion on its own crm-absent instance.
   6/6 green.
3. **`/metrics` CRM family.** `webphone_crm_lookups_total{outcome="hit|miss|failure"}`
   (counter) rendered only when `h.deps.CRM.Enabled()`; the family is ABSENT entirely when
   off — a disabled deploy must not publish zero-lines that read as "CRM broken". Pinned
   both directions: `TestMetricsRendersCRMLookupCounters` drives one real hit (known
   number), one miss (empty results), one failure (upstream server CLOSED → transport
   error) and asserts the three sample lines + HELP/TYPE + aggregates-only leak guard
   (no numbers, no contact names); `TestMetricsServesAggregatesOnly` asserts absence with
   the integration off.
4. **Full gates over my changes:** `nix develop -c go test -count=1 ./...` green (14 pkgs);
   island suite 79/79; `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` = 53 success / 0 failed
   (5 oxfmt line-wrap findings in my new .mjs files auto-fixed via
   `buildflow -s oxfmt --fix`, re-verified, committed); `nix flake check` = all checks
   passed (incl. the KVM-gated webphone-backup VM test and island-lint/oxlint).
5. **Docs:** CHANGELOG 2.6.0 addendum — five CRM entries (idempotent journal, resolver
   single-flight + counters, metrics family) plus the concurrent ux-raw-ideas train's two
   features (thread search E, audio output picker F) folded per the default fold policy,
   and the `crm.Client` `do()` chokepoint under Changed; AGENTS.md CRM-seam paragraph
   extended with the hardening facts (chokepoint closes the split brain, single-flight,
   counters → metrics gating, island UUID key contract); FEATURES.md — CRM row updated
   (single-flight, idempotency, metrics), Thread search + Audio output picker rows added.
6. **Concurrent-session cooperation held all night:** landed their uncommitted 21:46
   markdown/formatter sweep with attribution (8956fa9 — verified formatting-only first:
   `git diff -w` collapsed 709/711 to 94/96, all emphasis/list-marker normalization);
   re-verified the whole server suite green over their `recordCallIdem` /
   `apiContactSaved` refactor train (e8939b7, race on); landed their uncommitted
   follow-up doc updates (e01f75d). Never reverted anything of theirs.
7. **Release progress through step 5/9:** `release: v2.6.0` bump commit (b162e22) pushed;
   **signed tag `v2.6.0` created, pushed, and verified on origin (`807ca0c`)**; the FULL
   gate suite inside release.sh — buildflow full, `go test -count=1 ./...`, nix flake
   check, live smoke, vulnix over the runtime closure (every finding triaged
   distro-patched by the webphone-vulnix-triage CLI) — PASSED against the tagged tree.
   Daemon stalls hand-pushed at every phase boundary (precedent maintained, verified via
   `git ls-remote`, not push output).

## b) PARTIALLY DONE

1. **release.sh steps 6–9 (lychee → stack relock → stack E2E → aarch64 → gh release).**
   PARTIAL (2026-09-23 evening): lychee 0 errors, stack pin verified at `7197f1c`
   (the lock legally rides main; the stack's own CRM+TURN surgery is `be876ae`
   there), aarch64 re-verified by ELF `b7 00`. STILL OWED behind the host-load
   gate (load 19–104 all evening): browser E2E ×2, `gh release create`, smoke.
   - Run 1 died at the lychee gate: one broken relative link in an archived planning doc
     (`docs/planning/archived/2026-09-18_21-45_*` → status report moved under
     `docs/status/archived/` after the plan was archived). FIXED (7503561) and verified
     with a targeted lychee run (0 errors) before resuming.
   - Run 2 (resume) relocked the stack and died in the stack browser E2E waiting for
     `CONTACTS-ROUNDTRIP-OK` — AFTER the entire post-FS-restart transfer saga passed green
     (FS-RECOVERED → TRANSFER-BLIND-INITIATED → channel settle → `->9196` leg →
     TRANSFER-CALLER-RELEASED → TRANSFER-CALLEE-MEDIA).
   - Run 3 (resume) stalled EARLIER, at the transfer channel-settle wait
     (`fs_cli show channels | grep '^1 total'`, 180s) — the transfer initiated fine 22s in.
   - Diagnosis: two failures, DIFFERENT steps, same 180s marker-wait shape in the fragile
     post-restart tail = environment variance (host load 60–174 from parallel agent
     sessions), not a deterministic code regression. Run 1 of these two proved the whole
     transfer path works end-to-end.
   - Attempt 4 is CHAINED and LIVE: waits for load < 6 sustained (2 consecutive minutes),
     then fires `release.sh 2.6.0` automatically, logging to `/tmp/release26d.log`.
2. **Stack relock + pbx-artmann relock #4.** The stack re-relocked to webphone main during
   the resumes (pin needs re-verification against e01f75d on the next run — the pin and
   the v2.6.0 tag legally diverge per the 2026-09-20 owner decision and the v2.5.0
   precedent). The E2E against any relock has not passed yet; pbx-artmann relock #4 comes
   after the stack goes green.
   UPDATE 2026-09-23 evening: pin re-verified (`7197f1c` ≥ e01f75d); the stack landed its
   CRM+TURN surgery (`937b94f`+`be876ae`, eval checks green incl. new CRM coverage); E2E
   still owed (load-gated); pbx-artmann relock #4 blocked by a concurrent session's dirty
   AGENTS.md/FEATURES.md in that tree.
3. ~~**TODO_LIST harvest.** Deliberately untouched until the release completes (rows change~~ done (2026-09-23 docs-health HARVEST — release row rewritten around the TAIL)
   ~~state; the file was also mid-edit by the other session earlier tonight).~~

## c) NOT STARTED

1. `python3 scripts/webphone-smoke.py --expect-version 2.6.0` — the explicit post-release
   verification run (release.sh runs its own smoke during gates; the runbook's
   expect-version check is still owed after the tag exists).
2. pbx-artmann relock #4: rev-parse webphone main → pbx-artmann `flake.nix` telephony URL
   → `nix flake update telephony` → `nix run .#lock-drift-probe` → build both toplevels →
   verify webphone ExecStart store path moved → narrative commit + push.
3. The three owner questions from the 21:12 status report remain unanswered; I executed the
   documented defaults (fold policy → concurrent train's landed items folded into 2.6.0;
   UUID-key idempotency → implemented exactly as designed; push-daemon handling → kept
   hand-pushing at phase boundaries).

## d) TOTALLY FUCKED UP

1. **I fired release attempt 2 while my own background `nix flake check` was still
   running** (started ~23:40 for pre-verification, still building VM tests when the
   release E2E ran at 01:45). The E2E VM competed with my own KVM backup VM test + builds
   for CPU. Self-inflicted contention, plausibly the whole cause of that stall. From
   attempt 3 onward: nothing of mine runs during a release.
2. **Two of three attempts launched without a load check.** The host carries multiple
   parallel agent sessions tonight — 58 users, load 60–174; I directly observed
   `golangci-lint run` at 250% CPU, `go mod vendor`, `providers.test`, and an unrelated
   `nsfw-classifier-go` build. Attempt 3 started into a 20-minute quiet window that
   closed mid-E2E (load 149 during the run). The fix — require SUSTAINED quiet before
   starting — is in the live chained attempt, one attempt too late.
3. **Monitoring blind spot, twice:** release attempts ran behind `| tail -40`, which
   buffers ALL output until process exit — I was blind mid-run and reconstructed progress
   from git state and 7000-line logs instead of watching a file. From attempt 3: output
   redirects to `/tmp/release26d.log` and I tail the file.

## e) WHAT WE SHOULD IMPROVE

1. ~~**release.sh should gate on host load** — read `/proc/loadavg` in preconditions (e.g.
   refuse to start the E2E phases above load ~8, or at least WARN), or the stack E2E
   should auto-retry once on a marker-stall. Either would have saved ~40 minutes tonight.~~
   done at `a24496a` — `load_gate()` (1-min loadavg ≥8 refuses; `WEBPHONE_RELEASE_MAX_LOAD`
   override; tested both directions) + `assert_clean_tree()` at tag time. The
   load-gate-vs-retry ratification is an open ROADMAP question.
2. ~~**Codify the flake heuristic with step context** in AGENTS/runbook: "same E2E step
   stalls twice in a row = dig into code; different steps inside the post-FS-restart tail
   = host load, wait for quiet and retry." Tonight I had to derive this from two full
   logs; it should be a one-line rule.~~ done at `a24496a` — "Hard-won release rules
   (2026-09-23 tail)" in docs/release-runbook.md.
3. **Cross-session load protocol:** nothing tells a session "another session is running
   gates right now". A tiny flock convention (`/tmp/webphone-gates.lock` holding pid +
   scope) would let concurrent sessions yield instead of colliding.
4. ~~**The auto-commit daemon sweeping mid-release is a live hazard:** it committed f50e825
   (a docs file, harmless) BETWEEN the tag push and the lychee step, and earlier swept my
   in-flight edits mid-work twice. release.sh asserts a clean tree only in preconditions;
   asserting at each step boundary (or at least before the tag) would catch drift.~~ done
   (the "at least before the tag" variant) at `a24496a` — `assert_clean_tree()` runs at
   tag time; per-step-boundary asserts stay unimplemented (daemon-exclusion ask lives in
   ROADMAP infra).
5. ~~**Release-attempt logging:** always `> /tmp/release-<v>-<n>.log`; never pipe a
   30-minute multi-phase script through `tail`.~~ done at `a24496a` — runbook rule.

## f) NEXT (prioritized)

1. ~~**Verify the live chained attempt:** when load < 6 sustained, it fires~~ done (NEVER FIRED — verified: log gone, no gh release, TAIL row owns the retry)
   ~~`release.sh 2.6.0`; check `RELEASE-EXIT=0` at the end of `/tmp/release26d.log`.~~
2. On success: confirm the stack E2E passed, the stack relock commit exists and pins the
   intended webphone rev, aarch64 cross-builds verify by ELF machine bytes (`b7 00` at
   offset 0x12), and `gh release view v2.6.0` shows the extracted CHANGELOG body.
   PARTIAL 2026-09-23 evening: aarch64 `b7 00` re-verified; pin verified (`7197f1c`);
   E2E and the gh release object still owed (load-gated).
3. `python3 scripts/webphone-smoke.py --expect-version 2.6.0` → 0 failed.
4. pbx-artmann relock #4 (see c2): rev swap → lock-drift-probe → both toplevels →
   ExecStart store-path moved → narrative commit + push.
5. ~~TODO_LIST harvest: release row → DONE (v2.6.0, released 2026-09-23, tag `807ca0c`);~~ done (2026-09-23 sweep — release row rewritten, CRM row shrunk to e-h, pins row closed)
   ~~dedup-pins row → DONE (listRows / requireMultipartTo / pbx nil-client / ListThreads~~
   ~~all shipped earlier in this train); CRM follow-ups row shrinks to remaining~~
   ~~stack-side items (e)–(h): `crm.{url,token}` NixOS wiring, stack runbook cross-doc,~~
   ~~restore-drill, one HTTP-level integration test — (a)–(d) are SHIPPED this train.~~
6. ~~CHANGELOG: nothing owed — the addendum is already folded.~~ done (E+F were folded; the four A-D bullets completed by the 2026-09-23 sweep)
7. ~~Re-surface the three owner questions (21:12 report + this report §g); if still~~ done (routed — ROADMAP open questions + owner-calls row)
   ~~unanswered, note the defaults taken.~~
8. Closing sweep per runbook step 9: `git ls-remote` verify main + tag end states,
   narrative commit at the phase boundary.

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Load precondition or E2E auto-retry in release.sh?** Tonight's two failures were
   both load-shaped; a load check is cheap with no downside, a single E2E retry masks
   real regressions at +10 min per true failure. Which do you want — and should E2E
   retries count toward the "two consecutive over-budget" watch?
2. **Stack pin vs tag divergence:** the stack will pin webphone main (e01f75d — carrying
   the other session's `apiContactSaved` refactor) while the v2.6.0 tag pins b162e22.
   This matches the documented "ride main" decision and the v2.5.0 precedent
   (lock ≠ tag), but the delta now includes another session's refactor covered only by
   per-package tests + the upcoming stack E2E. Acceptable, or re-pin the stack to the
   tag itself for release trains?
3. **Host contention convention:** the load is partly OTHER projects' agent sessions
   (nsfw-classifier-go, providers.test, parallel nix builds) — not just webphone. Do you
   want a flock-file protocol (sessions yield to a held gate lock), staggered schedules,
   or is "wait for sustained quiet" the accepted de-facto protocol?
