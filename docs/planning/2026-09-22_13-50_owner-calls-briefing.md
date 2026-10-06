# Owner-Calls Briefing — the Sitting Decisions (SUPERB T11)

- **Date:** 2026-09-22 13:50 CEST, updated 2026-09-30 14:00 CEST (rows
  15–28 = everything accumulated since; rows 15/16 ratify autonomous
  calls already executed under the 2026-09-30 full-execution GO)
- **Purpose:** one sitting, ~30 minutes. Every row blocks or shapes
  downstream work in the 20-year plan
  (`docs/planning/2026-09-22_12-02_SUPERB-20-year-durability.md`).
  Each row: the question, what it gates, options, MY recommendation.
  Say "go with the recommendation" per row or override — verdicts land
  in their owning docs the same session.

| #  | Decision                                                                                                                                                                          | Gates                                       | Options                                                                                    | Recommendation                                                                                                                                               |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1  | **Train-cut**: cut v2.5.0 now (fold the [Unreleased] pile: styled 404, module options, VM test, htmx bundle, error-excellence, registration-loss rebuild) or hold per g2 cadence? | T6/T7 (release machine), announcement scope | (a) cut now, deploy once with everything; (b) deploy v2.4.0-chain only, cut later          | **(a) cut now** — the pile IS a user-visible theme (reliability + error excellence); one deploy carries all; the g2 rule wants theme-trains, and this is one |
| 2  | **`/livez` consumer**: wire anything to it (monitoring/systemd) or keep as contract-only?                                                                                         | T18c startup contract                       | (a) systemd watchdog consumes it; (b) fleet scraper; (c) contract-only for now             | **(c) contract-only**, revisit when a fleet scraper lands (the stack already fences `/healthz`)                                                              |
| 3  | **HSTS on prod**: enable `nginx.hsts` for pbx.artmann.tech?                                                                                                                       | stack config                                | (a) enable max-age 31536000; (b) short max-age first (31536000 → ratchet); (c) off         | **(b)** short window first (e.g. 300s → 1d → 1y) — https-only is proven but a WSS mishap bricks the phone until header expiry                                |
| 4  | **XFF sanitization flip**: does the stack's nginx sanitize `X-Forwarded-For` enough to flip webphone limiter keys to `KeyExtractorFromClientIP`?                                  | rate-limit fairness behind the proxy        | (a) flip now; (b) verify + flip later; (c) keep peer-host keys                             | **(c→b)** keep safe default until the stack's XFF handling is explicitly audited (one grep in the stack vhost); then flip                                    |
| 5  | **Disclosure posture** for announcements: name the forged-session hole publicly?                                                                                                  | T9 announcements                            | (a) full detail; (b) fix-acknowledged, no exploit detail; (c) no security mention          | **(b)** — single-tenant PBX, hole fixed before disclosure, detail only helps no one now                                                                      |
| 6  | **18-49 review annotate scope**: full ANNOTATE pass or judgment-sample?                                                                                                           | T16c docs kernel                            | (a) full; (b) f-items only                                                                 | **(b) f-items only** — the 22:26/22:29/23:43/00:14/01:04 pattern worked; full passes cost hours for marginal rows                                            |
| 7  | **Loopback `delivered`**: simulate `delivered` in loopback mode (like fax `transmitted`) or keep `sent` honest-terminal?                                                          | dev-mode fidelity                           | (a) simulate; (b) keep                                                                     | **(b) keep** — loopback is a seam for wiring tests, not a demo of provider semantics                                                                         |
| 8  | **Handler dual-layer**: keep `requireSession` helpers + `Sessions.Require` middleware both, or drop one?                                                                          | auth surface simplicity                     | (a) keep both (defense in depth, pages render anonymously); (b) middleware only            | **(a) keep** — the dual layer is pinned by contract tests; dropping buys nothing                                                                             |
| 9  | **gh-release-object habit**: every tag gets a Release object (notes from CHANGELOG)?                                                                                              | T7, future releases                         | (a) always (release.sh already does it); (b) majors only                                   | **(a) always** — it is already automated; the v2.3.0/v2.4.0 empty-body incident showed manual is worse                                                       |
| 10 | **Go module v2 policy**: `/v2` suffix on next major or NOT-DO?                                                                                                                    | future tags                                 | (a) adopt `/v2` at next major; (b) NOT-DO record (flake-consumed app, nobody `go get`s it) | **(b) NOT-DO** — record it, stop asking                                                                                                                      |
| 11 | **Recordings product intent**: panel in webphone (consent/jurisdiction/access model) or leave at stack level?                                                                     | T24 (major feature)                         | (a) build panel; (b) leave stack-level; (c) later                                          | **(c) later, decide by 2026-Q4** — the stack records server-side already; the webphone panel needs an access model first                                     |
| 12 | **TEMP-DIAG keep**: keep the stack E2E answer-phase dump (`b96d4c2` there)?                                                                                                       | E2E noise/verdict hygiene                   | (a) keep always; (b) gate behind env; (c) revert                                           | **(b) gate behind env** — it was diagnostic for the 1001 hunt; always-on adds noise                                                                          |
| 13 | **Oops ratification**: samber/oops stays a documented non-fix, or order a staged-adoption plan?                                                                                   | erraudit tier 3                             | (a) ratify non-fix; (b) adoption plan                                                      | **(a) ratify non-fix** — 2026-09-21 triage found 0 real findings; families are adopted                                                                       |
| 14 | **`backup.retentionDays`**: want the module option (auto-prune old snapshots)?                                                                                                    | T18a                                        | (a) yes, 30d default; (b) yes, 90d; (c) no (manual)                                        | **(a) yes, 30d default** — unbounded snapshot dirs are a slow-burn disk incident                                                                             |

