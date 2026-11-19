# Session self-audit: pbx-lens plan review + SUPERB plan closeout

> Claims verified as of **2026-11-19 13:37 CET**. Push/CI statements rot
> fastest; this file timestamps them.

**Scope.** This report covers ONLY this session's run (the merge
analysis → status report → SUPERB plan → plan review → this audit) and
what was noticed along the way. Verified this turn with commands:
CI verdict for `9331508c` via `scripts/ci-verdict.sh` (rc=0, GREEN),
commit ancestry via `git merge-base --is-ancestor` (true), tree state
via `git status`/`git log`, lens link-route facts via `rg` against
pbx-lens source (last turn). NOT verified: the full docs-gates battery
against the tree that now includes this file (not run — report written
first); lens's auth-redirect deep-link carry-over (one grep was
inconclusive; left as question g3); the live-report ratchet count
post-write (58 `.md` under `docs/status/` incl. INDEX before this
file; pin is 65 — under, but the authoritative check is docs-gates,
not `ls`).

## What this session was

Three completed requests (merge analysis, a–g status report, SUPERB
plan) inherited from the session summary, one plan-review turn, and
this audit. No product code was changed by me at any point — the plan
was the deliverable, per instruction.

## a) FULLY DONE

1. **Merge / common-data-layer analysis** — verdict delivered with
   evidence: NO merge, NO shared layer; YES read-only deep link +
   upstream OIDC extraction + recorded ruling. The four decisive
   facts (no call aggregate in webphone; uuid-vs-Call-ID join-key
   mismatch — with the uuid already on the wire; opposite retention
   doctrines; appkit-vs-do as the only structural fork) were measured,
   not guessed.
2. **Status report** `2026-11-15_11-17_pbx-lens-merge-vs-common-data-layer-decision-analysis.md`
   — written, INDEX regenerated (`--check` rc=0), committed by the
   concurrent session as `375fe473`.
3. **SUPERB plan**
   `docs/planning/2026-11-15_15-06_SUPERB-pbx-lens-readonly-integration-and-decision-closeout.md`
   — 24 macro / 96 micro tasks, mermaid execution graph, non-goals,
   guardrails, evidence appendix, 5-question owner-gated register.
   Committed `9331508c`, pushed `f034ce81..9331508c`, remote sha
   verified equal, ancestry in HEAD confirmed.
4. **CI verdict for `9331508c`: GREEN** — verified THIS turn
   (`scripts/ci-verdict.sh`, rc=0, workflow-name-verified). The item
   that was pending at the last handoff is closed. It only took four
   days and one audit request to actually run the check (see d1).
