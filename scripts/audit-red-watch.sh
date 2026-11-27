#!/usr/bin/env bash
# audit-red-watch: the fleet-audit rot detector in the watch lane
# (pareto T05). Replays the docs/review/libs reports' verify blocks via
# scripts/audit-verify.sh on a schedule, so a rotted verify line is
# caught the DAY it breaks instead of at the next manual sweep — the
# 00-52 §f 15 ask. One tracker issue per failing report (LIST-api dedup
# by title prefix; the SEARCH index lags for fresh issues — the
# 2026-11-04 199-duplicate class), a comment on recurrence.
#
# Tip-unchanged skip: when the checkout HEAD equals the PREVIOUS
# completed watch run's head SHA AND that run's audit step succeeded,
# the replay is skipped — same tree, same verdict, and the reports only
# change when HEAD does. A prior FAILURE at the same tip re-runs (the
# verdict may have been transient, and the issue stays the signal).
#
# Advisory: exits 0 unless its environment is broken; a rotted report
# is SIGNALED (issue/comment), never failed-on — the watch workflow
# stays green so the signal is the issue.
#
# usage:
#   audit-red-watch.sh [--dry-run]   # print the bundle, file nothing
#   audit-red-watch.sh --self-test   # fixtures through the decision core
set -euo pipefail

say() { echo "audit-red-watch: $*"; }

DRY_RUN=0
for arg in "$@"; do
	case "$arg" in
	--dry-run) DRY_RUN=1 ;;
	--self-test) self_test_mode=1 ;;
	*) echo "unknown arg $arg" >&2; exit 2 ;;
	esac
done

# ---------------------------------------------------------------- core
# The decision core (pure): given the audit-verify summary line, the
# open-issue titles for the failing report, and whether a previous
# successful audit ran at this exact tip, decide SKIP / NONE / EXISTS
# <n> / COMMENT <n> / CREATE.
decide() {
	python3 - "$1" "$2" "$3" <<'PY'
import json, sys

report, open_titles_json, prior_pass = sys.argv[1], sys.argv[2], sys.argv[3]
if prior_pass == "1":
    print(f"SKIP same-tip-verified {report}")
    raise SystemExit
same = json.loads(open_titles_json or "[]")
if same:
    print(f"EXISTS {same[0]}")
    raise SystemExit
print(f"CREATE {report}")
PY
}

# The watch lane's own history is the skip state (stateless across CI
# checkouts): the most recent COMPLETED run before this one with the
# SAME head SHA, and whether its audit step succeeded.
prior_audit_passed_at_this_tip() {
	local this_sha=$1
	[ -n "${GITHUB_RUN_ID:-}" ] || return 1 # local run: no history, always replay
	local past
	past=$(gh run list --workflow watch --status completed --json databaseId,headSha --limit 20 2>/dev/null) || return 1
	local pid
	pid=$(GH_JSON="$past" SHA="$this_sha" THIS="$GITHUB_RUN_ID" python3 - <<'PY'
import json, os
runs = json.loads(os.environ.get("GH_JSON") or "[]")
sha = os.environ["SHA"]
this = os.environ["THIS"]
for r in runs:
    if str(r["databaseId"]) == this:
        continue
    if r.get("headSha") == sha:
        print(r["databaseId"])
        break
PY
	)
	[ -n "$pid" ] || return 1
	local step
	step=$(gh run view "$pid" --json jobs --jq '[.jobs[].steps[] | select(.name == "audit-red-watch (fleet-audit rot replay, per-report issue)") | .conclusion] | first // empty' 2>/dev/null) || return 1
	[ "$step" = "success" ]
}

open_audit_issues() {
	# LIST api over the audit-red label, filtered client-side by the
	# title prefix for THIS report. NEVER the search index (it lags).
	gh issue list --state open --label audit-red --limit 200 --json number,title 2>/dev/null |
		filter_audit_titles "$1" || echo "[]"
}

# The exact-title filter (pure; exported for the self-test). Only a
# FULL title match counts — "audit-verify red: <report>" — so a
# partially-named or old-form issue never eats the signal.
filter_audit_titles() {
	REPORT="$1" python3 -c '
import json, os, sys
report = os.environ["REPORT"]
want = f"audit-verify red: {report}"
print(json.dumps([i["number"] for i in json.load(sys.stdin) if i["title"] == want]))
'
}

