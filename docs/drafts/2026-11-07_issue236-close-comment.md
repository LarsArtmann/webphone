Verdict: **a real, latent test-code data race — fixed** — closing on evidence.

Root cause (from the failed log of run 42308961304): `TestLoginClaimsAuditLineCarriesGroupsOnlyInTheJournal` swapped the process-global `slog.SetDefault` handler to its own `bytes.Buffer` while `t.Parallel()` let the sibling sso tests keep running — their own "oidc login claims" lines went through THIS test's buffer, and the race detector flagged the concurrent `bytes.Buffer.Write` against the test's `buf.String()` read (`internal/sso/sso_test.go:605`). Timing explains the intermittency: a docs-only head commit cannot introduce a race, and a loaded runner widened the window.

Why it is not just a flake: owning a process-global while parallel is unsound whether or not the detector fires, and the sibling lines were bleeding into the buffer the assertions read.

Fixed in `2c50b2c7`: the test runs sequentially now (the only sound way to own a global), mechanism noted at the site. The same commit healed the separate `encoding/json/v2` break the concurrent test-migration sweep (`43f70b9c`) had left on this same test — `domain.Extension` now marshals its raw value, so v2-strict `json.Marshal` accepts identity shapes again.

Successor run 42400725432 green, all jobs. Ledger row appended in `docs/ci-flake-ledger.md` (2026-11-07).
