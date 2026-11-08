# server.Deps accessor-collapse — design note (2026-11-08, portfolio plan P3-8)

> Status: DESIGN ONLY — no code moved. This note is the ONE home for the
> refactor candidate; the TODO_LIST hygiene row points here. Execution is
> gated on its own trigger (§ When it pays), not scheduled.

## The seam today

`server.Deps` (internal/server/server.go) grows one field per read model.
Seven fields are metaengine-backed read models with the same shape of
truth: owner-scoped, engine-backed (readmodels.db), nil-tolerant in test
compositions, and each consumer branch answers the honest 503 (or the
seam-specific refusal) when its field is nil:

| Field        | Backing collection   | Nil-503 convention site      |
| ------------ | -------------------- | ---------------------------- |
| `Contacts`   | readmodels.Contacts  | contacts handlers + api      |
| `Snippets`   | readmodels.Snippets  | snippet actions (Settings)   |
| `Transcripts`| readmodels.Transcripts | transcript POST + History  |
| `Reviews`    | readmodels call_reviews | review read/dispatch      |
| `Dials`      | readmodels.DialWindows | the dial-gate check+record |
| `Notes`      | readmodels call_notes  | note GET/PUT + History     |
| `AgentAudit` | readmodels agent_audit  | Settings recent-actions    |

(`Messages`/`Faxes` predate the engine and stay store-seam fields; they are
NOT in scope — their lifetime and seam differ.)

The cost is real but bounded: every new read model adds a field, a wiring
literal in app.go (`TestAppWiresEveryServerDep` pins the parity), a test
mutator arm, and one more nil-branch at its consumers. Seven is the number
at which the pattern starts generating mechanical work per addition.

## The candidate shape

Collapse the seven fields into ONE engine-handle field and keep per-model
accessors, so call sites stay typed and the DI surface shrinks to one fact:

```go
type Deps struct {
    // ReadModels is the engine-backed read-model handle. Nil keeps every
    // consumer's honest-degradation convention (the per-surface 503s are
    // policy, not field presence).
    ReadModels *readmodels.Handle
    // …every other field unchanged
}
```

with `readmodels.Handle` exposing `Contacts()`, `Snippets()`, … (methods,
not fields, so a nil Handle degrades at the accessor seam exactly once,
and a future per-model nil (a collection swapped at runtime) stays
representable). The accessor is the ONLY new branch; consumers keep their
typed `*readmodels.Contacts` locals.

## Migration order (each step its own train + suite)

1. Add `readmodels.Handle` (wrapping the engine, per-model getters) and
   `Deps.ReadModels`; app.go wires it alongside the seven fields
   (parallel-run, no behavior change).
2. Move consumers model-by-model: swap `deps.X` for `deps.ReadModels.X()`
   per model, targeted suites green, one commit per model. The
   nil-503 branches move INTO the accessors' callers unchanged (policy
   stays at the surface).
3. Unify test composition: `newTestReadmodels(t)` grows
   `WithContacts/WithSnippets/…` overrides; every
   `func(d *Deps) { d.X = … }` mutator becomes the Handle equivalent.
   `TestAppWiresEveryServerDep`'s literal updates ONCE (one field).
4. Delete the seven fields; the arch-test that pins the DI surface
   (branching-flow large-struct baseline entry) re-measures.

## Test-composition impact (the honest cost)

Roughly 30–40 test sites construct Deps with one read-model field each.
Step 3 is deliberately LAST so the mechanical rewrite happens once,
against a settled Handle shape, with grep-able call sites (`d.Contacts =`
→ `d.ReadModels = …WithContacts…`). The `TestDialGate*` family and the
notes/reviews 409-armed suites are the sensitive ones — their mutators
swap nil vs live stores to drive the degradation arms.

## When it pays (the trigger)

The collapse pays when the EIGHTH read model lands (the next one costs a
field + parity + mutator today, an accessor line after). Until then the
seven fields are honest, pinned, and cheap — executing now would churn
40 test sites to save one future field. Re-read this note at that train;
do not re-litigate the shape.

## What stays out of scope

The OPTIONAL seams (CRM, ASR, TextAgent, Push, PushDevices, Dedup,
UserAuth, SSO, Trace, Probe, Dashboard, Notifier) are features with
distinct lifetimes and config gates — collapsing them into any handle
would blur fail-closed boundaries that are per-seam policy. The collapse
is read-models-only, by construction.
