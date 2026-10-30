# Status — fax-truth train: the pre-switch sweep, the stale owner sheet, and what is left

**Snapshot-rot header: claims verified as of 18:33–18:38 CET, 2026-10-30. Report
written 20:29 CET.** The auto-commit daemon and at least one concurrent session
(dkim/Telnyx mail-fleet train in pbx-artmann) were active throughout — every
rev, count, and CI verdict below is a 18:3x snapshot, not a live claim. Nothing
was re-verified after 18:38.

**Session scope:** the follow-through plan
(`docs/planning/2026-10-30_15-50_SUPERB-switch-the-truth-and-close-the-seam-gaps.md`)
was already executed by the prior session except the owner-gated tail. This
session was the pre-switch verification sweep: trust the handover summary
**nothing**, re-verify the tri-repo state end-to-end, repair what drifted, and
leave the owner exactly one paste-chain from switching. One real defect was
found and fixed. Two hours later the switch still has not been reported as run.

---

## a) FULLY DONE (this session)

1. **Todo-list reconciliation.** The inherited list was stale (nearly all
   items done). Rebuilt: 5 completed groups, 3 genuinely-open (all
   owner/switch-gated).
2. **webphone verified** (18:34): `1ee39028`, working tree clean, zero
   unpushed commits, CI **green** on HEAD (run 41095609996, push-triggered,
   10m35s).
3. **Stack (nix-international-telephony) verified** (18:34): `f64f4fd`, clean,
   pushed, CI **green** on that rev (run 41089341722, 26m39s), and its
   `flake.lock` pins webphone **`e6fb14db`** — the ride pin, confirmed by
   parsing the lock (python, not jq — tool-shell rule).
4. **pbx-artmann verified** (18:34–18:35): `flake.lock` pins telephony
   **`f64f4fd`** (→ webphone `e6fb14db` transitively). Staged root
   `a2slrp3c…-nixos-system-pbx-26.11.20261027.31ecbac` still present in the
   store. `/tmp/pbx-toplevel-current` → `yyib5k9g…` = the **live**
   pre-switch generation tracker (this symlink tracks the deployed closure,
   not the staged one — naming gotcha worth remembering).
5. **Real-host fax eval re-verified** (18:38):
   `nix eval --json .#nixosConfigurations.pbx.config.services.webphone.settings.fax`
   → `{"receive":false,"send":false}` — the claim the whole switch rests on,
   re-proven on the current tree. (A pre-existing `stalwart 0.15.5` eval
   warning surfaced — belongs to the concurrent mail train, flagged in §f-46.)
6. **THE DRIFT — found and fixed.** The owner switch sheet
   (`pbx-artmann/docs/runbooks/owner-switch-2026-10-30-fax-truth.md`) was
   **stale**: written 15:00 for the FIRST staging, it demanded relock
   `05a442e` "on top" (impossible — HEAD had moved three commits past) and
   named superseded revs (`b035e74e`/`ad122b9`) in the title, the change
   paragraph, and the rollback. The owner's §0 blocking check would have
   **failed tonight**, stalling the switch at best. Seven edits: title +
   update notice carry the ride pin; the relock chain told whole
   (`05a442e` + `eda38d5`); §0 pre-flight made daemon-robust (grep the lock
   for `f64f4fd6`, `rev-list origin/master..HEAD --count` = 0 — a fixed HEAD
   expectation can never survive daemon commits); ride deltas summarized
   (always-on fax row, Message-gateway rename, conditional Paperless row);
   a §3 ride probe added (gateway row must read **"Message gateway"** —
   discriminates `e6fb14db` from the pre-ride binary); rollback corrected to
   revert **both** relocks + the on-host `nixos-rebuild switch --rollback`
   shortcut. Committed **`d540f4b`**, pushed.
7. **Daemon push-backlog cleared (side effect):** the push carried three
   stalled daemon commits (`a10d471`, `d87ad86`, `adb0ece` — all concurrent
   session docs) — `cc1e3f3..d540f4b` went up in one push. The manual-push
   bar was NOT invoked (oldest stall was ~8 min, far under 60).
8. **Books verified as claimed** (18:34): status report §UPDATE carries
   flake-FULL (incl. KVM backup VM) + smoke-in-CI verdicts + RIDE COMPLETE;
   TODO_LIST "Live smoke in CI" struck SHIPPED; the fax-enablement section
   carries the owner decision + the two re-open lanes.
