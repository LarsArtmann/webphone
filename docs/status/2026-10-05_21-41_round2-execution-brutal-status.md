# Round-2 execution — brutal status report (2026-10-05 21:41)

**Scope:** THIS session only (≈20:45–21:41 CEST): round-2 Full Execution of the
20:30 Pareto plan (`90ca9d1`) — Q1 + Q2 assistant legs, one unplanned gate
repair, and the hygiene closes. Prior sessions' work is context, not subject.
**End state:** remote main = `3b08058`, CI green, tree clean, working
directory at `docs/status/` parity with origin.

**Inputs:** round-2 plan Q1–Q17 + F-ids · briefing (32 rows + 16:50 addendum)
· CI verdicts on `90ca9d1`/`6a8eaac`/`3b08058` · vulnix + tw.css + git
evidence gathered this session.

---

## Opening self-critique — what did I forget / do worse / can still improve

1. **The CI check that caught the third break was luck-adjacent.** The
   handoff said `90ca9d1` CI was "in_progress, docs-only, expected green."
   I could have skimmed the plan commit's diff FIRST (it carried 261 lines
   of `error-contract.md` reflow — visible in `git show --stat` in one
   second) and predicted the break. Instead I trusted the expectation and
   only caught it because verifying CI was todo #1. Lesson: **"expected
   green" is a hypothesis, not a verdict — diff-reading the pushed commit
   beats trusting its description.**
2. **Negative control ran in the daemon-watched live tree.** I proved real
   drift still fails by `sed -i`-corrupting `docs/error-contract.md`, then
   restoring. The daemon commits within SECONDS — had it swept the
   corrupted intermediate, main would have gone red on garbage. Nothing
   broke, but the correct protocol was a temp worktree. Realized risk: zero;
   incurred risk: real.
3. **First briefing edit lost to the daemon race** ("modified since last
   read") — recovered by atomic `cat >>` heredoc. The round-1 session had
   already learned this; I reached for `edit` first anyway. Protocol for
   daemon-shared docs: read-immediately-then-append, or single atomic write.
4. **No full local gate this session.** I verified via full `go test ./...`,
   scoped `nix fmt` on the two css files, the build script's own artifact
   gates, and CI. I did NOT run `buildflow` or a full `nix flake check`
   locally. CI's job (observed) is `go test ./...` — the flake-only gates
   (treefmt whole-repo, island-lint, module checks, kvm backup test) and
   buildflow's detectors (gitleaks/codespell/markdown-lint) did NOT run over
   this session's changes. Risk assessed low (docs + test file + css with
   scoped fmt clean), but it is exactly ruling 34's question in the flesh.
5. **Minor waste:** markdownlint count grep pattern wrong on first try
   (returned 0; redone correctly); first GraphQL subscribe mutation used
   wrong payload fields (introspect first); histogram needed a second full
   markdownlint run. Each self-caught in seconds; none left wrong state.

---

## a) FULLY DONE (verified this session)

