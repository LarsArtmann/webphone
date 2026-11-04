# go-etag — library utilization audit

**Verdict:** the best adoption candidate among the indirect-only libraries: webphone hand-rolls RFC 7232 conditional-request logic for embedded assets while the library ships exactly that (plus a client-side caching transport webphone doesn't need). **Adoption: n/a (indirect-only); fit: MEDIUM.**
**Pin:** go-etag/entitytag + go-etag/server v0.6.1 (go.mod:100-101, `// indirect`) · **Latest:** v0.7.0 (local checkout tag) · **Currency:** one minor behind at the indirect pin · **Status:** indirect-only — pulled by httputil (root) and cqrs-htmx
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Modules | `go-etag/server` (middleware + config), `go-etag/entitytag` (tag type, parse/compare) |
| In webphone | indirect only; **zero imports** |

## What the library offers

`server`: `etag.New(ETagConfig)` middleware — generate strong/weak tags (FNV-64a default or custom hash), answer 304 on If-None-Match, buffering with `MaxBufferSize`, `SkipIfPresent`, per-request `Skip func(*http.Request) bool`, On304/OnETagGenerated hooks; Flush/Hijack-aware streaming. `entitytag`: `ETag` type with `StrongEqual`/`WeakEqual`, `ParseETag`/`ParseETagList`, `MatchesIfMatch`. `client`: caching `RoundTripper` (`FreshenOn304`, LRU `MaxEntries`) — no webphone use case today.

## Webphone overlap (verified)

- **Hand-rolled conditional GET for embedded assets**, internal/server/assets.go: `assetHeaders` computes a strong sha256 content ETag (:93-115) and `ifNoneMatch(header, etag)` (:126) does a hand-written list-compare of `If-None-Match` values; `/manifest.webmanifest` inherits the same via `serveEmbedded`.
- Everything else is deliberate no-validator policy: `/config.js` no-store, session/passkey/sso/csrf no-store, SSE no-cache, `/openapi.json` no-cache (apidocs.go:425).
- The htmx-ext composite ETag comes from cqrs-htmx (which itself rides this library).

## Fit analysis

| Option | What it buys | Cost / risk |
|---|---|---|
| **A. Swap `ifNoneMatch` for `entitytag.ParseETagList` + `MatchesIfMatch`** | spec-exact list semantics (`*`, W/ prefix handling) for ~10 deleted lines; zero new runtime surface (module already in graph) | smallest; almost pure win |
| **B. Full `etag.New` middleware for `/openapi.json` + `/manifest.webmanifest`** | conditional GETs on surfaces that re-serve big-ish static bytes; free 304s for polling clients | must `Skip` session/credential/SSE surfaces — a `Skip` misconfiguration would cache-gate an auth surface; test cost |
| **C. Client transport** | — | no webphone HTTP-client caching need (server-to-server calls are CRM/paperless/pbx APIs) |

## Ruling

Option A is a tidy-up worth doing in the next assets-touching train; Option B is a genuine small feature (fresher `/openapi.json` for consumers without a full re-download) gated on a careful `Skip` set + httpspec coverage. Either adoption should ride the **v0.7.0** bump (one minor, already released upstream). Do NOT let the middleware anywhere near session/config surfaces — the no-store doctrine is a security property (error-contract.md).

**Refresh trigger:** an assets train; a consumer complaining about `/openapi.json` size; the indirect pin bumping anyway via httputil/cqrs-htmx.

## How to verify

```console
rg -n "ifNoneMatch" internal/server/assets.go            # the hand-roll (:126)
rg -rn "go-etag" --type go internal/ | wc -l             # expect 0 direct imports
rg -n "go-etag" go.mod                                   # indirect pins
git -C /home/lars/projects/go-etag describe --tags --abbrev=0   # v0.7.0 vs pin v0.6.1
```