9. **Deliberate non-action, recorded:** no restage. The concurrent session
   was mid-edit (`TODO_LIST.md` + a status doc modified at check time) and
   the sheet's own §0 forbids staging a dirty working tree. The owner's
   leg 1 (`nix run .#deploy`) re-stages fresh — this is REQUIRED, not a
   confirmation formality (see §b-2).

## b) PARTIALLY DONE

1. **The train as a whole:** staged-but-not-switched. Every claim upstream of
   the switch is verified; the switch is what converts the fax-truth train
   from claim to user-visible honesty.
2. **Staging freshness:** the staged root `a2slrp3c` is docs-stale relative
   to pbx HEAD (daemon doc commits landed after the last guard run). Harmless
   for correctness — the switch legs evaluate the tree fresh — but it means
   leg 1 will rebuild a docs-churn child and print NEW store paths. Expect
   that; readlink is the truth.
3. **Post-switch books:** designed and specified (TODO INERT→LIVE flip with
   timestamp, status-report switch-verdict block, pbx CHANGELOG DEPLOYED
   ledger row at the file tail) — not executable until the switch runs.
4. **Owner sheet:** revs now accurate; but its §0 paragraph still names the
   14:57 mail-train files as THE hazard — historical color that could mislead
   a future reader about what "clean" means. The generic `git status --short`
   check covers it; a one-line genericization would help.
5. **The pin gap:** webphone `main` (`1ee39028`) sits ~7 commits past the
   pinned `e6fb14db` (docs/tests + the Paperless Settings row — zero
   deployed-host delta because paperless is unconfigured in prod). Recorded
   decision: no dedicated ride. The gap is known and bounded, but it should
   be restated in the post-switch books so it cannot grow silently.

## c) NOT STARTED

1. **T1 `[OWNER]`** — the switch legs (sheet §0–§4).
2. **T3** — live off-state proof (`wp-fax-off` present, `wp-compose-fax`
   absent, DE render, `/fax/send` honest 422/503, session-less 401,
   screenshots). Needs the switch + a session.
3. **T4** — post-switch bookkeeping. Needs T3.
4. **T8/T9/T10 `[OWNER]`** — native DE read; Telnyx facts (owning connection
   for `+1727289311`, fax-DID options/prices, fraud ticket); Paperless
   per-fax archive-marker ruling.
5. **Unknown-key alarm first real-fire drill** (VM stale-pin boot) — optional
   plan tail, never scheduled.
6. **Pre-switch `/version` baseline probe** — see §e-3; not done this session.

## d) TOTALLY FUCKED UP

**Nothing broke this session and no landed defect exists — but the sharpest
finding indicts the prior session's arc (also this assistant):** the second
ride (16:2x–18:2x) updated the stack pin, the pbx relock, the status report,
the CHANGELOG — and **forgot the one document the owner pastes from**. The
sheet went to bed demanding a relock three commits in the past. A standing
runbook that names revs is an unowned cache, and nobody owned it. That is the
defect class; the repair was 30 minutes, the near-miss was a dead-ended owner
at switch time.

Near-misses of THIS session, honestly:

1. **The sweep almost shipped on the handover's claims.** My verification
   checklist did not include "audit the owner sheet's revs". I caught the rot
   only because the sheet's title line carries a rev string and I cross-grepped
   store paths — and the first grep returned EMPTY (no store paths in the
   sheet at all). Read lazily, "no stale paths" would have read as "sheet
   fine" and I'd have stopped. The rot was in REV strings, not paths. The
   lesson: an empty verification result must be distinguished from a
   verification that covered the real risk surface.
2. **Own-artifact verification gap:** I diff-stat'ed the sheet edit and
   verified the edit strings applied (7/7), but did not re-read the full
   final file before push. During THIS report I re-read lines 1–45 (verified
   good); the §0–§4 region matched my edit strings but a full top-to-bottom
   read is the actual bar. Residual: §f-35.
3. **No `/version` baseline:** a single anonymous `GET /version` before the
   switch would timestamp the deployed generation and give the post-switch
   proof a public discriminator. Skipped; nothing public discriminates
   pre/post switch today, which is why §g-Q1 is genuinely unanswerable from
   here.
4. **Timestamp hygiene held but is worth repeating:** every claim in the
   handover summary and this report is already hours old at reading time.
   Timestamped, not timeless — per the ratified convention.