| #  | Item                                                                                                                                                                                                                                                                                                                                                                 | Proof                                                                                   |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| 1  | **Gate repair: registry pin made content-based.** Third daemon table-reflow (swept into `90ca9d1`) had re-broken the byte-exact pin; CI red. `TestErrorCodeRegistryIsFresh` now compares cell content with padding collapsed (`canonicalRow`); `-update` writer unchanged; `docs/error-contract.md` preamble documents the content-based pin                         | Commit `6a8eaac`; CI **SUCCESS** (3m39s); full `go test -count=1 ./...` green           |
| 2  | **Regression micro-test** `TestRegistryRowComparisonIgnoresPadding`: reflowed/header/separator rows canonicalize; real family change still reads as drift                                                                                                                                                                                                            | In `6a8eaac`, passing in CI                                                             |
| 3  | **Negative control:** corrupted `blob.escape` family in the live doc → test RED naming the code → restored byte-exact → green                                                                                                                                                                                                                                        | Ran live 21:0x, output captured                                                         |
| 4  | **Q1/F1.1 — briefing rows 33–34**: daemon-docs-format policy (reframed: pin now immune; recommendation keep-reflow) + proof-bar for green (recommendation: CI verdict on pushed head; release.sh keeps local full gate)                                                                                                                                              | Briefing tail, commit `3b08058`                                                         |
| 5  | **Q1/F1.5 — sweep-train attribution: FINISHED.** `81ea689` (15:42) = last commit touching go.mod/go.sum/vendor; nothing since                                                                                                                                                                                                                                        | `git log` evidence; answers status-report question 1                                    |
| 6  | **Q1/F1.2 — CHANGELOG sweep entry**: all six bumps old→new (templ 0.3.1020→0.3.1070, templ-components 1.19.4→1.20.0 + submodules, usermgmt 4.13.1→4.14.0, go-health 0.4.1→0.5.0, webauthn 0.18.1→0.18.2 indirect, otel 1.46.0→1.47.0 indirect; dashboard held 0.10.2) + tw.css consequence + registry-pin fix entry                                                  | `[Unreleased] ### Changed`, daemon commit `3c0c214` carried it verbatim (verified stat) |
| 7  | **Q1/F1.3 — tw.css class-identity check: FIRED REAL.** 1.20.0 moved button outline-warning/success text from amber/green-600 to -700, escaping the @theme remap → raw palette colors. Remap extended with `-700` entries (red-family precedent); artifact rebuilt via the sanctioned script; utilities resolve through `--warn`/`--ok` again; scoped `nix fmt` clean | `tw.css` diff shows `var(--warn)`/`var(--ok)` on -700; script gates passed              |
| 8  | **Q1/F1.4 — vulnix over the NEW runtime closure: CLEAN.** 8 derivations; only glibc 2.44-25 range-match noise; both CVEs (2026-5435, 2026-6238) distro-patched in locked nixpkgs; verdict "zero real advisories", exit 0                                                                                                                                             | Full scan output in session log                                                         |
| 9  | **Q2/F2.1 — AGENTS CI-check habit** (one line swapped, net-zero, 129/377 lines): verify end states with `ls-remote` **+ CI verdict**, check CI after EVERY push, receipts cited                                                                                                                                                                                      | `3b08058`                                                                               |
| 10 | **Q2/F2.2 — markdownlint invocation SETTLED**: `nix run nixpkgs#markdownlint-cli2 -- <docs>` (v0.23.3, library defaults); baseline over the three 10-05 docs = 172 findings (170 MD013 line-length + 2 MD026 heading-punct), all cosmetic; detect-only until row 18; recorded in the TODO row                                                                        | TODO_LIST tooling-hygiene row updated                                                   |
| 11 | **Q2/F2.3 — crush #3846 subscribe ATTEMPTED, honestly blocked**: `updateSubscription` needs the `notifications` scope; token has none; only one gh account. Owner one-liner recorded in the standing-watches row                                                                                                                                                     | TODO_LIST watches row; error captured verbatim                                          |
| 12 | **Q13 gate verified still shut**: latest crush release v0.97.1 (2026-09-29) predates the bug diagnosis; issue open, 0 comments — no fix release to re-test against                                                                                                                                                                                                   | `gh api` both checks                                                                    |
| 13 | **Post-push CI habit exercised** on both pushes (`6a8eaac`, `3b08058` — SUCCESS at 3m39s / 5m35s); remote = local verified via `ls-remote`                                                                                                                                                                                                                           | `gh run list` receipts                                                                  |

## b) PARTIALLY DONE

| Item                                             | State                                           | Remaining                                                                               |
| ------------------------------------------------ | ----------------------------------------------- | --------------------------------------------------------------------------------------- |
| F2.3 #3846 subscription                          | Invocation proven, scope blocked                | Owner: `gh auth refresh -s notifications` + mutation, or click Subscribe                |
| markdownlint lane                                | Invocation + baseline done                      | POSTURE (configure-to-house-style vs detect-only) = briefing row 18, owner sitting      |
| Local full-gate proof for this session's changes | go-test + scoped fmt + script gates + CI done   | `buildflow` + full `nix flake check` NOT run locally (see §e.1) — subsumed by ruling 34 |
| Sitting prep (Q1)                                | Rows 33–34 + sweep-audit closure block appended | THE SITTING itself (Q3) — owner, 34 rows                                                |

## c) NOT STARTED (all owner-terminal or gated — correctly untouched)

