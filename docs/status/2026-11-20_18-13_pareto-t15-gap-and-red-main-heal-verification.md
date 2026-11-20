# Pareto T15 gap + red-main heal verification — status

> Snapshot-rot header: claims verified as of 2026-11-20 18:13 CET.
> Scope: the delta since this session's 12:20 report
> (`2026-11-20_12-20_pareto-trains-t10-t15-midflight-status.md` — the
> full T01–T27 context, the a–g breakdown of the T10–T14 trains, and
> the three standing owner questions live THERE; this report covers
> what happened after, and what did not). CI claim verified via
> `scripts/ci-verdict.sh origin/main` at snapshot time.

## The ask

Owner: full status report, then WAIT. (Before that: the standing
"whole list, do not stop" directive — honored until this report.)

## a) FULLY DONE (this window)

- **Red-main heal VERIFIED end-to-end**: the art-dupl gate repair
  (baseline 109 → 114 with the two adjudicated intentional-group
  classes — the review-prescribed FCM/sso single-flight skeleton and
  the one-shot-companion constructor mirror; registry rows appended;
  CI-pinned v0.7.2, exact gate flags) converged: the bounded CI poll
  answered **GREEN on `30057a6e`** — main healed from the three-run
  red the unverified T12/T13/T14 pushes had left. Lesson recorded in
  the 12:20 report §d/§e: check CI after EVERY push.
- **The 12:20 report landed green** (daemon split it across
  `32b3e7c5` + `30057a6e`; cadence healed by archiving the
  gates-passing line-ownership snapshot; docs-gates battery PASS
  after the dprint repair).
- **T15's already-landed halves re-confirmed present at HEAD** (the
  daemon had swept them in `c016d918`/`32b3e7c5`): the FCM
  single-flight rework (`fetchAccessToken` cache-check-under-lock →
  joiner channel → leader-only `exchangeToken`), the `blob.Save`
  brand-before-write reorder, the Seq-as-retention-clock comment at
  `SweepBefore`, the dead zero-predicate and fax-trust-comment items
  the earlier verification session had already done.

## b) PARTIALLY DONE

- **T15's two missing tests — STILL missing** (re-verified by grep
  at 18:13): (1) the FCM single-flight concurrency test (N concurrent
  callers share ONE grant exchange; a cached second wave makes no
  new grant), (2) the blob malformed-extension edge test (brand
  refusal leaves NO orphaned file behind — the reorder's whole
  point). I had read the harness (`blob.New(t.TempDir…)`,
  `ParseRelPath`'s `..`-segment refusal) but written nothing when
  the WAIT arrived. This is the session's open seam: code landed,
  its pin tests did not.

## c) NOT STARTED (unchanged from the 12:20 report)

T16–T27 in plan order; T01–T03 deploy cycle (owner question g-1
pending); the final report-back rendering both plan tables. The
naming train's remainder (VoicemailItem split brain, authmap.Map →
Ownership, Shell ×2) and the unplanned vcard group-prefix fix (L5)
also remain.

## d) TOTALLY FUCKED UP

- **A six-hour dead window between the 12:20 WAIT and now**: the CI
  poll finished GREEN at ~12:35, the session then sat idle instead
  of either finishing T15's two small tests (minutes of work) or
  reporting — the todo list's "T15 missing tests" item was the
  cheapest open item on the board and outlived the whole window.
- No NEW self-inflicted damage this window (the 12:20 report §d owns
  the three-unverified-pushes red and the daemon-swept untested
  code; both healed).

## e) WHAT WE SHOULD IMPROVE

- **Never end a window holding a minutes-scale open item** — finish
  it or explicitly park it with an owner note; the current shape
  (code landed, pin tests absent) is exactly the "unverified claim"
  class this repo's own gates exist to catch.
- **The concurrent-session ledger**: another session's trains landed
  between my reports (`792d2949`, `e8d8c6ec` — a 20-skill battery +
  cadence drain). My reports should name that boundary explicitly
  (this one does) so outcomes are never double-claimed.
- Cadence arithmetic: `ls docs/status/*.md` counts INDEX.md — the
  honest unarchived count is files-minus-index; the battery's own
  number is the authority, not `ls | wc -l`.

## f) NEXT (up to 50; plan order unless noted)

1. **T15 closeout** (first): write + run the FCM single-flight test
   and the blob malformed-ext test; push-package suites
   (push/blob/cqrs); one narrative commit.
2. T16 cadence pre-drain automation (sweep that pre-computes the
   gates-passing archive set; WARN lane in docs-gates).
3. T17 stack window: drift audit + coverage arms (stack repo,
   mirrored-module eval check).
4. T18 ringhook operator docs + typed `ringhook.*` mirror +
   module-check case.
5. T19 android `lines[]` adoption (own repo, Gradle suite,
   WireContractTest).
6. T20 android wire-fixture triggers (execute only FIRED ones).
7. T21 SSO-only stack fixture (boot, login, whoami identity).
8. T22 fleet health batch (pma push heuristic, hub runner durable
   fix, ALERT_WEBHOOK_URL decision).
9. T23 CUE ruling record (AGENTS/README one-liner + pointer).
10. T24 notes-in-search spike (corpus boundary, query shape, FTS vs
    LIKE) → docs/planning note + row re-scope.
11. T25 hardening-parity pin per the 00:51 spec.
12. T26 BuildFlow bandit-provider issue (read draft, verify, file).
13. T27 flightrecorderhealth demand check → adopt or record gate.
14. T01 stack redeploy (needs g-1).
15. T02 prod authenticated capture + ring-guard positive path.
16. T03 night-run stability pair.
17. Final report-back: both plan tables (§2 T01–T27, §3 m01–m27)
    with outcomes.
18. Naming train remainder (VoicemailItem, authmap.Map → Ownership,
    Shell ×2).
19. vcard L5 group-prefix (`item1.TEL`) fix + Apple-export fixture.
20. Local dedup-check wrapper script (the pinned-detector dance
    cost three attempts; see 12:20 §e).
    21–50. The gated remainder as the owner unlocks: R21 register tick,
    R22/R23 lens lanes, v2.15.0 release order, ThreadGate Q1–Q5, the
    §5 DROP window ritual (dump + confirmations + migration
    citation), the four oldest reports' real annotate pass (kills
    the recurring cadence pressure at the root), the
    reopen-trigger registry sweep, loginpage/ExtraSection
    follow-ups, the cqrs-lint keep-visible list's quarterly audit,
    monthly tier-1/2 error-family re-measure (due 2026-12-08).

## g) Questions for the owner (unchanged + one new)

1. **Deploy authority (T01–T03)** — same as 12:20 g-1: does "the
   WHOLE list" authorize the production redeploy (stack relock,
   `nixos-rebuild switch` on the PBX host, `/version` verify,
   click-play), or does the standing cutover-authority question
   still gate it? Includes the v2.15.0 release-order call.
2. **Android window (T19/T20)** — same as 12:20 g-3: in scope for
   this train or a separate dedicated window?
3. **NEW — the blocking workflow-e2e posture**: the browser job is
   now blocking with 40 checks; today's six-hour gap and the
   earlier three-red stretch both argue the gate is doing its job,
   but if its flake rate grows with journey count, the owner call
   (keep blocking vs. N-green probation) cannot be self-answered.
