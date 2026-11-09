# Status: the legs session through its report and the Phase-6 tail — second snapshot

Snapshot-rot header: claims verified as of **2026-11-09 03:38 CET**
(local git state read directly; CI verdicts timestamped inline). This
report supersedes the posture claims of the 00-49 report (same
session; that report's §a work all stands, but its closing "main
green" statement is STALE — see b-1).

Scope: this session's run only (resumed from the 21-48 report's
STATUS-THEN-WAIT per the owner's execute mandate; the first snapshot
is `docs/status/2026-11-09_00-49_cross-repo-legs-hub-green-android-merged-status.md`
— its §a–g remain the session's substance; this snapshot adds the
report-push tail and corrects the rot).

---

## a) FULLY DONE (all of the 00-49 report's §a stands; verified there)

1. Webphone main healed from five red legs (registry, art-dupl, docs
   battery, crm DefaultTransport flake, treefmt canonicalization) —
   runs 42528105845 / 42531121332 / 42534218876 all-success on their
   shas.
2. Stack Leg 2 (ringhook) shipped + verified (7-spec unittest,
   ringhookCheck eval arm, dialplan-sim/format/changelog gates),
   pushed `89eca55..fccc1f4`; CI legs green on the pushed tip.
3. P4-9 android PR #7 merged via the repo's full local verify-gate;
   main's pre-existing dprint drift fixed; billing blocker documented.
4. P4-10 catalog hub run #240 GREEN end-to-end and published, through
   three infra repairs (runner uid renumber, crm vendorHash `37a1957`,
   the workflow step-scoping fix `7551d4a`).
