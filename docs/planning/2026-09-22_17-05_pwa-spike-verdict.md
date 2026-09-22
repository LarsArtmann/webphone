# PWA spike — verdict: NO service worker; manifest parked behind demand

- **Date:** 2026-09-22 17:05 CEST
- **Plan:** SUPERB T23 (PWA evaluation spike — "verdict doc, build only
  on a green verdict")
- **Verdict:** NOT-DO the service worker. A manifest-only "PWA-lite"
  is parked in ROADMAP behind a real mobile-install demand signal.

## What a PWA would buy here

1. Install-ability (homescreen icon, standalone window) — needs only a
   web manifest + icons. No service worker required.
2. Offline / instant cold loads — needs a service worker with a
   precache + runtime caching strategy.

## Why the service worker is a NOT-DO

- **Offline is architecturally impossible for the product's core.**
  Calls need the WSS REGISTER to the PBX; messages/fax need the
  gateway; every tab partial is server-rendered. A SW can only serve
  STALE UI during an outage — for a phone, a dialer that renders but
  cannot connect is WORSE than the honest "Network request failed…"
  toast. The whole failure→feedback philosophy (docs/error-contract.md)
  is built on honest surfaces; caching fights it.
- **Stale-island hazard, invisible to every gate we have.** The island
  modules are served VERBATIM and the consuming stack's E2E greps
  their bytes; the served-asset tests pin exact behavior. A SW that
  outlives a deploy would pair a NEW server with an OLD island (drifted
  `/config.js` contract, drifted DOM, drifted wire shapes) — and the
  E2E drives a FRESH chromium per run, so it would stay green while
  prod breaks. This is the worst bug class we know: remotely hellish,
  locally untestable by the existing gates.
- **SSE + cache interplay.** `/events` must never be cached (swap-safe
  live fragments), `/api/*`/`/phone-api/*` must always hit the network,
  and the CSP-hash-pinned inline script must not be double-wrapped by
  a SW fetch shim. Each is individually solvable (explicit bypass
  lists) — but every rule is a permanent liability carried for value
  item 2, which is already rejected above.
- CSP itself is NOT the blocker: a same-origin SW script and a
  same-origin manifest are both compatible with `script-src 'self'` /
  the current header set.

## Manifest-lite: parked, not rejected

Install-ability without any fetch interception (manifest + icon links,
no SW) carries ZERO staleness risk. Cost today: real PNG icons (the
repo has only `island/favicon.svg`, linked via `layout.templ`) —
maskable 192/512 PNGs must be generated and maintained. No mobile
install demand is on record for this desktop-internal tool, so the
manifest is parked in ROADMAP behind an owner demand signal (T11
console), not built speculatively.

## Consequences

- T23 contributes this verdict + the ROADMAP park. No code.
- If a concrete offline requirement ever appears (e.g. read-only
  message history on the train), the design must START from the
  bypass lists (never cache: `/events`, `/api/*`, `/phone-api/*`,
  `/config.js`, all POSTs) and a hard versioning story for the
  verbatim-served islands — and it must add a stale-detection gate
  before it may ship.
