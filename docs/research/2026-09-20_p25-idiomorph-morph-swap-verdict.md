# P25 verdict — idiomorph morph swaps for SSE/HTMX partials

**Date:** 2026-09-20 · **Branch:** [`experiment/idiomorph`](https://github.com/LarsArtmann/webphone/tree/experiment/idiomorph)
(head `8aaf1c7`) · **Verdict: PROMISING — wiring is trivial, local gates all
green; merge is gated on the consuming stack's browser E2E** (deliberately
NOT merged to main; the island DOM + bundle contract belongs to the stack
E2E first).

## What was trialed

Every live-update surface switched from innerHTML replacement to
idiomorph morphing (v0.7.4, bundled inside cqrs-htmx's
`extensions/idiomorph-ext.min.js` — no new dependency, one new script
tag):

| Surface                                                   | Mechanism               | Change                               |
| --------------------------------------------------------- | ----------------------- | ------------------------------------ |
| Thread list (`sse-swap="threads"`)                        | SSE push                | `hx-swap="morph:innerHTML"`          |
| Transcript (`#thread-transcript`, `sse-swap="thread"`)    | SSE push                | `hx-swap="morph:innerHTML"`          |
| Fax list (`sse-swap="fax"`)                               | SSE push                | `hx-swap="morph:innerHTML"`          |
| Voicemail panel re-fetch (`hx-get="/partials/voicemail"`) | `sse:voicemail` trigger | swap `innerHTML` → `morph:innerHTML` |
| Nav badge refresh (`refreshNav` in shell.js)              | `htmx.ajax`             | `swap: "morph:innerHTML"`            |

Server side: two route registrations
(`cqrshtmx.HTMXExtensionHandler(cqrshtmx.HTMXExtIdiomorph)` on the open
mux and the root mux mirror) plus the headExtras script tag. The asset
test map pins `/htmx-ext/idiomorph.js` (a green gate must prove it
scanned).

## Why it works mechanically

- The sse extension resolves the swap spec through htmx's standard
  `getSwapSpecification`, so an `hx-swap` attribute on an `sse-swap`
  element is honored — verified in the v2.2.4 extension source served
  by cqrs-htmx v4.11.0, not assumed.
- `morph:innerHTML` leaves the container element itself untouched, so
  `#thread-transcript`'s `data-page`/`data-thread` attributes and
  shell.js's document-level listeners survive every push — the
  live-transcript contract (paging cancel, read-marking, nav refresh)
  is unaffected by construction.
- The island is untouched: all changes are server markup/attributes.
  The E2E's greppable strings and `#reg-status`/call-card contracts
  cannot regress from this diff.

## What morphing buys

innerHTML replacement destroys and recreates every child node on each
push; morphing patches in place. The user-visible wins:

- **Drafts and focus survive** a live push into the thread list or
  transcript (the exact failure class the swap-safe-fragment rule
  works around today).
- **No flicker** on nav badge refreshes (diff patch, not re-render).
- Scroll anchors inside swapped regions stop jumping.

## Evidence from this session

- `GOEXPERIMENT=jsonv2 go test -count=1 ./...` — EXIT=0.
- `python3 scripts/webphone-smoke.py` — 28/28 on a real binary.
- `nix flake check` — EXIT=0 (treefmt, island-lint, module check,
  package build + sandbox tests).
- Live-server probe: `/htmx-ext/idiomorph.js` serves 200 with
  `Idiomorph` content; the shell carries the script tag; generated
  `_templ.go` output carries all four markup attributes.

## What is still unproven (the merge gate)

The stack's browser E2E
(`nix build -L .#telephony-browser` in nix-international-telephony)
exercises accept/reject, transfer, DTMF and the messages surface in a
real chromium. The stack pins webphone **main**, so proving the branch
requires either a temporary stack input override to the branch or
merging first and reverting on any red. Until that run happens, the
verdict is "all local evidence green, one browser-level gate open".

## Recommendation

Merge after one green stack browser-E2E run against the branch (the
runbook's step 7 command with a temporary `webphone` input override).
If red: revert is one commit — the whole experiment is two server
registrations, one script tag, four attributes, and one shell.js
option. No island code changed, so nothing else needs unwinding.
