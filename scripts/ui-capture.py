#!/usr/bin/env python3
"""Visual capture harness (T23): boots a fresh webphone binary (loopback
gateway — the whole product, zero PBX), seeds deterministic content over
HTTP, then drives headless chromium (selenium — the SAME driver family
the consuming stack's browser E2E uses) through the server tabs and
screenshots each in light and dark: the 20-shot matrix (10 surfaces x 2
themes).

M13 extension (2026-10-08): the boot turns the ASR seam ON (a fake
provider URL — nothing ever calls it at render time) and three surfaces
join the matrix: the Settings ASR row with provider detail, a History
transcript search (seeded via POST /api/transcripts), and a thread with
an AUDIO attachment whose transcribe button carries the same
data-transcribe-src shape the voicemail rows use (a bare boot has no
voicemail rows — PBX-owned — so the attachment variant is the
deterministic stand-in; the island's live call card needs a real SIP
call and stays covered by the island node specs).

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

Output: ui-shots/<n>-<name>-<light|dark>.png (20 files).
"""

from __future__ import annotations

import argparse
import http.client
import json
import os
import subprocess
import sys
import time
import urllib.parse
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
    ("media-transcribe", None),  # the attachment thread, resolved after seeding
    ("fax", "/fax"),
    ("voicemail", "/voicemail"),
    ("history", "/history"),
    ("history-transcript", "/history?q=revenue"),  # seeded transcript search
    ("contacts", "/contacts"),
    ("settings", "/settings"),
]
# Capture-time DOM assertion per surface: a screenshot is only evidence
# when the DOM beneath it is the real surface (this check would have
# caught the thread deep link rendering the messages list — the shots
# were byte-identical until the fix). A (selector, text) tuple also
# pins rendered copy inside the surface.
SURFACE_MARKERS = {
    "messages": ".wp-thread-row",  # seeded threads render rows
    "thread": "#wp-thread-head",  # the open conversation head
    "media-transcribe": "[data-transcribe-src]",  # audio attachment button (ASR on)
    "fax": ".wp-compose-fax",  # the compose form is always present
    "voicemail": "#voicemail-panel",
    "history": "#history-panel",
    "history-transcript": "[data-copy-transcript]",  # seeded transcript section
    "contacts": ".wp-export-link",  # unconditional in the panel
    "settings": ("#settings-panel dd.wp-service-on", "openai-compatible"),
}


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
                        f'--{boundary}\r\nContent-Disposition: form-data; '
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
    csrf = re.search(r'name="csrf-token" content="([^"]+)"', page).group(1)
    body = json.dumps({"extension": "1001", "password": "pw"}).encode()
    req = urllib.request.Request(base + "/api/session", data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-CSRF-Token", csrf)
    opener.open(req)
    csrf = json.loads(opener.open(base + "/api/csrf").read()).get("token", "")
    # The attachment thread goes in FIRST (see SEED_ATTACHMENT_THREAD):
    # recency ordering keeps it at the list's tail.
    remote, subject, body = SEED_ATTACHMENT_THREAD
    post(
        "/messages/send",
        {"to": remote, "body": body},
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
        if cookie.name == "webphone_session":
            session = cookie.value
    return thread_path, attach_thread_path, session


def capture(
    base: str,
    out_dir: str,
    thread_path: str | None,
    attach_thread_path: str | None,
    session: str,
) -> int:
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
            for name, path in TABS:
                if name == "thread":
                    if not thread_path:
                        continue
                    path = thread_path
                if name == "media-transcribe":
                    if not attach_thread_path:
                        continue
                    path = attach_thread_path
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
                marker = SURFACE_MARKERS[name]
                expected_text = None
                if isinstance(marker, tuple):
                    marker, expected_text = marker
                if not driver.find_elements("css selector", marker):
                    raise AssertionError(
                        f"surface {name!r} at {path} is missing {marker!r} — "
                        f"url={driver.current_url} anonymous_shell="
                        f"{bool(driver.find_elements('id', 'login-form'))} — "
                        "the shot would not be evidence; aborting"
                    )
                if expected_text and expected_text not in driver.find_element(
                    "css selector", "body"
                ).text:
                    raise AssertionError(
                        f"surface {name!r} rendered without {expected_text!r} — "
                        "the marker matched but the copy is wrong; aborting"
                    )
                target = os.path.join(out_dir, f"{shots + 1:02d}-{name}-{theme}.png")
                driver.save_screenshot(target)
                print(f"captured {target}")
                shots += 1
    finally:
        driver.quit()
    return shots


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="/tmp/webphone-bin")
    parser.add_argument("--port", type=int, default=18097)
    parser.add_argument("--out-dir", default="ui-shots")
    args = parser.parse_args()

    data_dir = "/tmp/wp-visual-data"
    os.makedirs(data_dir, exist_ok=True)
    os.makedirs(args.out_dir, exist_ok=True)
    base = f"http://127.0.0.1:{args.port}"
    # A CONFIG FILE, not env: the browser posts carry Origin/Referer, and
    # the CSRF middleware only trusts them with the fronted shape
    # (csrf.trusted_*) — the same configuration the NixOS module ships
    # behind caddy and the smoke's boot_configured exercises.
    config = {
        "addr": f"127.0.0.1:{args.port}",
        "data_dir": data_dir,
        "gateway": {"webhook_secret": "devsecret"},
        # ASR ON with an obviously-fake provider (M13): url alone turns
        # the seam on (openai wire); nothing calls it at render time —
        # the settings row shows Describe(), the buttons merely carry
        # data-transcribe-src.
        "asr": {
            "url": "http://127.0.0.1:9/v1/audio/transcriptions",
            "model": "ui-capture-whisper",
        },
        "csrf": {
            "trusted_proxies": ["127.0.0.1"],
            "trusted_origins": [base],
        },
    }
    config_path = os.path.join(data_dir, "ui-capture-config.json")
    with open(config_path, "w", encoding="utf-8") as fh:
        json.dump(config, fh)
    env = dict(os.environ, WEBPHONE_CONFIG=config_path)
    server = subprocess.Popen([args.binary], env=env)
    try:
        for _ in range(50):
            try:
                urllib.request.urlopen(base + "/livez", timeout=1)
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
        shots = capture(
            base, args.out_dir, thread_path, attach_thread_path, session
        )
        print(f"ui-capture: {shots} shots in {args.out_dir}/")
        return 0 if shots >= 20 else 1
    finally:
        server.terminate()
        server.wait(timeout=10)


if __name__ == "__main__":
    raise SystemExit(main())
