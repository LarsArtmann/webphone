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
import base64
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


def boot(binary: str, workdir: str) -> tuple[subprocess.Popen, str, str, Callable[[], None]]:  # type: ignore[name-defined]
    """Boot the binary on a fresh port + datadir; returns (proc, base,
    datadir, stop)."""
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
    driver_path = shutil.which("chromedriver")
    service = webdriver.ChromeService(executable_path=driver_path) if driver_path else None
    return webdriver.Chrome(options=options, service=service)


def wait(driver, predicate, timeout: float = TIMEOUT):
    from selenium.webdriver.support.ui import WebDriverWait

    return WebDriverWait(driver, timeout).until(predicate)


def journey_login_front_door(driver, base: str, c: Check) -> None:
    driver.get(base + "/login")
    wait(driver, lambda d: d.find_elements("css selector", "#login-form"))
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
                el for el in d.find_elements("css selector", "a.wp-nav-link")
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
            driver.current_url.endswith(path)
            and (current or "").endswith(path),
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
    before = len(
        driver.find_elements("css selector", "#thread-transcript .wp-bubble")
    )
    area = wait(
        driver,
        lambda d: d.find_elements("css selector", "form.wp-compose textarea[name=body]"),
    )[0]
    area.send_keys(body)
    driver.find_element(
        "css selector", "form.wp-compose button[type=submit]"
    ).click()
    try:
        wait(
            driver,
            lambda d: len(
                d.find_elements("css selector", "#thread-transcript .wp-bubble")
            )
            > before,
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
        lambda d: d.find_elements("css selector", "form input[name=name], form input[name='contact-name']"),
    )[0]
    number_input = driver.find_element(
        "css selector", "form input[name=number], form input[name='contact-number']"
    )
    name_input.send_keys(name)
    number_input.send_keys(number)
    driver.find_element(
        "css selector", "form button[type=submit]"
    ).click()
    try:
        wait(driver, lambda d: rows_named(d, True))
        c.ok("J7 contact add lands the row", True)
        delete = wait(
            driver,
            lambda d: d.find_elements("css selector", "button[hx-post*='/contacts/delete']"),
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
            lambda d: d.current_url.endswith("/fax")
            and d.find_elements("css selector", "#fax-panel"),
        )
        c.ok("J10 palette enter navigates", True)
    except Exception as err:  # noqa: BLE001 - journey assertion failure, not control flow
        c.ok("J10 Ctrl+K opens the palette", False, str(err)[:120])


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
        _, base, datadir, stop_server = boot(binary, workdir)
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
        finally:
            driver.quit()
    finally:
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
