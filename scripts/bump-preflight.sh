#!/usr/bin/env bash
# bump-preflight: the upstream-diff pre-read for dependency bumps
# (SUPERB error-agents M10-M11). The vendorHash incident class burned
# main twice; this reads a module's old->new delta BEFORE the bump:
# overlapping-file diffs (capped), removed/added files, best-effort
# removed exported symbols, and the usage-site map of the changed
# packages against THIS repo's imports.
#
# Fail-open offline (the repo's unreachable-peer doctrine): a proxy
# outage is a WARN + exit 0 — a pre-read must never block a train; a
# wrong verdict (refusing on transient network) would be worse than
# none. Wrong config (unparsable module path) still refuses.
#
# usage:
#   bump-preflight.sh <module-path> <old-version> <new-version> [--json]
#   bump-preflight.sh --self-test
set -euo pipefail

say() { echo "bump-preflight: $*"; }

run_python() {
	python3 - "$@" <<'PY'
import base64
import difflib
import io
import json
import re
import subprocess
import sys
import urllib.request
import zipfile

REPO = subprocess.run(
    ["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True
).stdout.strip()

SYMBOL_RE = re.compile(r"^(?:func|type|var|const)\s+(\w+)", re.MULTILINE)
EXPORTED = re.compile(r"^[A-Z]")


def escape_module(path):
    # proxy.golang.org escaping: every uppercase letter becomes
    # !lowercase, per the module proxy protocol.
    return "".join(
        ("!" + ch.lower()) if ch.isupper() else ch for ch in path
    )


def fetch(url, timeout=30):
    req = urllib.request.Request(url, headers={"User-Agent": "webphone-bump-preflight"})
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return resp.read()


def zip_files(data):
    z = zipfile.ZipFile(io.BytesIO(data))
    # module zips carry one top-level dir: <module-path>@<version>/ —
    # anchor on the "@" (module paths contain slashes, never "@").
    first = z.namelist()[0]
    prefix = first[: first.index("/", first.index("@")) + 1]
    return {
        n[len(prefix):]: z.read(n)
        for n in z.namelist()
        if not n.endswith("/") and n.startswith(prefix)
    }


def exported_syms(content):
    return {
        m.group(1)
        for m in SYMBOL_RE.finditer(content.decode("utf-8", errors="replace"))
        if EXPORTED.match(m.group(1))
    }


def preflight(module, old_v, new_v):
    esc = escape_module(module)
    base = f"https://proxy.golang.org/{esc}/@v/"
    old_zip = zip_files(fetch(base + old_v + ".zip"))
    new_zip = zip_files(fetch(base + new_v + ".zip"))
    old, new = old_zip, new_zip

    overlap = sorted(set(old) & set(new))
    removed = sorted(set(old) - set(new))
    added = sorted(set(new) - set(old))

    changed_files, diff_blocks = [], []
    removed_syms = {}
    for path in overlap:
        if not path.endswith(".go"):
            continue
        o, n = old[path], new[path]
        if o == n:
            continue
        changed_files.append(path)
        osym, nsym = exported_syms(o), exported_syms(n)
        gone = sorted(osym - nsym)
        if gone:
            removed_syms[path] = gone
        if len(diff_blocks) < 20:
            d = list(
                difflib.unified_diff(
                    o.decode("utf-8", errors="replace").splitlines(),
                    n.decode("utf-8", errors="replace").splitlines(),
                    fromfile=old_v + path,
                    tofile=new_v + path,
                    lineterm="",
                )
            )
            diff_blocks.append("\n".join(d[:80]))

    usage = {}
    pkg_dirs = sorted({p.rsplit("/", 1)[0] for p in changed_files if "/" in p})
    for d in pkg_dirs:
        imp = module + "/" + d
        out = subprocess.run(
            ["rg", "-l", "--fixed-strings", imp, "internal/", "cmd/", "go.mod"],
            capture_output=True, text=True, cwd=REPO,
        )
        sites = [x for x in out.stdout.splitlines() if x]
        if sites:
            usage[imp] = sites[:10]

    verdict = "CHANGES-DETECTED" if changed_files else "NO-GO-FILE-CHANGES"
    return {
        "module": module,
        "old": old_v,
        "new": new_v,
        "verdict": verdict,
        "changed_go_files": changed_files,
        "removed_files": removed[:20],
        "added_files": added[:20],
        "removed_exported_symbols": removed_syms,
        "usage_sites": usage,
        "diffs": diff_blocks,
    }


def render_human(r):
    out = [
        f"== bump-preflight: {r['module']} {r['old']} -> {r['new']} ==",
        f"verdict: {r['verdict']}",
        f"changed .go files: {len(r['changed_go_files'])}",
    ]
    if r["removed_exported_symbols"]:
        out.append("REMOVED exported symbols (breaking-change candidates):")
        for p, syms in r["removed_exported_symbols"].items():
            out.append(f"  {p}: {', '.join(syms[:10])}")
    else:
        out.append("removed exported symbols: (none detected — best-effort scan)")
    if r["usage_sites"]:
        out.append("usage sites of changed packages in this repo:")
        for imp, sites in r["usage_sites"].items():
            out.append(f"  {imp}: {len(sites)} file(s), e.g. {sites[0]}")
    else:
        out.append("usage sites: (this repo does not import the changed packages)")
    if r["diffs"]:
        out.append("")
        out.append("[first diffs, capped]")
        out.append(r["diffs"][0])
    return "\n".join(out)


def self_test():
    module = "github.com/LarsArtmann/go-health"
    try:
        r = preflight(module, "v0.5.0", "v0.5.1")
    except Exception as e:  # noqa: BLE001 — fail-open offline is the doctrine
        print(f"bump-preflight self-test: SKIP (proxy unreachable: {e})")
        return 0
    checks = [
        ("verdict present", bool(r["verdict"])),
        ("changed files listed", isinstance(r["changed_go_files"], list)),
        ("human report non-empty", len(render_human(r)) > 100),
        ("escape rule", escape_module("github.com/LarsArtmann/go-health")
         == "github.com/!lars!artmann/go-health"),
    ]
    failed = 0
    for name, ok in checks:
        print(f"  {'PASS' if ok else 'FAIL'}  {name}")
        failed += 0 if ok else 1
    print(f"bump-preflight self-test: {len(checks) - failed}/{len(checks)} passed")
    return 1 if failed else 0


args = [a for a in sys.argv[1:]]
if "--self-test" in args:
    raise SystemExit(self_test())
as_json = "--json" in args
positional = [a for a in args if not a.startswith("--")]
if len(positional) != 3:
    print("usage: bump-preflight.sh <module-path> <old-version> <new-version> [--json]", file=sys.stderr)
    raise SystemExit(1)
module, old_v, new_v = positional
try:
    report = preflight(module, old_v, new_v)
except Exception as e:  # noqa: BLE001 — fail-open offline is the doctrine
    print(f"bump-preflight: WARN proxy unreachable ({e}) — no verdict, train proceeds (fail-open)")
    raise SystemExit(0)
print(json.dumps(report, indent=2) if as_json else render_human(report))
PY
}

if [ "${1:-}" = "--self-test" ]; then
	run_python --self-test
	exit $?
fi
if [ $# -lt 3 ]; then
	say "usage: bump-preflight.sh <module-path> <old-version> <new-version> [--json]"
	exit 1
fi
run_python "$@"
