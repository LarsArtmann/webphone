# Lessons — war stories with evidence

Dated hard-won knowledge moved out of AGENTS.md (2026-09-22 size
pass). AGENTS carries the one-line RULE; this file carries the STORY
and the evidence. Newest last is NOT enforced — group by topic.

## CSP and styling

- **Inline `style` attributes are CSP-dead** (2026-09-21, seen live on
  pbx.artmann.tech): the server sets `style-src 'self'` with no
  `unsafe-inline`, and style ATTRIBUTES cannot be nonce'd or
  hash-allowlisted per value (Chrome's "Applying inline style violates
  …" fires silently — the rule just never applies). The avatar hue
  therefore buckets into `wp-av-h<deg>` classes (`avatarHueClass` in
  views/helpers.go + 36 rules in app.css); keep the helper and the CSS
  block in sync. Before adding ANY `style={ … }` to a template: it
  will not work in production.

## cqrs-htmx adoption

- The `setup` bundle, CQRS dispatch layer and usermgmt stay rejected
  (split-brain identity: this product's identity is the PBX extension
  - directory password, proven by the island's SIP REGISTER — a second
    user database would be a split brain). Security presets are NEVER
    adopted wholesale — the library's `RecommendedPermissionsPolicy`
    denies `microphone`, which would kill the WebRTC phone.
    `toastDetail` is a type alias of `cqrshtmx.ToastDetail`
    (root-package type, NOT dispatch-layer), so a wire-shape change
    upstream fails this build. Trap: the dispatch-layer `Notify*`
    options emit `{level,message}`, NOT the island's `{message,kind}`
    shape; their KIND vocabulary is dispatch-layer too — the server's
    `notifyToast` emits island kinds (ok/error/warn/info), which
    main.js's old `TOAST_KINDS` (copied from the dispatch vocabulary
    success/warning) silently recolored every success toast to info —
    the mapping now lives in ui.js `toastKindFor`, pinned by ui.test.mjs.
    Audit trail (deep-dives, plans, idiomorph verdict): `docs/research/`
  - `docs/planning/` 2026-09-18..20.

## CSRF

- `httputil.CSRFTokenHXHeaders`/`CSRFTokenHTMLMeta` HTML-escape — in
  templ build the value with `templ.JSONString` and let templ escape
  once; the raw-HTML helpers would double-escape and silently strip
  CSRF from every HTMX request (`TestShellRendersValidJSONCSRFHxHeaders`).
- Login/logout rotate the CSRF token (fixation defense); the island
  adopts the fresh one WITHOUT a reload via `GET /api/csrf`
  (`session.js adoptFreshCsrfToken`); every consumer reads live (meta
  tag, htmx per-request `hx-headers`). Adoption failure falls back to
  a reload. Tests that POST after logging in must use the client
  `login` helper (it adopts) — raw login POSTs leave a dead token.
- CSRF behind a TLS-terminating proxy needs `csrf.trusted_*` (found
  2026-09-19 via the stack E2E 403s): the browser sends
  `Origin: https://host` + `Sec-Fetch-Site: same-origin`, the listener
  sees plain HTTP, and httputil's attestation check reads the mismatch
  (scheme-only: `r.Host` matches) as a FORGED attestation → 403 on
  every POST. v2.0.0 shipped this; tab logins silently never worked in
  fronted deployments (island calls kept working, so the E2E stayed
  green). Fix: `csrf.trusted_proxies` (loopback nginx) +
  `csrf.trusted_origins` (the https vhost) — `requestScheme` honors
  X-Forwarded-Proto only from trusted proxies. Module and stack set
  both; `TestCSRFTrustsTheFrontingProxy` pins the request shape.

## Nix

- Cross-build trust trap (2026-09-20): `nix build --system
  aarch64-linux` prints "ignoring the client-specified setting
  'system' … not a trusted user" — and on truly restricted paths it
  silently builds the DEFAULT system with EXIT=0 (a false-green
  aarch64 gate). Verify cross-builds by the ELF machine bytes
  (`od -An -tx1 -j18 -N2 <binary>`: `b7 00` = EM_AARCH64, `3e 00` =
  EM_X86_64), never by exit code alone. On this host the flake eval
  still honored aarch64 despite the warning — the trap is the silent
  variant elsewhere. `nix flake check --all-systems` is NOT the fix:
  it is evaluation-only for other systems (verified 2026-09-19 — zero
  derivations built, "running 0 flake checks"). Checks DO cross-build
  if you name them (`nix build .#checks.aarch64-linux.island-lint`
  ran green cross-arch).
- `buildflow -s nix-hash-fix --fix` computes the right vendorHash but
  never writes it here (buildflow itself warns). Documented deviation:
  placeholder hash → `nix build` → read `got:` → apply. Only when
  go.mod/go.sum actually changed.
- vulnix: `vulnix --closure <out-path>` is the ONLY scoped mode —
  plain vulnix expands its input into the BUILD closure (bootstrap
  toolchains: dozens of findings that never deploy); `nix run
  .#vulnix` wraps the correct call. vulnix range-matches
  distro-patched versions: grep the flagged CVE ids against the LOCKED
  rev's glibc patches (`nix eval github:NixOS/nixpkgs/<rev>#glibc.patches`)
  before believing a finding (all 8 glibc findings of 2026-09-19/20
  were patch-covered). The exit-nonzero-on-findings shape is expected
  noise; the runtime closure itself carried zero real advisories at
  last scan.

## Tooling traps

