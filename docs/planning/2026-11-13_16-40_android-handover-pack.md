# Android handover pack — line picking + CallStyle server contracts

**Date:** 2026-11-13 (pareto plan T20/F66) · **Audience:** the `webphone-android`
session (P1 ring-while-closed shell) · **Status:** DRAFT for owner review —
every server-side fact here is verified against the tree at this date; the
app-side spec half is a proposal, not a landed contract.

This pack answers the two questions the android repo's parked rows keep
hitting: (1) WHICH line does the device speak for, and how does that change
at runtime (the "origin picker"); (2) what the ring payload can and cannot
feed an Android CallStyle notification.

## 1. What the app already consumes (verified)

| Fact                                         | Source                                         | Contract home                                                                                  |
| -------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `sipDomain`, `websocketPath` (island config) | `/config.js` (`window.PBX_CONFIG`)             | `internal/server/configjs_test.go`                                                             |
| Session identity, both credential arms       | `/api/session` + `/api/session/switch` answers | `docs/contracts/wire/session-identity-{brokered,password}.json` (byte-pinned, freshness-gated) |
| Push payloads, both transports               | ntfy JSON body / FCM `data`                    | `docs/contracts/wire/push-*.json` (7 fixtures)                                                 |
| Push registration                            | `POST /api/push-devices` (CSRF dance)          | asyncapi.yaml + the fixtures' request shape                                                    |

The build-time half of the digest handshake LANDED 2026-11-10: the android
repo vendors these fixtures (`scripts/sync-wire-fixtures.sh` there) and its
JVM tests decode them through the real parsers. The RUNTIME half (fetch a
digest at origin-config time, render "server older than app" on mismatch)
stays deferred with its named trigger: the first real server/app version-skew
incident. The `/openapi.json` + `/manifest.webmanifest` 304 ETag lane makes
the digest fetch effectively free when it fires.

## 2. Line picking — the server contract

The session identity carries the switchable set:

```json
{
  "extension": "1001",
  "did": "+1555…",
  "display_name": "Jane Operator",
  "numbers": ["+1555…"],
  "extensions": ["1001", "1002"],
  "sip_credential": "…"
}
```

- `extensions` lists EVERY same-email mapped line (`auth.users` mapping);
  it is empty for extension-password logins (single line, no switching).
- `POST /api/session/switch` re-mints the session for another line in that
  list: fresh password-file read + directory verify + CSRF rotation, old row
  deleted AFTER the new mint. Own-line switches are idempotent no-ops;
  cross-email extensions 403; identity layer off 404. The server never
  exposes the mapping to an anonymous caller — the picker can only offer
  lines the ANSWER already named. THERE IS NO server endpoint that lists an
  account's lines pre-login; the login form's extension field IS the picker
  at first sign-in.
- A live SIP call blocks the switch on the ISLAND side; the native shell
  inherits that rule by construction (a CallStyle-active call must not
  re-mint mid-call — the registration credential changes with the session).

**App-side proposal (for the android session to rule on):** Settings-level
"Active line" picker fed by `sessionIdentity.extensions`, switching through
`/api/session/switch` + full re-registration; hidden when the list has < 2
entries (byte-parity with single-line deployments). Push registrations are
per-owner, so after a switch the device MUST re-register push for the new
owner — the register request carries the session, so the natural sequence is
switch → re-register → refresh identity.

## 3. CallStyle — what the ring payload feeds, and what it cannot

The ring hook wire shape (`internal/server/hooks/ring.go`, the asyncapi
document is generated from it — never hand-edit the yaml):

```json
{ "owner": "1001", "from": "+1555…", "call_id": "<PBX Call-ID>", "state": "ring" | "end" }
```

- Routing is per-owner: a device registered for owner X receives only X's
  ring/end facts. `from` is the RAW remote number (the dialplan's), not a
  contact name.
- Delivery is ADVISORY: missed push = the PBX rings on as normal; the app
  must never be the only ringer. `priority 5` (ntfy) marks ring; `end`
  cleans the notification up. `call_id` correlates a ring with its end.
- What CallStyle needs that the payload does NOT carry: a display NAME. The
  server deliberately does not resolve contact names into the ring payload —
  push transports are third-party (ntfy/FCM), and putting caller names into
  those transports is a privacy regression the current posture forbids (the
  CRM data-mesh rule: names render server-side, never leave through push).
  The app's honest options: display the raw number (fixture-faithful), or
  resolve locally against its own synced contacts (app-side cache — an
  android-repo decision, not a server ask).

**App-side spec (proposal):** `Notification.CallStyle` on API 31+ with
`CommunicationStyle`/`CallStyle.forIncomingCall` — needs a full-screen
intent (USE_FULL_SCREEN_INTENT permission; Android 14 requires the user
grant for non-call apps), a 20s-ish post/cancel timeout, and the `end` fact
to cancel. API 26-30 fallback: heads-up + full-screen intent. The parked
android ROADMAP row (P3, "CallStyle shade on the API 31-33 direct-dial
path") is this spec's implementation vehicle; device verification stays that
repo's J23 row.

## 4. Open server-side asks (NONE filed — decisions, not gaps)

1. Ring-payload name enrichment: REJECTED by posture today (PII through
   third-party transports). Reopen only as an owner privacy ruling.
2. Pre-login line listing: REJECTED by posture today (anti-enumeration —
   the mapping is auth-gated). The login-time extension field is the
   picker.
3. A `config.js` key fixture in the sync set: deferred with its trigger
   (first key drift or a second consumer) — checked 2026-11-13: no drift,
   the app reads exactly `sipDomain`/`websocketPath`, both stable.

## 5. Wire-fixture follow-ups — trigger check (2026-11-13)

- (a) android-side loopback live-check script: trigger "device-train prep
  wanting deployment-level proof" — NOT FIRED (the first device train is
  still operator-blocked). Deferral stands.
- (b) `/config.js` key fixture: trigger "first key drift or a second config
  consumer" — NOT FIRED (verified above). Deferral stands.
