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

Budget decision (documented, 2026-10-02): this is a LOCAL harness, NOT a
flake check — headless chromium cannot run inside a nix build sandbox
without a KVM VM, and a KVM browser VM here would duplicate the stack's
proven browser-E2E lane at double the gate cost. The shots are eyeball
material (and optional pixel-diff input for the owner); timestamps and
relative times make byte-goldens flake by design.

Run (from the repo root):
  nix build .#webphone --out-link /tmp/wp-visual-bin
  nix shell nixpkgs#chromium nixpkgs#python312.withPackages(ps: [ ps.selenium ]) \\
    --command python3 scripts/ui-capture.py --binary /tmp/wp-visual-bin/bin/webphone

Output: ui-shots/<n>-<name>-<light|dark>.png (up to 24 files).
"""

from __future__ import annotations

import argparse
import glob
import http.client
import json
import os
import subprocess  # nosec B404 - deliberately spawns the local webphone binary under test
import sys
import threading
import time
import urllib.parse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import urllib.request

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
    "fax": ".wp-compose-fax",  # the compose form is always present
    "voicemail": "#voicemail-panel",
    "history": "#history-panel",
    "history-transcript": "[data-copy-transcript]",  # seeded transcript section
    "contacts": ".wp-export-link",  # unconditional in the panel
    "settings": ("#settings-panel dd .wp-service-on", "openai-compatible"),
    "not-found": (".wp-panel.wp-notfound", "404 — page not found"),
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

    def do_GET(self) -> None:  # noqa: N802 - http.server's spelling
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

    def do_POST(self) -> None:  # noqa: N802 - http.server's spelling
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

    def log_message(self, format: str, *args: object) -> None:  # noqa: A002 - http.server's signature
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


def capture(
    base: str,
    out_dir: str,
    thread_path: str | None,
    attach_thread_path: str | None,
    session: str,
    only: set[str] | None = None,
    crm_on: bool = False,
) -> int:
    # Selenium lives only in the capture env (see the AGENTS run recipe
    # and mypy.ini) — never in the lint shell.
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.support.ui import WebDriverWait

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
    try:
        driver.get(base + "/")
        WebDriverWait(driver, 10).until(
            lambda d: d.execute_script("return document.readyState") == "complete"
        )
        # The seed's session cookie, injected (see seed()): the
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
            # and /login 303s a live session home — drop the cookie for
            # these two shots, restore it after.
            driver.delete_cookie("webphone_session")
            for name, path in ANONYMOUS_SURFACES:
                if only is not None and name not in only:
                    continue
                driver.get(base + path)
                if name == "login-card":
                    # The island reveals the card once the anonymous resume
                    # probe fails — wait on the link itself (the surface's
                    # whole point) instead of a fixed sleep.
                    WebDriverWait(driver, 8).until(
                        lambda d: d.find_elements("css selector", "#passkey-login")
                    )
                time.sleep(0.4)  # settle: theme script, fonts
                marker, expected_text = ANONYMOUS_MARKERS[name]
                if not driver.find_elements("css selector", marker):
                    raise AssertionError(
                        f"surface {name!r} at {path} is missing {marker!r} — "
                        f"url={driver.current_url} — the shot would not be "
                        "evidence; aborting"
                    )
                if (
                    expected_text
                    and expected_text
                    not in driver.find_element("css selector", "body").text
                ):
                    raise AssertionError(
                        f"surface {name!r} rendered without {expected_text!r} — "
                        "the marker matched but the copy is wrong; aborting"
                    )
                target = os.path.join(out_dir, f"{shots + 1:02d}-{name}-{theme}.png")
                driver.save_screenshot(target)
                print(f"captured {target}")
                shots += 1
            driver.add_cookie({"name": "webphone_session", "value": session})
            for name, path in TABS:
                if only is not None and name not in only:
                    continue
                if name in ("thread", "crm-context"):
                    path = thread_path
                if name == "media-transcribe":
                    path = attach_thread_path
                if name == "crm-context" and not crm_on:
                    # Without the seam there is nothing to prove: the strip
                    # does not exist and the caller name is plain text. A
                    # filtered run of this surface errors in main() instead.
                    continue
                if path is None:
                    # An unresolved deep link (seed came up empty) skips
                    # the surface rather than shooting an error page.
                    continue
                driver.get(base + path)
                # The island reveals its call view only after the resume
                # probe settles (login view stays hidden meanwhile); the
                # loopback SIP refusal latency varies by pass, so wait on
                # the view itself instead of a fixed sleep.
                WebDriverWait(driver, 8).until(
                    lambda d: (
                        not d.find_element("css selector", "#phone-view").get_attribute(
                            "hidden"
                        )
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
                # A marker is a selector, a (selector, copy) tuple, or a
                # LIST of those pins (a surface that proves two things).
                pins = SURFACE_MARKERS[name]
                if isinstance(pins, str):
                    pins = [(pins, None)]
                elif isinstance(pins, tuple):
                    pins = [pins]
                for marker, expected_text in pins:
                    if not driver.find_elements("css selector", marker):
                        raise AssertionError(
                            f"surface {name!r} at {path} is missing {marker!r} — "
                            f"url={driver.current_url} anonymous_shell="
                            f"{bool(driver.find_elements('id', 'login-form'))} — "
                            "the shot would not be evidence; aborting"
                        )
                    if (
                        expected_text
                        and expected_text
                        not in driver.find_element("css selector", "body").text
                    ):
                        raise AssertionError(
                            f"surface {name!r} rendered without {expected_text!r} — "
                            "the marker matched but the copy is wrong; aborting"
                        )
                target = os.path.join(out_dir, f"{shots + 1:02d}-{name}-{theme}.png")
                driver.save_screenshot(target)
                print(f"captured {target}")
                shots += 1
                if name == "thread" and theme == "light":
                    # Layout facts belong here, not in a second script: a
                    # screenshot proves a surface RENDERS, not that the
                    # collapse reclaims width or that the master-detail split
                    # flips at the container threshold. Measured as COMPUTED
                    # STYLE, once per pass, on the surface that carries both
                    # panes (2026-10-29 closeout; a standalone proof script
                    # duplicated this config and hung for three hours).
                    check_thread_layout(driver, base, thread_path, out_dir)
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
        "--crm",
        action="store_true",
        # Turns on a fake Kith machine API beside the app (see module docstring):
        # the ONE way to shoot a CRM-linked caller name and the context strip.
        help="boot a fake CRM so crm-context surfaces render (link + strip)",
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
    args = parser.parse_args()

    data_dir = "/tmp/wp-visual-data"  # nosec B108 - LOCAL-ONLY visual harness by budget decision (AGENTS T23)
    os.makedirs(data_dir, exist_ok=True)
    os.makedirs(args.out_dir, exist_ok=True)
    # A failed run must not leave a mixed matrix: stale shots from an
    # earlier pass masquerade as completeness (the 2026-10-18 settings
    # abort left 10 fresh light shots beside 8 stale ones).
    if not args.surface:
        for stale in glob.glob(os.path.join(args.out_dir, "*.png")):
            os.remove(stale)
    base = f"http://127.0.0.1:{args.port}"
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
        only = set(args.surface) if args.surface else None
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
            args.out_dir,
            thread_path,
            attach_thread_path,
            session,
            only,
            crm_on=args.crm,
        )
        print(f"ui-capture: {shots} shots in {args.out_dir}/")
        if only is not None:
            # Every requested surface must have produced BOTH theme shots.
            expected = len(only) * 2
            return 0 if shots == expected else 1
        return 0 if shots >= 22 else 1
    finally:
        server.terminate()
        server.wait(timeout=10)


if __name__ == "__main__":
    raise SystemExit(main())
