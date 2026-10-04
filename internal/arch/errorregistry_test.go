package arch

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestErrorCodeRegistryIsFresh keeps docs/error-contract.md's error-code
// registry byte-honest with the source: every errorfamily code literal in
// non-test Go files must be documented, every documented code must still
// exist in source, and the family sets must match. The codes render as
// `[family:code]` in error strings and are grepped by journal recipes and
// the stack runbook, so a rename is a contract break — this test is the
// guard the family-adoption follow-up (f35) asked for. Regenerate the
// documented block after a deliberate code change:
//
//	go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update
var updateRegistry = flag.Bool("update", false, "rewrite the generated error-code registry block in docs/error-contract.md")

const (
	registryDocPath = "../docs/error-contract.md"
	registryBegin   = "<!-- error-code-registry: BEGIN (generated block; do not edit by hand; go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update rewrites it) -->"
	registryEnd     = "<!-- error-code-registry: END -->"
)

var (
	// Matches every constructor shape: New/Newf/Wrap/Wrapf/WrapOnce/WrapOncef
	// with an explicit family argument, plus the family-fixed helpers
	// New{Family}(f?) / Wrap{Family}(f?). Group 2 holds the fixed-family
	// suffix when present; group 3 is the argument text before the code and
	// group 4 the code itself (always the first quoted string after the
	// open paren — explicit-family forms pass errorfamily.X, an identifier,
	// and family-fixed forms pass err or nothing).
	ctorRe = regexp.MustCompile(`errorfamily\.((?:WrapOnce|Wrap|New)((?:Rejection|Conflict|Transient|Corruption|Infrastructure|Orchestration))?f?)\(([^)"]*)\"([^"]+)\"`)
	// Family argument inside group 3, e.g. `errorfamily.Rejection, `.
	familyArgRe = regexp.MustCompile(`errorfamily\.([A-Z][A-Za-z]*)\s*,`)
	familyRe    = regexp.MustCompile(`^(Rejection|Conflict|Transient|Corruption|Infrastructure|Orchestration)$`)
	// P5: codes are stable `<seam>.<op>` names, lowercase snake.
	codeShapeRe = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$`)
)

type codeEntry struct {
	families map[string]bool
	site     string
}

func TestErrorCodeRegistryIsFresh(t *testing.T) {
	entries := scanErrorFamilyCodes(t)

	docBytes, err := os.ReadFile(registryDocPath)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(docBytes)
	begin := strings.Index(doc, registryBegin)
	end := strings.Index(doc, registryEnd)
	if begin < 0 || end < 0 || end < begin {
		t.Fatalf("registry markers missing or misordered in %s", registryDocPath)
	}
	existing := parseRegistryRows(doc[begin+len(registryBegin) : end])
	generated := registryBlock(entries)

	var problems []string
	for code := range entries {
		if _, ok := existing[code]; !ok {
			problems = append(problems, "source code missing from the doc registry: "+code)
		}
	}
	for code, got := range existing {
		want, ok := entries[code]
		if !ok {
			problems = append(problems, "documented code has no source constructor: "+code)
			continue
		}
		if got != registryRow(code, want) {
			problems = append(problems, "registry drift for "+code+": doc has "+got)
		}
	}

	if *updateRegistry {
		rewritten := doc[:begin+len(registryBegin)] + "\n" +
			strings.Join(generated, "\n") + "\n\n" + doc[end:]
		if err := os.WriteFile(registryDocPath, []byte(rewritten), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote the registry block in %s (%d codes)", registryDocPath, len(entries))
		return
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		t.Errorf(
			"error-code registry drifted from source (%d problems); regenerate with: go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update\n%s",
			len(problems), strings.Join(problems, "\n"),
		)
	}
}

// scanErrorFamilyCodes walks every non-test Go file and extracts the
// (code → families, first site) map, failing on literals that break the
// code contract.
func scanErrorFamilyCodes(t *testing.T) map[string]*codeEntry {
	t.Helper()
	entries := map[string]*codeEntry{}
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", ".direnv", "result":
				return filepath.SkipAll
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(path), "../")
		for _, m := range ctorRe.FindAllStringSubmatch(string(data), -1) {
			code := m[4]
			if !codeShapeRe.MatchString(code) {
				t.Errorf("%s: suspicious errorfamily code %q — want a lowercase <seam>.<op> name (a mis-parse also lands here)", rel, code)
				continue
			}
			family := m[2]
			if family == "" {
				fa := familyArgRe.FindStringSubmatch(m[3])
				if fa == nil {
					t.Errorf("%s: cannot determine the family of code %q", rel, code)
					continue
				}
				family = fa[1]
			}
			if !familyRe.MatchString(family) {
				t.Errorf("%s: unknown errorfamily %q for code %q", rel, family, code)
				continue
			}
			entry := entries[code]
			if entry == nil {
				entry = &codeEntry{families: map[string]bool{}, site: rel}
				entries[code] = entry
			}
			entry.families[family] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no errorfamily codes found — the extractor likely broke")
	}
	return entries
}

func registryBlock(entries map[string]*codeEntry) []string {
	codes := make([]string, 0, len(entries))
	for code := range entries {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	rows := []string{"| Code | Families | First site |", "| ---- | -------- | ---------- |"}
	for _, code := range codes {
		rows = append(rows, registryRow(code, entries[code]))
	}
	return rows
}

func registryRow(code string, entry *codeEntry) string {
	families := make([]string, 0, len(entry.families))
	for family := range entry.families {
		families = append(families, family)
	}
	sort.Strings(families)
	return "| " + code + " | " + strings.Join(families, ", ") + " | " + entry.site + " |"
}

// parseRegistryRows reads the generated table back: code → full row text.
func parseRegistryRows(block string) map[string]string {
	rows := map[string]string{}
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		parts := strings.Split(strings.TrimSuffix(line, "|"), "|")
		if len(parts) != 4 {
			continue
		}
		code := strings.TrimSpace(parts[1])
		families := strings.TrimSpace(parts[2])
		if code == "" || code == "Code" || strings.HasPrefix(families, "--") {
			continue
		}
		rows[code] = line
	}
	return rows
}
