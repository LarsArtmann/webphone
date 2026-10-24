package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/server"
)

// runContracts serves the `webphone contracts` one-shot: render the
// machine-to-machine wire contracts (the /hooks/* ingest lanes, the
// gateway receipt, the SSE stream) as an AsyncAPI 3.0 document — no
// config, no database, no network. The consuming stack's bridge and
// the native shells build against this file; the freshness gate
// regenerates it and diffs (a dirty tree means the committed contract
// rotted against the code).
func runContracts(args []string) error {
	fs := flag.NewFlagSet("contracts", flag.ExitOnError)
	output := fs.String("o", "docs/contracts", "directory to write asyncapi.yaml into")
	_ = fs.Parse(args)

	body, err := server.RenderWireContractsYAML()
	if err != nil {
		return bootErr(bootGeneric, err)
	}

	if err := os.MkdirAll(*output, 0o755); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
			"create the contracts output directory "+*output))
	}
	target := filepath.Join(*output, "asyncapi.yaml")
	if err := os.WriteFile(target, body, 0o644); err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.contracts",
			"write "+target))
	}

	//nolint:erraudit // CLI output stream — write failures are not actionable
	fmt.Fprintf(os.Stdout, "contracts: wrote %s\n", target)
	return nil
}
