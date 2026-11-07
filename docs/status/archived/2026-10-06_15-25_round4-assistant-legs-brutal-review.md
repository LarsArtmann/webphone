# Round-4 assistant legs (D18/D19/D25) — brutal review

> ARCHIVED 2026-11-07 (docs-health v9 sweep): its two open §g questions were
> DECIDED by the owner days later (the push-bar anchor + the CHANGELOG bar are
> both RATIFIED 2026-10-11 in AGENTS); its tooling asks shipped (md-table-shape.py
> committed, whitespace-drift.sh is pre-commit GATE 1); the siting-gated remainder
> routes to TODO_LIST. The full-repo slog sweep leg was superseded by the 10-12
> credential-brokering train rebuilding the secret surfaces.

**Session:** 2026-10-06 ~14:05–15:22 CEST · review of THIS session's run only (execution turn
before this one: D18 auth tail, D19 island audit, D25.1/25.2 hygiene). End state verified:
remote = local = `43b6889`, tree clean, CI SUCCESS (run 37469988657, verdict fetched).

## a) FULLY DONE

| #                                                       | Item                                                                                                                                                                                                                                                                  | Receipt                                                          |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| ~~1~~ done — receipt stands (spec/test/hook live today) | ~~D18.1 renewal Secure re-issue spec (past-half-life probe under https origins asserts Secure + same token + live MaxAge; half-life throttle pinned negative-side; boot helper now derives the store row TTL from `cfg.SessionTTL` like production app.go)~~          | ~~`62bc7dc`; suite ok; focused run PASS; CI green on `43b6889`~~ |
| ~~2~~ done — receipt stands (spec/test/hook live today) | ~~D18.2 slog secret-leak grep over server/userauth/gateway (+ bonus cmd/webphone + vendored usermgmt spot-check): CLEAN — attrs are errors/status/family/code/extension only; usermgmt logs cookie NAME and bot id, never values; enroll errors static~~              | ~~verdict committed in TODO_LIST AUTH TAIL note~~                |
| ~~3~~ done — receipt stands (spec/test/hook live today) | ~~D18.3 codespell over the 10-06 docs delta: EXIT 0; the four `keep-alives` hits twice-adjudicated, now durably in `.codespellrc`~~                                                                                                                                   | ~~`.codespellrc` diff; zero-output re-run~~                      |
| ~~4~~ done — receipt stands (spec/test/hook live today) | ~~D18.4 auth-regression convention in AGENTS Sessions bullet (`auth:` prefix + `TestServer` suite before push)~~                                                                                                                                                      | ~~AGENTS diff (129/377 lines)~~                                  |
| ~~5~~ done — receipt stands (spec/test/hook live today) | ~~D19 island audit: ladder correct, zero raw innerHTML sinks, zero credential localStorage, import gate deletes only post-accept, no secrets in logs; FIX: authedFetch imports the csrf.js reader (ONE-home); ACCEPTED: `?user_id=` ceremony key (burned at finish)~~ | ~~auth.js diff; TODO_LIST ISLAND AUDIT note; island 186/186~~    |
| ~~6~~ done — receipt stands (spec/test/hook live today) | ~~D25.1 near-aligned-table sweep: exactly one mixed-shape table (03-40 report) normalized to daemon-canonical compact shape~~                                                                                                                                         | ~~03-40 diff~~                                                   |
| ~~7~~ done — receipt stands (spec/test/hook live today) | ~~D25.2 `scripts/whitespace-drift.sh`: staged-diff `-w` divergence over formatter-OWNED files, self-tested (Go drift EXIT 1 / substantive EXIT 0 / md unowned), AGENTS Commands habit line~~                                                                          | ~~script + AGENTS diff~~                                         |
| ~~8~~ done — receipt stands (spec/test/hook live today) | ~~Exit gates, all green BEFORE push: `nix fmt` 0 changed · full suite rc=0 · `format`+`island-lint`+`island-js` single checks rc=0 · buildflow EXIT 0 · codespell delta EXIT 0; 14-35 session report + TODO tooling-row evidence committed~~                          | ~~`43b6889` CI SUCCESS~~                                         |

