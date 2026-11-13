# SUPERB — Webphone for the Next 20 Years (Durability Pareto Plan)

> ARCHIVED 2026-11-07 (docs-health v9 sweep): EXECUTED — the six durability
> theses are load-bearing AGENTS rules (honest probes, boot contract, error
> families, morph swaps, CSP, backup drill). Row-level verdict pass DONE this
> sweep: all rows struck with verified evidence. Open legs route forward:
> announcements posting + the owner sitting (both TODO_LIST rows).

- **Date:** 2026-09-22 12:02 CEST
- **Input state:** TODO_LIST (8 rows), ROADMAP (open ideas, watches,
  long-tails, open questions), FEATURES (PLANNED/WORTH_CONSIDERING
  rows), the annotated 2026-09-19/20 reports' still-open f-items, the
  2026-09-22 status report (1001 anomaly closed, E2E ×2 green, both
  repos' PUSHERS DOWN, prod still on v2.1.0 with the forged-session
  hole).
- **Kind:** execution plan (point-in-time snapshot; TODO_LIST stays the
  living source — new items surfaced here get pointer rows there).

---

## 0) What "superb for 20 years" MEANS for this codebase

Six durability theses, each grounded in something that already happened
in this repo's first three weeks of life:

1. ~~**Security debt must never queue behind infrastructure.** A known~~ done — standing — deploy chain per-train (release-runbook); v2.5.0+ shipped
   ~~auth bypass (bogus creds mint sessions) has been live on prod since~~
   ~~2026-09-19 — not because the fix is missing (shipped in v2.2.0+,~~
   ~~validated), but because deploy = one owner ssh command behind a~~
   ~~broken pusher daemon. For 20 years: merged→live must be hours.~~
2. ~~**The rarest failure must leave evidence.** The 1001 anomaly took 3~~ done — standing — failure instrumentation shipped (dumps, watchdogs)
   ~~days + 5 E2E runs because browser #log vanished with the crashed~~
   ~~page. Fixed (pre-call dumps, sofia tripwires, cycle watchdog) — keep~~
   ~~every new failure path instrumented by default.~~
3. ~~**Dependencies rot on their schedule, not ours.** sip.js upstream is~~ done — standing — TODO_LIST § Standing watches (next 2026-12-20)
   ~~DORMANT at 0.21.2; chromium drifts monthly; nixpkgs churns. The~~
   ~~named-trigger watch system is the right shape — feed it.~~
4. ~~**Knowledge is the product.** AGENTS.md at 657+ lines vs a 377~~ done — standing — docs-gates ritual + AGENTS 377-line cap
   ~~budget; reports need annotation rituals; concurrent sessions race a~~
   ~~commit daemon. Docs-health is a standing cadence, not a one-off.~~
5. ~~**Boring releases, boring deploys.** release.sh is proven twice; the~~ done — standing — docs/release-runbook.md owns the train ritual
   ~~×2-green E2E rule and a re-baselined wall-time budget keep the train~~
   ~~trustworthy. Never make the train exciting.~~
6. ~~**The product surface still has user-visible gaps** (contacts~~ done — largely closed — later trains (contacts, failure UX, a11y: rows above)
   ~~staleness, message failure reasons, a11y). Durability includes being~~
   ~~worth keeping.~~

## 1) Pareto Breakdown

- **The 1% that delivers 51%:** SHIP THE FIX CHAIN (T1–T5). Push both
  repos → stack re-pin + full stack check → pbx-artmann relock → owner
  deploy + probes → post-switch verification + SMS-bridge journal grep.
  One afternoon. Closes the prod security hole, makes the anomaly fixes
  live, and proves the whole tri-repo chain end to end.
- **The 4% that delivers 64%:** THE DEPLOY/RELEASE MACHINE + CLEARING
  DECISIONS (T6–T11). Cut v2.5.0 (fold + runbook), re-baseline the E2E
  budget, post the announcements, fix the pusher + narrative-commit
  discipline, and ONE owner-calls batch session that clears ~14 open
  decisions (each of which blocks or shapes downstream work).
- **The 20% that delivers 80%:** THE DURABILITY KERNEL + VISIBLE PRODUCT
  (T12–T22). Contacts API completion, SSE contacts nudge (last stale
  surface), session/CSRF hardening, gate/tooling tail, docs kernel,
  island test tail, backup/probe contracts, upstream tails, the UX/a11y
  pass with its E2E coverage, messaging polish, generated DOM contract.
- **The other 20% to reach 100%** (T23–T27 + standing rituals + parked
  items): PWA spike, recordings decision+panel, retention/cleanup job,
  platform long tails — plus the standing cadences (quarterly watches,
  monthly erraudit re-measure, per-train runbook) and the consciously
  parked ideas (kept parked WITH named triggers — that is part of the
  100%).

## 2) Comprehensive Plan — 27 tasks, 30–100 min each

Sorted by importance / impact / effort / customer-value (P0 = now,
P1 = this week, P2 = next trains, P3 = on demand).

