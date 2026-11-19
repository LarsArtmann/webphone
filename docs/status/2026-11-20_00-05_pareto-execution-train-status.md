# Pareto execution train status — 2026-11-20 00:05

> Snapshot-rot header: claims verified as of 00:05 CET. CI on
> `origin/main` reads GREEN via `scripts/ci-verdict.sh`; two daemon
> commits (492b6064, 60909463) were unpushed at snapshot time. This
> report covers THIS session's run only: executing
> `docs/planning/2026-11-19_19-46_SUPERB-pareto-backlog-execution-plan.md`
> (T01–T27) under the owner's "whole TODO list" directive.

## a) FULLY DONE (verified this session)

1. **Phase 0 — plan landed + gates green.** The pareto plan file is
   committed/pushed; docs-gates battery full PASS (the freshness red I
   was healing was the CONCURRENT session's own 19:47 finding — they
   harvested it themselves; verified, not assumed).
2. **CI-red root cause found and FIXED (the deploy blocker).** The
   `catalog-contract` job (landed 2a46a9c6 by the concurrent session)
   fetched `https://catalog.home.lan/contracts/catalog-frontmatter.v2.schema.json`
   — a LAN name GitHub runners cannot resolve; three CI runs sat red on
   exactly that DNS error. Fix: vendored the byte-identical schema to
   `docs/contracts/catalog-frontmatter.v2.schema.json` (fetched from the
   live hub THIS session, sha256
   `d626ab6a14d0e607605c34743d03769380ebee9a0e1623a0dad203c6dc9baf14`
   recorded in the README with a refresh ritual), CI now validates
   against the pinned copy. **CI is GREEN as of this snapshot.**
3. **T04 — `usermgmt.db` backup gap CLOSED (R20-companion).** Both
   `package/backup-script.nix` branches carry the guarded-if identity
   DB; the backup VM test boots WITH passkey (raw settings lane) and
   asserts presence + `pragma integrity_check` in the single-snapshot
   AND dated-snapshot arms; module-check pins the string in both script
   shapes; both module-output goldens regenerated. Verified green:
   `webphone-module`, `webphone-backup-drill`, `webphone-backup` (KVM).
   CHANGELOG entry + TODO companion-debt clause struck with evidence.
   (Files landed via daemon sweeps — see d).1.)
4. **T05 — fax Resend×retention (R30): landed by the CONCURRENT
   session** (copy-blob arm + `TestResendOwnsItsBlobCopy`); I verified
   the suite green and the TODO strike already in place. Not double-built.
5. **T06 — workflow-e2e promoted to BLOCKING (R29).** Evidence: eight
   consecutive green runs (44540865097→44604487826, green even where
   other jobs failed, zero browser flakes). `continue-on-error`
   deleted, job renamed "workflow e2e (browser)", no branch-protection
   contexts existed to break (API-verified). AGENTS command comment +
   CHANGELOG + TODO strike updated. Committed WITH narrative (2f01f25b)
   and pushed.
6. **Formatter repairs on two fresh concurrent reports** (20:23
   catalog assessment, 20:51 catalog-ring plan): canonical dprint bytes
   applied (diff verified whitespace-only) so CI's format gate stays
   green despite mid-flight authoring.

## b) PARTIALLY DONE (in-flight, uncommitted-or-just-swept)

1. **T07–T09 — the three browser journeys** (R26/R27/R28), written
   into `scripts/workflow-e2e.py` (syntax-verified; already swept by
   the daemon):
   - `FakePhoneAPI` loopback stand-in (voicemail summary/messages rows,
     DELETE, one LIVE WAV recording + one DEAD 404 recording for the
     honest-toast-once contract, empty CDR history).
   - `boot()` gained a config-file lane (`WEBPHONE_CONFIG` JSON for the
     maps env cannot carry: `auth.users`, `identities`,
     `extension_password_files`, passkey rp pair) + `env_extra`
     (`WEBPHONE_DIAL__COOLDOWN=2m`, `WEBPHONE_PHONE_API_URL`).
   - J1 opens the break-glass disclosure before filling the extension
     form (passkey-on deployments hide it one `<details>` away).
   - First full run verdict: **all 10 pre-existing journeys still
     green** (no regression) + the dial-gate HTTP leg
     (records-then-denies) green. FAILURES driving the current rework:
     (i) the admin nav link does NOT render for an extension login —
     `isAdminSession` is LoginEmail-only BY DESIGN (admin.go), so the
     admin journey needs a mapped login; (ii) the switcher 403-toast
     assert raced on empty toast text; (iii) `dest.clear()` threw
     ElementNotInteractable and KILLED the run (voicemail journey
     never executed).
   - **Rework in progress**: full passkey lane via chromedriver's
     virtual authenticator + `--host-resolver-rules=MAP webphone.test
     127.0.0.1` (domain origin without /etc/hosts or TLS): CLI-minted
     enroll token → real /enroll ceremony → real /login ceremony → a
     MAPPED session, which unlocks the admin journey AND upgrades the
     switcher journey to the DESIGNED success path (switch to the
     second owned line, reload, whoami flips). `make_driver()` edit
     applied; journeys + wiring still to finish.
