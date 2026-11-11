# go-paperless — adjacent library audit (closeout T24)

**Verdict:** the fax-archiver adapter rides the SDK's typed surface exactly where it pays: the SDK's branded ids (`TagID`/`DocumentTypeID`/`CustomFieldID`, compile-time distinct since v0.5.0) replaced what would have been webphone's second brand layer, and `internal/paperless` stays a thin fire-and-forget adapter over `fax.Archiver` — a dead Paperless never affects a fax.
**Pin:** v0.5.0 (direct) · **Latest:** v0.5.2 (re-verified against remote tags 2026-11-11) · **Currency:** 2 minors behind, no consumed surface changed — v0.5.1 is `DocumentMeta.Content`/`ListTags`/redirect+config coding (webphone reads nothing back), v0.5.2 caps the SDK's internal Retry-After hints at MaxDelay (the adapter's fire-and-forget `do()` never surfaces them); bump rides the next fleet train, not this report · **Status:** direct dep
**Audited:** 2026-11-04 · **Prior reports:** the 2026-09-30 paperless seam plan (lazy metadata ids, duplicate refusal = inert success — all landed).
**Run all commands from the webphone checkout root** (repo-relative paths).

## Usage (verified)

`internal/paperless/paperless.go`: optional (`paperless.url`+`paperless.token` both-or-neither), nil-safe everywhere; metadata ids lazily ensured per first SUCCESS; inbound faxes only (outbound faxes never reach Paperless — the product's own blob store stays the only storage truth); the SDK rides `pbx.Client`-style timeout-bounded HTTP through the adapter's own `do()` chokepoint shape.

**Unused surface (ruled as-is):** outbound/document-update SDK paths (no product lane), any search/list API (webphone never reads Paperless back — write-only provenance by design).

## Ruling

Keep as-is. **Refresh trigger:** a product lane that reads documents back (today nothing does); the SDK shipping an outbound-fax-relevant feature (it has none — Paperless-ngx is a DMS, not a fax gateway).

## How to verify

```console
rg -n "sdk\\." internal/paperless/paperless.go | wc -l      # expect >= 6 (typed SDK calls)
rg -n "TagID|DocumentTypeID|CustomFieldID" internal/paperless/paperless.go | head -4   # SDK brands, no local duplicates
rg -n "go-paperless" go.mod                                 # DIRECT pin v0.5.0
```
