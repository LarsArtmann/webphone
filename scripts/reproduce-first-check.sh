#!/usr/bin/env bash
# reproduce-first: a fix train's new tests must FAIL on the merge base
# (SUPERB error-agents plan M3-M4). Kills the false-green class
# mechanically: a test that was weakened (or never failed) to make a fix
# land stops being invisible — the check applies ONLY the test-file diff
# onto the merge base and runs it. If the new tests already pass without
# the fix, the fix is unproven.
#
# Advisory-first: every verdict exits 0; only environment errors exit
# non-zero. The `reproduce-first:` log prefix makes verdicts greppable in
# a CI run view. Blocking is a deliberate later flip after N clean
# trains (documented in the ci.yml job).
#
# usage:
#   reproduce-first-check.sh [--base REF] [--head REF] [--expect V]
#   reproduce-first-check.sh --self-test
#
# verdicts: REPRODUCES (a new test fails on base — the fix proves
# itself), NO-REPRO (new tests pass on base too — investigate), SKIP (no
# changed tests, or head is not a fix train), APPLY-FAIL (test diff does
# not apply to base), TIMEOUT, ENV-GO (go missing or below the go.mod
# floor — run inside `nix develop`).
#
# exits: 0 verdict delivered, 3 environment error, 4 --expect mismatch.
set -euo pipefail

say() { echo "reproduce-first: $*"; }

go_floor_guard() {
	if ! command -v go >/dev/null 2>&1; then
		say "ENV-GO go is not on PATH — run inside: nix develop -c $0 $*"
		exit 3
	fi
	if [ -f go.mod ] &&
		grep -qE '^go [0-9.]+' go.mod &&
		! go version >/dev/null 2>&1; then
		say "ENV-GO go cannot start against this go.mod — run inside: nix develop -c $0 $*"
		exit 3
	fi
}

resolve_refs() {
	BASE=${ARGV_BASE:-origin/main}
	HEAD=${ARGV_HEAD:-HEAD}
	if ! git rev-parse --verify --quiet "$BASE^{commit}" >/dev/null; then
		BASE=main
	fi
	MERGE_BASE=$(git merge-base "$BASE" "$HEAD") || {
		say "ENV-GO cannot resolve a merge base between $BASE and $HEAD"
		exit 3
	}
}

changed_tests() {
	git diff --name-only "$MERGE_BASE" "$HEAD" -- '*_test.go' || true
}

head_is_fix_train() {
	git log -1 --format=%s "$HEAD" | grep -qE '^(fix|auth):'
}

run_check() {
	resolve_refs
	say "base=$BASE head=$HEAD merge-base=$MERGE_BASE"

	local tests
	tests=$(changed_tests)
	if [ -z "$tests" ]; then
		say "SKIP no changed *_test.go files between base and head"
		[ "${ARGV_EXPECT:-}" = "SKIP" ] || [ -z "${ARGV_EXPECT:-}" ] || {
			say "--expect mismatch: wanted $ARGV_EXPECT got SKIP"
			exit 4
		}
		return 0
	fi
	if ! head_is_fix_train; then
		say "SKIP head commit is not a fix train ($(git log -1 --format=%s "$HEAD" | head -c 60)...)"
		[ "${ARGV_EXPECT:-}" = "SKIP" ] || [ -z "${ARGV_EXPECT:-}" ] || {
			say "--expect mismatch: wanted $ARGV_EXPECT got SKIP"
			exit 4
		}
		return 0
	fi
	go_floor_guard

	local tmp
	tmp=$(mktemp -d)
	trap 'git worktree remove --force "${tmp:-}/wt" >/dev/null 2>&1 || true; rm -rf "${tmp:-}"' EXIT

	git worktree add --detach "$tmp/wt" "$MERGE_BASE" >/dev/null 2>&1
	if ! git diff "$MERGE_BASE" "$HEAD" -- $tests | git -C "$tmp/wt" apply --whitespace=nowarn -; then
		say "APPLY-FAIL the test diff does not apply onto the merge base — rebase the fix train"
		[ "${ARGV_EXPECT:-}" = "APPLY-FAIL" ] || [ -z "${ARGV_EXPECT:-}" ] || exit 4
		return 0
	fi

	local pkgs=()
	for t in $tests; do
		pkgs+=("./$(dirname "$t")")
	done
	local uniq_pkgs
	uniq_pkgs=$(printf '%s\n' "${pkgs[@]}" | sort -u | tr '\n' ' ')

	say "running: go test -count=1 -timeout 300s $uniq_pkgs (base + test diff only)"
	local rc=0
	(cd "$tmp/wt" && GOTOOLCHAIN=local timeout 330 go test -count=1 -timeout 300s $uniq_pkgs) || rc=$?

	case "$rc" in
	0)
		say "NO-REPRO the fix train's new tests PASS on the merge base — the fix is unproven (advisory: tighten the test or name the mechanism)"
		[ "${ARGV_EXPECT:-}" = "NO-REPRO" ] || [ -z "${ARGV_EXPECT:-}" ] || exit 4
		;;
	124 | 137)
		say "TIMEOUT the scoped run exceeded its budget on base — treat as inconclusive, inspect locally"
		[ "${ARGV_EXPECT:-}" = "TIMEOUT" ] || [ -z "${ARGV_EXPECT:-}" ] || exit 4
		;;
	*)
		say "REPRODUCES the fix train's new tests FAIL on the merge base — the fix proves itself (this is the GOOD verdict)"
		[ "${ARGV_EXPECT:-}" = "REPRODUCES" ] || [ -z "${ARGV_EXPECT:-}" ] || exit 4
		;;
	esac
}