| #                                                                                | Task (30–100 min)                                                                                                                                                                                                                                                                                                                                                              | Tier     | Priority | Impact             | Effort   | Customer value             | Depends on                                                   |
| -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | -------- | ------------------ | -------- | -------------------------- | ------------------------------------------------------------ |
| ~~T1~~ done — log T1 CLOSED; stack check green ×2                                | ~~Publish + re-pin + full stack check: push webphone `main` + stack `main`; stack `nix flake lock --update-input webphone`; full stack `nix flake check` on the new lock (the forgotten runbook step)~~                                                                                                                                                                        | ~~1%~~   | ~~P0~~   | ~~enormous~~       | ~~75m~~  | ~~security live~~          | ~~—~~                                                        |
| ~~T2~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                        | ~~pbx-artmann relock (`--update-input telephony`) + prod toplevel pre-build + push (needs stack tree clean + pushed)~~                                                                                                                                                                                                                                                         | ~~1%~~   | ~~P0~~   | ~~high~~           | ~~40m~~  | ~~deploy unblocked~~       | ~~T1~~                                                       |
| ~~T3~~ done — owner deployed (premise corrected; v2.5.0+ followed)               | ~~OWNER prod deploy: `nixos-rebuild test` → smoke `--base https://pbx.artmann.tech` (bogus-creds rejected green, styled 404, version) → `switch`~~                                                                                                                                                                                                                             | ~~1%~~   | ~~P0~~   | ~~critical~~       | ~~30m~~  | ~~THE security fix~~       | ~~T2~~                                                       |
| ~~T4~~ done — erraudit 0 (log); banner E2E green ×2                              | ~~Post-switch T18 verification (error-excellence plan): rejection-banner E2E on prod (self-send → 40310 reason), enforced erraudit green on deployed tree~~                                                                                                                                                                                                                    | ~~1%~~   | ~~P0~~   | ~~high~~           | ~~30m~~  | ~~trust~~                  | ~~T3~~                                                       |
| ~~T5~~ done — SMS lane restored; messaging FULLY_FUNCTIONAL                      | ~~OWNER SMS-bridge root cause: `journalctl -u telnyx-webhooks` grep on prod, restore the outbound SMS lane per findings~~                                                                                                                                                                                                                                                      | ~~1%~~   | ~~P0~~   | ~~high~~           | ~~30m~~  | ~~SMS works~~              | ~~T3 (same console session)~~                                |
| ~~T6~~ done — v2.5.0 folded + gated (tag 25740c6)                                | ~~v2.5.0 fold + gates: CHANGELOG/FEATURES/TODO_LIST fold, `webphoneVersion` bump, buildflow no-cache + go test + flake check + smoke~~                                                                                                                                                                                                                                         | ~~4%~~   | ~~P1~~   | ~~high~~           | ~~90m~~  | ~~versioned fix set~~      | ~~T3 (owner train-cut call in T11; can pre-stage after T1)~~ |
| ~~T7~~ done — release EXIT=0; gh object live                                     | ~~v2.5.0 release run + closeout: `release.sh` (stack gates incl. E2E, aarch64 ELF guard, vulnix, tag+push, lychee, gh object) + closing sweep~~                                                                                                                                                                                                                                | ~~4%~~   | ~~P1~~   | ~~high~~           | ~~60m~~  | ~~public artifact~~        | ~~T6~~                                                       |
| ~~T8~~ done — forced runs 384s/373s; 445s budget in ROADMAP                      | ~~E2E wall-time re-baseline: 2 forced-rebuild runs on a quiet machine, set the budget in the watch row (+ ROADMAP note)~~                                                                                                                                                                                                                                                      | ~~4%~~   | ~~P1~~   | ~~medium~~         | ~~45m~~  | ~~train speed~~            | ~~T7 (or any quiet window)~~                                 |
| ~~T9~~ done — routed — drafts unposted; TODO_LIST announcements row owns it      | ~~Announcements: owner picks channels + disclosure posture; extend drafts to v2.4.0/v2.5.0; post; link release objects~~                                                                                                                                                                                                                                                       | ~~4%~~   | ~~P1~~   | ~~medium~~         | ~~30m~~  | ~~public trust~~           | ~~T7, T10 (posture)~~                                        |
| ~~T10~~ done — pusher fixed; runbook discipline lines (log)                      | ~~Pusher daemon + commit discipline: diagnose why BOTH repos' pushes stopped 2026-09-22 (11h+ outage), fix/restart, verify a push lands; enforce narrative-commit-at-phase-boundary as a runbook checklist item~~                                                                                                                                                              | ~~4%~~   | ~~P1~~   | ~~high~~           | ~~45m~~  | ~~infra trust~~            | ~~— (independent)~~                                          |
| ~~T11~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST) | ~~OWNER-calls batch session (~14 decisions, one sitting): train-cut/cadence, `/livez` consumer, HSTS, XFF sanitization flip, disclosure posture, 18-49 annotate scope, loopback `delivered` semantics, handler dual-layer keep/drop, gh-release-object habit, Go module v2 policy, recordings product intent, TEMP-DIAG keep, oops ratification, `backup.retentionDays` want~~ | ~~4%~~   | ~~P1~~   | ~~high~~           | ~~90m~~  | ~~unblocks T7/T9/T24/T25~~ | ~~—~~                                                        |
| ~~T12~~ done — limiter + cap + OpenAPI shipped (AGENTS)                          | ~~Contacts API completion: rate limiter on `POST /api/contacts`, OpenAPI 3.1 surface, count-cap/pagination decision (+ tests each)~~                                                                                                                                                                                                                                           | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~90m~~  | ~~robustness~~             | ~~—~~                                                        |
| ~~T13~~ done — contacts SSE nudge shipped (AGENTS events list)                   | ~~SSE `contacts` nudge: event + island panel re-fetch (mirror the `voicemail` payload-less nudge), kills the last stale live surface; E2E-ready marker~~                                                                                                                                                                                                                       | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~60m~~  | ~~live UI~~                | ~~—~~                                                        |
| ~~T14~~ done — retry ladder shipped; rotation NOT-DO archived                    | ~~Session hardening pair: island CSRF-adoption retry ×3 with backoff before the reload fallback; CSRF rotation on session TTL refresh (spec verdict first — ROADMAP raw idea, P23 deferral)~~                                                                                                                                                                                  | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~90m~~  | ~~auth robustness~~        | ~~—~~                                                        |
| ~~T15~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                | ~~Gate/tooling tail: drift guard wired into buildflow (00:14 f10), `GOEXPERIMENT=jsonv2` deliberate removal sweep, release-script auto-lychee step~~                                                                                                                                                                                                                           | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~90m~~  | ~~train health~~           | ~~—~~                                                        |
| ~~T16~~ done — docs-health sweeps: AGENTS at cap; reports archived               | ~~Docs kernel: AGENTS.md size pass (657→<400, move detail to docs/), HARVEST routing of the annotated reports' still-open f-items into TODO_LIST/ROADMAP, 18-49 review ANNOTATE (owner-gated), 19-37 plan checkbox hygiene (7 open), FEATURES VERIFY pass~~                                                                                                                    | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~100m~~ | ~~knowledge lasts~~        | ~~T11 (scope calls)~~                                        |
| ~~T17~~ done — scenarios 10+11; smoke --expect-version; wait_marker              | ~~Island test/smoke tail: re-entrancy-guard test for `rebuildConnection`, `wait_marker` marker-alternatives (drop the `NOTIF-` substring hack), transient "rebuilding…" pill state, smoke `--expect-version` flag~~                                                                                                                                                            | ~~20%~~  | ~~P2~~   | ~~low-med~~        | ~~80m~~  | ~~diagnostics~~            | ~~—~~                                                        |
| ~~T18~~ done — retentionDays + README pointer shipped; notify NOT-DO             | ~~Backup + startup contracts: `backup.retentionDays` (if T11 wants it), README off-machine restic/borg pointer, `/startupz`→systemd `Type=notify` contract doc (+ optional wiring)~~                                                                                                                                                                                           | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~60m~~  | ~~data durability~~        | ~~T11~~                                                      |
| ~~T19~~ done — go-health quick-start + 12 release objects (log)                  | ~~Upstream tails: go-health `doc.go` `NewChecks` quick-start, cqrs-htmx v4.11.0 family gh release-object audit + the `NamedCheck.Timeout` retro-note in the next train's CHANGELOG~~                                                                                                                                                                                           | ~~20%~~  | ~~P2~~   | ~~low~~            | ~~60m~~  | ~~ecosystem~~              | ~~—~~                                                        |
| ~~T20~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins         | ~~UX/a11y pass + its E2E coverage: badge ringing/established states + `aria-live`, migration completion toast, thread-LIST `data-dial`, `data-sms` affordance, history ☆ save, logged-out-dial E2E scenario, contacts round-trip E2E, live-badge E2E assertion~~                                                                                                               | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~100m~~ | ~~felt polish~~            | ~~T13 (contacts nudge helps)~~                               |
| ~~T21~~ done — M17 trust batch; delivered badge; thumbnails shipped              | ~~Messaging polish batch: persist + display failure reasons (parity with fax), distinct delivered badge, image thumbnails, draft persistence (each behind its own test)~~                                                                                                                                                                                                      | ~~20%~~  | ~~P2~~   | ~~medium~~         | ~~100m~~ | ~~felt polish~~            | ~~—~~                                                        |
| ~~T22~~ done — docs/dom-contract.md generated; AGENTS links it                   | ~~Generated island DOM-contract file: emit the 35-id contract from `TestServedPageHoldsTheDomContract` output; AGENTS links it instead of hand-listing (stops enumeration drift)~~                                                                                                                                                                                             | ~~20%~~  | ~~P2~~   | ~~low~~            | ~~40m~~  | ~~drift-proofing~~         | ~~—~~                                                        |
| ~~T23~~ done — PWA verdict archived; PWA-LITE shipped 2026-10-23                 | ~~PWA evaluation spike (manifest + SW caching rules vs CSP/`/events` no-cache + hash-pinned island assets) — verdict doc, build only on a green verdict~~                                                                                                                                                                                                                      | ~~rest~~ | ~~P3~~   | ~~unknown~~        | ~~60m~~  | ~~install-ability~~        | ~~T11 (product call)~~                                       |
| ~~T24~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea      | ~~Recordings: product-intent decision (consent/jurisdiction/access model — T11) then the panel (playback via the phone-api proxy pattern, REC indicator, recorded badge on CDR rows, retention surfacing)~~                                                                                                                                                                    | ~~rest~~ | ~~P3~~   | ~~high-if-wanted~~ | ~~100m~~ | ~~major feature~~          | ~~T11~~                                                      |
| ~~T25~~ done — internal/retention sweeper shipped (ROADMAP)                      | ~~Retention/cleanup job: `retention_days` setting + module timer for old CDR rows, read fax/voicemail blobs, expired sessions (bounded deletion, drill-verified)~~                                                                                                                                                                                                             | ~~rest~~ | ~~P3~~   | ~~medium~~         | ~~90m~~  | ~~data hygiene~~           | ~~T18~~                                                      |
| ~~T26~~ done — /metrics, turn_rest, /api/export shipped                          | ~~Platform long-tail batch 1: metrics endpoint (Prometheus text), short-lived TURN REST creds via `/config.js`, per-extension data export, timezone-aware timestamps, MIME sniffing on attachments~~                                                                                                                                                                           | ~~rest~~ | ~~P3~~   | ~~low-med~~        | ~~100m~~ | ~~ops polish~~             | ~~—~~                                                        |
| ~~T27~~ done — gzip, favicon, signed tags, i18n guard, verdict records shipped   | ~~Platform long-tail batch 2: nginx gzip module option, `/favicon.ico` route, signed tags (`git tag -s`), i18n dynamic-template key-sync guard, webhook-idempotency durability decision (post-restart replay acceptable?), OpenAPI boundary decision record, limiter-key widening runbook line~~                                                                               | ~~rest~~ | ~~P3~~   | ~~low~~            | ~~100m~~ | ~~long-tail hygiene~~      | ~~—~~                                                        |

