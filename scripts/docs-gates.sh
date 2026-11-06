#!/usr/bin/env bash
# docs-gates.sh — ONE command for the docs-health ritual on webphone.
#
# Why a wrapper instead of three remembered steps (the 2026-11-06 v8 sweep's
# lesson, inherited from pbx-artmann's 2026-09-26 lost-cycle incident): the
# ORDER is load-bearing and every session re-derived it.
#
#   1. dprint fmt        — mutates FIRST, so every gate below judges the
#                          FORMATTED tree. docs-annotate rule F ("space
#                          stripped around inline code") is precisely the
#                          damage dprint's markdown reflow can inflict, so
#                          formatting after the gates invalidates them.
#                          This is also why rule F fixes must avoid the
#                          construct entirely rather than restore the space.
#   2. docs-gates run .  — the battery: freshness, audit-all, annotate,
#                          routing --strict, cadence, one verdict matrix.
#                          Ratchets (annotate_max / audit_max / cadence) live
#                          in .docs-gates.json; a ratchet is a SINKHOLE —
#                          lowering it is progress, raising it is a
#                          regression that needs evidence.
#   3. table-damage      — annotate.go repair-mangle --dry-run over the
#                          status corpus + the living docs + the plans. The
#                          mechanical damage classes (doubled strikes,
#                          collapsed empty cells, boundary quads) survive
#                          every other gate because uniformity is not shape.
#                          A planted fixture MUST trip it first: a gate that
#                          cannot fail is decoration.
#
# Run from a checkout (the hash/tag rules need .git):
#   nix develop -c bash scripts/docs-gates.sh
#
# Exit 0 = formatted + all gates green; 1 = a gate failed; 2 = environment.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)

fatal() {
	echo "docs-gates: FATAL: $*" >&2
	exit 2
}

command -v dprint >/dev/null 2>&1 ||
	fatal "dprint not on PATH — run: nix develop -c bash scripts/docs-gates.sh"
command -v docs-gates >/dev/null 2>&1 ||
	fatal "docs-gates not on PATH — install github:Artmann-Technologies/docs-gates"
command -v go >/dev/null 2>&1 ||
	fatal "go not on PATH — the table-damage gate runs the docs-health annotate.go"
ANNOTATE_GO="${DOCS_HEALTH_ANNOTATE_GO:-$HOME/.config/crush/skills/docs-health/assets/annotate.go}"
[ -f "$ANNOTATE_GO" ] ||
	fatal "annotate.go not found at $ANNOTATE_GO — the docs-health skill fan-out is missing (override via DOCS_HEALTH_ANNOTATE_GO)"
[ -d "$ROOT/.git" ] || fatal "no .git under $ROOT — the annotate hash/tag rules need a checkout"
[ -f "$ROOT/dprint.json" ] || fatal "no dprint.json under $ROOT — wrong repo?"

cd "$ROOT"

shopt -s nullglob
STATUS_FILES=(docs/status/*.md docs/status/archived/*.md)
PLAN_FILES=(docs/planning/*.md docs/planning/archived/*.md docs/reviews/*.md docs/research/*.md)
shopt -u nullglob

DAMAGE_FILES=(
	TODO_LIST.md FEATURES.md README.md ROADMAP.md AGENTS.md CHANGELOG.md
	"${STATUS_FILES[@]}" "${PLAN_FILES[@]}"
)

GATE_NAMES=()
GATE_RCS=()

run_gate() {
	local name="$1"
	shift
	GATE_NAMES+=("$name")
	echo "docs-gates: gate $name ..."
	local log
	log=$(mktemp)
	if "$@" >"$log" 2>&1; then
		GATE_RCS+=(0)
		cat "$log"
	else
		local rc=$?
		GATE_RCS+=("$rc")
		echo "docs-gates: gate $name FAILED (rc=$rc):" >&2
		cat "$log" >&2
	fi
	rm -f "$log"
}

echo "docs-gates: [1/3] dprint fmt"
dprint fmt

run_gate "battery" docs-gates run .

table_damage_gate() {
	local fixture rc=0
	fixture=$(mktemp -d)
	printf '| a | b |\n| --- | --- |\n| ~~~~x~~~~ | y |\n' >"$fixture/damaged.md"
	if go run "$ANNOTATE_GO" repair-mangle --dry-run "$fixture/damaged.md" >/dev/null 2>&1; then
		echo "the damage gate cannot FAIL: repair-mangle no longer detects the planted mangle" >&2
		rc=1
	fi
	rm -rf "$fixture"
	[ "$rc" -eq 0 ] || return 1
	local out rc=0
	out=$(go run "$ANNOTATE_GO" repair-mangle --dry-run "${DAMAGE_FILES[@]}" 2>&1) || rc=$?
	# Only the files that actually carry damage are signal; the per-file
	# "0 row(s)" lines are ~300 rows of noise on a clean corpus.
	printf '%s\n' "$out" | grep -v '0 row(s), 0 manual finding(s)' || true
	return "$rc"
}

run_gate "table-damage" table_damage_gate

echo "docs-gates: verdict matrix"
rc=0
for i in "${!GATE_NAMES[@]}"; do
	if [ "${GATE_RCS[$i]}" -eq 0 ]; then
		printf '  PASS  %s\n' "${GATE_NAMES[$i]}"
	else
		printf '  FAIL  %s (rc=%s)\n' "${GATE_NAMES[$i]}" "${GATE_RCS[$i]}"
		rc=1
	fi
done

if [ "$rc" -eq 0 ]; then
	echo "docs-gates: PASS — formatted + all doc gates green"
else
	echo "docs-gates: FAIL — the matrix above names the red gates" >&2
fi
exit "$rc"
