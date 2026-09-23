# Status Report — art-dupl `-t 3` Follow-up Sweep

> CLOSED 2026-09-23 (docs-health): (f) HARVESTED — helper micro-tests
> (+ the apiContactSaved pins), `-race`, error-contract cross-check,
> LSP-nolint fix-or-declare, and the runbook coordination note live as
> TODO rows; the baseline ratification + suppression-bucket doc +
> helper-test bar are owner calls (ROADMAP); the post-tag commits ride
> the v2.6.0 release TAIL re-run for their full-gate pass; AGENTS smoke
> line fixed 38→40 (2026-09-23 sweep); remote sync verified
> (origin == HEAD).

| | |
|---|---|
| **Written** | 2026-09-23 03:01 CEST |
| **Session scope** | The user's `-t 3 --type-aware --html` clone report (5 groups shown) → triage → 1 extraction + 3 accepts → verify → docs → this report |
| **Tree state at write time** | CLEAN; all session work daemon-committed (`e01f75d` docs, `f50e825` code, `670656d` AGENTS.md); prior baseline HEAD `b162e22` == v2.6.0 == remote |
| **Instruments** | art-dupl 0.7.0-81ce00b (`--sort total-tokens -t 3 --type-aware`); `nix develop -c go test -count=1`; `golangci-lint run ./internal/server/...` |
| **Scope discipline** | Per instruction: no research beyond what this session touched and noticed |

## Headline

The `-t 3` sweep found 5 clone groups, all priority low. **One was harmful and is now gone** (the
JSON contact-mutation epilogue → new `apiContactSaved` one-home in internal/server/contacts_api.go:85);
**three are deliberate residue** with on-record rationales (idempotency lock prologue,
empty-state markup, settings dt/dd rows). Re-run proves it: detected 49→47, shown 5→3, and the
3 shown are exactly the accepted trio. Tests 15/15, server lint 0 issues. The 2026-09-18 prior
triage held up under fresh inspection — the accept-with-rationale approach is durable.

Two honest self-caught issues, both fixed/recorded here: (1) I wrote a factually wrong location
("actions.go's import path") into a committed doc — it was actions.go's tab-form `saveContact`;
caught by grep at 03:00 and corrected inline. (2) I landed a second one-home helper with no
micro-test while "helper micro-tests" sat open on the previous report's own §f — a process
fuck-up, not a code one.

## Session timeline

1. Handoff from the 01-12 dedup-to-zero session (13 groups → 0 harmful).
2. User pasted a fresh `-t 3` HTML report: 49 detected / 5 shown / 44 suppressed, all low.
3. Loaded `deduplicate-code` skill; read all 5 sites at HEAD (concurrent-session rule honored).
4. Triage: extract groups #2+#4 (JSON contact tail), accept #1/#3/#5.
5. `apiContactSaved` extracted (3 edits, one file); exact notify→204 order preserved.
6. Verified: `go test -count=1 ./...` 15/15 ok; `golangci-lint run ./internal/server/...` 0 issues;
   art-dupl re-run: 49→47 detected, 5→3 shown = the accepted trio.
7. Docs: AGENTS.md one-home bullet + §h addendum in the 01-12 report; daemon committed.
8. Self-review for this report caught the §h location error → grepped real callers → fixed inline.

## Observations made this session (no action taken)

- **golangci-lint LSP keeps flagging `export_test.go:30` errcheck** although the line carries a
  valid `//nolint:staticcheck,errcheck` and the CLI gate prints 0 issues. The LSP integration
  appears not to honor nolint directives (or serves stale results). This warning has now
  polluted every diagnostics block this session.
- The 10 templ LSP warnings (`go.mod requires go >= 1.27.1 (running go 1.26.7)`) are the
  documented expected noise; all Go commands ran inside `nix develop -c`.
- The user's terminal executed the prompt's prose lines after the art-dupl command (exit 127
  on `"READ,"` etc.) — harmless; the report itself parsed fine. Multi-line prompts after a
  command want quoting or a heredoc.
- art-dupl 0.7.0 counter shape at `-t 3`: `Detected 47, 3 shown (13 non-actionable, 31 filtered
  suppressed)` — what the "filtered" bucket suppresses is not documented anywhere; relevant if
  `-t 3` becomes the ritual baseline.
