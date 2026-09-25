#!/usr/bin/env bash
# Release runbook for webphone — the dance that cut v2.0.0/v2.1.0, as one command.
#
# Usage:
#   scripts/release.sh X.Y.Z          # real release (gates run, tag pushes)
#   scripts/release.sh X.Y.Z --dry-run
#
# Preconditions you must do BEFORE calling this (runbook step 1):
#   - CHANGELOG: fold [Unreleased] into a dated "## [X.Y.Z] - YYYY-MM-DD" section
#   - FEATURES/TODO_LIST/ROADMAP synced, docs committed
#
# What this script does (runbook steps 2-8, fail-fast):
#   clean tree + main + synced with origin, tag free, CHANGELOG section exists
#   -> bump flake.nix webphoneVersion (package version AND /version ldflags)
#   -> gates: buildflow (no result cache), go test, nix flake check, smoke,
#      vulnix (runtime-closure advisory scan, network)
#   -> annotated tag, push main + tag, ls-remote verify
#   -> lychee link check (after push so tag links resolve)
#   -> stack: relock webphone input, commit, browser E2E + webphone VM test
#      + nix flake check, push
#   -> aarch64: cross-build package + island-lint check
#   -> gh release with the CHANGELOG section as notes
set -euo pipefail

VERSION="${1:-}"
DRY_RUN=0
case " ${*:2} " in *" --dry-run "*) DRY_RUN=1 ;; esac
if [ -z "$VERSION" ]; then
	echo "usage: $0 X.Y.Z [--dry-run]" >&2
	exit 2
fi
VERSION="${VERSION#v}"
TAG="v$VERSION"
REPO="$(cd "$(dirname "$0")/.." && pwd)"
STACK="${WEBPHONE_STACK_DIR:-$HOME/projects/nix-international-telephony}"

step() { printf '\n== %s ==\n' "$*"; }
run() {
	if [ "$DRY_RUN" = "1" ]; then echo "  [dry-run] $*"; else "$@"; fi
}

# Load precondition (2026-09-23 tail lesson: two E2E marker-stalls at host
# load 60-174 were load-shaped, not code-shaped; a quiet-host check is
# cheap and masks nothing, where an E2E auto-retry would mask real
# regressions — chosen as the default pending owner ratification, 02:47
# §g1). Override deliberately via WEBPHONE_RELEASE_MAX_LOAD=999.
load_gate() {
	local max="${WEBPHONE_RELEASE_MAX_LOAD:-8}"
	local one five
	read -r one five _ </proc/loadavg
	if awk -v l="$one" -v m="$max" 'BEGIN { exit !(l + 0 >= m + 0) }'; then
		echo "host load $one >= $max (5min: $five): the browser E2E and VM tests are timing-sensitive under contention. Wait for sustained quiet, or override with WEBPHONE_RELEASE_MAX_LOAD." >&2
		exit 1
	fi
	echo "ok: host load $one (< $max)"
}

# The daemon commits continuously; the gates take tens of minutes. A
# dirty tree AT TAG TIME means the tag would carry swept content nobody
# reviewed mid-release (2026-09-22: f50e825 landed between tag push and
# lychee) — fail and let the human inspect before re-running.
assert_clean_tree() {
	[ -z "$(git status --porcelain)" ] || {
		echo "tree went dirty mid-release (auto-commit daemon?): inspect git status, then re-run — the tag must not silently carry swept content" >&2
		exit 1
	}
}

cd "$REPO"

step "1/9 preconditions"
[ -z "$(git status --porcelain)" ] || {
	echo "tree not clean" >&2
	exit 1
}
[ "$(git branch --show-current)" = "main" ] || {
	echo "not on main" >&2
	exit 1
}
# Host health preflight (2026-09-24/25 outage: /run/binfmt vanished ~19:10
# and every nix build on the host died for 9+ hours with a cryptic
# 'getting attributes of path "/run/binfmt"' MID-GATE — after buildflow had
# already burned a full cycle. The kernel's binfmt_misc registration
# survives; the generation CANNOT recreate the symlink: its tmpfiles.d
# carries no binfmt rules and systemd-binfmt.service "finished OK" at the
# 22:35 reboot without creating it, so a service restart heals nothing —
# the interpreter symlink must be planted by hand (see the message).
# Failing HERE turns that into an actionable message instead of a
# mid-gate cryptic death.
if [ "$DRY_RUN" != "1" ] && [ ! -e /run/binfmt ]; then
	fix_bin="$(ls -d /nix/store/*-qemu-aarch64-binfmt-P/bin/* 2>/dev/null | head -1)"
	echo "host nix is broken: /run/binfmt is missing, so every nix build on this host fails (the generation cannot recreate it — restarting systemd-binfmt heals nothing). Fix (root): sudo mkdir -p /run/binfmt && sudo ln -s ${fix_bin:-<qemu-aarch64-binfmt-P binary>} /run/binfmt/aarch64-linux — full story + durable fix: docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md §0" >&2
	exit 1
