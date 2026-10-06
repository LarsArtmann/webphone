# CI-flake ledger

ONE home for GitHub Actions flake history — run id, date, failing test,
verdict. Folklore becomes data: before re-litigating a red CI run, check
here; after settling one, append a row. A flake is only "closed" when
the verdict column names the mechanism or the superseding fix.

| Run | Date | Test / stage | Verdict |
| --- | --- | --- | --- |
| 37496169265 | 2026-10-06 | `TestSQLiteSessionTTLExpiryAndSweep` (internal/session), 0.79s, on a commit that could not affect it (ci.yml comments + python lint) | CLOSED, hardening supersedes confirmation: never rerun empirically; the test was made deterministic via an injectable store clock and now passes deterministically — the original red's exact mechanism (~40 ms wall-clock window lost on a loaded runner) stays an inference. Evidence: 2026-10-06_23-22 status report. |
| 37524833682 | 2026-10-06 | vendorHash FOD mismatch (`specified WYqTip91…` vs `got 5ekZFK2m…`) in the sandbox goModules build | NOT a flake — deterministic infra: a post-train `go mod tidy` (commit 2578923) moved go.mod/go.sum AFTER the pin at 4597bc2. Re-pinned via `nix build .#webphone.goModules --rebuild`; successor run 37526835020 green. Rule: after any `go get`, tidy → repin BEFORE push (AGENTS vendorHash bullet). |
| 37536492580 | 2026-10-06 | full run, docs-only commit (round-5 plan table fix) | GREEN, no flake. Recorded as the seed "tonight's session" entry. |

Seeded 2026-10-06 (round-5 plan T06 / docs-health v7 r6). Wiring: TODO_LIST
gate-recovery row and AGENTS.md point here.
