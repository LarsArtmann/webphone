#!/usr/bin/env bash
# Rebuild internal/web/assets/health.css — the scoped Tailwind v4 build for
# the go-health-dashboard /health subtree (never merges with tw.css).
#
# The recipe (dark custom variant + @source scan roots) lives in the committed
# input file internal/web/assets/health.css.input, so this script is just the
# pinned invocation. tailwindcss v4 ONLY — v3 emits a different layer model
# and silently drops the dark variant.
#
# After a rebuild, review the diff and confirm the dark-selector set is
# preserved; the `checks.health-css` canary greps the output for `:where(.dark`
# and fails if the dark variant was lost.
set -euo pipefail
REPO="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO"

nix run nixpkgs#tailwindcss_4 -- \
	-i internal/web/assets/health.css.input \
	-o internal/web/assets/health.css
echo "rebuilt internal/web/assets/health.css"
