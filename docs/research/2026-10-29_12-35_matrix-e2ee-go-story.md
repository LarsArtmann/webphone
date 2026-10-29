# Matrix E2EE — the Go story (research, arms an owner decision)

Status: RESEARCH — no code, the encrypted-rooms WARN fence stands.
Trigger: TODO_LIST "Lane v2 fences" row (matrix E2EE "needs olm/megolm —
big"); §f item 12 of the 2026-10-29 whole-list report. Claims verified at
source 2026-10-29; one second-hand claim was corrected in the process (see
"corrections").

## The question

Can the webphone matrix lane decrypt (and later send) end-to-end-encrypted
messages from the single static Go binary — without CGO, without a second
daemon, keeping the hand-rolled 356-line CS-API client
(`internal/matrix/client.go`, zero matrix dependencies in go.mod today)?

## Verified findings

1. **libolm is end-of-life** (deprecated by the Matrix.org Foundation,
   late 2024); the blessed successor is **vodozemac (Rust)**. Any "just link
   libolm" idea is a dead end from the start.
2. **matrix-rust-sdk has NO official Go bindings** — verified at the repo's
   bindings README: Swift/Kotlin/JS/NodeJS via UniFFI only, and UniFFI does
   not target Go. There is no module to import. Using vodozemac from Go
   would mean a hand-rolled C-ABI bridge plus CGO — which poisons the flake's
   aarch64 cross-build posture and the no-cgo single-binary doctrine.
   REJECTED on structure, not preference. (An earlier fetch claimed Go
   bindings exist; the source says otherwise — recorded as a correction.)
3. **The practical Go path is mautrix-go** (`github.com/mautrix/go`,
   `maunium.net/go/mautrix` module path): its `crypto/` package is a full
   E2EE engine over **goolm, a pure-Go libolm reimplementation**
   (wire- and pickle-compatible), actively maintained at v0.32.0 — it is the
   crypto engine behind mautrix bridgev2. Megolm receive support is complete:
   `DecryptMegolmEvent`, `WaitForSession`, room-key requests and withheld
   notices, key import/export, server-side key backup, cross-signing,
   emoji-SAS verification.
4. **A minimal decrypt-capable client is small on paper**: `OlmMachine` +
   a crypto `Store` (~40-method interface; in-memory `MemoryStore` exists,
   a SQL store ships as a separate module) + a `StateStore` (membership);
   `machine.Load` then feed `ProcessSyncResponse` / `HandleToDeviceEvent`
   from the sync loop. Interactive verification is OPTIONAL for basic
   decrypt/encrypt — trust defaults to TOFU and our device shows as
   "unverified" to peers until cross-signing is done.

## Adoption paths for THIS repo

- **P1 — adopt mautrix.Client wholesale.** Replace the hand-rolled client
  with mautrix's sync loop entirely. Largest dep tree, redundant machinery
  for a receive lane, abandons the deliberate hand-roll. Not recommended.
- **P2 — OlmMachine bridge (RECOMMENDED).** Keep our sync loop and receiver;
  construct a minimal `*mautrix.Client` (same homeserver/token) that ONLY
  the machine uses for its own to-device uploads; `NewOlmMachine` takes the
  concrete client type (verified), our loop stays the driver feeding it sync
  data. One module dep; the machine owns the hard 20% (olm sessions, key
  requests, ratchets); we keep our wire client and fences.
- **P3 — goolm primitives only.** Import just `crypto/goolm` and hand-wire
  sessions/ratchets/to-device state. Zero client coupling but re-implements
  the machine's hardest parts with fresh bug classes. Only if the mautrix
  module dep is ruled unacceptable.

Decrypt-first shape (recommended sequencing): receive-only E2EE (decrypt
megolm events, request missed room keys, honor withheld) before ANY send —
`ShareGroupSession` drags in the verification/trust story for peers.

## Follow-on decisions the owner would owe

1. P1/P2/P3 ruling (recommendation: P2).
2. Crypto-store persistence: an own `matrix-crypto.db` (the usermgmt.db
   pattern) — and whether it rides the backup set (user data since the
   readmodels train).
3. Trust posture: TOFU with an "unverified device" presence in peers'
   clients (fine for an operator PBX lane?) vs. full cross-signing later.
4. Sequencing vs. the group-rooms spike (A1): E2EE DMs are independent
   (peer-keyed threads exist); E2EE GROUPS should land after A1 so room-keyed
   threads and room-key sharing arrive in one shape.

## Cost line

M–L for P2 decrypt-only (machine wiring + persistent store + sync feeding +
tests against recorded encrypted fixtures); send-encrypted adds the
verification story on top. The current fence (skip encrypted rooms, WARN)
stays until this is ruled and scheduled.
