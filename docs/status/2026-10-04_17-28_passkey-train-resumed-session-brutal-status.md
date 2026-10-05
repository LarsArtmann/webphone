# Passkey train, resumed session — brutal status + self-review (2026-10-04 17:28)

Scope: the resumed "BUILD IT" session (12:39 report → now) across all three
repos: webphone (the product), nix-international-telephony (lock bump),
pbx-artmann (deployment wiring + staged closure). This report covers ONLY
this session's run and what I noticed in it. Lineage: plan
`docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`, prior snapshot
`docs/status/2026-10-04_12-39_passkey-users-train.md` (annotated today).

End state in one line: **the code train is DONE, pushed, and staged
(`0ngvm4q7…`); the deploy is one owner-run switch away; nothing is
customer-visible yet, and the installer release channel silently went
stale again (see d7).**

---

## a) FULLY DONE (verified in this session)

1. **F15 — the enroll page.** `views/enroll.templ` (standalone, no island
   runtime, CSP-clean head) + `h.enrollPage`; render-pinned by
   `TestEnrollPageRendersStandalone`; views i18n keys en/de; the
   `TestNoUnusedDictionaryKeys` gate stays green.
2. **`webauthn.js` — the wire coercion** + a 9-spec round-trip suite pinned
   against go-webauthn's `URLEncodedBase64` (unpadded base64url; padded
   spellings tolerated; null userHandle omitted, not empty-stringed).
3. **`passkey.js` — the island login module** (begin → get → finish →
   REGISTER), 6-spec suite; `session.js` refactored to export
   `adoptServerSession` (the ONE post-session sequence) — the 175
   pre-existing island tests stayed green through the refactor, all 184
   green at close.
4. **`enroll.js` — the standalone enrollment flow** (3-spec suite),
   placed in `assets/enroll/` OUTSIDE island/app so the modulepreload
   closure invariant (`TestIslandModulesMatchImportClosure`) stays
   honest — the invariant is documented in webphone AGENTS now.
5. **F18/F19 — the dual-shape login card.** Passkey front door +
   break-glass `<details>` when on; byte-shaped pre-passkey card when
   off; `TestLoginCardAdaptsToPasskeyMode` pins both shapes; dom-contract
   gained a "conditional ids" section.
6. **F24 — island i18n** (9 passkey + 6 enroll keys, en/de parity) and
   views-side enroll.* keys.
7. **F17 — the CLI.** `webphone -enroll-passkey <email>` (register
   idempotent → 15-min one-time token → prints the URL) + the boot log
   line; `app.PasskeyRuntime` exported for it.
8. **The provider seam.** `PasskeyRuntime.WebAuthn` injection (tests ride
   a deterministic stub mirroring usermgmt's own; production keeps
   go-webauthn) + the userauth suite (8 tests: register idempotency,
   token burn/expiry uniformity, unmapped fail-closed, all three
   password-file failure classes, mapping resolution).
9. **Server handler suite** (6 tests): mode-off styled 404s, the uniform
   anti-enumeration 401, finish-mints-session-with-identity, token burn,
   CSRF gate. The contract test's 401 allowlist grew its FOURTH
   documented writer (passkey_api) with rationale.
10. **A real bug the tests caught:** `passkeyFinishLogin` originally
    answered without the password — the island can't REGISTER without
    it. Fixed (no-store) before it could ship.
11. **Quality gates in webphone:** erraudit nolints + an honest shutdown
    log; `nix fmt` ×2; vendorHash via the repo's documented
    placeholder→got: roundtrip; `nix flake check` all-green; the
    error-code registry regenerated (the freshness gate itself caught
    the 17 new `userauth.*` codes — the machinery working as designed).
12. **go-licenses root-caused and fixed** (not skipped): the nixpkgs
    wrapper bakes `GOROOT=<go-1.26.8>` via the `go` callPackage arg; a
    dependency importing Go 1.27 stdlib (`crypto/mldsa` via
    go-webauthn) killed the license-check. Devshell now overrides BOTH
    `go` and `buildGoModule` to 1.27.
