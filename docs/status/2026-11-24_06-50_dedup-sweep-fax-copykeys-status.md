# Status — dedup sweep session (pinned-gate verdict, fax `faxCopyKeys` extraction)

> Scope (claims verified as of 06:50, 2026-11-24, HEAD `d8773273` — HEAD moves as the
> daemon lands the fmt-repair): VERIFIED by command this session — CI-pinned art-dupl
> v0.7.2 `check . -t 1 --baseline-path .art-dupl-baseline.json` = **0 new vs 114**
> (run BEFORE and AFTER the code edit); the owner-pasted LOCAL v0.9.0-8ce90a4 report
> parsed and all 42 shown groups attributed; fax + full `internal/web/views` suite
> (i18n parity/unused/missing-key gates included) green; `internal/server` suite green
> (15.1s); `go build ./...` green; `nix fmt` 0 changed; docs-gates battery PASS after
> the dprint repair; post-extraction v0.9.0-8ce90a4 re-run (shown 42 → 40, the fax
> twin gone, the new inline binds clean); fork state (`v0.9.0` tag = `266f7893`, fork
> HEAD `9bf17036`, host binary `8ce90a4` = neither); remote main = `586ff894`.
> NOT verified, and why: **CI verdict** — impossible, the daemon's three commits
> (`92972b4b`, `d8773273`, + the fmt-repair riding) were UNPUSHED at report time
> (remote main `586ff894`); `scripts/ci-verdict.sh` has nothing to read yet. **Full
> `go test -count=1 ./...`** — deliberately scoped out (views-only change surface;
> targeted suites + build covered it); rides the next full battery. **Stack browser
> E2E** — not owed: served markup is byte-identical (proven at component level, see
> below). **Island gates** — no JS touched. **`scripts/format-check.sh`** (the
> authoritative treefmt check) — not run; `nix fmt`'s 0-changed is the cached
> answer, and the changed files (`.templ`, `*_templ.go`) are outside treefmt scope,
> so exposure is minimal but the authority was not invoked.

## The session in one paragraph

Owner ran the LOCAL art-dupl v0.9.0 (`-t 1 --suggest-generics --type-aware --timing
--html`) and said "deduplicate!". Per the registry protocol the paste is evidence of
ITS run, not of HEAD, so the first act was the authoritative pinned-v0.7.2 gate:
**0 new vs 114** — every group the CI universe sees is already baselined. All 42
paste groups were then attributed one by one: 41 re-sightings of standing
rulings/classes, and **one real finding the pinned gate has never seen** — the fax
capability copy-key twin switches (`faxSubtitleKey` ↔ `faxEmptyKey`, the same 4-arm
switch over `cap.Send`/`cap.Receive` differing only in the key-family prefix).
EXTRACTED as `faxCopyKeys(cap) (subtitle, empty string)` — one state mapping
deriving both families in lockstep, full i18n key literals kept (the unused-key scan
excludes `_test.go`, ruling out concat), inline binds at the two single-use call
sites. Verified byte-safe by the existing rendered-bytes pins; recorded in the
registry (ruling row + sweep log) and CHANGELOG.

## a) FULLY DONE

1. **Registry protocol honored end-to-end** — own pinned re-run first (never the
   paste), standing rulings consulted before judging, adjudication before any
   baseline decision, sweep-log line with status link appended.
2. **42/42 paste groups attributed** — 41 to standing rulings/noise classes
   (`newTestServer`+`signIn` call pair ×15, ProjectionHostOptions appends ×3,
   chatloop call-shape residue, per-package writeJSON writers, ErrNoteListFull 409
   arm, timer-drain idiom, wp-nav-badge/thread-when/segcount/avatar templ rows,
   VoicemailSummary two-vocabulary twin, signature/Props declaration noise), 1
   extracted (below).
3. **`faxCopyKeys` extraction** — the twin switches are ONE lockstep mapping now;
   `faxUnavailableKey` correctly left separate (inverse conditions, different suffix
   pattern — not a twin). templ regenerated with the exact go.mod CLI (v0.3.1118).
