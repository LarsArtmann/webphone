#!/usr/bin/env python3
"""Visual capture harness (T23): boots a fresh webphone binary (loopback
gateway — the whole product, zero PBX), seeds deterministic content over
HTTP, then drives headless chromium (selenium — the SAME driver family
the consuming stack's browser E2E uses) through the server tabs and
screenshots each in light and dark: the 24-shot matrix (12 surfaces x 2
themes).

The tenth surface is the styled 404 (`/nope`): it is a real page with
chrome, so it belongs in the matrix rather than being merely asserted —
a regression to stdlib text would show up as a broken shot, not just a
red test.

loginpage extension (2026-10-25, owner GO): the boot ALSO turns the
passkey mode on (a render-only config — no ceremony ever runs), and two
ANONYMOUS surfaces join the matrix ahead of the session'd tabs:
`login-card` (the island card with its /login link) and `login-page`
(the cqrs-htmx loginpage library page at /login — its own stylesheet,
its own ThemeScript). The session cookie is dropped for these two shots
and re-injected after (with a live session /login 303s home and the
card never renders). The library page follows prefers-color-scheme, NOT
the shell's wp-theme key (the documented forced-theme caveat), so each
theme pass emulates the media feature via CDP — which is exactly the
color a dark-OS user sees.

M13 extension (2026-10-08): the boot turns the ASR seam ON (a fake
provider BASE url — shell.js AUTO-transcribes seen audio, so the call
DOES go out and fail fast against the refused port; the button and
target div are the evidence) and three surfaces join the matrix: the Settings ASR row with provider detail, a History
transcript search (seeded via POST /api/transcripts), and a thread with
an AUDIO attachment whose transcribe button carries the same
data-transcribe-src shape the voicemail rows use (a bare boot has no
voicemail rows — PBX-owned — so the attachment variant is the
deterministic stand-in; the island's live call card needs a real SIP
call and stays covered by the island node specs).

CRM extension (2026-10-29): `--crm` boots a fake Kith machine API beside
the app and turns the seam on (url + token + a `contact_url` template),
so the CRM surfaces become EVIDENCE instead of claims: the `crm-context`
surface asserts the linked caller name (`.wp-crm-link`) AND the contact
context strip (`#wp-crm-context`) in the DOM before the shot. Without the
flag the seam stays off and the surface is skipped, which keeps the
default matrix exactly what a zero-config deployment renders.

Fax capability extension (2026-10-30): `--fax-off` boots with
`fax.send`/`fax.receive` explicitly false — what a deployed PBX without
a fax DID declares — and TIGHTENS the fax surface's DOM assertion from
"form or notice" to "notice present, compose form ABSENT", so a shot
cannot pass while the UI still lies about send/receive. Pair it with
`--surface fax` (plus `--surface settings`) for the targeted evidence
shot, or run the full matrix to see every surface as the off-deployment
renders it.

Remote/prod tracking extension (2026-11-10): `--base-url https://host`
points the SAME harness at a RUNNING deployment instead of a local boot
— no binary, no seeding; the session is minted over POST /api/session
with WEBPHONE_CAPTURE_EXTENSION / WEBPHONE_CAPTURE_PASSWORD (the PBX
directory verifies them for real; nothing is ever logged) and injected
into the browser exactly like the seed cookie. Markers run LENIENT
(warn + keep the shot) because real data legitimately lacks the seeded
fixtures, styled-404 surfaces are skipped, and the thread surfaces
become the first --threads (default 6) LIVE deep links off the real
list. The browser takes a REAL SIP registration for the run (the wss is
bridged on a deployed stack) — shots are honest, and the registration
dies with the run. Shots carry REAL customer data: ui-shots/ is
gitignored, keep it that way.

Screenshot tracking (`--track`): each run lands in
`<out-dir>/<host>/<UTC-stamp>/` (history is NEVER wiped), with a per-run
gallery.html + summary.json (deployment /version, --label, shot ledger)
and a host-level compare.html diffing the newest run against the
previous one — per surface: previous | current | pixel-diff heat map
(pillow; side-by-side only without it) with a changed-% badge. The WHOLE
page is captured (CDP captureBeyondViewport), not just the 1280x900
window, so below-the-fold layout rot is visible; `--viewport-only` opts
out and `--viewport WxH` (repeatable) adds extra passes (e.g. 390x844
for the mobile layout).

Conditional surfaces (2026-11-10): the matrix follows the deployment, not
a hardcoded list — `/enroll` joins the anonymous pass when the page
actually renders the form (probe GET /enroll, never a guessed link),
`admin` joins when the session's nav carries the admin link, and
`crm-context` joins when a live thread fragment carries the CRM strip.
Each conditional surface carries its own marker pin; the styled-404
skip stays the backstop for capabilities withheld mid-run.

Instrument honesty (2026-11-10): `--self-test` verifies every pure
helper (shot keys, viewport parsing, link scraping, marker
normalization, clip math, transcript probe words, prune retention, the
partial-compare tolerance — pixel math too, when pillow is installed)
without chromium; `--dry-run` logs in, resolves the surface list and
exits BEFORE a browser starts (the credential/surface sanity lane);
a ringing inbound call during a surface shot is detected (`#incoming-call`)
and that shot is re-taken once after a settle wait, so a run never
silently archives a banner shot as a clean matrix cell; tracked runs
additionally snapshot the wire facts (`/config.js` + the Settings
server section, verbatim) into `wire-facts.json`.

Diff controls (2026-11-10, calibration pending real data — defaults
preserve today's behavior): `--diff-threshold N` is the per-pixel
delta bar (default 12), `--badge-bar PCT` the compare badge/noise bar
(default 0.5%), and `--diff-clip N` enables the height-drift policy:
when a full-page shot's height changed, both images are cropped to
their common top N pixels and the diff runs on the clipped region
(marked `clipped` in summary.json) instead of skipping to
side-by-side.

TLS smoke recipe (the passkey-smoke shape, local only): the binary
serves PLAIN HTTP — put a TLS front in front of it
(`nix shell nixpkgs#caddy`, a Caddyfile with `https://webphone.test:18443`
+ `tls internal` + `reverse_proxy 127.0.0.1:18097`), then run the REMOTE
lane against it: `--base-url https://webphone.test:18443 --insecure
--host-map webphone.test=127.0.0.1` (chrome gets the `acceptInsecureCerts`
capability + a host-resolver rule; the HTTP probes skip verification
and resolve through the same map). This is
the only lane where verification may be skipped — NEVER point
--insecure at a production host. Note the login lane answers 429 when
rate limited and stays MANUAL by ruling: no auto-retry, a retry loop
against the per-IP limiter locks the operator out.

Fixture recipe (bare-boot remote lane): the local boot writes its
config to `/tmp/wp-visual-data/ui-capture-config.json` and seeds
extension 1001 with password `pw`; a manual boot with
`WEBPHONE_CONFIG=/tmp/wp-visual-data/ui-capture-config.json` on the
same port is a valid --base-url target for exercising the remote lane
without touching a real deployment.

Partial runs are compared gracefully (pinned by --self-test): a run
dir missing shots (interrupted pass) diffs the shots it has and
renders explicit `no previous`/`no current` boxes for the rest —
an interrupted run never corrupts the chain.

Budget decision (documented, 2026-10-02): this is a LOCAL harness, NOT a
flake check — headless chromium cannot run inside a nix build sandbox
without a KVM VM, and a KVM browser VM here would duplicate the stack's
proven browser-E2E lane at double the gate cost. The shots are eyeball
material (and optional pixel-diff input for the owner); timestamps and
relative times make byte-goldens flake by design.

Run (from the repo root):
  nix build .#webphone --out-link /tmp/wp-visual-bin
  nix shell nixpkgs#chromium nixpkgs#chromedriver \
    nixpkgs#python312.withPackages(ps: [ ps.selenium ps.pillow ]) \
    --command python3 scripts/ui-capture.py --binary /tmp/wp-visual-bin/bin/webphone

Remote (prod) tracked run:
  WEBPHONE_CAPTURE_EXTENSION=1001 WEBPHONE_CAPTURE_PASSWORD=... \
  nix shell nixpkgs#chromium nixpkgs#chromedriver \
    nixpkgs#python312.withPackages(ps: [ ps.selenium ps.pillow ]) \
    --command python3 scripts/ui-capture.py --base-url https://pbx.artmann.tech --track

Output: ui-shots/<n>-<name>-<light|dark>.png (up to 24 files); with
--track: ui-shots/<host>/<UTC-stamp>/ plus gallery.html, summary.json
and a host-level compare.html.
"""

from __future__ import annotations

import argparse
import base64
import glob
import html
import http.client
import json
import os
import re
import shutil
import socket
import ssl
import subprocess  # nosec B404 - deliberately spawns the local webphone binary under test
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
import webbrowser
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

# TLS verification context: None = verify (every real target). --insecure
# swaps in an unverified context for the LOCAL TLS-front smoke only.
_ssl_context: ssl.SSLContext | None = None

