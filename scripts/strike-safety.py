#!/usr/bin/env python3
"""strike-safety — the pre-write guard for a docs-health strike sweep.

Two defects in the archive tooling are known and unfixed upstream, and both
corrupt silently:

1. ESCAPED PIPES. A table cell may legally contain `\\|` inside a code span.
   The strike writer splits rows on `|`, so a row holding an escaped pipe is
   cut one cell too far and the strike starts or ends in the wrong cell — the
   content crosses a boundary and the row no longer means what it said. The
   damage gate `repair-mangle` catches the RESULT (doubled tildes), never the
   cause, and only after the file is written.

2. UNVALIDATED VERDICT SHAPE. `annotate rows <file> "3:v:UNMAPPED"` is written
   verbatim; the banned-word check in `check-rows` is a list of words, not a
   grammar, so a garbage verdict head survives every gate and reads as an
   authoritative ruling.

This script runs BEFORE the writer. It never edits anything.

    strike-safety.py rows <file>...            refuse escaped pipes / ragged rows
    strike-safety.py verdict <value>...        refuse a non-canonical verdict
    strike-safety.py --self-test               prove both refusals can fire

Exit 0 = safe to write; 1 = refused (the message names the line); 2 = usage.
"""

from __future__ import annotations

import re
import sys

# The house verdict vocabulary (docs-health SKILL.md ANNOTATE + the writers in
# annotate.go). A verdict value MUST start with one of these heads.
VERDICT_HEADS = (
    "done at",
    "done —",
    "done -",
    "won't implement",
    "wont implement",
    "not-do",
    "→",  # routed: "→ <home>"
    "superseded",
    "(docs-health pass",
)


def split_cells(row: str) -> list[str]:
    """Split a markdown table row on UNESCAPED pipes only."""
    cells: list[str] = []
    buf: list[str] = []
    i = 0
    while i < len(row):
        ch = row[i]
        if ch == "\\" and i + 1 < len(row):
            buf.append(row[i : i + 2])
            i += 2
            continue
        if ch == "|":
            cells.append("".join(buf))
            buf = []
            i += 1
            continue
        buf.append(ch)
        i += 1
    cells.append("".join(buf))
    return cells


def scan_rows(path: str) -> list[str]:
    """Return findings for one file."""
    findings: list[str] = []
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            lines = f.read().split("\n")
    except OSError as err:
        return [f"{path}: cannot read: {err}"]

    header_cells = 0
    in_table = False
    for lineno, line in enumerate(lines, start=1):
        stripped = line.strip()
        if not stripped.startswith("|"):
            in_table = False
            header_cells = 0
            continue
        cells = split_cells(stripped)
        if not in_table:
            in_table = True
            header_cells = len(cells)
            continue
        # A separator row (|---|:--:|) carries no content.
        if re.fullmatch(r"\|[\s|:-]+\|?", stripped):
            continue
        if len(cells) != header_cells:
            findings.append(
                f"{path}:{lineno}: row has {len(cells)} cells, header has {header_cells} "
                f"— an escaped pipe or a lost delimiter (a strike here would cross a cell boundary)"
            )
            continue
        if "~~" in stripped and "\\|" in stripped:
            findings.append(
                f"{path}:{lineno}: struck row contains an escaped pipe (`\\|`) — "
                f"the strike writer would split the cell in the wrong place; reword or un-strike first"
            )
    return findings


def check_verdict(value: str) -> str | None:
    """Return a finding when `value` is not a canonical verdict, else None."""
    v = value.strip()
    if not v:
        return "empty verdict"
    low = v.lower()
    # Bare evidence with no head is the common mistake; the writer accepts it.
    for head in VERDICT_HEADS:
        if low.startswith((head, "**" + head)):
            return None
    return f"non-canonical verdict {value!r} — must start with one of: " + ", ".join(
        repr(h) for h in VERDICT_HEADS
    )


EXPECTED_CASES = range(13)


