# SUPERB plan execution, trains A–C: the storage-ladder completion session — status at 2026-11-12 04:42 CET

> Claims verified as of **04:42 CET** (push/CI statements rot fastest; every verdict below was
> re-checked at that stamp). Mandate: execute the SUPERB 80/20 plan
> (`docs/planning/2026-11-11_20-25_SUPERB-storage-ladder-and-ui-craft-80-20.md`) —
> "GET SHIT DONE. The WHOLE TODO LIST." Session ran 2026-11-11 ~20:30 → 2026-11-12 04:42.

## a) FULLY DONE

1. **M01 — the three owner rulings, recorded** (the plan gated on them; the blanket GO
   stood in for the owner sitting, decisions made autonomously and written down for
   override):
   - _Fix-forward protocol_ (concurrent sessions' broken gates: confirm non-authorship →
     work non-overlapping → bounded-wait re-run → tracker issue naming surfaces; never
     revert) → AGENTS.md § Concurrent sessions.
   - _TODO strike canon_ (evidence-bearing inline strikes win; only genuinely obsolete
     rows get deleted, with the why in the commit) → AGENTS.md § Conventions.
   - _Detached-review shutdown race_ (the `TestAutoReviewDispatchRoundTrip` sandbox-only
     `sql: statement is closed`): **ACCEPT AND DOCUMENT** — test-harness-only, prod drains
     in order; reopen trigger = a second occurrence outside the buildflow sandbox →
     docs/lessons.md § Detached-review shutdown race.
