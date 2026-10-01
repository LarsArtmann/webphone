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
    user database would be a split brain) AND on footprint (2026-09-30:
    the shell-only adoption measured +10.40 MB / +68.2% binary delta —
    see the import-graph lesson below). Security presets are NEVER
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
- **Importing a package links its whole import graph — measure, don't
  assume the linker saves you (2026-09-30 footprint train).** Go links
  the init functions and package-level state of every transitively
  imported package regardless of function reachability: the shell-only
  setup adoption called `NewShell` (never `New`), kept the service path
  unreachable, passed every behavioral gate (httpspec chain parity, full
  suite, smoke 41+4) — and still shipped +72 modules (usermgmt,
  adminui, dashboardui, loginpage, casbin, appkit, datastar) into the
  binary: 15.25 → 25.65 MB. Embedded panel assets and package-level
  registries are exactly the weight pruning cannot touch. The fix was
  going one level down: the lifecycle value lived in
  `httputil.NewServer` (already linked) at +8 KB. Corollary: a
  composition root that "just wires" heavy submodules is a dependency
  on all of them at link time; the measured go/no-go gate in AGENTS is
  the only honest arbiter.

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
- Host nix can die while the STORE stays healthy (2026-09-24/25
  outage: `/run/binfmt` — the directory carrying the emulation
  interpreter symlink that nix stats for emulation detection —
  vanished ~19:10 and every `nix develop`/`nix build` on
  the host failed for 9+ hours with `getting attributes of path
  "/run/binfmt"`; the kernel's binfmt_misc registration was intact,
  only the symlink died. Unprivileged escapes are closed:
  `--option extra-platforms` and `--option sandbox` are restricted for
  untrusted users, `/run` is root-owned. The generation CANNOT
  recreate the symlink — its tmpfiles.d carries no binfmt rules and
  `systemd-binfmt.service` finished OK at the 22:35 reboot without
  creating it, so a service restart heals nothing; the fix plants the
  interpreter symlink by hand: `sudo mkdir -p /run/binfmt && sudo
  ln -s <the -qemu-aarch64-binfmt-P store binary from nix.conf's
  extra-sandbox-paths> /run/binfmt/aarch64-linux`. Durable fix is a
  host-config change (boot.binfmt.emulatedSystems, or drop the
  hand-rolled extra-sandbox-paths); release.sh preflights the
  missing symlink). Fallback that kept verification moving for hours:
  run gates with STORE toolchains directly —
  `/nix/store/*-go-1.27*/bin/go`
  with `GOTOOLCHAIN=local` exported (without it, store go tries to
  fetch its own toolchain and hits the same floor failure), the store
  nodejs for island tests, and a version-stamped smoke binary via
  `go build -ldflags "-X
  github.com/larsartmann/webphone/internal/server.buildVersion=vX.Y.Z"`
  — the ldflags variable is `internal/server.buildVersion` (see
  flake.nix's buildGoModule args); `main.displayVersion` is NOT it.
- Artifact checks need explicit out-links (2026-09-26 near-miss): a
  `nix build --no-link` leaves the PREVIOUS `result` symlink in place —
  a follow-up byte check then reads a STALE artifact (the first
  "aarch64 verification" read an x86_64 binary, e_machine=62; only the
  ELF byte rule caught it). Always pair `nix build … -o <path>` with
  the byte check; never trust a `result` symlink after `--no-link`.
- The auto-commit daemon's PUSH loop stalls silently (2026-09-24/25/26:
  ~45 min to 11 h, both repos, self-healed) — nothing distinguishes a
  slow push from a broken one without checking. End-state contract:
  `git ls-remote` vs HEAD at every phase boundary; the release ritual
  now depends on pushes, so treat an unpushed HEAD as unshipped work.

## Tooling traps

