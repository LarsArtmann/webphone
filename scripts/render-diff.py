#!/usr/bin/env python3
"""Render-diff: prove two webphone binaries render the same partial bytes.

The instrument from the 2026-09-24 dedup sweep (docs/status/
2026-09-24_16-56_art-dupl-t2-panelerror-sweep.md §b), committed so it
stops living in trashed /tmp files. Recipe rules (each one is a trap
the sweep already paid for):

- PARTIALS ONLY: the full shell carries a per-session CSRF token and a
  per-boot nonce and would false-diff; the tab partials are the render
  surface that matters.
- Login works because loopback dev mode skips PBX verification (WARNs).
- A tokenless login POST is a pinned 403 — always adopt the CSRF token
  via GET /api/csrf first (the Smoke helper dance).
- --error-instrument boots in webhook-gateway mode with a DEAD URL:
  POST /messages/send answers 502 and re-renders the panel with the
  error banner + a persisted failed row, so the error surface and the
  thread partial are exercised too. (A dead phone_api_url does NOT
  work: it blocks loopback login verification before any panel.)
- Per-server random thread IDs are normalized on both sides before the
  byte comparison (the only sanctioned non-determinism).

Usage:
  scripts/render-diff.py OLD_BIN NEW_BIN [--error-instrument]

Exit 0 = every partial byte-identical after normalization; anything
else prints the offending diff excerpt and exits 1.
"""

from __future__ import annotations

import argparse
import http.client
import os
import re
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
import uuid

PARTIALS = [
    "/partials/messages",
    "/partials/fax",
    "/partials/voicemail",
    "/partials/history",
    "/partials/contacts",
    "/partials/settings",
    "/partials/nav",
]

THREAD_ID = re.compile(
    rb"\b(?:t-|Thread:)[A-Za-z0-9_-]{16,}\b"
)  # branded ids: t-… / Thread:…
# Wall-clock stamps in failed-row/thread rendering (same length, different
# digits across boots — the second sanctioned non-determinism; the fixed
# probe content contains no clock-like text).
CLOCK = re.compile(rb"\b\d{1,2}:\d{2}(?::\d{2})?\b")
BOUNDARY = "----renderdiff" + uuid.uuid4().hex


def free_port() -> int:
    import socket

    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def boot(binary: str, port: int, data_dir: str, webhook_dead: bool) -> subprocess.Popen:
    env = dict(os.environ)
    env.update(
        {
            "WEBPHONE_ADDR": f"127.0.0.1:{port}",
            "WEBPHONE_DATA_DIR": data_dir,
            "WEBPHONE_GATEWAY__WEBHOOK_SECRET": "devsecret",
        }
    )
    if webhook_dead:
        env["WEBPHONE_GATEWAY__MODE"] = "webhook"
        env["WEBPHONE_GATEWAY__WEBHOOK_URL"] = "http://127.0.0.1:9"
    return subprocess.Popen(
        [binary],
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )


def wait_up(port: int, deadline_s: float = 20.0) -> bool:
    end = time.time() + deadline_s
    while time.time() < end:
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{port}/healthz", timeout=1).read()
            return True
        except (OSError, http.client.HTTPException):
            time.sleep(0.25)
    return False


def make_opener() -> urllib.request.OpenerDirector:
    import http.cookiejar

    return urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
    )


def request(
    opener,
    port: int,
    method: str,
    path: str,
    body: bytes | None = None,
    headers: dict | None = None,
) -> tuple[int, bytes]:
    req = urllib.request.Request(
        f"http://127.0.0.1:{port}{path}", data=body, method=method
    )
    for key, value in (headers or {}).items():
        req.add_header(key, value)
    try:
        with opener.open(req, timeout=10) as resp:
            return resp.status, resp.read()
    except urllib.error.HTTPError as err:
        return err.code, err.read()


