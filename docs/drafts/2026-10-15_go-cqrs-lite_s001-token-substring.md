> [!NOTE]
> This filing was drafted by glm-5.3-flash via Crush executing the 2026-10-15 metaengine-train gate run, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: S001 flags any identifier containing `token`/`key`/`secret`/… that is bound to a string literal ≥ 8 chars — so ordinary storage NAMES (`enrollTokensCollection = "enroll_tokens"`) come out as `critical` findings that gate builds, and the "use environment variables" remediation is unactionable for a `const` schema name. Ask: allowlist name-shaped values (pure snake_case identifiers) and skip `const` declarations where the suggestion cannot apply.

cqrs-lint v4.13.5 (RULES.md S001, `hardcoded-secrets`, critical/medium); every cite re-verified at `b0cf0ee01` today.

## Problem

The detector matches a fixed substring list against the lowercased identifier and reports the string literal's position:

```go
// cmd/cqrs-lint/pkg/rules/security/rules.go:37-45 (NewS001Detector)
secretFields := []string{
	"secret",
	"password",
	"passwd",
	"apikey",
	"api_key",
	"token",   // <-- matches any identifier CONTAINING "token"
	...
}
```

```go
// rules.go:55,59 — the only value-side escapes
if len(val) < 8 { return }
if lintutil.IsURLOrPlaceholder(val) { return }
```

Real hit from my repo (LarsArtmann/webphone, `internal/readmodels/enroll.go:24`) — a metaengine collection name, i.e. a SQLite table identifier:

```go
const enrollTokensCollection = "enroll_tokens"
// Potential hardcoded secret in field "enrollTokensCollection" —
// use environment variables or a secret manager   (critical, medium)
```

Nothing secret exists on that line — the value is the storage schema's own name — but `critical` severity trips the buildflow findings gate, so the consumer must suppress the site to ship.

## Impact

Every `*Collection`, `*TableName`, `*QueueName`-shaped identifier whose name happens to contain a trigger substring is a build-gating critical. The suggested fix ("load from environment variables") cannot be applied at all: a `const` used as a schema name must be a compile-time constant.

## Proposal

Two independent reductions, either alone kills this class:

1. **Value-shape allowlist (pre-collision):** skip literals that are pure snake/lowercase identifiers (`^[a-z][a-z0-9_]*$`). Secret material carries entropy (mixed case, digits, symbols, length); name-shaped values are schema vocabulary. `IsURLOrPlaceholder` already establishes the allowlist pattern at `rules.go:59` — this is one more clause beside it.
2. **Const-aware downgrade (at-failure):** when the binding is a `const` declaration, the "use env vars" suggestion is unactionable — downgrade severity (or reword the suggestion) so a true positive on a const still surfaces without gate-critical noise on names.

I like the rule's intent; this is only about the substring net's bycatch.

Affects `cmd/cqrs-lint` at `b0cf0ee01`; consumer: LarsArtmann/webphone on Go 1.27.

💘 Generated with Crush
