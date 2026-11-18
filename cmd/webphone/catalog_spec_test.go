package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestRunCatalogWritesSpecifications pins the hub cross-link tail: the
// service page's specifications frontmatter points at the bare filenames
// {asyncapi.yaml,openapi.json} in the service directory, so the export
// MUST ship both files — rendered from the same pure generators behind
// `webphone contracts` and GET /openapi.json, never hand-maintained
// copies.
//
// The paths MUST be bare filenames because @eventcatalog/core resolves a
// specification through TWO resolvers: the lint/MDX resolvers join the
// FULL path onto the resource page's dirname
// (core/src/utils/resource-files.ts), while the spec viewer pages join
// only the path's BASENAME (core/src/utils/collections/util.ts
// processSpecifications -> filename). A specifications/ subdirectory
// passed the full-path resolution below and still broke the federation
// hub — the viewer never found the file ("No OpenAPI spec file found";
// the AsyncAPI page parsed empty content and reported a misleading
// "asyncapi field missing"). Both resolutions are asserted now, and the
// service directory must not carry the old specifications/ subdir either.
func TestRunCatalogWritesSpecifications(t *testing.T) {
	out := t.TempDir()

	if err := runCatalog([]string{"-o", out}); err != nil {
		t.Fatalf("runCatalog: %v", err)
	}

	serviceDir := filepath.Join(out, "eventcatalog", "services", domain.CatalogServiceID)

	asyncapi, err := os.ReadFile(filepath.Join(serviceDir, "asyncapi.yaml"))
	if err != nil {
		t.Fatalf("read the exported asyncapi spec: %v", err)
	}
	if !strings.Contains(string(asyncapi), "asyncapi: 3") {
		t.Error("the exported asyncapi spec does not look like the wire-contract render (missing `asyncapi: 3`)")
	}

	openapi, err := os.ReadFile(filepath.Join(serviceDir, "openapi.json"))
	if err != nil {
		t.Fatalf("read the exported openapi spec: %v", err)
	}
	if !strings.Contains(string(openapi), `"openapi"`) {
		t.Error("the exported openapi spec does not look like the API contract render")
	}

	servicePage, err := os.ReadFile(filepath.Join(serviceDir, "index.mdx"))
	if err != nil {
		t.Fatalf("read the exported service page: %v", err)
	}
	specPaths := []string{
		"asyncapi.yaml",
		"openapi.json",
	}
	for _, needle := range specPaths {
		if !strings.Contains(string(servicePage), "path: "+needle) {
			t.Errorf("the service page frontmatter does not reference %s", needle)
		}
		if strings.Contains(string(servicePage), filepath.Join("specifications", needle)) {
			t.Errorf("specification path %q must be a bare filename, not a subdirectory path", needle)
		}
		// Resolve like the lint/MDX resolvers do: dirname(service page) + full path.
		fullResolved := filepath.Join(filepath.Dir(filepath.Join(serviceDir, "index.mdx")), needle)
		if _, statErr := os.Stat(fullResolved); statErr != nil {
			t.Errorf("specification %q does not resolve next to the service page: %v", needle, statErr)
		}
		// Resolve like the spec viewer pages do: dirname + BASENAME of the
		// declared path (processSpecifications reduces the path to its
		// basename before the page stats it).
		basenameResolved := filepath.Join(filepath.Dir(filepath.Join(serviceDir, "index.mdx")), filepath.Base(needle))
		if _, statErr := os.Stat(basenameResolved); statErr != nil {
			t.Errorf("specification %q does not resolve under the viewer's basename join: %v", needle, statErr)
		}
	}
	// Neither the tree root nor the service directory may carry a stray
	// specifications/ dir: those are the two shapes that broke the hub
	// while looking correct locally.
	for _, stray := range []string{
		filepath.Join(out, "eventcatalog", "specifications"),
		filepath.Join(serviceDir, "specifications"),
	} {
		if _, statErr := os.Stat(stray); statErr == nil {
			t.Errorf("the export still writes a stray specifications/ dir at %s, which the spec viewer never reads", stray)
		}
	}
}
