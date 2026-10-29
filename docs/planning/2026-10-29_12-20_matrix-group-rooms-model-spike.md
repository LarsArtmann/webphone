# Matrix group rooms — thread-model spike (arms an owner decision)

Status: SPIKE — analysis and a recommendation; NO code, no fence lifted.
Trigger: TODO_LIST "Lane v2 fences" row ("matrix group rooms — groups need a
thread-model change, not a fence lift", re-confirmed by the 2026-10-23
chat-networks plan's "group rooms follow" and the 10-29 §f item 11).
Signal groups carry the same model question and inherit whatever this
decides; the Sender half of the design is deliberately network-neutral.

## The domain question

What IS a conversation on a multi-party network? Today the product answers
"one extension ↔ one remote address": `Thread` keys on
`(owner, network, remote)` and `domain.Address` is a PEER id — matrix accepts
only `@localpart:server` user ids (`ParseAddress`, the docstring names the
doctrine: "a conversation stays keyed on the person, like every other
network"). A group room breaks both halves:

1. **Identity**: the room (`!room:server`) is the stable wire identifier, not
   any one peer; membership changes over the conversation's life.
2. **Authorship**: `Message` has `Direction` (in/out) and NO sender — in a DM
   the peer is implied. In a group, every inbound message has a distinct
   author, and "who said it" is display-critical.

The current fence is honest: the receiver maps rooms to peers via the
`m.direct` account-data map and skips non-DM rooms with a WARN
(`receiver.go`, "v1: DMs only"). Encrypted rooms stay skipped regardless.

## Options

### A1 — Peer-keyed DMs, room-keyed groups (RECOMMENDED)

Extend `Address` so matrix accepts BOTH peer and room shapes (`@…` and
`!…`); the thread key stays `(owner, network, remote)` byte-for-byte — for a
group, `remote` is the ROOM. One rule covers both regimes: _the thread's
remote is the conversation's stable wire identifier — peer id for 1:1, room
id for multi-party_ (which is exactly how the networks themselves model it).

- DM threads keep the 10-23 property: keyed on the person, so a NEW DM room
  with the same peer continues the SAME thread.
- Group threads are born-inbound (you cannot compose into a room id by
  accident — same doctrine as signal DMs; `ParseComposedAddress` stays
  peer-only, no change).
- The dial affordance, CRM enrichment, and networkLabel all branch on
  `Network()` / dialable shape — a `!…` remote renders with a raw-id fallback
  until room-name tracking lands; nothing dials.

New surface (the real cost):

1. `Message.Sender` — one nullable column + field ("" = DM-implied, so every
   existing row and both DM lanes are byte-compatible). Network-neutral:
   signal groups reuse it verbatim later.
2. Room display names — the receiver must track `m.room.name` /
   `m.room.canonical_alias` state events per room (a small cache beside the
   existing `m.direct` map) or group rows show the raw room id.
3. Outbound-to-room — the sender registry's matrix send path resolves a DM
   room via the inverse `m.direct` map today; a group thread sends straight
   to its room id (simpler than the DM case).
4. Receipts — `m.read` in a group is per-member; v1 can ignore group
   receipts (keep `RemoteRead` DM-only) without lying, since group delivery
   is not surfaced today.

Migration surface: NONE on threads (key shape just gains `!…` values);
messages table grows `sender` (nullable, versioned migration, v6); the
receiver's skip-fence narrows to encrypted rooms.

### A2 — Everything room-keyed (purer, rejected)

Key ALL matrix threads — DMs included — on the room id: one identity regime,
no dual shape. Rejected because it fragments DM history across rooms (a
peer's new DM room becomes a new thread, losing the continuity the 10-23
doctrine bought), and it migrates existing DM rows for zero user-visible
gain. Purity is not worth the continuity break.

### B — First-class multi-party Thread (participants list, YAGNI)

`Thread` grows `Participants []Address` + a kind axis. Nothing in v1 needs
it: display needs room name + per-message sender, not membership state;
maintaining participants from join/leave deltas is a stateful projection
(sync machinery, drift risk) that serves no current surface. Revisit only if
a feature needs member enumeration (e.g. "@here", member pickers).

### C — Per-peer shadow threads (ANTI-PATTERN, never)

Fanning a group room into per-peer DM threads lies three ways: replies must
go to the room (shadow thread implies peer), unread multiplies per member,
receipts misattribute. Recorded so it is never re-proposed.

## Recommendation

A1. It is the smallest model that tells the truth: two address shapes on one
network, each the stable wire identifier of its conversation class, plus one
nullable sender column that signal groups will reuse. Estimated effort M
(receiver room-name cache + sender plumbing + UI sender rendering + tests).

## What the owner decision needs to settle

1. Accept two thread-identity regimes inside `NetworkMatrix` (peer-id DMs,
   room-id groups) — or rule A2/A-B-C differently.
2. Accept the nullable `Message.Sender` column (schema v6).
3. Group receipts: v1 ignores them (recommended) or defers the whole lane.
4. E2EE groups stay fenced regardless (separate lane, olm/megolm research —
   TODO §f item 12; plaintext-only groups first).

## What this spike does NOT change

No fence lift, no config, no wire contract delta (`/hooks` + asyncapi
untouched — groups arrive via the receiver lane, not webhooks). The
skip-fence, the E2EE WARN, and signal's DM fence all stand until the owner
rules and an execution train lands.
