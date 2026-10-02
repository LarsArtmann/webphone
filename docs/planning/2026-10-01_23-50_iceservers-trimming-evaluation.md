# `iceServers` trimming evaluation (T11.4, 2026-10-01)

**Status:** evaluation only — NO config change. The plan's own rule
applies: no trimming without live path numbers, and the numbers come
from the owner's live-call ritual (T01) which has not run yet.

## Where the list actually lives

Webphone ships `ice_servers` from its config VERBATIM into
`/config.js` (`window.PBX_CONFIG.iceServers`); the optionally derived
TURN credentials (`turn_rest`) ride the same list. The DEPLOYMENT's
actual list is defined in the consuming stack
(nix-international-telephony), not in this repo — any trimming change
lands THERE, behind this evaluation's decision gate.

## What trimming could mean (and what it must not touch)

1. **Redundant STUN entries** — gathering with ONE STUN completes well
   under the 1000 ms cap (measured during the mic-prewarm train; the
   cap note lives in AGENTS.md). Extra STUN servers add parallel
   gathers, not path diversity, on a single-NAT household deployment.
   This is the one candidate worth evaluating.
2. **TURN entries** — NOT a trimming candidate. Relay is the fallback
   that makes calls work at all when hole-punching fails; removing TURN
   allocation cost is dominated by removing the only recovery path for
   this stack's #1 failure mode (NAT/ICE).
3. **`iceCandidatePoolSize`** — not configured today; out of scope.

## The evidence the decision needs (now shippable)

The ICE panel (T11.3) renders the exact numbers this evaluation gates
on, per call, with zero tooling:

```
setup: gather 812 ms · ice +0.6 s · first media +1.4 s
ice: connected  path: srflx → host
```

The owner ritual (T01) records, across a few real calls (wired and
Bluetooth, from the household network):

- `gather` ms and whether the `gather capped (>1 s, …)` variant ever
  appears (cap hits → the list is SLOW, the opposite of trimmable),
- `path:` selected-candidate types (`srflx → host` = healthy;
  `relay` appearing ROUTINELY → keep everything and investigate why
  direct paths fail),
- `rtt:` on the selected pair.

## Decision gate

| Evidence over several calls                    | Decision                                                                                                             |
| ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `path: srflx → host`, gather < 800 ms, no caps | A single STUN suffices → trim the STACK list to 1 STUN + TURN                                                        |
| Any routine `relay` path                       | Do NOT trim; investigate direct-path failure first                                                                   |
| `gather capped` seen more than once            | Do NOT trim; gathering is already at the budget — the list is too slow, consider fewer servers or a higher cap first |

## Revisit trigger

The panel ships with every call now. Any future call-quality
complaint starts by reading the `setup:` line; this note is the
standing interpretation of those numbers.
