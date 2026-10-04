# Status — Passkey users for the webphone (usermgmt train), 2026-10-04 12:39

Scope: the "BUILD IT" follow-up to the ugly-extensions question — passkey
(WebAuthn) login via an embedded `cqrs-htmx/usermgmt` identity layer, smart
numbers in the UI, break-glass extension login retained. Plan doc:
`docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md` (P/F task ids used
below). Primary repo: **webphone** (this repo); deployment wiring lands in
pbx-artmann; the stack (nix-international-telephony) needs only a lock bump
(settings are freeform JSON — verified).

---

## a) FULLY DONE

1. **Research spike (complete, load-bearing).** usermgmt embedding surface
   pinned: `NewService(ServiceConfig{EventStore, ReadModelDB, WebAuthn})` +
   `NewSQLEventStore(ctx, db, "sqlite")` (self-migrating) +
   `OptimizeSQLiteDB` + exported `ReadModel().FindByEmail`; ceremony wire
   contract mirrors usermgmt's own AuthHandler (begin `{email}` →
   `{options, session_key}`; finish `?user_id=` + RAW ceremony body); NO
   email-verification gate on login (verified in source); both repos PUBLIC
   (plain Go module dep, no private-fetch Nix plumbing).
2. **Plan doc** with Pareto 1/4/20/80, decisions D1–D9, P1–P12 + F1–F34
   task tables, mermaid graph, risks/mitigations, honest limits.
3. **F1 deps**: go.mod `usermgmt/v4 v4.13.1` + `usermgmt/webauthn/v4
   v4.12.0`; `go mod tidy && go mod vendor` (vendor/ synced).
4. **P2 config** (`internal/config/config.go`): `Auth.Passkey` +
   `PasskeyUser` structs (koanf/json tags), derived `Enabled()`,
   `validatePasskey()` — all-or-nothing (CRM/Paperless doctrine),
   rp_id == every origin's host, normalized extension keys, a password
   file required for EVERY mapped extension, email-key sanity. Compiles.
5. **P1 userauth** (`internal/userauth/userauth.go`): `Service` over
   `<dataDir>/usermgmt.db` (WAL, MaxOpenConns 1, pragmas via
   OptimizeSQLiteDB), SQL event store + SQL read models + webauthn
   provider, `Shutdown()` (satisfies `do.ShutdownerWithError`, asserted in
   app.go), `BeginLogin`/`FinishLogin→MappedUser` (usermgmt session
   deliberately ignored), `Resolve`, `MappedByExtension`,
   `SIPPassword` (read at login, trimmed, empty = Rejection — the
   2026-10-01 empty-secret class), `Register` idempotent via
   `ReadModel().FindByEmail` on `ErrEmailExists`. Compiles.
6. **F5 enroll tokens** (`internal/userauth/enroll.go`):
   `wp_enroll_tokens` table (sha256-at-rest, 15 min TTL, opportunistic
   sweep), verify = burn, unknown/expired/used answer one identical
   Rejection (no token-state oracle). Compiles.
7. **F10 app wiring** (`internal/app/app.go`): conditional
   `ProvideNamed "userauth"` (30 s boot-timeout-bounded `New`), eager
   `InvokeNamed` fail-fast (half-a-login-mode must fail the boot),
   `Deps.UserAuth`, `passkeyRuntime` adapter (extensions parsed once at
   boot), shutdown rides the container cascade. Compiles.
8. **P3 passkey login surface (handlers + routes)**:
   `internal/server/session_api.go` refactored —
   `verifyDirectoryCredentials` (the ONE fail-closed directory check
   behind BOTH login modes), `mintSession` (ONE session birth path:
   create + cookie + CSRF rotation), `sessionIdentityResponse`
   (extension + did + display_name + numbers), `numbersFor` (dedup,
   mapping order); `getSession` now returns the display identity for the
   boot-resume whoami. `internal/server/passkey_api.go`: begin/finish
   login + enroll verify/begin/finish handlers — enumeration-safe 401
   (unknown email ≡ no passkey), 429 lockout, 503 rejection/transient
   (family+code logged). Routes in `server.go` registered ONLY when
   `deps.UserAuth != nil` (disabled deployments keep the styled 404);
   login ceremonies share `loginLimiter`, enrollment gets
   `passkeyLimiter`.
9. **Compile state**: `internal/config`, `internal/userauth`,
   `internal/app` GREEN. `internal/server` has exactly ONE remaining
   error — the not-yet-written `h.enrollPage` (P5 in flight, see b/1).