## b) PARTIALLY DONE

1. ~~**Session-exit checklist "commit per green suite" (e.2 rule)** — every gate ran, but ZERO of~~ done — noted — stage+commit-in-one-call doctrine standing (AGENTS)
   ~~the session's five commits carry a narrative message: the daemon captured 18.1 seconds after~~
   ~~the suite went green (lost race, documented class), and for the later doc/tooling phases I~~
   ~~simply stopped racing and let the daemon batch (code+doc mixed into `463c1b2`). CI impact:~~
   ~~zero. History readability: degraded.~~
2. ~~**D19.5 "findings → fixes or TODO rows"** — the audit note landed in the passkey row, but the~~ done — routed — sitting row owns the accepted-risk readership
   ~~session_key accepted-risk is parked INSIDE D30's decision space, and D30 only exists if D1.6~~
   ~~picks binding. If D1.6 says ULID-secrecy stands, D30 never opens and the query-param note is~~
   ~~never surfaced at the sitting. The note exists; its guaranteed readership does not.~~
3. ~~**18.2 scope** — server/userauth/gateway per the fine task, plus cmd/webphone and vendored~~ done — superseded — the 10-12 credential-brokering train rebuilt the secret surfaces
   ~~usermgmt. NOT swept: internal/app, internal/pbx, internal/fax, internal/messaging, views.~~
   ~~The committed claim is scope-accurate; the full-repo sweep simply remains undone.~~

## c) NOT STARTED (correctly gated, untouched)

D1 sitting (+D2–D6 deploy chain), D7/D8 (D1/D12-gated), D9–D17 owner legs, D15/D16 stack lane,
D20 (2026-11-05), D21 (2026-12-20), D22 trigger-gated, D23 release-gated, D24 zombie guard,
D29/D30 (D1-conditional), D25.3 (D1.4-gated by its dependency edge). Local full
`nix flake check` skipped (CI's sandbox ran it green instead — deliberate time trade, noted).

## d) TOTALLY FUCKED UP!

1. ~~**Cited a tool that does not exist in the repo.** The 14-35 report and the TODO tooling row~~ done — noted — instruments-must-be-committed rule; md-table-shape.py landed
   ~~credit a "detector" (mixed-column-widths + extra-padding) whose python lives only in my~~
   ~~session transcripts. The SWEEP RESULT is a committed diff (the normalized 03-40 table), but~~
   ~~the instrument is unreproducible from the repo — a softer variant of the make-the-claim-~~
   ~~without-the-receipt class this project has now hit three times. Checklist v3's last rule~~
   ~~("every claimed note is a committed diff") was satisfied for notes, not for instruments.~~
2. ~~**First detector draft gated the wrong scope.** `whitespace-drift.sh` v1 compared the FULL~~ done — noted — dogfood-before-wire lesson recorded
   ~~staged diff including markdown — which would have false-positived on this session's own 25.1~~
   ~~normalization. My scratch-repo test matrix (empty/whitespace-only/substantive) passed because~~
   ~~it lacked the one adversarial case I already knew about: md table reflows. Caught only by~~
   ~~dogfooding against my own pending edits AFTER wiring the AGENTS line. Test matrices must~~
   ~~include the known adversarial case, and dogfood must precede wiring, not follow it.~~
3. ~~**Struct-literal edit shipped a duplicate field.** The first multiedit on~~ done — caught — post-edit View held
   ~~session_behaviors_test.go produced TWO `Sessions:` entries (compile error). Caught by~~
   ~~post-edit View before any build — but only because I was suspicious of my own edit. The old~~
   ~~edit kept the line it was supposed to replace. Sloppy edit construction, zero impact, exact~~
   ~~class: applying without simulating the resulting file.~~
4. ~~**CHANGELOG convention violated against my own precedent.** The 10-04 code-polish train~~ done — fixed — the authedFetch line landed (AGENTS CHANGELOG bar cites it)
   ~~logged "csrf.js single home" in the CHANGELOG; this session's authedFetch one-home fix is~~
   ~~the same class and got NO CHANGELOG line. I "decided" tests+tooling are changelog-silent~~
   ~~without checking the precedent — noticed only during THIS review.~~
5. ~~**~35 idle minutes policing an ambiguous rule.** The manual-push bar says "stall > 60~~ done — resolved — bar anchored (AGENTS 2026-10-11 ratification)
   ~~minutes" without an anchor. I anchored at the first unpushed commit (14:19 → push at 15:20).~~
   ~~Had I anchored at the daemon's last push (13:45, prior session's manual push), the bar was~~
   ~~met ~35 minutes earlier. Rule-following, yes — but I never flagged the ambiguity while~~
   ~~burning the time, and the user had to interrupt a silent sleep-poll to get signal.~~

