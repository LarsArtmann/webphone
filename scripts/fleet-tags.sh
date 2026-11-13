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
# matching remote tag. Go resolves exactly ONE tag form per module:
# <module-path-minus-major-suffix>/vX.Y.Z — for catalog/v4 that is
# catalog/v4.15.0, NOT catalog/v4/v4.15.0 (the 2026-11-13 incident: the
# release script cut the suffixed shape, the proxy could not see the
# release, and the old two-form heuristic here read it as tagged).
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
		label="<root>"
		tags=("$version")
		if [ "$dir" != "$repo" ]; then
			# Nested module: Go strips the module path's /vN major suffix to
			# get the tag prefix, so the ONE resolvable form is the dir path
			# minus that suffix + /vX.Y.Z (catalog/v4 -> catalog/v4.15.0). A
			# suffixed tag (catalog/v4/v4.15.0) is INERT to Go — its presence
			# must not mask an untagged release.
			prefix="${dir#"$repo"/}"
			major="$(sed -E -n 's/^module .*(\/v[0-9]+)$/\1/p' "$dir/go.mod" 2>/dev/null | head -1)"
			label="$prefix"
			if [ -n "$major" ]; then
				prefix="${prefix%"$major"}"
			fi
			tags=("$prefix/$version")
		fi
		found=0
		for tag in "${tags[@]}"; do
			if [ "$offline" = "1" ]; then
				if git -C "$repo" rev-parse -q --verify "refs/tags/$tag" >/dev/null 2>&1; then
					found=1
					break
				fi
			else
				if git -C "$repo" ls-remote origin "refs/tags/$tag" 2>/dev/null | grep -q .; then
					found=1
					break
				fi
			fi
			display_tag="$tag"
		done
		if [ "$found" = "1" ]; then
			continue
		fi
		printf 'UNTAGGED\t%s\t%s\t%s\n' "$repo" "$label" "${display_tag:-$version}"
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
	# Major-suffix modules: Go resolves catalog/v4.15.0 (dir minus the
	# /v4 suffix) and NEVER catalog/v4/v4.15.0. A module carrying ONLY the
	# suffixed shape must be FLAGGED (the 2026-11-13 catalog incident);
	# one carrying the resolvable shape must not be.
	mkdir -p "$tmp/fleet/catalog/v4" "$tmp/fleet/usermgmt/v4"
	printf 'module example.com/fleet/catalog/v4\n\ngo 1.27\n' >"$tmp/fleet/catalog/v4/go.mod"
	printf '## [v4.15.0] - 2026-04-04\nsuffix module\n' >"$tmp/fleet/catalog/v4/CHANGELOG.md"
	printf 'module example.com/fleet/usermgmt/v4\n\ngo 1.27\n' >"$tmp/fleet/usermgmt/v4/go.mod"
	printf '## [v4.21.0] - 2026-04-04\nresolvable-form sibling\n' >"$tmp/fleet/usermgmt/v4/CHANGELOG.md"
	git -C "$tmp/fleet" add catalog/v4 usermgmt/v4
	git -C "$tmp/fleet" commit -q -m "wrong-shape-tagged and correctly-tagged modules"
	git -C "$tmp/fleet" -c tag.forceSignAnnotated=false -c tag.gpgsign=false tag "catalog/v4/v4.15.0"
	git -C "$tmp/fleet" -c tag.forceSignAnnotated=false -c tag.gpgsign=false tag "usermgmt/v4.21.0"

	findings="$(sweep_repo "$tmp/fleet" 1 | LC_ALL=C sort)"
	expected="$(printf 'UNTAGGED\t%s\t<root>\tv9.9.9\nUNTAGGED\t%s\tcatalog/v4\tcatalog/v4.15.0\nUNTAGGED\t%s\tusermgmt\tusermgmt/v4.20.0\n' "$tmp/fleet" "$tmp/fleet" "$tmp/fleet")"
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
	echo "self-test: 5 checks green (untagged root, untagged nested, wrong-shape-tagged FLAGGED, resolvable-tagged-not-flagged, tagged-root-not-flagged)"
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
