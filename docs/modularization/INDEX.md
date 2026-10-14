# Modularization rulings index

Point-in-time rulings live here as HTML (grep-opaque — list this index
first when hunting prior rulings).

| Date       | Artifact                                    | Verdict                                                                                                                                           |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-10-14 | [2026-10-14_PROPOSAL.html](2026-10-14_PROPOSAL.html) | Single go.mod **RATIFIED** (third series verdict). Zero external Go consumers (verified GitHub-wide); every package under `internal/` blocks a split structurally; co-change 2% per 150 commits. Re-open ONLY on the named trigger: a second Go consumer importing `internal/domain` or `internal/gateway`. |

Earlier verdicts in the same series live in `docs/architecture-understanding/`:

- 2026-09-19 architecture review §03 "Should this be split into Go modules? No." — `2026-09-19_19-20_architecture-review.html`
- 2026-09-30 architecture review §05 "Go modules?" (re-run of the assessment, same verdict) — `2026-09-30_11-45_architecture-review.html`

Standing rule: AGENTS.md "Single go.mod is RATIFIED". The below-module
refinement (the `internal/server` carve) stays live in `TODO_LIST.md` —
its one home.
