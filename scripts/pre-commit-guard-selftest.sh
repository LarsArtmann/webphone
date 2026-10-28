#!/usr/bin/env bash
# Pre-commit gate selftest (T9, 2026-10-27; JS syntax cases added the same
# evening).
#
# Pins .githooks/pre-commit's contract the way commit-msg-guard-selftest
# pins the dep guard: the hook is only as good as its refusal path.
# Exercises both gates THROUGH the hook (not the scripts directly) against
# a THROWAWAY index (GIT_INDEX_FILE — the real index is never touched),
# asserting:
#   1. whitespace-only staged change in an owned file   => refuse (exit 1)
#   2. content staged change in an owned file           => allow  (exit 0)
#   3. markdown-only staged change                      => allow  (exit 0)
#   4. served asset with a JS syntax error              => refuse (exit 1)
#   5. served asset using ES-module syntax              => allow  (exit 0)
#   6. vendored bundle (generated, never hand-edited)   => allow  (exit 0)
#
# Cases 4-6 need node on PATH; the hook itself fails OPEN when it is
# missing, so the selftest WARNS instead of failing in that situation.
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

if ! command -v node >/dev/null 2>&1; then
	echo "pre-commit-guard-selftest: WARN — node not on PATH, JS syntax cases skipped" >&2
	unset GIT_INDEX_FILE
	echo "pre-commit-guard-selftest: 3 passed"
	exit 0
fi

# Case 4: a staged served asset that does not parse => refuse.
git read-tree HEAD
printf '(function () {\n  if (\n})();\n' >"$tmp/broken.js"
b=$(git hash-object -w "$tmp/broken.js")
git update-index --add --cacheinfo "100644,$b,internal/web/assets/shell.js"
run_case "staged JS syntax error is refused" 1

# Case 5: a staged ES-module asset (import/export at top level) => allow.
git read-tree HEAD
printf 'import { read } from "./io.js";\nexport const value = read;\n' >"$tmp/module.js"
b=$(git hash-object -w "$tmp/module.js")
git update-index --add --cacheinfo "100644,$b,internal/web/assets/island/app/main.js"
run_case "a staged ES-module asset is allowed" 0

# Case 6: the vendored bundle is generated => skipped, even when it would
# not parse (it is nobody's hand edit and update.sh owns it).
git read-tree HEAD
printf 'not javascript at all {{{\n' >"$tmp/vendor.js"
b=$(git hash-object -w "$tmp/vendor.js")
git update-index --add --cacheinfo "100644,$b,internal/web/assets/vendor/sip.min.js"
run_case "the vendored bundle bypasses the syntax gate" 0

# Cases 7-8 (2026-10-28): whitespace-drift's formatter adjudication, in a
# THROWAWAY repo (the real repo's HEAD is formatter-clean by CI, so a
# whitespace-only staged delta that IS canonical cannot be staged against
# it — the scratch repo's HEAD is deliberately gofmt-dirty to stage one).
# The adjudication resolves the treefmt config from the REAL repo (script
# location), so these need nix; like the node cases they WARN+skip without.
# The throwaway index must NOT leak into the scratch repo (it is a separate
# repository with its own index and HEAD).
unset GIT_INDEX_FILE
scratch="$(mktemp -d)"
trap 'rm -rf "$tmp" "$scratch"' EXIT
printf 'package x\n\ntype S struct {\nA int\nBB string\n}\n' >"$scratch/a.go"
(
	cd "$scratch" || exit 1
	git init -q .
	git add a.go
	git -c user.email=selftest@local -c user.name=selftest commit -qm dirty-base
)
run_ws_case() { # name expect staged-content
	local name="$1" expect="$2" content="$3" got=0
	printf '%s' "$content" >"$scratch/a.go"
	(cd "$scratch" && unset GIT_INDEX_FILE && git add a.go && \
		bash "$repo/scripts/whitespace-drift.sh") \
		>/dev/null 2>&1 || got=$?
	if [ "$got" -ne "$expect" ]; then
		fail "$name: expected exit $expect, got $got"
	fi
}
if command -v nix >/dev/null 2>&1 && \
	(cd "$repo" && nix eval --raw .#formatter.x86_64-linux >/dev/null 2>&1); then
	# Staged = gofmt-canonical alignment of the dirty base (formatter
	# output, whitespace-only delta) => acquit. Canonical bytes verified
	# against the config's own gofmt: tab-indented, name-aligned fields.
	run_ws_case "formatter-produced whitespace delta is allowed" 0 \
		$'package x\n\ntype S struct {\n\tA  int\n\tBB string\n}\n'
	# Staged = a hand misalignment the formatter would rewrite => refuse.
	run_ws_case "hand whitespace drift is still refused" 1 \
		$'package x\n\ntype S struct {\n\tA   int\n\tBB string\n}\n'
	ws_cases=2
else
	echo "pre-commit-guard-selftest: WARN — nix/treefmt unavailable, whitespace adjudication cases skipped" >&2
	ws_cases=0
fi

echo "pre-commit-guard-selftest: $((6 + ws_cases)) passed"
