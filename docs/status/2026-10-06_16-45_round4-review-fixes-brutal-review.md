# Round-4 review fixes — brutal review

**Session:** 2026-10-06 ~15:30–16:42 CEST · review of THIS session only (the
f.1/f.2/f.3/f.6/f.7 execution + three interim §g rulings + push/CI close).
End state: remote = `776f09b` CI SUCCESS (run 37480035135); local = `cb99714`
(docs-only exit-state line, unpushed, riding the daemon whose push-leg was
dead all session); a fresh daemon reflow pass (16:40ish) re-aligned my three
just-committed md edits — content-inert, detector ZERO, left to ride.

## a) FULLY DONE

| # | Item | Receipt |
| - | ---- | ------- |
| 1 | Three §g questions ruled explicitly, each marked INTERIM + sitting-owned: push-bar anchor = first-unpushed-commit timestamp; CHANGELOG bar = code-path changes log, test/docs silent; accepted-risk readership = forced onto the briefing | AGENTS Conventions (2 bullets, `14aaec1`); CHANGELOG line (`e684b99`); briefing row 35 (`d355b91`) |
| 2 | f.1: detector COMMITTED as `scripts/md-table-shape.py` — the d.1 "instrument only in transcript" fix; D25.1 claim reproduces from the repo (default + exact sweep set = ZERO) | `240ea98` + hardening commits ending `776f09b` |
| 3 | Detector hardened through FOUR dogfood rounds: wrapped-row joins, escaped-pipe handling, no prose gluing, display-width measurement; 8-case self-test incl. every adversarial shape that actually fired | `2402010` (daemon-captured), `776f09b` |
| 4 | f.7 canary FIRED (`2fbdfd7`, 16:18) with the D1.4 hypothesis INVERTED: the daemon is the ALIGNER; compact session appends are the mix source; corpus ZERO under v4 | TODO evidence; plan D1 row; report item 7 |
| 5 | f.6: full-repo slog sweep CLEAN (app/pbx/fax/messaging/views — one timezone Info, zero log calls in pbx+views, ids/errors/CDR-class attrs only) — b.3 closed | TODO AUTH TAIL note (`fbd0c5e`) |
| 6 | f.2 CHANGELOG line + f.3 briefing row 35 + plan re-count (35 rows, SEVEN §g verdicts, 6th push-lag datapoint, anchor refinement); briefing title's stale count dropped | `e684b99`, `d355b91` |
| 7 | Gates: fmt 0 · suite rc=0 (19 ok) · island 186/186 · buildflow EXIT 0 · codespell delta CLEAN (one pre-existing briefing hit, first adjudication) | session report § Gates |
| 8 | Pushed at the bar exactly as ruled (anchor stated: 15:26:54 → 16:26:54; heartbeat-visible wait); remote = local at push; CI verdict FETCHED | run 37480035135 SUCCESS |

## b) PARTIALLY DONE

1. **Corpus ZERO is true but only 1/3 mine.** I normalized exactly one table
   (samber-do Status column); the daemon's 16:18 aligning pass normalized the
   rest. The final claim is verified, but "the corpus is clean" overstates my
   contribution — the honest sentence is "the daemon aligned it, my detector
   now reads it correctly."
2. **Session report 16-10** was written mid-session with a placeholder
   exit-state, then patched twice (timestamp header, exit line). Its filename
   says 16-10 while its content claims ~16:50 — cosmetic drift I chose not to
   chase.
3. **cb99714 unpushed.** Followed the bar's letter (60m from 16:36 = 17:36,
   session ended first) but the practical result is: remote lacks the final
   report state, on a daemon that has not pushed in 4+ hours. The handoff is
   documented but the debt is real.

## c) NOT STARTED (correctly gated, untouched)

