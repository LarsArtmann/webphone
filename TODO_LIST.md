# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in ROADMAP.md (raw ideas + open questions); owner calls there too.

Last sweep: 2026-09-25 (docs-health VERIFY: every claim checked
against code/git, prod probed — closures appended).
Closed since the 2026-09-24 cycle: the stack-adoption quality train —
500-redaction (`a74f8e6`), 61MB upload envelope (`55cb19f`),
`errorfamily.LogErrorContext` (`88cf0ef`) + `errorfamilytest`
asserts in the five family files (`0bd6576`), httpspec conformance
19/19 (`e9f6310` — the middleware_test asserts stay DELIBERATELY:
they pin the panic lane httpspec does not cover), EmptyState ×6 +
`/assets/tw.css` (`f1a2853`), self-send train C (`37ffc53`);
send-failure (F) found ALREADY shipped in v2.6.0 (`d9ce6ff`:
session `did` payload + `thread.selfNotice`, pinned by
session_test.go); v2.7.0 folded but UNTAGGED (`78b30cf`, version
bump `5a689ef`, dep sweep cqrs-htmx v4.12.0 / httputil v1.3.0 /
go-error-family v0.10.2 / go-sse v0.6.1); v2.7.0 announcement
drafts (`357ffec`); the owner-terminal command sheet at
`docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md`
(`875e644`; §0 URGENT host-nix heal + the full v2.7.0 deploy
ritual); release.sh host-nix
preflight (`fe09106`); and the 2.6.0 chain DEPLOYED: prod
`/version` = v2.6.0 (probed 2026-09-25).
Closed since the 04:26 audit: both helper micro-test batches incl. the
apiSave/apiDelete gap tests (`7bf32a3`), full-package `-race` green,
error-contract sync + JSON-204 row + AGENTS rationale registry
(`d42902b`), train E provider-refusal 422 with both runbooks synced
(`6ac8962`), export_test nolint eliminated (`92715a0`), crm.Names
bounded concurrency + full-code-review report (`a43737f`), release.sh
load gate + clean-tree assert (`a24496a`), v2.6.0 announcement draft +
DOMAIN_LANGUAGE draft (`c5e92d9`), FOUC E2E scenario (stack `c2220d3`),
the stack CRM+TURN surgery (stack `937b94f`+`be876ae`: crm options,
per-response TURN credentials, config.js shadow+timer retired), the CRM
restore-drill call_logged round-trip (crm `1358e6f`), and the
island→CRM wire-contract test (`29e0d95`). Session 3 (release-tail
close): stack browser E2E ×2 GREEN at `271f5ef` (195s + 184s — the
FOUC scenario costs ~15s, g3 answered: no budget bump; it took four
stack harness commits because the scenario had never run live:
cache-dodging hard reloads + in-page theme-tick counting, since
chromedriver is blind to mid-navigation state), v2.6.0 GitHub
release published (CHANGELOG body), release smoke 41+4 green with
`/version` exactly v2.6.0, go-etag-split vendorHash repair (`0a7a732`
— the daemon-swept dep bump had broken `nix build` on main for ~2h),
full `nix flake check` green incl. the KVM backup VM, and pbx-artmann
relock #4 + re-pin (`8104448`/`438c348` → `20b2a18`): probe green,
both toplevels green, webphone ExecStart moved 2.5.0→2.6.0.
New cycle 2026-09-24: library-utilization audit of the LarsArtmann stack
(`docs/research/2026-09-24_larsartmann-stack-deep-dive.html`, overall 61/100)
harvested below + full pareto plan at
`docs/planning/2026-09-24_13-33_SUPERB-stack-adoption-pareto.md`.

