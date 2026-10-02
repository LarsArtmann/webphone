# Status: tier-2 family-adoption train — executed, gated, honestly reviewed

> ARCHIVED 2026-10-02 (docs-health v6 sweep): the train SHIPPED inside
> v2.8.0; tier-2 has read STAYS-0 at every measure since (0/0/0 on
> 2026-10-01). Open follow-ups are routed (TODO tooling-hygiene +
> watches rows, ROADMAP guard-ideas, owner briefing); oops/T15 stays
> owner-gated in the live error-excellence plan. Per-item verdicts inline.

**Session:** 2026-09-30, ~10:30–11:45 · **Scope of this report:** this
session's work and what it touched. Not a project-wide audit (owner
instruction). **Baseline → end:** erraudit tier-2 **132 → 0** findings;
owner's full audit-mode command **217 → 77** violations, all documented
residue; every gate green.

---

## a) FULLY DONE (verified this session)

~~1. **Research + mandate confirmation** — erraudit flags decoded against~~ done — this session (report of record)
   the tool's own `--help`; the SUPERB-error-excellence plan, its T02
   decision record, TODO_LIST row 30 (the family-adoption mandate) and
   the "swept-but-unused?" owner question located; go-error-family API
   (constructors, registry, `Error()` rendering, WrapOnce semantics)
   studied from source before any edit.
~~2. **Decision record FIRST** — seven principles (P1 classify at origin,~~ done — this session (report of record)
   P2 family-neutral propagation, P3 sentinels stay sentinels +
   registration, P4 rendered strings never change, P5 stable non-empty
   codes, P6 generic_return = one decision / zero new types, P7 no
   churn on ignores) + the per-seam family table, written into the plan
   appendix before conversion started.
~~3. **blob seam** — 7 sites converted (Infrastructure + `blob.escape`~~ done — this session (report of record)
   Rejection) + family_test (incl. wrap/Join survival).
~~4. **store seam** — db/messages/contacts/faxes/sweep (~55 sites) →~~ done — this session (report of record)
   Infrastructure `store.*` codes; `ErrNotFound`/`ErrListFull`
   registered Rejection in `init()`; the list-full wrap keeps
   `errors.Is` identity; db_test.go contract updated to the classified
   shape (Contains + family/code asserts).
~~5. **domain seam** — Parse refusals → Rejection (`domain.extension`/~~ done — this session (report of record)
   `domain.phone`), corrupt stored id → Corruption (`domain.id`);
   stale "no external dependencies" package doc corrected; arch test
   green.
~~6. **session seam** — 7 sites → Infrastructure `session.*` + family~~ done — this session (report of record)
   pins.
~~7. **config seam** — 27 sites: operator-input → Rejection `config.*`,~~ done — this session (report of record)
   the static defaults map is the one Orchestration site; the
   both-sources sentinel became an interface-typed classified value;
   ALL Contains-based config tests passed unchanged (message bodies
   preserved).
~~8. **pbx seam** — 8 sites; transport/decode Transient, build/encode~~ done — this session (report of record)
   Infrastructure, URL Rejection, non-2xx split 4xx-Rejection/
   5xx-Transient mirroring the gateway; sentinels registered
   (Disabled→Infrastructure, Unauthorized→Rejection); identity kept.
~~9. **crm seam** — 9 sites mirroring pbx; three sentinels registered;~~ done — this session (report of record)
   `errIsMiss` equality semantics untouched.
~~10. **gateway/messaging/fax service seams** — inbound store failures~~ done — this session (report of record)
    classified under the T02 code names (`store.thread_resolve`,
    `store.attachment_save`, `store.message_append`, `store.fax_spool`,
    `store.fax_create`, `store.outbound_status`); non-verdict webhook
    payload → Rejection; the polymorphic `"gateway: %w"` wraps and
    form-builder inner wraps marked family-neutral with reasoned
    nolints.
~~11. **server stragglers** — flexPages payload validations → Rejection~~ done — this session (report of record)
    (`webhook.fax_pages`); the three LogErrorContext log-context wraps
    marked family-neutral (a fixed-family wrap would clobber provider
    4xx Rejections in logs).
~~12. **cmd/webphone** — `propagatef`: ONE home for the nine startup~~ done — this session (report of record)
    wraps' family-neutral policy.
~~13. **erraudit's own rule satisfied** — `sentinel_concrete_type`:~~ done — this session (report of record)
    classified package-level sentinels must be declared as the `error`
    interface (real catch, fixed same session).
~~14. **Gates, all green:** full Go suite (15 packages, run repeatedly);~~ done — this session (report of record)
    tier-1 `--type-aware --disable-extensions` = 0 violations; tier-2
    `--enforce-go-error-family` = **0**; `--enforce-coded-errors` = 0;
    `nix flake check` **all checks passed** (incl. KVM backup VM,
    island-lint, treefmt, statix); smoke **41+4** on a fresh
    version-stamped binary (session survival across kill -9 proves the
    converted store paths live); `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`
    **exit 0** (after two real catches below).
