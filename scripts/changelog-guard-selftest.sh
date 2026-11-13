#!/usr/bin/env bash
# changelog-guard-selftest.sh — proves scripts/changelog-guard.sh can FAIL.
#
# The guard exists because the CHANGELOG bar had no teeth: the metaengine
# per-group-cap train landed with no [Unreleased] line and every gate stayed
# green. A guard that cannot go red would repeat exactly that failure, so this
# pins all four behaviours against a throwaway repo — the real history is never
# touched.
#
# Precedent: scripts/pre-commit-guard-selftest.sh.
# Exit 0 = every case behaved; 1 = a regression.
set -uo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
GUARD="$ROOT/scripts/changelog-guard.sh"
[ -x "$GUARD" ] || {
	echo "selftest: FATAL: $GUARD is not executable" >&2
	exit 2
}

SANDBOX=$(mktemp -d)
trap 'rm -rf "$SANDBOX"' EXIT

export GIT_AUTHOR_NAME=selftest GIT_AUTHOR_EMAIL=selftest@example.invalid
export GIT_COMMITTER_NAME=selftest GIT_COMMITTER_EMAIL=selftest@example.invalid
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null

fail=0
note_fail() {
	echo "selftest: FAIL — $1" >&2
	fail=1
}

fresh_repo() {
	local d="$SANDBOX/$1"
	mkdir -p "$d"
	git -C "$d" init -q -b main
	printf '# CHANGELOG\n\n## [Unreleased]\n' >"$d/CHANGELOG.md"
	mkdir -p "$d/internal/store"
	printf 'package store\n' >"$d/internal/store/store.go"
	git -C "$d" add -A
	git -C "$d" commit -q -m "base"
	echo "$d"
}

# ---------------------------------------------------------------- case 1 ----
# product code without a CHANGELOG delta MUST fail.
d=$(fresh_repo code-only)
printf 'package store\n\nfunc Added() {}\n' >"$d/internal/store/store.go"
git -C "$d" commit -qam "feat: add a thing with no changelog line"
if out=$(cd "$d" && bash "$GUARD" HEAD~1..HEAD 2>&1); then
	note_fail "case 1: a code-only commit passed"
else
	echo "selftest: PASS — code-only commit reds the guard"
fi
printf '%s\n' "$out" | grep -q 'internal/store/store.go' ||
	note_fail "case 1: the failure message did not name the offending file"

# ---------------------------------------------------------------- case 2 ----
# product code WITH a CHANGELOG delta passes.
d=$(fresh_repo code-plus-changelog)
printf 'package store\n\nfunc Added() {}\n' >"$d/internal/store/store.go"
printf '## [Unreleased]\n\n- **Store: added** — readers can act on it\n' >>"$d/CHANGELOG.md"
git -C "$d" add -A
git -C "$d" commit -q -m "feat: add a thing and say so"
if (cd "$d" && bash "$GUARD" HEAD~1..HEAD >/dev/null 2>&1); then
	echo "selftest: PASS — code + changelog passes"
else
	note_fail "case 2: a documented code change was rejected"
fi

# ---------------------------------------------------------------- case 3 ----
# docs-only deltas stay changelog-silent.
d=$(fresh_repo docs-only)
mkdir -p "$d/docs/status"
printf '# Status\n' >"$d/docs/status/2026-01-01_00-00_probe.md"
git -C "$d" add -A
git -C "$d" commit -q -m "docs: a status report"
if (cd "$d" && bash "$GUARD" HEAD~1..HEAD >/dev/null 2>&1); then
	echo "selftest: PASS — a docs-only commit passes"
else
	note_fail "case 3: a docs-only commit was rejected"
fi

# ---------------------------------------------------------------- case 4 ----
# tests and tooling are exempt even though they are .go / .sh.
d=$(fresh_repo tests-and-scripts)
printf 'package store\n\nfunc TestThing() {}\n' >"$d/internal/store/store_test.go"
mkdir -p "$d/scripts"
printf '#!/usr/bin/env bash\necho hi\n' >"$d/scripts/thing.sh"
git -C "$d" add -A
git -C "$d" commit -q -m "test: add a spec and a script"
if (cd "$d" && bash "$GUARD" HEAD~1..HEAD >/dev/null 2>&1); then
	echo "selftest: PASS — tests and shepherding scripts are exempt"
else
	note_fail "case 4: a test/tooling commit was rejected"
fi

# ---------------------------------------------------------------- case 5 ----
# an unresolvable base ref fails OPEN with a WARN (shallow clone / first push).
d=$(fresh_repo missing-base)
if out=$(cd "$d" && bash "$GUARD" nosuchref..HEAD 2>&1); then
	printf '%s\n' "$out" | grep -q 'NO-OP' ||
		note_fail "case 5: pass was silent instead of warning"
	echo "selftest: PASS — an unresolvable base fails open with a warning"
else
	note_fail "case 5: an unresolvable base red the guard"
fi

if [ "$fail" -eq 0 ]; then
	echo "selftest: 5 passed — the CHANGELOG guard is load-bearing"
else
	echo "selftest: FAILED" >&2
fi
exit "$fail"
