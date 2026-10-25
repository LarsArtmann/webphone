TL;DR: the AsyncAPI export always fabricates a default `kafka` server, so an HTTPS-only producer's document ships with a broker it does not have; consumers suppress it by overwriting the server with empty strings. Ask: emit no server (or a named-protocol default that matches the lanes) unless the catalog declares one.

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`asyncapi/exporter.go:56` (v4.9.0) hardcodes `protocol: "kafka"` for the implicit server. webphone's machine lanes are plain HTTPS webhooks + SSE (`docs/contracts/asyncapi.yaml`); every export starts by suppressing the fabricated broker via the `WithServer(name, "", "")` escape hatch. Without that call the document claims a kafka broker that exists nowhere in the deployment.

## Ask

When the catalog defines no server, emit `servers: {}` (valid AsyncAPI 3) instead of inventing one — or key the default off the declared channel bindings. The `WithServer(..., "", "")` workaround would become unnecessary.

💘 Generated with Crush
