# Contributing

Thanks for considering a contribution to the webphone.

## Setup

```console
nix develop        # Go, templ, golangci-lint, treefmt, esbuild
```

## The loop

```console
nix fmt                    # format everything (nix + go + prettier)
templ generate ./...       # after editing .templ files
go test ./...              # the suite
buildflow                  # the quality gate (lint, audit, checks)
nix flake check            # build + tests in the sandbox + treefmt
```

`GOEXPERIMENT=jsonv2` is required while templ-components depends on
`encoding/json/v2` (Go < 1.27); `nix develop` and `.buildflow.yml` set it.

## Ground rules

- The SIP island's DOM element ids are a published contract (see
  AGENTS.md) — browser E2E in the consuming telephony stack drives them.
  Never rename one without coordinating there.
- Port behavior verbatim first, refactor in a separate verified change.
- One home per fact: README sells, FEATURES inventories status,
  TODO_LIST holds open work, CHANGELOG logs history, AGENTS.md keeps
  session-durable knowledge.
