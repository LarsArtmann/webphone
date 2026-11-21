# Status — the 20-skill battery session (second run): final comprehensive report

> ARCHIVED 2026-11-21 (docs-health v10 sweep): pre-drain verified (check-rows --strict + bare-rows green), drained for cadence — the pin sat at 66/66 with 12 gates-passing candidates

> SNAPSHOT-ROT HEADER: claims verified as of **2026-11-20 15:25 CET**.
> The session's WORK window was **2026-11-19 ~19:55 → ~21:40 CET** (single
> continuous run; the machine slept, and this report was written on resume
> at 15:20). Overnight and this morning, OTHER sessions (the pareto train,
> T01–T27) moved HEAD from `40a86fc0` to `30057a6e` — I have NOT re-verified
> the tree past my 21:40 last-known-good state (docs-gates full PASS,
> `go test ./...` all 35 packages green, audit-verify 106/106); everything
> below describes MY session's run and what I noticed, not the overnight
> trains. The final CI verdict for my push was never read (see §d).

**Scope line.** The owner ordered all 20 skills "SUPERBLY fully and to the
MAX". Three overlapping sessions ran battery-class work on this repo that
same day (morning review train 07:41–08:40, an afternoon release-readiness

- CQRS-assessment + evening triple-review chain, and THIS session). This
  session's distinctive contribution: the mechanical gates nobody else ran
  (full no-cache buildflow, audit-verify replay, erraudit/art-dupl/banned-libs
  direct, banned-lib adjudication), a VERIFICATION pass over the morning
  artifacts, a 15-fix code train, and the docs-gates/TODO_LIST/doc-hygiene
  closeout.

## a) FULLY DONE

1. **All 20 skills loaded and executed** per the one-home doctrine — fresh
   same-day reports got VERIFY passes, uncovered skills got RUNS. Per-skill
   verdicts (full table with evidence: `docs/status/2026-11-19_21-20_skill-battery-verification.md`):
   samber-do EXEMPLARY (all six anti-patterns absent, DO-1 arch-test-pinned);
   cqrs-htmx + go-cqrs-lite CONFORMANT (chain order, zero deprecated
   re-exports, system.New composition root, ONE CAS dedup home,
   WithCapPerGroup caps, single sanctioned floor loop); dedup CLEAN at the
   pinned v0.7.2 gate (0 new clones; the v0.9.1 red is the documented
   version-skew phantom); go-error-modernization FULLY CLEAN (0 Go
   findings, 0 cargo-cult `Is` hand-rolls); how-to-golang CLEAN (0 banned
   direct deps — 3 `// indirect` transitives adjudicated: koanf→fsnotify,
   watermill→pkg/errors, mapstructure unused); go-modularize RULING HOLDS
   (the second go.mod is the erraudit probe fixture — test data);
   nix-private-go-repos NOT APPLICABLE (verified: no vendor/, go.work,
   replace, or pseudo-version sentinels); bdd-testing posture CONFORMANT
   (2 suites, black-box, zero focus, one deliberate real-clock spec).
