# Status: loginpage adoption — build green, gates + docs + push remain

Claims verified as of 16:10 CET, 2026-10-25. Plan: `docs/planning/2026-10-25_15-53_SUPERB-loginpage-adoption.md`.
Prior snapshot: `docs/status/2026-10-25_15-49_loginpage-adoption-mid-flight.md`.

## Headline

The library login page is now WIRED, STYLED, and TESTED at the unit level:
`/login` serves behind a nonce CSP, the login.css artifact builds with the
app tokens inlined, all targeted tests are green (`TestLogin*`,
`TestPasskey*`, views). The unpushed daemon train (6 commits, origin still
`c7ca563`) is now SELF-CONSISTENT — go.mod/go.sum ride together with the
vendorHash re-pin, so a daemon push no longer red-s mains. What remains is
the verification tail (full gates, smoke, buildflow, flake check), the docs
wave, and MY narrative `auth:` commit + push (paste #8, still owed).

## a) FULLY DONE this segment

1. Plan doc with pareto tiers (1%/4%/20% + tail), medium table (M1–M13),
   fine table (F1–F32), mermaid execution graph, guardrails.
2. **M1 vendorHash re-pin**: `sha256-TkRNv/…` applied to nix/packages.nix;
   `nix build .#webphone.goModules` verified SUCCESS. Committed by daemon
   (in the unpushed train).
3. **M3 honest suite**: island i18n pruned (6 dead passkey keys × en/de);
   `templ generate` run (phone_templ.go refreshed); passkey_api_test.go
   moved to `/api/auth/webauthn/login/*` + `passkeyFinish` const +
   `GET /login` added to the off-404 list + SA4006 fixed (`resp`→`_`);
   `TestLoginCardAdaptsToPasskeyMode` rewritten for the link card (also
   pins the dead form ids stay dead). GREEN.
4. **M4 /login tests** (`login_page_test.go`): nonce-CSP lane (exactly
   `'self'` + one base64url nonce; every inline script carries it; no
   hashes/unsafe-inline), wire ids (lp-login-form, lp-email,
   loginpage-config, csrf meta, login.css link, begin endpoint), 303 for a
   live session (CheckRedirect stopped at first hop), styled 404 when the
   mode is off. GREEN (`go test ./internal/server/ -run 'TestLogin|TestPasskey'` ok;
   views ok).
5. **M5 DOM contract**: conditional section rewritten — `passkey-login`
   link on `/`; lp-* ids documented as library-owned contract on `/login`.
6. **M6 styling — WITH A PRE-EXISTING BUG FOUND**: `.wp-sso` (SSO link,
   shipped 2026-10-25) and `.passkey-breakglass` had NO styles anywhere —
   the SSO-era login card rendered unstyled anchors/disclosures. Added
   `.island a.primary` (full-width primary-key treatment, covers both the
   SSO link and the new passkey link) + break-glass disclosure styles to
   island/style.css.
7. **M2 login.css artifact**: `build-tw-css.sh` REWRITTEN — sources the
   pinned modules from the module cache via `go list -m` (fixes the dead
   `vendor/` dependency, untracked since 2026-10-17), stages the same
   vendor/-shaped layout (tw.css.input byte-stable), adds a @theme parity
   gate between inputs, builds BOTH artifacts, artifact gates, nix fmt.
   `login.css.input` written (@import tailwindcss + ./app.css tokens +
   `.dark`-extended variant + identical @theme + full @source set incl.
   loginpage page/js + feedback alert/loading[Spinner]). BUILD SUCCEEDED;
   login.css verified: app tokens present, `.min-h-dvh`, `.bg-blue-600`,
   `var(--accent)` ×60, `.hidden`, `.dark` ×177. Discovered en route:
   `icons` is a SEPARATE module (`templ-components/icons`), and Spinner
   lives in `feedback/loading.templ` (no spinner.templ).

## b) PARTIALLY DONE

1. **tw.css refresh**: rebuilding surfaced +15 utility lines
   (field-sizing-content, max-h-80, min-h-10, invert) NOT referenced by any
   views markup — provenance is the module-cache component sources vs the
   old vendor snapshot. ADDITIVE-ONLY (dead utilities, no existing rule
   changed) → benign, but the classification is grep-verified only; commit
   it with the refresh note. Currently UNCOMMITTED (tw.css modified,
   login.css untracked — the daemon will sweep).
2. Everything in § c is sequenced after this — no partials beyond that.

## c) NOT STARTED

1. **M7 full local gates**: island node:test suite, `go test -count=1
   ./...`, session suite (`-run TestServer`), scoped `nix fmt`.
2. **M8 smoke boots**: passkey-OFF parity (`/login` styled 404) +
   passkey-ON render (JSON config boot, curl /login for styled+nonce+lp-*).
3. **M9 buildflow** full run.
4. **M10 `nix flake check`** (sandboxed build rides the re-pinned hash).
5. **M11 docs wave**: CHANGELOG `[Unreleased]`, README § Passkey sign-in,
   AGENTS.md bullet, FEATURES.md, TODO_LIST row, release-runbook stack-E2E
   obligation note.
6. **M12 narrative `auth:` commit + push + CI watch** (user paste #8 —
   explicitly requested, still owed).
7. **M13 close-out**: plan checkbox sweep, this file's successor, TODO
   harvest.

## d) TOTALLY FUCKED UP / regressions this segment

1. None landed broken — but three tool-discipline misses wasted cycles:
   dom-contract.md edit rejected (read-guard), build-script write rejected
   (my own sed had touched mtime), one accidental no-op multiedit.
2. The icons-module/layout discovery cost one failed script run (`cp:
   cannot stat …/icons/icon.templ`) — checking the module layout BEFORE
   writing the script would have caught it.
3. Exposure note: between 15:53 and the re-pin commit, go.mod sat in a
   daemon commit WITHOUT the hash — the window closed clean (origin never
   moved), but the plan's "deps as one atomic unit" lesson stands.

## e) WHAT WE SHOULD IMPROVE

1. Verify the source tree's module layout (go list -m + ls) BEFORE
   writing build scripts against remembered paths.
2. The stale-artifact trap: committed tw.css was NOT reproducible from its
   own input anymore (a past views/module change never triggered a
   rebuild). Consider a CI freshness grep (artifact ⊇ referenced classes)
   — candidate TODO row.
3. Same read-before-edit discipline note as before; the guard keeps being
   right and I keep being fast.

## f) NEXT (ordered)

1. Commit tw.css refresh + login.css (with the benign-drift note).
2. M7: island tests → `go test -count=1 ./...` → session suite → nix fmt.
3. M8: two smoke boots (OFF parity, ON render).
4. M9 buildflow; 5. M10 flake check.
6. M11 docs wave (CHANGELOG/README/AGENTS/FEATURES/TODO_LIST/runbook).
7. M12: `git status` review → whitespace-drift check → narrative `auth:`
   commit (detailed) → push → `gh run list` to green → ls-remote end state.
8. M13 close-out (plan checkboxes, harvest CI-freshness candidate row).

## g) QUESTIONS (cannot figure out myself)

1. **tw.css refresh**: commit the +15-utility refresh (additive, benign,
   artifact now honest to its input) — or revert tw.css and keep only
   login.css, leaving the shell artifact stale-but-known? My call is
   commit; veto if you want the shell artifact frozen this train.
2. **Commit granularity for M12**: ONE narrative `auth:` commit for the
   whole adoption (tests+css+docs together), or split (server+routing /
   island+views / css+script / docs)? The daemon already interleaved
   chore: commits, so my vote: one final narrative commit.
3. **Push timing**: push immediately after my narrative commit lands
   (paste #8 asked for it), or hold for your explicit GO given the daemon
   may race its own push?