## New since 2026-09-22 (rows 15–28; sources: TODO_LIST owner-calls row + the 2026-09-30 sessions)

| #  | Decision                                                                                                                                                                               | Gates                                        | Options                                                                             | Recommendation                                                                                                                           |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| 15 | **RATIFY: v2.8.0 release number** — cut 2026-09-30 under the execution GO (bridge ">= 2.8" claims now true; the never-tagged 2.7.0 content ships inside it)                            | announcement draft, bridge docs              | (a) ratify; (b) objection → too late for the tag, lands as a v2.8.1/v2.9.0 note     | **(a) ratify** — rewording path died with the tag; number honestly signals the module break                                              |
| 16 | **RATIFY: setup-shell adoption NO-GO** — footprint gate measured +10.40 MB / +68.2% vs ≤ +8 MB / ≤ +20%; lifecycle salvage rides v2.9.0; upstream seams (setup/v4.13.x) stay published | v2.9.0 salvage replay, upstream relationship | (a) ratify + salvage; (b) order deeper upstream slimming first                      | **(a) ratify** — the import-graph lesson says slimming won't close a 68% gap                                                             |
| 17 | **AGENTS.md compaction permission** — 498 → ≤377 lines (war stories → lessons.md, rules stay); asked 2026-09-29, unanswered                                                            | every future session's context cost          | (a) grant (quiet window); (b) keep as-is                                            | **(a) grant** — the file re-pays the edit every session; content rule: next add removes a line                                           |
| 18 | **markdownlint posture** — 5092 detect-only corpus findings invite a reflow disaster someday                                                                                           | T14 tooling hygiene                          | (a) configure markdownlint to house style; (b) record detect-only posture in AGENTS | **(b) detect-only posture** — reflowing 5000 findings churns every blame line for zero reader value                                      |
| 19 | **Push-lag threshold** — when is a silent daemon push stall BROKEN?                                                                                                                    | release preflight, tri-repo honesty          | (a) 10 min; (b) 1 h; (c) phase-boundary hand-pushes only                            | **(b) 1 h** + keep the phase-boundary hand-push habit; release.sh's ls-remote preflight catches the dangerous case (tag-time divergence) |
| 20 | **QMD indexing** — index webphone docs into the knowledge base for prior-ruling recall? (only `cv` is indexed today; sessions pay rediscovery cost)                                    | session speed on recurring questions         | (a) index docs/status+planning+reviews; (b) stay unindexed, record limitation       | **(a) index** — the go-cqrs-lite question cost a full session where a search would have answered                                         |
| 21 | **"Why not library X" rulings pointer** — one-line AGENTS pointer block vs strict one-home (the deep-dive owns the verdict)                                                            | future sessions' recall                      | (a) pointer block (3–5 named rulings, one line each); (b) strict one-home           | **(a) pointer block** — a pointer is not a second fact-home; recall cost of the alternative is proven                                    |
| 22 | **art-dupl `-t 3` baseline + dedup-registry ratification** (fifth surfacing; `docs/dedup-registry.md` as THE acceptance home)                                                          | future dedup sweeps                          | (a) ratify both; (b) tighten to -t 2 fleet-wide                                     | **(a) ratify** — -t 3 with per-site rationale beats noise-flooding -t 2                                                                  |
| 23 | **`settingsRow` build-or-retire** (fourth surfacing)                                                                                                                                   | settings tab structure                       | (a) build the helper; (b) retire the idea                                           | **(b) retire** — four surfacings without a customer need is the answer                                                                   |
| 24 | **Webhook 400 body-text dependents + `msg/`→`message/` idem-key rename** (wire-adjacent micro-decisions)                                                                               | bridge contract, dedupe keys                 | (a) do both on the next gateway train; (b) drop                                     | **(a) do both** — one train, two debts closed                                                                                            |
| 25 | **Missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy** (behavior contracts)                                                             | island/store behavior                        | per-item: keep-current vs change                                                    | keep missed-call REJECT (evidence-first), make `?q=` shareable (URL semantics), keep `Must*` panic (corrupt store should crash loudly)   |
| 26 | **`gateway.attachment_limit` knob** vs bridge-422-teaches design as final                                                                                                              | gateway seam surface                         | (a) add the knob; (b) keep 422-teaches                                              | **(b) keep** — the bridge teaches; a knob adds a config surface for one known consumer                                                   |
| 27 | **Sniff-fallback lifespan** (bridge magic-byte fallback for pre-`e6ea2c7` binaries): keep forever vs delete after deploy-confirmed                                                     | bridge code, webphone compat                 | (a) keep forever; (b) delete once prod > the honest-Type release                    | **(b) delete after v2.8.0 confirmed on prod** — the fallback has a natural expiry                                                        |
| 28 | **Existing-prod-data chmod/re-backup + destDir↔dataDir nesting legality** (UMask tightening follow-ups; stack verification timing)                                                     | prod data hygiene, module assertions         | (a) chmod in place + assert nesting; (b) re-backup only + leave nesting unasserted  | **(a) chmod in place + assert** — one owner command each; nesting is a real footgun                                                      |

