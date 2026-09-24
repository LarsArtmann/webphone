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
  `docs/planning/archived/2026-09-19_19-37_SUPERB-island-server-integration.md`;
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
  table: `docs/planning/archived/2026-09-19_19-37_SUPERB-island-server-integration.md`.
- Stack-side switchover of nix-international-telephony onto this
  service: DONE 2026-09-18 (stack imports `nixosModules.default`,
  nginx vhost proxies the service, browser E2E green after the
  accept/reject wiring fix). Fully resolved 2026-09-19 with the v2.1.0
  release: the stack input rides the v2.1.0 tag commit `d815004`
  (stack commit `2289e89`), and the stack's FULL `nix flake check` is
  green with that lock (browser E2E, webphone VM test included).

## WORTH_CONSIDERING cluster (one-line specs, SUPERB plan P26 2026-09-19)

- Session persistence: RESOLVED — SQLite-backed store shipped (2.5.0:
  sessions survive restarts) plus sliding TTL (7d idle / 30d absolute);
  the spike verdict lives at
  `docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md`.
- Retention/cleanup job: SHIPPED — `retention_days` (T25) deletes
  messages+attachments, faxes+documents and emptied threads on a daily
  sweep; sessions were already covered by the store's own expiry sweep.
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
- Backup snapshot retention: SHIPPED — `backup.retentionDays` (T18a)
  writes dated `snapshots/<date>/` history (hardlink basis) and prunes
  older than N days; VM-test-proven.
- Startup probe wiring: RESOLVED as a documented NOT-DO (T18c) — the
  unit deliberately stays `Type=simple`; `/startupz` is the readiness
  truth (README "Readiness vs systemd").
- nginx gzip for text assets: SHIPPED — `nginx.gzip.enable` module
  option (T27a).

## Standing watches (SUPERB plan P27 2026-09-19)

Drift gets caught by routine, not luck — each row names the trigger to
re-check:

- sip.js: revisit on a 0.22 release, a reconnect-hang fix, or a
  security advisory (evaluation: `docs/reviews/2026-09-18_sip-js-0.22-evaluation.md`).
- templ-components ThemeScript knob: DONE 2026-09-22 — shipped upstream
  as `PageProps.NoThemeScript` (v1.19.2); this repo consumes it and
  dropped the CSP hash pin plus the app.css `!important` color-scheme
  rules.
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

Source: `docs/planning/archived/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`
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
  `island-js`). Ports since: toasts, i18n parity, session.js feedback
  map, connection watchdog scenarios (incl. the live pill states),
  composer behaviors, calls state chip, shell error surfacing. Still
  grep-only until touched: dtmf-relay INFO shape — port into a DOM
  test when the code next changes.
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
- Release.sh load handling (from the v2.6.0 release night, two
  load-shaped E2E stalls): SHIPPED 2026-09-23 as the conservative
  default — `release.sh` `load_gate()` refuses at 1-min loadavg ≥8
  (`WEBPHONE_RELEASE_MAX_LOAD` override) plus `assert_clean_tree()`
  at tag time (`a24496a`). Remaining owner call: ratify, or switch to
  a single E2E auto-retry (+10 min per true failure, masks real
  regressions)? g3 CLOSED 2026-09-23 evening: with the FOUC scenario
  aboard the stack E2E ran green twice at 195s/184s against the 445s
  budget — no bump. (02:47 §g1, 15:41 §g)
- One-time force-push ratification (15:41 §g1): during the v2.6.0
  fold a `--force-with-lease` was used ONCE on a self-authored
  commit ~1 minute after pushing it (daemon raced `git add`). The
  runbook forbids force-push; ratify the exception shape (own commit,
  seconds old, lease-protected) or forbid outright.
- Stack pin vs tag on release trains: the v2.6.0-era stack lock pins
  post-tag main (`7197f1c` — carrying the CRM wire-contract test and
  the pre-surgery state; the stack's own surgery is `be876ae`) while
  the tag pins `807ca0c`. Matches the documented "ride main" decision
  and the v2.5.0 precedent (lock ≠ tag), but re-pinning the stack to
  the tag itself for release trains is the alternative. (02:47 §g2)
- Art-dupl ritual baseline: ratify `art-dupl 0.7.0 --sort
  total-tokens -t 3 --type-aware` as THE dedup baseline (`-t 1` is
  forensic-only — 43-45 shown groups of mostly idiom noise) and the
  `-t 3` sweep's triage (1 extract / 3 accept), and document what the
  0.7.0 "filtered suppressed" bucket hides before crowning it.
  (03:01 §g1, 01:12 §g3)
- Webhook 400 contract + idem key namespace (from the dedup trains):
  does any stack-side tooling/runbook grep the `/hooks/*/status` 400
  body texts or depend on the OLD validation precedence
  (provider_ref-before-status; both hooks now validate status first)?
  And is the in-memory idem key rename `msg/<ref>` → `message/<ref>`
  acceptable given no persistence? (01:12 §g1/g2, 03:01 §g2)
- Helper micro-test bar: must a one-home helper land ONLY with a
  micro-test pinning its contract (would have blocked
  `apiContactSaved`), or is suite-level coverage acceptable with
  micro-tests batched later? Decides whether the TODO row is a rule
  or a task. (03:01 §g3)
- Missed-call semantics: does a deliberately REJECTED incoming call
  count as missed (current: no — REJECT never dispatches
  `wp:call-missed`), and should the badge also clear when the
  island's _Recent calls_ panel opens (current: History tab click
  only)? (21:43 §g2)
