TL;DR: `validate.go` flags any 2xx response with a nil schema, including 204 No Content — which has no body by HTTP semantics — so a truthful 204 declaration fails validation unless the consumer lies with a fake empty-object schema. Ask: exempt 204 (and arguably 304) from the 2xx body-schema rule.

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`validate.go:170-176` (v4.9.0):

```go
for _, resp := range msg.Responses {
	if len(resp.StatusCode) == 3 && resp.StatusCode[0] == '2' && resp.Schema == nil {
		violations = append(violations, Violation{ ... "2xx response has no body schema" })
```

The status-class check does not exempt 204. webphone's DELETE and deregister endpoints answer 204; to pass validation each declares a typed empty object schema with the description "No response body" — the generated document then says the 204 has a JSON object body, which is wrong on the wire.

## Ask

Skip the body-schema violation when the status is `204` (bodies are prohibited there by RFC 9110 §6.4.1). A 204 should validate with `Schema == nil` and export as `{"description": "..."}` with no content key.

💘 Generated with Crush
