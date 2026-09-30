# Status: tier-2 family-adoption train — executed, gated, honestly reviewed

**Session:** 2026-09-30, ~10:30–11:45 · **Scope of this report:** this
session's work and what it touched. Not a project-wide audit (owner
instruction). **Baseline → end:** erraudit tier-2 **132 → 0** findings;
owner's full audit-mode command **217 → 77** violations, all documented
residue; every gate green.

---

## a) FULLY DONE (verified this session)

1. **Research + mandate confirmation** — erraudit flags decoded against
   the tool's own `--help`; the SUPERB-error-excellence plan, its T02
   decision record, TODO_LIST row 30 (the family-adoption mandate) and
   the "swept-but-unused?" owner question located; go-error-family API
   (constructors, registry, `Error()` rendering, WrapOnce semantics)
   studied from source before any edit.
2. **Decision record FIRST** — seven principles (P1 classify at origin,
   P2 family-neutral propagation, P3 sentinels stay sentinels +
   registration, P4 rendered strings never change, P5 stable non-empty
   codes, P6 generic_return = one decision / zero new types, P7 no
   churn on ignores) + the per-seam family table, written into the plan
   appendix before conversion started.
3. **blob seam** — 7 sites converted (Infrastructure + `blob.escape`
   Rejection) + family_test (incl. wrap/Join survival).
4. **store seam** — db/messages/contacts/faxes/sweep (~55 sites) →
   Infrastructure `store.*` codes; `ErrNotFound`/`ErrListFull`
   registered Rejection in `init()`; the list-full wrap keeps
   `errors.Is` identity; db_test.go contract updated to the classified
   shape (Contains + family/code asserts).
5. **domain seam** — Parse refusals → Rejection (`domain.extension`/
   `domain.phone`), corrupt stored id → Corruption (`domain.id`);
   stale "no external dependencies" package doc corrected; arch test
   green.
6. **session seam** — 7 sites → Infrastructure `session.*` + family
   pins.
7. **config seam** — 27 sites: operator-input → Rejection `config.*`,
   the static defaults map is the one Orchestration site; the
   both-sources sentinel became an interface-typed classified value;
   ALL Contains-based config tests passed unchanged (message bodies
   preserved).
8. **pbx seam** — 8 sites; transport/decode Transient, build/encode
   Infrastructure, URL Rejection, non-2xx split 4xx-Rejection/
   5xx-Transient mirroring the gateway; sentinels registered
   (Disabled→Infrastructure, Unauthorized→Rejection); identity kept.
9. **crm seam** — 9 sites mirroring pbx; three sentinels registered;
   `errIsMiss` equality semantics untouched.
10. **gateway/messaging/fax service seams** — inbound store failures
    classified under the T02 code names (`store.thread_resolve`,
    `store.attachment_save`, `store.message_append`, `store.fax_spool`,
    `store.fax_create`, `store.outbound_status`); non-verdict webhook
    payload → Rejection; the polymorphic `"gateway: %w"` wraps and
    form-builder inner wraps marked family-neutral with reasoned
    nolints.
11. **server stragglers** — flexPages payload validations → Rejection
    (`webhook.fax_pages`); the three LogErrorContext log-context wraps
    marked family-neutral (a fixed-family wrap would clobber provider
    4xx Rejections in logs).
12. **cmd/webphone** — `propagatef`: ONE home for the nine startup
    wraps' family-neutral policy.
13. **erraudit's own rule satisfied** — `sentinel_concrete_type`:
    classified package-level sentinels must be declared as the `error`
    interface (real catch, fixed same session).
14. **Gates, all green:** full Go suite (15 packages, run repeatedly);
    tier-1 `--type-aware --disable-extensions` = 0 violations; tier-2
    `--enforce-go-error-family` = **0**; `--enforce-coded-errors` = 0;
    `nix flake check` **all checks passed** (incl. KVM backup VM,
    island-lint, treefmt, statix); smoke **41+4** on a fresh
    version-stamped binary (session survival across kill -9 proves the
    converted store paths live); `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`
    **exit 0** (after two real catches below).
15. **Docs synced** — AGENTS erraudit bar rewritten (convention + rules
    - STAYS-0 duty), TODO_LIST row → DONE + watch row updated,
      CHANGELOG entry, plan execution log, P7 honesty correction (see d).
16. **Measured deltas** — audit mode 217→77 (residue: 40 ignored, 25
    nolint'd sites surfaced by `--no-suppress` design, 10
    generic_return, 2 counted-skip swallows); generic_return itself
    fell 42→10 as a side effect (functions stopped constructing stdlib
    errors internally); `erraudit tree` snapshot: flat, 8 named errors,
    as designed.

