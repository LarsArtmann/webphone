#!/usr/bin/env bash
# pre-push-gates.sh — the mechanical ride-along battery, ONE command.
#
# Three lesson repeaters were caught only by CI after push: the
# errorfamily-registry freshness (TestErrorCodeRegistryIsFresh — a new
# code without its registry row reds the arch test), the art-dupl
# new-clone gate (a line-moving refactor re-keys a ruled twin), and the
# docs/artifact freshness class the 2026-11-10 rule-F glue and stale
# tw.css landed. This battery runs every push-critical gate LOCALLY so
# the surprise happens before the push, not after:
#
#   1. errorfamily registry freshness (the arch test)
#   2. art-dupl check (pinned v0.7.2, new-only vs the baseline)
#   3. changelog-guard (code-path deltas carry an [Unreleased] line)
#   4. templ-generate --check (views match the pinned compiler)
#   5. build-tw-css --check  (tw.css/login.css are fresh rebuilds)
#   6. docs-gates --check    (the docs battery, non-mutating mode)
#
# Skippable per gate with --only a,b,c / --skip a,b,c. Exit 1 on the
# first red (fail fast — a red gate names its own fix).
set -uo pipefail

cd "$(git rev-parse --show-toplevel)"

ONLY=""
SKIP=""
for arg in "$@"; do
	case "$arg" in
	--only=*) ONLY="${arg#--only=}" ;;
	--skip=*) SKIP="${arg#--skip=}" ;;
	*)
		echo "usage: pre-push-gates.sh [--only=g1,g2] [--skip=g1,g2]" >&2
		exit 2
		;;
	esac
done

want() {
	local gate="$1"
	if [ -n "$ONLY" ]; then
		[[ ",$ONLY," == *",$gate,"* ]]
		return
	fi
	if [ -n "$SKIP" ]; then
		[[ ",$SKIP," != *",$gate,"* ]]
		return
	fi
	return 0
}

fail=0
run_gate() {
	local gate="$1"
	shift
	if ! want "$gate"; then
		echo "== $gate: skipped"
		return
	fi
	echo "== $gate"
	if ! "$@"; then
		echo "RED: $gate failed — fix before push (the gate's output names the fix)" >&2
		fail=1
		return 1
	fi
}

# env -u GOEXPERIMENT: the host-env leak flips encoding/json to v2 locally
# (AGENTS § Commands); the registry test must read the CI semantics.
run_gate registry env -u GOEXPERIMENT go test -count=1 ./internal/arch/ -run TestErrorCodeRegistryIsFresh
run_gate dup bash -c 'command -v art-dupl >/dev/null 2>&1 || go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@v0.7.2; "$(go env GOPATH)/bin/art-dupl" check . -t 1 --baseline-path .art-dupl-baseline.json'
run_gate changelog scripts/changelog-guard.sh
run_gate templ scripts/templ-generate.sh --check
run_gate twcss bash scripts/build-tw-css.sh --check
run_gate docs bash scripts/docs-gates.sh --check

if [ "$fail" -ne 0 ]; then
	echo "pre-push-gates: RED — the battery found drift before CI did"
	exit 1
fi
echo "pre-push-gates: all requested gates green"
