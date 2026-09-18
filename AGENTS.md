# AGENTS.md

Enduring context for AI sessions working in this repo.

## What this is

A standalone SIP.js WebRTC softphone UI (static site derivation),
extracted 2026-09-17 from `nix-international-telephony`
(`packages/webphone/`, v0.2.0 lineage). That stack now consumes this
flake as an input (`github:LarsArtmann/webphone`) and serves the package
at `https://<domain>/` behind nginx; the module there owns config.js
rendering, the wss proxy and TLS. This repo owns ONLY the UI and its
packaging: `src/` (ES modules, entry `app/main.js`) + `package/default.nix`
(esbuild: pinned sip.js tarball → `sip.min.js` browser global; modules →
single classic-script `app.js`).

## Commands

```console
nix flake check          # package build + treefmt + statix + deadnix (fast, <1 min)
nix fmt                  # treefmt: nixfmt (nix) + prettier (src/**/*.js, html, css)
nix build .#webphone     # the static site
./package/update.sh      # repin sip.js (arg: version, default: latest)
```

No Makefile, no justfile. BuildFlow local default is fast; a full
pipeline is unnecessary here (the whole gate realizes in under a minute).

## The DOM + bundle contract (DO NOT BREAK CASUALLY)

The consuming stack's VM test (`tests/webphone.nix`) and browser E2E
(`tests/browser-e2e.py`) drive this page remotely. Verified against the
upstream suites at extraction time; re-run them there after any change
to markup or the app bundle:

- **Served page asserts**: contains `WebPhone`, links `sip.min.js`;
  elements `id="keypad"`, `id="remember"`, `id="history-list"`,
  `id="vm-wrap"`, `id="contacts-wrap"`, `id="ice-wrap"`.
- **Served app.js asserts these exact strings** (so the bundle is built
  WITHOUT `--minify` on purpose — minification renames them away):
  `dtmf-relay`, `userAgent.reconnect()`, `.refer(`, `blindTransfer`,
  `attendedTransfer`, `Notification.requestPermission`,
  `titleFlashStart`, `ringToneStart`, `phone-api/history`,
  `phone-api/voicemail`, `candidate-pair`, `currentRoundTripTime`.
- **Browser E2E drives these hooks** (read `.text`/`textContent`,
  click, type): `#reg-status` (text `registered` /
  `registration rejected`), `#log`, `#login-error`, `#login-view`,
  `#phone-view` (`.hidden` toggles), `#ext`, `#pass`, `#dest`,
  `#incoming-from`, `#accept-btn`, `#ice-wrap summary`, `#ice-panel`
  (text contains `ice:`), `.call-card`, `.call-state-text`
  (text `in call`), `.hangup-btn`, `.transfer-btn`, `.transfer-dest`,
  `.transfer-row button`, keypad `button[data-tone]`, and
  `window.__pcs` (Map of live RTCPeerConnections — media proof).
- The event log (`#log`) stays **English** in both UI languages: it is
  operator-facing diagnostics and the runbook greps its phrasings.

## Hard-won knowledge

- CSP at the serving vhost is `default-src 'self'`-strict: no CDN, no
  webfonts, no inline styles/scripts — everything ships same-origin
  from this package. Keep it that way.
- config.js is generated at runtime by the serving PBX (short-lived
  TURN REST credentials); it is deliberately NOT in this package.
  `window.PBX_CONFIG` keys: `sipDomain`, `websocketPath`, `iceServers`,
  `phoneApi`, `contacts` (all optional; see src/app/config.js).
- sip.js is pinned by tarball hash (0.21.2). The 0.x series hangs in
  `userAgent.reconnect()` after transport loss — the bounded watchdog in
  `src/app/connection.js` (5s per attempt, full rebuild on timeout) is
  load-bearing; do not "simplify" it away.
- DTMF must be sent as `application/dtmf-relay` with `Signal=<d>`
  (equals form) — the colon form is 200-OK'd by FreeSWITCH and silently
  dropped.
- Module graph is kept acyclic on purpose: `state.js` (sessions map +
  focused/incoming ids) and `auth.js` (credentials) exist so that
  calls/ice/connection never import each other. Preserve that shape.
- FreeSWITCH executes transfers server-side on the REFER; the browser
  only sends it and parses the NOTIFY sipfrag verdict (`referOnNotify`).
- Formatting: treefmt (prettier) owns everything under `src/`;
  `.buildflow.yml` excludes `src/**` so BuildFlow's oxfmt cannot fight
  prettier (same war the telephony repo fought — pre-settled here).

## Conventions

- One home per fact: README sells, FEATURES inventories status,
  TODO_LIST holds open work, CHANGELOG logs history, this file keeps
  session-durable knowledge.
- Cite stable names (ids, function names, option names), not `file:line`.
- Behavior parity rules the extraction: when porting logic, port it
  verbatim first, refactor in a second, separately-verified commit.
