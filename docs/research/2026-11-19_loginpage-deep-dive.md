# loginpage/v4 deep-dive — the /login front door's utilization audit

> Snapshot-rot header: claims verified as of 2026-11-19 ~13:55 CET, at pin
> `loginpage/v4 v4.18.0` (module source read from the build cache). Push/CI
> claims rot fastest; everything below is timestamped to this minute.

Series: the library-deep-dive lane (`docs/research/README.md` index). This is
the FIRST loginpage deep-dive — the seam was adopted 2026-11-11 (the
`Config.ExtraSection` front-door train), so the audit covers ~8 days of
consumption, not years.

## Verdict

**92/100 — deep, deliberate adoption with one real gap (now fixed) and a
clean decline map.** The consuming seam (`internal/server/auth.NewLoginPage`)
uses the library exactly the way its doc comments prescribe; every unused
knob is either N/A by construction or declined by a recorded webphone ruling.
The audit found exactly one underutilization with user-visible value — the
favicon accent — and fixed it in the same session.

## Capability map (every exported knob of v4.18.0)

| Capability | Webphone use | Verdict |
| --- | --- | --- |
| `Config.Service` | passed conditionally (nil when passkey off) | **ADOPTED** — the nil-safe shape the ExtraSection docs prescribe; page still renders the extension form |
| `Config.Title` / `Config.Brand` | "Sign in" / "WebPhone" | ADOPTED |
| `Config.Redirect` | "/" | ADOPTED — post-login shell + boot-probe resume |
| `Config.CSSPath` | "/assets/login.css" | ADOPTED — the scoped Tailwind artifact, rebuilt by `scripts/build-tw-css.sh` |
| `Config.AuthPrefix` | "/api" | ADOPTED — the usermgmt-canonical ceremony paths |
| `Config.NonceFromRequest` | `httputil.NonceFromRequest` | ADOPTED — pairs with the /login CSP override (which this train extended with the report-uri/report-to tail) |
| `Config.NoRegistration` | true | ADOPTED — enrollment stays token-gated at /enroll |
| `Config.ExtraSection` | `views.LoginPageExtra(!passkeyOn, sso != nil)` | ADOPTED — the 2026-11-11 train's whole point; usage matches the upstream contract (per-request values would ride the render context — none needed today) |
| `Config.AccentColor` | **was unset → NOW "#17a467"** | **was the gap, fixed same session** — see below |
| `Config.Theme` / `Config.ThemeFromRequest` | unset | **DECLINED by ruling** — the login page follows `prefers-color-scheme` deliberately; forced-theme users see the OS theme there (recorded in AGENTS.md + login.css notes). `ThemeFromRequest` is the future door IF that ruling ever flips — one field, no rework |
| `Config.OAuth2Buttons` / `Config.NoOAuth2` | unset | **N/A by construction** — webphone configures ZERO OAuth2 providers on `usermgmt.Service` (its OIDC is the internal/sso seam, not a service provider), so button auto-detect finds none. `NoOAuth2` would be redundant belt-and-braces; setting it is optional hardening against a future service-provider config drift |
| `Config.NoWebAuthn` | unset | N/A — `Service.HasWebAuthn` already drives section visibility; the service-level flag is the truth |
| `Config.AutoRedirect` | unset | N/A — ExtraSection counts as a sign-in method, so OAuth2 is never "the only door"; the shortcut structurally cannot fire |
| `Config.CredentialName` / `Config.RegisterFirst` | unset | N/A — registration hidden |
| `Handler.ServeHTTP` (via manual mux route) | yes | ADOPTED — webphone mounts through its own mux chain (limiters, CSP override) instead of `Mount()`; equivalent, deliberate |
| `Handler.Mount` | no | N/A — same as above |
| `Handler.Routes()` (posture manifest) | no | **DECLINED by architecture** — webphone consumes no `cqrshtmx.RoutePosture` surface anywhere (the setup-bundle rejection); nothing to feed |
| `NewPageData` | no | N/A — the Handler renders; webphone never re-renders the page itself |
| `?error=` surface (`errorMessageFromQuery`, 300-char cap) | yes | ADOPTED — the `sso_error` → `error` rewrite rides it; sentences are short, cap never bites |
| embedded `loginJS` + ceremony endpoints | yes | ADOPTED — the library JS drives `/api/auth/webauthn/login/{begin,finish}` against webphone's own handlers |

## The finding (fixed): the favicon was indigo

`page.templ` renders `<link rel="icon" href={p.faviconURI()}>` — a
brand-initial square in `Config.AccentColor`, whose default is
`DefaultAccentColor = "#4f46e5"` (indigo). Webphone's brand is GREEN
(`--accent #17a467`; the 2026-11-11 owner veto: "my app is green, not
Tailwind default" — indigo/blue reads default-SaaS and is WRONG for this
product). The auth front door — the one page every anonymous visitor sees —
was rendering the library's indigo default in the browser tab.

Fixed in `internal/server/auth/auth.go` (`AccentColor: "#17a467"`), login
tests green. The shell's own PWA icons were already green; the two doors now
match.

## Misuse findings: none

- The ExtraSection usage matches the upstream contract verbatim (including
  the "counts as a sign-in method" semantics that keep `New` from warning).
- The nonce lane pairs correctly with the per-request CSP override (both
  read the same middleware nonce).
- NoOAuth2/OAuth2Buttons mutual exclusion is respected trivially (neither
  set).

## Version delta

v4.18.0 is the latest tag on the module (verified by the same-day
dependency currency sweep, `docs/research/2026-11-19_dependency-currency-sweep.md`
— the identity family is not in the behind list). No upgrade decisions owed.

## Refresh trigger

Re-open this audit when: the forced-theme ruling flips (→ `ThemeFromRequest`
adoption question), an OAuth2 provider ever lands on `usermgmt.Service`
(→ `NoOAuth2` hardening question), or the module bumps a minor (the Config
surface is the contract to re-read).
