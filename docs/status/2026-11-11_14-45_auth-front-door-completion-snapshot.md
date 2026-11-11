# Auth front-door train — completion snapshot

> Claims verified as of 14:45 CET, 2026-11-11. Supersedes the 12-31
> snapshot for completion state; that report holds the train's origin
> story and war stories. Planning doc:
> docs/planning/2026-11-11_13-44_SUPERB-auth-login-front-door-completion.md.

## The one-line state

`/login` is THE auth front door in every mode — library-released
(loginpage v4.18.0, tagged + pushed + proxy-resolved), webphone
rewired, ALL local gates green (Go full suite, island 330/330, tsc,
format, docs battery, art-dupl 0 new clones, smoke 54/54 + 7
scenarios, nix build), webphone + stack pushed — and CI verdicts on
the newest pushes were still in flight at snapshot time (three
transient proxy-flake reds and one docs-format red already explained
and healed; see the ledger below).

## Why the session churned (the honest answer to "why is your job never finishing")

1. **The sleep-poll CI pattern**: every "wait 4–5 min then check `gh
   run list`" call auto-backgrounded, and each report woke a new turn
   that only polled again. The cap should have been: ONE poll, report
   with "CI pending" as an explicit state.
2. **Concurrent sessions + the daemon**: three other Crush sessions
   and the auto-commit daemon were committing/pushing the same repos
   mid-train (my narrative commits were repeatedly swept into
   "chore:" blobs; layout.templ changed under me once; the vendorHash
   was re-pinned by another session before I got to it). Every edit
   needed a re-read + re-verify round.
3. **Real gate churn**: CI surfaced three NEW defect classes after
   each push (art-dupl clone groups ×2, the docs-format rc=20, plus
   proxy stream errors that are pure network flakes) — each needed a
   fix + re-push + re-watch cycle.

## Verification ledger (local, at HEAD 14:45)

| Gate | Verdict |
| --- | --- |
| `go test -count=1 ./...` (full) | PASS, zero FAILs |
| island `node --test` | 330/330 |
| `tsc -p . --noEmit` | exit 0 |
| treefmt / `scripts/format-check.sh` | green |
| docs battery (`scripts/docs-gates.sh`) | PASS (all seven gates) |
| art-dupl pinned v0.7.2 `-t 1` | 0 new clones (baseline 134) |
| `webphone-smoke.py` | 54/54 + restart/boot/asr/dialgate/faxcap/metrics/enroll all green |
| `nix build .#webphone` + goModules | green (vendorHash verified) |
| cqrs-htmx `loginpage` module tests | ok (v4.18.0, tagged, pushed, proxy-resolved) |

## CI ledger (why reds happened, each healed)

| Run (UTC) | Cause | Disposition |
| --- | --- | --- |
| 12:39Z | vendorHash stale after the loginpage bump (daemon commit rode without the tail) | re-pinned (concurrent session), verified `.#webphone.goModules` |
| 13:10Z | art-dupl 2 new clone groups + proxy stream error (transient) | clones DISSOLVED at the sites (b29e753b); proxy flake heals on re-run |
| 13:18Z | same 2 clone groups (pre-fix tip) | healed by b29e753b |
| 13:27Z | proxy stream errors again (go test -race fetch) + docs format rc=20 | transient + canonical bytes landed |
| 13:36Z | docs format rc=20 (non-canonical markdown in flight) | dprint-canonical bytes committed ("docs: dprint-canonical markdown bytes…") |
| 13:44Z | final verdict PENDING at snapshot time | local battery PASS at the same tree state |

## What remains (owner-visible)

1. CI verdicts on the newest pushes (13:40Z/13:42Z/13:44Z) — one
   `gh run list` check, no polling loop.
2. The stack's `webphone-relock` ritual rides the next stack train;
   the stack CI run on my e2e updates (43027908794) was still in
   flight at 27m (its history: the two most recent stack runs are
   success / transient-failure).
3. The stack browser E2E itself (445s budget, KVM VM) has NOT been
   executed against the new flow — the Python side is updated +
   syntax-checked, but the live run is the next session's first
   action.
4. `internal/cqrs/faxes_test.go` sits at 361/350 lines (buildflow
   file-size error) — ANOTHER session's file, left untouched by this
   train.