- Dynamic awk/grep patterns must ESCAPE `[`/`]`: release.sh's notes
  extractor matched `^## [2.4.0]` as a dynamic awk regex, where
  `[2.4.0]` is a bracket expression (one char of {2,.,4,0}) — it never
  matched the literal heading and shipped v2.3.0/v2.4.0 with EMPTY
  GitHub release bodies. The fold-check grep passed because ITS
  pattern had shell-escaped `\[`. Found by the 2026-09-20 release
  audit; both objects backfilled from CHANGELOG. The 2026-09-25
  variant: host gawk 5.4.1 WARNs on unknown escapes and DROPS the
  backslash, so the "escaped" pattern collapsed back into the bracket
  class and extracted 0 lines (caught only by idling re-test; a
  comment claiming the 2026-09-20 fix had kept it trusted for hours).
  Final fix: no regex at all — literal `index(line, heading) == 1`
  prefix match, identical in every awk — plus a guard refusing empty
  notes. Two lessons: escape classes in dynamic patterns, and a
  comment claiming a fix is not evidence the fix works.
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
- A gate script that can lie is worse than no gate script
  (2026-10-01, post-v2.8.0 self-review §d1): the battery ran
  `nix flake check 2>&1 | tail -4; echo rc=$?` — that rc is TAIL's,
  not nix's, so a FAILED drill printed "rc=0" and was reported
  green; only buildflow's independent nix step caught the real
  failure. Rule: every scripted gate uses `PIPESTATUS[0]` or the
  repo's documented `set -euo pipefail` pattern — never a bare `$?`
  after a pipe. Grep scripts/ for `| tail` + `$?` pairs before
  trusting any battery output.
- /tmp is not a holding area for verified-uncommitted work
  (2026-10-01): the setup-salvage worktree at /tmp/wp-shell died to
  a tmp cleanup; only luck (the owning session had already landed
  the work) avoided real loss. Park verified-uncommitted work in a
  BRANCH — cheap, greppable, survives reboot.
- Quiet-host windows for timing-sensitive gates (2026-10-01): nix/go
  evaluations running during the stack browser E2E caused a 180s
  timeout + full retry (~8 min) — the runbook documented this EXACT
  self-inflicted pattern from v2.6.0 and it repeated anyway. Two
  flakes across two releases is a pattern. Convention: touch
  /tmp/webphone-e2e-window before E2E/VM gates, have every session's
  nix/go launchers check it, remove after.

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
- The FOUC E2E harness arc (2026-09-23, four stack commits to the
  first green): TWO measurement-model traps, both knowable upfront.
  (1) Soft reloads (`driver.navigate().refresh()`) are served from
  cache and DODGE the URL-level request blocks the scenario relied on
  to force a fresh theme-preload — only
  `Page.reload {ignoreCache: true}` actually re-fetches. (2)
  chromedriver executes NOTHING mid-navigation: any polling loop on
  the driver side is blind exactly during the window the test
  measures; the recorder must be injected to run IN the page
  (`addScriptToEvaluateOnNewDocument` writing tick counters into the
  DOM, counted after load). Two theory-driven fixes (cache-disable,
  ignoreCache) each cost a ~6-min VM run before the instrumentation
  commit moved the diagnosis. Lesson: INSTRUMENT FIRST — a state
  timeline + document truth + server truth dump in the FIRST
  follow-up turns an unfixable-looking flake into a 20-minute fix
  (the PAIR1-DIAG pattern stays in the stack's browser-e2e.py as the
  permanent failure-path diagnostic).

## Provenance moved out of AGENTS.md (2026-10-01 compaction)

AGENTS.md was compacted to a rules-only doc (705 → ~370 lines) because
BuildFlow's `docs/agents-md-size` preflight caps it at 377. The rules
stay there; the dates, hashes, and train narratives moved here.

### Release / train chronology (v2.6.0 and after)

v2.6.0 (signed tag `807ca0c`) released + tail CLOSED 2026-09-23
evening: stack browser E2E ×2 green at `271f5ef` (195s/184s — the FOUC
scenario added ~15s, no budget bump), gh release object published,
smoke 41+4 with `/version` exactly v2.6.0 (use `--bin $(nix build
.#webphone)` locally — a bare `go build` reports Go's pseudo-version),
`nix flake check` green incl. the KVM backup VM, and pbx-artmann relock
#4 + re-pin at `20b2a18` (webphone ExecStart moved 2.5.0→2.6.0;
lock-drift-probe + both toplevels green). The stack rode train
`7197f1c` at that close, was forward-locked to `94ae28d` on 2026-09-24
(upstream vendorHash repair; stack `dea945c` rides it + the flipped
lowercase contacts assert awaiting the v2.7.0 relock). Post-close
reminders: dep bumps swept by the daemon need the vendorHash roundtrip
in the same breath (`0a7a732` repaired a ~2h broken `nix build`), and
the FOUC E2E harness lessons live in the stack repo's browser-e2e.py.

