# Whole-TODO-list execution session — status report

> **Snapshot-rot header (F85 convention): claims verified as of 2026-10-25 15:16 CET — `git ls-remote` == local HEAD `3768360`, working tree clean, every CI run on main today green (`gh run list`: 40195227784, 40195602264, 40199154068, 40200122826 all success; Dependency Graph dynamic runs green). Push/CI statements below rot fastest.**

Session window: ~13:20–15:16 CET, 2026-10-25. One assistant session (Crush/GLM-5.3). A SECOND session was demonstrably active on this checkout during the window (evidence in §d/§e); go-cqrs-lite also gained issue #90 from outside this session between my dupe search and my filings.

## a) FULLY DONE

1. **go.sum daemon-commit triage (the watched defect class)** — unpushed `5ccec34` deleted 27 go.sum lines (all stale `/go.mod` hash entries). Verified `go mod tidy` is a no-op on that state (canonical), verified the goModules FOD builds against the pin (insensitive to those lines; CI green on BOTH the +27 and −27 trees), daemon pushed it, CI 40195602264 green. No action needed — documented verdict.
2. **TODO row-hygiene, evidence re-verified at HEAD** (`45c41e4`): go-appkit section deleted (flight recorder `internal/app/flightrecorder.go` + latency metrics `internal/server/latency.go` verified shipped); branching-flow unskip item 3 marked EXECUTED (`skip_steps` = pytest-test only); catalog-push checkbox closed (`46cc6db` verified on origin/main; remaining leg is the hub re-pin, cross-repo); round-5 F74/F85/F86 closed (snapshot-rot header verified in both 10-18 reports + AGENTS Conventions line); stale passkey gopls handoff note retired (LSP diagnostics on `passkey_api.go` clean).
3. **Push-seam Settings surface** (the native-app train's deferred in-repo row): `settings.templ` gained a conditional "Ring while closed" DefinitionList row (`asrDetail` generalized to `serviceDetailExtra`, 2 call sites) + read-only `PushDevicesSection` (device id · transport · date, `wp-snippet-*` classes reused — no tw.css rebuild); 3 new i18n keys in BOTH maps; `panels.go` wiring (nil-safe, error-propagating, wake address never leaves the store); `TestSettingsPushSection` (view: off=absent, on=row+list+count, empty state) + `TestSettingsTabSurfacesPushDevices` (served tab: row + device id present, token absent; seam-off renders neither). i18n parity green. CHANGELOG entry. Zero-config default byte-identical.
4. **ASR #13 saved-to-History discoverability**: `calls.js` — marker node (hidden until first CONFIRMED save), `entry.transcriptSaved` count, `showTranscriptSaved` (persistent count + once-per-call polite announce), `renderCalls` re-labels on language switch; island i18n en/de `transcriptSavedToHistory`; `.call-transcript-saved` rule in island/style.css; new spec pins reveal/count/announce-once. Island suite 235/235 (documented glob form).
5. **ASR #21 whisper.cpp provider variant**: `asr.provider = "whispercpp"` — `inferencePath` const, `Client.path` field, `NewWhisperCppClient`, kind-aware `Describe()` ("whisper.cpp · model"), config validation accepts the third value, `app.go` picker branch, `TestWhisperCppInferencePath` (path + bearer + multipart shape + text decode + Describe), README `asr.provider` row updated. Default config untouched.
6. **metaengine v4.19.0→v4.20.0 bump train** (go-cqrs-lite/id rode to v4.7.3): codec contract read at source (`stream_codec.go` — strings as-is / JSON else, `EncodeStreamValueErr` supersedes the deprecated wrapper; webphone has ZERO direct stream-codec use); `go mod tidy`; vendorHash re-measured via the documented `--rebuild` recipe and re-pinned (`0c90e7b`); readmodels + app suites green, full `go test -race ./...` green; buildflow exit 0 (after fixing what the first run caught, see §d); `contracts-fresh`/`island-js`/`island-lint`/`format` checks built green; `webphone-module` eval green; KVM `webphone-backup-drill` green; asyncapi.yaml regenerated (the shared graph's YAML style flipped — proven content-inert: stripped-whitespace diff shows only `"`→`'` ref quoting + list indentation, 485/485 lines); committed `875c3f8` + docs `f33a8dc`; pushed; CI 40199154068 all four jobs green.
7. **The seven catalog-exporter upstream filings — go-cqrs-lite #91–#97**, all premises verified in the PINNED v4.9.0 source with file:line cites, dupe search over issues AND PRs clean, drafts in `docs/drafts/catalog-exporter-2026-10-25/` (voice checker: 0 FAIL / 0 WARN ×7), filed via `--body-file`, all seven verified landed (#91 response headers · #92 204 body-schema exemption · #93 WithParam types · #94 kafka default server · #95 multipart bodies · #96 operationId dots — `caseutil/convert.go:76-78` separator set quoted · #97 OpenAPI 3.1.0 — `exporter.go:14` pin). TODO row closed with the workaround-retirement watch (`3768360`).
8. **Docs**: CHANGELOG `[Unreleased]` gained 3 Added lines + 1 Changed line; TODO_LIST gained the 2026-10-25 sweep line; native-app deferred row + ASR rows updated to shipped state.
9. **End state**: local == remote at `3768360`, tree clean, 11 commits pushed this session (2 narrative feature commits, the bump pair, the contract regen, the filings commit, docs; the rest daemon chunks of my in-flight edits — all content verified intact).

## b) PARTIALLY DONE

1. **Local gate coverage for the whisper.cpp delta**: asr/config/app packages ran locally; the FULL local race suite last ran BEFORE the whisper.cpp code — the whole-tree guarantee for it comes from CI's `go test -race` job on `3768360` (green). A final local full-suite sweep after the last code change is owed (§f 2).
2. **Smoke coverage**: live smoke 48+4+8/0-failed ran after the Settings/ASR13 work but BEFORE whisper.cpp; the whispercpp wire has Go-level tests but no smoke scenario (§f 1).
3. **`nix flake check`**: reported "running 0 flake checks" (the KVM `requiredFeatures` gating makes nix skip the VM checks; the non-VM checks I built explicitly, all green). The full-flake-check PACKAGES verdict exists, but a literal full `nix flake check` that RUNS its checks was not produced locally this session.
4. **Vulnix**: NOT run over the bumped runtime closure (`nix run .#vulnix`); the bump-train notes' recipe doesn't include it, but release.sh gates on it — the next release will catch it (§f 3).
5. **aarch64**: the standing ELF-byte verify at final-gate cross-builds was not performed on the post-bump tree (§f 4).
6. **Push devices Settings row v1 is read-only by design** (registration/deregistration stay the shells' login/logout contract) — a per-device revoke affordance is a plausible v2 but is a product ruling, not a defect (§g 2).
7. **Plan-doc annotations**: `docs/planning/2026-10-24_20-05_dependency-train-notes.md` (now GO-executed) and the completion plan's upstream-candidates tail (now filed) were NOT annotated — the TODO rows carry the live state (one-home), but the source notes still read "when GO" / "carried until filed" (§f 9/10).
8. **go-appkit row deletion**: TODO section removed per docs-health, but I did not re-verify FEATURES.md carries the flight-recorder/latency rows (they shipped 2026-10-23 via another session; likely recorded, unverified by me).

## c) NOT STARTED (all gated, none forgotten)

Owner-terminal: /health host rebuild + SSO browser check · T13/T15 unified-auth cutover cluster · passkey deploy proofs + enroll-begin ULID-risk verdict · boot-contract D3 retry-loop call · v2.8.0 deploy tail (stack relock dance) · prod bridge journal leg · release announcements · owner-calls batch sitting (38-row briefing) · ASR #48 privacy call · /openapi.json ETag ruling · catalog v4.10.0 bump lean · mic pre-warm live ritual · visual-harness disposition · samber/do owner calls · markdownlint posture · devShell additions + lychee policy · pma daemon guard-hole fix (nix-config).
Cross-repo (NOT this repo): hub re-pin/build after the catalog push · hub cross-link pipeline home · webphone-android spec consumption + first device train · stack ring dialplan leg · stack browser E2E debt (mod_enum repair first) · signal-cli/matrix stack deployment · gateway stack-side bits · stack hardening tail.
Parked/scheduled: session messenger NO-GO (named triggers) · scheduled sends DECIDED-AGAINST · standing watches (quarterly 2026-12-20; erraudit monthly 2026-11-05) · quarterly dup baseline audit 2027-01-11 · modularize re-trigger ~2027-01.

## d) TOTALLY FUCKED UP

Nothing landed broken — every pushed state is CI-green and the tree is clean. The honest list of f*ck-ups and near-misses:

1. **The stale-vendorHash push-risk window (worst moment of the session)**: I ran `go get` on MAIN first and only then tried to retreat into a scratch worktree — by which time the daemon had already committed the half-bump (`efc862a`: bumped go.mod/go.sum + STALE pin). For ~10 minutes main carried a pushable red state; had the daemon's push leg fired in that window, CI reds on the hash mismatch (the 2026-10-06 class). It worked out because I re-pinned fast (`0c90e7b`) and the daemon was slow — that is luck, not process. Correct sequence: worktree/branch FIRST, go get inside it, land on main only when the pin is in the same commit set.
2. **`nix/packages.nix` concurrent-edit collision**: between my View and my Edit, the OTHER session rewrote the vendorHash line (an `AAAA…` placeholder flickered in my `rg` output); my edit failed "old_string not found", and the tree then settled to EXACTLY the hash I had measured (they ran the same deterministic recipe). Outcome correct, but I edited a file another session demonstrably held without re-Viewing immediately before the edit — the AGENTS rule I know verbatim.
3. **`nix flake check` "running 0 flake checks" nearly passed as a pass**: I initially read "all checks passed!" as gate-green. Probing found `requiredFeatures.kvm` makes nix SKIP the VM checks and the non-VM ones were only covered because I then built them explicitly. "0 checks ran" must be treated as a FAILURE of the gate, not a pass — this is the unverified-gate class from docs/lessons.md, caught late.
4. **First buildflow run raced my own parallel builds**: timing-regression noise across 15 steps (+650%…+62%) and a wasted exit-69 investigation. Run gates on an idle machine; don't parallelize nix/go-heavy work with the gate.
5. **Command hygiene (recurring, known traps)**: tripped the `rg -rn` replace-flag hook block TWICE (the hook's message says third strike on this host — it should have been zero); ran `node --test <dir>` (wrong form, spurious failure) before the documented glob form; `nix eval --raw builtins.currentSystem` without `--impure` (parsed as a path). Each cost a round trip; none cost correctness.
6. **Four stale-read edit failures** (TODO_LIST, CHANGELOG, README, and the packages.nix collision): all daemon/concurrent-session reflow tax, but each was avoidable with an immediate re-View before Edit.
7. **Worktree churn**: created `/tmp/wp-bump` + branch after they were already moot, deleted both unused — pure waste caused by (1).

## e) WHAT WE SHOULD IMPROVE

1. **Dependency bumps: worktree-first is the doctrine, not a preference.** The daemon commits main every few minutes; any multi-step bump on main has a pushable-red window. Proposal: record it in AGENTS (Conventions) alongside the existing dep-lane rule.
2. **Gate-verdict literacy**: "0 checks ran" / cached FOD / detect-only tools must print their RUN COUNT next to the verdict in my working habit — a gate that runs nothing is not green. Candidate one-liner for docs/lessons.md (§f 8).
3. **Re-View before EVERY edit on shared files** (TODO_LIST, i18n.go, settings.templ, nix/*, CHANGELOG): the daemon demonstrably reflows between my read and my write; four failures this session is four too many.
4. **Annotate source plan/notes docs when executing them** (dependency-train-notes GO outcome, completion-plan tail → issue numbers) so the next session doesn't re-derive state from TODO rows alone.
5. **Last-change-then-full-suite**: run the full local race suite AFTER the final code change of a session, not only before the last one — CI covering it is necessary but late.
6. **Concurrent-session awareness**: today's session overlapped another on the same checkout (daemon chunks mid-edit, packages.nix flicker, buildflow SQLITE_BUSY, go-cqrs-lite #90 appearing mid-session). Before filings or shared-file edits in such a window, check `git log --since` + `gh issue list` freshness first.
7. **Smoke parity for new wire dialects**: a provider variant without a smoke scenario ships with a thinner black-box guarantee than the feature it mirrors — add the fake-provider variant in the same train (§f 1).

## f) Up to 50 things to get done next

In-repo, actionable (this session's tails first):

1. Add a whisper.cpp scenario to the smoke's fake ASR provider (`asr_scenario` variant serving `/inference`), wire into main.
2. Full local `go test -race ./...` sweep on the final tree (post-whisper.cpp; CI already proved it, close the local loop).
3. `nix run .#vulnix` over the bumped runtime closure (pre-release dry run of the release.sh gate).
4. aarch64 cross-build + ELF-byte verify on the post-bump tree (standing final-gate rule).
5. Settings push row: SSE/morph live-update when devices register/deregister mid-session (currently render-time truth).
6. Settings push row: device count in the DefinitionList detail (e.g. "fcm+ntfy · 2 devices").
7. Ruling + (if yes) build: per-device revoke buttons in Settings (passkey-pattern `hx-delete` + panel re-render) — see §g 2.
8. docs/lessons.md one-liner: "`nix flake check` green with 0 checks run is not a pass" (requiredFeatures/KVM gating class).
9. Annotate `dependency-train-notes.md` with the executed-GO outcome + evidence hashes.
10. Annotate the completion plan's upstream tail with #91–#97.
11. Watch go-cqrs-lite #91–#97; on each fix landing, retire the local workaround in the same train (typed empty-object 204s → honest schemaless 204s + pin updates; Retry-After description text → typed header; `WithServer("","")` suppression; operationId string pins after #96).
12. After #96-style operationId changes: regenerate `/openapi.json` + sweep the three pin sites (Go tests, smoke, stack browser-e2e) for operationId literals.
13. Confirm nothing greps the OLD config.asr.provider rejection text (the enumeration grew to "openai, google, whispercpp").
14. Optionally extend the ui-capture ASR-on matrix with a saved-marker state on the call card.
15. Run buildflow once more on an idle machine to clear the timing-regression noise from the baseline record.
16. FEATURES.md check: flight-recorder + latency-metrics rows present (go-appkit TODO deletion follow-through).
17. Error-contract/table freshness spot-check after the provider-enum message change (TestErrorCodeRegistryIsFresh green, but eyeball the rendered row).
    Owner-gated (bring to the sitting, do not self-answer):
18. OWNER-calls batch sitting (38-row briefing + everything accrued since).
19. ASR #48 remote-leg-only transcription default (privacy).
20. /openapi.json ETag + golden fixture ruling.
21. catalog v4.10.0 bump — ride the next natural train; revisit when #91–#97 fixes land (v4.11+).
22. §g 2: Settings push-device revoke — yes/no.
23. §g 3: who implements #91–#97 upstream.
    Cross-repo (route, don't build here):
24. Hub re-pin/build (the catalog-pushed leg of the pair).
25. Hub cross-link pipeline home (generated pages need a links field, not a direct mdx edit).
26. webphone-android: consume `/openapi.json` + asyncapi ring contract; `/api/session` field check.
27. Stack leg: FS dialplan POSTs `webphone:/hooks/ring` (ring-while-closed goes live end-to-end).
28. Stack browser E2E debt: repair FreeSWITCH `mod_enum`, then one re-run covers the Settings/push/ASR13/whispercpp served-markup deltas + the island CALL_TRANSITIONS obligation.
29. Stack: signal-cli systemd unit + matrix homeserver config (feeds webphone via settings).
30. Chat lane v2: matrix groups + read receipts, matrix E2EE (big), signal inbound attachments, signal groups, compose-new-DM affordance.
31. Stack hardening tail: shared `secretFile` fixture, `fs-directory` into `operatorWatchedUnits`, sipcred wire vector into the stack fixture, xml_curl × fail2ban check, suite merge consideration.
32. Gateway stack-side bits: `ftypqt` sniff, E2E MMS-outbound, pbx-artmann FEATURES:87 text.
    Owner-terminal (deploy/ops):
33. /health host rebuild + logged-out SSO redirect + signed-in cards check.
34. T13/T15 cutover cluster (xml_curl staged cutover + rollback drill + T15 live proofs).
35. Passkey live proof + fail-closed password-file drill; enroll-begin ULID-risk verdict.
36. Boot-contract D3: systemd StartLimitBurst vs ratify-5s-retry.
37. v2.8.0+ deploy tail: stack relock dance (incl. the nginx→caddy `mkForce` swaps) + post-deploy smoke `--expect-version`.
38. Prod bridge failure journal leg (422 outbound SMS, command sheet §4 tree).
39. Post-release announcements (drafts exist; owner picks channels + disclosure posture).
40. Mic pre-warm live ritual (accept→speak, indicator timing, warm release).
41. Visual harness: eyeball the 20-shot matrix; per-release persistence decision; vision-CLI cross-check provider/key.
42. samber/do train owner calls: stack `/health` exposure policy, v2.9.0 fold decision.
43. End-to-end ring proof: stack leg + android leg + one real device.
    Standing/scheduled:
44. erraudit tier-2 monthly re-measure (due 2026-11-05).
45. Standing-watches quarterly re-check (due 2026-12-20: sip.js 0.22, templ-components, oxlint globals, E2E budget).
46. Quarterly art-dupl baseline/suppressed-set audit (first 2027-01-11).
47. Modularize re-trigger re-review (~2027-01; detection commands pre-armed).
48. pma daemon GIT_CONFIG guard-hole fix (nix-config owner edit; fleet-wide hook binding).
49. Dependency-bump lane adoption: `buildflow update`/Dependabot so bumps stop touching daemon commits at all.
50. Sessions credentials-at-rest periodic re-review (due 2026-12-30 or next session-storage change).

## g) Questions I cannot answer myself

1. **Was another session of yours running on this checkout between ~13:20 and ~15:00 today** (it committed `efc862a`, flickered `nix/packages.nix` through a placeholder hash, held the buildflow resume DB, and filed go-cqrs-lite #90 mid-session)? If it is still active I will re-check shared files before any follow-up; if it was you driving a second train, should my filings/commits have been coordinated with it differently?
2. **Settings push devices: read-only is shipped — do you want per-device revoke buttons in Settings** (passkey-pattern `hx-delete` + panel re-render), or does deregistration stay exclusively the shells' login/logout contract? A revoke that the device itself never learns about is a small island-honesty question I don't want to answer alone.
3. **go-cqrs-lite #91–#97: file-only, or implement?** You own the library — should a follow-up session implement the fixes upstream (the caseutil dot-separator and the 204 exemption are one-liners; headers/multipart/3.1.0 are real features), and if so, do any of them gate the next catalog bump train you want to ride?
