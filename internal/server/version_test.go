package server

import (
	"runtime/debug"
	"testing"
)

func fakeBuildInfo(revision, modified, vcsTime string) *debug.BuildInfo {
	info := &debug.BuildInfo{}
	if revision != "" {
		info.Settings = append(info.Settings, debug.BuildSetting{Key: "vcs.revision", Value: revision})
	}
	if modified != "" {
		info.Settings = append(info.Settings, debug.BuildSetting{Key: "vcs.modified", Value: modified})
	}
	if vcsTime != "" {
		info.Settings = append(info.Settings, debug.BuildSetting{Key: "vcs.time", Value: vcsTime})
	}
	return info
}

// The /version enrichment resolution (T12, 2026-09-30): ldflags
// injection wins, build-info vcs settings follow, unknown builds stay
// empty — and a modified tree names itself dirty.
func TestBuildCommitResolution(t *testing.T) {
	buildCommit, buildCommitDate = "injected-sha", ""
	t.Cleanup(func() { buildCommit, buildCommitDate = "", "" })
	if got := buildCommitWith(nil, false); got != "injected-sha" {
		t.Fatalf("ldflags injection must win: got %q", got)
	}

	buildCommit = ""
	if got := buildCommitWith(fakeBuildInfo("vcs-sha", "false", ""), true); got != "vcs-sha" {
		t.Fatalf("clean vcs fallback: got %q", got)
	}
	if got := buildCommitWith(fakeBuildInfo("vcs-sha", "true", ""), true); got != "vcs-sha-dirty" {
		t.Fatalf("dirty vcs fallback: got %q", got)
	}
	if got := buildCommitWith(nil, false); got != "" {
		t.Fatalf("unknown build must stay empty: got %q", got)
	}
}

func TestBuildCommitDateResolution(t *testing.T) {
	buildCommit, buildCommitDate = "", "2026-09-30T12:00:00Z"
	t.Cleanup(func() { buildCommit, buildCommitDate = "", "" })
	if got := buildCommitDateWith(nil, false); got != "2026-09-30T12:00:00Z" {
		t.Fatalf("ldflags injection must win: got %q", got)
	}

	buildCommitDate = ""
	info := fakeBuildInfo("sha", "false", "2026-09-30T12:34:56+02:00")
	if got := buildCommitDateWith(info, true); got != "2026-09-30T10:34:56Z" {
		t.Fatalf("vcs time fallback (normalized to UTC): got %q", got)
	}
	if got := buildCommitDateWith(nil, false); got != "" {
		t.Fatalf("unknown build must stay empty: got %q", got)
	}
}
