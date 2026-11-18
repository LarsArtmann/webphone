#!/usr/bin/env python3
"""Live smoke suite for the webphone binary — real HTTP, real process.

Recreates the ephemeral 14-check suite (the /tmp copy evaporated; this is
its in-repo home). It boots a FRESH binary + data dir (or reuses a running
server via --base) and verifies the surfaces only a running server proves:
CSRF gating, session cookies, signed-in SSE, webhook-to-SSE liveness,
swap-safe fragments, and the proxy/read contracts.

Usage:
  python3 scripts/webphone-smoke.py                 # build + boot + smoke
  python3 scripts/webphone-smoke.py --bin ./result/bin/webphone
  python3 scripts/webphone-smoke.py --base http://127.0.0.1:18099

Stdlib only (no pytest — the buildflow pytest step stays truthfully
skipped for this repo). Exit 0 iff every check passes.
"""

from __future__ import annotations

import argparse
import base64
import http.client
import http.cookiejar
import http.server
import json
import os
import pathlib
import re
import shutil
import socket
import subprocess  # nosec B404 - deliberately spawns the local webphone binary under test
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
from collections.abc import Callable
from typing import ClassVar
from urllib.parse import urlencode, urlparse

TIMEOUT = 10.0


def _version_tuple(text: str) -> tuple[int, ...]:
    parts = []
    for chunk in text.split("."):
        digits = ""
        for ch in chunk:
            if ch.isdigit():
                digits += ch
            else:
                break
        parts.append(int(digits or 0))
    return tuple(parts)


def _fmt(v: tuple[int, ...]) -> str:
    return ".".join(str(p) for p in v)


def ensure_go_toolchain(args: argparse.Namespace) -> None:
    """Re-exec under `nix develop -c` when the ambient go is below the
    go.mod floor.

    The trap this removes (bitten twice, 2026-09-20): the host exports
    GOTOOLCHAIN=local with an older go, so a bare
    `python3 scripts/webphone-smoke.py` dies mid-build with "go.mod
    requires go >= X (running go Y; GOTOOLCHAIN=local)". With flake.nix
    present the fix is one re-exec inside the devShell — the same
    suite, a working toolchain. --base/--bin modes and re-entrant calls
    skip the probe.
    """
    if args.base or args.bin or os.environ.get("WEBPHONE_SMOKE_REEXEC"):
        return
    root = pathlib.Path(__file__).resolve().parent.parent
    go_mod, flake = root / "go.mod", root / "flake.nix"
    if not go_mod.exists() or not flake.exists():
        return
    floor = None
    for line in go_mod.read_text().splitlines():
        m = re.match(r"^go (\d[0-9.]*)$", line.strip())
        if m:
            floor = _version_tuple(m.group(1))
            break
    if floor is None:
        return
    probe = subprocess.run(  # nosec B603 - argv-list (no shell) probe of the local go toolchain
        [args.go, "version"], capture_output=True, text=True, check=False
    )
    m = re.search(r"go version go(\d[0-9.]*)", probe.stdout)
    if m and _version_tuple(m.group(1)) >= floor:
        return
    print(
        "ambient go "
        + (m.group(1) if m else "unknown")
        + " < floor "
        + _fmt(floor)
        + "; re-executing via `nix develop -c` ...",
        flush=True,
    )
    env = dict(os.environ)
    env["WEBPHONE_SMOKE_REEXEC"] = "1"
    result = subprocess.run(  # nosec - argv-list self re-exec into `nix develop` (the documented host-go-below-floor fallback; B603 no shell/untrusted input, B607 PATH by design)
        [
            "nix",
            "develop",
            "-c",
            "python3",
            str(pathlib.Path(__file__).resolve()),
            *sys.argv[1:],
        ],
        env=env,
        check=False,
    )
    raise SystemExit(result.returncode)


class Check:
    def __init__(self) -> None:
        self.failures: list[str] = []
        self.passed = 0
        self.skipped = 0

    def ok(self, name: str, condition: bool, detail: str = "") -> bool:
        if condition:
            self.passed += 1
            print(f"  [ok]   {name}")
        else:
            self.failures.append(f"{name}: {detail}")
            print(f"  [FAIL] {name}: {detail}")
        return condition

    def skip(self, name: str, reason: str) -> None:
        self.skipped += 1
        print(f"  [skip] {name} ({reason})")


class Smoke:
    def __init__(self, base: str) -> None:
        self.base = base.rstrip("/")
        self.host = urlparse(base).hostname or "127.0.0.1"
        self.port = urlparse(base).port or 80
        self.jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.jar)
        )
        self.csrf = ""
        self.check = Check()

    def request(
        self,
        method: str,
        path: str,
        body: bytes | None = None,
        content_type: str = "",
        headers: dict[str, str] | None = None,
    ) -> tuple[int, bytes, dict[str, str]]:
        req = urllib.request.Request(self.base + path, data=body, method=method)
        if self.csrf and method not in ("GET", "HEAD"):
            req.add_header("X-CSRF-Token", self.csrf)
        if content_type:
            req.add_header("Content-Type", content_type)
        for key, value in (headers or {}).items():
            req.add_header(key, value)
        try:
            with self.opener.open(req, timeout=TIMEOUT) as resp:
                return resp.status, resp.read(), dict(resp.headers)
        except urllib.error.HTTPError as err:
            return err.code, err.read(), dict(err.headers)

    def login(self, extension: str = "1001", password: str = "pw") -> bool:  # nosec B107 - fixture credential against the loopback server this script boots
        status, _, _ = self.request(
            "POST",
            "/api/session",
            json.dumps({"extension": extension, "password": password}).encode(),
            "application/json",
        )
        if status != 201:
            return False
        return self.adopt_csrf()

    def adopt_csrf(self) -> bool:
        """Adopt the rotated CSRF token the way the island does (session.js)."""
        status, body, _ = self.request("GET", "/api/csrf")
        if status != 200:
            return False
        token = json.loads(body).get("token", "")
        if not token:
            return False
        self.csrf = token
        return True

    def hook(self, path: str, payload: dict) -> tuple[int, bytes, dict[str, str]]:
        body = json.dumps(payload).encode()
        return self.request(
            "POST",
            path,
            body,
            "application/json",
            {"Authorization": "Bearer test-secret"},
        )

    def sse_events(self, stop: threading.Event, sink: list[tuple[str, str]]) -> None:
        conn = http.client.HTTPConnection(self.host, self.port, timeout=TIMEOUT)
        cookie = "; ".join(f"{c.name}={c.value}" for c in self.jar)
        conn.request("GET", "/events", headers={"Cookie": cookie} if cookie else {})
        resp = conn.getresponse()
        event = ""
        try:
            while not stop.is_set():
                line = resp.fp.readline()
                if not line:
                    break
                text = line.decode("utf-8", "replace").strip()
                if text.startswith("event:"):
                    event = text.split(":", 1)[1].strip()
                elif text.startswith("data:") and event:
                    sink.append((event, text.split(":", 1)[1].strip()))
                    event = ""
        except (OSError, http.client.HTTPException):
            pass
        finally:
            conn.close()

    def wait_for(
        self, sink: list[tuple[str, str]], event: str, deadline: float
    ) -> str | None:
        while time.monotonic() < deadline:
            for name, data in sink:
                if name == event:
                    return data
            time.sleep(0.05)
        return None


def oversized_upload(s: Smoke) -> tuple[int, bytes]:
    """POST a multipart body just past the 61 MB upload cap to
    /messages/send and return the (status, body) of the early response.

    Sends via http.client in chunks with send errors swallowed: the
    server trips MaxBytesReader and answers 400 while the client is
    still uploading, so the write side may raise BrokenPipe before the
    response is readable — the response is the evidence, not the send.
    """
    boundary = "----webphonesmokeboundary"
    prefix = (
        f"--{boundary}\r\n"
        'Content-Disposition: form-data; name="to"\r\n\r\n'
        "+15551234567\r\n"
        f"--{boundary}\r\n"
        'Content-Disposition: form-data; name="body"\r\n\r\n'
        "hi\r\n"
        f"--{boundary}\r\n"
        'Content-Disposition: form-data; name="attachment"; '
        'filename="huge.bin"\r\n'
        "Content-Type: application/octet-stream\r\n\r\n"
    ).encode()
    suffix = f"\r\n--{boundary}--\r\n".encode()
    # 61 MB envelope + 64 KB: past uploadBodyLimit whatever the exact
    # multipart overhead is.
    payload = 61 * 1024 * 1024 + 64 * 1024 - len(prefix) - len(suffix)
    total = len(prefix) + payload + len(suffix)
    cookie = "; ".join(f"{c.name}={c.value}" for c in s.jar)
    conn = http.client.HTTPConnection(s.host, s.port, timeout=TIMEOUT)
    try:
        conn.putrequest("POST", "/messages/send")
        conn.putheader("Content-Type", f"multipart/form-data; boundary={boundary}")
        conn.putheader("Content-Length", str(total))
        conn.putheader("X-CSRF-Token", s.csrf)
        if cookie:
            conn.putheader("Cookie", cookie)
        conn.endheaders()
        chunks = [prefix, b"\0" * payload, suffix]
        try:
            for chunk in chunks:
                conn.send(chunk)
        except OSError:
            pass  # early 400 closes the socket mid-upload
        resp = conn.getresponse()
        return resp.status, resp.read()
    finally:
        conn.close()


def multipart_body(
    fields: list[tuple[str, str]],
    files: list[tuple[str, str, str, bytes]] | None = None,
    boundary: str = "----webphonesmokeboundary",
) -> tuple[bytes, str]:
    """Encode a multipart form body the way the composer/fax forms post.

    files entries are (field, filename, content_type, bytes). Returns the
    body plus the full Content-Type header value to send it with.
    """
    parts: list[bytes] = []
    for name, value in fields:
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="{name}"\r\n\r\n'
                f"{value}\r\n"
            ).encode()
        )
    for field, filename, ctype, payload in files or []:
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="{field}"; '
                f'filename="{filename}"\r\n'
                f"Content-Type: {ctype}\r\n\r\n"
            ).encode()
            + payload
            + b"\r\n"
        )
    parts.append(f"--{boundary}--\r\n".encode())
    return b"".join(parts), f"multipart/form-data; boundary={boundary}"


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def fronted_login(s: Smoke, csrf_cookie: str = "") -> tuple[int, bytes, dict[str, str]]:
    """POST /api/session the way a browser behind the TLS vhost does:
    Origin/Sec-Fetch-Site https + Host pbx.test while the listener sees
    plain http and an X-Forwarded-Proto header. csrf_cookie carries the
    Secure-flagged csrf_token by hand when the trusted origin is https
    (a real browser received it over the TLS hop; a plain-http harness
    must not rely on its cookie policy to resend it)."""
    headers = {
        "Host": "pbx.test",
        "Origin": "https://pbx.test",
        "Sec-Fetch-Site": "same-origin",
        "X-Forwarded-Proto": "https",
    }
    if csrf_cookie:
        headers["Cookie"] = csrf_cookie
    return s.request(
        "POST",
        "/api/session",
        json.dumps({"extension": "1001", "password": "pw"}).encode(),  # nosec B105 - fixture credential for the self-booted smoke server
        "application/json",
        headers=headers,
    )


def secure_csrf_cookie(headers: dict[str, str]) -> str:
    """Extract a Secure csrf_token cookie from raw Set-Cookie headers."""
    for raw in headers.get("Set-Cookie", "").split(","):
        parts = raw.split(";")
        name_value = parts[0].strip()
        if name_value.startswith("csrf_token=") and "secure" in raw.lower():
            return name_value
    return ""


