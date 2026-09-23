# Status 2026-09-22 15:06 — the whole-list execution: v2.5.0 shipped, code trains landed, gates RED mid-flight

Session start ~13:00 CEST, report cut 15:06 CEST. Directive: execute the
ENTIRE SUPERB 20-year plan (`docs/planning/2026-09-22_12-02_SUPERB-20-year-durability.md`)
in order, verify each step. This report covers only what this session
did and noticed.

## The one-paragraph truth

> CLOSED 2026-09-23 (docs-health): everything in (b)/(c) either
> shipped the same day (T16 docs kernel `c53ee81`; T18a/b/c retention;
> T20a-e affordances; T21a-d failed-bubble/thumbnails/drafts; T22 DOM
> contract; T23 PWA verdict; T25 retention; T26a-e platform tail;
> T27a-d platform tail 2 — 16:52/18:20/18:46/19:55 reports) or rode
> v2.6.0 (signed tag `807ca0c`, 2026-09-23). The RED gates of (d1)
> went green on the quiesced tree (19:55 gate scoreboard; the 02:47
> release gates re-ran everything green on the tagged tree).
> Owner-terminal items live as TODO_LIST rows (deploy, post-deploy
> probes, SMS lane, announcements, owner-calls batch).

The release machine worked end to end: **v2.5.0 is cut, gated, tagged,
published** (gh object live), the stack is relocked to it in-train
(browser E2E + VM test + full stack flake check green), pbx-artmann is
relocked to the tag chain with both-arch toplevels green — the owner
deploy is ONE command away. Before that, the session discovered prod
was NEVER in the emergency state the plan assumed (smoke `--base` 16/0:
bogus credentials rejected, `/version` v2.4.0 — the v2.1.0-hole premise
was stale). Four code trains landed on main after the release
(T12 contacts API, T13 SSE contacts nudge, T14 verdict+tests, T17 island
diagnostics tail) plus a repo-wide `GOEXPERIMENT=jsonv2` removal sweep —
**and the post-sweep buildflow + flake check are RED** (T12 gofmt drift
in the constructor alignment + a govulncheck toolchain mismatch,
possibly sweep fallout, possibly my mid-run edits racing the gates).
That red is the live fire right now.

## a) FULLY DONE (verified this session)

1. **Plan session debts** — mermaid render-verified via mmdc (48 KB
   SVG), fine table re-split to literal ≤12-minute units, "enourmous"
   typo + coverage recount (10 TODO rows incl. T10).
2. **T1 chain published + validated** — webphone main pushed through
   `5810208`; stack flake check GREEN twice (run #2 on the pre-bump
   tree, run #3 EXIT=0 on pinned `1a95a736` INCLUDING the daemon's
   overnight nixpkgs bump, which had also reintroduced operator.js
   formatting drift — re-fixed in `1a95a73` and pushed).
3. **T2 + T2' (tag relock)** — pbx-artmann relocked twice (`f42ef79`,
   then `36ae250` to the v2.5.0 stack rev `91ac2c82`): drift probe
   all-OK both times (narHash + sibling rev + webphone pin = the tag
   commit, no float), both-arch toplevels EXIT=0 both times, webphone
   store path moved `mxwzna36…` → `7az7rhn…` → `lq5fj47…` (**webphone-2.5.0**
   in the deploy closure). Pushed, verified via ls-remote.
4. **Prod premise CORRECTED** — smoke against prod: 16 passed / 0
   failed (bogus-creds rejected, styled 404, CSRF-gated, anonymous SSE
   rejected, phone-api fails closed). `/version` = v2.4.0. The security
   hole has been closed on prod since the 2026-09-20 owner deploy;
   TODO row rewritten to a normal-train deploy.
5. **T10 pusher daemon** — diagnosed: the committer is a long-lived
   agent session loop (GPG-signed "chore: auto-commit" sweeps; PID
   1854266 has webphone cwd since 02:34). Pushes verified RESTORED
   (test pushes moved origins in all three repos). Runbook gained the
   narrative-commit-at-phase-boundary + daemon-drift-recheck lines;
   ROADMAP gained the docs/status+planning exclusion ask.
