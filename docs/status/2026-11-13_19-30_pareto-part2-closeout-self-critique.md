# Status — pareto part-2 closeout: the honest self-critique (drain, bump, stack arms)

**Claims verified as of 19:25 CET.** Push/CI statements rot fastest: webphone
origin/main is RED at `a3e9c4d2` (the FOREIGN configjs quartet, tracker #252 —
re-checked at 19:2x, still red); the stack is pushed to `fb711a4` with its CI
verdict UNCHECKED; android and hub each carry ONE unpushed commit of mine.
Cadence: 64 unarchived reports against threshold 65 — this report makes it
65/65: PASSING but with ZERO headroom. The drain sizing error is mine (§d-7).

## Direct answers to the three opening questions

- **What did I forget?** Four concrete things: the go-cqrs-lite lib report's
  catalog Pin (still v4.14.0 + "bump rides the next train" — the bump LANDED
  today and I updated AGENTS but not the report); the android repo's mirror-row
  commit and the hub's TODO commit (both sitting unpushed for hours, noticed
  only at this closeout); the stack's CI verdict on `fb711a4` (the VM test ran
  locally green but the mirror's CI lane was never read); and a green CI run
  carrying MY commits has never existed — the only run covering them is the
  red one.
- **What could I have done better?** The WP-cardigan list is §d; the sharpest
  three: two multi-step command chains joined by newlines instead of `&&`
  burned two ~7-minute VM runs against stale code; edit anchors written from
  memory instead of copied from live bytes died twice (em-dash vs hyphen,
  `len("vm")` vs `len("vm-")`); and the voicemail arm shipped without a
  dump-first diagnostic, so discovering the device-less media pipeline took
  four VM runs instead of one.
- **What could I still improve?** §e has the systemic list; the meta-answer is
  that every failure this session was a rule I already knew (the runbook owns
  `&&`, the AGENTS own View-before-edit and explicit `git add` paths) applied
  too late rather than unknown.

## a) FULLY DONE

