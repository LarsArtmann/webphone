# Pareto execution — solo-mode resume status

> Claims verified as of **07:23 CET 2026-11-24** (push/CI statements rot fastest).
> Session: the owner's "WHOLE TODO LIST" order, resumed after the 04-06
> wave-1 handoff. Six interrupted agents were unresumable (every provider
> lane died), so the session switched to SOLO SERIAL execution and landed
> T06 + T07 fully and T08 ~70% from the clone at `e2d13698`.

## a) FULLY DONE (landed + pushed + verified)

1. **CI truth restored on the whole earlier chain** — `scripts/ci-verdict.sh`
   (T04's pagination fix) verified `586ff894` GREEN (all jobs: smoke,
   reproduce-first, nix build, race suite, catalog export, fast flake
   checks, workflow-e2e), and later `fd07c7f4` + `e2d13698` GREEN the same
   way. The "never verified" debt from the handoff is paid.
2. **Manual-push bar invoked correctly** — `f73c5249` sat unpushed ~2h50m
   (anchor: its 04:07:52+01:00 commit time; the daemon's push leg was dead
   again); pushed with the T06 chain on top; remote tip = local tip
   throughout the session's pushes.
3. **Pass B TODO strikes** (`c7ad8d4e`, in `fd07c7f4`'s push) — the T04
   (`362956ab`), T09 (`78a10215`) and T10 (`e81781fb`) harvest rows struck
   with landing evidence; strike-safety green; battery green in the clone
   before landing.
4. **T06 backlog truth LANDED** (`fd07c7f4`) — all 15 done-but-unstruck
   rows + 1 superseded row from the plan's §5.2 struck with re-verified
   evidence at HEAD (pbx-artmann deploy ledger, ginkgo v2.34.0 at
   go.mod:49, crm master==origin `41883f2e`, stack caFile CHANGELOG:1253 +
   web.nix:134, findings-gate baseline posture, go-idempotency#15 filed,
   create_url ×4, ci.yml webphone-module leg, openapi golden 80 KB,
   runbook train-hygiene :284); twins deduped to ONE homes (Kith e2e → the
   issue-#9 tails row; the drain cluster → the 18-08 row, T18/T26 carry
   it); stale premises retired (00-45 Q1–Q3: Q1 moot by B5's execution,
   Q2 already struck, Q3 stays open for the pack; the b5 one-liner row
   struck as redundant with the AGENTS cqrs-lint directive convention);
   `docs/status/INDEX.md` regenerated (the 04-06 report had landed
   without its index row — first daemon omission of the day).
5. **T07 own-defects reconciliation LANDED** (`9546b9e5` — the daemon
   swept the working-tree edits; content verified byte-exact afterwards —
   plus `e2d13698`'s TODO strike): vulnerability id `GO-2026-6859 /
   CVE-2026-97054` at both index.html sites (AGENTS:12 was `586ff894`);
   `RegisterStdlibDefaults` moved from the go-health row to the
   go-error-family row; Strengths counts re-derived (8 → **14** keyed rate
   limiters, 11 → **14** `metaengine.Query` declarations); the dead
   `metaengine/cap.go` verify path fixed to `metaengine/v4/cap.go`; the
   stale 2026-11-13 date label → 2026-11-22; Findings card 3 → FULLY
   RULED (feedback.Alert + panelHead declines added); card 5 → go-etag
   entitytag + server at DIRECT v0.6.2.
6. **The 06-50 report's deferred full suite DISCHARGED** (`e2d13698`) —
   the concurrent dedup session's §f 22 (c-6) deferral was executed by
   this session: `go test -count=1 ./...` green over the
   extraction-bearing tree (sole non-ok line = the no-test-files notice),
   island node specs 0 failed/skipped, `tsc -p . --noEmit` clean; the
   report item annotated resolved in place (its token table is 6/6
   generic — `c-6`/`count=1` are not tokens — so the harvest matcher can
   never bind it); the T07 TODO row struck; INDEX regenerated for 06-50
   (second daemon omission).

## b) PARTIALLY DONE (in flight)

1. **T08 dependency train — ~70%** (clone `/tmp/wp-clone-t08`, UNCOMMITTED):
   - m08.1 ✅ bump-preflight ×5 — no removed exported symbols (root adds
     `RecommendedSecurityMiddlewareWithCSP`/`CSPBuilder`, additive).
   - m08.2 ✅ `go get` + tidy landed the five named pins (cqrs-htmx root
     v4.18.1→v4.19.0, usermgmt v4.20.0→v4.21.0, loginpage
     v4.18.0→v4.19.0, identity-model v4.17.0→v4.18.0, go-branded-id
     v0.7.1→v0.8.0) PLUS two pre-adjudicated ride-alongs the graph forced:
     templ-components v1.27.0→v1.27.2 (the standing "ride the next templ
     train" ruling — this IS that train) and go-datastar/static
     v0.6.1→v0.8.0 (indirect).
   - m08.3 ✅ templ generate (CLI pinned to go.mod's v0.3.1118): 0 view
     changes; `nix fmt` 0 changed.
   - m08.4 ✅ build OK; session-behavior suite (`-run TestServer`) OK.
   - m08.5 ✅ after one fix (see d-1): full suite 37 pkgs ok, 0 FAIL.
   - m08.6 ⏳ vendorHash re-pin — NOT DONE.
   - m08.7 ⏳ buildflow + `nix flake check --no-build` — NOT DONE.
   - m08.8 ⏳ `auth:` commit + push + CI watch — NOT DONE.

## c) NOT STARTED (unchanged from the plan)

T16 (stack window A), T20 (announcements v2.9.0→v2.15.0 draft), T21
(android slice 1), T05 (audit-verify watch lane), T11/T12 (report
read-throughs), T13 (index polish + vision QA), T14
(RegisterStdlibDefaults adopt-or-decline), T15 (lessons + SKILLS
runbook), T17 (stack window B), T18/T26 (report drains), T19 (stack
residues), T25 (chat-network stack slice), T27 (gated register sweep,
LAST). Gated on the owner: T02, T22/T23, T24 — see g).

## d) TOTALLY FUCKED UP (honest list)

1. **templ-components v1.27.2 broke the tw.css gate** — the button
   component switched to semantic `accent-*` classes while webphone's
   `@theme` remap only defined `blue-*`; first rebuild still red because
   the accent family had no theme definition; fixed by adding the
   accent-100/300/400/500/600/800/900 mappings (same signal-green values
   as the blue remap) to BOTH `tw.css.input` AND `login.css.input` — the
   build script's one-home-per-remap guard caught the missing mirror on
   the first attempt.
2. **My own evidence tails flipped archived audit items over the
   generic_df boundary TWICE** (`disposition` +3 and `brand` +1; later
   `battery` ×3 in the discharge row) — each cost a full battery cycle to
   diagnose via audit-list diffs and reword. The DF-brittle ratchet pain
   is real and bites the WRITER, not just the corpus.
3. **The zai "no model move needed" premise for alejandra-t16 was
   FALSE** — the plain-continue resume rate-limited immediately; the
   model move to synthetic then worked for all three resumed agents
   before synthetic itself died.
4. **All four provider lanes are now dead for subagents** (zai
   rate-limits under concurrency, hyper + huggingface-fireworks no
   credits, synthetic subscription exhausted) — six resumed agents lost
   their reports; the session fell back to solo serial execution.
5. **The daemon swept my in-flight index.html edits** (`9546b9e5`) before
   my gates/narrative commit ran — I worked in the REAL repo for T07
   against my own clone protocol; verified the swept bytes matched my
   intended state, but the narrative commit was lost to the heuristic.
6. **The daemon landed TWO reports without INDEX rows today** (04-06 and
   06-50) — both healed by this session; the omission pattern is live.
7. **Wrong emoji codepoint** — `\U0001F7A1` (🞡) instead of `\U0001F7E1`
   (🟡) in the go-error-family row; caught on visual diff, fixed.
8. **A wrong handoff premise**: the t06 clone sat at `78a10215`, not the
   claimed clean base `94157138` — re-verified before trusting it (the
   discipline worked; the handoff summary was stale).

## e) WHAT WE SHOULD IMPROVE

1. **Solo serial beats the agent lottery** under provider exhaustion —
   deterministic, context-rich, no dispatch/recovery overhead. Keep
   agents only when a lane is PROVEN alive that hour.
2. **Work EVERYTHING in clones** — the daemon commits any dirty real-repo
   file within minutes; the clone protocol is the only safe workspace
   for multi-step edits.
3. **Pre-flight evidence tails** (`docs-gates freshness . --pre-flight` /
   `--margins`) BEFORE writing them into TODO rows — both boundary flips
   today were detectable in advance.
4. **Annotate-first for token-poor report items** — when `--explain-item`
   shows all-generic tokens, skip the harvest row entirely and annotate
   the item resolved (the row alone can never green the gate).
5. **INDEX regen belongs to the landing session** — daemon heuristic
   commits don't run scripts; every human-landed report should carry its
   INDEX row in the same commit.
6. **The `--verify` output of status-items is load-bearing** — read the
   full matched-line listing before writing verdicts (the skill's
   documented mis-strike incidents matched today's shape exactly).

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Landed this session (struck here so the ledger stays honest):

1. ~~Verify CI on `586ff894`~~ → done 2026-11-24 — GREEN (all jobs, workflow-verified).
2. ~~Manual-push bar: land `f73c5249`+chain~~ → done 2026-11-24 — pushed; tips verified.
3. ~~Pass B strikes (T04/T09/T10 rows)~~ → done 2026-11-24 — `c7ad8d4e`.
4. ~~T06 backlog truth (15+1 strikes, dedupes, stale premises)~~ → done 2026-11-24 — `fd07c7f4`.
5. ~~INDEX regen for the 04-06 report~~ → done 2026-11-24 — `fd07c7f4`.
6. ~~T07 index reconciliation (vuln id ×2, RegisterStdlibDefaults move,
   14-limit/14-Query counts, `metaengine/v4/cap.go`, FULLY-RULED card,
   etag v0.6.2)~~ → done 2026-11-24 — `9546b9e5`.
7. ~~T07 TODO row strike~~ → done 2026-11-24 — `e2d13698`.
8. ~~Discharge 06-50 §f 22 (c-6) with the green full-suite run~~ → done 2026-11-24 — `e2d13698`.
9. ~~INDEX regen for the 06-50 report~~ → done 2026-11-24 — `e2d13698`.
10. ~~CI GREEN verification on `fd07c7f4` + `e2d13698`~~ → done 2026-11-24 — workflow-verified GREEN.

Open, in execution order (each binds its TODO_LIST row):

11. **T08 tail (m08.6–m08.8): the vendorHash re-pin in ONE command,
    then buildflow + `nix flake check --no-build` (re-pin goldens), then
    the `auth:` commit + push + CI watch** — the whole tail of the
    cqrs-htmx root v4.19.0 / usermgmt v4.21.0 / loginpage v4.19.0 /
    identity-model v4.18.0 / go-branded-id v0.8.0 train (TODO row:
    "cqrs-htmx root v4.19.0 … vendorHash in ONE command, `auth:` commit,
    CI watch"; tokens cqrs-htmx v4.19.0 + vendorHash bind it).
12. **T16 stack window A**: the styled-404 shape rule in the stack's
    `tests/browser-e2e.py` + the dashboard-on VM leg + wire fixtures a+b
    (rows L1840/L1646; tokens: browser-e2e ×4, styled-404).
13. **T20 announcements**: one consolidated draft v2.9.0→v2.15.0 in the
    customer's voice; owner picks the channel (row L1778; tokens:
    announcements ×3).
14. **T21 android slice 1**: `lines[]` state, build + wire tests, the
    SIP-UA-vs-FreeSWITCH smoke checklist, origin-picker first screen,
    CallStyle wiring check (row L1865; tokens: android ×17).
15. **T05 red-window detection**: `audit-verify.sh` into the watch lane
    - tip-unchanged skip + the pma detection-lane ask (row 00-52 §f 15;
      tokens: tip-unchanged).
16. **T11 report read-through A** (11 reports incl. cqrs-htmx, the
    14-row go-cqrs-lite frame, templ-components) + provenance lines.
17. **T12 report read-through B** (14 reports) + the 25-family latest
    re-derivation with exact refspecs.
18. **T13 index polish**: one-line verdict cells, the go-cqrs-lite
    pin/latest pair, summary split, trace stylesheet, `index-render.jpg`
    - the vision-model check (row 00-52 §f 19–21).
19. **T14 `RegisterStdlibDefaults`**: read the 9 prod
    `RegisterClassifications` sites, decide adopt-or-decline, write it
    into go-error-family row 6.
20. **T15 lessons batch**: docs/lessons.md ×6 (awk-mangle tag query, the
    fleet-tags v-prefix hole, the ci-verdict page-one blind spot,
    relayed identifiers, the overlay negative-proof) + AGENTS one-liners
    - the SKILLS runbook (tokens: tip-unchanged, overlay).
21. **T17 stack window B**: `fs-directory` in operatorWatchedUnits, the
    secretFile fixture, the answer-count helper, the sipcred HMAC vector
    (row L1691).
22. **T18 report drain slice 1**: the two big 2026-11-10 reports (38+27
    rows) gates-first, ratchets re-measured DOWN (the ONE-home drain row).
23. **T19 stack residues slice 1**: the paperless module option +
    module-check case + the deploy.md secret column + demo-call recipe
    (row L1835; tokens: deploy.md).
24. **T25 chat-network stack slice 1**: the signal-cli unit + QR
    bootstrap + module-check + the matrix homeserver memo (row L1553;
    tokens: signal-cli ×2).
25. **T26 report drain slice 2**: the 2026-11-19…11-22 cohort (8
    reports) gates-first.
26. **T27 gated register sweep LAST**: one pass over every gated row +
    the upstream-tick drafts (go-structure-linter #13, cqrs-lint P012,
    the go-sse `[0.7.0]` note).
27. **Trash the stale /tmp clones** (`wp-clone-t03/-t06/-batch*/-report*`
    and the dead agents' dirs) with `trash` as tasks complete.
28. **CI watch after the T08 push** — the bumped module graph needs a
    green CI before any stack re-pin.
29. **Re-measure audit_max at the next sweep** (74 today; the corpus
    grew twice today via daemon-landed reports).
30. **Tracker issue for the daemon INDEX-omission pattern** (two today:
    04-06, 06-50) if a third lands.
31. **Tracker issue for the daemon push-leg stalls** (two nights
    running; the manual-push bar fired both times).
32. **The vendored-CHANGELOG blind spot** in fleet-tags (composite/wave
    headings naming submodule tags are not machine-swept — the T09
    residual).
33. **Fresh stack clone for T16** (`/tmp/stack-clone-t16` state unknown
    after the agent deaths — re-clone).

Gated on the owner (see g): 36. T02 post-sitting conversion; 37. T22 D2
veto window; 38. T23 D3 agent liveness slice 1; 39. T24 prod
verification window.

Standing maintenance (from tonight's observations):

39. The loginpage v4.19.0 ExtraSection contract re-check after the bump
    (the passkey/login surfaces ride it — the session suite covers it,
    but the smoke's login checks should re-run before the push).
40. `docs/watches/2026-11-08_standing-watch-calendar.md`: the Go 1.27.4
    roll row re-check (`nix eval nixpkgs#go_1_27.version` still 1.27.3).
41. The erraudit/branching-flow monthly re-measure rides the watch
    calendar (next due 2026-12-08).
42. The `grep -c "battery"`-class token margins: run
    `docs-gates freshness . --margins` before the next dense-docs train.
43. The `/tmp/go-appkit` symlink recipe for audit-verify in clones —
    note it in T15's runbook half.
44. The `test -d` guard for clone reuse in the landing recipe (the t06
    stale-base surprise).
45. Consider a `status-index.py --check` leg in CI's fast-checks so a
    daemon-committed report cannot red main silently (two omissions
    today were caught by hand).
46. The templ-components v1.27.2 accent swap: FEATURES/CHANGELOG line
    rides T08's m08.8 commit (the bump bullet).
47. After T08 lands: `scripts/fleet-tags.sh` re-sweep (the four new
    upstream tags must read as tagged).
48. The concurrent dedup session's remaining §f items (06-50 report)
    are theirs to harvest — leave alone unless still red at the next
    battery.
49. `git config core.hooksPath .githooks` sanity on any fresh clone used
    for the T08 commit (the hook gate runs on MY commits).

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Provider budget**: all four subagent lanes are dead tonight (zai
   rate-limits under concurrency, hyper and huggingface-fireworks want
   pre-paid credits, synthetic's subscription is exhausted). Do you want
   to top up credits (which provider?) so the remaining ~13 tasks can
   run 2–3 agents in parallel, or should I continue SOLO SERIAL (slower
   wall-clock, zero spend, working fine so far)?
2. **D2/D3 GO**: T22 (veto window) and T23 (agent liveness slice 1) stay
   gated on your explicit GO per the standing ruling — both would build
   behind OFF-by-default flags only. GO or defer to after the sitting?
3. **Capture credentials for T24**: provide
   `WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD` into this
   session's env only (never written to the repo) for the prod
   verification window — or park T24 until the sitting?