2. **M02 — storage-ladder train A: the transcripts engine store is DELETED**
   (`readmodels/transcripts.go` −455, its test −442, admin_reads −58). The gate+view
   surface moved to `transcriptsgate.go` (TranscriptGate, TranscriptSegmentView, both cap
   constants, ReassembleTranscripts). Server harness rides `cqrs.NewTranscripts` (the
   gate's Append already awaited its own projection fold). Ported coverage: the admin
   RecentAll pin, the 2230-segment window/caps table (direct projection folds — the
   raw-seed twin), the owner-cap trim (5010→5000 through EnforceCaps), the
   metacharacter-literal search pin, and the 8×25 concurrent-append no-loss test; the
   per-call floor stays decision-pinned (`TestTranscriptPerCallFloorEvictsOldest`).
   Registry regen (−8 `readmodels.transcript_*` codes); art-dupl: 0 clones reference the
   deleted file. Full `go test -count=1 ./...` green.
3. **M03 — train B: notes + reviews engine stores DELETED.** Gate/view homes
   `notesgate.go`/`reviewsgate.go` (gates, views, the 2000-cap constants, the sentinels
   with their Conflict/Rejection classifications). Engine lost the fields/accessors/Plan
   entries; `NewCoreSystem` grew the two decls; the harness drives `cqrs.NewNotes`/
   `NewReviews`. Ported the missing slot-shape pins (verdict-first slot creation,
   empty-body-clears-still-a-slot). **Both backends' write doors gained the D7
   `awaitOwnWrite` beat** — the aggregate header's "no D7 await" premise ("nothing
   re-reads in-request") retired with the twin: the History panel re-fetches right
   behind the SSE nudge, so read-your-writes is the honest contract. Registry −7 codes.
4. **M04 — train C: contacts + snippets + dialwindows engine stores DELETED.** Gate/view
   homes `contactsgate.go`/`snippetsgate.go`/`dialwindowsgate.go`. The Engine lost the
   three fields/accessors/Plan entries AND the aggregate-reader assertion (its last
   consumer was the contacts metrics count); the userless `notFound` envelope helper +
   its test retired with their last callers; the engine-store `dial_bench` went with its
   store. The harness + fuzz spec + API cap pin + store-failure lanes all boot the
   journal through ONE new `newCoreSystem` helper; the 500-contact API cap test seeds
   through the backend; the live-engine delete-miss envelope pins retired (the cqrs
   backends' own delete tests carry the contract). Registry regen — which also caught
   the CONCURRENT train's uncommitted codes (verified: HEAD's freshness test was RED
   before my regen, green after; the stash round-trip proved it). Committed `0e231da0`
   (narrative), pushed.
5. **CI red-main heal on `6485fd44`** (run 43090522377): the passkey ceremony's `user_id`
   parse prologue twins → extracted `passkeyUserIDQuery` (the one home both finish
   handlers share — the sessionAnswer pattern; commit carries `auth:` prefix, the
   session/passkey behavior suite green); the devices adapter/fake fold + the slot-Get
   twins → adjudicated ACCEPT into docs/dedup-registry.md (table row + sweep-log entry:
   test-fake class + deliberate per-seam-code parity). **Baseline re-recorded 137 → 117
   groups** — the ladder deletions dissolved twenty groups; the ratchet moved DOWN, the
   doctrine's expected direction. Post-write `check` = 0 new. The same run's tw.css
   freshness advisory was verified ALREADY HEALED at HEAD (the sibling's 37af2d80 rebuilt
   the artifacts; a local rebuild is a no-op).
6. **Cadence drain** (the docs-gates battery caught corpus growth over the 65-report
   threshold): four CLEAN archive candidates passed both annotate gates
   (`check-rows --strict` + `bare-rows`) before the `git mv`, banner first line
   `> ARCHIVED 2026-11-12 (docs-health v11 sweep)…`; 68 → 64 unarchived; INDEX.md
   regenerated; battery green.
7. **CHANGELOG + ledger for all three trains**: Removed entries (transcripts; notes/
   reviews; contacts/snippets/dialwindows + the notFound retirement) and three
   evidence-bearing strike rows in TODO_LIST (train A/B/C, each with the recipe, the
   ports, and the regen numbers).
8. **Release check (M13)**: v2.14.0 was ALREADY cut and pushed by the concurrent session
   before this one started (tag `v2.14.0` → `d0448143`, CI green ×2 watch runs) — marked
   done-by-sibling, not re-run (the plan's guardrail: attribute, don't re-litigate).

## b) PARTIALLY DONE

1. **M05 — the closeout train (push devices + agentaudit → readmodels = gates + views
   ONLY)** is IN FLIGHT: `pushgate.go` is written; the remaining steps are scoped and
   verified (agentauditgate.go; delete push.go/agentaudit.go/export.go/**readmodels.go —
   the Engine itself**/readmodels_test.go; `readmodels.Open` has ZERO prod consumers —
   the five test sites to flip are enumerated: idempotency_contract (via
   `sys.Dedup()`), ring_test, auth_audit_test, push_api_test, summary_api_test;
   `NewCoreSystem` grows the last two decls; the plan's purity arch test still to write).
   Stopped mid-train to write this report per the report-then-wait directive.
2. **Main is RED at the report stamp** — 0e231da0/0131bd0a/b0e71aaf all carry two
   failures, BOTH attributable to the concurrent session's live UI train:
   (a) `treefmt-check`: the daemon committed their in-flight `calls.js` bytes in
   non-canonical form (prettier wants the `/** @param */` cast wrapped); (b) the
   tw.css/login.css freshness advisory: their `app.css` edits landed without the artifact
   rebuild. Everything else on the run is green (race suite, nix build, smoke, art-dupl
   check = 0 new — the heal held). Per the just-recorded fix-forward protocol I did NOT
   touch their in-flight files (calls.js/sipverdict.js/app.css are uncommitted in the
   tree RIGHT NOW — actively being edited); their train closes both the moment it
   lands fmt + artifacts, exactly as their 37af2d80 did for the previous advisory red.
   The */15 watcher owns the tracker issue if it stands.
3. **M06–M08 (UI tokens/state/focus/grid)**: LANDED BY THE SIBLING mid-session
   (`37af2d80 feat(ui): state layers, focus-ring tokens, type ramp, nav fit, craft
   sweep`) — marked theirs per the plan's guardrail; I verified the artifacts were
   rebuilt fresh (no-op rebuild) and marked the plan rows done-by-sibling. M09–M12 (the
   verify passes: eyeball matrix, hover capture, messages polish, empty-state upgrade)
   were NOT started by either session at this stamp — ownership question in §g.

## c) NOT STARTED (from the plan, beyond the above)

- **M14** owner prod window (T05 authenticated capture, night-run pair, ring-guard
  positive path, diff noise-floor calibration) — owner-gated: needs the live prod host
  - the capture credentials; not present in this environment.
- **M15/M16** stack browser E2E (front door + voicemail playback arms), stack redeploy
  riding main, host diagnostics, webphone-relock completion — consuming-stack repo work
  (`nix-international-telephony`); I did not touch the stack this session.
- **M17** go-health-dashboard nonce lane (upstream implementation + tag + re-pin + CSP
  flip) — separate repo, a full upstream train of its own.
- **M18** bandit-provider issue file — still owner-banner-gated (the draft's checkmark
  was never given).
- **M19–M21** Trains 2/3 (History table, Reports dashboard, mobile dialer) — the plan's
  guardrail assigns these to the sibling's session unless handed over; untouched.
- **M22–M27** the tail (ringhook operator docs + ALERT_WEBHOOK_URL decision, android
  specs half, hygiene batch M24 — partially overtaken by this session's registry/
  lessons work — hub runner fix, wire-fixture trigger adjudication, standing rituals:
  erraudit monthly not due until 2026-12-08, dedup quarterly 2027-01-11).

## d) TOTALLY FUCKED UP (own admission, root cause each)

1. **The daemon ate my narrative commits THREE TIMES** (trains A, B-core, and the
   red-main heal landed as `chore: auto-commit N changed file(s)` blobs mixing my work
   with the sibling's push/UI refactors). Root cause: the docs battery + full suite run
   5–20 minutes between "code done" and "commit", and the daemon sweeps within seconds
   of any file going quiet. I did NOT rewrite the daemon commits to restore narratives —
   the live sibling's HEAD sits on them and a rebase would sabotage their session (the
   documented 2026-10-27 hazard). Cost: history is daemon-shaped for three trains; the
   narratives survive only in this report + the ledger strikes. Train C got its
   narrative commit (`0e231da0`) because I committed the moment the battery passed.
2. **I invented a struct instead of copying it**: my first `contactsgate.go` wrote a
   trimmed `ContactView` (no ID/CreatedAt) from memory; the build caught it in the cqrs
   fold and I had to recover the real shape from git history. Should have diffed ALL
   THREE view shapes against history before writing the gate files (snippets + dial
   were then verified that way — but that verification should have been step zero, not
   a repair).
3. **Two failed edit rounds on `transcriptsstore.go`'s comment sweep** — I pattern-matched
   the `//cqrs-lint:ignore(F031)` comment onto the wrong twin (it lived on the readmodels
   file I was deleting, not the cqrs one) and burned two tool calls. Same class as (2):
   edit-against-memory instead of edit-against-view.
4. **CHANGELOG writes lost races twice** (the edit tool reported success while the
   daemon/sibling had re-read the file between my read and write; the entry silently
   vanished once — caught by the immediate grep, re-applied atomically via python).
   The workaround (atomic python write + immediate verify) should have been the default
   from the first "file modified since last read" error.
5. **A modified `.githooks/pre-commit` sat in the tree for ~40 minutes unnoticed** (not
   mine; appeared mid-session among the sibling's files). I flagged it only while
   writing this report. Unresolved — see §g-2.

## e) WHAT WE SHOULD IMPROVE

1. **Commit at unit granularity, not train granularity, when the daemon is sweeping**:
   the battery-before-commit rule (adopted after three red docs tips) and the daemon's
   seconds-scale sweep are in direct conflict for anything longer than ~5 minutes. The
   working synthesis: commit CODE first (no battery owed), run the battery, commit MD in
   the same tool call — one train becomes two commits, both narrated.
2. **The freshness-regen is a SHARED contract across concurrent sessions**: HEAD was red
   on `TestErrorCodeRegistryIsFresh` because the sibling's 17-file sweep added codes
   without the `-update` regen. The convention says "same commit"; with two live
   sessions the honest statement is "same push" — whoever pushes next carries the
   regen. Worth one AGENTS sentence.
3. **View-shape extraction from `git show <sha>:<path>` should be a written step in the
   ladder recipe** (the faxes row said "moved", never "copy the struct verbatim from
   history"). One sentence in the struck recipe row saves the next train the repair.
4. **The tw-freshness lane reds CI while labeled advisory** — an advisory that exits 1
   is not advisory; either the lane's exit code or its label should tell the truth.
5. **`newCoreSystem` should have existed a train ago** — three separate test helpers
   were booting the journal by hand until train C unified them; the M05 flip would have
   been pure deletion if the helper had landed with train A.

## f) UP TO 50 THINGS TO DO NEXT (ranked; the plan's IDs kept where they apply)

1. **Finish M05** (in flight): write agentauditgate.go; delete the six readmodels files
   (incl. the Engine itself); flip the five test sites; NewCoreSystem += the two decls;
   write the readmodels-purity arch test; registry regen; art-dupl; full suite; battery;
   commit.
2. **Verify CI on the current tip once the sibling's UI train lands** (the two red lanes
   are theirs); re-run the battery if their landing sweeps my unstaged pushgate.go.
3. **Heal-or-wait decision on the treefmt red** (calls.js canonical bytes): after the
   sibling lands, verify; if the red survives their landing, run `nix fmt` + commit the
   reflow as the fix-forward.
4. ~~**Add the AGENTS sentence on the shared freshness-regen contract** (§e-2).~~ done — the SAME COMMIT bullet now carries the two-live-sessions SAME PUSH reading with the stash-proven evidence.
5. ~~**Add the "copy view shapes from history" step to the ladder recipe** (§e-3).~~ done — the Markdown formatter-UNOWNED rule now carries the recipe step zero: copy projection view shapes verbatim from `git show <sha>:<path>`.
6. **Investigate the modified `.githooks/pre-commit`** (§g-2) — whose edit, what it
   changes, gate the commit-msg contract selftest if it moved.
7. **M09** visual verify pass (≈18 shots × light/dark + reduced-motion) — ownership
   question in §g-1.
8. **M10** hover-state capture (`--hover` flag on ui-capture.py) + fix what it exposes.
9. **M11** messages polish (bubble burst grouping, rail 2-line clamp, pins, captures).
10. **M12** empty-state upgrade (icon-in-tonal-circle + action affordance; templ regen +
    the three pin sites).
11. **M14** owner prod window (T05 capture, night pair ≥8h, ring-guard path, diff noise
    floor) — needs §g-3.
12. **M15** stack browser E2E: front-door flow + voicemail playback arms (zero-config +
    default) vs the deployed flow.
13. **M16** stack redeploy riding main + host diagnostics read + the relock ritual rows.
14. **M17** dashboard nonce lane upstream (data-nonce + csp.go) → tag → re-pin webphone
    (vendorHash tail: tidy → re-pin → push → CI watch) → flip the `/health` CSP assertion
    → delete the unsafe-eval exception.
15. **M18** bandit-provider issue file (after the owner banner check).
16. **M19–M21** Trains 2/3 (History table; Reports/contacts/settings; mobile dialer) —
    sibling's unless handed over.
17. **M22** ringhook operator docs (stack deploy recipe) + typed webphone-module front
    design note + the ALERT_WEBHOOK_URL decision recorded + secret set.
18. **M23** webphone-android specs consumption (the wire-contract sync half).
19. **M24** hygiene batch: the synctest+projectionhost bubble lessons row; the
    runbook archive-banner recipe row; art-dupl clean set-diff check (the 11-11 sweep
    already exercised the clean set-diff — the row may strike with this session's
    evidence); lychee redirect-gap re-check.
20. **M25** hub runner durable fix (ExecStartPre provisioning) + hub TODO sync.
21. **M26** wire-fixture follow-ups (a)–(d) adjudicated against their recorded triggers.
22. **M27** standing rituals: erraudit monthly re-measure (due 2026-12-08), cadence
    candidates pass (59 DIRTY candidates owe drain trains), dedup quarterly prep.
23. Drain-train the next-oldest DIRTY cadence candidates (59 owe work; the 19-15
    report's §b-1 "register the 3 error codes" item is the battery's named UNHARVESTED
    — strike it with this session's regen evidence when the report is drained).
24. Sweep the remaining "the readmodels engine declares" parity comments in the cqrs
    contacts/snippets/dialwindows/pushdevices/agentaudit files (their twins are dead
    NOW — the sweep was per-train; train C swept only notes/reviews).
25. `export.go`'s `exportAllViews` dies with M05 — verify no other caller after the push/
    agentaudit deletion (it was shared by the deleted ExportAll methods).
26. Update the ONE-journal-world AGENTS bullet to "doctrine complete" after M05 (the
    bullet still says the Engine is a TEST implementer — after M05 there IS no Engine).
27. Backup-set mention audit after M05 (AGENTS says readmodels.db left the backup set —
    confirm the backup docs no longer name any engine artifact).
28. Re-check `Engine()`/`SQL()` accessor consumers (app.go health probe wiring) before
    deleting readmodels.go — the probe's critical-service contract names `sqlite` +
    `blob-dir`, but the wiring-test parity (`TestAppWiresEveryServerDep`) may reference
    the Engine.
29. Sweep `dialWindowsGlobalCeiling`-class comments in cqrs/dialwindows.go that name the
    readmodels declaration (same class as the notes/reviews sweep).
30. Consider hoisting `humanScaleVolume`'s canonical home comment (readmodels.go dies in
    M05; the const lives in threadsviews.go — the comment naming the deleted engine's
    rate example should move with it).
31. Delete `openTestEngine` from readmodels_test.go in the M05 commit (it exists only
    for push/agentaudit tests, which die with the train).
32. After M05: `internal/readmodels` package doc rewrite (the header still narrates
    readmodels.db + the DDL-ownership doctrine — keep the history, state the gates+
    views-only present).
33. Verify the island suite + smoke pins still green after the sibling's sipverdict.js
    train lands (their files, their pins — but the shared suite verdict is mine to
    check before my next push).
34. Night-run pair scheduling (part of M14) — even without the capture, the cron/
    reminder half can be recorded.
35. `scripts/ci-verdict.sh` on every push (this session ran it late twice).
36. Re-check whether `.art-dupl-baseline.json` should ride a `chore:` commit separately
    from narrative commits (the daemon mixed it into e3b39417).
37. The 59 DIRTY candidates' §b "RED MAIN" items are all historical (three regen cycles
    since) — a bulk-strike pass with the standing evidence would drain the DIRTY list
    faster than one-train-one-report.
38. Record the D7-await addition in the AGENTS D7 bullet (the await now covers notes +
    reviews + transcripts + contacts + snippets writes — the bullet's list may be
    stale).
39. Check whether `TestAppWiresEveryServerDep` (the wiring-parity pin) needs the M05
    Deps changes mirrored in app.go's literal (the push/agentaudit seams stay — only
    the test doubles change).
40. Sweep the SUPERB plan doc's errata section — mark M02/M03/M04/M06–M08/M13 with
    their landing commits (the plan is the canon; its errata should name what landed).
41. The `cqrs-lint` pinned detector run over the touched cqrs files (the plan's errata
    says six findings, two fixed — re-run after M05's store deletions to confirm the
    count moved).
42. `scripts/audit-verify.sh` replay over the fleet-libs reports (the touched reports
    claim code states my trains changed — the verify blocks may have drifted).
43. Consider a `docs/status/` note for THIS session's report being the third storage-
    ladder report (the INDEX regen already counted it).
44. The CHANGELOG `[Unreleased]` block is growing — the next release dance (2.15.0)
    will need the section split; not urgent.
45. Re-verify the `webphone-module` flake check after M05 (the module-check stand-ins
    don't touch readmodels, but the VM tests boot the binary — the Engine's death
    changes boot).
46. Cross-check `internal/retention/retention.go`'s comments (they name the readmodels
    engine's frozen lane — post-M05 the lane is fully gone, the comment may over-claim).
47. The `enroll.go` readmodels file still hosts EnrollGate + EnrollTokenView — verify
    their consumers after M05 (they were kept in train A as "the live cqrs seams"; the
    Engine's death doesn't touch them, but the file's header may need the same rewrite
    as the package doc).
48. Run `scripts/whitespace-drift.sh --selftest` once after the sibling's hook edit
    lands (their `.githooks/pre-commit` change may affect GATE 1).
49. The `docs/research/` cqrs-htmx round-3 audit's refresh trigger ("a bounded AuditLog
    projection") — unchanged, but the audit's adoption claims name readmodels stores
    that no longer exist; a one-line errata in the audit doc would keep it honest.
50. Book the next full `buildflow` at the next train START (this session ran suites
    piecemeal; the full run is owed at M05's end alongside the docs battery).

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **M09–M12 ownership**: the sibling landed M06–M08 (`37af2d80`) and is mid-flight on
   a sipverdict/hold-resume island train right now. Are the four verify/polish passes
   (eyeball matrix, hover capture, messages polish, empty-state upgrade) ALSO theirs,
   or should my session take them once M05 closes? I will not touch their lane without
   the ruling.
2. **The modified `.githooks/pre-commit`**: an uncommitted edit to the pre-commit hook
   appeared in the shared tree mid-session (not mine; the daemon hasn't swept it in
   ~40 minutes, which suggests a session is deliberately holding it). Is it yours or
   the sibling's, and should I leave it untouched (my default) or review/adjudicate it?
3. **M14's credentials and window**: the T05 authenticated prod capture needs
   `WEBPHONE_CAPTURE_EXTENSION`/`WEBPHONE_CAPTURE_PASSWORD` against `pbx.artmann.tech`
   and a live inbound call for the ring-guard path. Do you want to run that window
   yourself (I prepare the exact command lines), or hand me the env values and a
   quiet window so I execute the full matrix + the ≥8h night pair autonomously?

---

**Verdict**: the plan's 1% (the storage-ladder, S15) is 4 of 5 trains landed and
verified; the fifth is mid-flight with every consumer enumerated. The 4% gained the
rulings batch (S16 — recorded, awaiting owner override), the release (S5 — sibling,
shipped), and a red-main heal that took the art-dupl ratchet DOWN twenty groups. Main
is red at this stamp on two lanes owned by the concurrent UI train, both expected to
heal with its landing; everything my trains touched is green at HEAD (race suite, nix
build, smoke, art-dupl 0 new).
