# Dependency-train notes — metaengine v4.20.0 & catalog v4.10.0 (2026-10-24)

Assessment notes for the two upstream trains behind webphone's catalog
work (completion-plan N10/N12). Neither bump is started; these notes
price them so the next session decides, not rediscovers.

## metaengine v4.19.0 → v4.20.0 — a REAL train (418 commits)

The 12:44 audit called this a "routine bump"; that was wrong (erratum
recorded in the catalog deep-dive's addenda). What the delta carries:

- **Codec contract**: ADR for "encoded ingress decodes with the
  producer's codec" + the documented codec-contract cost measurement +
  no-decoder projection path proofs and write-path guards. webphone's
  readmodels seam rides `metaengine.Query` declarations + DedupStore +
  capabilities ASSERTED at Open — a codec-contract change can move what
  those assertions see.
- **typedfixture + C044 branded-ID fixtures** (dep budget bumped to 2):
  fixture tooling webphone's tests do not consume directly, but the
  branded-ID fixture shapes hint at record-level changes worth reading
  before any bump.
- Release commit strips replace directives (batch release) — the bump
  is a plain version bump, no module-path changes visible.

**Bump-train shape (when GO)**: read the codec ADR end-to-end → run
`go test ./internal/readmodels/` against v4.20.0 in a scratch branch →
vendorHash ritual (`nix build .#webphone.goModules --rebuild`) →
full suite + `nix flake check` → buildflow → push with the daemon
watched. Effort: M (the readmodels capability assertions are the risk
surface, not the SQL).

## catalog v4.9.0 → v4.10.0 — also a real train (213 commits)

webphone's apidocs/wirecontracts registries were built against v4.9.0
(the version the concurrent session pinned). The delta to v4.10.0
rides the same batch release and carries:

- **catalog/architecture**: hide-sections, check report, frontmatter
  export specs (the architecture-reference features — ADR "export
  selections key on kind and id, not custom properties").
- The same metaengine codec work (shared graph).

**Assessment**: nothing in the delta names response headers, 204
schema rules, multipart request bodies, the kafka default server, or
OpenAPI 3.1.0 — the seven upstream candidates from the completion plan
are NOT obviously fixed by the bump (verify per-candidate at filing
time against v4.10.0 before claiming otherwise). Bump value today is
low (no feature webphone needs); risk is low (same API shape). Verdict
candidate for the owner: RIDE the next natural train instead of a
dedicated bump, revisit if upstream filings land fixes on v4.11+.

## Upstream candidates (carried until filed)

Response-header support; 204 exempt from the 2xx-body-schema rule;
`WithParam` type support; no fabricated default kafka server; multipart
request bodies; operationId Pascal leaves dots; OpenAPI 3.1.0 emission.
Each needs a source-level repro against the PINNED version plus a dupe
search before filing (verify-before-filing + github-voice).
