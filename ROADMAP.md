# ROADMAP

Long-term direction and raw ideas not yet refined into actionable tasks.
Actionable work lives in TODO_LIST.md; shipped work in FEATURES.md.

## Open ideas

- Union test coverage (`-coverpkg=./...`) in the buildflow test-coverage
  step: per-package attribution hides cross-package coverage (tests in
  `internal/server` exercising `internal/store` count as 0% for the
  store). Blocked on a BuildFlow feature — the canonical
  `.buildflow.yml` key list has no per-step-args mechanism, and a
  repo-local coverage script would duplicate a step BuildFlow already
  orchestrates. Fleet-wide value → belongs upstream.
- sip.js: no 0.22 exists; upstream is dormant at 0.21.2 (an
  unreleased 0.21.3 tag adds only a SimpleUser option). Evaluation:
  `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`. Revisit only on
  a reconnect-hang fix, a security advisory, or a needed capability —
  any future bump re-runs the upstream browser E2E. SDK sweep
  2026-09-19 (SUPERB integration plan): **JsSIP 3.13.8 is the named
  fallback** — actively maintained (npm 2026-05), swap triggers and
  the full research table live in
  `docs/planning/2026-09-19_19-37_SUPERB-island-server-integration.md`;
  never swapped speculatively.
- Health-surface long shots (recorded 2026-09-19, options memo:
  `docs/architecture-understanding/2026-09-19_20-59_health-probes-fleet-options.md`):
  (a) stack-side go-health federation scraping webphone's `/livez`//`/startupz`
  for a fleet health view — zero webphone change; (b) a webphone UI health
  panel ONLY IF the CSP stance changes (the dashboard's Datastar HTML face
  needs `unsafe-eval`; the JSON probes stay the whole adopted surface
  otherwise). Owner calls.
- Server-side telephony inside webphone (Go SIP UA via sipgo,
  pion/webrtc media or B2BUA, FreeSWITCH ESL for originate): researched
  and rejected 2026-09-19 — the browser terminates the media either
  way, so server-side signaling only adds stateful hops to the hottest
  path, and the Go ESL client landscape is thin (top clients are
  Java/Python/Node). If server-initiated calls (click-to-call from
  other systems) ever become a product need, build at the stack/PBX
  layer — it already records calls and bridges webhooks. Research
  table: `docs/planning/2026-09-19_19-37_SUPERB-island-server-integration.md`.
- Stack-side switchover of nix-international-telephony onto this
  service: DONE 2026-09-18 (stack imports `nixosModules.default`,
  nginx vhost proxies the service, browser E2E green after the
  accept/reject wiring fix). Fully resolved 2026-09-19 with the v2.1.0
  release: the stack input rides the v2.1.0 tag commit `d815004`
  (stack commit `2289e89`), and the stack's FULL `nix flake check` is
  green with that lock (browser E2E, webphone VM test included).

## WORTH_CONSIDERING cluster (one-line specs, SUPERB plan P26 2026-09-19)

- Session persistence: move the in-memory TTL session store behind a
  restart-survivable backing (SQLite table + TTL sweep) — the island
  login flow is unchanged; weigh against "sessions are ephemeral by
  design" before building.
- Retention/cleanup job: bounded deletion for old CDR rows, read
  faxes/voicemail blobs, and expired sessions (a `retention_days`
  setting + a systemd timer in the module).
- PWA: RESOLVED 2026-09-22 — the service worker is a NOT-DO (stale
  cached island = a bug class invisible to every gate; offline is
  impossible for a phone: `docs/planning/2026-09-22_17-05_pwa-spike-verdict.md`).
  Parked behind an owner demand signal: manifest-LITE only (manifest
  - maskable PNG icons, NO fetch interception, zero staleness risk).
- Video calls: SIP.js video negotiation + a `<video>` call card —
  FreeSWITCH side needs a video-capable profile; large surface, only
  on demand.
- Recording UI (see raw ideas below): product-intent decision first
  (consent/jurisdiction), then the panel.
- Backup snapshot retention: optional `backup.retentionDays` pruning
  old snapshots (the module skeleton deliberately leaves retention to
  operator tooling; self-contained only if wanted) — 01:04 report
  §f/33, distinct from the blob/CDR retention idea above.
- Startup probe wiring: gate the systemd unit's `Type=notify`/health
  on `/startupz` semantics (document the contract first; the unit
  currently starts and stays up regardless) — §f/35.
- nginx gzip for text assets (app.css/shell.js/htmx bundles): micro
  win, one module option — §f/36.

## Standing watches (SUPERB plan P27 2026-09-19)

Drift gets caught by routine, not luck — each row names the trigger to
re-check:

- sip.js: revisit on a 0.22 release, a reconnect-hang fix, or a
  security advisory (evaluation: `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`).
- templ-components: ship a ThemeScript opt-out knob and this repo drops
  the CSP hash pin plus the app.css `!important` color-scheme rules.
- oxlint globals watchlist: any new browser global in the island needs
  an entry in `internal/web/assets/island/oxlint.json` (the gate fails
  closed on undeclared identifiers by design).
