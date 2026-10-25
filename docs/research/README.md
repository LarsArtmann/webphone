# Research index

One row per library deep-dive in this directory (the `library-deep-dive`
skill's output home). Newest last matches the filename dates; each row
carries the verdict a session needs before re-litigating an adoption.
Plans and status reports live in `docs/planning/` and `docs/status/`.

| Report | Library | Verdict (one line) |
|--------|---------|--------------------|
| [2026-09-18 cqrs-htmx deep-dive](2026-09-18_cqrs-htmx-deep-dive.html) | cqrs-htmx | First-pass adoption audit; superseded by the 09-19 revision below. |
| [2026-09-19 cqrs-htmx deep-dive](2026-09-19_cqrs-htmx-deep-dive.html) | cqrs-htmx | Authoritative cqrs-htmx audit — the root library + usermgmt seams webphone rides (layout.Base, DebugHandler, Recovery, SSE broadcaster). |
| [2026-09-19 modernc-sqlite deep-dive](2026-09-19_modernc-sqlite-deep-dive.html) | modernc.org/sqlite | CGO-free SQLite validated as the storage substrate (webphone.db + usermgmt.db + readmodels.db). |
| [2026-09-22 templ-components deep-dive](2026-09-22_templ-components-deep-dive.html) | templ-components | Adopted `layout.Base` (+ `NoThemeScript`), `display.EmptyState`, icons; `display.RelativeTime`/`CountBadge`/errorpage REJECTED (coexistence verdict). |
| [2026-09-24 larsartmann-stack deep-dive](2026-09-24_larsartmann-stack-deep-dive.html) | LarsArtmann family stack | Fleet-wide library landscape; per-module adopt/decline calls for the consuming stack. |
| [2026-10-23 go-appkit deep-dive](2026-10-23_go-appkit-deep-dive.html) | go-appkit | NOT a dependency, stays that way — every module re-wraps what webphone composes directly; `errorpages` declined (styled 404 must ride the session shell). |
| [2026-10-24 catalog deep-dive](2026-10-24_go-cqrs-lite-catalog-deep-dive.html) | go-cqrs-lite/catalog | Audit found zero usage and two hand-rolled doc surfaces; the gap closed the SAME DAY (eventcatalog + OpenAPI + AsyncAPI lanes shipped) — read its Addenda section first. |
