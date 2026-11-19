# Status report — cqrs-lint scoped-run verification, 11/11 findings confirmed as the armed keep-visible set (2026-11-19 19:53 CET)

> Claims verified as of 19:28–19:53 CET. Point-in-time snapshot; push/CI
> statements rot fastest. No Go code changed in this session; the only
> working-tree deltas are this report, the generated status index, and one
> comment-only `.buildflow.yml` ruling-record edit.

**Scope line — what this session VERIFIED, and how.** The owner pasted a
`buildflow -s cqrs-lint --format finding` run: 11 findings (2 warnings, 9
info), exit 0, **result cache 100% hit** — and this repo's documented trap
is exactly that a scoped verdict can be a cached set. The session proved
the findings live, reconciled them against every ruling registry, read all
9 finding sites in code, and adjudicated each. Verdict: **all 11 are the
documented armed keep-visible set, byte-for-byte** —
`.buildflow.yml` EXPECTED SCOPED COUNTS (2026-11-12) pins `cqrs-lint 11
(C014x2, D006, E023x4, F009, F021x3)` and today's run matches 11/11.

**Verification trail (evidence, not trust):**

1. **Cache honesty**: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s cqrs-lint
   --format finding` (19:33) returned the identical 11 finding IDs — the
   cached set was truthful.
2. **Registry reconciliation**: `.buildflow.yml:184` expected-counts pin ==
   today's set exactly; the 2026-11-11 triage report ruled 12 active then
   (`C014×2, E023×5, F021×3, F009, D006`), and E023 has healed 5→4 since.
3. **Site-level reads, 11/11**: every flagged site was read in code and
   matches its recorded rationale (table below). Webphone's own templ pin
   is intact (`go.mod`: `github.com/a-h/templ v0.3.1100`) — the doctor
   fingerprint drift (templ 0.3.1123, statix missing, nix 2.34.9) is
   injected by the **BuildFlow repo's dirty rebuild** (`result` →
   `buildflow-…-dirty`; concurrent session), not by webphone.
4. **Final scoped re-run** (19:44, the owner's exact command): same 11
   findings, exit 0.

**Explicitly NOT verified, and why:** the raw cqrs-lint CLI as an
independent third leg (the triage report's own method; I ran the BuildFlow
wrapper twice instead — the two legs I did run agree with the pinned
expected counts, but a raw-CLI leg would remove the stale-binary variable
the preflight itself warned about: binary 04fc6fb vs BuildFlow HEAD
48e854791); anything inside `/home/lars/projects/BuildFlow` (its tree is
DIRTY — `internal/cli/compliance_view.go`, `vendorHash.nix` — a concurrent
session owns it; I deliberately did not rebuild, which would have shipped
their in-flight bytes); the full buildflow battery; CI state (no push of
code from this session — and per the 19:47 sister report the tip
`2a46a9c6` carries another session's RED catalog job, run 44563793621,
which I confirm non-authorship of).

**Adjudication table (all 11):**

| #  | Finding                                   | Site reality (read this session)                                                                                                      | Ruling (one home)                                                       |
| -- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| 1  | C014 app.go:190 `time.Local = loc`        | Boot-time application of the validated `cfg.Timezone` (plan T26d) before any server starts                                            | keep-visible: deliberate boot zone (`.buildflow.yml` FULL TRIAGE block) |
| 2  | C014 reports.go:47 `Location: time.Local` | The Reports window rides the deployment zone the boot set                                                                             | keep-visible: deliberate deployment-zone windows (AGENTS.md § Reports)  |
| 3  | D006 flightrecorder.go:45 sentinel        | `errTraceRequestFailed` is the recorder trigger vocabulary's error signal (used at :134, `tc.Err`)                                    | keep-visible: recorder trigger vocabulary, erraudit-clean               |
| 4  | E023 chat.go:5                            | `db *sql.DB` is the `appFlagCursor` — matrix.sync_cursor via `store.FlagValue`/`SetFlag` in webphone.db (SQL lane outside the engine) | keep-visible: rule's own non-CQRS exemption class                       |
| 5  | E023 cqrs.go:13                           | database/sql for the DLQ handle (`dlqDB`)                                                                                             | keep-visible: DLQ engine-tier exempt — the rule's own text              |
| 6  | E023 cqrs.go:23                           | `_ "modernc.org/sqlite"` registers the driver FOR the DLQ handle                                                                      | same as #5                                                              |
| 7  | E023 server.go:11                         | database/sql = the sessions store (prod `NewSQLiteStore`)                                                                             | keep-visible: sessions stay SQL per the D2 ruling                       |
| 8  | F009 visibility.go:109                    | The deadline-bounded exponential poll inside the D7 `awaitOwnWrite` budget — in-process, not a durable timer                          | keep-visible: in-process D7 budget, scheduling module declined          |
| 9  | F021 threadsviews.go:78                   | 4 folds on threads — one per domain event (appended/flag/unread/deleted)                                                              | keep-visible: deliberate view decomposition                             |
| 10 | F021 threadsviews.go:96                   | 3 folds on messages                                                                                                                   | same                                                                    |
| 11 | F021 faxesgate.go:88                      | 3 folds = the fax lifecycle (receipt/advance/expunge), "the folds' one home"                                                          | same                                                                    |

## a) FULLY DONE

1. **Cache-trap check** — the documented scoped-verdict trap was honored
   first: no-cache re-run proved the 11 findings live (19:33).
2. **Full ruling reconciliation** — `.cqrs-lint.json` (11 permanent
   config-disables read), `.buildflow.yml` keep-visible block + expected
   counts, the 2026-11-11 triage report (50 suppressed / 12 active then,
   adjudication table). Three registries agree; nothing drifted.
3. **All 9 finding sites read and confirmed** against their rulings —
   zero sites had drifted from the recorded rationale; zero NEW findings
   beyond the armed set; E023's 5→4 heal is consistent with the 2026-11-12
   pin.
4. **Triage report consideration #14 closed** — the two C014 site
   rationales are now split in `.buildflow.yml` (app.go = boot-time zone
   set; reports.go = deployment-zone windows), exactly as the triage
   report's open item asked. Comment-only edit; no ruling changed.
5. **Buildflow housekeeping** — VACUUM executed on the 1.04 GB state DB
   (result: 1019 MB — the bytes are LIVE history rows, not free pages;
   `buildflow --help` exposes no purge CLI, so the size is a BuildFlow-side
   question). No concurrent buildflow process was running (checked first).
6. **Format gates green** — `nix fmt` (0 changed), then the authority
   `scripts/format-check.sh` (treefmt-check derivation built green).
7. **Final verdict** — the owner's exact command re-run at 19:44: same 11
   findings, exit 0.

## b) PARTIALLY DONE

1. **"fix it all" itself** — delivered as _verify-and-adjudicate_: every
   finding is now individually verified deliberate at its site. The
   _zeroing_ path (per-site directives / config-disable / code fixes) was
   deliberately NOT taken: it would re-litigate five armed owner rulings
   ("must stay armed"), and the doctrine forbids silent re-litigation.
   Blocked on an owner call (see g1).
2. **State-DB size fix** — VACUUM ran (the warning's documented fix) but
   could not shrink a DB whose bytes are live rows; the real lever
   (retention/purge) does not exist in the CLI and is BuildFlow-repo work.
3. **BuildFlow-side hazard mitigation** — diagnosed (binary staleness +
   fingerprint drift both trace to the concurrent session's dirty
   rebuild) but not fixed: no guard added, no upstream issue filed, and
   the owning session was not disturbed. Report-only by design.
4. **This report's docs-gates verification** — index regen + gates check
   executed at write time; the outcome is recorded in the chat handoff and
   any follow-up commits.

## c) NOT STARTED

1. The **raw cqrs-lint CLI cross-check** (third, wrapper-independent leg).
2. **Buildflow rebuild + reinstall** (`nix build . && nix run .#reinstall`
   in the BuildFlow repo) — blocked until the concurrent session lands and
   its tree is clean.
