#!/usr/bin/env python3
"""Visual capture harness (T23): boots a fresh webphone binary (loopback
gateway — the whole product, zero PBX), seeds deterministic content over
HTTP, then drives headless chromium (selenium — the SAME driver family
the consuming stack's browser E2E uses) through the six server tabs and
screenshots each in light and dark: the 14-shot matrix (7 surfaces x 2 themes).

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

Output: ui-shots/<n>-<name>-<light|dark>.png (14 files).
"""

from __future__ import annotations

import argparse
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
SEED_SNIPPETS = [
    ("Thanks, on it!", True),
    ("Please call back later", False),
]
SEED_CONTACTS = [("Marta Jensen", "+441632960961")]
TABS = [
    ("messages", "/messages"),
    ("thread", None),  # first thread, resolved after seeding
    ("fax", "/fax"),
    ("voicemail", "/voicemail"),
    ("history", "/history"),
    ("contacts", "/contacts"),
    ("settings", "/settings"),
]
# Capture-time DOM assertion per surface: a screenshot is only evidence
# when the DOM beneath it is the real surface (this check would have
# caught the thread deep link rendering the messages list — the shots
# were byte-identical until the fix).
SURFACE_MARKERS = {
    "messages": ".wp-thread-row",  # seeded threads render rows
    "thread": "#wp-thread-head",  # the open conversation head
    "fax": ".wp-compose-fax",  # the compose form is always present
    "voicemail": "#voicemail-panel",
    "history": "#history-panel",
    "contacts": ".wp-export-link",  # unconditional in the panel
    "settings": "#settings-panel",
}


def seed(base: str) -> tuple[str | None, str]:
    """Deterministic content over HTTP; returns (thread path, session cookie)."""
    csrf = ""
    cookie_processor = urllib.request.HTTPCookieProcessor()
    opener = urllib.request.build_opener(cookie_processor)

    def post(path: str, fields: dict[str, str], multipart: bool = False) -> None:
        if multipart:
            # /messages/send rides the multipart prologue (the composer
            # uploads attachments; plain fields travel as bare parts).
            boundary = "wp-visual-seed"
            body = "".join(
                f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n'
                for k, v in fields.items()
            ) + f"--{boundary}--\r\n"
            data = body.encode()
            ctype = f"multipart/form-data; boundary={boundary}"
        else:
            data = urllib.parse.urlencode(fields).encode()
            ctype = "application/x-www-form-urlencoded"
        req = urllib.request.Request(base + path, data=data, method="POST")
        req.add_header("Content-Type", ctype)
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
    for remote, _, message in SEED_THREADS:
        post("/messages/send", {"to": remote, "body": message}, multipart=True)
    for text, quick in SEED_SNIPPETS:
        post("/snippets/save", {"body": text, **({"quick": "1"} if quick else {})})
    for name, number in SEED_CONTACTS:
        post("/contacts/save", {"name": name, "number": number})
    # Thread path: the first seeded thread's link is on the list page.
    html = opener.open(base + "/messages").read().decode()
    marker = 'hx-get="/partials/messages/'
    idx = html.find(marker)
    thread_path = None
    if idx >= 0:
        start = idx + len('hx-get="/partials')
        thread_path = html[start : html.find('"', start)]
    # The minted session cookie rides back to the caller: the browser
    # gets it INJECTED (driver.add_cookie) instead of replaying the
    # island's login — the island gates its panel on the SIP WebSocket,
    # which a bare boot cannot serve (caddy bridges /sip on the stack).
    session = ""
    for cookie in cookie_processor.cookiejar:
        if cookie.name == "webphone_session":
            session = cookie.value
    return thread_path, session


def capture(base: str, out_dir: str, thread_path: str | None, session: str) -> int:
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
        WebDriverWait(driver, 10).until(lambda d: d.execute_script(
            "return document.readyState") == "complete")
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
                driver.get(base + path)
                time.sleep(0.4)  # settle: relative times, panels
                marker = SURFACE_MARKERS[name]
                if not driver.find_elements("css selector", marker):
                    raise AssertionError(
                        f"surface {name!r} at {path} is missing {marker!r} — "
                        "the shot would not be evidence; aborting")
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
            except Exception:
                time.sleep(0.2)
        else:
            print("server did not come up", file=sys.stderr)
            return 1
        thread_path, session = seed(base)
        if not session:
            print("seed did not mint a session cookie", file=sys.stderr)
            return 1
        shots = capture(base, args.out_dir, thread_path, session)
        print(f"ui-capture: {shots} shots in {args.out_dir}/")
        return 0 if shots >= 14 else 1
    finally:
        server.terminate()
        server.wait(timeout=10)


if __name__ == "__main__":
    raise SystemExit(main())
