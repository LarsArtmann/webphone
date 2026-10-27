# SUPERB plan execution — T1/T2/T5/T6 shipped, T8 blocked by session interference

**Claims verified as of 19:51 CET.** Push/CI statements rot fastest; everything below was
re-checked against HEAD (`055030c`) at that timestamp. Execution context: the owner GO for the
full 27-task plan landed at ~18:00; FOUR Crush sessions are active on this repo today (this
executor, the error-surface train, the TypeScript type-gate train, and the auto-commit daemon)
— the interference shape is documented in §d because it is the session's dominant risk.

## a) Done

- **State check (prerequisite).** The morning's red main was the concurrent session's
  TDD-first tests plus a transient `sse.go` missing-import break — both healed by their train
  before I touched anything. Verified green base, then started the plan.
- **T1 (issue #6, stop-the-lying slice) — SHIPPED, pushed, CI green.** `domain.Disposition`
  vocabulary (answered/missed/voicemail/ivr/failed/unknown) + the end-fact machine table
  (`DispositionFromEnd`: never-connected ⇒ missed, every media-bearing row ⇒ unknown — no
  automated path may claim answered) + the legacy-outcome demotion table
  (`LegacyOutcomeDisposition`: the retired `answered` wire value DEMOTES to unknown) +
  `AnsweredClass()` class law (exactly one member is the answered class — the #6 acceptance
  bar). `POST /api/calls` takes `disposition` (validated, 400 on foreign values) and demotes
  the legacy `outcome` field for stale tabs; the CRM client/resolver carry the typed
  disposition verbatim (Kith renders as-is); the island's `recordCrmCall` sends
  `unknown`/`missed`. Island wire-contract test, server e2e (4 new subtests incl. the demotion
  pin and the never-journal-a-refused-report pin), crm client tests. New code
  `domain.disposition` registered same-commit. The stack browser E2E is owed at the end of the
  #6 UI trains per the plan (T1+T2 are wire/JS; the DOM contract itself is untouched).
- **T2 (issue #6, operator confirm) — SHIPPED, pushed.** Data-model ruling: the verdict is a
  SEPARATE fact sharing the note's (owner, call) slot; `CallNoteSaved` now snapshots the whole
  slot (body + disposition) with the merge done SEAM-SIDE (the engine's update-fold passes a
  nil prev on missing rows — fold-merge panicked in tests; the house doctrine "merges stay in
  the store seam, folds stay plain inserts" was right and is now load-bearing knowledge). The
  note cap guards verdict slots too. `POST /api/calls/{id}/disposition`: persists the verdict
  FIRST (a fact about the human's knowledge, independent of CRM reachability), then journals
  the awaited correction upstream; CRM 502 keeps the verdict stored and the retry open; same
  verdict replay is an inert 204 (idem per call+verdict), a different verdict is a correction
  that journals again. Island: the ended card grows three chips (answered/voicemail/IVR,
  en+de, CRM-gated — zero config leaves the card byte-identical); success collapses the row to
  the saved label; a 502 keeps chips clickable. Machine-API contract registered with its full
  status vocabulary. Tests: store merge pins (note-save carries the verdict forward and vice
  versa, owner scoping, cap), 4 endpoint groups (correction, verdict-survives-outage,
  disabled-CRM shapes, guards), 3 island chip tests. 260/260 island tests green.
- **T5 (issue #9, typed crm module lane) — SHIPPED, pushed.** The app gained
  `crm.token_file` (the credential form of `crm.token`: at-most-one-of, read+trimmed at boot,
  missing/empty file fails the boot loudly — `config.crm.token_sources` +
  `config.crm.token_file` registered same-commit). The NixOS module gained the typed
  `services.webphone.crm` lane mirroring the shipped oauth2 lane: `enable`/`url`/`tokenFile`,
  token via `LoadCredential` (rotation = unit restart; missing file fails the UNIT START),
  eval-time fail-closed assertions (half shapes, knob-without-enable, store-leaked token), the
  oauth2-style typed/raw conflict (one home per fact), and the typed-off lane leaves raw
  `settings.crm` byte-untouched. Deliberate deviation from the plan's M30 mechanics: settings
  plain-defs instead of env vars (the oauth2 precedent; conflict-visible) — flagged in §g.
  `nix/module-check-crm.nix` (typed render / fail-closed / raw lane / conflict) wired into
  `webphone-module`; the whole check is green (including all pre-existing oauth2 cases after
  the `LoadCredential` mkMerge refactor).
- **T6 (issue #9, /healthz crm leg) — SHIPPED, pushed.** `crm.Client.HealthCheck`: one
  authenticated lookup round trip with a never-matching reserved-range number — 200-with-zero
  results proves reachable AND authorized; 401/403 yields the operator-fix rejection
  ("rotate the CRM token credential"); 5xx transient; deliberately NOT routed through the
  Resolver (the probe never pollutes the lookup cache/single-flight — pinned: 3 probes hit
  upstream 3 times). Wired three ways: the /healthz readiness list (NAMED check "crm", so the
  503 body names the seam), the NewChecks fallback arm (non-critical, matching the ASR-leg
  posture), and the prod probe via NAMED "crm" injector registration when configured (the asr
  precedent) + the `do.HealthcheckerWithContext` conformance assertion. Unconfigured
  deployments carry no crm check anywhere (zero config = byte-identical probes). Tests: client
  arms + cache-bypass, and the healthz integration triple (green names crm; revoked token →
  503 naming crm with the fix text on the very next request; unconfigured → no crm check).
- **Process verifications along the way:** full `./internal/...` sweep green after T1; the
  touched-package suites green after every train; whitespace-drift gate clean on every
  narrative commit; `nix fmt` clean at every gate; every push followed by a CI check (green on
  the batches that carried T1/T2/T5/T6).

## b) Partially done

- **T8 (daemon-audit orphan guard) — implemented TWICE, reverted TWICE, NOT landed.** The
  design is finished and validated piecewise: a reachability check (`merge-base --is-ancestor`
  against `origin/main`, falling back to `main`) turns the #5 false-flag class into an
  `ORPHANED … unreachable from <ref>` verdict instead of a flag (and never files an issue for
  rewritten-away content), plus two selftest asserts (detached-HEAD orphan commit →
  annotation, never FLAGGED). `bash -n` clean. What keeps killing it: the working-tree file
  was reverted to HEAD twice within seconds-to-minutes of my writes (mtime evidence 19:45:09,
  19:48:39), while the concurrent type-gate session was landing wide hand-commits
  (`feat(types)` at 19:46 swept broad). My second application was verified on disk
  (`grep`-confirmed) and gone minutes later. This is session interference, not a design
  failure — the same edits re-applied cleanly each time.
- **Discoverable defect found while debugging T8 (unfixed, fleet-wide):** `core.hooksPath` is
  set to the RELATIVE `.githooks` in the GLOBAL git config. A relative hooksPath resolves
  against the invoking CWD: tmp repos/subdirs either silently get no hooks or pick up a
  foreign repo's `.githooks`. This is at minimum a footgun for every selftest that commits in
  a scratch repo (the selftest's daemon-pattern commits can trip webphone's dep-guard hook
  depending on where git resolves the path), and it explains several of today's ghost-hook
  observations. Fix belongs in nix-config (absolute path or per-repo only).

## c) Not started

T3, T4, T7 (the remaining #6/#9 surfaces), T9–T27 — i.e. 21 of 27 plan tasks, including the
whole second half (T11–T15 #8 review artifacts, T16–T18 #7 cooldown, T19–T24 verification +
pbx-artmann side, T25–T27 docs polish/harvest). T9 (pre-commit drift gate) is next in the
tier-2 queue after T8; it touches only `.githooks/` and is far less interference-prone.

## d) Fucked up

1. **T8 lost twice to interference I should have anticipated.** After the FIRST revert I
   re-applied but still didn't secure the file with an immediate atomic stage+commit — the
   second application died the same way. The lesson is mechanical: in this repo, an edit that
   survives its gates must be committed in the SAME minute; anything longer is a coin flip
   with four other writers.
2. **The plan's commit protocol was largely defeated by the daemon.** T1/T2/T5/T6 all landed
   as scattered `chore: auto-commit` sweeps interleaved with the other sessions' files; my
   narrative commits (19ced86 T1, de091c9 T5) captured only the remainder each time. History
   is verifiable (every file's landing commit is traceable) but not narrative-shaped. The
   daemon's 2–5-minute cadence beat me on every train.
3. **Three sloppy edit-tool mistakes, all caught by gates but each costing a cycle:** a
   destructive `old_string` that deleted `TestResolverCachesLookups`'s stub setup (restored
   immediately); a comment continuation written without its `//` prefix (broke the readmodels
   build); a doubled closing brace in `crm/client.go` from anchoring an insertion on a
   mid-function line.
4. **Invented helper names before checking house idioms** (`readCappedJSON`,
   `domainWrappedNoteListFull`) — both replaced by the existing patterns on the next build.
   Reading the target file's siblings BEFORE writing is the habit; I did it for design but not
   consistently for micro-helpers.
5. **The island chip test burned 4–5 debug cycles** on two stub-semantics facts
   (`textContent` is a plain property, not derived from children; my failure-path mock never
   pushed to the recorder) that `helpers.mjs` documents. I wrote assertions before reading the
   stub's full contract.
6. **The engine's update-fold nil-prev panic** — I deviated from the documented
   store-seam-merge doctrine because the metaengine docs showed a prev-aware fold; the tests
   caught it, but the doctrine existed precisely for this and I should have weighted it higher
   than a doc example.

## e) Improvements

1. Ship the security valve first: the disposition vocabulary + demotion tables killed the
   active data-corruption defect before any UI existed (Pareto tier-1 worked as planned).
2. The persist-first/await-correct-second confirm contract gives #6 an honest failure mode at
   every point: the human verdict is never lost to a CRM outage, and retries are idempotent
   per (call, verdict) while remaining real corrections for changed verdicts.
3. The slot-snapshot fact (one event type carrying the full review slot) is replay-safe,
   self-describing in the event stream, and lets the cap guard both note and verdict with one
   `claimSlot` path — simpler than the two-fact design I first attempted.
4. `crm.Client.HealthCheck` rides the real authenticated endpoint with a reserved-range number
   and bypasses the resolver: one code path proves reachability AND authorization, and the leg
   can never go stale-green from a cached verdict.
5. The NixOS module's CRM lane is a deliberate clone of the oauth2 lane (credential doctrine,
   conflict guard, fail-closed eval assertions, extracted module-check file) — consistency
   with the shipped pattern made the whole lane reviewable in one pass and the check green
   first try after the mkMerge fix.
6. Every touched seam got its negative pin, not just the happy path: foreign disposition ⇒
   never journals; refused report ⇒ empty CRM stub; unconfigured ⇒ no crm check in the probe
   body; typed-off ⇒ raw lane byte-untouched.

## f) What's next

1. **T8 re-landing (atomic this time):** re-apply the orphan guard, run the selftest, and
   stage+commit within the same tool call — then verify the commit survived 60s before
   pushing. If it reverts again, stop and escalate (§g-1).
2. **T9 pre-commit drift gate** (`.githooks/pre-commit` running `whitespace-drift.sh` on the
   staged diff + selftest case) — small, `.githooks`-only, low interference.
3. **T20 baseline audit + T21 watcher check** (30 min each, read-mostly).
4. **T3 history disposition filter + T7 Settings tri-state** — Go-template + i18n.go work;
   i18n.go is shared with the type-gate session (they annotated i18n.js first) — re-View
   discipline applies.
5. **T4 `call_logged` payload contract note + pbx-artmann bridge comment** — the Kith wire
   carries the disposition verbatim since T1; the cross-repo comment closes the loop.
6. **T10 daemon forensics** — today's commit/push stalls are live evidence.
7. **T11–T15 (#8 review artifacts), T16–T18 (#7 cooldown), T19 external-claims verification.**
8. **T22–T24 pbx-artmann side of #9** — CA trust, token provisioning, stack e2e; the
   tri-repo ritual applies (webphone pushes BEFORE the stack re-pins).
9. **T25–T27 docs polish + harvest** — TODO rows to stable names, FEATURES/CHANGELOG for the
   #6 landing, HARVEST of plan-born items, M18 tail refresh.
10. **Stack browser E2E once after the #6/#7 UI trains** (release-runbook obligation; the
    DOM contract tests already cover T2's island additions at the unit level).
11. **Global `core.hooksPath` fix proposal to nix-config** (§b finding) — owner call, fleet
    scope.
12. **CHANGELOG entries** for T1/T2/T5/T6 (the [Unreleased] bar: these are committed
    code-path changes) — currently owed.
13. **AGENTS.md updates** — the store-seam-merge doctrine now has a proven failure mode worth
    one line (update-fold nil-prev), and the CRM lane joins the module options list.
14. **art-dupl + registry re-convergence check** — the other sessions' landed tests may have
    shifted both gates again; verify before the next full buildflow.
15. Remaining plan tasks in tier order per `docs/planning/2026-10-27_17-25_SUPERB-issue-triage-execution.md`.

## g) Questions

1. **T8 interference policy:** the type-gate session's wide hand-commits keep sweeping (and
   something reverted my file twice). Options: (a) I pause T8 until their train goes quiet and
   proceed with T9/T20/T21 (my recommendation — T9 is .githooks-only), (b) I re-apply T8
   atomically now and accept the revert risk, or (c) you pause the other session(s).
2. **Global relative `core.hooksPath=.githooks`** (set in `~/.gitconfig`): fix fleet-side in nix-config (absolute path), or scope it per-repo and
   remove the global? The relative form breaks every scratch-repo selftest on this host.
3. **T5 mechanics deviation:** the module writes `settings.crm` (oauth2 conflict-guard
   pattern, conflict-visible, off-lane untouched) instead of the plan's M30 env-var lane. Keep
   the deviation (my recommendation — it matches the shipped oauth2 lane exactly) or align to
   the plan text and rework to `WEBPHONE_CRM__*` env wiring?
