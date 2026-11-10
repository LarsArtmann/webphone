# Status: trunk-screened, reports-fix-pending-release

> **Claims verified as of 2026-11-10 16:03 CET** — a point-in-time snapshot; push/CI/deploy statements rot fastest. Session scope: the owner's Reports-panel + failed-Swiss-call report and the diagnosis that followed. No webphone code was changed; two pbx-artmann docs were edited.

## Scope line

**Verified here (fresh commands/artifacts, reproducible from this text alone):**

- Prod version: `GET https://pbx.artmann.tech/version` → `v2.12.0`, commit `dbc36933`, commitDate `2026-11-09T06:55:20Z` — i.e. prod predates today's Reports fix.
- Webphone fix trains on main: `git log --all` shows `cabdeeb8` (fix: parse CDR starts with seconds, 11:49), `abed8247` (honest all-undated state), `3d1c1ba2` (contacts-store naming), `12cee0b3` (NoAnswer stat), `c73f6be3` (changelog/todo); `git branch --contains cabdeeb8` → main; `git log origin/main..HEAD` → 2 UNPUSHED auto-commits (`38fa6dd9`, `c34aec9b`); working tree carries `M internal/push/fixtures.go`, `M internal/server/wirefixtures.go` (not authored here, left untouched).
- `internal/reports/reports.go:24-30` read: `startLayout = "2006-01-02 15:04:05"` with the 2026-11-10 triage comment; TODO_LIST.md:41-64 shows C1–C5 ticked done.
- Stack module: `options.nix:128-131` (`allowInternational` default **true**), `freeswitch.nix:188-190` (toll_allow), `freeswitch.nix:538-565` (`pstn_international` regex `^\+?([1-9]\d{7,14})$` + `lcrBridge` serial failover) — the dialplan gate is NOT what blocks the call.
- Prod host config: `hosts/pbx/default.nix:620-856` — five Telnyx gateways (US/GB/GV/PL/DE), all `register=false` NOREG, all `inviteIdentityHeader=true`, all `callerIdNumber` now `+`E.164; extension 1000 sets no `allowInternational` override.
- **Live trunk probe**: `scripts/trunk-wait-final-probe.sh` run 2026-11-10T15:19:04+01:00 → `INVITE 403`, `Reason: Q.850;cause=21;text="CALL_REJECTED"`, rc=0 (SCREENED); Call-ID `e753hq9mzqej@192.168.1.150`; the response's Remote-Party-ID carries `+17287289311` (indirect proof the +E.164 From train is live). Output parked at `/tmp/trunk-probe-1110.log`.
- Host reachability: `getent hosts pbx.artmann.tech` → `46.62.241.133` (DNS clean, not the dnsblockd block-page); `ssh root@pbx.artmann.tech` → **port 22 connection timed out**; raw `/dev/tcp/46.62.241.133/22` → closed/filtered; `netbird` binary absent on this dev box. The 2026-11-09 CHANGELOG says the ssh deploy legs worked yesterday → **new regression**.
- pbx-artmann TODO rows read: `telnyx-screening-ticket` (§88), `trunk-lift-watch` (§1366 — carries the 01:19Z update: +E.164 switch LIVE, closure `2id2kbrn` at 00:14:26Z, post-switch probe still 403), `siptrace-send-round` (§2083), `waitfinalpostswitch` (struck done, §1927).
- **Edited** (daemon will commit; NOT verified landed yet): `docs/runbooks/telnyx-trunk.md` "Latest verified repro" superseded to the 14:19Z probe with both closures named; `TODO_LIST.md` `trunk-lift-watch` row gained the 14:19Z persistence update + the ssh-filtered fact.

**NOT verified here (and why):**

