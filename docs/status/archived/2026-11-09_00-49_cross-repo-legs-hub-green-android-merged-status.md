> ARCHIVED 2026-11-11 (cadence sweep, this session): gate-clean (check-rows complete + bare-rows clean before the mv). The cross-repo legs it tracks shipped — the hub merged, the Android repo merged; v2.13.0 released 2026-11-10.

# Cross-repo legs executed: ringhook wired, hub green, android merged, main healed — full session status

Snapshot-rot header: claims verified as of **2026-11-09 00:49 CET** (local
runs and CI verdicts timestamped inline; push/CI statements rot fastest).

Scope of this report: THIS SESSION ONLY. Owner's standing order was
"READ, UNDERSTAND, RESEARCH, REFLECT … execute the WHOLE TODO LIST",
resumed after the 21-48 report's STATUS-THEN-WAIT. The session executed
the remaining portfolio legs (stack Leg 2 wiring, P4-9 android, P4-10
hub re-pin, P4-7 filings), healed webphone main from five red legs to
full green twice, and repaired three pieces of cross-repo CI
infrastructure it found broken on the way. Both other sessions' work is
attributed where their trains own the remaining red.

---

## a) FULLY DONE (verified this session)

1. **Webphone main healed from five red legs to full green — twice.**
   Run 42528105845 (the heal train `b029f7c1`) turned docs battery,
   format, art-dupl, nix build and smoke green; runs 42531121332 and
   42534218876 (HEAD `9c5de5bb`) concluded **SUCCESS with every leg
   green**, go test -race included. The five reds at session start:
   the docs battery (annotate 204 > pin 203 + 20 unformatted files),
   `TestErrorCodeRegistryIsFresh`, the art-dupl new-clones gate, the
   docs battery's `format` gate, and a flaky crm test riding CI.
   Evidence: `gh run view 42534218876` — all ✓.
2. **The registry red healed** — the sibling's committed thread
   aggregate added the `cqrs.thread*` code family without the
   same-commit `-update`; the generated block in
   `docs/error-contract.md` regenerated (+26 lines, 26 new rows),
   arch suite green, error-context router selftest 37/37.
3. **The art-dupl red adjudicated, not blessed blind** — baseline
   re-recorded with the CI-pinned v0.7.2 detector, 125 → 131 groups,
   and the six new groups adjudicated BEFORE the write (guard
   prologue, machine-fold arms, the dispatch twin the source comment
   names and the registry-visibility doctrine FORCES, append trivia,
   two journal-vs-legacy migration twins — the sixth group found via
   a baseline set-diff after the CI log proved truncated). Sweep-log
   row appended to `docs/dedup-registry.md`. Post-write check = 0 new.
