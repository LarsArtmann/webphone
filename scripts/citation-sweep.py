#!/usr/bin/env python3
"""citation-sweep — verify the fleet-libs reports' file:line citations.

A report that pins a claim to `internal/server/server.go:841` is only as
good as that line still existing. The tree moves daily; the reports do not.
This sweep re-resolves every cite in docs/review/libs/*.md against the tree
(the throwaway from the 2026-11-24 pareto plan, made a repo citizen as T10).

Cite shapes recognized on every line of a report:

  path.ext:NNN        repo-relative (internal/ cmd/ scripts/ docs/ package/ nix/)
  /abs/path.ext:NNN   absolute
  <lib>/path.ext:NNN  library-side, resolved against --libs-root (the fleet);
                      a slash-path whose first segment is not a fleet repo is
                      matched as a suffix against every fleet repo, so
                      metaengine/v4/encoded.go finds go-cqrs-lite
  name.ext:NNN        unique basename in the repo tree

Ranges `:12-34` and en-dash `:12–34` check the END line only: a range is in
range when its last line exists.

Classes:

  OK            resolved, the cited line exists
  OUT-OF-RANGE  resolved, but the file has fewer lines than the cite ends at
  MISSING       slash-path cite: no such file in the repo or the fleet
  AMBIGUOUS     basename (or fleet suffix) matches several files
  UNKNOWN       basename matches no file in the repo tree

Usage:

  scripts/citation-sweep.py [file ...]        # default docs/review/libs/*.md
  scripts/citation-sweep.py --self-test       # fixture-proven, can fail

Exit 0 = every cite OK; 1 = any bad cite (one line per bad cite, then a
summary); 2 = usage error. scripts/audit-verify.sh runs this as an ADVISORY
tail after its own summary and never gates its exit code on it.
"""

from __future__ import annotations

import argparse
import collections
import os
import re
import sys

EXTS = "go|templ|md|js|mjs|css|nix|sh|json|py|txt|yml|yaml|toml|input"
REPO_PREFIXES = ("internal/", "cmd/", "scripts/", "docs/", "package/", "nix/")
DEFAULT_LIBS_ROOT = "/home/lars/projects"
DEFAULT_REPORT_GLOB = "docs/review/libs/*.md"

# A path cite: absolute, or at least one slash. Path characters stop at
# whitespace, a backtick, a colon or parentheses (markdown prose fences).
PATH_RE = re.compile(
    r"(/[^\s`:()]+|[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)+)"
    r"\.(?:%s):([0-9]+)(?:[-–]([0-9]+))?" % EXTS
)
# A bare-name cite: not preceded by a word char, dot, slash or hyphen, so the
# tail of a path cite can never double-count as a basename cite.
BARE_RE = re.compile(
    r"(?<![\w/.-])([A-Za-z0-9_][A-Za-z0-9_.-]*\.(?:%s)):([0-9]+)(?:[-–]([0-9]+))?"
    % EXTS
)

CLASSES = ("OK", "OUT-OF-RANGE", "MISSING", "AMBIGUOUS", "UNKNOWN")