def login_session(opener, port: int) -> dict:
    """Login (loopback mode accepts credentials without PBX verify) + CSRF
    adoption; returns the header dict later requests need."""
    import json

    # Order matters (the smoke script's pinned-403 lesson): adopt a CSRF
    # token FIRST (the session POST is token-gated), then log in, then
    # RE-adopt — login rotates the token.
    status, token_body = request(opener, port, "GET", "/api/csrf")
    if status != 200:
        raise SystemExit(f"initial csrf adopt failed: {status}")
    token = json.loads(token_body).get("token", "")
    if not token:
        raise SystemExit("initial csrf adopt: empty token")

    form = json.dumps({"extension": "1001", "password": "pw"}).encode()
    status, body = request(
        opener,
        port,
        "POST",
        "/api/session",
        form,
        {"Content-Type": "application/json", "X-CSRF-Token": token},
    )
    if status != 201:
        raise SystemExit(f"login failed: {status} {body[:200]!r}")

    status, token_body = request(opener, port, "GET", "/api/csrf")
    if status != 200:
        raise SystemExit(f"post-login csrf adopt failed: {status}")
    token = json.loads(token_body).get("token", "")
    if not token:
        raise SystemExit("post-login csrf adopt: empty token")
    return {"X-CSRF-Token": token}


def seed_error_row(opener, port: int, headers: dict) -> None:
    """POST a send against the dead webhook gateway -> 502 + failed row."""
    parts = []
    for name, value in (("to", "+491700000000"), ("body", "render-diff probe")):
        parts.append(
            f'--{BOUNDARY}\r\nContent-Disposition: form-data; name="{name}"\r\n\r\n{value}\r\n'.encode()
        )
    body = b"".join(parts) + f"--{BOUNDARY}--\r\n".encode()
    status, _ = request(
        opener,
        port,
        "POST",
        "/messages/send",
        body,
        {**headers, "Content-Type": f"multipart/form-data; boundary={BOUNDARY}"},
    )
    if status != 502:
        raise SystemExit(f"error instrument expected 502, got {status}")


def collect(binary: str, error_instrument: bool) -> dict[str, bytes]:
    port = free_port()
    with tempfile.TemporaryDirectory(prefix="renderdiff-") as data_dir:
        proc = boot(binary, port, data_dir, error_instrument)
        try:
            if not wait_up(port):
                raise SystemExit(f"{binary}: server did not come up")
            opener = make_opener()
            headers = login_session(opener, port)
            if error_instrument:
                seed_error_row(opener, port, headers)
            out: dict[str, bytes] = {}
            for path in PARTIALS:
                status, body = request(opener, port, "GET", path, None, headers)
                if status != 200:
                    raise SystemExit(f"{path}: expected 200, got {status}")
                normalized = THREAD_ID.sub(b"t-<NORMALIZED>", body)
                out[path] = CLOCK.sub(b"<TIME>", normalized)
            return out
        finally:
            proc.terminate()
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("old_binary")
    parser.add_argument("new_binary")
    parser.add_argument(
        "--error-instrument",
        action="store_true",
        help="also exercise the 502 send-failure re-render (dead webhook URL)",
    )
    args = parser.parse_args()

    old = collect(args.old_binary, args.error_instrument)
    new = collect(args.new_binary, args.error_instrument)

    failures = 0
    for path in PARTIALS:
        same = old[path] == new[path]
        print(
            f"  {'OK ' if same else 'DIFF'}  {path} ({len(old[path])}B vs {len(new[path])}B)"
        )
        if not same:
            failures += 1
            for i, (a, b) in enumerate(
                zip(old[path].splitlines(), new[path].splitlines())
            ):
                if a != b:
                    print(f"       first differing line {i + 1}:")
                    print(f"       old: {a[:160]!r}")
                    print(f"       new: {b[:160]!r}")
                    break
    print(
        f"render-diff: {len(PARTIALS) - failures}/{len(PARTIALS)} byte-identical"
        + (" (error banner + failed row exercised)" if args.error_instrument else "")
    )
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
