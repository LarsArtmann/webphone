#!/usr/bin/env bash
# Self-healing buildflow wrapper (plan T20/T21, 2026-09-20).
#
# Why this exists: the host exports GOTOOLCHAIN=local with an older go,
# and buildflow's preflight states the CALLER's shell wins over
# project-side env — so bare `buildflow` fails every Go step, and a
# project-side env pin cannot fix it. This wrapper re-executes inside
# `nix develop -c` when the ambient go is below the go.mod floor.
#
# Gate promotion (T21): unless steps were requested explicitly (-s), the
# wrapper appends the secret and spelling scans that fast-mode skips, so
# ONE command runs the full quality story.
set -euo pipefail
cd "$(dirname "$0")/.."

floor() {
  awk '/^go [0-9]/ { print $2; exit }' go.mod
}
ambient() {
  go version 2>/dev/null | sed -n 's/^go version go\([0-9.]*\).*/\1/p'
}
below() {
  # lexical comparison is wrong in general; compare dotted numerics
  local a b IFS=.
  read -r -a a <<<"$1"
  read -r -a b <<<"$2"
  local i
  for i in 0 1 2; do
    local x=${a[i]:-0} y=${b[i]:-0}
    if ((10#$x < 10#$y)); then return 0; fi
    if ((10#$x > 10#$y)); then return 1; fi
  done
  return 1
}

RUN=(buildflow)
if [[ ! -f flake.nix || ! -f go.mod ]]; then
  exec "${RUN[@]}" "$@"
fi
if [[ -z "${WEBPHONE_BUILDFLOW_REEXEC:-}" ]] && command -v nix >/dev/null &&
  below "$(ambient)" "$(floor)"; then
  echo "ambient go $(ambient) < floor $(floor); re-executing via nix develop -c ..." >&2
  RUN=(nix develop -c buildflow)
  export WEBPHONE_BUILDFLOW_REEXEC=1
fi

EXPLICIT_STEP=false
for arg in "$@"; do
  case "$arg" in
    -s | --step) EXPLICIT_STEP=true ;;
  esac
done
if [[ "$EXPLICIT_STEP" == true ]]; then
  exec "${RUN[@]}" "$@"
fi

"${RUN[@]}" "$@"
EXTRA=()
buildflow list steps 2>/dev/null | grep -q gitleaks && EXTRA+=(-s gitleaks)
buildflow list steps 2>/dev/null | grep -q codespell && EXTRA+=(-s codespell)
if ((${#EXTRA[@]} > 0)); then
  "${RUN[@]}" "${EXTRA[@]}"
fi