- E2E wall-time budget: RE-BASELINED 2026-09-22 on the v2.5.0 release
  chain — two forced-rebuild runs measured 384s/373s (the suite carries
  the restart-resume, transfer and FS-outage drills; the old ~150s
  baseline predates them). Budget: 445s (max + ~15%); a run above it is
  a perf regression signal, not noise — watch TWO consecutive
  over-budget runs before investigating.
- CSP re-audit trigger: assets are same-origin by policy (CDN banned);
  if that stance ever changes, re-audit CSP against every moved
  script (idiomorph included) before shipping — 01:04 report §f/46.

## cqrs-htmx adoption long tail (plan P5-P7, 2026-09-18)

Source: `docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`
§Phases 5-7. Status 2026-09-20: P5-P6 shipped in full — OOB spike
(verdict: PARKED), toasts, `/version`, Server-Timing, `/openapi.json`,
`cqrshtmx.Chain` parity, the fuzz/OR1/RA1/TT1/DR1/VL1 hardening set
(the plan table carries every commit hash) — and the MD1 root-tag bump
trigger fired and was executed: cqrs-htmx v4.11.0's one wire change
(the SSE `retry:` hint) is stream-test-pinned, `BenchmarkHubFanOut`
re-ran clean (`docs/reviews/2026-09-18_hub-fanout-baseline.md`), and
the stack browser E2E passed on the bumped tree. What remains:

- OOB badge push (OO1-OO3 → M54/UB1): PARKED by verdict
  `docs/reviews/2026-09-18_oob-badge-spike-verdict.md` — weak demand
  (the badge already refreshes via TTL cache + drop-on-mutation
  invalidation), payload-contract risk, and the E2E-gate cost.
  Adoption only via that note's criteria: flag-gated prototype
  (`WEBPHONE_SSE_OOB=1`) behind a green stack browser E2E, with the
  `cqrshtmx.OOBHTML` signature re-checked against the then-current tag.
- Island JS test runner: LANDED 2026-09-20 (node:test + minimal DOM
  stubs under Nix — `internal/web/assets/island-tests/`, flake check
  `island-js`). First ports: toast rendering (announce kind/cap) and
  i18n en/de key parity. Still grep-only until touched: live pill,
  429 surfacing in the phone-api wrappers, dtmf-relay shape — port
  each into a DOM test when the code next changes.
- Conditional P7 decision records: `StructuredError` for
  `/api/session` (adopt only if the island branches on codes);
  ClientIP-trust note upstream in httputil (only if the stack proves
  XFF sanitized). The periodic vulnix rescan left the watchlist —
  `nix run .#vulnix` rides the release.sh gates every train since
  2026-09-20. DR1 (transitive-drift check as a buildflow step) is
  NOT-DO 2026-09-20: a drift check needs the network and is
  informational-only — informational checks do not belong in hard
  gates, the vulnix train gate covers the supply-side security angle,
  and version-drift only matters AT bump time when MD-style triggers
  fire. Revisit only if an unattended drift surprises a train.

## Open questions (owner calls)

- Rate-limit keys: does the stack's proxy sanitize `X-Forwarded-For`?
  Gates the flip from port-stripped peer-host keys
  (`remoteHostKey`) to `KeyExtractorFromClientIP`; standing since the
  2026-09-18 adoption (safe default shipped, flip rule documented).
- Stack `webphone` input policy: DECIDED 2026-09-20 — ride webphone
  `main` with a per-train lock bump (rationale + revisit trigger in
  AGENTS "Owner decisions"); no longer an owner call.
- GitHub Release objects per tag (notes/visibility) vs tags +
  CHANGELOG only (v2.0.0 has a tag, no Release object).
- Go module-path policy for v2+ tags: `/v2` suffix vs NOT-DO record
  (a flake-consumed app never `go get`s itself).
- Loopback gateway semantics: simulate `delivered` (like fax loopback
  resolves `transmitted`) or keep `sent` as the honest terminal state?
- Handler gating: keep the deliberate dual layer (`requireSession`
  helpers + `Sessions.Require` on `/events`, `/phone-api/`) or drop
  the middleware wiring?
- Recordings in the webphone UI (see raw ideas below) — product
  intent, consent/jurisdiction posture, per-extension vs operator
  access.
- Stack-side: keep or revert the TEMP-DIAG answer-phase dump in the
  stack's browser E2E (commit `b96d4c2` there).
- HSTS on prod (owner call): the `nginx.hsts` option ships opt-in —
  decide for `pbx.artmann.tech` once https-only is proven
  (01:04 report §f/21).
- Release cadence / pin policy (g2, owner call; recommendation
  recorded 2026-09-20): cut a train when `[Unreleased]` accumulates a
  user-visible theme (the v2.4.0 fold pattern: fix + feature + story),
  deploy once per train via the single owner command, and keep the
  stack riding webphone `main` until a hotfix cadence actually
  emerges — switching to tag pins buys reproducibility at the cost of
  a manual bump step on every fix. Revisit when a security fix ever
  needs to ship inside an hour.
