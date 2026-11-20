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
    binary: str,
    workdir: str,
    config: dict | None = None,
    env_extra: dict | None = None,
    port: int | None = None,
) -> tuple[subprocess.Popen, str, str, Callable[[], None]]:  # type: ignore[name-defined]
    """Boot the binary on a fresh port + datadir; returns (proc, base,
    datadir, stop). config (maps the env cannot carry: auth.users,
    identities) is written to a JSON file pointed at via WEBPHONE_CONFIG;
    env_extra rides the environment verbatim. port pins the listen port
    when the caller needs it BEFORE boot (the passkey rp_origins must
    carry the real origin, port included)."""
    if port is None:
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


def make_driver(insecure_origin: str = ""):
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
    if insecure_origin:
        # ...and a SECURE context: http://webphone.test:<port> is not one
        # by default, and navigator.credentials is undefined without it.
        # Grant it for exactly this origin (port included) — nothing else.
        options.add_argument(
            f"--unsafely-treat-insecure-origin-as-secure={insecure_origin}"
        )
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

            def _stripped_path(self) -> str:
                # The app prepends /phone-api to the configured phone API URL
                # (both the server-side client and the /phone-api proxy), so the
                # fake sees prefixed paths; normalize once at the handler door.
                path = self.path
                path = path.removeprefix("/phone-api")
                return path

            def do_GET(self):
                path = self._stripped_path()
                if path.endswith("/summary"):
                    fresh = sum(1 for m in outer.messages if not m["read"])
                    old = sum(1 for m in outer.messages if m["read"])
                    self._json(200, {"new": fresh, "old": old})
                    return
                if "/messages" in path:
                    self._json(200, {"messages": outer.messages})
                    return
                if path.startswith("/vm/"):
                    if path == "/vm/e2e-live-uuid.wav":
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
                if path.startswith("/history"):
                    self._json(200, {"entries": []})
                    return
                self._json(404, {"error": "not found"})

            def do_DELETE(self):
                uuid = self._stripped_path().rstrip("/").rsplit("/", 1)[-1]
                before = len(outer.messages)
                outer.messages = [m for m in outer.messages if m["uuid"] != uuid]
                self._json(204 if len(outer.messages) < before else 404, {})

        self._server = HTTPServer(("127.0.0.1", 0), Handler)
        self.port = self._server.server_address[1]
        self.base = f"http://127.0.0.1:{self.port}"
        # RELATIVE on purpose: the browser resolves it against the page origin,
        # so playback rides the session-gated same-origin /phone-api proxy
        # (CSP-clean media; the proxy forwards Range and copies 206 back).
        for message in self.messages:
            suffix = ".wav" if message["uuid"] == "e2e-live-uuid" else "/missing.wav"
            message["audio_url"] = f"/phone-api/vm/{message['uuid']}{suffix}"
        self._thread = threading.Thread(target=self._server.serve_forever, daemon=True)

    def start(self) -> FakePhoneAPI:
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


