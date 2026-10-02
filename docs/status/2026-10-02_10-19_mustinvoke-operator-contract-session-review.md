# Status Report + Brutal Self-Review: MustInvoke / Operator-Error-Contract Session

- **Date**: 2026-10-02 10:19 CEST
- **Scope**: This session ONLY (4 conversation rounds, zero file changes authored). No unrelated research performed, per instruction.
- **Format note**: `.md` at explicit path per operator demand — overrides the status-report skill's HTML-canonical default (flagged per skill contract, not propagated as a new default). brutal-self-review findings folded into (d)/(e) instead of a separate HTML report, per the same single-file instruction.
- **Tree state**: ~19 files modified by OTHER concurrent sessions (AGENTS.md, internal/**, island tests). I authored ZERO changes this session. Those in-flight files are not mine to judge, revert, or commit here.

> **HARVESTED 2026-10-02 11:15** (ANNOTATE, in place): the operator
> ordered execution ("fix!"). The boot-error contract SHIPPED the same
> morning — renderer + wiring + `App.Start` de-panic + pins live in
> `cmd/webphone/bootreport{,_test}.go` and `cmd/webphone/main.go`; the
> ruling lives in `docs/error-contract.md` § "Boot surface"; the D3
> owner call + stack runbook patch are harvested into `TODO_LIST.md`
> and `docs/planning/2026-10-02_11-05_boot-contract-stack-runbook-patch.md`
> (which also carries the records: InvokeAs rejected, ~33-call count
> correction, E2E-not-owed, 2026-10-22 re-audit tie-in). Section (f)
> items below are therefore DONE or ROUTED — do not re-harvest.

---

## Session chronology (what actually happened)

| Round | Prompt                         | What I did                                                                                                                                                               | Outcome                                                                                                            |
| ----- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| 1     | "Why do we use MustInvoke*?"   | Grepped all usage (37 matches), read `internal/app/app.go:100-246`, explained fail-fast boot semantics, DO-1 sanction, HW-4 suppression, framework-free service packages | Answered with evidence; loaded `samber-do-best-practices` skill                                                    |
| 2     | "What would our skills say?"   | Loaded `go-error-modernization` + `how-to-golang`; verified every call site (incl. `App.Start`, app.go:281-295) against DO-1                                             | Verdict table: "all compliant" — **later shown to grade the wrong axis**                                           |
| 3     | "Operator is a valid user too" | Read `cmd/webphone/main.go` (operator-visibility lines 44-67), grepped `docs/error-contract.md` (line 27 RecoveryMiddleware precedent, line 82 operator semantics)       | Conceded: boot panic = 1/5 contract parts, designed boot errors = 2/5; sketched boot renderer; said "say the word" |
| 4     | Status report (this file)      | Wrote this report                                                                                                                                                        | Waiting                                                                                                            |

---

## a) FULLY DONE

1. **Complete MustInvoke inventory + classification** — every call site located and classified: `New` provider closures, the handler provider (app.go:225-241), and boot-time `App.Start` (app.go:283, 291-292). Verified ZERO calls in request paths, service packages, or post-serve code. DO-1 conformance at the mechanism level is airtight.
2. **Three skills loaded and applied** — samber-do-best-practices (DO-1…DO-6), go-error-modernization (erraudit tiers, suppression-with-reason discipline), how-to-golang ("no panics in library code").
3. **The "why" answered durably** — fail-fast at the composition root, named errors (`sqlite`, `blob-dir`) for the probe's critical-service contract, `New` doc comment app.go:101-104 ("lazy leftovers would lie to the startup probe"), the one HW-4 suppression carrying an inline reason (app.go:131), services receiving plain constructor args.
4. **Round-3 concession with evidence** — `main.go:46` and `main.go:54` prove the repo already treats the operator as a boot-time audience ("Operators must SEE..."); error-contract.md:27 proves a per-request panic-presentation precedent exists that boot lacks.
5. **Zero damage** — no files touched, no tests broken, no git operations, nothing to roll back.

## b) PARTIALLY DONE

1. **Skills verdict** — mechanism-level compliance established (solid); presentation-level grading started only after the operator pushback and is ~60% mapped: two surfaces scored (panic 1/5, designed errors 2/5), remediation shape sketched in chat only.
2. **Boot-error-contract design** — the 5-line render for the panic case exists as chat prose; no design doc, no code, no tests.
3. **Memory upkeep** — durable discoveries identified as record-worthy (AGENTS.md composition-root rationale, operator ruling gap) but not yet written anywhere (see d-2).
4. **`App.Start` improvement** — identified during THIS report's drafting: all three `MustInvokeNamed` calls there sit in a method that already `return`s error, so `InvokeNamed` + `wrapf` removes them at zero cost. Analysis done; change not made.

## c) NOT STARTED

1. Boot panic renderer (recover → 5-part render → non-zero exit) in `cmd/webphone/run()`.
2. Structured boot-failure printer for designed errors (config, dataDir, timezone, paperless) replacing the bare `slog.Error` string at main.go:24.
3. Any edit to `docs/error-contract.md` (boot-surface section, operator-as-user ruling).
4. Stack-runbook mirror sync (tri-repo obligation).
5. All tests for the above (golden render test, five-parts-never-empty table test, exit-code drift test).
6. `App.Start` MustInvoke → Invoke conversion.
7. TODO_LIST.md harvest of section (f).
8. Everything in section (f) below.

## d) TOTALLY FUCKED UP

1. **Round-2 verdict graded the wrong axis.** I declared "compliant across the board" using the DI skills' lens (where panics are _created_) while the project's OWN `docs/error-contract.md` governs how failures are _presented_ — and I even cited it (line 27) without connecting it. The global AGENTS.md error doctrine (what/reassure/why/fix/escape) was in context the entire time. The operator had to supply the counterexample ("an operator is a valid user"). Root cause: I answered the lens I was asked instead of stepping up to the governing contract. **A top-tier engineer audits against the repo's named contract doc first.**
2. **Memory-protocol violation.** Global AGENTS.md mandates recording durable discoveries _at the moment of discovery_. Two qualify (MustInvoke rationale; the boot-surface contract gap + operator ruling) and neither is written anywhere — the knowledge lives only in this chat and evaporates with it. This report partially mops up, but AGENTS.md/error-contract.md edits remain owed.
3. **Precision slips I must own:**
   - "All 37 call sites" — the grep returned 37 _matches_, ~4 of which are comment lines (app.go:14, 129, 131, 135). Real calls ≈ 33. Wrong number stated confidently.
   - "systemd captures the panic and the unit fails" — asserted in round 2 BEFORE reading `cmd/webphone/main.go`. It happened to be true (stdlib semantics), but it was unverified in-repo at claim time. Same for the exit-code-2 claim.
   - HW-4 rationale ("the linter can't see transitive resolution") repeated from the code comment unverified against the samber-linter's actual rule text.
4. **"Say the word" instead of deciding.** A reversible, well-scoped improvement sat at the decision point and I asked permission — the engineering-mode protocol says recommend-and-execute for reversible changes. Cost: implementation is at 0% while the design was 90% done.
5. **Missed the `App.Start` conversion for three rounds.** I read app.go:281-295 repeatedly (even cited its `wrapf` calls as good behavior) and never noticed the panics there are gratuitous — the method already returns error. Found it only while writing this report.

## e) WHAT WE SHOULD IMPROVE

1. **Grade surfaces, not mechanisms.** When a repo has a named contract doc (error-contract, dom-contract, dedup-registry), audit behavior against THAT first; library skills answer "is the tool used right", never "does the user see the right thing".
2. **Keep the panic mechanism, add the presentation layer.** DO-1 blesses `Must*` at the composition root and that stays. What's missing is boot's analogue of `cqrshtmx.RecoveryMiddleware` (which logs stack + re-raises per request, error-contract.md:27): a recover-and-render wrapper in `run()` translating container panics into the 5-part operator contract, then exiting non-zero.
3. **Remove gratuitous panics where an error return already exists** — `App.Start`'s three `MustInvokeNamed` become `InvokeNamed` + `wrapf` for free. (The `New`/provider-closure panics stay: miswiring there has no recovery and the named error already lands in the panic text.)
4. **One home for the operator ruling**: `docs/error-contract.md` boot section, mirrored to the stack runbook (the AGENTS.md cross-doc sync rule), dated 2026-10-02.
5. **Verify numbers before citing them** — matches ≠ call sites; comments pollute greps; exit codes deserve a check.
6. **Write memory at discovery, not at report time** — the aggressive-update protocol exists precisely so a killed session loses nothing.
7. **Decide reversible things.** Present the recommendation, execute, note the alternative in one line.

## f) Up to 50 things we should get done next

_Brainstorm, not commitment list (status-report skill note). Impact-ordered within priority tiers. Items 1-12 are this session's direct debt; the rest are adjacent value spotted along the way._

**P0 — this session's direct debt**

1. Boot panic renderer in `cmd/webphone/run()`: recover → 5-part operator render → non-zero exit (mechanism unchanged, presentation fixed).
2. Structured boot-failure printer for designed errors (config/dataDir/timezone/paperless), replacing bare `slog.Error` (main.go:24).
3. Escape lines verified against the NixOS module: `nixos-rebuild switch --rollback` + `systemctl reset-failed webphone`.
4. Reassure lines per failure class (panic: SQLite opened-not-written; config errors: nothing touched).
5. Record the operator-as-user ruling (2026-10-02) as a new boot-surface section in `docs/error-contract.md`.
6. Sync the mirrored error-contract section in the stack's release runbook (tri-repo push-order ritual).
7. Exit-code taxonomy (1 = designed boot error, 2 = panic-rendered) documented in the runbook's systemd section.
8. Golden test: injected container panic asserts the 5-part render + exit code.
9. Table-driven test: every designed boot error renders ALL five parts — never an empty fix/escape line.
10. Convert `App.Start`'s three `MustInvokeNamed` → `InvokeNamed` + `wrapf` (app.go:283, 291-292).
11. Fold the MustInvoke rationale into AGENTS.md's composition-root bullet (late memory write — do now).
12. HARVEST this report's (f) into `TODO_LIST.md` (status-report skill: section f must not be entombed).

**P1 — near-term adjacent**

13. Language ruling for boot journal copy (expect English per D3 analog) — decide and record (see Q2).
14. Renderer scope decision: panics-only vs all boot failures through the 5-part printer (see Q1).
15. Restart policy decision: fail-and-wait for operator eyes vs `Restart=on-failure` auto-retry; document + module check (see Q3).
16. Re-grep the exact MustInvoke call count; correct the "37" figure wherever it lands in docs.
17. Independently verify the samber-linter HW-4 rule text behind the suppression (app.go:131).
18. Arch test: `do.MustInvoke*` must not appear outside `internal/app`.
19. Fully read `docs/error-contract.md` around line 82 (existing operator semantics) before editing — split-brain guard.
20. One-home check before adding the boot section (no duplicate operator-error doc may exist).
21. Extend `cmd/webphone/drift_test.go` for render/exit-code drift.
22. Tag the render with boot phase (New-phase vs Start-phase — "data touched" differs between them).
23. Version stamp in boot-failure output (rollback verification aid).
24. Keep the full stack trace AFTER the 5-part block (bug-filing aid, journal not spammed).
25. Smoke addition: unwritable dataDir boot asserts an operator-grade message.
26. Standing gates per change: `nix develop -c go test -count=1 ./...`, `nix fmt`, buildflow.
27. Concurrent-session discipline: re-View `main.go`/`app.go` immediately before each edit (~19 foreign-modified files in tree).
28. Attribution: commit only files I author; leave the daemon to the rest.
29. erraudit discipline for any new code: `propagatef` pattern for family-neutral wraps; errorfamily for any NEW error path (the renderer emits prints, not errors — expected zero new paths).
30. Config-class fix hints: IANA zone example for timezone, both-or-neither citation for paperless, permission hint for dataDir.

**P2 — backlog fuel**

31. README "Readiness vs systemd" section: one line on boot-failure presentation.
32. Include the 1/5 → 2/5 grading table as a dated audit appendix in error-contract.md.
33. `docs/lessons.md` war-story line: "graded the mechanism, missed the surface" (this session's miss).
34. Record `do.InvokeAs` as considered-and-rejected (all deps are concrete types) to prevent re-litigation.
35. Verify smoke `--expect-version` is unaffected by stderr shape changes.
36. Check the app-seam erraudit family pin covers `cmd/webphone` propagation.
37. Cite the exact call-site inventory (file:line) in the boot section for future audits.
38. Dashboard-enabled boot path: Start-phase failures (pusher) render distinctly from New-phase.
39. Tie the next boot-surface re-audit to the erraudit monthly cadence (next: 2026-10-22).
40. Design within item 2: slog structured attrs (what/why fields) vs prose in the printer.
41. Explicitly record that NO stack browser E2E is owed (no markup change) — prevent cargo-cult E2E runs.
42. When this report goes stale, update via docs-health ANNOTATE mode (never rewrite).
43. Keep the ruling project-local — nothing propagates to global skills (project-first memory rule).
44. Deliberately decide whether the boot section needs a contract-parse test like `TestServedPageHoldsTheDomContract` (probably overkill; decide, don't drift).
45. Leave other sessions' dev servers running during any testing here (AGENTS.md concurrent-session rule).
46. CHANGELOG entry when the implementation lands (one home for history).

## g) Questions I cannot figure out myself

1. **Renderer scope**: route ALL boot failures (config, dataDir, timezone, paperless included) through the 5-part printer for a uniform operator surface, or render only container panics and keep designed errors as enriched `slog` output? This shapes `main.go` structure and the test matrix. My lean: uniform — but it is a bigger diff.
2. **Language**: is boot journal copy English-only? D3 ruled the UI shell English and the island en/de; the boot journal surface was never ruled. My lean: English-only (operator trail precedent, `#log` channel).
3. **Restart + timing**: should a boot failure sit FAILED for operator eyes, or auto-retry via `Restart=on-failure`? And does the stack-side runbook/module sync (tri-repo push + relock ritual) ride THIS train or the next release? This is a release decision, not a code decision.

---

**Waiting for instructions.** Report is uncommitted by design (harness forbids commits without explicit request; the auto-commit daemon will pick it up).
