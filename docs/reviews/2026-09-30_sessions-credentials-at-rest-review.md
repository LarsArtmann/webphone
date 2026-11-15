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

> **Trigger note 2026-11-14 (docs-health sweep):** the re-review is OWED —
> a session-storage change landed after this review's snapshot (2026-10-28:
> the `login_email` column via the guarded idempotent ALTER, plus the
> identity-layer session rows minting through the same store). The re-review
> row lives in TODO_LIST (2026-11-14 September-corpus sweep residue);
> 2026-12-30 remains the fallback date.

## Re-review 2026-11-14 (the login_email trigger)

**Trigger honored:** the 2026-10-28 multi-user substrate landed the
`login_email` column and the identity-layer session rows — exactly the
"next session-storage change" the 09-30 verdict named. Every claim below
re-verified against current HEAD (`go test ./internal/session/...` green).

### Per-claim verification

| 09-30 claim | Still true | Evidence (file:line) |
| --- | --- | --- |
| `sessions` table: token PK, extension, password, created_at, expires_at | Yes — plus the new `login_email` column (NOT NULL DEFAULT '') | internal/session/sqlite.go:29-36 |
| Sessions table lives OUTSIDE the store/db.go ladder; package owns its DDL at open | Yes — ladder holds threads/messages/attachments/fax_jobs/contacts/snippets/call_transcripts/app_flags only | internal/store/db.go:163-283; sqlite.go:62-66 |
| `login_email` upgrade is a guarded idempotent ALTER (duplicate-column = already migrated) | Yes | sqlite.go:45, 67-71 |
| Bulk sweep of expired rows on every Create | Yes | sqlite.go:85-89 |
| Lazy delete of an expired row on read | Yes | sqlite.go:129-131 |
| Logout deletes the row | Yes | sqlite.go:137-139 |
| Plaintext password is load-bearing: feeds the `/phone-api` proxy | Yes — method renamed `Credentials()` → `PBXCredentials()` | service.go:77-79; internal/server/panels.go:186, internal/server/actions.go:280,461 |
| `GET /api/session` resumes with SIP credentials (requireSession-gated, no-store) | Yes — now split: derived `sip_credential` (HMAC of token; secret outside the DB) when `sip.derived_credentials` is on, raw `password` only when off | internal/server/auth/session_api.go:73-87; internal/server/auth/session_identity.go:86-95 |
| Sliding renewal, idle 7d / absolute 30d defaults | Yes | sqlite.go:145-162; internal/server/server.go:343; internal/config/config.go:683-684 |
| Token strength: 256-bit crypto-rand, no sequential IDs | Yes | service.go:34-35, 176-182 |
| UMask=0077 + StateDirectoryMode=0750 both units, eval-time pinned | Yes | package/nixos-module.nix:253-254,303-304; nix/module-check.nix:123-130; SECURITY.md:39 |

Identity-layer rows verified: ONE birth path `mintSession` → `Create(extension, password, loginEmail)` (session_identity.go:67-75); passkey persists `mapped.Email` (passkey_api.go:115), OIDC the same (sso_api.go:84), extension logins keep `""` byte-parity (session_api.go:55). NO `agent_mode` at rest — that field is a per-request, best-effort decoration of the identity response (session_identity.go:118-121, 152-167), never a column.

### Residual risk, rewritten against the identity-layer reality

Read access to `webphone.db` now discloses more than directory passwords: each row also carries `login_email`, the plaintext WHO behind a passkey/OIDC session (service.go:62-73) — a multi-user audit trail in miniature. The credential core is unchanged: identity-layer sessions still carry the extension's directory password (file-sourced at mint, passkey_api.go:104; sso_api.go:73), because the `/phone-api` proxy consumes it regardless of brokering — credential brokering only keeps the password off the BROWSER wire, not out of the row. The accepted equation therefore stands: host read access = PBX account compromise per stored extension, plus disclosure of which humans share those lines. Mitigations grown since 09-30: brokering exists as a deploy-time option, the file-permission hardening is eval-time pinned (module-check.nix:123-130), and both password and email age out with their row (sweeps + logout). Single-tenant household scope unchanged; host compromise still implies PBX compromise independently. Rotating one extension's directory password still fails closed at the next directory verify.

**RE-REVIEW COMPLETE 2026-11-14 — residual risk: LOW because the accepted 09-30 posture holds unchanged under the identity layer (plaintext directory passwords remain the deliberate, proxy-load-bearing payload; the only new at-rest content is low-sensitivity login emails that age out with their rows), the protective perimeter grew (credential-brokering option, eval-pinned file-permission hardening), and the original acceptance — host read access already implies call-control compromise in this single-tenant deployment — was not weakened by the change.** Next re-review: next session-storage change, or 2026-12-30, whichever first.
