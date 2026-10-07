# WebPhone

A self-hosted unified-communications web app: WebRTC phone calls, SMS/MMS
messaging, fax, voicemail, call history and contacts — one Go binary, one
page, one login (your PBX extension and directory password). No plugin, no
app install, no CDN, strict same-origin CSP.

The SIP call UI is a proven [sip.js](https://github.com/onsip/SIP.js) 0.21
island (multi-line calls, blind/attended transfer, DTMF, reconnect
watchdog, ICE diagnostics); everything around it — threads, faxes,
voicemail, history, contacts, live updates — is server-rendered
[templ](https://github.com/a-h/templ) + [HTMX](https://htmx.org), stored
in SQLite, and pushed to the browser over SSE.

Built as the dedicated home of the phone experience of
[nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony),
but usable against any SIP/WebSocket PBX (FreeSWITCH/sofia or compatible).

## What it does

| Capability      | Notes                                                                                                                                                                                                   |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Phone calls     | Multi-line, hold/focus/mute, blind + attended transfer, DTMF keypad                                                                                                                                     |
| Resilience      | Bounded reconnect watchdog with re-registration; live calls survive                                                                                                                                     |
| Call comfort    | Dial typeahead from your contacts, call state chip, missed-call badge, audio output picker for multi-output desks                                                                                       |
| SMS & MMS       | Threads with unread badges, attachments in/out, delivery receipts, live updates, search over remote + message bodies; pin/mute/archive a conversation, quick reply snippets, optimistic send with retry |
| Fax             | Send PDFs, receive documents, per-job status timeline, resend failed jobs, download; optional fire-and-forget Paperless-ngx archive of inbound faxes                                                    |
| Voicemail       | List, play (inline scrubber + speed control), delete — straight from the PBX's per-extension API                                                                                                        |
| Call history    | Server-side CDR records through the same API; outcome filters (missed/answered), day grouping, shareable filter URLs                                                                                    |
| Contacts        | Shared (config) + personal (per extension), vCard import/export, click-to-dial                                                                                                                          |
| Data export     | One session-gated download (Settings tab) zips every thread, fax job and personal contact of the signed-in extension                                                                                    |
| TURN auth       | Optional short-lived coturn REST credentials derived per response (`turn_rest.secret`) — long-lived TURN passwords never ship                                                                           |
| Live updates    | Per-extension SSE feed: threads, open transcripts, fax list, voicemail                                                                                                                                  |
| Sign-in         | Credentials verified against the PBX directory server-side (401 on rejection, 502 if the PBX is down); the SIP island and the tabs share that login                                                     |
| Diagnostics     | Live ICE/media panel with setup timings (gather, first media), pre-call mic/speaker self-test, English event log                                                                                        |
| i18n / themes   | Everything in English + German; dark + light themes with a manual toggle                                                                                                                                |
| Command palette | Ctrl/Cmd-K palette for tabs and actions (Call, New message, Cycle theme); `?` lists every keyboard shortcut                                                                                             |
| Accessibility   | Skip-to-content link, screen-reader live announcements, `aria-current` nav, labelled badges, focus moves to the new panel on swap                                                                       |
| Mobile          | Nav becomes a fixed bottom tab bar; a live call keeps the island front and center; 44px tap targets                                                                                                     |
| Deployment      | Single static binary, SQLite + content-addressed blob store, `/healthz`, five-part operator boot-error contract (exit 1 designed / 2 panic), NixOS module                                               |
| Recording       | Done PBX-side by the telephony stack (stereo WAV, `/recordings/` behind operator auth); this app shows CDR history rows only                                                                            |

## Quick start

```console
nix run github:LarsArtmann/webphone   # or: nix build .#webphone && ./result/bin/webphone
```

With zero configuration the server starts on `:8080` in **loopback**
gateway mode: messages and faxes are accepted instantly and marked
sent/transmitted, so the whole product is explorable without a PBX.

```console
curl -fsS http://127.0.0.1:8080/healthz   # readiness -> {"status":"ok",...} (sqlite ping + blob-dir write probe)
curl -fsS http://127.0.0.1:8080/livez     # liveness  -> 200 while the process serves (no checks run)
curl -fsS http://127.0.0.1:8080/startupz  # startup   -> 503 until the backing resources first pass, then latched 200
```

All three are session-free GETs whose bodies name checks and statuses
only. Readiness stays `/healthz` alone; `/livez` deliberately runs no
checks so a prober can tell "wedged process" from "degraded
dependencies".

### Health dashboard (optional, off by default)

Set `WEBPHONE_DASHBOARD__ENABLE=true` (or `dashboard.enable` in the
config file) and `/health` serves a live operator dashboard — check
cards over SSE (go-health-dashboard), the container's health checks
(sqlite, blob-dir, plus the dashboard's own pusher staleness as a
non-critical warn), a status trend, and JSON on `Accept:
application/json`. The page is CSP-strict: per-request nonces, the
Datastar SDK and its stylesheet served same-origin from the binary
(`unsafe-eval` is scoped to this subtree only, because the SDK compiles
its expressions). Probe aliases live at `/health/livez`,
`/health/readyz`, `/health/startupz` — same probe as the root triple,
never a second truth. Treat the subtree like the probe triple in front
of it: it is an operator surface, so fence it (the NixOS module's
Caddy vhost already proxies it unbuffered for the SSE stream).
`dashboard.title` names the deployment on the page.

Open the page, sign in on the phone panel (any extension format your PBX
accepts; against nothing it will just fail to register), and the tabs
unlock with the same login.

For real use you want, in front of or around the binary:

1. **TLS + `wss://<host>/sip`** proxied to your PBX's WebSocket transport
   (FreeSWITCH: sofia `wss-binding`). The island speaks SIP over that
   socket.
2. Optionally a **phone API** upstream (voicemail + CDR history).
3. Optionally a **gateway** for real SMS/MMS/fax delivery (webhook mode).

## Configuration

Config comes from an optional JSON file (path via `WEBPHONE_CONFIG`,
default `/etc/webphone/config.json`), overridden by `WEBPHONE_*`
environment variables. Env keys use `__` to nest:
`WEBPHONE_GATEWAY__MODE=webhook` → `gateway.mode`; single underscores stay
literal: `WEBPHONE_DATA_DIR` → `data_dir`. Scalars are env-friendly;
nested lists (`ice_servers`, `contacts`) and maps (`identities`) belong
in the JSON file.

| Key                                     | Default             | Meaning                                                                                                                                                                                                              |
| --------------------------------------- | ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `addr`                                  | `:8080`             | Listen address                                                                                                                                                                                                       |
| `data_dir`                              | `/var/lib/webphone` | SQLite DB + content-addressed attachment/fax files (created if missing)                                                                                                                                              |
| `sip_domain`                            | _empty_             | SIP domain the island registers at (rendered into `/config.js`)                                                                                                                                                      |
| `websocket_path`                        | `/sip`              | WSS path the island connects to (must be proxied to the PBX)                                                                                                                                                         |
| `phone_api_url`                         | _empty_ = disabled  | Base URL of the per-extension API, e.g. `https://pbx.example.com`                                                                                                                                                    |
| `session_ttl`                           | `7d`                | Sliding idle window: activity past its halfway point renews the session (a regularly used device never re-signs-in)                                                                                                  |
| `session_max_ttl`                       | `30d`               | Absolute session lifetime from sign-in — the cap that re-signs even a continuously renewed (or stolen) session                                                                                                       |
| `retention_days`                        | `0`                 | 0 keeps everything forever; a positive value makes a daily sweep delete messages (with attachments), fax jobs (with documents) and empty threads older than that many days                                           |
| `timezone`                              | _system local_      | IANA zone name owning every wall-clock rendering and log line; a typo fails validation instead of silently rendering UTC                                                                                             |
| `ice_servers`                           | _empty_             | STUN/TURN entries handed to the island (`urls`, `username`, `credential`)                                                                                                                                            |
| `turn_rest.secret`                      | _empty_             | coturn REST API shared secret (must match coturn's `static-auth-secret`); when set, `/config.js` derives short-lived username/credential pairs for every `turn:`/`turns:` entry instead of shipping static passwords |
| `turn_rest.ttl`                         | `48h`               | Validity window of the derived TURN credentials (per-response freshness; the pair expires with the username's unix timestamp)                                                                                        |
| `contacts`                              | _empty_             | Shared directory entries (`name`, `number`)                                                                                                                                                                          |
| `identities`                            | _empty_             | Extension → presented number (DID) shown as the user's own number (header, whoami, composers); display-only                                                                                                          |
| `gateway.mode`                          | `loopback`          | `loopback` or `webhook`                                                                                                                                                                                              |
| `gateway.webhook_url`                   | _empty_             | Provider base URL in webhook mode (required there)                                                                                                                                                                   |
| `gateway.webhook_secret`                | _empty_             | Shared secret; **also guards the inbound `/hooks/*` endpoints (fail-closed: hooks return 503 without it)**                                                                                                           |
| `gateway.webhook_secret_file`           | _empty_             | Read the shared secret from a runtime file at startup (single line) instead — one file both sides read beats keeping an env copy and the receiver's copy in sync; at most one of the two may be set                  |
| `csrf.trusted_proxies`                  | _empty_             | Local proxies whose `X-Forwarded-Proto` may be believed (the loopback front) — IP or CIDR entries                                                                                                                    |
| `csrf.trusted_origins`                  | _empty_             | Browser-facing origins counted as same-origin (the TLS vhost, e.g. `https://pbx.example.com`)                                                                                                                        |
| `crm.url`                               | _empty_ = disabled  | Ledger CRM base URL for the optional integration (e.g. `http://127.0.0.1:8080`)                                                                                                                                      |
| `crm.token`                             | _empty_             | Bearer token of the CRM's machine API (`-api-token` there); both keys together or neither                                                                                                                            |
| `paperless.url`                         | _empty_ = disabled  | Paperless-ngx base URL for the optional inbound-fax archive (e.g. `http://127.0.0.1:2280`); both keys together or neither                                                                                            |
| `paperless.token`                       | _empty_             | Paperless-ngx API token (Profile → My Profile → API token)                                                                                                                                                           |
| `asr.url`                               | _empty_ = disabled  | Base URL of the optional speech-to-text provider (e.g. `http://127.0.0.1:8081`); a URL alone enables — the OpenAI-compatible `/v1/audio/transcriptions` endpoint is assumed                                                       |
| `asr.token`                             | _empty_             | Optional bearer credential for the ASR provider; a token WITHOUT a URL fails closed (nowhere to send it)                                                                                                             |
| `asr.model`                             | `whisper-1`         | Model name the provider expects (e.g. `large-v3`); sent as the multipart `model` field                                                                                                                                |
| `auth.passkey.rp_id`                    | _empty_ = disabled  | WebAuthn relying-party ID (the registrable domain, e.g. `pbx.example.com`); setting the passkey keys enables the email-first login — all-or-nothing per validation                                                   |
| `auth.passkey.rp_display_name`          | `webphone`          | Name the browser shows in the passkey prompt                                                                                                                                                                         |
| `auth.passkey.rp_origins`               | _required_          | Browser-facing origins (`https://…`); each origin's host must equal `rp_id`                                                                                                                                          |
| `auth.passkey.users`                    | _required_          | Email → `{extensions: ["1000"], display_name: "Lars"}`: the mapping a verified passkey resolves to (first extension binds the session; every mapped extension's `identities` number shows in the whoami line)        |
| `auth.passkey.extension_password_files` | _required_          | Extension → runtime file carrying its SIP directory password (single line, read at login time, never cached; empty/missing fails closed)                                                                             |

`csrf.*` matters whenever TLS ends at a proxy: a truthful browser POST
then arrives with `Origin: https://host` while the listener sees plain
HTTP, and unconfigured the CSRF middleware rejects it as a forged
same-origin attestation (403 on every POST, logins included). The
NixOS module sets both keys when `caddy.enable` is on.

Example file:

```json
{
  "addr": "127.0.0.1:8080",
  "data_dir": "/var/lib/webphone",
  "sip_domain": "pbx.example.com",
  "websocket_path": "/sip",
  "phone_api_url": "https://pbx.example.com",
  "session_ttl": "168h",
  "session_max_ttl": "720h",
  "ice_servers": [{ "urls": ["stun:pbx.example.com:3478"] }],
  "contacts": [{ "name": "Support", "number": "2000" }],
  "identities": { "1001": "+49 30 12345678" },
  "gateway": {
    "mode": "webhook",
    "webhook_url": "http://127.0.0.1:8090",
    "webhook_secret": "long-random-string"
  },
  "csrf": {
    "trusted_proxies": ["127.0.0.1"],
    "trusted_origins": ["https://pbx.example.com"]
  },
  "auth": {
    "passkey": {
      "rp_id": "pbx.example.com",
      "rp_origins": ["https://pbx.example.com"],
      "users": {
        "lars@example.com": { "extensions": ["1000"], "display_name": "Lars" }
      },
      "extension_password_files": {
        "1000": "/var/lib/telephony-secrets/ext_1000"
      }
    }
  }
}
```

## Integration contracts

### Ledger CRM (optional)

With `crm.url` + `crm.token` configured, the webphone talks to a
Ledger CRM instance's machine API (the
CRM must run with `-api-token <same-token>`):

- **Caller names**: numbers rendered in History, Messages, Fax and
  Voicemail are resolved against the CRM's contacts (matched by digits,
  tolerant of trunk-prefix and country-code variants); matches render the
  contact name, everything else stays the raw number. Lookups are cached
  in-process (6 h, misses 5 min) and every failure degrades to the raw
  number — the phone never depends on the CRM.
- **Call journal**: after every call the island reports the outcome to
  `POST /api/calls` (session-gated); the server journals it on the
  matching CRM contact (`call_logged`: "Incoming call from +49… (2m 03s
  — answered)"). Unknown numbers are deliberately NOT logged — the
  integration never mints contacts.

Enable it on both sides: webphone gets `crm.url` + `crm.token`, the CRM
gets `-api-token <same value>` (its machine API is unmounted without
that flag).

### Paperless-ngx fax archive (optional)

With `paperless.url` + `paperless.token` configured, every INBOUND fax
is additionally filed into a Paperless-ngx instance (the storage truth
stays webphone's blob store — Paperless is a downstream copy):

each document uploads tagged `fax`, typed `Fax`, titled
`Fax from <number> <date>`, and carries the webphone fax id in the
`webphone-fax-id` custom field (the provenance link back). A server-side
content-hash duplicate refusal is inert — Paperless keeps one copy.
Archiving is fire-and-forget after the fax is persisted: a slow or dead
Paperless never delays or fails a fax, and failures only WARN in the
log. Outbound faxes are not archived (v1 scope).

### Passkey sign-in (optional)

With `auth.passkey.*` configured (OFF by default — zero config keeps
the login card byte-identical), the webphone embeds the
cqrs-htmx/usermgmt identity layer in-process and offers email + passkey
logins next to the extension form:

- **Login order is inverted**: `POST /api/auth/passkey/finish` verifies
  the ceremony, resolves the configured email→extension mapping
  (`auth.passkey.users`), sources the extension's SIP directory
  password from `auth.passkey.extension_password_files` (read per
  login, never cached — empty/missing fails CLOSED), verifies it
  against the PBX directory, and only then mints the session; the
  browser REGISTERs afterwards. The response is the typed
  session-identity shape (`extension`, `did`, `display_name`,
  `numbers`, plus the `password`, served no-store) — the same struct
  `GET /api/session` answers on resume.
- **Enrollment is CLI-minted**: `webphone -enroll-passkey <email>`
  prints a one-time link (`/enroll?token=…`; the token is stored
  sha256-hashed, expires after 15 minutes, and burns at first verify —
  unknown, expired and used tokens all answer the same 503, so the
  page never leaks token state). The standalone `/enroll` page runs
  verify → WebAuthn registration → finish without the island runtime.
- **Health**: with the mode on, `/healthz` gains a `userauth` leg
  probing the identity layer's own `usermgmt.db` (kept separate from
  `webphone.db` on purpose); a 503 naming `userauth` means that
  database is the broken leg.
- **Anti-enumeration**: unknown emails and credential-less accounts
  answer the same 401; the login flood bucket is shared with the
  extension login; operator-fixable rejections answer 503 with the
  honest `userauth.*` code in the journal (full table:
  docs/error-contract.md).

### The page (`window.PBX_CONFIG`)

The server renders `GET /config.js` from its own config — same contract
the static-site era consumed:

```js
window.PBX_CONFIG = {
  sipDomain: "pbx.example.com",
  websocketPath: "/sip",
  iceServers: [{ urls: ["stun:pbx.example.com:3478"] }],
  phoneApi: false,
  contacts: [{ name: "Support", number: "2000" }],
};
```

### Phone API (upstream, per-extension)

The server calls these with the signed-in extension's credentials (Basic
auth), both directly and through the transparent `/phone-api/*` proxy the
island uses:

| Method & path                                       | Response                                                                         |
| --------------------------------------------------- | -------------------------------------------------------------------------------- |
| `GET /phone-api/history?limit=N`                    | `{"entries":[{caller_id_number, destination_number, start, billsec, …}]}`        |
| `GET /phone-api/voicemail/{ext}/summary`            | `{"new":1,"old":0}`                                                              |
| `GET /phone-api/voicemail/{ext}/messages`           | `{"messages":[{uuid, cid_number, cid_name, seconds, created, read, audio_url}]}` |
| `DELETE /phone-api/voicemail/{ext}/messages/{uuid}` | _(2xx)_                                                                          |

### Outbound gateway (webhook mode)

The webphone POSTs `multipart/form-data` to `{webhook_url}/message` and
`{webhook_url}/fax` with `Authorization: Bearer {webhook_secret}`:

- fields: `kind` (`message`|`fax`), `owner` (sending extension), `to`, `body`
  (messages)
- files: one part per attachment (`attachment`) or the PDF (`document`)

The provider answers 2xx with `{"provider_ref": "..."}` — that reference
is what later status callbacks quote. Anything else marks the message/fax
failed.

### Inbound hooks (this server receives)

All hooks require `Authorization: Bearer {webhook_secret}` and are
**closed (503) when no secret is configured**.

```
POST /hooks/message        {"owner":"1001","from":"+441632960961","body":"hi",
                            "attachments":[{"name":"pic.jpg","mime_type":"image/jpeg","data_base64":"…"}]}
POST /hooks/fax            {"owner":"1001","from":"+441700000000","pages":2,
                            "provider_ref":"gw-123","pdf_base64":"…"}
POST /hooks/fax/status     {"provider_ref":"gw-123","status":"transmitted|failed","pages":2,"error":"…"}
POST /hooks/message/status {"provider_ref":"gw-123","status":"delivered|failed","error":"…"}
```

Providers may retry status callbacks: replays are deduped and answer
`202 Accepted` inertly (only successes are recorded, so failures stay
retryable).

### Bridging FreeSWITCH (example)

Any component that can POST JSON bridges the PBX; the payloads above are
the whole contract. A minimal two-way demo bridge (receives mod_sms
traffic, forwards as SMS; receives the webphone's multipart, hands to the
PBX) fits in ~30 lines of Python — this shape:

```python
# inbound: chatplan/lua/HTTP-hook on the PBX side calls
POST http://webphone:8080/hooks/message
     {"owner": "$to_ext", "from": "$from", "body": "$body"}
# outbound: receive the webphone's multipart at {bridge}/message and
# feed it into FreeSWITCH (event socket, mod_sms, or a carrier API),
# then answer {"provider_ref": "<id>"} and later call
# /hooks/fax/status or /hooks/message/status with that id.
```

### Live transcription (optional)

With `asr.url` set, every audio surface can be transcribed through one
session-gated endpoint. `POST /api/transcribe` takes the raw audio bytes
(any `audio/*` Content-Type) — optionally `?filename=` and `?lang=` — and
forwards them as a multipart file to the provider's OpenAI-compatible
`/v1/audio/transcriptions`, returning `{"text": "…"}`. The provider is any
OpenAI-compatible server; a disabled seam answers 404 and the UI hides
every affordance.

- **Live calls**: the island mixes the remote party and the operator's own
  voice through an `AudioContext` (closed on stop) and records 4-second
  segments, appending each segment's text into the call card. Capture
  pauses while the call is on hold (no audio flows — no silence
  hallucinations) and each segment carries the UI language as the
  provider's `language` hint.
- **Voicemail & MMS audio**: the Voicemail and Messages tabs render a
  transcribe button next to each playable clip (one delegated shell.js
  handler fetches the audio and posts it).
- **Auto-start**: with the seam on, transcription begins by itself — live
  capture starts when a call connects, and every rendered voicemail/MMS
  clip transcribes once per page life (morph re-renders never re-POST).
  The buttons stay as manual re-runs.
- **Persistence**: each live-call segment is also appended (fire-and-forget,
  `POST /api/transcripts`) to an owner-scoped SQLite table (schema v3,
  capped at 5000 segments per extension). The History tab renders the
  newest transcribed calls as their own "Call transcripts" section —
  deliberately NOT joined onto the CDR rows, which carry no call uuid to
  correlate with.

**Choosing a provider (2026 recommendation):** run ASR LOCAL, not hosted.
Call audio of real people is GDPR personal data; the US-inference APIs
(OpenAI, Groq, AssemblyAI) offer no EU residency. The stack already keeps
audio on your infrastructure — keep transcription there too.

- **Primary**: [Speaches](https://github.com/speaches-ai/speaches) (ex
  faster-whisper-server) with `large-v3-turbo` — first-class
  OpenAI-compatible `/v1/audio/transcriptions`, ships a flake.nix, CPU-viable
  for 4-second segments, MIT. Set `asr.url` to its base URL; no token
  needed.
- **Lightweight fallback**: whisper.cpp's `whisper-server` with a
  quantized `large-v3-turbo` (single static binary; its endpoint is
  `/inference`, so rewrite the path in the reverse proxy).
- **If cloud is ever required**: Deepgram Nova-3 via the EU endpoint
  (`api.eu.deepgram.com`, telephony-tuned models) behind a thin
  OpenAI-compatible adapter.
- Pass an explicit language where you can (`?lang=de` — auto-detect
  misfires DE↔EN on short phone chunks); the island does this
  automatically from the session language.

## Live updates (SSE)

Signed-in extensions get a per-extension feed at `GET /events` (HTMX SSE
extension). Events: `threads` (message list rows), `thread` (open
transcript bubbles), `fax` (fax list rows) — each carries an
innerHTML-safe fragment — and `voicemail`, a payload-less nudge that makes
an open voicemail tab re-fetch itself. The island's voicemail polls are
forwarded as nudges too, so finishing a call refreshes the tab.

## Deployment

The binary owns HTTP on `addr`; TLS, the WSS `/sip` proxy to the PBX, and
process supervision belong to the serving stack
([nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
today) or any reverse proxy in front:

```caddy
handle /sip {                      # WebSocket → PBX sofia wss-binding
  reverse_proxy https://pbx.internal:7443
}
handle /events {                   # SSE: unbuffered
  reverse_proxy 127.0.0.1:8080 {
    flush_interval -1
  }
}
handle {
  reverse_proxy 127.0.0.1:8080
}
request_body {
  max_size 64MB                     # attachments (≤5×10 MiB) + PDFs (≤20 MiB)
}
```

### Backups and restore

The state directory (`data_dir`, default `/var/lib/webphone`) holds
everything: `webphone.db` (SQLite) and `files/` (the content-addressed
blob tree for attachments and fax documents). Back up both together.

Online, per snapshot (no downtime — what `services.webphone.backup.*`
in the NixOS module wires as a daily timer):

```console
sqlite3 /var/lib/webphone/webphone.db ".backup '/var/lib/webphone-backup/webphone.db'"
rsync -a --delete /var/lib/webphone/files/ /var/lib/webphone-backup/files/
```

Restore (drill-verified cold path — `scripts/webphone-backup-drill.py`
boots a real binary, loads a webhook message with a binary attachment,
tars the data dir, restores it to scratch and pulls the attachment back
byte-identical). Re-run the drill any time the store layout changes or
after an upgrade:

```console
python3 scripts/webphone-backup-drill.py   # end-to-end restore drill (boots a throwaway binary + dirs)
```

1. stop the service,
2. copy `webphone.db` and `files/` back into the data directory,
3. start the service — tab sessions live in the SQLite database, so the
   restored session rows keep working (signed-in browsers stay signed
   in); everything else renders straight from the restored store.

Dated history (point-in-time restore): set
`services.webphone.backup.retentionDays` (default `null` = single
latest snapshot, the historical behavior). When set — e.g. `30` — each
daily run additionally writes `destDir/snapshots/<date>/` (db + blob
tree, unchanged blobs hardlinked against the previous snapshot, so a
month of history typically costs little more than one snapshot) and
deletes snapshot directories older than the given days. Restore from a
specific day: stop the service, copy `snapshots/<date>/webphone.db`
and `snapshots/<date>/files/` back into the data directory, start it.

Keep a copy OFF the machine: the snapshot directory (`destDir`) lives on
the same host as the service, so it is not a backup yet — a disk loss
takes both. Point an off-machine restic/borg repository (or any remote
rsync job) at `destDir`; the files are plain sqlite + blobs, so any
file-level backup tool handles them.

### Troubleshooting: 403 logins behind a TLS proxy

**Symptom:** every browser login (and every form POST) returns 403 once the
app is served behind a TLS-terminating proxy, while calls keep working and
`/healthz` is green. The server log names the cause exactly:

```console
WARN httputil: CSRF rejected request with forged same-origin attestation method=POST path=/api/session origin=https://pbx.example.org
```

**Why:** the browser truthfully sends `Origin: https://…` while the listener
sees plain HTTP; unless the proxy is trusted, that reads as a forged
attestation.

**Fix:** declare the fronting shape so the proxy's `X-Forwarded-Proto` is
believed and the https origin is trusted:

```json
{
  "csrf": {
    "trusted_proxies": ["127.0.0.1"],
    "trusted_origins": ["https://pbx.example.org"]
  }
}
```

The NixOS module ships exactly these defaults when `caddy.enable` is set
(derived from `caddy.hostName`); a startup log line `csrf fronting
trustedProxies=… trustedOrigins=…` shows the effective shape at boot.

`caddy.hsts.enable` (default **off**) adds Strict-Transport-Security to
the generated vhost; keep it off until the deployment is genuinely
https-only — HSTS pins browsers to https for `maxAge` (default 2 years).

### NixOS module

The flake ships `nixosModules.default` so the binary and its deployment
shape stay in sync:

```nix
inputs.webphone.nixosModules.default

services.webphone = {
  enable = true;
  package = inputs.webphone.packages."${pkgs.system}".webphone;
  settings = {
    sip_domain = "pbx.example.com";
    phone_api_url = "https://pbx.example.com";
    ice_servers = [{ urls = [ "stun:pbx.example.com:3478" ]; }];
  };
  environmentFile = "/run/secrets/webphone-env";   # WEBPHONE_GATEWAY__WEBHOOK_SECRET
  caddy = {
    enable = true;
    hostName = "phone.example.org";
    sipUpstream = "https://pbx.example.com:7443";   # bridge the SIP WebSocket to the PBX
  };
};
```

`settings` is the same JSON config the binary reads (rendered to a
`WEBPHONE_CONFIG` file); secrets belong in `environmentFile`, not the
world-readable config. The generated Caddy vhost terminates TLS
(automatic HTTPS) and proxies the app; with `caddy.sipUpstream` set it
also bridges the SIP WebSocket path to the PBX — without the bridge
the deployment must route that path itself. A flake check evaluates
the module, so `nix flake check` catches breakage, and a kvm-gated VM
test (`checks.x86_64-linux.webphone-backup`) proves the backup story
end to end.

Module options beyond `enable`/`package`/`settings`:

| Option                            | Default                    | Meaning                                                                                         |
| --------------------------------- | -------------------------- | ----------------------------------------------------------------------------------------------- |
| `dataDir`                         | `/var/lib/webphone`        | State directory (must stay under `/var/lib/` — asserted)                                        |
| `environmentFile`                 | _none_                     | systemd EnvironmentFile for secrets (`WEBPHONE_GATEWAY__WEBHOOK_SECRET`)                        |
| `environmentFiles`                | `[]`                       | Additional EnvironmentFiles loaded after `environmentFile` (later files win on dup keys)        |
| `memoryMax`                       | _uncapped_                 | systemd MemoryMax for the service                                                               |
| `csrf.trustedProxies`             | `[]`                       | Typed front for `settings.csrf.trusted_proxies`; beats the caddy-derived default when set       |
| `csrf.trustedOrigins`             | `[]`                       | Typed front for `settings.csrf.trusted_origins`; beats the caddy-derived default when set       |
| `serverTiming.enable`             | `false`                    | Server-Timing response headers (sets `WEBPHONE_DEBUG_TIMING=1`)                                 |
| `backup.enable`                   | `false`                    | Daily online snapshot timer (sqlite `.backup` + blob rsync)                                     |
| `backup.destDir`                  | `/var/lib/webphone-backup` | Snapshot destination                                                                            |
| `backup.calendar`                 | `*-*-* 04:30:00`           | Timer schedule                                                                                  |
| `backup.retentionDays`            | `null`                     | When set (e.g. `30`): daily dated `snapshots/<date>/` history + prune older than N days         |
| `caddy.enable` / `caddy.hostName` | _off_                      | Generated TLS vhost proxying the app (derives the csrf fronting defaults)                       |
| `caddy.sipUpstream`               | `null`                     | Bridge the SIP WebSocket path to the PBX (e.g. `https://pbx:7443`); null = deployment routes it |
| `caddy.hsts.enable` / `maxAge`    | _off_ / 2y                 | Strict-Transport-Security on the generated vhost                                                |

**Health probes behind the vhost:** `/healthz` (readiness), `/livez`
(process liveness) and `/startupz` (startup completion) ride the same
reverse proxy as the app. All three are session-free GETs whose bodies
name checks and statuses only, never secrets — safe to expose or
scrape. Fence a fleet health hub the Caddy way with a `remote_ip`
matcher in the vhost's `extraConfig` instead of per-location blocks;
`/metrics` (Prometheus text, aggregate counts only — never
per-extension data) fences the same way.

**Readiness vs systemd:** the service unit stays `Type=simple` by
DELIBERATE decision — the module does NOT wire `Type=notify`.
sd_notify would fire when the listener binds, which is a WEAKER
readiness signal than `/startupz` (503 until sqlite ping + blob-dir
write both first-pass, then latched 200), and a systemd watchdog
restart would kill live calls on a transient stall. Consumers that
need ordered startup should poll `/startupz` (or front it with a
one-shot `ExecStart=curl --retry` wait unit and `After=` ordering);
ongoing health belongs to `/healthz`, liveness to `/livez`. When the
boot itself fails, the binary prints a five-part operator report
(WHAT / REASSURE / WHY / FIX / ESCAPE, exit 1; a panic exits 2) —
the class table and fix hints live in
`docs/error-contract.md` § "Boot surface".

## Development

```console
nix develop                     # Go, templ, golangci-lint, esbuild, …
templ generate ./internal/web/views/   # after editing .templ files
go test ./...                          # json/v2 stable since Go 1.27
buildflow                       # the quality gate (format, lint, audit, checks)
nix flake check                 # package build + tests in the sandbox + treefmt
nix build .#webphone --system aarch64-linux   # cross-builds (pure Go)
./update.sh [version]           # repin vendored sip.js (esbuild IIFE bundle)
```

The flake is a slim entry point; its packages, checks (incl. the KVM
backup-VM test), NixOS module checks, devShell and formatting live as
flake-parts modules under [`nix/`](nix/).

The island sources live in `internal/web/assets/island/app/` (ES modules,
served as-is, CSP-strict); their DOM element ids are a published contract
(see [AGENTS.md](AGENTS.md)) driven by the consuming stack's browser E2E.

## License

MIT — see [LICENSE](LICENSE). The vendored sip.js keeps its own license
notice (`internal/web/assets/vendor/sip.min.js.LEGAL.txt`).
