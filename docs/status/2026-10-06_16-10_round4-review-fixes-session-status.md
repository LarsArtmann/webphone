# Round-4 review fixes — session status

**Session:** 2026-10-06 ~15:30–16:10 CEST · scope: the brutal review's
assistant-executable f) set (f.1/f.2/f.3/f.6/f.7) + explicit rulings on the three
open §g questions (the user's "keep going" demand authorized deciding them; every
ruling is interim and owned by the D1 sitting). Entry state verified: the 15-25
review file daemon-committed as `64fc3a4`; remote `43b6889` CI SUCCESS.

## The three §g rulings (all INTERIM, all sitting-owned)

| # | Question | Ruling | Receipt |
| --- | -------- | ------ | ------- |
| 1 | Push-bar anchor | FIRST UNPUSHED COMMIT's timestamp (`git log origin/main..HEAD --format=%cI \| tail -1`) — the daemon's own push state is unobservable while its push-leg is dead, so the conservative observable anchor wins; state the anchor whenever invoked | AGENTS Conventions bullet (`14aaec1`) + briefing-row-1.3 refinement in the plan |
| 2 | CHANGELOG bar for refactor-class changes | committed code-path changes get an `[Unreleased]` line — polish/one-home trains included (10-04 precedent re-affirmed); test-only and docs-only deltas stay silent | AGENTS Conventions bullet (`14aaec1`) + the applied line itself (`e684b99`) |
| 3 | Accepted-risk readership | FORCED onto the briefing: row 35 ratifies the passkey `?user_id=` ceremony-key acceptance — it can no longer die inside an unopened D30 | briefing row 35 + plan D1 row re-count (`d355b91`) |

## f) execution

1. **f.1 — `scripts/md-table-shape.py` COMMITTED** (`240ea98`): the d.1
   "instrument exists only in transcript" fix. Self-test 4/4 (mixed flags,
   compact clean, fully-aligned clean = D1.4's question not ours,
   unbalanced-padding flags); default scope `docs/status/` live = ZERO;
   explicit rerun over the exact D25.1 10-05/10-06 status+planning sweep set =
   ZERO — the committed claim now reproduces from the repo.
2. **Full-corpus evidence for D1.4** (dogfood, deliberately NOT normalized):
   25 findings across 9 files — 4 archived snapshots, FEATURES.md, and 4
   live docs including the briefing's rows-29–32/appended-rows mix. Rewriting
   archived snapshots trades hypothetical churn for real churn; the live-doc
   residue joins D1.4's decision space with the detector now a one-command
   check.
3. **Detector hardening (dogfood-before-wire, enforced on myself):** v1
   miscounted the corpus at 39 — 14 artifacts of two latent bug classes,
   both caught by running the committed instrument against THIS report's own
   tables: (a) wrapped table rows fragmented into ragged pseudo-rows
   (false positives on wrap points, false negatives where whole tables
   collapsed below the separator check), and (b) escaped pipes `\|` split
   cells that contain shell snippets. v3 joins open rows (bounded by blank
   lines so prose never glues), honors escaped pipes, and carries a 7-case
   self-test including both adversarial shapes plus wrapped-then-prose.
4. **f.2 — CHANGELOG line** (`e684b99`): the authedFetch `csrf.js` one-home fix
   logged under [Unreleased] ### Changed, matching the 10-04 precedent's shape.
5. **f.3 — briefing row 35** (`d355b91`): the `?user_id=` acceptance question
   with options + recommendation (accept: single-use, burned at finish,
   TLS-covered). The plan's D1 row now counts 35 briefing rows, SEVEN §g
   verdicts, the 6th push-lag datapoint (green/60m-stall → manual push), the
   anchor refinement, and the detector's corpus evidence; the briefing title
   dropped its stale "28 Decisions" count.
6. **f.6 — full-repo slog sweep: CLEAN** (b.3 closed; TODO evidence in
   `fbd0c5e`): app = one `log.Info` ("timezone applied", zone attr); pbx +
   web/views = ZERO log calls; fax = 6 slog.Warn (error + fax_id); messaging =
   3 slog.Warn (error + message id + owner extension + remote number —
   CDR-class data, the accepted 18.2 bar); no print/os.Stderr paths in any of
   the five.
7. **f.7 — canary ARMED** (TODO evidence in `240ea98`): no commit has touched
   the normalized 03-40 table since `455e26c`; the next daemon md commit
   touching it triggers a detector rerun — a re-introduced mix would prove the
   daemon formatter IS the near-aligned source for D1.4.

## Gates

- `nix fmt`: 0 changed · full Go suite: rc=0, 19 ok packages, zero FAIL lines ·
  island JS: 186/186 pass, 0 fail.
- codespell over the session docs delta: one PRE-EXISTING warning-level hit
  (a hyphenated pre- word at briefing line 121, introduced by daemon commit
  `1cb3445` before this session) — first adjudication, noted per the 18.3
  precedent (durable ignore only on the second hit), not session debt.
- Narrative-commit race: 4/5 won (`240ea98`, `e684b99`, `d355b91`,
  `14aaec1`); the f.6 TODO edit was captured by the daemon (`fbd0c5e`) with
  content verified in the diff — the documented lost-race class.

## Not done (correctly gated)

D1–D17 owner legs, D20 (2026-11-05), D21 (2026-12-20), D22–D24, D25.3
(D1.4-gated), D29/D30 (D1-conditional), f.4/f.5/f.8–f.15 (owner or
post-sitting). The f.8 pre-commit-hook idea stays unwired deliberately: the
daemon's commits traverse the same hook and the live-cycle test has not run.

---
*Exit state: see the final push + CI verdict below (filled at session close).*
