#!/usr/bin/env bash
# Whitespace-drift staged-diff detector (round-4 D25.2, 2026-10-06).
#
# Why this exists: the e62fe34 incident — a Go edit landed with one
# over-indented line that oxfmt accepted but the flake check's
# treefmt/gofmt-style formatter rejected, so CI went RED on a
# whitespace-only delta no local gate had surfaced. The asymmetry:
# "the formatter I ran was green" is NOT "CI's formatter is green".
#
# Scope: formatter-OWNED files only (Go minus *_templ.go, JS/MJS, CSS,
# Nix). Markdown is deliberately excluded — it is formatter-unowned in
# this repo, so whitespace changes there (e.g. table normalization to
# natural width) are intentional by definition, never drift.
#
# What it does: compares the staged diff against the same diff with
# whitespace ignored. A staged change that is PURELY whitespace (indent
# shifts, alignment padding) disappears from the -w view, so diverging
# line counts mean hand-edited whitespace a formatter did not
# canonicalize. Exit 1 names the fix; exit 0 is clean.
#
# Habit: run after `git add`, before `git commit`.
set -euo pipefail

if ! git rev-parse --git-dir >/dev/null 2>&1; then
	echo "whitespace-drift: not a git repository" >&2
	exit 2
fi

owned=$(git diff --cached --name-only --diff-filter=ACMR |
	grep -E '\.(go|js|mjs|css|nix)$' |
	grep -v '_templ\.go$' |
	grep -v '^vendor/' || true)

if [ -z "$owned" ]; then
	echo "whitespace-drift: no formatter-owned files staged (md and friends are unowned — nothing to gate)"
	exit 0
fi

# shellcheck disable=SC2086
full=$(git diff --cached -- $owned | wc -l)
# shellcheck disable=SC2086
ws=$(git diff --cached -w -- $owned | wc -l)

if [ "$full" -ne "$ws" ]; then
	echo "whitespace-drift: staged diff is $full lines but only $ws when whitespace is ignored" >&2
	echo "  -> hand-edited whitespace survived your formatter; run: nix fmt" >&2
	echo "  -> (oxfmt-green is NOT treefmt-green — the e62fe34 red)" >&2
	exit 1
fi

echo "whitespace-drift: clean ($full staged lines in formatter-owned files, none whitespace-only)"