2. **T01/T02/T03 — deploy + prod verification windows**: UNBLOCKED by
   the green CI but NOT STARTED. Stack pin gap confirmed (stack locks
   webphone at `cfba7252`, main is ~40+ commits past it); deploy
   mechanics verified from pbx-artmann README
   (`nixos-rebuild switch --flake .#pbx --target-host
   root@pbx.artmann.tech`); release-runbook stack-bump + pbx relock
   rituals read and listed.

## c) NOT STARTED

T10 export-or-confirm tool (R24); T11 catalog ReviewSaved parity (R31 —
NOTE: the concurrent session's 20:23 report lists the SAME work, see
g).3); T12 FindThread rename (R32a); T13 type tails (R33); T14 comment
truths (R34+R35); T15 fcr small-findings batch (R37); T16 cadence
pre-drain (R36); T17 stack drift + coverage arms (R04+R08); T18
ringhook docs (R16); T19/T20 android lines[] + triggers (R01/R18/R19);
T21 SSO-only fixture (R02); T22 fleet health batch (R03/R14/R15); T23
CUE ruling record (R05); T24 notes-search spike (R25); T25
hardening-parity pin (R12); T26 bandit issue filing (R13); T27
flightrecorderhealth demand check (R17). The gated rows (R21 owner
register, R22 stack lane, R23 spike, v2.15.0 order, ThreadGate Q1–Q5)
stay owner-scoped by plan.

## d) TOTALLY FUCKED UP (honest ledger)

1. **Daemon beat my narrative commit on T04.** I ran the verification
   battery BETWEEN editing and committing; the daemon swept all T04
   files into three `chore: auto-commit` heuristic commits, so my
   detailed message found "nothing to commit". The narrative survives
   only in the CHANGELOG + TODO strike. Lesson recorded: with the
   daemon + a concurrent session both live, commit IMMEDIATELY after
   the edit batch and run long verifications after (or amend the
   unpushed sweep per the pbx-artmann ritual).
2. **The CI red sat ~1h before I diagnosed it.** The island-storm
   closeout §f item 1 had already named the catalog job + commit; I was
   heads-down in T04 and re-derived it from scratch. Read sibling
   reports' §f FIRST at every gate red.
3. **Pipe-swallowed exit code** on the first T04 module-check run
   (`nix build … | tail && echo OK` — the runbook's own documented
   trap). Caught on the second look; the golden-regen detour (TWO
   goldens, not one) cost ~15 min.
4. **First e2e full run crashed mid-suite**: an unhandled
   ElementNotInteractableException in my dial journey aborted the
   whole run (voicemail journey never executed) — my new journeys did
   not follow the file's own try/except-per-assert pattern. The rework
   fixes the pattern, not just the element.

## e) WHAT WE SHOULD IMPROVE

- **Journey isolation**: every new e2e step wrapped in Check+try/except;
  one journey's failure must never orphan the rest.
- **Interactability**: prefer JS value-set + `requestSubmit()` for
  island form fills (the island's own listeners fire; no chromedriver
  interactability dependency).
- **Standing CI watcher**: start a `ci-verdict.sh` poll loop at train
  start instead of ad-hoc re-checks (I re-implemented it three times).
- **Sibling §f scan**: make "read the newest sibling report's §f" the
  first action on ANY red gate — it answered today's red within
  minutes of being written.
- **Vendored-contract drift**: the catalog schema copy needs a refresh
  trigger when the hub bumps to v3 (documented in
  docs/contracts/README.md); consider a lint when the live hub is
  reachable from this host.

## f) Up to 50 things to get done next (this train's tail, ordered)

1. Finish the passkey virtual-authenticator lane in workflow-e2e.py
   (enroll → login → mapped session).
2. Switcher journey: assert the SUCCESS path (switch to 1002, reload,
   whoami/`#wp-ext-switch` value flips).
3. Dial-gate journey: JS value-set + requestSubmit; assert
   `#wp-dial-cooldown` + "Dial anyway" chip.
4. Voicemail journey: run it (first execution ever — was orphaned by
   the crash); tune toast-once + delete + empty-state waits.
5. Full e2e green run, then a SECOND consecutive green run (blocking
   lane now demands stability).
