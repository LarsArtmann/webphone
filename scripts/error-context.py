#!/usr/bin/env python3
r"""error-context: the error-code router (SUPERB error-agents plan M1-M2).

One journal line or one `<seam>.<op>` code in -> the full context bundle
out: the error-contract registry row, the seam files that mint the code,
the family-test pin, lessons/flake-ledger mentions, and the owning
package's recent commits. Every later piece of the error-agents tooling
(watchers drafting ledger rows, incident bundles, future agent sessions)
consumes this bundle instead of starting from a blank page.

usage:
  error-context.py [--json] <code-or-journal-line>
  error-context.py --self-test

input normalization: a bare code (`asr.http`), a bracketed journal
fragment (`... [asr.http] upstream 500`), or a free-form line carrying a
registry code as a substring (longest match wins).

exits: 0 bundle delivered (verdict OK or DRIFT), 1 usage/environment
error, 2 unknown code (verdict carries nearest-code suggestions).
"""

import argparse
import datetime
import difflib
import json
import re
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
CONTRACT = REPO / "docs" / "error-contract.md"
REGISTRY_BEGIN = "<!-- error-code-registry: BEGIN"
REGISTRY_END = "<!-- error-code-registry: END"
CODE_RE = re.compile(r"^[a-z][a-z0-9_]*(?:\.[a-z0-9_]+)+$")
BRACKET_RE = re.compile(r"\[([a-z][a-z0-9_]*(?:\.[a-z0-9_]+)+)\]")
TESTFUNC_RE = re.compile(r"^func (Test\w+)")


def run(argv, **kw):
    return subprocess.run(argv, capture_output=True, text=True, **kw)


def normalize_target(raw):
    """Journal line -> code. Bracket first, then a whole-arg code, then
    None (the caller falls back to a substring scan over the registry)."""
    m = BRACKET_RE.search(raw)
    if m:
        return m.group(1)
    if CODE_RE.match(raw.strip()):
        return raw.strip()
    return None


def parse_registry(text=None):
    """The generated block's table: [(code, families, first_site)]."""
    if text is None:
        text = CONTRACT.read_text(encoding="utf-8")
    rows = []
    inside = False
    for line in text.splitlines():
        if REGISTRY_BEGIN in line:
            inside = True
            continue
        if REGISTRY_END in line:
            break
        if not inside or not line.startswith("|"):
            continue
        cells = [c.strip() for c in line.strip("|").split("|")]
        if len(cells) >= 3 and CODE_RE.match(cells[0]):
            rows.append((cells[0], cells[1], cells[2]))
    return rows


def resolve_code(raw, registry):
    code = normalize_target(raw)
    if code:
        return code
    hits = [c for c, _, _ in registry if c in raw]
    return max(hits, key=len) if hits else None


def seam_files(code, first_site):
    """Files minting/mentioning the code, constructor site ranked first."""
    out = run(["rg", "-l", "--fixed-strings", code, "internal/", "cmd/"], cwd=REPO)
    paths = [p for p in out.stdout.splitlines() if p]
    if not paths and first_site:
        return [], "DRIFT"
    counts = {}
    for p in paths:
        n = run(["rg", "-c", "--fixed-strings", code, p], cwd=REPO)
        counts[p] = int(n.stdout.strip()) if n.stdout.strip() else 1
    ranked = sorted(counts, key=lambda p: (p != first_site, -counts[p], p))
    files = [{"path": p, "matches": counts[p], "first_site": p == first_site} for p in ranked]
    return files, ("OK" if paths else "DRIFT")


def enclosing_test(path, line_no):
    try:
        lines = (REPO / path).read_text(encoding="utf-8").splitlines()
    except OSError:
        return None
    name = None
    for i in range(min(line_no, len(lines))):
        m = TESTFUNC_RE.match(lines[i])
        if m:
            name = m.group(1)
    return name


def family_pins(code):
    out = run(["rg", "-n", "--fixed-strings", code, "--glob", "*family_test.go", "internal/"], cwd=REPO)
    pins = []
    for line in out.stdout.splitlines():
        parts = line.split(":", 2)
        if len(parts) < 3:
            continue
        path, line_no, text = parts
        pins.append({"file": path, "line": line_no, "text": text.strip(), "test": enclosing_test(path, int(line_no))})
    return pins


def doc_mentions(code):
    mentions = {"lessons": [], "ledger": []}
    for key, path in (("lessons", REPO / "docs" / "lessons.md"), ("ledger", REPO / "docs" / "ci-flake-ledger.md")):
        if not path.exists():
            continue
        for i, text in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if code in text:
                mentions[key].append({"file": f"docs/{path.name}", "line": i, "text": text.strip()[:200]})
    return mentions


def recent_commits(first_site):
    if not first_site:
        return []
    out = run(["git", "log", "--oneline", "-5", "--", str(Path(first_site).parent)], cwd=REPO)
    commits = []
    for line in out.stdout.splitlines():
        sha, _, subject = line.partition(" ")
        commits.append({"sha": sha, "subject": subject})
    return commits


def nearest_codes(raw, registry):
    codes = [c for c, _, _ in registry]
    prefixes = [c for c in codes if c.startswith(raw) or raw.startswith(c.split(".")[0] + ".")]
    close = difflib.get_close_matches(raw, codes, n=5, cutoff=0.5)
    seen = []
    for c in prefixes + close:
        if c not in seen:
            seen.append(c)
    return seen[:5]


