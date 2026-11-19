#!/usr/bin/env python3
"""Fully automated WORKFLOW frontend E2E: a real headless Chromium drives
the real UI over a self-booted webphone binary.

Distinct from the two existing browser-touching tools: webphone-smoke.py
is HTTP black-box (no DOM, no JS), and ui-capture.py shoots SURFACES for
human eyes (local-only by budget decision). This harness CLICKS: the
login front door form, tab navigation without reload, the composer send,
a LIVE inbound webhook appearing through SSE->morph in the open page,
thread flags, snippet fills, contacts CRUD through the real forms, the
language cookie, the styled 404, and the command palette. DOM assertions
only — no screenshots, no visual budget.

Run inside the browser shell (the PYSEL dance):

    nix shell nixpkgs#chromium nixpkgs#chromedriver \\
      nixpkgs#python312.withPackages(ps: [ ps.selenium ]) \\
      --command python3 scripts/workflow-e2e.py [--bin /tmp/webphone-bin]

Verdict lines are printed per journey and tallied ("workflow e2e: N
passed, M failed"); counts are always the tally, never hardcoded.
"""

from __future__ import annotations

import argparse
import http.cookiejar
import json
import os
import re
import shutil
import socket
import subprocess  # nosec B404 - deliberately spawns the local webphone binary under test
import sys
import tempfile
import time
import urllib.error
import urllib.request

TIMEOUT = 10


class Check:
    def __init__(self) -> None:
        self.failures: list[str] = []
        self.passed = 0

    def ok(self, name: str, condition: bool, detail: str = "") -> bool:
        if condition:
            self.passed += 1
            print(f"  [ok]   {name}")
        else:
            self.failures.append(f"{name}: {detail}")
            print(f"  [FAIL] {name}: {detail}")
        return condition


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def multipart_body(fields: list[tuple[str, str]]) -> tuple[bytes, str]:
    boundary = "----webphoneworkflowboundary"
    parts = []
    for name, value in fields:
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="{name}"\r\n\r\n'
                f"{value}\r\n"
            ).encode()
        )
    parts.append(f"--{boundary}--\r\n".encode())
    return b"".join(parts), f"multipart/form-data; boundary={boundary}"


def boot(
    binary: str, workdir: str, config: dict | None = None, env_extra: dict | None = None
) -> tuple[subprocess.Popen, str, str, Callable[[], None]]:  # type: ignore[name-defined]
    """Boot the binary on a fresh port + datadir; returns (proc, base,
    datadir, stop). config (maps the env cannot carry: auth.users,
    identities) is written to a JSON file pointed at via WEBPHONE_CONFIG;
    env_extra rides the environment verbatim."""
    port = free_port()
    datadir = f"{workdir}/data-{port}"
    env = dict(os.environ)
    env.update(
        {
            "WEBPHONE_ADDR": f"127.0.0.1:{port}",
            "WEBPHONE_DATA_DIR": datadir,
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted test server
        }
    )
    if config is not None:
        config_path = f"{workdir}/config-{port}.json"
        with open(config_path, "w", encoding="utf-8") as handle:
            json.dump(config, handle)
        env["WEBPHONE_CONFIG"] = config_path
    if env_extra:
        env.update(env_extra)
    proc = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
        [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
    )
    base = f"http://127.0.0.1:{port}"
    deadline = time.monotonic() + TIMEOUT * 2
    while time.monotonic() < deadline:
        try:
            urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
            break
        except OSError:
            time.sleep(0.1)
    else:
        proc.kill()
        raise RuntimeError("workflow e2e: server did not become ready")

    def stop() -> None:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()

    return proc, base, datadir, stop


