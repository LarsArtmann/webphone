# Modularization rulings index

Point-in-time rulings live here as HTML (grep-opaque — list this index
first when hunting prior rulings).

| Date       | Artifact                                             | Verdict                                                                                                                                                                                                                                                                                                                                                                    |
| ---------- | ---------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-10-14 | [2026-10-14_PROPOSAL.html](2026-10-14_PROPOSAL.html) | Single go.mod **RATIFIED** (third series verdict). Zero external Go consumers (verified GitHub-wide); every package under `internal/` blocks a split structurally; co-change 2% per 150 commits. Re-open ONLY on the named trigger: a second Go consumer importing `internal/domain` or `internal/gateway`.                                                                |
| 2026-10-17 | [2026-10-17_PROPOSAL.html](2026-10-17_PROPOSAL.html) | **Re-affirmed** (fourth entry, post-metaengine train, all four lenses): scored 7-dimension review (4.0/5), D2 current/ideal pair, data-model readiness audit (domain split-ready; carries only go-error-family + nanoid), and the pre-armed DORMANT split blueprint with costed migration steps. Triggers now carry detection commands; next quarterly re-review ~2027-01. |

Earlier verdicts in the same series live in `docs/architecture-understanding/`:

- 2026-09-19 architecture review §03 "Should this be split into Go modules? No." — `2026-09-19_19-20_architecture-review.html`
- 2026-09-30 architecture review §05 "Go modules?" (re-run of the assessment, same verdict) — `2026-09-30_11-45_architecture-review.html`

Standing rule: AGENTS.md "Single go.mod is RATIFIED". The below-module
refinement (the `internal/server` carve) stays live in `TODO_LIST.md` —
its one home.
