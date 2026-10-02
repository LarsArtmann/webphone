#!/usr/bin/env python3
"""Visual capture harness (T23): boots a fresh webphone binary (loopback
gateway — the whole product, zero PBX), seeds deterministic content over
HTTP, then drives headless chromium (selenium — the SAME driver family
the consuming stack's browser E2E uses) through the six server tabs and
screenshots each in light and dark: the 12-shot matrix.

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

Output: ui-shots/<n>-<name>-<light|dark>.png (12 files).
"""

from __future__ import annotations

import argparse
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


def seed(base: str) -> str | None:
    """Deterministic content via the loopback gateway; returns thread path."""
    csrf = ""
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor())

    def post(path: str, fields: dict[str, str]) -> None:
        data = urllib.parse.urlencode(fields).encode()
        req = urllib.request.Request(base + path, data=data, method="POST")
        req.add_header("X-CSRF-Token", csrf)
        opener.open(req)

    # Login (loopback mode accepts any extension + skips PBX verify).
    body = urllib.parse.urlencode({"extension": "1001", "password": "pw"}).encode()
    opener.open(urllib.request.Request(base + "/login", data=body, method="POST"))
    resp = opener.open(base + "/api/csrf")
    csrf = resp.read().decode().strip().strip('"')
    for remote, _, message in SEED_THREADS:
        post("/messages/send", {"to": remote, "body": message})
    for text, quick in SEED_SNIPPETS:
        post("/snippets/save", {"body": text, **({"quick": "1"} if quick else {})})
    for name, number in SEED_CONTACTS:
        post("/contacts/save", {"name": name, "number": number})
    # Thread path: the first seeded thread's link is on the list page.
    html = opener.open(base + "/messages").read().decode()
    marker = 'hx-get="/partials/messages/'
    idx = html.find(marker)
    if idx < 0:
        return None
    start = idx + len('hx-get="/partials')
    return html[start : html.find('"', start)]


def capture(base: str, out_dir: str, thread_path: str | None) -> int:
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.support.ui import WebDriverWait

    options = Options()
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--use-fake-device-for-media-stream")
    options.add_argument("--use-fake-ui-for-media-stream")
    options.add_argument("--window-size=1280,900")
    driver = webdriver.Chrome(options=options)
    shots = 0
    try:
        driver.get(base + "/")
        WebDriverWait(driver, 10).until(lambda d: d.execute_script(
            "return document.readyState") == "complete")
        # Session via the island's own login flow (the API + cookie mint,
        # then the shell reveals the tabs).
        driver.find_element("id", "ext").send_keys("1001")
        driver.find_element("id", "pass").send_keys("pw")
        driver.find_element("id", "login-form").submit()
        time.sleep(1.0)  # the island mints the session cookie async
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
    env = dict(
        os.environ,
        WEBPHONE_ADDR=f"127.0.0.1:{args.port}",
        WEBPHONE_DATA_DIR=data_dir,
        WEBPHONE_GATEWAY__WEBHOOK_SECRET="devsecret",
    )
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
        thread_path = seed(base)
        shots = capture(base, args.out_dir, thread_path)
        print(f"ui-capture: {shots} shots in {args.out_dir}/")
        return 0 if shots >= 12 else 1
    finally:
        server.terminate()
        server.wait(timeout=10)


if __name__ == "__main__":
    raise SystemExit(main())