class Seed:
    """HTTP-side seeding: a session via the login API, threads via the
    composer + inbound hooks, one snippet, one contact. Returns the
    session cookie for the browser to inherit."""

    def __init__(self, base: str) -> None:
        self.base = base
        self.jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.jar)
        )
        self.csrf = ""

    def request(
        self, method: str, path: str, body: bytes | None = None, ctype: str = ""
    ) -> tuple[int, bytes]:
        req = urllib.request.Request(self.base + path, data=body, method=method)
        if self.csrf and method not in ("GET", "HEAD"):
            req.add_header("X-CSRF-Token", self.csrf)
        if ctype:
            req.add_header("Content-Type", ctype)
        try:
            with self.opener.open(req, timeout=TIMEOUT) as resp:
                return resp.status, resp.read()
        except urllib.error.HTTPError as err:
            return err.code, err.read()

    def login(self) -> None:
        _, page = self.request("GET", "/login")
        match = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if match:
            self.csrf = match.group(1)
        status, _ = self.request(
            "POST",
            "/api/session",
            json.dumps({"extension": "1001", "password": "pw"}).encode(),
            "application/json",
        )
        if status != 201:
            raise RuntimeError(f"seed login: {status}")
        _, body = self.request("GET", "/api/csrf")
        self.csrf = json.loads(body)["token"]

    def send(self, to: str, body: str) -> None:
        form, ctype = multipart_body([("to", to), ("body", body)])
        status, _ = self.request("POST", "/messages/send", form, ctype)
        if status != 200:
            raise RuntimeError(f"seed send to {to}: {status}")

    def hook_inbound(self, sender: str, body: str) -> None:
        payload = json.dumps({"owner": "1001", "from": sender, "body": body}).encode()
        req = urllib.request.Request(
            self.base + "/hooks/message", data=payload, method="POST"
        )
        req.add_header("Content-Type", "application/json")
        req.add_header("Authorization", "Bearer test-secret")  # nosec B105 - fixture secret for the self-booted test server
        with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:  # nosec B310 - webhook call against the local server under test
            if resp.status != 202:
                raise RuntimeError(f"seed hook: {resp.status}")

    def save_snippet(self, body: str) -> None:
        from urllib.parse import urlencode

        status, _ = self.request(
            "POST",
            "/snippets/save",
            urlencode({"body": body, "quick": "1"}).encode(),
            "application/x-www-form-urlencoded",
        )
        if status != 200:
            raise RuntimeError(f"seed snippet: {status}")

    def save_contact(self, name: str, number: str) -> None:
        from urllib.parse import urlencode

        status, _ = self.request(
            "POST",
            "/contacts/save",
            urlencode({"name": name, "number": number}).encode(),
            "application/x-www-form-urlencoded",
        )
        if status != 200:
            raise RuntimeError(f"seed contact: {status}")

    def session_cookie(self) -> str:
        for cookie in self.jar:
            if cookie.name == "webphone_session":
                return cookie.value
        raise RuntimeError("no session cookie after seed login")


def make_driver():
    from selenium import webdriver

    options = webdriver.ChromeOptions()
    for candidate in ("chromium", "chromium-browser", "google-chrome"):
        found = shutil.which(candidate)
        if found:
            options.binary_location = found
            break
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--window-size=1280,900")
    # The passkey ceremonies need a DOMAIN origin (WebAuthn rejects IP
    # rp_ids): map webphone.test onto the loopback server inside the
    # browser only — no /etc/hosts, no TLS front needed for the journey.
    options.add_argument("--host-resolver-rules=MAP webphone.test 127.0.0.1")
    driver_path = shutil.which("chromedriver")
    service = (
        webdriver.ChromeService(executable_path=driver_path) if driver_path else None
    )
    return webdriver.Chrome(options=options, service=service)


def wait(driver, predicate, timeout: float = TIMEOUT):
    from selenium.webdriver.support.ui import WebDriverWait

    return WebDriverWait(driver, timeout).until(predicate)


