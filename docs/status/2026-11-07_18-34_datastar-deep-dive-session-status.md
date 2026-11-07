# Datastar client-side adoption deep-dive — session status

> Snapshot: claims verified as of 18:34, 2026-11-07. Covers the session that
> answered the owner question "why don't we leverage Datastar fully on the
> client side???" with a library-deep-dive audit.
> **Verified with commands, at this timestamp:** the report train is PUSHED
> (`git merge-base --is-ancestor` for `06ca5b1f` + `fcbc038c` against
> `origin/main` = true; `gh run list --event push` shows the carrying runs
> 42362581057 + 42362720970 green — the 07:46–09:07 UTC failures were the
> concurrent UX train's red window, closed at 09:59 UTC), the verify-block
> claims still hold (`rg -o 'hx-[a-z-]+' … | wc -l` = 164), and the report
> render is a real page, not a blank (`magick identify`: 900x1688, mean
> 0.879, stdev 0.218, 21,628 colors). **NOT verified:** the dashboard library
> (`go-health-dashboard`) checkout location and the F4 nonce-lane change
> scope — deliberately not researched past the finding (owner gate, §g);
> full `nix flake check`/buildflow not run for this train (docs-only change;
> docs-gates + audit-verify + the CSP pin tests were the proportionate gates,
> all green — see §a).

The deliverable:
[docs/research/2026-11-07_datastar-deep-dive.html](../research/2026-11-07_datastar-deep-dive.html)
(verdict NO-GO, posture 78/100) + its wiring. The verdict in one line: 36% of
the 9,844 client-JS LOC is WebRTC/SIP-imperative code that Datastar's own
vocabulary (`data-ignore`, web components) exists to fence OFF; the tabs are
already declarative on htmx (164 pinned `hx-*` sites); expressions are
eval-shaped against the strict shell CSP. One actionable finding fell out
(F4): `/health`'s `'unsafe-eval'` can retire once go-health-dashboard adopts
the embedded engine's v1.0.3 nonce + Trusted Types lane.

## a) FULLY DONE

- **The deep-dive report, complete**: hero/scorecard/summary/current-usage,
  7 findings (F1–F7), the full-adoption cost bill, ranked opportunities,
  capability inventory (12 rows → score 78/100 with published derivation:
  N=12, N/A=3, applicable=9, Σ=7), JS-surface matrix, replayable verify
  appendix. Editorial template copied + edited in place (no line-splicing).
- **Skill done-bar, all six**: docs gates for the change PASS
  (freshness/audit-all/annotate/routing/table-damage); index row added to
  `docs/research/README.md` (with render thumbnail); AGENTS.md carries the
  durable verdict line (196/377 lines — extended an existing line, added
  none); chromium headless render preserved in-repo
  (`2026-11-07_datastar-deep-dive-render.jpg`, 900px) and linked from the
  index; every verify-block command replayed literally — the replay CAUGHT
  two stale claims (159→164 `hx-*` sites, 3→1 `Function(` occurrences) which
  were fixed in the report before shipping; screenshot preserved in-repo.
- **Ruling lineage closed**: `docs/review/libs/go-datastar.md` (the
  2026-11-04 adjacent NO-GO) gained the supersede pointer — its
  signal-patch refresh trigger was examined and does NOT fire; sharpened to
  "a product surface wanting server-pushed reactive state (live
  Reports-shaped)". The TODO_LIST reopen-trigger registry gained the row.
- **Evidence base, all verified at source**: engine bytes read directly from
  the module cache (`// Datastar v1.0.3`, 33,538 B, `data-nonce` +
  TrustedTypes lane present, exactly 1 `Function(`), the go-datastar SDK
  README + wire-format docs, upstream data-star.dev (agentic fetch:
  `data-ignore` = third-party escape hatch, JSON-up/HTML-down, file binds →
  base64 signals, Pro attributes paid, 11.90 KiB claim, morph default).
- **F4 filed with acceptance criteria**: TODO_LIST § 2026-11-07 Datastar
  deep-dive residue — upstream csp.go change + `app_test.go:135` re-pin.
- **Commit `06ca5b1f`** (narrative message) + the daemon-committed HTML/render
  (`fcbc038c`): both pushed, CI green (see scope line).

## b) PARTIALLY DONE

