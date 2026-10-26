#!/usr/bin/env bash
# main-red-watch: the observer half of the SUPERB error-agents plan
# (M6-M7). Watches the latest CI run on main; a red run becomes an issue
# in the tracker with a DRAFT ledger row — the human owns the verdict,
# the watcher only assembles evidence (recurrence matches against the
# flake ledger and lessons, run link, failing job names).
#
# State lives in the tracker, not in CI state files: dedup is an issue
# search for `red-main run <id>`; a NEWER red while a ci-red issue is
# open appends a comment instead of filing a duplicate.
#
# Advisory: this script exits 0 unless its environment is broken. A red
# main is SIGNALED (issue/comment/stdout), never failed-on — the watch
# workflow stays green so the signal is the issue, not a second red.
#
# usage:
#   main-red-watch.sh [--dry-run]     # print the bundle, file nothing
#   main-red-watch.sh --self-test    # fixtures through the decision core
set -euo pipefail

say() { echo "main-red-watch: $*"; }

# The decision core (pure): given the latest-run JSON, the open-issue
# search JSON for this run id, and the open ci-red issue number (-1
# when none), decide NONE / EXISTS / COMMENT <n> / CREATE.
decide() {
	python3 - "$1" "$2" "$3" <<'PY'
import json, sys

runs = json.loads(sys.argv[1] or "[]")
if not runs:
    print("NONE no-runs")
    raise SystemExit
run = runs[0]
conclusion = run.get("conclusion")
if conclusion != "failure":
    print(f"NONE conclusion={conclusion}")
    raise SystemExit
rid = run["databaseId"]
same = json.loads(sys.argv[2] or "[]")
if same:
    print(f"EXISTS {same[0]['number']}")
    raise SystemExit
red = int(sys.argv[3])
if red >= 0:
    print(f"COMMENT {red}")
    raise SystemExit
print(f"CREATE {rid}")
PY
}

open_ci_red_issue() {
	gh issue list --state open --label ci-red --limit 1 --json number --jq '.[0].number' 2>/dev/null || echo "-1"
}

grep_mentions() {
	# $1 = needle; echoes "file:line: text" hits from the two ledgers.
	local needle=$1
	grep -in "$needle" docs/ci-flake-ledger.md docs/lessons.md 2>/dev/null | head -5 || true
}

