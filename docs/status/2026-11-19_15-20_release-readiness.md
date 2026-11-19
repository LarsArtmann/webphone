# Status — go-release readiness analysis (v2.15.0 candidate)

> Snapshot-rot header: claims verified as of 2026-11-19 ~15:20 CET. Push/CI
> statements rot fastest; CI run 44529291480 was in flight at write time.

Posture: the go-release skill executed in **readiness-analysis mode** — the
full decision lifecycle up to (not including) the tag/push boundary. The
owner's release ritual (release.sh + tri-repo push dance + pbx-artmann
relock) is deliberately out of agent hands.

## Version determination

- Latest tag: **v2.14.0**. Since it: **422 commits (176 human, 246 daemon)**.
- Human commits include FEATURES: the line-ownership data model
  (`authmap.Line` end to end), the typed passkey lane for passkey-only
  deployments, the typed `services.webphone.textagent` + oauth2 NixOS module
  lanes, the docked sidebar rail, the admin read-only surface refinements,
  notes autosave, the e2e workflow test lane, and this train's fix-forward
  batch (SSE shutdown log honesty, /metrics nil-safety, /login CSP reporting,
  contacts-import escaping, UTF-8 transcript cap, disposition echo-guard,
  green login favicon).
- Semver judgment: **MINOR → v2.15.0** (features + fixes, no breaking wire
  change; the session wire's `lines[]` typed shape already shipped its
  breaking half pre-tag with fixtures regenerated).

## Gate status (verified this session, commands in the transcripts)

- `go test -count=1 ./...` — 37 packages ok (twice: baseline + post-fix).
- Island suite 369/369 pass; `tsc -p . --noEmit` green.
- `scripts/format-check.sh` green; `dprint check` green after fmt.
- `scripts/docs-gates.sh --check` — **full matrix PASS** (format,
  status-f-section, battery, table-damage, strike-safety) after the cadence
  drain (14 reports archived gates-first; 69→55 unarchived).
- `scripts/changelog-guard.sh` — **PASS** ("14 product-code file(s) changed
  and CHANGELOG.md carries the entry").
- `scripts/release-hygiene.sh` — lychee **ok** after repairing 11 broken
  relative links the archive moves created; origin/main-vs-HEAD mismatch
  resolved by the manual-push bar invocation (anchor: oldest unpushed commit
  2026-11-19T07:44:29+01:00, 365 min ≫ 60-min bar; red main + verified fix
  in hand). Pushed: `25301790..d0a9a3be`.

## CHANGELOG completeness

The `[Unreleased]` section is release-ready prose (concurrent sessions
maintained it through 2026-11-19; this train added the fix-forward batch and
favicon entries). changelog-guard mechanically confirms coverage.

## Blockers (the release ritual's preconditions)

1. **CI green on main** — run 44529291480 in flight at write time; verify via
   `scripts/ci-verdict.sh` before cutting.
2. **The dependency decision** — the same-day currency sweep
   (docs/research/2026-11-19_dependency-currency-sweep.md) found **10
   consumed-surface behinds, all additive**, incl. a fix-class cqrs-lite
   system bump (RawQuery+ProjectionSpec crash-loop repair). Ship v2.15.0 on
   current pins, or run the 4-train bump plan first and fold those CHANGELOG
   lines in? Owner call; the plan is spelled out either way.
3. **Clean clone discipline** — the v2.13.0 lesson: the release must run
   from a clean clone of green main, never a dirty worktree.

## The sequence (from docs/release-runbook.md, for the owner)

bump `webphoneVersion` in flake.nix → cut CHANGELOG → full gates →
`release.sh` → push webphone → stack re-pin (per-train lock bump) →
pbx-artmann relock (clean tree first) → verify end states with
`git ls-remote` + `scripts/ci-verdict.sh`.

## f) Up to 50 things we should get done next

1. Watch CI run 44529291480 to green via `scripts/ci-verdict.sh`, then hold
   until the owner picks the bump-vs-ship order for v2.15.0.
2. Run the T1 cqrs-lite trio bump train (system v4.19.0 fix-class first)
   per docs/research/2026-11-19_dependency-currency-sweep.md if the owner
   folds deps into v2.15.0.
3. Run the T2 templ v0.3.1118 + CLI-matched `templ generate` regen train
   (same sweep's plan; codegen pairing is mandatory).
4. Run the T3 modernc sqlite v1.62.0 train (backup-drill + vulnix gates).
5. Run the T4 small-SDKs train (httputil, paperless, idempotency WithClock
   migration retiring TimingScale 5, nanoid, oauth2).
6. Fix the fax Resend x retention document-integrity edge (fcr-02 M1,
   TODO row harvested 2026-11-19) before the release if cheap.