- `-t 1` (prior session: 43-45 shown) vs `-t 3` (5 shown): the higher threshold surfaces
  triageable work instead of idiom noise — evidence for prior §f item 18.

## a) FULLY DONE

1. **Skill-loaded triage of all 5 shown groups**, each read at current HEAD before judging
   (re-read-before-edit reflex held; one stale-mtime refusal on the 01-12 doc was re-read, not forced).
2. **`apiContactSaved` one-home extracted** — the `notifyContactsChanged` + bare-204 epilogue
   that `apiSaveContact` and `apiDeleteContact` both repeated. Its doc comment now owns the
   "JSON mutations answer 204, island re-fetches" contract; behavior (notify before 204) preserved
   verbatim. The tab handlers (`saveContact`/`deleteContact` in actions.go) keep
   notify + toast + partial — a 4-params-for-3-lines abstraction was declined per the skill's bar.
3. **Three accepts re-confirmed with rationale homes**: idempotency.go clock+lock prologue
   (in-code comment at the site, written last session); history.templ ↔ voicemail.templ
   error+needsAPI markup (2026-09-18 §a.5 prior triage; per-panel i18n keys make a shared
   component a 3-params-for-2-lines trade); settings.templ dt/dd rows (divergent value shapes —
   orUnset / plain / fmtInt / conditional / link — a `settingsRow` would half-cover 5 of 7 rows).
4. **Verification**: full `go test -count=1 ./...` 15/15 packages ok; `golangci-lint` on the
   touched package 0 issues; art-dupl re-run at the user's exact flags minus `--html`.
