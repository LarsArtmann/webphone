package server

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"testing"
	"time"
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

// TestVersionHandlerShape pins the /version payload at the Go level:
// the smoke probe key-checks the black-box surface only. The key SET is
// the contract (version, goVersion, title always; commit, commitDate
// only when known — their presence depends on the build environment,
// so absence is not asserted). Any new key is a deliberate change.
func TestVersionHandlerShape(t *testing.T) {
	server := newTestServer(t)
	resp, err := http.Get(server.URL + "/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type: got %q, want application/json", ct)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	allowed := map[string]bool{
		"version": true, "goVersion": true, "title": true,
		"commit": true, "commitDate": true,
	}
	for key, value := range payload {
		if !allowed[key] {
			t.Errorf("unexpected key %q in /version payload — extend the contract deliberately", key)
		}
		if s, ok := value.(string); !ok || s == "" {
			t.Errorf("key %q must be a non-empty string, got %#v", key, value)
		}
	}
	for _, key := range []string{"version", "goVersion", "title"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("required key %q missing from /version payload", key)
		}
	}
	if raw, ok := payload["commitDate"]; ok {
		if _, err := time.Parse(time.RFC3339, raw.(string)); err != nil {
			t.Errorf("commitDate must be RFC 3339: %q — %v", raw, err)
		}
	}
}
