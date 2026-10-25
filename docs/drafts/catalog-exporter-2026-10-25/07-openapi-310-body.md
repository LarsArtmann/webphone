TL;DR: the exporter pins `openAPIVersion = "3.0.3"`, so documents cannot opt into OpenAPI 3.1 (jsonSchema 2020-12, nullable via type arrays, richer examples) even when the generator's output is already compatible. Ask: an exporter option for 3.1.0 emission, 3.0.3 default.

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`openapi/exporter.go:14` (v4.9.0): `openAPIVersion = "3.0.3"`, documented as the target in `doc.go:1`. No option exists to change it.

## Ask

A low-risk path: an `WithOpenAPIVersion("3.1.0")`-style option that only swaps the version string (3.0 documents are largely valid 3.1), with any genuinely incompatible constructs gated per version later. Consumers pinned to 3.0 tooling keep today's default.

💘 Generated with Crush
