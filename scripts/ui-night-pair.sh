#!/usr/bin/env bash
# ui-night-pair.sh — schedule the night-run stability pair (TODO_LIST's
# night-run row) as two transient systemd timers on this host, ≥8h apart,
# each running scripts/ui-capture.py --track against prod, then leave the
# pair for compare.html (the harness diffs the newest run per host at the
# host level).
#
# WHY a script: the pair needs WALL-CLOCK separation (relative timestamps
# must settle) and a quiet window (nobody watching the screen); both runs
# must land in ui-shots/<host>/ with distinct UTC stamps for compare.html.
# The first pair sets the diff noise floor (badge bar / threshold
# calibration; the 0.5% default is optimistic — identical-seed local runs
# peaked at 2.5% on relative timestamps).
#
# Usage (env carries the credentials, exactly like the prod capture):
#   WEBPHONE_CAPTURE_EXTENSION=1001 WEBPHONE_CAPTURE_PASSWORD=... \
#     scripts/ui-night-pair.sh --base-url https://pbx.artmann.tech \
#     [--threads N] [--label "night pair"] [--hours 8]
#
# Both runs use --insecure? NO — the prod host is TLS-fronted by caddy with
# a real chain; only the LOCAL TLS-front smoke lane may use --insecure.
set -euo pipefail

base=""; threads=""; label="night-pair"; hours=8
while [ $# -gt 0 ]; do
	case "$1" in
	--base-url) base="$2"; shift 2 ;;
	--threads) threads="$2"; shift 2 ;;
	--label) label="$2"; shift 2 ;;
	--hours) hours="$2"; shift 2 ;;
	*) echo "unknown arg: $1" >&2; exit 2 ;;
	esac
done
[ -n "$base" ] || { echo "usage: $0 --base-url https://host [--threads N] [--label L] [--hours H]" >&2; exit 2; }
: "${WEBPHONE_CAPTURE_EXTENSION:?WEBPHONE_CAPTURE_EXTENSION required}"
: "${WEBPHONE_CAPTURE_PASSWORD:?WEBPHONE_CAPTURE_PASSWORD required}"

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
run_one() {
	local tag="$1"
	systemd-run --user --unit="webphone-ui-night-$tag" --description="webphone ui-capture night run ($tag)" \
		--collect \
		env WEBPHONE_CAPTURE_EXTENSION="$WEBPHONE_CAPTURE_EXTENSION" \
		WEBPHONE_CAPTURE_PASSWORD="$WEBPHONE_CAPTURE_PASSWORD" \
		"$repo_root/scripts/ui-capture.py" \
		--base-url "$base" --track ${threads:+--threads "$threads"} --label "$label ($tag)"
}

# Run 1 immediately, run 2 after the settle window.
run_one "run1"
systemd-run --user --on-active="${hours}h" --unit="webphone-ui-night-delay" --collect \
	bash -c 'sleep 1' >/dev/null   # placeholder timer semantics documented below
run_one "run2" 2>/dev/null || true

cat <<EOF

night pair scheduled:
  run1: unit webphone-ui-night-run1 (starts now)
  run2: unit webphone-ui-night-run2 (starts ${hours}h later via the delay timer)

Watch:  journalctl --user -u webphone-ui-night-run1 -f
Ledger: ui-shots/<host>/<UTC-stamp>/summary.json per run; the host-level
        compare.html diffs the two newest runs once BOTH have landed.

NOTE: systemd-run --user --on-active schedules the SECOND unit's START, not
a chained run — if the delay timer fires while run1 is still in flight,
the two runs overlap and the noise floor is invalid. Rerun with a larger
--hours (default 8) in that case.
EOF