- **Session-close verification (this turn's opening todos)**: push/CI end
  state — DONE now (see scope line; it was NOT done before the previous
  turn's "Done", a doctrine miss, §d4); report drift re-check — DONE (164
  holds at 18:34); render pixel-sanity — DONE (§a). Remaining open verify
  item: none from this session.
- **The deep-dive's verify block is HEAD-relative**, not sha-pinned: the
  `expect 164` line will drift the next time the UX train lands view edits.
  Deliberate (the appendix carries a snapshot note), but it is a known
  soft pin — improvement queued (§e1, §f11).

## c) NOT STARTED

- **F4 implementation** — the go-health-dashboard nonce-lane change. Nothing
  attempted: the library checkout was not located, csp.go was only read in
  the module cache, no branch/tag/release work. Blocked on the owner gate
  (§g1) by policy: library releases + the webphone bump tail (go get → tidy
  → vendorHash re-pin → push → CI-watch) are owner-driven, per the
  dependency-bump doctrine.
- **ROADMAP entry for the live-Reports trigger** — the sharpened refresh
  trigger is recorded in TODO_LIST's registry + the report, but not in
  ROADMAP (needs the owner's product answer first, §g2).
- **CI wiring of docs-gates** (pre-existing TODO item 10) — untouched by
  this session.

## d) TOTALLY FUCKED UP!

- **d1 — Planned with unarrived subagent results.** The agentic fetch had
  moved to background, yet my planning assumed its findings ("Rocket",
  "Pro paid", 11.9 KiB) as if delivered. The save: a local bundle grep
  showed `data-ignore`/`data-signals` are NOT literal in the minified
  engine, which forced re-verification at source before anything was
  written; the report then shipped only locally-verified or
  report-confirmed claims. The verify-external-claims discipline held at
  write time — the failure was upstream of it, in the thinking.
- **d2 — `grep -c` occurrence-vs-lines error.** The first draft claimed
  "3× Function(" — `grep -c -o` counts LINES on a minified single-line
  bundle; the truth is 1 occurrence. Caught by my own verify-block replay
  and fixed pre-ship. Embarrassing class: measurement tooling, not facts.
- **d3 — Doctrine miss at session close.** Declared the previous turn done
  without the ls-remote/CI check ("check CI after EVERY push"). It happened
  to be green — verified 7.5 h late, this turn. The miss is real even
  though the outcome was fine.
- **d4 — Could not visually verify the render** (this model accepts no
  image input). Substituted structural checks (parser-balance, no demo
  leftovers, byte size), and closed the gap retroactively with pixel
  statistics this turn (§a). Honest residual: nobody has LOOKED at the
  page; if the layout is broken in a way statistics can't see, only a
  vision pass would catch it (§g3).
- **d5 — Two edit-tool failures mid-compose** (read-before-edit on the
  fresh copy; entity-vs-literal apostrophe in the summary block). No
  damage, both retried correctly — recorded because they were avoidable
  friction with the template's literal-apostrophe convention.

## e) WHAT WE SHOULD IMPROVE!

1. **Sha-pin verify blocks in point-in-time reports**: `git show <sha>:path`
   inside the replay commands so a moving tree cannot rot an `expect N`
   (this report's 164 will drift; the fix is one command rewrite).
2. **Occurrence counts always via `grep -o | wc -l`** — never `grep -c` on
   minified artifacts (d2).
3. **Subagent results are evidence only once the report arrives** — cite the
   report, not the expectation (d1).
4. **End every train with the end-state check** (ls-remote + `gh run list`),
   not with the commit — the repo doctrine says it, this session skipped it
   once (d3).
5. **Pixel-statistics as the standing render gate** (`magick identify` +
   mean/stdev/colors): three numbers would have caught a blank or
   single-color render instantly; costs nothing; vision stays optional on
   top (d4).
6. **A concurrent train moves every count** — future audits should snapshot
   `git rev-parse HEAD` INTO the report prose so the reader can pin every
   number to a tree.

## f) Things to get done next (ranked, session-scoped)

1. Owner gate: green-light (or park) the F4 dashboard nonce-lane change (§g1).
2. Locate the go-health-dashboard checkout; read csp.go + page rendering at
   HEAD; size the change (emit `data-nonce` on `<html>`, grant
   `'nonce-…'`, drop `'unsafe-eval'`).
3. Implement the dashboard lane in the library repo; run ITS gates.
4. Release the dashboard version (tag; owner call per §g1).
5. webphone consumption: `go get` → `go mod tidy` → vendorHash re-pin recipe
   (`nix build .#webphone.goModules --rebuild` → apply `got:`).
6. Flip the pin: `app_test.go:135` asserts nonce, not unsafe-eval.
7. After F4 lands: update AGENTS.md's /health CSP sentence + the
   go-datastar.md verify block; CHANGELOG line (code-path change).
8. After F4 lands: ANNOTATE the deep-dive report (docs-health ANNOTATE mode,
   never rewrite) noting the finding closed.
9. Run `python3 scripts/status-index.py` — this report must land in the
   generated INDEX.md (done as part of this report's landing if the daemon
   has not yet swept it — verify).
10. docs/status cadence debt: this report is the 76th unarchived (threshold
    70) — the archive ritual is TODO_LIST item 11's bulk verdict pass; my
    report adds one row to that backlog, consciously.
11. Re-replay the deep-dive verify block at the next train boundary (the
    `expect 164` soft pin; see §e1 for the durable fix).
12. Re-check go-datastar.md's `expect 5` verify line after any
    dashboard-driven Go-file churn.
13. Optional: a `scripts/report-render-check.sh` one-liner wrapping the
    pixel-statistics gate (§e5) for future deep-dives.
14. If the owner green-lights live Reports (§g2): ROADMAP entry, then the
    canonical Datastar signal-patch pilot on that tab (the sharpened
    refresh trigger).

## g) Questions I can NOT figure out myself

1. **F4 green light**: may I implement + tag the go-health-dashboard
   nonce-lane change and run the webphone bump tail, or do you want the
   library release to stay owner-hands-on? (Doctrine: dependency bumps and
   library releases do not ride the daemon; the 2026-11-05 tagging
   precedent existed but was unblock-driven.)
2. **Is live Reports a product you actually want?** The Datastar ruling's
   sharpened refresh trigger points at a live-updating Reports tab; whether
   that trigger is live or a dead letter is a product call, not derivable
   from the code.
3. **Render verification ceiling**: should future report renders get a
   vision-model pass you run (and if so, how do you want it invoked), or is
   the structural + pixel-statistics gate the accepted standard?

— Generated by the Datastar deep-dive session; WAITING FOR INSTRUCTIONS.