| Task                          | What landed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Evidence                                                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| Report + gates triage         | The 15-14 report's literal pipes inside a code span (strike-safety refused the flight-recorder metrics label) fixed by rewording; dprint canonical alignment; docs-gates isolated cadence as the only remaining red                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | `8174f452`                                                                                                                           |
| T10/T11 drain tranche         | THREE 19-day-old reports (2026-10-25 03-54 / 15-16 / 21-52) archived gates-first — annotool rebuilt from skill assets, `check-rows --strict` + `bare-rows` green BEFORE each `git mv`; §a rows pass-struck, §b/§c adjudicated against live facts (CHANGELOG grep, FEATURES grep, module-check greps, the stack's pbx-prod read); the OIDC report's three still-open asks ROUTED to a new TODO_LIST "OIDC lane residuals" section; banner + index                                                                                                                                                                                                                                                                                                                                                                                           | `015ce46d` + `83d517f5` + `4df1166b` + `72752cfa`                                                                                    |
| T16 catalog bump, root-caused | `go get catalog/v4@v4.15.0` failed because the upstream release tag `catalog/v4/v4.15.0` is INERT to Go's tag mapping (module `…/catalog/v4` resolves `catalog/v4.15.0`; the proxy's `@v/list` proved it never saw the release). Release commit verified green in a detached worktree (build + tests), correctly-shaped tag pushed on the SAME commit (the dashboard-v0.13.0 unblock precedent), proxy pickup <1 min, then the full tail: preflight (no removed symbols), `go get` + tidy, contracts regen byte-identical, catalog render green, vendorHash re-pinned, full Go suite green, full smoke 54/54 + all scenarios                                                                                                                                                                                                               | webphone `aaf81eef` (daemon-swept go.mod/sum) + vendorHash `70040eae` + docs `f1a29421`; upstream tag `catalog/v4.15.0` on `8d84d4b` |
| fleet-tags.sh correction      | The bump falsified the script's two-form theory: it checked `catalog/v4/v4.15.0` and `catalog/v4/v4/v4.15.0` (both unresolvable) and MISSED `catalog/v4.15.0` — a wrong-shape tag read as "tagged". Rewritten to derive the prefix by STRIPPING the go.mod major suffix; a wrong-shape-tagged release now FLAGS; self-test grown to 5 checks including the incident class; C-locale sort fix (locale collation reordered `<root>` after `catalog`); live fleet re-sweep clean                                                                                                                                                                                                                                                                                                                                                              | `14115fd6` + `f1a29421`                                                                                                              |
| F54 go-idempotency clock      | verify-before-filing executed (four `time.Now()` sites store.go:153/174/194/218 + the sweep ticker; the library's OWN suite sleeps real time at 8 sites across 5 files; dupe-search clean; checkout at latest tag v0.3.2); draft in `docs/drafts/go-idempotency-clock-issue.md`, checker 0 FAIL / 0 WARN after matching the banner to the CHECKER's regex (the SKILL.md prose quote differs from `bannerBody` — the checker is the authority); FILED and verified landed                                                                                                                                                                                                                                                                                                                                                                   | go-idempotency#14                                                                                                                    |
| F55 + F56                     | etag #5 + appkit #59/#60/#61 all OPEN upstream — nothing new to adopt; F56's `# expect` anchor verified ALREADY GREEN by the audit-verify replay (5/5)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | gh issue list + `audit-verify.sh`                                                                                                    |
| T22 docs-truth batch          | FEATURES: flight-recorder surface row ADDED (limiter, boot log, disk floor, module option, backup-tar fact), /metrics row gains the snapshot counter family, /healthz row gains the crm + projections legs (the archived SUPERB report's routed gap closed), dashboard row rides v0.13.3's no-eval CSP, composition-root row names the pre-drain hub release; ROADMAP observability-depth entry; lessons.md: "plan says X, code pins Y — amend same turn" + "new dataDir artifacts owe the backup path an adjudication"; AGENTS: flight-recorder bullet caught up, templ pin names its one home, fleet-tags.sh joins Commands, S001 note corrected (the enroll site died in the ladder closeout), error-registry regen note gains the MOVE fact, currency line: catalog ADOPTED; go-appkit.md: go-flightrecorder now a CONSUMED direct dep | `fe0a7816` + `45fadc86` + `03044b77` + `68c7e588` + `8b5d978f`                                                                       |
| T23 baseline + nolint         | `buildflow -s branching-flow --save-baseline` from THIS checkout with the result cache disabled: 220 findings re-recorded at current paths (COMPOSITION_mixin re-homed, stale old-path DUPLICATE_TYPE entries gone), all under the severity cap; `erraudit nolint-audit`: 69 directives, 69 needed, 0 stale                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `03044b77`                                                                                                                           |
| T25 residue strikes           | The file-size train's three actionable rows struck with evidence (raceprobe decision `c46b9666`, the baseline refresh, the never-fired ledger condition); AGENTS regen-MOVE note                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `cdec4338` + `68c7e588`                                                                                                              |
| T19 cqrs-lint prep            | The TODO rows (file the suppression-SM + D005 drafts, track the `/*` lexer fix on the next bump) verified in place; drafts verified in BuildFlow; S001: only `server/wirefixtures.go:43` survives — AGENTS corrected; the filings themselves stay owner-tick-gated                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | `8b5d978f`                                                                                                                           |
| T20 wire-fixture + android    | Both deferral triggers re-checked UNFIRED (config.js keys stable, no second consumer, first device train operator-blocked); the handover pack drafts the line-picking contract (sessionIdentity `extensions`, `/api/session/switch`, no pre-login listing by anti-enumeration posture) + the CallStyle mapping (ring payload facts, the no-PII-through-push ruling, API 31-33 spec shape); android TODO_LIST carries the mirror row                                                                                                                                                                                                                                                                                                                                                                                                        | `317fb3cd`; android `66d8371`                                                                                                        |
| T21 hub runner                | SystemNix forgejo.nix: `UMask=002` + setgid 2770 cache root/state roots + a root ExecStartPre repair leg (walk bounded to the two small state roots — gomodcache et al are hash-keyed read-only hits); SystemNix `nix flake check --no-build` green; the hub TODO row records the landing; ringhook operator docs + the validated forward + the dialplan E2E leg were already standing                                                                                                                                                                                                                                                                                                                                                                                                                                                     | SystemNix `8c108bdef` (pushed), hub `efab146`                                                                                        |
| T27 calendar                  | The 2026-11-13 refresh: decay pass + baseline re-record annotated on their rows; fleet-tags joins as a TRIGGER-ONLY row (the catalog incident is its pinned class); NO dated row moved — none had fired                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | `a908026a`                                                                                                                           |
| T17 voicemail arms — VM-GREEN | The stack gained: the browser.nix deposit leg (vmclient into 1000's box pre-launch), the phone-api/operator seam enablement (the first run proved the panel honestly renders the withheld-capability empty state without it), the E2E playback arm asserting the honest observable set (playing                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | failed-toast                                                                                                                         |
| T18 ring guard + night pack   | The inline re-shoot-on-ringing block existed TWICE (desktop + mobile passes) — extracted to `ring_guard_reshoot` with an overridable probe; three forced-state self-test cases; a still-up banner after the settle wait now names itself; `ui-night-pair.sh` schedules the ≥8h pair as two transient systemd user units                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | `87bc067e` + `938b9fbd`                                                                                                              |
| Final battery                 | `nix flake check` ALL PASSED (incl. contracts-fresh, island suites, the backup VM drills, on the catalog-bumped vendorHash); fresh smoke 54/54 + restart 4 + boot-failure 8 + asr 9 + dialgate 9 + faxcap 8 + metrics 11 + enroll 5; full Go suite green; docs-gates fully green (before this report's birth)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | session log                                                                                                                          |
| Plan closeout strikes         | The SUPERB pareto plan's task table struck row by row with verdicts + evidence; owner-gated rows explicitly NOT DONE by design                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `378ffca2`                                                                                                                           |
| Pushes per the bar            | webphone pushed manually (anchor: `015ce46d` @ 15:40 local, 60-min bar met); stack pushed manually (anchor: `0d78f9f` @ 16:18 local, 71-min bar)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `a3e9c4d2..` + `fb711a4`                                                                                                             |

## b) PARTIALLY DONE

1. **A green CI run carrying MY commits does not exist.** The only run covering
   today's trains (`43455109901` on `a3e9c4d2`) is RED — on the FOREIGN
   configjs quartet, attributed and tracker-filed (#252) — but my baseline
   refresh, fleet-tags fix, and docs rode that run. My code-level verdicts are
   the local batteries (flake check, smoke, suite); CI-green proof pends the
   quartet's heal. Next session: verify a green run on a tip carrying them.
2. **go-cqrs-lite.md's catalog Pin is STALE** — still `catalog/v4 v4.14.0` with
   "bump rides the next go-cqrs-lite train", while the bump landed today. I
   updated AGENTS' currency line but missed the report's Pin/Currency block.
   Discovered at THIS closeout; unfixed (§f-2).
3. **T18 F61's wall-clock half** — the scheduler pack landed; the ≥8h pair
   itself needs its quiet window. First pair sets the diff noise floor.
4. **verify-live.sh's audio leg** is bash-syntax-checked and its CSRF
   extraction tested against the real response shape — never run against a
   live host (needs owner env creds; a probe must not invent credentials).
5. **`ui-night-pair.sh`** is syntax + usage-gate checked — never executed.
6. **The stack's CI verdict on `fb711a4`** — unread. The VM test is the strong
   local gate; the mirror's CI lane is unverified.
7. **Android (mirror row) + hub (TODO record) each carry one unpushed commit**
   of mine, hours old — noticed only at this closeout (§d-9).
8. **The webphone CI red is UNRESOLVED** — still failing at re-check; tracker
   #252 owns it; my fix-forward obligations (bounded wait + surfaces named +
   both options) are discharged, the heal is the authoring session's.

## c) NOT STARTED

- T24 authenticated prod capture, T26 quiet-host release hygiene, T01/T02
  owner sitting packs, T06 prod redeploy — owner-gated by design.
- ALERT_WEBHOOK_URL paste (hub A19) — owner.
- tag-release.sh tag-shape fix upstream (go-cqrs-lite) — awaiting the owner's
  direct-fix-or-file ruling (§g-1).
- The OIDC lane residuals (VM test, oauth2 golden, typed passkey options) —
  routed to TODO_LIST, not started.
- The next archive tranche (the six 2026-10-27 reports) — the cadence bank is
  now ZERO (§d-7); this is §f-1.

## d) TOTALLY FUCKED UP

1. **`git add -u` swept the concurrent session's four in-flight configjs files
   into a style commit.** Caught before push (the diff stat named files I never
   touched), repaired by `git reset --soft` + restaging only my four files —
   their work returned to the working tree untouched. But the CLASS repeated:
   part 1's sin was `git add -A`, this half's was `git add -u`, and both exist
   only because explicit paths were abandoned for one lazy command. The daemon
   then swept and PUSHED their half-written quartet anyway (`a3e9c4d2`),
   reding CI — that part was never mine to prevent.
2. **Two multi-step chains joined by newlines instead of `&&`.** A failed
   python edit step did not stop the build step, so TWO VM runs (~14 min)
   executed against stale code. The runbook rule ("stage+commit in ONE call",
   "&& the dependent steps") exists; I violated it twice under time pressure
   and ate the cost.
3. **Edit anchors written from memory, not copied from live bytes.** One died
   on an em-dash vs hyphen mismatch (the file's comment carried "—", my anchor
   "-"), one on `len("vm")` vs the file's `len("vm-")`. The AGENTS rule says
   View before Edit precisely for this; under time pressure I trusted my own
   transcription twice. The report update also hit it once ("already LANDED"
   case difference). Three anchor failures, all self-inflicted.
4. **The fleet-tags two-form theory shipped inside a "done" verdict.** The T15
   closeout asserted the fleet tags in TWO conventions — an interpretation of
   the bump's failure message, never derived from the authority (Go's tag
   mapping, the proxy's `@v/list`). The catalog incident then falsified it:
   the script had been validating a tag Go cannot resolve, i.e. a false-clean
   in the exact class the script was built to catch. Same-day correction +
   self-test pin, but the wrong verdict existed for hours.
5. **The voicemail arm had no dump-first diagnostic.** The device-less media
   pipeline discovery took FOUR VM runs (~28 min): wall clock → wp-playing →
   toast-or-playing → readyState, before currentSrc closed it. A try/except
   dump (page state, readyState, currentSrc, toasts) in the arm's FIRST
   version would have collapsed the ladder to two runs.
6. **The drain sizing ignored the session's own report births.** I banked to
   63/65, then wrote two more reports — 65/65 now: PASSING with ZERO headroom,
   one report birth from red. The whole point of banking was to survive this.
7. **The catalog lib-report Pin went stale within the hour.** T07's row
   proudly said "lib-report updated"; T16's catalog bump updated AGENTS but
   not go-cqrs-lite.md. The "audit corpus stays honest" goal was the ROW'S OWN
   why, and I missed the corpus half of it.
8. **My own report carried the pipes bug.** A metrics label with literal pipes
   inside a code span — the exact shape strike-safety exists to refuse — in a
   table I wrote. The tool caught it on first run; no false green shipped, but
   the author and the reviewer were the same session.
9. **Cross-repo closeout discipline lapsed.** The android mirror row and the
   hub TODO commit sat unpushed for hours, and the stack's CI verdict was
   never read — all noticed only when THIS report's push-state survey ran.
   The release-runbook obligation ("verify end states with ls-remote + CI")
   was applied to webphone and skipped for the three siblings.
10. **A green CI verdict for my own trains was never obtained.** The manual
    push landed my commits onto a tip that was already red from the quartet;
    the attribution is sound and the tracker stands, but "attribute the red"
    is not "prove the green" — my local batteries are strong, yet the CI-green
    proof simply does not exist yet.

## e) WHAT WE SHOULD IMPROVE

1. **Derive rules from the authority, not from a failure message.** "The proxy
   rejected the bare form" justified a two-form theory for hours; `@v/list` +
   the tag-mapping rule would have killed it in minutes. When a verdict rests
   on interpreting an error, re-derive from the spec/source before writing it.
2. **Every new browser arm ships with a dump-first except-block** (page state,
   element attrs, readyState, currentSrc, toast text) — the VM round-trip is
   too expensive to spend on blind iteration.
3. **`&&` between every dependent step, no exceptions** — the rule was in the
   runbook before this session; the session proved its price twice.
4. **Edit anchors: View the live bytes, then paste.** Em-dashes, unicode, and
   one-character identifiers are exactly where memory transcription fails.
5. **Drain sizing budgets the session's OWN outputs** — a tranche that banks 2
   while the session will write 2+ reports banks nothing. Target ≤61 when a
   reporting session is planned.
6. **The bump tail grows one step: the audit-corpus sweep.** After vendorHash,
   sweep every lib report's Pin/Currency block naming the bumped module (the
   AGENTS currency line is not the one home — the reports are).
7. **Cross-repo closeout ritual: for every repo touched, `ls-remote` + push
   state + CI verdict before declaring done** — siblings are not second-class.
8. **Stage explicit paths. Always.** `add -A` and `add -u` are both unsafe
   under concurrent sessions; the lesson now has two entries.
9. **The checker outranks the prose.** The github-voice SKILL.md's banner
   example diverged from the checker's regex; trust the mechanical gate first,
   read the prose second (one wasted round-trip).
10. **Fold pending verdicts before writing a report** (or write the report
    last): the 17-05 report shipped with "IN FLIGHT" caveats that needed a
    same-session errata pass an hour later.

## f) Up to 50 things to get done next (tiered, impact first)

<!-- check-rows:ledger -->

| #  | Item                                                                                                                                                                                                                          | Tier |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- |
| 1  | Archive tranche NOW (the six 2026-10-27 reports, gates-first) — this report sits AT the 65 threshold with zero headroom; target ≤63 in the same session it runs                                                               | 1    |
| 2  | Fix go-cqrs-lite.md's catalog Pin → v4.15.0 + drop the stale "bump rides the next train" Currency clause (the bump landed today)                                                                                              | 1    |
| 3  | Tracker #252 resolution: when the configjs quartet heals OR is reverted, verify CI green on a tip carrying MY commits too (the trains have no green run yet)                                                                  | 1    |
| 4  | Push the android repo's mirror-row commit (unpushed since ~17:00)                                                                                                                                                             | 1    |
| 5  | Push the hub repo's efab146 (unpushed since ~16:5x)                                                                                                                                                                           | 1    |
| 6  | Read the stack's CI verdict on `fb711a4` (the mirror's ci.yml lane)                                                                                                                                                           | 1    |
| 7  | Fold the bump-tail audit-corpus sweep into the release-runbook/bump checklist (lib-report Pin + currency lines) — §e-6's one home                                                                                             | 2    |
| 8  | Add the dump-first diagnostic helper for browser arms (page state, readyState, currentSrc, toasts on any arm failure) — ui-capture.py + the stack E2E                                                                         | 2    |
| 9  | Run the night pair in the next quiet window (`ui-night-pair.sh`); calibrate the diff threshold from the pair                                                                                                                  | 2    |
| 10 | Owner: go-idempotency#14 review tick (provenance banner awaits manual review)                                                                                                                                                 | 2    |
| 11 | Owner: the filing-posture ruling (direct-fix-or-file vs human tick) — it gates #13, #14 and the cqrs-lint drafts                                                                                                              | 2    |
| 12 | Owner: ALERT_WEBHOOK_URL paste (hub A19) + one deliberate red to prove alerting                                                                                                                                               | 2    |
| 13 | tag-release.sh: fix the tag-shape rule upstream (the release script produced the Go-invisible shape) — direct fix or file per #11                                                                                             | 2    |
| 14 | Owner: T24 authenticated prod capture + `--threads` deep-link matrix on real data                                                                                                                                             | 2    |
| 15 | Owner: T26 quiet-host hygiene (stale-baseline decision, timings re-baseline, SQLITE_BUSY correlation)                                                                                                                         | 2    |
| 16 | Owner: answer gf-logs' actual path (§g-2) so the runner provisioning leg provably covers it                                                                                                                                   | 2    |
| 17 | verify-live.sh audio leg: one live-host validation run with owner env creds (the leg has never executed against prod)                                                                                                         | 3    |
| 18 | OIDC residuals (TODO_LIST section): the NixOS VM test, the second oauth2 golden, typed passkey options                                                                                                                        | 3    |
| 19 | The history-500 class: the stack's `read_cdr_rows` unpacks empty CDR state to a 500 on fresh deployments (observed in VM runs 3-10: `GET /phone-api/history` 500 pre-call) — fix or file on the stack side with this evidence | 3    |
| 20 | The 2026-10-25 trio's §f prose rows were archived UNSTRUCK (ledger-exempt prose) — harvest-check their open inventory at the next docs-health sweep                                                                           | 3    |
| 21 | battery gate name (docs-gates "battery" step) — confirm the rc=20 format failures were only ever dprint, never a deeper tool fault (three same-day recurrences)                                                               | 3    |
| 22 | ui-night-pair.sh: replace the `sleep 1` placeholder bash with a direct second delayed unit (the script works but carries scaffolding)                                                                                         | 3    |
| 23 | The annotool rebuild command belongs in AGENTS' docs-gates bullet verbatim (it is there; verify it survives the next AGENTS compaction)                                                                                       | 3    |
| 24 | BUILDFLOW_NO_RESULT_CACHE=1 on every scoped verify run — consider making the wrapper warn when a scoped run reads a cached verdict                                                                                            | 3    |
| 25 | C-locale sort: audit the other repo scripts for locale-sensitive sorts (the fleet-tags self-test failure class)                                                                                                               | 3    |
| 26 | check-rows:ledger marker on every NEW report's §f table (this report carries it; retrofit-audit older live reports)                                                                                                           | 3    |
| 27 | The 22-50 file-size report's remaining §f rows (§b of its residue section) — confirm the sibling session struck theirs; the two sessions' rows must not fork                                                                  | 3    |
| 28 | `scripts/ui-capture.py` ring guard: the extracted helper's two call sites pass `shoot`/`target` positionally — consider kwargs for readability at the next touch                                                              | 3    |
| 29 | The smoke's anonymous-CSRF extraction now rides /login (front-door train) — verify-live.sh's leg mirrors it; keep the two in sync at the next auth change                                                                     | 3    |
| 30 | lessons.md: "anchor edits from live bytes" + "chains joined with &&" war stories (§d-2/3) — write them while the tracebacks are fresh                                                                                         | 3    |

## g) Questions I can NOT figure out myself

1. **The filing posture (third time asking — it gates real work).** For
   OWN-repo findings, is banner-and-file the default (go-idempotency#14's
   shape), or does every filing wait for your tick first? And does the same
   answer extend to direct CODE fixes in your repos (today's BuildFlow
   persist-fix, the catalog tag push, the SystemNix runner leg — all landed
   without a pre-touch tick)?
2. **Where does `gf-logs` live?** The hub row says "same treatment for
   `gf-logs`", but no fleet config defines that path. The runner leg
   provisions `ec-hub` + `gf-logs` with `|| true` tolerance — if gf-logs is
   on ANOTHER host's unit, my leg silently never covers it and the row's
   promise is half-kept. A hostname/unit name is enough.
3. **Cadence budgeting policy.** Should the drain tranche target include the
   session's own report births (drain to ≤61 when a reporting session is
   planned), or is riding at 65 acceptable with a next-session tranche as the
   recovery? I can size either; only you know whether you want the corpus
   ceiling treated as hard or as a ratchet-with-lag.

_Report: docs/status/2026-11-13_19-30_pareto-part2-closeout-self-critique.md.
Format `.md` per standing instruction. WAITING FOR INSTRUCTIONS._