def journey_login_front_door(driver, base: str, c: Check) -> None:
    driver.get(base + "/login")
    wait(driver, lambda d: d.find_elements("css selector", "#login-form"))
    # With the identity layer on, the extension form is the break-glass
    # lifeline one disclosure away — open it before filling.
    disclosures = driver.find_elements("css selector", "details > summary")
    if disclosures:
        disclosures[0].click()
    driver.find_element("css selector", "#ext").send_keys("1001")
    driver.find_element("css selector", "#pass").send_keys("pw")
    driver.find_element("css selector", "#login-btn").click()
    wait(driver, lambda d: d.current_url.rstrip("/") == base)
    c.ok(
        "J1 login lands on the shell",
        bool(driver.find_elements("css selector", "#reg-status")),
        f"url={driver.current_url}",
    )
    c.ok(
        "J1 nav renders",
        bool(driver.find_elements("css selector", "a.wp-nav-link")),
        "no nav links",
    )


TABS = [
    ("/messages", "messages-panel"),
    ("/fax", "fax-panel"),
    ("/voicemail", "voicemail-panel"),
    ("/history", "history-panel"),
    ("/contacts", "contacts-panel"),
    ("/settings", "settings-panel"),
]


def journey_tab_navigation(driver, base: str, c: Check) -> None:
    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "a.wp-nav-link"))
    driver.execute_script("window.__e2eMarker = 42;")
    for path, panel_id in TABS:
        link = wait(
            driver,
            lambda d, p=path: [
                el
                for el in d.find_elements("css selector", "a.wp-nav-link")
                if el.get_attribute("href").endswith(p)
            ],
        )[0]
        link.click()
        wait(driver, lambda d, pid=panel_id: d.find_elements("css selector", f"#{pid}"))
        wait(driver, lambda d, p=path: d.current_url.endswith(p))
        current = driver.execute_script(
            "return document.querySelector('a[aria-current=page]')?.getAttribute('href')"
        )
        c.ok(
            f"J2 {path} panel + aria-current + url",
            driver.current_url.endswith(path) and (current or "").endswith(path),
            f"url={driver.current_url} current={current}",
        )
    marker = driver.execute_script("return window.__e2eMarker")
    c.ok("J2 no reload across tabs", marker == 42, f"marker={marker}")


def _open_first_thread(driver, base: str) -> str:
    driver.get(base + "/messages")
    link = wait(
        driver,
        lambda d: d.find_elements("css selector", 'a[href^="/messages/"]'),
    )
    link = [
        el
        for el in driver.find_elements("css selector", 'a[href^="/messages/"]')
        if "/messages/" in el.get_attribute("href")
    ][0]
    href = link.get_attribute("href")
    link.click()
    wait(driver, lambda d: d.find_elements("css selector", "#thread-transcript"))
    return href