13. **Live loopback smoke, both modes:** off = `/enroll` 404 + untouched
    login card; on = passkey card + break-glass, `/enroll` 200, asset
    200, separate `usermgmt.db`; CLI mint → verify 200 → replay 503
    (burned) proven end-to-end against the running server.
14. **The smoke found a real bug:** `rp_id` validation compared
    host:port instead of hostname — spec-true fix (ports are
    RP-irrelevant) + the previously-missing `validatePasskey` suite
    (5 tests).
15. **Docs:** README config table + example, FEATURES row,
    error-contract rows, AGENTS invariant + runbook, CHANGELOG,
    TODO_LIST, pbx-artmann AGENTS/CHANGELOG/TODO §1/DOMAIN_LANGUAGE.
16. **The train:** sibling relocked+pushed (`890a526`, webphone
    `afa46ff→52f4d212`) with the ritual's local fallback evidence (104
    python tests + all-systems eval — CI runs keep dying to runner
    cancellations); pbx-artmann relocked, `lock-drift-probe` ALL-OK
    (webphone pins MATCH), staged `0ngvm4q7…` == flake eval, staged
    `webphone-config.json` byte-verified to carry `auth.passkey`
    alongside the stack-derived keys, both-arch eval + `nix flake
    check` + docs gates PASS. Webphone main pushed through `400eaff`
    (ls-remote verified).

## b) PARTIALLY DONE

1. **P12 deploy**: everything except the owner switch and the post-switch
   ritual (rotate `/tmp/pbx-toplevel-current`, record the fresh
   diff-closures baseline — the old `-current` target was GC'd, so the
   pre-switch diff-story leg is unreconstructable; stage-1 freshness was
   proven directly instead).
2. **Sibling CI verdict**: the run for `890a526` was still in_progress
   (~30 min) at session close; green unconfirmed (the fleet's runner
   cancellations strike at ~1 h). Local evidence stands as the fallback.
3. **`nix run .#deploy` composed guard**: never ran green end-to-end —
   secrets-preflight FAILS on the PRE-EXISTING
   `stalwart_relay_password` local-staging gap (mail train; every
   passkey-consumed secret is green). Components verified individually.
4. **BuildFlow full**: passing with warnings — the documented
   gomod-check vendor-consistency false positive (count honestly
   updated in AGENTS: 54→99 with the new deps).
5. **Live browser proof of the passkey itself**: BLOCKED on the switch
   (pbx-artmann TODO §1 row, with the fail-closed drill spelled out).

## c) NOT STARTED (noticed, owed)

1. **Plan-doc annotation** — `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`
   has ZERO struck items; the 12-39 status report is annotated but the
   plan's P/F task tables are not (docs-health: plans are snapshots too).
2. **TODO train-tail row went stale** — written at P9, never revisited
   after P12 completed it (fixing right after this report; see d8).
3. **`nix flake check --all-systems` on webphone** — only the plain x86
   check ran; the aarch64 build of the new pin rides the sibling CI's
   aarch64 lane (unconfirmed). Eval-only locally is the documented
   cheap floor.
4. **config family_test pins** for the new `config.auth.passkey.*`
   rejection codes (the codes are registry-fresh via erraudit-adjacent
   tooling, but config's own family pins don't cover them).
5. **Handler-level tests for enroll begin/finish** (service level is
   covered; the two thin handlers are not).
6. **verify-live.sh passkey probes** in pbx-artmann (`/enroll` 200/404
   gate, login-card passkey grep) — the no-ssh post-deploy suite
   doesn't know the new surface.
7. **ui-capture login shots** (optional polish, never claimed).
8. **`userauth` health**: the app's `/healthz` probes sqlite + blob
   only — the identity layer's own `usermgmt.db` is not health-probed.

## d) TOTALLY FUCKED UP (mistakes and near-misses — all caught, none shipped, honestly listed)

1. **`enrollFailed(0)`**: a network failure renders "Enrollment failed
   (HTTP 0)" — the status-0 sentinel leaks to the user, and the
   purpose-written `enrollNetFailed` copy is nearly-dead (only the
   outer catch uses it). Shipped wart; top of the fix list.
