#!/usr/bin/env bash
# status-predrain.sh — the cadence pre-drain sweep (pareto T16, 2026-11-21).
#
# WHY: cadence reds CI the day a new report lands while docs/status/ sits at
# its unarchived pin (66/66 as of 2026-11-21), and the archive ritual's own
# gates (annotate check-rows --strict + bare-rows) are only ever run BY HAND
# at drain time — so a session discovers which reports are archivable only
# after the red. This sweep pre-computes exactly that, oldest first:
#
#   READY   — passes BOTH archive gates; the exact ritual commands printed
#   BLOCKED — fails check-rows; the INCOMPLETE/UNTOUCHED table lines name
#             the harvest work standing between the report and the archive
#
# The sweep NEVER mutates and NEVER gates: the docs-gates wrapper wires it
# as a WARN lane so every local battery run points at the drain work before
# cadence goes red. (CI's --check mode has no annotate.go — the lane SKIPs
# there exactly like table-damage; the cadence battery itself stays the
# CI enforcement.)
#
# Usage:
#   nix develop -c bash scripts/status-predrain.sh [--limit N] [--self-test]
#
# Env:
#   PREDRAIN_STATUS_DIR  override the corpus dir (self-test fixture hook)
#   DOCS_HEALTH_ANNOTATE_GO  annotate.go source (default: the skill fan-out)
set -uo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
STATUS_DIR="${PREDRAIN_STATUS_DIR:-$ROOT/docs/status}"
ANNOTATE_GO="${DOCS_HEALTH_ANNOTATE_GO:-$HOME/.config/crush/skills/docs-health/assets/annotate.go}"
TOOL=/tmp/predrain-annotate/annotate
LIMIT=12

die() {
	echo "predrain: FATAL: $*" >&2
	exit 2
}

while [ $# -gt 0 ]; do
	case "$1" in
	--limit)
		LIMIT=${2:?--limit needs a value}
		shift 2
		;;
	--self-test)
		SELF_TEST=1
		shift
		;;
	*)
		die "unknown argument '$1'"
		;;
	esac
done
[ -d "$ROOT/.git" ] || die "no .git under $ROOT — check-rows resolves verdicts against git history"

# The tool is rebuilt per session (/tmp is wiped); rebuilt again when the
# skill fan-out source is newer than the cached binary.
if [ ! -x "$TOOL" ] || [ "$ANNOTATE_GO" -nt "$TOOL" ]; then
	command -v go >/dev/null 2>&1 || die "go not on PATH — run inside nix develop"
	mkdir -p "$(dirname "$TOOL")"
	(GOTOOLCHAIN=local GO111MODULE=off go build -o "$TOOL" "$ANNOTATE_GO") || die "cannot build the annotate tool from $ANNOTATE_GO"
fi

# Pins live in .docs-gates.json (read with python — the tool shell misinvokes jq).
read_pins() {
	python3 - "$ROOT/.docs-gates.json" <<'PY'
import json, sys
with open(sys.argv[1]) as f:
    cfg = json.load(f)
cad = cfg.get("cadence", {})
print(cad.get("max_unarchived", 66))
print(cad.get("oldest_days", 45))
PY
}
if [ "${SELF_TEST:-0}" != 1 ]; then
	pins=$(read_pins) || die "cannot read .docs-gates.json"
	MAX_UNARCHIVED=$(echo "$pins" | sed -n 1p)
	OLDEST_DAYS=$(echo "$pins" | sed -n 2p)
else
	MAX_UNARCHIVED=66
	OLDEST_DAYS=45
fi

age_days() {
	# "docs/status/2026-11-08_08-15_name.md" -> days since that stamp, -1 when unparsable
	local stamp
	stamp=$(basename "$1" | grep -oE '^[0-9]{4}-[0-9]{2}-[0-9]{2}')
	[ -n "$stamp" ] || {
		echo -1
		return
	}
	echo $((($(date +%s) - $(date -d "$stamp" +%s)) / 86400))
}

