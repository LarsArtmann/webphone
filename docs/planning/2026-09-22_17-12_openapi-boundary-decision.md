# OpenAPI boundary — decision record

- **Date:** 2026-09-22 17:12 CEST
- **Plan:** SUPERB T27f ("OpenAPI boundary decision record")
- **Decision:** `/openapi.json` documents the JSON API surface ONLY —
  the machine-consumed, versioned-contract endpoints the island calls.
  It deliberately does NOT attempt to describe the whole HTTP surface.

## In scope (documented, spec-vs-handler tested)

- `POST /api/session`, `GET /api/session` (island session mint /
  resume)
- `GET /api/csrf` (post-login token adoption)
- `GET`/`POST`/`DELETE /api/contacts` (personal contacts store)

The rule for additions: a JSON endpoint the ISLAND (or any scripted
client) consumes belongs in the spec, with its handler wired in the
spec-vs-handler test (`internal/server/middleware_test.go`).

## Out of scope (deliberately, with their homes)

- **htmx fragment routes** (`/messages`, `/partials/*`, form posts…):
  they are UI plumbing, not an API — request bodies are form-encoded,
  responses are HTML fragments (or swaps signaled via `HX-Trigger`).
  Their contract is the DOM contract (docs/dom-contract.md) + the
  failure→feedback map (docs/error-contract.md), pinned by Go render
  tests.
- **Provider webhooks** (`/hooks/*`): provider-facing; their payload
  contracts are documented in `internal/server/webhooks.go` doc
  comments + the stack's bridge plan docs; the security boundary is
  the shared secret (fail-closed), not an API schema.
- **Health probes** (`/healthz`, `/livez`, `/startupz`): ops surface;
  documented in README (monitoring section) and consumed by the
  module's nginx locations.

## Why this boundary

A spec that half-describes the fragment routes would be a LYING
document — worse than none. The JSON island API is the only surface
where a schema is a promise to a program rather than a description of
UI; that is what OpenAPI is for.
