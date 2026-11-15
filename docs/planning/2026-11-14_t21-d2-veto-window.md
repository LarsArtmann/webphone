# T21 D2 — veto-window mode (HITL L2): design memo (2026-11-14)

**Status:** DESIGN ONLY — owner-GO + stack-train gated; the send policy lives in the stack's textagent, webphone only surfaces it. Nothing is wired. Backs the TODO_LIST D2 row (veto window / HITL L2).

## 1. Gap and principle

Rungs today: L1 (`drafts`, everything waits for Approve) and L3 (`full`, a confident draft sends immediately). Missing L2: a confident draft holds T seconds; ONE click vetoes; silence sends. The timer and the hold live STACK-side (webphone is a restart-prone single binary; the agent is the sender of record) — webphone renders the countdown, never enforces the deadline.

## 2. Mode and hold state

- Fifth mode `veto` in the stack's `/mode` vocabulary. Webphone adds NO mode logic: the drafts picker renders `Reviews.Modes` verbatim (`draftModeOptions`), `SetMode`/`ResetMode` post the string, the header pill shows it — the moment the stack lists `veto`, every surface lights up.
- The held draft lives in the stack's review queue WITH a deadline (the review queue is the hold store; no new webphone state): `Draft.Status:"held"` so it rides `Reviews.Drafts` as today, plus `send_after` (absolute RFC3339 deadline — a stale poll still renders an honest remaining) and `sent_by:"veto_expired"|"owner"`. Below-confidence drafts stay `pending`: the window captures only drafts that would have auto-sent.

## 3. Veto UX

- PRIMARY: a per-draft chip in the `#wp-drafts` row — "sending in 47s" beside the contact meta, a Veto button beside Approve, NO `hx-confirm` (a veto prevents a send; the failure mode is cheap).
- POLL: the panel's `every 30s` is too coarse under a 60s window — the render tightens hx-trigger to `every 5s` while any hold is live (a rendered attribute, no JS), back to 30s otherwise.
- SECONDARY: the mode pill shows `veto · n held` (the `wp-draft-count` badge substrate exists). Toast-only is REJECTED: toasts are ephemeral, a hold is state.

## 4. Expiry and outcomes

- EXPIRE: the stack sends, status `sent`, `sent_by:"veto_expired"`, gone from the list on the next poll. The deadline is absolute persisted stack state: a webphone restart changes nothing; a stack crash fails safe back to `pending` (stack's call).
- VETO: one click returns the draft to `pending` — "not confident enough to auto-send", not "never send"; it rejoins the normal Approve/Edit/Discard loop. Approve-with-edit sends immediately, as today; the window never restarts.

## 5. Mode matrix (safety gates unchanged — only the send decision moves)

| mode        | confident draft       | veto? |
| ----------- | --------------------- | ----- |
| off         | nothing drafts        | n/a   |
| drafts      | awaits Approve        | n/a   |
| veto        | holds until now+T     | YES   |
| after-hours | time-scoped auto-send | no    |
| full        | sends immediately     | no    |

After-hours SUPERSEDES veto — a countdown nobody is awake to veto is fake oversight; `full` ignores it.

## 6. Audit lines

- Human veto rides the UNCHANGED `textAgentPanelAction` machinery (key `drafts.veto`, audit event `veto`, draft id on the path): journal `agent veto extension=1000 actor=lars@… draft=d-9F2 outcome=ok|failed` + durable row `AgentAuditAppended{Event:"veto", Subject:"d-9F2", Actor:"lars@…", Outcome}`.
- Expiry: observing `held → sent(sent_by:"veto_expired")` in a poll or nudge journals ONCE per draft — `agent expire … outcome=sent`, deduped on draft id; the stack keeps its own send-time audit. Mode flips to/from `veto` ride the existing `agent mode` line — no new shapes.

## 7. Owner open questions

1. T default — memo assumes 60s; the deployment default lives stack-side, surfaced via `send_after`.
2. Veto outcome: back to `pending` (assumed) or straight `discarded`?
3. Does READING the thread veto, or only explicit draft actions? (Assumed explicit-only: a silent false veto swallows messages.)
4. After-hours: supersedes (assumed) or veto applies inside it too?
5. Expiry audit: observed-once journaling (assumed) or a stack-emitted send event over the `wp:drafts-changed` nudge (cleaner, more stack work)?

## 8. What webphone asks of the stack (the contract half)

`veto` in `/mode` + `Reviews.Modes`; `Status:"held"` + `send_after` + `sent_by` on the Draft JSON; `POST /reviews/{id}/veto` → back to `pending`; stack-side deadline persistence + a send-time audit record.