**Everything else — standing rituals & consciously parked (part of the
100%):**

- Standing cadences: quarterly watches (due 2026-12-20; named triggers
  earlier), monthly erraudit tiers-1+2 re-measure (the count must
  shrink, never grow), per-train runbook (incl. aarch64 cross-builds of
  the checks you care about, closing sweep with process-death proof).
- Non-schedulable standing rows: "analyze the next E2E flake with the
  shipped dumps" (fires on occurrence); "confirm authorship of the
  json/v2 test change + diff-review `63f8b3c`" (curiosity-class, only if
  the files are touched again).
- Consciously PARKED with named triggers (do not schedule): local
  Playwright island E2E (trigger: stack E2E too slow for inner loops /
  a toast regression escapes), OOB badge push (verdict criteria), video
  calls (on demand), voicemail transcripts (if the PBX API ships them),
  dashboard-HTML (blocked on CSP stance), server-side telephony
  (REJECTED — research table), union coverage (blocked on BuildFlow
  upstream), structured `/api/session` errors + ClientIP-trust upstream
  notes (conditional).

## 3) Fine Breakdown — every task into ≤12-minute units

Sorted by parent-task priority; each unit independently verifiable.
(Effort is WORK time; long builds/flake checks are watch time — poll,
never block on them.)

