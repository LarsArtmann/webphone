# Status Report — T14/T15/T16/T20 executed: owner decisions recorded, sanitization aligned, own-number identity shipped

**Written:** 2026-09-20 11:39 CEST
**Session scope:** the four pasted TODO rows (island sanitization, own-number
visibility, owner decisions, standing watches) — mapped to the 10:14 SUPERB
pareto plan as T14 (decision batch), T15 (sanitization alignment), T16
(own-number DID surface), T20 (watches cadence). This report covers THIS
session's run only; no fresh research beyond it.

**Format override (flagged):** the status-report skill's canonical output is a
styled HTML dashboard; the user explicitly requested `.md`, which wins. The
brutal-self-review skill's content (what we got wrong) is folded into sections
d/e here instead of a second file at `docs/reviews/`.

**Verdict up front:** all four items executed, all gates green (go test,
flake check incl. island-lint + backup VM test, smoke 30/30, buildflow
EXIT=0, stack browser E2E EXIT=0 against the local tree). Nothing user-facing
broke. The failures this session were process failures (pipe-masked gate
verdict, edit-before-read round trips, a mis-typed test, sloppy ripgrep
flags) — enumerated honestly in d/e.

---

## a) FULLY DONE

1. **T14 — the four owner one-liners decided and recorded.** New AGENTS.md
   section "Owner decisions (2026-09-20)": (1) stack `webphone` input rides
   `main` with a per-train lock bump, revisit trigger named (sub-hour hotfix
   need); (2) pbx-artmann keeps `path:` while all three repos are co-located —
   explicitly superseding the earlier github-input recommendations in the
   15:37 SUPERB plan P8 and the 15:09 report, with the rationale that the
   2026-09-19 "burn" was an uncommitted tree, not the input type; (3)
   sanitization side = island keeps letters; (4) own-number feed = static
   config `identities` map now, stack `/phone-api` identity endpoint as the
   upgrade path, CDR-derive REJECTED on evidence (the stack's `api.py`
   reshapes FreeSWITCH `Master.csv`, where outbound `caller_id_number` is
   dialplan-dependent — webphone's own fixtures show the extension `1001`,
   not a DID — and it is absent before the first call).
2. **ROADMAP reconciliation.** The stack input-policy open question marked
   DECIDED with a pointer to AGENTS; no longer an owner call.
3. **TODO_LIST reconciliation.** The three done rows deleted (docs-health
   rule: done work is deleted, never struck); the sweep line updated with
   T14/T15/T16/T20.
4. **T15 — sanitization aligned end to end.** All three island sites
   (`calls.js` dial, `calls.js` blindTransfer, `panels.js` saveContact) now
   use `[^\d+*#a-zA-Z]/g`, matching `sanitizeDialable` exactly. Log line and
   `nothingDialable` (en/de) mention letters. The `ids.go` parity comment is
   TRUE again (it had been a corrected lie since the 19:18 review). Both
   sides pinned: served-asset grep rows for `calls.js` + `panels.js` in
   `TestStaticAssetsServe`, plus the existing Go-side table test.
5. **T16 — own-number identity shipped and tested.** New `identities` config
   map (ext → presented DID), validated at boot (keys must be normalized
   extensions so the signed-in lookup cannot silently miss; values must have
   dialable characters). Surfaces: signed-in shell header (`1001 · +49 …`),
   `POST /api/session` response gains `did`, island whoami line appends it
   via the `wp:session-opened` event detail, messages (thread list + open
   thread) and fax composers show "sending as …". Session-scoped and
   display-only — never rendered into the unauthenticated `/config.js`.
   Unmapped extensions get none of it (absence stays silent, pinned by the
   negative half of `TestIdentitySurfacesOwnNumber`). i18n keys in BOTH maps
   (`identity.from`), app.css styles, README key table + example +
   maps-sentence, FEATURES row, CHANGELOG [Unreleased] Added + Changed
   entries.
6. **T20 — standing watches converted to a dated quarterly re-check** (next
   due 2026-12-20) with named per-watch triggers. nanoid CLOSED claim
   verified against go.mod (v1.65.1); ThemeScript opt-out re-verified ABSENT
   at the consumed templ-components v1.18.0 tag (`base_templ.go:491` renders
   `ThemeScript` unconditionally) — the watch's "take the knob when it
   ships" instruction stands.
