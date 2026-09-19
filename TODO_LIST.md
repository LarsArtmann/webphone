# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in FEATURES.md as WORTH_CONSIDERING.

| Task                                                                                                                                                        | Status    | Priority | Effort | Evidence / notes                                                                                          |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | --------------------------------------------------------------------------------------------------------- |
| Split `hookFaxStatus`'s error mapping — every store error answers 404 while the message counterpart distinguishes 404/500; mind the provider-retry contract | 🔴 `TODO` | Low      | S      | `docs/status/2026-09-19_00-05_cqrs-htmx-adoption-execution-status.md` §e.1; `internal/server/webhooks.go` |
