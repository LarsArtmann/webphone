// Package arch holds architecture tests: the import-direction rules and
// the island module-graph invariant that AGENTS.md declares, enforced by
// the ordinary test suite so violations fail every gate, not just review.
package arch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type listPackage struct {
	ImportPath string
	Imports    []string
}

func goListInternal(t *testing.T) []listPackage {
	t.Helper()
	// Run from the module root: this test file lives in internal/arch.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the module root")
	}
	cmd := exec.Command("go", "list", "-json", "./internal/...")
	cmd.Dir = filepath.Join(filepath.Dir(thisFile), "..", "..")
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	var pkgs []listPackage
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg listPackage
		if err := dec.Decode(&pkg); err != nil {
			t.Fatal(err)
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// TestDomainImportsNothingInternal keeps the domain layer pure: it may
// not depend on any sibling internal package (AGENTS.md invariant).
func TestDomainImportsNothingInternal(t *testing.T) {
	const module = "github.com/larsartmann/webphone/internal/"
	for _, pkg := range goListInternal(t) {
		if pkg.ImportPath != module+"domain" {
			continue
		}
		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, module) {
				t.Errorf("internal/domain imports %s — the domain layer must import nothing internal", imp)
			}
		}
	}
}

// TestServicesNeverImportServerOrWeb keeps the dependency direction:
// services stay usable without the HTTP layer (AGENTS.md invariant).
func TestServicesNeverImportServerOrWeb(t *testing.T) {
	const module = "github.com/larsartmann/webphone/internal/"
	servicePkgs := map[string]bool{
		"blob": true, "config": true, "fax": true, "gateway": true,
		"messaging": true, "pbx": true, "session": true, "store": true,
		"vcard": true,
	}
	for _, pkg := range goListInternal(t) {
		suffix := strings.TrimPrefix(pkg.ImportPath, module)
		if !servicePkgs[suffix] {
			continue
		}
		for _, imp := range pkg.Imports {
			if imp == module+"server" || strings.HasPrefix(imp, module+"web") {
				t.Errorf("internal/%s imports %s — services must never import the server or web layers", suffix, imp)
			}
		}
	}
}

// TestIslandModulesStayIndependent enforces the island's acyclic module
// graph: calls, ice and connection never import each other — the shared
// state lives in state.js (AGENTS.md invariant).
func TestIslandModulesStayIndependent(t *testing.T) {
	const independent = "calls ice connection"

	appDir := filepath.Join("..", "web", "assets", "island", "app")
	entries, err := os.ReadDir(appDir)
	if err != nil {
		t.Fatal(err)
	}

	importRe := regexp.MustCompile(`from\s+"\./([^"]+)"`)
	imports := map[string]map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".js") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(appDir, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range importRe.FindAllStringSubmatch(string(data), -1) {
			module := strings.TrimSuffix(match[1], ".js")
			if imports[strings.TrimSuffix(name, ".js")] == nil {
				imports[strings.TrimSuffix(name, ".js")] = map[string]bool{}
			}
			imports[strings.TrimSuffix(name, ".js")][module] = true
		}
	}

	for _, module := range strings.Fields(independent) {
		for _, forbidden := range strings.Fields(independent) {
			if module == forbidden {
				continue
			}
			if imports[module][forbidden] {
				t.Errorf("island module %s.js imports %s.js — %s must never import each other (share state via state.js)", module, forbidden, independent)
			}
		}
	}
}
