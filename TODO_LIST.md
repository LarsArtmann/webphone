# TODO_LIST

Short- and mid-term improvement tasks. Done work is DELETED, never struck
through (docs-health: shipped work moves to CHANGELOG / FEATURES — one home
per fact). Long-shot ideas and owner calls live in ROADMAP.md; every row
cites its source report in Evidence.

Last sweep: 2026-10-04 (passkey train-tail session). Removed: the
AGENTS-over-cap row (trimmed 419→129 lines, dense-line style; passkey
detail moved to README § Passkey sign-in; jq lock-read guard added);
the passkey tail's code-polish items (2)–(8) (enrollFailed(0) copy,
plan-doc P/F annotation, csrf.js single home, userauth /healthz leg,
tier-2 pins, whoamiLine→ui.js + typed sessionIdentity, aarch64 eval —
all executed, gates green); the boot-contract runbook patch (applied to
the stack ops-runbook as `9a21893`, boot-surface + passkey rows; the
cancelled sibling CI run for `890a526` re-dispatched). Prior sweep
2026-10-02 (docs-health AUDIT v6).

## Open tasks

### Passkey train tail: deploy proof (owner terminal)

**Status:** 🟡 `IN_PROGRESS` (code polish DONE 2026-10-04; owner legs remain) · **Priority:** High · **Effort:** S

The code train AND the 3-repo deploy train are DONE (webphone `400eaff`
pushed; sibling `890a526`; pbx-artmann staged `0ngvm4q7…`, probe OK,
gates green — see CHANGELOG Unreleased + the 17-28 report); the code-polish tail landed 2026-10-04 (enrollFailed(0)→enrollNetFailed,
csrf.js single home, userauth /healthz leg, tier-2 pins, whoamiLine→ui.js,
typed sessionIdentity, plan-doc annotation, AGENTS trim — see CHANGELOG).
Remaining: the OWNER switch + post-switch ritual (rotate
`/tmp/pbx-toplevel-current`, record the fresh diff-closures baseline)
and the LIVE passkey proof with the fail-closed password-file drill
(pbx-artmann TODO §1 row). OWNER CALLS outstanding: ratify
runbook-only enroll + Lars-only v1 mapping; switch-now-vs-CI-first
(CI for `890a526` was re-dispatched 2026-10-04 after the ~1 h runner
cancellation — check its verdict); installer release republish timing
(stale since the relock).

**Evidence:** status `docs/status/2026-10-04_17-28_passkey-train-resumed-session-brutal-status.md` (§f harvest source); plan `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`

### Boot-contract tail: D3 retry-loop owner call + stack runbook patch

**Status:** 🟡 `PLANNED` (owner call + stack dispatch) · **Priority:** Medium · **Effort:** S

The boot-error contract SHIPPED 2026-10-02 (5-part render on every boot failure, exit 1 designed / 2 panic, EN-only, class-tagged; `App.Start` de-panicked; arch test confines samber/do; smoke `boot failure scenario` green). The stack runbook patch APPLIED 2026-10-04 (stack `9a21893`: boot-surface + passkey error-contract rows; the parked patch doc is spent). Remaining: (1) OWNER CALL (D3 follow-up) — cap systemd's boot-failure retry loop (`StartLimitBurst`/`StartLimitIntervalSec`) or ratify the 5s `Restart=on-failure` retry as desirable liveness (module deliberately unchanged this train); (2) re-grade the boot surfaces at the 2026-10-22 erraudit re-measure

**Evidence:** plan `docs/planning/2026-10-02_10-23_SUPERB-operator-boot-contract.md`; ruling `docs/error-contract.md` § "Boot surface"; pins `cmd/webphone/bootreport_test.go`; runbook patch = stack `nix-international-telephony@9a21893`

### v2.8.0 deploy tail

**Status:** 🟡 `PARTIALLY DONE` (owner terminal remains) · **Priority:** High · **Effort:** S-M

