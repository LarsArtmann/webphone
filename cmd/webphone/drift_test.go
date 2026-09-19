package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestFlakeVersionMatchesNewestTag guards the /version drift class: the
// ldflags injection reads webphoneVersion from flake.nix, and a tag cut
// without the flake bump (or the reverse) ships a binary whose /version
// lies about its own build. Needs the git repo, so the nix build sandbox
// (no .git) skips it; buildflow and local runs exercise it.
func TestFlakeVersionMatchesNewestTag(t *testing.T) {
	if _, err := os.Stat("../../.git"); err != nil {
		t.Skip("no .git (nix sandbox or exported tree): drift check needs tags")
	}
	out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		t.Skipf("git describe failed (no tags?): %v", err)
	}
	newest := strings.TrimSpace(string(out))
	if !strings.HasPrefix(newest, "v") {
		t.Fatalf("newest tag %q has no v prefix", newest)
	}

	flake, err := os.ReadFile("../../flake.nix")
	if err != nil {
		t.Skipf("flake.nix not readable here: %v", err)
	}
	match := regexp.MustCompile(`webphoneVersion\s*=\s*"([^"]+)"`).FindSubmatch(flake)
	if match == nil {
		t.Fatal("webphoneVersion not found in flake.nix")
	}
	version := string(match[1])
	if "v"+version != newest {
		// Release window: the runbook bumps the flake (commit) before the
		// tag is cut, so for those minutes the version is legitimately
		// ahead. That state is real only when the fold already happened:
		// a dated CHANGELOG section for the version exists. Anything else
		// is genuine drift.
		if versionAheadOf(newest, version) && changelogHasDatedSection(version) {
			t.Skipf("release window: %s folded in CHANGELOG, tag not yet cut", version)
		}
		t.Errorf("flake.nix webphoneVersion %q != newest tag %q: bump the flake (or fold the tag), never let /version drift", version, newest)
	}
}

// versionAheadOf reports whether semver version > the v-prefixed tag.
func versionAheadOf(vPrefixedTag, version string) bool {
	nums := func(s string) [3]int {
		s = strings.TrimPrefix(s, "v")
		var out [3]int
		for i, part := range strings.SplitN(s, ".", 3) {
			fmt.Sscanf(part, "%d", &out[i])
		}
		return out
	}
	tag, ver := nums(vPrefixedTag), nums(version)
	if ver[0] != tag[0] {
		return ver[0] > tag[0]
	}
	if ver[1] != tag[1] {
		return ver[1] > tag[1]
	}
	return ver[2] > tag[2]
}

func changelogHasDatedSection(version string) bool {
	changelog, err := os.ReadFile("../../CHANGELOG.md")
	if err != nil {
		return false
	}
	re := regexp.MustCompile(`(?m)^## \[` + regexp.QuoteMeta(version) + `\] - \d{4}-\d{2}-\d{2}`)
	return re.Match(changelog)
}

func TestVersionAheadOf(t *testing.T) {
	cases := []struct {
		tag, version string
		want         bool
	}{
		{"v2.1.0", "2.2.0", true},
		{"v2.1.0", "2.1.0", false},
		{"v2.9.0", "2.8.0", false},
		{"v2.1.0", "3.0.0", true},
		{"v2.1.1", "2.2.0", true},
	}
	for _, c := range cases {
		if got := versionAheadOf(c.tag, c.version); got != c.want {
			t.Errorf("versionAheadOf(%q, %q) = %v, want %v", c.tag, c.version, got, c.want)
		}
	}
}
