#!/usr/bin/env bash
# Whitespace-drift staged-diff detector (round-4 D25.2, 2026-10-06;
# formatter adjudication added 2026-10-28).
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
# whitespace ignored. A staged change whose whitespace-only lines
# disappear from the -w view is EITHER hand drift (the e62fe34 class)
# OR formatter OUTPUT — prettier/gofmt reflowing a neighbouring
# construct when real content changed (the 2026-10-28 false positive:
# a legitimate commit needed --no-verify because the heuristic cannot
# tell the two apart). So a divergence is not a verdict yet: the
# diverging staged blobs are materialized into a temp tree and handed
# to the SAME treefmt config CI's format check runs. Canonical bytes
# mean the delta is formatter output => allow. Anything else (or a
# formatter that cannot run) falls back to the refusal: hand-edited
# whitespace a formatter did not canonicalize. Exit 1 names the fix;
# exit 0 is clean.
#
# `--ref <sha>` runs the same divergence + adjudication over the diff
# sha~1..sha instead of the index (daemon-audit mode, 2026-10-26) — an
# auto-commit is audited after it landed, not before. The commit tree
# is materialized via git archive, so landed formatter reflows do not
# file false drift flags either.
#
# Habit: run after `git add`, before `git commit`.
set -euo pipefail

if ! git rev-parse --git-dir >/dev/null 2>&1; then
	echo "whitespace-drift: not a git repository" >&2
	exit 2
fi

# Resolve the flake's treefmt wrapper — the exact formatter config CI's
# format check uses (no second source of truth) — and parse the raw
# treefmt binary + config out of its exec line. Prints "<raw><TAB><cfg>".
# Empty return on ANY failure: the acquittal is best-effort, the byte
# heuristic stays the fallback, so a missing nix never widens the gate.
treefmt_pair() {
	local script_dir wrapper raw cfg
	command -v nix >/dev/null 2>&1 || return 1
	script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
	wrapper="$(cd "$script_dir" && nix eval --raw .#formatter.x86_64-linux 2>/dev/null)" || return 1
	[ -f "$wrapper/bin/treefmt" ] || return 1
	raw="$(sed -n 's/^exec \([^ ]*bin\/treefmt\) .*/\1/p' "$wrapper/bin/treefmt")"
	cfg="$(sed -n 's/.*--config-file=\([^ ]*\).*/\1/p' "$wrapper/bin/treefmt")"
	[ -n "$raw" ] && [ -x "$raw" ] && [ -n "$cfg" ] && [ -f "$cfg" ] || return 1
	printf '%s\t%s\n' "$raw" "$cfg"
}

# Ask CI's own formatter whether the materialized tree at $1 holds
# canonical bytes. Exit 0 = clean (formatter output), non-zero = hand
# drift OR formatter unavailable — both mean "no acquittal".
materialized_tree_is_canonical() {
	local tmp pair raw cfg
	tmp="$1"
	pair="$(treefmt_pair)" || return 1
	raw="${pair%%$'\t'*}"
	cfg="${pair#*$'\t'}"
	(
		cd "$tmp" || exit 1
		"$raw" --config-file="$cfg" --tree-root="$tmp" -C "$tmp" \
			--walk filesystem --ci --quiet >/dev/null 2>&1
	)
}

REF=""
if [ "${1:-}" = "--ref" ]; then
	[ -n "${2:-}" ] || {
		echo "whitespace-drift: --ref needs a commit sha" >&2
		exit 2
	}
	REF=$2
	shift 2
fi

if [ -n "$REF" ]; then
	owned=$(git diff --name-only "$REF~1" "$REF" --diff-filter=ACMR |
		grep -E '\.(go|js|mjs|css|nix)$' |
		grep -v '_templ\.go$' |
		grep -v '^vendor/' || true)

	if [ -z "$owned" ]; then
		echo "whitespace-drift: --ref $REF touches no formatter-owned files (nothing to gate)"
		exit 0
	fi

	full=$(git diff "$REF~1" "$REF" -- $owned | wc -l)
	ws=$(git diff -w "$REF~1" "$REF" -- $owned | wc -l)

	if [ "$full" -ne "$ws" ]; then
		diverging=""
		for p in $owned; do
			pf=$(git diff "$REF~1" "$REF" -- "$p" | wc -l)
			pw=$(git diff -w "$REF~1" "$REF" -- "$p" | wc -l)
			if [ "$pf" -ne "$pw" ]; then diverging="$diverging $p"; fi
		done
		if [ -n "${diverging# }" ]; then
			tmp="$(mktemp -d)"
			trap 'rm -rf "$tmp"' EXIT
			git archive "$REF" | tar -x -C "$tmp"
			if materialized_tree_is_canonical "$tmp"; then
				echo "whitespace-drift: --ref $REF whitespace delta is formatter output (treefmt --ci green) — not drift"
				exit 0
			fi
		fi
		echo "whitespace-drift: --ref $REF diff is $full lines but only $ws when whitespace is ignored" >&2
		echo "  -> a landed commit carries hand-edited whitespace a formatter did not canonicalize; run: nix fmt" >&2
		exit 1
	fi
	echo "whitespace-drift: --ref $REF clean ($full lines in formatter-owned files, none whitespace-only)"
	exit 0
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
	diverging=""
	for p in $owned; do
		pf=$(git diff --cached -- "$p" | wc -l)
		pw=$(git diff --cached -w -- "$p" | wc -l)
		if [ "$pf" -ne "$pw" ]; then diverging="$diverging $p"; fi
	done
	if [ -n "${diverging# }" ]; then
		tmp="$(mktemp -d)"
		trap 'rm -rf "$tmp"' EXIT
		# shellcheck disable=SC2086
		git checkout-index --prefix="$tmp/" -- $diverging
		if materialized_tree_is_canonical "$tmp"; then
			# shellcheck disable=SC2086
			echo "whitespace-drift: clean — whitespace delta over $diverging is formatter output, not hand drift"
			exit 0
		fi
	fi
	echo "whitespace-drift: staged diff is $full lines but only $ws when whitespace is ignored" >&2
	echo "  -> hand-edited whitespace survived your formatter; run: nix fmt" >&2
	echo "  -> (oxfmt-green is NOT treefmt-green — the e62fe34 red)" >&2
	exit 1
fi

echo "whitespace-drift: clean ($full staged lines in formatter-owned files, none whitespace-only)"
