# Status Report — docs-health AUDIT v5: living docs made superb + archive sweep (2026-10-01 17:27)

**When:** 2026-10-01 17:27 CEST · **Branch:** `main` (tree clean; daemon swept everything, HEAD `08a6479`)
**Session scope:** the owner mandate — "View ALL `**/2026-0*` files; execute the docs-health
SKILL PROPERLY; TODO_LIST / CHANGELOG / AGENTS / README / ROADMAP / FEATURES all SUPERB;
archive FULLY done and UPDATED (inline strikethrough) .md files." Documentation-only: **zero
product code touched.**
**Skill loaded:** docs-health SKILL.md + `check-rows.py` + `annotate-status-items.py` read
before acting; AUDIT mode (BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE).

> ARCHIVED 2026-10-03 (docs-health v6 sweep): SUPERSEDED by the v6 audit
> (the 00:35 + 02:09 reports + this closing leg) — every f-item below is
> closed: the 34-file annotate+archive program completed, the product trains
> all shipped, the owner legs ride their rows; the two standing owner
> questions (the four gated plans, the check-rows residue) remain open by
> design. Per-item verdicts inline.

---

## Headline

The six living docs now describe the **actual** current product — including the 2026-10-01
UI/UX train (M1–M8) that had shipped to `main` after the v2.8.0 tag and was **invisible in
every living doc**. `FEATURES` gained 8 rows, `CHANGELOG` the UI/UX entry, `README` 3 rows,
`AGENTS` the shell/a11y contract + the Ledger boundary ruling; `TODO_LIST` was rebuilt
(6 DONE rows deleted, 11 fresh rows harvested, un-padded into readable entries); `ROADMAP`
gained 6 open questions + a raw-ideas cluster. **20 fully-resolved historical files were
annotated inline and archived**; the previously-archived plan that carried zero
strikethroughs was repaired. Gates: the archived-dir completeness gate is CLEAN, check-rows
passes on every newly annotated file, and the scoped `go test cmd/webphone internal/server`
is green (DOM-contract intact).

---

## a) FULLY DONE (verified this session)

