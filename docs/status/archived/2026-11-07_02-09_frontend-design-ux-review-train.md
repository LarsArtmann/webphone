> ARCHIVED 2026-11-10 (docs-health drain): gates clean (check-rows --strict + bare-rows green), live facts carried in AGENTS/TODO_LIST — cadence headroom for the report flood.

# Frontend design & UX review train — session status

> Snapshot: claims verified as of 02:09, 2026-11-07. This report covers ONLY
> this session's run (the frontend design + UX review train). Repo state at
> writing: tree clean, HEAD `a8d089ba`, one commit not yet pushed
> (`origin/main` at `5b0b8bff`) — the daemon is committing the review report
> incrementally as I write it.

## What this session is

A whole-product frontend design and UX review of the served UI: both
stylesheets read in full, every templ panel skimmed via three scoped review
agents (views / island JS / shell), the visual harness run live (24
screenshots, layout assertions PASS), modern-web-guidance consulted for the
review criteria, and my own contrast math on the token pairs. Deliverable: a
Bauhaus-kit HTML report at
[docs/review/2026-11-07_01-51_frontend-design-ux-review.html](../../review/2026-11-07_01-51_frontend-design-ux-review.html)
— currently ~90% complete (all 18 finding cards, 12 strengths, fix queue
written; appendix matrix outstanding).

## a) FULLY DONE

- **Design-system core read in full**: `app.css` (2,920 lines),
  `island/style.css` (1,269), `layout.templ`, `theme-preload.js`,
  `manifest.webmanifest`, `settings.templ` (status-row region).
- **Contrast audit, measured**: every token pair computed from the served
  values. Result: everything passes WCAG except one systematic dark-theme
  failure — white ink on `--accent` fills = 3.21:1 (buttons, outbound
  bubbles), 1.87:1 on hover (worse than rest!), 2.52:1 bubble meta, 3.04:1
  on danger fills. Fix quantified: flip dark `--on-accent` to dark ink →
  6.06–10.42:1 everywhere, one token, both mirrored token blocks.
- **Visual harness run**: `scripts/ui-capture.py` over the built 2.9.0
  binary — 24 shots (12 surfaces × light/dark + collapsed), harness's own
  layout assertions PASS (collapse, persistence across reload, Messages
  split threshold at six viewport widths). Key shots inspected visually:
  thread light/dark, login, welcome, history transcript, contacts, fax,
  settings, voicemail, 404, media-transcribe.
- **One misread caught and retracted before publication**: a thumbnail-scale
  read of the media-transcribe shot suggested danger-red-on-green text;
  pixel sampling (ImageMagick histogram) proved the line renders white
  (#F7F8F9 on #0E8752) — no defect. The retraction discipline worked.
- **Three subagent reviews completed and received**: island JS (20 findings,
  verbatim in hand), views (18), shell (22). Their strongest verified
  findings are merged as report cards.
- **The report's written bulk**: head/hero/meta, scorecard, summary (2
  paragraphs + bottom line), 12 verified strengths, **18 finding cards**
  (F1–F18: one defect, twelve important, five nice), the 3-tranche fix
  queue, footer, scrollspy intact. HTML validated: 18 unique issue ids, all
  structural tags balanced.
- **My own verified findings beyond the agents'**: meta theme-color split
  brain (#0f766e teal-700 exists nowhere in the palette — both layout.templ
  AND manifest carry it); `crmDetail` missing the `wp-service-off` arm
  (settings.templ:355, confirmed in screenshot 11); English clocks mixing
  12h (Kitchen) and 24h ("Jan 2, 15:04") across panels (helpers.go:168-197);
  empty-state voice depth variance (Voicemail exemplary vs History terse);
  type-scale fragmentation (14 sizes below 1rem); radius/spacing token
  bypasses; uppercase-micro-label overuse (7+ sites).

## b) PARTIALLY DONE

- **The review report itself (~90%)**: the appendix matrix (the ~25 findings
  that did not earn their own cards) is NOT yet written, and the scorecard's
  "46 findings total" is a placeholder that must be reconciled with the
  real matrix row count. Tranche-2 list in the fix queue does not yet name
  every matrix item.
