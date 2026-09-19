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
import http.client
import json
import os
import re
import socket
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
from urllib.parse import urlparse

TIMEOUT = 10.0


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


class Smoke:
    def __init__(self, base: str) -> None:
        self.base = base.rstrip("/")
        self.host = urlparse(base).hostname or "127.0.0.1"
        self.port = urlparse(base).port or 80
        self.jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(self.jar))
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

    def login(self, extension: str = "1001", password: str = "pw") -> bool:
        status, _, _ = self.request(
            "POST",
            "/api/session",
            json.dumps({"extension": extension, "password": password}).encode(),
            "application/json",
        )
        return status == 201

    def hook(self, path: str, payload: dict) -> tuple[int, bytes]:
        body = json.dumps(payload).encode()
        return self.request("POST", path, body, "application/json",
                            {"Authorization": "Bearer test-secret"})

    def sse_events(self, stop: threading.Event, sink: list[str]) -> None:
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

    def wait_for(self, sink: list[str], event: str, deadline: float) -> str | None:
        while time.monotonic() < deadline:
            for name, data in sink:
                if name == event:
                    return data
            time.sleep(0.05)
        return None


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def run_checks(s: Smoke) -> int:
    c = s.check
    print(f"smoke against {s.base}")

    # 1. Shell page serves the island contract markers + CSRF meta.
    status, body, _ = s.request("GET", "/")
    page = body.decode("utf-8", "replace")
    c.ok("shell 200", status == 200, f"got {status}")
    c.ok("shell holds island login", 'id="login-view"' in page and 'id="reg-status"' in page,
         "island ids missing")
    match = re.search(r'name="csrf-token" content="([^"]+)"', page)
    c.ok("CSRF meta present", match is not None, "no csrf-token meta")
    if match:
        s.csrf = match.group(1)

    # 2. healthz is honest readiness JSON.
    status, body, _ = s.request("GET", "/healthz")
    c.ok("healthz ready", status == 200 and b'"ok"' in body.lower(), f"{status} {body[:80]!r}")

    # 3. version reports build metadata.
    status, body, _ = s.request("GET", "/version")
    c.ok("version reachable", status == 200 and b'"version"' in body, f"{status} {body[:80]!r}")

    # 4. config.js carries the PBX contract for the island.
    status, body, _ = s.request("GET", "/config.js")
    c.ok("config.js contract", status == 200 and b"window.PBX_CONFIG" in body and b"sipDomain" in body,
         f"{status}")

    # 5. Session POST is CSRF-gated: without the token it must not pass.
    plain = Smoke(s.base)
    plain.request("GET", "/")
    status, _, _ = plain.request(
        "POST", "/api/session",
        json.dumps({"extension": "1001", "password": "pw"}).encode(), "application/json")
    c.ok("session CSRF-gated", status in (403, 401), f"got {status}")

    # 6. Session with CSRF token issues the session cookie.
    c.ok("session login accepted", s.login(), "POST /api/session != 201")
    c.ok("session cookie issued", any(cookie.name == "webphone_session" for cookie in s.jar),
         "no webphone_session cookie")

    # 7. Signed-in SSE connects; anonymous SSE is rejected.
    stop = threading.Event()
    sink: list[tuple[str, str]] = []
    reader = threading.Thread(target=s.sse_events, args=(stop, sink), daemon=True)
    reader.start()
    time.sleep(0.4)
    c.ok("signed-in SSE stream live", len(sink) > 0, "no frames within 400ms")
    status, _, _ = Smoke(s.base).request("GET", "/events")
    c.ok("anonymous SSE rejected", status == 401, f"got {status}")

    # 8. Unknown hook path is a 404 even with the secret.
    status, _, _ = s.hook("/hooks/nope", {})
    c.ok("unknown hook 404", status == 404, f"got {status}")

    # 9. Inbound message webhook accepts.
    status, _, _ = s.hook("/hooks/message",
                          {"owner": "1001", "from": "+441632960961", "body": "smoke inbound"})
    c.ok("inbound webhook 202", status == 202, f"got {status}")

    # 10. The push lands live as a swap-safe threads fragment.
    data = s.wait_for(sink, "threads", time.monotonic() + TIMEOUT)
    c.ok("live threads event", data is not None, "no threads event in time")
    c.ok("threads fragment swap-safe",
         data is not None and "wp-thread-row" in data and "<section" not in data,
         f"payload {str(data)[:80]!r}")

    # 11. Thread row exists; opening it shows the transcript bubble region.
    _, body, _ = s.request("GET", "/partials/messages")
    rows = re.findall(r'hx-get="/partials/messages/([^"]+)"', body.decode())
    c.ok("thread list renders row", bool(rows), "no thread row")
    if rows:
        _, body, _ = s.request("GET", f"/partials/messages/{rows[0]}")
        transcript = body.decode()
        c.ok("transcript bubble", 'id="thread-transcript"' in transcript and 'sse-swap="thread"' in transcript,
             "swap region missing")

    # 12. Live thread push carries bubble fragments (not the whole panel).
    s.hook("/hooks/message", {"owner": "1001", "from": "+441632960961", "body": "smoke live"})
    data = s.wait_for(sink, "thread", time.monotonic() + TIMEOUT)
    c.ok("live thread event", data is not None, "no thread event in time")
    c.ok("thread fragment is bubbles only",
         data is not None and "wp-bubble" in data and "wp-compose" not in data and "<section" not in data,
         f"payload {str(data)[:80]!r}")

    # 13. Explicit read endpoint (the live-swap mark-read path).
    if rows:
        status, _, _ = s.request("POST", f"/messages/{rows[0]}/read")
        c.ok("live mark-read 204", status == 204, f"got {status}")

    # 14. Phone-api proxy fails closed while the PBX API is not configured.
    status, _, _ = s.request("GET", "/phone-api/history?limit=5")
    c.ok("phone-api proxy fails closed", status in (503, 401, 404), f"got {status}")

    stop.set()
    print(f"smoke: {c.passed} passed, {len(c.failures)} failed")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base", help="smoke an already-running server")
    parser.add_argument("--bin", help="webphone binary to boot (default: go build)")
    parser.add_argument("--go", default="go", help="go toolchain command for --bin build")
    args = parser.parse_args()

    if args.base:
        return run_checks(Smoke(args.base))

    port = free_port()
    workdir = tempfile.mkdtemp(prefix="webphone-smoke-")
    binary = args.bin
    if not binary:
        binary = f"{workdir}/webphone-bin"
        print(f"building {binary} …", flush=True)
        env = dict(os.environ)
        env["GOEXPERIMENT"] = "jsonv2"
        build = subprocess.run([args.go, "build", "-o", binary, "./cmd/webphone"],
                               env=env, capture_output=True, text=True)
        if build.returncode != 0:
            print(f"build failed: {build.stderr[:400]}", file=sys.stderr)
            return 2
    env = dict(os.environ)
    env.update({
        "WEBPHONE_ADDR": f"127.0.0.1:{port}",
        "WEBPHONE_DATA_DIR": f"{workdir}/data",
    })
    server = subprocess.Popen([binary], env=env,
                              stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        base = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()
                break
            except OSError:
                time.sleep(0.1)
        else:
            print("server did not become ready", file=sys.stderr)
            return 2
        return run_checks(Smoke(base))
    finally:
        server.terminate()
        try:
            server.wait(timeout=5)
        except subprocess.TimeoutExpired:
            server.kill()
        subprocess.run(["trash", workdir], capture_output=True)


if __name__ == "__main__":
    sys.exit(main())
