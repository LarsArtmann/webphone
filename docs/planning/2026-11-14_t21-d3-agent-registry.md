# T21 D3 — agent registry / scoped machine tokens: design memo (2026-11-14)

**Status:** DESIGN ONLY — the registry lives in the STACK; this memo is the CONTRACT webphone would hand the stack's review API. The webphone-side enforcement half is already V1-DESIGNED in the TODO_LIST D3 row (`auth.agents` config, `RequireAgent(scope)` constant-time middleware, loopback-only, `actor=agent:<name>` through the existing agentAudit shapes) and is NOT re-decided here. Nothing is wired; security review gates any build.

## 1. Today

ONE machine actor: the textagent bearer secret read from a file at boot (`textagent.New`), one Bearer header on every op, loopback only. Every future agent (summarizer, scheduler, integrations) would share it: no name, no scope, no revocation — one leaked secret is full send authority.

## 2. Token shape

- Opaque bearer minted per agent by the stack's registry: `wat_v1_<agent-id>_<random>`. Opaque over JWT for v1: nothing to decode, no clock skew, revocation is a registry lookup rather than a signature check. The token never lands in rendered config — a per-agent secret FILE (the `LoadCredential` posture; the same shape as today's textagent secret).
- Identity rides separately: the registry stamps `X-Agent-Name: scheduler` (or the review API resolves the name from the token server-side). Webphone's audit gains `actor=agent:<name>` through the EXISTING `agentAudit` journal line + `AgentAuditAppended` row — no new audit schema.

## 3. Scopes

| scope        | grants                      | review-API routes                 | webphone vocab |
| ------------ | --------------------------- | --------------------------------- | -------------- |
| draft-only   | create drafts, read queue   | `POST /draft`, `GET /reviews`     | draft+read     |
| send-allowed | draft-only + approve/send   | `+ POST /reviews/{id}/approve`    | send           |
| audit-read   | read the agent-audit stream | `GET /audit` (new)                | —              |

Default least privilege: `draft-only`. The scope is a CEILING — a `send-allowed` token still obeys the stack's mode gates (off/drafts/veto/after-hours/full decide send policy; the token cannot outrank them). `audit-read` is the read-only lane for observer agents (the row's agents-as-users idea).

## 4. Revocation

Registry allowlist (agent id, token hash, scopes, revoked-at); verification consults it under a ≤60s cache, so revocation lands within one TTL, and any downstream 401 from the registry forces an immediate re-check. Rotation = mint a new token, mark the old revoked-at, overlap ≤ TTL. The stack's kill switch stays the blunt instrument — it freezes conversations regardless of token scopes.

## 5. Verification contract (per request, review-API side)

1. Bearer parse → constant-time compare against the registry's token hash (LOOPBACK-ONLY for v1; TLS before any non-loopback exposure — the textagent posture).
2. Resolve agent + scopes → 401 unknown/revoked, 403 out-of-scope.
3. Enforce the §3 scope→route-family table server-side; clients never self-assert scope.
4. Stamp `X-Agent-Name` on anything that becomes an audit fact downstream.

## 6. Open questions for the owner

1. Build now, or defer until a SECOND real agent exists (the D3 row's own deferral ruling — this memo is ready for the security review either way)?
2. Does the stack accept the registry duty, or would webphone host tokens for stack-resident agents too? (Split-brain risk; memo assumes stack-side only.)
3. Per-agent audit prefixes (`agent scheduler:approve`) vs the `actor` attr only (memo assumes attr-only — `agent:` inside the actor value stays greppable).
4. `audit-read` now, or with the first real observer agent?

## 7. Deliberately NOT decided

JWT/mTLS shapes (the stack's front-door choice), per-agent rate buckets, the webphone `auth.agents` enforcement build (the row's V1 DESIGN stands unchanged), agents-as-users human accounts (D4 territory).