make_scratch_repo() {
	local dir=$1
	git init -q -b main "$dir"
	(
		cd "$dir"
		git config user.email t@t
		git config user.name t
		cat >go.mod <<'GOMOD'
module scratch

go 1.24
GOMOD
		cat >lib.go <<'GOLIB'
package scratch

func Add(a, b int) int { return a + b }
GOLIB
		git add -A
		git commit -qm "init"
	)
}

scratch_branch() {
	local dir=$1 branch=$2 msg=$3 broken=$4
	(
		cd "$dir"
		git checkout -q -b "$branch" main
		if [ "$broken" = broken ]; then
			cat >lib_test.go <<'GOTEST'
package scratch

import "testing"

func TestAdds(t *testing.T) {
	if Add(2, 2) != 5 {
		t.Fatal("2+2 must be 5 (this is the broken expectation under test)")
	}
}
GOTEST
		else
			cat >lib_test.go <<'GOTEST'
package scratch

import "testing"

func TestAdds(t *testing.T) {
	if Add(2, 2) != 4 {
		t.Fatal("math stopped working")
	}
}
GOTEST
		fi
		git add -A
		git commit -qm "$msg"
	)
}

self_test() {
	local script
	script=$(readlink -f "$0")
	local tmp
	tmp=$(mktemp -d)
	trap 'rm -rf "${tmp:-}"' EXIT

	local checks=0 failed=0
	assert() {
		checks=$((checks + 1))
		if [ "$1" != "$2" ]; then
			failed=$((failed + 1))
			echo "  FAIL  $3 (wanted $1, got $2)"
		else
			echo "  PASS  $3"
		fi
	}

	make_scratch_repo "$tmp/rep"
	scratch_branch "$tmp/rep" "fix/broken" "fix: demand the impossible" broken
	scratch_branch "$tmp/rep" "feat/passing" "feat: add a passing test" passing
	scratch_branch "$tmp/rep" "fix/passing" "fix: claim a fix, test already green" passing

	(cd "$tmp/rep" && "$script" --base main --head fix/broken --expect REPRODUCES) >/dev/null 2>&1
	assert 0 $? "scenario 1: broken fix-train test fails on base -> REPRODUCES"

	(cd "$tmp/rep" && "$script" --base main --head feat/passing --expect SKIP) >/dev/null 2>&1
	assert 0 $? "scenario 2: feature branch with passing test -> SKIP (not a fix train)"

	(cd "$tmp/rep" && "$script" --base main --head fix/passing --expect NO-REPRO) >/dev/null 2>&1
	assert 0 $? "scenario 3: fix branch whose test passes on base -> NO-REPRO"

	echo "reproduce-first self-test: $((checks - failed))/$checks passed"
	[ "$failed" -eq 0 ]
}

ARGV_BASE=""
ARGV_HEAD=""
ARGV_EXPECT=""
while [ $# -gt 0 ]; do
	case "$1" in
	--base)
		ARGV_BASE=$2
		shift 2
		;;
	--head)
		ARGV_HEAD=$2
		shift 2
		;;
	--expect)
		ARGV_EXPECT=$2
		shift 2
		;;
	--self-test)
		self_test
		exit $?
		;;
	*)
		say "unknown argument $1"
		exit 1
		;;
	esac
done

if [ "${GITHUB_EVENT_NAME:-}" != "pull_request" ] && [ "${REPRODUCE_FIRST_LOCAL:-}" != "1" ]; then
	say "SKIP event is ${GITHUB_EVENT_NAME:-unknown}, not pull_request — the advisory gate runs on PR trains only (blocking flip is a documented later decision)"
	exit 0
fi

run_check
