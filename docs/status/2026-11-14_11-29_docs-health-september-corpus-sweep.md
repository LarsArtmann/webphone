# Status: docs-health September-corpus sweep (every live 2026-0* file + the six living docs)

**Snapshot:** claims verified 2026-11-14 11:29 CET at HEAD `8f3520bb`
(= `origin/main`, CI GREEN via `scripts/ci-verdict.sh`, rc=0 — the
workflow-name-verified verdict, not the watch lane). Session scope: the
owner-directed docs-health AUDIT over ALL `**/2026-0*` files (155 found:
39 live September artifacts viewed in full — 14 `.md` read whole, 18 HTML
reports text-extracted including the one missing a `<body>` tag, 7 d2/svg
inventoried — and 116 archived files adjudicated mechanically by the
battery per the closed-history doctrine) plus the six living docs.
One correction landed mid-report: my first CI-class claim for the
2026-10-27 sha was an inference and the API proved it wrong — fixed in
the archived file (see §d1). Battery 8/8 PASS at HEAD; three small edits
(this report, the stale-diagram TODO row, the archived-report
correction) sit in the worktree for the daemon.

## a) FULLY DONE

1. **Red main root-caused and healed.** CI on today's two daemon pushes
   (f83ba3cf, b09f70c5) failed the docs-cadence gate (66 unarchived
   reports > 65). The archive ritual fixed the exact class: the
   2026-10-27 multicall report — the oldest DIRTY candidate — was
   adjudicated and archived gates-first; live corpus 66 → 65; battery
   green; pushed as `8f3520bb`; CI concluded success (rc=0, poll 4).
2. __Every live 2026-0_ file viewed_* (the owner directive): 14 `.md`
   full reads (4 announcement drafts, health-probes memo, sessions
   at-rest review, owner briefing, owner command sheet, hub-fanout +
   coverage baselines, 2 reviews), 18 HTML reports text-extracted in
   full (the 2026-09-18 deep-dive extracts zero chars because it lacks
   a `<body>` tag — extracted from `</head>` instead), the d2/svg set
   inventoried with dates.
3. **The multicall report's 50 §f items adjudicated against code**, not
   memory: 20 verified done (config fail-closed typing at
   `internal/config/config.go:918-927`, the bounded `Transcripts.Call`
   scan at `internal/cqrs/transcriptsstore.go:207`, the toast alias,
   apidocs idem/503 texts, the `postcall.js` focus helper, FEATURES'
   ring-queue row, the retired buildflow baseline, the v8-sweep archive,
   the devshell templ pin, the fleet lessons entry), 20 routed (14 to
   TODO rows minted this session, 6 to standing rows), 4 wont with
   reasons, the rest confirmed closed in-train.
4. **76-item annotate spec applied atomically**: `status-items --verify`
   resolved all 76 keys before any write; the §a table struck per-cell
   (11 rows, formatter-stable — re-proven after the sibling session's
   dprint pass re-aligned the columns); `check-rows --strict` +
   `bare-rows` green before AND after formatting; banner + `git mv` +
   `INDEX.md` regen (65 reports).
5. **September reviews annotated where code had moved past them**: the
   09-30 data-model review's migration steps 2+3 (config typing,
   schema_version ladder) and the 09-30 architecture review's four
   roadmap rows all carry `[Resolved …]` blocks citing today's file
   evidence; the 09-19 cqrs-htmx deep-dive carries the supersession
   banner (the 2026-10-25 round-3 audit is the state of record); the
   09-18 structural-health roadmap carries its resolution paragraph.