6. Land T07–T09; strike R26/R27/R28 with run evidence; CHANGELOG line.
7. T01 cycle-1: stack `nix flake lock --update-input webphone`, commit.
8. T01: stack gates (`telephony-webphone` VM test, `nix flake check`,
   `.#telephony-browser` E2E — served markup moved since v2.14.0, the
   obligation applies).
9. T01: pbx-artmann relock ritual (rev swap via git rev-parse, flake
   update telephony, lock-drift-probe, both toplevel builds).
10. T01: `nixos-rebuild switch --flake .#pbx --target-host
    root@pbx.artmann.tech` (owner confirmation pending — g).1).
11. T01: `/version` verify + unknown-key boot-WARN quiet check
    (pin-predates-key trap mitigation).
12. T01/R07: host diagnostics — systemctl status, journal tail, Caddy
    vhost health; one-paragraph owner-facing verdict note.
13. T02: prod authenticated capture (`ui-capture.py --base-url
    https://pbx.artmann.tech --track`, env creds).
14. T02/R11: ring-guard positive path through the real `/hooks/ring`.
15. T03/R10: night-run stability pair + anchor the repeat.
16. T10: export-or-confirm tool per the 2026-11-14 design (m10.1–m10.5).
17. T11: coordinate with the concurrent session (g).3) — ReviewSaved +
    derived parity.
18. T12: FindThread → ResolveThreadID (LSP rename + call sites).
19. T13: Locality constants, ASRProvider enum, LoginPageExtra params.
20. T14: apidocs response-header comment fix + favicon token unification.
21. T15: FCM single-flight, zero-predicate deletions, fax inbound cap,
    blob brand-before-write, seq comment (verify which the concurrent
    session already landed — R30 taught the check-first lesson).
22. T16: cadence pre-drain sweep script.
23. T17: stack drift audit + coverage arms.
24. T18: ringhook operator docs + typed mirror.
25. T19: android lines[] adoption (fixtures exist).
26. T20: android spec consumption + trigger check.
27. T21: SSO-only stack fixture verify.
28. T22: pma push heuristic, hub runner fix, ALERT_WEBHOOK_URL.
29. T23: CUE ruling one-liner + pointer.
30. T24: notes-in-search spike note.
31. T25: hardening-parity pin test.
32. T26: bandit-provider issue filing (verify-before-filing first).
33. T27: flightrecorderhealth demand check + gate verdict.
34. Cycle-2 re-bump + redeploy after the webphone-local trains land.
35. Owner register: the 3 questions from the 14:45 report + g) below.
36. Vendored catalog schema: add a "live hub reachable → byte-compare"
    local probe to the docs contracts README ritual.
37. e2e: assert the break-glass disclosure STATE (closed) so a markup
    regression can't silently re-primary the extension form.
38. e2e: line-switch reload assertion should pin the SSE resubscribe
    (the `#wp-sse-live` dot reappears) — one more boot-probe truth.
39. Consider pinning the FakePhoneAPI shapes against
    `internal/pbx/client.go` structs in a comment (drift alarm for the
    next wire change).
40. CHANGELOG: the e2e journey additions owe their `[Unreleased]` line
    when landed.
41. The daemon-race lesson → AGENTS "Release runbook" section: add the
    "commit before long verification" line (one sentence).
42. Re-run `scripts/audit-verify.sh` after the docs changes (report
    claims + tree drift).
43. status-index regen (`python3 scripts/status-index.py`) — this file
    joins the corpus.
44. Push-stall-watch before ending the session (unpushed daemon
    commits at snapshot).
45. Check the concurrent session's NEXT report for catalog-integrity
    overlap BEFORE building T11 (avoid the R30 double-train).
46. The `.buildflow.yml` + passkey.nix uncommitted edits seen mid-session
    belong to the concurrent session — leave untouched, re-verify at
    next gate.

## g) Questions for the owner (cannot be answered from here)

1. **Production cutover authority (T01):** CI is green and the stack +
   pbx-artmann relock is fully scripted from this host — may THIS
   session run the `nixos-rebuild switch --target-host
   root@pbx.artmann.tech` when the pre-deploy gates pass, or do you
   want to press that button yourself (the plan assumed agent-executed;
   it is still a production cutover)?
2. **v2.15.0 release order:** redeploy main-pin first and cut v2.15.0
   after this train's quality work lands (the plan's assumption), or
   tag v2.15.0 NOW off the green tip and redeploy the tag (the
   go-release memo's minor-readiness argument)?
3. **T11 overlap:** the concurrent session's 20:23 catalog assessment
   §f lists the same ReviewSaved/parity work as my T11 — yield T11 to
   them entirely (I skip to T12+), or split (they take the domain
   catalog row, I take the derived-parity test)?
