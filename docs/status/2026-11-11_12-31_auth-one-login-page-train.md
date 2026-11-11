# Auth → one dedicated `/login` page (loginpage ExtraSection train)

> Claims verified as of 12:31 CET, 2026-11-11. Push/CI statements rot
> fastest; the timestamps below are per-claim.

## Goal

Make auth ONE dedicated page at `/login`, hosted by the cqrs-htmx
loginpage library, in every deployment mode — instead of the old
topology (extension+password form buried inside the SIP island card in
the app shell; `/login` 404ing whenever the passkey mode was off; two
competing doors when it was on).

## DONE (verified)

### Library: cqrs-htmx loginpage v4.18.0 — SHIPPED

- `loginpage.Config.ExtraSection templ.Component`: consumer-provided
  sign-in section rendered inside the library card below the
  WebAuthn/OAuth2 sections, divider-separated. Counts as a sign-in
  method everywhere the page counts methods (no "No authentication
  method" alert, no `New` warning, `AutoRedirect` suppressed).
- `Config.Service` is now OPTIONAL when `ExtraSection` is set (nil
  service with no section still rejected at `New`) — every service read
  (`HasWebAuthn`, `ConfiguredOAuth2Providers`) nil-guarded in `New`,
  `buildPageData`, and the page template's alert branch.
- Six new tests
  (`TestNew_NilServiceWithExtraSection`,
  `TestServeHTTP_ExtraSectionOnlyRenders`,
  `TestPage_ExtraSectionDividerFollowsOtherSections`,
  `TestPage_ExtraSectionAloneHasNoDivider`,
  `TestServeHTTP_ExtraSectionBlocksAutoRedirect`, + a
  `TestSoleOAuthBeginURL` table row). Full module suite green:
  `GOWORK=off GOTOOLCHAIN=auto go test -count=1 ./...` → ok (12:00).
- README (new § Consumer sign-in sections + Configuration row) and
  CHANGELOG (v4.18.0 section) written.
- `page_templ.go` regenerated with the CI-canonical templ v0.3.1100.
- PUSHED: master `dc28e237..6a8d46a1` + tag `loginpage/v4.18.0`
  verified via `git ls-remote` (12:03). The pre-push release-train gate
  forced aligning 4 stale requires in `examples/system-demo` (system
  v4.18.0, metaengine v4.23.0, metaengine/sqliteengine v4.8.0) and
  `systemadapter` (sqliteengine v4.8.0) — both modules build.
- Module proxy resolves the tag: `go list -m ...loginpage/v4@v4.18.0`
  → v4.18.0 (12:06).

### Webphone: the wiring — LANDED (daemon-committed b30bec47/92324ef2 et al.)

- `loginpage/v4` pin v4.17.0 → v4.18.0, `go mod tidy` clean.
- `/login` now exists in EVERY mode (server.go builds `newLoginPage`
  unconditionally; the `GET /login` route left the UserAuth gate):
  library page renders the passkey ceremony when the identity layer is
  on, and `views.LoginPageExtra` (new templ component, regenerated
  with the pinned CLI) hosts the extension+password form (bare when
  passkey off — the primary method; break-glass disclosure when on) and
  the Pocket ID door (`#sso-login` → `/api/auth/oauth2/begin`) when the
  OIDC seam is wired.
- `assets/login/login-page.js` (new, served at `/assets/login-page.js`,
  embed grew `all:login`): POSTs `/api/session` with the page's CSRF
  meta token, handles 201/401/429 with English copy, honors the
  `pbx-extension` remember key (same localStorage key the old island
  form used), redirects home where the island boot probe resumes and
  REGISTERs — the proven passkey/SSO resume path.
- Anonymous `/` now answers 303 → `/login` (no-store). Deep links keep
  the SignInHint door, whose hint/compact copy no longer points at the
  island panel; new i18n keys `signin.hint`/`signin.link` in both maps.
- Island login card DELETED: `#login-view`, `ExtensionLoginForm`,
  `PhoneIslandProps`, `main.js` submit handler + remember prefill,
  `session.js createSession`, `ui.js` els entries, `ssoerror.js`
  (trashed), 21 dead i18n keys (en+de), `ShellProps.Passkey/SSO`,
  `types/island.d.ts` shell-side login ids. Resume-failure paths now
  redirect (`/login`, and `/login?error=…` after a rejected-credential
  sign-out; the stale-session transport-failure path keeps the honest
  offline state).
- SSO failure target: `/?sso_error=` → `/login?sso_error=`; the login
  page translates the reason code into the library alert's English
  sentence (`ssoLoginFailureText`).
- `docs/dom-contract.md` restructured: six ids left the always-on shell
  block; new `login-page-contract` fenced block (login-form, ext, pass,
  remember, login-error, sso-login); SSO/passkey prose rewritten; the
  `typecheck_pin_test.go` mirror now parses `LoginPageDomIds` too and
  the doc parser covers the new fence.

### Pins: all updated and GREEN (12:28)

`go test -count=1 ./internal/server/ ./internal/web/views/
./internal/web/assets/` → all three ok after rewriting:

- `login_page_test.go`: the 404-when-off pin became
  `TestLoginPageServesTheExtensionFormWhenPasskeyIsOff` (200, form ids,
  CSRF meta, NO lp-* artifacts, linked CSS+JS actually serve); the
  live-session 303 and nonce-CSP pins unchanged.
- `passkey_view_test.go`: the card test became
  `TestLoginPageExtraAdaptsToPasskeyMode` (primary vs break-glass vs
  SSO shapes, dead hand-rolled ids stay dead).
- `server_test.go` (contract test reads a signed-in shell; F19
  nav-free probe uses a fresh anonymous client), `contract_test.go`,
  `session_test.go`, `csp_report_test.go` (deep link carries the
  reporting lane), `notfound_test.go` (island fact `#phone-view`),
  `passkey_api_test.go` (`/login` left the mode-off 404 list),
  `sso_api_test.go` (three `/login?sso_error=` targets), the
  StaticAssetsServe map (`main.js` greps `resumeSession`; new
  `/assets/login-page.js` row).

## PARTIALLY DONE (in flight at snapshot time)

- `tsc -p . --noEmit` + island `node:test` suite: RUNNING at snapshot
  (background job 061). `login-page.js` carries strict-JSDoc types and
  a new `LoginPageDomIds` interface; the vtsls editor noise was ignored
  per the type-gate doctrine (tsc is the only authority).
- Island test SPECS (`island-tests/session.test.mjs`,
  `resume.test.mjs`) still reference the deleted `createSession` /
  login-form ids — expected reds until they are updated to the new
  resume+redirect flow.

## NOT DONE (honest list)

1. login.css.input lacks `@source` of `views/login_page.templ` — the
   new section's Tailwind utilities are NOT compiled into login.css
   yet; `scripts/build-tw-css.sh` has not been re-run (login.css in the
   tree is stale for this train).
2. `app.css` has no `.wp-welcome-signin` rule (the door's new link is
   unstyled; it inherits, so it works, but it is not styled on purpose
   yet) — and any app.css edit owes the login.css rebuild anyway.
3. `scripts/webphone-smoke.py` still pins the island login (`shell
   holds island login` expects `id="login-view"` at line ~1176; the
   restart scenario extracts the CSRF meta from anonymous `/`) — it
   will FAIL against the new binary until updated to fetch `/login`.
4. `scripts/ui-capture.py` anonymous surfaces + the fixture boot's
   login-card shot: not audited against the redirect.
5. The consuming stack's `tests/browser-e2e.py` still drives the island
   login form at `/` — needs the /login form flow + the stack's
   per-train relock; AGENTS obligation "served-markup changes owe a
   fresh stack browser E2E" is UNMET.
6. vendorHash not re-pinned after the go.sum change (goModules FOD will
   mismatch on the next nix build).
7. `nix fmt` / treefmt not run over the touched Go/JS; `buildflow` full
   run not done (train-start + train-end doctrine).
8. Docs: AGENTS.md (passkey bullet + login topology + smoke lines),
   README § Passkey sign-in / login, CHANGELOG [Unreleased] entry,
   error-contract.md login-error surfaces — all unwritten.
9. 10 webphone commits sit unpushed (daemon owns pushes); CI verdict
   against this train is pending — check `gh run list` after the push.
10. `login_page.go` + `server.go` comments still carry two overlapping
    "login ceremonies share the login flood budget" sentences (harmless
    duplication; collapse during the docs pass).

## War stories from this train (for docs/lessons.md candidates)

- The cqrs-htmx loginpage module does NOT build under its own
  go.work on this host: go.work replaces `go-cqrs-lite/catalog/v4` →
  `/home/lars/projects/go-cqrs-lite/catalog`, and that subdirectory no
  longer exists. Everything loginpage must run with `GOWORK=off
  GOTOOLCHAIN=auto`. The repo's flake devshell is ALSO broken (stale
  go 1.27.3 source hash — fixed-output derivation mismatch), so
  `nix develop -c` is not a fallback there either.