- Infra ask (upstream of this repo): teach the auto-commit daemon to
  EXCLUDE `docs/status/` and `docs/planning/` (or only sweep on
  quiescence) — its heuristic commits have twice swept half-written
  reports (2026-09-22 12:57: 11-file sweep) and reintroduced
  formatting drift in the stack (operator.js, fixed in `1a95a73`
  there). Until then: the runbook's narrative-commit-at-phase-boundary
  line is the mitigation.

## Harvested raw ideas (2026-09-19 docs-health sweep)

Grouped from the annotated 2026-09-18/19 status reports; nothing here
is committed work — refine into TODO_LIST only on demand.

- Recording integration (PBX records everything today; webphone has no
  surface): Recordings panel/in-island playback, live REC indicator,
  per-extension access via the phone-api proxy pattern, `*97` opt-out
  visibility, recorded-badge on CDR rows, retention surfacing,
  stereo-WAV compression job, consent/privacy README note.
- Browser-level gates: headless-browser console-cleanliness check on
  `/` (the gate hole that let three CSP violations ship), tab-SSE
  browser test, two-browser E2E under CPU-constrained TCG.
- Testing long tail: SSE handler edge tests (anonymous 401, heartbeat
  on the wire), store/domain edge-case round-out (contacts CRUD,
  branded-ID property round-trips), `-race` stress of SSE hubs,
  delivery-receipt SSE-event assertion, comment-vs-code contract
  sweep, provider error-text pinning, periodic `art-dupl` ritual.
- Messaging polish: persist + display failure reasons for failed
  messages (parity with fax), distinct delivered badge style (see
  TODO_LIST), random loopback provider refs (drop UnixNano), image
  thumbnails, draft persistence, fax cover pages.
- Calls/UX: shortcut help overlay, new-conversation UX (open the
  thread after first send), a11y pass (focus order post-swap, aria-live
  SSE regions), multiple-tab glare warning, SSE connection-loss banner
  beyond the live pill.
- Platform: short-lived TURN REST credentials via `/config.js`,
  metrics endpoint, SQLite backup/restore runbook + drill,
  per-extension data export, webhook payload versioning header,
  timezone-aware timestamps, server-side PDF page counting, MIME
  sniffing on attachments, rate-limit tuning knobs (only on operator
  demand), CSP nonce mode, `/favicon.ico` route for non-browser
  clients, reusable "HTMX tabs + island + session" pattern doc
  (third LarsArtmann app with this shape), i18n dynamic-template
  key-sync, island/app.css shared-token extraction, webhook-idempotency
  durability decision (memory TTL means a post-restart provider replay
  re-applies — acceptable?), OpenAPI boundary decision (extend beyond
  `/api/session` or record as deliberate), limiter-key widening runbook
  line (NAT offices), signed tags (`git tag -s`), nixpkgs lock + vulnix
  rescan cadence, daemon-config exclusion of `docs/status/` from
  heuristic commits (upstream infra decision), wrapper-flake bisect
  trick write-up (module-from-HEAD + package-from-rev).

## Local Playwright island E2E (consciously deferred, plan T27 2026-09-20)

A local (no-PBX) Playwright harness — login → tab click → force 401 →
toast assertion, plus golden screenshots of the four toast kinds — was
scoped in the 2026-09-20 testing plan and CONSCIOUSLY ROADMAP'd instead
of shipped: a chromium-in-devShell costs a 1-2 GB closure for what the
stack browser E2E already proves with a real PBX, and golden-image tests
flake on font/antialiasing drift. Trigger to revisit: the stack E2E
becomes too slow for inner-loop island work, or toast styling regressions
actually escape (they did not in the 2026-09-20 train — the island
node:test suite plus the E2E caught everything).

## Test-infra follow-ups (from the 2026-09-20 testing train)

- Long-tail coverage: `internal/blob`, `internal/fax`,
  `internal/messaging` still have no direct tests (exercised only
  through the server suite); `views` stays transitive BY DECISION (see
  `docs/reviews/2026-09-20_coverage-baseline.md`).
- The fuzz target (`FuzzContactsAPISave`) runs its seed corpus in CI;
  a scheduled longer `-fuzztime` run (and more targets: message send
  bodies, vcard parser) is unstarted.
- Shell copy stays English (decision D3); if a per-extension UX demand
  emerges, the shell toast strings move to a wp-lang lookup with en/de
  tables.

## Session persistence (raw idea, owner call)

Sliding-session TTL refresh (extend expiry on activity) with CSRF token
rotation at each refresh point, riding the island's existing adoption path
(`GET /api/csrf`). RESOLVED 2026-09-22: sliding sessions SHIPPED (2.5.0,
7d idle + 30d absolute); CSRF-rotation-on-slide is a recorded NOT-DO —
verdict with the threat model at
`docs/planning/2026-09-22_14-45_csrf-rotation-on-slide-verdict.md`
(double-submit makes token-only leaks inert; both-halves theft is bounded
by the absolute cap, not by rotation).
