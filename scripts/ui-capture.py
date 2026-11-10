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
import subprocess  # nosec B404 - deliberately spawns the local webphone binary under test
import sys
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

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
}
# Selectors that must NOT exist on a surface. Only --fax-off populates
# an entry (the fax surface must not carry the compose form while the
# boot declares send=false); the loop asserts absence after the pins.
SURFACE_ABSENT_MARKERS: dict[str, tuple[str, ...]] = {}


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
    marker = 'hx-get="/partials/messages/'
    starts = [m.start() for m in re.finditer(marker, html)]

    def link_at(idx: int) -> str | None:
        # Skips past 'hx-get="/partials' so the slice keeps the deep-link
        # shape "/messages/<thread-id>" (a leading slash driver.get needs).
        start = starts[idx] + len('hx-get="/partials')
        return html[start : html.find('"', start)]

    thread_path = link_at(0) if starts else None
    attach_thread_path = link_at(-1) if len(starts) > 1 else None
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
    opener = urllib.request.build_opener(jar)
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
    with urllib.request.urlopen(req, timeout=30) as resp:  # nosec B310 - read-only fetch of the server under capture
        return resp.read().decode("utf-8", "replace")


def resolve_remote_surfaces(
    base: str, session: str, threads_n: int
) -> list[tuple[str, str | None]]:
    """Build the surface list from LIVE data: the first `threads_n` thread
    deep links off the real list, plus the static tabs. Prints the resolved
    notes so the operator sees exactly what will be shot."""
    links: list[str] = []
    try:
        page = _get(base, "/messages", session)
        marker = 'hx-get="/partials/messages/'
        for match in re.finditer(re.escape(marker), page):
            start = match.start() + len('hx-get="/partials')
            link = page[start : page.find('"', start)]
            if link and link not in links:
                links.append(link)
    except (OSError, http.client.HTTPException) as exc:
        print(
            f"resolve: /messages unreadable ({exc}) — thread surfaces skipped",
            file=sys.stderr,
        )
    surfaces: list[tuple[str, str | None]] = [("messages", "/messages")]
    chosen = links[:threads_n]
    for idx, link in enumerate(chosen, 1):
        surfaces.append((f"thread-{idx:02d}", link))
    for idx, link in enumerate(chosen, 1):
        try:
            fragment = _get(base, "/partials" + link, session)
        except (OSError, http.client.HTTPException):
            continue
        if "data-transcribe-src" in fragment:
            print(f"media: thread-{idx:02d} carries transcribable audio ({link})")
    surfaces += [("fax", "/fax"), ("voicemail", "/voicemail"), ("history", "/history")]
    try:
        if "data-copy-transcript" in _get(base, "/history", session):
            surfaces.append(("history-transcript", "/history"))
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
        ("not-found", "/nope"),
    ]
    print("surfaces: " + ", ".join(name for name, _ in surfaces))
    return surfaces


