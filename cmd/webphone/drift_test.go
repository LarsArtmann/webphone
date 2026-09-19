package main

import (
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
	if "v"+string(match[1]) != newest {
		t.Errorf("flake.nix webphoneVersion %q != newest tag %q: bump the flake (or fold the tag), never let /version drift", match[1], newest)
	}
}
