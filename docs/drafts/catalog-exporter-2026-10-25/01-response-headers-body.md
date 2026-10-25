TL;DR: the OpenAPI Response type has no `headers` field, so a header-carrying contract like `429 + Retry-After` can only be expressed as prose in the description. Ask: add `Headers map[string]HeaderParam` to `Response` and render it.

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`openapi/types.go:55-58` (v4.9.0):

```go
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}
```

No headers. webphone's rate-limited surfaces have a real `Retry-After` contract (limiter middleware sets the header on every 429), and the generated `/openapi.json` has nowhere to put it — the agreement currently lives as a sentence inside the 429 response description, where codegen cannot see it.

## Ask

A `Headers` field on `Response` (rendered into the OpenAPI `headers` object, same shape as `Parameter` minus `location`) plus a `WithResponseHeader(...)`-style option or a `ResponseOption`. Webphone would move its Retry-After rows from description text to the typed field the moment it lands.

Scope: OpenAPI only; the AsyncAPI side has no response-header concept.

💘 Generated with Crush