### Setup-bundle footprint GO/NO-GO measurement

The cqrs-htmx `setup` bundle stays REJECTED: (1) split-brain identity —
identity is the PBX extension + directory password proven by the
island's SIP REGISTER, a second user database would be a split brain;
(2) the 2026-09-30 footprint measurement killed the scope-limited
runtime-shell adoption (plan
`docs/planning/2026-09-30_10-37_SUPERB-setup-shell-adoption.html`):
upstream seams shipped first (setup/v4.13.1: `DisableAuth`,
`DisableService`, `NewShell` — ADR-0054 in cqrs-htmx), webphone
composed the full shell behind its httpspec-pinned chain and every test
went green, but importing the setup package links its whole import
graph (+72 modules: usermgmt, adminui, dashboardui, loginpage, casbin,
appkit, datastar) even with `NewShell` — binary 15.25 → 25.65 MB,
+10.40 MB = +68.2% against the recorded gate (≤ +8 MB AND ≤ +20%):
NO-GO. The shell VALUE landed anyway via `httputil.NewServer` (already
a dependency): SSE-safe timeouts (ReadHeader 5s, Idle 60s, no
Read/Write deadlines), 30s graceful-shutdown budget for the SSE drain
— at +8 KB. A zero-usermgmt `shell` submodule upstream would dodge the
import graph; only worth it if the shell grows more value than the
lifecycle.

### flake.nix layout train (2026-09-29 nix-review)

flake.nix is a slim ENTRY (inputs, systems, imports, nixosModules);
the meat lives in `nix/packages.nix`, `nix/checks.nix`,
`nix/module-check.nix` (+ `module-check-base.nix` eval helpers and the
`module-check-csrf.nix` / `module-check-backup.nix` case groups),
`nix/vm-tests.nix`, `nix/apps.nix`, `nix/devshell.nix`,
`nix/treefmt.nix`. The `webphoneVersion` let-binding MUST stay in
flake.nix — `scripts/release.sh` greps/seds it there (`grep
"webphoneVersion = " flake.nix`); it reaches `nix/packages.nix` via a
`{ _module.args.webphoneVersion = ...; }` module. Gotcha: `self` is a
TOP-level-only flake-parts module arg — declare it on the module
function, never inside the `perSystem` pattern. Module hardening:
`backup.destDir` gets the same `/var/lib/` assertion as `dataDir`
(pinned by the `backup-destdir-assertion` linkFarm entry), both units
set `UMask=0077` + `StateDirectoryMode=0750`, and the backup oneshot
orders `after = [ "webphone.service" ]` so a Persistent catch-up at
boot cannot race db creation. The `| tee $out` check pattern is safe:
the locked stdenv setup sets `set -euo pipefail`.

### Second nix-review polish train (2026-10-01)

The module check split into base + csrf + backup case files (each
`checks.webphone-module` linkFarm entry preserved byte-for-byte —
15/15); the stand-in module gained `freeformType = attrsOf anything`
so a new top-level config key the webphone module writes can no longer
break the check; `services.webphone.package` moved to `mkPackageOption`
(still REQUIRED: `default = null`); `devShells.ci` joined the devShell
module — a minimal `mkShellNoCC` (go_1_27 + templ + golangci-lint,
`GOTOOLCHAIN=local`) that `.github/workflows/ci.yml`'s `go-tests` job
enters instead of the heavy interactive shell. The NixOS module was
split under the ~300-line guideline, output PROVEN byte-identical:
`package/nixos-module.nix` holds the config only (213 lines) and
imports `package/options.nix`, `package/caddy-vhost.nix`, and
`package/backup-script.nix`. The `webphoneVersion` hardcode stays an
accepted exception (Nix cannot read git tags without impurity);
`release.sh` owns the tag↔version lockstep.

### erraudit family-adoption train (2026-09-30)

