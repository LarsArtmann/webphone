# Session status — branching-flow mixin report triage (2026-10-05 14:50)

**Session window:** 2026-10-05 ~14:40–14:50 · **Repo:** webphone (main) · **Scope:** this
session only — the mixin-report review and what was noticed in passing. No new
research beyond it (per owner instruction).

**What this session was:** the owner pasted an 8-row `branching-flow mixins`
report and asked for a review. Verdict delivered: **8/8 suggestions REJECTED,
0 refactors, docs-only changes.** Companion evidence doc (written this
session): [2026-10-05_14-46_branching-flow-mixin-review-status.md](2026-10-05_14-46_branching-flow-mixin-review-status.md).

---

## a) FULLY DONE

1. **Skill routing honored** — `dedup-code` SKILL.md loaded BEFORE judging
   (mandatory activation flow); `status-report` loaded for this report.
2. **Dedup-registry protocol followed end-to-end**: registry read first
   (no re-litigation of standing rulings), detector re-run at HEAD
   (`branching-flow mixins .` matched the owner's paste 8/8 — attribution
   against HEAD, not the paste), every flagged struct re-read at HEAD.
3. **All 8 findings judged with type-level evidence** — verdicts grounded in
   facts, not vibes: `fax.Service` vs `messaging.Service` share field NAMES
   but not TYPES; `InboundMessage`'s doc comment states the no-lifecycle
   design; config koanf tags are the operator contract; `crm.Client` vs
   `pbx.Client` unexported fields can't cross packages without a new seam.
4. **Cross-repo precedent found and applied** — the CV project's own compose
   analysis (2026-03-06) concluded most compose flags are valid Go idioms,
   and its executed plan DISSOLVED `Base*` mixins into explicit fields. The
   detector's suggestion direction is the opposite of the owner's executed
   trajectory.
5. **Evidence doc written** — per-finding verdict table + standing detector
   policy note at `docs/status/2026-10-05_14-46_branching-flow-mixin-review-status.md`.
6. **Registry updated per protocol step 4** ("one line, nothing else") —
   sweep-log line at `docs/dedup-registry.md:76` linking the status doc.
7. **Concurrent-session discipline** — the pre-existing `gopls unusedparams`
   Info at `internal/server/passkey_api.go:211` belongs to another session;
   left untouched and attributed, not "fixed".

## b) PARTIALLY DONE

1. **Verification tail of the docs change** — the two markdown writes were
   never followed by `git log`/`git ls-remote` to confirm the auto-commit
   daemon picked them up (release-runbook obligation even for docs), and no
   markdown-lint detect pass ran on them (detect-only tooling, buildflow's
   lane). High confidence both are fine; unverified.
2. **QMD breakage noticed but undiagnosed** — `mcp_qmd_get` returned raw Go
   pointer garbage (`&{0x... ...}`) for BOTH docid and path lookups (2
   calls); worked around via the corpus path on disk (`~/projects/CV`,
   found via `mcp_qmd_status`). The bug was never reported or logged
   anywhere durable in-session.

## c) NOT STARTED

