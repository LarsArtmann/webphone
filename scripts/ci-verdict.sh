#!/usr/bin/env bash
# ci-verdict.sh <sha> — the CI-WORKFLOW conclusion for one commit sha.
#
# The 2026-11-09 misread-green class: the watch workflow's scheduled runs
# carry success conclusions and flood a name-less `gh run list`, so a
# reader (or agent) greping for any green row can publish a release on a
# commit CI never built. This script names the workflow and answers only
# the CI workflow's own verdict.
#
# The 2026-11-24 page-one blind spot (T04): the 15-minute watch lane
# fills page 1 of `actions/runs?head_sha=` after ~25 h on one tip and the
# CI run slid to page 2 while the script answered NO-RUN (observed:
# `10327ed8` sat red for 31 h). fetch_runs_json now walks EVERY page
# until total_count is covered — the endpoint has no server-side
# workflow-name filter (verified live 2026-11-24: ?name=CI is ignored) —
# under a page cap that answers a greppable PAGE-CAP-EXCEEDED instead of
# looping forever.
#
# Exit codes:
#   0  the CI workflow ran for the sha and concluded success
#   1  NO CI run exists for the sha (watch-green is not CI-green)
#   2  the CI workflow ran for the sha and failed
#   3  usage error, gh failure, or the page cap tripped
#
# On RED (exit 2) the script additionally BLAMES the failing tests: each
# `--- FAIL:` name is mapped to its source file and the last commit that
# touched it — the 2026-11-18 concurrent-train lesson (an hour lost
# attributing reds a per-file map answers instantly). Blame is
# informational: it never changes the exit code and a gh failure inside
# it degrades to a note.
#
# --self-test runs the verdict logic over fixture JSONs, the blame
# extraction over a fixture log, and the paging over a fake gh serving
# multi-page fixtures (no network).
set -uo pipefail

usage() {
	echo "usage: scripts/ci-verdict.sh <sha> | --self-test" >&2
	exit 3
}

# 20 pages x 100 rows = 2000 runs ~ 20 days of the 15-minute watch lane
# pinned to one tip; beyond that the crawl stops and says so (rc 3)
# instead of paging forever.
PAGE_CAP=20

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

