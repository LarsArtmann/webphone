# Status Report — Strong-ID Branded Train + branching-flow Suppression

**Date:** 2026-10-05 15:25 CEST
**Scope:** Single session. webphone strong-id report resolution (13 findings), plus the root-cause fix in branching-flow that the resolution exposed. No other research; everything below comes from this session's run.
**Repos touched:** `~/projects/webphone` (main), `~/projects/branching-flow` (main), `~/projects/go-branded-id` (read-only reference).
**Format note:** `.md` per explicit instruction in the dispatch prompt (status-report skill is HTML-canonical; standing dispatch exception applies).

**Verdict:** strong-id on webphone went **13 rows → 0** (2 legitimately suppressed, visible under `--include-suppressed`). webphone full suite 0 failures; buildflow **58 steps, 0 failed**. branching-flow strongid tests + touched-package golangci green. One cross-repo caveat until release: the PATH `branching-flow` binary (0.2.0 store build) still shows the 2 suppressed rows — only `go run ./cmd/branching-flow` from the repo shows 0.

---

## a) FULLY DONE

1. **webphone crm seam — `ContactRef` brand** (`internal/crm/client.go`, `resolver.go`): named brand + `NewContactRef` (wraps the CRM's opaque id verbatim — the CRM owns validity). `Match.ID`, `Client.LogCall`, `Resolver.LogCall` typed. Consequence: only a lookup can mint a ref; a phone number can no longer be passed where a contact ref belongs. `lookupResponse` stays raw (correct — it IS the wire).
2. **webphone paperless seam — branded metadata ids** (`internal/paperless/paperless.go`): `TagID`/`DocTypeID`/`FieldID` as `id.ID[Brand, int]`; `ensureMetadata` no longer returns three transposition-prone naked ints; SDK boundary converts via `.Get()`.
3. **webphone userauth seam — `usermgmt.UserID` end-to-end** (`internal/userauth/userauth.go`, `enroll.go`, `internal/server/passkey_api.go`, `cmd/webphone/main.go` untouched by luck of typing): `Register` returns the branded type; `EnrollToken.UserID`, `MintEnrollToken`, `BeginRegistration`, `FinishRegistration`, `FinishLogin` all typed. Browser-supplied ceremony keys are parsed **strictly** (`usermgmt.ParseUserID`) at the handler door (`parsePasskeyUserID`, 400 on garbage) — replacing the deprecated `usermgmt.NewUserID` that silently SHA-hashed garbage into pseudo-users (all 3 call sites removed; the gopls/staticcheck SA1019 warnings on `userauth.go` are gone). `wp_enroll_tokens` DB bytes unchanged (`.Get().String()` on write, strict parse on read; corrupt rows classify `Corruption` with the raw id in the message).
4. **Boundary-DTO rejects, documented at the site**: `apiContact.ID` and `sessionIdentity.DID` stay raw with reasoned `//nolint:branching-flow` comments. Reason: `id.ID` marshals the RAW value (no `Brand:` prefix, `id_json_v2.go`), so branding would silently change the island's wire bytes; DID is a config-validated identity with omitempty semantics and no id ever flows back in.
5. **branching-flow: strongid suppression implemented** (the rule that promised `--include-suppressed` but never emitted suppression records): `SuppressionDirective` field + `Suppressed()` on `strongid.Violation`; `core.ExtractIgnoreComments` + `Matching(line, IgnoreCommentTypeStrongID)` in the analyzer (same line / line above; bare, `:strong-id`, and explanation-qualified forms); `WithSuppression(finding.NolintSuppression(...))` in `to_findings.go`; suppression-aware runner registration in `container.go` (the `all` command); `dropSuppressed`/`activeItems` in the standalone `strong_id.go` (exit gate counts ACTIVE only, verified: `--exit-code` = 0 with suppressions aboard). 5 new tests pin typed/prev-line/explanation-qualified matching, the non-suppressing `:panic` directive, and finding-level `IsSuppressed`.
6. **Verification**: webphone `go test -count=1 ./...` → 0 failures (after regenerating the error-code registry for the new `userauth.enroll.userid` code, exactly as the failing test instructed); buildflow → 58 success / 0 failed (after one `nix fmt` fix to `sessionIdentity` alignment and one erraudit `context_loss` fix — both legitimate catches). branching-flow: `go test ./cmd/branching-flow/... ./pkg/strongid/... ./pkg/primitivetype/...` green; golangci 0 issues on `pkg/strongid/...` + `cmd/branching-flow/...`.
7. **Docs**: error-contract row "Passkey ceremony key malformed (400)"; AGENTS.md strong-id ruling (brands per seam + the raw-JSON-marshal gotcha); branching-flow CHANGELOG (Unreleased/Added) + TODO_LIST release item.
8. **Concurrent-session hygiene**: the other session's in-flight `pkg/primitivetype` refactor (G2 name-blocklist deletion) was never touched; its breakage landed mid-session and healed on its own; both daemons committed both changesets without collision (branching-flow HEAD `e33012ae` contains my strongid files; webphone HEAD `6863e8a`).

## b) PARTIALLY DONE

1. **branching-flow release**: code + CHANGELOG + TODO_LIST done; the release itself (tag, flake `version` bump per its AGENTS checklist, system-profile rebuild so PATH picks it up) is deliberately NOT done — owner ritual. Until then there is a **split brain between the installed 0.2.0 binary and the source**: PATH `branching-flow strong-id` on webphone shows 2 rows; the repo build shows 0.
2. **Passkey door behavior-change verification**: pinned by unit + handler tests (`TestPasskeyFinishRejectsAMalformedUserID`, the missing-user_id 400s, enroll-handler tests, userauth suite). The stack browser E2E was NOT re-run — no served markup changed (so not owed by the runbook rule), but the finish/enroll handshake is exactly what the E2E exercises.
3. **branching-flow lint scope**: golangci ran on the two touched package globs only, not the whole repo; branching-flow's own buildflow gate was never run (its tree also carried the other session's uncommitted refactor at the time — a full gate would have judged a mixed changeset).
4. **`--show-fix` suppression semantics**: fix output now renders the visible (non-suppressed) set; no test pins that interaction.