Closed 2026-09-26: the go-error-family ergonomics remainder (audit
finding #4, the last open slice of the error-family row):
`Family.HTTPStatus()` is now cited as the rejected alternative in the
`classifyForUser` comment (canonical Rejection→400 / Transient→503
mapping vs the pinned two-value 422/502 contract, verified against the
v0.10.2 source).
Also closed 2026-09-26, the post-heal gate battery (makes up the
outage-skipped legs of the 2026-09-25 sweep + self-review items
7/13): main is FULLY GREEN after the binfmt heal — `nix build
.#webphone` (2.7.0, no vendorHash drift), `nix flake check` all checks
passed incl. the KVM backup VM, loopback smoke 41+4, `nix run .#vulnix`
zero real advisories, buildflow full exit 0, erraudit tier 1 exit 0,
aarch64 cross-build ELF-verified (e_machine=183 — the healed
`/run/binfmt` path proven end-to-end); prod re-probed 2026-09-26 (19
passed / 0 failed, still v2.6.0). FEATURES.md gained the missing
templ-components adoption row; command sheet §4-§6 row-number refs
converted to stable names and §6 updated (v2.7.0 drafts exist since
`357ffec`).

| Task                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Status       | Priority | Effort | Evidence / notes                                                                                             |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | -------- | ------ | ------------------------------------------------------------------------------------------------------------ |
| v2.7.0 release tail + deploy (OWNER terminal — pbx-artmann AGENTS forbids assistant ssh/deploy): v2.7.0 is FOLDED but UNTAGGED (`78b30cf` fold, `5a689ef` version bump) — cut the signed tag, ride the runbook (stack lock bump → stack gates incl. the browser E2E the markup change owes → aarch64 → pbx-artmann relock #5 + re-pin → deploy; COMMAND: `cd ~/projects/pbx-artmann && nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech` → smoke → `nixos-rebuild switch`), then post-deploy `python3 scripts/webphone-smoke.py --base https://pbx.artmann.tech --expect-version 2.7.0`. NOTE: train C refuses self-sends LOCALLY — expect the 422 "messages cannot be sent to your own number" + failed row instead of the provider's 40310 text | 🔴 `TODO` | High | S | prod serves v2.6.0 today (probed 2026-09-25); ritual in `docs/release-runbook.md`; copy-paste commands in `docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md` |
| Outbound SMS bridge failure on prod: `POST /messages/send` answered 422 (gateway error) when probed 2026-09-19 — webphone-side classification is 502 + provider-rejection detail since `1d53f44`, and refusals answer 422 since `6ac8962` (`1d53f44` is live on prod since the 2.6.0 deploy; `6ac8962` rides 2.7.0), but the ROOT cause is stack-side (telnyx-webhooks bridge down/rejecting or Telnyx creds). OWNER (assistant ssh is forbidden by pbx-artmann AGENTS): journal the telnyx-webhooks unit for today, grep for sms/422/error lines, restart or fix creds per findings, send a test SMS; record the root cause here + in the stack runbook                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | 🔴 `TODO`    | High     | S      | 2026-09-19 probe; webphone-side fix `1d53f44`; refusal-422 `6ac8962`                                         |
| OWNER-calls batch session (decisions, one sitting — briefing at `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`): the original ~15 (train-cut/cadence, `/livez` consumer, HSTS, XFF sanitization, disclosure posture, loopback `delivered` semantics, handler dual-layer, gh-release habit, Go module v2 policy, recordings intent, TEMP-DIAG keep, oops ratification, `backup.retentionDays`) + the three CRM policy calls (typed NixOS `crm.{url,token}` options — SHIPPED as `crm.{enable,url,tokenFile}` in stack `be876ae`, ratify the shape; multi-contact "+N more"; English-only journal bodies) + self-send train-C semantics + templ-components history-blemish disposition (Tailwind spike RESOLVED 2026-09-24: wave 1 shipped, waves 2+3 rejected) + from the tail: g1 force-push ratification (one `--force-with-lease` on a self-authored 1-min-old pushed commit) + release.sh load-gate default ratification + g2 KVM-timeout policy (g3 CLOSED: no budget bump) + art-dupl `-t 3` baseline + suppression-bucket doc + webhook 400 body-text dependents + `msg/`→`message/` idem-key rename + helper micro-test bar + missed-call REJECT semantics + search `?q=` URL semantics + store `Must*` panic-on-corrupt policy | 🟡 `PLANNED` | High     | S      | ROADMAP "Open questions" carries the same list with context; questions route there, decisions land back here |
| Post the release announcements (drafts for v2.1.0–v2.3.0, v2.5.0, v2.6.0 AND v2.7.0 at `docs/announcements/`; owner picks channel(s), approves wording, decides the disclosure posture — fix-acknowledged, no exploit detail)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | 🟡 `PLANNED` | Low      | S      | v2.6.0 draft = `docs/announcements/2026-09-23_v2-6-0_drafts.md` (`c5e92d9`); v2.7.0 draft = `docs/announcements/2026-09-24_v2-7-0_drafts.md` (`357ffec`); v2.6.0 gh object verified published 2026-09-23      |
| Standing watches — QUARTERLY RE-CHECK, next due 2026-12-20 (named triggers fire earlier): sip.js 0.22 (swap ONLY on Chromium WebRTC breakage, a sip.js security advisory, or a needed capability — re-checked 2026-09-23: npm latest is still 0.21.2) · templ-components upstream v1.20.x watch (latest v1.19.3 — tag-only, no release notes; adopted v1.19.2, patch rideable at the next dep sweep; re-checked 2026-09-24, `sort -V` the tags) · oxlint globals (any new browser global → island oxlint.json) · E2E wall-time budget 445s — watch TWO consecutive over-budget runs (g3 ANSWERED 2026-09-23: green runs 195s/184s with the FOUC scenario aboard — no bump) · MONTHLY erraudit tier-2 re-measure next due 2026-10-22 (baseline 127 stdlib_constructor / 113 outside crm, must shrink) · nanoid CLOSED (v1.65.1)                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | 🟡 `PLANNED` | Low      | S      | AGENTS erraudit bar; re-checked 2026-09-23 + 2026-09-24 during the pareto cycle (C22/M22)                                     |
| templ-components adoption: DONE 2026-09-24. Spike GREEN (verdict `docs/planning/2026-09-24_16-38_tailwind-coexistence-verdict.md`); wave 1 SHIPPED (EmptyState ×6 true empty states, `/assets/tw.css` permanent 18.9KB build — regenerate on templ-components bumps, recipe in the verdict doc); wave 2 REJECTED on product grounds (RelativeTime: browser-locale strings break the per-extension language invariant + byte-stable pins + CSP kills its script; CountBadge: wrong shape for the `wp-nav-badge` pill); wave 3 errorpage KEEP (404 is shell-integrated, `.wp-error` is an htmx wire contract, family copy already lives server-side). spike route + probe REMOVED 2026-09-24 (`b7ed8fb`, zero refs after); stack browser E2E re-run rides the v2.7.0 release gate (markup changed)                                                     | 🟢 `IN PROGRESS` (release E2E pending) | Medium   | S      | verdict doc has data + recipe + dispositions |
| go-health `WithEvaluationHook` → health-check outcome counters on /metrics: RESOLVED-PARKED 2026-09-24 with evidence — the hook fires only inside `Probe.Evaluate`, which only the readiness handler / background refresh loop call; this wiring serves readiness via cqrs-htmx (no second readiness truth) and runs go-health live (refresh 0), so the counters would never fire (verified in v0.3.0 AND v0.4.0 sources; wiring was built, found dead, reverted in `f7028b4`). Revisit only if go-health hooks startup evaluations or cqrs-htmx grows a readiness hook. branded-id Valuer/Scanner stays PARKED at the store seam (adopt only on the next storage-format touch — verified: the library has no prefix-aware parser, so hand-rolled parseID stays) | ✅ `DONE` (parked with evidence) | Low | S | plan M18/M21; audit findings #5/#7; revert `f7028b4` |
