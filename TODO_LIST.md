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
| Redeploy production with webphone v2.1.0, then eyeball the browser console on `pbx.artmann.tech` after a login/call — URGENT: the deployed v2.0.0-era build 403s every browser login behind the TLS proxy (csrf attestation bug, fixed in v2.1.0), so tab sessions never open; server-side probes pass because they skip CSRF POSTs | 🔴 `TODO` | High | S | stack E2E found it 2026-09-19; fixed by `csrf.trusted_*` (v2.1.0, `d815004`); 22:08 report §b.1/§f.1 |