2. **Lost narrative commit messages**: the auto-commit daemon beat me on
   the sibling relock and several webphone batches — PUBLIC history now
   carries generic "chore: auto-commit" entries for a dependency train;
   the detail survives only in pbx-artmann's CHANGELOG.
3. **Partial multiedit ×2** (contract_test, i18n) — the known hazard;
   caught only because I re-viewed every time. The lesson stays: never
   trust "Applied 1 of 2".
4. **jq phantom readouts ×2** during the sibling relock: the tool shell
   misinvokes jq (a DOCUMENTED pbx-artmann gotcha I failed to apply
   proactively) and told me the lock hadn't moved while the file had —
   python/grep settled it. Nearly concluded a false state.
5. **First go-licenses override was wrong** (buildGoModule only): the
   wrapper still baked the 1.26.8 GOROOT. Should have read the nixpkgs
   wrapper FIRST; the second roundtrip found the `go` arg.
6. **Test-authoring sloppiness**: wrong hand-computed base64url vectors,
   a dropped email line, a racy intermediate-status assertion — three
   avoidable roundtrips (all caught by the suites; that's what they're
   for).
7. **FORGOT the installer release channel**: the relock makes the
   published `pbx-kexec-installer` release STALE again
   (`nix run .#release-freshness` would exit 1) — a documented,
   AGENTS-recorded gotcha I only remembered during this review. A
   recreate right now would boot pre-passkey code (recovered by step 5
   of the install flow, but still wrong).
8. **Self-inflicted doc rot within ONE session**: wrote the TODO
   train-tail row at P9, completed its contents by P12, left it
   claiming IN_PROGRESS at close.
9. **First pbx settings placement was wrong**:
   `services.telephony.webphone.settings` doesn't exist (eval error) —
   should have read the stack module's option surface before writing.
10. **`-current` loss discovered late**: the deploy-freshness FATAL
    surprised me mid-gate; the /tmp-root policy was in AGENTS all along
    — checking both roots BEFORE building would have made the handover
    cleaner.

## e) WHAT WE SHOULD IMPROVE (structural, from this session)

