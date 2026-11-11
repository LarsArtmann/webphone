#!/usr/bin/env bash
# stale-prod-watch: the deployment-drift observer (C2 of the 2026-11-10
# SUPERB trunk plan). Compares what prod ACTUALLY serves (public
# /version) against the newest GitHub release; a prod older than the
# threshold becomes an issue in the tracker — because the lesson of the
# 2026-11-09/10 voicemail round is that a fix on main is not a fix in
# prod until /version matches, and nothing else watches that gap.
#
# State lives in the tracker, not in CI state files: dedup is an issue
# search for `stale-prod:` in the title. A NEW drift pair while an old
# stale-prod issue is open files a fresh issue (the pairs are different
# facts); closing the old one is the human's verdict.
#
# Advisory: this script exits 0 unless its environment is broken. Drift
# is SIGNALED (issue/stdout), never failed-on — the watch workflow stays
# green so the signal is the issue, not a second red. Prod unreachable
# is deliberately NOT this watcher's signal: the pbx-artmann
# trunk-watch owns reachability (ssh canary + probe).
#
# usage:
#   stale-prod-watch.sh [--dry-run]   # print the verdict, file nothing
#   stale-prod-watch.sh --self-test  # fixtures through the decision core
#
# env:
#   STALE_PROD_URL              prod /version (default the live vhost)
#   STALE_PROD_THRESHOLD_HOURS  drift age that turns into an issue (24)
set -euo pipefail

say() { echo "stale-prod-watch: $*"; }

# The decision core (pure): given the prod version string, the newest
# release tag, the release's age in hours, the threshold, and the open
# stale-prod issue search JSON, decide NONE / EXISTS <n> / CREATE.
decide() {
	python3 - "$1" "$2" "$3" "$4" "$5" <<'PY'
import json, sys

prod, tag, age, threshold, issues = sys.argv[1], sys.argv[2], float(sys.argv[3]), float(sys.argv[4]), sys.argv[5]
if not prod:
    print("NONE prod-unreachable (trunk-watch owns reachability)")
    raise SystemExit
if not tag:
    print("NONE no-releases-yet")
    raise SystemExit
if prod == tag:
    print(f"NONE current ({prod})")
    raise SystemExit
if age <= threshold:
    print(f"NONE young (release is {age:.0f}h old, threshold {threshold:.0f}h)")
    raise SystemExit
same = json.loads(issues or "[]")
for issue in same:
    title = issue.get("title", "")
    if f"prod runs {prod}" in title and tag in title:
        print(f"EXISTS {issue['number']}")
        raise SystemExit
print(f"CREATE {prod} {tag} {age:.0f}")
PY
}

# prod_version fetches the public /version JSON; empty on any failure.
prod_version() {
	python3 - "$1" <<'PY'
import json, socket, sys, urllib.request

socket.setdefaulttimeout(20)
try:
    with urllib.request.urlopen(sys.argv[1]) as response:
        print(json.loads(response.read().decode()).get("version", ""))
except Exception:
    print("")
PY
}

