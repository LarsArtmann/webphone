# TODO_LIST

Short- and mid-term improvement tasks. Done work is DELETED, never struck
through (docs-health: shipped work moves to CHANGELOG / FEATURES — one home
per fact). Long-shot ideas and owner calls live in ROADMAP.md; every row
cites its source report in Evidence.

Last sweep: 2026-10-06 (docs-health v7: 23:22 §f harvest below; the 2026-09 verdict/decision cohort annotated + archived — 11 files; TODO/ROADMAP/FEATURES citation repoints; briefing rows 36–38). Prior sweep: 2026-10-05 (Pareto-plan execution session). Folded then: the 14:50 mixin-review session §f (daemon-commit verification DONE — remote `5a4f9db`, content intact; QMD `get` diagnosis DONE — crush `#3846`, see Standing watches; compose noise-floor DONE — registry sweep log 16:40; mixin/CRM/Paperless/vcard/errorfamily.Join owner calls → briefing rows 29–32); stack CI verdict recorded (RED by 1h timeout); erraudit re-measured early (tier-1/2a/2b = 0/0/0 after 7 reasoned suppressions; boot surfaces re-graded intact). Prior sweep 2026-10-04 (passkey train-tail session; AGENTS-over-cap row removed, tail code-polish executed, boot-contract runbook patched, `890a526` CI re-dispatched).

## Open tasks

### Gate-recovery session tail (2026-10-06 22:08 report)

**Status:** 🟡 `PARTIALLY DONE` (critical legs closed 2026-10-06 ~22:45; tails remain) · **Priority:** Medium · **Effort:** S-M

CLOSED this session: the red CI (vendorHash `WYqTip91…` stale after the post-train tidy — re-pinned `5ekZFK2m…`, CI run 37526835020 green) · `TestSQLiteSessionTTLExpiryAndSweep` hardened with an injectable store clock (both session stores; race structurally impossible, suite + buildflow green) · fresh-binary smoke boot (47+4+8, 0 failed) · buildflow verdict rc=0 · AGENTS lessons (kill builtin trap, vendorHash `--rebuild` recipe + tidy-then-repin sequencing, templ-components v1.20.0 poisoned require). REMAINING from the report's (f) list: ui-capture 14-shot visual pass on the v1.20.1 tree (f6) · vulnix `0/5 deterministic-retry` warning investigation (f12) · GOPROXY fallback chain + BuildFlow retry-wrapper (f9/f10, BuildFlow-repo/crush-config side) · module-cache janitor + leftover zero-byte `.tmp` sweep (f11/f16) · upstream releases: templ-components errorpage tagging discipline (f27). Stack browser E2E for the dep train (f8) rides the next stack session. 23:22 §f remainder folded 2026-10-06 (docs-health v7): whitespace-drift `.nix` gap + self-test (r1) · the two owed lessons stories — outage timeline + `/go.mod` h1 ≠ sha256 oracle (r2) · CI-flake ledger (r6) · stale BuildFlow binary rebuild (r10, BuildFlow-repo side; folds with f9/f10) · buildflow doctor gate in the release-runbook prelude + `timings --regressions` rebaseline (r22/r23) · v1.20.1 greppable-class byte check `wp-thread-row`/`wp-bubble` (r16) · fleet wedged-go audit + /tmp probe-artifact prune + gopls workspace close (r19/r20/r21) · run-37496169265 flake-window closure (r30) · quiesced-host local flake check incl. the KVM backup leg (r15) · kill-builtin lesson empirical verify at the next natural kill (r14) · DOM-contract v1.20.1 reasoning note (r17) · go-cqrs-lite release-tooling audit (r13, upstream) · owner legs routed to briefing rows 36–37: devShells dprint/prettier/ruff/lychee additions + lychee auth-vs-exclude (r25/r26). Dropped as landed: r27 (the go get → tidy → vendor → repin chain IS the AGENTS vendorHash bullet).

**Evidence:** report `docs/status/2026-10-06_22-08_buildflow-gate-recovery-go-module-train.md` (§f rows); CI `gh run list` 37526835020

### Passkey train tail: deploy proof (owner terminal)

**Status:** 🟡 `IN_PROGRESS` (code polish DONE 2026-10-04; owner legs remain) · **Priority:** High · **Effort:** S