~~15. **Docs synced** — AGENTS erraudit bar rewritten (convention + rules~~ done — this session (report of record)
    - STAYS-0 duty), TODO_LIST row → DONE + watch row updated,
      CHANGELOG entry, plan execution log, P7 honesty correction (see d).
~~16. **Measured deltas** — audit mode 217→77 (residue: 40 ignored, 25~~ done — this session (report of record)
    nolint'd sites surfaced by `--no-suppress` design, 10
    generic_return, 2 counted-skip swallows); generic_return itself
    fell 42→10 as a side effect (functions stopped constructing stdlib
    errors internally); `erraudit tree` snapshot: flat, 8 named errors,
    as designed.

## b) PARTIALLY DONE

~~- **The 40 ignored sites** — re-triaged against the 2026-09-21/22~~ folded into the standing STAYS-0 monthly re-measure (watch row); not separately walked
  documented triage (locations/shapes unchanged), NOT individually
  re-walked. The plan appendix now says exactly that; per-site walk is
  open follow-up.
~~- **docs/error-contract.md** — wire table stays truthful (zero behavior~~ routed — TODO tooling hygiene row (family totality + code registry, added 2026-10-02)
  change) but does not yet tell the "families are total" story or carry
  the code names; the stack-runbook cross-sync section could grow a
  family column.
~~- **docs/lessons.md** — two genuine lessons from this session are NOT~~ done — pipe-exit lesson landed in lessons.md; formatter-ownership rule lives in the AGENTS formatting bullet (nix fmt before gates)
  yet written: (1) `go test | tail` masks exit codes and let red builds
  into commits, (2) buildflow's oxfmt pass ≠ the flake's treefmt oxfmt
  on comment alignment (two-formatter war, Go edition).
~~- **classifyForUser × registered sentinels** — analysis done (send-path~~ routed — ROADMAP guard-ideas bullet (pin class, 2026-10-02)
  flows branch on `errors.Is` before classify), but the interaction has
  NO pin test: if a `store.ErrNotFound` ever leaks into the send path
  it now classifies Rejection→422 where untagged previously meant
  502. Deserves a test that pins today's intended answer.
~~- **Remote push** — local 22 commits ahead of origin at report time;~~ resolved by events — daemon pushes verified via ls-remote at every close since (AGENTS rule)
  daemon lagging (live instance of the pending push-lag-threshold owner
  question).

## c) NOT STARTED (deliberate, with reasoning)

~~- **samber/oops adoption (T15/D2)** — owner-gated by written plan; no~~ owner-gated — the live error-excellence plan (T15/D2)
  go.mod movement, no oops import. Correctly out of bounds.
~~- **Stack browser E2E re-run** — not owed by this train (no markup, no~~ record stands — not owed by this train; the owed run is the standing mod_enum-blocked obligation (TODO row)
  served-asset change; pinned suites cover the wire); rides the release
  gate anyway.
~~- **Old-vs-new binary render diff** — the dedup train's byte-truth~~ record stands — render-diff.py committed + live-proven by the 01:45 train instead
  tool; skipped in favor of the byte-pinned views/classify/i18n suites
  (all green). Defensible, but it was the strongest possible
  no-drift proof and I chose not to spend it.
~~- **FEATURES.md / dedup-registry sweep line** — judged not owed (no~~ record stands (judgment accepted)
  user-facing feature change; not an art-dupl sweep). Judgment calls,
  cheap to revisit.
~~- **RegisterStdlibDefaults** — deliberately not registered (YAGNI,~~ recorded here — standing no (fail-open correct); revisit only if a stdlib-error surface appears
  fail-open already correct); decision recorded only in my head, not in
  the docs.

## d) TOTALLY FUCKED UP (all caught and repaired same session)

~~1. **Committed RED builds — twice.** `go test … | tail -3 && git add …~~ process record — lesson landed in lessons.md (PIPESTATUS rule)
   && git commit` — the pipe fed `tail`'s exit 0 into the chain, so
   broken session and crm commits landed and had to be amended. Root
   cause is a shell-hygiene bug I kept in my command patterns. Only
   luck (daemon push stalled) kept red intermediates off origin.
~~2. **~6 tool calls of self-inflicted test compile errors** — invented~~ process record
   helpers (`newTestDB` cross-package), wrong signatures
   (`store.Open` two-value), mixed named/unnamed params, an edit that
   glued two lines, a wrong "Save fails in a writable tmpdir"
   assumption. I grepped existing tests before writing, but not
   carefully enough. Sloppy first drafts, each caught by the suite —
   that's what the suite is for, but the churn was avoidable.
