# Status report — OIDC / SSO: typed NixOS lane

**Snapshot:** 2026-10-25 15:16 CET (Sunday) · repo `~/projects/webphone`, branch `main`
**Scope:** the session that answered "Is it perfectly configured for NixOS?" and then
made the OIDC/SSO deployment lane perfect: typed options, systemd credentials,
eval-time fail-closed parity, tests, docs.
**Claims verified as of 15:16:** check verdicts, commit SHAs, remote head, live-boot
probe. Push/CI statements rot fastest — see §Facts.

---

## 0. What the session set out to do

Three gaps were named at the end of the previous answer:

1. Pocket ID client creation and secret-file perms live in the hosts repo.
2. `package/options.nix` had no typed OAuth block → `nix flake check` could not
   catch a malformed `settings.auth.oauth2`.
3. `client_secret_file` is a static absolute path only → no systemd
   `LoadCredential` provisioning.

Gap 1 is an IdP-provisioner concern and cannot be closed in this repo (the module
now _consumes_ the registration with zero host-side secret plumbing — that half is
closed). Gaps 2 and 3 are closed.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                  | Evidence                                                                                                                                                                                                                                 |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Typed `services.webphone.oauth2.{enable,issuer,clientId,clientSecretFile,redirectUrl,scopes,allowedGroups}`                                                                                                           | new `package/options-oauth2.nix`, imported by `package/nixos-module.nix`                                                                                                                                                                 |
| 2  | Secret provisioning through systemd `LoadCredential` (no host-side `640 root:webphone`)                                                                                                                               | `LoadCredential = [ "webphone-oidc-client-secret:<host path>" ]`; app reads `/run/credentials/webphone.service/webphone-oidc-client-secret`                                                                                              |
| 3  | Fail-fast secret: a missing source file fails the UNIT START, not the first login                                                                                                                                     | consequence of `LoadCredential` (systemd refuses to start)                                                                                                                                                                               |
| 4  | Callback derived from the vhost (`https://<caddy.hostName>/api/auth/oauth2/callback`) so vhost and IdP registration cannot drift                                                                                      | eval verified: `redirect_url == https://phone.example.org/...`                                                                                                                                                                           |
| 5  | WebAuthn prerequisites derived (`rp_id`/`rp_origins` from `caddy.hostName`, `mkDefault` so operator pins win) — enabling OIDC derives passkey mode from the shared users map, which the app otherwise rejects at boot | eval verified: `rp_id=phone.example.org`, `rp_origins=["https://phone.example.org"]`                                                                                                                                                     |
| 6  | Eval-time fail-closed mirror of the app's boot validation                                                                                                                                                             | 8 half-shapes each trip a named assertion: issuer (unset/http), secret, `openid` scope, underivable callback, missing `auth.users`, missing per-extension password file, knob-without-`enable`                                           |
| 7  | One home per fact: typed + raw `settings.auth.oauth2` for the same key = eval CONFLICT (not a silent winner)                                                                                                          | `oauth2-typed-raw-conflict` case via `builtins.tryEval`                                                                                                                                                                                  |
| 8  | Raw lane untouched when the typed lane is off (the consuming stack's shape)                                                                                                                                           | `oauth2-raw-lane` case: settings byte-identical, no `LoadCredential`, no derived rp                                                                                                                                                      |
| 9  | Store-path secret guard: a `/nix/store` string **or** a Nix path value in `auth.oauth2.client_secret_file` / `extension_password_files` fails the eval (path values are copied into the world-readable store)         | `oauth2-secret-store-guard` case; `builtins.isPath` avoids coercing (and thus copying) the offending value itself                                                                                                                        |
| 10 | Module checks                                                                                                                                                                                                         | new `nix/module-check-oauth2.nix` (6 cases), wired into `checks.webphone-module`; green                                                                                                                                                  |
| 11 | Golden fixture unchanged → zero config keeps the generated settings byte-identical                                                                                                                                    | md5 `125d72a26f8b242ac990016f7cc35907` (HEAD vs origin/main)                                                                                                                                                                             |
| 12 | Real-binary boot proof from the module-rendered config                                                                                                                                                                | `/healthz` 200 (`sqlite`, `blob-dir`, `userauth` ok), log `passkey auth mode=enabled rpid=phone.example.org users=1`                                                                                                                     |
| 13 | SSO route live and honest on IdP failure                                                                                                                                                                              | `GET /api/auth/oauth2/begin` → `303 Location: /?sso_error=unavailable`; login card carries `sso-login` + `/api/auth/oauth2/begin`                                                                                                        |
| 14 | Docs                                                                                                                                                                                                                  | README (config table +8 rows, module table +7 rows, new "OIDC / SSO sign-in" section with the LoadCredential + rotation contract), AGENTS module-options rule, CHANGELOG `[Unreleased]`, `docs/lessons.md` (recursion + priority lesson) |
| 15 | Formatting + eval gates                                                                                                                                                                                               | `nix fmt` clean; `nix build .#checks.x86_64-linux.{webphone-module,format}` green; `nix flake check --no-build` → "all checks passed"                                                                                                    |
| 16 | Working tree clean; all changes committed (daemon commits)                                                                                                                                                            | `git status` empty; local `main` 5 commits ahead of `origin/main` (`3768360`)                                                                                                                                                            |

## b) PARTIALLY DONE

| # | Item                                              | What is missing                                                                                                                                                                                                                 |
| - | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | `LoadCredential` verification                     | Only proven at EVAL level. No NixOS VM/systemd test asserts the credential directory exists, the file is mode 0400/owned by `webphone`, `ProtectSystem=strict` does not block the read, and the app actually reads it at login. |
| 2 | Live boot smoke                                   | Run manually (render JSON via `nix eval` → boot `/nix/store/kak026…/bin/webphone` → probe). Not codified as a check, so it cannot regress-guard.                                                                                |
| 3 | Raw-lane boot safety                              | The module asserts prerequisites only for the TYPED lane. A passkey-only raw `settings.auth.users` config behind `caddy.enable` still gets no derived rp and is left to the app's boot validation.                              |
| 4 | Module checker coverage of the typed/raw contract | Conflict, raw-lane, external-front and store-guard are pinned; a second golden with `oauth2.enable` (exact JSON) is not.                                                                                                        |
| 5 | Stack integration                                 | Not touched (separate repo). See §d/§e — the stack's oauth2-only lane is latently broken.                                                                                                                                       |
| 6 | FEATURES.md / TODO_LIST.md                        | Not updated (CHANGELOG carries the change; the follow-ups from §f are not filed anywhere durable).                                                                                                                              |
| 7 | CI verification                                   | Not possible yet: the daemon owns the push and the 5 local commits are unpushed; no `gh run list` verdict exists for them.                                                                                                      |

## c) NOT STARTED

1. NixOS VM test for the OIDC lane (`checks.x86_64-linux.webphone-oauth2`-style).
2. Full `nix flake check` with builds (package, island tests, KVM-gated backup/VM tests).
3. `buildflow` (full quality gate) and the `docs/agents-md-size` / markdown preflights.
4. `scripts/whitespace-drift.sh` after staging.
5. CI watch after the daemon push.
6. Fix of the stack's oauth2-only boot failure (other repo).
7. Migration of the stack from raw `settings.auth.oauth2` to the typed options.
8. Typed passkey options in this module (so passkey-only configs also get derived rp).
9. A README Nix snippet for the typed lane (prose + tables only today).
10. Rotation runbook note in `docs/release-runbook.md` (README/option docs carry it).

## d) TOTALLY FUCKED UP

Nothing shipped broken. Honest near-misses, in order of severity:

1. **Infinite recursion (caught before any commit).** The first wiring derived
   `rp_id` whenever the shared `auth.users` map was non-empty, i.e. it read
   `cfg.settings` from inside a definition of `settings`. Eval died with
   "infinite recursion encountered while evaluating `services.webphone.settings`".
   Fixed by keying the derivation on the typed lane only; the trap is now in
   `docs/lessons.md` because it is easy to re-introduce.
2. **A false doc claim removed before commit.** The first draft said the secret
   "rotates without restart" (true for the app's per-login file read, FALSE under
   `LoadCredential`, where systemd snapshots at unit start). Corrected in the
   option description, README, AGENTS and CHANGELOG scope.
3. **The daemon committed mid-flight**, bundling my in-flight nix files with another
   session's Go changes (`internal/server/api/*.go`) into heuristic `chore:`
   commits (`7ba252e`, `7555992`, `c75790f`, `152327b`, `b882788`). Content is
   correct; the commit narrative is not. This is the known daemon behaviour, not a
   data loss, but it is why the tree is "clean" without a narrative commit.
4. **The consuming stack has a latent boot failure I did not fix**
   (`~/projects/nix-international-telephony`): with `oauth2.enable` and
   `passkey.enable = false`, `modules/telephony/web.nix` writes `auth.users` but no
   `auth.passkey.rp_id`; the app derives passkey mode from `auth.users` and
   rejects the boot ("rp_id is required when the passkey mode is configured").
   Reported, out of repo, unresolved.

## e) WHAT WE SHOULD IMPROVE (process and design)

1. **Never claim a security property from eval evidence alone.** `LoadCredential`
   is a systemd runtime behaviour; the check proves the wiring, the boot smoke
   proves config validity, and neither proves the unit actually gets a 0400 copy.
   A VM test is the honest gate.
2. **Codify manual smokes.** The render→boot→probe proof is cheap (~1 min) and
   caught nothing new, but it would catch a regression the stand-in harness cannot
   (the stand-in accepts any `systemd.services` shape).
3. **Design for the `settings` freeform boundary earlier.** The module both reads
   and writes `settings`; the recursion class should be a checklist item before
   wiring any derivation. Rule to add: _assertions may read merged settings;
   definitions may not._ (Already in lessons.)
4. **Run the cage at train boundaries, not just the relevant check.** Only
   `webphone-module` + `format` + flake eval ran; full `nix flake check` builds and
   `buildflow` did not. The repo doctrine says do it at train start AND end.
5. **One-home docs discipline for derived constants.** The credentials path string
   exists in the module, the README, the option docs and AGENTS. It cannot drift
   silently (checks pin the module), but four copies is two too many.
6. **Cross-repo findings deserve a filed artifact** (issue/TODO), not just a chat
   sentence; the stack bug will otherwise be forgotten.

## f) UP TO 50 THINGS TO DO NEXT

**Verification / gates**

1. Add a NixOS VM test: module + fake secret file + stub IdP → unit starts, `/run/credentials/webphone.service/webphone-oidc-client-secret` exists, mode 0400, owner `webphone`, app boots, `/healthz` 200.
2. Same VM test asserts `ProtectSystem=strict` does not block reading the credential.
3. Codify the render→boot→`/healthz` smoke as a flake check (ephemeral port, temp data dir).
4. Assert in that check that the rendered `client_secret_file` is the credentials path (not the host path).
5. Run full `nix flake check` (builds + KVM-gated) and record the verdict in this report's successor.
6. Run `buildflow` (or `scripts/buildflow.sh`) and record exit/skip inventory.
7. Run `scripts/whitespace-drift.sh` right after staging on every future train.
8. Watch CI (`gh run list`) once the daemon pushes the 5 local commits; confirm green.
9. Add a second golden fixture for the oauth2-enabled settings JSON (exact JSON pin).
10. Add an eval case proving `LoadCredential` is absent with `oauth2.enable = false` even when raw `settings.auth.oauth2` is set (currently implied by the raw-lane case).

**Option surface**
11. Add typed passkey options (`services.webphone.passkey.{rpId,rpOrigins}`) so passkey-only configs get derived rp without raw settings.
12. Or assert `settings.auth.users` non-empty ⇒ rp resolvable, for the raw lane too (boot-safety at eval).
13. Assert mapping extensions lists are non-empty (eval mirror of the app).
14. Assert mapping keys look like emails (eval mirror; app still authoritative).
15. Decide whether `clientId`/`scopes` set without `enable` should trip the orphan assertion (today they do not, because they have defaults).
16. Add an option for secret delivery mode (`LoadCredential` default; host-path/credentialless opt-out for live rotation).
17. Add `LoadCredentialEncrypted` support (systemd-creds) as an alternative.
18. Validate that `clientSecretFile` is absolute when provided.
19. Add an eval case for `caddy.hostName` carrying a port (documents the rp_id host-mismatch trap).
20. Add an eval case proving a raw `settings.auth.passkey.rp_id` pin beats the derived default.
21. Unit-test `bareHTTPS` (uppercase scheme, userinfo, trailing slash, query, fragment, empty).

**Docs / ops**
22. Add a copy-paste Nix snippet for the typed lane to the README OIDC section.
23. Add a raw-lane (external front) snippet too.
24. Add the rotation runbook line (`systemctl restart webphone`) to `docs/release-runbook.md`.
25. Note the outbound IdP dependency in the ops docs (discovery is a live outbound HTTPS call from a hardened unit).
26. Update `FEATURES.md` with the typed OIDC lane.
27. File the §f follow-ups (at least items 1-3, 11, 22, 30) in `TODO_LIST.md`.
28. Add a CHANGELOG sub-bullet for the store-path secret guard.
29. Consider a docs/error-contract row? No new codes — explicitly record that decision instead.
30. Document the Pocket ID registration recipe (redirect URL, client type, secret rotation) once the owner confirms the exact pocket-id UI steps.

**Cross-repo (stack)**
31. Fix the stack's oauth2-only lane: derive `rp_id`/`rp_origins` from `cfg.domain` when `oauth2.enable` (mirroring this module), or assert `oauth2.enable ⇒ passkey.enable`.
32. Add a stack `telephony-eval` case for that shape (so it fails eval, not a boot).
33. Migrate the stack's `web.nix` to the typed `services.webphone.oauth2.*` lane and delete the raw settings write.
34. Teach the stack's secret-perms heal unit that `LoadCredential` makes the `640 root:webphone` grant optional for the OIDC secret (it is still needed for per-extension SIP password files).
35. Re-pin the stack to this webphone rev after the fix (release ritual; stack tree clean, narHash).

**Product / longer horizon**
36. Decide whether logout should propagate to the IdP (session-truth doctrine says no; document the silent re-login consequence in the README OIDC section).
37. Consider mapping IdP groups to per-extension access (today: one allowlist, all mappings) if the owner wants multi-tenant phone use.
38. Consider exposing SSO state in Settings ("signed in via Pocket ID") — currently no UI affordance.
39. Consider a read-only `/internal/oidc/status` machine endpoint (loopback) exposing discovery cache age — only if ops asks.
40. Consider a `sso.discovery` health leg (today `/healthz` has no IdP leg; an outage is deliberately login-only).

**Hardening / robustness**
41. Consider `LoadCredential` for the SIP per-extension password files too (same class of secret; today host paths).
42. Consider `oauth2.clientId` uniqueness/per-host guard if multiple webphone instances share an IdP client.
43. Re-check the 10 s discovery timeout against a slow-but-alive IdP (env-tunable?).
44. Consider caching negative discovery results briefly (today: none, per login attempt).
45. Verify the app's `sso.client_secret.read` journal message under `LoadCredential` (it still mentions the 640 grant; wording could name the credentials dir).
46. Consider pinning the credentials directory constant in one shared Nix binding (`package/` constant) and referencing it from docs.

**Process**
47. Add "definitions must not read merged `settings`" to the module authoring checklist in AGENTS or lessons (currently only in lessons).
48. Establish a monthly gate re-measure reminder (already in AGENTS; verify the cron/manual habit).
49. Record the OIDC train in the next release note (`webphoneVersion` bump + stack pin).
50. Keep a "one home per fact" audit of the four copies of the credential path (module, README, option doc, AGENTS) at the next docs-health pass.

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

1. **Stack policy:** should the consuming stack (a) derive `rp_id`/`rp_origins` when
   `oauth2.enable` (like this module now does), (b) require `passkey.enable`
   alongside `oauth2.enable`, or (c) migrate to the typed `services.webphone.oauth2`
   lane? (a) and (c) are the cleanest; I cannot choose a cross-repo direction.
2. **Secret rotation semantics:** is "rotate ⇒ `systemctl restart webphone`"
   acceptable, or do you want a `LoadCredential` opt-out (host-path mode) because
   live rotation without restart matters to you?
3. **Ownership of the push/CI step:** the 5 local commits are unpushed and the
   daemon's push leg is unobservable. Should I push manually now (invoking the
   manual-push bar would need the oldest unpushed commit's timestamp check) and
   watch CI to green, or leave it entirely to the daemon?

---

## Facts for the next session (avoid re-deriving)

- Remote head at snapshot: `3768360` (`origin/main`), local `main` at `b882788`
  (5 commits ahead, daemon-owned), working tree clean.
- Files added: `package/options-oauth2.nix`, `nix/module-check-oauth2.nix`,
  `docs/status/2026-10-25_15-16_oidc-typed-nixos-lane.md`.
- Files edited: `package/nixos-module.nix`, `nix/module-check.nix`, `README.md`,
  `AGENTS.md`, `CHANGELOG.md`, `docs/lessons.md`.
- Credential name/path used: `webphone-oidc-client-secret`,
  `/run/credentials/webphone.service/webphone-oidc-client-secret`.
- Derivation rule to remember: `deriveWebAuthnRP = cfg.caddy.enable && cfg.oauth2.enable`
  (keyed on typed only, because reading `settings` from a `settings` definition recurses).
- Precedence rules pinned: typed OIDC keys are PLAIN (conflict with raw), derived
  rp values are `mkDefault` (raw wins), csrf keeps its existing mkForce semantics.
- Test command: `nix build --no-link .#checks.x86_64-linux.webphone-module`
  (`.#checks.x86_64-linux.format` for the formatter, `nix flake check --no-build`
  for whole-flake eval).
- Boot smoke recipe: render `base.moduleSet { oauth2.enable = true; … }` settings
  to JSON, point `WEBPHONE_CONFIG` at it, run the `.#webphone` binary, probe
  `/healthz`; the SSO begin route answers `303 /?sso_error=unavailable` when the
  IdP is unreachable (expected, honest).
