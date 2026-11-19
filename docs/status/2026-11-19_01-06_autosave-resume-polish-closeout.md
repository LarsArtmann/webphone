# Notes-autosave resume: polish lanes, CI attribution, honest accounting

> Snapshot-rot header: claims verified as of 2026-11-19 01:06 CET.
> Scope: THIS session's run only (resuming the 21:55 closeout at the
> owner's "break this down and execute" order): the §f ledger's
> non-owner-gated items, two CI red attributions, one fix-forward push,
> and the brutal accounting. The prior session's report
> (`2026-11-18_21-55_notes-autosave-closeout-and-self-review.md`) owns
> the train itself; the 3 owner questions there are STILL OPEN and
> still gate the polish lanes.

## The session in one paragraph

Resumed with the §f ledger triage: the e2e advisory red turned out to
be a deterministic nix `getFlake "path:."` relative-path error (not
the browser flake, not my markup) that the concurrent e2e train
already fixed (`2076c4c0`) and CI proved green; the maxlength row was
already pinned by `TestCallNoteGuards`; and the real work this session
was three landed changes (echo-silent SSE nudges, island aria-live
parity, the lessons war story + 18:39 supersession banner), a
fix-forward push that healed a daemon-caused docs format red, and one
miss: that same push carried a concurrent session's report that
re-tipped the cadence gate, which a 30-second full local battery
would have caught.

## a) FULLY DONE (this session, verified)

