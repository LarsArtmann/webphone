#!/usr/bin/env bash
# fleet-tags.sh — per-submodule tag verification across the library fleet.
#
# The 2026-11-13 near-miss class: `awk -F/ '{print $NF}'` over `git
# ls-remote --tags` output mangled multi-module repos (last-component
# sorting made usermgmt tags look like root-module releases) and nearly
# wrote four false "behind latest" claims into the audit corpus. The
# ONLY honest tag query is a PER-SUBMODULE refspec:
#   git ls-remote origin "refs/tags/<module>/v<version>"
#
# The untagged-release sweep (the dashboard v0.13.0 class: the release
# commit pushed, the tag missing) keys on CHANGELOG.md files: the newest
# `## [vX.Y.Z]` heading of every CHANGELOG in a repo must have a
# matching remote tag — root module tag `vX.Y.Z`, nested module tag
# `<module-dir>/vX.Y.Z` (fleet convention: module dir == module path).
#
# Usage:
#   scripts/fleet-tags.sh [--repos dir...]      sweep; findings to stdout
#   scripts/fleet-tags.sh --self-test           fixture-verified behavior
#   scripts/fleet-tags.sh --offline --repos ... local tags instead of remote
set -euo pipefail

repo_latest_version() {
	# The newest version heading of one CHANGELOG: `## [v1.2.3]` (date or
	# link suffix tolerated). Empty when the file carries no version section.
	sed -n 's/^## \[\(v[0-9][0-9.]*[^]]*\)\].*/\1/p' "$1" 2>/dev/null | head -1 | sed 's/].*//; s/^v//' | sed 's/^/v/' | head -1 || true
}

sweep_repo() {
	repo="$1"
	offline="$2"
	find "$repo" -name CHANGELOG.md -type f -not -path "*/node_modules/*" -not -path "*/.pnpm/*" 2>/dev/null | while read -r changelog; do
		version="$(repo_latest_version "$changelog")"
		[ -n "$version" ] || continue
		dir="$(dirname "$changelog")"
		if [ "$dir" = "$repo" ]; then
			module=""
			tag="$version"
		else
			module="${dir#"$repo"/}"
			tag="$module/$version"
		fi
		if [ "$offline" = "1" ]; then
			if git -C "$repo" rev-parse -q --verify "refs/tags/$tag" >/dev/null 2>&1; then
				continue
			fi
		else
			if git -C "$repo" ls-remote origin "refs/tags/$tag" 2>/dev/null | grep -q .; then
				continue
			fi
		fi
		printf 'UNTAGGED\t%s\t%s\t%s\n' "$repo" "${module:-<root>}" "$tag"
	done
}

self_test() {
	tmp="$(mktemp -d)"
	trap 'rm -rf "$tmp"' EXIT
	# Fixture repo: a tagged root release, an UNTAGGED root release, and
	# a nested module whose tags carry the module prefix.
	git init -q "$tmp/fleet"
	git -C "$tmp/fleet" config user.email t@t
	git -C "$tmp/fleet" config user.name t
	printf '## [v1.2.3] - 2026-01-01\nold\n' >"$tmp/fleet/CHANGELOG.md"
	git -C "$tmp/fleet" add CHANGELOG.md
	git -C "$tmp/fleet" commit -q -m "root v1.2.3"
	# -c overrides: the host config can force annotated/signed tags (an
	# editor opens and 'false' dies); the fixture only needs a resolvable
	# refs/tags/<tag>, which a lightweight tag provides.
	git -C "$tmp/fleet" -c tag.forceSignAnnotated=false -c tag.gpgsign=false tag "v1.2.3"
	printf '## [v9.9.9] - 2026-02-02\nnew\n' >"$tmp/fleet/CHANGELOG.md"
	git -C "$tmp/fleet" commit -q -a -m "root v9.9.9 untagged"
	mkdir -p "$tmp/fleet/usermgmt"
	printf '## [v4.20.0] - 2026-03-03\nnested\n' >"$tmp/fleet/usermgmt/CHANGELOG.md"
	git -C "$tmp/fleet" add usermgmt/CHANGELOG.md
	git -C "$tmp/fleet" commit -q -m "nested v4.20.0 untagged"

	findings="$(sweep_repo "$tmp/fleet" 1 | sort)"
	expected="$(printf 'UNTAGGED\t%s\t<root>\tv9.9.9\nUNTAGGED\t%s\tusermgmt\tusermgmt/v4.20.0\n' "$tmp/fleet" "$tmp/fleet")"
	if [ "$findings" != "$expected" ]; then
		echo "self-test FAILED: findings mismatch" >&2
		echo "--- got ---" >&2
		printf '%s\n' "$findings" >&2
		echo "--- want ---" >&2
		printf '%s\n' "$expected" >&2
		return 1
	fi
	# The tagged root release must NOT be a finding (the mangled sweep's
	# false-positive class).
	if printf '%s\n' "$findings" | grep -q "v1.2.3"; then
		echo "self-test FAILED: the tagged v1.2.3 was flagged" >&2
		return 1
	fi
	echo "self-test: 3 checks green (untagged root, untagged nested, tagged-not-flagged)"
}

case "${1:-}" in
--self-test)
	self_test
	exit $?
	;;
esac

offline=0
if [ "${1:-}" = "--offline" ]; then
	offline=1
	shift
fi
if [ "${1:-}" != "--repos" ]; then
	echo "usage: $0 [--offline] --repos dir... | --self-test" >&2
	exit 2
fi
shift

status=0
for repo in "$@"; do
	[ -d "$repo/.git" ] || [ -f "$repo/.git" ] || {
		echo "SKIP (not a git repo): $repo" >&2
		continue
	}
	sweep_repo "$repo" "$offline" || status=1
done
exit $status
