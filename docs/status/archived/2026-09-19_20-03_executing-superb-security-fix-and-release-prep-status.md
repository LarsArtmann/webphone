# Status: executing the SUPERB hardening plan (continued) — 2026-09-19 20:03 CEST

Session continuation of
`2026-09-19_15-37_SUPERB-prod-recovery-and-hardening-plan.md`
execution (earlier snapshot: `2026-09-19_17-06_executing-superb-hardening-plan-status.md`).
This report covers 17:06 → 20:03.

## a) Fully done (verified, committed, green)

1. **Prod deploy state confirmed.** The owner deployed v2.1.0 during the
   session: `/version` = v2.1.0, `/healthz` ok. The 17:06 report's #1
   blocker is gone. The owner's console paste (17:41 UTC) is the
   in-browser eyeball evidence: REGISTER, session created, csrf adopted,
   a real call with media, history loaded.
2. **Post-deploy CSRF probes vs prod — all green.** `GET /api/csrf` →
   200 with `Secure` cookie, `Cache-Control: no-store`, `Vary: Cookie`;
   tokenless `POST /api/session` → 403 fail-closed.
3. **Stack VM test (P20.2) executed and green.** First run FAILED on my
   assertion (I asserted `trusted_origins == ["https://localhost"]`; the
   module derives `["https://pbx.test"]` from the vhost hostName — the
   green browser E2E proves that value matches what browsers send).
   Fixed the assertion, re-ran with `--no-eval-cache`: green.
4. **SECURITY FIX — forged-session vulnerability (found + fixed this
   session).** Live probe: `POST /api/session` with a WRONG password on
   prod answered **201 Created**. The handler trusted the island's claim
   that a SIP REGISTER proved the credentials; since tab partials, fax/
   attachment streams and SSE fragments scope by the session alone, a
   forged session could read ANY extension's stored threads, transcripts,
   fax documents and contacts. Fix: `pbx.Client.VerifyCredentials`
   (cheapest authenticated phone-api call = the same directory REGISTER
   checks) enforced in `createSession` — 401 on rejected credentials,
   502 when the PBX is unreachable, skip+WARN at boot when no phone API
   (loopback dev). Pinned by `TestVerifyCredentials` (5 subtests),
   `TestSessionCreationVerifiesCredentials` (401 + no session minted +
   201 opens tabs), `TestSessionCreationFailsClosedWhenPbxDown`; the
   contract test gained a third allowlisted 401 writer (login gate).
   Full Go suite green; smoke 26/26. CHANGELOG carries the Security
   entry; AGENTS.md sessions invariant rewritten; README sign-in row
   corrected; plan doc carries the execution addendum.
5. **Send-gateway 502 classification** (your 422 paste): a bridge outage
   during message/fax send answered 422 with the gateway's internal
   error text in the page. Now: upstream failures log the detail
   server-side and answer **502** with a localized "saved as failed"
   note (en+de); validation mistakes stay 422 with their reason.
   `TestSendClassifiesGatewayOutageAs502` pins it; harness gained a
   config-tweakable constructor (`newTestServerWithConfig`).
6. **Smoke `--base` foreign mode rewritten** (M1.5 tool): secret/
   injection-dependent checks now SKIP with a stated reason instead of
   failing falsely; the mode asserts bogus credentials are REJECTED
   (a 201 flags a pre-v2.1.1 build). Local mode unchanged (26/26).
   Against prod: 10 passed, 1 honest RED (the vulnerability, live),
   14 justified skips — it flips green exactly when v2.1.1 deploys.
7. **P14 `scripts/release.sh`** — the manual runbook as one fail-fast
   command (preconditions, fold-check, version bump, gates, tag+push+
   ls-remote verify, lychee, stack relock+gates, aarch64, gh release
   from the CHANGELOG section) with `--dry-run`. Both paths PROVEN:
   negative (fold-check rejects an unfolded tree) and positive (9-step
   dry-run, EXIT=0). The dry-run caught a real bug (stack relock ran
   `git -C <stack> nix flake lock` — nix as a git subcommand) — fixed.
