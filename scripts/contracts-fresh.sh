#!/usr/bin/env bash
# Freshness check for the committed wire-contract tree: regenerate
# docs/contracts/asyncapi.yaml AND the consumer fixtures
# (docs/contracts/wire/*.json) from the source of truth (the handler
# wire types and push payload builders, via `webphone contracts`) and
# diff against the committed files. Exit 1 on drift — the committed
# contract rotted against the code. The flake check `contracts-fresh`
# (nix/checks.nix) runs the same recipe against the built package in
# the sandbox.
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

nix develop -c go run ./cmd/webphone contracts -o "$tmp" >/dev/null

if ! diff -u docs/contracts/asyncapi.yaml "$tmp/asyncapi.yaml"; then
	echo "contracts-fresh: docs/contracts/asyncapi.yaml is STALE — regenerate with:" >&2
	echo "  nix develop -c go run ./cmd/webphone contracts" >&2
	exit 1
fi
if ! diff -ru docs/contracts/wire "$tmp/wire"; then
	echo "contracts-fresh: docs/contracts/wire/ is STALE — regenerate with:" >&2
	echo "  nix develop -c go run ./cmd/webphone contracts" >&2
	echo "  (then re-vendor into webphone-android: its scripts/sync-wire-fixtures.sh)" >&2
	exit 1
fi
echo "contracts-fresh: asyncapi.yaml + wire/ fixtures fresh"