3. **Upstream BuildFlow asks** (see f rows 4–7) — none filed yet.
4. **Zeroing the armed findings** — blocked on g1.
5. **Deprecated `--go-upgrade-blame all` source hunt** — the doctor
   failure-cache says the alias was used once (last 2026-11-17); it is NOT
   in webphone's `scripts/` or `.github/`; origin unknown.
6. **Healed-E023 archaeology** — which site the triage's fifth E023
   finding lost, and when (one `git log -S` away; not run).

## d) TOTALLY FUCKED UP

Nothing damaged: no reverts, no lost work, tree clean except my own
documented edits, every gate I ran ended green. Honest stumbles, for the
record:

1. **Wrong order of research.** I read the finding SITES before the
   RULING registry. The `.buildflow.yml` EXPECTED SCOPED COUNTS section —
   which answers the entire adjudication in one line — was found only in
   the second round. Two rounds re-derived a fact that was already pinned.
2. **Pipeline `$?` masking.** `pgrep -fa buildflow | head -5; echo $?`
   read `head`'s exit code, not pgrep's — the same class as the repo's
   documented PIPESTATUS lesson (bash-only, unavailable in this shell).
   Harmless here (empty output was itself the signal), but sloppy.
3. **Ran VACUUM expecting a shrink.** The warning offered VACUUM as THE
   fix; I executed it before asking what makes the DB 1 GB (live run
   history). Verified after instead of predicting before — cost was one
   redundant step, not data.
