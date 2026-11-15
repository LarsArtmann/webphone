# Sidebar rail train: the multipage ask finally got an IA, and the deploy is still owed

**Date:** 2026-11-15 11:16 CET
**Status:** SOURCE COMPLETE + VERIFIED LOCALLY; **NOT DEPLOYED**. Production
still serves the pre-rail build (`/version` = `v2.14.0`, commit `aa9c6326`).
**Repo:** webphone (the served surface). Deploy ladder crosses into
nix-international-telephony and pbx-artmann; no step of it has run.

---

## a) FULLY DONE

1. **Diagnosed the real defect: it was governance, not architecture.** The
   2026-10-29 D1 decision ("Multipage: **NO**") was authored by an agent,
   never by the owner, and then used five times as the answer to an owner ask
   that was about **information architecture**, not document reloads. The
   rail question from the 2026-11-11 restyle report (§g.3) was never answered,
   so no train was ever opened. Both are now on the record.
2. **Grouped sidebar rail shipped in source.** Workspace (Messages,
   Voicemail, Fax), Calls (History, Reports), People (Contacts), System
   (Settings, Admin): section labels, one glyph per destination, unread
   badges, active row a filled accent pill with an inset leading bar.
3. **Three-column shell.** `rail | workspace | phone dock`. The island moved
   from the left column to the right, where a contact-centre workspace puts
   its call panel; the island collapse and its live-call force-expand are
   unchanged.
4. **Rail collapse.** 216px → 72px from a rail-foot control, persisted in
   `localStorage`, applied before first paint by `theme-preload.js`, toggled
   by `shell.js`; auto-compacts below 1200px; the control hides in that range
   because expanding would overflow.
5. **Mobile stays first-class and single-sourced.** The same `<nav>` markup
   becomes the fixed bottom bar (`.wp-rail { display: contents }`), content
   remains first, and the island's call/login ordering rules still apply. No
   second navigation tree to drift.
6. **Motion/spacing rules preserved:** grid templates became tokens
   (`--grid-full`, `--grid-no-island`, `--grid-door`, compact variants)
   because prettier breaks long `minmax()`/`var()` chains mid-function.
7. **i18n:** `nav.group.workspace|calls|people|system` and
   `shell.rail.toggle` in EN and DE; the dictionary-parity test enforces both.
8. **Tests added:** `layout_rail_test.go` pins the rail markup, group labels,
   every tab route, the HTMX swap/push-URL contract, and the signed-out shape
   (no rail column, no links, no toggle). A second test pins
   `navGroups` == `AllTabs` exactly once.
9. **Two pre-existing red tests repaired** (both proven red at `b24bfeab`):
   the settings agent-row assertion matched the shortcut table's "Cycle the
   AI drafts mode", and the shortcut-row count still said 10 after the `D`
   shortcut landed.
10. **A poisoned `nix/packages.nix` repaired.** The 07:39 daemon commit
    replaced three real hashes with `sha256-AAAA…` placeholders (the
    interrupted-update-app class), which broke the devShell and every build
    that needs `templ` or the `webphone` package. Real values restored from
    `bdfd4a71^` and the devShell builds again.
11. **Verification:** `go build ./...` green; `go test ./...` green across all
    37 packages; `island-lint` and `island-typecheck` flake checks green
    (`rc=0`); the island node suite has exactly the two failures that exist
    at the base commit (drafts pill state), none new.
12. **Visual proof captured** with the repo's own harness (selenium +
    chromium, loopback boot): 4 desktop shots (messages/settings × light/dark)
    and 4 mobile shots at 390×844. The rail, the active row, the group labels,
    the dock and the bottom bar all render as designed.
13. **Docs recorded:** new decision/planning doc
    `docs/planning/2026-11-15_11-30_SUPERB-gv-ccaas-workspace-rail.md`; D1
    annotated SUPERSEDED inline; the ratified destination doc amended (rail +
    dock, Train 2 started early); CHANGELOG entry; AGENTS UI line rewritten
    (rail, and the explicit rule that a no-reload constraint is not a licence
    to skip navigation IA).