D1–D17 owner legs, D2–D5 deploy chain, D20 (2026-11-05), D21 (2026-12-20),
D22–D24, D25.3 (D1.4-gated — though the canary verdict just made that note
nearly writable), D29/D30, f.4/f.5/f.8–f.15. No local `nix flake check`
(CI's sandbox judged the push; same deliberate trade as round 4).

## d) TOTALLY FUCKED UP!

1. **I committed the d.1 fix while committing a new d.1.** v1 of the detector
   went in at `240ea98` with a 4-case self-test that lacked the two adversarial
   shapes present in MY OWN session output — my report's own table wrapped
   mid-cell and contained an escaped pipe. The instrument shipped unverified
   against the exact corpus it had just measured, and its "39 findings"
   evidence line (committed into TODO + the plan) was artifact-inflated. It
   took three more rounds and two count corrections to reach truth. The rule
   "instruments must be committed" was necessary but not sufficient — the
   missing rule is "committed instruments must have run clean against the
   output that motivated them, BEFORE the evidence lands."
2. **Sloppy edit construction — twice.** The report renumbering multiedit
   included two literal no-ops (old == new), then a follow-up edit merged
   f.2's body under f.3's header ("4. **f.3 — briefing row 35** ... logged
   under [Unreleased]"), producing duplicate/garbled numbering I had to view
   and re-fix. Same class as round 4's d.3 (duplicate struct field): editing
   without simulating the resulting file.
3. **Fixture arithmetic done in my head, twice wrong.** The emoji self-test
   case shipped with header display-width 12 vs body 10 — my miscount — and
   my first "fix" was still wrong. Width-sensitive fixtures must be built by
   code that pads to width, not by mental arithmetic.
4. **A lying helper name in v3's first draft** (`is_default_scope` checking
   for archived paths) — caught on review minutes later, but it went through
   an edit round as a lie. Plus one pasted-garbage line (`sum(... if False)`)
   in a table_blocks rewrite that only died because I viewed the file before
   the next edit.
5. **An unexplained observation was left dangling.** `gh run list` showed two
   CI runs (`357ffec`, `2bbbc2e`) on heads that exist nowhere in the repo
   history, completing around my push. I internally said "not my concern" and
   wrote nothing down. Either they are fork/PR-triggered runs (benign) or
   something odd happened — an honest report flags it; mine didn't until now.

## e) WHAT WE SHOULD IMPROVE!

1. **Dogfood the instrument against its motivating output BEFORE committing
   the evidence.** Tonight's four-round saga would have been one round.
2. **Build width-sensitive test fixtures programmatically** (a pad-to-width
   helper in the self-test) — human arithmetic is the bug, remove the human.
3. **Simulate every multiedit** (read the post-state mentally or diff after);
   no-op edits in a batch are a smell that the batch was constructed hastily.
4. **Closing-push policy needs a ruling** — the bar governs mid-session
   stalls; session end currently leaves a documented-but-real unpushed debt
   whenever the daemon's leg is dead (tonight: 4h+).
5. **Write down every unexplained observation** (the phantom CI heads) — the
   error-contract discipline for dashboards applies to my own reports too.
6. **The D1.4 convention note (D25.3) is now nearly free**: two canary
   datapoints (16:18 reflow compact→aligned; 16:40 re-align of my fresh
   compact edit) prove aligned is the daemon's fixed point and compact is
   not stable. Whoever writes the note should say: append rows in the
   aligned shape OR accept one normalizing reflow; never hand-fight the
   daemon with compact normalizations (what D25.1 did to 03-40 was already
   undone within the hour — by the daemon, benignly).

## f) Up to 50 things to get done next

1. Owner: D1 THE SITTING — 35 briefing rows, SEVEN §g verdicts, now also
   ratifying/overturning my three interim rulings (anchor, CHANGELOG bar,
   row-35 acceptance).
2. Owner: D26.3 daemon push-leg diagnosis (pma logs) — dead 4h+ this session;
   single highest-value infra fix left.
3. Push `cb99714` when the bar is met (17:36) or when the daemon recovers;
   fetch the CI verdict (docs-only, low risk).
4. Owner: D2→D3→D4→D5 deploy chain — five+ sessions of green code (now incl.
   the daemon's flake.lock nixpkgs bump 494ce7fd→151fa4e8) still not live.
5. D7 post-sitting paper closes; D8 v2.9.0 release train folding the auth +
   auth-tail + review-fixes trains.
6. f.9: island test pinning authedFetch's X-CSRF-Token sourcing from the meta
   (still only indirectly covered by the 186).
7. f.8: whitespace-drift pre-commit hook live-cycle test against a real daemon
   cycle before wiring.
8. f.15: CI arm for whitespace-drift.sh (owner call; currently habit-only).
9. D25.3: write the coupling-home convention note once D1.4 rules — evidence
   is now committed (two canary datapoints, corpus ZERO).
10. Explain the phantom CI heads (`357ffec`, `2bbbc2e`): fork PRs targeting
    main? If unexplainable, escalate — unexplained CI activity is not nothing.
11. AGENTS formatting bullet: document that the daemon reformats
    `scripts/*.py` + shell (style-only, self-test-verified tonight) — the
    bullet currently scopes formatting to treefmt/oxfmt-owned sets.
12. lessons.md: the display-width lesson (daemon aligns emoji to terminal
    cells; codepoint-length tooling misreads it) — a trap for any future
    width-sensitive tool.
13. Watch the next daemon pass re-align the samber-do table (already observed
    once at 16:40) — third datapoint = convention settled de facto.
14. Optional: pad-to-width fixture helper inside md-table-shape's self-test
    (e.2 above).
15. D12/D10/D11/D13/D14/D17 owner verdicts; D15/D16 stack batches; D20/D21
    date-gated watches; D22–D24 (unchanged gates).

## g) Questions I can NOT figure out myself

1. **Session-close push exception:** should the manual-push bar gain a
   closing clause (a green-gated final push at session end, independent of
   the 60m stall), or is the documented handoff — trailing docs commit left
   for a dead-legged daemon — the intended cost?
2. **Do interim rulings bias the sitting?** My three §g answers are written
   into AGENTS/briefing/CHANGELOG as INTERIM + sitting-owned. Keep that
   pattern for future blocked-on-owner questions, or should such rulings
   live ONLY in the briefing row (never in AGENTS) until ratified, so the
   operating docs never carry unratified defaults?
3. **Daemon formatting scope:** the daemon reformats `scripts/` (python +
   shell, style-only) though no formatter owns those files per AGENTS.
   Accept as the implicit rule (daemon = formatter of last resort over all
   non-Go/non-island files), or should its scope be pinned/documented
   somewhere it can be argued with?

---
*Written 2026-10-06 16:45 CEST, immediately after the execution session; no
post-hoc fixes applied during the review — everything actionable waits in f).*
