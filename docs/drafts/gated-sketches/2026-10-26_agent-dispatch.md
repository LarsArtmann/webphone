# GATED SKETCH — agent dispatch on red main (NOT implemented)

Status: sketch only, recorded by the SUPERB error-agents plan (M17).
The OBSERVER half shipped (watch.yml → main-red-watch files the issue);
this is the DISPATCH half — an agent acting on that issue — and it
waits for an owner GO. Trigger criteria at the bottom.

## Evidence pack (the PR template an agent would receive)

The red-watch issue body already carries most of this. The dispatch
template adds what a fix attempt needs:

```markdown
## dispatch: red-main run <id>

- run link: <url>
- failing job: <job>
- prev green SHA: <sha> (last successful run before the red)
- failing commit: <sha> (head of the red run)
- repro commands: (reproduce-first verdict on the red commit, if any)
- error-context bundles: (one per `<seam>.<op>` code in the log tail)
- recurrence: (ledger/lessons matches — known class or new?)
- rollback line: `git revert <failing-commit>` is the standing first
  move when diagnosis exceeds the budget — a reverted red main beats a
  speculative fix landing at 2am.
```

## Runtime decision menu

| Runtime | Mechanics | Secrets needed | Cost profile | Notes |
|---|---|---|---|---|
| claude-code-action | GitHub Action, comments `@claude fix` on the issue | Anthropic key as repo secret | per-attempt API cost | sandboxed to the PR; well-trodden |
| Devin | cloud session fed the evidence pack URL | Devin API key | seat/credits | strongest autonomy, least repo-native |
| Crush headless on a self-hosted runner | `crush run` with the issue body as prompt | local provider (this host's stack) | zero marginal API cost | self-hosted doctrine fits; needs a runner with nix + KVM for the gates |
| (do nothing — the shipped state) | human reads the issue | none | zero | the default until GO |

Every option funnels into ONE invariant: the agent opens a DRAFT PR —
it never pushes main, never re-baselines, never marks a flake "closed"
in the ledger (the ledger's verdict column stays human-owned).

## Trigger criteria (owner calls, recorded not made)

1. GO/NO-GO on dispatch at all (the shipped watchers file evidence;
   dispatch is the risky half).
2. If GO: runtime choice from the menu (secrets + cost implications
   named there).
3. Guardrails budget: max attempts per red run, max concurrent
   dispatches, and the rollback-first rule made explicit.
4. Which labels opt a red-main issue into dispatch (e.g. only
   `ci-red` + `dispatch-ok`, never by default).
