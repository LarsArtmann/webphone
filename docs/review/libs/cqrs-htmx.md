# cqrs-htmx — library utilization audit

**Verdict:** AFFIRMED at the prior round's ceiling: 100/100 — the round-3 audit's three opportunities were all shipped the same day by the SUPERB train and remain wired. This report re-verifies the pins and re-states the boundaries; it does not re-litigate. **Adoption 100/100 (very high; affirmed)** — 14 of 14 post-train rows per the round-3 rubric.
**Pin:** root /v4 v4.18.1, identity-model/v4 v4.17.0, loginpage/v4 v4.18.0, usermgmt/v4 v4.20.0, usermgmt/webauthn/v4 v4.13.0 (all direct) · **Latest (family tags):** root v4.18.1 / usermgmt v4.20.0 / loginpage v4.18.0 / identity-model v4.17.0 · **Currency:** pins = latest, re-verified against remote tags 2026-11-11 (the 2026-11-07 family train: root v4.16.0→v4.18.1, identity-model v4.15.0→v4.17.0, usermgmt v4.19.0→v4.20.0; loginpage v4.16.0→v4.18.0 on 2026-11-11) · **Status:** direct dep
**Audited:** 2026-11-04 · **Currency re-verified:** 2026-11-11 (remote tags + tree; the ExtraSection adoption below) · **Authoritative prior:** [2026-10-25 round-3 deep-dive](../../research/2026-10-25_cqrs-htmx-deep-dive.html) (71/100 at audit time → **re-scored 100/100 same day**, §07 Addendum; supersedes the 09-18/09-19 rounds). **Movement: 100 → 100 affirmed at this audit (the three round-3 opportunities are all SHIPPED and pinned)**.

## Status