class Sweeper:
    """Resolves cites against one repo root plus one library fleet root."""

    def __init__(self, root: str, libs_root: str) -> None:
        self.root = os.path.abspath(root)
        self.libs_root = os.path.abspath(libs_root)
        self._basenames: dict[str, list[str]] | None = None
        self._lib_dirs: list[str] | None = None
        self._line_counts: dict[str, int] = {}

    # -- indexes -----------------------------------------------------------

    def basenames(self) -> dict[str, list[str]]:
        if self._basenames is None:
            idx: dict[str, list[str]] = collections.defaultdict(list)
            for dirpath, dirnames, filenames in os.walk(self.root):
                dirnames[:] = [d for d in dirnames if d not in (".git", "node_modules")]
                for name in filenames:
                    idx[name].append(os.path.join(dirpath, name))
            self._basenames = dict(idx)
        return self._basenames

    def lib_dirs(self) -> list[str]:
        if self._lib_dirs is None:
            try:
                self._lib_dirs = sorted(
                    e.name for e in os.scandir(self.libs_root) if e.is_dir()
                )
            except OSError:
                self._lib_dirs = []
        return self._lib_dirs

    def line_count(self, path: str) -> int:
        if path not in self._line_counts:
            with open(path, encoding="utf-8", errors="replace") as f:
                self._line_counts[path] = sum(1 for _ in f)
        return self._line_counts[path]

    # -- resolution --------------------------------------------------------

    def resolve_slash_path(self, path: str) -> tuple[list[str], str]:
        """Return (candidate files, origin) for a cite carrying a slash."""
        if os.path.isabs(path):
            return ([path] if os.path.isfile(path) else [], "absolute")
        if path.startswith(REPO_PREFIXES):
            p = os.path.join(self.root, path)
            return ([p] if os.path.isfile(p) else [], "repo-relative")
        head = path.partition("/")[0]
        if head in self.lib_dirs():
            # <lib>/…-qualified: the fleet repo IS named, never guess another.
            p = os.path.join(self.libs_root, path)
            return ([p] if os.path.isfile(p) else [], "lib-qualified")
        # No repo prefix, no fleet head: match the path as a suffix against
        # every fleet repo (cheap existence tests, cached per cite path).
        cands = [
            os.path.join(self.libs_root, d, path)
            for d in self.lib_dirs()
            if os.path.isfile(os.path.join(self.libs_root, d, path))
        ]
        return (cands, "fleet-suffix")

    def classify(self, cite: str, path: str, end: int, slash: bool) -> tuple[str, str]:
        """Return (class, reason) for one cite."""
        if slash:
            cands, origin = self.resolve_slash_path(path)
            if not cands:
                return ("MISSING", f"no such file ({origin}: {path})")
            if len(cands) > 1:
                shown = ", ".join(sorted(cands)[:3])
                more = f", … +{len(cands) - 3} more" if len(cands) > 3 else ""
                return (
                    "AMBIGUOUS",
                    f"{origin} matches {len(cands)} files: {shown}{more}",
                )
            target = cands[0]
        else:
            cands = self.basenames().get(path, [])
            if not cands:
                return ("UNKNOWN", f"no file named {path} under the repo root")
            if len(cands) > 1:
                shown = ", ".join(sorted(cands)[:3])
                more = f", … +{len(cands) - 3} more" if len(cands) > 3 else ""
                return (
                    "AMBIGUOUS",
                    f"basename matches {len(cands)} files: {shown}{more}",
                )
            target = cands[0]
        n = self.line_count(target)
        if end > n:
            return ("OUT-OF-RANGE", f"{target} has {n} lines, cite ends at {end}")
        return ("OK", target)

    # -- driving ------------------------------------------------------------

    def sweep(self, report: str) -> tuple[list[str], collections.Counter]:
        findings: list[str] = []
        counts: collections.Counter = collections.Counter()
        label = (
            os.path.relpath(report, self.root)
            if report.startswith(self.root)
            else report
        )
        with open(report, encoding="utf-8", errors="replace") as f:
            lines = f.read().split("\n")
        for lineno, line in enumerate(lines, start=1):
            for m in PATH_RE.finditer(line):
                # group(1) stops before the extension; the whole match minus
                # the :line tail is the cited path (the throwaway's trick).
                path = m.group(0).rsplit(":", 1)[0]
                end = int(m.group(3) or m.group(2))
                cls, reason = self.classify(m.group(0), path, end, slash=True)
                counts[cls] += 1
                if cls != "OK":
                    findings.append(
                        f"{label}:{lineno}: {m.group(0)} -> {cls} ({reason})"
                    )
            for m in BARE_RE.finditer(line):
                name, end = m.group(1), int(m.group(3) or m.group(2))
                cls, reason = self.classify(m.group(0), name, end, slash=False)
                counts[cls] += 1
                if cls != "OK":
                    findings.append(
                        f"{label}:{lineno}: {m.group(0)} -> {cls} ({reason})"
                    )
        return findings, counts


def run(
    files: list[str], root: str, libs_root: str
) -> tuple[list[str], collections.Counter]:
    sweeper = Sweeper(root, libs_root)
    findings: list[str] = []
    counts: collections.Counter = collections.Counter()
    for report in files:
        f, c = sweeper.sweep(report)
        findings.extend(f)
        counts.update(c)
    return findings, counts


