# Session-Persistence Design Spike (T11) — Verdict: PERSIST

- **Created:** 2026-09-20 17:41 CEST
- **Task:** T11 of `2026-09-20_17-23_SUPERB-testing-error-feedback-pareto-plan.md`
- **Decision gate:** D2 — persist sessions to SQLite vs keep in-memory.
- **Verdict:** **PERSIST**, implemented as T12 in this same train
  (explicitly approved: "GET SHIT DONE. The WHOLE TODO LIST.").
- **Status:** EXECUTED (see [Execution notes](#execution-notes) at the bottom).

---

## 1. What a session IS (domain)

One signed-in extension. Identity comes from the PBX directory: the
island's successful SIP REGISTER proved the credentials, and
`POST /api/session` re-verifies them server-side before minting. The
session carries exactly two payloads:

1. **Owner scoping** — every tab query (threads, faxes, contacts,
   attachments, SSE fragments) filters by the session's extension.
2. **PBX credentials** — extension + directory password, consumed by the
   `/phone-api` proxy (voicemail re-fetch, history fetch). There is NO
   server-side PBX session handle to hold instead: the phone-api is
   stateless basic-auth per request, so the password is the credential.

Lifecycle: minted on login → live until TTL expiry (default 24h,
`SessionTTL`) or logout delete → gone. The cookie (`webphone_session`,
HttpOnly, SameSite=Lax, Max-Age=TTL) and the store row share the same
TTL by construction (`SetCookie(w, r, token, ttl)` is called with the
same config value the store was built with).

## 2. Consumer inventory (every reader of session state)

| Consumer                             | What it reads                       | Restart impact if persisted                                                                             |
| ------------------------------------ | ----------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `session.Store.Attach` / `Require`   | token → Session                     | none — same seam                                                                                        |
| `requireSession` helper (actions.go) | Session extension                   | none                                                                                                    |
| `/phone-api` proxy handlers          | `PBXCredentials()` (ext + password) | none — password rides the row                                                                           |
| ExtensionHubs (SSE)                  | extension + negotiated lang         | hub map rebuilt empty on boot; first push re-negotiates lang from the `wp-lang` cookie — no stale state |
| Login/hook rate limiters             | peer IP                             | not session state — unchanged                                                                           |
| `/healthz`, `/startupz`              | sqlite ping + blob-dir write probe  | sqlite check ALREADY covers the DB sessions would live in — no new probe needed                         |
| CSRF (nosurf)                        | its own cookie                      | independent of sessions — unchanged                                                                     |

## 3. Design (T12 implements this)

**Table** (owns its migration, same `webphone.db` file):

```sql
CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,       -- the raw cookie token (256-bit)
    extension  TEXT NOT NULL,
    password   TEXT NOT NULL,
    created_at INTEGER NOT NULL,        -- unix seconds
    expires_at INTEGER NOT NULL         -- unix seconds; TTL unchanged
);
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expires_at);
```

**Seam:** `Store` becomes an interface (`Create/Get/Delete/Attach/
Require`); the in-memory map moves to `NewMemStore` (tests, loopback
dev), the SQLite-backed `NewSQLiteStore(db, ttl)` is wired in
`cmd/webphone/main.go`. Same TTL semantics: expiry is enforced on Get
(expired row → dead + lazily deleted), sweep on Create (parity with
today's `gcLocked`).

**Restart test:** create → close DB → reopen → same token still live;
expired-at-restart rows are dead on first read.

**Migration for existing installs:** an empty table means every user
re-signs-in once — acceptable and honest (no resurrection of anything).

## 4. Security review (fail-closed preserved)

The one REAL cost of persistence, stated plainly:

- **Credentials at rest.** The sessions table stores directory
  passwords in the SQLite file. Today NOTHING in webphone.db is a
  credential (messages/faxes/contacts are private but not credentials).
  Persistence introduces the first one. Threat model:
  - _DB theft alone_ (backup leak, file copy): with the password in the
    row, token-hashing adds nothing (the row already grants full
    access), so the token is stored raw — no security theater.
  - _Mitigations that hold:_ the dataDir is 0700 under systemd
    hardening; exposure is bounded by the SAME 24h TTL that bounds the
    in-memory copy's lifetime (sweep deletes expired rows, so long-lived
    backups only ever carry already-expired rows unless taken within the
    TTL window); the backup module's destDir is a separate directory
    from dataDir, so snapshots are not auto-mirrored into the service
    tree. Encryption with an auto-generated key next to the DB was
    REJECTED as theater; a key sourced from the environmentFile was
    REJECTED for now as an ops-breaking mandatory secret — revisit if a
    deployment demands it (config knob, not schema change).
- **Fail-closed unchanged.** A missing/expired/foreign row is still 401;
  the 401-writer allowlist test still pins every rejection site. An
  attacker with a stolen cookie gets exactly what they get today — the
  session until TTL — nothing more.
- **Cookie theft window:** unchanged in shape (HttpOnly, Lax, Max-Age),
  unchanged in duration (TTL), now merely unbothered by server restarts.
- **Idle extension reuse:** another person signing in on the same
  extension mints a SECOND independent row (same as today's map — two
  live sessions per extension are already possible). Scoping stays
  per-row, not per-extension-unique. No new cross-talk.
- **TTL sweep on read:** an expired row is deleted on first Get — a
  restart cannot resurrect a session that expired while the server was
  down (ExpiresAt is absolute time, not idle-extended).

## 5. Restart-behavior change surface (what users will notice)

- Surviving: tab sessions, drafts in open morph surfaces, SSE feed
  (reconnects automatically), the SIP registration (was never
  server-side), contacts/history/fax data (already persisted).
- Still lost on restart: in-flight outbound sends mid-POST (unchanged —
  the HTTP request dies with the process), the notifier's in-flight
  events (unchanged).
- The shipped "Tab session ended" toast becomes a vanishingly rare path
  instead of a per-restart certainty. It stays (hard boot loops,
  manual cookie clears).

## 6. Verdict

PERSIST. It deletes the #1 error-feedback failure class at the root
instead of narrating it; the only real cost is bounded credential
at-rest exposure, which the threat model above judges acceptable for
this product (self-hosted PBX front, dataDir already 0700, TTL-bounded,
fail-closed unchanged). D2 = **yes**, executed in T12 with the design
above.

## Execution notes (post-implementation)

- `session.Store` is the interface; `NewMemStore` (legacy map, tests +
  loopback) and `NewSQLiteStore` (prod) both satisfy it.
- `NewSQLiteStore` owns its `CREATE TABLE IF NOT EXISTS` migration and
  sweeps expired rows on Create and on Get (lazy delete of the row it
  just found dead).
- `cmd/webphone/main.go` wires SQLite-backed sessions over the same
  `webphone.db`; in-memory remains the default ONLY where no DB exists
  (tests). Restart round-trip is pinned by
  `TestSQLiteSessionStoreSurvivesRestart`; TTL-by-read and sweep by
  `TestSQLiteSessionTTL...`; cookie Max-Age parity is unchanged.