Tier 1 enforced (`--type-aware --disable-extensions`, must exit 0;
package-level sentinels must be the `error` INTERFACE — concrete
`*Error` sentinels trip `sentinel_concrete_type`); tier 2 enforced
green (`--enforce-go-error-family` — 132 stdlib_constructor findings on
2026-09-30 → 0 after converting every seam to go-error-family
constructors with stable dot-notation codes; `--enforce-coded-errors`
also 0). Rules: constructors classify at ORIGIN (P1); propagation over
polymorphic inner errors wraps family-NEUTRALLY — `fmt.Errorf("…: %w")`

- reasoned nolint, a fixed-family Wrap would clobber the inner
  classification (P2; homes: `cmd/webphone.propagatef`, the `"gateway: %w"`
  service wraps, LogErrorContext log-context wraps, gateway form-builder
  inner wraps); sentinels stay `errors.New` vars (P3) and classify via
  `init()` registration in the owning package (store.ErrNotFound/
  ErrListFull, pbx.ErrDisabled/ErrUnauthorized, crm.ErrDisabled/
  ErrUnauthorized/ErrNotFound); functions keep the bare `error` return —
  typed structs only where callers branch (ErrInvalidSend pattern, P6);
  defer-close ignores are standard practice (P7). Tier 3 owner-only full
  audit (never gates; `--no-suppress --enforce-samber-oops
--enforce-generic-return` shows the residue: generic_return decisions,
  40 defer-close ignores, 2 counted-skip swallows, the nolint'd neutral
  wraps). Re-measure tiers 1+2 monthly (next: 2026-10-22). `erraudit
tree` draws hierarchy edges ONLY from package-level declarations and
  dedupes same-named sentinels — the 7 package-level sentinels across
  crm/pbx/store show as 4 rows at max depth 0 BY DESIGN.

### Health-dashboard seam (go-health-dashboard v0.10.1, 2026-10-01)

Config-gated `dashboard.{enable,title}`, DEFAULT OFF — an operator
surface the deployment deliberately exposes (fence it like the probe
triple; the module's Caddy vhost proxies `/health/*` unbuffered like
`/events`). Mounted at `/health` (subtree patterns `/health` AND
`/health/` — Go mux exact-vs-subtree, one without the other 404s half
the surface); probe aliases at `/health/{livez,readyz,startupz}` (SAME
probe instance as the root triple). The dashboard registers in the
container under the name `dashboard` (its pusher-staleness check shows
as non-critical warn). CSP: subtree override via `httputil.Nonce` +
`dashboard.RecommendedCSP(nonce)` — `unsafe-eval` scoped to `/health`
ONLY (Datastar SDK compiles expressions); every other surface keeps the
strict app policy. Its stylesheet is its OWN scoped build
`/assets/health.css` (prettier-formatted, treefmt owns it; regenerated
from the go-health-dashboard + templ-components layout/display/
feedback/utils/datastar sources with `nix run nixpkgs#tailwindcss_4` —
NEVER nixpkgs#tailwindcss v3; the app's `tw.css` stays the
adopted-components set, the two builds never merge). Probe refresh: 1s
background loop ONLY while the dashboard is enabled (the pusher reads
CachedResponse; live mode would freeze at boot); dashboard off =
interval 0. The smoke boots WITH the dashboard on (standing coverage).

### Composition-root train (samber/do v2, 2026-10-01)

ONE container in `internal/app` owns object lifetime; `cmd/webphone`
owns process concerns (config, logging, signals, the HTTP listener).
The injector lives ONLY in `internal/app` (services hold resolved deps,
never the container — service packages stay framework-free; the do
lifecycle-interface conformance is asserted adapter-side in app.go);
the critical pair registers NAMED (`sqlite`, `blob-dir`) because the
names are the probe's critical-service contract; both + the handler are
EAGERLY invoked in `New` (a never-invoked lazy service health-checks as
silently passing — go-health gotcha); the probe is built ONCE and
threaded to both server.Deps and the dashboard (never registered in the
injector it reads — `*health.Probe` conforms to the health-check
interface, self-registration recurses); shutdown order = HTTP drain →
`Probe.Shutdown` → `do.Shutdown` cascade (dashboard pusher,
`Database.Shutdown` — idempotent). `Deps.Probe`/`Deps.Dashboard`
nil-fallback keeps every existing test composition working.
`samber-linter` HW-4 warns the named pair is lazy-registered; suppressed
with a reason at the ProvideNamed sites because the pair IS eagerly
resolved via `MustInvokeNamed` in `New` (the linter cannot see
transitive resolution).