## e) WHAT WE SHOULD IMPROVE

1. **Sheet-freshness rule (one home: release-runbook + AGENTS):** any relock
   invalidates every standing owner sheet's rev references — the relock
   ritual ends with "does an owner sheet / runbook name the old revs?" and
   updates or supersede-marks it in the same commit.
2. **Machine-check the sheet:** the deploy app should PRINT the pinned
   telephony/webphone revs; a guard (or the sheet itself) then quotes deploy
   output instead of hand-written revs. Better: a check that diffs sheet-named
   revs against `flake.lock` and refuses a stale sheet.
3. **`/version` baseline probe:** add "record the live `/version` BEFORE the
   switch" to the sheet's §0 — turns the anonymous endpoint into a
   pre/post discriminator and gives the evidence block a public anchor.
4. **Verification bar for pushed prose:** diff-stat is not verification;
   pushed docs get one full re-read (this session's residual, §f-35).
5. **pbx relock message contract:** the ride relock `eda38d5` carried a
   generic Dependabot-style message; the stack's lock-move-guard demands rev
   ranges, and the first relock `05a442e` followed it. pbx's relock app (or
   runbook) should enforce the same shape so `git log --grep relock` stays a
   reliable index.
6. **Daemon push-stall observability:** the stall recurred (~8 min, benign,
   cleared by my push). The T15 `push-stall-watch` timer is shipped in
   SystemNix but ARMED coverage is unverified: is it active on THIS host, and
   does its sweep cover the pbx-artmann checkout? (§f-31.)
7. **Shared-history pushing:** my push published the concurrent session's
   half-narrative commits. Harmless this time, but "pushing publishes shared
   history — check `git log origin..HEAD` before any manual push" deserves
   one line in the runbook.
8. **Status-report discoverability:** this report supersedes the sweep state
   in `2026-10-30_02-25_fax-capability-honesty.md`; a one-line forward
   pointer was appended there (done this session) — keep that habit for every
   follow-up report.

## f) UP TO 50 THINGS TO GET DONE NEXT

`[O]` = owner, `[A]` = assistant. Sorted: switch-critical first.

**The switch + live proof (P0):**
1. `[O]` Leg 0: confirm pbx tree clean + fully pushed (concurrent train settled).
2. `[O]` Leg 1: `nix run .#deploy` — guards green, deploy-freshness FRESH, legs printed (expect NEW docs-churn paths vs `a2slrp3c`).
3. `[O]` Leg 2: `nixos-rebuild test` + the two journal probes — NO `unknown configuration keys` WARN.
4. `[O]` Leg 3: restart `telephony-secrets-perms webphone telephony-agent`.
5. `[O]` Leg 4: `./tests/installer/verify-live.sh` — the PASS battery.
6. `[O]` Journal quiet-check over a LONG window (hours, not `-n 40`) — the pin-gap alarm's silence is the acceptance test.
7. `[O]` Leg: `nixos-rebuild switch` — make it permanent.
8. `[A]` Authenticated `GET /fax`: `wp-fax-off` ≥ 1, `wp-compose-fax` = 0 (execution mode per §g-Q2).
9. `[A]` `GET /settings`: fax row copy + the **Message gateway** ride probe (proves `e6fb14db` live).
10. `[A]` `wp-lang=de` cookie render — the DE notice reads idiomatically in situ.
11. `[A]` `POST /fax/send` (CSRF dance): honest 422/503, no fake success.
12. `[A]` Session-less `GET /fax` → 401 (the notice never leaks anonymously).
13. `[A]` Screenshot the live off-state panel (light + dark) into the evidence block.
14. `[A]` TODO_LIST §Fax enablement: INERT → LIVE flip with switch timestamp + verdict.
15. `[A]` Status report: switch-verdict block (revs, timestamps, probe outputs).
16. `[A]` pbx CHANGELOG: DEPLOYED ledger row at the FILE TAIL (old → new closure paths, live-since, riders: the concurrent mail train named).
17. `[A]` Post-switch `nix run .#deploy` re-stage so staged == live.
18. `[A]` `/version` before/after comparison (baseline first — §f-29).
19. `[A]` Mark the owner sheet DEPLOYED (dated note) or supersede it — it must not survive as a live doc.

