TL;DR: `operationId` derives from `strings.ToLower(method) + caseutil.ToPascal(msg.ID)`, and `ToPascal` splits only on space/underscore/hyphen — so a dotted message id keeps its dots in the operationId, which OpenAPI codegen tools reject. Ask: treat `.` as a word separator in `caseutil.ToPascal` (or sanitize the emitted operationId).

go-cqrs-lite catalog/v4 v4.9.0, consumer LarsArtmann/webphone (2026-10-24 train).

## Evidence

`openapi/exporter.go:163` (v4.9.0): `OperationID: strings.ToLower(method) + caseutil.ToPascal(string(msg.ID))`, and `internal/caseutil/convert.go:76-78`:

```go
words := strings.FieldsFunc(s, func(r rune) bool {
	return r == ' ' || r == '_' || r == '-'
})
```

webphone's message ids are dotted (`commands.wire.hooks.fax`), so the generated operationIds read `postCommands.wire.hooks.fax` — generators that emit client method names from operationId choke on the dots.

## Ask

Add `.` to the `FieldsFunc` separator set. That single change turns the id into `postCommandsWireHooksFax` for every existing dotted-id consumer, and non-dotted ids are unaffected.

💘 Generated with Crush