## Closed since 2026-09-22 (recorded here so the sitting skips them)

- **go-cqrs-lite question (old g-Q1): CLOSED 2026-09-30 as wontfix** —
  drift-checked (`system/README.md` unchanged since v4.10.0),
  `system.New` read in source (domain composition root, no HTTP
  surface — the old "fights the middleware chain" argument was wrong;
  the real reason: nothing event-sourced to compose), ADR-0123 v5
  impact zero-direct (revisit trigger: event-sourced state or a second
  Go binary — recorded in ROADMAP). Full closure in the 12:58 status
  report's appendix.
- **erraudit tier-2 family-adoption intent**: DONE 2026-09-30 — every
  seam converted, tier-2 = 0; the go-error-family dep is used, not
  swept-but-unused.
- **Next-train sequencing (old g-Q3)**: settled by events — upstream
  shipped the setup seams; release tail ran first per the TODO PREREQ.
- SMS-bridge triage: the owner command sheet §4 now carries the full
  decision tree (unit status → journal grep → cred fixes → restart →
  test SMS) — run it whenever prod SMS misbehaves.

## Not-decisions (already DECIDED, listed so the batch stays one sitting)

- Stack rides webphone `main` + per-train lock bump (2026-09-20).
- pbx-artmann stays a rev pin (2026-09-21 policy; bumped today to
  `1a95a736`).
- Shell copy stays English (D3, 2026-09-20); island copy en/de.
- No Go-side telephony, no speculative sip.js swap (named triggers
  only), no Playwright while the stack E2E suffices.

## Addendum 2026-10-05 16:50 — pre-sitting sweep results + rows 29–32

Verification sweep (P1 of the 15:25 Pareto plan, executed):

- **Stack CI = RED**: every recent `main` run cancels at GitHub's 1h
  ceiling inside `nix flake check` (aarch64 VM leg green in 6m16s;
  browser E2E is on-demand and never fires; `890a526`'s re-dispatch
  was superseded by later pushes). Confirms the deploy train's
  mod_enum/timeout diagnosis (F4.2) must precede any lock bump.