## b) PARTIALLY DONE

1. **P5 enroll surface**: routes registered + all five handlers written;
   MISSING: `views.EnrollPage` templ page, `h.enrollPage` handler, island
   `enroll.js`. This is the one known compile breaker in the tree.
2. **P12 git state**: 6 daemon auto-commits on `main` (ahead of origin,
   NOT pushed) mixing plan + code — the narrative phase commits the
   webphone AGENTS demands are still owed.
3. **Vendor/lock health**: go.mod/vendor consistent; the Nix `vendorHash`
   in `nix/packages.nix` NOT yet bumped (nix build of the package will
   fail until `buildflow -s nix-hash-fix --fix`).

## c) NOT STARTED

- **F14 island WebAuthn coercion + login UI**: `webauthn.js` (base64url
  buffer coercion both directions — highest-risk unproven part),
  `passkey.js`, `enroll.js`, `phone.templ` passkey section (email +
  button, break-glass `<details>` around the existing form), `main.js`
  wiring (finish response → `connect()` → whoami display_name+numbers),
  `configjs.go` passkey flag.
- **F17-CLI**: `webphone -enroll-passkey <email>` (Register-if-missing →
  mint → print URL from `rp_origins[0]`); `main.go` boot log line.
- **P7 tests**: userauth service tests (stub WebAuthnProvider —
  usermgmt's own strategy), token burn/expiry/uniformity, SIPPassword
  failure classes, server handler tests (off = 404, 401 uniformity,
  session mint, 429, CSRF), family pins for the new seam codes.
- **P8**: i18n en/de keys + parity/unused tests; dom-contract ids +
  served-page contract update.
- **P9 webphone docs**: FEATURES/CHANGELOG/README/AGENTS/error-contract/
  TODO_LIST harvest.
- **P11 gates**: templ generate, `nix fmt` (island/shell/css — buildflow
  skips island files), buildflow full, vendorHash, `nix flake check`,
  loopback smoke (off-mode byte-shape regression + /enroll 404).
- **P10 pbx-artmann wiring** (nothing exists yet):
  `services.webphone.settings.auth.passkey` values (rp_id
  `pbx.artmann.tech`, users `lars@artmann.tech → ["1000"]`, password
  file `/var/lib/telephony-secrets/telephony_ext_1000`),
  `secrets-perms.nix` grant (`telephony_ext_1000/1001` →
  `root:webphone 640`), enroll operator surface, AGENTS/CHANGELOG/
  DOMAIN_LANGUAGE updates, docs gates.
- **P12 train**: push webphone → stack `nix flake update webphone` + push
  (CI-green-first: `gh run list`) → pbx-artmann `nix flake update
  telephony` + `lock-drift-probe` (webphone pin must MATCH) → both-arch
  eval + toplevel build + `-o /tmp/pbx-toplevel-root` staging +
  `deploy-freshness` + hand over the switch command (owner-run).

## d) TOTALLY FUCKED UP!

1. **Worst self-caught hazard this session**: the first draft of
   `passkey_api.go` carried a placeholder `verifyPBXCredentials` stub
   returning `true` — had that shape survived to deploy, passkey login
   would have minted sessions WITHOUT directory verification (the exact
   forged-session class the check exists for). Caught and replaced by the
   real shared `verifyDirectoryCredentials` the same hour. RULE going
   forward: no placeholder security seams, not even in a first draft —
   the file must be honest at every pause.
2. **multiedit clobbered `probeRefreshIf`'s body** in app.go (supplied a
   new_string without the function body — a replacement that deletes code
   is a loss, not an edit). Caught by immediate re-view; restored.
3. **"Applied 1 of 2" misread**: in session_api.go one edit failed
   silently and I first concluded the WRONG one had applied. After any
   partial multiedit: re-view before reasoning about state.
4. **First-draft userauth.go** had trailing-comma syntax errors, a
   phantom `mustUserID`, and a dummy `UserIDByEmail` returning
   `("", false)` — a lying-name function (the naming anti-pattern in
   person). All removed; email→ID now rides the exported
   `ReadModel().FindByEmail`.
5. **Process miss**: dove into code before reporting the two demanded
   plan TABLE VIEWS in chat (they exist only in the plan doc), and worked
   in one big uncommitted blob instead of the webphone AGENTS' "small
   committed units, a narrative commit per phase boundary".