def self_test() -> int:
    import os
    import tempfile

    failures: list[str] = []

    def expect(name: str, ok: bool) -> None:
        print(f"selftest: {'PASS' if ok else 'FAIL'} — {name}")
        if not ok:
            failures.append(name)

    with tempfile.TemporaryDirectory() as tmp:
        # Case 1: an escaped pipe inside a code span in a struck cell.
        p = os.path.join(tmp, "escaped.md")
        with open(p, "w") as f:
            f.write(
                "| # | Task | Impact |\n"
                "| - | ---- | ------ |\n"
                "| 1 | `a \\| b` ~~old~~ done at `abc1234` | High |\n"
            )
        expect("escaped pipe in a struck row is refused", len(scan_rows(p)) >= 1)

        # Case 2: the same escaped pipe WITHOUT a strike is legal (it is a
        # plain table cell) — the guard must not become a corpus linter.
        p = os.path.join(tmp, "legal.md")
        with open(p, "w") as f:
            f.write(
                "| # | Task | Impact |\n"
                "| - | ---- | ------ |\n"
                "| 1 | `a \\| b` is a legal cell | High |\n"
            )
        expect("a clean escaped-pipe cell is allowed", scan_rows(p) == [])

        # Case 3: a ragged row (a lost delimiter) is refused.
        p = os.path.join(tmp, "ragged.md")
        with open(p, "w") as f:
            f.write(
                "| # | Task | Impact |\n| - | ---- | ------ |\n| 1 | two cells merged High |\n"
            )
        expect("ragged row is refused", len(scan_rows(p)) >= 1)

        # Case 4: a well-formed struck row passes.
        p = os.path.join(tmp, "clean.md")
        with open(p, "w") as f:
            f.write(
                "| # | Task | Impact |\n"
                "| - | ---- | ------ |\n"
                "| ~~1~~ | ~~old~~ done at `abc1234` | ~~High~~ |\n"
            )
        expect("well-formed struck row passes", scan_rows(p) == [])

    # Verdict vocabulary.
    expect("`done at <hash>` accepted", check_verdict("done at `abc1234`") is None)
    expect(
        "`done — evidence` accepted", check_verdict("done — docs/lessons.md") is None
    )
    expect("routed `→ TODO_LIST` accepted", check_verdict("→ TODO_LIST §Owner") is None)
    expect(
        "`**Won't implement — x.**` accepted",
        check_verdict("**Won't implement — noise.**") is None,
    )
    expect("NOT-DO accepted", check_verdict("NOT-DO — out of scope.") is None)
    expect(
        "superseded accepted", check_verdict("Superseded — 2026-11-07 report.") is None
    )
    expect("`UNMAPPED` refused", check_verdict("UNMAPPED") is not None)
    expect("empty verdict refused", check_verdict("  ") is not None)
    expect("bare evidence refused", check_verdict("it works now") is not None)

    if failures:
        print(f"selftest: FAILED — {len(failures)} case(s)")
        return 1
    print(f"selftest: {len(EXPECTED_CASES)} passed — both refusals fire")
    return 0


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print(__doc__, file=sys.stderr)
        return 2
    verb, args = argv[1], argv[2:]
    if verb == "--self-test":
        return self_test()
    if verb == "rows":
        if not args:
            print("strike-safety: rows needs at least one file", file=sys.stderr)
            return 2
        findings: list[str] = []
        for path in args:
            findings.extend(scan_rows(path))
        if findings:
            for f in findings:
                print(f"strike-safety: REFUSED: {f}", file=sys.stderr)
            return 1
        print(f"strike-safety: {len(args)} file(s) safe to strike")
        return 0
    if verb == "verdict":
        problems = [p for p in (check_verdict(v) for v in args) if p]
        if problems:
            for p in problems:
                print(f"strike-safety: REFUSED: {p}", file=sys.stderr)
            return 1
        print(f"strike-safety: {len(args)} verdict(s) canonical")
        return 0
    print(f"strike-safety: unknown verb {verb!r}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
