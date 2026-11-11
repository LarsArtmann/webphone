#!/usr/bin/env bash
# cadence-candidates.sh — the archive-candidate scan behind the docs-gates
# cadence budget (.docs-gates.json max_unarchived).
#
# For every LIVE status report (docs/status/*.md, not archived/), run the
# two gates an archive must pass BEFORE the git mv (the 2026-10-15 lesson):
#   1. annotate check-rows --strict   (no UNTOUCHED/INCOMPLETE ledger tables)
#   2. annotate bare-rows             (0 bare data rows outside ledger tables)
# and print the passing reports OLDEST FIRST — those are the candidates a
# same-push archive can legally spend when a new report files into the
# cadence budget. DIRTY lines name the oldest reports that owe drain
# trains (the 2026-11-11 18-08 row: 102 bare rows, zero adjudication).
#
# The annotate binary is NOT the prebuilt fan-out copy (stale by design);
# the script builds the session-fresh tool from the skill's annotate.go
# into /tmp/annotool (rebuilds whenever tmpfiles wipes it). Override the
# path with ANNOTATE=<path>.
set -uo pipefail

REPO="$(git rev-parse --show-toplevel)" || exit 1
TOOL="${ANNOTATE:-/tmp/annotool/annotate}"
SRC="$HOME/.config/crush/skills/docs-health/assets/annotate.go"

if [ ! -x "$TOOL" ]; then
  mkdir -p "$(dirname "$TOOL")" || exit 1
  GOTOOLCHAIN=local GO111MODULE=off go build -o "$TOOL" "$SRC" || exit 1
fi

cd "$REPO" || exit 1
clean=0
dirty=0
for f in docs/status/*.md; do
  [ -f "$f" ] || continue
  base="$(basename "$f")"
  [ "$base" = "INDEX.md" ] && continue
  if "$TOOL" check-rows --strict "$f" >/dev/null 2>&1 &&
    "$TOOL" bare-rows "$f" >/dev/null 2>&1; then
    echo "CLEAN  $base"
    clean=$((clean + 1))
  else
    echo "DIRTY  $base"
    dirty=$((dirty + 1))
  fi
done
echo "----"
echo "$clean clean candidate(s) (archive-eligible, oldest first), $dirty dirty (owe drain trains)."
[ "$clean" -gt 0 ]