def journey_composer_send(driver, base: str, c: Check, body: str) -> None:
    _open_first_thread(driver, base)
    before = len(driver.find_elements("css selector", "#thread-transcript .wp-bubble"))
    area = wait(
        driver,
        lambda d: d.find_elements(
            "css selector", "form.wp-compose textarea[name=body]"
        ),
    )[0]
    area.send_keys(body)
    driver.find_element("css selector", "form.wp-compose button[type=submit]").click()
    try:
        wait(
            driver,
            lambda d: (
                len(d.find_elements("css selector", "#thread-transcript .wp-bubble"))
                > before
            ),
        )
        c.ok("J3 composer send lands the bubble", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J3 composer send lands the bubble", False, str(err)[:120])


def journey_live_inbound(driver, base: str, c: Check, sender: str, body: str) -> None:
    from selenium.common.exceptions import StaleElementReferenceException

    driver.get(base + "/messages")
    wait(driver, lambda d: d.find_elements("css selector", 'a[href^="/messages/"]'))
    driver.execute_script("window.__e2eMarker = 7;")
    seed = Seed(base)
    seed.hook_inbound(sender, body)

    def sender_row(d):
        try:
            return [
                el
                for el in d.find_elements("css selector", 'a[href^="/messages/"]')
                if sender in el.get_attribute("href") or sender in el.text
            ]
        except StaleElementReferenceException:
            return []  # the SSE morph swapped the rail mid-poll; the next poll re-finds

    try:
        wait(driver, sender_row)
        c.ok("J4 live inbound appears via SSE", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J4 live inbound appears via SSE", False, str(err)[:120])
    marker = driver.execute_script("return window.__e2eMarker")
    c.ok("J4 live inbound never reloaded", marker == 7, f"marker={marker}")


def journey_thread_pin(driver, base: str, c: Check) -> None:
    driver.get(base + "/messages")
    buttons = wait(
        driver,
        lambda d: d.find_elements("css selector", "button[hx-post*='/pin']"),
    )
    if not buttons:
        c.ok("J5 pin toggles the row state", False, "no pin button rendered")
        return
    buttons[0].click()
    try:
        wait(
            driver,
            lambda d: d.find_elements("css selector", ".wp-thread-row.wp-pinned"),
        )
        c.ok("J5 pin toggles the row state", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J5 pin toggles the row state", False, str(err)[:120])


def journey_snippet_fill(driver, base: str, c: Check, snippet: str) -> None:
    _open_first_thread(driver, base)
    chips = wait(
        driver,
        lambda d: d.find_elements("css selector", "button[data-snippet]"),
    )
    if not chips:
        c.ok("J6 snippet chip fills the composer", False, "no snippet chip rendered")
        return
    chips[0].click()
    value = driver.find_element(
        "css selector", "form.wp-compose textarea[name=body]"
    ).get_attribute("value")
    c.ok("J6 snippet chip fills the composer", value == snippet, f"value={value!r}")


def journey_contacts(driver, base: str, c: Check, name: str, number: str) -> None:
    from selenium.common.exceptions import StaleElementReferenceException

    def rows_named(d, want: bool):
        try:
            found = [
                el
                for el in d.find_elements("css selector", ".wp-thread-remote")
                if el.text == name
            ]
            return found if want else (not found)
        except StaleElementReferenceException:
            return [] if want else True  # mid-morph; the next poll re-finds

    driver.get(base + "/contacts")
    wait(driver, lambda d: d.find_elements("css selector", "#contacts-panel"))
    name_input = wait(
        driver,
        lambda d: d.find_elements(
            "css selector", "form input[name=name], form input[name='contact-name']"
        ),
    )[0]
    number_input = driver.find_element(
        "css selector", "form input[name=number], form input[name='contact-number']"
    )
    name_input.send_keys(name)
    number_input.send_keys(number)
    driver.find_element("css selector", "form button[type=submit]").click()
    try:
        wait(driver, lambda d: rows_named(d, True))
        c.ok("J7 contact add lands the row", True)
        delete = wait(
            driver,
            lambda d: d.find_elements(
                "css selector", "button[hx-post*='/contacts/delete']"
            ),
        )[0]
        delete.click()
        driver.switch_to.alert.accept()
        wait(driver, lambda d: rows_named(d, False))
        c.ok("J7 contact delete removes the row", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J7 contact add lands the row", False, str(err)[:120])


def journey_language_cookie(driver, base: str, c: Check) -> None:
    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "a.wp-nav-link"))
    driver.add_cookie({"name": "wp-lang", "value": "de", "path": "/"})
    driver.refresh()
    try:
        wait(
            driver,
            lambda d: [
                el
                for el in d.find_elements("css selector", "a.wp-nav-link")
                if "Nachrichten" in el.text
            ],
        )
        c.ok("J8 wp-lang=de relabels the nav", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J8 wp-lang=de relabels the nav", False, str(err)[:120])


def journey_styled_404(driver, base: str, c: Check) -> None:
    driver.get(base + "/definitely/not/a/workflow/path")
    try:
        wait(driver, lambda d: d.find_elements("css selector", ".wp-notfound"))
        panel = driver.find_element("css selector", ".wp-notfound")
        c.ok(
            "J9 styled 404 with a way home",
            "/definitely/not/a/workflow/path" in panel.text
            and bool(panel.find_elements("css selector", "a.wp-notfound-back")),
            panel.text[:80],
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J9 styled 404 with a way home", False, str(err)[:120])


def journey_palette(driver, base: str, c: Check) -> None:
    from selenium.webdriver.common.keys import Keys

    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "a.wp-nav-link"))
    driver.find_element("css selector", "body").send_keys(Keys.CONTROL, "k")
    try:
        wait(driver, lambda d: d.find_elements("css selector", "#wp-palette"))
        c.ok("J10 Ctrl+K opens the palette", True)
        box = driver.find_element("css selector", "#wp-palette input")
        box.send_keys("fax")
        time.sleep(0.2)
        box.send_keys(Keys.ENTER)
        wait(
            driver,
            lambda d: (
                d.current_url.endswith("/fax")
                and d.find_elements("css selector", "#fax-panel")
            ),
        )
        c.ok("J10 palette enter navigates", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J10 Ctrl+K opens the palette", False, str(err)[:120])


def wav_silence(seconds: float = 0.6, rate: int = 8000) -> bytes:
    """A minimal valid WAV (PCM16 mono) of silence — enough for
    Chromium to keep an <audio> element playing."""
    import struct

    frames = int(rate * seconds)
    data = b"\x00\x00" * frames
    header = struct.pack(
        "<4sI4s4sIHHIIHH4sI",
        b"RIFF",
        36 + len(data),
        b"WAVE",
        b"fmt ",
        16,
        1,
        1,
        rate,
        rate * 2,
        2,
        16,
        b"data",
        len(data),
    )
    return header + data


class FakePhoneAPI:
    """The loopback stand-in for the telephony stack's phone API: the
    voicemail summary/messages rows, DELETE, the message recordings
    (one live WAV, one dead 404 — the honest-toast-once contract), and
    an empty CDR history. The webphone server directory-verifies logins
    against GET summary, so this must be up before boot."""

    def __init__(self) -> None:
        import threading
        from http.server import BaseHTTPRequestHandler, HTTPServer

        self.messages: list[dict] = [
            {
                "uuid": "e2e-live-uuid",
                "cid_number": "+441632960901",
                "cid_name": "Live Recording",
                "seconds": 3,
                "created": int(time.time()),
                "read": False,
                "audio_url": "",
            },
            {
                "uuid": "e2e-dead-uuid",
                "cid_number": "+441632960902",
                "cid_name": "Dead Recording",
                "seconds": 4,
                "created": int(time.time()) - 60,
                "read": True,
                "audio_url": "",
            },
        ]
        outer = self
        wav = wav_silence()

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, fmt, *args):  # silence the default stderr spam
                return

            def _json(self, code: int, payload) -> None:
                body = json.dumps(payload).encode()
                self.send_response(code)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                if code != 204:
                    self.wfile.write(body)

            def do_GET(self):  # noqa: N802 - http.server API
                if self.path.endswith("/summary"):
                    fresh = sum(1 for m in outer.messages if not m["read"])
                    old = sum(1 for m in outer.messages if m["read"])
                    self._json(200, {"new": fresh, "old": old})
                    return
                if "/messages" in self.path:
                    self._json(200, {"messages": outer.messages})
                    return
                if self.path.startswith("/vm/"):
                    if self.path == "/vm/e2e-live-uuid.wav":
                        self.send_response(200)
                        self.send_header("Content-Type", "audio/wav")
                        self.send_header("Content-Length", str(len(wav)))
                        self.end_headers()
                        self.wfile.write(wav)
                    else:
                        self.send_response(404)
                        self.send_header("Content-Length", "0")
                        self.end_headers()
                    return
                if self.path.startswith("/history"):
                    self._json(200, {"entries": []})
                    return
                self._json(404, {"error": "not found"})

            def do_DELETE(self):  # noqa: N802 - http.server API
                uuid = self.path.rstrip("/").rsplit("/", 1)[-1]
                before = len(outer.messages)
                outer.messages = [m for m in outer.messages if m["uuid"] != uuid]
                self._json(204 if len(outer.messages) < before else 404, {})

        self._server = HTTPServer(("127.0.0.1", 0), Handler)
        self.port = self._server.server_address[1]
        self.base = f"http://127.0.0.1:{self.port}"
        for message in self.messages:
            suffix = ".wav" if message["uuid"] == "e2e-live-uuid" else "/missing.wav"
            message["audio_url"] = f"{self.base}/vm/{message['uuid']}{suffix}"
        self._thread = threading.Thread(target=self._server.serve_forever, daemon=True)

    def start(self) -> "FakePhoneAPI":
        self._thread.start()
        return self

    def stop(self) -> None:
        self._server.shutdown()
        self._server.server_close()


def _first_thread_id(driver, base: str) -> str:
    driver.get(base + "/messages")
    link = wait(
        driver,
        lambda d: d.find_elements("css selector", 'a[href^="/messages/"]'),
    )[0]
    return link.get_attribute("href").rstrip("/").rsplit("/", 1)[-1]


def journey_admin(driver, base: str, c: Check, thread_id: str) -> None:
    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "a.wp-nav-link"))
    admin_links = [
        el
        for el in driver.find_elements("css selector", "a.wp-nav-link")
        if el.get_attribute("href").endswith("/admin")
    ]
    c.ok("J-admin nav link renders for the mapped admin", bool(admin_links))
    if not admin_links:
        return
    admin_links[0].click()
    wait(driver, lambda d: d.find_elements("css selector", "#admin-panel"))
    c.ok("J-admin /admin panel renders", True)
    c.ok(
        "J-admin thread inventory lists the seeded thread",
        bool(driver.find_elements("css selector", "a.wp-admin-thread-row")),
    )
    driver.get(base + f"/admin/threads/{thread_id}")
    wait(driver, lambda d: d.find_elements("css selector", "#admin-thread"))
    c.ok("J-admin thread deep link renders", True)