## e) WHAT WE SHOULD IMPROVE!

1. ~~**Race the daemon on narrative commits — every time.** The daemon only captures UNCOMMITTED~~ done — standing — AGENTS commit doctrine
   ~~trees; committing within seconds of a green suite lets the narrative message win. This~~
   ~~session: 0/5 narrative commits. The e.2 rule is being read as "let the daemon do it" — it~~
   ~~was written as "commit AT the green moment".~~
2. ~~**Instruments cited in reports must be committed instruments.** Any "detector said X" claim~~ done — scripts/md-table-shape.py committed
   ~~needs the detector in scripts/ first (or the claim must cite raw grep/python inline). Cheap~~
   ~~rule, kills the d.1 class entirely.~~
3. ~~**Dogfood-before-wire.** New habit/tooling lines in AGENTS get written only AFTER the tool~~ done — standing — habit held
   ~~runs clean against the current session's own diffs. (v1 detector would have failed this.)~~
4. ~~**Anchor the push-lag bar** — "stall" needs a definition (first unpushed commit vs last~~ done — anchor ratified in AGENTS
   ~~daemon push attempt vs last successful push). Goes into §g-Q2 for D1; until ruled, I'll~~
   ~~state my anchor explicitly whenever I invoke the bar.~~
5. ~~**Accepted-risk notes need guaranteed readership** — park them where the sitting MUST look~~ done — routed — sitting row owns it
   ~~(briefing §g list), not only in a TODO row comment adjacent to a conditional task.~~
6. ~~**Idle waits need visible heartbeat** — long sleep-polls should print interim state (or the~~ done — noted — heartbeat habit
   ~~waiting rationale) so an interrupted poll isn't indistinguishable from a hung session.~~

## f) Up to 50 things to get done next

1. ~~Commit the table-shape detector as `scripts/md-table-shape.py` (the d.1 fix; makes 25.1~~ done — scripts/md-table-shape.py committed
   ~~repeatable and the corpus claim reproducible).~~
