#!/usr/bin/env bash
# Self-test for format-check.sh (M21/F86, 2026-10-29).
# Pins: the script exists+is executable, names the authority derivation,
# exits 0 on a clean tree, and fails (set -e) if the flake lacks the
# check. Case 4 proves the wrapper FAILS CLOSED: a flake without the
# check derivation must exit nonzero, never silently pass.
set -uo pipefail
SCRIPT="$(cd "$(dirname "$0")" && pwd)/format-check.sh"
REPO="$(dirname "$SCRIPT")/.."
fails=0
pass() { echo "ok - $1"; }
fail() {
	echo "FAIL - $1"
	fails=$((fails + 1))
}

[ -x "$SCRIPT" ] && pass "executable" || fail "not executable"
grep -q "checks.x86_64-linux.format" "$SCRIPT" &&
	pass "names the authority derivation" ||
	fail "does not name the authority derivation"

out="$(cd "$REPO" && "$SCRIPT" 2>&1)"
rc=$?
[ $rc -eq 0 ] && pass "clean tree exits 0" || fail "clean tree exited $rc"
echo "$out" | grep -q "checks.x86_64-linux.format" &&
	pass "output names the check" ||
	fail "output missing the check name"
echo "$out" | grep -q "ONLY authority" &&
	pass "prints the authority note" ||
	fail "authority note missing"

sandbox="$(mktemp -d)"
trap 'rm -rf "$sandbox"' EXIT
printf '{ outputs = { self }: { checks.x86_64-linux.nope = self.outPath; }; }\n' >"$sandbox/flake.nix"
if (cd "$sandbox" && "$SCRIPT" >/dev/null 2>&1); then
	fail "missing check derivation must fail closed"
else
	pass "missing check derivation fails closed"
fi

echo "selftest: $fails failed"
exit $fails
