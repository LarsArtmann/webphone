package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4/eventcatalog"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/server"
)

// runCatalog serves the `webphone catalog` one-shot: a pure render of
// the event catalog into <output>/eventcatalog/ — no config, no
// database, no network. The architecture-catalog hub builds this exact
// tree (the -plain variant emits plain frontmatter-ID refs for the
// hub's per-source governance lint).
func runCatalog(args []string) error {
	fs := flag.NewFlagSet("catalog", flag.ExitOnError)
	output := fs.String("o", "docs/catalog", "directory to write the catalog into")
	plain := fs.Bool("plain", false, "emit plain frontmatter-ID refs (governance lint exports)")
	_ = fs.Parse(args)

	dir := filepath.Join(*output, "eventcatalog") //nolint:branching-flow // flag.String never returns a nil pointer

	var opts []eventcatalog.Option
	if *plain { //nolint:branching-flow // flag.Bool never returns a nil pointer
		opts = append(opts, eventcatalog.WithPlainRefIDs())
	}

	if err := eventcatalog.NewExporter(dir, opts...).Export(domain.BuildCatalog()); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.catalog",
			"export eventcatalog tree to "+dir))
	}

	// Wire-contract spec files (the service page's specifications
	// frontmatter points here): the SAME pure renders behind
	// `webphone contracts` and GET /openapi.json, so the hub's inventory
	// page links live contract bytes, never hand-maintained copies.
	//
	// They live DIRECTLY IN the service directory with bare-filename
	// frontmatter paths because @eventcatalog/core resolves a
	// specification through TWO resolvers, and only a bare filename
	// satisfies both: the lint/MDX resolvers join the FULL path onto
	// the resource page's directory (core/src/utils/resource-files.ts),
	// while the spec viewer pages join only the path's BASENAME
	// (core/src/utils/collections/util.ts processSpecifications ->
	// filename: getBasename(path), consumed by
	// pages/docs/[type]/[id]/[version]/{asyncapi,spec}/[filename].astro).
	// A specifications/ subdirectory satisfied the first and silently
	// 404'd the viewer in the federation hub ("No OpenAPI spec file
	// found"; the AsyncAPI page parsed empty content and reported a
	// misleading "asyncapi field missing"). A tree-root specifications/
	// dir, the shape before that, matched nothing at all.
	serviceDir := filepath.Join(dir, "services", domain.CatalogServiceID)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.catalog",
			"create "+serviceDir))
	}
	wire, err := server.RenderWireContractsYAML()
	if err != nil {
		return bootErr(bootGeneric, err)
	}
	if _, err := atomicwrite.WriteIfChanged(filepath.Join(serviceDir, "asyncapi.yaml"), wire); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.catalog",
			"write the asyncapi specification"))
	}
	apiBody, err := server.RenderOpenAPIContract()
	if err != nil {
		return bootErr(bootGeneric, err)
	}
	if _, err := atomicwrite.WriteIfChanged(filepath.Join(serviceDir, "openapi.json"), apiBody); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.catalog",
			"write the openapi specification"))
	}

	_, _ = fmt.Fprintf(os.Stdout, "catalog: wrote eventcatalog tree to %s\n", dir)
	return nil
}