The code train AND the 3-repo deploy train are DONE (webphone `400eaff`
pushed; sibling `890a526`; pbx-artmann staged `0ngvm4q7…`, probe OK,
gates green — see CHANGELOG Unreleased + the 17-28 report); AUTH AUDIT 2026-10-06 (13:10): the passkey HTTP surface was code-read audited; one accepted-risk
observation needs an owner verdict — `POST /api/auth/passkey/enroll/begin` gates on the account
ULID alone (any well-formed usermgmt UserID drives a registration ceremony; the enrollment TOKEN
is only enforced one step earlier at /verify). Protection = ULID unguessability (128-bit, never
exposed except to valid-token holders) + passkeyLimiter + mode default-OFF. Options: bind the
ceremony to a token-verified server-side session (usermgmt upstream change) or ratify ULID-secrecy
as the standing posture. Companion reconfirmation: session-row password stays plaintext at rest
(spike-verdict ratified; 0700 DB + UMask 0077). Same session FIXED: session-cookie Secure behind
the TLS proxy, real-config CSRF rotation, constant-time webhook secret (CHANGELOG § Security).
AUTH TAIL 2026-10-06 (D18): renewal Secure re-issue pinned end-to-end (incl. the half-life
throttle + same-token assertions; `62bc7dc`); slog secret-leak grep over server/userauth/gateway
CLEAN (attrs are errors/status/family/code/extension only; vendored usermgmt logs the cookie NAME
and bot id, never values; enroll-token errors are static strings; gateway logs nothing); codespell
over the CHANGELOG/lessons/status/planning delta EXIT 0 — the four `keep-alives` hits are the
twice-adjudicated plural noun, now durably in `.codespellrc` ignore-words. Full-repo sweep close
(b.3/f.6) 2026-10-06 15:50: the five unswept packages are CLEAN — app (one `log.Info` "timezone
applied" zone attr), pbx + web/views (ZERO log calls), fax (6 slog.Warn: error + fax_id),
messaging (3 slog.Warn: error + message id + owner extension + remote number — CDR-class data,
the accepted 18.2 bar); no print/os.Stderr paths anywhere in the five. ISLAND AUDIT 2026-10-06 (D19): clean — adoption ladder correct
(3 backoff retries, reload last resort); zero raw innerHTML sinks (whoami/snippet/lightbox/errors
all textContent/value/src); zero credential persistence in localStorage; legacy pbx-contacts
import deletes only after EVERY row is server-accepted, else retries next login; no
password/token in any log line. FIXED: authedFetch now imports the csrf.js meta reader (ONE-home
rule; it had drifted to an inline querySelector). ACCEPTED: passkey/enroll finish carry the
one-time ceremony session_key as ?user_id= (burned at finish, worthless without the WebAuthn
assertion + CSRF; exposure = proxy access logs only) — moving it into the POST body is a cheap
option inside D30's decision space.
the code-polish tail landed 2026-10-04 (enrollFailed(0)→enrollNetFailed,
csrf.js single home, userauth /healthz leg, tier-2 pins, whoamiLine→ui.js,
typed sessionIdentity, plan-doc annotation, AGENTS trim — see CHANGELOG).
Remaining: the OWNER switch + post-switch ritual (rotate
`/tmp/pbx-toplevel-current`, record the fresh diff-closures baseline)
and the LIVE passkey proof with the fail-closed password-file drill
(pbx-artmann TODO §1 row). OWNER CALLS outstanding: ratify
runbook-only enroll + Lars-only v1 mapping; switch-now-vs-CI-first
(VERDICT 2026-10-05: every recent stack `main` CI run cancels at the 1h
ceiling inside `nix flake check` — aarch64 VM leg green 6m16s, browser
E2E on-demand never fires; the mod_enum/timeout diagnosis precedes any
lock bump); installer release republish timing (stale since the
relock). HANDOFF NOTE: the gopls unused-`r` Info at
`internal/server/passkey_api.go:236` is still live at HEAD (file last
touched 2026-10-05 14:52 by this train) — the owning session resolves
it; no cross-session fixes.

**Evidence:** status `docs/status/2026-10-04_17-28_passkey-train-resumed-session-brutal-status.md` (§f harvest source); plan `docs/planning/2026-10-04_12-17_SUPERB-passkey-users.md`

### Boot-contract tail: D3 retry-loop owner call + stack runbook patch

**Status:** 🟡 `PLANNED` (owner call + stack dispatch) · **Priority:** Medium · **Effort:** S

The boot-error contract SHIPPED 2026-10-02 (5-part render on every boot failure, exit 1 designed / 2 panic, EN-only, class-tagged; `App.Start` de-panicked; arch test confines samber/do; smoke `boot failure scenario` green). The stack runbook patch APPLIED 2026-10-04 (stack `9a21893`: boot-surface + passkey error-contract rows; the parked patch doc is spent). Remaining: (1) OWNER CALL (D3 follow-up) — cap systemd's boot-failure retry loop (`StartLimitBurst`/`StartLimitIntervalSec`) or ratify the 5s `Restart=on-failure` retry as desirable liveness (module deliberately unchanged this train). The boot-surface re-grade landed EARLY 2026-10-05 (contract table vs code: 7 classes + drift pins intact; see the erraudit line in Standing watches)

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

**Evidence:** DONE: codespell real-binary + `.codespellrc` (vendor/go.sum skipped, pre-emptive ignored), AGENTS buildflow-claim reconcile, render-diff script committed (`scripts/render-diff.py`; LIVE-VERIFIED 2026-10-01: 7/7 partials byte-identical + 7/7 error arm), BuildFlow freshness noted via doctor. v6 stragglers DONE 2026-10-04: health-css fmt appended (live-verified: raw rebuild → in-script fmt → byte-identical with HEAD), smoke mypy 8→0 (annotations were the lie; ruff clean), "error families are total" + 104-code registry table in error-contract.md generated and freshness-pinned by `TestErrorCodeRegistryIsFresh` (negative-tested; fixes the SkipAll-walk latent bug in the sibling arch test too). REMAINING (owner): markdownlint posture (briefing row 18) — invocation SETTLED 2026-10-05: `nix run nixpkgs#markdownlint-cli2 -- <docs...>` (v0.23.3, no config = library defaults; baseline over the three 10-05 session docs = 172 findings, 170 MD013 line-length + 2 MD026 heading-punct, all cosmetic — detect-only until the row-18 posture rules otherwise). Round-3 C2 gate-debt close 2026-10-06 03:15 (01-40 report addendum has the detail): buildflow EXIT 0 over `43091cc` (incl. `2b52fbf`), island JS 186/186 pass, on-demand codespell EXIT 0 (raw scan: one pre-existing warning-level `keep-alives` in CHANGELOG prose — not session debt). Round-3 C10 daemon zero-churn observation 2026-10-06 03:20: five daemon commits since the `2b52fbf` regen (`e9caa75`, `43091cc`, `be4d5e0`, `b245b73`, `e62fe34`) left `docs/error-contract.md` untouched — md5 stable at `d25d4a50…`; the churn kill now has observational backing, not just the 130-row byte-verification (canary stays open: any future daemon commit touching the file must show a zero reflow diff). C9 landed same session: writer framing + uniform-row-length pins in `rewriteRegistryBlock`, 00-52 report table re-flowed. Round-4 D25 hygiene 2026-10-06: near-aligned-table
sweep over every 10-05/10-06 status+planning doc found exactly ONE mixed-shape table (the 03-40
report) — normalized to the daemon-canonical compact shape; the detector (mixed column widths +
extra cell padding) now reports ZERO across the corpus. `scripts/whitespace-drift.sh` landed:
staged-diff `-w` divergence check over formatter-OWNED files only (Go/JS/MJS/CSS/Nix; md
deliberately excluded — unowned means no drift definition), self-tested in a scratch repo
(whitespace-only Go edit EXIT 1 with the `nix fmt` fix, substantive EXIT 0, md-only "unowned"),
wired as an AGENTS Commands habit line — kills the e62fe34 class at add-time. D18.3 durability:
`keep-alives` joined `.codespellrc` ignore-words after its second adjudication. d.1 fix 2026-10-06
15:40 (hardened twice by 16:45): the detector is now a COMMITTED instrument —
`scripts/md-table-shape.py` (self-test 8/8: mixed flags, compact clean, fully-aligned clean =
D1.4's question, unbalanced-padding flags, mid-cell wrapped rows clean, prose-after-table never
glues, escaped pipes never split a cell, display-width-aligned emoji columns clean); default
scope docs/status/ live = ZERO; explicit rerun over the exact 10-05/10-06 status+planning sweep
set = ZERO, reproducing the claim from the repo. v1 miscounted wrapped tables (fragmented rows)
and escaped pipes; v3/v4 were caught by dogfooding against the session's own report and the
daemon's own output (e.3 rule) — v4 measures DISPLAY width because the daemon's aligner pads
emoji to terminal cells, not codepoints. Full-corpus run 16:45: ZERO findings (v3's 25 shrank to
1 real finding after the daemon's 16:18 aligning pass + display-width fix; the samber-do
scorecard's Status column was normalized to compact on the spot).
f.7 canary FIRED 2026-10-06 16:18 (`2fbdfd7`): the daemon's reflow pass touched the normalized
03-40 table and REFLOWED it compact → fully-aligned (content-inert, no mix introduced; detector
ZERO post-pass). VERDICT INVERTED for D1.4: the daemon IS the aligner — near-aligned mixes come
from sessions appending compact rows onto daemon-aligned tables (the briefing rows-29–34
pattern), and the daemon normalizes them on its next md pass. The same daemon pass reformatted
`scripts/md-table-shape.py` (style-only; self-test still 8/8) and bumped `flake.lock` nixpkgs
(494ce7fd → 151fa4e8 — rides the next push; CI judges).
Was: 09-26 §d3/§f11-12; 09-24 12:41 §f19-20; 04:32 §f14-15

### OWNER-calls batch session

**Status:** 🟡 `PLANNED` (briefing ready, sitting owed) · **Priority:** High · **Effort:** S

OWNER-calls batch session (decisions, one sitting — briefing at `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`): the original ~15 + CRM policy trio + self-send train-C semantics + templ-components history-blemish + tail items g1 force-push ratification / release.sh load-gate default / g2 KVM-timeout / art-dupl `-t 3` baseline (FIFTH surfacing) + registry ratification (`docs/dedup-registry.md` as THE home) + `settingsRow` build-or-retire (FOURTH surfacing) + suppression-bucket doc + dedup suppression-SCOPE ruling (shown-only vs periodic suppressed-set audit; 12:59 report g3, unrouted until v6) + webhook 400 body-text dependents + `msg/`→`message/` idem-key rename + helper micro-test bar + missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy. NEW since 2026-10-06 (briefing rows 36–38): the `auth:`-prefix × daemon-sweep interface (manual `auth:` commit before the sweep window vs accept a daemon `chore:` commit when the pinned session-behavior suite provably ran — 23:22 g2) · devShells.default tool additions (dprint/prettier/ruff/lychee, silences 4 buildflow warnings) + lychee authenticate-vs-exclude policy (22:08 f13/f14) · docs-only express-push ratification (23:22 g3). NEW since 2026-10-05 (briefing rows 29–32, addendum in the briefing doc): mixin/dupe detector-policy home (registry sweep-log only vs AGENTS promotion) · `config.CRM`/`Paperless` URL+Token twin standing ruling (resurfaced by TWO detectors) · `vcard.Card`/`SharedContact`/`apiSharedContact` triad boundary-layer rejection ratification · `errorfamily.Join` gap (add upstream constructor vs keep 4 nolint aggregate sites). NEW since 2026-10-02 (docs-health v6 sweep): ratify "routed-as-resolved → archive" for the four live owner-gated plans (error-excellence, 20-year-durability, 04-29 pareto, 13-33 stack-adoption — v4's g1); ratify train-level `v` verdict markers as the standing annotation hash bar (v3's g3, six sweeps running); the check-rows marker-cell baseline for the 37 pre-2026-09-18 archived files (restyle or accept). NEW since 2026-09-26: push-lag threshold policy (when is a silent daemon push stall BROKEN — 10 min? 1 h) + session push ratification (phase-boundary hand-pushes) · is the next webphone release 2.8.0 (making the bridge's ">= 2.8" claims true) or reword them · `gateway.attachment_limit` knob vs bridge-422-teaches design as final · sniff-fallback lifespan (keep forever vs delete-after-deploy-confirmed) · tier-2 family-adoption intent (dep swept-but-unused?) · report erraudit `tree` same-name-dedupe upstream? (owner's own tool) · AGENTS restructure permission · existing-prod-data chmod/re-backup for the UMask tightening · destDir nesting legality · stack verification now-vs-train-close

**Evidence:** briefing updated 2026-10-01 to 28 rows (15–28: v2.8.0 + setup NO-GO ratifications, compaction permission, tooling postures, push-lag, QMD, dedup baseline, gateway micro-decisions) + a closed-since section (go-cqrs-lite wontfix etc.)

### Post the release announcements

**Status:** 🟡 `PLANNED` (owner posts) · **Priority:** Low · **Effort:** S

Post the release announcements (drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 at `docs/announcements/`; owner picks channel(s), approves wording, decides the disclosure posture — fix-acknowledged, no exploit detail)

**Evidence:** v2.8.0 drafts A/B/C = `docs/announcements/2026-09-30_v2-8-0_drafts.md` (covers the merged 2.7.0+2.8.0 story); v2.7.0 drafts annotated superseded, links repointed; gh object for v2.8.0 verified published 2026-10-01

### Standing watches

**Status:** 🟡 `PLANNED` · **Priority:** Low · **Effort:** S

Standing watches — QUARTERLY RE-CHECK, next due 2026-12-20 (named triggers fire earlier): sip.js 0.22 (swap ONLY on Chromium WebRTC breakage, a sip.js security advisory, or a needed capability — re-checked 2026-09-23: npm latest still 0.21.2) · templ-components upstream v1.20.x watch (v1.19.4 now adopted in go.mod — daemon sweep; tw.css verified class-identical across the 1.19.2→1.19.4 ride 2026-09-30, no regen needed; next regen only when an adopted component's class set actually changes) · oxlint globals (any new browser global → island oxlint.json) · E2E wall-time budget 445s — greens 195s/184s with FOUC aboard; watch TWO consecutive over-budget runs · MONTHLY erraudit tier-2 re-measure next due 2026-11-05 (RE-MEASURED EARLY 2026-10-05: tier-1 0, tier-2a 0, tier-2b 0 — the passkey train (10-01→10-04) had regressed tier-2 with 7 findings post-09-30-pin: 2 bare constructors at enroll-boot sites, 4 errors.Join shutdown aggregates [errorfamily has no Join — briefing row 32], 1 bool-blank FP where `Duplicate()` returns (int64, bool, bool); all suppressed with reasoned nolints, boot-surface re-grade intact; prior early measure 2026-10-01 with one context_loss FP) · QMD `get`/`multi_get` render bug — blocked UPSTREAM on crush `#3846` (MCP embedded-resource content rendered as Go struct pointer; qmd 2.8.3 is spec-compliant; workaround = `query` tool + disk reads at the corpus root; re-test `get` after the next Crush upgrade; SUBSCRIBE attempt 2026-10-05 blocked — gh token lacks the `notifications` scope; owner one-liner: `gh auth refresh -s notifications` then the `updateSubscription` GraphQL mutation on the issue's node id, or click Subscribe in the web UI) · go-health M18 park watch: revisit ONLY if go-health hooks startup evaluations or cqrs-htmx grows a readiness hook (hook fires only inside `Probe.Evaluate`, which this wiring never calls — verified v0.3.0 AND v0.4.0, wiring reverted `f7028b4`) · branded-id Valuer/Scanner stays PARKED at the store seam (adopt on next storage-format touch) · sessions credentials-at-rest periodic re-review: next due 2026-12-30 OR on the next session-storage change (the 2026-10-06 auth train's changes shipped WITH the 13:10 audit's at-rest reconfirmation — posture stands; baseline `docs/reviews/2026-09-30_sessions-credentials-at-rest-review.md`)

**Evidence:** AGENTS erraudit bar + templ-components bullet; re-checked 2026-09-23/24 during the pareto cycle (C22/M22)

### internal/server god-package carve

**Status:** 🟡 `PLANNED` (trigger-based, not date-based) · **Priority:** Medium · **Effort:** M

internal/server god-package carve — TRIGGER: the next file added to internal/server (today 19 files / 3268 LOC + 6073 test LOC / 12-of-15 sibling imports / most-touched package since 09-23 at 51 file-events; exported surface only 29 doc lines, so it is an OPAQUE composition surface, not a god module — but it sits ON the >1000-LOC/>20-file review threshold). Carve `server/api` (JSON endpoints: contacts, calls, session, csrf) + `server/hooks` (webhook ingest + idempotency) out of the wiring (Deps, chain, mount stay); tests move with their files; the contract_test three-401-writer allowlist (actions/webhooks/session_api) + DOM-contract pins update mechanically; render-diff harness proves byte parity if wanted
Trigger adjudication (2026-10-06 session): the 09-30 baseline reading
fires on `passkey_api.go` (landed 10-04), but the round-2 plan
(2026-10-05 20:30, §F15 "do NOT start") supersedes — the trigger is
the NEXT file landing in `internal/server` AFTER that plan; stays
gated until then.

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