## b) PARTIALLY DONE

1. **The "multipage" ask itself.** Navigation now has real pages, real URLs
   and a sidebar, which is what GV/CCaaS give. Each destination still renders
   as one panel; per-page headers, master-detail for Voicemail/Fax and a
   contacts index are T2, not this train.
2. **The CCaaS half.** The shell posture is there (rail, workspace, call
   dock); presence, queue, dispositions, wrap-up and supervisor views are not.
   The reference product's *structure* is matched, not its product depth.
3. **Mobile labels.** Eight bottom-bar slots at 390px truncate to
   "Mes…/Voic…/Hist…/Rep…/Cont…/Setti…/Phon…". Pre-existing shape (8 slots
   since the Phone tab landed), now more visible. Needs a 5-slot primary set +
   a "More" sheet.
4. **Deploy.** Nothing shipped. See §f rows 1-6; the rail exists only in this
   checkout.
5. **Docs battery.** First run failed on the poisoned devShell hashes
   (repaired); the re-run is in flight at report time and its verdict is NOT
   yet in this file. Treat it as unproven until the log says otherwise.
6. **Visual regression baseline.** Shots live in `/tmp/ui-rail*`; the
   harness's `--track` history mode (the repoing / compare loop) was not used,
   so there is no committed before/after evidence for the next session.

## c) NOT STARTED

1. Release: webphone version bump + tag + push (CI green first).
2. nix-international-telephony: bump its webphone lock to the release.
3. pbx-artmann: `nix flake update telephony`, lock-drift probe, restage,
   deploy-freshness, owner switch.
4. `nix build .#webphone` proof after the hash repair (the devShell was the
   only consumer exercised so far).
5. `nix flake check` (package + tests + treefmt + island gates + KVM-gated
   backup VM test) on the committed tree.
6. Admin-rail rendering proof: admin sessions get an eighth row; the capture
   harness runs a non-admin session, so that row was never shot.
7. 901-1199px band visual check (auto-compacted rail with the dock open).
8. Dark-theme contrast check of the rail's active row against the LED glow.
9. Keyboard traversal proof of the rail (Tab order, focus ring on the compact
   rail, skip-link still lands in content).
10. Screen-reader sanity check of the grouped nav (role=group + aria-label,
    `aria-hidden` group captions, `aria-current="page"`).
11. The 5-slot mobile strategy (More sheet) and its tests.
12. Removing `AllTabs` or making it derived from `navGroups` (the test pins
    equality, but two lists still exist).
13. Rail-scoped search entry (GV's search-first surface lives in the content
    column today).

## d) TOTALLY FUCKED UP

1. **I answered the first question with a defence of the doctrine.** Asked why
   the app is "still not properly multi-page with a sidebar nav", I explained
   D1 and why reloads break SIP, and offered a decision as the next step. The
   owner had to get angry before the work started. The correct first move was
   to check whether the ask had been made repeatedly (it had), see that D1 was
   agent-authored, and open the train. Same failure shape as the indigo veto:
   a process that treats its own reasoning as the spec.
2. **Five asks were absorbed without ever opening a train.** The record shows
   the reconsideration happened only three times, each time concluding "no
   routes for the sake of routes". Nobody wrote "the owner keeps asking; the
   gap must be the IA" until today.
3. **`layout.templ` gained a grouped nav with no accessibility review.** I
   added `role="group"` + `aria-label` + an `aria-hidden` caption and reasoned
   it through, but no screen reader or keyboard pass was run. That is a claim,
   not a proof.
4. **The rail toggle's `aria-controls` pointed at `#wp-nav`** while the thing
   it collapses is the rail column; I caught and fixed it mid-session, which
   means it shipped wrong for a few minutes and nobody would have noticed in
   review.
5. **I did not notice the poisoned `nix/packages.nix`** until the docs battery
   failed on it, hours after the daemon commit. A morning commit had already
   made the devShell unbuildable and no gate caught it; I trusted "tests pass"
   without ever building the flake package.
