# Status report — a green main, a six-hour blind spot, and the fix for it

**Snapshot: 2026-10-28 10:30 CET.** `HEAD` = `8edd0ea`, **pushed**, working tree clean, **CI green**
(`gh run view 40662665439`, all five jobs, on `b43d1ab` — the watcher commit is script-only and its run
was still in flight at the minute this was written; see §9.1).

This report covers **this session** (2026-10-27 22:00 → 2026-10-28 10:30, with a break overnight). The
previous two reports (`2026-10-27_22-00_…`, `2026-10-28_02-40_…`, `2026-10-28_03-45_…`) hold the
detailed analysis of the type-gate campaign; this one adds what happened after 03:45 — and the answer to
*"what did you forget?"* is not flattering:

> **While I slept, main went red again, and the watcher I trust did not say a word.** A dependency bump
> landed at 04:14 CET without its vendored-module hash; CI failed on both `nix build` and `fast flake
> checks`; and `main-red-watch` reported **NONE** for ~6 hours because it asked GitHub for "the newest
> run on main" and inspected **its own always-green cron run** instead of the CI run beside it. I found
> it at 10:09, re-derived the hash, verified the build plus a live smoke, pushed, and CI is green.

---

## 1. What was asked

Report on what I did, what I forgot, what I could have done better, what I could still improve — then
wait.

---

## 2. The numbers

| Metric | 03:45 (last report) | **10:30** | How measured |
|---|---|---|---|
| `origin/main` | `3ebfff3` | **`8edd0ea`** | `git rev-parse` |
| CI on main | 5/5 green | **5/5 green** | run 40662665439 @ `b43d1ab` |
| Main red window | — | **~6 h 05 min** (04:14 → 10:20 CET) | ci.yml run 40625533810 created 03:14Z, fixed by `b43d1ab` at 09:2xZ |
| Cause | — | **missing vendorHash after `go-health v0.8.1`** (3rd occurrence of this class in this repo) | CI log: `specified sha256-CbfEd3… got sha256-3Jos…` |
| Red-main watcher | reports NONE | **scoped to `ci.yml`**, self-test 9/9 | `scripts/main-red-watch.sh --self-test` |
| `nix build .#webphone` | green (pre-bump) | **green on the bumped tree** | `nix build .#webphone --no-link` |
| Live smoke | 48 / 4 / 8 | **48 / 4 / 8, 0 failed** on the bumped binary | `scripts/webphone-smoke.py --bin <nix store path>` |
| `go mod tidy` | — | **one cleanup** (retired `go-health v0.7.0` go.sum lines) | `git diff go.sum` |
| `scripts/error-context.py --self-test` | never run this leg | **37/37 passed** | direct run |
| Type gate / island specs / flake check / buildflow | 0 errors / 261 / passed / 69 steps 0 failed | unchanged (nothing touched since) | prior runs |

---

## 3. FULLY DONE

### 3.1 The overnight red, fixed and verified

- **Diagnosed**: `go-health v0.8.1` landed as a `deps:` commit whose message says *"Verified: build and
  vet green"* — `go build` and `go vet` **cannot** see a Nix fixed-output hash, so the verification
  looked complete and was structurally unable to catch the failure. Both `nix build .#webphone` and
  `fast flake checks` died on `webphone-2.9.0-go-modules`.
- **Fixed with the documented recipe**: `go mod tidy` (dropped the retired v0.7.0 go.sum lines), then
  `nix build .#webphone.goModules --rebuild` to read `got:`, then the pin. My local derivation and CI's
  agreed independently on `sha256-3JosIH0io0ZUPDF5ypQFPAQGdmkQ9X/wOxR3qISgXkg=`.
- **Verified end to end**: `nix build .#webphone` green from that hash, then the **live smoke on the
  built binary**: 48 + restart 4 + boot-failure 8, zero failures.

### 3.2 The watcher's blind spot, found and closed

`main-red-watch.sh` called `gh run list --branch main --limit 1` — no workflow filter. Because the
`watch` lane runs on a 15-minute cron and is green by design, the "newest run" was almost always the
watcher's own; the red CI run was never inspected. **Proof captured both ways**: the unscoped query
returns `watch success`, the scoped query returns `CI success`. The fix adds `--workflow=ci.yml`, a
comment recording the six-hour silence, and a self-test case that fails if the scoping is ever removed
(**9/9 passing**).

### 3.3 What I had forgotten from the earlier legs, now closed

- **CHANGELOG gaps**: the shell memory/UX work (peaks-cache cap, palette memoisation, panel-scoped
  Retry), the `answerReviewDuplicate` extraction, and the `go-health` bump each got their line — the
  ratified bar says committed code-path changes carry one, and three of mine had not.
- **`error-context.py --self-test`** (37 checks) run for the first time this leg: green.

### 3.4 Everything from the earlier legs still stands

Type gate: 0 errors over 26 files, enforced in CI `fast-checks`, 10 pins, `docs/type-gate.md`. Pre-commit
gate 2 (`node --check` on staged assets) with 6 selftest cases. `art-dupl`: 0 new groups against an
80-group baseline re-recorded with the CI-pinned detector, both new groups extracted or ruled in
`docs/dedup-registry.md`. Buildflow: 69 steps, 0 failed. `nix flake check`: all checks passed (on the
pre-bump tree). Island specs 261. `go test ./...` green (31 packages).