5. **Docs updated**: AGENTS.md one-home-helpers bullet extended with `apiContactSaved`;
   §h follow-up addendum appended to the 01-12 status report (annotate-don't-rewrite).
6. **Daemon landed everything**; tree clean at 03:00:43; no manual commits (harness contract).
7. **§h factual error corrected** (import path → tab-form `saveContact`) after grepping the
   actual `contactSaveFailed` callers (exactly two: actions.go:231, contacts_api.go:129).

## b) PARTIALLY DONE

1. **Helper test coverage**: the suite is green, but `apiContactSaved` has no micro-test pinning
   the notify→204 order or the contract — and neither did the 9 helpers from the prior train.
   I also never inventoried WHICH existing tests cover `apiSaveContact`/`apiDeleteContact`;
   "15/15 ok" is not proof the epilogue contract is pinned anywhere.
2. **Verification scope**: lint ran on `./internal/server/...` only (the touched package), not
   the full `./internal/...` sweep the prior session ran. Smoke test skipped — a judgment call
   I made silently instead of stating a bar. No `-race`, no buildflow (consistent with prior
   session's deliberate skips, but here unremarked until now).
3. **`-t 3` ritual baseline**: this session gathered the supporting evidence (5 triageable vs
   `-t 1`'s ~45 noise groups) but the baseline is not ratified or written into AGENTS.md.
4. **Prior report's g.1–g.3 questions**: asked at 01-12, still unanswered — re-asked in §g below.
5. **HARVEST**: two status reports now carry §f lists (20 + this one); TODO_LIST.md/ROADMAP.md
   have not been fed — awaiting instruction per "THEN WAIT".

## c) NOT STARTED

Everything on the carry-forward list (prior report §f, still open as far as this session could
see — none of it was touched): `-race` on internal/server; full buildflow on a quiet machine;
island JS suite re-run (parallel session touched panels.js + island-tests); error-contract
cross-check vs the consolidated webhook tail; AGENTS.md smoke-line 38→40+4 drift fix; runbook
note about the v2.6.0-mid-refactor coordination hazard; `wp-empty` extraction trigger bar;
`ValidOutboundStatus` roadmap-only decision; docs/status archive sweep; aarch64 ELF check;
erraudit tier-2 re-measure (due 2026-10-22 per AGENTS.md). Plus everything new in §f below.

## d) TOTALLY FUCKED UP

Nothing is broken: no compile breakage, no bad intermediate reached history (the two red
moments last session were already verified clean), tree green at every checkpoint, no ghost
systems (`apiContactSaved` is wired at both call sites), no split brains (the nudge stays
one-home in `notifyContactsChanged`; the new helper composes it), nothing useful was removed,
no scope creep — no unrelated files touched.

What WAS fucked up, process-grade:

1. **A factually wrong claim landed in a committed doc.** §h said the extracted epilogue was
   "paired with `contactSaveFailed` in actions.go's import path". The real pairing is
   actions.go's tab-form `saveContact` (actions.go:231); the vCard import loop is not a caller.
   Root cause: I had READ `saveContact` at lines 211-237 minutes earlier and still wrote
   "import path" — I let the handoff summary's phrasing ("skip closure in the vCard import
   loop") bleed into a sentence about a different site. The grep took 5 seconds; the wrong
   sentence lived ~20 minutes in history before self-review caught it. Fixed inline 03:01.
2. **Second untested helper landed while the micro-tests item sat open on my own prior §f.**
   Knowing-and-skipping is worse than forgetting. The suite passed, but the new helper's
   contract (nudge order, 204) is unpinned by any test I can name.

## e) WHAT WE SHOULD IMPROVE

Direct answers to "what did you forget / do better / still improve":

- **Forgot**: to grep the `contactSaveFailed` callers BEFORE writing prose about them; to
  inventory existing coverage of the two JSON contact handlers; to state the smoke-skip bar
  out loud; that a ratified ritual baseline belongs in AGENTS.md (rules file), not just in
  timestamped reports.
- **Could have done better**: helper + micro-test should land in the same change — "tests after
  every change" includes tests FOR the change; prose file:line claims deserve the same
  grep-before-write discipline I apply to code edits; verification scope (server-only lint)
  should have been stated at verification time, not excavated in retrospective.
- **Could still improve**:
  1. **Acceptance-rationale registry**: rationale homes are scattered across in-code comments,
     the 2026-09-18 archived report, and §h. One registry (a short AGENTS.md dedup paragraph
     pointing at in-code anchors) would end the scatter — this session PROVED rationales get
     re-litigated otherwise (I re-read all three before accepting).
  2. **golangci-lint LSP nolint false positive**: every session pays an attention tax on the
     `export_test.go:30` warning; either fix the LSP integration/config or formally declare
     the CLI gate as the only truth (and make the LSP warning ignorable).
  3. **Skip-statement habit**: when a gate is deliberately skipped (smoke, -race, buildflow),
     name it in the final message with the reason — the prior session did this well; this one
     did it half-well.
  4. **art-dupl suppression opacity**: document what the 0.7.0 "filtered suppressed" bucket
     hides before crowning `-t 3` the baseline (§g Q1).

## f) TOP THINGS WE SHOULD GET DONE NEXT (up to 50 — brainstorm, most are ROADMAP fuel)

Carry = prior report §f (2026-09-23 01-12, items there numbered 1-20; still open). New = observed this session.

| # | Item | Source | Priority | Size |
|---|------|--------|----------|------|
| 1 | Micro-test `apiContactSaved`: pins notify-then-204 order + both JSON handlers route through it | New | High | S |
| 2 | Inventory which tests cover `apiSaveContact`/`apiDeleteContact`; add gap tests | New | High | S |
| ~~3~~ | ~~HARVEST both §f lists (01-12 + this report) into TODO_LIST.md/ROADMAP.md (docs-health)~~ done — 2026-09-23 docs-health HARVEST | ~~Carry~~ | ~~High~~ | ~~M~~ |
| 4 | Answer/ratify `-t 3` as THE dedup ritual baseline + document art-dupl's "filtered suppressed" bucket; one AGENTS.md line | Carry+New | High | S |
| 5 | `go test -race ./internal/server/...` once, on a quiet machine | Carry | Medium | S |
| 6 | Helper micro-tests for the prior train's 9 helpers (must, OrClock, updatedOrNotFound, formatFor, crmNumbers, applyStatusWebhook, recordCallIdem, contactSaveFailed, + idem behavior) | Carry | Medium | M |
| ~~7~~ | ~~Full buildflow run (with gitleaks/codespell) on a quiet machine; process findings~~ done — 02:47 release gates ran it on the tagged tree; post-tag deltas ride the TAIL | ~~Carry~~ | ~~Medium~~ | ~~M~~ |
| 8 | Process g.1/g.2 answers when given: webhook 400 body-text dependents; `msg/`→`message/` idem key rename acceptance | Carry | Medium | S |
| 9 | Cross-check docs/error-contract.md vs `applyStatusWebhook` consolidation (both-sides-in-sync rule) | Carry | Medium | S |
| 10 | Decide whether the JSON-204 mutation contract belongs in error-contract.md too | New | Low | S |
| 11 | Fix or formally ignore the golangci-lint LSP nolint false positive on `export_test.go:30` | New | Medium | S |
| 12 | Full-tree lint `./internal/...` re-run (this session covered server only) | New | Low | S |
| ~~13~~ | ~~Smoke re-run after the epilogue consolidation (cheap binary-level confidence)~~ done — release gates smoke green on the tagged tree | ~~New~~ | ~~Low~~ | ~~S~~ |
| ~~14~~ | ~~AGENTS.md smoke line: documented "38-check" vs observed 40+4 drift~~ done — fixed 38 to 40 by the 2026-09-23 sweep | ~~Carry~~ | ~~Low~~ | ~~S~~ |
| ~~15~~ | ~~Runbook note: v2.6.0 was cut while a refactor session was mid-flight (coordination hazard)~~ done — TODO runbook-hardening row | ~~Carry~~ | ~~Medium~~ | ~~S~~ |
| 16 | Acceptance-rationale registry: one home for all dedup ACCEPT decisions (see §e.1) | New | Medium | S |
| ~~17~~ | ~~Island JS suite re-run (parallel session's panels.js + island-tests changes, gate-relevant)~~ done — 79/79 green, 02:47 | ~~Carry~~ | ~~Medium~~ | ~~S~~ |
| 18 | `wp-empty` extraction trigger: is "10th simple usage" still the bar (9 sites accepted) | Carry | Low | S |
| ~~19~~ | ~~`ValidOutboundStatus` stays roadmap-only UNLESS webhook-valid and service-apply sets diverge~~ done — routed ROADMAP (composer cluster line) | ~~Carry~~ | ~~Low~~ | ~~S~~ |
| ~~20~~ | ~~Sweep docs/status archive: confirm no pre-2026-09-18 dup registers linger~~ done — 2026-09-23 archive sweep — no pre-2026-09-18 dup registers linger | ~~Carry~~ | ~~Low~~ | ~~S~~ |
| 21 | aarch64 cross-build ELF verification at next release train (post-refactor byte check) | Carry | Medium | S |
| 22 | templ `settingsRow` component: only if `-t 3` keeps surfacing the dt/dd rows (declined today) | New | Low | S |
| ~~23~~ | ~~erraudit tier-2 re-measure (family-adoption count; due 2026-10-22 per AGENTS.md)~~ done — standing row — next 2026-10-22 | ~~Carry~~ | ~~Medium~~ | ~~M~~ |
| 24 | Concurrent-session rule: make re-read-before-edit the reflex (validated again this session: one stale-mtime refusal on the 01-12 doc) | Carry | Low | S |
| ~~25~~ | ~~Verify remote sync (`git ls-remote`) at next gate: 4+ commits landed since the `b162e22` verification~~ done — verified 2026-09-23: origin == HEAD | ~~New~~ | ~~Low~~ | ~~S~~ |
| ~~26~~ | ~~Declare CLI lint the single lint truth in AGENTS.md if the LSP false positive can't be fixed (ties to #11)~~ done — merged into the TODO LSP row | ~~New~~ | ~~Low~~ | ~~S~~ |

No padding beyond 26: the honest backlog is these; items 27-50 would be invented.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Baseline ratification (merges prior g.3 + §f.4):** Is `art-dupl 0.7.0 --sort total-tokens -t 3
   --type-aware` now THE dedup ritual baseline for future sessions — and do you ratify this sweep's
   triage (1 extract / 3 accept), including the declined tab-handler abstraction and the declined
   `settingsRow` templ component?
2. **Still unanswered from 01:00 (prior g.1 + g.2):** (a) Does any stack-side tooling/runbook grep
   the webhook 400 body texts, or depend on the message hook's OLD validation precedence
   (provider_ref before status)? (b) Is the in-memory idem key rename `msg/<ref>` → `message/<ref>`
   acceptable, given no persistence?
3. **Testing bar for one-home helpers:** Should a helper land ONLY with a micro-test pinning its
   contract (would have blocked `apiContactSaved` today), or is suite-level coverage acceptable
   with micro-tests batched later as §f.6? This decides whether §f.1 is a rule or a task.

---

_Point-in-time snapshot — goes stale. Feed (f) to `docs-health` HARVEST; annotate, never rewrite, when bringing current later._