v2.8.0 deploy tail (OWNER terminal — pbx-artmann AGENTS forbids assistant ssh/deploy): the RELEASE SIDE IS DONE 2026-10-01 — v2.8.0 signed+pushed (number decided autonomously: 2.8.0 so the bridge ">= 2.8" claims stay true; ships the never-tagged 2.7.0 content + nix-review train + honest Content-Type + error-family + caddy module rename). REMAINING (all owner): deploy (stack lock bump → stack gates incl. the browser E2E the EmptyState markup change owes → aarch64 → pbx-artmann relock #5 + re-pin → deploy; COMMAND: `cd ~/projects/pbx-artmann && nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech` → smoke → `nixos-rebuild switch`), then post-deploy `python3 scripts/webphone-smoke.py --base https://pbx.artmann.tech --expect-version <V>`. NOTE: train C refuses self-sends LOCALLY — expect the 422 "messages cannot be sent to your own number" + failed row instead of the provider's 40310 text; the bridge's declared-type preference reaches prod only after the stack relock that rides this tail. If the release number is NOT 2.8.0, reword the bridge-side ">= 2.8" doc claims there to commit/date form (webphone's own AGENTS line is already commit-pinned at `e6ea2c7`). AT THE STACK RELOCK (same train — both stack lines reference the REMOVED nginx option and eval-error against the new rev): `modules/telephony/web.nix` `nginx.enable = lib.mkForce false;` → `caddy.enable = lib.mkForce false;` (comment wording too) and `modules/telephony/default.nix` vhost split-brain assertion `config.services.webphone.nginx.enable` → `caddy.enable` + message wording (webphone caddy rename, 2026-09-30)

**Evidence:** tag `v2.8.0` pushed + gh object published 2026-10-01; stack `3afcf57` (lock bump + caddy guard, E2E×1 green after one self-inflicted load flake + VM + flake check green); aarch64 ELF `b700`; lychee 0 after repointing the never-tagged v2.7.0 links; number = 2.8.0 (bridge ">= 2.8" claims true; ratify at the sitting — briefing row 15)

### Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error)

**Status:** 🔴 `TODO` (owner journal leg) · **Priority:** High · **Effort:** S

Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error) when probed 2026-09-19 — webphone-side classification is 502 + provider-rejection detail since `1d53f44`, refusals answer 422 since `6ac8962` (`1d53f44` live on prod since the 2.6.0 deploy; `6ac8962` rides the untagged release), but the ROOT cause is stack-side (telnyx-webhooks bridge down/rejecting or Telnyx creds). OWNER (assistant ssh is forbidden by pbx-artmann AGENTS): journal the telnyx-webhooks unit, grep for sms/422/error lines, restart or fix creds per findings, send a test SMS; record the root cause here + in the stack runbook. Also still unproven live: the 2.6.0 post-deploy sign-in rejection-banner (self-send) browser check — narrows to a journal-only confirmation if a human test SMS worked since 2026-09-25

