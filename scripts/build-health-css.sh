#!/usr/bin/env bash
# Rebuild internal/web/assets/health.css — the scoped Tailwind v4 build for
# the go-health-dashboard /health subtree (never merges with tw.css).
#
# The recipe (dark custom variant + @source scan roots) lives in the committed
# input file internal/web/assets/health.css.input, so this script is just the
# pinned invocation. tailwindcss v4 ONLY — v3 emits a different layer model
# and silently drops the dark variant.
#
# The toolchain rides the FLAKE LOCKED nixpkgs rev (not the registry default):
# the `checks.health-css` drift gate rebuilds with the locked tailwindcss_4,
# so a registry-version rebuild would mint an artifact the gate rejects the
# moment the registry moves. After a deliberate nixpkgs bump, re-run this
# script and commit the diff.
set -euo pipefail
REPO="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO"

rev="$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)"
# Build OUT of tree, then move into place: the artifact sits inside a
# Tailwind @source scan root, so an in-place -o would feed the previous
# build's own class names back into the next (self-perpetuating zombie
# classes — emerald-100 survived three rebuilds that way). The drift
# check builds outside the tree too; both sides now see the clean scan.
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
nix run "github:NixOS/nixpkgs/$rev#tailwindcss_4" -- \
	-i internal/web/assets/health.css.input \
	-o "$tmp"
mv "$tmp" internal/web/assets/health.css
echo "rebuilt internal/web/assets/health.css with locked nixpkgs $rev"