| Sub                                                                                                   | Unit (≤12 min)                                                                                                                          | Effort  | Verify by                |
| ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------------------ |
| ~~1a~~ done — log T1 CLOSED; stack check green ×2                                                     | ~~`git status` + `ls-remote` both repos; confirm clean trees and what is unpushed~~                                                     | ~~3m~~  | ~~output read~~          |
| ~~1b~~ done — log T1 CLOSED; stack check green ×2                                                     | ~~Push webphone `main`; verify `ls-remote == HEAD`~~                                                                                    | ~~5m~~  | ~~hash match~~           |
| ~~1c~~ done — log T1 CLOSED; stack check green ×2                                                     | ~~Push stack `main`; verify~~                                                                                                           | ~~5m~~  | ~~hash match~~           |
| ~~1d~~ done — log T1 CLOSED; stack check green ×2                                                     | ~~Stack: `nix flake lock --update-input webphone`; grep the locked `rev`~~                                                              | ~~10m~~ | ~~rev == webphone HEAD~~ |
| ~~1e-i~~ done — log T1 CLOSED; stack check green ×2                                                   | ~~Launch full stack `nix flake check` in background; note shell id~~                                                                    | ~~2m~~  | ~~shell id~~             |
| ~~1e-ii~~ done — log T1 CLOSED; stack check green ×2                                                  | ~~Poll to completion; read the verdict line from the log (never assume)~~                                                               | ~~10m~~ | ~~"all checks passed"~~  |
| ~~1f~~ done — log T1 CLOSED; stack check green ×2                                                     | ~~Record chain state in TODO_LIST redeploy row~~                                                                                        | ~~5m~~  | ~~row updated~~          |
| ~~2a~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                             | ~~Verify stack tree clean + pushed before the path: relock~~                                                                            | ~~5m~~  | ~~git status~~           |
| ~~2b~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                             | ~~pbx-artmann: swap pinned rev + `nix flake update telephony` + drift probe~~                                                           | ~~10m~~ | ~~lock diff + probe OK~~ |
| ~~2c-i~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                           | ~~Launch `nix build .#pbx-toplevel` (x86_64) in background~~                                                                            | ~~2m~~  | ~~build log~~            |
| ~~2c-ii~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                          | ~~Launch aarch64 toplevel cross-build in background~~                                                                                   | ~~2m~~  | ~~build log~~            |
| ~~2c-iii~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                         | ~~Watch both EXITs; verify webphone store path MOVED vs old closure~~                                                                   | ~~10m~~ | ~~EXIT=0 + path moved~~  |
| ~~2d~~ done — log T2 CLOSED (relock 91ac2c82, both archs)                                             | ~~Commit + push pbx-artmann `master` (narrative message)~~                                                                              | ~~10m~~ | ~~ls-remote~~            |
| ~~3a~~ done — owner deployed (premise corrected; v2.5.0+ followed)                                    | ~~OWNER: `nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech`~~                                                       | ~~12m~~ | ~~unit active~~          |
| ~~3b~~ done — owner deployed (premise corrected; v2.5.0+ followed)                                    | ~~Probe: `webphone-smoke.py --base https://pbx.artmann.tech` — bogus-creds rejected green, styled 404, `/version` shows the new build~~ | ~~10m~~ | ~~0 failed~~             |
| ~~3c~~ done — owner deployed (premise corrected; v2.5.0+ followed)                                    | ~~OWNER: `nixos-rebuild switch`; confirm service healthy (`/healthz` triple)~~                                                          | ~~8m~~  | ~~200 ×3~~               |
| ~~4a~~ done — erraudit 0; banner E2E green (runs 6/7)                                                 | ~~Rejection-banner E2E on prod: self-send shows the 40310 reason banner~~                                                               | ~~12m~~ | ~~banner seen~~          |
| ~~4b~~ done — erraudit 0; banner E2E green (runs 6/7)                                                 | ~~Enforced erraudit set green on the deployed tree~~                                                                                    | ~~10m~~ | ~~EXIT=0~~               |
| ~~4c~~ done — erraudit 0; banner E2E green (runs 6/7)                                                 | ~~Record T11/T18 completion in the error-excellence plan doc~~                                                                          | ~~5m~~  | ~~EXECUTED marker~~      |
| ~~5a~~ done — SMS lane restored; messaging FULLY_FUNCTIONAL                                           | ~~OWNER: `journalctl -u telnyx-webhooks --since today, grep -iE "sms 422 error"` paste~~                                                | ~~10m~~ | ~~tail in hand~~         |
| ~~5b~~ done — SMS lane restored; messaging FULLY_FUNCTIONAL                                           | ~~Restore lane per findings (bridge restart / creds / Telnyx console) + send a test SMS~~                                               | ~~12m~~ | ~~delivered~~            |
| ~~5c~~ done — SMS lane restored; messaging FULLY_FUNCTIONAL                                           | ~~Record root cause in TODO_LIST (row deletion) + stack runbook~~                                                                       | ~~5m~~  | ~~row gone~~             |
| ~~6a-i~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                              | ~~Fold CHANGELOG `Unreleased` → dated `2.5.0` section~~                                                                                 | ~~10m~~ | ~~heading check~~        |
| ~~6a-ii~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                             | ~~Sync FEATURES rows to the fold~~                                                                                                      | ~~8m~~  | ~~rows updated~~         |
| ~~6a-iii~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                            | ~~Sync TODO_LIST/ROADMAP (delete done rows, add new pointer rows)~~                                                                     | ~~7m~~  | ~~rows updated~~         |
| ~~6b~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                                | ~~Bump `webphoneVersion` (one let-binding, package + `/version` ldflags)~~                                                              | ~~10m~~ | ~~`/version` probe~~     |
| ~~6c-i~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                              | ~~Gates: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` (read the verdict)~~                                                                   | ~~12m~~ | ~~EXIT=0~~               |
| ~~6c-ii~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                             | ~~Gates: `go test -count=1 ./...` via nix develop~~                                                                                     | ~~8m~~  | ~~EXIT=0~~               |
| ~~6d-i~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                              | ~~Gates: `nix flake check`~~                                                                                                            | ~~10m~~ | ~~green~~                |
| ~~6d-ii~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                             | ~~Gates: `webphone-smoke.py` (38+4 checks)~~                                                                                            | ~~10m~~ | ~~0 failed~~             |
| ~~6e~~ done — v2.5.0 folded + gated (log, tag 25740c6)                                                | ~~Explicit narrative commit of the fold~~                                                                                               | ~~10m~~ | ~~git log~~              |
| ~~7a~~ done — release.sh EXIT=0; gh object live (log)                                                 | ~~Run `scripts/release.sh` (watch: stack E2E in-train, aarch64 `b700` guard, vulnix triage)~~                                           | ~~12m~~ | ~~step logs~~            |
| ~~7b~~ done — release.sh EXIT=0; gh object live (log)                                                 | ~~Verify tag + gh release object + lychee link check~~                                                                                  | ~~12m~~ | ~~URLs live~~            |
| ~~7c~~ done — release.sh EXIT=0; gh object live (log)                                                 | ~~Post-train stack re-pin (runbook step 6) + stack gates green~~                                                                        | ~~12m~~ | ~~lock rev~~             |
| ~~7d~~ done — release.sh EXIT=0; gh object live (log)                                                 | ~~Closing sweep: `pgrep` proof, `git ls-remote` ×3, daemon-pushed confirmation~~                                                        | ~~12m~~ | ~~checklist~~            |
| ~~8a-i~~ done — forced runs 384s/373s; 445s budget in ROADMAP                                         | ~~Forced-rebuild E2E run 1 on a quiet machine; record wall-time~~                                                                       | ~~12m~~ | ~~run EXIT=0~~           |
| ~~8a-ii~~ done — forced runs 384s/373s; 445s budget in ROADMAP                                        | ~~Forced-rebuild E2E run 2; record wall-time~~                                                                                          | ~~12m~~ | ~~run EXIT=0~~           |
| ~~8b~~ done — forced runs 384s/373s; 445s budget in ROADMAP                                           | ~~Set the budget (max observed + ~15%) in TODO_LIST watch row + ROADMAP watches section~~                                               | ~~10m~~ | ~~rows updated~~         |
| ~~8c~~ done — forced runs 384s/373s; 445s budget in ROADMAP                                           | ~~Note the 151s→new history (scenario growth) in the same row~~                                                                         | ~~5m~~  | ~~row text~~             |
| ~~9a~~ done — routed — drafts unposted; TODO_LIST announcements row owns it                           | ~~Post announcements per T10 posture; add discussion links to release objects~~                                                         | ~~12m~~ | ~~links live~~           |
| ~~10a~~ done — pusher fixed; runbook discipline lines (log)                                           | ~~Diagnose pusher: identify the committing actor + why pushes stopped~~                                                                 | ~~12m~~ | ~~root cause named~~     |
| ~~10b~~ done — pusher fixed; runbook discipline lines (log)                                           | ~~Fix/restart the pusher; land a test push (any repo) and verify origin moves~~                                                         | ~~12m~~ | ~~ls-remote~~            |
| ~~10c~~ done — pusher fixed; runbook discipline lines (log)                                           | ~~Add "explicit narrative commit at phase boundaries" to the runbook checklist section~~                                                | ~~10m~~ | ~~AGENTS edit~~          |
| ~~10d~~ done — pusher fixed; runbook discipline lines (log)                                           | ~~Decide + record the `docs/status` daemon-exclusion ask (upstream infra) as a ROADMAP line~~                                           | ~~10m~~ | ~~row added~~            |
| ~~11a~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Owner decisions 1–3: train-cut/cadence NOW, `/livez` consumer, HSTS on prod~~                                                         | ~~12m~~ | ~~AGENTS DECIDED lines~~ |
| ~~11b~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Owner decisions 4–6: XFF sanitization stance, disclosure posture, 18-49 annotate scope~~                                              | ~~12m~~ | ~~rows updated~~         |
| ~~11c~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Owner decisions 7–9: loopback `delivered`, handler dual-layer, gh-release habit~~                                                     | ~~12m~~ | ~~recorded~~             |
| ~~11d~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Owner decisions 10–12: module v2 policy, recordings intent, TEMP-DIAG keep~~                                                          | ~~12m~~ | ~~recorded~~             |
| ~~11e~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Owner decisions 13–14: oops ratification, `backup.retentionDays` want~~                                                               | ~~10m~~ | ~~TODO row closed~~      |
| ~~11f~~ done — routed — sitting owed; briefing + fresh packet own it (TODO_LIST)                      | ~~Write every verdict into its owning doc (AGENTS/ROADMAP/TODO_LIST) in one pass~~                                                      | ~~12m~~ | ~~docs-health check~~    |
| ~~12a-i~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                             | ~~Wire `KeyedRateLimiter` on `POST /api/contacts` (mirror login wiring)~~                                                               | ~~10m~~ | ~~code reads~~           |
| ~~12a-ii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                            | ~~429 behavior test (burst over limit → limited; other extensions unaffected)~~                                                         | ~~10m~~ | ~~429 test green~~       |
| ~~12a-iii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                           | ~~Full go-test gate~~                                                                                                                   | ~~5m~~  | ~~EXIT=0~~               |
| ~~12b-i~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                             | ~~Author OpenAPI 3.1 schema for `/api/contacts` (mirror the session spec)~~                                                             | ~~10m~~ | ~~spec served~~          |
| ~~12b-ii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                            | ~~Contract test: served spec matches the implementation~~                                                                               | ~~10m~~ | ~~spec test green~~      |
| ~~12b-iii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                           | ~~Gate run~~                                                                                                                            | ~~5m~~  | ~~EXIT=0~~               |
| ~~12c-i~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                             | ~~Decision: count-cap vs pagination (record rationale)~~                                                                                | ~~5m~~  | ~~DECIDED~~              |
| ~~12c-ii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                            | ~~Implement the chosen bound in store + handler~~                                                                                       | ~~8m~~  | ~~code reads~~           |
| ~~12c-iii~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                           | ~~Tests: bound enforced at the boundary row~~                                                                                           | ~~10m~~ | ~~tests green~~          |
| ~~12c-iv~~ done — limiter + cap + OpenAPI shipped (AGENTS contacts bullet)                            | ~~Gate run~~                                                                                                                            | ~~5m~~  | ~~EXIT=0~~               |
| ~~13a-i~~ done — contacts SSE nudge shipped (AGENTS events list)                                      | ~~Server: `contacts` SSE event (payload-less nudge, ExtensionHub fan-out) + unit test~~                                                 | ~~10m~~ | ~~unit test~~            |
| ~~13a-ii~~ done — contacts SSE nudge shipped (AGENTS events list)                                     | ~~Wire the nudge into contact mutation paths + gate~~                                                                                   | ~~8m~~  | ~~EXIT=0~~               |
| ~~13b-i~~ done — contacts SSE nudge shipped (AGENTS events list)                                      | ~~Island: panel re-fetch on the nudge (mirror voicemail handler)~~                                                                      | ~~8m~~  | ~~island test~~          |
| ~~13b-ii~~ done — contacts SSE nudge shipped (AGENTS events list)                                     | ~~session/panel integration test + node gate~~                                                                                          | ~~7m~~  | ~~node --test green~~    |
| ~~13c-i~~ done — contacts SSE nudge shipped (AGENTS events list)                                      | ~~Morph-surface id check (contacts panel rows carry stable ids)~~                                                                       | ~~8m~~  | ~~id test~~              |
| ~~13c-ii~~ done — contacts SSE nudge shipped (AGENTS events list)                                     | ~~Add the E2E-ready marker constant~~                                                                                                   | ~~5m~~  | ~~marker grep~~          |
| ~~14a-i~~ done — retry ladder shipped; rotation NOT-DO archived                                       | ~~Adoption retry ×3 with backoff before the reload fallback (session.js)~~                                                              | ~~10m~~ | ~~island test~~          |
| ~~14a-ii~~ done — retry ladder shipped; rotation NOT-DO archived                                      | ~~session.test cases: fail ×3 → reload; success on retry 2~~                                                                            | ~~10m~~ | ~~tests green~~          |
| ~~14a-iii~~ done — retry ladder shipped; rotation NOT-DO archived                                     | ~~node gate~~                                                                                                                           | ~~4m~~  | ~~EXIT=0~~               |
| ~~14b-i~~ done — retry ladder shipped; rotation NOT-DO archived                                       | ~~CSRF-rotation-on-TTL-refresh threat-model notes (vs adoption cost; P23 context)~~                                                     | ~~10m~~ | ~~notes written~~        |
| ~~14b-ii~~ done — retry ladder shipped; rotation NOT-DO archived                                      | ~~Verdict doc written + TODO_LIST pointer row~~                                                                                         | ~~10m~~ | ~~verdict doc~~          |
| ~~14c-i~~ done — retry ladder shipped; rotation NOT-DO archived                                       | ~~Per verdict: rotate token on session slide (server)~~                                                                                 | ~~10m~~ | ~~unit test~~            |
| ~~14c-ii~~ done — retry ladder shipped; rotation NOT-DO archived                                      | ~~Per verdict: island adopts rotated token without reload~~                                                                             | ~~10m~~ | ~~island test~~          |
| ~~14c-iii~~ done — retry ladder shipped; rotation NOT-DO archived                                     | ~~Contract tests + gates~~                                                                                                              | ~~10m~~ | ~~EXIT=0~~               |
| ~~15a-i~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                   | ~~Determine expression path (buildflow step vs documented NOT-DO)~~                                                                     | ~~8m~~  | ~~decision note~~        |
| ~~15a-ii~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                  | ~~Implement the drift guard or write the NOT-DO rationale~~                                                                             | ~~10m~~ | ~~gate runs~~            |
| ~~15b-i~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                   | ~~Sweep devShell/flake/scripts/AGENTS for `GOEXPERIMENT=jsonv2` mentions~~                                                              | ~~8m~~  | ~~list~~                 |
| ~~15b-ii~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                  | ~~Remove deliberate mentions; document the unavoidable ones~~                                                                           | ~~8m~~  | ~~diff read~~            |
| ~~15b-iii~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                 | ~~Full gates after removal (buildflow + flake check)~~                                                                                  | ~~9m~~  | ~~green~~                |
| ~~15c-i~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                   | ~~Add auto-lychee step to release.sh (network-fail policy decided)~~                                                                    | ~~10m~~ | ~~dry-run~~              |
| ~~15c-ii~~ done — drift-gate NOT-DO; jsonv2 sweep; self-heal shipped                                  | ~~Exercise the step on a docs-only edit~~                                                                                               | ~~8m~~  | ~~links live~~           |
| ~~16a-i~~ done — docs-health sweeps: AGENTS at cap; reports archived                                  | ~~Inventory AGENTS sections; mark MOVE vs KEEP per block~~                                                                              | ~~10m~~ | ~~annotated list~~       |
| ~~16a-ii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                 | ~~Extract long blocks → docs/ files (first half)~~                                                                                      | ~~12m~~ | ~~files created~~        |
| ~~16a-iii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                | ~~Extract long blocks → docs/ files (second half)~~                                                                                     | ~~12m~~ | ~~files created~~        |
| ~~16a-iv~~ done — docs-health sweeps: AGENTS at cap; reports archived                                 | ~~Rewrite AGENTS links; verify <400 lines; gates~~                                                                                      | ~~6m~~  | ~~wc -l <400~~           |
| ~~16b-i~~ done — docs-health sweeps: AGENTS at cap; reports archived                                  | ~~Collect still-open f-items from the annotated reports~~                                                                               | ~~8m~~  | ~~list~~                 |
| ~~16b-ii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                 | ~~Route each: TODO_LIST (bounded) vs ROADMAP (vague) vs already-done delete~~                                                           | ~~10m~~ | ~~rows placed~~          |
| ~~16b-iii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                | ~~Docs-health verify pass (one home per fact, no dups)~~                                                                                | ~~7m~~  | ~~no dups~~              |
| ~~16c-i~~ done — docs-health sweeps: AGENTS at cap; reports archived                                  | ~~18-49 review ANNOTATE (if T11 confirms scope) — first half~~                                                                          | ~~10m~~ | ~~inline markers~~       |
| ~~16c-ii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                 | ~~18-49 review ANNOTATE — second half + routing~~                                                                                       | ~~8m~~  | ~~markers + rows~~       |
| ~~16d-i~~ done — docs-health sweeps: AGENTS at cap; reports archived                                  | ~~19-37 plan checkbox hygiene: tick boxes against shipped reality~~                                                                     | ~~10m~~ | ~~boxes ticked~~         |
| ~~16d-ii~~ done — docs-health sweeps: AGENTS at cap; reports archived                                 | ~~FEATURES VERIFY spot-check of the new rows~~                                                                                          | ~~10m~~ | ~~verified~~             |
| ~~17a-i~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                 | ~~Author `rebuildConnection` re-entrancy-guard test (concurrent triggers)~~                                                             | ~~10m~~ | ~~island test~~          |
| ~~17a-ii~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                | ~~Run + verify single rebuild outcome~~                                                                                                 | ~~8m~~  | ~~EXIT=0~~               |
| ~~17b-i~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                 | ~~`wait_marker` explicit alternatives (drop `NOTIF-` substring matching)~~                                                              | ~~8m~~  | ~~E2E eval~~             |
| ~~17b-ii~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                | ~~Dry-eval the E2E with the new markers~~                                                                                               | ~~7m~~  | ~~eval green~~           |
| ~~17c-i~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                 | ~~Transient "rebuilding…" pill state + i18n keys en/de~~                                                                                | ~~10m~~ | ~~island test~~          |
| ~~17c-ii~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                | ~~Island tests for the transient state + node gate~~                                                                                    | ~~10m~~ | ~~EXIT=0~~               |
| ~~17d-i~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                 | ~~Smoke `--expect-version` flag + probe implementation~~                                                                                | ~~10m~~ | ~~smoke run~~            |
| ~~17d-ii~~ done — scenarios 10+11; smoke --expect-version; wait_marker                                | ~~Smoke run with the flag against a fresh binary~~                                                                                      | ~~8m~~  | ~~0 failed~~             |
| ~~18a-i~~ done — backup.retentionDays shipped (module options)                                        | ~~`backup.retentionDays` module option + prune logic in the oneshot~~                                                                   | ~~10m~~ | ~~code reads~~           |
| ~~18a-ii~~ done — backup.retentionDays shipped (module options)                                       | ~~Oneshot test + module flake-check stand-ins~~                                                                                         | ~~10m~~ | ~~module check~~         |
| ~~18b~~ done — README restic/borg pointer                                                             | ~~README off-machine backup pointer (restic/borg) + "destDir is not a backup" note~~                                                    | ~~12m~~ | ~~doc read~~             |
| ~~18c-i~~ done — NOT-DO — unit stays Type=simple (AGENTS.md)                                          | ~~`/startupz`→systemd `Type=notify` contract doc~~                                                                                      | ~~10m~~ | ~~doc~~                  |
| ~~18c-ii~~ done — NOT-DO — unit stays Type=simple (AGENTS.md)                                         | ~~Optional wiring per verdict + module check~~                                                                                          | ~~12m~~ | ~~module check~~         |
| ~~19a-i~~ done — go-health quick-start + 12 release objects (log)                                     | ~~go-health `doc.go` `NewChecks` quick-start snippet + local verify~~                                                                   | ~~10m~~ | ~~snippet tested~~       |
| ~~19a-ii~~ done — go-health quick-start + 12 release objects (log)                                    | ~~Upstream PR + CI green~~                                                                                                              | ~~12m~~ | ~~PR green~~             |
| ~~19b-i~~ done — go-health quick-start + 12 release objects (log)                                     | ~~cqrs-htmx family release-object audit (list missing objects)~~                                                                        | ~~10m~~ | ~~list~~                 |
| ~~19b-ii~~ done — go-health quick-start + 12 release objects (log)                                    | ~~Cut the missing release objects~~                                                                                                     | ~~12m~~ | ~~URLs~~                 |
| ~~19c~~ done — go-health quick-start + 12 release objects (log)                                       | ~~`NamedCheck.Timeout` retro-note in next cqrs-htmx CHANGELOG~~                                                                         | ~~12m~~ | ~~entry~~                |
| ~~20a-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~Badge states (ringing vs established) + `aria-live` announcement~~                                                                    | ~~10m~~ | ~~island tests~~         |
| ~~20a-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~Island tests + node gate~~                                                                                                            | ~~10m~~ | ~~EXIT=0~~               |
| ~~20b-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~Migration completion toast ("imported N contacts") en/de~~                                                                            | ~~8m~~  | ~~island test~~          |
| ~~20b-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~Island test for the toast~~                                                                                                           | ~~7m~~  | ~~EXIT=0~~               |
| ~~20c-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~Thread-LIST `data-dial` (button-outside-anchor restructure)~~                                                                         | ~~12m~~ | ~~markup reads~~         |
| ~~20c-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~Render test + gates~~                                                                                                                 | ~~10m~~ | ~~render test~~          |
| ~~20d-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~`data-sms` affordance (history/voicemail rows → compose prefilled)~~                                                                  | ~~12m~~ | ~~markup reads~~         |
| ~~20d-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~Render/integration tests~~                                                                                                            | ~~12m~~ | ~~render test~~          |
| ~~20e-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~History tab ☆ save-as-contact action~~                                                                                                | ~~12m~~ | ~~action wired~~         |
| ~~20e-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~Render test + gate~~                                                                                                                  | ~~10m~~ | ~~render test~~          |
| ~~20f-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~E2E scenario: logged-out `data-dial` → toast + `#ext` focus~~                                                                         | ~~12m~~ | ~~code~~                 |
| ~~20f-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~E2E run + green ×2~~                                                                                                                  | ~~12m~~ | ~~E2E green ×2~~         |
| ~~20g-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~E2E scenario: island ☆ save → contacts tab shows it~~                                                                                 | ~~12m~~ | ~~code~~                 |
| ~~20g-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~E2E run + green ×2~~                                                                                                                  | ~~12m~~ | ~~E2E green ×2~~         |
| ~~20h-i~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                            | ~~E2E assertion: `#call-badge` during a live call~~                                                                                     | ~~8m~~  | ~~code~~                 |
| ~~20h-ii~~ done — badge/a11y/data-dial/data-sms/star-save shipped; E2E pins                           | ~~E2E run + green ×2~~                                                                                                                  | ~~12m~~ | ~~E2E green ×2~~         |
| ~~21a-i~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                        | ~~Persist message failure reasons server-side~~                                                                                         | ~~10m~~ | ~~store test~~           |
| ~~21a-ii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                       | ~~Display the reason in thread + list surfaces~~                                                                                        | ~~10m~~ | ~~UI check~~             |
| ~~21a-iii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                      | ~~Tests + gate~~                                                                                                                        | ~~8m~~  | ~~EXIT=0~~               |
| ~~21b-i~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                        | ~~Distinct delivered badge style (messages list)~~                                                                                      | ~~10m~~ | ~~CSS + markup~~         |
| ~~21b-ii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                       | ~~Render test~~                                                                                                                         | ~~8m~~  | ~~render test~~          |
| ~~21c-i~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                        | ~~Decide server-side vs CSS thumbnails~~                                                                                                | ~~5m~~  | ~~DECIDED~~              |
| ~~21c-ii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                       | ~~Implement image thumbnails in threads~~                                                                                               | ~~12m~~ | ~~UI check~~             |
| ~~21c-iii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                      | ~~Thumbnail test~~                                                                                                                      | ~~10m~~ | ~~test green~~           |
| ~~21d-i~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                        | ~~Draft persistence per thread (localStorage)~~                                                                                         | ~~10m~~ | ~~island test~~          |
| ~~21d-ii~~ done — M17 trust batch; delivered badge; thumbnails FULLY_FUNCTIONAL                       | ~~Island test + node gate~~                                                                                                             | ~~10m~~ | ~~EXIT=0~~               |
| ~~22a-i~~ done — docs/dom-contract.md generated; AGENTS links it                                      | ~~Golden-write mechanism + on-mismatch policy decision~~                                                                                | ~~12m~~ | ~~policy note~~          |
| ~~22a-ii~~ done — docs/dom-contract.md generated; AGENTS links it                                     | ~~Test emits the contract file; diff reviewed~~                                                                                         | ~~10m~~ | ~~file generated~~       |
| ~~22b~~ done — docs/dom-contract.md generated; AGENTS links it                                        | ~~AGENTS/stack-doc switch to linking the generated contract~~                                                                           | ~~12m~~ | ~~link live~~            |
| ~~22c~~ done — docs/dom-contract.md generated; AGENTS links it                                        | ~~Stack E2E operator note: contract file = greppable cross-check source~~                                                               | ~~10m~~ | ~~doc note~~             |
| ~~23a-i~~ done — PWA verdict archived; PWA-LITE shipped 2026-10-23                                    | ~~PWA spike: manifest + SW rules vs CSP/`/events`/hash-pinned assets~~                                                                  | ~~12m~~ | ~~inventory~~            |
| ~~23a-ii~~ done — PWA verdict archived; PWA-LITE shipped 2026-10-23                                   | ~~Verdict doc written~~                                                                                                                 | ~~10m~~ | ~~verdict doc~~          |
| ~~23b-i~~ done — PWA verdict archived; PWA-LITE shipped 2026-10-23                                    | ~~(only on green verdict) minimal SW + manifest~~                                                                                       | ~~12m~~ | ~~smoke~~                |
| ~~23b-ii~~ done — PWA verdict archived; PWA-LITE shipped 2026-10-23                                   | ~~Install smoke verification~~                                                                                                          | ~~10m~~ | ~~smoke~~                |
| ~~24a~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing     | ~~Recordings product-intent verdict (from T11) → decision doc~~                                                                         | ~~12m~~ | ~~DECIDED~~              |
| ~~24b-i~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing   | ~~Recordings panel: phone-api proxy route + tests~~                                                                                     | ~~12m~~ | ~~route test~~           |
| ~~24b-ii~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing  | ~~Playback UI panel~~                                                                                                                   | ~~12m~~ | ~~panel test~~           |
| ~~24b-iii~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing | ~~Integration tests + gate~~                                                                                                            | ~~12m~~ | ~~panel test~~           |
| ~~24c-i~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing   | ~~REC indicator (island)~~                                                                                                              | ~~10m~~ | ~~island test~~          |
| ~~24c-ii~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing  | ~~Recorded badge on CDR rows~~                                                                                                          | ~~10m~~ | ~~render test~~          |
| ~~24c-iii~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing | ~~Tests + gate~~                                                                                                                        | ~~8m~~  | ~~EXIT=0~~               |
| ~~24d-i~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing   | ~~Access model + `*97` opt-out visibility~~                                                                                             | ~~8m~~  | ~~docs~~                 |
| ~~24d-ii~~ done — routed — recordings stay owner-intent; ROADMAP owns the idea; repo records nothing  | ~~README consent note~~                                                                                                                 | ~~10m~~ | ~~docs~~                 |
| ~~25a-i~~ done — internal/retention sweeper shipped (ROADMAP)                                         | ~~`retention_days` config plumb~~                                                                                                       | ~~8m~~  | ~~config test~~          |
| ~~25a-ii~~ done — internal/retention sweeper shipped (ROADMAP)                                        | ~~Store deletion queries~~                                                                                                              | ~~12m~~ | ~~store tests~~          |
| ~~25a-iii~~ done — internal/retention sweeper shipped (ROADMAP)                                       | ~~Store tests (aged rows deleted, recent kept)~~                                                                                        | ~~10m~~ | ~~store tests~~          |
| ~~25b-i~~ done — internal/retention sweeper shipped (ROADMAP)                                         | ~~Module systemd timer definition~~                                                                                                     | ~~10m~~ | ~~module check~~         |
| ~~25b-ii~~ done — internal/retention sweeper shipped (ROADMAP)                                        | ~~Flake-check stand-ins for the new timer~~                                                                                             | ~~10m~~ | ~~module check~~         |
| ~~25c-i~~ done — internal/retention sweeper shipped (ROADMAP)                                         | ~~Drill: create aged fixtures~~                                                                                                         | ~~10m~~ | ~~fixtures exist~~       |
| ~~25c-ii~~ done — internal/retention sweeper shipped (ROADMAP)                                        | ~~Run + verify bounded deletion~~                                                                                                       | ~~12m~~ | ~~drill green~~          |
| ~~26a-i~~ done — /metrics, turn_rest, /api/export shipped                                             | ~~Metrics endpoint skeleton + metric types~~                                                                                            | ~~10m~~ | ~~code reads~~           |
| ~~26a-ii~~ done — /metrics, turn_rest, /api/export shipped                                            | ~~Wire counts/durations from existing log data~~                                                                                        | ~~12m~~ | ~~code reads~~           |
| ~~26a-iii~~ done — /metrics, turn_rest, /api/export shipped                                           | ~~Probe test~~                                                                                                                          | ~~8m~~  | ~~probe test~~           |
| ~~26b-i~~ done — /metrics, turn_rest, /api/export shipped                                             | ~~Short-lived TURN REST creds generation + `/config.js` key~~                                                                           | ~~12m~~ | ~~config test~~          |
| ~~26b-ii~~ done — /metrics, turn_rest, /api/export shipped                                            | ~~Config test + gate~~                                                                                                                  | ~~10m~~ | ~~config test~~          |
| ~~26c-i~~ done — /metrics, turn_rest, /api/export shipped                                             | ~~Per-extension export route + zip assembly~~                                                                                           | ~~12m~~ | ~~route test~~           |
| ~~26c-ii~~ done — /metrics, turn_rest, /api/export shipped                                            | ~~Threads/faxes/contacts/vCard inclusion~~                                                                                              | ~~12m~~ | ~~export test~~          |
| ~~26c-iii~~ done — /metrics, turn_rest, /api/export shipped                                           | ~~Export test~~                                                                                                                         | ~~12m~~ | ~~export test~~          |
| ~~26d-i~~ done — /metrics, turn_rest, /api/export shipped                                             | ~~Timezone-aware timestamps (settings choice)~~                                                                                         | ~~10m~~ | ~~settings~~             |
| ~~26d-ii~~ done — /metrics, turn_rest, /api/export shipped                                            | ~~Render tests~~                                                                                                                        | ~~10m~~ | ~~render test~~          |
| ~~26e-i~~ done — /metrics, turn_rest, /api/export shipped                                             | ~~MIME sniffing helper + correct Content-Type serve~~                                                                                   | ~~10m~~ | ~~header test~~          |
| ~~26e-ii~~ done — /metrics, turn_rest, /api/export shipped                                            | ~~Header test + gate~~                                                                                                                  | ~~8m~~  | ~~header test~~          |
| ~~27a-i~~ done — nginx.gzip.enable module option shipped (ROADMAP)                                    | ~~nginx gzip module option + vhost wiring~~                                                                                             | ~~10m~~ | ~~module check~~         |
| ~~27a-ii~~ done — nginx.gzip.enable module option shipped (ROADMAP)                                   | ~~Module check~~                                                                                                                        | ~~8m~~  | ~~module check~~         |
| ~~27b~~ done — favicon route shipped (internal/server/assets.go)                                      | ~~`/favicon.ico` route + test~~                                                                                                         | ~~12m~~ | ~~probe test~~           |
| ~~27c-i~~ done — git tag -s in release.sh                                                             | ~~Signed tags (`git tag -s`) config in release.sh~~                                                                                     | ~~8m~~  | ~~tag verify~~           |
| ~~27c-ii~~ done — git tag -s in release.sh                                                            | ~~Tag sign dry-run verify~~                                                                                                             | ~~7m~~  | ~~signature ok~~         |
| ~~27d-i~~ done — i18n parity + unused-key tests (views/i18n_test.go)                                  | ~~i18n dynamic-template key-sync guard (enumerate every `t("…")` key)~~                                                                 | ~~12m~~ | ~~guard test~~           |
| ~~27d-ii~~ done — i18n parity + unused-key tests (views/i18n_test.go)                                 | ~~Guard test run~~                                                                                                                      | ~~10m~~ | ~~guard test~~           |
| ~~27e~~ done — webhook-idem verdict archived; in-memory 1h shipped                                    | ~~Webhook idempotency durability decision record (post-restart replay: acceptable?)~~                                                   | ~~12m~~ | ~~DECIDED~~              |
| ~~27f~~ done — OpenAPI boundary verdict archived                                                      | ~~OpenAPI boundary decision record (extend beyond session/contacts or stop)~~                                                           | ~~12m~~ | ~~DECIDED~~              |
| ~~27g~~ done — AGENTS limiter-key note (remoteHostKey; XFF stance)                                    | ~~Limiter-key widening runbook line (NAT offices)~~                                                                                     | ~~10m~~ | ~~runbook~~              |

