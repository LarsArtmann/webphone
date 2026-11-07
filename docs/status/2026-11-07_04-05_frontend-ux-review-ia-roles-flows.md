# Frontend UX review — IA/roles/flows extension train (session status #2)

> Snapshot: claims verified as of 04:05, 2026-11-07. Covers this session's
> work AFTER the 02:09 status report
> ([2026-11-07_02-09_frontend-design-ux-review-train.md](2026-11-07_02-09_frontend-design-ux-review-train.md)):
> the owner's challenge that the review was missing the whole layout design —
> pages, roles × UX, BDD flows — and the work done in response. Repo state:
> review report has UNCOMMITTED modifications (daemon mid-cycle), 3 commits
> not yet pushed, tree otherwise clean.

## What this phase was

The owner's push ("you're missing the whole layout design — what pages
should we have? how should roles shape the UI? what are the user flows,
BDD-style? THINK HARDER") named the real gap: the review audited components
and tokens but never modeled the PRODUCT. The fix, now in the report as
three new chapters:

- **§03 Information architecture & layout** — the zone-by-zone layout
  assessment (header / island / main column / overlay / feedback layers),
  the page inventory, **two additions** (focused signed-out layout;
  Settings split by audience), and **two recorded rejections** (a "Calls"
  tab — would split-brain call state with the island; a separate onboarding
  route — the dismissible welcome panel already owns it).
- **§04 Roles & permissions × UI/UX** — five identities (anonymous,
  extension owner, multi-line, admin, operator) + the capability-flag axis
  in a matrix against the current UI, plus the four-rule doctrine the
  codebase mostly obeys (render-absence over disabled controls; roles gate
  visibility / capabilities gate affordance, server-rendered; cross-identity
  access read-only and chipped; writes need your own session).
- **§05 User flows, BDD-style** — ~45 Given/When/Then scenarios across 9
  features, each tagged with its behavioral pin (`island-tests/*.test.mjs`,
  Go tests, smoke scenarios) or **[GAP]**; all 8 GAPs map 1:1 to finding
  cards, making each finding one scenario away from being a pinned test.
- **Two new finding cards** the IA thinking surfaced: **F19** (signed-out
  page shows the full desk with dead tabs — a floor plan instead of a door)
  and **F20** (Settings serves two masters — operator truth, personal
  preferences and personal data in one column).
- Sidebar nav extended (Fix queue / IA & layout / Roles × UI / User flows),
  appendix renumbered 06, scorecard updated (15 important, 20 carded ·
  matrix pending).

## a) FULLY DONE

- IA chapter §03 complete: zone table, page inventory, 2 additions, 2
  explicit rejections with rationale (rejected ideas are recorded so they
  stay rejected).
- Roles chapter §04 complete: identity × UI matrix, capability axis, the
  R1–R4 doctrine, per-role gap notes (anonymous = F19, Settings = F20,
  admin search + read-only badge carried in matrix items).
- Flows chapter §05 complete: 9 features, ~45 scenarios, pin-mapped, 8
  named GAPs (focus-on-replace, ring-time names, drafts-poll dirty guard,
  prefill gating, palette reachability + ARIA, logout-with-call confirm,
  tab-failure visibility nuance).
- Finding cards F19 + F20 written with evidence and fixes.
- Scorecard/nav/summary reconciliation for the new sections.
- Structural validation re-run: 20 unique issue ids, 9/9 sections nav-
  linked, all container tags balanced, 10 gherkin blocks, 8 GAP tags.
- Honest acknowledgment to the owner that the IA layer was missing — the
  critique was accepted, not defended.

## b) PARTIALLY DONE

- **The report as a whole (~92%)**: the appendix matrix (the ~25 non-carded
  findings from the three agents) is still unwritten, and the scorecard's
  "20 carded · matrix pending" placeholder still needs its true total.
- **Condensed agent re-emits**: dispatched hours ago; agent_status now shows
  all three agents finished their follow-ups, but the condensed reports
  never woke this session — from here the matrix must be written from the
  agents' original full reports (received and cited) or a third request.
- **Self-review discipline**: the two daemon-era edit failures (mtime guard
  tripping on daemon commits under me) were handled correctly but slowly —
  the fix (re-View before edit) is known and should be reflexive.
- **Status hygiene**: report #1 (02:09) was written but the status index was
  NOT regenerated at the time — fixed by this report's publication step
  below.

## c) NOT STARTED

- Appendix matrix + scorecard total reconciliation + full rendered
  (browser) verification of the report.
- `python3 scripts/status-index.py` regen covering both status reports
  (executing right after this file lands).
- All 50 fix-queue items — this remains a review-only train; zero product
  files touched (the only modified file is the report itself).
- Runtime verification of the 8 [GAP] scenarios and the hypothesis-flagged
  findings (palette containment, double-Enter, live-context race,
  drafts-poll loss).
- The IA fixes themselves (F19 anonymous layout, F20 Settings grouping) —
  designed, not implemented.

## d) TOTALLY FUCKED UP

No corruption, no data loss, repo healthy — but honest failures, led by the
one the owner caught:

1. **The review shipped without its top layer.** I audited tokens, markup
   and JS — the artifacts — and never modeled the product: no site map, no
   role matrix, no journey specs. Root cause: I let the component-level
   rubric (CSS, a11y, forms) define the review's ceiling. A design review's
   first chapter is the layout and IA; mine opened with contrast ratios.
   The owner's "THINK HARDER" was correct and the fix took one focused
   pass — meaning the gap was choice of lens, not lack of material.
