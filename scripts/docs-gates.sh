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
#   nix develop -c bash scripts/docs-gates.sh            # human: format, then judge
#   nix develop -c bash scripts/docs-gates.sh --check    # CI: judge only, never mutate
#
# --check exists because `dprint fmt` MUTATES: a CI run that formats before
# judging would certify a tree it had just repaired. CI must run `dprint
# check` and fail on drift the author left behind. The table-damage gate
# needs the docs-health skill's annotate.go, which is a local fan-out file
# and not in CI — under --check it reports SKIP (visible in the matrix,
# never silent) instead of failing the build for a missing ritual tool.
#
# Exit 0 = formatted + all gates green; 1 = a gate failed; 2 = environment.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)

fatal() {
	echo "docs-gates: FATAL: $*" >&2
	exit 2
}

MODE=human
case "${1:-}" in
--check) MODE=check ;;
"") ;;
*) fatal "unknown argument '$1' (the only mode flag is --check)" ;;
esac

command -v dprint >/dev/null 2>&1 ||
	fatal "dprint not on PATH — run: nix develop -c bash scripts/docs-gates.sh"
# Fail CLOSED on a missing battery: a CI leg that cannot find the tool must
# not quietly certify the docs. The devShell provides it (config.packages
# .docs-gates), so absence is a real misconfiguration.
command -v docs-gates >/dev/null 2>&1 ||
	fatal "docs-gates not on PATH — run: nix develop -c bash scripts/docs-gates.sh (or: nix build .#docs-gates && PATH=\"$(nix build --no-link --print-out-paths .#docs-gates)/bin:$PATH\")"
ANNOTATE_GO="${DOCS_HEALTH_ANNOTATE_GO:-$HOME/.config/crush/skills/docs-health/assets/annotate.go}"
if [ ! -f "$ANNOTATE_GO" ]; then
	if [ "$MODE" = check ]; then
		echo "docs-gates: WARN: annotate.go not found at $ANNOTATE_GO — the table-damage gate will SKIP (local-only ritual step)" >&2
	else
		fatal "annotate.go not found at $ANNOTATE_GO — the docs-health skill fan-out is missing (override via DOCS_HEALTH_ANNOTATE_GO)"
	fi
fi
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
GATE_SKIPPED=()

skip_gate() {
	GATE_NAMES+=("$1")
	GATE_RCS+=(0)
	GATE_SKIPPED+=(1)
	echo "docs-gates: gate $1 SKIPPED — $2"
}

run_gate() {
	local name="$1"
	shift
	GATE_NAMES+=("$name")
	GATE_SKIPPED+=(0)
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

# Kills the VACUOUS PASS. The freshness gate reads the newest report's §f and
# certifies "nothing unharvested" when that section is MISSING — a report with
# no harvest list is not fresh, it is unexamined (observed twice:
# 2026-11-06_10-45 and 2026-11-07_19-25). The battery cannot see the
# difference, so the wrapper asserts the shape BEFORE trusting its verdict.
status_f_gate() {
	# DOCS_GATES_STATUS_DIR exists so this gate can be proved RED against a
	# fixture (a self-test that cannot fail is decoration).
	local dir="${DOCS_GATES_STATUS_DIR:-docs/status}" newest
	newest=$(find "$dir" -maxdepth 1 -name '*.md' ! -name 'INDEX.md' -printf '%f\n' | sort | tail -1)
	[ -n "$newest" ] || {
		echo "no status report found under docs/status/ — nothing to certify" >&2
		return 1
	}
	if ! grep -qE '^## f\)' "$dir/$newest"; then
		{
			echo "the NEWEST status report carries no '## f)' section: $dir/$newest"
			echo "The freshness gate would PASS it vacuously (nothing to check)."
			echo "Fix: add '## f) Up to 50 things we should get done next' with ranked rows"
			echo "whose tokens match their TODO_LIST home, then re-run this wrapper."
		} >&2
		return 1
	fi
	echo "docs-freshness-precondition: $dir/$newest carries '## f)' — the newest-report verdict is meaningful"
}

if [ "$MODE" = check ]; then
	echo "docs-gates: [1/4] dprint check (no mutation in --check mode)"
	run_gate "format" dprint check
else
	echo "docs-gates: [1/4] dprint fmt"
	dprint fmt
fi

run_gate "status-f-section" status_f_gate

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

if [ ! -f "$ANNOTATE_GO" ] || ! command -v go >/dev/null 2>&1; then
	skip_gate "table-damage" "needs go + the docs-health annotate.go (local ritual step; the battery above still ran)"
else
	run_gate "table-damage" table_damage_gate
fi

echo "docs-gates: verdict matrix"
rc=0
for i in "${!GATE_NAMES[@]}"; do
	if [ "${GATE_SKIPPED[$i]}" -eq 1 ]; then
		printf '  SKIP  %s\n' "${GATE_NAMES[$i]}"
	elif [ "${GATE_RCS[$i]}" -eq 0 ]; then
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
