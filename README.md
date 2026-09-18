# WebPhone

A self-hosted browser phone. Call, hold, transfer, and pick up voicemail
from any device with a microphone — no plugin, no app install, no CDN.
One static page speaking SIP over WebRTC via [sip.js](https://github.com/onsip/SIP.js),
packaged as a single reproducible Nix derivation.

Built as the dedicated home for the phone UI of
[nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
(which consumes this flake as an input), but usable against any
FreeSWITCH/sofia (or compatible) WebSocket endpoint.

## What it does

| Capability          | Notes                                                                     |
| ------------------- | ------------------------------------------------------------------------- |
| Multi-line calling  | Several concurrent calls, hold/focus/mute per call                        |
| Transfer            | Blind and attended (REFER / REFER-with-Replaces, executed server-side)    |
| DTMF                | Keypad via `application/dtmf-relay` INFO (FreeSWITCH-compatible form)     |
| Incoming calls      | System notification, ring tone, tab-title flash, accept/reject            |
| Resilience          | Bounded reconnect watchdog with re-registration and call preservation     |
| Voicemail & history | In-browser playback/deletion and CDR history via a per-extension API      |
| Diagnostics         | Live ICE/media panel that names the suspected cause (e.g. blocked TURN)   |
| i18n                | English and German UI                                                     |
| Dark & light        | Full token-based themes following `prefers-color-scheme`                  |

## Get started

```console
nix build github:LarsArtmann/webphone          # static site in result/share/webphone
```

The build produces a self-contained static site:

```
share/webphone/
├── index.html            the phone
├── style.css             design tokens, dark + light
├── favicon.svg
├── app.js                the application (ES modules bundled by esbuild)
├── sip.min.js            pinned sip.js 0.21 as a browser global (SIP.*)
└── sip.min.js.LEGAL.txt  upstream license notice
```

Serve it over **HTTPS** and give it three same-origin companions:

1. `GET /config.js` — a runtime-rendered script setting `window.PBX_CONFIG`:

   ```js
   window.PBX_CONFIG = {
     sipDomain: "pbx.example.com",
     websocketPath: "/sip",
     iceServers: [{ urls: ["stun:pbx.example.com:3478"] }],
     phoneApi: false,        // enables the voicemail/history panels
     contacts: [{ name: "Support", number: "2000" }],
   };
   ```

2. `wss://<host>/sip` — a WebSocket (TLS) proxy to your SIP server's
   `wss` transport (FreeSWITCH: sofia's `wss-binding`).
3. Optionally `GET /phone-api/*` — the per-extension voicemail/CDR API
   (see the telephony stack for a reference implementation).

[nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
wires all of this up for you (nginx vhost, TLS, TURN credentials, the API):

```nix
inputs.webphone.url = "github:LarsArtmann/webphone";
# services.telephony.webphone.package defaults to this flake's package.
```

## Development

```console
nix build .#webphone     # the package
nix flake check          # build + treefmt + statix + deadnix
nix fmt                  # treefmt: nixfmt (nix) + prettier (js/html/css)
./package/update.sh      # repin sip.js (arg: version, default latest)
```

The UI sources live in `src/` as ES modules (`app/main.js` is the entry;
esbuild bundles them to the single served `app.js`). If you serve the
site behind the same strict CSP as the telephony stack
(`script-src 'self'`, no CDN, no webfonts) everything just works.

Browser end-to-end coverage (real registration + calls against a live
FreeSWITCH) lives in the consuming stack's test suite; the DOM element
contract those tests drive is documented in
[AGENTS.md](AGENTS.md) and must not be broken casually.

## License

MIT — see [LICENSE](LICENSE). The bundled sip.js keeps its own license
notice (`sip.min.js.LEGAL.txt`).