render_body() {
	local report=$1 summary=$2 tail=$3 date=$4
	cat <<EOF
## audit-verify red: $report

The watch lane's replay of the fleet-audit reports failed for this
report — a verify line's claim no longer holds against the tree. The
fix lives in the REPORT (the tree moved) or the TREE (the claim broke),
never by editing the script (docs/lessons.md § audit-verify self-test).

### Summary
$summary

### Failing commands
$tail

### Draft ledger row (owner fills the verdict)
| $date | audit-verify red on \`$report\` | TBD — owner verdict; row drafted by audit-red-watch |

<sub>drafted by scripts/audit-red-watch.sh — evidence assembly only, verdict is human</sub>
EOF
}

extract_failing_reports() {
	# audit-verify.sh prints "FAIL <cmd>" / per-report verdict lines;
	# its last line is the tally. Failing REPORT names come from the
	# per-report blocks: lines like "report: <name>.md". Everything is
	# best effort — the issue body carries the raw tail regardless.
	sed 's/\x1b\[[0-9;]*m//g' |
		grep -E "^FAIL|failed|✗" | head -10 || true
}

self_test() {
	local checks=0
	local failed=0
	failed=0
	assert() {
		checks=$((checks + 1))
		if [ "$1" != "$2" ]; then
			failed=$((failed + 1))
			echo "  FAIL  $3 (wanted '$1', got '$2')"
		else
			echo "  PASS  $3"
		fi
	}
	assert "SKIP same-tip-verified go-etag.md" "$(decide "go-etag.md" "[]" "1")" "same-tip prior pass -> SKIP"
	assert "CREATE go-etag.md" "$(decide "go-etag.md" "[]" "0")" "fresh failure -> CREATE"
	assert "EXISTS 7" "$(decide "go-etag.md" "[7]" "0")" "issue already tracks this report -> EXISTS"
	# The dedup lane filters by the EXACT title from the LIST payload
	# (the SEARCH index lags — the 2026-11-04 199-duplicate class).
	local lane
	lane='[{"number":8,"title":"audit-verify red: go-etag.md"},{"number":9,"title":"audit-verify red: go-health.md"},{"number":10,"title":"some other audit issue"}]'
	assert "[8]" "$(printf '%s' "$lane" | filter_audit_titles "go-etag.md")" "exact-title filter keeps only this report's issue"
	assert "[]" "$(printf '%s' "$lane" | filter_audit_titles "templ-components.md")" "no issue for an untracked report"
	echo "audit-red-watch: self-test $checks checks, $failed failed"
	[ "$failed" -eq 0 ]
}

if [ "${self_test_mode:-0}" = "1" ]; then
	self_test
	exit 0
fi

# A local harness for the open-issue lane shape (the real lane needs gh).
open_audit_issues_for_test() {
	echo "[]"
}
# ------------------------------------------------------------- the run
if [ ! -x scripts/audit-verify.sh ]; then
	say "scripts/audit-verify.sh missing — nothing to replay"
	exit 0
fi

THIS_SHA=$(git rev-parse HEAD)
if prior_audit_passed_at_this_tip "$THIS_SHA"; then
	say "SKIP: the previous watch run already replayed green at this tip ($THIS_SHA)"
	exit 0
fi

say "replaying the fleet-audit verify blocks at $THIS_SHA"
LOG=$(mktemp)
if bash scripts/audit-verify.sh >"$LOG" 2>&1; then
	say "NONE: all verify claims hold ($(tail -1 "$LOG"))"
	rm -f "$LOG"
	exit 0
fi

SUMMARY=$(sed 's/\x1b\[[0-9;]*m//g' "$LOG" | tail -5 | tr '\n' ' ')
TAIL=$(extract_failing_reports <"$LOG" | head -10)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# The failing reports: audit-verify exits nonzero listing per-report
# blocks; derive the report set from the log's own "report" markers,
# falling back to ONE umbrella issue when the marker form changed.
REPORTS=$(sed 's/\x1b\[[0-9;]*m//g' "$LOG" | grep -oE '^report: [^ ]+\.md' | awk '{print $2}' | sort -u)
if [ -z "$REPORTS" ]; then
	REPORTS="docs/review/libs/index.html"
fi

rm -f "$LOG"

for report in $REPORTS; do
	name=$(basename "$report")
	same=$(open_audit_issues "$name")
	verdict=$(decide "$name" "$same" "0")
	case "$verdict" in
	SKIP*) continue ;;
	EXISTS\ *)
		number=${verdict#EXISTS }
		say "EXISTS: issue #$number already tracks $name"
		if [ "$DRY_RUN" = "0" ]; then
			gh issue comment "$number" --body "recurrence at $DATE ($THIS_SHA): $SUMMARY" >/dev/null 2>&1 || true
		fi
		;;
	CREATE\ *)
		name=${verdict#CREATE }
		say "CREATE: filing the audit-verify red issue for $name"
		if [ "$DRY_RUN" = "0" ]; then
			gh label create audit-red --color D93F0B 2>/dev/null || true
			body=$(render_body "$name" "$SUMMARY" "$TAIL" "$DATE")
			gh issue create \
				--title "audit-verify red: $name" \
				--label audit-red \
				--body "$body" >/dev/null 2>&1 || say "WARN: issue create failed (the signal stays in this log)"
		fi
		;;
	esac
done

say "done"
exit 0