4. **Fix-first reflex.** I drafted a `fmt.Errorf → errors.New` swap for
   the D006 sentinel before the arming block vetoed it (the triage ruled
   that site keep-visible ON PURPOSE). Caught before landing, zero
   damage — but the reflex is exactly what ruling-first reading prevents.

## e) WHAT WE SHOULD IMPROVE

1. **Ordering doctrine for scoped lint verdicts**: read the ruling
   registries FIRST (expected counts → config-disables → directives),
   sites second, cache-bypass third. Worth a `docs/lessons.md` row.
2. **Mechanical expected-count replay**: a `scripts/audit-verify.sh`-style
   check that re-runs the scoped lint and diffs finding IDs against the
   armed set, so drift is caught by the machine, not by a session's memory.
3. **Three-leg standard for scoped verdicts**: wrapper no-cache + raw CLI
   (+ the pinned expected counts as the third leg). The triage report
   already used the raw CLI; future sessions should too.
4. **BuildFlow UX**: scoped runs should print "expected N (armed set)
   vs found M" from the repo's own pin — upstream ask material.
5. **Keep expected counts fresh**: the 2026-11-12 pin is only useful if
   every adjudication updates it (it did — that is why today's check was
   a one-liner once found). Ratchet discipline applies.
6. **Concurrent-session awareness paid off**: the dirty BuildFlow tree
   correctly stayed untouched; the fix-forward protocol (confirm
   non-authorship, leave the surface) was applied to the RED catalog CI
   job and the binary freshness advisory.

## f) Up to 50 things we should get done next

`[NEW]` = born from this session. `[LEDGER]` = already tracked; home named
so nothing is re-derived. Ordered roughly by impact.

