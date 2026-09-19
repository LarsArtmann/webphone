# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in ROADMAP.md (raw ideas + open questions); owner calls there too.

| Task                                                                                                                                                                | Status    | Priority | Effort | Evidence / notes                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | --------------------------------------------------------------------------------------------------- |
| Re-run the stack browser E2E: island changed after the last green run (429 surfacing, toasts, live pill) — the accept/reject bug class is what only the E2E catches | 🔴 `TODO` | High     | M      | `docs/status/archived/2026-09-19_06-42_…` §c.3/§f.5; island commits `91d018c`, `d60c166`, `718cbe7` |
| Stack-side full `nix flake check` with the new webphone lock (FreeSWITCH/operator derivations unexercised since the bump) + re-run the stack's webphone VM test     | 🔴 `TODO` | High     | M      | 07:43 report §b.4/§f.3; 06:42 report §f.15                                                          |
| Redeploy the production vhost and eyeball the browser console on `pbx.artmann.tech` (the CSP fixes are verified locally + via the stack E2E, not in production)     | 🔴 `TODO` | Medium   | S      | 22:08 report §b.1/§f.1 (owner/ops action)                                                           |
