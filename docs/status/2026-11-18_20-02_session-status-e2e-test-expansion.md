# Session status — SUPERB E2E + integration + workflow-frontend test expansion

> Claims verified as of 2026-11-18 20:02. CI run 44374290116 for pushed
> HEAD `157d3b5e` was still PENDING at snapshot time — push/CI statements
> below carry that caveat explicitly. Trigger: owner ask "a LOT more E2E
>
> - integration and fully automated workflow frontend tests". One home
>   for the plan + fine ledger: docs/planning/2026-11-18_18-21_SUPERB-e2e-workflow-test-expansion.md
>   (164-row task ledger, per-train strike notes with evidence).

## a) Fully done (verified green)

- **Plan doc** with Pareto breakdown (1%/4%/20% levers), mermaid
  execution graph, coarse train table + 164-row fine ledger, per-train
  landing evidence (all sections struck). dprint-canonical.
- **Smoke run_checks workflow block** (+29 checks, suite now
  `smoke: 92 passed, 0 failed`): csp-report intake (204 malformed /
  413 oversize — an oversize report must be VALID JSON SHAPE or the
  decoder errors before the cap), manifest + favicon, /api/export
  (anon 401 + real zip magic), snippets save/too-long/delete, contacts
  save/upsert/export/vCard-import/delete, composer send -> thread ->
  deep link -> bogus-id gone-notice, thread flags (pin/archive/unarchive
  /mute row-lane 204 vs head-variant 200 swap), MMS inbound with honest
  image/png attachment (anon 401, typed stream), fax send/document
  stream/inbound received row/no-resend-on-healthy.
- **Four new smoke scenarios** (+48 checks, all 0 failed):
  `msglifecycle` 17 (webhook gateway against a switchable fake provider:
  Bearer receipt, delivered verdict + badge, provider 422 refusal as
  non-retryable bubble, 5xx outage as retryable 502, retry lane, hook
  replay 202, SSE threads+thread), `ratelimit` 7 (login/compose/hooks/
  csrf limiter fences with Retry-After; login lane drives SUCCESSFUL
  logins — loopback never verifies credentials, and CSRF-class
  rejections die upstream of the limiter), `drafts` 11 (fake agent
  review API: ai-draft prefill, approve relay + drafts panel — the
  stack-side agent sends, not webphone — discard, mode, absent review,
  disposition), `push` 13 (fake ntfy: off-boot 404 + inert ring,
  write-only token, rotation, urgent ring wake on the rotated topic,
  end stand-down, post-deregister silence).
  Full smoke replay on a fresh binary: **200 checks, 0 failed** across
  all 12 scenarios.
- **Four Go workflow-chain tests** (new files only, no concurrent-session
  collisions): `internal/server/workflow_chains_test.go`
  (TestInboundWorkflowChain: hook -> BOTH SSE events with payload shapes
  -> unread 2 -> archive -> inbound AUTO-UNARCHIVES over HTTP;
  TestForeignOwnerSurfacesStayDenied: cross-owner attachment/fax-doc
  404s + foreign deep link -> gone-notice), `internal/server/voicemail_chain_test.go`
  (fake PBX mailbox under the /phone-api prefix the pbx client
  prepends; panel rows, delete proxies, payload-less voicemail nudge),
  `internal/app/boot_cycle_test.go` (composition-root cycle: front-door
  login -> CSRF rotation -> live SSE -> composer send -> inbound hook ->
  both nudges on one connection). Full `go test ./...` green for these.
- **Two island specs** for the real gaps: `notify.test.mjs` (title flash
  interval/idempotent-start/stop-restore/restoreTitle) and
  `shortcuts.test.mjs` (full keymap: letters/brackets/media act,
  unmapped pass through, modifier chords inert). Every other planned
  spec row was verified ALREADY PINNED by existing specs before writing
  duplicates (recorded per-row in the plan ledger).