**Owner tail (decision-grade):**
20. `[O]` T8: native DE read of the 16 new strings; `[A]` fixes + re-pin + parity tests if flagged.
21. `[O]` T9: Mission Control — which connection owns `+1727289311`.
22. `[O]` T9: fax-DID options + prices; is T.38 offered on the FQDN connection?
23. `[O]` T9: outbound-fraud ticket status (outbound 403s block any future fax send too).
24. `[A]` Fold the Telnyx facts into the decision doc's re-open triggers.
25. `[O]` T10: Paperless per-fax archive marker — build or decline; if build, file the fax-timeline shape.
26. `[O]` Unknown-key alarm first real-fire drill (deliberate stale-pin VM boot) — optional.
27. `[O]` Backup drill scheduling post-switch (`readmodels.db` is user data since the metaengine train).

**Process + hardening (noticed this session):**
28. `[A]` Write the sheet-freshness rule into release-runbook + AGENTS (§e-1).
29. `[A]` Pre-switch `/version` baseline probe — sheet §0 line + do it next session.
30. `[A]` Deploy app prints pinned telephony/webphone revs; guard diffs sheet revs vs lock (§e-2).
31. `[A]` Verify `push-stall-watch` timer is ARMED on this host and covers pbx-artmann checkouts (§e-6).
32. `[A]` Runbook line: manual pushes publish shared history — inspect `origin..HEAD` first (§e-7).
33. `[A]` Watch smoke-in-CI over the next pushes; if it flakes, pin retry-once (the fax-test lesson: wait, don't probe once).
34. `[A]` Audit `tests/fax.nix` (stack) for other single-probe `wait_until_succeeds` — the ESL-binding race class may not be unique to mod_spandsp.
35. `[A]` Full top-to-bottom re-read of the final owner sheet (own-artifact verification bar, §e-4).
36. `[A]` Verify the stack's telephony-eval has a POSITIVE faxDid fixture (trunk WITH a DID → assertion silent), not just the negative.
37. `[A]` FEATURES.md: confirm unknown-key WARN, Paperless row, fax lanes are inventoried.
38. `[A]` webphone CHANGELOG: confirm T11/T29/T30/T7 lines exist under [Unreleased].
39. `[A]` Re-read the vulnix verdict post-switch (closure changes) vs the buildflow warning list.
40. `[A]` Restate the ~7-commit pin gap in the post-switch books so it cannot grow silently (§b-5).
41. `[A]` Coordinate the shared-closure narrative with the dkim/Telnyx session — their 18:28–18:32 status docs ride the same DEPLOYED row.
42. `[A]` Confirm the `/tmp/pbx-toplevel-current` reboot-repair recipe is written in the pbx runbook (currently tribal knowledge).
43. `[A]` Tier-1/tier-2 re-measure is due **2026-11-05** (AGENTS cadence) — calendar it.
44. `[A]` Post-switch journal watch for the first hours (no sentry: grep ERROR/WARN classes).
45. `[A]` Flag the `stalwart 0.15.5` eval warning to the mail-train session (observed during the fax eval; not this train's to fix).
46. `[A]` Confirm the ride decision stands: no dedicated ride for the 7 docs/tests commits unless a user-visible delta appears.
47. `[A]` Archive the 15:50 plan doc with DONE markers so it does not rot into folklore.
48. `[A]` dom-contract: confirm the conditional fax nodes carry a deployed-verdict note post-switch.
49. `[A]` TODO_LIST: strike plan-execution rows this session closed (sheet refresh, sweep) if any name them.
50. `[A]` Genericize the owner sheet §0 hazard paragraph (name the CHECK, not the 14:57 file list — §b-4).

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Has the switch run — and if not, when do you plan to run it?** I cannot
   detect it: no-ssh doctrine, and no public endpoint discriminates the
   pre/post-switch binary today (see §e-3 — the `/version` baseline would
   have fixed that; I can start recording it next session).
2. **Who executes the authenticated live proof (T3, items 8–13)?** Either you
   paste the curl/DOM outputs and I archive them, or you hand me a temporary
   session cookie and I run the whole battery + screenshots myself. This
   decides the execution mode before the switch, not after.
3. **Do you want the T8/T9/T10 owner tail closed in this session right after
   the switch** (DE read, Mission Control facts, archive-marker ruling), **or
   scheduled to the next train?** The decision doc's re-open triggers stay
   live either way; this is purely a scheduling call.

---

*Verified as of 18:33–18:38 CET snapshot; written 20:29 CET. Next state
change that matters: the owner's leg 0–2.*