fetch_runs_json() {
	# $1: full sha. Prints ONE bare-JSON array of run rows
	# {id, workflowName, headSha, conclusion, createdAt} covering EVERY
	# workflow run on the sha — ALL pages of actions/runs?head_sha=, not
	# just page 1 (the 2026-11-24 blind spot above; the endpoint has no
	# server-side workflow filter). Rows carry id/createdAt so the blame
	# leg can reuse the same fetch instead of re-querying gh. Pages are
	# followed until total_count is covered (or a page comes back empty);
	# PAGE_CAP bounds the crawl. Returns 3 on gh or parse failure.
	local sha="$1" page collected=0 total=0 n=0 rows_acc=''
	for ((page = 1; page <= PAGE_CAP; page++)); do
		local envelope mapped stats
		envelope=$(gh api "repos/LarsArtmann/webphone/actions/runs?head_sha=$sha&per_page=100&page=$page" \
			--jq '{total: .total_count, runs: [.workflow_runs[] | {id: .id, workflowName: .name, headSha: .head_sha, conclusion: .conclusion, createdAt: .created_at}]}' 2>&1) || {
			echo "gh api failed (page $page): $envelope" >&2
			return 3
		}
		# First line "<total> <n>", then one mapped row per line.
		mapped=$(printf '%s' "$envelope" | python3 -c '
import json, sys
e = json.load(sys.stdin)
print(e["total"], len(e["runs"]))
for r in e["runs"]:
    print(json.dumps(r))
') || {
			echo "page $page envelope not parseable: $envelope" >&2
			return 3
		}
		stats=${mapped%%$'\n'*}
		total=${stats%% *}
		n=${stats##* }
		collected=$((collected + n))
		rows_acc+="${mapped#*$'\n'}"$'\n'
		if [ "$n" -eq 0 ] || [ "$collected" -ge "$total" ]; then
			break
		fi
	done
	if [ "$collected" -lt "$total" ]; then
		echo "PAGE-CAP-EXCEEDED: read $collected of $total runs on $sha after $PAGE_CAP pages; raise PAGE_CAP or query a fresher tip" >&2
		return 3
	fi
	printf '%s' "$rows_acc" | python3 -c '
import json, sys
print(json.dumps([json.loads(line) for line in sys.stdin if line.strip()]))'
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
	# $1: the runs JSON already fetched by fetch_runs_json (all pages) —
	# the blame leg reuses the verdict's rows instead of re-querying (and
	# re-paging) gh. Picks the newest FAILED CI run, fetches its
	# failed-step log and maps every failing test to its file +
	# last-touching commit. Degrades to a note on any gh failure (blame
	# never masks the verdict).
	local runs="$1" run_id log
	run_id=$(printf '%s\n' "$runs" | python3 -c '
import json, sys
runs = json.load(sys.stdin)
fails = [r for r in runs if r.get("workflowName") == "CI" and r.get("conclusion") == "failure"]
if fails:
    print(max(fails, key=lambda r: r.get("createdAt") or "")["id"])
')
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

	# Paging fixtures (the 2026-11-24 T04 blind spot): page 1 holds 100
	# watch-lane rows, the CI run sits on page 2 — the exact shape the old
	# one-page read answered NO-RUN for. A fake `gh` (exported function;
	# the child `bash "$self"` imports it) serves actions/runs pages and
	# the --log-failed log per FAKE_GH_MODE. No network.
	# shellcheck disable=SC2329  # invoked indirectly: exported below, imported by the child `bash "$self"`
	gh() {
		local args="$*"
		case "$args" in
		*--log-failed*)
			printf '2026-11-24T10:00:00Z job\tgo test ./...\n--- FAIL: TestSSOOnlySessionResolvesIdentity (1.0s)\nFAIL\tgithub.com/larsartmann/webphone/internal/server\t1.0s\n'
			return 0
			;;
		*actions/runs*)
			local page
			page=$(printf '%s' "$args" | sed -n 's/.*[?&]page=\([0-9]*\).*/\1/p')
			[ -n "$page" ] || page=1
			# shellcheck disable=SC2016  # the single-quoted python payload must not be expanded by bash
			FAKE_GH_PAGE="$page" python3 -c '
import json, os
mode, page = os.environ["FAKE_GH_MODE"], int(os.environ["FAKE_GH_PAGE"])
def row(i, name, conclusion):
    return {"id": 1000 + i, "workflowName": name, "headSha": "f" * 40,
            "conclusion": conclusion, "createdAt": "2026-11-23T00:00:00Z"}
if mode == "endless":  # a full watch page every time; total never covered
    runs, total = [row(i, "watch", "success") for i in range(100)], 99999
else:
    ci = "failure" if mode == "flood-red" else "success"
    pages = {1: [row(i, "watch", "success") for i in range(100)],
             2: [row(100, "CI", ci), row(101, "watch", "success")]}
    runs, total = pages.get(page, []), 102
# emulate `gh api --jq` post-projection shape: {total, runs}
print(json.dumps({"total": total, "runs": runs}))
'
			return 0
			;;
		esac
		echo "fake gh: unexpected invocation: $args" >&2
		return 1
	}
	export -f gh
	self="$0"
	paging_case() {
		local name="$1" mode="$2" want="$3" marker="${4:-}"
		local got code
		got=$(FAKE_GH_MODE="$mode" bash "$self" ffffffffffffffffffffffffffffffffffffffff 2>&1)
		code=$?
		if [ "$code" != "$want" ]; then
			echo "self-test FAIL: $name (exit $code, want $want): $got" >&2
			fail=1
		elif [ -n "$marker" ] && ! printf '%s\n' "$got" | grep -q "$marker"; then
			echo "self-test FAIL: $name (marker '$marker' missing): $got" >&2
			fail=1
		else
			echo "self-test ok: $name (exit $code${marker:+, marker present})"
		fi
	}
	paging_case "100 watch rows page 1, CI failure on page 2 => RED" flood-red 2
	paging_case "100 watch rows page 1, CI success on page 2 => GREEN" flood-green 0
	paging_case "endless watch pages => page cap, greppable rc 3" endless 3 "PAGE-CAP-EXCEEDED"
	unset -f gh

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
# runs the --commit flag misses). Query the API directly — PAGED: one
# page is not enough once the watch lane floods the listing (see the
# header note) — and map into the same JSON shape the verdict parser
# consumes.
full_sha=$(git rev-parse --verify --quiet "$sha^{commit}") || full_sha="$sha"
if ! runs=$(fetch_runs_json "$full_sha"); then
	exit 3
fi
printf '%s\n' "$runs" | verdict_from_json
verdict=$?
if [ "$verdict" = 2 ]; then
	blame_failed_tests "$runs"
fi
exit "$verdict"