| What                                                                                                                                                                                                         | Where                                         | Evidence                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------- | --------------------------------------------------------------------------------- |
| e2e advisory red ATTRIBUTED: deterministic nix error (relative `path:.` in the workflow's PYSEL expr), journeys never ran                                                                                    | CI run 44397661442 step log                   | owning session's fix `2076c4c0`; journeys GREEN on runs 44412428019 + 44419228536 |
| §f.7 (maxlength boundary) found ALREADY PINNED — the ledger row was stale                                                                                                                                    | `TestCallNoteGuards`                          | 4000 CJK = 204, 4001 = 413, both asserted                                         |
| Echo-silent SSE nudges: `NotesGate.Set` answers `(changed bool, error)`; a same-body echo over an existing slot skips the dispatch entirely (decider guard stays the race authority; first save still mints) | `notesgate.go`, `callnotes.go`, `note_api.go` | commit `a164ca3b`; readmodels+cqrs+arch+server green                              |
| Echo pins both levels: cqrs `TestNoteSetNamesWhetherItChanged`; server `TestCallNoteSaveNudgesHistory` extended (echo = 204 + NO nudge; revision = exactly one more)                                         | the two test files                            | 11 call sites updated to the new signature; suites green                          |
| Island aria-live parity: `.call-note-status` span carries `aria-live="polite"` (History-span twin)                                                                                                           | `postcall.js` `noteRow`                       | commit `d0cf4cb2`; spec pin in the debounce test; 9/9; tsc exit 0                 |
| lessons war story: carried-body terminating-retry invariant + module-global flush-map isolation + the stub lessons                                                                                           | `docs/lessons.md`                             | commit `273ec280`                                                                 |
| 18:39 report supersession banner (its "shell.js bug still live" claim can no longer mislead)                                                                                                                 | the 18:39 report header                       | commit `273ec280`                                                                 |
| AGENTS post-call-notes bullet carries the echo semantics; CHANGELOG lines for both code changes                                                                                                              | `AGENTS.md`, `CHANGELOG.md`                   | `184da4dc` + daemon's `22a6756c` (content preserved, message eaten)               |
| Gates at the working tree: island 364/364; Go 37/37 rc=0; tsc; sandboxed format-check; docs-gates --check                                                                                                    | this session's runs                           | all exit 0 (docs-gates at its then-tree; later commits changed the verdict)       |
| FIX-FORWARD: daemon-committed TODO_LIST row was dprint-raw -> CI format rc=20 red; canonicalized + pushed under the manual-push bar (red main + verified format-only fix)                                    | commit `86614220`                             | CI format gate PASSED on that run (was FAIL on `0ee214d6`)                        |
| My surfaces on pushed CI: e2e journeys, go-race, smoke, nix build, art-dupl, reproduce ALL GREEN                                                                                                             | runs 44419228536, 44419228536-era             | job-level view; sole reds were the two docs-battery states below                  |

## b) PARTIALLY DONE

- **CI green end-state not reached.** The `86614220` push healed the
  format red but the run still concluded failure at `FAIL cadence`:
  the concurrent docs session's interrupted pareto report (`52e09341`)
  re-tipped live reports 65 -> 66 (the exact fight their `9a17c6f7`
  "drain one more" commit was mid-way through). Their active lane;
  per the concurrent-session protocol I did not touch their archive
  ritual. My two docs commits (`22a6756c`, `184da4dc`) were still
  daemon-unpushed at report time.
- **buildflow full battery red, attributed not fixed:** file-size-check
  47 error findings with the baseline matching 0/220 entries (the
  BuildFlow binary is stale vs its own repo — preflight warn names it)
  and jscpd 1 finding (`calls-reason.test.mjs`, the e2e train's).
  Neither tool runs in CI. My honest share: `callnotes_test.go`
  crossed 347 -> 368 against the 350 max FRESH (see d); `callnotes.go`
  was already over at 396.

## c) NOT STARTED (from the 21:55 §f ledger, deliberately)

- §f.8 `withIsolatedNoteRows` test helper — the per-URL assertion
  scoping is already landed in the specs and the war story documents
  the trap; §f.14 keeps a production reset seam deferred (served-byte
  cost). Revisit only if a future spec re-hits the leak.
- §f.13 ui-capture shot of the note status line — the CI e2e journeys
  now cover this repo's served-markup obligation (and they run green);
  the stack-side browser E2E remains a release-runbook obligation,
  not a this-repo one.

## d) TOTALLY FUCKED UP

- **I pushed a red I could have preflighted.** The manual push of
  `86614220` carried the concurrent session's `52e09341` (cadence
  66/65) alongside my format fix. I ran only scoped checks (dprint
  tree check + status-index) before pushing, not the 30-second full
  docs battery that names cadence. "Attempts verify, they must not
  discover" — my push discovered. The fix itself was right; the
  rider was foreseeable.
- **The aria-live CHANGELOG line missed its code commit** by ~40
  minutes (remembered during closeout; the daemon took it as
  `22a6756c` with the narrative eaten). The bar says the line rides
  the same commit as the code-path change.
- **First CI poll loop treated rc=2 as terminal** — but rc=2 is
  red-OR-pending, and the docs literally say a poll loop must wait
  out pending. Wrote the buggy version first, knew the rule.
- **Re-hit the documented PIPESTATUS gap** (`GO_SUITE_RC=unknown`) and
  had to re-run the full Go suite for an honest exit code.
- **Crossed callnotes_test.go over the file-size budget unpremeditated**
  (347 -> 368 vs max 350): I did not check the budget before adding
  the 21-line pin. Ruled keep-the-pin-and-document rather than
  deleting a just-landed correctness pin to duck a threshold whose
  gate is tree-wide red — but the crossing itself was sloppy.

## e) WHAT WE SHOULD IMPROVE

1. Before ANY manual push, run the FULL local docs battery when the
   push carries foreign docs commits — the bar's "verified fix" must
   verify the whole push, not just my file.
2. CHANGELOG lines ride the code commit, always — a post-hoc
   remember is a daemon snack.
3. Check the file-size budget BEFORE adding test code to a file at
   99% of its threshold.
4. Copy poll loops from the documented semantics, not from memory of
   them.
5. The daemon's raw TODO_LIST commits keep costing rc=20 reds — the
   systemic fix is a dprint check in the pma push leg (fleet-wide
   policy call), not per-red canonicalization.

## f) Next things

<!-- check-rows:ledger -->

| #  | Thing                                                                                                                                        | Why now                              | Size |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ---- |
| 1  | Answer the 3 owner questions (Save button, AI-draft autosave, failure loudness) — unchanged from the 21:55 report                            | gates every polish lane              | S    |
| 2  | Docs session drains cadence 66 -> 65 (their active lane; my report adds +1 when it lands)                                                    | main's only CI red                   | S    |
| 3  | Reinstall the BuildFlow binary (stale `6c86a21` vs repo HEAD `f5a0f13`) and re-check the 0/220 baseline match                                | file-size battery is unjudgeable     | S    |
| 4  | file-size closeout train: re-baseline or split; adjudicate callnotes_test.go 368 and callnotes.go 412                                        | 47 error findings                    | M    |
| 5  | jscpd duplicate in calls-reason.test.mjs (e2e train's file)                                                                                  | battery red                          | XS   |
| 6  | Interstitial "Saving…" is announced by both status spans — the a11y guide says interstitials are noise; tie the fix to owner Q3              | SR chatter per typing pause          | S    |
| 7  | ui-capture shot of the History note box + status line in situ                                                                                | visual ledger                        | S    |
| 8  | Disposition saves (`SetDisposition` lane) publish NO SSE nudge — other tabs go stale on verdict changes (pre-existing, noticed this session) | same changed-verdict pattern applies | S    |
| 9  | Stack-side browser E2E for the history.templ markup change                                                                                   | release-runbook obligation           | M    |
| 10 | Consider `aria-atomic` on the note status spans                                                                                              | announce precision                   | XS   |
| 11 | Measure the `Notes.Set` pre-read cost if autosave volume ever matters (one keyed projection read per save)                                   | unmeasured, believed tiny            | XS   |
| 12 | The 21:55 report's §f rows get adjudicated by the next docs-gates annotate sweep — no manual strike owed                                     | docs-health owns it                  | XS   |
| 13 | Daemon-side dprint gate for committed markdown (pma module, fleet-wide policy call)                                                          | the recurring rc=20 class            | M    |

## g) Questions the owner must answer (cannot be derived)

1. The three 21:55 questions stand (Save button semantics; AI-draft
   autosave doctrine; failure loudness) — every remaining polish lane
   keys on them.
2. Manual-push bar scope: when a bar-sanctioned push must carry
   another session's commits, must the WHOLE push be battery-green
   first (my miss tonight says yes — ratify it?), or is attributed-red
   acceptable when the red is provably the other session's active
   lane?
3. Keep the cqrs-level `TestNoteSetNamesWhetherItChanged` pin even
   though it pushed `callnotes_test.go` over the file-size budget
   (my ruling: yes, correctness pins outrank size thresholds and the
   split train owns the carve), or trim it under 350 now?