- Dynamic awk/grep patterns must ESCAPE `[`/`]`: release.sh's notes
  extractor matched `^## [2.4.0]` as a dynamic awk regex, where
  `[2.4.0]` is a bracket expression (one char of {2,.,4,0}) — it never
  matched the literal heading and shipped v2.3.0/v2.4.0 with EMPTY
  GitHub release bodies. The fold-check grep passed because ITS
  pattern had shell-escaped `\[`. Found by the 2026-09-20 release
  audit; both objects backfilled from CHANGELOG.
- Verify dependency internals at the CONSUMED tag (module cache or
  `git show v4.9.0:<path>`), never master: the 2026-09-18 audit
  over-credited v4.9.0's `ServeSSE` with a `retry:` hint that only
  exists on master — tag-checking before the port caught it. (True at
  the time; v4.11.0 DOES ship the hint — MD1's bump trigger fired and
  was executed 2026-09-20.)
- htmx/cqrs-htmx BUMP CHECKLIST (plan T25, 2026-09-20): re-verify at
  the new version (1) the `responseHandling` contract — entry keys
  `code`, `swap`, `error`, `target`, `select` in the SERVED
  htmx.min.js plus the htmx-config markers pinned in
  `TestServedPageHoldsTheDomContract`; (2) the `/events` `retry:` hint
  (`TestSSEStreamCarriesConnectedThenEvents`); (3) `HX-Trigger` still
  fires BEFORE the swap decision (the skip-guard in shell.js §3c
  depends on it); (4) the sse extension still fires cancelable
  `htmx:sseBeforeMessage`.
- json/v2 went STABLE in Go 1.27 (2026-09-22 sweep): no
  `GOEXPERIMENT` anywhere — flake devShell, scripts, buildflow env
  and docs all dropped it after the full suite proved green without
  it (13/13 packages). The old trap (commands failing outside the
  shell for a missing flag) is gone; the remaining toolchain trap is
  the host's below-floor go — use `nix develop -c`.

## Telephony

- sip.js is pinned at 0.21.2 (vendored tarball + IIFE bundle). The 0.x
  series hangs in `userAgent.reconnect()` after transport loss; the
  bounded watchdog in the island's `connection.js` (5s per attempt,
  full rebuild on timeout) is load-bearing. Do not "simplify" it away.
  The watchdog ALSO rebuilds on registration loss (2026-09-22, the
  1001-anomaly fix): a Registerer that had reached `Registered` and
  later goes `Unregistered`/`Terminated` outside logout/rebuild is
  dead weight (retrying `register()` on it never recovers), so
  `registrationLost` tears the agent down and builds a fresh
  Registerer; a never-registered rejection keeps the `regRejected`
  pill (bogus-login UX + the E2E `registration rejected` contract).
  Pinned by `island-tests/connection.test.mjs`. The stack E2E carries
  the sofia tripwire for recurrences (`REGS-AT-RECONNECT`/
  `REGS-AT-DIAL` dumps in its browser.nix).
- **sip.js fires NO stateChange when a Registerer re-registers without
  leaving `Registered`** (a transport loss does not demote it): any UI
  that keys on registration state must be refreshed by the recovery
  path itself, never only by the listener. The 2026-09-22 E2E chain:
  reconnect succeeded, the pill stayed on the stale backoff text, the
  suite read "stuck", reload-fell-back, and the resumed pages never
  clicked login (breaking the notification-permission marker). The
  reconnect success path in `connection.js` sets the pill explicitly.
  The stack E2E suite runs ~293-322s under the 2026-09-22 scenario
  set (restart + transfer + FS-outage drills); budget re-baselined to
  445s on 2026-09-22 (two forced-rebuild runs 384s/373s).
- SDK decision: KEEP sip.js 0.21.2; **JsSIP 3.13.8 is the named
  fallback** (the only maintained alternative). Swap ONLY on Chromium
  WebRTC breakage, a sip.js security advisory, or a needed capability
  — never speculatively; a swap is a full island call-path rewrite
  plus a stack browser-E2E re-run. All Go-side telephony REJECTED
  (sipgo / pion / ESL): the browser terminates media anyway,
  server-side signaling only adds state to the hottest path.

## Server details

- `pbx.Client` owns the timeout-bounded HTTP client; the `/phone-api`
  proxy must ride `PhoneAPI.HTTPClient()`, never `http.DefaultClient`.
  Join path and query separately — `url.JoinPath` percent-encodes `?`
  (a test caught upstream receiving `history%3Flimit=30`).
- htmx loads deferred from `headExtras`, after the `htmx-config` meta
  that disables its inline indicator-style injection (`app.css` ships
  the same rules so hx-indicator keeps working). The meta must precede
  the script or htmx never reads it; Shell leaves `HTMXSrc` unset so
  `layout.Base` does not also emit a synchronous htmx tag.
- The app.css `[data-theme]` `color-scheme` rules used to carry
  `!important` because the CSP-hash-pinned framework theme script wrote
  an inline `colorScheme` that could undo a forced theme. RESOLVED
  2026-09-22: templ-components v1.19.2 shipped the `NoThemeScript`
  knob; webphone consumes it, serves its own same-origin
  `/assets/theme-preload.js` (which sets `data-theme` pre-paint — the
  library script never did), and dropped the hash plus the
  `!important`s. Lesson: when a dependency forces a CSP exception, the
  fix belongs upstream — and the exception class (hash pinning
  dependency bytes) was the real recurring cost, not the one-time pin.

## E2E

- Stack browser E2E flake mode: a SLOW run can die at the transfer
  step — FreeSWITCH hangs the call with RECOVERY_ON_TIMER_EXPIRE ~90s
  after DTLS-ready (ICE/media inactivity ceiling), the island removes
  the dead call card, the E2E's fresh `.transfer-btn` click goes
  StaleElementReference. Verdict rule: registration + DTMF + ICE
  stats green before a ~90s death = flake, not an island regression —
  re-run once before digging.