- **Q3 THE OWNER SITTING** (34 rows; the 1%→51% lever; briefing fully prepped)
- **Q4 DEPLOY TRAIN** (stack CI 1h-ceiling/mod_enum diagnosis → lock bump → gates + E2E → aarch64 ELF → pbx relock → rebuild/drill/switch → smoke) — owner
- **Q5 prod SMS-bridge 422 triage** — owner (pack ready)
- **Q6 post-sitting paper closes** — gated on Q3 outcomes
- **Q7 release announcements** — owner
- **Q8 erraudit re-measure** — date-gated 2026-11-05 (measured early 10-05: 0/0/0)
- **Q9 stack batch 2** (`services.webphone.paperless`, `ftypqt` sniff, E2E MMS-outbound, pbx FEATURES:87) + **Q10 stack docs batch** — stack lane, deploy-gated
- **Q11 mic pre-warm live ritual**, **Q12 visual-harness disposition** — owner, post-deploy
- **Q14 full local re-proof** — conditional on ruling 34
- **Q15 internal/server carve** — trigger-gated (no new server file landed this session)
- **Q16 quarterly watches** — date-gated 2026-12-20
- **Q17 scheduled sends** — stays dead unless the sitting revives it

## d) TOTALLY FUCKED UP!

1. **Main was CI-red a THIRD time today on arrival** (`90ca9d1`, ~18:33–18:58):
   my own 20:33 plan commit — authored by the prior session-leg of me — swept
   in the daemon's reflow of `error-contract.md` (261 lines) without
   diff-reading it, re-breaking the just-repaired byte-exact pin. The
   "docs-only, expected green" label was wrong by 261 lines. Fixed this
   session at the root (content-based pin); the miss itself belongs here.
2. **Self-inflicted near-miss:** the sed corruption negative control ran in
   the daemon-watched live tree (§Opening.2). One daemon tick away from a
   red-garbage commit. Protocol violation with a lucky outcome.
3. Nothing else was broken, reverted, or left red. No owner files touched;
   guardrails (no owner legs, no speculative code) held — the only
   code-adjacent changes are the test-side pin and the remap completion,
   both verified.

## e) WHAT WE SHOULD IMPROVE