| #  | Item                                                                                                                                                                           | Source / home                                                  |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------- |
| 1  | Owner ruling: zero the 11 armed cqrs-lint findings (directives/config/fixes) vs ratify keep-visible — the standing blocker on this whole class                                 | `[NEW]`, blocked on g1                                         |
| 2  | Record the "ruling registry before sites" ordering lesson in docs/lessons.md                                                                                                   | `[NEW]`                                                        |
| 3  | Mechanical cqrs-lint expected-count replay (diff scoped finding IDs vs the armed set)                                                                                          | `[NEW]`                                                        |
| 4  | BuildFlow upstream ask: scoped runs self-print expected-vs-found from the repo pin <!-- fresh:skip -->                                                                         | `[NEW]`                                                        |
| 5  | BuildFlow upstream ask: PATH precedence guard — provider binaries must not shadow a devshell's pinned toolchain (templ 0.3.1123-over-0.3.1100 codegen class)                   | `[NEW]`                                                        |
| 6  | BuildFlow upstream ask: state-history retention (purge CLI / GC knob) — 1 GB live rows, no lever                                                                               | `[NEW]`                                                        |
| 7  | BuildFlow upstream ask: scope the baseline-staleness alarm to the audited tool (scoped-run noise by design)                                                                    | `[NEW]`                                                        |
| 8  | Rebuild + reinstall the buildflow binary AFTER the concurrent session lands (tree clean, ≥48e854791)                                                                           | `[NEW]`, blocked                                               |
| 9  | Hunt the deprecated `--go-upgrade-blame all` invocation (last 2026-11-17; absent from webphone scripts/CI)                                                                     | `[NEW]`                                                        |
| 10 | Healed-E023 archaeology: identify the triage's fifth E023 site and the commit that healed it                                                                                   | `[NEW]`                                                        |
| 11 | Optional: localize dprint plugins (`nix build .#dprint-plugins` + `BUILDFLOW_DPRINT_PLUGINS`) to kill the cold-fetch hang class                                                | `[NEW]`, optional                                              |
| 12 | `io-history-ratchet` advisory: one `--profile-io` comparison vs the last green run when convenient                                                                             | `[NEW]`, advisory                                              |
| 13 | Verify the 2 go-auto-upgrade failure-cache suppression entries still apply at the next full run                                                                                | `[NEW]`, advisory                                              |
| 14 | One-line AGENTS.md pointer to the `.buildflow.yml` expected-counts pin (findable without opening the YAML)                                                                     | `[NEW]`, tiny                                                  |
| 15 | internal/server carve — the below-module refinement                                                                                                                            | `[LEDGER]` TODO_LIST (one home)                                |
| 16 | html-hub carve residual — depthbudget named pin, ratchet direction DOWN                                                                                                        | `[LEDGER]` TODO_LIST carve row                                 |
| 17 | D4 admin write-shaped delegation (acting happens per-line; only reads today)                                                                                                   | `[LEDGER]` TODO_LIST                                           |
| 18 | metaengine nested addressing: per-CALL transcript floor needs owner+call caps — the only remaining WithCapPerGroup reopen trigger                                              | `[LEDGER]` TODO_LIST reopen registry                           |
| 19 | go-appkit asks #59/#60/#61: adopt when upstream ships (webphone migrates LAST)                                                                                                 | `[LEDGER]` docs/review/libs/go-appkit.md                       |
| 20 | appkit/flightrecorderhealth v0.1.7 candidate (health-batch trace snapshots)                                                                                                    | `[LEDGER]` demand-gated TODO_LIST row                          |
| 21 | Session protocol NO-GO reopen triggers                                                                                                                                         | `[LEDGER]` TODO_LIST                                           |
| 22 | C044 prefixed-key → bare-value migration train (versioned data migration; priced, not a lint cleanup)                                                                          | `[LEDGER]` .cqrs-lint.json reason                              |
| 23 | D005 re-enable when upstream token-attachment fix ships                                                                                                                        | `[LEDGER]` .cqrs-lint.json reason                              |
| 24 | F005 upcaster registry at the FIRST real v2 event                                                                                                                              | `[LEDGER]` .cqrs-lint.json reason                              |
| 25 | F58 drain train: annotate_max 260 back DOWN                                                                                                                                    | `[LEDGER]` docs-gates ratchet                                  |
| 26 | audit_max 72 ratchet down + re-derive generic_df from the living-line count at the next sweep (rule ~0.5%, never copy the number)                                              | `[LEDGER]` docs-gates                                          |
| 27 | Monthly tier re-measure (errorfamily tiers 1+2 at `--max --no-suppress`) — due 2026-12-08 per watch calendar                                                                   | `[LEDGER]` watch calendar                                      |
| 28 | Quarterly art-dupl baseline audit (regen ONLY with CI-pinned detector v0.7.2)                                                                                                  | `[LEDGER]` dedup doctrine                                      |
| 29 | workflow-e2e CI job: advisory → required once clean trains prove it                                                                                                            | `[LEDGER]` CI posture                                          |
| 30 | Stack browser E2E ~90s transfer-step flake: stack-side diagnosis (re-run-once policy stands)                                                                                   | `[LEDGER]` AGENTS.md § stack E2E                               |
| 31 | RED CI job on tip: "catalog export matches the hub frontmatter contract" (run 44563793621, commit 2a46a9c6 — another session's); watch for its fix, tracker-issue if it stands | `[LEDGER]` 19:47 sister report                                 |
| 32 | readmodels.db inert orphan on upgraded deployments — backup-set note (A43) stands; operator-docs cleanup eventually                                                            | `[LEDGER]` ladder retirement doc                               |
| 33 | Attended transfer M19 NO-GO — reopen only with an FS lab                                                                                                                       | `[LEDGER]` docs/research/2026-10-27_attended-transfer-spike.md |
| 34 | Push ring end-to-end: consuming-stack FS dialplan must POST /hooks/ring (stack-repo work, NOT webphone)                                                                        | `[LEDGER]` AGENTS.md § push seam                               |
| 35 | webphone-android P1 shell: registration via the CSRF dance + ring wiring                                                                                                       | `[LEDGER]` sibling repo                                        |
| 36 | Bounded AuditLog projection consumption — stays an owner call (usermgmt projection UNWIRED by choice)                                                                          | `[LEDGER]` AGENTS.md § owner decisions                         |
| 37 | Fleet-tags sweep cadence: keep every scored family at upstream latest (last full re-verify 2026-11-13)                                                                         | `[LEDGER]` docs/review/libs/                                   |

## g) Questions for the owner (not answerable from webphone)

1. **Does "fix it all" override the armed rulings?** All 11 findings are
   the documented keep-visible set ("must stay armed", 2026-11-11 triage).
   Zeroing them (per-site directives, config-disable, or the errors.New
   swap) re-litigates those rulings; verification-as-delivered does not.
   Which do you want?
2. **BuildFlow state DB at ~1 GB of live run history**: acceptable, or do
   you want a retention policy (a BuildFlow-side purge CLI/GC ask — I
   found no such lever in the binary)?
3. **The BuildFlow PATH/fingerprint hazard** (its dirty rebuild injects
   templ 0.3.1123 over webphone's pinned v0.3.1100 for out-of-shell runs
   — the RED MAIN codegen class): file the upstream PATH-precedence issue
   NOW, or wait for the in-flight BuildFlow session to land first?