**Evidence:** 2026-09-19 probe; webphone-side fix `1d53f44`; refusal-422 `6ac8962`; OWNER TRIAGE PACK ready 2026-10-01: command sheet §4 decision tree (unit status → journal grep → the bridge's own actionable cred strings → restart → test SMS; self-send 422 = expected, not a failure)

### Scheduled sends (M21.6): NO-GO on the current gateway seam

**Status:** ⛔ `DECIDED-AGAINST` (gateway-seam dependency; revisit only with a deferred-send seam) · **Priority:** Low · **Effort:** M

Deferred/scheduled message sending was cut from T18 (M21.6): the gateway seam has NO deferred-send concept (loopback sends immediately; the webhook bridge is fire-and-forget to the provider), so a scheduled-send UI would show a state the transport cannot honor — violating the island-honesty contract. Revisit only if the provider/stack grows a deferred-send lane; the snippets seam (M21) already covers the "reply faster" problem it was meant to solve.

**Evidence:** design note `docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md` (M21.6 cut); owner ratification pending (2026-10-02 §g).

### Gateway honest-Content-Type follow-ups

**Status:** 🟡 `PLANNED` (stack/pbx side only now) · **Priority:** Medium · **Effort:** S

Gateway honest-Content-Type follow-ups (webphone side of `docs/status/archived/2026-09-29_04-32_*` §f): byte-exact part-HEADER-block golden (current golden greps fields individually; a full header-block pin catches reordering) · document the webphone×bridge compat matrix next to the AGENTS gateway-seam bullet (old binary → sniff lane; new → declared lane) · consider a webhook-mode smoke probe (loopback-only today; the webhook lane has zero black-box coverage in this repo). Stack/pbx side (NOT this repo): `ftypqt`→`video/quicktime` sniff fix, stack E2E MMS-outbound coverage, pbx-artmann FEATURES:87 stale sniff text — route via the next stack session/relock

**Evidence:** webphone side DONE `0aab677`: full-body golden `TestProviderFormPartHeaderBlockGolden` (field order, header order, CRLF framing), AGENTS compat matrix, webhook smoke probe DECIDED-AGAINST (golden + bridge contract tests own the shape). Still routed via next stack session: `ftypqt` sniff fix, E2E MMS-outbound, pbx-artmann FEATURES:87 text

### Tooling hygiene batch: markdown-lint posture decision

**Status:** 🟡 `PARTIALLY DONE` 2026-10-01 · **Priority:** Low · **Effort:** S

Tooling hygiene batch: markdown-lint posture decision (5092 detect-only corpus findings — configure markdownlint to house style or record the detect-only posture in AGENTS so nobody "fixes" the corpus by reflowing) · codespell policy for the 2 `pre-emptive` warnings in archived status snapshots (exclude `docs/status/**` like the stack, or fix the words) · reconcile AGENTS buildflow-full claim vs observed skips (gitleaks/codespell/markdown-lint "skipped by build mode 'full'" on 2026-09-26 — AGENTS or `.buildflow.yml` is wrong) · BuildFlow binary freshness advisory (stale binary skews verdicts — rebuild in the BuildFlow repo) · commit the reusable render-diff script (partials-only, webhook-502 error instrument — mechanics in the archived 16:56 report §b) + park its recipe in AGENTS/lessons (carried since the 09-24 18:05 report)

**Evidence:** DONE: codespell real-binary + `.codespellrc` (vendor/go.sum skipped, pre-emptive ignored), AGENTS buildflow-claim reconcile, render-diff script committed (`scripts/render-diff.py`; LIVE-VERIFIED 2026-10-01: 7/7 partials byte-identical + 7/7 error arm), BuildFlow freshness noted via doctor. v6 stragglers DONE 2026-10-04: health-css fmt appended (live-verified: raw rebuild → in-script fmt → byte-identical with HEAD), smoke mypy 8→0 (annotations were the lie; ruff clean), "error families are total" + 104-code registry table in error-contract.md generated and freshness-pinned by `TestErrorCodeRegistryIsFresh` (negative-tested; fixes the SkipAll-walk latent bug in the sibling arch test too). REMAINING (owner): markdownlint posture (briefing row 18). Was: 09-26 §d3/§f11-12; 09-24 12:41 §f19-20; 04:32 §f14-15

### OWNER-calls batch session

**Status:** 🟡 `PLANNED` (briefing ready, sitting owed) · **Priority:** High · **Effort:** S

OWNER-calls batch session (decisions, one sitting — briefing at `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`): the original ~15 + CRM policy trio + self-send train-C semantics + templ-components history-blemish + tail items g1 force-push ratification / release.sh load-gate default / g2 KVM-timeout / art-dupl `-t 3` baseline (FIFTH surfacing) + registry ratification (`docs/dedup-registry.md` as THE home) + `settingsRow` build-or-retire (FOURTH surfacing) + suppression-bucket doc + dedup suppression-SCOPE ruling (shown-only vs periodic suppressed-set audit; 12:59 report g3, unrouted until v6) + webhook 400 body-text dependents + `msg/`→`message/` idem-key rename + helper micro-test bar + missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy. NEW since 2026-10-02 (docs-health v6 sweep): ratify "routed-as-resolved → archive" for the four live owner-gated plans (error-excellence, 20-year-durability, 04-29 pareto, 13-33 stack-adoption — v4's g1); ratify train-level `v` verdict markers as the standing annotation hash bar (v3's g3, six sweeps running); the check-rows marker-cell baseline for the 37 pre-2026-09-18 archived files (restyle or accept). NEW since 2026-09-26: push-lag threshold policy (when is a silent daemon push stall BROKEN — 10 min? 1 h) + session push ratification (phase-boundary hand-pushes) · is the next webphone release 2.8.0 (making the bridge's ">= 2.8" claims true) or reword them · `gateway.attachment_limit` knob vs bridge-422-teaches design as final · sniff-fallback lifespan (keep forever vs delete-after-deploy-confirmed) · tier-2 family-adoption intent (dep swept-but-unused?) · report erraudit `tree` same-name-dedupe upstream? (owner's own tool) · AGENTS restructure permission · existing-prod-data chmod/re-backup for the UMask tightening · destDir nesting legality · stack verification now-vs-train-close

**Evidence:** briefing updated 2026-10-01 to 28 rows (15–28: v2.8.0 + setup NO-GO ratifications, compaction permission, tooling postures, push-lag, QMD, dedup baseline, gateway micro-decisions) + a closed-since section (go-cqrs-lite wontfix etc.)

### Post the release announcements

**Status:** 🟡 `PLANNED` (owner posts) · **Priority:** Low · **Effort:** S

Post the release announcements (drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 at `docs/announcements/`; owner picks channel(s), approves wording, decides the disclosure posture — fix-acknowledged, no exploit detail)

**Evidence:** v2.8.0 drafts A/B/C = `docs/announcements/2026-09-30_v2-8-0_drafts.md` (covers the merged 2.7.0+2.8.0 story); v2.7.0 drafts annotated superseded, links repointed; gh object for v2.8.0 verified published 2026-10-01

### Standing watches

**Status:** 🟡 `PLANNED` · **Priority:** Low · **Effort:** S

Standing watches — QUARTERLY RE-CHECK, next due 2026-12-20 (named triggers fire earlier): sip.js 0.22 (swap ONLY on Chromium WebRTC breakage, a sip.js security advisory, or a needed capability — re-checked 2026-09-23: npm latest still 0.21.2) · templ-components upstream v1.20.x watch (v1.19.4 now adopted in go.mod — daemon sweep; tw.css verified class-identical across the 1.19.2→1.19.4 ride 2026-09-30, no regen needed; next regen only when an adopted component's class set actually changes) · oxlint globals (any new browser global → island oxlint.json) · E2E wall-time budget 445s — greens 195s/184s with FOUC aboard; watch TWO consecutive over-budget runs · MONTHLY erraudit tier-2 re-measure next due 2026-10-22 (RE-MEASURED EARLY 2026-10-01: tier-1 0, tier-2 0, coded-errors 0 — one new-rule FP suppressed with a reasoned nolint: the upgraded erraudit gained a context_loss rule that flags scan OUT-params at store/messages.go:183, garbage-by-definition on scan failure) · go-health M18 park watch: revisit ONLY if go-health hooks startup evaluations or cqrs-htmx grows a readiness hook (hook fires only inside `Probe.Evaluate`, which this wiring never calls — verified v0.3.0 AND v0.4.0, wiring reverted `f7028b4`) · branded-id Valuer/Scanner stays PARKED at the store seam (adopt on next storage-format touch)

**Evidence:** AGENTS erraudit bar + templ-components bullet; re-checked 2026-09-23/24 during the pareto cycle (C22/M22)

### internal/server god-package carve

**Status:** 🟡 `PLANNED` (trigger-based, not date-based) · **Priority:** Medium · **Effort:** M

internal/server god-package carve — TRIGGER: the next file added to internal/server (today 19 files / 3268 LOC + 6073 test LOC / 12-of-15 sibling imports / most-touched package since 09-23 at 51 file-events; exported surface only 29 doc lines, so it is an OPAQUE composition surface, not a god module — but it sits ON the >1000-LOC/>20-file review threshold). Carve `server/api` (JSON endpoints: contacts, calls, session, csrf) + `server/hooks` (webhook ingest + idempotency) out of the wiring (Deps, chain, mount stay); tests move with their files; the contract_test three-401-writer allowlist (actions/webhooks/session_api) + DOM-contract pins update mechanically; render-diff harness proves byte parity if wanted

**Evidence:** arch review 2026-09-30 finding #1 + §05 go-modularize verdict (NO go.mod split — zero Go consumers, Nix vendorHash + release-ritual cost, arch test already enforces the DAG; revisit trigger = a second Go consumer)

### Mic pre-warm live verification + gates

**Status:** 🟡 `PARTIALLY DONE` (code + gates green; live ritual owner) · **Priority:** Medium · **Effort:** S

`mic.js` pre-warm (speak ASAP after accept) shipped with T07 (2026-10-01, code + island pins) and the T11 perf train added the mic-indicator timing baseline; suite/smoke green. Remaining: ONE live-call ritual on the stack (accept→speak sub-second, indicator at ring, warm release on reject/missed) — owner terminal.

**Evidence:** `docs/status/2026-10-01_04-07_mic-prewarm-accept-latency-train-status.md` §b/§f; T11 timing baseline `scripts/perf-baseline.py`.

### Island-honesty train follow-ups

**Status:** 🟡 `PARTIALLY DONE` (webphone side done; stack lane remains) · **Priority:** Medium · **Effort:** S

Hold machine + `#offline-banner` shipped (`cc98c2e`); T15 landed the pending-copy split, Terminated/watchdog `holdPending` clearing, and the T23 visual pass covers the screenshot leg; vulnix (zero real advisories) + aarch64 exit-green done at the T22 close-out (ELF-byte verify owed at each final-gate cross-build). Remaining: the stack browser E2E the T11–T19 served-markup delta owes (blocked by the stack FreeSWITCH `mod_enum` build break — repair that first), then the runbook obligation closes.

**Evidence:** `docs/status/2026-10-01_05-26_island-honesty-hold-offline-banner-train-status.md` §b/§f; T15/T22/T23 session reports 2026-10-02.

### samber/do composition-root + dashboard train follow-ups

**Status:** 🟡 `PARTIALLY DONE` (webphone side done; two owner calls) · **Priority:** Medium · **Effort:** S

health.css: input + staged rebuild script (`scripts/build-health-css.sh`, locked-nixpkgs + staged build — the artifact sat inside its own @source root) + `checks.health-css` canary landed; the 2026-10-02 rebuild fixed a genuinely stale artifact (zombie classes, missing `--blur-xs`). Family pins landed (T22: `store.thread_flag`, `store.count_archived`, snippets). The local-behind-remote divergence is LIVE daemon behavior (observed twice: named commit rewritten 2026-10-02, same content new hashes — never verify via push logs, only `git ls-remote`). Remaining: stack `/health` exposure policy (remote_ip vs PublicMode vs basic auth) + the v2.9.0 fold decision (default: one release, owner §g2) — both owner.

**Evidence:** `docs/status/2026-10-01_02-12_samber-do-composition-root-health-dashboard-train.md`; T21/T22 session report 2026-10-02 11:43 §a; T18–T23 session report 2026-10-02 11:41 §d.

### Visual harness: shot disposition + optional vision cross-check

**Status:** 🟡 `PARTIALLY DONE` (harness shipped; disposition is owner) · **Priority:** Low · **Effort:** S

`scripts/ui-capture.py` is the persisted harness (T23): config-file boot with the fronted CSRF shape, HTTP seeding (multipart sends, meta-CSRF → `/api/session` → rotated token), session-cookie INJECTION into chromium (the island's login gates on the SIP WS a bare boot cannot serve), per-surface DOM assertions before every shot, and the 14-shot matrix (7 surfaces × light/dark) into `ui-shots/`. LOCAL-ONLY by budget decision — see the AGENTS command entry. Remaining (owner §g1): eyeball the current matrix; decide per-release vs per-train persistence (default: per release); the optional vision-CLI cross-check needs a provider/key decision.

**Evidence:** T23 in `docs/status/2026-10-02_11-41_t18-t23-train-session7-status.md` §a/§b; run recipe in AGENTS § Commands.

### Cross-repo obligations (stack / pbx-artmann)

**Status:** 🟡 `PLANNED` · **Priority:** Medium · **Effort:** S-M

Owner/stack legs that block or derive from webphone work: repair the stack FreeSWITCH `mod_enum` build → run the stack browser E2E → relock to webphone `cc98c2e` (or newer) + pbx-artmann re-pin; **the cascade-fix train (2026-10-04, island style scoped to `.island` + history filter `wp-mini`) AND the passkey-tail session (same day: csrf.js module, whoamiLine move, enroll error copy) changed served markup/assets — the stack browser E2E re-run above covers both trains' obligation**; stack `services.webphone.paperless` module option + smoke arm; the WebTransport-not-adopted verdict doc; telephony `deploy.md` secret PATH column; ops-runbook demo-call recipe (`originate user/1000 &playback(local_stream://moh)` + the `/var/lib/telephony-secrets/` password path); MOH audibility + `/recordings/` + CDR check; the gateway stack-side bits (`ftypqt`→`video/quicktime` sniff, E2E MMS-outbound, pbx-artmann FEATURES:87 stale text). No assistant ssh — verified handovers only.

**Evidence:** 02:54 / 04:07 / 05:26 reports §f; owner command sheet `docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md`.
