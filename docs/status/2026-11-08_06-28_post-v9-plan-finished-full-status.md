# Status — the post-v9 execution plan FINISHED: P2 remainder + all nine P3 + the bug the review found (2026-11-08 06:28)

**Snapshot-rot header: every claim below was verified as of 2026-11-08
06:28 CET.** This session continued from the 2026-11-07 20:18 halfway
report after the owner's "GET SHIT DONE — THE WHOLE TODO LIST" order,
and later answered the "is the plan superb? if not make it perfect"
challenge by auditing my own deliverables before delivery. Everything
is pushed; the final commit's CI run `42450705344` was `in_progress`
at the timestamp above — the two PRIOR runs (`42440660594` for
`48cda2c8`, `42440158730`/`42439390608` for `20832456`) are
**completed success, all seven jobs**, and the in-flight run carries
only docs deltas on top of a green tree. Full-suite local legs: `go
test -count=1 ./...` rc=0, docs battery 5/5 PASS, index check exit 0,
erraudit 0 violations.

## a) FULLY DONE

| #   | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A1  | **P1-5b closed — buildflow's findings gate green and attributable.** The halfway RED attributed to three first-party error findings, all fixed: `cue-vet` (ci.yml lacked the fleet `permissions:` block → `contents: read` declared with the policy comment), `erraudit` (`errors.As` → `errors.AsType` at the gitignore canary, the gateway tests' established shape), `go-humanize` (deleted `formatThousands`, `humanize.Comma` at the call sites, the map-case test now pins the LIBRARY against the same expectations — a format-regression canary for the views). Gate green on the full rerun.                                                                                                                                                                                                                                                                                                   |
| A2  | **Paperless token parity: a real bug found by the "make it perfect" review and fixed end to end.** `paperless.token_file` did not exist while its CRM twin did — an operator following the documented CRM pattern got the key silently dropped by koanf and a misleading "token is set without paperless.url" boot failure. Fixed: struct field, resolve-before-validate in `Load`, `errPaperlessTokenBothSources` rejection, `TestLoadResolvesPaperlessTokenFile` (four arms), README docpin row, two error-registry rows (`-update`), CHANGELOG line. The docpin gate caught the undocumented key the moment the field existed.                                                                                                                                                                                                                                                                       |
| A3  | **P2-9 branching-flow governance VERIFIED (and a footgun documented).** Live scan 96 → 95 (the transport fix below removed one); 95/95 findings resolve against the 157-entry baseline — 63 already matched, 33 new sites from the landed trains (agent audit, dial gate, CRM context, live context, chat loop) added per-finding via `baseline add`. **Footgun on the record:** `--save-baseline` writes the UNACCEPTED remainder, not a snapshot — against a green gate it EMPTIED the file (0 findings); recovery from history `ed9ccb63` (the daemon had swept the emptied file, so `git restore` gave back the empty version first). One real fix from the scan: the CRM transport's unchecked `http.DefaultTransport` type assertion now guards with a documented-unreachable fallback.                                                                                                           |
| A4  | **P2-6 quiet-host benchmark verdict delivered** (commit `e9c8836b`): load 0.75/2.06/9.72, spreads 1.4–8.1 % (vs ~2× on 11-06), every shape at or below its 2026-09-24 median (1×1 33.2 vs 35.0 ns; 100×10 104.3 vs 105.0 ns). The 2026-11-06 "measurement attempted, verdict carried" entry is ANSWERED. No regression from either bump train.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| A5  | **P3-1 owner sitting pack, SECOND EDITION** (`docs/planning/2026-11-07_21-30_owner-sitting-pack.md`): 58 rulings (A product/UX ×7, B privacy/data ×8, C infra ×11, D tooling ×18, E cross-repo ×14) + a 38-row F bulk table folding the ENTIRE 2026-09-22 briefing; each ruling carries a recommendation, the other branch's cost, and rows tagged [ratify] where code already shipped. Mechanically consistency-checked: 58 headings = 58 template refs, 38 F-rows, zero broken links. TODO's OWNER-calls row names it THE ONE answer surface; the packet and briefing remain as evidence appendices. The pack's B4 changed character mid-session: the paperless gap it described shipped as a fix, so the row records the standing rule (every bearer-token seam lands with its `_file` twin) instead of asking the question.                                                                         |
| A6  | **P3-6 standing-watch calendar** (`docs/watches/2026-11-08_standing-watch-calendar.md`): one dated table, every row with cadence, next due, named trigger, evidence home, on-fire action. The lapsed erraudit monthly check was RE-MEASURED same-session (0 violations, `nolint-audit` 68/68 needed, 0 stale) instead of silently moving the date — next due 2026-12-08; quarterly buildflow 12-20, credentials 12-30, modularize ~2027-01, dedup 2027-01-11. The Fleet-libs reopen-trigger registry is referenced, never duplicated. TODO's Standing-watches row now points at the calendar.                                                                                                                                                                                                                                                                                                           |
| A7  | **P3-3 detection column** (`docs/error-contract.md` § "Detection — how anyone KNOWS it happened"): twelve quiet classes answered with the FIRST place to look (CRM down → `/healthz` strict leg; archiver → journal; push transports; receiver loops; dial-gate store; blob readiness; hooks fail-closed; ASR; flight-recorder off; passkey dismissal; transcription skips), plus the ruled absence (latency classes have no detection because nothing measures them — the design note owns it). Two §f rows struck with verified evidence (readmodels Contacts move: `store/contacts.go` gone, `app.go:428` wires `engine.Contacts()`; empty-vs-filtered audit = the fleet-libs train's per-site rulings) — audit-all dropped 74 → 72, exactly AT the pin: no raise bought.                                                                                                                            |
| A8  | **P3-4 latency design note** (`docs/planning/2026-11-08_latency-measure-first-design.md`): dial-press → first-audio and ASR round-trip with t0/t1 at user-perceivable boundaries, four named legs each (island-overhead / signaling / media-attach / audio-gap; capture-flush / POST / provider / render), reporting home decided (island beacons → server histograms, opt-in flag default OFF = byte-identical), ~160 LOC cost estimate, no code per the L3 ruling.                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| A9  | **P3-5 four decision briefs** (`docs/planning/2026-11-08_decision-briefs.md`): paperless seam shape (Settings row recommended; per-fax markers' correlation machinery costed), F1 ink flip + F19 signed-out door (both already SHIPPED 11-07 per the review's recommendation — the briefs record reversal costs), attended transfer (M19 NO-GO stands; the "warn" option is a permanent lie mid-call; GO-day trigger = the stack lab scenario). Linked from the TODO Paperless row.                                                                                                                                                                                                                                                                                                                                                                                                                     |
| A10 | **P3-7 cross-repo handover pack** (`docs/planning/2026-11-08_cross-repo-handover-pack.md`): four legs as verified runbooks — stack `crm.caFile` pass-through (webphone module side SHIPPED + pinned; the STACK option tree has no caFile — read-verified), `/hooks/ring` dialplan leg (zero references in the stack today — read-verified), `mod_enum` CI-ceiling repair (`applications.enum` at freeswitch-modules.nix:29), the #9 owner sequence. Sequenced by blast radius per effort; zero invented paths. Linked from the TODO Cross-repo row.                                                                                                                                                                                                                                                                                                                                                     |
| A11 | **P3-8 native-app preflight** (`docs/planning/2026-11-08_native-app-preflight.md`): the android repo is QUIET (clean tree, last commit 11-05, 2 commits/7d) — HALF the KMP gate met, spike STAYS deferred; the first-device-train checklist (7 legs: SIP UA vs live WSS, incoming INVITE/REFER, CallStyle, origin picker, sessionIdentity field check, the stack ring hook, push docs). Read-only proven: `git status` re-checked identical after. Linked from the TODO Native-app row.                                                                                                                                                                                                                                                                                                                                                                                                                 |
| A12 | **P3-9 parked-decisions audit** (`docs/planning/2026-11-08_parked-decisions-audit.md`): all four parked items (scheduled sends NO-GO, session messenger, ssetest `WithLastEventID` ride, #8 review dimensions) verified carrying live triggers and single homes; NO-GO premises re-verified at HEAD. One wording risk found (ratification noted in two homes) with the pack named the answer surface. Linked from the TODO Scheduled-sends row.                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| A13 | **P3-2 filings: one draft shipped voice-clean, one filing verified to a NO.** docs-gates four-defect draft (`docs/drafts/2026-11-08_docs-gates-four-sweep-defects.md`) — DF verdict flips on corpus growth, standalone verbs ignoring `.docs-gates.json`, escaped-pipe cell splitting, unvalidated verdict cells; one reproduction each, first-hand from this session; checker `0 FAIL 0 WARN` after it caught two real defects (banner regex, missing attribution footer) and I caught the nested-fence truncation in my own extractor. The htmx `[object Event]` upstream filing is a **verified NO** (`docs/drafts/2026-11-08_htmx-object-event-verification-no.md`): the served ext is the cqrs-htmx FORK, the current fork has no console logging, upstream `htmx-extensions` sse.js has none either — premise dissolved at source. The two branching-flow drafts from 2026-10-11 already existed. |
| A14 | **The cadence drain the final report triggered.** My completion report made 66 unarchived reports (threshold 65); the oldest gate-clean live report (`2026-11-04_18-31_fleet-libs-audit-status.md`) went through the FULL ritual — `check-rows --strict` + `bare-rows` green BEFORE the `git mv` (two other gate-clean candidates identified first), banner, two citations repointed, INDEX regenerated. Also fixed the one annotate finding in a concurrent session's new report (a word glued to a closing backtick — the mechanical-unblock class, per the sitting pack's own D4 recommendation). Battery 5/5 PASS at the end state.                                                                                                                                                                                                                                                                 |
| A15 | **CI verification by explicit conclusion reads, not watch exits**: run `42440660594` (all seven jobs success: fast flake checks incl. docs battery, `go test -race`, nix build, live-boot smoke, three advisories) — the first green day after the findings-gate fixes; the concurrent session's `gh run watch --exit-status` defect (§c1 of its report) was honored by never trusting the watch exit alone.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |

## b) PARTIALLY DONE

1. **Filing the docs-gates issue** — draft committed and voice-clean; the
   POST is owner-gated (unsolicited banner unchecked). Fold the
   `--save-baseline` remainder-write footgun into the draft as a fifth
   defect before posting (discovered after the draft was written).
2. **The final CI run** (`42450705344`) was in_progress at report time; the
   tree it carries differs from the verified-green `48cda2c8` only by docs
   and the two daemon sweeps, but the conclusion is not yet quoted — verify
   on your next look.
3. **AGENTS.md updates for the session's lessons** — the memory-maintenance
   rule says update at discovery, and three lessons (the `--save-baseline`
   remainder semantics, `ContactContext` returning a value type making
   `(nil, true)` impossible, the annotool archive-ritual recipe already
   being documented) are only partially landed (the annotool recipe IS in
   AGENTS; the other two are in commit messages and the completion report,
   not yet in AGENTS bullets).

## c) NOT STARTED

1. **The owner sitting itself** — the pack is ready; 58 rulings + F bulk
   await answers (that is the point of the deliverable, not a gap).
2. **The `annotate_max` 205 drain** — the corpus sits at 206/205-scale
   territory every time any session grows a doc; it needs a deliberate
   ANNOTATE session (the concurrent session's own §c3 says the same).
3. **Nothing else from the plan** — all 27 tasks (P0-1 … P3-9) are executed
   or verified to a ruling; the halfway report's remaining items (P2-6,
   P2-9, P3-1…P3-9) are all closed this session.

## d) TOTALLY FUCKED UP

1. **`buildflow --save-baseline` emptied the 124-entry baseline** and my
   first recovery (`git restore`) gave back the EMPTY committed version
   because the daemon had swept the emptied file between save and restore —
   the real recovery needed history archaeology (`git show ed9ccb63:…`).
   Two compounding mistakes: trusting a blanket re-save without reading its
   semantics first, and restoring from HEAD instead of checking whether
   HEAD still held the good version.
2. **The B4 token-parity flip-flop**: correct claim → wrong "correction"
   (misattributed a grep hit to the Paperless struct when it was the CRM
   struct) → wrote resolution code against a nonexistent field → the
   COMPILE failure proved the original claim right. Three reasoning steps
   spent to arrive back at the first conclusion; the grep without struct
   context was the root error.
3. **My own draft extractor truncated the filing body at the first inner
   code fence** — the exact nested-fence failure the github-voice skill
   documents (`sed` stops at the first fence; use split/rsplit) — and I hit
   it anyway with a regex, producing a false "missing banner" FAIL that
   cost one debug cycle before the second extraction with `rindex` fixed
   it.
4. **The sitting pack's first edition shipped with a placeholder E-row,
   ten missing ROADMAP rulings, and an answer template lagging the IDs** —
   caught by my own audit before the owner read it, but the audit should
   have been part of writing, not a second pass after the owner's challenge.
5. **One `edit` tool refusal chain on TODO_LIST.md** (shared file, modified
   by the concurrent session between my read and edit) — the guard worked,
   but I initially applied the edit against a stale read; re-read-then-edit
   cost a round trip that reading the section FIRST would have saved.

## e) WHAT WE SHOULD IMPROVE

1. **Run the mechanical checker AS PART OF writing**, not after: the pack
   consistency check (IDs ↔ headings, links) and the voice checker both
   caught real defects on their first run — every deliverable with an
   existing checker gets checked before it is declared done.
2. **Read struct fields with their struct in context** (`go doc`, or read
   the enclosing type block): a bare grep across config.go cannot tell
   which struct owns a hit — the flip-flop class, now twice-caught.
3. **Understand a destructive tool's semantics before the first real
   invocation** (`--save-baseline`): one docs grep would have revealed
   "writes the unaccepted remainder" and saved the empty-file incident
   entirely.
4. **Prefer honest strikes over pin moves when the corpus grows**: today's
   74→72 came from striking two genuinely-landed rows; the ratchet stayed
   honest AND the pin held. Make that the default reflex on every
   DF-brittleness trip.
5. **For shared files under a concurrent session, re-read the section in
   the SAME tool call as the edit** — the daemon and the sibling session
   both race TODO_LIST.md; two refusals this session were both stale-read
   artifacts.
6. **Archive-drain candidates should be PRE-IDENTIFIED gate-clean** —
   scanning for check-rows/bare-rows-clean live reports BEFORE the cadence
   gate trips turns the drain from a strike job into a `git mv`.

## f) NEXT (up to 50, ranked by leverage; §f of the completion report holds the top 5 — this is the fuller board)

1. Answer the sitting pack (58 rulings + F bulk) — unlocks M05, M07–M11,
   M13, the pma#348 PR, the coverage floor, AGENTS compaction, and ~38
   briefing rows.
2. Fold the `--save-baseline` footgun into the docs-gates draft, then post
   it (owner review → `gh issue create --body-file`, never `-`).
3. Verify run `42450705344` concluded success (docs-only delta on a green
   tree; conclusion read, not the watch).
4. The `annotate_max` drain: a deliberate ANNOTATE session over the
   archived corpus's 206 findings (the v9 recipe, ~15 min/file).
5. The archive-residue live side: the next ten oldest live reports through
   the proven v10 recipe (per-cell strikes with transcribed evidence).
6. Paperless Settings row (brief 1, ~20 LOC + i18n + pin) once the sitting
   ratifies the shape.
7. The latency beacon train (~160 LOC per the design note) once the opt-in
   flag posture is blessed.
8. AGENTS.md: add the `--save-baseline` remainder semantics + the
   `ContactContext` value-type contract to the standing bullets.
9. Missed-badge clearing on the island panel opening (pack A4b, ~10 min)
   once A4 is answered.
10. Search `?q=` deep-linking (pack A5, one `pushState`) once A5 is
    answered.
11. Stack leg: `crm.caFile` pass-through (handover pack Leg 1) — small,
    unblocks private-CA CRM deployments.
12. Stack leg: `/hooks/ring` dialplan POST + VM-test scenario (Leg 2) —
    unlocks ring-while-closed E2E.
13. Stack leg: `mod_enum` repair-or-exclusion (Leg 3) — unblocks every
    relock.
14. pma#348: authorize + land the push-leg drain/sweep fix (pack C1, ~100
    min).
