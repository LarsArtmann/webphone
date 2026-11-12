# Deepening trains + review series — session status

**Snapshot:** 2026-11-12 05:22 · claims verified against source/tree at write time ·
the concurrent session's storage-ladder train was moving readmodels gates under this
session the whole night (notes/reviews/agentaudit/export) — attribution noted per item.

## Scope

Six skills executed on owner order: architecture-visualization, architecture-review,
go-modularize, data-model-review, full-code-review, improve-codebase-architecture
(the last after the owner corrected me — it exists; `disable-model-invocation` had
kept it out of my trigger list). Plus deepening-candidate execution and cross-train
residue fixes. Companion reports all committed: D2 pair + 4th architecture review
(`docs/architecture-understanding/2026-11-11_15-22*`), 5th modularization entry
(`docs/modularization/2026-11-11_PROPOSAL.html`), 5th data-model review +
3rd full-code-review (`docs/reviews/`), pareto plan, deepening-candidates report
(`2026-11-11_20-23_deepening-candidates.html`).

## a) Fully done

1. **Visualization pair** — current + improved D2 renders of the post-journal-world
   architecture; INDEX rows added.
2. **4th architecture review** — 7-dimension re-score **4.29** (was 4.5). All four
   2026-09-30 findings verified: three FIXED (fax Receipt.Resolution, fail-closed
   boot typing, schema_version), views-prop typing still open. Findings: P1
   `internal/server` god-package (106 files, 342 exports, Ce=25), P2 views prop
   reach-through, P3 store-ladder retirement date.
3. **5th data-model review** — carried items re-verified (push.Transport,
   DocumentPath, machineSurface all still open with honest site markers);
   sso.Complete recommended CLOSED as site-ruled; two new findings (store
   ThreadSummary transitional home, ThreadFlag double definition).
4. **3rd full-code-review** — ~480 files, ten read-only cluster sweeps + lead's own
   server-hub pass. **Lead-verification contract held**: 7 fabricated subagent
   claims rejected (one whole server-core sweep was fiction and redone by hand).
   Ten fixes landed same day, all verified: passkey Settings own-account (HIGH,
   the shared-extension revocation hole), ended-call Close staleness (HIGH),
   textagent connection pool, admin inventory engine-side sort, transcript sweep
   termination guard, summarize verdict allowlist, dial-gate body cap, badge-error
   logs, note maxlength, six stale readmodels.db comment sites.
