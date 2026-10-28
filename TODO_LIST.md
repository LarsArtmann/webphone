# TODO_LIST

Short- and mid-term improvement tasks. Done work is DELETED, never struck
through (docs-health: shipped work moves to CHANGELOG / FEATURES — one home
per fact). Long-shot ideas and owner calls live in ROADMAP.md; every row
cites its source report in Evidence.

Last sweep: 2026-10-28 (whole-list execution session, concurrent with the dialer train):
whitespace-drift formatter adjudication + `__wpClock` clock seam + branching-flow
triage (36 reasoned nolints + fresh 78-entry governed baseline) SHIPPED · advisory
tw-freshness CI lane + drifted artifacts rebuilt SHIPPED · chat lane v2 (compose-new
by address shape + matrix m.read receipts) SHIPPED · catalog cross-link
(specifications frontmatter + spec files in the export tree) SHIPPED · fleet ask
FILED as pma#348 (verified at source) · cqrs-htmx M12/M13/M14 closed as
ALREADY-SHIPPED upstream (issue #32 + e7b62cd9 ring buffer; usermgmt README §
Projection readiness; crush-config a9aeb2a — no duplicates filed) · dial-code
registry healed (the concurrent train's split registration) · KMP spike DEFERRED
(live concurrent call-stack train in webphone-android — Gradle surgery would
collide; Tier-0 may-stay-unstarted by its own plan).
Prior sweep: 2026-10-27 (multi-call dock + agent-UX SUPERB train executed end-to-end: M1–M14 + M16–M18 shipped, tails recorded below) (whole-list execution session: the unpushed go.sum
daemon commit triaged canonical (`go mod tidy` no-op; vendorHash insensitive;
pushed green) · stale rows closed with evidence re-verified at HEAD (go-appkit
flight recorder + latency metrics SHIPPED, branching-flow unskip EXECUTED,
F74/F85/F86 closed, catalog push confirmed on origin/main, the stale passkey
gopls handoff retired — LSP clean) · push-seam Settings surface SHIPPED
(Ring-while-closed row + read-only device list, conditional, token never
rendered) · ASR #13 saved-to-History affordance + #21 whisper.cpp
`/inference` provider variant SHIPPED · metaengine v4.19.0→v4.20.0 bump train
EXECUTED (codec contract read at source, capability assertions held, all
gates green incl. KVM drill, vendorHash re-pinned, asyncapi regenerated
content-inertly) · the seven catalog-exporter candidates FILED upstream
(go-cqrs-lite #91–#97, premises source-verified, dupes searched, drafts in
docs/drafts/catalog-exporter-2026-10-25/)) · GitHub-issue review session:
all five open issues handled — #5 daemon-audit flag closed as healed
(flagged commits rewritten away in the day's rebase, tree canonical at
HEAD), #6–#8 dialer-adoption claims verified with corrections posted and
triaged into the new section below, #9 Kith switch-on corrected (no crm
module options exist; webphone slice scoped below)).
Prior sweep: 2026-10-18 (round-5 resumed execution closed the data-model, metaengine, ASR, and process tails: metaengine row CLOSED — G3 shipped via v2.9.0 (`0f1031d` verified in-tag), F019 Volume hints landed on all five query declarations, F026 adjudicated NOT-DO (no cursor pagination at HEAD to cache); series-3 harvest row executed via round-5 E9–E11; ASR hardening/polish/test-debt bullets closed (#22/#23/#37/#39–#44/#47/#5 + #11/#12); island D-train executed (ef29ea9); module-cache janitor landed + live-swept; upstream templ-components #27 + branching-flow #9/#15/#16 verified ALREADY-FILED-COMPLETED — no duplicate filings; go-modularize §f 1–40 annotated with routing verdicts; carve row LOC refresh 4,056 wiring / 5,204 tree, two methods agree).
Prior sweep: 2026-10-12 (TODO-execution session: settingsRow→DefinitionList adoption, the first art-dupl `-t 1` deep sweep + baseline + CI dup-gate, the daemon dep-guard commit hook + push-stall watchdog, and ASR test debt #32/#33/#34/#36/#38 — rows updated below; still open from that plan: voicemail transcribe-button test #37, 429-combined-load #5, ASR retry/backoff + toast #43/#44, the docs cluster, the branching-flow unskip re-triage). Same-day prior sweep (SUPERB unified-auth train tails harvest: the 10-52 report's non-gated (f) items executed same-session — stack suite asserts + module-mirror eval check + docs; the owner-gated T13/T15 cluster, the three owner questions, and the hardening tail row below are what remains; 06-55 residuals verified against stack code and routed). Prior sweep: 2026-10-11 (branching-flow sitting harvest: all five dedup open owner calls ratified+recorded — `-t 1`/zero/upstream-FP-filing, settingsRow→ADOPT `display.DefinitionList`, registry=ONE home, quarterly suppressed-set audits, CRM/Paperless routed upstream; new section below + branching-flow#8/#9 filed + the stale branching-flow NO-nolint AGENTS note corrected). Prior sweep: 2026-10-08 (SUPERB ASR train M7 harvest: 16-54/18-15 §f + plan remainders folded into the ASR-follow-ups row below; visual-harness row refreshed to the 18-shot matrix). Prior sweep: 2026-10-06 (docs-health v7: 23:22 §f harvest below; the 2026-09 verdict/decision cohort annotated + archived — 11 files; TODO/ROADMAP/FEATURES citation repoints; briefing rows 36–38). Prior sweep: 2026-10-05 (Pareto-plan execution session). Folded then: the 14:50 mixin-review session §f (daemon-commit verification DONE — remote `5a4f9db`, content intact; QMD `get` diagnosis DONE — crush `#3846`, see Standing watches; compose noise-floor DONE — registry sweep log 16:40; mixin/CRM/Paperless/vcard/errorfamily.Join owner calls → briefing rows 29–32); stack CI verdict recorded (RED by 1h timeout); erraudit re-measured early (tier-1/2a/2b = 0/0/0 after 7 reasoned suppressions; boot surfaces re-graded intact). Prior sweep 2026-10-04 (passkey train-tail session; AGENTS-over-cap row removed, tail code-polish executed, boot-contract runbook patched, `890a526` CI re-dispatched).

## Open tasks

### Island type gate tails (2026-10-27/28 — gate SHIPPED and CI-WIRED; refinements below)

The gate itself is done and enforced: 680 type errors → 0 over 26 served files, `island-typecheck`
now runs in CI's `fast-checks` job (it was green locally and unwired in CI until 2026-10-28), the
how-to is [docs/type-gate.md](docs/type-gate.md), and the pins from `typecheck_pin_test.go` grew from
5 to 10 assertions (excluded glob reaches the bundle and nothing else, no narrowed include, the
`IslandDomIds` mirror of docs/dom-contract.md, plus a `globMatcher` table test).
Evidence: [docs/status/2026-10-27_22-00_island-type-gate-complete-status.md](docs/status/2026-10-27_22-00_island-type-gate-complete-status.md).

The three refinement rows that sat here (whitespace-drift false positive,
`__wpClock`, branching-flow triage) are SHIPPED 2026-10-28 — the gate now
adjudicates formatter output with the flake's own treefmt over the
materialized staged blobs (`--ref` mode too), the shell reads all time
through one exported clock holder with the global deleted, and the
branching-flow set is fully governed (36 reasoned nolints + a fresh
78-entry baseline wired into `.buildflow.yml`). See CHANGELOG [Unreleased].

- **DECLINED with reason — do not re-litigate without new evidence:**
  - _Split `shell.js` at its seams_ (voicemail player / palette / composer). The gate now makes a
    split SAFE (a broken import is a compile error), but the classic-script load order is a
    served-byte contract the stack E2E greps, and the payoff is readability only. Reopen with an E2E
    budget.
  - _Type-check `island-tests/*.mjs`._ Needs `@types/node` (not in the devshell) plus a third
    tsconfig, and the specs' stubs are deliberately partial duck types — the checker would fight the
    design the `hasMethod` doctrine rests on.
  - _Move the shell's copy into a typed key map._ The shell must work when the island scripts fail
    (that is why it is a classic script); a shared map adds a load-order dependency in the one place
    the shell's independence is the point.
  - _`tsc --explainFiles` in CI._ The coverage pin derives the checked set from the EMBEDDED tree, so
    a file that stops being included fails the pin — proved 2026-10-28 by narrowing the include to a
    single file (both coverage pins failed with an actionable message).

### Error-surface + crash-containment plan (2026-10-27 — P0–P2 SHIPPED, host legs + evidence-gated tails open)

Plan + execution log: [docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md](docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md).
Shipped same-session (see the plan's execution log + CHANGELOG): recover() guards on ALL
five bare-goroutine sites + containment test, the island catch sweep (1 bug fixed:
summarize network-dead split; justifications at every silent site), the call-failure
reason card + THE SIP-cause mapping (error-contract.md, coordinates with #6 below),
oxlint `no-empty`/`no-unused-vars`/`no-console(allow info·warn·error)` gates + finding
fixes, the shutdown release-count log, push draft wrap-code pins, errors.AsType, the
502-proxy + island-JS error-contract sections (+ stack runbook sync), lessons entries.
**Priority:** High · **Effort:** S remainder.

1. **Host-leg verdicts (OWNER, read-only ssh — banned in the agent shell)**: run
   `ssh root@pbx.artmann.tech 'systemctl show webphone -p NRestarts -p ExecMainStartTimestamp; journalctl -u webphone --since today | grep -iE "panic|fatal"'`
   and paste the output; a panic line reopens P0 forensics on the named seam. Same
   pass covers the Caddy /events flush+logging config check and the dialplan
   `state:end` ring-hook probe (plan M03/M53/M54). Prod `/version` already verified
   REMOTELY: v2.9.0 at `ec5de5e` (deploy fired; `agent:true` live).
2. **`/events` heartbeat + server `retry:` hint — NO-GO until drop-class evidence**
   (plan T19): recovery is guaranteed by the upstream recreate loop; no long-idle
   drop has ever been observed. Reopen trigger: a journal/SSE-log pattern showing
   feeds dying while the process stayed up.
3. **Upstream filing (owner + github-voice)**: the htmx SSE ext `console.error("[object
   Event]")` defect — diagnosis source-verified, draft note ready:
   [docs/research/2026-10-27_htmx-sse-ext-object-event-logging-note.md](docs/research/2026-10-27_htmx-sse-ext-object-event-logging-note.md).
4. **Post-deploy deep-verify (plan T18/T22 tail)**: after the next release reaches
   prod, confirm the reason card + offline toast + reconnect pill title render on
   the live deployment (browser E2E owed only if served markup changed — it did
   not: all surfaces are island-JS-created).

### Language stance: island type gate SHIPPED, Go→JS + full KMP rejected (2026-10-27 plan)

> **Item 1 CLOSED 2026-10-27 21:00** — the gate is enforced and green over all 26 served asset
> files (`tsc -p . --noEmit`, wired as `island-typecheck`), with 680 type errors retired and ten
> real defects fixed on the way. `tsconfig.all.json` is deleted and the debt list was replaced by
> stronger coverage pins. Execution record:
> [docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md](docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md)
> §9. Item 2 (the error-surface conventions) landed in the same train; items 3–4 stay open.

Plan + evidence: [docs/planning/2026-10-27_16-04_SUPERB-typescript-and-platform-language-pareto-plan.html](docs/planning/2026-10-27_16-04_SUPERB-typescript-and-platform-language-pareto-plan.html)
(the plan owns the task count; this row points at it). **Priority:** High · **Effort:** M.

1. **Island type gate (GO — Tier 1)**: JSDoc `// @ts-check` + `tsc --noEmit` + hand-written
   `types/globals.d.ts` (PBX_CONFIG) and `types/sip.d.ts` (vendored 0.21.2) + a sandbox-safe
   `island-typecheck` nix check + a progress pin (checked-file set never shrinks, no bare
   `@ts-ignore`). Served bytes stay IDENTICAL, so no stack browser E2E is owed. T01–T06 + T23 is
   the scaffold (~4 h); the per-file annotation campaign (T07–T22) is one commit per file.
2. **Island error-surface conventions (T24) — convention DESIGN only**: the typed rejection shape
   plus the no-silent-catch rule, worded so the annotations and the error doctrine agree. The
   runtime half (catch sweep, oxlint gate rules, `error-contract.md` island section) belongs to the
   sibling plan
   [docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md](docs/planning/2026-10-27_16-05_SUPERB-error-surface-crash-containment-plan.md)
   (T5/T11/T12); de-overlap recorded in both files.
3. **KMP protocol-layer spike (CONDITIONAL — T26/T27, `~/projects/webphone-android`)**:
   the codec + digest move to `commonMain` cleanly; `SipEngine` needs its okhttp transport behind
   `expect`/`actual` first (re-estimated at 180 min, not 15). Android stays the ONLY target, then
   STOP — no second target without a linuxX64 WebRTC 3-call audio smoke. Gate: a committed second
   target in ROADMAP. **DEFERRED 2026-10-28**: the android repo carried a LIVE concurrent
   call-stack train (daemon commits minutes old, Kotlin files mid-edit) — Gradle surgery on
   settings.gradle.kts + a new shared/ module would collide head-on. Tier-0 by the plan's own
   wording ("may stay unstarted indefinitely"); execute when that repo is quiet.
4. **REJECTED, do not re-propose**: Go server → JavaScript (57,349 LOC plus the
   single-binary/nix deployment premise for zero product gain) · full Kotlin Multiplatform (the
   2026-10-23 platform decision stands: Linux WebRTC is the weak leg) · a Kotlin/Wasm web target
   (rewrites the island + templ/htmx model) · island → `.ts` with emit (deferred behind three
   named triggers: a justified npm dependency, a JSDoc overhead ceiling, framework adoption).

### Dialer-adoption issues #6–#8 + Kith switch-on #9 (2026-10-27 — claims verified; #6 + the #9 config/health slice SHIPPED same evening)

UPDATE 2026-10-27 evening: #6 is SHIPPED — typed `domain.Disposition`
(answered/missed/voicemail/ivr/failed/unknown per the stop-the-lying
ruling, NOT the dialer's seven-value set), legacy outcome demotion,
ended-card confirm, `call_logged` verdict travel, history verdict
filter + badge. #9's typed `crm` NixOS lane, `crm.token_file`, and the
`/healthz` CRM leg are SHIPPED. #7/#8 remain unstarted. The per-item
rows keep their original wording for the audit trail.

Owner-filed adoption candidates from the dialer postmortem
(matthewdonsemail-lab/dialer PRs #20/#54); every gap claim verified against
HEAD, corrections posted as issue comments. Separate trains, none started.

1. **Call dispositions** (#6): operator-confirmed taxonomy
   (`connected`/`voicemail`/`ivr`/`no_answer`/`wrong_number`/`dnc`/`disconnected`)
   in `internal/domain`, default `connected` on Established, confirm through
   the CallNote UI; ONE total CDR-status→disposition mapping (unknown ⇒
   failure class, never answered); awaited write with surfaced failure; carry
   the disposition in the Kith `call_logged` payload. Correction recorded:
   `Outcome` IS filled today — the island sends
   `established ? "answered" : "missed"` (panels.js:78 → calls_api.go:108),
   so a voicemail call journals as "answered" in Kith right now.
2. **Per-remote redial cooldown** (#7): opt-in config (1–4h default when on),
   persisted (Owner, remote) window in the read-model (restart-proof),
   inline "dial anyway" chip in the typeahead row (never a modal — the #54
   lesson), state visible in the row. Verified: no cooldown exists today
   (typeahead.js suggests only; the history ↻ button dials unconditionally).
3. **AI call review artifact** (#8): UPDATE 2026-10-28 (multi-actor train
   read at HEAD): MOSTLY SHIPPED by the concurrent dialer train — the
   post-call review is AUTOMATIC (`settleEndedCall` dispatches
   `POST /api/calls/{id}/review`, 202 detached, idempotent with the manual
   button; island calls.js), the artifact IS persisted
   (`domain.ReviewState` pending/present/failed via readmodels
   `call_reviews`, rendered in History — `TestHistoryTranscriptReviewStates`
   pins the four renderings incl. the not-configured chip). Remaining
   open slice: the summary/points/score DIMENSIONS the plan named (today
   the artifact stores the draft text + confidence only) — reopen only if
   the operator asks for richer artifacts.
4. **Kith CRM switch-on, webphone slice** (#9): typed module options
   `crm.{enable,url,token_file}` (LoadCredential pattern like oauth2 — no
   crm options exist in package/nixos-module.nix today, and its leak guard
   rejects path-valued settings secrets) + auth-failed surfacing (Settings
   shows only connected/unset, settings.templ:202; no /healthz crm leg;
   `ErrUnauthorized` must turn a surface red within one poll cycle). The
   issue's steps 1/2/4 (PKI trust, token provisioning, e2e verify) are
   pbx-artmann-side; the cited diagnosis doc lives there too.

### Multi-actor HITL (2026-10-28 — substrate SHIPPED same evening; escalation tiers below)

Plan + HITL ladder + actor model:
[docs/planning/2026-10-28_21-54_SUPERB-multi-actor-hitl-pareto-plan.html](docs/planning/2026-10-28_21-54_SUPERB-multi-actor-hitl-pareto-plan.html).
SHIPPED in the same train: session `LoginEmail` (passkey/OIDC mint,
guarded migration, byte-parity for extension logins), audit lines naming
the human actor (`agent *` + logout; `TestAgentAuditNamesTheHumanActor`
pins both shapes), `authmap.MappedByLogin` whoami preference +
deterministic shared-extension lookup, drafts-panel pending age
(`draftAge`, foreign-wire tolerant). **Priority:** per-row · **Effort:**
per-row.

1. **Extension switcher (D1)** — SHIPPED 2026-10-28 (whole-list
   execution session): `POST /api/session/switch` re-mints the session
   for another extension mapped to the SAME login email (fresh
   password-file read + directory verify + CSRF rotation + old row
   deleted after the mint; tight PBX-bound rate bucket; `auth switch`
   audit line names the human move; extension logins refused, own-line
   switch an idempotent no-op, 404 with the identity layer off). The
   session identity carries `extensions`; more than one renders a
   JS-created select by the whoami line (no served-markup delta, no DOM
   contract change); selecting a line POSTs + reloads through the ONE
   boot path (re-REGISTER, hub resubscribe, panels — the no-reload
   sequence stays future polish), and a LIVE CALL blocks the switch.
   Pinned: session_switch_test.go (6 cases) + session-switch.test.mjs
   (5 cases).
2. **Veto-window mode / HITL L2 (D2)**: draft auto-sends after T unless a
   human vetoes; the missing rung between approve-everything (L1) and
   confidence-gated autopilot (L3, shipped agent-side). Lives in the
   stack's textagent (send policy) — webphone side only surfaces age +
   countdown (the age substrate landed). Owner GO + stack train.
3. **Agent registry + scoped machine tokens (D3)**: today ONE machine
   actor (the textagent bearer secret, loopback). Named agents with
   scopes (draft/read/send) + per-agent audit prefixes + an API-token
   lane for agents-as-users. Owner GO; security review first.
4. **Delegation + roles (D4)**: approve-on-behalf-of, per-contact scopes,
   admin/member split — needs the actor substrate (shipped) AND a second
   real user first. Owner call.
5. **Durable audit read model (D5, F107 reopen trigger)**: multi-actor
   makes in-UI audit history a real operator need; the journal-line home
   stands until then.
6. **Branching-flow baseline staleness (observed 2026-10-28 train-start
   buildflow)**: `.buildflow-baseline.json`'s 78 accepted entries matched
   nothing in 3 consecutive runs (morning session's governed baseline —
   suspect host-absolute IDs again). Re-record from THIS checkout with
   `buildflow -s branching-flow --save-baseline` or retire it; owned by
   the branching-flow governance row, not this train.

### Daemon push-stall forensics (2026-10-27 T10 — root cause found, fleet fix pending)

The evening's push stalls are the daemon's push leg dying with its
PROCESS, and fleet deploys bounce that process: the 20:19:41
`switch-to-configuration test` (SystemNix) terminated
`projects-management-automation.service`; pushes stopped 20:09:15,
resumed 20:40:36 (webphone's oldest unpushed commit waited 30m37s;
follow-on gaps hit 54m and 70m). The journal logs NOTHING about pushes
(the T21 LOCAL-timer verdict stands), and the commit-message LLM
provider (127.0.0.1:52625) refused connections 13:06→20:59 — ~8h of
heuristic-only messages fleet-wide, though the commit leg itself stayed
fast (177 ms fails). FLEET ASK FILED 2026-10-28 as
projects-management-automation#348 (verified at source first: Stop()
  drops pending batches with no drain, Start() has no unpushed sweep,
  per-push logging already exists — so the ask is the SIGTERM flush, the
  startup sweep, and a shutdown-completion line; draft in
  docs/drafts/pma-push-leg-durability-issue-2026-10-28.md). Until it
  lands, a fleet deploy window predicts a push stall — run
  `scripts/push-stall-watch.sh` after every `switch-to-configuration`.

### Multi-call dock + agent-UX train (2026-10-26/27 SUPERB plan — SHIPPED, two tails)

Pareto plan: [docs/planning/2026-10-26_08-11_SUPERB-multicall-dock-agent-ux.md](docs/planning/2026-10-26_08-11_SUPERB-multicall-dock-agent-ux.md)
(20 medium / 120 micro tasks; round-two errata in the plan doc: M15 search
already shipped, notes home = readmodels not the store ladder, M18 durable
journal NO-GO, M19 NO-GO on this host). M1–M14 + M16–M18 shipped across
2026-10-26/27 (ring queue, dock, captions, ended-card summarize chain,
agent-mode pill, draft-ready push, call notes with AI prefill, locality
labels, audit lines). Open tails:

1. **Stack browser E2E re-run** (release-runbook obligation, ~445s
   budget): any served-markup change owes it; this train touched the
   calls island + History rows. Owner-runnable from the stack checkout.
2. **M18 tail — Settings last-N agent-actions row**: needs a durable
   agent-action read model; the trigger is an operator who needs in-UI
   history (the `agent <event>` journal line covers audits today).
3. **M19 re-open trigger — attended transfer**: the stack grows an FS
   lab with an attended-REFER-with-Replaces VM scenario; the island
   delta is small (sketch in
   `docs/research/2026-10-27_attended-transfer-spike.md`).

### Error-agents tooling train (2026-10-26 SUPERB plan — ALL 20 M-tasks SHIPPED same day)

Pareto plan: [docs/planning/2026-10-26_10-33_SUPERB-error-agents.md](docs/planning/2026-10-26_10-33_SUPERB-error-agents.md)
(20 medium / 79 micro tasks, mermaid execution graph; PURE TOOLING — zero
product code, advisory-first; source: the 2026-10-26 practice scan
`docs/research/2026-10-26_ai-agent-error-management.html`). SHIPPED, with
evidence:

1. **P0 — Error-code router** (M1–M2) SHIPPED: `scripts/error-context.py`
   (`--self-test` 37/37; README troubleshooting one-liner + AGENTS
   Commands line).
2. **P0 — Reproduce-first gate** (M3–M5) SHIPPED:
   `scripts/reproduce-first-check.sh` (`--self-test` 3/3 scenarios) + the
   advisory `reproduce-first` CI job (`continue-on-error`; blocking flip
   is a documented later decision).
3. **P1 — Observers** (M6–M9) SHIPPED: `watch.yml` */15 cron →
   `main-red-watch.sh` (tracker issue with DRAFT ledger row; `--self-test`
   8/8) + `daemon-audit.sh` (guarded files + drift via the new
   `whitespace-drift.sh --ref` mode; `--self-test` 3/3). NO dispatch.
   First live catch on day one: the drill-era incident class (six daemon
   commits, zero CI runs, red main ~2.5h) is the ledger row for run
   40301017699 — exactly what the watcher closes.
4. **P2 tails** (M10–M20) SHIPPED: `bump-preflight.sh` (verified on two
   real histories; `--self-test` 4/4), `incident-bundle.sh`
   (`--self-test` 6/6), the KVM-gated `webphone-error-drill` VM test
   (built green locally, ~20s VM time), the two gated sketches in
   `docs/drafts/gated-sketches/` (LLM prose gate, agent dispatch — owner
   GO recorded as trigger criteria, nothing wired). Plan errata: drill
   scenarios B/C stay smoke-owned (one home per fact).

REMAINING (owner calls, not tasks): the reproduce-first blocking flip
(after N clean advisory trains), the two gated GOs, and the daemon
hook-override fix that lives in the nix-config pma module.

### cqrs-htmx leverage-maximization tails (2026-10-25 SUPERB train — M1–M11 SHIPPED same day)

The critical path is done: identity-train stub tail + vendorHash (un-reds
main), projections readiness on /healthz, hub Shutdown pre-drain, clone
detection + auth-audit journal verified live, report re-scored 100/100
(§07 Addendum). What remains are the optional tails:

1. **Upstream: bounded AuditLog projection ask** (M12) — RESOLVED 2026-10-28
   without filing: the ask is ALREADY SHIPPED upstream. Issue #32
   ("AuditLog projection: bound the in-memory entries") existed, and
   `e7b62cd9` (2026-10-27, on origin/master) implements the ring buffer
   with cap + eviction + `Dropped()` and no way to re-enable unbounded
   mode — verified at source, NOT in the pinned usermgmt v4.16.0. Close-
   out comment posted on #32. Remaining webphone action: bump usermgmt
   once it tags, then the read-side audit surface unblocks (wire, not
   file).
2. **Upstream: usermgmt README pointer** (M13) — RESOLVED 2026-10-28:
   `usermgmt/README.md` § "Projection readiness (health checks)" already
   carries the paragraph + the `ReadinessHandler(ProjectionReadinessCheck(svc))`
   example + the `OnProjectionFailed` twin (landed via the
   `docs/usermgmt-projection-readiness-pointer` branch, on master). No PR
   needed.
3. **Process lessons → crush-config** (M14) — RESOLVED 2026-10-28 as
   already shipped during the 10-25 train: crush-config `a9aeb2a` carries
   both rules ("Verify 'exactly N' claims with symbol-level greps" +
   "Run a cheap build+test baseline at session start").
4. **loginpage next-tag passive watch** — favicon + NewPageData fixes
   ride the tag AFTER v4.13.0; bump when it lands, nothing to do now
   (report finding 5).

Evidence: docs/planning/2026-10-25_19-26_SUPERB-cqrs-htmx-leverage-maximization.md
(errata filled); docs/research/2026-10-25_cqrs-htmx-deep-dive.html §07.

### loginpage-adoption tail (2026-10-25 train — shipped, two candidates)

The library `/login` page, path rename (`/api/auth/webauthn/login/*`),
nonce CSP lane, and `login.css` artifact are SHIPPED (CHANGELOG
[Unreleased]; smoke-verified both modes). Two candidates stay open:

1. **tw.css freshness grep in CI** — SHIPPED 2026-10-28: the advisory
   `tw-freshness` CI job rebuilds both artifacts via
   `scripts/build-tw-css.sh` (flake-locked tailwind, go.mod-pinned module
   sources — byte-stable by construction) and diffs against the
   committed bytes; both artifacts were rebuilt and committed with the
   lane (they had drifted behind the agent-UX + review view trains —
   the lane's first run caught exactly that class). Advisory like
   reproduce-first (the daemon pushes in-flight view edits before a
   train's tail rebuilds); flip to blocking after clean trains. Evidence:
   docs/status/2026-10-25_16-10_loginpage-adoption-build-green.md.
2. **Forced-theme users see the OS theme on `/login`** — the library
   ThemeScript follows `prefers-color-scheme` only; the shell's
   localStorage forced theme does not apply there. Accepted caveat for
   now (documented in README § Passkey sign-in); revisit only if the
   owner notices.

**Priority:** Low · **Effort:** S each.

### Chat networks follow-ups (2026-10-23 train — matrix + signal shipped, session deferred)

Foundation (network axis + sender registry + receivers) and both shippable lanes
landed (`internal/matrix`, `internal/signal`, store v5, config gates, Settings rows;
plan + research evidence:
[docs/planning/2026-10-23_15-40_SUPERB-chat-networks-pareto-plan.html](docs/planning/2026-10-23_15-40_SUPERB-chat-networks-pareto-plan.html)).
Remaining, unscheduled:

- **Session messenger: NO-GO** (researched at source: no Go library anywhere, no
  official headless daemon/bot API; `libsession-util` is C++ config-layer only;
  the community `session.js` is Bun + bus-factor 1). Reopen ONLY on a named
  trigger: (a) an official/foundation headless API, (b) a maintained Go
  implementation, (c) an owner call accepting a research-grade `session.js`
  sidecar (then the stack repo owns the service first). **Priority:** parked ·
  **Effort:** n/a until a trigger fires.
- **Stack deployment work** (nix-international-telephony repo, NOT this one):
  signal-cli systemd unit (nixpkgs 0.14.9, `daemon --socket`, linked-device QR
  bootstrap ritual, 3-month freshness watch) + matrix homeserver choice/config;
  both feed webphone via `settings`/`environmentFile` — no webphone module
  change required. Served-markup changes (network prefix rows) owe the standing
  stack browser E2E on the next relock. **Priority:** Medium · **Effort:** M.
- **Lane v2 fences**: read receipts (matrix `m.read` → delivered) and the
  compose-new-DM affordance are SHIPPED 2026-10-28 (`ParseComposedAddress`
  shape routing + `chatloop.Sink.RemoteRead`; CHANGELOG [Unreleased]). Still
  open: matrix group rooms (threads key (owner, network, remote) one-peer —
  groups need a thread-model change, not a fence lift), matrix E2EE (needs
  olm/megolm — big), signal inbound attachments (signal-cli side plumbing),
  signal groups (same model issue). **Priority:** Low · **Effort:** M each.

### /health operator surface (2026-10-23 — stack option shipped, deploy pending)

`services.telephony.webphone.dashboard` now mounts the go-health dashboard at
`/health` with the nginx location UNBUFFERED and fenced (operator SSO in
`forwardAuth` mode, loopback-only otherwise); pbx-artmann sets it. **Shipped
2026-10-23:** webphone and the stack are pushed, pbx-artmann relocked
(`2fed91b`), and the host evaluates — the `pbx.artmann.tech` vhost renders
`= /health` + `/health/` with `auth_request /_telephony_auth_operator` and
`error_page 401 = @telephony_signin_operator`, and
`services.webphone.settings.dashboard = { enable = true; title = "PBX webphone"; }`.
What remains is the HOST REBUILD plus a browser check that logged-out `/health`
redirects to the SSO gate and a signed-in operator sees the cards.
**Priority:** High · **Effort:** S.

### Catalog adoption SHIPPED locally (2026-10-24 — push pending; supersedes the 2026-10-19 inventory-only record)

webphone is now a full federated source of catalog.home.lan:
`internal/domain/catalog.go` (the four seam events — message.received,
message.delivery.advanced, thread.flag.changed, fax.advanced — service
meta, telephony-domain membership, owner lars), a `catalog` CLI
subcommand (versioned + `-plain` export), catalog/export tests
(`internal/domain/catalog_test.go`), and `.eventcatalog-hub.json`.
The 2026-10-19 "inventory-only, an export would fabricate message
vocabulary" record is superseded by the 2026-10-24 coverage plan: the
seam events are webphone's real emitted vocabulary, documented AS
change notifications in their summaries; NO commands are declared
(the build test asserts zero), so nothing fabricates a CQRS write
model and no cross-service consumer edges render.

- [x] **Catalog adoption pushed + verified** (2026-10-25): `46cc6db`
      (the `catalog` command) is on origin/main. REMAINING LEG IS
      CROSS-REPO: the eventcatalog-hub re-pin + build — the hub CI clones
      webphone@main and runs the export; ritual unchanged (webphone-push
      first, stack re-pin after, tri-repo rule). Owned by the hub/stack
      session, not this repo.

### Wire contracts SHIPPED (2026-10-24 — same catalog universe, the OTHER two lanes)

The openapi/asyncapi lanes of
[docs/planning/2026-10-24_16-50_SUPERB-catalog-contracts-plan.md](docs/planning/2026-10-24_16-50_SUPERB-catalog-contracts-plan.md)
landed the same day (status:
[docs/status/2026-10-24_19-15_catalog-contracts-execution-status.md](docs/status/2026-10-24_19-15_catalog-contracts-execution-status.md)):
`/openapi.json` is GENERATED from the handler DTOs via
go-cqrs-lite/catalog over all eleven session-gated surfaces (the
hand-written 227-line const is gone; drift test + seam pins keep it
honest), and `webphone contracts` renders the machine-to-machine lanes
(five `/hooks/*` types, the `{"provider_ref"}` receipt, five SSE kinds)
as the committed `docs/contracts/asyncapi.yaml`. The split rule
(browser JSON→OpenAPI, machine lanes→AsyncAPI, never both) + the
passkey/SSO conditional-surface note live in
[docs/contracts/README.md](docs/contracts/README.md). This section
records the decision the 16-50 plan's M1 gate asked for: ADOPTED by
ship, extending the eventcatalog supersession above — the three catalog
lanes (eventcatalog federation, OpenAPI, AsyncAPI) now cover the three
real vocabularies (seam events, browser JSON, machine-to-machine).
Remaining tail lives in
[docs/planning/2026-10-24_19-27_SUPERB-contracts-completion.md](docs/planning/2026-10-24_19-27_SUPERB-contracts-completion.md)
(freshness gate, upstream exporter filings, android consumption).
HARVEST 2026-10-24 ~20:05 (the freshness gate + research index + the
three-catalog ruling SHIPPED since that plan was written):

- [x] **File the seven catalog-exporter candidates upstream — FILED
      2026-10-25: go-cqrs-lite #91 (response headers), #92 (204 body-schema
      exemption), #93 (WithParam types), #94 (no fabricated kafka server),
      #95 (multipart bodies), #96 (operationId dots), #97 (OpenAPI 3.1.0).**
      Each premise verified at source against the pinned v4.9.0 (file:line
      cites in the bodies; `caseutil.ToPascal`'s separator set and
      `validate.go`'s 2xx rule quoted), dupe search over issues AND PRs
      clean, drafts voice-checked (0 FAIL / 0 WARN) and kept at
      [docs/drafts/catalog-exporter-2026-10-25/](docs/drafts/catalog-exporter-2026-10-25/).
      Watch the issues; the local workarounds (typed empty objects, the
      Retry-After description text, `WithServer("", "")`) retire when the
      fixes land.
- [x] **metaengine v4.19.0→v4.20.0 bump train — SHIPPED 2026-10-25.**
      Executed per the notes' GO recipe: codec contract read at source
      (`stream_codec.go`: strings as-is / JSON else, `Err` write
      variants; webphone has ZERO direct stream-codec use — the engine
      encodes inside the query declarations), readmodels' Open-asserted
      capabilities held, full suite + race suite + buildflow (exit 0) +
      format/island/contracts-fresh checks + the KVM backup/restore
      drill all green, vendorHash re-pinned via the documented rebuild
      recipe, and the asyncapi.yaml regenerated content-inertly (the
      shared graph's YAML style flipped; every ref/name identical).
      go-cqrs-lite/id rode to v4.7.3 with it.
- [ ] **OWNER-GATED: catalog v4.9.0→v4.10.0 bump** — 213 commits
      (architecture-reference + export-selection features webphone does
      not use; the seven upstream candidates are NOT obviously fixed).
      Lean: ride the next natural train. Same notes file. **Priority:**
      Low · **Effort:** S (decision) / M (bump).
- [ ] **webphone-android consumes the specs** — fetch/parse
      `/openapi.json` in the client bootstrap, derive the push-devices
      client calls, parse the asyncapi ring contract, verify the CSRF-dance
      paths against the spec (completion-plan N14; other repo's gates).
      **Priority:** Medium · **Effort:** M.
- [ ] **OWNER CALL: /openapi.json ETag + golden fixture** — cheap
      client caching / byte-pinned generated spec; recorded, not built
      (a disabled deployment must not promise; an ETag is a product
      decision). **Priority:** Low · **Effort:** S.

### Declarative/DDD model train (2026-10-14 plan — EXECUTED)

**Status:** 🟡 `EXECUTED — trains A/B/C/E DONE 2026-10-14, D DONE 2026-10-18 (island CALL_TRANSITIONS + table-gated hold machine, commit ef29ea9); only owner calls remain` (A: delivery machine, CI green · B: re-adjudicated delivered-by-prior-train, backfills landed — B1.5 resolution-branch fax test, B1.6 P5 annotation, B2.3 README notes · C: AppendToThread + faxMoves/FaxAdvance, C2.5 re-adjudicated against the pinned panel contract (same-status verdicts refresh, dedup stays the seam's job) · E: typed events + `Notifier.Project` projection, memo, docs sweep · D: declarative CALL_TRANSITIONS/applyTransition/viewStateOf in state.js behind the hold machine, sip.js arrivals table-gated with loud refusal warns, 2 new specs) · **Priority:** High · **Effort:** M — full plan + 60-row fine breakdown + execution graph: [docs/planning/2026-10-14_17-15_SUPERB-declarative-ddd-model.md](docs/planning/2026-10-14_17-15_SUPERB-declarative-ddd-model.md).

- **Tier 1% (51%):** DONE — delivery machine as data (`deliveryMoves` + pure `DeliveryAdvance` + `FailureKind` typed; retry literal → `IsRetryable`), byte-parity held.
- **Tier 4% (64%):** DONE (landed with the series-3 train; verified + backfilled this session) — `gateway.Receipt.Resolution` enum, `identities` key + shared-contact validation at `Load`.
- **Tier 20% (80%):** DONE — `domain.AppendToThread` invariant fn; `faxMoves` + `FaxAdvance` (replay semantics re-adjudicated: same-status verdicts are legal refreshing moves; inert class = events against the inbound `received` fact).
- **Tier 100% (partial):** typed domain events + `Notifier.Project` SSE map DONE; DOMAIN_LANGUAGE ratification memo DRAFTED (owner-gated); docs sweep DONE; island `CALL_TRANSITIONS` (D) DONE 2026-10-18 — the stack browser E2E obligation it carried rides the next relock (standing AGENTS rule).
- **Owner calls minted:** A1.7 out-of-order verdict tightening (recommend: keep today's overwrite); E2 ratify-vs-delete DOMAIN_LANGUAGE.md (memo in the file, ROADMAP cross-linked); g3 Train D sequencing (lean: own fresh session + stack browser E2E).

**Evidence:** `docs/reviews/2026-09-30_data-model-review.html` (three open findings); analysis session 2026-10-14.

### SUPERB unified-auth train tails: owner-gated cutover + hardening remainder (2026-10-12 13:45 harvest)

**Status:** 🟡 `PARTIALLY DONE` (T08/T11/T12/T14 closed 2026-10-12; the same-day harvest session executed the 10-52 report's non-gated items — VM suites green with the new asserts; only the rows below remain) · **Priority:** High (cutover cluster) / Medium (tail) · **Effort:** M

- **OWNER-GATED — T13/T15 cutover cluster (do not self-answer; questions in ROADMAP § Open questions):** pbx-artmann relock to stack HEAD (clean tree — narHash covers all; release-runbook dance) → staged prod cutover `telephony.directory.mode = "xml_curl"` + explicit `staticFallbackExtensions` (operator + ring groups; members that cannot webphone-login MUST be listed) → pre-cutover backup + the <2-min rollback drill rehearsed → verify the derivation secret file lands with correct owner/mode for BOTH webphone (DynamicUser read) and the bridge LoadCredential → post-cutover CI + journal watch (`auth logout`, `directory query` lines prove the seam live) → T15 live proofs (derived-credential REGISTER browser+one softphone, logout-revocation live, static extensions unaffected, rollback drill ONCE on prod) → pbx-artmann CHANGELOG entry written IN the cutover commit.
- **Stack hardening tail:** port the bridge answer-count journal + seam-content-assert pattern into the OTHER webphone suites (shared helper in `tests/common.nix`) · shared `secretFile` fixture helper (`user = <service>` pattern) in `tests/common.nix` · consider merging telephony-webphone-dynamic-directory INTO telephony-webphone (one machine, halved CI cost; keep assertions readable) · wire `fs-directory.service` into `operatorWatchedUnits` under xml_curl + consider a health leg joining telephony-health (verified open 2026-10-12, pbx.nix `operatorWatchedUnits` lacks it) · pin the sipcred wire vector into the stack fixture (the VM proves behavior; the cross-repo HMAC vector is untested there) · xml_curl × fail2ban interplay check (digest-failure bans still apply?) · fs-directory journal noise under REGISTER storms (proxy-side rate limit?) · stack-repo one-liner at next touch: name the single-go-mod re-open trigger in the stack's consuming docs (routed from the 10-14 go-modularize §f item 40, 2026-10-18).
- **Webphone/infra tail:** daemon push-stall watchdog LANDED 2026-10-12 (`scripts/push-stall-watch.sh`, exit 1 when the oldest unpushed commit exceeds the 60-min ratified anchor; `--threshold-minutes`/`--no-fetch`; 4 arms self-tested in a scratch repo) · daemon-enforcement hook LANDED 2026-10-12 (`.githooks/commit-msg` blocks daemon-pattern commits touching `go.mod`/`go.sum`/`go.work*`/`flake.lock`/`.buildflow.yml`; `core.hooksPath` now RESOLVES — it had dangled since repo birth; `--no-verify` bypasses for real tool bumps) · **GUARD HOLE triaged 2026-10-15 (M10 μ44, evidence `89cf3b9`): the hook NEVER binds daemon commits** — the pma unit exports `GIT_CONFIG core.hooksPath=/var/empty` (env config overrides repo + global config), so daemon git runs hookless; hook logic itself VERIFIED (temp-index repro refuses exit 1; `scripts/commit-msg-guard-selftest.sh` pins 3 cases). FIX = owner edit of the nix-config pma module (drop the GIT_CONFIG_* override; fleet-wide call — hooks then bind the daemon in ALL projects). Interim exposure: daemon can still land go.mod/.buildflow.yml — watch daemon commits touching guarded files. · smoke-harness vacuous-pass guard: VERIFIED CLEAN 2026-10-12 (bool() guards at every wire compare in webphone-smoke.py; no double-extraction compares shipped — apply the rule when a credential-equality leg lands).
- Closed 2026-10-12 by the harvest session (rows deliberately NOT minted): VM suites value-pins (`sipCredential:true`, auth-logout audit assert, static auth-register leg), `apps.check-vm` sequential VM driver, FS module-mirror eval check + ONE-home list, ci-posture ledger, CI-registration investigation (`25cf08d` HAS run 38211100788 — the 10-52 claim was stale; `c2caead` was simply unpushed), README/error-contract/lessons/runbook docs, wsprobe exit-code docs, AGENTS tri-repo FS line. The /tmp passkey smoke script evaporated with /tmp — its arms live in the VM suite + go tests now; recreate only on demand.

**Evidence:** `docs/status/2026-10-12_10-52_SUPERB-train-completion-status.md` (f)+§g; stack runbook § Dynamic SIP directory; session VM run log (telephony-webphone + dynamic-directory green with the new asserts, 2026-10-12 ~13:35).

### Dedup/branching-flow sitting (2026-10-11 ratifications — execution tail)

**Status:** 🟡 `PARTIALLY DONE` (all five owner calls answered + recorded in registry/AGENTS; executions below) · **Priority:** Medium · **Effort:** M

1. **settingsRow ADOPTION** (owner: BUILD → resolved to ADOPT) — EXECUTED 2026-10-12: settings.templ renders `display.DefinitionList` (Class `wp-settings`), tw.css `@source` rebuilt, old-vs-new render diff 6/7 panels byte-identical (radius literals synced to the redesign train's 9/12px), registry ruling #2 annotated EXECUTED. templ-components ALREADY ships the component, so no new one anywhere.
2. **branching-flow upstream filings** (owner: file): (a) phantom-analysis class scoping/config — 13/14 rows this session were non-identity plumbing, per-finding nolint + baseline exist but no class-level knob; (b) panic-analysis FP precision — 3 guard shapes (closure nil-guard, map-to-slice invariant, i<j two-pointer) produced 6/8 false rows. Drafts land in `docs/drafts/`, evidence: docs/status/2026-10-11_02-19 triage doc.
3. **Re-triage the branching-flow skip** — EXECUTED 2026-10-11 (stale premise retired): the tool runs UNSKIPPED (19 baselined findings with in-source rationales; the SDK severity cap keeps new findings non-gating until the error_rules-bearing BuildFlow release); `skip_steps` now holds only pytest-test (verified 2026-10-25).
4. **art-dupl `-t 1` + baseline ritual** — EXECUTED 2026-10-12: first deep sweep 867→58 actionable groups, ZERO harmful; `.art-dupl-baseline.json` recorded (tool v0.7.2) and CI gained the new-only `dup-gate` job; sweep-log line in the registry. Unfixable detector FPs: none surfaced this sweep (upstream-filing duty stays with future sweeps).
5. **Daemon dep-lane move** (owner: MOVE): configure pma to stop bumping go.mod/go.sum (owner-side daemon config); bumps ride `buildflow update`/Dependabot with the tidy → re-pin → CI-watch tail (AGENTS Conventions row added).
6. **Quarterly baseline/suppressed-set audit** — first 2027-01-11 (registry Protocol step 1).

**Evidence:** registry 2026-10-11 entries + ratified-rulings section; status docs `2026-10-11_02-19` (triage) + `2026-10-11_02-48` (session report); AGENTS.md dedup/nolint/dep-lane/INTERIM edits

### ASR/SUPERB train follow-ups (2026-10-08 harvest: 16-54/18-15 §f + plan remainder)

**Status:** 🟡 `PARTIALLY DONE — hardening + polish + test debt CLOSED 2026-10-18; #13 affordance + #21 whisper.cpp variant SHIPPED 2026-10-25; remains: #48 privacy call, owner gates` · **Priority:** Medium · **Effort:** S-M

SHIPPED by the train (see CHANGELOG `[Unreleased]`): transcript persistence with per-extension + per-call caps and the near-silence RMS guard (M12), History transcript search + hint (M15), copy-to-clipboard on card/History/chips (M16), export-zip transcripts leg (M17), Settings provider kind·model row (M18), queued-hold + concurrency + cap-speedup specs (M19/M20), single-pass `Recent` (M21), confidence marker REJECTED on provider evidence (M22, decisions doc), ui-capture ASR-on surfaces (M13, 18-shot matrix), buildflow-at-train-start lesson + AGENTS rule + `.buildflow.yml` load-bearing warning (M14). REMAINING, grouped:

- **OWNER-gated (14-08 report §g — do not self-answer):** M1 first-live provider + retention/autoplay ratification; M2 Speaches deployment on the consuming stack (stack-repo work); M4 live-reality transcription (throwaway GCP key or Speaches instance); M6 stack browser E2E re-run (served markup changed twice on 10-08); 429 cooldown Retry-After vs fixed 10 s; ship `SILENCE_RMS` 0.004 as-is.
- **Provider hardening (16-54 §f):** CLOSED 2026-10-18 — retry/backoff + "transcription unavailable" toast (#43/#44) VERIFIED ALREADY-SHIPPED; provider latency metric (#23) landed as the rolling outcomes tracker + debug line; optional non-critical ASR health leg (#22) landed (recent-outcomes verdict, never probes); memory ceiling (#47) landed as the /api/transcribe concurrency ceiling (4 in flight, 503 past a short queue wait). #21 `/inference` whisper.cpp native variant SHIPPED 2026-10-25 (`asr.provider = "whispercpp"`: same multipart + {"text"} wire, /inference path, kind-aware Settings row).
- **Client polish (16-54 §f):** CLOSED except #48 — language hint (#11/#12), DOM cap (#41), per-call clear affordance (#39), transcript download (#40) landed (round-5 E16); "saved to History" discoverability (18-15 #13) SHIPPED 2026-10-25: the first server-confirmed segment save reveals a persistent saved-count line on the call card (en/de), announced once per call; REMAINS: remote-leg-only transcription default — privacy call (#48, owner).
- **Test debt (16-54 §f):** ALL CLOSED — 2026-10-12 batch (#32/#33/#36/#38/#34) plus 2026-10-18: voicemail panels transcribe-button test (#37) VERIFIED PRE-SHIPPED · 429 combined live+auto load (18-15 #5) pinned by the flood spec (transcribe + transcripts share one bucket).
- **Docs/smoke:** CLOSED 2026-10-12: NixOS module README `asr.*` options (#26 — already shipped by the google-wire train; all 7 rows verified against `config.ASR` fields) · error-contract ASR freshness (VERIFIED: all 8 `asr.*` codes in code match the table; `TestErrorCodeRegistryIsFresh` enforces it continuously) · smoke ASR leg with a fake provider (already shipped as `asr_scenario` in webphone-smoke.py — openai + google wires, wired into main) · "audio leaves the box" flag note (#49, README § Live transcription) · near-live segmented-design note (#50, same section).

**Evidence:** plan `docs/planning/2026-10-08_07-36_SUPERB-asr-provider-rollout.md`; decisions `docs/planning/2026-10-08_14-05_SUPERB-asr-decisions-and-spikes.md`; pause report `docs/status/2026-10-08_14-08_SUPERB-asr-train-m12-m25-execution-report.md` §g; source §f lists in `docs/status/2026-10-07_16-54_*` and `2026-10-07_18-15_*`

### Gate-recovery session tail (2026-10-06 22:08 report)

**Status:** 🟡 `PARTIALLY DONE` (critical legs closed 2026-10-06 ~22:45; tails remain) · **Priority:** Medium · **Effort:** S-M

CLOSED this session: the red CI (vendorHash `WYqTip91…` stale after the post-train tidy — re-pinned `5ekZFK2m…`, CI run 37526835020 green) · `TestSQLiteSessionTTLExpiryAndSweep` hardened with an injectable store clock (both session stores; race structurally impossible, suite + buildflow green) · fresh-binary smoke boot (47+4+8, 0 failed) · buildflow verdict rc=0 · AGENTS lessons (kill builtin trap, vendorHash `--rebuild` recipe + tidy-then-repin sequencing, templ-components v1.20.0 poisoned require). REMAINING from the report's (f) list: ui-capture 14-shot visual pass on the v1.20.1 tree (f6 — re-run 2026-10-18 on the round-5 tree, see the round-5 tail row) · vulnix `0/5 deterministic-retry` warning INVESTIGATED-CLOSED 2026-10-12 (f12: the string exists nowhere in this repo — it is vulnix-the-tool's internal eval-recovery message; NOT reproducible at HEAD, two consecutive `nix run .#vulnix` runs byte-identical — 8 derivations, glibc range-match noise triaged distro-patched, zero real advisories, no retry line; the gate's verdict rides the fixture-checked `webphone-vulnix-triage` CLI, never vulnix's raw exit) · GOPROXY fallback chain + BuildFlow retry-wrapper (f9/f10, BuildFlow-repo/crush-config side — the one remainder, F86) · module-cache janitor + leftover zero-byte `.tmp` sweep CLOSED 2026-10-18 (`scripts/module-cache-janitor.sh`: report-only default, `--sweep` deletes only zero-byte + age-floored files, self-tested; 24 stale zero-byte .tmp swept live) · upstream releases: templ-components errorpage tagging discipline CLOSED-UPSTREAM 2026-10-18 (verified: issue #27 + docs/release-checklist.md + the release-smoke workflow already encode the discipline — no duplicate filed). Stack browser E2E for the dep train (f8) rides the next stack session. 23:22 §f remainder folded 2026-10-06 (docs-health v7); closed-later-that-night (round-5 plan execution): whitespace-drift `.nix` gap + self-test (r1 — `.nix` matcher already landed; scratch-repo self-test: whitespace-only `.nix` staged edit exits 1, content edit exits 0) · the two owed lessons stories — outage timeline + `/go.mod` h1 ≠ sha256 oracle (r2, docs/lessons.md Tooling traps + Nix) · CI-flake ledger (r6, docs/ci-flake-ledger.md; AGENTS links it) · v1.20.1 greppable-class byte check `wp-thread-row`/`wp-bubble` (r16 — served-payload probe on a fresh loopback binary: both classes present in list + transcript) · DOM-contract v1.20.1 reasoning note (r17, dom-contract.md § Notes) · run-37496169265 flake-window closure (r30 — ledger records "hardening supersedes confirmation" as the OR-arm closure) · v7-delta hygiene (codespell clean; markdownlint baseline: default MD013@80 is noise, no repo config — one real MD018 hit in lessons.md fixed). REMAINING: stale BuildFlow binary rebuild (r10, BuildFlow-repo side; folds with f9/f10) · buildflow doctor gate in the release-runbook prelude + `timings --regressions` rebaseline (r22/r23) · fleet wedged-go audit + /tmp probe-artifact prune + gopls workspace close (r19/r20/r21; local wchan scan 2026-10-06 ~00:10: one locks_lock_inode_wait hit = a concurrent monitor365 cargo check, left alone per concurrent-session doctrine) · quiesced-host local flake check incl. the KVM backup leg (r15 — DEFERRED: host not quiesced, a concurrent session's `go test -race` suites were running; re-run when idle) · kill-builtin lesson empirical verify at the next natural kill (r14) · go-cqrs-lite release-tooling audit (r13, upstream) · owner legs routed to briefing rows 36–37: devShells dprint/prettier/ruff/lychee additions + lychee auth-vs-exclude (r25/r26). Dropped as landed: r27 (the go get → tidy → vendor → repin chain IS the AGENTS vendorHash bullet).

**Evidence:** report `docs/status/2026-10-06_22-08_buildflow-gate-recovery-go-module-train.md` (§f rows); CI `gh run list` 37526835020

### Passkey train tail: deploy proof (owner terminal)

**Status:** 🟡 `IN_PROGRESS` (code polish DONE 2026-10-04; owner legs remain) · **Priority:** High · **Effort:** S

The code train AND the 3-repo deploy train are DONE (webphone `400eaff`
pushed; sibling `890a526`; pbx-artmann staged `0ngvm4q7…`, probe OK,
gates green — see CHANGELOG Unreleased + the 17-28 report); AUTH AUDIT 2026-10-06 (13:10): the passkey HTTP surface was code-read audited; one accepted-risk
observation needs an owner verdict — `POST /api/auth/passkey/enroll/begin` gates on the account
ULID alone (any well-formed usermgmt UserID drives a registration ceremony; the enrollment TOKEN
is only enforced one step earlier at /verify). Protection = ULID unguessability (128-bit, never
exposed except to valid-token holders) + passkeyLimiter + mode default-OFF. Options: bind the
ceremony to a token-verified server-side session (usermgmt upstream change) or ratify ULID-secrecy
as the standing posture. Companion reconfirmation: session-row password stays plaintext at rest
(spike-verdict ratified; 0700 DB + UMask 0077). Same session FIXED: session-cookie Secure behind
the TLS proxy, real-config CSRF rotation, constant-time webhook secret (CHANGELOG § Security).
AUTH TAIL 2026-10-06 (D18): renewal Secure re-issue pinned end-to-end (incl. the half-life
throttle + same-token assertions; `62bc7dc`); slog secret-leak grep over server/userauth/gateway
CLEAN (attrs are errors/status/family/code/extension only; vendored usermgmt logs the cookie NAME
and bot id, never values; enroll-token errors are static strings; gateway logs nothing); codespell
over the CHANGELOG/lessons/status/planning delta EXIT 0 — the four `keep-alives` hits are the
twice-adjudicated plural noun, now durably in `.codespellrc` ignore-words. Full-repo sweep close
(b.3/f.6) 2026-10-06 15:50: the five unswept packages are CLEAN — app (one `log.Info` "timezone
applied" zone attr), pbx + web/views (ZERO log calls), fax (6 slog.Warn: error + fax_id),
messaging (3 slog.Warn: error + message id + owner extension + remote number — CDR-class data,
the accepted 18.2 bar); no print/os.Stderr paths anywhere in the five. ISLAND AUDIT 2026-10-06 (D19): clean — adoption ladder correct
(3 backoff retries, reload last resort); zero raw innerHTML sinks (whoami/snippet/lightbox/errors
all textContent/value/src); zero credential persistence in localStorage; legacy pbx-contacts
import deletes only after EVERY row is server-accepted, else retries next login; no
password/token in any log line. FIXED: authedFetch now imports the csrf.js meta reader (ONE-home
rule; it had drifted to an inline querySelector). ACCEPTED: passkey/enroll finish carry the
one-time ceremony session_key as ?user_id= (burned at finish, worthless without the WebAuthn
assertion + CSRF; exposure = proxy access logs only) — moving it into the POST body is a cheap
option inside D30's decision space.
the code-polish tail landed 2026-10-04 (enrollFailed(0)→enrollNetFailed,
csrf.js single home, userauth /healthz leg, tier-2 pins, whoamiLine→ui.js,
typed sessionIdentity, plan-doc annotation, AGENTS trim — see CHANGELOG).
Remaining: the OWNER switch + post-switch ritual (rotate
`/tmp/pbx-toplevel-current`, record the fresh diff-closures baseline)
and the LIVE passkey proof with the fail-closed password-file drill
(pbx-artmann TODO §1 row). OWNER CALLS outstanding: ratify
runbook-only enroll + Lars-only v1 mapping; switch-now-vs-CI-first
(VERDICT 2026-10-05: every recent stack `main` CI run cancels at the 1h
ceiling inside `nix flake check` — aarch64 VM leg green 6m16s, browser
E2E on-demand never fires; the mod_enum/timeout diagnosis precedes any
lock bump); installer release republish timing (stale since the
relock). HANDOFF NOTE: RESOLVED 2026-10-25 — LSP diagnostics on
`internal/server/passkey_api.go` are clean at HEAD (the note's owning
session is long gone; verified by this sweep).

**Evidence:** status `docs/status/2026-10-04_17-28_passkey-train-resumed-session-brutal-status.md` (§f harvest source); plan `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`

### Boot-contract tail: D3 retry-loop owner call + stack runbook patch

**Status:** 🟡 `PLANNED` (owner call + stack dispatch) · **Priority:** Medium · **Effort:** S

The boot-error contract SHIPPED 2026-10-02 (5-part render on every boot failure, exit 1 designed / 2 panic, EN-only, class-tagged; `App.Start` de-panicked; arch test confines samber/do; smoke `boot failure scenario` green). The stack runbook patch APPLIED 2026-10-04 (stack `9a21893`: boot-surface + passkey error-contract rows; the parked patch doc is spent). Remaining: (1) OWNER CALL (D3 follow-up) — cap systemd's boot-failure retry loop (`StartLimitBurst`/`StartLimitIntervalSec`) or ratify the 5s `Restart=on-failure` retry as desirable liveness (module deliberately unchanged this train). The boot-surface re-grade landed EARLY 2026-10-05 (contract table vs code: 7 classes + drift pins intact; see the erraudit line in Standing watches)

**Evidence:** plan `docs/planning/2026-10-02_10-23_SUPERB-operator-boot-contract.md`; ruling `docs/error-contract.md` § "Boot surface"; pins `cmd/webphone/bootreport_test.go`; runbook patch = stack `nix-international-telephony@9a21893`

### v2.8.0 deploy tail

**Status:** 🟡 `PARTIALLY DONE` (owner terminal remains) · **Priority:** High · **Effort:** S-M

v2.8.0 deploy tail (OWNER terminal — pbx-artmann AGENTS forbids assistant ssh/deploy): the RELEASE SIDE IS DONE 2026-10-01 — v2.8.0 signed+pushed (number decided autonomously: 2.8.0 so the bridge ">= 2.8" claims stay true; ships the never-tagged 2.7.0 content + nix-review train + honest Content-Type + error-family + caddy module rename). REMAINING (all owner): deploy (stack lock bump → stack gates incl. the browser E2E the EmptyState markup change owes → aarch64 → pbx-artmann relock #5 + re-pin → deploy; COMMAND: `cd ~/projects/pbx-artmann && nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech` → smoke → `nixos-rebuild switch`), then post-deploy `python3 scripts/webphone-smoke.py --base https://pbx.artmann.tech --expect-version <V>`. NOTE: train C refuses self-sends LOCALLY — expect the 422 "messages cannot be sent to your own number" + failed row instead of the provider's 40310 text; the bridge's declared-type preference reaches prod only after the stack relock that rides this tail. If the release number is NOT 2.8.0, reword the bridge-side ">= 2.8" doc claims there to commit/date form (webphone's own AGENTS line is already commit-pinned at `e6ea2c7`). AT THE STACK RELOCK (same train — both stack lines reference the REMOVED nginx option and eval-error against the new rev): `modules/telephony/web.nix` `nginx.enable = lib.mkForce false;` → `caddy.enable = lib.mkForce false;` (comment wording too) and `modules/telephony/default.nix` vhost split-brain assertion `config.services.webphone.nginx.enable` → `caddy.enable` + message wording (webphone caddy rename, 2026-09-30)

**Evidence:** tag `v2.8.0` pushed + gh object published 2026-10-01; stack `3afcf57` (lock bump + caddy guard, E2E×1 green after one self-inflicted load flake + VM + flake check green); aarch64 ELF `b700`; lychee 0 after repointing the never-tagged v2.7.0 links; number = 2.8.0 (bridge ">= 2.8" claims true; ratify at the sitting — briefing row 15)

### Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error)

**Status:** 🔴 `TODO` (owner journal leg) · **Priority:** High · **Effort:** S

Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error) when probed 2026-09-19 — webphone-side classification is 502 + provider-rejection detail since `1d53f44`, refusals answer 422 since `6ac8962` (`1d53f44` live on prod since the 2.6.0 deploy; `6ac8962` rides the untagged release), but the ROOT cause is stack-side (telnyx-webhooks bridge down/rejecting or Telnyx creds). OWNER (assistant ssh is forbidden by pbx-artmann AGENTS): journal the telnyx-webhooks unit, grep for sms/422/error lines, restart or fix creds per findings, send a test SMS; record the root cause here + in the stack runbook. Also still unproven live: the 2.6.0 post-deploy sign-in rejection-banner (self-send) browser check — narrows to a journal-only confirmation if a human test SMS worked since 2026-09-25

**Evidence:** 2026-09-19 probe; webphone-side fix `1d53f44`; refusal-422 `6ac8962`; OWNER TRIAGE PACK ready 2026-10-01: command sheet §4 decision tree (unit status → journal grep → the bridge's own actionable cred strings → restart → test SMS; self-send 422 = expected, not a failure)

### Scheduled sends (M21.6): NO-GO on the current gateway seam

**Status:** ⛔ `DECIDED-AGAINST` (gateway-seam dependency; revisit only with a deferred-send seam) · **Priority:** Low · **Effort:** M

Deferred/scheduled message sending was cut from T18 (M21.6): the gateway seam has NO deferred-send concept (loopback sends immediately; the webhook bridge is fire-and-forget to the provider), so a scheduled-send UI would show a state the transport cannot honor — violating the island-honesty contract. Revisit only if the provider/stack grows a deferred-send lane; the snippets seam (M21) already covers the "reply faster" problem it was meant to solve.

**Evidence:** design note `docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md` (M21.6 cut); owner ratification pending (2026-10-02 §g).

### Gateway honest-Content-Type follow-ups

**Status:** 🟡 `PLANNED` (stack/pbx side only now) · **Priority:** Medium · **Effort:** S

Gateway honest-Content-Type follow-ups (webphone side of `docs/status/archived/2026-09-29_04-32_*` §f): byte-exact part-HEADER-block golden (current golden greps fields individually; a full header-block pin catches reordering) · document the webphone×bridge compat matrix next to the AGENTS gateway-seam bullet (old binary → sniff lane; new → declared lane) · consider a webhook-mode smoke probe (loopback-only today; the webhook lane has zero black-box coverage in this repo). Stack/pbx side (NOT this repo): `ftypqt`→`video/quicktime` sniff fix, stack E2E MMS-outbound coverage, pbx-artmann FEATURES:87 stale sniff text — route via the next stack session/relock

**Evidence:** webphone side DONE `0aab677`: full-body golden `TestProviderFormPartHeaderBlockGolden` (field order, header order, CRLF framing), AGENTS compat matrix, webhook smoke probe DECIDED-AGAINST (golden + bridge contract tests own the shape). Still routed via next stack session: `ftypqt` sniff fix, E2E MMS-outbound, pbx-artmann FEATURES:87 text

### Tooling hygiene batch: markdown-lint posture decision

**Status:** 🟡 `PARTIALLY DONE` 2026-10-01 · **Priority:** Low · **Effort:** S

Tooling hygiene batch: markdown-lint posture decision (5092 detect-only corpus findings — configure markdownlint to house style or record the detect-only posture in AGENTS so nobody "fixes" the corpus by reflowing) · codespell policy for the 2 `pre-emptive` warnings in archived status snapshots (exclude `docs/status/**` like the stack, or fix the words) · reconcile AGENTS buildflow-full claim vs observed skips (gitleaks/codespell/markdown-lint "skipped by build mode 'full'" on 2026-09-26 — AGENTS or `.buildflow.yml` is wrong) · BuildFlow binary freshness advisory (stale binary skews verdicts — rebuild in the BuildFlow repo) · commit the reusable render-diff script (partials-only, webhook-502 error instrument — mechanics in the archived 16:56 report §b) + park its recipe in AGENTS/lessons (carried since the 09-24 18:05 report)

**Evidence:** DONE: codespell real-binary + `.codespellrc` (vendor/go.sum skipped, pre-emptive ignored), AGENTS buildflow-claim reconcile, render-diff script committed (`scripts/render-diff.py`; LIVE-VERIFIED 2026-10-01: 7/7 partials byte-identical + 7/7 error arm), BuildFlow freshness noted via doctor. v6 stragglers DONE 2026-10-04: health-css fmt appended (live-verified: raw rebuild → in-script fmt → byte-identical with HEAD), smoke mypy 8→0 (annotations were the lie; ruff clean), "error families are total" + 104-code registry table in error-contract.md generated and freshness-pinned by `TestErrorCodeRegistryIsFresh` (negative-tested; fixes the SkipAll-walk latent bug in the sibling arch test too). REMAINING (owner): markdownlint posture (briefing row 18) — invocation SETTLED 2026-10-05: `nix run nixpkgs#markdownlint-cli2 -- <docs...>` (v0.23.3, no config = library defaults; baseline over the three 10-05 session docs = 172 findings, 170 MD013 line-length + 2 MD026 heading-punct, all cosmetic — detect-only until the row-18 posture rules otherwise). Round-3 C2 gate-debt close 2026-10-06 03:15 (01-40 report addendum has the detail): buildflow EXIT 0 over `43091cc` (incl. `2b52fbf`), island JS 186/186 pass, on-demand codespell EXIT 0 (raw scan: one pre-existing warning-level `keep-alives` in CHANGELOG prose — not session debt). Round-3 C10 daemon zero-churn observation 2026-10-06 03:20: five daemon commits since the `2b52fbf` regen (`e9caa75`, `43091cc`, `be4d5e0`, `b245b73`, `e62fe34`) left `docs/error-contract.md` untouched — md5 stable at `d25d4a50…`; the churn kill now has observational backing, not just the 130-row byte-verification (canary stays open: any future daemon commit touching the file must show a zero reflow diff). C9 landed same session: writer framing + uniform-row-length pins in `rewriteRegistryBlock`, 00-52 report table re-flowed. Round-4 D25 hygiene 2026-10-06: near-aligned-table
sweep over every 10-05/10-06 status+planning doc found exactly ONE mixed-shape table (the 03-40
report) — normalized to the daemon-canonical compact shape; the detector (mixed column widths +
extra cell padding) now reports ZERO across the corpus. `scripts/whitespace-drift.sh` landed:
staged-diff `-w` divergence check over formatter-OWNED files only (Go/JS/MJS/CSS/Nix; md
deliberately excluded — unowned means no drift definition), self-tested in a scratch repo
(whitespace-only Go edit EXIT 1 with the `nix fmt` fix, substantive EXIT 0, md-only "unowned"),
wired as an AGENTS Commands habit line — kills the e62fe34 class at add-time. D18.3 durability:
`keep-alives` joined `.codespellrc` ignore-words after its second adjudication. d.1 fix 2026-10-06
15:40 (hardened twice by 16:45): the detector is now a COMMITTED instrument —
`scripts/md-table-shape.py` (self-test 8/8: mixed flags, compact clean, fully-aligned clean =
D1.4's question, unbalanced-padding flags, mid-cell wrapped rows clean, prose-after-table never
glues, escaped pipes never split a cell, display-width-aligned emoji columns clean); default
scope docs/status/ live = ZERO; explicit rerun over the exact 10-05/10-06 status+planning sweep
set = ZERO, reproducing the claim from the repo. v1 miscounted wrapped tables (fragmented rows)
and escaped pipes; v3/v4 were caught by dogfooding against the session's own report and the
daemon's own output (e.3 rule) — v4 measures DISPLAY width because the daemon's aligner pads
emoji to terminal cells, not codepoints. Full-corpus run 16:45: ZERO findings (v3's 25 shrank to
1 real finding after the daemon's 16:18 aligning pass + display-width fix; the samber-do
scorecard's Status column was normalized to compact on the spot).
f.7 canary FIRED 2026-10-06 16:18 (`2fbdfd7`): the daemon's reflow pass touched the normalized
03-40 table and REFLOWED it compact → fully-aligned (content-inert, no mix introduced; detector
ZERO post-pass). VERDICT INVERTED for D1.4: the daemon IS the aligner — near-aligned mixes come
from sessions appending compact rows onto daemon-aligned tables (the briefing rows-29–34
pattern), and the daemon normalizes them on its next md pass. The same daemon pass reformatted
`scripts/md-table-shape.py` (style-only; self-test still 8/8) and bumped `flake.lock` nixpkgs
(494ce7fd → 151fa4e8 — rides the next push; CI judges).
Was: 09-26 §d3/§f11-12; 09-24 12:41 §f19-20; 04:32 §f14-15

### OWNER-calls batch session

**Status:** 🟡 `PLANNED` (briefing ready, sitting owed) · **Priority:** High · **Effort:** S

OWNER-calls batch session (decisions, one sitting — briefing at `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`): the original ~15 + CRM policy trio + self-send train-C semantics + templ-components history-blemish + tail items g1 force-push ratification / release.sh load-gate default / g2 KVM-timeout / art-dupl `-t 3` baseline (FIFTH surfacing) + registry ratification (`docs/dedup-registry.md` as THE home) + `settingsRow` build-or-retire (FOURTH surfacing) + suppression-bucket doc + dedup suppression-SCOPE ruling (shown-only vs periodic suppressed-set audit; 12:59 report g3, unrouted until v6) + webhook 400 body-text dependents + `msg/`→`message/` idem-key rename + helper micro-test bar + missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy. NEW since 2026-10-06 (briefing rows 36–38): the `auth:`-prefix × daemon-sweep interface (manual `auth:` commit before the sweep window vs accept a daemon `chore:` commit when the pinned session-behavior suite provably ran — 23:22 g2) · devShells.default tool additions (dprint/prettier/ruff/lychee, silences 4 buildflow warnings) + lychee authenticate-vs-exclude policy (22:08 f13/f14) · docs-only express-push ratification (23:22 g3). NEW since 2026-10-05 (briefing rows 29–32, addendum in the briefing doc): mixin/dupe detector-policy home (registry sweep-log only vs AGENTS promotion) · `config.CRM`/`Paperless` URL+Token twin standing ruling (resurfaced by TWO detectors) · `vcard.Card`/`SharedContact`/`apiSharedContact` triad boundary-layer rejection ratification · `errorfamily.Join` gap (add upstream constructor vs keep 4 nolint aggregate sites). NEW since 2026-10-02 (docs-health v6 sweep): ratify "routed-as-resolved → archive" for the four live owner-gated plans (error-excellence, 20-year-durability, 04-29 pareto, 13-33 stack-adoption — v4's g1); ratify train-level `v` verdict markers as the standing annotation hash bar (v3's g3, six sweeps running); the check-rows marker-cell baseline for the 37 pre-2026-09-18 archived files (restyle or accept). NEW since 2026-09-26: push-lag threshold policy (when is a silent daemon push stall BROKEN — 10 min? 1 h) + session push ratification (phase-boundary hand-pushes) · is the next webphone release 2.8.0 (making the bridge's ">= 2.8" claims true) or reword them · `gateway.attachment_limit` knob vs bridge-422-teaches design as final · sniff-fallback lifespan (keep forever vs delete-after-deploy-confirmed) · tier-2 family-adoption intent (dep swept-but-unused?) · report erraudit `tree` same-name-dedupe upstream? (owner's own tool) · AGENTS restructure permission · existing-prod-data chmod/re-backup for the UMask tightening · destDir nesting legality · stack verification now-vs-train-close

**Evidence:** briefing at 38 rows (15–28 added 2026-10-01; 29–32 on 10-05; 33–35 on 10-06 by the round-2 sessions; 36–38 on 10-06 by docs-health v7) + a closed-since section (go-cqrs-lite wontfix etc.)

### Post the release announcements

**Status:** 🟡 `PLANNED` (owner posts) · **Priority:** Low · **Effort:** S

Post the release announcements (drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 at `docs/announcements/`; owner picks channel(s), approves wording, decides the disclosure posture — fix-acknowledged, no exploit detail)

**Evidence:** v2.8.0 drafts A/B/C = `docs/announcements/2026-09-30_v2-8-0_drafts.md` (covers the merged 2.7.0+2.8.0 story); v2.7.0 drafts annotated superseded, links repointed; gh object for v2.8.0 verified published 2026-10-01

### Standing watches

**Status:** 🟡 `PLANNED` · **Priority:** Low · **Effort:** S

Standing watches — QUARTERLY RE-CHECK, next due 2026-12-20 (named triggers fire earlier): sip.js 0.22 (swap ONLY on Chromium WebRTC breakage, a sip.js security advisory, or a needed capability — re-checked 2026-09-23: npm latest still 0.21.2) · templ-components upstream v1.20.x watch (v1.19.4 now adopted in go.mod — daemon sweep; tw.css verified class-identical across the 1.19.2→1.19.4 ride 2026-09-30, no regen needed; next regen only when an adopted component's class set actually changes) · oxlint globals (any new browser global → island oxlint.json) · E2E wall-time budget 445s — greens 195s/184s with FOUC aboard; watch TWO consecutive over-budget runs · MONTHLY erraudit tier-2 re-measure next due 2026-11-05 (RE-MEASURED EARLY 2026-10-05: tier-1 0, tier-2a 0, tier-2b 0 — the passkey train (10-01→10-04) had regressed tier-2 with 7 findings post-09-30-pin: 2 bare constructors at enroll-boot sites, 4 errors.Join shutdown aggregates [errorfamily has no Join — briefing row 32], 1 bool-blank FP where `Duplicate()` returns (int64, bool, bool); all suppressed with reasoned nolints, boot-surface re-grade intact; prior early measure 2026-10-01 with one context_loss FP) · QMD `get`/`multi_get` render bug — blocked UPSTREAM on crush `#3846` (MCP embedded-resource content rendered as Go struct pointer; qmd 2.8.3 is spec-compliant; workaround = `query` tool + disk reads at the corpus root; re-test `get` after the next Crush upgrade; SUBSCRIBE attempt 2026-10-05 blocked — gh token lacks the `notifications` scope; owner one-liner: `gh auth refresh -s notifications` then the `updateSubscription` GraphQL mutation on the issue's node id, or click Subscribe in the web UI) · go-health M18 park watch: revisit ONLY if go-health hooks startup evaluations or cqrs-htmx grows a readiness hook (hook fires only inside `Probe.Evaluate`, which this wiring never calls — verified v0.3.0 AND v0.4.0, wiring reverted `f7028b4`) · branded-id Valuer/Scanner stays PARKED at the store seam (adopt on next storage-format touch) · sessions credentials-at-rest periodic re-review: next due 2026-12-30 OR on the next session-storage change (the 2026-10-06 auth train's changes shipped WITH the 13:10 audit's at-rest reconfirmation — posture stands; baseline `docs/reviews/2026-09-30_sessions-credentials-at-rest-review.md`) · go-cqrs-lite upstream filings 2026-10-15: #71 (TypedReader.Scan truncation signal) + #72 (keep-newest-N cap primitive) — WATCH both until addressed; when the library lands them, re-visit webphone's converging-loop trims + enroll sweep as the named trigger · QUARTERLY modularize re-trigger re-review next due ~2027-01 (fourth series entry 2026-10-17; detection commands live in `docs/modularization/2026-10-17_PROPOSAL.html` §09; the named reopen triggers fire earlier than any calendar)

**Evidence:** AGENTS erraudit bar + templ-components bullet; re-checked 2026-09-23/24 during the pareto cycle (C22/M22)

### internal/server god-package carve

**Status:** 🟢 `DONE — executed 2026-10-17 (round-5 E7 train): server/hooks + server/api landed, adjudicated closed at two phases` · **Priority:** Medium · **Effort:** M

internal/server god-package carve — TRIGGER: the next file added to internal/server (26 files / 4859 non-test LOC + 8630 test LOC re-measured 2026-10-14, two independent count methods; the 09-30 baseline read 19 files / 3268 LOC / 6073 test LOC / 12-of-15 sibling imports / most-touched package since 09-23 at 51 file-events; exported surface only 29 doc lines, so it is an OPAQUE composition surface, not a god module — but it sits ON the >1000-LOC/>20-file review threshold).
**EXECUTED (2026-10-17, round-5 plan E7/E8):** phase 1 `internal/server/hooks` (webhook ingest + idem gates; `hooks.DedupWindow`/`Gate`/`BearerGate`/`RetryAfterHint` exported; app/dedup.go adapter + sweeper retargeted; 401-writer contract test scans the new package; error-code registry follows `webhook.fax_pages`) and phase 2 `internal/server/api` (the five TRUE leaves: csrf, contacts, calls, transcribe, transcripts — moved verbatim under a narrow `api.Server` seam: `RequireSession` + `NotifyContactsChanged` exported on the wiring hub so the 401 wording and the SSE push keep one home; `api.ContactSaveFailed` shared with the tab handlers; contacts fuzz + recordCallIdem contract tests moved with their code). Verification: full server/api/hooks/app/arch suites green, session-behavior suite green, live smoke 47+4+8+9/0-failed against the carved binary, branching-flow baseline refreshed for moved paths with the `api.Server` IFACE_COMPLETE rationale in-source.
**Phase-3 ADJUDICATION (session_api STAYS):** the round-5 plan ordered the session carve last under an `auth:` prefix, but the dependency map vetoes it — `verifyDirectoryCredentials`/`mintSession`/`sessionIdentityResponse` are the shared auth spine of BOTH login modes (passkey_api reuses all three), and the handlers also ride wiring-private state (cookies, csrfCfg); moving the three HTTP handlers alone would split the login subsystem across packages behind an 8-leg interface (a shuffle, not a carve). The honest future move is BOTH login modes together into an auth package — its own train, reopened on the next auth-surface change. Likewise `passkey_settings` (re-renders the settings partial; view-coupled) and `internal_directory` (machine surface over wiring-owned seams: Sessions, Config.SIP, remoteHostKey) stay in the wiring. Post-carve counts: internal/server 19 non-test files (under the 20-file trigger), hooks +2, api +6. LOC refresh 2026-10-18 (two methods agree): wiring-only 4,056 non-test LOC; whole tree incl. api/hooks 27 files / 5,204 LOC (was 26/4,999 flat on 10-17 — the round-5 trains added ~150 net wiring LOC).
Trigger adjudication (2026-10-06 session): the 09-30 baseline reading
fires on `passkey_api.go` (landed 10-04), but the round-2 plan
(2026-10-05 20:30, §F15 "do NOT start") supersedes — the trigger is
the NEXT file landing in `internal/server` AFTER that plan; stays
gated until then.
**TRIGGER FIRED — adjudicated 2026-10-12 (TODO-execution session):** SEVEN files
landed after the 2026-10-05 20:30 plan: `transcribe.go`+test (10-07 16:32),
`transcripts_api.go`+test (10-07 17:31/43), `internal_directory.go`+test
(10-12 05:28), `passkey_settings.go` (10-12 06:11), `auth_audit.go`+test
(10-12 06:16). The package is now 25 non-test files, ~13k LOC incl. tests.
Dependency map (done, so the executing train skips the research): every API
file hangs methods off the `handlers` hub (pages.go) — session_api alone
calls 5 shared server methods (`sessionIdentityResponse`, `identityFor`,
`verifyDirectoryCredentials`, `mintSession`, `numbersFor`); contacts calls
`notifyContactsChanged`; calls uses `callsIdem` + `recordCallIdem`; webhooks
uses `hooksIdem` + `applyStatusWebhook` + `unread`. Carve design that falls
out: the new packages get their OWN deps structs built at construction from
`handlers` (limiters, idem stores, cookie policy, csrfCfg), with the shared
methods either passed as a narrow interface server implements or moved into
the carving package where they have exactly one consumer. Phase order for
the train: (1) `server/hooks` (webhooks+idempotency, smaller, no session
surface), (2) `server/api` (csrf_api is a zero-dependency leaf — carve it
first inside the package, then calls, contacts, session LAST with an
`auth:` prefix + the TestServer session-behavior suite), (3) newer JSON
files (transcribe, transcripts, passkey_*, internal_directory, auth_audit)
are a SEPARATE adjudication — the row names four endpoints; do not gold-plate
the first carve. Deliberately NOT started inside the 2026-10-12 TODO session:
an auth-surface structural move stacked on a day with two self-inflicted
reds and 13 unpushed commits is the fastest-over-best trap; the map above is
the session's honest contribution.

**Evidence:** arch review 2026-09-30 finding #1 + §05 go-modularize verdict (NO go.mod split — zero Go consumers, Nix vendorHash + release-ritual cost, arch test already enforces the DAG; revisit trigger = a second Go consumer); 2026-10-17 4th series entry re-affirms and pre-arms the dormant split blueprint: `docs/modularization/2026-10-17_PROPOSAL.html`; execution + adjudications recorded above (2026-10-17).

### Mic pre-warm live verification + gates

**Status:** 🟡 `PARTIALLY DONE` (code + gates green; live ritual owner) · **Priority:** Medium · **Effort:** S

`mic.js` pre-warm (speak ASAP after accept) shipped with T07 (2026-10-01, code + island pins) and the T11 perf train added the mic-indicator timing baseline; suite/smoke green. Remaining: ONE live-call ritual on the stack (accept→speak sub-second, indicator at ring, warm release on reject/missed) — owner terminal.

**Evidence:** `docs/status/archived/2026-10-01_04-07_mic-prewarm-accept-latency-train-status.md` §b/§f; T11 timing baseline `scripts/perf-baseline.py`.

### Island-honesty train follow-ups

**Status:** 🟡 `PARTIALLY DONE` (webphone side done; stack lane remains) · **Priority:** Medium · **Effort:** S

Hold machine + `#offline-banner` shipped (`cc98c2e`); T15 landed the pending-copy split, Terminated/watchdog `holdPending` clearing, and the T23 visual pass covers the screenshot leg; vulnix (zero real advisories) + aarch64 exit-green done at the T22 close-out (ELF-byte verify owed at each final-gate cross-build). Remaining: the stack browser E2E the T11–T19 served-markup delta owes (blocked by the stack FreeSWITCH `mod_enum` build break — repair that first), then the runbook obligation closes.

**Evidence:** `docs/status/archived/2026-10-01_05-26_island-honesty-hold-offline-banner-train-status.md` §b/§f; T15/T22/T23 session reports 2026-10-02.

### samber/do composition-root + dashboard train follow-ups

**Status:** 🟡 `PARTIALLY DONE` (webphone side done; two owner calls) · **Priority:** Medium · **Effort:** S

health.css: input + staged rebuild script (`scripts/build-health-css.sh`, locked-nixpkgs + staged build — the artifact sat inside its own @source root) + `checks.health-css` canary landed; the 2026-10-02 rebuild fixed a genuinely stale artifact (zombie classes, missing `--blur-xs`). Family pins landed (T22: `store.thread_flag`, `store.count_archived`, snippets). The local-behind-remote divergence is LIVE daemon behavior (observed twice: named commit rewritten 2026-10-02, same content new hashes — never verify via push logs, only `git ls-remote`). Remaining: stack `/health` exposure policy (remote_ip vs PublicMode vs basic auth) + the v2.9.0 fold decision (default: one release, owner §g2) — both owner.

**Evidence:** `docs/status/archived/2026-10-01_02-12_samber-do-composition-root-health-dashboard-train.md`; T21/T22 session report 2026-10-02 11:43 §a; T18–T23 session report 2026-10-02 11:41 §d.

### Visual harness: shot disposition + optional vision cross-check

**Status:** 🟡 `PARTIALLY DONE` (harness shipped; disposition is owner) · **Priority:** Low · **Effort:** S

`scripts/ui-capture.py` is the persisted harness (T23): config-file boot with the fronted CSRF shape, HTTP seeding (multipart sends incl. an audio attachment, meta-CSRF → `/api/session` → rotated token, transcript seed via `/api/transcripts`), session-cookie INJECTION into chromium (the island's login gates on the SIP WS a bare boot cannot serve), per-surface DOM assertions (marker + optional rendered-copy pin) before every shot, and the ASR-ON 20-shot matrix (10 surfaces × light/dark — M13 extension 2026-10-08, styled 404 added 2026-10-23) into `ui-shots/`. LOCAL-ONLY by budget decision — see the AGENTS command entry. Remaining (owner §g1): eyeball the current matrix; decide per-release vs per-train persistence (default: per release); the optional vision-CLI cross-check needs a provider/key decision.

**Evidence:** T23 in `docs/status/archived/2026-10-02_11-41_t18-t23-train-session7-status.md` §a/§b; run recipe in AGENTS § Commands.

### Cross-repo obligations (stack / pbx-artmann)

**Status:** 🟡 `PLANNED` · **Priority:** Medium · **Effort:** S-M

Owner/stack legs that block or derive from webphone work: repair the stack FreeSWITCH `mod_enum` build → run the stack browser E2E → relock to webphone `cc98c2e` (or newer) + pbx-artmann re-pin; **the cascade-fix train (2026-10-04, island style scoped to `.island` + history filter `wp-mini`) AND the passkey-tail session (same day: csrf.js module, whoamiLine move, enroll error copy) changed served markup/assets — the stack browser E2E re-run above covers both trains' obligation**; stack `services.webphone.paperless` module option + smoke arm; the WebTransport-not-adopted verdict doc; telephony `deploy.md` secret PATH column; ops-runbook demo-call recipe (`originate user/1000 &playback(local_stream://moh)` + the `/var/lib/telephony-secrets/` password path); MOH audibility + `/recordings/` + CDR check; the gateway stack-side bits (`ftypqt`→`video/quicktime` sniff, E2E MMS-outbound, pbx-artmann FEATURES:87 stale text). No assistant ssh — verified handovers only.

**Evidence:** 02:54 / 04:07 / 05:26 reports §f; owner command sheet `docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md`.

### Round-5 plan tail: process folds + row hygiene (2026-10-17 harvest)

**Status:** 🟡 `PARTIALLY DONE — 2026-10-18 closed F75/F76/F79–F84; 2026-10-25 row-hygiene sweep closed F74/F85/F86; remains d2 + markdownlint posture (both owner-gated)` · **Priority:** Low · **Effort:** S — full plan: [docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md](docs/planning/2026-10-17_20-55_SUPERB-round5-pareto-plan.md) (E17–E23, F74–F87)

- CLOSED 2026-10-18: lessons folds (self-verifying scripts + verify-before-describing, appended to the daemon-race entry; IO-storm scoping refinement folded into the local-suite corollary) · module-cache janitor (`scripts/module-cache-janitor.sh`, self-tested, 24 live sweeps) · templ-components tagging note VERIFIED ALREADY-UPSTREAM (#27 + release-checklist + workflow) · branching-flow drafts VERIFIED ALREADY-FILED-COMPLETED (#9, #15/#16) · go-modularize §f 1–40 routing audit (annotated inline; items 30/31 → ROADMAP, 40 → cross-repo row) · carve LOC refresh (wiring 4,056 / tree 27 files 5,204 LOC, two methods agree) · metaengine row CLOSED (G3 via v2.9.0 verified in-tag) · F019 Volume hints landed on all five declarations (`humanScaleVolume`); F026 NOT-DO (no cursor pagination at HEAD to cache — reopens only if a surface adopts cursor paging).
- CLOSED 2026-10-25 (row-hygiene sweep, evidence re-verified at HEAD): F74 (the run's DOM assertions + shot inventory recorded in the 2026-10-18 closing reports; the EYEBALL stays owner §g1) · F85 (snapshot-rot header adopted: both 10-18 reports carry it, AGENTS Conventions line landed) · F86 (verified at source: the retry-or-kill wrapper ask is SUBSUMED by BuildFlow's 10-13 retry-economics train — starvation-floor retry + the no-output watchdog cover the 4.5h network-wedge class; the fleet GOPROXY fallback chain stays the owner's §g-3 network-posture decision, crush-config side). REMAINS: d2 cheat-sheet upstream fix (`style.stroke-dash` on v0.9.0; blocked on the sitting's Q3) + human eyeball of the 2026-10-17 D2 pair (image-blind assistant ceiling) · markdownlint posture execution (post-sitting row-18 verdict).

**Evidence:** status reports `docs/status/2026-10-17_20-16_*` + `2026-10-17_20-34_*`; round-5 plan Errata §1–§3; 2026-10-18 closing report.

### Native-app train: ring-while-closed legs (2026-10-23 ship, webphone side closed)

**Status:** 🟡 `P2 SHIPPED server-side; P1 shell scaffolded (webphone-android v0.1, BUILD GREEN + 16KB-aligned); the two remaining legs live OUTSIDE this repo` · **Priority:** High · **Effort:** M (each leg)

- **Stack leg (nix-international-telephony):** the FS dialplan must POST `webphone:/hooks/ring` (Bearer webhook secret; `{owner, from, call_id, state: ring|end}`) on inbound-call ring + ring-end — the same bridge pattern as the Telnyx webhooks. Until this lands, the seam is inert (fire-and-forget 202, no devices rung). Config to expose: `services.telephony.webphone.push.*` (freeform `settings` already carries it; document in the stack's deploy docs).
- **Shell leg (webphone-android):** first device train — exercise the SIP UA against the live WSS gateway (REGISTER/INVITE/BYE/DTMF/hold + incoming INVITE + REFER/NOTIFY are implemented but UNTESTED against FreeSWITCH), compose the real CallStyle notification (v0.1 posts the FSI heads-up), the origin-picker settings UI (App.origin is a constant today), and the `/api/session` field-name check against the served sessionIdentity shape.
- **Deferred row SHIPPED 2026-10-25 (this repo):** the Settings tab now carries the push seam state row + the extension's registered devices (read-only; the wake address never renders) — the .templ/i18n surface landed once the chat-networks train went quiet. The served-markup delta rides the standing stack browser E2E obligation (next relock).

**Evidence:** decision doc `docs/planning/2026-10-23_16-13_native-app-platform-decision.md` (status: DECIDED + SHIPPING); `~/projects/webphone-android` README honesty markers; `internal/push` + `internal/server/hooks/ring_test.go` + `internal/server/push_api_test.go` for the shipped server contract.

- [x] **Hub cross-link: webphone's wire contracts from its inventory page**
      (eventcatalog-hub) — SHIPPED 2026-10-28 via webphone's OWN exported page:
      the catalog declares EventCatalog `specifications` (asyncapi + openapi)
      and `webphone catalog` renders both spec files into the export tree from
      the pure generators behind `webphone contracts` and GET /openapi.json
      (`server.RenderOpenAPIContract`, newly exported), pinned by a cmd-level
      export test. The SystemNix inventory-module option is unnecessary now.
      (Was: DISCOVERED 2026-10-24: the hub's `services/` AND `domains/` are
      GENERATED (gitignored; merge.py unions source trees, bodies preserved
      first-source-wins). A direct `services/webphone/index.mdx` edit is
      drift (tested: appended content sits untracked until the next build
      clobbers it — restored byte-exact 2026-10-24).)
