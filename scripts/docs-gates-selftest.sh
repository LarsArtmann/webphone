#!/usr/bin/env bash
# docs-gates-selftest.sh — proves the wrapper's own gates can FAIL.
#
# A gate that cannot go red is decoration. This pins the one gate that is
# purely ours (the battery and table-damage gates are the upstream tool's and
# carry their own fixtures):
#
#   status_f_gate — the newest docs/status report MUST carry a `## f)`
#   section. Without it `docs-gates freshness` certifies "0 unharvested
#   open §f items" for a report that has no §f at all: a PASS that means
#   "did not look" (observed 2026-11-06_10-45 and 2026-11-07_19-25).
#
# Both directions are exercised against fixtures, so the real corpus is
# never mutated. Precedent: scripts/pre-commit-guard-selftest.sh.
#
# Exit 0 = the gate reds on the offending shape and greens on the good one.
set -uo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 2

GATE_DIR=$(mktemp -d)
trap 'rm -rf "$GATE_DIR"' EXIT

# Extract the gate body from the wrapper: ONE definition, no second copy to
# drift (the selftest must exercise the shipped code, not a paraphrase).
gate_body="$GATE_DIR/gate.sh"
sed -n '/^status_f_gate()/,/^}/p' scripts/docs-gates.sh >"$gate_body"
if [ ! -s "$gate_body" ]; then
	echo "selftest: FAIL — could not extract status_f_gate from scripts/docs-gates.sh" >&2
	exit 1
fi

fail=0

run_gate() {
	local dir="$1"
	( set -e
		# shellcheck disable=SC1090
		source "$gate_body"
		DOCS_GATES_STATUS_DIR="$dir" status_f_gate )
}

echo "selftest: case 1 — newest report WITHOUT '## f)' must fail"
mkdir -p "$GATE_DIR/red"
printf '# Status\n\n## a) FULLY DONE\n\n- x\n\n## e) Notes\n\n- y\n' \
	>"$GATE_DIR/red/2026-11-07_20-00_no-f.md"
if run_gate "$GATE_DIR/red" >/dev/null 2>&1; then
	echo "selftest: FAIL — a §f-less newest report was certified fresh" >&2
	fail=1
else
	echo "selftest: PASS — red on the offending shape"
fi

echo "selftest: case 2 — newest report WITH '## f)' must pass"
mkdir -p "$GATE_DIR/green"
printf '# Status\n\n## f) Up to 50 things\n\n| # | Task |\n| - | ---- |\n| 1 | z |\n' \
	>"$GATE_DIR/green/2026-11-07_20-00_has-f.md"
if run_gate "$GATE_DIR/green" >/dev/null 2>&1; then
	echo "selftest: PASS — green on the good shape"
else
	echo "selftest: FAIL — a report carrying §f was rejected" >&2
	fail=1
fi

echo "selftest: case 3 — the ORDER matters (a later report wins over an earlier good one)"
mkdir -p "$GATE_DIR/mixed"
printf '# Status\n\n## f) Up to 50 things\n\n| # | Task |\n| - | ---- |\n| 1 | z |\n' \
	>"$GATE_DIR/mixed/2026-11-07_09-00_has-f.md"
printf '# Status\n\n## a) FULLY DONE\n\n- x\n' \
	>"$GATE_DIR/mixed/2026-11-07_20-00_no-f.md"
if run_gate "$GATE_DIR/mixed" >/dev/null 2>&1; then
	echo "selftest: FAIL — an older fresh report masked a newer §f-less one" >&2
	fail=1
else
	echo "selftest: PASS — the NEWEST report is the one judged"
fi

if [ "$fail" -eq 0 ]; then
	echo "selftest: 3 passed — the §f precondition can fail"
else
	echo "selftest: FAILED — the §f precondition is not load-bearing"
fi
exit "$fail"