def restart_scenario(binary: str, workdir: str, port: int, env: dict) -> int:
    """Sessions survive a full process kill over the same data dir.

    The SQLite-backed session store (T12) must make a service restart
    invisible to a signed-in tab: login → SIGKILL the server → reboot on
    the same data dir → the old session cookie still opens session-gated
    surfaces. Before T12 this was the silent-401 failure class the
    "Tab session ended" toast could only narrate.
    """
    c = Check()
    base = f"http://127.0.0.1:{port}"
    s = Smoke(base)
    _, body, _ = s.request("GET", "/login")
    m = re.search(
        r'name="csrf-token" content="([^"]+)"', body.decode("utf-8", "replace")
    )
    if m:
        s.csrf = m.group(1)
    c.ok("restart: login accepted", s.login(), "POST /api/session != 201")
    cookie = "; ".join(f"{x.name}={x.value}" for x in s.jar)
    c.ok("restart: session cookie captured", bool(cookie), "jar empty")

    def boot() -> subprocess.Popen:
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("restart scenario: server did not become ready")

    srv = boot()
    try:
        # Kill -9: no graceful shutdown, no cleanup — exactly the crash case.
        srv.kill()
        srv.wait(timeout=5)
    finally:
        pass
    srv = boot()  # same port, same data dir, fresh process
    try:
        status, _, _ = s.request("GET", "/partials/messages")
        c.ok(
            "restart: session survives kill -9",
            status == 200,
            f"session-gated surface answered {status} (401 = persistence broken)",
        )
        # The fresh process must still fail an anonymous probe the same way.
        status, _, _ = Smoke(base).request("GET", "/partials/messages")
        c.ok(
            "restart: anonymous still rejected",
            status == 401,
            f"got {status} (fail-closed posture regressed)",
        )
    finally:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()
    print(f"restart scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def boot_failure_scenario(binary: str, workdir: str, port: int) -> int:
    """A designed boot failure renders the five-part operator contract.

    The operator is a user (ruling 2026-10-02): every boot failure owes
    WHAT/REASSURE/WHY/FIX/ESCAPE on the journal surface, version-stamped
    and class-tagged, with a documented exit code (1 = designed). The
    scenario parks the data dir under a REGULAR FILE, so MkdirAll fails
    with ENOTDIR in the storage phase (app.New's first failure).
    """
    c = Check()
    blocker = pathlib.Path(workdir) / "not-a-dir"
    blocker.write_text("blocks mkdir")
    env = dict(os.environ)
    env["WEBPHONE_ADDR"] = f"127.0.0.1:{port}"
    env["WEBPHONE_DATA_DIR"] = str(blocker / "data")
    try:
        proc = subprocess.run(  # nosec B603 - argv-list (no shell) launch of the local binary (the boot-failure scenario)
            [binary],
            env=env,
            capture_output=True,
            text=True,
            timeout=TIMEOUT,
            check=False,
        )
        code, out = proc.returncode, proc.stdout + proc.stderr
    except subprocess.TimeoutExpired:
        code, out = -1, "boot-failure probe timed out: binary did not exit"
    c.ok(
        "boot failure: designed error exits 1",
        code == 1,
        f"exit {code} (taxonomy: 1 designed, 2 panic)",
    )
    for marker in ("WHAT:", "REASSURE:", "WHY:", "FIX:", "ESCAPE:"):
        c.ok(f"boot failure: contract marker {marker}", marker in out, "missing")
    c.ok(
        "boot failure: storage class named",
        "class=data-dir" in out,
        "header missing class=data-dir",
    )
    c.ok(
        "boot failure: version stamped",
        re.search(r"version=\S+", out) is not None,
        "no version= in the report header",
    )
    print(f"boot failure scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


class _FakeASRHandler(http.server.BaseHTTPRequestHandler):
    """In-process transcription provider standing in for both wires."""

    wire = "openai"
    hits: ClassVar[list[dict[str, str]]] = []

    def do_POST(self) -> None:
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        _FakeASRHandler.hits.append(
            {
                "path": self.path,
                "auth": self.headers.get("Authorization", ""),
                "gkey": self.headers.get("x-goog-api-key", ""),
                "ctype": self.headers.get("Content-Type", ""),
                "body": body.decode("utf-8", "replace"),
            }
        )
        if _FakeASRHandler.wire == "google":
            payload = json.dumps(
                {"results": [{"alternatives": [{"transcript": "smoke google"}]}]}
            ).encode()
        else:
            payload = json.dumps({"text": "smoke hello"}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def log_message(self, *_args: object) -> None:
        pass


def _start_fake_asr(wire: str) -> tuple[str, Callable[[], None]]:
    _FakeASRHandler.wire = wire
    _FakeASRHandler.hits = []
    httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), _FakeASRHandler)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()

    def stop() -> None:
        httpd.shutdown()
        httpd.server_close()

    return f"http://127.0.0.1:{httpd.server_port}", stop


def asr_scenario(binary: str, workdir: str) -> int:
    """The transcription seam end-to-end against fake providers.

    No real speech runs here — what this proves is the WIRING: webphone
    booted with asr.* speaks both provider wires correctly (the
    OpenAI-compatible multipart and the Google v2 JSON recognize), a
    posted segment comes back as text, the island's save call persists
    it, and the History tab renders the section. Real-browser audio is
    the stack E2E's territory; this is the server-side half of M4.
    """
    c = Check()

    def boot(env_extra: dict[str, str], datadir: str) -> tuple[subprocess.Popen, str]:
        port = free_port()
        env = dict(os.environ)
        env.update(
            {
                "WEBPHONE_ADDR": f"127.0.0.1:{port}",
                "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
                "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            }
        )
        env.update(env_extra)
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("asr scenario: server did not become ready")

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    provider, stop_provider = _start_fake_asr("openai")
    srv, base = boot({"WEBPHONE_ASR__URL": provider}, "data-asr-openai")
    try:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        _, body, _ = s.request("GET", "/config.js")
        c.ok(
            "asr: /config.js advertises the seam",
            b'"asr":true' in body,
            body[:160].decode("utf-8", "replace"),
        )
        c.ok("asr: login accepted", s.login(), "POST /api/session != 201")
        status, body, _ = s.request(
            "POST",
            "/api/transcribe?lang=de&filename=seg.webm",
            b"FAKEAUDIO",
            "audio/webm",
        )
        c.ok(
            "asr: openai wire transcribes",
            status == 200 and b"smoke hello" in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
        multipart_hits = [
            h
            for h in _FakeASRHandler.hits
            if h["path"] == "/v1/audio/transcriptions"
            and h["ctype"].startswith("multipart/form-data")
        ]
        c.ok(
            "asr: openai wire saw honest multipart",
            len(multipart_hits) == 1,
            f"hits: {len(multipart_hits)}",
        )
        save = json.dumps(
            {
                "callId": "smoke-call-1",
                "direction": "in",
                "remote": "+49123456789",
                "startedAt": int(time.time() * 1000),
                "text": "smoke hello",
            }
        ).encode()
        status, _, _ = s.request("POST", "/api/transcripts", save, "application/json")
        c.ok("asr: transcript segment saved", status == 204, f"status {status}")
        status, body, _ = s.request("GET", "/partials/history")
        c.ok(
            "asr: history renders the transcript",
            status == 200 and b"smoke hello" in body,
            f"status {status}",
        )
    finally:
        stop_server(srv)
        stop_provider()

    provider, stop_provider = _start_fake_asr("google")
    srv, base = boot(
        {
            "WEBPHONE_ASR__PROVIDER": "google",
            "WEBPHONE_ASR__PROJECT": "smoke-proj",
            "WEBPHONE_ASR__TOKEN": "gkey",  # nosec B105 - fixture token for the in-process fake ASR provider
            "WEBPHONE_ASR__URL": provider,
        },
        "data-asr-google",
    )
    try:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        c.ok("asr(g): login accepted", s.login(), "POST /api/session != 201")
        status, body, _ = s.request(
            "POST", "/api/transcribe", b"FAKEAUDIO", "audio/webm"
        )
        c.ok(
            "asr(g): google wire transcribes",
            status == 200 and b"smoke google" in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
        recognize_hits = [
            h
            for h in _FakeASRHandler.hits
            if "recognizers/_:recognize" in h["path"]
            and "smoke-proj" in h["path"]
            and h["gkey"] == "gkey"
        ]
        c.ok(
            "asr(g): recognizer path + api key on the wire",
            len(recognize_hits) == 1,
            f"hits: {len(recognize_hits)}",
        )
    finally:
        stop_server(srv)
        stop_provider()

    print(f"asr scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def dialgate_scenario(binary: str, workdir: str) -> int:
    """The advisory dial gate (issue #7) end-to-end on a self-booted server.

    Two boots prove both halves of the opt-in contract: with
    WEBPHONE_DIAL__COOLDOWN set, the first ask for a remote is allowed
    (and recorded), the immediate replay is denied with the cooldown
    state the island phrases (last_at + window), the one-shot override
    re-opens exactly one dial, and /config.js advertises the window;
    with zero config the route answers the disabled-seam 404 and
    /config.js stays silent (byte-identical off).
    """
    c = Check()

    def boot(env_extra: dict[str, str], datadir: str) -> tuple[subprocess.Popen, str]:
        port = free_port()
        env = dict(os.environ)
        env.update(
            {
                "WEBPHONE_ADDR": f"127.0.0.1:{port}",
                "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
                "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            }
        )
        env.update(env_extra)
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("dialgate scenario: server did not become ready")

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    def login_csrf(base: str) -> Smoke:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        c.ok("dialgate: login accepted", s.login(), "POST /api/session != 201")
        return s

    # OFF boot: the mounted route answers 404, /config.js stays silent.
    srv, base = boot({}, "data-dialgate-off")
    try:
        s = login_csrf(base)
        status, _, _ = s.request(
            "POST", "/api/dials", b'{"remote":"+4917012345678"}', "application/json"
        )
        c.ok("dialgate: off answers 404", status == 404, f"status {status}")
        _, body, _ = s.request("GET", "/config.js")
        c.ok(
            "dialgate: config.js silent when off",
            b"dialCooldownSeconds" not in body,
            body[:160].decode("utf-8", "replace"),
        )
    finally:
        stop_server(srv)

    # ON boot: allow -> deny with state -> override -> denied again.
    srv, base = boot({"WEBPHONE_DIAL__COOLDOWN": "1h"}, "data-dialgate-on")
    try:
        s = login_csrf(base)
        _, body, _ = s.request("GET", "/config.js")
        c.ok(
            "dialgate: config.js advertises the window",
            b'"dialCooldownSeconds":3600' in body,
            body[:160].decode("utf-8", "replace"),
        )
        status, body, _ = s.request(
            "POST", "/api/dials", b'{"remote":"+4917012345678"}', "application/json"
        )
        c.ok(
            "dialgate: first ask allowed",
            status == 200 and b'"allowed":true' in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
        status, body, _ = s.request(
            "POST", "/api/dials", b'{"remote":"+4917012345678"}', "application/json"
        )
        c.ok(
            "dialgate: replay denied with state",
            status == 200
            and b'"allowed":false' in body
            and b'"window":3600' in body
            and b'"last_at"' in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
        status, body, _ = s.request(
            "POST",
            "/api/dials",
            b'{"remote":"+4917012345678","override":true}',
            "application/json",
        )
        c.ok(
            "dialgate: override allowed",
            status == 200 and b'"allowed":true' in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
        status, body, _ = s.request(
            "POST", "/api/dials", b'{"remote":"+4917012345678"}', "application/json"
        )
        c.ok(
            "dialgate: override is one-shot",
            status == 200 and b'"allowed":false' in body,
            f"{status} {body[:120].decode('utf-8', 'replace')}",
        )
    finally:
        stop_server(srv)

    print(f"dialgate scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def metrics_scenario(binary: str, workdir: str) -> int:
    """The island latency-beacon intake (measure-first design), end-to-end.

    Two boots prove the opt-in contract: with
    WEBPHONE_METRICS__ISLAND_BEACON the session-gated route folds a valid
    beacon into the /metrics histograms (204, then labeled buckets), a
    garbage body is a silent 204 (instrumentation never alerts), an
    unauthenticated POST is a 401, and /config.js advertises "metrics";
    with zero config the route answers 404, /config.js stays silent, and
    /metrics carries no island family (byte-identical off).
    """
    c = Check()

    def boot(env_extra: dict[str, str], datadir: str) -> tuple[subprocess.Popen, str]:
        port = free_port()
        env = dict(os.environ)
        env.update(
            {
                "WEBPHONE_ADDR": f"127.0.0.1:{port}",
                "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
                "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            }
        )
        env.update(env_extra)
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("metrics scenario: server did not become ready")

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    def login_csrf(base: str) -> Smoke:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        c.ok("metrics: login accepted", s.login(), "POST /api/session != 201")
        return s

    beacon = (
        b'{"interval":"dial_to_audio",'
        b'"legs_ms":{"island_overhead":12,"signaling":1500,'
        b'"media_attach":80,"audio_gap":240}}'
    )

    # OFF boot: 404 route, silent config.js, no island family on /metrics.
    srv, base = boot({}, "data-metrics-off")
    try:
        s = login_csrf(base)
        status, _, _ = s.request(
            "POST", "/api/metrics/latency", beacon, "application/json"
        )
        c.ok("metrics: off answers 404", status == 404, f"status {status}")
        _, body, _ = s.request("GET", "/config.js")
        c.ok(
            "metrics: config.js silent when off",
            b'"metrics"' not in body,
            body[:160].decode("utf-8", "replace"),
        )
        _, metrics, _ = s.request("GET", "/metrics")
        c.ok(
            "metrics: no island family when off",
            b"webphone_island_latency_seconds" not in metrics,
            "island family present on a disabled deploy",
        )
    finally:
        stop_server(srv)

    # ON boot: advertise, fold a valid beacon, stay silent on garbage,
    # fence the anonymous, and expose the labeled buckets.
    srv, base = boot({"WEBPHONE_METRICS__ISLAND_BEACON": "true"}, "data-metrics-on")
    try:
        s = login_csrf(base)
        _, body, _ = s.request("GET", "/config.js")
        c.ok(
            "metrics: config.js advertises the intake",
            b'"metrics":true' in body,
            body[:160].decode("utf-8", "replace"),
        )
        status, _, _ = s.request(
            "POST", "/api/metrics/latency", beacon, "application/json"
        )
        c.ok("metrics: valid beacon answers 204", status == 204, f"status {status}")
        status, _, _ = s.request(
            "POST", "/api/metrics/latency", b"{not json at all", "application/json"
        )
        c.ok(
            "metrics: garbage body is a silent 204",
            status == 204,
            f"status {status}",
        )
        _, metrics, _ = s.request("GET", "/metrics")
        c.ok(
            "metrics: histogram bucket appears with the leg label",
            b'webphone_island_latency_seconds_bucket{interval="dial_to_audio",leg="signaling"'
            in metrics,
            metrics[:400].decode("utf-8", "replace"),
        )
        c.ok(
            "metrics: count line folds the beacon",
            b'webphone_island_latency_seconds_count{interval="dial_to_audio",leg="signaling"} 1'
            in metrics,
            "count != 1 after one valid beacon",
        )
        status, _, _ = s.request(
            "POST", "/api/metrics/latency", beacon, "application/json"
        )
        anon_status = urllib.request.Request(
            base + "/api/metrics/latency", data=beacon, method="POST"
        )
        try:
            urllib.request.urlopen(anon_status, timeout=2)  # nosec B310 - probe of the local server under test
            anon_code = 0
        except urllib.error.HTTPError as err:
            anon_code = err.code
        c.ok(
            "metrics: anonymous POST is fenced",
            anon_code in (401, 403),
            f"anon status {anon_code}",
        )
    finally:
        stop_server(srv)

    print(f"metrics scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def agent_scenario(binary: str, workdir: str) -> int:
    """The textagent seam's boot honesty (T05) on self-booted servers.

    With WEBPHONE_TEXTAGENT__URL + SECRET_FILE configured, /config.js
    advertises "agent":true (the island's summarize gate rides it) and
    the summarize route is LIVE (a session ask reaches the transcript
    lookup — "no transcript", not the disabled-seam "not found"); with
    zero config the flag stays absent and the route answers the
    disabled-seam 404 — byte-identical off.
    """
    c = Check()

    def boot(env_extra: dict[str, str], datadir: str) -> tuple[subprocess.Popen, str]:
        port = free_port()
        env = dict(os.environ)
        env.update(
            {
                "WEBPHONE_ADDR": f"127.0.0.1:{port}",
                "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
                "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            }
        )
        env.update(env_extra)
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("agent scenario: server did not become ready")

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    def login_csrf(base: str) -> Smoke:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        c.ok("agent: login accepted", s.login(), "POST /api/session != 201")
        return s

    # OFF boot: the flag is absent, the route answers the disabled seam.
    srv, base = boot({}, "data-agent-off")
    try:
        _, body, _ = Smoke(base).request("GET", "/config.js")
        c.ok(
            "agent: config.js silent when off",
            b'"agent":true' not in body,
            "config.js advertised the seam without config",
        )
        s = login_csrf(base)
        status, rbody, _ = s.request("POST", "/api/calls/smoke-call/summary-draft")
        c.ok(
            "agent: summarize route answers the disabled seam",
            status == 404 and b"not found" in rbody,
            f"got {status} {rbody[:60]!r}",
        )
    finally:
        stop_server(srv)

    # ON boot: the flag advertises, the route reaches the transcript gate.
    secret = f"{workdir}/agent-secret"
    with open(secret, "w", encoding="utf-8") as fh:
        fh.write("review-secret\n")
    srv, base = boot(
        {
            "WEBPHONE_TEXTAGENT__URL": "http://127.0.0.1:9",  # deliberately dead: the route must fail PAST the seam gate
            "WEBPHONE_TEXTAGENT__SECRET_FILE": secret,
        },
        "data-agent-on",
    )
    try:
        _, body, _ = Smoke(base).request("GET", "/config.js")
        c.ok(
            "agent: config.js advertises the seam",
            b'"agent":true' in body,
            "config.js missing the agent flag with the seam wired",
        )
        s = login_csrf(base)
        status, rbody, _ = s.request("POST", "/api/calls/smoke-call/summary-draft")
        c.ok(
            "agent: summarize route is live",
            status == 404 and b"transcript" in rbody,
            f"got {status} {rbody[:60]!r}",
        )
    finally:
        stop_server(srv)

    print(f"agent scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def faxcap_scenario(binary: str, workdir: str) -> int:
    """The fax capability declaration, end-to-end on a self-booted server.

    The Fax tab must state what the DEPLOYMENT can do instead of offering
    an affordance it cannot honour: with fax.send/fax.receive off there is
    no compose form, the panel carries the capability notice (and Settings
    names the missing halves); with zero config the panel is the historical
    one (form + plain copy), so the smoke also pins the byte-parity path.
    """
    c = Check()

    def boot(env_extra: dict[str, str], datadir: str) -> tuple[subprocess.Popen, str]:
        port = free_port()
        env = dict(os.environ)
        env.update(
            {
                "WEBPHONE_ADDR": f"127.0.0.1:{port}",
                "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
                "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            }
        )
        env.update(env_extra)
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("faxcap scenario: server did not become ready")

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    def login(base: str) -> Smoke:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            s.csrf = m.group(1)
        c.ok("faxcap: login accepted", s.login(), "POST /api/session != 201")
        return s

    # OFF boot: no send form, the notice + the Settings row instead.
    srv, base = boot(
        {"WEBPHONE_FAX__SEND": "false", "WEBPHONE_FAX__RECEIVE": "false"},
        "data-faxcap-off",
    )
    try:
        s = login(base)
        _, body, _ = s.request("GET", "/fax")
        page = body.decode("utf-8", "replace")
        c.ok(
            "faxcap: no compose form without a lane",
            "wp-compose-fax" not in page,
            "the send form rendered on a deployment with no fax lane",
        )
        c.ok(
            "faxcap: panel says fax is unavailable",
            "wp-fax-off" in page and "Fax is not available on this server" in page,
            page[:200],
        )
        _, body, _ = s.request("GET", "/settings")
        settings = body.decode("utf-8", "replace")
        c.ok(
            "faxcap: settings names the missing directions",
            "not configured — nothing can be sent or received" in settings,
            "Settings lost the fax row",
        )
    finally:
        stop_server(srv)

    # Default boot: the historical panel (parity path), no notice.
    srv, base = boot({}, "data-faxcap-on")
    try:
        s = login(base)
        _, body, _ = s.request("GET", "/fax")
        page = body.decode("utf-8", "replace")
        c.ok(
            "faxcap: default keeps the compose form",
            "wp-compose-fax" in page,
            "the zero-config Fax tab lost its send form",
        )
        c.ok(
            "faxcap: default renders no capability notice",
            "wp-fax-off" not in page,
            "a capability notice leaked into the default panel",
        )
        _, body, _ = s.request("GET", "/settings")
        settings = body.decode("utf-8", "replace")
        c.ok(
            "faxcap: settings states the available lanes",
            "sending and receiving available" in settings,
            "Settings lost the always-on fax lanes row (owner decision 2, 2026-10-30)",
        )
    finally:
        stop_server(srv)

    print(f"faxcap scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def enroll_scenario(binary: str, workdir: str) -> int:
    """The -enroll-passkey CLI one-shot end-to-end (the 2026-11-09
    split-brain class).

    One config file drives BOTH sides: the CLI one-shot runs under the
    SAME config and data dir as the server, so its token lands in the
    same system.db journal the server verifies from — mint on a fresh
    dir, boot the server, verify the fresh token 200, replay the burned
    token 503, and prove the /enroll page (the conditional surface)
    renders with passkey enabled.
    """
    c = Check()

    datadir = f"{workdir}/data-enroll"
    os.makedirs(datadir, exist_ok=True)
    pwfile = f"{workdir}/pw-1001"
    with open(pwfile, "w", encoding="utf-8") as fh:
        fh.write("pw\n")
    cfg = {
        # The CLI one-shot validates the WHOLE config: a real addr shape
        # up front, rewritten to the server's port before boot.
        "addr": "127.0.0.1:18099",
        "data_dir": datadir,
        "gateway": {"webhook_secret": "test-secret"},  # nosec B105 - fixture secret for the self-booted smoke server
        "auth": {
            "users": {"lars@example.com": {"extensions": ["1001"]}},
            "extension_password_files": {"1001": pwfile},
            "passkey": {
                "rp_id": "webphone.test",
                "rp_origins": ["https://webphone.test"],
            },
        },
    }
    cfg_path = f"{workdir}/config-enroll.json"
    with open(cfg_path, "w", encoding="utf-8") as fh:
        json.dump(cfg, fh)

    def cli_mint() -> tuple[int, str, str]:
        env = dict(os.environ)
        env["WEBPHONE_CONFIG"] = cfg_path
        done = subprocess.run(  # nosec B603 - argv-list (no shell) run of the CLI one-shot under test
            [binary, "-enroll-passkey", "lars@example.com"],
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        return done.returncode, done.stdout, done.stderr

    def boot() -> tuple[subprocess.Popen, str]:
        port = free_port()
        cfg["addr"] = f"127.0.0.1:{port}"
        with open(cfg_path, "w", encoding="utf-8") as fh:
            json.dump(cfg, fh)
        env = dict(os.environ)
        env["WEBPHONE_CONFIG"] = cfg_path
        srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                return srv, base
            except OSError:
                time.sleep(0.1)
        srv.kill()
        raise RuntimeError("enroll scenario: server did not become ready")

    # Mint FIRST: the one-shot opens the journal alone, the server joins
    # it after — the shared-journal contract without a write race.
    rc, out, err = cli_mint()
    c.ok("enroll: CLI one-shot exits 0", rc == 0, err.strip()[:160])
    m = re.search(r"token=([A-Za-z0-9_-]+)", out)
    c.ok("enroll: the link carries a token", m is not None, out.strip()[:160])
    if not m:
        print(f"enroll scenario: {c.passed} passed, {len(c.failures)} failed")
        for failure in c.failures:
            print(f"  FAILED: {failure}")
        return 1
    token = m.group(1)

    srv, base = boot()
    try:
        s = Smoke(base)
        status, body, _ = s.request("GET", "/enroll")
        page = body.decode("utf-8", "replace")
        c.ok(
            "enroll: the page renders with passkey on",
            status == 200,
            f"status {status}",
        )
        m = re.search(r'name="csrf-token" content="([^"]+)"', page)
        if m:
            s.csrf = m.group(1)

        status, body, _ = s.request(
            "POST",
            "/api/auth/passkey/enroll/verify",
            json.dumps({"token": token}).encode(),
            "application/json",
        )
        c.ok(
            "enroll: fresh token verifies 200",
            status == 200 and b"lars@example.com" in body,
            f"status {status} body {body[:120].decode('utf-8', 'replace')}",
        )

        status, _, _ = s.request(
            "POST",
            "/api/auth/passkey/enroll/verify",
            json.dumps({"token": token}).encode(),
            "application/json",
        )
        c.ok("enroll: burned token answers 503", status == 503, f"status {status}")
    finally:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    print(f"enroll scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def run_checks(
    s: Smoke,
    boot_configured: Callable[[], tuple[str, Callable[[], None]]] | None = None,
    foreign: bool = False,
    expect_version: str | None = None,
) -> int:
    c = s.check
    print(f"smoke against {s.base}")

    # 1. The login page is the front door in EVERY mode: the extension
    #    form + CSRF meta render anonymously, and the anonymous root
    #    navigation redirects here (the signed-in shell is checked after
    #    the login below). The redirect-following client lands on the
    #    login page for both.
    status, body, _ = s.request("GET", "/login")
    page = body.decode("utf-8", "replace")
    c.ok("login page 200", status == 200, f"got {status}")
    c.ok(
        "login page holds the extension form",
        'id="login-form"' in page and 'id="login-error"' in page,
        "login form ids missing",
    )
    c.ok(
        "login form drives its own module",
        "/assets/login-page.js" in page,
        "the standalone submit module is not referenced",
    )
    status_root, body_root, _ = s.request("GET", "/")
    c.ok(
        "anonymous root lands on the front door",
        status_root == 200 and b"login-form" in body_root,
        f"root GET answered {status_root} without the login form",
    )
    match = re.search(r'name="csrf-token" content="([^"]+)"', page)
    c.ok("CSRF meta present", match is not None, "no csrf-token meta")
    if match:
        s.csrf = match.group(1)

    # 2. healthz is honest readiness JSON.
    status, body, _ = s.request("GET", "/healthz")
    c.ok(
        "healthz ready",
        status == 200 and b'"ok"' in body.lower(),
        f"{status} {body[:80]!r}",
    )

    # 2b. the go-health probe pair: liveness is fetch-free and answers
    # pass; startup latches to pass once the backing checks first ran
    # (both are session-free JSON — the NixOS vhost proxies them).
    status, body, _ = s.request("GET", "/livez")
    c.ok(
        "livez pass",
        status == 200 and b'"status":"pass"' in body,
        f"{status} {body[:80]!r}",
    )
    status, body, _ = s.request("GET", "/startupz")
    c.ok(
        "startupz latched pass",
        status == 200 and b'"status":"pass"' in body,
        f"{status} {body[:80]!r}",
    )

    # 2c. the health dashboard (this suite boots with
    # WEBPHONE_DASHBOARD__ENABLE=true): HTML page under the dashboard's
    # nonce'd CSP, the embedded Datastar SDK served same-origin, and the
    # namespaced probe alias answering from the same probe as /livez.
    status, body, headers = s.request("GET", "/health")
    dash_page = body.decode("utf-8", "replace")
    c.ok(
        "dashboard html serves",
        status == 200 and "/health/datastar.js" in dash_page and "nonce=" in dash_page,
        f"{status} {body[:80]!r}",
    )
    csp = headers.get("Content-Security-Policy", "")
    c.ok(
        "dashboard nonce csp",
        "unsafe-eval" not in csp and "nonce-" in csp,
        csp[:100],
    )
    status, body, _ = s.request("GET", "/health/datastar.js")
    c.ok("dashboard sdk same-origin", status == 200, f"got {status}")
    status, body, _ = s.request("GET", "/health/livez")
    c.ok(
        "dashboard probe alias pass",
        status == 200 and b'"status":"pass"' in body,
        f"{status} {body[:80]!r}",
    )
    # The trailing-slash form is the SUBTREE pattern: it must reach the
    # same dashboard, not the 404 that answered here while the prefix
    # route was registered without its bare sibling.
    status, body, _ = s.request("GET", "/health/")
    c.ok(
        "dashboard subtree serves",
        status == 200 and b"/health/datastar.js" in body,
        f"{status} {body[:80]!r}",
    )

    # 3. version reports build metadata. With --expect-version the probe
    # FAILS unless the served build is exactly that version (accepts the
    # form with or without the leading v) — the post-deploy check that
    # the switch actually switched.
    status, body, _ = s.request("GET", "/version")
    c.ok(
        "version reachable",
        status == 200 and b'"version"' in body,
        f"{status} {body[:80]!r}",
    )
    # Key-checked (not shape-pinned): when the flake injected commit
    # metadata, both keys must be present and non-empty — the enrichment
    # that kills the store-path chain-verification dance.
    if b'"commit"' in body:
        c.ok(
            "version commit enrichment present",
            re.search(rb'"commit":"[0-9a-f]{7,40}(-dirty)?"', body) is not None,
            f"{body[:160]!r}",
        )
        c.ok(
            "version commitDate present",
            re.search(rb'"commitDate":"\d{4}-\d{2}-\d{2}T', body) is not None,
            f"{body[:160]!r}",
        )
    if expect_version is not None:
        want = expect_version.lstrip("v")
        detail = f"{body[:120]!r}"
        served = re.search(rb'"version":"([^"]+)"', body)
        if served and (
            served[1].startswith(b"v0.0.0-")
            or b"-0." in served[1]  # tag-based pseudo-version: vX.Y.Z-0.<ts>-<hash>
            or b"+dirty" in served[1]
            or b"devel" in served[1]
        ):
            detail += (
                " — the served build is a bare `go build` artifact (Go's "
                "pseudo-version); rebuild via nix: --bin $(nix build .#webphone)"
            )
        c.ok(
            f"version is exactly {expect_version}",
            f'"version":"v{want}"'.encode() in body,
            detail,
        )

    # 3b. openapi.json publishes the machine API contract (3.0.3 — the
    # go-cqrs-lite catalog exporter's spec version; upstream may emit
    # 3.1.0 later, the pin moves with the generator's truth). The dialer
    # surface must be in the published contract too: /api/dials shipped
    # with the T19-T27 dialer train and a contract that silently drops a
    # route is the drift class the generated doc exists to prevent.
    status, body, _ = s.request("GET", "/openapi.json")
    c.ok(
        "openapi published",
        status == 200 and b'"openapi": "3.0.3"' in body,
        f"{status} {body[:80]!r}",
    )
    c.ok(
        "openapi carries /api/dials",
        status == 200 and b"/api/dials" in body,
        f"{status} dials={b'/api/dials' in body}",
    )

    # 4. config.js carries the PBX contract for the island.
    status, body, _ = s.request("GET", "/config.js")
    c.ok(
        "config.js contract",
        status == 200 and b"window.PBX_CONFIG" in body and b"sipDomain" in body,
        f"{status}",
    )

    # 4a. theme preload: the same-origin asset serves with a JavaScript
    # content type, and the shell page carries ZERO inline <script> tags
    # (strict CSP — a forced theme must not depend on inline bytes).
    status, body, headers = s.request("GET", "/assets/theme-preload.js")
    ctype = next((v for k, v in headers.items() if k.lower() == "content-type"), "")
    c.ok(
        "theme preload serves",
        status == 200 and "javascript" in ctype.lower() and b"data-theme" in body,
        f"{status} {ctype!r}",
    )
    # 4b. /partials/nav contract (F19): a signed-out session gets ZERO
    # links — the desk's tab row is for sessions that can use it; the
    # signed-in re-fetch below is the only badge source. A fresh Smoke
    # keeps it session-free even mid-suite.
    anon_nav = Smoke(s.base)
    # 4a-bis. The zero-inline-script contract belongs to the SHELL: an
    # anonymous deep link renders it session-free (the login page is the
    # nonce'd inline-script surface by design, so `page` above must not
    # be the probe).
    status, shell_anon, _ = anon_nav.request("GET", "/messages")
    shell_page = shell_anon.decode("utf-8", "replace")
    inline = re.search(r"<script(?![^>]*\bsrc=)[^>]*>.{0,120}", shell_page, re.DOTALL)
    c.ok(
        "shell has zero inline scripts",
        status == 200 and inline is None and b'id="reg-status"' in shell_anon,
        f"{status} inline <script> without src found: {inline.group(0)[:140]!r}"
        if inline
        else f"{status} not the shell",
    )
    status, body, _ = anon_nav.request("GET", "/partials/nav")
    nav_anon = body.decode("utf-8", "replace")
    c.ok(
        "nav partial anonymous shape",
        status == 200 and "wp-nav-link" not in nav_anon,
        f"{status} anonymous nav should be link-free: {nav_anon[:80]!r}",
    )

    # 4c. Unknown paths render the STYLED 404: shell chrome around the
    # not-found panel (literal 404, the requested path echoed back, a way
    # home, NO reload button), status stays 404. Works in --base foreign
    # mode too, so the post-deploy probe validates prod. The island fact
    # is #phone-view — the login card left the island (front-door train).
    status, body, _ = anon_nav.request("GET", "/definitely/not/a/path")
    nf = body.decode("utf-8", "replace")
    c.ok(
        "styled 404",
        status == 404
        and 'class="wp-panel wp-notfound"' in nf
        and "/definitely/not/a/path" in nf
        and "Back to Messages" in nf
        and 'id="phone-view"' in nf,
        f"{status} bare-or-unstyled: {nf[:80]!r}",
    )

    # 5. Session POST is CSRF-gated: without the token it must not pass.
    plain = Smoke(s.base)
    plain.request("GET", "/")
    status, _, _ = plain.request(
        "POST",
        "/api/session",
        json.dumps({"extension": "1001", "password": "pw"}).encode(),  # nosec B105 - fixture credential (this CSRF-gate test EXPECTS the rejection)
        "application/json",
    )
    c.ok("session CSRF-gated", status in (403, 401), f"got {status}")

    # 6. Session with CSRF token issues the session cookie. A FOREIGN
    # server is probed with bogus credentials on purpose instead: a
    # hardened build (v2.1.1+, server-side credential verification) must
    # reject them with 401; a 201 means the server mints sessions without
    # verification (the forged-session vulnerability fixed in v2.1.1).
    # Everything downstream that needs a real session, the webhook secret,
    # or injected traffic is honestly skipped in foreign mode.
    pre_login_token = s.csrf
    sink: list[tuple[str, str]] = []
    stop: threading.Event | None = None
    if foreign:
        status, _, _ = s.request(
            "POST",
            "/api/session",
            json.dumps({"extension": "1001", "password": "definitely-wrong"}).encode(),  # nosec B105 - negative-test credential: this login MUST be rejected
            "application/json",
        )
        c.ok(
            "bogus credentials rejected",
            status == 401,
            f"got {status}"
            + (
                " (server mints sessions without verifying credentials: pre-v2.1.1 build)"
                if status == 201
                else ""
            ),
        )
        for name in (
            "login rotated the CSRF token",
            "stale CSRF token rejected",
            "adopted CSRF token accepted",
            "session cookie issued",
            "signed-in SSE stream live",
            "unknown hook 404",
            "inbound webhook 202",
            "live threads event",
            "threads fragment swap-safe",
            "nav partial signed-in badge",
            "thread list renders row",
            "transcript bubble",
            "live thread event",
            "thread fragment is bubbles only",
            "live mark-read 204",
            "oversized upload rejected",
        ):
            c.skip(
                name, "needs a self-booted server: real credentials + webhook secret"
            )
    else:
        c.ok("session login accepted", s.login(), "POST /api/session != 201")
        c.ok(
            "login rotated the CSRF token",
            s.csrf != pre_login_token and s.adopt_csrf(),
            "adoption missing or stale token reused",
        )
        # The signed-in root now renders the SHELL (the island contract
        # markers live here since the front door took the login form).
        status, shell, _ = s.request("GET", "/")
        c.ok(
            "signed-in shell serves the island",
            status == 200
            and b'id="reg-status"' in shell
            and b'id="phone-view"' in shell,
            f"{status} island ids missing",
        )
        # 6b. The pre-login token is dead: POSTs with it must 403, while the
        # same probe with the adopted token passes CSRF (unknown path = 404).
        status, _, _ = s.request(
            "POST",
            "/api/csrf-rotate-probe",
            b"",
            "",
            {"X-CSRF-Token": pre_login_token},
        )
        c.ok("stale CSRF token rejected", status == 403, f"got {status}")
        status, _, _ = s.request(
            "POST",
            "/api/csrf-rotate-probe",
            b"",
            "",
            {"X-CSRF-Token": s.csrf},
        )
        c.ok("adopted CSRF token accepted", status == 404, f"got {status}")
        c.ok(
            "session cookie issued",
            any(cookie.name == "webphone_session" for cookie in s.jar),
            "no webphone_session cookie",
        )

        # 6c. Boot resume: a live cookie gets the session's SIP credentials
        # back (the island re-registers without the login form), the
        # credential answer is never cacheable, and logout kills the
        # resume like every other gated surface. The sliding renewal is
        # unit-pinned (renewDue + Attach cookie parity); the smoke pins
        # only the HTTP wiring.
        status, body, headers = s.request("GET", "/api/session")
        c.ok("resume 200 with live cookie", status == 200, f"got {status}")
        c.ok(
            "resume is no-store",
            headers.get("Cache-Control") == "no-store",
            headers.get("Cache-Control", "<missing>"),
        )
        resumed = {}
        try:
            resumed = json.loads(body)
        except ValueError:
            pass
        c.ok(
            "resume carries the SIP credentials",
            resumed.get("extension") == "1001" and bool(resumed.get("password")),
            f"keys={sorted(resumed)}",
        )
        status, _, _ = s.request("DELETE", "/api/session")
        c.ok("logout 204", status == 204, f"got {status}")
        status, _, _ = s.request("GET", "/api/session")
        c.ok("resume after logout 401s", status == 401, f"got {status}")
        # Logout rotated the CSRF cookie as well (the island reloads for a
        # fresh page; the harness re-adopts via the GET).
        c.ok(
            "re-login after resume scenario",
            s.adopt_csrf() and s.login(),
            "POST /api/session != 201",
        )

        # 7. Signed-in SSE connects.
        stop = threading.Event()
        reader = threading.Thread(target=s.sse_events, args=(stop, sink), daemon=True)
        reader.start()
        time.sleep(0.4)
        c.ok("signed-in SSE stream live", len(sink) > 0, "no frames within 400ms")
    status, _, _ = Smoke(s.base).request("GET", "/events")
    c.ok("anonymous SSE rejected", status == 401, f"got {status}")

    if not foreign:
        # 8. Unknown hook path is a 404 even with the secret.
        status, _, _ = s.hook("/hooks/nope", {})
        c.ok("unknown hook 404", status == 404, f"got {status}")

        # 9. Inbound message webhook accepts.
        status, _, _ = s.hook(
            "/hooks/message",
            {"owner": "1001", "from": "+441632960961", "body": "smoke inbound"},
        )
        c.ok("inbound webhook 202", status == 202, f"got {status}")

        # 10. The push lands live as a swap-safe threads fragment.
        data = s.wait_for(sink, "threads", time.monotonic() + TIMEOUT)
        c.ok("live threads event", data is not None, "no threads event in time")
        c.ok(
            "threads fragment swap-safe",
            data is not None and "wp-thread-row" in data and "<section" not in data,
            f"payload {str(data)[:80]!r}",
        )

        # 10b. Signed-in nav re-renders with the unread badge (before the
        # thread opens below and marks it read — the badge's only honest
        # window in this suite).
        status, body, _ = s.request("GET", "/partials/nav?active=messages")
        nav_authed = body.decode("utf-8", "replace")
        c.ok(
            "nav partial signed-in badge",
            status == 200
            and "wp-nav-link" in nav_authed
            and "wp-nav-badge" in nav_authed,
            f"{status} badge missing: {nav_authed[:80]!r}",
        )

        # 10c. The Reports tab (issue #12): a deep link renders the full
        # shell with the panel, and the partial answers on its own. The
        # loopback smoke runs WITHOUT a phone API, so the panel must say
        # which knob is missing rather than render a quiet week of zeros —
        # the same honesty the History tab owes.
        status, body, _ = s.request("GET", "/reports")
        reports_page = body.decode("utf-8", "replace")
        c.ok(
            "reports deep link renders the panel",
            status == 200
            and 'id="reports-panel"' in reports_page
            and 'data-tab="reports"' in reports_page,
            f"{status} panel or nav entry missing",
        )
        status, body, _ = s.request("GET", "/partials/reports")
        reports_partial = body.decode("utf-8", "replace")
        c.ok(
            "reports names the missing phone API",
            status == 200
            and "phone_api_url" in reports_partial
            and "wp-stat-value" not in reports_partial,
            f"{status} disabled-API shape wrong: {reports_partial[:80]!r}",
        )

        # 11. Thread row exists; opening it shows the transcript bubble region.
        _, body, _ = s.request("GET", "/partials/messages")
        rows = re.findall(r'hx-get="/partials/messages/([^"]+)"', body.decode())
        c.ok("thread list renders row", bool(rows), "no thread row")
        if rows:
            _, body, _ = s.request("GET", f"/partials/messages/{rows[0]}")
            transcript = body.decode()
            c.ok(
                "transcript bubble",
                'id="thread-transcript"' in transcript
                and 'sse-swap="thread"' in transcript,
                "swap region missing",
            )

        # 12. Live thread push carries bubble fragments (not the whole panel).
        s.hook(
            "/hooks/message",
            {"owner": "1001", "from": "+441632960961", "body": "smoke live"},
        )
        data = s.wait_for(sink, "thread", time.monotonic() + TIMEOUT)
        c.ok("live thread event", data is not None, "no thread event in time")
        c.ok(
            "thread fragment is bubbles only",
            data is not None
            and "wp-bubble" in data
            and "wp-compose" not in data
            and "<section" not in data,
            f"payload {str(data)[:80]!r}",
        )

        # 13. Explicit read endpoint (the live-swap mark-read path).
        if rows:
            status, _, _ = s.request("POST", f"/messages/{rows[0]}/read")
            c.ok("live mark-read 204", status == 204, f"got {status}")
            if stop is not None:
                stop.set()

        # 13b. The upload body cap: a multipart just past the 61 MB
        # envelope must answer 400 naming the size — the only unbounded
        # request path the audit found (finding #1), pinned live.
        status, upload_body = oversized_upload(s)
        c.ok(
            "oversized upload rejected",
            status == 400 and b"too large" in upload_body,
            f"got {status}: {upload_body[:120]!r}",
        )

    # 13c. Notes API roundtrip (T05): the one-slot-per-call contract on a
    # fresh binary — PUT records, GET reads back, DELETE frees (the 409's
    # remedy proves itself at boot, not only in the Go suite), and the
    # slot reads EMPTY after the delete (a 200 with no body, the
    # notes-precedent shape).
    note_call = "smoke-note-call"
    status, _, _ = s.request(
        "PUT",
        f"/api/calls/{note_call}/note",
        json.dumps({"body": "smoke note"}).encode(),
        "application/json",
    )
    c.ok("notes: PUT records the slot", status == 204, f"got {status}")
    status, body, _ = s.request("GET", f"/api/calls/{note_call}/note")
    c.ok(
        "notes: GET reads it back",
        status == 200 and b"smoke note" in body,
        f"{status} {body[:60]!r}",
    )
    status, _, _ = s.request("DELETE", f"/api/calls/{note_call}/note")
    c.ok("notes: DELETE frees the slot", status == 204, f"got {status}")
    status, body, _ = s.request("GET", f"/api/calls/{note_call}/note")
    c.ok(
        "notes: GET after delete reads the empty slot",
        status == 200 and b'"body":""' in body,
        f"{status} {body[:40]!r}",
    )

    # 13d. SUPERB E2E train (2026-11-18): the product's WRITE surfaces as
    # workflows against the same loopback boot — composer send, thread
    # flags, snippets, contacts CRUD/import/export, MMS attachments, the
    # fax lifecycle, the csp-report intake, and the manifest/export open
    # surfaces. Session-gated lanes run only against a self-booted server
    # (foreign mode has no session to write with).
    # (a) csp-report intake: malformed stays best-effort 204, an oversized
    # report hits the 4 KiB cap with 413 — the observable-CSP lane's fence.
    status, _, _ = Smoke(s.base).request("POST", "/csp-report", b"not json", "application/json")
    c.ok("csp-report malformed 204", status == 204, f"got {status}")
    # A report that is VALID JSON SHAPE past the 4 KiB cap hits the
    # honest 413 (a syntactically broken report errors before the reader
    # reaches the cap and stays best-effort 204 — see the malformed probe).
    status, _, _ = Smoke(s.base).request(
        "POST", "/csp-report", b'{"documentURI":"' + b"x" * 6144 + b'"}', "application/json"
    )
    c.ok("csp-report oversized 413", status == 413, f"got {status}")

    # (b) PWA-lite manifest + favicon: the install surface stays honest.
    status, body, headers = s.request("GET", "/manifest.webmanifest")
    mtype = next((v for k, v in headers.items() if k.lower() == "content-type"), "")
    c.ok(
        "manifest serves",
        status == 200 and "manifest+json" in mtype and b'"name"' in body,
        f"{status} {mtype!r}",
    )
    status, body, headers = s.request("GET", "/favicon.svg")
    ftype = next((v for k, v in headers.items() if k.lower() == "content-type"), "")
    c.ok("favicon serves", status == 200 and "svg" in ftype.lower(), f"{status} {ftype!r}")

    # (c) data export: anonymous is fenced; the session answer is a real zip.
    anon_export = Smoke(s.base)
    status, _, _ = anon_export.request("GET", "/api/export")
    c.ok("export anonymous 401", status == 401, f"got {status}")

    if not foreign:
        status, body, _ = s.request("GET", "/api/export")
        c.ok(
            "export answers a zip",
            status == 200 and body[:4] == b"PK\x03\x04",
            f"{status} magic={body[:4]!r}",
        )

        # (d) Composer send: a new conversation via the real form shape,
        # then the thread list and transcript both reflect the bubble.
        send_to = "+441632960111"
        body_bytes, ctype = multipart_body([("to", send_to), ("body", "smoke outbound workflow")], [])
        status, body, _ = s.request("POST", "/messages/send", body_bytes, ctype)
        c.ok(
            "workflow send accepted",
            status == 200 and b'messages-panel' in body,
            f"{status} {body[:80]!r}",
        )
        status, body, _ = s.request("GET", "/partials/messages")
        listing = body.decode("utf-8", "replace")
        sent_rows = re.findall(r'hx-get="/partials/messages/([^"]+)"', listing)
        sent_thread = ""
        for row in sent_rows:
            if f'thread-{row}"' in listing and send_to in listing:
                sent_thread = row
                break
        if not sent_thread and sent_rows:
            sent_thread = sent_rows[0]
        c.ok(
            "workflow send thread listed",
            bool(sent_thread) and send_to in listing,
            f"rows={sent_rows[:3]}",
        )
        if sent_thread:
            _, body, _ = s.request("GET", f"/partials/messages/{sent_thread}")
            transcript = body.decode("utf-8", "replace")
            c.ok(
                "workflow send bubble in transcript",
                "smoke outbound workflow" in transcript and "wp-status-" in transcript,
                "bubble or status badge missing",
            )
            # (e) Deep link renders the OPEN thread in the full shell; a
            # bogus id degrades to the list + the gone-notice (no 404).
            status, body, _ = s.request("GET", f"/messages/{sent_thread}")
            c.ok(
                "workflow deep link opens thread",
                status == 200 and 'id="thread-transcript"' in body.decode("utf-8", "replace"),
                f"got {status}",
            )
            status, body, _ = s.request("GET", "/messages/zzz-not-a-thread")
            c.ok(
                "workflow bogus thread degrades to gone-notice",
                status == 200 and "no longer exists" in body.decode("utf-8", "replace"),
                f"got {status}",
            )

            # (f) Thread flags (T18): pin marks the row, archive moves it to
            # the archived view and back, mute is presentation-only. The
            # head-target variant answers the bare 204 the row lane does.
            status, _, _ = s.request("POST", f"/messages/{sent_thread}/pin?on=1")
            c.ok("workflow pin 204", status == 204, f"got {status}")
            _, body, _ = s.request("GET", "/partials/messages")
            c.ok(
                "workflow pinned row marked",
                f'thread-{sent_thread}"' in body.decode("utf-8", "replace")
                and "wp-pinned" in body.decode("utf-8", "replace"),
                "wp-pinned missing from the list",
            )
            status, _, _ = s.request("POST", f"/messages/{sent_thread}/archive?on=1")
            c.ok("workflow archive 204", status == 204, f"got {status}")
            _, body, _ = s.request("GET", "/partials/messages")
            unarchived_view = body.decode("utf-8", "replace")
            _, body, _ = s.request("GET", "/partials/messages?archived=1")
            archived_view = body.decode("utf-8", "replace")
            c.ok(
                "workflow archived row moved",
                f"/partials/messages/{sent_thread}" not in unarchived_view
                and f"/partials/messages/{sent_thread}" in archived_view,
                "row did not move to the archived view",
            )
            status, _, _ = s.request("POST", f"/messages/{sent_thread}/archive?on=0")
            c.ok("workflow unarchive 204", status == 204, f"got {status}")
            status, _, _ = s.request("POST", f"/messages/{sent_thread}/mute?on=1")
            c.ok("workflow mute row-lane 204", status == 204, f"got {status}")
            status, body, _ = s.request(
                "POST",
                f"/messages/{sent_thread}/mute?on=0",
                b"",
                "",
                {"HX-Target": "wp-thread-head"},
            )
            c.ok(
                "workflow mute head-variant swaps the head",
                status == 200 and b"wp-thread-head" in body,
                f"got {status}",
            )

        # (g) Snippets: save renders the settings list, an over-long body
        # is refused, delete removes the row.
        status, body, _ = s.request(
            "POST", "/snippets/save", urlencode({"body": "smoke snippet workflow"}).encode(), "application/x-www-form-urlencoded"
        )
        c.ok(
            "workflow snippet saved",
            status == 200 and b"smoke snippet workflow" in body,
            f"{status}",
        )
        status, _, _ = s.request(
            "POST", "/snippets/save", urlencode({"body": "x" * 4000}).encode(), "application/x-www-form-urlencoded"
        )
        c.ok("workflow snippet over-long 422", status == 422, f"got {status}")
        status, body, _ = s.request("GET", "/partials/settings")
        snippet_ids = re.findall(r'hx-post="/snippets/delete\?id=([^"]+)"', body.decode("utf-8", "replace"))
        if snippet_ids:
            status, body, _ = s.request("POST", f"/snippets/delete?id={snippet_ids[-1]}")
            c.ok(
                "workflow snippet deleted",
                status == 200 and f"/snippets/delete?id={snippet_ids[-1]}".encode() not in body,
                f"got {status}",
            )
        else:
            c.ok("workflow snippet deleted", False, "no snippet id in settings partial")

        # (h) Contacts: save, dup-rename (one row), export, import, delete.
        contact_number = "+441632960222"
        status, body, _ = s.request(
            "POST",
            "/contacts/save",
            urlencode({"name": "Smoke Contact", "number": contact_number}).encode(),
            "application/x-www-form-urlencoded",
        )
        c.ok(
            "workflow contact saved",
            status == 200 and b"Smoke Contact" in body,
            f"{status}",
        )
        status, body, _ = s.request(
            "POST",
            "/contacts/save",
            urlencode({"name": "Smoke Renamed", "number": contact_number}).encode(),
            "application/x-www-form-urlencoded",
        )
        contacts_partial = body.decode("utf-8", "replace")
        c.ok(
            "workflow contact upsert keeps one row",
            status == 200
            and contacts_partial.count("Smoke Renamed") == 1
            and "Smoke Contact" not in contacts_partial.replace("Smoke Renamed", ""),
            f"{status} rows changed unexpectedly",
        )
        status, body, _ = s.request("GET", "/contacts/export")
        c.ok(
            "workflow contact exported",
            status == 200 and b"Smoke Renamed" in body,
            f"got {status}",
        )
        vcard = (
            "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Imported One\r\n"
            f"TEL;TYPE=CELL:{contact_number[:5]}33301\r\nEND:VCARD\r\n"
            "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Imported Two\r\n"
            f"TEL;TYPE=CELL:{contact_number[:5]}33302\r\nEND:VCARD\r\n"
        ).encode()
        body_bytes, ctype = multipart_body([], [("vcard", "two.vcf", "text/vcard", vcard)])
        status, body, _ = s.request("POST", "/contacts/import", body_bytes, ctype)
        imported = body.decode("utf-8", "replace")
        c.ok(
            "workflow contacts imported",
            status == 200 and "Imported One" in imported and "Imported Two" in imported,
            f"got {status}",
        )
        status, body, _ = s.request("GET", "/partials/contacts")
        delete_ids = re.findall(r'hx-post="/contacts/delete\?id=([^"]+)"', body.decode("utf-8", "replace"))
        if delete_ids:
            status, body, _ = s.request("POST", f"/contacts/delete?id={delete_ids[0]}")
            c.ok(
                "workflow contact deleted",
                status == 200 and f"/contacts/delete?id={delete_ids[0]}".encode() not in body,
                f"got {status}",
            )
        else:
            c.ok("workflow contact deleted", False, "no contact id in panel")

        # (i) MMS inbound with an attachment: the thread shows the honest
        # attachment link, the stream is session-fenced, and the served
        # bytes carry the declared type.
        png_1px = base64.b64encode(
            bytes.fromhex(
                "89504e470d0a1a0a0000000d494844520000000100000001080600000"
                "01f15c4890000000d49444154789c626001000000ffff030000060005"
                "57bfabd40000000049454e44ae426082"
            )
        ).decode()
        status, _, _ = s.hook(
            "/hooks/message",
            {
                "owner": "1001",
                "from": "+441632960444",
                "body": "smoke mms",
                "attachments": [
                    {"name": "dot.png", "mime_type": "image/png", "data_base64": png_1px}
                ],
            },
        )
        c.ok("workflow mms hook 202", status == 202, f"got {status}")
        _, body, _ = s.request("GET", "/partials/messages")
        mms_rows = re.findall(r'hx-get="/partials/messages/([^"]+)"', body.decode("utf-8", "replace"))
        attachment_id = ""
        # Only the MMS thread carries an attachment row, so hunt for the
        # wp-attachment marker directly (the rail preview would match any
        # thread's view — pinned rows sort first and broke the naive hunt).
        for row in mms_rows:
            _, tbody, _ = s.request("GET", f"/partials/messages/{row}")
            ttext = tbody.decode("utf-8", "replace")
            match = re.search(r'class="wp-attachment[^"]*"[^>]*href="/attachments/([^"]+)"', ttext) or re.search(
                r'href="/attachments/([^"]+)"[^>]*class="wp-attachment', ttext
            )
            if match:
                attachment_id = match.group(1)
                break
        c.ok("workflow mms attachment linked", bool(attachment_id), "no /attachments link in the thread")
        if attachment_id:
            status, _, _ = anon_export.request("GET", f"/attachments/{attachment_id}")
            c.ok("workflow attachment anonymous 401", status == 401, f"got {status}")
            status, body, headers = s.request("GET", f"/attachments/{attachment_id}")
            atype = next((v for k, v in headers.items() if k.lower() == "content-type"), "")
            c.ok(
                "workflow attachment streams its type",
                status == 200 and atype == "image/png",
                f"{status} {atype!r}",
            )

        # (j) Fax lifecycle on the loopback gateway: send a PDF-shaped
        # document, the row appears, the document streams back, an inbound
        # fax lands as a received row, and resend never renders on a
        # non-failed row.
        pdf = b"%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n"
        body_bytes, ctype = multipart_body(
            [("to", "+441632960333")], [("document", "smoke.pdf", "application/pdf", pdf)]
        )
        status, body, _ = s.request("POST", "/fax/send", body_bytes, ctype)
        fax_partial = body.decode("utf-8", "replace")
        c.ok(
            "workflow fax sent row",
            status == 200 and "wp-fax-row" in fax_partial,
            f"got {status}",
        )
        fax_doc_ids = re.findall(r'href="/fax/([^"]+)/document"', fax_partial)
        if fax_doc_ids:
            status, body, headers = s.request("GET", f"/fax/{fax_doc_ids[0]}/document")
            dtype = next((v for k, v in headers.items() if k.lower() == "content-type"), "")
            c.ok(
                "workflow fax document streams",
                status == 200 and "pdf" in dtype.lower() and body.startswith(b"%PDF"),
                f"{status} {dtype!r}",
            )
        else:
            c.ok("workflow fax document streams", False, "no document link in fax panel")
        status, _, _ = s.hook(
            "/hooks/fax",
            {
                "owner": "1001",
                "from": "+441632960555",
                "pages": 1,
                "pdf_base64": base64.b64encode(pdf).decode(),
            },
        )
        c.ok("workflow inbound fax 202", status == 202, f"got {status}")
        _, body, _ = s.request("GET", "/partials/fax")
        fax_after = body.decode("utf-8", "replace")
        c.ok(
            "workflow inbound fax received row",
            "wp-fax-in" in fax_after and "/resend" not in fax_after,
            "no inbound row or a resend button rendered on a healthy row",
        )

    # 14. Phone-api proxy fails closed while the PBX API is not configured.
    status, _, _ = s.request("GET", "/phone-api/history?limit=5")
    c.ok("phone-api proxy fails closed", status in (503, 401, 404), f"got {status}")

    # 15. The TLS-fronted login shape. Without csrf.trusted_* the scheme
    # mismatch (https Origin vs plain-http listener) 403s EVERY login —
    # the 2026-09-19 prod outage; with the fronting configured it passes.
    fronted_plain = Smoke(s.base)
    fronted_plain.request("GET", "/")
    status, _, _ = fronted_login(fronted_plain)
    c.ok("fronted login 403s when unconfigured", status == 403, f"got {status}")

    if boot_configured is None:
        print("smoke: fronted-configured probe skipped (--base mode)")
    else:
        base2, stop_configured = boot_configured()
        try:
            fronted_trusted = Smoke(base2)
            _, body2, hdrs2 = fronted_trusted.request("GET", "/")
            m2 = re.search(
                r'name="csrf-token" content="([^"]+)"',
                body2.decode("utf-8", "replace"),
            )
            if m2:
                fronted_trusted.csrf = m2.group(1)
            status, _, _ = fronted_login(fronted_trusted, secure_csrf_cookie(hdrs2))
            c.ok("fronted login 201 when configured", status == 201, f"got {status}")
        finally:
            stop_configured()

    if stop is not None and not stop.is_set():
        stop.set()
    skipped = f", {c.skipped} skipped (foreign mode)" if c.skipped else ""
    print(f"smoke: {c.passed} passed, {len(c.failures)} failed{skipped}")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


class _FakeProviderHandler(http.server.BaseHTTPRequestHandler):
    """The webhook-gateway bridge stand-in: records outbound receipts and
    answers them with provider refs. `mode` switches the verdict:
    ok (200 + receipt), refuse (422 provider refusal), down (503)."""

    mode = "ok"
    hits: ClassVar[list[dict[str, str]]] = []
    counter: ClassVar[int] = 0

    def do_POST(self) -> None:
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length)
        _FakeProviderHandler.counter += 1
        _FakeProviderHandler.hits.append(
            {
                "path": self.path,
                "auth": self.headers.get("Authorization", ""),
                "ctype": self.headers.get("Content-Type", ""),
                "body": raw.decode("utf-8", "replace"),
            }
        )
        if _FakeProviderHandler.mode == "refuse":
            payload = b"destination not allowed by provider policy"
            self.send_response(422)
            ref = ""
        elif _FakeProviderHandler.mode == "down":
            payload = b"upstream unavailable"
            self.send_response(503)
            ref = ""
        else:
            ref = f"gw-smoke-{_FakeProviderHandler.counter}"
            payload = json.dumps({"provider_ref": ref}).encode()
            self.send_response(200)
        _FakeProviderHandler.hits[-1]["ref"] = ref
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def log_message(self, *_args: object) -> None:
        pass


def _start_fake_provider() -> tuple[str, Callable[[], None]]:
    _FakeProviderHandler.mode = "ok"
    _FakeProviderHandler.hits = []
    _FakeProviderHandler.counter = 0
    httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), _FakeProviderHandler)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()

    def stop() -> None:
        httpd.shutdown()
        httpd.server_close()

    return f"http://127.0.0.1:{httpd.server_port}", stop


def _scenario_boot(
    binary: str, workdir: str, datadir: str, env_extra: dict[str, str], name: str
) -> tuple[subprocess.Popen, str, Callable[[], None]]:
    """The shared scenario boot: free port, loopback env, readiness poll."""

    def stop_server(srv: subprocess.Popen) -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    port = free_port()
    env = dict(os.environ)
    env.update(
        {
            "WEBPHONE_ADDR": f"127.0.0.1:{port}",
            "WEBPHONE_DATA_DIR": f"{workdir}/{datadir}",
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
        }
    )
    env.update(env_extra)
    srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
        [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
    )
    base = f"http://127.0.0.1:{port}"
    deadline = time.monotonic() + TIMEOUT
    while time.monotonic() < deadline:
        try:
            urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
            return srv, base, lambda: stop_server(srv)
        except OSError:
            time.sleep(0.1)
    stop_server(srv)
    raise RuntimeError(f"{name}: server did not become ready")


def _thread_id_for(s: Smoke, needle: str) -> str:
    """Find the thread id whose DETAIL (not rail preview) matches needle."""
    _, body, _ = s.request("GET", "/partials/messages")
    rows = re.findall(r'hx-get="/partials/messages/([^"]+)"', body.decode("utf-8", "replace"))
    for row in rows:
        _, tbody, _ = s.request("GET", f"/partials/messages/{row}")
        if needle in tbody.decode("utf-8", "replace"):
            return row
    return ""


def msglifecycle_scenario(binary: str, workdir: str) -> int:
    """The outbound messaging lifecycle over the REAL webhook gateway.

    Loopback mode makes every send succeed, so the honest transport test
    points the gateway at a controllable fake provider: the send receipt,
    the delivered verdict, the refusal (a provider 4xx with its reason),
    the outage class (5xx -> retryable), the retry lane, hook replay
    idempotence, and the SSE events along the way.
    """
    c = Check()
    provider, stop_provider = _start_fake_provider()
    srv, base, stop = _scenario_boot(
        binary,
        workdir,
        "data-msglifecycle",
        {
            "WEBPHONE_GATEWAY__MODE": "webhook",
            "WEBPHONE_GATEWAY__WEBHOOK_URL": provider,
        },
        "msglifecycle scenario",
    )
    try:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace"))
        if m:
            s.csrf = m.group(1)
        c.ok("msglifecycle: login accepted", s.login(), "POST /api/session != 201")
        sink: list[tuple[str, str]] = []
        ev_stop = threading.Event()
        reader = threading.Thread(target=s.sse_events, args=(ev_stop, sink), daemon=True)
        reader.start()
        time.sleep(0.4)

        body_bytes, ctype = multipart_body(
            [("to", "+441632960777"), ("body", "lifecycle one")]
        )
        status, body, _ = s.request("POST", "/messages/send", body_bytes, ctype)
        c.ok(
            "msglifecycle: send accepted",
            status == 200 and b"messages-panel" in body,
            f"got {status}",
        )
        receipts = [h for h in _FakeProviderHandler.hits if h["path"] == "/message"]
        c.ok(
            "msglifecycle: provider saw the message",
            len(receipts) >= 1
            and receipts[0]["auth"] == "Bearer test-secret"
            and "+441632960777" in receipts[0]["body"],
            f"hits={_FakeProviderHandler.hits[:1]}",
        )
        data = s.wait_for(sink, "threads", time.monotonic() + TIMEOUT)
        c.ok("msglifecycle: SSE threads on send", data is not None, "no threads event")

        ref = ""
        for hit in _FakeProviderHandler.hits:
            if hit.get("ref"):
                ref = str(hit["ref"])
        c.ok("msglifecycle: receipt ref known", bool(ref), "no provider_ref captured")
        thread = _thread_id_for(s, "lifecycle one")
        c.ok("msglifecycle: outbound bubble persisted", bool(thread), "thread not found")

        if ref:
            status, _, _ = s.hook(
                "/hooks/message/status", {"provider_ref": ref, "status": "delivered"}
            )
            c.ok("msglifecycle: delivered hook 202", status == 202, f"got {status}")
            _, tbody, _ = s.request("GET", f"/partials/messages/{thread}")
            c.ok(
                "msglifecycle: delivered badge lands",
                b"wp-status-delivered" in tbody,
                "no delivered badge in the transcript",
            )
            data = s.wait_for(sink, "thread", time.monotonic() + TIMEOUT)
            c.ok("msglifecycle: SSE thread on verdict", data is not None, "no thread event")
            status, _, _ = s.hook(
                "/hooks/message/status", {"provider_ref": ref, "status": "delivered"}
            )
            c.ok("msglifecycle: hook replay 202", status == 202, f"got {status}")

        # Refusal: the provider's 4xx reason is user feedback (422) and
        # the bubble keeps it with NO retry affordance.
        _FakeProviderHandler.mode = "refuse"
        body_bytes, ctype = multipart_body(
            [("to", "+441632960888"), ("body", "lifecycle refused")]
        )
        status, body, _ = s.request("POST", "/messages/send", body_bytes, ctype)
        c.ok(
            "msglifecycle: refusal answers 422 with the reason",
            status == 422 and b"provider policy" in body,
            f"got {status}: {body[:120]!r}",
        )
        refused_thread = _thread_id_for(s, "lifecycle refused")
        if refused_thread:
            _, tbody, _ = s.request("GET", f"/partials/messages/{refused_thread}")
            ttext = tbody.decode("utf-8", "replace")
            c.ok(
                "msglifecycle: refused bubble keeps reason, no retry",
                "wp-status-failed" in ttext and "wp-retry" not in ttext,
                "failed badge/retry affordance wrong",
            )
        else:
            c.ok(
                "msglifecycle: refused bubble keeps reason, no retry",
                False,
                "refused thread not found",
            )

        # Outage class: a provider 5xx is infrastructure (502) and the
        # bubble stays retryable.
        _FakeProviderHandler.mode = "down"
        body_bytes, ctype = multipart_body(
            [("to", "+441632960999"), ("body", "lifecycle outage")]
        )
        status, body, _ = s.request("POST", "/messages/send", body_bytes, ctype)
        c.ok("msglifecycle: outage answers 502", status == 502, f"got {status}")
        outage_thread = _thread_id_for(s, "lifecycle outage")
        if outage_thread:
            _, tbody, _ = s.request("GET", f"/partials/messages/{outage_thread}")
            ttext = tbody.decode("utf-8", "replace")
            c.ok(
                "msglifecycle: outage bubble retryable",
                "wp-status-failed" in ttext and "wp-retry" in ttext,
                "retry affordance missing on the transient failure",
            )
            # Retry once the provider recovers (the retry lane rides
            # ?thread= exactly like the rendered form).
            _FakeProviderHandler.mode = "ok"
            body_bytes, ctype = multipart_body([("body", "lifecycle retry")])
            status, body, _ = s.request(
                "POST", f"/messages/send?thread={outage_thread}", body_bytes, ctype
            )
            c.ok(
                "msglifecycle: retry send accepted",
                status == 200 and b"wp-bubble" in body,
                f"got {status}",
            )
            status, _, _ = s.hook(
                "/hooks/message/status",
                {"provider_ref": f"gw-smoke-{_FakeProviderHandler.counter}", "status": "failed", "error": "carrier rejected"},
            )
            c.ok("msglifecycle: failed verdict with reason 202", status == 202, f"got {status}")
            _, tbody, _ = s.request("GET", f"/partials/messages/{outage_thread}")
            c.ok(
                "msglifecycle: failed detail lands",
                b"carrier rejected" in tbody,
                "provider reason missing from the transcript",
            )
        else:
            c.ok("msglifecycle: outage bubble retryable", False, "outage thread not found")

        ev_stop.set()
    finally:
        stop()
        stop_provider()
    print(f"msglifecycle scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def ratelimit_scenario(binary: str, workdir: str) -> int:
    """The limiter fences on their own boot (buckets are per-IP, so a
    dedicated datadir keeps the exhaustion run from poisoning any other
    scenario's traffic): login lockout with Retry-After, the compose
    bucket's .wp-error fragment, the hooks lane, and the csrf lane."""
    c = Check()
    srv, base, stop = _scenario_boot(binary, workdir, "data-ratelimit", {}, "ratelimit scenario")
    try:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace"))
        if m:
            s.csrf = m.group(1)
        c.ok("ratelimit: clean login first", s.login(), "the clean-bucket login failed")

        # Login lockout: 30/min + burst 5. Credentials are not verified
        # on the loopback boot, so the consuming request is a SUCCESSFUL
        # login — each rotates the CSRF token, and the next one is scraped
        # from the layout meta (the shell carries the current token on
        # every page; /api/csrf would pollute the csrf lane below).
        attacker = Smoke(base)
        _, page, _ = attacker.request("GET", "/login")
        m = re.search(r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace"))
        if m:
            attacker.csrf = m.group(1)
        saw_429 = False
        retry_after = ""
        for _ in range(40):
            status, _, headers = attacker.request(
                "POST",
                "/api/session",
                json.dumps({"extension": "1001", "password": "pw"}).encode(),
                "application/json",
            )
            if status == 429:
                saw_429 = True
                retry_after = headers.get("Retry-After", "")
                break
            _, page, _ = attacker.request("GET", "/")
            m = re.search(
                r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
            )
            if m:
                attacker.csrf = m.group(1)
        c.ok("ratelimit: login bucket 429s", saw_429, "40 logins never tripped the limiter")
        c.ok("ratelimit: 429 carries Retry-After", bool(retry_after), "missing header")

        # Compose bucket: 20/min burst 10 -> the 11th rapid send answers
        # 429 with the .wp-error fragment htmx swaps into #wp-tab-error.
        limited_body = b""
        for i in range(12):
            body_bytes, ctype = multipart_body(
                [("to", f"+44163296011{i}"), ("body", "rl")]
            )
            status, body, headers = s.request("POST", "/messages/send", body_bytes, ctype)
            if status == 429:
                limited_body = body
                c.ok("ratelimit: compose 429 has Retry-After", bool(headers.get("Retry-After")), "missing header")
                break
        c.ok(
            "ratelimit: compose bucket 429s",
            b"wp-error" in limited_body,
            f"12 rapid sends never answered a .wp-error 429: {limited_body[:80]!r}",
        )

        # Hooks lane: 60/min burst 60 -> 61 rapid bad-secret probes trip
        # the fence before the secret gate is even consulted.
        hook_429 = False
        for _ in range(63):
            status, _, _ = s.request(
                "POST", "/hooks/message", b"{}", "application/json", {"Authorization": "Bearer wrong"}
            )
            if status == 429:
                hook_429 = True
                break
        c.ok("ratelimit: hooks lane 429s", hook_429, "63 rapid hooks never tripped")

        # CSRF lane: same 60/60 budget on GET /api/csrf.
        csrf_429 = False
        for _ in range(63):
            status, _, _ = s.request("GET", "/api/csrf")
            if status == 429:
                csrf_429 = True
                break
        c.ok("ratelimit: csrf lane 429s", csrf_429, "63 rapid csrf fetches never tripped")
    finally:
        stop()
    print(f"ratelimit scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def adminswitch_scenario(binary: str, workdir: str) -> int:
    """The admin surface (read-only, all lines) and the line switcher on
    a config-file boot with the identity layer: one admin account owning
    two mapped extensions, one plain account, and the isolation fences
    between them."""
    c = Check()

    datadir = f"{workdir}/data-adminswitch"
    os.makedirs(datadir, exist_ok=True)
    password_files: dict[str, str] = {}
    for ext in ("1001", "1002", "1003"):
        path = f"{workdir}/pw-{ext}"
        with open(path, "w", encoding="utf-8") as fh:
            fh.write("pw\n")
        password_files[ext] = path
    cfg = {
        "addr": "127.0.0.1:18099",
        "data_dir": datadir,
        "gateway": {"webhook_secret": "test-secret"},  # nosec B105 - fixture secret for the self-booted smoke server
        "auth": {
            "users": {
                "admin@example.com": {"extensions": ["1001", "1002"], "admin": True},
                "user@example.com": {"extensions": ["1003"]},
            },
            "extension_password_files": password_files,
            "passkey": {
                "rp_id": "webphone.test",
                "rp_origins": ["https://webphone.test"],
            },
        },
    }
    cfg_path = f"{workdir}/config-adminswitch.json"
    port = free_port()
    cfg["addr"] = f"127.0.0.1:{port}"
    with open(cfg_path, "w", encoding="utf-8") as fh:
        json.dump(cfg, fh)
    env = dict(os.environ)
    env["WEBPHONE_CONFIG"] = cfg_path
    env["WEBPHONE_GATEWAY__WEBHOOK_SECRET"] = "test-secret"  # nosec B105 - fixture secret for the self-booted smoke server
    srv = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
        [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
    )
    base = f"http://127.0.0.1:{port}"
    ready = False
    deadline = time.monotonic() + TIMEOUT
    while time.monotonic() < deadline:
        try:
            urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
            ready = True
            break
        except OSError:
            time.sleep(0.1)
    if not ready:
        srv.kill()
        print("adminswitch scenario: 0 passed, 1 failed")
        print("  FAILED: adminswitch: server did not become ready")
        return 1

    def stop() -> None:
        srv.terminate()
        try:
            srv.wait(timeout=5)
        except subprocess.TimeoutExpired:
            srv.kill()

    try:
        def fresh_login(extension: str) -> Smoke:
            client = Smoke(base)
            _, page, _ = client.request("GET", "/login")
            m = re.search(
                r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
            )
            if m:
                client.csrf = m.group(1)
            client.login(extension)
            return client

        # Seed a thread owned by the PLAIN line, then prove the admin
        # sees it cross-line while the plain line cannot see the admin's.
        admin = fresh_login("1001")
        plain = fresh_login("1003")
        c.ok("adminswitch: admin login", any(n.name == "webphone_session" for n in admin.jar), "no session")
        c.ok("adminswitch: plain login", any(n.name == "webphone_session" for n in plain.jar), "no session")
        status, _, _ = plain.hook(
            "/hooks/message", {"owner": "1003", "from": "+441632960123", "body": "plain line thread"}
        )
        c.ok("adminswitch: plain thread seeded", status == 202, f"got {status}")

        status, body, _ = admin.request("GET", "/admin")
        admin_page = body.decode("utf-8", "replace")
        c.ok(
            "adminswitch: admin page renders",
            status == 200 and 'id="admin-panel"' in admin_page,
            f"got {status}",
        )
        status, body, _ = admin.request("GET", "/partials/admin")
        admin_partial = body.decode("utf-8", "replace")
        c.ok(
            "adminswitch: admin sees the plain line's thread",
            status == 200 and "+441632960123" in admin_partial,
            f"cross-line row missing: {admin_partial[:80]!r}",
        )
        rows = re.findall(r'hx-get="/partials/admin/threads/([^"]+)"', admin_partial)
        if rows:
            status, body, _ = admin.request("GET", f"/admin/threads/{rows[0]}")
            c.ok(
                "adminswitch: admin deep link opens the thread",
                status == 200 and 'id="thread-transcript"' in body.decode("utf-8", "replace"),
                f"got {status}",
            )
        else:
            c.ok("adminswitch: admin deep link opens the thread", False, "no admin row link")

        status, body, _ = plain.request("GET", "/admin")
        c.ok(
            "adminswitch: non-admin gets the styled 404",
            status == 404 and b"wp-notfound" in body,
            f"got {status}",
        )
        status, body, _ = plain.request("GET", "/partials/nav")
        c.ok(
            "adminswitch: non-admin nav hides the tab",
            status == 200 and b"/admin" not in body,
            "admin link rendered for a non-admin",
        )

        # Line switcher: the admin's two mapped extensions re-mint the
        # session; the other account's line and unmapped extensions stay
        # out of reach.
        status, body, _ = admin.request(
            "POST",
            "/api/session/switch",
            json.dumps({"extension": "1002"}).encode(),
            "application/json",
        )
        c.ok("adminswitch: switch to mapped line", status in (200, 204), f"got {status}: {body[:80]!r}")
        admin.adopt_csrf()
        status, body, _ = admin.request("GET", "/api/session")
        switched = json.loads(body) if body else {}
        c.ok(
            "adminswitch: switched identity answers 1002",
            switched.get("extension") == "1002",
            f"identity={switched.get('extension')}",
        )
        status, body, _ = admin.request("GET", "/partials/messages")
        c.ok(
            "adminswitch: switched line isolates other owners",
            status == 200 and b"+441632960123" not in body,
            "the 1003 thread leaked across the line switch",
        )
        status, _, _ = admin.request(
            "POST",
            "/api/session/switch",
            json.dumps({"extension": "1003"}).encode(),
            "application/json",
        )
        c.ok("adminswitch: foreign line 403s", status == 403, f"got {status}")
        status, _, _ = plain.request(
            "POST",
            "/api/session/switch",
            json.dumps({"extension": "1003"}).encode(),
            "application/json",
        )
        c.ok("adminswitch: extension login cannot switch", status == 403, f"got {status}")
    finally:
        stop()
    print(f"adminswitch scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


class _FakeAgentHandler(http.server.BaseHTTPRequestHandler):
    """The textagent review-API stand-in: drafts with counting ids, the
    approve/discard verdicts, mode pinning, and the reviews listing."""

    hits: ClassVar[list[dict[str, str]]] = []
    counter: ClassVar[int] = 0

    def _json(self, status: int, payload: dict) -> None:
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:
        _FakeAgentHandler.hits.append({"path": self.path, "method": "GET"})
        self._json(
            200,
            {
                "drafts": [],
                "frozen": [],
                "suppressed": [],
                "mode": "off",
                "mode_source": "config",
                "default_mode": "off",
            },
        )

    def do_POST(self) -> None:
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length).decode("utf-8", "replace")
        _FakeAgentHandler.hits.append({"path": self.path, "method": "POST", "body": body})
        _FakeAgentHandler.counter += 1
        draft_id = f"d-smoke-{_FakeAgentHandler.counter}"
        if self.path.startswith("/reviews/") and self.path.endswith("/discard"):
            self._json(200, {})
            return
        if self.path.startswith("/reviews/") and self.path.endswith("/approve"):
            self._json(200, {"id": draft_id, "text": "smoke agent draft", "status": "approved"})
            return
        if self.path == "/mode":
            self._json(200, {"mode": "off", "mode_source": "runtime", "default_mode": "off"})
            return
        self._json(
            200,
            {
                "id": draft_id,
                "text": "smoke agent draft",
                "status": "draft",
                "confidence": 80,
            },
        )

    def log_message(self, *_args: object) -> None:
        pass


def _start_fake_agent() -> tuple[str, Callable[[], None]]:
    _FakeAgentHandler.hits = []
    _FakeAgentHandler.counter = 0
    httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), _FakeAgentHandler)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()

    def stop() -> None:
        httpd.shutdown()
        httpd.server_close()

    return f"http://127.0.0.1:{httpd.server_port}", stop


def drafts_scenario(binary: str, workdir: str) -> int:
    """The agent drafts workflow against a fake review API: the compose
    button's draft lands prefilled in the composer, approve SENDS it into
    the thread, discard drops it, the mode switch round-trips, and the
    review/disposition surfaces answer their absent-value contracts."""
    c = Check()
    agent, stop_agent = _start_fake_agent()
    agent_secret = f"{workdir}/agent-secret"
    with open(agent_secret, "w", encoding="utf-8") as fh:
        fh.write("agent-secret")  # nosec B105 - fixture secret for the fake review API
    srv, base, stop = _scenario_boot(
        binary,
        workdir,
        "data-drafts",
        {
            "WEBPHONE_TEXTAGENT__URL": agent,
            "WEBPHONE_TEXTAGENT__SECRET_FILE": agent_secret,
        },
        "drafts scenario",
    )
    try:
        s = Smoke(base)
        _, page, _ = s.request("GET", "/login")
        m = re.search(r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace"))
        if m:
            s.csrf = m.group(1)
        c.ok("drafts: login accepted", s.login(), "POST /api/session != 201")
        _, body, _ = s.request("GET", "/config.js")
        c.ok("drafts: config.js advertises the agent", b'"agent":true' in body, body[:120].decode("utf-8", "replace"))

        body_bytes, ctype = multipart_body(
            [("to", "+441632960555"), ("body", "say hi about the invoice")]
        )
        status, body, _ = s.request("POST", "/messages/ai-draft", body_bytes, ctype)
        c.ok(
            "drafts: ai-draft prefills the composer",
            status == 200 and b"smoke agent draft" in body,
            f"got {status}: {body[:100]!r}",
        )
        agent_hits = [h for h in _FakeAgentHandler.hits if h["path"] == "/draft"]
        c.ok(
            "drafts: agent saw the ask",
            len(agent_hits) >= 1 and "invoice" in agent_hits[0].get("body", ""),
            f"hits={agent_hits[:1]}",
        )

        status, body, _ = s.request(
            "POST",
            "/messages/drafts/d-smoke-1/approve",
            urlencode({"text": ""}).encode(),
            "application/x-www-form-urlencoded",
        )
        c.ok("drafts: approve answers the panel", status == 200, f"got {status}")
        approve_hits = [h for h in _FakeAgentHandler.hits if h["path"].endswith("/approve")]
        c.ok("drafts: agent saw the approve", len(approve_hits) >= 1, "no approve hit")
        _, body, _ = s.request("GET", "/partials/messages")
        listing = body.decode("utf-8", "replace")
        c.ok(
            "drafts: approved draft reached the thread",
            "+441632960555" in listing,
            "thread for the draft destination missing",
        )
        draft_thread = _thread_id_for(s, "smoke agent draft")
        c.ok("drafts: draft text landed as a bubble", bool(draft_thread), "draft bubble missing")

        body_bytes, ctype = multipart_body([("to", "+441632960556"), ("body", "hint")])
        status, _, _ = s.request("POST", "/messages/ai-draft", body_bytes, ctype)
        status, _, _ = s.request(
            "POST",
            "/messages/drafts/d-smoke-2/discard",
            b"",
            "",
        )
        c.ok("drafts: discard answers", status == 200, f"got {status}")
        discard_hits = [h for h in _FakeAgentHandler.hits if h["path"].endswith("/discard")]
        c.ok("drafts: agent saw the discard", len(discard_hits) >= 1, "no discard hit")

        status, body, _ = s.request(
            "POST",
            "/messages/drafts/mode",
            urlencode({"mode": "off"}).encode(),
            "application/x-www-form-urlencoded",
        )
        c.ok("drafts: mode toggle answers", status == 200, f"got {status}")

        status, body, _ = s.request("GET", "/api/calls/no-such-call/review")
        c.ok(
            "drafts: absent review reads empty",
            status == 200 and body == b"",
            f"got {status} {body[:40]!r}",
        )
        status, _, _ = s.request(
            "POST",
            "/api/calls/smoke-call-1/disposition",
            json.dumps(
                {"disposition": "missed", "number": "+441632960555", "direction": "in", "seconds": 0}
            ).encode(),
            "application/json",
        )
        c.ok("drafts: disposition records", status == 204, f"got {status}")
    finally:
        stop()
        stop_agent()
    print(f"drafts scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


class _FakeNtfyHandler(http.server.BaseHTTPRequestHandler):
    """The ntfy stand-in: records the JSON publish bodies the push fanout
    sends (topic in the body, priority per urgency)."""

    hits: ClassVar[list[dict[str, object]]] = []

    def do_POST(self) -> None:
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length)
        try:
            payload = json.loads(raw)
        except ValueError:
            payload = {"raw": raw.decode("utf-8", "replace")}
        _FakeNtfyHandler.hits.append(payload)
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", "2")
        self.end_headers()
        self.wfile.write(b"{}")

    def log_message(self, *_args: object) -> None:
        pass


def _start_fake_ntfy() -> tuple[str, Callable[[], None]]:
    _FakeNtfyHandler.hits = []
    httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), _FakeNtfyHandler)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()

    def stop() -> None:
        httpd.shutdown()
        httpd.server_close()

    return f"http://127.0.0.1:{httpd.server_port}", stop


def push_scenario(binary: str, workdir: str) -> int:
    """The ring-while-closed seam end-to-end against a fake ntfy: the
    off-boot answers 404 (the shell must not believe registration
    worked), the configured boot registers a device whose token is
    write-only, rotation replaces it, the ring hook wakes the device on
    its CURRENT topic with the urgent priority, the end state stands it
    down, and a deregistered device hears nothing."""
    c = Check()

    def fresh_login(base_url: str) -> Smoke:
        client = Smoke(base_url)
        _, page, _ = client.request("GET", "/login")
        m = re.search(
            r'name="csrf-token" content="([^"]+)"', page.decode("utf-8", "replace")
        )
        if m:
            client.csrf = m.group(1)
        client.login()
        return client

    # Off boot: no push config = the surface does not exist.
    srv_off, base_off, stop_off = _scenario_boot(
        binary, workdir, "data-push-off", {}, "push scenario"
    )
    try:
        s = fresh_login(base_off)
        status, _, _ = s.request("GET", "/api/push-devices")
        c.ok("push: off boot answers 404", status == 404, f"got {status}")
        status, _, _ = s.hook(
            "/hooks/ring",
            {"owner": "1001", "from": "+441632960000", "call_id": "fs-off", "state": "ring"},
        )
        c.ok("push: off boot ring stays 202 inert", status == 202, f"got {status}")
    finally:
        stop_off()

    ntfy, stop_ntfy = _start_fake_ntfy()
    srv, base, stop = _scenario_boot(
        binary, workdir, "data-push-on", {"WEBPHONE_PUSH__NTFY__URL": ntfy}, "push scenario"
    )
    try:
        s = fresh_login(base)
        status, body, _ = s.request("GET", "/api/push-devices")
        c.ok(
            "push: list starts empty",
            status == 200 and json.loads(body).get("devices") == [],
            f"got {status} {body[:60]!r}",
        )
        status, _, _ = s.request(
            "POST",
            "/api/push-devices",
            json.dumps(
                {"device_id": "smoke-device-1", "transport": "ntfy", "token": "smoketopic111"}
            ).encode(),
            "application/json",
        )
        c.ok("push: register 204", status == 204, f"got {status}")
        status, body, _ = s.request("GET", "/api/push-devices")
        listed = json.loads(body)
        c.ok(
            "push: list never echoes the token",
            status == 200
            and len(listed.get("devices", [])) == 1
            and b"smoketopic111" not in body,
            f"got {status} {body[:80]!r}",
        )
        status, _, _ = s.request(
            "POST",
            "/api/push-devices",
            json.dumps(
                {"device_id": "smoke-device-1", "transport": "ntfy", "token": "smoketopic222"}
            ).encode(),
            "application/json",
        )
        c.ok("push: rotation 204", status == 204, f"got {status}")
        status, body, _ = s.request("GET", "/api/push-devices")
        c.ok(
            "push: rotation keeps one row",
            len(json.loads(body).get("devices", [])) == 1,
            f"rows={json.loads(body).get('devices')}",
        )

        status, _, _ = s.hook(
            "/hooks/ring",
            {"owner": "1001", "from": "+441632960000", "call_id": "fs-ring-1", "state": "ring"},
        )
        c.ok("push: ring hook 202", status == 202, f"got {status}")
        deadline = time.monotonic() + TIMEOUT
        ring_hit = None
        while time.monotonic() < deadline:
            for hit in _FakeNtfyHandler.hits:
                if hit.get("topic") == "smoketopic222":
                    ring_hit = hit
                    break
            if ring_hit:
                break
            time.sleep(0.1)
        c.ok(
            "push: ring wakes the rotated topic urgently",
            ring_hit is not None and str(ring_hit.get("priority")) == "5" and "fs-ring-1" in str(ring_hit.get("message", "")),
            f"hits={_FakeNtfyHandler.hits[-2:]}",
        )
        status, _, _ = s.hook(
            "/hooks/ring",
            {"owner": "1001", "from": "+441632960000", "call_id": "fs-ring-1", "state": "end"},
        )
        c.ok("push: end state 202", status == 202, f"got {status}")
        deadline = time.monotonic() + TIMEOUT
        end_hit = None
        while time.monotonic() < deadline:
            for hit in _FakeNtfyHandler.hits:
                if hit.get("topic") == "smoketopic222" and str(hit.get("priority")) != "5":
                    end_hit = hit
                    break
            if end_hit:
                break
            time.sleep(0.1)
        c.ok("push: end stands down at lower priority", end_hit is not None, f"hits={_FakeNtfyHandler.hits[-2:]}")

        status, _, _ = s.request("DELETE", "/api/push-devices?id=smoke-device-1")
        c.ok("push: deregister 204", status == 204, f"got {status}")
        before = len(_FakeNtfyHandler.hits)
        s.hook(
            "/hooks/ring",
            {"owner": "1001", "from": "+441632960000", "call_id": "fs-ring-2", "state": "ring"},
        )
        time.sleep(0.6)
        c.ok(
            "push: deregistered device hears nothing",
            len(_FakeNtfyHandler.hits) == before,
            f"publishes after deregister: {_FakeNtfyHandler.hits[before:]}",
        )
    finally:
        stop()
        stop_ntfy()
    print(f"push scenario: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base", help="smoke an already-running server")
    parser.add_argument(
        "--bin",
        help="webphone binary to boot (default: bare 'go build', whose "
        "/version reports Go's pseudo-version — pair --expect-version "
        "with a released build: --bin $(nix build .#webphone))",
    )
    parser.add_argument(
        "--go", default="go", help="go toolchain command for --bin build"
    )
    parser.add_argument(
        "--expect-version",
        help="fail unless /version reports exactly this build version "
        "(leading v optional) — the post-deploy did-it-switch probe",
    )
    args = parser.parse_args()

    if args.expect_version and not args.base and not args.bin:
        parser.error(
            "--expect-version needs --base or --bin: the default bare "
            "`go build` binary reports Go's pseudo-version, never a release "
            "version — pass --bin $(nix build .#webphone)"
        )

    ensure_go_toolchain(args)

    if args.base:
        return run_checks(
            Smoke(args.base), foreign=True, expect_version=args.expect_version
        )

    port = free_port()
    workdir = tempfile.mkdtemp(prefix="webphone-smoke-")
    binary = args.bin
    if not binary:
        binary = f"{workdir}/webphone-bin"
        print(f"building {binary} …", flush=True)
        env = dict(os.environ)
        build = subprocess.run(  # nosec B603 - argv-list (no shell) build of the binary under test
            [args.go, "build", "-o", binary, "./cmd/webphone"],
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        if build.returncode != 0:
            print(f"build failed: {build.stderr[:400]}", file=sys.stderr)
            return 2

    def boot_configured() -> tuple[str, Callable[[], None]]:
        """Boot a second server with the TLS-fronting csrf shape configured
        (the NixOS module ships these defaults when caddy.enable)."""
        port2 = free_port()
        cfg = {
            "addr": f"127.0.0.1:{port2}",
            "data_dir": f"{workdir}/data-fronted",
            "gateway": {"webhook_secret": "test-secret"},  # nosec B105 - fixture secret for the self-booted smoke server
            "csrf": {
                "trusted_proxies": ["127.0.0.1"],
                "trusted_origins": ["https://pbx.test"],
            },
        }
        cfg_path = f"{workdir}/config-fronted.json"
        with open(cfg_path, "w", encoding="utf-8") as fh:
            json.dump(cfg, fh)
        env2 = dict(os.environ)
        env2["WEBPHONE_CONFIG"] = cfg_path
        srv2 = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
            [binary], env=env2, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base2 = f"http://127.0.0.1:{port2}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base2 + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                break
            except OSError:
                time.sleep(0.1)
        else:
            srv2.kill()
            raise RuntimeError("configured server did not become ready")

        def stop() -> None:
            srv2.terminate()
            try:
                srv2.wait(timeout=5)
            except subprocess.TimeoutExpired:
                srv2.kill()

        return base2, stop

    env = dict(os.environ)
    env.update(
        {
            "WEBPHONE_ADDR": f"127.0.0.1:{port}",
            "WEBPHONE_DATA_DIR": f"{workdir}/data",
            # Without a configured secret the hooks fail CLOSED (503) — the
            # suite exercises the open path, so it configures one.
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",  # nosec B105 - fixture secret for the self-booted smoke server
            # The dashboard is part of the product surface now: boot it
            # enabled so its checks are standing smoke coverage.
            "WEBPHONE_DASHBOARD__ENABLE": "true",
        }
    )
    server = subprocess.Popen(  # nosec B603 - argv-list (no shell) launch of the local binary under test
        [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
    )
    try:
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()  # nosec B310 - readiness probe of the local server under test
                break
            except OSError:
                time.sleep(0.1)
        else:
            print("server did not become ready", file=sys.stderr)
            return 2
        rc = run_checks(
            Smoke(base), boot_configured, expect_version=args.expect_version
        )
        rc = max(rc, restart_scenario(binary, workdir, port, env))
        rc = max(rc, boot_failure_scenario(binary, workdir, free_port()))
        rc = max(rc, asr_scenario(binary, workdir))
        rc = max(rc, dialgate_scenario(binary, workdir))
        rc = max(rc, agent_scenario(binary, workdir))
        rc = max(rc, faxcap_scenario(binary, workdir))
        rc = max(rc, metrics_scenario(binary, workdir))
        rc = max(rc, enroll_scenario(binary, workdir))
        rc = max(rc, msglifecycle_scenario(binary, workdir))
        rc = max(rc, ratelimit_scenario(binary, workdir))
        rc = max(rc, adminswitch_scenario(binary, workdir))
        rc = max(rc, drafts_scenario(binary, workdir))
        rc = max(rc, push_scenario(binary, workdir))
        return rc
    finally:
        server.terminate()
        try:
            server.wait(timeout=5)
        except subprocess.TimeoutExpired:
            server.kill()
        # The workdir is the script's OWN scratch (a fresh temp dir it
        # created), so the fleet's trash-only rule degrades safely:
        # prefer the mandated trash tool, fall back to shutil.rmtree
        # where it does not exist (GitHub runners — the 2026-10-30
        # smoke-in-CI first run died here AFTER passing every scenario).
        if shutil.which("trash"):
            subprocess.run(["trash", workdir], capture_output=True, check=False)  # nosec - argv-list cleanup via the mandated trash tool (B603 no shell/untrusted input, B607 PATH by design)
        else:
            shutil.rmtree(workdir, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
