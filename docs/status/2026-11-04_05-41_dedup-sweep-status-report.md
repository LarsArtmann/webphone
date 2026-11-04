# Status report — dedup sweep session, 2026-11-04

**Claims verified as of 05:41 CET.** Push/CI statements rot fastest; timestamps below are the anchor.

## Scope

**Verified here (fresh commands this session):**
- `check . -t 1 --baseline-path .art-dupl-baseline.json` with the CI-pinned `art-dupl@v0.7.2` (built from the tag into `/tmp/artdupl072`, `version` = dev/v0.7.2 tag source): **"Found total 0 clone groups" vs the 86-entry baseline** — run twice (session start and end), both clean.
- The owner-pasted SARIF (`~/.crush/shell-output/output-3792907013.log`, host `art-dupl v0.9.1-0.20261017225421`, `-t 3`): parsed with python — 130 per-instance results → **29 distinct groups**, each fragment re-read at HEAD via `sed` at the reported file:line (line numbers matched HEAD, so the paste is at/near current HEAD).
- `grep` repo-wide for `WriteByte('\n')` and `Unwrap() []error` (non-test): exactly 3 files.
- Final `git status`: only `docs/dedup-registry.md` (M) + the new sweep status doc (??) changed by this session.

**NOT verified here (and why):**
- **CI state** — I made no push; the daemon owns doc commits. `gh run list` was not read (nothing pushed by me to verify).
- **buildflow / `nix flake check`** — deliberately not run: this session changed ZERO non-doc bytes; markdown is formatter-unowned. No gate can move on this delta.
- **The 86 baselined groups themselves** — not re-adjudicated (quarterly audit due 2027-01-11 per the ratified suppression-scope ruling; not due now).
- **The three newest main commits** (`2d25c3f7` CUE/OpenAPI verdicts, `cabf2edf`+`a87aee62` "stage pending work before dependency update", `965929cb` nixpkgs bump) — noticed, not investigated (out of scope per owner instruction).
- **The host v0.9.1 binary's provenance** — owner's environment; I verified its *output*, not its build.

## a) FULLY DONE

