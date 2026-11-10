#!/usr/bin/env bash
# changelog-guard.sh — the CHANGELOG coverage bar, enforced.
#
# The bar is RATIFIED (AGENTS § Conventions): committed code-path changes get
# an `[Unreleased]` line; test-only and docs-only deltas stay changelog-silent.
# It had no guard, and the metaengine per-group-cap train (2026-11-06) landed
# with no line while every gate stayed green — a bar nobody can check is a
# preference, not a bar.
#
#   scripts/changelog-guard.sh [<base>..<head>]     default: origin/main..HEAD
#   scripts/changelog-guard.sh --list                only print what it counts
#
# Exit 0  = the range is exempt, or it changed product code AND CHANGELOG.md
# Exit 1  = product code changed without a CHANGELOG.md delta (quote the rule)
# Exit 0  = base ref unresolvable, with a WARN (fail-open: a shallow or
#           first-push checkout must not red the build for missing history)
#
# "Product code" is deliberately narrow — the bar exists for changes a USER or
# an OPERATOR can feel: Go sources, templ views, the served JS/CSS assets, and
# the packaging (flake/nix). Tooling scripts, tests, workflows and docs are
# exempt, because a guard that reds on every docs sweep gets routed around.
set -uo pipefail

RANGE=""
LIST_ONLY=0
for arg in "$@"; do
	case "$arg" in
	--list) LIST_ONLY=1 ;;
	--help | -h)
		sed -n '2,20p' "$0"
		exit 0
		;;
	*) RANGE="$arg" ;;
	esac
done

command -v git >/dev/null 2>&1 || {
	echo "changelog-guard: FATAL: git not on PATH" >&2
	exit 2
}
git rev-parse --git-dir >/dev/null 2>&1 || {
	echo "changelog-guard: FATAL: not inside a git checkout" >&2
	exit 2
}

[ -n "$RANGE" ] || RANGE="origin/main..HEAD"
base="${RANGE%%..*}"

if ! git rev-parse --verify --quiet "$base^{commit}" >/dev/null; then
	echo "changelog-guard: WARN: base ref '$base' does not resolve (shallow clone or first push) — guard is a NO-OP for $RANGE" >&2
	exit 0
fi

files=$(git diff --name-only "$RANGE" 2>/dev/null) || {
	echo "changelog-guard: WARN: could not diff '$RANGE' — guard is a NO-OP" >&2
	exit 0
}

# Path filter, one line per reason. Keep the shapes explicit: a glob that
# silently stops matching is the failure mode this whole script exists for.
is_product_code() {
	local f="$1"
	case "$f" in
	CHANGELOG.md) return 1 ;;
	# tests never count
	*_test.go) return 1 ;;
	internal/web/assets/island-tests/*) return 1 ;;
	internal/web/assets/vendor/*) return 1 ;;
	# docs never count
	*.md) return 1 ;;
	docs/*) return 1 ;;
	# workflows are tooling
	.github/*) return 1 ;;
	# shepherding scripts are tooling (they do not ship)
	scripts/*) return 1 ;;
	# product code: Go, templ views + their codegen, served assets, packaging
	*.go | *_templ.go) return 0 ;;
	*.templ) return 0 ;;
	internal/web/assets/*) return 0 ;;
	nix/*) return 0 ;;
	flake.nix | flake.lock) return 0 ;;
	go.mod | go.sum) return 0 ;;
	package/*) return 0 ;;
	cmd/*) return 0 ;;
	esac
	return 1
}

product=()
while IFS= read -r f; do
	[ -n "$f" ] || continue
	is_product_code "$f" && product+=("$f")
done <<<"$files"

changelog_changed=0
while IFS= read -r f; do
	[ "$f" = "CHANGELOG.md" ] && changelog_changed=1
done <<<"$files"

if [ "$LIST_ONLY" -eq 1 ]; then
	echo "changelog-guard: range $RANGE"
	echo "changelog-guard: CHANGELOG.md changed = $changelog_changed"
	echo "changelog-guard: product-code files (${#product[@]}):"
	printf '  %s\n' "${product[@]}"
	exit 0
fi

if [ "${#product[@]}" -eq 0 ]; then
	echo "changelog-guard: PASS — no product-code change in $RANGE (docs/tests/tooling only)"
	exit 0
fi

if [ "$changelog_changed" -eq 1 ]; then
	echo "changelog-guard: PASS — ${#product[@]} product-code file(s) changed and CHANGELOG.md carries the entry"
	exit 0
fi

{
	echo "changelog-guard: FAIL — product code changed but CHANGELOG.md did not ($RANGE)"
	echo
	echo "Changed product-code files (${#product[@]}):"
	printf '  %s\n' "${product[@]}"
	echo
	echo "The ratified bar (AGENTS § Conventions): committed code-path changes get an"
	echo "[Unreleased] line; test-only and docs-only deltas stay changelog-silent."
	echo
	echo "Fix: add one [Unreleased] line naming the user/operator-visible effect, e.g."
	echo "  - **What changed** — why it matters (file or option the reader can act on)"
	echo
	echo "If this range genuinely carries no user-visible change, say so in the"
	echo "commit message; the guard judges only whether CHANGELOG.md moved."
} >&2
exit 1
