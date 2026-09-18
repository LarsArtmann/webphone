# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in FEATURES.md as WORTH_CONSIDERING.

| Task                                                                                                                                                | Status    | Priority | Effort | Evidence / notes                                                                     |
| --------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ------------------------------------------------------------------------------------ |
| Upstream switchover in nix-international-telephony (swap input to this service, WSS proxy, config.js → server config migration, browser E2E re-run) | 🔴 `TODO` | High     | M      | Report (f4); DOM contract asserted by `internal/server/server_test.go`               |
| Split `server_test.go` by concern (sse/webhooks/proxy/session) and dedupe the 14-line helper clone                                                  | 🔴 `TODO` | Medium   | M      | `internal/server/server_test.go` (jscpd: line 703 vs 618-631)                        |
| Tag v2.0.0 on GitHub so CHANGELOG compare/release links resolve (lychee 404 × 2)                                                                    | 🔴 `TODO` | Medium   | S      | `CHANGELOG.md`; buildflow lychee findings 2026-09-18                                 |
| Nixpkgs channel bump to clear 14 build-chain CVEs (incl. glibc CVE-2026-5450, 9.8)                                                                  | 🔴 `TODO` | Medium   | S      | vulnix findings, buildflow full run 2026-09-18                                       |
| sip.js 0.22 bump (evaluation report: docs/reviews/, gated on the upstream browser E2E re-run)                                                       | 🔴 `TODO` | Low      | M      | `./update.sh` repins; bundle-contract strings must survive; re-run upstream VM suite |
