package arch

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestExportedModelAPIHasProductionConsumers: every exported method on
// the MODEL packages' types (authmap, domain — the vocabulary layer
// everything else consumes) must be referenced by at least one
// production file OUTSIDE its declaring package. Test files do not
// count, same-package files do not count (a symbol whose only callers
// live in its own package does not need to be exported), and generated
// code does count (templ's *_templ.go calls .String() — that IS a
// production consumer).
//
// This is the Line.Display() class (2026-11-18 status report §e.4): a
// "one-home helper" that shipped with zero production consumers is
// speculative API, and nothing flagged it at gate time. A finding is
// fixed by consuming the symbol at a real surface, unexporting it when
// the concept is package-internal law, or deleting it — never by
// widening the exemption table.
//
// Exemptions are methods whose consumers are IMPLICIT by interface
// contract (encoding/json, the error interface) — no call site can
// exist by construction.
//
// The rule is deliberately scoped to the model packages; widening it
// to every internal package would drown interface implementations and
// wire DTOs in false positives (the type-safety triage's
// single-implementer seams are the documented counter-example class).
const modelAPIRoot = "." // resolved against the repo root below

// repoRoot answers this checkout's root (the test's CWD is its package
// dir, not the module root).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile))) // internal/arch/model_api_test.go -> repo root
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("%s does not look like the repo root: %v", root, err)
	}
	return root
}

var modelPackages = []string{
	"internal/authmap",
	"internal/domain",
}

var implicitInterfaceMethods = map[string]bool{
	"MarshalJSON":   true,
	"UnmarshalJSON": true,
	"Error":         true,
}

var exportedMethodRe = regexp.MustCompile(`func \([a-z] \*?([A-Z][A-Za-z0-9]*)\) ([A-Z][A-Za-z0-9]*)\(`)

func nonTestGoFiles(t *testing.T, root string, skipVendor bool) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipVendor && (d.Name() == "vendor" || d.Name() == ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

func TestExportedModelAPIHasProductionConsumers(t *testing.T) {
	root := repoRoot(t)
	declared := map[string]map[string]bool{} // method -> declaring dirs
	for _, pkg := range modelPackages {
		for _, file := range nonTestGoFiles(t, filepath.Join(root, pkg), false) {
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			for _, match := range exportedMethodRe.FindAllStringSubmatch(string(body), -1) {
				method := match[2]
				if implicitInterfaceMethods[method] {
					continue
				}
				if declared[method] == nil {
					declared[method] = map[string]bool{}
				}
				declared[method][filepath.Dir(file)] = true
			}
		}
	}
	if len(declared) == 0 {
		t.Fatal("the scan found no exported model methods — the walk is broken, not the model empty")
	}

	consumers := nonTestGoFiles(t, root, true) // the repo tree, vendor-free
	var dead []string
	for method, dirs := range declared {
		useRe := regexp.MustCompile(`\.` + regexp.QuoteMeta(method) + `\(`)
		used := false
		for _, file := range consumers {
			if dirs[filepath.Dir(file)] {
				continue // same-package uses do not justify the export
			}
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			if useRe.Match(body) {
				used = true
				break
			}
		}
		if !used {
			dead = append(dead, method+" ("+strings.Join(sortedKeys(dirs), ", ")+")")
		}
	}
	if len(dead) > 0 {
		t.Errorf("exported model API with zero production consumers (consume, unexport, or delete — never exempt):\n  %s",
			strings.Join(dead, "\n  "))
	}
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	// Deterministic order: findings must read the same on every run.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
