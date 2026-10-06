#!/usr/bin/env python3
r"""md-table-shape: detect NEAR-ALIGNED markdown tables (the churn-bait class).

A table column is NEAR-ALIGNED when its cell widths vary across rows AND at
least one cell in that column carries extra/unbalanced padding. Pure-compact
columns (single-space cells, naturally varying widths) are the daemon-canonical
shape and are clean; fully hand-aligned columns (uniform padded widths) are a
separate question (briefing row: daemon-format coupling home) and are NOT
flagged here. Only the mixed state is churn bait: the next formatter pass
rewrites it, so diffs lie about being substantive.

Limitations: cells are compared as raw text (escaped `\|` inside a cell is
kept as content and never splits a cell).

Usage:
  md-table-shape.py [--self-test] [paths ...]   # default: docs/status/ (live,
                                               # i.e. archived/ excluded — a
                                               # subset of the D25.1 scope)

History (2026-10-06, one session, four dogfood rounds): v1 reported 39
findings — wrapped-row fragments and escaped-pipe splits; v3 fixed both and
reported 25; v4 measures DISPLAY width (unicodedata W/F = 2) because the
daemon's aligner pads emoji columns to terminal cells, not codepoints — the
daemon's own 16:18 aligning pass then left exactly ONE real finding
(samber-do scorecard, normalized to compact on the spot). The corpus now
reports ZERO. The 03-40 canary verdict from the same pass: the daemon
REFLOWS compact tables to fully-aligned (content-inert) — the daemon is the
ALIGNER; near-aligned mixes come from sessions appending compact rows onto
aligned tables. Rewriting archived snapshots was declined (hypothetical
churn vs real churn); D1.4 owns the convention.

Exit 0 = clean (or self-test pass), exit 1 = findings (or self-test failure).
Committed 2026-10-06 as the d.1 fix: instruments cited in reports must be
committed instruments (round-4 brutal review e.2).
"""

from __future__ import annotations

import re
import sys
import unicodedata
from pathlib import Path

DEFAULT_PATHS = ["docs/status/"]


def is_archived(path: Path) -> bool:
    return "archived" in path.parts


def split_cells(line: str) -> list[str]:
    stripped = line.strip()
    stripped = stripped.removeprefix("|")
    stripped = stripped.removesuffix("|")
    return re.split(r"(?<!\\)\|", stripped)


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


def display_width(text: str) -> int:
    """Terminal cells, not codepoints: the daemon pads emoji/wide glyphs to
    DISPLAY width, so alignment must be compared on the same axis."""
    return sum(2 if unicodedata.east_asian_width(char) in "WF" else 1 for char in text)


def is_compact(cell: str) -> bool:
    content = cell.strip()
    if content == "":
        return len(cell) <= 2
    return cell == f" {content} "


def table_blocks(lines: list[str]):
    """Yield (header_physical_index, [header, separator, *body]) per GFM table.

    Rows may wrap mid-cell (a physical line without a trailing `|` continues
    the row until a line ends with one); blank lines never continue a row, so
    prose after a table cannot glue onto it.
    """

    def read_row(index: int) -> tuple[str, int]:
        parts = [lines[index]]
        cursor = index
        while not parts[-1].rstrip().endswith("|"):
            if cursor + 1 >= len(lines) or not lines[cursor + 1].strip():
                break
            cursor += 1
            parts.append(lines[cursor])
        return " ".join(part.strip() for part in parts), cursor + 1

    index = 0
    while index < len(lines):
        if (
            lines[index].strip().startswith("|")
            and index + 1 < len(lines)
            and is_separator(lines[index + 1])
        ):
            block = [lines[index], lines[index + 1]]
            cursor = index + 2
            while cursor < len(lines) and lines[cursor].strip().startswith("|"):
                row, cursor = read_row(cursor)
                block.append(row)
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
            widths = {display_width(cell) for cell in cells}
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

WRAPPED_TABLE = """| name | ruling |
| --- | --- |
| anchor | FIRST UNPUSHED COMMIT's timestamp (`git log origin/main..HEAD \
--format=%cI | tail -1`) — the observable anchor |
| changelog | code-path changes log; tests and docs stay silent |
"""

WRAPPED_THEN_PROSE = (
    WRAPPED_TABLE
    + """
Prose after the table must not glue onto the last row: the empty line
bounds the join, and this paragraph never becomes cell content.
"""
)


ESCAPED_PIPE_TABLE = """| name | value |
| --- | --- |
| anchor | `git log origin/main..HEAD --format=%cI \\| tail -1` |
| other | plain |
"""

EMOJI_ALIGNED_TABLE = """| Impact   | Owner |
| -------- | ----- |
| \U0001f525\U0001f525\U0001f525   | Lars |
| \U0001f525\U0001f525     | Lars |
"""


def self_test() -> int:
    cases = [
        ("mixed widths flag", MIXED_TABLE, 1),
        ("compact is clean", COMPACT_TABLE, 0),
        ("fully aligned is clean (D1.4's question, not ours)", ALIGNED_TABLE, 0),
        ("unbalanced padding flags", UNBALANCED_TABLE, 1),
        ("mid-cell wrapped rows stay clean", WRAPPED_TABLE, 0),
        ("prose after a wrapped table never glues", WRAPPED_THEN_PROSE, 0),
        ("escaped pipes never split a cell", ESCAPED_PIPE_TABLE, 0),
        (
            "display-width-aligned emoji columns are clean",
            EMOJI_ALIGNED_TABLE,
            0,
        ),
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
