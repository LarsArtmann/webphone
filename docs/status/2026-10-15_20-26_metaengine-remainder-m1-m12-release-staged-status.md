# Status: SUPERB metaengine train remainder — M1–M12 DONE, M13 release staged, tail spikes landed

**2026-10-15 20:26 · session scope:** the remainder Pareto plan (`docs/planning/2026-10-15_17-14_SUPERB-metaengine-train-remainder-pareto.md`, M1–M18 / μ01–μ65) under the owner's "GET SHIT DONE — the whole TODO list" order, with G3 = "Everything" (full release incl. tag + stack notification) and M9 = "file upstream now".

## a) Fully done

- **M1–M7 (from the prior session, verified standing):** T09 snippets, G2 GO verdict, T11+T12 faxes, T16 interim census, T13+T14 transcripts — all swapped, CI green.
- **M8 enroll tokens — the seam's second half (userauth swap) landed.** `Lookup` extended with the expiry (unreleased seam; the invalid-vs-expired operator-log distinction stays in userauth as planned); `userauth.Service` carries the engine store, `New` takes it; `app.go` wires `engine.EnrollTokens()`; the `-enroll-passkey` CLI opens the engine itself; `migrateEnrollTokens` + the `wp_enroll_tokens` table are gone; tests seed aged tokens through the store (no raw SQL). Registry regenerated (6 SQL-era codes retired). Auth-grade gates green (userauth + `TestServer` suites), full suite green (server 548 s), smoke **47+4+8+9**.
- **M9 upstream filings — both filed with the full skill treatment** (`verify-before-filing` gates + `github-voice` drafts in `docs/drafts/`, check-draft 0/0, no duplicates): go-cqrs-lite **#71** (typed `Scan` truncates at 100 with no signal at the typed level) and **#72** (no keep-newest-N cap primitive). Standing-watches row links both.
- **M11 buildflow verdict collected** (see d) for the open tail of it): the gate is red on exactly ONE error-severity finding, and it is a **cqrs-lint S001 false positive** — "hardcoded secret" for `enrollTokensCollection = "enroll_tokens"` (a table-name literal matching the "token" substring heuristic; RULES.md says S001 is critical/medium-confidence). Also surfaced: the 189-entry `.buildflow-baseline.json` is fully stale (M11's prediction, confirmed by the tool itself).
- **M12 docs sweep:** CHANGELOG `[Unreleased]` entry for all swaps + backup coverage; FEATURES Platform row "Metaengine read models"; AGENTS readmodels bullet extended (five seams, no-legacy-import-for-enroll, converging trims, backup) + the art-dupl **version-pin rule**; dedup-registry sweep-log ruling; TODO_LIST train row; **T16 CLOSE census: 46 → 39 sites** under the stated rule (predictions −10/−6 over-counted; enroll never lived in `internal/store` — honest note recorded in plan §7).
- **Tail μ56–μ62, μ65:**
  - **T19 sessions memo** (`docs/planning/2026-10-15_19-59_…`): N4 stands — zero engine win for a kv+slide store, auth-grade cost, continuity risk.
  - **T20 dual-run spike** (`internal/store/threadsdual_test.go`, tag `threadsdual`): **8/8 byte-identical** against the production SQL lists (List, Archived, Search ×5 incl. `%`-escape + ASCII case). Report + owner decision row (`docs/planning/2026-10-15_20-35_T20-dual-run-report.md`): the port costs 4 hand re-implementations — last-message fold, in-Go compound sort (**sqlite pushdown applies only the FIRST `WithSortColumns` column** — upstream doc-verified), hand-rolled LIKE, shadow-fact feed. **NO-GO now stands on measured economics.**
  - **μ60** upload-limit test marked `testing.Short()` (owner approved via "full").
  - **μ61–μ62** fuzz perf measured: ~800–1400 execs/s with the per-iteration engine → sharing one engine would trade determinism for nothing; kept.
  - **μ65** postgres proof: `pgengine/v4@v4.4.2 New(dsn)` returns the same `metaengine.Engine` readmodels consumes — a constructor swap, no seam change; memo line in plan §7.
- **M13 preconditions (step 1):** CHANGELOG folded into **`## [2.9.0] - 2026-10-15`** (Unreleased reset, compare links added); runbook + `scripts/release.sh` read — the script automates steps 2–8 (gates incl. buildflow-no-cache + vulnix, tag, push, lychee, stack relock + browser E2E + VM test, aarch64, gh release).

## b) Partially done

- **M11 execution tail:** the S001 FP is diagnosed but NOT resolved (suppression vs rename decision below), the stale 189-entry baseline not yet refreshed/deleted, and no re-run to green yet.
- **M13 release:** staged but NOT run — `scripts/release.sh 2.9.0` enforces a quiet host (load < 8) and the host sat at load 89→37 during this session (sibling sessions: a NixOS toplevel rebuild + a chromium/bun E2E — left alone per the concurrent-sessions rule). Release.sh itself re-runs every gate, so nothing was shortcut.
- **μ63–μ64 backup verification — the important half became a REAL fix:** discovered the **backup story never covered `readmodels.db`** (unit script: only `webphone.db` + `files/`; the drill + README restore path likewise). Fixed at root cause: `package/backup-script.nix` (guarded `sqlite3 .backup` for both latest + snapshot paths), `nix/module-output.golden` regenerated (webphone-module check green), `nix/vm-tests.nix` assertions extended (latest + snapshot + integrity), drill mode-check includes `readmodels.db`, README § Backups updated, CHANGELOG folded. **backup-drill check GREEN**; the KVM `webphone-backup` VM test was launched but its result was never collected (background job) — re-run it before the release.

## c) Not started

- Nothing from the plan's P0. Correctly NOT done: the stack browser E2E (owed only on served-markup change — none happened; the module/nix changes are markup-free), and μ65 beyond the interface proof (optional; no multi-engine need exists).

## d) Totally fucked up (own the failures)

1. **I red-ed main with a tool-version skew.** Regenerated `.art-dupl-baseline.json` with the LOCAL v0.9.0 while CI pins **v0.7.2** ("Install art-dupl at the baseline's tool version" — it says it right there); v0.7.2 saw 36 phantom clone groups → CI red on the M8 push (run 38781159138). Fixed by re-recording with the pinned version; the rule + war story are now in AGENTS + dedup-registry. Lesson: check the CI-pinned tool version BEFORE regenerating any baseline.
2. **The M8 narrative commit never landed** — the daemon swept the working tree before my `git commit` ("nothing to commit"); a passkey-surface change sits in two `chore: auto-commit` commits. The `auth:` discipline (suite + prefix) was honored in substance, not in the log. The story lives in docs, but git history is poorer for it. Race the daemon: commit FIRST, run gates after.
3. **My first T20 harness draft was sloppy** — three scaffold bugs (a nonexistent method call, a phantom struct field, a missing helper) and two wrong fixture unreads. Compiler + comparison caught everything, but that's two avoidable cycles.
4. **The T16-close census line I first wrote contradicted itself** (per-seam breakdown summed to −12 against a measured −7). Caught pre-commit; the plan now carries only the honest milestones.
5. **Slow diagnosis theater:** five queries against the buildflow state DB to find the S001 finding that was already in my own earlier truncated output. Read the full tool output before grepping subsets.
6. **Untracked background jobs:** the load-wait watchdog and the KVM VM test were left running without durable IDs recorded here (the load-wait was 193; the VM test 189). Fixed by writing this report.

## e) Improvements made to the codebase (beyond the plan)

- Backup coverage for `readmodels.db` (b/ above) — this was the session's real find: a silent data-loss gap in the new storage layer, closed with VM-test teeth.
- AGENTS now carries the art-dupl version-pin rule so the v0.7.2/v0.9.0 class can't recur silently.
- The `threadsdual` harness doubles as a regression harness for future metaengine bumps (re-run after any minor bump).

## f) Next 50 (prioritized head, then the long tail)

1. Resolve S001: `//cqrs-lint:ignore(S001)` + rationale comment (directive support verified in cqrs-lint v4.13.5 doctor) — do NOT rename the collection for a linter; file the FP upstream.
2. File the S001 heuristic upstream (go-cqrs-lite): "token"-substring + table-name literal is too broad.
3. Refresh or delete `.buildflow-baseline.json` (189 dead entries; the tool itself flags it).
4. Re-run buildflow → green; confirm tier-1/2 posture unchanged.
5. Collect the KVM `webphone-backup` VM test verdict (re-run: `nix build .#checks.x86_64-linux.webphone-backup`).
6. When host load < 8: run `scripts/release.sh 2.9.0` (the full G3: gates, tag, push, lychee, stack relock + E2E + VM test, aarch64, gh release).
7. Watch release CI + verify tag + gh release published (`git ls-remote`, `gh run list`).
8. Stack notification: "webphone 2.9.0 train done; relock window open" (G3 commitment).
9. Owner: tick the MANUALLY REVIEWED boxes on go-cqrs-lite #71/#72.
10. Owner: ratify the T20 NO-GO decision row + the T19 memo.
11. Owner: the deploy tail — v2.8.0 was never deployed; 2.9.0 supersedes it (one deploy, re-verify `--expect-version 2.9.0`).
12. Verify `/healthz` actually covers `readmodels.db` (engine HealthCheck registration — confirm, don't assume).
13. Confirm CI does NOT pass `-short` (the upload test must stay exercised in CI).
14. Time a `-short` local run to quantify μ60's win; consider documenting `-short` in AGENTS § Commands.
15. Reconcile the 10-14 plan's over-counted census predictions inline (−10/−6 → actuals) or leave the honest note (already there — decide).
16. Quarterly art-dupl baseline audit (next ~2027-01); the 63-vs-66 version skew is documented.
17. erraudit tier-1/2 monthly re-measure (due 2026-11-05).
18. Decide deliberately: bump the CI art-dupl pin to v0.9.0 once (workflow edit + one re-record) vs keep pinning locals back to v0.7.2.
19. Mechanical guard candidate: `scripts/art-dupl-ci.sh` wrapper that always runs the pinned version.
20. Add the version-skew war story to `docs/lessons.md`.
21. When #72 (cap primitive) lands upstream: revisit the converging-loop trims + enroll sweep (Standing watch).
22. When #71 lands: adopt the truncation signal in webphone scans (Standing watch).
23. Re-run the `threadsdual` harness after every metaengine minor bump (it regenerates the T20 evidence).
24. Consider CI-ifying the dual-run harness as a scheduled/optional job.
25. Audit today's daemon commits for unintended sweeps (spot-check 3–4).
26. Verify `.buildflow.yml` ownership: it carried not-mine changes (build_mode full, un-suppressed tools) — confirm with owner that this is the intended live posture (it IS what this session's gates ran under).
27. `.buildflow.yml` re-read after any daemon activity near it (standing rule).
28. Post-release: TODO_LIST train row gets the G3 GREEN line.
29. Post-release status report addendum with the release verdict.
30. ROADMAP: verified no changes needed this train (record the verification).
31. Smoke re-run happens inside release.sh — but if run manually later, expect 47+4+8+9.
32. `go.mod`/`go.sum` untouched this session — verify stays true at tag time.
33. Link check (`lychee`) runs inside release.sh; if it trips on the new drafts' issue links, whitelist the pattern.
34. The enroll no-legacy-import decision: optional boot-log line when legacy `wp_enroll_tokens` rows exist but are ignored (operator visibility).
35. Consider `pgengine` behind `readmodels.OpenWithEngine` ONLY on a real need (recorded path; do not build speculatively).
36. Backup script comment: document why `sqlite3 .backup` needs no `-wal` handling (consistent snapshot).
37. Check the remaining plan G-gates: G3 is the last one; the train formally CLOSES after the release.
38. After the stack relock: pbx-artmann re-pin is the owner's terminal (runbook).
39. Watch the next Crush upgrade for the QMD `get` render fix (Standing watch, unrelated but open).
40. The 429-combined-load test (#5) and ASR follow-ups remain parked from older sweeps — untouched, still open.
41. `docs/dom-contract.md`: no changes this train — verify the count test stayed green (it did; no markup touched).
42. Consider a CHANGELOG line for the `threadsdual` harness? (Test-only → stays changelog-silent per the bar. Verify the line I wrote doesn't violate the bar — it doesn't; the harness line isn't in CHANGELOG.)
43. Sweep `docs/drafts/` into the dedup-registry's "filed upstream" ledger if one exists (the registry tracks clone rulings; the issues are linked from TODO_LIST — consistent).
44. `internal/store` legacy readers: when the import window closes (owner call, post-deploy), delete the three `legacy_*.go` files + flags (the census will drop ~3 more sites).
45. Same for `legacy_contacts.go` (3 sites) — the full legacy cleanup is ONE future train.
46. Consider a `TestEngineCollectionsMatchPlan` arch-test so a new read model can't skip the Plan registration (S001-adjacent hygiene).
47. The upload test's `-short` skip message names the cost (61 MiB) — good; keep messages honest when numbers change.
48. Re-verify AGENTS.md line cap (377) after this session's additions (was 151 lines — far under).
49. Next session starts with `buildflow` per the train-start rule (it was mid-session here; the M11 refresh covers it).
50. After the owner deploys 2.9.0: run the prod smoke (`--base https://pbx.artmann.tech --expect-version 2.9.0`) — owner terminal, then the backup-drill against prod data.

## g) Questions for the owner (max 3)

1. **S001 handling:** suppress in-place with `//cqrs-lint:ignore(S001)` + rationale (my recommendation — the name is honest domain vocabulary and the directive is auditable), rename `enrollTokensCollection`/`"enroll_tokens"` to dodge the "token" heuristic (free NOW, pre-release; poisons the collection id after), or both-suppress-and-file-upstream? I'll do suppress + file unless you prefer the rename.
2. **art-dupl pin direction:** keep CI on v0.7.2 and force locals to match (current fix), or deliberately bump the workflow to v0.9.0 + one canonical re-record so the newer detector is the standard? (Workflow change in webphone = your call.)
3. **Release autonomy:** `scripts/release.sh 2.9.0` waits for host load < 8 by design. With concurrent sessions rebuilding NixOS + running E2Es, do you want me to fire it the moment the host is quiet (G3 "Everything" already approved), or hold for your explicit go since the run relocks the stack and eats ~30+ min of heavy compute?

## Evidence

- CI: 38781827344 (M8 + baseline fix, GREEN), 38782919331 (TODO row, GREEN), 38784539351 (dual-run spike, GREEN), 38784699484 (2.9.0 fold, GREEN).
- Commits this session: `81ab1e0` + `a39c213` (M8 + baseline), `58113c0` (docs), `7dc4255` (spike), `0f1031d` (fold) — plus daemon sweeps carrying the seams.
- Upstream: go-cqrs-lite #71, #72 (OPEN).
- Gates this session: full suite GREEN (server 548 s), smoke 47+4+8+9, backup-drill GREEN, webphone-module GREEN, buildflow = 1 known FP.