# classify <file> -> prints "READY" or "BLOCKED <why-first-line>"
classify() {
	local f="$1" why
	if out=$("$TOOL" check-rows --strict "$f" 2>&1); then
		if "$TOOL" bare-rows "$f" >/dev/null 2>&1; then
			echo "READY"
			return
		fi
		echo "BLOCKED bare-rows: data rows outside ledger tables must be struck or ledger-marked"
		return
	fi
	why=$(printf '%s\n' "$out" | grep -m1 -E 'INCOMPLETE|UNTOUCHED')
	echo "BLOCKED check-rows: ${why:-unharvested rows remain}"
}

sweep() {
	local dir="$1" limit="$2"
	local files=() f cls age oldest_age=-1 ready=0 blocked=0
	while IFS= read -r f; do
		files+=("$f")
	done < <(ls "$dir"/*.md 2>/dev/null | grep -v '/INDEX\.md$' | sort)
	local total=${#files[@]}
	echo "pre-drain sweep over $total unarchived report(s) (pin: $MAX_UNARCHIVED; oldest-age pin: ${OLDEST_DAYS}d)"
	echo
	for f in "${files[@]}"; do
		cls=$(classify "$f")
		age=$(age_days "$f")
		[ "$age" -gt "$oldest_age" ] && oldest_age=$age
		case "$cls" in
		READY*)
			ready=$((ready + 1))
			if [ "$ready" -le "$limit" ]; then
				echo "READY   $(basename "$f")  (age ${age}d)"
				echo "    # ARCHIVED $(date +%F) (docs-health v10 sweep): pre-drain verified (check-rows --strict + bare-rows green), drained for cadence"
				echo "    git mv $f $ROOT/docs/status/archived/$(basename "$f")"
			fi
			;;
		BLOCKED*)
			blocked=$((blocked + 1))
			if [ "$blocked" -le "$limit" ]; then
				echo "BLOCKED $(basename "$f")  (age ${age}d) — ${cls#BLOCKED }"
			fi
			;;
		esac
	done
	echo
	echo "totals: $total unarchived (pin $MAX_UNARCHIVED), oldest ${oldest_age}d (pin ${OLDEST_DAYS}d); READY $ready / BLOCKED $blocked"
	if [ "$ready" -gt 0 ] && [ "$total" -ge "$MAX_UNARCHIVED" ]; then
		echo "cadence at the pin with $ready archivable report(s): drain BEFORE the next report lands (see READY lines), then status-index.py + docs-gates"
	fi
}

self_test() {
	local fixture
	fixture=$(mktemp -d)
	mkdir -p "$fixture/status"
	cat >"$fixture/status/2026-01-01_00-00_ready.md" <<'MD'
# Ready report

| a | b |
| --- | --- |
| ~~x~~ | ~~done 2026-01-02 — evidence~~ |
| ~~y~~ | ~~done 2026-01-02 — evidence~~ |
MD
	cat >"$fixture/status/2026-01-01_00-01_blocked.md" <<'MD'
# Blocked report

| a | b |
| --- | --- |
| ~~x~~ | done 2026-01-02 — evidence |
| open row | still to do |
MD
	touch "$fixture/status/INDEX.md"
	local out ready_lines blocked_lines
	out=$(PREDRAIN_STATUS_DIR="$fixture/status" "$0" 2>&1)
	ready_lines=$(printf '%s\n' "$out" | grep -c '^READY')
	blocked_lines=$(printf '%s\n' "$out" | grep -c '^BLOCKED')
	rm -rf "$fixture"
	if [ "$ready_lines" -ne 1 ] || [ "$blocked_lines" -ne 1 ]; then
		echo "self-test FAILED (ready=$ready_lines blocked=$blocked_lines, want 1/1):" >&2
		printf '%s\n' "$out" >&2
		return 1
	fi
	echo "self-test: OK (fixture classified READY=1 BLOCKED=1)"
	return 0
}

if [ "${SELF_TEST:-0}" = 1 ]; then
	self_test
	exit $?
fi

[ -d "$STATUS_DIR" ] || die "no status dir at $STATUS_DIR"
sweep "$STATUS_DIR" "$LIMIT"