4. **The crm DefaultTransport flake root-caused and fixed at the
   source** — `TestContactContextFailureClassesStayDistinct/absent`
   failed in CI with `http: CloseIdleConnections called`: the no-CA
   client rode the process-global transport while every
   `httptest.Server.Close` in the test binary closes idle connections
   on exactly that global (stdlib-verified at
   `net/http/httptest/server.go`'s "help them out" block). Fix: the
   client clones the default transport (the CA path already did);
   production-neutral, CHANGELOG Fixed entry, 5× `-race` green.
5. **Docs battery green**: my report's wrap-point F/spacing reworded
   to prose (the known dprint-strips-space corollary), the sibling's
   20 unformatted files formatted in place by the battery's own fmt
   step, and the Phase-5 report's three unharvested §f items
   harvested to TODO_LIST (stack browser E2E, sys.Dedup()
   unification, the metrics-count decision). Battery matrix 9/9 PASS
   at session end.
6. **Stack Leg 2 (the ringhook wiring) SHIPPED end-to-end**:
   `services.telephony.ringhook.{enable,port,webphoneUrl,
   webhookSecretFile}` options (options.nix), the `ring.nix` service
   module (hardened systemd unit mirroring messaging.nix: DynamicUser,
   ProtectSystem=strict, LoadCredential webphone_secret,
   Restart=on-failure, after webphone.service; the secret DERIVES from
   the messaging bridge's gatewaySecretFile when that lane is on),
   `pbx.nix` derives the loopback ringHookUrl, and the FreeSWITCH
   dialplan arms render ONLY when enabled — off stays byte-identical.
   Assertions: webphone.enable + a secret file required.
7. **Stack Leg 2 verified before push**: `tests/test_ringhook.py`
   (7 stdlib-unittest specs: signed forward, end fact, malformed 400s,
   dead-webphone 202, missing-credential 503, health, 404 paths) —
   351-test discovery green; the new `ringhookCheck` eval arm (unit
   env + credential derivation + rendered curl arms + the off-dialplan
   carries none) threaded into `telephony-eval` (RC 0); dialplan-sim,
   changelog-headings and format checks green; `nix fmt` canonical.
   Pushed `89eca55..fccc1f4` (8 commits incl. the daemon's sweeps).
8. **Stack CI legs green on the pushed tip** — run 42529951959:
   aarch64 VM test ✓, stdlib python suites ✓, and the
   `nix flake check` leg finished in **46m40s — under the old
   mod_enum ceiling** — with every VM suite green except ONE
   (attribution below). Leg 1's crm.caFile confirmed pre-landed;
   Leg 2 verified in-tree.
9. **P4-9 android: the Dependabot PR merged with honest local
   verification** — Actions is billing-blocked on the mirror account
   (jobs never start; documented), so PR #7 (AGP 9.4.1→9.5.1, one
   line) verified through the repo's own full `scripts/verify-gate.sh`
   (build, unit tests, ktlint, 16 KB zipalign, contrast 50/50, dprint)
   on the PR branch, then merged (`1ba926b`). Also fixed main's
   pre-existing dprint drift in `.config/metadata.yaml` (`2ef4ac7`) —
   the verify gate's step 6 had been red-for-everyone since 11-01,
   masked only by the billing outage. Zero open PRs remain.
10. **P4-10 catalog hub: run #240 GREEN end-to-end and published**
    (export, merge, governance lint, 1417-page site build, dist
    assertions, publish) — reached through three real repairs (below).
    The webphone Forgejo mirror force-synced to the GitHub tip first
    (`mirror-sync-sources.sh webphone` → ✓ at tip), per the hub's own
    ritual.
11. **Hub repair 1 — the CI host's runner state-root ownership**: five
    consecutive hub failures ("Build-state root … missing or
    unwritable after 6 retries") root-caused to a NixOS switch
    renumbering the runner's DynamicUser — the state dirs sat owned by
    the old uid 63182, group-r-x only, while the job now runs as
    62668. Repaired in place (chown to the live uid), re-dispatched;
    the durable-fix row (ExecStartPre provisioning or static uid) is
    filed in the hub's TODO_LIST (commit `b1a4353`).
12. **Hub repair 2 — the crm source's stale go-modules vendorHash**:
    with the mount fixed, the build died on a hash mismatch in the
    kith-crm go-modules FOD at the `66f37f3` tip. Placeholder probe →
    read `got:` → applied (`crm` `37a1957`), crm builds green, mirror
    re-synced.
13. **Hub repair 3 — the workflow's step-scoped variable**: run #239
    built the ENTIRE site green (1417 pages) then died with
    `current_contract: unbound variable` — the Verify-dist step copied
    its argument list from verify-site.sh but not the derivation line
    feeding `--contract`, and each step is its own shell. Fixed in
    `7551d4a` (derive in step, comment why), verify-gates self-test
    parity intact. The ci-log "first failure" banner had
    misattributed this to the webphone governance lint — the lint was
    clean (proved by a local reproduction at the correct resource
    root).
14. **P4-7 upstream filings batch closed out** — the two
    branching-flow drafts (phantom scoping, panic FP) were ALREADY
    FILED and CLOSED upstream (#8, #9, 2026-10-11/17; the plan's
    "refresh + post" sub-items were stale). file-size-check's
    configurable cap: ALREADY SHIPPED upstream (buildflow #91 closed
    COMPLETED 2026-11-06 — the flag-Changed-gated override + regression
    test); no filing owed. The two real asks now have committed drafts:
    the go-idempotency contract clock race
    (`docs/drafts/2026-11-08_go-idempotency_contract-clock-race.md` —
    all five verify-before-filing gates run at go-idempotency HEAD
    `9a82577`, no clock seam exists, TimingScale is the only lever and
    its own docs admit the wall-clock design) and the erraudit tree
    same-name dedupe
    (`docs/drafts/2026-11-08_erraudit-tree-same-name-dedupe.md` —
    reproduced live at erraudit `c78b8db5`: 4 of 6 same-named sentinel
    declarations hidden, the surviving pick arbitrary). Both bodies
    pass check-draft (0 FAIL, 0 WARN); both carry the unsolicited
    banner UNCHECKED — the sitting pack's D-slots are unanswered, so
    posting is not the agent's call (P4-7's own acceptance allows
    "committed drafts with exact reproductions; checks pass").
15. **The 4 known-red server pins on webphone resolved by their owner**
    — the go-test reds (admin fallback, thread pagination, metrics,
    SSE fragments) were the sibling session's named "4 known-red
    server pins"; their resumed session fixed them the same evening
    (runs 42531121332+ green). Correctly NOT touched by this session:
    taking over a mid-flight core swap would have been split-brain.
16. **TODO_LIST + handover pack carry every verdict with run ids** —
    the cross-repo row's per-leg state, the ring-while-closed stack
    leg flipped from "must POST" to "LANDED", the catalog row's
    executed state, the pack's three leg banners + the run-record
    (commit `9c5de5bb`).
17. **Housekeeping clean**: git worktree + temp probe files removed;
    the four repos' end states surveyed (webphone clean of my work,
    stack clean, android clean + pushed, hub clean + pushed).

## b) PARTIALLY DONE

1. **Stack Leg 3 acceptance (twice-consecutive-green CI)** — the
   ceiling half is PROVEN (46m40s under the old 1h envelope, with the
   enum exclusion landed and ruling-commented), but the run red-ed on
   ONE suite: `telephony-voicemail`'s quota-exceeded log assertion at
   179s. That red is the sibling stack session's OWN named-open item
   (their 21:49 report title carries it: "quota model correction red")
   from their MWI/voicemail train — not this session's legs (ringhook
   rides a default-off byte-identical dialplan there). The second
   green cannot be earned until their train heals it.
