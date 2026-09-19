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

| Capability    | Notes                                                                                                                                               |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Phone calls   | Multi-line, hold/focus/mute, blind + attended transfer, DTMF keypad                                                                                 |
| Resilience    | Bounded reconnect watchdog with re-registration; live calls survive                                                                                 |
| SMS & MMS     | Threads with unread badges, attachments in/out, delivery receipts, live updates                                                                     |
| Fax           | Send PDFs, receive documents, provider status (transmitted/failed), download                                                                        |
| Voicemail     | List, play, delete — straight from the PBX's per-extension API                                                                                      |
| Call history  | Server-side CDR records through the same API                                                                                                        |
| Contacts      | Shared (config) + personal (per extension), vCard import/export, click-to-dial                                                                      |
| Live updates  | Per-extension SSE feed: threads, open transcripts, fax list, voicemail                                                                              |
| Sign-in       | Credentials verified against the PBX directory server-side (401 on rejection, 502 if the PBX is down); the SIP island and the tabs share that login |
| Diagnostics   | Live ICE/media panel that names the suspected cause (e.g. blocked TURN)                                                                             |
| i18n / themes | Everything in English + German; dark + light themes with a manual toggle                                                                            |
| Deployment    | Single static binary, SQLite + content-addressed blob store, `/healthz`, NixOS module                                                               |
| Recording     | Done PBX-side by the telephony stack (stereo WAV, `/recordings/` behind operator auth); this app shows CDR history rows only                        |

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
nested lists (`ice_servers`, `contacts`) belong in the JSON file.

| Key                      | Default             | Meaning                                                                                                    |
| ------------------------ | ------------------- | ---------------------------------------------------------------------------------------------------------- |
| `addr`                   | `:8080`             | Listen address                                                                                             |
| `data_dir`               | `/var/lib/webphone` | SQLite DB + content-addressed attachment/fax files (created if missing)                                    |
| `sip_domain`             | _empty_             | SIP domain the island registers at (rendered into `/config.js`)                                            |
| `websocket_path`         | `/sip`              | WSS path the island connects to (must be proxied to the PBX)                                               |
| `phone_api_url`          | _empty_ = disabled  | Base URL of the per-extension API, e.g. `https://pbx.example.com`                                          |
| `session_ttl`            | `24h`               | Tab-session lifetime (cookie + server store)                                                               |
| `ice_servers`            | _empty_             | STUN/TURN entries handed to the island (`urls`, `username`, `credential`)                                  |
| `contacts`               | _empty_             | Shared directory entries (`name`, `number`)                                                                |
| `gateway.mode`           | `loopback`          | `loopback` or `webhook`                                                                                    |
| `gateway.webhook_url`    | _empty_             | Provider base URL in webhook mode (required there)                                                         |
| `gateway.webhook_secret` | _empty_             | Shared secret; **also guards the inbound `/hooks/*` endpoints (fail-closed: hooks return 503 without it)** |
| `csrf.trusted_proxies`   | _empty_             | Local proxies whose `X-Forwarded-Proto` may be believed (loopback nginx) — IP or CIDR entries              |
| `csrf.trusted_origins`   | _empty_             | Browser-facing origins counted as same-origin (the TLS vhost, e.g. `https://pbx.example.com`)              |

`csrf.*` matters whenever TLS ends at a proxy: a truthful browser POST
then arrives with `Origin: https://host` while the listener sees plain
HTTP, and unconfigured the CSRF middleware rejects it as a forged
same-origin attestation (403 on every POST, logins included). The
NixOS module sets both keys when `nginx.enable` is on.

Example file:

```json
{
  "addr": "127.0.0.1:8080",
  "data_dir": "/var/lib/webphone",
  "sip_domain": "pbx.example.com",
  "websocket_path": "/sip",
  "phone_api_url": "https://pbx.example.com",
  "session_ttl": "24h",
  "ice_servers": [{ "urls": ["stun:pbx.example.com:3478"] }],
  "contacts": [{ "name": "Support", "number": "2000" }],
  "gateway": {
    "mode": "webhook",
    "webhook_url": "http://127.0.0.1:8090",
    "webhook_secret": "long-random-string"
  },
  "csrf": {
    "trusted_proxies": ["127.0.0.1"],
    "trusted_origins": ["https://pbx.example.com"]
  }
}
```

## Integration contracts

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

```nginx
location / { proxy_pass http://127.0.0.1:8080; }
location /sip {                       # WebSocket → PBX sofia wss-binding
  proxy_pass https://pbx.internal:7443;
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
}
location /events {                    # SSE: no buffering
  proxy_pass http://127.0.0.1:8080;
  proxy_buffering off;
}
client_max_body_size 64m;             # attachments (≤5×10 MiB) + PDFs (≤20 MiB)
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
byte-identical):

1. stop the service,
2. copy `webphone.db` and `files/` back into the data directory,
3. start the service — sessions are in-memory by design, so nothing
   else to replay; sign in and the tabs render from the restored store.

Keep a copy off the machine: the snapshot directory is plain files, so
any rsync/restic pipeline can pick it up.

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

The NixOS module ships exactly these defaults when `nginx.enable` is set
(derived from `nginx.hostName`); a startup log line `csrf fronting
trustedProxies=… trustedOrigins=…` shows the effective shape at boot.

`nginx.hsts.enable` (default **off**) adds Strict-Transport-Security to
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
  nginx = {
    enable = true;
    hostName = "phone.example.org";
  };
};
```

`settings` is the same JSON config the binary reads (rendered to a
`WEBPHONE_CONFIG` file); secrets belong in `environmentFile`, not the
world-readable config. The generated vhost terminates TLS, proxies `/`
and upgrades the SIP WebSocket path with a long read timeout. A flake
check evaluates the module, so `nix flake check` catches breakage.

## Development

```console
nix develop                     # Go, templ, golangci-lint, esbuild, …
templ generate ./internal/web/views/   # after editing .templ files
GOEXPERIMENT=jsonv2 go test ./...      # required: templ-components uses encoding/json/v2
buildflow                       # the quality gate (format, lint, audit, checks)
nix flake check                 # package build + tests in the sandbox + treefmt
nix build .#webphone --system aarch64-linux   # cross-builds (pure Go)
./update.sh [version]           # repin vendored sip.js (esbuild IIFE bundle)
```

The island sources live in `internal/web/assets/island/app/` (ES modules,
served as-is, CSP-strict); their DOM element ids are a published contract
(see [AGENTS.md](AGENTS.md)) driven by the consuming stack's browser E2E.

## License

MIT — see [LICENSE](LICENSE). The vendored sip.js keeps its own license
notice (`internal/web/assets/vendor/sip.min.js.LEGAL.txt`).
