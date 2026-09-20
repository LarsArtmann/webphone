#!/usr/bin/env bash
# Release runbook for webphone — the dance that cut v2.0.0/v2.1.0, as one command.
#
# Usage:
#   scripts/release.sh X.Y.Z          # real release (gates run, tag pushes)
#   scripts/release.sh X.Y.Z --dry-run
#
# Preconditions you must do BEFORE calling this (runbook step 1):
#   - CHANGELOG: fold [Unreleased] into a dated "## [X.Y.Z] - YYYY-MM-DD" section
#   - FEATURES/TODO_LIST/ROADMAP synced, docs committed
#
# What this script does (runbook steps 2-8, fail-fast):
#   clean tree + main + synced with origin, tag free, CHANGELOG section exists
#   -> bump flake.nix webphoneVersion (package version AND /version ldflags)
#   -> gates: buildflow (no result cache), go test, nix flake check, smoke,
#      vulnix (runtime-closure advisory scan, network)
#   -> annotated tag, push main + tag, ls-remote verify
#   -> lychee link check (after push so tag links resolve)
#   -> stack: relock webphone input, commit, browser E2E + webphone VM test
#      + nix flake check, push
#   -> aarch64: cross-build package + island-lint check
#   -> gh release with the CHANGELOG section as notes
set -euo pipefail

VERSION="${1:-}"
DRY_RUN=0
case " ${*:2} " in *" --dry-run "*) DRY_RUN=1 ;; esac
if [ -z "$VERSION" ]; then
	echo "usage: $0 X.Y.Z [--dry-run]" >&2
	exit 2
fi
VERSION="${VERSION#v}"
TAG="v$VERSION"
REPO="$(cd "$(dirname "$0")/.." && pwd)"
STACK="${WEBPHONE_STACK_DIR:-$HOME/projects/nix-international-telephony}"

step() { printf '\n== %s ==\n' "$*"; }
run() {
	if [ "$DRY_RUN" = "1" ]; then echo "  [dry-run] $*"; else "$@"; fi
}

cd "$REPO"

step "1/9 preconditions"
[ -z "$(git status --porcelain)" ] || {
	echo "tree not clean" >&2
	exit 1
}
[ "$(git branch --show-current)" = "main" ] || {
	echo "not on main" >&2
	exit 1
}
git fetch origin --tags --quiet
[ "$(git rev-parse main)" = "$(git rev-parse origin/main)" ] || {
	echo "main diverged from origin/main" >&2
	exit 1
}
RESUME_TAG=0
if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null; then
	if git ls-remote --tags origin "refs/tags/$TAG" | grep -q .; then
		echo "tag $TAG already on origin: resuming after step 5 (stack, aarch64, gh release)"
		RESUME_TAG=1
	else
		echo "tag $TAG exists locally but not on origin; delete it first (git tag -d $TAG)" >&2
		exit 1
	fi
fi
echo "ok: clean main at $(git rev-parse --short HEAD), $TAG is free"

step "2/9 fold-check: CHANGELOG has a dated $VERSION section"
grep -q "^## \[$VERSION\] " CHANGELOG.md || {
	echo "CHANGELOG has no '## [$VERSION] - date' section; fold [Unreleased] first (runbook step 1)" >&2
	exit 1
}
echo "ok"

step "3/9 version bump: flake.nix webphoneVersion = $VERSION"
if grep -q "webphoneVersion = \"$VERSION\";" flake.nix; then
	echo "already $VERSION"
else
	if [ "$DRY_RUN" = "1" ]; then
		echo "  [dry-run] sed flake.nix webphoneVersion -> $VERSION + commit 'release: $TAG'"
	else
		sed -i "s/webphoneVersion = \"[^\"]*\";/webphoneVersion = \"$VERSION\";/" flake.nix
		git add flake.nix
		git commit -m "release: $TAG"
	fi
fi

step "4/9 gates (fail-fast)"
run env BUILDFLOW_NO_RESULT_CACHE=1 buildflow
run env GOEXPERIMENT=jsonv2 go test -count=1 ./...
run nix flake check
run python3 scripts/webphone-smoke.py
# Vulnix cadence (TODO row closed 2026-09-20): every train scans the
# RUNTIME closure. Range-matched findings against patched nixpkgs versions
# (the glibc class) trip this — triage against the locked tree's patch level
# before shipping, as documented in AGENTS.md.
run nix run .#vulnix

step "5/9 tag + push + verify"
if [ "$RESUME_TAG" = "1" ]; then
	echo "skipped: $TAG already pushed"
else
	run git tag -a "$TAG" -m "webphone $TAG"
	run git push origin main "$TAG"
	if [ "$DRY_RUN" != "1" ]; then
		git ls-remote --tags origin "refs/tags/$TAG^{}" | grep -q . || {
			echo "tag $TAG not resolvable on remote (peeled object missing)" >&2
			exit 1
		}
		echo "remote has $TAG"
	fi
fi

step "6/9 link check (after push so the new tag link resolves)"
run nix run nixpkgs#lychee -- .

step "7/9 stack: relock, gates, push ($STACK)"
[ -d "$STACK" ] || {
	echo "stack dir missing: $STACK" >&2
	exit 1
}
[ -z "$(git -C "$STACK" status --porcelain)" ] || {
	echo "stack tree not clean" >&2
	exit 1
}
run bash -c "cd '$STACK' && nix flake lock --update-input webphone"
if [ "$DRY_RUN" != "1" ]; then
	git -C "$STACK" add flake.lock
	git -C "$STACK" diff --cached --quiet || git -C "$STACK" commit -m "chore: bump webphone input to $TAG"
fi
run bash -c "cd '$STACK' && nix build -L .#telephony-browser"
run bash -c "cd '$STACK' && nix build -L .#checks.x86_64-linux.telephony-webphone"
run bash -c "cd '$STACK' && nix flake check"
run bash -c "cd '$STACK' && git push"

step "8/9 aarch64 cross-builds"
run nix build .#webphone --system aarch64-linux
run nix build .#checks.aarch64-linux.island-lint

step "9/9 GitHub release"
if [ "$DRY_RUN" = "1" ]; then
	echo "  [dry-run] gh release create $TAG --verify-tag --title $TAG --notes <CHANGELOG $VERSION section>"
else
	notes="$(mktemp)"
	awk -v want="## [$VERSION]" '
    $0 ~ "^" want {on=1; next}
    /^## / && on {exit}
    on {print}
  ' CHANGELOG.md >"$notes"
	gh release create "$TAG" --verify-tag --title "$TAG" --notes-file "$notes"
	rm -f "$notes"
fi

printf '\nrelease %s complete. Remaining manual steps: owner deploy (nixos-rebuild),\npost-deploy probes (smoke --base), pbx-artmann relock + toplevel pre-build.\n' "$TAG"
