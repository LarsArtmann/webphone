#!/usr/bin/env bash
# Rebuild internal/web/assets/health.css — the scoped Tailwind v4 build for
# the go-health-dashboard /health subtree (kept separate from tw.css; the two
# never merge).
#
# Why a script and not a static input: the class names live in the
# go-health-dashboard sources. The repo VENDORS its deps, so the stable path
# is vendor/github.com/larsartmann/go-health-dashboard (a module-cache path
# would be version-specific). The @source lines are appended to the committed
# input (internal/web/assets/health.css.input). tailwindcss v4 ONLY — v3 emits
# a different layer model and silently drops the dark variant.
#
# Run inside `nix develop` (for the go floor) or any shell with go + nix.
set -euo pipefail
REPO="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO"

DASH_DIR="$REPO/vendor/github.com/larsartmann/go-health-dashboard"
[ -d "$DASH_DIR" ] || {
	echo "vendored dashboard missing; run 'go mod vendor' first" >&2
	exit 1
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
{
	cat internal/web/assets/health.css.input
	printf '\n/* Appended by scripts/build-health-css.sh: dynamic @source roots. */\n'
	printf '@source "%s/*.templ";\n' "$DASH_DIR"
	printf '@source "%s/internal/app/dashboard.go";\n' "$REPO"
} > "$tmp/input.css"

nix run nixpkgs#tailwindcss_4 -- -i "$tmp/input.css" -o internal/web/assets/health.css
echo "rebuilt internal/web/assets/health.css"