fi
git fetch origin --tags --quiet
[ "$(git rev-parse main)" = "$(git rev-parse origin/main)" ] || {
	echo "main diverged from origin/main" >&2
	exit 1
}
RESUME_TAG=0
if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null; then
	if git ls-remote --tags origin "refs/tags/$TAG" | grep -q .; then
		echo "tag $TAG already on origin: resuming after step 5 (stack, aarch64, gh release)"
		RESUME_TAG=1
	else
		echo "tag $TAG exists locally but not on origin; delete it first (git tag -d $TAG)" >&2
		exit 1
	fi
fi
echo "ok: clean main at $(git rev-parse --short HEAD), $TAG is free"

step "2/9 fold-check: CHANGELOG has a dated $VERSION section"
grep -q "^## \[$VERSION\] " CHANGELOG.md || {
	echo "CHANGELOG has no '## [$VERSION] - date' section; fold [Unreleased] first (runbook step 1)" >&2
	exit 1
}
echo "ok"

step "3/9 version bump: flake.nix webphoneVersion = $VERSION"
if grep -q "webphoneVersion = \"$VERSION\";" flake.nix; then
	echo "already $VERSION"
else
	if [ "$DRY_RUN" = "1" ]; then
		echo "  [dry-run] sed flake.nix webphoneVersion -> $VERSION + commit 'release: $TAG'"
	else
		sed -i "s/webphoneVersion = \"[^\"]*\";/webphoneVersion = \"$VERSION\";/" flake.nix
		git add flake.nix
		git commit -m "release: $TAG"
	fi
fi

step "4/9 gates (fail-fast)"
# scripts/buildflow.sh, NOT bare buildflow (attempt-1 trap, 2026-09-24):
# the wrapper re-execs inside `nix develop -c` when the ambient go is below
# the go.mod floor — a bare call launched outside a develop shell fails all
# Go steps on the host's older GOTOOLCHAIN=local go. The wrapper also
# appends the gitleaks/codespell scans that fast mode skips (T21).
run env BUILDFLOW_NO_RESULT_CACHE=1 scripts/buildflow.sh
# Same trap class for the direct go call: webphone-smoke.py self-heals the
# same way, and `nix develop -c` is cheap here (the preflight above already
# proved nix works; the devShell is cached) — so the test suite runs on the
# go.mod-floor toolchain no matter how this script was launched.
run nix develop -c go test -count=1 ./...
run nix flake check
run python3 scripts/webphone-smoke.py
# Vulnix cadence (TODO row closed 2026-09-20): every train scans the
# RUNTIME closure. Range-matched findings against patched nixpkgs versions
# (the glibc class) trip this — triage against the locked tree's patch level
# before shipping, as documented in AGENTS.md.
run nix run .#vulnix

step "5/9 tag + push + verify"
if [ "$RESUME_TAG" = "1" ]; then
	echo "skipped: $TAG already pushed"
else
	if [ "$DRY_RUN" != "1" ]; then
		assert_clean_tree
	fi
	# Signed tags (plan T27c): the release is an artifact others may
	# clone — a GPG signature makes provenance checkable offline
	# (the daemon's sweep commits already sign, so the key exists).
	run git tag -s "$TAG" -m "webphone $TAG"
	if [ "$DRY_RUN" != "1" ]; then
		git tag -v "$TAG" >/dev/null 2>&1 || {
			echo "tag $TAG did not verify against a trusted key" >&2
			exit 1
		}
	fi
	run git push origin main "$TAG"
	if [ "$DRY_RUN" != "1" ]; then
		git ls-remote --tags origin "refs/tags/$TAG^{}" | grep -q . || {
			echo "tag $TAG not resolvable on remote (peeled object missing)" >&2
			exit 1
		}
		echo "remote has $TAG"
	fi