5. P4-7 filings: branching-flow pair already filed+closed upstream;
   go-idempotency clock + erraudit tree drafts committed
   gate-verified and check-clean; file-size-check already shipped
   upstream (#91).
6. The 00-49 status report itself: written, battery-green at write
   time, index regenerated (61 reports), harvested §f, pushed
   (`6f0c746a`).
7. The post-report format-gate heal: CI's rc=20 at `6f0c746a`
   diagnosed (see d-1), the daemon's `ecbfb6ae` canonical bytes found
   unpushed, pushed under the manual-push bar; the tip's fast-flake
   leg then ran GREEN (run 42538856915: format ✓, go test -race ✓,
   build ✓, smoke ✓).

## b) PARTIALLY DONE

1. **"Webphone main green" is stale — main is RED again at the tip,
   and the reds are the sibling's, not mine.** After their SUPERB
   Phase 6 landed (A39–A40 fax journal `b759c603`, the 2.11.0
   changelog + version bump `eefa68bb`) and their P7 closeout docs
   (`06deaaec`, `0e49e939`), CI ran: run 42538856915 red on the
   **art-dupl leg** (new clone groups from the freshly landed fax
   journal code — the same gate my session healed for the threads
   aggregate), and run 42539029787 (their final status report push)
   concluded FAILURE on shas since rewritten in my local log
   (`06deaaec`/`0e49e939` — their closeout was amended). Their own
   commit message names the work in flight ("drain the annotate
   regression (rule-F span) — SUPERB P7 closeout"). NOT touched by
   this session per the concurrent-session doctrine — but the
   00-49 report's posture line rot-ted within the hour, which is
   exactly what the snapshot-rot header exists to flag.
2. **Stack Leg 3 twice-green acceptance** — blocked on the sibling's
   named-open voicemail quota red (unchanged from 00-49).
3. **The transport-isolation Clone() class** — crm fixed; pbx/asr/
   matrix/paperless harvested (TODO) but unfixed.
4. **The erraudit tree fix** — drafted, unfixed.
5. **Hub freshness loop** — #240 published; the serving side presumed
   on cadence, unverified.
6. **Android CI truth** — merged main still has no CI verdict
   (billing).

## c) NOT STARTED

Unchanged from the 00-49 report §c: P4-8 (device-gated), P4-11
(owner), P4-12 (the erraudit monthly re-measure overdue since 11-05),
P4-13, P4-14 (sitting-gated), the sitting pack's D-slots, and the two
filings' posting (owner tick).

## d) TOTALLY FUCKED UP

1. **I battery-verified the working tree and pushed a different tree.**
   The report push's format rc=20: my local battery ran while the
   sibling's release-prep docs were in my WORKING tree — the battery's
   own fmt step formatted them IN PLACE (dirty, unstaged), the battery
   passed, and my commit carried only my three files. CI checked out
   the COMMIT, whose bytes were still their unformatted originals →
   red. The divergence class: battery-on-working-tree ≠
   battery-on-commit when the fmt step's fixes are left unstaged. The
   correct ritual: after the battery, `git add` EVERYTHING the fmt
   touched (or commit first, then battery the clean tree), THEN push.
   The daemon's `ecbfb6ae` sweep carried the canonical bytes and my
   manual push landed them — but the red CI run between was mine to
   prevent.
2. **The 00-49 report's closing posture went stale in under an hour**
   ("main green" → sibling's Phase 6 landed → art-dupl red). The
   snapshot-rot header warns about this; the honest correction is THIS
   report's b-1 rather than a silent edit of a pushed report — but the
   sharper version of the lesson: a posture claim should carry its
   commit sha so the rot is checkable ("green at `ecbfb6ae`" ages
   honestly; "green" does not).
3. **Carried from the 00-49 report §d (all still true)**: the
   pipeline-rc lie (tail's exit code captured as the gate verdict),
   the one-level-too-high lint reproduction (the trap documented in
   the script's own comment), the three-cycle misattribution of the
   hub failure (phase banner vs last error line), the ringhook
   test-helper path bug, the near-blessed sixth clone group (saved by
   the set-diff, late), the double dispatch miss, the `--rebuild`
   probe on a never-built derivation, the daemon-adversarial edit
   staleness round trips.

## e) WHAT WE SHOULD IMPROVE

1. **Battery-on-commit**: the docs-health ritual gains a tail step —
   after committing, confirm `git status` is clean so the battery's
   fmt output is INSIDE the commit, and quote the commit sha in any
   posture claim. (This session produced the counter-example; the rule
   should outlive it.)
2. **Push-time inventory**: before pushing, read
   `git log origin/main..HEAD` AND `git log HEAD..origin/main` (post-
   fetch) so gap commits — the sibling's or the daemon's — are seen
   before their state rides my push into CI.
3. **Carried from 00-49 §e**: temp-file rc capture as the first
   reflex; read the reproducing script's comments first; the phase
   banner names WHERE not WHAT; set-diff machine state over human
   summaries; bare filenames for dispatch APIs; stubs assert the URL
   not just the status.

## f) NEXT TASKS (ledger — harvest inventory, not commitments)

<!-- check-rows:ledger -->

| #  | Task                                                                                                                                                                                                  | Impact | Effort | Category |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 1  | Re-record the art-dupl baseline for the sibling's Phase-6 fax-journal groups (adjudicate-before-write, same as the threads-aggregate repair) — OR their train does it; the red blocks every later run | High   | S      | webphone |
| 2  | Read run 42539029787's failing legs (their final-report push) — art-dupl was #238856915's red; the second red is unnamed in this session's observations                                               | High   | S      | webphone |
| 3  | Heal the stack `telephony-voicemail` quota assertion (sibling's named-open red) — unblocks Leg-3 twice-green                                                                                          | High   | M      | stack    |
| 4  | Earn the second consecutive green stack CI run (Leg 3 acceptance), record both run ids in the handover pack                                                                                           | High   | S      | stack    |
| 5  | Restore android Actions billing (owner) — pushes since 11-08 ~21:25 UTC never start jobs                                                                                                              | High   | S      | android  |
| 6  | Post-billing android CI green on the merged main (AGP 9.5.1 + dprint fix)                                                                                                                             | Med    | S      | android  |
| 7  | Publish the go-idempotency clock filing on the owner tick (draft committed, gates+checks clean)                                                                                                       | Med    | S      | filings  |
| 8  | Publish the erraudit tree same-name-dedupe filing on the owner tick (draft + live reproduction committed)                                                                                             | Med    | S      | filings  |
| 9  | Fix erraudit tree itself: package-qualified rows + deterministic pick (own tool; possibly go-finding's MergeIter)                                                                                     | Med    | M      | erraudit |
| 10 | Clone the transport-isolation fix into pbx, asr, matrix, paperless (TODO-harvested 2026-11-09)                                                                                                        | Med    | M      | webphone |
| 11 | Hub runner durable fix: ExecStartPre provisioning / static uid for the build-state roots (in-place chown done; durable half in the hub TODO)                                                          | Med    | M      | hub      |
| 12 | Hub ALERT_WEBHOOK_URL decision (channel pending: ntfy / email / Forgejo webhook)                                                                                                                      | Med    | S      | hub      |
| 13 | Verify catalog.home.lan serves run #240's publish (hourly sync) and the freshness WARN cleared                                                                                                        | Med    | S      | hub      |
| 14 | P4-12 standing-watch sweep: erraudit tier-1+2 monthly re-measure OVERDUE (was due 11-05) + named triggers checked                                                                                     | Med    | M      | watches  |
| 15 | Ringhook end-to-end ring proof (push device + shim enabled + inbound call rings the closed app)                                                                                                       | High   | M      | stack    |
| 16 | The ring-group boundary ruling (owner §g-1 of the 21-48 report)                                                                                                                                       | Med    | S      | owner    |
| 17 | Ringhook health leg joining telephony-health + the deploy-docs enable recipe + the operator command sheet row + a typed mirror in webphone's own module (the four doc/ops faces of the shipped lane)  | Med    | S      | stack    |
| 18 | The sitting pack D-slots D1–D18 (gates D10 posture, D12 docs-gates filing, D17 posting)                                                                                                               | High   | S      | owner    |
| 19 | The 2.10.0/2.11.0 release-content decision (sibling §g-1) — their 2.11.0 bump landed; the cut-vs-hold question needs their A35 state as input                                                         | High   | S      | owner    |
| 20 | The metrics-count decision (sibling §g-2)                                                                                                                                                             | Med    | S      | owner    |
| 21 | The D4 per-phase release cadence amendment (sibling §g-3)                                                                                                                                             | Med    | S      | owner    |
| 22 | The sibling's SUPERB complete closeout (their final report pushed at 00:13 UTC): their §f will re-open harvest items — next sweep reads THEIR report the way this session read theirs at 22:46        | Med    | M      | docs     |
| 23 | The docs-gates four sweep defects filing (D12 draft exists; owner tick owed) — the DF-brittleness re-confirmed live again today (error-contract growth)                                               | Med    | S      | docs     |
| 24 | The battery-on-commit rule (d-1): AGENTS docs-gates entry gains the tail step + the sha-quoted posture rule                                                                                           | Med    | S      | process  |
| 25 | Push-time inventory reflex (e-2): gap-commit reading before push — AGENTS candidate beside the manual-push bar                                                                                        | Med    | S      | process  |
| 26 | P4-13 parked-decisions re-confirmation read-through                                                                                                                                                   | Low    | S      | docs     |
| 27 | P4-14 coverage floor — sitting-gated (C5 GO)                                                                                                                                                          | Low    | L      | webphone |
| 28 | P4-11 release announcements drafts refresh (owner)                                                                                                                                                    | Low    | S      | owner    |
| 29 | P4-8 native first-device train (device-gated)                                                                                                                                                         | High   | L      | android  |
| 30 | Verify the other three Dependabot PRs the plan named (only AGP #7 was open) at the next Dependabot cycle                                                                                              | Low    | S      | android  |
| 31 | Android M050 emulator re-run with the hardened harness (existing TODO top row)                                                                                                                        | Med    | 30min  | android  |
| 32 | Android emulator script hardening: timeouts + mandatory log output (existing TODO)                                                                                                                    | Med    | 20min  | android  |
| 33 | webphone-android consumes the OpenAPI/asyncapi specs (TODO wire-contracts)                                                                                                                            | Med    | M      | android  |
| 34 | Dashboard unsafe-eval→nonce retirement (TODO datastar residue: upstream lane, release, re-pin)                                                                                                        | Low    | M      | webphone |
| 35 | Phase-4 KVM release preflight (TODO harvest)                                                                                                                                                          | Med    | M      | webphone |
| 36 | Stack residues: dashboard-on VM test + styled-404 E2E shape (TODO cross-repo row)                                                                                                                     | Low    | M      | stack    |
| 37 | The flush-announce island spec (TODO harvest)                                                                                                                                                         | Low    | S      | webphone |
| 38 | The type-safety roadmap items (Transport/ParseTransport, sso.Finish, FaxJob.RelPath, asr/matrix wiring structs, machineSurface bool-last)                                                             | Low    | M      | webphone |
| 39 | catalog v4.9.0→v4.10.0 bump decision (owner-gated)                                                                                                                                                    | Low    | S      | owner    |
| 40 | ssetest WithLastEventID ride (trigger-gated)                                                                                                                                                          | Low    | S      | webphone |
| 41 | Issue #236's race ledger row: close against a green successor when the core-swap train settles                                                                                                        | Low    | S      | webphone |
| 42 | Fleet-libs reopen-trigger sweep (quarterly; last 11-04)                                                                                                                                               | Low    | S      | webphone |
| 43 | Multi-actor HITL escalation tiers (TODO, unstarted)                                                                                                                                                   | Low    | M      | webphone |
| 44 | Error-surface crash-containment tails (TODO, unstarted)                                                                                                                                               | Low    | M      | webphone |
| 45 | Dialer-adoption #6–#8 tails (TODO, unstarted)                                                                                                                                                         | Low    | M      | webphone |
| 46 | Confirm crm's own gates on the vendorHash repair (hub proves the build; crm's treefmt/buildflow verdicts unread)                                                                                      | Low    | S      | crm      |
| 47 | Hub stale-sources.yml green on its next scheduled run (its failures today were the mount class, now repaired)                                                                                         | Low    | S      | hub      |
| 48 | TODO_LIST consolidated read so no row is homed twice (00-49 + this report's harvests)                                                                                                                 | Low    | S      | docs     |
| 49 | The webphone export docs gain the one-level-too-high lint-trap note (the hub script carries it; the source side does not)                                                                             | Low    | S      | docs     |
| 50 | The build-state root inode-cap history (hub runs 11/63) has no tracked owner beside the uid-class row                                                                                                 | Low    | S      | hub      |

## g) QUESTIONS (cannot be figured out from the tree)

1. **The stack voicemail quota red** (f-3): does the sibling stack
   session resume to fix its own named-open quota-model correction,
   or does THIS session's line take it? It blocks the Leg-3
   twice-green acceptance; both sessions editing the voicemail/quota
   surface at once is the split-brain I want your call on first.
2. **The two committed filings** (f-7/f-8): the go-idempotency clock
   ask and the erraudit tree dedupe are drafted, gate-verified and
   check-clean — post them now on your review, or hold until the
   sitting answers D17 (and the D12 sibling draft)?
3. **Art-dupl ownership on the Phase-6 red** (f-1/f-2): the fax-journal
   clone groups are the sibling's freshly landed code mid-closeout —
   do I re-record the baseline adjudicated (as I did for the threads
   aggregate), or does their P7 closeout own it? Their commit messages
   say they were draining annotate when CI went red on art-dupl, so
   the trains may already be crossed.

---

**Session posture**: the session's own work stands verified (legs,
hub, android, filings, the 00-49 report); the tip's current red is the
sibling's Phase-6 closeout surface, correctly untouched by this
session. The session WAITS here per the owner's instruction.
