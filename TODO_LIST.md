# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in ROADMAP.md (raw ideas + open questions); owner calls there too.

| Task                                                                                                                                                            | Status    | Priority | Effort | Evidence / notes                          |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ----------------------------------------- |
| idiomorph swap experiment for SSE/HTMX partials, gated on the stack's browser E2E | 🟡 `PLANNED` | Low   | M      | cqrs-htmx superb-adoption plan (open item) |
| Owner decision: stack pin policy (riding webphone `main` vs release tags) — live PBX runs unreleased commits today | 🔴 `TODO` | Medium   | S      | 11:02 report §g.1                         |
| Owner decision: pbx-artmann `path:` vs github input for the stack (the live path input burned them once) | 🔴 `TODO` | Medium   | S      | 11:02 report §g.2/§f.28                   |
| Inbound fax feed: convert the stack's rxfax TIFFs to PDF → webphone `/hooks/fax` once messaging is live | 🔴 `TODO` | Low      | M      | SUPERB plan deferred item                 |
| Eyeball the browser console on `pbx.artmann.tech` after a login/call (the CSP fixes + the whole v2 stack were redeployed to production 2026-09-19 and verified server-side: healthz, config.js, hooks 401/202, bridge self-tests — only the in-browser eyeball remains) | 🔴 `TODO` | Low   | S      | 22:08 report §b.1/§f.1; redeploy + server-side probes done 2026-09-19 (SUPERB plan T12/T13) |