- **Merge fidelity management**: the views/shell agents' full reports arrived
  earlier; because I caught myself reconstructing their findings from
  memory, I requested condensed one-line re-emits (delivered ~15 min ago,
  still pending arrival as wake-ups). The matrix waits for them rather
  than risking paraphrase drift. The island agent's 20 findings are verbatim
  and already fully merged.
- **Staged verification**: per-edit HTML structure checks ran and passed,
  but no full read-through of the assembled report in a browser has
  happened yet.

## c) NOT STARTED

- Appendix matrix + scorecard count reconciliation + fix-queue completion.
- Full rendered verification of the report (open-in-browser pass).
- `python3 scripts/status-index.py` regen for this file.
- **Any code fix**: this train is review-only so far — zero product files
  touched (confirmed: tree clean except the report + this status file;
  `ui-shots/` is gitignored, the report HTML is tracked and daemon-committed).
- The hypothesis-flagged findings' runtime verification pass (palette focus
  containment, double-Enter double-send, live-context race, drafts-poll
  loss) — all correctly marked hypothesis in the cards, none reproduced.

## d) TOTALLY FUCKED UP

Nothing destroyed, nothing corrupted, no data loss — but four honest
stumbles:

1. **Hallucinated report arrival, then caught it**: early on I believed the
   views/shell reports had arrived when they hadn't; later I nearly wrote
   the appendix matrix from reconstructed memory. The condensed-reemit
   request was the correct recovery, but the correct move was checking
   arrival before citing. (The stable names I cited — `role="img"` rows,
   drafts poll, segment counter — did come from real received reports, but
   the near-miss pattern is real.)
2. **Lost a strength during an append**: my "add four strengths" edit
   replaced the empty-states strength instead of appending after it;
   caught on the very next verification, restored. Sloppy anchor choice.
3. **Agent rate-limit churn**: the island reviewer failed twice on provider
   rate limits while three agents ran concurrently; I retried into the same
   saturated provider before serializing. ~10 minutes wasted; no findings
   lost (the agent kept its session — 51 tool calls total).