1. **Split brain — `csrfToken()` now has three island homes**
   (session.js module-internal, passkey.js private, enroll.js private).
   Deliberate at the time (session.js's is unexported); still drift
   bait. ONE exported home (or a tiny `csrf.js`) is the right shape.
2. **`whoamiLine` lives in passkey.js but serves the resume path** —
   identity display belongs in a neutral module (ui.js/session.js), not
   the passkey feature module.
3. **Typed responses**: finish/createSession answer `map[string]any` —
   a small typed struct would pin the wire shape at compile time.
4. **Testing gaps**: enroll begin/finish handlers, `userauth.Shutdown`,
   no browser-level passkey E2E (the stack's suite owns that surface —
   a passkey scenario belongs there post-deploy).
5. **CLI parsing**: hand-rolled `os.Args[1]` match; fine for one mode,
   won't survive a second flag.
6. **Docs machinery**: webphone lacks pbx-artmann's docs-annotate gate —
   the plan-doc rot (c1) is exactly what that gate catches. Porting it
   (or running pbx's binary ad-hoc) is cheap insurance.
7. **Commit cadence vs the daemon**: commit within minutes of each
   green step; a narrated train beats heuristic history in public repos.
8. **Lock reads in this shell**: jq lies — always python/grep (a
   one-line AGENTS rule already exists in pbx-artmann; consider a
   session-start reminder in webphone too).

## f) Next (impact-sorted; ids referenced nowhere else — this list is the harvest source)

| #  | Task                                                                                                                                                                                            | Repo     | Impact                        | Effort |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ----------------------------- | ------ |
| 1  | Owner switch (the handed-over `nixos-rebuild switch … --target-host root@pbx.artmann.tech`) + post-switch: rotate `-current`, record the diff-closures baseline, `verify-live.sh`               | pbx      | High (everything waits on it) | 10min  |
| 2  | LIVE passkey proof: enroll via CLI link, email+passkey login (whoami `Lars · +17287289311`), calls work; then the fail-closed drill (`: > telephony_ext_1000` → login fails honestly → restore) | pbx      | High                          | 10min  |
| 3  | `nix run .#release-freshness` + `./installer/publish-release.sh` — the installer channel is STALE after this relock (d7)                                                                        | pbx      | High (recreate readiness)     | 20min  |
| 4  | Confirm sibling CI verdict for `890a526` (re-dispatch if cancelled again)                                                                                                                       | stack    | Med                           | 2min   |
| 5  | Fix `enrollFailed(0)` → `enrollNetFailed` mapping in enroll.js + spec                                                                                                                           | webphone | Med (UX wart)                 | S      |
| 6  | Annotate the plan doc's P/F task tables (docs-health: plans are snapshots)                                                                                                                      | webphone | Med                           | S      |
| 7  | Fix the stale TODO train-tail row (harvest; doing post-report)                                                                                                                                  | webphone | Med                           | S      |
| 8  | csrfToken single exported home across island + enroll                                                                                                                                           | webphone | Med                           | S      |
| 9  | `userauth` health check in `/healthz` (usermgmt.db ping)                                                                                                                                        | webphone | Med                           | S      |
| 10 | config family_test pins for `config.auth.passkey.*`                                                                                                                                             | webphone | Low                           | S      |
| 11 | Enroll begin/finish handler tests + `Shutdown` test                                                                                                                                             | webphone | Low                           | S      |
| 12 | `whoamiLine` → neutral module (ui.js)                                                                                                                                                           | webphone | Low                           | S      |
| 13 | Typed session-identity response structs                                                                                                                                                         | webphone | Low                           | S      |
| 14 | `nix flake check --all-systems` on webphone (aarch64 eval)                                                                                                                                      | webphone | Low                           | S      |
| 15 | verify-live.sh: `/enroll` + login-card passkey probes                                                                                                                                           | pbx      | Med                           | S      |
| 16 | AGENTS.md size warn (411>377): move passkey detail to README                                                                                                                                    | webphone | Low                           | S      |
| 17 | Stack ops-runbook: webphone error-contract passkey rows (rides the pre-existing boot-contract patch row)                                                                                        | stack    | Low                           | S      |
| 18 | Browser-E2E passkey scenario in the stack suite (post-deploy)                                                                                                                                   | stack    | Med                           | M      |
| 19 | Consider stack-level `services.telephony.webphone.passkey.*` typed options (nicer than raw settings JSON)                                                                                       | stack    | Low                           | M      |
| 20 | Stage `stalwart_relay_password` locally (pre-existing mail gap; unblocks `#deploy`)                                                                                                             | pbx      | Med                           | S      |
| 21 | Ratify (or reject) the runbook-only enroll surface + Lars-only v1 mapping (see g)                                                                                                               | pbx      | Low                           | owner  |
| 22 | ui-capture login shots for the train record                                                                                                                                                     | webphone | Low                           | S      |
| 23 | Session-start lock-read guard: use python/grep, never jq (webphone AGENTS one-liner)                                                                                                            | webphone | Low                           | S      |
| 24 | Commit-faster rule for narrated trains (daemon race)                                                                                                                                            | process  | Low                           | S      |

(24 real items — padding to 50 would invent work; the ROADMAP-worthy
extras: multi-extension switcher, usermgmt identity endpoints beyond
passkey, both already recorded there.)

## g) Questions I can NOT figure out myself

1. **Switch now or CI-first?** The staged closure is locally proven
   (probe, both-arch eval, byte-verified config, gates), but the
   sibling's CI run for `890a526` was still in_progress at close and
   that fleet's runners cancel at ~1 h. Deploy on local evidence now,
   or wait for/force a green CI verdict first? Your risk call, not
   mine.
2. **Ratify the two autonomous decisions from the 12:39 §g?** (a)
   runbook-only enroll command — no NixOS wrapper script; (b) Lars-only
   mapping in v1 (Alice pre-granted but unmapped). Keep both, or do you
   want the wrapper / Alice mapped before the switch?
3. **Republish the kexec installer release now or after the switch
   verifies?** It's stale either way (d7); publishing now protects an
   emergency recreate, publishing after keeps the channel aligned to a
   PROVEN closure. (I can prepare the dry-run either way; the publish
   script is yours to run.)

---

_Written by the resumed session immediately after close; the assistant
now WAITS for instructions._
