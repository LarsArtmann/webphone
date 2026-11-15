# T12 — Settings-schema export spike (design memo)

2026-11-14. Status: RECOMMENDATION (owner ratification pending). Spike only — no production code. Executes the 2026-11-13 routed row ("settings schema GENERATED from the config structs; the openapi.json doctrine applied to settings") — the named trigger the REJECTED CUE ruling deferred.

## What already exists

- `internal/server/apidocs.go` — the precedent: catalog from wire DTOs → openapi exporter → /openapi.json, byte-pinned by TestOpenapiGoldenFixture.
- `internal/config/unknown.go` — `schemaPaths()` ALREADY reflects over `reflect.TypeFor[Config]()` by koanf tags (incl. the dynamic-key map set for identities/passkey users): half a schema walker, feeding only the unknown-key WARN.
- `internal/config/docpin_test.go` — `configLeafPaths` + README drift tests: the tripwire class already exists for prose.
- `cmd/webphone/contracts.go` + `scripts/contracts-fresh.sh` + the nix `contracts-fresh` check — committed, regenerated artifacts (docs/contracts/) with a freshness gate: ready-made home AND enforcement.

## Three paths, judged

(a) **Reflection walk → JSON Schema draft 2020-12**, rendered by `webphone contracts`. The doctrine applied literally: the artifact cannot drift because it is never hand-carried. Two honest traps: durations ride the wire as STRINGS (`"7d"` — naive reflection emits integer nanoseconds; map `time.Duration` → `type: string` explicitly), and gateway.mode / asr.provider enums are Go constants invisible to reflection (small pinned enum table). Descriptions: lift struct comments via go/doc at generation time, so prose stays next to the field. Pointer-bools (fax.send/receive, unset = on): `default: true` + `x-unset-means-on: true`; NOT `oneOf` null — null never rides the config wire (koanf sees absent-or-bool only), and the marker is what tells a future form generator to preserve absence rather than coerce it to false. ~200 lines stdlib, zero new dependencies.

(b) **go:generate struct-tag codegen** (templ/catalog style) writing the schema. Misfit: templ emits Go SOURCE the compiler needs; a schema is DATA — the directive adds a second sync point (annotation + the wrong-CLI-version trap scripts/templ-generate.sh documents) without adding truth. The house regen ritual is subcommands + freshness gates, not directives.

(c) **Hand-written schema + reflection drift test** failing on key/type mismatch (the TestServedPageHoldsTheDomContract class). Cheapest honest tripwire, and it works — but it re-creates the README-table pattern the doctrine retires: schema prose, x-extensions and if/then blocks rot silently between pin points; every change stays a hand edit. Keep this class as the golden byte-pin ON TOP of (a), not instead of it.

## Recommendation: (a), riding the existing contracts ritual

Generate from the structs; extend `webphone contracts` to write the schema beside asyncapi.yaml; wire it into scripts/contracts-fresh.sh + the nix `contracts-fresh` check (drift dies at regen — the gate diffs, CI refuses); pin bytes with an openapi-style golden test. Path (c)'s walker is reused INSIDE the generator, keeping its tripwire value.

## Deliverable home: docs/contracts/config.schema.json — file, not a route

The docs/contracts split rule: browser-facing JSON lives in /openapi.json, machine lanes live here. The settings schema is a third class — deployment tooling + editors — and committed-and-regenerated is exactly that tree's charter, already guarded by the freshness gate. An HTTP route would publish the internal config shape (secret field NAMES included) for no consumer; CLI generation keeps the zero-config boot byte-identical.

## Consumers

The NixOS module (freeform `settings` submodule + typed fronts in package/options.nix): a nix check or VM-test assertion validates the RENDERED config.json against the schema, so eval-time drift between module options and the binary's loader fails a check — also the shared reference for the open "typed settings.textagent.*" and "stack drift audit" rows. Human config.json editors get completion/validation once README links the schema. A future admin UI reads `default` / `x-unset-means-on` instead of re-learning the tri-state fax convention. And it closes the CUE schema-export deferral: machine-checkable contract, no cuelang.org/go cost.

## Follow-up TODO_LIST row

- [ ] **Ship `config.schema.json`** (spike verdict 2026-11-14): extend
      `webphone contracts` to reflect `internal/config` into JSON Schema
      draft 2020-12 at `docs/contracts/config.schema.json` (`time.Duration`
      → string, pinned enums, `x-unset-means-on` for the fax pointer-bools,
      go/doc comment lift), wire it into scripts/contracts-fresh.sh + the
      nix `contracts-fresh` check, pin bytes with a golden test — consumers:
      the NixOS module's settings assertions, config editors, a future admin
      UI (closes the CUE schema-export deferral).