2. **The fix train — 15 fixes landed, all test-verified green**:
   (1) fax resend now owns its blob copy (fcr-02 M1, HIGH — data-integrity
   class) + `TestResendOwnsItsBlobCopy` pin; (2) `review.saved` joined the
   event catalog + `CatalogVersion` 1.0.0→1.1.0 + per-message
   `WithVersion` stamps (the library defaults messages to "1.0.0"
   regardless of the builder — all five events rode the stale stamp until
   the option was added) + the seam-only ruling codified + a
   pairs↔catalog count cross-check; (3) three pipe-into-tee Nix checks
   (statix/deadnix/vulnix-triage) could never fail — FIXED, checks rebuild
   green; (4) signal data-URI filename percent-escaped (F-G, inbound
   remote-controlled); (5) dead second parse in `sendMessage`; (6) MarkRead
   swallow → debug log; (7) transcriptsgate inverted comment; (8) apidocs
   stale response-header claim + the AGENTS.md bullet; (9) four dead
   zero-predicates deleted (Address/Phone/Token/RelPath `IsZero` — zero
   callers incl. tests and generated files); (10) fax `Receive` trust
   justification (L3's second sanctioned option); (11–12) naming small
   class: Mast*→Last*, `max`→`ceiling`/`peak`, `extensionIds`→`extensionIDs`;
   (13) 0400→0440 LoadCredential comment drift ×5 sites (vm-tests empirical
   pin), redundant mkIf, missing maintainers meta ×2; (14) AGENTS.md
   middleware-chain listing gains ReportingEndpointsMiddleware; (15) the
   templ-components icon-floor claim re-pinned 29→26 with dated rationale.
3. **Gates, final state at session end (21:40)**: `go test -count=1 ./...`
   all 35 packages green (twice); session/auth suite, arch tests (depth
   budget + DAG), vet green; `audit-verify` 106/106 after one stale-claim
   repair; `BUILDFLOW_NO_RESULT_CACHE=1 buildflow` — everything green
   except the file-size class (below); docs-gates **full matrix PASS**
   (format, status-f-section, battery, table-damage, strike-safety).
4. **The file-size blocker provenance PROVEN**: the 47 `go-file-too-large`
   error findings arrive from the BuildFlow side (provider registration
   changed 2026-11-16, `registration_2.go`) — four days AFTER v2.14.0's
   green preflight (2026-11-11) — not from the corpus; the §D10 owner call
   (sitting since 2026-11-08) now bites release.sh's buildflow step.
5. **Release verdict**: v2.15.0 (MINOR, per the 15:20 readiness report)
   folded-ready, BLOCKED on the §D10 disposition; the tag/dance stays
   owner-executed per that report's explicit boundary. No tag, no push of
   release material, no fold — exactly as the boundary requires.
6. **TODO_LIST discipline**: 5 harvest rows struck with evidence (M1, M2,
   Catalog T11 delta, apidocs, zero-predicates + inbound-fax cap), the
   naming row amended, 7 new rows appended (file-size disposition +
   release-blocker, vm-tests split, erraudit JS triage, deepening owner
   pick, bump-train handoff, go-structure-linter tail, F005, T01 stack
   lock, T18 FCM invalidateToken), and three audit-all flare rows routed
   (A30 → resolved-in-place annotation; CI-watch → struck; F005 → TODO)
   keeping the ratchet at its honest 72 — the concurrent session's raise
   to 75 was REVERTED DOWN once my routing made 72 true again.
7. **Docs hygiene closeout**: the oldest unarchived report (2026-11-05)
   archived GATES-FIRST (ledger marker, check-rows + bare-rows clean,
   git mv, banner) taking cadence 67→66; status index regenerated; CHANGELOG
   [Unreleased] carries the four behavior-bearing fixes; AGENTS.md carries
   the two durable lessons (WithVersion stamp gotcha; file-size gate
   provenance + release bite).
8. **Served-markup exemption (stated now, as the runbook requires)**: my
   train changed NO served bytes — no `.templ`, no island assets, no
   `app.css`; the changes are Go internals, catalog JSON, Nix checks, CI
   pins, and docs. The runbook's stack browser-E2E obligation is therefore
   EXEMPT for this train (I failed to state this at 21:40 — see §d).

## b) PARTIALLY DONE

1. **erraudit JS `.catch` triage** — 10 findings, one class, all carrying
   intent comments in a form the detector does not recognize; adjudicated
   and routed (directive-form row), NOT fixed: today not gating (buildflow's
   erraudit step reads Go scope only).
2. **Dependency bump trains** — the go-ecosystem-upgrade skill ran its
   assess/baseline/enumerate/classify phases (the 08:22 sweep verified:
   10 direct deps behind, all additive, reopen-registry NONE FIRED) and I
   verified its claims + version surfaces (go 1.27.3 patch floor is
   dep-imposed by loginpage v4.17.0); the four EXECUTION trains are armed
   with exact commands and handed off, deliberately not run in this session
   (each is a per-train push with CI-watch; folding them into an
   already-overloaded audit session would have been the F10 failure mode).
3. **Naming adjudication train** — the smallest-blast-radius class shipped;
   the four real renames (FindThread→ResolveThreadID, VoicemailItem,
   authmap.Map→Ownership, Shell×2) stay on their row (they need the
   central adjudication the review itself asked for).
4. **go-release** — readiness verdict only (see a5); the fold and the
   dance are owner-owned per the 15:20 report's boundary.
5. **The full-code-review unfixed tail** — of ~30 findings across the four
   morning parts, this session fixed 7 and routed the rest: FCM
   single-flight, blob brand-before-write, Seq-as-retention-clock, F-A doc
   comments, F-D mutex-across-network, F-F Fan panic pin, F1 stale arch-test
   subset, F3 trim duplication, F4 Seq retention clock, F5 whole-corpus
   reads, F7 boot-helper duplication, F13/F14 island file sizes.
6. **arch-verify depth** — the subagent died to rate limits and I absorbed
   it manually at reduced depth: scores + roadmap + top deepening
   recommendation extracted and key claims re-checked (depth-budget
   numbers, the deepening brief's candidate structure), but not every
   file:line citation in the morning reports was re-verified.

## c) NOT STARTED

1. Bump-train EXECUTION (Trains 1–4, per
   `docs/research/2026-11-19_dependency-currency-sweep.md`).
2. The v2.15.0 fold + release.sh dance (owner; blocked on §D10).
3. The §D10 file-size disposition itself (owner call).
4. vm-tests.nix split; the deepening grilling (owner pick gated);
   the systemscenario spike (owner-taste row); aarch64 cross-build and
   the standalone `nix run .#vulnix` (both release-time legs).
5. The final CI verdict read for this session's commits — the session
   ended before it, and the overnight trains have since pushed several
   times; whatever CI says NOW reflects their head, not mine (§d6).

## d) TOTALLY FUCKED UP (the honest list)

1. **The subagent rate-limit cascade**: I launched 10 parallel audit
   subagents against one provider; FIVE died to rate limits mid-flight
   and I re-absorbed their work manually. The fleet-deep-dive runbook I
   had literally loaded hours earlier documents the cap ("concurrency at
   3") and its ~40-minute cost. Pure self-inflicted waste.
2. **I nearly duplicated three concurrent sessions' work wholesale**: the
   morning git log already showed the review train's reports before I
   planned the fan-out; I adapted only after the first reports were
   dispatched. The correct move — shrink to verification + gaps — came
   one full planning cycle late.
3. **My status report shipped violating two gates I then had to repair**:
   it lacked the `## f)` section the docs-gates wrapper REQUIRES of the
   newest report (documented in the wrapper I had already run), and its
   skill table contained a raw `| tee $out` code span that split a table
   row (the exact escaped-pipe class strike-safety exists for, documented
   in AGENTS.md). The concurrent session patched the §f before I did.
4. **Duplicate concurrent fix**: my setup-go v5.5.0 SHA pin for the
   morning's cue-vet finding was superseded within the hour by the
   concurrent session's v7.1.0 pin (tag-verified, both sites). Same
   defect, two sessions, two fixes — I did not check whether the file's
   author (commit `2a46a9c6`, authored TODAY) was still active before
   fixing their finding.
5. **Small tool-discipline stumbles**: ran a scoped `buildflow -s` and
   trusted a red verdict before applying the documented no-cache bypass
   (the cache-trap AGENTS.md warns about — the no-cache rerun exonerated);
   passed `-s` twice to buildflow (cli.validation rejection); ran
   fleet-tags.sh without its required `--repos`; used a derivation NAME
   where a flake check ATTRIBUTE was needed (one wasted build cycle).
6. **Session end mid-verification**: the last full test sweep and the CI
   verdict read were still outstanding when the session ended; both were
   superseded by the overnight trains. My last verified state stands at
   21:40 (docs-gates full PASS); the push/CI claims for my commits were
   never closed out — the exact "push statements rot fastest" failure the
   repo's own conventions rank first.
7. **Never stated the served-markup exemption at train end** (required by
   the release runbook when a train skips the stack E2E) — stated in §a8
   now, eighteen hours late.

## e) WHAT WE SHOULD IMPROVE

1. **Read concurrent-session signals FIRST** (git log timestamps, newest
   docs/status, TODO_LIST harvest sections) and scope a battery run to
   verification + gaps BEFORE dispatching anything.
2. **A session-end checklist**: full test sweep verdict read, CI verdict
   via `scripts/ci-verdict.sh`, status index regen, docs-gates --check,
   explicit served-markup exemption statement, then commit — in that
   order, before the machine sleeps.
3. **Never put raw pipes inside markdown table cells** — reword around
   them (strike-safety's fixture-proven class).
4. **Read the docs-gates wrapper contract before writing status reports**:
   the newest report needs `## f)` with rows whose tokens match their
   TODO_LIST homes.
5. **Subagent concurrency cap 3, sequential waves** (the fleet runbook's
   own rule); absorb failed subagents' tasks directly instead of
   relaunching into a rate-limited provider.
6. **Scoped buildflow runs get BUILDFLOW_NO_RESULT_CACHE=1 by default** —
   the cache trap has now cost this repo three separate stumbles.
7. **Route findings at write time** (the fleet rule held: zero findings
   were lost to report-entombment this session — keep it that way).

## f) Up to 50 things we should get done next

Ranked by impact; tokens ride their TODO_LIST homes (the freshness gate
matches these).

1. **§D10 file-size disposition** (owner) — split-on-sight vs raise-cap
   (`max_file_size`, buildflow #91) vs warn-only; BLOCKS the release gate.
2. **v2.15.0 fold + release** (owner) — CHANGELOG dated section, version
   bump, `scripts/release.sh`, gates, tag, push.
3. **Stack lock bump, T01 cycle-1** — `nix flake lock --update-input
   webphone`, commit, stack gates (runbook steps 6–7).
4. **Stack browser E2E** for the release train's served-markup verdict
   (445s budget; this train was exempt — the next will not be).
5. **aarch64 cross-build** at the release (ELF-byte verified).
6. **Lychee link check** after the next archive move (the v2.13.0 class).
7. **Bump Train 1** — go-cqrs-lite engine wave v4.19.0/v4.26.0/v4.9.1
   (the RawQuery crash-loop fix) + the prod-data compat gate.
8. **Bump Train 2** — templ v0.3.1118 with CLI-paired `templ generate`.
9. **Bump Train 3** — modernc sqlite v1.62.0 (+ backup-drill check).
10. **Bump Train 4** — small SDKs ×5 incl. `idempotency.WithClock` (retire
    the TimingScale band-aid; do NOT file the 2026-11-08 draft).
11. **Deepening pick** (owner) — five candidates; top: messaging read
    surface onto the ThreadGate (design already sketched).
12. **Deepening execution** of the picked candidate (grill → interface →
    tests survive).
13. **Naming adjudication train** — the 4 renames + the ruling row.
14. **FCM token-mint single-flight** (fcr small-findings batch).
15. **Blob brand-before-write** edge (same row).
16. **Seq-as-retention-clock comment** (same row).
17. **T18 FCM invalidateToken on 401/403** (push transports).
18. **erraudit JS `.catch` directive-form triage** (10 sites, one class).
19. **vm-tests.nix split** (one file per KVM test).
20. **favicon green split brain** (derive from one token source).
21. **asr.Locality constants + ASRProvider enum** (type tails row).
22. **LoginPageExtra positional bools → question-form params**.
23. **push.Transport through PushDevices.Upsert + SavePushDevice** (the
    parity rule's owed second change).
24. **Disposition echo guard + `SetDisposition` changed-bool symmetry**.
25. **Typed feed maps widened to all three owner lanes** (identities +
    matrix.owner + signal.owner).
26. **F-A duplicated/misplaced doc comments** (3 sites, one class).
27. **F-D FCM mutex-across-network** (single-flight shape).
28. **F-F Fan panic-containment pin**.
29. **F1 services-never-import-server test stale subset**.
30. **F3 duplicated overflow-trim logic** (dial.go).
31. **F4/F5 transcripts Seq-clock + whole-corpus read posture**.
32. **F7 boot-helper duplication across cqrs** (4 prod + 5 test copies).
33. **F13/F14 island file splits** (calls.js 1.7k lines; shell.js 3.3k,
    section renumbering).
34. **Catalog verification ring** — live v2-contract run against the REAL
    hub + CI step checking the `-plain` tree.
35. **DLQ operator visibility** (surface the dead-letter store).
36. **Metrics off the frozen webphone.db ladder** (ReadCounts/FlagPresent).
37. **Snapshot assessment for long-lived streams** (measured trigger).
38. **readmodels.db orphan operator note + test-only store/messages.go
    deletion**.
39. **Depth-pin runway design note** (five packages near budget).
40. **Bring .templ under the depth ratchet** (registry act).
41. **Server-root test drain rides carves only** (standing policy).
42. **Cadence pre-drain automation** (pre-compute the gates-passing set).
43. **Flip `workflow-e2e` CI job to blocking** (after N clean trains).
44. **Dial-gate chip + voicemail browser journeys** (need the fake-SIP/
    `--phone-api` harness lanes).
45. **Hardening-parity pin** (backup oneshot vs main unit) + the
    dedup-registry row.
46. **File the drafted BuildFlow bandit-provider issue** (owner banner).
47. **Ownership hot-reload IMPLEMENT arm** (owner trigger call pending).
48. **SSO-only identity resolution in the stack fixture**.
49. **webphone-android adopts the `lines[]` wire** (mirror row there).
50. **pma daemon push heuristic** (cross-repo pma train; age heuristic).

## g) Questions only the owner can answer

1. **The §D10 file-size ruling** — split-on-sight, raise the cap via
   buildflow #91's now-working `max_file_size`, or warn-only? It is the
   single gate standing between main and the v2.15.0 release, and every
   option is a standing policy, not a code change I can derive.
2. **The deepening pick** — which of the 08:10 brief's five candidates do
   you want deepened (top recommendation: collapse the messaging read
   surface onto the ThreadGate)? The skill is deliberately stopped at the
   pick; everything downstream is gated on it.
3. **Bump trains vs the release: which first?** Train 1 carries a real
   fix (the RawQuery crash-loop) and the release is §D10-blocked — do I
   (or a successor session) execute Trains 1–4 now while the release
   waits, or hold all dependency motion until v2.15.0 has cut, so the
   release folds on the current pins?