## b) PARTIALLY DONE

- **The 40 ignored sites** — re-triaged against the 2026-09-21/22
  documented triage (locations/shapes unchanged), NOT individually
  re-walked. The plan appendix now says exactly that; per-site walk is
  open follow-up.
- **docs/error-contract.md** — wire table stays truthful (zero behavior
  change) but does not yet tell the "families are total" story or carry
  the code names; the stack-runbook cross-sync section could grow a
  family column.
- **docs/lessons.md** — two genuine lessons from this session are NOT
  yet written: (1) `go test | tail` masks exit codes and let red builds
  into commits, (2) buildflow's oxfmt pass ≠ the flake's treefmt oxfmt
  on comment alignment (two-formatter war, Go edition).
- **classifyForUser × registered sentinels** — analysis done (send-path
  flows branch on `errors.Is` before classify), but the interaction has
  NO pin test: if a `store.ErrNotFound` ever leaks into the send path
  it now classifies Rejection→422 where untagged previously meant
  502. Deserves a test that pins today's intended answer.
- **Remote push** — local 22 commits ahead of origin at report time;
  daemon lagging (live instance of the pending push-lag-threshold owner
  question).

## c) NOT STARTED (deliberate, with reasoning)

- **samber/oops adoption (T15/D2)** — owner-gated by written plan; no
  go.mod movement, no oops import. Correctly out of bounds.
- **Stack browser E2E re-run** — not owed by this train (no markup, no
  served-asset change; pinned suites cover the wire); rides the release
  gate anyway.
- **Old-vs-new binary render diff** — the dedup train's byte-truth
  tool; skipped in favor of the byte-pinned views/classify/i18n suites
  (all green). Defensible, but it was the strongest possible
  no-drift proof and I chose not to spend it.
- **FEATURES.md / dedup-registry sweep line** — judged not owed (no
  user-facing feature change; not an art-dupl sweep). Judgment calls,
  cheap to revisit.
- **RegisterStdlibDefaults** — deliberately not registered (YAGNI,
  fail-open already correct); decision recorded only in my head, not in
  the docs.

## d) TOTALLY FUCKED UP (all caught and repaired same session)

1. **Committed RED builds — twice.** `go test … | tail -3 && git add …
   && git commit` — the pipe fed `tail`'s exit 0 into the chain, so
   broken session and crm commits landed and had to be amended. Root
   cause is a shell-hygiene bug I kept in my command patterns. Only
   luck (daemon push stalled) kept red intermediates off origin.
2. **~6 tool calls of self-inflicted test compile errors** — invented
   helpers (`newTestDB` cross-package), wrong signatures
   (`store.Open` two-value), mixed named/unnamed params, an edit that
   glued two lines, a wrong "Save fails in a writable tmpdir"
   assumption. I grepped existing tests before writing, but not
   carefully enough. Sloppy first drafts, each caught by the suite —
   that's what the suite is for, but the churn was avoidable.
3. **One overclaim in the decision record.** P7 originally said the 40
   ignored sites were "Re-verified 2026-09-30, unchanged" — my
   verification was a re-triage against documented priors, not a
   per-site walk. Corrected to the honest wording this session. Closest
   I came to lying in the docs; flagging it here so the correction is
   visible.

## e) WHAT WE SHOULD IMPROVE

- **Never pipe a test run into a commit chain.** Check `$?` of the test
  command itself. Candidate lessons.md entry + personal standing rule.
