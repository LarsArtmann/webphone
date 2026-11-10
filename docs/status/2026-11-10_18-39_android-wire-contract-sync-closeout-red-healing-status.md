# Android wire-contract sync CLOSEOUT + the red-healing side-train — 2026-11-10 18:39 CET

> Claims verified as of 18:39 CET. Push/CI statements rot fastest: the
> verdicts below name their runs/shas; the concurrent sessions are LIVE
> and will move both repos past this snapshot.

**Scope line.** This report covers ONE session: the completion of the
android wire-contract sync train (the blocked resume from 15:57) plus
the webphone red-main healing it pulled in. VERIFIED with: android
`scripts/verify-gate.sh` fully green at `a58a68c`; both android remotes
at `a58a68c` via `git ls-remote`; webphone CI run 42855672039
**success** at `fc17582`; buildflow full run + `--failed-only` rerun
(rc=0) on the final tree; `audit-verify.sh` 91/91; docs battery all 9
gates PASS at pins (annotate 203/203, audit 54/72, cadence under
threshold). NOT verified: anything the two CONCURRENT sessions landed
after 18:00 (webphone tip is already past my `fc17582` — their v2.13.0
release train; android has 5 unpushed daemon commits from the
wake-lane session: `708b0d8..26d9c8c`); the android CI verdict (Actions
billing dead there — red CI = billing, not code); live/emulator
behavior of the drift fix (JVM-only authority, by decision below).

---

## a) FULLY DONE

1. **The android half of the wire-contract train — ALL of it, pushed
   both remotes (`a58a68c`).** FcmReceiver + NtfySubscriptionService
   delegate to `WireFacts` (the ONE parse home); `SessionApi` reads the
   REAL wire (exactly-one credential arm `sip_credential`/`password`,
   SIP facts from `/config.js` via `fetchIslandConfig`, honest-null
   when un-registerable); `registerPayload` extracted pure;
   `scripts/sync-wire-fixtures.sh` (copy + `--check` byte-diff gate);
   9 fixtures vendored byte-pinned (dprint-excluded); `WireContractTest`
   — 12 Robolectric tests over every fixture, **meta-proven** (the
   historical `sip_password` drift reintroduced ⇒ brokered-arm red ⇒
   restored ⇒ green); full verify gate green (assemble, tests, ktlint +
   KGP parity, 16 KB zipalign, contrast, dprint); CHANGELOG / TODO_LIST
   / AGENTS / README rows all landed.
2. **The three blocking §g questions answered by evidence + mandate**
   (recorded as struck verdicts in the 15-57 report): wake-lane session
   IS active (their ring-or-silence plan committed 16:12, owner-gated
   waves, zero code overlap) ⇒ standalone landing was right; push both
   remotes with local gates as authority; JVM suite is the local
   authority, emulator stays the device train's job.
3. **d1.1 deferred rows with triggers, both repos**: loopback
   live-check script, `/config.js` key fixture (consider), runtime
   digest handshake mirror — webphone `3f955f2` + android TODO_LIST.
4. **The 15-57 midflight status report fully annotated** — all 14 §f
   items + 3 §g questions carry verdicts; cross-repo hashes cited by
   repo name so the B/hash gate resolves nothing foreign.
5. **Webphone red-main healed through THREE red classes** (each
   attributed before touching): (i) dprint format rc=20 — my own
   report tables, their plan doc, their v2.13.0 FEATURES fold;
   (ii) docs cadence — 68→62 reports via the gates-first archive
   ritual (check-rows --strict + bare-rows BEFORE every `git mv`,
   banners, INDEX regen, all gates green at pins); (iii) their
   `nix/apps.nix` bytes. CI green at `fc17582` (run 42855672039).