15. The upstream go-idempotency ask (the concurrent session's §b1 —
    TimingScale 5 is a band-aid; the verify-before-filing + github-voice
    ritual is owed).
16. erraudit `tree` same-name-dedupe upstream filing (pack D17, owner's own
    tool, near-zero fix cost).
17. `file-size-check` configurable-cap upstream ask (pack D10) — makes the
    46-entry baseline derivable.
18. The flock gate wrapper (pack D18, ~10 lines) — closes the
    gate-collision class fleet-wide.
19. M05 config strictness (pack B3 GO) — the file-strict/env-warn flip +
    deployment `settings` grep.
20. M07–M09 group rooms (pack B1 GO, 3 × 100 min) after the sitting.
21. M10–M11 E2EE decrypt-only (pack B2 GO) + the backup-set extension.
22. The D3 agent-liveness train (design exists: `2026-10-29_16-17` doc;
    owner-GO + security review owed).
23. The `/health` unsafe-eval → nonce-lane retirement (TODO_LIST Datastar
    residue row).
24. Batch identity `by-phone?numbers=` on the CRM producer (TODO §CRM tail
    2) when a list exceeds ~20 numbers per render.
25. The provider-failure rule statement in `docs/error-contract.md` (pack
    A3 — the typed machine is already right; state it).