---

## 4. PARTIALLY DONE

| Item | State | Missing |
|---|---|---|
| **Red-main monitoring** | watcher scoped to `ci.yml`, pinned by a self-test | the scoping is *asserted*, not *demonstrated live* — defeating it needs a fresh red CI run (see §9.1) |
| `nix flake check` on the **bumped** tree | not run; CI's `fast-checks` ran the equivalent set (format, island js/lint/type, statix, deadnix, vulnix-triage, module eval, package build) | the KVM VM tests are excluded from CI by design; a local `nix flake check` would cover them |
| buildflow | 0 failures (last run pre-bump) | not re-run since the `go-health` bump + hash re-pin; the changed inputs are the module set and one Nix pin |
| branching-flow findings | baseline retired (suppressed nothing) | the 104 findings still need a rule-family triage |
| shared `GOMODCACHE` | isolated cache used for every command here | the shared cache's state is unknown (see §9.3) |
| the other session's review train | 6 breaks healed earlier | still in flight; `app.css` gained its review-artifact styles mid-session (not mine, untouched) |
| `nix flake check --all-systems` | x86_64 green | aarch64 not evaluated |
| visual matrix / stack browser E2E | not run | no behavioral served-markup change |

---

## 5. TOTALLY FUCKED UP (honest list)

1. **I trusted my own 03:45 claim for six hours.** "CI green, main green" was true when written and
   false within 30 minutes. Nothing in my workflow re-checked it — and the one mechanism that exists for
   exactly that (`main-red-watch`) was silently broken. **The report should have carried the watcher's
   health as a fact, not just the run's.** This is the single biggest takeaway of the leg.
2. **The dependency bump's "verification" was structurally unable to catch its own failure.** *"build and
   vet green"* is true and irrelevant: neither tool reads `vendorHash`. My fix was correct, but the
   commit that caused six hours of red is exactly the case AGENTS.md warns about ("bumps do NOT ride the
   daemon") — and the hook that enforces that binds humans only, because the daemon overrides
   `core.hooksPath`. So the guard exists, is documented, and cannot fire.
3. **The watcher bug is mine to own in a subtler sense**: it was written with the right intent
   ("independent of the push lane") and the wrong query, and its own self-test only exercised the
   *decision core*, never the *argument*. A test of a pure function cannot catch a wrong input.
4. **I still cannot commit a narrative.** Every commit I made tonight and this morning lost its message
   to the daemon (`chore: auto-commit N changed files`), including the vendorHash fix. The only commits
   carrying my text are the ones small enough to beat the race.
5. **I have not re-run `buildflow` or `nix flake check` since the bump** — the two gates most likely to
   disagree with `go build`. CI covers most of it, but "CI ran it" and "I ran it" are different claims,
   and my reports have been careful about that distinction everywhere else.
6. **`gopls` in this session is useless and I kept reading its output**: ~1,600 phantom errors plus a
   phantom checksum failure (`v3.5.0.mod`) that was real once (cache corruption) and then stale. It cost
   minutes twice.
7. **Two stale webphone processes** (6h52m, 8h25m, an older store path) are listening on `:8080`; not
   mine, correctly left alone — but I never checked before starting the smoke, and my `webphone version`
   invocation actually booted a server that could not bind.

---

## 6. WHAT WE SHOULD IMPROVE

1. **Monitoring must observe the thing it claims to observe.** `main-red-watch` is now scoped; the next
   trap is the same shape — a query without a `--workflow` filter, a status API that reports the caller.
   Rule: every watcher names its subject explicitly, and its self-test covers the *invocation*, not only
   the decision.
2. **A dependency bump owes a `nix build`, not a `go build`.** The one-line fix to the bump ritual: the
   claim in the commit message should be `nix build .#webphone` + smoke, because only that reads the
   hash. Worth a template line in AGENTS and in `scripts/bump-preflight.sh`'s output.
3. **The daemon bypasses every hook by construction** (`GIT_CONFIG_*` override). Either the pma unit
   stops exporting it, or the repo stops pretending the guards bind the daemon — today the go.mod guard
   is advisory against exactly the actor that keeps breaking it.
4. **Reports should carry a "verified at HH:MM" stamp per claim, and a re-check step**, because a
   six-hour-old green statement is a liability. The repo's status-report convention already says this;
   my own report did not obey it for the CI row.
5. **One `scripts/gate.sh`** (fmt → tsc → go test → nix build → flake check → smoke) so a bump is
   verified by the same order every time.
6. **`crush` should not run gopls here** (it cannot see the module cache; its diagnostics are noise).
7. **Per-session `GOMODCACHE`** so one full disk does not poison every agent's build.
8. **`nix flake check` should be part of any "I changed a Nix file" loop** — I changed `vendorHash` and
   verified with `nix build`, which is correct, but the KVM-backed checks have not seen this tree.

---

## 7. NOT STARTED (deliberately, reasons in TODO_LIST)

Split `shell.js`; type-check `island-tests/*.mjs`; typed key map for the shell's copy;
`tsc --explainFiles` in CI; aarch64 flake check; the visual matrix; the stack browser E2E; the
branching-flow triage of 104 findings; a deliberate `buildflow timings` rebaseline; the `gate.sh`
entry point.

---

## 8. THE PLAN — next actions, ordered

### P0 — do not let this happen a fourth time (~30 min)
1. Extend `scripts/bump-preflight.sh` to print the **mandatory bump tail** (`go mod tidy` →
   `nix build .#webphone.goModules --rebuild` → pin → `nix build .#webphone` → smoke) at the end of every
   run, whatever the bump.
2. Add that tail to `AGENTS.md`'s dependency bullet as one line naming `nix build`, not `go build`.
3. Re-run `buildflow` and `nix flake check` on this post-bump tree and record both verdicts in
   TODO_LIST's tail row.
4. Leave a red-main issue open until the fix commit's CI is green (the watcher files CREATE, but nothing
   closes it — verify the close path).