8. **v2.1.1 CHANGELOG folded** (real content, not pre-draft): Security
   (credential verification + Secure flag), Fixed (502 taxonomy, CSRF
   fronting, hookFaxStatus, provider_ref index, …), Added (release.sh,
   foreign-mode smoke, HSTS, drift guard, #log lines…).
9. **P23 closed as a documented verdict** (spec decision, no code):
   sessions are fixed-TTL — there IS no refresh point to hook; CSRF
   already rotates at the two real lifecycle events (login/logout) via
   the proven adoption path. Forcing sliding sessions to gain a rotation
   point = user-visible behavior change + island 403-trap risk for
   marginal gain. Verdict in the plan; ROADMAP carries the
   sliding-session idea; commit `fcd77e3`.
10. **Your three questions answered** (see d) with code where the fix
    was webphone-side.
11. **P24 stack fax-feed module WRITTEN + committed by daemon**
    (verification still open, see b): `telephony.fax.feed.*` options
    (owner/from/webphoneUrl/secretFile/sweepInterval) in options.nix,
    wiring in `fax-feed.nix` (tiff2pdf → python3 poster → Bearer-secret
    POST to `/hooks/fax` as `pdf_base64`; systemd path unit for instant
    pickup + timer sweep as retry policy; failed files stay in place;
    assertions: requires fax.enable + webphone.enable + secretFile;
    unit hardening: NoNewPrivileges/PrivateTmp/ProtectSystem/
    ReadWritePaths). Registered in default.nix.

## b) Partially done

1. ~~**P24 fax feed** — module written, committed; NOT verified:~~ done (stack-side; fax feed later drilled end-to-end (fixture + VM test))
   ~~the stack `nix flake check` is running in the background right now;~~
   ~~the loopback VM test (M24.4: seeded TIFF → feed → webphone 202) is~~
   ~~NOT written; the fixture generation hit a snag (ImageMagick produced~~
   ~~a 16-bit TIFF that `tiff2pdf` rejects — needs `-depth 8`).~~
2. ~~**Owed gates (stale)** — `nix fmt` (prettier pass over the island~~ done (gates green across the 09-20/09-22 trains (buildflow no-cache runs recorded))
   ~~JS hand-edited earlier), `nix flake check` (webphone),~~
   ~~`BUILDFLOW_NO_RESULT_CACHE=1 buildflow` not run this session~~
   ~~(Go suite + smoke ARE green on current HEAD).~~
3. ~~**pbx-artmann** — ahead 2 (unpushed), lock stale again vs the~~ done (relock ritual executed (#1-#3 through 2026-09-22))
   ~~stack's new commits; re-lock + toplevel pre-build pending.~~
4. ~~**P8 decisions** — memos wait for your DECIDED (pin policy, input~~ done (DECIDED 2026-09-20 (AGENTS Owner decisions))
   ~~type).~~
5. ~~**M1.5 post-v2.1.1** — the prod probe red is CORRECT today; rerun~~ done (superseded: prod verified on v2.4.0 2026-09-22 (bogus rejected, 16/0))
   ~~after the v2.1.1 deploy must show `bogus credentials rejected` green.~~

## c) Not started (plan remainder)

- **P5**: two real consecutive browser-E2E greens via `--no-eval-cache`
  (+ optional REGS assert upgrade).
- **P15**: `gh release create v2.1.0` (tag exists), later v2.1.1 +
  announcement draft.
- **P16**: vulnix runtime closure + aarch64 island-lint cross-build.
- **P17**: backup/restore story — nothing landed yet (again).
- **P25**: idiomorph experiment + verdict.
- **P27 remainder**: FEATURES VERIFY pass, `#log`-English check,
  union-coverage note.
- **M2.x**: full in-browser eyeball (your paste covers login + call +
  history; messages-compose and fax-download not yet eyeballed).

> Resolved 2026-09-22 (docs-health): P5 CLOSED (anomaly fixed in 2.5.0,
> E2E green x2); P15 release objects live through v2.5.0 (v2.1.1
> superseded by v2.2.0); P16 rides release.sh per train; P17 shipped
> (2.4.0 backup story + retentionDays); P25 idiomorph shipped v2.4.0;
> P27 VERIFY passes done; M2.x superseded by the stack browser E2E.

## d) Totally fucked up / what I forgot

1. **The v2.1.0 you deployed SHIPS the forged-session hole.** Not my
   fuckup to hide: the vulnerability predates today, but I probed it
   LIVE ON PROD (minted a real session for ext 9999 — read nothing,
   but a live-prod mutation). The honest framing: prod is currently
   exploitable for stored-message/fax/contacts reads; the fix exists;
   v2.1.1 must deploy SOON. The smoke now guards this permanently.
2. **My first bogus-creds probe was sloppy**: I POSTed without a CSRF
   token, got 403, and nearly misread it as "gate works". The 201 only
   showed up once the probe did the full token dance. Lesson applied:
   the smoke's foreign mode does the dance.
3. **The P20.2 assertion failure was MY error** — I asserted a value
   (`https://localhost`) without reading how the module derives the
   origin. Read-the-derivation-first, assert-second.
4. **release.sh shipped with a broken relock command** that only the
   dry-run caught (`git -C … nix flake lock`). Dry-runs exist for
   exactly this; but I should have traced the command before the first
   write.
5. **The 16-bit TIFF fixture** — ImageMagick's default depth broke
   tiff2pdf; one wasted build cycle (`-depth 8` fixes it).
6. **Stale TODO row**: TODO_LIST still carries the "redeploy v2.1.0
   URGENT" row from before the deploy (the new v2.1.1 row sits next to
   it) — pruning is owed in the next doc pass.
7. **Per-task commit discipline was partially defeated** by the daemon
   racing me: the security fix landed as three commits (`ee84b2f`
   core, `a16ae42` tests, `39eca0a` contract+history) instead of one.
8. **gopls false positive** (`refreshCSRF unused`) — ignored per the
   known artifact; CLI-verified in use.

## e) What I could still improve (beyond the plan)

- The island fires `/phone-api/*` refreshes BEFORE login (the 401s in
  your first paste) — cosmetic console noise; could gate the initial
  refresh on session adoption. Cosmetic, not queued.
- The feed's `from` placeholder ("0000") connects to your own-number
  question: set `fax.feed.from` to the trunk faxDid and the Fax tab
  shows a real sender — a small data win until the own-DID story lands.
- release.sh could gain a `--from-tag` resume mode; YAGNI until a
  release actually fails mid-flight.

## f) Next 50 (ordered; owner-blocked items marked)

1. ~~Verify stack `nix flake check` result (running) + commit any fixups.~~ done (done (stack flake check green in every train since))
2. ~~Fix the TIFF fixture (`-depth 8`) + write `tests/fax-feed.nix`~~ done (stack-side; fax feed drilled end-to-end)
   ~~(seeded TIFF → feed → webphone 202 → fed/ archive; bad-TIFF left~~
   ~~for retry) → M24.4.~~
3. ~~Register the fax-feed VM test in the stack flake checks + green run.~~ done (stack-side)
4. ~~README (stack) fax-feed section → M24.6.~~ done (stack-side)
5. ~~Owed: `nix fmt` + `nix flake check` (webphone) +~~ done (done (gates green across trains))
   ~~`BUILDFLOW_NO_RESULT_CACHE=1 buildflow`.~~
6. ~~Cut v2.1.1: `scripts/release.sh 2.1.1` (gates, tag, push, lychee,~~ done (superseded: v2.1.1 → the fixes rode v2.2.0+)
   ~~stack relock + browser E2E — this run doubles as P5 greens #1,~~
   ~~aarch64, gh release).~~
7. ~~Browser E2E green #2 via `--no-eval-cache` → P5 closed.~~ done (CLOSED 2026-09-22 (anomaly fixed, ×2 green))
8. ~~`gh release create v2.1.0 --verify-tag` (M15.1) + CHANGELOG link~~ done (done (gh objects through v2.5.0))
   ~~refs (M15.2) + announcement draft (M15.3).~~
9. ~~P16: `nix run .#vulnix` + `nix build .#checks.aarch64-linux.island-lint`.~~ done (rides release.sh per train)
10. ~~P17: backup/restore story (inventory → rsync/restic pattern →~~ done (done (2.4.0 + retentionDays))
    ~~restore drill on scratch → module timer skeleton).~~
11. ~~P25: idiomorph branch + E2E + verdict doc.~~ done (shipped v2.4.0)
12. ~~P27: FEATURES VERIFY, `#log`-English check, union-coverage note.~~ done (done (VERIFY passes))
13. ~~Prune the stale v2.1.0-redeploy TODO row (doc pass).~~ done (superseded: prod premise corrected 2026-09-22)
14. ~~Re-lock pbx-artmann to the settled stack + `nix build~~ done (relock ritual executed through #3 (2026-09-22))
    ~~.#nixosConfigurations.pbx.config.system.build.toplevel` pre-build.~~
15. ~~**[OWNER] Deploy v2.1.1** (`nixos-rebuild test` → probes → switch).~~ done (superseded: prod on v2.4.0 verified; owner deploy row in TODO_LIST)
16. ~~**[OWNER] Grep the telnyx-webhooks bridge journal** for the 422'd~~ done (still open — owner TODO row (SMS lane))
    ~~send; restore the SMS lane.~~
17. ~~**[OWNER] P8 DECIDED lines** (pin policy + input type) → AGENTS.~~ done (DECIDED 2026-09-20 (AGENTS))
18. ~~**[OWNER] Own-number feed decision** (TODO row has the 3 options).~~ done (DECIDED: static identities map (AGENTS); /phone-api feed = upgrade path)
19. ~~Post-deploy: smoke `--base` must be fully green (bogus-creds 401);~~ done (smoke --base 16/0 verified 2026-09-22)
    ~~verify the startup `pbx credential verification` journal line.~~
20. ~~Post-deploy: in-browser eyeball of messages-compose + fax-download.~~ done (superseded: browser truth via the stack E2E)
21. ~~Record the forged-session incident in the 11:02 report's timeline~~ done (story lives in docs/lessons.md + CHANGELOG Security)
    ~~appendix (doc pass).~~
22. ~~Consider SECURITY.md (disclosure contact) — the incident shows the~~ **Won't implement — SECURITY.md not adopted.**
    ~~repo needs one.~~
23. ~~Consider a static per-release `gh release` template.~~ **Won't implement — release template not adopted.**
24. ~~openapi: `POST /messages/send` + `/fax/send` 502 documentation~~ **Won't implement — openapi stays on /api surfaces.**
    ~~(the openapi doc covers session/csrf only today).~~
25. ~~Island: gate initial `/phone-api` refreshes on session adoption~~ done (superseded: wp:session-opened gating shipped (no cold-login 401 storm))
    ~~(kill the 401 console noise).~~
26. ~~Smoke: fold the new `-depth 8` TIFF fixture into the repo (reusable~~ done (stack-side fixture)
    ~~by fax tests).~~
27. ~~fax-feed: alert hook on repeated failed sweeps (resilience.nix~~ done (stack-side)
    ~~OnFailure pattern).~~
28. ~~fax-feed: pages count from TIFF metadata (tiffinfo) instead of the~~ done (stack-side)
    ~~omitted field.~~
29. ~~fax-feed: `done/` retention policy (prune fed TIFFs).~~ done (stack-side)
30. ~~Consider forwarding fax TIFFs to restic backups explicitly (they~~ done (stack-side)
    ~~land under recordings/fax already — check the backup scope).~~
31. ~~Drift guard: also pin the stack's `webphone` input to the newest~~ done (DECIDED ride-main (AGENTS))
    ~~tag (currently rides main by decision — document-only note).~~
32. ~~Smoke: add a check that the startup WARN for PBX-less mode appears~~ **Won't implement — boot-log WARN smoke check not adopted (loopback dev mode documented).**
    ~~in the journal (boot-log contract).~~
33. ~~release.sh: print the resolved versions (flake + newest tag) at~~ done (release.sh prints versions at step 1)
    ~~step 1 for the transcript.~~
34. ~~Docs: AGENTS "Release runbook" step 1 now says "fold" — reference~~ done (runbook references release.sh as the one-command path)
    ~~release.sh as the one-command path.~~
35. ~~`parseOwnerFrom`: consider rejecting the literal "0000" in prod~~ **Won't implement — placeholder DID cosmetics not adopted.**
    ~~docs (placeholder noise in the Fax tab) — or embrace it as the~~
    ~~documented unknown-sender marker.~~
36. ~~Verify `tiff2pdf` output opens in a browser (the Fax tab serves~~ done (stack-side (verified in the fax-feed drill))
    ~~the PDF as-is; multipage OK).~~
37. ~~fax-feed VM test: also assert the path-unit instant pickup (drop a~~ done (stack-side)
    ~~TIFF while running, no manual start).~~
38. ~~Consider `Restart = on-failure` + `RestartSec` on the feed service~~ done (stack-side)
    ~~for tight retry loops (timer sweep may suffice; test will tell).~~
39. ~~Confirm LoadCredential + PrivateTmp interplay (CREDENTIALS_DIRECTORY~~ done (stack-side (asserted in the VM test))
    ~~visible under PrivateTmp — yes, systemd mounts creds in /run/credls;~~
    ~~assert in the VM test anyway).~~
40. ~~CHANGELOG: v2.1.1 Added section should mention the fax feed once it~~ done (stack CHANGELOG owns the fax-feed entry)
    ~~lands (it's a stack feature, not webphone — only if visible).~~
41. ~~Todo tool state: P5/P14/P23/P24 statuses updated at interrupt.~~ done (done (session state recorded))
42. ~~Session report pointer appended to the plan doc execution log.~~ done (done (plan execution log entries))
43. ~~`nix run nixpkgs#lychee -- .` BEFORE tagging (release.sh runs it~~ done (lychee pre-pass folded into release.sh flow)
    ~~after push; a pre-pass avoids tag-rele/tag dance).~~
44. ~~Double-check `webphoneVersion` sed idempotence after the 2.1.1 run.~~ done (version bump verified by the drift test)
45. ~~Keep `--no-eval-cache` for ALL stack VM-test reruns (cache-hit trap).~~ done (--no-eval-cache lesson recorded (lessons/runbook))
46. ~~ROADMAP: own-DID endpoint idea could ride the operator window's~~ done (superseded: identities map shipped; feed upgrade path recorded)
    ~~phone-api instead (one API, two consumers).~~
47. ~~AGENTS: add the fax-feed module to the tri-repo integration state~~ done (AGENTS tri-repo section current)
    ~~section (stack-side contract).~~
48. ~~The 3 allowed 401 writers: add to docs-health's FEATURES VERIFY~~ done (401-writer allowlist test shipped (contract test))
    ~~checklist.~~
49. ~~Consider renaming `--base` smoke mode's summary line to include~~ **Won't implement — --base summary wording unchanged.**
    ~~"foreign" explicitly (minor UX).~~
50. ~~Celebrate the part that worked: negative-probe discipline (P3 gate,~~ done (negative-probe discipline institutionalized in smoke + release.sh)
    ~~bogus-creds dance, release.sh dry-run) caught three real bugs~~
    ~~before they could ship.~~

## g) Owner questions (3)

1. **v2.1.1 timing**: prod is exploitable for stored-thread/fax/contact
   reads (no password needed). OK to cut v2.1.1 and hand you the deploy
   block as soon as gates go green (target: today)?
2. **SMS bridge**: the 422 you hit is the stack's telnyx-webhooks bridge
   failing. Do you want the bridge journal grep + fix as part of this
   session (stack-side), or is Telnyx messaging knowingly not live yet?
3. **Own-number feed**: which source should own the extension→DID
   mapping — a new `/phone-api` identity endpoint (stack-side work),
   a static config map in webphone, or CDR-derived? (TODO row has the
   tradeoffs.)
