package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestRunCatalogWritesSpecifications pins the hub cross-link tail: the
// service page's specifications frontmatter points at
// specifications/{asyncapi.yaml,openapi.json}, so the export MUST ship
// both files — rendered from the same pure generators behind
// `webphone contracts` and GET /openapi.json, never hand-maintained
// copies.
//
// The files live INSIDE the service directory because that is where core
// resolves them (core/src/utils/resource-files.ts joins the resource
// page's dirname with the spec path). A tree-root specifications/ dir
// passed this test and still broke the federation hub at lint
// (`refs/file-exists`) and at render (astro ENOENT), which is why the
// last assertion below resolves every referenced path the way core does.
func TestRunCatalogWritesSpecifications(t *testing.T) {
	out := t.TempDir()

	if err := runCatalog([]string{"-o", out}); err != nil {
		t.Fatalf("runCatalog: %v", err)
	}

	asyncapi, err := os.ReadFile(filepath.Join(out, "eventcatalog", "services", domain.CatalogServiceID, "specifications", "asyncapi.yaml"))
	if err != nil {
		t.Fatalf("read the exported asyncapi spec: %v", err)
	}
	if !strings.Contains(string(asyncapi), "asyncapi: 3") {
		t.Error("the exported asyncapi spec does not look like the wire-contract render (missing `asyncapi: 3`)")
	}

	openapi, err := os.ReadFile(filepath.Join(out, "eventcatalog", "services", domain.CatalogServiceID, "specifications", "openapi.json"))
	if err != nil {
		t.Fatalf("read the exported openapi spec: %v", err)
	}
	if !strings.Contains(string(openapi), `"openapi"`) {
		t.Error("the exported openapi spec does not look like the API contract render")
	}

	servicePage, err := os.ReadFile(filepath.Join(out, "eventcatalog", "services", domain.CatalogServiceID, "index.mdx"))
	if err != nil {
		t.Fatalf("read the exported service page: %v", err)
	}
	specPaths := []string{
		"specifications/asyncapi.yaml",
		"specifications/openapi.json",
	}
	for _, needle := range specPaths {
		if !strings.Contains(string(servicePage), needle) {
			t.Errorf("the service page frontmatter does not reference %s", needle)
		}
		if strings.Contains(string(servicePage), "/"+needle) {
			t.Errorf("specification path %q must be service-dir-relative, not root-absolute", needle)
		}
		// Resolve exactly like core does: dirname(service page) + path.
		resolved := filepath.Join(
			filepath.Dir(filepath.Join(out, "eventcatalog", "services", domain.CatalogServiceID, "index.mdx")),
			needle,
		)
		if _, statErr := os.Stat(resolved); statErr != nil {
			t.Errorf("specification %q does not resolve next to the service page: %v", needle, statErr)
		}
	}
	// The tree root must NOT carry a stray specifications/ dir: that was
	// the shape that broke the hub while looking correct locally.
	if _, statErr := os.Stat(filepath.Join(out, "eventcatalog", "specifications")); statErr == nil {
		t.Error("the export still writes a tree-root specifications/ dir, which core never reads")
	}
}
