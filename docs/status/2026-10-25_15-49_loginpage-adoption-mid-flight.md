# Status: cqrs-htmx/loginpage adoption — mid-flight (phase 4 of 9)

Claims verified as of 15:49 CET, 2026-10-25. Build state: `go build ./...` GREEN
inside `nix develop`. Origin/main = `c7ca563` (CI green); HEAD = `fee3e57`
(TWO unpushed daemon commits, see § d).

## Mission

Owner directive: adopt `/home/lars/projects/cqrs-htmx/loginpage` ("and co")
as the login implementation, replacing the hand-rolled passkey auth UI that
kept producing auth bugs.

## The chosen architecture (researched, decided, mostly wired)

- Pin the PUBLISHED tag `loginpage/v4.13.0` — its go.mod requires exactly
  what webphone already pins (usermgmt v4.14.3, templ-components v1.23.0,
  httputil v1.5.0). ZERO dependency version bumps. The cqrs-htmx working
  tree has moved past the tag (usermgmt v4.16.0 requires) — deliberately
  NOT ridden.
- `/login` = the library page (passkey ceremony ONLY, `NoRegistration` —
  enrollment stays token-gated at `/enroll`; no OAuth2 buttons — the island
  card keeps the Pocket ID link, one door per method).
- Library JS wire contract verified IDENTICAL to webphone's ceremony
  endpoints: begin `{email}` → `{options, session_key}`; finish
  `?user_id=<session_key>` + raw assertion; then redirect `/` where the
  island boot probe signs in + REGISTERs (the SSO lane's proven resume path).
- Server-side handlers UNCHANGED (they mint the webphone session + SIP
  credential — the thing the library does NOT do). Only the PATHS moved to
  the usermgmt-canonical `/api/auth/webauthn/login/{begin,finish}` the
  library posts to. No external caller greps the old paths (verified:
  smoke script, stack browser E2E, Android shell, openapi/wirecontracts —
  zero references).
- CSP: `/login` alone overrides the strict app CSP with a per-request
  nonce lane (`script-src 'self' 'nonce-…'`) — same pattern as the proven
  health-dashboard subtree. The shell keeps zero-inline-script strictness.
- Theme: library ThemeScript (nonce'd) follows prefers-color-scheme; the
  login Tailwind build's dark variant extended to also match `.dark`
  (see open question 2 for the forced-theme caveat).

## a) FULLY DONE

1. Research: loginpage tag vs HEAD diff, feature set at v4.13.0
   (NonceFromRequest/ThemeFromRequest/NoRegistration/AuthPrefix all
   present at the tag), usermgmt CHANGELOG reviewed (v4.15/v4.16 changes
   are why we stay on the tag), httputil CSRF helpers verified compatible
   (`csrf-token` meta + `X-CSRF-Token` header — same library as our
   middleware).
2. External pin-site audit: smoke script, stack E2E, Android shell,
   apidocs/wirecontracts, CSP test (probes only `/`), notfound tests —
   none reference the passkey card or old paths. DOM contract doc is the
   only external-contract surface to update.
3. Dependency added: `loginpage/v4 v4.13.0` in go.mod/go.sum, tidied.
4. `userauth.Service.Usermgmt()` accessor added (read-only page wiring;
   ceremonies stay behind the seam).
5. `internal/server/login_page.go` written: `newLoginPage()` (Config:
   Brand "WebPhone", Redirect "/", CSSPath "/assets/login.css",
   AuthPrefix "/api", NoRegistration, NonceFromRequest),
   `loginContentSecurityPolicy(nonce)`, `loginPage` handler (live session
   → 303 `/`; CSP override; nil-handler → styled 404).
6. `server.go`: handler built in `New` when `deps.UserAuth != nil`
   (build failure → loud log + nil, /login 404s); routes registered:
   `GET /login` behind `httputil.Nonce`; passkey login begin/finish
   RENAMED to `/api/auth/webauthn/login/{begin,finish}` (same handlers,
   same login flood budget); enroll routes untouched.
7. `pages.go`: `loginPageHandler` field on `handlers`.
8. `phone.templ`: passkey form block → `<a id="passkey-login"
   class="primary wp-passkey-link" href="/login">` + break-glass
   (SSO/extension-only shapes untouched → passkey-off deployments
   byte-identical).
9. Island cleanup: `passkey.js` + `passkey.test.mjs` DELETED (trash);
   `main.js` import+call removed; `ui.js` passkey els removed;
   `webauthn.js` lost its now-dead `prepareLoginOptions` (login half —
   the library owns login coercion now) + header re-scoped; STAYS for the
   enroll flow (verified: `assets/enroll/enroll.js` imports
   `prepareRegistrationOptions`/`serializeCredential` from it).
10. `webauthn.test.mjs`: dead helper's import + 2 test blocks removed.
11. `go build ./...` green after all Go-side edits.

## b) PARTIALLY DONE

1. **i18n.js key prune BLOCKED mid-edit**: the multiedit was REJECTED by
   the read-before-edit guard (I had only sed-viewed the file). Six dead
   keys × en/de maps still present: `passkeyEmail`, `passkeyUnknown`,
   `passkeyThrottled`, `passkeyBeginFailed`, `passkeyFinishFailed`,
   `passkeyNetFailed` (keep `passkeySection`/`passkeySignin`/
   `passkeyBreakglass` — the link + break-glass use them).
2. Island JS edits (main/ui/webauthn/tests) are correct but UNCOMMITTED —
   the daemon will sweep them; i18n.js must land in the same conceptual
   unit to keep the island-test suite green.
3. phone.templ edited but `templ generate` NOT run → `*_templ.go` stale →
   any Go test run right now exercises the OLD markup (false confidence).

## c) NOT STARTED

1. **vendorHash re-pin** (see § d — now the most time-critical item).
2. `passkey_api_test.go`: still pins `/api/auth/passkey/begin|finish`
   (`passkeyBegin` const + literal URLs) → WILL FAIL; move to
   `/api/auth/webauthn/login/*`.
3. `views/passkey_view_test.go` `TestLoginCardAdaptsToPasskeyMode`: still
   asserts the form markup → rewrite for the link shape.
4. New `/login` tests: CSP nonce lane (nonce in header, scripts carry it,
   no unsafe-inline), live-session 303, passkey-off styled 404, CSRF meta
   present, `lp-*` ids conditional.
5. `docs/dom-contract.md`: conditional passkey ids section rewrite
   (`passkey-login` link on `/`; `lp-*` library ids on `/login`);
   `TestServedPageHoldsTheDomContract` reads this doc.
6. **login.css artifact**: `login.css.input` (@import tailwindcss +
   @import app.css for tokens + extended dark variant incl. `.dark` +
   @theme remap + @source = current set + loginpage
   page.templ/page_templ.go/assets/login.js + templ-components feedback
   alert/spinner sources) + build script extension + consistency gate vs
   tw.css.input. Without it `/login` renders unstyled (CSSPath points at
   it already).
7. app.css: styling for `a.primary`/`.wp-passkey-link` (anchor vs button
   rendering unverified).
8. Gates: templ generate, nix fmt (island+css scope), island node:test
   suite, `go test -count=1 ./...`, session-behavior suite
   (`go test ./internal/server/ -run TestServer` — the `auth:` convention),
   buildflow, `nix flake check`.
9. Docs: README § Passkey sign-in rewrite, CHANGELOG `[Unreleased]`,
   AGENTS.md (passkey bullet + DOM-contract note + login.css in the tw
   build story), FEATURES.md, TODO_LIST row, release-runbook note (markup
   change owes fresh stack browser E2E).
10. Smoke boot of the passkey-on binary (needs the HOSTNAME rp_id dance)
    + verify `/login` renders styled + ceremony round-trip.

## d) TOTALLY FUCKED UP (and what I did about it)

1. **`go.mod`/`go.sum` landed in daemon commits WITHOUT the vendorHash
   re-pin** — `fee3e57` + `b1cdd0d` are unpushed (verified
   `origin/main=c7ca563`, CI green there). If the daemon pushes before I
   re-pin, `nix flake check`'s FOD mismatches = red main (the exact
   2026-10-06/10-09 class; AGENTS.md calls a bare go.mod delta in a
   daemon commit a defect). Root cause: I deferred the re-pin to the end
   of my plan; the daemon does not respect my phases. Fix path: run the
   recipe NOW (`nix build .#webphone.goModules --rebuild` → read `got:` →
   apply) before the 60-min push-stall bar is met (first unpushed commit
   ~15:2x; bar at ~16:2x).
2. `go get` before the import existed → `go mod tidy` dropped the module;
   wasted a cycle re-adding after imports landed. Order should have been
   imports-first.
3. i18n.js edit rejected for not View-ing the file first (tool-discipline
   slip; blocked phase 4 completion).
4. An older CI run (40202082229, 14:24) on `a77e510` shows `failure`,
   then `c7ca563` (14:28) green — pre-session noise, already recovered;
   noting for the record, NOT mine.

## e) WHAT WE SHOULD IMPROVE (session lessons)

1. Dependency-add order: imports → `go get` → tidy → vendorHash re-pin in
   ONE unit, BEFORE any daemon-visible intermediate state (the daemon is
   adversarial to phased plans — AGENTS.md says so; I planned in phases
   anyway).
2. Run `templ generate` + targeted tests IMMEDIATELY after each .templ
   edit; a stale `*_templ.go` makes green tests lie.
3. View-tool (not bash cat/sed) every file before editing it — the guard
   exists for a reason.
4. `scripts/build-tw-css.sh` has a LATENT pre-existing breakage found
   during research: it copies from `vendor/github.com/larsartmann/
   templ-components/` but vendor/ was trashed 2026-10-17 and does not
   exist — the script cannot run as-is today. The login.css work must fix
   this (resolve module dirs via `go list -m -f {{.Dir}}` instead of the
   vendor tree).
5. LSP diagnostics were stale/noisy all session (phantom typecheck
   warnings after edits landed) — trust `go build`, not the panel.

## f) NEXT (ordered, ~30 concrete items)

1. vendorHash re-pin NOW (§ d.1 recipe) — beats the daemon's push.
2. View + edit `island/app/i18n.js`: drop the six dead keys × en/de.
3. `templ generate ./internal/web/views/`.
4. Fix `passkey_api_test.go` paths (`passkeyBegin` const, finish URLs,
   off-mode 404 list) → `/api/auth/webauthn/login/*`.
5. Rewrite `TestLoginCardAdaptsToPasskeyMode` for the link card.
6. Add `/login` server tests (CSP nonce, 303, 404-off, CSRF meta, lp-*
   ids, `Cache-Control: no-store`).
7. Update `docs/dom-contract.md` (conditional ids: `passkey-login` on /;
   `lp-*` on /login; note the library owns that markup).
8. Write `internal/web/assets/login.css.input`; extend/generalize
   `scripts/build-tw-css.sh` (fix the vendor/ dependency, stage loginpage
   + feedback sources from the module cache, add @theme/variant parity
   gate between inputs); build + commit `login.css`.
9. app.css rule for `.wp-passkey-link` (verify against `.wp-sso` style).
10. `nix fmt` (scoped: island, css, Go artifacts).
11. Island tests: `nix run nixpkgs#nodejs -- --test --test-force-exit
    internal/web/assets/island-tests/*.test.mjs`.
12. `go test -count=1 ./...` + `-run TestServer` session suite.
13. buildflow full run (inside nix develop).
14. `nix flake check`.
15. Smoke boot: passkey-on config + HOSTNAME rp_id dance; curl /login
    (styled, nonce'd CSP, lp-* ids); loopback ceremony round-trip if
    feasible; default boot byte-parity check (passkey off → /login 404).
16. Docs: CHANGELOG `[Unreleased]` (code-path change → line owed), README
    § Passkey sign-in (new /login flow + paths), AGENTS.md (adopted-
    loginpage bullet: paths, CSP nonce lane, login.css, theme caveat,
    loginpage pinned at v4.13.0 + why not HEAD), FEATURES.md, TODO_LIST
    row, release-runbook E2E obligation note.
17. Verify daemon push lands + CI green (`gh run list`) after vendorHash
    commit.
18. Consider (owner call): whether `internal/server/passkey_api.go`'s
    header comment needs the new paths named.
19. Health dashboard parity check: confirm the /login nonce lane adds no
    `/metrics` label leakage (aggregates only — no action expected).
20. Android shell: no action needed (verified — resumes via session API;
    /login stays in-shell by NavigationVerdict).

## g) QUESTIONS (cannot figure out myself)

1. **Stack E2E obligation**: served markup changed (island passkey card →
   link; new /login page). Verified nothing in the stack's browser E2E
   references the old card, but the runbook says markup changes owe a
   fresh stack browser E2E. Run it now against the local stack, or do you
   take it at the next release train?
2. **/login theme**: forced light/dark lives in `localStorage["wp-theme"]`
   — unreadable server-side, so /login follows the OS theme
   (prefers-color-scheme) even for users who forced a theme. Ship as-is
   (one transient page), or plumb a `wp-theme` COOKIE the shell writes +
   `ThemeFromRequest` reads (new cookie state, SSR-correct theme)?
3. **Daemon commits vs the `auth:` convention**: AGENTS.md demands an
   `auth:` prefix on passkey-surface commits; the daemon already committed
   the core wiring as `chore:` (unpushed `fee3e57`+`b1cdd0d`). Amend that
   history (rewrite two daemon commits), or leave them and let the
   narrative `auth:`-prefixed commit + CHANGELOG carry the convention?
