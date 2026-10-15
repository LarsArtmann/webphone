#!/usr/bin/env bash
# Commit-msg guard selftest (M10 mu46, 2026-10-15).
#
# Why this exists: the 2026-10-15 triage proved the dep-guard hook fires
# correctly (reproduce: temp-index + daemon-pattern message => exit 1) yet
# daemon commit 89cf3b9 landed .buildflow.yml anyway — because the pma
# systemd unit overrides core.hooksPath to /var/empty via GIT_CONFIG_* env,
# which beats both repo and global config. The hook can never bind that
# invocation; the fleet fix lives in the unit (see TODO_LIST). This test
# pins the HOOK's own contract so a future edit can not silently break the
# refusal path the way the deployment did.
#
# What it does: exercises .githooks/commit-msg against a THROWAWAY index
# (GIT_INDEX_FILE — the real index is never touched), asserting:
#   1. daemon-pattern message + staged protected file  => refuse (exit 1)
#   2. daemon-pattern message + no protected files     => allow  (exit 0)
#   3. human message + staged protected file           => allow  (exit 0)
set -euo pipefail

repo="$(git rev-parse --show-toplevel)"
hook="$repo/.githooks/commit-msg"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

fail() {
	echo "commit-msg-guard-selftest: FAIL — $1" >&2
	exit 1
}

[ -f "$hook" ] || fail "hook missing at $hook"

run_case() {
	local name="$1" msg_file="$2" expect="$3"
	local got=0
	bash "$hook" "$msg_file" >/dev/null 2>&1 || got=$?
	if [ "$got" -ne "$expect" ]; then
		fail "$name: expected exit $expect, got $got"
	fi
}

# Case 1: daemon pattern + protected file staged => refuse.
msg="$tmp/daemon.msg"
printf 'chore: auto-commit 1 changed file(s) (heuristic)\n' >"$msg"
export GIT_INDEX_FILE="$tmp/index"
git read-tree HEAD
printf 'guard-test\n' >"$tmp/blob"
b=$(git hash-object -w "$tmp/blob")
git update-index --add --cacheinfo "100644,$b,.buildflow.yml"
run_case "daemon commit of protected file is refused" "$msg" 1

# Case 2: daemon pattern, nothing protected staged => allow.
git read-tree HEAD
run_case "daemon commit without protected files is allowed" "$msg" 0

# Case 3: human message + protected file staged => allow.
git update-index --add --cacheinfo "100644,$b,.buildflow.yml"
printf 'fix: real human work\n' >"$msg"
run_case "human commit of protected file is allowed" "$msg" 0

unset GIT_INDEX_FILE
echo "commit-msg-guard-selftest: 3 passed"