5. **Deepening candidates executed (3 of 4)**:
   - **Identity tail (C2)** — `sessionAnswer` is the ONE credential-answer policy;
     six dead passkey nil-gates removed (route registration owns the seam-off gate).
   - **Push fan-out (C4)** — `push.Fanout.Fan(ctx, owner, fact)` owns detach +
     budget + prune-on-unregistered; the ring hook's synchronous loop (contract
     violation, filed finding) and notifyDraftReady's private copy both deleted;
     registry armed in App.Start via an adapter; race-tested.
   - **Thread list vocabulary (C1)** — ThreadSummary render shape + the
     pinned/archived/muted vocabulary live in readmodels (ONE home, ended the
     store's double definition); store's ladder spelling unexported with its death
     dated; templ regenerated.
6. **Call lifecycle core (C3)** — `sipverdict.js` (the re-INVITE verdict lane) +
   holdSession rewired: hold/resume now settles on the far-end VERDICT, not the
   send — the island honesty contract's oldest gap, filed the same day, closed the
   same night. Test stubs ride the verdict lane (accept()/refuse(), synchronous
   delegate fire).
7. **Cross-train residue tonight** — hold-test stubs made pollution-safe (the
   dialer-focus red in shell.test.mjs was MY microtask leak, caught by their green
   baseline); app.css five physical text-aligns → logical (their craft train landed
   past the M24.5 styles gate); `readmodels/pushgate.go` orphan (their half-move,
   duplicate of committed push.go) removed to unbreak the build.
8. **CHANGELOG Unreleased** batch for the ten fixes; **TODO_LIST** harvest rows for
   the open findings (hold verdict, ring fan-out, thread-split probe, blob-orphan
   delete).
9. **Final gates at snapshot**: full Go suite green, island **333/333**, tsc **0
   errors** (the webauthn.js reds from yesterday are gone — the concurrent
   session's train resolved them), race clean on push/hooks/api.

## b) Partially done

1. **Call lifecycle split (C3's larger half)** — intent interface exists for
   hold/resume only; calls.js is still ~2 000 lines with post-call products riding
   the transport path. The extraction seams are named in the deepening report.
2. **Views prop typing (arch P2)** — thread vocabulary moved; `pbx.CDR`,
   `pbx.Voicemail*`, textagent/reports props still reach through (admin.templ,
   history.templ, voicemail.templ untouched).
3. **Server carve (arch P1)** — dormant blueprint, trigger fired (auth surfaces
   changed today), not executed.
4. **TODO_LIST rows vs reality** — hold-verdict and ring-fan-out rows are DONE in
   code but NOT struck in TODO_LIST; the blob-orphan row landed only its TODO
   (no code); CHANGELOG entries for the four deepening refactors are NOT written
   (only the ten fixes have entries).

## c) Not started

- Thread-split race probe (concurrent first-sight, the unverified data-loss claim).
- Blob-orphan compensating delete on persist failure.
- `internal/server` three-way carve (auth/html/platform).
- docs-gates battery over this session's doc volume (annotate/cadence ratchets
  unmeasured — the report/TODO volume this session added is exactly what the
  ratchets measure).
- buildflow full run at train end (only scoped package suites ran).
- CI verdict verification on the night's pushed commits (daemon push anchor was
  34 min at last check — under the 60-min bar; verdicts unverified per the
  workflow-name rule).
- arch error-registry drift (7 problems seen ~21:30 during the readmodels rename
  churn) — attribution and current state unverified.

## d) Totally fucked up (owned)

1. **The first server-core sweep was fiction** — 8 of 9 medium+ findings fabricated
   (nonexistent functions, races that aren't there). Caught only because the lead
   re-read every claim; a full dispatch wasted and the report nearly poisoned.
2. **Provider fleet burned ~40 min** — four re-dispatches of the same review across
   gemini (no credits), hyper (no credits), huggingface (no credits), synthetic
   (rate-limited) before doing the review myself. Provider health should have been
   probed once, up front.
3. **My holdSession edit shipped a SYNTAX ERROR** (`finally` without `try`) that
   sat in the tree through daemon commits with the island suite red — while the
   concurrent session shipped v2.14.0 around it. test-before-commit was skipped at
   exactly the wrong moment.
4. **The verdict-lane stub then leaked microtasks** across test files (three
   diagnostic rounds to find the cross-file DOM pollution; the fix was firing the
   delegate synchronously).
5. **The 5th modularization entry got publicly challenged** ("36 packages and 80K
   lines is fine?!?!") — my framing buried the package-depth debt under the
   module-boundary ruling. The honest answer was always two sentences: the go.mod
   boundary is closed by construction; `internal/server`'s god-package owes the
   carve, and its trigger fired this week.

## e) What we should improve

1. **Lead-verify sweep prompts before dispatching**: strict quote contracts reduce
   but do not eliminate fabrication; the quote-must-exist check should be a grep
   step, not trust.
2. **Probe provider health once** (one 1-token call per provider) before any
   fan-out; fall back to lead-owns-the-cluster immediately on two failures.
3. **Commit-per-verified-step** held mostly, but the two syntax/stub incidents both
   came from editing island JS without running the suite first — the rule is
   "scoped suite before every commit," no exceptions at 21:00.
4. **Depth debt needs a budget, not a narrative**: per-package LOC/export ceilings
   as an arch test would have surfaced the server god-package mechanically instead
   of via reviews.
5. **Cross-session tangles are the norm** (four sessions + daemon): every fix
   tonight touched a file another session also touched; the attribution discipline
   in this report is the minimum viable hygiene.

## f) Next queue (priority-ordered, ~34)

**Integrity (P0/P1):**

1. Strike the two landed TODO_LIST rows (hold verdict, ring fan-out) + add
   CHANGELOG entries for the four deepening refactors.
2. Thread-split race probe (two-goroutine Receive, unseen remote).
3. Blob-orphan compensating delete + the sweep/orphan decision.
4. Verify arch error-registry state post-readmodels-rename; regenerate if drifted.
5. CI verdicts for the night's pushed commits (`scripts/ci-verdict.sh`).
6. Rebase check: my thread-vocabulary commit vs their ladder train's final
   readmodels shape (both touched threads.go territory).

**Structure (P1):**
7. Execute the `internal/server` carve — auth → html → platform, dormant
blueprint, trigger fired.
8. Per-package LOC/export ceilings as an arch test (the depth budget).
9. Views prop typing for the remaining infra types (CDR, Voicemail, textagent).
10. Store-ladder retirement TODO row (detection command + delete date).
11. calls.js post-call extraction (postcall.js) per the deepening report.
12. Delete the dead fax vocabulary doors (ParseFaxStatus consumers gone) or
re-point the webhook switch.
13. `DeliveryStep.Event()` unknown-Kind rejection + round-trip test.
14. `SendRefused` FailureKind validation.
15. arch-test services allowlist → denylist (new packages classified by default).

**UI/honesty (P2):**
16. Hold/resume: add the re-INVITE verdict pins the new tests enable
(rejected/timeout arms at the UI level).
17. Matomo-free glyph cleanup: panels.js text glyphs → inline SVG icons.
18. `unreadDivider` aria-label fix (em-dash leak).
19. German locale weekday/month fixes (reports.templ, crm.go) + DE pins.
20. Admin CDR party fallback (`firstNonEmpty` reuse).
21. Note textarea cap is bytes-vs-chars mismatched for multibyte — align.
22. login-page.js localStorage try/catch.
23. Dial-form double-submit guard.
24. `htmx.ajax` unhandled-rejection catches (restoreLastTab/live-context).

**Docs/ritual (P2):**
25. Run the docs-gates battery; measure the ratchets after this session's volume.
26. buildflow full run; erraudit nolint-audit after the sweep-refactor commits.
27. Annotate tonight's four companion reports into INDEX cross-refs (done for
arch-understanding; reviews/modularization INDEX rows pending).
28. Answer the standing cqrs-lint gating + C044 owner questions (pending rows).
29. File the cqrs-lint `/*` lexer draft upstream (owner tick pending).

**Smaller verified leftovers (P3):**
30. Signal receipt-noise WARN → benign-skip parity.
31. matrix oversized-sync decode (streaming) + garbage-cursor self-heal probes.
32. FCM token invalidation on 401/403.
33. CRM response body caps (io.LimitReader, 1 MiB).
34. `Resolver.Name` dead surface delete (or mark test-only).

## g) Questions only the owner can answer

1. **Depth budget:** should the quarterly modularize watch be replaced by hard
   per-package ceilings (LOC/export counts as an arch test) with the server carve
   as the first enforcement — or do you want the carve as a one-off train and the
   watch left as-is?
2. **readmodels is mid-rename under another session** (notes/reviews/agentaudit/
   export gates). My thread-vocabulary commit touches threads.go/summary_view.go
   in the same package: do you want me to rebase my residue on their final shape
   when they land, or stay off readmodels entirely until they call it done?
3. **The two remaining integrity items** (thread-split race probe, blob-orphan
   delete) plus the server carve: one combined cqrs-integrity train, or the
   standing queue?
