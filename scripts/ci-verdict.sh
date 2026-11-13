#!/usr/bin/env bash
# ci-verdict.sh <sha> — the CI-WORKFLOW conclusion for one commit sha.
#
# The 2026-11-09 misread-green class: the watch workflow's scheduled runs
# carry success conclusions and flood a name-less `gh run list`, so a
# reader (or agent) greping for any green row can publish a release on a
# commit CI never built. This script names the workflow and answers only
# the CI workflow's own verdict.
#
# Exit codes:
#   0  the CI workflow ran for the sha and concluded success
#   1  NO CI run exists for the sha (watch-green is not CI-green)
#   2  the CI workflow ran for the sha and failed
#   3  usage error or gh failure
#
# --self-test runs the verdict logic over fixture JSONs (no network).
set -uo pipefail

usage() {
	echo "usage: scripts/ci-verdict.sh <sha> | --self-test" >&2
	exit 3
}

verdict_from_json() {
	# stdin: gh run list --json workflowName,headSha,conclusion JSON for ONE sha
	python3 -c '
import json, sys

runs = json.load(sys.stdin)
ci = [r for r in runs if r.get("workflowName") == "CI"]
watch = [r for r in runs if r.get("workflowName") != "CI"]

if not runs:
    print("NO-RUNS: no workflow runs recorded for this sha")
    sys.exit(1)

if not ci:
    names = ", ".join(sorted({r.get("workflowName", "?") for r in watch}))
    print("MASCARADE: no CI-WORKFLOW run for this sha; green rows were: " + names)
    print("MASCARADE: the scheduled watch workflow cannot vouch for this commit")
    sys.exit(1)

conclusions = sorted((c or "pending") for c in (r.get("conclusion") for r in ci))
if any(c != "success" for c in conclusions):
    print(f"RED: CI concluded {conclusions}")
    sys.exit(2)

print(f"GREEN: CI ({len(ci)} run(s)) concluded success")
sys.exit(0)
'
}

if [ "${1:-}" = "--self-test" ]; then
	fail=0
	check() {
		local name="$1" fixture="$2" want="$3"
		local got
		got=$(echo "$fixture" | verdict_from_json)
		local code=$?
		if [ "$code" != "$want" ]; then
			echo "self-test FAIL: $name (exit $code, want $want): $got" >&2
			fail=1
		else
			echo "self-test ok: $name (exit $code)"
		fi
	}

	check empty '[]' 1
	check watch-only '[{"workflowName":"watch","headSha":"a","conclusion":"success"},{"workflowName":"watch","headSha":"a","conclusion":"success"}]' 1
	check ci-green '[{"workflowName":"CI","headSha":"a","conclusion":"success"},{"workflowName":"watch","headSha":"a","conclusion":"success"}]' 0
	check ci-red '[{"workflowName":"CI","headSha":"a","conclusion":"failure"}]' 2
	check ci-pending '[{"workflowName":"CI","headSha":"a","conclusion":null},{"workflowName":"CI","headSha":"a","conclusion":"success"}]' 2

	if [ "$fail" = 0 ]; then
		echo "self-test: all cases pass"
		exit 0
	fi
	exit 1
fi

sha="${1:-}"
case "$sha" in
"" | -* | help) usage ;;
esac

# `gh run list --commit <sha>` answers EMPTY on this repo even when runs
# exist (verified live 2026-11-12: the API's head_sha filter returns the
# runs the --commit flag misses). Query the API directly and map into the
# same JSON shape the verdict parser consumes.
full_sha=$(git rev-parse --verify --quiet "$sha^{commit}") || full_sha="$sha"
runs=$(gh api "repos/LarsArtmann/webphone/actions/runs?head_sha=$full_sha&per_page=100" \
	--jq '{runs: [.workflow_runs[] | {workflowName: .name, headSha: .head_sha, conclusion: .conclusion}]}' 2>&1) || {
	echo "gh api failed: $runs" >&2
	exit 3
}
# The parser reads a bare array; unwrap the envelope.
echo "$runs" | python3 -c 'import json, sys; print(json.dumps(json.load(sys.stdin)["runs"]))' | verdict_from_json