### Dedup trains (2026-09-22 / 09-23 / 09-24)

One-home helpers: `listRows[T]` (store/db.go) owns the query→close→
scan→`rows.Err()` lifecycle; `pbx.do()` is the single disabled-policy
chokepoint (`VerifyCredentials` delegates to `VoicemailSummary`);
`session.makeSession` owns the session birth invariant; server
`requireMultipartTo` is the send-form prologue. The late-night pass
added `domain.must`, `domain.OrClock`, `store.updatedOrNotFound`,
`views.formatFor`, `server.crmNumbers`, `server.applyStatusWebhook`,
`server.recordCallIdem`, `server.contactSaveFailed`,
`server.apiContactSaved`, `views.panelHead`, `views.errorBanner`,
`views.panelError`, `views.identityLine` — all pinned byte-exact in
`views/panels_test.go` (incl. the panelHead both-langs backfill) and
proven by an old-vs-new binary render diff (7/7 partials
byte-identical with error banners + identity lines exercised live).
Each carries its own micro-test. The dedup acceptance registry
([docs/dedup-registry.md](dedup-registry.md)) is the ONE home for every
accepted/declined clone ruling; `-t 3` is the working baseline pending
owner ratification.

### Mic pre-warm seam (2026-10-01)

`mic.js`: `onInvite` starts `getUserMedia` while the phone rings (mic
indicator lights at ring — owner decision: speak ASAP after accept), and
a custom media stream factory passed to
`SIP.Web.defaultSessionDescriptionHandlerFactory` (0.21.2 accepts the
factory argument, verified in the vendored bundle) hands the warm stream
to the session one-shot; every non-answer exit (reject, caller gave up,
logout) releases the device, and a take/release during a PENDING
acquisition stops the late stream. `iceGatheringTimeout: 1000` in the
factory options caps the pre-200 wait (the 0.21.2 default is 5000).
Pinned by `mic.test.mjs` + the connection mic-wiring tests.

### Typography craft train (2026-09-30 font-design)

The CSS root is `font-size: 93.75%` (app.css `html` rule) — a
PERCENTAGE, never px, so the whole rem scale tracks the browser
font-size preference; `.island` mirrors it as `1rem`. The html rule
also owns the rendering baseline (`-webkit-font-smoothing`,
`text-rendering: optimizeLegibility`, `font-synthesis: none` —
no synthetic bold for 550-750 weights), inherited by the island
stylesheet — do NOT duplicate it there. Headings carry `text-wrap:
balance`; `.wp-bubble-body` carries `text-wrap: pretty`; `#log`/`.ice`
use the full local mono stack. `html lang` follows the session
(layout.templ passes `Locale: string(props.Lang)` to `layout.Base` —
library default is a hardcoded "en").

### Shell & accessibility train (2026-10-01 UI/UX)

The shell (shell.js + layout.templ + app.css) owns the command palette,
the "?" shortcut help + Settings cheat-sheet, the skip-to-content link,
the `#wp-tab-skeleton` reveal + panel transition on navigation swaps,
the morph focus-to-heading move, the `#wp-live` live-region
announcements (both new DOM-contract ids), the fixed bottom tab bar
(mobile), `aria-current` on the nav, and the optimistic send bubble with
draft-restore-on-failure. OPERATOR RULING (2026-10-01): contacts depth
(manager search/sections/edit, single-vCard export) is LEDGER's domain
(~/projects/crm) — this app keeps its per-extension personal-contacts
store + `/api/contacts` seam but does NOT grow a contacts manager (an
in-train workstream was reverted, `6989b99`).

### Paperless seam (2026-10-01)

Optional, OFF by default — `paperless.url`+`paperless.token`
both-or-neither build `internal/paperless.Archiver` (go-paperless
v0.4.2) injected as `fax.New`'s archiver; nil-safe everywhere. Inbound
faxes only, fire-and-forget after persist+notify
(`fax.archiveInbound`: re-reads the spooled PDF, 2-min bound, WARN on
failure); metadata ids (tag `fax` / type `Fax` / field
`webphone-fax-id`) lazily ensured per first SUCCESS; duplicate refusal =
inert success; blob store stays the only storage truth.