1. **CI's coverage vs the local gate stack.** Observed this session: the CI
   job is `go test ./...`; treefmt/island-lint/module/kvm checks and
   buildflow's detectors are LOCAL-only. Every daemon auto-commit
   (3 today) gets CI-verified for Go but never format/lint-checked before
   push. Either extend CI (a flake-check job or the `ci` devshell +
   treefmt) or ratify CI-go-tests-only as the bar (ruling 34's decision).
2. **Diff-read every commit before/after push** — "expected green" labels
   from handoffs are hypotheses. The 261-line reflow was visible in one
   `git show --stat`.
3. **Negative controls in worktrees, never in the daemon-watched tree.**
4. **Atomic writes for daemon-shared docs** (`cat >>`/single write), not
   edit-tool round trips (one lost edit this session).
5. **The `-update` writer still emits unpadded rows** — the daemon will
   re-pad after every regeneration (churn, now harmless). Optional polish:
   emit the aligned form the daemon produces (reverse-engineer its column
   widths) and the file goes quiet.
6. **GraphQL hygiene:** introspect payload types before mutating (first
   subscribe attempt failed on a nonexistent field).
7. **The `passkey_api.go:211` unused-`r` gopls Info is STILL live** (third
   session carrying it) — the concurrent passkey session never took it;
   keep it visible until that session or the owner kills it.

## f) Up to 50 things to get done next (owners marked; gates named)

**Owner — the levers (do these first):**

1. Q3: THE SITTING — 34 rows in dependency order (~90m; flips 9/16 TODO rows)
2. Q4 deploy train, step 1: diagnose the stack CI 1h-ceiling failure inside `nix flake check` (mod_enum build log)
3. Q4: stack lock bump to ≥`3b08058` once diagnosis lands
4. Q4: stack gates + browser E2E (445s budget; one re-run on the transfer flake)
5. Q4: aarch64 cross-build + ELF byte verification
6. Q4: pbx-artmann clean-tree check → relock → re-pin
7. Q4: `nixos-rebuild test` → passkey fail-closed password-file drill → `switch`
8. Q4: rotate `/tmp/pbx-toplevel-current` + fresh diff-closures baseline
9. Q4: smoke `--base https://pbx.artmann.tech --expect-version <V>`
10. Q5: prod SMS 422 triage (journal `telnyx-webhooks` → §4 decision tree → test SMS → record root cause)
11. Q7: pick announcement channels; approve v2.8.0 draft A/B/C; post
12. Q11: mic pre-warm live ritual (accept→speak sub-second; reject/missed warm release)
13. Q12: visual-harness disposition (eyeball 14 shots; persistence ruling)
14. Sitting rulings 33/34 (daemon-format policy; proof-bar) — now with this session's evidence
15. Row 18: markdownlint posture (configure vs detect-only) — invocation + baseline ready
16. Grant gh `notifications` scope (or web-Subscribe) → finish the #3846 watch (one-liner in TODO)
17. Kill or own the `passkey_api.go:211` unused-`r` finding (third session carrying it)

**Assistant — executable after/between owner legs:**
18. Q6: record all sitting verdicts into briefing + TODO (same sitting, no drift)
19. Q6: registry standing rows for ratified 29–32 + vcard triad (if ratified)
20. Q6: ruled markdownlint posture implementation (config or AGENTS note)
21. Q6: AGENTS edits within the 377 cap for every ruled posture
22. Q6: TODO/ROADMAP verdict updates for every outcome
23. Q8: erraudit tier-1+2 re-measure (due 2026-11-05; both must stay 0)
24. Q8: boot-surface re-grade vs error-contract §"Boot surface"; bump next-due in AGENTS
25. Q13: re-test `mcp_qmd_get` after the next Crush release (watch recorded)
26. Q14 (conditional on ruling 34): clean-cache buildflow + full `nix flake check`, OR record CI-as-bar and close
27. Optional polish: make the registry `-update` writer emit daemon-aligned rows (kills regen churn)
28. Fold this report's §e.1 CI-coverage question into the sitting if ruling 34 stays open
29. Sweep-log line in dedup-registry for this session's registry-pin change? (No — not a clone ruling; skip unless a detector resurfaces it. Kept here as the reminder NOT to.)

**Stack lane (deploy-gated):**
30. Q9: `services.webphone.paperless` module option + smoke arm
31. Q9: `ftypqt`→`video/quicktime` sniff fix
32. Q9: E2E MMS-outbound coverage
33. Q9: pbx-artmann FEATURES:87 stale sniff text
34. Q10: WebTransport verdict doc
35. Q10: telephony deploy.md secret PATH column
36. Q10: ops-runbook demo-call recipe + secrets path
37. Q10: MOH + `/recordings/` + CDR checks

**Gated/standing:**
38. Q15: server-carve micro-plan WHEN the next file lands in internal/server
39. Q16: quarterly watches 2026-12-20 (sip.js 0.22, templ-components, oxlint globals, E2E budget)
40. Q17: keep scheduled sends dead unless the sitting revives it
41. Re-check stack CI recovery state before the deploy train (read-only `gh run list` on the stack)
42. Verify the v2.9.0 fold question (sitting part 1) against the freshly-green webphone CI story

_(42 real items; padding to 50 would be inventory theater.)_

## g) Questions I can NOT figure out myself (max 3)

1. **Sitting now?** Q3 is fully prepped (34 rows, sweep-audit closure aboard,
   rows 33–34 carry this session's evidence). Do you want to run the sitting
   next — and if so, in one block or split (release/deploy rows first, then
   the rest)?
2. **CI bar policy (the ruling-34 core, sharpened by today):** today caught
   three reds ONLY because CI was checked — yet CI runs go-tests only, and
   the flake/format gates stayed local (and unrun this session). Do you want
   (a) CI = the ratified bar as-is, (b) CI extended (flake-check job) so
   daemon pushes get the full gate, or (c) local full gates before
   sittings/releases only?
3. **gh token scope:** may I have you run `gh auth refresh -s notifications`
   (then I finish the #3846 subscribe), or will you click Subscribe on the
   issue yourself? I cannot grant scopes, and the watch is otherwise blind
   until the quarterly re-check.

---

_Snapshot report; goes stale. HARVEST note: §f items 18–28 are routing
candidates — TODO_LIST already carries the rows this session updated
(markdownlint invocation, #3846 watch); items 2–13 are the round-2 plan's
owner legs and live there already._
