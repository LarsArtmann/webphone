## Problem

`-t 1` on a `.templ`-heavy consumer floods the new-clones gate with markup shape-coincidences: parallel-by-design UI sections and 3-5 line span/conditional shapes collide on structure while every token that carries MEANING (ids, translations, class hooks, conditionals over different fields) differs. Every one of these gets adjudicated ACCEPT; none survives the extraction bar. The baseline then grows as a log of blessed noise, and each UI train re-reds the gate.

## Evidence

One consumer's adjudication record, all with the pinned v0.7.2 `-t 1` (this repo's CI gate):

- 2026-10-23 (chat-networks train): 9 new groups, 5 templ span pairs judged coincidental.
- 2026-10-27 (CRM train): label-for pair; the sweep-log ruling added: *"if a fifth markup-shape group lands, FILE UPSTREAM (art-dupl short-span matching in `.templ`), do not keep blessing them."*
- 2026-10-29 (CRM completion train): 4 markup-shape groups accepted in one sweep.
- 2026-11-05 (admin-surface train): **10 templ-shape groups in a single landing** — settings-head twins (`admin.templ:86` vs `:110`), chip/muted span pairs, a `formatStamp` tail, `<a` blocks, `if x > 0` conditionals over different fields. Baseline grew 86 → 98 in one train; every group ACCEPT.

Sweep log with per-group rulings: `docs/dedup-registry.md` (LarsArtmann/webphone). Baseline: 86 → 98 groups in one day, ~70% of the growth is templ markup shape.

Compounding factor: `//art-dupl:accept` in `.templ` files is currently unusable (#9 — directives force rendered-output markup), so the baseline is the only recourse even where the intent is per-site.

## What would change the verdicts

Any one of these turns the flood into signal:

1. **Templ-aware suppression that doesn't render** (unblocks #9) — per-site intent becomes expressible and the baseline stops accumulating.
2. **A markup-shape tolerance for `.templ` fragments**: weight the match on element structure but require divergence in id/class/text-content to BREAK the clone (parallel sections with different ids are not clones of each other in any sense that survives review).
3. **A `.templ` similarity floor** above `-t 1`'s effective bar for template fragments (e.g. require N shared non-attribute tokens, not just shape).

Happy to supply the full 14-group adjudication set or repo access to the consumer as a fixture.
💘 Generated with Crush

Assisted-by: Crush:glm-5.3-flash
