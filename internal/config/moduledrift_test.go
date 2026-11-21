package config

import (
	"encoding/json/v2"
	"os"
	"strings"
	"testing"
)

// TestModuleGoldenSettingsAreKnownConfig is the module-drift audit's pin
// (pareto T17, 2026-11-21, TODO rows 171/1592): the NixOS module's
// RENDERED settings (both module-output goldens — the exact bytes the
// module-check golden case pins) are fed through the SAME unknown-key
// walk the boot WARN uses. A module writing a key this binary dropped
// would silently disable a capability at deploy time (the 2026-10-30
// fax-capability class), so the drift reds HERE, at the repo that owns
// both halves, instead of surfacing as a vanished setting in
// production. The reverse direction (config keys without module
// surface) is not drift: settings is freeform and operators may set
// anything the binary knows.
func TestModuleGoldenSettingsAreKnownConfig(t *testing.T) {
	for _, golden := range []string{
		"../../nix/testdata/module-output.golden",
		"../../nix/testdata/module-output-oauth2.golden",
	} {
		raw, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("read %s: %v (run from the package dir; the goldens are repo testdata)", golden, err)
		}
		settings := goldenSettingsJSON(t, golden, raw)
		if unknown := unknownConfigPaths(settings); len(unknown) > 0 {
			t.Fatalf("%s renders settings keys this binary IGNORES (the boot WARN class — a module/binary drift): %s", golden, strings.Join(unknown, ", "))
		}
	}
}

// goldenSettingsJSON extracts the "== settings JSON ==" section from a
// module-output golden and decodes it as the raw map the koanf merge
// would carry.
func goldenSettingsJSON(t *testing.T, name string, raw []byte) map[string]any {
	t.Helper()
	const marker = "== settings JSON =="
	start := strings.Index(string(raw), marker)
	if start < 0 {
		t.Fatalf("%s carries no %q section — golden format drift; update this test with it", name, marker)
	}
	rest := string(raw[start+len(marker):])
	if end := strings.Index(rest, "\n== "); end >= 0 {
		rest = rest[:end]
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(rest)), &settings); err != nil {
		t.Fatalf("%s: settings section is not JSON: %v", name, err)
	}
	return settings
}
