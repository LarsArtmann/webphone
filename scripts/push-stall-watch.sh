#!/usr/bin/env bash
# Push-stall watchdog: exits 1 when the FIRST unpushed commit is older than
# the threshold (default 60 minutes) — AGENTS § Conventions' manual-push bar
# as a runnable check. The daemon's push leg is unobservable while stalled,
# so the oldest unpushed commit's committer timestamp is the conservative
# observable anchor (the stall fired twice on 2026-10-12). LOCAL-ONLY by
# structure (T21, 2026-10-27): unpushed commits exist only on the machine
# that owns the daemon, so GitHub-hosted CI cannot observe a stall (its
# checkout IS the remote tip — the range origin/main..HEAD is always empty
# there). Automation means a LOCAL timer next to the daemon, not
# watch.yml; for humans and sessions the exit code and stderr line ARE
# the alert (modeled 2026-10-27: a 40-min stall correctly stays silent at
# the 60-min bar and fires STALL at 30).
#
# Usage: scripts/push-stall-watch.sh [--threshold-minutes N] [--no-fetch]
# Exit codes: 0 in-sync-or-fresh, 1 STALL, 2 setup error.
set -eu

threshold=60
do_fetch=1
while [ $# -gt 0 ]; do
	case "$1" in
	--threshold-minutes)
		threshold="$2"
		shift 2
		;;
	--no-fetch)
		do_fetch=0
		shift
		;;
	*)
		echo "usage: $0 [--threshold-minutes N] [--no-fetch]" >&2
		exit 2
		;;
	esac
done

up="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)"
if [ -z "$up" ] || ! git rev-parse --verify -q "$up" >/dev/null 2>&1; then
	echo "push-stall-watch: no usable upstream for branch $(git branch --show-current)" >&2
	exit 2
fi
if [ "$do_fetch" = 1 ]; then
	git fetch -q "${up%%/*}" 2>/dev/null ||
		echo "push-stall-watch: fetch failed; judging against the stale remote ref" >&2
fi

first="$(git log "$up"..HEAD --format=%cI | tail -n 1 || true)"
if [ -z "$first" ]; then
	echo "push-stall-watch: in sync with $up"
	exit 0
fi

count="$(git log "$up"..HEAD --oneline | wc -l | tr -d ' ')"
now="$(date +%s)"
then="$(date -d "$first" +%s)"
age_min=$(((now - then) / 60))
echo "push-stall-watch: $count unpushed commit(s), oldest $age_min min ago ($first)"
if [ "$age_min" -ge "$threshold" ]; then
	echo "push-stall-watch: STALL — over the ${threshold}min bar; invoke the manual-push bar (state the anchor)" >&2
	exit 1
fi