def fetch_version(base: str) -> str | None:
    """The deployment's /version (open route); None when fenced or absent."""
    try:
        with urllib.request.urlopen(base + "/version", timeout=10) as resp:  # nosec B310 - read-only version probe
            return str(json.loads(resp.read()).get("version") or "") or None
    except (OSError, ValueError, http.client.HTTPException):
        return None


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
    prev_path: str, cur_path: str, diff_path: str, threshold: int = 12
) -> tuple[float, bool] | None:
    """Heat-map the changed pixels into diff_path; returns the changed
    fraction, or None without pillow / on a size mismatch (content height
    changed — the side-by-side carries the signal there)."""
    try:
        from PIL import Image, ImageChops, ImageEnhance
    except ImportError:
        return None
    prev = Image.open(prev_path).convert("RGB")
    cur = Image.open(cur_path).convert("RGB")
    if prev.size != cur.size:
        return None
    grey = ImageChops.difference(prev, cur).convert("L")
    hist = grey.histogram()
    total = prev.size[0] * prev.size[1]
    changed = sum(hist[threshold:]) / total
    mask = grey.point(lambda v: 255 if v >= threshold else 0)
    faded = ImageEnhance.Brightness(prev).enhance(0.35)
    heat = Image.composite(Image.new("RGB", prev.size, (235, 20, 60)), faded, mask)
    os.makedirs(os.path.dirname(diff_path), exist_ok=True)
    heat.save(diff_path)
    return changed, True


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
.missing{color:#666;border:1px dashed #444;border-radius:6px;padding:24px;text-align:center}
</style></head><body>
"""


def build_gallery(
    run_dir: str, base: str, version: str | None, label: str | None
) -> None:
    parts = [f"<h1>{html.escape(base)}</h1><p>run {html.escape(os.path.basename(run_dir))}"]
    if version:
        parts.append(f" · {html.escape(version)}")
    if label:
        parts.append(f" · {html.escape(label)}")
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
) -> None:
    """One compare.html per host: the newest run against the previous one,
    per surface — previous | current | diff, worst offenders first."""

    def _collect(run: str) -> dict[str, str]:
        return {
            shot_key(p): os.path.basename(p)
            for p in glob.glob(os.path.join(run, "*.png"))
        }

    prev_files, cur_files = _collect(prev_dir), _collect(cur_dir)
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

    def _badge(key: str) -> str:
        if key not in prev_files:
            return '<span class="badge new">new</span>'
        pct = changed.get(key)
        if pct is not None and pct >= 0.5:
            return f'<span class="badge">{pct:.2f}%</span>'
        return ""

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
    sections = []
    for key in keys:
        diff_rel = f"diffs/{key}.png" if key in changed else None
        sections.append(
            f"<section><h2>{html.escape(key)}{_badge(key)}</h2>"
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
    label: str | None,
    full_page: bool,
    viewports: list[tuple[int, int]],
) -> None:
    """History + diff bookkeeping for --track runs: summary.json, the
    per-run gallery, and the host-level compare against the previous run."""
    host_dir = os.path.dirname(run_dir)
    runs = sorted(
        d
        for d in os.listdir(host_dir)
        if os.path.isdir(os.path.join(host_dir, d)) and re.match(r"^\d{8}T\d{6}Z$", d)
    )
    run_id = os.path.basename(run_dir)
    earlier = [d for d in runs if d < run_id]
    prev_name = earlier[-1] if earlier else None
    prev_dir = os.path.join(host_dir, prev_name) if prev_name else None
    version = fetch_version(base)
    prev_by_key: dict[str, str] = {}
    if prev_dir:
        for path in glob.glob(os.path.join(prev_dir, "*.png")):
            prev_by_key[shot_key(path)] = path
    shots = []
    diffed: list[tuple[str, float]] = []
    for path in sorted(glob.glob(os.path.join(run_dir, "*.png"))):
        key = shot_key(path)
        entry: dict[str, object] = {
            "key": key,
            "file": os.path.basename(path),
            "bytes": os.path.getsize(path),
            "prev_file": None,
            "changed_pct": None,
            "diff_file": None,
        }
        prev_path = prev_by_key.get(key)
        if prev_path:
            entry["prev_file"] = os.path.relpath(prev_path, run_dir)
            diff_path = os.path.join(run_dir, "diffs", f"{key}.png")
            result = pixel_diff(prev_path, path, diff_path)
            if result is None:
                entry["size_match"] = False
                print(f"diff: {key} — height changed; side-by-side only")
            else:
                changed_pct = result[0] * 100
                entry["changed_pct"] = round(changed_pct, 3)
                entry["diff_file"] = os.path.relpath(diff_path, run_dir)
                entry["size_match"] = True
                diffed.append((key, changed_pct))
        shots.append(entry)
    summary = {
        "run": run_id,
        "base": base,
        "version": version,
        "label": label,
        "full_page": full_page,
        "viewports": [f"{w}x{h}" for w, h in viewports],
        "prev_run": prev_name,
        "shots": shots,
    }
    with open(os.path.join(run_dir, "summary.json"), "w", encoding="utf-8") as fh:
        json.dump(summary, fh, indent=2, sort_keys=True)
        fh.write("\n")
    build_gallery(run_dir, base, version, label)
    print(f"gallery: {os.path.join(run_dir, 'gallery.html')}")
    if prev_dir:
        build_compare(host_dir, prev_dir, run_dir, version, label, dict(diffed))
        print(f"compare: {os.path.join(host_dir, 'compare.html')} (vs {prev_name})")
    else:
        print("compare: no previous run — first run in this history")
    diffed.sort(key=lambda pair: pair[1], reverse=True)
    notable = [(key, pct) for key, pct in diffed if pct >= 0.5]
    print(f"track: {len(shots)} shot(s), {len(diffed)} differ from {prev_name or '—'}")
    for key, pct in notable[:10]:
        print(f"  changed {pct:5.2f}%  {key}")
    if diffed and not notable:
        print("  nothing above the 0.5% noise bar")
    print(f"history: {len(runs)} run(s) under {host_dir}")


def _parse_viewport(text: str) -> tuple[int, int]:
    width, _, height = text.lower().partition("x")
    if not width.isdigit() or not height.isdigit():
        raise SystemExit(
            f"bad --viewport {text!r} (want WxH, e.g. 1280x900 or 390x844)"
        )
    return int(width), int(height)


def _sanitize_dirname(base: str) -> str:
    return re.sub(r"[^A-Za-z0-9._-]+", "_", urllib.parse.urlsplit(base).netloc or "local")


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
) -> int:
    """Drive headless chromium through every (surface, theme, viewport)
    cell and screenshot it; returns the shot count.

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
    import shutil

    chrome = shutil.which("chromium")
    if chrome:
        options.binary_location = chrome
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--use-fake-device-for-media-stream")
    options.add_argument("--use-fake-ui-for-media-stream")
    options.add_argument("--window-size=1280,900")
    driver_path = shutil.which("chromedriver")
    from selenium.webdriver.chrome.service import Service

    service = Service(executable_path=driver_path) if driver_path else None
    driver = webdriver.Chrome(service=service, options=options)
    shots = 0

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
        # A marker is a selector, a (selector, copy) tuple, or a LIST of
        # those pins (a surface that proves two things). Strict mode raises
        # inside _marker_fail; lenient mode warns and keeps going.
        if isinstance(pins, str):
            pins = [(pins, None)]
        elif isinstance(pins, tuple):
            pins = [pins]
        for marker, expected_text in pins:
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
            # server-rendered tabs ride it directly.
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
                for name, path in ANONYMOUS_SURFACES:
                    if only is not None and name not in only:
                        continue
                    driver.get(base + path)
                    if name == "login-card":
                        # The island reveals the card once the anonymous
                        # resume probe fails — wait on the sign-in affordance
                        # itself (the passkey link when the mode is on, the
                        # classic form otherwise) instead of a fixed sleep.
                        WebDriverWait(driver, 8).until(
                            lambda d: d.find_elements("css selector", "#passkey-login")
                            or d.find_elements("css selector", "#login-form")
                        )
                    time.sleep(0.4)  # settle: theme script, fonts
                    if lenient and _is_styled_404():
                        print(
                            f"skip: {name} at {path} is the styled 404 "
                            "(capability off on this deployment)"
                        )
                        continue
                    _check_pins(name, path, ANONYMOUS_MARKERS[name])
                    target = _target(name, theme, vp)
                    shoot(driver, target, full_page)
                    print(f"captured {target}")
                    shots += 1
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
                    if lenient and _is_styled_404():
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
                            expandable[0].click()
                    _check_pins(name, path, SURFACE_MARKERS[name])
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
                    print(f"captured {target}")
                    shots += 1
                    if (
                        name == "thread"
                        and theme == "light"
                        and layout_proof
                        and len(viewports) == 1
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
    return shots


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
        "--label",
        default=None,
        help="free-text note stored in summary.json and the galleries",
    )
    args = parser.parse_args()

    full_page = not args.viewport_only
    viewports = (
        [_parse_viewport(v) for v in args.viewport] if args.viewport else [(1280, 900)]
    )
    remote = args.base_url is not None
    base = (args.base_url or f"http://127.0.0.1:{args.port}").rstrip("/")
    only = set(args.surface) if args.surface else None

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
        extension = os.environ.get("WEBPHONE_CAPTURE_EXTENSION")
        password = os.environ.get("WEBPHONE_CAPTURE_PASSWORD")
        if not extension or not password:
            print(
                "remote capture needs credentials in the environment:\n"
                "  WEBPHONE_CAPTURE_EXTENSION  the extension to sign in as\n"
                "  WEBPHONE_CAPTURE_PASSWORD   that extension's directory password\n"
                "(they verify against the real PBX directory; neither is logged)",
                file=sys.stderr,
            )
            return 2
        try:
            urllib.request.urlopen(base + "/livez", timeout=10)  # nosec B310 - readiness probe of the remote server under capture
        except (OSError, http.client.HTTPException) as exc:
            print(f"{base} is unreachable ({exc})", file=sys.stderr)
            return 1
        session = login_remote(base, extension, password)
        if not session:
            return 1
        surfaces = resolve_remote_surfaces(base, session, args.threads)
        if only is not None:
            # A typo must fail loudly: a filtered run that matches nothing
            # would print "0 shots" and look like a clean pass.
            unknown = sorted(
                only
                - {name for name, _ in surfaces}
                - {name for name, _ in ANONYMOUS_SURFACES}
            )
            if unknown:
                print(
                    f"unknown surface(s): {', '.join(unknown)}; resolved: "
                    f"{', '.join(name for name, _ in surfaces)}",
                    file=sys.stderr,
                )
                return 2
        print(
            "prod capture: shots carry REAL data — ui-shots/ is gitignored; "
            "keep it that way"
        )
        shots = capture(
            base,
            out_dir,
            session,
            surfaces,
            only,
            viewports=viewports,
            full_page=full_page,
            lenient=True,
        )
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
            thread_path, attach_thread_path, session = seed(base)
            if not session:
                print("seed did not mint a session cookie", file=sys.stderr)
                return 1
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
                known = {name for name, _ in TABS} | {
                    name for name, _ in ANONYMOUS_SURFACES
                }
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
            shots = capture(
                base,
                out_dir,
                session,
                surfaces,
                only,
                viewports=viewports,
                full_page=full_page,
                layout_proof=True,
            )
            exit_code = 0 if shots >= 22 else 1
        finally:
            server.terminate()
            server.wait(timeout=10)

    print(f"ui-capture: {shots} shots in {out_dir}/")
    if args.track:
        track(out_dir, base, args.label, full_page, viewports)
    if only is not None and not remote:
        # Every requested surface must have produced BOTH theme shots
        # (per viewport).
        expected = len(only) * 2 * len(viewports)
        return 0 if shots == expected else 1
    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