## 4) Execution Graph (mermaid)

```mermaid
flowchart TD
    subgraph TIER1["1% → 51% — SHIP THE FIX CHAIN (P0, one afternoon)"]
        T1[T1 publish both repos + stack re-pin + full stack flake check]
        T2[T2 pbx-artmann relock + toplevel]
        T3[T3 OWNER deploy + prod probes]
        T4[T4 T18 post-switch verification]
        T5[T5 OWNER SMS-bridge root cause]
        T1-->T2-->T3-->T4
        T3-->T5
    end
    subgraph TIER4["4% → 64% — THE DEPLOY/RELEASE MACHINE"]
        T10[T10 fix pusher + narrative-commit discipline]
        T11[T11 OWNER-calls batch: ~14 decisions]
        T6[T6 v2.5.0 fold + gates]
        T7[T7 v2.5.0 release run + closeout]
        T8[T8 E2E budget re-baseline]
        T9[T9 announcements posted]
        T10-->T1
        T11-->T6
        T6-->T7-->T8
        T7-->T9
        T11-->T9
    end
    subgraph TIER20["20% → 80% — DURABILITY KERNEL + VISIBLE PRODUCT"]
        T12[T12 contacts API completion]
        T13[T13 SSE contacts nudge]
        T14[T14 adoption retry + CSRF rotation]
        T15[T15 gate/tooling tail]
        T16[T16 docs kernel]
        T17[T17 island test/smoke tail]
        T18b[T18 backup + startup contracts]
        T19[T19 upstream tails]
        T20[T20 UX/a11y pass + E2E coverage]
        T21[T21 messaging polish]
        T22[T22 generated DOM contract]
        T13-->T20
        T11-->T16
        T11-->T18b
    end
    subgraph REST["other 20% → 100%"]
        T23[T23 PWA spike]
        T24[T24 recordings decision + panel]
        T25[T25 retention/cleanup job]
        T26[T26 platform tail 1]
        T27[T27 platform tail 2]
        T11-->T23
        T11-->T24
        T18b-->T25
    end
    subgraph RITUAL["standing rituals — always on"]
        R1[quarterly watches<br/>next 2026-12-20]
        R2[monthly erraudit<br/>tiers 1+2 re-measure]
        R3[per-train runbook<br/>E2E x2, aarch64, vulnix, closing sweep]
        R4[flake-budget watch<br/>two consecutive over-budget runs]
    end
    TIER1-->TIER4-->TIER20-->REST
```