- webphone `main` end-state: `5a4f9db` on origin; all 2026-10-05
  session docs landed (plan commit `32905ce`; daemon applied table
  reflows only, content intact).
- `passkey_api.go:211` gopls unused-`r` Info: still live at HEAD;
  the passkey train last touched that file 2026-10-05 14:52 —
  HANDED OFF to that session (multi-session rule), not fixed here.

Assistant legs executed since the plan (evidence for the sitting):

- **P8 erraudit re-measured early**: tier-1 = 0, tier-2a = 0,
  tier-2b = 0 — after suppressing 7 post-09-30 passkey-train
  findings (2 bare constructors at enroll-boot sites, 4 `errors.Join`
  shutdown aggregates [errorfamily has no Join], 1 bool-blank false
  positive where `Duplicate()` returns `(int64, bool, bool)`).
  Boot-surface re-grade: contract table matches code, all drift pins
  present. Next due 2026-11-05.
- **P9 QMD `get` garbage output**: root cause is
  **crush #3846** (open upstream, 0 comments): MCP results with
  embedded-resource content render as a raw Go struct pointer
  (`&{0x… map[] <nil>}`). qmd 2.8.3 (latest) is spec-compliant; the
  bug is Crush's client. No local fix possible; workaround = `query`
  (text results render fine) + disk reads. Unblock = next Crush
  release. Status-report question 3 answered: not owner-known, not
  qmd's bug.
- **P14 detector noise floor**: `branching-flow compose` no longer
  exists in the installed build (command drift); `stats`+`dupe`
  baseline recorded in the registry sweep log — family total 178,
  mixins 8 (the rejected set), dupe 2 actionable (both triaged, 0
  refactors).

New rows for the sitting:

