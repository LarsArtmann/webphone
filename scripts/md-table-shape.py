#!/usr/bin/env python3
"""md-table-shape: detect NEAR-ALIGNED markdown tables (the churn-bait class).

A table column is NEAR-ALIGNED when its cell widths vary across rows AND at
least one cell in that column carries extra/unbalanced padding. Pure-compact
columns (single-space cells, naturally varying widths) are the daemon-canonical
shape and are clean; fully hand-aligned columns (uniform padded widths) are a
separate question (briefing row: daemon-format coupling home) and are NOT
flagged here. Only the mixed state is churn bait: the next formatter pass
rewrites it, so diffs lie about being substantive.

Limitations: naive pipe split (no \\| escapes, no pipes inside code spans) —
fine for this corpus, which contains none in table rows.

Usage:
  md-table-shape.py [--self-test] [paths ...]   # default: docs/status/ (live,
                                               # i.e. archived/ excluded — the
                                               # D25.1 sweep scope)

The 2026-10-06 full-corpus run (all docs/ + root md) reported 39 findings
across 10 files (6 archived, FEATURES.md, 4 live planning docs incl. the
briefing's rows-29–32/appended-rows mix): recorded as evidence for the D1.4
daemon-format coupling-home verdict, NOT normalized — rewriting archived
snapshots trades hypothetical churn for real churn.

Exit 0 = clean (or self-test pass), exit 1 = findings (or self-test failure).
Committed 2026-10-06 as the d.1 fix: instruments cited in reports must be
committed instruments (round-4 brutal review e.2).
"""

from __future__ import annotations

import sys
from pathlib import Path

DEFAULT_PATHS = ["docs/status/"]


def is_archived(path: Path) -> bool:
    return "archived" in path.parts


def split_cells(line: str) -> list[str]:
    stripped = line.strip()
    if stripped.startswith("|"):
        stripped = stripped[1:]
    if stripped.endswith("|"):
        stripped = stripped[:-1]
    return stripped.split("|")


def is_separator(line: str) -> bool:
    if not line.strip().startswith("|"):
        return False
    cells = split_cells(line)
    if not cells:
        return False
    for cell in cells:
        content = cell.strip()
        if not content or content.strip(":") != "-" * len(content.strip(":")):
            return False
    return True


def is_compact(cell: str) -> bool:
    content = cell.strip()
    if content == "":
        return len(cell) <= 2
    return cell == f" {content} "


def table_blocks(lines: list[str]):
    """Yield (header_index, [header, separator, *body]) for each GFM table."""
    index = 0
    while index < len(lines):
        if lines[index].strip().startswith("|") and index + 1 < len(lines) and is_separator(lines[index + 1]):
            block = [lines[index], lines[index + 1]]
            cursor = index + 2
            while cursor < len(lines) and lines[cursor].strip().startswith("|"):
                block.append(lines[cursor])
                cursor += 1
            yield index, block
            index = cursor
        else:
            index += 1


def find_near_aligned(lines: list[str]) -> list[str]:
    findings: list[str] = []
    for header_index, block in table_blocks(lines):
        rows = [split_cells(line) for line in block]
        data_rows = rows[0:1] + rows[2:]  # skip the separator (row 2 of the block)
        width = max(len(cells) for cells in data_rows) if data_rows else 0
        for column in range(width):
            cells = [row[column] if column < len(row) else "" for row in data_rows]
            widths = {len(cell) for cell in cells}
            padded = sum(1 for cell in cells if not is_compact(cell))
            if len(widths) > 1 and padded > 0:
                findings.append(
                    f"line {header_index + 1}: column {column + 1} near-aligned "
                    f"(widths {sorted(widths)}, {padded} padded cell(s))"
                )
    return findings


def scan_path(path: Path, skip_archived: bool = False) -> list[str]:
    findings: list[str] = []
    files = sorted(path.rglob("*.md")) if path.is_dir() else [path]
    for file in files:
        if skip_archived and is_archived(file):
            continue
        try:
            lines = file.read_text(encoding="utf-8").splitlines()
        except (OSError, UnicodeDecodeError) as error:
            findings.append(f"{file}: unreadable ({error})")
            continue
        for finding in find_near_aligned(lines):
            findings.append(f"{file}: {finding}")
    return findings


MIXED_TABLE = """| name | note |
| ---- | ---- |
| a    | short |
| b | a much longer note |
"""

COMPACT_TABLE = """| name | note |
| --- | --- |
| a | short |
| b | a much longer note |
"""

ALIGNED_TABLE = """| name | note             |
| ---- | ---------------- |
| a    | short            |
| b    | a longer note    |
"""

UNBALANCED_TABLE = """| name | note |
| --- | --- |
|  a  | x |
| b | y |
"""


def self_test() -> int:
    cases = [
        ("mixed widths flag", MIXED_TABLE, 1),
        ("compact is clean", COMPACT_TABLE, 0),
        ("fully aligned is clean (D1.4's question, not ours)", ALIGNED_TABLE, 0),
        ("unbalanced padding flags", UNBALANCED_TABLE, 1),
    ]
    failures = 0
    for label, document, expected in cases:
        count = len(find_near_aligned(document.splitlines()))
        status = "PASS" if count == expected else "FAIL"
        if count != expected:
            failures += 1
        print(f"self-test: {status} {label} (findings={count}, expected={expected})")
    return 1 if failures else 0


def main(argv: list[str]) -> int:
    if "--self-test" in argv:
        return self_test()
    paths = [Path(argument) for argument in argv if not argument.startswith("--")]
    default = not paths
    if default:
        paths = [Path(argument) for argument in DEFAULT_PATHS]
    findings: list[str] = []
    for path in paths:
        if not path.exists():
            print(f"md-table-shape: no such path: {path}", file=sys.stderr)
            return 2
        findings.extend(scan_path(path, skip_archived=default))
    for finding in findings:
        print(f"md-table-shape: {finding}")
    print(f"md-table-shape: {len(findings)} finding(s)")
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