4. **Verification chain** — `TestFaxPanelDeclaresTheDeploymentCapability` (4 states ×
   2 languages), `TestFaxPanelDefaultMarkupIsUnchanged` (element-pinned literals),
   `TestFaxListEmptyStateFollowsTheCapability` (isolated empty side) prove byte
   safety INCLUDING the family-swap case: a plain-state family swap breaks
   DefaultMarkup's `<p class="wp-panel-sub">Send PDFs as faxes…</p>` exact pin; any
   suffix swap breaks the per-state distinct-string assertions; FaxList isolates the
   empty family. Full views suite (i18n gates) + server suite + build green.
5. **Post-change detector verification** — pinned gate still **0 new** (no baseline
   write owed: the group never existed in the pinned universe); the v0.9.0 binary's
   universe dropped 42 → 40 shown, the twin gone, the two new binds created no fresh
   group.
6. **Records** — registry ruling row (Extracted table) + dated sweep-log line with
   status-doc link; CHANGELOG `[Unreleased]` entry; docs-gates battery green after
   the dprint repair.
7. **On-sight fixes surfaced by the session** — the rotting count in `ci.yml`'s
   dup-gate comment ("86 groups, re-recorded 2026-11-04" vs the file's 114) replaced
   with a pointer to living truth (recordedAt + registry); the sweep-log link gap
   closed.
8. **The detector-pin question answered with evidence** — the fork HAS a clean
   `v0.9.0` tag (`266f7893`); fork HEAD `9bf17036` carries untagged daemon commits;
   the host binary `0.9.0-8ce90a4` matches neither. Pinning the TAG is the only
   reproducible upgrade form.

## b) PARTIALLY DONE

1. **This status report** — written now; the (f) list is a ~30-item honest brainstorm,
   not the padded 50 (see e-6).
2. **Join-body trigger bookkeeping** — today's grep finds the `WriteByte('\n')`
   builder in exactly 2 sites (`app.go:794`, `main.go:262`); the registry row's
   "exactly 3" was true at its 11-04 sighting, the readmodels member has evidently
   dissolved since. The row is historical-by-design ("sites at last sighting") and
   the sweep log carries today's count — but a dated inline annotation on the ROW
   would make the reconciliation explicit (annotate-not-rewrite doctrine).
3. **Detector-upgrade preparation** — the fork tag exists and the path is mapped
   (tag → CI pin bump → one-time re-adjudication of the new universe → re-record),
   but the upgrade itself is an owner call (question g1) and NOT executed.
4. **AGENTS.md templ-pin rot** — noticed (AGENTS.md prose says the pinned compiler
   is v0.3.1100; go.mod says v0.3.1118) and reported here, but the prose refresh is
   not written yet — it wants the bump-train provenance first (a 5-minute `git log
   -G templ go.mod`) so the correction cites its train like the rest of the file.

## c) NOT STARTED

1. **CI verdict on this session's commits** — blocked until the daemon pushes;
   then `scripts/ci-verdict.sh d8773273` (and the fmt-repair commit).
2. **The detector-pin upgrade or advisory-lane decision itself** — awaiting g1.
3. **HARVEST of section (f) into TODO_LIST/ROADMAP** — deliberately not started:
   the owner instruction for this session was "report, then wait"; docs-health
   HARVEST owns the transfer on instruction.
4. **AGENTS.md templ-CLI prose refresh** — see b-4.
5. **Quarterly baseline + suppressed-set audit** — next cadence date 2027-01-11;
   not owed yet, listed so it is not forgotten.
6. **Full-suite run (`go test -count=1 ./...`) + island gates for this change** —
   deliberately scoped out; the next natural battery covers it.

## d) TOTALLY FUCKED UP

Nothing in this session reached that class — stated with evidence, not comfort: no
test failed late, no gate stayed red, no wrong claim survived into the record
uncorrected, no committed code is broken (the daemon's commits carry green
targeted suites + build). The two near-misses are disclosed under (e) with their
costs. The closest thing to a (d) finding is SYSTEMIC, not session-level, and it is
honest to name it here anyway: **the canonical clone gate was blind to a real,
harmful clone for ~8 weeks** (the fax twin lived under a green gate from the
capability train until today, because only the unpinned local detector could see
it). The gate did exactly what it was designed to do — nothing new since the
baseline — and that design is what (e-4) puts on the owner's desk.

## e) WHAT WE SHOULD IMPROVE

1. **Verify-at-source before tool invocation** — I ran `templ@v0.3.1100` from
   AGENTS.md prose without reading go.mod first (the file states the exact-match
   rule in the same breath). Cost: one wasted generation + a version-check warning.
   The durable fix is to stop carrying literal versions in AGENTS.md for anything
   with a living pin: say "match go.mod" and give the `go list -m` one-liner.
2. **Push/CI verification as a hard tail step** — I checked `git ls-remote` only
   when writing this report, and it caught a live fact: the daemon's commits sat
   UNPUSHED. The AGENTS.md rule ("check CI after every push") must be a step of
   every train's tail, not a report-time afterthought.
3. **On-sight fixes shouldn't wait for the report cycle** — both doc rots
   (ci.yml count, AGENTS.md templ pin) were noticed mid-session; one was fixed on
   sight, the other deferred pending provenance. The 2026-09-06 standing permission
   covers two-liners precisely so they don't age into report items.
4. **The pinned-detector blindness (the big one)** — three phantom-red incidents
   bought gate stability by freezing detection at v0.7.2's matcher; the fax twin is
   the receipts. Options for the owner: (a) upgrade the pin to the fork's `v0.9.0`
   tag — the gate sees strictly more, at the cost of a one-time full
   re-adjudication + re-record (all 114 identities re-key; old hash references go
   historical); (b) keep v0.7.2 canonical and run the fork build as a report-only
   advisory lane each sweep (no baseline, no gate teeth). Either is a one-line
   ritual change + registry note; neither is mine to pick.
5. **Never copy counts into prose comments** — the ci.yml comment carried "86
   groups" and rotted within three weeks; the fix points at living truth. Same
   discipline the docs-gates ratchet notes already teach.
6. **Honest quota over padded quota** — (f) is ~30 grounded items, not 50. The
   remaining 20 would have been fabrication, and a fabricated backlog is worse
   than a short one (docs-health HARVEST would enshrine noise).
7. **Argue the safety claim, don't assert it** — my mid-session message said
   "byte-safety proven by the pins" without the swap-ambiguity reasoning; this
   review verified the reasoning holds (a-4). The claim was true but under-argued
   in the moment; the argument belongs in the record the first time.

## f) What we should get done next (honest brainstorm — ~30 grounded items)

1. `scripts/ci-verdict.sh` on this session's commits once the daemon pushes; if the
   push stalls past the 60-min bar, invoke it (anchor: first unpushed commit
   06:40).
2. Detector-pin decision (g1), then execute exactly one of:
3. (upgrade path) bump `ci.yml`'s pin to the fork's `v0.9.0` tag, dry-run the full
   new universe, classify every group against standing rulings BEFORE the pin
   bump, re-record the baseline (registry-protocol act), annotate the ritual
   change in the registry protocol section.
4. (advisory path) add a report-only fork-build lane (local script or CI
   continue-on-error step), registry note defining its non-gating status.
5. AGENTS.md templ-CLI prose refresh (match-go.mod phrasing + `go list -m`
   one-liner; provenance for the v0.3.1118 bump via `git log -G templ go.mod`).
6. HARVEST this report's (f) into TODO_LIST/ROADMAP (docs-health, on instruction).
7. Dated inline annotation on the join-body registry row (b-2).
8. Dated inline note on the AGENTS.md-side stale `v0.3.1100` claim once fixed (5).
9. Registry housekeeping: two "## Sweep log" headings exist (~line 98 + ~line 243);
   consolidate carefully or add a one-line note that the second is the active one
   (annotate-first, the file is append-only by doctrine).
10. Registry note recording the local binary's provenance chain
    (`8ce90a4` ≠ tag `266f7893` ≠ fork HEAD `9bf17036`) so future sweeps don't
    re-derive it.
11. If the advisory lane lands: pre-write standing-class notes for the two groups
    its generics suggestions already nag about (the join-body pair, the
    note/review `Get` pair) so the nag is pre-attributed.
12. Understand the shown-count divergence (pinned 101 vs v0.9 42 at the same
    `-t 1`) before any upgrade — it calibrates the re-adjudication workload.
13. Direct unit table for `faxCopyKeys` — redundant today (coverage proven), do it
    only when the file is next touched.
14. Quarterly baseline + suppressed-set audit (cadence date 2027-01-11): re-derive
    `generic_df`-style staleness, check the 114 for dissolved groups, recount the
    jscpd island-test class (15 → 16 at the last sighting).
15. `scripts/dedup-sweep.sh` wrapper pinning the whole ritual (detector version +
    flags + check + baseline set-diff helper) — the TOOL TRAP lessons are prose
    today; a script makes them un-misrememberable.
16. Sweep-log line numbers rot: ruling rows cite `file:line` from sighting dates
    (e.g. `disposition_api.go:71-74`); a slow pass converting cites to symbol
    names per the cite-stable-names convention (quarterly-audit-sized, not urgent).
17. Fork-repo release ritual: daemon commits sit untagged ahead of `v0.9.0` —
    decide whether/how fork-HEAD daemon commits ever ride a release tag.
18. Upstream relationship (fork of `golangci/dupl`): decide whether detector fixes
    (the short-span `.templ` matching class, art-dupl#14 territory) flow upstream
    via PR (verify-before-filing gates first).
19. Baseline re-record discipline: next clone-introducing train carries the
    ride-along (standing doctrine — restated here only because the fax twin's
    train DIDN'T, which is how the blind spot survived).
20. The reports meta-span trio (`wp-reports-day-meta` ×3) stays baselined per
    do-not-re-litigate; touch it ONLY via the quarterly audit if it still earns a
    `reportsMeta` one-home then.
21. Extraction-shadow groups (the 10-28 lesson "a duplicated CALL is the
    extraction's shadow"): add the class note to the registry PROTOCOL section so
    future sweeps stop re-arguing them.
    ~~22. Full `go test -count=1 ./...` + island gates ride the next battery (c-6).~~ done 2026-11-24 — the deferred full gates RAN GREEN the same day (pareto session): `go test -count=1 ./...` over the extraction-bearing tree (sole non-ok line the no-test-files notice), island node specs 0 failed/skipped, `tsc -p . --noEmit` clean
22. When the daemon pushes, confirm the fmt-repair diff landed canonical markdown
    (battery-on-commit lesson: quote the commit sha in any posture claim).
23. Keep `.crush/shell-output/output-917265081.log` (the paste's raw evidence)
    until the next sweep supersedes it — it is the only byte-exact record of the
    owner's run.
24. If g1 lands as "upgrade": schedule it as its OWN train (one train, one baseline
    write — the 11-09 doctrine), never as a rider.
25. If g1 lands as "advisory": define where the advisory report lands (docs/?/
    .crush/?) so it doesn't entomb.
26. The untracked file in the FORK repo's checkout (`docs/status/2026-11-24_06-44_docs-health-…`,
    another session's docs-health audit) — leave alone; cross-session awareness
    only.
27. CHANGELOG: when the release dance next cuts, fold today's `faxCopyKeys` entry
    into the release notes wording.
28. Re-check that the extraction's sweep-log claim "no baseline write owed" holds
    at next sweep (the pinned gate's universe must still be 0-new at HEAD+1).
29. Teach the next session the one-line version of this finding: "the gate's
    0-new is version-relative truth; the local detector sees a different,
    larger universe" — it is now a standing-class fact worth a registry
    protocol sentence.

## g) Questions I can NOT figure out myself

1. **Detector-pin policy:** upgrade the CI dup-gate to the fork's `v0.9.0` tag
   (catches real clones like the fax twin; costs a one-time full re-adjudication +
   re-record, and every historical hash reference in the registry goes
   read-only-historical), or keep `v0.7.2` canonical and run the fork build as a
   report-only advisory lane? This is a ratified-ritual change (2026-10-11
   sitting) — I will not pick it unilaterally.
2. **Intent of `--suggest-generics`:** was your local run one-off curiosity, or do
   you want the generics-suggestion lane as a recurring sweep surface (it would
   need its own adjudication cadence and pre-written standing-class notes)?
3. **Advisory-lane cadence:** if the fork build stays in the loop, do you want it
   automated (CI report-only step or a local script in the ritual), or is the
   manual local run you did today the intended cadence?
