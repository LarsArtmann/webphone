TL;DR: the OpenAPI exporter can only emit `application/json` request bodies — there is no way to declare a multipart upload, so file-upload endpoints are invisible to the generated contract. Ask: support `multipart/form-data` content (a `WithRequestBodyContent`/media-type option is enough).

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

Request bodies are built exclusively through `jsonContentWithExample` (`openapi/exporter.go:125`), and `MediaType` (types.go:60-63) carries only schema/example — there is no encoding or multipart shape anywhere in the openapi package. webphone's `/api/transcribe` takes a multipart audio upload; it is absent from the generated `/openapi.json` today (a JSON-only body would be a lie, so the endpoint is left out).

## Ask

A media-type-aware body option (`WithBodyContent(mediaType, schema)`) or a multipart helper that maps named file/string parts to the OpenAPI multipart schema shape. Scope: OpenAPI only.

💘 Generated with Crush
