# Verdict — CSRF token rotation on session TTL refresh (SUPERB T14b)

- **Date:** 2026-09-22
- **Question:** when a sliding session renews (activity past the idle
  halfway point), should the CSRF token rotate with it? (ROADMAP raw
  idea "Session persistence"; SUPERB plan P23 deferred it pending
  exactly this verdict.)
- **Kind:** decision record — NOT-DO, with the threat model that
  justifies staying put.

## Threat model, honestly

The CSRF defense is the double-submit pair: HttpOnly session cookie +
`X-CSRF-Token` header (meta tag / `hx-headers` / island `authedFetch`).
An attacker needs BOTH halves to forge a request.

1. **Token-only leak** (token exfiltrated, cookie not): with
   double-submit this leak is INERT — the header token without the
   cookie cannot forge anything. Rotating on slide would bound
   nothing that matters.
2. **Cookie-only leak** (cookie stolen, token not): rotation does not
   help either — the attacker lacks the token regardless; the cookie's
   power is bounded by the 30d absolute cap and idle window.
3. **Both halves stolen** (= the session is fully compromised): the
   attacker's own requests slide the session and receive every rotated
   token in their own responses. Rotation does not lock them out;
   only `session_max_ttl` (30d absolute) and logout-everywhere would.

The one class rotation genuinely helps — a token captured ONCE by an
attacker who can NEVER observe traffic again but somehow keeps making
the victim's browser send the OLD token while the victim keeps using
the new one — requires an attacker position (live request forgery
without the cookie) that double-submit already defeats.

## Adoption cost (why it is not free)

- The slide happens inside ordinary request handling; every response
  would carry the new token (an `X-CSRF-Token` response header), and
  EVERY consumer (htmx per-request headers, island `authedFetch`, the
  meta tag) must adopt it mid-session — an adoption path that today
  exists only at login/logout.
- In-flight requests race the rotation: a POST that started with the
  old token lands after rotation — the middleware must accept a
  grace window (weakening the check) or reject (user-visible 403s).
- Fixation defense — the reason login/logout rotate — does not apply
  mid-session: the session identity does not change on slide.

## Decision

**NOT-DO, recorded.** The pairs that ARE worth rotating rotate already
(login/logout = fixation defense); the absolute cap bounds the real
compromise scenario; the residual vector is empty under double-submit.
Revisit only if the CSRF scheme ever changes shape (e.g. adopting
per-request synchronizer tokens for a multi-tenant deployment).

Ship status of the sibling item: the adoption RETRY ladder (x3 with
backoff before the reload fallback) is implemented in session.js and
now pinned by two island tests (recover-on-retry-2 without reload;
reload only after three consecutive failures).
