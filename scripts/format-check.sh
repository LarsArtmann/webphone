#!/usr/bin/env bash
# THE format gate: builds the flake's own format check, the single
# authority for "is the tree treefmt-clean".
#
# `nix fmt` is NOT evidence: treefmt v2 caches, and its "formatted 0
# files" can be a 37ms cache short-circuit that skips every formatter
# while the sandboxed check (cold cache, --ci) still reds. The
# 2026-10-29 class. This wrapper always builds the check derivation —
# same bytes CI sees — and prints which one it ran.
#
# The check evaluates the COMMITTED (tracked) tree: it proves what CI
# will see, not your uncommitted scratch. Commit first, then run this.
set -euo pipefail
echo "format-check: building .#checks.x86_64-linux.format (the ONLY authority; nix fmt output is not evidence)" >&2
exec nix build --no-link --print-out-paths ".#checks.x86_64-linux.format" "$@"