render_body() {
	local prod=$1 tag=$2 age=$3 threshold=$4 url=$5 release_date=$6 prod_commit=$7
	cat <<EOF
## stale prod — the deployment lags the release

| fact | value |
| ---- | ----- |
| prod serves | **$prod** ($prod_commit) |
| newest release | **$tag** ($release_date, $age h old) |
| drift | $age h > $threshold h threshold |
| prod /version | $url |

The lesson this watcher closes: **a fix on main is not a fix in prod
until /version matches** — the 2026-11-09/10 voicemail round shipped
signed and gated while prod kept serving the old binary for a day.

Fix path: the release runbook's deploy ritual — stage the stack pin,
\`nixos-rebuild switch\` on the host, verify \`/version\` matches
\`$tag\`, then close this issue with the deployed sha.

<sub>drafted by scripts/stale-prod-watch.sh — evidence only, the deploy is human-owned</sub>
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
	local same
	same='[{"number":7,"title":"stale-prod: prod runs v2.12.0, release v2.13.0 is 30h old"}]'

	assert "NONE prod-unreachable (trunk-watch owns reachability)" \
		"$(decide "" "v2.13.0" "30" "24" "[]")" "prod unreachable -> NONE"
	assert "NONE no-releases-yet" \
		"$(decide "v2.12.0" "" "30" "24" "[]")" "no releases -> NONE"
	assert "NONE current (v2.13.0)" \
		"$(decide "v2.13.0" "v2.13.0" "30" "24" "[]")" "prod current -> NONE"
	assert "NONE young (release is 2h old, threshold 24h)" \
		"$(decide "v2.12.0" "v2.13.0" "2" "24" "[]")" "young release -> NONE"
	assert "CREATE v2.12.0 v2.13.0 30" \
		"$(decide "v2.12.0" "v2.13.0" "30" "24" "[]")" "drift past threshold -> CREATE"
	assert "EXISTS 7" \
		"$(decide "v2.12.0" "v2.13.0" "30" "24" "$same")" "same pair already tracked -> EXISTS"
	assert "CREATE v2.11.0 v2.13.0 30" \
		"$(decide "v2.11.0" "v2.13.0" "30" "24" "$same")" "different pair files fresh -> CREATE"

	local body
	body=$(render_body "v2.12.0" "v2.13.0" "30" "24" "https://pbx.artmann.tech/version" "2026-11-10" "2c2cc950")
	assert "1" "$(grep -c 'a fix on main is not a fix in prod' <<<"$body")" "body carries the lesson"
	assert "1" "$(grep -c 'prod serves | \*\*v2.12.0\*\*' <<<"$body")" "body names the prod version"
	assert "1" "$(grep -c 'close this issue with the deployed sha' <<<"$body")" "body carries the fix path"
	echo "stale-prod-watch self-test: $((checks - failed))/$checks passed"
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
URL="${STALE_PROD_URL:-https://pbx.artmann.tech/version}"
THRESHOLD="${STALE_PROD_THRESHOLD_HOURS:-24}"

PROD=$(prod_version "$URL")
[ -n "$PROD" ] || say "prod unreachable at $URL — trunk-watch owns reachability, skipping"

RELEASE_JSON=$(gh release list --limit 1 --json tagName,createdAt 2>/dev/null || echo "[]")
TAG=$(python3 -c 'import json,sys; runs=json.loads(sys.argv[1] or "[]"); print(runs[0]["tagName"] if runs else "")' "$RELEASE_JSON")
RELEASE_DATE=$(python3 -c 'import json,sys; runs=json.loads(sys.argv[1] or "[]"); print(runs[0]["createdAt"] if runs else "")' "$RELEASE_JSON")
AGE=$(
	python3 - "$RELEASE_DATE" <<'PY'
import sys
from datetime import datetime, timezone

if not sys.argv[1]:
    print("-1")
    raise SystemExit
created = datetime.fromisoformat(sys.argv[1].replace("Z", "+00:00"))
print(f"{(datetime.now(timezone.utc) - created).total_seconds() / 3600:.1f}")
PY
)

SAME_JSON=$(gh issue list --state open --search "in:title stale-prod" --json number,title 2>/dev/null || echo "[]")
DECISION=$(decide "$PROD" "$TAG" "${AGE:-0}" "$THRESHOLD" "$SAME_JSON")
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

PROD_COMMIT=$(
	python3 - "$URL" <<'PY'
import json, socket, sys, urllib.request

socket.setdefaulttimeout(20)
try:
    with urllib.request.urlopen(sys.argv[1]) as response:
        print(json.loads(response.read().decode()).get("commit", "")[:12])
except Exception:
    print("unknown")
PY
)
BODY=$(render_body "$PROD" "$TAG" "${AGE%.*}" "$THRESHOLD" "$URL" "${RELEASE_DATE%%T*}" "$PROD_COMMIT")

if [ "$DRY_RUN" = "1" ]; then
	say "DRY-RUN would create: stale-prod: prod runs $PROD, release $TAG is ${AGE%.*}h old"
	say "$BODY"
	exit 0
fi

gh issue create \
	--label stale-prod \
	--title "stale-prod: prod runs $PROD, release $TAG is ${AGE%.*}h old" \
	--body "$BODY"
say "filed the stale-prod issue"