## c) NOT STARTED (owed or noticed during the session)

1. **webphone push**: `main...origin/main [ahead 7]` at session end — the daemon commits but had not pushed; per the runbook the stack tree must see webphone main before pbx-artmann re-pins. Needs `git ls-remote` confirmation once the daemon pushes.
2. **HARVEST**: section (f) below is docs-health HARVEST input for `TODO_LIST.md`/`ROADMAP.md` — not harvested (dispatch says wait).
3. **webphone CHANGELOG entry** for the user-visible passkey behavior change (garbage `user_id`: 400 instead of the old hash-masked 401).
4. **Lesson not recorded**: `nix flake check` / treefmt-check builds the **committed** git tree — uncommitted fixes are invisible to the drv (same-hash rebuild proved it). Belongs in docs/lessons.md.
5. **Remaining usermgmt v4 deprecations** in webphone (pre-existing, intentionally untouched): `ErrEmailExists`, `ErrUserNotFound`, `ErrAccountLocked`, `ErrNoCredentials`, `WebAuthnProvider` → direct `identity-model/v4` imports.
6. **Exit-code golden test** in branching-flow for the suppressed-only case (verified manually once, unpinned).

## d) TOTALLY FUCKED UP

Nothing in the END STATE is broken — both repos are green and committed. But two process fuckups cost real cycles and must be named:

1. **I used a suppression mechanism that did not exist.** My first `//nolint:branching-flow` attempt (placed above the field) was never verified; the second attempt (same line) also failed; only then did I read branching-flow's source and discover `pkg/strongid/to_findings.go` had NO suppression support at all. Two edit rounds were spent on a fiction. I should have read the tool's suppression implementation FIRST (the `--include-suppressed` help text was the only evidence it existed), or verified the first directive immediately.
2. **Verification-by-vibes moments** (each small, together a pattern):
   - `nix build .#checks.x86_64-linux.treefmt` printed nothing and I moved on WITHOUT checking the exit code — the "success" was never actually confirmed.
   - Grepped callers for the crm/userauth retyping but missed TEST call sites (`match.ID != "01M"`, `enrollCredential` returning `string`) → two avoidable compile round-trips.
   - The erraudit `context_loss` finding on my `Wrapf` surfaced only at the final full gate, not when I wrote `enroll.go` — running the single erraudit step right after the edit would have saved a full gate cycle.
   - One multi-value-return compile error in `strong_id.go` (`len(...), Render(...)`) — sloppy.
   - Wrote an unsolicited comment into `sessionIdentity`, then removed it — should not have been written.
   - `rg -rn` misuse twice (the `-r` REPLACE flag mangled output I initially misread).

## e) WHAT WE SHOULD IMPROVE

