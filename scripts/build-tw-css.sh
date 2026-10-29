#!/usr/bin/env bash
# Rebuild the scoped Tailwind v4 artifacts under internal/web/assets/:
#
#   tw.css    — the adopted templ-components surfaces for the shell
#               (display.Button, forms.Input/Textarea, display.EmptyState,
#               layout.Base skip link, icons)
#   login.css — the cqrs-htmx loginpage surface at /login (the library
#               page + the feedback.Alert/Spinner components it renders),
#               with the app.css design tokens INLINED (the page links
#               exactly ONE stylesheet, so the @theme remap's var()
#               references must resolve)
#
# NEVER merges with health.css (the scoped dashboard build) and never
# builds with tailwindcss v3.
#
# The recipes (dark custom variant + @theme token remap + @source scan
# sets) live in the committed input files internal/web/assets/*.input, so
# this script is just the pinned invocation. The toolchain rides the FLAKE
# LOCKED nixpkgs rev (not the registry default) — same contract as
# build-health-css.sh.
#
# Module sources resolve via `go list -m -f {{.Dir}}` (run inside
# `nix develop`): the untracked vendor/ tree this script once copied from
# was trashed 2026-10-17 — the module cache is the one source of truth,
# and the staged tree keeps the historical vendor/-shaped layout so the
# committed @source paths stay byte-stable.
#
# STAGED build (mirrors build-health-css.sh): Tailwind's implicit content
# detection leaks the surrounding repo/git context, and the artifacts sit
# inside a @source root, so an in-place -o would feed the previous build's
# own classes back into the next. Only the input + the scanned sources go
# into a temp tree that replicates the repo-relative layout.
set -euo pipefail
REPO="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO"

rev="$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)"

# Radius drift gate: the inputs' radius literals must equal the app.css
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

# @theme parity gate: the login input's NON-GRAY token remap must be
# byte-equal to tw.css.input's (one home for the brand remap; the login
# input only ADDS sources and the .dark-extended variant). The GRAY
# family is deliberately ABSENT from the login input — the loginpage's
# standard dark idiom needs Tailwind's stock palette (the shell's
# gray-to-token remap inverts the dark page: gray-900=--text makes the
# dark page background light; caught by the T23 dark shot, 2026-10-25) —
# so the gate compares only the non-gray token declarations.
if ! diff <(sed -n '/@theme inline {/,/^}/p' internal/web/assets/tw.css.input | grep -E '^\s*--' | grep -v -- '--color-gray-') \
	<(sed -n '/@theme inline {/,/^}/p' internal/web/assets/login.css.input | grep -E '^\s*--' | grep -v -- '--color-gray-') >/dev/null; then
	echo "login.css.input @theme brand slots diverged from tw.css.input — copy the non-gray remap back over (one home per remap)" >&2
	exit 1
fi
if grep -q -- '--color-gray-' internal/web/assets/login.css.input; then
	echo "login.css.input remaps the gray family — the loginpage's dark idiom needs Tailwind's stock palette; see the input's header comment" >&2
	exit 1
fi

COMP="vendor/github.com/larsartmann/templ-components"
LOGIN="vendor/github.com/larsartmann/cqrs-htmx/loginpage/v4"

# Resolve the pinned module sources from the module cache (requires the
# dev shell's go). A COLD cache (fresh CI runner, new machine) answers
# an EMPTY Dir — heal it with an explicit download instead of cp-ing
# from "/"; a module that still resolves empty means go.mod and the
# inputs disagree. icons is its OWN module (templ-components/icons),
# not a subdir.
moddir() {
	local dir
	dir="$(go list -m -f "{{.Dir}}" "$1" 2>/dev/null || true)"
	if [ -z "$dir" ]; then
		go mod download "$1" >/dev/null
		dir="$(go list -m -f "{{.Dir}}" "$1" 2>/dev/null || true)"
	fi
	if [ -z "$dir" ]; then
		echo "build-tw-css: module $1 resolved no source dir — is it in go.mod?" >&2
		exit 1
	fi
	printf '%s' "$dir"
}
TC_DIR="$(moddir github.com/larsartmann/templ-components)"
ICONS_DIR="$(moddir github.com/larsartmann/templ-components/icons)"
LP_DIR="$(moddir github.com/larsartmann/cqrs-htmx/loginpage/v4)"

stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/internal/web/assets" "$stage/$COMP/display" \
	"$stage/$COMP/forms" "$stage/$COMP/layout" "$stage/$COMP/icons" \
	"$stage/$COMP/feedback" "$stage/$LOGIN/assets"
cp internal/web/assets/tw.css.input internal/web/assets/login.css.input "$stage/internal/web/assets/"
cp internal/web/assets/app.css "$stage/internal/web/assets/"
cp -r internal/web/views "$stage/internal/web/views"
for f in \
	display/definition_list.templ display/definition_list_templ.go \
	display/empty_state.templ display/empty_state_templ.go \
	display/button_go.go display/button.templ display/button_templ.go \
	forms/input.templ forms/input_templ.go forms/input_classes.go \
	forms/label.templ forms/label_templ.go \
	forms/textarea.templ forms/textarea_templ.go \
	layout/base.templ layout/base_templ.go \
	feedback/alert.templ feedback/alert_templ.go \
	feedback/loading.templ feedback/loading_templ.go; do
	cp "$TC_DIR/$f" "$stage/$COMP/$f"
done
for f in icon.templ icon_templ.go; do
	cp "$ICONS_DIR/$f" "$stage/$COMP/icons/$f"
done
tw() { # input output
	nix run "github:NixOS/nixpkgs/$rev#tailwindcss_4" -- \
		-i "$stage/internal/web/assets/$1" \
		-o "$stage/$2"
}

# ORDER MATTERS: the shell build runs BEFORE the loginpage sources land
# in the stage, AND the shell build auto-scans the staged assets dir —
# so the login input and its comments must never mention dark-variant
# utility names (a literal utility in a comment EMITS it into tw.css;
# the 2026-10-25 leak). Belt and suspenders: neither stage order nor
# comment prose can leak the other build's classes.
tw tw.css.input tw.css

for f in page.templ page_templ.go assets/login.js; do
	cp "$LP_DIR/$f" "$stage/$LOGIN/$f"
done
tw login.css.input login.css

mv "$stage/tw.css" internal/web/assets/tw.css
mv "$stage/login.css" internal/web/assets/login.css

# Gates: the manual-dark fingerprint, the remapped component utilities, and
# the token-driven values must all be present in the shell artifact.
grep -q 'data-theme="dark"' internal/web/assets/tw.css
grep -qF '.bg-blue-600' internal/web/assets/tw.css
grep -qF '.ring-gray-300' internal/web/assets/tw.css
grep -qF 'var(--accent)' internal/web/assets/tw.css

# Gates: the login artifact carries the app tokens (the remap resolves
# through them), the page-shell utilities the library markup needs, the
# token-driven values, and the .dark variant the page's ThemeScript class
# activates.
grep -q -- '--accent:' internal/web/assets/login.css
grep -qF '.min-h-dvh' internal/web/assets/login.css
grep -qF '.bg-blue-600' internal/web/assets/login.css
grep -qF 'var(--accent)' internal/web/assets/login.css
grep -qF '.hidden' internal/web/assets/login.css
grep -qF '.dark' internal/web/assets/login.css

# A raw rebuild re-breaks the treefmt format gate (prettier owns *.css);
# hand the fresh artifacts straight to the formatter. SCOPED to the
# artifacts on purpose: a repo-wide nix fmt would trample concurrent
# in-flight edits. TREEFMT_NO_CACHE=1 is LOAD-BEARING (2026-10-29): the
# v2 evaluation cache answered "0 changed in 7ms" on non-canonical
# artifact bytes, shipping a rebuild that reds the format check; the
# flag forces the formatter to actually run (276ms, collapses the file).
TREEFMT_NO_CACHE=1 nix fmt internal/web/assets/tw.css internal/web/assets/login.css
echo "rebuilt internal/web/assets/{tw,login}.css with locked nixpkgs $rev"
