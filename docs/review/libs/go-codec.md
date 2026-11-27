# go-codec — library utilization audit

_Produced by C (subagents 6-9) of the 2026-11-22 full-fleet re-audit's fan-out (runbook step 7); every other line is the audit's own._

**Verdict:** NO-GO for direct PROD adoption stands — it is the event-store payload codec layer, and webphone's HTTP/JSON surfaces are not codec territory. The journal's wire vocabulary is now imported TEST-side to pin retention-tombstone replay compatibility. **Adoption: n/a (prod indirect by design + one test-side import).**
**Pin:** v0.3.3 (direct require since 2026-11-11 — the test import below) · **Latest:** v0.3.3 · **Currency:** current · **Status:** direct (test-only), prod stays indirect (a direct require of every cqrs-htmx module and most go-cqrs-lite modules — metaengine/v4/encoded.go:12 is the engine's import webphone rides; webphone pins metaengine/v4 v4.26.0, go.mod:28)
**Audited:** 2026-11-04 · **Updated:** 2026-11-11 (the D014 tombstone wire pin) · **2026-11-22 re-verify** — pin still v0.3.3 = latest; [Unreleased] CI-only; verdict unchanged. · **Prior reports:** none (first audit).

## Status

|             |                                                                                                                                           |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Module      | `github.com/larsartmann/go-codec`                                                                                                         |
| In webphone | prod: indirect only, zero imports; **test: 1 import** (internal/cqrs/tombstone_wire_test.go — `codec.EncodingCBOR` / `codec.ForEncoding`) |

## What the library offers

`Codec{Encoding(), Encode, Decode}` protocol over canonical CBOR / JSON / Raw; `DeterministicCodec` (signing-safe marker); strict compact CBOR (unknown-field rejection); zero-alloc buffer encoding; `ObservableCodec` (metrics/hooks); envelope helpers (`WrapEncode`/`UnwrapDecode`), `TranscodeToJSON`, streaming. It is the encoding seam that go-cqrs-lite stores, signing and encryption consume.

## Webphone overlap analysis

| Surface                                         | What webphone does                                                       | Codec relevance                                                                                                                                                     |
| ----------------------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP JSON DTOs (internal/server/api/*, apidocs) | `encoding/json`                                                          | none — deterministic CBOR targets store/signing bytes, not wire APIs                                                                                                |
| Gateway multipart forms                         | `go-playground/form` + honest part headers (byte-pinned golden)          | none                                                                                                                                                                |
| Engine persistence                              | metaengine/sqliteengine encodes payloads **through go-codec internally** | transitive benefit already realized: webphone's read-model views and event payloads persist in the engine's canonical encoding without webphone touching the module |
| Matrix/signal/chat receivers                    | bespoke JSON against upstream protocols                                  | none — wire compatibility owns the format                                                                                                                           |

**No hand-rolled codec exists to replace** — `rg "go-codec|msgp|CBOR" internal/` shows no parallel implementation.

## The test-side import (2026-11-11, D014 triage)

`internal/cqrs/tombstone_wire_test.go` encodes PRE-tag witness shapes with `event.DefaultCodec`, decodes them through the fold's `DecodePayloadAuto` path, and re-encodes via `codec.ForEncoding` to prove the tombstone facts' exact-name json tags are byte-neutral on the journal wire — a retention-replay compatibility pin, not an adoption of the codec as an application seam. This is the ONE place the module's vocabulary (`codec.Encoding` + the CBOR constant) is consumed directly, and it sits exactly where it belongs: at the journal's own encoding boundary.

## Ruling

Do not adopt directly. The one named future trigger: if webphone ever persists its own versioned binary blobs outside metaengine (it currently does not — blobs are media files, everything structured is engine-owned), reuse `CBORCodec` rather than inventing an ad-hoc encoding. Until then this module earns its keep invisibly.

**Refresh trigger:** a self-managed versioned blob format; a signing/audit requirement over stored events (then `DeterministicCodec` is the entry point).

**2026-11-22 re-verify** — pin v0.3.3 equals the remote latest (checkout tag v0.3.3). The [Unreleased] section is 4 lines of CI-only movement: golangci-lint raised to v2.15.0 with the fleet errorfamily wrapcheck exemption added to `.golangci.yml`; Added and Fixed both read "Nothing yet" — zero library-code change, zero intent touching webphone's consumed surface. That surface is unchanged: the one test-side import (internal/cqrs/tombstone_wire_test.go:9, `EncodingCBOR` at :93, `ForEncoding` at :139) and the transitive engine seam, which moved upstream to `metaengine/v4/encoded.go:12` when go-cqrs-lite carved v4 (verify-block path updated to match). No verdict movement: no direct prod adoption.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-codec" go.mod                                   # DIRECT pin v0.3.3 (test import)
rg -n "go-codec" --type go internal/ cmd/ | wc -l        # expect 1 (tombstone_wire_test.go only)
rg -n "ForEncoding|EncodingCBOR" internal/cqrs/tombstone_wire_test.go | head -2   # the wire pin
rg -n "go-codec" /home/lars/projects/go-cqrs-lite/metaengine/v4/encoded.go | head -1   # the transitive consumer (v4 path since the go-cqrs-lite carve)
```