def mint_enroll_token(
    binary: str, config_path: str, datadir: str, addr: str, email: str
) -> str:
    """Mint a one-time enrollment token with the `-enroll-passkey` CLI
    one-shot, under the SAME config file + data dir the booted server
    holds (the shared-journal contract). Retried once: both processes
    open the SQLite journal, and a busy lock is a race, not a verdict."""
    os.makedirs(
        datadir, exist_ok=True
    )  # SQLite cannot open the enroll journal in a missing parent dir
    env = dict(os.environ)
    env.update(
        {
            "WEBPHONE_CONFIG": config_path,
            "WEBPHONE_DATA_DIR": datadir,
            "WEBPHONE_ADDR": addr,
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret, same shape the server booted with
        }
    )
    done = None
    for _ in (1, 2):
        done = subprocess.run(  # nosec B603 - argv-list (no shell) run of the CLI one-shot under test
            [binary, "-enroll-passkey", email],
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        match = re.search(r"token=([A-Za-z0-9_-]+)", done.stdout)
        if match:
            return match.group(1)
        time.sleep(1)
    raise RuntimeError(
        f"enroll token mint failed: rc={done.returncode} err={(done.stderr or '')[:160]}"
    )


def journey_passkey(driver, domain_base: str, c: Check, token: str, email: str) -> None:
    """The full passkey lane on the DOMAIN origin: CLI-minted token →
    the /enroll ceremony → the /login ceremony → a MAPPED session
    (LoginEmail set) — the substrate the admin + switcher journeys
    need, driven through chromedriver's virtual authenticator."""
    from selenium.webdriver.common.virtual_authenticator import (
        VirtualAuthenticatorOptions,
    )

    opts = VirtualAuthenticatorOptions()
    opts.protocol = VirtualAuthenticatorOptions.Protocol.CTAP2
    opts.transport = VirtualAuthenticatorOptions.Transport.INTERNAL
    opts.has_resident_key = True
    opts.is_user_verified = True
    driver.add_virtual_authenticator(opts)

    driver.get(domain_base + "/enroll?token=" + token)
    # The module prefills the token input at load — its presence with the
    # value is the proof the submit handler is bound (never click before).
    wait(
        driver,
        lambda d: (lambda els: bool(els) and els[0].get_attribute("value") == token)(
            d.find_elements("css selector", "#enroll-token")
        ),
        timeout=15,
    )
    try:
        driver.find_element("css selector", "#enroll-credential-name").send_keys("e2e")
        driver.find_element("css selector", "#enroll-form button[type=submit]").click()
        wait(
            driver,
            lambda d: d.find_elements("css selector", "#enroll-form[hidden]"),
            timeout=20,
        )
        c.ok("J-passkey enroll ceremony completes", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        error = driver.find_elements("css selector", "#enroll-error:not([hidden])")
        detail = error[0].text[:80] if error else ""
        c.ok(
            "J-passkey enroll ceremony completes",
            False,
            f"{str(err)[:100]} {detail}".strip(),
        )

    driver.get(domain_base + "/login")
    wait(driver, lambda d: d.find_elements("css selector", "#lp-login-form"))
    try:
        driver.find_element("css selector", "#lp-email").send_keys(email)
        driver.find_element("css selector", "#lp-login-btn").click()
        wait(
            driver,
            lambda d: (
                d.current_url.rstrip("/") == domain_base
                and bool(d.find_elements("css selector", "#reg-status"))
            ),
            timeout=20,
        )
        c.ok("J-passkey login lands on the shell", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        error = driver.find_elements("css selector", "#lp-error:not([hidden])")
        detail = error[0].text[:80] if error else ""
        c.ok(
            "J-passkey login lands on the shell",
            False,
            f"{str(err)[:100]} {detail}".strip(),
        )
    try:
        who = driver.find_element("css selector", "#whoami-ext").text
        c.ok(
            "J-passkey whoami names the mapped user on the first line",
            "Ops User" in who and "+49 30 1110001" in who,
            f"whoami={who[:80]}",
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-passkey whoami names the mapped user", False, str(err)[:120])


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
    """The DESIGNED success path: a mapped multi-line session switches
    to its second owned line. The island POSTs /api/session/switch and
    RELOADS through the one boot path; the re-rendered switcher + whoami
    carry the new active line (the extension-login 403-deny shape is a
    different session class this harness no longer drives here)."""
    driver.get(base + "/")
    select = wait(
        driver, lambda d: d.find_elements("css selector", "#wp-ext-switch"), timeout=15
    )[0]
    options = select.find_elements("css selector", "option")
    labels = [opt.text.strip() for opt in options]
    c.ok(
        "J-switch renders one option per owned line (DID · extension)",
        len(options) == 2 and all("·" in label for label in labels),
        f"labels={labels}",
    )
    current = select.get_attribute("value")
    c.ok(
        "J-switch active line is the first owned line",
        current == "1001",
        f"value={current}",
    )
    if current != "1001":
        return
    try:
        from selenium.webdriver.support.select import Select

        Select(select).select_by_value("1002")

        def switched(d):
            try:
                els = d.find_elements("css selector", "#wp-ext-switch")
                return bool(els) and els[0].get_attribute("value") == "1002"
            except Exception:  # noqa: BLE001 - mid-reload staleness, the next poll re-finds
                return False

        wait(driver, switched, timeout=25)
        who = driver.find_element("css selector", "#whoami-ext").text
        c.ok(
            "J-switch reload lands with whoami on the second line",
            "+49 30 1110002" in who,
            f"whoami={who[:80]}",
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok(
            "J-switch reload lands with whoami on the second line",
            False,
            str(err)[:120],
        )
    try:
        wait(
            driver,
            lambda d: d.find_elements("css selector", "#wp-sse-live"),
            timeout=15,
        )
        c.ok("J-switch SSE resubscribes after the reload", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-switch SSE resubscribes after the reload", False, str(err)[:120])


def journey_dial_gate(driver, base: str, c: Check, seed: Seed, remote: str) -> None:
    def gate(extension_override: bool) -> tuple[int, bytes]:
        body = json.dumps({"remote": remote, "override": extension_override}).encode()
        return seed.request("POST", "/api/dials", body, "application/json")

    status, first = gate(False)
    _, second = gate(False)
    override_status, third = gate(True)
    first_verdict = json.loads(first)
    second_verdict = json.loads(second)
    third_verdict = json.loads(third)
    c.ok(
        "J-dial gate records, denies, then the override re-allows",
        status == 200
        and override_status == 200
        and first_verdict.get("allowed") is True
        and second_verdict.get("allowed") is False
        and third_verdict.get("allowed") is True,
        f"first={first_verdict} second={second_verdict} third={third_verdict}",
    )

    driver.get(base + "/")
    wait(driver, lambda d: d.find_elements("css selector", "#dial-form"))
    # dest.clear()/send_keys trips chromedriver interactability on this
    # island form; set the value + submit through the DOM instead — the
    # island's own listeners fire either way.
    driver.execute_script(
        "document.getElementById('dest').value = arguments[0];"
        "document.getElementById('dial-form').requestSubmit();",
        remote,
    )
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
    from selenium.common.exceptions import StaleElementReferenceException

    # The boot-time voicemail re-fetch morphs the panel shortly after
    # load: re-find + click with a stale-retry instead of holding a
    # reference across the swap.
    def click_retry(selector: str) -> None:
        for _ in range(8):
            try:
                driver.find_element("css selector", selector).click()
                return
            except StaleElementReferenceException:
                time.sleep(0.4)
        driver.find_element("css selector", selector).click()

    driver.get(base + "/voicemail")
    try:
        rows = wait(driver, lambda d: d.find_elements("css selector", "[data-vm-play]"))
        c.ok("J-vm rows render from the phone API", len(rows) == 2, f"rows={len(rows)}")
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-vm rows render from the phone API", False, str(err)[:120])
        return

    try:
        click_retry("[data-vm-play='e2e-live-uuid']")
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

    try:
        # Toasts live 6s and dedup identical consecutive messages: the
        # "exactly once" contract is the PEAK count over a full settle
        # window, one above the pre-click baseline and never two.
        before = len(driver.find_elements("css selector", "#toasts .toast"))
        click_retry("[data-vm-play='e2e-dead-uuid']")
        wait(
            driver,
            lambda d, b=before: (
                len(d.find_elements("css selector", "#toasts .toast")) > b
            ),
            timeout=15,
        )
        peak = before + 1
        settle = time.monotonic() + 7
        while time.monotonic() < settle:
            peak = max(
                peak, len(driver.find_elements("css selector", "#toasts .toast"))
            )
            time.sleep(0.2)
        c.ok(
            "J-vm dead recording toasts exactly once",
            peak == before + 1,
            f"peak={peak} before={before}",
        )
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J-vm dead recording toasts exactly once", False, str(err)[:120])

    # htmx asks hx-confirm through window.confirm at request time:
    # override it instead of racing switch_to.alert.
    driver.execute_script("window.confirm = function () { return true; };")
    for uuid, label in (("e2e-live-uuid", "live"), ("e2e-dead-uuid", "dead")):
        try:
            click_retry(f"[hx-post*='uuid={uuid}']")
            wait(
                driver,
                lambda d, u=uuid: (
                    not d.find_elements("css selector", f"[data-vm-play='{u}']")
                ),
                timeout=15,
            )
            c.ok(f"J-vm delete removes the {label} row", True)
        except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
            c.ok(f"J-vm delete removes the {label} row", False, str(err)[:120])
    try:
        wait(
            driver,
            lambda d: d.find_elements(
                "css selector", "#voicemail-panel .wp-empty-state"
            ),
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
    domain_base = ""
    port: int | None = None
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
        for ext in ("1001", "1002"):
            with open(f"{workdir}/sip-password-{ext}", "w", encoding="utf-8") as handle:
                handle.write("pw")  # nosec B105 - fixture credential for the self-booted test server
        # The port is pinned BEFORE boot: the passkey rp_origins must
        # carry the REAL origin (go-webauthn compares scheme+host WITH
        # the non-default port), and the browser needs the same origin
        # granted secure-context treatment for navigator.credentials.
        port = free_port()
        domain_base = f"http://webphone.test:{port}"
        # auth.users folds into passkey mode (the ONE identity home), so
        # the config carries the full passkey half: rp pair + one SIP
        # password file per mapped extension (read per LOGIN, never at
        # boot — fixture paths are enough for the journeys). The file is
        # written HERE because the enroll token must be minted BEFORE
        # boot: the one-shot opens the journal alone, the server joins
        # it after (the smoke's shared-journal ordering — minting against
        # a live server raced the verify 503).
        config = {
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
                    "rp_origins": [domain_base],
                },
            },
            "identities": {
                "1001": "+49 30 1110001",
                "1002": "+49 30 1110002",
            },
        }
        config_path = f"{workdir}/config-{port}.json"
        with open(config_path, "w", encoding="utf-8") as handle:
            json.dump(config, handle)
        token = ""
        try:
            token = mint_enroll_token(
                binary,
                config_path,
                f"{workdir}/data-{port}",
                f"127.0.0.1:{port}",
                "ops@example.org",
            )
        except Exception as err:  # noqa: BLE001 - pre-boot step, judged as a check below
            print(f"workflow e2e: enroll mint failed before boot: {err}")
        _, base, datadir, stop_server = boot(
            binary,
            workdir,
            port=port,
            env_extra={
                "WEBPHONE_CONFIG": config_path,
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

        driver = make_driver(domain_base if port else "")
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
            # flag, two owned lines with DIDs, the passkey rp pair) and
            # the dial-gate + voicemail lanes (env + fake phone API) need
            # THIS harness's boot config — an externally driven --base
            # server owes them nothing, so they run only on the self-booted
            # lane. Every lane is isolated: one crash must never orphan
            # the rest (the first full run's ElementNotInteractable lesson).
            if stop_server is not None:

                def attempt(name: str, fn) -> None:
                    try:
                        fn()
                    except Exception as err:  # noqa: BLE001 - journey isolation, not control flow
                        c.ok(name, False, str(err)[:160])

                c.ok("J-passkey CLI mints the enroll token", bool(token), token[:24])
                if token:
                    attempt(
                        "J-passkey ceremonies",
                        lambda: journey_passkey(
                            driver, domain_base, c, token, "ops@example.org"
                        ),
                    )

                    def admin_lane() -> None:
                        thread_id = _first_thread_id(driver, domain_base)
                        journey_admin(driver, domain_base, c, thread_id)

                    attempt("J-admin lane", admin_lane)
                    attempt(
                        "J-switch lane",
                        lambda: journey_line_switcher(driver, domain_base, c),
                    )
                attempt(
                    "J-dial lane",
                    lambda: journey_dial_gate(driver, base, c, seed, "+441632960999"),
                )
                attempt("J-vm lane", lambda: journey_voicemail(driver, base, c))
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