render_body() {
	local run_id=$1 run_url=$2 run_title=$3 failed_jobs=$4 recurrence=$5 log_codes=$6 date=$7
	cat <<EOF
## red main — run $run_id

Run: $run_url
Failing job(s): **$failed_jobs**
Head commit: $run_title

### Draft ledger row (owner fills the verdict — the watcher only drafts)

| $run_id | $date | \`$failed_jobs\` on a push to main (head: $run_title) | TBD — owner verdict; row drafted by main-red-watch |

### Recurrence candidates (grep vs docs/ci-flake-ledger.md + docs/lessons.md)
${recurrence:-(no matches — apparently a new class)}

### Error codes seen in the failed-log tail (best effort; log access may be 403 for this token)
$log_codes

<sub>drafted by scripts/main-red-watch.sh — evidence assembly only, verdict is human</sub>
EOF
}

self_test() {
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
	local green pending red same
	green='[{"databaseId":111,"conclusion":"success"}]'
	pending='[{"databaseId":112,"conclusion":"IN_PROGRESS"}]'
	red='[{"databaseId":113,"conclusion":"failure","displayTitle":"fix: x"}]'
	same='[{"number":7,"title":"red-main run 113: build"}]'

	assert "NONE conclusion=success" "$(decide "$green" "[]" "-1")" "green main -> NONE"
	assert "NONE conclusion=IN_PROGRESS" "$(decide "$pending" "[]" "-1")" "in-flight run -> NONE"
	assert "CREATE 113" "$(decide "$red" "[]" "-1")" "fresh red, no issues -> CREATE"
	assert "EXISTS 7" "$(decide "$red" "$same" "-1")" "issue already tracks this run -> EXISTS"
	assert "COMMENT 9" "$(decide "$red" "[]" "9")" "newer red, open ci-red issue -> COMMENT"

	local body
	body=$(render_body 113 "https://example/run/113" "fix: x" "fast flake checks" "- docs/lessons.md:1: prior class" "n/a" "2026-10-26")
	assert "1" "$(grep -c 'red main — run 113' <<<"$body")" "body names the run"
	assert "1" "$(grep -c 'TBD — owner verdict' <<<"$body")" "draft row keeps the human verdict"
	assert "1" "$(grep -c 'prior class' <<<"$body")" "recurrence hits render"
	echo "main-red-watch self-test: $((checks - failed))/$checks passed"
	[ "$failed" -eq 0 ]
}

DRY_RUN=0
if [ "${1:-}" = "--self-test" ]; then
	self_test
	exit $?
fi
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

command -v gh >/dev/null 2>&1 || {
	say "ENV gh is not on PATH"
	exit 3
}
: "${GH_TOKEN:=${GITHUB_TOKEN:-}}"
GH_REPO="${GH_REPO:-LarsArtmann/webphone}"
export GH_REPO GH_TOKEN

LIST_JSON=$(gh run list --branch main --limit 1 --json databaseId,conclusion,displayTitle,headSha,url)
SAME_JSON=$(gh issue list --state open --search "in:title red-main run" --json number,title 2>/dev/null || echo "[]")
RED_NUM=$(open_ci_red_issue)
[ -n "$RED_NUM" ] || RED_NUM=-1

DECISION=$(decide "$LIST_JSON" "$SAME_JSON" "$RED_NUM")
case "$DECISION" in
NONE*)
	say "$DECISION"
	exit 0
	;;
EXISTS*)
	say "$DECISION (already tracked — the tracker is the state store)"
	exit 0
	;;
esac

RUN_ID=$(GH_REPO= python3 -c 'import json,sys; print(json.loads(sys.argv[1])[0]["databaseId"])' "$LIST_JSON")
RUN_URL=$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])[0].get("url",""))' "$LIST_JSON")
RUN_TITLE=$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])[0].get("displayTitle","")[:80])' "$LIST_JSON")
FAILED_JOBS=$(gh run view "$RUN_ID" --json jobs --jq '[.jobs[] | select(.conclusion=="failure") | .name] | join(", ")' 2>/dev/null || echo "unknown")
[ -n "$FAILED_JOBS" ] || FAILED_JOBS="unknown"

RECURRENCE=""
IFS=',' read -ra JOB_NAMES <<<"$FAILED_JOBS"
for job in "${JOB_NAMES[@]}"; do
	job=$(echo "$job" | sed 's/^ *//;s/ *$//')
	[ -n "$job" ] || continue
	while IFS= read -r hit; do
		[ -n "$hit" ] && RECURRENCE+="- $hit"$'\n'
	done < <(grep_mentions "$job")
done

LOG_CODES=""
LOG_SNIP=$(gh run view "$RUN_ID" --log-failed 2>/dev/null | grep -oE '\b[a-z]+\.[a-z_]+' | grep -vE '^(github|go|golang|com|org|md|nix|test|yaml|json|e\.g)' | sort | uniq -c | sort -rn | head -5 || true)
[ -n "$LOG_SNIP" ] && LOG_CODES="$LOG_SNIP"

DATE=$(date +%F)
BODY=$(render_body "$RUN_ID" "$RUN_URL" "$RUN_TITLE" "$FAILED_JOBS" "$RECURRENCE" "$LOG_CODES" "$DATE")

if [ "$DRY_RUN" = 1 ]; then
	say "DRY-RUN decision=$DECISION run=$RUN_ID jobs=$FAILED_JOBS"
	echo "--- issue body ---"
	echo "$BODY"
	exit 0
fi

if [ "$DECISION" = "CREATE $RUN_ID" ]; then
	gh label create ci-red --description "a main-branch CI run failed" --color D73A4E >/dev/null 2>&1 || true
	ISSUE_URL=$(gh issue create --title "red-main run $RUN_ID: $FAILED_JOBS" --body "$BODY" --label ci-red)
	say "filed $ISSUE_URL for run $RUN_ID"
else
	NUM=${DECISION#COMMENT }
	gh issue comment "$NUM" --body "$BODY" >/dev/null
	say "appended run $RUN_ID evidence to existing ci-red issue #$NUM"
fi