def build_bundle(raw, registry):
    code = resolve_code(raw, registry)
    registered = row = next(((c, f, s) for c, f, s in registry if c == code), None) if code else None
    if not code or not registered:
        probe = code or (normalize_target(raw) or raw.strip())
        return {
            "target": raw,
            "verdict": "UNKNOWN",
            "suggestions": nearest_codes(probe, registry),
        }
    first_site = row[2]
    files, verdict = seam_files(code, first_site)
    bundle = {
        "target": raw,
        "code": code,
        "verdict": verdict,
        "generated_at": datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds"),
        "registry": {"families": row[1], "first_site": first_site} if row else None,
        "seam_files": files,
        "family_pins": family_pins(code),
        "mentions": doc_mentions(code),
        "recent_commits": recent_commits(first_site),
    }
    if verdict == "DRIFT":
        bundle["drift_hint"] = (
            "code is in the error-contract registry but absent from source — "
            "run: go test ./internal/arch -run TestErrorCodeRegistryIsFresh"
        )
    return bundle


def render_human(b):
    out = [f"== error-context: {b.get('code', b['target'])} ==", f"verdict: {b['verdict']}"]
    if b["verdict"] == "UNKNOWN":
        out.append("nearest registry codes: " + (", ".join(b["suggestions"]) or "(none)"))
        out.append("the code never minted — check spelling or the registry: docs/error-contract.md")
        return "\n".join(out)
    r = b.get("registry")
    out.append(f"registry: {b['code']} | families: {r['families']} | first site: {r['first_site']}")
    out.append("")
    out.append("[seam files]")
    if b["seam_files"]:
        for f in b["seam_files"]:
            tag = "  (first site)" if f["first_site"] else ""
            out.append(f"  {f['path']} ({f['matches']} match(es)){tag}")
    else:
        out.append("  (none — see drift hint)")
    out.append("")
    out.append("[family test pins]")
    if b["family_pins"]:
        for p in b["family_pins"]:
            at = f" in {p['test']}" if p["test"] else ""
            out.append(f"  {p['file']}:{p['line']}{at}: {p['text'][:120]}")
    else:
        out.append("  (no family_test.go pin found)")
    out.append("")
    out.append("[lessons / flake-ledger mentions]")
    for key in ("lessons", "ledger"):
        rows = b["mentions"][key]
        if rows:
            for m in rows:
                out.append(f"  {m['file']}:{m['line']}: {m['text'][:140]}")
        else:
            out.append(f"  {key}: (no mentions — clean slate)")
    out.append("")
    out.append(f"[recent commits — {Path(r['first_site']).parent if r and r['first_site'] else 'n/a'}]")
    if b["recent_commits"]:
        for c in b["recent_commits"]:
            out.append(f"  {c['sha']} {c['subject'][:110]}")
    else:
        out.append("  (no git history available)")
    if b["verdict"] == "DRIFT":
        out.append("")
        out.append(f"DRIFT: {b['drift_hint']}")
    return "\n".join(out)


def self_test():
    registry = parse_registry()
    checks = []

    def check(name, ok):
        checks.append((name, bool(ok)))

    check("registry parsed (>100 rows)", len(registry) > 100)
    picks = {
        "asr.empty": "Rejection",
        "blob.open": "Infrastructure",
        "crm.http": "runtime-split",
    }
    for code, family in picks.items():
        b = build_bundle(code, registry)
        check(f"{code}: verdict OK", b["verdict"] == "OK")
        check(f"{code}: registry row present", bool(b["registry"]))
        check(f"{code}: families mention {family}", family in (b["registry"] or {}).get("families", ""))
        check(f"{code}: seam files non-empty", len(b["seam_files"]) > 0)
        check(f"{code}: first site ranked first", b["seam_files"][0]["first_site"])
        check(f"{code}: family pin found", len(b["family_pins"]) > 0)
        human = render_human(b)
        for section in ("[seam files]", "[family test pins]", "[lessons / flake-ledger mentions]", "[recent commits"):
            check(f"{code}: section {section!r} present", section in human)

    b = build_bundle("[asr.http] upstream answered 500 while transcribing a segment", registry)
    check("journal-line bracket normalized", b["code"] == "asr.http")

    b = build_bundle("zzz.no.such.code", registry)
    check("unknown code verdict", b["verdict"] == "UNKNOWN")
    check("unknown code suggestions non-empty", len(b["suggestions"]) > 0)

    b = build_bundle("asr.htt", registry)
    check("typo suggests nearest code", "asr.http" in b["suggestions"])

    drifted = registry + [("zzz.fake.drift", "Transient", "internal/nope/nope.go")]
    b = build_bundle("zzz.fake.drift", drifted)
    check("contract-only code yields DRIFT", b["verdict"] == "DRIFT")
    check("DRIFT names the registry test", "TestErrorCodeRegistryIsFresh" in b.get("drift_hint", ""))

    failed = [n for n, ok in checks if not ok]
    for n, ok in checks:
        print(f"  {'PASS' if ok else 'FAIL'}  {n}")
    print(f"error-context self-test: {len(checks) - len(failed)}/{len(checks)} passed")
    return 1 if failed else 0


def main(argv=None):
    ap = argparse.ArgumentParser(description="route an error code or journal line to its context bundle")
    ap.add_argument("target", nargs="?", help="error code or journal line")
    ap.add_argument("--json", action="store_true", help="machine-readable bundle")
    ap.add_argument("--self-test", action="store_true", help="run the pinned self-test scenarios")
    args = ap.parse_args(argv)

    if args.self_test:
        return self_test()
    if not args.target:
        ap.error("a code or journal line is required (or --self-test)")

    registry = parse_registry()
    if not registry:
        print("error-context: cannot parse the registry block in docs/error-contract.md", file=sys.stderr)
        return 1
    bundle = build_bundle(args.target, registry)
    print(json.dumps(bundle, indent=2) if args.json else render_human(bundle))
    return 2 if bundle["verdict"] == "UNKNOWN" else 0


if __name__ == "__main__":
    sys.exit(main())
