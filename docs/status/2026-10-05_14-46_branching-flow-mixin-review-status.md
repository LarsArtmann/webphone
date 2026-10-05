# Branching-flow mixin report review — all 8 findings REJECTED

**Date:** 2026-10-05 14:46
**Tool:** `branching-flow mixins . --format markdown`
**Verdict:** 8/8 suggestions REJECTED, 0 refactors, no code changes. Zero harmful
duplication unchanged.

## Method

1. Re-ran `branching-flow mixins .` at HEAD — output matched the owner's
   paste exactly (8 rows). Every struct re-read at HEAD, not from the paste.
2. Judged each with the dedup-code bar ("an abstraction would take more
   parameters than the duplicated code has lines" → accept the similarity)
   plus project doctrine: explicit over implicit, config koanf/json tags ARE
   the operator contract, domain invariants must stay legible, non-domain
   names rejected.
3. Precedent consulted: the CV compose analysis
   (docs/status/archived/2026-03-06_13-46_BRANCHING-FLOW-COMPOSE-ANALYSIS.md
   in that repo) reached the same family of conclusion — "most flagged
   'issues' are valid Go idioms" — and its executed plan went the OPPOSITE
   direction of the detector: converting `Base*` embedded mixins (e.g.
   `BaseDomainEvent`) to explicit fields.

## Cross-cutting rejection rationale

- Go embedding is not inheritance. A "mixin" embed hides fields behind an
  indirection hop: constructors write through the mixin, readers chase two
  levels, per-field doc comments detach from the type that owns them.
- For config structs the `koanf`/`json` tags ARE the operator-facing wire
  contract (env nesting `WEBPHONE_...__...` documented per seam). Scattering
  fields into embedded mixins obfuscates exactly the surface the README
  documents.
- The suggested names (`SharedMixin`, `PasskeyMixin`, `ClientMixin`, …) are
  implementation-leaking non-domain names — the Manager/Helper class the
  naming doctrine rejects.
- Detector semantics: it flags structs with field-count/shape similarity to
  a sibling. Architectural ROLE similarity (service with deps, API client,
  boundary DTO, result carrier) is not duplication — Go embedding can only
  carry IDENTICAL types, so in most rows here it could not even absorb the
  "shared" fields.

## Per-finding verdicts

| Finding                       | Verdict | Why                                                                                                                                                                                                                                                                                                                                                                                                       |
| ----------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `fax.Service` (medium)        | REJECT  | 7 explicit deps; the only shape-sibling is `messaging.Service` whose same-named fields have DIFFERENT types (`*store.Faxes` vs `*store.Messages`, `gateway.FaxGateway` vs `gateway.MessageGateway`, different `ChangeFunc` signatures). An embed can absorb at most blobs/clock/identities (3 declarations, zero logic). The shared shape is the service ROLE, not duplication.                           |
| `config.Passkey` (low)        | REJECT  | 5 fields with distinct koanf keys + doc comments; `Enabled()` derives from the exact field set (half-config fails closed). A mixin scatters the operator contract and obscures the derived invariant.                                                                                                                                                                                                     |
| `config.PasskeyUser` (low)    | REJECT  | 2 fields. Indirection costs more than the 2 lines saved.                                                                                                                                                                                                                                                                                                                                                  |
| `config.CRM` (low)            | REJECT  | Closest call: `config.Paperless` shares the URL+Token shape. Still REJECT — two INDEPENDENT integrations (separately toggled, separately validated both-or-neither, per-seam doc comments); embedding a shared struct couples them and hides `crm.url` / `paperless.url` behind an embed hop.                                                                                                             |
| `crm.Client` (low)            | REJECT  | `base`+`client` are type-identical to `pbx.Client`, but they are unexported fields across packages — sharing means an exported package + new import coupling two seams whose timeouts (15s vs 3s), error codes (`pbx.*` vs `crm.*`) and family tests are deliberately separate. The deliberate similarity is already pinned at the `do()`-chokepoint level (AGENTS.md).                                   |
| `domain.InboundMessage` (low) | REJECT  | Strongest reject. Its doc comment states the design: "carries no status: inbound delivery is a fact, not a lifecycle". `Message` carries Status/FailureKind/ProviderRef and even the attachments field differs in type (`[]Attachment` vs `[]AttachmentContent`). A mixin would weld a lifecycle-free boundary type to a lifecycle-carrying aggregate — a false sameness the data-model doctrine forbids. |
| `server.idemStore` (low)      | REJECT  | 3 unexported fields, single package, constructor adjacent. The method-pair shape already has a standing ACCEPT ruling (idempotency clock+lock prologue). Nothing to share.                                                                                                                                                                                                                                |
| `store.SweepResult` (low)     | REJECT  | 4-field result carrier with ONE producer (`Sweep`). No second site exists; "mixin" is meaningless here.                                                                                                                                                                                                                                                                                                   |

## Standing note for future runs of this detector

Treat `branching-flow mixins` rows as candidate detections, not findings:
verify the sibling actually shares IDENTICAL field types and a shared
behavior, and weight the operator-contract cost for config types. The
project's executed direction (CV precedent) is explicit fields over embedded
mixins. A row is only actionable if the embedded struct removes LOGIC
duplication, not field-declaration similarity.
