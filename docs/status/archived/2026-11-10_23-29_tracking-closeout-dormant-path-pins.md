# Tracking-train closeout follow-up — the dormant-path pins, the harvest, and the honest ledger

> ARCHIVED-REPORT POSTSCRIPT (same evening): landing this report tipped
> the corpus to 66 live reports and red the cadence gate; the ritual
> drained the OLDEST report (2026-10-23_15-25, 93 bare rows) — every row
> now carries a verified verdict (check-rows --strict + bare-rows green
> before the mv; the annotate ratchet back at its 203 pin). The tooling
> lesson it cost: the annotate auditor recognizes a FIXED verdict-word
> set (done/resolved/superseded/fixed/adopted/decided/verified/won't
> implement/not-do/duplicate/struck/moot/no longer, or the `→ routed`
> arrow) — my `routed (...)` / `wont-do (...)` prefixes were invisible
> to it and cost a battery iteration to decode.

> Claims verified as of 23:29. This session executed every remaining
> owner-NEUTRAL item from the tracking train's self-review
> (docs/status/2026-11-10_19-12_tracking-train-self-review.md §f), then
> struck the satisfied rows in that same review. Base tip at session
> start `93d56fc7`; my landing shas `c1da54d3` (harness + lesson, daemon
> commit), `368e0e25` (docs batch, daemon commit), `9a9196a4` (strikes,
> pathspec commit) — CI success on both watchable shas (`368e0e25` run
> 42889360532, `9a9196a4` run 42892063213). Remote tip has since moved
> past me (other sessions' archive sweeps + strike canonicalization,
> `c5074bf8`); none of their diffs touched my files. The working tree
> carries a concurrent session's large in-flight Go refactor
> (catalog.go, internal/server/*, a staged `importutil.go` deletion,
> nix/apps.nix, .buildflow.yml) — observed, never touched, and excluded
> from my commit by a pathspec.

## a) FULLY DONE (verified, not claimed)

- **T28's second half — the EMPTY-run discard guard is pinned.**
  `--self-test` grew 29 → 35 checks: an empty crashed run dir is
  removed by `discard_empty_run`, a run carrying shots survives it.
  Green in BOTH environments (bare host python3 and the capture env
  `/tmp/pysel-path` → python3.12-env with pillow; the pixel block
  asserted there too).
- **§f-3's pin half — the `--open` decision is extracted and pinned.**
  New pure helper `open_target(produced, out_dir, has_display)` (the
  one decision `--open` makes: raise the compare artifact when a
  previous run exists, else the gallery, out-dir fallback, headless
  refusal), call site behavior-identical. Three checks pin all three
  branches. The webbrowser side effect stays stdlib-trusted.
- **§f-11 — the fake-CRM dead-server lesson graduated into
  docs/lessons.md** (§ tracking-train process lessons, fifth entry:
  "A helper server configured by a dying process dies with it"). It
  had lived only in the plan ledger.
- **T13's selector positively verified against the served markup**
  (evidence, not a live call): phone.templ renders the banner with the
  `hidden` attribute and the island shows/hides it by toggling exactly
  that attribute, so `#incoming-call:not([hidden])` matches the
  ringing state and nothing else. Recorded as a plan-ledger row.
- **§f-10 — the harvest is complete.** TODO_LIST's tracking-residue
  section now carries: the T05-blocked chain row (command staged,
  unblock order spelled out), the ring-guard positive-path row, and
  the landed-pins `[x]` row. Plus the earlier night-run and `--open`
  rows, the section is the complete dispatch queue for this train.
- **CHANGELOG `[Unreleased]` exists again** — the 2.13.0 cut had
  consumed the block; the harness refactor + pins carry a Changed
  entry (polish bar honored).
- **The plan ledger gained 2 evidence rows** (21 total): the pin
  landing + the selector verification, timestamped 19:5x.
- **Three satisfied §f rows struck** in the 19:12 self-review (items
  10 harvest, 11 lessons, 15 T28-second-half) with evidence, in the
  14:42 report's established strike format; strike-safety reported the
  file safe and the docs-gates battery passed after.
- **Gates, all green**: scoped `buildflow -s ruff-check / -s
  mypy-check / -s todo-check` exit 0; `scripts/docs-gates.sh` full
  battery PASS (freshness, audit-all, annotate, routing, cadence,
  table-damage, strike-safety); `ui-capture.py --self-test` 35/35 in
  both envs.
- **Push + CI watched with heartbeat monitors** (never silent sleeps):
  the daemon's push leg was alive both times (push observed at hb 3
  and hb 12 of the monitors); CI success on `368e0e25` (9m16s) and
  `9a9196a4` (9m20s).

## b) PARTIALLY DONE

- **§f-3 / item 7 — the dormant paths' LIVE halves.** The decision
  logic is now pinned, but a real-display `--open` pass and the ring
  guard's positive path (a genuine inbound call mid-run triggering the
  re-shoot) still need T05's authenticated prod session plus a caller.
  The `webbrowser.open` side effect is deliberately unpinned.
- **T13** — selector verified against markup, guard negative-tested on
  an idle boot; the positive live exercise remains (same prod-session
  dependency as above).
- **T09/T10** — the instrument flags and the clip path landed in the
  morning session; the calibration VALUES (threshold, badge bar) still
  need real prod noise data. One honest local number stands: identical
  -seed runs 3 min apart peak at 2.5% changed, so the 0.5% badge bar
  is optimistic (expect 2–3%).
- **T15 admin surface** — carried from the morning session unchanged:
  detection is landed and self-tested, dormant-but-correct behind the
  product wall (`isAdminSession` needs `LoginEmail`; the
  credential lane never sets it).

## c) NOT STARTED (all owner-gated, all staged)

- **T05** — the authenticated prod capture. The one command:
  `nix run .#ui-track -- --base-url https://pbx.artmann.tech --track`.
  Blocked ONLY on `WEBPHONE_CAPTURE_EXTENSION` /
  `WEBPHONE_CAPTURE_PASSWORD` (asked twice, unanswered).
- **T05 act 2 → T06 → T07 → T08** — the gallery walkthrough, the
  ranked ugliness inventory, the first proven fix, the first real prod
  compare chain, the mobile pass. All downstream of T05.
- **T25** — the night-run stability pair (laptop on overnight, runs
  ≥8h apart).
- **OG-2** — lighthouse scores (my recommendation stands: NO-GO; the
  ruling is the owner's).
- **OG-3** — where tracking shots may live (laptop-only vs synced) +
  the default prune N.
- **§g answers** — the three questions below.
- **T16 on real prod** — the CRM-strip surface joins when prod's real
  Kith publishes context; verify inside T05's run.

## d) TOTALLY FUCKED UP (honest list)

- **The PIPESTATUS trap, AGAIN, on the same day it was documented.**
  My first gate check was `buildflow -s ruff-check | tail -5; echo
  EXIT=$?` — tail's exit code, not buildflow's. AGENTS warns about
  exactly this and the morning self-review §d had named it hours
  earlier; I reached for the pattern anyway. Caught it within a minute
  (the 4ms cached-step durations looked suspicious) and re-ran with
  redirect-to-file + direct `$?`, which is the only valid shape here.
  No false verdict shipped — but the repeat is the finding.
- **The daemon ate two of my three commit narratives.** `c1da54d3` and
  `368e0e25` landed as `chore: auto-commit N changed file(s)` because
  I kept editing across gate runs instead of committing my own batches
  at boundaries in one add+commit call. Content was verified intact
  afterwards (grep the landed symbols), so nothing was lost — but the
  2026-10-17 lesson exists precisely to prevent this, and I had the
  window to commit each batch myself. Only the third batch (the
  strikes) got a real narrative message, and only because the tree's
  foreign staged deletion forced the pathspec form.
- **The missing-files surprise.** After the docs battery, `git status`
  showed only 3 of my 5 edited files — the daemon had already
  committed `ui-capture.py` + lessons.md mid-flight. I verified
  integrity immediately (the landed content grep), but I had been
  editing for ~10 minutes without tracking that the daemon was
  consuming my work as I went. Awareness lag, not data loss.
- **A near-miss handled, worth naming: the foreign staged deletion.**
  Another session had `internal/app/importutil.go` STAGED for deletion
  in the shared index while I committed. A bare `git commit` would
  have destroyed their in-flight work. The pathspec commit
  (`git commit -m … -- <only-my-file>`) excluded it; the pre-commit
  gate correctly reported nothing formatter-owned to gate. This is the
  pattern that should be DEFAULT on this tree, not a recovery.

## e) WHAT WE SHOULD IMPROVE

1. **Never pipe a verdict.** Gate runs go `command > /tmp/log 2>&1;
   EXIT=$?` — the pipeline form reports the last command's exit and
   has now cost this repo two near-misses in one day. Consider a
   shell alias/habit: if `| tail` appears, `$?` is FORBIDDEN.
2. **Commit my own batches at boundaries** with narrative messages in
   ONE add+commit call, and only then run the next gate. The daemon
   narrating my work is narrative loss even when content survives.
3. **Re-check `git status` after every battery/gate run** on this
   tree — the daemon commits continuously, and "which of my files are
   still uncommitted" is a question that goes stale in minutes.
4. **Pathspec commits as the default** whenever the tree carries
   foreign staged content (tonight that was a deletion — the worst
   case). The race-proof pattern worked perfectly under pressure.
5. **Batch ALL doc edits before the battery.** I ran docs-gates twice
   because strikes came after the first pass. One edit batch → one
   battery → one commit.
6. **Pin self-test counts in ONE home.** The ledger said "29 checks",
   this session made it 35; the next sweep should re-count rather than
   trust any prose number (the count moved twice today).
7. **The heartbeat monitor pattern worked** (bounded loops, printed
   heartbeats, early exit on push/verdict) — use it as the standard
   shape for every push/CI wait instead of ad-hoc sleeps.

## f) NEXT (ranked, roughly Pareto-ordered)

1. **[OG-1] T05** — the authenticated prod capture; only creds missing.
2. **T05 act 2** — owner gallery walkthrough → ranked top-5 ugliness
   inventory into docs/planning/.
3. **T06** — the first UI fix, proven by a before/after tracked run.
4. **T07** — the second authenticated run → the first REAL prod
   compare chain; judge badges, byte-deltas and clipped marks in anger.
5. **T08** — the mobile prod pass (`--viewport 390x844`) + the
   default-or-on-demand ruling.
6. **T09/T10 values** — calibrate threshold + badge bar on prod data.
7. **Real-display `--open` pass** inside T05's session.
8. **Ring-guard positive path** — a real inbound call mid-run
   triggering the re-shoot (T05's session + a caller).
9. **T16 on real prod** — CRM-strip joins when prod's Kith publishes
   context (verify in T05's run).
10. **T25** — the night-run stability pair (quiet window, ≥8h apart).
11. **Noise-floor table** — observed per-surface noise into the
    harness docstring after T07, making the badge bar measured.
12. **[OG-3]** — shot-storage home ruling; `--prune` awaits it.
13. **Default prune N** — set once OG-3 rules (current: no pruning
    without the flag).
14. **[OG-2]** — lighthouse ruling (recommendation: NO-GO, honest lane
    is a separate opt-in app).
15. **T15's honest closure** — identity-carrying capture lane or the
    admin surface dies (product wall, owner call).
16. **Upstream buildflow nix-hash-fix issue** — file with today's
    commit-order evidence IF the owner wants it filed (§g-1).
17. **`--emulate-feature`** — the one-line opt-in a11y axis if T06's
    fixes surface motion/contrast uglies.
18. **T20 polish** — confirm the owner accepts prod runs landing in
    the repo's `ui-shots/` by default (tied to OG-3).
19. **CHANGELOG fold** — the `[Unreleased]` block rides the next
    version cut.
20. **Keep the 429-stays-manual ruling** in the harness docstring when
    the AGENTS command block is next edited.
21. **Sweep the plan ledger first** at the next session start — it is
    this train's status home (21 evidence rows), not the reports.
22. **Re-count the self-test checks** at the next sweep (29 → 35
    today; trust re-counts, not prose).
23. **Re-verify `nix run .#ui-track -- --help` after the concurrent
    session's refactor lands** — their in-flight diff touches
    nix/apps.nix, which hosts the ui-track app.
24. **Re-run `--self-test` after any foreign landing that touches
    scripts/** — scripts/ was swept twice tonight by other sessions
    (shellcheck, ruff); neither hit ui-capture.py, but the file is in
    their blast radius.
25. **`--crawl` mode** — true link discovery beyond the curated matrix,
    only if T05's shots prove the matrix too narrow (14:42 §f-9).
26. **`--min-height` blank-shot probe** — a blank/1px shot should warn
    (14:42 §f-20; the styled-404 skip covers one class, dead renders
    another).
27. **PWA manifest/icon surfaces ruling** — probably NOT (no visual
    value); record the verdict when touched (14:42 §f-16).
28. **Per-surface viewport clipping** — check whether `--diff-clip`
    (landed) supersedes the 14:42 §f-22 idea; strike the row if so.
29. **Re-seed thread quirk** — the Test C re-seeding observation goes
    to lessons.md only if it bites again (14:42 §f-26, dormant).
30. **Reopen-trigger registry sweep** for this train — none expected
    (no library adoptions; pillow is capture-env-only); verify and
    close the 14:42 §f-28 row.
31. **Daemon-hot multi-edit hardening** — for large batches on shared
    files, consider the atomic python-rewrite pattern from the start
    (19:12 §e-2), not just for appends.
32. **`--track` compare chain verification on LOCAL data** — the two
    tracked local runs in `ui-shots/127.0.0.1_18097/` could birth a
    second compare generation to exercise multi-hop diffing before
    prod data exists (cheap, no owner input).
33. **docs-gates ratchet re-measure at the next sweep** — tonight's
    battery passed at annotate 203 / audit 72; the strikes moved the
    corpus again, re-derive rather than copy (the standing rule).
34. **The foreign refactor's CI verdict** — when the concurrent Go
    refactor lands, attribute its run before acting on any red (the
    mixed-tree ritual; it owns the tree right now, not me).
35. **Admin-lane question timing** — only worth asking after OG-3's
    storage ruling lands; the admin surface is dormant-by-design until
    then (deferred §g-2).

## g) QUESTIONS (cannot answer myself)

1. **May I have the T05 credentials** (`WEBPHONE_CAPTURE_EXTENSION` /
   `WEBPHONE_CAPTURE_PASSWORD` for pbx.artmann.tech) — asked twice
   before; every item in §f 1–11 is downstream of this one input.
2. **May I file the buildflow nix-hash-fix issue upstream**
   (LarsArtmann/buildflow) with today's evidence — commit-order proof
   (go.mod 11-08 OLDER than the vendorHash pin 11-10) plus the step's
   own "build green but N lock-drift warnings" diagnose lines — or do
   you want to carry it yourself, since it is your tool?
3. **OG-3, the storage ruling**: where may tracking shots live
   (laptop-only vs a synced store, given real caller data on disk),
   and what should `--prune` default to once ruled (my default stays
   NO pruning without the flag — silent deletion of evidence is worse
   than a full disk)?