- cqrs-htmx's pre-push hook runs the release-train check in
  ENFORCING mode (train lag > 0 fails the push), while its pre-commit
  runs the same check `--advisory`. A library push can be parked
  indefinitely by examples-module lag; the sanctioned fix is bumping
  the lagging requires (done), not --no-verify (the hook's own output
  names the advisory/planning distinction).
- The default-branch name is `master` in cqrs-htmx (not `main`);
  `git push origin main` silently resolves to nothing useful.
- The auto-commit daemon committed the loginpage feature mid-edit
  across THREE commits (config+handler, templ+tests, README+CHANGELOG)
  while the pre-commit hook was failing on the unrelated workspace
  breakage — the tag therefore points at a daemon-shaped history. The
  tag content is what ships; it is verified green.

## Verification ledger (commands + verdicts, 12:00–12:28)

| Command                                                                            | Verdict             |
| ---------------------------------------------------------------------------------- | ------------------- |
| `GOWORK=off GOTOOLCHAIN=auto go test -count=1 ./...` (cqrs-htmx/loginpage)         | ok                  |
| `GOWORK=off GOTOOLCHAIN=auto go build ./...` (system-demo, systemadapter)          | ok                  |
| `go list -m github.com/larsartmann/cqrs-htmx/loginpage/v4@v4.18.0`                 | v4.18.0             |
| `git ls-remote origin refs/tags/loginpage/v4.18.0`                                 | present             |
| `go build ./...` (webphone)                                                        | ok                  |
| `go test -count=1 ./internal/server/ ./internal/web/views/ ./internal/web/assets/` | ok ×3               |
| `tsc -p . --noEmit` + island node:test                                             | running at snapshot |
