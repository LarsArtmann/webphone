# Attended-transfer spike (M19) — NO-GO on this host

**Snapshot:** claims verified 2026-10-27 on webphone `main` and the
stack checkout `~/projects/nix-international-telephony`. Verdict up
front: **NO-GO for implementing consult-then-transfer now** — the stack
side cannot be verified here, and an unverifiable bridge must not get
island UI. Everything the island needs is already shipped; the missing
piece is a FreeSWITCH lab that proves `Replaces`.

## What was checked (F111)

- **The island path is real and shipped.** `internal/web/assets/island/app/calls.js`
  documents the split: "FreeSWITCH executes the actual transfer
  server-side on the REFER (the REFERing party's partner leg is
  re-routed through the dialplan on blind; the two existing calls are
  bridged on attended via Replaces) — the browser only sends REFER and
  waits for the NOTIFY sipfrag verdict." Blind transfer works in the
  stack E2E today; attended rides the same transport with a Replaces
  body.
- **The stack has NO Replaces-aware dialplan handling.**
  `modules/freeswitch.nix` mentions neither `Replaces` as a dialplan
  concept nor an attended-bridge extension; the only case-insensitive
  hits are overlay wording. Whether the mirrored module set's vanilla
  templates would bridge an attended REFER between two live legs is
  UNVERIFIED — that is exactly the gap the spike was to close.
- **This host has no FreeSWITCH lab.** No `freeswitch`/`fs_cli`
  binaries, no FS service, and the stack's VM tests carry no
  attended-REFER scenario. F112's loopback smoke is therefore NOT RUN —
  recorded as not-run, never as passing.

## Why NO-GO (and why that is the honest verdict)

Attended transfer's failure mode is silent: a REFER the dialplan
cannot bridge 200-OKs and then does nothing (the DTMF lesson — the
colon form was 200-OK'd and dropped). Shipping a "Bridge to held"
button against an unverified bridge would put a lie in the dock. The
island's honesty contract forbids exactly that.

## The GO day (re-open trigger)

1. The stack grows an FS lab scenario: two registered loopback legs, an
   attended REFER with Replaces, asserting the bridged audio + both
   BYEs (a VM test in `tests/`, next to the telephony-eval check).
2. Only then: the island consult-flow sketch below (a small delta —
   hold + the existing transfer pending path compose; no new state
   machine).

## The consult-flow sketch (F113, carried forward verbatim)

Two-call gesture in the dock: (1) hold the live call (existing hold
machine — the held blip already shows); (2) dial the consult target —
the second call lands in the dock as its own entry; (3) on the consult
leg's Established, the transfer row of the CONSULT call offers "Bridge
to held" — sending REFER with Replaces toward the held dialog and
parsing the NOTIFY sipfrag exactly like the shipped blind path; (4) the
NOTIFY verdict tears down both local legs on success (the PBX bridged
them server-side). No new island state machine: hold + the existing
transfer pending path compose; the only new UI is the "Bridge to held"
button scoped to the case where exactly one held + one established
consult call exist.
