#!/usr/bin/env bash
# ui-night-pair.sh — schedule the night-run stability pair (TODO_LIST's
# night-run row) as two transient systemd user units ≥8h apart, each
# running scripts/ui-capture.py --track against prod. The harness's
# host-level compare.html diffs the two newest runs once BOTH landed.
#
# WHY a script: the pair needs WALL-CLOCK separation (relative timestamps
# must settle) and a quiet window; both runs land in ui-shots/<host>/ with
# distinct UTC stamps. The first pair sets the diff noise floor (badge
# bar / threshold calibration; the 0.5% default is optimistic —
# identical-seed local runs peaked at 2.5% on relative timestamps).
#
# Usage (env carries the credentials, exactly like the prod capture):
#   WEBPHONE_CAPTURE_EXTENSION=1001 WEBPHONE_CAPTURE_PASSWORD=... \
#     scripts/ui-night-pair.sh --base-url https://pbx.artmann.tech \
#     [--threads N] [--label "night pair"] [--hours 8]
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
capture_argv=(
	"$repo_root/scripts/ui-capture.py"
	--base-url "$base"
	--track
	--label "$label"
)
[ -z "$threads" ] || capture_argv+=(--threads "$threads")

# Run 1 starts now; run 2 is scheduled ${hours}h later. Two independent
# transient units — the delay lives on run2's OWN start, so the runs can
# never overlap unless run1 somehow runs longer than the settle window
# (rerun with a larger --hours if a ledger pair ever looks overlapped).
systemd-run --user --unit=webphone-ui-night-run1 --collect \
	--description="webphone ui-capture night pair, run 1 (now)" \
	--setenv=WEBPHONE_CAPTURE_EXTENSION="$WEBPHONE_CAPTURE_EXTENSION" \
	--setenv=WEBPHONE_CAPTURE_PASSWORD="$WEBPHONE_CAPTURE_PASSWORD" \
	"${capture_argv[@]}"

systemd-run --user --unit=webphone-ui-night-run2 --collect \
	--on-active="${hours}h" \
	--description="webphone ui-capture night pair, run 2 (now+${hours}h)" \
	--setenv=WEBPHONE_CAPTURE_EXTENSION="$WEBPHONE_CAPTURE_EXTENSION" \
	--setenv=WEBPHONE_CAPTURE_PASSWORD="$WEBPHONE_CAPTURE_PASSWORD" \
	"${capture_argv[@]}"

cat <<EOF

night pair scheduled:
  run1: unit webphone-ui-night-run1 — starts now
  run2: unit webphone-ui-night-run2 — starts in ${hours}h (--on-active)

Watch:   journalctl --user -u webphone-ui-night-run1 -f   (then run2)
Ledger:  ui-shots/<host>/<UTC-stamp>/summary.json per run; compare.html
         at the host level diffs the newest pair once run2 has landed.
Cleanup: the units are --collect (gone when done); list with
         systemctl --user list-units 'webphone-ui-night-*'
EOF