6. **T4b** — enforced erraudit on the chain-to-deploy: EXIT=0, 0
   findings.
7. **T11 prep** — one-page 14-decision briefing at
   `docs/planning/2026-09-22_13-50_owner-calls-briefing.md` (options +
   recommendation each).
8. **T6+T7 v2.5.0 RELEASED** — fold (CHANGELOG/FEATURES/TODO), version
   bump, release.sh inside the devShell: buildflow no-cache, go test,
   flake check, smoke, vulnix, tag `25740c6`, push, lychee, STACK
   relock in-train (browser E2E green ~292s, telephony-webphone VM test
   green, full stack flake check green, pushed as `91ac2c8`), aarch64
   `b700` ELF guard, gh release object with the folded notes.
   EXIT=0. https://github.com/LarsArtmann/webphone/releases/tag/v2.5.0
9. **T8 E2E budget re-baselined** — two forced-rebuild runs on the
   release chain: 384s and 373s, both EXIT=0 → budget **445s** set in
   the TODO watch row + ROADMAP (old 151s figure retired).
10. **T19 upstream tails** — (a) go-health container-free `NewChecks`
    quick-start in doc.go, vetted, pushed (`60a0aa9`); (b) family
    release-object audit: all 12 latest cqrs-htmx module tags had NO gh
    objects (last objects were the 2026-07 v4.3.0 era) — root v4.12.0
    cut with full CHANGELOG notes + 11 module sync objects + catalog
    standalone, all verified live; (c) NamedCheck.Timeout retro-note:
    already shipped in the family CHANGELOG — verified, nothing owed.
11. **T12 contacts API completion** — write limiter on
    `POST /api/contacts` (60/min burst 60, import-friendly; 429 +
    Retry-After test-pinned), OpenAPI 3.1 surface for all three
    operations with a spec-vs-handler contract test, per-extension
    count cap 500 enforced ATOMICALLY in the store upsert (renames
    never consume a slot; ErrListFull → 422). Store + server suites
    green at commit time (`8cd327c`).
12. **T13 SSE contacts nudge** — payload-less `contacts` event from all
    five mutation paths (island API, tab save/delete, vCard import);
    the Contacts tab re-fetches via `hx-trigger="sse:contacts"` with
    morph + stable ids on the stateful inputs; the island dropdown
    re-fetches via a document-level `htmx:sseMessage` listener (keyed
    on `detail.type`, decoded from the served sse extension source).
    Server nudge test (payload-less, save+delete), render-wiring test,
    3 island tests (fetch-exactly-once, negatives, rendered rows;
    helpers.mjs grew `replaceChildren`). All suites green at commit
    time (`4408342`).
13. **T14 session hardening** — (a) the CSRF adoption retry ladder
    (already implemented) now PINNED by two island tests: recover on
    retry 2 without reload; reload only after exactly three failures;
    (b) rotation-on-slide verdict written as a decision record
    (NOT-DO: double-submit makes token-only leaks inert, both-halves
    theft is bounded by the 30d cap; ROADMAP section RESOLVED) —
    committed `43f544e`.
14. **T15b GOEXPERIMENT sweep (code part)** — full suite proved green
    WITHOUT the flag (13/13 packages) → removed from flake devShell,
    buildGoModule env, `.buildflow.yml`, release.sh, webphone-smoke.py,
    README, CONTRIBUTING, AGENTS (6 mentions rewritten). **Gates after
    it: RED — see d).**
15. **T17 island diagnostics tail (code part)** — 10th+11th connection
    scenarios (concurrent rebuild triggers collapse into ONE rebuild;
    transient "rebuilding…" pill first shown, then registered — new
    `regRebuilding` key en/de, i18n parity green), smoke
    `--expect-version` flag verified positive AND negative against
    prod, stack `wait_marker` NOTIF prefix hack replaced with an
    explicit two-marker alternation. Island suites 14/14 green at the
    time; **commit + gate verification still open** (see b).

