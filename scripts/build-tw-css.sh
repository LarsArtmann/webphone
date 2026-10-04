#!/usr/bin/env bash
# Rebuild internal/web/assets/tw.css — the Tailwind v4 build for the adopted
# templ-components surfaces (display.Button, forms.Input/Textarea,
# display.EmptyState, layout.Base skip link, icons). NEVER merges with
# health.css (the scoped dashboard build) and never builds with tailwindcss
# v3.
#
# The recipe (dark custom variant + @theme token remap + @source scan set)
# lives in the committed input file internal/web/assets/tw.css.input, so this
# script is just the pinned invocation. The toolchain rides the FLAKE LOCKED
# nixpkgs rev (not the registry default) — same contract as
# build-health-css.sh.
#
# STAGED build (mirrors build-health-css.sh): Tailwind's implicit content
# detection leaks the surrounding repo/git context, and the artifact sits
# inside a @source root, so an in-place -o would feed the previous build's
# own classes back into the next. Only the input + the scanned sources go
# into a temp tree that replicates the repo-relative layout.
set -euo pipefail
REPO="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO"

rev="$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)"

# Radius drift gate: the input's radius literals must equal the app.css
# tokens (7/10/14 "Graphite desk" radii). They are duplicated only because
# Tailwind's own token names --radius-md/--radius-lg shadow the app names
# in the cascade; a silent mismatch would square the adopted buttons.
radius_pair() { # tailwind-token app-token
	local inv appv
	inv="$(grep -m1 -- "--$1: " internal/web/assets/tw.css.input | grep -oE '[0-9.]+px' || true)"
	appv="$(grep -m1 -- "  --$2: " internal/web/assets/app.css | grep -oE '[0-9.]+px' || true)"
	if [ -z "$inv" ] || [ -z "$appv" ] || [ "$inv" != "$appv" ]; then
		echo "tw.css radius drift: input --$1=$inv vs app.css --$2=$appv — update the @theme radii" >&2
		exit 1
	fi
}
radius_pair radius-md radius
radius_pair radius-lg radius-lg

COMP="vendor/github.com/larsartmann/templ-components"
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/internal/web/assets" "$stage/$COMP/display" \
	"$stage/$COMP/forms" "$stage/$COMP/layout" "$stage/$COMP/icons"
cp internal/web/assets/tw.css.input "$stage/internal/web/assets/"
cp -r internal/web/views "$stage/internal/web/views"
for f in \
	display/empty_state.templ display/empty_state_templ.go \
	display/button_go.go display/button.templ display/button_templ.go \
	forms/input.templ forms/input_templ.go forms/input_classes.go \
	forms/label.templ forms/label_templ.go \
	forms/textarea.templ forms/textarea_templ.go \
	layout/base.templ layout/base_templ.go \
	icons/icon.templ icons/icon_templ.go; do
	cp "$COMP/$f" "$stage/$COMP/$f"
done

nix run "github:NixOS/nixpkgs/$rev#tailwindcss_4" -- \
	-i "$stage/internal/web/assets/tw.css.input" \
	-o "$stage/tw.css"
mv "$stage/tw.css" internal/web/assets/tw.css

# Gates: the manual-dark fingerprint, the remapped component utilities, and
# the token-driven values must all be present in the artifact.
grep -q 'data-theme="dark"' internal/web/assets/tw.css
grep -qF '.bg-blue-600' internal/web/assets/tw.css
grep -qF '.ring-gray-300' internal/web/assets/tw.css
grep -qF 'var(--accent)' internal/web/assets/tw.css

# A raw rebuild re-breaks the treefmt format gate (prettier owns *.css);
# hand the fresh artifact straight to the formatter. SCOPED to the artifact
# on purpose: a repo-wide nix fmt would trample concurrent in-flight edits.
nix fmt internal/web/assets/tw.css
echo "rebuilt internal/web/assets/tw.css with locked nixpkgs $rev"