26. The "Add to CRM" link for unknown callers (TODO §CRM tail 1).
27. M19 attended transfer: schedule the stack FS-lab scenario (the GO-day
    trigger) as a stack-repo task, not a webphone one.
28. Retire the `/hooks` 400-body-text + `msg/`→`message/` idem-rename rows
    wholesale once F5/F6 answer "all yes".
29. `settingsRow` retire-or-build (F2 — FOURTH surfacing; the sitting
    closes it forever).
30. The suppression-bucket doc in dedup-registry (F3) + the quarterly
    suppressed-set audit confirmation (F4).
31. The devShell tool additions ratification (F8) — silences 4 buildflow
    warnings, one-line nix diff.
32. lychee archived-dir excludes (F9) — three `exclude_path` lines in
    lychee.toml.
33. The auth-package carve re-trigger check (TODO §carve — dormant; run
    the detection commands once this quarter).
34. `webphone-backup` + `webphone-backup-drill` as ONE `nix flake check`
    gate arm (TODO §CRM tail 4 — the KVM-gated pair verified only
    individually).
35. Native-app: merge/rebase the four open Dependabot PRs in
    webphone-android (housekeeping, zero webphone risk).
36. Native-app: the `/api/session` field-name assert on the Kotlin side
    (preflight checklist item 5).