## 5) Verschlimmbessern guard — what this plan deliberately does NOT do

- NO Go-side telephony, NO sipgo/pion/ESL (rejected 2026-09-19, tabled).
- NO second user database / cqrs-htmx `setup` usermgmt (split brain).
- NO speculative sip.js 0.22 bump or JsSIP swap (named triggers only).
- NO samber/oops adoption (documented non-fix, tier-3 audit only).
- NO local Playwright E2E while the stack E2E suffices (trigger parked).
- NO CSP relaxation, no CDN, no inline styles (policy pins).
- NO rewrite of AGENTS content away — the size pass MOVES detail to
  docs/, it does not delete knowledge.
- NO new framework/abstraction layers: every task above is additive,
  test-first, and reversible (feature-scoped diffs).
- Behavior parity rule: any port/refactor ships verbatim first, adjusts
  in a separately verified change.

## 6) Coverage proof — every open TODO is in this plan

| Source                                                                                                                    | Items  | Landed in                      |
| ------------------------------------------------------------------------------------------------------------------------- | ------ | ------------------------------ |
| ~~TODO_LIST (10 rows, incl. 2026-09-22 additions)~~ done — snapshot — point-in-time coverage proof; living docs supersede | ~~10~~ | ~~T1-T5, T7-T9, T10, T11~~     |
| ~~My 2026-09-22 status report f-list (30)~~ done — snapshot — point-in-time coverage proof; living docs supersede         | ~~30~~ | ~~T1-T22 (all mapped)~~        |
| ~~ROADMAP open ideas + WORTH_CONSIDERING~~ done — snapshot — point-in-time coverage proof; living docs supersede          | ~~16~~ | ~~T18, T22-T27 + parked list~~ |
| ~~ROADMAP open questions (owner calls)~~ done — snapshot — point-in-time coverage proof; living docs supersede            | ~~12~~ | ~~T11~~                        |
| ~~ROADMAP long-tails (test-infra, cqrs)~~ done — snapshot — point-in-time coverage proof; living docs supersede           | ~~8~~  | ~~T15, T19, T17 + parked~~     |
| ~~22:26 report still-open f-items (13)~~ done — snapshot — point-in-time coverage proof; living docs supersede            | ~~13~~ | ~~T12, T13, T16, T20~~         |
| ~~22:29/01:04 still-open~~ done — snapshot — point-in-time coverage proof; living docs supersede                          | ~~11~~ | ~~T11, T14, T15, T16, T19~~    |
| ~~Standing watches (4 named)~~ done — snapshot — point-in-time coverage proof; living docs supersede                      | ~~4~~  | ~~RITUAL + T8~~                |