SEED_THREADS = [
    ("+441632960961", "Contract question", "Can you send the revised contract?"),
    ("+441632960962", "Delivery update", "The parcel arrives Thursday."),
]
# Seeded FIRST so the recency-ordered thread list keeps the two threads
# above at the top (their "thread" shot stays comparable) and the
# attachment thread lands LAST — resolvable as the list's final link.
SEED_ATTACHMENT_THREAD = ("+441632960963", "Voice memo", "Notes from the site visit.")
# A minimal honest RIFF/WAVE header (8 kHz mono 8-bit, empty data chunk):
# nothing parses it — the DECLARED audio/wav part type is what makes the
# attachment render as audio (attachmentMimeType trusts a set header).
SEED_WAV = (
    b"RIFF"
    + (36).to_bytes(4, "little")
    + b"WAVEfmt "
    + (16).to_bytes(4, "little")
    + b"\x01\x00\x01\x00"
    + (8000).to_bytes(4, "little")
    + (8000).to_bytes(4, "little")
    + b"\x01\x00\x08\x00data"
    + (0).to_bytes(4, "little")
)
SEED_TRANSCRIPT_TEXT = "Quarterly revenue figures are up twelve percent."
SEED_SNIPPETS = [
    ("Thanks, on it!", True),
    ("Please call back later", False),
]
SEED_CONTACTS = [("Marta Jensen", "+441632960961")]
TABS = [
    ("messages", "/messages"),
    ("thread", None),  # first thread, resolved after seeding
    ("crm-context", None),  # the same thread WITH the fake CRM on (--crm only)
    ("media-transcribe", None),  # the attachment thread, resolved after seeding
    ("fax", "/fax"),
    ("voicemail", "/voicemail"),
    ("history", "/history"),
    ("history-transcript", "/history?q=revenue"),  # seeded transcript search
    ("contacts", "/contacts"),
    ("settings", "/settings"),
    ("not-found", "/nope"),  # the styled 404: a real page, shot like any other
]
# Capture-time DOM assertion per surface: a screenshot is only evidence
# when the DOM beneath it is the real surface (this check would have
# caught the thread deep link rendering the messages list — the shots
# were byte-identical until the fix). A (selector, text) tuple also
# pins rendered copy inside the surface.
SURFACE_MARKERS = {
    "messages": ".wp-thread-row",  # seeded threads render rows
    "thread": "#wp-thread-head",  # the open conversation head
    # The two halves of the CRM integration, asserted in the DOM: the caller
    # name is a LINK into the CRM, and the context strip carries the CRM's
    # published facts. A tuple list keeps both pins on one surface.
    "crm-context": [
        (".wp-crm-link", "Ada Lovelace"),
        ("#wp-crm-context", "Analytical Engines Ltd"),
        ("#wp-crm-context", "Retainer 2027"),
    ],
    "media-transcribe": "[data-transcribe-src]",  # audio attachment button (ASR on)
    # The compose form exists only when the booted config declares outbound
    # fax; a deployment without it renders the capability notice instead.
    # Both are honest fax surfaces, so either satisfies the marker — the
    # harness must not demand a form the product deliberately withholds.
    "fax": ".wp-compose-fax, .wp-fax-off",
    "voicemail": "#voicemail-panel",
    "history": "#history-panel",
    "history-transcript": "[data-copy-transcript]",  # seeded transcript section
    "contacts": ".wp-export-link",  # unconditional in the panel
    "settings": ("#settings-panel dd .wp-service-on", "openai-compatible"),
    "not-found": (".wp-panel.wp-notfound", "404 — page not found"),
}

# F20 split Settings into My preferences / My data / a COLLAPSED server
# section. The server facts (the ASR provider row among them) are in the
# DOM but not visible until the details expands — Selenium's .text sees
# only visible text, so the copy assert needs the section opened first.
SURFACE_EXPAND = {
    "settings": "details.wp-settings-server > summary",
}

# Anonymous surfaces (2026-10-25): shot per theme BEFORE the session'd
# tabs, with the session cookie deleted — the login card only renders
# pre-session, and /login bounces a live session home (303). Render-only
# passkey config (see main()); no ceremony ever runs, so the rp_id never
# faces a browser — it only has to pass boot validation.
ANONYMOUS_SURFACES = [
    ("login-card", "/"),
    ("login-page", "/login"),
]
ANONYMOUS_MARKERS = {
    # The card's anchor text is the island i18n default (server-rendered,
    # swapped client-side); the library page's copy is the library's —
    # pin the WIRE id only, not prose we do not own.
    "login-card": ("#passkey-login", "Sign in with passkey"),
    "login-page": ("#lp-login-form", None),
    # The token-gated enroll page (conditional: shot only when the
    # deployment renders the form — see probe_enroll).
    "enroll": ("#enroll-form", None),
}
# Selectors that must NOT exist on a surface. Only --fax-off populates
# an entry (the fax surface must not carry the compose form while the
# boot declares send=false); the loop asserts absence after the pins.
SURFACE_ABSENT_MARKERS: dict[str, tuple[str, ...]] = {}

# A ringing inbound call mid-pass (the island banner, hidden when idle):
# the ring guard re-shoots the affected cell once instead of archiving
# a banner shot as a clean matrix cell.
RINGING_SELECTOR = "#incoming-call:not([hidden])"

# Remote-only conditional surfaces resolve markers at resolve time
# (thread-NN, crm-context on real data); the admin panel's wire id is
# unconditional markup of the admin surface itself.
SURFACE_MARKERS["admin"] = "#admin-panel"


FAKE_CRM_TOKEN = "ui-capture-crm-token"  # nosec B105 - LOCAL-ONLY visual harness; never a real credential
FAKE_CRM_CONTACT_ID = "01VISUALCONTACT0000000000"
# One contact the phone can describe: a company, a valuable deal with a
# counted tail, an outstanding task that is not yet due, and a recent touch.
FAKE_CRM_CONTEXT = {
    "id": FAKE_CRM_CONTACT_ID,
    "name": "Ada Lovelace",
    "company": "Analytical Engines Ltd",
    "open_deals": [
        {
            "title": "Retainer 2027",
            "stage": "proposal",
            "value_cents": 1200000,
            "currency": "EUR",
        }
    ],
    "open_deal_count": 3,
    "open_tasks": 2,
    "next_task": {"title": "Send the revised offer", "due_at": "2027-11-02T09:00:00Z"},
    "last_activity_at": "2026-10-24T11:03:00Z",
}
FAKE_CRM_LOOKUP = {
    "results": [
        {
            "id": FAKE_CRM_CONTACT_ID,
            "first_name": "Ada",
            "last_name": "Lovelace",
            "email": "ada@example.org",
        }
    ]
}


class _FakeCRMHandler(BaseHTTPRequestHandler):
    """The Kith machine API, reduced to the three routes the phone calls.

    Every answer is bearer-gated: a harness whose token is wrong must fail
    loudly here rather than quietly render the no-CRM product and pass.
    """

    def do_GET(self) -> None:
        if not self._authorized():
            return
        path = urllib.parse.urlsplit(self.path).path
        if path == "/api/contacts/by-phone":
            self._json(200, FAKE_CRM_LOOKUP)

            return
        if path == f"/api/contacts/{FAKE_CRM_CONTACT_ID}/context":
            self._json(200, FAKE_CRM_CONTEXT)

            return
        self._json(404, {"error": "no such route"})

    def do_POST(self) -> None:
        if not self._authorized():
            return
        if (
            urllib.parse.urlsplit(self.path).path
            == f"/api/contacts/{FAKE_CRM_CONTACT_ID}/calls"
        ):
            self.send_response(204)
            self.send_header("Content-Length", "0")
            self.end_headers()

            return
        self._json(404, {"error": "no such route"})

    def _authorized(self) -> bool:
        if self.headers.get("Authorization") == f"Bearer {FAKE_CRM_TOKEN}":
            return True
        self._json(401, {"error": "unauthorized"})

        return False

    def _json(self, status: int, payload: dict[str, object]) -> None:
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args: object) -> None:
        # Silence: the harness's stdout is the shot ledger, not access logs.
        return


def start_fake_crm() -> tuple[str, str]:
    """Boot the fake machine API on an ephemeral loopback port.

    Returns (base url, token). The server is a daemon thread, so it dies with
    the process; the caller needs no shutdown bookkeeping.
    """
    server = ThreadingHTTPServer(("127.0.0.1", 0), _FakeCRMHandler)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    host, port = server.server_address[0], server.server_address[1]

    return f"http://{host}:{port}", FAKE_CRM_TOKEN