5. **Plan review turn** — verified the plan's riskiest factual claim
   against lens SOURCE, not the plan's own prose: `SourceKind =
   "freeswitch-cdr"` (`pbx-lens/internal/calls/domain.go:53`), route
   `GET /calls/{kind}/{id}` (`internal/web/server.go:123`), detail
   resolves by the source-NATIVE id (`internal/web/queries.go:607-612`,
   param `nativeID`; recordings are uuid-keyed per the comment at
   server.go:572-575). The href shape `/calls/freeswitch-cdr/<uuid>`
   is confirmed correct. Delivered verdict (good plan, correctly
   prioritized 1%) + 5 concrete gaps + 3 questions.

## b) PARTIALLY DONE

1. **Plan review follow-through** — the 5 gaps were identified and
   reported but NOT folded into the plan doc (review mode; owner has
   the questions). The plan as committed still lacks: the dom-contract
   pin-site grep ritual, the tw.css/utility-class note, the templ CLI
   pin (`go run …/cmd/templ@v0.3.1100`, not devshell templ), the
   hybrid-History exclusion (island-local rows carry no uuid → no
   link), and the T18 "depends: all" vs graph-subset inconsistency.
2. **`usermgmt.db` backup coverage (T03)** — verified omitted from
   webphone's `package/backup-script.nix`; the stack/host-side
   coverage question is still open (that sweep is T03.1–T03.3,
   unexecuted).
3. **Review questions** — 2 of 3 are owner-gated by design; the
   third (lens auth deep-link carry-over) got one inconclusive grep,
   no answer.
4. **Ratchet posture after this file** — `docs/status/INDEX.md` was
   regenerated and `--check` passed, but the full docs-gates battery
   was not run against the tree containing this report.

## c) NOT STARTED

None of the plan's 24 macro tasks. Explicitly: **zero code changed**
for the link, the config seam, the backup fix, or the docs (T01–T24
all pending). The last-20% tier (T16–T24: OIDC sketch, upstream ask,
two-binary memo, vocabulary module, promotion notes, identity/retention
memo, Reports spike, revisit date) is intentionally owner-gated and
may never be started — that is by design, not neglect.

## d) TOTALLY FUCKED UP

1. **The CI verdict hung for four days on a one-command check.** At
   plan-push time I polled ~16 times over ~10 minutes, then yielded
   with "pending". The review turn had a free slot to run it and
   DIDN'T — I reviewed plan quality while my own pushed commit sat
   unverified, even though the handoff listed the check as next step
   #1. It was green all along. Wasted polls then, four days of
   ambiguity now resolved only because the owner demanded an audit.
2. **Unnormalized LOC figures in a decision doc.** The analysis's
   first answer quoted test+generated-inflated totals (88 759 vs the
   honest 34 492) and had to be corrected mid-session. Wrong numbers
   in the doc that justifies a NO-GO is the exact class of error the
   doc existed to prevent.
3. **`rg -rn` replace-flag trap hit again** during the lens
   verification — third strike on this host per the hook's own count.
   The hook caught it; my fingers apparently still think rg is grep.
4. **Review-turn verification was reactive, not proactive.** The
   freeswitch-cdr/nativeID check that de-risked the plan's core href
   took three tool calls — it should have been IN the plan's evidence
   appendix at writing time. The plan shipped asserting a route shape
   it had not verified against lens source.

## e) WHAT WE SHOULD IMPROVE

1. **Verify factual claims against source at PLAN-WRITING time.** Any
   wire shape, route literal, or field name that a plan depends on
   gets one rg against the owning repo before the plan is committed —
   evidence appendix or it didn't happen.
2. **Never yield a turn with my own pushed commit unverified.**
   `scripts/ci-verdict.sh <sha>` costs one call; pending is not done.
3. **Plan docs must embed the repo's served-markup rituals** — the
   three pin-site grep, the tw.css rebuild trigger, the templ CLI
   pin. All three were missing from T07/T08 and every one has a RED
   MAIN precedent in AGENTS/lessons.
4. **Polling discipline.** Sixteen inline polls is a background job
   or a single deferred check with an explicit handoff marker. (The
   handoff marker existed; the polls were still wasted.)
5. **Write `rg -n`, never `rg -rn`.** Same class as the jq/`$?`
   shell traps already in AGENTS — worth a line in the concurrent
   sessions/traps section if it strikes a fourth time.
6. **Make hybrid-surface exclusions explicit in verification
   sections** ("History is island-log + CDR; local rows get no link")
   so parity tests are written against the real shape.
7. **Fix the T18 dependency-word nit** when the plan is next touched
   ("all" → the graph's actual feed set).

## f) What to get done next (Pareto order; items 1–8 unblock everything)

1. Fold the 5 review gaps into the plan doc (one edit, one commit).
2. Owner ticks the 5-question register in the plan.
3. Answer review question g2 (affordance shape) before T07 renders.
4. T05 — decode CDR `uuid` into `pbx.CDR` + `views.CDREntry`; rewrite
   the two "no call uuid" comments; extend the decode test.
5. T06 — `lens.url` config seam (CRM-pattern: empty=off, absolute
   http(s), no userinfo/fragment, trailing-slash normalize; document
   browser-facing vs loopback).
6. T07 — History-row anchor (en+de i18n, `target="_blank"` +
   `rel="noopener noreferrer"`, off-state byte-parity test, escaping
   test with a crafted uuid, existing utility classes only).
7. T08 — server-side link builder wired through `historyPanel`;
   on/off parity + hostile-uuid tests; dom-contract decision +
   three-pin-site grep.
8. T09 — threat model + degraded-mode contract (lens down = dead
   link, never a page error; decide the non-https public-host stance).
9. T03 — sweep the stack (modules/, packages/, host docs) for a
   `usermgmt.db`/identity backup lane; record covered-vs-gap.
10. T04 — IF a gap: guarded `.backup` in both backup branches +
    presence/integrity VM assertions + restore drill.
11. T10 — smoke boot check (link present with lens URL, byte-parity
    without) + capture-surface decision.
12. T01 — one-line ruling in BOTH AGENTS files (webphone <377 lines,
    lens re-View first).
13. T02 — reopen triggers + revisit date into the consolidated
    registry; TODO row 1658 cross-check.
14. T11 — README seam paragraph, FEATURES row, AGENTS bullet,
    CHANGELOG `[Unreleased]` line.
15. T12 — stack repo: typed `services.telephony.webphone.lens.url`
    option + eval case.
16. T18 — full battery: `nix fmt`, docs-gates `--check`, buildflow,
    `go test -count=1 ./...`, closeout status report, INDEX regen,
    push, CI verdict — ALL in the same session as the push.
17. T13 — `scripts/loc-normalized.sh` + self-test (kills the d2
   error class permanently).
18. T14 — measure both binaries (size + idle RSS) into the evidence
    appendix.
19. T15 — `go mod graph` diff + co-change query.
20. Verify lens auth deep-link carry-over (read lens's auth
    middleware; if it drops the path through the login redirect, that
    is a lens-side fix to file BEFORE the link ships).
21. T16 — OIDC extraction API sketch from `oidclogin` vs `sso`.
22. T17 — upstream cqrs-htmx OIDC-provider ask draft (provenance
    banner, like the appkit asks).
23. T19 — one-repo/two-binaries memo (cost the AGENTS-cap arithmetic
    honestly).
24. T20 — shared call-vocabulary module memo + gate-2 trigger.
25. T21 — promotion notes: lens backup lib + FTS read-model pattern.
26. T22 — identity/retention alignment memo (MaxUsers=1 vs
    multi-user; GDPR semantics).
27. T23 — Reports-tab-becomes-lens-backed spike.
28. T24 — set the revisit date + owner-tick note.
29. After this report lands: confirm the docs-gates ratchet stayed
    green (report count vs `max_unarchived: 65`).
30. Check TODO_LIST row 1658 still matches the plan's task ids.

(Items beyond 30 are the plan's own micro table — 96 rows, each
≤12 min, already sorted; executing f4–f16 IS the 1%+4%+20% of the
plan. Nothing else was invented here.)

## g) Questions I cannot answer myself

1. **Execute or hold?** The plan is committed and CI-green; every
   execution step is gated on your tick (the plan's own register says
   so). Start at T05 now?
2. **Affordance shape for the History link** — trailing text link,
   icon-only, or row-level affordance? This is an owner's-eyes call
   under the quiet-chrome doctrine; I can build any of the three but
   must not pick for you.
3. **Is `usermgmt.db` deliberately host-managed** (backed up outside
   this repo's script, by the stack or the host), or is its omission
   from `package/backup-script.nix` a bug T04 should fix? T03 can
   search the stack, but only you know the deployment's intent.

## Self-review answers (brutal-self-review, folded per your scope)

- **Did I lie?** No. One number set (LOC totals) was wrong first-pass
  and corrected the same session; nothing known-false stands.
- **Ghost systems?** None created — this session shipped docs and a
  plan only; the plan's non-goals explicitly fence against the
  ghost-system shapes (no private module, no merge).
- **Split brains?** One watched candidate, already fenced in the
  plan: T12's typed `lens.url` vs the raw `settings.lens.url` lane
  (one-home-per-fact eval conflict rule applies; the plan states the
  raw lane stays untouched when the typed one is off).
- **Tests?** No code changed, so none owed this session; the plan
  specifies the parity/escaping/decode tests as first-class micro
  tasks, not afterthoughts.
- **Scope creep?** The plan touches two other repos (stack, lens
  AGENTS) — each is a separate commit by guardrail, not creep; the
  review turn added zero scope.
