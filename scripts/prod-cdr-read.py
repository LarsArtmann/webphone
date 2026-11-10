#!/usr/bin/env python3
"""One-shot prod CDR read: mint a session, pull /phone-api/history, print
the C3 inbound-window + D2 row-level verdicts. Read-only; nothing stored
outside stdout."""
import json
import re
import sys
import urllib.request
import urllib.error
import ssl
from pathlib import Path

BASE = "https://pbx.artmann.tech"
PW_FILE = Path.home() / ".pbx-prod-secrets" / "telephony_ext_1000"
EXT = "1000"
LIMIT = 500

ctx = ssl.create_default_context()


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *a, **k):
        return None


import http.cookiejar
jar = http.cookiejar.CookieJar()
opener = urllib.request.build_opener(NoRedirect, urllib.request.HTTPCookieProcessor(jar))


def get(path, cookie=None):
    req = urllib.request.Request(BASE + path)
    if cookie:
        req.add_header("Cookie", cookie)
    try:
        with opener.open(req, timeout=30) as r:
            return r.status, r.read().decode("utf-8", "replace"), r.headers
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", "replace"), e.headers


def post_json(path, payload, csrf, cookie=None):
    body = json.dumps(payload).encode()
    req = urllib.request.Request(BASE + path, data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-CSRF-Token", csrf)
    if cookie:
        req.add_header("Cookie", cookie)
    try:
        with opener.open(req, timeout=30) as r:
            return r.status, r.read().decode("utf-8", "replace"), r.headers
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", "replace"), e.headers


password = PW_FILE.read_text().strip()

# 1. anonymous page -> csrf meta
rc, page, _ = get("/")
m = re.search(r'name="csrf-token" content="([^"]+)"', page)
if rc != 200 or not m:
    print(f"FAIL: cannot read the anonymous page (rc={rc})")
    sys.exit(1)
csrf = m.group(1)

# 2. mint the session
rc, body, headers = post_json("/api/session", {"extension": EXT, "password": password}, csrf)
if rc not in (200, 201):
    print(f"FAIL: session mint rc={rc}: {body[:200]}")
    sys.exit(1)
setc = headers.get("Set-Cookie", "")
mm = re.search(r"webphone_session=([^;]+)", setc)
if not mm:
    print("FAIL: no session cookie in the mint response")
    sys.exit(1)
cookie = f"webphone_session={mm.group(1)}"

# 3. adopt the rotated csrf (harmless for GETs, kept for parity)
rc, cbody, _ = get("/api/csrf", cookie)
rot = ""
if rc == 200:
    try:
        rot = json.loads(cbody).get("token", "")
    except json.JSONDecodeError:
        rot = ""

# 4. the CDR page
rc, hbody, _ = get(f"/phone-api/history?limit={LIMIT}", cookie)
if rc != 200:
    print(f"FAIL: /phone-api/history rc={rc}: {hbody[:200]}")
    sys.exit(1)
data = json.loads(hbody)
entries = data.get("entries", [])
print(f"CDR rows fetched: {len(entries)}")

inbound = [e for e in entries if e.get("context") == "public"]
outbound = [e for e in entries if e.get("context") != "public"]
ans_in = [e for e in inbound if e.get("billsec", 0) > 0]
ans_out = [e for e in outbound if e.get("billsec", 0) > 0]
undated = [e for e in entries if not re.match(r"^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$", e.get("start", ""))]
print(f"inbound: {len(inbound)} (answered {len(ans_in)}) | outbound: {len(outbound)} (answered {len(ans_out)})")
print(f"rows failing the seconds-parse: {len(undated)}")

print("\n=== inbound rows (the C3 window) ===")
for e in inbound[:20]:
    print(f"  start={e.get('start','?')} from={e.get('caller_id_number','?')} name={e.get('caller_id_name','')!r} billsec={e.get('billsec',0)}")

print("\n=== answered rows detail (the D2 24-answered story) ===")
for e in ans_out:
    print(f"  start={e.get('start','?')} dest={e.get('destination_number','?')} name={e.get('caller_id_name','')!r} billsec={e.get('billsec',0)}")

print("\n=== rows with garbled/odd caller names (the pifo42h1 hunt) ===")
for e in entries:
    n = e.get("caller_id_name", "")
    if n and not re.match(r"^[\w\s.+\-()]*$", n) or (n and re.search(r"[^A-Za-z0-9 +.\-()]", n)):
        print(f"  start={e.get('start','?')} ctx={e.get('context','?')} from={e.get('caller_id_number','?')} name={n!r} dest={e.get('destination_number','?')} billsec={e.get('billsec',0)}")

# the literal pifo42h1 hunt
hits = [e for e in entries if "pifo" in json.dumps(e).lower()]
print(f"\nliteral pifo42h1 rows: {len(hits)}")
for e in hits:
    print(f"  {json.dumps(e)}")
