# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in FEATURES.md as WORTH_CONSIDERING.

| Task                                                                                                                                                                         | Status    | Priority | Effort | Evidence / notes                                                                                            |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| Upstream switchover in nix-international-telephony (swap input to this service, WSS proxy, config.js → server config migration, browser E2E re-run)                          | 🔴 `TODO` | High     | M      | Report (f4); DOM contract asserted by `internal/server/server_test.go`                                      |
| Tag v2.0.0 on GitHub so CHANGELOG compare/release links resolve (lychee 404 × 2)                                                                                             | 🔴 `TODO` | Medium   | S      | `CHANGELOG.md`; buildflow lychee findings 2026-09-18                                                        |
| Track glibc CVE-2026-5450 (9.8) until nixpkgs ships a fix — nixpkgs bumped to 2026-09-17 and it is still unfixed; runtime closure is otherwise clean (8 derivations, vulnix) | 🔴 `TODO` | Medium   | S      | vulnix runtime-closure scan 2026-09-18; build-time-only advisories (binutils/gcc/bootstrap Go) never deploy |