- Search `?q=` URL semantics: should a typed query push `?q=` into
  browser history (deep-linkable, back-button unwinds filters) or
  stay ephemeral as implemented? (21:43 §g1)
- Store `Must*` panic policy: store scans call
  `MustParsePhone`/`MustContactID`/`MustParseExtension` on raw DB
  strings — a corrupted/hand-edited row panics the server at query
  time. Panic-on-corrupt deliberate (data only ever written through
  validated paths) or degrade to an error? One decision covers all
  `Must` call sites in scans. (17:40 §g3)
- Concurrent-session breakage policy: when another session's
  committed-but-broken code blocks the shared gate, fix-and-commit
  immediately (unblock everyone, risk colliding with their next edit)
  or keep the hands-off rule and report only? (17:40 §g1 — HEAD went
  red twice on 2026-09-22 under this policy)
- Infra ask (upstream of this repo): teach the auto-commit daemon to
  EXCLUDE `docs/status/` and `docs/planning/` (or only sweep on
  quiescence) — its heuristic commits have twice swept half-written
  reports (2026-09-22 12:57: 11-file sweep) and reintroduced
  formatting drift in the stack (operator.js, fixed in `1a95a73`
  there). Grew two sharper variants from the v2.6.0 night: a
  PRE-SWEEP BUILD GATE (never commit a non-compiling tree — three
  sessions hit committed-broken main on 2026-09-22) and asserting a
  clean tree at each release.sh step boundary (the daemon swept
  mid-release between tag push and lychee). Until then: the runbook's
  narrative-commit-at-phase-boundary line is the mitigation.
- Cross-session gate protocol (from the v2.6.0 release night, load
  60-174 from parallel agent sessions): a flock convention
  (`/tmp/webphone-gates.lock` holding pid + scope) so concurrent
  sessions yield instead of colliding, staggered schedules, or
  "sustained quiet" as the accepted de-facto rule — owner call.
- `nix flake check --all-systems`: RESOLVED as a NOT-DO (runbook §8,
  verified 2026-09-19) — for other systems it is evaluation-only
  (zero derivations built, gates nothing); the aarch64 gate is the
  explicit cross-build of the package + the checks that matter, with
  the ELF machine-bytes assert. The 21:43 buildflow warning is the
  reminder, not an open question.

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
- Platform: short-lived TURN REST credentials via `/config.js` (T26b,
  TODO row), metrics endpoint (SHIPPED — `/metrics`, T26a), SQLite
  backup/restore runbook + drill (SHIPPED),
  per-extension data export (T26c zip: messages JSON, contacts vCard,
  fax list), webhook payload versioning header,
  timezone-aware timestamps (SHIPPED — `timezone` key, T26d),
  server-side PDF page counting, MIME
  sniffing on attachments (SHIPPED — T26e), rate-limit tuning knobs
  (only on operator demand), CSP nonce mode (moot while zero inline
  scripts holds), `/favicon.ico` route (SHIPPED — T27b), reusable
  "HTMX tabs + island + session" pattern doc
  (third LarsArtmann app with this shape), i18n dynamic-template
  key-sync (SHIPPED — referenced-keys guard, T27d), island/app.css
  shared-token extraction, signed tags (SHIPPED — `git tag -s` in
  release.sh, T27c), nixpkgs lock + vulnix rescan cadence (rides
  release.sh every train), daemon-config exclusion of `docs/status/`
  from heuristic commits (upstream infra decision), wrapper-flake
  bisect trick write-up (module-from-HEAD + package-from-rev),
  vendored-sip.js cache headers + build-time integrity pin (esbuild
  output hash checked so `./update.sh` breakage surfaces at build),
  CSP report-only companion or report-uri endpoint, `/assets/*`
  unknown-path 404 parity with the styled 404, smoke `--expect-csp`
  assertion.

## Composer/UX raw ideas (2026-09-22 trains, unshipped)

From the send-failure and composer train brainstorms. The first six
ideas SHIPPED 2026-09-22/23 in v2.6.0 (dial typeahead, jump-to-
latest chip, missed-call badge, thread search, absolute-time-on-
hover, audio output picker — see FEATURES); what remains here is
unshipped fuel — refine into TODO_LIST only on demand.

- Peer hub: contact → thread + history + voicemail in one view.
- Tailwind v4 scoped-layer coexistence spike for templ-components
  (designed in the 2026-09-22 deep-dive report; owner call pending —
  one-component proof before any adoption).
- Unicode-insensitive thread search: SQLite `LIKE` folds ASCII only
  ("MÜNCHEN" does not match "münchen") — needs `lower()` collation
  or an FTS5 column; a real design decision, not a patch (small ADR;
  21:43 report §e4).
- Search depth follow-ons: clear-button affordance, result count +
  active-filter chip, deep-link return with the query preserved,
  CRM display-name matching (today the query matches raw numbers
  only), debounce indicator (21:43 §e5/§f27-28).
- Typeahead follow-ons: also search thread remotes (contacts only
  today), recent-calls recency boost in `rankContacts`.
- Jump-chip coalescing of rapid pushes; hover `title`s for the nav
  badge and voicemail rows too.
- Missed-badge / jump-chip localization — only if the owner ever
  overturns the language-neutral-shell decision (D3).
- Voicemail transcription (demand-gated; verify the phone API even
  exposes it first).
- `ValidOutboundStatus` extraction: stays un-built UNLESS the
  webhook-valid set and the service-apply set ever diverge (two
  deliberate distinct contracts today).

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