1. **Verify the smallest unit immediately**: after ANY suppression/annotation/lint-directive edit, re-run the exact tool before touching anything else. "Test after changes" applies to annotations too.
2. **Read the enforcement mechanism before relying on it**: a flag's help text is a claim, not evidence (this session's core lesson — same shape as verify-external-claims, aimed inward at our own tooling).
3. **Always `nix fmt` before gates** — the AGENTS rule exists because hand-edited Go drifts; I re-learned it at the cost of a gate round.
4. **When retyping a signature, grep tests too** (`rg` over `*_test.go` for the old literal/param shape) before compiling.
5. **Run the single relevant gate step after the single relevant edit** (`buildflow -s erraudit`, `-s go-lint`) instead of saving everything for the monolithic final gate.
6. **Close the tool/source split fast**: an installed binary that disagrees with source makes every future strong-id run lie. Release or don't claim.
7. **Exit codes, not silence**: every "it printed nothing" verification needs an explicit `$?`.
8. **Narrative commits**: the daemon shredded this train into 8 `chore: auto-commit N file(s)` commits (webphone) — history has no phase boundaries. Where a train matters, commit phase boundaries deliberately (harness permitting).
9. **Type-model follow-up (design)**: `identities` config values are validated ("has dialable characters") but stay strings; parsing them into `domain.Phone` at config load would eventually let `sessionIdentity.DID` become honest AND open branding the whoami surfaces.
10. **Upstream idea for go-branded-id**: an opt-in JSON marshal that emits the `"Brand:value"` form would make wire DTOs like `apiContact.ID` brandable byte-identically — the one thing that blocked the last two findings. Feature-request candidate (verify-before-filing first).

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

_Brainstorm, not commitment — HARVEST should route these (most are ROADMAP fuel). Ordered by impact within groups._

**This train's loose ends**

1. Release branching-flow (tag + flake `version`, its AGENTS checklist) so source and installed binary agree.
2. Rebuild system profile; confirm PATH `branching-flow strong-id` shows 0 rows on webphone.
3. Confirm webphone push landed (`git ls-remote`) — 7 commits were unpushed at session end.
4. HARVEST this report's section (f) into webphone `TODO_LIST.md` / `ROADMAP.md` (docs-health).
5. Add the webphone CHANGELOG entry for the malformed-`user_id` 400 behavior change.
6. Re-run the stack browser E2E once against a build with the strict `ParseUserID` door.
7. Record the "flake checks build the committed tree" lesson in docs/lessons.md.
8. Smoke a fresh binary (`scripts/webphone-smoke.py`) with passkey mode on, covering enroll→login once.

**branching-flow hardening**
9. Golden test: `strong-id --exit-code` exits 0 when only suppressed findings remain.
10. Test: `--show-fix` interaction with suppressions (visible set rendered).
11. Run branching-flow's own buildflow gate on a clean tree (post-primitivetype-landing).
12. Whole-repo golangci (not just touched globs).
13. Docs page for strong-id suppression semantics (CLI help already claims it; the docs should teach it).
14. Consider backfilling the same suppression audit for the OTHER analyzers — which rules accept `//nolint:branching-flow` without implementing it? (roleak/panicanalyzer/phantom verified; dupe/splitbrain/contextguard/nakedreturn/flagparam/ifacecomplete/doanalyzerv2/boolblind UNAUDITED.)
15. Release-notes cross-link from the `--include-suppressed` flag help to the implemented rules.
16. Unit test for `newStandardRunnerWithSuppression` + `--include-suppressed` widening (runner-level, not just dropSuppressed).

**webphone type-safety continuation**
17. Migrate the 5 remaining `usermgmt` v4 re-export deprecations to direct `identity-model/v4` imports (kills 10+ diagnostics).
18. Parse `identities` config values into `domain.Phone` at config load; type `identityFor` accordingly.
19. Then re-evaluate `sessionIdentity.DID` branding (blocked today by omitempty wire semantics, not just JSON shape).
20. `sessionIdentity.Extension` is still a raw wire string — fine as DTO, but document it next to the DID nolint reasoning so the next strong-id sweep doesn't re-litigate.
21. Upstream go-branded-id: opt-in prefixed JSON marshal (see e.10) — would un-brand-block `apiContact.ID` honestly.
22. Re-evaluate `apiContact.ID` branding IF 21 lands; keep the nolint reason in sync.
23. gopls `embedlit` hint at `session_api.go:127` (pre-existing) — one-line struct-literal cleanup.
24. gopls `slicescontains` hint at `userauth.go:203` (pre-existing) — `slices.Contains` simplification.
25. gopls `unusedparams` hint at `passkey_api.go:211` (pre-existing) — the stub-provider seam's `r` param.
26. `nilness` tautology warning at `userauth.go:155` (pre-existing, `Shutdown` nil-check shape).

