# T15 — Model routing for the post-call review (issue #8)

Decision record, 2026-10-28. Status: DECIDED — the review chain rides the
existing `textagent` seam. No dedicated client, no new config surface.

## The question

The automatic post-call review (issue #8: dispatch → draft → durable
artifact → History render) needs a drafting model behind it. Two routings
were on the table:

1. **The existing `textagent` seam** — the stack's
   `services.telephony.textagent` review API on loopback
   (`textagent.url` + `textagent.secret_file`, both-or-neither,
   fail-closed half-config). The same seam the manual summarize button
   and the note-prefill chain already ride.
2. **A dedicated review client** — a new optional config block (e.g.
   `review.url`/`review.model`), its own transport, auth, verdict
   vocabulary, family pins, and disabled-seam behavior, with a local
   LLM as the imagined default.

## Decision

Route through `textagent`. The T11 dispatch, the detached job, the manual
button, and the note prefill share ONE drafting path (`DraftFromTranscript`
→ the agent's `/draft` op) and ONE idempotency key — the plan's own
constraint ("do NOT re-implement the summarize/draft chain"). The T12
artifact slot is webphone-owned either way: whatever drafts the text, the
durable record of pending/present/failed lives in `readmodels.call_reviews`
and renders in History.

## Why the seam wins

- **Proven end to end**: the summarize chain shipped with verdict
  passthrough, outage 502s, idempotency, and the push fan-out. A dedicated
  client would re-earn every one of those properties in a second vocabulary.
- **Local by default already**: the stack wires the review API on loopback
  (`127.0.0.1:8071`) — the "local default" the plan asked for is the seam's
  deployment shape, not a new provider option. Caller audio never leaves the
  host for the review step; the GDPR posture matches the ASR doctrine.
- **One failure vocabulary**: agent refusals (`*textagent.Error`) and
  transport failures already classify distinctly (`refused`/`failed`
  audit outcomes, `agent.summarize` codes). A second client would fork
  that taxonomy the same week it landed.
- **Zero config = byte-identical holds**: no new block, no new gate, no
  new health leg. The review surfaces gate on the SAME
  `textagent.url`-present condition the summarize button always had
  (the T14 not-configured chip renders from `h.textAgentOn()`).

## What the dedicated client would have cost

A `review.*` config pair with both-or-neither validation, a second
bearer-secret file in the NixOS module, a disabled-seam 404 contract,
family pins + registry rows, and a routing decision inside every call
site (which client drafts?). None of that has a consumer today — the
score dims the artifact reserves (summary/points/score, optional-empty)
stay empty until an agent actually emits them.

## Reopen triggers

1. The owner wants a DIFFERENT model for reviews than for the drafts
   queue (a genuine routing split, not a hypothetical).
2. Kith's review API cannot express a review capability the artifact
   needs (e.g. structured score dims) and the API cannot be extended.
3. A local-only review agent exists that must bypass the drafts queue
   entirely (reviews without drafts).

## Config sketch (the chosen shape — unchanged)

```console
# already shipped; the review chain adds NO keys
WEBPHONE_TEXTAGENT__URL=http://127.0.0.1:8071
WEBPHONE_TEXTAGENT__SECRET_FILE=/run/keys/textagent-secret
```

Off = no review routes, no artifact writes, no History render changes
beyond the not-configured chip.
