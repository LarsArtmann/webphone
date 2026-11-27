# library-policy upstream ask — crypto/sha1 exemption mechanism

Filed 2026-11-27 (pareto T08 deps train hit it). FILED: LarsArtmann/library-policy#80.

## Title

crypto/sha1 ban has no exemption path; coturn TURN REST credentials are HMAC-SHA1 by protocol

## Body

````markdown
> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: The compiled-in `crypto/sha1` ban has no per-policy or per-path exemption, so a repo whose one SHA-1 use is the coturn REST credential scheme — which validates exactly `base64(HMAC-SHA1(secret, username))` — can neither suppress the finding nor demote it below error severity. Ask: a justified-exemption mechanism (per-policy path allowlist, or an inline directive the scanner honors), or a protocol exception for coturn TURN REST in the compiled security policy.

## Problem

BuildFlow's `library-policy` provider (restored 2026-11-23 as a v3 provider) runs the external binary with the policy set compiled in — `policies/security.go:68` bans `crypto/sha1` at error severity, and there is no on-disk config that can amend a compiled set (a repo-side `library-policy.yaml` replaces the whole set rather than extending it).

`webphone` derives coturn "use-auth-secret" TURN credentials at `internal/server/platform/configjs.go:117` (`turnRESTCredentials`):

```go
mac := hmac.New(sha1.New, []byte(secret))
```
````

coturn validates exactly `base64(HMAC-SHA1(secret, unix-expiry-username))`; SHA-256 would break every TURN server behind that config. The finding is real as written — the import IS `crypto/sha1` — but it is protocol-mandated, and none of the existing suppression paths reach it:

- `//nolint:library-policy` at the site: no effect (verified live — the scanner is import-pattern based and does not parse comments).
- BuildFlow's `exclude` list: cannot reach the external binary — it walks the workdir itself (verified live: the finding survived an exact-path exclude entry).
- BuildFlow config: only `error_rules` exists, which elevates; there is no per-rule demotion key.
- A repo-side `library-policy.yaml`: would restate the entire compiled policy set and drift against it.

So the only working suppression today is skipping the whole tool in `.buildflow.yml` `skip_steps`, which blinds every other policy (json v2, testify, Charm) repo-wide.

## Suggestion

A scoped escape hatch, any one of:

1. A per-policy `exempt_paths` (or `exempt_globs`) field on `PolicySpec`, matched against the finding's file path.
2. An inline directive the scanner honors, e.g. `// library-policy:ignore(crypto/sha1) <reason>` on or above the flagged import — the same shape the golangci-style linters already use.
3. A compiled-in protocol exception for the coturn TURN REST credential shape (`hmac.New(sha1.New, ...)` feeding a base64 credential beside a unix-expiry username), if a scoped mechanism is not wanted.

Any of the three lets webphone un-skip the tool (currently skipped with a dated reason and this issue as the reopen trigger in `.buildflow.yml`).

Environment: library-policy `policies/security.go` at `321255c`; BuildFlow `ff4aaa4`; webphone at `b95db765`.

💘 Generated with Crush

```
```
