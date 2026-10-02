#!/usr/bin/env python3
"""Asset-perf timing baseline (T10.6): cold GET vs gzip GET vs ETag
revalidation, measured with urllib + time.perf_counter (curl is banned
in the assistant harness). Run against a locally booted binary; prints
a markdown table for the perf-plan appendix."""

import gzip
import io
import sys
import time
import urllib.error
import urllib.request

BASE = sys.argv[1] if len(sys.argv) > 1 else "http://127.0.0.1:18123"
ASSET = "/assets/vendor/sip.min.js"  # the payload that dominates first load
REPEATS = 5


def timed_get(path, headers=None):
    req = urllib.request.Request(BASE + path, headers=headers or {})
    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            body = resp.read()
            status = resp.status
            hdrs = resp.headers
    except urllib.error.HTTPError as err:
        # 304 surfaces as an error; it IS the response we want to measure.
        body = err.read()
        status = err.code
        hdrs = err.headers
    elapsed = (time.perf_counter() - start) * 1000
    return status, hdrs, body, elapsed


def median(values):
    ordered = sorted(values)
    return ordered[len(ordered) // 2]


def fmt(values):
    return f"{median(values):8.1f} ms (median of {len(values)})"


def main():
    rows = []

    status, hdrs, body, _ = timed_get(ASSET)
    assert status == 200, status
    etag = hdrs["ETag"]
    cold_bytes = len(body)

    plain_times, gzip_times = [], []
    gzip_body_len = None
    for _ in range(REPEATS):
        _, _, _, ms = timed_get(ASSET)
        plain_times.append(ms)
        _, hdrs2, body2, ms2 = timed_get(ASSET, {"Accept-Encoding": "gzip"})
        if gzip_body_len is None:
            gzip_body_len = len(body2)
            gunzip = gzip.GzipFile(fileobj=io.BytesIO(body2)).read()
            assert gunzip == body, "gzip roundtrip must reproduce the bytes"
        gzip_times.append(ms2)

    reval_times, reval_bytes = [], None
    for _ in range(REPEATS):
        status3, _, body3, ms3 = timed_get(ASSET, {"If-None-Match": etag})
        if reval_bytes is None:
            assert status3 == 304, f"expected 304, got {status3}"
            reval_bytes = len(body3)
        reval_times.append(ms3)

    _, _, head_body, _ = timed_get("/")
    modulepreload_count = head_body.count(b'rel="modulepreload"')

    rows.append(("cold GET (no gzip)", cold_bytes, fmt(plain_times)))
    rows.append(("GET with Accept-Encoding: gzip", gzip_body_len, fmt(gzip_times)))
    rows.append(("revalidate (If-None-Match → 304)", reval_bytes, fmt(reval_times)))

    print(f"| request for `{ASSET}` | bytes on wire | wall time |")
    print("| --- | --- | --- |")
    for label, nbytes, timing in rows:
        print(f"| {label} | {nbytes} | {timing} |")
    print()
    print(
        f"- ETag: `{etag[:23]}…` (strong, sha256 of content); "
        f"Cache-Control: `{hdrs['Cache-Control']}`; "
        f"Vary: `{hdrs['Vary']}`"
    )
    saving = 100 - (gzip_body_len * 100 // cold_bytes)
    print(
        f"- gzip saves {saving}% of the dominant payload; the 304 revalidation "
        f"transfers {reval_bytes} bytes ({100 - (reval_bytes * 100 // cold_bytes)}% less than cold)."
    )
    print(
        f"- served page head carries {modulepreload_count} modulepreload links "
        f"(the island ESM graph, {modulepreload_count} modules)."
    )


if __name__ == "__main__":
    main()
