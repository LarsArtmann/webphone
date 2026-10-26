#!/usr/bin/env bash
# incident-bundle: assemble the evidence pack for an incident BEFORE any
# diagnosis (SUPERB error-agents M12-M13). Flight-recorder snapshot
# window, CI-run window, commit window, and an error-context sweep of a
# supplied log land in one drafted bundle under docs/drafts/ — with
# lessons/CHANGELOG/ledger SKELETONS whose verdicts stay TBD. No doc is
# ever auto-merged: the owner fills every verdict (exoskeleton doctrine).
#
# usage:
#   incident-bundle.sh --since <iso-ts> --dataDir <dir> [--log-file <path>]
#                     [--slug <name>] [--out <dir>]
#   incident-bundle.sh --self-test
set -euo pipefail

say() { echo "incident-bundle: $*"; }
REPO=$(git rev-parse --show-toplevel)

self_test() {
	local tmp
	tmp=$(mktemp -d)
	trap 'rm -rf "${tmp:-}"' EXIT
	mkdir -p "$tmp/data/traces"
	head -c 2048 /dev/urandom >"$tmp/data/traces/2026-10-26T10-00-00Z.trace.gz"
	local logf="$tmp/journal.log"
	cat >"$logf" <<'LOG'
Oct 26 10:00:01 host webphone[7]: journal: send failed: [gateway.http] provider 502
Oct 26 10:00:02 host webphone[7]: journal: classify: store.updated applied
LOG
	local out="$tmp/bundle"
	"$REPO/scripts/incident-bundle.sh" \
		--since "$(date -u -d '2 hours ago' +%Y-%m-%dT%H:%M:%SZ)" \
		--dataDir "$tmp/data" \
		--log-file "$logf" \
		--slug selftest --out "$out" >/dev/null 2>&1

	local checks=0 failed=0
	assert() {
		checks=$((checks + 1))
		if [ "$1" != "$2" ]; then
			failed=$((failed + 1))
			echo "  FAIL  $3 (wanted '$1', got '$2')"
		else
			echo "  PASS  $3"
		fi
	}
	local bundle
	bundle=$(find "$out" -name bundle.md | head -1)
	assert "1" "$([ -n "$bundle" ] && echo 1 || echo 0)" "bundle.md written"
	local body
	body=$(cat "$bundle")
	has() {
		grep -q "$1" <<<"$body" && echo 1 || echo 0
	}
	assert "1" "$(has 'gateway.http')" "code sweep found gateway.http"
	assert "1" "$(has 'verdict: TBD')" "lessons skeleton keeps verdict TBD"
	assert "1" "$(has 'CHANGELOG candidate')" "changelog candidate present"
	assert "1" "$(has 'flake-ledger row')" "ledger skeleton present"
	assert "1" "$(has 'flight-recorder snapshots')" "traces section present"
	echo "incident-bundle self-test: $((checks - failed))/$checks passed"
	[ "$failed" -eq 0 ]
}

[ "${1:-}" = "--self-test" ] && {
	self_test
	exit $?
}

SINCE="" DATADIR="" LOGFILE="" SLUG="" OUT=""
while [ $# -gt 0 ]; do
	case "$1" in
	--since)
		SINCE=$2
		shift 2
		;;
	--dataDir)
		DATADIR=$2
		shift 2
		;;
	--log-file)
		LOGFILE=$2
		shift 2
		;;
	--slug)
		SLUG=$2
		shift 2
		;;
	--out)
		OUT=$2
		shift 2
		;;
	*)
		say "unknown argument $1"
		exit 1
		;;
	esac
done
[ -n "$SINCE" ] && [ -n "$DATADIR" ] || {
	say "--since and --dataDir are required"
	exit 1
}
SLUG=${SLUG:-incident}
OUT=${OUT:-"$REPO/docs/drafts/incident-$(date +%F-%H%M)-$SLUG"}
mkdir -p "$OUT"

{
	echo "# incident bundle — $SLUG"
	echo
	echo "assembled: $(date -u +%FT%TZ) · window since: $SINCE · evidence only, verdicts are human"
	echo

	echo "## flight-recorder snapshots (window)"
	if [ -d "$DATADIR/traces" ]; then
		FOUND=$(find "$DATADIR/traces" -type f -newermt "$SINCE" -printf '%p %s bytes\n' 2>/dev/null || true)
		if [ -n "$FOUND" ]; then
			echo "$FOUND"
			echo
			echo "read one: \`gunzip <file> && go tool trace <unzipped>\` (best-speed gzip, go tool trace native)"
		else
			echo "(no snapshots in window — the flight recorder only writes on >2s requests and 5xx)"
		fi
	else
		echo "(WARN: $DATADIR/traces does not exist — recorder seam off or wrong dataDir)"
	fi
	echo

	echo "## CI runs (window)"
	if [ "${BUNDLE_SKIP_GH:-}" = "1" ]; then
		echo "(skipped in self-test)"
	elif command -v gh >/dev/null 2>&1; then
		gh run list --limit 20 --json createdAt,displayTitle,conclusion,url 2>/dev/null |
			python3 - "$SINCE" <<'PY' || echo "(gh window unavailable — check GH_TOKEN)"
import json, sys
since = sys.argv[1]
try:
    runs = json.load(sys.stdin)
except Exception:
    print("(gh returned nothing — token scope?)")
    raise SystemExit
hits = [r for r in runs if r.get("createdAt", "") >= since]
if not hits:
    print("(no runs in window)")
for r in hits:
    print(f"- {r['createdAt']} {r['conclusion']}: {r['displayTitle'][:70]} {r['url']}")
PY
	else
		echo "(gh unavailable)"
	fi
	echo

	echo "## commits (window)"
	git -C "$REPO" log --since="$SINCE" --oneline | head -20 || true
	echo

	echo "## error-context sweep of the supplied log"
	if [ -n "$LOGFILE" ] && [ -f "$LOGFILE" ]; then
		CODES=$(grep -oE '\[[a-z]+\.[a-z_0-9.]+\]' "$LOGFILE" | tr -d '[]' | sort -u | head -5)
		if [ -n "$CODES" ]; then
			while IFS= read -r code; do
				[ -n "$code" ] || continue
				echo "### $code"
				python3 "$REPO/scripts/error-context.py" "$code" 2>/dev/null | sed 's/^/    /' || echo "    (error-context unavailable)"
				echo
			done <<<"$CODES"
		else
			echo "(no bracketed error codes found in the log)"
		fi
	else
		echo "(no --log-file supplied — journald needs root: journalctl -u webphone > file, then re-run)"
	fi

	echo "## drafted skeletons — every verdict stays with the owner"
	cat <<'SKELETON'

### lessons.md row (verdict: TBD)

| date | train | what happened | verdict: TBD — owner rules after reading the evidence above |

### CHANGELOG candidate line

- CHANGELOG candidate: (one line, Fixed/Added section per nature — delete if test/docs-only)

### docs/ci-flake-ledger.md row (if CI-adjacent)

- flake-ledger row: | <run id> | <date> | <stage/test> | TBD — owner verdict; drafted by incident-bundle |

> Runbook pointer: assemble THIS bundle first, then diagnose — the
> plan (SUPERB error-agents M13) keeps README untouched until the first
> real incident proves the bundle shape. TODO_LIST cites this seam.
SKELETON
} >"$OUT/bundle.md"

say "bundle written: $OUT/bundle.md"
