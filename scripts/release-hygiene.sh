#!/usr/bin/env bash
# Train-hygiene checklist, scripted (plan T24, 2026-09-20).
#
# The post-train sweeps the release runbook expects, as ONE command:
#   1. hub-fanout benchmark — ONLY when cqrs-htmx or go-sse moved this
#      train (MD1 rule); skipped otherwise, forced with --fanout.
#   2. vulnix over the RUNTIME closure (network).
#   3. lychee link check (network; run AFTER the push so tag links
#      resolve; skipped with --no-net).
#   4. origin/main vs local HEAD (the daemon may have stalled).
# Exit 0 iff every executed step passes.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

NO_NET=false
FORCE_FANOUT=false
for arg in "$@"; do
	case "$arg" in
	--no-net) NO_NET=true ;;
	--fanout) FORCE_FANOUT=true ;;
	esac
done

FAILED=0
step() {
	local name="$1"
	shift
	echo "=== $name"
	if "$@"; then
		echo "--- ok: $name"
	else
		echo "--- FAILED: $name"
		FAILED=1
	fi
}

# 1. Hub fan-out (MD1): only when the SSE-carrying libraries moved.
if $FORCE_FANOUT || git log --since="14 days ago" -p -1 -- go.sum go.mod 2>/dev/null |
	grep -qE "github.com/larsartmann/cqrs-htmx|github.com/larsartmann/go-sse"; then
	step "hub-fanout benchmark (MD1)" \
		nix develop -c go test -count=1 -bench BenchmarkHubFanOut -benchtime 1s -run '^$' ./internal/server/
else
	echo "=== hub-fanout: skipped (no cqrs-htmx/go-sse change in the train)"
fi

# 2. vulnix over the runtime closure.
if ! $NO_NET; then
	nix build -o /tmp/webphone-hygiene .#webphone >/dev/null 2>&1
	step "vulnix runtime closure" nix run .#vulnix
else
	echo "=== vulnix: skipped (--no-net)"
fi

# 3. Link check.
if ! $NO_NET; then
	step "lychee link check" nix run nixpkgs#lychee -- .
else
	echo "=== lychee: skipped (--no-net)"
fi

# 4. Push-state verification.
echo "=== origin/main vs HEAD"
git fetch origin main >/dev/null 2>&1
LOCAL="$(git rev-parse HEAD)"
REMOTE="$(git rev-parse origin/main 2>/dev/null || echo unknown)"
if [[ "$LOCAL" == "$REMOTE" ]]; then
	echo "--- ok: origin/main tracks HEAD ($LOCAL)"
else
	echo "--- MISMATCH: local $LOCAL vs origin $REMOTE (daemon stalled? push)"
	FAILED=1
fi

if ((FAILED)); then
	echo "train hygiene: FAILED"
	exit 1
fi
echo "train hygiene: all executed steps green"
