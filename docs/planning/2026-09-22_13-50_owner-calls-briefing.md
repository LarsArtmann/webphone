# Owner-Calls Briefing — the 14 Decisions (SUPERB T11)

- **Date:** 2026-09-22 13:50 CEST
- **Purpose:** one sitting, ~15 minutes. Every row blocks or shapes
  downstream work in the 20-year plan
  (`docs/planning/2026-09-22_12-02_SUPERB-20-year-durability.md`).
  Each row: the question, what it gates, options, MY recommendation.
  Say "go with the recommendation" per row or override — verdicts land
  in their owning docs the same session.

| #  | Decision | Gates | Options | Recommendation |
| -- | -------- | ----- | ------- | -------------- |
| 1  | **Train-cut**: cut v2.5.0 now (fold the [Unreleased] pile: styled 404, module options, VM test, htmx bundle, error-excellence, registration-loss rebuild) or hold per g2 cadence? | T6/T7 (release machine), announcement scope | (a) cut now, deploy once with everything; (b) deploy v2.4.0-chain only, cut later | **(a) cut now** — the pile IS a user-visible theme (reliability + error excellence); one deploy carries all; the g2 rule wants theme-trains, and this is one |
| 2  | **`/livez` consumer**: wire anything to it (monitoring/systemd) or keep as contract-only? | T18c startup contract | (a) systemd watchdog consumes it; (b) fleet scraper; (c) contract-only for now | **(c) contract-only**, revisit when a fleet scraper lands (the stack already fences `/healthz`) |
| 3  | **HSTS on prod**: enable `nginx.hsts` for pbx.artmann.tech? | stack config | (a) enable max-age 31536000; (b) short max-age first (31536000 → ratchet); (c) off | **(b)** short window first (e.g. 300s → 1d → 1y) — https-only is proven but a WSS mishap bricks the phone until header expiry |
| 4  | **XFF sanitization flip**: does the stack's nginx sanitize `X-Forwarded-For` enough to flip webphone limiter keys to `KeyExtractorFromClientIP`? | rate-limit fairness behind the proxy | (a) flip now; (b) verify + flip later; (c) keep peer-host keys | **(c→b)** keep safe default until the stack's XFF handling is explicitly audited (one grep in the stack vhost); then flip |
| 5  | **Disclosure posture** for announcements: name the forged-session hole publicly? | T9 announcements | (a) full detail; (b) fix-acknowledged, no exploit detail; (c) no security mention | **(b)** — single-tenant PBX, hole fixed before disclosure, detail only helps no one now |
| 6  | **18-49 review annotate scope**: full ANNOTATE pass or judgment-sample? | T16c docs kernel | (a) full; (b) f-items only | **(b) f-items only** — the 22:26/22:29/23:43/00:14/01:04 pattern worked; full passes cost hours for marginal rows |
| 7  | **Loopback `delivered`**: simulate `delivered` in loopback mode (like fax `transmitted`) or keep `sent` honest-terminal? | dev-mode fidelity | (a) simulate; (b) keep | **(b) keep** — loopback is a seam for wiring tests, not a demo of provider semantics |
| 8  | **Handler dual-layer**: keep `requireSession` helpers + `Sessions.Require` middleware both, or drop one? | auth surface simplicity | (a) keep both (defense in depth, pages render anonymously); (b) middleware only | **(a) keep** — the dual layer is pinned by contract tests; dropping buys nothing |
| 9  | **gh-release-object habit**: every tag gets a Release object (notes from CHANGELOG)? | T7, future releases | (a) always (release.sh already does it); (b) majors only | **(a) always** — it is already automated; the v2.3.0/v2.4.0 empty-body incident showed manual is worse |
| 10 | **Go module v2 policy**: `/v2` suffix on next major or NOT-DO? | future tags | (a) adopt `/v2` at next major; (b) NOT-DO record (flake-consumed app, nobody `go get`s it) | **(b) NOT-DO** — record it, stop asking |
| 11 | **Recordings product intent**: panel in webphone (consent/jurisdiction/access model) or leave at stack level? | T24 (major feature) | (a) build panel; (b) leave stack-level; (c) later | **(c) later, decide by 2026-Q4** — the stack records server-side already; the webphone panel needs an access model first |
| 12 | **TEMP-DIAG keep**: keep the stack E2E answer-phase dump (`b96d4c2` there)? | E2E noise/verdict hygiene | (a) keep always; (b) gate behind env; (c) revert | **(b) gate behind env** — it was diagnostic for the 1001 hunt; always-on adds noise |
| 13 | **Oops ratification**: samber/oops stays a documented non-fix, or order a staged-adoption plan? | erraudit tier 3 | (a) ratify non-fix; (b) adoption plan | **(a) ratify non-fix** — 2026-09-21 triage found 0 real findings; families are adopted |
| 14 | **`backup.retentionDays`**: want the module option (auto-prune old snapshots)? | T18a | (a) yes, 30d default; (b) yes, 90d; (c) no (manual) | **(a) yes, 30d default** — unbounded snapshot dirs are a slow-burn disk incident |

## Not-decisions (already DECIDED, listed so the batch stays one sitting)

- Stack rides webphone `main` + per-train lock bump (2026-09-20).
- pbx-artmann stays a rev pin (2026-09-21 policy; bumped today to
  `1a95a736`).
- Shell copy stays English (D3, 2026-09-20); island copy en/de.
- No Go-side telephony, no speculative sip.js swap (named triggers
  only), no Playwright while the stack E2E suffices.
