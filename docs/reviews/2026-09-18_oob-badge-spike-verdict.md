# OOB badge push spike — verdict: PARKED

**Date:** 2026-09-18 · **Plan items:** OO1–OO3 (M29/M30), UB1 gated on this verdict.

## Idea

Ride the unread-badge update on the existing `threads` SSE event: the
server appends an out-of-band fragment (library `cqrshtmx.OOBHTML(id,
html, swapStrategy)`) to the swap-safe `threads` payload, so the nav
badge updates live without a new event name, a new endpoint, or island
polling.

## Prototype shape (not merged)

```go
// notifier.go, only when WEBPHONE_SSE_OOB=1:
badge := views.UnreadBadge(views.UnreadBadgeProps{Count: n})
payload := renderedThreads + cqrshtmx.OOBHTML("vm-badge", renderedBadge)
hubs.Publish(ext, sseEventThreads, payload)
```

Server-only: htmx processes `hx-swap-oob` fragments inside SSE event
data on swap; the island needs no change. Invariant §2.3 holds — the
flag gates every byte of the behavior.

## Why parked

1. **The E2E gate cannot run here.** Adoption of OOB changes SSE payload
   bytes; the plan (and AGENTS.md's E2E rule) requires validating that
   against the upstream stack browser E2E before anything defaults on.
   That loop lives in `nix-international-telephony` and needs its full
   stack (browser, PBX); this session stayed in-repo by design.
2. **Weak demand signal.** The badge already refreshes via TTL'd cache
   (5 s) plus drop-on-mutation invalidation; OOB push is polish, not a
   correctness gap.
3. **Risk asymmetry.** A malformed OOB fragment inside a `threads`
   event could corrupt the swap-safe-fragment contract that several
   tests pin — the one payload contract we deliberately keep frozen.

## Adoption criteria (when revisited)

- Flag-gated prototype (`WEBPHONE_SSE_OOB=1`, default off) merged behind
  the stack E2E proving: badge swaps, tab content intact, no double-swap,
  draft state preserved.
- UB1 (badge push from unreadCache invalidation points) follows only on
  a green run.
- Re-check `cqrshtmx.OOBHTML` signature against the then-current tag —
  this verdict was written against v4.9.0.