|         |                                                                                                                                                                                  |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Modules | root /v4 + identity-model + loginpage + usermgmt + usermgmt/webauthn (5 direct modules of 15 in the family)                                                                      |
| Usage   | 15 files: internal/userauth/* (the identity lane), internal/server/{server,login_page,sse,pages,panels,toast,passkey_api,hooks/webhooks}.go + readiness/middleware/healthz tests |

## Family deltas since the 2026-11-06 report (all consumed or ruled)

- **loginpage v4.18.0 `Config.ExtraSection` (2026-11-11) — ADOPTED same day.** `/login` is now THE auth front door in EVERY mode: the library page hosts webphone's own lanes in its ExtraSection slot (`views.LoginPageExtra` — the extension + SIP-password form ALWAYS, the Pocket ID door `#sso-login` when the OIDC seam is on; internal/server/login_page.go:58). The island card carries no login form anymore; every lane lands on `/` where the boot probe resumes. `NoRegistration` stays (enrollment is CLI-token-gated); zero library OAuth2 buttons (the SSO door is webphone's own lane).
- **root v4.18.1 — pristine embedded assets, passively consumed.** htmx core 2.0.10→2.0.11 and idiomorph 0.7.4→0.8.0 with sha256 pins (server.go:626 mounts `htmx.ExtensionsHandler(htmx.ExtSSE, htmx.ExtIdiomorph)` behind `/htmx-ext.js`): webphone's morph-swap surfaces ride idiomorph 0.8.0's focus-restore refinements without touching a byte locally.
- **root v4.18.0 — HTTP QUERY (RFC 10008) surface: not applicable.** `MethodOverrideMiddleware`/`RequireAnyMethod`/`RequireContentType`/`AcceptQuery` serve CQRS-over-HTTP apps; webphone's tabs are server-rendered panels on POST/GET and the root `App` pipeline is already ruled out below. The Go floor 1.27.3 matches webphone's go.mod.
- **usermgmt v4.20.0**: `ArchitectureCatalog()` (events-only published language — webphone's catalog lane already rides go-cqrs-lite/catalog for the whole app), the session-lifecycle audit journal (decorates usermgmt's OWN SessionStore — webphone deliberately ignores usermgmt sessions; the durable auth-audit journal webphone reads is unchanged), `GrantRoles`/`RevokeRoles` durable role grants (webphone's admin is the shared `auth.users` mapping, not library roles). **identity-model v4.17.0**: `SessionStarted`/`SessionEnded` events + strict `ParseTenantID`/`ParseBotID` — the passkey door consumes `ParseUserID` + sentinels only.

## Round-3 verdicts — re-verified in this tree

| Round-3 finding                                     | Status in this tree  | Evidence                                                                                                                                                                                                             |
| --------------------------------------------------- | -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ProjectionReadinessCheck + OnProjectionFailed wired | ✅                   | internal/server/server.go:585 (`ProjectionReadinessCheck` in the /healthz checks); internal/userauth/userauth.go:131-133 (`OnProjectionFailed` WARN hook)                                                            |
| Identity train fully landed (usermgmt v4.16+ then)  | ✅ + bumped          | pins now usermgmt v4.20.0 / identity-model v4.17.0 / loginpage v4.18.0 (2026-11-07..11 trains); clone detection pinned by test; durable auth-audit journal verified queryable in a live boot (AGENTS.md, 2026-10-25) |
| `ExtensionHubs.Shutdown` pre-drain                  | ✅                   | internal/server/sse.go:116 (releases live /events streams before the 30s HTTP drain; cmd/webphone/main.go:196)                                                                                                       |
| `setup` bundle stays REJECTED                       | ✅ unchanged         | cmd/webphone/main.go:189-193 records the measured +68% binary delta go/no-go; lifecycle rides `httputil.NewServer`                                                                                                   |
| Remaining gap: memory-only AuditLog read projection | ✅ still the one gap | UPSTREAM (usermgmt audit_log.go, DefaultAuditLogLimit=10_000, memory-only); webphone's durable alternative is its own auth-audit journal + `agent_audit` store                                                       |

## How webphone uses it (surface summary)

- **Root /v4**: `Chain`, `ContextEnrichmentMiddleware(nil)`, `RequestLoggingSlog`, `RecoveryMiddleware`, `NamedCheck` + `ReadinessHandler` + `DebugHandler` (the ONE readiness truth, per-check timeouts), `ProjectionReadinessCheck` + `ProjectionStatusProvider` (Deps override for tests), `Broadcaster`/`NewBroadcaster` (the SSE wrap), `HTMXScriptHandler` + `HTMXExtensionsHandler(HTMXExtSSE, HTMXExtIdiomorph)` (vendored ext bundle), `SafeDetail` (error redaction), `ToastDetail` (HX-Trigger wire shape).
- **loginpage**: the `/login` front door — `loginpage.New(Config{Service (nil-safe), Title, Brand, Redirect:"/", CSSPath:"/assets/login.css", AuthPrefix:"/api", NonceFromRequest, NoRegistration:true, ExtraSection: views.LoginPageExtra(...)})` (internal/server/login_page.go:58-70); the ExtraSection hosts the extension form + Pocket ID door, counts as a sign-in method, and suppresses AutoRedirect by construction; webphone overrides CSP per-subtree with the nonce lane; anonymous root navigation 303s here, live sessions skip to `/`.
- **usermgmt + webauthn**: in-process identity service on its own `usermgmt.db` (`NewSQLEventStore` → `OptimizeSQLiteDB` → `NewService(ServiceConfig{WebAuthn: waprovider.New(...), ReadModelDialect:"sqlite", OnProjectionFailed: ...})`, internal/userauth/userauth.go:79-146); webphone's own session store is the single session truth — the library's FinishLogin session is deliberately ignored; SIP directory password sourced from operator files at login, never cached.
- **identity-model**: `ParseUserID`/sentinels (`ErrUserNotFound`/`ErrNoCredentials`/`ErrAccountLocked`/`ErrEmailExists`) mapped to HTTP at the passkey API door (internal/server/passkey_api.go:37-103).

## Not-applicable map (each a claim, with its reason)

| Module/surface                                                                                                                       | Why not                                                                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `setup/v4` bundle                                                                                                                    | measured +68% binary delta + split-brain identity; rejected 2026-09-30, evidence pinned in main.go:189-193                                                                                              |
| `usermgmt/oauth2`, loginpage OAuth2 buttons, `usermgmt/totp`                                                                         | SSO is webphone's own OIDC lane (internal/sso); login page renders zero OAuth2 buttons by config; no TOTP requirement                                                                                   |
| usermgmt sessions / `NewSessionMiddleware`                                                                                           | webphone owns the session store + sliding renewal (internal/session); the docstring at userauth.go:6-8 records the split                                                                                |
| usermgmt `AuditLog` projection                                                                                                       | memory-only upstream projection; webphone reads its own journal instead — reconsider only when upstream ships a bounded projection                                                                      |
| root `App`/`Command`/`Query` CQRS pipeline, `CSRFProtect`, `Response` builder, ACK protocol, Sync handlers, `JournalSSEStore` replay | webphone is not a CQRS HTTP app: tabs are server-rendered panels; CSRF/security/composition ride httputil + own middleware (round-3 ruling); SSE replay is the go-sse seam ([go-sse report](go-sse.md)) |
| `adminui`, `dashboardui`, `auditlog`, `datastar`, `systemadapter`, `health/v4` modules                                               | health rides go-health/go-health-dashboard directly; datastar only ships via the dashboard bundle; admin/audit surfaces have no product home here                                                       |
| root v4.18 HTTP QUERY ladder (`MethodOverrideMiddleware`, `AcceptQuery`, `ScriptHandler4`, `ETagResults`)                            | webphone is not a CQRS HTTP app (standing round-3 ruling); htmx stays on the 2.x default handler; conditional GETs ride go-etag/server on webphone's own documents ([go-etag report](go-etag.md))       |

## Score

Carried from the round-3 §07 Addendum: **14/14 post-train rows fully leveraged → 100/100**. This audit re-verified the three shipped findings in the current tree (table above) and found no regressions, no new missed capability in the adopted modules, and pins at or above audit-time latest.

## Ruling

Stay the course. The two live items are upstream-shaped: (1) the bounded AuditLog read projection (the standing refresh trigger), (2) any `loginpage` option that would let the forced-theme users get the OS theme on `/login` (minor, note-only). Nothing to adopt locally.

**Refresh trigger:** usermgmt major bump; a bounded AuditLog projection upstream; any second identity consumer (then re-check `adminui`).

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "ProjectionReadinessCheck" internal/server/server.go | head -2
rg -n "OnProjectionFailed" internal/userauth/userauth.go | head -2
rg -n "ExtensionHubs" internal/server/sse.go | head -3
rg -n "setup" go.mod | wc -l                              # expect 0
rg -n "ExtraSection" internal/server/login_page.go        # the v4.18.0 adoption
rg -n "ExtensionsHandler" internal/server/server.go       # the pristine-asset mount
```
