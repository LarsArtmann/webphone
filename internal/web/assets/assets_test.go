package assets

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestIslandModulesMatchImportClosure pins the fact IslandModules relies
// on: main.js's static import closure covers EVERY module under
// island/app. If someone adds a module outside the closure (or a dynamic
// import splits the graph), preloading the directory listing would fetch
// files the page may never run — this test fails so the preload strategy
// gets revisited instead of silently wasting requests.
func TestIslandModulesMatchImportClosure(t *testing.T) {
	modules := IslandModules()
	if len(modules) == 0 {
		t.Fatal("IslandModules() is empty — the island graph vanished")
	}

	// Every URL points at a real embedded file.
	for _, url := range modules {
		name := strings.TrimPrefix(url, "/assets/")
		if _, err := fs.ReadFile(embedded, name); err != nil {
			t.Errorf("IslandModules lists %s but the file is not embedded: %v", url, err)
		}
	}

	// Walk main.js's static import closure (named + side-effect imports;
	// the island uses no dynamic imports).
	fromRe := regexp.MustCompile(`from "\./([A-Za-z0-9_.-]+)"`)
	sideEffectRe := regexp.MustCompile(`import "\./([A-Za-z0-9_.-]+)"`)
	closure := map[string]bool{"main.js": true}
	var walk func(name string)
	walk = func(name string) {
		src, err := fs.ReadFile(embedded, "island/app/"+name)
		if err != nil {
			t.Fatalf("closure walk: %v", err)
		}
		for _, m := range fromRe.FindAllStringSubmatch(string(src), -1) {
			if !closure[m[1]] {
				closure[m[1]] = true
				walk(m[1])
			}
		}
		for _, m := range sideEffectRe.FindAllStringSubmatch(string(src), -1) {
			if !closure[m[1]] {
				closure[m[1]] = true
				walk(m[1])
			}
		}
	}
	walk("main.js")

	if len(closure) != len(modules) {
		t.Fatalf("main.js closure covers %d modules but IslandModules lists %d — preload strategy no longer matches the graph (closure: %v)",
			len(closure), len(modules), closure)
	}
}
