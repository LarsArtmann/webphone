# go-etag — library utilization audit

**Verdict:** the best adoption candidate among the indirect-only libraries: webphone hand-rolled RFC 7232 conditional-request logic for embedded assets while the library ships exactly that. **Adoption: LANDED 2026-11-04, BOTH halves — option A (`assetHeaders` rides `entitytag.NewETag` + `MatchesIfNoneMatch`; the hand-rolled `ifNoneMatch` deleted) AND option B (`etag.New(etag.DefaultETagConfig())` mounted on `/openapi.json` + `/manifest.webmanifest` only — the mount is the Skip set, `TestConditionalGETRootDocuments` pins 304/weak-candidate/stale/fence incl. config.js staying bare).**
**Pin:** go-etag/entitytag v0.6.2 + go-etag/server v0.6.2 (both direct; bumped from v0.6.1 on 2026-11-18, commit 087c14d4) · **Latest:** entitytag v0.6.2 / server v0.6.2 — CURRENT at both submodules · **Currency:** current · **Status:** direct dep (entitytag + server)
**Root-tag drift, resolved:** the go-etag#5 drift filed 2026-11-04 (root v0.7.0 tagged with no cut CHANGELOG section; Release run 41773058677 failed as designed) is closed — the [0.7.0] section is cut, and root v0.7.1 (2026-11-18) fired release.yml end to end for the first time (run 44318110155, Release Latest — upstream TODO_LIST row 4). v0.7.1 was released as ONE staircase with the nested train v0.6.2; the root tag numbering one minor ahead of the nested modules is by-design (the deprecated root shim rides the nested train and stays one minor ahead until v1.0.0 deletes it — upstream [0.7.1] Changed). webphone consumes neither the root module nor the shim.
**Audited:** 2026-11-04 · **2026-11-22 re-verify** — v0.6.1 → v0.6.2 treadmill receipted below; verdict unchanged. · **Prior reports:** none (first audit).

## Status

|             |                                                                                       |
| ----------- | ------------------------------------------------------------------------------------- |
| Modules     | `go-etag/server` (middleware + config), `go-etag/entitytag` (tag type, parse/compare) |
| In webphone | direct: `entitytag` (assets.go) + `server` (server.go route table, TWO mounts)        |

## What the library offers

`server`: `etag.New(ETagConfig)` middleware — generate strong/weak tags (FNV-64a default or custom hash), answer 304 on If-None-Match, buffering with `MaxBufferSize`, `SkipIfPresent`, per-request `Skip func(*http.Request) bool`, On304/OnETagGenerated hooks; Flush/Hijack-aware streaming. `entitytag`: `ETag` type with `StrongEqual`/`WeakEqual`, `ParseETag`/`ParseETagList`, `MatchesIfMatch`. `client`: caching `RoundTripper` (`FreshenOn304`, LRU `MaxEntries`) — no webphone use case today.

## Webphone overlap (verified)

- **Hand-rolled conditional GET for embedded assets**, internal/server/platform/assets.go (moved from internal/server/assets.go by the 2026-11-12 platform carve): `assetHeaders` computes a strong sha256 content ETag and the entitytag swap replaced the old `ifNoneMatch(header, etag)` hand-roll; `/manifest.webmanifest` inherits the same via `serveEmbedded`.
- Everything else is deliberate no-validator policy: `/config.js` no-store, session/passkey/sso/csrf no-store, SSE no-cache. (`/openapi.json` was no-cache-only at audit time (apidocs.go, now :485); it now carries the middleware ETag per the landed Option B above — the `no-cache` directive itself stays, so the fetch still revalidates, just bodylessly on a match.)
- The htmx-ext composite ETag comes from cqrs-htmx (which itself rides this library).

## Fit analysis

| Option                                                                          | What it buys                                                                                                                  | Cost / risk                                                                                                         |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| **A. Swap `ifNoneMatch` for `entitytag.ParseETagList` + `MatchesIfMatch`**      | spec-exact list semantics (`*`, W/ prefix handling) for ~10 deleted lines; zero new runtime surface (module already in graph) | smallest; almost pure win                                                                                           |
| **B. Full `etag.New` middleware for `/openapi.json` + `/manifest.webmanifest`** | conditional GETs on surfaces that re-serve big-ish static bytes; free 304s for polling clients                                | must `Skip` session/credential/SSE surfaces — a `Skip` misconfiguration would cache-gate an auth surface; test cost |
| **C. Client transport**                                                         | —                                                                                                                             | no webphone HTTP-client caching need (server-to-server calls are CRM/paperless/pbx APIs)                            |

## Ruling

Option A AND option B both LANDED 2026-11-04. The B mount is deliberately minimal: only `/openapi.json` (the generated contract revalidates with `no-cache` on every consumer fetch — the ETag makes revalidation free) and `/manifest.webmanifest` (same stable-per-build property) are wrapped; session/credential/SSE no-store surfaces are structurally untouched because the middleware never sees them — no `Skip` misconfiguration is possible when the mount is two lines in the route table. Do NOT widen the mount to any session/config surface — the no-store doctrine is a security property (error-contract.md). `/version` deliberately stays bare (a one-line string needs no revalidation mechanism).

**Refresh trigger:** an assets train; a consumer complaining about `/openapi.json` size; the indirect pin bumping anyway via httputil/cqrs-htmx.

**2026-11-22 re-verify** — both pins moved v0.6.1 → v0.6.2 on 2026-11-18 (webphone commit 087c14d4, the train-4-era treadmill). Upstream's own [0.6.2] section: `entitytag` and `client` are internal-layout re-tags (per-topic test-suite splits) with zero public-API change; `server` and `metrics` pick up go-error-family v0.11.0 → v0.12.0 (nil-receiver-safe `Unwrap`, no API change); go directives settled back to minor-form `go 1.27`. For webphone's consumed surface — `entitytag.NewETag`/`MatchesIfNoneMatch` (internal/server/platform/assets.go:121-124) and the two `etag.New` wraps (server.go:726, :872) — it is a no-code treadmill: pins move, bytes don't. The cqrs-htmx train receipt says the same ("go-directive floor + release scripts only, no code"). Verify-block replay today: `ifNoneMatch` still 0, `entitytag` uses intact, `conditionalGET` still 4, both modules direct at go.mod:34-35. No verdict movement.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "ifNoneMatch" internal/server/platform/assets.go       # expect 0 (deleted 2026-11-04; file moved to platform/ 2026-11-12)
rg -n "entitytag" internal/server/platform/assets.go         # expect the NewETag + MatchesIfNoneMatch use
rg -n "conditionalGET" internal/server/server.go         # expect 4: 1 comment + 1 construction + 2 wraps (openapi.json, manifest)
rg -n "go-etag" go.mod                                   # BOTH modules DIRECT (v0.6.2, lines 34-35)
git -C /home/lars/projects/go-etag describe --tags --abbrev=0   # v0.7.1 — root tag only; the nested entitytag/server tags are v0.6.2 (by-design staircase)
```