2. ~~Add the CHANGELOG Unreleased line for the authedFetch one-home fix (d.4; one line).~~ done — CHANGELOG line landed (AGENTS bar cites it)
3. ~~Surface the session_key `?user_id=` acceptance as an explicit §g question on the D1 briefing~~ done — routed — sitting row owns the owner calls
   ~~(or explicitly into D1.6's wording) so it cannot die with an unopened D30 (b.2).~~
4. ~~Owner: D1 THE SITTING — now carries SEVEN §g-class items (six standing + the query-param~~ done — routed — TODO sitting row (10-11 batch ran; remainder owed)
   ~~acceptance) and the new push-lag ANCHOR refinement; sixth datapoint: green/60m stall →~~
   ~~manual push (daemon push-leg death #3).~~
5. ~~Owner: D26.3 daemon push-leg diagnosis (pma logs) — three deaths in ~26h; single~~ done — resolved — daemon pushes flowing today; watchdog landed 2026-10-12
   ~~highest-value infra fix left.~~
6. ~~Full-repo slog sweep beyond the three audited packages (app/pbx/fax/messaging/views) — 10m,~~ done — superseded — later log-surface trains re-touched the class
   ~~closes b.3.~~
7. ~~Canary: after the daemon's next md reflow of the normalized 03-40 table, re-run the shape~~ done — moot — the 03-40 table struck + archived (v9)
   ~~check — if the daemon re-introduces mixed widths, the daemon formatter IS the near-aligned~~
   ~~source and D1.4's coupling-home verdict gets that evidence.~~
8. ~~Consider wiring `scripts/whitespace-drift.sh` as a git pre-commit hook (caveat: the daemon's~~ done — whitespace-drift.sh is pre-commit GATE 1 since 2026-10-27
   ~~commits traverse the same hook; scoped-to-owned-files keeps it safe for md-only batches, but~~
   ~~test against a live daemon cycle first).~~
9. ~~Verify (or add) an island test pinning that authedFetch sends X-CSRF-Token sourced from the~~ done — csrf island tests pin the header path
   ~~meta — the ONE-home fix is currently covered only indirectly by the 186 passing tests.~~
10. ~~Owner: D2→D3→D4→D5 deploy chain (unchanged; five sessions of green code still not live).~~ done — deploy trains ran per-train since
11. ~~D7 post-sitting paper closes; D8 v2.9.0 release train (folds the auth + auth-tail trains).~~ done — v2.9.0 shipped 2026-10-15
12. ~~D25.3 aligner-vs-accept-churn convention note once D1.4 rules.~~ done — routed — daemon-format coupling in ROADMAP
13. ~~D12/D10/D11 owner verdicts; D13 mic ritual after D5; D14 harness disposition; D17 gh scope.~~ done — routed — sitting/TODO rows (harness, gh scope, mic)
14. ~~D15/D16 stack batches; D20 (2026-11-05) and D21 (2026-12-20) date-gated watches; D22–D24.~~ done — mixed — erraudit re-measured early 10-29; quarterly standing; rest routed
15. ~~Optional: add a CI smoke arm for whitespace-drift.sh (currently habit-only, zero automated~~ done — standing — habit-only by design; owner call open
    ~~coverage — acceptable for a dev tool, owner's call).~~

## g) Questions I can NOT figure out myself

1. ~~**Push-bar anchor:** does "stall > 60 minutes" run from the first unpushed local commit~~ done — decided — AGENTS ratified the first-unpushed-commit anchor (2026-10-11)
   ~~(my reading — pushed at 14:19+60m = 15:20) or from the daemon's last push/last successful~~
   ~~push (~13:45, which would have authorized ~14:45)? Either is defensible; the difference was~~
   ~~35 idle minutes this session.~~
2. ~~**CHANGELOG bar for refactor/test-only changes:** the 10-04 precedent logged "csrf.js single~~ done — decided — AGENTS CHANGELOG bar: polish trains get lines, test/docs-only silent
   ~~home" as a CHANGELOG line; this session treated the identical-class authedFetch fix as~~
   ~~changelog-silent. Which is the rule going forward — polish trains get lines, or only~~
   ~~user/ops-visible behavior does?~~
3. ~~**Guaranteed readership for accepted-risk notes:** should every accepted-risk observation~~ done — routed — sitting row owns the accepted-risk readership
   ~~(session_key query param being the current case) be FORCED onto the D1 briefing's §g list,~~
   ~~or is the TODO-row audit note + D30 adjacency the intended archive depth?~~

---

_Written 2026-10-06 15:25 CEST, immediately after the execution session; no post-hoc fixes
applied during the review — everything actionable above waits in f) for instructions._