2. **The transport-isolation hardening class** — the crm client now
   clones the default transport, but the SAME asymmetry (a bare
   `&http.Client{Timeout: …}` riding the process-global transport) is
   still present in `pbx`, `asr`, `matrix` and `paperless`. Production
   is unaffected (one client per process); only parallel-test flakes
   of the same class lurk there. Harvested as a TODO row this session.
3. **The erraudit tree defect itself** — reproduced, verified at
   source, and drafted; the FIX (package-qualified rows and/or a
   deterministic pick, possibly reaching go-finding's MergeIter) is
   not written. Own tool, near-zero fix cost per D17's own reasoning.
4. **The catalog hub's freshness loop** — run #240 published, but the
   hourly sync pulling it to catalog.home.lan and the site-freshness
   WARN clearing were not directly observed before this report; the
   publish artifact is verified, the serving side is presumed on its
   documented cadence.
5. **Android CI truth** — the merge is locally verified and pushed,
   but NO CI run has executed on the merged main (billing); the
   post-billing green is still owed.

## c) NOT STARTED

1. **P4-8 native first-device train** — SIP UA vs live WSS, incoming
   INVITE, real CallStyle, origin picker, sessionIdentity assert
   (needs the physical device; OWNER+assist).
2. **P4-11 release announcements** — v2.9.x drafts refresh, owner
   picks channels (OWNER).
3. **P4-12 standing-watch sweep** — the erraudit tier-1+2 monthly
   re-measure is OVERDUE (next calendar date was 2026-11-05, last run
   2026-10-05); the named triggers unchecked this session.
4. **P4-13 parked-decisions re-confirmation** — the one-read-through
   pass over the parked audit (session messenger, ssetest ride, KMP
   spike, #8 dimensions, scheduled sends).
5. **P4-14 coverage floor** — sitting-gated (C5 GO pending).
6. **The sitting pack's unanswered D-slots** — D1–D18 (including D17's
   filing-authorization and D10's cap posture) all still blank at
   `docs/planning/2026-11-07_21-30_owner-sitting-pack.md:47`.
7. **The go-idempotency and erraudit filings' actual posting** —
   drafts committed; posting awaits the owner tick (deliberately, see
   a-14).

## d) TOTALLY FUCKED UP

1. **The pipeline-exit-code lie, repeated from the documented trap.**
   The android verify-gate run captured `GATE_RC=$?` after a
   `timeout … | tail -30` pipeline — that is tail's exit code, not the
   gate's; the gate had actually FAILED on step 6 (dprint) while my
   captured RC read 0. The AGENTS-documented rule is explicit: capture
   every gate rc to a temp file because this tool shell has no
   PIPESTATUS — and I had even cited that rule earlier in the session
   (the explicit-rc ritual). I caught the real verdict because the
   dprint output was visibly red, not because my rc discipline worked.
   Cost: the gate verdict was unknown for one round trip.
2. **The hub lint reproduction ran one level too high — the exact
   trap the script's own comment documents.** `lint-sources.sh`
   carries a 2026-10-19 paragraph: scanning the export dir instead of
   the resource root breaks owner-ref resolution and produces exactly
   the false "owner does not exist" errors I then "reproduced". I
   concluded webphone's export was lint-dirty from a bad invocation;
   re-reading the script corrected it (the export is clean at the
   right root). Cost: one false failure theory + a round trip.
3. **Misattributed the hub failure for three cycles.** The ci-log
   "first failure" banner named the webphone governance lint; the
   REAL error (`current_contract: unbound variable`) sat 1417 built
   pages later at the very end of the log. I chased the lint,
   reproduced it wrong (d-2), probed the wrong sha — all before
   grepping the log tail past the banner. The lesson now obvious:
   the banner names a phase, the verdict needs the last error line.
4. **My ringhook unittest had a real bug** — the unknown-path probe
   used the helper that posts to `/ring` (expecting 404 from a
   different path), so it asserted against a malformed-body 400. The
   shim was right; the test was wrong. Fixed before landing (7/7
   green), but it shipped one red iteration.
5. **The art-dupl adjudication nearly blessed an unseen group.** I
   adjudicated the five groups the CI log showed, but the re-record
   counted +6 — the sixth (the journal-vs-legacy ThreadFlag struct
   twin) existed only in the baseline set-diff. I caught it by
   counting, then enumerated the diff; had I trusted the truncated
   log, one unreviewed group would have been blessed silently. The
   save was real and late.
6. **The first hub dispatch called the wrong workflow** — the dispatch
   script defaulted to `stale-sources.yml`; passing build.yml as a
   PATH form 404'd (Forgejo wants the bare filename). Two dispatch
   attempts before the build actually ran.
7. **The crm vendorHash probe used `--rebuild` on a never-built
   derivation** — nix refuses ("cannot be substituted"), one wasted
   build cycle before the plain build that prints `got:`.
8. **Two edit-tool staleness round trips** — the daemon touched files
   between my reads and edits (mod-time only, twice), and my own
   scripted rewrite invalidated the tracker once. Each cost a re-read
   cycle; no damage, but the re-View-before-Edit reflex was applied
   late in a session where the daemon is adversarial by design.

## e) WHAT WE SHOULD IMPROVE

1. **Capture gate exit codes to temp files as the FIRST reflex**, not
   a post-mortem correction — the documented rule exists precisely
   because `| tail` and pipes lie (d-1). Any command whose verdict
   matters gets `rc` captured before display.
2. **Read the reproducing script's own comments before reproducing** —
   lint-sources.sh documented the exact false-reproduction shape I
   performed (d-2). Cost of reading: 30 seconds; cost of skipping: a
   false failure theory.
3. **Phase banners name WHERE, not WHAT** — always pull the last error
   line of the step log before theorizing (d-3). The ci-log first-
   failure marker is a heuristic, not a verdict.
4. **Set-diff machine-readable state over trusting human summaries** —
   the truncated CI log under-reported the new clone groups by one
   (d-5); the baseline JSON set-diff is the ground truth. Same
   reflex now proven for baselines, goldens and inventories.
5. **Dispatch/workflow-name APIs want bare filenames** — path-form
   identifiers 404 on Forgejo dispatch (d-6); record the working form
   beside the script usage.
6. **When a failing output was truncated, re-pull the FULL log before
   the second theory** (d-3 and d-5 share this root).
7. **Test helpers deserve a path audit**: a helper that hardcodes the
   happy path (`self.post` → /ring) makes negative-path probes lie
   (d-4). Next island/Go test file: assert the URL inside the stub,
   not just the status.

## f) NEXT TASKS (ledger — harvest inventory, not commitments)

<!-- check-rows:ledger -->

| #  | Task                                                                                                                                                                                                                              | Impact | Effort | Category |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 1  | Heal the stack `telephony-voicemail` quota-exceeded assertion (the sibling session-2's named-open red) — unblocks the stack CI green                                                                                              | High   | M      | stack    |
| 2  | Earn the SECOND consecutive green stack CI run after 1 lands — the Leg 3 twice-green acceptance, then record both run ids in the handover pack                                                                                    | High   | S      | stack    |
| 3  | Restore android Actions billing (owner, Billing & plans) — every push since 11-08 ~21:25 UTC fails in seconds with jobs never started                                                                                             | High   | S      | android  |
| 4  | Post-billing android CI green on main — first real CI verdict on the merged AGP 9.5.1 + the dprint fix                                                                                                                            | Med    | S      | android  |
| 5  | Publish the go-idempotency contract-clock filing on the owner tick — draft committed, banner unchecked, gates+checks clean                                                                                                        | Med    | S      | filings  |
| 6  | Publish the erraudit tree same-name-dedupe filing on the owner tick — draft committed with the live reproduction                                                                                                                  | Med    | S      | filings  |
| 7  | Fix the erraudit tree display itself (own tool): package-qualified rows for same-name sentinels + a deterministic pick — possibly reaches go-finding's MergeIter dedup                                                            | Med    | M      | erraudit |
| 8  | Clone the transport-isolation fix into pbx, asr, matrix, paperless (the crm Clone() pattern) — kills the latent parallel-test flake class at those seams; then the 5× `-race` sweep per package                                   | Med    | M      | webphone |
| 9  | Hub runner durable fix: ExecStartPre provisioning (or static uid / group-write) for the build-state roots so the next DynamicUser renumber cannot red every build                                                                 | Med    | M      | hub      |
| 10 | Hub ALERT_WEBHOOK_URL decision (the failure log says "channel decision pending: ntfy / email / Forgejo webhook") — pick the channel, set the secret so failures alert                                                             | Med    | S      | hub      |
| 11 | Verify catalog.home.lan actually serves run #240's publish (the hourly sync half) and the freshness WARN cleared                                                                                                                  | Med    | S      | hub      |
| 12 | P4-12 standing-watch sweep: the erraudit tier-1+2 monthly re-measure is OVERDUE (calendar date was 11-05; last 10-05) + the named triggers checked now                                                                            | Med    | M      | watches  |
| 13 | Ringhook end-to-end ring proof: a push device registered + the shim enabled (VM or live host) + an inbound call rings the closed app — the seam's acceptance bar                                                                  | High   | M      | stack    |
| 14 | The ring-group boundary ruling (owner §g-1 of the 21-48 report): hook lane on every owned-extension ring vs inbound-DID only — decides whether ring groups get arms                                                               | Med    | S      | stack    |
| 15 | Consider a ringhook health leg joining telephony-health (the shim exposes /ringhook/health; operatorWatchedUnits lacks it)                                                                                                        | Low    | S      | stack    |
| 16 | The sitting pack D-slots (D1–D18) — the owner answers gate D10's cap posture, D12's docs-gates filing, D17's posting authorization                                                                                                | High   | S      | owner    |
| 17 | The 2.10.0 release-content decision (sibling 22-46 §g-1): cut from the green tree as-is vs hold for A35's importer                                                                                                                | High   | S      | owner    |
| 18 | The metrics-count decision (sibling 22-46 §g-2): messages/threads totals during the seam window — re-point vs sum vs frozen                                                                                                       | Med    | S      | owner    |
| 19 | Per-phase release cadence D4 amendment (sibling 22-46 §g-3) — decides three phase-gate release items in their plan                                                                                                                | Med    | S      | owner    |
| 20 | Sibling Phase-5 continuation: A35 importer, A37 parity legs, A38 gates (incl. the stack browser E2E — already harvested) — their train, this row is the watch                                                                     | High   | L      | webphone |
| 21 | Docs-gates four sweep defects: the D12 draft exists (2026-11-08) — owner review tick, then post; this session re-confirmed the DF-brittleness live (error-contract growth flipped a pin)                                          | Med    | S      | docs     |
| 22 | P4-13 parked-decisions re-confirmation read-through (one pass, no action; re-run next quarter)                                                                                                                                    | Low    | S      | docs     |
| 23 | P4-14 coverage floor implementation — sitting-gated on C5 GO                                                                                                                                                                      | Low    | L      | webphone |
| 24 | P4-11 release announcements drafts refresh (v2.9.x) — owner picks channels + disclosure posture                                                                                                                                   | Low    | S      | owner    |
| 25 | P4-8 native first-device train (needs the device): REGISTER/INVITE/BYE/DTMF vs the live WSS, incoming INVITE, CallStyle, origin picker, sessionIdentity assert — README honesty markers flip per leg                              | High   | L      | android  |
| 26 | Verify the other three Dependabot PRs the plan named (firebase-messaging, okhttp, gradle-wrapper) — only AGP #7 was open; confirm the rest merged or re-opened at the next Dependabot cycle                                       | Low    | S      | android  |
| 27 | Android M050 emulator re-run with the hardened harness (existing TODO top row)                                                                                                                                                    | Med    | 30min  | android  |
| 28 | Android emulator smoke/train script hardening: timeout-bound boot + install, mandatory log output (existing TODO)                                                                                                                 | Med    | 20min  | android  |
| 29 | webphone-android consumes the OpenAPI/asyncapi specs in bootstrap (TODO wire-contracts row)                                                                                                                                       | Med    | M      | android  |
| 30 | The dashboard unsafe-eval→nonce retirement (TODO datastar residue): upstream lane, library release, webphone re-pin of the CSP assertion                                                                                          | Low    | M      | webphone |
| 31 | Phase-4 KVM release preflight (TODO harvest from 21-48)                                                                                                                                                                           | Med    | M      | webphone |
| 32 | Stack residues: the dashboard-on VM test leg + the styled-404 shape rule in the stack's browser E2E (TODO cross-repo row)                                                                                                         | Low    | M      | stack    |
| 33 | The flush-announce island spec (TODO harvest from 21-48)                                                                                                                                                                          | Low    | S      | webphone |
| 34 | The type-safety roadmap open items: pbx Transport defined type + ParseTransport, sso.Finish value object, FaxJob.DocumentPath as domain.RelPath, asr/matrix wiring structs, machineSurface bool-last                              | Low    | M      | webphone |
| 35 | catalog v4.9.0→v4.10.0 bump decision (owner-gated TODO wire-contracts row; 213 commits)                                                                                                                                           | Low    | S      | owner    |
| 36 | The ssetest WithLastEventID ride (trigger-gated: a Last-Event-ID replay surface must exist first)                                                                                                                                 | Low    | S      | webphone |
| 37 | Issue #236's race ledger row: close it against a green successor run when the core-swap train settles                                                                                                                             | Low    | S      | webphone |
| 38 | The daemon-vs-Go-edit race mitigation: decide between atomic commit-at-each-green-point and stash-on-pause as the standing rule (two daemon ate-my-WIP incidents across sessions now)                                             | Med    | S      | process  |
| 39 | Fleet-libs reopen-trigger sweep (quarterly cadence; last sweep 11-04) — one read of the ONE table                                                                                                                                 | Low    | S      | webphone |
| 40 | Multi-actor HITL escalation tiers (TODO 2026-10-28 row) — unstarted since landing                                                                                                                                                 | Low    | M      | webphone |
| 41 | Error-surface crash-containment tails (TODO 2026-10-27 row: host legs + evidence-gated tails)                                                                                                                                     | Low    | M      | webphone |
| 42 | Dialer-adoption #6–#8 tails (TODO 2026-10-27 row)                                                                                                                                                                                 | Low    | M      | webphone |
| 43 | Add the ringhook lane to the stack's deploy docs (the option table + the secret-slot note) — the module is shipped but undocumented operator-side                                                                                 | Med    | S      | stack    |
| 44 | The webphone NixOS module gains a typed `ringhook` front (enable/port/webphoneUrl/webhookSecretFile threaded from the stack level) — the stack module owns the shim; the app module may want the typed mirror for non-stack hosts | Low    | M      | stack    |
| 45 | Confirm the crm repo's OWN gates on the vendorHash repair commit (the hub build proves the build; crm's treefmt/buildflow verdicts were not read)                                                                                 | Low    | S      | crm      |
| 46 | Hub stale-sources.yml: confirm it goes green on its next scheduled run now that the mount is repaired (its two failures today were the mount class)                                                                               | Low    | S      | hub      |
| 47 | The operator command sheet gains the ringhook enable recipe (config keys + secret slot + what to expect in the dialplan)                                                                                                          | Low    | S      | stack    |
| 48 | TODO_LIST freshness: the 21-48 report's three harvested rows plus this report's new rows (8, 9, 10, 11, 43–45) — one consolidated read at next sweep so no row is homed twice                                                     | Low    | S      | docs     |
| 49 | The webphone export's governance-lint hygiene stays green at the RESOURCE root — add the one-level-too-high trap as a comment in the webphone export docs (the hub script carries it; the source side does not)                   | Low    | S      | docs     |
| 50 | The build-state root ENOSPC history (hub runs 11/63, cited by the mount error) — a row in the hub TODO exists for the uid class; the inode-cap half of that old lesson has no tracked owner                                       | Low    | S      | hub      |

## g) QUESTIONS (cannot be figured out from the tree)

1. **The stack voicemail quota red** (f-1): is the sibling stack
   session resuming to fix its own named-open quota-model correction,
   or do you want THIS session's line to take it now? It is the one
   blocker on the Leg-3 twice-green acceptance, and both sessions
   touching the voicemail/quota surface at once is the split-brain I
   want your call on before acting.
2. **The two committed filings** (f-5/f-6): the go-idempotency clock
   ask and the erraudit tree dedupe are drafted, gate-verified and
   check-clean, but the sitting pack's D-slots are unanswered — do I
   post them now on your review, or hold until you answer D17 (and
   D12's sibling draft) in the sitting?
3. **Android verification posture while billing is down** (f-3/f-4):
   local verify-gate substituting for CI was my call this session —
   is that the standing rule until you restore Actions billing, or
   should merges wait for real CI even at the cost of the
   Dependabot queue going stale?

---

**Session posture**: webphone main green twice (42531121332,
42534218876); stack pushed with legs verified (acceptance blocked on
the sibling's named voicemail red); android merged (billing-documented);
hub green and published. The session WAITS here per the owner's
instruction.