6. **A real root-cause fix: the Tailwind CWD leak.**
   `scripts/build-tw-css.sh` now runs tailwind with **cwd = the stage**:
   v4's implicit content detection was scanning the REPO's git tree, so
   plan-doc PROSE emitted utilities into the committed artifacts
   (`.top-5` from "ranks the top-5 ugliest surfaces"). Empirically
   proven (cwd=repo ⇒ 1 hit; cwd=stage ⇒ 0, outputs differ); the
   rebuild shrank both artifacts (−543/+318 lines) of leaked-from-docs
   content; verified NO app source renders Tailwind utilities outside
   the staged views (only `.wp-error`, hand-rolled app.css); the
   tw-freshness advisory job is green in CI for the first time today.
7. **Tail verification**: full `BUILDFLOW_NO_RESULT_CACHE=1` run on the
   final tree — 80 steps, one local-load flake
   (`TestAutoReviewDispatchRoundTrip`: detached-job window missed under
   4-way parallel load; 3/3 green isolated; CI green on identical
   bytes) cleared by the `--failed-only` rerun (rc=0).

## b) PARTIALLY DONE

1. **"Webphone CI stays green" is true AT `fc17582` and already
   stale** — the concurrent v2.13.0 release train moved main past it
   within minutes (their session, their pushes; my last check of their
   newest run was green at `fc17582`). Greenness is now THEIR train's
   to hold.