2. **Condensed-reemit follow-ups are lost in delivery.** Two of three
   agent wake-up reports never surfaced despite the agents finishing; hours
   were spent assuming arrival instead of verifying it. The lesson from
   status report #1 ("verify report arrival before merging") was applied
   too late to save the wait.
3. **Edit friction with the daemon**: two multiedit rejections (file
   modified under me by auto-commits) and the earlier lost-strength edit —
   each recovered, each preventable by an immediate re-View reflex.
4. **Idle waits**: the session sat ~2h on agent wake-ups that never came;
   a deadline ("write the matrix from originals at T+30min") would have
   closed the deliverable instead of stalling it.

## e) WHAT WE SHOULD IMPROVE

- **Structure-first reviews**: next design review opens with IA + roles +
  flows (the §03–05 chapters), then components. The owner should never
  have to ask for the layout layer.
- **Map flows→pins while the agents run**: coverage mapping belongs in the
  review agents' task lists, not in a later synthesis pass.
- **Agent wake-up deadlines**: if a follow-up report doesn't arrive within
  a fixed budget, proceed from the originals and say so in the artifact.
- **Verify delivery, not dispatch**: agent_status "finished" means the
  agent stopped, not that the report reached the session.
- **Product-side**: the report now carries the full improvement queue —
  headline items unchanged (dark `--on-accent` flip; island screen-reader
  pass; palette reachability + combobox; ring-time names), joined by the
  two IA corrections (F19, F20).

## f) NEXT TASKS (up to 50)

1. Write the appendix matrix from the agents' original reports; reconcile
   the scorecard total; drop the "matrix pending" placeholder.
2. Browser-render verification pass of the full report (anchors, scrollspy,
   both themes of reading comfort).
3. Regen `docs/status/INDEX.md` (both status reports indexed).
4. Decide + implement the dark `--on-accent` flip (F1) in both token
   blocks; re-run contrast math + ui-capture.
5. Add the `wp-service-off` arm to `crmDetail` (F4, one line).
6. Point layout.templ + manifest at the real palette colors (F2, both
   files one fact).
7. Implement the signed-out focused layout (F19) — nav hidden/collapsed
   when anonymous, welcome + login centered; DOM contract ids preserved.
8. Split Settings into My preferences / My data / This server (F20).
9. Gate "Prefill from AI draft" on the seam prop (F6).
10. Unify English wall-clock on Kitchen format (F5).
11. Remove `role="img"` from the four row surfaces (F10).
12. Drafts-poll skip-if-dirty guard (F11).
13. Palette combobox ARIA (F12).
14. Palette coarse-pointer entry button (F12).
15. Segment counter UTF-16 units (F13).
16. Island focus-on-replace helper + five call sites (F8).
17. Incoming-call caller-name resolution (F7).
18. `role="status"` on `#dial-error` (F9).
19. Drop the live region from `#reg-status` (F9).
20. Localize island-built aria-labels (F9).
21. Fix the malformed incoming-call announce (F9).
22. Icon-action labels/confirm pattern for thread head + History rows (F3).
23. Chip-X 24px hit area (F14).
24. History empty state rewording (F15).
25. Type-scale consolidation tokens (F16).
26. Radius/spacing token adoption (F17).
27. Uppercase-voice zoning + reconnect-pill copy (F18).
28. Admin tab: `q` filter + `hx-push-url` + a read-only badge.
29. Brand as the page's `h1`.
30. `tabindex="-1"` on panel-head `h2`.
31. List semantics for rail + row collections.
32. `maxlength` hints aligned with server caps.
33. `htmx-indicator` on submits beyond fax.
34. Contacts forms `hx-disabled-elt`.
35. i18n keys for hardcoded panel aria-labels.
36. Emoji/dingbat glyphs → icons package; strip ✖ from `history.missed`.
37. German rewords (Stumm aus / hat abbestellt / Rechner / DE noMatch
    recovery / contactsMigrated singular).
38. Vocabulary drift: Mailbox unify; "transcript lines" rename.
39. Login in-flight guard + connecting pill state.
40. i18n key for the hardcoded 429 toast.
41. `beforeunload` guard while calls are live.
42. In-call `document.title` signal.
43. `agoLabel` "just now" branch (kills "Last call now ago").
44. Voicemail network-failure copy; surface silent `!res.ok` paths.
45. Logout confirmation on live calls.
46. Mute/hold/resume settled-state announces.
47. `.catch` + `tabRefreshFailed` on bare `htmx.ajax` calls.
48. Transfer row: empty-destination error + Escape closes.
49. Toast dismissal affordance.
50. Turn the 8 [GAP] flow scenarios into pinned tests (the flows chapter
    is the executable spec — this closes the loop).

## g) QUESTIONS (I cannot answer these myself)

1. **Review-only or fix train?** The review is now ~92% complete; the
   50-item queue is fully specified. Do you want Tranche-1 fixes (F1 token
   flip + F4 one-liner, re-shot) implemented in this session after the
   report closes, or does fixing start as its own owner-initiated train?
2. **F1 fix direction (taste)**: dark `--on-accent` → dark ink (my
   recommendation: keeps the vivid dark accent, matches the hardware-key
   vernacular, one token) vs darkening the dark accent fill to ~#108351
   with white ink? The report recommends the ink flip; the palette identity
   is yours to overrule.
3. **F19's shape (product call)**: for the signed-out page, do you want the
   tab row GONE entirely (pure door: welcome + login only — my
   recommendation), or tabs kept but visibly inert (preserves the "there
   is more here" preview at the cost of promising what's locked)?