4. **Edit-anchor churn**: three edits failed on entity/whitespace mismatches
   (`&mdash;` vs literal em dash, prettier's `</span\n>` style) and one
   dropped a closing `</div>` — all caught by the verify-after-edit habit,
   but each was preventable by grepping exact bytes first.

Also genuinely forgot until late: `ui-shots/` tracking status (gitignored —
verified), the one-unpushed-commit state (daemon push leg lagging one
commit — within the 60-min manual-push bar, no action needed yet), and the
status-index regen obligation for this very file.

## e) WHAT WE SHOULD IMPROVE

- **Verify agent-report arrival before merging from memory.** The condensed
  re-emit pattern is good insurance; invoking it proactively (at dispatch
  time, as part of the task) beats retrofitting it.
- **Pixel-verify any screenshot-derived claim before it becomes a finding.**
  The transcription-failure misread would have shipped as a false card in a
  less disciplined pass. ImageMagick one-liner = cheap truth.
- **Serialize subagents on a rate-limited provider** (or stagger dispatch)
  instead of retrying all three into the same limit.
- **For long HTML artifacts: draft content in one `write`, not incremental
  `multiedit` on a copied template** — the entity/whitespace anchor failures
  all came from editing prettier-formatted bytes incrementally.
- **The scorecard-total discipline**: never let a count land before the
  section it counts exists (the "46" placeholder is exactly that trap).
- Product-side improvements are the report's job — headline: the dark
  `--on-accent` flip (the only WCAG AA failure in the product), the island
  screen-reader pass (focus-on-replace, `#dial-error` live region), the
  palette's missing touch path + combobox wiring, and ring-time caller
  names.

## f) NEXT TASKS (up to 50)

1. Finish the review report: write the appendix matrix, reconcile the
   scorecard totals, complete Tranche-2 naming.
2. Full rendered verification pass of the report (browser, both
   scrollspy + anchor links).
3. Regen `docs/status/INDEX.md` (`python3 scripts/status-index.py`).
4. Implement the dark-theme `--on-accent` ink flip (app.css +
   island/style.css), re-run contrast math, re-run ui-capture.
5. Add the `wp-service-off` arm to `crmDetail` (one line).
6. Point layout.templ + manifest.webmanifest theme-colors at the real
   palette (both files, one fact).
7. Gate "Prefill from AI draft" on the same prop as the "AI review not
   configured" line.
8. Unify English wall-clock on Kitchen format (helpers.go `formatWhen`).
9. Chip-X hit area to 24px (visual stays 18px, hit-expander).
10. Reword History's empty state to the Voicemail three-beat shape.
11. Remove `role="img"` from the four row surfaces (fax/history/admin/
    live-context), restore readable row content.
12. Drafts-poll skip-if-dirty guard (don't swap under an unsaved composer).
13. Palette combobox ARIA (role/expanded/activedescendant/listbox/option).
14. Palette/help coarse-pointer entry button in the header actions.
15. SMS segment counter: count UTF-16 units for the UCS-2 path.
16. Island focus-on-replace helper + call it from the five transition sites.
17. Resolve incoming-call numbers to contact names on the ring surface.
18. `role="status"` on `#dial-error`.
19. Drop the live region from `#reg-status` (offline banner owns announcements).
20. Route island-built aria-labels through defineMessages.
21. Fix the malformed incoming-call announce (stray argument).
22. Type-scale consolidation to a 3-step secondary scale (tokens first).
23. Radius/spacing token adoption (fold 8px outliers; 2–3 spacing tokens).
24. Uppercase-voice zoning + humanize the reconnecting pill copy.
25. Icon-action pattern: labels or confirm for the thread-head trio and
    History's copy/download/delete row.
26. Admin tab: one `q` filter + `hx-push-url` (reuse the History pattern).
27. Make the brand the page's `h1` (styled as today).
28. `tabindex="-1"` on the panel-head `h2` for post-swap focus.
29. List semantics for the thread rail and row collections (ul/li).
30. Compose "to" inputmode decision (document the matrix-address exception
    or add `inputmode="text"` with hint).
31. `maxlength` hints aligned with server caps (composer, contacts, snippets).
32. `htmx-indicator` on submits beyond fax (reuse `faxSendIndicator`).
33. Drafts badge: count-phrase label instead of repeating the heading.
34. Move `aria-describedby` onto the inputs; name the ✱/# keypad keys.
35. `hx-disabled-elt` on the contacts save/import forms.
36. i18n keys for the two hardcoded panel aria-labels (jump-latest,
    Attachments).
37. Swap the seven emoji/dingbat glyphs for icons-package SVGs; strip ✖ out
    of the `history.missed` key.
38. German rewords: "Stumm aus", "hat abbestellt", "(Rechner der Anfrage)",
    restore the DE no-match recovery line, fix the singular in
    contactsMigrated.
39. Vocabulary drift: DE "Voicemail"→"Mailbox" in the toast; EN
    `admin.calls.lines`→"transcript lines".
40. Island login: in-flight guard + connecting pill state.
41. i18n key for the hardcoded-English 429 toast (`noteThrottled`).
42. `beforeunload` guard while calls are live.
43. In-call `document.title` signal (reuse the 1s render tick).
44. `agoLabel` "just now"/"gerade eben" branch (kills "Last call now ago").
45. Separate voicemail network-failure copy; surface the silent `!res.ok`
    paths (history/contacts loads).
46. Logout confirmation when calls are live (second-click pattern).
47. Polite announces for settled mute/hold/resume transitions.
48. `.catch` + `tabRefreshFailed` key on the bare `htmx.ajax` calls.
49. Transfer row: empty-destination error surface + Escape closes.
50. Runtime-verification pass for every hypothesis-flagged finding, then a
    final matrix "verified" sweep.

## g) QUESTIONS (I cannot answer these myself)

1. **Review-only or fix-tranche train?** Should I implement Tranche 1 (the
   `--on-accent` flip + the crmDetail dot, re-shot) in this session after
   the report lands, or stop at the review and let fixes be a separate
   owner-initiated train?
2. **Finding F1's fix direction is a taste call**: keep the vivid dark
   accent and flip dark `--on-accent` to dark ink (my recommendation —
   matches the hardware-key vernacular, one token), or darken the dark
   accent fill to ~#108351 and keep white ink? The report recommends the
   ink flip; the palette identity is yours to overrule.
3. **Matrix fidelity vs speed**: the two condensed agent re-emits are still
   pending arrival. Write the appendix matrix now from the agents' original
   full reports (all received; minor paraphrase risk on ~10 rows), or hold
   the train until the condensed copies wake me?