2. **The midflight report's annotations are complete but the report is
   still LIVE (unarchived)** — by design (it is today's context), but
   it now carries zero open rows and will be archive-eligible at the
   next cadence sweep.
3. **tw-freshness is green but still ADVISORY** (`continue-on-error`) —
   the leak it exists to catch shipped silently for days because
   nobody gated it. The flip to blocking is an owner call (§g-1).
4. **Android verify authority is local-only** — the mirror CI cannot
   run (billing). Every android claim above rests on the local gate +
   pushed tips, not on Actions.

## c) NOT STARTED (deliberately deferred, triggers recorded)

1. Loopback live-check script (fixtures pin SHAPE, nothing pins a
   DEPLOYMENT) — trigger: device-train prep.
2. Runtime contract digest handshake ("server older than app" surface)
   — trigger: first real version-skew incident.
3. `/config.js` key fixture (config keys are regex-read, unpinned) —
   trigger: first config.js drift or second consumer.
4. KMP protocol-layer spike — quiet-repo gate unmet (their repo is
   LOUD right now).
5. Emulator/device verification of the drift fix (a REAL credential +
   domain now reaches REGISTER) — the first device train's leg.
6. The android forgejo→github mirror-lag watch — moot for this train
   (both remotes pushed directly), but the standing PAT/workflow-scope
   row stays open.

## d) TOTALLY FUCKED UP (owned — each healed in-session)

1. **dprint mangled the byte-pinned fixtures.** I vendored fixtures
   into a dprint-owned tree and ran a repo-wide `dprint fmt` without
   asking what it would touch; the pretty-printed fixtures would have
   silently broken the sync gate's byte diff. Caught only because I
   re-ran `sync-wire-fixtures.sh --check` as an afterthought. Fix:
   dprint exclude + re-vendor; the exclude is now part of the landed
   config.
2. **I committed the LEAKED css artifacts before fixing the leak.**
   First rebuild (cwd=repo) landed the `.top-5`-carrying artifacts as
   `99f1139a`; the script fix + isolated rebuild shrank them three
   minutes later (`2830e42f`). For that window, main carried bundle
   content that existed only because a plan doc mentions a utility
   name. Correct order was: finish the leak diagnosis (the cwd
   hypothesis was already in hand), rebuild ONCE.
3. **The B/hash +5.** My §f verdicts cited android-repo hashes as
   `done at <sha>`; the annotate gate resolves hashes against THIS
   repo's history ⇒ 5 findings past the 203 pin ⇒ battery red. Cost an
   empirical bisect (swap-bytes-and-measure) to attribute to MY file.
   The verdict grammar (single-hash `done at` = resolvable-here) is now
   learned the expensive way; cross-repo verdicts say "in
   webphone-android (`sha`)".
4. **My own rc=20 — twice.** The 15-57 report's tables were committed
   un-dprint'd (the exact BATTERY-ON-COMMIT class AGENTS documents),
   and my first §f annotation pass skipped the annotate→format→re-check
   canon. Both healed; both were avoidable by running the battery on MY
   bytes before letting the daemon see them.
5. **Six narrative commits lost to daemon races.** The WireContractTest
   first version, the code half (2c578b5), the fixture/dprint fix, the
   docs batch, the archive sweep (13 files!), and the annotation fix
   all landed as bare `chore:` — including the drift fix itself, whose
   only narrative survives in the test commit + CHANGELOG. I knew the
   "stage+commit+verify in ONE call the moment a unit is green" rule
   and still wrote-then-verified-then-lost, repeatedly.
6. **Two buildflow verdicts conflated mid-narrative.** I described the
   15:32 run's findings-gate ERROR and the 16:07 run's 4 step-failures
   as one story before `buildflow history` separated them (15:32 was
   step-green; 16:07 had the chronic nix flakes). Also piped `tail` on
   backgrounded long-runs twice and destroyed the step tables I then
   had to re-derive via `--failed-only`. Long runs now log to files.

## e) WHAT WE SHOULD IMPROVE

1. **Byte-pinned artifacts + repo-wide formatters don't mix.** Any
   future vendored/generated file needs its formatter-exclude in the
   SAME change as the file lands. (Also: read `git status` output
   AFTER a format run, not before the next step.)
2. **The annotate verdict grammar should be pre-flightable.** A
   `--dry-run --verbose` pass over one spec before a 15-item sweep
   would have shown the B/hash findings for free; the bisect cost 10+
   minutes. Consider teaching the tool a `repo:` prefix for foreign
   hashes (upstream docs-gates ask).
3. **Diagnose before rebuild.** When a rebuild's output surprises
   (`.top-5` from nowhere), the surprising BYTES are evidence — freeze,
   hypothesize, isolate (cwd swap took 2 minutes), then rebuild once.
   Committing surprise bytes to unblock a gate is how leaks ship.
4. **The cadence gate is structurally fighting the report flood.**
   Today: 62 live reports, threshold 65, five active trains writing
   3-5 reports/day. Archiving is becoming a standing tax on whoever is
   last to push. Either the archive ritual becomes part of
   report-WRITING (author annotates+archives their own report at close)
   or the threshold needs re-deriving from the real train count.
5. **The daemon eats narrative history.** Six `chore:` commits today
   carrying semantic work — future forensics (and reviewers) get
   nothing. The hooksPath-override fix lives in the pma module (fleet
   policy, owner); a cheaper local mitigation: commit in the SAME tool
   call as the unit's green gate, every time, no exceptions.
6. **Advisory gates rot silently.** tw-freshness was advisory exactly
   while its subject leaked for days. Advisory should have a sunset:
   N clean runs ⇒ auto-flip to blocking (a small ci.yml comment update
   per flip, but the reminder should be mechanical, not memory).
7. **`TestAutoReviewDispatchRoundTrip` needs a load-proof wait.** The
   detached-job wait window missed under parallel local load (3/3 green
   isolated, CI green). An event-driven wait (channel/poll deadline)
   instead of a fixed sleep removes the whole class.
8. **ktlint trust without understanding.** `ktlintFormat` reordered my
   imports in a way that looked wrong ("fail" sorted after
   "assertTrue"); I accepted it because the gate passed. Understand the
   rule (probably alias/case ordering) or the acceptance is cargo cult.

## f) NEXT (impact-sorted; most route to rows that already exist)

1. **Owner: flip tw-freshness advisory→blocking** now that the build is
   deterministic (stage-cwd) — the silent-leak days are the precedent.
2. **Visual sanity of the SHRUNK tw.css/login.css before v2.13.0
   ships**: one `scripts/ui-capture.py` run (T23 harness) over the
   surfaces the adopted components render; the −543 lines are
   gate-verified but pixel-unverified.
3. **First device train leg**: install a build with the drift fix, sign
   in once, watch REGISTER succeed with the real credential (the
   no-ring chain's first proven link).
4. Wake-lane session (theirs, in flight): their 5 unpushed android
   commits need their wave-end push ritual (both remotes).
5. Harden `TestAutoReviewDispatchRoundTrip`'s wait (event-driven).
6. File the docs-gates upstream ask: `repo:`-prefixed foreign hashes in
   verdicts (the B/hash +5 class).
7. Loopback live-check script (trigger may fire at device-train prep).
8. `/config.js` key fixture + re-vendor (consider row).
9. Runtime digest handshake (trigger: first version-skew incident).
10. Archive-eligibility sweep: the 15-57 report (zero open rows) + the
    16-03 report once their session closes theirs out.
11. Annotate-tool pre-flight habit: `--dry-run --verbose` one spec
    before any sweep (write it into the docs-health skill's sweep
    notes).
12. Android: `WireFacts` could carry `draft_ready` fields
    (`contact`/`draft_id`) when a consumer exists — today the ring
    router ignores them by design; do not model before the consumer.
13. Android verify-gate in CI once billing revives (M035 rehearsal row).
14. Webphone CHANGELOG: the tw.css leak fix + script change are
    code-path changes that owe an [Unreleased] line if v2.13.0 didn't
    fold them (the concurrent release train may have — verify at next
    touch).
15. The `keySet`/`equals` org.json idioms deserve a one-line note in
    the android AGENTS test gotchas (JVM suite keeps rediscovering
    them).
16. Consider an `annotate_max` re-measure note: my sweep's strikes
    net-zero'd (203→203); the pin is honest but a drain pass on the
    17-07-era archived rule-F rows would buy headroom.
17. Android ntfy read-auth decision (F39-F41, theirs) lands soon — the
    fixture set gains an auth-header case if it changes the wire.
18. The 15-57 report's §f-11 (config.js fixture) and §f-8/9 verdict
    wording now carry rare tokens — future freshness sweeps should
    match them; re-verify at the next docs battery.
19. Mirror PAT `workflow` scope + Actions billing rows (owner, J01
    cluster) — still the cheapest unblock for android CI.
20. Emulator pixel train (their M20) after their waves land.

## g) QUESTIONS FOR THE OWNER (cannot be determined from here)

1. **Flip tw-freshness to blocking?** It was advisory exactly while the
   Tailwind CWD leak shipped silently; the build is now deterministic
   (stage-cwd) and green in CI. Blocking means a mid-train daemon push
   of view edits can red main until the artifacts rebuild — the
   tradeoff the advisory comment names. Your CI-policy call.
2. **The shrunk css artifacts (−543/+318 lines) ride the concurrent
   v2.13.0 release as-is, or do you want the visual sanity pass (f-2)
   BEFORE it ships?** The gates say the removed content was leaked from
   docs prose; pixels have not confirmed it. The release train is live
   right now, so this answer is time-sensitive.
3. **Should the commit-daemon's hooksPath override be fixed fleet-wide
   (pma module, GIT_CONFIG env override) so daemon commits run the
   guard hooks?** Today it again committed dependency-adjacent and
   formatter-hot files mid-edit; the AGENTS documents the guard hole as
   a known fleet-policy deferral. It ate six narrative commits this
   session alone.

---

## Concurrent-session boundary (kept clean)

The v2.13.0 release session (webphone) and the ring-or-silence session
(android) were live throughout. I never staged their in-flight files
(healed only their COMMITTED bytes, mechanically, each time
attributed); their unpushed android commits (`708b0d8..26d9c8c`) and
webphone tree (.buildflow.yml, docs/lessons.md) are theirs, untouched.
