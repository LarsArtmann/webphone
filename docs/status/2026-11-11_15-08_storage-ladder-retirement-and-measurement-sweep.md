# Storage-ladder retirement (faxes + enroll), watcher dedup repair, and the measurement sweep — 2026-11-11 15:08 CET

> Claims verified as of 15:08 CET. Push/CI statements rot fastest: the
> verdicts below name their runs/shas; the concurrent sessions are LIVE
> and will move both repos past this snapshot.

**Scope line.** The continuation of the whole-TODO-list execution session:
the big storage refactor (the prod-dead readmodels fax + enroll engine
stores deleted, their tests moved to the cqrs journal backends), the
red-main watcher dedup defect found by its own verification and fixed,
the cqrs-lint advisory triage (the 185-warning premise decayed to six
findings — two fixed, four adjudicated), the Phase-4 KVM preflight legs,
the crm cross-gate confirmation, the catalog.home.lan publish verify,
and the docs-health cadence archive that paid for this report.

## a) FULLY DONE

1. **readmodels faxes store + EnrollTokens DELETED** (TODO row 2026-11-09
   - the doctrine-sweep rider): `faxes.go` gone (FaxView/faxInput moved
     into faxesgate.go; FaxesProjection/FaxJobOfView/FaxGate kept — the
     fold one-home the live journal projection consumes), `enroll.go`
     slimmed to EnrollGate + EnrollTokenView, the Engine lost the
     faxes/enroll fields, accessors and Plan entries. Tests moved: the fax
     double boots `cqrs.FaxesDomainDecl` + `NewFaxes`
     (fax/teststore_test.go + archiver_test.go, synctest archiver green on
     the journal backend), userauth/session_switch/passkey_api tests boot
     `cqrs.NewEnrollSystem` + `NewEnrollTokens`. The dying readmodels
     coverage PORTED to cqrs (ListAll newest-first + ByProviderRef guards +
     the 11-row faxMoves legality mirror + the idempotent second-key arm);
     the admin-reads fax pins moved to cqrs/faxes_test.go. Registry regen'd
     (−14 codes); art-dupl `-t 1`: 0 clones reference readmodels/faxes.go —
     the 3 baselined twin groups dissolved exactly as the registry row
     promised. Full `go test -count=1 ./...` EXIT=0. Commits: 169a3df3,
     7b0a023b, 19b08bcf (daemon-swept), 74856f7e (docs).
2. **Red-main watcher dedup repaired** (TODO row's "live look" demand):
   the LIST-API verification caught TWO defects — the dedup lane rode
   `gh issue list --search` (the lagging search index, the 2026-11-04
   duplicate class) and `--label ci-red --limit 1` pointed the COMMENT
   verdict at ANY ci-red issue (six open daemon-audit flags share the
   label BY DESIGN → the next red main would have spammed evidence onto
   a daemon-audit flag every 15 min). Fixed: `red_main_issues()` = LIST
   api over ci-red filtered by the `red-main run` title prefix;
   decide() untouched; self-test 11/11 with a lane fixture; live dry-run
   answered CREATE against the real red tip of that hour. Commit
   fafdbb74.
3. **cqrs-lint advisory triage CLOSED** (TODO row 187): the pinned
   detector v4.21.0 over ./internal/cqrs reports SIX findings, not the
   row's decayed 185. TWO REAL FIXED: `SearchThreads` rode the engine's
   silent 100-row ceiling (search missed hits past 100 messages) →
   `WithLimit(0)` lifted per the package's own contract; all 13
   hand-built cqrs query decls gained `Volume(humanScaleVolume)` (the
   readmodels-side consistency). FOUR NOISE adjudicated with site
   rationale in the TODO strike (F032 scope, F031-variadic false
   positive — all four callers pass explicit limits, F021×2 deliberate
   projection models, F009 best-effort await). Commit 6d2306ed.
4. **Phase-4 KVM preflight legs GREEN** (TODO row 270): webphone-module
   (eval), webphone-backup-drill, webphone-backup — all three named
   builds pass on KVM at 6d2306ed.
5. **crm cross-gates confirmed** (TODO row 216): crm's full test suite
   EXIT=0 at e383ba9; its treefmt step is honestly not-applicable (no
   treefmt.toml); the webphone vendorHash repair never touched crm.
6. **catalog.home.lan publish verified** (TODO row 209, observable
   half): root 200 with generation stamp 2026-11-10T22:59:53 (run #240's
   window); `/docs/domains/telephony/1.0.0` server-renders webphone's
   CURRENT event vocabulary — the freshness proof that needs no hub
   access.
7. **Docs-gates battery PASS on the final tree**: format (the dprint
   repair commit 75ba79c5 healed the rc=20 reds that the pre-repair
   tips shipped), freshness, audit-all, annotate, routing, cadence
   (66→65 via the 18-39 android-report archive, both annotate gates
   green before the move), table-damage, strike-safety, status-f-section.

## b) PARTIALLY DONE

1. **The front-door snapshot report's §f** (the sibling session's 14-45
   report, guarded red by the status-f-section gate): completed
   mechanically from the report's OWN remaining-items (four rows, one
   resolved in-flight — see c-1). Their train owns the content; the
   section is the mechanical gate completion.
