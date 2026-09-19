# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in ROADMAP.md (raw ideas + open questions); owner calls there too.

| Task                                                                                                                                                            | Status    | Priority | Effort | Evidence / notes                          |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ----------------------------------------- |
| Cut v2.1.0 per the release runbook (fold CHANGELOG, tag, lychee, stack bump) — HEAD reports `v2.0.0` while dozens of commits newer; `/version` drifts until then | 🔴 `TODO` | Medium   | M      | 11:02 report §f.11; owner ceremony        |
| CSRF token rotation on login: island must adopt a fresh token post-login (no reload) before `InvalidateCSRFCookie` can ship | 🟡 `PLANNED` | Medium | M | AGENTS CSRF constraint (2); superb-adoption plan residue |
| idiomorph swap experiment for SSE/HTMX partials, gated on the stack's browser E2E | 🟡 `PLANNED` | Low   | M      | cqrs-htmx superb-adoption plan (open item) |
| Owner decision: stack pin policy (riding webphone `main` vs release tags) — live PBX runs unreleased commits today | 🔴 `TODO` | Medium   | S      | 11:02 report §g.1                         |
| Owner decision: pbx-artmann `path:` vs github input for the stack (the live path input burned them once) | 🔴 `TODO` | Medium   | S      | 11:02 report §g.2/§f.28                   |
| Inbound fax feed: convert the stack's rxfax TIFFs to PDF → webphone `/hooks/fax` once messaging is live | 🔴 `TODO` | Low      | M      | SUPERB plan deferred item                 |
| Redeploy the production vhost and eyeball the browser console on `pbx.artmann.tech` (the CSP fixes are verified locally + via the stack E2E, not in production) | 🔴 `TODO` | Medium   | S      | 22:08 report §b.1/§f.1 (owner/ops action) |
