package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnknownConfigPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  map[string]any
		want []string
	}{
		{
			name: "clean config yields nothing",
			raw: map[string]any{
				"addr":    "127.0.0.1:8080",
				"fax":     map[string]any{"send": false, "receive": true},
				"gateway": map[string]any{"mode": "loopback", "webhook_url": ""},
				"identities": map[string]any{
					"1001": "+493012345678",
				},
				"auth": map[string]any{
					"passkey": map[string]any{
						"users": map[string]any{
							"a@b.c": map[string]any{"extensions": []any{"1001"}},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "unknown leaf at top level",
			raw:  map[string]any{"wat": 1, "addr": "x"},
			want: []string{"wat"},
		},
		{
			name: "unknown key inside a known section",
			raw:  map[string]any{"fax": map[string]any{"send": true, "nonsense": 1}},
			want: []string{"fax.nonsense"},
		},
		{
			name: "unknown section reported once, subtree collapsed",
			raw:  map[string]any{"future": map[string]any{"deep": map[string]any{"deeper": 1}}},
			want: []string{"future"},
		},
		{
			name: "unknown key inside a known section keeps siblings checked",
			raw: map[string]any{
				"gateway": map[string]any{"mode": "webhook", "bogus": true, "also_bogus": 2},
			},
			want: []string{"gateway.also_bogus", "gateway.bogus"},
		},
		{
			name: "dynamic-key map children are exempt",
			raw: map[string]any{
				"identities": map[string]any{"1001": "+49", "new-ext-tomorrow": "+49"},
				"auth": map[string]any{
					"passkey": map[string]any{
						"users": map[string]any{
							"new@user.example": map[string]any{"whatever": 1},
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := unknownConfigPaths(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("unknownConfigPaths() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("unknownConfigPaths() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestUnknownConfigPathsCoversRealSchema(t *testing.T) {
	t.Parallel()
	known, dynamic := schemaPaths()
	// The detection is only as good as the schema walk: pin a few
	// load-bearing paths (the fax lanes, the secret file pairs, a
	// dynamic-key map) so a refactor that changes tags reds here instead
	// of silently widening the unknown set.
	for _, path := range []string{
		"fax.send", "fax.receive",
		"gateway.webhook_secret_file", "sip.credential_secret_file",
		"crm.token_file", "asr.url", "push.fcm.credentials_file",
		"auth.passkey.users", "identities", "dial.cooldown",
	} {
		if !known[path] {
			t.Errorf("schemaPaths() is missing known path %q", path)
		}
	}
	for _, path := range []string{"identities", "auth.passkey.users"} {
		if !dynamic[path] {
			t.Errorf("schemaPaths() does not mark %q as a dynamic-key map", path)
		}
	}
}

func captureWarns(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestLoadWarnsOnUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := `{"fax": {"send": false, "nonsense": true}, "not_a_key": 1}`
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEBPHONE_CONFIG", path)
	buf := captureWarns(t)

	if _, err := Load(); err != nil {
		t.Fatalf("Load() error = %v, want nil (unknown keys warn, never fail boot)", err)
	}
	out := buf.String()
	if !strings.Contains(out, "unknown configuration keys") {
		t.Fatalf("Load() logged no unknown-key warning; output:\n%s", out)
	}
	if !strings.Contains(out, "fax.nonsense") || !strings.Contains(out, "not_a_key") {
		t.Fatalf("warning does not name the unknown paths; output:\n%s", out)
	}
	if strings.Contains(out, "nonsense=true") {
		t.Fatalf("warning leaked a config VALUE (paths only); output:\n%s", out)
	}
}

func TestLoadCleanConfigWarnsNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := `{"fax": {"send": false, "receive": false}, "gateway": {"mode": "loopback"}}`
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEBPHONE_CONFIG", path)
	buf := captureWarns(t)

	if _, err := Load(); err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if out := buf.String(); strings.Contains(out, "unknown configuration keys") {
		t.Fatalf("clean config produced a warning; output:\n%s", out)
	}
}

func TestWarnUnknownConfigKeysCapsTheList(t *testing.T) {
	t.Parallel()
	raw := map[string]any{}
	for i := 0; i < unknownKeyWarnCap+5; i++ {
		raw[strings.Repeat("x", i+1)] = i
	}
	// Direct slog capture through the exported shape: the WARN is a log
	// line, not an error, so the assertion is on the emitted text.
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	warnUnknownConfigKeys(raw)
	out := buf.String()
	if !strings.Contains(out, "total=25") {
		t.Fatalf("warning does not carry the full count; output:\n%s", out)
	}
	if !strings.Contains(out, "xx,") { // key #2 shown, key #21 (21 x's) absent
		t.Fatalf("warning does not show keys; output:\n%s", out)
	}
	if strings.Contains(out, strings.Repeat("x", 21)+",") {
		t.Fatalf("warning exceeded the cap; output:\n%s", out)
	}
}

func TestUnknownConfigPathsNeverPanicsOnOddShapes(t *testing.T) {
	t.Parallel()
	// A config file can carry arrays where the schema wants objects and
	// vice versa; detection must degrade to "no unknowns", never panic.
	got := unknownConfigPaths(map[string]any{
		"fax":     []any{"send"},
		"gateway": "loopback",
		"weird":   map[any]any{1: 2},
	})
	if len(got) != 1 || got[0] != "weird" {
		t.Fatalf("unknownConfigPaths() = %v, want [weird]", got)
	}
}