**Docs / memory**
27. AGENTS.md: add the committed-tree flake-check gotcha next to the existing Gotcha bullet about `self`.
28. Error-contract: the malformed-user_id row cites both the malformed and missing tests — tighten to the precise test home.
29. AGENTS.md strong-id ruling: mention `DocTypeID` (the report had suggested `TypeID`) so future sweeps don't rename it back.
30. Record in lessons.md: `rg -rn` foot-gun in this shell (replace-flag misinvocation), alongside the existing jq warning.
31. Dedup registry: append a sweep-log line for this session's accepted similarities (`NewContactRef` wrapper vs domain Parse pattern; the three paperless brand blocks).
32. FEATURES.md: the passkey door's strict-parse behavior belongs in the passkey section if it inventories wire behavior.
33. branching-flow TODO_LIST release item: add "verify webphone strong-id 0-rows post-release" as the acceptance check.

**Tooling / gates**
34. Webphone: add a scripted strong-id check (pinned post-release binary, `--exit-code`) so the report can never silently regrow.
35. Same pattern for branching-flow: dogfood `strong-id ./... --exit-code` in its own CI once released.
36. `nix fmt` as an explicit pre-gate step in my session checklist (memory-worthy, not just AGENTS-worthy).
37. Investigate the buildflow "regression verdict: step(s) slower" noise — cold-cache runs flag ~15 steps every time; baseline tuning or a warm-run policy would un-nihilate the verdict line.
38. vulnix retry warning ("0/5 retried runs recovered") — the AGENTS notes it gates via `webphone-vulnix-triage`; confirm the warning is known-noise or file it.
39. PostHog telemetry 403 (buildflow warning) — disable via `BUILDFLOW_TELEMETRY_DISABLED=true` in the shell or fix endpoint.
40. The 9-tools-unavailable buildflow health warning — AGENTS calls it noise; consider suppressing in webphone's buildflow config for honest verdict lines.

**Hygiene / pre-existing debt noticed (not introduced this session)**
41. `internal/crm/family_test.go` uses `client.LogCall` with a real HTTP stub — fine, but the branded-ref construction in tests could share one helper.
42. `userauth_test.go:173` errcheck (`testDB.Close` unchecked) — pre-existing defer without nolint; one-line fix.
43. `lookupResponse` in crm duplicates `Match`'s name-derivation logic inline — acceptable, but a comment pairing them would prevent drift.
44. `passkey_api.go` maps `enrolled.UserID.Get().String()` — if go-branded-id grows `ValueString()` sugar, adopt it everywhere `.Get().String()` appears (≥4 sites now).
45. Consider a `userauth` package-level `MustParseUserID` re-export so server tests don't import usermgmt directly for the Must form.
46. The daemon's heuristic commit messages lose train narratives — decide whether phase-boundary manual commits are wanted (harness conflict to resolve).
47. `docs/status/` non-archived reports are accumulating (2026-10-05 already had 2 from the other session) — periodic archive sweep per docs-health.
48. go-branded-id pin check: v0.7.0 is current in webphone's go.mod — keep on the daemon sweep.
49. `internal/server/passkey_api.go` header comment says "finish takes ?user_id=<session_key>" — still true, but the strict-ULID door deserves one sentence there.
50. Re-run `branching-flow strong-id --include-tests` once on webphone — test files were excluded by default this session; there may be literal-bearing test helpers worth typing or suppressing deliberately.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Release timing for branching-flow**: should I cut the release now (tag + flake `version` + profile rebuild) so the installed binary stops lying about suppressions — or is branching-flow riding a release train with other pending work (the concurrent session's primitivetype G2 change is also freshly committed) and strong-id suppression should just wait for it?
2. **Gate or advisory?**: is `strong-id` meant to become an enforced gate (`--exit-code` in CI/BuildFlow) or stay an advisory manual sweep? This decides how strict suppression reasons and the AGENTS ruling need to be, and whether item 34 is a must or a nice-to-have.
3. **Deprecation migration scope**: should the remaining `usermgmt` v4 re-export deprecations (direct `identity-model/v4` imports; ~10 standing diagnostics) be migrated in the next webphone train, or are they parked until cqrs-htmx removes the re-exports?

---

_Point-in-time snapshot. Section (f) is HARVEST input — not yet routed into TODO_LIST/ROADMAP (dispatch says wait for instructions)._
