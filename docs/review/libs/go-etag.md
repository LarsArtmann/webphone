# go-etag — library utilization audit

**Verdict:** the best adoption candidate among the indirect-only libraries: webphone hand-rolled RFC 7232 conditional-request logic for embedded assets while the library ships exactly that. **Adoption: LANDED 2026-11-04, BOTH halves — option A (`assetHeaders` rides `entitytag.NewETag` + `MatchesIfNoneMatch`; the hand-rolled `ifNoneMatch` deleted) AND option B (`etag.New(etag.DefaultETagConfig())` mounted on `/openapi.json` + `/manifest.webmanifest` only — the mount is the Skip set, `TestConditionalGETRootDocuments` pins 304/weak-candidate/stale/fence incl. config.js staying bare).**
**Pin:** go-etag/entitytag v0.6.1 + go-etag/server v0.6.1 (BOTH direct since 2026-11-04) · **Latest:** entitytag submodule v0.6.1 (the repo-root v0.7.0 tag did NOT include an entitytag bump — the CHANGELOG drift T20 files upstream) · **Currency:** current at the submodule · **Status:** direct dep (entitytag + server)
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Modules | `go-etag/server` (middleware + config), `go-etag/entitytag` (tag type, parse/compare) |
| In webphone | direct: `entitytag` (assets.go) + `server` (server.go route table, TWO mounts) |

## What the library offers

`server`: `etag.New(ETagConfig)` middleware — generate strong/weak tags (FNV-64a default or custom hash), answer 304 on If-None-Match, buffering with `MaxBufferSize`, `SkipIfPresent`, per-request `Skip func(*http.Request) bool`, On304/OnETagGenerated hooks; Flush/Hijack-aware streaming. `entitytag`: `ETag` type with `StrongEqual`/`WeakEqual`, `ParseETag`/`ParseETagList`, `MatchesIfMatch`. `client`: caching `RoundTripper` (`FreshenOn304`, LRU `MaxEntries`) — no webphone use case today.

## Webphone overlap (verified)

- **Hand-rolled conditional GET for embedded assets**, internal/server/assets.go: `assetHeaders` computes a strong sha256 content ETag (:93-115) and `ifNoneMatch(header, etag)` (:126) does a hand-written list-compare of `If-None-Match` values; `/manifest.webmanifest` inherits the same via `serveEmbedded`.
- Everything else is deliberate no-validator policy: `/config.js` no-store, session/passkey/sso/csrf no-store, SSE no-cache. (`/openapi.json` was no-cache-only at audit time (apidocs.go:425); it now carries the middleware ETag per the landed Option B above — the `no-cache` directive itself stays, so the fetch still revalidates, just bodylessly on a match.)
- The htmx-ext composite ETag comes from cqrs-htmx (which itself rides this library).

## Fit analysis

| Option | What it buys | Cost / risk |
|---|---|---|
| **A. Swap `ifNoneMatch` for `entitytag.ParseETagList` + `MatchesIfMatch`** | spec-exact list semantics (`*`, W/ prefix handling) for ~10 deleted lines; zero new runtime surface (module already in graph) | smallest; almost pure win |
| **B. Full `etag.New` middleware for `/openapi.json` + `/manifest.webmanifest`** | conditional GETs on surfaces that re-serve big-ish static bytes; free 304s for polling clients | must `Skip` session/credential/SSE surfaces — a `Skip` misconfiguration would cache-gate an auth surface; test cost |
| **C. Client transport** | — | no webphone HTTP-client caching need (server-to-server calls are CRM/paperless/pbx APIs) |

## Ruling

Option A AND option B both LANDED 2026-11-04. The B mount is deliberately minimal: only `/openapi.json` (the generated contract revalidates with `no-cache` on every consumer fetch — the ETag makes revalidation free) and `/manifest.webmanifest` (same stable-per-build property) are wrapped; session/credential/SSE no-store surfaces are structurally untouched because the middleware never sees them — no `Skip` misconfiguration is possible when the mount is two lines in the route table. Do NOT widen the mount to any session/config surface — the no-store doctrine is a security property (error-contract.md). `/version` deliberately stays bare (a one-line string needs no revalidation mechanism).

**Refresh trigger:** an assets train; a consumer complaining about `/openapi.json` size; the indirect pin bumping anyway via httputil/cqrs-htmx.

## How to verify

```console
rg -n "ifNoneMatch" internal/server/assets.go            # expect 0 (deleted 2026-11-04)
rg -n "entitytag" internal/server/assets.go              # expect the NewETag + MatchesIfNoneMatch use
rg -n "conditionalGET" internal/server/server.go         # expect 4: 1 comment + 1 construction + 2 wraps (openapi.json, manifest)
rg -n "go-etag" go.mod                                   # BOTH modules DIRECT since 2026-11-04
git -C /home/lars/projects/go-etag describe --tags --abbrev=0   # v0.7.0 vs pin v0.6.1
```