## e) WHAT WE SHOULD IMPROVE!

- Compile + narrative commit after EVERY P/F task batch, not per phase.
- Never a placeholder function again (see d/1) — write the real seam or
  nothing; "temporary" is a lie security audits cannot afford.
- Island WebAuthn coercion is the riskiest unproven component: write the
  base64url round-trip island test FIRST, then the ceremony code.
- Report the paste-ritual artifacts (tables) in chat before executing.
- Re-run `arch_test` early after adding a package (userauth must stay
  server/web-free; the config→runtime adapter lives in app.go only).

## f) Next tasks (impact-sorted; ids map to the plan doc)

| # | Task | Why next |
| --- | --- | --- |
| 1 | F15 `views/enroll.templ` + `h.enrollPage` | THE compile blocker |
| 2 | F14 island `webauthn.js` coercion + round-trip test | highest risk, unblocks all UI |
| 3 | F14 island `enroll.js` (verify→begin→ceremony→finish) | makes enrollment usable |
| 4 | F14 `phone.templ` passkey section + break-glass `<details>` | the visible feature |
| 5 | F14 `passkey.js` + `main.js` wiring + whoami name/numbers | the "smart numbers" win |
| 6 | F20 `configjs.go` passkey flag (+test) | island gates on it |
| 7 | `templ generate` + views render tests | gates green |
| 8 | F24 i18n en/de + parity tests | gate-pinned |
| 9 | F25 dom-contract ids + contract test | gate-pinned |
| 10 | F26 userauth service tests (stub provider) | proves register→enroll→login→map |
| 11 | F27 token burn/expiry/uniformity + SIPPassword failure classes | security pins |
| 12 | F28 server handler tests (off=404, 401 uniform, mint, 429, CSRF) | contract pins |
| 13 | F29 family pins for the new seam codes | tier-2 green |
| 14 | F17-CLI `-enroll-passkey` + main.go boot log line | operator onboarding |
| 15 | P9 webphone docs sweep + TODO_LIST harvest | docs-health duty |
| 16 | `nix fmt` (island/shell/css) BEFORE gates | known gate-killer |
| 17 | buildflow full (expect the ~54 vendor-consistency false positive) | quality gate |
| 18 | vendorHash via `buildflow -s nix-hash-fix --fix` | nix build unblocks |
| 19 | `nix flake check` (module-output.golden should be untouched) | sandbox gates |
| 20 | Loopback smoke: off-mode login page byte-shape + /enroll 404 | regression pin |
| 21 | Narrative phase commits, then push webphone `main` | train step 1 |
| 22 | P10 pbx settings.auth.passkey values | deployment truth |
| 23 | P10 secrets-perms grant `telephony_ext_1000/1001` → root:webphone 640 | login path needs it |
| 24 | P10 enroll operator surface (wrapper vs runbook — see g/1) | onboarding ergonomics |
| 25 | P10 pbx docs (AGENTS/CHANGELOG/TODO §1/DOMAIN_LANGUAGE) + docs gates | repo ritual |
| 26 | Stack: `nix flake update webphone` + push, CI-green-first check | train step 2 |
| 27 | pbx: `nix flake update telephony` + `lock-drift-probe` (pin MATCH) | train step 3 |
| 28 | pbx gates: both-arch eval + toplevel build + flake check --no-build | pre-staging |
| 29 | pbx: staging root `-o /tmp/pbx-toplevel-root` + `deploy-freshness` FRESH | pre-handover |
| 30 | `nix run .#deploy` composed guard + hand over switch command | owner runs it |
| 31 | Post-deploy: CLI enroll + browser ceremony + verify-live.sh | proof |
| 32 | Optional: per-train ui-capture login shots (LOCAL-ONLY budget) | polish |

## g) Top questions I can NOT figure out myself

1. **Enrollment operator surface**: a `webphone-enroll <email>` wrapper in
   `environment.systemPackages` (systemd-run as User=webphone,
   `WEBPHONE_CONFIG` lifted from the unit) — or runbook-only sudo
   one-liner? Wrapper = better UX + one more moving part; runbook = zero
   surface. Owner taste decides.
2. **Alice too?** Should `1001` get a mapped passkey user in v1, or
   Lars-only (Alice has no real person yet)?
3. **Train depth this session**: full ride (push webphone → stack lock →
   pbx relock → staged + gates) or stop after webphone-repo green for
   review? The paste says push; main is shared by the floating input.