- CI verdict on today's webphone fix trains (cited "pushed" from git topology only; `gh run list` not run).
- Who authored the 2 unpushed webphone auto-commits / the modified fixture files (inferred "another concurrent session" from AGENTS.md's concurrent-sessions doctrine; unconfirmed).
- Deploy-status of closures `04sm922dd` / `2id2kbrn` (trusted the TODO row's own verification; corroborated indirectly by the probe's RPID; host unreachable for a direct read).
- The trunk-leg From wire bytes — never observed. That is exactly what the pending siptrace capture exists to produce.
- Inbound-call health (the panel shows **0 Incoming**; uninvestigated — out of session scope).
- The fixed Reports panel in a real browser (impossible until prod runs a binary carrying `cabdeeb8`).

## Session narrative (what was asked, what was found)

The owner pasted the live Reports panel (73 calls, "Records without a readable timestamp: 73", "This week — 0", empty By-day, garbled top contact `pifo42h1`, 24 answered / 0 missed) plus a failed-call card (`+41417607701`, "The call failed before it connected") and asked why the panel reads as nonsense and why Swiss numbers still fail. Two answers:

1. **Reports**: every CDR row failed the date parse (FreeSWITCH writes `start_stamp` **with seconds**, webphone parsed minute-only) → all 73 rows counted Undated, temporal views honest-but-empty. **Already fixed today** on webphone main; prod runs yesterday's binary, so the owner is looking at the pre-fix panel. The fix set also carries the honest all-undated rendering, contacts-store naming (kills the `pifo42h1` garble), and a NoAnswer card that will make the 49 failed dials legible.
2. **Swiss call**: nothing Swiss-specific and nothing webphone-side. Every outbound PSTN call dies at the carrier: Telnyx screening answers every authed INVITE from the PBX with `403 Q.850;cause=21 CALL_REJECTED` ~1 s in, pre-call-record, all destinations. Both engineering-prescribed fixes (X-Telnyx-Username header, +E.164 trunk From) are deployed and exhausted; the fresh probe today confirms the screening persists 15 days after first live reproduction (2026-10-26).

## a) FULLY DONE

- **Reports root-cause diagnosis confirmed and attributed** — 73/73 undated rows = `startLayout` minute-only parse vs seconds-carrying `start_stamp`; verified the fix (`cabdeeb8`) is on main and identified prod staleness (`/version` = `dbc36933`, 2026-11-09) as the reason the owner still sees the bug. Evidence: git log + HTTPS /version read.
- **Swiss-call failure root-caused** — not toll_allow (default true), not dialplan regex (matches), not webphone: carrier screening 403. Evidence: dialplan read + live probe rc=0/403/cause=21.
- **Fresh repro evidence produced for ticket #832080** — full paste-ready verdict block with UTC timestamp, Call-ID, Reason cause, RPID; parked at `/tmp/trunk-probe-1110.log`. Cost: zero (a screened probe rings nobody).
- **Runbook superseded per its own convention** — `docs/runbooks/telnyx-trunk.md` "Latest verified repro" now carries the 2026-11-10T14:19:04Z block with both fix closures named and the config-exhausted conclusion.
- **Tracking row updated** — `trunk-lift-watch` (TODO_LIST.md §1366) gained the 14:19Z persistence datapoint and the new ssh-filtered blocker, referencing `siptrace-send-round`.
- **Boundary respected** — did not touch the uncommitted fixture edits in webphone (not mine), did not start a release, did not send external mail (owner act per the 2026-10-28 precedent).

## b) PARTIALLY DONE

- **Telnyx re-engage reply** — the draft is paste-ready in the runbook (§394-442) and now has a fresher repro; the `<E164-RESULT>` slot still says "either" instead of the recorded "still 403 at 00:19Z + 14:19Z", and the required **siptrace capture does not exist** (Telnyx explicitly asked for it; their newest mail is a 4-day-old auto-solve reminder). Blocker: ssh:22 filtered from this box. Effort to finish: S once host access returns.
- **Webphone reports fix delivery** — code complete, tested, pushed (train commits on origin/main; CI verdict unchecked). Remaining: the release dance (fold → bump → gates → stack re-pin → browser E2E → pbx relock → owner switch). Blocker: tree carries another session's uncommitted fixture edits; release is a multi-repo ritual with an owner leg. Effort: L.
- **pbx-artmann doc edits** — written; not verified committed/pushed (daemon owns the commit; the push/CI check convention not yet executed). Effort: S.
- **Swiss presence** — outbound to +41 will work the day the screening lifts (36-country profile includes CH); an inbound +41 lane remains blocked on Swiss-entity regulatory requirements (2026-10-27 research). No action possible from this repo.

## c) NOT STARTED

- **Siptrace capture on the host** (`sofia profile external siptrace on` → one probe → off → journal grep) — the committed next act, blocked on host access. Priority: highest technical act on the ticket.
- **Inbound-call health check** — the panel's `0 Incoming` was noticed and never followed up; inbound has its own 403-burst death class and Telnyx's voice-IP feed phase-in (11-01..15) is inside its risk window. Priority: high (the owner may have a silent second outage).
- **Island dial-failure honesty** — mapping SIP hangup causes to distinct messages (carrier-reject vs timeout vs auth) instead of the single `callFailedUnknown` string (`i18n.js:252`). Planned, never started; blocked only on prioritization.
- **Stale-prod detector** — nothing compares prod `/version` commit vs origin/main; today's 4.5 h fix→user gap went unnoticed. Not started.
- **Scheduled wait-final probe** — lift-day detection is manual; could sit unnoticed for days. Not started.
- **api-dial interim lane wiring** — the working outbound alternative (call-control API, rides the 36-country profile, +49 live-proven) exists for the agent lane only; wiring it as a user-facing stopgap is an unstarted owner decision.

## d) TOTALLY FUCKED UP

- **All outbound PSTN calling is dead** (15th day): Telnyx screening 403s every authed INVITE from the PBX, pre-call-record, all destinations, both prescribed fixes live and insufficient. Severity: blocks the product's core promise (a phone that cannot dial out). Root cause: carrier-side, exact D-code unknown (that is the siptrace ask). Workaround: the api-dial call-control lane (not wired to the phone).
- **The owner experienced an already-fixed bug for ~4.5 hours** — the Reports fix landed on main 11:49; prod still serves v2.12.0 (built 2026-11-09). Severity: trust erosion ("WTF?!?!?!?"); the panel's dishonest state was fixed in code the owner cannot see. Root cause: release cadence + no staleness alarm. Workaround: none for the user; release pending.
- **ssh:22 from the dev box is now filtered** (timed out on name AND raw IP; worked 2026-11-09 per CHANGELOG). Severity: blocks the siptrace capture, the identity-train switch, `push-secrets.sh`, backup pulls — every `--target-host` leg. Root cause: unknown (firewall change not found in the trees read this session). Workaround: none found (netbird absent here; Hetzner console is the owner's path).
- **49 of 73 outbound calls failed with zero in-product explanation** — the island answered every failure with the generic "The call failed before it connected"; the 2026-10-27 owner ruling (an error without a visible surface is a bug) applies to the carrier-reject class. Severity: chronic UX dishonesty during the whole screening saga. Root cause: no cause→message mapping in the island. Workaround: the operator runbook (operator-only).
- **The Telnyx ticket is one silent step from auto-solving** — the newest thread mail is a pending-response reminder; the reply requires evidence only the host can produce, and the host is unreachable. Severity: the only resolution path decays while blocked. Workaround: Hetzner console access.

## e) WHAT WE SHOULD IMPROVE

- **A fix on main is not a fix in prod.** Same-day fixes sat invisible because nothing compares prod `/version` vs origin/main. Fix: a watcher arm (like main-red-watch) that files an issue when prod's commit falls behind main by >N hours.
- **Carrier failures deserve carrier-grade messages.** The island's single unknown-failure string hid a 15-day outage. Fix: map the SIP hangup cause (FreeSWITCH already knows it) to distinct, honest copy; the error-contract doctrine extends to the dial path.
- **Lift-day detection should not depend on someone running a script.** The probe is excellent but manual. Fix: a low-frequency timer (e.g. hourly) that records the verdict and alerts on CHANGED.
- **Evidence parked in /tmp rots.** The runbook references `/tmp/trunk-probe-11xx.log` files that die on reboot; ticket attachments deserve a durable home (e.g. `reports/` or a docs evidence dir) with the /tmp copy as scratch.
- **Host-access loss should be loud.** ssh:22 died silently between two sessions; the deploy legs that depend on it had no canary. Fix: a nightly `ssh -o ConnectTimeout=5 root@pbx.artmann.tech true` health leg.
- **Verify-external-claims applies to other sessions' TODO updates too.** This session trusted the `2id2kbrn` deploy-status row (corroborated indirectly); a direct deploy-status read belongs in the post-ssh checklist.
- **Diagnosis breadth before depth**: a single probe answered the question, but `trunk-debug-matrix.sh` (7 cells) and the home-IP A/B would have produced the full evidence pack in one pass instead of one datapoint per session.

## f) Next things to get done (ranked; harvest ground for docs-health)

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Restore ssh:22 dev-box→PBX reachability (identify the firewall change; NetBird or console fallback) | Critical | S-M | Bug |
| 2 | Run the runbook's siptrace capture (on→probe→off→journal grep) once host is reachable | Critical | S | Bug/diagnosis |
| 3 | Send the re-engage reply to Telnyx (owner act) with both failed-fix results + the siptrace capture | Critical | S | Owner/ticket |
| 4 | Fill the reply draft's `<E164-RESULT>` slot with the recorded 00:19Z + 14:19Z still-403 results | High | S | Documentation |
| 5 | Release webphone carrying the Reports fix (fold → bump v2.13.0 → buildflow/flake/test/smoke gates → release.sh) | Critical | L | Release |
| 6 | Complete the release tail: stack re-pin + browser E2E (served-bytes obligation) + pbx relock + owner switch | Critical | L | Release |
| 7 | Check inbound-call health (0 Incoming anomaly + Telnyx voice-IP feed phase-in window 11-01..15) | High | S | Diagnosis |
| 8 | Re-run the Nov-10 IP-range sweep (`sip.telnyx.com/voice.json` vs `allowedCidrs` pins) — the runbook says ALWAYS pre-switch and the 00:14Z switch already happened | High | S | Ops |
| 9 | Re-run the home-IP A/B discriminator post-both-fixes (freshness for the ticket evidence) | High | S | Diagnosis |
| 10 | Run `scripts/trunk-debug-matrix.sh` (7 cells) for the deep sweep trunk-lift-watch names for oddities | Medium | S | Diagnosis |
| 11 | Map SIP hangup causes to distinct island messages (retire the conflation in `callFailedUnknown`) | High | M | Feature |
| 12 | Wire a stale-prod watcher arm (prod /version commit vs origin/main; file an issue past a threshold) | Medium | M | Tooling |
| 13 | Schedule the wait-final probe (hourly timer; alert on exit 3 CHANGED) | Medium | S | Tooling |
| 14 | Decide the api-dial stopgap: wire call-control as interim outbound dial or formally decline | High | L | Decision/feature |
| 15 | Answer Telnyx's inbound-MMS re-ask (GV-side message ID pull from the sending seat) | Medium | S | Owner/ticket |
| 16 | Verify CI green on today's webphone fix trains (`gh run list` against cabdeeb8..12cee0b3) | Medium | S | Hygiene |
| 17 | Identify the 2 unpushed webphone auto-commits + the modified fixture files (orphaned vs live session) | Medium | S | Hygiene |
| 18 | Verify the pbx-artmann doc edits landed (daemon commit/push + CI) per the push-verification convention | Medium | S | Hygiene |
| 19 | Verify closure `2id2kbrn` deploy-status directly once the host is reachable (retire the indirect proof) | Low | S | Hygiene |
| 20 | Pre-stage the lift-day drill checklist (probe exit 3 → real dial → CDR row → inbound-health green) | High | S | Prep |
| 21 | Identify the `pifo42h1` top-contact (1 call, 1h45m) via CDR + the new contacts-store naming | Low | S | Data |
| 22 | Confirm the 24 Answered calls are all-internal once dated Reports data lands (validates the panel story) | Medium | S | Verification |
| 23 | Add a NoAnswer-by-cause bucket to Reports so the next screening is visible in the tab itself | Medium | M | Feature idea |
| 24 | Pick `reports.weekly_talk_budget` now that dated weeks render (bar currently dead at zero budget) | Low | S | Config |
| 25 | Durable evidence home for probe capture blocks (move /tmp-parked repro evidence into the repo) | Low | S | Cleanup |
| 26 | Run the identity-train owner legs (provisioner restart, hikaru passkey enroll, Super Admin 3-surface check) | Medium | M | Owner/adjacent |
| 27 | Dry-run the siptrace procedure's tooling assumptions (fs_cli present, journal grep pattern matches) | Low | S | Prep |
| 28 | First REAL outbound fax to a fax machine (delivered verdict still unproven; 11-09 train dialed a voice line) | Medium | S | Feature validation |
| 29 | Check the stack browser E2E status for the reports train (served bytes changed; release-blocking obligation) | High | M | Gate |
| 30 | Fold CHANGELOG `[Unreleased]` + sync FEATURES/TODO_LIST as release step 1 | Medium | S | Release |
| 31 | Run the vulnix gate as part of the release (release.sh requirement) | Medium | S | Release gate |
| 32 | Post-deploy verification: prod /version carries the new tag + fresh prod ui-capture of the fixed Reports panel | Medium | S | Verification |
| 33 | Add a lessons entry: "a fix on main is not a fix in prod until /version matches" (the stale-binary class) | Medium | S | Documentation |
| 34 | Owner decision: a +41 inbound lane (Swiss DID needs a Swiss entity per the 10-27 research) | Low | — | Decision |
| 35 | Name the `9100` and `2000`/`*98` report contacts properly through the contacts store (C3 verification on live data) | Low | S | Data |
| 36 | Host-access canary: nightly `ssh -o ConnectTimeout=5 … true` health leg with an alarm | Medium | S | Tooling |
| 37 | Write the trunk config-exhausted state into pbx-artmann AGENTS.md (session-context home, runbook is the detail) | Low | S | Documentation |
| 38 | Consider a denial-reason surface for `pstn_denied` (today it hangs up silently by design; make future gate changes visible) | Low | S | Feature idea |

HARVEST note: items 1–5, 7, 9, 11, 12, 14, 20 belong in TODO_LIST.md; the remainder are ROADMAP-fuel or one-command hygiene. Items 2–4 are already owned by the existing `siptrace-send-round` / `trunk-lift-watch` rows — no duplicates.

## g) Questions I cannot answer myself

1. **ssh:22 from this dev box started timing out today** (name and raw IP; DNS clean; it worked yesterday for the deploy legs). Was filtering this box deliberate hardening (then I route via NetBird/Hetzner console and stop touching ssh), or is it a regression to fix in the firewall config? What I tried: ssh to FQDN and IP, raw /dev/tcp:22, known_hosts check, netbird (absent here), the repo firewall configs read without finding a change.
2. **Do you want the v2.13.0 release started now** (once the in-flight fixture edits clear the tree), or held until the concurrent trains finish? The Reports fix has been user-visible-blocking for ~4.5 h; the release carries multiple trains and ends with an owner switch.
3. **While the screening persists**: should I wire the api-dial call-control lane as the interim outbound dial path (it works, +49-proven, places calls from Telnyx's infrastructure), or does the phone stay dial-less until Telnyx lifts the block?

---

*Report convention: snapshot-rot header per AGENTS.md; scope line per status-report skill. Verified claims cite the command/artifact; gaps are named, not implied.*
