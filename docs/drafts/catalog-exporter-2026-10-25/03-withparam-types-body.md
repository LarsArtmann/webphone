TL;DR: `WithParam(name, location, description string, required bool)` has no type parameter, so every declared query/path parameter exports as an untyped string and consumer-side type docs (integer, boolean) cannot be expressed. Ask: add a typed variant (variadic option or a `WithParamType`) without breaking the existing signature.

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`message_config.go:153` (v4.9.0):

```go
func WithParam(name, location, description string, required bool) MessageOption {
```

All four knobs are strings/bool; the emitted parameter carries no `schema` (OpenAPI defaults it to string). webphone's pagination/limit-style params can't declare `integer` or an enum today.

## Ask

Either a `WithParamSchema(name, location string, schema *Schema)` companion, or a functional option on `WithParam`. The string signature can stay for compatibility; the typed escape hatch is the point.

💘 Generated with Crush