1. **HARVEST of this report into TODO_LIST.md** — deliberately deferred:
   the owner ordered "WAIT FOR INSTRUCTIONS" after writing (overrides the
   skill's run-HARVEST-now nudge).
2. **AGENTS.md policy line for the mixin detector** — deliberately not
   started: AGENTS.md is line-capped (377, BuildFlow preflight) and the
   registry is the ONE home for rulings by design. Listed as owner call
   instead (see g).
3. **QMD `get` diagnosis** — not attempted (crush_logs MCP triage); out of
   the review's scope, in scope for a follow-up if the owner wants.

## d) TOTALLY FUCKED UP!

1. **Byte-accuracy failure on the registry edit** — I retyped the 2026-10-04
   sweep-log line as `wp_mini` where the file says `wp-mini`; the edit
   failed AND I initially misattributed it to concurrent-daemon drift before
   checking bytes (`cat -A`). One wasted round trip, root cause: retyping
   instead of verbatim copying, plus blaming the daemon before blaming
   myself. Recovered same minute; final line landed exactly once in
   chronological position.
2. **Two burned tool calls on a broken MCP tool** — repeated `mcp_qmd_get`
   after the first garbage response instead of pivoting to disk
   immediately. Cost: one round trip + noise. (Tool bug is QMD's; the
   repeat was mine.)

Nothing code-level was fucked up: zero Go/templ/JS files touched, zero
tests affected, zero contract surfaces (DOM/config/wire) near the change.

## e) WHAT WE SHOULD IMPROVE!

1. **Edit anchoring discipline**: prefer a SHORT unique anchor copied
   verbatim over a long retyped line; when an edit fails, check MY bytes
   first, daemon drift second (the daemon is adversarial, but so is
   human transcription).
2. **End-state verification even for docs-only changes**: one `git log -3`
   after writes in a multi-session repo; cheap, catches daemon misses and
   accidental cross-session tangles.
3. **Surface tool breakage immediately**: a broken MCP tool should get a
   crush_logs check and an in-report note the moment it misbehaves, not a
   silent workaround.
4. **Markdown hygiene**: run the detect-only markdown-lint over new status
   docs before ending the session (buildflow lane; two files, seconds).
5. **Detector reports ≠ findings** (now codified in the 14:46 doc): shape
   detectors need a triage bar — identical field types + shared LOGIC
   before any embed suggestion is actionable. Future sessions inherit this
   via the registry line.

## f) Things to get done next (session-derived; deliberately NOT padded to 50)

Scope discipline: the owner capped research at "what this session did and
noticed" — inventing 40 unrelated items would be fabrication. Honest list:

1. Verify the daemon committed the mixin-review status doc + registry line
   (`git log`, `git ls-remote`); amend if mangled.
2. markdown-lint detect over the two new markdown files.
3. HARVEST this report + the 14:46 doc into TODO_LIST.md (docs-health)
   once the owner green-lights.
4. Owner call: ratify the mixin-detector triage policy as a standing
   registry row (like the art-dupl `-t 3` baseline call) so mixin rows
   stop needing fresh judgment every run.
5. Owner call: standing ruling (or explicit non-ruling) for the
   `config.CRM` / `config.Paperless` URL+Token twin — the one genuinely
   debatable shape-similarity in the set; a row would pre-answer future
   detector + art-dupl resurfacing.
6. Diagnose `mcp_qmd_get` garbage output (crush_logs `[mcp]` section;
   reproduce; fix or file upstream against the QMD server).
7. Re-check `internal/server/passkey_api.go:211` unused-`r` gopls Info once
   the concurrent passkey session lands — confirm it resolved or hand it
   to that session; do NOT fix cross-session.
8. The registry's four OPEN OWNER CALLS are all still open (noticed while
   reading it): `-t 3` baseline ratification, `settingsRow` build-or-retire,
   registry-as-one-home ratification, suppressed-set audit scope.
9. If webphone has never had a `branching-flow compose` baseline (unverified
   this session), one run would calibrate how noisy the detector family is
   for THIS codebase before anyone acts on its output again.
10. Consider a one-line pointer in AGENTS.md's dedup bullet to "mixin
    reports triaged 2026-10-05" ONLY if the registry line proves
    insufficient later — the 377-line cap argues against doing it now.

## g) Questions for the owner (cannot be figured out from here)

1. **Mixin-detector policy home**: is the registry sweep-log line + status
   doc enough, or do you want it promoted to an AGENTS.md hard-won rule
   (at the cost of the 377-line cap) / a new standing-rulings row?
2. **CRM/Paperless twin**: standing ruling row (pre-empts future
   resurfacing) or leave the rejection in the sweep log only?
3. **QMD `get` bug**: known breakage on your side, or shall I run a
   diagnosis session (crush_logs → reproduce → upstream fix/file)? It
   affects every future QMD-powered lookup in this environment.

---

**Process note:** written as `.md` at the owner's explicit path demand —
this overrides the status-report skill's HTML-canonical default for this
instance only. Commit intentionally skipped (harness: never commit without
explicit request); the auto-commit daemon owns pickup.
