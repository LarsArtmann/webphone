# Status Report — 2026-09-30 13:17 CEST — `-t 1` Dedup Sweep (Zero Harmful Duplication)

**Session scope:** one task — owner-requested `-t 1` art-dupl dedup sweep from a pasted
report, judged against `docs/dedup-registry.md` (the ONE acceptance home). No other work.

**One-line verdict:** the codebase is CLEAN at `-t 1` (the deepest threshold swept) —
42/42 shown clone groups attributed, 0 new harmful, 0 extractions needed, 3 new standing
rulings recorded.

## Evidence

1. Re-ran `art-dupl --sort total-tokens -t 1 --suggest-generics --timing --rich-text`
   at HEAD (dropped `--type-aware` per the tool's own precedence note): **identical to
   the paste** — 560 detected, 42 shown (395 non-actionable, 123 filtered suppressed).
2. Every occurrence site of every non-ruled group re-read at HEAD, not one site per
   group: all five `data-dial` buttons, both `wp-dir-chip` chips (+ the group's third,
   avatar site), both delivery-verdict guards (+ the fax twin at webhooks.go), the
   delete-affordance pair, all seven plain submit buttons, both composer textareas.

## New standing rulings (all ACCEPT)

- **`data-dial` call buttons ×5** — the attribute IS the shell.js delegated-handler
  contract; sites differ in dial source (4 expressions, 2 guarded by non-empty checks),
  class (`wp-mini` ×4 vs the thread-header `wp-primary`), and sibling actions
  (delete/sms). A component takes 3 params for a 1-line body.
- **`wp-dir-chip` pair** — each chip pairs its own domain predicate
  (`job.Direction == domain.FaxInbound` vs `cdr.Context == "public"`) with its own
  glyph/label helpers; `dirChip(...)` would take 3 params for a 1-line body.
- **Delivery-verdict guard pair** (`case StatusDelivered, StatusFailed:` with empty
  default arm) — two validation LAYERS, not one logic: `hookMessageStatus` 400s
  untrusted wire input; `DeliveryReceipt` rejects the typed domain value (errorfamily
  Rejection, code `messaging.verdict`) for non-HTTP callers. Defense-in-depth by design.
  The `FaxTransmitted, FaxFailed` twin is the same idiom on a different enum.
- **`wp-danger` delete-affordance pair** — same widget idiom, different endpoints
  (contact id vs voicemail uuid), confirm keys and labels.

## Widened rulings

- The composer/submit row now covers the full messages composer widget set ×2
  (textarea, segcount, send) plus the cross-tab plain submit-button family
  (contacts save/import, history filter, messages loadOlder/retry).

## Attributed to existing rulings (no change)

settings dt/dd (owner call), wp-nav-badge spans (incl. the SignedIn site),
T()-windows (welcome li ×3, h2 trio, 1-token T()/format fragments ×15), closing-tag
fragments (muted spans ×3, export/needsAPI.post), needsAPI pair, `i++` ×3,
idempotency clock prologue, `data-i18n` island (spans, summary pair, log-wrap
details, accept/reject/vm-refresh), panelHead + contactSaveFailed call sites,
row-head strong pair, avatars pair, EmptyState conditionals, segcount pair,
head trivia (scripts/meta/stylesheets — distinct assets and loading semantics),
single-widget shapes (history dir options ×3, en/de options, hidden inputs ×3,
Body/FailureDetail conditionals), coincidental single-token pairs (`default:` ×2,
`wp-tab-error`/`calls` divs), and `wp-error`/`wp-notice` (errorBanner is the
one-home; the notice is a different role).

~~**Done:** zero harmful duplication at `-t 1`. Registry rows + sweep log updated.~~ done (2026-09-30: registry rows + sweep log updated; re-confirmed at the 2026-10-01 UI/UX train)