def self_test() -> int:
    import glob
    import tempfile

    failures: list[str] = []

    def expect(name: str, ok: bool) -> None:
        print(f"selftest: {'PASS' if ok else 'FAIL'} — {name}")
        if not ok:
            failures.append(name)

    with tempfile.TemporaryDirectory() as tmp:
        root = os.path.join(tmp, "root")
        libs = os.path.join(tmp, "libs")
        for d in (
            "internal/server",
            "internal/domain",
            "internal/web",
            "docs/review/libs",
        ):
            os.makedirs(os.path.join(root, d))
        for d in ("go-fixture", "go-other/nested", "go-clone/nested"):
            os.makedirs(os.path.join(libs, d))

        def put(path: str, n: int, marker: str) -> None:
            with open(path, "w", encoding="utf-8") as f:
                f.writelines(f"line {i} {marker}\n" for i in range(1, n + 1))

        put(os.path.join(root, "internal/server/a.go"), 100, "repo")
        put(os.path.join(root, "internal/domain/dup.go"), 10, "one")
        put(os.path.join(root, "internal/web/dup.go"), 10, "two")
        put(os.path.join(libs, "go-fixture/store.go"), 200, "lib")
        put(os.path.join(libs, "go-other/nested/only.go"), 50, "suffix")
        put(os.path.join(libs, "go-other/nested/twin.go"), 50, "twinA")
        put(os.path.join(libs, "go-clone/nested/twin.go"), 50, "twinB")

        report = os.path.join(root, "docs/review/libs/report.md")
        with open(report, "w", encoding="utf-8") as f:
            f.write(
                "good `internal/server/a.go:50`, boundary `internal/server/a.go:100`\n"  # OK, OK
                "stale `internal/server/a.go:150` — the file shrank\n"  # OUT-OF-RANGE
                "en-dash range `internal/server/a.go:10–20`\n"  # OK
                "lib-qualified `go-fixture/store.go:42` and lib-stale `go-fixture/store.go:999`\n"
                "fleet suffix unique `nested/only.go:3`\n"  # OK
                "fleet suffix twin `nested/twin.go:3` — two repos carry it\n"  # AMBIGUOUS
                "bare unique `a.go:100` ok but `a.go:101` off by one\n"  # OK, OUT-OF-RANGE
                "bare ambiguous `dup.go:5`\n"  # AMBIGUOUS
                "bare unknown `nowhere.go:7`\n"  # UNKNOWN
                "slash missing `internal/server/gone.go:5`\n"  # MISSING
            )
        clean = os.path.join(root, "docs/review/libs/clean.md")
        with open(clean, "w", encoding="utf-8") as f:
            f.write("only good cites: `internal/server/a.go:1` and `a.go:100`\n")

        files = sorted(glob.glob(os.path.join(root, "docs/review/libs/*.md")))
        findings, counts = run(files, root, libs)
        got = dict(counts)
        # report.md: OK 6 (a.go:50, a.go:100 boundary, en-dash range,
        # store.go:42, only.go:3, a.go:100 bare) + clean.md's 2 OK cites.
        expect("OK counted (report 6 + clean 2)", got.get("OK") == 8)
        expect("OUT-OF-RANGE fires (repo, lib, bare)", got.get("OUT-OF-RANGE") == 3)
        expect("MISSING fires (slash path, no file)", got.get("MISSING") == 1)
        expect("AMBIGUOUS fires (basename + fleet suffix)", got.get("AMBIGUOUS") == 2)
        expect("UNKNOWN fires (basename matches none)", got.get("UNKNOWN") == 1)
        expect(
            "bad cite reported as <file>:<line>: <cite> -> <class>",
            any(
                l.startswith(
                    "docs/review/libs/report.md:2: internal/server/a.go:150 -> OUT-OF-RANGE"
                )
                for l in findings
            ),
        )
        expect(
            "lib-qualified cite resolves against the libs root (in-range silent, stale loud)",
            not any("store.go:42" in l for l in findings)
            and any("go-fixture/store.go:999 -> OUT-OF-RANGE" in l for l in findings),
        )
        cfind, ccounts = run([clean], root, libs)
        expect("clean file has zero findings", not cfind and ccounts["OK"] == 2)

    if failures:
        print(f"selftest: FAILED — {len(failures)} case(s)")
        return 1
    print("selftest: 9 cases passed — every class can fire")
    return 0


def main(argv: list[str]) -> int:
    default_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    ap = argparse.ArgumentParser(
        prog="citation-sweep.py",
        description="verify file:line citations in the fleet-libs reports",
    )
    ap.add_argument(
        "files", nargs="*", help="report files (default: docs/review/libs/*.md)"
    )
    ap.add_argument(
        "--root", default=default_root, help="repo root (default: this script's repo)"
    )
    ap.add_argument(
        "--libs-root", default=DEFAULT_LIBS_ROOT, help="fleet root for <lib>/… cites"
    )
    ap.add_argument(
        "--self-test", action="store_true", help="run the fixture self-test"
    )
    args = ap.parse_args(argv)

    if args.self_test:
        return self_test()

    files: list[str] = []
    if args.files:
        for name in args.files:
            p = (
                name
                if os.path.isabs(name) or os.path.exists(name)
                else os.path.join(args.root, name)
            )
            if not os.path.isfile(p):
                print(f"citation-sweep: no such file: {name}", file=sys.stderr)
                return 2
            files.append(p)
    else:
        import glob as _glob

        files = sorted(_glob.glob(os.path.join(args.root, DEFAULT_REPORT_GLOB)))
        if not files:
            print(
                f"citation-sweep: no reports under {os.path.join(args.root, DEFAULT_REPORT_GLOB)}",
                file=sys.stderr,
            )
            return 2

    findings, counts = run(files, args.root, args.libs_root)
    for line in findings:
        print(line)
    total = sum(counts.values())
    parts = ", ".join(f"{c} {counts.get(c, 0)}" for c in CLASSES)
    print(f"citation-sweep: {len(files)} report(s), {total} cite(s): {parts}")
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