2. **The full buildflow run**: first execution died on a go-step flake
   (`TestAutoReviewDispatchRoundTrip`, `sql: statement is closed` — the
   detached review goroutine racing the fixture engine's cleanup under
   sandbox timing; passes plain AND `-race` locally incl. the whole
   server package at 137s) plus the faxes_test.go file-size error; the
   re-run after the file split is FULLY GREEN (EXIT=0, 3897 log lines,
   findings gate green at 45 baselined).

## c) NOT STARTED (deliberately deferred, triggers recorded)

1. **faxes_test.go over the 350-line cap** (surfaced by the sibling's
   14-45 report item 4): FIXED in this session before it could red
   another gate — the two fax-gate tests this session added split into
   `faxesgate_test.go` (270 + 101 lines); `buildflow -s
   file-size-check` green. Recorded as the cap's first catch-by-another-
   session.
2. **The transient proxy class**: the 13:27Z CI runs failed `[setup
   failed]` on proxy.golang.org stream errors (projectionadapter v4.6.2,
   identity-model v4.17.0) — retried green on the next pushes; no code
   defect. Watch: if the class recurs within a day it owes a flake-ledger
   row.
3. **The remaining 16 open TODO rows** are all genuinely external:
   stack-repo work (redeploy, coverage arms, ringhook docs), owner-gated
   (prod capture T05, bandit issue's owner banner, ALERT_WEBHOOK_URL,
   catalog bump, dashboard tag), overnight-window calibration (the
   night-run pair, the ring-guard positive path), cross-repo halves
   (android specs consumption, dashboard nonce lane upstream,
   wire-fixture triggers).

## d) TOTALLY FUCKED UP (owned — each healed in-session)

1. **The battery-on-commit class BIT ME**: markdown edits (dedup
   registry row, TODO strikes) landed without the dprint repair pass —
   three CI tips red on `docs-gates: gate format FAILED (rc=20)` before
   the repair commit (`75ba79c5`) carried canonical bytes. The runbook
   rule exists; the daemon's commit-and-push speed turned the gap into
   three red tips in ~20 minutes. Healed and the battery now runs BEFORE
   every commit in this session's flow.
2. **The archive banner split the claims block** (the sed-after-line-4
   insertion cut the report's verified-claims blockquote mid-sentence);
   caught on the immediate head-read, repaired before commit. The
   archive ritual's banner placement belongs AFTER the title, BEFORE the
   claims block — the 18-39 report's head is the shape to copy.

## e) WHAT WE SHOULD IMPROVE

- The daemon's commit race ate two narrative commit messages this
  session (the readmodels deletion, the docs train); amend-while-unpushed
  recovered one, the other is heuristic-message-only. The stage+commit in
  ONE call rule is necessary but insufficient against a daemon that
  commits within seconds — the amend-while-unpushed recovery should be
  the documented default reflex.
- cqrs-lint's F031-through-variadic and F032-scope false positives are
  recorded in the TODO strike, not filed upstream: one live sighting
  each, below the verify-before-filing bar. If either recurs in another
  package, the upstream report writes itself from the strike's rationale.

## f) Up to 50 things we should get done next

| # | Item                                                                                                                     | Verdict                                                                                                                 |
| - | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| 1 | CI verdict on THIS report's push (the docs + file-split train at `7aa38c83` and its successors)                          | the 7aa38c83 run completed success; successors ride the watch lane + one `gh run list` check                            |
| 2 | The stack browser E2E against the one-page front door + the stack `webphone-relock` ritual                               | rides the sibling session's stated next action (their 14-45 report §What-remains items 2-3) — stack-repo, not this tree |
| 3 | The night-run stability pair (≥8h apart on prod) + the ring-guard positive path during a real inbound call               | owner-present calibration rows (TODO 82/97), first prod capture session                                                 |
| 4 | The cqrs-lint noise classes (F031-variadic, F032-scope) → upstream go-cqrs-lite reports IF they recur in another package | trigger recorded in the TODO strike's rationale                                                                         |
| 5 | The BuildFlow bandit-provider issue filing (drafted, owner banner unchecked)                                             | stays owner-gated (TODO 116)                                                                                            |