| What | Evidence |
| --- | --- |
| Ratified dedup sweep executed at HEAD | pinned v0.7.2 `check . -t 1` = **0 new** vs 86 baseline (start + end of session) |
| Pasted v0.9.1 `-t 3` report fully triaged | 29/29 groups attributed: 24 declaration/signature shape noise + 5 standing-ruling re-sightings (table in the sweep status doc) |
| The one real logic clone adjudicated | multi-error `Error()` join ×3 (`bootJoinError`/`ShutdownError`/`notFoundError`): ruling located (10-27 erraudit status §29), extraction trigger grep-verified **not fired** (exactly 3 sites) |
| Registry: missing ruling row added | error-join triple now a standing row with its grep-able trigger |
| Registry: duplicate rail-row ruling merged; broken 3-line table fragment removed | final bytes verified via `grep -n`; concepts carried (update rule #4) |
| Sweep log line + status doc | `docs/status/2026-11-04_05-36_dedup-sweep-t1-status.md`, registry tail |
| Working tree verified | `git status --short` = exactly the two intended doc paths |

## b) PARTIALLY DONE

| Item | What works | What remains | Effort |
| --- | --- | --- | --- |
| Big shape-noise group attribution (e.g. `f49b…` 43 sites) | Every occurrence's *shape* verified via the SARIF snippet text (each is a lone field/signature line) | The 43 host files were not each opened in full; justified by the class (declaration noise needs no file context) but rigor is snippet-level, not file-level | S if ever needed |
| CI verification | Local gate clean at HEAD | CI verdict after the daemon commits the docs — unchecked | S |
| v0.9.x anomaly | Observed + recorded: at `-t 3` **zero** groups carried `non_actionable_pattern` despite ~20 defined denylist patterns (e.g. the 43-site group should plausibly match one) | Unexplained — SARIF export may never attach the property; needs tool source/docs read before any upstream filing | S–M |
| Upstream filing of the short-span noise class | Class note recorded (10-29 ruling + today's status doc) | Not filed; `verify-before-filing` gate not run; no explicit ask | M |

## c) NOT STARTED

- **HARVEST of this report's (f) into TODO_LIST.md** (docs-health) — waiting on this report landing; not yet run.
- **Quarterly baseline/suppressed-set audit** — due 2027-01-11 (protocol date; unchanged).
- **Buildflow tiers 1+2 monthly re-measure** — **due 2026-11-05 (tomorrow)** per AGENTS (last 2026-10-05). Untouched this session.
- **Dependency-update train** (`cabf2edf`/`a87aee62`/`965929cb`) — in progress or abandoned by another session; not started *by me*, deliberately.
- **Upstream art-dupl issue(s)** — see (b).

## d) TOTALLY FUCKED UP

**Nothing currently broken.** Honest near-misses, all self-repaired this session:

1. **I reintroduced the exact defect I was deleting.** My first write of the two new registry rows used two physical lines (shape cell on line 1, ruling/why cells on line 2) — the same malformed-table pattern as the fragment I removed. Caught during post-edit verification, joined to single-line rows, re-verified. Root cause: I copied the *broken fragment's* structure instead of the file's one-line row convention. Severity: none (final bytes correct). Lesson applied below in (e).
2. **First edit attempt failed on an assumed byte shape** — I matched `||` (what the View rendering showed) instead of the file's actual `| ` prefix; assertion caught it, bytes inspected via `repr`, second attempt clean. Cost: one round trip.
3. **Pre-existing damage found, not caused by me:** the broken table fragment had sat in the ONE clone-ruling home since the 2026-10-29 17:4x closeout entry (~6 days) — a previous session shipped malformed markdown into `dedup-registry.md`. Now repaired and merged into the standing ruling row.
4. **Repo-level observation, unowned:** main carries two `chore: stage pending work before dependency update` commits plus a nixpkgs bump — a dependency train that per AGENTS must NOT ride the daemon and must land with the vendorHash tail. If it stalled mid-train, that is the most consequential unknown on main right now (not investigated — out of scope).

## e) WHAT WE SHOULD IMPROVE

**Session self-critique (the owner's three questions):**
- *What did I forget?* (1) To state the detector-version policy proactively — the owner had to ask why v0.7.2 (answer: CI's voice; baselines are only valid per detector version — the 10-15 red-main incident run 38781159138 proved a v0.9.0-recorded baseline reds CI with phantom clones, and v0.9.1 still reports 43–47 phantom "new" against the same baseline). (2) I did not surface the v0.9.x actionability anomaly in my first summary — it appeared only when writing this report.
- *What could I have done better?* (1) Copy the file's own row convention when adding rows (near-miss #1 would be impossible then). (2) Verify exact bytes with `repr` *before* composing an edit, not after a failed assertion. (3) The 43-site group could have had a random-sample file-level spot-check for full rigor.
- *What could I still improve (systemically)?* See the concrete items below.

**Process/design improvements:**
1. **Baseline entries carry no rationale.** `.art-dupl-baseline.json` = hash/files/tokens only; the "why" lives in registry rows, and no mechanical hash↔row link exists. Quarterly audits re-derive attributions by grep. Fix: cite the baseline hash in each ruling row (or a generated cross-walk).
2. **Named triggers are prose, not tripwires.** The error-join ruling says "extract when a fourth appears" — enforceable as a 5-line arch-test asserting `count(WriteByte('\n')) == 3`, going red the day the trigger fires.
3. **Registry markdown integrity is ungated.** The broken fragment survived 6 days because nothing parses the file. Fix: extend a selftest to assert one physical line per table row / no orphan single-cell rows.
4. **Sweep triage is hand-rolled each time.** This session wrote the SARIF parser inline twice (two iterations to get `region.snippet`). Fix: a committed `scripts/dupl-summary.py` (distinct groups, sites, snippets, ranked).
5. **The pinned detector is rebuilt into `/tmp` per session.** Fix: expose `art-dupl@v0.7.2` as a flake devShell package so sweeps are one command and version drift is impossible locally.
6. **Non-canonical pastes lack a protocol hook.** I attributed the v0.9.1 paste correctly, but ad hoc. Fix: one protocol sentence — "owner-pasted runs: record detector version + threshold in the sweep-log line, adjudicate only against your own pinned re-run."

## f) Next tasks (ranked; honest count — padding to 50 would fabricate work)

| # | Task | Impact | Effort | Category |
| --- | --- | --- | --- | --- |
| 1 | Buildflow tiers 1+2 monthly re-measure — **due 2026-11-05** | Critical | M | Quality |
| 2 | HARVEST this report's list into TODO_LIST.md (docs-health) | Critical | S | Process |
| 3 | Tripwire arch-test: exactly 3 multi-error aggregates, or the deferred extraction fires | High | S | Quality |
| 4 | Verify the daemon committed the two doc files; re-read registry post-commit (daemon truncation precedent) | High | S | Cleanup |
| 5 | `gh run list` CI verdict after those doc commits land | High | S | Process |
| 6 | Adjudicate the dependency-train residue on main (`cabf2edf`/`a87aee62`/`965929cb`): finish or revert the stalled bump per the vendorHash doctrine | High | M | Cleanup |
| 7 | Owner decision task: upgrade CI art-dupl pin v0.7.2 → current (re-record baseline in the new universe + re-adjudicate delta) or ratify v0.7.2 as the permanent CI voice | High | L | Process |
| 8 | Pin `art-dupl@v0.7.2` into the flake devShell (kill the /tmp rebuild) | Medium | S | Cleanup |
| 9 | `scripts/dupl-summary.py` — SARIF triage in one command | Medium | M | Quality |
| 10 | Baseline-hash ↔ registry-row cross-walk (citations or generated table) | Medium | M | Quality |
| 11 | Registry table-integrity selftest (one line per row; no orphan fragments) | Medium | S | Quality |
| 12 | Investigate v0.9.x SARIF actionability anomaly; document or file upstream | Medium | S–M | Documentation |
| 13 | File upstream (verify-before-filing first): declaration/signature-span noise class + the "N tokens" inconsistency (`f49b` claims 43 tokens for 1-statement field lines) | Medium | M | Quality |
| 14 | Protocol amendment: sweep-log must name any pasted run's detector version + threshold | Low | S | Documentation |
| 15 | Full-run-groups vs baseline set-diff script as quarterly-audit pre-work (due 2027-01-11) | Low | M | Quality |
| 16 | Type-safety roadmap: `push.Transport` defined type + `ParseTransport` | Medium | M | Feature |
| 17 | Type-safety roadmap: `sso.Finish` value object into `Complete` (auth: convention) | Medium | M | Feature |
| 18 | Type-safety roadmap: `FaxJob.DocumentPath` as `domain.RelPath` | Medium | S | Feature |
| 19 | Type-safety roadmap: asr/matrix wiring structs | Medium | M | Feature |
| 20 | Type-safety roadmap: `machineSurface` bool-last | Low | S | Quality |
| 21 | AuditLog READ projection upstream ask (memory-only, unbounded — the 10-25 audit's remaining gap) | Medium | L | Feature |
| 22 | `internal/server` carve refinement (standing TODO_LIST item) | Low | L | Cleanup |
| 23 | ntfy/fcm `authed(ctx)` helper — owner call only if the guard prologue grows | Low | S | Quality |
| 24 | Attended transfer M19 — stays NO-GO until the stack ships Replaces dialplan handling | Low | L | Feature |
| 25 | Session chat-network lane — parked at NO-GO; reopen only on TODO_LIST triggers | Low | — | Feature |
| 26 | webphone-android end-to-end ring once both push transports exist on both sides | Medium | L | Feature |

Items 16–25 are standing AGENTS/TODO inventory restated for HARVEST routing, not new discovery from this session.

## g) Questions I cannot answer myself

1. **Should the CI art-dupl pin be upgraded to the current v0.9.x (accepting a baseline re-record + delta re-adjudication in its noisier universe), or is v0.7.2 ratified as the permanent CI voice?** I can execute either; only you can price the re-baseline vs the stale-detector risk.
2. **Do you want detector-noise classes filed upstream to LarsArtmann/art-dupl** (after the verify-before-filing gate), or does the pinned-only doctrine make upstream filing noise?
3. **Is a dependency-update train currently in flight by another session** (the two "stage pending work" commits + nixpkgs bump on main)? I assumed yes and stayed off `go.mod`/flake territory; if it is abandoned residue, it needs triage per the bump doctrine.
