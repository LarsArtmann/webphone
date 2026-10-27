#!/usr/bin/env bash
# Pre-commit drift-gate selftest (T9, 2026-10-27).
#
# Pins .githooks/pre-commit's contract the way commit-msg-guard-selftest
# pins the dep guard: the hook is only as good as its refusal path.
# Exercises whitespace-drift THROUGH the hook (not the script directly)
# against a THROWAWAY index (GIT_INDEX_FILE — the real index is never
# touched), asserting:
#   1. whitespace-only staged change in an owned file  => refuse (exit 1)
#   2. content staged change in an owned file          => allow  (exit 0)
#   3. markdown-only staged change                     => allow  (exit 0)
set -euo pipefail

repo="$(git rev-parse --show-toplevel)"
hook="$repo/.githooks/pre-commit"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

fail() {
	echo "pre-commit-guard-selftest: FAIL — $1" >&2
	exit 1
}

[ -f "$hook" ] || fail "hook missing at $hook"

run_case() {
	local name="$1" expect="$2"
	local got=0
	bash "$hook" >/dev/null 2>&1 || got=$?
	if [ "$got" -ne "$expect" ]; then
		fail "$name: expected exit $expect, got $got"
	fi
}

export GIT_INDEX_FILE="$tmp/index"
git read-tree HEAD

# Case 1: whitespace-only re-indent of an owned file => refuse.
sed '2s/^/  /' "$repo/flake.nix" >"$tmp/ws.nix"
b=$(git hash-object -w "$tmp/ws.nix")
git update-index --add --cacheinfo "100644,$b,flake.nix"
run_case "whitespace-only staged drift is refused" 1

# Case 2: real content change in an owned file => allow.
sed '2s/^/# gate /' "$repo/flake.nix" >"$tmp/content.nix"
b=$(git hash-object -w "$tmp/content.nix")
git update-index --add --cacheinfo "100644,$b,flake.nix"
run_case "content staged change is allowed" 0

# Case 3: markdown-only change => allow (formatter-unowned).
git read-tree HEAD
printf 'selftest scratch\n' >"$tmp/note.md"
b=$(git hash-object -w "$tmp/note.md")
git update-index --add --cacheinfo "100644,$b,NOTE.md"
run_case "markdown-only staged change is allowed" 0

unset GIT_INDEX_FILE
echo "pre-commit-guard-selftest: 3 passed"
