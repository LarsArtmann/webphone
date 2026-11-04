# go-codec — library utilization audit

**Verdict:** NO-GO for direct adoption — it is the event-store payload codec layer, and webphone's HTTP/JSON surfaces are not codec territory. Webphone benefits from it transitively inside metaengine/sqliteengine. **Adoption: n/a (indirect-only, by design).**
**Pin:** v0.3.3 (go.mod:75, `// indirect`) · **Latest:** v0.3.3 · **Currency:** current · **Status:** indirect-only — a direct require of every cqrs-htmx module and most go-cqrs-lite modules (metaengine/encoded.go:12 is the engine's import webphone rides)
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-codec` |
| In webphone | indirect only; **zero imports** in webphone code |

## What the library offers

`Codec{Encoding(), Encode, Decode}` protocol over canonical CBOR / JSON / Raw; `DeterministicCodec` (signing-safe marker); strict compact CBOR (unknown-field rejection); zero-alloc buffer encoding; `ObservableCodec` (metrics/hooks); envelope helpers (`WrapEncode`/`UnwrapDecode`), `TranscodeToJSON`, streaming. It is the encoding seam that go-cqrs-lite stores, signing and encryption consume.

## Webphone overlap analysis

| Surface | What webphone does | Codec relevance |
|---|---|---|
| HTTP JSON DTOs (internal/server/api/*, apidocs) | `encoding/json` | none — deterministic CBOR targets store/signing bytes, not wire APIs |
| Gateway multipart forms | `go-playground/form` + honest part headers (byte-pinned golden) | none |
| Engine persistence | metaengine/sqliteengine encodes payloads **through go-codec internally** | transitive benefit already realized: webphone's read-model views and event payloads persist in the engine's canonical encoding without webphone touching the module |
| Matrix/signal/chat receivers | bespoke JSON against upstream protocols | none — wire compatibility owns the format |

**No hand-rolled codec exists to replace** — `rg "go-codec|msgp|CBOR" internal/` shows no parallel implementation.

## Ruling

Do not adopt directly. The one named future trigger: if webphone ever persists its own versioned binary blobs outside metaengine (it currently does not — blobs are media files, everything structured is engine-owned), reuse `CBORCodec` rather than inventing an ad-hoc encoding. Until then this module earns its keep invisibly.

**Refresh trigger:** a self-managed versioned blob format; a signing/audit requirement over stored events (then `DeterministicCodec` is the entry point).

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -n "go-codec" go.mod                                   # indirect pin
rg -rn "go-codec" --type go internal/ cmd/ | wc -l        # expect 0
rg -n "go-codec" /home/lars/projects/go-cqrs-lite/metaengine/encoded.go | head -1   # the transitive consumer
```