| #  | Decision                                                                                                                                                                                                                                                                                    | Gates                             | Options                                                                                                                                 | Recommendation                                                                                                                                    |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 29 | Detector-policy home: mixin/dupe verdicts as registry sweep-log lines only, or promote a standing AGENTS hard-won rule?                                                                                                                                                                     | future runs                       | (a) registry only (14:46 + 16:40 lines); (b) promote to AGENTS                                                                          | **(a) registry only** — one-home doctrine + the 377-line cap; promote only on a third unhandled resurfacing                                       |
| 30 | `config.CRM`/`Paperless` URL+Token twin — standing ruling? (now resurfaced by TWO detectors: mixins 10-05 + dupe 16:40)                                                                                                                                                                     | dedup registry                    | (a) standing ACCEPT-similarity row; (b) sweep-log rejection only                                                                        | **(a) standing row** — pre-empts every future resurfacing; the embed itself stays rejected                                                        |
| 31 | `vcard.Card`/`domain.SharedContact`/`server.apiSharedContact` Name+Number triad — ratify the boundary-layer rejection?                                                                                                                                                                      | dedup registry                    | (a) ratify as standing; (b) sweep-log only                                                                                              | **(a) ratify** — parse/domain/wire layers are doctrine (same class as the InboundMessage reject)                                                  |
| 32 | `errorfamily.Join` gap — add a Join/Aggregate constructor to go-error-family, or keep the 4 `nolint:erraudit` shutdown-aggregate sites?                                                                                                                                                     | tier-2 posture                    | (a) upstream lib train now; (b) keep nolints                                                                                            | **(b) keep nolints** until a third aggregate site appears; then (a)                                                                               |
| 33 | Daemon-docs-formatting policy: the auto-commit daemon reflows markdown tables (broke the registry pin's byte-exact form twice, and a third reflow landed in the 20:33 plan commit before the pin went content-based). Keep the behavior, or exclude `docs/**` from daemon formatting?       | daemon config, doc churn          | (a) keep reflowing (pins must be formatting-insensitive); (b) exclude docs/**                                                           | **(a) keep** — the pin is now content-based (2026-10-05 fix), reflows are content-inert, and (b) needs daemon-side config for zero remaining harm |
| 34 | Proof-bar for "green": is CI-success on the pushed head the ratified bar for sittings/releases, or a local full gate (clean-cache buildflow + full `nix flake check`)? (Today's 4h-red window was caught by CI, not local gates.)                                                           | release ritual, sitting readiness | (a) CI verdict on the pushed head (release.sh keeps its own local full gate at release time); (b) local full gates before every sitting | **(a) CI on the pushed head** — it caught what local runs missed twice today; the release ritual keeps the heavyweight local proof                |
| 35 | Passkey enroll ceremony carries the one-time key as a `?user_id=` query param on the email link (single-use, burned at finish, TLS-covered) — ratify the acceptance, or fold into row 1.6's token-binding verdict? Surfaced 2026-10-06 so the acceptance cannot die inside an unopened D30. | auth posture                      | (a) accept as-is (status quo); (b) revisit only if D1.6 picks binding and D30 opens anyway                                              | **(a) accept** — single-use + burned-at-finish + TLS makes the query param a cheap option inside D30's decision space, not a standing exposure    |
| 36 | `auth:`-prefix × daemon-sweep interface: when a train touches session/cookie/passkey surfaces, must the session manually commit with the `auth:` prefix BEFORE the daemon's sweep window (slower, prefix always attributable), or is a daemon `chore:` commit acceptable when the pinned session-behavior suite provably ran? (Born 2026-10-06 23:22 §g2.) | convention, auth trains           | (a) manual `auth:` commit first, always; (b) suite-proof substitutes for the prefix                                               | **(a) manual-first** — the prefix is the cheap, greppable provenance; (b) asks a future reader to trust a CI lookup for every auth-shaped commit |
| 37 | devShells.default tool coverage: add dprint/prettier/ruff/lychee to silence the four standing buildflow "tool unavailable" warnings (heavier shell closure), and should lychee authenticate (`GITHUB_TOKEN`) or keep the exclude list? (22:08 §f13/f14.) | tooling posture                  | (a) add tools + lychee token; (b) add tools + keep excludes; (c) leave warnings as documented noise                              | **(b)** — the warnings are detect-only noise otherwise, and a CI token needs a secret decision first                                   |
| 38 | Docs-only express push: ratify the 2026-10-06 docs-only hand-push (b/4) as a standing narrow exception to the never-push rule when main is red and the fix is documentation, or censure it? (23:22 §g3.) | push policy                      | (a) ratify as standing narrow exception; (b) censure — daemon-only pushes, always                                                  | **(a) ratify** — matches the AGENTS manual-push bar's spirit (red main + verified fix in hand); docs changes carry zero build risk                        |

## Round-2 sweep-audit outcomes (2026-10-05 21:10, pre-sitting closure of the 15:42 train)

- **Sweep train = FINISHED** (status-report question 1): `81ea689`
  (15:42) is the last commit touching go.mod/go.sum/vendor; no dependency
  commits since; tree clean apart from session docs. Nothing in-flight.
- **CHANGELOG**: [Unreleased] ### Changed now carries the sweep entry —
  all six bumps old→new (templ 0.3.1020→0.3.1070, templ-components
  1.19.4→1.20.0 + submodules, usermgmt 4.13.1→4.14.0, go-health
  0.4.1→0.5.0, webauthn 0.18.1→0.18.2 indirect, otel 1.46.0→1.47.0
  indirect; go-health-dashboard held at 0.10.2) + the tw.css consequence.
- **tw.css class-identity check (watches-row obligation): CLASSES
  CHANGED** — templ-components 1.20.0 moved the button
  outline-warning/success variants from amber/green-600 text to -700,
  which fell outside the @theme token remap and resolved to RAW palette
  colors. Fixed: remap gained `-700` entries (matching the red-family
  precedent), artifact rebuilt via `scripts/build-tw-css.sh`; the
  utilities resolve through `--warn`/`--ok` again.
- **vulnix over the NEW runtime closure: CLEAN** — 8 derivations
  scanned; only glibc 2.44-25 range-match noise; both CVEs
  (2026-5435, 2026-6238) distro-patched in the locked nixpkgs rev;
  triage CLI verdict "zero real advisories", exit 0.
- **Registry pin made content-based** (incident note for row 33): a
  THIRD daemon reflow had landed in the 20:33 plan commit and re-broke
  the byte-exact pin (CI red again). `TestErrorCodeRegistryIsFresh` now
  compares cell content with padding collapsed; a micro-test pins that
  reflowed rows canonicalize while real drift still fails by name.
