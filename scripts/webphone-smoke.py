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
import http.cookiejar
import json
import os
import pathlib
import re
import socket
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
from collections.abc import Callable
from urllib.parse import urlparse

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
    probe = subprocess.run(
        [args.go, "version"], capture_output=True, text=True, check=False
    )
    m = re.search(r"go version go(\d[0-9.]*)", probe.stdout)
    if m and _version_tuple(m.group(1)) >= floor:
        return
    print(
        "ambient go " + (m.group(1) if m else "unknown")
        + " < floor " + _fmt(floor)
        + "; re-executing via `nix develop -c` ...",
        flush=True,
    )
    env = dict(os.environ)
    env["WEBPHONE_SMOKE_REEXEC"] = "1"
    result = subprocess.run(
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

    def login(self, extension: str = "1001", password: str = "pw") -> bool:
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

    def hook(self, path: str, payload: dict) -> tuple[int, bytes]:
        body = json.dumps(payload).encode()
        return self.request(
            "POST",
            path,
            body,
            "application/json",
            {"Authorization": "Bearer test-secret"},
        )

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
        json.dumps({"extension": "1001", "password": "pw"}).encode(),
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
    _, body, _ = s.request("GET", "/")
    m = re.search(r'name="csrf-token" content="([^"]+)"', body.decode("utf-8", "replace"))
    if m:
        s.csrf = m.group(1)
    c.ok("restart: login accepted", s.login(), "POST /api/session != 201")
    cookie = "; ".join(f"{x.name}={x.value}" for x in s.jar)
    c.ok("restart: session cookie captured", bool(cookie), "jar empty")

    def boot() -> subprocess.Popen:
        srv = subprocess.Popen(
            [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base + "/healthz", timeout=1).read()
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
    print(
        f"restart scenario: {c.passed} passed, {len(c.failures)} failed"
    )
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def run_checks(
    s: Smoke,
    boot_configured: Callable[[], tuple[str, Callable[[], None]]] | None = None,
    foreign: bool = False,
) -> int:
    c = s.check
    print(f"smoke against {s.base}")

    # 1. Shell page serves the island contract markers + CSRF meta.
    status, body, _ = s.request("GET", "/")
    page = body.decode("utf-8", "replace")
    c.ok("shell 200", status == 200, f"got {status}")
    c.ok(
        "shell holds island login",
        'id="login-view"' in page and 'id="reg-status"' in page,
        "island ids missing",
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

    # 3. version reports build metadata.
    status, body, _ = s.request("GET", "/version")
    c.ok(
        "version reachable",
        status == 200 and b'"version"' in body,
        f"{status} {body[:80]!r}",
    )

    # 3b. openapi.json publishes the session API contract (3.1.0).
    status, body, _ = s.request("GET", "/openapi.json")
    c.ok(
        "openapi published",
        status == 200 and b'"openapi": "3.1.0"' in body,
        f"{status} {body[:80]!r}",
    )

    # 4. config.js carries the PBX contract for the island.
    status, body, _ = s.request("GET", "/config.js")
    c.ok(
        "config.js contract",
        status == 200 and b"window.PBX_CONFIG" in body and b"sipDomain" in body,
        f"{status}",
    )

    # 4b. /partials/nav contract (AGENTS-documented, now smoked): labels
    # render anonymously — badges NEVER do; the signed-in re-fetch below
    # is the only badge source. A fresh Smoke keeps it session-free even
    # mid-suite.
    anon_nav = Smoke(s.base)
    status, body, _ = anon_nav.request("GET", "/partials/nav")
    nav_anon = body.decode("utf-8", "replace")
    c.ok(
        "nav partial anonymous shape",
        status == 200 and "wp-nav-link" in nav_anon and "wp-nav-badge" not in nav_anon,
        f"{status} labels/badges wrong: {nav_anon[:80]!r}",
    )

    # 4c. Unknown paths render the STYLED 404: shell chrome around the
    # error panel, status stays 404 (error-page parity; works in --base
    # foreign mode, so the post-deploy probe validates prod too).
    status, body, _ = anon_nav.request("GET", "/definitely/not/a/path")
    nf = body.decode("utf-8", "replace")
    c.ok(
        "styled 404",
        status == 404
        and 'class="wp-panel"' in nf
        and "data-reload" in nf
        and 'id="login-view"' in nf,
        f"{status} bare-or-unstyled: {nf[:80]!r}",
    )

    # 5. Session POST is CSRF-gated: without the token it must not pass.
    plain = Smoke(s.base)
    plain.request("GET", "/")
    status, _, _ = plain.request(
        "POST",
        "/api/session",
        json.dumps({"extension": "1001", "password": "pw"}).encode(),
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
            json.dumps({"extension": "1001", "password": "definitely-wrong"}).encode(),
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

    stop_set = stop is not None and not stop.is_set()
    if stop_set:
        stop.set()
    skipped = f", {c.skipped} skipped (foreign mode)" if c.skipped else ""
    print(f"smoke: {c.passed} passed, {len(c.failures)} failed{skipped}")
    for failure in c.failures:
        print(f"  FAILED: {failure}")
    return 1 if c.failures else 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base", help="smoke an already-running server")
    parser.add_argument("--bin", help="webphone binary to boot (default: go build)")
    parser.add_argument(
        "--go", default="go", help="go toolchain command for --bin build"
    )
    args = parser.parse_args()

    ensure_go_toolchain(args)

    if args.base:
        return run_checks(Smoke(args.base), foreign=True)

    port = free_port()
    workdir = tempfile.mkdtemp(prefix="webphone-smoke-")
    binary = args.bin
    if not binary:
        binary = f"{workdir}/webphone-bin"
        print(f"building {binary} …", flush=True)
        env = dict(os.environ)
        env["GOEXPERIMENT"] = "jsonv2"
        build = subprocess.run(
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
        (the NixOS module ships these defaults when nginx.enable)."""
        port2 = free_port()
        cfg = {
            "addr": f"127.0.0.1:{port2}",
            "data_dir": f"{workdir}/data-fronted",
            "gateway": {"webhook_secret": "test-secret"},
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
        srv2 = subprocess.Popen(
            [binary], env=env2, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        base2 = f"http://127.0.0.1:{port2}"
        deadline = time.monotonic() + TIMEOUT
        while time.monotonic() < deadline:
            try:
                urllib.request.urlopen(base2 + "/healthz", timeout=1).read()
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
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "test-secret",
        }
    )
    server = subprocess.Popen(
        [binary], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
    )
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
        rc = run_checks(Smoke(base), boot_configured)
        rc = max(rc, restart_scenario(binary, workdir, port, env))
        return rc
    finally:
        server.terminate()
        try:
            server.wait(timeout=5)
        except subprocess.TimeoutExpired:
            server.kill()
        subprocess.run(["trash", workdir], capture_output=True, check=False)


if __name__ == "__main__":
    sys.exit(main())