37. The voicemail/contacts/history reload-honesty island specs — verify the
    M20 copy pins still hold after this week's shell edits (specs exist;
    re-run and re-assert).
38. Keep `docs-gates` `--check` honest about skips: the table-damage SKIP
    in CI prints visibly — consider a CI-side annotate.go copy to un-SKIP
    it (the wrapper documents the fan-out file as the blocker).
39. The stale `pma` unit `core.hooksPath` override (pack D6) — if the
    sitting authorizes, the fix is fleet-wide and small.
40. README reflow audit: dprint re-aligns config-table columns; confirm the
    `paperless.token_file` row's padding survived the daemon's sweeps
    byte-canonically (cosmetic, formatter-owned).
41. Benchowance follow-up: re-run `BenchmarkHubFanOut` after the NEXT
    cqrs-htmx/go-sse bump (the baseline file's standing trigger — no action
    until then).
42. The `wt`/badge pack A4a REJECT-semantics note in AGENTS (one sentence)
    once A4 answers.
43. Cross-check the stack runbook's error-contract copy against the new
    Detection section (the sync rule; the runbook's ladder is a different
    surface, but the section pointer is owed).
44. QMD `get`/`multi_get` re-test after the next Crush upgrade (standing
    watch; upgrade-triggered).
45. Session-credentials at-rest review: pre-schedule the 2026-12-30 ritual
    owner-side (the calendar carries it; a calendar entry on the owner's
    side makes it real).