### P1 — process fixes (this week)
5. File the daemon-detach + commit-swallowing defect against the pma unit.
6. Rewrite the whitespace-drift hook around `treefmt --ci` (formatter output is not drift).
7. Decide whether `tsc` also runs in the Go CI job; record the decision where CI lives.
8. Make `scripts/buildflow.sh` and the devshell export a session-scoped `GOMODCACHE`.
9. Add `scripts/gate.sh` (one order, one exit code) and call it from the release runbook.
10. Teach the watcher to also assert `nix build`'s presence in the CI run it inspects (the two failures
    tonight were build-level).

### P2 — type the rest of the shell's DOM surface
11. `ShellDomIds` for shell-owned ids (`wp-live`, `wp-tab-skeleton`, `wp-tab-error`, `wp-lightbox`).
12. `inputElement`/`formElement` twins; retire the last inline casts in `shell.js`.
13. `hasMethod` as a type predicate, then drop the follow-on casts.
14. Type-check the `.mjs` specs (third tsconfig, `@types/node`) or record the final NO.

### P3 — evidence
15. Fold this leg into `FEATURES.md` (type gate row: "CI-enforced"; watcher row: "scoped").
16. Cross-link `docs/type-gate.md` ↔ `docs/dom-contract.md`.
17. Incident note: "the watch lane reported green while CI was red" with the two `gh run list` proofs.
18. Incident note: module-cache corruption + the sumdb check that separates it from an attack.
19. Deliberate `buildflow timings` rebaseline.
20. Triage the 104 branching-flow findings by rule family, then re-record the baseline.
21. Re-run `scripts/error-context.py --self-test` after every code-table change (37 checks; green today).

### P4 — payload (next train)
22. Split `shell.js` (voicemail player / palette / composer / notes), load order unchanged, E2E budgeted.
23. Typed key map for the shell copy with an explicit "island failed to load" fallback.
24. Injectable `now()` at construction; delete `window.__wpClock`.
25. Spec the shell's load-failure breadcrumb (`#log` line).
26. Bound the History panel's rendered transcript segments per group.
27. `./update.sh` prints the sip.js public surface to re-read.
28. Island dictionary en/de key parity test (catches the duplicate-key class at `go test` time).
29. One home for "the tab failed" (`#wp-error` slot vs `.wp-error` banner).
30. `history.templ` dir/outcome selects join `filterOptions` for consistency.
31. Devshell prints the checked-file count from `tsc --listFiles`.
32. ADR: "the served tree IS the source tree".
33. Protocol header in `docs/dedup-registry.md`: `check` (new-only) vs `baseline` (all) counts differ.
34. A `main-red-watch` case for "CI red but its run is OLDER than a newer watch run" (the exact shape of
    tonight's miss, beyond the workflow filter).
35. Check whether the `:8080` dev servers belong to a service unit; if they are orphans, say so in
    AGENTS so the next session does not wonder.

---

## 9. Questions I cannot answer myself

1. **Who owns the `go-health v0.8.1` bump?** It landed as an owner-authored `deps:` commit with
   `--no-verify` and a "build and vet green" verification that cannot see the Nix hash. If that was you
   (or another session), fine — the fix is in. If it was an unattended agent, the same commit shape will
   recur, and the durable fix belongs in the bump tooling (P0-1), not in this report. Tell me which, and
   whether the bump tail should refuse to commit at all without a `nix build` transcript.
2. **Should the watcher also close the red-main issue it opens, or is closing always human?** Tonight it
   never even opened one; now that it is scoped, the open/close lifecycle matters. If closing stays
   human, the issue tracker will accumulate stale reds (five `daemon-audit` issues are already open) —
   do you want an explicit "verified green, closing" comment from the watcher as evidence, or silence?
3. **What should happen to the shared `/mnt/buildcache/go-mod`?** Unchanged from the 03:45 question:
   corrupted by a full disk, then emptied by another process, and re-extraction there produced files Go
   could not read; I finished on an isolated cache. A full rebuild re-downloads gigabytes and breaks
   other agents' builds while it runs — operator call.
