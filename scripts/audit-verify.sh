#!/usr/bin/env bash
# audit-verify.sh — replay the fleet-audit reports' "How to verify" blocks.
#
# Every docs/review/libs/*.md carries a verify block whose commands pin the
# audit's claims to the tree. This script extracts those console fences,
# runs each command from the repo root, and checks the machine-readable
# claims: `# expect N` (count), `# expect >= N` / `# expect > N`, and
# `# expect vX.Y.Z` (substring). Prose expects run without a numeric
# oracle: a zero-exit command with output passes, a non-zero exit fails.
#
# Usage:
#   scripts/audit-verify.sh                  # replay all reports
#   scripts/audit-verify.sh go-etag.md ...   # replay named reports
#
# Exit 0 only when every command ran and every checkable claim held.
# Stale claims are the harness's product: fix the REPORT (the tree moved)
# or the TREE (the claim broke) — never the script.

set -u

cd "$(git rev-parse --show-toplevel)" || exit 1

LIBS_DIR="docs/review/libs"

if [ "$#" -gt 0 ]; then
	REPORTS=""
	for name in "$@"; do
		REPORTS="$REPORTS $LIBS_DIR/$name"
	done
else
	REPORTS=$(ls "$LIBS_DIR"/*.md)
fi

total_cmds=0
total_pass=0
total_fail=0
failed=()

# run_claim <file> <command> <expect-clause|''>
run_claim() {
	_file=$1
	_cmd=$2
	_expect=$3
	total_cmds=$((total_cmds + 1))

	# eval, not $_cmd: the commands carry quoted patterns and pipes that
	# word-splitting must not re-chop. Content is committed repo prose.
	_out=$(eval "$_cmd" 2>&1)
	_status=$?
	_lines=$(printf '%s\n' "$_out" | grep -c .)

	verdict=RUN-OK
	if [ "$_status" -ne 0 ] && [ -z "$_out" ]; then
		verdict=FAIL
	fi

	if [ -n "$_expect" ]; then
		# numeric with comparator: ">= 14" / "> 0"
		case "$_expect" in
		">="*)
			cond=ge
			_n=$(printf '%s' "$_expect" | sed 's/^>= *//;s/[^0-9].*$//')
			;;
		">"*)
			cond=gt
			_n=$(printf '%s' "$_expect" | sed 's/^> *//;s/[^0-9].*$//')
			;;
		*)
			cond=eq
			_n=$(printf '%s' "$_expect" | grep -oE '^[0-9]+')
			;;
		esac

		if [ -n "$_n" ]; then
			_got=$(printf '%s\n' "$_out" | grep -oE '[0-9]+' | sort -rn | head -1)
			_got=${_got:-0}
			_ok=""
			[ "$cond" = ge ] && [ "$_got" -ge "$_n" ] && _ok=yes
			[ "$cond" = gt ] && [ "$_got" -gt "$_n" ] && _ok=yes
			[ "$cond" = eq ] && { [ "$_lines" -eq "$_n" ] || [ "$(printf '%s' "$_out" | tr -d ' ')" = "$_n" ]; } && _ok=yes
			if [ -n "$_ok" ]; then
				verdict=PASS
			else
				verdict="FAIL: want $_expect, got $_got"
			fi
		else
			# version substring: "v0.9.0 ..." — the claim is contained text
			_v=$(printf '%s' "$_expect" | grep -oE '^v[0-9]+\.[0-9]+\.[0-9]+')
			if printf '%s' "$_out" | grep -q "$_v"; then
				verdict=PASS
			else
				verdict="FAIL: want $_v in output, got '$(printf '%s' "$_out" | head -1)'"
			fi
		fi
	fi

	case "$verdict" in
	PASS* | RUN-OK) total_pass=$((total_pass + 1)) ;;
	*)
		total_fail=$((total_fail + 1))
		failed+=("$_file: $_cmd => $verdict")
		;;
	esac
	printf '%-24s %-4s %s\n' "${_file#"$LIBS_DIR"/}" "$verdict" "$(printf '%s' "$_cmd" | cut -c1-72)"
}

for report in $REPORTS; do
	[ -f "$report" ] || {
		echo "no such report: $report"
		exit 1
	}
	# Extract the How-to-verify console fence; skip comment lines.
	_block=$(awk '/^## How to verify/{f=1; next} f && /^```console/{c=1; next} f && c && /^```/{exit} c {print}' "$report")
	while IFS= read -r line; do
		[ -z "$line" ] && continue
		case "$line" in
		\#*) continue ;;
		esac
		# Split a trailing `# expect ...` clause off the command.
		expect=""
		case "$line" in
		*"# "*)
			cmd=${line%%\#*}
			expect=$(printf '%s' "$line" | sed 's/^.*# *expect *//')
			[ "$expect" = "$line" ] && expect=""
			;;
		*)
			cmd=$line
			;;
		esac
		cmd=$(printf '%s' "$cmd" | sed 's/[[:space:]]*$//')
		[ -z "$cmd" ] && continue
		run_claim "$report" "$cmd" "$expect"
	done <<EOF
$_block
EOF
done

echo
echo "audit-verify: $total_pass passed, $total_fail failed of $total_cmds commands"
if [ "$total_fail" -gt 0 ]; then
	printf '%s\n' "${failed[@]}"
	exit 1
fi
exit 0