6. **Owner-facing docs de-conflicted**: the 2026-09-22 owner briefing
   and the command sheet's §5 now point to the 2026-11-07 sitting pack
   (the ONE answer surface since 2026-11-08); the health-probes memo
   records the 2026-11-13 `unsafe-eval` retirement (dashboard
   v0.13.3's data-nonce root); the sessions at-rest review carries its
   fired re-review trigger (`login_email` landed 2026-10-28).
7. **AGENTS.md ratchet prose corrected with provenance**: the file's
   pins are `annotate_max` 260 / `generic_df` 27; the prose claimed 203
   / 22. The 203→260 raise was deliberate and evidenced (`5e4dc514`,
   2026-11-12 DF-brittle flare, measured 249) — but a later daemon
   commit stripped the evidence key from `.docs-gates.json`, so the
   AGENTS sentence now carries the surviving record and the re-derive
   rule.
8. **Coverage baseline debt paid** (the file's own per-release rule owed
   v2.12.0, v2.13.0, v2.14.0): full `-cover` run green, island suite
   340/340, dated section appended with the movers table — the
   `readmodels` 73→11 collapse and the `store` 82→71 giveback are
   STRUCTURAL (the ladder retirement moved engine + tests to `cqrs`
   at 78.9% and deleted the legacy importers), not regressions.
9. **Nine candidate living-doc lines pre-flighted** before landing (the
   additions-only model): one reword required — the token
   `credentials` collided with an OIDC-lane archived item — and the
   re-run came back SAFE.
10. **Toolchain**: the docs-health annotate binary built from the skill
    assets per the AGENTS recipe, self-test green, used for every
    strike (no hand-rolled sed).

## b) PARTIALLY DONE

1. **The e6da8ab CI verdict (§f1 of the archived report)** — the
   conclusion (RED at the sha) was verified, but my first annotation
   GUESSED the failing class ("docs-cadence") from today's identical
   red. The API check (asked for by this report's §d review) shows the
   real failing job was **`art-dupl check (new clones only)`** —
   healed 2026-11-11 when the templ clones dissolved at their sites.
   All three affected verdicts in the archived file now carry the
   `[CORRECTED 2026-11-14 11:29 …]` inline amendment. The verified
   fact set is complete now; the lesson is §e2.
2. **The archived 116 files**: adjudicated by the battery's
   annotate/audit-all/routing gates (all green, 210 findings tolerated
   under the pinned 260) — deliberately NOT read file-by-file per the
   skill's closed-history rule. Any file a gate flags gets opened; none
   did.
3. **README + FEATURES + ROADMAP + CHANGELOG audits**: verified by
   targeted spot-checks (dashboard v0.13.3 CSP row, textagent lane,
   ring-queue shortcuts, `/login` front door, append-only
   `[Unreleased]`), not claim-by-claim over every row — the battery's
   freshness/audit-all replay is the standing mechanical net for the
   rest.
4. **AGENTS.md**: the ratchet drift was found and fixed and the 377-line
   cap holds (215), but the file's remaining ~200 lines were not
   re-verified claim-by-claim this session (the file is the repo's
   memory; a full verify sweep is a train of its own).
5. **The sitting pack (2026-11-07) folding of briefing rows**: the
   supersession banners now point there, but I did not re-verify that
   all 38 briefing rows actually landed in the pack's §F table (the
   pack post-dates the briefing; the TODO row asserts the fold).

## c) NOT STARTED

1. **The work the six minted TODO rows describe** (pill trio, notes
   cluster, summarize cluster, smoke probes, typed textagent lane,
   sessions re-review, coverpkg step) — routing only; zero code moved.
2. **The F58 archived-corpus drain** (the annotate pin's way back down
   from 260) — the T24 release-tails row owns it; not started here.
3. **QMD indexing, recordings-by-Q4, XFF flip, sniff-fallback delete,
   TEMP-DIAG** — owner calls / stack-side, homed in ROADMAP/the pack,
   untouched (correctly — they are not mine to decide).
4. **The architecture-understanding diagram refresh** — noticed during
   the corpus read (the 2026-09-30 "16 packages" pair predates the
   ladder retirement, push/chat packages and the server carve), minted
   as a Low TODO row; not started.
5. **A full claim-by-claim AGENTS verify sweep** — see §b4; not
   started.

## d) TOTALLY FUCKED UP (honest list)

1. **I wrote an inferred CI class into a permanent annotation.** The
   §b1/§f1/§g1 verdicts in the archived multicall report claimed the
   e6da8ab red was "the docs-cadence class" — an inference from today's
   identical-looking red, stated as fact. The API proved the failing
   job was `art-dupl check`. Caught during this report's §d review and
   corrected inline with `[CORRECTED …]` blocks, but the first version
   shipped to the archive committed at `8f3520bb` was wrong for ~90
   minutes. The annotation-rule consequence: verdicts must cite the
   VERIFIED fact (the conclusion), never the extrapolated cause.
2. **Three failed tool dry-runs before reading the source.** `rows`
   mode rejected `1` and `M11` and the spec rejected `**M11 …**` before
   I opened `annotate.go` and learned the grammar (row-id = the literal
   first-cell text; bold key-rows only). Reading the tool's usage block
   first would have cost one call, not four.
3. **The §a strike script aborted on padded cells.** I hard-coded the
   header's padding (`**M11 summarize backend**     |`) instead of
   regex-matching cell boundaries; the assertion abort was atomic (no
   damage) but the retry cost a round trip. Exact-match editing is for
   the edit tool, not for generated table bytes.
4. **The pre-flight needed a reword I should have predicted.** I knew
   the DF-brittle generic-token rule (AGENTS carries it) and still
   drafted the sessions row with the token `credentials`; the gate
   caught it, but the rule says draft with rare tokens from the start.
5. **An invisible annotation.** The structural-health resolution first
   landed as an HTML comment — unreadable to a browser reader, which
   defeats the entire annotate doctrine ("a reader scans and sees").
   Converted to a visible `<p>` immediately; should have been visible
   by construction.
6. **One typo crashed the HTML extraction mid-batch**
   (`…deep-dive.html2` in a dict), costing a re-run of four tails; and
   the first extraction truncated at 7 000 chars per file, requiring a
   second pass for the tails I then needed anyway — the cap saved
   nothing.
7. **The coverage debt accumulated at all.** Three releases shipped
   without the baseline's own re-run — the ritual was written down and
   still slipped; this session paid it, but the debt existing is the
   failure (a watch-calendar entry would prevent a repeat).

## e) WHAT WE SHOULD IMPROVE

1. **Verdict discipline: cite the verified fact, nothing more.** The
   §d1 correction is the standing example — `ci-verdict.sh` answers the
   CONCLUSION; the failing JOB needs the check-runs API, and an
   annotation that names a cause must have run that call.
2. **Read the tool's usage/source before the first dry-run**, not after
   the third failure — the annotate grammar, the per-cell strike rule
   and the `--verify`-then-read-in-full protocol are all written down;
   the failure was skipping the reading order.
3. **Draft §f/new-row tokens rare from the first keystroke** — the
   pre-flight exists to catch collisions, but drafting with quoted
   phrases from the matching TODO row (the AGENTS corollary) avoids the
   round trip entirely.
4. **A watch-calendar entry for the coverage baseline** (the file says
   "re-run at every release" — that is exactly the class the watch
   calendar exists for; the erraudit/buildflow entries proved the
   pattern works).
5. **Archived-report banners should claim only what the gates
   certified** — the v8-era banner wording "every table row is struck
   or routed" over-claimed for its file (no `~~` markers exist in it);
   today's banner states exactly what was done. Future archive banners
   copy today's form.
6. **The battery's annotate pin re-measure** belongs at every sweep
   (the corpus grew again today: this report + one TODO row; the
   measured count was 210 tolerated of 260 — re-derive, never drift).

## f) Up to 50 things we should get done next

Priority-ordered; ★ = needs the owner or another repo. Tokens bind to
their TODO_LIST/ROADMAP homes (rows 1-7 minted today; the rest carry
their standing homes).

1. **Sessions at-rest re-review** (owed by the 2026-09-30 review's own
   trigger — `login_email` landed 2026-10-28 after its snapshot).
2. **Summarize micro-cluster** (hide the ended-card Summarize button
   when ASR is off; `agentAudit` for the summarize producer lane; the
   deleted-transcript idempotency pin; `error-context.py --self-test`
   - error-contract prose for the `push.draft_*`/`agent.summarize`
     families; ring_test records drafts natively; the duplicate-replay
     no-second-push pin; the server test-harness extraction; the FCM
     `SendDraft` wire test).
3. **Notes micro-cluster** (explicit note-slot delete API, the
   glanceable "noted" chip on History rows, note bodies in the
   transcript search corpus, the Settings 2000-cap visibility row).
4. **Agent-pill UX trio** (per-mode aria/tooltip semantics, pill
   refresh on `wp:drafts-changed`, a cycle keyboard shortcut).
5. **Smoke boot probes + lychee excludes** (grep the `agent` config
   flag at boot, round-trip the notes API on a fresh binary, exclude
   the archived docs trees from the lychee preflight).
6. **Typed `settings.textagent.*` module lane** (the summarize chain
   made the seam first-class; the NixOS module has no typed option).
7. **The `buildflow -coverpkg` union coverage step** (the mechanical
   half; the policy question sits in ROADMAP).
8. **F58 — the full archived-corpus drain** (the annotate pin 260's way
   back DOWN) before the ~2026-12-07 cadence window ★priority call.
9. **Cadence drain continuation + ratchet re-measure** at the next
   sweep (the DIRTY candidates list, annotate/audit pins re-derived).
10. **NixOS VM test for the OIDC lane** (LoadCredential lands, mode
    0400, the app reads it at login).
11. **Second module-check golden with `oauth2.enable`** (exact rendered
    JSON beside the zero-config golden).
12. **Typed passkey options** in the module (derived `rp_id`/origins for
    passkey-only configs).
13. **Stack drift audit** — diff the stack's webphone module assertions
    against `internal/config` (rides the next stack module touch) ★.
14. **Record the CUE ruling durably** (REJECTED; schema-export deferred
    with a named trigger) as an owner decision line ★.
15. **Schema-export spike** — settings schema GENERATED from the config
    structs.
16. **Delete the dead sibling legacy readers** (`legacy_faxes.go` et al.;
    keep the app-flag helpers).
17. **Design the §5 export-or-confirm tooling** before any DROP window.
18. **T05 — the authenticated prod capture** (`nix run .#ui-track --`…,
    then the ranked ugliness inventory → T06/T07 calibration) ★.
19. **Answer the sitting pack** (58 rulings; closes the OWNER-calls
    batch row and the superseded briefing/command-sheet pointers) ★.
20. **Voicemail stack redeploy** riding `main`, then click play and
    record the toast-named verdict ★.
21. **File the suppression state-machine draft upstream** (owner tick
    gates it) ★.
22. **File the D005 first-token draft upstream** (same gate) ★.
23. **Confirm the enroll.go S001 heuristic change upstream** ★.
24. **`erraudit nolint-audit` decay pass** after the C007/At-command
    train.
25. **The 2026-12-08 erraudit/buildflow calendar re-measure** (tier 1+2).
26. **AGENTS.md b5 one-liner** (per-site cqrs-lint directive convention;
    owner call) ★.
27. **File the bandit stdout findings upstream** (owner tick) ★.
28. **Stale `# nosec B105` audit** (verify-only; touch-triggered).
29. **Ring-while-closed end-to-end** (stack dialplan ring hook + the
    android shell consuming it) ★.
30. **webphone-android first device train** (the scaffold vs live
    FreeSWITCH) ★.
31. **C5 trunk lift day** (matrix capture → first PSTN dial + CDR +
    recording → inbound green) ★.
32. **A3.4 — send the assembled Telnyx reply into ticket #832080** ★.
33. **C3.2 — one real inbound mobile call to the DID** (closes the
    zero-inbound ruling) ★.
34. **E3 — pick the `reports.weekly_talk_budget` number** and set it ★.
35. **E5.2–E5.4 — hikaru's passkey one-time link + enroll + the three
    Super Admin surfaces** ★.
36. **E6 — the first real outbound fax** (owner picks the destination) ★.
37. **Row-33/34/35 confirmations** (the `pifo42h1` recording, port 9100,
    internal test-destination names) ★.
38. **UI/UX Train 2** — per-surface desktop layouts under the shared
    shell (the ratified destination's next train).
39. **The stale architecture-understanding diagrams** (TODO row minted
    today; regenerate the D2/SVG pair vs the ~38-package reality).
40. **XFF flip decision** (ROADMAP standing watch `KeyExtractorFromClientIP`)
    ★.
41. **TEMP-DIAG keep/gate decision** (ROADMAP standing watch) ★.
42. **Sniff-fallback lifespan** — the delete-after-prod-confirm trigger
    has long fired (prod > v2.8.0 since 2026-11-06); stack-side ★.
43. **Recording UI product-intent decision** — briefing row 11 said
    decide by 2026-Q4, which is NOW ★.
44. **Domain-language glossary ratify-or-delete** (ROADMAP open
    question, memo on the file since 2026-10-14) ★.
45. **Coverage teeth policy** (ROADMAP open question: floor vs
    informational) ★.
46. **QMD indexing of webphone docs** (routed to ROADMAP today) ★.
47. **Attended-transfer GO day** (the ROADMAP watch: FS lab + Replaces
    VM scenario before any island consult-flow work) ★.
48. **Summarize UX polish** (ROADMAP: drafts-panel deep link, ended-card
    stack cap, the `duplicate:true` original-draft-id wire idea).
49. **Re-verify the sitting pack's §F table against the briefing's 38
    rows** (§b5 above — confirm the fold no row behind).
50. **docs-health sweep ritual, next pass** (the battery + annotate pin
    re-measure + the cadence-candidates list; keep the queue at ≤65).

## g) Three questions I cannot answer myself

1. **F58 drain priority** — the annotate pin sits at 260 (raised with
   evidence on 2026-11-12; measured 210 today). Do you want a dedicated
   archived-corpus drain train to walk the pin back down BEFORE the
   ~2026-12-07 cadence window, or continue tranche-by-tranche with the
   cadence sweeps and accept the raised pin until it drains naturally?
2. **Prod currency** — the command sheet last verified prod at v2.9.0 on
   2026-11-06; the stack rides webphone `main` per-train. Is prod
   expected to be on v2.14.0 already, and do you want a
   `--expect-version 2.14.0` prod smoke added to the next sweep's owner
   legs (I cannot probe the deployment from here)?
3. **Recordings decision deadline** — briefing row 11 committed to a
   product-intent decision by 2026-Q4 (this quarter): build the
   webphone recording panel (consent/access model first), or re-affirm
   stack-level-only and record the NOT-DO?

_Report ends. Awaiting instructions._
