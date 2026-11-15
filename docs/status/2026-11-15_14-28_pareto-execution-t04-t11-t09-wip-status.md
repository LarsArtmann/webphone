# Pareto execution — T04–T11 done, T09 in flight (well, in a VM) — status claims verified as of 14:28 CET

> Snapshot-rot header: push/CI statements below are timestamped; the CI
> verdict on `73054c22` was read at 14:28 CET (RED — attribution inside).
> Plan of record:
> `docs/planning/2026-11-14_11-41_SUPERB-open-queue-pareto-plan.md`.
> Session context: the owner's standing "GET SHIT DONE" GO; the interim
> report (2026-11-15_12-20) covered T01–T03 + T07/T12–T16/T19–T24 — this
> one covers the rest of the tier-1/4/20% queue.

## a) FULLY done (committed AND pushed)

| Task                                  | What landed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Push       |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| CI heal (pre-work)                    | `collapseToggle` templ component = the ONE home for the rail train's twin toggles; art-dupl baseline re-recorded 110 → 111 (post-write `check` = 0 new), sweep-log row in docs/dedup-registry.md; tw.css/login.css rebuilt (the rail train's app.css tokens had never regenerated login.css)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `b9f09157` |
| T04 (F18–F24)                         | `TestSummaryDraftFailedAttemptNeverRecordsIdem` (both failure arms leave the idem window empty), `TestSummaryDraftReplayFansNoSecondPush` (Settle-backed, exactly one draft fact), error-context self-test pins `agent.summarize`/`push.fcm_send`/`cqrs.note_list_full` (**67 checks**; the plan's `push.draft_*` names codes that no longer exist — the fanout wraps under `push.transport`/`push.*_send`), error-contract § "The summarize chain's two error lanes" (`textagent.*` verdicts vs the `agent.summarize` 502 wrap), hooks `fakeTransport` records drafts natively (lane-separable), `TestFcmSendDraftFullWire` (draft data + NORMAL priority), `agentharness_test.go` = the ONE home for fakeProvider/stubReviewAPI/asrClientFor/recordingTransport/waitForDrafts/mustExt | `3fd06ed4` |
| T05 (F25–F28)                         | Smoke `agent scenario` (6 checks: off ⇒ config.js silent + disabled-seam 404; on ⇒ `"agent":true` + the route reaches the transcript gate), smoke §13c notes PUT/GET/DELETE roundtrip (suite: **58 passed, 0 failed** — twice), `docs/status/archived` joined the lychee point-in-time shield, AGENTS smoke verdict-line prose + error-context 67-checks prose                                                                                                                                                                                                                                                                                                                                                                                                                          | `d671f961` |
| T06 (F29–F33)                         | `package/textagent.nix` (`enable`/`url`/`secretFile`; LoadCredential; eval both-or-neither + absolute-http(s) mirror; the secretFile joins the eval-time store-leak guard), `nix/module-check-textagent.nix` (4 cases: typed render incl. the sanctioned-file secrets pass, fail-closed halves, raw lane untouched, typed/raw conflict)                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `b9fdcaa0` |
| T08 (F38–F40)                         | FIRST recorded cross-package union coverage: **75.8%** at `b9fdcaa0` (`-coverpkg=./...` + `go tool cover -func`), dated section appended in docs/reviews/2026-09-20_coverage-baseline.md; the views 1.3%-direct vs 27.0%-in-`./...` contrast is the structural-undershoot proof. BuildFlow's `test-coverage` cannot carry `-coverpkg` (binary verified; no config key reaches the argv) — the knob ask is drafted for the owner: docs/drafts/2026-11-15_buildflow-test-coverage-coverpkg.md                                                                                                                                                                                                                                                                                             | `92a66ead` |
| T10 (F47–F49)                         | `nix/module-output-oauth2.nix` (single source, module-output.nix recipe) + committed golden (66 lines) pinning the full three-section rendering WITH the SSO lane on — the settings JSON carries the callback (`https://phone.example.org/api/auth/oauth2/callback`) and passkey `rp_id`/`rp_origins` derivations from `caddy.hostName`, so the F48 derivation assertion IS the fixture's bytes; `oauth2-enabled-output-golden` case green (35 module-check cases total)                                                                                                                                                                                                                                                                                                                | `dfebb31b` |
| T11 (F50–F53)                         | `package/passkey.nix` — `services.webphone.passkey.enable` derives `settings.auth.passkey.{rp_id,rp_origins}` from `caddy.hostName` for PASSKEY-ONLY deployments (mkDefault; an operator pin wins), with the OIDC lane's users-map + password-file assertions mirrored; `nix/module-check-passkey.nix` (3 cases: typed render incl. no-credential, fail-closed halves, raw lane + pin-beats-derivation)                                                                                                                                                                                                                                                                                                                                                                                 | `73054c22` |
| vendorHash repair (unplanned, urgent) | Daemon commit `e85830ef` (13:39) STOMPED the three real hashes in nix/packages.nix to `AAAA…` placeholders (templ hash, templ vendorHash, webphone vendorHash) — origin main carried them ~30 min, breaking every fresh package build. Restored from the parent commit (`0ca42edd`, pushed 14:26); `nix build .#webphone.goModules` verified green (the restored webphone hash matches the fresh `got:` exactly)                                                                                                                                                                                                                                                                                                                                                                        | `0ca42edd` |

## b) PARTIALLY done

- **T09 (OIDC VM test, F41–F46)** — written, registered, KVM-gated
  (`webphone-oauth2` in nix/vm-tests.nix, pushed via the daemon's
  `a6648cd1`/`924e21a0`), eval-fixed twice (caddy.enable + hostName for
  the two derivations; https issuer per the bareHTTPS assertion), and
  the VM now BOOTS through the credential asserts' first half — the
  file EXISTS at /run/credentials — but fails at
  `stat -c %a = 400` (24.4 s in). Two open threads:
  1. the credential mode assert: systemd's LoadCredential landed the
     file but the mode read something else — need the actual stat value
     (or `systemd-analyze cat` of the unit) before re-pinning the
     assert;
  2. the harness's `caddy.enable = true` makes the VM's caddy attempt
     REAL ACME (Let's Encrypt lookups that fail without network) —
     cosmetic noise but a design smell; the backup/error-drill VM tests
     run caddy-free, and this one probably should too (explicit
     `redirectUrl` + hand-written rp instead of the caddy derivation).
- **Train end** — the per-task batteries ran green (docs-gates --check
  after every docs commit; smoke twice; module checks; island suite
  untouched this half), but the FINAL battery (full `go test -count=1`,
  `nix flake check`, CI-watch to green) has NOT run yet.

## c) NOT started

- **T18 (F83–F87)** — night-run stability pair + ring-guard positive path.
- **T17 (F79–F82)** — architecture-understanding D2/SVG refresh vs the
  ~38-package reality (the row 1668 wording).
- **TODO row strike for T09** (row 147) — gates on the VM test going
  green; the row stays honestly open until then.
- **CI-watch to green** on the post-restore tip (`0ca42edd`) — CI on
  `73054c22` was RED at 14:28 (check-runs attribution not yet pulled;
  prime suspect: the placeholder hashes broke the package build job the
  restore now fixes, plus possibly the rail session's standing reds).

## d) TOTALLY fucked up (owned, with the shape it taught)

1. **`git add -A` swept a concurrent session's in-flight edit onto
   main.** At 13:39 I staged the passkey work with `git add -A` while a
   foreign session had `nix/packages.nix` placeholder-stomped in the
   worktree; the daemon committed the sweep (`e85830ef`) and the
   placeholders landed on ORIGIN for ~30 minutes, breaking every fresh
   package build. Lesson the hard way: with a live concurrent session,
   stage EXPLICIT paths, never `-A`; and `git show --stat` a daemon
   commit that carries my files before trusting it.
2. **The vendorHash recipe's warning was ignored at train start.** The
   recipe says a plain build masks a stale/fake FOD; I ran `nix build`
   on CHECKS only (which don't build the package) and never
   `nix build .#webphone.goModules` up front — the placeholder sat
   until the VM test forced the fresh build. New train-start step:
   run the goModules FOD check BEFORE the first task.
3. **Asserts written before reading the wire.** The smoke 13c "empty
   200" assert failed because `GetCallNote` honestly answers
   `{"body":""}` — I wrote the expectation from AGENTS prose instead of
   note_api.go's writeJSONBody shape. Same class: the VM's 0400 stat
   assert. Read the exact shape FIRST, then write the assert.
4. **The VM test enabled caddy** (for the two derivations) without
   thinking about what the rendered vhost DOES in a sandbox (ACME
   retries against a nonexistent network). Follow the existing
   caddy-free VM pattern instead.

## e) Improvements (concrete, adopted from d)

- Stage explicit paths when any concurrent session is live; re-View
  before every Edit (already did this after two freshness trips).
- Train-start recon gains `nix build .#webphone.goModules` (the FOD
  freshness check) beside `git log origin/main..HEAD` + ci-verdict.
- Every new assert quotes the wire/site truth (read the handler, run
  the probe, then pin).
- VM tests follow the caddy-free pattern; derivations that need the
  vhost get explicit values in VM configs unless the vhost itself is
  the thing under test.
- The daemon interleave is handled by ONE narrative commit per task,
  but the daemon still fragments files it commits first — accept the
  fragmentation (content is identical), never rebase to clean it.

## f) Next 50 (the queue, sorted impact-first)

1. Debug the VM credential-mode assert (read the actual stat; pin the
   honest mode; systemd docs say LoadCredential copies are 0400 — find
   why the VM disagreed).
2. Decide caddy-free vs caddy-quiet for the VM config (explicit
   redirectUrl + rp; kill the ACME noise).
3. Green `webphone-oauth2`; strike TODO row 147 with the run evidence.
4. CHANGELOG line for the VM test (runtime proof) + the F45/F46 plan
   checkboxes.
5. `nix flake check` over the whole flake (KVM tests included on this
   host) — first full flake check of the session.
6. Full `nix develop -c go test -count=1 ./...` train-end run.
7. Island suite + `tsc -p . --noEmit` (untouched this half; the shell
   contract test still owes one run against the collapseToggle render).
8. `nix fmt` + `scripts/docs-gates.sh --check` train-end.
9. `timeout 90 bash scripts/ci-verdict.sh 0ca42edd` — poll to a final
   verdict (wait for pending to stop saying pending).
10. Attribute the `73054c22` red via check-runs API BEFORE claiming the
    restore fixed it.
11. If the red names MY surfaces: fix-forward; if it names the rail
    session's surfaces: their train (the §g-1 question below).
12. T18 F83: stage the quiet-window overnight `--track` run command (the
    night-run stability pair's first half).
13. T18 F84: define "stable" — pass count + duration band — for the
    night-run stability pair.
14. T18 F85: wire the ring-guard positive path (a real ring during
    `--track`, the quiet-window harness's ring arm).
15. T18 F86: run the ring-guard positive path once; record the evidence
    in the capture-harness docs.
16. T18 F87: strike/update TODO rows 221+240 (the night-run stability
    pair ledger rows) with the run evidence.
17. T17 F79: re-derive the package import graph (`go list`) for the
    architecture-understanding diagrams.
18. T17 F80: update the d2 sources (two dated generations, the
    38-package reality).
19. T17 F81: render the architecture-understanding SVGs; keep the old
    files (the dated-snapshot convention).
20. T17 F82: strike TODO row 1668 with the new file paths.
21. File a tracker issue for the 13:39 placeholder stomp IF §g-1 says
    daemon defect (else: adopt the ritual, write it into AGENTS).
22. Sweep TODO_LIST for the strike-evidence rows T01–T11 (all eleven
    plan rows now have evidence; verify each strike carries commit shas).
23. `python3 scripts/status-index.py` regen after THIS report lands.
24. Watch-calendar check: api-dial tripwire due 2026-11-17 (two days).
25. erraudit tier re-measure 2026-12-08 entry — verify still armed.
26. The F58 drain (annotate pin 260's way down) — the ★priority call
    before ~2026-12-07 cadence window.
27. Consider the T15 residue: the six small strike commits worth of
    Errata rows — verify none regressed during this session's edits.
28. The rail session's CHANGELOG sidebar-rail entry + my entries: check
    CHANGELOG ordering survived the daemon merges.
29. Verify the collapseToggle extraction against ui-capture (the T23
    harness pins header shots; the extraction changed the source, not
    the bytes — one gallery run would prove it visually).
30. `scripts/bump-preflight.sh` — no bumps this session; confirm none
    of my commits touched go.mod/go.sum (the commit-msg guard holds for
    mine; the daemon's e85830ef touched only packages.nix).
31. The `push.draft_*` code names in AGENTS.md ("push.draft_devices/
    push.draft_send wraps") are STALE prose (the codes don't exist) —
    an errata strike candidate.
32. The push-lane prose in docs/dedup-registry.md's 2026-11-13 row
    references — re-check the registry sweep-log has the 111-group
    state recorded (done) and the check stays 0-new at train end.
33. `buildflow` full run at train end (the battery; BUILDFLOW_NO_RESULT_CACHE=1).
34. `nix run .#vulnix` — release-adjacent, cheap, unrun this session.
35. `scripts/docs-gates.sh` mutating run (fmt first) if the train-end
    --check flags anything.
36. Consider filing the BuildFlow coverpkg ask (§ owner tick in the
    draft) — pending §g-2's sibling answer on filing etiquette.
37. The webphone-backup VM test + webphone-error-drill: re-run both
    inside the train-end `nix flake check` (KVM present) — neither has
    run since my module edits.
38. Read the backup VM test's two EMPTY machine.succeed() calls (lines
    with bare parens in nix/vm-tests.nix — they look like daemon-mangled
    corpses; fix or remove).
39. `webphone-backup-drill` (sandboxed): run in the battery.
40. Smoke the released-shape binary with --expect-version at the NEXT
    release (not now — no release this train).
41. The agent scenario smoke: consider adding a textagent-on notes
    interaction probe (cheap, rides the existing scenario).
42. Consider a `cqrs.note_missing` family_test literal pin (the
    self-test's `cqrs.note_list_full` pin exists; `note_missing` has no
    family_test.go home — its registry row cites callnotes.go only).
43. CHANGELOG ordering: my T06/T11 lines above the rail session's
    Changed entries — verify dprint kept the grouping stable.
44. AGENTS.md § Commands: the smoke line now lists "agent scenario" —
    one re-read at train end for prose drift against the 9 verdict lines.
45. The interim report's §g owner questions (3) from 12:20 — still
    unanswered; fold into this report's §g so the sitting pack is ONE
    surface.
46. architecture-visualization skill: T17 says "regenerate via the
    skill" — load the skill BEFORE touching the d2 sources.
47. The `docs/review/libs/` fleet-audit currency: 2026-11-13 re-verify
    is fresh; nothing owed until the next upstream tag sweep.
48. `scripts/push-stall-watch.sh` — run before invoking the manual-push
    bar next time (state the anchor).
49. `TODO_LIST` appends this session made (D2/D3 design paragraphs,
    pbx-lens row, appkit PARK row): re-read for strike-canon compliance
    (evidence-bearing, dated).
50. Sleep the cadence candidates list until the next docs-health sweep
    (the 65/45d ratchets own the due date; today's runs were green).

## g) 3 questions for the owner (the sitting pack)

1. **The 13:39 placeholder stomp** of nix/packages.nix (`e85830ef`):
   was that YOUR deliberate hash-reset ritual (the `.buildflow.yml`
   nix-hash-fix recipe writes placeholders before filling them), a
   concurrent session's in-flight state I wrongly swept with
   `git add -A`, or a daemon defect I should file a tracker issue for?
   I restored the real hashes (verified by the FOD `got:` match) — the
   question decides whether I also write the "stage explicit paths
   under concurrent sessions" lesson into AGENTS.md § Concurrent
   sessions as a RULE (I already behave that way now).
2. **The OIDC VM test's "app reads it" bar**: does the current shape
   (credential lands 0400 + rendered config names the credential path +
   begin 302s to the issuer + the fail-closed restart arm) satisfy
   F43's "assert the app reads it at login", or do you want the FULL
   token-exchange dance (a mock IdP serving discovery/token/jwks in the
   VM, ~80 more lines of test)? The lazy-discovery design means the
   secret is only read at the token exchange, so the full dance is the
   only way to prove the secret's bytes reach the app's OIDC client.
3. **VM-test caddy posture**: kill the ACME noise by making the OIDC VM
   caddy-free (explicit `redirectUrl` + hand-written `rp_id`/`rp_origins`
   — loses the derivation coverage in the VM, keeps it in the eval
   checks), or keep caddy.enable and just never assert on caddy (the
   ACME retries stay as harmless log noise)? My lean: caddy-free VM +
   the derivations stay covered by module-output-oauth2.golden.

## Verification ledger (what was actually run, per task)

- T04: `go test ./internal/server/...` (21.9 s green), push/readmodels/
  cqrs suites green, `error-context.py --self-test` 67/67.
- T05: full smoke suite TWICE (57/1 first — my wrong assert; then
  **58 passed, 0 failed** with all 8 scenario verdict lines).
- T06/T10/T11: `nix build .#checks.x86_64-linux.webphone-module` green
  (35 cases incl. the 4 textagent + 3 passkey + oauth2 golden).
- T08: the union run itself (all packages `ok`), `go tool cover -func`
  total 75.8%.
- T09: VM boots (24.4 s), fails at the 0400 assert — OPEN.
- Docs: `docs-gates.sh --check` green after every docs-bearing commit
  (T05/T08/T10/T11).
- Train end: NOT yet run (see b).
