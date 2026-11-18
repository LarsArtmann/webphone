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
# On RED (exit 2) the script additionally BLAMES the failing tests: each
# `--- FAIL:` name is mapped to its source file and the last commit that
# touched it — the 2026-11-18 concurrent-train lesson (an hour lost
# attributing reds a per-file map answers instantly). Blame is
# informational: it never changes the exit code and a gh failure inside
# it degrades to a note.
#
# --self-test runs the verdict logic over fixture JSONs plus the blame
# extraction over a fixture log (no network).
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

extract_failed_tests() {
	# stdin: a failed-run log (--log-failed); stdout: the unique failing
	# Go test names, one per line. Pure (no network) so the self-test
	# covers it with a fixture.
	grep -oE -- '--- FAIL: [A-Za-z0-9_./]+' | awk '{print $3}' | sort -u
}

blame_test() {
	# $1: one Go test name. Prints "test -> file (last touched: sha subject)"
	# against the CURRENT checkout. Unknown tests print the lookup miss —
	# a name with no file is a signal too (renamed/deleted since the run).
	local test="$1" file last
	file=$(rg -l --type go "func ${test}\(" . 2>/dev/null | head -1)
	if [ -z "$file" ]; then
		echo "  $test -> (no declaring file in this checkout)"
		return 0
	fi
	last=$(git log -1 --format='%h %cd %s' --date=short -- "$file")
	echo "  $test -> $file${last:+ (last touched: $last)}"
}

blame_failed_tests() {
	# $1: full sha. Fetches the newest FAILED CI run's failed-step log and
	# maps every failing test to its file + last-touching commit. Degrades
	# to a note on any gh failure (blame never masks the verdict).
	local sha="$1" run_id log
	run_id=$(gh api "repos/LarsArtmann/webphone/actions/runs?head_sha=$sha&per_page=100" \
		--jq '[.workflow_runs[] | select(.name == "CI" and .conclusion == "failure")] | max_by(.created_at) | .id // empty' 2>/dev/null)
	if [ -z "$run_id" ]; then
		echo "blame: no concluded-failure CI run for this sha (pending runs cannot be blamed)"
		return 0
	fi
	log=$(gh run view "$run_id" --log-failed 2>/dev/null)
	if [ -z "$log" ]; then
		echo "blame: could not fetch run $run_id failed logs"
		return 0
	fi
	local tests
	tests=$(printf '%s\n' "$log" | extract_failed_tests)
	if [ -z "$tests" ]; then
		echo "blame: no --- FAIL: lines in the failed log (non-test red: format/lint/build?)"
		return 0
	fi
	echo "blame: failing tests -> declaring file (newest commit touching it):"
	while IFS= read -r test; do
		[ -n "$test" ] && blame_test "$test"
	done <<-EOF
		$tests
	EOF
	return 0
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

	# Blame extraction: a fixture log with noise around two FAIL lines.
	fixture_log='2026-11-18T18:00:00Z some job	go test ./...
--- FAIL: TestVoicemailChainOverProxy (12.3s)
    voicemail_test.go:402: boom
--- FAIL: TestProbeFakeMailboxVerify (3.1s)
FAIL	github.com/larsartmann/webphone/internal/server	18.5s
--- PASS: TestUnrelated (0.1s)
no test files'
	got=$(printf '%s' "$fixture_log" | extract_failed_tests | tr '\n' ' ')
	want='TestProbeFakeMailboxVerify TestVoicemailChainOverProxy '
	if [ "$got" != "$want" ]; then
		echo "self-test FAIL: blame extraction (got '$got', want '$want')" >&2
		fail=1
	else
		echo "self-test ok: blame extraction (2 names, sorted, noise dropped)"
	fi
	# Blame mapping: a REAL test in this checkout must resolve to its
	# declaring file (no network, just rg+git over the live repo).
	got=$(blame_test TestSSOOnlySessionResolvesIdentity)
	case "$got" in
	*"internal/server/sso_api_test.go"*) echo "self-test ok: blame mapping (test -> declaring file)" ;;
	*)
		echo "self-test FAIL: blame mapping (got: $got)" >&2
		fail=1
		;;
	esac
	got=$(blame_test TestNoSuchTestAnywhere)
	case "$got" in
	*"no declaring file"*) echo "self-test ok: blame mapping miss (unknown test named honestly)" ;;
	*)
		echo "self-test FAIL: blame mapping miss (got: $got)" >&2
		fail=1
		;;
	esac

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
verdict=$?
if [ "$verdict" = 2 ]; then
	blame_failed_tests "$full_sha"
fi
exit "$verdict"
