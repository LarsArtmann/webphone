#!/usr/bin/env bash
# scripts/lsp.sh — start a Go language server from the PREBUILT flake output.
#
#   scripts/lsp.sh gopls serve
#   scripts/lsp.sh golangci-lint-langserver
#
# Why this exists (LSP outage 2026-11-21..27): the project crushrc started
# both servers with `nix develop -c <server>`. That re-fetches and
# re-evaluates the flake on EVERY start — so two servers x every open
# session contended for the same flake fetch lock — and the devShell's
# shellHook wrote its banner to stdout, which is the LSP wire: the client
# read "webphone dev shell — go1.27.x" where the JSON-RPC initialize
# result belonged and closed the stream. Every start, every session.
#
# So the servers come from ONE prebuilt store output (flake output
# `go-lsp`: go, gopls, golangci-lint-langserver, templ). This wrapper
# resolves it by the hash of the files that define it and execs the binary
# directly: a warm start runs no nix at all, takes no lock, and owns a
# clean stdout. Only a cold start (first run, or after flake.lock /
# devshell.nix / packages.nix changed, or after the store path was
# collected) builds the output once.
#
# Same toolchain policy as the devShell (GOTOOLCHAIN=local, the pinned Go
# on PATH): the host go is older than go.mod's floor.
set -euo pipefail

TOOL=${1:-}
if (($# > 0)); then
	shift
fi
case "$TOOL" in
gopls | golangci-lint-langserver) ;;
*)
	echo "lsp.sh: usage: scripts/lsp.sh gopls|golangci-lint-langserver [args]" >&2
	exit 2
	;;
esac

ROOT=$(cd "$(dirname "$0")/.." && pwd)

KEY=$(cat "$ROOT/flake.lock" "$ROOT/nix/devshell.nix" "$ROOT/nix/packages.nix" |
	sha256sum | cut -c1-16)
LINK_DIR=${XDG_CACHE_HOME:-$HOME/.cache}/webphone/go-lsp
LINK="$LINK_DIR/$KEY"

if [[ ! -x "$LINK/bin/$TOOL" ]]; then
	echo "lsp.sh: building the Go LSP toolset (cold start for flake key $KEY)..." >&2
	if ! OUT=$(nix build --no-link --print-out-paths "$ROOT#go-lsp"); then
		echo "lsp.sh: FATAL: nix build .#go-lsp failed — the LSP cannot start (see nix error above)" >&2
		exit 1
	fi
	mkdir -p "$LINK_DIR"
	# Publish atomically: concurrent cold starts (both servers at once) must
	# never observe a half-written link.
	ln -sfn "$OUT" "$LINK.tmp.$$"
	mv -T "$LINK.tmp.$$" "$LINK"
fi

exec env PATH="$LINK/bin:$PATH" GOTOOLCHAIN=local "$LINK/bin/$TOOL" "$@"
