#!/usr/bin/env python3
"""Pre-upload frontmatter contract check for the catalog export.

The hub (eventcatalog-hub) publishes a versioned JSON Schema contract for
every catalog page's frontmatter
(https://catalog.home.lan/contracts/catalog-frontmatter.v<N>.schema.json
— the byte-identical dist artifact of its merge gate's schema). Validating
the export HERE, before pushing, means a schema break surfaces in this
repo's CI instead of as a red hub build naming this source.

The contract's shape is deliberately simple (per-collection `required` and
`allowed` key sets + open `x-*` extension namespaces), so this check needs
no JSON Schema library: it reads required/allowed the same way the hub's
merge gate does, walks the export tree, and fails on the two classes the
hub cares about — a page MISSING a required key, or carrying a key outside
allowed + the extension pattern. Collection = the first path segment
(events/, services/, ...); nested sidecar pages (versioned copies,
ubiquitous-language) are exempt from required like the hub's primary-page
rule, but unknown-key checking applies everywhere.

Usage:
    scripts/check-catalog-contract.py --export <tree> [--contract URL | FILE]
    scripts/check-catalog-contract.py --self-test

--contract defaults to the live contract URL; pass a local file for
offline runs. Exit 0 clean · 1 violations · 2 usage/fetch failure.
"""

import argparse
import json
import re
import sys
import urllib.request
from pathlib import Path

DEFAULT_CONTRACT_URL = (
    "https://catalog.home.lan/contracts/catalog-frontmatter.v2.schema.json"
)
EXTENSION_PATTERN = re.compile(r"^x-")


def load_contract(source):
    if "://" not in source:
        return json.loads(Path(source).read_text())
    with urllib.request.urlopen(source, timeout=30) as response:
        return json.loads(response.read().decode())


def parse_frontmatter(text):
    """Return the frontmatter's top-level KEY LIST (order-insensitive),
    or None when the file has no --- fence. Line-level only: the values
    are the hub's business, the key set is this check's."""
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        return None
    for end in range(1, len(lines)):
        if lines[end].strip() == "---":
            break
    else:
        return None
    keys = []
    for line in lines[1:end]:
        if not line.strip() or line[0] in " \t-#" or line.startswith("---"):
            continue
        match = re.match(r"^([A-Za-z_][A-Za-z0-9_-]*):", line)
        if match:
            keys.append(match.group(1))
    return keys


def violations(export_root, contract):
    problems = []
    root = Path(export_root)
    if not root.is_dir():
        return [f"export tree missing: {root}"]
    for collection, spec in sorted(contract.get("properties", {}).items()):
        base = root / collection
        if not base.is_dir():
            continue
        required = list(spec.get("required", []))
        allowed = set(spec.get("allowed", [])) | set(required)
        for mdx in sorted(base.rglob("*.mdx")):
            keys = parse_frontmatter(mdx.read_text(encoding="utf-8", errors="replace"))
            if keys is None:
                problems.append(
                    f"{collection}: {mdx.relative_to(root)} has no --- frontmatter fence"
                )
                continue
            rel = mdx.relative_to(root).parts
            primary = len(rel) == 2 or (len(rel) == 3 and mdx.name == "index.mdx")
            if primary:
                missing = [key for key in required if key not in keys]
                if missing:
                    problems.append(
                        f"{collection}: {mdx.relative_to(root)} missing required {missing}"
                    )
            unknown = sorted(
                key
                for key in keys
                if key not in allowed and not EXTENSION_PATTERN.match(key)
            )
            if unknown:
                problems.append(
                    f"{collection}: {mdx.relative_to(root)} carries unknown key(s) {unknown}"
                )
    return problems


def self_test():
    import tempfile

    contract = {
        "properties": {
            "events": {
                "required": ["id", "name", "version", "summary"],
                "allowed": [
                    "id",
                    "name",
                    "version",
                    "summary",
                    "owners",
                    "producers",
                    "consumers",
                    "schemaPath",
                ],
            },
        }
    }
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "events/good.v1.0.0").mkdir(parents=True)
        (root / "events/good.v1.0.0/index.mdx").write_text(
            "---\nid: good\nname: Good\nversion: 1.0.0\nsummary: s\nowners:\n    - t\n---\nb\n"
        )
        (root / "events/bad.v1.0.0").mkdir(parents=True)
        (root / "events/bad.v1.0.0/index.mdx").write_text(
            "---\nid: bad\nsummery: typo-key\nx-custom: fine\n---\nb\n"
        )
        (root / "events/nofence.v1.0.0").mkdir(parents=True)
        (root / "events/nofence.v1.0.0/index.mdx").write_text("no fence\n")
        problems = violations(root, contract)
        assert any(
            "bad.v1.0.0" in p and "missing required ['name', 'version', 'summary']" in p
            for p in problems
        ), problems
        assert any("bad.v1.0.0" in p and "summery" in p for p in problems), problems
        assert any("nofence" in p for p in problems), problems
        assert not any("good.v1.0.0" in p for p in problems), problems
        # x-* extensions stay open by contract
        assert not any("x-custom" in p for p in problems), problems
        assert violations(root / "does-not-exist", contract) == [
            "export tree missing: " + str(root / "does-not-exist")
        ]
    print("check-catalog-contract self-test: OK")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--export", help="the catalog export tree to check")
    parser.add_argument(
        "--contract", default=DEFAULT_CONTRACT_URL, help="contract URL or local file"
    )
    parser.add_argument("--self-test", action="store_true")
    args = parser.parse_args()
    if args.self_test:
        self_test()
        return 0
    if not args.export:
        parser.print_usage()
        return 2
    try:
        contract = load_contract(args.contract)
    except (OSError, ValueError) as error:
        print(
            f"check-catalog-contract: contract unreadable at {args.contract}: {error}",
            file=sys.stderr,
        )
        return 2
    schema_version = contract.get("x-hub-schema-version")
    print(f"check-catalog-contract: contract v{schema_version} from {args.contract}")
    problems = violations(args.export, contract)
    for problem in problems:
        print(f"  VIOLATION: {problem}")
    if problems:
        print(
            f"check-catalog-contract: {len(problems)} violation(s) — fix the exporter, never widen the contract",
            file=sys.stderr,
        )
        return 1
    print("check-catalog-contract: clean")
    return 0


if __name__ == "__main__":
    sys.exit(main())
