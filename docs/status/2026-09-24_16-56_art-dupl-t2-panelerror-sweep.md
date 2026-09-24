# Status Report — art-dupl `-t 2` Sweep #2: `panelError` / `errorBanner` / `identityLine`

**Date:** 2026-09-24 16:56 CEST
**Session scope:** Triage the 16 clone groups from the user-pasted `art-dupl --sort total-tokens -t 2 --type-aware --rich-text --explain --html` run (130 detected / 16 shown / 118 suppressed) and eliminate the harmful ones. No other work.
**Tree state at write time:** a concurrent session landed a tailwind-coexistence spike (`spike.templ`, `/assets/tw.css`, a `headExtras` link line) mid-flight — untouched per the concurrent-session rules; its clone group in the re-run is theirs.

---

## Headline

**All 16 shown groups resolved: 3 extractions (one family, 8 call sites) + 13 accepts; every remaining clone re-attributed in the re-run.** The anticipated `wp-error` family (15:58 report e.6/f.15: "pre-judge whether they're accepted-unless-flagged or extraction candidates") got FLAGGED by this run (groups #4/#10/#13) and is now extracted: `views.errorBanner` owns the ONE `.wp-error` node (wire contract — selected by htmx responseHandling), `views.panelError` owns the conditional wrapper (×5 panels), `views.identityLine` owns the from-identity line (×3 composer panels, same family decision, closes f.15). The re-run at the identical flags shows the two cross-file `wp-error` groups dissolved; what remains maps line-by-line to the accepted set, the on-record "helper call sites ARE the dedup" class (now incl. the three `@panelError` call lines), and the concurrent session's spike noise.

**Discipline repairs from the 15:58 confession list, honored this time:**

- **e.1 (helper + micro-test in the same change, hard gate):** the golden-capture test was written and run GREEN against the PRE-refactor code first (`TestPanelsRenderTheSharedErrorAndIdentityNodes` pins the exact node bytes across all five panels, hostile input doubles as the escaping probe); the three new components got byte-pinning micro-tests at birth; and the standing High-priority `panelHead` backfill (15:58 f.1, the twice-confessed miss) closed in the same change set.
- **e.2 (prove templ refactors at the byte level):** old binary built from a detached HEAD worktree, new binary from the working tree, both booted in loopback with identical config, every tab partial fetched post-login and compared: **7/7 byte-identical**, with the identity line exercised live on messages/fax and the error banner exercised live on the 502 send-failure body (the one diff — a per-server random thread ID — normalizes away).
- **e.3 (state the verification scope at verification time):** running full suite, golangci-lint `./internal/web/...`, treefmt on touched files, art-dupl re-run at the user's exact flags, smoke 41+4. Skipping: `nix flake check` (KVM backup VM + island JS untouched by a views-only change; the suite covers both Go sides), stack browser E2E (markup structure changed AGAIN — parked on the release TAIL relock per the 15:58 g.3 precedent, re-raised below), vulnix/buildflow (nothing in the dependency surface moved; the modified `go.mod`/`go.sum` in the tree are the concurrent session's, not mine).

## a) Triage — all 16 groups