(Nothing dropped; parked items stay parked WITH triggers — that is
their correct state, not an omission.)

— Planned 2026-09-22 12:02. Execution order: T1 the moment the pusher
issue is resolved (or via owner push); everything else follows the
graph.

## 7) Execution log (live)

- 2026-09-22 ~13:30 **T1 CLOSED**: webphone + stack pushed; stack
  flake check GREEN ×2 (run #2 on the pre-bump tree, run #3 EXIT=0 on
  the pinned `1a95a736` INCLUDING the daemon's nixpkgs bump — which
  had reintroduced operator.js drift, re-fixed in `1a95a73` before the
  check).
- 2026-09-22 ~14:10 **T2 CLOSED**: pbx-artmann relocked to `1a95a736`
  (probe all-OK, both-arch toplevels green, webphone store path moved),
  narrative commit `f42ef79` pushed.
- 2026-09-22 13:52 **PROD PREmise CORRECTED**: smoke `--base` 16/0 —
  bogus creds REJECTED, `/version` v2.4.0: the owner already deployed
  the v2.4.0 chain; the "v2.1.0 hole live" premise was stale. T3/T5
  become owner-terminal handovers (pbx-artmann AGENTS: no assistant
  ssh).
- 2026-09-22 ~14:05 **T10 CLOSED**: committer = long-lived agent
  session loop (GPG-signed heuristic sweeps; caught twice mid-session).
  Pushes verified restored (test pushes moved origins). Runbook gains
  the phase-boundary + daemon-drift lines; ROADMAP gains the
  docs/status+planning exclusion ask.
