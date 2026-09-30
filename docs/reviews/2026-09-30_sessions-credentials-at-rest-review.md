# Sessions credentials-at-rest review (periodic re-review)

**Date:** 2026-09-30 · **Trigger:** the AGENTS sessions invariant ("credentials
at rest accepted, swept on read/Create") owes a periodic re-review; T17 of the
2026-09-30 pareto plan. **Scope:** what is stored, why plaintext is load-bearing,
what protects it, what changed since the 2026-09-20 spike verdict, verdict.

## What is at rest

`webphone.db` (SQLite, modernc) table `sessions`
(`internal/session/sqlite.go:25`):

| Column     | Content                                    | Notes                                                                                   |
| ---------- | ------------------------------------------ | --------------------------------------------------------------------------------------- |
| token      | opaque session token (PK, plaintext)       | minted by `mintToken` (crypto-rand)                                                     |
| extension  | PBX extension                              | owner scoping key for every query                                                       |
| password   | the PBX **directory password** (plaintext) | feeds `Session.Credentials()` → the `/phone-api` proxy and the island's SIP re-register |
| created_at | unix millis                                |                                                                                         |
| expires_at | unix millis                                | idle 7d / absolute 30d (`normalized()`)                                                 |

## Why plaintext is load-bearing (not an oversight)

- `Session.Credentials()` (`internal/session/service.go:48`) hands the SAME
  extension + directory password the login proved to the server-side
  `/phone-api` proxy. A hash cannot proxy.
- `GET /api/session` returns SIP credentials to the island on resume
  (requireSession-gated, `no-store`) so REGISTER re-runs without the user
  retyping. A hash cannot register.
- The alternative (server-side credential vault / per-session tokens minted
  against the PBX) needs a PBX API that FreeSWITCH's directory does not offer
  here; the spike verdict (`docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md`)
  already weighed this and accepted the trade-off.

## Protections (current state, verified 2026-09-30)

- **Expiry + sweeps:** absolute 30d cap; expired rows deleted in bulk on every
  Create (`sqlite.go:67`), lazily on read (`sqlite.go:103`), logout deletes
  (`sqlite.go:111`). A stolen DB file's credential rows self-age-out — the
  DIRECTORY password itself does not expire with them, but the file stops
  being a live session forge kit after expiry.
- **File permissions (new since the spike, v2.8.0):** both systemd units set
  `UMask=0077` + `StateDirectoryMode=0750`; the private-comms hardening
  review pinned that message threads/faxes/voicemail blobs — and with them
  `webphone.db` — are no longer world-readable on disk. Existing files keep
  old modes (owner chmod/re-backup question sits in the owner-calls briefing
  row 28).
- **Token strength:** crypto-random mint; no sequential IDs.
- **Transport:** the consuming stack fronts everything with TLS (WSS for SIP).

## Residual risk (accepted)

Read access to `webphone.db` on the host = disclosure of directory passwords =
PBX account compromise for those extensions until the PBX password is rotated.
Single-tenant household deployment; host compromise already implies
call-control compromise via the PBX itself. Rotating one extension's directory
password invalidates every stored row for it (fail-closed at next verify).

## Changes since the 2026-09-20 spike verdict

1. Sliding renewal landed (`Renew`, idle-halfway re-auth) — changes expiry
   dynamics, not at-rest content.
2. v2.8.0 UMask/StateDirectoryMode hardening (above) closes the
   world-readable gap the spike noted.
3. Nothing else touched the schema or the credential flow.

## Verdict

**Posture stands; no hardening required now.** The 2026-09-20 acceptance is
still sound, the v2.8.0 permission hardening improved the at-rest position,
and the only open item (existing prod files keeping pre-hardening modes) is
already an owner-calls row (briefing row 28), not a code debt. Next re-review:
on the next session-storage change, or 2026-12-30 whichever comes first.