def seed(base: str) -> tuple[str | None, str | None, str]:
    """Deterministic content over HTTP; returns (thread path, attachment
    thread path, session cookie)."""
    csrf = ""
    cookie_processor = urllib.request.HTTPCookieProcessor()
    opener = urllib.request.build_opener(cookie_processor)

    def post(
        path: str,
        fields: dict[str, str],
        multipart: bool = False,
        files: list[tuple[str, str, bytes, str]] | None = None,
    ) -> None:
        if multipart or files:
            # /messages/send rides the multipart prologue (the composer
            # uploads attachments; plain fields travel as bare parts).
            boundary = "wp-visual-seed"
            parts = [
                f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n'.encode()
                for k, v in fields.items()
            ]
            for field, filename, content, ctype in files or []:
                parts.append(
                    (
                        f"--{boundary}\r\nContent-Disposition: form-data; "
                        f'name="{field}"; filename="{filename}"\r\n'
                        f"Content-Type: {ctype}\r\n\r\n"
                    ).encode()
                    + content
                    + b"\r\n"
                )
            data = b"".join(parts) + f"--{boundary}--\r\n".encode()
            ctype_header = f"multipart/form-data; boundary={boundary}"
        else:
            data = urllib.parse.urlencode(fields).encode()
            ctype_header = "application/x-www-form-urlencoded"
        req = urllib.request.Request(base + path, data=data, method="POST")
        req.add_header("Content-Type", ctype_header)
        req.add_header("X-CSRF-Token", csrf)
        opener.open(req)

    # Session via the JSON API (the island's own flow): the anonymous
    # page ships a meta CSRF token, /api/session consumes it, and
    # GET /api/csrf adopts the rotated one (the session.js dance).
    import re

    page = opener.open(base + "/").read().decode()
    token = re.search(r'name="csrf-token" content="([^"]+)"', page)
    if token is None:
        raise RuntimeError("anonymous page carries no csrf meta token")
    csrf = token.group(1)
    body = json.dumps({"extension": "1001", "password": "pw"}).encode()  # nosec B105 - fixture credential for the local harness server
    req = urllib.request.Request(base + "/api/session", data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-CSRF-Token", csrf)
    opener.open(req)
    csrf = json.loads(opener.open(base + "/api/csrf").read()).get("token", "")
    # The attachment thread goes in FIRST (see SEED_ATTACHMENT_THREAD):
    # recency ordering keeps it at the list's tail.
    remote, subject, note = SEED_ATTACHMENT_THREAD
    post(
        "/messages/send",
        {"to": remote, "body": note},
        files=[("attachment", f"{subject}.wav", SEED_WAV, "audio/wav")],
    )
    for remote, _, message in SEED_THREADS:
        post("/messages/send", {"to": remote, "body": message}, multipart=True)
    for text, quick in SEED_SNIPPETS:
        post("/snippets/save", {"body": text, **({"quick": "1"} if quick else {})})
    for name, number in SEED_CONTACTS:
        post("/contacts/save", {"name": name, "number": number})
    # A transcript for the History search surface: the island's own
    # fire-and-forget report shape (POST /api/transcripts).
    payload = json.dumps(
        {
            "callId": "ui-capture-call",
            "direction": "out",
            "remote": "+441632960961",
            "startedAt": int(time.time() * 1000) - 3_600_000,
            "text": SEED_TRANSCRIPT_TEXT,
        }
    ).encode()
    req = urllib.request.Request(base + "/api/transcripts", data=payload, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-CSRF-Token", csrf)
    opener.open(req)
    # Thread paths: the list page's deep links, first (newest of the two
    # plain threads) and last (the attachment thread).
    html = opener.open(base + "/messages").read().decode()
    links = scrape_thread_links(html)

    def link_at(idx: int) -> str | None:
        # Skips past 'hx-get="/partials' so the slice keeps the deep-link
        # shape "/messages/<thread-id>" (a leading slash driver.get needs).
        return links[idx] if idx < len(links) else None

    thread_path = link_at(0) if links else None
    attach_thread_path = link_at(-1) if len(links) > 1 else None
    # The minted session cookie rides back to the caller: the browser
    # gets it INJECTED (driver.add_cookie) instead of replaying the
    # island's login — the island gates its panel on the SIP WebSocket,
    # which a bare boot cannot serve (caddy bridges /sip on the stack).
    session = ""
    for cookie in cookie_processor.cookiejar:
        if cookie.name == "webphone_session" and cookie.value:
            session = cookie.value
    return thread_path, attach_thread_path, session


def login_remote(base: str, extension: str, password: str) -> str:
    """Mint a session over POST /api/session (the island's own flow) and
    return the webphone_session cookie value; "" on any refusal.

    The credentials verify against the REAL PBX directory on a deployment,
    so a wrong pair is a hard, named stop — never a retry loop (the login
    lane is rate limited per IP and a loop would lock the operator out).
    """
    jar = urllib.request.HTTPCookieProcessor()
    handler = (
        urllib.request.HTTPSHandler(context=_ssl_context) if _ssl_context else None
    )
    opener = urllib.request.build_opener(*([handler] if handler else []), jar)
    try:
        page = opener.open(base + "/", timeout=30).read().decode("utf-8", "replace")
    except (OSError, http.client.HTTPException) as exc:
        print(f"login: cannot reach {base} ({exc})", file=sys.stderr)
        return ""
    token = re.search(r'name="csrf-token" content="([^"]+)"', page)
    if token is None:
        print(
            f"login: {base}/ carries no csrf meta token — not a webphone page?",
            file=sys.stderr,
        )
        return ""
    body = json.dumps({"extension": extension, "password": password}).encode()
    req = urllib.request.Request(base + "/api/session", data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-CSRF-Token", token.group(1))
    try:
        opener.open(req, timeout=30)
    except urllib.error.HTTPError as exc:
        if exc.code in (401, 403):
            print(
                f"login: refused ({exc.code}): WEBPHONE_CAPTURE_EXTENSION / "
                "WEBPHONE_CAPTURE_PASSWORD failed the directory verification",
                file=sys.stderr,
            )
        elif exc.code == 429:
            print(
                "login: rate limited (429) — wait a minute and retry",
                file=sys.stderr,
            )
        else:
            print(f"login: HTTP {exc.code} from /api/session", file=sys.stderr)
        return ""
    for cookie in jar.cookiejar:
        if cookie.name == "webphone_session" and cookie.value:
            return cookie.value
    print("login: no webphone_session cookie was minted", file=sys.stderr)
    return ""


def _get(base: str, path: str, session: str | None = None) -> str:
    req = urllib.request.Request(base + path)
    if session:
        req.add_header("Cookie", f"webphone_session={session}")
    with urllib.request.urlopen(req, timeout=30, context=_ssl_context) as resp:  # nosec B310 - read-only fetch of the server under capture
        return resp.read().decode("utf-8", "replace")


def resolve_remote_surfaces(base: str, session: str, threads_n: int) -> dict:
    """Build the surface list from LIVE data: the first `threads_n` thread
    deep links off the real list, plus the static tabs and the CONDITIONAL
    surfaces the deployment proves (CRM strip in a thread fragment, admin
    in the nav). Returns {"surfaces", "threads_found", "threads_requested"}
    and prints the resolved notes so the operator sees exactly what will be
    shot."""
    links: list[str] = []
    admin = False
    try:
        page = _get(base, "/messages", session)
        links = scrape_thread_links(page)
        admin = has_admin_nav(page)
    except (OSError, http.client.HTTPException) as exc:
        print(
            f"resolve: /messages unreadable ({exc}) — thread surfaces skipped",
            file=sys.stderr,
        )
    if admin:
        print(
            "admin: the session's nav carries the admin link — /admin joins the matrix"
        )
    if 0 < len(links) < threads_n:
        print(
            f"note: {len(links)} live thread(s) < --threads {threads_n} — shooting all of them"
        )
    surfaces: list[tuple[str, str | None]] = [("messages", "/messages")]
    chosen = links[:threads_n]
    for idx, link in enumerate(chosen, 1):
        surfaces.append((f"thread-{idx:02d}", link))
        # The open-conversation head is the same evidence pin the local
        # "thread" surface uses — a live thread page must show it; when real
        # data shifts beneath a tracked run, leniency degrades to a warning.
        SURFACE_MARKERS[f"thread-{idx:02d}"] = "#wp-thread-head"
    crm_link: str | None = None
    for idx, link in enumerate(chosen, 1):
        try:
            fragment = _get(base, "/partials" + link, session)
        except (OSError, http.client.HTTPException):
            continue
        if "data-transcribe-src" in fragment:
            print(f"media: thread-{idx:02d} carries transcribable audio ({link})")
        if crm_link is None and fragment_has_crm_strip(fragment):
            crm_link = link
    if crm_link:
        # One home for the pin: the real CRM strip's wire id, selector-only
        # (the local fixture's copy pins do not hold against real Kith data).
        print(
            f"crm: thread fragment {crm_link} carries #wp-crm-context — crm-context joins"
        )
        SURFACE_MARKERS["crm-context"] = "#wp-crm-context"
    surfaces += [("fax", "/fax"), ("voicemail", "/voicemail"), ("history", "/history")]
    try:
        history_html = _get(base, "/history", session)
        if "data-copy-transcript" in history_html:
            # A dense shot needs a q= that actually matches: scrape a word
            # out of the first transcript body instead of guessing a term.
            word = transcript_probe_word(history_html)
            path = f"/history?q={urllib.parse.quote(word)}" if word else "/history"
            surfaces.append(("history-transcript", path))
            if word:
                print(
                    f"history: dense search rides q={word!r} (scraped from a live transcript)"
                )
        else:
            print(
                "note: no call transcripts on this deployment — "
                "history-transcript skipped"
            )
    except (OSError, http.client.HTTPException):
        pass
    surfaces += [
        ("contacts", "/contacts"),
        ("settings", "/settings"),
    ]
    if crm_link:
        surfaces.append(("crm-context", crm_link))
    if admin:
        surfaces.append(("admin", "/admin"))
    surfaces += [("not-found", "/nope")]
    print("surfaces: " + ", ".join(name for name, _ in surfaces))
    return {
        "surfaces": surfaces,
        "threads_found": len(links),
        "threads_requested": threads_n,
    }


def fetch_version_info(base: str) -> dict | None:
    """The deployment's /version (open route) as {"version", "commit"};
    None when fenced or absent."""
    try:
        with urllib.request.urlopen(
            base + "/version", timeout=10, context=_ssl_context
        ) as resp:  # nosec B310 - read-only version probe
            payload = json.loads(resp.read())
    except (OSError, ValueError, http.client.HTTPException):
        return None
    info = {"version": str(payload.get("version") or "") or None}
    commit = str(payload.get("commit") or "").strip()
    info["commit"] = commit[:12] if commit else None
    return info if info["version"] or info["commit"] else None


def shoot(driver, target: str, full_page: bool) -> None:
    """Save a screenshot; full_page captures the ENTIRE document via CDP
    (captureBeyondViewport), not just the window — below-the-fold layout
    rot is exactly what this harness exists to expose."""
    if not full_page:
        driver.save_screenshot(target)
        return
    shot = driver.execute_cdp_cmd(
        "Page.captureScreenshot",
        {"format": "png", "captureBeyondViewport": True},
    )
    with open(target, "wb") as fh:
        fh.write(base64.b64decode(shot["data"]))


def shot_key(path: str) -> str:
    """`01-messages-light.png` -> `messages-light` — the STABLE compare key
    (numbering drifts between runs, the surface name does not)."""
    match = re.match(r"^\d+-(.+)\.png$", os.path.basename(path))
    return match.group(1) if match else os.path.splitext(os.path.basename(path))[0]


def pixel_diff(
    prev_path: str,
    cur_path: str,
    diff_path: str,
    threshold: int = 12,
    clip: int | None = None,
) -> tuple[float, bool] | None:
    """Heat-map the changed pixels into diff_path; returns (changed
    fraction, clipped) — or None without pillow / on an undiffable pair.
    With --diff-clip set, a height drift crops both shots to their common
    top-N region and diffs there (clipped=True, the summary marks it);
    without the bar a size mismatch stays a skip (side-by-side carries it)."""
    try:
        from PIL import Image, ImageChops, ImageEnhance
    except ImportError:
        return None
    prev = Image.open(prev_path).convert("RGB")
    cur = Image.open(cur_path).convert("RGB")
    box = clip_box(prev.size, cur.size, clip)
    if box is None:
        return None
    clipped = prev.size != cur.size
    if clipped:
        prev = prev.crop((0, 0, box[0], box[1]))
        cur = cur.crop((0, 0, box[0], box[1]))
    grey = ImageChops.difference(prev, cur).convert("L")
    hist = grey.histogram()
    total = box[0] * box[1]
    changed = sum(hist[threshold:]) / total
    mask = grey.point(lambda v: 255 if v >= threshold else 0)
    faded = ImageEnhance.Brightness(prev).enhance(0.35)
    heat = Image.composite(Image.new("RGB", prev.size, (235, 20, 60)), faded, mask)
    os.makedirs(os.path.dirname(diff_path), exist_ok=True)
    heat.save(diff_path)
    return changed, clipped


_REPORT_CSS = """<!doctype html><html><head><meta charset="utf-8">
<title>webphone shots</title><style>
body{font:14px/1.4 system-ui,sans-serif;margin:16px;background:#111;color:#eee}
h1{font-size:18px;margin:0 0 4px} p{color:#999;margin:4px 0 16px}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(340px,1fr));gap:16px}
figure{margin:0;background:#1c1c1c;padding:8px;border-radius:8px}
img{width:100%;display:block;border:1px solid #333}
figcaption{color:#bbb;padding-top:6px;font-size:12px}
section{margin-bottom:24px}
section h2{font-size:15px;margin:0 0 8px}
.row{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.badge{background:#c8143c;color:#fff;border-radius:4px;padding:1px 6px;font-size:12px;margin-left:6px}
.badge.new{background:#287f6c}
.size{color:#8899a6;font-weight:normal;font-size:12px;margin-left:6px}
.clipped{background:#8a6d1a;color:#fff;border-radius:4px;padding:1px 6px;font-size:12px;margin-left:6px}
.missing{color:#666;border:1px dashed #444;border-radius:6px;padding:24px;text-align:center}
</style></head><body>
"""


def build_gallery(
    run_dir: str,
    base: str,
    version: str | None,
    label: str | None,
    wire_note: str | None = None,
) -> None:
    parts = [
        f"<h1>{html.escape(base)}</h1><p>run {html.escape(os.path.basename(run_dir))}"
    ]
    if version:
        parts.append(f" · {html.escape(version)}")
    if label:
        parts.append(f" · {html.escape(label)}")
    if wire_note:
        parts.append(f" · {html.escape(wire_note)}")
    parts.append("</p>" + '<div class="grid">')
    for path in sorted(glob.glob(os.path.join(run_dir, "*.png"))):
        rel = os.path.basename(path)
        key = shot_key(path)
        parts.append(
            f'<figure><a href="{html.escape(rel)}"><img src="{html.escape(rel)}" '
            f'loading="lazy" alt="{html.escape(key)}"></a>'
            f"<figcaption>{html.escape(key)}</figcaption></figure>"
        )
    parts.append("</div></body></html>")
    with open(os.path.join(run_dir, "gallery.html"), "w", encoding="utf-8") as fh:
        fh.write("".join(parts))


def build_compare(
    host_dir: str,
    prev_dir: str,
    cur_dir: str,
    version: str | None,
    label: str | None,
    changed: dict[str, float],
    badge_bar: float = 0.5,
    clipped_keys: set[str] | None = None,
) -> None:
    """One compare.html per host: the newest run against the previous one,
    per surface — previous | current | diff, worst offenders first, with
    byte-size deltas and links into both run galleries."""

    def _collect(run: str) -> dict[str, str]:
        return {
            shot_key(p): os.path.basename(p)
            for p in glob.glob(os.path.join(run, "*.png"))
        }

    prev_files, cur_files = _collect(prev_dir), _collect(cur_dir)
    clipped_keys = clipped_keys or set()
    prev_version = None
    try:
        with open(os.path.join(prev_dir, "summary.json"), encoding="utf-8") as fh:
            prev_version = json.load(fh).get("version")
    except (OSError, ValueError):
        pass

    def _img(run: str, fname: str | None, caption: str) -> str:
        if not fname:
            return f'<div class="missing">no {html.escape(caption)}</div>'
        src = html.escape(os.path.relpath(os.path.join(run, fname), host_dir))
        return (
            f'<figure><img src="{src}" loading="lazy" alt="{html.escape(caption)}">'
            f"<figcaption>{html.escape(caption)}</figcaption></figure>"
        )

    def _badge(key: str, clipped: bool = False) -> str:
        mark = '<span class="clipped">clipped</span>' if clipped else ""
        if key not in prev_files:
            return f'<span class="badge new">new</span>{mark}'
        pct = changed.get(key)
        if pct is not None and pct >= badge_bar:
            return f'<span class="badge">{pct:.2f}%</span>{mark}'
        return mark

    def _sizes(key: str) -> str:
        prev_f, cur_f = prev_files.get(key), cur_files.get(key)
        if not prev_f or not cur_f:
            return ""
        prev_n = os.path.getsize(os.path.join(prev_dir, prev_f))
        cur_n = os.path.getsize(os.path.join(cur_dir, cur_f))
        delta = (cur_n - prev_n) / prev_n * 100 if prev_n else 0
        return (
            f'<span class="size">{_bytes_text(prev_n)} → {_bytes_text(cur_n)} '
            f"({delta:+.1f}%)</span>"
        )

    keys = sorted(
        set(prev_files) | set(cur_files),
        key=lambda k: (-changed.get(k, 0.0), k),
    )
    head = (
        f"<h1>UI diff — {html.escape(os.path.basename(prev_dir))} → "
        f"{html.escape(os.path.basename(cur_dir))}</h1><p>"
    )
    versions = " · ".join(v for v in (prev_version, version) if v)
    if versions:
        head += html.escape(versions) + " · "
    if label:
        head += html.escape(label) + " · "
    head += html.escape(os.path.basename(host_dir)) + "</p>"
    head += (
        f'<p><a href="{html.escape(os.path.relpath(prev_dir, host_dir))}/gallery.html">'
        f'previous gallery</a> · <a href="{html.escape(os.path.relpath(cur_dir, host_dir))}'
        f'/gallery.html">current gallery</a></p>'
    )
    sections = []
    for key in keys:
        diff_rel = f"diffs/{key}.png" if key in changed else None
        sections.append(
            f"<section><h2>{html.escape(key)}{_badge(key, key in clipped_keys)}{_sizes(key)}</h2>"
            "<div class='row'>"
            + _img(prev_dir, prev_files.get(key), "previous")
            + _img(cur_dir, cur_files.get(key), "current")
            + _img(cur_dir, diff_rel, "diff")
            + "</div></section>"
        )
    with open(os.path.join(host_dir, "compare.html"), "w", encoding="utf-8") as fh:
        fh.write(_REPORT_CSS + head + "".join(sections) + "</body></html>")


def track(
    run_dir: str,
    base: str,
    facts: dict | None,
    label: str | None,
    full_page: bool,
    viewports: list[tuple[int, int]],
    ledger: list[dict[str, str]] | None = None,
    wire: dict[str, str | None] | None = None,
    threads: dict | None = None,
    diff_threshold: int = 12,
    badge_bar: float = 0.5,
    diff_clip: int | None = None,
    prune_keep: int | None = None,
) -> str | None:
    """History + diff bookkeeping for --track runs: summary.json, the
    per-run gallery, and the host-level compare against the previous run.
    Version facts are fetched by main() while the server is still up (the
    local boot is already torn down by the time this runs). Returns the
    produced HTML path (compare when a previous run exists, else the
    gallery) so --open can raise it."""
    version = (facts or {}).get("version")
    host_dir = os.path.dirname(run_dir)
    runs = _run_dirs(host_dir)
    run_id = os.path.basename(run_dir)
    earlier = [d for d in runs if d < run_id]
    prev_name = earlier[-1] if earlier else None
    prev_dir = os.path.join(host_dir, prev_name) if prev_name else None
    prev_by_key: dict[str, str] = {}
    if prev_dir:
        for path in glob.glob(os.path.join(prev_dir, "*.png")):
            prev_by_key[shot_key(path)] = path
    urls = {entry["key"]: entry["url"] for entry in ledger or []}
    shots = []
    diffed: list[tuple[str, float]] = []
    clipped_keys: set[str] = set()
    # The layout-proof artifact (zz-thread-collapsed) is evidence for the
    # proof, not a matrix cell — it must not enter the diff ledger.
    matrix = [
        p
        for p in sorted(glob.glob(os.path.join(run_dir, "*.png")))
        if not os.path.basename(p).startswith("zz-")
    ]
    for path in matrix:
        key = shot_key(path)
        entry: dict[str, object] = {
            "key": key,
            "file": os.path.basename(path),
            "url": urls.get(key),
            "bytes": os.path.getsize(path),
            "prev_file": None,
            "changed_pct": None,
            "diff_file": None,
            "clipped": False,
        }
        prev_path = prev_by_key.get(key)
        if prev_path:
            entry["prev_file"] = os.path.relpath(prev_path, run_dir)
            diff_path = os.path.join(run_dir, "diffs", f"{key}.png")
            result = pixel_diff(
                prev_path, path, diff_path, threshold=diff_threshold, clip=diff_clip
            )
            if result is None:
                entry["size_match"] = False
                print(
                    f"diff: {key} — size mismatch beyond the clip bar; side-by-side only"
                )
            else:
                changed_pct, clipped = result
                entry["changed_pct"] = round(changed_pct * 100, 3)
                entry["diff_file"] = os.path.relpath(diff_path, run_dir)
                entry["size_match"] = True
                entry["clipped"] = clipped
                if clipped:
                    clipped_keys.add(key)
                    print(
                        f"diff: {key} — height drifted; diffed the common top {diff_clip}px (clipped)"
                    )
                diffed.append((key, changed_pct * 100))
        shots.append(entry)
    if wire is not None:
        with open(
            os.path.join(run_dir, "wire-facts.json"), "w", encoding="utf-8"
        ) as fh:
            json.dump(wire, fh, indent=2, sort_keys=True)
            fh.write("\n")
    summary = {
        "run": run_id,
        "base": base,
        "version": version,
        "commit": (facts or {}).get("commit"),
        "label": label,
        "full_page": full_page,
        "viewports": [f"{w}x{h}" for w, h in viewports],
        "prev_run": prev_name,
        "threads_requested": (threads or {}).get("threads_requested"),
        "threads_found": (threads or {}).get("threads_found"),
        "wire_facts": {
            "config_js": wire is not None and wire.get("config_js") is not None,
            "settings_server_html": wire is not None
            and wire.get("settings_server_html") is not None,
        },
        "shots": shots,
    }
    with open(os.path.join(run_dir, "summary.json"), "w", encoding="utf-8") as fh:
        json.dump(summary, fh, indent=2, sort_keys=True)
        fh.write("\n")
    wire_note = None
    if wire is not None:
        bits = ["config.js ✓" if wire.get("config_js") else "config.js —"]
        bits.append(
            "settings server section ✓"
            if wire.get("settings_server_html")
            else "settings server section —"
        )
        wire_note = "wire: " + ", ".join(bits)
    build_gallery(run_dir, base, version, label, wire_note)
    print(f"gallery: {os.path.join(run_dir, 'gallery.html')}")
    produced = os.path.join(run_dir, "gallery.html")
    if prev_dir:
        build_compare(
            host_dir,
            prev_dir,
            run_dir,
            version,
            label,
            dict(diffed),
            badge_bar=badge_bar,
            clipped_keys=clipped_keys,
        )
        print(f"compare: {os.path.join(host_dir, 'compare.html')} (vs {prev_name})")
        produced = os.path.join(host_dir, "compare.html")
    else:
        print("compare: no previous run — first run in this history")
    diffed.sort(key=lambda pair: pair[1], reverse=True)
    notable = [(key, pct) for key, pct in diffed if pct >= badge_bar]
    print(f"track: {len(shots)} shot(s), {len(diffed)} differ from {prev_name or '—'}")
    for key, pct in notable[:10]:
        print(f"  changed {pct:5.2f}%  {key}")
    if diffed and not notable:
        print(f"  nothing above the {badge_bar:g}% noise bar")
    print(f"history: {len(runs)} run(s) under {host_dir}")
    if prune_keep is not None:
        removed = prune_runs(host_dir, prune_keep)
        if removed:
            print(f"prune: removed {len(removed)} old run(s): {', '.join(removed)}")
        else:
            print(f"prune: nothing to remove (keep {prune_keep})")
    return produced


def _parse_viewport(text: str) -> tuple[int, int]:
    width, _, height = text.lower().partition("x")
    if not width.isdigit() or not height.isdigit():
        raise SystemExit(
            f"bad --viewport {text!r} (want WxH, e.g. 1280x900 or 390x844)"
        )
    return int(width), int(height)


def _sanitize_dirname(base: str) -> str:
    return re.sub(
        r"[^A-Za-z0-9._-]+", "_", urllib.parse.urlsplit(base).netloc or "local"
    )


def scrape_thread_links(page: str) -> list[str]:
    """The /messages deep links off a rendered list page, in list order,
    deduped — the ONE link-scrape home (seed() and the remote resolver
    share it)."""
    marker = 'hx-get="/partials/messages/'
    links: list[str] = []
    for match in re.finditer(re.escape(marker), page):
        start = match.start() + len('hx-get="/partials')
        link = page[start : page.find('"', start)]
        if link and link not in links:
            links.append(link)
    return links


def normalize_pins(pins) -> list[tuple[str, str | None]]:
    """A marker is a selector, a (selector, copy) tuple, or a LIST of
    those pins — one normalized shape for the pin loop."""
    if isinstance(pins, str):
        return [(pins, None)]
    if isinstance(pins, tuple):
        return [pins]
    return list(pins)


def clip_box(
    size_a: tuple[int, int], size_b: tuple[int, int], clip: int | None
) -> tuple[int, int] | None:
    """The common region two shots can be diffed on, or None when
    unclip-able: equal sizes need no clip; any drift (height — full-page
    content growth, or width — scrollbar/rounding drift) only diffs when
    --diff-clip gave a bar, cropping BOTH shots to their common top-left
    region (the clipped mark in the compare says so)."""
    if size_a == size_b:
        return size_a
    if not clip or clip <= 0:
        return None
    return (min(size_a[0], size_b[0]), min(size_a[1], size_b[1], clip))


def has_admin_nav(page: str) -> bool:
    return 'href="/admin"' in page


def fragment_has_crm_strip(fragment: str) -> bool:
    return 'id="wp-crm-context"' in fragment


def transcript_probe_word(page: str) -> str | None:
    """A word scraped from the first transcript group body, for a
    ?q= search that will actually match real data (the seeded q=revenue
    only exists on the local fixture)."""
    block = re.search(r'<details class="wp-transcript-call".*?</details>', page, re.DOTALL)
    if not block:
        return None
    text = re.sub(r"<[^>]+>", " ", block.group(0))
    for word in re.findall(r"[A-Za-z]{6,}", text):
        return word
    return None


def probe_enroll(base: str) -> bool:
    """True when the deployment renders the token-gated enroll form —
    probed directly (the page is reached via distributed token links,
    never linked from the login surface)."""
    try:
        return 'id="enroll-form"' in _get(base, "/enroll")
    except (OSError, http.client.HTTPException, urllib.error.HTTPError):
        return False


def _run_dirs(host_dir: str) -> list[str]:
    return sorted(
        d
        for d in os.listdir(host_dir)
        if os.path.isdir(os.path.join(host_dir, d)) and re.match(r"^\d{8}T\d{6}Z$", d)
    )


def prune_runs(host_dir: str, keep: int) -> list[str]:
    """Newest-N retention over one host's run history; returns the removed
    run names. A keep < 1 refuses (a retention foot-gun is not a feature)."""
    if keep < 1:
        return []
    runs = _run_dirs(host_dir)
    doomed = runs[:-keep] if len(runs) > keep else []
    for name in doomed:
        shutil.rmtree(os.path.join(host_dir, name))
    return doomed


def snapshot_wire_facts(base: str, session: str | None) -> dict[str, str | None]:
    """Verbatim wire facts for the run dir: /config.js (the island's whole
    config surface) and the Settings server section (the service rows as
    rendered). Raw text, not an interpretation — forensics, not claims."""
    facts: dict[str, str | None] = {"config_js": None, "settings_server_html": None}
    try:
        facts["config_js"] = _get(base, "/config.js")
    except (OSError, http.client.HTTPException, urllib.error.HTTPError):
        pass
    try:
        page = _get(base, "/settings", session)
        match = re.search(
            r'<details class="wp-settings-server">.*?</details>', page, re.DOTALL
        )
        facts["settings_server_html"] = match.group(0) if match else None
    except (OSError, http.client.HTTPException, urllib.error.HTTPError):
        pass
    return facts


def _bytes_text(n: int) -> str:
    if n < 1024:
        return f"{n} B"
    if n < 1024 * 1024:
        return f"{n / 1024:.1f} KB"
    return f"{n / (1024 * 1024):.2f} MB"


def _apply_host_maps(mappings: list[str] | None) -> None:
    """Point chrome AND the python HTTP lane at the same addresses: the
    --host-map pairs patch socket.getaddrinfo (urllib resolves through
    it; chrome gets --host-resolver-rules in capture())."""
    mapped = {
        host: ip
        for host, sep, ip in (m.partition("=") for m in mappings or [])
        if host and ip
    }
    if not mapped:
        return
    original = socket.getaddrinfo

    def patched(host, *args, **kwargs):
        return original(mapped.get(host, host), *args, **kwargs)

    socket.getaddrinfo = patched  # type: ignore[assignment]


def capture(
    base: str,
    out_dir: str,
    session: str,
    surfaces: list[tuple[str, str | None]],
    only: set[str] | None = None,
    viewports: list[tuple[int, int]] | None = None,
    full_page: bool = True,
    lenient: bool = False,
    layout_proof: bool = False,
    anonymous_surfaces: list[tuple[str, str]] | None = None,
    insecure: bool = False,
    host_maps: list[str] | None = None,
) -> list[dict[str, str]]:
    """Drive headless chromium through every (surface, theme, viewport)
    cell and screenshot it; returns the shot LEDGER (key, file, url per
    shot — the count is len()).

    lenient=True (remote --base-url runs): a missing marker is a printed
    warning and a KEPT shot, and a styled-404 page skips its surface —
    real data legitimately lacks the seeded fixtures the strict pins
    assume, and a capability a deployment withheld is honest UI, not a
    broken run.
    """
    # Selenium lives only in the capture env (see the AGENTS run recipe
    # and mypy.ini) — never in the lint shell.
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.support.ui import WebDriverWait

    viewports = viewports or [(1280, 900)]
    options = Options()
    # The NIX chromium/chromedriver (Selenium Manager's downloaded driver
    # is a generic binary that cannot exec on nix). Both come from the
    # documented nix shell; fall back to PATH discovery.
    chrome = shutil.which("chromium")
    if chrome:
        options.binary_location = chrome
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--use-fake-device-for-media-stream")
    options.add_argument("--use-fake-ui-for-media-stream")
    options.add_argument("--window-size=1280,900")
    if insecure:
        # The LOCAL TLS-front smoke only (self-signed cert): the W3C
        # capability is what chromedriver honors — the switch form is
        # silently ignored.
        options.set_capability("acceptInsecureCerts", True)
    for mapping in host_maps or []:
        host, sep, ip = mapping.partition("=")
        if host and ip:
            options.add_argument(f"--host-resolver-rules=MAP {host} {ip}")
    driver_path = shutil.which("chromedriver")
    from selenium.webdriver.chrome.service import Service

    driver_path = shutil.which("chromedriver")

    service = Service(executable_path=driver_path) if driver_path else None
    driver = webdriver.Chrome(service=service, options=options)
    shots = 0
    ledger: list[dict[str, str]] = []
    anonymous = (
        anonymous_surfaces if anonymous_surfaces is not None else ANONYMOUS_SURFACES
    )

    def _target(name: str, theme: str, vp: tuple[int, int]) -> str:
        suffix = "" if len(viewports) == 1 else f"-{vp[0]}x{vp[1]}"
        return os.path.join(out_dir, f"{shots + 1:02d}-{name}-{theme}{suffix}.png")

    def _marker_fail(surface: str, path: str, marker: str, text: str | None) -> None:
        detail = (
            f"surface {surface!r} at {path} is missing {marker!r}"
            if text is None
            else (
                f"surface {surface!r} rendered without {text!r} "
                "(the marker matched but the copy is wrong)"
            )
        )
        if lenient:
            print(f"warn: {detail} — shot kept, marker unverified")
            return
        raise AssertionError(
            f"{detail} — url={driver.current_url} — the shot would not be "
            "evidence; aborting"
        )

    def _check_pins(name: str, path: str, pins) -> None:
        # Strict mode raises inside _marker_fail; lenient mode warns and
        # keeps going.
        for marker, expected_text in normalize_pins(pins):
            if not driver.find_elements("css selector", marker):
                _marker_fail(name, path, marker, None)
                return
            if (
                expected_text
                and expected_text
                not in driver.find_element("css selector", "body").text
            ):
                _marker_fail(name, path, marker, expected_text)
                return

    def _is_styled_404() -> bool:
        return bool(driver.find_elements("css selector", ".wp-panel.wp-notfound"))

    try:
        for vp in viewports:
            driver.set_window_size(*vp)
            driver.get(base + "/")
            WebDriverWait(driver, 10).until(
                lambda d: d.execute_script("return document.readyState") == "complete"
            )
            # The session cookie, injected (see seed()/login_remote()): the
            # server-rendered tabs ride it directly. Anonymous runs inject
            # nothing (selenium rejects an empty cookie value).
            if session:
                driver.add_cookie({"name": "webphone_session", "value": session})
            for theme in ("light", "dark"):
                # Theme pinned via the persisted key (theme-preload.js applies
                # it pre-paint — no cycling the 3-state toggle from "auto").
                driver.get(base + "/")
                driver.execute_script(f"localStorage.setItem('wp-theme', '{theme}')")
                # The library /login page follows prefers-color-scheme, not
                # wp-theme (the documented forced-theme caveat) — emulate the
                # media feature so the anonymous shots see exactly what a
                # same-scheme OS user sees. Shell surfaces stay pinned via
                # wp-theme, so this is a no-op for the tabs.
                driver.execute_cdp_cmd(
                    "Emulation.setEmulatedMedia",
                    {"features": [{"name": "prefers-color-scheme", "value": theme}]},
                )
                # Anonymous surfaces first: the login card renders pre-session
                # and /login bounces a live session home (303). Drop the
                # cookie for these shots, restore it after.
                driver.delete_cookie("webphone_session")
                for name, path in anonymous:
                    if only is not None and name not in only:
                        continue
                    driver.get(base + path)
                    if name == "login-card":
                        # The island reveals the card once the anonymous
                        # resume probe fails — wait on the sign-in affordance
                        # itself (the passkey link when the mode is on, the
                        # classic form otherwise) instead of a fixed sleep.
                        WebDriverWait(driver, 8).until(
                            lambda d: (
                                d.find_elements("css selector", "#passkey-login")
                                or d.find_elements("css selector", "#login-form")
                            )
                        )
                    time.sleep(0.4)  # settle: theme script, fonts
                    if lenient and name != "not-found" and _is_styled_404():
                        # The deliberate /nope surface IS the styled 404 and
                        # must still be shot; this skip is for capabilities a
                        # deployment withheld (e.g. /login without passkey).
                        print(
                            f"skip: {name} at {path} is the styled 404 "
                            "(capability off on this deployment)"
                        )
                        continue
                    _check_pins(name, path, ANONYMOUS_MARKERS[name])
                    target = _target(name, theme, vp)
                    shoot(driver, target, full_page)
                    if driver.find_elements("css selector", RINGING_SELECTOR):
                        print(
                            f"warn: {name}: an inbound call is ringing — re-shooting "
                            "once after a settle wait"
                        )
                        time.sleep(3)
                        shoot(driver, target, full_page)
                    ledger.append(
                        {
                            "key": shot_key(target),
                            "file": os.path.basename(target),
                            "url": driver.current_url,
                        }
                    )
                    print(f"captured {target}")
                    shots += 1
                if session:
                    driver.add_cookie({"name": "webphone_session", "value": session})
                for name, path in surfaces:
                    if only is not None and name not in only:
                        continue
                    if path is None:
                        # An unresolved deep link (seed came up empty / the
                        # deployment lacks the capability) skips the surface
                        # rather than shooting an error page.
                        continue
                    driver.get(base + path)
                    # The island reveals its call view only after the resume
                    # probe settles (login view stays hidden meanwhile); the
                    # SIP/WebSocket latency varies by pass, so wait on the
                    # view itself instead of a fixed sleep.
                    WebDriverWait(driver, 8).until(
                        lambda d: (
                            not d.find_element(
                                "css selector", "#phone-view"
                            ).get_attribute("hidden")
                        )
                    )
                    time.sleep(0.4)  # settle: relative times, panels
                    if name == "history-transcript":
                        # Transcript groups ship as collapsed <details>; open
                        # them so the shot carries the text, not just rows.
                        driver.execute_script(
                            "document.querySelectorAll('details.wp-transcript-call')"
                            ".forEach((d) => d.setAttribute('open', ''))"
                        )
                    if lenient and name != "not-found" and _is_styled_404():
                        print(
                            f"skip: {name} at {path} is the styled 404 "
                            "(capability off on this deployment)"
                        )
                        continue
                    for expand_sel in (
                        [SURFACE_EXPAND[name]] if name in SURFACE_EXPAND else []
                    ):
                        expandable = driver.find_elements("css selector", expand_sel)
                        if expandable and expandable[0].is_displayed():
                            # Narrow viewports park collapsed sections under
                            # the fixed bottom tab bar — scroll out first, and
                            # fall back to a JS click when something still
                            # overlays the summary.
                            driver.execute_script(
                                "arguments[0].scrollIntoView({block: 'center'});",
                                expandable[0],
                            )
                            try:
                                expandable[0].click()
                            except Exception:
                                driver.execute_script(
                                    "arguments[0].click();", expandable[0]
                                )
                    _check_pins(name, path, SURFACE_MARKERS.get(name, []))
                    for absent in SURFACE_ABSENT_MARKERS.get(name, ()):
                        if driver.find_elements("css selector", absent):
                            message = (
                                f"surface {name!r} renders {absent!r} but the boot "
                                "declared it away — the shot would prove a lie"
                            )
                            if lenient:
                                print(f"warn: {message}")
                            else:
                                raise AssertionError(message + "; aborting")
                    target = _target(name, theme, vp)
                    shoot(driver, target, full_page)
                    if driver.find_elements("css selector", RINGING_SELECTOR):
                        print(
                            f"warn: {name}: an inbound call is ringing — re-shooting "
                            "once after a settle wait"
                        )
                        time.sleep(3)
                        shoot(driver, target, full_page)
                    ledger.append(
                        {
                            "key": shot_key(target),
                            "file": os.path.basename(target),
                            "url": driver.current_url,
                        }
                    )
                    print(f"captured {target}")
                    shots += 1
                    if (
                        name == "thread"
                        and theme == "light"
                        and layout_proof
                        and vp == (1280, 900)
                    ):
                        # Layout facts belong here, not in a second script: a
                        # screenshot proves a surface RENDERS, not that the
                        # collapse reclaims width or that the master-detail
                        # split flips at the container threshold. Measured as
                        # COMPUTED STYLE, once per pass, on the surface that
                        # carries both panes (2026-10-29 closeout; a standalone
                        # proof script duplicated this config and hung for
                        # three hours). Remote runs skip it: a tracking pass
                        # must not fail on layout claims, and it mutates the
                        # collapsed preference mid-run.
                        check_thread_layout(driver, base, path, out_dir)
    finally:
        driver.quit()
    return ledger


def check_thread_layout(driver, base: str, thread_path: str, out_dir: str) -> None:
    """Assert the collapsible island and the Messages master-detail.

    Every claim asserted here is one a screenshot cannot make: which grid
    direction is live, whether the column is really removed, whether the
    preference survives a navigation, and where the container query flips.
    """
    # Imported here, not at module scope: selenium lives only in the capture
    # env (see the AGENTS run recipe), and this module is imported by nothing
    # else — but keeping the import local mirrors capture()'s own arrangement.
    from selenium.webdriver.support.ui import WebDriverWait

    probe_js = """
    const col = document.querySelector('.wp-main-col');
    const split = document.querySelector('.wp-msg-split');
    const rail = document.querySelector('.wp-msg-rail');
    const island = document.querySelector('.wp-island');
    const toggle = document.getElementById('wp-island-toggle');
    const nameEl = document.querySelector('.wp-rail-name');
    const row = document.querySelector('.wp-rail-row');
    return {
      collapsed: document.documentElement.classList.contains('wp-island-collapsed'),
      colWidth: Math.round(col.getBoundingClientRect().width),
      splitDir: split ? getComputedStyle(split).flexDirection : null,
      railDisplay: rail ? getComputedStyle(rail).display : null,
      islandDisplay: island ? getComputedStyle(island).display : null,
      ariaExpanded: toggle ? toggle.getAttribute('aria-expanded') : null,
      toggleLabel: toggle ? toggle.textContent.trim() : null,
      rowHeight: row ? Math.round(row.getBoundingClientRect().height) : null,
      nameTruncated: nameEl ? nameEl.scrollWidth > nameEl.clientWidth + 1 : null,
      nameText: nameEl ? nameEl.textContent.trim() : null,
    };
    """
    failures: list[str] = []

    expanded = driver.execute_script(probe_js)
    print(f"layout: expanded {expanded}")
    if expanded["splitDir"] != "row":
        failures.append(f"the split must be a row, got {expanded['splitDir']!r}")
    if expanded["islandDisplay"] == "none":
        failures.append("the island column must be visible while expanded")
    if expanded["ariaExpanded"] != "true":
        failures.append(
            f"aria-expanded must be 'true', got {expanded['ariaExpanded']!r}"
        )
    if not (expanded["toggleLabel"] or "").strip():
        failures.append("the collapse control has no visible label (undiscoverable)")
    if expanded["nameTruncated"]:
        failures.append(
            f"the rail still ellipsizes the name ({expanded['nameText']!r})"
        )
    if expanded["rowHeight"] is not None and expanded["rowHeight"] > 90:
        failures.append(f"the rail row grew to {expanded['rowHeight']}px")

    driver.find_element("id", "wp-island-toggle").click()
    time.sleep(0.6)
    collapsed = driver.execute_script(probe_js)
    print(f"layout: collapsed {collapsed}")
    if collapsed["islandDisplay"] != "none":
        failures.append(
            f"the collapsed island must be display:none, got {collapsed['islandDisplay']!r}"
        )
    if collapsed["ariaExpanded"] != "false":
        failures.append(
            f"aria-expanded must be 'false', got {collapsed['ariaExpanded']!r}"
        )
    if collapsed["splitDir"] != "row":
        failures.append(
            f"the split must survive the collapse, got {collapsed['splitDir']!r}"
        )
    if collapsed["colWidth"] <= expanded["colWidth"] + 300:
        failures.append(
            "the collapsed column must reclaim the island width "
            f"({expanded['colWidth']} -> {collapsed['colWidth']})"
        )
    if (
        driver.execute_script("return localStorage.getItem('wp-island-collapsed')")
        != "1"
    ):
        failures.append("the collapse preference was not persisted")
    driver.save_screenshot(os.path.join(out_dir, "zz-thread-collapsed.png"))

    # Reload: the pre-paint class must survive navigation.
    driver.get(base + thread_path)
    WebDriverWait(driver, 12).until(
        lambda d: d.find_element("css selector", "#wp-thread-head").is_displayed()
    )
    time.sleep(0.5)
    after = driver.execute_script(probe_js)
    print(f"layout: after reload {after}")
    if after["islandDisplay"] != "none" or after["ariaExpanded"] != "false":
        failures.append(
            "the stored collapse did not survive a reload "
            f"(display={after['islandDisplay']!r}, aria={after['ariaExpanded']!r})"
        )

    # Boundary widths around the 44rem (704px) container threshold: prove the
    # query flips exactly there and never leaves a dead band or an oscillation.
    for width in (640, 680, 700, 704, 720, 780):
        driver.set_window_size(width, 900)
        time.sleep(0.45)
        direction, col_width = driver.execute_script(
            "const s=document.querySelector('.wp-msg-split');"
            "const c=document.querySelector('.wp-main-col');"
            "return [s?getComputedStyle(s).flexDirection:null,"
            "Math.round(c.getBoundingClientRect().width)];"
        )
        want = "row" if col_width >= 704 else "column"
        print(
            f"layout: viewport {width} -> column {col_width} -> {direction} (want {want})"
        )
        if direction != want:
            failures.append(
                f"at viewport {width} the column is {col_width}px so the split "
                f"must be {want}, got {direction!r}"
            )
    driver.set_window_size(1280, 900)

    # Leave the workspace expanded: the collapse is a preference, and a proof
    # run must not change what the next pass starts from.
    driver.find_element("id", "wp-island-toggle").click()
    time.sleep(0.3)
    if driver.execute_script(
        "return document.documentElement.classList.contains('wp-island-collapsed')"
    ):
        failures.append("the island did not expand again after the proof")

    if failures:
        raise AssertionError("layout proof failed:\n  - " + "\n  - ".join(failures))
    print("layout: PASS (collapse, persistence, split threshold)")


def discard_empty_run(out_dir: str) -> None:
    """A crashed capture must not leave an EMPTY run dir in the history —
    the next run's compare would silently treat the empty dir as its
    "previous" and badge every surface "new" (observed live, 2026-11-10).
    Only an empty dir is removed; a partial one stays (partial runs are
    compared gracefully)."""
    if not glob.glob(os.path.join(out_dir, "*.png")):
        shutil.rmtree(out_dir, ignore_errors=True)
        print(f"run discarded (no shots captured): {out_dir}")


def self_test() -> int:
    """Verify every pure helper this harness depends on — NO browser, NO
    server. Pillow pixel math runs only when the capture env carries it;
    everything else must pass everywhere. Non-zero exit on any failure."""
    failures: list[str] = []

    def check(name: str, ok: bool, detail: str = "") -> None:
        if ok:
            print(f"  ok   {name}")
        else:
            failures.append(name)
            print(f"  FAIL {name}: {detail}")

    check(
        "shot_key strips numbering",
        shot_key("07-thread-01-dark.png") == "thread-01-dark",
    )
    check(
        "shot_key unnumbered passthrough",
        shot_key("zz-thread-collapsed.png") == "zz-thread-collapsed",
    )
    check("viewport parse", _parse_viewport("390x844") == (390, 844))
    check("viewport parse caps", _parse_viewport("1280X900") == (1280, 900))
    try:
        _parse_viewport("wide")
        check("viewport junk rejected", False, "no SystemExit")
    except SystemExit:
        check("viewport junk rejected", True)
    check(
        "sanitize host",
        _sanitize_dirname("https://pbx.artmann.tech") == "pbx.artmann.tech",
    )
    check(
        "sanitize local",
        _sanitize_dirname("http://127.0.0.1:18097") == "127.0.0.1_18097",
    )

    fixture = (
        '<a hx-get="/partials/messages/t1">x</a>'
        '<a hx-get="/partials/messages/t2">y</a>'
        '<a hx-get="/partials/messages/t1">dup</a>'
    )
    check(
        "thread scrape order + dedup",
        scrape_thread_links(fixture) == ["/messages/t1", "/messages/t2"],
        repr(scrape_thread_links(fixture)),
    )
    check("pins: str", normalize_pins(".a") == [(".a", None)])
    check("pins: tuple", normalize_pins((".a", "txt")) == [(".a", "txt")])
    check(
        "pins: list",
        normalize_pins([(".a", None), (".b", "x")]) == [(".a", None), (".b", "x")],
    )

    check("clip: equal sizes", clip_box((800, 900), (800, 900), None) == (800, 900))
    check("clip: no bar refuses drift", clip_box((800, 900), (800, 1400), None) is None)
    check(
        "clip: bar bounds height", clip_box((800, 900), (800, 1400), 600) == (800, 600)
    )
    check(
        "clip: width drift crops too",
        clip_box((800, 900), (815, 900), 600) == (800, 600),
    )

    check("admin nav detected", has_admin_nav('<a href="/admin">Admin</a>'))
    check("admin nav exact", not has_admin_nav('<a href="/admin2">x</a>'))
    check("crm strip detected", fragment_has_crm_strip('<div id="wp-crm-context">'))
    check("crm strip absent", not fragment_has_crm_strip("<div>nope</div>"))
    word = transcript_probe_word(
        '<details class="wp-transcript-call"><summary>s</summary>'
        "<p>Quarterly revenue figures up</p></details>"
    )
    check("transcript probe word", word == "Quarterly", repr(word))
    check("transcript probe absent", transcript_probe_word("<p>nothing</p>") is None)

    with tempfile.TemporaryDirectory() as td:
        prev = os.path.join(td, "20260101T000000Z")
        cur = os.path.join(td, "20260102T000000Z")
        os.makedirs(prev)
        os.makedirs(cur)
        with open(os.path.join(prev, "01-a-light.png"), "wb") as fh:
            fh.write(b"png")
        for name in ("01-a-light.png", "02-b-light.png"):
            with open(os.path.join(cur, name), "wb") as fh:
                fh.write(b"png")
        build_compare(td, prev, cur, "vX", None, {}, badge_bar=0.5)
        with open(os.path.join(td, "compare.html"), encoding="utf-8") as fh:
            out = fh.read()
        check(
            "partial run compares gracefully",
            "no previous" in out and "02-b-light" in out and "01-a-light" in out,
        )

    with tempfile.TemporaryDirectory() as td:
        for stamp in ("20260101T000000Z", "20260102T000000Z", "20260103T000000Z"):
            os.makedirs(os.path.join(td, stamp))
        removed = prune_runs(td, 2)
        check(
            "prune keeps newest N",
            removed == ["20260101T000000Z"]
            and sorted(os.listdir(td)) == ["20260102T000000Z", "20260103T000000Z"],
            repr(removed),
        )
        check("prune refuses keep<1", prune_runs(td, 0) == [])

    try:
        from PIL import Image
    except ImportError:
        print("  skip pixel_diff math (pillow not installed — capture env only)")
    else:
        with tempfile.TemporaryDirectory() as td:

            def join(name: str) -> str:
                return os.path.join(td, name)

            Image.new("RGB", (10, 10), (0, 0, 0)).save(join("a.png"))
            Image.new("RGB", (10, 10), (255, 255, 255)).save(join("b.png"))
            Image.new("RGB", (10, 20), (0, 0, 0)).save(join("c.png"))
            same = pixel_diff(
                join("a.png"), join("a.png"), join("d0.png"), threshold=12
            )
            check(
                "pixel: identical = 0%", same is not None and same[0] == 0.0, repr(same)
            )
            full = pixel_diff(
                join("a.png"), join("b.png"), join("d1.png"), threshold=12
            )
            check(
                "pixel: full change = 100%",
                full is not None and abs(full[0] - 1.0) < 1e-9,
                repr(full),
            )
            skip = pixel_diff(join("a.png"), join("c.png"), join("d2.png"))
            check("pixel: height drift skips w/o clip", skip is None, repr(skip))
            clipped = pixel_diff(join("a.png"), join("c.png"), join("d3.png"), clip=10)
            check(
                "pixel: clip path diffs top-N",
                clipped is not None and clipped[1] is True,
                repr(clipped),
            )

    if failures:
        print(f"self-test: {len(failures)} FAILURE(S)")
        return 1
    print("self-test: all checks pass")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="/tmp/webphone-bin")  # nosec B108 - LOCAL-ONLY visual harness by budget decision (AGENTS T23)
    parser.add_argument(
        "--base-url",
        default=None,
        metavar="URL",
        help="capture a RUNNING deployment (e.g. https://pbx.artmann.tech) "
        "instead of booting --binary; sign in via WEBPHONE_CAPTURE_EXTENSION "
        "/ WEBPHONE_CAPTURE_PASSWORD; markers run lenient (warn, not abort)",
    )
    parser.add_argument(
        "--crm",
        action="store_true",
        # Turns on a fake Kith machine API beside the app (see module docstring):
        # the ONE way to shoot a CRM-linked caller name and the context strip.
        help="boot a fake CRM so crm-context surfaces render (link + strip)",
    )
    parser.add_argument(
        "--fax-off",
        action="store_true",
        # Boots the app with fax.send/receive=false (the deployed PBX's
        # declaration) and demands the notice markup, not the form.
        help="boot with fax.send/receive=false; the fax surface must show "
        "the capability notice, NOT the compose form",
    )
    parser.add_argument("--port", type=int, default=18097)
    parser.add_argument("--out-dir", default="ui-shots")
    parser.add_argument(
        "--surface",
        action="append",
        metavar="NAME",
        # A single surface re-shot in seconds instead of a 24-shot pass.
        # Repeatable: --surface thread --surface messages. Unknown names are
        # rejected below so a typo cannot silently shoot nothing.
        help="shoot only this surface (repeatable); default shoots the matrix",
    )
    parser.add_argument(
        "--track",
        action="store_true",
        help="history mode: the run lands in <out-dir>/<host>/<UTC-stamp>/ "
        "(never wiped) with gallery.html + summary.json, and compare.html at "
        "the host level diffs it against the previous run",
    )
    parser.add_argument(
        "--threads",
        type=int,
        default=6,
        metavar="N",
        help="remote mode: capture the first N live thread deep links (default 6)",
    )
    parser.add_argument(
        "--viewport",
        action="append",
        metavar="WxH",
        default=None,
        help="window size for the matrix (repeatable for extra passes, e.g. "
        "390x844 for the mobile layout; default 1280x900)",
    )
    parser.add_argument(
        "--viewport-only",
        action="store_true",
        help="capture just the window instead of the full page height",
    )
    parser.add_argument(
        "--anonymous",
        action="store_true",
        help="remote mode: capture ONLY the anonymous surfaces (login card + "
        "/login) — no credentials needed; the deployment's public face",
    )
    parser.add_argument(
        "--label",
        default=None,
        help="free-text note stored in summary.json and the galleries",
    )
    parser.add_argument(
        "--self-test",
        action="store_true",
        help="verify every pure helper (scrapes, clip math, prune, "
        "partial-compare tolerance; pixel math when pillow is present) "
        "without starting a browser; non-zero exit on any failure",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="log in, resolve the surface list and exit BEFORE a browser "
        "starts — the credential/surface sanity lane",
    )
    parser.add_argument(
        "--insecure",
        action="store_true",
        help="skip TLS verification (the LOCAL TLS-front smoke ONLY — "
        "never point this at a production host)",
    )
    parser.add_argument(
        "--host-map",
        action="append",
        metavar="HOST=IP",
        default=None,
        help="point chrome's DNS at an IP (repeatable; e.g. "
        "webphone.test=127.0.0.1 for the local TLS front)",
    )
    parser.add_argument(
        "--diff-threshold",
        type=int,
        default=12,
        metavar="N",
        help="per-pixel delta counted as changed (default 12)",
    )
    parser.add_argument(
        "--badge-bar",
        type=float,
        default=0.5,
        metavar="PCT",
        help="compare badge / console noise bar in %% (default 0.5)",
    )
    parser.add_argument(
        "--diff-clip",
        type=int,
        default=None,
        metavar="N",
        help="height-drift policy: diff the common top N pixels of "
        "full-page pairs whose heights changed (default: skip to "
        "side-by-side)",
    )
    parser.add_argument(
        "--prune",
        type=int,
        default=None,
        metavar="N",
        help="with --track: keep only the newest N runs per host "
        "(removes older run dirs after the compare is written)",
    )
    parser.add_argument(
        "--open",
        action="store_true",
        help="open the produced compare/gallery HTML when the run ends "
        "(no-op with a note on a headless host)",
    )
    args = parser.parse_args()

    if args.self_test:
        return self_test()
    if args.insecure:
        global _ssl_context
        _ssl_context = ssl._create_unverified_context()
        print("insecure: TLS verification OFF — local smoke lane only")
    _apply_host_maps(args.host_map)

    full_page = not args.viewport_only
    viewports = (
        [_parse_viewport(v) for v in args.viewport] if args.viewport else [(1280, 900)]
    )
    remote = args.base_url is not None
    base = (args.base_url or f"http://127.0.0.1:{args.port}").rstrip("/")
    only = set(args.surface) if args.surface else None
    threads_info: dict = {}
    wire: dict[str, str | None] | None = None

    if args.track:
        # History lives under <out-dir>/<host>/<UTC-stamp>/ and is NEVER
        # wiped — the diff chain (compare.html) is the whole point.
        host = _sanitize_dirname(base)
        run_id = time.strftime("%Y%m%dT%H%M%SZ", time.gmtime())
        out_dir = os.path.join(args.out_dir, host, run_id)
    else:
        out_dir = args.out_dir
    os.makedirs(out_dir, exist_ok=True)
    # A failed run must not leave a mixed matrix: stale shots from an
    # earlier pass masquerade as completeness (the 2026-10-18 settings
    # abort left 10 fresh light shots beside 8 stale ones). Track mode
    # never wipes: a run is history, not a working set.
    if not args.surface and not args.track:
        for stale in glob.glob(os.path.join(out_dir, "*.png")):
            os.remove(stale)

    if remote:
        anonymous = list(ANONYMOUS_SURFACES)
        if probe_enroll(base):
            print(
                "enroll: the deployment renders the token-gated form — "
                "/enroll joins the anonymous pass"
            )
            anonymous.append(("enroll", "/enroll"))
        threads_info: dict = {}
        if args.anonymous:
            surfaces = []
            session = ""
            version_facts = fetch_version_info(base)
        else:
            extension = os.environ.get("WEBPHONE_CAPTURE_EXTENSION")
            password = os.environ.get("WEBPHONE_CAPTURE_PASSWORD")
            if not extension or not password:
                print(
                    "remote capture needs credentials in the environment:\n"
                    "  WEBPHONE_CAPTURE_EXTENSION  the extension to sign in as\n"
                    "  WEBPHONE_CAPTURE_PASSWORD   that extension's directory password\n"
                    "(they verify against the real PBX directory; neither is logged)\n"
                    "or pass --anonymous to shoot only the login surfaces",
                    file=sys.stderr,
                )
                return 2
            try:
                urllib.request.urlopen(
                    base + "/livez", timeout=10, context=_ssl_context
                )  # nosec B310 - readiness probe of the remote server under capture
            except (OSError, http.client.HTTPException) as exc:
                print(f"{base} is unreachable ({exc})", file=sys.stderr)
                return 1
            session = login_remote(base, extension, password)
            if not session:
                return 1
            version_facts = fetch_version_info(base)
            if version_facts:
                commit = version_facts.get("commit") or "sha fenced"
                print(f"deployment: {version_facts['version']} ({commit})")
            resolved = resolve_remote_surfaces(base, session, args.threads)
            surfaces = resolved["surfaces"]
            threads_info = resolved
        wire = snapshot_wire_facts(base, session) if args.track else None
        if only is not None:
            # A typo must fail loudly: a filtered run that matches nothing
            # would print "0 shots" and look like a clean pass.
            unknown = sorted(
                only - {name for name, _ in surfaces} - {name for name, _ in anonymous}
            )
            if unknown:
                print(
                    f"unknown surface(s): {', '.join(unknown)}; resolved: "
                    f"{', '.join(name for name, _ in surfaces)}",
                    file=sys.stderr,
                )
                return 2
        if args.dry_run:
            print("dry-run: surfaces resolved; no browser started")
            print("  anonymous: " + ", ".join(n for n, _ in anonymous))
            print(
                "  session:   "
                + (", ".join(n for n, _ in surfaces) or "(anonymous mode)")
            )
            return 0
        print(
            "prod capture: shots carry REAL data — ui-shots/ is gitignored; "
            "keep it that way"
        )
        try:
            ledger = capture(
                base,
                out_dir,
                session,
                surfaces,
                only,
                viewports=viewports,
                full_page=full_page,
                lenient=True,
                anonymous_surfaces=anonymous,
                insecure=args.insecure,
                host_maps=args.host_map,
            )
        except Exception:
            if args.track:
                discard_empty_run(out_dir)
            raise
        shots = len(ledger)
        exit_code = 0 if shots > 0 else 1
    else:
        data_dir = "/tmp/wp-visual-data"  # nosec B108 - LOCAL-ONLY visual harness by budget decision (AGENTS T23)
        os.makedirs(data_dir, exist_ok=True)
        # A CONFIG FILE, not env: the browser posts carry Origin/Referer, and
        # the CSRF middleware only trusts them with the fronted shape
        # (csrf.trusted_*) — the same configuration the NixOS module ships
        # behind caddy and the smoke's boot_configured exercises.
        config = {
            "addr": f"127.0.0.1:{args.port}",
            "data_dir": data_dir,
            "gateway": {"webhook_secret": "devsecret"},  # nosec B105 - the documented dev-only secret (AGENTS smoke recipe uses the same value)
            # ASR ON with an obviously-fake provider (M13): url alone turns
            # the seam on (openai wire, BASE url — the client appends the
            # endpoint path). shell.js auto-transcribes seen audio, so the
            # call goes out and fails fast against the refused port; the
            # settings row shows Describe(), the buttons carry
            # data-transcribe-src.
            "asr": {
                "url": "http://127.0.0.1:9",
                "model": "ui-capture-whisper",
            },
            "csrf": {
                "trusted_proxies": ["127.0.0.1"],
                "trusted_origins": [base],
            },
            # Passkey ON, render-only (2026-10-25): turns the mode on so the
            # two anonymous login surfaces exist; no ceremony ever runs, so
            # the rp_id never faces a browser — it only has to satisfy boot
            # validation (hostname-shaped, host of every origin entry).
            "auth": {
                "passkey": {
                    "rp_id": "webphone.test",
                    "rp_origins": ["https://webphone.test"],
                    "users": {"ui-capture@example.com": {"extensions": ["1001"]}},
                    "extension_password_files": {
                        "1001": os.path.join(data_dir, "ext1001.pass")
                    },
                }
            },
        }
        if args.fax_off:
            config["fax"] = {"send": False, "receive": False}
            # Tighten the fax pin from "form or notice" to notice-only, and
            # forbid the form outright — the whole point of the off-state shot.
            SURFACE_MARKERS["fax"] = [
                (
                    ".wp-fax-off",
                    None,
                ),
            ]
            SURFACE_ABSENT_MARKERS["fax"] = (".wp-compose-fax",)
        if args.crm:
            crm_url, crm_token = start_fake_crm()
            config["crm"] = {
                "url": crm_url,
                "token": crm_token,
                # The human UI lives on a different origin than the machine API in
                # every real deployment, so the template is config, never inferred.
                "contact_url": "https://crm.example/contacts/{id}",
            }
        with open(os.path.join(data_dir, "ext1001.pass"), "w", encoding="utf-8") as fh:
            fh.write("ui-capture-fixture\n")  # nosec B105 - LOCAL-ONLY visual harness; never a real credential
        config_path = os.path.join(data_dir, "ui-capture-config.json")
        with open(config_path, "w", encoding="utf-8") as fh:
            json.dump(config, fh)
        env = dict(os.environ, WEBPHONE_CONFIG=config_path)
        server = subprocess.Popen([args.binary], env=env)  # nosec B603 - argv-list (no shell) launch of the local binary under test
        exit_code = 1
        try:
            for _ in range(50):
                try:
                    urllib.request.urlopen(base + "/livez", timeout=1)  # nosec B310 - readiness probe of the local server under test
                    break
                except (OSError, http.client.HTTPException):
                    time.sleep(0.2)
            else:
                print("server did not come up", file=sys.stderr)
                return 1
            version_facts = fetch_version_info(base)
            thread_path, attach_thread_path, session = seed(base)
            if not session:
                print("seed did not mint a session cookie", file=sys.stderr)
                return 1
            anonymous = list(ANONYMOUS_SURFACES)
            if probe_enroll(base):
                anonymous.append(("enroll", "/enroll"))
            surfaces: list[tuple[str, str | None]] = []
            for name, path in TABS:
                if name == "thread":
                    path = thread_path
                elif name == "crm-context":
                    # Without the seam there is nothing to prove: the strip
                    # does not exist and the caller name is plain text. A
                    # filtered run of this surface errors below instead.
                    path = thread_path if args.crm else None
                elif name == "media-transcribe":
                    path = attach_thread_path
                surfaces.append((name, path))
            if only is not None:
                # A typo must fail loudly: a filtered run that matches nothing
                # would print "0 shots" and look like a clean pass.
                known = {name for name, _ in TABS} | {name for name, _ in anonymous}
                if "crm-context" in only and not args.crm:
                    print(
                        "surface 'crm-context' needs --crm: without the fake CRM the "
                        "strip does not exist and the shot would prove nothing",
                        file=sys.stderr,
                    )

                    return 2
                unknown = sorted(only - known)
                if unknown:
                    print(
                        f"unknown surface(s): {', '.join(unknown)}; known: "
                        f"{', '.join(sorted(known))}",
                        file=sys.stderr,
                    )
                    return 2
            wire = snapshot_wire_facts(base, session) if args.track else None
            if args.dry_run:
                print("dry-run: fixture seeded, surfaces resolved; no browser started")
                print("  anonymous: " + ", ".join(n for n, _ in anonymous))
                print("  session:   " + ", ".join(n for n, _ in surfaces))
                return 0
            try:
                ledger = capture(
                    base,
                    out_dir,
                    session,
                    surfaces,
                    only,
                    viewports=viewports,
                    full_page=full_page,
                    layout_proof=True,
                    anonymous_surfaces=anonymous,
                    insecure=args.insecure,
                    host_maps=args.host_map,
                )
            except Exception:
                if args.track:
                    discard_empty_run(out_dir)
                raise
            shots = len(ledger)
            exit_code = 0 if shots >= 22 else 1
        finally:
            server.terminate()
            server.wait(timeout=10)

    print(f"ui-capture: {shots} shots in {out_dir}/")
    if args.track:
        produced = track(
            out_dir,
            base,
            version_facts,
            args.label,
            full_page,
            viewports,
            ledger=ledger,
            wire=wire,
            threads=threads_info or None,
            diff_threshold=args.diff_threshold,
            badge_bar=args.badge_bar,
            diff_clip=args.diff_clip,
            prune_keep=args.prune,
        )
        if args.open:
            if os.environ.get("DISPLAY") or os.environ.get("WAYLAND_DISPLAY"):
                webbrowser.open("file://" + os.path.abspath(produced or out_dir))
            else:
                print("--open: no display on this host — open the HTML manually")
    if only is not None and not remote:
        # Every requested surface must have produced BOTH theme shots
        # (per viewport).
        expected = len(only) * 2 * len(viewports)
        return 0 if shots == expected else 1
    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
