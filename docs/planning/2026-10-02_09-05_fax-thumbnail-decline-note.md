# Fax first-page thumbnail (C4) — evaluation note, 2026-10-02

Owner-facing decision record for M14's thumbnail item. Status: DECLINED
for now, with rationale and a re-open trigger.

## What C4 asked

A first-page thumbnail of each fax PDF rendered inline in the fax list.

## Options measured

| Option                                                            | Cost                                                                                                       | Verdict                                                                                                                                                                                                     |
| ----------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| pdf.js client-side render                                         | vendor pdf.js (~1.2 MB js + worker) served same-origin; canvas rendering per row; CSP-compatible but heavy | REJECTED — the lean-serving posture is a recorded product stance (the setup-shell adoption was NO-GO'd at +10 MB binary; +1.2 MB of always-loaded JS for a list ornament is the same trade in smaller coins |
| Server-side rasterize (poppler `pdftoppm` in the runtime closure) | new runtime dependency in the NixOS module + binary closure; PDF → PNG per fax (cache? where?)             | REJECTED — deployment-shape change (module + closure growth) for a cosmetic; also needs a thumbnail cache to avoid re-rasterizing on every panel render                                                     |
| Pure-Go rasterizer                                                | no sane pure-Go PDF rasterizer exists (pdfium is cgo)                                                      | NOT AVAILABLE                                                                                                                                                                                               |
| Styled placeholder card (doc glyph + page count)                  | trivial                                                                                                    | REJECTED — cosmetics pretending to be a preview is dishonest UI                                                                                                                                             |

## What shipped instead (M14)

- C5 status timeline (queued → sending → delivered/failed steps) —
  `FaxSteps` in fax.templ.
- C6 resend — `POST /fax/{id}/resend` (failed outbound rows only;
  `fax.Service.Resend` reuses the spooled PDF, original row kept as
  evidence).

## Re-open trigger

If the consuming stack ever gains a thumbnail/preview service (or a
legit need arrives: e.g. operator triage of inbound faxes at a glance),
the honest home is server-side rendering behind a session-gated
thumbnail handler with a blob-store cache — not client pdf.js.