7. **Verification gates, all green with proof:**
   - `go test -count=1 ./...` — all packages ok (run twice during the session).
   - `nix fmt` — 4 files normalized (all benign: alignment, prettier wraps).
   - `nix flake check` — all checks passed, including island-lint, treefmt,
     the sandbox package tests, and the kvm-gated backup VM test.
   - `webphone-smoke.py` (dev shell) — 30 passed, 0 failed.
   - `buildflow` — EXIT=0, "52 success, 0 failed" (second run; see d/1).
   - Stack browser E2E (`.#telephony-browser` with
     `--override-input webphone <local tree>`) — EXIT=0 on a re-run with the
     exit code captured, result symlink built, `E2E-OK` marker found, zero
     error lines. This is the markup gate for BOTH island changes.
   - Closing sweep: the E2E's leftover host server (`/tmp/wpshoot/
     webphone-bin`, port 18099) was killed (SIGTERM then SIGKILL) and the
     port proven free; `pgrep` clean.

## b) PARTIALLY DONE

1. **"Compose prefill" interpretation.** T16's wording said "compose
   prefill"; I shipped a passive "sending as" identity line above the
   composers (there is no from-input to prefill). The choice is defensible
   and documented in CHANGELOG, but the interpretation was never confirmed
   against the owner's intent — question g/2.
2. **Own-number feature's real-world value.** Code is done and tested, but
   prod has NO `identities` configured (and v2.4.0 is not deployed yet), so
   the user-visible answer to "what is my number?" exists only in test and
   loopback fixtures today. The feature is inert until the owner fills real
   DIDs — which I cannot know (question g/1).
3. **E2E wall-time budget.** Today's gate run took 172.82s vs the 151s
   baseline (+21s). The watch's trigger is TWO consecutive over-budget runs;
   only one data point exists, and the datum was recorded into the watches
   row after the fact (retroactively, not at observation time — see e/6).
4. **Standing watches themselves.** The re-check SCHEDULE exists; the actual
   watch work (sip.js 0.22 existence scan, templ-components knob check on
   each bump, oxlint globals hygiene) is now dated work due 2026-12-20, not
   done today beyond the two verifications noted above.
5. **pbx-artmann decision propagation.** DECIDED (keep `path:`) is recorded
   in AGENTS, but the two superseded recommendation documents (P8, 15:09
   report) are not annotated with DECIDED pointers — a docs split-brain
   risk until the T17 ANNOTATE pass covers them.
6. **HARVEST closure.** Per the status-report skill, section (f) belongs in
   TODO_LIST/ROADMAP, not entombed here. The pre-existing remaining rows are
   already in TODO_LIST; the session-derived NEW items below (f/16–f/24
   especially) await an explicit HARVEST run.

## c) NOT STARTED (surrounding plan work this session did not own)

1. T1 — redeploy prod with v2.4.0 (the security exposure — sessions minted
   without credential verification — is STILL LIVE on prod).
2. T2 — restore the prod outbound SMS lane (root cause stack-side).
3. T4/T5/T6 — 1001 registration anomaly: sofia dump instrumentation → island
   rebuild-on-`Unregistered` fix → two green E2E runs.
4. T7 — release.sh step-8 ELF machine-type guard.
5. T8 — vulnix triage extraction + fixture test.
6. T9 — styled-404 anonymous smoke check (foreign-mode safe).
7. T10 — module csrf conflict pin (typed AND raw simultaneously).
8. T11 — styled-404 i18n decision.
9. T12 — stack-side csrf rendered-settings assertion.
10. T13 — stack uncommitted-tree reconciliation (owner keep/discard).
11. T17 — docs-health ANNOTATE over the docs/status range.
12. T18 — v2.1–v2.3.0 announcements (channels + disclosure posture).
13. T19 — "on flake, read the transfer_dbg dumps first" ritual into AGENTS.
14. Roadmap owner calls: XFF-sanitization answer (rate-limit key flip), HSTS
    on prod, GitHub Release objects policy, `/v2` module-path policy,
    loopback `delivered` semantics, handler-gating dual layer, recordings UI.

## d) TOTALLY FUCKED UP

Nothing shipped broke, no data was lost, and every gate is green — but these
process failures happened and deserve the blunt label:

1. **Pipe-masked gate verdict (the AGENTS trap, repeated).** The first
   buildflow run was piped through `tail -20`, which hid its exit code — the
   exact "pipeline masking" lesson already in AGENTS cross-cutting lessons.
   I caught it, but only after presenting the tail as if it were a verdict;
   the honest rerun (`> log; echo EXIT=$?`) is what produced EXIT=0. A green
   banner I almost trusted without proof.
2. **Edit-before-View, four times.** calls.js, config_test.go, README.md,
   CHANGELOG/FEATURES — each rejected with "you must read the file first"
   because I had read them via bash (`sed`/`cat`) instead of the View tool.
   The rule is written down; I re-violated it after the first rejection
   instead of internalizing it. Four wasted round trips.
3. **Test written against the wrong types.** `TestIdentitySurfacesOwnNumber`
   first failed to compile with six `strings.Contains`-on-`[]byte` errors —
   the adjacent tests in the same file visibly used `string(body)`. One more
   cycle spent.
4. **Unreliable grep flags.** I used `rg -rn` twice: in ripgrep `-r` is the
   REPLACE flag, so some displayed lines may have shown substituted text
   rather than file contents (and `--include` is grep's flag, not rg's). No
   decision in this session rested on a distorted line, but the reads were
   untrustworthy and I shouldn't have to say that.
5. **Stray tool call.** A `read_mcp_resource` invocation with a missing
   parameter at session start — pure noise, a fat-fingered parallel call.
6. **Process-proof fumbling.** The first leftover-server check reported
   "smoke processes ALIVE" ambiguously; the shell's `kill` builtin is
   unsupported, so termination took pkill + SIGTERM + SIGKILL across three
   attempts before `pgrep`/port checks proved the process gone. The
   closing-sweep rule was eventually satisfied, but clumsily.

## e) WHAT WE SHOULD IMPROVE

1. Treat the edit tool's View-first contract as a hard sequence, not a
   suggestion to rediscover per file.
2. Gate invocations NEVER go through `| tail`/`| grep` filters — always
   `cmd > log 2>&1; echo EXIT=$?` first, then inspect the log. AGENTS
   documents this trap; it now has a second in-session instance.
3. Verify CLI flag semantics (`rg -r`, `--include`) before trusting output;
   grep habits do not transfer to ripgrep.
4. When a DECISION supersedes an earlier written recommendation, annotate
   the superseded document immediately (or fold it into the T17 ANNOTATE
   scope explicitly) — otherwise two documents contradict and only one knows.
5. Plan wording that admits interpretation ("prefill") should be resolved as
   an EXPLICIT recorded assumption, not silently engineered.
6. Record measurements at the moment of observation (the 172.8s datum went
   into the TODO row only during this report).
7. Config-backed features should ship with at least one live exercising
   surface (smoke or E2E fixture) so the feature is proven over real HTTP,
   not only Go tests — identities currently has Go-test coverage only.
8. Prefer explicit per-task commits over the daemon's heuristic ones when
   work is authorized; today's history is again "auto-commit N files" noise
   that a future `git log` reader cannot triage.
9. Give the new user-facing copy a native-ear pass ("sending as" /
   "Senden als") before the next train.
10. Consider closing the no-reload parity gap: after island login the shell
    header still shows no DID until the next full page load (mirrors today's
    extension behavior, but the `wp:session-opened` event could update both).

**Self-review addendum (brutal-self-review questions):** Forgot: recording
the E2E datum immediately; annotating superseded recommendations. Ghost
systems: none created — every new surface is wired, rendered, and test-pinned.
Split brains: the dial alphabet is now pinned in three places (ids.go
comment, ids_test comment, two served-asset greps) — acceptable as deliberate
redundant pinning, but the JS literal still lives in three island sites; a
shared island helper would give it ONE home (deferred: any island refactor
costs a full E2E gate run, and the Verschlimmbesserung guard favored
additive-only this session). Lies: none — every "green" claim above carries
its command and exit code. Scope creep: held; the stack repo and
pbx-artmann were deliberately NOT touched. Removed something useful: no —
the only removals were three TODO rows whose work now exists in code and
docs.

## f) Up to 50 things we should get done next

Brainstorm per the skill's rule — most items beyond the first block are
TODO_LIST/ROADMAP fuel, not commitments.

**Immediate (plan-gated, highest impact):**
1. T1: deploy v2.4.0 to prod (owner ssh) — the unverified-session security
   exposure is still live; bogus-creds probe must be green before `switch`.
2. T2: restore the prod SMS lane (journalctl triage → fix → send/receive test).
3. T13: reconcile the stack's uncommitted `flake.lock` + `operator.js`.
4. T3: train-cut decision — v2.5.0 now vs hold per g2 cadence.
5. Populate real `identities` on prod (needs g/1 data) as part of the deploy.
6. Second stack browser E2E run to pair with today's 172.8s (budget trigger
   needs two consecutive over-budget points).
7. If the second run confirms over-budget: investigate the +21s delta;
   else re-baseline the budget.

**Own-number identity follow-through:**
8. Stack `/phone-api` identity endpoint (the DECIDED upgrade path; kills
   config duplication and covers extensions without manual map upkeep).
9. Typed NixOS module option for `identities` (+ flake-check assertion)
   instead of relying on freeform `settings` pass-through.
10. Client-side header DID update on `wp:session-opened` (no-reload parity).
11. Smoke: an identities-configured scenario exercising header, whoami and
    composer lines over real HTTP.
12. Stack browser E2E: configure one identity and assert the whoami/header
    DID in-browser.
13. Annotate the superseded P8/15:09 recommendations with DECIDED pointers
    (fold into T17 scope).
14. README NixOS-module section: one sentence on `identities` pass-through.
15. ROADMAP fuel: contact-name resolution next to the DID in composers;
    header tooltip i18n ("your number" / "Ihre Nummer").

**Remaining plan tasks:**
16. T4: sofia registration dump instrumentation in the E2E reconnect phase.
17. T5: island rebuild-on-`Unregistered` fix (asset tripwire test first).
18. T6: E2E ×2 green with recorded wall times (pairs with 6/7).
19. T7: release.sh ELF machine-type guard (`od -j18 -N2` = `b7 00`).
20. T8: extract vulnix triage into `scripts/vulnix-triage.sh` + fixture test.
21. T9: styled-404 anonymous smoke check (foreign-mode safe).
22. T10: module csrf conflict pin (typed AND raw set) + README precedence.
23. T11: styled-404 i18n decision (en/de keys or recorded English-only line).
24. T12: stack-side csrf rendered-settings assertion in its VM test.
25. T17: docs-health ANNOTATE over docs/status (owner confirms range).
26. T18: post the v2.1–v2.3.0 announcements (channels + disclosure posture).
27. T19: write the flake-analysis ritual ("read transfer_dbg dumps first")
    into AGENTS.

**Roadmap owner calls:**
28. XFF sanitization answer → flip `remoteHostKey` to
    `KeyExtractorFromClientIP`.
29. HSTS on `pbx.artmann.tech` decision.
30. GitHub Release objects vs tags-only policy (backfill v2.0.0 object?).
31. Go `/v2` module-path policy record.
32. Loopback gateway `delivered` vs honest `sent` semantics.
33. Handler gating dual layer: keep or drop the middleware wiring.
34. Recordings UI product intent (consent/jurisdiction posture).

**Sanitization/identity polish:**
35. Shared island sanitize helper (one regex home in JS); costs an E2E gate
    run — batch with another island change.
36. Property/fuzz test over a shared fixture table for the Go↔island
    alphabet (full-skills-sweep item 21).
37. `inputmode` UX decision: login ext input is `inputmode="numeric"` while
    letters are now legal dialables.
38. whoami DID tooltip/title with en/de i18n keys.
39. CHANGELOG wording pass on the two new entries before the next train.

**Testing/debt:**
40. `-race` stress of the SSE hubs.
41. SSE handler edge tests (anonymous 401, heartbeat on the wire).
42. Store/domain edge-case round-out (contacts CRUD, branded-ID round-trips).
43. Delivery-receipt SSE-event assertion.
44. Headless console-cleanliness gate (the class that let CSP violations ship).
45. ROADMAP tail refinement on demand: gzip, retentionDays, startupz wiring,
    session persistence, PWA.

**Process/hygiene:**
46. HARVEST this report's (f) list into TODO_LIST/ROADMAP (docs-health).
47. Explicit commit-per-task when authorized (readable history beats daemon
    heuristics).
48. Institutionalize the process-death proof inside `webphone-smoke.py`
    teardown (assert the booted binary is gone, not just the check count).
49. Clean stale artifact dirs on this host (`/tmp/wpshoot`,
    `/tmp/webphone-smoke-*`) — owner machine hygiene.
50. When pbx-artmann next relocks, verify the `path:` tree-cleanliness
    precondition is still documented at the relock site, not only in AGENTS.

## g) Questions I cannot figure out myself

1. **Which real DID belongs to which extension on prod** (and preferred
   display format: E.164 or spaced)? No repo contains this data; without it
   the identities feature stays inert. And should the values ship via
   pbx-artmann's config.json or the module's typed settings once option 9
   exists?
2. **What did "compose prefill" mean to you?** I shipped a passive "sending
   as" line above the message and fax composers (no from-input exists to
   prefill). If you intended something active — a visible/hidden from-field,
   a dial prefix, or per-thread identity selection — say so and I'll rework.
3. **Do you confirm keeping pbx-artmann's `path:` input?** My DECIDED line
   supersedes two earlier written recommendations (P8 and the 15:09 report
   both recommended github input). Confirm `path:` stays, or ask for the
   github-input migration and I'll rework the AGENTS tri-repo contract and
   the relock runbook accordingly.

---

**Waiting for instructions.** Per the status-report skill, section (f) is the
input for a docs-health HARVEST pass; on your word I will route the new items
into TODO_LIST/ROADMAP.