- **Browser workflow harness** `scripts/workflow-e2e.py` (the NEW layer
  the owner named): real headless Chromium clicks the real UI over a
  self-booted binary — J1 login through the REAL front-door form (no
  cookie injection), J2 all six tabs (panel id + aria-current + URL +
  no-reload window marker), J3 composer send, J4 LIVE inbound via
  /hooks/message appearing through SSE->morph with no reload (stale-
  element-safe predicates), J5 pin, J6 snippet chip fill, J7 contacts
  add/delete through the real forms incl. the hx-confirm alert, J10
  Ctrl+K palette navigation, J9 styled 404, J8 wp-lang=de relabel (last
  — it flips the UI language). **`workflow e2e: 20 passed, 0 failed`,
  twice on fresh boots**; `--self-test` green; DOM-only (no screenshots).
- **CI wiring**: `workflow-e2e` job (build + the PYSEL store-path dance
  — nix >= 2.34 parses `nixpkgs#python312.withPackages(...)` as a PATH —
  - chromium/chromedriver, 15m budget). ADVISORY (continue-on-error)
    per the reproduce-first/tw-freshness precedent; flip is a TODO row.
- **Docs**: AGENTS.md Commands carries the four new smoke scenario names
  - the workflow-e2e run line; TODO_LIST.md gained the deferred-rows
    section (4 rows with reopen triggers); plan ledger closed.
- **Ship**: narrative commits `83e5c770` (smoke trains), `4a0d0c38`
  (Go chains + island specs), `157d3b5e` (ledger close); trains E/F/G
  rode daemon chore commits. Pushed (`84b2ab1e..157d3b5e`), remote
  verified via `git ls-remote`. ChangeLOG-silent (test-only, per the
  ratified bar).

## b) Partially done / in flight (verified as of 20:02)