| # | Sites | Verdict | Why |
| - | ----- | ------- | --- |
| 1 | settings.templ:21+23 (dt/dd window) | ACCEPT | On prior record ×2 ("divergent value shapes"); the `settingsRow` trigger question stays the owner's (15:58 g.2) — it surfaced again here, third time |
| 2 | layout.templ:135/137 + messages.templ:102 (`wp-nav-badge`) | ACCEPT | Deliberate hand-roll per AGENTS.md templ table; conditions and count sources differ |
| 3 | layout.templ:152/193/198 + settings.templ:40 (T() paragraphs) | ACCEPT | panelHead-internal vs call-site-adjacent paragraph noise; different keys/classes |
| 4 | history.templ:46 + voicemail.templ:39 (error + needsAPI) | EXTRACT (`panelError`) + ACCEPT | needsAPI half on prior record; error half → the new wrapper |
| 5 | vcard.go:160/164 + i18n_test.go:78 (`i++`/`continue`) | ACCEPT | Loop-skip idiom; the two vcard branches handle different escapes |
| 6 | contacts.templ:58 + messages.templ:98/276 | ACCEPT | Coincidental closing-`</span>` display fragments |
| 7 | idempotency.go:44/54 | ACCEPT | On record with its in-code rationale (`idempotency.go:39`) — verified at the site, not re-litigated |
| 8 | fax.templ:40 + messages.templ:233 (`wp-attach`) | ACCEPT | One empty dropzone div per form; island JS wiring differs per form |
| 9 | messages.templ:143/235 (`wp-segcount`) | ACCEPT | On record: a component call is the same length |
| 10 | messages.templ:176/179 (error + selfNotice) | EXTRACT (`panelError`) + ACCEPT | error half → wrapper; selfNotice is a distinct one-off |
| 11 | actions.go:231 + contacts_api.go:129 (`contactSaveFailed(w); return`) | ACCEPT | Helper call sites ARE the dedup (on record) |
| 12 | contacts.templ:45 + history.templ:53 (`wp-empty`) | ACCEPT | Deliberate hand-roll; different conditions/keys |
| 13 | fax.templ:20 + messages.templ:31 (`@panelHead` + error, MEDIUM) | EXTRACT (`panelError`) | The flagship: the conditional error banner flagged cross-file at MEDIUM |
| 14 | phone.templ:43/72 (`data-i18n` spans) | ACCEPT | Island DOM contract, ported verbatim — untouchable |
| 15 | history.templ:68 + voicemail.templ:72 (row heads) | ACCEPT | Different content expressions; an abstraction would take more params than lines |
| 16 | contacts.templ:55/71 (shared vs personal rows) | ACCEPT | Different domain rules: mutable personal (delete affordance) vs read-only shared (tagged); avatar pair already on record |

Plus the unflagged family pre-judgment the 15:58 report asked for: **`wp-identity` ×3 extracted** (`identityLine`) — same optional-prologue-line shape, shared `identity.from` key, extraction strictly shortens all three sites; **fax.templ:76 `wp-error` span deliberately NOT folded in** — different element and context (per-job inline status inside a muted row, not a panel-level banner).

## b) Verification

| Gate | Result |
| ---- | ------ |
| `nix develop -c go test -count=1 ./...` | 13 test-bearing packages ok (server incl. DOM-contract + strict-CSP pins; views incl. golden + 3 new micro-tests + panelHead backfill; arch invariants) |
| `golangci-lint run ./internal/web/...` | 0 issues |
| gofmt + treefmt on touched files | clean, 0 reformats (b.3 data point: treefmt DOES process `.templ` sources here) |
| art-dupl re-run at the user's exact flags | 16 groups held (the 15:58 "rebalanced shapes" effect); wp-error cross-file groups gone; remainder attributed in §a |
| Old-vs-new binary render diff | 7/7 partials byte-identical (modulo per-server nanoid), error banner + identity lines exercised live |
| `python3 scripts/webphone-smoke.py` | 41 + 4 green |

Instrument mechanics for the render diff (reusable): partials only — the shell carries a per-session CSRF token and would false-diff; the live error path comes from webhook-gateway mode with a dead URL (`POST /messages/send` → 502 re-renders the panel with its banner); a dead `phone_api_url` does NOT work as the error instrument because it blocks loopback login verification with 502 before any panel renders, and a tokenless login POST is a pinned 403 (smoke §5), not a usable probe.

## c) Honest notes

- Two dead-end render-diff designs before the working one (403 tokenless login, 502 login-verification-via-dead-phone-API) — cheap, but a five-minute read of the smoke's session section (§5/§6: CSRF-gated login, adopt-then-post) would have skipped both.
- The 15:58 report's `wp-error`/`wp-identity` sites were quoted from memory of the report; I re-counted them by grep before deciding (3 identity sites, not the "×2" the report guessed).
- The prior archived reports (01-12/03-01) were consulted via the 15:58 report's triage summary plus on-site in-code verification, not re-read in full — sufficient here because nothing on record was re-litigated.
- `go.mod`/`go.sum` modifications and the tailwind spike in the tree are the concurrent session's; my suite/lint runs executed WITH their code present and green, which is also the attribution caveat for any future failure.

## d) Still open (owner)

1. Baseline ratification `-t 3` vs `-t 2` (15:58 f.4) — this run again shows `-t 2` yielding ~16 mostly-idiom groups vs the on-record `-t 3` trio.
2. `settingsRow` trigger (15:58 g.2) — surfaced a third time (group #1).
3. Stack browser E2E timing (15:58 g.3) — served markup structure changed again (panel prologues now route through components; ids/classes/strings intact, render-diff byte-proof in hand).

---

_Point-in-time snapshot — goes stale. Feed (d) to `docs-health` HARVEST; annotate, never rewrite, when bringing current later._
