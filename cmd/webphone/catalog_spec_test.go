package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunCatalogWritesSpecifications pins the hub cross-link tail: the
// service page's specifications frontmatter points at
// /specifications/{asyncapi.yaml,openapi.json}, so the export MUST ship
// both files — rendered from the same pure generators behind
// `webphone contracts` and GET /openapi.json, never hand-maintained
// copies.
func TestRunCatalogWritesSpecifications(t *testing.T) {
	out := t.TempDir()

	if err := runCatalog([]string{"-o", out}); err != nil {
		t.Fatalf("runCatalog: %v", err)
	}

	asyncapi, err := os.ReadFile(filepath.Join(out, "eventcatalog", "specifications", "asyncapi.yaml"))
	if err != nil {
		t.Fatalf("read the exported asyncapi spec: %v", err)
	}
	if !strings.Contains(string(asyncapi), "asyncapi: 3") {
		t.Error("the exported asyncapi spec does not look like the wire-contract render (missing `asyncapi: 3`)")
	}

	openapi, err := os.ReadFile(filepath.Join(out, "eventcatalog", "specifications", "openapi.json"))
	if err != nil {
		t.Fatalf("read the exported openapi spec: %v", err)
	}
	if !strings.Contains(string(openapi), `"openapi"`) {
		t.Error("the exported openapi spec does not look like the API contract render")
	}

	servicePage, err := os.ReadFile(filepath.Join(out, "eventcatalog", "services", "webphone", "index.mdx"))
	if err != nil {
		t.Fatalf("read the exported service page: %v", err)
	}
	for _, needle := range []string{
		"/specifications/asyncapi.yaml",
		"/specifications/openapi.json",
	} {
		if !strings.Contains(string(servicePage), needle) {
			t.Errorf("the service page frontmatter does not reference %s", needle)
		}
	}
}