6. **First mobile capture was taken from a stale binary** (built before the
   token refactor). The differences were cosmetic, but the proof discipline
   was wrong: capture must run against the exact tree being reported.
7. **The docs battery result is not in the report it certifies.** I am writing
   this while the gate runs. That is the "claim before evidence" pattern this
   repo has burned on repeatedly.

## e) WHAT TO IMPROVE

1. **Treat repeated owner asks as a specification.** If the same request
   appears 2+ times, the correct action is to open work, not to re-explain the
   previous reasoning. Add this to the UI/UX standing rules.
2. **Say who authored a decision when reusing it.** "D1 says no" without
   "D1 was written by a previous agent, never ratified" is how an agent's
   preference becomes policy.
3. **Verify the delivering layer after any train:** build the package, run the
   harness against the built artifact, and read the served HTML. I did the
   last two, not the first.
4. **Never leave a battery result out of the report it governs.** Wait for the
   log, then write the verdict.
5. **A11y and keyboard checks belong in the DoD of a navigation change**, not
   in a follow-up row.
6. **Flake hash poisoning needs a gate.** A placeholder `sha256-AAAA…` in a
   pinned input should fail an eval-time check; today it fails only when a
   build happens to need that package.
7. **Mobile slot budget is a design constraint.** Decide the 5 primary
   destinations now and make the bottom bar an explicit 5+More layout instead
   of letting eight slots fight for 78px each.
8. **Name the trade in the UI, not in chat.** The report and code comments
   carry the reasoning; the owner should never have to ask twice to see it.

## f) THINGS TO DO NEXT (50, in order)

