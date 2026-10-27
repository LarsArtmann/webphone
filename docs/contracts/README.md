# Wire contracts

The machine-to-machine contracts of the webphone binary live here as a
committed, regenerated artifact:

- `asyncapi.yaml` — the AsyncAPI 3.0 document declaring every lane the
  consuming stack and the native shells build against: the five
  secret-gated `/hooks/*` ingest lanes (ring, message, fax, fax/status,
  message/status), the gateway acceptance receipt (`{"provider_ref"}`),
  and the per-extension SSE stream kinds.

## One source of truth

The document is **generated from the handler wire types** (via
`go-cqrs-lite/catalog`) — the same Go structs the handlers decode. It is
never hand-edited; a payload change is a code change plus a regen:

```console
nix develop -c go run ./cmd/webphone contracts   # writes docs/contracts/asyncapi.yaml
```

The freshness gate in `nix flake check` regenerates and diffs — a dirty
tree means the committed contract rotted against the code.

## Consumers

- **pbx-artmann's `telnyx-webhooks.py` bridge** — the Telnyx webhook
  lanes and the receipt envelope it answers with.
- **The FreeSWITCH dialplan** — the `/hooks/ring` wake lane
  (ring-while-closed; the dialplan must never depend on push health, the
  lane answers 202 inertly when push is unconfigured).
- **webphone-android** — the push-device registry and the ring contract
  its CallStyle notification rides, plus the `draft_ready` push fact
  (M14): an internal-only device payload riding the SAME transports and
  registry (no HTTP surface of its own, hence nothing in either
  document — the payload shape is pinned by `internal/push` tests).

## The split rule

Browser-facing JSON belongs in the served **OpenAPI** document
(`/openapi.json`, `internal/server/apidocs.go`); machine-to-machine
lanes (`/hooks/*`, gateway receipt, SSE) belong in **this** AsyncAPI
document. One seam, one document, never both — a lane documented in the
wrong file is a bug, not redundancy. Related: the passkey/SSO surfaces
are **conditional** (registered only when `auth.passkey.*` / SSO is
configured, styled 404 otherwise) and therefore deliberately absent from
both documents — a disabled deployment must not promise a surface it
does not mount.

## One home each: why there are THREE catalog builders

webphone declares three catalogs and that is the ruling, not an
accident (2026-10-24): `internal/domain.BuildCatalog` (the EventCatalog
federation — domain-seam events; domain cannot import server, and the
hub consumes it), `internal/server/apidocs.go` (the served OpenAPI —
references session/api handler DTOs, so it MUST live in server), and
`internal/server/wirecontracts.go` (this document's AsyncAPI source —
references hooks + gateway wire types, likewise server-owned). A merge
would either invert an import edge or split-brain a lane; the shared
parts (title/version consts, the split rule) stay documented here.