| #   | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Evidence                                                        |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
~~| a1  | **Inventory + skill load.** All `docs/**/2026-0*` artifacts enumerated (status, planning, reviews, research, announcements, architecture-understanding, HTML/D2/SVG); docs-health SKILL.md + tooling read first.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | glob/ls; SKILL.md read                                          |~~ done — this session (report of record)
~~| a2  | **VERIFY fleet-wide.** Living-doc claims re-derived against code: DOM-contract id count **35 → 38** (recounted `docs/dom-contract.md`, gained `offline-banner`, `wp-live`, `wp-tab-skeleton`), the UI/UX features confirmed present (`shell.js` palette, `layout.templ` skip link + `#wp-tab-skeleton` + `#wp-live`, `app.css` day separators + mobile bottom bar, `settings.templ`/`i18n.go` palette keys), island test files counted (16), git log of the UI/UX train read.                                                                                                                                                                                                                                                                                           | greps, `docs/dom-contract.md`, `git log`                        |~~ done — this session (report of record)
~~| a3  | **FEATURES.md made true.** +8 rows: Calling `Fast answer (mic pre-warm)` (🟡 PARTIALLY_FUNCTIONAL — code green, live-unverified); Messaging `Optimistic send bubble`, `Transcript day separators`; Live updates `Morph accessibility`; Platform `Command palette + shortcut help`, `Skip-to-content + form associations`, `Mobile shell`, `Tab loading skeleton`. DOM count 35→38; Browser-E2E row → v2.8.0 gates (`3afcf57`) + owed-run note; aarch64 refreshed.                                                                                                                                                                                                                                                                                                       | FEATURES.md:24,45,46,103,136-139,128,134                        |~~ done — this session (report of record)
~~| a4  | **CHANGELOG.md made true.** `[Unreleased] → Added` gained the "Trust + accessibility batch" entry covering the whole 2026-10-01 UI/UX train (incl. the contacts-manager revert as Ledger's domain).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | CHANGELOG.md:67                                                 |~~ done — this session (report of record)
~~| a5  | **AGENTS.md made true.** New "Shell & accessibility contract (2026-10-01 UI/UX train)" bullet (palette, skip link, skeleton, `#wp-live`, bottom bar, `aria-current`, optimistic bubble) + the **OPERATOR RULING** that contacts depth is Ledger's domain (revert `6989b99`); the concurrent-sessions paragraph gained the daemon edit-race rule (re-View / atomic writes).                                                                                                                                                                                                                                                                                                                                                                                              | AGENTS.md shell bullet + concurrent-sessions section            |~~ done — this session (report of record)
~~| a6  | **README.md made true.** Capability table +3 rows (`Command palette`, `Accessibility`, `Mobile`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | README.md:37-39                                                 |~~ done — this session (report of record)
~~| a7  | **TODO_LIST.md rebuilt (BUILD rules enforced).** The skill-violating **6 🟢 DONE rows were DELETED** (setup-bundle, erraudit family adoption, 09-29 nix-review batch, `/version` ldflags, fax→Paperless, templ-components — each now in CHANGELOG/FEATURES); the 2168-char padded table was un-padded into readable `###` entries; **11 fresh rows harvested** from the 2026-10-01 reports/plans (UI/UX M9–M26 remainder, UI/UX gates + stack-E2E obligation, verification+performance plan, ring-silence bug, island boot language split-brain, mic-prewarm verification, island-honesty follow-ups, nix-review batch 2, samber/do+dashboard follow-ups, visual-gate harness, cross-repo obligations). Result: 21 open tasks, every row citing its source in Evidence. | TODO_LIST.md (full rewrite)                                     |~~ done — this session (report of record)
~~| a8  | **ROADMAP.md harvested.** +6 open questions (UI/UX execution shape, v2.9.0 fold timing, stack `/health` exposure policy, ring-silence authorization, token-hygiene parity, contacts-ownership confirmation) + a `## Raw ideas (2026-10-01 harvest)` cluster.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | ROADMAP.md:297-330                                              |~~ done — this session (report of record)
~~| a9  | **ANNOTATE + ARCHIVE (the core ask).** **20 files `git mv`-ed** into `archived/`: 16 fully-resolved status reports whose annotation the v4 sweep had completed but whose **archive leg never ran**, plus `02-47` (release-tail items resolved + struck), plus `13-17` dedup sweep, plus the `send-failure-ux` and `composer-ux` plans (self-declared EXECUTED; coarse "this train" rows struck).                                                                                                                                                                                                                                                                                                                                                                        | git R-count; `docs/status/archived/`, `docs/planning/archived/` |~~ done — this session (report of record)
~~| a10 | **Repaired the archived plan that failed the gate.** `docs/planning/archived/2026-09-30_13-14_SUPERB-post-review-pareto-execution-plan.md` had **zero strikethroughs**; 12 T-rows struck inline (`T01/T04–T12/T18/T19 done at…`) + an inline resolution note; converted to the sanctioned whole-line-strike form so `check-rows` passes.                                                                                                                                                                                                                                                                                                                                                                                                                                | check-rows: complete                                            |~~ done — this session (report of record)
~~| a11 | **Stale-reference sweep after the moves.** Repointed every pointer to the two moved plans (CHANGELOG ×2, `docs/error-contract.md`, the 13-33 plan), verified no live doc cites a moved status report, and confirmed no archived path cited in the living docs is missing.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | greps (empty)                                                   |~~ done — this session (report of record)
~~| a12 | **Gates run:** archived-dir completeness `grep -rLn '~~'` → **CLEAN**; `check-rows.py` on the three newly annotated files → **complete**; `nix develop -c go test -count=1 ./cmd/webphone ./internal/server` → **ok** (DOM-contract + CSP tests green).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | command outputs                                                 |~~ done — this v6 sweep re-runs the same gates at its close
~~| a13 | **Not-archived decisions made deliberately:** 2026-09-30/10-01 reports stay live (current, still open); the four owner-gated plans (`error-excellence`, `20-year-durability`, `13-33`, `04-29`) stay live (owner g1 + parked M18); the operative verdict/decision docs (csrf, pwa, webhook-idem, openapi, session-persistence, tailwind, command sheet, briefing, announcements) stay live per skill SKIP policy.                                                                                                                                                                                                                                                                                                                                                       | session reasoning                                               |~~ done — this session (report of record)

---

## b) PARTIALLY DONE

| #  | Item                                                                | What's missing / why                                                                                                                                                                                                                                                                                                                            |
| -- | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
~~| b1 | **Annotations of the 20 recent reports (2026-09-30 / 2026-10-01).** | Left UN-annotated on purpose: they are 1 day old and explicitly "waiting for instructions" (open action items). Annotating a live report prematurely would be wrong; they need the next sweep once their trains close.                                                                                                                          |~~ done — this v6 sweep annotated + archived them all (the 34-file program complete)
~~| b2 | **The four owner-gated plans.**                                     | `04-29 pareto`, `13-33 stack-adoption`, `error-excellence`, `20-year-durability` were classified FULLY-RESOLVED-or-routed by the v4 report but carry owner-open remainder (its `g1` question, `M18` parked). Left live; not annotated. Owner ratification pending.                                                                              |~~ owner — still live pending the sitting (the standing v6 default)
~~| b3 | **check-rows uniformity over the whole archived corpus.**           | 37 archived files (mostly pre-2026-09-18) are flagged INCOMPLETE because they use the older **marker-cell format** (`\| ~~ \| 7 \| …`) which tickles the annotate-rows↔check-rows format mismatch. The hard `grep` gate passes on all of them. Pre-existing, owner-pending "accepted baseline" question (the repo's own 04-26 report flags it). |~~ owner — the g2 tooling question stands; this sweep verified the newly annotated files file-by-file
~~| b4 | **Full quality gate.**                                              | Only the scoped `go test cmd/webphone internal/server` + marker/link gates ran — matching the docs-only-delta precedent (no code changed). `buildflow`, `nix flake check`, the island suite, `vulnix` were NOT run.                                                                                                                             |~~ resolved by events — the later batteries ran (the master-todo trains gated everything)
~~| b5 | **FEATURES depth on the new rows.**                                 | The UI/UX feature rows cite code presence, not a live/exercised verification; the E2E obligation is recorded but the run is owed. Honestly labeled where uncertain (mic pre-warm = PARTIALLY_FUNCTIONAL).                                                                                                                                       |~~ routed — the live verification rides the TODO island-honesty E2E row
~~| b6 | **`05-11` (v4 report) and `04-26` (v3 report).**                    | Annotated with the legs this session completed (archive of 16, gate clean) and the resolved items, but both still carry genuine open items (the 6-plan leg, `g1`, DOMAIN_LANGUAGE hash-bar `g3`, the 3-file restyle) → kept live.                                                                                                               |~~ done — v3 + v4 were archived by the earlier v6 legs
~~| b7 | **The harvest is a synthesis, not a 1:1 copy.**                     | ~150 §f items across the recent reports were collapsed into 11 TODO rows + ROADMAP questions. Some low-value items were intentionally dropped; the mapping is by theme, not by line.                                                                                                                                                            |~~ record stands — the synthesis is by design

---

## c) NOT STARTED

- **Any product/code work.** None attempted — the mandate is documentation-only.
- **The remaining archive candidates** (the four owner-gated plans + the ~20 recent reports) — blocked on their own closure, not on effort.
- **The 50 next-things backlog** in (f) — all unstarted.
- **health.css build-script commit / CI rebuild** (noted in the 02:12 report) — not touched; it is a code/docs task owned by a future train.
- **A fresh stack browser E2E** after the UI/UX markup change — recorded as an obligation, not run (needs the stack repo + PBX).
- **`docs/DOMAIN_LANGUAGE.md` review** — it exists now; not read this session (out of scope, no drift detected).

---

## d) TOTALLY FUCKED UP _(honest accounting; nothing shipped broken, tree green)_

~~1. **Chopped body text in my first TODO_LIST regeneration.** My title-stripping logic derived a heading from the first separator, then removed that prefix from the prose — but when the separator sat beyond 90 chars the prefix came from a _truncated_ title, so bodies began mid-sentence ("04-32_*`§f): byte-exact…"). Caught on the read-back; regenerated with`body = full task`. **One wasted regeneration cycle from a self-inflicted off-by-separator bug.**~~ process record
~~2. **Edit path typo `/home/labs/...`.** The AGENTS concurrent-sessions edit failed with "file not found" because I typed `labs` for `lars`. Trivial, but a wasted round trip and exactly the "compose-then-verify" failure class the repo's own reports name.~~ process record
~~3. **Per-cell strikes tripped check-rows.** I first struck DONE table rows by wrapping each cell (`| ~~T01~~ … | ~~…~~ |`), which check-rows correctly flagged as **PARTIAL** (mixed tables). The sanctioned pattern is the whole-line strike (`~~| T01 | … |~~ verdict`), which removes the row from table syntax. Converted all three files; passed. **I created the exact "planted-miss class" the skill warns about, then fixed it — it cost a detection cycle.**~~ process record — the whole-line form became the standard
~~4. **I trusted report-cited numbers instead of re-running them.** The FEATURES Browser-E2E and aarch64 rows were refreshed from the reports' own ledgers (stack `3afcf57`, ELF `b700`), not from a re-run — correct as of the sources, but not independently re-proven this session.~~ process record
~~5. **No prior baseline for the health score.** This is the first v5 audit in-session; I reported the findings-as-found with fixes rather than inventing a "prior score" (per the format's "never invent a prior state" rule).~~ process record
~~6. **I did not write a session report until asked.** The docs-health AUDIT mode says print inline; the repo convention wants a status report. I printed inline and paused — the owner had to ask for this file.~~ process record

---

## e) WHAT WE SHOULD IMPROVE

~~1. **Generate TODO_LIST via a parser, never a bespoke title heuristic.** The chop bug came from inventing a title-stripper; a small committed helper (parse table → emit entries) would make this repeatable and bug-free. (Candidate: `scripts/docs/` helper.)~~ not adopted — below the bar (no committed helper)
~~2. **Match the strike format to the tool from line one.** Use whole-line strikes (or `annotate-status-items.py`) for any table, never per-cell hand-wrapping — check-rows exists to catch the difference and did.~~ record stands — whole-line strikes are the standard form
~~3. **Archive insurance: verify BEFORE `git mv`, every time.** The 16 files sat annotated-but-unarchived for two days precisely because the v4 sweep annotated first and never moved them. The order is annotate → verify → archive; make it mechanical.~~ done — the annotate→verify→archive order is mechanical in the v6 sweep
~~4. **The annotate-rows↔check-rows format mismatch is a tooling defect worth routing upstream.** `check-rows.py`'s docstring claims the marker-cell format "counts as struck" but its `classify_row` requires every cell to carry `~~`; that is why 37 archived files false-flag. Either fix check-rows to accept the marker column, or migrate the corpus — one or the other, not both dangling.~~ owner — the g2 check-rows question
~~5. **Re-derive living-doc counts from a standing gate.** The 35→38 DOM drift recurred (a prior sweep fixed 34→35); the DOM-contract list should be parsed by a small check that writes the count into FEATURES (or FEATURES should stop citing a number and point at the file).~~ done in part — TestServedPageHoldsTheDomContract is the standing gate; FEATURES keeps a hand-cited count
~~6. **Keep AGENTS.md leanness in view.** It grew again (shell/a11y + Ledger ruling + daemon note). The additive posture is correct (enduring rules), but the compaction (`498 → ≤377`, owner-gated) is now more urgent — the next content add should pay for itself by removing a line.~~ done — the compaction landed (404→377) + the preflight cap
~~7. **A docs-only delta still deserves the marker/link gate + a written scope statement** (which the reports' precedents do); I matched that, but should say it explicitly at close, not only in the final message.~~ process record
~~8. **Cite the source report for every harvested TODO row** (done) **and** avoid re-listing the same item in ROADMAP + TODO (the 04-26 report warned of double-listing). I mirrored questions into ROADMAP and tasks into TODO — reviewed once, but a standing dedup check would help.~~ process record

---

## f) Up to 50 things we should get done next

> Ordered roughly by value. **[P1]** immediate, **[P2]** next, **[P3]** later/roadmap, **[OWNER]** needs an owner call, **[XREPO]** cross-repo.

**Docs / the mandate itself**

~~1. **[P1]** ANNOTATE the 20 recent (2026-09-30/10-01) reports once their trains close, then archive the fully-resolved ones.~~ done — this v6 sweep (the 34-file annotate + archive program complete)
~~2. **[P1]** Decide the four owner-gated plans (`error-excellence`, `20-year`, `13-33`, `04-29`): archive (routed-as-resolved) or annotate-and-keep. (v4 `g1`.)~~ owner — the standing q3 (keep live pending the sitting)
~~3. **[P1]** Resolve/ratify the 37-file check-rows marker-cell residue — fix `check-rows.py` to accept the marker column, or migrate the corpus.~~ owner — the g2 tooling question
~~4. **[P2]** Convert TODO_LIST maintenance into a committed parse/emit helper so regeneration is deterministic.~~ not adopted — below the bar (no committed helper)
~~5. **[P2]** Add a standing DOM-contract count check (or drop the number from FEATURES) to stop the 34→35→38 recurrence.~~ done in part — the DOM-contract test is the standing gate; the FEATURES number stays hand-cited
~~6. **[P2]** Load `agents-quality-guide.md` and run the 5-dimension AGENTS rubric; feed the compaction row.~~ not adopted — below the bar (no rubric run recorded)
~~7. **[P2]** Re-verify the post-move archived dirs are link-clean (a full markdown link check, not just cited-path greps).~~ done — the v6 sweep link-checks the touched files at its close
~~8. **[P3]** Reconcile any duplicate listings between the new ROADMAP questions and the owner-calls TODO row.~~ process record — reviewed at the v6 harvest-parity re-read

**Living docs still to fold/sync**
~~9. **[P2]** Fold `[Unreleased]` → a version section on the next release (the UI/UX + paperless + mic + honesty batch all ride it).~~ routed — TODO health row (the v2.9.0 fold decision, owner)
~~10. **[P2]** FEATURES: flip the mic pre-warm row to FULLY_FUNCTIONAL once a real call verifies it.~~ routed — the flip gates on the TODO mic-row live-call ritual
~~11. **[P3]** README FAQ for "why answering is fast now / what the mic indicator at ring means" (04:07 §f 30).~~ not adopted — below the bar
~~12. **[P3]** Update `docs/release-runbook.md` with the stack-E2E obligation for the UI/UX markup change (06:59 §f 12).~~ done — the runbook rule landed
~~13. **[P3]** `docs/dom-contract.md` prose note for `wp-live` / `wp-tab-skeleton` (06:59 §f 14).~~ done — dom-contract lists both ids

**Verification / gates**
~~14. **[P1]** Run `buildflow` (full) + `nix flake check` + a fresh smoke boot on current HEAD (no gate ran on the UI/UX train).~~ resolved by events — the batteries ran (T03 and later)
~~15. **[P1][XREPO]** Re-run the stack browser E2E (served markup changed: skeleton, palette, mobile bar, day separators).~~ routed — TODO island-honesty row (the owed stack browser E2E)
~~16. **[P1]** Add the owed pins: `aria-current="page"/"false"`, skeleton reveal/hide, optimistic-bubble-morph edge.~~ done — the pins landed (master-todo session 2)
~~17. **[P2]** Fix the overpromising `#wp-live` "connection recovery" comment in `layout.templ`.~~ done — the comment fixed (session 2)
~~18. **[P2]** Run the KVM `webphone-backup` VM test + `webphone-backup-drill` once post-nix-split.~~ done — ran green (the 20:12 battery)
~~19. **[P2]** `nix run .#vulnix` + aarch64 ELF verify on HEAD post-UI/UX train.~~ done — T22 close-out (vulnix zero-real-advisories + aarch64 exit-green)

**The 2026-10-01 product trains (from the harvest)**
~~20. **[P1]** Live-call ritual (accept→speak, ICE path/rtt, MOH audibility) — retires ~5 open items.~~ routed — TODO mic row (the live-call ritual, owner terminal)
~~21. **[P1]** Ring-silence `AudioContext` gesture fix + pin (owner authorization pending).~~ done — T06 shipped (owner ruled KEEP)
~~22. **[P1]** Island boot language split-brain fix (read `wp-lang` cookie) + `TestShellHtmlLangFollowsSessionLang`.~~ done — T05 shipped + pinned (TestShellHtmlLangFollowsSessionLang)
~~23. **[P2]** Mic pre-warm live verification + `devicechange` re-warm + the three pin tests.~~ done in part — the pins landed; live verification rides the TODO mic row; devicechange rides ROADMAP
~~24. **[P2]** Island-honesty follow-ups: split `holdFailed`/`resumeFailed`, clear `holdPending` on Terminated/watchdog, VM-timeout root-cause, errcheck dispatch.~~ done — T15 landed (the split + Terminated/watchdog clearing); the VM root-cause was not adopted
~~25. **[P2]** UI/UX M9 dial affordances; M12 segment countdown; M11 history filters.~~ done — T12/T13 shipped
~~26. **[P3]** UI/UX M13 voicemail playback, M14 fax depth, M15 visual tokens, M16 URL state.~~ done — T14/T16 shipped
~~27. **[P3]** UI/UX M17 feedback/trust (reconnect banner, undo, retry-in-banner, confirm consistency, spinner, success pulse).~~ done — T15/T17 shipped
~~28. **[P3]** UI/UX M18 onboarding/demo, M19 mobile extras, M20 theming depth, M24 i18n/RTL, M25 call depth, M26 shell sizing.~~ done — T16/T17/T18/T19 shipped
~~29. **[P3]** Perf plan: ETag+304 for `/assets/*` (reconcile `server.go` composite-ETag first).~~ done — T10 shipped (content ETag + 304)
~~30. **[P3]** Perf plan: scoped gzip for static handlers; `modulepreload` for the island graph; curl timing baseline.~~ done — T10/T11 shipped (scoped gzip, modulepreload, perf-baseline.py)
~~31. **[P3]** Outgoing-call mic warm (mirror of the shipped incoming fix).~~ done — the dial-focus mic warm shipped
~~32. **[P3]** `iceServers` trimming eval; trickle-ICE investigation (named-trigger).~~ not adopted — the trim-eval note stands; trickle stays policy-rejected

**Nix / tooling**
~~33. **[P2]** Commit the module-output golden (vhost + backup scripts) + a new check entry (05:56 §f 1).~~ done — nix/module-output.golden + the check entry
~~34. **[P2]** Release-time `webphoneVersion` ↔ `git describe` guard in release.sh (with the bump-before-tag edge).~~ done — release.sh tag↔webphoneVersion guard (override documented)
~~35. **[P3]** `actionlint` over `.github/workflows/ci.yml`; dedupe `devShells.ci`/`default` Go env.~~ done — actionlint ran clean + the devshell dedupe landed (T21c)
~~36. **[P3]** Record the accepted exceptions (hardcoded version, module-check permissiveness) + the declined `go-standard` migration decision.~~ done — recorded in the CHANGELOG nix-review batch-2 entry (T21e)
~~37. **[P3]** Commit `health.css` + its build script; wire the CI rebuild (02:12 §f).~~ done — build-health-css.sh + checks.health-css canary (TODO health row)
~~38. **[P3]** `nix/README.md` section index for the 8 `nix/` files.~~ not adopted — below the bar

**Samber/do + dashboard**
~~39. **[P3]** Add `family_test.go` pins for the new error codes (dashboard train).~~ done — T22 family pins
~~40. **[P2]** Investigate the local-main-behind-remote divergence (`17684fc` vs `241b908`).~~ done — documented as LIVE daemon behavior (TODO row)
~~41. **[OWNER]** Decide the stack `/health` exposure policy.~~ owner — TODO health row (the /health exposure policy)

**Cross-repo (stack / pbx-artmann)** — handovers only, no assistant ssh
~~42. **[P1][XREPO]** Repair the stack FreeSWITCH `mod_enum` build → E2E → relock to `cc98c2e`+.~~ routed — TODO cross-repo row (the repair-first chain)
~~43. **[P2][XREPO]** Ops-runbook demo-call recipe + `/var/lib/telephony-secrets/` path; `deploy.md` secret PATH column.~~ routed — TODO cross-repo row
~~44. **[P2][XREPO]** Check MOH audibility + `/recordings/` + CDR.~~ routed — TODO cross-repo row
~~45. **[P3][XREPO]** WebTransport-not-adopted verdict doc; gateway stack-side bits (`ftypqt` sniff, MMS-outbound, FEATURES:87).~~ routed — TODO cross-repo row
~~46. **[P3][XREPO]** Stack `services.webphone.paperless` module option + smoke arm.~~ routed — TODO cross-repo row
~~47. **[P3][XREPO]** Route the gomod-check + codespell false positives upstream in BuildFlow.~~ other repo — BuildFlow

**Owner/process**
~~48. **[OWNER]** The owner-calls sitting (28-row briefing) — ratifies v2.8.0 + setup NO-GO, grants AGENTS compaction.~~ routed — TODO owner row (the sitting)
~~49. **[OWNER]** deploy v2.8.0 + pbx-artmann relock #5; SMS-bridge journal leg.~~ routed — TODO deploy-tail row
~~50. **[P3]** Add a committable docs-health helper (TODO parse/emit + archive order guard) so the next sweep does not re-derive this session's fixes.~~ not adopted — below the bar (no committed helper)

---

## g) Three questions I CANNOT answer myself

~~1. **Archive the four owner-open plans or keep them live?** The v4 report asked (`g1`) whether "routed-as-resolved" suffices for archiving `error-excellence`, `20-year-durability`, `13-33 stack-adoption`, and `04-29 pareto`. Their remaining items now live in the TODO/ROADMAP rows. Ratify "archive with routed markers", or keep them live until their owner-open items close?~~ owner — still open (the v6 default: keep live pending the sitting)
~~2. **Fix `check-rows.py` or migrate the corpus?** 37 archived files false-flag on the older marker-cell format because the two skill assets disagree on what "struck" means. Should I patch the tool to accept the marker column (one-file fix, keeps history untouched), or restyle the corpus to the strict uniform format (much larger, edits historical snapshots)?~~ owner — still open (the 37-file residue question)
~~3. **Is the docs-only-delta gate scope acceptable?** I ran the marker/link gates + scoped `go test cmd/webphone internal/server` (precedent), but NOT `buildflow`/`nix flake check`/smoke. Do you want the full battery run over this docs commit anyway (it costs a quiet-machine window), or is the scoped precedent the standing answer?~~ process record — the scoped precedent stood (the v6 sweep also runs scoped)