def journey_line_switcher(driver, base: str, c: Check) -> None:
    driver.get(base + "/")
    select = wait(
        driver, lambda d: d.find_elements("css selector", "#wp-ext-switch"), timeout=15
    )[0]
    options = select.find_elements("css selector", "option")
    labels = [opt.text.strip() for opt in options]
    c.ok(
        "J-switch renders one option per owned line",
        len(options) == 2 and all("·" in label for label in labels),
        f"labels={labels}",
    )
    current = select.get_attribute("value")
    other = [opt.get_attribute("value") for opt in options if opt.get_attribute("value") != current][0]
    try:
        from selenium.webdriver.support.select import Select

        Select(select).select_by_value(other)
        wait(
            driver,
            lambda d: len(d.find_elements("css selector", "#toasts .toast")) > 0,
        )
        toast = driver.find_elements("css selector", "#toasts .toast")[-1].text
        c.ok(
            "J-switch extension login denied with a visible announce",
            "switch" in toast.lower() or "403" in toast,
            f"toast={toast[:80]}",
        )
        c.ok(
            "J-switch select resets to the active line",
            select.get_attribute("value") == current,
            f"value={select.get_attribute('value')}",
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-switch deny path announces and resets", False, str(err)[:120])


def journey_dial_gate(driver, base: str, c: Check, seed: "Seed", remote: str) -> None:
    def gate(extension_override: bool) -> tuple[int, bytes]:
        body = json.dumps(
            {"remote": remote, "override": extension_override}
        ).encode()
        return seed.request("POST", "/api/dials", body, "application/json")

    status, first = gate(False)
    _, second = gate(False)
    first_verdict = json.loads(first)
    second_verdict = json.loads(second)
    c.ok(
        "J-dial gate records then denies the same remote",
        status == 200
        and first_verdict.get("allowed") is True
        and second_verdict.get("allowed") is False,
        f"first={first_verdict} second={second_verdict}",
    )

    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "#dial-form"))
    dest = driver.find_element("css selector", "#dest")
    dest.clear()
    dest.send_keys(remote)
    driver.find_element("css selector", "#dial-form button[type=submit]").click()
    try:
        chip = wait(
            driver,
            lambda d: d.find_elements("css selector", "#wp-dial-cooldown"),
            timeout=15,
        )[0]
        buttons = chip.find_elements("css selector", "button")
        c.ok(
            "J-dial denied dial renders the inline state + Dial anyway chip",
            bool(buttons),
            f"chip-buttons={len(buttons)}",
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-dial denied dial renders the inline state chip", False, str(err)[:120])


def journey_voicemail(driver, base: str, c: Check) -> None:
    driver.get(base + "/voicemail")
    rows = wait(driver, lambda d: d.find_elements("css selector", "[data-vm-play]"))
    c.ok("J-vm rows render from the phone API", len(rows) == 2, f"rows={len(rows)}")

    live_btn = driver.find_element("css selector", "[data-vm-play='e2e-live-uuid']")
    live_btn.click()
    try:
        wait(
            driver,
            lambda d: d.execute_script(
                "var a = document.getElementById('vm-audio-e2e-live-uuid');"
                " return !!a && !a.paused && a.currentTime > 0;"
            ),
            timeout=15,
        )
        c.ok("J-vm live recording plays", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-vm live recording plays", False, str(err)[:120])

    dead_btn = driver.find_element("css selector", "[data-vm-play='e2e-dead-uuid']")
    dead_btn.click()
    try:
        wait(
            driver,
            lambda d: len(d.find_elements("css selector", "#toasts .toast")) == 1,
            timeout=15,
        )
        c.ok("J-vm dead recording toasts exactly once", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        toasts = len(driver.find_elements("css selector", "#toasts .toast"))
        c.ok("J-vm dead recording toasts exactly once", False, f"toasts={toasts} {str(err)[:80]}")

    for uuid, label in (("e2e-live-uuid", "live"), ("e2e-dead-uuid", "dead")):
        try:
            row_button = driver.find_element(
                "css selector", f"[hx-post*='uuid={uuid}']"
            )
            row_button.click()
            alert = wait(driver, lambda d: d.switch_to.alert)
            alert.accept()
            wait(
                driver,
                lambda d, u=uuid: not d.find_elements(
                    "css selector", f"[data-vm-play='{u}']"
                ),
            )
            c.ok(f"J-vm delete removes the {label} row", True)
        except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
            c.ok(f"J-vm delete removes the {label} row", False, str(err)[:120])
    try:
        wait(
            driver,
            lambda d: d.find_elements("css selector", "#voicemail-panel .wp-empty-state"),
        )
        c.ok("J-vm empty state after the last delete", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-vm empty state after the last delete", False, str(err)[:120])


def self_test() -> int:
    failures = 0

    def check(name: str, ok: bool) -> None:
        nonlocal failures
        print(f"  [{'ok' if ok else 'FAIL'}] {name}")
        if not ok:
            failures += 1

    body, ctype = multipart_body([("to", "+123"), ("body", "hi")])
    check(
        "multipart shape",
        ctype.startswith("multipart/form-data; boundary=")
        and b'name="to"\r\n\r\n+123' in body,
    )
    check("free port binds", isinstance(free_port(), int))
    print(f"self-test: {2 - failures} passed, {failures} failed")
    return 1 if failures else 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--bin", help="webphone binary to boot (default: go build)")
    parser.add_argument("--base", help="drive an already-running server")
    parser.add_argument("--self-test", action="store_true", help="pure helpers only")
    parser.add_argument("--keep", action="store_true", help="keep the datadir")
    args = parser.parse_args()

    if args.self_test:
        return self_test()

    binary = args.bin
    workdir = tempfile.mkdtemp(prefix="webphone-workflow-e2e-")
    stop_server: Callable[[], None] | None = None  # type: ignore[name-defined]
    phone_api: FakePhoneAPI | None = None
    base = args.base
    if base:
        print(f"workflow e2e against {base}")
    else:
        if not binary:
            binary = f"{workdir}/webphone-bin"
            print(f"building {binary} ...")
            build = subprocess.run(  # nosec B603 - argv-list (no shell) build of the binary under test
                ["go", "build", "-o", binary, "./cmd/webphone"],
                capture_output=True,
                text=True,
                check=False,
            )
            if build.returncode != 0:
                print(f"build failed: {build.stderr[:400]}", file=sys.stderr)
                return 2
        # The fake phone API must be up FIRST: the server
        # directory-verifies every login against its summary route.
        phone_api = FakePhoneAPI().start()
        # auth.users folds into passkey mode (the ONE identity home), so
        # the config carries the full passkey half: rp pair + one SIP
        # password file per mapped extension (read per LOGIN, never at
        # boot — fixture paths are enough for the journeys).
        for ext in ("1001", "1002"):
            with open(f"{workdir}/sip-password-{ext}", "w", encoding="utf-8") as handle:
                handle.write("pw")  # nosec B105 - fixture credential for the self-booted test server
        _, base, datadir, stop_server = boot(
            binary,
            workdir,
            config={
                "auth": {
                    "users": {
                        "ops@example.org": {
                            "display_name": "Ops User",
                            "extensions": ["1001", "1002"],
                            "admin": True,
                        }
                    },
                    "extension_password_files": {
                        "1001": f"{workdir}/sip-password-1001",
                        "1002": f"{workdir}/sip-password-1002",
                    },
                    "passkey": {
                        "rp_id": "webphone.test",
                        "rp_origins": ["https://webphone.test"],
                    },
                },
                "identities": {
                    "1001": "+49 30 1110001",
                    "1002": "+49 30 1110002",
                },
            },
            env_extra={
                "WEBPHONE_DIAL__COOLDOWN": "2m",
                "WEBPHONE_PHONE_API_URL": phone_api.base,
            },
        )
        print(f"workflow e2e against {base} (datadir {datadir})")
    c = Check()
    seed = Seed(base)
    try:
        seed.login()
        seed.send("+441632960801", "workflow seed outbound")
        seed.hook_inbound("+441632960802", "workflow seed inbound")
        seed.save_snippet("workflow snippet body")
        seed.save_contact("Workflow Seed", "+441632960803")

        driver = make_driver()
        try:
            # J1 logs the BROWSER in through the real front door; the seed
            # session (HTTP-side) only created the data — same extension,
            # same store, so the browser session sees every seeded row.
            journey_login_front_door(driver, base, c)
            journey_tab_navigation(driver, base, c)
            journey_composer_send(driver, base, c, "workflow browser send")
            journey_live_inbound(
                driver, base, c, "+441632960804", "workflow live inbound"
            )
            journey_thread_pin(driver, base, c)
            journey_snippet_fill(driver, base, c, "workflow snippet body")
            journey_contacts(driver, base, c, "Browser Contact", "+441632960805")
            journey_palette(driver, base, c)
            journey_styled_404(driver, base, c)
            journey_language_cookie(driver, base, c)
            # The ownership-model journeys (config-file boot: the admin
            # flag, two owned lines with DIDs) and the dial-gate +
            # voicemail lanes (env + fake phone API) need THIS harness's
            # boot config — an externally driven --base server owes them
            # nothing, so they run only on the self-booted lane.
            if stop_server is not None:
                journey_line_switcher(driver, base, c)
                thread_id = _first_thread_id(driver, base)
                journey_admin(driver, base, c, thread_id)
                journey_dial_gate(driver, base, c, seed, "+441632960999")
                journey_voicemail(driver, base, c)
        finally:
            driver.quit()
    finally:
        if phone_api is not None:
            phone_api.stop()
        if stop_server is not None:
            stop_server()
        if not args.keep:
            shutil.rmtree(workdir, ignore_errors=True)

    print(f"workflow e2e: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


if __name__ == "__main__":
    sys.exit(main())