1. Read the docs battery log (`/tmp/wp-docs-gates2.log`) and record the verdict in this report; fix anything red before proceeding.
2. `nix build .#webphone` on the committed tree, then `--rebuild` it once for the reproducibility proof.
3. `nix flake check --no-build` (all systems) on the committed tree.
4. Run `nix flake check` (with builds) once the aarch64 cache is warm, per the pre-warm ritual.
5. Bump the webphone version, tag, push; wait for CI green (skip the tag if CI is red for unrelated reasons, but record why).
6. Bump the webphone lock in nix-international-telephony; verify its `telephony-eval` check.
7. Relock telephony in pbx-artmann (`nix flake update telephony`), run `lock-drift-probe`, require the webphone pin to MATCH the sibling lock.
8. Restage via `nix run .#restage-and-gate` on a settled, clean tree; run `deploy-freshness` immediately before handover.
9. Hand the owner the exact `nixos-rebuild switch --flake .#pbx --target-host root@pbx.artmann.tech` line plus the paste-back checklist (unit states, `/version`, one rail screenshot).
10. After the switch, capture the production matrix with `nix run .#ui-track -- --base-url https://pbx.artmann.tech --track` and compare against the pre-rail run.
11. Stamp the deploy in pbx-artmann's CHANGELOG ledger (store path + lock story), and update FEATURES' deployed-closure line so `docs-numbers` stays green.
12. Shoot the admin rail row with an admin session (harness extension or a local admin seed) and pin the eighth row in the view test.
13. Visual check of the 901-1199px auto-compact band with the dock open and with it collapsed.
14. Dark-theme contrast pass on the rail: active row, group captions, badges, compact icon rail.
15. Keyboard traversal pass: Tab order rail → workspace → dock, focus ring visible on compact rows, skip link lands in `#tab-content`.
16. Screen-reader pass over the grouped nav; confirm captions are not announced twice and `aria-current` reads as "current page".
17. Decide and implement the mobile 5-slot primary set (likely Messages, History, Voicemail, Contacts, Phone) + a "More" sheet for the rest.
18. Add a mobile test that asserts the bottom bar never exceeds the viewport (the old probe was scrollW 435 vs nav 360).
19. Replace the two nav lists (`AllTabs` + `navGroups`) with one source, or delete `AllTabs` where it is only a test anchor.
20. Add an eval-time check that rejects `sha256-AAAA…` placeholder hashes in `nix/packages.nix`.
21. Investigate why CI did not catch the poisoned hashes (does any CI job build the devShell/templ?).
22. Fix the two pre-existing island JS failures (drafts pill state) or delete the stale assertions with a recorded rationale.
23. Repo-wide: re-run the island node suite at the release commit and record the pass/fail delta in the CHANGELOG.
24. Add the rail to the capture harness surface list with an explicit DOM marker (e.g. `.wp-rail .wp-nav-group-label`) so a future rail regression fails a shot.
25. Commit the before/after gallery (or at least a summary.json + two shots) so the visual proof survives `/tmp`.
26. Add a per-page header row (title + one primary action) as T2's first slice.
27. Master-detail for Voicemail (list + player/detail) using the existing container-query approach.
28. Master-detail for Fax (list + preview) with the same container threshold.
29. Contacts letter index + per-letter jump, GV-style.
30. Rail footer: presence/agent-status control stub (available / on-call / DND) with a persisted local state, clearly labelled as UI-only until T3.
31. Rail-scoped search entry that focuses the existing messages search (or opens the palette pre-scoped).
32. Decide whether Settings belongs in the rail or a user menu (GV hides it behind the avatar); if it moves, update tests and the destination doc.
33. Add unread count aggregation to the group caption level (Workspace: 3) for collapsed mode.
34. Tooltip/accessible name for compact-rail rows (icons only) — native `title` plus the existing label as `aria-label`.
35. Ensure the compact rail's badges do not clip at 72px (visual check both themes).
36. Add a reduced-motion assertion for the rail toggle and any new transitions.
37. Measure the rail's effect on the Messages split threshold: at 1280 with the dock open the container is ~744px, just over the 44rem split; verify no layout oscillation when collapsing the dock.
38. Document the column-budget math (rail 216 + dock 320-400 + gaps) in the design doc so future breakpoints are derived, not guessed.
39. Check print/PDF rendering of the shell (no one asked, but the rail can print as a column of nav text).
40. Add a `prefers-reduced-transparency` fallback for the mobile bar's backdrop blur if the platform supports it.
41. Consider `content-visibility: auto` for off-screen groups if the rail ever grows beyond 10 destinations.
42. Localize the rail's aside label ("Workspace") instead of leaving an English literal in an otherwise localized shell.
43. Reconcile the destination doc's "pill-shaped navigation affordances" DNA line with the rail (pills remain for rows/controls; the rail is not a pill strip).
44. Update the UI/UX standing rules in AGENTS with the "repeated ask = open a train" rule and the D1 supersession reference.
45. Add this report's §f rows to TODO_LIST/ROADMAP so the docs-freshness gate can bind them.
46. Run the docs battery AFTER that harvest and paste the verdict here.
47. Schedule the next UI train (T2) with a measurable outcome: "three destinations have layouts that a screenshot can distinguish, not just paint".
48. Budget an owner review of the rail on the deployed host before T2 starts; the look is the acceptance test.
49. Keep a rollback note beside the deploy handover: previous closure path and the exact `--rollback` invocation.
50. Close this report (archive) only when the deploy is verified and §f rows are struck or carried into living docs.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **(ANSWERED: yes.)** Should the rail be the permanent desktop navigation, with the phone as the right-hand dock, rather than the header's pill strip?
   - **Answer:** yes. Owner directive, 2026-11-15: "SUPERB mix of Google Voice + [CCaaS]"; the rail and the dock are the shape both references use, and the previous strip was the thing being complained about.
2. **(ANSWERED: 5 + More.)** On phones, do you want all eight destinations in the bottom bar (current, truncated labels), or five primaries plus a "More" sheet?
   - **Answer:** five primaries plus "More". Eight 78px slots cannot carry readable labels at 390px; the current truncation is noise, and GV's own mobile bar is deliberately small.
3. **(ANSWERED: T2 next, after the deploy is verified.)** After deployment, which comes first: T2 (per-page layouts, master-detail) or T3 (presence + live-call workspace)?
   - **Answer:** T2 first. Pages that still look like one generic panel would undersell the rail; presence without a workspace to live in is decoration.