46. The `#wp-palette` disclosure pin + notes-nudge pins: re-run the served-
    bytes greps once before the next release (the three pin sites rule).
47. `go.mod` templ CLI pin reminder: the wrapper guards it; delete the
    AGENTS stale-binary note if the devshell templ ever catches up to
    v0.3.1100+ (the wrapper becomes a no-op guard — keep or retire).
48. The `93111401`-class daemon sweeps of dprint-formatted files: consider
    the pre-sweep build gate ask (pack D5) — it is the same pma bucket as
    the push leg.
49. Webphone-Android README honesty markers: refresh after the first device
    train (the seven legs each flip a marker).
50. The 2027-01-11 dedup quarterly + the modularize re-review: both sit on
    the calendar; nothing owed before then.

## g) Three questions I cannot answer myself

1. **Should the second CI-heal session stand down, or take a disjoint
   scope?** Two sessions overlapped this train again (its reports landed
   mid-train; its report edits touched my gate state once via the corpus).
   My recommendation is the pack's C10 rule, but only you know whether the
   other session's mandate (CI heal + BuildFlow repo) is done.
2. **Do you want the docs-gates issue POSTED after folding the fifth
   defect, or held for the sitting?** It is your own repo and an
   unsolicited AI-found filing — the banner awaits your tick; I defaulted
   to draft-only.
3. **The sitting pack's F-bulk: is "all yes" acceptable as a default?** 38
   rows where the recommendation is the status quo — if you trust the
   defaults, one line closes them; if not, name the exceptions and I will
   split them into individual pack rows for the next sitting.
