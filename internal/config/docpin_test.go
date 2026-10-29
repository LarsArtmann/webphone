package config

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// README.md's Configuration table is the operator's only list of knobs:
// a renamed or deleted key leaves it lying, and nothing but a reader's
// attention keeps the two in step. These tests do it instead (the
// pattern ai-phone's tests/unit/test_docs_match_code.py established).

// configLeafPaths walks the Config tree by its koanf tags. Maps and
// slices are leaves: README documents `auth.passkey.users` as one map
// with an inline shape, and `ice_servers` as one list.
func configLeafPaths(t *testing.T) []string {
	t.Helper()
	var out []string
	var walk func(rt reflect.Type, prefix string)
	walk = func(rt reflect.Type, prefix string) {
		for i := range rt.NumField() {
			field := rt.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := strings.Split(field.Tag.Get("koanf"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			path := name
			if prefix != "" {
				path = prefix + "." + name
			}
			if field.Type.Kind() == reflect.Struct {
				walk(field.Type, path)
				continue
			}
			out = append(out, path)
		}
	}
	walk(reflect.TypeOf(Config{}), "")
	sort.Strings(out)
	return out
}

// readmeConfigTable parses README.md's "## Configuration" table into
// key -> default cell. The first cell is the koanf path, the second the
// documented default.
func readmeConfigTable(t *testing.T) map[string]string {
	t.Helper()
	rows := readmeConfigRows(t)
	out := make(map[string]string, len(rows))
	for _, cells := range rows {
		out[cells[0]] = cells[1]
	}
	return out
}

// readmeConfigRows returns the non-header cells of every table row in
// README.md's "## Configuration" section.
func readmeConfigRows(t *testing.T) [][]string {
	t.Helper()
	data, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	var rows [][]string
	inSection := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			inSection = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Configuration"
			continue
		}
		if !inSection || !strings.HasPrefix(line, "| `") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			continue
		}
		key := strings.Trim(strings.TrimSpace(cells[1]), "`")
		if key == "" {
			continue
		}
		rows = append(rows, []string{key, strings.TrimSpace(cells[2])})
	}
	if len(rows) == 0 {
		t.Fatal("README.md: no Configuration table rows found, the table moved or changed shape")
	}
	return rows
}

// docFiles are the documents a config key may be described in. The
// operator-facing README table is the primary home; AGENTS.md carries
// the module-facing view of options whose full shape lives in the NixOS
// module.
var docFiles = []string{"../../README.md", "../../AGENTS.md"}

// undocumentedByDesign lists leaf keys that deliberately appear in no
// operator doc, each with the reason. Empty means every key is written
// down somewhere a reader can find it: a new knob either gets a line or
// an entry here.
var undocumentedByDesign = map[string]string{}

func TestREADMEConfigTableNamesOnlyRealKeys(t *testing.T) {
	real := map[string]bool{}
	for _, path := range configLeafPaths(t) {
		real[path] = true
	}
	for key := range readmeConfigTable(t) {
		if !real[key] {
			t.Errorf("README.md Configuration table names %q, which is not a config key (renamed? the key list is derived from the struct's koanf tags, so update the table)", key)
		}
	}
}

func TestEveryConfigKeyIsDocumented(t *testing.T) {
	docs := make([]string, 0, len(docFiles))
	for _, name := range docFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		docs = append(docs, string(data))
	}
	for _, path := range configLeafPaths(t) {
		if _, ok := undocumentedByDesign[path]; ok {
			continue
		}
		documented := false
		for _, body := range docs {
			if strings.Contains(body, path) {
				documented = true
				break
			}
		}
		if !documented {
			t.Errorf("config key %q appears in no operator doc (%s): document it or add it to undocumentedByDesign with a reason", path, strings.Join(docFiles, ", "))
		}
	}
}

// The documented defaults are claims about code. Pin the ones this
// package owns: config.defaults() and the turn_rest.ttl literal Load
// seeds. Provider-specific fallbacks (asr.location, the OAuth2 client
// id, the passkey display name) are applied by their consumers, so
// their README cells stay prose.
func TestREADMEConfigTableDefaultsMatchCode(t *testing.T) {
	cfg := defaults()
	want := map[string]string{
		"addr":            cfg.Addr,
		"websocket_path":  cfg.WebsocketPath,
		"session_ttl":     cfg.SessionTTL.String(),
		"session_max_ttl": cfg.SessionMaxTTL.String(),
		"turn_rest.ttl":   (48 * time.Hour).String(),
		"gateway.mode":    string(cfg.Gateway.Mode),
	}
	documented := readmeConfigTable(t)
	keys := make([]string, 0, len(want))
	for key := range want {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		cell, ok := documented[key]
		if !ok {
			t.Errorf("README.md Configuration table has no row for %q", key)
			continue
		}
		if !sameDefault(cell, want[key]) {
			t.Errorf("README.md default for %q is %q, code says %q", key, cell, want[key])
		}
	}
}

// sameDefault compares a README default cell with a code value. The
// cells are prose: `7d` and `48h` for durations, backticked tokens
// elsewhere, `_empty_` for the zero value.
func sameDefault(cell, want string) bool {
	if strings.TrimSpace(cell) == "_empty_" || strings.TrimSpace(cell) == "_required_" {
		return want == ""
	}
	cell = strings.TrimSpace(strings.Trim(strings.TrimSpace(cell), "`"))
	if cell == want {
		return true
	}
	cellDuration, ok := parseDocDuration(cell)
	if !ok {
		return false
	}
	wantDuration, err := time.ParseDuration(want)
	return err == nil && cellDuration == wantDuration
}

// parseDocDuration accepts `d` day units (the table writes `7d`/`30d`);
// time.ParseDuration does not.
func parseDocDuration(s string) (time.Duration, bool) {
	if days, found := strings.CutSuffix(s, "d"); found {
		d, err := time.ParseDuration(days + "h")
		return d * 24, err == nil
	}
	d, err := time.ParseDuration(s)
	return d, err == nil
}
