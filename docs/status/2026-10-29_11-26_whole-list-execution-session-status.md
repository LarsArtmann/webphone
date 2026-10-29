# Status — whole-list execution session (2026-10-28 ~10:45 → 11:30 CET, report written 2026-10-29 11:26)

Claims verified as of **2026-10-29 11:26 CET**: my last narrative commit `6eb6522`;
the docs sweep landed via daemon commit `3f8b1ca` (2026-10-29 01:40) and was then
BUILT UPON by a later session's 2026-10-29 sweep (docpin tests — not mine, untouched);
9 overnight commits sat unpushed past the 60-min bar and were pushed manually at
report time (`dc5a4a0..0d2fb36`, anchor 2026-10-29 01:40, ~9.7 h) — **the CI run
for that push is IN FLIGHT and unverified**. *(Annotated 2026-10-29 12:20 by the
follow-up session: that run, 40857662429, was RED — the daemon had committed
`layout.templ` ahead of its i18n key (healed by the next commit) plus the first
two clone groups; both classes were fixed and CI went green on 40862216157.
The missing final-gate evidence also closed: full buildflow re-run with
`--failed-only` over the fixed tree = exit 0, log `/tmp/buildflow-retry-2026-10-29.log`.)*
The session ran concurrent with the
dialer train (#7/#8) in `internal/server`, `views/`, `readmodels/` — dodged by file
selection throughout.

Scope: execute the TODO_LIST. A live concurrent train owned the dialer rows (#7/#8),
so those were left to it (attribution rule); everything else executable by an agent
in this repo was executed, adjudicated, or closed-as-already-done with source evidence.

## a) FULLY DONE (this session, all verified green at their commits)

1. **Whitespace-drift false-positive fix** (`cd0d0c4`): the pre-commit gate now
   materializes the diverging staged blobs into a temp tree and asks the flake's own
   treefmt (config parsed out of the `.#formatter` wrapper) whether the bytes are
   canonical — formatter reflow output ACQUITS instead of forcing `--no-verify`.
   The daemon-audit `--ref` mode adjudicates landed commits identically (verified:
   canonical reflow → exit 0, hand drift → exit 1). Selftest grew 6→8 cases, all
   green; AGENTS Commands line + hook header updated.
2. **Clock seam** (`8e240bb` via daemon): `window.__wpClock` deleted; all shell time
   reads (error throttle + relative-time tick) go through one exported
   `clock` holder; the three specs that monkey-patched `Date.now` drive the holder.
   tsc strict green, island suite 261/261.
3. **tw-freshness advisory CI lane + drifted artifacts rebuilt** (`0dd0a86`): the
   committed tw.css/login.css were STALE behind the agent-UX + review view trains
   (found by running the rebuild — +15/+81 lines). Lane is advisory like
   reproduce-first (daemon pushes in-flight views before a train's tail rebuilds);
   artifacts recommitted byte-fresh. The lane proved itself within the hour: it
   red-annotated an intermediate commit where new view changes had landed post-rebuild.
4. **Catalog cross-link** (`ba9613e`): the exported eventcatalog service page carries
   EventCatalog `specifications` (asyncapi + openapi); `webphone catalog` renders both
   spec files into the export tree from the pure generators (`server.RenderOpenAPIContract`
   newly exported). Pinned by a cmd-level export test + a domain spec pin; live export
   verified into a temp dir.
5. **Chat lane v2 — compose by address shape** (`6eb6522`): `domain.ParseComposedAddress`
   routes matrix `@local:server` + session `05…` through the sender registry; sms
   dialable otherwise; Signal stays compose-born-inbound (phone-shaped = ambiguous).
   Form copy/i18n both languages, `inputmode="tel"` dropped, no pin sites affected.
6. **Chat lane v2 — matrix read receipts** (`6eb6522`): `m.receipt` ephemera → a peer's
   `m.read` on the homeserver event id (the sender's provider ref) flips the outbound
   bubble to delivered via the SAME `DeliveryReceipt` path webhooks take;
   `chatloop.Sink` grew `RemoteRead` (compile-time complete; both fakes updated);
   own echoes skipped; uncorrelated refs benign (`store.ErrNotFound`). Receiver test
   pins both directions.
7. **Branching-flow triage — 115 findings → 0 ungoverned**: 36 reasoned
   statement/signature nolints across 24 files (invariant-by-construction index math,
   never-closed buffered channel, composition-root panic/MustInvoke doctrine,
   flag-pointer derefs, the phantom-string classes the strong-id/BD003 + engine-layer
   + open-roadmap rulings already decided), then a fresh 78-entry baseline
   (TypeDecl decided architecture: branded-id generics, typed error products, catalog
   mirrors, the six single-implementer Deps seams + 2 nolint-immune FLAG_PARAM sites)
   wired via the `baseline:` key — suppression verified ACTIVE.
8. **Error-code registry heal** (`63e9f25`): the dial train's daemon-split commits
   left 7 `dial_*` codes unregistered (the suite's only red); regenerated, arch green,
   full `go test -count=1 ./...` green (31/31).
9. **Fleet ask FILED**: projects-management-automation **#348** (SIGTERM flush +
   startup unpushed sweep + shutdown line) — verified at source first (Stop() drops
   batches with no drain; Start() has no sweep; per-push logging already exists, so
   the ask was corrected from the TODO's "log the push leg" premise). Draft kept in
   docs/drafts/.
10. **cqrs-htmx M12/M13/M14 closed as ALREADY-SHIPPED upstream** (no duplicates
    filed — verify-before-filing caught all three stale premises): M12 = issue #32
    exists AND `e7b62cd9` implements the ring buffer (verified not in pinned
    v4.16.0; close-out comment posted); M13 = usermgmt README § "Projection
    readiness" already on master; M14 = crush-config `a9aeb2a` already carries both
    lesson rules.
11. **Two manual-push bar invocations, anchors stated**: (1) red main + verified fix
    in hand (09:52 go-tests red = the registry drift; my heal) → pushed, CI 40671674989
    **success**; (2) report-time push of the 9 overnight commits (anchor 01:40,
    ~9.7 h past the bar) → **CI in flight, unverified**.

## b) PARTIALLY DONE

- **Chat lane v2**: 2 of ~5 fences shipped (compose-DM, read receipts). Remaining
  four adjudicated-open with reasons in the TODO row: matrix group rooms (thread
  model keys one peer — a model change, not a fence lift), matrix E2EE (olm/megolm),
  signal inbound attachments (signal-cli side), signal groups.
- **TODO_LIST harvest**: my row closures + sweep line landed (`3f8b1ca`) but the
  header was superseded ~14 h later by another session's newer sweep — expected
  docs-health churn, listed here for the audit trail.
- **Final full buildflow gate**: launched in the background at train end and NEVER
  VERIFIED (see d); partial evidence stands (full go suite green, targeted
  branching-flow steps green, CI green on the mid-train push).

## c) NOT STARTED (deliberately)

- **KMP spike (T26/T27)** — DEFERRED with stated reason: webphone-android carried a
  LIVE concurrent call-stack train (daemon commits minutes old, Kotlin files
  mid-edit); Gradle surgery on settings.gradle.kts + a new shared/ module would
  collide head-on. Tier-0 by its own plan ("may stay unstarted indefinitely").
- **Dialer #7/#8 rows** — the concurrent session's active train (T11–T15 of its plan
  shipped while I watched); left untouched, attribution rule.
- **Owner-gated rows** (untouched, unchanged): host-leg ssh verdicts, post-deploy
  deep-verify, htmx SSE-ext upstream filing, release announcements, stack browser
  E2E re-run, reproduce-first + gated-GO flips, M18/M19 triggers, owner-calls batch
  sitting, mic-pre-warm ritual, visual-harness eyeball, d2 cheat-sheet fix,
  markdownlint posture.

## d) TOTALLY FUCKED UP (honest near-misses — nothing destructive landed)

1. **The final background buildflow was never checked.** I backgrounded the full
   gate, pivoted to docs, and the session boundary ate the job + `/tmp` log. Worst
   actual consequence: missing final-gate EVIDENCE, not a red (go suite + CI green
   in between) — but "started a gate and never read its verdict" is exactly the
   class the repo's own lessons file condemns.
2. **A botched edit deleted 5 lines of `matrix_test.go`** (replaced a block instead
   of inserting); caught and repaired one tool-call later, verified by re-view + vet
   + suite. Process slip, zero landed damage.
3. **Three nolints initially suppressed nothing** (FLAG_PARAM ×2 ignored placement,
   one nil-deref finding pointed at a continuation line): found only by diffing the
   saved baseline against expectations and re-placing. Cost: one extra scan round.
4. **Narrative-commit racing**: 5 of my 8 intended narrative commits were swept by
   the daemon as `chore:` before my `git commit` ran (content always landed; the
   message did not). The known add+commit-in-ONE-call rule was followed and the
   daemon still won twice — the window is the tool-call boundary itself.
5. **The TODO header I wrote lived ~14 h** before a later session's sweep replaced
   it (normal churn, but I verified closure content survived only by grepping at
   report time).

## e) WHAT WE SHOULD IMPROVE

- **Never background a gate without a read-back step in the same phase** — add to
  crush-config lessons (candidate: "a backgrounded verification is an unverified
  verification until its output is read; read it before pivoting").
- **View-touching trains need a `scripts/build-tw-css.sh` habit line** — the
  artifacts drifted TWICE in one day (agent-UX/review trains, then again mid-push).
  The advisory lane now catches it, but the fix (rebuild) should be a named step in
  the train checklist (AGENTS Commands already documents the script; the HABIT for
  view edits is the gap).
- **FLAG_PARAM cannot be nolint'd** (v0.6.4) — the two sites ride the baseline;
  consider `rule_accepts` for FLAG_PARAM only if more appear, else they'll re-noise
  every scan from moved checkouts (upstream #13's host-absolute ids).
- **The daemon push stall recurred** (~9.7 h overnight) exactly as pma#348 predicts;
  until it lands, `push-stall-watch.sh` after every switch-to-configuration is the
  only guard — consider wiring it into watch.yml as a */15 observer (a new small task).
- **Baseline portability**: host-absolute ids bind this checkout; a CI-side
  re-baseline is impossible (sandbox). Upstream #13 remains the durable fix.

## f) NEXT (up to 50, priority order; ✋ = owner)

1. Verify the in-flight CI run for `dc5a4a0..0d2fb36` (pushed at report time).
2. usermgmt bump train once a tag carries `e7b62cd9` → then WIRE the bounded
   AuditLog read projection (the read-side audit surface unblocks; M18-adjacent).
3. Stack browser E2E re-run — the owed set now includes the compose-form markup
   (mine) + review/dial-cooldown surfaces (dialer train). ✋ owner-runnable
4. Blocking-flip decision for `tw-freshness` after N clean advisory trains (N=?). ✋
5. Blocking-flip decision for `reproduce-first` (same pattern). ✋
6. Implement pma#348 (SIGTERM flush + startup sweep) in projects-management-automation.
7. Add `push-stall-watch.sh` as a watch.yml */15 observer (local timer, T21 verdict).
8. htmx SSE-ext `console.error("[object Event]")` upstream filing (draft ready). ✋
9. Host-leg ssh verdicts (NRestarts + panic grep + Caddy /events + ring hook probe). ✋
10. Post-deploy deep-verify of the error-surface train surfaces on prod. ✋
11. Chat: matrix group-rooms data-model spike (thread key is one-peer — needs a
    model decision before any fence lift).
12. Chat: matrix E2EE research (olm/megolm Go story).
13. Chat: signal inbound attachments (signal-cli side plumbing, stack-adjacent).
14. Chat: signal groups (same model issue as matrix groups).
15. Typeahead suggestions for chat-network peers (thread-derived source; today only
    phone numbers are suggested). Product call. ✋
16. KMP spike T26/T27 once webphone-android is quiet (deferred this session).
17. M18: durable agent-action read model (trigger: operator need). ✋
18. M19: attended transfer once the stack grows an FS Replaces lab. ✋
19. Release announcements posting (drafts ready, channel + posture). ✋
20. Owner-calls batch sitting (38+ rows in the briefing doc). ✋
21. Mic pre-warm live-call ritual. ✋
22. Visual-harness matrix eyeball + persistence policy. ✋
23. `Rule accepts` decision for FLAG_PARAM if new sites appear.
24. Re-baseline branching-flow from any moved checkout (upstream #13 follow-up).
    *(Annotated 2026-10-29 12:30: measured 164 live findings vs the governed 78 —
    a re-record would bless ~86 ungoverned ones from the dialer/docpin/review
    trains; the refresh is BLOCKED on the 2026-11-05 triage cadence, not just on
    a checkout move. Items 25/26 shipped same day: crush-config 4468af5 + the
    AGENTS Formatting bullet.)*
25. crush-config lessons candidates from this session: backgrounded-gate read-back;
    daemon-race at the tool-call boundary (add+commit in one call still loses).
26. Add the `nix fmt`/`build-tw-css.sh` habit line to the view-edit train checklist
    (AGENTS or release-runbook).
27. Consider a compose-form island spec (server-side tested; the form markup is
    greppable) — small.
28. erraudit tier-2 monthly re-measure — due 2026-11-05.
29. Sessions credentials-at-rest periodic re-review — due 2026-12-30.
30. Quarterly standing watches — due 2026-12-20 (sip.js, templ-components, oxlint
    globals, E2E budget).
31. Quarterly modularize re-trigger re-review — due ~2027-01.
32. Dialer #7/#8 row harvest once the concurrent train closes (their session's).
33. go-cqrs-lite #71/#72 watch (TypedReader truncation + keep-newest-N).
34. loginpage next-tag passive watch (bump when tagged).
35. Stack chat-networks deployment (signal-cli unit + matrix homeserver). ✋ stack repo
36. `/health` dashboard deploy verify on the host. ✋
37. d2 cheat-sheet upstream fix (blocked on owner Q3). ✋
38. Markdownlint posture execution (owner row-18 verdict). ✋
39. Watch pma#348 + cqrs-htmx #32 to closure.
40. If more daemon stalls: cite pma#348 in the ledger rows.

## g) QUESTIONS (cannot be answered from the repo)

1. **tw-freshness blocking flip**: the artifact-drift class fired TWICE within one
   day; does that justify flipping the lane to blocking NOW, or hold the
   ratified "after N clean advisory trains" bar (and what N)?
2. **pma#348 ownership**: may a future session implement the SIGTERM flush +
   startup push sweep in projects-management-automation directly, or is the daemon
   owner-only territory?
3. **Chat compose scope**: should the compose typeahead suggest chat-network peers
   (sourced from existing threads/CRM), or is address-shape compose enough for now?

— Session record: 8 narrative commits (`cd0d0c4`…`6eb6522`) + 2 verified upstream
filings/comments (pma#348, cqrs-htmx #32) + 1 manual push beyond the bar mid-train
(CI green) + 1 at report time (CI in flight). No reverts, no lost work.
