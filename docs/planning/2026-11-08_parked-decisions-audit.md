# Parked-decisions audit — nothing lost, every trigger live (2026-11-08)

**Status:** AUDIT (P3-9; confirm nothing is accidentally lost) · One
paragraph per parked item, each carrying: the parked verdict, the live
trigger, the ONE home, and what this audit verified at HEAD.

## 1. Scheduled sends — NO-GO stands, ratification owed (pack F-side)

**Parked verdict:** cut from T18 (M21.6); the gateway seam has no
deferred-send concept — loopback sends immediately, the webhook bridge is
fire-and-forget to the provider — so a scheduled-send UI would show a state
the transport cannot honor, violating the island-honesty contract. The
snippets seam (M21) already covers the "reply faster" problem it was meant
to solve.

**Live trigger:** the provider/stack grows a deferred-send lane. Nothing
else re-opens it — not a UI idea, not a scheduler library.

**One home:** TODO §Scheduled sends; the design note
`docs/planning/2026-10-02_10-02_T18-m21-m22-seam-design.md` (M21.6 cut).

**Verified today:** the gateway seam is unchanged on both paths (loopback
immediate; webhook fire-and-forget) — the NO-GO premise still holds at the
paperless-parity train. Owner ratification is pack §F-side wording; a
deferred-send lane would arrive as a stack/provider feature with its own
train.

## 2. Session messenger — parked, DM-first roadmap intact

**Parked verdict:** the `session` network axis exists (`domain.Network`,
`05…` 66-hex addresses, `ParseComposedAddress` routes it, and the lane is
compose-born-inbound) but has NO transport — session messenger integration
stays parked pending a real peer (the protocol NO-GO with reopen triggers
lives in TODO §Chat networks).

**Live trigger:** a transport decision — an actual session-messenger peer
the owner wants to talk to. No transport, no work; the address shape and
sender-registry routing are already built so the integration cost starts
at the transport, not the model.

**One home:** TODO §Chat networks follow-ups (session deferred); AGENTS §
Chat-network seam (the `session` axis + compose routing).

**Verified today:** `domain.ParseComposedAddress` still routes `05…`
shapes to the sender registry; the registry still answers Rejection for
the unregistered session sender (never silent) — parked, not broken.

## 3. The ssetest `WithLastEventID` branding ride — trigger-gated, correctly dormant

**Parked verdict:** go-sse v0.7.0 landed (2026-11-05 wave) with the ssetest
helper's `WithLastEventID(sse.EventID)` available; webphone does NOT adopt
it yet because no Last-Event-ID replay surface exists — the `/events`
heartbeat/replay row is NO-GO until drop-class evidence exists.

**Live trigger:** drop-class evidence (a client demonstrably missing events
across a reconnect that Last-Event-ID replay would have covered), or any
new surface that wants replay. Schedule the ride ONLY then; adopting a
test helper for a capability the product does not offer is inventory, not
work.

**One home:** TODO §Fleet-libs closeout residue, item 1's residual line.

**Verified today:** `/events` still rides `Broadcaster.ServeSSE` without a
replay buffer; no `WithLastEventID` consumer exists in the tree.

## 4. Review-artifact dimensions (#8's open slice) — reopen on operator ask

**Parked verdict:** #8 is MOSTLY SHIPPED (automatic post-call review,
persisted `domain.ReviewState` pending/present/failed, History render
pinned incl. the not-configured chip). The deliberately-unbuilt slice is
the richer artifact shape (summary/points/score dimensions) — today the
artifact stores the draft text + confidence only.

**Live trigger:** an operator asks for richer artifacts. The typed
`ReviewState` machine is the anchor — dimensions extend the PRESENT
rendering's payload, and the wholesale-replace slot semantics
(`call_reviews`, 2000/extension REJECT cap) already bound the storage.

**One home:** TODO §Dialer-adoption issues item 3 (the UPDATE paragraph).

**Verified today:** `TestHistoryTranscriptReviewStates` still pins the four
renderings; the review slot still replaces wholesale; no dimension field
exists anywhere in `readmodels` — parked, not half-built.

## The audit's one finding

All four parked items carry live triggers and single homes — **nothing was
lost**. The one wording risk found: item 1's TODO row says "owner
ratification pending (2026-10-02 §g)" while the owner sitting pack also
carries it; the pack row (F-side, via §Scheduled sends) is now the answer
surface and the TODO row should close when the sitting is answered — one
fact, one answer, then one home.
