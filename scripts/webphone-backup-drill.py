#!/usr/bin/env python3
"""Backup/restore drill for the webphone data directory (SUPERB P17).

Proves the documented pattern end to end on scratch:
  1. boot a binary with a temp data dir, POST one inbound webhook message
     carrying a binary attachment (thread + message + blob in SQLite +
     the blob store)
  2. stop the service, tar the whole data dir (the backup step)
  3. restore the archive into a fresh location (the disaster), reboot
  4. sign in against the restored store and pull the attachment back out
     byte-identical — proof the backup is restorable, not just copyable
"""

import base64
import http.cookiejar
import json
import os
import shutil
import socket
import sqlite3
import subprocess
import tarfile
import tempfile
import time
import urllib.request

BIN = os.environ["WP_BIN"]
SECRET = "drill-secret"
PORT = 18099
PAYLOAD = b"webphone backup drill payload \xf0\x9f\x93\x9e " + bytes(range(256))


def wait_port(port, timeout=30):
    deadline = time.time() + timeout
    while time.time() < deadline:
        with socket.socket() as s:
            if s.connect_ex(("127.0.0.1", port)) == 0:
                return
        time.sleep(0.2)
    raise SystemExit(f"port {port} never opened")


class Client:
    def __init__(self):
        self.jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.jar)
        )
        self.csrf = ""

    def request(self, method, path, data=None, content_type=None, headers=None):
        req = urllib.request.Request(
            f"http://127.0.0.1:{PORT}{path}", data=data, method=method
        )
        if content_type:
            req.add_header("Content-Type", content_type)
        for key, value in (headers or {}).items():
            req.add_header(key, value)
        if self.csrf and method not in ("GET", "HEAD"):
            req.add_header("X-CSRF-Token", self.csrf)
        try:
            resp = self.opener.open(req, timeout=10)
            return resp.status, resp.read()
        except urllib.error.HTTPError as err:
            return err.code, err.read()

    def login(self):
        status, body = self.request("GET", "/api/csrf")
        assert status == 200, "csrf fetch failed"
        self.csrf = json.loads(body)["token"]
        status, _ = self.request(
            "POST",
            "/api/session",
            json.dumps({"extension": "1000", "password": "pw"}).encode(),
            "application/json",
        )
        assert status == 201, f"loopback login expected 201, got {status}"
        status, body = self.request("GET", "/api/csrf")
        assert status == 200, "csrf adoption failed"
        self.csrf = json.loads(body)["token"]


def boot(data_dir):
    # Kill only THIS drill's binary (never a real webphone service that
    # may share the host): match the exact temp binary path. pkill is
    # probed (absent in the nix sandbox, where no prior drill process
    # can exist) so the drill also runs as a flake check.
    if shutil.which("pkill"):
        subprocess.run(["pkill", "-f", BIN], capture_output=True)
        time.sleep(0.5)
    env = dict(
        os.environ,
        WEBPHONE_ADDR=f"127.0.0.1:{PORT}",
        WEBPHONE_DATA_DIR=data_dir,
        WEBPHONE_GATEWAY__WEBHOOK_SECRET=SECRET,
    )
    proc = subprocess.Popen(
        [BIN], env=env, stdout=open(data_dir + ".log", "wb"), stderr=subprocess.STDOUT
    )
    wait_port(PORT)
    return proc


def attachment_row(data_dir):
    db = sqlite3.connect(os.path.join(data_dir, "webphone.db"))
    row = db.execute("select id, path from attachments limit 1").fetchone()
    db.close()
    return row


def main():
    src = tempfile.mkdtemp(prefix="wp-src-")
    archive = tempfile.mktemp(prefix="wp-backup-", suffix=".tar.gz")
    dst = tempfile.mktemp(prefix="wp-restore-")

    # 1. load the source store through the real webhook surface
    proc = boot(src)
    try:
        hook = Client()
        status, _ = hook.request(
            "POST",
            "/hooks/message",
            json.dumps(
                {
                    "owner": "1000",
                    "from": "+15550001111",
                    "body": "drill inbound with attachment",
                    "attachments": [
                        {
                            "name": "drill.bin",
                            "mime_type": "application/octet-stream",
                            "data_base64": base64.b64encode(PAYLOAD).decode(),
                        }
                    ],
                }
            ).encode(),
            "application/json",
            {"Authorization": f"Bearer {SECRET}"},
        )
        assert status == 202, f"hook expected 202, got {status}"
        time.sleep(0.5)
        att = attachment_row(src)
        assert att, "no attachment row landed in the source store"
        att_id, att_path = att
        assert os.path.exists(os.path.join(src, "files", att_path)), (
            "blob missing on disk"
        )
        print(f"[1] source loaded: attachment {att_id} at {att_path}")
    finally:
        proc.terminate()
        proc.wait(timeout=10)

    # 2. backup: archive the whole data dir while the service is stopped
    with tarfile.open(archive, "w:gz") as tar:
        tar.add(src, arcname=".")
    print(f"[2] backup written: {archive} ({os.path.getsize(archive)} bytes)")

    # 3. disaster: restore the archive into a fresh location, reboot
    os.makedirs(dst)
    with tarfile.open(archive) as tar:
        tar.extractall(dst, filter="data")
    proc = boot(dst)
    try:
        client = Client()
        client.login()
        status, body = client.request("GET", f"/attachments/{att_id}")
        assert status == 200, f"attachment download failed: {status} {body[:200]}"
        assert body == PAYLOAD, "restored attachment bytes differ"
        print("[3] restored store serves the attachment byte-identical")
        status, _ = client.request("GET", "/partials/messages")
        assert status == 200, "restored messages panel failed to render"
        print("[4] restore drill PASSED: sqlite rows + blob content survive")
    finally:
        proc.terminate()
        proc.wait(timeout=10)
        shutil.rmtree(src, ignore_errors=True)
        os.remove(archive)
        shutil.rmtree(dst, ignore_errors=True)


if __name__ == "__main__":
    main()
