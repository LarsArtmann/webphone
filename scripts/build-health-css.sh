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
# STAGED build: only the input + the three @source roots go into a temp
# tree. Two reasons: (1) Tailwind's implicit content detection leaks the
# surrounding repo/git context (a build over a HEAD-only snapshot even
# differs because vendor/ is untracked), so only a staged build is
# reproducible; (2) the artifact itself sits inside a @source root, so
# an in-place -o feeds the previous build's own classes back into the
# next (self-perpetuating zombie classes — emerald-100 survived three
# rebuilds that way). The committed artifact cannot be byte-verified by
# a flake check (the flake source never sees untracked vendor/); the
# checks.health-css CANARY pins the dark variant instead.
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/internal/web/assets" "$stage/internal/app" \
	"$stage/vendor/github.com/larsartmann/go-health-dashboard" \
	"$stage/vendor/github.com/larsartmann/templ-components"
cp internal/web/assets/health.css.input "$stage/internal/web/assets/"
cp internal/app/dashboard.go "$stage/internal/app/"
cp -r vendor/github.com/larsartmann/go-health-dashboard/. \
	"$stage/vendor/github.com/larsartmann/go-health-dashboard/"
cp -r vendor/github.com/larsartmann/templ-components/. \
	"$stage/vendor/github.com/larsartmann/templ-components/"
nix run "github:NixOS/nixpkgs/$rev#tailwindcss_4" -- \
	-i "$stage/internal/web/assets/health.css.input" \
	-o "$stage/health.css"
mv "$stage/health.css" internal/web/assets/health.css
# A raw rebuild re-breaks the treefmt format gate (prettier owns *.css —
# t21-t22 report f1); hand the fresh artifact straight to the formatter.
# SCOPED to the artifact on purpose: a repo-wide nix fmt would trample
# concurrent in-flight edits.
nix fmt internal/web/assets/health.css
echo "rebuilt internal/web/assets/health.css with locked nixpkgs $rev"
