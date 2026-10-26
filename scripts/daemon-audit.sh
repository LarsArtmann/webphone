#!/usr/bin/env bash
# daemon-audit: mechanically catch the auto-commit daemon's guard-hole
# class (SUPERB error-agents M9). The commit-msg hook binds HUMAN
# commits only — a live env-level git-config override lets every daemon
# commit run WITHOUT hooks (the 2026-10-15 guard hole, incident 89cf3b9).
# This script audits the daemon's landed work instead of trusting the
# hook: a daemon-pattern commit that touches a guarded file
# (.buildflow.yml / go.mod / go.sum / go.work* / flake.lock) or carries
# whitespace drift in a formatter-owned file gets flagged.
#
# Advisory: exits 0 on every verdict; flags are SIGNALS (stdout / the
# ci-red issue lane with --file-issue), never failures — the watch
# workflow stays green, the issue is the alarm.
#
# usage:
#   daemon-audit.sh [--sha <ref>] [--dry-run] [--file-issue]
#   daemon-audit.sh --self-test
set -euo pipefail

say() { echo "daemon-audit: $*"; }
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

self_test() {
	local script_dir
	script_dir=$(cd "$(dirname "$0")" && pwd)
	local tmp
	tmp=$(mktemp -d)
	trap 'rm -rf "${tmp:-}"' EXIT

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

	git init -q -b main "$tmp/r"
	(
		cd "$tmp/r"
		git config user.email t@t
		git config user.name t
		printf 'module scratch\n\ngo 1.24\n' >go.mod
		printf 'package x\nfunc F() {\n\ty := 1\n\t_ = y\n}\n' >x.go
		git add -A
		git commit -qm "init"
		git rev-parse HEAD >"$tmp/init.sha"
		# daemon-pattern commit touching a guarded file
		printf 'module scratch\n\ngo 1.25\n' >go.mod
		git add -A
		git commit -qm "chore: auto-commit 1 changed file(s) (heuristic)"
		git rev-parse HEAD >"$tmp/gomod.sha"
		# daemon-pattern commit reindenting owned code (the e62fe34 class)
		printf 'package x\nfunc F() {\n        y := 1\n        _ = y\n}\n' >x.go
		git add -A
		git commit -qm "chore: auto-commit 2 changed file(s) (heuristic)"
	)
	local out
	out=$(cd "$tmp/r" && "$script_dir/daemon-audit.sh" --sha "$(cat "$tmp/gomod.sha")" 2>&1)
	assert "1" "$(grep -c 'guarded file touched: go.mod' <<<"$out")" "daemon commit w/ go.mod -> guarded flag"
	out=$(cd "$tmp/r" && "$script_dir/daemon-audit.sh" --sha HEAD 2>&1)
	assert "1" "$(grep -c 'whitespace drift' <<<"$out")" "daemon commit w/ reindent -> drift flag"
	out=$(cd "$tmp/r" && "$script_dir/daemon-audit.sh" --sha "$(cat "$tmp/init.sha")" 2>&1)
	assert "1" "$(grep -c 'SKIP not a daemon-pattern commit' <<<"$out")" "human commit -> SKIP"

	echo "daemon-audit self-test: $((checks - failed))/$checks passed"
	[ "$failed" -eq 0 ]
}

SHA=""
FILE_ISSUE=0
DRY_RUN=0
while [ $# -gt 0 ]; do
	case "$1" in
	--sha)
		SHA=$2
		shift 2
		;;
	--file-issue)
		FILE_ISSUE=1
		shift
		;;
	--dry-run)
		DRY_RUN=1
		shift
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
[ -n "$SHA" ] && audited_sha=$SHA
audited_sha=${audited_sha:-HEAD}
SUBJECT=$(git log -1 --format=%s "$audited_sha")

case "$SUBJECT" in
"chore: auto-commit"*) ;;
*)
	say "SKIP not a daemon-pattern commit: $(echo "$SUBJECT" | head -c 60)"
	exit 0
	;;
esac

FLAGS=""
GUARDED="\.buildflow\.yml|go\.mod|go\.sum|go\.work|go\.work\.sum|flake\.lock"
TOUCHED=$(git diff --name-only "$audited_sha~1" "$audited_sha" || true)
for f in $TOUCHED; do
	if echo "$f" | grep -qE "^($GUARDED)$"; then
		FLAGS+="- guarded file touched: $f"$'\n'
	fi
done

DRIFT_OUT=""
if ! DRIFT_OUT=$("$SCRIPT_DIR/whitespace-drift.sh" --ref "$audited_sha" 2>&1); then
	FLAGS+="- whitespace drift in the daemon's formatter-owned diff:"$'\n'"$(echo "$DRIFT_OUT" | sed 's/^/    /')"$'\n'
fi

if [ -z "$FLAGS" ]; then
	say "CLEAN $audited_sha ($SUBJECT) — guarded files untouched, no drift"
	exit 0
fi

BODY=$(cat <<EOF
## daemon-audit flag — $audited_sha

Commit: \`$SUBJECT\`

$FLAGS
The daemon runs without the commit-msg hook (the documented env-level
git-config override, guard hole 2026-10-15). A human must judge whether
this landing is legitimate — dependency bumps and buildflow config do
not ride the daemon (owner decision 2026-10-11).

<sub>drafted by scripts/daemon-audit.sh — flag only, verdict is human</sub>
EOF
)

if [ "$DRY_RUN" = 1 ]; then
	say "DRY-RUN flags on $audited_sha:"
	echo "$BODY"
	exit 0
fi
if [ "$FILE_ISSUE" = 1 ]; then
	: "${GH_TOKEN:=${GITHUB_TOKEN:-}}"
	GH_REPO="${GH_REPO:-LarsArtmann/webphone}"
	export GH_REPO GH_TOKEN
	command -v gh >/dev/null 2>&1 || {
		say "ENV gh missing — cannot file the issue; printing instead"
		echo "$BODY"
		exit 0
	}
	URL=$(gh issue create --title "daemon-audit: $audited_sha touched guarded paths" --body "$BODY" --label ci-red 2>/dev/null || gh issue create --title "daemon-audit: $audited_sha touched guarded paths" --body "$BODY")
	say "filed $URL"
fi
say "FLAGGED $audited_sha:"
echo "$FLAGS"