- **Formatter split brain (Go edition):** `buildflow format` left the
  flake's treefmt check RED — the two oxfmt configurations disagree on
  comment/vertical alignment. I worked around it with `nix fmt` (the
  gate's own formatter). This is exactly the two-formatters-one-file-set
  war AGENTS documents for the island, now mild but real for Go.
  Belongs in BuildFlow (config parity) or an explicit "nix fmt before
  flake check" rule.
- **The todo-checker BUG rule matched the word "bug" in prose** —
  caught at the gate, reworded. Worth remembering when writing comments
  near the word.
- **gopls ran stale diagnostics all session** (kept reporting
  long-fixed errors). I learned to trust builds over the LSP pane;
  restarting the LSP earlier would have removed noise.
- **Registration semantics need a pin** (the classifyForUser×NotFound
  case in b).
- **Residue deserves one enumerated home** — this report now enumerates
  the 77; the plan appendix could carry the same table so nobody
  re-derives it.
- **Real-path tests beat constructed-contract tests** where cheap: the
  ErrListFull pin constructs the wrap instead of driving 500 saves
  (milliseconds in SQLite); driving the real cap would be strictly
  stronger.

## f) Next things (session-grounded; 35 — no filler)

1. Verify the daemon push landed (`git ls-remote`), else hand-push per
   owner's call.
2. Pin `classifyForUser` × registered `ErrNotFound` (what status may a
   leaked NotFound produce — today's intended answer, test-enforced).
3. Replace the constructed ErrListFull wrap pin with a real 500-save
   cap drive.
4. lessons.md: the `| tail` exit-code-masking lesson.
5. lessons.md: buildflow-format vs treefmt alignment skew lesson.
6. Route the oxfmt/treefmt config parity to BuildFlow (owner repo).
7. error-contract.md: family totality + code names; sync the stack
   runbook side.
8. Per-site walk of the 40 ignored findings (closes P7's follow-up).
9. Fold the erraudit tree snapshot (flat 8, by design) into the plan
   appendix.
10. Owner release decision: fold this train into the v2.8.0 tag; ride
    the runbook (stack lock bump → stack gates → aarch64 → pbx relock
    #5 → deploy → smoke).
11. Stack browser E2E at the release gate (owed by the release).
12. T15 oops decision: adopt via bridge/ or permanently ratify
    non-adoption.
13. Decide RegisterStdlibDefaults (context/sql/os taxonomy at the app
    boundary) — currently an undocumented no.
14. Check whether any operator runbook greps exact startup-failure
    lines that now carry `[family:code]` prefixes (journalctl recipes).
15. Check the vcard import log line ("store rejected …") for grep
    dependents (its store errors now carry prefixes).
16. Verify webhook 400 body texts for the flexPages validations are
    pinned (they answer fixed copy; confirm a test holds it).
17. Consider `errorfamily.HandleError` at main.go exit (family →
    sysexits exit codes) instead of hardcoded `os.Exit(1)`.
18. Consider `/metrics` label for error codes on failure paths (bounded
    cardinality — codes are a closed set).
19. Session `SQLiteStore.Get` still swallows scan errors (returns
    false) — pre-existing; candidate for a debug log now that it's
    visible.
20. `.WithContext("path", path)` on pbx/store wraps for operator logs
    (library feature unused so far; failure-path only).
21. Cross-repo consistency: should Ledger CRM's API mirror the family
    vocabulary on its error responses?
22. Inform the webphone×bridge compat-matrix TODO row: 4xx texts
    unchanged (pinned) — no bridge action needed for this train.
23. Add `--enforce-coded-errors` to the documented tier-2 invocation in
    AGENTS (it held at 0; make it part of the bar).
24. Publish a small "error code registry" table (grep-generated) —
    codes are now a stable contract worth a page.
25. LSP restart hygiene when diagnostics contradict builds.
26. blob family test's chmod trick is Linux-only — fine for this
    product; note it in the test.
27. Re-check the concurrent setup-shell-adoption train before any
    shared-file edit (their T01/T02 owner ratifications pending).
28. OWNER-calls batch: mark "tier-2 family-adoption intent" RESOLVED
    (executed); the "2.8 vs reword" question is now sharper (this train
    is in the fold).
29. v2.8.0 announcement draft owes an error-architecture paragraph.
30. Monthly tier-2 re-measure 2026-10-22 must read STAYS-0.
31. Consider teaching erraudit (owner's tool) a rule: flag family-fixed
    wraps around polymorphic causes (the P2 trap) — the inverse of
    today's constructor rule.
32. Probe: does `WrapOnce` deserve use anywhere here (currently unused;
    our neutral wraps are deliberate)? Answer likely no — document.
33. Confirm gopls/golangci stale-cache warning noise is tooling-only
    (it was this session).
34. Double-check `server.actions` line 509's `"store rejected …"`
    import-log composition under the new prefixes (visual review).
35. Consider a tiny `TestCodesAreStable` golden over all `errorfamily`
    code literals (guards renames breaking log consumers).

## g) Questions I can NOT answer myself

1. **Residue bar:** is "77 documented" the ratified standing posture
   for your audit-mode flag set, or do you want it driven lower
   (concrete error returns, oops enrichment, per-site ignore cleanup)?
   This subsumes T15 but is broader.
2. **Formatter authority:** fix the buildflow-oxfmt ↔ treefmt-oxfmt
   alignment skew in BuildFlow (config parity), or accept `nix fmt` as
   the mandatory pre-flake-check step? BuildFlow is your repo — your
   call which side owns Go alignment.
3. **Release sequencing:** does this train ride the untagged v2.8.0
   fold (one tag, one stack relock, one deploy), or should it sit on
   main behind the pending release? (Compounds the bridge's ">= 2.8"
   wording question already in the OWNER-calls list.)

---

_Format note: written as .md per explicit owner instruction (the
status-report skill's canonical format is HTML). Auto-commit daemon
picks this file up; not hand-committed per harness rules._