fi

step "6/9 link check (after push so the new tag link resolves)"
run nix run nixpkgs#lychee -- .

step "7/9 stack: relock, gates, push ($STACK)"
[ -d "$STACK" ] || {
	echo "stack dir missing: $STACK" >&2
	exit 1
}
[ -z "$(git -C "$STACK" status --porcelain)" ] || {
	echo "stack tree not clean" >&2
	exit 1
}
run bash -c "cd '$STACK' && nix flake lock --update-input webphone"
if [ "$DRY_RUN" != "1" ]; then
	git -C "$STACK" add flake.lock
	git -C "$STACK" diff --cached --quiet || git -C "$STACK" commit -m "chore: bump webphone input to $TAG"
fi
if [ "$DRY_RUN" != "1" ]; then
	load_gate
fi
run bash -c "cd '$STACK' && nix build -L .#telephony-browser"
run bash -c "cd '$STACK' && nix build -L .#checks.x86_64-linux.telephony-webphone"
run bash -c "cd '$STACK' && nix flake check"
run bash -c "cd '$STACK' && git push"

step "8/9 aarch64 cross-builds"
run nix build .#webphone --system aarch64-linux
run nix build .#checks.aarch64-linux.island-lint
if [ "$DRY_RUN" != "1" ]; then
	# Untrusted-client trap: --system is a RESTRICTED nix setting. An
	# untrusted client gets "ignoring the client-specified setting
	# 'system'" and may silently build the DEFAULT system with EXIT=0 —
	# a false-green aarch64 gate. Verify the produced ELF by its machine
	# bytes (e_machine at file offset 0x12): b700 = EM_AARCH64,
	# 3e00 = EM_X86_64. Never trust the exit code alone.
	aarch64_bin="$(nix build --no-link --print-out-paths .#webphone --system aarch64-linux)/bin/webphone"
	machine="$(od -An -tx1 -j18 -N2 "$aarch64_bin" | tr -d ' \n')"
	if [ "$machine" != "b700" ]; then
		echo "aarch64 guard: ELF machine bytes are '$machine', want 'b700' (EM_AARCH64) — the cross-build silently produced a different architecture (untrusted-client --system trap)" >&2
		exit 1
	fi
	echo "aarch64 guard: ELF machine bytes are b700 (EM_AARCH64)"
fi

step "9/9 GitHub release"
if [ "$DRY_RUN" = "1" ]; then
	echo "  [dry-run] gh release create $TAG --verify-tag --title $TAG --notes <CHANGELOG $VERSION section>"
else
	notes="$(mktemp)"
	# Literal prefix match, NOT a dynamic regex: as a regex, the heading
	# `## [2.4.0]` makes the brackets a one-char class ({2,.,4,0}) that
	# never matches the literal heading — the bug that shipped
	# v2.3.0/v2.4.0 with empty release notes (audit 2026-09-20). The
	# 2026-09-20 "fix" (escaped brackets through -v) only works on awks
	# whose string parser PRESERVES unknown escapes: host gawk 5.4.1
	# warns and DROPS the backslash, collapsing back into the class bug
	# (found 2026-09-25 pre-tag — 0 extracted lines). index() is literal
	# substring matching in every awk; == 1 anchors it to the line start.
	awk -v want="## [$VERSION]" '
    index($0, want) == 1 {on=1; next}
    /^## / && on {exit}
    on {print}
  ' CHANGELOG.md >"$notes"
	# Belt and braces: refuse to publish tag notes that came out empty.
	if [ ! -s "$notes" ]; then
		echo "extracted CHANGELOG $VERSION section is EMPTY — refusing to publish empty release notes (see the awk comment above)" >&2
		rm -f "$notes"
		exit 1
	fi
	gh release create "$TAG" --verify-tag --title "$TAG" --notes-file "$notes"
	rm -f "$notes"
fi

printf '\nrelease %s complete. Remaining manual steps: owner deploy (nixos-rebuild),\npost-deploy probes (smoke --base), pbx-artmann relock + toplevel pre-build.\n' "$TAG"