~~3. **One overclaim in the decision record.** P7 originally said the 40~~ process record — corrected in-session (visible above)
   ignored sites were "Re-verified 2026-09-30, unchanged" — my
   verification was a re-triage against documented priors, not a
   per-site walk. Corrected to the honest wording this session. Closest
   I came to lying in the docs; flagging it here so the correction is
   visible.

## e) WHAT WE SHOULD IMPROVE

~~- **Never pipe a test run into a commit chain.** Check `$?` of the test~~ rule landed — lessons.md PIPESTATUS entry
  command itself. Candidate lessons.md entry + personal standing rule.
~~- **Formatter split brain (Go edition):** `buildflow format` left the~~ rule landed — AGENTS formatting bullet (nix fmt before the gates)
  flake's treefmt check RED — the two oxfmt configurations disagree on
  comment/vertical alignment. I worked around it with `nix fmt` (the
  gate's own formatter). This is exactly the two-formatters-one-file-set
  war AGENTS documents for the island, now mild but real for Go.
  Belongs in BuildFlow (config parity) or an explicit "nix fmt before
  flake check" rule.
~~- **The todo-checker BUG rule matched the word "bug" in prose** —~~ rule of thumb (word "bug" in prose)
  caught at the gate, reworded. Worth remembering when writing comments
  near the word.
~~- **gopls ran stale diagnostics all session** (kept reporting~~ rule — compiler over LSP cache (lessons.md)
  long-fixed errors). I learned to trust builds over the LSP pane;
  restarting the LSP earlier would have removed noise.
~~- **Registration semantics need a pin** (the classifyForUser×NotFound~~ routed — ROADMAP guard-ideas bullet (2026-10-02)
  case in b).
~~- **Residue deserves one enumerated home** — this report now enumerates~~ done — §a16 enumerates the 77
  the 77; the plan appendix could carry the same table so nobody
  re-derives it.
~~- **Real-path tests beat constructed-contract tests** where cheap: the~~ rule recorded; the real-cap drive routed to ROADMAP
  ErrListFull pin constructs the wrap instead of driving 500 saves
  (milliseconds in SQLite); driving the real cap would be strictly
  stronger.

## f) Next things (session-grounded; 35 — no filler)

~~1. Verify the daemon push landed (`git ls-remote`), else hand-push per~~ done — standing close-out ritual (ls-remote asserted at every train close since)
   owner's call.
~~2. Pin `classifyForUser` × registered `ErrNotFound` (what status may a~~ routed — ROADMAP guard-ideas bullet
   leaked NotFound produce — today's intended answer, test-enforced).
~~3. Replace the constructed ErrListFull wrap pin with a real 500-save~~ routed — ROADMAP guard-ideas bullet
   cap drive.
~~4. lessons.md: the `| tail` exit-code-masking lesson.~~ done — lessons.md PIPESTATUS entry
~~5. lessons.md: buildflow-format vs treefmt alignment skew lesson.~~ done — AGENTS formatting bullet owns the rule
~~6. Route the oxfmt/treefmt config parity to BuildFlow (owner repo).~~ owner — BuildFlow config parity (g2)
~~7. error-contract.md: family totality + code names; sync the stack~~ routed — TODO tooling hygiene row
   runbook side.
~~8. Per-site walk of the 40 ignored findings (closes P7's follow-up).~~ folded into the STAYS-0 monthly re-measure (watch row)
~~9. Fold the erraudit tree snapshot (flat 8, by design) into the plan~~ rides the 2026-10-22 re-measure if still wanted (LOW)
   appendix.
~~10. Owner release decision: fold this train into the v2.8.0 tag; ride~~ done — folded into v2.8.0 (tagged 2026-10-01; this train shipped inside)
    the runbook (stack lock bump → stack gates → aarch64 → pbx relock
    #5 → deploy → smoke).
~~11. Stack browser E2E at the release gate (owed by the release).~~ routed — TODO island-honesty + cross-repo rows (mod_enum-blocked)
~~12. T15 oops decision: adopt via bridge/ or permanently ratify~~ owner-gated — live error-excellence plan
    non-adoption.
~~13. Decide RegisterStdlibDefaults (context/sql/os taxonomy at the app~~ recorded here — standing no (see c)
    boundary) — currently an undocumented no.
~~14. Check whether any operator runbook greps exact startup-failure~~ routed — rides the boot-contract stack-runbook patch application (TODO row)
    lines that now carry `[family:code]` prefixes (journalctl recipes).
~~15. Check the vcard import log line ("store rejected …") for grep~~ routed — same leg (boot-contract runbook patch, TODO row)
    dependents (its store errors now carry prefixes).
~~16. Verify webhook 400 body texts for the flexPages validations are~~ covered by P4 (rendered strings never change) + the unchanged Contains tests (§a7)
    pinned (they answer fixed copy; confirm a test holds it).
~~17. Consider `errorfamily.HandleError` at main.go exit (family →~~ not adopted — exit taxonomy shipped as bootreport 1/2 constants (2026-10-02)
    sysexits exit codes) instead of hardcoded `os.Exit(1)`.
~~18. Consider `/metrics` label for error codes on failure paths (bounded~~ deferred — revisit with the error-code registry table (TODO tooling row)
    cardinality — codes are a closed set).
~~19. Session `SQLiteStore.Get` still swallows scan errors (returns~~ routed — ROADMAP guard-ideas bullet (debug log, 2026-10-02)
    false) — pre-existing; candidate for a debug log now that it's
    visible.
~~20. `.WithContext("path", path)` on pbx/store wraps for operator logs~~ not adopted — failure-path only; revisit on operator-log demand
    (library feature unused so far; failure-path only).
~~21. Cross-repo consistency: should Ledger CRM's API mirror the family~~ ROADMAP fuel — cross-repo vocabulary mirroring (owner taste)
    vocabulary on its error responses?
~~22. Inform the webphone×bridge compat-matrix TODO row: 4xx texts~~ done — the matrix bullet landed in AGENTS at `0aab677` (2026-10-01)
    unchanged (pinned) — no bridge action needed for this train.
~~23. Add `--enforce-coded-errors` to the documented tier-2 invocation in~~ done — AGENTS erraudit bar names both flags (tier-2 enforced green)
    AGENTS (it held at 0; make it part of the bar).
~~24. Publish a small "error code registry" table (grep-generated) —~~ routed — TODO tooling hygiene row (2026-10-02)
    codes are now a stable contract worth a page.
~~25. LSP restart hygiene when diagnostics contradict builds.~~ rule — compiler over LSP cache (lessons.md)
~~26. blob family test's chmod trick is Linux-only — fine for this~~ record stands (Linux-only, fine)
    product; note it in the test.
~~27. Re-check the concurrent setup-shell-adoption train before any~~ moot — that train closed with its NO-GO verdict the same day (15:49 report)
    shared-file edit (their T01/T02 owner ratifications pending).
~~28. OWNER-calls batch: mark "tier-2 family-adoption intent" RESOLVED~~ done — the intent question resolved by execution; briefing carries the rest
    (executed); the "2.8 vs reword" question is now sharper (this train
    is in the fold).
~~29. v2.8.0 announcement draft owes an error-architecture paragraph.~~ done — drafts written 2026-10-01 (T15, A/B/C incl. the error-architecture story)
~~30. Monthly tier-2 re-measure 2026-10-22 must read STAYS-0.~~ standing watch — TODO watches row (re-measured 0/0/0 early on 2026-10-01)
~~31. Consider teaching erraudit (owner's tool) a rule: flag family-fixed~~ owner — their tool (erraudit upstream)
    wraps around polymorphic causes (the P2 trap) — the inverse of
    today's constructor rule.
~~32. Probe: does `WrapOnce` deserve use anywhere here (currently unused;~~ not adopted — neutral wraps are deliberate (documented in the seam nolints)
    our neutral wraps are deliberate)? Answer likely no — document.
~~33. Confirm gopls/golangci stale-cache warning noise is tooling-only~~ record stands (tooling-only noise)
    (it was this session).
~~34. Double-check `server.actions` line 509's `"store rejected …"`~~ record stands — P4 keeps rendered strings stable; the composition unchanged
    import-log composition under the new prefixes (visual review).
~~35. Consider a tiny `TestCodesAreStable` golden over all `errorfamily`~~ routed — rides the error-code registry table (TODO tooling row)
    code literals (guards renames breaking log consumers).

## g) Questions I can NOT answer myself

~~1. **Residue bar:** is "77 documented" the ratified standing posture~~ owner — 77-documented posture (g1; briefing holds release/quality postures)
   for your audit-mode flag set, or do you want it driven lower
   (concrete error returns, oops enrichment, per-site ignore cleanup)?
   This subsumes T15 but is broader.
~~2. **Formatter authority:** fix the buildflow-oxfmt ↔ treefmt-oxfmt~~ owner — BuildFlow config parity (g2)
   alignment skew in BuildFlow (config parity), or accept `nix fmt` as
   the mandatory pre-flake-check step? BuildFlow is your repo — your
   call which side owns Go alignment.
~~3. **Release sequencing:** does this train ride the untagged v2.8.0~~ done — folded into v2.8.0 (tagged + shipped 2026-10-01)
   fold (one tag, one stack relock, one deploy), or should it sit on
   main behind the pending release? (Compounds the bridge's ">= 2.8"
   wording question already in the OWNER-calls list.)

---

_Format note: written as .md per explicit owner instruction (the
status-report skill's canonical format is HTML). Auto-commit daemon
picks this file up; not hand-committed per harness rules._