- **CI verdict at pushed HEAD 157d3b5e**: run 44374290116 PENDING
  (~5 min elapsed at snapshot; typical 8-12 min). Prior run (44370689079,
  the concurrent session's 18:31 push) failed on a stale vendorHash pin
  AND 5 art-dupl groups — both predate my commits. The CURRENT tree's
  `.#webphone.goModules --rebuild` builds GREEN locally (rc=0), so the
  build jobs should pass at my HEAD; the outstanding risks are named in
  (d).
- **Full buildflow battery**: first run green on every step EXCEPT the
  go step carrying MY one flake (see (d)); the post-fix re-run was
  blocked by a HOST nix eval breakage (placeholder-hash FOD for
  templ-0.3.1100, the documented concurrent-eval class; flake.lock
  untouched by me) — not re-run since because the session was told to
  stop. Store-toolchain fallback covered the full go suite.

## c) Not started (deliberately deferred, rows + triggers in TODO_LIST)

- Admin + line-switcher BLACK-BOX scenario (mapped-login requirement;
  no black-box WebAuthn/RS256-OIDC lane from Python — plan B19-B27).
- Dial-gate chip browser journey (island SIP-boot bound — plan E29).
- Voicemail browser journey (needs fake PBX inside the browser boot).
- Flipping the workflow-e2e CI job to blocking (needs N clean trains).
- The 80/20 "other 20%" remainder beyond this session: perf/SLO lane,
  webphone-android digest-handshake half (foreign repo, pre-existing row).

## d) Totally fucked up / broken (precise, with blame-avoidance evidence)

1. **MY one flake, FIXED**: TestAppBootCycleLoginSendHookSSE raced its
   SSE subscription past the send under full-suite load (SSE events are
   not replayed). Fixed by gating on the hub's `connected` greeting
   before any traffic; 3x green standalone + app+server packages green
   under load after. Committed.
2. **Concurrent session's island reds (NOT mine)**: 2 postcall-note
   specs red (`/Saved/` assertion) since their 18:40 daemon commit of
   postcall.js + its spec. Attributed via isolation runs + git log;
   untouched per the fix-forward protocol. Blocks CI `fast-checks`
   (island-js) at my HEAD.
3. **Server depth-budget red (NOT mine)**: internal/server 5212
   hand-written LOC > 5200 pin — test files are excluded from the
   count, so the concurrent session's production admin/panels growth
   owns the bump-or-carve. Blocks CI `go-tests` at my HEAD.
4. **art-dupl: 3 clone groups at my HEAD (NOT mine)**: all in
   admin.templ/history.templ — the concurrent session's view edits
   (5 groups at their 18:31 run). Blocks CI `dup-gate`.
5. **Host nix eval breakage (environment, not repo)**: placeholder-hash
   FOD (`sha256-AAAA...`) for templ-0.3.1100 sources — devshell eval
   intermittently fails; flake.lock clean; quiet-window retry did not
   clear it within one bounded wait. CI is unaffected (fresh eval).
6. **docs-gates corpus reds (pre-existing)**: cadence + status-f-section
   - format over the corpus — verified identical WITHOUT my doc present
     (concurrent session's mid-flight TODO_LIST/FEATURES state).

## e) Improvements for having a working system

- The smoke's scenario set now covers every product write surface; the
  next regression in send/verdict/limits/drafts/push lands as a RED
  CHECK NAME, not a user report.
- The workflow harness gives the repo its first click-level UI test
  lane — future UI trains can extend it with one journey function
  (the stale-element-safe predicate pattern is the template).
- The `_scenario_boot` helper in the smoke makes future scenarios ~10
  lines of boot instead of 30.
- Honest negative findings recorded: quick-cap "5 snippets" is a UI
  fact not a server fence; admin/switch need mapped logins (test-seam
  ask rowed); pbx client's /phone-api prefix is now written down in a
  test (it cost this session 20 minutes).

## f) The next 50 things (ranked; rough effort, S/M/L)

**Land-clean-on-main cluster (do first)**

1. Watch CI 44374290116 to conclusion; record the verdict. (S)
2. Fix-forward the 2 postcall-note island reds (concurrent session's —
   coordinate or wait past bounded window, then file the tracker issue
   per protocol). (S)
3. Bump-or-carve the server depth budget 5212-> (concurrent session's
   production growth; the carve precedent is api/hooks/auth/platform).
   (M)
4. Baseline-or-fix the 3 art-dupl groups in admin.templ/history.templ
   (a deliberate `.art-dupl-baseline.json` regen with the PINNED v0.7.2
   detector, or de-dup the chips). (S/M)
5. Re-run the full buildflow battery once the host nix eval clears;
   record the posture claim with the sha. (S)
6. Flip workflow-e2e to blocking after N clean advisory trains (N=3?). (S)
7. Add a smoke scenario for /api/session/switch happy-path behind a
   mapped login once a test seam exists (pairs with row 33). (M)

**Browser-lane growth (the new layer's 80/20)**
8. Voicemail browser journey behind a `--phone-api` fake flag. (M)
9. Failed-send Retry/Dismiss buttons in the browser (webhook gateway +
switchable fake provider, reusing the smoke's _FakeProviderHandler
pattern in python). (M)
10. Thread archived-view toggle + back navigation in the browser. (S)
11. History transcript section + delete transcript chip in the browser. (M)
12. Settings surfaces: passkey list revoke, push device list, snippets
manage in the browser. (M)
13. Dial-gate chip journey (needs the SIP lane or a stack-side assert). (L)
14. Mobile viewport pass (390x844) over the J1-J10 journeys. (M)
15. Reduced-motion + dark-theme pass (assert no FOUC markers). (S)
16. Keyboard-only navigation pass (tab order + focus visible). (M)
17. Command palette: '?' help mode + every command entry. (S)
18. Login failure copy in the browser (wrong password -> #login-error
text). (S)
19. SSE reconnect in the browser: kill the server's hub mid-session,
assert the reconnect announce (needs a controllable seam). (L)
20. Optimistic send bubble + draft-restore-on-failure in the browser
(pairs with 9). (M)

**Smoke-lane growth**
21. Inbound fax with provider_ref + its status hook replay (idem) in
run_checks. (S)
22. /api/calls/{id}/review happy path with a fake agent (extend the
drafts scenario's fake with the review artifact wire). (M)
23. Transcripts: /api/transcripts POST + History partial section + trim
shape in a scenario. (M)
24. Dial slots listing GET /api/dials after the dialgate scenario's
allow. (S)
25. CSP report intake journal greppability (needs a log-tail seam or
assert via /metrics if a counter exists — verify first). (M)
26. /api/metrics/latency legs matrix (every fixed leg name folds). (S)
27. Enrollment burn-through: 3-token cap behavior in enroll_scenario. (M)
28. Boot-failure matrix: bad gateway URL, bad push credentials, bad
textagent secret-file (each refuses with its WHAT/FIX contract). (M)
29. Restart scenario: assert SSE reconnect + draft survival across the
SIGKILL reboot. (M)
30. Fronted-login scenario extension: full CSRF + Secure cookie attrs
on the configured boot. (S)

**Go-chain growth**
31. MMS attachment auto-transcribe flow at HTTP level (ASR stub +
transcript partial markers). (M)
32. Review artifact read model trim at HTTP level (2000 cap). (M)
33. Notes cap 409 at HTTP level (fill the cap, assert Conflict + copy). (M)
34. Agent-mode pill state transitions via /api/session agent_mode
fields (probe stub). (M)
35. chat-network compose: matrix-shaped address routes through the
sender registry (stub sender) end-to-end. (L)
36. Self-send rejection end-to-end with identities configured (422 +
persisted-but-refused). (S)
37. /phone-api proxy auth: Basic creds reach the PBX (assert in the
fake). (S)
38. Backup drill extension: system.db rows survive the restore (KVM
lane exists; add the row assert). (M)
39. OpenAPI golden: pin the new /api/metrics/latency + push routes if
missing (verify apidocs covers them). (S)
40. Contract freshness: `webphone contracts` regen check in CI (the
yaml is generated; a drift check would catch hand-edits). (S)

**Infrastructure / meta**
41. Expose a deterministic WebAuthn stub seam for black-box passkey
flows (unlocks rows 7 + the admin/switch scenario; owner call on
the seam's config shape). (L)
42. The smoke's `Check`/`Smoke` + workflow-e2e's `Check`/`Seed` are
near-identical: one shared `scripts/_harness.py` (hyphen-free,
importable) if a third consumer appears. (S)
43. ui-capture's `tap_through` could ride workflow-e2e's stale-safe
predicates (visual + workflow in one boot). (M)
44. Perf: island latency beacons asserted end-to-end in the browser
lane (needs the metrics env boot). (M)
45. Track smoke runtime: 12 scenarios now; watch the 15m CI budget,
split into two jobs if it approaches 10m. (S)
46. Add `workflow-e2e` to release.sh's pre-switch gates once blocking. (S)
47. Docs: FEATURES.md row for the workflow harness + the four scenarios
(test infra inventory). (S)
48. Lessons.md entry: "SSE subscriptions are not replayed — gate test
traffic on the hub's `connected` greeting" (cost this session one
battery run). (S)
49. Lessons.md entry: the pbx client prepends /phone-api to every
upstream path — fake PBXes must serve the prefix (cost 20 min). (S)
50. Sweep this status report's rows into TODO_LIST at the next
docs-health harvest (this report is the source). (S)

## g) Up to 3 questions I genuinely cannot figure out myself

1. **Who owns the current main red?** The three blockers at my pushed
   HEAD (postcall specs, depth budget, art-dupl groups) all originate
   in the concurrent session's in-flight admin/line-ownership train —
   per the fix-forward protocol I did not touch their files, but if
   that session has ended, someone must decide: do I fix them forward
   in a follow-up, or do they land their own remedy? (Protocol says
   tracker issue after a bounded wait; the wait window is still open.)
2. **Should the workflow-e2e CI job block now?** I landed it advisory
   per the house precedent, but the owner asked for "fully automated"
   — if your intent is "gating from day one", say so and I flip
   `continue-on-error` after the first clean run.
3. **Is a black-box passkey/SSO test seam wanted?** The admin surface
   and the line switcher are un-testable black-box without a mapped
   login, and neither WebAuthn ceremonies nor an RS256 OIDC IdP can be
   faked from stdlib Python. A config-gated deterministic WebAuthn
   stub (like the Go tests' PasskeyRuntime) would unlock the whole
   identity-layer E2E lane — that is an owner architecture call, not
   mine to make.