## b) PARTIALLY DONE

1. **T15a drift guard** — decided + documented (NOT a new buildflow
   step): the treefmt check inside `nix flake check` IS the drift gate
   (it caught operator.js today); the process gap was fixed by the
   runbook lines. Not yet recorded in the plan log/TODO.
   ~~recorded later~~ done (16:52 session closed the plan-log entries)
2. **T15c release.sh auto-lychee** — verified ALREADY SHIPPED (step 6,
   ran green in the release). ~~Not yet recorded in the plan log.~~
   done (same)
3. **T17 final verification** — the T17 files (connection.js, i18n.js,
   connection.test.mjs) were UNCOMMITTED at report time; the stack's
   browser.nix marker edit was daemon-swept into stack `1ea7dfc`
   (push state unverified) and has NOT been E2E-validated.
4. **T9 announcements** — v2.5.0 drafts written (headline + one-liner,
   `docs/announcements/2026-09-22_v2-5-0_drafts.md`) and committed;
   posting waits on owner channels + posture (briefing #5).
5. **T18a backup.retentionDays** — ~~investigation started (module backup
   block + oneshot script read); NO code written yet.~~ done (shipped
   same day: `3f90b01`/`32e1ac0`/`c893683`, VM-verified)
6. **Plan execution log** — ~~appended through T19/T2' but NOT updated
   for T12/T13/T14/T15/T17.~~ done (16:52 session: entries closed)
7. **CHANGELOG/FEATURES for the post-2.5.0 trains** — ~~T12/T13/T14/T17
   features are on main with NO `[Unreleased]` entries yet.~~ done
   (`4c4b36b` synced every train; the 2.6.0 fold `711fff5` carried
   them into the release) Same class
   of miss the runbook's fold step exists for — caught here before any
   next release, but it is a discipline miss (see e).

## c) NOT STARTED (from the plan, in order)

- ~~**T16 docs kernel**: AGENTS size pass (657+ lines → <400), HARVEST
  routing of report f-items, 18-49 ANNOTATE (owner-gated scope),
  19-37 plan checkbox hygiene, FEATURES VERIFY pass.~~ done (AGENTS
  705→337 `c53ee81`; 18-49 ANNOTATE + plan hygiene closed 18:46;
  HARVEST = the 2026-09-22 evening docs-health sweep)
- ~~**T18 rest**: README off-machine restic/borg pointer (18b — cheap),
  `/startupz`→systemd `Type=notify` contract doc (18c).~~ done (same
  day: README dated-history row; T18c documented NOT-DO verdict)
- ~~**T20 UX/a11y pass** (badge states + aria-live, migration toast,
  thread-list `data-dial`, `data-sms`, history ☆ save, three E2E
  scenarios).~~ done (`03b9431`, `7a6e8ac`, `bf4476e`, `5ae71b1`;
  E2E scenarios ×2 green 19:55)
- ~~**T21 messaging polish** (persist/display failure reasons, delivered
  badge, image thumbnails, draft persistence).~~ done (`1bec154`,
  `5ae71b1`, `25b0cf9`)
- ~~**T22 generated DOM-contract file** (emit from
  TestServedPageHoldsTheDomContract; AGENTS links it).~~ done
  (`docs/dom-contract.md` golden file, verified green 18:20)
- ~~**T23 PWA spike** (verdict doc; build gated on owner #11).~~ done
  (NOT-DO verdict `docs/planning/2026-09-22_17-05_pwa-spike-verdict.md`)
- **T24 recordings** (product-intent decision = owner #11 → then panel). ← still open (owner-calls row)
- ~~**T25 retention/cleanup job** (gated on T18a).~~ done (`d857568`)
- ~~**T26 platform tail 1** (metrics endpoint, TURN creds via config.js,
  per-extension export, timezone timestamps, MIME sniffing).~~ done
  (metrics 18:46+19:11; TURN `c1971c4`; export `37aae5a`; timezone
  + sniffing `d131e11`)
- ~~**T27 platform tail 2** (gzip option, favicon route, signed tags,
  i18n key-sync guard, idempotency durability decision, OpenAPI
  boundary decision, limiter-key runbook line).~~ done (`9f93537`+
  `44c0b8e`, `d131e11`, `4f067f8`, verdicts `0afe26d`)
- **Owner-gated**: T3 deploy command, T4a prod banner check, T5 SMS
  journal grep, T9 posting, T11 the 14 decisions themselves. ← still open (TODO_LIST owner rows)

## d) TOTALLY FUCKED UP (honest ledger)

1. **Post-sweep gates RED — the live fire.** `BUILDFLOW-RC=69`,
   `FLAKECHECK-RC=1`:
   - treefmt-check failure: MY T12 constructor edit in server.go missed
     the gofmt field-alignment after lengthening
     `contactsLimiter` (the diff names the `unread:`/`hooksIdem:`
     lines). Trivial `nix fmt`/gofmt fix — but it means **T12 was
     committed with formatting drift**; the earlier targeted test runs
     passed while the format gate did not run until now.
   - govulncheck step failed with a Go-version mismatch ("rebuild
     govulncheck with the current Go version") — possibly caused by
     removing `GOEXPERIMENT` from `.buildflow.yml` env, possibly
     pre-existing; NOT yet triaged. Candidate fix: restore the env for
     that step or rebuild the tool in the devShell.
   - Confounder: the gates ran while I kept editing the tree (T17
     files) and the daemon kept sweeping — the run is not a clean
     signal for anything except "re-run on a quiesced tree".
2. **Fabricated a git hash mid-edit.** When re-pinning pbx-artmann I
   typed a full 40-char rev knowing only the 8-char prefix — caught by
   my own check immediately and corrected with `git rev-parse`, but a
   fabricated-hash edit into a deployment flake is the exact bug class
   this repo's rituals exist to kill. Never write a rev that was not
   read from `git rev-parse`/`ls-remote` output.
3. **Release run #1 failed on a DOCUMENTED trap** — I launched
   `release.sh` OUTSIDE the devShell; host go 1.26.7 + go.mod 1.27.1
   failed every Go step (EXIT=69). The runbook had warned about this
   exact class all along. (The sweep in T15b removes the GOEXPERIMENT
   half of that trap; the toolchain half remains — release.sh still
   must run inside `nix develop -c`.)
4. **Invented test helpers** (`jsonStringField`, `urlQueryEscape`) in
   the SSE test — compiled-fail loop fixed by parsing with
   encoding/json + net/url like the rest of the suite does.
5. **Digit miscounting in test phone numbers — twice** — the count-cap
   test's "rename" number didn't match the filled rows (7-digit vs
   10-digit tails), cost one red test cycle each time.
6. **Narrative commits repeatedly split by the daemon** — five separate
   times my staged batch was partially swept into "chore: auto-commit"
   commits seconds before mine (fold, T12, T13, T14, T17), leaving my
   narrative commit carrying 1-5 of its files. Content always landed,
   history is polluted. Mitigation that DID work twice: amend the
   daemon's unpushed commit into the narrative message.
7. **Stub gap shipped a false green** — panels test passed while
   renderContacts actually THREW (`replaceChildren` missing from the
   stub); caught only because I chased a logged error the assertions
   didn't cover. Lesson: assertions must cover the render outcome, not
   just "no crash" (fixed in the same change).
8. **The vtsls `connection.test.mjs:174` hint from the prior session**
   turned out to be noise (points at a comment line; vtsls's LSP cannot
   even load the module under the host-go trap). Chasing it further
   would have been wasted — noted, closed.

## e) WHAT WE SHOULD IMPROVE

1. **Gate discipline: run the FULL gate before committing a train, or
   explicitly label the commit pre-gate.** T12/T13 passed targeted
   suites but the formatting gate ran only later. The cheap fix:
   `nix fmt` + `nix develop -c go vet` before every commit, full gates
   per train.
2. **Never race the gates.** Background gates + concurrent edits =
   unreadable signal (today's red is partly contention). Quiesce the
   tree (commit or stash) BEFORE launching buildflow/flake check, and
   edit nothing until the verdict is read.
3. **Hash discipline**: a rev/hash goes into a file ONLY as a
   shell-substituted `git rev-parse` value, never typed.
4. **Daemon contention protocol**: when a multi-file narrative commit
   matters, `git commit` IMMEDIATELY after staging (the daemon's sweep
   window is ~seconds), and re-check `git show --stat` after every
   commit; amend unpushed daemon commits that stole the batch.
5. **CHANGELOG entries per train, not per release**: T12-T17 features
   have no `[Unreleased]` bullets yet. Write the bullet in the same
   commit as the feature.
6. **release.sh should self-heal like buildflow.sh/smoke.py** (re-exec
   inside `nix develop -c` when the ambient go is below the floor) —
   one small script change kills the trap I fell into.
7. **Smoke `--expect-version` should ride the deploy TODO row's owner
   command** (done for the flag; row text update pending).

## f) NEXT — up to 50, in order

**Fire first:**

1. ~~Commit the pending T17 files (connection.js, i18n.js, connection.test.mjs).~~ done (files committed; gates re-ran green on the quiesced tree)
2. ~~`nix fmt` the server.go constructor alignment; verify with gofmt.~~ done (nix fmt fixed; treefmt green since)
3. ~~Triage the govulncheck failure (restore `.buildflow.yml` env for that~~ done (govulncheck green when quiesced, 16:52 session)
   ~~step OR rebuild govulncheck on 1.27) — decide sweep-adjust vs revert~~
   ~~of that one env line.~~
4. ~~Re-run BOTH gates on the quiesced tree; read the verdicts. GREEN~~ done (19:55 gate scoreboard all green)
   ~~before anything else ships.~~
5. ~~Verify the stack's daemon commit `1ea7dfc` (my browser.nix marker~~ done (stack 1ea7dfc validated by a green forced E2E; pbx-artmann relock #3 done 18:20)
   ~~edit) is pushed; run one forced E2E to validate the marker change;~~
   ~~relock pbx-artmann if the stack moves (owner deploy carries it).~~

**Then (plan order):**
6. ~~CHANGELOG `[Unreleased]` bullets for T12/T13/T14/T17 + FEATURES rows.~~ done (4c4b36b synced every train; the 2.6.0 fold 711fff5 carried them into the release)
7. ~~Plan execution-log update (T12-T17, gate incident).~~ done (16:52 session closed the plan-log entries)
8. ~~T18a `backup.retentionDays` (null default = keep-forever; prune in~~ done (T18a shipped same day, 3f90b01/32e1ac0/c893683, VM-verified)
~~the oneshot; module check stand-ins).~~
9. ~~T18b README off-machine restic/borg pointer.~~ done (T18b README dated-history row, same day)
10. ~~T18c `/startupz`→systemd contract doc (+ optional wiring).~~ done (T18c documented NOT-DO verdict, same day)
11. ~~T16b HARVEST routing of still-open report f-items.~~ done (2026-09-22 evening docs-health sweep)
12. ~~T16d 19-37 plan checkbox hygiene + FEATURES VERIFY.~~ done (18:46 session)
13. ~~T16a AGENTS size pass (657 → <400; move detail to docs/).~~ done (c53ee81, AGENTS 705 to 337 lines)
14. ~~T22 generated DOM-contract file + AGENTS/stack-doc link swap.~~ done (T22 dom-contract.md golden file, verified green 18:20)
15. ~~T21a persist + display message failure reasons.~~ done (1bec154)
16. ~~T21b distinct delivered badge.~~ done (1bec154)
17. ~~T21c image thumbnails (decide server-side vs CSS first).~~ done (5ae71b1)
18. ~~T21d draft persistence per thread.~~ done (25b0cf9)
19. ~~T20a badge ringing/established states + aria-live.~~ done (03b9431)
20. ~~T20b migration completion toast en/de.~~ done (5ae71b1)
21. ~~T20c thread-LIST `data-dial` restructure.~~ done (7a6e8ac)
22. ~~T20d `data-sms` affordance (history/voicemail → compose prefilled).~~ done (bf4476e)
23. ~~T20e history ☆ save-as-contact.~~ done (bf4476e)
24. ~~T20f/g/h the three E2E scenarios (logged-out dial, contacts~~ done (19:55 E2E x2 green, 198s + 237s)
~~round-trip, live badge) — each ×2 green.~~
25. ~~T23a PWA spike verdict doc.~~ done (verdict doc written, 16:52)
26. ~~T26a metrics endpoint (Prometheus text).~~ done (metrics + module location, 9f93537)
27. ~~T26b short-lived TURN REST creds via /config.js.~~ done (c1971c4)
28. ~~T26c per-extension data export (zip).~~ done (37aae5a)
29. ~~T26d timezone-aware timestamps.~~ done (d131e11)
30. ~~T26e MIME sniffing on attachments.~~ done (5ae71b1)
31. ~~T27a nginx gzip module option.~~ done (9f93537 + 44c0b8e)
32. ~~T27b /favicon.ico route.~~ done (d131e11)
33. ~~T27c signed tags (git tag -s) in release.sh.~~ done (4f067f8)
34. ~~T27d i18n dynamic-template key-sync guard.~~ done (d131e11)
35. ~~T27e webhook idempotency durability decision record.~~ done (0afe26d)
36. ~~T27f OpenAPI boundary decision record.~~ done (0afe26d)
37. ~~T27g limiter-key widening runbook line.~~ done (0afe26d + error-contract section)
38. release.sh self-heal into nix develop (the e/6 improvement).
39. ~~T25a-c retention/cleanup job (after T18a).~~ done (d857568)

**Owner console (one sitting):**
40. T3: run the deploy command (in the TODO row) + smoke
`--base … --expect-version 2.5.0`.
41. T4a: rejection-banner check (needs a real extension session).
42. T5: `journalctl -u telnyx-webhooks` grep + SMS lane restore.
43. T11: the 14-decision batch (briefing doc; recommendations ready).
44. T9: pick channels + posture, post the drafts.

**Standing (dated):**
45. Monthly erraudit tiers-1+2 re-measure (next: 2026-10-22).
46. Quarterly watches re-check (next: 2026-12-20).
47. Per-train runbook incl. aarch64 named checks + closing sweep.
48. E2E budget watch from the new 445s baseline (two consecutive
over-budget runs before digging).
49. Decide/act on the daemon docs/status exclusion ask (ROADMAP line).
50. ~~Next train fold: the accumulating [Unreleased] (T12-T17 + whatever~~ done (v2.6.0 folded and tagged 807ca0c)
~~lands) — g2 theme already forming ("live surfaces + honesty").~~

## g) Questions I cannot answer myself

1. **Deploy timing**: the v2.5.0 chain is fully staged on pbx-artmann
   (`36ae250`, toplevels green). Do you want to run the owner deploy
   command NOW (I am barred from it by pbx-artmann's AGENTS), or hold
   until the post-sweep gate reds here are re-run green so the deploy
   carries a fully-gated main? (Prod currently serves v2.4.0 content —
   safe, just older.)
2. **The auto-commit daemon session** (the overnight webphone-cwd agent
   loop, still alive): I documented the docs/status+planning exclusion
   ask in ROADMAP, but reconfiguring or stopping that session is your
   call — it also caused today's formatting-drift reintroduction and
   five split commits. Keep it as-is, or do you want it stopped/taught
   the exclusions before the next train?
3. **Announcement posture**: briefing #5 recommends "fix-acknowledged,
   no exploit detail" for the v2.5.0 announcement drafts. Approve that
   (and the draft wording), or do you want the security story told
   differently — including whether to name the forged-session hole at
   all, given it never was live-with-fix-missing (prod was already on
   v2.4.0)?

— Reported 2026-09-22 15:06 CEST. Background: no servers booted by me
remain; the only background processes were the gate runs (verdicts
above) and both E2E baseline runs (completed EXIT=0).