- 2026-09-22 ~14:15 **T4b GREEN**: enforced erraudit 0 findings on the
  chain-to-deploy. (4a banner check needs a real extension session —
  prepared as an owner step.)
- 2026-09-22 ~14:20 **T11 PREP**: 14-decision briefing written
  (`docs/planning/2026-09-22_13-50_owner-calls-briefing.md`).
- 2026-09-22 ~15:10 **T6+T7 CLOSED**: v2.5.0 folded, gated, tagged
  (`25740c6`), gh object live; stack relocked in-train (`91ac2c82`,
  browser E2E green, VM test green, stack flake check green, pushed);
  aarch64 `b700` guard green. Release EXIT=0. (First run failed on the
  documented host-go trap — release.sh outside the devShell; rerun
  inside `nix develop -c`.)
- 2026-09-22 ~15:20 **T2' (relock to the tag)**: pbx-artmann pinned to
  `91ac2c82` (probe OK, webphone pin = tag commit), both-arch
  toplevels building.
- 2026-09-22 ~14:50 **T19 CLOSED**: (a) go-health `NewChecks`
  container-free quick-start pushed (`60a0aa9`); (b) family
  release-object audit — 12 latest module tags had NO gh objects (the
  2026-07 v4.3.0 era was the last with objects); all 12 cut (root with
  full CHANGELOG notes, modules with the established sync-link note);
  intermediate historical versions stay tag-only by decision. (c)
  NamedCheck.Timeout retro-note: already shipped in the v4.x
  CHANGELOG — verified.
- Session debts CLOSED: plan mermaid render-verified (mmdc → 48 KB
  SVG), fine table re-split to literal ≤12-minute units, "enourmous"
  typo + coverage recount (10 TODO rows).
- 2026-09-22 ~15:40 **T12 CLOSED**: contacts API hardening landed
  (`contactsLimiter` 60/min burst 60 on POST /api/contacts, atomic
  500-per-extension cap — renames never consume slots, ErrListFull →
  422 — and OpenAPI 3.1 for GET/POST/DELETE /api/contacts with a
  spec-vs-handler test). Green at `8cd327c`.
- 2026-09-22 ~15:50 **T13 CLOSED**: the payload-less `contacts` SSE
  nudge fires from all five mutation paths (tab save/delete/import +
  island API); contacts.templ re-fetches on `sse:contacts` (morph,
  stable ids); island panels.js listens for the bubbling
  `htmx:sseMessage` (decoded against the served sse.min.js: the ext
  fires the trigger AND a bubbling event with the raw MessageEvent as
  detail). Green at `4408342`.
- 2026-09-22 ~15:55 **T14 CLOSED**: two island tests pin the CSRF
  adoption retry ladder (recover on retry 2 / reload only after 3
  failures); rotation-on-slide recorded as NOT-DO
  (`docs/planning/archived/2026-09-22_14-45_csrf-rotation-on-slide-verdict.md`,
  ROADMAP RESOLVED). `43f544e`.
- 2026-09-22 ~16:00 **T17 CLOSED**: island scenarios 10+11
  (re-entrancy collapse; transient `regRebuilding` pill en/de), smoke
  `--expect-version` (verified pos+neg against prod), stack
  `wait_marker` NOTIF-prefix hack → explicit alternation. Daemon swept
  the files into `3e5f220`/`d094d9e`.
- 2026-09-22 ~16:05 **T15 CLOSED**: (a) NOT-DO — treefmt-in-flake-check
  IS the formatting drift gate, no extra check warranted; (b) the
  GOEXPERIMENT=jsonv2 sweep executed (13/13 pkgs green without the
  flag; flake/buildflow/release.sh/smoke/docs updated); (c)
  release-script self-heal verified ALREADY SHIPPED as release.sh
  step 6 (the "outside devShell" failure mode is detected and
  reported by the script itself; the deeper `nix develop -c` re-exec
  improvement stays in TODO as release.sh hardening).
- 2026-09-22 ~16:10 **GATE INCIDENT + FIX**: post-train gates were RED
  (`BUILDFLOW-RC=69`, `FLAKECHECK-RC=1`) — root causes: (1) gofmt
  alignment drift in the server.go New() constructor (the
  contactsLimiter field lengthening; fixed surgically with
  `gofmt -w internal/server/server.go`), (2) a govulncheck Go-version
  mismatch (triage below), (3) confounders: T17 edits + daemon sweeps
  landed while the gates ran. RESOLVED 2026-09-22 ~17:00: gofmt fixed
  surgically; govulncheck GREEN on a quiesced run (the mismatch was a
  raced-run artifact — no env restore needed); the remaining erraudit
  findings were the CONCURRENT CRM train's (7 in internal/crm/client.go,
  then 2 more in store/messages.go) — fixed 2026-09-22 ~17:10 after
  their session went quiet (context vars on the error paths + reasoned
  nolints on the best-effort drain; their tests assert errors.Is
  sentinels, not strings). Enforced tier back to 0 findings.
- 2026-09-22 ~19:35 **T26b + T26c CLOSED**: (b) `turn_rest.secret` +
  `turn_rest.ttl` (48h default) — /config.js derives coturn REST pairs
  per response (username = unix expiry, credential =
  base64(HMAC-SHA1(secret, username))); static passthrough when unset;
  boot validation rejects dead config (secret without a turn:/turns:
  URL) and non-positive TTLs; config + handler tests green (the
  handler test verifies the actual HMAC, not a grep). (c)
  `GET /api/export` (session-gated, Settings-tab download link): zip
  with messages.json (all threads + messages + attachment manifest),
  faxes.json, contacts.vcf — owner-scoped store reads, blobs
  deliberately excluded (manifest only); 401 pin for anonymous.
  Gates on the converged tree: go suite green, erraudit tier-1 = 0,
  buildflow full = RC 0 (the earlier 69 was a host-go invocation, not
  a tree problem — AGAIN the pipeline-exit-code lesson), vulnix app =
  zero real advisories (all distro-patched). ThemeScript knob shipped
  end-to-end by the CRM session (v1.19.2; CSP now script-src 'self'
  with ZERO inline scripts — the hash-pin machinery is gone).
- 2026-09-22 ~19:35 **E2E: RUN 6 FULLY GREEN** (198s, half the 445s
  budget) after TWO driver-authoring bugs surfaced in runs 4/5: (1)
  the logged-out-dial-guard scenario was UNREACHABLE — a real logout
  RELOADS to the login-card shell (session.js), which renders no
  data-dial buttons; the scenario now seeds its own contact,
  reproduces the in-place signed-out island DOM (#phone-view hidden +
  #login-view shown) and pins BOTH the guard toast + #ext focus AND
  the safe-by-construction reload (zero data-dial buttons after).
  (2) ElementClickInterceptedException: transient toasts (transfer
  verdicts) overlay the nav bar; click_tab now retries until the
  click lands. Run 7 (the ×2-green) in flight.
