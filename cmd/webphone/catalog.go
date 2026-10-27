package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-cqrs-lite/catalog/v4/eventcatalog"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
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

	dir := filepath.Join(*output, "eventcatalog")

	var opts []eventcatalog.Option
	if *plain {
		opts = append(opts, eventcatalog.WithPlainRefIDs())
	}

	if err := eventcatalog.NewExporter(dir, opts...).Export(domain.BuildCatalog()); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.catalog",
			"export eventcatalog tree to "+dir))
	}

	fmt.Fprintf(os.Stdout, "catalog: wrote eventcatalog tree to %s\n", dir)
	return nil
}
