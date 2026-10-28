#!/usr/bin/env bash
# Module-cache janitor: reports and (with --sweep) deletes the zero-byte
# .tmp download leftovers interrupted `go` commands leave under
# GOMODCACHE/cache/download (24 found on 2026-10-18). These never become
# anything — a download either finished (the .tmp was renamed) or died;
# a zero-byte corpse only confuses the next forensics pass. The
# wedged-download class it complements is documented in docs/lessons.md:
# hung go processes sit in locks_lock_inode_wait on the module-cache
# flock while their artifacts stay behind.
#
# Second class (2026-10-28): a zero-byte FINAL artifact under @v/ — the
# full-disk corruption that makes `go` report a checksum mismatch (which
# reads as a supply-chain attack, while sum.golang.org confirms go.sum is
# right). Reported, never swept: repairing means deleting the corpse and
# re-downloading, a deliberate act.
#
# Safety: default is REPORT-ONLY. --sweep deletes ONLY files that are
# zero-byte AND older than the age floor (default 60 min), so an
# in-flight download's .tmp is never touched. Non-zero .tmp files are
# always reported, never deleted. Deletion is `rm` on purpose: a
# zero-byte file holds nothing to recover, so the trash rule's
# data-loss concern does not apply and trashing would pile cache junk
# into the trash store.
#
# Usage: scripts/module-cache-janitor.sh [--sweep] [--min-age-minutes N] [--self-test]
# Exit codes: 0 ok (report or successful sweep/self-test), 2 setup error.
set -eu

# janitor_run <cache-dir> <min-age-minutes> <sweep 0|1>: the one report
# body, shared by the real run and the self-test fixtures.
janitor_run() {
	cache_dir="$1"
	age="$2"
	do_sweep="$3"

	# Sweepable: zero-byte, older than the age floor.
	stale="$(find "$cache_dir" -type f -name '*.tmp' -size 0 -mmin +"$age" 2>/dev/null || true)"
	stale_count="$(printf '%s' "$stale" | grep -c . || true)"
	# Guarded: zero-byte but fresh (a download may still be mid-flight).
	fresh="$(find "$cache_dir" -type f -name '*.tmp' -size 0 -mmin -"$age" 2>/dev/null || true)"
	fresh_count="$(printf '%s' "$fresh" | grep -c . || true)"
	# Wedged: non-zero .tmp older than a day — reported, never deleted
	# (its bytes are evidence of how far the download got).
	wedged="$(find "$cache_dir" -type f -name '*.tmp' ! -size 0 -mmin +1440 2>/dev/null || true)"
	wedged_count="$(printf '%s' "$wedged" | grep -c . || true)"
	# Corrupt: a zero-byte FINAL artifact (.mod/.info/.zip under @v). This
	# is the full-disk class (2026-10-28): the file was created and never
	# written, so `go` reports a checksum mismatch that reads like a
	# supply-chain attack while sum.golang.org says go.sum is right.
	# REPORT-ONLY: repairing it means deleting the corpse and re-downloading,
	# which is a deliberate act, not a sweep.
	corrupt="$(find "$cache_dir" -type f -path '*/@v/*' \( -name '*.mod' -o -name '*.info' -o -name '*.zip' \) -size 0 2>/dev/null || true)"
	corrupt_count="$(printf '%s' "$corrupt" | grep -c . || true)"

	echo "module-cache-janitor: $cache_dir — $stale_count stale zero-byte .tmp (>${age}min), $fresh_count fresh (guarded), $wedged_count wedged non-zero (report-only), $corrupt_count zero-byte module files (CORRUPT, report-only)"
	[ -z "$stale" ] || printf '%s\n' "$stale" | sed 's/^/  sweepable: /'
	[ -z "$wedged" ] || printf '%s\n' "$wedged" | sed 's/^/  wedged:   /'
	if [ -n "$corrupt" ]; then
		printf '%s\n' "$corrupt" | sed 's/^/  corrupt:  /'
		echo "  repair: delete these and run 'go mod download' (a zero-byte .mod reports as a checksum mismatch, not as a missing file)"
	fi

	if [ "$do_sweep" = 1 ]; then
		if [ -n "$stale" ]; then
			printf '%s\n' "$stale" | while IFS= read -r f; do rm -- "$f"; done
			# Self-verify the sweep before claiming success.
			left="$(find "$cache_dir" -type f -name '*.tmp' -size 0 -mmin +"$age" 2>/dev/null | grep -c . || true)"
			if [ "$left" != "0" ]; then
				echo "module-cache-janitor: sweep FAILED — $left sweepable file(s) remain" >&2
				return 1
			fi
			echo "module-cache-janitor: swept $stale_count stale zero-byte .tmp file(s)"
		else
			echo "module-cache-janitor: nothing to sweep"
		fi
	fi
	return 0
}

# self_test: fixtures in a temp cache dir prove the three classes and
# the sweep's guarded boundaries before this script claims to work.
self_test() {
	tmp="$(mktemp -d)"
	trap 'rm -rf "$tmp"' RETURN
	dl="$tmp/cache/download/x/@v"
	mkdir -p "$dl"
	: >"$dl/stale.tmp"
	touch -d '2 hours ago' "$dl/stale.tmp"
	: >"$dl/fresh.tmp"
	touch -d '1 minute ago' "$dl/fresh.tmp"
	printf 'bytes' >"$dl/wedged.tmp"
	touch -d '2 days ago' "$dl/wedged.tmp"
	: >"$dl/v1.0.0.mod"

	out="$(janitor_run "$tmp" 60 0)"
	echo "$out" | sed 's/^/  self-test: /'
	case "$out" in
	*" 1 stale zero-byte .tmp"*" 1 fresh (guarded)"*" 1 wedged non-zero"*" 1 zero-byte module files"*) ;;
	*)
		echo "module-cache-janitor: self-test FAILED at the report shape" >&2
		return 1
		;;
	esac
	janitor_run "$tmp" 60 1 >/dev/null
	if [ -e "$dl/stale.tmp" ] || [ ! -e "$dl/fresh.tmp" ] || [ ! -e "$dl/wedged.tmp" ]; then
		echo "module-cache-janitor: self-test FAILED at the sweep boundaries" >&2
		return 1
	fi
	if [ ! -e "$dl/v1.0.0.mod" ]; then
		echo "module-cache-janitor: self-test FAILED — the report-only corrupt class was deleted by the sweep" >&2
		return 1
	fi
	echo "module-cache-janitor: self-test passed (stale swept, fresh + wedged + corrupt kept)"
	return 0
}

sweep=0
min_age=60
while [ $# -gt 0 ]; do
	case "$1" in
	--sweep)
		sweep=1
		shift
		;;
	--min-age-minutes)
		min_age="$2"
		shift 2
		;;
	--self-test)
		shift
		self_test "$@"
		exit $?
		;;
	*)
		echo "usage: $0 [--sweep] [--min-age-minutes N] [--self-test]" >&2
		exit 2
		;;
	esac
done

go_bin="$(command -v go || true)"
if [ -z "$go_bin" ]; then
	go_bin="$(ls /nix/store/*-go-1.*/bin/go 2>/dev/null | head -n 1 || true)"
fi
if [ -z "$go_bin" ]; then
	echo "module-cache-janitor: no go binary found (host go missing and no nix store go)" >&2
	exit 2
fi
cache="$("$go_bin" env GOMODCACHE)"
if [ -z "$cache" ] || [ ! -d "$cache" ]; then
	echo "module-cache-janitor: GOMODCACHE '$cache' is not a directory" >&2
	exit 2
fi

janitor_run "$cache" "$min_age" "$sweep"
