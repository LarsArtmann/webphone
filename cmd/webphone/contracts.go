package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/larsartmann/go-atomic-write"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/server"
)

// runContracts serves the `webphone contracts` one-shot: render the
// machine-to-machine wire contracts (the /hooks/* ingest lanes, the
// gateway receipt, the SSE stream) as an AsyncAPI 3.0 document, plus
// the consumer wire fixtures (wire/*.json — the byte-pinned payload
// examples the native shells decode in their own tests) — no config,
// no database, no network. The consuming stack's bridge and the native
// shells build against these files; the freshness gate regenerates
// them and diffs (a dirty tree means the committed contract rotted
// against the code).
func runContracts(args []string) error {
	fs := flag.NewFlagSet("contracts", flag.ExitOnError)
	output := fs.String("o", "docs/contracts", "directory to write asyncapi.yaml and wire/ fixtures into")
	_ = fs.Parse(args)

	body, err := server.RenderWireContractsYAML()
	if err != nil {
		return bootErr(bootGeneric, err)
	}

	if err := os.MkdirAll(*output, 0o755); err != nil { //nolint:branching-flow // flag.String never returns a nil pointer
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
			"create the contracts output directory "+*output)) //nolint:branching-flow // flag.String never returns a nil pointer
	}
	target := filepath.Join(*output, "asyncapi.yaml") //nolint:branching-flow // flag.String never returns a nil pointer
	if _, err := atomicwrite.WriteIfChanged(target, body); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
			"write "+target))
	}
	_, _ = fmt.Fprintf(os.Stdout, "contracts: wrote %s\n", target)

	fixtures := server.RenderWireFixtureFiles()
	wireDir := filepath.Join(*output, "wire") //nolint:branching-flow // flag.String never returns a nil pointer
	if err := os.MkdirAll(wireDir, 0o755); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
			"create the wire fixtures directory "+wireDir)) //nolint:branching-flow // flag.String never returns a nil pointer
	}
	names := make([]string, 0, len(fixtures))
	for name := range fixtures {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fixture := filepath.Join(wireDir, name)
		if _, err := atomicwrite.WriteIfChanged(fixture, fixtures[name]); err != nil {
			return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
				"write "+fixture))
		}
		_, _ = fmt.Fprintf(os.Stdout, "contracts: wrote %s\n", fixture)
	}
	return nil
}
