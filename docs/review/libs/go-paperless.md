# go-paperless — adjacent library audit (closeout T24)

_Produced by orchestrator of the 2026-11-22 full-fleet re-audit's fan-out (runbook step 7); every other line is the audit's own._

**Verdict:** the fax-archiver adapter rides the SDK's typed surface exactly where it pays: the SDK's branded ids (`TagID`/`DocumentTypeID`/`CustomFieldID`, compile-time distinct since v0.5.0) replaced what would have been webphone's second brand layer, and `internal/paperless` stays a thin fire-and-forget adapter over `fax.Archiver` — a dead Paperless never affects a fax.
**Pin:** v0.5.3 (direct, bumped 2026-11-20 — deps train 4, commit 8b4a45e8) · **Latest:** v0.5.3 · **Currency:** CURRENT — v0.5.3 (2026-11-14) is CI/docs-only for the consumed surface (an api-diff release gate + mechanical drift guards, all in-repo); the earlier minors (v0.5.1 `DocumentMeta.Content`/`ListTags`, v0.5.2 the SDK's internal Retry-After MaxDelay cap) landed 2026-11-12; webphone reads none of the changed surface back (the adapter's fire-and-forget `do()`) · **Status:** direct dep
**Audited:** 2026-11-04 · **re-verified:** 2026-11-13, 2026-11-22 · **Prior reports:** the 2026-09-30 paperless seam plan (lazy metadata ids, duplicate refusal = inert success — all landed).
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
rg -n "go-paperless" go.mod                                 # DIRECT pin v0.5.3
```

## Index-cell adjudications (2026-11-27 tighten train)

### From the index verdict cell (go-paperless)

write-only provenance, blob store stays the storage truth; v0.5.3 landed 2026-11-20 (train 4) — CI/docs-only for the consumed surface
