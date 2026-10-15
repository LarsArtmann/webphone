# T19 — Sessions on the metaengine? Evaluation memo (no cutover)

**Date:** 2026-10-15 · **Verdict:** N4 STANDS — sessions stay on `session/sqlite.go`.
**Re-triggers:** a storage-format touch (schema/migration work in `session/sqlite.go` anyway) or a multi-node need. Neither exists today.

## What the session store IS today

`internal/session` `NewSQLiteStore`: one keyed row per session (token hash →
extension, directory password, timestamps) in `webphone.db` under the hand
migration ladder. The seams that matter are not storage-shaped: `makeSession`
pins the birth invariant, sliding renewal re-issues the cookie past the idle
halfway point (`session_ttl` 7d idle / `session_max_ttl` 30d absolute), and
the TLS-fronted Secure mint + the uniform anti-enumeration 401s are pinned by
the session-behavior suite (`go test ./internal/server/ -run TestServer`,
`auth:` convention). The directory password rides the row because the
`/phone-api` proxy needs it per request.

## What the metaengine would buy

Nothing the store needs. The engine's wins this train were: durable dedup
windows (restart-loss), fold-expressed read models with sorted scans
(lists), converging cap enforcement (bounded growth). Sessions are a
single-key kv lookup with an expiry slide — no list shape, no cap, no fold
ladder, no idempotency replay. The engine would sit on the SAME sqlite file
class, adding one indirection and one more engine collection to plan.

## What it would cost

- The full `auth:` grade: session store/cookie changes owe the
  session-behavior suite; the `makeSession` birth invariant, the sliding
  re-issue specs and the four-writer 401 allowlist test would all re-pin.
- Continuity: live session rows would need another one-time legacy import
  (or every operator upgrade logs every browser out — the 2026-10-01 outage
  class was exactly a silently-empty credential source at login).
- Risk asymmetry: the store is 9 SQL call sites of boring, pinned, correct
  code. The engine integration is proven for READ models; a security seam is
  the worst possible first write-model consumer.

## Decision

Keep sessions SQL. Revisit ONLY on the named re-triggers above — and even
then, weigh whether the change is really about storage or about the
multi-node coordination problem (which needs more than an engine swap:
session revocation and CSRF rotation across nodes).
